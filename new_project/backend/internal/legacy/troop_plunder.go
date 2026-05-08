package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"html"
	"strings"
	"time"
)

const troopCarryTechnicID = 11

type plunderTargetSnapshot struct {
	UID       int
	CityName  string
	Soldiers  int64
	Defences  int64
	Stationed int64
	Heroes    int64
	Resource  TroopResource
	Protected TroopResource
}

func (s plunderTargetSnapshot) minimalEligible() bool {
	return s.Soldiers <= 0 && s.Defences <= 0 && s.Stationed <= 0 && s.Heroes <= 0
}

func (r *Repository) loadPlunderTargetSnapshotTx(ctx context.Context, tx *sql.Tx, cid int) (plunderTargetSnapshot, error) {
	var (
		snapshot       plunderTargetSnapshot
		woodMax        int64
		rockMax        int64
		ironMax        int64
		foodMax        int64
		goldMax        int64
		cityName       sql.NullString
		totalSoldiers  sql.NullInt64
		totalDefences  sql.NullInt64
		totalStationed sql.NullInt64
		totalHeroes    sql.NullInt64
	)

	if err := tx.QueryRowContext(ctx, `
select
	c.uid,
	coalesce(c.name, ''),
	m.wood,
	m.rock,
	m.iron,
	m.food,
	m.gold,
	m.wood_max,
	m.rock_max,
	m.iron_max,
	m.food_max,
	m.gold_max
from sys_city c
join mem_city_resource m on m.cid = c.cid
where c.cid = ?`, cid).Scan(
		&snapshot.UID,
		&cityName,
		&snapshot.Resource.Wood,
		&snapshot.Resource.Rock,
		&snapshot.Resource.Iron,
		&snapshot.Resource.Food,
		&snapshot.Resource.Gold,
		&woodMax,
		&rockMax,
		&ironMax,
		&foodMax,
		&goldMax,
	); err != nil {
		return plunderTargetSnapshot{}, err
	}
	snapshot.CityName = firstNonEmpty(cityName.String, formatCIDLabel(cid))
	snapshot.Protected = TroopResource{
		Wood: min64(snapshot.Resource.Wood, woodMax),
		Rock: min64(snapshot.Resource.Rock, rockMax),
		Iron: min64(snapshot.Resource.Iron, ironMax),
		Food: min64(snapshot.Resource.Food, foodMax),
		Gold: min64(snapshot.Resource.Gold, goldMax),
	}

	if err := tx.QueryRowContext(ctx, "select coalesce(sum(count), 0) from sys_city_soldier where cid = ? and count > 0", cid).Scan(&totalSoldiers); err != nil {
		return plunderTargetSnapshot{}, err
	}
	if err := tx.QueryRowContext(ctx, "select coalesce(sum(count), 0) from sys_city_defence where cid = ? and count > 0", cid).Scan(&totalDefences); err != nil {
		return plunderTargetSnapshot{}, err
	}
	if err := tx.QueryRowContext(ctx, "select count(*) from sys_troops where targetcid = ? and state = 4", cid).Scan(&totalStationed); err != nil {
		return plunderTargetSnapshot{}, err
	}
	if err := tx.QueryRowContext(ctx, "select count(*) from sys_city_hero where cid = ? and state in (1, 7)", cid).Scan(&totalHeroes); err != nil {
		return plunderTargetSnapshot{}, err
	}

	snapshot.Soldiers = totalSoldiers.Int64
	snapshot.Defences = totalDefences.Int64
	snapshot.Stationed = totalStationed.Int64
	snapshot.Heroes = totalHeroes.Int64
	return snapshot, nil
}

func (r *Repository) troopCarryCapacityTx(ctx context.Context, tx *sql.Tx, cid int, soldiers map[int]int64) (int64, error) {
	totalCarry := int64(0)
	for sid, count := range soldiers {
		if sid <= 0 || count <= 0 {
			continue
		}

		var carry sql.NullInt64
		if err := tx.QueryRowContext(ctx, "select carry from cfg_soldier where sid = ?", sid).Scan(&carry); err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return 0, err
		}
		totalCarry += carry.Int64 * count
	}

	var techLevel sql.NullInt64
	if err := tx.QueryRowContext(ctx, "select level from sys_city_technic where cid = ? and tid = ?", cid, troopCarryTechnicID).Scan(&techLevel); err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if techLevel.Int64 > 0 {
		totalCarry = totalCarry * (10 + techLevel.Int64) / 10
	}

	return totalCarry, nil
}

