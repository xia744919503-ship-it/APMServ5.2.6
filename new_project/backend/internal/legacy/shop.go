package legacy

import "context"

func (r *Repository) MyShop(ctx context.Context, uid int) (ShopSnapshot, error) {
	snapshot := ShopSnapshot{
		Medals: []ShopMedal{},
		Groups: []ShopGroup{},
	}
	if r.db == nil {
		return snapshot, nil
	}

	if err := r.db.QueryRowContext(ctx, `
select
	coalesce(u.name, ''),
	coalesce(u.lastcid, 0),
	coalesce(city.name, ''),
	coalesce(u.money, 0),
	coalesce(goods.count, 0),
	coalesce(res.gold, 0),
	coalesce(u.honour, 0)
from sys_user u
left join sys_city city on city.cid = u.lastcid
left join mem_city_resource res on res.cid = u.lastcid
left join sys_goods goods on goods.uid = u.uid and goods.gid = 0
where u.uid = ?`, uid).Scan(
		&snapshot.Wallet.UserName,
		&snapshot.Wallet.FocusCID,
		&snapshot.Wallet.FocusCity,
		&snapshot.Wallet.Yuanbao,
		&snapshot.Wallet.Gift,
		&snapshot.Wallet.Gold,
		&snapshot.Wallet.Honour,
	); err != nil {
		return ShopSnapshot{}, err
	}

	medalCounts := map[int]int64{}
	rows, err := r.db.QueryContext(ctx, `
select tid, count
from sys_things
where uid = ? and tid in (30000, 30001, 30002, 30003)`, uid)
	if err != nil {
		return ShopSnapshot{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var thingID int
		var count int64
		if err := rows.Scan(&thingID, &count); err != nil {
			return ShopSnapshot{}, err
		}
		medalCounts[thingID] = count
	}
	if err := rows.Err(); err != nil {
		return ShopSnapshot{}, err
	}

	for _, thingID := range []int{30000, 30001, 30002, 30003} {
		snapshot.Medals = append(snapshot.Medals, ShopMedal{
			ThingID: thingID,
			Name:    shopMedalLabel(thingID),
			Count:   medalCounts[thingID],
		})
	}

	totalBought := map[int]int64{}
	totalRows, err := r.db.QueryContext(ctx, `
select sid, count
from log_shop_buy_cnt
where uid = ?`, uid)
	if err != nil {
		return ShopSnapshot{}, err
	}
	defer totalRows.Close()

	for totalRows.Next() {
		var itemID int
		var count int64
		if err := totalRows.Scan(&itemID, &count); err != nil {
			return ShopSnapshot{}, err
		}
		totalBought[itemID] = count
	}
	if err := totalRows.Err(); err != nil {
		return ShopSnapshot{}, err
	}

	todayBought := map[int]int64{}
	todayRows, err := r.db.QueryContext(ctx, `
select shopid, coalesce(sum(count), 0)
from log_shop
where uid = ? and date(from_unixtime(time)) = curdate()
group by shopid`, uid)
	if err != nil {
		return ShopSnapshot{}, err
	}
	defer todayRows.Close()

	for todayRows.Next() {
		var itemID int
		var count int64
		if err := todayRows.Scan(&itemID, &count); err != nil {
			return ShopSnapshot{}, err
		}
		todayBought[itemID] = count
	}
	if err := todayRows.Err(); err != nil {
		return ShopSnapshot{}, err
	}

	itemRows, err := r.db.QueryContext(ctx, `
select
	id,
	gid,
	`+"`group`"+` as group_id,
	name,
	coalesce(description, ''),
	coalesce(pack, 1),
	coalesce(price, 0),
	coalesce(oriprice, 0),
	coalesce(totalCount, 0),
	coalesce(userbuycnt, 0),
	coalesce(daybuycnt, 0),
	coalesce(battledaybuycnt, 0),
	coalesce(position, 0),
	coalesce(commend, 0),
	coalesce(hot, 0),
	coalesce(battleshop, 0),
	coalesce(creditPrice, 0),
	coalesce(medalPrice, 0),
	coalesce(medalTypeId, 30000),
	coalesce(battleGoodsType, 0)
from cfg_shop
where onsale = 1 and starttime <= unix_timestamp() and endtime > unix_timestamp()
order by position asc, id asc`)
	if err != nil {
		return ShopSnapshot{}, err
	}
	defer itemRows.Close()

	groupMap := map[int]*ShopGroup{}
	groupOrder := make([]int, 0, 8)
	for itemRows.Next() {
		var (
			item       ShopItem
			commended  int
			hot        int
			battleShop int
		)
		if err := itemRows.Scan(
			&item.ID,
			&item.GID,
			&item.GroupID,
			&item.Name,
			&item.Description,
			&item.Pack,
			&item.Price,
			&item.OriginalPrice,
			&item.TotalCount,
			&item.UserLimit,
			&item.DayLimit,
			&item.BattleDayLimit,
			&item.Position,
			&commended,
			&hot,
			&battleShop,
			&item.CreditPrice,
			&item.MedalPrice,
			&item.MedalTypeID,
			&item.BattleGoodsType,
		); err != nil {
			return ShopSnapshot{}, err
		}

		item.GroupLabel = shopGroupLabel(item.GroupID)
		item.Commended = commended != 0
		item.Hot = hot != 0
		item.BattleShop = battleShop != 0
		item.MedalTypeLabel = shopMedalLabel(item.MedalTypeID)
		item.BoughtTotal = totalBought[item.ID]
		item.BoughtToday = todayBought[item.ID]
		if item.BattleShop && item.BattleDayLimit > 0 && item.BoughtToday == 0 {
			item.BoughtToday = item.BoughtTotal
		}

		group := groupMap[item.GroupID]
		if group == nil {
			group = &ShopGroup{
				ID:    item.GroupID,
				Label: shopGroupLabel(item.GroupID),
				Items: []ShopItem{},
			}
			groupMap[item.GroupID] = group
			groupOrder = append(groupOrder, item.GroupID)
		}

		group.Items = append(group.Items, item)
		group.ItemCount++
	}
	if err := itemRows.Err(); err != nil {
		return ShopSnapshot{}, err
	}

	for _, groupID := range groupOrder {
		if group := groupMap[groupID]; group != nil {
			snapshot.Groups = append(snapshot.Groups, *group)
		}
	}

	return snapshot, nil
}

func shopGroupLabel(groupID int) string {
	switch groupID {
	case 0:
		return "常用道具"
	case 1:
		return "军务补给"
	case 2:
		return "内政增益"
	case 3:
		return "礼包宝箱"
	case 6:
		return "战场兑换"
	default:
		return "其他货架"
	}
}

func shopMedalLabel(thingID int) string {
	switch thingID {
	case 30000:
		return "汉室勋章"
	case 30001:
		return "平定黄巾勋章"
	case 30002:
		return "袁军官渡勋章"
	case 30003:
		return "曹军官渡勋章"
	default:
		return "未知勋章"
	}
}
