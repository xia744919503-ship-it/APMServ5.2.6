package legacy

import (
	"fmt"
	"html"
	"strings"
	"time"
)

const troopReportTitleScout = 7

type scoutReportInput struct {
	OriginCID            int
	OriginCity           string
	TargetCID            int
	TargetCity           string
	DepartTime           int64
	SettleTime           int64
	PathSeconds          int
	AttackerScouts       int64
	AttackerSurvivors    int64
	DefenderScoutsBefore int64
	DefenderScoutsAfter  int64
	TargetSnapshot       plunderTargetSnapshot
}

func (input scoutReportInput) attackerLosses() int64 {
	loss := input.AttackerScouts - input.AttackerSurvivors
	if loss < 0 {
		return 0
	}
	return loss
}

func (input scoutReportInput) defenderLosses() int64 {
	loss := input.DefenderScoutsBefore - input.DefenderScoutsAfter
	if loss < 0 {
		return 0
	}
	return loss
}

func (input scoutReportInput) routeLabel() string {
	from := firstNonEmpty(input.OriginCity, formatCIDLabel(input.OriginCID))
	target := firstNonEmpty(input.TargetCity, formatCIDLabel(input.TargetCID))
	return from + " -> " + target
}

func buildRichScoutAttackerReport(input scoutReportInput) string {
	outcomeClass := "report-outcome-alert"
	outcomeText := "侦察受阻"
	note := "我军斥候在目标城池遭遇拦截，侦察链路中断。"
	if input.AttackerSurvivors > 0 {
		outcomeClass = "report-outcome-win"
		outcomeText = "侦察回传"
		note = "侦察部队已到点并回传情报，正在返程。"
	}

	return buildRichLegacyBattleReport(
		"侦察战报",
		outcomeClass,
		outcomeText,
		input.routeLabel(),
		note,
		[][2]string{
			{"行动", "侦察"},
			{"路线", input.routeLabel()},
			{"到点时间", formatReportTime(input.SettleTime)},
			{"去程耗时", formatReportDuration(input.PathSeconds)},
			{"派出斥候", formatReportInt(input.AttackerScouts)},
			{"斥候损失", formatReportInt(input.attackerLosses())},
			{"敌方斥候损失", formatReportInt(input.defenderLosses())},
			{"情报状态", outcomeText},
		},
		buildRichReportSection("交战摘要", buildScoutBattleTable(input)),
		buildRichReportSection("目标军情", buildScoutIntelTable(input)),
		buildRichReportSection("目标资源快照", buildRichReportResourceTable("到点资源", input.TargetSnapshot.Resource)),
	)
}

func buildRichScoutDefenderReport(input scoutReportInput) string {
	note := "敌军斥候已抵达我方城池，请注意加强防御与反侦察。"
	if input.AttackerSurvivors <= 0 {
		note = "敌军斥候在侦察过程中被拦截，侦察链路已被切断。"
	}

	return buildRichLegacyBattleReport(
		"敌军侦察告警",
		"report-outcome-alert",
		"侦察来袭",
		input.routeLabel(),
		note,
		[][2]string{
			{"事件", "敌军侦察"},
			{"来源", firstNonEmpty(input.OriginCity, formatCIDLabel(input.OriginCID))},
			{"目标", firstNonEmpty(input.TargetCity, formatCIDLabel(input.TargetCID))},
			{"到点时间", formatReportTime(input.SettleTime)},
			{"敌军斥候", formatReportInt(input.AttackerScouts)},
			{"拦截击毁", formatReportInt(input.attackerLosses())},
			{"本城斥候损失", formatReportInt(input.defenderLosses())},
			{"当前警戒", "建议加派斥候"},
		},
		buildRichReportSection("交战摘要", buildScoutBattleTable(input)),
		buildRichReportSection("本城军情", buildScoutIntelTable(input)),
	)
}

func buildScoutBattleTable(input scoutReportInput) string {
	rows := [][2]string{
		{"我军派出斥候", formatReportInt(input.AttackerScouts)},
		{"我军存活斥候", formatReportInt(input.AttackerSurvivors)},
		{"我军斥候损失", formatReportInt(input.attackerLosses())},
		{"守军斥候(战前)", formatReportInt(input.DefenderScoutsBefore)},
		{"守军斥候(战后)", formatReportInt(input.DefenderScoutsAfter)},
		{"守军斥候损失", formatReportInt(input.defenderLosses())},
	}
	return buildScoutKeyValueTable("侦察交战结算", rows)
}

func buildScoutIntelTable(input scoutReportInput) string {
	rows := [][2]string{
		{"目标坐标", formatCIDLabel(input.TargetCID)},
		{"目标总兵力", formatReportInt(input.TargetSnapshot.Soldiers)},
		{"目标城防", formatReportInt(input.TargetSnapshot.Defences)},
		{"目标驻军队列", formatReportInt(input.TargetSnapshot.Stationed)},
		{"目标守将数量", formatReportInt(input.TargetSnapshot.Heroes)},
	}
	return buildScoutKeyValueTable("目标情报摘要", rows)
}

func buildScoutKeyValueTable(title string, rows [][2]string) string {
	var body strings.Builder
	body.WriteString(`<table width="249" border="0" cellpadding="0" cellspacing="0">`)
	body.WriteString(`<tr class="TitleBattleYellow"><td colspan="2" height="25" align="center" valign="middle">`)
	body.WriteString(html.EscapeString(title))
	body.WriteString(`</td></tr>`)
	for _, row := range rows {
		body.WriteString(`<tr><td width="120" height="25" align="center" valign="middle" class="TextArmyCount">`)
		body.WriteString(html.EscapeString(row[0]))
		body.WriteString(`</td><td width="129" height="25" align="center" valign="middle" class="TextArmyCount">`)
		body.WriteString(html.EscapeString(row[1]))
		body.WriteString(`</td></tr>`)
	}
	body.WriteString(`</table>`)
	return body.String()
}

func formatReportTime(unix int64) string {
	if unix <= 0 {
		return "--"
	}
	return time.Unix(unix, 0).Format("2006-01-02 15:04:05")
}

func formatReportDuration(seconds int) string {
	if seconds <= 0 {
		return "--"
	}
	return fmt.Sprintf("%d秒", seconds)
}
