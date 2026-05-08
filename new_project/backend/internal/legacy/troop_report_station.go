package legacy

const troopReportTitleStation = 9

type stationReportInput struct {
	HomeCID     int
	HomeCity    string
	TargetCID   int
	TargetCity  string
	DepartTime  int64
	SettleTime  int64
	PathSeconds int
	Soldiers    []Soldier
}

func (input stationReportInput) outboundRouteLabel() string {
	from := firstNonEmpty(input.HomeCity, formatCIDLabel(input.HomeCID))
	target := firstNonEmpty(input.TargetCity, formatCIDLabel(input.TargetCID))
	return from + " -> " + target
}

func (input stationReportInput) returnRouteLabel() string {
	from := firstNonEmpty(input.TargetCity, formatCIDLabel(input.TargetCID))
	target := firstNonEmpty(input.HomeCity, formatCIDLabel(input.HomeCID))
	return from + " -> " + target
}

func buildRichStationArrivalReport(input stationReportInput) string {
	return buildRichLegacyBattleReport(
		"驻军回执",
		"report-outcome-win",
		"驻军就位",
		input.outboundRouteLabel(),
		"部队已抵达目标城池并进入驻扎待命，可在军务司中继续查看或下达撤回。",
		[][2]string{
			{"行动", "驻军"},
			{"路线", input.outboundRouteLabel()},
			{"出发时间", formatReportTime(input.DepartTime)},
			{"到点时间", formatReportTime(input.SettleTime)},
			{"去程耗时", formatReportDuration(input.PathSeconds)},
			{"驻军总数", formatReportInt(reportSoldierTotal(input.Soldiers))},
			{"当前状态", "驻扎待命"},
		},
		buildRichReportSection("驻军兵力", buildRichReportSoldierTable(input.Soldiers)),
	)
}

func buildRichStationReturnReport(input stationReportInput) string {
	return buildRichLegacyBattleReport(
		"驻军回执",
		"report-outcome-win",
		"回营完成",
		input.returnRouteLabel(),
		"驻军部队已完成撤回并返回本城，兵力重新编入城内调度。",
		[][2]string{
			{"行动", "回营"},
			{"路线", input.returnRouteLabel()},
			{"返程开始", formatReportTime(input.DepartTime)},
			{"回营时间", formatReportTime(input.SettleTime)},
			{"返程耗时", formatReportDuration(input.PathSeconds)},
			{"回营兵力", formatReportInt(reportSoldierTotal(input.Soldiers))},
			{"当前状态", "已归城"},
		},
		buildRichReportSection("回营兵力", buildRichReportSoldierTable(input.Soldiers)),
	)
}
