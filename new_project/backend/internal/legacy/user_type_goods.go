package legacy

import (
	"context"
	"database/sql"
	"math"
	"strings"
)

type UserTypeGoodsItem struct {
	GID         int    `json:"gid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       int    `json:"image"`
	Value       int64  `json:"value"`
	Group       int    `json:"group"`
	Count       int64  `json:"count"`
}

type UserTypeGoodsSnapshot struct {
	Type      int                 `json:"type"`
	TimeLeft  int64               `json:"timeLeft"`
	Cost      int64               `json:"cost,omitempty"`
	GoodsList []UserTypeGoodsItem `json:"goodsList"`
	Legacy    []any               `json:"legacy"`
}

func (r *Repository) UserTypeGoods(ctx context.Context, uid int, goodsType int, timeLeft int64) (UserTypeGoodsSnapshot, error) {
	if timeLeft < 0 {
		timeLeft = 0
	}

	gids, cost, err := userTypeGoodsSelection(goodsType, timeLeft)
	if err != nil {
		return UserTypeGoodsSnapshot{}, err
	}

	snapshot := UserTypeGoodsSnapshot{
		Type:     goodsType,
		TimeLeft: timeLeft,
		Cost:     cost,
	}
	if r.db == nil {
		snapshot.GoodsList = []UserTypeGoodsItem{}
		snapshot.Legacy = userTypeGoodsLegacyPayload(goodsType, cost, snapshot.GoodsList)
		return snapshot, nil
	}

	items, err := r.queryUserTypeGoods(ctx, uid, gids)
	if err != nil {
		return UserTypeGoodsSnapshot{}, err
	}

	snapshot.GoodsList = items
	snapshot.Legacy = userTypeGoodsLegacyPayload(goodsType, cost, items)
	return snapshot, nil
}

func (r *Repository) UseUserTypeGoods(ctx context.Context, uid int, cid int, goodsType int, gid int) (CityDetail, error) {
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return CityDetail{}, err
	} else if !allowed {
		return CityDetail{}, ErrForbidden
	}
	if r.db == nil {
		return r.fixtureCityDetail(cid), nil
	}
	if goodsType < 4 || goodsType > 10 {
		return CityDetail{}, newInvalidError("unsupported goods type")
	}

	gids, _, err := userTypeGoodsSelection(goodsType, 0)
	if err != nil {
		return CityDetail{}, err
	}
	if !intInSlice(gid, gids) {
		return CityDetail{}, newInvalidError("goods does not belong to this type")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return CityDetail{}, err
	}
	defer tx.Rollback()

	now, err := r.currentUnixTimeTx(ctx, tx)
	if err != nil {
		return CityDetail{}, err
	}
	if err := r.useUserTypeGoodsTx(ctx, tx, uid, cid, goodsType, gid, now); err != nil {
		return CityDetail{}, err
	}
	if err := tx.Commit(); err != nil {
		return CityDetail{}, err
	}
	return r.CityDetail(ctx, cid)
}

func (r *Repository) useUserTypeGoodsTx(ctx context.Context, tx *sql.Tx, uid int, cid int, goodsType int, gid int, now int64) error {
	switch goodsType {
	case 4:
		if err := r.useMoraleGoodsTx(ctx, tx, uid, cid, gid, now); err != nil {
			return err
		}
	case 5:
		if err := r.useGoldRateGoodsTx(ctx, tx, uid, gid); err != nil {
			return err
		}
	case 6:
		if err := r.usePeopleGoodsTx(ctx, tx, uid, cid, gid); err != nil {
			return err
		}
	case 7, 8, 9, 10:
		if err := r.useResourceRateGoodsTx(ctx, tx, uid, goodsType, gid); err != nil {
			return err
		}
	default:
		return newInvalidError("unsupported goods type")
	}

	if err := r.consumeUserGoodsTx(ctx, tx, uid, gid); err != nil {
		return err
	}
	return r.recalculateCityProduction(ctx, tx, cid)
}

func (r *Repository) useResourceRateGoodsTx(ctx context.Context, tx *sql.Tx, uid int, goodsType int, gid int) error {
	bufferType := goodsType - 6
	column := map[int]string{
		7:  "goods_food_add",
		8:  "goods_wood_add",
		9:  "goods_rock_add",
		10: "goods_iron_add",
	}[goodsType]
	if column == "" {
		return newInvalidError("unsupported resource goods type")
	}

	delay := int64(86400)
	if gid >= 44 && gid <= 47 {
		delay *= 7
	}
	if _, err := tx.ExecContext(ctx, `
insert into mem_user_buffer (uid, buftype, endtime)
values (?, ?, unix_timestamp() + ?)
on duplicate key update endtime = endtime + ?`, uid, bufferType, delay, delay); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
update sys_city_res_add a
join sys_city c on c.cid = a.cid
set `+column+` = 25, a.resource_changing = 1
where c.uid = ?`, uid); err != nil {
		return err
	}
	return nil
}

