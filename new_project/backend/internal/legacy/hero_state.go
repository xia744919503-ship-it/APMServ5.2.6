package legacy

import "fmt"

func heroStateDisplayLabel(state int) string {
	switch state {
	case 0:
		return "空闲"
	case 1:
		return "城守"
	case 2:
		return "出征"
	case 3:
		return "战斗"
	case 4:
		return "驻守"
	case 5:
		return "俘虏"
	case 6:
		return "投奔"
	case 7:
		return "主将"
	case 8:
		return "军师"
	default:
		return fmt.Sprintf("状态 %d", state)
	}
}
