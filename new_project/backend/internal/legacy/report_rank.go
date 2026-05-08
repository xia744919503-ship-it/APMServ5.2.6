package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const legacyPageSize = 10

var (
	reportTagPattern        = regexp.MustCompile(`(?s)<[^>]*>`)
	reportWhitespacePattern = regexp.MustCompile(`\s+`)
)

type rankingSpec struct {
	Kind    string
	Title   string
	Table   string
	Query   string
	Columns []RankingColumn
}

func (r *Repository) Reports(ctx context.Context, uid int, filter string, page int) (ReportPage, error) {
	filter = normalizeReportFilter(filter)
	if r.db == nil {
		return ReportPage{
			Filter: filter,
			Items:  []ReportSummary{},
		}, nil
	}

	whereClause, args := reportWhere(uid, filter)
	var total int
	if err := r.db.QueryRowContext(ctx, "select count(*) from sys_report where "+whereClause, args...).Scan(&total); err != nil {
		return ReportPage{}, err
	}

	pageCount := 0
	if total > 0 {
		pageCount = (total + legacyPageSize - 1) / legacyPageSize
		page = clamp(page, 0, pageCount-1)
	} else {
		page = 0
	}

	items := make([]ReportSummary, 0, legacyPageSize)
	if total > 0 {
		queryArgs := append(append([]any{}, args...), page*legacyPageSize, legacyPageSize)
		rows, err := r.db.QueryContext(ctx, `
select
	id,
	origincid,
	origincity,
	happencid,
	happencity,
	title,
	type,
	`+"`time`"+`,
	`+"`read`"+`,
	battleid
from sys_report
where `+whereClause+`
order by id desc
limit ?, ?`, queryArgs...)
		if err != nil {
			return ReportPage{}, err
		}
		defer rows.Close()

		for rows.Next() {
			var summary ReportSummary
			var createdUnix int64
			var readFlag int
			if err := rows.Scan(
				&summary.ID,
				&summary.OriginCID,
				&summary.OriginCity,
				&summary.HappenCID,
				&summary.HappenCity,
				&summary.Title,
				&summary.Type,
				&createdUnix,
				&readFlag,
				&summary.BattleID,
			); err != nil {
				return ReportPage{}, err
			}

			createdAt := time.Unix(createdUnix, 0)
			summary.Read = readFlag != 0
			summary.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
			summary.Headline = reportDisplayHeadline(summary)

			body, err := r.readReportBody(ctx, uid, summary.ID, createdAt)
			if err == nil {
				summary.Snippet = reportSnippet(body)
			}
			if summary.Snippet == "" {
				summary.Snippet = formatReportLocation(summary)
			}

			items = append(items, summary)
		}

		if err := rows.Err(); err != nil {
			return ReportPage{}, err
		}
	}

	return ReportPage{
		Filter:    filter,
		Page:      page,
		PageCount: pageCount,
		Total:     total,
		Items:     items,
	}, nil
}

func (r *Repository) ReportDetail(ctx context.Context, uid int, id int) (ReportDetail, error) {
	if r.db == nil {
		return ReportDetail{}, sql.ErrNoRows
	}

	var summary ReportSummary
	var createdUnix int64
	var readFlag int
	err := r.db.QueryRowContext(ctx, `
select
	id,
	origincid,
	origincity,
	happencid,
	happencity,
	title,
	type,
	`+"`time`"+`,
	`+"`read`"+`,
	battleid
from sys_report
where uid = ? and id = ? and state = 0`, uid, id).Scan(
		&summary.ID,
		&summary.OriginCID,
		&summary.OriginCity,
		&summary.HappenCID,
		&summary.HappenCity,
		&summary.Title,
		&summary.Type,
		&createdUnix,
		&readFlag,
		&summary.BattleID,
	)
	if err != nil {
		return ReportDetail{}, err
	}

	createdAt := time.Unix(createdUnix, 0)
	summary.Read = readFlag != 0
	summary.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
	summary.Headline = reportDisplayHeadline(summary)

	body, err := r.readReportBody(ctx, uid, id, createdAt)
	if err != nil {
		return ReportDetail{}, err
	}
	summary.Snippet = reportSnippet(body)
	if summary.Snippet == "" {
		summary.Snippet = formatReportLocation(summary)
	}
	if err := r.markReportRead(ctx, uid, summary.ID, createdAt, body, summary.Read); err != nil {
		return ReportDetail{}, err
	}
	summary.Read = true

	return ReportDetail{
		Summary:      summary,
		HTMLDocument: wrapReportHTML(summary.Headline, body),
	}, nil
}

