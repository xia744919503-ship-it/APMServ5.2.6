package legacy

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"
)

const systemMailFromName = "系统"
const autoMailNotice = "【系统自动提示信息：官方不会以任何形式在个人信件中通知用户中奖，如果您收到此类信件，请不要相信，更不要向信息发布者汇款，谨防受骗！】\n\n"

const legacyMaxMailCount = 100

const (
	mailInboxFullMessage      = "\u6536\u4ef6\u7bb1\u5df2\u6ee1\uff0c\u8bf7\u5220\u9664\u591a\u4f59\u4fe1\u4ef6\u540e\u518d\u53d1\u9001\u3002"
	mailOutboxFullMessage     = "\u53d1\u4ef6\u7bb1\u5df2\u6ee1\uff0c\u8bf7\u5220\u9664\u591a\u4f59\u4fe1\u4ef6\u540e\u518d\u53d1\u9001\u3002"
	mailEnemyBlockedMessage   = "\u4f60\u5728\u5bf9\u65b9\u4ec7\u4eba\u540d\u5355\u4e2d\uff0c\u65e0\u6cd5\u53d1\u4fe1"
	mailContentIllegalMessage = "\u4fe1\u4ef6\u5185\u5bb9\u5305\u542b\u975e\u6cd5\u5b57\u7b26\uff0c\u4e0d\u80fd\u53d1\u9001"
)

func (r *Repository) MailPage(ctx context.Context, uid int, folder string, page int) (MailPage, error) {
	folder = normalizeMailFolder(folder)
	if r.db == nil {
		return MailPage{
			Folder: folder,
			Counts: MailCounts{},
			Items:  []MailSummary{},
		}, nil
	}

	counts, err := r.mailCounts(ctx, uid)
	if err != nil {
		return MailPage{}, err
	}

	total := totalForMailFolder(counts, folder)
	pageCount := 0
	if total > 0 {
		pageCount = (total + legacyPageSize - 1) / legacyPageSize
		page = clamp(page, 0, pageCount-1)
	} else {
		page = 0
	}

	items := make([]MailSummary, 0, legacyPageSize)
	if total > 0 {
		rows, err := r.mailListRows(ctx, uid, folder, page)
		if err != nil {
			return MailPage{}, err
		}
		defer rows.Close()

		items, err = scanMailRows(rows, folder)
		if err != nil {
			return MailPage{}, err
		}
	}

	return MailPage{
		Folder:    folder,
		Page:      page,
		PageCount: pageCount,
		Total:     total,
		Counts:    counts,
		Items:     items,
	}, nil
}

func (r *Repository) MailDetail(ctx context.Context, uid int, folder string, id int) (MailDetail, error) {
	folder = normalizeMailFolder(folder)
	if r.db == nil {
		return MailDetail{}, sql.ErrNoRows
	}

	summary, content, err := r.mailDetail(ctx, uid, folder, id)
	if err != nil {
		return MailDetail{}, err
	}

	switch folder {
	case "inbox":
		if _, err := r.db.ExecContext(ctx, "update sys_mail_box set `read` = 1 where `mid` = ? and `uid` = ?", id, uid); err != nil {
			return MailDetail{}, err
		}
		summary.Read = true
	case "system":
		if _, err := r.db.ExecContext(ctx, "update sys_mail_sys_box set `read` = 1 where `mid` = ? and `uid` = ?", id, uid); err != nil {
			return MailDetail{}, err
		}
		summary.Read = true
	}

	if err := r.clearMailAlarmIfNoUnread(ctx, uid); err != nil {
		return MailDetail{}, err
	}

	summary.Snippet = mailSnippet(content)
	if summary.Snippet == "" {
		summary.Snippet = normalizeMailTitle(summary.Title)
	}

	return MailDetail{
		Folder:       folder,
		Summary:      summary,
		HTMLDocument: wrapMailHTML(summary, content),
	}, nil
}

