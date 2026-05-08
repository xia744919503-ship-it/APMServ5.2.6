package legacy

import (
	"context"
	"database/sql"
	"math"
	"math/rand"
	"strings"
	"time"
)

const (
	heroBuildingHotel         = 10
	heroBuildingOffice        = 11
	heroRecruitRefreshSeconds = 10800
	heroRecruitTaskGoal       = 84
	heroRecruitDefaultLoyalty = 70
)

func (r *Repository) cityHotelRecruitSnapshot(ctx context.Context, uid int, cid int) (int, []HeroRecruit, error) {
	if r.db == nil {
		recruits := fixtureHeroRecruits(cid)
		return 3, recruits, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, nil, err
	}
	defer tx.Rollback()

	hotelLevel, err := r.cityBuildingLevelTx(ctx, tx, cid, heroBuildingHotel)
	if err != nil {
		return 0, nil, err
	}
	if hotelLevel <= 0 {
		if err := tx.Commit(); err != nil {
			return 0, nil, err
		}
		return 0, []HeroRecruit{}, nil
	}

	if err := r.ensureCityRecruitPoolTx(ctx, tx, cid, hotelLevel); err != nil {
		return 0, nil, err
	}
	recruits, err := r.hotelRecruitsTx(ctx, tx, cid)
	if err != nil {
		return 0, nil, err
	}

	if err := tx.Commit(); err != nil {
		return 0, nil, err
	}
	return hotelLevel, recruits, nil
}

