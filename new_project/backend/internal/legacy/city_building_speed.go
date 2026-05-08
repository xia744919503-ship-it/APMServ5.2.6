package legacy

import (
	"context"
	"database/sql"
	"math"
)

const (
	buildingSpeedTechnicID     = 17
	buildingSpeedHeroBuffType  = 2
	buildingHeroBuffMultiplier = 1.25
	buildingGameSpeedRate      = 1.0
)

func (r *Repository) buildingSpeedRateTx(ctx context.Context, tx *sql.Tx, cid int) (float64, error) {
	speedAdd := 0.0

	var technicLevel int64
	err := tx.QueryRowContext(ctx, `
select level
from sys_city_technic
where cid = ? and tid = ?`, cid, buildingSpeedTechnicID).Scan(&technicLevel)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if err == nil && technicLevel > 0 {
		speedAdd += float64(technicLevel * 10)
	}

	var chiefHID int
	if err := tx.QueryRowContext(ctx, `
select chiefhid
from sys_city
where cid = ?`, cid).Scan(&chiefHID); err != nil {
		return 0, err
	}

	if chiefHID > 0 {
		heroAdd, err := r.buildingSpeedHeroAddTx(ctx, tx, chiefHID)
		if err != nil {
			return 0, err
		}
		speedAdd += heroAdd
	}

	return 1.0 / (1.0 + 0.01*speedAdd), nil
}

func (r *Repository) buildingSpeedHeroAddTx(ctx context.Context, tx *sql.Tx, hid int) (float64, error) {
	var (
		affairsBase  float64
		affairsAdd   float64
		affairsAddOn float64
	)

	err := tx.QueryRowContext(ctx, `
select affairs_base, affairs_add, affairs_add_on
from sys_city_hero
where hid = ?`, hid).Scan(&affairsBase, &affairsAdd, &affairsAddOn)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	hasBuff, err := r.heroHasBuildingSpeedBuffTx(ctx, tx, hid)
	if err != nil {
		return 0, err
	}

	multiplier := 1.0
	if hasBuff {
		multiplier = buildingHeroBuffMultiplier
	}

	return (affairsBase+affairsAdd)*multiplier + affairsAddOn, nil
}

func (r *Repository) heroHasBuildingSpeedBuffTx(ctx context.Context, tx *sql.Tx, hid int) (bool, error) {
	var count int
	if err := tx.QueryRowContext(ctx, `
select count(*)
from mem_hero_buffer
where hid = ? and buftype = ? and endtime > unix_timestamp()`,
		hid,
		buildingSpeedHeroBuffType,
	).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func buildingQueueDurationSeconds(baseUpgradeTime int64, speedRate float64) int64 {
	if baseUpgradeTime <= 0 {
		return 0
	}
	return int64(math.Floor(float64(baseUpgradeTime) * speedRate / buildingGameSpeedRate))
}

func buildingDestroyDurationSeconds(baseUpgradeTime int64, speedRate float64) int64 {
	if baseUpgradeTime <= 0 {
		return 0
	}
	return int64(math.Floor(float64(baseUpgradeTime) * 0.5 * speedRate / buildingGameSpeedRate))
}

func buildingDisplayDurationSeconds(baseUpgradeTime int64, speedRate float64) int64 {
	if baseUpgradeTime <= 0 {
		return 0
	}
	return int64(math.Ceil(float64(baseUpgradeTime) * speedRate / buildingGameSpeedRate))
}
