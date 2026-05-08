package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
)

const (
	collegeBuildingID            = 7
	cityResearchHeroBuffType     = 4
	cityResearchHeroBuffBase     = 10.0
	cityResearchHeroBuffBoost    = 12.5
	cityResearchGameSpeedRate    = 1.0
	maxCityResearchLevel         = 10
	cityResearchBusyReason       = "another research is already in progress"
	cityResearchInProgressReason = "research is already in progress"
)

type CityResearchSnapshot struct {
	CID       int                  `json:"cid"`
	Position  int                  `json:"position"`
	Level     int                  `json:"level"`
	ActiveTID int                  `json:"activeTid"`
	Options   []CityResearchOption `json:"options"`
}

type CityResearchOption struct {
	TID                  int                     `json:"tid"`
	Name                 string                  `json:"name"`
	Description          string                  `json:"description"`
	Level                int64                   `json:"level"`
	ShareLevel           int64                   `json:"shareLevel"`
	ResearchCID          int                     `json:"researchCid"`
	State                int                     `json:"state"`
	StateLabel           string                  `json:"stateLabel"`
	StateStartTime       int64                   `json:"stateStartTime"`
	StateEndTime         int64                   `json:"stateEndTime"`
	SecondsLeft          int64                   `json:"secondsLeft"`
	LevelDescription     string                  `json:"levelDescription"`
	NextLevel            int                     `json:"nextLevel"`
	NextLevelDescription string                  `json:"nextLevelDescription"`
	WoodNeed             int64                   `json:"woodNeed"`
	RockNeed             int64                   `json:"rockNeed"`
	IronNeed             int64                   `json:"ironNeed"`
	FoodNeed             int64                   `json:"foodNeed"`
	GoldNeed             int64                   `json:"goldNeed"`
	UpgradeDuration      int64                   `json:"upgradeDuration"`
	CanUpgrade           bool                    `json:"canUpgrade"`
	Reason               string                  `json:"reason"`
	Conditions           []CityResearchCondition `json:"conditions"`
}

type CityResearchCondition struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	RequiredLevel int    `json:"requiredLevel"`
	CurrentLevel  int    `json:"currentLevel"`
	Satisfied     bool   `json:"satisfied"`
}

type cityResearchCollegeRecord struct {
	ID       int
	Position int
	Level    int
}

type cityResearchConfig struct {
	Wood        int64
	Rock        int64
	Iron        int64
	Food        int64
	Gold        int64
	UpgradeTime int64
	Description string
}

type cityResearchResourceSnapshot struct {
	Wood   int64
	Rock   int64
	Iron   int64
	Food   int64
	Gold   int64
	People int64
}

type cityResearchRow struct {
	ID             int
	UID            int
	TID            int
	Level          int64
	CID            int
	State          int
	StateStartTime int64
	StateEndTime   int64
}

type cityResearchOptionRecord struct {
	TID         int
	Name        string
	Description string
	Level       int64
	ShareLevel  int64
	ResearchCID int
	State       int
	StateStart  int64
	StateEnd    int64
}

func (r *Repository) CityResearchSnapshot(ctx context.Context, uid int, cid int, position int) (CityResearchSnapshot, error) {
	if position <= 0 {
		return CityResearchSnapshot{}, newInvalidError("invalid college position")
	}
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return CityResearchSnapshot{}, err
	} else if !allowed {
		return CityResearchSnapshot{}, ErrForbidden
	}
	if r.db == nil {
		return r.fixtureCityResearchSnapshot(cid, position), nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return CityResearchSnapshot{}, err
	}
	defer tx.Rollback()

	if err := r.settleCityResearchQueueTx(ctx, tx, cid); err != nil {
		return CityResearchSnapshot{}, fmt.Errorf("settle city research queue: %w", err)
	}

	snapshot, err := r.cityResearchSnapshotTx(ctx, tx, uid, cid, position)
	if err != nil {
		return CityResearchSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return CityResearchSnapshot{}, err
	}
	return snapshot, nil
}

