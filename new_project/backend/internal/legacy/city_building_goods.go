package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"math"
)

var buildingSpeedGoodsReduceSeconds = map[int]int64{
	67: 15 * 60,
	68: 60 * 60,
	69: 150 * 60,
	70: 8 * 60 * 60,
	71: 10 * 60 * 60,
	72: 0,
	73: 0,
}

func (r *Repository) BuildingSpeedGoods(ctx context.Context, uid int, cid int, position int) (BuildingSpeedGoodsSnapshot, error) {
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return BuildingSpeedGoodsSnapshot{}, err
	} else if !allowed {
		return BuildingSpeedGoodsSnapshot{}, ErrForbidden
	}
	if r.db == nil {
		return BuildingSpeedGoodsSnapshot{}, ErrDatabaseUnavailable
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return BuildingSpeedGoodsSnapshot{}, err
	}
	defer tx.Rollback()

	if err := r.settleCityBuildingQueueTx(ctx, tx, cid); err != nil {
		return BuildingSpeedGoodsSnapshot{}, fmt.Errorf("settle city building queue: %w", err)
	}

	building, err := r.loadCityBuildingInfoBuildingTx(ctx, tx, cid, position)
	if err != nil {
		return BuildingSpeedGoodsSnapshot{}, err
	}
	if building.State == 0 {
		return BuildingSpeedGoodsSnapshot{}, newInvalidError("building is not busy")
	}

	now, err := r.currentUnixTimeTx(ctx, tx)
	if err != nil {
		return BuildingSpeedGoodsSnapshot{}, err
	}
	remaining := building.StateEndTime - now
	if remaining < 0 {
		remaining = 0
	}

	goods, err := r.loadBuildingSpeedGoodsTx(ctx, tx, uid, remaining)
	if err != nil {
		return BuildingSpeedGoodsSnapshot{}, err
	}

	return BuildingSpeedGoodsSnapshot{
		Type:      0,
		Time:      remaining,
		Cost:      buildingInstantCost(remaining),
		Position:  position,
		Building:  building,
		GoodsList: goods,
	}, tx.Commit()
}

func (r *Repository) UseBuildingSpeedGoods(ctx context.Context, uid int, cid int, position int, gid int) (CityDetail, error) {
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

	building, err := r.loadCityBuildingInfoBuildingTx(ctx, tx, cid, position)
	if err != nil {
		return CityDetail{}, err
	}
	if building.State == 0 {
		return CityDetail{}, newInvalidError("building is not busy")
	}

	now, err := r.currentUnixTimeTx(ctx, tx)
	if err != nil {
		return CityDetail{}, err
	}
	remaining := building.StateEndTime - now
	if remaining <= 0 {
		if err := r.settleCityBuildingQueueTx(ctx, tx, cid); err != nil {
			return CityDetail{}, err
		}
		if err := tx.Commit(); err != nil {
			return CityDetail{}, err
		}
		return r.CityDetail(ctx, cid)
	}

	reduce, ok := buildingSpeedGoodsReduceSeconds[gid]
	if !ok {
		return CityDetail{}, newInvalidError("goods cannot speed up buildings")
	}
	if err := r.adjustUserGoodsTx(ctx, tx, uid, gid, -1); err != nil {
		return CityDetail{}, err
	}

	if reduce <= 0 || reduce >= remaining {
		if err := r.finishBuildingQueueRowTx(ctx, tx, cid, building); err != nil {
			return CityDetail{}, err
		}
	} else {
		newEnd := building.StateEndTime - reduce
		if err := r.shiftBuildingQueueEndTx(ctx, tx, cid, building, newEnd); err != nil {
			return CityDetail{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return CityDetail{}, err
	}
	return r.CityDetail(ctx, cid)
}

func (r *Repository) loadBuildingSpeedGoodsTx(ctx context.Context, tx *sql.Tx, uid int, remaining int64) ([]SpeedGoods, error) {
	rows, err := tx.QueryContext(ctx, `
select c.gid, c.name, c.description, c.image, coalesce(g.count, 0)
from cfg_goods c
left join sys_goods g on g.uid = ? and g.gid = c.gid
where c.gid in (67,68,69,70,71,72,73)
order by c.gid`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]SpeedGoods, 0, 7)
	for rows.Next() {
		item := SpeedGoods{}
		if err := rows.Scan(&item.GID, &item.Name, &item.Description, &item.Image, &item.Count); err != nil {
			return nil, err
		}
		item.ReduceTime = buildingSpeedGoodsReduceSeconds[item.GID]
		if item.GID == 73 {
			item.Cost = buildingInstantCost(remaining)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) finishBuildingQueueRowTx(ctx context.Context, tx *sql.Tx, cid int, building Building) error {
	table := "mem_building_upgrading"
	if building.State == 2 {
		table = "mem_building_destroying"
	}
	rows, err := r.loadCityBuildingQueueRowsTx(ctx, tx, table, cid, math.MaxInt64)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.Position == building.Position && row.BID == building.BID {
			if building.State == 1 {
				if _, err := tx.ExecContext(ctx, `
update sys_building
set level = ?, state = 0, state_starttime = 0, state_endtime = 0
where id = ? and cid = ?`, row.Level, row.ID, cid); err != nil {
					return err
				}
				if _, err := tx.ExecContext(ctx, "delete from mem_building_upgrading where id = ?", row.ID); err != nil {
					return err
				}
			} else if row.Level <= 0 {
				if _, err := tx.ExecContext(ctx, "delete from sys_building where id = ? and cid = ?", row.ID, cid); err != nil {
					return err
				}
				if _, err := tx.ExecContext(ctx, "delete from mem_building_destroying where id = ?", row.ID); err != nil {
					return err
				}
			} else {
				if _, err := tx.ExecContext(ctx, `
update sys_building
set level = ?, state = 0, state_starttime = 0, state_endtime = 0
where id = ? and cid = ?`, row.Level, row.ID, cid); err != nil {
					return err
				}
				if _, err := tx.ExecContext(ctx, "delete from mem_building_destroying where id = ?", row.ID); err != nil {
					return err
				}
			}
			return r.refreshCityAfterBuildingQueueChangeTx(ctx, tx, cid)
		}
	}
	return newInvalidError("building queue row is missing")
}

func (r *Repository) shiftBuildingQueueEndTx(ctx context.Context, tx *sql.Tx, cid int, building Building, newEnd int64) error {
	table := "mem_building_upgrading"
	if building.State == 2 {
		table = "mem_building_destroying"
	}
	if _, err := tx.ExecContext(ctx, `
update sys_building
set state_endtime = ?
where cid = ? and xy = ? and state <> 0`, newEnd, cid, building.Position); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, fmt.Sprintf(`
update %s
set state_endtime = ?
where cid = ? and xy = ?`, table), newEnd, cid, building.Position)
	return err
}

func (r *Repository) refreshCityAfterBuildingQueueChangeTx(ctx context.Context, tx *sql.Tx, cid int) error {
	if err := r.ensureCityResAdd(ctx, tx, cid); err != nil {
		return err
	}
	if err := r.recalculateCityProduction(ctx, tx, cid); err != nil {
		return err
	}
	if err := r.recalculateCityPeopleMaxTx(ctx, tx, cid); err != nil {
		return err
	}
	return r.recalculateCityGoldMaxTx(ctx, tx, cid)
}

func buildingInstantCost(remaining int64) int64 {
	if remaining <= 0 {
		return 0
	}
	return int64(math.Ceil(float64(remaining) / 3600.0))
}
