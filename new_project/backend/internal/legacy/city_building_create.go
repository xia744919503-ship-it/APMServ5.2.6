package legacy

import (
	"context"
	"database/sql"
)

var cityBuildingFieldSlotUnlockLevels = map[int]int{
	1:  5,
	2:  8,
	10: 1,
	11: 1,
	12: 5,
	13: 8,
	21: 1,
	22: 1,
	23: 5,
	24: 8,
	31: 1,
	32: 1,
	33: 3,
	34: 6,
	35: 9,
	41: 1,
	42: 1,
	43: 3,
	44: 6,
	45: 7,
	46: 9,
	51: 1,
	52: 1,
	53: 3,
	54: 6,
	55: 7,
	56: 9,
	60: 1,
	61: 1,
	62: 2,
	63: 4,
	64: 7,
	65: 7,
	70: 2,
	71: 2,
	72: 4,
	73: 7,
	81: 4,
	82: 7,
}

var cityBuildingInnerSlotSet = map[int]struct{}{
	0:   {},
	100: {},
	101: {},
	102: {},
	103: {},
	104: {},
	105: {},
	110: {},
	111: {},
	112: {},
	113: {},
	114: {},
	115: {},
	120: {},
	122: {},
	123: {},
	124: {},
	125: {},
	132: {},
	133: {},
	134: {},
	135: {},
	140: {},
	141: {},
	142: {},
	143: {},
	144: {},
	145: {},
	150: {},
	151: {},
	152: {},
	153: {},
	154: {},
	155: {},
	199: {},
}

var repeatableCityBuildingIDs = map[int]struct{}{
	1:  {},
	2:  {},
	3:  {},
	4:  {},
	5:  {},
	9:  {},
	17: {},
}

type cityBuildingCandidate struct {
	BID         int
	Name        string
	Description string
}

func (r *Repository) CityBuildingPlacementOptions(ctx context.Context, uid int, cid int, position int) (CityBuildingPlacementOptions, error) {
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return CityBuildingPlacementOptions{}, err
	} else if !allowed {
		return CityBuildingPlacementOptions{}, ErrForbidden
	}

	slot, err := r.cityBuildingSlot(position, 10)
	if err != nil {
		return CityBuildingPlacementOptions{}, err
	}

	if r.db == nil {
		return CityBuildingPlacementOptions{
			Slot: slot,
		}, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return CityBuildingPlacementOptions{}, err
	}
	defer tx.Rollback()

	if err := r.settleCityBuildingQueueTx(ctx, tx, cid); err != nil {
		return CityBuildingPlacementOptions{}, err
	}

	governmentLevel, err := r.loadCityGovernmentLevelTx(ctx, tx, cid)
	if err != nil {
		return CityBuildingPlacementOptions{}, err
	}

	slot, err = r.cityBuildingSlot(position, governmentLevel)
	if err != nil {
		return CityBuildingPlacementOptions{}, err
	}

	occupied, err := r.cityBuildingSlotOccupiedTx(ctx, tx, cid, position)
	if err != nil {
		return CityBuildingPlacementOptions{}, err
	}
	slot.Occupied = occupied

	response := CityBuildingPlacementOptions{
		Slot:    slot,
		Options: []CityBuildingPlacementOption{},
	}

	if !occupied {
		candidates, err := r.listCityBuildingCandidatesTx(ctx, tx, cid, slot)
		if err != nil {
			return CityBuildingPlacementOptions{}, err
		}

		globalReason := ""
		if !slot.Unlocked {
			globalReason = "government level is not enough for this field slot"
		} else if err := r.ensureBuildingQueueCapacityTx(ctx, tx, uid, cid); err != nil {
			globalReason = err.Error()
		}
		speedRate, err := r.buildingSpeedRateTx(ctx, tx, cid)
		if err != nil {
			return CityBuildingPlacementOptions{}, err
		}

		options := make([]CityBuildingPlacementOption, 0, len(candidates))
		for _, candidate := range candidates {
			option, err := r.buildCityBuildingPlacementOptionTx(ctx, tx, uid, cid, slot, candidate, globalReason, speedRate)
			if err != nil {
				return CityBuildingPlacementOptions{}, err
			}
			options = append(options, option)
		}
		response.Options = options
	}

	if err := tx.Commit(); err != nil {
		return CityBuildingPlacementOptions{}, err
	}

	return response, nil
}

