package legacy

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrDatabaseUnavailable = errors.New("legacy database is unavailable")

type Repository struct {
	db         *sql.DB
	status     DatabaseStatus
	worldTypes map[int]string
}

func NewRepository(dsn string) (*Repository, error) {
	repo := &Repository{
		status: DatabaseStatus{
			Connected: false,
			Mode:      "fixture",
			Source:    "bloodwar@127.0.0.1:3306",
			Message:   "Legacy database is offline, using fixture data.",
		},
		worldTypes: defaultWorldTypes(),
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return repo, err
	}
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(30 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return repo, err
	}

	repo.db = db
	repo.status = DatabaseStatus{
		Connected:   true,
		Mode:        "live",
		Source:      "bloodwar@127.0.0.1:3306",
		Message:     "Connected to the legacy bloodwar database.",
		ConnectedAt: time.Now(),
	}

	_ = repo.loadWorldTypes(ctx)

	return repo, nil
}

func (r *Repository) Close() error {
	if r.db == nil {
		return nil
	}

	return r.db.Close()
}

func (r *Repository) Status() DatabaseStatus {
	return r.status
}

func (r *Repository) Summary(ctx context.Context) (Counts, error) {
	if r.db == nil {
		return Counts{
			Users:         896,
			Cities:        6930,
			WorldTiles:    250000,
			ActiveTroops:  0,
			ActiveBattles: 0,
		}, nil
	}

	return Counts{
		Users:         r.count(ctx, "select count(*) from sys_user"),
		Cities:        r.count(ctx, "select count(*) from sys_city"),
		WorldTiles:    r.count(ctx, "select count(*) from mem_world"),
		ActiveTroops:  r.count(ctx, "select count(*) from sys_troops"),
		ActiveBattles: r.count(ctx, "select count(*) from sys_battle where state=0"),
	}, nil
}

func (r *Repository) Announcement(ctx context.Context) (string, error) {
	if r.db == nil {
		return "This is the modern refactor workspace for the legacy RXSG single-player server. The legacy database is currently offline, so fixture data is being shown.", nil
	}

	var content sql.NullString
	if err := r.db.QueryRowContext(ctx, "select content from sys_announce where id=1").Scan(&content); err != nil {
		return "", err
	}

	return content.String, nil
}

func (r *Repository) FeaturedCities(ctx context.Context, limit int) ([]CityCard, error) {
	return r.listCities(ctx, limit)
}

func (r *Repository) ListCities(ctx context.Context, limit int) ([]CityCard, error) {
	return r.listCities(ctx, limit)
}

func (r *Repository) CityDetail(ctx context.Context, cid int) (CityDetail, error) {
	if r.db == nil {
		return r.fixtureCityDetail(cid), nil
	}

	if err := r.settleCityBuildingQueue(ctx, cid); err != nil {
		return CityDetail{}, err
	}
	if err := r.settleCityDraftQueue(ctx, cid); err != nil {
		return CityDetail{}, err
	}
	if err := r.settleCityResearchQueue(ctx, cid); err != nil {
		return CityDetail{}, err
	}

	detail := CityDetail{}
	query := `
select
	c.cid,
	c.name,
	coalesce(u.name, ''),
	r.wood,
	r.rock,
	r.iron,
	r.food,
	r.gold,
	r.people,
	r.people_max,
	r.people_stable,
	r.people_building,
	r.food_max,
	r.wood_max,
	r.rock_max,
	r.iron_max,
	r.gold_max,
	r.morale,
	r.morale_stable,
	r.tax,
	r.complaint,
	r.gold_rate
from sys_city c
join mem_city_resource r on r.cid = c.cid
left join sys_user u on u.uid = c.uid
where c.cid = ?`

	var owner sql.NullString
	err := r.db.QueryRowContext(ctx, query, cid).Scan(
		&detail.Summary.CID,
		&detail.Summary.Name,
		&owner,
		&detail.Summary.Resources.Wood,
		&detail.Summary.Resources.Rock,
		&detail.Summary.Resources.Iron,
		&detail.Summary.Resources.Food,
		&detail.Summary.Resources.Gold,
		&detail.Summary.Resources.People,
		&detail.Summary.Resources.PeopleMax,
		&detail.Summary.Resources.PeopleStable,
		&detail.Summary.Resources.PeopleBuilding,
		&detail.Summary.Resources.FoodMax,
		&detail.Summary.Resources.WoodMax,
		&detail.Summary.Resources.RockMax,
		&detail.Summary.Resources.IronMax,
		&detail.Summary.Resources.GoldMax,
		&detail.Morale,
		&detail.MoraleStable,
		&detail.Tax,
		&detail.Complaint,
		&detail.GoldRate,
	)
	if err != nil {
		return CityDetail{}, err
	}

	detail.Summary.Owner = owner.String
	detail.Summary.X, detail.Summary.Y = coordinatesFromCID(detail.Summary.CID)

	production, err := r.loadCityProduction(ctx, cid)
	if err != nil {
		return CityDetail{}, err
	}
	detail.Production = production

	_ = r.db.QueryRowContext(ctx, "select count(*) from sys_city_hero where cid = ?", cid).Scan(&detail.HeroCount)

	buildings, err := r.queryBuildings(ctx, cid)
	if err != nil {
		return CityDetail{}, err
	}
	detail.Buildings = buildings

	soldiers, err := r.querySoldiers(ctx, cid)
	if err != nil {
		return CityDetail{}, err
	}
	detail.Soldiers = soldiers

	defences, err := r.queryDefences(ctx, cid)
	if err != nil {
		return CityDetail{}, err
	}
	detail.DefenceList = defences

	return detail, nil
}

