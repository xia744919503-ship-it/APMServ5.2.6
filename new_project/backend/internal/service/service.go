package service

import (
	"context"
	"time"

	"rxsg-new-project/backend/internal/legacy"
)

type ProjectProfile struct {
	LegacyName      string   `json:"legacyName"`
	LegacyStack     []string `json:"legacyStack"`
	ModernStack     []string `json:"modernStack"`
	KeyEntrypoints  []string `json:"keyEntrypoints"`
	MigrationStages []string `json:"migrationStages"`
}

type DashboardOverview struct {
	Status         legacy.DatabaseStatus `json:"status"`
	Counts         legacy.Counts         `json:"counts"`
	Announcement   string                `json:"announcement"`
	FeaturedCities []legacy.CityCard     `json:"featuredCities"`
	Project        ProjectProfile        `json:"project"`
	SnapshotAt     time.Time             `json:"snapshotAt"`
}

type Service struct {
	repo *legacy.Repository
}

func New(repo *legacy.Repository) Service {
	return Service{repo: repo}
}

func (s Service) Health() map[string]any {
	return map[string]any{
		"ok":       true,
		"service":  "rxsg-modern-refactor-api",
		"database": s.repo.Status(),
		"time":     time.Now(),
	}
}

func (s Service) Overview(ctx context.Context) (DashboardOverview, error) {
	counts, err := s.repo.Summary(ctx)
	if err != nil {
		return DashboardOverview{}, err
	}

	announcement, err := s.repo.Announcement(ctx)
	if err != nil {
		return DashboardOverview{}, err
	}

	cities, err := s.repo.FeaturedCities(ctx, 8)
	if err != nil {
		return DashboardOverview{}, err
	}

	return DashboardOverview{
		Status:         s.repo.Status(),
		Counts:         counts,
		Announcement:   announcement,
		FeaturedCities: cities,
		Project: ProjectProfile{
			LegacyName: "Legacy RXSG (PHP + Flash + MySQL)",
			LegacyStack: []string{
				"PHP 5.x with mysql_* style procedural code",
				"Flash `BloodWar.swf` as the primary client shell",
				"AMFPHP and direct PHP scripts as the mixed API layer",
				"MySQL 5.1 with MyISAM tables",
				"APMServ Windows integrated runtime",
			},
			ModernStack: []string{
				"Vue 3 + TypeScript + Vite",
				"Pinia for page state",
				"Element Plus for management-heavy screens",
				"Legacy asset-driven DOM rendering for the world map scene",
				"Go as the new API and legacy adapter layer",
			},
			KeyEntrypoints: []string{
				"`www/htdocs/index.php` loads the Flash game shell",
				"`server/game/*.php` holds city, battle, troop and technology logic",
				"`server/config/db.php` defines the legacy database connection",
				"`MySQL5.1/data/bloodwar` stores the full game data set",
			},
			MigrationStages: []string{
				"1. Understand the legacy layout, database and main execution flow",
				"2. Build a read-first Go adapter instead of rewriting all PHP at once",
				"3. Rebuild overview, city and world scenes in Vue while reusing legacy assets",
				"4. Continue with login, city operations, troops, technology and battle modules",
			},
		},
		SnapshotAt: time.Now(),
	}, nil
}

func (s Service) Cities(ctx context.Context, limit int) ([]legacy.CityCard, error) {
	return s.repo.ListCities(ctx, limit)
}

func (s Service) CityDetail(ctx context.Context, cid int) (legacy.CityDetail, error) {
	return s.repo.CityDetail(ctx, cid)
}

func (s Service) WorldMap(ctx context.Context, cid int, radius int) (legacy.WorldMap, error) {
	return s.repo.WorldMap(ctx, cid, radius)
}

func (s Service) WorldMapAt(ctx context.Context, cid int, focusX int, focusY int, radius int) (legacy.WorldMap, error) {
	return s.repo.WorldMapAt(ctx, cid, focusX, focusY, radius)
}

func (s Service) CommanderOptions(ctx context.Context, limit int) ([]legacy.CommanderOption, error) {
	return s.repo.CommanderOptions(ctx, limit)
}

func (s Service) SessionUser(ctx context.Context, uid int) (legacy.SessionUser, error) {
	return s.repo.SessionUser(ctx, uid)
}

