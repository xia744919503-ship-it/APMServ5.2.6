package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	chargeExchangeRate int64 = 10
	chargeLogType      int64 = 0
	chargeRewardType   int64 = 6
	maxChargeExchange  int64 = 100000
)

type chargeAccount struct {
	Passport string
	PassType string
}

type chargeActGift struct {
	ActID       int
	MoneyLimit  int64
	DayCount    int64
	MailTitle   string
	MailContent string
}

type chargeBoxDetail struct {
	Sort  int
	Type  int
	Count int64
}

func (r *Repository) ExchangeCharge(ctx context.Context, uid int, exchangeCount int64) (ChargeSnapshot, error) {
	if exchangeCount <= 0 {
		return ChargeSnapshot{}, newInvalidError("充值数量无效。")
	}
	if exchangeCount > maxChargeExchange {
		return ChargeSnapshot{}, newInvalidError("单次兑换数量过大。")
	}
	if r.db == nil {
		return ChargeSnapshot{}, newInvalidError("当前环境不支持充值兑换。")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ChargeSnapshot{}, err
	}
	defer tx.Rollback()

	account, err := r.chargeAccountTx(ctx, tx, uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return ChargeSnapshot{}, newInvalidError("未找到当前主公。")
		}
		return ChargeSnapshot{}, err
	}
	if strings.TrimSpace(account.Passport) == "" || strings.TrimSpace(account.PassType) == "" {
		return ChargeSnapshot{}, newInvalidError("当前主公缺少旧服通行证信息，无法充值。")
	}

	yuanbao := exchangeCount * chargeExchangeRate
	now := time.Now().Unix()
	orderID := fmt.Sprintf("rexue-%d-%d", uid, time.Now().UnixNano())
	day := now - ((now + 8*3600) % 86400)

	if _, err := tx.ExecContext(ctx, `
update sys_user
set money = money + ?
where uid = ?`, yuanbao, uid); err != nil {
		return ChargeSnapshot{}, err
	}

	if _, err := tx.ExecContext(ctx, `
insert into log_money (uid, count, time, type)
values (?, ?, ?, ?)`, uid, yuanbao, now, chargeLogType); err != nil {
		return ChargeSnapshot{}, err
	}

	if _, err := tx.ExecContext(ctx, `
insert into pay_log (orderid, type, payname, passport, passtype, money, code, time)
values (?, ?, 'xiaonei', ?, ?, ?, 'rexue', ?)`, orderID, chargeLogType, account.Passport, account.PassType, yuanbao, now); err != nil {
		return ChargeSnapshot{}, err
	}

	if _, err := tx.ExecContext(ctx, `
insert into pay_day_money (day, money)
values (?, ?)
on duplicate key update money = money + values(money)`, day, yuanbao); err != nil {
		return ChargeSnapshot{}, err
	}

	if err := r.applyFirstPayGiftTx(ctx, tx, uid, yuanbao, now); err != nil {
		return ChargeSnapshot{}, err
	}
	if err := r.applyActPayGiftTx(ctx, tx, uid, yuanbao, now); err != nil {
		return ChargeSnapshot{}, err
	}

	if err := tx.Commit(); err != nil {
		return ChargeSnapshot{}, err
	}

	return r.MyCharge(ctx, uid)
}

func (r *Repository) chargeAccountTx(ctx context.Context, tx *sql.Tx, uid int) (chargeAccount, error) {
	account := chargeAccount{}
	err := tx.QueryRowContext(ctx, `
select
	coalesce(passport, ''),
	coalesce(passtype, '')
from sys_user
where uid = ?`, uid).Scan(&account.Passport, &account.PassType)
	if err != nil {
		return chargeAccount{}, err
	}
	return account, nil
}

