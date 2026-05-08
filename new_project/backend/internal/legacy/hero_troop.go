package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (r *Repository) CityHeroes(ctx context.Context, uid int, cid int, limit int) (HeroRoster, error) {
	limit = clamp(limit, 1, 24)

	if allowed, err := r.UserOwnsCity(ctx, uid, cid); err != nil {
		return HeroRoster{}, err
	} else if !allowed {
		return HeroRoster{}, ErrForbidden
	}

	if r.db == nil {
		return r.fixtureHeroRoster(cid), nil
	}

	city, err := r.cityCardByID(ctx, cid)
	if err != nil {
		return HeroRoster{}, err
	}

	query := fmt.Sprintf(`
select
	h.hid,
	h.uid,
	h.cid,
	case when trim(coalesce(h.name, '')) = '' then concat('HID ', h.hid) else h.name end as display_name,
	h.sex,
	h.face,
	h.state,
	h.level,
	h.loyalty,
	h.exp,
	h.command_base + h.command_add_on as command_stat,
	h.affairs_base + h.affairs_add + h.affairs_add_on as affairs_stat,
	h.bravery_base + h.bravery_add + h.bravery_add_on as bravery_stat,
	h.wisdom_base + h.wisdom_add + h.wisdom_add_on as wisdom_stat,
	h.affairs_add,
	h.bravery_add,
	h.wisdom_add,
	coalesce(m.force, 0),
	coalesce(m.force_max, 0),
	coalesce(m.energy, 0),
	coalesce(m.energy_max, 0)
from sys_city_hero h
left join mem_hero_blood m on m.hid = h.hid
where h.cid = ? and h.uid = ?
order by h.state asc, h.level desc, h.hid asc
limit %d`, limit)

	rows, err := r.db.QueryContext(ctx, query, cid, uid)
	if err != nil {
		return HeroRoster{}, err
	}
	defer rows.Close()

	items := make([]HeroCard, 0, limit)
	for rows.Next() {
		item := HeroCard{}
		if err := rows.Scan(
			&item.HID,
			&item.UID,
			&item.CID,
			&item.Name,
			&item.Sex,
			&item.Face,
			&item.State,
			&item.Level,
			&item.Loyalty,
			&item.Exp,
			&item.Command,
			&item.Affairs,
			&item.Bravery,
			&item.Wisdom,
			&item.AffairsAdd,
			&item.BraveryAdd,
			&item.WisdomAdd,
			&item.Force,
			&item.ForceMax,
			&item.Energy,
			&item.EnergyMax,
		); err != nil {
			return HeroRoster{}, err
		}

		item.Available = clamp(item.Level-item.AffairsAdd-item.BraveryAdd-item.WisdomAdd, 0, item.Level)
		item.StateName = heroStateDisplayLabel(item.State)
		item.StateLabel = item.StateName
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return HeroRoster{}, err
	}

	if len(items) == 0 {
		roster := r.fixtureHeroRoster(cid)
		roster.CID = city.CID
		roster.CityName = city.Name
		roster.Owner = city.Owner
		roster.Count = len(roster.Items)
		return roster, nil
	}

	hotelLevel := 0
	recruits := []HeroRecruit{}
	if level, hotelRecruits, err := r.cityHotelRecruitSnapshot(ctx, uid, cid); err != nil {
		return HeroRoster{}, err
	} else {
		hotelLevel = level
		recruits = hotelRecruits
	}

	officeLevel := 0
	if err := r.db.QueryRowContext(ctx, `
select coalesce(max(level), 0)
from sys_building
where cid = ? and bid = ?`, cid, officeBuildingID).Scan(&officeLevel); err != nil {
		return HeroRoster{}, err
	}
	recruitCapacity := officeLevel - len(items)
	if recruitCapacity < 0 {
		recruitCapacity = 0
	}

	return HeroRoster{
		CID:             city.CID,
		CityName:        city.Name,
		Owner:           city.Owner,
		Count:           len(items),
		Items:           items,
		HotelLevel:      hotelLevel,
		RecruitCapacity: recruitCapacity,
		Recruits:        recruits,
	}, nil
}

