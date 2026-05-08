package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (r *Repository) MyTasks(ctx context.Context, uid int) (TaskSnapshot, error) {
	snapshot := TaskSnapshot{
		Categories: []TaskCategory{},
	}
	if r.db == nil {
		return snapshot, nil
	}

	groupRows, err := r.db.QueryContext(ctx, `
select
	g.id,
	g.type,
	g.name,
	g.description
from cfg_task_group g
where g.id in (
	select distinct t.`+"`group`"+`
	from sys_user_task u
	join cfg_task t on t.id = u.tid
	where u.uid = ? and u.state = 0
)
order by g.type asc, g.id asc`, uid)
	if err != nil {
		return snapshot, err
	}
	defer groupRows.Close()

	groupMap := make(map[int]*TaskGroup)
	groupIDsByType := make(map[int][]int)
	categoryOrder := make([]int, 0, 8)
	for groupRows.Next() {
		group := &TaskGroup{
			Tasks: []TaskCard{},
		}
		if err := groupRows.Scan(&group.ID, &group.Type, &group.Name, &group.Description); err != nil {
			return TaskSnapshot{}, err
		}
		group.TypeLabel = taskGroupTypeLabel(group.Type)
		groupMap[group.ID] = group
		if _, exists := groupIDsByType[group.Type]; !exists {
			categoryOrder = append(categoryOrder, group.Type)
		}
		groupIDsByType[group.Type] = append(groupIDsByType[group.Type], group.ID)
	}
	if err := groupRows.Err(); err != nil {
		return TaskSnapshot{}, err
	}

	taskRows, err := r.db.QueryContext(ctx, `
select
	t.id,
	t.`+"`group`"+`,
	t.pretid,
	t.name,
	coalesce(t.content, ''),
	coalesce(t.todo, '')
from sys_user_task u
join cfg_task t on t.id = u.tid
where u.uid = ? and u.state = 0
order by t.`+"`group`"+` asc, t.id asc`, uid)
	if err != nil {
		return snapshot, err
	}
	defer taskRows.Close()

	tasksByGroup := make(map[int][]*TaskCard)
	taskByID := make(map[int]*TaskCard)
	taskIDs := make([]int, 0, 128)
	for taskRows.Next() {
		task := &TaskCard{
			Goals:   []TaskGoal{},
			Rewards: []TaskReward{},
		}
		if err := taskRows.Scan(
			&task.ID,
			&task.GroupID,
			&task.PreTaskID,
			&task.Name,
			&task.Content,
			&task.Todo,
		); err != nil {
			return TaskSnapshot{}, err
		}

		tasksByGroup[task.GroupID] = append(tasksByGroup[task.GroupID], task)
		taskByID[task.ID] = task
		taskIDs = append(taskIDs, task.ID)

		if _, exists := groupMap[task.GroupID]; !exists {
			group := &TaskGroup{
				ID:          task.GroupID,
				Type:        -1,
				TypeLabel:   taskGroupTypeLabel(-1),
				Name:        fmt.Sprintf("任务组 %d", task.GroupID),
				Description: "",
				Tasks:       []TaskCard{},
			}
			groupMap[task.GroupID] = group
			if _, ok := groupIDsByType[group.Type]; !ok {
				categoryOrder = append(categoryOrder, group.Type)
			}
			groupIDsByType[group.Type] = append(groupIDsByType[group.Type], task.GroupID)
		}
	}
	if err := taskRows.Err(); err != nil {
		return TaskSnapshot{}, err
	}

	if len(taskIDs) == 0 {
		return snapshot, nil
	}

	eval := newTaskEvalContext(r, uid)
	goalsByTask := make(map[int][]TaskGoal, len(taskIDs))
	goalQuery, goalArgs := taskIDInQuery(taskIDs)
	goalRows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
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
where g.tid in (%s)
order by g.tid asc, g.id asc`, goalQuery), append([]any{uid}, goalArgs...)...)
	if err != nil {
		return TaskSnapshot{}, err
	}
	defer goalRows.Close()

	for goalRows.Next() {
		var (
			goal     TaskGoal
			reduce   int
			joinedID sql.NullInt64
		)
		if err := goalRows.Scan(
			&goal.ID,
			&goal.TaskID,
			&goal.Sort,
			&goal.Type,
			&goal.Count,
			&reduce,
			&goal.Content,
			&joinedID,
		); err != nil {
			return TaskSnapshot{}, err
		}

		goal.Reduce = reduce != 0
		completed, current, target, trackable, err := eval.goalStatus(ctx, goal.TaskID, joinedID.Valid && joinedID.Int64 > 0, goal)
		if err != nil {
			return TaskSnapshot{}, err
		}
		goal.Completed = completed
		goal.Current = current
		goal.Target = target
		goal.Trackable = trackable
		goal.StatusLabel = taskGoalStatusLabel(goal)

		goalsByTask[goal.TaskID] = append(goalsByTask[goal.TaskID], goal)
	}
	if err := goalRows.Err(); err != nil {
		return TaskSnapshot{}, err
	}

	rewardsByTask := make(map[int][]TaskReward, len(taskIDs))
	rewardRows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
select
	tid,
	sort,
	type,
	count,
	coalesce(name, '')
from cfg_task_reward
where tid in (%s)
order by tid asc, type asc, sort asc`, goalQuery), goalArgs...)
	if err != nil {
		return TaskSnapshot{}, err
	}
	defer rewardRows.Close()

	for rewardRows.Next() {
		var (
			taskID int
			reward TaskReward
		)
		if err := rewardRows.Scan(&taskID, &reward.Sort, &reward.Type, &reward.Count, &reward.Name); err != nil {
			return TaskSnapshot{}, err
		}
		rewardsByTask[taskID] = append(rewardsByTask[taskID], reward)
	}
	if err := rewardRows.Err(); err != nil {
		return TaskSnapshot{}, err
	}

	for _, task := range taskByID {
		task.Goals = goalsByTask[task.ID]
		task.Rewards = rewardsByTask[task.ID]
		task.GoalCount = len(task.Goals)
		task.CompletedGoals = 0
		for _, goal := range task.Goals {
			if goal.Completed {
				task.CompletedGoals++
			}
		}
		task.Completed = task.GoalCount > 0 && task.CompletedGoals == task.GoalCount
	}

	for _, groupType := range categoryOrder {
		category := TaskCategory{
			Type:   groupType,
			Label:  taskGroupTypeLabel(groupType),
			Groups: []TaskGroup{},
		}

		for _, groupID := range groupIDsByType[groupType] {
			source := groupMap[groupID]
			if source == nil {
				continue
			}

			group := *source
			group.Tasks = make([]TaskCard, 0, len(tasksByGroup[groupID]))
			for _, task := range tasksByGroup[groupID] {
				group.Tasks = append(group.Tasks, *task)
				group.Total++
				category.TaskCount++
				snapshot.Summary.TaskCount++
				if task.Completed {
					group.Completed++
					category.Completed++
					snapshot.Summary.CompletedTasks++
				}
			}
			category.Groups = append(category.Groups, group)
			category.GroupCount++
			snapshot.Summary.GroupCount++
			if group.Total > 0 && group.Completed == group.Total {
				snapshot.Summary.CompletedGroups++
			}
		}

		if category.GroupCount > 0 {
			snapshot.Categories = append(snapshot.Categories, category)
			snapshot.Summary.CategoryCount++
		}
	}

	snapshot.Summary.PendingTasks = snapshot.Summary.TaskCount - snapshot.Summary.CompletedTasks
	return snapshot, nil
}

