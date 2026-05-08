package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	unionBuildingHonglu          = 12
	unionCreateGoldCost    int64 = 10000
	unionMemberRateByHonglu      = 10
	unionEventAdd                = 0
	unionEventQuit               = 1
	unionEventChangeName         = 4
	unionRelationChangeCooldown  = 900
	unionGoalJoin                = 67
	unionNameMaxRunes            = 8
	unionIntroMaxRunes           = 200
	unionAnnouncementMaxRunes    = 500
)

type unionUserRecord struct {
	Name    string
	UnionID int
	Pos     int
	LastCID int
}

type unionRecord struct {
	ID       int
	Name     string
	LeaderID int
	Member   int
}

type unionRelationRecord struct {
	Type int
	Time int64
}

func (r *Repository) CreateUnion(ctx context.Context, uid int, name string) (UnionSnapshot, error) {
	cleanName := strings.TrimSpace(name)
	if err := validateUnionName(cleanName); err != nil {
		return UnionSnapshot{}, err
	}
	if r.db == nil {
		return UnionSnapshot{}, newInvalidError("当前环境不支持联盟操作。")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return UnionSnapshot{}, err
	}
	defer tx.Rollback()

	user, err := r.unionUserRecordTx(ctx, tx, uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return UnionSnapshot{}, newInvalidError("未找到当前主公。")
		}
		return UnionSnapshot{}, err
	}
	if user.UnionID > 0 {
		return UnionSnapshot{}, newInvalidError("你已经加入其他联盟。")
	}
	if err := r.ensureUnionNameAvailableTx(ctx, tx, cleanName, 0); err != nil {
		return UnionSnapshot{}, err
	}

	hongluLevel, err := r.userHongluLevelTx(ctx, tx, uid)
	if err != nil {
		return UnionSnapshot{}, err
	}
	if hongluLevel < 2 {
		return UnionSnapshot{}, newInvalidError("你的鸿胪寺等级不足2级，不能创建联盟。")
	}
	if user.LastCID <= 0 {
		return UnionSnapshot{}, newInvalidError("当前没有可用于创建联盟的主城。")
	}

	result, err := tx.ExecContext(ctx, `
update mem_city_resource
set gold = gold - ?
where cid = ? and gold >= ?`, unionCreateGoldCost, user.LastCID, unionCreateGoldCost)
	if err != nil {
		return UnionSnapshot{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return UnionSnapshot{}, err
	}
	if affected == 0 {
		return UnionSnapshot{}, newInvalidError("当前主城黄金不足10000，不能创建联盟。")
	}

	now := time.Now().Unix()
	result, err = tx.ExecContext(ctx, `
insert into sys_union (name, leader, creator, member, createtime)
values (?, ?, ?, 1, ?)`, cleanName, uid, uid, now)
	if err != nil {
		return UnionSnapshot{}, err
	}
	unionID64, err := result.LastInsertId()
	if err != nil {
		return UnionSnapshot{}, err
	}
	unionID := int(unionID64)

	if _, err := tx.ExecContext(ctx, `
update sys_user
set union_id = ?, union_pos = 1
where uid = ?`, unionID, uid); err != nil {
		return UnionSnapshot{}, err
	}
	if _, err := tx.ExecContext(ctx, "delete from sys_union_apply where uid = ?", uid); err != nil {
		return UnionSnapshot{}, err
	}
	if err := r.completeUnionGoalTx(ctx, tx, uid); err != nil {
		return UnionSnapshot{}, err
	}
	if err := r.refreshUnionStatsTx(ctx, tx, unionID); err != nil {
		return UnionSnapshot{}, err
	}
	if err := r.notifyUnionChangeTx(ctx, tx, uid, unionID, 1, unionDisplayName(user.Name, uid)); err != nil {
		return UnionSnapshot{}, err
	}
	if err := r.addUnionEventTx(ctx, tx, unionID, unionEventAdd, fmt.Sprintf("%s 创建联盟 %s ！", unionDisplayName(user.Name, uid), cleanName), now); err != nil {
		return UnionSnapshot{}, err
	}

	if err := tx.Commit(); err != nil {
		return UnionSnapshot{}, err
	}
	return r.MyUnion(ctx, uid)
}

