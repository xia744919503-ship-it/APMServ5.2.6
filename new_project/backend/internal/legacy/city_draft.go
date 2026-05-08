package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
)

const (
	barracksBuildingID            = 9
	citySoldierTrainTechnicID     = 8
	citySiegeTrainTechnicID       = 5
	citySoldierTrainHeroBuffType  = 3
	citySoldierTrainHeroBuffBase  = 10.0
	citySoldierTrainHeroBuffBoost = 12.5
	citySoldierDraftNeedTime      = 3
	citySoldierDraftGameSpeedRate = 1.0
)

type CityBarracksSnapshot struct {
	CID           int                       `json:"cid"`
	Position      int                       `json:"position"`
	Level         int                       `json:"level"`
	QueueCapacity int                       `json:"queueCapacity"`
	QueueCount    int                       `json:"queueCount"`
	FreePeople    int64                     `json:"freePeople"`
	Options       []CityBarracksDraftOption `json:"options"`
	Queue         []CityBarracksQueueItem   `json:"queue"`
}

type CityBarracksDraftOption struct {
	SID           int                          `json:"sid"`
	Name          string                       `json:"name"`
	Description   string                       `json:"description"`
	Count         int64                        `json:"count"`
	HP            int64                        `json:"hp"`
	AP            int64                        `json:"ap"`
	DP            int64                        `json:"dp"`
	Range         int64                        `json:"range"`
	Speed         int64                        `json:"speed"`
	Carry         int64                        `json:"carry"`
	FoodUse       float64                      `json:"foodUse"`
	WoodNeed      int64                        `json:"woodNeed"`
	RockNeed      int64                        `json:"rockNeed"`
	IronNeed      int64                        `json:"ironNeed"`
	FoodNeed      int64                        `json:"foodNeed"`
	GoldNeed      int64                        `json:"goldNeed"`
	PeopleNeed    int64                        `json:"peopleNeed"`
	DraftDuration int64                        `json:"draftDuration"`
	CanDraft      bool                         `json:"canDraft"`
	Reason        string                       `json:"reason"`
	Conditions    []CityBarracksDraftCondition `json:"conditions"`
}

type CityBarracksDraftCondition struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	RequiredLevel int    `json:"requiredLevel"`
	CurrentLevel  int    `json:"currentLevel"`
	Satisfied     bool   `json:"satisfied"`
}

type CityBarracksQueueItem struct {
	ID             int    `json:"id"`
	SID            int    `json:"sid"`
	Name           string `json:"name"`
	Count          int64  `json:"count"`
	State          int    `json:"state"`
	StateLabel     string `json:"stateLabel"`
	DraftInterval  int64  `json:"draftInterval"`
	NeedTime       int64  `json:"needTime"`
	StateStartTime int64  `json:"stateStartTime"`
	EndTime        int64  `json:"endTime"`
	SecondsLeft    int64  `json:"secondsLeft"`
	AccMark        bool   `json:"accMark"`
}

type cityBarracksRecord struct {
	ID       int
	Position int
	Level    int
}

type cityDraftResourceSnapshot struct {
	Wood           int64
	Rock           int64
	Iron           int64
	Food           int64
	Gold           int64
	People         int64
	PeopleWorking  int64
	PeopleBuilding int64
}

type citySoldierDraftRecord struct {
	SID         int
	Name        string
	Description string
	HP          int64
	AP          int64
	DP          int64
	Range       int64
	Speed       int64
	Carry       int64
	TimeNeed    int64
	WoodNeed    int64
	RockNeed    int64
	IronNeed    int64
	FoodNeed    int64
	GoldNeed    int64
	PeopleNeed  int64
	FoodUse     float64
	Count       int64
}

type cityDraftQueueRow struct {
	ID             int
	CID            int
	Position       int
	SID            int
	Name           string
	Count          int64
	State          int
	DraftInterval  int64
	StateStartTime int64
	NeedTime       int64
	AccMark        bool
}

func (r *Repository) CityBarracksSnapshot(ctx context.Context, uid int, cid int, position int) (CityBarracksSnapshot, error) {
	if position <= 0 {
		return CityBarracksSnapshot{}, newInvalidError("invalid barracks position")
	}
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return CityBarracksSnapshot{}, err
	} else if !allowed {
		return CityBarracksSnapshot{}, ErrForbidden
	}
	if r.db == nil {
		return r.fixtureCityBarracksSnapshot(cid, position), nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return CityBarracksSnapshot{}, err
	}
	defer tx.Rollback()

	if err := r.settleCityDraftQueueTx(ctx, tx, cid); err != nil {
		return CityBarracksSnapshot{}, fmt.Errorf("settle city draft queue: %w", err)
	}

	snapshot, err := r.cityBarracksSnapshotTx(ctx, tx, uid, cid, position)
	if err != nil {
		return CityBarracksSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return CityBarracksSnapshot{}, err
	}
	return snapshot, nil
}

