package legacy

import (
	"context"
	"database/sql"
	"html"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var battleNewsTagPattern = regexp.MustCompile(`(?s)<[^>]*>`)

func (r *Repository) BattleFieldState(ctx context.Context, uid int, battlefieldID int, unionID int, cid int, fieldName string) (BattleFieldState, error) {
	state := BattleFieldState{
		FieldName:     strings.TrimSpace(fieldName),
		CID:           cid,
		BattlefieldID: battlefieldID,
		BID:           battlefieldID,
		UnionID:       unionID,
		CanSend:       true,
		News:          []BattleFieldNewsItem{},
		Rows:          []BattleFieldTroopRow{},
		Cities:        []BattleFieldCityInfo{},
		CurrentTroops: []BattleFieldCurrentTroop{},
		WinPoints:     []BattleFieldWinPoint{},
	}
	if state.FieldName == "" {
		state.FieldName = formatCIDLabel(cid)
	}
	if battlefieldID <= 0 || cid <= 0 {
		return state, nil
	}
	if r.db == nil {
		return state, nil
	}

	resolvedBattlefieldID, resolvedBID, resolvedUnionID, err := r.resolveBattleFieldIDs(ctx, uid, battlefieldID, unionID)
	if err != nil {
		return BattleFieldState{}, err
	}
	battlefieldID = resolvedBattlefieldID
	unionID = resolvedUnionID
	state.BattlefieldID = battlefieldID
	state.BID = resolvedBID
	state.UnionID = unionID

	info, err := r.battleFieldInfo(ctx, uid, battlefieldID, unionID, state.FieldName)
	if err != nil {
		return BattleFieldState{}, err
	}
	state.Info = info
	state.FieldName = firstNonEmpty(info.Name, state.FieldName)

	news, total, err := r.battleFieldNews(ctx, battlefieldID, unionID, 0, 10)
	if err != nil {
		return BattleFieldState{}, err
	}
	state.News = news
	state.NewsTotal = total

	currentTroops, currentTroopCID, err := r.battleFieldCurrentTroops(ctx, uid, battlefieldID)
	if err != nil {
		return BattleFieldState{}, err
	}
	state.CurrentTroops = currentTroops

	rows, err := r.battleFieldTroopRows(ctx, battlefieldID, cid)
	if err != nil {
		return BattleFieldState{}, err
	}
	for i := range rows {
		applyBattleFieldPermissions(&rows[i], unionID, cid, currentTroopCID, &state.CanSend)
	}
	state.Rows = rows

	cities, err := r.battleFieldCities(ctx, battlefieldID, unionID)
	if err != nil {
		return BattleFieldState{}, err
	}
	state.Cities = cities

	if state.BID == 2001 {
		winPoints, err := r.battleFieldWinPoints(ctx, battlefieldID)
		if err != nil {
			return BattleFieldState{}, err
		}
		if len(winPoints) == 0 && battlefieldID == state.BID {
			winPoints, err = r.battleFieldSampleWinPoints(ctx, state.BID)
			if err != nil {
				return BattleFieldState{}, err
			}
		}
		state.WinPoints = winPoints
	}

	return state, nil
}

func (r *Repository) resolveBattleFieldIDs(ctx context.Context, uid int, requestedID int, unionID int) (int, int, int, error) {
	battlefieldID := requestedID
	bid := requestedID
	resolvedUnionID := unionID

	var fieldBID int
	var fieldUnionID int
	err := r.db.QueryRowContext(ctx, `
select
	coalesce(f.bid, 0),
	coalesce(s.unionid, 0)
from sys_user_battle_field f
left join sys_user_battle_state s on s.uid = ? and s.battlefieldid = f.id
where f.id = ?
limit 1`, uid, requestedID).Scan(&fieldBID, &fieldUnionID)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, 0, err
	}
	if fieldBID > 0 {
		bid = fieldBID
		if fieldUnionID > 0 && resolvedUnionID <= 0 {
			resolvedUnionID = fieldUnionID
		}
		return battlefieldID, bid, resolvedUnionID, nil
	}

	var currentBattlefieldID int
	var currentBID int
	var currentUnionID int
	err = r.db.QueryRowContext(ctx, `
select
	coalesce(s.battlefieldid, 0),
	coalesce(s.bid, 0),
	coalesce(s.unionid, 0)
from sys_user_battle_state s
where s.uid = ?
	and (? <= 0 or s.bid = ? or s.battlefieldid = ?)
limit 1`, uid, requestedID, requestedID, requestedID).Scan(&currentBattlefieldID, &currentBID, &currentUnionID)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, 0, err
	}
	if currentBattlefieldID > 0 {
		battlefieldID = currentBattlefieldID
	}
	if currentBID > 0 {
		bid = currentBID
	}
	if currentUnionID > 0 {
		resolvedUnionID = currentUnionID
	}

	return battlefieldID, bid, resolvedUnionID, nil
}

func (r *Repository) battleFieldInfo(ctx context.Context, uid int, battlefieldID int, unionID int, fallbackName string) (BattleFieldInfo, error) {
	info := BattleFieldInfo{
		Name:        strings.TrimSpace(fallbackName),
		BID:         battlefieldID,
		MinPeople:   1,
		MaxPeople:   5,
		MaxLevel:    10,
		Level:       10,
		Content:     "",
		PeopleTotal: 0,
	}
	if battlefieldID <= 0 {
		return info, nil
	}

	var battleBID int
	var battleLevel int
	if err := r.db.QueryRowContext(ctx, `
select
	coalesce(f.bid, 0),
	coalesce(f.level, 0),
	coalesce(f.state, 0),
	coalesce(f.starttime, 0),
	coalesce(f.endtime, 0),
	coalesce(f.winner, -1)
from sys_user_battle_field f
where f.id = ?`, battlefieldID).Scan(&battleBID, &battleLevel, &info.State, &info.StartTime, &info.EndTime, &info.Winner); err != nil && err != sql.ErrNoRows {
		return BattleFieldInfo{}, err
	}
	if battleBID > 0 {
		info.BID = battleBID
	}
	if battleLevel > 0 {
		info.Level = battleLevel
	}

	var name sql.NullString
	var content sql.NullString
	if err := r.db.QueryRowContext(ctx, `
select
	coalesce(name, ''),
	coalesce(minpeople, 0),
	coalesce(maxpeople, 0),
	coalesce(maxlevel, 0),
	coalesce(content, '')
from cfg_battle_field
where id = ?`, info.BID).Scan(&name, &info.MinPeople, &info.MaxPeople, &info.MaxLevel, &content); err != nil {
		if err != sql.ErrNoRows {
			return BattleFieldInfo{}, err
		}
	} else {
		info.Name = firstNonEmpty(strings.TrimSpace(name.String), info.Name)
		info.Content = strings.TrimSpace(content.String)
	}
	if info.Name == "" {
		info.Name = formatCIDLabel(battlefieldID)
	}
	if info.MinPeople <= 0 {
		info.MinPeople = 1
	}
	if info.MaxPeople <= 0 {
		info.MaxPeople = 5
	}
	if info.MaxLevel <= 0 {
		info.MaxLevel = info.Level
	}
	if info.Level <= 0 {
		info.Level = info.MaxLevel
	}
	info.Image = "battle_gdzz.jpg"

	if err := r.db.QueryRowContext(ctx, `select count(*) from sys_user_battle_state where battlefieldid = ?`, battlefieldID).Scan(&info.PeopleTotal); err != nil {
		return BattleFieldInfo{}, err
	}
	if info.PeopleTotal == 0 && uid > 0 {
		var currentBattleID int
		if err := r.db.QueryRowContext(ctx, `
select coalesce(battlefieldid, 0)
from sys_user_battle_state
where uid = ? and unionid = ?`, uid, unionID).Scan(&currentBattleID); err != nil && err != sql.ErrNoRows {
			return BattleFieldInfo{}, err
		} else if currentBattleID > 0 && currentBattleID != battlefieldID {
			_ = r.db.QueryRowContext(ctx, `select count(*) from sys_user_battle_state where battlefieldid = ?`, currentBattleID).Scan(&info.PeopleTotal)
		}
	}

	return info, nil
}

