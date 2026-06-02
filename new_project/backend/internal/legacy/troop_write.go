package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	troopGroundBuildingID   = 8
	minimalTroopPathSeconds = 5
	scoutSoldierSID         = 3
)

type troopRecord struct {
	ID        int
	UID       int
	CID       int
	StartCID  int
	TargetCID int
	HID       int
	Task      int
	State     int
	StartTime int64
	PathTime  int
	EndTime   int64
	Soldiers  string
	Resource  string
}

type troopSoldierStats struct {
	People  int64
	FoodUse float64
}

func (r *Repository) DispatchCityTroop(ctx context.Context, uid int, cid int, targetCID int, soldiers map[int]int64, heroID int, task int, resource TroopResource) (TroopPage, error) {
	if targetCID <= 0 || targetCID == cid {
		return TroopPage{}, newInvalidError("目标城池无效。")
	}
	if task < 0 || task > 3 {
		return TroopPage{}, newInvalidError("only task=0, task=1, task=2, task=3 are supported")
	}
	if taskRequiresHero(task) && heroID <= 0 {
		return TroopPage{}, newInvalidError("task=3 requires hero")
	}
	soldierPayload, _, err := normalizeTroopSoldiers(soldiers)
	if err != nil {
		return TroopPage{}, err
	}
	resourcePayload, err := normalizeTroopResource(resource)
	if err != nil {
		return TroopPage{}, err
	}
	if task == 1 && !resourcePayload.isZero() {
		return TroopPage{}, newInvalidError("task=1 不支持携带资源。")
	}

	if task == 2 {
		if soldierPayload[scoutSoldierSID] <= 0 {
			return TroopPage{}, newInvalidError("task=2 requires scout sid=3")
		}
		if !resourcePayload.isZero() {
			return TroopPage{}, newInvalidError("task=2 does not support resources")
		}
	}
	if task == 3 {
		if len(soldierPayload) == 1 && soldierPayload[scoutSoldierSID] > 0 {
			return TroopPage{}, newInvalidError("task=3 does not support scout-only army")
		}
		if !resourcePayload.isZero() {
			return TroopPage{}, newInvalidError("task=3 does not support resources")
		}
	}

	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return TroopPage{}, err
	} else if !allowed {
		return TroopPage{}, ErrForbidden
	}
	targetExists, err := r.cityExists(ctx, targetCID)
	if err != nil {
		return TroopPage{}, err
	}
	if !targetExists {
		return TroopPage{}, newInvalidError("target city does not exist")
	}
	targetOwnedByUser, err := r.UserOwnsCity(ctx, uid, targetCID)
	allowed := targetOwnedByUser
	if task == 2 || task == 3 {
		allowed = !targetOwnedByUser
	}
	if err == nil && !allowed && task == 2 {
		return TroopPage{}, newInvalidError("task=2 must target a non-owned city/field")
	}
	if err == nil && !allowed && task == 3 {
		return TroopPage{}, newInvalidError("task=3 must target a non-owned city")
	}
	if err != nil {
		return TroopPage{}, err
	} else if !allowed {
		return TroopPage{}, newInvalidError("当前最小派遣链只支持自己的城池之间调兵。")
	}

	if r.db == nil {
		return r.fixtureDispatchTroop(uid, cid, targetCID, soldierPayload, heroID, task, resourcePayload)
	}

	if err := r.settleDueTroops(ctx, uid); err != nil {
		return TroopPage{}, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return TroopPage{}, err
	}
	defer tx.Rollback()

	groundLevel, err := r.cityBuildingLevel(ctx, tx, cid, troopGroundBuildingID)
	if err != nil {
		return TroopPage{}, err
	}
	if groundLevel <= 0 {
		return TroopPage{}, newInvalidError("需要校场后才能派遣部队。")
	}

	activeTroops, err := r.cityActiveTroopCount(ctx, tx, uid, cid)
	if err != nil {
		return TroopPage{}, err
	}
	if activeTroops >= groundLevel {
		return TroopPage{}, newInvalidError("校场容量不足，无法继续派遣。")
	}

	if heroID > 0 {
		if err := r.ensureDispatchHeroTx(ctx, tx, uid, cid, heroID); err != nil {
			return TroopPage{}, err
		}
	}
	for soldierSID, soldierCount := range soldierPayload {
		availableSoldiers, err := r.citySoldierCount(ctx, tx, cid, soldierSID)
		if err != nil {
			return TroopPage{}, err
		}
		if availableSoldiers < soldierCount {
			return TroopPage{}, newInvalidError("兵力不足，无法派遣。")
		}
	}
	if task == 3 {
		targetSnapshot, err := r.loadPlunderTargetSnapshotTx(ctx, tx, targetCID)
		if err != nil {
			return TroopPage{}, err
		}
		if !targetSnapshot.minimalEligible() {
			return TroopPage{}, newInvalidError("当前最小 task=3 只支持无兵、无城防、无驻军、无守将的目标城池。")
		}
	}

	for soldierSID, soldierCount := range soldierPayload {
		result, err := tx.ExecContext(ctx, `
update sys_city_soldier
set count = count - ?
where cid = ? and sid = ? and count >= ?`,
			soldierCount,
			cid,
			soldierSID,
			soldierCount,
		)
		if err != nil {
			return TroopPage{}, err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return TroopPage{}, err
		}
		if affected == 0 {
			return TroopPage{}, newInvalidError("兵力不足，无法派遣。")
		}
	}
	if task == 0 && !resourcePayload.isZero() {
		if err := r.subtractCityResourcesTx(ctx, tx, cid, resourcePayload); err != nil {
			return TroopPage{}, err
		}
	}

	if err := r.ensureCityResAdd(ctx, tx, cid); err != nil {
		return TroopPage{}, err
	}
	if _, err := tx.ExecContext(ctx, "update sys_city_res_add set resource_changing = 1 where cid = ?", cid); err != nil {
		return TroopPage{}, err
	}

	now := time.Now().Unix()
	soldiersRaw := formatTroopSoldiersRaw(soldierPayload)
	resourceRaw := formatTroopResourceRaw(resourcePayload)
	soldierStats, err := r.troopSoldierStatsTx(ctx, tx, soldierPayload)
	if err != nil {
		return TroopPage{}, err
	}
	if _, err := tx.ExecContext(ctx, `
insert into sys_troops (uid, cid, startcid, hid, targetcid, task, state, starttime, pathtime, endtime, soldiers, resource, people, fooduse)
values (?, ?, ?, ?, ?, ?, 0, ?, ?, ?, ?, ?, ?, ?)`,
		uid,
		cid,
		cid,
		heroID,
		targetCID,
		task,
		now,
		minimalTroopPathSeconds,
		now+minimalTroopPathSeconds,
		soldiersRaw,
		resourceRaw,
		soldierStats.People,
		soldierStats.FoodUse,
	); err != nil {
		return TroopPage{}, err
	}
	if heroID > 0 {
		if _, err := tx.ExecContext(ctx, "update sys_city_hero set state = 2 where hid = ? and uid = ? and cid = ?", heroID, uid, cid); err != nil {
			return TroopPage{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return TroopPage{}, err
	}

	return r.MyTroops(ctx, uid, 40)
}

func (r *Repository) CallbackTroop(ctx context.Context, uid int, troopID int) (TroopPage, error) {
	if troopID <= 0 {
		return TroopPage{}, newInvalidError("无效的部队编号。")
	}

	if r.db == nil {
		return r.fixtureCallbackTroop(uid, troopID), nil
	}

	if err := r.settleDueTroops(ctx, uid); err != nil {
		return TroopPage{}, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return TroopPage{}, err
	}
	defer tx.Rollback()

	record, err := r.userTroopRecord(ctx, tx, uid, troopID)
	if err != nil {
		if err == sql.ErrNoRows {
			return TroopPage{}, newInvalidError("未找到该部队。")
		}
		return TroopPage{}, err
	}

	if record.State == 1 {
		return TroopPage{}, newInvalidError("部队已经在返程中。")
	}
	if record.State == 3 {
		return TroopPage{}, newInvalidError("战斗中的部队不能撤回。")
	}
	if !canCallbackTroopState(record.State) {
		return TroopPage{}, newInvalidError("当前部队状态不支持撤回。")
	}

	now := time.Now().Unix()
	startCID, targetCID, endTime := callbackTroopReturnRoute(record, now)

	if _, err := tx.ExecContext(ctx, `
update sys_troops
set state = 1, starttime = ?, endtime = ?, startcid = ?, targetcid = ?
where id = ?`,
		now,
		endTime,
		startCID,
		targetCID,
		record.ID,
	); err != nil {
		return TroopPage{}, err
	}

	if callbackTroopMarksCityResources(record.State) {
		if err := r.setTroopHeroStateTx(ctx, tx, uid, record.CID, record.HID, 2); err != nil {
			return TroopPage{}, err
		}
		if err := r.ensureCityResAdd(ctx, tx, record.CID); err != nil {
			return TroopPage{}, err
		}
		if _, err := tx.ExecContext(ctx, "update sys_city_res_add set resource_changing = 1 where cid = ?", record.CID); err != nil {
			return TroopPage{}, err
		}
		if err := r.ensureCityResAdd(ctx, tx, record.TargetCID); err != nil {
			return TroopPage{}, err
		}
		if _, err := tx.ExecContext(ctx, "update sys_city_res_add set resource_changing = 1 where cid = ?", record.TargetCID); err != nil {
			return TroopPage{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return TroopPage{}, err
	}

	return r.MyTroops(ctx, uid, 40)
}

func (r *Repository) settleDueTroops(ctx context.Context, uid int) error {
	if r.db == nil {
		return nil
	}

	query, args := dueTroopsQuery(uid, time.Now().Unix())
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	due := make([]troopRecord, 0, 8)
	for rows.Next() {
		record := troopRecord{}
		if err := rows.Scan(
			&record.ID,
			&record.UID,
			&record.CID,
			&record.StartCID,
			&record.TargetCID,
			&record.HID,
			&record.Task,
			&record.State,
			&record.StartTime,
			&record.PathTime,
			&record.EndTime,
			&record.Soldiers,
			&record.Resource,
		); err != nil {
			return err
		}
		due = append(due, record)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, record := range due {
		if err := r.settleDueTroop(ctx, record); err != nil {
			return err
		}
	}

	return nil
}

func dueTroopsQuery(uid int, now int64) (string, []any) {
	return `
select id, uid, cid, startcid, targetcid, hid, task, state, starttime, pathtime, endtime, soldiers, resource
from sys_troops
where (uid = ? or targetcid in (select cid from sys_city where uid = ?))
  and task in (0, 1, 2, 3, 4) and endtime > 0 and endtime <= ? and state in (0, 1, 5)
order by id asc`,
		[]any{
			uid,
			uid,
			now,
		}
}

func (r *Repository) settleDueTroop(ctx context.Context, record troopRecord) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	current, err := r.userTroopRecord(ctx, tx, record.UID, record.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	if !isSettledTroopTask(current.Task) || !isDueTroopState(current.State) || current.EndTime > time.Now().Unix() {
		return tx.Commit()
	}

	soldierCounts := parseTroopSoldierCounts(current.Soldiers)
	resources := parseTroopResourcePayload(current.Resource)

	if current.Task == 2 && current.State == 0 {
		attackerScouts := soldierCounts[scoutSoldierSID]
		if attackerScouts <= 0 {
			if err := r.restoreTroopHeroStateTx(ctx, tx, current.UID, current.CID, current.HID); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, "delete from sys_troops where id = ?", current.ID); err != nil {
				return err
			}
			return tx.Commit()
		}

		targetSnapshot, targetIsCity, err := r.loadScoutTargetSnapshotTx(ctx, tx, current.TargetCID)
		if err != nil {
			return err
		}
		fromCity := r.cityNameTx(ctx, tx, current.CID)
		targetCity := firstNonEmpty(targetSnapshot.CityName, r.cityNameTx(ctx, tx, current.TargetCID))

		defenderScouts := int64(0)
		if targetIsCity {
			defenderScouts, err = r.citySoldierCount(ctx, tx, current.TargetCID, scoutSoldierSID)
			if err != nil {
				return err
			}
		}

		attackerSurvivors := attackerScouts
		defenderLosses := int64(0)
		if defenderScouts > 0 {
			switch {
			case attackerScouts == defenderScouts*2:
				// Legacy tie branch: both sides lose half of their scouts.
				attackerSurvivors = attackerScouts / 2
				defenderLosses = defenderScouts / 2
			case attackerScouts > defenderScouts*2:
				defenderLosses = defenderScouts
				attackerSurvivors = attackerScouts - defenderScouts*2
			default:
				attackerSurvivors = 0
				// Legacy defender-win branch: defender loses attacker scout count.
				defenderLosses = attackerScouts
			}
		}

		defenderRemaining := defenderScouts - defenderLosses
		if defenderRemaining < 0 {
			defenderRemaining = 0
		}
		if targetIsCity && defenderRemaining != defenderScouts {
			if err := r.setCitySoldierCountTx(ctx, tx, current.TargetCID, scoutSoldierSID, defenderRemaining); err != nil {
				return err
			}
			if err := r.ensureCityResAdd(ctx, tx, current.TargetCID); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, "update sys_city_res_add set resource_changing = 1 where cid = ?", current.TargetCID); err != nil {
				return err
			}
		}

		settleAt := time.Now().Unix()
		scoutReport := scoutReportInput{
			OriginCID:            current.CID,
			OriginCity:           fromCity,
			TargetCID:            current.TargetCID,
			TargetCity:           targetCity,
			DepartTime:           current.StartTime,
			SettleTime:           settleAt,
			PathSeconds:          current.PathTime,
			AttackerScouts:       attackerScouts,
			AttackerSurvivors:    attackerSurvivors,
			DefenderScoutsBefore: defenderScouts,
			DefenderScoutsAfter:  defenderRemaining,
			TargetSnapshot:       targetSnapshot,
		}
		attackerReport := buildRichScoutAttackerReport(scoutReport)
		if err := r.writeTroopReportTx(ctx, tx, current.UID, troopReportTitleScout, current.CID, fromCity, current.TargetCID, targetCity, attackerReport); err != nil {
			return err
		}
		if targetSnapshot.UID > 0 && targetSnapshot.UID != current.UID {
			defenderReport := buildRichScoutDefenderReport(scoutReport)
			if err := r.writeTroopReportTx(ctx, tx, targetSnapshot.UID, troopReportTitleScout, current.CID, fromCity, current.TargetCID, targetCity, defenderReport); err != nil {
				return err
			}
		}

		if current.HID > 0 && attackerScouts > 0 && attackerSurvivors < attackerScouts {
			bloodLoss := 100 - (100*attackerSurvivors)/attackerScouts
			if bloodLoss > 0 {
				if _, err := tx.ExecContext(ctx, "update mem_hero_blood set force = force - ? where hid = ?", bloodLoss, current.HID); err != nil {
					return err
				}
			}
		}

		returnSoldiersRaw := formatScoutReturnSoldiersRaw(soldierCounts, attackerSurvivors)
		if returnSoldiersRaw == "" {
			if err := r.restoreTroopHeroStateTx(ctx, tx, current.UID, current.CID, current.HID); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, "delete from sys_troops where id = ?", current.ID); err != nil {
				return err
			}
			return tx.Commit()
		}

		if _, err := tx.ExecContext(ctx, `
update sys_troops
set state = 1, starttime = ?, endtime = ?, soldiers = ?, resource = ?
where id = ? and state = 0`,
			settleAt,
			settleAt+int64(current.PathTime),
			returnSoldiersRaw,
			formatTroopResourceRaw(TroopResource{}),
			current.ID,
		); err != nil {
			return err
		}
		return tx.Commit()
	}

	if current.Task == 0 && current.State == 0 {
		fromCity := r.cityNameTx(ctx, tx, current.CID)
		targetCity := r.cityNameTx(ctx, tx, current.TargetCID)
		targetUID := 0
		var targetOwner sql.NullInt64
		if err := tx.QueryRowContext(ctx, "select uid from sys_city where cid = ?", current.TargetCID).Scan(&targetOwner); err != nil {
			if err != sql.ErrNoRows {
				return err
			}
		} else {
			targetUID = int(targetOwner.Int64)
		}

		if !resources.isZero() {
			if err := r.addCityResourcesTx(ctx, tx, current.TargetCID, resources); err != nil {
				return err
			}
		}
		if err := r.ensureCityResAdd(ctx, tx, current.TargetCID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "update sys_city_res_add set resource_changing = 1 where cid = ?", current.TargetCID); err != nil {
			return err
		}

		settleAt := time.Now().Unix()
		if _, err := tx.ExecContext(ctx, `
update sys_troops
set state = 1, starttime = ?, endtime = ?, resource = ?
where id = ? and state = 0`,
			settleAt,
			settleAt+int64(current.PathTime),
			formatTroopResourceRaw(TroopResource{}),
			current.ID,
		); err != nil {
			return err
		}

		reportInput := transportReportInput{
			OriginCID:     current.CID,
			OriginCity:    fromCity,
			TargetCID:     current.TargetCID,
			TargetCity:    targetCity,
			DepartTime:    current.StartTime,
			SettleTime:    settleAt,
			PathSeconds:   current.PathTime,
			Payload:       resources,
			ReturnStarted: true,
		}
		senderReport := buildRichTransportSenderReport(reportInput)
		if err := r.writeTroopReportTx(ctx, tx, current.UID, troopReportTitleTransport, current.CID, fromCity, current.TargetCID, targetCity, senderReport); err != nil {
			return err
		}
		if targetUID > 0 && targetUID != current.UID {
			receiverReport := buildRichTransportReceiverReport(reportInput)
			if err := r.writeTroopReportTx(ctx, tx, targetUID, troopReportTitleTransport, current.CID, fromCity, current.TargetCID, targetCity, receiverReport); err != nil {
				return err
			}
		}
		return tx.Commit()
	}
	if current.Task == 1 && current.State == 0 {
		fromCity := r.cityNameTx(ctx, tx, current.CID)
		targetCity := r.cityNameTx(ctx, tx, current.TargetCID)
		soldierNames := r.loadSoldierNames(ctx)
		reportSoldiers, _ := parseTroopSoldiers(current.Soldiers, soldierNames)
		settleAt := time.Now().Unix()
		if _, err := tx.ExecContext(ctx, `
update sys_troops
set state = 4, startcid = ?, starttime = ?, endtime = 0, resource = ?
where id = ? and state = 0`,
			current.CID,
			settleAt,
			formatTroopResourceRaw(TroopResource{}),
			current.ID,
		); err != nil {
			return err
		}
		if err := r.setTroopHeroStateTx(ctx, tx, current.UID, current.CID, current.HID, 4); err != nil {
			return err
		}

		reportInput := stationReportInput{
			HomeCID:     current.CID,
			HomeCity:    fromCity,
			TargetCID:   current.TargetCID,
			TargetCity:  targetCity,
			DepartTime:  current.StartTime,
			SettleTime:  settleAt,
			PathSeconds: current.PathTime,
			Soldiers:    reportSoldiers,
		}
		arrivalReport := buildRichStationArrivalReport(reportInput)
		if err := r.writeTroopReportTx(ctx, tx, current.UID, troopReportTitleStation, current.CID, fromCity, current.TargetCID, targetCity, arrivalReport); err != nil {
			return err
		}
		return tx.Commit()
	}
	if current.Task == 3 && current.State == 0 {
		targetSnapshot, err := r.loadPlunderTargetSnapshotTx(ctx, tx, current.TargetCID)
		if err != nil {
			return err
		}

		fromCity := r.cityNameTx(ctx, tx, current.CID)
		targetCity := firstNonEmpty(targetSnapshot.CityName, r.cityNameTx(ctx, tx, current.TargetCID))
		soldierNames := r.loadSoldierNames(ctx)
		reportSoldiers, _ := parseTroopSoldiers(current.Soldiers, soldierNames)

		loot := TroopResource{}
		note := "目标城池无守军，部队已完成掠夺并开始返程。"
		if !targetSnapshot.minimalEligible() {
			note = "目标城池在部队到达前出现了守军或驻军，当前最小掠夺链未进入战斗结算，部队空载返程。"
		} else {
			carryCapacity, err := r.troopCarryCapacityTx(ctx, tx, current.CID, soldierCounts)
			if err != nil {
				return err
			}
			loot = plunderLootForSettlement(targetSnapshot, carryCapacity)
			if loot.isZero() {
				note = "目标城池的资源均处于仓储保护范围内，本次未掠得资源，部队空载返程。"
			} else {
				if err := r.subtractCityResourcesTx(ctx, tx, current.TargetCID, loot); err != nil {
					return err
				}
				if err := r.ensureCityResAdd(ctx, tx, current.TargetCID); err != nil {
					return err
				}
				if _, err := tx.ExecContext(ctx, "update sys_city_res_add set resource_changing = 1 where cid = ?", current.TargetCID); err != nil {
					return err
				}
			}
		}

		now := time.Now().Unix()
		if _, err := tx.ExecContext(ctx, `
update sys_troops
set state = 1, starttime = ?, endtime = ?, resource = ?
where id = ? and state = 0`,
			now,
			now+int64(current.PathTime),
			formatTroopResourceRaw(loot),
			current.ID,
		); err != nil {
			return err
		}

		attackerReport := buildRichPlunderAttackerReport(fromCity, targetCity, reportSoldiers, loot, note)
		if err := r.writeTroopReportTx(ctx, tx, current.UID, 8, current.CID, fromCity, current.TargetCID, targetCity, attackerReport); err != nil {
			return err
		}
		if targetSnapshot.UID > 0 && targetSnapshot.UID != current.UID {
			defenderNote := "敌军已完成掠夺并开始返程。"
			if loot.isZero() {
				defenderNote = "敌军抵达但未掠得资源。"
			}
			if !targetSnapshot.minimalEligible() {
				defenderNote = "敌军抵达时城内已存在守军或驻军，本次未发生掠夺结算。"
			}
			defenderReport := buildRichPlunderDefenderReport(fromCity, targetCity, reportSoldiers, loot, defenderNote)
			if err := r.writeTroopReportTx(ctx, tx, targetSnapshot.UID, 8, current.CID, fromCity, current.TargetCID, targetCity, defenderReport); err != nil {
				return err
			}
		}

		return tx.Commit()
	}

	targetCity := current.CID
	if current.Task == 1 && current.State == 0 {
		targetCity = current.TargetCID
	}
	if current.Task == 1 && current.State == 1 && current.StartCID > 0 {
		stationCity := firstPositive(current.StartCID, current.TargetCID)
		fromCity := r.cityNameTx(ctx, tx, stationCity)
		targetCityName := r.cityNameTx(ctx, tx, current.CID)
		soldierNames := r.loadSoldierNames(ctx)
		reportSoldiers, _ := parseTroopSoldiers(current.Soldiers, soldierNames)
		reportInput := stationReportInput{
			HomeCID:     current.CID,
			HomeCity:    targetCityName,
			TargetCID:   stationCity,
			TargetCity:  fromCity,
			DepartTime:  current.StartTime,
			SettleTime:  time.Now().Unix(),
			PathSeconds: current.PathTime,
			Soldiers:    reportSoldiers,
		}
		returnReport := buildRichStationReturnReport(reportInput)
		if err := r.writeTroopReportTx(ctx, tx, current.UID, troopReportTitleStation, stationCity, fromCity, current.CID, targetCityName, returnReport); err != nil {
			return err
		}
	}
	if len(soldierCounts) > 0 {
		if err := r.addCitySoldiersTx(ctx, tx, targetCity, soldierCounts); err != nil {
			return err
		}
	}
	if !resources.isZero() {
		if err := r.addCityResourcesTx(ctx, tx, targetCity, resources); err != nil {
			return err
		}
	}
	if err := r.ensureCityResAdd(ctx, tx, targetCity); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "update sys_city_res_add set resource_changing = 1 where cid = ?", targetCity); err != nil {
		return err
	}
	if err := r.restoreTroopHeroStateTx(ctx, tx, current.UID, targetCity, current.HID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "delete from sys_troops where id = ?", current.ID); err != nil {
		return err
	}

	return tx.Commit()
}