func (r *Repository) StartCitySoldierDraft(ctx context.Context, uid int, cid int, position int, sid int, count int) (CityBarracksSnapshot, error) {
	if position <= 0 {
		return CityBarracksSnapshot{}, newInvalidError("invalid barracks position")
	}
	if sid <= 0 || count <= 0 {
		return CityBarracksSnapshot{}, newInvalidError("invalid soldier draft request")
	}
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return CityBarracksSnapshot{}, err
	} else if !allowed {
		return CityBarracksSnapshot{}, ErrForbidden
	}
	if r.db == nil {
		return r.fixtureCityBarracksSnapshot(cid, position), nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return CityBarracksSnapshot{}, err
	}
	defer tx.Rollback()

	if err := r.settleCityDraftQueueTx(ctx, tx, cid); err != nil {
		return CityBarracksSnapshot{}, fmt.Errorf("settle city draft queue: %w", err)
	}

	barracks, err := r.loadCityBarracksTx(ctx, tx, cid, position)
	if err != nil {
		if err == sql.ErrNoRows {
			return CityBarracksSnapshot{}, newInvalidError("barracks does not exist at the selected position")
		}
		return CityBarracksSnapshot{}, fmt.Errorf("load barracks: %w", err)
	}

	queueCount, err := r.cityBarracksQueueCountTx(ctx, tx, cid, position)
	if err != nil {
		return CityBarracksSnapshot{}, fmt.Errorf("load barracks queue count: %w", err)
	}
	if queueCount >= barracks.Level+1 {
		return CityBarracksSnapshot{}, newInvalidError("barracks queue is full")
	}

	draftRecord, err := r.loadCitySoldierDraftRecordTx(ctx, tx, cid, sid)
	if err != nil {
		if err == sql.ErrNoRows {
			return CityBarracksSnapshot{}, newInvalidError("city soldier branch is not configured")
		}
		return CityBarracksSnapshot{}, fmt.Errorf("load city soldier config: %w", err)
	}

	conditions, err := r.loadCitySoldierDraftConditionsTx(ctx, tx, uid, cid, position, sid)
	if err != nil {
		return CityBarracksSnapshot{}, fmt.Errorf("load soldier draft conditions: %w", err)
	}
	if reason := firstUnsatisfiedCityDraftConditionReason(conditions); reason != "" {
		return CityBarracksSnapshot{}, newInvalidError(reason)
	}

	resources, err := r.cityDraftResourceSnapshotTx(ctx, tx, cid)
	if err != nil {
		return CityBarracksSnapshot{}, fmt.Errorf("load city draft resources: %w", err)
	}
	if !cityHasDraftResources(resources, draftRecord, int64(count)) {
		return CityBarracksSnapshot{}, newInvalidError("not enough resources to start soldier draft")
	}
	if cityDraftFreePeople(resources) < draftRecord.PeopleNeed*int64(count) {
		return CityBarracksSnapshot{}, newInvalidError("not enough free people to start soldier draft")
	}

	now, err := r.currentUnixTimeTx(ctx, tx)
	if err != nil {
		return CityBarracksSnapshot{}, fmt.Errorf("load current unix time: %w", err)
	}
	speedRate, err := r.citySoldierDraftSpeedRateTx(ctx, tx, cid, sid)
	if err != nil {
		return CityBarracksSnapshot{}, fmt.Errorf("load soldier draft speed rate: %w", err)
	}
	draftInterval := citySoldierDraftIntervalSeconds(draftRecord.TimeNeed, speedRate)
	needTime := citySoldierDraftLegacyNeedTime(draftInterval)

	if err := r.reserveCitySoldierDraftCostTx(ctx, tx, cid, draftRecord, int64(count)); err != nil {
		return CityBarracksSnapshot{}, fmt.Errorf("reserve city soldier draft cost: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
insert into sys_city_draftqueue (cid, xy, sid, count, queuetime, state, draft_interval, state_starttime, needtime, accmark)
values (?, ?, ?, ?, ?, 0, ?, 0, ?, 0)`,
		cid,
		position,
		sid,
		count,
		now,
		draftInterval,
		needTime,
	); err != nil {
		return CityBarracksSnapshot{}, fmt.Errorf("insert city soldier draft queue: %w", err)
	}
	if err := r.ensureActiveCityDraftQueueTx(ctx, tx, cid, position); err != nil {
		return CityBarracksSnapshot{}, fmt.Errorf("activate city soldier draft queue: %w", err)
	}

	snapshot, err := r.cityBarracksSnapshotTx(ctx, tx, uid, cid, position)
	if err != nil {
		return CityBarracksSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return CityBarracksSnapshot{}, err
	}
	return snapshot, nil
}

func (r *Repository) CancelCitySoldierDraft(ctx context.Context, uid int, cid int, position int, queueID int) (CityBarracksSnapshot, error) {
	if position <= 0 {
		return CityBarracksSnapshot{}, newInvalidError("invalid barracks position")
	}
	if queueID <= 0 {
		return CityBarracksSnapshot{}, newInvalidError("invalid soldier draft queue id")
	}
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return CityBarracksSnapshot{}, err
	} else if !allowed {
		return CityBarracksSnapshot{}, ErrForbidden
	}
	if r.db == nil {
		return r.fixtureCityBarracksSnapshot(cid, position), nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return CityBarracksSnapshot{}, err
	}
	defer tx.Rollback()

	if err := r.settleCityDraftQueueTx(ctx, tx, cid); err != nil {
		return CityBarracksSnapshot{}, fmt.Errorf("settle city draft queue: %w", err)
	}

	if _, err := r.loadCityBarracksTx(ctx, tx, cid, position); err != nil {
		if err == sql.ErrNoRows {
			return CityBarracksSnapshot{}, newInvalidError("barracks does not exist at the selected position")
		}
		return CityBarracksSnapshot{}, fmt.Errorf("load barracks: %w", err)
	}

	queue, err := r.loadCityDraftQueueRowTx(ctx, tx, cid, position, queueID)
	if err != nil {
		if err == sql.ErrNoRows {
			return CityBarracksSnapshot{}, newInvalidError("soldier draft queue item does not exist")
		}
		return CityBarracksSnapshot{}, fmt.Errorf("load city soldier draft queue: %w", err)
	}
	draftRecord, err := r.loadCitySoldierDraftRecordTx(ctx, tx, cid, queue.SID)
	if err != nil {
		if err == sql.ErrNoRows {
			return CityBarracksSnapshot{}, newInvalidError("city soldier branch is not configured")
		}
		return CityBarracksSnapshot{}, fmt.Errorf("load city soldier config: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "delete from sys_city_draftqueue where id = ? and cid = ? and xy = ?", queue.ID, cid, position); err != nil {
		return CityBarracksSnapshot{}, fmt.Errorf("delete city soldier draft queue: %w", err)
	}
	if err := r.refundCitySoldierDraftCostTx(ctx, tx, cid, draftRecord, queue.Count); err != nil {
		return CityBarracksSnapshot{}, fmt.Errorf("refund city soldier draft cost: %w", err)
	}
	if queue.State == 1 {
		if _, err := tx.ExecContext(ctx, "delete from mem_city_draft where id = ?", queue.ID); err != nil {
			return CityBarracksSnapshot{}, fmt.Errorf("delete active city soldier draft queue: %w", err)
		}
		if err := r.ensureActiveCityDraftQueueTx(ctx, tx, cid, position); err != nil {
			return CityBarracksSnapshot{}, fmt.Errorf("activate next city soldier draft queue: %w", err)
		}
	}

	snapshot, err := r.cityBarracksSnapshotTx(ctx, tx, uid, cid, position)
	if err != nil {
		return CityBarracksSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return CityBarracksSnapshot{}, err
	}
	return snapshot, nil
}

func (r *Repository) settleCityDraftQueue(ctx context.Context, cid int) error {
	if r.db == nil {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := r.settleCityDraftQueueTx(ctx, tx, cid); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) settleCityDraftQueueTx(ctx context.Context, tx *sql.Tx, cid int) error {
	now, err := r.currentUnixTimeTx(ctx, tx)
	if err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx, `
select id, cid, xy, sid, count
from mem_city_draft
where cid = ? and state_endtime <= ?
order by state_endtime, id`, cid, now)
	if err != nil {
		return err
	}
	defer rows.Close()

	items := make([]cityDraftQueueRow, 0, 4)
	for rows.Next() {
		item := cityDraftQueueRow{}
		if err := rows.Scan(&item.ID, &item.CID, &item.Position, &item.SID, &item.Count); err != nil {
			return err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, item := range items {
		if err := r.settleCityDraftQueueRowTx(ctx, tx, item); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) settleCityDraftQueueRowTx(ctx context.Context, tx *sql.Tx, row cityDraftQueueRow) error {
	if err := r.addCitySoldiersTx(ctx, tx, row.CID, map[int]int64{row.SID: row.Count}); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "delete from sys_city_draftqueue where id = ?", row.ID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "delete from mem_city_draft where id = ?", row.ID); err != nil {
		return err
	}
	if err := r.ensureActiveCityDraftQueueTx(ctx, tx, row.CID, row.Position); err != nil {
		return err
	}

	uid, err := r.cityOwnerUIDTx(ctx, tx, row.CID)
	if err != nil {
		return err
	}
	if uid > 0 {
		if err := r.updateUserPrestigeTx(ctx, tx, uid); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) cityBarracksSnapshotTx(ctx context.Context, tx *sql.Tx, uid int, cid int, position int) (CityBarracksSnapshot, error) {
	barracks, err := r.loadCityBarracksTx(ctx, tx, cid, position)
	if err != nil {
		if err == sql.ErrNoRows {
			return CityBarracksSnapshot{}, newInvalidError("barracks does not exist at the selected position")
		}
		return CityBarracksSnapshot{}, fmt.Errorf("load barracks: %w", err)
	}

	now, err := r.currentUnixTimeTx(ctx, tx)
	if err != nil {
		return CityBarracksSnapshot{}, fmt.Errorf("load current unix time: %w", err)
	}
	resources, err := r.cityDraftResourceSnapshotTx(ctx, tx, cid)
	if err != nil {
		return CityBarracksSnapshot{}, fmt.Errorf("load city draft resources: %w", err)
	}
	options, err := r.loadCityBarracksDraftOptionsTx(ctx, tx, uid, cid, position, barracks.Level, resources)
	if err != nil {
		return CityBarracksSnapshot{}, fmt.Errorf("load city soldier draft options: %w", err)
	}
	queue, err := r.loadCityBarracksQueueItemsTx(ctx, tx, cid, position, now)
	if err != nil {
		return CityBarracksSnapshot{}, fmt.Errorf("load city soldier draft queue: %w", err)
	}

	return CityBarracksSnapshot{
		CID:           cid,
		Position:      position,
		Level:         barracks.Level,
		QueueCapacity: barracks.Level + 1,
		QueueCount:    len(queue),
		FreePeople:    cityDraftFreePeople(resources),
		Options:       options,
		Queue:         queue,
	}, nil
}

func (r *Repository) loadCityBarracksTx(ctx context.Context, tx *sql.Tx, cid int, position int) (cityBarracksRecord, error) {
	record := cityBarracksRecord{}
	err := tx.QueryRowContext(ctx, `
select id, xy, level
from sys_building
where cid = ? and xy = ? and bid = ?
limit 1`, cid, position, barracksBuildingID).Scan(
		&record.ID,
		&record.Position,
		&record.Level,
	)
	return record, err
}

func (r *Repository) cityDraftResourceSnapshotTx(ctx context.Context, tx *sql.Tx, cid int) (cityDraftResourceSnapshot, error) {
	snapshot := cityDraftResourceSnapshot{}
	err := tx.QueryRowContext(ctx, `
select wood, rock, iron, food, gold, people, people_working, people_building
from mem_city_resource
where cid = ?`, cid).Scan(
		&snapshot.Wood,
		&snapshot.Rock,
		&snapshot.Iron,
		&snapshot.Food,
		&snapshot.Gold,
		&snapshot.People,
		&snapshot.PeopleWorking,
		&snapshot.PeopleBuilding,
	)
	return snapshot, err
}

func (r *Repository) loadCityBarracksDraftOptionsTx(
	ctx context.Context,
	tx *sql.Tx,
	uid int,
	cid int,
	position int,
	barracksLevel int,
	resources cityDraftResourceSnapshot,
) ([]CityBarracksDraftOption, error) {
	rows, err := tx.QueryContext(ctx, `
select
	s.sid,
	coalesce(s.name, ''),
	coalesce(s.description, ''),
	coalesce(s.hp, 0),
	coalesce(s.ap, 0),
	coalesce(s.dp, 0),
	coalesce(s.range, 0),
	coalesce(s.speed, 0),
	coalesce(s.carry, 0),
	coalesce(s.time_need, 0),
	coalesce(s.wood_need, 0),
	coalesce(s.rock_need, 0),
	coalesce(s.iron_need, 0),
	coalesce(s.food_need, 0),
	coalesce(s.gold_need, 0),
	coalesce(s.people_need, 0),
	coalesce(s.food_use, 0),
	coalesce(c.count, 0)
from cfg_soldier s
left join sys_city_soldier c on c.cid = ? and c.sid = s.sid
where s.fromcity = 1
order by s.sid`, cid)
	if err != nil {
		return nil, err
	}

	records := make([]citySoldierDraftRecord, 0, 12)
	for rows.Next() {
		record := citySoldierDraftRecord{}
		var (
			name        sql.NullString
			description sql.NullString
			woodRaw     float64
			rockRaw     float64
			ironRaw     float64
			foodRaw     float64
			goldRaw     float64
			foodUseRaw  float64
		)
		if err := rows.Scan(
			&record.SID,
			&name,
			&description,
			&record.HP,
			&record.AP,
			&record.DP,
			&record.Range,
			&record.Speed,
			&record.Carry,
			&record.TimeNeed,
			&woodRaw,
			&rockRaw,
			&ironRaw,
			&foodRaw,
			&goldRaw,
			&record.PeopleNeed,
			&foodUseRaw,
			&record.Count,
		); err != nil {
			return nil, err
		}

		record.Name = strings.TrimSpace(name.String)
		record.Description = strings.TrimSpace(description.String)
		record.WoodNeed = int64(math.Round(woodRaw))
		record.RockNeed = int64(math.Round(rockRaw))
		record.IronNeed = int64(math.Round(ironRaw))
		record.FoodNeed = int64(math.Round(foodRaw))
		record.GoldNeed = int64(math.Round(goldRaw))
		record.FoodUse = foodUseRaw
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	if barracksLevel <= 0 {
		return []CityBarracksDraftOption{}, nil
	}

	options := make([]CityBarracksDraftOption, 0, len(records))
	for _, record := range records {
		conditions, err := r.loadCitySoldierDraftConditionsTx(ctx, tx, uid, cid, position, record.SID)
		if err != nil {
			return nil, err
		}
		speedRate, err := r.citySoldierDraftSpeedRateTx(ctx, tx, cid, record.SID)
		if err != nil {
			return nil, err
		}

		reason := firstUnsatisfiedCityDraftConditionReason(conditions)
		if reason == "" && !cityHasDraftResources(resources, record, 1) {
			reason = "Not enough resources"
		}
		if reason == "" && cityDraftFreePeople(resources) < record.PeopleNeed {
			reason = "Not enough free people"
		}

		options = append(options, CityBarracksDraftOption{
			SID:           record.SID,
			Name:          record.Name,
			Description:   record.Description,
			Count:         record.Count,
			HP:            record.HP,
			AP:            record.AP,
			DP:            record.DP,
			Range:         record.Range,
			Speed:         record.Speed,
			Carry:         record.Carry,
			FoodUse:       record.FoodUse,
			WoodNeed:      record.WoodNeed,
			RockNeed:      record.RockNeed,
			IronNeed:      record.IronNeed,
			FoodNeed:      record.FoodNeed,
			GoldNeed:      record.GoldNeed,
			PeopleNeed:    record.PeopleNeed,
			DraftDuration: citySoldierDraftIntervalSeconds(record.TimeNeed, speedRate),
			CanDraft:      reason == "",
			Reason:        reason,
			Conditions:    conditions,
		})
	}
	return options, nil
}

func (r *Repository) loadCityBarracksQueueItemsTx(ctx context.Context, tx *sql.Tx, cid int, position int, now int64) ([]CityBarracksQueueItem, error) {
	rows, err := tx.QueryContext(ctx, `
select
	d.id,
	d.sid,
	coalesce(s.name, ''),
	d.count,
	d.state,
	d.draft_interval,
	d.state_starttime,
	d.needtime,
	d.accmark
from sys_city_draftqueue d
left join cfg_soldier s on s.sid = d.sid
where d.cid = ? and d.xy = ?
order by d.state desc, d.queuetime, d.id`, cid, position)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]CityBarracksQueueItem, 0, 6)
	for rows.Next() {
		row := cityDraftQueueRow{}
		var name sql.NullString
		if err := rows.Scan(
			&row.ID,
			&row.SID,
			&name,
			&row.Count,
			&row.State,
			&row.DraftInterval,
			&row.StateStartTime,
			&row.NeedTime,
			&row.AccMark,
		); err != nil {
			return nil, err
		}

		endTime := int64(0)
		secondsLeft := row.NeedTime
		if row.State == 1 && row.StateStartTime > 0 {
			endTime = row.StateStartTime + row.NeedTime
			secondsLeft = maxDraftInt64(0, endTime-now)
		}

		items = append(items, CityBarracksQueueItem{
			ID:             row.ID,
			SID:            row.SID,
			Name:           strings.TrimSpace(name.String),
			Count:          row.Count,
			State:          row.State,
			StateLabel:     cityDraftQueueStateLabel(row.State),
			DraftInterval:  row.DraftInterval,
			NeedTime:       row.NeedTime,
			StateStartTime: row.StateStartTime,
			EndTime:        endTime,
			SecondsLeft:    secondsLeft,
			AccMark:        row.AccMark,
		})
	}
	return items, rows.Err()
}

func (r *Repository) loadCityDraftQueueRowTx(ctx context.Context, tx *sql.Tx, cid int, position int, queueID int) (cityDraftQueueRow, error) {
	row := cityDraftQueueRow{}
	err := tx.QueryRowContext(ctx, `
select id, cid, xy, sid, count, state, draft_interval, state_starttime, needtime, accmark
from sys_city_draftqueue
where id = ? and cid = ? and xy = ?`, queueID, cid, position).Scan(
		&row.ID,
		&row.CID,
		&row.Position,
		&row.SID,
		&row.Count,
		&row.State,
		&row.DraftInterval,
		&row.StateStartTime,
		&row.NeedTime,
		&row.AccMark,
	)
	return row, err
}

func (r *Repository) cityBarracksQueueCountTx(ctx context.Context, tx *sql.Tx, cid int, position int) (int, error) {
	var count int
	err := tx.QueryRowContext(ctx, `
select count(*)
from sys_city_draftqueue
where cid = ? and xy = ?`, cid, position).Scan(&count)
	return count, err
}

func (r *Repository) loadCitySoldierDraftRecordTx(ctx context.Context, tx *sql.Tx, cid int, sid int) (citySoldierDraftRecord, error) {
	record := citySoldierDraftRecord{}
	var (
		name        sql.NullString
		description sql.NullString
		woodRaw     float64
		rockRaw     float64
		ironRaw     float64
		foodRaw     float64
		goldRaw     float64
		foodUseRaw  float64
	)
	err := tx.QueryRowContext(ctx, `
select
	s.sid,
	coalesce(s.name, ''),
	coalesce(s.description, ''),
	coalesce(s.hp, 0),
	coalesce(s.ap, 0),
	coalesce(s.dp, 0),
	coalesce(s.range, 0),
	coalesce(s.speed, 0),
	coalesce(s.carry, 0),
	coalesce(s.time_need, 0),
	coalesce(s.wood_need, 0),
	coalesce(s.rock_need, 0),
	coalesce(s.iron_need, 0),
	coalesce(s.food_need, 0),
	coalesce(s.gold_need, 0),
	coalesce(s.people_need, 0),
	coalesce(s.food_use, 0),
	coalesce(c.count, 0)
from cfg_soldier s
left join sys_city_soldier c on c.cid = ? and c.sid = s.sid
where s.sid = ? and s.fromcity = 1`, cid, sid).Scan(
		&record.SID,
		&name,
		&description,
		&record.HP,
		&record.AP,
		&record.DP,
		&record.Range,
		&record.Speed,
		&record.Carry,
		&record.TimeNeed,
		&woodRaw,
		&rockRaw,
		&ironRaw,
		&foodRaw,
		&goldRaw,
		&record.PeopleNeed,
		&foodUseRaw,
		&record.Count,
	)
	if err != nil {
		return citySoldierDraftRecord{}, err
	}

	record.Name = strings.TrimSpace(name.String)
	record.Description = strings.TrimSpace(description.String)
	record.WoodNeed = int64(math.Round(woodRaw))
	record.RockNeed = int64(math.Round(rockRaw))
	record.IronNeed = int64(math.Round(ironRaw))
	record.FoodNeed = int64(math.Round(foodRaw))
	record.GoldNeed = int64(math.Round(goldRaw))
	record.FoodUse = foodUseRaw
	return record, nil
}

func (r *Repository) loadCitySoldierDraftConditionsTx(ctx context.Context, tx *sql.Tx, uid int, cid int, position int, sid int) ([]CityBarracksDraftCondition, error) {
	rows, err := tx.QueryContext(ctx, `
select pre_type, pre_id, pre_level
from cfg_soldier_condition
where sid = ?
order by pre_type, pre_id`, sid)
	if err != nil {
		return nil, err
	}

	type rawCondition struct {
		PreType  int
		PreID    int
		PreLevel int
	}
	rawConditions := make([]rawCondition, 0, 4)
	for rows.Next() {
		item := rawCondition{}
		if err := rows.Scan(&item.PreType, &item.PreID, &item.PreLevel); err != nil {
			rows.Close()
			return nil, err
		}
		rawConditions = append(rawConditions, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	conditions := make([]CityBarracksDraftCondition, 0, 4)
	for _, item := range rawConditions {
		condition := CityBarracksDraftCondition{
			RequiredLevel: item.PreLevel,
			Satisfied:     true,
		}
		switch item.PreType {
		case 0:
			condition.Type = "building"
			condition.Name = r.cityDraftBuildingNameTx(ctx, tx, item.PreID)
			currentLevel, err := r.cityDraftConditionBuildingLevelTx(ctx, tx, cid, position, item.PreID)
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

func (r *Repository) cityDraftConditionBuildingLevelTx(ctx context.Context, tx *sql.Tx, cid int, position int, buildingID int) (int, error) {
	var level sql.NullInt64
	query := `
select max(level)
from sys_building
where cid = ? and bid = ?`
	args := []any{cid, buildingID}
	if buildingID == barracksBuildingID {
		query = `
select level
from sys_building
where cid = ? and bid = ? and xy = ?
limit 1`
		args = append(args, position)
	}
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&level); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	if !level.Valid {
		return 0, nil
	}
	return int(level.Int64), nil
}

func (r *Repository) cityDraftConditionTechnicLevelTx(ctx context.Context, tx *sql.Tx, uid int, technicID int) (int, error) {
	var level sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
select max(level)
from sys_technic
where uid = ? and tid = ?`, uid, technicID).Scan(&level); err != nil {
		return 0, err
	}
	if !level.Valid {
		return 0, nil
	}
	return int(level.Int64), nil
}