func (r *Repository) battleFieldNews(ctx context.Context, battlefieldID int, unionID int, page int, pageSize int) ([]BattleFieldNewsItem, int, error) {
	if battlefieldID <= 0 {
		return []BattleFieldNewsItem{}, 0, nil
	}
	if page < 0 {
		page = 0
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `select count(*) from log_battle_news where battleid = ?`, battlefieldID).Scan(&total); err != nil {
		if isMissingLegacyTableError(err) {
			return []BattleFieldNewsItem{}, 0, nil
		}
		return nil, 0, err
	}
	if total == 0 {
		return []BattleFieldNewsItem{}, 0, nil
	}

	rows, err := r.db.QueryContext(ctx, `
select
	coalesce(battleid, 0),
	coalesce(unionid, 0),
	coalesce(content, ''),
	coalesce(log_time, 0)
from log_battle_news
where battleid = ?
order by log_time desc
limit ?, ?`, battlefieldID, page*pageSize, pageSize)
	if err != nil {
		if isMissingLegacyTableError(err) {
			return []BattleFieldNewsItem{}, 0, nil
		}
		return nil, 0, err
	}
	defer rows.Close()

	items := []BattleFieldNewsItem{}
	for rows.Next() {
		item := BattleFieldNewsItem{}
		if err := rows.Scan(&item.BattleID, &item.UnionID, &item.Content, &item.LogTime); err != nil {
			return nil, 0, err
		}
		item.ID = len(items) + 1 + page*pageSize
		item.Content = removeBattleNewsTags(strings.TrimSpace(item.Content))
		item.Time = formatBattleNewsTime(item.LogTime)
		item.OwnUnion = unionID > 0 && item.UnionID == unionID
		item.Color = 16451364
		if item.OwnUnion {
			item.Color = 8161263
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *Repository) BattleFieldNewsPage(ctx context.Context, uid int, battlefieldID int, unionID int, page int, pageSize int) (BattleFieldNewsPage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}
	if battlefieldID <= 0 || unionID <= 0 {
		resolvedBattlefieldID, _, resolvedUnionID, err := r.resolveBattleFieldIDs(ctx, uid, battlefieldID, unionID)
		if err != nil {
			return BattleFieldNewsPage{}, err
		}
		if battlefieldID <= 0 {
			battlefieldID = resolvedBattlefieldID
		}
		if unionID <= 0 {
			unionID = resolvedUnionID
		}
	}

	items, total, err := r.battleFieldNews(ctx, battlefieldID, unionID, page-1, pageSize)
	if err != nil {
		return BattleFieldNewsPage{}, err
	}
	pageCount := 1
	if total > 0 {
		pageCount = (total + pageSize - 1) / pageSize
	}
	return BattleFieldNewsPage{
		Page:      page,
		PageSize:  pageSize,
		Total:     total,
		PageCount: pageCount,
		Items:     items,
		ReadOnly:  true,
		Message:   "战场讯息为旧服只读数据；分页已按旧服 getBattleNews 读取，不会写入。",
	}, nil
}

func (r *Repository) BattleQuitPreview(ctx context.Context, uid int) (BattleQuitPreview, error) {
	preview := BattleQuitPreview{
		Result:   0,
		Message:  "战场退出写接口未接入；当前仅显示旧服退出结果预览，不会退出战场。",
		ReadOnly: true,
	}
	if uid <= 0 || r.db == nil {
		return preview, nil
	}

	var winner int
	var state int
	var unionID int
	var level int
	err := r.db.QueryRowContext(ctx, `
select
	coalesce(f.winner, 0),
	coalesce(f.state, 0),
	coalesce(s.unionid, 0),
	coalesce(f.level, 0)
from sys_user_battle_state s
left join sys_user_battle_field f on s.battlefieldid = f.id
where s.uid = ?`, uid).Scan(&winner, &state, &unionID, &level)
	if err != nil {
		if err == sql.ErrNoRows {
			preview.Message = "你已经不在战场中；战场退出写接口未接入。"
			return preview, nil
		}
		return BattleQuitPreview{}, err
	}

	switch {
	case state == 2:
		preview.Result = 2
		preview.Message = "战场尚未结束，退出将按旧服规则结算；当前写接口未接入，不会退出战场。"
	case winner == unionID:
		preview.Result = 1
		preview.Message = "旧服预览：当前阵营为胜利方；退出写接口未接入，不会退出战场。"
	case winner == -1:
		preview.Result = -1
		preview.HonourDelta = 0 - level*level*20
		preview.Message = "旧服预览：中途退出可能扣除荣誉 " + strconv.Itoa(-preview.HonourDelta) + "；退出写接口未接入，不会退出战场。"
	default:
		preview.Result = 0
		preview.Message = "旧服预览：当前阵营不是胜利方；退出写接口未接入，不会退出战场。"
	}

	return preview, nil
}

func (r *Repository) BattleTroopDetail(ctx context.Context, uid int, troopID int) (BattleTroopDetail, error) {
	if troopID <= 0 {
		return BattleTroopDetail{}, newInvalidError("invalid battle troop id")
	}
	if r.db == nil {
		return BattleTroopDetail{ID: troopID, ReadOnly: true}, nil
	}
	return r.battleTroopDetail(ctx, troopID)
}

func (r *Repository) BattleArmySendPreview(ctx context.Context, uid int, troopID int, targetCID int, targetName string) (BattleArmySendPreview, error) {
	if troopID <= 0 {
		return BattleArmySendPreview{}, newInvalidError("invalid battle troop id")
	}
	if targetCID <= 0 {
		return BattleArmySendPreview{}, newInvalidError("invalid battle target id")
	}
	troop, err := r.battleTroopDetail(ctx, troopID)
	if err != nil {
		return BattleArmySendPreview{}, err
	}
	if troop.UID != uid {
		return BattleArmySendPreview{}, newForbiddenError("只能查看自己的战场部队派遣预览。")
	}
	if troop.State != 4 {
		return BattleArmySendPreview{}, newInvalidError("部队未驻扎，不能派遣。")
	}
	if troop.CID == targetCID {
		return BattleArmySendPreview{}, newInvalidError("目标据点不能与当前据点相同。")
	}

	originCity, err := r.battleCitySnapshot(ctx, troop.CID)
	if err != nil {
		return BattleArmySendPreview{}, err
	}
	targetCity, err := r.battleCitySnapshot(ctx, targetCID)
	if err != nil {
		return BattleArmySendPreview{}, err
	}
	if targetCity.BattlefieldID != 0 && originCity.BattlefieldID != 0 && targetCity.BattlefieldID != originCity.BattlefieldID {
		return BattleArmySendPreview{}, newInvalidError("只能前往相邻的据点。")
	}
	if err := validateBattleCanGoto(targetCID, originCity.NextXY); err != nil {
		return BattleArmySendPreview{}, err
	}

	target := strings.TrimSpace(targetName)
	if target == "" {
		target = firstNonEmpty(targetCity.Name, r.battleCityName(ctx, targetCID))
	}
	arrival := int64(0)
	if troop.PathTime > 0 {
		arrival = time.Now().Unix() + int64(troop.PathTime)
	}
	return BattleArmySendPreview{
		Troop:    troop,
		TargetID: targetCID,
		Target:   target,
		PathTime: troop.PathTime,
		Arrival:  arrival,
		ReadOnly: true,
		Message:  "只读预览：派遣写接口未接入，不会移动部队。",
	}, nil
}

func (r *Repository) BattleCampaignPreview(ctx context.Context, uid int, cid int, targetCID int, heroID int, soldiers map[int]int64, useFlag bool) (BattleCampaignPreview, error) {
	preview := BattleCampaignPreview{
		CID:       cid,
		TargetCID: targetCID,
		HeroID:    heroID,
		UseFlag:   useFlag,
		ReadOnly:  true,
		Message:   "只读预览：已按旧服战场出征规则计算耗粮和到达时间；出征写接口未接入，不会扣粮、扣兵或派发部队。",
	}
	if cid <= 0 {
		preview.Blocked = true
		preview.Reason = "当前城池无效。"
		return preview, nil
	}
	if r.db == nil {
		preview.Blocked = true
		preview.Reason = "战场出征接口未接入。"
		preview.FieldName = "战场目标"
		preview.Target = formatCIDLabel(targetCID)
		return preview, nil
	}
	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return BattleCampaignPreview{}, err
	} else if !allowed {
		return BattleCampaignPreview{}, ErrForbidden
	}

	normalized, _, err := normalizeTroopSoldiers(soldiers)
	if err != nil {
		preview.Blocked = true
		preview.Reason = "请选择出征士兵。"
	} else {
		preview.Soldiers, preview.SoldierCount = parseTroopSoldiers(formatTroopSoldiersRaw(normalized), r.loadSoldierNames(ctx))
		if len(normalized) == 1 && normalized[scoutSoldierSID] > 0 {
			preview.Blocked = true
			preview.Reason = "斥候不能单独出征。"
		}
	}
	if heroID <= 0 && preview.Reason == "" {
		preview.Blocked = true
		preview.Reason = "请选择将领后执行战场出征。"
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return BattleCampaignPreview{}, err
	}
	defer tx.Rollback()

	field, err := r.battleUserFieldSnapshot(ctx, uid)
	if err != nil {
		return BattleCampaignPreview{}, err
	}
	if field.BattlefieldID <= 0 && preview.Reason == "" {
		preview.Blocked = true
		preview.Reason = "你尚未加入战场。"
	}

	if targetCID <= 0 && preview.Reason == "" {
		preview.Blocked = true
		preview.Reason = "请选择战场出征目标。"
	}
	if targetCID > 0 {
		target, err := r.battleCampaignTarget(ctx, uid, field, targetCID)
		if err != nil {
			if preview.Reason == "" {
				preview.Blocked = true
				preview.Reason = err.Error()
			}
		} else {
			preview.Target = target.Name
		}
	} else {
		preview.Target = "战场目标"
	}
	preview.FieldName = firstNonEmpty(preview.Target, "战场目标")

	groundLevel, err := r.cityBuildingLevel(ctx, tx, cid, troopGroundBuildingID)
	if err != nil {
		return BattleCampaignPreview{}, err
	}
	preview.GroundLevel = groundLevel
	if groundLevel <= 0 && preview.Reason == "" {
		preview.Blocked = true
		preview.Reason = "需要校场后才能派遣部队。"
	}

	activeTroops, err := r.cityActiveTroopCount(ctx, tx, uid, cid)
	if err != nil {
		return BattleCampaignPreview{}, err
	}
	preview.CurrentCityTroops = activeTroops
	if groundLevel > 0 && activeTroops >= groundLevel && preview.Reason == "" {
		preview.Blocked = true
		preview.Reason = "校场容量不足，无法继续派遣。"
	}

	battleTroops, err := r.battleCampaignTroopCount(ctx, uid, field.BattlefieldID)
	if err != nil {
		return BattleCampaignPreview{}, err
	}
	preview.CurrentBattleTroops = battleTroops
	if battleTroops >= 2 && preview.Reason == "" {
		preview.Blocked = true
		preview.Reason = "每个战场最多只能派遣 2 支部队。"
	}

	if useFlag {
		flagCount, err := r.userGoodsCount(ctx, uid, 59)
		if err != nil {
			return BattleCampaignPreview{}, err
		}
		preview.FlagCount = flagCount
		if flagCount <= 0 && preview.Reason == "" {
			preview.Blocked = true
			preview.Reason = "没有军旗，不能使用军旗加成。"
		}
	}

	if heroID > 0 && preview.Reason == "" {
		if err := r.ensureDispatchHeroTx(ctx, tx, uid, cid, heroID); err != nil {
			preview.Blocked = true
			preview.Reason = err.Error()
		}
	}

	if len(normalized) > 0 {
		for sid, count := range normalized {
			available, err := r.citySoldierCount(ctx, tx, cid, sid)
			if err != nil {
				return BattleCampaignPreview{}, err
			}
			if available < count && preview.Reason == "" {
				preview.Blocked = true
				preview.Reason = "兵力不足，无法派遣。"
			}
		}

		stats, pathTime, err := r.battleCampaignSoldierStatsTx(ctx, tx, cid, heroID, normalized)
		if err != nil {
			return BattleCampaignPreview{}, err
		}
		preview.People = stats.People
		preview.PathTime = pathTime
		if pathTime > 0 {
			preview.Arrival = time.Now().Unix() + int64(pathTime)
		}
		foodRate, err := r.battleCampaignFoodRate(ctx, field.BattlefieldID)
		if err != nil {
			return BattleCampaignPreview{}, err
		}
		preview.FoodUse = int64(math.Ceil(stats.FoodUse * float64(foodRate)))
		preview.Capacity = int64(groundLevel) * 100000
		if useFlag {
			preview.Capacity = int64(math.Ceil(float64(preview.Capacity) * 1.25))
		}
		if preview.Capacity > 0 && preview.SoldierCount > preview.Capacity && preview.Reason == "" {
			preview.Blocked = true
			preview.Reason = "校场等级不足，最多可派遣 " + strconv.FormatInt(preview.Capacity, 10) + " 人。"
		}
		food, err := r.battleCampaignCityFood(ctx, tx, cid)
		if err != nil {
			return BattleCampaignPreview{}, err
		}
		if food < preview.FoodUse && preview.Reason == "" {
			preview.Blocked = true
			preview.Reason = "粮食不足，无法战场出征。"
		}
	}

	if preview.Reason == "" {
		preview.Reason = "战场出征写接口未接入。"
	}
	preview.Blocked = true
	return preview, nil
}