func canCallbackTroopState(state int) bool {
	switch state {
	case 0, 2, 4, 5:
		return true
	default:
		return false
	}
}

func isSettledTroopTask(task int) bool {
	switch task {
	case 0, 1, 2, 3, 4:
		return true
	default:
		return false
	}
}

func taskRequiresHero(task int) bool {
	switch task {
	case 3, 4:
		return true
	default:
		return false
	}
}

func isDueTroopState(state int) bool {
	switch state {
	case 0, 1, 5:
		return true
	default:
		return false
	}
}

func callbackTroopMarksCityResources(state int) bool {
	switch state {
	case 2, 4, 5:
		return true
	default:
		return false
	}
}

func callbackTroopReturnRoute(record troopRecord, now int64) (int, int, int64) {
	endTime := now + int64(record.PathTime)
	if record.State == 0 {
		elapsed := now - record.StartTime
		if elapsed < 0 {
			elapsed = 0
		}
		endTime = now + elapsed
	}

	startCID := record.StartCID
	targetCID := record.TargetCID
	if record.State == 4 || record.State == 5 {
		startCID = record.TargetCID
		targetCID = record.CID
	}
	return startCID, targetCID, endTime
}

func (r *Repository) cityBuildingLevel(ctx context.Context, tx *sql.Tx, cid int, bid int) (int, error) {
	var level sql.NullInt64
	if err := tx.QueryRowContext(ctx, "select level from sys_building where cid = ? and bid = ? limit 1", cid, bid).Scan(&level); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return int(level.Int64), nil
}

