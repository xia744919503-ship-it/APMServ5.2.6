package legacy

import (
	"context"
	"database/sql"
	"time"
)

type shopPurchaseItem struct {
	ID             int
	GID            int
	Pack           int
	Price          int64
	TotalCount     int64
	UserLimit      int
	DayLimit       int
	BattleDayLimit int
	BattleShop     bool
	CreditPrice    int64
	MedalPrice     int64
	MedalTypeID    int
	BattleGoods    int
}

func (r *Repository) BuyShopItem(ctx context.Context, uid int, itemID int, count int, payType int, cityID int) (ShopSnapshot, error) {
	if itemID <= 0 || count <= 0 {
		return ShopSnapshot{}, newInvalidError("商品或数量无效。")
	}
	if payType != 0 && payType != 1 {
		return ShopSnapshot{}, newInvalidError("支付方式无效。")
	}
	if r.db == nil {
		return ShopSnapshot{}, newInvalidError("当前环境不支持商城购买。")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ShopSnapshot{}, err
	}
	defer tx.Rollback()

	item, err := r.shopPurchaseItem(ctx, tx, itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ShopSnapshot{}, newInvalidError("商品已下架。")
		}
		return ShopSnapshot{}, err
	}

	if item.BattleShop {
		if err := r.buyBattleShopItemTx(ctx, tx, uid, item, count); err != nil {
			return ShopSnapshot{}, err
		}
	} else {
		if err := r.buyNormalShopItemTx(ctx, tx, uid, item, count, payType); err != nil {
			return ShopSnapshot{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return ShopSnapshot{}, err
	}

	_ = cityID
	return r.MyShop(ctx, uid)
}

func (r *Repository) shopPurchaseItem(ctx context.Context, tx *sql.Tx, itemID int) (shopPurchaseItem, error) {
	item := shopPurchaseItem{}
	var battleShop int
	err := tx.QueryRowContext(ctx, `
select
	id,
	gid,
	coalesce(pack, 1),
	coalesce(price, 0),
	coalesce(totalCount, 0),
	coalesce(userbuycnt, 0),
	coalesce(daybuycnt, 0),
	coalesce(battledaybuycnt, 0),
	coalesce(battleshop, 0),
	coalesce(creditPrice, 0),
	coalesce(medalPrice, 0),
	coalesce(medalTypeId, 30000),
	coalesce(battleGoodsType, 0)
from cfg_shop
where id = ?
	and onsale = 1
	and starttime <= unix_timestamp()
	and endtime > unix_timestamp()`, itemID).Scan(
		&item.ID,
		&item.GID,
		&item.Pack,
		&item.Price,
		&item.TotalCount,
		&item.UserLimit,
		&item.DayLimit,
		&item.BattleDayLimit,
		&battleShop,
		&item.CreditPrice,
		&item.MedalPrice,
		&item.MedalTypeID,
		&item.BattleGoods,
	)
	if err != nil {
		return shopPurchaseItem{}, err
	}
	if item.Pack <= 0 {
		item.Pack = 1
	}
	item.BattleShop = battleShop != 0
	return item, nil
}

func (r *Repository) buyNormalShopItemTx(ctx context.Context, tx *sql.Tx, uid int, item shopPurchaseItem, count int, payType int) error {
	if item.TotalCount == 0 {
		return newInvalidError("商品已售罄。")
	}

	totalBought, todayBought, err := r.shopPurchaseCountsTx(ctx, tx, uid, item.ID)
	if err != nil {
		return err
	}
	if item.UserLimit > 0 && totalBought+count > item.UserLimit {
		return newInvalidError("超过商品总限购。")
	}
	if item.DayLimit > 0 && todayBought+count > item.DayLimit {
		return newInvalidError("超过商品日限购。")
	}

	if payType == 0 {
		result, err := tx.ExecContext(ctx, `
update sys_user
set money = money - ?, last_pay = ?
where uid = ? and money >= ?`, item.Price*int64(count), payType, uid, item.Price*int64(count))
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return newInvalidError("元宝不足，无法购买。")
		}
	} else {
		if err := r.ensureGiftWalletTx(ctx, tx, uid); err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `
update sys_goods
set count = count - ?
where uid = ? and gid = 0 and count >= ?`, item.Price*int64(count), uid, item.Price*int64(count))
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return newInvalidError("礼金不足，无法购买。")
		}
		if _, err := tx.ExecContext(ctx, "update sys_user set last_pay = ? where uid = ?", payType, uid); err != nil {
			return err
		}
	}

	if err := r.addUserGoodsTx(ctx, tx, uid, item.GID, int64(item.Pack*count)); err != nil {
		return err
	}
	if err := r.bumpShopBuyCountTx(ctx, tx, uid, item.ID, count); err != nil {
		return err
	}

	now := time.Now().Unix()
	if _, err := tx.ExecContext(ctx, `
insert into log_shop (uid, shopid, count, price, time)
values (?, ?, ?, ?, ?)`, uid, item.ID, count, item.Price, now); err != nil {
		return err
	}

	return nil
}