func (r *Repository) useGoldRateGoodsTx(ctx context.Context, tx *sql.Tx, uid int, gid int) error {
	delay := int64(86400)
	if gid == 55 {
		delay *= 7
	}
	if _, err := tx.ExecContext(ctx, `
update mem_city_resource m
join sys_city c on c.cid = m.cid
set m.gold_rate = 125
where c.uid = ?`, uid); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
insert into mem_user_buffer (uid, buftype, endtime)
values (?, 15, unix_timestamp() + ?)
on duplicate key update endtime = endtime + ?`, uid, delay, delay)
	return err
}

func (r *Repository) usePeopleGoodsTx(ctx context.Context, tx *sql.Tx, uid int, cid int, gid int) error {
	var people int64
	var peopleMax int64
	if err := tx.QueryRowContext(ctx, `
select people, people_max
from mem_city_resource
where cid = ?`, cid).Scan(&people, &peopleMax); err != nil {
		return err
	}
	if people >= peopleMax {
		return newInvalidError("city population is full")
	}

	add := int64(math.Ceil(float64(peopleMax) * 0.2))
	if add < 100 {
		add = 100
	}
	if _, err := tx.ExecContext(ctx, "update mem_city_resource set people = people + ? where cid = ?", add, cid); err != nil {
		return err
	}
	return r.markUserCitiesResourceChangingTx(ctx, tx, uid)
}

func (r *Repository) useMoraleGoodsTx(ctx context.Context, tx *sql.Tx, uid int, cid int, gid int, now int64) error {
	var last sql.NullInt64
	if err := tx.QueryRowContext(ctx, "select last_anming from mem_city_schedule where cid = ?", cid).Scan(&last); err != nil && err != sql.ErrNoRows {
		return err
	}
	if last.Valid && now-last.Int64 < 259200 {
		return newInvalidError("morale goods is cooling down")
	}

	if _, err := tx.ExecContext(ctx, `
update mem_city_resource
set morale = 100,
	complaint = 0,
	morale_stable = 100 - tax,
	people_stable = people_max
where cid = ?`, cid); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
insert into mem_city_schedule (cid, last_anming)
values (?, ?)
on duplicate key update last_anming = ?`, cid, now, now)
	return err
}

func (r *Repository) markUserCitiesResourceChangingTx(ctx context.Context, tx *sql.Tx, uid int) error {
	_, err := tx.ExecContext(ctx, `
update sys_city_res_add a
join sys_city c on c.cid = a.cid
set a.resource_changing = 1
where c.uid = ?`, uid)
	return err
}

func (r *Repository) consumeUserGoodsTx(ctx context.Context, tx *sql.Tx, uid int, gid int) error {
	result, err := tx.ExecContext(ctx, `
update sys_goods
set count = count - 1
where uid = ? and gid = ? and count >= 1`, uid, gid)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return newInvalidError("goods count is not enough")
	}
	return nil
}

func (r *Repository) queryUserTypeGoods(ctx context.Context, uid int, gids []int) ([]UserTypeGoodsItem, error) {
	placeholders := strings.TrimRight(strings.Repeat("?,", len(gids)), ",")
	args := make([]any, 0, len(gids)+1)
	args = append(args, uid)
	for _, gid := range gids {
		args = append(args, gid)
	}

	rows, err := r.db.QueryContext(ctx, `
select c.gid, c.name, coalesce(c.description, ''), coalesce(c.image, 0), coalesce(c.value, 0), coalesce(c.`+"`group`"+`, 0), coalesce(g.count, 0)
from cfg_goods c
left join sys_goods g on g.gid = c.gid and g.uid = ?
where c.gid in (`+placeholders+`)
order by c.value`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]UserTypeGoodsItem, 0, len(gids))
	var instantComplete *UserTypeGoodsItem
	for rows.Next() {
		item := UserTypeGoodsItem{}
		if err := rows.Scan(&item.GID, &item.Name, &item.Description, &item.Image, &item.Value, &item.Group, &item.Count); err != nil {
			return nil, err
		}
		if item.GID == 73 {
			copy := item
			instantComplete = &copy
			continue
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if instantComplete != nil {
		items = append([]UserTypeGoodsItem{*instantComplete}, items...)
	}
	return items, nil
}

func userTypeGoodsSelection(goodsType int, timeLeft int64) ([]int, int64, error) {
	switch goodsType {
	case 0:
		return []int{67, 68, 69, 70, 71, 72, 73}, int64(math.Floor(float64(timeLeft)*4.5/3600.0) + 25), nil
	case 4:
		return []int{58}, 0, nil
	case 5:
		return []int{54, 55}, 0, nil
	case 6:
		return []int{57}, 0, nil
	case 7:
		return []int{2, 44}, 0, nil
	case 8:
		return []int{3, 45}, 0, nil
	case 9:
		return []int{4, 46}, 0, nil
	case 10:
		return []int{5, 47}, 0, nil
	default:
		return nil, 0, newInvalidError("unsupported goods type")
	}
}

func userTypeGoodsLegacyPayload(goodsType int, cost int64, items []UserTypeGoodsItem) []any {
	if goodsType == 0 {
		return []any{cost, items}
	}
	return []any{items}
}

func intInSlice(value int, values []int) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}