func (r *Repository) cityDraftBuildingNameTx(ctx context.Context, tx *sql.Tx, buildingID int) string {
	var name sql.NullString
	if err := tx.QueryRowContext(ctx, "select name from cfg_building where bid = ?", buildingID).Scan(&name); err != nil {
		return fmt.Sprintf("Building %d", buildingID)
	}
	return firstNonEmpty(strings.TrimSpace(name.String), fmt.Sprintf("Building %d", buildingID))
}

func (r *Repository) cityDraftTechnicNameTx(ctx context.Context, tx *sql.Tx, technicID int) string {
	var name sql.NullString
	if err := tx.QueryRowContext(ctx, "select name from cfg_technic where tid = ?", technicID).Scan(&name); err != nil {
		return fmt.Sprintf("Technic %d", technicID)
	}
	return firstNonEmpty(strings.TrimSpace(name.String), fmt.Sprintf("Technic %d", technicID))
}

func (r *Repository) citySoldierDraftSpeedRateTx(ctx context.Context, tx *sql.Tx, cid int, sid int) (float64, error) {
	speedAdd := 0.0
	technicID := 0
	switch {
	case sid >= 9 && sid <= 12:
		technicID = citySiegeTrainTechnicID
	case sid > 0 && sid <= 8:
		technicID = citySoldierTrainTechnicID
	}

	if technicID > 0 {
		var level sql.NullInt64
		if err := tx.QueryRowContext(ctx, `
select level
from sys_city_technic
where cid = ? and tid = ?`, cid, technicID).Scan(&level); err != nil && err != sql.ErrNoRows {
			return 0, err
		} else if level.Valid && level.Int64 > 0 {
			speedAdd += float64(level.Int64 * 100)
		}
	}

	hid, err := r.cityDraftSpeedHeroIDTx(ctx, tx, cid)
	if err != nil {
		return 0, err
	}
	if hid > 0 {
		heroAdd, err := r.cityDraftHeroSpeedAddTx(ctx, tx, hid)
		if err != nil {
			return 0, err
		}
		speedAdd += heroAdd
	}

	return 1.0 / (10.0 + 0.1*speedAdd), nil
}

