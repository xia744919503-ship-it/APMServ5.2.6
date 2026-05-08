package legacy

import (
	"context"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	portalFunctionPattern = regexp.MustCompile(`(?s)function\s+(\w+)\s*\(\)\s*\{\s*return\s+"((?:\\.|[^"])*)";\s*\}`)
	portalHeadingPattern  = regexp.MustCompile(`(?s)<font[^>]*><b>(.*?)</b></font>`)
	portalHTMLTagPattern  = regexp.MustCompile(`(?s)<[^>]+>`)
	portalMailPattern     = regexp.MustCompile(`mailto:([^'"]+)`)
	portalTextPattern     = regexp.MustCompile(`(?is)<%s[^>]*>(.*?)</%s>`)
	portalParagraphs      = regexp.MustCompile(`(?is)<p[^>]*>(.*?)</p>`)
)

func (r *Repository) LegacyPortal(ctx context.Context) (PortalSnapshot, error) {
	snapshot := defaultPortalSnapshot()
	passUpdatedAt := ""

	if raw, modTime := readLegacyPortalFile("pass.js"); raw != "" {
		passUpdatedAt = formatPortalTime(modTime)
		applyPassJS(&snapshot, raw, passUpdatedAt)
	}

	if raw, modTime := readLegacyPortalFile("index_notice.html"); raw != "" {
		applyNoticeHTML(&snapshot, raw, formatPortalTime(modTime))
	}

	if announcement, err := r.Announcement(ctx); err == nil {
		lines := splitPortalLines(announcement)
		if len(lines) > 0 {
			snapshot.AnnouncementLines = lines
		}
	}

	if len(snapshot.AnnouncementLines) == 0 {
		snapshot.AnnouncementLines = splitPortalLines(snapshot.Notice.Body)
	}

	snapshot.Boards = buildPortalBoards(snapshot)
	snapshot.Topics = buildPortalTopics(snapshot, passUpdatedAt)
	return snapshot, nil
}

func defaultPortalSnapshot() PortalSnapshot {
	return PortalSnapshot{
		PassType:     "12ha",
		HomeButton:   "专区首页",
		SupportEmail: "webgamekf@12ha.com",
		Notice: PortalNotice{
			Title:       "热血公告",
			Greeting:    "尊敬的玩家：",
			Body:        "军机厅正在整理旧版门页告示，请稍候翻阅。",
			Signature:   "热血三国运营团队",
			SourceLabel: "index_notice.html",
			SourceURL:   "http://rx.12ha.com",
		},
		Links: []PortalLink{
			{ID: "home", Label: "专区首页", URL: "http://rx.12ha.com", Note: "旧版官网首页入口", Group: "portal"},
			{ID: "forum", Label: "主公论坛", URL: "http://wgbbs.12ha.com", Note: "旧版主公论坛入口", Group: "forum"},
			{ID: "union-forum", Label: "联盟论坛", URL: "http://rxbbs.12ha.com", Note: "旧版联盟论坛入口", Group: "forum"},
			{ID: "help", Label: "帮助页面", URL: "http://rx.12ha.com/News/detail_1990.html", Note: "旧版帮助与规则入口", Group: "help"},
			{ID: "register", Label: "注册账号", URL: "http://passport.12ha.com/GCP_Register.aspx", Note: "旧版通行证注册入口", Group: "account"},
			{ID: "pay", Label: "充值入口", URL: "http://passport.12ha.com/GCP_Return_RXSG.html?u=", Note: "旧版充值跳转地址", Group: "account"},
		},
	}
}

func readLegacyPortalFile(name string) (string, time.Time) {
	path := filepath.Join(legacyHTDocsDirectory(), name)
	body, err := os.ReadFile(path)
	if err != nil {
		return "", time.Time{}
	}

	info, err := os.Stat(path)
	if err != nil {
		return string(body), time.Time{}
	}

	return string(body), info.ModTime()
}

