package legacy

type transportReportInput struct {
	OriginCID     int
	OriginCity    string
	TargetCID     int
	TargetCity    string
	DepartTime    int64
	SettleTime    int64
	PathSeconds   int
	Payload       TroopResource
	ReturnStarted bool
}

const troopReportTitleTransport = 2

func (input transportReportInput) routeLabel() string {
	from := firstNonEmpty(input.OriginCity, formatCIDLabel(input.OriginCID))
	target := firstNonEmpty(input.TargetCity, formatCIDLabel(input.TargetCID))
	return from + " -> " + target
}

func (input transportReportInput) returnStateLabel() string {
	if input.ReturnStarted {
		return "Returning"
	}
	return "Holding"
}

func buildRichTransportSenderReport(input transportReportInput) string {
	note := "Transport payload has been delivered and the convoy is now returning."
	if reportResourceTotal(input.Payload) <= 0 {
		note = "Transport convoy reached destination with an empty payload and is now returning."
	}

	return buildRichLegacyBattleReport(
		"Transport Receipt",
		"report-outcome-win",
		"Delivered",
		input.routeLabel(),
		note,
		[][2]string{
			{"Action", "Transport"},
			{"Route", input.routeLabel()},
			{"Departed At", formatReportTime(input.DepartTime)},
			{"Arrived At", formatReportTime(input.SettleTime)},
			{"Outbound ETA", formatReportDuration(input.PathSeconds)},
			{"Payload Total", formatReportInt(reportResourceTotal(input.Payload))},
			{"Return State", input.returnStateLabel()},
		},
		buildRichReportSection("Delivery Payload", buildRichReportResourceTable("Delivered", input.Payload)),
	)
}

func buildRichTransportReceiverReport(input transportReportInput) string {
	note := "A friendly transport convoy has delivered supplies to this city."
	if reportResourceTotal(input.Payload) <= 0 {
		note = "A transport convoy arrived, but the delivered payload is empty."
	}

	return buildRichLegacyBattleReport(
		"Inbound Transport Notice",
		"report-outcome-alert",
		"Supplies Arrived",
		input.routeLabel(),
		note,
		[][2]string{
			{"Event", "Inbound Transport"},
			{"From", firstNonEmpty(input.OriginCity, formatCIDLabel(input.OriginCID))},
			{"To", firstNonEmpty(input.TargetCity, formatCIDLabel(input.TargetCID))},
			{"Departed At", formatReportTime(input.DepartTime)},
			{"Arrived At", formatReportTime(input.SettleTime)},
			{"Payload Total", formatReportInt(reportResourceTotal(input.Payload))},
			{"Sender Return", input.returnStateLabel()},
		},
		buildRichReportSection("Received Payload", buildRichReportResourceTable("Received", input.Payload)),
	)
}
