package legacy

import (
	"context"
	"database/sql"
	"time"
)

type claimTaskInfo struct {
	ID int
}

func (r *Repository) ClaimTaskReward(ctx context.Context, uid int, taskID int) (TaskSnapshot, error) {
	if taskID <= 0 {
		return TaskSnapshot{}, newInvalidError("任务编号无效。")
	}
	if taskID >= 200000 {
		return TaskSnapshot{}, newInvalidError("当前任务窗暂不支持悬赏任务领奖。")
	}
	if isUnsupportedEpicTask(taskID) {
		return TaskSnapshot{}, newInvalidError("当前阶段暂未接入黄巾史诗任务领奖。")
	}
	if r.db == nil {
		return TaskSnapshot{}, newInvalidError("当前环境不支持任务领奖。")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return TaskSnapshot{}, err
	}
	defer tx.Rollback()

	task, err := r.claimTaskInfo(ctx, tx, uid, taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			return TaskSnapshot{}, newInvalidError("未找到可领奖的任务。")
		}
		return TaskSnapshot{}, err
	}

	goals, err := r.loadClaimTaskGoals(ctx, uid, taskID)
	if err != nil {
		return TaskSnapshot{}, err
	}
	for _, goal := range goals {
		if !goal.Completed {
			return TaskSnapshot{}, newInvalidError("任务条件尚未完成。")
		}
	}

	cityID, err := r.userLastCIDTx(ctx, tx, uid)
	if err != nil {
		return TaskSnapshot{}, err
	}

	rewards, err := r.loadClaimTaskRewards(ctx, tx, taskID)
	if err != nil {
		return TaskSnapshot{}, err
	}
	if len(rewards) == 0 {
		rewards, err = r.loadSpecialTaskRewards(ctx, tx, uid, taskID)
		if err != nil {
			return TaskSnapshot{}, err
		}
	}

	skipFollowups := false
	if isHuangjinDonationTask(task.ID) {
		skipFollowups, err = r.applyHuangjinDonationTaskTx(ctx, tx, uid, cityID, task.ID, rewards)
		if err != nil {
			return TaskSnapshot{}, err
		}
	} else {
		for _, reward := range rewards {
			if err := r.applyTaskRewardTx(ctx, tx, uid, cityID, reward); err != nil {
				return TaskSnapshot{}, err
			}
		}
	}

	if skipFollowups {
		if err := tx.Commit(); err != nil {
			return TaskSnapshot{}, err
		}
		return r.MyTasks(ctx, uid)
	}

	if !taskCanRecomplete(task.ID) {
		if _, err := tx.ExecContext(ctx, `
update sys_user_task
set state = 1
where uid = ? and tid = ? and state = 0`, uid, task.ID); err != nil {
			return TaskSnapshot{}, err
		}
	}

	for _, goal := range goals {
		if !goal.Reduce {
			continue
		}
		if err := r.applyTaskGoalReductionTx(ctx, tx, uid, cityID, goal); err != nil {
			return TaskSnapshot{}, err
		}
	}

	if isDailyBattleTask(task.ID) {
		if _, err := tx.ExecContext(ctx, `
insert into log_everyday_task (uid, tid, gettime)
values (?, ?, ?)
on duplicate key update gettime = values(gettime)`, uid, task.ID, time.Now().Unix()); err != nil {
			return TaskSnapshot{}, err
		}
	}

	if err := r.activateTaskTriggersTx(ctx, tx, uid, task.ID); err != nil {
		return TaskSnapshot{}, err
	}

	if err := tx.Commit(); err != nil {
		return TaskSnapshot{}, err
	}

	return r.MyTasks(ctx, uid)
}

func (r *Repository) claimTaskInfo(ctx context.Context, tx *sql.Tx, uid int, taskID int) (claimTaskInfo, error) {
	task := claimTaskInfo{}
	err := tx.QueryRowContext(ctx, `
select t.id
from cfg_task t
join sys_user_task u on u.tid = t.id
where u.uid = ? and u.tid = ? and u.state = 0`, uid, taskID).Scan(&task.ID)
	if err != nil {
		return claimTaskInfo{}, err
	}
	return task, nil
}

