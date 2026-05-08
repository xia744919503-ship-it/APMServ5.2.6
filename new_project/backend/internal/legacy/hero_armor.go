package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type heroArmorSlotDef struct {
	spart     int
	part      int
	partLabel string
	slotLabel string
}

type heroArmorHeader struct {
	HID   int
	CID   int
	Name  string
	Level int
	State int
}

type heroArmorRecord struct {
	SID        int
	ArmorID    int
	Name       string
	Part       int
	Type       int
	HeroLevel  int
	HP         int
	HPMax      int
	OriHPMax   int
	Value      int
	Attribute  string
	CurrentHID int
}

const (
	heroBuildingMarket      = 13
	heroRenovateLogType     = 100
	heroArmorRecycleLogType = 9
)

var heroArmorSlotDefs = []heroArmorSlotDef{
	{spart: 10, part: 1, partLabel: "头盔", slotLabel: "头盔位"},
	{spart: 20, part: 2, partLabel: "护颈", slotLabel: "护颈位"},
	{spart: 30, part: 3, partLabel: "护肩", slotLabel: "护肩位"},
	{spart: 40, part: 4, partLabel: "铠甲", slotLabel: "铠甲位"},
	{spart: 50, part: 5, partLabel: "披风", slotLabel: "披风位"},
	{spart: 60, part: 6, partLabel: "腰带", slotLabel: "腰带位"},
	{spart: 70, part: 7, partLabel: "护手", slotLabel: "护手位"},
	{spart: 80, part: 8, partLabel: "鞋履", slotLabel: "鞋履位"},
	{spart: 90, part: 9, partLabel: "戒指", slotLabel: "左戒位"},
	{spart: 91, part: 9, partLabel: "戒指", slotLabel: "右戒位"},
	{spart: 100, part: 10, partLabel: "佩印", slotLabel: "佩印位"},
	{spart: 110, part: 11, partLabel: "兵器", slotLabel: "兵器一"},
	{spart: 111, part: 11, partLabel: "兵器", slotLabel: "兵器二"},
	{spart: 112, part: 11, partLabel: "兵器", slotLabel: "兵器三"},
	{spart: 120, part: 12, partLabel: "坐骑", slotLabel: "坐骑位"},
}

func (r *Repository) HeroArmorSnapshot(ctx context.Context, uid int, cid int, hid int) (HeroArmorSnapshot, error) {
	if hid <= 0 {
		return HeroArmorSnapshot{}, newInvalidError("无效的武将编号。")
	}
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return HeroArmorSnapshot{}, err
	} else if !allowed {
		return HeroArmorSnapshot{}, ErrForbidden
	}
	if r.db == nil {
		return r.fixtureHeroArmorSnapshot(cid, hid), nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	defer tx.Rollback()

	snapshot, err := r.heroArmorSnapshotTx(ctx, tx, uid, cid, hid)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return HeroArmorSnapshot{}, err
	}
	return snapshot, nil
}

func (r *Repository) EquipHeroArmor(ctx context.Context, uid int, cid int, hid int, sid int, spart int) (HeroArmorSnapshot, error) {
	if hid <= 0 || sid <= 0 || spart <= 0 {
		return HeroArmorSnapshot{}, newInvalidError("装备穿戴参数无效。")
	}
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return HeroArmorSnapshot{}, err
	} else if !allowed {
		return HeroArmorSnapshot{}, ErrForbidden
	}
	if r.db == nil {
		return r.fixtureEquippedHeroArmorSnapshot(cid, hid, sid, spart), nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	defer tx.Rollback()

	hero, err := r.cityHeroArmorHeaderTx(ctx, tx, uid, cid, hid)
	if err != nil {
		if err == sql.ErrNoRows {
			return HeroArmorSnapshot{}, newInvalidError("未找到该武将。")
		}
		return HeroArmorSnapshot{}, err
	}
	if !heroAllowsPointAllocation(hero.State) {
		return HeroArmorSnapshot{}, newInvalidError("将领不在本城或没有效忠于你。只能给本城内效忠于你的将领换装。")
	}

	slotDef, ok := heroArmorSlotDefByKey(spart)
	if !ok {
		return HeroArmorSnapshot{}, newInvalidError("装备部位无效。")
	}

	armor, err := r.userArmorRecordTx(ctx, tx, uid, sid)
	if err != nil {
		if err == sql.ErrNoRows {
			return HeroArmorSnapshot{}, newInvalidError("该件装备不存在。")
		}
		return HeroArmorSnapshot{}, err
	}
	if armor.Part != slotDef.part {
		return HeroArmorSnapshot{}, newInvalidError("不能装备在这个部位。")
	}
	if armor.CurrentHID != 0 {
		return HeroArmorSnapshot{}, newInvalidError("该件装备已经被其他武将使用了。")
	}
	if armorDurabilityValue(armor.HP) <= 0 {
		return HeroArmorSnapshot{}, newInvalidError("装备已经没有耐久，不能使用，请先修复。")
	}
	if armor.HeroLevel > hero.Level {
		return HeroArmorSnapshot{}, newInvalidError(fmt.Sprintf("这件装备需要将领等级达到%d级才能使用。", armor.HeroLevel))
	}

	if oldSID, err := r.heroArmorSlotSIDTx(ctx, tx, hid, spart); err != nil {
		return HeroArmorSnapshot{}, err
	} else if oldSID > 0 {
		if _, err := tx.ExecContext(ctx, "update sys_user_armor set hid = 0 where sid = ?", oldSID); err != nil {
			return HeroArmorSnapshot{}, err
		}
	}

	if _, err := tx.ExecContext(ctx, "update sys_user_armor set hid = ? where sid = ? and uid = ?", hid, sid, uid); err != nil {
		return HeroArmorSnapshot{}, err
	}
	if _, err := tx.ExecContext(ctx, `
insert into sys_hero_armor (hid, spart, sid, armorid)
values (?, ?, ?, ?)
on duplicate key update sid = values(sid), armorid = values(armorid)`, hid, spart, sid, armor.ArmorID); err != nil {
		return HeroArmorSnapshot{}, err
	}

	if err := r.refreshHeroArmorBonusesTx(ctx, tx, uid, cid, hid, true); err != nil {
		return HeroArmorSnapshot{}, err
	}

	snapshot, err := r.heroArmorSnapshotTx(ctx, tx, uid, cid, hid)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return HeroArmorSnapshot{}, err
	}
	return snapshot, nil
}

