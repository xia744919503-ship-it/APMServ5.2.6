package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type legacyUserRecord struct {
	UID      int
	Passport string
	PassType string
	State    int
}

func (r *Repository) LegacyDoLogin(ctx context.Context, payload LegacyLoginPayload, ip int64) (LegacyLoginResult, error) {
	result := LegacyLoginResult{}

	// Keep the same early version guard behavior: for some callers a mismatch only returns [2].
	serverVersion := r.stateValue(ctx, 3, int64(payload.Version))
	if payload.Version > 0 && serverVersion > 0 && int64(payload.Version) != serverVersion {
		if payload.LoginType == 2 {
			result.Raw = []any{2}
			return result, nil
		}
		result.Raw = []any{0, "client_version_old"}
		return result, nil
	}

	serverState := r.stateValue(ctx, 2, 1)
	if serverState != 1 {
		if payload.LoginType == 1 {
			result.Raw = []any{2}
			return result, nil
		}
		result.Raw = []any{0, r.serverStateMessage(ctx, int(serverState))}
		return result, nil
	}

	if payload.LoginType != 0 {
		result.Raw = []any{0, "unsupported_login_type"}
		return result, nil
	}

	passType := strings.TrimSpace(payload.PassType)
	if passType == "" {
		passType = "local"
	}
	passport := strings.TrimSpace(payload.Passport)
	if passport == "" {
		result.Raw = []any{0, "invalid_user_pwd"}
		return result, nil
	}

	record, err := r.findLegacyUserByPassport(ctx, passport, passType)
	if err != nil && err != sql.ErrNoRows {
		return result, err
	}

	sid := time.Now().UnixNano() & 0x7fffffff
	now := time.Now().Unix()
	if record.UID == 0 {
		created, createErr := r.createLegacyUser(ctx, passport, passType, sid, ip, now)
		if createErr != nil {
			return result, createErr
		}
		record = created
	} else {
		// Mirror the old flow: clear stale queue row before entering online/queue check.
		if err := r.clearUserQueue(ctx, record.UID); err != nil {
			return result, err
		}

		forbidden, err := r.userForbiddenSeconds(ctx, record.UID)
		if err != nil {
			return result, err
		}
		if forbidden > 0 {
			result.Raw = []any{0, fmt.Sprintf("account_temp_locked:%d", forbidden)}
			return result, nil
		}

		locked, err := r.userLocked(ctx, record.UID, record.State)
		if err != nil {
			return result, err
		}
		if locked {
			result.Raw = []any{0, "account_locked"}
			return result, nil
		}
	}

	online, err := r.onlineUserCount(ctx)
	if err != nil {
		return result, err
	}
	maxOnline := r.stateValue(ctx, 4, 2000)
	queueSize, err := r.queueSize(ctx)
	if err != nil {
		return result, err
	}

	if int64(online) >= maxOnline || queueSize > 100 || (queueSize > 0 && int64(queueSize+online+200) >= maxOnline) {
		queueCount, err := r.enqueueUser(ctx, record.UID, sid, ip, now)
		if err != nil {
			return result, err
		}

		result.Raw = []any{1, 1, record.UID, sid, queueCount}
		result.UID = record.UID
		result.SID = sid
		result.Queued = true
		result.QueueCount = queueCount
		return result, nil
	}

	if err := r.applyLegacyRealLogin(ctx, record.UID, sid, ip, now); err != nil {
		return result, err
	}

	result.Raw = []any{1, 2, record.UID, sid}
	result.Logged = true
	result.UID = record.UID
	result.SID = sid
	return result, nil
}