func (r *Repository) CreateCityBuilding(ctx context.Context, uid int, cid int, position int, bid int) (CityDetail, error) {
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
		return CityDetail{}, err
	}

	governmentLevel, err := r.loadCityGovernmentLevelTx(ctx, tx, cid)
	if err != nil {
		return CityDetail{}, err
	}

	slot, err := r.cityBuildingSlot(position, governmentLevel)
	if err != nil {
		return CityDetail{}, err
	}

	config, goodsRequirements, err := r.validateCityBuildingCreateTx(ctx, tx, uid, cid, slot, bid)
	if err != nil {
		return CityDetail{}, err
	}

	now, err := r.currentUnixTimeTx(ctx, tx)
	if err != nil {
		return CityDetail{}, err
	}
	speedRate, err := r.buildingSpeedRateTx(ctx, tx, cid)
	if err != nil {
		return CityDetail{}, err
	}

	if err := r.reserveCityUpgradeResourcesTx(ctx, tx, cid, config); err != nil {
		return CityDetail{}, err
	}
	if err := r.consumeUpgradeGoodsTx(ctx, tx, uid, goodsRequirements); err != nil {
		return CityDetail{}, err
	}

	finishAt := now + buildingQueueDurationSeconds(config.UpgradeTime, speedRate)
	result, err := tx.ExecContext(ctx, `
insert into sys_building (cid, xy, bid, level, state, state_starttime, state_endtime)
values (?, ?, ?, 0, 1, ?, ?)`, cid, position, bid, now, finishAt)
	if err != nil {
		return CityDetail{}, err
	}

	buildingID, err := result.LastInsertId()
	if err != nil {
		return CityDetail{}, err
	}

	if _, err := tx.ExecContext(ctx, `
insert into mem_building_upgrading (id, cid, xy, bid, level, state_endtime)
values (?, ?, ?, ?, ?, ?)`,
		buildingID,
		cid,
		position,
		bid,
		1,
		finishAt,
	); err != nil {
		return CityDetail{}, err
	}

	if err := tx.Commit(); err != nil {
		return CityDetail{}, err
	}

	return r.CityDetail(ctx, cid)
}

func (r *Repository) validateCityBuildingCreateTx(ctx context.Context, tx *sql.Tx, uid int, cid int, slot CityBuildingPlacementSlot, bid int) (cityBuildingUpgradeConfig, []cityBuildingCondition, error) {
	if !slot.Unlocked {
		return cityBuildingUpgradeConfig{}, nil, newInvalidError("government level is not enough for this field slot")
	}

	if err := r.ensureCityBuildingSlotEmptyTx(ctx, tx, cid, slot.Position); err != nil {
		return cityBuildingUpgradeConfig{}, nil, err
	}

	if _, err := r.loadCityBuildingCandidateTx(ctx, tx, cid, slot, bid); err != nil {
		return cityBuildingUpgradeConfig{}, nil, err
	}

	config, err := r.loadCityBuildingUpgradeConfigTx(ctx, tx, bid, 1)
	if err != nil {
		if err == sql.ErrNoRows {
			return cityBuildingUpgradeConfig{}, nil, newInvalidError("next building level is not configured")
		}
		return cityBuildingUpgradeConfig{}, nil, err
	}

	if discountActive, err := r.hasBuildingDiscountBufferTx(ctx, tx, uid, 1); err != nil {
		return cityBuildingUpgradeConfig{}, nil, err
	} else if discountActive {
		config.Wood = int64(float64(config.Wood) * 0.7)
		config.Rock = int64(float64(config.Rock) * 0.7)
		config.Iron = int64(float64(config.Iron) * 0.7)
		config.Food = int64(float64(config.Food) * 0.7)
	}

	if err := r.ensureBuildingQueueCapacityTx(ctx, tx, uid, cid); err != nil {
		return cityBuildingUpgradeConfig{}, nil, err
	}
	if err := r.ensureCityPeopleForUpgradeTx(ctx, tx, cid, config.UpgradePeople); err != nil {
		return cityBuildingUpgradeConfig{}, nil, err
	}

	goodsRequirements, err := r.validateBuildingUpgradeConditionsTx(ctx, tx, uid, cid, bid, 1)
	if err != nil {
		return cityBuildingUpgradeConfig{}, nil, err
	}
	if err := r.ensureCityUpgradeResourcesAvailableTx(ctx, tx, cid, config); err != nil {
		return cityBuildingUpgradeConfig{}, nil, err
	}

	return config, goodsRequirements, nil
}