func (r *Repository) WorldMap(ctx context.Context, cid int, radius int) (WorldMap, error) {
	if r.db == nil {
		return r.fixtureWorldMap(cid, radius), nil
	}

	center, err := r.cityCardByID(ctx, cid)
	if err != nil {
		return WorldMap{}, err
	}

	return r.worldMapAround(ctx, center, center.X, center.Y, radius)
}

func (r *Repository) WorldMapAt(ctx context.Context, cid int, focusX int, focusY int, radius int) (WorldMap, error) {
	if r.db == nil {
		return r.fixtureWorldMapAt(cid, focusX, focusY, radius), nil
	}

	center, err := r.cityCardByID(ctx, cid)
	if err != nil {
		return WorldMap{}, err
	}

	return r.worldMapAround(ctx, center, focusX, focusY, radius)
}

func (r *Repository) worldMapAround(ctx context.Context, center CityCard, focusX int, focusY int, radius int) (WorldMap, error) {
	radius = clamp(radius, 4, 14)
	x0 := clamp(focusX, 0, 999)
	y0 := clamp(focusY, 0, 999)

	requested := make([]int, 0, (radius*2+1)*(radius*2+1))
	wids := make([]int, 0, cap(requested))
	for y := y0 - radius; y <= y0+radius; y++ {
		for x := x0 - radius; x <= x0+radius; x++ {
			if x < 0 || y < 0 || x > 999 || y > 999 {
				continue
			}

			tileCID := y*1000 + x
			requested = append(requested, tileCID)
			wids = append(wids, cidToWid(tileCID))
		}
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(wids)), ",")
	query := fmt.Sprintf(`
select
	w.wid,
	w.type,
	w.ownercid,
	w.state,
	w.level,
	w.province,
	w.jun,
	coalesce(c.name, ''),
	coalesce(u.name, '')
from mem_world w
left join sys_city c on c.cid = w.ownercid
left join sys_user u on u.uid = c.uid
where w.wid in (%s)`, placeholders)

	args := make([]any, len(wids))
	for index, wid := range wids {
		args[index] = wid
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return WorldMap{}, err
	}
	defer rows.Close()

	tileMap := make(map[int]WorldTile, len(wids))
	for rows.Next() {
		var tile WorldTile
		var cityName sql.NullString
		var ownerName sql.NullString

		if err := rows.Scan(
			&tile.WID,
			&tile.Type,
			&tile.OwnerCID,
			&tile.State,
			&tile.Level,
			&tile.Province,
			&tile.Jun,
			&cityName,
			&ownerName,
		); err != nil {
			return WorldMap{}, err
		}

		tile.CID = widToCID(tile.WID)
		tile.X, tile.Y = coordinatesFromCID(tile.CID)
		tile.TypeName = r.worldTypes[tile.Type]
		tile.CityName = cityName.String
		tile.OwnerName = ownerName.String
		tileMap[tile.WID] = tile
	}

	tiles := make([]WorldTile, 0, len(requested))
	for _, tileCID := range requested {
		wid := cidToWid(tileCID)
		tile, exists := tileMap[wid]
		if !exists {
			x, y := coordinatesFromCID(tileCID)
			tile = WorldTile{
				WID:      wid,
				CID:      tileCID,
				X:        x,
				Y:        y,
				Type:     1,
				TypeName: r.worldTypes[1],
			}
		}
		tiles = append(tiles, tile)
	}

	return WorldMap{
		Center: center,
		FocusX: x0,
		FocusY: y0,
		Radius: radius,
		Tiles:  tiles,
	}, nil
}