func (r *Repository) DeleteMail(ctx context.Context, uid int, folder string, ids []int, page int) (MailPage, error) {
	folder = normalizeMailFolder(folder)
	if r.db == nil {
		return MailPage{}, ErrDatabaseUnavailable
	}

	filteredIDs := uniquePositiveInts(ids)
	if len(filteredIDs) == 0 {
		return r.MailPage(ctx, uid, folder, page)
	}

	query, args, err := deleteMailQuery(uid, folder, filteredIDs)
	if err != nil {
		return MailPage{}, err
	}

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return MailPage{}, err
	}

	if folder == "inbox" || folder == "system" {
		if err := r.clearMailAlarmIfNoUnread(ctx, uid); err != nil {
			return MailPage{}, err
		}
	}

	return r.MailPage(ctx, uid, folder, page)
}

func (r *Repository) SendMail(ctx context.Context, uid int, toName string, title string, content string) (MailDetail, error) {
	if r.db == nil {
		return MailDetail{}, ErrDatabaseUnavailable
	}

	if err := r.checkMailCapacity(ctx, uid); err != nil {
		return MailDetail{}, err
	}

	senderName, err := r.mailUserDisplayName(ctx, uid)
	if err != nil {
		return MailDetail{}, err
	}

	recipient, err := r.mailRecipient(ctx, toName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return MailDetail{}, fmt.Errorf("recipient %q was not found: %w", strings.TrimSpace(toName), sql.ErrNoRows)
		}
		return MailDetail{}, err
	}

	if err := r.checkMailEnemyRestriction(ctx, uid, recipient.UID); err != nil {
		return MailDetail{}, err
	}

	if err := r.checkBannedMailContent(ctx, uid, senderName, content); err != nil {
		return MailDetail{}, err
	}

	normalizedTitle := normalizeMailTitle(title)
	normalizedContent := autoMailNotice + strings.TrimSpace(content)
	if strings.TrimSpace(content) == "" {
		normalizedContent = autoMailNotice
	}
	postedAt := time.Now().Unix()

	contentResult, err := r.db.ExecContext(ctx, `
insert into sys_mail_content (content, posttime)
values (?, ?)`, normalizedContent, postedAt)
	if err != nil {
		return MailDetail{}, err
	}

	contentID, err := contentResult.LastInsertId()
	if err != nil {
		return MailDetail{}, err
	}

	mailResult, err := r.db.ExecContext(ctx, `
insert into sys_mail_box (uid, name, fromuid, fromname, contentid, title, `+"`read`"+`, recvstate, sendstate, posttime)
values (?, ?, ?, ?, ?, ?, 0, 0, 0, ?)`,
		recipient.UID,
		recipient.Name,
		uid,
		senderName,
		contentID,
		normalizedTitle,
		postedAt,
	)
	if err != nil {
		return MailDetail{}, err
	}

	mailID, err := mailResult.LastInsertId()
	if err != nil {
		return MailDetail{}, err
	}

	if _, err := r.db.ExecContext(ctx, `
insert into sys_alarm (uid, mail)
values (?, 1)
on duplicate key update mail = 1`, recipient.UID); err != nil {
		return MailDetail{}, err
	}

	return r.MailDetail(ctx, uid, "outbox", int(mailID))
}

func (r *Repository) mailCounts(ctx context.Context, uid int) (MailCounts, error) {
	var counts MailCounts
	err := r.db.QueryRowContext(ctx, `
select
	(select count(*) from sys_mail_box where uid = ? and recvstate = 0),
	(select count(*) from sys_mail_box where fromuid = ? and sendstate = 0),
	(select count(*) from sys_mail_sys_box where uid = ?),
	(select count(*) from sys_mail_box where uid = ? and recvstate = 0 and `+"`read`"+` = 0),
	(select count(*) from sys_mail_sys_box where uid = ? and `+"`read`"+` = 0)`,
		uid, uid, uid, uid, uid,
	).Scan(
		&counts.Inbox,
		&counts.Outbox,
		&counts.System,
		&counts.UnreadInbox,
		&counts.UnreadSystem,
	)
	if err != nil {
		return MailCounts{}, err
	}

	return counts, nil
}