func (r *Repository) buildCityBuildingPlacementOptionTx(ctx context.Context, tx *sql.Tx, uid int, cid int, slot CityBuildingPlacementSlot, candidate cityBuildingCandidate, globalReason string, speedRate float64) (CityBuildingPlacementOption, error) {
	config, err := r.loadCityBuildingUpgradeConfigTx(ctx, tx, candidate.BID, 1)
	if err != nil {
		if err == sql.ErrNoRows {
			return CityBuildingPlacementOption{
				BID:         candidate.BID,
				Name:        candidate.Name,
				Description: candidate.Description,
				Level:       1,
				CanBuild:    false,
				Reason:      "next building level is not configured",
			}, nil
		}
		return CityBuildingPlacementOption{}, err
	}

	if discountActive, err := r.hasBuildingDiscountBufferTx(ctx, tx, uid, 1); err != nil {
		return CityBuildingPlacementOption{}, err
	} else if discountActive {
		config.Wood = int64(float64(config.Wood) * 0.7)
		config.Rock = int64(float64(config.Rock) * 0.7)
		config.Iron = int64(float64(config.Iron) * 0.7)
		config.Food = int64(float64(config.Food) * 0.7)
	}

	option := CityBuildingPlacementOption{
		BID:         candidate.BID,
		Name:        candidate.Name,
		Description: candidate.Description,
		Level:       1,
		Wood:        config.Wood,
		Rock:        config.Rock,
		Iron:        config.Iron,
		Food:        config.Food,
		Gold:        config.Gold,
		People:      config.UpgradePeople,
		Duration:    buildingDisplayDurationSeconds(config.UpgradeTime, speedRate),
		CanBuild:    true,
	}

	if globalReason != "" {
		option.CanBuild = false
		option.Reason = globalReason
		return option, nil
	}

	if err := r.ensureCityPeopleForUpgradeTx(ctx, tx, cid, config.UpgradePeople); err != nil {
		option.CanBuild = false
		option.Reason = err.Error()
		return option, nil
	}

	if _, err := r.validateBuildingUpgradeConditionsTx(ctx, tx, uid, cid, candidate.BID, 1); err != nil {
		option.CanBuild = false
		option.Reason = err.Error()
		return option, nil
	}

	if err := r.ensureCityUpgradeResourcesAvailableTx(ctx, tx, cid, config); err != nil {
		option.CanBuild = false
		option.Reason = err.Error()
		return option, nil
	}

	if !slot.Unlocked {
		option.CanBuild = false
		option.Reason = "government level is not enough for this field slot"
	}

	return option, nil
}

func (r *Repository) cityBuildingSlot(position int, governmentLevel int) (CityBuildingPlacementSlot, error) {
	if position == 199 {
		return CityBuildingPlacementSlot{
			Position:    position,
			Inner:       true,
			WallSlot:    true,
			Unlocked:    true,
			UnlockLevel: 1,
		}, nil
	}

	if unlockLevel, ok := cityBuildingFieldSlotUnlockLevels[position]; ok {
		return CityBuildingPlacementSlot{
			Position:    position,
			Unlocked:    governmentLevel >= unlockLevel,
			UnlockLevel: unlockLevel,
		}, nil
	}

	if _, ok := cityBuildingInnerSlotSet[position]; ok {
		return CityBuildingPlacementSlot{
			Position:    position,
			Inner:       true,
			Unlocked:    true,
			UnlockLevel: 1,
		}, nil
	}

	return CityBuildingPlacementSlot{}, newInvalidError("building slot is invalid")
}

