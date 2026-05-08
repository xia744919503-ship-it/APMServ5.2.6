package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"math"
)

const (
	buildingQueueBufferType        = 11
	buildingDiscountBasicBuffer    = 12
	buildingDiscountAdvancedBuffer = 13
	buildingDiscountMasterBuffer   = 14
	houseCapacityBuildingID        = 5
	governmentCapacityBuildingID   = 6
)

type cityBuildingTarget struct {
	ID       int
	BID      int
	Level    int
	Position int
	State    int
}

type cityBuildingUpgradeConfig struct {
	Wood          int64
	Rock          int64
	Iron          int64
	Food          int64
	Gold          int64
	UpgradePeople int64
	UpgradeTime   int64
}

type cityBuildingCondition struct {
	PreType  int
	PreID    int
	PreLevel int64
}

type cityBuildingQueueRow struct {
	ID           int
	CID          int
	Position     int
	BID          int
	Level        int
	StateEndTime int64
}

func (r *Repository) UpgradeCityBuilding(ctx context.Context, uid int, cid int, position int) (CityDetail, error) {
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return CityDetail{}, err
	} else if !allowed {
		return CityDetail{}, ErrForbidden
	}

	if r.db == nil {
		return r.fixtureCityDetail(cid), nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return CityDetail{}, err
	}
	defer tx.Rollback()

	if err := r.settleCityBuildingQueueTx(ctx, tx, cid); err != nil {
		return CityDetail{}, fmt.Errorf("settle city building queue: %w", err)
	}

	target, err := r.loadCityBuildingTargetTx(ctx, tx, cid, position)
	if err != nil {
		if err == sql.ErrNoRows {
			return CityDetail{}, newInvalidError("building does not exist")
		}
		return CityDetail{}, fmt.Errorf("load target building: %w", err)
	}
	if target.State != 0 {
		return CityDetail{}, newInvalidError("building is already busy")
	}

	cityType, err := r.loadCityTypeTx(ctx, tx, cid)
	if err != nil {
		return CityDetail{}, fmt.Errorf("load city type: %w", err)
	}

	nextLevel := target.Level + 1
	if nextLevel > maxBuildingLevel(cityType, target.BID) {
		return CityDetail{}, newInvalidError("building has reached the level cap")
	}

	config, err := r.loadCityBuildingUpgradeConfigTx(ctx, tx, target.BID, nextLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			return CityDetail{}, newInvalidError("next building level is not configured")
		}
		return CityDetail{}, fmt.Errorf("load building upgrade config: %w", err)
	}

	if discountActive, err := r.hasBuildingDiscountBufferTx(ctx, tx, uid, nextLevel); err != nil {
		return CityDetail{}, fmt.Errorf("load building discount buffer: %w", err)
	} else if discountActive {
		config.Wood = int64(float64(config.Wood) * 0.7)
		config.Rock = int64(float64(config.Rock) * 0.7)
		config.Iron = int64(float64(config.Iron) * 0.7)
		config.Food = int64(float64(config.Food) * 0.7)
	}

	if err := r.ensureBuildingQueueCapacityTx(ctx, tx, uid, cid); err != nil {
		return CityDetail{}, fmt.Errorf("check building queue capacity: %w", err)
	}
	if err := r.ensureCityPeopleForUpgradeTx(ctx, tx, cid, config.UpgradePeople); err != nil {
		return CityDetail{}, fmt.Errorf("check city people for upgrade: %w", err)
	}

	goodsRequirements, err := r.validateBuildingUpgradeConditionsTx(ctx, tx, uid, cid, target.BID, nextLevel)
	if err != nil {
		return CityDetail{}, fmt.Errorf("validate building upgrade conditions: %w", err)
	}

	now, err := r.currentUnixTimeTx(ctx, tx)
	if err != nil {
		return CityDetail{}, fmt.Errorf("load current unix time: %w", err)
	}
	speedRate, err := r.buildingSpeedRateTx(ctx, tx, cid)
	if err != nil {
		return CityDetail{}, fmt.Errorf("load building speed rate: %w", err)
	}

	if err := r.reserveCityUpgradeResourcesTx(ctx, tx, cid, config); err != nil {
		return CityDetail{}, fmt.Errorf("reserve city upgrade resources: %w", err)
	}
	if err := r.consumeUpgradeGoodsTx(ctx, tx, uid, goodsRequirements); err != nil {
		return CityDetail{}, fmt.Errorf("consume building upgrade goods: %w", err)
	}

	finishAt := now + buildingQueueDurationSeconds(config.UpgradeTime, speedRate)
	result, err := tx.ExecContext(ctx, `
update sys_building
set state = 1, state_starttime = ?, state_endtime = ?
where id = ? and cid = ? and state = 0`, now, finishAt, target.ID, cid)
	if err != nil {
		return CityDetail{}, fmt.Errorf("mark building upgrade in progress: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return CityDetail{}, err
	}
	if affected == 0 {
		return CityDetail{}, newInvalidError("building state changed before the upgrade started")
	}

	if _, err := tx.ExecContext(ctx, `
insert into mem_building_upgrading (id, cid, xy, bid, level, state_endtime)
values (?, ?, ?, ?, ?, ?)
on duplicate key update
	cid = values(cid),
	xy = values(xy),
	bid = values(bid),
	level = values(level),
	state_endtime = values(state_endtime)`,
		target.ID,
		cid,
		position,
		target.BID,
		nextLevel,
		finishAt,
	); err != nil {
		return CityDetail{}, fmt.Errorf("enqueue building upgrade: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return CityDetail{}, fmt.Errorf("commit building upgrade: %w", err)
	}

	return r.CityDetail(ctx, cid)
}

func (r *Repository) DestroyCityBuilding(ctx context.Context, uid int, cid int, position int) (CityDetail, error) {
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return CityDetail{}, err
	} else if !allowed {
		return CityDetail{}, ErrForbidden
	}

	if r.db == nil {
		return r.fixtureCityDetail(cid), nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return CityDetail{}, err
	}
	defer tx.Rollback()

	if err := r.settleCityBuildingQueueTx(ctx, tx, cid); err != nil {
		return CityDetail{}, fmt.Errorf("settle city building queue: %w", err)
	}

	target, err := r.loadCityBuildingTargetTx(ctx, tx, cid, position)
	if err != nil {
		if err == sql.ErrNoRows {
			return CityDetail{}, newInvalidError("building does not exist")
		}
		return CityDetail{}, fmt.Errorf("load target building: %w", err)
	}

	switch target.State {
	case 0:
	case 1:
		return CityDetail{}, newInvalidError("building is already upgrading")
	case 2:
		return CityDetail{}, newInvalidError("building is already demolishing")
	default:
		return CityDetail{}, newInvalidError("building is already busy")
	}

	if err := r.ensureBuildingQueueCapacityTx(ctx, tx, uid, cid); err != nil {
		return CityDetail{}, fmt.Errorf("check building queue capacity: %w", err)
	}

	if target.BID == governmentCapacityBuildingID && target.Level == 1 {
		return CityDetail{}, newInvalidError("government hall level 1 cannot be demolished")
	}

	config, err := r.loadCityBuildingUpgradeConfigTx(ctx, tx, target.BID, target.Level)
	if err != nil {
		if err == sql.ErrNoRows {
			return CityDetail{}, newInvalidError("current building level is not configured")
		}
		return CityDetail{}, fmt.Errorf("load building destroy config: %w", err)
	}

	now, err := r.currentUnixTimeTx(ctx, tx)
	if err != nil {
		return CityDetail{}, fmt.Errorf("load current unix time: %w", err)
	}
	speedRate, err := r.buildingSpeedRateTx(ctx, tx, cid)
	if err != nil {
		return CityDetail{}, fmt.Errorf("load building speed rate: %w", err)
	}

	finishAt := now + buildingDestroyDurationSeconds(config.UpgradeTime, speedRate)
	result, err := tx.ExecContext(ctx, `
update sys_building
set state = 2, state_starttime = ?, state_endtime = ?
where id = ? and cid = ? and state = 0`, now, finishAt, target.ID, cid)
	if err != nil {
		return CityDetail{}, fmt.Errorf("mark building demolish in progress: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return CityDetail{}, err
	}
	if affected == 0 {
		return CityDetail{}, newInvalidError("building state changed before the demolish started")
	}

	if _, err := tx.ExecContext(ctx, `
insert into mem_building_destroying (id, cid, xy, bid, level, state_endtime)
values (?, ?, ?, ?, ?, ?)
on duplicate key update
	cid = values(cid),
	xy = values(xy),
	bid = values(bid),
	level = values(level),
	state_endtime = values(state_endtime)`,
		target.ID,
		cid,
		position,
		target.BID,
		target.Level-1,
		finishAt,
	); err != nil {
		return CityDetail{}, fmt.Errorf("enqueue building demolish: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return CityDetail{}, fmt.Errorf("commit building demolish: %w", err)
	}

	return r.CityDetail(ctx, cid)
}

func (r *Repository) CancelCityBuildingAction(ctx context.Context, uid int, cid int, position int) (CityDetail, error) {
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return CityDetail{}, err
	} else if !allowed {
		return CityDetail{}, ErrForbidden
	}

	if r.db == nil {
		return r.fixtureCityDetail(cid), nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return CityDetail{}, err
	}
	defer tx.Rollback()

	if err := r.settleCityBuildingQueueTx(ctx, tx, cid); err != nil {
		return CityDetail{}, fmt.Errorf("settle city building queue: %w", err)
	}

	target, err := r.loadCityBuildingTargetTx(ctx, tx, cid, position)
	if err != nil {
		if err == sql.ErrNoRows {
			return CityDetail{}, newInvalidError("building does not exist")
		}
		return CityDetail{}, fmt.Errorf("load target building: %w", err)
	}

	switch target.State {
	case 1:
		config, err := r.loadCityBuildingUpgradeConfigTx(ctx, tx, target.BID, target.Level+1)
		if err != nil {
			if err == sql.ErrNoRows {
				return CityDetail{}, newInvalidError("next building level is not configured")
			}
			return CityDetail{}, fmt.Errorf("load building cancel config: %w", err)
		}

		if target.Level == 0 {
			result, err := tx.ExecContext(ctx, `
delete from sys_building
where id = ? and cid = ? and state = 1`, target.ID, cid)
			if err != nil {
				return CityDetail{}, fmt.Errorf("delete pending building: %w", err)
			}
			affected, err := result.RowsAffected()
			if err != nil {
				return CityDetail{}, err
			}
			if affected == 0 {
				return CityDetail{}, newInvalidError("building state changed before the cancel completed")
			}
		} else {
			result, err := tx.ExecContext(ctx, `
update sys_building
set state = 0, state_starttime = 0, state_endtime = 0
where id = ? and cid = ? and state = 1`, target.ID, cid)
			if err != nil {
				return CityDetail{}, fmt.Errorf("reset building upgrade state: %w", err)
			}
			affected, err := result.RowsAffected()
			if err != nil {
				return CityDetail{}, err
			}
			if affected == 0 {
				return CityDetail{}, newInvalidError("building state changed before the cancel completed")
			}
		}

		if _, err := tx.ExecContext(ctx, `
delete from mem_building_upgrading
where id = ?`, target.ID); err != nil {
			return CityDetail{}, fmt.Errorf("delete building upgrade queue: %w", err)
		}
		if err := r.refundCityBuildingResourcesTx(ctx, tx, cid, config); err != nil {
			return CityDetail{}, fmt.Errorf("refund building cancel resources: %w", err)
		}
	case 2:
		result, err := tx.ExecContext(ctx, `
update sys_building
set state = 0, state_starttime = 0, state_endtime = 0
where id = ? and cid = ? and state = 2`, target.ID, cid)
		if err != nil {
			return CityDetail{}, fmt.Errorf("reset building demolish state: %w", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return CityDetail{}, err
		}
		if affected == 0 {
			return CityDetail{}, newInvalidError("building state changed before the cancel completed")
		}

		if _, err := tx.ExecContext(ctx, `
delete from mem_building_destroying
where id = ?`, target.ID); err != nil {
			return CityDetail{}, fmt.Errorf("delete building demolish queue: %w", err)
		}
	case 0:
		return CityDetail{}, newInvalidError("building is idle and cannot be cancelled")
	default:
		return CityDetail{}, newInvalidError("building action cannot be cancelled")
	}

	if err := tx.Commit(); err != nil {
		return CityDetail{}, fmt.Errorf("commit building cancel: %w", err)
	}

	return r.CityDetail(ctx, cid)
}

func (r *Repository) settleCityBuildingQueue(ctx context.Context, cid int) error {
	if r.db == nil {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := r.settleCityBuildingQueueTx(ctx, tx, cid); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) settleCityBuildingQueueTx(ctx context.Context, tx *sql.Tx, cid int) error {
	now, err := r.currentUnixTimeTx(ctx, tx)
	if err != nil {
		return err
	}

	upgradeRows, err := r.loadCityBuildingQueueRowsTx(ctx, tx, "mem_building_upgrading", cid, now)
	if err != nil {
		return err
	}
	destroyRows, err := r.loadCityBuildingQueueRowsTx(ctx, tx, "mem_building_destroying", cid, now)
	if err != nil {
		return err
	}
	if len(upgradeRows) == 0 && len(destroyRows) == 0 {
		return nil
	}

	for _, item := range upgradeRows {
		if _, err := tx.ExecContext(ctx, `
update sys_building
set level = ?, state = 0, state_starttime = 0, state_endtime = 0
where id = ? and cid = ?`, item.Level, item.ID, cid); err != nil {
			return err
		}
	}
	if len(upgradeRows) > 0 {
		if _, err := tx.ExecContext(ctx, `
delete from mem_building_upgrading
where cid = ? and state_endtime <= ?`, cid, now); err != nil {
			return err
		}
	}

	for _, item := range destroyRows {
		if item.Level <= 0 {
			if _, err := tx.ExecContext(ctx, `
delete from sys_building
where id = ? and cid = ?`, item.ID, cid); err != nil {
				return err
			}
			continue
		}

		if _, err := tx.ExecContext(ctx, `
update sys_building
set level = ?, state = 0, state_starttime = 0, state_endtime = 0
where id = ? and cid = ?`, item.Level, item.ID, cid); err != nil {
			return err
		}
	}
	if len(destroyRows) > 0 {
		if _, err := tx.ExecContext(ctx, `
delete from mem_building_destroying
where cid = ? and state_endtime <= ?`, cid, now); err != nil {
			return err
		}
	}

	if err := r.ensureCityResAdd(ctx, tx, cid); err != nil {
		return err
	}
	if err := r.recalculateCityProduction(ctx, tx, cid); err != nil {
		return err
	}
	if err := r.recalculateCityPeopleMaxTx(ctx, tx, cid); err != nil {
		return err
	}
	if err := r.recalculateCityGoldMaxTx(ctx, tx, cid); err != nil {
		return err
	}
	return nil
}

func (r *Repository) loadCityBuildingQueueRowsTx(ctx context.Context, tx *sql.Tx, table string, cid int, now int64) ([]cityBuildingQueueRow, error) {
	query := fmt.Sprintf(`
select id, cid, xy, bid, level, state_endtime
from %s
where cid = ? and state_endtime <= ?
order by state_endtime, id`, table)

	rows, err := tx.QueryContext(ctx, query, cid, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	queueRows := make([]cityBuildingQueueRow, 0, 4)
	for rows.Next() {
		var item cityBuildingQueueRow
		if err := rows.Scan(&item.ID, &item.CID, &item.Position, &item.BID, &item.Level, &item.StateEndTime); err != nil {
			return nil, err
		}
		queueRows = append(queueRows, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return queueRows, nil
}

func (r *Repository) loadCityBuildingTargetTx(ctx context.Context, tx *sql.Tx, cid int, position int) (cityBuildingTarget, error) {
	target := cityBuildingTarget{}
	err := tx.QueryRowContext(ctx, `
select id, bid, level, xy, state
from sys_building
where cid = ? and xy = ?`, cid, position).Scan(
		&target.ID,
		&target.BID,
		&target.Level,
		&target.Position,
		&target.State,
	)
	return target, err
}

func (r *Repository) loadCityTypeTx(ctx context.Context, tx *sql.Tx, cid int) (int, error) {
	var cityType int
	err := tx.QueryRowContext(ctx, "select type from sys_city where cid = ?", cid).Scan(&cityType)
	return cityType, err
}

func (r *Repository) loadCityBuildingUpgradeConfigTx(ctx context.Context, tx *sql.Tx, bid int, level int) (cityBuildingUpgradeConfig, error) {
	config := cityBuildingUpgradeConfig{}
	var (
		woodRaw   float64
		rockRaw   float64
		ironRaw   float64
		foodRaw   float64
		goldRaw   float64
		peopleRaw float64
		timeRaw   float64
	)
	err := tx.QueryRowContext(ctx, `
select upgrade_wood, upgrade_rock, upgrade_iron, upgrade_food, upgrade_gold, upgrade_people, upgrade_time
from cfg_building_level
where bid = ? and level = ?`, bid, level).Scan(
		&woodRaw,
		&rockRaw,
		&ironRaw,
		&foodRaw,
		&goldRaw,
		&peopleRaw,
		&timeRaw,
	)
	if err != nil {
		return cityBuildingUpgradeConfig{}, err
	}
	config.Wood = int64(math.Round(woodRaw))
	config.Rock = int64(math.Round(rockRaw))
	config.Iron = int64(math.Round(ironRaw))
	config.Food = int64(math.Round(foodRaw))
	config.Gold = int64(math.Round(goldRaw))
	config.UpgradePeople = int64(math.Round(peopleRaw))
	config.UpgradeTime = int64(math.Round(timeRaw))
	return config, err
}

func (r *Repository) hasBuildingDiscountBufferTx(ctx context.Context, tx *sql.Tx, uid int, nextLevel int) (bool, error) {
	query := `
select count(*)
from mem_user_buffer
where uid = ? and buftype = ? and endtime > unix_timestamp()`
	args := []any{uid, buildingDiscountMasterBuffer}

	switch {
	case nextLevel <= 5:
		query = `
select count(*)
from mem_user_buffer
where uid = ? and buftype in (?, ?, ?) and endtime > unix_timestamp()`
		args = []any{uid, buildingDiscountBasicBuffer, buildingDiscountAdvancedBuffer, buildingDiscountMasterBuffer}
	case nextLevel <= 8:
		query = `
select count(*)
from mem_user_buffer
where uid = ? and buftype in (?, ?) and endtime > unix_timestamp()`
		args = []any{uid, buildingDiscountAdvancedBuffer, buildingDiscountMasterBuffer}
	}

	var count int
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) ensureBuildingQueueCapacityTx(ctx context.Context, tx *sql.Tx, uid int, cid int) error {
	var queueCount int
	if err := tx.QueryRowContext(ctx, `
select count(*)
from sys_building
where cid = ? and state > 0`, cid).Scan(&queueCount); err != nil {
		return err
	}

	limit := 2
	var hasBuffer int
	if err := tx.QueryRowContext(ctx, `
select count(*)
from mem_user_buffer
where uid = ? and buftype = ? and endtime > unix_timestamp()`, uid, buildingQueueBufferType).Scan(&hasBuffer); err != nil {
		return err
	}
	if hasBuffer > 0 {
		limit = 5
	}
	if queueCount >= limit {
		return newInvalidError("building upgrade queue is full")
	}
	return nil
}

func (r *Repository) ensureCityPeopleForUpgradeTx(ctx context.Context, tx *sql.Tx, cid int, requiredPeople int64) error {
	if requiredPeople <= 0 {
		return nil
	}

	var people int64
	if err := tx.QueryRowContext(ctx, `
select people
from mem_city_resource
where cid = ?`, cid).Scan(&people); err != nil {
		return err
	}
	if people < requiredPeople {
		return newInvalidError("city people are not enough for this upgrade")
	}
	return nil
}

func (r *Repository) validateBuildingUpgradeConditionsTx(ctx context.Context, tx *sql.Tx, uid int, cid int, bid int, nextLevel int) ([]cityBuildingCondition, error) {
	rows, err := tx.QueryContext(ctx, `
select pre_type, pre_id, pre_level
from cfg_building_condition
where bid = ? and levelid = ?
order by pre_type, pre_id`, bid, nextLevel)
	if err != nil {
		return nil, err
	}

	conditions := make([]cityBuildingCondition, 0, 4)
	for rows.Next() {
		var condition cityBuildingCondition
		if err := rows.Scan(&condition.PreType, &condition.PreID, &condition.PreLevel); err != nil {
			rows.Close()
			return nil, err
		}
		conditions = append(conditions, condition)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	goodsRequirements := make([]cityBuildingCondition, 0, 2)
	for _, condition := range conditions {
		switch condition.PreType {
		case 0:
			var buildingLevel int64
			if err := tx.QueryRowContext(ctx, `
select coalesce(max(level), 0)
from sys_building
where cid = ? and bid = ?`, cid, condition.PreID).Scan(&buildingLevel); err != nil {
				return nil, err
			}
			if buildingLevel < condition.PreLevel {
				return nil, newInvalidError("building prerequisite is not met")
			}
		case 1:
			var technicLevel int64
			if err := tx.QueryRowContext(ctx, `
select coalesce(max(level), 0)
from sys_technic
where uid = ? and tid = ?`, uid, condition.PreID).Scan(&technicLevel); err != nil {
				return nil, err
			}
			if technicLevel < condition.PreLevel {
				return nil, newInvalidError("technic prerequisite is not met")
			}
		case 2:
			var goodsCount int64
			if err := tx.QueryRowContext(ctx, `
select coalesce(max(count), 0)
from sys_goods
where uid = ? and gid = ?`, uid, condition.PreID).Scan(&goodsCount); err != nil {
				return nil, err
			}
			if goodsCount < condition.PreLevel {
				return nil, newInvalidError("goods prerequisite is not met")
			}
			goodsRequirements = append(goodsRequirements, condition)
		default:
			return nil, newInvalidError("unsupported building prerequisite")
		}
	}

	return goodsRequirements, nil
}

func (r *Repository) reserveCityUpgradeResourcesTx(ctx context.Context, tx *sql.Tx, cid int, config cityBuildingUpgradeConfig) error {
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
		return newInvalidError("city resources are not enough")
	}
	return nil
}

func (r *Repository) consumeUpgradeGoodsTx(ctx context.Context, tx *sql.Tx, uid int, requirements []cityBuildingCondition) error {
	for _, item := range requirements {
		if item.PreLevel <= 0 {
			continue
		}
		if err := r.adjustUserGoodsTx(ctx, tx, uid, item.PreID, -item.PreLevel); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) recalculateCityPeopleMaxTx(ctx context.Context, tx *sql.Tx, cid int) error {
	var peopleMax int64
	if err := tx.QueryRowContext(ctx, `
select coalesce(sum(level * (level + 1) * 500), 0)
from sys_building
where cid = ? and bid = ?`, cid, houseCapacityBuildingID).Scan(&peopleMax); err != nil {
		return err
	}

	_, err := tx.ExecContext(ctx, `
update mem_city_resource
set people_max = ?, people_stable = round(? * morale_stable * 0.01)
where cid = ?`, peopleMax, peopleMax, cid)
	return err
}

func (r *Repository) recalculateCityGoldMaxTx(ctx context.Context, tx *sql.Tx, cid int) error {
	var goldMax int64
	if err := tx.QueryRowContext(ctx, `
select coalesce(max(level * (level + 1) * 5000000), 0)
from sys_building
where cid = ? and bid = ?`, cid, governmentCapacityBuildingID).Scan(&goldMax); err != nil {
		return err
	}

	_, err := tx.ExecContext(ctx, `
update mem_city_resource
set gold_max = ?
where cid = ?`, goldMax, cid)
	return err
}

func (r *Repository) currentUnixTimeTx(ctx context.Context, tx *sql.Tx) (int64, error) {
	var now int64
	err := tx.QueryRowContext(ctx, "select unix_timestamp()").Scan(&now)
	return now, err
}

func maxBuildingLevel(cityType int, bid int) int {
	if bid >= 5 {
		return 10
	}

	switch cityType {
	case 1:
		return 12
	case 2:
		return 15
	case 3:
		return 18
	case 4:
		return 20
	default:
		return 10
	}
}

func maxInt64(left int64, right int64) int64 {
	if left > right {
		return left
	}
	return right
}

func (r *Repository) refundCityBuildingResourcesTx(ctx context.Context, tx *sql.Tx, cid int, config cityBuildingUpgradeConfig) error {
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

func buildingCancelRefundAmount(amount int64) int64 {
	if amount <= 0 {
		return 0
	}
	return int64(math.Floor(float64(amount) * 0.66))
}