func computePlunderLoot(snapshot plunderTargetSnapshot, capacity int64) TroopResource {
	if capacity <= 0 {
		return TroopResource{}
	}

	loot := TroopResource{}
	remaining := capacity

	available := []struct {
		amount int64
		assign func(int64)
	}{
		{
			amount: max64(0, snapshot.Resource.Gold-snapshot.Protected.Gold),
			assign: func(value int64) { loot.Gold = value },
		},
		{
			amount: max64(0, snapshot.Resource.Food-snapshot.Protected.Food),
			assign: func(value int64) { loot.Food = value },
		},
		{
			amount: max64(0, snapshot.Resource.Wood-snapshot.Protected.Wood),
			assign: func(value int64) { loot.Wood = value },
		},
		{
			amount: max64(0, snapshot.Resource.Rock-snapshot.Protected.Rock),
			assign: func(value int64) { loot.Rock = value },
		},
		{
			amount: max64(0, snapshot.Resource.Iron-snapshot.Protected.Iron),
			assign: func(value int64) { loot.Iron = value },
		},
	}

	for _, item := range available {
		if remaining <= 0 {
			break
		}
		if item.amount <= 0 {
			continue
		}

		take := min64(remaining, item.amount)
		item.assign(take)
		remaining -= take
	}

	return loot
}

func (r *Repository) cityNameTx(ctx context.Context, tx *sql.Tx, cid int) string {
	if cid <= 0 {
		return "--"
	}

	var name sql.NullString
	if err := tx.QueryRowContext(ctx, "select name from sys_city where cid = ?", cid).Scan(&name); err != nil {
		return formatCIDLabel(cid)
	}
	return firstNonEmpty(name.String, formatCIDLabel(cid))
}

func (r *Repository) writeTroopReportTx(ctx context.Context, tx *sql.Tx, uid int, title int, originCID int, originCity string, happenCID int, happenCity string, content string) error {
	if uid <= 0 {
		return nil
	}

	postedAt := time.Now().Unix()
	if _, err := tx.ExecContext(ctx, `
insert into sys_report (uid, origincid, origincity, happencid, happencity, title, type, `+"`time`"+`, `+"`read`"+`, battleid, content, state)
values (?, ?, ?, ?, ?, ?, 0, ?, 0, 0, ?, 0)`,
		uid,
		originCID,
		originCity,
		happenCID,
		happenCity,
		title,
		postedAt,
		content,
	); err != nil {
		return err
	}

	_, err := tx.ExecContext(ctx, `
insert into sys_alarm (uid, report)
values (?, 1)
on duplicate key update report = 1`, uid)
	return err
}

func buildPlunderAttackerReport(originCity string, targetCity string, soldiers []Soldier, loot TroopResource, note string) string {
	var body strings.Builder
	body.WriteString(html.EscapeString(originCity))
	body.WriteString(" 向 ")
	body.WriteString(html.EscapeString(targetCity))
	body.WriteString(" 发起掠夺任务并已到达目标。<br/>")
	body.WriteString(html.EscapeString(note))
	body.WriteString("<br/>")
	body.WriteString(`<table width="567" border="0" cellpadding="1" cellspacing="1" bgcolor="#FFFFFF">`)
	body.WriteString(`<tr><td height="25" colspan="2" align="center" class="TitleBlueWhite">掠夺结算</td></tr>`)
	body.WriteString(`<tr><td height="25" colspan="2" align="center" class="TextArmyCount">`)
	body.WriteString(buildReportSoldierTable(soldiers))
	body.WriteString(`</td></tr>`)
	body.WriteString(`<tr><td height="25" colspan="2" align="center" class="TextArmyCount">`)
	body.WriteString(buildReportResourceTable("携带资源", loot))
	body.WriteString(`</td></tr>`)
	body.WriteString(`</table>`)
	return body.String()
}

func buildPlunderDefenderReport(originCity string, targetCity string, soldiers []Soldier, loot TroopResource, note string) string {
	var body strings.Builder
	body.WriteString("一支军队从 ")
	body.WriteString(html.EscapeString(originCity))
	body.WriteString(" 对 ")
	body.WriteString(html.EscapeString(targetCity))
	body.WriteString(" 发起了掠夺。<br/>")
	body.WriteString(html.EscapeString(note))
	body.WriteString("<br/>")
	body.WriteString(`<table width="567" border="0" cellpadding="1" cellspacing="1" bgcolor="#FFFFFF">`)
	body.WriteString(`<tr><td height="25" colspan="2" align="center" class="TitleBlueWhite">敌军情报</td></tr>`)
	body.WriteString(`<tr><td height="25" colspan="2" align="center" class="TextArmyCount">`)
	body.WriteString(buildReportSoldierTable(soldiers))
	body.WriteString(`</td></tr>`)
	body.WriteString(`<tr><td height="25" colspan="2" align="center" class="TextArmyCount">`)
	body.WriteString(buildReportResourceTable("损失资源", loot))
	body.WriteString(`</td></tr>`)
	body.WriteString(`</table>`)
	return body.String()
}