func (r *Repository) BattleArmyAttackPreview(ctx context.Context, uid int, troopID int, targetTroopID int, targetName string) (BattleArmyAttackPreview, error) {
	if troopID <= 0 || targetTroopID <= 0 {
		return BattleArmyAttackPreview{}, newInvalidError("invalid battle troop id")
	}
	troop, err := r.battleTroopDetail(ctx, troopID)
	if err != nil {
		return BattleArmyAttackPreview{}, err
	}
	if troop.UID != uid {
		return BattleArmyAttackPreview{}, newForbiddenError("只能查看自己的战场部队攻击预览。")
	}
	if troop.State != 4 {
		return BattleArmyAttackPreview{}, newInvalidError("部队未驻扎，不能攻击。")
	}
	target, err := r.battleTroopDetail(ctx, targetTroopID)
	if err != nil {
		return BattleArmyAttackPreview{}, err
	}
	userField, err := r.battleUserFieldSnapshot(ctx, uid)
	if err != nil {
		return BattleArmyAttackPreview{}, err
	}
	if target.BattleUnionID == userField.UnionID && userField.UnionID > 0 {
		return BattleArmyAttackPreview{}, newInvalidError("同一阵营不能攻击。")
	}

	pathTime := troop.PathTime
	if troop.CID == target.CID {
		if target.State != 4 {
			return BattleArmyAttackPreview{}, newInvalidError("同据点目标未驻扎，不能攻击。")
		}
		pathTime = 10
	} else {
		originCity, err := r.battleCitySnapshot(ctx, troop.CID)
		if err != nil {
			return BattleArmyAttackPreview{}, err
		}
		targetCity, err := r.battleCitySnapshot(ctx, target.CID)
		if err != nil {
			return BattleArmyAttackPreview{}, err
		}
		if err := r.validateBattleAttackSpecialGuards(ctx, userField, targetCity.CID); err != nil {
			return BattleArmyAttackPreview{}, err
		}
		if err := validateBattleCanGoto(targetCity.CID, originCity.NextXY); err != nil {
			return BattleArmyAttackPreview{}, err
		}
		target.SoldiersRaw = ""
		target.Soldiers = []Soldier{}
		target.SoldierCount = 0
	}

	targetLabel := strings.TrimSpace(targetName)
	if targetLabel == "" {
		targetLabel = firstNonEmpty(target.Name, target.TargetName, r.battleCityName(ctx, target.CID))
	}
	arrival := int64(0)
	if pathTime > 0 {
		arrival = time.Now().Unix() + int64(pathTime)
	}
	return BattleArmyAttackPreview{
		Troop:      troop,
		Target:     target,
		TargetID:   targetTroopID,
		TargetName: targetLabel,
		PathTime:   pathTime,
		Arrival:    arrival,
		ReadOnly:   true,
		Message:    "只读预览：攻击写接口未接入，不会进入交战。",
	}, nil
}

