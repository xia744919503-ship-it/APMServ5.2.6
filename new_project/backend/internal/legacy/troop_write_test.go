package legacy

import (
	"context"
	"strings"
	"testing"
)

func TestFormatTroopSoldiersRawUsesLegacyTypeCountPrefix(t *testing.T) {
	raw := formatTroopSoldiersRaw(map[int]int64{
		7: 12,
		3: 5,
	})

	if raw != "2,3,5,7,12," {
		t.Fatalf("formatTroopSoldiersRaw() = %q, want legacy type-count format", raw)
	}
}

func TestParseTroopSoldierCountsAcceptsLegacyAndPairFormats(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want map[int]int64
	}{
		{
			name: "legacy type-count prefix",
			raw:  "2,3,5,7,12,",
			want: map[int]int64{3: 5, 7: 12},
		},
		{
			name: "pair-only compatibility",
			raw:  "3,5,7,12,",
			want: map[int]int64{3: 5, 7: 12},
		},
		{
			name: "duplicate soldiers are merged",
			raw:  "3,3,5,3,2,7,12,",
			want: map[int]int64{3: 7, 7: 12},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTroopSoldierCounts(tt.raw)
			if len(got) != len(tt.want) {
				t.Fatalf("len(parseTroopSoldierCounts(%q)) = %d, want %d: %#v", tt.raw, len(got), len(tt.want), got)
			}
			for sid, wantCount := range tt.want {
				if got[sid] != wantCount {
					t.Fatalf("sid %d count = %d, want %d: %#v", sid, got[sid], wantCount, got)
				}
			}
		})
	}
}

func TestParseTroopSoldiersAcceptsLegacyFormat(t *testing.T) {
	items, total := parseTroopSoldiers("2,3,5,7,12,", map[int]string{
		3: "Scout",
		7: "Light Cavalry",
	})

	if total != 17 {
		t.Fatalf("total = %d, want 17", total)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2: %#v", len(items), items)
	}
	if items[0].SID != 3 || items[0].Name != "Scout" || items[0].Count != 5 {
		t.Fatalf("items[0] = %#v, want sid=3 name=Scout count=5", items[0])
	}
	if items[1].SID != 7 || items[1].Name != "Light Cavalry" || items[1].Count != 12 {
		t.Fatalf("items[1] = %#v, want sid=7 name=Light Cavalry count=12", items[1])
	}
}

func TestFormatScoutReturnSoldiersRawKeepsMixedTroopsWhenScoutsDie(t *testing.T) {
	raw := formatScoutReturnSoldiersRaw(map[int]int64{
		scoutSoldierSID: 5,
		7:               12,
	}, 0)

	if raw != "1,7,12," {
		t.Fatalf("formatScoutReturnSoldiersRaw() = %q, want non-scout survivors to return", raw)
	}
}

func TestFormatScoutReturnSoldiersRawDeletesTroopWhenNoSoldiersSurvive(t *testing.T) {
	raw := formatScoutReturnSoldiersRaw(map[int]int64{
		scoutSoldierSID: 5,
	}, 0)

	if raw != "" {
		t.Fatalf("formatScoutReturnSoldiersRaw() = %q, want empty troop payload", raw)
	}
}

func TestBuildWorldTargetSnapshotIsSafeScoutTarget(t *testing.T) {
	snapshot := buildWorldTargetSnapshot(225185, 3, "Forest")

	if snapshot.UID != 0 {
		t.Fatalf("UID = %d, want unowned field target", snapshot.UID)
	}
	if snapshot.CityName != "Forest[185,225]" {
		t.Fatalf("CityName = %q, want type name plus coordinate label", snapshot.CityName)
	}
	if !snapshot.Resource.isZero() || !snapshot.Protected.isZero() {
		t.Fatalf("field resources = %#v protected = %#v, want zero-value resources", snapshot.Resource, snapshot.Protected)
	}
	if !snapshot.minimalEligible() {
		t.Fatalf("field target should be minimal eligible: %#v", snapshot)
	}
}

func TestFormatWorldTargetNameFallsBackToTypeID(t *testing.T) {
	got := formatWorldTargetName(225185, 9, " ")
	if got != "Type9[185,225]" {
		t.Fatalf("formatWorldTargetName() = %q, want fallback type label", got)
	}
}

func TestDueTroopsQueryIncludesOwnedTargets(t *testing.T) {
	query, args := dueTroopsQuery(42, 123456)

	if !strings.Contains(query, "uid = ? or targetcid in (select cid from sys_city where uid = ?)") {
		t.Fatalf("dueTroopsQuery() = %q, want current uid troops plus troops targeting owned cities", query)
	}
	for _, want := range []string{
		"task in (0, 1, 2, 3, 4)",
		"endtime > 0",
		"endtime <= ?",
		"state in (0, 1, 5)",
		"order by id asc",
	} {
		if !strings.Contains(query, want) {
			t.Fatalf("dueTroopsQuery() = %q, want condition %q", query, want)
		}
	}
	if len(args) != 3 {
		t.Fatalf("len(args) = %d, want 3: %#v", len(args), args)
	}
	if args[0] != 42 || args[1] != 42 || args[2] != int64(123456) {
		t.Fatalf("args = %#v, want uid, uid, now", args)
	}
}

func TestSettledTroopTaskMatrixIncludesLegacyOccupyGathering(t *testing.T) {
	for _, task := range []int{0, 1, 2, 3, 4} {
		if !isSettledTroopTask(task) {
			t.Fatalf("isSettledTroopTask(%d) = false, want true", task)
		}
	}
	for _, task := range []int{-1, 5, 6, 7} {
		if isSettledTroopTask(task) {
			t.Fatalf("isSettledTroopTask(%d) = true, want false", task)
		}
	}
}