func buildReportSoldierTable(soldiers []Soldier) string {
	var body strings.Builder
	body.WriteString(`<table width="249" border="0" cellpadding="0" cellspacing="0">`)
	body.WriteString(`<tr class="TitleBattleYellow"><td width="120" height="25" align="center" valign="middle">军队</td><td width="129" height="25" align="center" valign="middle">数量</td></tr>`)
	for _, soldier := range soldiers {
		body.WriteString(`<tr><td width="120" height="25" align="center" valign="middle" class="TextArmyCount">`)
		body.WriteString(html.EscapeString(soldier.Name))
		body.WriteString(`</td><td width="129" height="25" align="center" valign="middle" class="TextArmyCount">`)
		body.WriteString(fmt.Sprintf("%d", soldier.Count))
		body.WriteString(`</td></tr>`)
	}
	body.WriteString(`</table>`)
	return body.String()
}

func buildReportResourceTable(title string, payload TroopResource) string {
	return `<table width="249" border="0" cellpadding="0" cellspacing="0">` +
		`<tr class="TitleBattleYellow"><td colspan="2" height="25" align="center" valign="middle">` + html.EscapeString(title) + `</td></tr>` +
		buildReportResourceRow("黄金", payload.Gold) +
		buildReportResourceRow("粮草", payload.Food) +
		buildReportResourceRow("木材", payload.Wood) +
		buildReportResourceRow("石料", payload.Rock) +
		buildReportResourceRow("铁锭", payload.Iron) +
		`</table>`
}

func buildReportResourceRow(label string, value int64) string {
	return `<tr><td width="120" height="25" align="center" valign="middle" class="TextArmyCount">` + html.EscapeString(label) + `</td><td width="129" height="25" align="center" valign="middle" class="TextArmyCount">` + fmt.Sprintf("%d", value) + `</td></tr>`
}