func (r *Repository) Ranking(ctx context.Context, kind string, page int) (RankingPage, error) {
	kind = strings.ToLower(strings.TrimSpace(kind))
	if kind == "" {
		kind = "user"
	}

	specs := rankingSpecs()
	spec, ok := specs[kind]
	if !ok {
		spec = specs["user"]
	}

	if r.db == nil {
		return RankingPage{
			Kind:    spec.Kind,
			Title:   spec.Title,
			Columns: spec.Columns,
			Rows:    []map[string]string{},
		}, nil
	}

	var total int
	if err := r.db.QueryRowContext(ctx, "select count(*) from `"+spec.Table+"`").Scan(&total); err != nil {
		return RankingPage{}, err
	}

	pageCount := 0
	if total > 0 {
		pageCount = (total + legacyPageSize - 1) / legacyPageSize
		page = clamp(page, 0, pageCount-1)
	} else {
		page = 0
	}

	updatedAt := time.Now().Format("2006-01-02 15:04:05")
	_ = r.db.QueryRowContext(ctx, "select from_unixtime(`value`, '%Y-%m-%d %H:%i:%s') from mem_state where `state` = 10").Scan(&updatedAt)

	rowsData := make([]map[string]string, 0, legacyPageSize)
	if total > 0 {
		rows, err := r.db.QueryContext(ctx, spec.Query, page*legacyPageSize, legacyPageSize)
		if err != nil {
			return RankingPage{}, err
		}
		defer rows.Close()

		rowsData, err = scanRankingRows(rows)
		if err != nil {
			return RankingPage{}, err
		}
	}

	return RankingPage{
		Kind:      spec.Kind,
		Title:     spec.Title,
		UpdatedAt: updatedAt,
		Page:      page,
		PageCount: pageCount,
		Total:     total,
		Columns:   spec.Columns,
		Rows:      rowsData,
	}, nil
}

func reportWhere(uid int, filter string) (string, []any) {
	switch filter {
	case "type0":
		return "uid = ? and type = 0 and state = 0", []any{uid}
	case "type1":
		return "uid = ? and type = 1 and state = 0", []any{uid}
	case "type2":
		return "uid = ? and type = 2 and state = 0", []any{uid}
	case "type3":
		return "uid = ? and type = 3 and state = 0", []any{uid}
	default:
		return "uid = ? and `read` = 0 and state = 0", []any{uid}
	}
}

func normalizeReportFilter(filter string) string {
	switch filter {
	case "type0", "type1", "type2", "type3":
		return filter
	default:
		return "unread"
	}
}