func (r *Repository) OffloadHeroArmor(ctx context.Context, uid int, cid int, hid int, spart int) (HeroArmorSnapshot, error) {
	if hid <= 0 || spart <= 0 {
		return HeroArmorSnapshot{}, newInvalidError("装备卸下参数无效。")
	}
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return HeroArmorSnapshot{}, err
	} else if !allowed {
		return HeroArmorSnapshot{}, ErrForbidden
	}
	if r.db == nil {
		return r.fixtureOffloadedHeroArmorSnapshot(cid, hid, spart), nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	defer tx.Rollback()

	hero, err := r.cityHeroArmorHeaderTx(ctx, tx, uid, cid, hid)
	if err != nil {
		if err == sql.ErrNoRows {
			return HeroArmorSnapshot{}, newInvalidError("未找到该武将。")
		}
		return HeroArmorSnapshot{}, err
	}
	if !heroAllowsPointAllocation(hero.State) {
		return HeroArmorSnapshot{}, newInvalidError("将领不在本城或没有效忠于你。只能给本城内效忠于你的将领换装。")
	}

	sid, err := r.heroArmorSlotSIDTx(ctx, tx, hid, spart)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	if sid <= 0 {
		return HeroArmorSnapshot{}, newInvalidError("该件装备不存在。")
	}

	if _, err := tx.ExecContext(ctx, "update sys_user_armor set hid = 0 where sid = ?", sid); err != nil {
		return HeroArmorSnapshot{}, err
	}
	if _, err := tx.ExecContext(ctx, "delete from sys_hero_armor where hid = ? and spart = ?", hid, spart); err != nil {
		return HeroArmorSnapshot{}, err
	}

	if err := r.refreshHeroArmorBonusesTx(ctx, tx, uid, cid, hid, true); err != nil {
		return HeroArmorSnapshot{}, err
	}

	snapshot, err := r.heroArmorSnapshotTx(ctx, tx, uid, cid, hid)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return HeroArmorSnapshot{}, err
	}
	return snapshot, nil
}

func (r *Repository) RepairHeroArmor(ctx context.Context, uid int, cid int, hid int, sid int) (HeroArmorSnapshot, error) {
	if hid <= 0 || sid <= 0 {
		return HeroArmorSnapshot{}, newInvalidError("装备修理参数无效。")
	}
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return HeroArmorSnapshot{}, err
	} else if !allowed {
		return HeroArmorSnapshot{}, ErrForbidden
	}
	if r.db == nil {
		return r.fixtureRepairedHeroArmorSnapshot(cid, hid, sid), nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	defer tx.Rollback()

	if _, err := r.cityHeroArmorHeaderTx(ctx, tx, uid, cid, hid); err != nil {
		if err == sql.ErrNoRows {
			return HeroArmorSnapshot{}, newInvalidError("未找到该武将。")
		}
		return HeroArmorSnapshot{}, err
	}

	armor, err := r.userArmorRecordTx(ctx, tx, uid, sid)
	if err != nil {
		if err == sql.ErrNoRows {
			return HeroArmorSnapshot{}, newInvalidError("该件装备不存在。")
		}
		return HeroArmorSnapshot{}, err
	}
	if armor.CurrentHID > 0 && armor.CurrentHID != hid {
		return HeroArmorSnapshot{}, newInvalidError("请先切换到装备所在武将。")
	}

	hp := armorDurabilityValue(armor.HP)
	if hp <= 0 {
		return HeroArmorSnapshot{}, newInvalidError("装备已经没有耐久，不能修理，只能修复了。")
	}

	goldNeed := int64(armor.HPMax-hp) * 100
	if goldNeed <= 0 {
		return HeroArmorSnapshot{}, newInvalidError("这件装备没有损坏，不需要修理。")
	}

	result, err := tx.ExecContext(ctx, `
update mem_city_resource
set gold = gold - ?
where cid = ? and gold >= ?`, goldNeed, cid, goldNeed)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	if affected == 0 {
		return HeroArmorSnapshot{}, newInvalidError("本城黄金不足。")
	}

	reduce := max(1, (armor.HPMax-hp+9)/10)
	newHPMax := max(0, armor.HPMax-reduce)
	if _, err := tx.ExecContext(ctx, `
update sys_user_armor
set hp = ?, hp_max = ?
where sid = ? and uid = ?`, newHPMax*10, newHPMax, sid, uid); err != nil {
		return HeroArmorSnapshot{}, err
	}

	if armor.CurrentHID == hid {
		if err := r.refreshHeroArmorBonusesTx(ctx, tx, uid, cid, hid, false); err != nil {
			return HeroArmorSnapshot{}, err
		}
	}

	snapshot, err := r.heroArmorSnapshotTx(ctx, tx, uid, cid, hid)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return HeroArmorSnapshot{}, err
	}
	return snapshot, nil
}