func deleteMailQuery(uid int, folder string, ids []int) (string, []any, error) {
	placeholders := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, 0, len(ids)+1)

	switch folder {
	case "outbox":
		args = append(args, uid)
		for _, id := range ids {
			args = append(args, id)
		}
		return "update sys_mail_box set sendstate = 1 where fromuid = ? and mid in (" + placeholders + ")", args, nil
	case "system":
		args = append(args, uid)
		for _, id := range ids {
			args = append(args, id)
		}
		return "delete from sys_mail_sys_box where uid = ? and mid in (" + placeholders + ")", args, nil
	case "inbox":
		args = append(args, uid)
		for _, id := range ids {
			args = append(args, id)
		}
		return "update sys_mail_box set recvstate = 1 where uid = ? and mid in (" + placeholders + ")", args, nil
	default:
		return "", nil, fmt.Errorf("unsupported mail folder %q", folder)
	}
}

func totalForMailFolder(counts MailCounts, folder string) int {
	switch folder {
	case "outbox":
		return counts.Outbox
	case "system":
		return counts.System
	default:
		return counts.Inbox
	}
}

func normalizeMailFolder(folder string) string {
	switch folder {
	case "outbox", "system", "inbox":
		return folder
	default:
		return "inbox"
	}
}

type mailRecipientInfo struct {
	UID  int
	Name string
}

func (r *Repository) mailListRows(ctx context.Context, uid int, folder string, page int) (*sql.Rows, error) {
	offset := page * legacyPageSize

	switch folder {
	case "outbox":
		return r.db.QueryContext(ctx, `
select
	i.mid,
	i.fromname,
	i.name,
	i.title,
	i.`+"`read`"+`,
	i.posttime,
	c.content
from sys_mail_box i
left join sys_mail_content c on c.mid = i.contentid
where i.fromuid = ? and i.sendstate = 0
order by i.posttime desc
limit ?, ?`, uid, offset, legacyPageSize)
	case "system":
		return r.db.QueryContext(ctx, `
select
	i.mid,
	?,
	coalesce(u.name, ''),
	i.title,
	i.`+"`read`"+`,
	i.posttime,
	c.content
from sys_mail_sys_box i
left join sys_mail_sys_content c on c.mid = i.contentid
left join sys_user u on u.uid = i.uid
where i.uid = ?
order by i.posttime desc
limit ?, ?`, systemMailFromName, uid, offset, legacyPageSize)
	default:
		return r.db.QueryContext(ctx, `
select
	i.mid,
	i.fromname,
	i.name,
	i.title,
	i.`+"`read`"+`,
	i.posttime,
	c.content
from sys_mail_box i
left join sys_mail_content c on c.mid = i.contentid
where i.uid = ? and i.recvstate = 0
order by i.posttime desc
limit ?, ?`, uid, offset, legacyPageSize)
	}
}