func (r *Repository) cityDraftSpeedHeroIDTx(ctx context.Context, tx *sql.Tx, cid int) (int, error) {
	var (
		chiefHID      int
		generalHID    int
		counsellorHID int
	)
	if err := tx.QueryRowContext(ctx, `
select chiefhid, generalid, counsellorid
from sys_city
where cid = ?`, cid).Scan(&chiefHID, &generalHID, &counsellorHID); err != nil {
		return 0, err
	}
	switch {
	case generalHID > 0:
		return generalHID, nil
	case chiefHID > 0:
		return chiefHID, nil
	case counsellorHID > 0:
		return counsellorHID, nil
	default:
		return 0, nil
	}
}

func (r *Repository) cityDraftHeroSpeedAddTx(ctx context.Context, tx *sql.Tx, hid int) (float64, error) {
	var (
		braveryBase  sql.NullFloat64
		braveryAdd   sql.NullFloat64
		braveryAddOn sql.NullFloat64
	)
	err := tx.QueryRowContext(ctx, `
select bravery_base, bravery_add, bravery_add_on
from sys_city_hero
where hid = ?`, hid).Scan(&braveryBase, &braveryAdd, &braveryAddOn)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	multiplier := citySoldierTrainHeroBuffBase
	hasBuff, err := r.cityDraftHeroHasSpeedBuffTx(ctx, tx, hid)
	if err != nil {
		return 0, err
	}
	if hasBuff {
		multiplier = citySoldierTrainHeroBuffBoost
	}

	return (braveryBase.Float64+braveryAdd.Float64)*multiplier + braveryAddOn.Float64, nil
}