type taskEvalContext struct {
	repo *Repository
	uid  int

	userLoaded bool
	user       taskEvalUser

	cityLoaded bool
	city       taskEvalCity

	unionMembersLoaded bool
	unionMembers       int64

	unionNamedCitiesLoaded bool
	unionNamedCities       int64

	namedCitiesLoaded bool
	namedCities       int64

	productionLoaded bool
	productionByBid  map[int]int64

	maxBuildingLoaded bool
	maxBuildingLevel  map[int]int64

	technicLoaded bool
	technicLevels map[int]int64

	cityTypeLoaded bool
	cityTypeCounts map[int]int64

	goodsCounts    map[int]int64
	soldierCounts  map[int]int64
	defenceCounts  map[int]int64
	thingCounts    map[int]int64
	heroCounts     map[int]int64
	actEventCounts map[int]int64
}

type taskEvalUser struct {
	LastCID   int
	UnionID   int
	OfficePos int
	Nobility  int64
	Money     int64
	Prestige  int64
}

type taskEvalCity struct {
	Gold      int64
	Food      int64
	Wood      int64
	Rock      int64
	Iron      int64
	People    int64
	PeopleMax int64
	Morale    int64
	Complaint int64
	Tax       int64
}

func newTaskEvalContext(repo *Repository, uid int) *taskEvalContext {
	return &taskEvalContext{
		repo:             repo,
		uid:              uid,
		productionByBid:  map[int]int64{},
		maxBuildingLevel: map[int]int64{},
		technicLevels:    map[int]int64{},
		cityTypeCounts:   map[int]int64{},
		goodsCounts:      map[int]int64{},
		soldierCounts:    map[int]int64{},
		defenceCounts:    map[int]int64{},
		thingCounts:      map[int]int64{},
		heroCounts:       map[int]int64{},
		actEventCounts:   map[int]int64{},
	}
}