func (r *Repository) BattlePatrolPreview(ctx context.Context, uid int, troopID int, targetTroopID int) (BattlePatrolPreview, error) {
	if troopID <= 0 || targetTroopID <= 0 {
		return BattlePatrolPreview{}, newInvalidError("invalid battle troop id")
	}
	troop, err := r.battleTroopDetail(ctx, troopID)
	if err != nil {
		return BattlePatrolPreview{}, err
	}
	if troop.UID != uid {
		return BattlePatrolPreview{}, newForbiddenError("只能查看自己的战场部队侦察预览。")
	}
	target, err := r.battleTroopDetail(ctx, targetTroopID)
	if err != nil {
		return BattlePatrolPreview{}, err
	}
	targetCity, err := r.battleCitySnapshot(ctx, target.CID)
	if err != nil {
		return BattlePatrolPreview{}, err
	}
	pigeonCount, err := r.userGoodsCount(ctx, uid, 140)
	if err != nil {
		return BattlePatrolPreview{}, err
	}

	lines := []string{
		"战场侦察结果：",
		"地点 " + firstNonEmpty(targetCity.Name, r.battleCityName(ctx, target.CID)),
		"敌将 " + firstNonEmpty(target.Hero, "--"),
		"等级 " + strconv.Itoa(target.Level),
	}
	for _, soldier := range target.Soldiers {
		lines = append(lines, soldier.Name+" "+strconv.FormatInt(soldier.Count, 10))
	}

	blocked := pigeonCount <= 0
	message := "只读预览：侦察写接口未接入，不会扣除信鸽，也不会写入战报。"
	if blocked {
		message = "你没有足够的信鸽用来侦察，请到商城购买。只读预览未扣除信鸽。"
	}

	return BattlePatrolPreview{
		Troop:       troop,
		Target:      target,
		TargetID:    targetTroopID,
		TargetName:  firstNonEmpty(target.Name, target.TargetName, r.battleCityName(ctx, target.CID)),
		TargetCity:  firstNonEmpty(targetCity.Name, r.battleCityName(ctx, target.CID)),
		Report:      strings.Join(lines, "\n"),
		ReportLines: lines,
		PigeonCount: pigeonCount,
		Blocked:     blocked,
		ReadOnly:    true,
		Message:     message,
	}, nil
}