func (r *Repository) cityDraftHeroHasSpeedBuffTx(ctx context.Context, tx *sql.Tx, hid int) (bool, error) {
	var count int
	if err := tx.QueryRowContext(ctx, `
select count(*)
from mem_hero_buffer
where hid = ? and buftype = ? and endtime > unix_timestamp()`,
		hid,
		citySoldierTrainHeroBuffType,
	).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) reserveCitySoldierDraftCostTx(ctx context.Context, tx *sql.Tx, cid int, record citySoldierDraftRecord, count int64) error {
	if count <= 0 {
		return nil
	}
	woodCost := record.WoodNeed * count
	rockCost := record.RockNeed * count
	ironCost := record.IronNeed * count
	foodCost := record.FoodNeed * count
	goldCost := record.GoldNeed * count
	peopleCost := record.PeopleNeed * count

	result, err := tx.ExecContext(ctx, `
update mem_city_resource
set wood = wood - ?, rock = rock - ?, iron = iron - ?, food = food - ?, gold = gold - ?, people = people - ?
where cid = ?
	and wood >= ?
	and rock >= ?
	and iron >= ?
	and food >= ?
	and gold >= ?
	and people >= ?`,
		woodCost,
		rockCost,
		ironCost,
		foodCost,
		goldCost,
		peopleCost,
		cid,
		woodCost,
		rockCost,
		ironCost,
		foodCost,
		goldCost,
		peopleCost,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return newInvalidError("city resources changed before soldier draft started")
	}
	return nil
}

