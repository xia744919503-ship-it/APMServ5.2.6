package legacy

import "testing"

func TestRemoveBattleNewsTagsReturnsSingleLinePlainText(t *testing.T) {
	raw := `<font color="#ff0000">汉军</font>&nbsp;攻占<br/>据点 <b>洛阳</b>`

	got := removeBattleNewsTags(raw)
	want := "汉军 攻占 据点 洛阳"
	if got != want {
		t.Fatalf("removeBattleNewsTags() = %q, want %q", got, want)
	}
}

func TestBattleTaskGroupIDsMatchLegacyUnionMapping(t *testing.T) {
	tests := []struct {
		name    string
		bid     int
		unionID int
		want    []int
	}{
		{name: "1001 uses Han task groups", bid: 1001, unionID: 3, want: []int{60000, 60001, 60002, 60003, 60004}},
		{name: "Han union", bid: 2001, unionID: 1, want: []int{60000, 60001, 60002, 60003, 60004}},
		{name: "Dong union", bid: 2001, unionID: 3, want: []int{60005, 60006, 60007, 60008, 60009, 60010}},
		{name: "Yuan union", bid: 2001, unionID: 4, want: []int{60011, 60012, 60013, 60014, 60015, 60016, 60017, 60018}},
		{name: "unsupported union", bid: 2001, unionID: 2, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := battleTaskGroupIDs(tt.bid, tt.unionID)
			if len(got) != len(tt.want) {
				t.Fatalf("len(battleTaskGroupIDs(%d, %d)) = %d, want %d: %#v", tt.bid, tt.unionID, len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("battleTaskGroupIDs(%d, %d)[%d] = %d, want %d: %#v", tt.bid, tt.unionID, i, got[i], tt.want[i], got)
				}
			}
		})
	}
}