func min64(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func buildRichPlunderAttackerReport(originCity string, targetCity string, soldiers []Soldier, loot TroopResource, note string) string {
	return buildRichLegacyBattleReport(
		"掠夺战报",
		"report-outcome-win",
		"掠夺成功",
		originCity+" -> "+targetCity,
		note,
		[][2]string{
			{"行动", "掠夺"},
			{"路线", originCity + " -> " + targetCity},
			{"出征兵力", formatReportInt(reportSoldierTotal(soldiers))},
			{"缴获总量", formatReportInt(reportResourceTotal(loot))},
		},
		buildRichReportSection("出征兵力", buildRichReportSoldierTable(soldiers)),
		buildRichReportSection("缴获资源", buildRichReportResourceTable("本次缴获", loot)),
	)
}

func buildRichPlunderDefenderReport(originCity string, targetCity string, soldiers []Soldier, loot TroopResource, note string) string {
	return buildRichLegacyBattleReport(
		"敌情战报",
		"report-outcome-alert",
		"遭遇掠夺",
		originCity+" -> "+targetCity,
		note,
		[][2]string{
			{"事件", "敌军掠夺"},
			{"来源", originCity},
			{"目标", targetCity},
			{"损失总量", formatReportInt(reportResourceTotal(loot))},
		},
		buildRichReportSection("敌军兵力", buildRichReportSoldierTable(soldiers)),
		buildRichReportSection("损失资源", buildRichReportResourceTable("本次损失", loot)),
	)
}

func buildRichLegacyBattleReport(title string, outcomeClass string, outcome string, route string, note string, summaryRows [][2]string, sections ...string) string {
	var body strings.Builder
	body.WriteString(`<div class="report-shell">`)
	body.WriteString(`<div class="report-head">`)
	body.WriteString(`<div class="report-head-copy">`)
	body.WriteString(`<span class="report-head-kicker">军情战报</span>`)
	body.WriteString(`<strong>`)
	body.WriteString(html.EscapeString(title))
	body.WriteString(`</strong>`)
	body.WriteString(`<small>`)
	body.WriteString(html.EscapeString(route))
	body.WriteString(`</small>`)
	body.WriteString(`</div>`)
	body.WriteString(`<span class="report-outcome `)
	body.WriteString(outcomeClass)
	body.WriteString(`">`)
	body.WriteString(html.EscapeString(outcome))
	body.WriteString(`</span>`)
	body.WriteString(`</div>`)
	body.WriteString(`<div class="report-note">`)
	body.WriteString(html.EscapeString(note))
	body.WriteString(`</div>`)
	body.WriteString(`<div class="report-summary">`)
	body.WriteString(buildRichReportSummaryTable(summaryRows))
	body.WriteString(`</div>`)
	body.WriteString(`<div class="report-grid">`)
	for _, section := range sections {
		body.WriteString(section)
	}
	body.WriteString(`</div>`)
	body.WriteString(`</div>`)
	return body.String()
}

func buildRichReportSection(title string, content string) string {
	var body strings.Builder
	body.WriteString(`<section class="report-section">`)
	body.WriteString(`<div class="report-section-title">`)
	body.WriteString(html.EscapeString(title))
	body.WriteString(`</div>`)
	body.WriteString(`<div class="report-section-body">`)
	body.WriteString(content)
	body.WriteString(`</div>`)
	body.WriteString(`</section>`)
	return body.String()
}

func buildRichReportSummaryTable(rows [][2]string) string {
	var body strings.Builder
	body.WriteString(`<table width="567" border="0" cellpadding="0" cellspacing="0" class="report-summary-table">`)
	body.WriteString(`<tr class="TitleBlueWhite"><td colspan="4" height="28" align="center" valign="middle">战况摘要</td></tr>`)
	for index, row := range rows {
		if index%2 == 0 {
			body.WriteString(`<tr>`)
		}
		body.WriteString(`<td width="110" height="26" align="center" valign="middle" class="TitleBattleYellow">`)
		body.WriteString(html.EscapeString(row[0]))
		body.WriteString(`</td>`)
		body.WriteString(`<td width="173" height="26" align="center" valign="middle" class="TextArmyCount">`)
		body.WriteString(html.EscapeString(row[1]))
		body.WriteString(`</td>`)
		if index%2 == 1 {
			body.WriteString(`</tr>`)
		}
	}
	if len(rows)%2 == 1 {
		body.WriteString(`<td width="110" height="26" align="center" valign="middle" class="TitleBattleYellow">--</td>`)
		body.WriteString(`<td width="173" height="26" align="center" valign="middle" class="TextArmyCount">--</td></tr>`)
	}
	body.WriteString(`</table>`)
	return body.String()
}

func buildRichReportSoldierTable(soldiers []Soldier) string {
	var body strings.Builder
	body.WriteString(`<table width="249" border="0" cellpadding="0" cellspacing="0">`)
	body.WriteString(`<tr class="TitleBattleYellow"><td width="120" height="25" align="center" valign="middle">兵种</td><td width="129" height="25" align="center" valign="middle">数量</td></tr>`)
	for _, soldier := range soldiers {
		body.WriteString(`<tr><td width="120" height="25" align="center" valign="middle" class="TextArmyCount">`)
		body.WriteString(html.EscapeString(soldier.Name))
		body.WriteString(`</td><td width="129" height="25" align="center" valign="middle" class="TextArmyCount">`)
		body.WriteString(formatReportInt(soldier.Count))
		body.WriteString(`</td></tr>`)
	}
	body.WriteString(`</table>`)
	return body.String()
}

func buildRichReportResourceTable(title string, payload TroopResource) string {
	return `<table width="249" border="0" cellpadding="0" cellspacing="0">` +
		`<tr class="TitleBattleYellow"><td colspan="2" height="25" align="center" valign="middle">` + html.EscapeString(title) + `</td></tr>` +
		buildRichReportResourceRow("黄金", payload.Gold) +
		buildRichReportResourceRow("粮草", payload.Food) +
		buildRichReportResourceRow("木材", payload.Wood) +
		buildRichReportResourceRow("石料", payload.Rock) +
		buildRichReportResourceRow("铁锭", payload.Iron) +
		buildRichReportResourceRow("合计", reportResourceTotal(payload)) +
		`</table>`
}

func buildRichReportResourceRow(label string, value int64) string {
	return `<tr><td width="120" height="25" align="center" valign="middle" class="TextArmyCount">` + html.EscapeString(label) + `</td><td width="129" height="25" align="center" valign="middle" class="TextArmyCount">` + formatReportInt(value) + `</td></tr>`
}

func reportSoldierTotal(soldiers []Soldier) int64 {
	total := int64(0)
	for _, soldier := range soldiers {
		total += soldier.Count
	}
	return total
}

func reportResourceTotal(payload TroopResource) int64 {
	return payload.Gold + payload.Food + payload.Wood + payload.Rock + payload.Iron
}

func formatReportInt(value int64) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}

	raw := fmt.Sprintf("%d", value)
	if len(raw) <= 3 {
		return sign + raw
	}

	var body strings.Builder
	for index, ch := range raw {
		if index > 0 && (len(raw)-index)%3 == 0 {
			body.WriteByte(',')
		}
		body.WriteRune(ch)
	}
	return sign + body.String()
}