func (r *Repository) StartCityResearch(ctx context.Context, uid int, cid int, position int, tid int) (CityResearchSnapshot, error) {
	if position <= 0 {
		return CityResearchSnapshot{}, newInvalidError("invalid college position")
	}
	if tid <= 0 {
		return CityResearchSnapshot{}, newInvalidError("invalid technology id")
	}
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return CityResearchSnapshot{}, err
	} else if !allowed {
		return CityResearchSnapshot{}, ErrForbidden
	}
	if r.db == nil {
		return r.fixtureCityResearchSnapshot(cid, position), nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return CityResearchSnapshot{}, err
	}
	defer tx.Rollback()

	if err := r.settleCityResearchQueueTx(ctx, tx, cid); err != nil {
		return CityResearchSnapshot{}, fmt.Errorf("settle city research queue: %w", err)
	}

	college, err := r.loadCityCollegeTx(ctx, tx, cid, position)
	if err != nil {
		if err == sql.ErrNoRows {
			return CityResearchSnapshot{}, newInvalidError("college does not exist at the selected position")
		}
		return CityResearchSnapshot{}, fmt.Errorf("load college: %w", err)
	}

	activeCount, err := r.cityResearchActiveCountTx(ctx, tx, cid)
	if err != nil {
		return CityResearchSnapshot{}, fmt.Errorf("load research active count: %w", err)
	}
	if activeCount > 0 {
		return CityResearchSnapshot{}, newInvalidError(cityResearchBusyReason)
	}

	record, err := r.loadCityResearchOptionRecordTx(ctx, tx, uid, cid, tid)
	if err != nil {
		if err == sql.ErrNoRows {
			return CityResearchSnapshot{}, newInvalidError("technology is not configured")
		}
		return CityResearchSnapshot{}, fmt.Errorf("load research option: %w", err)
	}
	if record.Level >= maxCityResearchLevel {
		return CityResearchSnapshot{}, newInvalidError("technology has reached the level cap")
	}

	nextLevel := int(record.Level) + 1
	config, err := r.loadCityResearchConfigTx(ctx, tx, tid, nextLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			return CityResearchSnapshot{}, newInvalidError("next technology level is not configured")
		}
		return CityResearchSnapshot{}, fmt.Errorf("load research config: %w", err)
	}

	resources, err := r.cityResearchResourceSnapshotTx(ctx, tx, cid)
	if err != nil {
		return CityResearchSnapshot{}, fmt.Errorf("load city resources: %w", err)
	}
	if !cityResearchHasResources(resources, config) {
		return CityResearchSnapshot{}, newInvalidError("not enough resources to start research")
	}

	conditions, err := r.loadCityResearchConditionsTx(ctx, tx, uid, cid, tid, nextLevel)
	if err != nil {
		return CityResearchSnapshot{}, fmt.Errorf("load research conditions: %w", err)
	}
	if reason := firstUnsatisfiedCityResearchConditionReason(conditions); reason != "" {
		return CityResearchSnapshot{}, newInvalidError(reason)
	}

	now, err := r.currentUnixTimeTx(ctx, tx)
	if err != nil {
		return CityResearchSnapshot{}, fmt.Errorf("load current unix time: %w", err)
	}
	speedRate, err := r.cityResearchSpeedRateTx(ctx, tx, cid)
	if err != nil {
		return CityResearchSnapshot{}, fmt.Errorf("load research speed rate: %w", err)
	}
	finishAt := now + cityResearchDurationSeconds(config.UpgradeTime, speedRate)

	if err := r.reserveCityResearchCostTx(ctx, tx, cid, config); err != nil {
		return CityResearchSnapshot{}, fmt.Errorf("reserve city research cost: %w", err)
	}

	researchID, err := r.upsertCityResearchTx(ctx, tx, uid, cid, tid, int(record.Level), finishAt)
	if err != nil {
		return CityResearchSnapshot{}, fmt.Errorf("upsert city research row: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
insert into mem_technic_upgrading (id, cid, tid, level, state_endtime)
values (?, ?, ?, ?, ?)
on duplicate key update
	cid = values(cid),
	tid = values(tid),
	level = values(level),
	state_endtime = values(state_endtime)`,
		researchID,
		cid,
		tid,
		nextLevel,
		finishAt,
	); err != nil {
		return CityResearchSnapshot{}, fmt.Errorf("insert research queue row: %w", err)
	}

	snapshot, err := r.cityResearchSnapshotTx(ctx, tx, uid, cid, college.Position)
	if err != nil {
		return CityResearchSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return CityResearchSnapshot{}, err
	}
	return snapshot, nil
}

func (r *Repository) CancelCityResearch(ctx context.Context, uid int, cid int, position int, tid int) (CityResearchSnapshot, error) {
	if position <= 0 {
		return CityResearchSnapshot{}, newInvalidError("invalid college position")
	}
	if tid <= 0 {
		return CityResearchSnapshot{}, newInvalidError("invalid technology id")
	}
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return CityResearchSnapshot{}, err
	} else if !allowed {
		return CityResearchSnapshot{}, ErrForbidden
	}
	if r.db == nil {
		return r.fixtureCityResearchSnapshot(cid, position), nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return CityResearchSnapshot{}, err
	}
	defer tx.Rollback()

	if err := r.settleCityResearchQueueTx(ctx, tx, cid); err != nil {
		return CityResearchSnapshot{}, fmt.Errorf("settle city research queue: %w", err)
	}

	if _, err := r.loadCityCollegeTx(ctx, tx, cid, position); err != nil {
		if err == sql.ErrNoRows {
			return CityResearchSnapshot{}, newInvalidError("college does not exist at the selected position")
		}
		return CityResearchSnapshot{}, fmt.Errorf("load college: %w", err)
	}

	active, err := r.loadCityActiveResearchTx(ctx, tx, cid, tid)
	if err != nil {
		if err == sql.ErrNoRows {
			return CityResearchSnapshot{}, newInvalidError("technology is not being researched")
		}
		return CityResearchSnapshot{}, fmt.Errorf("load active research: %w", err)
	}

	targetLevel := int(active.Level) + 1
	config, err := r.loadCityResearchConfigTx(ctx, tx, tid, targetLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			return CityResearchSnapshot{}, newInvalidError("next technology level is not configured")
		}
		return CityResearchSnapshot{}, fmt.Errorf("load research config: %w", err)
	}

	if active.Level == 0 {
		if _, err := tx.ExecContext(ctx, "delete from sys_technic where id = ?", active.ID); err != nil {
			return CityResearchSnapshot{}, fmt.Errorf("delete research row: %w", err)
		}
	} else {
		if _, err := tx.ExecContext(ctx, `
update sys_technic
set cid = 0, state = 0, state_starttime = 0, state_endtime = 0
where id = ?`, active.ID); err != nil {
			return CityResearchSnapshot{}, fmt.Errorf("reset research row: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, "delete from mem_technic_upgrading where id = ?", active.ID); err != nil {
		return CityResearchSnapshot{}, fmt.Errorf("delete research queue row: %w", err)
	}
	if err := r.refundCityResearchCostTx(ctx, tx, cid, config); err != nil {
		return CityResearchSnapshot{}, fmt.Errorf("refund research resources: %w", err)
	}

	snapshot, err := r.cityResearchSnapshotTx(ctx, tx, uid, cid, position)
	if err != nil {
		return CityResearchSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return CityResearchSnapshot{}, err
	}
	return snapshot, nil
}

func (r *Repository) settleCityResearchQueue(ctx context.Context, cid int) error {
	if r.db == nil {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := r.settleCityResearchQueueTx(ctx, tx, cid); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) settleCityResearchQueueTx(ctx context.Context, tx *sql.Tx, cid int) error {
	now, err := r.currentUnixTimeTx(ctx, tx)
	if err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx, `
select id, cid, tid, level
from mem_technic_upgrading
where cid = ? and state_endtime <= ?
order by state_endtime, id`, cid, now)
	if err != nil {
		return err
	}
	defer rows.Close()

	items := make([]cityResearchRow, 0, 2)
	for rows.Next() {
		item := cityResearchRow{}
		if err := rows.Scan(&item.ID, &item.CID, &item.TID, &item.Level); err != nil {
			return err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, item := range items {
		if err := r.settleCityResearchRowTx(ctx, tx, item); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) settleCityResearchRowTx(ctx context.Context, tx *sql.Tx, row cityResearchRow) error {
	var uid int
	if err := tx.QueryRowContext(ctx, "select uid from sys_technic where id = ?", row.ID).Scan(&uid); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
update sys_technic
set level = ?, cid = 0, state = 0, state_starttime = 0, state_endtime = 0
where id = ?`, row.Level, row.ID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "delete from mem_technic_upgrading where id = ?", row.ID); err != nil {
		return err
	}
	return r.updateCityResearchLevelsTx(ctx, tx, uid, row.TID, int(row.Level))
}

