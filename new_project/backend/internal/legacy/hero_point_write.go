package legacy

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
)

type cityHeroPointRecord struct {
	State       int
	Level       int
	AffairsBase int
	AffairsAdd  int
	BraveryBase int
	BraveryAdd  int
	WisdomBase  int
	WisdomAdd   int
}

type heroAttributeBonuses struct {
	Command   int
	Affairs   int
	Bravery   int
	Wisdom    int
	ForceMax  int
	EnergyMax int
	Speed     int
	Attack    int
	Defence   int
}

func (r *Repository) AddHeroPoint(ctx context.Context, uid int, cid int, hid int, stat string, amount int) (HeroRoster, error) {
	if hid <= 0 {
		return HeroRoster{}, newInvalidError("无效的武将编号。")
	}

	stat = normalizeHeroPointStat(stat)
	if stat == "" {
		return HeroRoster{}, newInvalidError("无效的加点属性。")
	}
	if amount <= 0 {
		return HeroRoster{}, newInvalidError("加点数量无效。")
	}

	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return HeroRoster{}, err
	} else if !allowed {
		return HeroRoster{}, ErrForbidden
	}

	if r.db == nil {
		return r.fixtureAddHeroPoint(cid, hid, stat, amount)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return HeroRoster{}, err
	}
	defer tx.Rollback()

	hero, err := r.cityHeroPointRecord(ctx, tx, uid, cid, hid)
	if err != nil {
		if err == sql.ErrNoRows {
			return HeroRoster{}, newInvalidError("未找到该武将。")
		}
		return HeroRoster{}, err
	}
	if !heroAllowsPointAllocation(hero.State) {
		return HeroRoster{}, newInvalidError("武将不在城内，无法加点。")
	}
	if heroAvailablePoints(hero) < amount {
		return HeroRoster{}, newInvalidError("没有剩余潜力点。")
	}

	switch stat {
	case "affairs":
		hero.AffairsAdd += amount
		_, err = tx.ExecContext(ctx, "update sys_city_hero set affairs_add = ? where hid = ?", hero.AffairsAdd, hid)
	case "bravery":
		hero.BraveryAdd += amount
		_, err = tx.ExecContext(ctx, "update sys_city_hero set bravery_add = ? where hid = ?", hero.BraveryAdd, hid)
	case "wisdom":
		hero.WisdomAdd += amount
		_, err = tx.ExecContext(ctx, "update sys_city_hero set wisdom_add = ? where hid = ?", hero.WisdomAdd, hid)
	}
	if err != nil {
		return HeroRoster{}, err
	}

	bonuses, err := r.loadHeroAttributeBonuses(ctx, tx, uid, hid)
	if err != nil {
		return HeroRoster{}, err
	}

	braveryBuffed, wisdomBuffed, err := r.loadHeroPointBuffs(ctx, tx, hid)
	if err != nil {
		return HeroRoster{}, err
	}

	bravery := hero.BraveryBase + hero.BraveryAdd
	if braveryBuffed {
		bravery = bravery * 3 / 2
	}

	wisdom := hero.WisdomBase + hero.WisdomAdd
	if wisdomBuffed {
		wisdom = wisdom * 5 / 4
	}

	forceMax := 100 + hero.Level/5 + (bravery+bonuses.Bravery)/3 + bonuses.ForceMax
	energyMax := 100 + hero.Level/5 + (wisdom+bonuses.Wisdom)/3 + bonuses.EnergyMax

	if _, err := tx.ExecContext(ctx, `
update sys_city_hero
set command_add_on = ?, affairs_add_on = ?, bravery_add_on = ?, wisdom_add_on = ?,
	force_max_add_on = ?, energy_max_add_on = ?, speed_add_on = ?, attack_add_on = ?, defence_add_on = ?
where hid = ?`,
		bonuses.Command,
		bonuses.Affairs,
		bonuses.Bravery,
		bonuses.Wisdom,
		bonuses.ForceMax,
		bonuses.EnergyMax,
		bonuses.Speed,
		bonuses.Attack,
		bonuses.Defence,
		hid,
	); err != nil {
		return HeroRoster{}, err
	}

	if err := r.upsertHeroBlood(ctx, tx, hid, forceMax, energyMax); err != nil {
		return HeroRoster{}, err
	}
	if err := r.recalculateHeroFee(ctx, tx, uid, cid); err != nil {
		return HeroRoster{}, err
	}

	chiefHID, err := r.cityOfficeHolder(ctx, tx, uid, cid, "chiefhid")
	if err != nil {
		return HeroRoster{}, err
	}
	if chiefHID == hid {
		if err := r.ensureCityResAdd(ctx, tx, cid); err != nil {
			return HeroRoster{}, err
		}
		if _, err := tx.ExecContext(ctx, "update sys_city_res_add set resource_changing = 1 where cid = ?", cid); err != nil {
			return HeroRoster{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return HeroRoster{}, err
	}

	return r.CityHeroes(ctx, uid, cid, 24)
}

func (r *Repository) cityHeroPointRecord(ctx context.Context, tx *sql.Tx, uid int, cid int, hid int) (cityHeroPointRecord, error) {
	record := cityHeroPointRecord{}
	err := tx.QueryRowContext(ctx, `
select state, level, affairs_base, affairs_add, bravery_base, bravery_add, wisdom_base, wisdom_add
from sys_city_hero
where uid = ? and cid = ? and hid = ?`,
		uid,
		cid,
		hid,
	).Scan(
		&record.State,
		&record.Level,
		&record.AffairsBase,
		&record.AffairsAdd,
		&record.BraveryBase,
		&record.BraveryAdd,
		&record.WisdomBase,
		&record.WisdomAdd,
	)
	return record, err
}

func (r *Repository) loadHeroAttributeBonuses(ctx context.Context, tx *sql.Tx, uid int, hid int) (heroAttributeBonuses, error) {
	rows, err := tx.QueryContext(ctx, `
select coalesce(c.attribute, '')
from sys_user_armor u
left join sys_hero_armor h on h.hid = u.hid and h.sid = u.sid
left join cfg_armor c on c.id = h.armorid
where u.uid = ? and u.hid = ? and u.hp > 0`,
		uid,
		hid,
	)
	if err != nil {
		return heroAttributeBonuses{}, err
	}
	defer rows.Close()

	bonuses := heroAttributeBonuses{}
	for rows.Next() {
		var attribute string
		if err := rows.Scan(&attribute); err != nil {
			return heroAttributeBonuses{}, err
		}
		applyArmorAttributeBonuses(attribute, &bonuses)
	}

	return bonuses, rows.Err()
}

func (r *Repository) loadHeroPointBuffs(ctx context.Context, tx *sql.Tx, hid int) (bool, bool, error) {
	rows, err := tx.QueryContext(ctx, `
select buftype
from mem_hero_buffer
where hid = ? and buftype in (3, 4) and endtime > unix_timestamp()`,
		hid,
	)
	if err != nil {
		return false, false, err
	}
	defer rows.Close()

	braveryBuffed := false
	wisdomBuffed := false
	for rows.Next() {
		var buffType int
		if err := rows.Scan(&buffType); err != nil {
			return false, false, err
		}
		if buffType == 3 {
			braveryBuffed = true
		}
		if buffType == 4 {
			wisdomBuffed = true
		}
	}

	return braveryBuffed, wisdomBuffed, rows.Err()
}

func (r *Repository) upsertHeroBlood(ctx context.Context, tx *sql.Tx, hid int, forceMax int, energyMax int) error {
	var (
		currentForce  int
		currentEnergy int
	)

	err := tx.QueryRowContext(ctx, "select `force`, `energy` from mem_hero_blood where hid = ?", hid).Scan(&currentForce, &currentEnergy)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}

		if forceMax > 100 {
			currentForce = 100
		} else {
			currentForce = forceMax
		}
		if energyMax > 100 {
			currentEnergy = 100
		} else {
			currentEnergy = energyMax
		}

		_, err = tx.ExecContext(ctx, "insert into mem_hero_blood (hid, `force`, force_max, `energy`, energy_max) values (?, ?, ?, ?, ?)",
			hid,
			currentForce,
			forceMax,
			currentEnergy,
			energyMax,
		)
		return err
	}

	currentForce = clamp(currentForce, 0, forceMax)
	currentEnergy = clamp(currentEnergy, 0, energyMax)
	_, err = tx.ExecContext(ctx, "update mem_hero_blood set `force` = ?, force_max = ?, `energy` = ?, energy_max = ? where hid = ?",
		currentForce,
		forceMax,
		currentEnergy,
		energyMax,
		hid,
	)
	return err
}

