package legacy

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type unionMembership struct {
	UnionID int
	Pos     int
}

type unionPermissionHints struct {
	CanCreate bool
	CanApply  bool
}

func (r *Repository) MyUnion(ctx context.Context, uid int) (UnionSnapshot, error) {
	snapshot := UnionSnapshot{
		Members:   []UnionMember{},
		Relations: []UnionRelation{},
		Events:    []UnionEvent{},
		Directory: []UnionDirectoryCard{},
	}
	if r.db == nil {
		return snapshot, nil
	}

	membership, err := r.userUnionMembership(ctx, uid)
	if err != nil {
		return snapshot, err
	}

	application, err := r.unionApplication(ctx, uid)
	if err != nil {
		return snapshot, err
	}

	directory, err := r.unionDirectory(ctx, uid)
	if err != nil {
		return snapshot, err
	}

	hints, err := r.unionPermissionHints(ctx, uid, membership)
	if err != nil {
		return snapshot, err
	}

	snapshot.Application = application
	snapshot.Directory = directory
	snapshot.Permissions = unionPermissions(membership, application != nil, len(directory) > 0, hints)

	if membership.UnionID <= 0 {
		return snapshot, nil
	}

	summary, found, err := r.unionSummary(ctx, membership.UnionID, membership.Pos)
	if err != nil {
		return snapshot, err
	}
	if !found {
		return snapshot, nil
	}

	members, err := r.unionMembers(ctx, membership.UnionID)
	if err != nil {
		return snapshot, err
	}

	relations, err := r.unionRelations(ctx, membership.UnionID)
	if err != nil {
		return snapshot, err
	}

	events, err := r.unionEvents(ctx, membership.UnionID)
	if err != nil {
		return snapshot, err
	}

	snapshot.Joined = true
	snapshot.Summary = &summary
	snapshot.Members = members
	snapshot.Relations = relations
	snapshot.Events = events
	return snapshot, nil
}

func (r *Repository) userUnionMembership(ctx context.Context, uid int) (unionMembership, error) {
	membership := unionMembership{}
	err := r.db.QueryRowContext(ctx, `
select
	coalesce(union_id, 0),
	coalesce(union_pos, 0)
from sys_user
where uid = ?`, uid).Scan(&membership.UnionID, &membership.Pos)
	if err != nil {
		return unionMembership{}, err
	}
	return membership, nil
}