func (r *Repository) cityResearchSnapshotTx(ctx context.Context, tx *sql.Tx, uid int, cid int, position int) (CityResearchSnapshot, error) {
	college, err := r.loadCityCollegeTx(ctx, tx, cid, position)
	if err != nil {
		if err == sql.ErrNoRows {
			return CityResearchSnapshot{}, newInvalidError("college does not exist at the selected position")
		}
		return CityResearchSnapshot{}, fmt.Errorf("load college: %w", err)
	}

	now, err := r.currentUnixTimeTx(ctx, tx)
	if err != nil {
		return CityResearchSnapshot{}, fmt.Errorf("load current unix time: %w", err)
	}
	resources, err := r.cityResearchResourceSnapshotTx(ctx, tx, cid)
	if err != nil {
		return CityResearchSnapshot{}, fmt.Errorf("load city resources: %w", err)
	}
	speedRate, err := r.cityResearchSpeedRateTx(ctx, tx, cid)
	if err != nil {
		return CityResearchSnapshot{}, fmt.Errorf("load research speed rate: %w", err)
	}
	activeCount, err := r.cityResearchActiveCountTx(ctx, tx, cid)
	if err != nil {
		return CityResearchSnapshot{}, fmt.Errorf("load research active count: %w", err)
	}

	options, err := r.loadCityResearchOptionsTx(ctx, tx, uid, cid, resources, speedRate, activeCount > 0, now)
	if err != nil {
		return CityResearchSnapshot{}, fmt.Errorf("load research options: %w", err)
	}

	activeTID := 0
	for _, option := range options {
		if option.State == 1 && option.ResearchCID == cid {
			activeTID = option.TID
			break
		}
	}

	return CityResearchSnapshot{
		CID:       cid,
		Position:  college.Position,
		Level:     college.Level,
		ActiveTID: activeTID,
		Options:   options,
	}, nil
}