func (r *Repository) recalculateHeroFee(ctx context.Context, tx *sql.Tx, uid int, cid int) error {
	var heroFee int64
	if err := tx.QueryRowContext(ctx, `
select coalesce(sum(level * 20 + (greatest(affairs_base + affairs_add - 90, 0) + greatest(bravery_base + bravery_add - 90, 0) + greatest(wisdom_base + wisdom_add - 90, 0)) * 50), 0)
from sys_city_hero
where cid = ? and uid = ? and state < 5`,
		cid,
		uid,
	).Scan(&heroFee); err != nil {
		return err
	}

	_, err := tx.ExecContext(ctx, "update mem_city_resource set hero_fee = ? where cid = ?", heroFee, cid)
	return err
}

func normalizeHeroPointStat(stat string) string {
	switch strings.ToLower(strings.TrimSpace(stat)) {
	case "affairs":
		return "affairs"
	case "bravery":
		return "bravery"
	case "wisdom":
		return "wisdom"
	default:
		return ""
	}
}

func heroAllowsPointAllocation(state int) bool {
	return state == 0 || state == 1 || state == 7 || state == 8
}

func heroAvailablePoints(hero cityHeroPointRecord) int {
	return clamp(hero.Level-hero.AffairsAdd-hero.BraveryAdd-hero.WisdomAdd, 0, hero.Level)
}