func (r *Repository) cityActiveTroopCount(ctx context.Context, tx *sql.Tx, uid int, cid int) (int, error) {
	var count int
	err := tx.QueryRowContext(ctx, cityActiveTroopCountSQL(), cid, uid).Scan(&count)
	return count, err
}

func cityActiveTroopCountSQL() string {
	return "select count(*) from sys_troops where cid = ? and uid = ? and state < 4"
}

func (r *Repository) citySoldierCount(ctx context.Context, tx *sql.Tx, cid int, sid int) (int64, error) {
	var count sql.NullInt64
	if err := tx.QueryRowContext(ctx, "select count from sys_city_soldier where cid = ? and sid = ?", cid, sid).Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return count.Int64, nil
}

func (r *Repository) cityExists(ctx context.Context, cid int) (bool, error) {
	if r.db == nil {
		return true, nil
	}

	// Check sys_city for actual cities
	var cityCount int
	if err := r.db.QueryRowContext(ctx, "select count(*) from sys_city where cid = ?", cid).Scan(&cityCount); err != nil {
		return false, err
	}
	if cityCount > 0 {
		return true, nil
	}

	// Check mem_world for fields (野地). Fields have type>0 and are valid dispatch targets.
	// For fields, cid = y*1000 + x maps back to wid
	wid := cidToWid(cid)
	var worldType int
	if err := r.db.QueryRowContext(ctx, "select type from mem_world where wid = ?", wid).Scan(&worldType); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	// type > 0 means it's a field, valid for dispatch (scout task=2)
	return worldType > 0, nil
}