func (r *Repository) loadCityCollegeTx(ctx context.Context, tx *sql.Tx, cid int, position int) (cityResearchCollegeRecord, error) {
	record := cityResearchCollegeRecord{}
	err := tx.QueryRowContext(ctx, `
select id, xy, level
from sys_building
where cid = ? and xy = ? and bid = ?
limit 1`, cid, position, collegeBuildingID).Scan(
		&record.ID,
		&record.Position,
		&record.Level,
	)
	return record, err
}

func (r *Repository) cityResearchActiveCountTx(ctx context.Context, tx *sql.Tx, cid int) (int, error) {
	var count int
	err := tx.QueryRowContext(ctx, `
select count(*)
from sys_technic
where cid = ? and state = 1`, cid).Scan(&count)
	return count, err
}

func (r *Repository) cityResearchResourceSnapshotTx(ctx context.Context, tx *sql.Tx, cid int) (cityResearchResourceSnapshot, error) {
	snapshot := cityResearchResourceSnapshot{}
	err := tx.QueryRowContext(ctx, `
select wood, rock, iron, food, gold, people
from mem_city_resource
where cid = ?`, cid).Scan(
		&snapshot.Wood,
		&snapshot.Rock,
		&snapshot.Iron,
		&snapshot.Food,
		&snapshot.Gold,
		&snapshot.People,
	)
	return snapshot, err
}

func (r *Repository) loadCityResearchOptionsTx(
	ctx context.Context,
	tx *sql.Tx,
	uid int,
	cid int,
	resources cityResearchResourceSnapshot,
	speedRate float64,
	activeBusy bool,
	now int64,
) ([]CityResearchOption, error) {
	rows, err := tx.QueryContext(ctx, `
select
	t.tid,
	coalesce(t.name, ''),
	coalesce(t.description, ''),
	coalesce(s.level, 0),
	coalesce(c.level, 0),
	coalesce(s.cid, 0),
	coalesce(s.state, 0),
	coalesce(s.state_starttime, 0),
	coalesce(s.state_endtime, 0)
from cfg_technic t
left join sys_technic s on s.uid = ? and s.tid = t.tid
left join sys_city_technic c on c.cid = ? and c.tid = t.tid
order by t.tid`, uid, cid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]cityResearchOptionRecord, 0, 24)
	for rows.Next() {
		record := cityResearchOptionRecord{}
		var name sql.NullString
		var description sql.NullString
		if err := rows.Scan(
			&record.TID,
			&name,
			&description,
			&record.Level,
			&record.ShareLevel,
			&record.ResearchCID,
			&record.State,
			&record.StateStart,
			&record.StateEnd,
		); err != nil {
			return nil, err
		}

		record.Name = strings.TrimSpace(name.String)
		record.Description = strings.TrimSpace(description.String)
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	options := make([]CityResearchOption, 0, len(records))
	for _, record := range records {
		option, err := r.cityResearchOptionSnapshotTx(ctx, tx, uid, cid, record, resources, speedRate, activeBusy, now)
		if err != nil {
			return nil, err
		}
		options = append(options, option)
	}
	return options, nil
}

