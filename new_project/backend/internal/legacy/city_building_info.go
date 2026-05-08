package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"math"
)

func (r *Repository) CityBuildingInfo(ctx context.Context, uid int, cid int, position int) (CityBuildingInfo, error) {
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return CityBuildingInfo{}, err
	} else if !allowed {
		return CityBuildingInfo{}, ErrForbidden
	}

	if r.db == nil {
		detail := r.fixtureCityDetail(cid)
		for _, building := range detail.Buildings {
			if building.Position == position {
				return CityBuildingInfo{
					Building:  building,
					Current:   fixtureCityBuildingLevelInfo(building),
					Resources: detail.Summary.Resources,
				}, nil
			}
		}
		return CityBuildingInfo{}, sql.ErrNoRows
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return CityBuildingInfo{}, err
	}
	defer tx.Rollback()

	if err := r.settleCityBuildingQueueTx(ctx, tx, cid); err != nil {
		return CityBuildingInfo{}, fmt.Errorf("settle city building queue: %w", err)
	}

	building, err := r.loadCityBuildingInfoBuildingTx(ctx, tx, cid, position)
	if err != nil {
		return CityBuildingInfo{}, err
	}

	resources, err := r.loadCityResourceSnapshotTx(ctx, tx, cid)
	if err != nil {
		return CityBuildingInfo{}, err
	}

	currentLevel := building.Level
	if currentLevel <= 0 {
		currentLevel = 1
	}
	current, err := r.loadCityBuildingLevelInfoTx(ctx, tx, building.BID, currentLevel, 0)
	if err != nil {
		return CityBuildingInfo{}, err
	}

	response := CityBuildingInfo{
		Building:  building,
		Current:   current,
		Resources: resources,
	}

	cityType, err := r.loadCityTypeTx(ctx, tx, cid)
	if err != nil {
		return CityBuildingInfo{}, err
	}

	nextLevel := building.Level + 1
	if building.Level <= 0 {
		nextLevel = 1
	}
	if nextLevel > maxBuildingLevel(cityType, building.BID) {
		return response, tx.Commit()
	}

	speedRate, err := r.buildingSpeedRateTx(ctx, tx, cid)
	if err != nil {
		return CityBuildingInfo{}, err
	}

	next, err := r.loadCityBuildingLevelInfoTx(ctx, tx, building.BID, nextLevel, speedRate)
	if err != nil {
		if err == sql.ErrNoRows {
			return response, tx.Commit()
		}
		return CityBuildingInfo{}, err
	}

	if discountActive, err := r.hasBuildingDiscountBufferTx(ctx, tx, uid, nextLevel); err != nil {
		return CityBuildingInfo{}, err
	} else if discountActive {
		next.WoodNeed = int64(float64(next.WoodNeed) * 0.7)
		next.RockNeed = int64(float64(next.RockNeed) * 0.7)
		next.IronNeed = int64(float64(next.IronNeed) * 0.7)
		next.FoodNeed = int64(float64(next.FoodNeed) * 0.7)
	}

	conditions, err := r.loadCityBuildingUpgradeConditionsInfoTx(ctx, tx, uid, cid, building.BID, nextLevel)
	if err != nil {
		return CityBuildingInfo{}, err
	}
	next.Conditions = conditions
	next.CanUpgrade = true
	for _, condition := range conditions {
		if !condition.CanUpgrade {
			next.CanUpgrade = false
			next.Reason = "building prerequisite is not met"
			break
		}
	}
	if building.State != 0 {
		next.CanUpgrade = false
		next.Reason = "building is already busy"
	} else if err := r.ensureBuildingQueueCapacityTx(ctx, tx, uid, cid); err != nil {
		next.CanUpgrade = false
		next.Reason = err.Error()
	} else if resources.People < next.PeopleNeed {
		next.CanUpgrade = false
		next.Reason = "city people are not enough for this upgrade"
	} else if resources.Wood < next.WoodNeed ||
		resources.Rock < next.RockNeed ||
		resources.Iron < next.IronNeed ||
		resources.Food < next.FoodNeed ||
		resources.Gold < next.GoldNeed {
		next.CanUpgrade = false
		next.Reason = "city resources are not enough"
	}

	response.Next = &next
	return response, tx.Commit()
}

func (r *Repository) loadCityBuildingInfoBuildingTx(ctx context.Context, tx *sql.Tx, cid int, position int) (Building, error) {
	item := Building{}
	err := tx.QueryRowContext(ctx, `
select b.bid, cb.name, b.level, b.xy, b.state, b.state_starttime, b.state_endtime
from sys_building b
join cfg_building cb on cb.bid = b.bid
where b.cid = ? and b.xy = ?`, cid, position).Scan(
		&item.BID,
		&item.Name,
		&item.Level,
		&item.Position,
		&item.State,
		&item.StateStartTime,
		&item.StateEndTime,
	)
	if err != nil {
		return Building{}, err
	}
	item.Name = legacyBuildingName(item.BID, item.Name)
	return item, nil
}

