package legacy

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"unicode/utf8"
)

const (
	legacyMaxUserName = 8
	legacyMaxCityName = 8
)

func (r *Repository) LegacyCreateRole(ctx context.Context, payload LegacyRoleCreatePayload) (LegacyRoleCreateResult, error) {
	result := LegacyRoleCreateResult{
		Raw: []any{},
		UID: payload.UID,
	}

	if payload.UID <= 0 || payload.SID <= 0 {
		result.Raw = []any{0, "invalid_user_auth"}
		return result, nil
	}
	if r.db == nil {
		result.Raw = []any{0, "legacy_database_unavailable"}
		return result, nil
	}

	authOK, err := r.legacySessionMatches(ctx, payload.UID, payload.SID)
	if err != nil {
		return result, err
	}
	if !authOK {
		result.Raw = []any{0, "invalid_user_auth"}
		return result, nil
	}

	if err := validateLegacyRolePayload(payload); err != nil {
		result.Raw = []any{0, err.Error()}
		return result, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	userState, err := r.userStateTx(ctx, tx, payload.UID)
	if err != nil {
		return result, err
	}
	if userState != 3 {
		result.Raw = []any{0, "cant_duplicate_create"}
		return result, nil
	}

	exists, err := r.userHasAnyCityTx(ctx, tx, payload.UID)
	if err != nil {
		return result, err
	}
	if exists {
		result.Raw = []any{0, "cant_duplicate_create"}
		return result, nil
	}

	username := strings.TrimSpace(payload.UserName)
	cityName := strings.TrimSpace(payload.CityName)
	flagChar := strings.TrimSpace(payload.FlagChar)
	if err := r.ensureNameNotBannedTx(ctx, tx, username, "invalid_char"); err != nil {
		if errors.Is(err, ErrInvalid) {
			result.Raw = []any{0, err.Error()}
			return result, nil
		}
		return result, err
	}
	if err := r.ensureNameNotBannedTx(ctx, tx, cityName, "name_illegal"); err != nil {
		if errors.Is(err, ErrInvalid) {
			result.Raw = []any{0, err.Error()}
			return result, nil
		}
		return result, err
	}
	if err := r.ensureUserNameAvailableTx(ctx, tx, payload.UID, username); err != nil {
		if errors.Is(err, ErrInvalid) {
			result.Raw = []any{0, err.Error()}
			return result, nil
		}
		return result, err
	}

	cityID, province, err := r.allocateStarterCityTx(ctx, tx, payload.Province)
	if err != nil {
		if errors.Is(err, ErrInvalid) {
			result.Raw = []any{0, err.Error()}
			return result, nil
		}
		return result, err
	}

	if err := r.createStarterCityTx(ctx, tx, payload.UID, cityID, province, cityName); err != nil {
		return result, err
	}
	if err := r.updateStarterUserProfileTx(ctx, tx, payload.UID, cityID, username, flagChar, payload.Sex, payload.Face); err != nil {
		return result, err
	}
	if err := r.insertStarterMailsTx(ctx, tx, payload.UID); err != nil {
		return result, err
	}
	if err := r.insertStarterTasksTx(ctx, tx, payload.UID); err != nil {
		return result, err
	}
	if err := r.insertStarterGoodsTx(ctx, tx, payload.UID); err != nil {
		return result, err
	}

	if err := tx.Commit(); err != nil {
		return result, err
	}
	committed = true

	result.CID = cityID
	user, userErr := r.SessionUser(ctx, payload.UID)
	if userErr == nil {
		result.User = &user
	}
	return result, nil
}

func validateLegacyRolePayload(payload LegacyRoleCreatePayload) error {
	username := strings.TrimSpace(payload.UserName)
	cityName := strings.TrimSpace(payload.CityName)
	flagChar := strings.TrimSpace(payload.FlagChar)

	if utf8.RuneCountInString(username) < 1 {
		return newInvalidError("city_holder_name_not_null")
	}
	if utf8.RuneCountInString(username) > legacyMaxUserName {
		return newInvalidError("city_holder_name_too_long")
	}
	if utf8.RuneCountInString(cityName) > legacyMaxCityName {
		return newInvalidError("city_name_too_long")
	}
	if containsLegacyIllegalUserChars(username) {
		return newInvalidError("no_illege_char")
	}
	if containsLegacyIllegalCityChars(cityName) {
		return newInvalidError("name_illegal")
	}
	flagLen := utf8.RuneCountInString(flagChar)
	if flagLen == 0 {
		return newInvalidError("enter_flag_char")
	}
	if flagLen > 1 {
		return newInvalidError("single_char")
	}

	return nil
}

func containsLegacyIllegalUserChars(value string) bool {
	return strings.Contains(value, "<") || strings.Contains(value, "'") || strings.Contains(value, "\\")
}

func containsLegacyIllegalCityChars(value string) bool {
	return strings.Contains(value, "'") || strings.Contains(value, "\\")
}

func (r *Repository) legacySessionMatches(ctx context.Context, uid int, sid int64) (bool, error) {
	if r.db == nil {
		return false, nil
	}
	var count int
	if err := r.db.QueryRowContext(ctx, "select count(*) from sys_sessions where uid = ? and sid = ?", uid, sid).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) userStateTx(ctx context.Context, tx *sql.Tx, uid int) (int, error) {
	var state int
	if err := tx.QueryRowContext(ctx, "select state from sys_user where uid = ?", uid).Scan(&state); err != nil {
		if err == sql.ErrNoRows {
			return 0, newInvalidError("invalid_user_auth")
		}
		return 0, err
	}
	return state, nil
}

func (r *Repository) userHasAnyCityTx(ctx context.Context, tx *sql.Tx, uid int) (bool, error) {
	var count int
	if err := tx.QueryRowContext(ctx, "select count(*) from sys_city where uid = ?", uid).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) ensureNameNotBannedTx(ctx context.Context, tx *sql.Tx, value string, reason string) error {
	var count int
	if err := tx.QueryRowContext(ctx, "select count(*) from cfg_baned_name where instr(?, `name`) > 0", value).Scan(&count); err != nil {
		// Older snapshots may miss this table; keep behavior permissive instead of failing hard.
		if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			return nil
		}
		return err
	}
	if count > 0 {
		return newInvalidError(reason)
	}
	return nil
}

