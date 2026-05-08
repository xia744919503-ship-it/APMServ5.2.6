package legacy

import (
	"context"
	"database/sql"
	"fmt"
)

const officeBuildingID = 11

type cityOffice struct {
	cityColumn              string
	heroState               int
	clearChiefLoyaltyOnZero bool
	markResourceChanging    bool
}

var (
	cityChiefOffice = cityOffice{
		cityColumn:              "chiefhid",
		heroState:               1,
		clearChiefLoyaltyOnZero: true,
		markResourceChanging:    true,
	}
	cityGeneralOffice = cityOffice{
		cityColumn: "generalid",
		heroState:  7,
	}
	cityCounsellorOffice = cityOffice{
		cityColumn: "counsellorid",
		heroState:  8,
	}
)

func (r *Repository) UpdateCityChief(ctx context.Context, uid int, cid int, hid int) (HeroRoster, error) {
	return r.updateCityOffice(ctx, uid, cid, hid, cityChiefOffice)
}

func (r *Repository) UpdateCityGeneral(ctx context.Context, uid int, cid int, hid int) (HeroRoster, error) {
	return r.updateCityOffice(ctx, uid, cid, hid, cityGeneralOffice)
}

func (r *Repository) UpdateCityCounsellor(ctx context.Context, uid int, cid int, hid int) (HeroRoster, error) {
	return r.updateCityOffice(ctx, uid, cid, hid, cityCounsellorOffice)
}

func (r *Repository) updateCityOffice(ctx context.Context, uid int, cid int, hid int, office cityOffice) (HeroRoster, error) {
	if hid < 0 {
		return HeroRoster{}, newInvalidError("invalid hero id")
	}

	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return HeroRoster{}, err
	} else if !allowed {
		return HeroRoster{}, ErrForbidden
	}

	if r.db == nil {
		return r.fixtureUpdatedOfficeRoster(cid, hid, office.heroState), nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return HeroRoster{}, err
	}
	defer tx.Rollback()

	oldHID, err := r.cityOfficeHolder(ctx, tx, uid, cid, office.cityColumn)
	if err != nil {
		return HeroRoster{}, err
	}

	if hid == oldHID {
		if err := tx.Commit(); err != nil {
			return HeroRoster{}, err
		}
		return r.CityHeroes(ctx, uid, cid, 24)
	}

	if err := r.ensureOfficeBuilding(ctx, tx, cid); err != nil {
		return HeroRoster{}, err
	}

	if oldHID > 0 {
		busy, err := r.heroIsInTroop(ctx, tx, oldHID)
		if err != nil {
			return HeroRoster{}, err
		}
		if busy {
			return HeroRoster{}, newInvalidError("将领出征中，不能任命。")
		}

		if _, err := tx.ExecContext(ctx, "update sys_city_hero set state = 0 where hid = ?", oldHID); err != nil {
			return HeroRoster{}, err
		}
	}

	if hid > 0 {
		if err := r.ensureAvailableCityHero(ctx, tx, uid, cid, hid); err != nil {
			return HeroRoster{}, err
		}

		loyalty, err := r.cityHeroLoyalty(ctx, tx, hid)
		if err != nil {
			return HeroRoster{}, err
		}

		if _, err := tx.ExecContext(ctx, "update sys_city_hero set state = ? where hid = ?", office.heroState, hid); err != nil {
			return HeroRoster{}, err
		}
		if _, err := tx.ExecContext(ctx, "update mem_city_resource set chief_loyalty = ? where cid = ?", loyalty, cid); err != nil {
			return HeroRoster{}, err
		}
	} else if office.clearChiefLoyaltyOnZero {
		if _, err := tx.ExecContext(ctx, "update mem_city_resource set chief_loyalty = 0 where cid = ?", cid); err != nil {
			return HeroRoster{}, err
		}
	}

	if office.markResourceChanging {
		if err := r.ensureCityResAdd(ctx, tx, cid); err != nil {
			return HeroRoster{}, err
		}
		if _, err := tx.ExecContext(ctx, "update sys_city_res_add set resource_changing = 1 where cid = ?", cid); err != nil {
			return HeroRoster{}, err
		}
	}

	query := fmt.Sprintf("update sys_city set %s = ? where cid = ?", office.cityColumn)
	if _, err := tx.ExecContext(ctx, query, hid, cid); err != nil {
		return HeroRoster{}, err
	}

	if err := tx.Commit(); err != nil {
		return HeroRoster{}, err
	}

	return r.CityHeroes(ctx, uid, cid, 24)
}

func (r *Repository) cityOfficeHolder(ctx context.Context, tx *sql.Tx, uid int, cid int, column string) (int, error) {
	query := fmt.Sprintf("select %s from sys_city where cid = ? and uid = ?", column)

	var hid int
	if err := tx.QueryRowContext(ctx, query, cid, uid).Scan(&hid); err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrForbidden
		}
		return 0, err
	}

	return hid, nil
}

func (r *Repository) ensureAvailableCityHero(ctx context.Context, tx *sql.Tx, uid int, cid int, hid int) error {
	var count int
	if err := tx.QueryRowContext(ctx, `
select count(*)
from sys_city_hero
where cid = ? and uid = ? and hid = ? and state = 0`, cid, uid, hid).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return newInvalidError("任命将领失败。")
	}

	return nil
}

func (r *Repository) cityHeroLoyalty(ctx context.Context, tx *sql.Tx, hid int) (int, error) {
	var loyalty int
	if err := tx.QueryRowContext(ctx, "select loyalty from sys_city_hero where hid = ?", hid).Scan(&loyalty); err != nil {
		return 0, err
	}

	return loyalty, nil
}

func (r *Repository) ensureOfficeBuilding(ctx context.Context, tx *sql.Tx, cid int) error {
	var count int
	if err := tx.QueryRowContext(ctx, "select count(*) from sys_building where cid = ? and bid = ?", cid, officeBuildingID).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return newInvalidError("需要官府后才能任命官职。")
	}

	return nil
}

func (r *Repository) heroIsInTroop(ctx context.Context, tx *sql.Tx, hid int) (bool, error) {
	if hid <= 0 {
		return false, nil
	}

	var count int
	if err := tx.QueryRowContext(ctx, "select count(*) from sys_troops where hid = ? and uid != 0", hid).Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *Repository) fixtureUpdatedOfficeRoster(cid int, hid int, officeState int) HeroRoster {
	roster := r.fixtureHeroRoster(cid)
	for index := range roster.Items {
		item := &roster.Items[index]
		if hid > 0 && item.HID == hid {
			item.State = officeState
			item.StateName = heroStateDisplayLabel(officeState)
			item.StateLabel = item.StateName
			continue
		}

		if item.State == officeState {
			item.State = 0
			item.StateName = heroStateDisplayLabel(0)
			item.StateLabel = item.StateName
		}
	}

	return roster
}