func (r *Repository) RepairAllHeroArmor(ctx context.Context, uid int, cid int, hid int, sids []int) (HeroArmorSnapshot, error) {
	batchSIDs := normalizeHeroArmorBatchSIDs(sids)
	if hid <= 0 || len(batchSIDs) == 0 {
		return HeroArmorSnapshot{}, newInvalidError("装备批量修理参数无效。")
	}
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return HeroArmorSnapshot{}, err
	} else if !allowed {
		return HeroArmorSnapshot{}, ErrForbidden
	}
	if r.db == nil {
		return r.fixtureRepairAllHeroArmorSnapshot(cid, hid, batchSIDs), nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	defer tx.Rollback()

	if _, err := r.cityHeroArmorHeaderTx(ctx, tx, uid, cid, hid); err != nil {
		if err == sql.ErrNoRows {
			return HeroArmorSnapshot{}, newInvalidError("未找到该武将。")
		}
		return HeroArmorSnapshot{}, err
	}

	armors := make([]heroArmorRecord, 0, len(batchSIDs))
	var goldNeed int64
	hasEquippedArmor := false
	for _, sid := range batchSIDs {
		armor, err := r.userArmorRecordTx(ctx, tx, uid, sid)
		if err != nil {
			if err == sql.ErrNoRows {
				return HeroArmorSnapshot{}, newInvalidError("该件装备不存在。")
			}
			return HeroArmorSnapshot{}, err
		}
		if armor.CurrentHID > 0 && armor.CurrentHID != hid {
			return HeroArmorSnapshot{}, newInvalidError("请先切换到装备所在武将。")
		}

		hp := armorDurabilityValue(armor.HP)
		if hp <= 0 {
			return HeroArmorSnapshot{}, newInvalidError("装备已经没有耐久，不能修理，只能修复了。")
		}

		goldNeed += int64(armor.HPMax-hp) * 100
		if armor.CurrentHID == hid {
			hasEquippedArmor = true
		}
		armors = append(armors, armor)
	}
	if goldNeed <= 0 {
		return HeroArmorSnapshot{}, newInvalidError("这批装备没有损坏，不需要修理。")
	}

	result, err := tx.ExecContext(ctx, `
update mem_city_resource
set gold = gold - ?
where cid = ? and gold >= ?`, goldNeed, cid, goldNeed)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	if affected == 0 {
		return HeroArmorSnapshot{}, newInvalidError("本城黄金不足。")
	}

	for _, armor := range armors {
		hp := armorDurabilityValue(armor.HP)
		reduce := max(1, (armor.HPMax-hp+9)/10)
		newHPMax := max(0, armor.HPMax-reduce)
		if _, err := tx.ExecContext(ctx, `
update sys_user_armor
set hp = ?, hp_max = ?
where sid = ? and uid = ?`, newHPMax*10, newHPMax, armor.SID, uid); err != nil {
			return HeroArmorSnapshot{}, err
		}
	}

	if hasEquippedArmor {
		if err := r.refreshHeroArmorBonusesTx(ctx, tx, uid, cid, hid, false); err != nil {
			return HeroArmorSnapshot{}, err
		}
	}

	snapshot, err := r.heroArmorSnapshotTx(ctx, tx, uid, cid, hid)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return HeroArmorSnapshot{}, err
	}
	return snapshot, nil
}

func (r *Repository) RenovateHeroArmor(ctx context.Context, uid int, cid int, hid int, sid int) (HeroArmorSnapshot, error) {
	if hid <= 0 || sid <= 0 {
		return HeroArmorSnapshot{}, newInvalidError("装备修复参数无效。")
	}
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return HeroArmorSnapshot{}, err
	} else if !allowed {
		return HeroArmorSnapshot{}, ErrForbidden
	}
	if r.db == nil {
		return r.fixtureRenovatedHeroArmorSnapshot(cid, hid, sid), nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	defer tx.Rollback()

	if _, err := r.cityHeroArmorHeaderTx(ctx, tx, uid, cid, hid); err != nil {
		if err == sql.ErrNoRows {
			return HeroArmorSnapshot{}, newInvalidError("未找到该武将。")
		}
		return HeroArmorSnapshot{}, err
	}

	armor, err := r.userArmorRecordTx(ctx, tx, uid, sid)
	if err != nil {
		if err == sql.ErrNoRows {
			return HeroArmorSnapshot{}, newInvalidError("该件装备不存在。")
		}
		return HeroArmorSnapshot{}, err
	}
	if armor.CurrentHID > 0 && armor.CurrentHID != hid {
		return HeroArmorSnapshot{}, newInvalidError("请先切换到装备所在武将。")
	}
	if armor.OriHPMax <= 0 {
		return HeroArmorSnapshot{}, newInvalidError("该件装备数据异常，无法修复。")
	}

	hp := armorDurabilityValue(armor.HP)
	moneyNeed := int64(armor.OriHPMax-armor.HPMax) + int64((armor.HPMax-hp+9)/10)
	if moneyNeed <= 0 {
		return HeroArmorSnapshot{}, newInvalidError("这件装备没有损毁，不需要修复。")
	}

	result, err := tx.ExecContext(ctx, `
update sys_user
set money = money - ?
where uid = ? and money >= ?`, moneyNeed, uid, moneyNeed)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	if affected == 0 {
		return HeroArmorSnapshot{}, newInvalidError("你没有足够的元宝，请充值后再修复。")
	}

	now := time.Now().Unix()
	if _, err := tx.ExecContext(ctx, `
insert into log_money (uid, count, time, type)
values (?, ?, ?, ?)`, uid, -moneyNeed, now, heroRenovateLogType); err != nil {
		return HeroArmorSnapshot{}, err
	}

	if _, err := tx.ExecContext(ctx, `
update sys_user_armor
set hp = ?, hp_max = ?
where sid = ? and uid = ?`, armor.OriHPMax*10, armor.OriHPMax, sid, uid); err != nil {
		return HeroArmorSnapshot{}, err
	}

	if armor.CurrentHID == hid {
		if err := r.refreshHeroArmorBonusesTx(ctx, tx, uid, cid, hid, false); err != nil {
			return HeroArmorSnapshot{}, err
		}
	}

	snapshot, err := r.heroArmorSnapshotTx(ctx, tx, uid, cid, hid)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return HeroArmorSnapshot{}, err
	}
	return snapshot, nil
}