func (r *Repository) userTroopRecord(ctx context.Context, tx *sql.Tx, uid int, troopID int) (troopRecord, error) {
	record := troopRecord{}
	err := tx.QueryRowContext(ctx, `
select id, uid, cid, startcid, targetcid, hid, task, state, starttime, pathtime, endtime, soldiers, resource
from sys_troops
where id = ? and uid = ?`,
		troopID,
		uid,
	).Scan(
		&record.ID,
		&record.UID,
		&record.CID,
		&record.StartCID,
		&record.TargetCID,
		&record.HID,
		&record.Task,
		&record.State,
		&record.StartTime,
		&record.PathTime,
		&record.EndTime,
		&record.Soldiers,
		&record.Resource,
	)
	return record, err
}

func (r *Repository) addCitySoldiersTx(ctx context.Context, tx *sql.Tx, cid int, soldiers map[int]int64) error {
	for sid, count := range soldiers {
		if count <= 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
insert into sys_city_soldier (cid, sid, count)
values (?, ?, ?)
on duplicate key update count = count + values(count)`,
			cid,
			sid,
			count,
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) setCitySoldierCountTx(ctx context.Context, tx *sql.Tx, cid int, sid int, count int64) error {
	if sid <= 0 {
		return nil
	}
	if count <= 0 {
		_, err := tx.ExecContext(ctx, "update sys_city_soldier set count = 0 where cid = ? and sid = ?", cid, sid)
		return err
	}
	_, err := tx.ExecContext(ctx, `
insert into sys_city_soldier (cid, sid, count)
values (?, ?, ?)
on duplicate key update count = values(count)`,
		cid,
		sid,
		count,
	)
	return err
}

func (r *Repository) ensureDispatchHeroTx(ctx context.Context, tx *sql.Tx, uid int, cid int, hid int) error {
	if hid <= 0 {
		return nil
	}

	var state int
	if err := tx.QueryRowContext(ctx, `
select state
from sys_city_hero
where cid = ? and uid = ? and hid = ?`, cid, uid, hid).Scan(&state); err != nil {
		if err == sql.ErrNoRows {
			return newInvalidError("出征将领无效。")
		}
		return err
	}
	if state == 2 || state == 3 || state == 5 || state == 6 {
		return newInvalidError("将领出征中，不能再次派遣。")
	}

	busy, err := r.heroIsInTroop(ctx, tx, hid)
	if err != nil {
		return err
	}
	if busy {
		return newInvalidError("将领出征中，不能再次派遣。")
	}
	return nil
}

func (r *Repository) setTroopHeroStateTx(ctx context.Context, tx *sql.Tx, uid int, cid int, hid int, state int) error {
	if hid <= 0 {
		return nil
	}
	_, err := tx.ExecContext(ctx, "update sys_city_hero set state = ? where hid = ? and uid = ? and cid = ?", state, hid, uid, cid)
	return err
}

func (r *Repository) restoreTroopHeroStateTx(ctx context.Context, tx *sql.Tx, uid int, cid int, hid int) error {
	if hid <= 0 {
		return nil
	}

	state := 0
	heroCID := cid
	if err := tx.QueryRowContext(ctx, "select cid from sys_city_hero where hid = ? and uid = ?", hid, uid).Scan(&heroCID); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	var chiefHID, generalHID, counsellorHID sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
select chiefhid, generalid, counsellorid
from sys_city
where cid = ? and uid = ?`, heroCID, uid).Scan(&chiefHID, &generalHID, &counsellorHID); err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	} else {
		switch int64(hid) {
		case chiefHID.Int64:
			state = 1
		case generalHID.Int64:
			state = 7
		case counsellorHID.Int64:
			state = 8
		}
	}
	return r.setTroopHeroStateTx(ctx, tx, uid, heroCID, hid, state)
}