func (c *taskEvalContext) goalStatus(ctx context.Context, taskID int, joined bool, goal TaskGoal) (bool, int64, int64, bool, error) {
	if joined {
		return true, 1, 1, false, nil
	}

	if isManualOnlyTask(taskID) {
		return false, 0, 1, false, nil
	}

	switch goal.Sort {
	case 1:
		return c.resourceGoalStatus(ctx, taskID, goal)
	case 2:
		current, err := c.goodCount(ctx, goal.Type)
		return current >= goal.Count, current, goal.Count, true, err
	case 3:
		current, err := c.soldierCount(ctx, goal.Type)
		return current >= goal.Count, current, goal.Count, true, err
	case 4:
		current, err := c.defenceCount(ctx, goal.Type)
		return current >= goal.Count, current, goal.Count, true, err
	case 5:
		current, err := c.thingCount(ctx, goal.Type)
		return current >= goal.Count, current, goal.Count, true, err
	case 6:
		current, err := c.maxBuildingLevelForUser(ctx, goal.Type)
		return current >= goal.Count, current, goal.Count, true, err
	case 7:
		current, err := c.technicLevel(ctx, goal.Type)
		target := goal.Count + 1
		return current >= target, current, target, true, err
	case 8:
		current, err := c.cityTypeCount(ctx, goal.Type)
		return current >= goal.Count, current, goal.Count, true, err
	case 9:
		current, err := c.heroCount(ctx, goal.Type)
		return current >= goal.Count, current, goal.Count, true, err
	case 101:
		current, err := c.actEventCount(ctx, goal.Type)
		return current >= goal.Count, current, goal.Count, true, err
	default:
		return false, 0, goal.Count, false, nil
	}
}