func (r *Repository) listCities(ctx context.Context, limit int) ([]CityCard, error) {
	limit = clamp(limit, 1, 80)
	if r.db == nil {
		cities := r.fixtureCities()
		if limit < len(cities) {
			return cities[:limit], nil
		}
		return cities, nil
	}

	query := fmt.Sprintf(`
select
	c.cid,
	c.name,
	coalesce(u.name, ''),
	r.wood,
	r.rock,
	r.iron,
	r.food,
	r.gold,
	r.people,
	r.people_max
from sys_city c
join mem_city_resource r on r.cid = c.cid
left join sys_user u on u.uid = c.uid
order by (r.wood + r.rock + r.iron + r.food + r.gold) desc
limit %d`, limit)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cities := make([]CityCard, 0, limit)
	for rows.Next() {
		card := CityCard{}
		var owner sql.NullString
		if err := rows.Scan(
			&card.CID,
			&card.Name,
			&owner,
			&card.Resources.Wood,
			&card.Resources.Rock,
			&card.Resources.Iron,
			&card.Resources.Food,
			&card.Resources.Gold,
			&card.Resources.People,
			&card.Resources.PeopleMax,
		); err != nil {
			return nil, err
		}

		card.Owner = owner.String
		card.X, card.Y = coordinatesFromCID(card.CID)
		cities = append(cities, card)
	}

	return cities, rows.Err()
}

func (r *Repository) cityCardByID(ctx context.Context, cid int) (CityCard, error) {
	if r.db == nil {
		for _, city := range r.fixtureCities() {
			if city.CID == cid {
				return city, nil
			}
		}
		return r.fixtureCities()[0], nil
	}

	var card CityCard
	var owner sql.NullString
	err := r.db.QueryRowContext(ctx, `
select
	c.cid,
	c.name,
	coalesce(u.name, ''),
	r.wood,
	r.rock,
	r.iron,
	r.food,
	r.gold,
	r.people,
	r.people_max
from sys_city c
join mem_city_resource r on r.cid = c.cid
left join sys_user u on u.uid = c.uid
where c.cid = ?`, cid).Scan(
		&card.CID,
		&card.Name,
		&owner,
		&card.Resources.Wood,
		&card.Resources.Rock,
		&card.Resources.Iron,
		&card.Resources.Food,
		&card.Resources.Gold,
		&card.Resources.People,
		&card.Resources.PeopleMax,
	)
	if err != nil {
		return CityCard{}, err
	}

	card.Owner = owner.String
	card.X, card.Y = coordinatesFromCID(card.CID)
	return card, nil
}