func (r *Repository) RecruitCityHero(ctx context.Context, uid int, cid int, recruitID int) (HeroRoster, error) {
	if recruitID <= 0 {
		return HeroRoster{}, newInvalidError("无效的招募编号。")
	}
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return HeroRoster{}, err
	} else if !allowed {
		return HeroRoster{}, ErrForbidden
	}
	if r.db == nil {
		return r.fixtureRecruitedHeroRoster(cid, recruitID)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return HeroRoster{}, err
	}
	defer tx.Rollback()

	hotelLevel, err := r.cityBuildingLevelTx(ctx, tx, cid, heroBuildingHotel)
	if err != nil {
		return HeroRoster{}, err
	}
	if hotelLevel <= 0 {
		return HeroRoster{}, newInvalidError("当前城池尚未建造客栈，不能招募武将。")
	}

	recruit, err := r.recruitHeroRecordTx(ctx, tx, recruitID)
	if err != nil {
		if err == sql.ErrNoRows {
			return HeroRoster{}, newInvalidError("未找到该招募将领。")
		}
		return HeroRoster{}, err
	}
	if recruit.CID != cid {
		return HeroRoster{}, newInvalidError("该将领不在当前城池的招贤榜中。")
	}

	hasPosition, err := r.cityHasHeroPositionTx(ctx, tx, uid, cid)
	if err != nil {
		return HeroRoster{}, err
	}
	if !hasPosition {
		return HeroRoster{}, newInvalidError("当前官府席位已满，无法继续招募武将。")
	}

	result, err := tx.ExecContext(ctx, `
update mem_city_resource
set gold = gold - ?
where cid = ? and gold >= ?`, recruit.GoldNeed, cid, recruit.GoldNeed)
	if err != nil {
		return HeroRoster{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return HeroRoster{}, err
	}
	if affected == 0 {
		return HeroRoster{}, newInvalidError("当前城池黄金不足，无法招募该武将。")
	}

	result, err = tx.ExecContext(ctx, `
insert into sys_city_hero (
	uid,
	name,
	sex,
	face,
	cid,
	state,
	level,
	exp,
	affairs_base,
	bravery_base,
	wisdom_base,
	affairs_add,
	bravery_add,
	wisdom_add,
	loyalty,
	herotype,
	command_base
) values (?, ?, ?, ?, ?, 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		uid,
		recruit.Name,
		recruit.Sex,
		recruit.Face,
		cid,
		recruit.Level,
		recruit.Exp,
		recruit.AffairsBase,
		recruit.BraveryBase,
		recruit.WisdomBase,
		recruit.AffairsAdd,
		recruit.BraveryAdd,
		recruit.WisdomAdd,
		heroRecruitDefaultLoyalty,
		recruit.HeroType,
		recruit.Loyalty,
	)
	if err != nil {
		return HeroRoster{}, err
	}
	heroID64, err := result.LastInsertId()
	if err != nil {
		return HeroRoster{}, err
	}
	heroID := int(heroID64)

	forceMax := 100 + int(math.Floor(float64(recruit.Level)/5.0)) + int(math.Floor(float64(recruit.BraveryBase+recruit.BraveryAdd)/3.0))
	energyMax := 100 + int(math.Floor(float64(recruit.Level)/5.0)) + int(math.Floor(float64(recruit.WisdomBase+recruit.WisdomAdd)/3.0))
	if _, err := tx.ExecContext(ctx, `
insert into mem_hero_blood (hid, `+"`force`"+`, force_max, `+"`energy`"+`, energy_max)
values (?, 100, ?, 100, ?)`, heroID, forceMax, energyMax); err != nil {
		return HeroRoster{}, err
	}

	if _, err := tx.ExecContext(ctx, "delete from sys_recruit_hero where id = ?", recruitID); err != nil {
		return HeroRoster{}, err
	}
	if err := r.recalculateHeroFee(ctx, tx, uid, cid); err != nil {
		return HeroRoster{}, err
	}
	if _, err := tx.ExecContext(ctx, `
replace into sys_user_goal (uid, gid)
values (?, ?)`, uid, heroRecruitTaskGoal); err != nil {
		return HeroRoster{}, err
	}

	if err := tx.Commit(); err != nil {
		return HeroRoster{}, err
	}
	return r.CityHeroes(ctx, uid, cid, 24)
}

type recruitHeroRecord struct {
	ID          int
	Name        string
	Sex         int
	Face        int
	CID         int
	Level       int
	Exp         int64
	AffairsBase int
	BraveryBase int
	WisdomBase  int
	AffairsAdd  int
	BraveryAdd  int
	WisdomAdd   int
	Loyalty     int
	GoldNeed    int64
	HeroType    int
}

func (r *Repository) recruitHeroRecordTx(ctx context.Context, tx *sql.Tx, recruitID int) (recruitHeroRecord, error) {
	record := recruitHeroRecord{}
	err := tx.QueryRowContext(ctx, `
select
	id,
	coalesce(name, ''),
	coalesce(sex, 0),
	coalesce(face, 0),
	coalesce(cid, 0),
	coalesce(level, 0),
	coalesce(exp, 0),
	coalesce(affairs_base, 0),
	coalesce(bravery_base, 0),
	coalesce(wisdom_base, 0),
	coalesce(affairs_add, 0),
	coalesce(bravery_add, 0),
	coalesce(wisdom_add, 0),
	coalesce(loyalty, 0),
	coalesce(gold_need, 0),
	coalesce(herotype, 0)
from sys_recruit_hero
where id = ?`, recruitID).Scan(
		&record.ID,
		&record.Name,
		&record.Sex,
		&record.Face,
		&record.CID,
		&record.Level,
		&record.Exp,
		&record.AffairsBase,
		&record.BraveryBase,
		&record.WisdomBase,
		&record.AffairsAdd,
		&record.BraveryAdd,
		&record.WisdomAdd,
		&record.Loyalty,
		&record.GoldNeed,
		&record.HeroType,
	)
	if err != nil {
		return recruitHeroRecord{}, err
	}
	record.Name = strings.TrimSpace(record.Name)
	return record, nil
}

func (r *Repository) hotelRecruitsTx(ctx context.Context, tx *sql.Tx, cid int) ([]HeroRecruit, error) {
	rows, err := tx.QueryContext(ctx, `
select
	id,
	coalesce(name, ''),
	coalesce(sex, 0),
	coalesce(face, 0),
	coalesce(cid, 0),
	coalesce(level, 0),
	coalesce(affairs_base, 0),
	coalesce(bravery_base, 0),
	coalesce(wisdom_base, 0),
	coalesce(affairs_add, 0),
	coalesce(bravery_add, 0),
	coalesce(wisdom_add, 0),
	coalesce(loyalty, 0),
	coalesce(gold_need, 0)
from sys_recruit_hero
where cid = ?
order by id desc`, cid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]HeroRecruit, 0, 12)
	for rows.Next() {
		item := HeroRecruit{}
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Sex,
			&item.Face,
			&item.CID,
			&item.Level,
			&item.AffairsBase,
			&item.BraveryBase,
			&item.WisdomBase,
			&item.AffairsAdd,
			&item.BraveryAdd,
			&item.WisdomAdd,
			&item.Loyalty,
			&item.GoldNeed,
		); err != nil {
			return nil, err
		}
		item.Name = strings.TrimSpace(item.Name)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ensureCityRecruitPoolTx(ctx context.Context, tx *sql.Tx, cid int, hotelLevel int) error {
	if hotelLevel <= 0 {
		return nil
	}

	if _, err := tx.ExecContext(ctx, `
insert into mem_city_schedule (cid, last_reset_recruit)
values (?, 0)
on duplicate key update cid = cid`, cid); err != nil {
		return err
	}

	var lastReset int64
	if err := tx.QueryRowContext(ctx, `
select coalesce(last_reset_recruit, 0)
from mem_city_schedule
where cid = ?`, cid).Scan(&lastReset); err != nil {
		return err
	}

	now := time.Now().Unix()
	blockSize := heroRecruitRefreshSeconds / hotelLevel
	if blockSize <= 0 {
		blockSize = 1
	}
	lastBlock := (lastReset + 8*3600) / int64(blockSize)
	currentBlock := (now + 8*3600) / int64(blockSize)
	blockDelta := currentBlock - lastBlock
	if blockDelta > 0 {
		rows, err := tx.QueryContext(ctx, `
select id
from sys_recruit_hero
where cid = ?
order by id asc
limit ?`, cid, blockDelta)
		if err != nil {
			return err
		}
		deleteIDs := make([]int, 0, blockDelta)
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
			deleteIDs = append(deleteIDs, id)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
		for _, id := range deleteIDs {
			if _, err := tx.ExecContext(ctx, "delete from sys_recruit_hero where id = ?", id); err != nil {
				return err
			}
		}
	}

	var recruitCount int
	if err := tx.QueryRowContext(ctx, `
select count(*)
from sys_recruit_hero
where cid = ?`, cid).Scan(&recruitCount); err != nil {
		return err
	}

	if recruitCount < hotelLevel {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := recruitCount; i < hotelLevel; i++ {
			if err := r.generateRecruitHeroTx(ctx, tx, cid, hotelLevel, rng); err != nil {
				return err
			}
		}
	}

	if blockDelta > 0 {
		if _, err := tx.ExecContext(ctx, `
update mem_city_schedule
set last_reset_recruit = ?
where cid = ?`, now, cid); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) generateRecruitHeroTx(ctx context.Context, tx *sql.Tx, cid int, hotelLevel int, rng *rand.Rand) error {
	sex := 1
	if rng.Intn(10) == 0 {
		sex = 0
	}

	name, err := r.randomHeroNameTx(ctx, tx, sex)
	if err != nil {
		return err
	}
	face := rng.Intn(70) + 1001
	if sex == 0 {
		face = rng.Intn(9) + 1
	}

	level := rng.Intn(max(1, hotelLevel*7)) + 1
	var heroExpRaw float64
	if err := tx.QueryRowContext(ctx, `
select coalesce(total_exp, 0)
from cfg_hero_level
where level = ?`, level).Scan(&heroExpRaw); err != nil {
		return err
	}
	heroExp := int64(math.Round(heroExpRaw))

	affairsRate := rng.Intn(71) + 30
	braveryRate := rng.Intn(71) + 30
	wisdomRate := rng.Intn(71) + 30
	allRate := affairsRate + braveryRate + wisdomRate
	allBase := rng.Intn(131) + 30

	affairsBase := allBase * affairsRate / allRate
	braveryBase := allBase * braveryRate / allRate
	wisdomBase := allBase * wisdomRate / allRate

	affairsAdd := int(math.Round(float64(level*affairsRate) / float64(allRate)))
	braveryAdd := int(math.Round(float64(level*braveryRate) / float64(allRate)))
	wisdomAdd := level - affairsAdd - braveryAdd

	loyalty := rng.Intn(81) + 10
	goldNeed := int64(level * 1000)
	_, err = tx.ExecContext(ctx, `
insert into sys_recruit_hero (
	name,
	sex,
	face,
	cid,
	level,
	exp,
	affairs_base,
	bravery_base,
	wisdom_base,
	affairs_add,
	bravery_add,
	wisdom_add,
	loyalty,
	gold_need,
	gen_time
) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		name,
		sex,
		face,
		cid,
		level,
		heroExp,
		affairsBase,
		braveryBase,
		wisdomBase,
		affairsAdd,
		braveryAdd,
		wisdomAdd,
		loyalty,
		goldNeed,
		time.Now().Unix(),
	)
	return err
}