func (r *Repository) ensureUserNameAvailableTx(ctx context.Context, tx *sql.Tx, uid int, username string) error {
	var count int
	if err := tx.QueryRowContext(ctx, "select count(*) from sys_user where name = ? and uid <> ?", username, uid).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return newInvalidError("used_city_holder_name")
	}
	return nil
}

func (r *Repository) allocateStarterCityTx(ctx context.Context, tx *sql.Tx, requestedProvince int) (int, int, error) {
	province := requestedProvince
	if province < 0 {
		province = 0
	}

	for attempt := 0; attempt < 20; attempt++ {
		cid, resolvedProvince, err := r.pickStarterCIDTx(ctx, tx, province)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				if province > 0 {
					return 0, 0, newInvalidError("province_is_full")
				}
				return 0, 0, newInvalidError("no_starter_land")
			}
			return 0, 0, err
		}

		var exists int
		if err := tx.QueryRowContext(ctx, "select count(*) from sys_city where cid = ?", cid).Scan(&exists); err != nil {
			return 0, 0, err
		}
		if exists == 0 {
			return cid, resolvedProvince, nil
		}
	}

	return 0, 0, newInvalidError("failed_to_allocate_city")
}

func (r *Repository) pickStarterCIDTx(ctx context.Context, tx *sql.Tx, province int) (int, int, error) {
	var (
		wid              int
		resolvedProvince int
		err              error
	)
	if province > 0 {
		err = tx.QueryRowContext(ctx, `
select wid, province
from mem_world
where type = 1 and ownercid = 0 and state = 0 and province = ?
order by rand()
limit 1`, province).Scan(&wid, &resolvedProvince)
	} else {
		err = tx.QueryRowContext(ctx, `
select wid, province
from mem_world
where type = 1 and ownercid = 0 and state = 0
order by rand()
limit 1`).Scan(&wid, &resolvedProvince)
	}
	if err != nil {
		return 0, 0, err
	}
	return widToCID(wid), resolvedProvince, nil
}

