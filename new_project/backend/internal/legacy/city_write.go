package legacy

import (
	"context"
	"database/sql"
	"math"
)

const (
	gameSpeedRate  = 1
	foodBuildingID = 1
	woodBuildingID = 2
	rockBuildingID = 3
	ironBuildingID = 4
	globalFoodRate = 1000
	globalWoodRate = 1000
	globalRockRate = 500
	globalIronRate = 400
)

func (r *Repository) UpdateCityTax(ctx context.Context, uid int, cid int, tax int) (CityDetail, error) {
	if tax < 0 {
		tax = 0
	}
	if tax > 100 {
		tax = 100
	}

	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return CityDetail{}, err
	} else if !allowed {
		return CityDetail{}, ErrForbidden
	}

	if r.db != nil {
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return CityDetail{}, err
		}
		defer tx.Rollback()

		if _, err := tx.ExecContext(ctx, "update mem_city_resource set tax = ? where cid = ?", tax, cid); err != nil {
			return CityDetail{}, err
		}
		if _, err := tx.ExecContext(ctx, "update mem_city_resource set morale_stable = greatest(0, least(100 - tax - complaint, 100)) where cid = ?", cid); err != nil {
			return CityDetail{}, err
		}
		if err := tx.Commit(); err != nil {
			return CityDetail{}, err
		}
	}

	return r.CityDetail(ctx, cid)
}

func (r *Repository) UpdateCityProduction(ctx context.Context, uid int, cid int, settings ProductionSettings) (CityDetail, error) {
	settings.FoodRate = clamp(settings.FoodRate, 0, 100)
	settings.WoodRate = clamp(settings.WoodRate, 0, 100)
	settings.RockRate = clamp(settings.RockRate, 0, 100)
	settings.IronRate = clamp(settings.IronRate, 0, 100)

	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return CityDetail{}, err
	} else if !allowed {
		return CityDetail{}, ErrForbidden
	}

	if r.db != nil {
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return CityDetail{}, err
		}
		defer tx.Rollback()

		if err := r.ensureCityResAdd(ctx, tx, cid); err != nil {
			return CityDetail{}, err
		}
		if _, err := tx.ExecContext(ctx, `
update sys_city_res_add
set food_rate = ?, wood_rate = ?, rock_rate = ?, iron_rate = ?, resource_changing = 1
where cid = ?`,
			settings.FoodRate,
			settings.WoodRate,
			settings.RockRate,
			settings.IronRate,
			cid,
		); err != nil {
			return CityDetail{}, err
		}

		if err := r.recalculateCityProduction(ctx, tx, cid); err != nil {
			return CityDetail{}, err
		}
		if err := tx.Commit(); err != nil {
			return CityDetail{}, err
		}
	}

	return r.CityDetail(ctx, cid)
}

func (r *Repository) loadCityProduction(ctx context.Context, cid int) (ProductionState, error) {
	if r.db == nil {
		return ProductionState{
			Settings:      ProductionSettings{FoodRate: 100, WoodRate: 100, RockRate: 100, IronRate: 100},
			FoodAdd:       1000,
			FoodArmyUse:   0,
			WoodAdd:       1000,
			RockAdd:       500,
			IronAdd:       400,
			HeroFee:       0,
			GoldAdd:       50000,
			PeopleWorking: 100,
		}, nil
	}

	if _, err := r.db.ExecContext(ctx, "insert into sys_city_res_add (cid) values (?) on duplicate key update cid = cid", cid); err != nil {
		return ProductionState{}, err
	}

	state := ProductionState{}
	err := r.db.QueryRowContext(ctx, `
select
	a.food_rate,
	a.wood_rate,
	a.rock_rate,
	a.iron_rate,
	r.food_add,
	r.food_army_use,
	r.wood_add,
	r.rock_add,
	r.iron_add,
	r.hero_fee,
	floor(r.tax * r.people * 0.01 * ? * r.gold_rate * 0.01 - r.hero_fee),
	r.people_working
from sys_city_res_add a
join mem_city_resource r on r.cid = a.cid
where a.cid = ?`, gameSpeedRate, cid).Scan(
		&state.Settings.FoodRate,
		&state.Settings.WoodRate,
		&state.Settings.RockRate,
		&state.Settings.IronRate,
		&state.FoodAdd,
		&state.FoodArmyUse,
		&state.WoodAdd,
		&state.RockAdd,
		&state.IronAdd,
		&state.HeroFee,
		&state.GoldAdd,
		&state.PeopleWorking,
	)
	return state, err
}