func rankingSpecs() map[string]rankingSpec {
	return map[string]rankingSpec{
		"user": {
			Kind:  "user",
			Title: "君主排行",
			Table: "rank_user",
			Query: userRankingQuery(),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "君主"},
				{Key: "nobility", Label: "爵位"},
				{Key: "prestige", Label: "声望"},
				{Key: "city", Label: "城池"},
				{Key: "people", Label: "人口"},
				{Key: "union_name", Label: "联盟"},
			},
		},
		"union": {
			Kind:  "union",
			Title: "联盟排行",
			Table: "rank_union",
			Query: unionRankingQuery(),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "联盟"},
				{Key: "leader", Label: "盟主"},
				{Key: "member", Label: "成员"},
				{Key: "famouscity", Label: "名城"},
				{Key: "prestige", Label: "声望"},
			},
		},
		"hero_level": {
			Kind:  "hero_level",
			Title: "将领等级排行",
			Table: "rank_hero",
			Query: heroRankingQuery("rank_hero"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "将领"},
				{Key: "owner_name", Label: "归属"},
				{Key: "level", Label: "等级"},
				{Key: "affairs", Label: "内政"},
				{Key: "bravery", Label: "勇武"},
				{Key: "wisdom", Label: "智谋"},
			},
		},
		"hero_affairs": {
			Kind:  "hero_affairs",
			Title: "将领内政排行",
			Table: "rank_hero_affairs",
			Query: heroRankingQuery("rank_hero_affairs"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "将领"},
				{Key: "owner_name", Label: "归属"},
				{Key: "level", Label: "等级"},
				{Key: "affairs", Label: "内政"},
				{Key: "bravery", Label: "勇武"},
				{Key: "wisdom", Label: "智谋"},
			},
		},
		"hero_bravery": {
			Kind:  "hero_bravery",
			Title: "将领勇武排行",
			Table: "rank_hero_bravery",
			Query: heroRankingQuery("rank_hero_bravery"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "将领"},
				{Key: "owner_name", Label: "归属"},
				{Key: "level", Label: "等级"},
				{Key: "affairs", Label: "内政"},
				{Key: "bravery", Label: "勇武"},
				{Key: "wisdom", Label: "智谋"},
			},
		},
		"hero_wisdom": {
			Kind:  "hero_wisdom",
			Title: "将领智谋排行",
			Table: "rank_hero_wisdom",
			Query: heroRankingQuery("rank_hero_wisdom"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "将领"},
				{Key: "owner_name", Label: "归属"},
				{Key: "level", Label: "等级"},
				{Key: "affairs", Label: "内政"},
				{Key: "bravery", Label: "勇武"},
				{Key: "wisdom", Label: "智谋"},
			},
		},
		"jungong": {
			Kind:  "jungong",
			Title: "军功排行",
			Table: "rank_jungong",
			Query: meritRankingQuery("rank_jungong"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "君主"},
				{Key: "union_name", Label: "联盟"},
				{Key: "jungong", Label: "军功"},
				{Key: "juanxian", Label: "捐献"},
				{Key: "qinwang", Label: "勤王"},
				{Key: "gongpin", Label: "贡品"},
			},
		},
		"juanxian": {
			Kind:  "juanxian",
			Title: "捐献排行",
			Table: "rank_juanxian",
			Query: meritRankingQuery("rank_juanxian"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "君主"},
				{Key: "union_name", Label: "联盟"},
				{Key: "jungong", Label: "军功"},
				{Key: "juanxian", Label: "捐献"},
				{Key: "qinwang", Label: "勤王"},
				{Key: "gongpin", Label: "贡品"},
			},
		},
		"qinwang": {
			Kind:  "qinwang",
			Title: "勤王排行",
			Table: "rank_qinwang",
			Query: meritRankingQuery("rank_qinwang"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "君主"},
				{Key: "union_name", Label: "联盟"},
				{Key: "jungong", Label: "军功"},
				{Key: "juanxian", Label: "捐献"},
				{Key: "qinwang", Label: "勤王"},
				{Key: "gongpin", Label: "贡品"},
			},
		},
		"gongpin": {
			Kind:  "gongpin",
			Title: "贡品排行",
			Table: "rank_gongpin",
			Query: meritRankingQuery("rank_gongpin"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "君主"},
				{Key: "union_name", Label: "联盟"},
				{Key: "jungong", Label: "军功"},
				{Key: "juanxian", Label: "捐献"},
				{Key: "qinwang", Label: "勤王"},
				{Key: "gongpin", Label: "贡品"},
			},
		},
		"jungong_union": {
			Kind:  "jungong_union",
			Title: "联盟军功排行",
			Table: "rank_jungong_union",
			Query: meritUnionRankingQuery("rank_jungong_union"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "联盟"},
				{Key: "leader", Label: "盟主"},
				{Key: "jungong", Label: "军功"},
				{Key: "juanxian", Label: "捐献"},
				{Key: "qinwang", Label: "勤王"},
				{Key: "gongpin", Label: "贡品"},
			},
		},
		"juanxian_union": {
			Kind:  "juanxian_union",
			Title: "联盟捐献排行",
			Table: "rank_juanxian_union",
			Query: meritUnionRankingQuery("rank_juanxian_union"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "联盟"},
				{Key: "leader", Label: "盟主"},
				{Key: "jungong", Label: "军功"},
				{Key: "juanxian", Label: "捐献"},
				{Key: "qinwang", Label: "勤王"},
				{Key: "gongpin", Label: "贡品"},
			},
		},
		"qinwang_union": {
			Kind:  "qinwang_union",
			Title: "联盟勤王排行",
			Table: "rank_qinwang_union",
			Query: meritUnionRankingQuery("rank_qinwang_union"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "联盟"},
				{Key: "leader", Label: "盟主"},
				{Key: "jungong", Label: "军功"},
				{Key: "juanxian", Label: "捐献"},
				{Key: "qinwang", Label: "勤王"},
				{Key: "gongpin", Label: "贡品"},
			},
		},
		"gongpin_union": {
			Kind:  "gongpin_union",
			Title: "联盟贡品排行",
			Table: "rank_gongpin_union",
			Query: meritUnionRankingQuery("rank_gongpin_union"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "联盟"},
				{Key: "leader", Label: "盟主"},
				{Key: "jungong", Label: "军功"},
				{Key: "juanxian", Label: "捐献"},
				{Key: "qinwang", Label: "勤王"},
				{Key: "gongpin", Label: "贡品"},
			},
		},
		"city_people": {
			Kind:  "city_people",
			Title: "城池人口排行",
			Table: "rank_city",
			Query: cityRankingQuery("rank_city"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "城池"},
				{Key: "city_type", Label: "类型"},
				{Key: "owner_name", Label: "君主"},
				{Key: "union_name", Label: "联盟"},
				{Key: "coord", Label: "坐标"},
				{Key: "people", Label: "人口"},
			},
		},
		"city_type": {
			Kind:  "city_type",
			Title: "城池类型排行",
			Table: "rank_city_type",
			Query: cityRankingQuery("rank_city_type"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "城池"},
				{Key: "city_type", Label: "类型"},
				{Key: "owner_name", Label: "君主"},
				{Key: "union_name", Label: "联盟"},
				{Key: "coord", Label: "坐标"},
				{Key: "people", Label: "人口"},
			},
		},
		"military": {
			Kind:  "military",
			Title: "兵力排行",
			Table: "rank_military",
			Query: militaryRankingQuery("rank_military", false, false),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "君主"},
				{Key: "union_name", Label: "联盟"},
				{Key: "army", Label: "兵力"},
				{Key: "attack", Label: "攻击"},
				{Key: "defence", Label: "防御"},
			},
		},
		"military_attack": {
			Kind:  "military_attack",
			Title: "攻击排行",
			Table: "rank_military_attack",
			Query: militaryRankingQuery("rank_military_attack", true, false),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "君主"},
				{Key: "union_name", Label: "联盟"},
				{Key: "attack", Label: "攻击"},
				{Key: "defence", Label: "防御"},
				{Key: "army", Label: "兵力"},
			},
		},
		"military_defence": {
			Kind:  "military_defence",
			Title: "防御排行",
			Table: "rank_military_defence",
			Query: militaryRankingQuery("rank_military_defence", false, true),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "君主"},
				{Key: "union_name", Label: "联盟"},
				{Key: "defence", Label: "防御"},
				{Key: "attack", Label: "攻击"},
				{Key: "army", Label: "兵力"},
			},
		},
		"rich": {
			Kind:  "rich",
			Title: "财富排行",
			Table: "rank_rich",
			Query: richRankingQuery("rank_rich"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "君主"},
				{Key: "union_name", Label: "联盟"},
				{Key: "total", Label: "总资产"},
				{Key: "day", Label: "日收益"},
				{Key: "month", Label: "月收益"},
			},
		},
		"rich_day": {
			Kind:  "rich_day",
			Title: "日收益排行",
			Table: "rank_rich_day",
			Query: richRankingQuery("rank_rich_day"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "君主"},
				{Key: "union_name", Label: "联盟"},
				{Key: "day", Label: "日收益"},
				{Key: "total", Label: "总资产"},
				{Key: "month", Label: "月收益"},
			},
		},
		"rich_month": {
			Kind:  "rich_month",
			Title: "月收益排行",
			Table: "rank_rich_month",
			Query: richRankingQuery("rank_rich_month"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "君主"},
				{Key: "union_name", Label: "联盟"},
				{Key: "month", Label: "月收益"},
				{Key: "total", Label: "总资产"},
				{Key: "day", Label: "日收益"},
			},
		},
		"battle_total": {
			Kind:  "battle_total",
			Title: "功勋总榜",
			Table: "rank_battle_total",
			Query: battleRankingQuery("rank_battle_total"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "君主"},
				{Key: "union_name", Label: "联盟"},
				{Key: "total", Label: "总功勋"},
				{Key: "week", Label: "周功勋"},
				{Key: "day", Label: "日功勋"},
			},
		},
		"battle_week": {
			Kind:  "battle_week",
			Title: "功勋周榜",
			Table: "rank_battle_week",
			Query: battleRankingQuery("rank_battle_week"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "君主"},
				{Key: "union_name", Label: "联盟"},
				{Key: "week", Label: "周功勋"},
				{Key: "total", Label: "总功勋"},
				{Key: "day", Label: "日功勋"},
			},
		},
		"battle_day": {
			Kind:  "battle_day",
			Title: "功勋日榜",
			Table: "rank_battle_day",
			Query: battleRankingQuery("rank_battle_day"),
			Columns: []RankingColumn{
				{Key: "rank", Label: "排名"},
				{Key: "name", Label: "君主"},
				{Key: "union_name", Label: "联盟"},
				{Key: "day", Label: "日功勋"},
				{Key: "week", Label: "周功勋"},
				{Key: "total", Label: "总功勋"},
			},
		},
	}
}