func applyPassJS(snapshot *PortalSnapshot, raw string, updatedAt string) {
	values := map[string]string{}
	for _, match := range portalFunctionPattern.FindAllStringSubmatch(raw, -1) {
		values[match[1]] = html.UnescapeString(strings.ReplaceAll(match[2], `\"`, `"`))
	}

	if value := strings.TrimSpace(values["getPassType"]); value != "" {
		snapshot.PassType = value
	}
	if value := strings.TrimSpace(values["getHomeButton"]); value != "" {
		snapshot.HomeButton = value
	}

	linkDefs := []struct {
		id       string
		label    string
		valueKey string
		note     string
		group    string
	}{
		{id: "home", label: defaultString(snapshot.HomeButton, "专区首页"), valueKey: "getHomeUrl", note: "旧版官网首页入口", group: "portal"},
		{id: "forum", label: "主公论坛", valueKey: "getBBSUrl", note: "旧版主公论坛入口", group: "forum"},
		{id: "union-forum", label: "联盟论坛", valueKey: "getUnionBBSUrl", note: "旧版联盟论坛入口", group: "forum"},
		{id: "help", label: "帮助页面", valueKey: "getHelpUrl", note: "旧版帮助与规则入口", group: "help"},
		{id: "register", label: "注册账号", valueKey: "getRegisterUrl", note: "旧版通行证注册入口", group: "account"},
		{id: "pay", label: "充值入口", valueKey: "getPayUrl", note: "旧版充值跳转地址", group: "account"},
	}

	links := make([]PortalLink, 0, len(linkDefs))
	for _, def := range linkDefs {
		url := strings.TrimSpace(values[def.valueKey])
		if url == "" {
			url = portalLinkURL(snapshot.Links, def.id)
		}
		if url == "" {
			continue
		}

		links = append(links, PortalLink{
			ID:    def.id,
			Label: def.label,
			URL:   url,
			Note:  def.note,
			Group: def.group,
		})
	}
	if len(links) > 0 {
		snapshot.Links = links
	}

	ruleHTML := strings.TrimSpace(values["getUserRule"])
	if ruleHTML == "" {
		return
	}

	snapshot.Rules = parsePortalRules(ruleHTML)
	if email := extractPortalMail(ruleHTML); email != "" {
		snapshot.SupportEmail = email
	}
	if snapshot.Notice.UpdatedAt == "" {
		snapshot.Notice.UpdatedAt = updatedAt
	}
}

func applyNoticeHTML(snapshot *PortalSnapshot, raw string, updatedAt string) {
	title := portalTagText(raw, "h1")
	greeting := portalTagText(raw, "h3")
	paragraphMatches := portalParagraphs.FindAllStringSubmatch(raw, -1)
	paragraphs := make([]string, 0, len(paragraphMatches))
	for _, match := range paragraphMatches {
		line := cleanPortalText(match[1])
		if line != "" {
			paragraphs = append(paragraphs, line)
		}
	}

	if title != "" {
		snapshot.Notice.Title = title
	}
	if greeting != "" {
		snapshot.Notice.Greeting = greeting
	}
	if len(paragraphs) > 0 {
		snapshot.Notice.Body = paragraphs[0]
	}
	if len(paragraphs) > 1 {
		snapshot.Notice.Signature = paragraphs[len(paragraphs)-1]
	}
	snapshot.Notice.SourceLabel = "index_notice.html"
	if snapshot.Notice.SourceURL == "" {
		snapshot.Notice.SourceURL = portalLinkURL(snapshot.Links, "home")
	}
	if updatedAt != "" {
		snapshot.Notice.UpdatedAt = updatedAt
	}
}