func scanMailRows(rows *sql.Rows, folder string) ([]MailSummary, error) {
	items := make([]MailSummary, 0, legacyPageSize)
	for rows.Next() {
		var item MailSummary
		var fromName sql.NullString
		var toName sql.NullString
		var title sql.NullString
		var content sql.NullString
		var readFlag int
		var createdUnix int64

		if err := rows.Scan(
			&item.ID,
			&fromName,
			&toName,
			&title,
			&readFlag,
			&createdUnix,
			&content,
		); err != nil {
			return nil, err
		}

		item.Folder = folder
		item.FromName = strings.TrimSpace(fromName.String)
		item.ToName = strings.TrimSpace(toName.String)
		item.Title = normalizeMailTitle(title.String)
		item.Read = readFlag != 0
		item.CreatedAt = time.Unix(createdUnix, 0).Format("2006-01-02 15:04:05")
		item.Snippet = mailSnippet(content.String)
		if item.Snippet == "" {
			item.Snippet = item.Title
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) mailDetail(ctx context.Context, uid int, folder string, id int) (MailSummary, string, error) {
	var (
		query string
		args  []any
	)

	switch folder {
	case "outbox":
		query = `
select
	i.mid,
	i.fromname,
	i.name,
	i.title,
	i.` + "`read`" + `,
	i.posttime,
	c.content
from sys_mail_box i
left join sys_mail_content c on c.mid = i.contentid
where i.mid = ? and i.fromuid = ? and i.sendstate = 0`
		args = []any{id, uid}
	case "system":
		query = `
select
	i.mid,
	?,
	coalesce(u.name, ''),
	i.title,
	i.` + "`read`" + `,
	i.posttime,
	c.content
from sys_mail_sys_box i
left join sys_mail_sys_content c on c.mid = i.contentid
left join sys_user u on u.uid = i.uid
where i.mid = ? and i.uid = ?`
		args = []any{systemMailFromName, id, uid}
	default:
		query = `
select
	i.mid,
	i.fromname,
	i.name,
	i.title,
	i.` + "`read`" + `,
	i.posttime,
	c.content
from sys_mail_box i
left join sys_mail_content c on c.mid = i.contentid
where i.mid = ? and i.uid = ? and i.recvstate = 0`
		args = []any{id, uid}
	}

	var summary MailSummary
	var fromName sql.NullString
	var toName sql.NullString
	var title sql.NullString
	var content sql.NullString
	var readFlag int
	var createdUnix int64
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&summary.ID,
		&fromName,
		&toName,
		&title,
		&readFlag,
		&createdUnix,
		&content,
	); err != nil {
		return MailSummary{}, "", err
	}

	summary.Folder = folder
	summary.FromName = strings.TrimSpace(fromName.String)
	summary.ToName = strings.TrimSpace(toName.String)
	summary.Title = normalizeMailTitle(title.String)
	summary.Read = readFlag != 0
	summary.CreatedAt = time.Unix(createdUnix, 0).Format("2006-01-02 15:04:05")
	summary.Snippet = mailSnippet(content.String)
	if summary.Snippet == "" {
		summary.Snippet = summary.Title
	}

	return summary, content.String, nil
}

func (r *Repository) clearMailAlarmIfNoUnread(ctx context.Context, uid int) error {
	counts, err := r.mailCounts(ctx, uid)
	if err != nil {
		return err
	}

	if counts.UnreadInbox == 0 && counts.UnreadSystem == 0 {
		if _, err := r.db.ExecContext(ctx, "update sys_alarm set `mail` = 0 where uid = ?", uid); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) mailUserDisplayName(ctx context.Context, uid int) (string, error) {
	var displayName string
	err := r.db.QueryRowContext(ctx, `
select
	case
		when trim(coalesce(name, '')) = '' then concat('UID ', uid)
		else name
	end as display_name
from sys_user
where uid = ?`, uid).Scan(&displayName)
	if err != nil {
		return "", err
	}

	return displayName, nil
}

func (r *Repository) mailRecipient(ctx context.Context, name string) (mailRecipientInfo, error) {
	normalized := strings.TrimSpace(name)
	if normalized == "" {
		return mailRecipientInfo{}, sql.ErrNoRows
	}

	recipient := mailRecipientInfo{}
	err := r.db.QueryRowContext(ctx, `
select
	uid,
	case
		when trim(coalesce(name, '')) = '' then concat('UID ', uid)
		else name
	end as display_name
from sys_user
where trim(coalesce(name, '')) = ?
	or case
		when trim(coalesce(name, '')) = '' then concat('UID ', uid)
		else name
	end = ?
order by uid asc
limit 1`, normalized, normalized).Scan(&recipient.UID, &recipient.Name)
	if err != nil {
		return mailRecipientInfo{}, err
	}

	return recipient, nil
}

func (r *Repository) checkMailCapacity(ctx context.Context, uid int) error {
	var inboxCount int
	if err := r.db.QueryRowContext(ctx, `
select count(*)
from sys_mail_box
where uid = ? and recvstate = 0`, uid).Scan(&inboxCount); err != nil {
		return err
	}
	if inboxCount > legacyMaxMailCount {
		return newInvalidError(mailInboxFullMessage)
	}

	var outboxCount int
	if err := r.db.QueryRowContext(ctx, `
select count(*)
from sys_mail_box
where fromuid = ? and sendstate = 0`, uid).Scan(&outboxCount); err != nil {
		return err
	}
	if outboxCount > legacyMaxMailCount {
		return newInvalidError(mailOutboxFullMessage)
	}

	return nil
}

func (r *Repository) checkMailEnemyRestriction(ctx context.Context, uid int, recipientUID int) error {
	var blockedCount int
	if err := r.db.QueryRowContext(ctx, `
select count(*)
from sys_user_relation
where uid = ? and tuid = ? and type = 1`, recipientUID, uid).Scan(&blockedCount); err != nil {
		return err
	}
	if blockedCount > 0 {
		return newForbiddenError(mailEnemyBlockedMessage)
	}

	return nil
}

func (r *Repository) checkBannedMailContent(ctx context.Context, uid int, senderName string, content string) error {
	rows, err := r.db.QueryContext(ctx, "select content from cfg_baned_mail_content")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var bannedContent sql.NullString
		if err := rows.Scan(&bannedContent); err != nil {
			return err
		}

		term := strings.TrimSpace(bannedContent.String)
		if term == "" || !strings.Contains(content, term) {
			continue
		}

		if _, err := r.db.ExecContext(ctx, `
insert into log_illegal_user (uid, name, count)
values (?, ?, 1)
on duplicate key update count = count + 1`, uid, senderName); err != nil {
			return err
		}

		return newInvalidError(mailContentIllegalMessage)
	}

	return rows.Err()
}

func normalizeMailTitle(title string) string {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return "无标题"
	}

	return trimmed
}