func (r *Repository) RenovateAllHeroArmor(ctx context.Context, uid int, cid int, hid int, sids []int) (HeroArmorSnapshot, error) {
	batchSIDs := normalizeHeroArmorBatchSIDs(sids)
	if hid <= 0 || len(batchSIDs) == 0 {
		return HeroArmorSnapshot{}, newInvalidError("装备批量修复参数无效。")
	}
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return HeroArmorSnapshot{}, err
	} else if !allowed {
		return HeroArmorSnapshot{}, ErrForbidden
	}
	if r.db == nil {
		return r.fixtureRenovateAllHeroArmorSnapshot(cid, hid, batchSIDs), nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	defer tx.Rollback()

	if _, err := r.cityHeroArmorHeaderTx(ctx, tx, uid, cid, hid); err != nil {
		if err == sql.ErrNoRows {
			return HeroArmorSnapshot{}, newInvalidError("未找到该武将。")
		}
		return HeroArmorSnapshot{}, err
	}

	armors := make([]heroArmorRecord, 0, len(batchSIDs))
	var moneyNeed int64
	hasEquippedArmor := false
	for _, sid := range batchSIDs {
		armor, err := r.userArmorRecordTx(ctx, tx, uid, sid)
		if err != nil {
			if err == sql.ErrNoRows {
				return HeroArmorSnapshot{}, newInvalidError("该件装备不存在。")
			}
			return HeroArmorSnapshot{}, err
		}
		if armor.CurrentHID > 0 && armor.CurrentHID != hid {
			return HeroArmorSnapshot{}, newInvalidError("请先切换到装备所在武将。")
		}
		if armor.OriHPMax <= 0 {
			return HeroArmorSnapshot{}, newInvalidError("该件装备数据异常，无法修复。")
		}

		hp := armorDurabilityValue(armor.HP)
		moneyNeed += int64(armor.OriHPMax-armor.HPMax) + int64((armor.HPMax-hp+9)/10)
		if armor.CurrentHID == hid {
			hasEquippedArmor = true
		}
		armors = append(armors, armor)
	}
	if moneyNeed <= 0 {
		return HeroArmorSnapshot{}, newInvalidError("这批装备没有损毁，不需要修复。")
	}

	result, err := tx.ExecContext(ctx, `
update sys_user
set money = money - ?
where uid = ? and money >= ?`, moneyNeed, uid, moneyNeed)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	if affected == 0 {
		return HeroArmorSnapshot{}, newInvalidError("你没有足够的元宝，请充值后再修复。")
	}

	now := time.Now().Unix()
	if _, err := tx.ExecContext(ctx, `
insert into log_money (uid, count, time, type)
values (?, ?, ?, ?)`, uid, -moneyNeed, now, heroRenovateLogType); err != nil {
		return HeroArmorSnapshot{}, err
	}

	for _, armor := range armors {
		if _, err := tx.ExecContext(ctx, `
update sys_user_armor
set hp = ?, hp_max = ?
where sid = ? and uid = ?`, armor.OriHPMax*10, armor.OriHPMax, armor.SID, uid); err != nil {
			return HeroArmorSnapshot{}, err
		}
	}

	if hasEquippedArmor {
		if err := r.refreshHeroArmorBonusesTx(ctx, tx, uid, cid, hid, false); err != nil {
			return HeroArmorSnapshot{}, err
		}
	}

	snapshot, err := r.heroArmorSnapshotTx(ctx, tx, uid, cid, hid)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return HeroArmorSnapshot{}, err
	}
	return snapshot, nil
}

func (r *Repository) RecycleHeroArmor(ctx context.Context, uid int, cid int, hid int, sid int) (HeroArmorSnapshot, error) {
	if hid <= 0 || sid <= 0 {
		return HeroArmorSnapshot{}, newInvalidError("装备回收参数无效。")
	}
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return HeroArmorSnapshot{}, err
	} else if !allowed {
		return HeroArmorSnapshot{}, ErrForbidden
	}
	if r.db == nil {
		return r.fixtureRecycledHeroArmorSnapshot(cid, hid, sid), nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	defer tx.Rollback()

	if _, err := r.cityHeroArmorHeaderTx(ctx, tx, uid, cid, hid); err != nil {
		if err == sql.ErrNoRows {
			return HeroArmorSnapshot{}, newInvalidError("未找到该武将。")
		}
		return HeroArmorSnapshot{}, err
	}

	marketLevel, err := r.cityBuildingLevelTx(ctx, tx, cid, heroBuildingMarket)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	if marketLevel < 5 {
		return HeroArmorSnapshot{}, newInvalidError("市场达到5级才能回收装备。")
	}

	nobility, err := r.effectiveUserNobilityTx(ctx, tx, uid)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	if nobility < 1 {
		return HeroArmorSnapshot{}, newInvalidError("爵位达到“公士”才能回收装备。")
	}

	armor, err := r.userArmorRecordTx(ctx, tx, uid, sid)
	if err != nil {
		if err == sql.ErrNoRows {
			return HeroArmorSnapshot{}, newInvalidError("该件装备不存在。")
		}
		return HeroArmorSnapshot{}, err
	}
	if armor.CurrentHID != 0 {
		return HeroArmorSnapshot{}, newInvalidError("请先卸下装备后再回收。")
	}
	if armor.OriHPMax <= 0 {
		return HeroArmorSnapshot{}, newInvalidError("该件装备数据异常，无法回收。")
	}

	hp := armorDurabilityValue(armor.HP)
	goldAdd := int64(max(1, (hp/armor.OriHPMax))*armor.Value) * 500
	if _, err := tx.ExecContext(ctx, `
update mem_city_resource
set gold = gold + ?
where cid = ?`, goldAdd, cid); err != nil {
		return HeroArmorSnapshot{}, err
	}
	if _, err := tx.ExecContext(ctx, `
delete from sys_user_armor
where sid = ? and uid = ?`, sid, uid); err != nil {
		return HeroArmorSnapshot{}, err
	}

	now := time.Now().Unix()
	if _, err := tx.ExecContext(ctx, `
insert into log_armor (uid, armorid, count, time, type)
values (?, ?, ?, ?, ?)`, uid, armor.ArmorID, -1, now, heroArmorRecycleLogType); err != nil {
		return HeroArmorSnapshot{}, err
	}

	snapshot, err := r.heroArmorSnapshotTx(ctx, tx, uid, cid, hid)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return HeroArmorSnapshot{}, err
	}
	return snapshot, nil
}