func parsePortalRules(raw string) []PortalRuleSection {
	matches := portalHeadingPattern.FindAllStringSubmatchIndex(raw, -1)
	sections := make([]PortalRuleSection, 0, len(matches))
	for index, match := range matches {
		start := match[1]
		end := len(raw)
		if index+1 < len(matches) {
			end = matches[index+1][0]
		}

		title := cleanPortalText(raw[match[2]:match[3]])
		body := raw[start:end]
		body = strings.ReplaceAll(body, "<br/>", "\n")
		body = strings.ReplaceAll(body, "<br />", "\n")
		body = strings.ReplaceAll(body, "<br>", "\n")
		body = cleanPortalText(body)
		lines := splitPortalLines(body)
		if title == "" || len(lines) == 0 {
			continue
		}

		sections = append(sections, PortalRuleSection{
			ID:    "rule-" + strconvItoa(index+1),
			Title: title,
			Items: lines,
		})
	}
	return sections
}

func extractPortalMail(raw string) string {
	match := portalMailPattern.FindStringSubmatch(raw)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func buildPortalBoards(snapshot PortalSnapshot) []PortalBoard {
	return []PortalBoard{
		{
			Key:    "announce",
			Label:  "官方公告",
			Keeper: "军机厅",
			Brief:  "门页告示与游戏公告",
			URL:    defaultString(snapshot.Notice.SourceURL, portalLinkURL(snapshot.Links, "home")),
		},
		{
			Key:    "help",
			Label:  "帮助指引",
			Keeper: "教习官",
			Brief:  "帮助页、客服与使用说明",
			URL:    portalLinkURL(snapshot.Links, "help"),
		},
		{
			Key:    "rule",
			Label:  "游戏规则",
			Keeper: "军法司",
			Brief:  "帐号、漏洞与处罚规则",
			URL:    portalLinkURL(snapshot.Links, "help"),
		},
		{
			Key:    "portal",
			Label:  "站点门廊",
			Keeper: "驿站吏",
			Brief:  "官网、论坛与账号入口",
			URL:    portalLinkURL(snapshot.Links, "home"),
		},
	}
}

func buildPortalTopics(snapshot PortalSnapshot, passUpdatedAt string) []PortalTopic {
	topics := []PortalTopic{}
	noticeUpdatedAt := defaultString(snapshot.Notice.UpdatedAt, passUpdatedAt)

	if strings.TrimSpace(snapshot.Notice.Body) != "" {
		content := joinPortalLines(snapshot.Notice.Greeting, snapshot.Notice.Body, snapshot.Notice.Signature)
		topics = append(topics, PortalTopic{
			ID:          "notice-main",
			BoardKey:    "announce",
			Title:       defaultString(snapshot.Notice.Title, "热血公告"),
			Summary:     summaryPortalLine(snapshot.Notice.Body),
			Content:     content,
			Author:      "运营团队",
			Role:        "门页告示",
			UpdatedAt:   noticeUpdatedAt,
			SourceLabel: defaultString(snapshot.Notice.SourceLabel, "index_notice.html"),
			SourceURL:   snapshot.Notice.SourceURL,
			Tags:        []string{"置顶", "门页"},
			Sticky:      true,
		})
	}

	for index, line := range snapshot.AnnouncementLines {
		topics = append(topics, PortalTopic{
			ID:          "announce-" + strconvItoa(index+1),
			BoardKey:    "announce",
			Title:       "【公告】" + summaryPortalLine(line),
			Summary:     line,
			Content:     line,
			Author:      "系统公告",
			Role:        "官帖",
			UpdatedAt:   noticeUpdatedAt,
			SourceLabel: "sys_announce",
			SourceURL:   portalLinkURL(snapshot.Links, "home"),
			Tags:        []string{"公告", stickyTag(index < 2)},
			Sticky:      index < 2,
		})
		if index >= 5 {
			break
		}
	}

	helpURL := portalLinkURL(snapshot.Links, "help")
	if helpURL != "" {
		topics = append(topics, PortalTopic{
			ID:          "help-entry",
			BoardKey:    "help",
			Title:       "【帮助】旧版帮助页面入口",
			Summary:     "保留原版帮助页地址，可对照查看旧版说明与规则页面。",
			Content:     "此入口对应旧版站外帮助页，可用于对照原版说明页与规则页。\n地址：" + helpURL,
			Author:      "教习官",
			Role:        "指引帖",
			UpdatedAt:   passUpdatedAt,
			SourceLabel: "pass.js",
			SourceURL:   helpURL,
			Tags:        []string{"置顶", "帮助"},
			Sticky:      true,
		})
	}

	if snapshot.SupportEmail != "" {
		topics = append(topics, PortalTopic{
			ID:          "help-support",
			BoardKey:    "help",
			Title:       "【客服】处罚申诉与帮助邮箱",
			Summary:     "旧版规则中保留了客服邮箱，可作为原版帮助链的一部分查看。",
			Content:     "客服帮助邮箱：\n" + snapshot.SupportEmail + "\n\n规则说明中保留了该联系方式，用于处罚申诉与帮助说明。",
			Author:      "军机厅",
			Role:        "客服帖",
			UpdatedAt:   passUpdatedAt,
			SourceLabel: "pass.js",
			SourceURL:   "mailto:" + snapshot.SupportEmail,
			Tags:        []string{"帮助", "客服"},
		})
	}

	for index, section := range snapshot.Rules {
		topics = append(topics, PortalTopic{
			ID:          section.ID,
			BoardKey:    "rule",
			Title:       "【规则】" + section.Title,
			Summary:     summaryPortalLine(strings.Join(section.Items, " ")),
			Content:     strings.Join(section.Items, "\n"),
			Author:      "军法司",
			Role:        "规则摘录",
			UpdatedAt:   passUpdatedAt,
			SourceLabel: "pass.js",
			SourceURL:   helpURL,
			Tags:        []string{"规则", stickyTag(index < 2)},
			Sticky:      index < 2,
		})
	}

	for _, link := range snapshot.Links {
		if link.Group == "help" {
			continue
		}

		topics = append(topics, PortalTopic{
			ID:          "portal-" + link.ID,
			BoardKey:    "portal",
			Title:       "【门廊】" + link.Label,
			Summary:     link.Note,
			Content:     link.Note + "\n地址：" + link.URL,
			Author:      "驿站吏",
			Role:        "门廊引路",
			UpdatedAt:   passUpdatedAt,
			SourceLabel: "pass.js",
			SourceURL:   link.URL,
			Tags:        []string{link.Group},
		})
	}

	return topics
}

func portalTagText(raw string, tag string) string {
	pattern := regexp.MustCompile(strings.ReplaceAll(portalTextPattern.String(), "%s", tag))
	match := pattern.FindStringSubmatch(raw)
	if len(match) < 2 {
		return ""
	}
	return cleanPortalText(match[1])
}

func cleanPortalText(raw string) string {
	text := portalHTMLTagPattern.ReplaceAllString(raw, "")
	text = html.UnescapeString(text)
	text = strings.ReplaceAll(text, "\u00a0", " ")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "\r", "")
	return strings.TrimSpace(text)
}

func splitPortalLines(raw string) []string {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	lines := make([]string, 0, 8)
	for _, item := range strings.Split(normalized, "\n") {
		line := strings.TrimSpace(item)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func joinPortalLines(lines ...string) string {
	items := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		items = append(items, line)
	}
	return strings.Join(items, "\n")
}

func summaryPortalLine(raw string) string {
	line := strings.TrimSpace(raw)
	if len([]rune(line)) <= 28 {
		return line
	}
	runes := []rune(line)
	return string(runes[:28]) + "…"
}

func portalLinkURL(links []PortalLink, id string) string {
	for _, link := range links {
		if link.ID == id {
			return link.URL
		}
	}
	return ""
}

func stickyTag(sticky bool) string {
	if sticky {
		return "置顶"
	}
	return ""
}

func formatPortalTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04")
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func legacyHTDocsDirectory() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("www", "htdocs")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "www", "htdocs"))
}

func strconvItoa(value int) string {
	return strconv.Itoa(value)
}