func (c *taskEvalContext) resourceGoalStatus(ctx context.Context, taskID int, goal TaskGoal) (bool, int64, int64, bool, error) {
	user, err := c.userInfo(ctx)
	if err != nil {
		return false, 0, goal.Count, false, err
	}

	switch goal.Type {
	case 9:
		return user.Prestige >= goal.Count, user.Prestige, goal.Count, true, nil
	case 11:
		if user.UnionID <= 0 {
			return false, 0, goal.Count, true, nil
		}
		current, err := c.unionMemberCount(ctx)
		return current >= goal.Count, current, goal.Count, true, err
	case 17:
		current := int64(user.OfficePos)
		return current >= goal.Count, current, goal.Count, true, nil
	case 18:
		return user.Nobility >= goal.Count, user.Nobility, goal.Count, true, nil
	case 19, 22, 122:
		return user.Money >= goal.Count, user.Money, goal.Count, true, nil
	case 20:
		current, err := c.namedCityCount(ctx)
		return current >= goal.Count, current, goal.Count, true, err
	case 21:
		if user.UnionID <= 0 {
			return false, 0, goal.Count, true, nil
		}
		current, err := c.unionNamedCityCount(ctx)
		return current >= goal.Count, current, goal.Count, true, err
	case 12:
		current, err := c.baseProduction(ctx, foodBuildingID)
		return current >= goal.Count, current, goal.Count, true, err
	case 13:
		current, err := c.baseProduction(ctx, woodBuildingID)
		return current >= goal.Count, current, goal.Count, true, err
	case 14:
		current, err := c.baseProduction(ctx, rockBuildingID)
		return current >= goal.Count, current, goal.Count, true, err
	case 15:
		current, err := c.baseProduction(ctx, ironBuildingID)
		return current >= goal.Count, current, goal.Count, true, err
	case 31, 32, 33, 34, 35, 36, 37, 38, 39:
		current, target, completed, err := c.battleGoalProgress(ctx, taskID, goal.Type)
		return completed, current, target, true, err
	}

	city, err := c.cityInfo(ctx)
	if err != nil {
		return false, 0, goal.Count, false, err
	}

	var current int64
	switch goal.Type {
	case 1:
		current = city.Gold
	case 2:
		current = city.Food
	case 3:
		current = city.Wood
	case 4:
		current = city.Rock
	case 5:
		current = city.Iron
	case 6:
		current = city.People
	case 7:
		current = city.Morale
	case 8:
		current = city.Complaint
	case 10:
		current = city.PeopleMax
	case 16:
		current = city.People * city.Tax / 100
	default:
		return false, 0, goal.Count, false, nil
	}

	return current >= goal.Count, current, goal.Count, true, nil
}

func (c *taskEvalContext) userInfo(ctx context.Context) (taskEvalUser, error) {
	if c.userLoaded {
		return c.user, nil
	}

	user := taskEvalUser{}
	err := c.repo.db.QueryRowContext(ctx, `
select
	coalesce(lastcid, 0),
	coalesce(union_id, 0),
	coalesce(officepos, 0),
	coalesce(nobility, 0),
	coalesce(money, 0),
	coalesce(prestige, 0)
from sys_user
where uid = ?`, c.uid).Scan(
		&user.LastCID,
		&user.UnionID,
		&user.OfficePos,
		&user.Nobility,
		&user.Money,
		&user.Prestige,
	)
	if err != nil {
		return taskEvalUser{}, err
	}

	c.userLoaded = true
	c.user = user
	return c.user, nil
}

func (c *taskEvalContext) cityInfo(ctx context.Context) (taskEvalCity, error) {
	if c.cityLoaded {
		return c.city, nil
	}

	user, err := c.userInfo(ctx)
	if err != nil {
		return taskEvalCity{}, err
	}
	if user.LastCID <= 0 {
		c.cityLoaded = true
		return c.city, nil
	}

	err = c.repo.db.QueryRowContext(ctx, `
select
	coalesce(gold, 0),
	coalesce(food, 0),
	coalesce(wood, 0),
	coalesce(rock, 0),
	coalesce(iron, 0),
	coalesce(people, 0),
	coalesce(people_max, 0),
	coalesce(morale, 0),
	coalesce(complaint, 0),
	coalesce(tax, 0)
from mem_city_resource
where cid = ?`, user.LastCID).Scan(
		&c.city.Gold,
		&c.city.Food,
		&c.city.Wood,
		&c.city.Rock,
		&c.city.Iron,
		&c.city.People,
		&c.city.PeopleMax,
		&c.city.Morale,
		&c.city.Complaint,
		&c.city.Tax,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			return taskEvalCity{}, err
		}
	}

	c.cityLoaded = true
	return c.city, nil
}

