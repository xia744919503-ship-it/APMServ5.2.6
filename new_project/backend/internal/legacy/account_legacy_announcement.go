package legacy

import (
	"context"
	"database/sql"
	"strings"
)

func (r *Repository) LegacyLoginAnnouncement(ctx context.Context) (string, error) {
	if r.db == nil {
		return "", nil
	}

	var raw sql.NullString
	if err := r.db.QueryRowContext(ctx, "select content from sys_announce where id = 1 limit 1").Scan(&raw); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(raw.String), nil
}
