package legacy

var legacyBuildingNames = map[int]string{
	1:  "农田",
	2:  "伐木场",
	3:  "采石场",
	4:  "铁矿",
	5:  "民房",
	6:  "官府",
	7:  "书院",
	8:  "校场",
	9:  "兵营",
	10: "客栈",
	11: "招贤馆",
	12: "寺庙",
	13: "市场",
	14: "铁匠铺",
	15: "工匠作坊",
	16: "马厩",
	17: "仓库",
	18: "驿站",
	19: "烽火台",
	20: "城墙",
}

func legacyBuildingName(bid int, fallback string) string {
	if name, ok := legacyBuildingNames[bid]; ok {
		return name
	}
	return fallback
}
