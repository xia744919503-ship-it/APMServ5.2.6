package legacy

import "fmt"

func reportDisplayHeadline(summary ReportSummary) string {
	titleLabel := reportTitleLabel(summary.Title)
	location := formatReportLocation(summary)
	if location == "" {
		return fmt.Sprintf("%s #%d", titleLabel, summary.ID)
	}

	return fmt.Sprintf("%s #%d · %s", titleLabel, summary.ID, location)
}

func reportTitleLabel(title int) string {
	switch title {
	case troopReportTitleTransport:
		return "运输回执"
	case troopReportTitleStation:
		return "驻军回执"
	case troopReportTitleScout:
		return "侦察战报"
	case 8:
		return "掠夺战报"
	default:
		return "战报"
	}
}