func (r *Repository) randomHeroNameTx(ctx context.Context, tx *sql.Tx, sex int) (string, error) {
	firstName, err := r.randomTableNameTx(ctx, tx, "mem_cfg_firstname")
	if err != nil {
		return "", err
	}
	if sex == 0 {
		lastName, err := r.randomTableNameTx(ctx, tx, "mem_cfg_girlname")
		if err != nil {
			return "", err
		}
		return firstName + lastName, nil
	}

	lastName, err := r.randomTableNameTx(ctx, tx, "mem_cfg_boyname")
	if err != nil {
		return "", err
	}
	return firstName + lastName, nil
}

func (r *Repository) randomTableNameTx(ctx context.Context, tx *sql.Tx, table string) (string, error) {
	query := "select coalesce(name, '') from " + table + " order by rand() limit 1"
	var name string
	if err := tx.QueryRowContext(ctx, query).Scan(&name); err != nil {
		return "", err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", newInvalidError("招募姓名表为空。")
	}
	return name, nil
}

func (r *Repository) cityBuildingLevelTx(ctx context.Context, tx *sql.Tx, cid int, bid int) (int, error) {
	var level sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
select max(level)
from sys_building
where cid = ? and bid = ?`, cid, bid).Scan(&level); err != nil {
		return 0, err
	}
	if !level.Valid {
		return 0, nil
	}
	return int(level.Int64), nil
}

func (r *Repository) cityHasHeroPositionTx(ctx context.Context, tx *sql.Tx, uid int, cid int) (bool, error) {
	officeLevel, err := r.cityBuildingLevelTx(ctx, tx, cid, heroBuildingOffice)
	if err != nil {
		return false, err
	}
	if officeLevel <= 0 {
		return false, nil
	}

	var heroCount int
	if err := tx.QueryRowContext(ctx, `
select count(*)
from sys_city_hero
where cid = ? and uid = ?`, cid, uid).Scan(&heroCount); err != nil {
		return false, err
	}
	return officeLevel > heroCount, nil
}

func fixtureHeroRecruits(cid int) []HeroRecruit {
	return []HeroRecruit{
		{
			ID:          90001,
			Name:        "韩当",
			Sex:         1,
			Face:        1005,
			CID:         cid,
			Level:       12,
			AffairsBase: 42,
			BraveryBase: 61,
			WisdomBase:  37,
			AffairsAdd:  3,
			BraveryAdd:  6,
			WisdomAdd:   3,
			Loyalty:     68,
			GoldNeed:    12000,
		},
		{
			ID:          90002,
			Name:        "甄宓",
			Sex:         0,
			Face:        6,
			CID:         cid,
			Level:       9,
			AffairsBase: 55,
			BraveryBase: 24,
			WisdomBase:  63,
			AffairsAdd:  4,
			BraveryAdd:  1,
			WisdomAdd:   4,
			Loyalty:     74,
			GoldNeed:    9000,
		},
	}
}

func (r *Repository) fixtureRecruitedHeroRoster(cid int, recruitID int) (HeroRoster, error) {
	roster := r.fixtureHeroRoster(cid)
	roster.Count = len(roster.Items) + 1
	roster.Items = append([]HeroCard{
		{
			HID:        90000 + recruitID,
			CID:        cid,
			Name:       "新募将",
			Sex:        1,
			Face:       1006,
			State:      0,
			StateName:  heroStateDisplayLabel(0),
			StateLabel: heroStateDisplayLabel(0),
			Level:      12,
			Loyalty:    heroRecruitDefaultLoyalty,
			Exp:        120000,
			Command:    68,
			Affairs:    46,
			Bravery:    67,
			Wisdom:     40,
			Available:  3,
			Force:      100,
			ForceMax:   124,
			Energy:     100,
			EnergyMax:  117,
		},
	}, roster.Items...)
	if len(roster.Recruits) > 0 {
		filtered := make([]HeroRecruit, 0, len(roster.Recruits)-1)
		for _, item := range roster.Recruits {
			if item.ID != recruitID {
				filtered = append(filtered, item)
			}
		}
		roster.Recruits = filtered
	}
	return roster, nil
}

func max(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