func userRankingQuery() string {
	return `
select
	` + "`rank`" + ` as rank,
	` + "`name`" + ` as name,
	` + "`nobility`" + ` as nobility,
	` + "`prestige`" + ` as prestige,
	` + "`city`" + ` as city,
	` + "`people`" + ` as people,
	coalesce(` + "`union`" + `, '') as union_name
from rank_user
where ` + "`rank`" + ` > ?
order by ` + "`rank`" + `
limit ?`
}

func unionRankingQuery() string {
	return `
select
	` + "`rank`" + ` as rank,
	` + "`name`" + ` as name,
	coalesce(` + "`leader`" + `, '') as leader,
	` + "`member`" + ` as member,
	coalesce(` + "`famouscity`" + `, '') as famouscity,
	` + "`prestige`" + ` as prestige
from rank_union
where ` + "`rank`" + ` > ?
order by ` + "`rank`" + `
limit ?`
}

func heroRankingQuery(table string) string {
	return `
select
	` + "`rank`" + ` as rank,
	` + "`name`" + ` as name,
	coalesce(` + "`user`" + `, '') as owner_name,
	` + "`level`" + ` as level,
	` + "`affairs`" + ` as affairs,
	` + "`bravery`" + ` as bravery,
	` + "`wisdom`" + ` as wisdom
from ` + table + `
where ` + "`rank`" + ` > ?
order by ` + "`rank`" + `
limit ?`
}