func (r *Repository) MyTroops(ctx context.Context, uid int, limit int) (TroopPage, error) {
	limit = clamp(limit, 1, 48)
	if r.db == nil {
		return r.fixtureTroopPage(uid), nil
	}
	if err := r.settleDueTroops(ctx, uid); err != nil {
		return TroopPage{}, err
	}

	soldierNames := r.loadSoldierNames(ctx)
	query := fmt.Sprintf(`
select
	t.id,
	t.uid,
	t.cid,
	t.startcid,
	t.targetcid,
	t.hid,
	t.task,
	t.state,
	t.starttime,
	t.pathtime,
	t.endtime,
	t.people,
	t.fooduse,
	t.soldiers,
	t.resource,
	coalesce(origin_city.name, ''),
	coalesce(target_city.name, ''),
	coalesce(h.name, ''),
	coalesce(h.level, 0)
from sys_troops t
left join sys_city origin_city on origin_city.cid = case when t.startcid > 0 then t.startcid else t.cid end
left join sys_city target_city on target_city.cid = t.targetcid
left join sys_city_hero h on h.hid = t.hid
where t.uid = ?
order by
	case when t.state < 4 then 0 else 1 end asc,
	t.endtime desc,
	t.id desc
limit %d`, limit)

	rows, err := r.db.QueryContext(ctx, query, uid)
	if err != nil {
		return TroopPage{}, err
	}
	defer rows.Close()

	page := TroopPage{
		Items: make([]TroopCard, 0, limit),
	}
	now := time.Now().Unix()
	for rows.Next() {
		item := TroopCard{}
		var fromCity sql.NullString
		var targetCity sql.NullString
		var heroName sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.UID,
			&item.CID,
			&item.StartCID,
			&item.TargetCID,
			&item.HeroID,
			&item.Task,
			&item.State,
			&item.StartTime,
			&item.PathTime,
			&item.EndTime,
			&item.People,
			&item.FoodUse,
			&item.SoldiersRaw,
			&item.ResourceRaw,
			&fromCity,
			&targetCity,
			&heroName,
			&item.HeroLevel,
		); err != nil {
			return TroopPage{}, err
		}

		item.FromCity = firstNonEmpty(fromCity.String, formatCIDLabel(firstPositive(item.StartCID, item.CID)))
		item.TargetCity = firstNonEmpty(targetCity.String, formatCIDLabel(item.TargetCID))
		item.HeroName = firstNonEmpty(heroName.String, "--")
		item.TaskLabel = troopTaskLabel(item.Task)
		item.StateLabel = troopStateLabel(item.State)
		item.Soldiers, item.SoldierCount = parseTroopSoldiers(item.SoldiersRaw, soldierNames)
		item.Resources = parseTroopResourcePayload(item.ResourceRaw)
		item.Resource = item.Resources
		if item.People <= 0 && item.SoldierCount > 0 {
			item.People = item.SoldierCount
		}
		if item.EndTime > now {
			item.SecondsLeft = item.EndTime - now
		}

		switch item.State {
		case 0:
			page.Moving++
		case 1:
			page.Returning++
		case 3:
			page.Battling++
		case 4:
			page.Stationed++
		case 5:
			page.Gathering++
		}

		page.Items = append(page.Items, item)
	}
	if err := rows.Err(); err != nil {
		return TroopPage{}, err
	}

	page.Total = len(page.Items)
	return page, nil
}

func (r *Repository) fixtureHeroRoster(cid int) HeroRoster {
	detail := r.fixtureCityDetail(cid)
	items := []HeroCard{
		{
			HID:        1001,
			CID:        detail.Summary.CID,
			Name:       "执政官",
			Sex:        1,
			Face:       3,
			State:      1,
			StateName:  heroStateDisplayLabel(1),
			StateLabel: heroStateDisplayLabel(1),
			Level:      48,
			Loyalty:    100,
			Exp:        2480000,
			Command:    132,
			Affairs:    168,
			Bravery:    104,
			Wisdom:     152,
			AffairsAdd: 18,
			Available:  30,
			Force:      100,
			ForceMax:   100,
			Energy:     100,
			EnergyMax:  100,
		},
		{
			HID:        1002,
			CID:        detail.Summary.CID,
			Name:       "先锋将",
			Sex:        1,
			Face:       6,
			State:      0,
			StateName:  heroStateDisplayLabel(0),
			StateLabel: heroStateDisplayLabel(0),
			Level:      45,
			Loyalty:    96,
			Exp:        2115000,
			Command:    154,
			Affairs:    88,
			Bravery:    176,
			Wisdom:     92,
			BraveryAdd: 20,
			Available:  25,
			Force:      96,
			ForceMax:   100,
			Energy:     82,
			EnergyMax:  100,
		},
		{
			HID:        1003,
			CID:        detail.Summary.CID,
			Name:       "军师",
			Sex:        0,
			Face:       3,
			State:      0,
			StateName:  heroStateDisplayLabel(0),
			StateLabel: heroStateDisplayLabel(0),
			Level:      44,
			Loyalty:    99,
			Exp:        1980000,
			Command:    118,
			Affairs:    149,
			Bravery:    82,
			Wisdom:     186,
			WisdomAdd:  16,
			Available:  28,
			Force:      88,
			ForceMax:   100,
			Energy:     100,
			EnergyMax:  100,
		},
	}

	return HeroRoster{
		CID:             detail.Summary.CID,
		CityName:        detail.Summary.Name,
		Owner:           detail.Summary.Owner,
		Count:           len(items),
		Items:           items,
		HotelLevel:      3,
		RecruitCapacity: 2,
		Recruits:        fixtureHeroRecruits(detail.Summary.CID),
	}
}