func (s Service) LoginByPassport(ctx context.Context, passport string, password string) (legacy.SessionUser, error) {
	return s.repo.LoginByPassport(ctx, passport, password)
}

func (s Service) LegacyDoLogin(ctx context.Context, payload legacy.LegacyLoginPayload, ip int64) (legacy.LegacyLoginResult, error) {
	return s.repo.LegacyDoLogin(ctx, payload, ip)
}

func (s Service) LegacyLoginAnnouncement(ctx context.Context) (string, error) {
	return s.repo.LegacyLoginAnnouncement(ctx)
}

func (s Service) LegacyCheckQueue(ctx context.Context, payload legacy.LegacyQueueCheckPayload, ip int64) (legacy.LegacyLoginResult, error) {
	return s.repo.LegacyCheckQueue(ctx, payload, ip)
}

func (s Service) LegacyCreateRole(ctx context.Context, payload legacy.LegacyRoleCreatePayload) (legacy.LegacyRoleCreateResult, error) {
	return s.repo.LegacyCreateRole(ctx, payload)
}

func (s Service) UserCities(ctx context.Context, uid int, limit int) ([]legacy.CityCard, error) {
	return s.repo.UserCities(ctx, uid, limit)
}

func (s Service) UserRelations(ctx context.Context, uid int) (legacy.RelationPage, error) {
	return s.repo.UserRelations(ctx, uid)
}

func (s Service) MyUnion(ctx context.Context, uid int) (legacy.UnionSnapshot, error) {
	return s.repo.MyUnion(ctx, uid)
}

func (s Service) CreateUnion(ctx context.Context, uid int, unionName string) (legacy.UnionSnapshot, error) {
	return s.repo.CreateUnion(ctx, uid, unionName)
}

func (s Service) ApplyJoinUnion(ctx context.Context, uid int, unionID int) (legacy.UnionSnapshot, error) {
	return s.repo.ApplyJoinUnion(ctx, uid, unionID)
}

func (s Service) CancelJoinUnionApply(ctx context.Context, uid int) (legacy.UnionSnapshot, error) {
	return s.repo.CancelJoinUnionApply(ctx, uid)
}

func (s Service) LeaveUnion(ctx context.Context, uid int) (legacy.UnionSnapshot, error) {
	return s.repo.LeaveUnion(ctx, uid)
}

func (s Service) UpdateUnionProfile(ctx context.Context, uid int, unionName string, intro string, announcement string) (legacy.UnionSnapshot, error) {
	return s.repo.UpdateUnionProfile(ctx, uid, unionName, intro, announcement)
}

func (s Service) SetUnionRelation(ctx context.Context, uid int, targetUnionID int, relationType int) (legacy.UnionSnapshot, error) {
	return s.repo.SetUnionRelation(ctx, uid, targetUnionID, relationType)
}

func (s Service) RemoveUnionRelation(ctx context.Context, uid int, targetUnionID int, relationType int) (legacy.UnionSnapshot, error) {
	return s.repo.RemoveUnionRelation(ctx, uid, targetUnionID, relationType)
}

func (s Service) MyTasks(ctx context.Context, uid int) (legacy.TaskSnapshot, error) {
	return s.repo.MyTasks(ctx, uid)
}

func (s Service) BattleTasksSnapshot(ctx context.Context, uid int, bidHint int, unionIDHint int) (legacy.TaskSnapshot, error) {
	return s.repo.BattleTasksSnapshot(ctx, uid, bidHint, unionIDHint)
}

func (s Service) ClaimTaskReward(ctx context.Context, uid int, taskID int) (legacy.TaskSnapshot, error) {
	return s.repo.ClaimTaskReward(ctx, uid, taskID)
}

func (s Service) MyShop(ctx context.Context, uid int) (legacy.ShopSnapshot, error) {
	return s.repo.MyShop(ctx, uid)
}

func (s Service) BuyShopItem(ctx context.Context, uid int, itemID int, count int, payType int, cityID int) (legacy.ShopSnapshot, error) {
	return s.repo.BuyShopItem(ctx, uid, itemID, count, payType, cityID)
}

func (s Service) MyCharge(ctx context.Context, uid int) (legacy.ChargeSnapshot, error) {
	return s.repo.MyCharge(ctx, uid)
}

func (s Service) ExchangeCharge(ctx context.Context, uid int, exchangeCount int) (legacy.ChargeSnapshot, error) {
	return s.repo.ExchangeCharge(ctx, uid, int64(exchangeCount))
}