func (c *taskEvalContext) unionMemberCount(ctx context.Context) (int64, error) {
	if c.unionMembersLoaded {
		return c.unionMembers, nil
	}

	user, err := c.userInfo(ctx)
	if err != nil {
		return 0, err
	}
	if user.UnionID <= 0 {
		c.unionMembersLoaded = true
		return 0, nil
	}

	if err := c.repo.db.QueryRowContext(ctx, `
select coalesce(member, 0)
from sys_union
where id = ?`, user.UnionID).Scan(&c.unionMembers); err != nil {
		if err != sql.ErrNoRows {
			return 0, err
		}
	}

	c.unionMembersLoaded = true
	return c.unionMembers, nil
}

func (c *taskEvalContext) unionNamedCityCount(ctx context.Context) (int64, error) {
	if c.unionNamedCitiesLoaded {
		return c.unionNamedCities, nil
	}

	user, err := c.userInfo(ctx)
	if err != nil {
		return 0, err
	}
	if user.UnionID <= 0 {
		c.unionNamedCitiesLoaded = true
		return 0, nil
	}

	if err := c.repo.db.QueryRowContext(ctx, `
select count(*)
from sys_city
where type > 0 and uid in (
	select uid from sys_user where union_id = ?
)`, user.UnionID).Scan(&c.unionNamedCities); err != nil {
		return 0, err
	}

	c.unionNamedCitiesLoaded = true
	return c.unionNamedCities, nil
}

func (c *taskEvalContext) namedCityCount(ctx context.Context) (int64, error) {
	if c.namedCitiesLoaded {
		return c.namedCities, nil
	}

	if err := c.repo.db.QueryRowContext(ctx, `
select count(*)
from sys_city
where uid = ? and type > 0`, c.uid).Scan(&c.namedCities); err != nil {
		return 0, err
	}

	c.namedCitiesLoaded = true
	return c.namedCities, nil
}

func (c *taskEvalContext) baseProduction(ctx context.Context, bid int) (int64, error) {
	if !c.productionLoaded {
		user, err := c.userInfo(ctx)
		if err != nil {
			return 0, err
		}
		if user.LastCID > 0 {
			rows, err := c.repo.db.QueryContext(ctx, `
select
	bid,
	coalesce(sum(level * (level + 1) * 50), 0)
from sys_building
where cid = ? and bid in (?, ?, ?, ?)
group by bid`, user.LastCID, foodBuildingID, woodBuildingID, rockBuildingID, ironBuildingID)
			if err != nil {
				return 0, err
			}
			defer rows.Close()

			for rows.Next() {
				var currentBid int
				var currentValue int64
				if err := rows.Scan(&currentBid, &currentValue); err != nil {
					return 0, err
				}
				c.productionByBid[currentBid] = currentValue
			}
			if err := rows.Err(); err != nil {
				return 0, err
			}
		}
		c.productionLoaded = true
	}

	return c.productionByBid[bid], nil
}

func (c *taskEvalContext) maxBuildingLevelForUser(ctx context.Context, bid int) (int64, error) {
	if !c.maxBuildingLoaded {
		rows, err := c.repo.db.QueryContext(ctx, `
select
	b.bid,
	coalesce(max(b.level), 0)
from sys_building b
join sys_city city on city.cid = b.cid
where city.uid = ?
group by b.bid`, c.uid)
		if err != nil {
			return 0, err
		}
		defer rows.Close()

		for rows.Next() {
			var currentBid int
			var maxLevel int64
			if err := rows.Scan(&currentBid, &maxLevel); err != nil {
				return 0, err
			}
			c.maxBuildingLevel[currentBid] = maxLevel
		}
		if err := rows.Err(); err != nil {
			return 0, err
		}

		c.maxBuildingLoaded = true
	}

	return c.maxBuildingLevel[bid], nil
}