func (r *Repository) refundCitySoldierDraftCostTx(ctx context.Context, tx *sql.Tx, cid int, record citySoldierDraftRecord, count int64) error {
	if count <= 0 {
		return nil
	}

	_, err := tx.ExecContext(ctx, `
update mem_city_resource
set wood = wood + ?, rock = rock + ?, iron = iron + ?, food = food + ?, gold = gold + ?, people = people + ?
where cid = ?`,
		buildingCancelRefundAmount(record.WoodNeed*count),
		buildingCancelRefundAmount(record.RockNeed*count),
		buildingCancelRefundAmount(record.IronNeed*count),
		buildingCancelRefundAmount(record.FoodNeed*count),
		buildingCancelRefundAmount(record.GoldNeed*count),
		record.PeopleNeed*count,
		cid,
	)
	return err
}

func (r *Repository) ensureActiveCityDraftQueueTx(ctx context.Context, tx *sql.Tx, cid int, position int) error {
	var activeCount int
	if err := tx.QueryRowContext(ctx, `
select count(*)
from sys_city_draftqueue
where cid = ? and xy = ? and state = 1`, cid, position).Scan(&activeCount); err != nil {
		return err
	}
	if activeCount > 0 {
		return nil
	}

	row := cityDraftQueueRow{}
	err := tx.QueryRowContext(ctx, `
select id, cid, xy, sid, count, needtime
from sys_city_draftqueue
where cid = ? and xy = ?
order by queuetime, id
limit 1`, cid, position).Scan(
		&row.ID,
		&row.CID,
		&row.Position,
		&row.SID,
		&row.Count,
		&row.NeedTime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	now, err := r.currentUnixTimeTx(ctx, tx)
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `
update sys_city_draftqueue
set state = 1, state_starttime = ?
where id = ? and state = 0`, now, row.ID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return nil
	}

	_, err = tx.ExecContext(ctx, `
insert into mem_city_draft (id, cid, xy, sid, count, state_endtime)
values (?, ?, ?, ?, ?, ?)
on duplicate key update
	cid = values(cid),
	xy = values(xy),
	sid = values(sid),
	count = values(count),
	state_endtime = values(state_endtime)`,
		row.ID,
		row.CID,
		row.Position,
		row.SID,
		row.Count,
		now+row.NeedTime,
	)
	return err
}