func (s Service) LegacyPortal(ctx context.Context) (legacy.PortalSnapshot, error) {
	return s.repo.LegacyPortal(ctx)
}

func (s Service) GuidesByGroup(ctx context.Context, group int) (legacy.GuideGroup, error) {
	return s.repo.GuidesByGroup(ctx, group)
}

func (s Service) ActivityList(ctx context.Context) (legacy.ActivityList, error) {
	return s.repo.ActivityList(ctx)
}

func (s Service) UserTypeGoods(ctx context.Context, uid int, goodsType int, timeLeft int64) (legacy.UserTypeGoodsSnapshot, error) {
	return s.repo.UserTypeGoods(ctx, uid, goodsType, timeLeft)
}

func (s Service) UseUserTypeGoods(ctx context.Context, uid int, cid int, goodsType int, gid int) (legacy.CityDetail, error) {
	return s.repo.UseUserTypeGoods(ctx, uid, cid, goodsType, gid)
}

func (s Service) AddUserRelation(ctx context.Context, uid int, targetName string, relationType int) (legacy.RelationPage, error) {
	return s.repo.AddUserRelation(ctx, uid, targetName, relationType)
}

func (s Service) RemoveUserRelation(ctx context.Context, uid int, targetUID int, relationType int) (legacy.RelationPage, error) {
	return s.repo.RemoveUserRelation(ctx, uid, targetUID, relationType)
}

func (s Service) TouchLegacySession(ctx context.Context, uid int, sid int64, ip int64) error {
	return s.repo.TouchLegacySession(ctx, uid, sid, ip)
}

func (s Service) UpdateCityTax(ctx context.Context, uid int, cid int, tax int) (legacy.CityDetail, error) {
	return s.repo.UpdateCityTax(ctx, uid, cid, tax)
}

func (s Service) UpdateCityProduction(ctx context.Context, uid int, cid int, settings legacy.ProductionSettings) (legacy.CityDetail, error) {
	return s.repo.UpdateCityProduction(ctx, uid, cid, settings)
}

func (s Service) UpgradeCityBuilding(ctx context.Context, uid int, cid int, position int) (legacy.CityDetail, error) {
	return s.repo.UpgradeCityBuilding(ctx, uid, cid, position)
}

func (s Service) DestroyCityBuilding(ctx context.Context, uid int, cid int, position int) (legacy.CityDetail, error) {
	return s.repo.DestroyCityBuilding(ctx, uid, cid, position)
}

func (s Service) CancelCityBuildingAction(ctx context.Context, uid int, cid int, position int) (legacy.CityDetail, error) {
	return s.repo.CancelCityBuildingAction(ctx, uid, cid, position)
}

func (s Service) CityBuildingPlacementOptions(ctx context.Context, uid int, cid int, position int) (legacy.CityBuildingPlacementOptions, error) {
	return s.repo.CityBuildingPlacementOptions(ctx, uid, cid, position)
}

func (s Service) CityBuildingInfo(ctx context.Context, uid int, cid int, position int) (legacy.CityBuildingInfo, error) {
	return s.repo.CityBuildingInfo(ctx, uid, cid, position)
}

func (s Service) CreateCityBuilding(ctx context.Context, uid int, cid int, position int, bid int) (legacy.CityDetail, error) {
	return s.repo.CreateCityBuilding(ctx, uid, cid, position, bid)
}

func (s Service) BuildingSpeedGoods(ctx context.Context, uid int, cid int, position int) (legacy.BuildingSpeedGoodsSnapshot, error) {
	return s.repo.BuildingSpeedGoods(ctx, uid, cid, position)
}

func (s Service) UseBuildingSpeedGoods(ctx context.Context, uid int, cid int, position int, gid int) (legacy.CityDetail, error) {
	return s.repo.UseBuildingSpeedGoods(ctx, uid, cid, position, gid)
}

func (s Service) CityHeroes(ctx context.Context, uid int, cid int, limit int) (legacy.HeroRoster, error) {
	return s.repo.CityHeroes(ctx, uid, cid, limit)
}

func (s Service) UpdateCityChief(ctx context.Context, uid int, cid int, hid int) (legacy.HeroRoster, error) {
	return s.repo.UpdateCityChief(ctx, uid, cid, hid)
}