func (r *Repository) heroArmorSnapshotTx(ctx context.Context, tx *sql.Tx, uid int, cid int, hid int) (HeroArmorSnapshot, error) {
	hero, err := r.cityHeroArmorHeaderTx(ctx, tx, uid, cid, hid)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}

	equippedRows, err := r.heroArmorEquippedTx(ctx, tx, uid, hid)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	equippedBySpart := make(map[int]HeroArmorItem, len(equippedRows))
	for spart, row := range equippedRows {
		item := buildHeroArmorItem(row)
		item.Equipped = true
		item.SlotKey = spart
		if slotDef, ok := heroArmorSlotDefByKey(spart); ok {
			item.PartLabel = slotDef.partLabel
			item.SlotLabel = slotDef.slotLabel
		}
		equippedBySpart[spart] = item
	}

	slots := make([]HeroArmorSlot, 0, len(heroArmorSlotDefs))
	for _, slotDef := range heroArmorSlotDefs {
		slot := HeroArmorSlot{
			Spart:     slotDef.spart,
			Part:      slotDef.part,
			PartLabel: slotDef.partLabel,
			SlotLabel: slotDef.slotLabel,
		}
		if item, ok := equippedBySpart[slotDef.spart]; ok {
			copyItem := item
			slot.Equipped = &copyItem
		}
		slots = append(slots, slot)
	}

	inventoryRows, err := r.userArmorInventoryTx(ctx, tx, uid)
	if err != nil {
		return HeroArmorSnapshot{}, err
	}
	inventory := make([]HeroArmorItem, 0, len(inventoryRows))
	for _, row := range inventoryRows {
		item := buildHeroArmorItem(row)
		item.Equipped = false
		inventory = append(inventory, item)
	}

	snapshot := HeroArmorSnapshot{
		HID:            hero.HID,
		CID:            hero.CID,
		HeroName:       hero.Name,
		HeroLevel:      hero.Level,
		HeroState:      hero.State,
		HeroStateLabel: heroStateDisplayLabel(hero.State),
		Slots:          slots,
		Inventory:      inventory,
	}
	normalizeHeroArmorSnapshotDerived(&snapshot)
	return snapshot, nil
}

func (r *Repository) cityHeroArmorHeaderTx(ctx context.Context, tx *sql.Tx, uid int, cid int, hid int) (heroArmorHeader, error) {
	row := heroArmorHeader{}
	err := tx.QueryRowContext(ctx, `
select hid, cid, coalesce(name, ''), coalesce(level, 0), coalesce(state, 0)
from sys_city_hero
where uid = ? and cid = ? and hid = ?`, uid, cid, hid).Scan(
		&row.HID,
		&row.CID,
		&row.Name,
		&row.Level,
		&row.State,
	)
	if err != nil {
		return heroArmorHeader{}, err
	}
	row.Name = strings.TrimSpace(row.Name)
	if row.Name == "" {
		row.Name = fmt.Sprintf("HID %d", row.HID)
	}
	return row, nil
}

func (r *Repository) heroArmorEquippedTx(ctx context.Context, tx *sql.Tx, uid int, hid int) (map[int]heroArmorRecord, error) {
	rows, err := tx.QueryContext(ctx, `
select
	h.spart,
	u.sid,
	h.armorid,
	coalesce(c.name, ''),
	coalesce(c.part, 0),
	coalesce(c.type, 0),
	coalesce(c.hero_level, 0),
	coalesce(u.hp, 0),
	coalesce(u.hp_max, 0),
	coalesce(c.ori_hp_max, 0),
	coalesce(c.value, 0),
	coalesce(c.attribute, '')
from sys_hero_armor h
join sys_user_armor u on u.sid = h.sid and u.hid = h.hid
left join cfg_armor c on c.id = h.armorid
where h.hid = ? and u.uid = ?`, hid, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make(map[int]heroArmorRecord)
	for rows.Next() {
		var spart int
		record := heroArmorRecord{}
		if err := rows.Scan(
			&spart,
			&record.SID,
			&record.ArmorID,
			&record.Name,
			&record.Part,
			&record.Type,
			&record.HeroLevel,
			&record.HP,
			&record.HPMax,
			&record.OriHPMax,
			&record.Value,
			&record.Attribute,
		); err != nil {
			return nil, err
		}
		record.Name = strings.TrimSpace(record.Name)
		items[spart] = record
	}
	return items, rows.Err()
}

func (r *Repository) userArmorInventoryTx(ctx context.Context, tx *sql.Tx, uid int) ([]heroArmorRecord, error) {
	rows, err := tx.QueryContext(ctx, `
select
	u.sid,
	u.armorid,
	coalesce(c.name, ''),
	coalesce(c.part, 0),
	coalesce(c.type, 0),
	coalesce(c.hero_level, 0),
	coalesce(u.hp, 0),
	coalesce(u.hp_max, 0),
	coalesce(c.ori_hp_max, 0),
	coalesce(c.value, 0),
	coalesce(c.attribute, ''),
	coalesce(u.hid, 0)
from sys_user_armor u
left join cfg_armor c on c.id = u.armorid
where u.uid = ? and u.hid = 0
order by c.part asc, c.hero_level desc, u.sid desc`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]heroArmorRecord, 0)
	for rows.Next() {
		record := heroArmorRecord{}
		if err := rows.Scan(
			&record.SID,
			&record.ArmorID,
			&record.Name,
			&record.Part,
			&record.Type,
			&record.HeroLevel,
			&record.HP,
			&record.HPMax,
			&record.OriHPMax,
			&record.Value,
			&record.Attribute,
			&record.CurrentHID,
		); err != nil {
			return nil, err
		}
		record.Name = strings.TrimSpace(record.Name)
		items = append(items, record)
	}
	return items, rows.Err()
}

func (r *Repository) userArmorRecordTx(ctx context.Context, tx *sql.Tx, uid int, sid int) (heroArmorRecord, error) {
	record := heroArmorRecord{}
	err := tx.QueryRowContext(ctx, `
select
	u.sid,
	u.armorid,
	coalesce(c.name, ''),
	coalesce(c.part, 0),
	coalesce(c.type, 0),
	coalesce(c.hero_level, 0),
	coalesce(u.hp, 0),
	coalesce(u.hp_max, 0),
	coalesce(c.ori_hp_max, 0),
	coalesce(c.value, 0),
	coalesce(c.attribute, ''),
	coalesce(u.hid, 0)
from sys_user_armor u
left join cfg_armor c on c.id = u.armorid
where u.uid = ? and u.sid = ?`, uid, sid).Scan(
		&record.SID,
		&record.ArmorID,
		&record.Name,
		&record.Part,
		&record.Type,
		&record.HeroLevel,
		&record.HP,
		&record.HPMax,
		&record.OriHPMax,
		&record.Value,
		&record.Attribute,
		&record.CurrentHID,
	)
	if err != nil {
		return heroArmorRecord{}, err
	}
	record.Name = strings.TrimSpace(record.Name)
	return record, nil
}