func (r *Repository) createStarterCityTx(ctx context.Context, tx *sql.Tx, uid int, cid int, province int, cityName string) error {
	wid := cidToWid(cid)
	updateResult, err := tx.ExecContext(ctx, "update mem_world set ownercid = ?, `type` = 0 where wid = ? and ownercid = 0 and `type` = 1", cid, wid)
	if err != nil {
		return err
	}
	affected, err := updateResult.RowsAffected()
	if err == nil && affected == 0 {
		return newInvalidError("starter_land_conflict")
	}

	if _, err := tx.ExecContext(ctx, "replace into sys_city (`cid`,`uid`,`name`,`type`,`state`,`province`) values (?,?,?,0,2,?)", cid, uid, cityName, province); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "delete from sys_building where cid = ?", cid); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "replace into sys_building (`cid`,`xy`,`bid`,`level`) values (?,120,6,1)", cid); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "replace into sys_city_res_add (cid,food_rate,wood_rate,rock_rate,iron_rate,chief_add) values (?,80,80,80,80,0)", cid); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
replace into mem_city_resource (
	`+"`cid`,`people`,`food`,`wood`,`rock`,`iron`,`gold`,`food_max`,`wood_max`,`rock_max`,`iron_max`,`gold_max`,`food_add`,`wood_add`,`rock_add`,`iron_add`,`lastupdate`"+`
) values (
	?,50000,50000000,500000,500000,500000,500000,10000000,10000000,10000000,10000000,100000000,10000,10000,10000,10000,unix_timestamp()
)`, cid); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
replace into mem_city_schedule (`+"`cid`,`create_time`,`next_good_event`,`next_bad_event`"+`)
values (
	?,
	unix_timestamp(),
	unix_timestamp()-(unix_timestamp()+8*3600)%86400 + 86400 + 86400 * rand(),
	unix_timestamp()-(unix_timestamp()+8*3600)%86400 + 86400 + 86400 * rand()
)`, cid); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
insert into mem_user_schedule (uid,start_new_protect)
values (?,unix_timestamp())
on duplicate key update start_new_protect = unix_timestamp()`, uid); err != nil {
		return err
	}

	return nil
}

func (r *Repository) updateStarterUserProfileTx(ctx context.Context, tx *sql.Tx, uid int, cid int, username string, flagChar string, sex int, face int) error {
	_, err := tx.ExecContext(ctx, `
update sys_user
set state = 0, lastcid = ?, `+"`name`"+` = ?, face = ?, sex = ?, flagchar = ?
where uid = ?`, cid, username, face, sex, flagChar, uid)
	return err
}

func (r *Repository) insertStarterMailsTx(ctx context.Context, tx *sql.Tx, uid int) error {
	rows := []struct {
		ContentID int
		Title     string
	}{
		{ContentID: 12, Title: "客服帮助"},
		{ContentID: 13, Title: "欢迎进驻《热血三国》"},
		{ContentID: 14, Title: "热血三国游戏说明"},
		{ContentID: 5, Title: "新手保护提醒"},
	}
	for _, row := range rows {
		if _, err := tx.ExecContext(ctx, `
insert into sys_mail_sys_box (`+"`uid`,`contentid`,`title`,`read`,`posttime`"+`)
values (?, ?, ?, 0, unix_timestamp())`, uid, row.ContentID, row.Title); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, "insert into sys_alarm (`uid`,`mail`) values (?,1) on duplicate key update `mail`=1", uid); err != nil {
		return err
	}
	return nil
}

func (r *Repository) insertStarterTasksTx(ctx context.Context, tx *sql.Tx, uid int) error {
	_, err := tx.ExecContext(ctx, "insert into sys_user_task (uid,tid,state) values (?,1,0) on duplicate key update state=state", uid)
	return err
}

func (r *Repository) insertStarterGoodsTx(ctx context.Context, tx *sql.Tx, uid int) error {
	starterGoods := []struct {
		GID   int
		Count int64
	}{
		{GID: 50101, Count: 1},
		{GID: 67, Count: 2},
	}
	for _, item := range starterGoods {
		if err := r.addUserGoodsTx(ctx, tx, uid, item.GID, item.Count); err != nil {
			return err
		}
	}
	return nil
}