func (p TroopResource) isZero() bool {
	return p.Gold == 0 && p.Food == 0 && p.Wood == 0 && p.Rock == 0 && p.Iron == 0
}

func normalizeTroopSoldiers(soldiers map[int]int64) (map[int]int64, int64, error) {
	normalized := make(map[int]int64, len(soldiers))
	var total int64
	for sid, count := range soldiers {
		if sid <= 0 || count <= 0 {
			continue
		}
		normalized[sid] += count
		total += count
	}
	if total <= 0 {
		return nil, 0, newInvalidError("派遣兵力无效。")
	}
	return normalized, total, nil
}

func normalizeTroopResource(resource TroopResource) (TroopResource, error) {
	if resource.Gold < 0 || resource.Food < 0 || resource.Wood < 0 || resource.Rock < 0 || resource.Iron < 0 {
		return TroopResource{}, newInvalidError("运输资源不能为负数。")
	}
	return resource, nil
}

func (r *Repository) troopSoldierStatsTx(ctx context.Context, tx *sql.Tx, soldiers map[int]int64) (troopSoldierStats, error) {
	stats := troopSoldierStats{}
	for sid, count := range soldiers {
		if sid <= 0 || count <= 0 {
			continue
		}

		var peopleNeed sql.NullInt64
		var foodUse sql.NullFloat64
		if err := tx.QueryRowContext(ctx, "select coalesce(people_need, 0), coalesce(food_use, 0) from cfg_soldier where sid = ?", sid).Scan(&peopleNeed, &foodUse); err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return troopSoldierStats{}, err
		}
		stats.People += peopleNeed.Int64 * count
		stats.FoodUse += foodUse.Float64 * float64(count)
	}
	return stats, nil
}