func meritRankingQuery(table string) string {
	return `
select
	` + "`rank`" + ` as rank,
	` + "`name`" + ` as name,
	coalesce(` + "`union`" + `, '') as union_name,
	` + "`jungong`" + ` as jungong,
	` + "`juanxian`" + ` as juanxian,
	` + "`qinwang`" + ` as qinwang,
	` + "`gongpin`" + ` as gongpin
from ` + table + `
where ` + "`rank`" + ` > ?
order by ` + "`rank`" + `
limit ?`
}

func meritUnionRankingQuery(table string) string {
	return `
select
	` + "`rank`" + ` as rank,
	` + "`name`" + ` as name,
	coalesce(` + "`leader`" + `, '') as leader,
	` + "`jungong`" + ` as jungong,
	` + "`juanxian`" + ` as juanxian,
	` + "`qinwang`" + ` as qinwang,
	` + "`gongpin`" + ` as gongpin
from ` + table + `
where ` + "`rank`" + ` > ?
order by ` + "`rank`" + `
limit ?`
}

func cityRankingQuery(table string) string {
	return `
select
	` + "`rank`" + ` as rank,
	` + "`name`" + ` as name,
	` + "`type`" + ` as city_type,
	coalesce(` + "`user`" + `, '') as owner_name,
	coalesce(` + "`union`" + `, '') as union_name,
	concat('[', mod(` + "`cid`" + `, 1000), ',', floor(` + "`cid`" + ` / 1000), ']') as coord,
	` + "`people`" + ` as people
from ` + table + `
where ` + "`rank`" + ` > ?
order by ` + "`rank`" + `
limit ?`
}