func (r *Repository) cityResearchOptionSnapshotTx(
	ctx context.Context,
	tx *sql.Tx,
	uid int,
	cid int,
	record cityResearchOptionRecord,
	resources cityResearchResourceSnapshot,
	speedRate float64,
	activeBusy bool,
	now int64,
) (CityResearchOption, error) {
	option := CityResearchOption{
		TID:            record.TID,
		Name:           record.Name,
		Description:    record.Description,
		Level:          record.Level,
		ShareLevel:     record.ShareLevel,
		ResearchCID:    record.ResearchCID,
		State:          record.State,
		StateLabel:     cityResearchStateLabel(record.State),
		StateStartTime: record.StateStart,
		StateEndTime:   record.StateEnd,
	}

	if record.State == 1 && record.StateEnd > 0 {
		option.SecondsLeft = maxDraftInt64(0, record.StateEnd-now)
	}

	if record.Level > 0 {
		option.LevelDescription = r.cityResearchLevelDescriptionTx(ctx, tx, record.TID, int(record.Level))
	}

	nextLevel := int(record.Level) + 1
	option.NextLevel = nextLevel
	if nextLevel > maxCityResearchLevel {
		option.CanUpgrade = false
		option.Reason = "technology has reached the level cap"
		return option, nil
	}

	config, err := r.loadCityResearchConfigTx(ctx, tx, record.TID, nextLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			option.CanUpgrade = false
			option.Reason = "next technology level is not configured"
			return option, nil
		}
		return CityResearchOption{}, err
	}

	option.WoodNeed = config.Wood
	option.RockNeed = config.Rock
	option.IronNeed = config.Iron
	option.FoodNeed = config.Food
	option.GoldNeed = config.Gold
	option.UpgradeDuration = cityResearchDurationSeconds(config.UpgradeTime, speedRate)
	option.NextLevelDescription = config.Description

	conditions, err := r.loadCityResearchConditionsTx(ctx, tx, uid, cid, record.TID, nextLevel)
	if err != nil {
		return CityResearchOption{}, err
	}
	option.Conditions = conditions

	reason := ""
	if record.State == 1 && record.ResearchCID == cid {
		reason = cityResearchInProgressReason
	}
	if reason == "" && activeBusy {
		reason = cityResearchBusyReason
	}
	if reason == "" && !cityResearchHasResources(resources, config) {
		reason = "Not enough resources"
	}
	if reason == "" {
		reason = firstUnsatisfiedCityResearchConditionReason(conditions)
	}

	option.CanUpgrade = reason == ""
	option.Reason = reason
	return option, nil
}

func (r *Repository) loadCityResearchOptionRecordTx(ctx context.Context, tx *sql.Tx, uid int, cid int, tid int) (cityResearchOptionRecord, error) {
	record := cityResearchOptionRecord{}
	var name sql.NullString
	var description sql.NullString
	err := tx.QueryRowContext(ctx, `
select
	t.tid,
	coalesce(t.name, ''),
	coalesce(t.description, ''),
	coalesce(s.level, 0),
	coalesce(c.level, 0),
	coalesce(s.cid, 0),
	coalesce(s.state, 0),
	coalesce(s.state_starttime, 0),
	coalesce(s.state_endtime, 0)
from cfg_technic t
left join sys_technic s on s.uid = ? and s.tid = t.tid
left join sys_city_technic c on c.cid = ? and c.tid = t.tid
where t.tid = ?`, uid, cid, tid).Scan(
		&record.TID,
		&name,
		&description,
		&record.Level,
		&record.ShareLevel,
		&record.ResearchCID,
		&record.State,
		&record.StateStart,
		&record.StateEnd,
	)
	if err != nil {
		return cityResearchOptionRecord{}, err
	}
	record.Name = strings.TrimSpace(name.String)
	record.Description = strings.TrimSpace(description.String)
	return record, nil
}