func formatTroopResourceRaw(payload TroopResource) string {
	return fmt.Sprintf("%d,%d,%d,%d,%d", payload.Gold, payload.Food, payload.Wood, payload.Rock, payload.Iron)
}

func (r *Repository) addCityResourcesTx(ctx context.Context, tx *sql.Tx, cid int, payload TroopResource) error {
	if payload.isZero() {
		return nil
	}
	if err := r.ensureCityResourceRowTx(ctx, tx, cid); err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `
update mem_city_resource
set wood = coalesce(wood, 0) + ?, rock = coalesce(rock, 0) + ?, iron = coalesce(iron, 0) + ?, food = coalesce(food, 0) + ?, gold = coalesce(gold, 0) + ?
where cid = ?`,
		payload.Wood,
		payload.Rock,
		payload.Iron,
		payload.Food,
		payload.Gold,
		cid,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected <= 0 {
		return newInvalidError("target city resource row is missing; resources were not added")
	}
	return nil
}

func (r *Repository) ensureCityResourceRowTx(ctx context.Context, tx *sql.Tx, cid int) error {
	_, err := tx.ExecContext(ctx, ensureCityResourceRowSQL(), cid)
	return err
}

func ensureCityResourceRowSQL() string {
	return `
insert ignore into mem_city_resource (
	` + "`cid`,`people`,`food`,`wood`,`rock`,`iron`,`gold`,`food_max`,`wood_max`,`rock_max`,`iron_max`,`gold_max`,`food_add`,`wood_add`,`rock_add`,`iron_add`,`lastupdate`" + `
)
select cid,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,unix_timestamp()
from sys_city
where cid = ?`
}