func (r *Repository) ApplyJoinUnion(ctx context.Context, uid int, unionID int) (UnionSnapshot, error) {
	if unionID <= 0 {
		return UnionSnapshot{}, newInvalidError("联盟编号无效。")
	}
	if r.db == nil {
		return UnionSnapshot{}, newInvalidError("当前环境不支持联盟操作。")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return UnionSnapshot{}, err
	}
	defer tx.Rollback()

	user, err := r.unionUserRecordTx(ctx, tx, uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return UnionSnapshot{}, newInvalidError("未找到当前主公。")
		}
		return UnionSnapshot{}, err
	}
	if user.UnionID > 0 {
		return UnionSnapshot{}, newInvalidError("你已经加入其他联盟。")
	}

	hongluLevel, err := r.userHongluLevelTx(ctx, tx, uid)
	if err != nil {
		return UnionSnapshot{}, err
	}
	if hongluLevel <= 0 {
		return UnionSnapshot{}, newInvalidError("你尚未建造鸿胪寺，不能申请加入联盟。")
	}

	application, err := r.unionApplicationTx(ctx, tx, uid)
	if err != nil {
		return UnionSnapshot{}, err
	}
	if application != nil {
		return UnionSnapshot{}, newInvalidError(fmt.Sprintf("你已经申请加入［%s］,去鸿胪寺撤消原申请之后才能重新申请。", application.UnionName))
	}

	targetUnion, err := r.unionRecordTx(ctx, tx, unionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return UnionSnapshot{}, newInvalidError("目标联盟不存在。")
		}
		return UnionSnapshot{}, err
	}

	if _, err := tx.ExecContext(ctx, `
insert into sys_union_apply (uid, unionid, name, time)
values (?, ?, ?, ?)`, uid, unionID, targetUnion.Name, time.Now().Unix()); err != nil {
		return UnionSnapshot{}, err
	}

	if err := tx.Commit(); err != nil {
		return UnionSnapshot{}, err
	}
	return r.MyUnion(ctx, uid)
}

func (r *Repository) CancelJoinUnionApply(ctx context.Context, uid int) (UnionSnapshot, error) {
	if r.db == nil {
		return UnionSnapshot{}, newInvalidError("当前环境不支持联盟操作。")
	}

	if _, err := r.db.ExecContext(ctx, "delete from sys_union_apply where uid = ?", uid); err != nil {
		return UnionSnapshot{}, err
	}
	return r.MyUnion(ctx, uid)
}