func (r *Repository) loadClaimTaskGoals(ctx context.Context, uid int, taskID int) ([]TaskGoal, error) {
	eval := newTaskEvalContext(r, uid)
	rows, err := r.db.QueryContext(ctx, `
select
	g.id,
	g.tid,
	g.sort,
	g.type,
	g.count,
	g.reduce,
	coalesce(g.content, ''),
	ug.uid
from cfg_task_goal g
left join sys_user_goal ug on ug.gid = g.id and ug.uid = ?
where g.tid = ?
order by g.id asc`, uid, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	goals := []TaskGoal{}
	for rows.Next() {
		var (
			goal     TaskGoal
			reduce   int
			joinedID sql.NullInt64
		)
		if err := rows.Scan(
			&goal.ID,
			&goal.TaskID,
			&goal.Sort,
			&goal.Type,
			&goal.Count,
			&reduce,
			&goal.Content,
			&joinedID,
		); err != nil {
			return nil, err
		}

		goal.Reduce = reduce != 0
		completed, current, target, trackable, err := eval.goalStatus(ctx, taskID, joinedID.Valid && joinedID.Int64 > 0, goal)
		if err != nil {
			return nil, err
		}
		goal.Completed = completed
		goal.Current = current
		goal.Target = target
		goal.Trackable = trackable
		goals = append(goals, goal)
	}

	return goals, rows.Err()
}

func (r *Repository) loadClaimTaskRewards(ctx context.Context, tx *sql.Tx, taskID int) ([]TaskReward, error) {
	rows, err := tx.QueryContext(ctx, `
select sort, type, count, coalesce(name, '')
from cfg_task_reward
where tid = ?
order by type asc, sort asc`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rewards := []TaskReward{}
	for rows.Next() {
		reward := TaskReward{}
		if err := rows.Scan(&reward.Sort, &reward.Type, &reward.Count, &reward.Name); err != nil {
			return nil, err
		}
		rewards = append(rewards, reward)
	}

	return rewards, rows.Err()
}

func (r *Repository) loadSpecialTaskRewards(ctx context.Context, tx *sql.Tx, uid int, taskID int) ([]TaskReward, error) {
	switch taskID {
	case 250:
		var salary int64
		if err := tx.QueryRowContext(ctx, `
select coalesce(p.salary, 0)
from sys_user u
left join cfg_office_pos p on p.id = u.officepos
where u.uid = ?`, uid).Scan(&salary); err != nil {
			return nil, err
		}
		return []TaskReward{{Sort: 1, Type: 1, Count: salary, Name: "黄金"}}, nil
	case 251:
		var salary int64
		if err := tx.QueryRowContext(ctx, `
select coalesce(n.salary, 0)
from sys_user u
left join cfg_nobility n on n.id = u.nobility
where u.uid = ?`, uid).Scan(&salary); err != nil {
			return nil, err
		}
		return []TaskReward{
			{Sort: 1, Type: 2, Count: salary, Name: "粮草"},
			{Sort: 1, Type: 3, Count: salary, Name: "木材"},
			{Sort: 1, Type: 4, Count: salary, Name: "石料"},
			{Sort: 1, Type: 5, Count: salary, Name: "铁锭"},
		}, nil
	case 279:
		gold, err := r.unionFamousCityGoldTx(ctx, tx, uid)
		if err != nil {
			return nil, err
		}
		return []TaskReward{{Sort: 1, Type: 1, Count: gold, Name: "黄金"}}, nil
	default:
		return nil, nil
	}
}

func (r *Repository) unionFamousCityGoldTx(ctx context.Context, tx *sql.Tx, uid int) (int64, error) {
	var unionID int
	if err := tx.QueryRowContext(ctx, "select coalesce(union_id, 0) from sys_user where uid = ?", uid).Scan(&unionID); err != nil {
		return 0, err
	}
	if unionID == 0 {
		return 0, nil
	}

	rows, err := tx.QueryContext(ctx, `
select c.type, count(*)
from sys_city c
join sys_user u on u.uid = c.uid
where u.union_id = ? and c.type > 0
group by c.type`, unionID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var gold int64
	for rows.Next() {
		var cityType int
		var count int64
		if err := rows.Scan(&cityType, &count); err != nil {
			return 0, err
		}

		switch cityType {
		case 1:
			gold += 10000 * count
		case 2:
			gold += 30000 * count
		case 3:
			gold += 100000 * count
		case 4:
			gold += 300000 * count
		}
	}

	return gold, rows.Err()
}

func (r *Repository) userLastCIDTx(ctx context.Context, tx *sql.Tx, uid int) (int, error) {
	var cityID int
	if err := tx.QueryRowContext(ctx, "select coalesce(lastcid, 0) from sys_user where uid = ?", uid).Scan(&cityID); err != nil {
		return 0, err
	}
	return cityID, nil
}

func (r *Repository) applyTaskRewardTx(ctx context.Context, tx *sql.Tx, uid int, cityID int, reward TaskReward) error {
	switch reward.Sort {
	case 1:
		return r.applyTaskResourceTx(ctx, tx, uid, cityID, reward.Type, reward.Count)
	case 2:
		if reward.Type == 0 {
			return r.applyTaskResourceTx(ctx, tx, uid, cityID, 19, reward.Count)
		}
		return r.adjustUserGoodsTx(ctx, tx, uid, reward.Type, reward.Count)
	case 3:
		return r.adjustCitySoldiersTx(ctx, tx, cityID, reward.Type, reward.Count)
	case 4:
		return r.adjustCityDefenceTx(ctx, tx, cityID, reward.Type, reward.Count)
	case 5:
		return r.adjustUserThingsTx(ctx, tx, uid, reward.Type, reward.Count)
	default:
		return newInvalidError("当前任务奖励类型暂未支持。")
	}
}

func (r *Repository) applyTaskGoalReductionTx(ctx context.Context, tx *sql.Tx, uid int, cityID int, goal TaskGoal) error {
	switch goal.Sort {
	case 1:
		return r.applyTaskResourceTx(ctx, tx, uid, cityID, goal.Type, -goal.Count)
	case 2:
		return r.adjustUserGoodsTx(ctx, tx, uid, goal.Type, -goal.Count)
	case 3:
		return r.adjustCitySoldiersTx(ctx, tx, cityID, goal.Type, -goal.Count)
	case 4:
		return r.adjustCityDefenceTx(ctx, tx, cityID, goal.Type, -goal.Count)
	case 5:
		return r.adjustUserThingsTx(ctx, tx, uid, goal.Type, -goal.Count)
	case 6:
		return r.cutUserArmorTx(ctx, tx, uid, goal.Type, int(goal.Count))
	case 101:
		_, err := tx.ExecContext(ctx, `
insert into temp_act_event (uid, type, eid, count)
values (?, ?, ?, ?)
on duplicate key update count = count + values(count)`, uid, goal.Sort, goal.Type, -goal.Count)
		return err
	default:
		return newInvalidError("当前任务扣除条件暂未支持。")
	}
}

func (r *Repository) applyTaskResourceTx(ctx context.Context, tx *sql.Tx, uid int, cityID int, resourceType int, count int64) error {
	switch resourceType {
	case 1:
		return r.adjustCityResourceTx(ctx, tx, cityID, "gold", count)
	case 2:
		return r.adjustCityResourceTx(ctx, tx, cityID, "food", count)
	case 3:
		return r.adjustCityResourceTx(ctx, tx, cityID, "wood", count)
	case 4:
		return r.adjustCityResourceTx(ctx, tx, cityID, "rock", count)
	case 5:
		return r.adjustCityResourceTx(ctx, tx, cityID, "iron", count)
	case 6:
		return r.adjustCityResourceTx(ctx, tx, cityID, "people", count)
	case 7:
		_, err := tx.ExecContext(ctx, `
update mem_city_resource
set morale = least(100, greatest(0, morale + ?)),
	people_stable = people_max * least(100, greatest(0, morale + ?)) * 0.01
where cid = ?`, count, count, cityID)
		return err
	case 8:
		_, err := tx.ExecContext(ctx, `
update mem_city_resource
set complaint = greatest(0, complaint + ?)
where cid = ?`, count, cityID)
		return err
	case 9:
		_, err := tx.ExecContext(ctx, `
update sys_user
set prestige = prestige + ?, warprestige = warprestige + ?
where uid = ?`, count, count, uid)
		return err
	case 17:
		_, err := tx.ExecContext(ctx, "update sys_user set officepos = ? where uid = ?", count, uid)
		return err
	case 18:
		_, err := tx.ExecContext(ctx, "update sys_user set nobility = ? where uid = ?", count, uid)
		return err
	case 19:
		if err := r.ensureGiftWalletTx(ctx, tx, uid); err != nil {
			return err
		}
		return r.adjustUserGoodsTx(ctx, tx, uid, 0, count)
	case 20, 122:
		_, err := tx.ExecContext(ctx, "update sys_user set money = money + ? where uid = ?", count, uid)
		return err
	case 30:
		_, err := tx.ExecContext(ctx, "update sys_user set honour = honour + ? where uid = ?", count, uid)
		return err
	default:
		return newInvalidError("当前资源奖励类型暂未支持。")
	}
}

func (r *Repository) adjustCityResourceTx(ctx context.Context, tx *sql.Tx, cityID int, column string, delta int64) error {
	if cityID <= 0 {
		return newInvalidError("当前任务没有可结算的城池。")
	}

	query := `
update mem_city_resource
set ` + column + ` = ` + column + ` + ?
where cid = ?`
	if delta < 0 {
		query = `
update mem_city_resource
set ` + column + ` = ` + column + ` + ?
where cid = ? and ` + column + ` >= ?`
		result, err := tx.ExecContext(ctx, query, delta, cityID, -delta)
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return newInvalidError("任务扣除条件不足。")
		}
		return nil
	}

	_, err := tx.ExecContext(ctx, query, delta, cityID)
	return err
}

func (r *Repository) adjustUserGoodsTx(ctx context.Context, tx *sql.Tx, uid int, gid int, delta int64) error {
	if delta >= 0 {
		return r.addUserGoodsTx(ctx, tx, uid, gid, delta)
	}

	result, err := tx.ExecContext(ctx, `
update sys_goods
set count = count + ?
where uid = ? and gid = ? and count >= ?`, delta, uid, gid, -delta)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return newInvalidError("任务扣除道具不足。")
	}
	return nil
}

func (r *Repository) adjustUserThingsTx(ctx context.Context, tx *sql.Tx, uid int, tid int, delta int64) error {
	if delta >= 0 {
		_, err := tx.ExecContext(ctx, `
insert into sys_things (uid, tid, count)
values (?, ?, ?)
on duplicate key update count = count + values(count)`, uid, tid, delta)
		return err
	}

	result, err := tx.ExecContext(ctx, `
update sys_things
set count = count + ?
where uid = ? and tid = ? and count >= ?`, delta, uid, tid, -delta)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return newInvalidError("任务扣除勋章/道具不足。")
	}
	return nil
}