func (r *Repository) loadCityResearchConfigTx(ctx context.Context, tx *sql.Tx, tid int, level int) (cityResearchConfig, error) {
	config := cityResearchConfig{}
	var (
		description sql.NullString
		woodRaw     float64
		rockRaw     float64
		ironRaw     float64
		foodRaw     float64
		goldRaw     float64
	)
	err := tx.QueryRowContext(ctx, `
select
	coalesce(upgrade_wood, 0),
	coalesce(upgrade_rock, 0),
	coalesce(upgrade_iron, 0),
	coalesce(upgrade_food, 0),
	coalesce(upgrade_gold, 0),
	coalesce(upgrade_time, 0),
	coalesce(description, '')
from cfg_technic_level
where tid = ? and level = ?`, tid, level).Scan(
		&woodRaw,
		&rockRaw,
		&ironRaw,
		&foodRaw,
		&goldRaw,
		&config.UpgradeTime,
		&description,
	)
	if err != nil {
		return cityResearchConfig{}, err
	}

	config.Wood = int64(math.Round(woodRaw))
	config.Rock = int64(math.Round(rockRaw))
	config.Iron = int64(math.Round(ironRaw))
	config.Food = int64(math.Round(foodRaw))
	config.Gold = int64(math.Round(goldRaw))
	config.Description = strings.TrimSpace(description.String)
	return config, nil
}