func (r *Repository) cityOwnerUIDTx(ctx context.Context, tx *sql.Tx, cid int) (int, error) {
	var uid int
	err := tx.QueryRowContext(ctx, "select uid from sys_city where cid = ?", cid).Scan(&uid)
	return uid, err
}

func (r *Repository) updateUserPrestigeTx(ctx context.Context, tx *sql.Tx, uid int) error {
	var (
		buildingPrestige sql.NullInt64
		soldierPrestige  sql.NullInt64
		troopPrestige    sql.NullInt64
		warPrestige      sql.NullInt64
	)
	if err := tx.QueryRowContext(ctx, `
select
	coalesce((select sum(r.people_building)
		from sys_city c
		join mem_city_resource r on r.cid = c.cid
		where c.uid = ?), 0),
	coalesce((select sum(cfg.people_need * s.count)
		from sys_city c
		join sys_city_soldier s on s.cid = c.cid
		join cfg_soldier cfg on cfg.sid = s.sid
		where c.uid = ?), 0),
	coalesce((select sum(people)
		from sys_troops
		where uid = ?), 0),
	coalesce((select warprestige
		from sys_user
		where uid = ?), 0)`,
		uid,
		uid,
		uid,
		uid,
	).Scan(
		&buildingPrestige,
		&soldierPrestige,
		&troopPrestige,
		&warPrestige,
	); err != nil {
		return err
	}

	prestige := buildingPrestige.Int64 + soldierPrestige.Int64 + troopPrestige.Int64 + warPrestige.Int64
	if prestige < 0 {
		prestige = 0
	}
	_, err := tx.ExecContext(ctx, "update sys_user set prestige = ? where uid = ?", prestige, uid)
	return err
}