func (r *Repository) adjustCitySoldiersTx(ctx context.Context, tx *sql.Tx, cityID int, sid int, delta int64) error {
	if cityID <= 0 {
		return newInvalidError("当前任务没有可结算的城池。")
	}

	if delta >= 0 {
		if err := r.addCitySoldiersTx(ctx, tx, cityID, map[int]int64{sid: delta}); err != nil {
			return err
		}
		if err := r.ensureCityResAdd(ctx, tx, cityID); err != nil {
			return err
		}
		return r.recalculateCityProduction(ctx, tx, cityID)
	}

	result, err := tx.ExecContext(ctx, `
update sys_city_soldier
set count = count + ?
where cid = ? and sid = ? and count >= ?`, delta, cityID, sid, -delta)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return newInvalidError("任务扣除兵力不足。")
	}
	if err := r.ensureCityResAdd(ctx, tx, cityID); err != nil {
		return err
	}
	return r.recalculateCityProduction(ctx, tx, cityID)
}

func (r *Repository) adjustCityDefenceTx(ctx context.Context, tx *sql.Tx, cityID int, did int, delta int64) error {
	if cityID <= 0 {
		return newInvalidError("当前任务没有可结算的城池。")
	}

	if delta >= 0 {
		_, err := tx.ExecContext(ctx, `
insert into sys_city_defence (cid, did, count)
values (?, ?, ?)
on duplicate key update count = count + values(count)`, cityID, did, delta)
		return err
	}

	result, err := tx.ExecContext(ctx, `
update sys_city_defence
set count = count + ?
where cid = ? and did = ? and count >= ?`, delta, cityID, did, -delta)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return newInvalidError("任务扣除城防不足。")
	}
	return nil
}

