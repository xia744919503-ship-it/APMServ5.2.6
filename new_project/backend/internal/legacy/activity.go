package legacy

import (
	"context"
	"database/sql"
)

func (r *Repository) ActivityList(ctx context.Context) (ActivityList, error) {
	if r.db == nil {
		return ActivityList{Items: fixtureActivityItems()}, nil
	}

	rows, err := r.db.QueryContext(ctx, "select `content`, `link`, `interval` from sys_activity where inuse=1 order by id")
	if err != nil {
		return ActivityList{}, err
	}
	defer rows.Close()

	items := make([]ActivityItem, 0, 8)
	for rows.Next() {
		var item ActivityItem
		if err := rows.Scan(&item.Content, &item.Link, &item.Interval); err != nil {
			return ActivityList{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return ActivityList{}, err
	}
	if len(items) == 0 {
		return ActivityList{}, sql.ErrNoRows
	}

	return ActivityList{Items: items}, nil
}

func fixtureActivityItems() []ActivityItem {
	return []ActivityItem{
		{Content: "暂无活动公告", Link: "", Interval: "1800"},
	}
}
