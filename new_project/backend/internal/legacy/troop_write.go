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

func (r *Repository) DispatchCityTroop(ctx context.Context, uid int, cid int, targetCID int, soldierSID int, soldierCount int, task int, resource TroopResource) (TroopPage, error) {
	if targetCID <= 0 || targetCID == cid {
		return TroopPage{}, newInvalidError("目标城池无效。")
	}
	if soldierSID <= 0 || soldierCount <= 0 {
		return TroopPage{}, newInvalidError("派遣兵力无效。")
	}
	if task < 0 || task > 3 {
		return TroopPage{}, newInvalidError("only task=0, task=1, task=2, task=3 are supported")
	}
	resourcePayload, err := normalizeTroopResource(resource)
	if err != nil {
		return TroopPage{}, err
	}
	if task == 1 && !resourcePayload.isZero() {
		return TroopPage{}, newInvalidError("task=1 不支持携带资源。")
	}

	if task == 2 {
		if soldierSID != scoutSoldierSID {
			return TroopPage{}, newInvalidError("task=2 only supports scout sid=3")
		}
		if !resourcePayload.isZero() {
			return TroopPage{}, newInvalidError("task=2 does not support resources")
		}
	}
	if task == 3 {
		if soldierSID == scoutSoldierSID {
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
		return r.fixtureDispatchTroop(uid, cid, targetCID, soldierSID, soldierCount, task, resourcePayload)
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

	availableSoldiers, err := r.citySoldierCount(ctx, tx, cid, soldierSID)
	if err != nil {
		return TroopPage{}, err
	}
	if availableSoldiers < int64(soldierCount) {
		return TroopPage{}, newInvalidError("兵力不足，无法派遣。")
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

	if _, err := tx.ExecContext(ctx, `
update sys_city_soldier
set count = count - ?
where cid = ? and sid = ? and count >= ?`,
		soldierCount,
		cid,
		soldierSID,
		soldierCount,
	); err != nil {
		return TroopPage{}, err
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
	soldiersRaw := formatTroopSoldiersRaw(map[int]int64{soldierSID: int64(soldierCount)})
	resourceRaw := formatTroopResourceRaw(resourcePayload)
	if _, err := tx.ExecContext(ctx, `
insert into sys_troops (uid, cid, startcid, hid, targetcid, task, state, starttime, pathtime, endtime, soldiers, resource, people, fooduse)
values (?, ?, ?, 0, ?, ?, 0, ?, ?, ?, ?, ?, ?, 0)`,
		uid,
		cid,
		cid,
		targetCID,
		task,
		now,
		minimalTroopPathSeconds,
		now+minimalTroopPathSeconds,
		soldiersRaw,
		resourceRaw,
		soldierCount,
	); err != nil {
		return TroopPage{}, err
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
	if record.State != 0 && record.State != 2 && record.State != 4 {
		return TroopPage{}, newInvalidError("当前部队状态不支持撤回。")
	}

	now := time.Now().Unix()
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
	if record.Task == 1 && record.State == 4 {
		startCID = record.TargetCID
		targetCID = record.CID
	}

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

	if record.State == 2 || record.State == 4 {
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

	rows, err := r.db.QueryContext(ctx, `
select id, uid, cid, startcid, targetcid, hid, task, state, starttime, pathtime, endtime, soldiers, resource
from sys_troops
where uid = ? and task in (0, 1, 2, 3) and endtime > 0 and endtime <= ? and state in (0, 1)
order by id asc`,
		uid,
		time.Now().Unix(),
	)
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
	if (current.Task != 0 && current.Task != 1 && current.Task != 2 && current.Task != 3) || (current.State != 0 && current.State != 1) || current.EndTime > time.Now().Unix() {
		return tx.Commit()
	}

	soldierCounts := parseTroopSoldierCounts(current.Soldiers)
	resources := parseTroopResourcePayload(current.Resource)

	if current.Task == 2 && current.State == 0 {
		attackerScouts := soldierCounts[scoutSoldierSID]
		if attackerScouts <= 0 {
			if _, err := tx.ExecContext(ctx, "delete from sys_troops where id = ?", current.ID); err != nil {
				return err
			}
			return tx.Commit()
		}

		targetSnapshot, err := r.loadPlunderTargetSnapshotTx(ctx, tx, current.TargetCID)
		if err != nil {
			return err
		}
		fromCity := r.cityNameTx(ctx, tx, current.CID)
		targetCity := firstNonEmpty(targetSnapshot.CityName, r.cityNameTx(ctx, tx, current.TargetCID))

		defenderScouts, err := r.citySoldierCount(ctx, tx, current.TargetCID, scoutSoldierSID)
		if err != nil {
			return err
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
		if defenderRemaining != defenderScouts {
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

		if attackerSurvivors <= 0 {
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
			formatTroopSoldiersRaw(map[int]int64{scoutSoldierSID: attackerSurvivors}),
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
			loot = computePlunderLoot(targetSnapshot, carryCapacity)
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
	if _, err := tx.ExecContext(ctx, "delete from sys_troops where id = ?", current.ID); err != nil {
		return err
	}

	return tx.Commit()
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
	err := tx.QueryRowContext(ctx, "select count(*) from sys_troops where cid = ? and uid = ? and state < 4", cid, uid).Scan(&count)
	return count, err
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

func (p TroopResource) isZero() bool {
	return p.Gold == 0 && p.Food == 0 && p.Wood == 0 && p.Rock == 0 && p.Iron == 0
}

func normalizeTroopResource(resource TroopResource) (TroopResource, error) {
	if resource.Gold < 0 || resource.Food < 0 || resource.Wood < 0 || resource.Rock < 0 || resource.Iron < 0 {
		return TroopResource{}, newInvalidError("运输资源不能为负数。")
	}
	return resource, nil
}

func formatTroopResourceRaw(payload TroopResource) string {
	return fmt.Sprintf("%d,%d,%d,%d,%d", payload.Gold, payload.Food, payload.Wood, payload.Rock, payload.Iron)
}

func (r *Repository) addCityResourcesTx(ctx context.Context, tx *sql.Tx, cid int, payload TroopResource) error {
	_, err := tx.ExecContext(ctx, `
update mem_city_resource
set wood = wood + ?, rock = rock + ?, iron = iron + ?, food = food + ?, gold = gold + ?
where cid = ?`,
		payload.Wood,
		payload.Rock,
		payload.Iron,
		payload.Food,
		payload.Gold,
		cid,
	)
	return err
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
	parts := make([]string, 0, len(soldiers)*2)
	for sid, count := range soldiers {
		if sid <= 0 || count <= 0 {
			continue
		}
		parts = append(parts, strconv.Itoa(sid), strconv.FormatInt(count, 10))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ",") + ","
}

func parseTroopSoldierCounts(raw string) map[int]int64 {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	soldiers := make(map[int]int64, len(parts)/2)
	for index := 0; index+1 < len(parts); index += 2 {
		sidText := strings.TrimSpace(parts[index])
		countText := strings.TrimSpace(parts[index+1])
		if sidText == "" || countText == "" {
			continue
		}

		sid, err := strconv.Atoi(sidText)
		if err != nil || sid <= 0 {
			continue
		}
		count, err := strconv.ParseInt(countText, 10, 64)
		if err != nil || count <= 0 {
			continue
		}
		soldiers[sid] += count
	}
	return soldiers
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

func (r *Repository) fixtureDispatchTroop(uid int, cid int, targetCID int, soldierSID int, soldierCount int, task int, resource TroopResource) (TroopPage, error) {
	page := r.fixtureTroopPage(uid)
	page.Items = append([]TroopCard{
		{
			ID:           9901,
			UID:          uid,
			CID:          cid,
			StartCID:     cid,
			TargetCID:    targetCID,
			FromCity:     formatCIDLabel(cid),
			TargetCity:   formatCIDLabel(targetCID),
			HeroID:       0,
			HeroName:     "--",
			Task:         task,
			TaskLabel:    troopTaskLabel(task),
			State:        0,
			StateLabel:   troopStateLabel(0),
			StartTime:    time.Now().Unix(),
			EndTime:      time.Now().Add(minimalTroopPathSeconds * time.Second).Unix(),
			PathTime:     minimalTroopPathSeconds,
			SecondsLeft:  minimalTroopPathSeconds,
			People:       int64(soldierCount),
			FoodUse:      0,
			SoldiersRaw:  fmt.Sprintf("%d,%d,", soldierSID, soldierCount),
			ResourceRaw:  formatTroopResourceRaw(resource),
			Resources:    resource,
			Resource:     resource,
			Soldiers:     []Soldier{{SID: soldierSID, Name: fmt.Sprintf("SID %d", soldierSID), Count: int64(soldierCount)}},
			SoldierCount: int64(soldierCount),
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