func (r *Repository) cutUserArmorTx(ctx context.Context, tx *sql.Tx, uid int, armorID int, count int) error {
	if count <= 0 {
		return nil
	}
	result, err := tx.ExecContext(ctx, `
delete from sys_user_armor
where uid = ? and armorid = ? and hid = 0
limit ?`, uid, armorID, count)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if int(affected) < count {
		return newInvalidError("任务扣除装备不足。")
	}
	return nil
}

func (r *Repository) activateTaskTriggersTx(ctx context.Context, tx *sql.Tx, uid int, taskID int) error {
	rows, err := tx.QueryContext(ctx, `
select id, coalesce(`+"`default`"+`, 0)
from cfg_task
where pretid = ?`, taskID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var triggerID int
		var state int
		if err := rows.Scan(&triggerID, &state); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
insert into sys_user_task (uid, tid, state)
values (?, ?, ?)
on duplicate key update state = values(state)`, uid, triggerID, state); err != nil {
			return err
		}
	}

	return rows.Err()
}

func (r *Repository) applyHuangjinDonationTaskTx(ctx context.Context, tx *sql.Tx, uid int, cityID int, taskID int, rewards []TaskReward) (bool, error) {
	progress := struct {
		Max int64
		Cur int64
	}{}
	if err := tx.QueryRowContext(ctx, `
select maxvalue, curvalue
from huangjin_progress
where tid = ?`, taskID).Scan(&progress.Max, &progress.Cur); err != nil {
		return false, err
	}

	if progress.Cur >= progress.Max {
		if _, err := tx.ExecContext(ctx, `
update sys_user_task
set state = 1
where tid = ?`, taskID); err != nil {
			return false, err
		}
		return true, nil
	}

	unionID, err := r.userUnionIDTx(ctx, tx, uid)
	if err != nil {
		return false, err
	}

	for _, reward := range rewards {
		if reward.Sort == 5 {
			if err := r.recordHuangjinRewardTx(ctx, tx, uid, unionID, reward); err != nil {
				return false, err
			}
		}
		if err := r.applyTaskRewardTx(ctx, tx, uid, cityID, reward); err != nil {
			return false, err
		}
	}

	if _, err := tx.ExecContext(ctx, `
update huangjin_progress
set curvalue = least(maxvalue, curvalue + 1)
where tid = ?`, taskID); err != nil {
		return false, err
	}

	if progress.Cur >= progress.Max-1 {
		if _, err := tx.ExecContext(ctx, `
update sys_user_task
set state = 1
where tid = ?`, taskID); err != nil {
			return false, err
		}

		var unfinished int
		if err := tx.QueryRowContext(ctx, `
select count(*)
from huangjin_progress
where curvalue < maxvalue`).Scan(&unfinished); err != nil {
			return false, err
		}
		if unfinished == 0 {
			if _, err := tx.ExecContext(ctx, `
update mem_state
set value = 1
where state = 5`); err != nil {
				return false, err
			}
			if err := r.triggerHuangjinCityTaskTx(ctx, tx); err != nil {
				return false, err
			}
		}
	}

	return false, nil
}

func (r *Repository) userUnionIDTx(ctx context.Context, tx *sql.Tx, uid int) (int, error) {
	var unionID int
	if err := tx.QueryRowContext(ctx, `
select coalesce(union_id, 0)
from sys_user
where uid = ?`, uid).Scan(&unionID); err != nil {
		return 0, err
	}
	return unionID, nil
}

func (r *Repository) recordHuangjinRewardTx(ctx context.Context, tx *sql.Tx, uid int, unionID int, reward TaskReward) error {
	var column string
	switch reward.Type {
	case 11001:
		column = "jungong"
	case 12001:
		column = "juanxian"
	case 13001:
		column = "qinwang"
	case 14001:
		column = "gongpin"
	default:
		return nil
	}

	if _, err := tx.ExecContext(ctx, `
insert into huangjin_task_log (uid, `+column+`)
values (?, ?)
on duplicate key update `+column+` = `+column+` + values(`+column+`)`, uid, reward.Count); err != nil {
		return err
	}
	if unionID <= 0 {
		return nil
	}

	_, err := tx.ExecContext(ctx, `
insert into huangjin_task_log_union (unionid, `+column+`)
values (?, ?)
on duplicate key update `+column+` = `+column+` + values(`+column+`)`, unionID, reward.Count)
	return err
}

func (r *Repository) triggerHuangjinCityTaskTx(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
insert into sys_user_task (uid, tid, state)
select uid, 10501, 0
from sys_user_task
where tid = 243 and state = 1
on duplicate key update state = values(state)`)
	return err
}

func taskCanRecomplete(taskID int) bool {
	return (taskID > 10000 && taskID < 10600) ||
		(taskID > 11000 && taskID < 15000) ||
		(taskID > 100021 && taskID < 100025) ||
		(taskID > 100140 && taskID < 100144) ||
		taskID == 100171 ||
		taskID == 100201
}

func isUnsupportedEpicTask(taskID int) bool {
	return false
}

func isHuangjinDonationTask(taskID int) bool {
	return taskID > 11000 && taskID < 15000
}

func isDailyBattleTask(taskID int) bool {
	switch taskID {
	case 6001, 6101, 6102, 6201, 6202, 6211, 6212:
		return true
	default:
		return false
	}
}