func applyArmorAttributeBonuses(raw string, bonuses *heroAttributeBonuses) {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	if len(parts) == 0 {
		return
	}

	attributeCount, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || attributeCount*2+1 != len(parts) {
		return
	}

	for index := 1; index+1 < len(parts); index += 2 {
		attributeType, err := strconv.Atoi(strings.TrimSpace(parts[index]))
		if err != nil {
			continue
		}

		value, err := strconv.Atoi(strings.TrimSpace(parts[index+1]))
		if err != nil {
			continue
		}

		switch attributeType {
		case 1:
			bonuses.Command += value
		case 2:
			bonuses.Affairs += value
		case 3:
			bonuses.Bravery += value
		case 4:
			bonuses.Wisdom += value
		case 5:
			bonuses.ForceMax += value
		case 6:
			bonuses.EnergyMax += value
		case 8:
			bonuses.Attack += value
		case 9:
			bonuses.Defence += value
		case 11:
			bonuses.Speed += value
		}
	}
}

func (r *Repository) fixtureAddHeroPoint(cid int, hid int, stat string, amount int) (HeroRoster, error) {
	roster := r.fixtureHeroRoster(cid)
	for index := range roster.Items {
		item := &roster.Items[index]
		if item.HID != hid {
			continue
		}
		if !heroAllowsPointAllocation(item.State) {
			return HeroRoster{}, newInvalidError("武将不在城内，无法加点。")
		}
		if item.Available < amount {
			return HeroRoster{}, newInvalidError("没有剩余潜力点。")
		}

		switch stat {
		case "affairs":
			item.AffairsAdd += amount
			item.Affairs += amount
		case "bravery":
			item.BraveryAdd += amount
			item.Bravery += amount
		case "wisdom":
			item.WisdomAdd += amount
			item.Wisdom += amount
		}

		item.Available -= amount
		item.ForceMax = 100 + item.Level/5 + item.Bravery/3
		item.EnergyMax = 100 + item.Level/5 + item.Wisdom/3
		item.Force = clamp(item.Force, 0, item.ForceMax)
		item.Energy = clamp(item.Energy, 0, item.EnergyMax)
		return roster, nil
	}

	return HeroRoster{}, newInvalidError("未找到该武将。")
}