func militaryRankingQuery(table string, attackFirst bool, defenceFirst bool) string {
	mainStat := "`army` as army,\n\t`attack` as attack,\n\t`defence` as defence"
	if attackFirst {
		mainStat = "`attack` as attack,\n\t`defence` as defence,\n\t`army` as army"
	}
	if defenceFirst {
		mainStat = "`defence` as defence,\n\t`attack` as attack,\n\t`army` as army"
	}

	return `
select
	` + "`rank`" + ` as rank,
	` + "`name`" + ` as name,
	coalesce(` + "`unionname`" + `, '') as union_name,
	` + mainStat + `
from ` + table + `
where ` + "`rank`" + ` > ?
order by ` + "`rank`" + `
limit ?`
}

func richRankingQuery(table string) string {
	return `
select
	` + "`rank`" + ` as rank,
	` + "`name`" + ` as name,
	coalesce(` + "`unionname`" + `, '') as union_name,
	` + "`total`" + ` as total,
	` + "`day`" + ` as day,
	` + "`month`" + ` as month
from ` + table + `
where ` + "`rank`" + ` > ?
order by ` + "`rank`" + `
limit ?`
}

func battleRankingQuery(table string) string {
	return `
select
	` + "`rank`" + ` as rank,
	` + "`name`" + ` as name,
	coalesce(` + "`unionname`" + `, '') as union_name,
	` + "`total`" + ` as total,
	` + "`week`" + ` as week,
	` + "`day`" + ` as day
from ` + table + `
where ` + "`rank`" + ` > ?
order by ` + "`rank`" + `
limit ?`
}

func scanRankingRows(rows *sql.Rows) ([]map[string]string, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]string, 0, legacyPageSize)
	for rows.Next() {
		values := make([]sql.RawBytes, len(columns))
		scanTargets := make([]any, len(columns))
		for index := range values {
			scanTargets[index] = &values[index]
		}

		if err := rows.Scan(scanTargets...); err != nil {
			return nil, err
		}

		row := make(map[string]string, len(columns))
		for index, column := range columns {
			row[column] = string(values[index])
		}
		result = append(result, row)
	}

	return result, rows.Err()
}

func (r *Repository) readReportBody(ctx context.Context, uid int, id int, createdAt time.Time) (string, error) {
	body, err := readReportCache(id, createdAt)
	if err == nil && strings.TrimSpace(body) != "" {
		return body, nil
	}

	var content sql.NullString
	if err := r.db.QueryRowContext(ctx, "select content from sys_report where uid = ? and id = ?", uid, id).Scan(&content); err != nil {
		return "", err
	}

	return content.String, nil
}

func (r *Repository) markReportRead(ctx context.Context, uid int, id int, createdAt time.Time, body string, alreadyRead bool) error {
	clearContent := false
	if strings.TrimSpace(body) != "" {
		if err := ensureReportCache(id, createdAt, body); err == nil {
			clearContent = true
		}
	}

	if !alreadyRead {
		query := "update sys_report set `read` = 1 where uid = ? and id = ?"
		if clearContent {
			query = "update sys_report set `read` = 1, content = '' where uid = ? and id = ?"
		}
		if _, err := r.db.ExecContext(ctx, query, uid, id); err != nil {
			return err
		}
	}

	return r.clearReportAlarmIfNoUnread(ctx, uid)
}