func mailSnippet(content string) string {
	plain := reportWhitespacePattern.ReplaceAllString(reportTagPattern.ReplaceAllString(content, " "), " ")
	plain = html.UnescapeString(strings.TrimSpace(plain))
	if plain == "" {
		return ""
	}

	runes := []rune(plain)
	if len(runes) <= 72 {
		return plain
	}

	return string(runes[:72]) + "..."
}

func wrapMailHTML(summary MailSummary, content string) string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8" />
<title>` + html.EscapeString(summary.Title) + `</title>
<style type="text/css">
body {
	margin: 0;
	padding: 16px;
	background: #17292B;
	color: #F4E9C8;
	font: 12px/1.6 "Microsoft YaHei", sans-serif;
}
.mail-shell {
	min-height: calc(100vh - 32px);
	padding: 14px 16px 18px;
	border: 1px solid #3F5A5E;
	background: linear-gradient(180deg, rgba(27, 49, 52, 0.98), rgba(16, 27, 29, 0.98));
	box-sizing: border-box;
}
.mail-body {
	margin-top: 14px;
	padding: 14px 16px;
	border: 1px solid #45666A;
	background: rgba(9, 18, 19, 0.92);
	word-break: break-word;
}
.mail-body p:first-child {
	margin-top: 0;
}
.mail-body p:last-child {
	margin-bottom: 0;
}
.mail-empty {
	color: #C7B48A;
}
.mail-text {
	white-space: normal;
}
a {
	color: #9FD9E3;
}
</style>
</head>
<body>
<div class="mail-shell">
  <div class="mail-body">` + normalizeMailBody(content) + `</div>
</div>
</body>
</html>`
}

func normalizeMailBody(content string) string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return `<div class="mail-empty">暂无内容</div>`
	}

	if strings.Contains(trimmed, "<") && strings.Contains(trimmed, ">") {
		return trimmed
	}

	escaped := html.EscapeString(trimmed)
	escaped = strings.ReplaceAll(escaped, "\r\n", "<br />")
	escaped = strings.ReplaceAll(escaped, "\n", "<br />")
	return `<div class="mail-text">` + escaped + `</div>`
}

func uniquePositiveInts(values []int) []int {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[int]struct{}, len(values))
	result := make([]int, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}

	return result
}
