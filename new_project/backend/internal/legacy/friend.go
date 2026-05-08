package legacy

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

const maxUserRelationCount = 20

const (
	friendRelationTypeFriend = 0
	friendRelationTypeEnemy  = 1
)

const (
	friendInvalidTypeMessage   = "关系类型不存在。"
	friendUserNotFoundMessage  = "你输入的玩家不存在，请重新输入。"
	friendRelationFullMessage  = "好友或仇人列表已满，最多只能保存20人。"
	friendRelationSelfMessage  = "不能将自己加入关系名单。"
	friendRelationEmptyMessage = "请输入主公名字。"
)

func (r *Repository) UserRelations(ctx context.Context, uid int) (RelationPage, error) {
	if r.db == nil {
		return RelationPage{
			Limit: maxUserRelationCount,
			Items: []RelationCard{},
		}, nil
	}

	rows, err := r.db.QueryContext(ctx, `
select
	r.tuid,
	coalesce(nullif(trim(coalesce(u.name, '')), ''), concat('UID ', r.tuid)) as display_name,
	coalesce(u.passport, ''),
	coalesce(u.passtype, ''),
	r.type,
	coalesce(n.name, ''),
	coalesce(u.nobility, 0),
	coalesce(city.city_count, 0),
	coalesce(city.default_cid, 0),
	coalesce(city.default_city, ''),
	r.time
from sys_user_relation r
left join sys_user u on u.uid = r.tuid
left join sys_union n on n.id = u.union_id
left join (
	select
		c.uid,
		count(c.cid) as city_count,
		coalesce(max(case when c.cid = su.lastcid then c.cid end), min(c.cid)) as default_cid,
		coalesce(max(case when c.cid = su.lastcid then c.name end), min(c.name)) as default_city
	from sys_city c
	join sys_user su on su.uid = c.uid
	group by c.uid, su.lastcid
) city on city.uid = r.tuid
where r.uid = ?
order by r.type asc, city.city_count desc, r.time desc, r.tuid asc`, uid)
	if err != nil {
		return RelationPage{}, err
	}
	defer rows.Close()

	page := RelationPage{
		Limit: maxUserRelationCount,
		Items: make([]RelationCard, 0, maxUserRelationCount),
	}

	for rows.Next() {
		item := RelationCard{}
		var updatedUnix int64
		if err := rows.Scan(
			&item.UID,
			&item.Name,
			&item.Passport,
			&item.PassType,
			&item.RelationType,
			&item.UnionName,
			&item.Nobility,
			&item.CityCount,
			&item.DefaultCID,
			&item.DefaultCity,
			&updatedUnix,
		); err != nil {
			return RelationPage{}, err
		}

		item.RelationLabel = relationTypeLabel(item.RelationType)
		if updatedUnix > 0 {
			item.UpdatedAt = time.Unix(updatedUnix, 0).Format("2006-01-02 15:04:05")
		}

		switch item.RelationType {
		case friendRelationTypeFriend:
			page.FriendCount++
		case friendRelationTypeEnemy:
			page.EnemyCount++
		}

		page.Items = append(page.Items, item)
	}
	if err := rows.Err(); err != nil {
		return RelationPage{}, err
	}

	page.Total = len(page.Items)
	return page, nil
}

func (r *Repository) AddUserRelation(ctx context.Context, uid int, targetName string, relationType int) (RelationPage, error) {
	if r.db == nil {
		return RelationPage{}, ErrDatabaseUnavailable
	}

	normalizedName := strings.TrimSpace(targetName)
	if normalizedName == "" {
		return RelationPage{}, newInvalidError(friendRelationEmptyMessage)
	}

	if _, err := normalizeRelationType(relationType); err != nil {
		return RelationPage{}, err
	}

	recipient, err := r.mailRecipient(ctx, normalizedName)
	if err != nil {
		if err == sql.ErrNoRows {
			return RelationPage{}, newInvalidError(friendUserNotFoundMessage)
		}
		return RelationPage{}, err
	}

	if recipient.UID == uid {
		return RelationPage{}, newInvalidError(friendRelationSelfMessage)
	}

	exists, err := r.userRelationExists(ctx, uid, recipient.UID)
	if err != nil {
		return RelationPage{}, err
	}

	if !exists {
		count, err := r.userRelationCount(ctx, uid)
		if err != nil {
			return RelationPage{}, err
		}
		if count >= maxUserRelationCount {
			return RelationPage{}, newInvalidError(friendRelationFullMessage)
		}
	}

	now := time.Now().Unix()
	if exists {
		if _, err := r.db.ExecContext(ctx, `
update sys_user_relation
set type = ?, time = ?
where uid = ? and tuid = ?`, relationType, now, uid, recipient.UID); err != nil {
			return RelationPage{}, err
		}
	} else {
		if _, err := r.db.ExecContext(ctx, `
insert into sys_user_relation (uid, tuid, type, time)
values (?, ?, ?, ?)`, uid, recipient.UID, relationType, now); err != nil {
			return RelationPage{}, err
		}
	}

	return r.UserRelations(ctx, uid)
}

func (r *Repository) RemoveUserRelation(ctx context.Context, uid int, targetUID int, relationType int) (RelationPage, error) {
	if r.db == nil {
		return RelationPage{}, ErrDatabaseUnavailable
	}

	if targetUID <= 0 {
		return RelationPage{}, newInvalidError(friendUserNotFoundMessage)
	}

	if _, err := normalizeRelationType(relationType); err != nil {
		return RelationPage{}, err
	}

	if _, err := r.db.ExecContext(ctx, `
delete from sys_user_relation
where uid = ? and tuid = ? and type = ?`, uid, targetUID, relationType); err != nil {
		return RelationPage{}, err
	}

	return r.UserRelations(ctx, uid)
}

func (r *Repository) userRelationCount(ctx context.Context, uid int) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `
select count(*)
from sys_user_relation
where uid = ?`, uid).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) userRelationExists(ctx context.Context, uid int, targetUID int) (bool, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `
select count(*)
from sys_user_relation
where uid = ? and tuid = ?`, uid, targetUID).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func normalizeRelationType(value int) (int, error) {
	switch value {
	case friendRelationTypeFriend, friendRelationTypeEnemy:
		return value, nil
	default:
		return 0, newInvalidError(friendInvalidTypeMessage)
	}
}

func relationTypeLabel(value int) string {
	if value == friendRelationTypeEnemy {
		return "仇人"
	}
	return "好友"
}