func (r *Repository) LegacyCheckQueue(ctx context.Context, payload LegacyQueueCheckPayload, ip int64) (LegacyLoginResult, error) {
	result := LegacyLoginResult{}
	if payload.UID <= 0 || payload.SID <= 0 {
		result.Raw = []any{0}
		return result, nil
	}

	serverState := r.stateValue(ctx, 2, 1)
	if serverState != 1 {
		result.Raw = []any{0, "server_is_updating"}
		return result, nil
	}

	if r.db == nil {
		result.Raw = []any{1, 0}
		result.Logged = true
		result.UID = payload.UID
		result.SID = payload.SID
		return result, nil
	}

	queueID, found, err := r.findQueueRow(ctx, payload.UID, payload.SID, ip)
	if err != nil {
		return result, err
	}
	if !found {
		result.Raw = []any{0}
		return result, nil
	}

	online, err := r.onlineUserCount(ctx)
	if err != nil {
		return result, err
	}
	maxOnline := r.stateValue(ctx, 4, 2000)
	queueOrder, err := r.queueOrder(ctx, queueID)
	if err != nil {
		return result, err
	}

	if int64(online)+int64(queueOrder) < maxOnline {
		if _, err := r.db.ExecContext(ctx, "delete from mem_queue where id = ?", queueID); err != nil {
			return result, err
		}
		now := time.Now().Unix()
		if err := r.applyLegacyRealLogin(ctx, payload.UID, payload.SID, ip, now); err != nil {
			return result, err
		}

		result.Raw = []any{1, 0}
		result.Logged = true
		result.UID = payload.UID
		result.SID = payload.SID
		return result, nil
	}

	now := time.Now().Unix()
	if _, err := r.db.ExecContext(ctx, "update mem_queue set `lastupdate` = ? where id = ?", now, queueID); err != nil {
		return result, err
	}

	result.Raw = []any{1, 1, queueOrder}
	result.UID = payload.UID
	result.SID = payload.SID
	result.Queued = true
	result.QueueCount = queueOrder
	return result, nil
}

func (r *Repository) stateValue(ctx context.Context, state int, fallback int64) int64 {
	if r.db == nil {
		return fallback
	}

	var value sql.NullInt64
	if err := r.db.QueryRowContext(ctx, "select `value` from mem_state where `state` = ? limit 1", state).Scan(&value); err != nil {
		return fallback
	}
	if !value.Valid {
		return fallback
	}
	return value.Int64
}

func (r *Repository) serverStateMessage(ctx context.Context, serverState int) string {
	switch serverState {
	case 0:
		return r.announcementByID(ctx, 2, "server_not_start")
	case 2:
		return r.announcementByID(ctx, 3, "server_is_updating")
	default:
		return "server_is_updating"
	}
}