func (r *Repository) heroArmorSlotSIDTx(ctx context.Context, tx *sql.Tx, hid int, spart int) (int, error) {
	var sid sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
select sid
from sys_hero_armor
where hid = ? and spart = ?`, hid, spart).Scan(&sid); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	if !sid.Valid {
		return 0, nil
	}
	return int(sid.Int64), nil
}

func (r *Repository) refreshHeroArmorBonusesTx(ctx context.Context, tx *sql.Tx, uid int, cid int, hid int, updateChiefResource bool) error {
	hero, err := r.cityHeroPointRecord(ctx, tx, uid, cid, hid)
	if err != nil {
		return err
	}

	bonuses, err := r.loadHeroAttributeBonuses(ctx, tx, uid, hid)
	if err != nil {
		return err
	}

	braveryBuffed, wisdomBuffed, err := r.loadHeroPointBuffs(ctx, tx, hid)
	if err != nil {
		return err
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
		return err
	}

	if err := r.upsertHeroBlood(ctx, tx, hid, forceMax, energyMax); err != nil {
		return err
	}

	if updateChiefResource && hero.State == 1 {
		if err := r.ensureCityResAdd(ctx, tx, cid); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "update sys_city_res_add set resource_changing = 1 where cid = ?", cid); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) effectiveUserNobilityTx(ctx context.Context, tx *sql.Tx, uid int) (int, error) {
	var nobility int
	if err := tx.QueryRowContext(ctx, `
select coalesce(nobility, 0)
from sys_user
where uid = ?`, uid).Scan(&nobility); err != nil {
		return 0, err
	}

	var bufparam sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
select bufparam
from mem_user_buffer
where uid = ? and buftype in (16, 18)
order by bufparam desc
limit 1`, uid).Scan(&bufparam); err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if !bufparam.Valid {
		return nobility, nil
	}

	effective := nobility + int(bufparam.Int64)
	switch int(bufparam.Int64) {
	case 5:
		if effective > 19 {
			effective = 19
		}
	case 2:
		if effective > 18 {
			effective = 18
		}
	}
	if effective > nobility {
		return effective, nil
	}
	return nobility, nil
}

func buildHeroArmorItem(record heroArmorRecord) HeroArmorItem {
	partLabel := heroArmorPartLabel(record.Part)
	item := HeroArmorItem{
		SID:           record.SID,
		ArmorID:       record.ArmorID,
		Name:          firstNonEmpty(record.Name, fmt.Sprintf("装备 #%d", record.ArmorID)),
		Part:          record.Part,
		PartLabel:     partLabel,
		Type:          record.Type,
		HeroLevel:     record.HeroLevel,
		Durability:    armorDurabilityValue(record.HP),
		DurabilityMax: clamp(record.HPMax, 0, 1<<30),
		OriginalDurabilityMax: clamp(record.OriHPMax, 0, 1<<30),
		RecycleGold:   heroArmorRecycleGoldValue(record.Value),
		Attributes:    parseHeroArmorAttributes(record.Attribute),
	}
	normalizeHeroArmorItemDerived(&item)
	return item
}

func normalizeHeroArmorSnapshotDerived(snapshot *HeroArmorSnapshot) {
	if snapshot == nil {
		return
	}
	for index := range snapshot.Slots {
		if snapshot.Slots[index].Equipped == nil {
			continue
		}
		normalizeHeroArmorItemDerived(snapshot.Slots[index].Equipped)
	}
	for index := range snapshot.Inventory {
		normalizeHeroArmorItemDerived(&snapshot.Inventory[index])
	}
}

func normalizeHeroArmorItemDerived(item *HeroArmorItem) {
	if item == nil {
		return
	}
	if item.OriginalDurabilityMax <= 0 {
		item.OriginalDurabilityMax = max(item.DurabilityMax, item.Durability)
	}
	if item.OriginalDurabilityMax < item.DurabilityMax {
		item.OriginalDurabilityMax = item.DurabilityMax
	}

	durabilityGap := max(0, item.DurabilityMax-item.Durability)
	durabilityLoss := max(0, item.OriginalDurabilityMax-item.DurabilityMax)

	item.RepairGoldNeed = int64(durabilityGap) * 100
	item.RenovateMoneyNeed = int64(durabilityLoss) + int64((durabilityGap+9)/10)

	if item.Equipped {
		item.RecycleGold = 0
	}
}

func heroArmorRecycleGoldValue(value int) int64 {
	if value <= 0 {
		return 0
	}
	return int64(value) * 500
}

func parseHeroArmorAttributes(raw string) []HeroArmorAttribute {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	if len(parts) < 3 {
		return []HeroArmorAttribute{}
	}

	count, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || count <= 0 {
		return []HeroArmorAttribute{}
	}

	items := make([]HeroArmorAttribute, 0, count)
	for index := 1; index+1 < len(parts); index += 2 {
		attributeType, err := strconv.Atoi(strings.TrimSpace(parts[index]))
		if err != nil {
			continue
		}
		value, err := strconv.Atoi(strings.TrimSpace(parts[index+1]))
		if err != nil {
			continue
		}
		items = append(items, HeroArmorAttribute{
			Type:  attributeType,
			Label: heroArmorAttributeLabel(attributeType),
			Value: value,
		})
	}
	return items
}

func heroArmorAttributeLabel(attributeType int) string {
	switch attributeType {
	case 1:
		return "统率"
	case 2:
		return "内政"
	case 3:
		return "勇武"
	case 4:
		return "智谋"
	case 5:
		return "体力"
	case 6:
		return "精力"
	case 7:
		return "生命"
	case 8:
		return "攻击"
	case 9:
		return "防御"
	case 10:
		return "射程"
	case 11:
		return "速度"
	case 12:
		return "负重"
	default:
		return fmt.Sprintf("属性%d", attributeType)
	}
}