func (r *Repository) loadCityBuildingLevelInfoTx(ctx context.Context, tx *sql.Tx, bid int, level int, speedRate float64) (CityBuildingLevelInfo, error) {
	info := CityBuildingLevelInfo{Conditions: []CityBuildingUpgradeCondition{}}
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
select b.bid, b.name, b.description, l.level, l.description,
	l.upgrade_wood, l.upgrade_rock, l.upgrade_iron, l.upgrade_food, l.upgrade_gold, l.upgrade_people, l.upgrade_time
from cfg_building b
join cfg_building_level l on l.bid = b.bid
where b.bid = ? and l.level = ?`, bid, level).Scan(
		&info.BID,
		&info.Name,
		&info.Description,
		&info.Level,
		&info.LevelDescription,
		&woodRaw,
		&rockRaw,
		&ironRaw,
		&foodRaw,
		&goldRaw,
		&peopleRaw,
		&timeRaw,
	)
	if err != nil {
		return CityBuildingLevelInfo{}, err
	}
	info.Name = legacyBuildingName(info.BID, info.Name)
	info.WoodNeed = int64(math.Round(woodRaw))
	info.RockNeed = int64(math.Round(rockRaw))
	info.IronNeed = int64(math.Round(ironRaw))
	info.FoodNeed = int64(math.Round(foodRaw))
	info.GoldNeed = int64(math.Round(goldRaw))
	info.PeopleNeed = int64(math.Round(peopleRaw))
	info.UpgradeTime = int64(math.Round(timeRaw))
	if speedRate > 0 {
		info.UpgradeTime = buildingDisplayDurationSeconds(info.UpgradeTime, speedRate)
	}
	info.CanUpgrade = true
	return info, nil
}

func (r *Repository) loadCityBuildingUpgradeConditionsInfoTx(ctx context.Context, tx *sql.Tx, uid int, cid int, bid int, nextLevel int) ([]CityBuildingUpgradeCondition, error) {
	rows, err := tx.QueryContext(ctx, `
select pre_type, pre_id, pre_level
from cfg_building_condition
where bid = ? and levelid = ?
order by pre_type, pre_id`, bid, nextLevel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conditions := make([]CityBuildingUpgradeCondition, 0, 4)
	for rows.Next() {
		var item CityBuildingUpgradeCondition
		if err := rows.Scan(&item.PreType, &item.PreID, &item.UpgradeNeed); err != nil {
			return nil, err
		}
		item.CanUpgrade = true
		switch item.PreType {
		case 0:
			item.Type = r.conditionBuildingNameTx(ctx, tx, item.PreID)
			if err := tx.QueryRowContext(ctx, `
select coalesce(max(level), 0)
from sys_building
where cid = ? and bid = ?`, cid, item.PreID).Scan(&item.CurrentOwn); err != nil {
				return nil, err
			}
		case 1:
			item.Type = r.conditionTechnicNameTx(ctx, tx, item.PreID)
			if err := tx.QueryRowContext(ctx, `
select coalesce(max(level), 0)
from sys_technic
where uid = ? and tid = ?`, uid, item.PreID).Scan(&item.CurrentOwn); err != nil {
				return nil, err
			}
		case 2:
			item.Type = r.conditionGoodsNameTx(ctx, tx, item.PreID)
			if err := tx.QueryRowContext(ctx, `
select coalesce(max(count), 0)
from sys_goods
where uid = ? and gid = ?`, uid, item.PreID).Scan(&item.CurrentOwn); err != nil {
				return nil, err
			}
		default:
			item.Type = "unknown"
		}
		item.CanUpgrade = item.CurrentOwn >= item.UpgradeNeed
		conditions = append(conditions, item)
	}
	return conditions, rows.Err()
}

func (r *Repository) loadCityResourceSnapshotTx(ctx context.Context, tx *sql.Tx, cid int) (ResourceSnapshot, error) {
	snapshot := ResourceSnapshot{}
	err := tx.QueryRowContext(ctx, `
select wood, rock, iron, food, gold, people, people_max, people_stable, people_building, food_max, wood_max, rock_max, iron_max, gold_max
from mem_city_resource
where cid = ?`, cid).Scan(
		&snapshot.Wood,
		&snapshot.Rock,
		&snapshot.Iron,
		&snapshot.Food,
		&snapshot.Gold,
		&snapshot.People,
		&snapshot.PeopleMax,
		&snapshot.PeopleStable,
		&snapshot.PeopleBuilding,
		&snapshot.FoodMax,
		&snapshot.WoodMax,
		&snapshot.RockMax,
		&snapshot.IronMax,
		&snapshot.GoldMax,
	)
	return snapshot, err
}

func (r *Repository) conditionBuildingNameTx(ctx context.Context, tx *sql.Tx, bid int) string {
	var name string
	if err := tx.QueryRowContext(ctx, "select name from cfg_building where bid = ?", bid).Scan(&name); err != nil {
		return fmt.Sprintf("building:%d", bid)
	}
	return legacyBuildingName(bid, name)
}

func (r *Repository) conditionTechnicNameTx(ctx context.Context, tx *sql.Tx, tid int) string {
	var name string
	if err := tx.QueryRowContext(ctx, "select name from cfg_technic where tid = ?", tid).Scan(&name); err != nil {
		return fmt.Sprintf("technic:%d", tid)
	}
	return name
}

func (r *Repository) conditionGoodsNameTx(ctx context.Context, tx *sql.Tx, gid int) string {
	var name string
	if err := tx.QueryRowContext(ctx, "select name from cfg_goods where gid = ?", gid).Scan(&name); err != nil {
		return fmt.Sprintf("goods:%d", gid)
	}
	return name
}

func fixtureCityBuildingLevelInfo(building Building) CityBuildingLevelInfo {
	return CityBuildingLevelInfo{
		BID:              building.BID,
		Name:             building.Name,
		Description:      building.Name,
		Level:            building.Level,
		LevelDescription: building.Name,
		Conditions:       []CityBuildingUpgradeCondition{},
		CanUpgrade:       building.State == 0,
	}
}