func (r *Repository) ensureCityResAdd(ctx context.Context, tx *sql.Tx, cid int) error {
	_, err := tx.ExecContext(ctx, "insert into sys_city_res_add (cid) values (?) on duplicate key update cid = cid", cid)
	return err
}

func (r *Repository) recalculateCityProduction(ctx context.Context, tx *sql.Tx, cid int) error {
	var settings ProductionSettings
	var goodsFoodAdd, goodsWoodAdd, goodsRockAdd, goodsIronAdd int64
	if err := tx.QueryRowContext(ctx, `
select
	food_rate,
	wood_rate,
	rock_rate,
	iron_rate,
	coalesce(goods_food_add, 0),
	coalesce(goods_wood_add, 0),
	coalesce(goods_rock_add, 0),
	coalesce(goods_iron_add, 0)
from sys_city_res_add
where cid = ?`, cid).Scan(
		&settings.FoodRate,
		&settings.WoodRate,
		&settings.RockRate,
		&settings.IronRate,
		&goodsFoodAdd,
		&goodsWoodAdd,
		&goodsRockAdd,
		&goodsIronAdd,
	); err != nil {
		return err
	}

	var people int64
	if err := tx.QueryRowContext(ctx, "select people from mem_city_resource where cid = ?", cid).Scan(&people); err != nil {
		return err
	}

	var foodWorkers, woodWorkers, rockWorkers, ironWorkers sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
select
	coalesce(sum(case when b.bid = ? then l.using_people else 0 end), 0),
	coalesce(sum(case when b.bid = ? then l.using_people else 0 end), 0),
	coalesce(sum(case when b.bid = ? then l.using_people else 0 end), 0),
	coalesce(sum(case when b.bid = ? then l.using_people else 0 end), 0)
from sys_building b
join cfg_building_level l on l.bid = b.bid and l.level = b.level
where b.cid = ?`, foodBuildingID, woodBuildingID, rockBuildingID, ironBuildingID, cid).Scan(
		&foodWorkers,
		&woodWorkers,
		&rockWorkers,
		&ironWorkers,
	); err != nil {
		return err
	}

	foodNeed := float64(foodWorkers.Int64) * float64(settings.FoodRate) / 100.0
	woodNeed := float64(woodWorkers.Int64) * float64(settings.WoodRate) / 100.0
	rockNeed := float64(rockWorkers.Int64) * float64(settings.RockRate) / 100.0
	ironNeed := float64(ironWorkers.Int64) * float64(settings.IronRate) / 100.0
	peopleWorking := foodNeed + woodNeed + rockNeed + ironNeed

	multiplier := 0.0
	if peopleWorking > 0 && people > 0 {
		multiplier = math.Min(1, float64(people)/peopleWorking)
	}

	foodAdd := float64(globalFoodRate*gameSpeedRate) * float64(foodWorkers.Int64) * float64(settings.FoodRate) / 100.0 * multiplier
	woodAdd := float64(globalWoodRate*gameSpeedRate) * float64(woodWorkers.Int64) * float64(settings.WoodRate) / 100.0 * multiplier
	rockAdd := float64(globalRockRate*gameSpeedRate) * float64(rockWorkers.Int64) * float64(settings.RockRate) / 100.0 * multiplier
	ironAdd := float64(globalIronRate*gameSpeedRate) * float64(ironWorkers.Int64) * float64(settings.IronRate) / 100.0 * multiplier
	foodAdd *= 1 + float64(goodsFoodAdd)/100.0
	woodAdd *= 1 + float64(goodsWoodAdd)/100.0
	rockAdd *= 1 + float64(goodsRockAdd)/100.0
	ironAdd *= 1 + float64(goodsIronAdd)/100.0

	_, err := tx.ExecContext(ctx, `
update mem_city_resource
set people_working = ?, food_add = ?, wood_add = ?, rock_add = ?, iron_add = ?
where cid = ?`,
		int64(math.Round(peopleWorking)),
		int64(math.Round(foodAdd)),
		int64(math.Round(woodAdd)),
		int64(math.Round(rockAdd)),
		int64(math.Round(ironAdd)),
		cid,
	)
	return err
}
