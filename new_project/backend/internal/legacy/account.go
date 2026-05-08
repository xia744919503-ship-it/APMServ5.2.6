package legacy

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var ErrForbidden = errors.New("forbidden")

func (r *Repository) CommanderOptions(ctx context.Context, limit int) ([]CommanderOption, error) {
	limit = clamp(limit, 1, 40)
	if r.db == nil {
		return []CommanderOption{
			{UID: 710, Name: "襄平", Passport: "710", PassType: "npc", CityCount: 1103, DefaultCID: 215265, DefaultCity: "洛阳"},
			{UID: 518, Name: "许昌", Passport: "518", PassType: "npc", CityCount: 494, DefaultCID: 225185, DefaultCity: "长安"},
			{UID: 357, Name: "赤壁", Passport: "357", PassType: "npc", CityCount: 253, DefaultCID: 165335, DefaultCity: "邺县"},
		}, nil
	}

	query := fmt.Sprintf(`
select
	u.uid,
	case when trim(coalesce(u.name, '')) = '' then concat('UID ', u.uid) else u.name end as display_name,
	u.passport,
	u.passtype,
	count(c.cid) as city_count,
	coalesce(max(case when c.cid = u.lastcid then c.cid end), min(c.cid)) as default_cid,
	coalesce(max(case when c.cid = u.lastcid then c.name end), min(c.name)) as default_city
from sys_user u
join sys_city c on c.uid = u.uid
group by u.uid, u.name, u.passport, u.passtype, u.lastcid
having count(c.cid) > 0
order by city_count desc, u.uid asc
limit %d`, limit)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]CommanderOption, 0, limit)
	for rows.Next() {
		item := CommanderOption{}
		if err := rows.Scan(
			&item.UID,
			&item.Name,
			&item.Passport,
			&item.PassType,
			&item.CityCount,
			&item.DefaultCID,
			&item.DefaultCity,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) SessionUser(ctx context.Context, uid int) (SessionUser, error) {
	if r.db == nil {
		options, _ := r.CommanderOptions(ctx, 5)
		for _, item := range options {
			if item.UID == uid {
				return SessionUser{
					UID:         item.UID,
					Name:        item.Name,
					Passport:    item.Passport,
					PassType:    item.PassType,
					CityCount:   item.CityCount,
					DefaultCID:  item.DefaultCID,
					DefaultCity: item.DefaultCity,
					Sex:         0,
					Face:        0,
					UserSex:     0,
					UserFace:    0,
					OfficePos:   "平民",
					Nobility:    "平民",
					UnionName:   "无联盟",
					UnionPos:    "",
				}, nil
			}
		}
		return SessionUser{}, sql.ErrNoRows
	}

	row := r.db.QueryRowContext(ctx, `
select
	u.uid,
	case when trim(coalesce(u.name, '')) = '' then concat('UID ', u.uid) else u.name end as display_name,
	coalesce(u.passport, '') as passport,
	coalesce(u.passtype, '') as passtype,
	count(c.cid) as city_count,
	coalesce(max(case when c.cid = u.lastcid then c.cid end), min(c.cid), 0) as default_cid,
	coalesce(max(case when c.cid = u.lastcid then c.name end), min(c.name), '') as default_city,
	coalesce(u.sex, 0) as sex,
	coalesce(u.face, 0) as face,
	cast(coalesce(u.prestige, 0) as signed) as prestige,
	cast(coalesce(rank_user.`+"`rank`"+`, 0) as signed) as user_rank,
	coalesce(u.officepos, 0) as officepos_id,
	coalesce(cfg_office_pos.name, '平民') as officepos,
	coalesce(u.nobility, 0) as nobility_id,
	coalesce(cfg_nobility.name, '平民') as nobility,
	coalesce(u.union_id, 0) as union_id,
	case when coalesce(u.union_id, 0) = 0 then '无联盟' else coalesce(sys_union.name, '') end as unionname,
	coalesce(u.union_pos, 0) as union_pos_id
from sys_user u
left join sys_city c on c.uid = u.uid
left join rank_user on rank_user.uid = u.uid
left join cfg_office_pos on cfg_office_pos.id = u.officepos
left join cfg_nobility on cfg_nobility.id = u.nobility
left join sys_union on sys_union.id = u.union_id
where u.uid = ?
group by u.uid, u.name, u.passport, u.passtype, u.lastcid, u.sex, u.face, u.prestige, rank_user.`+"`rank`"+`, u.officepos, cfg_office_pos.name, u.nobility, cfg_nobility.name, u.union_id, sys_union.name, u.union_pos`, uid)

	user := SessionUser{}
	err := row.Scan(
		&user.UID,
		&user.Name,
		&user.Passport,
		&user.PassType,
		&user.CityCount,
		&user.DefaultCID,
		&user.DefaultCity,
		&user.Sex,
		&user.Face,
		&user.Prestige,
		&user.Rank,
		&user.OfficePosID,
		&user.OfficePos,
		&user.NobilityID,
		&user.Nobility,
		&user.UnionID,
		&user.UnionName,
		&user.UnionPosID,
	)
	if err != nil {
		return SessionUser{}, err
	}
	normalizeSessionUser(&user)

	return user, nil
}

func (r *Repository) SessionUserByPassport(ctx context.Context, passport string) (SessionUser, error) {
	passport = strings.TrimSpace(passport)
	if passport == "" {
		return SessionUser{}, sql.ErrNoRows
	}

	if r.db == nil {
		options, _ := r.CommanderOptions(ctx, 18)
		for _, item := range options {
			if strings.EqualFold(item.Passport, passport) {
				return SessionUser{
					UID:         item.UID,
					Name:        item.Name,
					Passport:    item.Passport,
					PassType:    item.PassType,
					CityCount:   item.CityCount,
					DefaultCID:  item.DefaultCID,
					DefaultCity: item.DefaultCity,
					Sex:         0,
					Face:        0,
					UserSex:     0,
					UserFace:    0,
					OfficePos:   "平民",
					Nobility:    "平民",
					UnionName:   "无联盟",
					UnionPos:    "",
				}, nil
			}
		}
		return SessionUser{}, sql.ErrNoRows
	}

	row := r.db.QueryRowContext(ctx, `
select
	u.uid,
	case when trim(coalesce(u.name, '')) = '' then concat('UID ', u.uid) else u.name end as display_name,
	coalesce(u.passport, '') as passport,
	coalesce(u.passtype, '') as passtype,
	count(c.cid) as city_count,
	coalesce(max(case when c.cid = u.lastcid then c.cid end), min(c.cid), 0) as default_cid,
	coalesce(max(case when c.cid = u.lastcid then c.name end), min(c.name), '') as default_city,
	coalesce(u.sex, 0) as sex,
	coalesce(u.face, 0) as face,
	cast(coalesce(u.prestige, 0) as signed) as prestige,
	cast(coalesce(rank_user.`+"`rank`"+`, 0) as signed) as user_rank,
	coalesce(u.officepos, 0) as officepos_id,
	coalesce(cfg_office_pos.name, '平民') as officepos,
	coalesce(u.nobility, 0) as nobility_id,
	coalesce(cfg_nobility.name, '平民') as nobility,
	coalesce(u.union_id, 0) as union_id,
	case when coalesce(u.union_id, 0) = 0 then '无联盟' else coalesce(sys_union.name, '') end as unionname,
	coalesce(u.union_pos, 0) as union_pos_id
from sys_user u
left join sys_city c on c.uid = u.uid
left join rank_user on rank_user.uid = u.uid
left join cfg_office_pos on cfg_office_pos.id = u.officepos
left join cfg_nobility on cfg_nobility.id = u.nobility
left join sys_union on sys_union.id = u.union_id
where trim(coalesce(u.passport, '')) = ?
group by u.uid, u.name, u.passport, u.passtype, u.lastcid, u.sex, u.face, u.prestige, rank_user.`+"`rank`"+`, u.officepos, cfg_office_pos.name, u.nobility, cfg_nobility.name, u.union_id, sys_union.name, u.union_pos
order by u.uid asc
limit 1`, passport)

	user := SessionUser{}
	err := row.Scan(
		&user.UID,
		&user.Name,
		&user.Passport,
		&user.PassType,
		&user.CityCount,
		&user.DefaultCID,
		&user.DefaultCity,
		&user.Sex,
		&user.Face,
		&user.Prestige,
		&user.Rank,
		&user.OfficePosID,
		&user.OfficePos,
		&user.NobilityID,
		&user.Nobility,
		&user.UnionID,
		&user.UnionName,
		&user.UnionPosID,
	)
	if err != nil {
		return SessionUser{}, err
	}
	normalizeSessionUser(&user)

	return user, nil
}

func normalizeSessionUser(user *SessionUser) {
	user.UserSex = user.Sex
	user.UserFace = user.Face
	if user.UserFace > 0 {
		user.UserFace--
	}
	if user.UnionID > 0 {
		user.UnionPos = sessionUnionPositionLabel(user.UnionPosID)
	} else {
		user.UnionPos = ""
	}
}

func sessionUnionPositionLabel(value int) string {
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

func (r *Repository) LoginByPassport(ctx context.Context, passport string, password string) (SessionUser, error) {
	passport = strings.TrimSpace(passport)
	password = strings.TrimSpace(password)
	if passport == "" || password == "" {
		return SessionUser{}, sql.ErrNoRows
	}

	user, err := r.SessionUserByPassport(ctx, passport)
	if err != nil {
		return SessionUser{}, err
	}

	passType := strings.ToLower(strings.TrimSpace(user.PassType))
	switch passType {
	case "":
		return user, nil
	case "npc":
		return SessionUser{}, newInvalidError("当前账号为内置主公，不支持密码登录，请使用主公直登。")
	default:
		return user, nil
	}
}

func (r *Repository) UserCities(ctx context.Context, uid int, limit int) ([]CityCard, error) {
	limit = clamp(limit, 1, 160)
	if r.db == nil {
		cities := r.fixtureCities()
		if limit < len(cities) {
			return cities[:limit], nil
		}
		return cities, nil
	}

	query := fmt.Sprintf(`
select
	c.cid,
	c.name,
	case when trim(coalesce(u.name, '')) = '' then concat('UID ', u.uid) else u.name end as display_name,
	r.wood,
	r.rock,
	r.iron,
	r.food,
	r.gold,
	r.people,
	r.people_max
from sys_city c
join mem_city_resource r on r.cid = c.cid
left join sys_user u on u.uid = c.uid
where c.uid = ?
order by (r.wood + r.rock + r.iron + r.food + r.gold) desc, c.cid asc
limit %d`, limit)

	rows, err := r.db.QueryContext(ctx, query, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]CityCard, 0, limit)
	for rows.Next() {
		card := CityCard{}
		if err := rows.Scan(
			&card.CID,
			&card.Name,
			&card.Owner,
			&card.Resources.Wood,
			&card.Resources.Rock,
			&card.Resources.Iron,
			&card.Resources.Food,
			&card.Resources.Gold,
			&card.Resources.People,
			&card.Resources.PeopleMax,
		); err != nil {
			return nil, err
		}

		card.X, card.Y = coordinatesFromCID(card.CID)
		items = append(items, card)
	}

	return items, rows.Err()
}

func (r *Repository) TouchLegacySession(ctx context.Context, uid int, sid int64, ip int64) error {
	if r.db == nil {
		return nil
	}

	_, err := r.db.ExecContext(ctx, `
insert into sys_sessions (uid, sid, ip)
values (?, ?, ?)
on duplicate key update sid = values(sid), ip = values(ip)`, uid, sid, ip)
	return err
}

func (r *Repository) UserOwnsCity(ctx context.Context, uid int, cid int) (bool, error) {
	if r.db == nil {
		return true, nil
	}

	var count int
	if err := r.db.QueryRowContext(ctx, "select count(*) from sys_city where uid = ? and cid = ?", uid, cid).Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}