func (c *taskEvalContext) technicLevel(ctx context.Context, tid int) (int64, error) {
	if !c.technicLoaded {
		rows, err := c.repo.db.QueryContext(ctx, `
select tid, level
from sys_technic
where uid = ?`, c.uid)
		if err != nil {
			return 0, err
		}
		defer rows.Close()

		for rows.Next() {
			var technicID int
			var level int64
			if err := rows.Scan(&technicID, &level); err != nil {
				return 0, err
			}
			c.technicLevels[technicID] = level
		}
		if err := rows.Err(); err != nil {
			return 0, err
		}

		c.technicLoaded = true
	}

	return c.technicLevels[tid], nil
}

func (c *taskEvalContext) cityTypeCount(ctx context.Context, cityType int) (int64, error) {
	if !c.cityTypeLoaded {
		rows, err := c.repo.db.QueryContext(ctx, `
select type, count(*)
from sys_city
where uid = ?
group by type`, c.uid)
		if err != nil {
			return 0, err
		}
		defer rows.Close()

		for rows.Next() {
			var currentType int
			var count int64
			if err := rows.Scan(&currentType, &count); err != nil {
				return 0, err
			}
			c.cityTypeCounts[currentType] = count
		}
		if err := rows.Err(); err != nil {
			return 0, err
		}

		c.cityTypeLoaded = true
	}

	return c.cityTypeCounts[cityType], nil
}

func (c *taskEvalContext) goodCount(ctx context.Context, gid int) (int64, error) {
	return c.singleCount(ctx, c.goodsCounts, gid, `
select coalesce(count, 0)
from sys_goods
where uid = ? and gid = ?`)
}

func (c *taskEvalContext) soldierCount(ctx context.Context, sid int) (int64, error) {
	return c.singleCount(ctx, c.soldierCounts, sid, `
select coalesce(s.count, 0)
from sys_city_soldier s
join sys_user u on u.lastcid = s.cid
where u.uid = ? and s.sid = ?`)
}

func (c *taskEvalContext) defenceCount(ctx context.Context, did int) (int64, error) {
	return c.singleCount(ctx, c.defenceCounts, did, `
select coalesce(d.count, 0)
from sys_city_defence d
join sys_user u on u.lastcid = d.cid
where u.uid = ? and d.did = ?`)
}

func (c *taskEvalContext) thingCount(ctx context.Context, tid int) (int64, error) {
	return c.singleCount(ctx, c.thingCounts, tid, `
select coalesce(count, 0)
from sys_things
where uid = ? and tid = ?`)
}

func (c *taskEvalContext) heroCount(ctx context.Context, npcid int) (int64, error) {
	return c.singleCount(ctx, c.heroCounts, npcid, `
select count(*)
from sys_city_hero
where uid = ? and npcid = ?`)
}

func (c *taskEvalContext) actEventCount(ctx context.Context, eid int) (int64, error) {
	return c.singleCount(ctx, c.actEventCounts, eid, `
select coalesce(count, 0)
from temp_act_event
where uid = ? and eid = ?`)
}

func (c *taskEvalContext) singleCount(ctx context.Context, cache map[int]int64, key int, query string) (int64, error) {
	if value, ok := cache[key]; ok {
		return value, nil
	}

	var value int64
	if err := c.repo.db.QueryRowContext(ctx, query, c.uid, key).Scan(&value); err != nil {
		if err != sql.ErrNoRows {
			return 0, err
		}
	}

	cache[key] = value
	return value, nil
}

