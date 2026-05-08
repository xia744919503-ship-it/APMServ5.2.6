package legacy

import (
	"context"
	"database/sql"
)

func (r *Repository) GuidesByGroup(ctx context.Context, group int) (GuideGroup, error) {
	if group <= 0 {
		return GuideGroup{}, newInvalidError("invalid guide group")
	}

	if r.db == nil {
		// No DB configured and no fixture data - return error instead of partial data
		return GuideGroup{}, newInvalidError("no database configured and no fixture data for guides")
	}

	rows, err := r.db.QueryContext(ctx, `
select gid, `+"`group`"+`, pregid, name, content, triggertype, coalesce(triggerdetails, ''), showpos, distype, coalesce(disdetails, '')
from cfg_guide
where `+"`group`"+` = ?
order by gid`, group)
	if err != nil {
		return GuideGroup{}, err
	}
	defer rows.Close()

	items := make([]Guide, 0, 16)
	for rows.Next() {
		var item Guide
		if err := rows.Scan(
			&item.GID,
			&item.Group,
			&item.PreGID,
			&item.Name,
			&item.Content,
			&item.TriggerType,
			&item.TriggerDetails,
			&item.ShowPos,
			&item.DisType,
			&item.DisDetails,
		); err != nil {
			return GuideGroup{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return GuideGroup{}, err
	}
	if len(items) == 0 {
		return GuideGroup{}, sql.ErrNoRows
	}

	return GuideGroup{Group: group, Items: items}, nil
}

func fixtureGuidesByGroup(group int) []Guide {
	if group != 1 {
		return []Guide{}
	}
	return []Guide{
		{
			GID:        6,
			Group:      1,
			PreGID:     5,
			Name:       "第6步，选择空地",
			Content:    "请您先选择一块风水宝地，然后在该“空地”上建造“民房”一间，私虽陋室，却蕴龙行。",
			ShowPos:    "570,130,100,60",
			DisType:    1,
			DisDetails: "",
		},
	}
}