func (r *Repository) BattleMembersSnapshot(ctx context.Context, uid int) (BattleMembersSnapshot, error) {
	snapshot := BattleMembersSnapshot{
		Rows:     []BattleMemberRow{},
		ReadOnly: true,
		Message:  "只读预览：已读取旧服战场成员和邀请列表；成员邀请/取消写接口未接入。",
	}
	if uid <= 0 || r.db == nil {
		return snapshot, nil
	}

	var battlefieldID int
	var createUID int
	err := r.db.QueryRowContext(ctx, `
select
	coalesce(s.battlefieldid, 0),
	coalesce(f.createuid, 0)
from sys_user_battle_state s
left join sys_user_battle_field f on s.battlefieldid = f.id
where s.uid = ?
limit 1`, uid).Scan(&battlefieldID, &createUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return snapshot, nil
		}
		return BattleMembersSnapshot{}, err
	}
	if battlefieldID <= 0 {
		return snapshot, nil
	}
	snapshot.IsCreator = uid == createUID

	rows, err := r.db.QueryContext(ctx, `
select
	s.uid,
	coalesce(u.honour, 0),
	coalesce(u.name, ''),
	coalesce(c.name, ''),
	coalesce((select count(*) from sys_troops t where t.uid = s.uid and t.battlefieldid = s.battlefieldid), 0)
from sys_user_battle_state s
left join sys_user u on s.uid = u.uid
left join cfg_battle_union c on c.bid = s.bid and c.unionid = s.unionid
where s.battlefieldid = ?
order by s.jointime asc, s.uid asc`, battlefieldID)
	if err != nil {
		return BattleMembersSnapshot{}, err
	}
	defer rows.Close()

	for rows.Next() {
		item := BattleMemberRow{State: "已加入"}
		var name sql.NullString
		var camp sql.NullString
		if err := rows.Scan(&item.UID, &item.Honour, &name, &camp, &item.HeroCount); err != nil {
			return BattleMembersSnapshot{}, err
		}
		item.ID = item.UID
		item.Name = firstNonEmpty(strings.TrimSpace(name.String), formatUIDLabel(item.UID))
		item.Camp = strings.TrimSpace(camp.String)
		snapshot.Rows = append(snapshot.Rows, item)
	}
	if err := rows.Err(); err != nil {
		return BattleMembersSnapshot{}, err
	}
	snapshot.InCount = len(snapshot.Rows)

	invites, err := r.db.QueryContext(ctx, `
select
	coalesce(i.id, 0),
	coalesce(i.touid, 0),
	coalesce(i.toname, ''),
	coalesce(u.honour, 0)
from sys_battle_invite i
left join sys_user u on i.touid = u.uid
where i.battlefieldid = ?
order by i.time desc, i.id desc`, battlefieldID)
	if err != nil {
		if isMissingLegacyTableError(err) {
			return snapshot, nil
		}
		return BattleMembersSnapshot{}, err
	}
	defer invites.Close()

	for invites.Next() {
		item := BattleMemberRow{
			State:   "邀请中",
			Cancel:  snapshot.IsCreator,
			Invited: true,
		}
		var name sql.NullString
		if err := invites.Scan(&item.InviteID, &item.UID, &name, &item.Honour); err != nil {
			return BattleMembersSnapshot{}, err
		}
		item.ID = item.InviteID
		item.Name = firstNonEmpty(strings.TrimSpace(name.String), formatUIDLabel(item.UID))
		snapshot.Rows = append(snapshot.Rows, item)
	}
	if err := invites.Err(); err != nil {
		return BattleMembersSnapshot{}, err
	}

	return snapshot, nil
}

func (r *Repository) battleTroopDetail(ctx context.Context, troopID int) (BattleTroopDetail, error) {
	soldierNames := r.loadSoldierNames(ctx)
	item := BattleTroopDetail{ReadOnly: true}
	var name sql.NullString
	var unionName sql.NullString
	var heroName sql.NullString
	var targetName sql.NullString
	var currentName sql.NullString
	err := r.db.QueryRowContext(ctx, `
select
	s.id,
	s.uid,
	s.cid,
	coalesce(s.targetcid, 0),
	coalesce(s.battleunionid, 0),
	coalesce(s.hid, 0),
	case
		when s.uid <= 897 then coalesce(bu.name, '')
		else coalesce(u2.name, '')
	end as display_name,
	coalesce(bu.name, '') as union_name,
	case
		when s.uid <= 897 then coalesce(bh.name, '')
		else coalesce(ch.name, '')
	end as hero_name,
	case
		when s.uid <= 897 then coalesce(bh.level, 0)
		else coalesce(ch.level, 0)
	end as hero_level,
	coalesce(s.state, 0),
	coalesce(s.pathtime, 0),
	coalesce(s.endtime, 0),
	coalesce(s.soldiers, ''),
	coalesce(bc_target.name, ''),
	coalesce(bc_current.name, '')
from sys_troops s
left join cfg_battle_union bu on s.battleunionid = bu.unionid
left join cfg_battle_hero bh on s.hid = bh.hid
left join sys_city_hero ch on s.hid = ch.hid
left join sys_user u2 on s.uid = u2.uid
left join sys_battle_city bc_target on s.targetcid = bc_target.cid
left join sys_battle_city bc_current on s.cid = bc_current.cid
where s.id = ?`, troopID).Scan(
		&item.ID,
		&item.UID,
		&item.CID,
		&item.TargetCID,
		&item.BattleUnionID,
		&item.HeroID,
		&name,
		&unionName,
		&heroName,
		&item.Level,
		&item.State,
		&item.PathTime,
		&item.EndTime,
		&item.SoldiersRaw,
		&targetName,
		&currentName,
	)
	if err != nil {
		return BattleTroopDetail{}, err
	}

	item.Name = firstNonEmpty(strings.TrimSpace(name.String), formatUIDLabel(item.UID))
	item.Union = strings.TrimSpace(unionName.String)
	item.Hero = firstNonEmpty(strings.TrimSpace(heroName.String), "--")
	item.StateLabel = battleTroopDetailStateLabel(item.State)
	item.TargetName = battleTroopTargetName(item.State, item.TargetCID, item.CID, targetName.String, currentName.String)
	item.Soldiers, item.SoldierCount = parseTroopSoldiers(item.SoldiersRaw, soldierNames)
	if item.EndTime > 0 {
		item.SecondsLeft = maxInt64(0, item.EndTime-time.Now().Unix())
	}
	buffers, err := r.battleTroopBuffers(ctx, troopID)
	if err != nil {
		return BattleTroopDetail{}, err
	}
	item.Buffers = buffers
	return item, nil
}