func (c *taskEvalContext) battleGoalProgress(ctx context.Context, taskID int, goalType int) (int64, int64, bool, error) {
	now := time.Now().Unix()
	dayStart := now - now%86400 + 1
	dayEnd := dayStart + 86400

	var claimedToday int
	if err := c.repo.db.QueryRowContext(ctx, `
select count(*)
from log_everyday_task
where uid = ? and tid = ? and gettime >= ? and gettime <= ?`, c.uid, taskID, dayStart, dayEnd).Scan(&claimedToday); err != nil {
		return 0, 0, false, err
	}
	if claimedToday > 0 {
		return 0, battleGoalTarget(goalType), false, nil
	}

	var since sql.NullInt64
	if err := c.repo.db.QueryRowContext(ctx, `
select gettime
from log_everyday_task
where uid = ? and tid = ?
order by gettime desc
limit 1`, c.uid, taskID).Scan(&since); err != nil && err != sql.ErrNoRows {
		return 0, 0, false, err
	}

	windowStart := int64(0)
	if since.Valid {
		windowStart = since.Int64
	}

	var (
		query string
		args  []any
	)
	switch goalType {
	case 31:
		query = `
select count(*)
from log_battle_honour
where uid = ? and result = 0 and quittime >= ? and quittime <= ? and (battleid = 1001 or battleid = 2001)`
		args = []any{c.uid, windowStart, now}
	case 32:
		query = `
select count(*)
from log_battle_honour
where uid = ? and result = 0 and level = 10 and battleid = 1001 and quittime >= ? and quittime <= ?`
		args = []any{c.uid, windowStart, now}
	case 33:
		query = `
select count(*)
from log_battle_honour
where uid = ? and result = 0 and battleid = 1001 and quittime >= ? and quittime <= ?`
		args = []any{c.uid, windowStart, now}
	case 34:
		query = `
select count(*)
from log_battle_honour
where uid = ? and result = 0 and unionid = 3 and level = 10 and battleid = 2001 and quittime >= ? and quittime <= ?`
		args = []any{c.uid, windowStart, now}
	case 35:
		query = `
select count(*)
from log_battle_honour
where uid = ? and result = 0 and unionid = 3 and battleid = 2001 and quittime >= ? and quittime <= ?`
		args = []any{c.uid, windowStart, now}
	case 36:
		query = `
select count(*)
from log_battle_honour
where uid = ? and result = 0 and unionid = 4 and level = 10 and battleid = 2001 and quittime >= ? and quittime <= ?`
		args = []any{c.uid, windowStart, now}
	case 37:
		query = `
select count(*)
from log_battle_honour
where uid = ? and result = 0 and unionid = 4 and battleid = 2001 and quittime >= ? and quittime <= ?`
		args = []any{c.uid, windowStart, now}
	case 38:
		query = `
select count(*)
from log_battle_honour
where uid = ? and battleid = 1001 and result = 0`
		args = []any{c.uid}
	case 39:
		query = `
select count(*)
from log_battle_honour
where uid = ? and battleid = 2001 and result = 0`
		args = []any{c.uid}
	default:
		return 0, 0, false, nil
	}

	var count int64
	if err := c.repo.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, 0, false, err
	}

	target := battleGoalTarget(goalType)
	if count > target && (goalType == 38 || goalType == 39) {
		count = target
	}
	return count, target, count >= target, nil
}

func battleGoalTarget(goalType int) int64 {
	switch goalType {
	case 31, 33, 35, 37:
		return 10
	case 32, 34, 36, 38, 39:
		return 1
	default:
		return 0
	}
}

func isManualOnlyTask(taskID int) bool {
	return (taskID >= 60000 && taskID <= 60024) || (taskID >= 60100 && taskID <= 60144)
}

func taskGroupTypeLabel(taskType int) string {
	switch taskType {
	case 0:
		return "主线任务"
	case 1:
		return "政务任务"
	case 2:
		return "委托任务"
	case 3:
		return "战场任务"
	case 6:
		return "活动任务"
	default:
		return "其他任务"
	}
}

func taskGoalStatusLabel(goal TaskGoal) string {
	if goal.Trackable {
		if goal.Target > 0 {
			return formatTaskProgress(goal.Current) + " / " + formatTaskProgress(goal.Target)
		}
		return formatTaskProgress(goal.Current)
	}
	if goal.Completed {
		return "已达成"
	}
	return "未达成"
}

func formatTaskProgress(value int64) string {
	return strconv.FormatInt(value, 10)
}

func taskIDInQuery(taskIDs []int) (string, []any) {
	placeholders := strings.TrimRight(strings.Repeat("?,", len(taskIDs)), ",")
	args := make([]any, 0, len(taskIDs))
	for _, id := range taskIDs {
		args = append(args, id)
	}
	return placeholders, args
}