func (s Service) UpdateCityGeneral(ctx context.Context, uid int, cid int, hid int) (legacy.HeroRoster, error) {
	return s.repo.UpdateCityGeneral(ctx, uid, cid, hid)
}

func (s Service) UpdateCityCounsellor(ctx context.Context, uid int, cid int, hid int) (legacy.HeroRoster, error) {
	return s.repo.UpdateCityCounsellor(ctx, uid, cid, hid)
}

func (s Service) AddHeroPoint(ctx context.Context, uid int, cid int, hid int, stat string, amount int) (legacy.HeroRoster, error) {
	return s.repo.AddHeroPoint(ctx, uid, cid, hid, stat, amount)
}

func (s Service) RecruitCityHero(ctx context.Context, uid int, cid int, recruitID int) (legacy.HeroRoster, error) {
	return s.repo.RecruitCityHero(ctx, uid, cid, recruitID)
}

func (s Service) HeroArmorSnapshot(ctx context.Context, uid int, cid int, hid int) (legacy.HeroArmorSnapshot, error) {
	return s.repo.HeroArmorSnapshot(ctx, uid, cid, hid)
}

func (s Service) EquipHeroArmor(ctx context.Context, uid int, cid int, hid int, sid int, spart int) (legacy.HeroArmorSnapshot, error) {
	return s.repo.EquipHeroArmor(ctx, uid, cid, hid, sid, spart)
}

func (s Service) OffloadHeroArmor(ctx context.Context, uid int, cid int, hid int, spart int) (legacy.HeroArmorSnapshot, error) {
	return s.repo.OffloadHeroArmor(ctx, uid, cid, hid, spart)
}

func (s Service) RepairHeroArmor(ctx context.Context, uid int, cid int, hid int, sid int) (legacy.HeroArmorSnapshot, error) {
	return s.repo.RepairHeroArmor(ctx, uid, cid, hid, sid)
}

func (s Service) RepairAllHeroArmor(ctx context.Context, uid int, cid int, hid int, sids []int) (legacy.HeroArmorSnapshot, error) {
	return s.repo.RepairAllHeroArmor(ctx, uid, cid, hid, sids)
}

func (s Service) RenovateHeroArmor(ctx context.Context, uid int, cid int, hid int, sid int) (legacy.HeroArmorSnapshot, error) {
	return s.repo.RenovateHeroArmor(ctx, uid, cid, hid, sid)
}

func (s Service) RenovateAllHeroArmor(ctx context.Context, uid int, cid int, hid int, sids []int) (legacy.HeroArmorSnapshot, error) {
	return s.repo.RenovateAllHeroArmor(ctx, uid, cid, hid, sids)
}

func (s Service) RecycleHeroArmor(ctx context.Context, uid int, cid int, hid int, sid int) (legacy.HeroArmorSnapshot, error) {
	return s.repo.RecycleHeroArmor(ctx, uid, cid, hid, sid)
}

func (s Service) MyTroops(ctx context.Context, uid int, limit int) (legacy.TroopPage, error) {
	return s.repo.MyTroops(ctx, uid, limit)
}

func (s Service) DispatchCityTroop(ctx context.Context, uid int, cid int, targetCID int, soldiers map[int]int64, heroID int, task int, resource legacy.TroopResource) (legacy.TroopPage, error) {
	return s.repo.DispatchCityTroop(ctx, uid, cid, targetCID, soldiers, heroID, task, resource)
}

func (s Service) BattleFieldState(ctx context.Context, uid int, battlefieldID int, unionID int, cid int, fieldName string) (legacy.BattleFieldState, error) {
	return s.repo.BattleFieldState(ctx, uid, battlefieldID, unionID, cid, fieldName)
}

func (s Service) BattleQuitPreview(ctx context.Context, uid int) (legacy.BattleQuitPreview, error) {
	return s.repo.BattleQuitPreview(ctx, uid)
}

func (s Service) BattleTroopDetail(ctx context.Context, uid int, troopID int) (legacy.BattleTroopDetail, error) {
	return s.repo.BattleTroopDetail(ctx, uid, troopID)
}

func (s Service) BattleArmySendPreview(ctx context.Context, uid int, troopID int, targetCID int, targetName string) (legacy.BattleArmySendPreview, error) {
	return s.repo.BattleArmySendPreview(ctx, uid, troopID, targetCID, targetName)
}