func (r *Repository) battleTroopBuffers(ctx context.Context, troopID int) ([]BattleTroopBuffer, error) {
	rows, err := r.db.QueryContext(ctx, `
select coalesce(buftype, 0), coalesce(bufparam, 0)
from mem_troops_buffer
where troopid = ?
order by buftype asc`, troopID)
	if err != nil {
		if isMissingLegacyTableError(err) {
			return []BattleTroopBuffer{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	items := []BattleTroopBuffer{}
	for rows.Next() {
		item := BattleTroopBuffer{}
		if err := rows.Scan(&item.BufType, &item.BufParam); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) battleCityName(ctx context.Context, cid int) string {
	if cid <= 0 || r.db == nil {
		return "--"
	}
	var name sql.NullString
	if err := r.db.QueryRowContext(ctx, `select coalesce(name, '') from sys_battle_city where cid = ? limit 1`, cid).Scan(&name); err == nil {
		return firstNonEmpty(strings.TrimSpace(name.String), formatCIDLabel(cid))
	}
	return formatCIDLabel(cid)
}

type battleCitySnapshot struct {
	CID           int
	BattlefieldID int
	UnionID       int
	Name          string
	NextXY        string
}

type battleCampaignTargetSnapshot struct {
	CID  int
	Name string
	XY   int
}

type battleUserFieldSnapshot struct {
	BattlefieldID int
	BID           int
	UnionID       int
}

func (r *Repository) battleCitySnapshot(ctx context.Context, cid int) (battleCitySnapshot, error) {
	if cid <= 0 {
		return battleCitySnapshot{}, newInvalidError("据点不存在")
	}
	item := battleCitySnapshot{}
	var name sql.NullString
	var nextXY sql.NullString
	err := r.db.QueryRowContext(ctx, `
select
	coalesce(cid, 0),
	coalesce(battlefieldid, 0),
	coalesce(unionid, 0),
	coalesce(name, ''),
	coalesce(nextxy, '')
from sys_battle_city
where cid = ?
limit 1`, cid).Scan(&item.CID, &item.BattlefieldID, &item.UnionID, &name, &nextXY)
	if err != nil {
		if err == sql.ErrNoRows {
			return battleCitySnapshot{}, newInvalidError("据点不存在")
		}
		return battleCitySnapshot{}, err
	}
	item.Name = firstNonEmpty(strings.TrimSpace(name.String), formatCIDLabel(item.CID))
	item.NextXY = strings.TrimSpace(nextXY.String)
	return item, nil
}

func (r *Repository) battleCampaignTarget(ctx context.Context, uid int, field battleUserFieldSnapshot, targetCID int) (battleCampaignTargetSnapshot, error) {
	if targetCID <= 0 {
		return battleCampaignTargetSnapshot{}, newInvalidError("请选择战场出征目标。")
	}
	city, err := r.battleCitySnapshot(ctx, targetCID)
	if err != nil {
		return battleCampaignTargetSnapshot{}, err
	}
	if field.BattlefieldID > 0 && city.BattlefieldID > 0 && city.BattlefieldID != field.BattlefieldID {
		return battleCampaignTargetSnapshot{}, newInvalidError("目标不在当前战场。")
	}

	target := battleCampaignTargetSnapshot{
		CID:  city.CID,
		Name: firstNonEmpty(city.Name, formatCIDLabel(city.CID)),
		XY:   battleCityXY(city.CID),
	}
	if field.BID <= 0 || field.UnionID <= 0 || target.XY <= 0 {
		return target, nil
	}

	var needHonour sql.NullInt64
	err = r.db.QueryRowContext(ctx, `
select coalesce(needhonour, 0)
from cfg_battle_start_city
where bid = ? and xy = ? and unionid = ?
limit 1`, field.BID, target.XY, field.UnionID).Scan(&needHonour)
	if err != nil {
		if err == sql.ErrNoRows {
			return battleCampaignTargetSnapshot{}, newInvalidError("该据点不能作为当前阵营的战场出征点。")
		}
		if isMissingLegacyTableError(err) {
			return target, nil
		}
		return battleCampaignTargetSnapshot{}, err
	}
	if needHonour.Int64 <= 0 {
		return target, nil
	}

	var honour sql.NullInt64
	err = r.db.QueryRowContext(ctx, `select coalesce(honour, 0) from sys_user where uid = ?`, uid).Scan(&honour)
	if err != nil && err != sql.ErrNoRows && !isMissingLegacyTableError(err) {
		return battleCampaignTargetSnapshot{}, err
	}
	if honour.Int64 < needHonour.Int64 {
		return battleCampaignTargetSnapshot{}, newInvalidError("战场荣誉不足，不能从该据点出征。")
	}
	return target, nil
}

func (r *Repository) battleUserFieldSnapshot(ctx context.Context, uid int) (battleUserFieldSnapshot, error) {
	if uid <= 0 {
		return battleUserFieldSnapshot{}, nil
	}
	item := battleUserFieldSnapshot{}
	err := r.db.QueryRowContext(ctx, `
select
	coalesce(s.battlefieldid, 0),
	coalesce(f.bid, s.bid, 0),
	coalesce(s.unionid, 0)
from sys_user_battle_state s
left join sys_user_battle_field f on f.id = s.battlefieldid
where s.uid = ?
limit 1`, uid).Scan(&item.BattlefieldID, &item.BID, &item.UnionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return battleUserFieldSnapshot{}, nil
		}
		return battleUserFieldSnapshot{}, err
	}
	return item, nil
}

func (r *Repository) validateBattleAttackSpecialGuards(ctx context.Context, field battleUserFieldSnapshot, targetCID int) error {
	if field.BID != 2001 || field.BattlefieldID <= 0 {
		return nil
	}
	targetXY := battleCityXY(targetCID)
	switch targetXY {
	case 767:
		point, err := r.battleWinPoint(ctx, field.BattlefieldID, 4)
		if err != nil {
			return err
		}
		if point > 0 {
			return newInvalidError("曹操粮草尚未耗尽，不能贸然攻击许都！")
		}
	case 101:
		point, err := r.battleWinPoint(ctx, field.BattlefieldID, 3)
		if err != nil {
			return err
		}
		if point > 0 {
			return newInvalidError("袁绍粮草尚未耗尽，不能贸然攻击邺城！")
		}
	}
	return nil
}

func (r *Repository) battleWinPoint(ctx context.Context, battlefieldID int, unionID int) (int, error) {
	var point int
	err := r.db.QueryRowContext(ctx, `
select coalesce(point, 0)
from sys_battle_winpoint
where battlefieldid = ? and unionid = ?
limit 1`, battlefieldID, unionID).Scan(&point)
	if err != nil {
		if err == sql.ErrNoRows || isMissingLegacyTableError(err) {
			return 0, nil
		}
		return 0, err
	}
	return point, nil
}

func (r *Repository) userGoodsCount(ctx context.Context, uid int, gid int) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `
select coalesce(count, 0)
from sys_goods
where uid = ? and gid = ?
limit 1`, uid, gid).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows || isMissingLegacyTableError(err) {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

func (r *Repository) battleCampaignTroopCount(ctx context.Context, uid int, battlefieldID int) (int, error) {
	if uid <= 0 || battlefieldID <= 0 {
		return 0, nil
	}
	var count int
	err := r.db.QueryRowContext(ctx, `select count(*) from sys_troops where uid = ? and battlefieldid = ?`, uid, battlefieldID).Scan(&count)
	return count, err
}

func (r *Repository) battleCampaignFoodRate(ctx context.Context, battlefieldID int) (int, error) {
	if battlefieldID <= 0 {
		return 1, nil
	}
	var hours int
	err := r.db.QueryRowContext(ctx, `
select ceil((coalesce(endtime, unix_timestamp()) - unix_timestamp()) / 3600)
from sys_user_battle_field
where id = ?`, battlefieldID).Scan(&hours)
	if err != nil {
		if err == sql.ErrNoRows {
			return 1, nil
		}
		return 0, err
	}
	if hours <= 0 {
		hours = 1
	}
	return hours, nil
}

func (r *Repository) battleCampaignCityFood(ctx context.Context, tx *sql.Tx, cid int) (int64, error) {
	var food sql.NullInt64
	err := tx.QueryRowContext(ctx, `select coalesce(food, 0) from mem_city_resource where cid = ?`, cid).Scan(&food)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return food.Int64, nil
}

func (r *Repository) battleCampaignSoldierStatsTx(ctx context.Context, tx *sql.Tx, cid int, heroID int, soldiers map[int]int64) (troopSoldierStats, int, error) {
	stats := troopSoldierStats{}
	minSpeed := 0.0
	infantryRate, err := r.battleCampaignTechnicRateTx(ctx, tx, cid, 12, 0.1)
	if err != nil {
		return troopSoldierStats{}, 0, err
	}
	cavalryRate, err := r.battleCampaignTechnicRateTx(ctx, tx, cid, 13, 0.05)
	if err != nil {
		return troopSoldierStats{}, 0, err
	}
	heroRate, err := r.battleCampaignHeroSpeedRateTx(ctx, tx, heroID)
	if err != nil {
		return troopSoldierStats{}, 0, err
	}

	for sid, count := range soldiers {
		if sid <= 0 || count <= 0 {
			continue
		}
		var peopleNeed sql.NullInt64
		var foodUse sql.NullFloat64
		var speed sql.NullFloat64
		if err := tx.QueryRowContext(ctx, `
select coalesce(people_need, 0), coalesce(food_use, 0), coalesce(speed, 0)
from cfg_soldier
where sid = ?`, sid).Scan(&peopleNeed, &foodUse, &speed); err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return troopSoldierStats{}, 0, err
		}
		stats.People += peopleNeed.Int64 * count
		stats.FoodUse += foodUse.Float64 * float64(count)
		effectiveSpeed := speed.Float64
		if effectiveSpeed > 0 {
			if sid < 7 && sid != scoutSoldierSID {
				effectiveSpeed *= infantryRate
			} else {
				effectiveSpeed *= cavalryRate
			}
			effectiveSpeed *= heroRate
			if minSpeed <= 0 || effectiveSpeed < minSpeed {
				minSpeed = effectiveSpeed
			}
		}
	}
	if minSpeed <= 0 {
		return stats, 0, nil
	}
	return stats, int(math.Floor(162000 / minSpeed)), nil
}

func (r *Repository) battleCampaignTechnicRateTx(ctx context.Context, tx *sql.Tx, cid int, tid int, step float64) (float64, error) {
	var level sql.NullInt64
	err := tx.QueryRowContext(ctx, `select coalesce(level, 0) from sys_city_technic where cid = ? and tid = ?`, cid, tid).Scan(&level)
	if err != nil {
		if err == sql.ErrNoRows {
			return 1, nil
		}
		return 0, err
	}
	return 1 + float64(level.Int64)*step, nil
}

func (r *Repository) battleCampaignHeroSpeedRateTx(ctx context.Context, tx *sql.Tx, heroID int) (float64, error) {
	if heroID <= 0 {
		return 1, nil
	}
	var speedAdd sql.NullFloat64
	err := tx.QueryRowContext(ctx, `select coalesce(speed_add_on, 0) from sys_city_hero where hid = ?`, heroID).Scan(&speedAdd)
	if err != nil {
		if err == sql.ErrNoRows {
			return 1, nil
		}
		return 0, err
	}
	return 1 + speedAdd.Float64*0.01, nil
}

func validateBattleCanGoto(targetCID int, nextXY string) error {
	targetXY := battleCityXY(targetCID)
	if targetXY == 0 {
		return newInvalidError("只能前往相邻的据点。")
	}
	parts := strings.Split(nextXY, ",")
	if len(parts) == 0 {
		return newInvalidError("只能前往相邻的据点。")
	}
	count, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || count <= 0 {
		count = len(parts) - 1
	}
	for i := 0; i < count && i+1 < len(parts); i++ {
		value, err := strconv.Atoi(strings.TrimSpace(parts[i+1]))
		if err != nil {
			continue
		}
		switch value {
		case targetXY:
			return nil
		case -targetXY:
			return newInvalidError("前往该据点的通路尚未打开。")
		}
	}
	return newInvalidError("只能前往相邻的据点。")
}

func battleCityXY(cid int) int {
	if cid < 0 {
		return -((-cid) % 1000)
	}
	return cid % 1000
}

func battleTroopDetailStateLabel(state int) string {
	switch state {
	case 2, 3, 4:
		return troopStateLabel(state)
	default:
		return troopStateLabel(state)
	}
}

func battleTroopTargetName(state int, targetCID int, currentCID int, targetName string, currentName string) string {
	switch state {
	case 0, 1, 2, 3:
		return firstNonEmpty(strings.TrimSpace(targetName), formatCIDLabel(targetCID))
	case 4:
		return firstNonEmpty(strings.TrimSpace(currentName), formatCIDLabel(currentCID))
	default:
		return "--"
	}
}

func (r *Repository) battleFieldTroopRows(ctx context.Context, battlefieldID int, cid int) ([]BattleFieldTroopRow, error) {
	soldierNames := r.loadSoldierNames(ctx)
	rows, err := r.db.QueryContext(ctx, `
select
	s.id,
	s.uid,
	s.cid,
	coalesce(s.targetcid, 0),
	coalesce(s.startcid, 0),
	coalesce(s.battlefieldid, 0),
	coalesce(s.battleunionid, 0),
	coalesce(s.hid, 0),
	case
		when s.uid < 897 then coalesce(bu.name, '')
		else coalesce(u2.name, '')
	end as display_name,
	coalesce(bu.name, '') as union_name,
	case
		when s.uid < 897 then coalesce(bh.name, '')
		else coalesce(ch.name, '')
	end as hero_name,
	case
		when s.uid < 897 then coalesce(bh.level, 0)
		else coalesce(ch.level, 0)
	end as hero_level,
	coalesce(s.state, 0),
	coalesce(s.soldiers, '')
from sys_troops s
left join cfg_battle_union bu on s.battleunionid = bu.unionid
left join cfg_battle_hero bh on s.hid = bh.hid
left join sys_city_hero ch on s.hid = ch.hid
left join sys_user u2 on s.uid = u2.uid
where s.battlefieldid = ?
	and s.cid = ?
	and (
		(s.uid < 897 and (s.state = 4 or s.state = 3))
		or
		(s.uid > 897 and ((s.state = 4 or s.state = 3) or s.targetcid = ?))
	)
order by s.uid asc, s.id asc`, battlefieldID, cid, cid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []BattleFieldTroopRow{}
	for rows.Next() {
		item := BattleFieldTroopRow{}
		var name sql.NullString
		var unionName sql.NullString
		var heroName sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.UID,
			&item.CID,
			&item.TargetCID,
			&item.StartCID,
			&item.BattlefieldID,
			&item.BattleUnionID,
			&item.HeroID,
			&name,
			&unionName,
			&heroName,
			&item.Level,
			&item.State,
			&item.SoldiersRaw,
		); err != nil {
			return nil, err
		}
		item.Name = firstNonEmpty(strings.TrimSpace(name.String), formatUIDLabel(item.UID))
		item.Union = strings.TrimSpace(unionName.String)
		item.Hero = firstNonEmpty(strings.TrimSpace(heroName.String), "--")
		item.StateLabel = troopStateLabel(item.State)
		item.Soldiers, item.SoldierCount = parseTroopSoldiers(item.SoldiersRaw, soldierNames)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *Repository) battleFieldCurrentTroops(ctx context.Context, uid int, battlefieldID int) ([]BattleFieldCurrentTroop, int, error) {
	soldierNames := r.loadSoldierNames(ctx)
	rows, err := r.db.QueryContext(ctx, `
select
	t.id,
	t.cid,
	coalesce(t.soldiers, ''),
	coalesce(h.face, 0),
	coalesce(h.sex, 0),
	coalesce(h.name, ''),
	coalesce(h.level, 0)
from sys_troops t
left join sys_city_hero h on t.hid = h.hid
where t.uid = ? and t.battlefieldid = ?
order by case when t.state = 4 then 0 else 1 end asc, t.id asc`, uid, battlefieldID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []BattleFieldCurrentTroop{}
	currentCID := 0
	for rows.Next() {
		item := BattleFieldCurrentTroop{}
		var heroName sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.CID,
			&item.SoldiersRaw,
			&item.Face,
			&item.Sex,
			&heroName,
			&item.HeroLevel,
		); err != nil {
			return nil, 0, err
		}
		item.HeroName = firstNonEmpty(strings.TrimSpace(heroName.String), "--")
		item.Soldiers, item.SoldierCount = parseTroopSoldiers(item.SoldiersRaw, soldierNames)
		if currentCID == 0 && item.CID > 0 {
			currentCID = item.CID
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, currentCID, nil
}

func (r *Repository) battleFieldCities(ctx context.Context, battlefieldID int, unionID int) ([]BattleFieldCityInfo, error) {
	rows, err := r.db.QueryContext(ctx, `
select
	coalesce(cid, 0),
	coalesce(battlefieldid, 0),
	coalesce(name, ''),
	coalesce(uid, 0),
	coalesce(unionid, 0),
	coalesce(hasuser, 0)
from sys_battle_city
where battlefieldid = ?
order by cid asc`, battlefieldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []BattleFieldCityInfo{}
	for rows.Next() {
		item := BattleFieldCityInfo{}
		var name sql.NullString
		var hasUser int
		if err := rows.Scan(&item.CID, &item.BattlefieldID, &name, &item.UID, &item.UnionID, &hasUser); err != nil {
			return nil, err
		}
		item.Name = firstNonEmpty(strings.TrimSpace(name.String), formatCIDLabel(item.CID))
		item.HasUser = hasUser != 0
		item.Flag, item.FlagChar = battleCityFlag(item.UnionID, unionID, item.HasUser)
		item.FlagLabel = battleCityFlagLabel(item.Flag)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *Repository) battleFieldWinPoints(ctx context.Context, battlefieldID int) ([]BattleFieldWinPoint, error) {
	if battlefieldID <= 0 {
		return []BattleFieldWinPoint{}, nil
	}
	rows, err := r.db.QueryContext(ctx, `
select
	coalesce(battlefieldid, 0),
	coalesce(unionid, 0),
	coalesce(point, 0),
	coalesce(nextreset, 0),
	coalesce(`+"`interval`"+`, 60),
	coalesce(bid, 0),
	coalesce(pointcount, 0),
	coalesce(pointname, ''),
	coalesce(state, 0)
from sys_battle_winpoint
where battlefieldid = ?
order by unionid asc`, battlefieldID)
	if err != nil {
		if isMissingLegacyTableError(err) {
			return []BattleFieldWinPoint{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	items := []BattleFieldWinPoint{}
	for rows.Next() {
		item := BattleFieldWinPoint{}
		var pointName sql.NullString
		if err := rows.Scan(
			&item.BattlefieldID,
			&item.UnionID,
			&item.Point,
			&item.NextReset,
			&item.Interval,
			&item.BID,
			&item.PointCount,
			&pointName,
			&item.State,
		); err != nil {
			return nil, err
		}
		item.PointName = strings.TrimSpace(pointName.String)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *Repository) battleFieldSampleWinPoints(ctx context.Context, bid int) ([]BattleFieldWinPoint, error) {
	if bid <= 0 {
		return []BattleFieldWinPoint{}, nil
	}
	var sampleBattlefieldID int
	err := r.db.QueryRowContext(ctx, `
select coalesce(min(battlefieldid), 0)
from sys_battle_winpoint
where bid = ?`, bid).Scan(&sampleBattlefieldID)
	if err != nil {
		if isMissingLegacyTableError(err) || err == sql.ErrNoRows {
			return []BattleFieldWinPoint{}, nil
		}
		return nil, err
	}
	if sampleBattlefieldID <= 0 {
		return []BattleFieldWinPoint{}, nil
	}
	return r.battleFieldWinPoints(ctx, sampleBattlefieldID)
}

func applyBattleFieldPermissions(row *BattleFieldTroopRow, unionID int, cid int, currentTroopCID int, canSend *bool) {
	if row.BattleUnionID == unionID {
		row.CanAttack = false
		row.CanPatrol = false
	} else {
		row.CanAttack = true
		row.CanPatrol = cid != currentTroopCID
		*canSend = false
	}

	switch {
	case cid == currentTroopCID:
		row.CanView = true
	case row.BattleUnionID != unionID:
		row.CanView = false
	default:
		row.CanView = true
	}
}

func battleCityFlagLabel(flag int) string {
	switch flag {
	case 0:
		return "敌方据点"
	case 1:
		return "我方据点"
	case 2:
		return "我方NPC据点"
	case 3:
		return "争夺中"
	case 4:
		return "敌方玩家据点"
	case 5:
		return "空据点"
	default:
		return ""
	}
}

func battleCityFlag(cityUnionID int, myUnionID int, hasUser bool) (int, string) {
	switch {
	case cityUnionID == -1:
		return 5, ""
	case cityUnionID == 0:
		return 3, ""
	case cityUnionID == myUnionID:
		if hasUser {
			return 1, battleUnionFlagText(myUnionID)
		}
		return 2, battleUnionFlagText(myUnionID)
	default:
		if hasUser {
			return 4, battleUnionFlagText(cityUnionID)
		}
		return 0, battleUnionFlagText(cityUnionID)
	}
}

func battleUnionFlagText(unionID int) string {
	switch unionID {
	case 1:
		return "汉"
	case 2:
		return "黄"
	case 3:
		return "董"
	case 4:
		return "袁"
	case 5:
		return "曹"
	default:
		return ""
	}
}

func formatUIDLabel(uid int) string {
	if uid > 0 {
		return "UID " + strconv.Itoa(uid)
	}
	return "--"
}

func formatBattleNewsTime(unix int64) string {
	if unix <= 0 {
		return ""
	}
	return time.Unix(unix, 0).Format("15:04")
}

func removeBattleNewsTags(value string) string {
	plain := strings.NewReplacer(
		"<br>", " ",
		"<br/>", " ",
		"<br />", " ",
		"</p>", " ",
		"</div>", " ",
		"</tr>", " ",
	).Replace(value)
	plain = battleNewsTagPattern.ReplaceAllString(plain, " ")
	plain = html.UnescapeString(plain)
	return strings.Join(strings.Fields(plain), " ")
}

func isMissingLegacyTableError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "doesn't exist") ||
		strings.Contains(message, "no such table") ||
		strings.Contains(message, "unknown table")
}