func (r *Repository) fixtureTroopPage(uid int) TroopPage {
	cities := r.fixtureCities()
	items := []TroopCard{
		{
			ID:          9001,
			UID:         uid,
			CID:         cities[0].CID,
			StartCID:    cities[0].CID,
			TargetCID:   cities[1].CID,
			FromCity:    cities[0].Name,
			TargetCity:  cities[1].Name,
			HeroID:      1002,
			HeroName:    "先锋将",
			HeroLevel:   45,
			Task:        3,
			TaskLabel:   troopTaskLabel(3),
			State:       0,
			StateLabel:  troopStateLabel(0),
			StartTime:   time.Now().Add(-5 * time.Minute).Unix(),
			EndTime:     time.Now().Add(18 * time.Minute).Unix(),
			PathTime:    1380,
			SecondsLeft: int64((18 * time.Minute) / time.Second),
			People:      12600,
			FoodUse:     84,
			SoldiersRaw: "4,6000,6,4200,8,2400,",
			ResourceRaw: "0,0,0,0,0",
			Resources:   TroopResource{},
			Resource:    TroopResource{},
			Soldiers: []Soldier{
				{SID: 4, Name: "Spearman", Count: 6000},
				{SID: 6, Name: "Archer", Count: 4200},
				{SID: 8, Name: "Heavy Cavalry", Count: 2400},
			},
			SoldierCount: 12600,
		},
		{
			ID:          9002,
			UID:         uid,
			CID:         cities[0].CID,
			StartCID:    cities[0].CID,
			TargetCID:   cities[0].CID,
			FromCity:    cities[0].Name,
			TargetCity:  cities[0].Name,
			HeroID:      1001,
			HeroName:    "执政官",
			HeroLevel:   48,
			Task:        1,
			TaskLabel:   troopTaskLabel(1),
			State:       4,
			StateLabel:  troopStateLabel(4),
			StartTime:   time.Now().Add(-90 * time.Minute).Unix(),
			EndTime:     0,
			PathTime:    0,
			SecondsLeft: 0,
			People:      8800,
			FoodUse:     56,
			SoldiersRaw: "2,3000,4,2800,6,3000,",
			ResourceRaw: "0,0,0,0,0",
			Resources:   TroopResource{},
			Resource:    TroopResource{},
			Soldiers: []Soldier{
				{SID: 2, Name: "Militia", Count: 3000},
				{SID: 4, Name: "Spearman", Count: 2800},
				{SID: 6, Name: "Archer", Count: 3000},
			},
			SoldierCount: 8800,
		},
	}

	return TroopPage{
		Total:     len(items),
		Moving:    1,
		Returning: 0,
		Stationed: 1,
		Battling:  0,
		Gathering: 0,
		Items:     items,
	}
}

func (r *Repository) loadSoldierNames(ctx context.Context) map[int]string {
	if r.db == nil {
		return map[int]string{}
	}

	rows, err := r.db.QueryContext(ctx, "select sid, name from cfg_soldier")
	if err != nil {
		return map[int]string{}
	}
	defer rows.Close()

	names := make(map[int]string, 16)
	for rows.Next() {
		var sid int
		var name string
		if err := rows.Scan(&sid, &name); err != nil {
			return map[int]string{}
		}
		names[sid] = name
	}
	return names
}

func parseTroopSoldiers(raw string, soldierNames map[int]string) ([]Soldier, int64) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "0" {
		return []Soldier{}, 0
	}

	parts := strings.Split(raw, ",")
	items := make([]Soldier, 0, len(parts)/2)
	var total int64
	for index := 0; index+1 < len(parts); index += 2 {
		sidText := strings.TrimSpace(parts[index])
		countText := strings.TrimSpace(parts[index+1])
		if sidText == "" || countText == "" {
			continue
		}

		sid, err := strconv.Atoi(sidText)
		if err != nil {
			continue
		}
		count, err := strconv.ParseInt(countText, 10, 64)
		if err != nil {
			continue
		}

		items = append(items, Soldier{
			SID:   sid,
			Name:  firstNonEmpty(soldierNames[sid], fmt.Sprintf("SID %d", sid)),
			Count: count,
		})
		total += count
	}

	return items, total
}

func heroStateLabel(state int) string {
	switch state {
	case 0:
		return "待命"
	case 1:
		return "在任"
	case 2:
		return "回城"
	case 3:
		return "交战"
	case 4:
		return "在野"
	case 5:
		return "驻守"
	case 7:
		return "守城"
	default:
		return fmt.Sprintf("状态 %d", state)
	}
}

func troopTaskLabel(task int) string {
	switch task {
	case 0:
		return "运输"
	case 1:
		return "驻军"
	case 2:
		return "侦察"
	case 3:
		return "掠夺"
	case 4:
		return "占领"
	case 5:
		return "防守"
	case 7:
		return "战场调动"
	case 8:
		return "战场支援"
	case 9:
		return "战场追击"
	default:
		return fmt.Sprintf("任务 %d", task)
	}
}

func troopStateLabel(state int) string {
	switch state {
	case 0:
		return "行军"
	case 1:
		return "返程"
	case 2:
		return "待战"
	case 3:
		return "交战"
	case 4:
		return "驻扎"
	case 5:
		return "采集"
	default:
		return fmt.Sprintf("状态 %d", state)
	}
}

func formatCIDLabel(cid int) string {
	if cid <= 0 {
		return "--"
	}

	x, y := coordinatesFromCID(cid)
	return fmt.Sprintf("[%d,%d]", x, y)
}

func firstPositive(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