func cityDraftFreePeople(snapshot cityDraftResourceSnapshot) int64 {
	free := snapshot.People - snapshot.PeopleWorking - snapshot.PeopleBuilding
	if free < 0 {
		return 0
	}
	return free
}

func cityHasDraftResources(snapshot cityDraftResourceSnapshot, record citySoldierDraftRecord, count int64) bool {
	if count <= 0 {
		return true
	}
	return snapshot.Wood >= record.WoodNeed*count &&
		snapshot.Rock >= record.RockNeed*count &&
		snapshot.Iron >= record.IronNeed*count &&
		snapshot.Food >= record.FoodNeed*count &&
		snapshot.Gold >= record.GoldNeed*count
}

func firstUnsatisfiedCityDraftConditionReason(conditions []CityBarracksDraftCondition) string {
	for _, condition := range conditions {
		if condition.Satisfied {
			continue
		}
		return fmt.Sprintf("%s Lv.%d required", firstNonEmpty(condition.Name, condition.Type), condition.RequiredLevel)
	}
	return ""
}

func citySoldierDraftIntervalSeconds(baseTime int64, speedRate float64) int64 {
	if baseTime <= 0 {
		return 1
	}
	duration := int64(math.Floor(float64(baseTime) * speedRate / citySoldierDraftGameSpeedRate))
	if duration < 1 {
		return 1
	}
	return duration
}

func citySoldierDraftLegacyNeedTime(interval int64) int64 {
	if interval <= 0 {
		return 0
	}
	return citySoldierDraftNeedTime
}

func cityDraftQueueStateLabel(state int) string {
	if state == 1 {
		return "Training"
	}
	return "Queued"
}

func (r *Repository) fixtureCityBarracksSnapshot(cid int, position int) CityBarracksSnapshot {
	return CityBarracksSnapshot{
		CID:           cid,
		Position:      position,
		Level:         1,
		QueueCapacity: 2,
		QueueCount:    0,
		FreePeople:    10,
		Options: []CityBarracksDraftOption{
			{
				SID:           1,
				Name:          "Worker",
				Description:   "Fixture soldier branch",
				Count:         100,
				WoodNeed:      150,
				IronNeed:      10,
				FoodNeed:      50,
				PeopleNeed:    1,
				DraftDuration: 6,
				CanDraft:      true,
				Conditions: []CityBarracksDraftCondition{
					{Type: "building", Name: "Barracks", RequiredLevel: 1, CurrentLevel: 1, Satisfied: true},
				},
			},
			{
				SID:           2,
				Name:          "Militia",
				Description:   "Fixture soldier branch",
				Count:         50,
				WoodNeed:      100,
				IronNeed:      50,
				FoodNeed:      80,
				PeopleNeed:    1,
				DraftDuration: 3,
				CanDraft:      true,
				Conditions: []CityBarracksDraftCondition{
					{Type: "building", Name: "Barracks", RequiredLevel: 1, CurrentLevel: 1, Satisfied: true},
				},
			},
		},
		Queue: []CityBarracksQueueItem{},
	}
}

func maxDraftInt64(left int64, right int64) int64 {
	if left > right {
		return left
	}
	return right
}
