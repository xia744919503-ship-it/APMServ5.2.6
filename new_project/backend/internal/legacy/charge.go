package legacy

import (
	"context"
	"strings"
	"time"
)

func (r *Repository) MyCharge(ctx context.Context, uid int) (ChargeSnapshot, error) {
	snapshot := ChargeSnapshot{
		Summary: ChargeSummary{
			ExchangeRate: 10,
			ReadOnly:     true,
		},
		Buckets: defaultChargeBuckets(),
		Events:  []ChargeEvent{},
	}
	if r.db == nil {
		return snapshot, nil
	}

	var (
		passport string
		passtype string
	)
	if err := r.db.QueryRowContext(ctx, `
select
	coalesce(u.name, ''),
	coalesce(city.name, ''),
	coalesce(u.money, 0),
	coalesce(goods.count, 0),
	coalesce(u.passport, ''),
	coalesce(u.passtype, '')
from sys_user u
left join sys_city city on city.cid = u.lastcid
left join sys_goods goods on goods.uid = u.uid and goods.gid = 0
where u.uid = ?`, uid).Scan(
		&snapshot.Summary.UserName,
		&snapshot.Summary.FocusCity,
		&snapshot.Summary.Yuanbao,
		&snapshot.Summary.Gift,
		&passport,
		&passtype,
	); err != nil {
		return ChargeSnapshot{}, err
	}

	snapshot.Summary.ReadOnly = strings.TrimSpace(passport) == "" || strings.TrimSpace(passtype) == ""

	if passport != "" && passtype != "" {
		if err := r.db.QueryRowContext(ctx, `
select
	coalesce(floor(sum(money) / ?), 0),
	count(*)
from pay_log
where passport = ? and passtype = ?`, snapshot.Summary.ExchangeRate, passport, passtype).Scan(&snapshot.Summary.TotalPaid, &snapshot.Summary.PayCount); err != nil {
			return ChargeSnapshot{}, err
		}

		if err := r.db.QueryRowContext(ctx, `
select coalesce(floor(sum(money) / ?), 0)
from pay_log
where passport = ? and passtype = ? and date(from_unixtime(time)) = curdate()`, snapshot.Summary.ExchangeRate, passport, passtype).Scan(&snapshot.Summary.TodayPaid); err != nil {
			return ChargeSnapshot{}, err
		}

		hasOrderTable := 0
		if err := r.db.QueryRowContext(ctx, `
select count(*)
from information_schema.tables
where table_schema = database() and table_name = 'log_51_charge'`).Scan(&hasOrderTable); err == nil && hasOrderTable > 0 {
			_ = r.db.QueryRowContext(ctx, `
select count(*)
from log_51_charge
where passport = ? and state = 0`, passport).Scan(&snapshot.Summary.PendingOrders)
		}
	}

	for index, bucket := range snapshot.Buckets {
		snapshot.Buckets[index].Yuanbao = bucket.MinMoney * int64(snapshot.Summary.ExchangeRate)
		if err := r.db.QueryRowContext(ctx, chargeBucketQuery(bucket), chargeBucketArgs(bucket)...).Scan(&snapshot.Buckets[index].PlayerCount); err != nil {
			return ChargeSnapshot{}, err
		}
	}

	rows, err := r.db.QueryContext(ctx, `
select
	a.actid,
	coalesce(a.name, ''),
	coalesce(p.money_limit, 0),
	coalesce(p.daycnt, 0),
	coalesce(p.mailtitle, ''),
	a.starttime,
	a.endtime
from cfg_act a
join cfg_act_paygift p on p.actid = a.actid
where a.type = 2
order by a.endtime desc
limit 8`)
	if err != nil {
		return ChargeSnapshot{}, err
	}
	defer rows.Close()

	now := time.Now().Unix()
	for rows.Next() {
		event := ChargeEvent{}
		var startUnix int64
		var endUnix int64
		if err := rows.Scan(
			&event.ActID,
			&event.Name,
			&event.MoneyLimit,
			&event.DayCount,
			&event.MailTitle,
			&startUnix,
			&endUnix,
		); err != nil {
			return ChargeSnapshot{}, err
		}

		event.StartAt = time.Unix(startUnix, 0).Format("2006-01-02 15:04:05")
		event.EndAt = time.Unix(endUnix, 0).Format("2006-01-02 15:04:05")
		event.Active = now >= startUnix && now <= endUnix
		snapshot.Events = append(snapshot.Events, event)
	}
	if err := rows.Err(); err != nil {
		return ChargeSnapshot{}, err
	}

	return snapshot, nil
}

func defaultChargeBuckets() []ChargeBucket {
	return []ChargeBucket{
		{ID: "r1", Label: "1-10 元", MinMoney: 1, MaxMoney: 10},
		{ID: "r2", Label: "11-30 元", MinMoney: 11, MaxMoney: 30},
		{ID: "r3", Label: "31-50 元", MinMoney: 31, MaxMoney: 50},
		{ID: "r4", Label: "51-100 元", MinMoney: 51, MaxMoney: 100},
		{ID: "r5", Label: "101-200 元", MinMoney: 101, MaxMoney: 200},
		{ID: "r6", Label: "201-300 元", MinMoney: 201, MaxMoney: 300},
		{ID: "r7", Label: "301-500 元", MinMoney: 301, MaxMoney: 500},
		{ID: "r8", Label: "501 元以上", MinMoney: 501, MaxMoney: 0},
	}
}

func chargeBucketQuery(bucket ChargeBucket) string {
	if bucket.MaxMoney > 0 {
		return `
select count(*)
from (
	select passport
	from pay_log
	group by passport
	having floor(sum(money) / 10) >= ? and floor(sum(money) / 10) <= ?
) p`
	}

	return `
select count(*)
from (
	select passport
	from pay_log
	group by passport
	having floor(sum(money) / 10) >= ?
) p`
}

func chargeBucketArgs(bucket ChargeBucket) []any {
	if bucket.MaxMoney > 0 {
		return []any{bucket.MinMoney, bucket.MaxMoney}
	}
	return []any{bucket.MinMoney}
}