func (r *Repository) clearReportAlarmIfNoUnread(ctx context.Context, uid int) error {
	var unread int
	if err := r.db.QueryRowContext(ctx, "select count(*) from sys_report where uid = ? and `read` = 0 and state = 0", uid).Scan(&unread); err != nil {
		return err
	}
	if unread > 0 {
		return nil
	}

	_, err := r.db.ExecContext(ctx, "update sys_alarm set report = 0 where uid = ?", uid)
	return err
}

func readReportCache(id int, createdAt time.Time) (string, error) {
	if createdAt.IsZero() {
		return "", os.ErrNotExist
	}

	reportPath := legacyReportPath(id, createdAt)
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func ensureReportCache(id int, createdAt time.Time, body string) error {
	if createdAt.IsZero() || strings.TrimSpace(body) == "" {
		return nil
	}

	reportPath := legacyReportPath(id, createdAt)
	if _, err := os.Stat(reportPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		return err
	}

	return os.WriteFile(reportPath, []byte(body), 0o644)
}

func legacyReportPath(id int, createdAt time.Time) string {
	return filepath.Join(legacyGameDirectory(), "report_data", createdAt.Format("20060102"), fmt.Sprintf("%d.html", id))
}

func legacyGameDirectory() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("www", "htdocs", "server", "game")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "www", "htdocs", "server", "game"))
}

func formatReportHeadline(summary ReportSummary) string {
	location := formatReportLocation(summary)
	if location == "" {
		return fmt.Sprintf("战报 #%d", summary.ID)
	}

	return fmt.Sprintf("战报 #%d · %s", summary.ID, location)
}

func formatReportLocation(summary ReportSummary) string {
	origin := strings.TrimSpace(summary.OriginCity)
	happen := strings.TrimSpace(summary.HappenCity)

	switch {
	case origin != "" && happen != "" && origin != happen:
		return origin + " -> " + happen
	case happen != "":
		return happen
	case origin != "":
		return origin
	default:
		return ""
	}
}