func heroArmorPartLabel(part int) string {
	for _, slotDef := range heroArmorSlotDefs {
		if slotDef.part == part {
			return slotDef.partLabel
		}
	}
	return fmt.Sprintf("部位%d", part)
}

func heroArmorSlotDefByKey(spart int) (heroArmorSlotDef, bool) {
	for _, slotDef := range heroArmorSlotDefs {
		if slotDef.spart == spart {
			return slotDef, true
		}
	}
	return heroArmorSlotDef{}, false
}

func armorDurabilityValue(hp int) int {
	if hp <= 0 {
		return 0
	}
	return (hp + 9) / 10
}

func normalizeHeroArmorBatchSIDs(sids []int) []int {
	result := make([]int, 0, len(sids))
	seen := make(map[int]struct{}, len(sids))
	for _, sid := range sids {
		if sid <= 0 {
			continue
		}
		if _, ok := seen[sid]; ok {
			continue
		}
		seen[sid] = struct{}{}
		result = append(result, sid)
	}
	return result
}

func (r *Repository) fixtureHeroArmorSnapshot(cid int, hid int) HeroArmorSnapshot {
	roster := r.fixtureHeroRoster(cid)
	hero := roster.Items[0]
	for _, item := range roster.Items {
		if item.HID == hid {
			hero = item
			break
		}
	}

	slots := make([]HeroArmorSlot, 0, len(heroArmorSlotDefs))
	for _, slotDef := range heroArmorSlotDefs {
		slot := HeroArmorSlot{
			Spart:     slotDef.spart,
			Part:      slotDef.part,
			PartLabel: slotDef.partLabel,
			SlotLabel: slotDef.slotLabel,
		}
		if slotDef.spart == 10 {
			equipped := HeroArmorItem{
				SID:           70001,
				ArmorID:       50000,
				Name:          "新手头盔",
				Part:          1,
				PartLabel:     slotDef.partLabel,
				SlotKey:       slotDef.spart,
				SlotLabel:     slotDef.slotLabel,
				Type:          3,
				HeroLevel:     10,
				Durability:    100,
				DurabilityMax: 100,
				OriginalDurabilityMax: 100,
				RecycleGold:   500,
				Equipped:      true,
				Attributes: []HeroArmorAttribute{
					{Type: 3, Label: "勇武", Value: 6},
				},
			}
			slot.Equipped = &equipped
		}
		slots = append(slots, slot)
	}

	snapshot := HeroArmorSnapshot{
		HID:            hero.HID,
		CID:            cid,
		HeroName:       hero.Name,
		HeroLevel:      hero.Level,
		HeroState:      hero.State,
		HeroStateLabel: hero.StateLabel,
		Slots:          slots,
		Inventory: []HeroArmorItem{
			{
				SID:           70011,
				ArmorID:       50001,
				Name:          "新手护颈",
				Part:          2,
				PartLabel:     "护颈",
				Type:          3,
				HeroLevel:     10,
				Durability:    100,
				DurabilityMax: 100,
				OriginalDurabilityMax: 100,
				RecycleGold:   500,
				Attributes: []HeroArmorAttribute{
					{Type: 2, Label: "内政", Value: 4},
				},
			},
			{
				SID:           70012,
				ArmorID:       1431,
				Name:          "小乔的玉戒",
				Part:          9,
				PartLabel:     "戒指",
				Type:          3,
				HeroLevel:     30,
				Durability:    120,
				DurabilityMax: 120,
				OriginalDurabilityMax: 120,
				RecycleGold:   1000,
				Attributes: []HeroArmorAttribute{
					{Type: 4, Label: "智谋", Value: 10},
				},
			},
			{
				SID:           70013,
				ArmorID:       1436,
				Name:          "落月",
				Part:          12,
				PartLabel:     "坐骑",
				Type:          3,
				HeroLevel:     30,
				Durability:    100,
				DurabilityMax: 100,
				OriginalDurabilityMax: 100,
				RecycleGold:   1000,
				Attributes: []HeroArmorAttribute{
					{Type: 11, Label: "速度", Value: 12},
				},
			},
		},
	}
	normalizeHeroArmorSnapshotDerived(&snapshot)
	return snapshot
}

func (r *Repository) fixtureEquippedHeroArmorSnapshot(cid int, hid int, sid int, spart int) HeroArmorSnapshot {
	snapshot := r.fixtureHeroArmorSnapshot(cid, hid)
	slotDef, ok := heroArmorSlotDefByKey(spart)
	if !ok {
		normalizeHeroArmorSnapshotDerived(&snapshot)
		return snapshot
	}

	selectedIndex := -1
	for index := range snapshot.Inventory {
		if snapshot.Inventory[index].SID == sid {
			selectedIndex = index
			break
		}
	}
	if selectedIndex == -1 {
		for index := range snapshot.Inventory {
			if snapshot.Inventory[index].Part == slotDef.part {
				selectedIndex = index
				break
			}
		}
	}
	if selectedIndex == -1 {
		normalizeHeroArmorSnapshotDerived(&snapshot)
		return snapshot
	}

	item := snapshot.Inventory[selectedIndex]
	item.Equipped = true
	item.SlotKey = spart
	item.SlotLabel = slotDef.slotLabel
	item.PartLabel = slotDef.partLabel

	for index := range snapshot.Slots {
		if snapshot.Slots[index].Spart == spart {
			copyItem := item
			snapshot.Slots[index].Equipped = &copyItem
			break
		}
	}

	snapshot.Inventory = append(snapshot.Inventory[:selectedIndex], snapshot.Inventory[selectedIndex+1:]...)
	normalizeHeroArmorSnapshotDerived(&snapshot)
	return snapshot
}