func (r *Repository) buyBattleShopItemTx(ctx context.Context, tx *sql.Tx, uid int, item shopPurchaseItem, count int) error {
	if item.TotalCount == 0 {
		return newInvalidError("商品已售罄。")
	}
	if item.BattleGoods != 0 {
		return newInvalidError("当前战场货架暂不支持该商品类型。")
	}

	totalBought, _, err := r.shopPurchaseCountsTx(ctx, tx, uid, item.ID)
	if err != nil {
		return err
	}
	if item.UserLimit > 0 && totalBought+count > item.UserLimit {
		return newInvalidError("超过商品总限购。")
	}
	if item.BattleDayLimit > 0 && totalBought+count > item.BattleDayLimit {
		return newInvalidError("超过战场货架限购。")
	}

	creditNeed := item.CreditPrice * int64(count)
	medalNeed := item.MedalPrice * int64(count)

	if creditNeed > 0 {
		result, err := tx.ExecContext(ctx, `
update sys_user
set honour = honour - ?
where uid = ? and honour >= ?`, creditNeed, uid, creditNeed)
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return newInvalidError("荣誉不足，无法兑换。")
		}
	}

	if medalNeed > 0 {
		result, err := tx.ExecContext(ctx, `
update sys_things
set count = count - ?
where uid = ? and tid = ? and count >= ?`, medalNeed, uid, item.MedalTypeID, medalNeed)
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return newInvalidError("勋章不足，无法兑换。")
		}
	}

	if err := r.addUserGoodsTx(ctx, tx, uid, item.GID, int64(item.Pack*count)); err != nil {
		return err
	}
	if err := r.bumpShopBuyCountTx(ctx, tx, uid, item.ID, count); err != nil {
		return err
	}

	return nil
}

func (r *Repository) shopPurchaseCountsTx(ctx context.Context, tx *sql.Tx, uid int, itemID int) (int, int, error) {
	var totalBought int
	if err := tx.QueryRowContext(ctx, `
select coalesce(count, 0)
from log_shop_buy_cnt
where uid = ? and sid = ?`, uid, itemID).Scan(&totalBought); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}

	var todayBought int
	if err := tx.QueryRowContext(ctx, `
select coalesce(sum(count), 0)
from log_shop
where uid = ? and shopid = ? and date(from_unixtime(time)) = curdate()`, uid, itemID).Scan(&todayBought); err != nil {
		return 0, 0, err
	}

	return totalBought, todayBought, nil
}

func (r *Repository) ensureGiftWalletTx(ctx context.Context, tx *sql.Tx, uid int) error {
	_, err := tx.ExecContext(ctx, `
insert into sys_goods (uid, gid, count)
values (?, 0, 0)
on duplicate key update count = count`, uid)
	return err
}

func (r *Repository) addUserGoodsTx(ctx context.Context, tx *sql.Tx, uid int, gid int, count int64) error {
	if count <= 0 {
		return nil
	}

	_, err := tx.ExecContext(ctx, `
insert into sys_goods (uid, gid, count)
values (?, ?, ?)
on duplicate key update count = count + values(count)`, uid, gid, count)
	return err
}

func (r *Repository) bumpShopBuyCountTx(ctx context.Context, tx *sql.Tx, uid int, itemID int, count int) error {
	_, err := tx.ExecContext(ctx, `
insert into log_shop_buy_cnt (uid, sid, count)
values (?, ?, ?)
on duplicate key update count = count + values(count)`, uid, itemID, count)
	return err
}