func (r *Repository) listCityBuildingCandidatesTx(ctx context.Context, tx *sql.Tx, cid int, slot CityBuildingPlacementSlot) ([]cityBuildingCandidate, error) {
	query := `
select bid, name, description
from cfg_building
where (` + "`inner`" + ` = 0 or bid = 5)
order by bid`
	args := []any{}

	switch {
	case slot.WallSlot:
		query = `
select bid, name, description
from cfg_building
where bid = 20`
	case slot.Inner:
		query = `
select b.bid, b.name, b.description
from cfg_building b
where b.` + "`inner`" + ` = 1
	and b.bid <> 20
	and (
		b.bid in (5, 9, 17)
		or not exists (
			select 1
			from sys_building s
			where s.cid = ? and s.bid = b.bid
		)
	)
order by b.bid`
		args = append(args, cid)
	}

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]cityBuildingCandidate, 0, 8)
	for rows.Next() {
		var item cityBuildingCandidate
		if err := rows.Scan(&item.BID, &item.Name, &item.Description); err != nil {
			return nil, err
		}
		item.Name = legacyBuildingName(item.BID, item.Name)
		candidates = append(candidates, item)
	}

	return candidates, rows.Err()
}

func (r *Repository) loadCityBuildingCandidateTx(ctx context.Context, tx *sql.Tx, cid int, slot CityBuildingPlacementSlot, bid int) (cityBuildingCandidate, error) {
	candidate := cityBuildingCandidate{}
	query := `
select bid, name, description
from cfg_building
where bid = ? and ` + "`inner`" + ` = 0`
	args := []any{bid}

	switch {
	case slot.WallSlot:
		query = `
select bid, name, description
from cfg_building
where bid = 20 and bid = ?`
	case slot.Inner:
		query = `
select b.bid, b.name, b.description
from cfg_building b
where b.bid = ?
	and b.` + "`inner`" + ` = 1
	and b.bid <> 20
	and (
		b.bid in (5, 9, 17)
		or not exists (
			select 1
			from sys_building s
			where s.cid = ? and s.bid = b.bid
		)
	)`
		args = append(args, cid)
	}

	if err := tx.QueryRowContext(ctx, query, args...).Scan(&candidate.BID, &candidate.Name, &candidate.Description); err != nil {
		if err == sql.ErrNoRows {
			return cityBuildingCandidate{}, newInvalidError("building cannot be created in this slot")
		}
		return cityBuildingCandidate{}, err
	}
	candidate.Name = legacyBuildingName(candidate.BID, candidate.Name)

	return candidate, nil
}

func (r *Repository) loadCityGovernmentLevelTx(ctx context.Context, tx *sql.Tx, cid int) (int, error) {
	var level int
	err := tx.QueryRowContext(ctx, `
select coalesce(max(level), 0)
from sys_building
where cid = ? and bid = ?`, cid, governmentCapacityBuildingID).Scan(&level)
	return level, err
}

func (r *Repository) cityBuildingSlotOccupiedTx(ctx context.Context, tx *sql.Tx, cid int, position int) (bool, error) {
	var count int
	if err := tx.QueryRowContext(ctx, `
select count(*)
from sys_building
where cid = ? and xy = ?`, cid, position).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) ensureCityBuildingSlotEmptyTx(ctx context.Context, tx *sql.Tx, cid int, position int) error {
	occupied, err := r.cityBuildingSlotOccupiedTx(ctx, tx, cid, position)
	if err != nil {
		return err
	}
	if occupied {
		return newInvalidError("building slot is already occupied")
	}
	return nil
}

func (r *Repository) ensureCityUpgradeResourcesAvailableTx(ctx context.Context, tx *sql.Tx, cid int, config cityBuildingUpgradeConfig) error {
	var snapshot cityBuildingUpgradeConfig
	if err := tx.QueryRowContext(ctx, `
select wood, rock, iron, food, gold, people
from mem_city_resource
where cid = ?`, cid).Scan(
		&snapshot.Wood,
		&snapshot.Rock,
		&snapshot.Iron,
		&snapshot.Food,
		&snapshot.Gold,
		&snapshot.UpgradePeople,
	); err != nil {
		return err
	}

	if snapshot.Wood < config.Wood ||
		snapshot.Rock < config.Rock ||
		snapshot.Iron < config.Iron ||
		snapshot.Food < config.Food ||
		snapshot.Gold < config.Gold {
		return newInvalidError("city resources are not enough")
	}

	return nil
}