func TestLegacyCampaignTaskHeroRequirementMatrix(t *testing.T) {
	for _, task := range []int{3, 4} {
		if !taskRequiresHero(task) {
			t.Fatalf("taskRequiresHero(%d) = false, want true", task)
		}
	}
	for _, task := range []int{0, 1, 2, 5} {
		if taskRequiresHero(task) {
			t.Fatalf("taskRequiresHero(%d) = true, want false", task)
		}
	}
}

func TestFixtureDispatchRejectsLegacyPlunderWithoutHero(t *testing.T) {
	repo := &Repository{}
	_, err := repo.DispatchCityTroop(context.Background(), 1, 101001, 102002, map[int]int64{1: 10}, 0, 3, TroopResource{})
	if err == nil {
		t.Fatalf("DispatchCityTroop task=3 without hero succeeded, want legacy hero requirement error")
	}
	if !strings.Contains(err.Error(), "task=3 requires hero") {
		t.Fatalf("DispatchCityTroop task=3 without hero error = %q, want hero requirement", err.Error())
	}
}

func TestFixtureDispatchKeepsLegacyOccupyUnsupported(t *testing.T) {
	repo := &Repository{}
	_, err := repo.DispatchCityTroop(context.Background(), 1, 101001, 102002, map[int]int64{1: 10}, 9001, 4, TroopResource{})
	if err == nil {
		t.Fatalf("DispatchCityTroop task=4 succeeded, want unsupported until occupy settlement is implemented")
	}
	if !strings.Contains(err.Error(), "task=3 are supported") {
		t.Fatalf("DispatchCityTroop task=4 error = %q, want supported task range", err.Error())
	}
}

func TestCallbackTroopStateMatrixIncludesLegacyGathering(t *testing.T) {
	allowed := []int{0, 2, 4, 5}
	for _, state := range allowed {
		if !canCallbackTroopState(state) {
			t.Fatalf("canCallbackTroopState(%d) = false, want true", state)
		}
	}

	blocked := []int{1, 3, 6, 7, 8, 9}
	for _, state := range blocked {
		if canCallbackTroopState(state) {
			t.Fatalf("canCallbackTroopState(%d) = true, want false", state)
		}
	}
}

func TestDueTroopStateMatrixIncludesLegacyGathering(t *testing.T) {
	for _, state := range []int{0, 1, 5} {
		if !isDueTroopState(state) {
			t.Fatalf("isDueTroopState(%d) = false, want true", state)
		}
	}
	for _, state := range []int{2, 3, 4, 6} {
		if isDueTroopState(state) {
			t.Fatalf("isDueTroopState(%d) = true, want false", state)
		}
	}
}

func TestCallbackTroopReturnRouteForMovingAndStationaryLegacyStates(t *testing.T) {
	now := int64(1000)

	moving := troopRecord{CID: 10, StartCID: 10, TargetCID: 20, State: 0, StartTime: 940, PathTime: 300}
	startCID, targetCID, endTime := callbackTroopReturnRoute(moving, now)
	if startCID != 10 || targetCID != 20 || endTime != 1060 {
		t.Fatalf("moving route = start %d target %d end %d, want 10 20 1060", startCID, targetCID, endTime)
	}

	for _, state := range []int{4, 5} {
		record := troopRecord{CID: 10, StartCID: 10, TargetCID: 20, State: state, StartTime: 700, PathTime: 300}
		startCID, targetCID, endTime := callbackTroopReturnRoute(record, now)
		if startCID != 20 || targetCID != 10 || endTime != 1300 {
			t.Fatalf("state %d route = start %d target %d end %d, want 20 10 1300", state, startCID, targetCID, endTime)
		}
		if !callbackTroopMarksCityResources(state) {
			t.Fatalf("callbackTroopMarksCityResources(%d) = false, want true", state)
		}
	}
}

func TestCityActiveTroopCountSQLCountsReturningCapacity(t *testing.T) {
	query := cityActiveTroopCountSQL()

	for _, want := range []string{
		"from sys_troops",
		"cid = ?",
		"uid = ?",
		"state < 4",
	} {
		if !strings.Contains(query, want) {
			t.Fatalf("cityActiveTroopCountSQL() = %q, want %q", query, want)
		}
	}
	if strings.Contains(query, "state = 0") {
		t.Fatalf("cityActiveTroopCountSQL() = %q, must include returning state=1 capacity", query)
	}
}

func TestDefendedTargetPlunderReturnsEmptyLoot(t *testing.T) {
	snapshot := plunderTargetSnapshot{
		Soldiers: 1,
		Resource: TroopResource{
			Gold: 1000,
			Food: 1000,
			Wood: 1000,
			Rock: 1000,
			Iron: 1000,
		},
	}

	if snapshot.minimalEligible() {
		t.Fatalf("minimalEligible() = true, want defended target to block plunder")
	}
	if loot := plunderLootForSettlement(snapshot, 5000); !loot.isZero() {
		t.Fatalf("plunderLootForSettlement() = %#v, want empty loot for defended target", loot)
	}
}

func TestEnsureCityResourceRowSQLOnlyCreatesRowsForExistingCities(t *testing.T) {
	query := ensureCityResourceRowSQL()

	for _, want := range []string{
		"insert ignore into mem_city_resource",
		"select cid,0,0,0,0,0,0",
		"from sys_city",
		"where cid = ?",
	} {
		if !strings.Contains(query, want) {
			t.Fatalf("ensureCityResourceRowSQL() = %q, want %q", query, want)
		}
	}
}