func (r *Repository) LeaveUnion(ctx context.Context, uid int) (UnionSnapshot, error) {
	if r.db == nil {
		return UnionSnapshot{}, newInvalidError("当前环境不支持联盟操作。")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return UnionSnapshot{}, err
	}
	defer tx.Rollback()

	user, err := r.unionUserRecordTx(ctx, tx, uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return UnionSnapshot{}, newInvalidError("未找到当前主公。")
		}
		return UnionSnapshot{}, err
	}
	if user.UnionID <= 0 {
		return UnionSnapshot{}, newInvalidError("你当前未加入联盟。")
	}

	unionInfo, err := r.unionRecordTx(ctx, tx, user.UnionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return UnionSnapshot{}, newInvalidError("当前联盟不存在。")
		}
		return UnionSnapshot{}, err
	}

	now := time.Now().Unix()
	if unionInfo.LeaderID == uid {
		if unionInfo.Member > 1 {
			return UnionSnapshot{}, newInvalidError("盟主在联盟仍有成员时不能直接退出。")
		}

		if _, err := tx.ExecContext(ctx, "update sys_user set union_id = 0, union_pos = 0 where uid = ?", uid); err != nil {
			return UnionSnapshot{}, err
		}
		if _, err := tx.ExecContext(ctx, "delete from sys_union where id = ?", unionInfo.ID); err != nil {
			return UnionSnapshot{}, err
		}
		if _, err := tx.ExecContext(ctx, "delete from sys_union_relation where unionid = ? or target = ?", unionInfo.ID, unionInfo.ID); err != nil {
			return UnionSnapshot{}, err
		}
		if _, err := tx.ExecContext(ctx, "delete from sys_union_event where unionid = ?", unionInfo.ID); err != nil {
			return UnionSnapshot{}, err
		}
		if _, err := tx.ExecContext(ctx, "delete from mem_union_event where unionid = ?", unionInfo.ID); err != nil {
			return UnionSnapshot{}, err
		}
		if _, err := tx.ExecContext(ctx, "delete from sys_union_invite where unionid = ?", unionInfo.ID); err != nil {
			return UnionSnapshot{}, err
		}
		if _, err := tx.ExecContext(ctx, "delete from sys_union_apply where unionid = ? or uid = ?", unionInfo.ID, uid); err != nil {
			return UnionSnapshot{}, err
		}
		if _, err := tx.ExecContext(ctx, "delete from huangjin_task_log_union where unionid = ?", unionInfo.ID); err != nil {
			return UnionSnapshot{}, err
		}
		if _, err := tx.ExecContext(ctx, "delete from sys_union_city where unionid = ?", unionInfo.ID); err != nil {
			return UnionSnapshot{}, err
		}
		if _, err := tx.ExecContext(ctx, "delete from rank_union where uid = ?", unionInfo.ID); err != nil {
			return UnionSnapshot{}, err
		}
		if err := r.notifyUnionChangeTx(ctx, tx, uid, unionInfo.ID, 0, unionDisplayName(user.Name, uid)); err != nil {
			return UnionSnapshot{}, err
		}
	} else {
		if user.Pos != 0 {
			return UnionSnapshot{}, newInvalidError("官员及以上职位不能直接退出联盟。")
		}
		if _, err := tx.ExecContext(ctx, "update sys_user set union_id = 0, union_pos = 0 where uid = ?", uid); err != nil {
			return UnionSnapshot{}, err
		}
		if _, err := tx.ExecContext(ctx, "delete from sys_union_apply where uid = ?", uid); err != nil {
			return UnionSnapshot{}, err
		}
		if err := r.refreshUnionStatsTx(ctx, tx, unionInfo.ID); err != nil {
			return UnionSnapshot{}, err
		}
		if err := r.notifyUnionChangeTx(ctx, tx, uid, unionInfo.ID, 0, unionDisplayName(user.Name, uid)); err != nil {
			return UnionSnapshot{}, err
		}
		if err := r.addUnionEventTx(ctx, tx, unionInfo.ID, unionEventQuit, fmt.Sprintf("%s 退出了联盟！", unionDisplayName(user.Name, uid)), now); err != nil {
			return UnionSnapshot{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return UnionSnapshot{}, err
	}
	return r.MyUnion(ctx, uid)
}

func (r *Repository) UpdateUnionProfile(ctx context.Context, uid int, name string, intro string, announcement string) (UnionSnapshot, error) {
	cleanName := strings.TrimSpace(name)
	cleanIntro := strings.TrimSpace(intro)
	cleanAnnouncement := strings.TrimSpace(announcement)
	if err := validateUnionName(cleanName); err != nil {
		return UnionSnapshot{}, err
	}
	if utf8.RuneCountInString(cleanIntro) > unionIntroMaxRunes {
		return UnionSnapshot{}, newInvalidError("联盟简介不能超过200字。")
	}
	if utf8.RuneCountInString(cleanAnnouncement) > unionAnnouncementMaxRunes {
		return UnionSnapshot{}, newInvalidError("联盟公告不能超过500字。")
	}
	if r.db == nil {
		return UnionSnapshot{}, newInvalidError("当前环境不支持联盟操作。")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return UnionSnapshot{}, err
	}
	defer tx.Rollback()

	user, err := r.unionUserRecordTx(ctx, tx, uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return UnionSnapshot{}, newInvalidError("未找到当前主公。")
		}
		return UnionSnapshot{}, err
	}
	if user.UnionID <= 0 || user.Pos <= 0 || user.Pos > 2 {
		return UnionSnapshot{}, newInvalidError("只有盟主或副盟主可以修改联盟信息。")
	}
	if err := r.ensureUnionNameAvailableTx(ctx, tx, cleanName, user.UnionID); err != nil {
		return UnionSnapshot{}, err
	}

	unionInfo, err := r.unionRecordTx(ctx, tx, user.UnionID)
	if err != nil {
		return UnionSnapshot{}, err
	}

	if _, err := tx.ExecContext(ctx, `
update sys_union
set name = ?, intro = ?, announcement = ?
where id = ?`, cleanName, cleanIntro, cleanAnnouncement, user.UnionID); err != nil {
		return UnionSnapshot{}, err
	}
	if unionInfo.Name != cleanName {
		if err := r.addUnionEventTx(ctx, tx, user.UnionID, unionEventChangeName, fmt.Sprintf("%s 将联盟名称改为 %s !", unionDisplayName(user.Name, uid), cleanName), time.Now().Unix()); err != nil {
			return UnionSnapshot{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, "update rank_union set name = ? where uid = ?", cleanName, user.UnionID); err != nil {
		return UnionSnapshot{}, err
	}

	if err := tx.Commit(); err != nil {
		return UnionSnapshot{}, err
	}
	return r.MyUnion(ctx, uid)
}

func (r *Repository) SetUnionRelation(ctx context.Context, uid int, targetUnionID int, relationType int) (UnionSnapshot, error) {
	if relationType < 0 || relationType > 2 {
		return UnionSnapshot{}, newInvalidError("外交关系类型无效。")
	}
	if targetUnionID <= 0 {
		return UnionSnapshot{}, newInvalidError("目标联盟编号无效。")
	}
	if r.db == nil {
		return UnionSnapshot{}, newInvalidError("当前环境不支持联盟操作。")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return UnionSnapshot{}, err
	}
	defer tx.Rollback()

	user, err := r.unionUserRecordTx(ctx, tx, uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return UnionSnapshot{}, newInvalidError("未找到当前主公。")
		}
		return UnionSnapshot{}, err
	}
	if user.UnionID <= 0 || user.Pos <= 0 || user.Pos > 2 {
		return UnionSnapshot{}, newInvalidError("只有盟主或副盟主可以调整外交关系。")
	}

	unionInfo, err := r.unionRecordTx(ctx, tx, user.UnionID)
	if err != nil {
		return UnionSnapshot{}, err
	}
	if unionInfo.ID == targetUnionID {
		return UnionSnapshot{}, newInvalidError("不能和自己的联盟建立外交关系。")
	}

	targetUnion, err := r.unionRecordTx(ctx, tx, targetUnionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return UnionSnapshot{}, newInvalidError("目标联盟不存在。")
		}
		return UnionSnapshot{}, err
	}

	now := time.Now().Unix()
	oldRelation, err := r.unionRelationRecordTx(ctx, tx, unionInfo.ID, targetUnion.ID)
	if err != nil {
		return UnionSnapshot{}, err
	}
	if oldRelation != nil && oldRelation.Type == relationType {
		return r.MyUnion(ctx, uid)
	}
	if oldRelation != nil && oldRelation.Time > now-unionRelationChangeCooldown {
		return UnionSnapshot{}, newInvalidError("外交状态调整过于频繁，请稍后再试。")
	}

	if _, err := tx.ExecContext(ctx, `
insert into sys_union_relation (type, unionid, target, time)
values (?, ?, ?, ?)
on duplicate key update type = values(type), time = values(time)`, relationType, unionInfo.ID, targetUnion.ID, now); err != nil {
		return UnionSnapshot{}, err
	}

	relationLabel := unionRelationLabel(relationType)
	if err := r.addUnionEventTx(ctx, tx, unionInfo.ID, unionEventAdd+5+relationType, fmt.Sprintf("%s 将与 %s 的关系设为%s。", unionDisplayName(user.Name, uid), targetUnion.Name, relationLabel), now); err != nil {
		return UnionSnapshot{}, err
	}
	if err := r.addUnionEventTx(ctx, tx, targetUnion.ID, unionEventAdd+5+relationType, fmt.Sprintf("%s 将与你方的关系设为%s。", unionInfo.Name, relationLabel), now); err != nil {
		return UnionSnapshot{}, err
	}

	if err := tx.Commit(); err != nil {
		return UnionSnapshot{}, err
	}
	return r.MyUnion(ctx, uid)
}