func (r *Repository) applyFirstPayGiftTx(ctx context.Context, tx *sql.Tx, uid int, yuanbao int64, now int64) error {
	if yuanbao < 50 {
		return nil
	}

	var endTime sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
select value
from mem_state
where state = 21`).Scan(&endTime); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	if endTime.Valid && endTime.Int64 > 0 && now >= endTime.Int64 {
		return nil
	}

	var gifted int
	if err := tx.QueryRowContext(ctx, `
select count(*)
from pay_user_gift
where uid = ?`, uid).Scan(&gifted); err != nil {
		return err
	}
	if gifted > 0 {
		return nil
	}

	nowTime := time.Now().Unix()
	goods := []struct {
		GID   int
		Count int64
	}{
		{GID: 24, Count: 1},
		{GID: 40, Count: 1},
		{GID: 56, Count: 1},
		{GID: 30, Count: 2},
		{GID: 96, Count: 1},
	}
	for _, item := range goods {
		if err := r.chargeGrantGoodsTx(ctx, tx, uid, item.GID, item.Count, nowTime); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `
insert into pay_user_gift (uid, time)
values (?, ?)`, uid, nowTime); err != nil {
		return err
	}

	return r.sendSystemMailTx(
		ctx,
		tx,
		uid,
		"新服首冲送大礼",
		"亲爱的玩家：\n\n感谢您参加本次“新服首冲送大礼”充值活动，您已获得：迁城令*1、建筑图纸*1、徭役令*1、珍珠*2、白色装备箱*1，请注意查收您的物品栏，祝您游戏愉快！\n\n           《热血三国》运营团队",
		nowTime,
	)
}

func (r *Repository) applyActPayGiftTx(ctx context.Context, tx *sql.Tx, uid int, yuanbao int64, now int64) error {
	rows, err := tx.QueryContext(ctx, `
select
	a.actid,
	coalesce(p.money_limit, 0),
	coalesce(p.daycnt, 0),
	coalesce(p.mailtitle, ''),
	coalesce(p.mailcontent, '')
from cfg_act a
join cfg_act_paygift p on p.actid = a.actid
where a.type = 2 and ? >= a.starttime and ? <= a.endtime
order by a.actid asc`, now, now)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		act := chargeActGift{}
		if err := rows.Scan(&act.ActID, &act.MoneyLimit, &act.DayCount, &act.MailTitle, &act.MailContent); err != nil {
			return err
		}
		if act.MoneyLimit <= 0 || yuanbao < act.MoneyLimit {
			continue
		}

		giveCount := yuanbao / act.MoneyLimit
		if giveCount <= 0 {
			continue
		}

		details, err := r.chargeBoxDetailsTx(ctx, tx, act.ActID)
		if err != nil {
			return err
		}
		if len(details) == 0 {
			continue
		}

		for _, detail := range details {
			grantedToday, err := r.chargeGrantedTodayTx(ctx, tx, uid, detail)
			if err != nil {
				return err
			}
			if giveCount > act.DayCount-grantedToday {
				giveCount = act.DayCount - grantedToday
			}
			if giveCount <= 0 {
				break
			}
			if err := r.applyChargeBoxDetailTx(ctx, tx, uid, detail, giveCount, now); err != nil {
				return err
			}
		}

		if giveCount > 0 {
			content := strings.ReplaceAll(act.MailContent, "{giveCount}", strconv.FormatInt(giveCount, 10))
			if err := r.sendSystemMailTx(ctx, tx, uid, act.MailTitle, content, now); err != nil {
				return err
			}
		}
	}

	return rows.Err()
}

func (r *Repository) chargeBoxDetailsTx(ctx context.Context, tx *sql.Tx, actID int) ([]chargeBoxDetail, error) {
	rows, err := tx.QueryContext(ctx, `
select sort, type, count
from cfg_box_details
where srctype = 1 and srcid = ?
order by id asc`, actID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	details := make([]chargeBoxDetail, 0, 8)
	for rows.Next() {
		detail := chargeBoxDetail{}
		if err := rows.Scan(&detail.Sort, &detail.Type, &detail.Count); err != nil {
			return nil, err
		}
		details = append(details, detail)
	}

	return details, rows.Err()
}

func (r *Repository) chargeGrantedTodayTx(ctx context.Context, tx *sql.Tx, uid int, detail chargeBoxDetail) (int64, error) {
	var query string
	switch detail.Sort {
	case 2:
		query = `
select coalesce(sum(count), 0)
from log_goods
where uid = ? and type = ? and gid = ? and curdate() = date(from_unixtime(time))`
	case 5:
		query = `
select coalesce(sum(count), 0)
from log_things
where uid = ? and type = ? and tid = ? and curdate() = date(from_unixtime(time))`
	case 6:
		query = `
select coalesce(sum(count), 0)
from log_armor
where uid = ? and type = ? and armorid = ? and curdate() = date(from_unixtime(time))`
	default:
		return 0, newInvalidError("当前充值礼包类型暂未支持。")
	}

	var granted int64
	if err := tx.QueryRowContext(ctx, query, uid, chargeRewardType, detail.Type).Scan(&granted); err != nil {
		return 0, err
	}
	return granted, nil
}

func (r *Repository) applyChargeBoxDetailTx(ctx context.Context, tx *sql.Tx, uid int, detail chargeBoxDetail, giveCount int64, now int64) error {
	total := detail.Count * giveCount
	if total <= 0 {
		return nil
	}

	switch detail.Sort {
	case 2:
		return r.chargeGrantGoodsTx(ctx, tx, uid, detail.Type, total, now)
	case 5:
		return r.chargeGrantThingsTx(ctx, tx, uid, detail.Type, total, now)
	case 6:
		return r.chargeGrantArmorTx(ctx, tx, uid, detail.Type, total, now)
	default:
		return newInvalidError("当前充值礼包类型暂未支持。")
	}
}

func (r *Repository) chargeGrantGoodsTx(ctx context.Context, tx *sql.Tx, uid int, gid int, count int64, now int64) error {
	if err := r.addUserGoodsTx(ctx, tx, uid, gid, count); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
insert into log_goods (uid, gid, count, time, type)
values (?, ?, ?, ?, ?)`, uid, gid, count, now, chargeRewardType)
	return err
}