func (r *Repository) subtractCityResourcesTx(ctx context.Context, tx *sql.Tx, cid int, payload TroopResource) error {
	result, err := tx.ExecContext(ctx, `
update mem_city_resource
set wood = wood - ?, rock = rock - ?, iron = iron - ?, food = food - ?, gold = gold - ?
where cid = ? and wood >= ? and rock >= ? and iron >= ? and food >= ? and gold >= ?`,
		payload.Wood,
		payload.Rock,
		payload.Iron,
		payload.Food,
		payload.Gold,
		cid,
		payload.Wood,
		payload.Rock,
		payload.Iron,
		payload.Food,
		payload.Gold,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected <= 0 {
		return newInvalidError("资源不足，无法运输。")
	}
	return nil
}

func formatTroopSoldiersRaw(soldiers map[int]int64) string {
	parts := make([]string, 0, 1+len(soldiers)*2)
	ids := make([]int, 0, len(soldiers))
	for sid := range soldiers {
		if sid > 0 && soldiers[sid] > 0 {
			ids = append(ids, sid)
		}
	}
	sort.Ints(ids)
	parts = append(parts, strconv.Itoa(len(ids)))
	for _, sid := range ids {
		count := soldiers[sid]
		parts = append(parts, strconv.Itoa(sid), strconv.FormatInt(count, 10))
	}
	if len(ids) == 0 {
		return ""
	}
	return strings.Join(parts, ",") + ","
}

func formatScoutReturnSoldiersRaw(soldiers map[int]int64, survivingScouts int64) string {
	returning := make(map[int]int64, len(soldiers))
	for sid, count := range soldiers {
		returning[sid] = count
	}
	returning[scoutSoldierSID] = survivingScouts
	return formatTroopSoldiersRaw(returning)
}

func parseTroopSoldierCounts(raw string) map[int]int64 {
	pairs := parseTroopSoldierPairs(raw)
	soldiers := make(map[int]int64, len(pairs))
	for _, pair := range pairs {
		soldiers[pair.sid] += pair.count
	}
	return soldiers
}

type troopSoldierPair struct {
	sid   int
	count int64
}

func parseTroopSoldierPairs(raw string) []troopSoldierPair {
	tokens := troopSoldierTokens(raw)
	if len(tokens) == 0 {
		return []troopSoldierPair{}
	}

	start := 0
	if typeCount, err := strconv.Atoi(tokens[0]); err == nil && typeCount >= 0 && len(tokens) == 1+typeCount*2 {
		start = 1
	}

	pairs := make([]troopSoldierPair, 0, (len(tokens)-start)/2)
	for index := start; index+1 < len(tokens); index += 2 {
		sid, err := strconv.Atoi(tokens[index])
		if err != nil || sid <= 0 {
			continue
		}
		count, err := strconv.ParseInt(tokens[index+1], 10, 64)
		if err != nil || count <= 0 {
			continue
		}
		pairs = append(pairs, troopSoldierPair{sid: sid, count: count})
	}
	return pairs
}

func troopSoldierTokens(raw string) []string {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	tokens := make([]string, 0, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token != "" {
			tokens = append(tokens, token)
		}
	}
	return tokens
}

func parseTroopResourcePayload(raw string) TroopResource {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	values := make([]int64, 0, 5)
	for _, part := range parts {
		if len(values) >= 5 {
			break
		}
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		value, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			values = append(values, 0)
			continue
		}
		values = append(values, value)
	}
	for len(values) < 5 {
		values = append(values, 0)
	}
	return TroopResource{
		Gold: values[0],
		Food: values[1],
		Wood: values[2],
		Rock: values[3],
		Iron: values[4],
	}
}