func (r *Repository) RemoveUnionRelation(ctx context.Context, uid int, targetUnionID int, relationType int) (UnionSnapshot, error) {
	if targetUnionID <= 0 {
		return UnionSnapshot{}, newInvalidError("目标联盟编号无效。")
	}
	if relationType < 0 || relationType > 2 {
		return UnionSnapshot{}, newInvalidError("外交关系类型无效。")
	}
	if r.db == nil {
		return UnionSnapshot{}, newInvalidError("当前环境不支持联盟操作。")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return UnionSnapshot{}, err
	}
	defer tx.Rollback()

	user, err := r.unionUserRecordTx(ctx, tx, uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return UnionSnapshot{}, newInvalidError("未找到当前主公。")
		}
		return UnionSnapshot{}, err
	}
	if user.UnionID <= 0 || user.Pos <= 0 || user.Pos > 2 {
		return UnionSnapshot{}, newInvalidError("只有盟主或副盟主可以调整外交关系。")
	}

	unionInfo, err := r.unionRecordTx(ctx, tx, user.UnionID)
	if err != nil {
		return UnionSnapshot{}, err
	}
	targetUnion, err := r.unionRecordTx(ctx, tx, targetUnionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return UnionSnapshot{}, newInvalidError("目标联盟不存在。")
		}
		return UnionSnapshot{}, err
	}

	result, err := tx.ExecContext(ctx, `
update sys_union_relation
set type = 3, time = ?
where unionid = ? and target = ? and type = ?`, time.Now().Unix(), unionInfo.ID, targetUnionID, relationType)
	if err != nil {
		return UnionSnapshot{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return UnionSnapshot{}, err
	}
	if affected == 0 {
		return UnionSnapshot{}, newInvalidError("未找到对应的外交关系记录。")
	}

	relationLabel := unionRelationLabel(relationType)
	now := time.Now().Unix()
	if err := r.addUnionEventTx(ctx, tx, unionInfo.ID, unionEventAdd+5+relationType, fmt.Sprintf("%s 取消了与 %s 的%s关系。", unionDisplayName(user.Name, uid), targetUnion.Name, relationLabel), now); err != nil {
		return UnionSnapshot{}, err
	}
	if err := r.addUnionEventTx(ctx, tx, targetUnion.ID, unionEventAdd+5+relationType, fmt.Sprintf("%s 取消了与你方的%s关系。", unionInfo.Name, relationLabel), now); err != nil {
		return UnionSnapshot{}, err
	}

	if err := tx.Commit(); err != nil {
		return UnionSnapshot{}, err
	}
	return r.MyUnion(ctx, uid)
}