func (r *Repository) announcementByID(ctx context.Context, id int, fallback string) string {
	if r.db == nil {
		return fallback
	}

	var value sql.NullString
	if err := r.db.QueryRowContext(ctx, "select content from sys_announce where id = ? limit 1", id).Scan(&value); err != nil {
		return fallback
	}
	trimmed := strings.TrimSpace(value.String)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func (r *Repository) findLegacyUserByPassport(ctx context.Context, passport string, passType string) (legacyUserRecord, error) {
	if r.db == nil {
		user, err := r.SessionUserByPassport(ctx, passport)
		if err != nil {
			return legacyUserRecord{}, err
		}
		return legacyUserRecord{
			UID:      user.UID,
			Passport: user.Passport,
			PassType: user.PassType,
		}, nil
	}

	row := r.db.QueryRowContext(ctx, `
select uid, trim(coalesce(passport, '')), trim(coalesce(passtype, '')), state
from sys_user
where trim(coalesce(passport, '')) = ? and trim(coalesce(passtype, '')) = ?
order by uid asc
limit 1`, passport, passType)

	record := legacyUserRecord{}
	if err := row.Scan(&record.UID, &record.Passport, &record.PassType, &record.State); err != nil {
		return legacyUserRecord{}, err
	}
	return record, nil
}

func (r *Repository) createLegacyUser(ctx context.Context, passport string, passType string, sid int64, ip int64, now int64) (legacyUserRecord, error) {
	if r.db == nil {
		return legacyUserRecord{}, newInvalidError("legacy database is unavailable for account creation")
	}

	maxUserCount := r.stateValue(ctx, 100, 1000000)
	var userCount int64
	if err := r.db.QueryRowContext(ctx, "select count(*) from sys_user where uid > 1000").Scan(&userCount); err != nil {
		return legacyUserRecord{}, err
	}
	if userCount >= maxUserCount {
		return legacyUserRecord{}, newInvalidError("server_full")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return legacyUserRecord{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	insert, err := tx.ExecContext(ctx, `
insert into sys_user (passtype, passport, `+"`group`"+`, state, regtime, domainid, honour)
values (?, ?, '0', 3, ?, 0, 0)`, passType, passport, now)
	if err != nil {
		return legacyUserRecord{}, err
	}

	uid64, err := insert.LastInsertId()
	if err != nil {
		return legacyUserRecord{}, err
	}
	uid := int(uid64)

	if _, err := tx.ExecContext(ctx, "insert into sys_sessions (uid, sid, ip) values (?, ?, ?)", uid, sid, ip); err != nil {
		return legacyUserRecord{}, err
	}
	if _, err := tx.ExecContext(ctx, "insert into sys_online (uid, lastupdate, onlineupdate, onlinetime) values (?, ?, ?, 0)", uid, now, now); err != nil {
		return legacyUserRecord{}, err
	}
	if err := tx.Commit(); err != nil {
		return legacyUserRecord{}, err
	}

	return legacyUserRecord{
		UID:      uid,
		Passport: passport,
		PassType: passType,
		State:    3,
	}, nil
}

func (r *Repository) clearUserQueue(ctx context.Context, uid int) error {
	if r.db == nil {
		return nil
	}
	_, err := r.db.ExecContext(ctx, "delete from mem_queue where uid = ?", uid)
	return err
}

func (r *Repository) userForbiddenSeconds(ctx context.Context, uid int) (int64, error) {
	if r.db == nil {
		return 0, nil
	}

	var delta sql.NullInt64
	if err := r.db.QueryRowContext(ctx, "select login - unix_timestamp() from sys_user_forbidden where uid = ? limit 1", uid).Scan(&delta); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	if !delta.Valid || delta.Int64 <= 0 {
		return 0, nil
	}
	return delta.Int64, nil
}

func (r *Repository) userLocked(ctx context.Context, uid int, state int) (bool, error) {
	if state == 5 {
		return true, nil
	}
	if r.db == nil {
		return false, nil
	}

	var count int
	if err := r.db.QueryRowContext(ctx, "select count(*) from sys_user_state where uid = ? and forbiend > unix_timestamp()", uid).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) onlineUserCount(ctx context.Context) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	var count int
	if err := r.db.QueryRowContext(ctx, "select count(*) from sys_online where unix_timestamp() - lastupdate < 30").Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) queueSize(ctx context.Context) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	var count int
	if err := r.db.QueryRowContext(ctx, "select count(*) from mem_queue").Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) enqueueUser(ctx context.Context, uid int, sid int64, ip int64, now int64) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	insert, err := r.db.ExecContext(ctx, "insert into mem_queue (uid, sid, ip, lastupdate) values (?, ?, ?, ?)", uid, sid, ip, now)
	if err != nil {
		return 0, err
	}

	qid, err := insert.LastInsertId()
	if err != nil {
		return 0, err
	}

	var queueCount int
	if err := r.db.QueryRowContext(ctx, "select count(*) from mem_queue where id < ?", qid).Scan(&queueCount); err != nil {
		return 0, err
	}
	return queueCount, nil
}

func (r *Repository) applyLegacyRealLogin(ctx context.Context, uid int, sid int64, ip int64, now int64) error {
	if r.db == nil {
		return nil
	}

	if _, err := r.db.ExecContext(ctx, "insert into sys_sessions (uid, sid, ip) values (?, ?, ?) on duplicate key update sid = values(sid), ip = values(ip)", uid, sid, ip); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, "insert into sys_online (uid, lastupdate, onlineupdate, onlinetime) values (?, ?, ?, 0) on duplicate key update lastupdate = values(lastupdate), onlineupdate = values(onlineupdate)", uid, now, now); err != nil {
		return err
	}
	return nil
}

func (r *Repository) findQueueRow(ctx context.Context, uid int, sid int64, ip int64) (int64, bool, error) {
	if r.db == nil {
		return 0, false, nil
	}

	var queueID int64
	err := r.db.QueryRowContext(ctx, "select id from mem_queue where uid = ? and sid = ? and ip = ? limit 1", uid, sid, ip).Scan(&queueID)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return queueID, true, nil
}

func (r *Repository) queueOrder(ctx context.Context, queueID int64) (int, error) {
	if r.db == nil {
		return 0, nil
	}

	var queueOrder int
	if err := r.db.QueryRowContext(ctx, "select count(*) from mem_queue where id < ?", queueID).Scan(&queueOrder); err != nil {
		return 0, err
	}
	return queueOrder, nil
}