func (r *Repository) chargeGrantThingsTx(ctx context.Context, tx *sql.Tx, uid int, tid int, count int64, now int64) error {
	if _, err := tx.ExecContext(ctx, `
insert into sys_things (uid, tid, count)
values (?, ?, ?)
on duplicate key update count = count + values(count)`, uid, tid, count); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
insert into log_things (uid, tid, count, time, type)
values (?, ?, ?, ?, ?)`, uid, tid, count, now, chargeRewardType)
	return err
}

func (r *Repository) chargeGrantArmorTx(ctx context.Context, tx *sql.Tx, uid int, armorID int, count int64, now int64) error {
	var hpMax int64
	if err := tx.QueryRowContext(ctx, `
select coalesce(ori_hp_max, 0)
from cfg_armor
where id = ?`, armorID).Scan(&hpMax); err != nil {
		return err
	}

	for index := int64(0); index < count; index++ {
		if _, err := tx.ExecContext(ctx, `
insert into sys_user_armor (uid, armorid, hp, hp_max, hid)
values (?, ?, ?, ?, 0)`, uid, armorID, hpMax*10, hpMax); err != nil {
			return err
		}
	}

	_, err := tx.ExecContext(ctx, `
insert into log_armor (uid, armorid, count, time, type)
values (?, ?, ?, ?, ?)`, uid, armorID, count, now, chargeRewardType)
	return err
}

func (r *Repository) sendSystemMailTx(ctx context.Context, tx *sql.Tx, uid int, title string, content string, now int64) error {
	contentResult, err := tx.ExecContext(ctx, `
insert into sys_mail_sys_content (content, posttime)
values (?, ?)`, content, now)
	if err != nil {
		return err
	}

	contentID, err := contentResult.LastInsertId()
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
insert into sys_mail_sys_box (uid, contentid, title, `+"`read`"+`, posttime)
values (?, ?, ?, 0, ?)`, uid, contentID, title, now); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
insert into sys_alarm (uid, mail)
values (?, 1)
on duplicate key update mail = 1`, uid)
	return err
}