func validateUnionName(name string) error {
	if name == "" {
		return newInvalidError("联盟名称不能为空。")
	}
	if utf8.RuneCountInString(name) > unionNameMaxRunes {
		return newInvalidError("联盟名称不能超过8字。")
	}
	if strings.ContainsAny(name, "'\\") {
		return newInvalidError("联盟名称包含非法字符。")
	}
	return nil
}

func (r *Repository) ensureUnionNameAvailableTx(ctx context.Context, tx *sql.Tx, name string, currentUnionID int) error {
	var bannedCount int
	if err := tx.QueryRowContext(ctx, `
select count(*)
from cfg_baned_name
where instr(?, name) > 0`, name).Scan(&bannedCount); err != nil {
		return err
	}
	if bannedCount > 0 {
		return newInvalidError("联盟名称包含非法字符。")
	}

	var existingID int
	err := tx.QueryRowContext(ctx, `
select id
from sys_union
where name = ?`, name).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil && existingID != currentUnionID {
		return newInvalidError("该联盟名称已被使用。")
	}
	return nil
}

func (r *Repository) unionUserRecordTx(ctx context.Context, tx *sql.Tx, uid int) (unionUserRecord, error) {
	record := unionUserRecord{}
	err := tx.QueryRowContext(ctx, `
select
	coalesce(name, ''),
	coalesce(union_id, 0),
	coalesce(union_pos, 0),
	coalesce(lastcid, 0)
from sys_user
where uid = ?`, uid).Scan(&record.Name, &record.UnionID, &record.Pos, &record.LastCID)
	if err != nil {
		return unionUserRecord{}, err
	}
	record.Name = strings.TrimSpace(record.Name)
	return record, nil
}

func (r *Repository) unionRecordTx(ctx context.Context, tx *sql.Tx, unionID int) (unionRecord, error) {
	record := unionRecord{}
	err := tx.QueryRowContext(ctx, `
select
	id,
	coalesce(name, ''),
	coalesce(leader, 0),
	coalesce(member, 0)
from sys_union
where id = ?`, unionID).Scan(&record.ID, &record.Name, &record.LeaderID, &record.Member)
	if err != nil {
		return unionRecord{}, err
	}
	record.Name = strings.TrimSpace(record.Name)
	return record, nil
}