func (r *Repository) queryBuildings(ctx context.Context, cid int) ([]Building, error) {
	rows, err := r.db.QueryContext(ctx, `
select b.bid, cb.name, b.level, b.xy, b.state, b.state_starttime, b.state_endtime
from sys_building b
join cfg_building cb on cb.bid = b.bid
where b.cid = ?
order by b.xy`, cid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	buildings := make([]Building, 0, 48)
	for rows.Next() {
		item := Building{}
		if err := rows.Scan(
			&item.BID,
			&item.Name,
			&item.Level,
			&item.Position,
			&item.State,
			&item.StateStartTime,
			&item.StateEndTime,
		); err != nil {
			return nil, err
		}
		item.Name = legacyBuildingName(item.BID, item.Name)
		buildings = append(buildings, item)
	}

	return buildings, rows.Err()
}

func (r *Repository) querySoldiers(ctx context.Context, cid int) ([]Soldier, error) {
	rows, err := r.db.QueryContext(ctx, `
select s.sid, cs.name, s.count
from sys_city_soldier s
join cfg_soldier cs on cs.sid = s.sid
where s.cid = ?
order by s.sid`, cid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	soldiers := make([]Soldier, 0, 16)
	for rows.Next() {
		item := Soldier{}
		if err := rows.Scan(&item.SID, &item.Name, &item.Count); err != nil {
			return nil, err
		}
		soldiers = append(soldiers, item)
	}

	return soldiers, rows.Err()
}

func (r *Repository) queryDefences(ctx context.Context, cid int) ([]Defence, error) {
	rows, err := r.db.QueryContext(ctx, `
select d.did, d.name, coalesce(cd.count, 0)
from cfg_defence d
left join sys_city_defence cd on cd.did = d.did and cd.cid = ?
order by d.did`, cid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	defences := make([]Defence, 0, 5)
	for rows.Next() {
		item := Defence{}
		if err := rows.Scan(&item.DID, &item.Name, &item.Count); err != nil {
			return nil, err
		}
		defences = append(defences, item)
	}

	return defences, rows.Err()
}

func (r *Repository) loadWorldTypes(ctx context.Context) error {
	if r.db == nil {
		return ErrDatabaseUnavailable
	}

	rows, err := r.db.QueryContext(ctx, "select type, name from cfg_world_type")
	if err != nil {
		return err
	}
	defer rows.Close()

	types := defaultWorldTypes()
	for rows.Next() {
		var worldType int
		var name string
		if err := rows.Scan(&worldType, &name); err != nil {
			return err
		}
		types[worldType] = name
	}

	r.worldTypes = types
	return rows.Err()
}

func (r *Repository) count(ctx context.Context, query string) int {
	if r.db == nil {
		return 0
	}

	var value int
	if err := r.db.QueryRowContext(ctx, query).Scan(&value); err != nil {
		return 0
	}

	return value
}

func coordinatesFromCID(cid int) (int, int) {
	return cid % 1000, cid / 1000
}

func cidToWid(cid int) int {
	y := cid / 1000
	x := cid % 1000
	return (y/10)*10000 + (x/10)*100 + (y%10)*10 + (x % 10)
}

func widToCID(wid int) int {
	y := (wid/10000)*10 + ((wid % 100) / 10)
	x := ((wid % 10000) / 100 * 10) + (wid % 10)
	return y*1000 + x
}

func clamp(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func defaultWorldTypes() map[int]string {
	return map[int]string{
		0: "City",
		1: "Plain",
		2: "Desert",
		3: "Forest",
		4: "Grassland",
		5: "Mountain",
		6: "Lake",
		7: "Swamp",
	}
}

func (r *Repository) fixtureCities() []CityCard {
	return []CityCard{
		{
			CID:   215265,
			Name:  "Luoyang",
			Owner: "Han Empire",
			X:     265,
			Y:     215,
			Resources: ResourceSnapshot{
				Wood: 21285000, Rock: 21285000, Iron: 21285000, Food: 218754100, Gold: 47053054, People: 94140, PeopleMax: 104600,
				PeopleStable: 94140, PeopleBuilding: 0,
				FoodMax: 300000000, WoodMax: 300000000, RockMax: 300000000, IronMax: 300000000, GoldMax: 100000000,
			},
		},
		{
			CID:   225185,
			Name:  "Chang'an",
			Owner: "Dong Zhuo",
			X:     185,
			Y:     225,
			Resources: ResourceSnapshot{
				Wood: 12285000, Rock: 12285000, Iron: 12285000, Food: 104212888, Gold: 47053054, People: 94140, PeopleMax: 104600,
				PeopleStable: 94140, PeopleBuilding: 0,
				FoodMax: 300000000, WoodMax: 300000000, RockMax: 300000000, IronMax: 300000000, GoldMax: 100000000,
			},
		},
		{
			CID:   165335,
			Name:  "Ye County",
			Owner: "Yuan Shao",
			X:     335,
			Y:     165,
			Resources: ResourceSnapshot{
				Wood: 12285000, Rock: 12285000, Iron: 12285000, Food: 54093059, Gold: 47053054, People: 94140, PeopleMax: 104600,
				PeopleStable: 94140, PeopleBuilding: 0,
				FoodMax: 300000000, WoodMax: 300000000, RockMax: 300000000, IronMax: 300000000, GoldMax: 100000000,
			},
		},
		{
			CID:   195295,
			Name:  "Huai County",
			Owner: "Sima Yi",
			X:     295,
			Y:     195,
			Resources: ResourceSnapshot{
				Wood: 11519397, Rock: 4810000, Iron: 9030000, Food: 29710908, Gold: 30350051, People: 51075, PeopleMax: 68100,
				PeopleStable: 51075, PeopleBuilding: 0,
				FoodMax: 300000000, WoodMax: 300000000, RockMax: 300000000, IronMax: 300000000, GoldMax: 100000000,
			},
		},
	}
}

func (r *Repository) fixtureCityDetail(cid int) CityDetail {
	cities := r.fixtureCities()
	selected := cities[0]
	for _, item := range cities {
		if item.CID == cid {
			selected = item
			break
		}
	}

	return CityDetail{
		Summary:      selected,
		Morale:       90,
		MoraleStable: 90,
		Tax:          10,
		Complaint:    0,
		GoldRate:     100,
		HeroCount:    6,
		Production:   ProductionState{Settings: ProductionSettings{FoodRate: 100, WoodRate: 100, RockRate: 100, IronRate: 100}, FoodAdd: 1000, FoodArmyUse: 0, WoodAdd: 1000, RockAdd: 500, IronAdd: 400, HeroFee: 0, GoldAdd: 50000, PeopleWorking: 100},
		Buildings: []Building{
			{BID: 2, Name: "Lumber Mill", Level: 20, Position: 1, State: 0},
			{BID: 4, Name: "Iron Mine", Level: 20, Position: 2, State: 0},
			{BID: 1, Name: "Farm", Level: 20, Position: 13, State: 0},
			{BID: 3, Name: "Quarry", Level: 20, Position: 21, State: 0},
			{BID: 6, Name: "Government Hall", Level: 20, Position: 81, State: 0},
			{BID: 20, Name: "Wall", Level: 20, Position: 199, State: 0},
		},
		Soldiers: []Soldier{
			{SID: 1, Name: "Worker", Count: 11974585},
			{SID: 2, Name: "Militia", Count: 537635},
			{SID: 3, Name: "Scout", Count: 6778869},
			{SID: 4, Name: "Spearman", Count: 2090801},
			{SID: 6, Name: "Archer", Count: 576037},
			{SID: 8, Name: "Heavy Cavalry", Count: 319509},
		},
		DefenceList: []Defence{
			{DID: 1, Name: "陷阱", Count: 0},
			{DID: 2, Name: "拒马", Count: 0},
			{DID: 3, Name: "箭塔", Count: 0},
			{DID: 4, Name: "滚木", Count: 0},
			{DID: 5, Name: "擂石", Count: 0},
		},
	}
}

func (r *Repository) fixtureWorldMap(cid int, radius int) WorldMap {
	center := r.fixtureCities()[0]
	for _, item := range r.fixtureCities() {
		if item.CID == cid {
			center = item
			break
		}
	}

	return r.fixtureWorldMapAt(center.CID, center.X, center.Y, radius)
}

func (r *Repository) fixtureWorldMapAt(cid int, focusX int, focusY int, radius int) WorldMap {
	center := r.fixtureCities()[0]
	for _, item := range r.fixtureCities() {
		if item.CID == cid {
			center = item
			break
		}
	}

	radius = clamp(radius, 4, 12)
	focusX = clamp(focusX, 0, 999)
	focusY = clamp(focusY, 0, 999)
	tiles := make([]WorldTile, 0, (radius*2+1)*(radius*2+1))

	for y := focusY - radius; y <= focusY+radius; y++ {
		for x := focusX - radius; x <= focusX+radius; x++ {
			if x < 0 || y < 0 {
				continue
			}

			tileCID := y*1000 + x
			tileType := 1 + ((x + y) % 6)
			tile := WorldTile{
				WID:      cidToWid(tileCID),
				CID:      tileCID,
				X:        x,
				Y:        y,
				Type:     tileType,
				TypeName: r.worldTypes[tileType],
				Level:    (x + y) % 10,
			}

			if x == center.X && y == center.Y {
				tile.Type = 0
				tile.TypeName = r.worldTypes[0]
				tile.OwnerCID = center.CID
				tile.CityName = center.Name
				tile.OwnerName = center.Owner
				tile.Level = 3
			}

			for _, city := range r.fixtureCities()[1:] {
				if abs(city.X-x) <= 1 && abs(city.Y-y) <= 1 && city.CID != center.CID {
					tile.Type = 0
					tile.TypeName = r.worldTypes[0]
					tile.OwnerCID = city.CID
					tile.CityName = city.Name
					tile.OwnerName = city.Owner
					tile.Level = 2
				}
			}

			tiles = append(tiles, tile)
		}
	}

	return WorldMap{
		Center: center,
		FocusX: focusX,
		FocusY: focusY,
		Radius: radius,
		Tiles:  tiles,
	}
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