func reportSnippet(body string) string {
	plain := reportWhitespacePattern.ReplaceAllString(reportTagPattern.ReplaceAllString(body, " "), " ")
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

func wrapReportHTML(title string, body string) string {
	return `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<title>` + html.EscapeString(title) + `</title>
<style type="text/css">
html {
	scrollbar-arrow-color: #000000;
	scrollbar-base-color: #17292B;
	scrollbar-dark-shadow-color: #000000;
	scrollbar-track-color: #1A2B2F;
	scrollbar-face-color: #17292B;
	scrollbar-shadow-color: #000000;
	scrollbar-highlight-color: #000000;
	scrollbar-3d-light-color: #17292B;
}
body {
	background:
		radial-gradient(circle at top, #31545a 0%, #17292b 42%, #081012 100%);
	font-size: 9pt;
	color: #FFFFFF;
	margin: 0;
	padding: 12px;
	font-family: SimSun, "Microsoft YaHei", serif;
}
body, td, th {
	color: #FFFFFF;
}
table {
	border-collapse: collapse;
}
.NormalText {
	font-size: 12px;
	font-weight: normal;
	color: #FFFFFF;
}
.WinGreen {
	font-size: 14px;
	font-weight: bold;
	color: #00FF00;
	background: #225A5D;
}
.LoseRed {
	font-size: 14px;
	color: #CC3333;
	font-weight: bold;
	background: #225A5D;
}
.TitleBlueWhite {
	font-size: 14px;
	color: #FFFFFF;
	font-weight: bold;
	background: #225A5D;
}
.TitleRedWhite {
	font-size: 14px;
	color: #FFFFFF;
	font-weight: bold;
	background: #5D2522;
}
.TitleListWhite {
	font-size: 14px;
	color: #FFFFFF;
	font-weight: bold;
	background: #17292B;
}
.NameBlue {
	font-size: 12px;
	font-weight: normal;
	color: #00D8FF;
}
.TitleBattleYellow {
	font-size: 12px;
	font-weight: bold;
	color: #FFD200;
	background: #2D2414;
}
.TextArmyCount {
	font-size: 12px;
	font-weight: normal;
	color: #FFFFFF;
	background: #000000;
}
.report-screen {
	max-width: 620px;
	margin: 0 auto;
	padding: 10px;
	border: 1px solid #6f877f;
	background: rgba(5, 11, 13, 0.78);
	box-shadow:
		inset 0 0 0 1px rgba(255, 255, 255, 0.08),
		0 14px 30px rgba(0, 0, 0, 0.35);
}
.report-stage {
	border: 1px solid #24373b;
	background:
		linear-gradient(180deg, rgba(54, 95, 100, 0.2), rgba(5, 10, 11, 0.36)),
		#0d1719;
}
.report-stage-title {
	padding: 10px 14px;
	border-bottom: 1px solid #36555a;
	background: linear-gradient(180deg, #30585d 0%, #21393c 100%);
	box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.18);
}
.report-stage-title strong {
	display: block;
	font-size: 16px;
	letter-spacing: 1px;
	color: #f8efbe;
}
.report-stage-title span {
	display: block;
	margin-top: 4px;
	font-size: 12px;
	color: #d1d9c8;
}
.report-stage-body {
	padding: 12px;
}
.report-shell {
	display: grid;
	gap: 10px;
}
.report-head {
	display: flex;
	justify-content: space-between;
	align-items: center;
	gap: 12px;
	padding: 10px 12px;
	border: 1px solid #395257;
	background: linear-gradient(180deg, rgba(43, 75, 79, 0.95), rgba(19, 31, 35, 0.95));
}
.report-head-copy strong {
	display: block;
	font-size: 18px;
	color: #fff4c9;
}
.report-head-copy small,
.report-head-kicker {
	display: block;
	font-size: 12px;
	color: #d9e0d1;
}
.report-head-kicker {
	margin-bottom: 4px;
	letter-spacing: 2px;
	color: #89c0c7;
}
.report-outcome {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	min-width: 104px;
	padding: 7px 14px;
	border: 1px solid rgba(255, 255, 255, 0.12);
	font-size: 14px;
	font-weight: bold;
	letter-spacing: 1px;
}
.report-outcome-win {
	background: linear-gradient(180deg, #386b40 0%, #1f3c24 100%);
	color: #dff9bf;
}
.report-outcome-alert {
	background: linear-gradient(180deg, #734237 0%, #451f19 100%);
	color: #ffd8bf;
}
.report-note {
	padding: 10px 12px;
	border: 1px solid #473d28;
	background: linear-gradient(180deg, rgba(67, 52, 26, 0.92), rgba(31, 22, 11, 0.92));
	color: #f2df9d;
	line-height: 1.7;
}
.report-summary,
.report-grid {
	display: grid;
	gap: 10px;
}
.report-grid {
	grid-template-columns: repeat(2, minmax(0, 1fr));
}
.report-section {
	border: 1px solid #2f4448;
	background: rgba(5, 11, 12, 0.72);
}
.report-section-title {
	padding: 8px 10px;
	border-bottom: 1px solid #334b50;
	background: linear-gradient(180deg, #244146 0%, #13272b 100%);
	color: #f0df98;
	font-size: 13px;
	font-weight: bold;
}
.report-section-body {
	padding: 8px;
}
.report-summary-table,
.report-section table {
	width: 100%;
	background: #fff;
}
.report-summary-table td,
.report-section td {
	border: 1px solid #fff;
}
@media (max-width: 560px) {
	.report-grid {
		grid-template-columns: 1fr;
	}
}
</style>
</head>
<body><div class="report-screen"><div class="report-stage"><div class="report-stage-title"><strong>` + html.EscapeString(title) + `</strong><span>热血三国军情回执</span></div><div class="report-stage-body">` + body + `</div></div></div></body>
</html>`
}