func (s Service) BattleCampaignPreview(ctx context.Context, uid int, cid int, targetCID int, heroID int, soldiers map[int]int64, useFlag bool) (legacy.BattleCampaignPreview, error) {
	return s.repo.BattleCampaignPreview(ctx, uid, cid, targetCID, heroID, soldiers, useFlag)
}

func (s Service) BattleArmyAttackPreview(ctx context.Context, uid int, troopID int, targetTroopID int, targetName string) (legacy.BattleArmyAttackPreview, error) {
	return s.repo.BattleArmyAttackPreview(ctx, uid, troopID, targetTroopID, targetName)
}

func (s Service) BattlePatrolPreview(ctx context.Context, uid int, troopID int, targetTroopID int) (legacy.BattlePatrolPreview, error) {
	return s.repo.BattlePatrolPreview(ctx, uid, troopID, targetTroopID)
}

func (s Service) BattleMembersSnapshot(ctx context.Context, uid int) (legacy.BattleMembersSnapshot, error) {
	return s.repo.BattleMembersSnapshot(ctx, uid)
}

func (s Service) BattleFieldNewsPage(ctx context.Context, uid int, battlefieldID int, unionID int, page int, pageSize int) (legacy.BattleFieldNewsPage, error) {
	return s.repo.BattleFieldNewsPage(ctx, uid, battlefieldID, unionID, page, pageSize)
}

func (s Service) CityBarracksSnapshot(ctx context.Context, uid int, cid int, position int) (legacy.CityBarracksSnapshot, error) {
	return s.repo.CityBarracksSnapshot(ctx, uid, cid, position)
}

func (s Service) StartCitySoldierDraft(ctx context.Context, uid int, cid int, position int, sid int, count int) (legacy.CityBarracksSnapshot, error) {
	return s.repo.StartCitySoldierDraft(ctx, uid, cid, position, sid, count)
}

func (s Service) CancelCitySoldierDraft(ctx context.Context, uid int, cid int, position int, queueID int) (legacy.CityBarracksSnapshot, error) {
	return s.repo.CancelCitySoldierDraft(ctx, uid, cid, position, queueID)
}

func (s Service) CityResearchSnapshot(ctx context.Context, uid int, cid int, position int) (legacy.CityResearchSnapshot, error) {
	return s.repo.CityResearchSnapshot(ctx, uid, cid, position)
}

func (s Service) StartCityResearch(ctx context.Context, uid int, cid int, position int, tid int) (legacy.CityResearchSnapshot, error) {
	return s.repo.StartCityResearch(ctx, uid, cid, position, tid)
}

func (s Service) CancelCityResearch(ctx context.Context, uid int, cid int, position int, tid int) (legacy.CityResearchSnapshot, error) {
	return s.repo.CancelCityResearch(ctx, uid, cid, position, tid)
}

func (s Service) CallbackTroop(ctx context.Context, uid int, troopID int) (legacy.TroopPage, error) {
	return s.repo.CallbackTroop(ctx, uid, troopID)
}

func (s Service) MailPage(ctx context.Context, uid int, folder string, page int) (legacy.MailPage, error) {
	return s.repo.MailPage(ctx, uid, folder, page)
}

func (s Service) MailDetail(ctx context.Context, uid int, folder string, id int) (legacy.MailDetail, error) {
	return s.repo.MailDetail(ctx, uid, folder, id)
}

func (s Service) DeleteMail(ctx context.Context, uid int, folder string, ids []int, page int) (legacy.MailPage, error) {
	return s.repo.DeleteMail(ctx, uid, folder, ids, page)
}

func (s Service) SendMail(ctx context.Context, uid int, toName string, title string, content string) (legacy.MailDetail, error) {
	return s.repo.SendMail(ctx, uid, toName, title, content)
}

func (s Service) Reports(ctx context.Context, uid int, filter string, page int) (legacy.ReportPage, error) {
	return s.repo.Reports(ctx, uid, filter, page)
}

func (s Service) ReportDetail(ctx context.Context, uid int, id int) (legacy.ReportDetail, error) {
	return s.repo.ReportDetail(ctx, uid, id)
}

func (s Service) Ranking(ctx context.Context, kind string, page int) (legacy.RankingPage, error) {
	return s.repo.Ranking(ctx, kind, page)
}