func (r *Repository) findUnionByNameTx(ctx context.Context, tx *sql.Tx, name string) (unionRecord, error) {
	record := unionRecord{}
	err := tx.QueryRowContext(ctx, `
select
	id,
	coalesce(name, ''),
	coalesce(leader, 0),
	coalesce(member, 0)
from sys_union
where name = ?`, name).Scan(&record.ID, &record.Name, &record.LeaderID, &record.Member)
	if err != nil {
		return unionRecord{}, err
	}
	record.Name = strings.TrimSpace(record.Name)
	return record, nil
}

func (r *Repository) unionApplicationTx(ctx context.Context, tx *sql.Tx, uid int) (*UnionApplication, error) {
	application := UnionApplication{}
	var createdUnix int64
	err := tx.QueryRowContext(ctx, `
select
	unionid,
	coalesce(name, ''),
	coalesce(time, 0)
from sys_union_apply
where uid = ?`, uid).Scan(&application.UnionID, &application.UnionName, &createdUnix)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if createdUnix > 0 {
		application.CreatedAt = time.Unix(createdUnix, 0).Format("2006-01-02 15:04:05")
	}
	return &application, nil
}

func (r *Repository) userHongluLevelTx(ctx context.Context, tx *sql.Tx, uid int) (int, error) {
	var level sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
select max(b.level)
from sys_building b
join sys_city c on c.cid = b.cid
where b.bid = ? and c.uid = ?`, unionBuildingHonglu, uid).Scan(&level); err != nil {
		return 0, err
	}
	if !level.Valid {
		return 0, nil
	}
	return int(level.Int64), nil
}

func (r *Repository) unionRelationRecordTx(ctx context.Context, tx *sql.Tx, unionID int, targetUnionID int) (*unionRelationRecord, error) {
	record := unionRelationRecord{}
	err := tx.QueryRowContext(ctx, `
select
	type,
	coalesce(time, 0)
from sys_union_relation
where unionid = ? and target = ?`, unionID, targetUnionID).Scan(&record.Type, &record.Time)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func (r *Repository) refreshUnionStatsTx(ctx context.Context, tx *sql.Tx, unionID int) error {
	if _, err := tx.ExecContext(ctx, `
update sys_union
set
	member = (select count(*) from sys_user where union_id = ?),
	prestige = coalesce((select sum(prestige) from sys_user where union_id = ?), 0)
where id = ?`, unionID, unionID, unionID); err != nil {
		return err
	}

	var famousCityCount int
	if err := tx.QueryRowContext(ctx, `
select count(*)
from sys_city c
join sys_user u on u.uid = c.uid
where u.union_id = ? and c.type > 0`, unionID).Scan(&famousCityCount); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
insert into sys_union_city (unionid, count)
values (?, ?)
on duplicate key update count = values(count)`, unionID, famousCityCount); err != nil {
		return err
	}
	return nil
}

func (r *Repository) completeUnionGoalTx(ctx context.Context, tx *sql.Tx, uid int) error {
	_, err := tx.ExecContext(ctx, `
replace into sys_user_goal (uid, gid)
values (?, ?)`, uid, unionGoalJoin)
	return err
}

func (r *Repository) notifyUnionChangeTx(ctx context.Context, tx *sql.Tx, uid int, unionID int, state int, nickname string) error {
	_, err := tx.ExecContext(ctx, `
insert into mem_union_buf (uid, nick, union_id, state, updatetime)
values (?, ?, ?, ?, ?)`, uid, nickname, unionID, state, time.Now().Unix())
	return err
}

func (r *Repository) addUnionEventTx(ctx context.Context, tx *sql.Tx, unionID int, eventType int, content string, now int64) error {
	if _, err := tx.ExecContext(ctx, `
insert into sys_union_event (unionid, type, content, evttime)
values (?, ?, ?, ?)`, unionID, eventType, content, now); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
insert into mem_union_event (unionid, type, content, evttime)
values (?, ?, ?, ?)`, unionID, eventType, content, now)
	return err
}

func unionDisplayName(name string, uid int) string {
	if strings.TrimSpace(name) == "" {
		return fmt.Sprintf("UID %d", uid)
	}
	return strings.TrimSpace(name)
}