func (r *Repository) loadCityResearchConditionsTx(ctx context.Context, tx *sql.Tx, uid int, cid int, tid int, level int) ([]CityResearchCondition, error) {
	rows, err := tx.QueryContext(ctx, `
select pre_type, pre_id, pre_level
from cfg_technic_condition
where tid = ? and level = ?
order by pre_type, pre_id`, tid, level)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type rawCondition struct {
		PreType  int
		PreID    int
		PreLevel int
	}
	rawConditions := make([]rawCondition, 0, 6)
	for rows.Next() {
		item := rawCondition{}
		if err := rows.Scan(&item.PreType, &item.PreID, &item.PreLevel); err != nil {
			return nil, err
		}
		rawConditions = append(rawConditions, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	conditions := make([]CityResearchCondition, 0, len(rawConditions))
	for _, item := range rawConditions {
		condition := CityResearchCondition{
			RequiredLevel: item.PreLevel,
			Satisfied:     true,
		}
		switch item.PreType {
		case 0:
			condition.Type = "building"
			condition.Name = r.cityDraftBuildingNameTx(ctx, tx, item.PreID)
			currentLevel, err := r.cityResearchConditionBuildingLevelTx(ctx, tx, cid, item.PreID)
			if err != nil {
				return nil, err
			}
			condition.CurrentLevel = currentLevel
			condition.Satisfied = currentLevel >= item.PreLevel
		case 1:
			condition.Type = "technic"
			condition.Name = r.cityDraftTechnicNameTx(ctx, tx, item.PreID)
			currentLevel, err := r.cityDraftConditionTechnicLevelTx(ctx, tx, uid, item.PreID)
			if err != nil {
				return nil, err
			}
			condition.CurrentLevel = currentLevel
			condition.Satisfied = currentLevel >= item.PreLevel
		default:
			condition.Type = "unknown"
			condition.Name = fmt.Sprintf("Condition %d", item.PreID)
		}

		conditions = append(conditions, condition)
	}
	return conditions, nil
}

func (r *Repository) cityResearchConditionBuildingLevelTx(ctx context.Context, tx *sql.Tx, cid int, buildingID int) (int, error) {
	var level sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
select max(level)
from sys_building
where cid = ? and bid = ?`, cid, buildingID).Scan(&level); err != nil {
		return 0, err
	}
	if !level.Valid {
		return 0, nil
	}
	return int(level.Int64), nil
}

func (r *Repository) loadCityActiveResearchTx(ctx context.Context, tx *sql.Tx, cid int, tid int) (cityResearchRow, error) {
	row := cityResearchRow{}
	err := tx.QueryRowContext(ctx, `
select id, uid, tid, level, cid, state, state_starttime, state_endtime
from sys_technic
where cid = ? and tid = ? and state = 1
limit 1`, cid, tid).Scan(
		&row.ID,
		&row.UID,
		&row.TID,
		&row.Level,
		&row.CID,
		&row.State,
		&row.StateStartTime,
		&row.StateEndTime,
	)
	return row, err
}

func (r *Repository) upsertCityResearchTx(ctx context.Context, tx *sql.Tx, uid int, cid int, tid int, currentLevel int, finishAt int64) (int, error) {
	now, err := r.currentUnixTimeTx(ctx, tx)
	if err != nil {
		return 0, err
	}

	existing := cityResearchRow{}
	err = tx.QueryRowContext(ctx, `
select id, uid, tid, level, cid, state, state_starttime, state_endtime
from sys_technic
where uid = ? and tid = ?
limit 1`, uid, tid).Scan(
		&existing.ID,
		&existing.UID,
		&existing.TID,
		&existing.Level,
		&existing.CID,
		&existing.State,
		&existing.StateStartTime,
		&existing.StateEndTime,
	)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if err == sql.ErrNoRows {
		result, err := tx.ExecContext(ctx, `
insert into sys_technic (uid, tid, level, cid, state, state_starttime, state_endtime)
values (?, ?, 0, ?, 1, ?, ?)`, uid, tid, cid, now, finishAt)
		if err != nil {
			return 0, err
		}
		id, err := result.LastInsertId()
		return int(id), err
	}

	if _, err := tx.ExecContext(ctx, `
update sys_technic
set level = ?, cid = ?, state = 1, state_starttime = ?, state_endtime = ?
where id = ?`, currentLevel, cid, now, finishAt, existing.ID); err != nil {
		return 0, err
	}
	return existing.ID, nil
}

func (r *Repository) reserveCityResearchCostTx(ctx context.Context, tx *sql.Tx, cid int, config cityResearchConfig) error {
	result, err := tx.ExecContext(ctx, `
update mem_city_resource
set wood = wood - ?, rock = rock - ?, iron = iron - ?, food = food - ?, gold = gold - ?
where cid = ?
	and wood >= ?
	and rock >= ?
	and iron >= ?
	and food >= ?
	and gold >= ?`,
		config.Wood,
		config.Rock,
		config.Iron,
		config.Food,
		config.Gold,
		cid,
		config.Wood,
		config.Rock,
		config.Iron,
		config.Food,
		config.Gold,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return newInvalidError("city resources changed before research started")
	}
	return nil
}

func (r *Repository) refundCityResearchCostTx(ctx context.Context, tx *sql.Tx, cid int, config cityResearchConfig) error {
	_, err := tx.ExecContext(ctx, `
update mem_city_resource
set wood = wood + ?, rock = rock + ?, iron = iron + ?, food = food + ?, gold = gold + ?
where cid = ?`,
		buildingCancelRefundAmount(config.Wood),
		buildingCancelRefundAmount(config.Rock),
		buildingCancelRefundAmount(config.Iron),
		buildingCancelRefundAmount(config.Food),
		buildingCancelRefundAmount(config.Gold),
		cid,
	)
	return err
}

func (r *Repository) cityResearchSpeedRateTx(ctx context.Context, tx *sql.Tx, cid int) (float64, error) {
	hid, err := r.cityResearchSpeedHeroIDTx(ctx, tx, cid)
	if err != nil {
		return 0, err
	}
	if hid <= 0 {
		return 1, nil
	}

	var (
		wisdomBase  sql.NullFloat64
		wisdomAdd   sql.NullFloat64
		wisdomAddOn sql.NullFloat64
	)
	err = tx.QueryRowContext(ctx, `
select wisdom_base, wisdom_add, wisdom_add_on
from sys_city_hero
where hid = ?`, hid).Scan(&wisdomBase, &wisdomAdd, &wisdomAddOn)
	if err != nil {
		if err == sql.ErrNoRows {
			return 1, nil
		}
		return 0, err
	}

	multiplier := cityResearchHeroBuffBase
	hasBuff, err := r.cityResearchHeroHasSpeedBuffTx(ctx, tx, hid)
	if err != nil {
		return 0, err
	}
	if hasBuff {
		multiplier = cityResearchHeroBuffBoost
	}

	speedAdd := (wisdomBase.Float64+wisdomAdd.Float64)*multiplier + wisdomAddOn.Float64
	if speedAdd <= 0 {
		return 1, nil
	}
	return 1.0 / (1.0 + 0.01*speedAdd), nil
}

func (r *Repository) cityResearchSpeedHeroIDTx(ctx context.Context, tx *sql.Tx, cid int) (int, error) {
	var (
		chiefID      sql.NullInt64
		generalID    sql.NullInt64
		counsellorID sql.NullInt64
	)
	if err := tx.QueryRowContext(ctx, `
select chiefhid, generalid, counsellorid
from sys_city
where cid = ?`, cid).Scan(&chiefID, &generalID, &counsellorID); err != nil {
		return 0, err
	}

	switch {
	case counsellorID.Valid && counsellorID.Int64 > 0:
		return int(counsellorID.Int64), nil
	case generalID.Valid && generalID.Int64 > 0:
		return int(generalID.Int64), nil
	case chiefID.Valid && chiefID.Int64 > 0:
		return int(chiefID.Int64), nil
	default:
		return 0, nil
	}
}

func (r *Repository) cityResearchHeroHasSpeedBuffTx(ctx context.Context, tx *sql.Tx, hid int) (bool, error) {
	var count int
	if err := tx.QueryRowContext(ctx, `
select count(*)
from mem_hero_buffer
where hid = ? and buftype = ? and endtime > unix_timestamp()`,
		hid,
		cityResearchHeroBuffType,
	).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) updateCityResearchLevelsTx(ctx context.Context, tx *sql.Tx, uid int, tid int, level int) error {
	rows, err := tx.QueryContext(ctx, `
select cid
from sys_city
where uid = ?
order by cid`, uid)
	if err != nil {
		return err
	}
	defer rows.Close()

	cityIDs := make([]int, 0, 8)
	for rows.Next() {
		var cityID int
		if err := rows.Scan(&cityID); err != nil {
			return err
		}
		cityIDs = append(cityIDs, cityID)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, cityID := range cityIDs {
		appliedLevel := 0
		candidateLevel := level + 1
		for candidateLevel > 0 {
			appliedLevel, err = r.cityResearchAppliedLevelTx(ctx, tx, cityID, tid, candidateLevel)
			if err != nil {
				return err
			}
			if appliedLevel > 0 {
				break
			}
			candidateLevel--
		}
		if appliedLevel == 0 {
			continue
		}

		if _, err := tx.ExecContext(ctx, `
insert into sys_city_technic (cid, tid, level)
values (?, ?, ?)
on duplicate key update level = values(level)`, cityID, tid, appliedLevel); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) cityResearchAppliedLevelTx(ctx context.Context, tx *sql.Tx, cid int, tid int, level int) (int, error) {
	rows, err := tx.QueryContext(ctx, `
select pre_id, pre_level
from cfg_technic_condition
where tid = ? and level = ? and pre_type = 0
order by pre_id`, tid, level)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type rawCondition struct {
		BuildingID    int
		RequiredLevel int
	}
	rawConditions := make([]rawCondition, 0, 4)
	for rows.Next() {
		item := rawCondition{}
		if err := rows.Scan(&item.BuildingID, &item.RequiredLevel); err != nil {
			return 0, err
		}
		rawConditions = append(rawConditions, item)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	rows.Close()

	for _, item := range rawConditions {
		currentLevel, err := r.cityResearchConditionBuildingLevelTx(ctx, tx, cid, item.BuildingID)
		if err != nil {
			return 0, err
		}
		if currentLevel < item.RequiredLevel {
			return 0, nil
		}
	}
	return level, nil
}

func (r *Repository) cityResearchLevelDescriptionTx(ctx context.Context, tx *sql.Tx, tid int, level int) string {
	var description sql.NullString
	if err := tx.QueryRowContext(ctx, `
select description
from cfg_technic_level
where tid = ? and level = ?`, tid, level).Scan(&description); err != nil {
		return ""
	}
	return strings.TrimSpace(description.String)
}

func cityResearchHasResources(snapshot cityResearchResourceSnapshot, config cityResearchConfig) bool {
	return snapshot.Wood >= config.Wood &&
		snapshot.Rock >= config.Rock &&
		snapshot.Iron >= config.Iron &&
		snapshot.Food >= config.Food &&
		snapshot.Gold >= config.Gold
}

func firstUnsatisfiedCityResearchConditionReason(conditions []CityResearchCondition) string {
	for _, condition := range conditions {
		if condition.Satisfied {
			continue
		}
		return fmt.Sprintf("%s Lv.%d required", firstNonEmpty(condition.Name, condition.Type), condition.RequiredLevel)
	}
	return ""
}

func cityResearchDurationSeconds(baseTime int64, speedRate float64) int64 {
	if baseTime <= 0 {
		return 1
	}
	duration := int64(math.Floor(float64(baseTime) * speedRate / cityResearchGameSpeedRate))
	if duration < 1 {
		return 1
	}
	return duration
}

func cityResearchStateLabel(state int) string {
	if state == 1 {
		return "Researching"
	}
	return "Idle"
}

func (r *Repository) fixtureCityResearchSnapshot(cid int, position int) CityResearchSnapshot {
	return CityResearchSnapshot{
		CID:      cid,
		Position: position,
		Level:    1,
		Options:  []CityResearchOption{},
	}
}