func (r *Repository) fixtureDispatchTroop(uid int, cid int, targetCID int, soldiers map[int]int64, heroID int, task int, resource TroopResource) (TroopPage, error) {
	page := r.fixtureTroopPage(uid)
	soldierNames := r.loadSoldierNames(context.Background())
	reportSoldiers, soldierTotal := parseTroopSoldiers(formatTroopSoldiersRaw(soldiers), soldierNames)
	page.Items = append([]TroopCard{
		{
			ID:           9901,
			UID:          uid,
			CID:          cid,
			StartCID:     cid,
			TargetCID:    targetCID,
			FromCity:     formatCIDLabel(cid),
			TargetCity:   formatCIDLabel(targetCID),
			HeroID:       heroID,
			HeroName:     "--",
			Task:         task,
			TaskLabel:    troopTaskLabel(task),
			State:        0,
			StateLabel:   troopStateLabel(0),
			StartTime:    time.Now().Unix(),
			EndTime:      time.Now().Add(minimalTroopPathSeconds * time.Second).Unix(),
			PathTime:     minimalTroopPathSeconds,
			SecondsLeft:  minimalTroopPathSeconds,
			People:       soldierTotal,
			FoodUse:      0,
			SoldiersRaw:  formatTroopSoldiersRaw(soldiers),
			ResourceRaw:  formatTroopResourceRaw(resource),
			Resources:    resource,
			Resource:     resource,
			Soldiers:     reportSoldiers,
			SoldierCount: soldierTotal,
		},
	}, page.Items...)
	page.Total = len(page.Items)
	page.Moving++
	return page, nil
}

func (r *Repository) fixtureCallbackTroop(uid int, troopID int) TroopPage {
	page := r.fixtureTroopPage(uid)
	items := make([]TroopCard, 0, len(page.Items))
	for _, item := range page.Items {
		if item.ID == troopID {
			continue
		}
		items = append(items, item)
	}
	page.Items = items
	page.Total = len(items)
	page.Moving = 0
	if page.Stationed > 0 {
		page.Stationed = 1
	}
	return page
}