func (r *Repository) fixtureOffloadedHeroArmorSnapshot(cid int, hid int, spart int) HeroArmorSnapshot {
	snapshot := r.fixtureHeroArmorSnapshot(cid, hid)
	for index := range snapshot.Slots {
		if snapshot.Slots[index].Spart != spart || snapshot.Slots[index].Equipped == nil {
			continue
		}
		item := *snapshot.Slots[index].Equipped
		item.Equipped = false
		item.SlotKey = 0
		item.SlotLabel = ""
		snapshot.Inventory = append([]HeroArmorItem{item}, snapshot.Inventory...)
		snapshot.Slots[index].Equipped = nil
		break
	}
	normalizeHeroArmorSnapshotDerived(&snapshot)
	return snapshot
}

func (r *Repository) fixtureRepairedHeroArmorSnapshot(cid int, hid int, sid int) HeroArmorSnapshot {
	snapshot := r.fixtureHeroArmorSnapshot(cid, hid)
	for index := range snapshot.Inventory {
		if snapshot.Inventory[index].SID != sid {
			continue
		}
		item := &snapshot.Inventory[index]
		if item.Durability < item.DurabilityMax {
			reduce := max(1, (item.DurabilityMax-item.Durability+9)/10)
			item.DurabilityMax = max(0, item.DurabilityMax-reduce)
			item.Durability = item.DurabilityMax
		}
		normalizeHeroArmorSnapshotDerived(&snapshot)
		return snapshot
	}
	for index := range snapshot.Slots {
		if snapshot.Slots[index].Equipped == nil || snapshot.Slots[index].Equipped.SID != sid {
			continue
		}
		item := snapshot.Slots[index].Equipped
		if item.Durability < item.DurabilityMax {
			reduce := max(1, (item.DurabilityMax-item.Durability+9)/10)
			item.DurabilityMax = max(0, item.DurabilityMax-reduce)
			item.Durability = item.DurabilityMax
		}
		normalizeHeroArmorSnapshotDerived(&snapshot)
		return snapshot
	}
	normalizeHeroArmorSnapshotDerived(&snapshot)
	return snapshot
}

func (r *Repository) fixtureRenovatedHeroArmorSnapshot(cid int, hid int, sid int) HeroArmorSnapshot {
	snapshot := r.fixtureHeroArmorSnapshot(cid, hid)
	for index := range snapshot.Inventory {
		if snapshot.Inventory[index].SID == sid {
			item := &snapshot.Inventory[index]
			item.DurabilityMax = max(item.DurabilityMax, item.Durability)
			item.Durability = item.DurabilityMax
			normalizeHeroArmorSnapshotDerived(&snapshot)
			return snapshot
		}
	}
	for index := range snapshot.Slots {
		if snapshot.Slots[index].Equipped != nil && snapshot.Slots[index].Equipped.SID == sid {
			item := snapshot.Slots[index].Equipped
			item.DurabilityMax = max(item.DurabilityMax, item.Durability)
			item.Durability = item.DurabilityMax
			normalizeHeroArmorSnapshotDerived(&snapshot)
			return snapshot
		}
	}
	normalizeHeroArmorSnapshotDerived(&snapshot)
	return snapshot
}

func (r *Repository) fixtureRepairAllHeroArmorSnapshot(cid int, hid int, sids []int) HeroArmorSnapshot {
	snapshot := r.fixtureHeroArmorSnapshot(cid, hid)
	sidSet := make(map[int]struct{}, len(sids))
	for _, sid := range normalizeHeroArmorBatchSIDs(sids) {
		sidSet[sid] = struct{}{}
	}

	for index := range snapshot.Inventory {
		if _, ok := sidSet[snapshot.Inventory[index].SID]; !ok {
			continue
		}
		item := &snapshot.Inventory[index]
		if item.Durability < item.DurabilityMax {
			reduce := max(1, (item.DurabilityMax-item.Durability+9)/10)
			item.DurabilityMax = max(0, item.DurabilityMax-reduce)
			item.Durability = item.DurabilityMax
		}
	}
	for index := range snapshot.Slots {
		if snapshot.Slots[index].Equipped == nil {
			continue
		}
		if _, ok := sidSet[snapshot.Slots[index].Equipped.SID]; !ok {
			continue
		}
		item := snapshot.Slots[index].Equipped
		if item.Durability < item.DurabilityMax {
			reduce := max(1, (item.DurabilityMax-item.Durability+9)/10)
			item.DurabilityMax = max(0, item.DurabilityMax-reduce)
			item.Durability = item.DurabilityMax
		}
	}
	normalizeHeroArmorSnapshotDerived(&snapshot)
	return snapshot
}

func (r *Repository) fixtureRenovateAllHeroArmorSnapshot(cid int, hid int, sids []int) HeroArmorSnapshot {
	snapshot := r.fixtureHeroArmorSnapshot(cid, hid)
	sidSet := make(map[int]struct{}, len(sids))
	for _, sid := range normalizeHeroArmorBatchSIDs(sids) {
		sidSet[sid] = struct{}{}
	}

	for index := range snapshot.Inventory {
		if _, ok := sidSet[snapshot.Inventory[index].SID]; !ok {
			continue
		}
		item := &snapshot.Inventory[index]
		item.DurabilityMax = max(item.DurabilityMax, item.Durability)
		item.Durability = item.DurabilityMax
	}
	for index := range snapshot.Slots {
		if snapshot.Slots[index].Equipped == nil {
			continue
		}
		if _, ok := sidSet[snapshot.Slots[index].Equipped.SID]; !ok {
			continue
		}
		item := snapshot.Slots[index].Equipped
		item.DurabilityMax = max(item.DurabilityMax, item.Durability)
		item.Durability = item.DurabilityMax
	}
	normalizeHeroArmorSnapshotDerived(&snapshot)
	return snapshot
}

func (r *Repository) fixtureRecycledHeroArmorSnapshot(cid int, hid int, sid int) HeroArmorSnapshot {
	snapshot := r.fixtureHeroArmorSnapshot(cid, hid)
	filtered := make([]HeroArmorItem, 0, len(snapshot.Inventory))
	for _, item := range snapshot.Inventory {
		if item.SID != sid {
			filtered = append(filtered, item)
		}
	}
	snapshot.Inventory = filtered
	normalizeHeroArmorSnapshotDerived(&snapshot)
	return snapshot
}