func (r *Repository) unionApplication(ctx context.Context, uid int) (*UnionApplication, error) {
	application := UnionApplication{}
	var createdUnix int64
	err := r.db.QueryRowContext(ctx, `
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

func (r *Repository) unionDirectory(ctx context.Context, uid int) ([]UnionDirectoryCard, error) {
	rows, err := r.db.QueryContext(ctx, `
select
	n.id,
	coalesce(n.name, ''),
	coalesce(n.leader, 0),
	coalesce(leader_user.name, ''),
	coalesce(n.member, 0),
	1 + (
		select count(*)
		from sys_union other_union
		where other_union.prestige > n.prestige
	),
	coalesce(n.prestige, 0),
	coalesce(n.intro, ''),
	case when apply.uid is null then 0 else 1 end
from sys_union n
left join sys_user leader_user on leader_user.uid = n.leader
left join sys_union_apply apply on apply.unionid = n.id and apply.uid = ?
order by
	case when n.prestige > 0 then 0 else 1 end,
	n.prestige desc,
	n.member desc,
	n.id asc
limit 18`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]UnionDirectoryCard, 0, 18)
	for rows.Next() {
		item := UnionDirectoryCard{}
		var applied int
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.LeaderUID,
			&item.LeaderName,
			&item.MemberCount,
			&item.Rank,
			&item.Prestige,
			&item.Intro,
			&applied,
		); err != nil {
			return nil, err
		}
		item.Name = strings.TrimSpace(item.Name)
		item.LeaderName = strings.TrimSpace(item.LeaderName)
		item.Intro = strings.TrimSpace(item.Intro)
		item.IsApplied = applied > 0
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) unionSummary(ctx context.Context, unionID int, myPosition int) (UnionSummary, bool, error) {
	var summary UnionSummary
	var intro sql.NullString
	var announcement sql.NullString
	var leaderName sql.NullString
	var creatorName sql.NullString
	err := r.db.QueryRowContext(ctx, `
select
	n.id,
	n.name,
	n.leader,
	coalesce(leader_user.name, ''),
	coalesce(creator_user.name, ''),
	coalesce(n.member, 0),
	1 + (
		select count(*)
		from sys_union other_union
		where other_union.prestige > n.prestige
	),
	coalesce(n.prestige, 0),
	coalesce(n.intro, ''),
	coalesce(n.announcement, '')
from sys_union n
left join sys_user leader_user on leader_user.uid = n.leader
left join sys_user creator_user on creator_user.uid = n.creator
where n.id = ?`, unionID).Scan(
		&summary.ID,
		&summary.Name,
		&summary.LeaderUID,
		&leaderName,
		&creatorName,
		&summary.MemberCount,
		&summary.Rank,
		&summary.Prestige,
		&intro,
		&announcement,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return UnionSummary{}, false, nil
		}
		return UnionSummary{}, false, err
	}

	summary.LeaderName = strings.TrimSpace(leaderName.String)
	summary.CreatorName = strings.TrimSpace(creatorName.String)
	summary.Intro = strings.TrimSpace(intro.String)
	summary.Announcement = strings.TrimSpace(announcement.String)
	summary.MyPosition = myPosition
	summary.MyPositionLabel = unionPositionLabel(myPosition)

	if err := r.db.QueryRowContext(ctx, `
select count(*)
from sys_city c
join sys_user u on u.uid = c.uid
where u.union_id = ?`, unionID).Scan(&summary.CityCount); err != nil {
		return UnionSummary{}, false, err
	}

	return summary, true, nil
}

func (r *Repository) unionMembers(ctx context.Context, unionID int) ([]UnionMember, error) {
	rows, err := r.db.QueryContext(ctx, `
select
	u.uid,
	case when trim(coalesce(u.name, '')) = '' then concat('UID ', u.uid) else u.name end as display_name,
	coalesce(u.passport, ''),
	coalesce(u.passtype, ''),
	coalesce(u.union_pos, 0),
	coalesce(rank_user.`+"`rank`"+`, 0),
	coalesce(u.nobility, 0),
	count(c.cid) as city_count,
	coalesce(max(case when c.cid = u.lastcid then c.cid end), min(c.cid), 0) as default_cid,
	coalesce(max(case when c.cid = u.lastcid then c.name end), min(c.name), '') as default_city,
	coalesce(sys_online.lastupdate, 0)
from sys_user u
left join sys_city c on c.uid = u.uid
left join rank_user on rank_user.uid = u.uid
left join sys_online on sys_online.uid = u.uid
where u.union_id = ?
group by u.uid, u.name, u.passport, u.passtype, u.union_pos, rank_user.`+"`rank`"+`, u.nobility, u.lastcid, sys_online.lastupdate
order by
	case u.union_pos
		when 1 then 0
		when 2 then 1
		when 3 then 2
		when 4 then 3
		else 4
	end,
	city_count desc,
	u.uid asc`, unionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]UnionMember, 0, 32)
	for rows.Next() {
		item := UnionMember{}
		var lastOnlineUnix int64
		if err := rows.Scan(
			&item.UID,
			&item.Name,
			&item.Passport,
			&item.PassType,
			&item.Position,
			&item.Rank,
			&item.Nobility,
			&item.CityCount,
			&item.DefaultCID,
			&item.DefaultCity,
			&lastOnlineUnix,
		); err != nil {
			return nil, err
		}

		item.PositionLabel = unionPositionLabel(item.Position)
		if lastOnlineUnix > 0 {
			item.LastOnline = time.Unix(lastOnlineUnix, 0).Format("2006-01-02 15:04:05")
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) unionRelations(ctx context.Context, unionID int) ([]UnionRelation, error) {
	rows, err := r.db.QueryContext(ctx, `
select
	ur.target,
	coalesce(target_union.name, ''),
	coalesce(leader_user.name, ''),
	ur.type,
	coalesce(target_union.member, 0),
	1 + (
		select count(*)
		from sys_union other_union
		where other_union.prestige > target_union.prestige
	),
	coalesce(target_union.prestige, 0)
from sys_union_relation ur
left join sys_union target_union on target_union.id = ur.target
left join sys_user leader_user on leader_user.uid = target_union.leader
where ur.unionid = ? and ur.type in (0, 1, 2)
order by ur.type asc, target_union.prestige desc, ur.target asc`, unionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]UnionRelation, 0, 12)
	for rows.Next() {
		item := UnionRelation{}
		if err := rows.Scan(
			&item.UnionID,
			&item.Name,
			&item.LeaderName,
			&item.RelationType,
			&item.MemberCount,
			&item.Rank,
			&item.Prestige,
		); err != nil {
			return nil, err
		}

		item.RelationLabel = unionRelationLabel(item.RelationType)
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) unionEvents(ctx context.Context, unionID int) ([]UnionEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
select
	type,
	content,
	evttime
from sys_union_event
where unionid = ?
order by evttime desc
limit 8`, unionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]UnionEvent, 0, 8)
	for rows.Next() {
		item := UnionEvent{}
		var createdUnix int64
		if err := rows.Scan(&item.Type, &item.Content, &createdUnix); err != nil {
			return nil, err
		}
		if createdUnix > 0 {
			item.CreatedAt = time.Unix(createdUnix, 0).Format("2006-01-02 15:04:05")
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) unionPermissionHints(ctx context.Context, uid int, membership unionMembership) (unionPermissionHints, error) {
	hints := unionPermissionHints{}
	if membership.UnionID > 0 {
		return hints, nil
	}

	var hongluLevel sql.NullInt64
	if err := r.db.QueryRowContext(ctx, `
select max(b.level)
from sys_building b
join sys_city c on c.cid = b.cid
where b.bid = ? and c.uid = ?`, unionBuildingHonglu, uid).Scan(&hongluLevel); err != nil {
		return hints, err
	}
	level := int64(0)
	if hongluLevel.Valid {
		level = hongluLevel.Int64
	}
	hints.CanApply = level > 0
	if level < 2 {
		return hints, nil
	}

	var lastCID int
	if err := r.db.QueryRowContext(ctx, "select coalesce(lastcid, 0) from sys_user where uid = ?", uid).Scan(&lastCID); err != nil {
		return hints, err
	}
	if lastCID <= 0 {
		return hints, nil
	}

	var gold sql.NullInt64
	if err := r.db.QueryRowContext(ctx, "select coalesce(gold, 0) from mem_city_resource where cid = ?", lastCID).Scan(&gold); err != nil {
		if err == sql.ErrNoRows {
			return hints, nil
		}
		return hints, err
	}

	if gold.Valid && gold.Int64 >= unionCreateGoldCost {
		hints.CanCreate = true
	}
	return hints, nil
}

func unionPermissions(membership unionMembership, hasApplication bool, hasDirectory bool, hints unionPermissionHints) UnionPermissions {
	permissions := UnionPermissions{
		CanCreate:      membership.UnionID <= 0 && hints.CanCreate,
		CanApply:       membership.UnionID <= 0 && !hasApplication && hasDirectory && hints.CanApply,
		CanCancelApply: membership.UnionID <= 0 && hasApplication,
		CanLeave:       membership.UnionID > 0,
	}
	if membership.Pos > 0 && membership.Pos <= 2 {
		permissions.CanEditProfile = true
		permissions.CanManageRelations = true
	}
	return permissions
}

func unionPositionLabel(value int) string {
	switch value {
	case 1:
		return "盟主"
	case 2:
		return "副盟主"
	case 3:
		return "长老"
	case 4:
		return "官员"
	default:
		return "成员"
	}
}

func unionRelationLabel(value int) string {
	switch value {
	case 0:
		return "友好"
	case 1:
		return "中立"
	case 2:
		return "敌对"
	default:
		return "未设定"
	}
}
