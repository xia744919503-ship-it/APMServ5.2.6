<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import {
  battleArmyAttackPreview,
  battleCampaignPreview,
  battleArmySendPreview,
  battleFieldState,
  battleMembers,
  battleNews,
  battlePatrolPreview,
  battleQuitPreview,
  battleTroopDetail,
  applyUnion,
  buildingInfo,
  buildingOptions,
  buildingSpeedGoods,
  cancelBuildingAction,
  cityDetail,
  cityHeroes,
  cityResearchSnapshot,
  commanderOptions,
  cancelUnionApply,
  createBuilding,
  createRole,
  currentUser,
  dashboardOverview,
  deleteMail,
  dispatchCityTroop,
  destroyBuilding,
  legacyActivities,
  legacyGuides,
  legacyLogin,
  mailDetail,
  mailPage,
  myCharge,
  myCities,
  myRelations,
  myShop,
  myTroops,
  myUnion,
  rankings,
  recruitHero,
  reports,
  reportDetail,
  researchTech,
  trainTroop,
  upgradeBuilding,
  useBuildingSpeedGoods,
  useUserTypeGoods,
  userTypeGoods,
  worldMap,
  buyShopItem,
  type ActivityItem,
  type ActivityList,
  type BattleArmyAttackPreview,
  type BattleCampaignPreview,
  type BattleArmySendPreview,
  type BattleFieldCurrentTroop,
  type BattleFieldNewsItem,
  type BattleFieldNewsPage,
  type BattleFieldState,
  type BattleFieldTroopRow,
  type BattleMembersSnapshot,
  type BattlePatrolPreview,
  type BattleTroopDetail,
  type Building,
  type BuildingInfoResult,
  type BuildingOption,
  type BuildingOptionsResult,
  type BuildingSpeedGoodsResult,
  type CityCard,
  type CityDetail,
  type CityListResult,
  type CityResearchSnapshot,
  type CommanderOption,
  type DashboardOverview,
  type Guide,
  type Hero,
  type HeroRoster,
  type LoginResult,
  type MailItem,
  type MailPage,
  type RankItem,
  type RelationCard,
  type RelationPage,
  type ReportItem,
  type ReportPage,
  type SessionUser,
  type ShopItem,
  type ShopSnapshot,
  type SpeedGoods,
  type TechItem,
  type TroopPage,
  type TroopType,
  type UnionSnapshot,
  type UserTypeGoodsItem,
  type WorldMap,
  type ChargeSnapshot
} from "./api";
import {
  asset,
  buildingImageByBid,
  buildingIntroByBid,
  cityHallPlot,
  cityWallPlot,
  innerGridPoint,
  innerPlots,
  isInnerCityHallGrid,
  isInnerWallGrid,
  outerPlotsForGovernmentLevel,
  provinceAssets,
  type Plot
} from "./assets";
import {
  battleTasks,
  claimTaskReward,
  myTasks,
  type TaskCard,
  type TaskCategory,
  type TaskSnapshot
} from "./api";

type Screen = "boot" | "login" | "create-role" | "city";
type CityView = "inner" | "outer";
type LeftInfoTab = "resource" | "commander" | "army" | "defence";

const STAGE_WIDTH = 1000;
const STAGE_HEIGHT = 600;
const CITY_VIEW_LEFT = 264;
const CITY_VIEW_TOP = 0;
const CITY_MAP_WIDTH = 731;
const CITY_MAP_HEIGHT = 550;
const GUIDE_TIP_WIDTH = 468;
const GUIDE_ARROW_WIDTH = 76;
const GUIDE_ARROW_HEIGHT = 76;
const GUIDE_TIP_MARGIN = 10;
const GUIDE_TEXT_CHARS_PER_LINE = 30;
const GUIDE_TEXT_LINE_HEIGHT = 23;
const GUIDE_TEXT_MIN_HEIGHT = 44;

const screen = ref<Screen>("boot");
const cityView = ref<CityView>("inner");
const leftInfoTab = ref<LeftInfoTab>("resource");
const cityDropdownOpen = ref(false);
const stageScale = ref(1);
const loading = ref(false);
const error = ref("");
const LEFT_INFO_TIP_SEPARATOR = "────────────";
type LeftInfoTipRow = { label: string; value: string; color?: string };
type LeftInfoTipInputRow = { label: string; value: string | number; danger?: boolean; color?: string } | { separator: true };
const leftInfoTip = ref({ visible: false, x: 45, y: 39, title: "", rows: [] as LeftInfoTipRow[] });
const flashToolTip = ref({ visible: false, x: 0, y: 0, text: "" });
const user = ref<SessionUser | null>(null);
const lastLogin = ref<LoginResult | null>(null);
const loginCommanders = ref<CommanderOption[]>([]);
const loginDashboard = ref<DashboardOverview | null>(null);
const loginSnapshotMessage = ref("读取中...");
const city = ref<CityDetail | null>(null);
const cityList = ref<CityCard[]>([]);
const showBuildingLevels = ref(false);
const showTopBuildingList = ref(false);
const buildingListDialogVisible = ref(false);
const craftDialogVisible = ref(false);
const tacticDialogVisible = ref(false);
const lordDialogVisible = ref(false);
const equipmentDialogVisible = ref(false);
const treasureDialogVisible = ref(false);
const activeBottomFunction = ref("");
const hoveredBottomFunction = ref("");
const chatChannelMenuVisible = ref(false);
const chatChannelLabel = ref("世界");
const chatChannelItems = ["世界", "联盟", "私聊", "战场"];
const chatSendDialogVisible = ref(false);
const chatControlDialogVisible = ref(false);
const hoveredTopButton = ref("");
const pressedTopButton = ref("");
const topButtonOverAssets = new Set(["innercity", "outercity", "map", "battle"]);
const topButtonHoverOnAssets = new Set(["hero", "army", "union", "mission"]);
const flashConfirm = ref<{
  visible: boolean;
  title: string;
  message: string;
  resolve: ((value: boolean) => void) | null;
}>({
  visible: false,
  title: "确认",
  message: "",
  resolve: null
});
const guideVisible = ref(false);
const guideItems = ref<Guide[]>([]);
const currentGuide = ref<Guide | null>(null);
const selectedProvince = ref(10);
const hoveredProvince = ref<number | null>(null);
const selectedPlot = ref<Plot | null>(null);
const hoveredInnerPlot = ref<Plot | null>(null);
const hoveredAnnounce = ref(false);
const buildPanel = ref<BuildingOptionsResult | null>(null);
const buildingPanel = ref<Building | null>(null);
const buildingInfoPanel = ref<BuildingInfoResult | null>(null);
const speedGoodsPanel = ref<BuildingSpeedGoodsResult | null>(null);
const useGoodsPanel = ref({
  visible: false,
  type: 0,
  loading: false,
  error: "",
  goodsList: [] as UserTypeGoodsItem[]
});
const selectedBid = ref<number | null>(null);
const taskSnapshot = ref<TaskSnapshot | null>(null);
const taskPanelVisible = ref(false);
const selectedTaskId = ref<number | null>(null);
const selectedTaskCategoryType = ref<number | null>(null);
const selectedTaskGroupId = ref<number | null>(null);
const BATTLE_TASK_CATEGORY_TYPE = 3;
const announcementVisible = ref(false);
const buildingTip = ref({ visible: false, x: 45, y: 39, text: "" });
const governmentSubDialog = ref<"" | "tax" | "name" | "pacify" | "levy" | "produce" | "fields" | "cities">("");
const governmentTaxValue = ref(0);
const governmentCityName = ref("");
const governmentPacifyAction = ref("赈灾");
const governmentLevyResource = ref("黄金");
const governmentCityItems = ref<CityCard[]>([]);
const cityPopulation = computed(() => city.value?.summary.resources.people ?? 0);
const selectedCityId = computed(() => city.value?.summary.cid ?? user.value?.defaultCid ?? cityList.value[0]?.cid ?? 0);
const governmentTaxPreviewText = computed(() => {
  const population = cityPopulation.value;
  const income = Math.floor(population * Math.max(0, governmentTaxValue.value) * 0.01);
  return `每小时可征收黄金${income}。过于沉重的税收将导致民心下降，请慎重调节税率。`;
});
const governmentPacifyText = computed(() => {
  const population = cityPopulation.value;
  const foodCost = Math.floor(population * 0.1);
  const goldCost = Math.floor(population * 0.1);
  if (governmentPacifyAction.value === "赈灾") return `赈灾消耗粮食${foodCost}，提升民心5，减少民怨15。`;
  if (governmentPacifyAction.value === "祈福") return `祈福消耗黄金${goldCost}，提升民心25，减少民怨5。`;
  if (governmentPacifyAction.value === "祭天") return `祭天消耗粮食${foodCost}，消耗黄金${goldCost}，避免1次天灾，有机会增加1次天赐。`;
  return `增丁消耗粮食${foodCost}，增加人口${Math.floor(population * 0.05)}。不超过人口上限。`;
});
const governmentLevyPreview = computed(() => {
  const population = cityPopulation.value;
  const amount = Math.floor(population * 0.1);
  if (governmentLevyResource.value === "黄金") return `可征收黄金${amount}。`;
  if (governmentLevyResource.value === "粮食") return `可征收粮食${amount}。`;
  if (governmentLevyResource.value === "木材") return `可征收木材${amount}。`;
  if (governmentLevyResource.value === "石料") return `可征收石料${amount}。`;
  return `可征收铁锭${amount}。`;
});
const governmentProduceRows = computed(() => {
  const summary = city.value?.summary.resources;
  return [
    { label: "粮食", value: summary?.food ?? 0, rate: "25", title: "资源生产" },
    { label: "木材", value: summary?.wood ?? 0, rate: "25", title: "资源生产" },
    { label: "石料", value: summary?.rock ?? 0, rate: "25", title: "资源生产" },
    { label: "铁锭", value: summary?.iron ?? 0, rate: "25", title: "资源生产" }
  ];
});
const governmentFieldRows = computed(() => {
  const baseX = city.value?.summary.x ?? 250;
  const baseY = city.value?.summary.y ?? 250;
  return [
    { type: "农田", field: "平地", level: "1", x: baseX + 1, y: baseY, state: "和平" },
    { type: "木场", field: "森林", level: "1", x: baseX, y: baseY + 1, state: "和平" },
    { type: "石场", field: "山地", level: "2", x: baseX + 1, y: baseY + 1, state: "和平" },
    { type: "铁矿", field: "荒漠", level: "2", x: baseX - 1, y: baseY + 1, state: "和平" }
  ].map((row) => ({ ...row, position: `(${row.x},${row.y})` }));
});

// Dialog refs
const mailPanelVisible = ref(false);
const mailFolder = ref<"inbox" | "sent" | "sys" | "compose">("inbox");
const mailItems = ref<MailItem[]>([]);
const mailPageInfo = ref({ total: 0, page: 1, pageSize: 10 });
const mailDetailView = ref<MailItem | null>(null);

type ReportFilter = "all" | "attack" | "defend" | "scout" | "unread";
const reportPanelVisible = ref(false);
const reportFilter = ref<ReportFilter>("all");
const reportItems = ref<ReportItem[]>([]);
const reportPageInfo = ref({ total: 0, page: 1, pageSize: 10 });
const reportDetailView = ref<ReportItem | null>(null);

const shopPanelVisible = ref(false);
const shopSnapshot = ref<ShopSnapshot | null>(null);
const selectedShopGroupId = ref<number | null>(null);
const friendDialogVisible = ref(false);
const friendRelationTab = ref<"friends" | "blacklist">("friends");
const relationSnapshot = ref<RelationPage | null>(null);
const statDialogVisible = ref(false);
const forumDialogVisible = ref(false);
const websiteDialogVisible = ref(false);
const helpDialogVisible = ref(false);
const chargeDialogVisible = ref(false);
const chargeSnapshot = ref<ChargeSnapshot | null>(null);

const worldMapPanelVisible = ref(false);
const worldMapData = ref<WorldMap | null>(null);
const worldMapCenter = ref<WorldMap["center"] | null>(null);
const worldMapInputX = ref("");
const worldMapInputY = ref("");
const worldMapMode = ref<"city" | "map">("city");
const worldMapAlert = ref("");
const WORLD_OCCUPY_DISABLED_NOTE = "占领接口未接入，暂不可执行。";
type WorldGridInfo = {
  title: string;
  text: string;
  city: boolean;
  empty: boolean;
  targetCid: number;
};
const worldGridTip = ref<WorldGridInfo & { visible: boolean; x: number; y: number }>({ visible: false, x: 0, y: 0, title: "", text: "", city: false, empty: false, targetCid: 0 });
const selectedWorldGrid = ref<(WorldGridInfo & { x: number; y: number }) | null>(null);
type WorldTerrainTile = {
  key: string;
  x: number;
  y: number;
  mx: number;
  my: number;
  image: string;
  title: string;
};
const WORLD_GRID_CENTER_X = 314;
const WORLD_GRID_CENTER_Y = 250;
const WORLD_GRID_OFFSET_X = 54;
const WORLD_GRID_OFFSET_Y = 27;
const WORLD_GRID_IMAGE_HEIGHT = 75;
const WORLD_GRID_LOAD_RANGE = 9;
const WORLD_MAP_MIN_X = 1;
const WORLD_MAP_MAX_X = 500;
const WORLD_MAP_MIN_Y = 1;
const WORLD_MAP_MAX_Y = 500;
const WORLD_MOVE_VERT_STRAIGHT = 4;
const WORLD_MOVE_HORI_STRAIGHT = 3;
const WORLD_MOVE_OBLIQUE = 5;
const WORLD_TERRAIN_IMAGES = ["land", "grass", "forest", "hill", "land", "grass", "desert", "lake", "swamp"];
const WORLD_TERRAIN_NAMES = ["空地", "平地", "森林", "湖泊", "山地", "荒漠", "沼泽", "草原"];
const WORLD_STATE_NAMES = ["正常", "免战", "保护", "交战"];

const unionPanelVisible = ref(false);
const unionSnapshot = ref<UnionSnapshot | null>(null);

const heroPanelVisible = ref(false);
const heroRoster = ref<HeroRoster | null>(null);

const barracksPanelVisible = ref(false);
const troopsData = ref<TroopPage | null>(null);
const cityTroopItems = computed<TroopType[]>(() =>
  city.value?.soldiers?.map((item) => ({
    tid: item.sid,
    name: item.name,
    count: item.count,
    injured: 0
  })) ?? []
);
const barracksTroopItems = computed(() => troopsData.value?.troops.length ? troopsData.value.troops : cityTroopItems.value);

const collegePanelVisible = ref(false);
const researchSnapshot = ref<CityResearchSnapshot | null>(null);

type RankKind = "power" | "level" | "city";
const rankPanelVisible = ref(false);
const rankKind = ref<RankKind>("power");
const rankItems = ref<RankItem[]>([]);
const rankPageInfo = ref({ total: 0, page: 1, pageSize: 20 });

const battlePanelVisible = ref(false);
const battleMenuVisible = ref(false);
const battleCampaignVisible = ref(false);
const battleCampaignMode = ref<"world" | "battle">("world");
const battleFieldViewVisible = ref(false);
const battleActionVisible = ref(false);
const battleAttackVisible = ref(false);
const battlePatrolVisible = ref(false);
const battleTroopViewVisible = ref(false);
const battleTroopDetailVisible = ref(false);
const battleInfoVisible = ref(false);
const battleUsersVisible = ref(false);
const battleInfoTab = ref<"info" | "news">("info");
const battleMapId = ref(1001);
const battleFieldSnapshot = ref<BattleFieldState | null>(null);
const selectedBattleFieldRow = ref<BattleFieldTroopRow | null>(null);
const selectedBattleCurrentTroop = ref<BattleFieldCurrentTroop | null>(null);
const selectedBattleTroopDetail = ref<BattleTroopDetail | null>(null);
const battleSendPreviewData = ref<BattleArmySendPreview | null>(null);
const battleCampaignPreviewData = ref<BattleCampaignPreview | null>(null);
const battleAttackPreviewData = ref<BattleArmyAttackPreview | null>(null);
const battlePatrolPreviewData = ref<BattlePatrolPreview | null>(null);
const battleMembersSnapshot = ref<BattleMembersSnapshot | null>(null);
let battleCampaignPreviewSeq = 0;
const battleCampaignUseFlag = ref(false);
const battleCampaignTargetId = ref("1");
const battleCampaignHeroId = ref("0");
const battleCampaignFieldName = ref("前线战场");
const battleCampaignPathNeedTime = ref("0:30");
const battleCampaignArriveTime = ref("0:30");
const battleCampaignStartReason = ref("");
const battleCampaignTask = ref<2 | 3>(3);
const battleInfoPage = ref(1);
const battleNewsTotal = ref(0);
const battleNewsPageSize = 10;
const battleNewsItems = ref<{ time: string; evtContent: string; color: number }[]>([]);
const battleCampaignSoldiers = ref<{ sid: number; name: string; count: number; takecount: number }[]>([]);
const battleCampaignTargets = ref<{ id: string; name: string }[]>([]);
const battleCampaignHeroes = ref([
  { id: "0", heroname: "请选择" }
]);
const battleHeroSlots = [
  { index: 0, title: "查看旧服战场部队详情；召回、加速、撤退仍为只读。" },
  { index: 1, title: "查看旧服战场部队详情；召回、加速、撤退仍为只读。" }
];
const battleFieldObjects = [
  { key: "left-banner", className: "battle-field-object left-banner" },
  { key: "right-banner", className: "battle-field-object right-banner" },
  { key: "center-marker", className: "battle-field-object center-marker" }
];
const battleFieldRows = computed(() => battleFieldSnapshot.value?.rows ?? []);
const battleCurrentTroops = computed(() => battleFieldSnapshot.value?.currentTroops ?? []);
const battleHeroSlotViews = computed(() =>
  battleHeroSlots.map((slot) => {
    const troop = battleCurrentTroops.value[slot.index];
    return {
      ...slot,
      troop,
      label: troop?.heroName ? troop.heroName.slice(0, 1) : "",
      image: troop ? battleHeroImage(troop) : ""
    };
  })
);
const battleActiveMapId = computed(() => battleFieldSnapshot.value?.bid || battleMapId.value);
const battleWinPointWidth = (index: number) => {
  const point = Math.min(Math.max(Number(battleFieldSnapshot.value?.winPoints?.[index]?.point ?? 0), 0), 10000);
  return Math.round((174 * point) / 10000);
};
const battleBloodItems = computed(() => {
  if (battleActiveMapId.value !== 2001) return [];
  if ((battleFieldSnapshot.value?.winPoints?.length ?? 0) < 2) return [];
  return [
    { side: "left", width: battleWinPointWidth(0) },
    { side: "right", width: battleWinPointWidth(1) }
  ];
});
const battleUnionId = computed(() => flashObjectNumber(userInfo.value, ["union_id", "unionId"], 0));
const battleFieldCanSendPreview = computed(() => !!battleSelectedCurrentTroop.value?.id && battleFieldSendTargetCid() > 0);
const battleFieldSendDisabledReason = computed(() => {
  if (!battleFieldSnapshot.value) return "读取战场部队后可打开只读派遣预览。";
  if (!battleSelectedCurrentTroop.value?.id) return "暂无己方战场部队，不能派遣。";
  if (battleFieldSendTargetCid() <= 0) return "暂无可派遣目标，不能派遣。";
  return "";
});
const battleFieldSendTitle = computed(() =>
  battleFieldCanSendPreview.value ? "接口未接入，打开只读派遣预览。" : battleFieldSendDisabledReason.value
);
const battleFieldReadonlyNote = computed(() => {
  const disabledReason = battleFieldSendDisabledReason.value;
  if (disabledReason && disabledReason !== "读取战场部队后可打开只读派遣预览。") {
    return `只读预览：${disabledReason} 派遣、巡逻、攻击写接口未接入，不会写入或派发请求。`;
  }
  return battleFieldSnapshot.value
    ? "只读预览：已读取旧服战场部队数据；派遣、巡逻、攻击写接口未接入，不会写入或派发请求。"
    : "只读预览：战场部队接口未接入，不会写入或派发请求。";
});
const battleTroopViewNote = computed(() =>
  battleFieldSnapshot.value
    ? "只读预览：已读取旧服己方战场部队，不会派发部队。"
    : "只读预览，不会派发部队。"
);
const battleTroopDetailNote = computed(() =>
  selectedBattleTroopDetail.value
    ? "只读预览：已读取旧服战场部队详情；召回、加速、撤退写接口未接入。"
    : battleFieldSnapshot.value
      ? "只读预览：已读取旧服战场部队列表；召回、加速、撤退写接口未接入。"
      : "只读预览：战场部队详情接口未接入，不会写入或派发请求。"
);
const battleSelectedDetail = computed(() => {
  const detail = selectedBattleTroopDetail.value;
  const row = selectedBattleFieldRow.value;
  const troop = selectedBattleCurrentTroop.value;
  return {
    user: detail?.name || row?.name || user.value?.name || "-",
    union: detail?.union || row?.union || userUnionName.value || "-",
    hero: detail?.hero || row?.hero || troop?.heroName || "未知",
    level: formatFlashInteger(detail?.level ?? row?.level ?? troop?.heroLevel),
    state: detail?.stateLabel || row?.stateLabel || "待命",
    target: detail?.targetName || (row ? battleFieldTargetLabel(row) : selectedCityLabel.value || city.value?.summary.name || "-"),
    leftTime: formatDuration(detail?.secondsLeft ?? 0),
    arrival: detail?.endTime ? formatBattleClock(detail.endTime) : "0:00",
    soldiers: detail?.soldiers?.length ? detail.soldiers : row?.soldiers?.length ? row.soldiers : troop?.soldiers ?? []
  };
});
const battleSelectedCurrentTroop = computed(() => selectedBattleCurrentTroop.value ?? battleCurrentTroops.value[0] ?? null);
function battleSoldierKey(soldier: TroopType, index = 0) {
  return Number((soldier as unknown as { sid?: number }).sid ?? soldier.tid ?? index);
}

const battleActionSoldiers = computed<TroopType[]>(() => {
  const currentTroop = battleSelectedCurrentTroop.value;
  if (currentTroop?.soldiers?.length) return currentTroop.soldiers;
  return battleCampaignSoldiers.value.map((soldier) => ({
    tid: soldier.sid,
    name: soldier.name,
    count: soldier.count,
    injured: 0
  }));
});
const battleTargetSoldiers = computed(() => selectedBattleFieldRow.value?.soldiers ?? []);
const battleActionTargetName = computed(() => {
  if (battleSendPreviewData.value?.target) return battleSendPreviewData.value.target;
  const row = selectedBattleFieldRow.value;
  return row ? battleFieldTargetLabel(row) : battleFieldName();
});
const battleActionHeroName = computed(() => battleSelectedCurrentTroop.value?.heroName || "未选择");
const battleActionHeroLevel = computed(() => formatFlashInteger(battleSelectedCurrentTroop.value?.heroLevel));
const battleActionPathTime = computed(() => formatDuration(battleSendPreviewData.value?.pathTime ?? 0));
const battleActionArrival = computed(() => battleSendPreviewData.value?.arrival ? formatBattleClock(battleSendPreviewData.value.arrival) : "0:00");
const battleAttackTargetName = computed(() => {
  if (battleAttackPreviewData.value?.targetName) return battleAttackPreviewData.value.targetName;
  const row = selectedBattleFieldRow.value;
  return row ? row.name || battleFieldTargetLabel(row) : battleFieldName();
});
const battleAttackMyHeroName = computed(() => battleAttackPreviewData.value?.troop.hero || battleSelectedCurrentTroop.value?.heroName || "未选择");
const battleAttackTargetHeroName = computed(() => battleAttackPreviewData.value?.target.hero || selectedBattleFieldRow.value?.hero || "未知");
const battleAttackPathTime = computed(() => formatDuration(battleAttackPreviewData.value?.pathTime ?? 0));
const battleAttackArrival = computed(() => battleAttackPreviewData.value?.arrival ? formatBattleClock(battleAttackPreviewData.value.arrival) : "0:00");
const battleTroopPreviewIsMyArmy = computed(() => {
  if (selectedBattleCurrentTroop.value || selectedBattleTroopDetail.value?.uid === (lastLogin.value?.uid ?? user.value?.uid ?? 0)) return true;
  const row = selectedBattleFieldRow.value;
  if (!row) return true;
  const currentUid = lastLogin.value?.uid ?? user.value?.uid ?? 0;
  if (currentUid > 0 && row.uid === currentUid) return true;
  const currentName = user.value?.name || lastLogin.value?.user?.name || city.value?.summary.owner || "";
  return !!currentName && row.name === currentName;
});
const battleActionReadonlyNote = computed(() =>
  battleSendPreviewData.value?.message ||
  (battleFieldSnapshot.value
    ? "只读预览：已读取旧服战场兵力；派遣接口未接入，不会写入或派发请求。"
    : "只读预览：派遣接口未接入，不会写入或派发请求。")
);
const battleAttackReadonlyNote = computed(() =>
  battleAttackPreviewData.value?.message ||
  (battleFieldSnapshot.value
    ? "只读预览：已读取旧服双方部队；攻击接口未接入，不会写入或派发请求。"
    : "只读预览：攻击接口未接入，不会写入或派发请求。")
);
const battlePatrolReadonlyNote = computed(() =>
  battlePatrolPreviewData.value?.message ||
  (battleFieldSnapshot.value
    ? "只读预览：已读取旧服侦察目标；巡逻接口未接入，不会扣除信鸽或写入战报。"
    : "只读预览：巡逻接口未接入，不会扣除信鸽或写入战报。")
);
const battlePatrolReportLines = computed(() => battlePatrolPreviewData.value?.reportLines ?? []);
const battleCampaignFoodCarry = computed(() =>
  battleCampaignMode.value === "battle" && battleCampaignPreviewData.value
    ? formatFlashInteger(battleCampaignPreviewData.value.foodUse)
    : String(battleCampaignSoldiers.value.reduce((total, soldier) => total + soldier.takecount, 0))
);
const battleCampaignSelectedSoldiers = computed(() =>
  battleCampaignSoldiers.value
    .filter((soldier) => soldier.takecount > 0)
    .map((soldier) => ({ sid: soldier.sid, count: soldier.takecount }))
);
const isWorldScoutCampaign = computed(() => battleCampaignMode.value === "world" && battleCampaignTask.value === 2);
const battleCampaignActionLabel = computed(() => isWorldScoutCampaign.value ? "侦察" : "出征");
const battleCampaignTitle = computed(() => battleCampaignMode.value === "battle" ? "战场出征" : isWorldScoutCampaign.value ? "地图侦察" : "地图出征");
const battleCampaignLeftTitle = computed(() => isWorldScoutCampaign.value ? "侦察部队" : "出征部队");
const battleCampaignTargetLabel = computed(() => battleCampaignMode.value === "battle" ? "出征目标" : isWorldScoutCampaign.value ? "侦察目标" : "目标选择");
const battleCampaignFieldLabel = computed(() => battleCampaignMode.value === "battle" ? "出征目标" : isWorldScoutCampaign.value ? "侦察目标" : "目标名称");
const battleCampaignPathLabel = computed(() => battleCampaignMode.value === "battle" ? "单程时间" : isWorldScoutCampaign.value ? "侦察耗时" : "行军耗时");
const battleCampaignFoodLabel = computed(() => battleCampaignMode.value === "battle" ? "战场耗粮" : "携带粮草");
const battleCampaignHeroReadonlyReason = computed(() =>
  battleCampaignMode.value === "world" && battleCampaignTask.value === 3 && battleCampaignHeroId.value === "0"
    ? "请选择将领后执行掠夺。"
    : ""
);
const battleCampaignStartDisabledReason = computed(() => {
  if (battleCampaignMode.value === "battle") return battleCampaignPreviewData.value?.reason || "战场出征接口未接入。";
  if (battleCampaignSelectedSoldiers.value.length === 0) return `请选择${battleCampaignActionLabel.value}士兵。`;
  if (Number.parseInt(battleCampaignTargetId.value, 10) <= 0) return `缺少旧服编号，不能${battleCampaignActionLabel.value}。`;
  if (battleCampaignHeroReadonlyReason.value) return battleCampaignHeroReadonlyReason.value;
  return "";
});
const battleCampaignStartNote = computed(() => battleCampaignStartReason.value || battleCampaignStartDisabledReason.value);
const battleFieldInfo = computed(() => battleFieldSnapshot.value?.info ?? null);
const battleInfoMeta = computed(() => [
  { label: "战场名称", value: battleFieldInfo.value?.name || battleFieldName() },
  { label: "当前城池", value: selectedCityLabel.value || city.value?.summary.name || "-" },
  { label: "难度", value: formatFlashInteger(battleFieldInfo.value?.level || 10) },
  { label: "战场兵力", value: `${formatFlashInteger(battleFieldInfo.value?.minPeople || 1)}~${formatFlashInteger(battleFieldInfo.value?.maxPeople || 5)}` },
  { label: "参加人数", value: battleFieldSnapshot.value ? formatFlashInteger(battleFieldInfo.value?.peopleTotal ?? battleInviteDisplayUsers.value.length) : "0" }
]);
const battleInfoContent = computed(() =>
  battleFieldInfo.value?.content?.trim() ||
  "只读预览：已读取旧服战场说明/部队/城池快照；派遣、巡逻、攻击、退出写接口未接入，不会写入或派发任何战场请求。"
);
const battleInfoImageName = computed(() => battleFieldInfo.value?.image || "battle_gdzz.jpg");
const battleInfoImageSrc = computed(() => asset(battleInfoImageName.value));
const battleNewsPageCount = computed(() => Math.max(1, Math.ceil(battleNewsTotal.value / battleNewsPageSize)));
const battleInfoNewsPageItems = computed(() => battleNewsItems.value);
const battleNewsReadonlyTitle = computed(() =>
  battleFieldSnapshot.value ? "战场讯息读取已接入；分页按旧服日志读取，写入接口未接入。" : "战场讯息接口未接入。"
);
const battleNewsReadonlyNote = computed(() =>
  battleFieldSnapshot.value ? "战场讯息为旧服只读数据；分页已接入，写入接口未接入，不会派发写请求。" : "战场讯息接口未接入，只读预览不会写入。"
);
const battleInviteCreator = ref(true);
const battleInviteName = ref("");
const battleInviteSnapshotRows = computed(() => {
  const memberMap = new Map<number, { id: number; name: string; camp: string; state: string; herocount: number; honour: number; cancel: boolean }>();
  for (const row of battleFieldRows.value) {
    const id = row.uid || row.id;
    const existing = memberMap.get(id);
    const camp = row.battleUnionId && row.battleUnionId === battleUnionId.value ? "我方" : row.union || "敌方";
    if (existing) {
      existing.herocount += row.hero ? 1 : 0;
      existing.honour += row.soldierCount;
      if (row.stateLabel && existing.state === "参战") existing.state = row.stateLabel;
      continue;
    }
    memberMap.set(id, {
      id,
      name: row.name || `玩家${id}`,
      camp,
      state: row.stateLabel || "参战",
      herocount: row.hero ? 1 : 0,
      honour: row.soldierCount,
      cancel: false
    });
  }
  for (const cityItem of battleFieldSnapshot.value?.cities ?? []) {
    if (!cityItem.hasUser) continue;
    const id = cityItem.uid || cityItem.cid;
    if (memberMap.has(id)) continue;
    memberMap.set(id, {
      id,
      name: cityItem.name || `城池${formatCityCode(cityItem.cid)}`,
      camp: cityItem.flagLabel || cityItem.flagChar || "-",
      state: "驻守",
      herocount: 0,
      honour: 0,
      cancel: false
    });
  }
  return [...memberMap.values()];
});
const battleInviteDisplayUsers = computed(() =>
  battleMembersSnapshot.value?.rows?.length
    ? battleMembersSnapshot.value.rows
    : battleInviteSnapshotRows.value
);
const battleInviteDisplayCount = computed(() =>
  battleMembersSnapshot.value
    ? `${formatFlashInteger(battleMembersSnapshot.value.inCount)}/10`
    : battleFieldSnapshot.value
    ? `${battleInviteDisplayUsers.value.length}/10`
    : "0/10"
);
const activityItems = ref([
  { content: "暂无活动公告", link: "", interval: 1800 }
]);

const flashSoldierSlots = [
  { sid: 1, name: "民夫", count: 0 },
  { sid: 2, name: "义兵", count: 0 },
  { sid: 3, name: "斥候", count: 0 },
  { sid: 4, name: "长枪兵", count: 0 },
  { sid: 5, name: "刀盾兵", count: 0 },
  { sid: 6, name: "弓箭兵", count: 0 },
  { sid: 7, name: "轻骑兵", count: 0 },
  { sid: 8, name: "铁骑兵", count: 0 },
  { sid: 9, name: "辎重车", count: 0 },
  { sid: 10, name: "床弩", count: 0 },
  { sid: 11, name: "冲车", count: 0 },
  { sid: 12, name: "投石车", count: 0 }
];
const flashDefenceSlots = [
  { did: 1, name: "陷阱", count: 0 },
  { did: 2, name: "拒马", count: 0 },
  { did: 3, name: "箭塔", count: 0 },
  { did: 4, name: "滚木", count: 0 },
  { did: 5, name: "擂石", count: 0 }
];

const provincePixelCache = new Map<number, ImageData>();
const currentUnix = ref(Math.floor(Date.now() / 1000));
const resourceSyncUnix = ref(currentUnix.value);
let timerID = 0;
let leftPanelTimerCounter15 = 15;
let leftPanelTimerCounter60 = 60;
const activityCurrentIndex = ref(0);
let activityScrollTimeLeft = 0;
let buildRequestSeq = 0;

const guideRect = computed(() => parseGuideRect(currentGuide.value?.disdetails || currentGuide.value?.showpos));
const guideShowRect = computed(() => parseGuideRect(currentGuide.value?.showpos));
const guideText = computed(() => currentGuide.value?.content.replace(/\\n/g, "\n") || "");
const guideTipMetrics = computed(() => {
  const contentHeight = estimateGuideTextHeight(guideText.value);
  return {
    height: contentHeight + 50,
    skipTop: contentHeight + 5
  };
});
const stageStyle = computed(() => ({
  transform: `scale(${stageScale.value})`
}));
const guideTipPosition = computed(() => {
  return guidePlacement.value.tip;
});
const guideArrow = computed(() => {
  return guidePlacement.value.arrow;
});
const effectiveGuideVisible = computed(() => {
  return guideVisible.value && guideMatchesCurrentScene(currentGuide.value);
});
const guidePlacement = computed(() => {
  const rect = guideShowRect.value ?? { x: 300, y: 200, w: 0, h: 0 };
  const hiddenArrow = !guideShowRect.value;
  const preferred =
    rect.x < STAGE_WIDTH / 10 ? 4 :
    rect.y < STAGE_HEIGHT / 10 ? 1 :
    rect.x > 9 * STAGE_WIDTH / 10 ? 2 :
    rect.y > 9 * STAGE_HEIGHT / 10 ? 3 :
    rect.x > 2 * STAGE_WIDTH / 3 ? 2 :
    rect.y > 2 * STAGE_HEIGHT / 3 ? 3 :
    rect.x < STAGE_WIDTH / 3 ? 4 :
    1;
  let placement = tryGuidePlacement(preferred, rect, hiddenArrow);
  for (const direction of [preferred, 1, 2, 3, 4, preferred]) {
    placement = tryGuidePlacement(direction, rect, hiddenArrow);
    if (!placement.corrected) break;
  }
  return placement;
});
const worldMapCities = computed<WorldMap["cities"]>(() => {
  const byID = new Map<number, WorldMap["cities"][number]>();
  for (const cityItem of worldMapData.value?.cities ?? []) {
    byID.set(cityItem.cid, cityItem);
  }
  if (city.value) {
    const summary = city.value.summary;
    byID.set(summary.cid, {
      cid: summary.cid,
      name: summary.name,
      owner: summary.owner,
      ownerId: user.value?.uid ?? 0,
      x: summary.x,
      y: summary.y,
      level: 10,
      flagChar: ""
    });
  }
  return Array.from(byID.values());
});
const worldTerrainTiles = computed<WorldTerrainTile[]>(() => {
  const center = currentWorldMapCenter();
  const cities = new Map<string, { name: string; level: number }>();
  for (const cityItem of worldMapCities.value) {
    const point = flashCityPoint(cityItem);
    cities.set(`${point.x},${point.y}`, { name: cityItem.name, level: cityItem.level });
  }

  const tiles: WorldTerrainTile[] = [];
  for (let my = center.y - WORLD_GRID_LOAD_RANGE; my <= center.y + WORLD_GRID_LOAD_RANGE; my++) {
    for (let mx = center.x - WORLD_GRID_LOAD_RANGE; mx <= center.x + WORLD_GRID_LOAD_RANGE; mx++) {
      const dx = mx - center.x;
      const dy = my - center.y;
      if (dx + dy < -12 || dx + dy > 12 || Math.abs(dx - dy) > 9) continue;
      const cityItem = cities.get(`${mx},${my}`);
      tiles.push({
        key: `${mx}-${my}`,
        x: WORLD_GRID_CENTER_X + dx * WORLD_GRID_OFFSET_X - dy * WORLD_GRID_OFFSET_X,
        y: WORLD_GRID_CENTER_Y + dx * WORLD_GRID_OFFSET_Y + dy * WORLD_GRID_OFFSET_Y,
        mx,
        my,
        image: worldTerrainImage(mx, my, cityItem?.level),
        title: cityItem ? `${cityItem.name}[${mx},${my}]` : `${mx},${my}`
      });
    }
  }
  return tiles;
});

const loginForm = ref({
  passport: localStorage.getItem("rxsg_passport") || "",
  password: ""
});

const roleForm = ref({
  userName: "",
  cityName: "",
  flagChar: "H",
  sex: 0,
  face: 1
});

const debugParams = new URLSearchParams(window.location.search);
const debugScreen = debugParams.get("debugScreen");
const debugBattlefieldId = Number.parseInt(debugParams.get("debugBattlefieldId") || "", 10);
if (Number.isFinite(debugBattlefieldId) && debugBattlefieldId > 0) {
  battleMapId.value = debugBattlefieldId;
}
const occupiedByPosition = computed(() => {
  const map = new Map<number, Building>();
  for (const item of city.value?.buildings ?? []) {
    if (item.bid === 6) {
      map.set(cityHallPlot.position, item);
    } else if (item.bid === 20) {
      map.set(cityWallPlot.position, item);
    } else {
      map.set(item.position, item);
    }
  }
  return map;
});
const governmentLevel = computed(() => {
  const cityHall = city.value?.buildings.find((building) => building.bid === 6);
  return cityHall?.level ?? 1;
});
const visibleOuterPlots = computed(() => outerPlotsForGovernmentLevel(governmentLevel.value));

const resources = computed(() => city.value?.summary.resources);
const production = computed(() => city.value?.production);
const idlePeople = computed(() => (resources.value?.people ?? 0) - (production.value?.peopleWorking ?? 0) - (resources.value?.peopleBuilding ?? 0));
const leftFoodArmyUse = computed(() => Math.floor(production.value?.foodArmyUse ?? 0));
const leftFoodAdd = computed(() => Math.floor((production.value?.foodAdd ?? 0) - leftFoodArmyUse.value));
const leftGoldAdd = computed(() => Math.floor(production.value?.goldAdd ?? 0));
const leftGoldTaxAdd = computed(() => leftGoldAdd.value + Math.floor(production.value?.heroFee ?? 0));
const leftIdlePeople = computed(() => idlePeople.value);
const leftFoodValue = computed(() => flashTickedResourceValue(resources.value?.food, resources.value?.foodMax, leftFoodAdd.value));
const leftWoodValue = computed(() => flashTickedResourceValue(resources.value?.wood, resources.value?.woodMax, production.value?.woodAdd));
const leftRockValue = computed(() => flashTickedResourceValue(resources.value?.rock, resources.value?.rockMax, production.value?.rockAdd));
const leftIronValue = computed(() => flashTickedResourceValue(resources.value?.iron, resources.value?.ironMax, production.value?.ironAdd));
const visibleActivityItems = computed(() => {
  const items = activityItems.value;
  const count = Math.min(6, items.length);
  if (count === 0) return [];
  return Array.from({ length: count }, (_, offset) => items[(activityCurrentIndex.value + offset) % items.length]);
});
const heroRosterItems = computed(() => heroRoster.value?.items ?? heroRoster.value?.heroes ?? []);
const heroRecruitCapacity = computed(() => heroRoster.value?.recruitCapacity ?? heroRoster.value?.recruitFreeCount ?? 0);
const leftHeroItems = computed(() => heroRosterItems.value);
const leftSoldierItems = computed(() => {
  const byID = new Map([...barracksTroopItems.value, ...cityTroopItems.value].map((item) => [item.tid, item]));
  return flashSoldierSlots.map((slot) => {
    const troop = byID.get(slot.sid);
    return {
      sid: slot.sid,
      name: slot.name,
      count: troop?.count ?? slot.count
    };
  });
});
const leftDefenceItems = computed(() => {
  const raw = city.value?.defenceList ?? [];
  const byID = new Map(raw.map((item) => [Number(item.did), item]));
  return flashDefenceSlots.map((slot) => {
    const item = byID.get(slot.did);
    return {
      did: slot.did,
      name: slot.name,
      count: item?.count ?? slot.count
    };
  });
});
const userInfo = computed(() => (user.value ?? {}) as SessionUser & Record<string, unknown>);
const userPrestige = computed(() => formatFlashInteger(typeof userInfo.value.prestige === "number" ? userInfo.value.prestige : 0));
const userRank = computed(() => formatFlashInteger(typeof userInfo.value.rank === "number" ? userInfo.value.rank : 0));
const userOffice = computed(() => String(userInfo.value.officepos ?? userInfo.value.office ?? "平民"));
const userNobility = computed(() => String(userInfo.value.nobility ?? userInfo.value.nobilityname ?? "平民"));
const userUnionName = computed(() => {
  const unionName = userInfo.value.unionname ?? userInfo.value.union;
  return typeof unionName === "string" && unionName.trim() ? unionName : "无联盟";
});
const userUnionPosition = computed(() => {
  const unionPosition = userInfo.value.union_pos ?? userInfo.value.unionpos;
  return typeof unionPosition === "string" && unionPosition.trim() ? unionPosition : "";
});
const lordName = computed(() => user.value?.name || city.value?.summary.owner || "");
const lordInfoRows = computed(() => [
  { label: "君主", value: lordName.value },
  { label: "声望", value: userPrestige.value },
  { label: "排名", value: userRank.value },
  { label: "官职", value: userOffice.value },
  { label: "爵位", value: userNobility.value },
  { label: "联盟", value: userUnionName.value },
  { label: "职位", value: userUnionPosition.value || "无" },
  { label: "城池", value: city.value?.summary.name ?? "" }
]);
const equipmentSlots = [
  "武器",
  "坐骑",
  "头盔",
  "铠甲",
  "披风",
  "戒指"
];
const equipmentReadonlyNote = "装备写接口未接入，当前仅保留只读预览。";
const treasureReadonlyNote = "宝物背包接口未接入，当前仅显示旧服只读预览。";
const treasureRows = computed(() => [
  { name: "加速道具", text: "可从资源栏 + 号或建筑面板进入对应使用窗口。" },
  { name: "资源宝物", text: `民心 ${city.value?.morale ?? 0}，税率 ${city.value?.tax ?? 0}%` },
  { name: "君主宝物", text: treasureReadonlyNote }
]);
const treasureWallet = computed(() => {
  const value = userInfo.value.gold ?? userInfo.value.money;
  return typeof value === "number" ? value : 0;
});
const selectedProvinceName = computed(
  () => provinceAssets.find((province) => province.id === selectedProvince.value)?.name ?? ""
);
const selectedCityLabel = computed(() => {
  const selected = cityList.value.find((item) => item.cid === selectedCityId.value);
  const item = selected ?? city.value?.summary;
  return item ? `${item.name}${formatCityCode(item.cid)}` : "";
});
const activeProvince = computed(() => hoveredProvince.value ?? selectedProvince.value);
const faceImage = computed(() => {
  const sexValue =
    typeof userInfo.value.usersex === "number"
      ? userInfo.value.usersex
      : typeof userInfo.value.sex === "number"
        ? userInfo.value.sex
        : roleForm.value.sex;
  const rawFaceValue =
    typeof userInfo.value.userface === "number"
      ? userInfo.value.userface
      : typeof userInfo.value.face === "number"
        ? userInfo.value.face
        : roleForm.value.face - 1;
  const faceValue = rawFaceValue + 1;
  const sexName = sexValue === 1 ? "male" : "female";
  const face = Math.min(Math.max(faceValue, 1), 10);
  return asset(`player/player_${sexName}_${face}.jpg`);
});
const buildingRemaining = computed(() => {
  if (!buildingPanel.value || buildingPanel.value.stateEndTime <= 0) return 0;
  return Math.max(0, buildingPanel.value.stateEndTime - currentUnix.value);
});
const buildingStateText = computed(() => {
  if (!buildingPanel.value) return "";
  if (buildingPanel.value.state === 1) return `正在升级  剩余:${formatDuration(buildingRemaining.value)}`;
  if (buildingPanel.value.state === 2) return `正在拆除  剩余:${formatDuration(buildingRemaining.value)}`;
  return "";
});
const buildingDescriptionText = computed(() => {
  return buildingInfoPanel.value?.current.levelDescription || "";
});
const buildingNeedText = computed(() => {
  const next = buildingInfoPanel.value?.next;
  if (!next || buildingPanel.value?.state !== 0) return "";
  const parts = [
    next.peopleNeed > 0 ? `人口 ${formatNumber(next.peopleNeed)}` : "",
    next.foodNeed > 0 ? `粮食 ${formatNumber(next.foodNeed)}` : "",
    next.woodNeed > 0 ? `木材 ${formatNumber(next.woodNeed)}` : "",
    next.rockNeed > 0 ? `石料 ${formatNumber(next.rockNeed)}` : "",
    next.ironNeed > 0 ? `铁锭 ${formatNumber(next.ironNeed)}` : "",
    next.goldNeed > 0 ? `黄金 ${formatNumber(next.goldNeed)}` : ""
  ].filter(Boolean);
  const conditionText = next.conditions
    .map((item) => `${item.type} ${formatNumber(item.currentOwn)}/${formatNumber(item.upgradeNeed)}`)
    .join(" ");
  const needText = parts.length > 0 ? `升级到: ${buildingPanel.value?.level ? buildingPanel.value.level + 1 : 1} ${parts.join(" ")}` : "";
  return conditionText ? `${needText} ${conditionText}` : needText;
});
const upgradingBuildings = computed(() =>
  (city.value?.buildings ?? [])
    .filter((building) => building.state > 0)
    .map((building) => ({
      ...building,
      task: building.state === 2 ? "拆除" : building.level === 0 ? "建造" : "升级",
      nextLevel: building.state === 2 ? Math.max(0, building.level - 1) : building.level + 1,
      timeLeft: formatDuration(Math.max(0, building.stateEndTime - currentUnix.value)),
      endTime: building.stateEndTime > 0 ? new Date(building.stateEndTime * 1000).toLocaleString("zh-CN") : ""
    }))
);
const topBuildingQueueItems = computed(() => upgradingBuildings.value.slice(0, 5));
const buildingListQueueItems = computed(() => upgradingBuildings.value);
const buildingListOuterItems = computed(() =>
  (city.value?.buildings ?? [])
    .filter((building) => building.bid !== 6 && (building.bid === 20 || !innerPlots.some((plot) => plot.position === building.position)))
    .sort((a, b) => a.position - b.position)
);
const buildingListInnerItems = computed(() =>
  (city.value?.buildings ?? [])
    .filter((building) => building.bid === 6 || (building.bid !== 20 && innerPlots.some((plot) => plot.position === building.position)))
    .sort((a, b) => a.position - b.position)
);
const taskCategories = computed(() => taskSnapshot.value?.categories ?? []);
const emptyBattleTaskCategory = computed<TaskCategory>(() => ({
  type: BATTLE_TASK_CATEGORY_TYPE,
  label: "战场任务",
  groupCount: 0,
  taskCount: 0,
  completed: 0,
  groups: []
}));

function makeLocalTaskSnapshot(categories = [emptyBattleTaskCategory.value]): TaskSnapshot {
  const taskCount = categories.reduce((total, category) => total + category.taskCount, 0);
  const completedTasks = categories.reduce((total, category) => total + category.completed, 0);
  return {
    categories,
    summary: {
      groupCount: categories.reduce((total, category) => total + category.groupCount, 0),
      taskCount,
      completedTasks,
      pendingTasks: Math.max(0, taskCount - completedTasks)
    }
  };
}

const selectedTaskCategory = computed(() => {
  const categories = taskCategories.value;
  if (categories.length === 0) return null;
  return categories.find((category) => category.type === selectedTaskCategoryType.value) ?? categories[0] ?? null;
});
const selectedTaskGroups = computed(() => selectedTaskCategory.value?.groups ?? []);
const selectedTaskGroup = computed(() => {
  const groups = selectedTaskGroups.value;
  if (groups.length === 0) return null;
  return groups.find((group) => group.id === selectedTaskGroupId.value) ?? groups[0] ?? null;
});
const selectedTask = computed(() => {
  const tasks = selectedTaskGroup.value?.tasks ?? [];
  if (tasks.length === 0) return null;
  return tasks.find((task) => task.id === selectedTaskId.value) ?? tasks[0] ?? null;
});
const isBattleTaskPanel = computed(() => selectedTaskCategory.value?.type === BATTLE_TASK_CATEGORY_TYPE);
const taskClaimDisabled = computed(() => !selectedTask.value?.completed || isBattleTaskPanel.value);
const taskClaimTitle = computed(() => {
  if (isBattleTaskPanel.value) return "战场任务奖励写接口未接入，当前仅可查看任务进度。";
  return selectedTask.value?.completed ? "领取奖励" : (selectedTask.value?.todo || "任务尚未完成。");
});
const taskClaimReadonlyNote = computed(() =>
  isBattleTaskPanel.value
    ? "战场任务奖励写接口未接入，当前仅可查看旧服战场任务进度。"
    : "任务未完成时奖励领取按钮禁用，仅可查看任务进度。"
);
const shopGroups = computed(() => shopSnapshot.value?.groups ?? []);
const selectedShopGroup = computed(() => {
  const groups = shopGroups.value;
  if (groups.length === 0) return null;
  return groups.find((group) => group.id === selectedShopGroupId.value) ?? groups[0] ?? null;
});
const selectedShopItems = computed(() => selectedShopGroup.value?.items ?? []);
const friendRows = computed(() =>
  (relationSnapshot.value?.items ?? []).filter((item) =>
    friendRelationTab.value === "friends" ? item.relationType === 0 : item.relationType === 1
  )
);
const friendReadonlyNote = "关系写接口禁用，当前仅可查看旧服关系名单。";
const statSummaryRows = computed(() => [
  { label: "君主", value: lordName.value || "-" },
  { label: "城池", value: selectedCityLabel.value || city.value?.summary.name || "-" },
  { label: "声望", value: userPrestige.value },
  { label: "排名", value: userRank.value },
  { label: "人口", value: formatFlashInteger(resources.value?.people) },
  { label: "民心", value: formatFlashInteger(city.value?.morale) },
  { label: "税率", value: `${city.value?.tax ?? 0}%` },
  { label: "将领", value: formatFlashInteger(city.value?.heroCount) }
]);
const statResourceRows = computed(() => [
  { label: "粮食", current: leftFoodValue.value, max: resources.value?.foodMax, add: leftFoodAdd.value },
  { label: "木材", current: leftWoodValue.value, max: resources.value?.woodMax, add: production.value?.woodAdd },
  { label: "石料", current: leftRockValue.value, max: resources.value?.rockMax, add: production.value?.rockAdd },
  { label: "铁锭", current: leftIronValue.value, max: resources.value?.ironMax, add: production.value?.ironAdd },
  { label: "黄金", current: resources.value?.gold, max: resources.value?.goldMax, add: leftGoldAdd.value }
]);
const portalBoardRows = [
  { title: "官方公告", text: "旧版公告与论坛入口", count: "只读" },
  { title: "玩家交流", text: "打开旧服论坛频道", count: "外链" },
  { title: "活动讨论", text: "活动、战报、攻略交流", count: "外链" }
];
const forumTopicRows = [
  { type: "公告", title: "热血三国论坛入口", date: "旧服" },
  { type: "攻略", title: "新手、城池、战场讨论区", date: "只读" },
  { type: "反馈", title: "迁移版暂不内嵌论坛发帖", date: "外链" }
];
const websiteLinkRows = [
  { title: "游戏官网", text: "旧版游戏官网入口", url: "https://www.uuyx.com/" },
  { title: "活动中心", text: "充值、公告、活动页面", url: "https://www.uuyx.com/activity/" },
  { title: "客服中心", text: "账号与旧服资料说明", url: "https://www.uuyx.com/service/" }
];
const helpRuleRows = [
  { title: "城池建设", text: "点击空地建造，点击建筑查看升级、拆除与加速入口。" },
  { title: "资源生产", text: "左侧资源栏显示当前值、容量与产量；官府内可查看生产比例。" },
  { title: "世界地图", text: "顶部地图按钮进入世界地图，选择野地或城池后显示只读操作面板。" }
];
const helpTopicRows = [
  { title: "新手指南", text: "创建角色、建造民房、领取任务奖励", url: "http://action.uuyx.com/GameHelp/Help1/index.html" },
  { title: "战斗说明", text: "出征、战场、侦察与战报", url: "http://action.uuyx.com/GameHelp/Help1/index.html" },
  { title: "道具说明", text: "加速道具、资源宝物、充值兑换", url: "http://action.uuyx.com/GameHelp/Help1/index.html" }
];
const chargeSummaryRows = computed(() => {
  const summary = chargeSnapshot.value?.summary;
  return [
    { label: "账号", value: summary?.userName || lordName.value || "-" },
    { label: "当前城池", value: summary?.focusCity || city.value?.summary.name || "-" },
    { label: "元宝", value: formatFlashInteger(summary?.yuanbao) },
    { label: "礼金", value: formatFlashInteger(summary?.gift) },
    { label: "今日充值", value: formatFlashInteger(summary?.todayPaid) },
    { label: "累计充值", value: formatFlashInteger(summary?.totalPaid) }
  ];
});
const chargeBucketRows = computed(() => chargeSnapshot.value?.buckets ?? []);
const chargeEventRows = computed(() => chargeSnapshot.value?.events ?? []);
const useGoodsDialogStyle = computed(() => {
  const count = useGoodsPanel.value.goodsList.length;
  let width = 467;
  if (count === 1) {
    width = 160;
  } else if (count > 1 && count < 7) {
    width = count * 80 + (count - 1) * 13 + 40;
  } else if (count >= 7) {
    width = 682;
  }
  return {
    width: `${width}px`,
    left: `${Math.round((1000 - width) / 2)}px`
  };
});

const WORLD_MAP_CENTER_X = 314;
const WORLD_MAP_CENTER_Y = 250;
const WORLD_MAP_GRID_OFFSET_X = 54;
const WORLD_MAP_GRID_OFFSET_Y = 27;

function formatNumber(value: number | undefined) {
  if (value == null) return "0";
  return value.toLocaleString("zh-CN");
}

function cityMorale(item: CityCard) {
  return (item as unknown as { morale?: number }).morale ?? "-";
}

function formatFlashInteger(value: number | undefined) {
  if (value == null || !Number.isFinite(value)) return "0";
  return String(Math.floor(value));
}

function formatFlashRoundedInteger(value: number | undefined) {
  if (value == null || !Number.isFinite(value)) return "0";
  return String(Math.round(value));
}

function formatBattleNewsTime(unix: number | undefined) {
  if (!unix || !Number.isFinite(unix)) return "";
  const date = new Date(unix * 1000);
  const hour = String(date.getHours()).padStart(2, "0");
  const minute = String(date.getMinutes()).padStart(2, "0");
  return `${hour}:${minute}`;
}

function formatBattleClock(unix: number | undefined) {
  if (!unix || !Number.isFinite(unix)) return "0:00";
  const date = new Date(unix * 1000);
  const hour = String(date.getHours()).padStart(2, "0");
  const minute = String(date.getMinutes()).padStart(2, "0");
  return `${hour}:${minute}`;
}

function flashTickedResourceValue(current: number | undefined, max: number | undefined, add: number | undefined) {
  if (current == null || !Number.isFinite(current)) return 0;
  const capacity = max ?? Number.POSITIVE_INFINITY;
  const rate = add ?? 0;
  const elapsed = Math.max(0, currentUnix.value - resourceSyncUnix.value);
  if (current >= capacity && rate >= 0) return current;
  return Math.max(0, current + (rate * elapsed) / 3600);
}

function setCityDetail(nextCity: CityDetail) {
  city.value = nextCity;
  resourceSyncUnix.value = currentUnix.value;
}

function leftInfoTipStyle() {
  return {
    left: `${leftInfoTip.value.x}px`,
    top: `${leftInfoTip.value.y}px`,
    height: `${42 + leftInfoTip.value.rows.length * 16}px`
  };
}

function showLeftInfoTip(title: string, rows: LeftInfoTipInputRow[]) {
  hideFlashToolTip();
  leftInfoTip.value = {
    visible: true,
    x: 45,
    y: 39,
    title,
    rows: rows.map((row) => {
      if ("separator" in row) {
        return { label: LEFT_INFO_TIP_SEPARATOR, value: "" };
      }
      return {
        label: row.label,
        value: String(row.value),
        color: row.color ?? (row.danger ? "#ff0000" : undefined)
      };
    })
  };
}

function hideLeftInfoTip() {
  leftInfoTip.value = { ...leftInfoTip.value, visible: false };
}

function showFlashToolTip(event: MouseEvent, text: string) {
  flashToolTip.value = {
    visible: true,
    x: event.clientX + 11,
    y: event.clientY + 18,
    text
  };
}

function moveFlashToolTip(event: MouseEvent) {
  if (!flashToolTip.value.visible) return;
  flashToolTip.value = {
    ...flashToolTip.value,
    x: event.clientX + 11,
    y: event.clientY + 18
  };
}

function hideFlashToolTip() {
  flashToolTip.value = { ...flashToolTip.value, visible: false };
}

function showMoraleTip() {
  const morale = city.value?.morale ?? 0;
  const moraleStable = city.value?.moraleStable ?? morale;
  let trend = "稳定";
  let color = "#ffffff";
  if (moraleStable > morale) {
    trend = "上升";
    color = "#00ff00";
  } else if (moraleStable < morale) {
    trend = "下降";
    color = "#ff0000";
  }
  showLeftInfoTip("民心", [
    { label: "当前民心", value: morale },
    { label: "当前民怨", value: city.value?.complaint ?? 0 },
    { separator: true },
    { label: "民心变化", value: trend, color }
  ]);
}

function showFoodTip() {
  const food = leftFoodValue.value;
  const foodMax = resources.value?.foodMax ?? 0;
  const foodArmyUseText = `-${formatFlashInteger(leftFoodArmyUse.value)}`;
  showLeftInfoTip("粮食", [
    { label: "当前数量", value: formatFlashInteger(food), danger: food > foodMax || food <= 0 },
    { label: "容量上限", value: formatFlashInteger(foodMax) },
    { label: "粮食产量", value: formatFlashInteger(production.value?.foodAdd) },
    { label: "军队耗粮", value: foodArmyUseText, danger: leftFoodArmyUse.value > 0 },
    { label: "实际产量", value: formatFlashInteger(leftFoodAdd.value), danger: leftFoodAdd.value <= 0 }
  ]);
}

function showResourceTip(kind: "wood" | "rock" | "iron") {
  const names = { wood: "木材", rock: "石料", iron: "铁锭" };
  const current = kind === "wood" ? leftWoodValue.value : kind === "rock" ? leftRockValue.value : leftIronValue.value;
  const max = resources.value?.[`${kind}Max` as "woodMax" | "rockMax" | "ironMax"] ?? 0;
  const add = production.value?.[`${kind}Add` as "woodAdd" | "rockAdd" | "ironAdd"] ?? 0;
  showLeftInfoTip(names[kind], [
    { label: "当前数量", value: formatFlashInteger(current), danger: current > max || current <= 0 },
    { label: "容量上限", value: formatFlashInteger(max) },
    { label: "实际产量", value: formatFlashInteger(add) }
  ]);
}

function showPeopleTip() {
  const people = resources.value?.people ?? 0;
  const peopleMax = resources.value?.peopleMax ?? 0;
  const peopleStable = resources.value?.peopleStable ?? people;
  const peopleBuilding = resources.value?.peopleBuilding ?? 0;
  const free = leftIdlePeople.value;
  let trend = "稳定";
  let color = "#ffffff";
  if (peopleStable > people && (city.value?.tax ?? 0) < 100) {
    trend = "上升";
    color = "#00ff00";
  } else if (peopleStable < people && (city.value?.tax ?? 0) > 0) {
    trend = "下降";
    color = "#ff0000";
  }
  showLeftInfoTip("人口", [
    { label: "当前人口", value: formatFlashInteger(people) },
    { label: "人口上限", value: formatFlashInteger(peopleMax) },
    { separator: true },
    { label: "劳动人口", value: formatFlashInteger(production.value?.peopleWorking) },
    { label: "建筑人口", value: formatFlashInteger(peopleBuilding) },
    { label: "空闲人口", value: formatFlashInteger(free), danger: free < 0 },
    { separator: true },
    { label: "人口变化", value: trend, color }
  ]);
}

function showGoldTip() {
  const heroFee = Math.floor(production.value?.heroFee ?? 0);
  const heroFeeText = heroFee > 0 ? `-${formatFlashInteger(heroFee)}` : "0";
  showLeftInfoTip("黄金", [
    { label: "当前数量", value: formatFlashInteger(resources.value?.gold) },
    { label: "容量上限", value: formatFlashInteger(resources.value?.goldMax) },
    { label: "征税收入", value: formatFlashInteger(leftGoldTaxAdd.value) },
    { label: "将领俸禄", value: heroFeeText, danger: heroFee > 0 },
    { separator: true },
    { label: "实际收入", value: formatFlashInteger(leftGoldAdd.value), danger: leftGoldAdd.value <= 0 }
  ]);
}

function formatCityCode(value: number | undefined) {
  if (value == null) return "";
  const cid = Math.max(0, Math.floor(value));
  return `(${cid % 1000},${Math.floor(cid / 1000)})`;
}

function flashObjectNumber(item: unknown, keys: string[], fallback: number) {
  const record = (item ?? {}) as Record<string, unknown>;
  for (const key of keys) {
    const value = record[key];
    if (typeof value === "number" && Number.isFinite(value)) return value;
    if (typeof value === "string") {
      const parsed = Number.parseInt(value, 10);
      if (Number.isFinite(parsed)) return parsed;
    }
  }
  return fallback;
}

function flashObjectText(item: unknown, keys: string[], fallback: string) {
  const record = (item ?? {}) as Record<string, unknown>;
  for (const key of keys) {
    const value = record[key];
    if (typeof value === "string" && value.trim()) return value;
    if (typeof value === "number" && Number.isFinite(value)) return String(value);
  }
  return fallback;
}

function getCityChiefName(item: CityCard) {
  return (item as unknown as { chiefname?: string }).chiefname ?? item.owner;
}

function formatDuration(seconds: number) {
  const safe = Math.max(0, Math.floor(seconds));
  const hours = Math.floor(safe / 3600);
  const minutes = Math.floor((safe % 3600) / 60);
  const secs = safe % 60;
  if (hours > 0) return `${hours}:${String(minutes).padStart(2, "0")}:${String(secs).padStart(2, "0")}`;
  return `${minutes}:${String(secs).padStart(2, "0")}`;
}

function status(message = "") {
  error.value = message;
}

async function withLoading(task: () => Promise<void>) {
  loading.value = true;
  status();
  try {
    await task();
  } catch (err) {
    status(err instanceof Error ? err.message : String(err));
  } finally {
    loading.value = false;
  }
}

async function boot() {
  await withLoading(async () => {
    if (debugScreen === "login") {
      screen.value = "login";
      void loadLoginSnapshot();
      return;
    }
    if (debugScreen === "create-role") {
      screen.value = "create-role";
      lastLogin.value = {
        logged: true,
        queued: false,
        uid: Number(debugParams.get("uid") || 0),
        sid: Number(debugParams.get("sid") || 0)
      };
      return;
    }
    if (debugScreen === "city") {
      await debugEnterCity();
      return;
    }
    const me = await currentUser();
    user.value = me.user;
    if (!me.user) {
      screen.value = "login";
      void loadLoginSnapshot();
      return;
    }
    if (me.user.cityCount <= 0) {
      screen.value = "create-role";
      return;
    }
    await enterCity();
  });
}

async function loadLoginSnapshot() {
  loginSnapshotMessage.value = "读取中...";
  try {
    const [commanders, dashboard] = await Promise.all([
      commanderOptions(4),
      dashboardOverview()
    ]);
    loginCommanders.value = commanders.items ?? [];
    loginDashboard.value = dashboard;
    loginSnapshotMessage.value = "";
  } catch (error) {
    loginCommanders.value = [];
    loginDashboard.value = null;
    loginSnapshotMessage.value = error instanceof Error && error.message ? error.message : "登录快照读取失败";
  }
}

async function debugEnterCity() {
  const passport = debugParams.get("debugPassport") || "codex_city_debug";
  const result = await legacyLogin(passport, "");
  lastLogin.value = result;
  if (!result.logged || !result.uid || !result.sid) throw new Error(String(result.raw?.[1] ?? "调试登录失败"));
  user.value = result.user ?? null;
  if (!result.user || result.user.cityCount <= 0) {
    const created = await createRole({
      uid: result.uid,
      sid: result.sid,
      userName: "",
      cityName: "",
      province: selectedProvince.value,
      flagChar: "H",
      sex: 1,
      face: 1
    });
    user.value = created.user ?? user.value;
    await enterCity(created.cid);
    guideVisible.value = debugParams.get("guide") !== "0";
    return;
  }
  await enterCity(result.user.defaultCid, [result.user.defaultCid]);
  guideVisible.value = debugParams.get("guide") === "1";
}

async function submitLogin() {
  await withLoading(async () => {
    localStorage.setItem("rxsg_passport", loginForm.value.passport);
    const result = await legacyLogin(loginForm.value.passport.trim(), loginForm.value.password);
    lastLogin.value = result;
    if (!result.logged || !result.uid || !result.sid) throw new Error(String(result.raw?.[1] ?? "登录未完成"));
    user.value = result.user ?? null;
    if (!result.user || result.user.cityCount <= 0) {
      screen.value = "create-role";
      return;
    }
    await enterCity();
  });
}

async function submitRole() {
  await withLoading(async () => {
    const uid = lastLogin.value?.uid ?? user.value?.uid;
    const sid = lastLogin.value?.sid;
    if (!uid || !sid) throw new Error("缺少登录会话，请重新登录");
    const result = await createRole({
      uid,
      sid,
      userName: roleForm.value.userName.trim(),
      cityName: roleForm.value.cityName.trim(),
      province: selectedProvince.value,
      flagChar: roleForm.value.userName.trim().slice(0, 1),
      sex: roleForm.value.sex,
      face: roleForm.value.face - 1
    });
    if (!result.cid) throw new Error(String(result.raw?.[1] ?? "创建角色失败"));
    user.value = result.user ?? user.value;
    await enterCity(result.cid);
    guideVisible.value = true;
  });
}

async function enterCity(cid?: number, fallbackCityIds: number[] = []) {
  let cities: CityListResult = { items: [] };
  try {
    cities = await myCities();
  } catch (error) {
    if (!fallbackCityIds.length) throw error;
  }
  cityList.value = cities.items;
  const target = cid ?? user.value?.defaultCid ?? cities.items[0]?.cid ?? fallbackCityIds[0];
  if (!target) {
    screen.value = "create-role";
    return;
  }
  setCityDetail(await cityDetail(target));
  heroRoster.value = null;
  troopsData.value = null;
  await loadActivityList();
  await loadGuideGroup();
  screen.value = "city";
}

function selectCityLabel(cityItem: CityCard) {
  return `${cityItem.name}${cityItem.cid === selectedCityId.value ? "" : ""}`;
}

async function changeCity(event: Event) {
  const value = Number((event.target as HTMLSelectElement).value);
  if (!Number.isFinite(value) || value <= 0) return;
  await enterCity(value);
}

function toggleCityDropdown() {
  cityDropdownOpen.value = !cityDropdownOpen.value;
}

async function selectCityFromDropdown(cid: number) {
  cityDropdownOpen.value = false;
  if (!Number.isFinite(cid) || cid <= 0 || cid === selectedCityId.value) return;
  await enterCity(cid);
}

function closeCityDropdown(event: MouseEvent) {
  const target = event.target as HTMLElement | null;
  if (target?.closest(".flash-city-combo")) return;
  cityDropdownOpen.value = false;
}

async function loadGuideGroup() {
  const payload = await legacyGuides(1);
  guideItems.value = payload.items;
  currentGuide.value =
    payload.items.find((item) => item.gid === 6) ??
    payload.items.find((item) => item.distype === 1 && item.showpos) ??
    payload.items[0] ??
    null;
}

function normalizeActivityItems(payload: ActivityList) {
  return payload.items.map((item: ActivityItem) => ({
    content: item.content,
    link: item.link,
    interval: flashActivityInterval(item.interval)
  }));
}

async function loadActivityList() {
  try {
    const payload = await legacyActivities();
    const items = normalizeActivityItems(payload);
    if (items.length === 0) return;
    activityItems.value = items;
    activityCurrentIndex.value = 0;
    activityScrollTimeLeft = flashActivityInterval(items[Math.min(5, items.length - 1)]?.interval);
  } catch {
    // Keep the current activity text if the legacy command adapter is unavailable.
  }
}

function parseGuideRect(raw: string | undefined) {
  if (!raw) return null;
  const parts = raw.split(",").map((part) => Number(part.trim()));
  if (parts.length < 4 || parts.some((part) => Number.isNaN(part))) return null;
  return {
    x: parts[0],
    y: parts[1],
    w: parts[2],
    h: parts[3]
  };
}

function isMainCityView(view?: CityView) {
  return screen.value === "city" &&
    !worldMapPanelVisible.value &&
    !battlePanelVisible.value &&
    (!view || cityView.value === view);
}

function guideMatchesCurrentScene(guide: Guide | null) {
  if (!guide || screen.value !== "city") return false;
  switch (guide.gid) {
    case 6:
    case 8:
      return isMainCityView("inner") &&
        !buildPanel.value &&
        !buildingPanel.value &&
        !speedGoodsPanel.value &&
        !taskPanelVisible.value;
    case 7:
      return isMainCityView("inner") && !!buildPanel.value;
    case 9:
      return isMainCityView("inner") && !!buildingPanel.value && !speedGoodsPanel.value;
    case 10:
      return !!speedGoodsPanel.value;
    case 11:
      return true;
    case 12:
    case 13:
      return !!taskPanelVisible.value;
    case 14:
      return isMainCityView();
    default:
      return true;
  }
}

function estimateGuideTextHeight(text: string) {
  const normalized = text.replace(/<[^>]*>/g, "");
  const lines = normalized.split("\n").flatMap((line) => {
    const length = Math.max(1, line.trim().length);
    return Array.from({ length: Math.max(1, Math.ceil(length / GUIDE_TEXT_CHARS_PER_LINE)) });
  });
  return Math.max(GUIDE_TEXT_MIN_HEIGHT, lines.length * GUIDE_TEXT_LINE_HEIGHT + 20);
}

function clampGuideArrow(arrow: { x: number; y: number }) {
  return {
    x: Math.min(STAGE_WIDTH - GUIDE_ARROW_WIDTH - GUIDE_TIP_MARGIN, Math.max(0, arrow.x)),
    y: Math.min(STAGE_HEIGHT - GUIDE_ARROW_HEIGHT - GUIDE_TIP_MARGIN, Math.max(0, arrow.y))
  };
}

function correctGuideTip(tip: { x: number; y: number }) {
  let corrected = false;
  let x = tip.x;
  let y = tip.y;
  if (x <= 0) {
    x = 0;
    corrected = true;
  } else if (x + GUIDE_TIP_WIDTH > STAGE_WIDTH) {
    x = STAGE_WIDTH - GUIDE_TIP_WIDTH - GUIDE_TIP_MARGIN;
    corrected = true;
  }
  if (y <= 0) {
    y = 0;
    corrected = true;
  } else if (y + guideTipMetrics.value.height > STAGE_HEIGHT) {
    y = STAGE_HEIGHT - guideTipMetrics.value.height - GUIDE_TIP_MARGIN;
    corrected = true;
  }
  return { x, y, corrected };
}

function tryGuidePlacement(
  direction: number,
  rect: { x: number; y: number; w: number; h: number },
  hiddenArrow = false
) {
  let arrow = { x: rect.x + (rect.w - GUIDE_ARROW_WIDTH) / 2, y: rect.y + rect.h };
  let image = "arrow_1.png";
  let transform = "";
  let tip = { x: rect.x + (GUIDE_ARROW_WIDTH - GUIDE_TIP_WIDTH) / 2, y: rect.y + rect.h + GUIDE_ARROW_HEIGHT };
  if (direction === 2) {
    image = "arrow_2.png";
    arrow = { x: rect.x - rect.w, y: rect.y + (rect.h - GUIDE_ARROW_HEIGHT) / 2 };
    tip = { x: arrow.x - GUIDE_TIP_WIDTH, y: arrow.y + (GUIDE_ARROW_HEIGHT - guideTipMetrics.value.height) / 2 };
  } else if (direction === 3) {
    transform = "rotate(180deg)";
    arrow = { x: rect.x + (rect.w - GUIDE_ARROW_WIDTH) / 2, y: rect.y - GUIDE_ARROW_HEIGHT };
    tip = { x: rect.x + (GUIDE_ARROW_WIDTH - GUIDE_TIP_WIDTH) / 2, y: arrow.y - guideTipMetrics.value.height };
  } else if (direction === 4) {
    image = "arrow_2.png";
    transform = "rotate(180deg)";
    arrow = { x: rect.x + rect.w, y: rect.y + (rect.h - GUIDE_ARROW_HEIGHT) / 2 };
    tip = { x: arrow.x + GUIDE_ARROW_WIDTH, y: arrow.y };
  }
  arrow = clampGuideArrow(arrow);
  const correctedTip = correctGuideTip(tip);
  return {
    corrected: correctedTip.corrected,
    tip: { x: Math.round(correctedTip.x), y: Math.round(correctedTip.y) },
    arrow: {
      image,
      x: Math.round(arrow.x),
      y: Math.round(arrow.y),
      w: GUIDE_ARROW_WIDTH,
      h: GUIDE_ARROW_HEIGHT,
      hidden: hiddenArrow,
      transform
    }
  };
}

function updateStageScale() {
  const width = window.innerWidth || STAGE_WIDTH;
  const height = window.innerHeight || STAGE_HEIGHT;
  stageScale.value = Math.min(1, width / STAGE_WIDTH, height / STAGE_HEIGHT);
}

function guideByID(gid: number) {
  return guideItems.value.find((item) => item.gid === gid) ?? null;
}

function setGuide(gid: number) {
  currentGuide.value = guideByID(gid);
  guideVisible.value = !!currentGuide.value;
}

function showNextGuideFrom(gid: number) {
  const next = guideItems.value.find((item) => item.pregid === gid);
  if (!next) {
    guideVisible.value = false;
    return;
  }
  currentGuide.value = next;
  guideVisible.value = true;
}

function firstGuideBuildPlot() {
  const rect = guideRect.value;
  const candidates = innerPlots.filter((plot) => !positionBuilding(plot.position));
  if (rect) {
    const guided = candidates.find((plot) => plotIntersectsGuide(plot, rect));
    if (guided) return guided;
  }
  return candidates[0] ?? null;
}

function builtHousePlot() {
  const house =
    city.value?.buildings.find((building) => building.bid === 5 && building.state !== 0) ??
    city.value?.buildings.find((building) => building.bid === 5);
  if (!house) return null;
  return innerPlots.find((plot) => plot.position === house.position) ?? null;
}

async function handleGuideHotspotClick() {
  if (!effectiveGuideVisible.value) return;
  const gid = currentGuide.value?.gid;
  if (!gid) return;

  if (gid === 6) {
    const plot = firstGuideBuildPlot();
    if (plot) await openBuild(plot);
    return;
  }

  if (gid === 7) {
    const option = buildPanel.value?.options.find((item) => item.bid === 5 && item.canBuild);
    if (option) {
      await confirmBuild(option);
    }
    return;
  }

  if (gid === 8) {
    const plot = builtHousePlot();
    if (plot) await openBuild(plot);
    return;
  }

  if (gid === 9) {
    await requestSpeedSelectedBuilding();
    return;
  }

  if (gid === 10) {
    const item = speedGoodsPanel.value?.goodsList.find((goods) => goods.gid === 67 && goods.count > 0);
    if (item) {
      await useSpeedGoods(item);
    }
    return;
  }

  if (gid === 11) {
    await openTaskPanel();
    return;
  }

  if (gid === 12) {
    const firstTask = taskSnapshot.value?.categories?.flatMap((category) => category.groups ?? [])
      .flatMap((group) => group.tasks ?? [])[0];
    if (firstTask) await selectTask(firstTask.id);
    return;
  }

  if (gid === 13) {
    await handleClaimReward();
  }
}

async function openBuild(plot: Plot) {
  if (!city.value) return;
  buildingListDialogVisible.value = false;
  const requestSeq = ++buildRequestSeq;
  const guideLocked = effectiveGuideVisible.value && currentGuide.value?.distype === 1;
  const targetRect = guideRect.value;
  const currentGuideID = currentGuide.value?.gid ?? 0;
  const redirectedPlot =
    guideLocked && currentGuideID === 8 ? builtHousePlot() :
    null;
  const activePlot = redirectedPlot ?? plot;
  const bypassGuideRect =
    (currentGuideID === 8 && !!positionBuilding(activePlot.position) && positionBuilding(activePlot.position)?.bid === 5);
  if (guideLocked && targetRect && !bypassGuideRect && !plotIntersectsGuide(activePlot, targetRect)) return;
  const wasGuideTarget = guideLocked && (bypassGuideRect || !targetRect || plotIntersectsGuide(plot, targetRect));
  selectedPlot.value = activePlot;
  selectedBid.value = null;
  buildPanel.value = null;
  buildingPanel.value = null;
  buildingInfoPanel.value = null;

  const occupied = positionBuilding(activePlot.position);
  if (occupied) {
    await withLoading(async () => {
      const info = await buildingInfo(city.value!.summary.cid, occupied.position);
      if (requestSeq !== buildRequestSeq) return;
      buildingInfoPanel.value = info;
      buildingPanel.value = info.building;
    });
    if (requestSeq !== buildRequestSeq) return;
    if (wasGuideTarget && currentGuide.value?.gid === 8 && occupied.bid === 5) showNextGuideFrom(8);
    return;
  }

  await withLoading(async () => {
    const options = await buildingOptions(city.value!.summary.cid, activePlot.position);
    if (requestSeq !== buildRequestSeq) return;
    buildPanel.value = options;
    selectedBid.value = options.options.find((item) => item.canBuild)?.bid ?? null;
  });
  if (requestSeq !== buildRequestSeq) return;
  if (wasGuideTarget && buildPanel.value && currentGuide.value?.gid === 6) showNextGuideFrom(6);
}

async function openGovernmentFromCityManage() {
  hideBuildingTip();
  leaveGuideForManualNavigation();
  await openBuild(cityHallPlot);
}

async function openGovernmentSubDialog(kind: "tax" | "name" | "pacify" | "levy" | "produce" | "fields" | "cities") {
  governmentSubDialog.value = kind;
  governmentTaxValue.value = city.value?.tax ?? 0;
  governmentCityName.value = city.value?.summary.name ?? "";
  if (kind === "cities") {
    await withLoading(async () => {
      governmentCityItems.value = (await myCities()).items;
    });
  }
}

function closeGovernmentSubDialog() {
  governmentSubDialog.value = "";
}

async function openLeftBuildingList() {
  hideBuildingTip();
  leaveGuideForManualNavigation();
  closeBuild();
  buildingListDialogVisible.value = true;
}

function closeBuildingListDialog() {
  buildingListDialogVisible.value = false;
}

function toggleTopBuildingList() {
  craftDialogVisible.value = false;
  tacticDialogVisible.value = false;
  showTopBuildingList.value = !showTopBuildingList.value;
}

function openCraftDialog() {
  closeMainSceneOverlays();
  showTopBuildingList.value = false;
  tacticDialogVisible.value = false;
  craftDialogVisible.value = true;
}

function closeCraftDialog() {
  craftDialogVisible.value = false;
}

async function openCraftBuildingOverview() {
  closeCraftDialog();
  await openLeftBuildingList();
}

async function openCraftProduceDialog() {
  closeCraftDialog();
  await openGovernmentFromCityManage();
  await openGovernmentSubDialog("produce");
}

function openTacticDialog() {
  closeMainSceneOverlays();
  showTopBuildingList.value = false;
  craftDialogVisible.value = false;
  tacticDialogVisible.value = true;
}

function closeTacticDialog() {
  tacticDialogVisible.value = false;
}

function openTacticWorldMap() {
  closeTacticDialog();
  void openWorldMapPanel();
}

function openTacticBattle() {
  closeTacticDialog();
  openBattlePanel();
}

function bottomFunctionImage(name: string) {
  const suffix =
    activeBottomFunction.value === name ? "_down" :
    hoveredBottomFunction.value === name ? "_on" :
    "";
  return asset(`function_${name}${suffix}.png`);
}

function topButtonImage(name: string, selected = false, hasDown = true) {
  const suffix =
    pressedTopButton.value === name && hasDown ? "_down" :
    selected ? "_on" :
    hoveredTopButton.value === name && topButtonHoverOnAssets.has(name) ? "_on" :
    hoveredTopButton.value === name && topButtonOverAssets.has(name) ? "_over" :
    "";
  return asset(`topbutton_${name}${suffix}.png`);
}

async function openTopBuildingQueueItem(building: Building) {
  const plot =
    building.bid === 6 ? cityHallPlot :
    building.bid === 20 ? cityWallPlot :
    innerPlots.find((item) => item.position === building.position) ??
      visibleOuterPlots.value.find((item) => item.position === building.position) ??
      { position: building.position, gridX: 0, gridY: 0, x: 0, y: 0, w: 0, h: 0 };
  await openBuild(plot);
}

async function openBuildingFromList(building: Building) {
  buildingListDialogVisible.value = false;
  cityView.value = building.bid === 20 || buildingListOuterItems.value.some((item) => item.position === building.position) ? "outer" : "inner";
  await openTopBuildingQueueItem(building);
}

async function openLeftResourceProduce() {
  await openGovernmentFromCityManage();
  await openGovernmentSubDialog("produce");
}

async function openLeftTax() {
  await openGovernmentFromCityManage();
  await openGovernmentSubDialog("tax");
}

async function openLeftCityFields() {
  await openGovernmentFromCityManage();
  await openGovernmentSubDialog("fields");
}

async function inspectGovernmentField(row: { x: number; y: number }) {
  closeBuild();
  closeGovernmentSubDialog();
  await openWorldMapPanel();
  setWorldMapCenter({ x: row.x, y: row.y });
  selectWorldGrid(row.x, row.y);
}

function showMissingFlashDialog(name: string) {
  closeMainSceneOverlays();
  status(`${name}窗口还没有接入。`);
}

function openLordDialog() {
  closeMainSceneOverlays();
  lordDialogVisible.value = true;
}

function openEquipmentDialog() {
  closeMainSceneOverlays();
  equipmentDialogVisible.value = true;
}

function openTreasureDialog() {
  closeMainSceneOverlays();
  treasureDialogVisible.value = true;
}

function selectChatChannel(channel: string) {
  chatChannelLabel.value = channel;
  chatChannelMenuVisible.value = false;
}

function handleChatSend() {
  closeMainSceneOverlays();
  chatSendDialogVisible.value = true;
}

function handleChatControl() {
  closeMainSceneOverlays();
  chatControlDialogVisible.value = true;
}

function showAnnouncementHover() {
  hoveredAnnounce.value = true;
  setBuildingTipAtPoint(675, 453, "公告栏", -70);
}

function hideAnnouncementHover() {
  hoveredAnnounce.value = false;
  hideBuildingTip();
}

function openAnnouncement() {
  announcementVisible.value = true;
  hideAnnouncementHover();
}

async function openUseGoodsDialog(type: number) {
  useGoodsPanel.value = {
    visible: true,
    type,
    loading: true,
    error: "",
    goodsList: []
  };
  try {
    const payload = await userTypeGoods(type);
    useGoodsPanel.value = {
      visible: true,
      type,
      loading: false,
      error: "",
      goodsList: payload.goodsList ?? []
    };
  } catch (err) {
    useGoodsPanel.value = {
      ...useGoodsPanel.value,
      loading: false,
      error: err instanceof Error && err.message !== "404 Not Found" ? err.message : ""
    };
  }
}

function closeUseGoodsDialog() {
  useGoodsPanel.value = {
    visible: false,
    type: 0,
    loading: false,
    error: "",
    goodsList: []
  };
}

function requestFlashConfirm(message: string, title = "确认") {
  if (flashConfirm.value.resolve) {
    flashConfirm.value.resolve(false);
  }
  return new Promise<boolean>((resolve) => {
    flashConfirm.value = {
      visible: true,
      title,
      message,
      resolve
    };
  });
}

function resolveFlashConfirm(value: boolean) {
  const resolve = flashConfirm.value.resolve;
  flashConfirm.value = {
    visible: false,
    title: "确认",
    message: "",
    resolve: null
  };
  resolve?.(value);
}

async function useGeneralGoods(item: UserTypeGoodsItem) {
  if (!city.value) return;
  if (item.count <= 0) {
    status(`${item.name}数量不足。`);
    return;
  }
  if (!(await requestFlashConfirm(`确定使用${item.name}?`))) return;
  const goodsType = useGoodsPanel.value.type;
  await withLoading(async () => {
    setCityDetail(await useUserTypeGoods(goodsType, item.gid, city.value!.summary.cid));
    useGoodsPanel.value = {
      visible: true,
      type: goodsType,
      loading: false,
      error: "",
      goodsList: (await userTypeGoods(goodsType)).goodsList ?? []
    };
  });
}

function plotStageRect(plot: Plot) {
  return {
    x: CITY_VIEW_LEFT + plot.x,
    y: CITY_VIEW_TOP + plot.y,
    w: plot.w,
    h: plot.h
  };
}

function plotIntersectsGuide(plot: Plot, rect: { x: number; y: number; w: number; h: number }) {
  const stagePlot = plotStageRect(plot);
  return stagePlot.x < rect.x + rect.w && stagePlot.x + stagePlot.w > rect.x && stagePlot.y < rect.y + rect.h && stagePlot.y + stagePlot.h > rect.y;
}

async function confirmBuild(option?: BuildingOption) {
  const bid = option?.bid ?? selectedBid.value;
  if (!city.value || !selectedPlot.value || !bid) return;
  if (effectiveGuideVisible.value && currentGuide.value?.gid === 7 && bid !== 5) return;
  await withLoading(async () => {
    setCityDetail(await createBuilding(city.value!.summary.cid, selectedPlot.value!.position, bid));
    buildPanel.value = null;
    selectedPlot.value = null;
    selectedBid.value = null;
  });
  if (currentGuide.value?.gid === 7 && bid === 5) showNextGuideFrom(7);
}

function closeBuild() {
  buildRequestSeq++;
  buildPanel.value = null;
  buildingPanel.value = null;
  buildingInfoPanel.value = null;
  speedGoodsPanel.value = null;
  selectedPlot.value = null;
  selectedBid.value = null;
}

async function upgradeSelectedBuilding() {
  if (!city.value || !selectedPlot.value || !buildingPanel.value) return;
  await withLoading(async () => {
    const position = buildingPanel.value!.position;
    setCityDetail(await upgradeBuilding(city.value!.summary.cid, position));
    buildingPanel.value = city.value?.buildings.find((item) => item.position === position) ?? null;
    buildingInfoPanel.value = buildingPanel.value ? await buildingInfo(city.value!.summary.cid, position) : null;
  });
}

async function requestDestroySelectedBuilding() {
  if (!buildingPanel.value) return;
  if (buildingPanel.value.bid === 6 && buildingPanel.value.level <= 1) {
    status("官府1级不能拆除。");
    return;
  }
  if (!(await requestFlashConfirm(`确定拆除${buildingPanel.value.name}(等级${buildingPanel.value.level})?`))) return;
  await destroySelectedBuilding();
}

async function destroySelectedBuilding() {
  if (!city.value || !selectedPlot.value || !buildingPanel.value) return;
  await withLoading(async () => {
    setCityDetail(await destroyBuilding(city.value!.summary.cid, buildingPanel.value!.position));
    closeBuild();
  });
}

async function cancelSelectedBuildingAction() {
  if (!city.value || !selectedPlot.value || !buildingPanel.value) return;
  await withLoading(async () => {
    const position = buildingPanel.value!.position;
    setCityDetail(await cancelBuildingAction(city.value!.summary.cid, position));
    buildingPanel.value = city.value?.buildings.find((item) => item.position === position) ?? null;
    buildingInfoPanel.value = buildingPanel.value ? await buildingInfo(city.value!.summary.cid, position) : null;
  });
}

async function requestSpeedSelectedBuilding() {
  if (!city.value || !buildingPanel.value) return;
  const position = buildingPanel.value.position;
  await withLoading(async () => {
    speedGoodsPanel.value = await buildingSpeedGoods(city.value!.summary.cid, position);
  });
  if (currentGuide.value?.gid === 9) showNextGuideFrom(9);
}

async function openTaskPanel() {
  leaveSceneForBottomFunction();
  await withLoading(async () => {
    taskSnapshot.value = await myTasks();
    const firstCategory = taskSnapshot.value.categories[0] ?? null;
    selectedTaskCategoryType.value = firstCategory?.type ?? null;
    const firstGroup = firstCategory?.groups[0] ?? null;
    selectedTaskGroupId.value = firstGroup?.id ?? null;
    const firstTask = firstGroup?.tasks[0] ?? null;
    selectedTaskId.value = firstTask?.id ?? null;
    taskPanelVisible.value = true;
  });
  if (currentGuide.value?.gid === 11) setGuide(12);
}

async function openBattleTaskPanel() {
  taskSnapshot.value = makeLocalTaskSnapshot();
  selectedTaskCategoryType.value = BATTLE_TASK_CATEGORY_TYPE;
  selectedTaskGroupId.value = null;
  selectedTaskId.value = null;
  taskPanelVisible.value = true;
  loading.value = true;
  try {
    const snapshot = await battleTasks({
      bid: battleActiveMapId.value || battleMapId.value,
      unionId: battleFieldSnapshot.value?.unionId || battleUnionId.value || 1
    });
    const battleCategory = snapshot.categories.find((category) => category.type === BATTLE_TASK_CATEGORY_TYPE) ?? emptyBattleTaskCategory.value;
    taskSnapshot.value = makeLocalTaskSnapshot([battleCategory]);
    const firstGroup = battleCategory.groups[0] ?? null;
    const firstTask = firstGroup?.tasks[0] ?? null;
    selectedTaskCategoryType.value = BATTLE_TASK_CATEGORY_TYPE;
    selectedTaskGroupId.value = firstGroup?.id ?? null;
    selectedTaskId.value = firstTask?.id ?? null;
    taskPanelVisible.value = true;
  } catch {
    taskSnapshot.value = makeLocalTaskSnapshot();
  } finally {
    loading.value = false;
  }
}

async function selectTask(taskId: number) {
  selectedTaskId.value = taskId;
  if (currentGuide.value?.gid === 12) setGuide(13);
}

async function handleClaimReward() {
  const taskId = selectedTaskId.value;
  if (!taskId) return;
  if (isBattleTaskPanel.value) {
    status("战场任务奖励写接口未接入，当前仅可查看任务进度。");
    return;
  }
  await withLoading(async () => {
    taskSnapshot.value = await claimTaskReward(taskId);
    const firstCategory = taskSnapshot.value.categories.find((category) => category.type === selectedTaskCategoryType.value) ?? taskSnapshot.value.categories[0] ?? null;
    selectedTaskCategoryType.value = firstCategory?.type ?? null;
    const firstGroup = firstCategory?.groups.find((group) => group.id === selectedTaskGroupId.value) ?? firstCategory?.groups[0] ?? null;
    selectedTaskGroupId.value = firstGroup?.id ?? null;
    const nextTask = firstGroup?.tasks.find((task) => task.id === taskId) ?? firstGroup?.tasks[0] ?? null;
    selectedTaskId.value = nextTask?.id ?? null;
    taskPanelVisible.value = false;
  });
  if (currentGuide.value?.gid === 13) setGuide(14);
}

async function useSpeedGoods(item: SpeedGoods) {
  if (!city.value || !speedGoodsPanel.value) return;
  if (item.count <= 0) {
    status(`${item.name}数量不足。`);
    return;
  }
  if (!(await requestFlashConfirm(`确定使用${item.name}?`))) return;
  const position = speedGoodsPanel.value.position;
  await withLoading(async () => {
    setCityDetail(await useBuildingSpeedGoods(city.value!.summary.cid, position, item.gid));
    buildingPanel.value = city.value?.buildings.find((building) => building.position === position) ?? null;
    buildingInfoPanel.value = buildingPanel.value ? await buildingInfo(city.value!.summary.cid, position) : null;
    speedGoodsPanel.value = null;
  });
  if (currentGuide.value?.gid === 10 && item.gid === 67) {
    const house = city.value.buildings.find((building) => building.bid === 5 && building.level >= 1 && building.state === 0);
    if (house) setGuide(11);
  }
}

// Mail dialog
function closeMailPanel() {
  mailPanelVisible.value = false;
  mailDetailView.value = null;
  if (mailFolder.value === "compose") mailFolder.value = "inbox";
}

function closeReportPanel() {
  reportPanelVisible.value = false;
  reportDetailView.value = null;
}

function closeShopPanel() {
  shopPanelVisible.value = false;
}

function closeRankPanel() {
  rankPanelVisible.value = false;
}

function closeFriendDialog() {
  friendDialogVisible.value = false;
}

function closeStatDialog() {
  statDialogVisible.value = false;
}

function closeForumDialog() {
  forumDialogVisible.value = false;
}

function closeWebsiteDialog() {
  websiteDialogVisible.value = false;
}

function closeHelpDialog() {
  helpDialogVisible.value = false;
}

function closeChargeDialog() {
  chargeDialogVisible.value = false;
}

function closeUtilityDialogs() {
  closeMailPanel();
  closeReportPanel();
  closeShopPanel();
  closeRankPanel();
  closeFriendDialog();
  closeStatDialog();
  closeForumDialog();
  closeWebsiteDialog();
  closeHelpDialog();
  closeChargeDialog();
}

async function openMailPanel(folder: "inbox" | "sent" | "sys" = "inbox") {
  closeMainSceneOverlays();
  mailFolder.value = folder;
  await withLoading(async () => {
    const result = await mailPage(folder, mailPageInfo.value.page);
    mailItems.value = result.items;
    mailPageInfo.value = { total: result.total, page: result.page, pageSize: result.pageSize };
    mailDetailView.value = null;
    mailPanelVisible.value = true;
  });
}

async function selectMailItem(item: MailItem) {
  mailDetailView.value = item;
}

async function deleteMailItem(id: number) {
  if (!(await requestFlashConfirm("确定删除此邮件?"))) return;
  await withLoading(async () => {
    const result = await deleteMail(mailFolder.value, id, mailPageInfo.value.page);
    mailItems.value = result.items;
    mailPageInfo.value = { total: result.total, page: result.page, pageSize: result.pageSize };
    if (mailDetailView.value?.id === id) mailDetailView.value = null;
  });
}

// Report dialog
async function openReportPanel(filter: ReportFilter = "all") {
  closeMainSceneOverlays();
  reportFilter.value = filter;
  await withLoading(async () => {
    const result = await reports(filter, reportPageInfo.value.page);
    reportItems.value = result.items;
    reportPageInfo.value = { total: result.total, page: result.page, pageSize: result.pageSize };
    reportDetailView.value = null;
    reportPanelVisible.value = true;
  });
}

async function selectReportItem(item: ReportItem) {
  reportDetailView.value = item;
}

// Shop dialog
async function openShopPanel(groupId?: number) {
  closeMainSceneOverlays();
  selectedShopGroupId.value = Number.isFinite(groupId) ? groupId! : null;
  await withLoading(async () => {
    shopSnapshot.value = await myShop();
    if (selectedShopGroupId.value == null) {
      selectedShopGroupId.value = shopSnapshot.value.groups[0]?.id ?? null;
    } else if (!shopSnapshot.value.groups.some((group) => group.id === selectedShopGroupId.value)) {
      selectedShopGroupId.value = shopSnapshot.value.groups[0]?.id ?? null;
    }
    shopPanelVisible.value = true;
  });
}

async function openUseGoodsShopPanel() {
  const firstItem = useGoodsPanel.value.goodsList[0];
  const goodsGroup = Number.parseInt(String(firstItem?.group ?? ""), 10);
  await openShopPanel(Number.isFinite(goodsGroup) ? 2 + goodsGroup : undefined);
}

async function openSpeedGoodsShopPanel() {
  const firstItem = speedGoodsPanel.value?.goodsList[0];
  const goodsGroup = Number.parseInt(String(firstItem?.group ?? ""), 10);
  speedGoodsPanel.value = null;
  await openShopPanel(Number.isFinite(goodsGroup) ? 2 + goodsGroup : undefined);
}

async function handleBuyItem(item: ShopItem) {
  await withLoading(async () => {
    shopSnapshot.value = await buyShopItem(item.id, 1, 0, city.value?.summary.cid ?? 0);
    if (!shopSnapshot.value.groups.some((group) => group.id === selectedShopGroupId.value)) {
      selectedShopGroupId.value = shopSnapshot.value.groups[0]?.id ?? null;
    }
  });
}

// World map dialog
function closeBattleSceneDialogs() {
  battleMenuVisible.value = false;
  battleCampaignVisible.value = false;
  battleFieldViewVisible.value = false;
  battleActionVisible.value = false;
  battleAttackVisible.value = false;
  battlePatrolVisible.value = false;
  battleTroopViewVisible.value = false;
  battleTroopDetailVisible.value = false;
  battleInfoVisible.value = false;
  battleUsersVisible.value = false;
}

function openMailComposePanel() {
  mailDetailView.value = null;
  mailFolder.value = "compose";
}

function closeMailComposePanel() {
  mailFolder.value = "inbox";
}

function closeMainSceneOverlays() {
  closeBuild();
  closeUseGoodsDialog();
  closeGovernmentSubDialog();
  announcementVisible.value = false;
  taskPanelVisible.value = false;
  closeUtilityDialogs();
  unionPanelVisible.value = false;
  heroPanelVisible.value = false;
  barracksPanelVisible.value = false;
  collegePanelVisible.value = false;
  rankPanelVisible.value = false;
  craftDialogVisible.value = false;
  tacticDialogVisible.value = false;
  lordDialogVisible.value = false;
  equipmentDialogVisible.value = false;
  treasureDialogVisible.value = false;
  chatSendDialogVisible.value = false;
  chatControlDialogVisible.value = false;
  selectedWorldGrid.value = null;
  worldMapAlert.value = "";
  hideWorldGridTip();
  closeBattleSceneDialogs();
}

function leaveGuideForManualNavigation() {
  if (guideVisible.value) guideVisible.value = false;
}

function openCityScene(view: "inner" | "outer") {
  leaveGuideForManualNavigation();
  closeMainSceneOverlays();
  worldMapPanelVisible.value = false;
  battlePanelVisible.value = false;
  cityView.value = view;
}

function leaveSceneForBottomFunction() {
  leaveGuideForManualNavigation();
  if (!worldMapPanelVisible.value && !battlePanelVisible.value) return;
  worldMapPanelVisible.value = false;
  battlePanelVisible.value = false;
  cityView.value = "inner";
}

async function openWorldMapPanel() {
  if (!city.value) return;
  leaveGuideForManualNavigation();
  closeMainSceneOverlays();
  battlePanelVisible.value = false;
  worldMapMode.value = "city";
  worldMapPanelVisible.value = true;
  await withLoading(async () => {
    worldMapData.value = await worldMap(city.value!.summary.cid, 20);
    setWorldMapCenter(worldMapData.value.center);
  });
}

// Union dialog
async function openUnionPanel() {
  leaveSceneForBottomFunction();
  closeMainSceneOverlays();
  await withLoading(async () => {
    unionSnapshot.value = await myUnion();
    unionPanelVisible.value = true;
  });
}

function unionApplyDisabledReason(item: UnionSnapshot["applyList"][number]) {
  if (unionSnapshot.value?.application || item.isApplied) return "已有待审核联盟申请，可先取消后再申请其他联盟。";
  if (unionSnapshot.value?.permissions?.canApply === false) return "当前状态不能申请联盟。";
  if (!item.id) return "缺少联盟编号，不能申请。";
  return "";
}

async function handleApplyUnion(item: UnionSnapshot["applyList"][number]) {
  const reason = unionApplyDisabledReason(item);
  if (reason) {
    status(reason);
    return;
  }
  if (!(await requestFlashConfirm(`确定申请加入${item.name}?`))) return;
  await withLoading(async () => {
    unionSnapshot.value = await applyUnion(item.id);
  });
}

async function handleCancelUnionApply() {
  if (!unionSnapshot.value?.application) return;
  if (!(await requestFlashConfirm(`确定取消申请${unionSnapshot.value.application.unionName}?`))) return;
  await withLoading(async () => {
    unionSnapshot.value = await cancelUnionApply();
  });
}

// Hero dialog
async function openHeroPanel() {
  if (!city.value) return;
  leaveSceneForBottomFunction();
  closeMainSceneOverlays();
  await withLoading(async () => {
    heroRoster.value = await cityHeroes(city.value!.summary.cid, 100);
    heroPanelVisible.value = true;
  });
}

async function handleRecruitHero() {
  if (!city.value) return;
  const recruitId = heroRoster.value?.recruits?.[0]?.id;
  if (!recruitId) return;
  await withLoading(async () => {
    heroRoster.value = await recruitHero(city.value!.summary.cid, recruitId);
  });
}

async function openFriendDialog(tab: "friends" | "blacklist" = "friends") {
  closeMainSceneOverlays();
  friendRelationTab.value = tab;
  friendDialogVisible.value = true;
  await withLoading(async () => {
    relationSnapshot.value = await myRelations();
  });
}

function openStatDialog() {
  closeMainSceneOverlays();
  statDialogVisible.value = true;
}

function openForumDialog() {
  closeMainSceneOverlays();
  forumDialogVisible.value = true;
}

function openWebsiteDialog() {
  closeMainSceneOverlays();
  websiteDialogVisible.value = true;
}

function openHelpDialog() {
  closeMainSceneOverlays();
  helpDialogVisible.value = true;
}

async function openChargeDialog() {
  closeMainSceneOverlays();
  chargeDialogVisible.value = true;
  await withLoading(async () => {
    chargeSnapshot.value = await myCharge();
  });
}

// Barracks dialog
async function openBarracksPanel() {
  leaveSceneForBottomFunction();
  closeMainSceneOverlays();
  await withLoading(async () => {
    const result = await myTroops(100);
    troopsData.value = result;
    barracksPanelVisible.value = true;
  });
}

async function handleTrainTroop(tid: number, count: number) {
  await withLoading(async () => {
    const result = await trainTroop(tid, count);
    if (result.success) {
      const troopsResult = await myTroops(100);
      troopsData.value = troopsResult;
    }
  });
}

// College/Research dialog
async function openCollegePanel(position = 2) {
  if (!city.value) return;
  closeMainSceneOverlays();
  await withLoading(async () => {
    researchSnapshot.value = await cityResearchSnapshot(city.value!.summary.cid, position);
    collegePanelVisible.value = true;
  });
}

async function handleResearchTech(tid: number) {
  if (!city.value) return;
  await withLoading(async () => {
    await researchTech(city.value!.summary.cid, 2, tid);
    researchSnapshot.value = await cityResearchSnapshot(city.value!.summary.cid, 2);
  });
}

// Ranking dialog
async function openRankPanel(kind: RankKind = "power") {
  closeMainSceneOverlays();
  rankKind.value = kind;
  await withLoading(async () => {
    const result = await rankings(kind, rankPageInfo.value.page);
    rankItems.value = result.items;
    rankPageInfo.value = { total: result.total, page: result.page, pageSize: result.pageSize };
    rankPanelVisible.value = true;
  });
}

// Battle dialog
async function openBattlePanel() {
  leaveGuideForManualNavigation();
  closeMainSceneOverlays();
  worldMapPanelVisible.value = false;
  battlePanelVisible.value = true;
  try {
    await loadBattleFieldState();
  } catch (error) {
    status(error instanceof Error ? error.message : "战场数据读取失败。");
  }
}

function openBattleMenu() {
  battleMenuVisible.value = true;
}

function closeBattleMenu() {
  battleMenuVisible.value = false;
}

async function refreshBattleCampaignRoster() {
  if (!city.value) return;
  if (!heroRoster.value) {
    heroRoster.value = await cityHeroes(city.value.summary.cid, 100);
  }
  syncBattleCampaignRoster();
}

async function openBattleCampaignDialog() {
  battleMenuVisible.value = false;
  battleCampaignMode.value = "battle";
  resetBattleCampaignDraft();
  battleCampaignFieldName.value = battleFieldName();
  syncBattleCampaignTargetsFromField();
  battleCampaignVisible.value = true;
  await withLoading(async () => {
    await refreshBattleCampaignRoster();
    await loadBattleFieldState();
  });
  await refreshBattleCampaignPreview();
}

function openBattleInfoDialog() {
  battleMenuVisible.value = false;
  battleInfoVisible.value = true;
  battleInfoTab.value = "info";
  battleInfoPage.value = 1;
  void loadBattleFieldState();
}

function openBattleUsersDialog() {
  battleMenuVisible.value = false;
  battleUsersVisible.value = true;
  void withLoading(async () => {
    await loadBattleFieldState();
    battleMembersSnapshot.value = await battleMembers();
    battleInviteCreator.value = battleMembersSnapshot.value.isCreator;
  });
}

function openBattleTaskDialog() {
  battleMenuVisible.value = false;
  void openBattleTaskPanel();
}

async function quitBattleField() {
  battleMenuVisible.value = false;
  await withLoading(async () => {
    const fallback = "战场退出接口未接入。";
    status(fallback);
    try {
      const preview = await battleQuitPreview();
      status(`${fallback}${preview.message ? ` ${preview.message}` : ""}`);
    } catch {
      status(fallback);
    }
  });
}

function closeBattleCampaignDialog() {
  battleCampaignVisible.value = false;
  battleCampaignPreviewSeq++;
  resetBattleCampaignDraft();
}

function battleFieldName() {
  return battleFieldSnapshot.value?.fieldName || battleCampaignFieldName.value || "城池争夺";
}

function battleCityName(cid: number) {
  const item = battleFieldSnapshot.value?.cities.find((cityItem) => cityItem.cid === cid);
  return item?.name || (cid > 0 ? `城池${formatCityCode(cid)}` : "-");
}

function battleFieldTargetLabel(row: BattleFieldTroopRow) {
  return row.targetCid > 0 ? battleCityName(row.targetCid) : battleFieldName();
}

function battleFieldSendTargetCid() {
  const troop = battleSelectedCurrentTroop.value;
  if (!troop?.id) return 0;
  const selectedTarget = selectedBattleFieldRow.value?.targetCid || selectedBattleFieldRow.value?.cid || 0;
  const campaignTarget = battleCampaignMode.value === "battle" && battleCampaignTargets.value.some((item) => item.id === battleCampaignTargetId.value)
    ? battleCampaignTargetCid()
    : 0;
  const openCityTarget = battleFieldSnapshot.value?.cities.find((item) => item.cid > 0 && item.cid !== troop.cid)?.cid || 0;
  return selectedTarget || campaignTarget || openCityTarget || 0;
}

function selectBattleFieldRow(row: BattleFieldTroopRow) {
  selectedBattleFieldRow.value = row;
  selectedBattleCurrentTroop.value = null;
  selectedBattleTroopDetail.value = null;
  battleSendPreviewData.value = null;
  battleAttackPreviewData.value = null;
  battlePatrolPreviewData.value = null;
}

function selectBattleCurrentTroop(troop: BattleFieldCurrentTroop) {
  selectedBattleCurrentTroop.value = troop;
  selectedBattleFieldRow.value = null;
  selectedBattleTroopDetail.value = null;
  battleSendPreviewData.value = null;
  battleAttackPreviewData.value = null;
  battlePatrolPreviewData.value = null;
}

function ensureBattleTroopDetailSelection() {
  if (selectedBattleFieldRow.value || selectedBattleCurrentTroop.value) return;
  const firstCurrentTroop = battleCurrentTroops.value[0];
  if (firstCurrentTroop) {
    selectBattleCurrentTroop(firstCurrentTroop);
    return;
  }
  const firstRow = battleFieldRows.value[0];
  if (firstRow) selectBattleFieldRow(firstRow);
}

function applyBattleNewsPage(page: BattleFieldNewsPage) {
  battleInfoPage.value = page.page;
  battleNewsTotal.value = page.total;
  battleNewsItems.value = page.items.map((item: BattleFieldNewsItem) => ({
    time: item.time || formatBattleNewsTime(item.logTime),
    evtContent: item.content,
    color: item.color
  }));
}

async function loadBattleNewsPage(page: number) {
  const snapshot = battleFieldSnapshot.value;
  if (!snapshot) return;
  await withLoading(async () => {
    const newsPage = await battleNews({
      battlefieldId: snapshot.battlefieldId,
      unionId: snapshot.unionId,
      page,
      pageSize: battleNewsPageSize
    });
    applyBattleNewsPage(newsPage);
  });
}

async function loadBattleFieldState() {
  const cid = selectedCityId.value;
  if (!cid) return;
  await withLoading(async () => {
    battleFieldSnapshot.value = await battleFieldState({
      battlefieldId: battleMapId.value,
      unionId: battleUnionId.value,
      cid,
      name: battleCampaignFieldName.value || "城池争夺"
    });
    battleInfoPage.value = 1;
    battleNewsTotal.value = battleFieldSnapshot.value.newsTotal ?? battleFieldSnapshot.value.news?.length ?? 0;
    battleNewsItems.value = (battleFieldSnapshot.value.news ?? []).map((item: BattleFieldNewsItem) => ({
      time: item.time || formatBattleNewsTime(item.logTime),
      evtContent: item.content,
      color: item.color
    }));
    ensureBattleTroopDetailSelection();
    if (battleCampaignMode.value === "battle") syncBattleCampaignTargetsFromField();
  });
}

async function openBattleFieldViewDialog() {
  battleMenuVisible.value = false;
  battleActionVisible.value = false;
  battleAttackVisible.value = false;
  battlePatrolVisible.value = false;
  battleTroopViewVisible.value = false;
  battleTroopDetailVisible.value = false;
  battleSendPreviewData.value = null;
  battleAttackPreviewData.value = null;
  battlePatrolPreviewData.value = null;
  battleFieldViewVisible.value = true;
  try {
    await loadBattleFieldState();
    ensureBattleTroopDetailSelection();
  } catch (error) {
    status(error instanceof Error ? error.message : "战场部队读取失败。");
  }
}

function closeBattleFieldViewDialog() {
  battleFieldViewVisible.value = false;
  battleActionVisible.value = false;
  battleAttackVisible.value = false;
  battlePatrolVisible.value = false;
}

async function openBattleActionDialog() {
  battleMenuVisible.value = false;
  battleTroopDetailVisible.value = false;
  battleAttackVisible.value = false;
  battlePatrolVisible.value = false;
  battleSendPreviewData.value = null;
  if (!battleFieldSnapshot.value) await loadBattleFieldState();
  const troop = battleSelectedCurrentTroop.value;
  const target = battleFieldSendTargetCid();
  if (!troop?.id || target <= 0) {
    status(battleFieldSendDisabledReason.value || "暂无可派遣部队或目标。");
    return;
  }
  if (troop?.id && target > 0) {
    await withLoading(async () => {
      battleSendPreviewData.value = await battleArmySendPreview({
        troopId: troop.id,
        targetCid: target,
        targetName: battleCityName(target)
      });
    });
  }
  battleActionVisible.value = true;
}

function closeBattleActionDialog() {
  battleActionVisible.value = false;
}

async function handleBattleFieldPatrol(row: BattleFieldTroopRow) {
  selectBattleFieldRow(row);
  if (!(await requestFlashConfirm("确定巡逻该部队?"))) return;
  battleActionVisible.value = false;
  battleAttackVisible.value = false;
  battleTroopDetailVisible.value = false;
  battlePatrolPreviewData.value = null;
  const troop = battleSelectedCurrentTroop.value;
  if (!troop?.id || !row.id) {
    status("请选择己方战场部队后再巡逻。");
    return;
  }
  await withLoading(async () => {
    battlePatrolPreviewData.value = await battlePatrolPreview({
      troopId: troop.id,
      targetTroopId: row.id
    });
  });
  battlePatrolVisible.value = true;
}

function closeBattlePatrolDialog() {
  battlePatrolVisible.value = false;
}

async function openBattleAttackDialog() {
  battleMenuVisible.value = false;
  battleTroopDetailVisible.value = false;
  battleActionVisible.value = false;
  battlePatrolVisible.value = false;
  battleAttackPreviewData.value = null;
  if (!battleFieldSnapshot.value) await loadBattleFieldState();
  const troop = battleSelectedCurrentTroop.value;
  const target = selectedBattleFieldRow.value;
  if (troop?.id && target?.id) {
    await withLoading(async () => {
      battleAttackPreviewData.value = await battleArmyAttackPreview({
        troopId: troop.id,
        targetTroopId: target.id,
        targetName: target.name || battleFieldTargetLabel(target)
      });
    });
  }
  battleAttackVisible.value = true;
}

function closeBattleAttackDialog() {
  battleAttackVisible.value = false;
}

async function openBattleTroopViewDialog(forceList = false) {
  battleMenuVisible.value = false;
  battleFieldViewVisible.value = false;
  battleActionVisible.value = false;
  battleAttackVisible.value = false;
  battlePatrolVisible.value = false;
  battleTroopDetailVisible.value = false;
  selectedBattleTroopDetail.value = null;
  try {
    await loadBattleFieldState();
  } catch (error) {
    status(error instanceof Error ? error.message : "战场部队读取失败。");
  }
  if (!forceList && battleCurrentTroops.value[0]) {
    await openBattleTroopDetailDialog(undefined, battleCurrentTroops.value[0]);
    return;
  }
  battleTroopViewVisible.value = true;
}

function closeBattleTroopViewDialog() {
  battleTroopViewVisible.value = false;
}

async function openBattleTroopDetailDialog(row?: BattleFieldTroopRow, troop?: BattleFieldCurrentTroop) {
  if (row) selectBattleFieldRow(row);
  if (troop) selectBattleCurrentTroop(troop);
  ensureBattleTroopDetailSelection();
  selectedBattleTroopDetail.value = null;
  battleFieldViewVisible.value = false;
  battleActionVisible.value = false;
  battleAttackVisible.value = false;
  battlePatrolVisible.value = false;
  battleTroopViewVisible.value = false;
  if (!battleFieldSnapshot.value) await loadBattleFieldState();
  const troopID = row?.id ?? troop?.id ?? selectedBattleFieldRow.value?.id ?? selectedBattleCurrentTroop.value?.id ?? 0;
  if (troopID > 0) {
    await withLoading(async () => {
      selectedBattleTroopDetail.value = await battleTroopDetail(troopID);
    });
  }
  battleTroopDetailVisible.value = true;
}

function closeBattleTroopDetailDialog() {
  battleTroopDetailVisible.value = false;
}

function handleBattleTroopDetailPreview() {
  void openBattleTroopDetailDialog();
}

function handleBattleAttackPreview() {
  void openBattleAttackDialog();
}

function closeBattleInfoDialog() {
  battleInfoVisible.value = false;
}

function closeBattleUsersDialog() {
  battleUsersVisible.value = false;
}

function setBattleInfoTab(tab: "info" | "news") {
  battleInfoTab.value = tab;
  if (tab === "news") void loadBattleNewsPage(battleInfoPage.value);
}

function prevBattleInfoPage() {
  const page = Math.max(1, battleInfoPage.value - 1);
  if (page !== battleInfoPage.value) void loadBattleNewsPage(page);
}

function nextBattleInfoPage() {
  const page = Math.min(battleNewsPageCount.value, battleInfoPage.value + 1);
  if (page !== battleInfoPage.value) void loadBattleNewsPage(page);
}

function toggleBattleCampaignUseFlag() {
  battleCampaignUseFlag.value = !battleCampaignUseFlag.value;
  void refreshBattleCampaignPreview();
}

function resetBattleCampaignDraft() {
  battleCampaignUseFlag.value = false;
  battleCampaignHeroId.value = "0";
  battleCampaignTask.value = 3;
  battleCampaignStartReason.value = "";
  battleCampaignPreviewData.value = null;
  for (const item of battleCampaignSoldiers.value) {
    item.takecount = 0;
  }
}

function battleCampaignTargetCid() {
  const target = battleCampaignTargetId.value;
  const firstNumber = Number.parseInt(target.split(",")[0] ?? target, 10);
  return Number.isFinite(firstNumber) && firstNumber > 0 ? firstNumber : 0;
}

function worldTargetId(grid = selectedWorldGrid.value) {
  if (!grid) return 0;
  return grid.targetCid || grid.x + grid.y * 1000;
}

function syncBattleCampaignTargetsFromField() {
  const targetCities = battleFieldSnapshot.value?.cities.filter((item) => item.cid > 0) ?? [];
  if (targetCities.length === 0) {
    battleCampaignTargets.value = [{ id: "battle", name: "战场目标" }];
    battleCampaignTargetId.value = "battle";
    void refreshBattleCampaignPreview();
    return;
  }
  battleCampaignTargets.value = targetCities.map((item) => ({
    id: String(item.cid),
    name: `${item.name}${item.flagLabel ? `(${item.flagLabel})` : ""}`
  }));
  if (!battleCampaignTargets.value.some((target) => target.id === battleCampaignTargetId.value)) {
    battleCampaignTargetId.value = battleCampaignTargets.value[0]?.id ?? "battle";
  }
  void refreshBattleCampaignPreview();
}

async function refreshBattleCampaignPreview() {
  if (battleCampaignMode.value !== "battle" || !city.value || !battleCampaignVisible.value) return;
  const seq = ++battleCampaignPreviewSeq;
  try {
    const preview = await battleCampaignPreview({
      cid: city.value.summary.cid,
      targetCid: battleCampaignTargetCid(),
      heroId: Number.parseInt(battleCampaignHeroId.value, 10) || 0,
      soldiers: battleCampaignSelectedSoldiers.value,
      useFlag: battleCampaignUseFlag.value
    });
    if (seq !== battleCampaignPreviewSeq) return;
    battleCampaignPreviewData.value = preview;
    battleCampaignPathNeedTime.value = formatDuration(preview.pathTime);
    battleCampaignArriveTime.value = preview.arrival ? formatBattleClock(preview.arrival) : "0:00";
    battleCampaignFieldName.value = preview.target || preview.fieldName || battleFieldName();
    battleCampaignStartReason.value = "";
  } catch (error) {
    if (seq !== battleCampaignPreviewSeq) return;
    battleCampaignStartReason.value = error instanceof Error ? error.message : "战场出征预览读取失败。";
  }
}

async function startBattleCampaignDispatch() {
  if (!city.value) return;
  const disabledReason = battleCampaignStartDisabledReason.value;
  if (disabledReason) {
    battleCampaignStartReason.value = disabledReason;
    return;
  }
  const soldiers = battleCampaignSelectedSoldiers.value;
  const firstSoldier = soldiers[0];
  if (!firstSoldier) return;
  battleCampaignStartReason.value = "";
  await dispatchCityTroop(city.value.summary.cid, {
    targetCid: battleCampaignTargetCid(),
    soldierSid: firstSoldier.sid,
    soldierCount: firstSoldier.count,
    soldiers,
    heroId: Number.parseInt(battleCampaignHeroId.value, 10),
    task: battleCampaignTask.value
  });
  closeBattleCampaignDialog();
}

function syncBattleCampaignRoster() {
  const troops = cityTroopItems.value;
  const existingCounts = new Map(battleCampaignSoldiers.value.map((soldier) => [soldier.sid, soldier.takecount]));
  battleCampaignSoldiers.value = troops.map((troop) => ({
    sid: troop.tid,
    name: troop.name,
    count: Math.max(0, Number(troop.count) || 0),
    takecount: Math.max(0, Math.min(Math.max(0, Number(troop.count) || 0), existingCounts.get(troop.tid) ?? 0))
  }));
  const heroes = heroRosterItems.value;
  battleCampaignHeroes.value = [
    { id: "0", heroname: "请选择" },
    ...heroes.map((hero) => ({ id: String(hero.hid), heroname: hero.name }))
  ];
  void refreshBattleCampaignPreview();
}

function setBattleSoldierCount(index: number, delta: number) {
  const item = battleCampaignSoldiers.value[index];
  if (!item) return;
  item.takecount = Math.max(0, Math.min(item.count, item.takecount + delta));
  battleCampaignStartReason.value = "";
  void refreshBattleCampaignPreview();
}

function onBattleSoldierInput(index: number, event: Event) {
  const item = battleCampaignSoldiers.value[index];
  if (!item) return;
  const input = event.target as HTMLInputElement;
  const parsed = Number.parseInt(input.value, 10);
  item.takecount = Number.isNaN(parsed) ? 0 : Math.max(0, Math.min(item.count, parsed));
  input.value = String(item.takecount);
  battleCampaignStartReason.value = "";
  void refreshBattleCampaignPreview();
}

function battleSoldierIcon(sid: number) {
  return asset(`images/army_${sid}.png`);
}

function battleHeroImage(troop: BattleFieldCurrentTroop) {
  const sex = troop.sex === 0 ? "girl" : "boy";
  const face = Math.max(1, Number(troop.face || 10));
  return asset(`images/herox/hero_${sex}_${face}.jpg`);
}

function leftHeroItemStyle(index: number) {
  return {
    left: `${(index & 1) === 0 ? 8 : 125}px`,
    top: `${Math.floor(index / 2) * 47 + 8}px`
  };
}

function leftWarItemStyle(index: number) {
  return {
    left: `${(index & 1) === 0 ? 6 : 124}px`,
    top: `${Math.floor(index / 2) * 40 + 8}px`
  };
}

function leftHeroImage(hero: Hero) {
  const sex = flashObjectNumber(hero, ["sex", "herosex"], 1) === 0 ? "girl" : "boy";
  const face = flashObjectNumber(hero, ["face", "heroface"], 10);
  return asset(`images/herox/hero_${sex}_${face}.jpg`);
}

function leftHeroState(hero: Hero) {
  return flashObjectText(hero, ["statename", "stateLabel", "stateName"], hero.cityName ? "驻守" : "空闲");
}

async function setLeftInfoTab(tab: LeftInfoTab) {
  leftInfoTab.value = tab;
  if (tab === "commander" && city.value && !heroRoster.value) {
    heroRoster.value = await cityHeroes(city.value.summary.cid, 100);
  } else if (tab === "army" && !troopsData.value) {
    troopsData.value = await myTroops(100);
  }
}

async function onLeftPanelTimer() {
  if (screen.value !== "city" || !city.value || loading.value) return;
  scrollActivityContent();
  --leftPanelTimerCounter15;
  if (leftPanelTimerCounter15 < 0) {
    leftPanelTimerCounter15 = 15;
    try {
      setCityDetail(await cityDetail(city.value.summary.cid));
    } catch {
      // Keep the visible city state if a timer refresh fails.
    }
  }
  --leftPanelTimerCounter60;
  if (leftPanelTimerCounter60 < 0) {
    leftPanelTimerCounter60 = 60;
    try {
      if (leftInfoTab.value === "commander") {
        heroRoster.value = await cityHeroes(city.value.summary.cid, 100);
      } else if (leftInfoTab.value === "army") {
        troopsData.value = await myTroops(100);
      }
    } catch {
      // Timer refresh is best-effort, matching Flash's background city commands.
    }
  }
}

function flashActivityInterval(value: unknown) {
  const parsed = Number.parseInt(String(value ?? ""), 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : 1800;
}

function scrollActivityContent() {
  const items = activityItems.value;
  if (items.length === 0) return;
  --activityScrollTimeLeft;
  if (activityScrollTimeLeft > 0) return;
  activityCurrentIndex.value = (activityCurrentIndex.value + 1) % items.length;
  const lastVisibleIndex = (activityCurrentIndex.value + Math.min(6, items.length) - 1) % items.length;
  activityScrollTimeLeft = flashActivityInterval(items[lastVisibleIndex]?.interval);
}

function takeAllBattleSoldiers() {
  for (const item of battleCampaignSoldiers.value) {
    item.takecount = item.count;
  }
  battleCampaignStartReason.value = "";
  void refreshBattleCampaignPreview();
}

function takeNoBattleSoldiers() {
  for (const item of battleCampaignSoldiers.value) {
    item.takecount = 0;
  }
  battleCampaignStartReason.value = "";
  void refreshBattleCampaignPreview();
}

function clampWorldMapCoord(value: number, min: number, max: number) {
  return Math.min(Math.max(Math.trunc(value), min), max);
}

function currentWorldMapCenter() {
  return worldMapCenter.value ?? worldMapData.value?.center ?? {
    x: city.value?.summary.x ?? 250,
    y: city.value?.summary.y ?? 250
  };
}

function setWorldMapCenter(center: { x: number; y: number }) {
  const next = {
    x: clampWorldMapCoord(center.x, WORLD_MAP_MIN_X, WORLD_MAP_MAX_X),
    y: clampWorldMapCoord(center.y, WORLD_MAP_MIN_Y, WORLD_MAP_MAX_Y)
  };
  worldMapCenter.value = next;
  worldMapInputX.value = String(next.x);
  worldMapInputY.value = String(next.y);
  hideWorldGridTip();
}

function moveWorldMap(dx: number, dy: number) {
  const center = currentWorldMapCenter();
  setWorldMapCenter({
    x: center.x + dx,
    y: center.y + dy
  });
}

let lastWorldMapPointerMove = { vector: "", at: 0 };

function moveWorldMapByPointer(dx: number, dy: number) {
  lastWorldMapPointerMove = { vector: `${dx},${dy}`, at: Date.now() };
  moveWorldMap(dx, dy);
}

function moveWorldMapByClick(dx: number, dy: number) {
  const vector = `${dx},${dy}`;
  if (lastWorldMapPointerMove.vector === vector && Date.now() - lastWorldMapPointerMove.at < 350) return;
  moveWorldMap(dx, dy);
}

function resetWorldMapToCity() {
  if (!city.value) return;
  setWorldMapCenter({
    x: city.value.summary.x,
    y: city.value.summary.y
  });
}

function submitWorldMapMove() {
  const x = Number.parseInt(worldMapInputX.value, 10);
  const y = Number.parseInt(worldMapInputY.value, 10);
  if (!Number.isFinite(x) || !Number.isFinite(y)) {
    setWorldMapCenter(currentWorldMapCenter());
    return;
  }
  setWorldMapCenter({ x, y });
}

function handleWorldMiniMapClick(event: MouseEvent) {
  const target = event.currentTarget;
  if (!(target instanceof HTMLElement)) return;
  const rect = target.getBoundingClientRect();
  setWorldMapCenter({
    x: Math.floor(event.clientX - rect.left) + 1,
    y: Math.floor(event.clientY - rect.top) + 1
  });
}

function worldGridFromPanelPoint(localX: number, localY: number) {
  const center = currentWorldMapCenter();
  const localTileX = localX - WORLD_GRID_CENTER_X - 107 / 2;
  const localTileY = localY - WORLD_GRID_CENTER_Y - WORLD_GRID_IMAGE_HEIGHT / 2;
  const dx = Math.round((localTileX / WORLD_GRID_OFFSET_X + localTileY / WORLD_GRID_OFFSET_Y) / 2);
  const dy = Math.round((localTileY / WORLD_GRID_OFFSET_Y - localTileX / WORLD_GRID_OFFSET_X) / 2);
  const x = center.x + dx;
  const y = center.y + dy;
  if (x < WORLD_MAP_MIN_X || x > WORLD_MAP_MAX_X || y < WORLD_MAP_MIN_Y || y > WORLD_MAP_MAX_Y) return null;
  return { x, y };
}

function worldGridInfo(x: number, y: number) {
  const cityItem = worldMapCities.value.find((item) => {
    const point = flashCityPoint(item);
    return point.x === x && point.y === y;
  });
  if (cityItem) {
    return {
      title: `${cityItem.name}:(${x},${y})`,
      text: `君主:${cityItem.owner || "-"}\n声望:${cityItem.level * 1000}\n联盟:-\n状态:${WORLD_STATE_NAMES[0]}`,
      city: true,
      empty: false,
      targetCid: cityItem.cid
    };
  }

  const type = worldTerrainType(x, y);
  if (type === 0) {
    return {
      title: `空地:(${x},${y})`,
      text: "这是一块尚未开发的空地。",
      city: false,
      empty: true,
      targetCid: 0
    };
  }

  const level = worldTerrainLevel(x, y);
  const tile = worldMapData.value?.tiles.find((item) => item.x === x && item.y === y);
  return {
    title: `${WORLD_TERRAIN_NAMES[type] ?? "野地"}:(${x},${y})  ${level}级`,
    text: worldFieldResourceText(type, level),
    city: false,
    empty: false,
    targetCid: tile?.cid || x + y * 1000
  };
}

function handleWorldGridMove(event: MouseEvent) {
  const target = event.currentTarget;
  if (!(target instanceof HTMLElement)) return;
  const rect = target.getBoundingClientRect();
  const localX = event.clientX - rect.left;
  const localY = event.clientY - rect.top;
  const grid = worldGridFromPanelPoint(localX, localY);
  if (!grid) {
    hideWorldGridTip();
    return;
  }
  const info = worldGridInfo(grid.x, grid.y);
  const width = info.city ? 240 : 200;
  const height = info.city ? 100 : 85;
  let x = localX + 20;
  let y = localY;
  if (x + width >= CITY_MAP_WIDTH) x = localX - width - 10;
  if (y + height >= CITY_MAP_HEIGHT) y = localY - height - 10;
  worldGridTip.value = {
    visible: true,
    x: Math.max(0, Math.round(x)),
    y: Math.max(0, Math.round(y)),
    ...info
  };
}

function handleWorldGridClick(event: MouseEvent) {
  const target = event.currentTarget;
  if (!(target instanceof HTMLElement)) return;
  const rect = target.getBoundingClientRect();
  const grid = worldGridFromPanelPoint(event.clientX - rect.left, event.clientY - rect.top);
  if (!grid) {
    selectedWorldGrid.value = null;
    hideWorldGridTip();
    return;
  }
  selectWorldGrid(grid.x, grid.y);
}

function selectWorldGrid(x: number, y: number) {
  const info = worldGridInfo(x, y);
  if (info.empty) {
    selectedWorldGrid.value = null;
    worldMapAlert.value = "空地不能执行操作。";
    hideWorldGridTip();
    return;
  }
  worldMapAlert.value = "";
  selectedWorldGrid.value = { x, y, ...info };
  hideWorldGridTip();
}

function hideWorldGridTip() {
  worldGridTip.value = { ...worldGridTip.value, visible: false };
}

function selectedWorldGridMarkerStyle() {
  if (!selectedWorldGrid.value) return {};
  const center = currentWorldMapCenter();
  return {
    left: `${WORLD_GRID_CENTER_X + (selectedWorldGrid.value.x - center.x) * WORLD_GRID_OFFSET_X - (selectedWorldGrid.value.y - center.y) * WORLD_GRID_OFFSET_X + 54}px`,
    top: `${WORLD_GRID_CENTER_Y + (selectedWorldGrid.value.x - center.x) * WORLD_GRID_OFFSET_Y + (selectedWorldGrid.value.y - center.y) * WORLD_GRID_OFFSET_Y + 36}px`
  };
}

function worldGridActionStyle() {
  if (!selectedWorldGrid.value) return {};
  return selectedWorldGrid.value.x >= currentWorldMapCenter().x
    ? { left: "210px", top: "82px" }
    : { left: "430px", top: "82px" };
}

function closeWorldGridAction() {
  selectedWorldGrid.value = null;
}

function closeWorldMapAlert() {
  worldMapAlert.value = "";
}

function inspectWorldTarget() {
  if (!selectedWorldGrid.value) return;
  worldMapAlert.value = `${selectedWorldGrid.value.title}\n${selectedWorldGrid.value.text}`;
}

async function scoutWorldTarget() {
  if (!selectedWorldGrid.value) return;
  await openWorldCampaignDialog(2);
}

async function openWorldCampaignDialog(task: 2 | 3 = 3) {
  if (!selectedWorldGrid.value) return;
  worldMapAlert.value = "";
  battlePanelVisible.value = true;
  battleMenuVisible.value = false;
  battleCampaignMode.value = "world";
  resetBattleCampaignDraft();
  battleCampaignTask.value = task;
  battleCampaignFieldName.value = selectedWorldGrid.value.title;
  battleCampaignTargets.value = [{
    id: String(worldTargetId()),
    name: selectedWorldGrid.value.title
  }];
  battleCampaignTargetId.value = battleCampaignTargets.value[0].id;
  battleCampaignVisible.value = true;
  await withLoading(async () => {
    await refreshBattleCampaignRoster();
  });
}

function selectWorldCity(cityItem: WorldMap["cities"][number]) {
  const point = flashCityPoint(cityItem);
  selectWorldGrid(point.x, point.y);
}

function flashCityPoint(cityItem: WorldMap["cities"][number]) {
  const mapX = Math.floor(Number(cityItem.x));
  const mapY = Math.floor(Number(cityItem.y));
  if (Number.isFinite(mapX) && Number.isFinite(mapY) && mapX >= 0 && mapY >= 0) {
    return { x: mapX, y: mapY };
  }

  const cid = Math.max(0, Math.floor(Number(cityItem.cid)));
  const cidX = cid % 1000;
  const cidY = Math.floor(cid / 1000);
  return {
    x: Number.isFinite(cidX) ? cidX : 0,
    y: Number.isFinite(cidY) ? cidY : 0
  };
}

function worldTerrainImage(x: number, y: number, cityLevel?: number) {
  if (cityLevel !== undefined) {
    const cityIndex = Math.min(4, Math.max(0, Math.floor(cityLevel / 3)));
    return `world_tile_${cityIndex === 0 ? "city" : `city${cityIndex}`}.png`;
  }
  const index = worldTerrainType(x, y);
  return `world_tile_${WORLD_TERRAIN_IMAGES[index]}.png`;
}

function worldTerrainType(x: number, y: number) {
  return Math.abs((x * 11 + y * 7 + x * y) % WORLD_TERRAIN_IMAGES.length);
}

function worldTerrainLevel(x: number, y: number) {
  return Math.max(1, Math.abs((x * 5 + y * 3) % 10));
}

function worldFieldResourceText(type: number, level: number) {
  const bonus = level * 2 + (level > 0 ? 3 : 0);
  if (type === 1) return "可用于建立新城。";
  if (type === 2) return `木材产量增加${bonus}%`;
  if (type === 3) return `粮食产量增加${bonus}%`;
  if (type === 4) return `石料产量增加${level + 2}%`;
  if (type === 5) return `铁锭产量增加${bonus}%`;
  return `资源产量增加${bonus}%`;
}

function worldMapCityDotStyle(cityItem: WorldMap["cities"][number]) {
  const point = flashCityPoint(cityItem);
  return {
    left: `${point.x + 9}px`,
    top: `${point.y + 9}px`
  };
}

function worldMapCityLabelStyle(cityItem: WorldMap["cities"][number]) {
  const point = flashCityPoint(cityItem);
  const left = Math.min(Math.max(point.x + 13, 0), 452);
  const top = Math.min(Math.max(point.y + 7, 0), 498);
  return {
    left: `${left}px`,
    top: `${top}px`,
    zIndex: cityItem.name === city.value?.summary.name ? "80" : "20"
  };
}

function buildingImage(building: Building) {
  return asset(buildingImageByBid[building.bid] ?? "CityInnerPanel_house.png");
}

function outerBuildingImage(building: Building) {
  const imageByBid: Record<number, string> = {
    1: "CityPanel_farm.png",
    2: "CityPanel_logcamp.png",
    3: "CityPanel_stonepit.png",
    4: "CityPanel_mine.png"
  };
  return asset(imageByBid[building.bid] ?? "mycity_field.png");
}

function buildingDialogImage(building: Building) {
  return asset(buildingIntroByBid[building.bid] ?? buildingImageByBid[building.bid] ?? "CityInnerPanel_house.png");
}

function positionBuilding(position: number) {
  return occupiedByPosition.value.get(position);
}

function isOccupied(position: number) {
  return occupiedByPosition.value.has(position);
}

function isBusy(building: Building | undefined) {
  return !!building && building.state !== 0;
}

function buildingLevelLabel(building: Building | undefined) {
  if (!building) return "";
  return String(displayBuildingLevel(building));
}

function buildingLevelText(building: Building | undefined) {
  const label = buildingLevelLabel(building);
  return label ? `Lv.${label}` : "";
}

function buildingLevelImage(building: Building | undefined) {
  if (!building) return "";
  return asset(`level_${buildingLevelLabel(building)}.png`);
}

function displayBuildingLevel(building: Building) {
  if (building.state === 1 && building.level === 0) return 1;
  return building.level;
}

function setBuildingTip(event: MouseEvent, text: string, offsetX = 0) {
  const cityViewElement = (event.currentTarget as HTMLElement).closest(".city-view") as HTMLElement | null;
  if (!cityViewElement) return;
  const rect = cityViewElement.getBoundingClientRect();
  const scaleX = rect.width / CITY_MAP_WIDTH;
  const scaleY = rect.height / CITY_MAP_HEIGHT;
  const x = Math.round((event.clientX - rect.left) / scaleX + offsetX);
  const y = Math.round((event.clientY - rect.top) / scaleY + 20);
  buildingTip.value = {
    visible: true,
    x: Math.min(600, Math.max(0, x)),
    y: Math.min(530, Math.max(0, y)),
    text
  };
}

function localCityPoint(event: MouseEvent) {
  const cityViewElement = (event.currentTarget as HTMLElement).closest(".city-view") as HTMLElement | null;
  if (!cityViewElement) return null;
  const rect = cityViewElement.getBoundingClientRect();
  const scaleX = rect.width / CITY_MAP_WIDTH;
  const scaleY = rect.height / CITY_MAP_HEIGHT;
  return {
    x: (event.clientX - rect.left) / scaleX,
    y: (event.clientY - rect.top) / scaleY
  };
}

function innerPlotFromPoint(x: number, y: number) {
  const point = innerGridPoint(x, y);
  if (isInnerWallGrid(point.gridX, point.gridY)) return cityWallPlot;
  if (point.gridX < 0 || point.gridX >= 6 || point.gridY < 0 || point.gridY >= 6) return null;
  if (isInnerCityHallGrid(point.gridX, point.gridY)) return cityHallPlot;
  return innerPlots.find((plot) => plot.gridX === point.gridX && plot.gridY === point.gridY) ?? null;
}

function isAnnouncePoint(x: number, y: number) {
  return x >= 653 && x <= 697 && y >= 411 && y <= 453;
}

function setBuildingTipAtPoint(x: number, y: number, text: string, offsetX = 0) {
  buildingTip.value = {
    visible: true,
    x: Math.min(600, Math.max(0, Math.round(x + offsetX))),
    y: Math.min(530, Math.max(0, Math.round(y + 20))),
    text
  };
}

function handleInnerHitMove(event: MouseEvent) {
  const point = localCityPoint(event);
  if (!point) return;
  const plot = innerPlotFromPoint(point.x, point.y);
  hoveredInnerPlot.value = plot;
  hoveredAnnounce.value = false;
  if (!plot) {
    if (isAnnouncePoint(point.x, point.y)) {
      hoveredAnnounce.value = true;
      setBuildingTipAtPoint(point.x, point.y, "公告栏", -70);
      return;
    }
    hideBuildingTip();
    return;
  }
  const text = plot.position === cityWallPlot.position ? wallBuildingTipText() : buildingTipTextForPlot(plot);
  setBuildingTipAtPoint(point.x, point.y, text);
}

async function handleInnerHitClick(event: MouseEvent) {
  const point = localCityPoint(event);
  if (!point) return;
  const plot = innerPlotFromPoint(point.x, point.y);
  if (!plot) {
    if (isAnnouncePoint(point.x, point.y)) {
      announcementVisible.value = true;
      hideBuildingTip();
    }
    return;
  }
  hideBuildingTip();
  await openBuild(plot);
}

function clearInnerHit() {
  hoveredInnerPlot.value = null;
  hoveredAnnounce.value = false;
  hideBuildingTip();
}

function hideBuildingTip() {
  buildingTip.value = { ...buildingTip.value, visible: false };
}

function buildingTipTextForPlot(plot: Plot) {
  const building = positionBuilding(plot.position);
  if (building) return `${building.name} ${displayBuildingLevel(building)}级`;
  return "空地";
}

function wallBuildingTipText() {
  const wall = positionBuilding(cityWallPlot.position);
  return wall ? `${wall.name} ${displayBuildingLevel(wall)}级` : "无城墙";
}

function speedGoodsImage(item: SpeedGoods) {
  return asset(`item_${item.gid}.png`);
}

function useGoodsImage(item: UserTypeGoodsItem) {
  return asset(`item_${item.gid}.png`);
}

function speedGoodsEffect(item: SpeedGoods) {
  if (item.gid === 73) return `立即完成，消耗 ${item.cost} 元宝`;
  if (item.gid === 72) return "缩短当前剩余时间30%";
  return `缩短 ${formatDuration(item.reduceTime)}`;
}

function buildingIntro(option: BuildingOption) {
  return asset(buildingIntroByBid[option.bid] ?? buildingImageByBid[option.bid] ?? "CityInnerPanel_house.png");
}

function buildOptionDescription(option: BuildingOption) {
  const cost = [
    `当前人口 ${formatNumber(resources.value?.people)}`,
    `消耗粮食 ${formatNumber(option.food)}`,
    `消耗木材 ${formatNumber(option.wood)}`,
    `消耗石料 ${formatNumber(option.rock)}`,
    `消耗铁锭 ${formatNumber(option.iron)}`,
    `消耗黄金 ${formatNumber(option.gold)}`
  ].join(" ");
  return option.description ? `${option.description} ${cost}` : cost;
}

function selectBuildOption(option: BuildingOption) {
  selectedBid.value = option.bid;
  if (effectiveGuideVisible.value && currentGuide.value?.gid === 7 && option.bid === 5 && option.canBuild) {
    void confirmBuild(option);
  }
}

function prevFace() {
  roleForm.value.face = roleForm.value.face <= 1 ? 10 : roleForm.value.face - 1;
}

function nextFace() {
  roleForm.value.face = roleForm.value.face >= 10 ? 1 : roleForm.value.face + 1;
}

async function getProvincePixels(province: (typeof provinceAssets)[number]) {
  const cached = provincePixelCache.get(province.id);
  if (cached) return cached;
  const image = new Image();
  image.src = asset(province.image);
  await image.decode();
  const canvas = document.createElement("canvas");
  canvas.width = Math.round(province.w);
  canvas.height = Math.round(province.h);
  const context = canvas.getContext("2d", { willReadFrequently: true });
  if (!context) return null;
  context.drawImage(image, 0, 0, canvas.width, canvas.height);
  const pixels = context.getImageData(0, 0, canvas.width, canvas.height);
  provincePixelCache.set(province.id, pixels);
  return pixels;
}

async function provinceAtEvent(event: MouseEvent) {
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect();
  const mapX = event.clientX - rect.left;
  const mapY = event.clientY - rect.top;
  for (const province of provinceAssets) {
    const localX = Math.floor(mapX - province.x);
    const localY = Math.floor(mapY - province.y);
    if (localX < 0 || localY < 0 || localX >= province.w || localY >= province.h) continue;
    const pixels = await getProvincePixels(province);
    if (!pixels) continue;
    const index = (localY * pixels.width + localX) * 4;
    if (pixels.data[index] || pixels.data[index + 1] || pixels.data[index + 2]) return province.id;
  }
  return -1;
}

async function moveProvince(event: MouseEvent) {
  const province = await provinceAtEvent(event);
  hoveredProvince.value = province > 0 ? province : null;
}

async function pickProvince(event: MouseEvent) {
  const province = await provinceAtEvent(event);
  if (province > 0) selectedProvince.value = province;
}

onMounted(() => {
  updateStageScale();
  window.addEventListener("resize", updateStageScale);
  window.addEventListener("codex:battle-troop-detail-preview", handleBattleTroopDetailPreview);
  window.addEventListener("codex:battle-attack-preview", handleBattleAttackPreview);
  timerID = window.setInterval(() => {
    currentUnix.value = Math.floor(Date.now() / 1000);
    void onLeftPanelTimer();
  }, 1000);
  document.addEventListener("mousedown", closeCityDropdown);
  void boot();
});

onUnmounted(() => {
  if (timerID) window.clearInterval(timerID);
  window.removeEventListener("resize", updateStageScale);
  window.removeEventListener("codex:battle-troop-detail-preview", handleBattleTroopDetailPreview);
  window.removeEventListener("codex:battle-attack-preview", handleBattleAttackPreview);
  document.removeEventListener("mousedown", closeCityDropdown);
});
</script>

<template>
  <main class="page">
    <section class="stage" :class="`screen-${screen}`" :style="stageStyle">
      <div v-if="screen === 'boot'" class="boot">载入中...</div>

      <template v-else-if="screen === 'login'">
        <img class="full-bg" :src="asset('board_login.jpg')" alt="" />
        <form class="login-form" @submit.prevent="submitLogin">
          <label class="login-input account">
            <span class="sr-only">账号</span>
            <input v-model="loginForm.passport" maxlength="24" autocomplete="username" />
          </label>
          <label class="login-input password">
            <span class="sr-only">密码</span>
            <input v-model="loginForm.password" type="password" autocomplete="current-password" />
          </label>
          <button
            class="login-submit"
            type="submit"
            :disabled="loading"
            :title="loading ? '正在进入游戏，请稍候。' : undefined"
            :aria-describedby="loading ? 'login-loading-note' : undefined"
          >进入游戏</button>
          <span id="login-loading-note" class="sr-only">正在进入游戏，请稍候。</span>
        </form>
        <section class="login-snapshot" aria-label="服务器快照">
          <div class="login-snapshot-title">服务器快照</div>
          <div v-if="loginSnapshotMessage" class="login-snapshot-message">{{ loginSnapshotMessage }}</div>
          <template v-else>
            <div class="login-commander-list">
              <button
                v-for="commander in loginCommanders"
                :key="commander.uid"
                class="login-commander-row"
                :class="{ selected: loginForm.passport === commander.passport }"
                type="button"
                @click="loginForm.passport = commander.passport"
              >
                <strong>{{ commander.name }}</strong>
                <span>城 {{ commander.cityCount }}</span>
                <em>{{ commander.passport }}</em>
              </button>
              <div v-if="loginCommanders.length === 0" class="login-commander-empty">暂无可用君主</div>
            </div>
            <div class="login-dashboard-grid">
              <span class="login-dashboard-row"><b>君主</b>{{ loginDashboard?.counts.users ?? 0 }}</span>
              <span class="login-dashboard-row"><b>城池</b>{{ loginDashboard?.counts.cities ?? 0 }}</span>
              <span class="login-dashboard-row"><b>野地</b>{{ loginDashboard?.counts.worldTiles ?? 0 }}</span>
              <span class="login-dashboard-row"><b>部队</b>{{ loginDashboard?.counts.activeTroops ?? 0 }}</span>
              <div v-if="!loginDashboard" class="login-dashboard-empty">暂无概览数据</div>
            </div>
            <div class="login-featured-cities">
              <span
                v-for="item in loginDashboard?.featuredCities ?? []"
                :key="item.cid"
                class="login-featured-city"
              >
                {{ item.name }}({{ item.x }},{{ item.y }})
              </span>
              <span v-if="(loginDashboard?.featuredCities ?? []).length === 0" class="login-featured-empty">暂无推荐城池</span>
            </div>
          </template>
        </section>
      </template>

      <template v-else-if="screen === 'create-role'">
        <img class="full-bg dimmed-bg" :src="asset('board_login2.jpg')" alt="" />
        <img class="role-title-img" :src="asset('title.png')" alt="" />
        <div class="role-title">创建角色</div>

        <section class="role-left">
          <div class="face-title">
            <img class="face-select-img" :src="asset('select_headimg.png')" alt="" />
            <span>选择头像</span>
          </div>
          <div class="role-field province-select">
            <span>所在州郡</span>
            <span class="role-select-shell">
              <select v-model.number="selectedProvince">
                <option v-for="province in provinceAssets" :key="province.id" :value="province.id">
                  {{ province.name }}
                </option>
              </select>
            </span>
          </div>

          <div class="face-frame">
            <img class="face-photo" :src="faceImage" alt="" />
            <button class="face-prev" type="button" @click="prevFace">&lt;</button>
            <button class="face-next" type="button" @click="nextFace">&gt;</button>
            <div class="sex-row">
              <button type="button" :class="{ active: roleForm.sex === 1 }" @click="roleForm.sex = 1">男</button>
              <button type="button" :class="{ active: roleForm.sex === 0 }" @click="roleForm.sex = 0">女</button>
            </div>
          </div>

          <label class="role-field name-field">
            <span>君主名</span>
            <input v-model="roleForm.userName" maxlength="8" />
          </label>
          <label class="role-field city-name-field">
            <span>城池名</span>
            <input v-model="roleForm.cityName" maxlength="8" />
          </label>
          <div class="province-info">{{ selectedProvinceName }}：请选择主公起兵之地。</div>
          <label class="role-rule">
            <input type="checkbox" checked />
            <span class="role-rule-check" aria-hidden="true"></span>
            <span>我已经阅读并同意用户协议</span>
          </label>
          <button
            class="start-role"
            type="button"
            :disabled="loading"
            :title="loading ? '正在创建角色，请稍候。' : undefined"
            :aria-describedby="loading ? 'create-role-loading-note' : undefined"
            @click="submitRole"
          >开始游戏</button>
          <span id="create-role-loading-note" class="sr-only">正在创建角色，请稍候。</span>
        </section>

        <section class="role-map" @mousemove="moveProvince" @click="pickProvince" @mouseleave="hoveredProvince = null">
          <img class="china-map" :src="asset('sanguo_worldmap.png')" alt="" />
          <img
            v-for="province in provinceAssets"
            :key="province.id"
            v-show="activeProvince === province.id"
            class="province-piece"
            :src="asset(province.image)"
            :style="{ left: `${province.x}px`, top: `${province.y}px`, width: `${province.w}px`, height: `${province.h}px` }"
            alt=""
          />
        </section>
      </template>

      <template v-else-if="screen === 'city' && city">
        <img class="full-bg city-bg" :src="asset('board_login2.jpg')" alt="" />
        <img class="left-board" :src="asset('leftboard_new.jpg')" alt="" />
        <img class="top-board" :src="asset('board_topbar.png')" alt="" />
        <div class="top-tabs">
          <button class="top-tab-btn city-tab" data-testid="top-scene-tab-inner" data-action="open-city-scene" data-scene="inner" aria-label="城内" title="城内" type="button" :aria-pressed="cityView === 'inner' && !worldMapPanelVisible && !battlePanelVisible" @mouseenter="hoveredTopButton = 'innercity'" @mouseleave="hoveredTopButton = ''; pressedTopButton = ''" @pointerdown="pressedTopButton = 'innercity'" @mousedown="pressedTopButton = 'innercity'" @click="openCityScene('inner')" @pointerup="pressedTopButton = ''" @mouseup="pressedTopButton = ''" @keydown.enter.prevent="openCityScene('inner')" @keydown.space.prevent="openCityScene('inner')">
            <img :src="topButtonImage('innercity', cityView === 'inner' && !worldMapPanelVisible && !battlePanelVisible)" alt="城内" />
          </button>
          <button class="top-tab-btn city-tab" data-testid="top-scene-tab-outer" data-action="open-city-scene" data-scene="outer" aria-label="城池" title="城池" type="button" :aria-pressed="cityView === 'outer' && !worldMapPanelVisible && !battlePanelVisible" @mouseenter="hoveredTopButton = 'outercity'" @mouseleave="hoveredTopButton = ''; pressedTopButton = ''" @pointerdown="pressedTopButton = 'outercity'" @mousedown="pressedTopButton = 'outercity'" @click="openCityScene('outer')" @pointerup="pressedTopButton = ''" @mouseup="pressedTopButton = ''" @keydown.enter.prevent="openCityScene('outer')" @keydown.space.prevent="openCityScene('outer')">
            <img :src="topButtonImage('outercity', cityView === 'outer' && !worldMapPanelVisible && !battlePanelVisible)" alt="城池" />
          </button>
          <button class="top-tab-btn city-tab" data-testid="top-scene-tab-map" data-action="open-world-map" data-scene="map" aria-label="地图" title="地图" type="button" :aria-pressed="worldMapPanelVisible" @mouseenter="hoveredTopButton = 'map'" @mouseleave="hoveredTopButton = ''; pressedTopButton = ''" @pointerdown="pressedTopButton = 'map'" @mousedown="pressedTopButton = 'map'" @click="openWorldMapPanel()" @pointerup="pressedTopButton = ''" @mouseup="pressedTopButton = ''" @keydown.enter.prevent="openWorldMapPanel()" @keydown.space.prevent="openWorldMapPanel()">
            <img :src="topButtonImage('map', worldMapPanelVisible)" alt="地图" />
          </button>
          <button class="top-tab-btn city-tab" data-testid="top-scene-tab-battle" data-action="open-battle" data-scene="battle" aria-label="战场" title="战场" type="button" :aria-pressed="battlePanelVisible" @mouseenter="hoveredTopButton = 'battle'" @mouseleave="hoveredTopButton = ''; pressedTopButton = ''" @pointerdown="pressedTopButton = 'battle'" @mousedown="pressedTopButton = 'battle'" @click="openBattlePanel()" @pointerup="pressedTopButton = ''" @mouseup="pressedTopButton = ''" @keydown.enter.prevent="openBattlePanel()" @keydown.space.prevent="openBattlePanel()">
            <img :src="topButtonImage('battle', battlePanelVisible)" alt="战场" />
          </button>
        </div>
        <div class="top-mini-functions">
          <button
            class="mini-function-btn build"
            :class="{ selected: showTopBuildingList }"
            data-testid="top-mini-build"
            data-action="toggle-building-queue"
            aria-label="建筑队列"
            :aria-pressed="showTopBuildingList"
            title="建筑队列"
            type="button"
            @click="toggleTopBuildingList"
          ></button>
          <button class="mini-function-btn craft" data-testid="top-mini-craft" data-action="open-craft-dialog" aria-label="工匠" title="工匠" type="button" @click="openCraftDialog"></button>
          <button class="mini-function-btn tactic" data-testid="top-mini-tactic" data-action="open-tactic-dialog" aria-label="计谋" title="计谋" type="button" @click="openTacticDialog"></button>
          <button
            class="mini-function-btn level"
            :class="{ selected: showBuildingLevels }"
            data-testid="top-mini-level"
            data-action="toggle-building-levels"
            aria-label="显示建筑等级"
            :aria-pressed="showBuildingLevels"
            title="显示建筑等级"
            type="button"
            @click="showBuildingLevels = !showBuildingLevels"
          ></button>
        </div>
        <div v-if="showTopBuildingList" class="top-building-list-panel" data-testid="top-building-list">
          <template v-if="topBuildingQueueItems.length > 0">
            <button
              v-for="item in topBuildingQueueItems"
              :key="`top-building-${item.position}-${item.state}`"
              class="top-building-list-row"
              data-testid="top-building-list-row"
              data-action="open-building-queue-item"
              :data-position="item.position"
              type="button"
              @click="openTopBuildingQueueItem(item)"
            >
              <span class="top-building-name">{{ item.name }}</span>
              <span class="top-building-task">{{ item.task }} Lv.{{ item.level }}-&gt;{{ item.nextLevel }}</span>
              <span class="top-building-time">{{ item.timeLeft }}</span>
            </button>
          </template>
          <div v-else class="top-building-empty">暂无建筑队列</div>
        </div>
        <div v-if="craftDialogVisible" class="modal-layer utility-overlay-layer">
          <div class="utility-mini-dialog" data-testid="craft-dialog">
            <button class="dialog-close" data-testid="craft-close" data-action="close-craft-dialog" type="button" @click="closeCraftDialog">关闭</button>
            <div class="utility-mini-title">工匠</div>
            <button data-testid="craft-action-building-queue" data-action="building-queue" type="button" @click="toggleTopBuildingList">建筑队列</button>
            <button data-testid="craft-action-building-overview" data-action="building-overview" type="button" @click="openCraftBuildingOverview">建筑总览</button>
            <button data-testid="craft-action-produce" data-action="produce" type="button" @click="openCraftProduceDialog">资源生产</button>
          </div>
        </div>
        <div v-if="tacticDialogVisible" class="modal-layer utility-overlay-layer">
          <div class="utility-mini-dialog" data-testid="tactic-dialog">
            <button class="dialog-close" data-testid="tactic-close" data-action="close-tactic-dialog" type="button" @click="closeTacticDialog">关闭</button>
            <div class="utility-mini-title">计谋</div>
            <button data-testid="tactic-action-world-map" data-action="open-world-map" type="button" @click="openTacticWorldMap">打开地图</button>
            <button data-testid="tactic-action-battle" data-action="open-battle" type="button" @click="openTacticBattle">打开战场</button>
          </div>
        </div>
        <div class="top-actions">
          <button class="top-tab-btn func-tab" data-testid="top-func-tab-hero" data-action="open-hero-panel" aria-label="将领" title="将领" type="button" :aria-pressed="heroPanelVisible" @mouseenter="hoveredTopButton = 'hero'" @mouseleave="hoveredTopButton = ''; pressedTopButton = ''" @pointerdown="pressedTopButton = 'hero'" @pointerup="pressedTopButton = ''" @click="openHeroPanel"><img :src="topButtonImage('hero', heroPanelVisible)" alt="将领" /></button>
          <button class="top-tab-btn func-tab" data-testid="top-func-tab-army" data-action="open-barracks-panel" aria-label="军队" title="军队" type="button" :aria-pressed="barracksPanelVisible" @mouseenter="hoveredTopButton = 'army'" @mouseleave="hoveredTopButton = ''; pressedTopButton = ''" @pointerdown="pressedTopButton = 'army'" @pointerup="pressedTopButton = ''" @click="openBarracksPanel"><img :src="topButtonImage('army', barracksPanelVisible)" alt="军队" /></button>
          <button class="top-tab-btn func-tab" data-testid="top-func-tab-union" data-action="open-union-panel" aria-label="联盟" title="联盟" type="button" :aria-pressed="unionPanelVisible" @mouseenter="hoveredTopButton = 'union'" @mouseleave="hoveredTopButton = ''; pressedTopButton = ''" @pointerdown="pressedTopButton = 'union'" @pointerup="pressedTopButton = ''" @click="openUnionPanel"><img :src="topButtonImage('union', unionPanelVisible)" alt="联盟" /></button>
          <button class="top-tab-btn func-tab" data-testid="top-func-tab-mission" data-action="open-task-panel" aria-label="任务" title="任务" type="button" :aria-pressed="taskPanelVisible" @mouseenter="hoveredTopButton = 'mission'" @mouseleave="hoveredTopButton = ''; pressedTopButton = ''" @pointerdown="pressedTopButton = 'mission'" @pointerup="pressedTopButton = ''" @click="openTaskPanel"><img :src="topButtonImage('mission', taskPanelVisible)" alt="任务" /></button>
        </div>
        <div class="left-data">
          <div class="left-bottom-panel" aria-hidden="true"></div>
          <div class="left-view-tabs">
            <button
              class="left-vert-btn"
              data-testid="left-tab-resource"
              data-action="select-left-tab"
              data-tab="resource"
              :class="{ selected: leftInfoTab === 'resource' }"
              :aria-pressed="leftInfoTab === 'resource'"
              type="button"
              @mousedown.prevent="setLeftInfoTab('resource')"
              @click="setLeftInfoTab('resource')"
              @keydown.enter.prevent="setLeftInfoTab('resource')"
              @keydown.space.prevent="setLeftInfoTab('resource')"
            >
              资源
            </button>
            <button
              class="left-vert-btn"
              data-testid="left-tab-commander"
              data-action="select-left-tab"
              data-tab="commander"
              :class="{ selected: leftInfoTab === 'commander' }"
              :aria-pressed="leftInfoTab === 'commander'"
              type="button"
              @mousedown.prevent="setLeftInfoTab('commander')"
              @click="setLeftInfoTab('commander')"
              @keydown.enter.prevent="setLeftInfoTab('commander')"
              @keydown.space.prevent="setLeftInfoTab('commander')"
            >
              将领
            </button>
            <button
              class="left-vert-btn"
              data-testid="left-tab-army"
              data-action="select-left-tab"
              data-tab="army"
              :class="{ selected: leftInfoTab === 'army' }"
              :aria-pressed="leftInfoTab === 'army'"
              type="button"
              @mousedown.prevent="setLeftInfoTab('army')"
              @click="setLeftInfoTab('army')"
              @keydown.enter.prevent="setLeftInfoTab('army')"
              @keydown.space.prevent="setLeftInfoTab('army')"
            >
              军队
            </button>
            <button
              class="left-vert-btn"
              data-testid="left-tab-defence"
              data-action="select-left-tab"
              data-tab="defence"
              :class="{ selected: leftInfoTab === 'defence' }"
              :aria-pressed="leftInfoTab === 'defence'"
              type="button"
              @mousedown.prevent="setLeftInfoTab('defence')"
              @click="setLeftInfoTab('defence')"
              @keydown.enter.prevent="setLeftInfoTab('defence')"
              @keydown.space.prevent="setLeftInfoTab('defence')"
            >
              城防
            </button>
          </div>
          <div class="left-rank-strip">声望 {{ userPrestige }}&nbsp;&nbsp;排名 {{ userRank }}</div>
          <div class="left-hero-panel">
            <div class="lord-portrait-frame">
              <img class="lord-portrait" data-testid="lord-portrait" role="button" tabindex="0" aria-label="君主信息" :src="faceImage" alt="" @click="openLordDialog" @keydown.enter.prevent="openLordDialog" @keydown.space.prevent="openLordDialog" />
            </div>
            <div class="lord-meta-panel">
              <div class="lord-meta-row row-king" data-testid="lord-meta-king"><span data-action="open-lord" role="button" tabindex="0" @click="openLordDialog" @keydown.enter.prevent="openLordDialog" @keydown.space.prevent="openLordDialog">君主:</span><strong>{{ user?.name || city.summary.owner }}</strong></div>
              <div class="lord-meta-row row-prestige" data-testid="lord-meta-prestige"><span data-action="open-rank" role="button" tabindex="0" @click="openRankPanel()" @keydown.enter.prevent="openRankPanel()" @keydown.space.prevent="openRankPanel()">声望:</span><strong>{{ userPrestige }}</strong></div>
              <div class="lord-meta-row row-rank" data-testid="lord-meta-rank"><span data-action="open-rank" role="button" tabindex="0" @click="openRankPanel()" @keydown.enter.prevent="openRankPanel()" @keydown.space.prevent="openRankPanel()">排名:</span><strong data-action="open-rank" role="button" tabindex="0" @click="openRankPanel()" @keydown.enter.prevent="openRankPanel()" @keydown.space.prevent="openRankPanel()">{{ userRank }}</strong></div>
              <div class="lord-meta-row row-office" data-testid="lord-meta-office"><span data-action="open-task" role="button" tabindex="0" @click="openTaskPanel" @keydown.enter.prevent="openTaskPanel" @keydown.space.prevent="openTaskPanel">官职:</span><strong>{{ userOffice }}</strong></div>
              <div class="lord-meta-row row-title" data-testid="lord-meta-title"><span data-action="open-task" role="button" tabindex="0" @click="openTaskPanel" @keydown.enter.prevent="openTaskPanel" @keydown.space.prevent="openTaskPanel">爵位:</span><strong>{{ userNobility }}</strong></div>
              <div class="lord-meta-row row-union" data-testid="lord-meta-union"><span data-action="open-union" role="button" tabindex="0" @click="openUnionPanel" @keydown.enter.prevent="openUnionPanel" @keydown.space.prevent="openUnionPanel">联盟:</span><strong>{{ userUnionName }}</strong></div>
              <div class="lord-meta-row row-post" data-testid="lord-meta-post"><span data-action="open-union" role="button" tabindex="0" @click="openUnionPanel" @keydown.enter.prevent="openUnionPanel" @keydown.space.prevent="openUnionPanel">职位:</span><strong>{{ userUnionPosition }}</strong></div>
            </div>
            <div class="left-action-column">
              <button class="playerinfo-btn king" data-testid="playerinfo-king" data-action="open-lord" type="button" @click="openLordDialog">君主</button>
              <button class="playerinfo-btn armor" data-testid="playerinfo-equipment" data-action="open-equipment" type="button" @click="openEquipmentDialog">装备</button>
              <button class="playerinfo-btn inventory" data-testid="playerinfo-inventory" data-action="open-treasure" type="button" @click="openTreasureDialog">宝物</button>
            </div>
          </div>
          <div class="city-summary-strip">
            <img class="summary-real-icon morale-icon" role="button" tabindex="0" aria-label="民心宝物" :src="asset('city_popularity.png')" alt="" @mouseenter="showMoraleTip" @mouseleave="hideLeftInfoTip" @click="openUseGoodsDialog(4)" @keydown.enter.prevent="openUseGoodsDialog(4)" @keydown.space.prevent="openUseGoodsDialog(4)" />
            <span class="summary-value morale-value" @mouseenter="showFlashToolTip($event, '民心')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ city.morale }}</span>
            <span class="summary-slash">/</span>
            <span class="summary-value complaint-value" @mouseenter="showFlashToolTip($event, '民怨')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ city.complaint }}</span>
            <img class="summary-real-icon tax-icon" role="button" tabindex="0" aria-label="税率" :src="asset('city_tax.png')" alt="" @mouseenter="showFlashToolTip($event, '税率')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openLeftTax" @keydown.enter.prevent="openLeftTax" @keydown.space.prevent="openLeftTax" />
            <span class="summary-value tax-value" @mouseenter="showFlashToolTip($event, '税率')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ city.tax }}%</span>
            <img class="summary-real-icon gold-icon" role="button" tabindex="0" aria-label="黄金宝物" :src="asset('city_gold.png')" alt="" @mouseenter="showGoldTip" @mouseleave="hideLeftInfoTip" @click="openUseGoodsDialog(5)" @keydown.enter.prevent="openUseGoodsDialog(5)" @keydown.space.prevent="openUseGoodsDialog(5)" />
            <span class="summary-value gold-value" @mouseenter="showFlashToolTip($event, '黄金数量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashRoundedInteger(resources?.gold) }}</span>
            <span class="summary-value gold-add" :class="{ negative: leftGoldAdd < 0 }" @mouseenter="showFlashToolTip($event, '黄金产量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(leftGoldAdd) }}</span>
            <img class="summary-real-icon people-icon" role="button" tabindex="0" aria-label="人口宝物" :src="asset('city_population.png')" alt="" @mouseenter="showPeopleTip" @mouseleave="hideLeftInfoTip" @click="openUseGoodsDialog(6)" @keydown.enter.prevent="openUseGoodsDialog(6)" @keydown.space.prevent="openUseGoodsDialog(6)" />
            <span class="summary-value people-value" @mouseenter="showFlashToolTip($event, '当前人口')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(resources?.people) }}</span>
            <span class="summary-value idle-value" :class="{ negative: leftIdlePeople < 0 }" @mouseenter="showFlashToolTip($event, '空闲人口')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(leftIdlePeople) }}</span>
            <button class="left-plus-btn morale-plus" data-testid="left-plus-morale" data-action="open-use-goods" data-goods-type="4" type="button" aria-label="使用宝物提高民心" title="使用宝物提高民心，消除民怨" @mouseenter="showFlashToolTip($event, '使用宝物提高民心，消除民怨')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openUseGoodsDialog(4)"></button>
            <button class="left-plus-btn gold-plus" data-testid="left-plus-gold" data-action="open-use-goods" data-goods-type="5" type="button" aria-label="使用宝物增加黄金税收" title="使用宝物增加黄金税收" @mouseenter="showFlashToolTip($event, '使用宝物增加黄金税收')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openUseGoodsDialog(5)"></button>
            <button class="left-plus-btn people-plus" data-testid="left-plus-people" data-action="open-use-goods" data-goods-type="6" type="button" aria-label="使用宝物增加人口" title="使用宝物增加人口" @mouseenter="showFlashToolTip($event, '使用宝物增加人口')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openUseGoodsDialog(6)"></button>
            <button class="mycity-btn mycity-building" data-testid="left-mycity-building" data-action="open-building-list" type="button" aria-label="建筑信息" title="建筑信息" @mouseenter="showFlashToolTip($event, '建筑信息')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openLeftBuildingList"></button>
            <button class="mycity-btn mycity-labor" data-testid="left-mycity-labor" data-action="open-resource-produce" type="button" aria-label="资源生产" title="资源生产" @mouseenter="showFlashToolTip($event, '资源生产')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openLeftResourceProduce"></button>
            <button class="mycity-btn mycity-field" data-testid="left-mycity-field" data-action="open-city-fields" type="button" aria-label="附属野地" title="附属野地" @mouseenter="showFlashToolTip($event, '附属野地')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openLeftCityFields"></button>
          </div>
          <div class="left-city-header">
            <img class="left-city-title" :src="asset('title.png')" alt="" />
            <div class="flash-city-combo" data-testid="city-switcher" :data-city-x="city.summary.x" :data-city-y="city.summary.y" @click.self="toggleCityDropdown" @mouseleave="cityDropdownOpen = false">
              <button class="city-list" data-testid="city-switcher-button" data-action="toggle-city-switcher" :class="{ open: cityDropdownOpen }" type="button" :aria-expanded="cityDropdownOpen" aria-haspopup="listbox" @click="toggleCityDropdown" @keydown.enter.prevent="toggleCityDropdown" @keydown.space.prevent="toggleCityDropdown" @keydown.escape.prevent="cityDropdownOpen = false">
                {{ selectedCityLabel }}
              </button>
              <div v-if="cityDropdownOpen" class="city-list-menu" data-testid="city-switcher-menu" role="listbox">
                <button
                  v-for="item in cityList.slice(0, 10)"
                  :key="item.cid"
                  class="city-list-option"
                  data-testid="city-switcher-option"
                  data-action="select-city"
                  :class="{ selected: item.cid === selectedCityId }"
                  :data-cid="item.cid"
                  :aria-selected="item.cid === selectedCityId"
                  role="option"
                  type="button"
                  @click="selectCityFromDropdown(item.cid)"
                >
                  {{ item.name }}{{ formatCityCode(item.cid) }}
                </button>
              </div>
            </div>
          </div>
          <div class="left-box city-resource-box">
            <div v-show="leftInfoTab === 'resource'" class="resource-stack">
              <div class="resource-icon-box food"><img role="button" tabindex="0" aria-label="粮食宝物" :src="asset('resource_food.png')" alt="" @mouseenter="showFoodTip" @mouseleave="hideLeftInfoTip" @click="openUseGoodsDialog(7)" @keydown.enter.prevent="openUseGoodsDialog(7)" @keydown.space.prevent="openUseGoodsDialog(7)" /></div>
              <div class="resource-icon-box wood"><img role="button" tabindex="0" aria-label="木材宝物" :src="asset('resource_wood.png')" alt="" @mouseenter="showResourceTip('wood')" @mouseleave="hideLeftInfoTip" @click="openUseGoodsDialog(8)" @keydown.enter.prevent="openUseGoodsDialog(8)" @keydown.space.prevent="openUseGoodsDialog(8)" /></div>
              <div class="resource-icon-box rock"><img role="button" tabindex="0" aria-label="石料宝物" :src="asset('resource_rock.png')" alt="" @mouseenter="showResourceTip('rock')" @mouseleave="hideLeftInfoTip" @click="openUseGoodsDialog(9)" @keydown.enter.prevent="openUseGoodsDialog(9)" @keydown.space.prevent="openUseGoodsDialog(9)" /></div>
              <div class="resource-icon-box iron"><img role="button" tabindex="0" aria-label="铁锭宝物" :src="asset('resource_iron.png')" alt="" @mouseenter="showResourceTip('iron')" @mouseleave="hideLeftInfoTip" @click="openUseGoodsDialog(10)" @keydown.enter.prevent="openUseGoodsDialog(10)" @keydown.space.prevent="openUseGoodsDialog(10)" /></div>
              <div class="resource-value-box food" @mouseenter="showFlashToolTip($event, '粮食数量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(leftFoodValue) }}</div>
              <div class="resource-value-box wood" @mouseenter="showFlashToolTip($event, '木材数量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(leftWoodValue) }}</div>
              <div class="resource-value-box rock" @mouseenter="showFlashToolTip($event, '石料数量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(leftRockValue) }}</div>
              <div class="resource-value-box iron" @mouseenter="showFlashToolTip($event, '铁锭数量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(leftIronValue) }}</div>
              <div class="resource-add-box food" :class="{ negative: leftFoodAdd < 0 }" @mouseenter="showFlashToolTip($event, '粮食产量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(leftFoodAdd) }}</div>
              <div class="resource-add-box wood" @mouseenter="showFlashToolTip($event, '木材产量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(production?.woodAdd) }}</div>
              <div class="resource-add-box rock" @mouseenter="showFlashToolTip($event, '石料产量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(production?.rockAdd) }}</div>
              <div class="resource-add-box iron" @mouseenter="showFlashToolTip($event, '铁锭产量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(production?.ironAdd) }}</div>
              <button class="left-plus-btn resource-plus food" data-testid="left-plus-resource-food" data-action="open-use-goods" data-goods-type="7" type="button" aria-label="使用宝物增加粮食产量" title="使用宝物增加粮食产量" @mouseenter="showFlashToolTip($event, '使用宝物增加粮食产量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openUseGoodsDialog(7)"></button>
              <button class="left-plus-btn resource-plus wood" data-testid="left-plus-resource-wood" data-action="open-use-goods" data-goods-type="8" type="button" aria-label="使用宝物增加木材产量" title="使用宝物增加木材产量" @mouseenter="showFlashToolTip($event, '使用宝物增加木材产量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openUseGoodsDialog(8)"></button>
              <button class="left-plus-btn resource-plus rock" data-testid="left-plus-resource-rock" data-action="open-use-goods" data-goods-type="9" type="button" aria-label="使用宝物增加石料产量" title="使用宝物增加石料产量" @mouseenter="showFlashToolTip($event, '使用宝物增加石料产量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openUseGoodsDialog(9)"></button>
              <button class="left-plus-btn resource-plus iron" data-testid="left-plus-resource-iron" data-action="open-use-goods" data-goods-type="10" type="button" aria-label="使用宝物增加铁锭产量" title="使用宝物增加铁锭产量" @mouseenter="showFlashToolTip($event, '使用宝物增加铁锭产量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openUseGoodsDialog(10)"></button>
              <img class="line-leftboard" :src="asset('line_leftboard.png')" alt="" />
              <div class="activity-panel">
                <div class="activity-content">
                  <a
                    v-for="(item, index) in visibleActivityItems"
                    :key="`${index}-${item.content}`"
                    :href="item.link || undefined"
                    :target="item.link ? '_blank' : undefined"
                  >
                    {{ item.content }}
                  </a>
                </div>
              </div>
            </div>
            <div v-show="leftInfoTab === 'commander'" class="left-info-stack commander">
              <div
                v-for="(hero, index) in leftHeroItems"
                :key="hero.hid"
                class="left-info-item hero"
                :style="leftHeroItemStyle(index)"
              >
                <div class="left-info-hero-frame">
                  <img :src="leftHeroImage(hero)" alt="" @error="($event.target as HTMLImageElement).src = asset('frame_pic.png')" />
                </div>
                <div class="left-info-hero-board">
                  <div class="left-info-hero-name">{{ hero.name }}</div>
                  <div class="left-info-hero-level">{{ hero.level }}级</div>
                  <div class="left-info-hero-state">{{ leftHeroState(hero) }}</div>
                </div>
              </div>
            </div>
            <div v-show="leftInfoTab === 'army'" class="left-info-stack army">
              <div
                v-for="(item, index) in leftSoldierItems"
                :key="item.sid"
                class="left-info-item soldier"
                :style="leftWarItemStyle(index)"
              >
                <div class="left-info-icon-board"><img :src="asset(`images/army_${item.sid}.png`)" alt="" /></div>
                <div class="left-info-text-board">
                  <div class="left-info-name">{{ item.name }}</div>
                  <div class="left-info-count">{{ formatFlashInteger(item.count) }}</div>
                </div>
              </div>
            </div>
            <div v-show="leftInfoTab === 'defence'" class="left-info-stack defence">
              <div
                v-for="(item, index) in leftDefenceItems"
                :key="item.did"
                class="left-info-item defence"
                :style="leftWarItemStyle(index)"
              >
                <div class="left-info-icon-board defence"><img :src="asset(`images/defence_${item.did}.png`)" alt="" /></div>
                <div class="left-info-text-board">
                  <div class="left-info-name">{{ item.name }}</div>
                  <div class="left-info-count">{{ formatFlashInteger(item.count) }}</div>
                </div>
              </div>
            </div>
          </div>
          <div v-if="leftInfoTip.visible" class="left-info-tip" :style="leftInfoTipStyle()">
            <div class="left-info-tip-title">{{ leftInfoTip.title }}</div>
            <div class="left-info-tip-names">
              <div v-for="(row, index) in leftInfoTip.rows" :key="`tip-name-${index}`">{{ row.label }}</div>
            </div>
            <div class="left-info-tip-values">
              <div v-for="(row, index) in leftInfoTip.rows" :key="`tip-value-${index}`" :style="{ color: row.color || undefined }">{{ row.value }}</div>
            </div>
          </div>
          <div v-if="flashToolTip.visible" class="flash-tooltip" :style="{ left: `${flashToolTip.x}px`, top: `${flashToolTip.y}px` }">
            {{ flashToolTip.text }}
          </div>
        </div>
        <div v-if="!worldMapPanelVisible && !battlePanelVisible" class="city-view" :class="cityView === 'inner' ? 'city-view-inner' : 'city-view-outer'">
          <template v-if="cityView === 'inner'">
            <img
              class="inner-map"
              :src="asset(positionBuilding(cityWallPlot.position) ? 'map_innercity_high.jpg' : 'map_innercity_low.jpg')"
              alt=""
            />
            <img
              class="city-wall-img"
              :class="{
                busy: isBusy(positionBuilding(cityWallPlot.position)),
                hovered: hoveredInnerPlot?.position === cityWallPlot.position || selectedPlot?.position === cityWallPlot.position
              }"
              :src="asset(positionBuilding(cityWallPlot.position) ? 'CityInnerPanel_citywall.png' : 'CityInnerPanel_townwall.png')"
              alt=""
            />
            <img
              v-if="showBuildingLevels"
              class="building-level-img citywall-level"
              :class="{ 'townwall-level': !positionBuilding(cityWallPlot.position) }"
              :src="positionBuilding(cityWallPlot.position) ? buildingLevelImage(positionBuilding(cityWallPlot.position)) : asset('level_0.png')"
              alt=""
            />
            <div
              class="cityhall-hit"
              data-testid="cityhall-hit"
              aria-hidden="true"
              :class="{
                occupied: isOccupied(cityHallPlot.position),
                busy: isBusy(positionBuilding(cityHallPlot.position)),
                hovered: hoveredInnerPlot?.position === cityHallPlot.position,
                selected: selectedPlot?.position === cityHallPlot.position
              }"
              :data-position="cityHallPlot.position"
              :data-grid-x="cityHallPlot.gridX"
              :data-grid-y="cityHallPlot.gridY"
              :style="{
                left: `${cityHallPlot.x}px`,
                top: `${cityHallPlot.y}px`,
                width: `${cityHallPlot.w}px`,
                height: `${cityHallPlot.h}px`
              }"
            >
              <img class="cityhall-img" :src="asset('building_cityhall.png')" alt="" />
              <img
                v-if="showBuildingLevels && positionBuilding(cityHallPlot.position)"
                class="building-level-img cityhall-level"
                :src="buildingLevelImage(positionBuilding(cityHallPlot.position))"
                alt=""
              />
              <span v-if="showBuildingLevels && positionBuilding(cityHallPlot.position)" class="building-level-label cityhall-level-label">{{ buildingLevelText(positionBuilding(cityHallPlot.position)) }}</span>
            </div>
            <div
              v-for="plot in innerPlots"
              :key="plot.position"
              class="plot-hit"
              data-testid="inner-plot-hit"
              aria-hidden="true"
              :class="{
                occupied: isOccupied(plot.position),
                busy: isBusy(positionBuilding(plot.position)),
                hovered: hoveredInnerPlot?.position === plot.position,
                selected: selectedPlot?.position === plot.position
              }"
              :data-position="plot.position"
              :data-grid-x="plot.gridX"
              :data-grid-y="plot.gridY"
              :style="{ left: `${plot.x}px`, top: `${plot.y}px`, width: `${plot.w}px`, height: `${plot.h}px` }"
            >
              <span v-if="!positionBuilding(plot.position)" class="empty-building-shadow" aria-hidden="true"></span>
              <img
                v-if="positionBuilding(plot.position)"
                :src="buildingImage(positionBuilding(plot.position)!)"
                :alt="positionBuilding(plot.position)!.name"
              />
              <img
                v-if="showBuildingLevels && positionBuilding(plot.position)"
                class="building-level-img"
                :src="buildingLevelImage(positionBuilding(plot.position))"
                alt=""
              />
              <span v-if="showBuildingLevels && positionBuilding(plot.position)" class="building-level-label">{{ buildingLevelText(positionBuilding(plot.position)) }}</span>
            </div>
            <button
              class="inner-hit-layer"
              data-testid="inner-hit-layer"
              type="button"
              aria-label="城内建筑格"
              @mousemove="handleInnerHitMove"
              @mouseleave="clearInnerHit"
              @click="handleInnerHitClick"
            />
            <button
              class="announce-entry-btn"
              :class="{ hovered: hoveredAnnounce }"
              data-testid="announcement-entry"
              data-action="open-announcement"
              type="button"
              aria-label="公告"
              @mouseenter="showAnnouncementHover"
              @mouseleave="hideAnnouncementHover"
              @click="openAnnouncement"
            >
              <img :src="asset('announce_entry.png')" alt="" />
            </button>
            <div
              v-if="buildingTip.visible"
              class="building-tip"
              :style="{ left: `${buildingTip.x}px`, top: `${buildingTip.y}px` }"
            >
              {{ buildingTip.text }}
            </div>
          </template>
            <template v-else>
              <img class="outer-map" :src="asset('map_outercity.jpg')" alt="" />
              <div
                class="outer-citywall"
                data-testid="outer-citywall-visual"
                data-position="wall"
                aria-hidden="true"
                :class="{ selected: selectedPlot?.position === cityWallPlot.position }"
              >
                <img :src="asset(positionBuilding(cityWallPlot.position) ? 'building_outercity.png' : 'building_outertown.png')" alt="" />
              </div>
              <button
                class="outer-citywall-hit"
                data-testid="outer-citywall-hit"
                data-action="open-build"
                data-position="wall"
                :data-grid-x="cityWallPlot.gridX"
                :data-grid-y="cityWallPlot.gridY"
                :class="{ selected: selectedPlot?.position === cityWallPlot.position }"
                type="button"
                aria-label="城墙"
                @click="openBuild(cityWallPlot)"
              ></button>
              <button
                v-for="plot in visibleOuterPlots"
                :key="`outer-${plot.position}`"
                class="outer-plot-hit"
                :class="{ occupied: isOccupied(plot.position), busy: isBusy(positionBuilding(plot.position)) }"
                data-testid="outer-plot-hit"
                data-action="open-build"
                :data-position="plot.position"
                :data-grid-x="plot.gridX"
                :data-grid-y="plot.gridY"
                :style="{ left: `${plot.x}px`, top: `${plot.y}px`, width: `${plot.w}px`, height: `${plot.h}px` }"
                type="button"
                :aria-label="buildingTipTextForPlot(plot)"
                @click="openBuild(plot)"
              >
              <img
                v-if="positionBuilding(plot.position)"
                class="outer-building-img"
                :src="outerBuildingImage(positionBuilding(plot.position)!)"
                :alt="positionBuilding(plot.position)!.name"
              />
                <img
                  v-if="showBuildingLevels && positionBuilding(plot.position)"
                  class="building-level-img outer-level"
                  :src="buildingLevelImage(positionBuilding(plot.position))"
                  alt=""
                />
                <span v-if="showBuildingLevels && positionBuilding(plot.position)" class="building-level-label outer-level-label">{{ buildingLevelText(positionBuilding(plot.position)) }}</span>
              </button>
          </template>
        </div>
        <div class="bottom-chat">
          <div class="flash-chat-shell">
            <div class="chat-log-overlay" aria-live="polite">
              <div class="chat-log-tabs" aria-hidden="true">
                <button class="chat-log-tab selected" type="button" tabindex="-1" disabled title="聊天记录只读" aria-describedby="bottom-chat-readonly-note">世界</button>
                <button class="chat-log-tab" type="button" tabindex="-1" disabled title="聊天记录只读" aria-describedby="bottom-chat-readonly-note">联盟</button>
                <button class="chat-log-tab" type="button" tabindex="-1" disabled title="聊天记录只读" aria-describedby="bottom-chat-readonly-note">私聊</button>
                <button class="chat-log-tab" type="button" tabindex="-1" disabled title="聊天记录只读" aria-describedby="bottom-chat-readonly-note">战场</button>
              </div>
              <div class="chat-log-lines">
                <div class="chat-log-empty">聊天读取和发送接口尚未接入旧服，当前不显示聊天记录。</div>
              </div>
            </div>
            <div class="chat-console">
              <div class="chat-entry">
                <button class="chat-channel-btn" data-testid="chat-channel-button" type="button" @click="chatChannelMenuVisible = !chatChannelMenuVisible">{{ chatChannelLabel }}</button>
                <input class="chat-input" type="text" value="" disabled tabindex="-1" placeholder="聊天发送未接入" title="聊天发送接口未接入，当前仅显示公告入口。" aria-describedby="bottom-chat-readonly-note" />
                <button class="chat-send-btn readonly" data-testid="chat-send-button" type="button" aria-label="查看聊天未接入说明" title="聊天发送接口未接入，当前仅显示公告入口。" @click="handleChatSend"></button>
                <button class="chat-control-btn first" data-testid="chat-control-button" type="button" aria-label="聊天控制" title="聊天控制" @click="handleChatControl"></button>
                <span id="bottom-chat-readonly-note" class="bottom-chat-readonly-note">聊天发送接口未接入，当前为只读公告栏。</span>
                <div v-if="chatChannelMenuVisible" class="chat-channel-menu" data-testid="chat-channel-menu">
                  <button
                    v-for="channel in chatChannelItems"
                    :key="channel"
                    class="chat-channel-option"
                    data-action="select-chat-channel"
                    :data-channel="channel"
                    type="button"
                    :class="{ selected: channel === chatChannelLabel }"
                    @click="selectChatChannel(channel)"
                  >
                    {{ channel }}
                  </button>
                </div>
              </div>
            </div>
            <div class="bottom-right-shell"></div>
            <div class="right-functions">
              <button class="function-btn friend" data-testid="bottom-function-friend" data-action="open-friend" type="button" aria-label="好友" title="好友" @pointerenter="hoveredBottomFunction = 'friend'" @pointerdown="activeBottomFunction = 'friend'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="leaveSceneForBottomFunction(); openFriendDialog()"><img :src="bottomFunctionImage('friend')" alt="好友" /></button>
              <button class="function-btn report" data-testid="bottom-function-report" data-action="open-report" type="button" aria-label="报告" title="报告" @pointerenter="hoveredBottomFunction = 'report'" @pointerdown="activeBottomFunction = 'report'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="leaveSceneForBottomFunction(); openReportPanel()"><img :src="bottomFunctionImage('report')" alt="报告" /></button>
              <button class="function-btn mail" data-testid="bottom-function-mail" data-action="open-mail" type="button" aria-label="信件" title="信件" @pointerenter="hoveredBottomFunction = 'mail'" @pointerdown="activeBottomFunction = 'mail'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="leaveSceneForBottomFunction(); openMailPanel('inbox')"><img :src="bottomFunctionImage('mail')" alt="信件" /></button>
              <button class="function-btn rank" data-testid="bottom-function-rank" data-action="open-rank" type="button" aria-label="排行" title="排行" @pointerenter="hoveredBottomFunction = 'rank'" @pointerdown="activeBottomFunction = 'rank'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="leaveSceneForBottomFunction(); openRankPanel()"><img :src="bottomFunctionImage('rank')" alt="排行" /></button>
              <button class="function-btn stat" data-testid="bottom-function-stat" data-action="open-stat" type="button" aria-label="统计" title="统计" @pointerenter="hoveredBottomFunction = 'stat'" @pointerdown="activeBottomFunction = 'stat'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="leaveSceneForBottomFunction(); openStatDialog()"><img :src="bottomFunctionImage('stat')" alt="统计" /></button>
              <button class="function-btn forum" data-testid="bottom-function-forum" data-action="open-forum" type="button" aria-label="论坛" title="论坛" @pointerenter="hoveredBottomFunction = 'forum'" @pointerdown="activeBottomFunction = 'forum'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="leaveSceneForBottomFunction(); openForumDialog()"><img :src="bottomFunctionImage('forum')" alt="论坛" /></button>
              <button class="function-btn website" data-testid="bottom-function-website" data-action="open-website" type="button" aria-label="官网" title="官网" @pointerenter="hoveredBottomFunction = 'website'" @pointerdown="activeBottomFunction = 'website'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="leaveSceneForBottomFunction(); openWebsiteDialog()"><img :src="bottomFunctionImage('website')" alt="官网" /></button>
              <button class="function-btn help" data-testid="bottom-function-help" data-action="open-help" type="button" aria-label="帮助" title="帮助" @pointerenter="hoveredBottomFunction = 'help'" @pointerdown="activeBottomFunction = 'help'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="leaveSceneForBottomFunction(); openHelpDialog()"><img :src="bottomFunctionImage('help')" alt="帮助" /></button>
              <button class="function-btn charge" data-testid="bottom-function-charge" data-action="open-charge" type="button" aria-label="充值" title="充值" @pointerenter="hoveredBottomFunction = 'charge'" @pointerdown="activeBottomFunction = 'charge'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="leaveSceneForBottomFunction(); openChargeDialog()"><img :src="bottomFunctionImage('charge')" alt="充值" /></button>
              <button class="function-btn shop" data-testid="bottom-function-shop" data-action="open-shop" type="button" aria-label="商城" title="商城" @pointerenter="hoveredBottomFunction = 'shop'" @pointerdown="activeBottomFunction = 'shop'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="leaveSceneForBottomFunction(); openShopPanel()"><img :src="bottomFunctionImage('shop')" alt="商城" /></button>
            </div>
          </div>
        </div>
        <div v-if="effectiveGuideVisible" class="guide-layer">
          <button
            v-if="guideRect"
            class="guide-hotspot"
            data-action="advance-guide-hotspot"
            type="button"
            tabindex="-1"
            :style="{
              left: `${guideRect.x}px`,
              top: `${guideRect.y}px`,
              width: `${guideRect.w}px`,
              height: `${guideRect.h}px`
            }"
            aria-hidden="true"
            @click="handleGuideHotspotClick"
          />
          <div
            v-if="guideRect"
            class="guide-red-circle"
            :style="{
              left: `${guideRect.x}px`,
              top: `${guideRect.y}px`,
              width: `${guideRect.w}px`,
              height: `${guideRect.h}px`
            }"
          />
          <div class="guide-mask guide-mask-top" :style="{ height: `${guideRect?.y ?? 0}px` }" />
          <div
            class="guide-mask guide-mask-left"
            :style="{ top: `${guideRect?.y ?? 0}px`, width: `${guideRect?.x ?? 0}px`, height: `${guideRect?.h ?? 0}px` }"
          />
          <div
            class="guide-mask guide-mask-right"
            :style="{
              left: `${(guideRect?.x ?? 0) + (guideRect?.w ?? 0)}px`,
              top: `${guideRect?.y ?? 0}px`,
              width: `${1000 - (guideRect?.x ?? 0) - (guideRect?.w ?? 0)}px`,
              height: `${guideRect?.h ?? 0}px`
            }"
          />
          <div
            class="guide-mask guide-mask-bottom"
            :style="{ top: `${(guideRect?.y ?? 0) + (guideRect?.h ?? 0)}px`, height: `${600 - (guideRect?.y ?? 0) - (guideRect?.h ?? 0)}px` }"
          />
          <div
            class="guide-tip"
            :style="{
              left: `${guideTipPosition.x}px`,
              top: `${guideTipPosition.y}px`,
              height: `${guideTipMetrics.height}px`
            }"
          >
            <div class="guide-tip-content">{{ guideText }}</div>
            <button
              class="guide-tip-skip"
              type="button"
              :style="{ top: `${guideTipMetrics.skipTop}px` }"
              @click="guideVisible = false"
            >
              跳过引导
            </button>
          </div>
          <img
            v-if="!guideArrow.hidden"
            class="guide-arrow"
            :src="asset(guideArrow.image)"
            :style="{
              left: `${guideArrow.x}px`,
              top: `${guideArrow.y}px`,
              width: `${guideArrow.w}px`,
              height: `${guideArrow.h}px`,
              transform: guideArrow.transform
            }"
            alt=""
          />
        </div>

        <div v-if="flashConfirm.visible" class="modal-layer flash-confirm-layer" data-testid="flash-confirm-layer">
          <div class="flash-confirm-dialog" data-testid="flash-confirm-dialog">
            <img class="flash-confirm-title-img" :src="asset('title.png')" alt="" />
            <h2>{{ flashConfirm.title }}</h2>
            <div class="flash-confirm-message">{{ flashConfirm.message }}</div>
            <div class="flash-confirm-actions">
              <button class="flash-confirm-ok" data-testid="flash-confirm-ok" type="button" @click="resolveFlashConfirm(true)">确定</button>
              <button class="flash-confirm-cancel" data-testid="flash-confirm-cancel" type="button" @click="resolveFlashConfirm(false)">取消</button>
            </div>
          </div>
        </div>

        <div v-if="announcementVisible" class="modal-layer announcement-layer" data-testid="announcement-layer">
          <div class="announcement-dialog" data-testid="announcement-dialog">
            <img class="announcement-title-img" :src="asset('title.png')" alt="" />
            <h2>今日公告</h2>
            <div class="announcement-top-canvas">
              <div class="announcement-header latest">最新战报</div>
              <div class="announcement-invasion">您当前没有遭到入侵</div>

              <div class="announcement-left-canvas">
                <div class="announcement-left-header">
                  <strong>每日登录奖励</strong>
                  <span>(爵位达到公士以上，每天可以领取奖励。)</span>
                </div>
                <div class="announcement-reward-row row-1">
                  <span class="announcement-reward-icon"></span>
                  <span class="announcement-reward-content">爵位达到公士以上免费领取</span>
                  <button type="button" data-testid="announcement-reward-button" data-action="view-announcement-report" data-reward-index="1" disabled aria-describedby="announcement-reward-readonly-note" title="公告奖励领取接口未接入。">条件不足</button>
                </div>
                <div class="announcement-line line-1"></div>
                <div class="announcement-reward-row row-2 alt">
                  <span class="announcement-reward-icon"></span>
                  <span class="announcement-reward-content">需要&lt;5元宝&gt;</span>
                  <button type="button" data-testid="announcement-reward-button" data-action="view-announcement-report" data-reward-index="2" disabled aria-describedby="announcement-reward-readonly-note" title="公告奖励领取接口未接入。">条件不足</button>
                </div>
                <div class="announcement-line line-2"></div>
                <div class="announcement-reward-row row-3">
                  <span class="announcement-reward-icon"></span>
                  <span class="announcement-reward-content special">需要&lt;5元宝&gt;</span>
                  <button type="button" data-testid="announcement-reward-button" data-action="view-announcement-report" data-reward-index="3" disabled aria-describedby="announcement-reward-readonly-note" title="公告奖励领取接口未接入。">条件不足</button>
                </div>
                <div class="announcement-line line-3"></div>
                <div class="announcement-reward-row row-4 alt">
                  <span class="announcement-reward-icon"></span>
                  <span class="announcement-reward-content special">充值后可免费领取</span>
                  <button type="button" data-testid="announcement-reward-button" data-action="view-announcement-report" data-reward-index="4" disabled aria-describedby="announcement-reward-readonly-note" title="公告奖励领取接口未接入。">条件不足</button>
                </div>
                <div class="announcement-line line-4"></div>
                <div class="announcement-reward-row row-5">
                  <span class="announcement-reward-icon hidden"></span>
                  <span class="announcement-reward-content special"></span>
                  <button type="button" data-testid="announcement-reward-button" data-action="view-announcement-report" data-reward-index="5" disabled aria-describedby="announcement-reward-readonly-note" title="公告奖励领取接口未接入。">条件不足</button>
                </div>
                <div id="announcement-reward-readonly-note" class="announcement-reward-note">公告奖励领取接口未接入，当前仅显示旧版预览。</div>
              </div>

              <div class="announcement-right-title">英雄榜</div>
              <div class="announcement-right-card card-1">
                <strong>位高权重</strong>
                <span class="hero-name">{{ user?.name || '暂无' }}</span>
                <span>官职：{{ userOffice }} 爵位：{{ userNobility }}</span>
              </div>
              <div class="announcement-right-card card-2">
                <strong>攻无不克</strong>
                <span>暂无</span>
              </div>
              <div class="announcement-right-card card-3">
                <strong>固若金汤</strong>
                <span>暂无</span>
              </div>
              <div class="announcement-frame-canvas"></div>
            </div>
            <button class="announcement-close" data-testid="announcement-close" data-action="close-announcement" type="button" @click="announcementVisible = false">关闭</button>
          </div>
        </div>

        <div v-if="buildingListDialogVisible" class="modal-layer">
          <div class="building-list-dialog" data-testid="building-list-dialog">
            <img class="building-list-title-img" :src="asset('title.png')" alt="" />
            <h2>建筑信息</h2>
            <div class="building-list-board building-list-queue" data-testid="building-list-queue">
              <div class="building-list-grid queue-grid">
                <div class="building-list-head">
                  <span>正在建造</span>
                  <span>任务</span>
                  <span>当前等级</span>
                  <span>目标等级</span>
                  <span>剩余时间</span>
                  <span>完成时间</span>
                  <span>操作</span>
                </div>
                <button
                  v-for="item in buildingListQueueItems"
                  :key="`queue-${item.position}-${item.state}`"
                  class="building-list-row"
                  data-testid="building-list-queue-row"
                  data-action="open-building-from-list"
                  :data-bid="item.bid"
                  :data-position="item.position"
                  type="button"
                  @click="openBuildingFromList(item)"
                >
                  <span>{{ item.name }}</span>
                  <span>{{ item.task }}</span>
                  <span>{{ item.level }}</span>
                  <span>{{ item.nextLevel }}</span>
                  <span>{{ item.timeLeft }}</span>
                  <span>{{ item.endTime }}</span>
                  <span>查看</span>
                </button>
                <div v-if="buildingListQueueItems.length === 0" class="building-list-empty">暂无建造队列</div>
              </div>
            </div>
            <div class="building-list-board building-list-outer" data-testid="building-list-outer">
              <div class="building-list-grid compact-grid">
                <div class="building-list-head">
                  <span>城外建筑</span>
                  <span>等级</span>
                  <span>状态</span>
                </div>
                <button
                  v-for="item in buildingListOuterItems"
                  :key="`outer-${item.position}`"
                  class="building-list-row"
                  data-testid="building-list-outer-row"
                  data-action="open-building-from-list"
                  :data-bid="item.bid"
                  :data-position="item.position"
                  type="button"
                  @click="openBuildingFromList(item)"
                >
                  <span>{{ item.name }}</span>
                  <span>{{ displayBuildingLevel(item) }}</span>
                  <span>{{ item.state === 0 ? '空闲' : item.state === 2 ? '拆除中' : '升级中' }}</span>
                </button>
                <div v-if="buildingListOuterItems.length === 0" class="building-list-empty compact-empty">暂无城外建筑</div>
              </div>
            </div>
            <div class="building-list-board building-list-inner" data-testid="building-list-inner">
              <div class="building-list-grid compact-grid">
                <div class="building-list-head">
                  <span>城内建筑</span>
                  <span>等级</span>
                  <span>状态</span>
                </div>
                <button
                  v-for="item in buildingListInnerItems"
                  :key="`inner-${item.position}`"
                  class="building-list-row"
                  data-testid="building-list-inner-row"
                  data-action="open-building-from-list"
                  :data-bid="item.bid"
                  :data-position="item.position"
                  type="button"
                  @click="openBuildingFromList(item)"
                >
                  <span>{{ item.name }}</span>
                  <span>{{ displayBuildingLevel(item) }}</span>
                  <span>{{ item.state === 0 ? '空闲' : item.state === 2 ? '拆除中' : '升级中' }}</span>
                </button>
                <div v-if="buildingListInnerItems.length === 0" class="building-list-empty compact-empty">暂无城内建筑</div>
              </div>
            </div>
            <div class="building-list-remain">可同时建造 1 个队列，扩展队列未接入旧服写接口</div>
            <button
              class="building-list-sync"
              data-testid="building-list-sync"
              data-action="activate-building-queue"
              type="button"
              disabled
              title="未接入旧服建筑队列扩展接口"
            >
              增加队列
            </button>
            <button
              class="building-list-close"
              data-testid="building-list-close"
              data-action="close-building-list"
              type="button"
              @click="closeBuildingListDialog"
            >
              关闭
            </button>
          </div>
        </div>

        <div v-if="buildingPanel && selectedPlot" class="modal-layer">
          <div
            class="building-panel"
            data-testid="building-panel"
            :data-panel-kind="buildingPanel.bid === 6 ? 'government' : 'building'"
            :class="{ 'government-building-panel': buildingPanel.bid === 6 }"
          >
            <button class="building-close" data-testid="building-close" type="button" @click="closeBuild">关闭</button>
            <img class="building-title-img" :src="asset('title.png')" alt="" />
            <h2>{{ buildingPanel.name }}(等级{{ displayBuildingLevel(buildingPanel) }})</h2>
            <div class="building-item">
              <div class="building-icon-frame">
                <img :src="buildingDialogImage(buildingPanel)" :alt="buildingPanel.name" />
              </div>
              <button
                v-if="buildingPanel.state === 0"
                class="building-upgrade-btn"
                data-action="upgrade-building"
                type="button"
                @click="upgradeSelectedBuilding"
              >
                升级
              </button>
              <button
                v-if="buildingPanel.state === 0"
                class="building-destroy-btn"
                data-action="destroy-building"
                type="button"
                @click="requestDestroySelectedBuilding"
              >
                拆除
              </button>
              <button
                v-if="buildingPanel.bid === 7 && buildingPanel.state === 0"
                class="building-research-btn"
                data-testid="building-research-btn"
                data-action="open-college-research"
                type="button"
                @click="openCollegePanel(buildingPanel.position)"
              >
                研究
              </button>
              <button
                v-if="buildingPanel.state !== 0"
                class="building-speed-btn"
                data-action="speed-building"
                type="button"
                @click="requestSpeedSelectedBuilding"
              >
                加速
              </button>
              <button
                v-if="buildingPanel.state !== 0"
                class="building-cancel-btn"
                data-action="cancel-building"
                type="button"
                @click="cancelSelectedBuildingAction"
              >
                取消
              </button>
              <div class="building-info-board" data-testid="building-info-board">
                <div class="building-description">
                  {{ buildingDescriptionText }}
                </div>
                <div class="building-state">
                  {{ buildingNeedText || buildingStateText }}
                </div>
              </div>
            </div>
            <div v-if="buildingPanel.bid === 6" class="government-panel" data-testid="government-panel">
              <div class="government-actions">
                <button type="button" data-testid="government-action-building-list" data-action="overview" @click="openLeftBuildingList">建筑总览</button>
                <button type="button" data-testid="government-action-tax" data-action="tax" @click="openGovernmentSubDialog('tax')">调整税率</button>
                <button type="button" data-testid="government-action-pacify" data-action="pacify" @click="openGovernmentSubDialog('pacify')">安抚百姓</button>
                <button type="button" data-testid="government-action-levy" data-action="levy" @click="openGovernmentSubDialog('levy')">征收物资</button>
                <button type="button" data-testid="government-action-produce" data-action="produce" @click="openGovernmentSubDialog('produce')">资源生产</button>
                <button type="button" data-testid="government-action-rename" data-action="rename" @click="openGovernmentSubDialog('name')">城池改名</button>
                <button type="button" data-testid="government-action-fields" data-action="fields" @click="openGovernmentSubDialog('fields')">附属野地</button>
                <button type="button" data-testid="government-action-cities" data-action="cities" @click="openGovernmentSubDialog('cities')">所有城池</button>
              </div>
              <div class="government-grid">
                <div class="government-grid-head">
                  <span>正在建造</span>
                  <span>任务</span>
                  <span>当前等级</span>
                  <span>目标等级</span>
                  <span>剩余时间</span>
                  <span>完成时间</span>
                </div>
                <button
                  v-for="item in upgradingBuildings"
                  :key="`${item.position}-${item.state}`"
                  class="government-grid-row"
                  type="button"
                  @click="openBuild(item.bid === 6 ? cityHallPlot : { ...cityHallPlot, position: item.position })"
                >
                  <span>{{ item.name }}</span>
                  <span>{{ item.task }}</span>
                  <span>{{ item.level }}</span>
                  <span>{{ item.nextLevel }}</span>
                  <span>{{ item.timeLeft }}</span>
                  <span>{{ item.endTime }}</span>
                </button>
              </div>
            </div>
            <div v-if="buildingPanel.bid === 6 && governmentSubDialog" class="government-sub-layer">
              <div v-if="governmentSubDialog === 'tax'" class="government-popup change-tax-popup">
                <button class="popup-close" type="button" @click="closeGovernmentSubDialog">关闭</button>
                <button class="popup-confirm tax-confirm" type="button" @click="closeGovernmentSubDialog">确定</button>
                <div class="popup-title">调整税率</div>
                <div class="popup-label tax-label-people">人口</div>
                <div class="popup-value tax-value-people">{{ formatNumber(cityPopulation) }}</div>
                <div class="popup-label tax-label-current">民心</div>
                <div class="popup-value tax-value-current">{{ city?.morale ?? 0 }}</div>
                <div class="popup-label tax-label-new">税收</div>
                <input v-model.number="governmentTaxValue" class="tax-stepper" type="number" min="0" max="100" />
                <div class="tax-preview">{{ governmentTaxPreviewText }}</div>
                <textarea class="tax-textarea" readonly></textarea>
              </div>

              <div v-if="governmentSubDialog === 'name'" class="government-popup change-name-popup" data-testid="government-dialog-name">
                <button class="popup-close" data-testid="government-dialog-close" type="button" @click="closeGovernmentSubDialog">关闭</button>
                <button class="popup-confirm name-confirm" data-testid="government-name-confirm" data-action="confirm-rename" type="button" disabled title="城池改名写接口未接入，当前仅显示预览。" aria-describedby="government-name-readonly-note">确定</button>
                <div class="popup-title">城池改名</div>
                <div class="popup-label name-label-current">城池</div>
                <div class="popup-value name-value-current">{{ city?.summary.name }}</div>
                <div class="popup-label name-label-new">新名称</div>
                <div class="popup-value name-input-wrap">
                  <input v-model="governmentCityName" maxlength="20" />
                </div>
                <div class="name-tip">每天只能更改一次城池名，请慎重考虑！</div>
                <div id="government-name-readonly-note" class="government-readonly-note">城池改名写接口未接入，当前仅显示预览。</div>
              </div>

              <div v-if="governmentSubDialog === 'pacify'" class="government-popup pacify-popup" data-testid="government-dialog-pacify">
                <button class="popup-close" data-testid="government-dialog-close" type="button" @click="closeGovernmentSubDialog">关闭</button>
                <button class="popup-confirm pacify-confirm" data-testid="government-pacify-confirm" data-action="confirm-pacify" type="button" disabled title="安抚百姓写接口未接入，当前仅显示预览。" aria-describedby="government-pacify-readonly-note">确定</button>
                <div class="popup-title">安抚百姓</div>
                <div class="popup-label pacify-label-people">人口</div>
                <div class="popup-value pacify-value-people">{{ formatNumber(cityPopulation) }}</div>
                <div class="popup-label pacify-label-morale">民心</div>
                <div class="popup-value pacify-value-morale">{{ city?.morale ?? 0 }}</div>
                <div class="popup-label pacify-label-action">行为</div>
                <span class="government-select-shell pacify-select-shell">
                  <select v-model="governmentPacifyAction" class="pacify-select">
                    <option>赈灾</option>
                    <option>祈福</option>
                    <option>祭天</option>
                    <option>增丁</option>
                  </select>
                </span>
                <textarea class="pacify-preview" readonly :value="governmentPacifyText"></textarea>
                <div id="government-pacify-readonly-note" class="government-readonly-note">安抚百姓写接口未接入，当前仅显示预览。</div>
              </div>

              <div v-if="governmentSubDialog === 'levy'" class="government-popup levy-popup" data-testid="government-dialog-levy">
                <button class="popup-close" data-testid="government-dialog-close" type="button" @click="closeGovernmentSubDialog">关闭</button>
                <button class="popup-confirm levy-confirm" data-testid="government-levy-confirm" data-action="confirm-levy" type="button" disabled title="征收物资写接口未接入，当前仅显示预览。" aria-describedby="government-levy-readonly-note">确定</button>
                <div class="popup-title">征收物资</div>
                <div class="popup-label levy-label-people">人口</div>
                <div class="popup-value levy-value-people">{{ formatNumber(cityPopulation) }}</div>
                <div class="popup-label levy-label-morale">民心</div>
                <div class="popup-value levy-value-morale">{{ city?.morale ?? 0 }}</div>
                <div class="popup-label levy-label-resource">征收</div>
                <span class="government-select-shell levy-select-shell">
                  <select v-model="governmentLevyResource" class="levy-select">
                    <option>黄金</option>
                    <option>粮食</option>
                    <option>木材</option>
                    <option>石料</option>
                    <option>铁锭</option>
                  </select>
                </span>
                <div class="levy-preview">{{ governmentLevyPreview }}</div>
                <textarea class="levy-textarea" readonly></textarea>
                <div id="government-levy-readonly-note" class="government-readonly-note">征收物资写接口未接入，当前仅显示预览。</div>
              </div>

              <div v-if="governmentSubDialog === 'produce'" class="government-full-dialog produce-dialog-panel" data-testid="government-dialog-produce">
                <button class="dialog-close government-full-close" data-testid="government-dialog-close" type="button" @click="closeGovernmentSubDialog">关闭</button>
                <div class="government-full-title wide">资源生产</div>
                <div class="government-list-panel produce-board-panel">
                  <div class="produce-head">
                    <span>资源</span>
                    <span>比例</span>
                    <span>存量</span>
                    <span>修改</span>
                  </div>
                  <div v-for="row in governmentProduceRows" :key="row.label" class="produce-row readonly-row" data-testid="government-produce-row" :data-rate-key="row.label" aria-readonly="true">
                    <span>{{ row.label }}</span>
                    <span><input type="number" :value="row.rate" min="0" max="100" readonly /></span>
                    <span>{{ formatNumber(row.value) }}</span>
                    <span><button type="button" class="city-list-action readonly-shell" data-testid="government-produce-row-confirm" data-action="confirm-production" disabled title="未接入旧服生产比例写接口">修改</button></span>
                  </div>
                  <div class="produce-footer">
                    <span>可用劳力</span>
                    <span>{{ formatNumber(cityPopulation) }}</span>
                    <span>最大劳力</span>
                    <span>{{ formatNumber(cityPopulation) }}</span>
                    <span>开工需求</span>
                    <span>0</span>
                    <span>生产效率</span>
                    <span>100%</span>
                  </div>
                </div>
                <button class="government-return-btn produce-confirm-btn" data-testid="government-produce-confirm" data-action="confirm-production" type="button" disabled title="未接入旧服生产比例写接口">修改</button>
              </div>

              <div v-if="governmentSubDialog === 'fields'" class="government-full-dialog city-field-dialog-panel" data-testid="government-dialog-fields">
                <button class="dialog-close government-full-close field-close" data-testid="government-dialog-close" type="button" @click="closeGovernmentSubDialog">关闭</button>
                <div class="government-full-title">附属野地</div>
                <div class="government-list-panel field-list-panel">
                  <div class="city-field-head">
                    <span>野地</span>
                    <span>位置</span>
                    <span>等级</span>
                    <span>状态</span>
                    <span>查看</span>
                  </div>
                  <div v-for="(row, index) in governmentFieldRows" :key="index" class="city-field-row" data-testid="government-field-row" :data-world-x="row.x" :data-world-y="row.y">
                    <span>{{ row.field }}</span>
                    <span>{{ row.position }}</span>
                    <span>{{ row.level }}</span>
                    <span>{{ row.state }}</span>
                    <span><button type="button" class="city-field-action" data-testid="government-field-inspect" data-action="inspect-field" @click="inspectGovernmentField(row)">查看</button></span>
                  </div>
                </div>
              </div>

              <div v-if="governmentSubDialog === 'cities'" class="government-full-dialog city-list-dialog-panel">
                <button class="dialog-close government-full-close city-list-close" type="button" @click="closeGovernmentSubDialog">关闭</button>
                <button class="government-return-btn" type="button" @click="closeGovernmentSubDialog">返回</button>
                <div class="government-full-title">所有城池</div>
                <div class="government-list-panel city-list-panel">
                  <div class="city-list-head">
                    <span>城池</span>
                    <span>位置</span>
                    <span>城守</span>
                    <span>人口</span>
                    <span>民心</span>
                    <span></span>
                  </div>
                  <div v-for="item in governmentCityItems" :key="item.cid" class="city-list-row readonly-row" aria-readonly="true">
                    <span>{{ item.name }}</span>
                    <span>{{ formatCityCode(item.cid) }}</span>
                    <span>{{ getCityChiefName(item) }}</span>
                    <span>{{ formatNumber(item.resources.people) }}</span>
                    <span>{{ cityMorale(item) }}</span>
                    <span>
                      <button type="button" class="city-list-action readonly-shell" disabled title="城池切换写接口未接入">查看</button>
                    </span>
                  </div>
                  <div v-if="governmentCityItems.length === 0" class="government-empty list-empty">暂无城池数据。</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-if="buildPanel && selectedPlot" class="modal-layer">
          <div class="build-panel" data-testid="build-panel">
            <button class="build-close" data-testid="build-close" type="button" @click="closeBuild">关闭</button>
            <img class="build-title-img" :src="asset('title.png')" alt="" />
            <h2>建造建筑</h2>
            <div v-if="buildPanel.slot.occupied" class="build-message">这里已经有建筑。</div>
            <div v-else-if="!buildPanel.slot.unlocked" class="build-message">官府等级不足，需要{{ buildPanel.slot.unlockLevel }}级。</div>
            <div v-else class="build-list">
              <div
                v-for="option in buildPanel.options"
                :key="option.bid"
                class="build-option"
                data-testid="build-option"
                data-action="select-build-option"
                :data-bid="option.bid"
                :class="{ active: selectedBid === option.bid }"
                :aria-disabled="!option.canBuild"
                @click="selectBuildOption(option)"
              >
                <img :src="buildingIntro(option)" alt="" />
                <button
                  class="create-building-btn"
                  data-testid="create-building-btn"
                  data-action="create-building"
                  type="button"
                  :disabled="!option.canBuild"
                  :title="option.canBuild ? undefined : option.reason"
                  :aria-describedby="option.canBuild ? undefined : `build-option-reason-${option.bid}`"
                  @click.stop="confirmBuild(option)"
                >建造</button>
                <strong>{{ option.name }}</strong>
                <span>{{ buildOptionDescription(option) }}</span>
                <small :id="`build-option-reason-${option.bid}`">{{ option.canBuild ? `建造时间 ${formatDuration(option.duration)}` : option.reason }}</small>
              </div>
            </div>
          </div>
        </div>

        <div v-if="speedGoodsPanel" class="speed-goods-layer">
          <div class="speed-goods-dialog" data-testid="speed-goods-dialog">
            <button class="speed-goods-close" data-testid="speed-goods-close" type="button" @click="speedGoodsPanel = null">关闭</button>
            <img class="speed-goods-title-img" :src="asset('title.png')" alt="" />
            <h2>使用宝物</h2>
            <div class="speed-goods-list">
              <button
                v-for="item in speedGoodsPanel.goodsList"
                :key="item.gid"
                class="speed-goods-item"
                data-testid="speed-goods-item"
                data-action="use-speed-goods"
                :data-gid="item.gid"
                :disabled="item.count <= 0"
                type="button"
                :title="item.count <= 0 ? `${item.name}数量不足。` : `${item.name}: ${speedGoodsEffect(item)}`"
                :aria-describedby="item.count <= 0 ? `speed-goods-reason-${item.gid}` : undefined"
                @click.stop="useSpeedGoods(item)"
              >
                <span class="speed-goods-frame">
                  <img :src="speedGoodsImage(item)" :alt="item.name" @error="($event.target as HTMLImageElement).src = asset('board_listdown.png')" />
                  <span class="speed-goods-count">x{{ item.count }}</span>
                </span>
                <span class="speed-goods-name">{{ item.name }}</span>
                <span class="speed-goods-effect" :id="`speed-goods-reason-${item.gid}`">{{ item.count <= 0 ? '数量不足' : speedGoodsEffect(item) }}</span>
              </button>
            </div>
            <button class="speed-goods-buy" data-testid="speed-goods-buy" type="button" @click="openSpeedGoodsShopPanel">购买</button>
          </div>
        </div>

        <div v-if="useGoodsPanel.visible" class="use-goods-layer">
          <div class="use-goods-dialog" :style="useGoodsDialogStyle" :data-goods-type="useGoodsPanel.type">
            <button class="use-goods-close" data-testid="use-goods-close" type="button" @click="closeUseGoodsDialog">关闭</button>
            <img class="use-goods-title-img" :src="asset('title.png')" alt="" />
            <h2>使用宝物</h2>
            <div class="use-goods-list">
              <div v-if="useGoodsPanel.loading" class="use-goods-message" data-testid="use-goods-loading">加载中...</div>
              <div v-else-if="useGoodsPanel.error" class="use-goods-message" data-testid="use-goods-error">{{ useGoodsPanel.error }}</div>
              <div v-else-if="useGoodsPanel.goodsList.length === 0" class="use-goods-message" data-testid="use-goods-empty">暂无可用宝物</div>
              <template v-else>
                <button
                  v-for="item in useGoodsPanel.goodsList"
                  :key="item.gid"
                  class="use-goods-item"
                  data-testid="use-goods-item"
                  data-action="use-general-goods"
                  :data-gid="item.gid"
                  :disabled="item.count <= 0"
                  type="button"
                  :title="item.count <= 0 ? `${item.name}数量不足。` : item.description"
                  :aria-describedby="item.count <= 0 ? `use-goods-reason-${item.gid}` : undefined"
                  @click.stop="useGeneralGoods(item)"
                >
                  <span class="use-goods-frame">
                    <img :src="useGoodsImage(item)" :alt="item.name" @error="($event.target as HTMLImageElement).src = asset('frame_pic.png')" />
                    <span class="use-goods-count">x{{ item.count }}</span>
                  </span>
                  <span class="use-goods-name">{{ item.name }}</span>
                  <span v-if="item.count <= 0" class="use-goods-effect" :id="`use-goods-reason-${item.gid}`">数量不足</span>
                </button>
              </template>
            </div>
            <button class="use-goods-buy" data-testid="use-goods-buy" data-action="open-use-goods-shop" type="button" @click="openUseGoodsShopPanel">购买</button>
          </div>
        </div>

        <div v-if="taskPanelVisible" class="modal-layer task-layer" :class="{ 'battle-overlay-layer': battlePanelVisible }">
          <div class="task-dialog">
            <button class="task-close" data-testid="task-close" data-action="close-task-panel" type="button" @click="taskPanelVisible = false; selectedTaskId = null; selectedTaskGroupId = null; selectedTaskCategoryType = null">关闭</button>
            <img class="task-title-img" :src="asset('title.png')" alt="" />
            <h2>任务</h2>
            <div class="task-category-tabs">
              <button
                v-for="(category, index) in taskCategories.slice(0, 6)"
                :key="category.type"
                class="task-category-tab"
                data-testid="task-category-tab"
                data-action="select-task-category"
                :data-category-type="category.type"
                :aria-pressed="selectedTaskCategoryType === category.type"
                :class="[`kind-${index}`, { active: selectedTaskCategoryType === category.type }]"
                type="button"
                @click="selectedTaskCategoryType = category.type; selectedTaskGroupId = category.groups[0]?.id ?? null; selectedTaskId = category.groups[0]?.tasks[0]?.id ?? null"
              >
                {{ category.label }}
              </button>
            </div>
            <div class="task-groups">
              <div v-if="selectedTaskGroups.length === 0" class="task-empty">暂无战场任务</div>
              <button
                v-for="group in selectedTaskGroups"
                :key="group.id"
                class="task-group-item"
                data-testid="task-group-item"
                data-action="select-task-group"
                :data-group-id="group.id"
                :aria-pressed="selectedTaskGroupId === group.id"
                :class="{ active: selectedTaskGroupId === group.id }"
                type="button"
                @click="selectedTaskGroupId = group.id; selectedTaskId = group.tasks[0]?.id ?? null"
              >
                <span class="task-group-name">{{ group.name }}</span>
                <span class="task-group-progress">{{ group.completed }}/{{ group.total }}</span>
              </button>
            </div>
            <div class="task-list">
              <div v-if="(selectedTaskGroup?.tasks ?? []).length === 0" class="task-empty">暂无任务</div>
              <button
                v-for="task in selectedTaskGroup?.tasks ?? []"
                :key="task.id"
                class="task-item"
                data-testid="task-item"
                data-action="select-task"
                :data-task-id="task.id"
                :aria-pressed="selectedTaskId === task.id"
                :class="{ active: selectedTaskId === task.id, completed: task.completed }"
                type="button"
                @click="selectTask(task.id)"
              >
                <span class="task-name">{{ task.name }}</span>
                <span class="task-progress">{{ task.completedGoals }}/{{ task.goalCount }}</span>
              </button>
            </div>
            <div class="task-detail">
              <div class="task-detail-title">{{ selectedTask?.name ?? '未选择任务' }}</div>
              <div class="task-detail-body">{{ selectedTask?.content || selectedTask?.todo || '请选择左侧任务查看详情。' }}</div>
              <div v-if="selectedTask?.goals?.length" class="task-goals">
                <div v-for="goal in selectedTask.goals" :key="goal.id" class="task-goal-line">
                  <span>{{ goal.content }}</span>
                  <span>{{ goal.current }}/{{ goal.target }}</span>
                </div>
              </div>
              <div v-if="selectedTask?.rewards?.length" class="task-rewards">
                <span class="task-reward-label">奖励</span>
                <span v-for="(reward, index) in selectedTask.rewards" :key="`${reward.type}-${index}`" class="task-reward-item">
                  {{ reward.name }} x{{ reward.count }}
                </span>
              </div>
            </div>
            <button
              v-if="selectedTaskId"
              class="task-claim-btn"
              data-testid="task-claim-button"
              data-action="claim-task-reward"
              type="button"
              :disabled="taskClaimDisabled"
              :class="{ disabled: taskClaimDisabled }"
              :title="taskClaimTitle"
              aria-describedby="task-claim-readonly-note"
              @click="handleClaimReward"
            >
              领取奖励
            </button>
            <div id="task-claim-readonly-note" class="readonly-note">{{ taskClaimReadonlyNote }}</div>
          </div>
        </div>

        <!-- Mail Dialog -->
        <div v-if="mailPanelVisible" class="modal-layer">
          <div class="dialog-panel mail-dialog">
            <button class="dialog-close" data-testid="mail-close" data-action="close-mail" type="button" @click="mailPanelVisible = false; mailDetailView = null">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>邮件</h2>
            <div class="mail-tabs">
              <button data-testid="mail-tab-inbox" data-action="select-mail-folder" data-folder="inbox" :aria-pressed="mailFolder === 'inbox'" :class="{ active: mailFolder === 'inbox' }" type="button" @click="openMailPanel('inbox')">收件箱</button>
              <button data-testid="mail-tab-sent" data-action="select-mail-folder" data-folder="sent" :aria-pressed="mailFolder === 'sent'" :class="{ active: mailFolder === 'sent' }" type="button" @click="openMailPanel('sent')">发件箱</button>
              <button data-testid="mail-tab-sys" data-action="select-mail-folder" data-folder="sys" :aria-pressed="mailFolder === 'sys'" :class="{ active: mailFolder === 'sys' }" type="button" @click="openMailPanel('sys')">系统邮件</button>
              <button data-testid="mail-tab-compose" data-action="open-mail-compose" data-folder="compose" :aria-pressed="mailFolder === 'compose'" :class="{ active: mailFolder === 'compose' }" type="button" @click="openMailComposePanel">写信</button>
            </div>
            <div v-if="mailFolder === 'compose'" class="mail-compose">
              <label class="mail-compose-row">
                <span>收件人</span>
                <input type="text" value="" readonly aria-describedby="mail-compose-readonly-note" placeholder="收件人" />
              </label>
              <label class="mail-compose-row">
                <span>标题</span>
                <input type="text" value="" readonly aria-describedby="mail-compose-readonly-note" placeholder="标题" />
              </label>
              <label class="mail-compose-row mail-compose-content">
                <span>正文</span>
                <textarea readonly aria-describedby="mail-compose-readonly-note" placeholder="正文"></textarea>
              </label>
              <div id="mail-compose-readonly-note" class="readonly-note">邮件发送写接口未接入，当前仅显示旧版写信面板。</div>
              <div class="mail-compose-actions">
                <button data-testid="mail-compose-send" data-action="send-mail" type="button" disabled title="邮件发送写接口未接入。">发送</button>
                <button data-testid="mail-compose-cancel" data-action="cancel-mail-compose" type="button" @click="closeMailComposePanel">取消</button>
              </div>
            </div>
            <div v-else-if="mailDetailView" class="mail-detail">
              <h3>{{ mailDetailView.title }}</h3>
              <div class="mail-meta">
                <span>{{ mailDetailView.type === 3 ? '系统' : (mailDetailView.type === 2 ? '收件人' : '发件人') }}: {{ mailDetailView.type === 3 ? '' : (mailDetailView.type === 2 ? mailDetailView.to : mailDetailView.from) }}</span>
                <span>{{ new Date(mailDetailView.sendTime * 1000).toLocaleString('zh-CN') }}</span>
              </div>
              <p class="mail-content">{{ mailDetailView.content }}</p>
              <button data-testid="mail-detail-back" data-action="back-mail-list" type="button" @click="mailDetailView = null">返回列表</button>
              <button v-if="mailDetailView.type !== 3" data-testid="mail-detail-delete" data-action="delete-mail" :data-mail-id="mailDetailView.id" type="button" @click="deleteMailItem(mailDetailView.id)">删除</button>
            </div>
            <div v-else class="mail-list">
              <div v-if="mailItems.length === 0" class="empty-list">暂无邮件</div>
              <button
                v-for="item in mailItems"
                :key="item.id"
                class="mail-item"
                data-testid="mail-item"
                data-action="open-mail-detail"
                :data-mail-id="item.id"
                :class="{ unread: !item.read }"
                type="button"
                @click="selectMailItem(item)"
              >
                <span class="mail-item-title">{{ item.title }}</span>
                <span class="mail-item-time">{{ new Date(item.sendTime * 1000).toLocaleDateString('zh-CN') }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Report Dialog -->
        <div v-if="reportPanelVisible" class="modal-layer">
          <div class="dialog-panel report-dialog">
            <button class="dialog-close" data-testid="report-close" data-action="close-report" type="button" @click="reportPanelVisible = false; reportDetailView = null">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>战报</h2>
            <div class="report-tabs">
              <button data-testid="report-tab-all" data-action="select-report-filter" data-filter="all" :aria-pressed="reportFilter === 'all'" :class="{ active: reportFilter === 'all' }" type="button" @click="openReportPanel('all')">全部</button>
              <button data-testid="report-tab-attack" data-action="select-report-filter" data-filter="attack" :aria-pressed="reportFilter === 'attack'" :class="{ active: reportFilter === 'attack' }" type="button" @click="openReportPanel('attack')">进攻</button>
              <button data-testid="report-tab-defend" data-action="select-report-filter" data-filter="defend" :aria-pressed="reportFilter === 'defend'" :class="{ active: reportFilter === 'defend' }" type="button" @click="openReportPanel('defend')">防御</button>
              <button data-testid="report-tab-scout" data-action="select-report-filter" data-filter="scout" :aria-pressed="reportFilter === 'scout'" :class="{ active: reportFilter === 'scout' }" type="button" @click="openReportPanel('scout')">侦察</button>
              <button data-testid="report-tab-unread" data-action="select-report-filter" data-filter="unread" :aria-pressed="reportFilter === 'unread'" :class="{ active: reportFilter === 'unread' }" type="button" @click="openReportPanel('unread')">未读</button>
            </div>
            <div v-if="reportDetailView" class="report-detail">
              <h3>{{ reportDetailView.title }}</h3>
              <p class="report-content">{{ reportDetailView.content }}</p>
              <button data-testid="report-detail-back" data-action="back-report-list" type="button" @click="reportDetailView = null">返回列表</button>
            </div>
            <div v-else class="report-list">
              <div v-if="reportItems.length === 0" class="empty-list">暂无战报</div>
              <button
                v-for="item in reportItems"
                :key="item.id"
                class="report-item"
                data-testid="report-item"
                data-action="open-report-detail"
                :data-report-id="item.id"
                :class="{ unread: !item.read }"
                type="button"
                @click="selectReportItem(item)"
              >
                <span class="report-item-title">{{ item.title }}</span>
                <span class="report-item-time">{{ new Date(item.reportTime * 1000).toLocaleDateString('zh-CN') }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Shop Dialog -->
        <div v-if="shopPanelVisible" class="modal-layer">
          <div class="dialog-panel shop-dialog" data-testid="shop-dialog">
            <button class="dialog-close" data-testid="shop-close" data-action="close-shop" type="button" @click="shopPanelVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>商店</h2>
            <div class="shop-currency">
              <span>黄金: {{ shopSnapshot?.wallet.gold ?? 0 }}</span>
              <span>元宝: {{ shopSnapshot?.wallet.yuanbao ?? 0 }}</span>
              <span>礼金: {{ shopSnapshot?.wallet.gift ?? 0 }}</span>
            </div>
            <div class="shop-group-tabs">
              <button
                v-for="group in shopGroups"
                :key="group.id"
                type="button"
                data-testid="shop-group-tab"
                data-action="select-shop-group"
                :data-group-id="group.id"
                :aria-pressed="selectedShopGroupId === group.id"
                :class="{ active: selectedShopGroupId === group.id }"
                @click="selectedShopGroupId = group.id"
              >
                {{ group.label }}
              </button>
            </div>
            <div class="shop-list">
              <div v-if="selectedShopItems.length === 0" class="empty-list">商店暂无商品</div>
              <div
                v-for="item in selectedShopItems"
                :key="item.id"
                class="shop-item"
                data-testid="shop-item"
                :data-item-id="item.id"
              >
                <img :src="asset(`item_${item.gid}.png`)" alt="" @error="($event.target as HTMLImageElement).src = asset('board_listdown.png')" />
                <div class="shop-item-info">
                  <strong>{{ item.name }}</strong>
                  <span>{{ item.description }}</span>
                  <span class="shop-item-price">{{ item.price }} {{ item.battleShop ? item.medalTypeLabel : '元宝' }}</span>
                  <span class="shop-item-stock">库存: {{ item.totalCount === -1 ? '不限' : item.totalCount }} / 已购: {{ item.boughtToday }}</span>
                </div>
                <button
                  type="button"
                  data-action="buy-shop-item"
                  :data-item-id="item.id"
                  title="购买接口已接入；请确认资源充足。"
                  @click="handleBuyItem(item)"
                >
                  购买
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Friend Dialog -->
        <div v-if="friendDialogVisible" class="modal-layer">
          <div class="dialog-panel friend-dialog">
            <button class="dialog-close" type="button" @click="friendDialogVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>好友</h2>
            <div class="friend-tabs">
              <button
                data-testid="friend-tab-friends"
                type="button"
                :class="{ active: friendRelationTab === 'friends' }"
                @click="friendRelationTab = 'friends'"
              >
                好友
              </button>
              <button
                data-testid="friend-tab-blacklist"
                type="button"
                :class="{ active: friendRelationTab === 'blacklist' }"
                @click="friendRelationTab = 'blacklist'"
              >
                黑名单
              </button>
              <button
                data-testid="friend-tab-requests"
                type="button"
                disabled
                title="好友申请读取接口未接入，当前仅可查看关系名单。"
                aria-describedby="friend-write-readonly-note"
              >
                申请
              </button>
            </div>
            <div class="friend-form">
              <span class="friend-input-shell">
                <input disabled placeholder="输入君主名" title="关系写接口禁用" aria-describedby="friend-write-readonly-note" />
              </span>
              <span class="friend-select-shell">
                <select disabled title="关系写接口禁用" aria-describedby="friend-write-readonly-note">
                  <option value="0">好友</option>
                  <option value="1">黑名单</option>
                </select>
              </span>
              <button data-action="add-friend" type="button" disabled title="关系写接口禁用" aria-describedby="friend-write-readonly-note">添加</button>
            </div>
            <div id="friend-write-readonly-note" class="friend-readonly-note">{{ friendReadonlyNote }}</div>
            <div class="friend-list">
              <div class="friend-head">
                <span>君主</span>
                <span>关系</span>
                <span>联盟 / 城池</span>
                <span>更新时间</span>
                <span>操作</span>
              </div>
              <div v-if="friendRows.length === 0" class="empty-list">
                {{ friendRelationTab === 'friends' ? '暂无好友' : '暂无黑名单' }}
              </div>
              <div
                v-for="row in friendRows"
                :key="`${row.uid}-${row.relationType}`"
                class="friend-row"
                data-testid="friend-row"
              >
                <strong>{{ row.name }}</strong>
                <span>{{ row.relationLabel }}</span>
                <span>{{ row.unionName || '无联盟' }} / {{ row.defaultCity || '-' }}</span>
                <span>{{ row.updatedAt || '-' }}</span>
                <button
                  data-action="delete-friend"
                  type="button"
                  disabled
                  title="关系写接口禁用"
                  aria-describedby="friend-write-readonly-note"
                >
                  删除
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Statistics Dialog -->
        <div v-if="statDialogVisible" class="modal-layer">
          <div class="dialog-panel stat-dialog" data-testid="stat-dialog">
            <button class="dialog-close" type="button" @click="statDialogVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>统计</h2>
            <div class="stat-summary-grid" data-testid="stat-summary-grid">
              <div
                v-for="row in statSummaryRows"
                :key="row.label"
                class="stat-summary-item"
                data-testid="stat-summary-item"
              >
                <span>{{ row.label }}</span>
                <strong>{{ row.value }}</strong>
              </div>
            </div>
            <div class="stat-resource-list">
              <div class="stat-resource-head">
                <span>资源</span>
                <span>当前</span>
                <span>容量</span>
                <span>产量</span>
              </div>
              <div
                v-for="row in statResourceRows"
                :key="row.label"
                class="stat-resource-row"
                data-testid="stat-resource-row"
              >
                <strong>{{ row.label }}</strong>
                <span>{{ formatFlashInteger(row.current) }}</span>
                <span>{{ formatFlashInteger(row.max) }}</span>
                <span>{{ formatFlashInteger(row.add) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Forum Dialog -->
        <div v-if="forumDialogVisible" class="modal-layer">
          <div class="dialog-panel portal-dialog forum-dialog" data-testid="forum-dialog">
            <button class="dialog-close" type="button" @click="forumDialogVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>论坛</h2>
            <div class="portal-board-list">
              <div v-for="row in portalBoardRows" :key="row.title" class="portal-board-row">
                <strong>{{ row.title }}</strong>
                <span>{{ row.text }}</span>
                <em>{{ row.count }}</em>
              </div>
            </div>
            <div class="portal-topic-list">
              <a
                v-for="row in forumTopicRows"
                :key="row.title"
                class="portal-topic-row"
                data-testid="portal-topic-row"
                href="http://bbs.uuyx.com/"
                target="_blank"
                rel="noopener noreferrer"
              >
                <em>{{ row.type }}</em>
                <strong>{{ row.title }}</strong>
                <span>{{ row.date }}</span>
              </a>
            </div>
          </div>
        </div>

        <!-- Website Dialog -->
        <div v-if="websiteDialogVisible" class="modal-layer">
          <div class="dialog-panel portal-dialog website-dialog" data-testid="website-dialog">
            <button class="dialog-close" type="button" @click="websiteDialogVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>官网</h2>
            <div class="portal-notice-box">
              <strong>热血三国</strong>
              <span>旧服官网入口</span>
              <p>迁移版保留原有外部入口；点击右侧链接会在新窗口打开。</p>
              <em>不会在游戏内写入状态。</em>
            </div>
            <div class="website-link-list">
              <a
                v-for="row in websiteLinkRows"
                :key="row.title"
                class="website-link-row"
                data-testid="website-link-row"
                :href="row.url"
                target="_blank"
                rel="noopener noreferrer"
              >
                <strong>{{ row.title }}</strong>
                <span>{{ row.text }}</span>
                <em>{{ row.url }}</em>
              </a>
            </div>
          </div>
        </div>

        <!-- Help Dialog -->
        <div v-if="helpDialogVisible" class="modal-layer">
          <div class="dialog-panel portal-dialog help-dialog" data-testid="help-dialog">
            <button class="dialog-close" type="button" @click="helpDialogVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>帮助</h2>
            <div class="help-rule-list">
              <div v-for="row in helpRuleRows" :key="row.title" class="help-rule-section">
                <h3>{{ row.title }}</h3>
                <p>{{ row.text }}</p>
              </div>
            </div>
            <div class="help-topic-list">
              <a
                v-for="row in helpTopicRows"
                :key="row.title"
                class="help-topic-row"
                data-testid="help-topic-row"
                :href="row.url"
                target="_blank"
                rel="noopener noreferrer"
              >
                <strong>{{ row.title }}</strong>
                <span>{{ row.text }}</span>
              </a>
            </div>
          </div>
        </div>

        <!-- Charge Dialog -->
        <div v-if="chargeDialogVisible" class="modal-layer">
          <div class="dialog-panel charge-dialog" data-testid="charge-dialog">
            <button class="dialog-close" type="button" @click="chargeDialogVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>充值</h2>
            <div class="charge-summary-grid" data-testid="charge-summary-grid">
              <div v-for="row in chargeSummaryRows" :key="row.label" class="charge-summary-item">
                <span>{{ row.label }}</span>
                <strong>{{ row.value }}</strong>
              </div>
            </div>
            <div class="charge-disabled-note">充值兑换写接口未接入，兑换按钮禁用；当前仅显示旧服只读数据。</div>
            <div class="charge-bucket-list">
              <div v-if="chargeBucketRows.length === 0" class="empty-list">暂无充值档位</div>
              <div
                v-for="row in chargeBucketRows"
                :key="row.id"
                class="charge-bucket-row"
                data-testid="charge-bucket-row"
                role="listitem"
                aria-disabled="true"
                title="充值兑换接口未接入，当前禁用。"
              >
                <strong>{{ row.label }}</strong>
                <span>{{ row.minMoney }}-{{ row.maxMoney }} 元</span>
                <em>{{ row.yuanbao }} 元宝</em>
              </div>
            </div>
            <div class="charge-event-list">
              <div v-if="chargeEventRows.length === 0" class="empty-list">暂无充值活动</div>
              <div
                v-for="row in chargeEventRows"
                :key="row.actId"
                class="charge-event-row"
                :class="{ active: row.active }"
                data-testid="charge-event-row"
              >
                <strong>{{ row.name }}</strong>
                <span>{{ row.mailTitle || '活动奖励' }}</span>
                <em>{{ row.startAt }} - {{ row.endAt }}</em>
              </div>
            </div>
          </div>
        </div>

        <!-- World Map Dialog -->
        <div v-if="worldMapPanelVisible" class="modal-layer flash-panel-layer worldmap-layer">
          <div class="worldmap-stage worldmap-map-stage">
            <div class="worldmap-top-hud">
              <span class="worldmap-top-resource">粮 {{ formatFlashInteger(leftFoodValue) }}</span>
              <span class="worldmap-top-resource">木 {{ formatFlashInteger(leftWoodValue) }}</span>
              <span class="worldmap-top-resource">石 {{ formatFlashInteger(leftRockValue) }}</span>
              <span class="worldmap-top-resource">铁 {{ formatFlashInteger(leftIronValue) }}</span>
              <span class="worldmap-top-user">UID {{ user?.uid ?? lastLogin?.uid ?? 0 }}</span>
            </div>
            <div class="worldmap-utility-strip" data-testid="worldmap-utility-strip">
              <button class="function-btn report" data-testid="worldmap-function-report" data-action="open-report" type="button" aria-label="报告" title="报告" @pointerenter="hoveredBottomFunction = 'report'" @pointerdown="activeBottomFunction = 'report'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="openReportPanel()"><img :src="bottomFunctionImage('report')" alt="报告" /></button>
              <button class="function-btn mail" data-testid="worldmap-function-mail" data-action="open-mail" type="button" aria-label="信件" title="信件" @pointerenter="hoveredBottomFunction = 'mail'" @pointerdown="activeBottomFunction = 'mail'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="openMailPanel('inbox')"><img :src="bottomFunctionImage('mail')" alt="信件" /></button>
              <button class="function-btn rank" data-testid="worldmap-function-rank" data-action="open-rank" type="button" aria-label="排行" title="排行" @pointerenter="hoveredBottomFunction = 'rank'" @pointerdown="activeBottomFunction = 'rank'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="openRankPanel()"><img :src="bottomFunctionImage('rank')" alt="排行" /></button>
              <button class="function-btn stat" data-testid="worldmap-function-stat" data-action="open-stat" type="button" aria-label="统计" title="统计" @pointerenter="hoveredBottomFunction = 'stat'" @pointerdown="activeBottomFunction = 'stat'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="openStatDialog"><img :src="bottomFunctionImage('stat')" alt="统计" /></button>
              <button class="function-btn forum" data-testid="worldmap-function-forum" data-action="open-forum" type="button" aria-label="论坛" title="论坛" @pointerenter="hoveredBottomFunction = 'forum'" @pointerdown="activeBottomFunction = 'forum'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="openForumDialog"><img :src="bottomFunctionImage('forum')" alt="论坛" /></button>
              <button class="function-btn website" data-testid="worldmap-function-website" data-action="open-website" type="button" aria-label="官网" title="官网" @pointerenter="hoveredBottomFunction = 'website'" @pointerdown="activeBottomFunction = 'website'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="openWebsiteDialog"><img :src="bottomFunctionImage('website')" alt="官网" /></button>
              <button class="function-btn help" data-testid="worldmap-function-help" data-action="open-help" type="button" aria-label="帮助" title="帮助" @pointerenter="hoveredBottomFunction = 'help'" @pointerdown="activeBottomFunction = 'help'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="openHelpDialog"><img :src="bottomFunctionImage('help')" alt="帮助" /></button>
              <button class="function-btn charge" data-testid="worldmap-function-charge" data-action="open-charge" type="button" aria-label="充值" title="充值" @pointerenter="hoveredBottomFunction = 'charge'" @pointerdown="activeBottomFunction = 'charge'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="openChargeDialog"><img :src="bottomFunctionImage('charge')" alt="充值" /></button>
              <button class="function-btn shop" data-testid="worldmap-function-shop" data-action="open-shop" type="button" aria-label="商城" title="商城" @pointerenter="hoveredBottomFunction = 'shop'" @pointerdown="activeBottomFunction = 'shop'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="openShopPanel()"><img :src="bottomFunctionImage('shop')" alt="商城" /></button>
              <button class="function-btn friend" data-testid="worldmap-function-friend" data-action="open-friend" type="button" aria-label="好友" title="好友" @pointerenter="hoveredBottomFunction = 'friend'" @pointerdown="activeBottomFunction = 'friend'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="openFriendDialog()"><img :src="bottomFunctionImage('friend')" alt="好友" /></button>
            </div>
            <div class="worldmap-map-footer">
              <span>{{ selectedWorldGrid?.title ?? "未选择目标" }}</span>
              <strong>{{ worldMapInputX }},{{ worldMapInputY }}</strong>
            </div>
            <div class="worldmap-sideboard">
              <div class="worldmap-side-section">
                <div class="worldmap-target-name">{{ selectedWorldGrid?.title ?? "未选择目标" }}</div>
                <div class="worldmap-target-text">{{ selectedWorldGrid?.text ?? "" }}</div>
              </div>
            </div>
            <div class="worldmap-console">
              <div class="worldmap-console-channel">
                <button
                  class="worldmap-channel-btn"
                  data-testid="worldmap-channel-button"
                  type="button"
                  :aria-expanded="chatChannelMenuVisible"
                  aria-haspopup="menu"
                  @click="chatChannelMenuVisible = !chatChannelMenuVisible"
                >
                  {{ chatChannelLabel }}
                </button>
                <div v-if="chatChannelMenuVisible" class="worldmap-channel-menu" data-testid="worldmap-channel-menu">
                  <button
                    v-for="channel in chatChannelItems"
                    :key="channel"
                    class="worldmap-channel-option"
                    data-action="select-chat-channel"
                    :data-channel="channel"
                    type="button"
                    :class="{ selected: channel === chatChannelLabel }"
                    @click="selectChatChannel(channel)"
                  >
                    {{ channel }}
                  </button>
                </div>
                <strong>{{ worldMapInputX }},{{ worldMapInputY }}</strong>
              </div>
              <div class="worldmap-console-frame">
                <div class="worldmap-console-title">{{ selectedWorldGrid?.title ?? "未选择目标" }}</div>
                <div class="worldmap-console-tags">
                  <span>目标</span>
                  <span>{{ selectedWorldGrid?.city ? "城池" : "野地" }}</span>
                </div>
                <div class="worldmap-console-message">{{ selectedWorldGrid?.text ?? "请选择地图目标。" }}</div>
              </div>
            </div>
            <div class="worldmap-view-shell">
              <div v-show="worldMapMode === 'map'" class="worldmap-map-panel">
                <button class="worldmap-mini-map-hit" data-testid="worldmap-mini-map-hit" data-action="locate-world-map" type="button" aria-label="地图定位" @click="handleWorldMiniMapClick">
                  <img class="worldmap-map-img" :src="asset('ditu.png')" alt="" />
                </button>
              </div>
              <div v-show="worldMapMode === 'city'" class="worldmap-city-panel">
                <img class="worldmap-city-grid" :src="asset('ditu2.png')" alt="" />
              </div>
              <div v-show="worldMapMode === 'city'" class="worldmap-tip">
                <button
                  v-for="cityItem in worldMapCities"
                  :key="`dot-${cityItem.cid}`"
                  class="worldmap-city-dot"
                  data-testid="worldmap-city-dot"
                  data-action="select-world-city"
                  :data-cid="cityItem.cid"
                  :data-world-x="flashCityPoint(cityItem).x"
                  :data-world-y="flashCityPoint(cityItem).y"
                  type="button"
                  :style="worldMapCityDotStyle(cityItem)"
                  :aria-label="`地图城池点 ${cityItem.name}`"
                  :title="`${cityItem.name}[${flashCityPoint(cityItem).x},${flashCityPoint(cityItem).y}]\n君主:${cityItem.owner}`"
                  @click="selectWorldCity(cityItem)"
                ></button>
                <button
                  v-for="cityItem in worldMapCities"
                  :key="`label-${cityItem.cid}`"
                  class="worldmap-city"
                  :class="{ hidden: cityItem.name !== city?.summary.name }"
                  data-testid="worldmap-city-label"
                  data-action="select-world-city"
                  :data-cid="cityItem.cid"
                  :data-world-x="flashCityPoint(cityItem).x"
                  :data-world-y="flashCityPoint(cityItem).y"
                  type="button"
                  :style="worldMapCityLabelStyle(cityItem)"
                  :aria-label="`地图城池 ${cityItem.name}`"
                  :title="`${cityItem.name}[${flashCityPoint(cityItem).x},${flashCityPoint(cityItem).y}]\n君主:${cityItem.owner}`"
                  @pointerdown.stop
                  @mousedown.stop
                  @click="selectWorldCity(cityItem)"
                >
                  <span class="city-name">{{ cityItem.name }}</span>
                </button>
              </div>
            </div>
            <div class="worldmap-grid-panel">
              <img
                v-for="tile in worldTerrainTiles"
                :key="tile.key"
                class="worldmap-grid-tile"
                :src="asset(tile.image)"
                :style="{ left: `${tile.x}px`, top: `${tile.y}px` }"
                :title="tile.title"
                alt=""
              />
              <button
                class="worldmap-grid-hit-layer"
                data-testid="worldmap-grid-hit-layer"
                data-action="select-world-grid"
                type="button"
                aria-label="地图地块"
                @mousemove="handleWorldGridMove"
                @mouseleave="hideWorldGridTip"
                @click="handleWorldGridClick"
              ></button>
              <div class="worldmap-selected-grid" :style="selectedWorldGridMarkerStyle()"></div>
            </div>
            <div
              v-if="worldGridTip.visible"
              class="worldmap-grid-tip"
              :class="{ city: worldGridTip.city }"
              :style="{ left: `${worldGridTip.x}px`, top: `${worldGridTip.y}px` }"
            >
              <div class="worldmap-grid-tip-title">{{ worldGridTip.title }}</div>
              <div class="worldmap-grid-tip-text">{{ worldGridTip.text }}</div>
            </div>
            <div v-if="selectedWorldGrid" class="worldmap-action-dialog" :style="worldGridActionStyle()">
              <button class="worldmap-action-close" data-testid="worldmap-action-close" data-action="close-world-action" type="button" aria-label="关闭地图动作" @click="closeWorldGridAction"></button>
              <div class="worldmap-action-title">{{ selectedWorldGrid.title }}</div>
              <div class="worldmap-action-body">
                <div>{{ selectedWorldGrid.text }}</div>
              </div>
              <div class="worldmap-action-buttons">
                <button v-if="selectedWorldGrid.city" data-testid="worldmap-action-inspect" data-action="inspect" type="button" @click="inspectWorldTarget">查看</button>
                <button v-if="selectedWorldGrid.city" data-testid="worldmap-action-scout" data-action="scout" data-legacy-task="2" type="button" @click="scoutWorldTarget">侦察</button>
                <button v-if="selectedWorldGrid.city" data-testid="worldmap-action-campaign" data-action="campaign" data-legacy-task="3" type="button" @click="openWorldCampaignDialog(3)">出征</button>
                <button v-if="selectedWorldGrid.city" data-action="occupy" data-legacy-task="4" type="button" disabled :title="WORLD_OCCUPY_DISABLED_NOTE" aria-describedby="worldmap-city-occupy-disabled-note">占领</button>
                <button v-if="!selectedWorldGrid.city" data-action="occupy" data-legacy-task="4" type="button" disabled :title="WORLD_OCCUPY_DISABLED_NOTE" aria-describedby="worldmap-city-occupy-disabled-note">占领</button>
                <button v-if="!selectedWorldGrid.city" data-testid="worldmap-action-scout" data-action="scout" data-legacy-task="2" type="button" @click="scoutWorldTarget">侦察</button>
              </div>
              <div id="worldmap-city-occupy-disabled-note" class="worldmap-action-note">{{ WORLD_OCCUPY_DISABLED_NOTE }}</div>
            </div>
            <div v-if="worldMapAlert" class="worldmap-alert">
              <div class="worldmap-alert-message">{{ worldMapAlert }}</div>
              <button data-testid="worldmap-alert-confirm" data-action="close-world-alert" type="button" @click="closeWorldMapAlert">确定</button>
            </div>
            <div class="worldmap-switch-panel">
              <div class="worldmap-side-strip top">
                <button
                  class="worldmap-side-btn zoom-in"
                  data-testid="worldmap-city-view"
                  data-action="switch-worldmap-mode"
                  data-mode="city"
                  :aria-pressed="worldMapMode === 'city'"
                  :class="{ active: worldMapMode === 'city' }"
                  type="button"
                  aria-label="城池视图"
                  @click="worldMapMode = 'city'"
                ></button>
                <button
                  class="worldmap-side-btn zoom-out"
                  data-testid="worldmap-map-view"
                  data-action="switch-worldmap-mode"
                  data-mode="map"
                  :aria-pressed="worldMapMode === 'map'"
                  :class="{ active: worldMapMode === 'map' }"
                  type="button"
                  aria-label="地图视图"
                  @click="worldMapMode = 'map'"
                ></button>
              </div>
            </div>
            <div class="worldmap-close-panel">
              <div class="worldmap-side-strip bottom">
                <button class="worldmap-side-btn focus-city" data-testid="worldmap-focus-city-view" data-action="switch-worldmap-mode" data-mode="city" type="button" aria-label="城池视图" @click="worldMapMode = 'city'"></button>
                <button class="sr-only" data-testid="worldmap-side-locator-city" data-action="reset-worldmap-city" type="button" aria-label="返回本城" @click="resetWorldMapToCity"></button>
              </div>
            </div>
            <div class="worldmap-control-panel">
              <button class="worldmap-control upleft" data-testid="worldmap-move-upleft" data-action="move-worldmap" data-direction="upleft" type="button" aria-label="左上" @pointerdown.prevent="moveWorldMapByPointer(-WORLD_MOVE_OBLIQUE, 0)" @click="moveWorldMapByClick(-WORLD_MOVE_OBLIQUE, 0)"></button>
              <button class="worldmap-control up" data-testid="worldmap-move-up" data-action="move-worldmap" data-direction="up" type="button" aria-label="上移" @pointerdown.prevent="moveWorldMapByPointer(-WORLD_MOVE_VERT_STRAIGHT, -WORLD_MOVE_VERT_STRAIGHT)" @click="moveWorldMapByClick(-WORLD_MOVE_VERT_STRAIGHT, -WORLD_MOVE_VERT_STRAIGHT)"></button>
              <button class="worldmap-control upright" data-testid="worldmap-move-upright" data-action="move-worldmap" data-direction="upright" type="button" aria-label="右上" @pointerdown.prevent="moveWorldMapByPointer(0, -WORLD_MOVE_OBLIQUE)" @click="moveWorldMapByClick(0, -WORLD_MOVE_OBLIQUE)"></button>
              <button class="worldmap-control left" data-testid="worldmap-move-left" data-action="move-worldmap" data-direction="left" type="button" aria-label="左移" @pointerdown.prevent="moveWorldMapByPointer(-WORLD_MOVE_HORI_STRAIGHT, WORLD_MOVE_HORI_STRAIGHT)" @click="moveWorldMapByClick(-WORLD_MOVE_HORI_STRAIGHT, WORLD_MOVE_HORI_STRAIGHT)"></button>
              <button class="worldmap-control mycity" data-testid="worldmap-move-mycity" data-action="reset-worldmap-city" type="button" title="返回" aria-label="返回本城" @click="resetWorldMapToCity"></button>
              <button class="worldmap-control right" data-testid="worldmap-move-right" data-action="move-worldmap" data-direction="right" type="button" aria-label="右移" @pointerdown.prevent="moveWorldMapByPointer(WORLD_MOVE_HORI_STRAIGHT, -WORLD_MOVE_HORI_STRAIGHT)" @click="moveWorldMapByClick(WORLD_MOVE_HORI_STRAIGHT, -WORLD_MOVE_HORI_STRAIGHT)"></button>
              <button class="worldmap-control downleft" data-testid="worldmap-move-downleft" data-action="move-worldmap" data-direction="downleft" type="button" aria-label="左下" @pointerdown.prevent="moveWorldMapByPointer(0, WORLD_MOVE_OBLIQUE)" @click="moveWorldMapByClick(0, WORLD_MOVE_OBLIQUE)"></button>
              <button class="worldmap-control down" data-testid="worldmap-move-down" data-action="move-worldmap" data-direction="down" type="button" aria-label="下移" @pointerdown.prevent="moveWorldMapByPointer(WORLD_MOVE_VERT_STRAIGHT, WORLD_MOVE_VERT_STRAIGHT)" @click="moveWorldMapByClick(WORLD_MOVE_VERT_STRAIGHT, WORLD_MOVE_VERT_STRAIGHT)"></button>
              <button class="worldmap-control downright" data-testid="worldmap-move-downright" data-action="move-worldmap" data-direction="downright" type="button" aria-label="右下" @pointerdown.prevent="moveWorldMapByPointer(WORLD_MOVE_OBLIQUE, 0)" @click="moveWorldMapByClick(WORLD_MOVE_OBLIQUE, 0)"></button>
              <button class="worldmap-control move" data-testid="worldmap-move-submit" data-action="submit-worldmap-coordinates" type="button" aria-label="跳转坐标" @click="submitWorldMapMove"></button>
              <input :value="worldMapInputX" class="worldmap-input x" type="text" inputmode="numeric" maxlength="3" aria-label="X" @input="worldMapInputX = ($event.target as HTMLInputElement).value" @keydown.enter.prevent="submitWorldMapMove" @blur="submitWorldMapMove" />
              <input :value="worldMapInputY" class="worldmap-input y" type="text" inputmode="numeric" maxlength="3" aria-label="Y" @input="worldMapInputY = ($event.target as HTMLInputElement).value" @keydown.enter.prevent="submitWorldMapMove" @blur="submitWorldMapMove" />
            </div>
          </div>
        </div>

        <!-- Union Dialog -->
        <div v-if="unionPanelVisible" class="modal-layer">
          <div class="dialog-panel union-dialog" data-testid="union-dialog">
            <button class="dialog-close" data-testid="union-close" data-action="close-union" type="button" @click="unionPanelVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>联盟</h2>
            <div v-if="unionSnapshot?.union" class="union-info">
              <h3>{{ unionSnapshot.union.name }}</h3>
              <p>盟主: {{ unionSnapshot.union.leader }}</p>
              <p>等级: {{ unionSnapshot.union.level }}</p>
              <p>成员: {{ unionSnapshot.union.memberCount }}/{{ unionSnapshot.union.maxMembers }}</p>
              <p>公告: {{ unionSnapshot.union.announcement }}</p>
              <div class="union-members">
                <h4>成员列表</h4>
                <div v-for="member in unionSnapshot.union.members" :key="member.uid" class="union-member">
                  <span>{{ member.name }}</span>
                  <span>Lv.{{ member.level }}</span>
                  <span>{{ ['成员', '官员', '盟主'][member.position - 1] }}</span>
                  <span>贡献: {{ member.contribute }}</span>
                </div>
              </div>
            </div>
            <div v-else class="union-none">
              <p>您还没有加入联盟</p>
              <div
                v-if="unionSnapshot?.application"
                id="union-application-note"
                class="union-application"
                data-testid="union-application"
              >
                <span>已申请: {{ unionSnapshot.application.unionName }}</span>
                <span>{{ unionSnapshot.application.createdAt }}</span>
                <button
                  data-testid="union-cancel-apply"
                  data-action="cancel-union-apply"
                  type="button"
                  title="取消当前联盟申请"
                  aria-describedby="union-application-note"
                  @click="handleCancelUnionApply"
                >
                  取消
                </button>
              </div>
              <div v-if="unionSnapshot?.applyList.length" class="union-apply-list">
                <h4>可申请联盟</h4>
                <button
                  v-for="u in unionSnapshot.applyList"
                  :key="u.id"
                  data-testid="union-apply-button"
                  data-action="apply-union"
                  :data-union-id="u.id"
                  type="button"
                  :disabled="!!unionApplyDisabledReason(u)"
                  :title="unionApplyDisabledReason(u) || `申请加入${u.name}`"
                  :aria-describedby="unionApplyDisabledReason(u) ? 'union-application-note' : undefined"
                  @click="handleApplyUnion(u)"
                >
                  {{ u.isApplied ? `${u.name}(已申请)` : u.name }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <div v-if="lordDialogVisible" class="modal-layer">
          <div class="dialog-panel lord-dialog" data-testid="lord-dialog">
            <button class="dialog-close" data-testid="lord-close" data-action="close-lord" type="button" @click="lordDialogVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>君主信息</h2>
            <div class="lord-dialog-body">
              <div class="lord-card">
                <div class="lord-card-portrait"><img :src="faceImage" alt="" /></div>
                <strong>{{ lordName }}</strong>
                <span>{{ userOffice }} / {{ userNobility }}</span>
              </div>
              <div class="lord-info-list">
                <div v-for="row in lordInfoRows" :key="row.label" class="lord-info-row">
                  <span>{{ row.label }}</span>
                  <strong>{{ row.value }}</strong>
                </div>
              </div>
            </div>
            <div class="lord-dialog-actions">
              <button data-testid="lord-action-rank" data-action="open-rank" type="button" @click="openRankPanel()">排行</button>
              <button data-testid="lord-action-task" data-action="open-task" type="button" @click="openTaskPanel">任务</button>
              <button data-testid="lord-action-union" data-action="open-union" type="button" @click="openUnionPanel">联盟</button>
            </div>
          </div>
        </div>

        <div v-if="equipmentDialogVisible" class="modal-layer">
          <div class="dialog-panel equipment-dialog" data-testid="equipment-dialog">
            <button class="dialog-close" data-testid="equipment-close" data-action="close-equipment" type="button" @click="equipmentDialogVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>装备</h2>
            <div class="equipment-owner">
              <img :src="faceImage" alt="" />
              <strong>{{ lordName }}</strong>
              <span>当前装备只读预览</span>
            </div>
            <div class="equipment-slot-grid">
              <div v-for="slot in equipmentSlots" :key="slot" class="equipment-slot" data-testid="equipment-slot">
                <span class="equipment-slot-name">{{ slot }}</span>
                <strong>未装备</strong>
                <em>旧服装备接口未接入</em>
                <div class="equipment-actions">
                  <button type="button" data-action="offload-equipment" disabled :title="equipmentReadonlyNote" aria-describedby="equipment-readonly-note">卸下</button>
                  <button type="button" data-action="repair-equipment" disabled :title="equipmentReadonlyNote" aria-describedby="equipment-readonly-note">修理</button>
                  <button type="button" data-action="renovate-equipment" disabled :title="equipmentReadonlyNote" aria-describedby="equipment-readonly-note">改造</button>
                </div>
              </div>
            </div>
            <div class="equipment-inventory">
              <h3>装备包裹</h3>
              <button class="equipment-repair-all" data-action="repair-all-equipment" type="button" disabled :title="equipmentReadonlyNote" aria-describedby="equipment-readonly-note">全部修理</button>
              <button class="equipment-renovate-all" data-action="renovate-all-equipment" type="button" disabled :title="equipmentReadonlyNote" aria-describedby="equipment-readonly-note">全部改造</button>
              <div class="equipment-inventory-row" data-testid="equipment-inventory-row">
                <strong>暂无装备</strong>
                <span>包裹为空</span>
                <em>当前为旧服只读兼容窗口</em>
                <div class="equipment-actions">
                  <button type="button" data-action="equip-equipment" disabled :title="equipmentReadonlyNote" aria-describedby="equipment-readonly-note">装备</button>
                  <button type="button" data-action="repair-equipment" disabled :title="equipmentReadonlyNote" aria-describedby="equipment-readonly-note">修理</button>
                  <button type="button" data-action="renovate-equipment" disabled :title="equipmentReadonlyNote" aria-describedby="equipment-readonly-note">改造</button>
                  <button type="button" data-action="recycle-equipment" disabled :title="equipmentReadonlyNote" aria-describedby="equipment-readonly-note">回收</button>
                </div>
              </div>
              <div id="equipment-readonly-note" class="readonly-note">{{ equipmentReadonlyNote }}</div>
            </div>
          </div>
        </div>

        <div v-if="treasureDialogVisible" class="modal-layer">
          <div class="dialog-panel treasure-dialog" data-testid="treasure-dialog">
            <button class="dialog-close" data-testid="treasure-close" data-action="close-treasure" type="button" @click="treasureDialogVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>宝物</h2>
            <div class="treasure-list">
              <div v-for="row in treasureRows" :key="row.name" class="treasure-row" data-testid="treasure-row" role="listitem" aria-describedby="treasure-readonly-note">
                <span>{{ row.name }}</span>
                <strong>{{ row.text }}</strong>
              </div>
            </div>
            <div class="treasure-wallet">
              <span>元宝</span>
              <strong>{{ formatFlashInteger(treasureWallet) }}</strong>
            </div>
            <div id="treasure-readonly-note" class="readonly-note treasure-readonly-note">{{ treasureReadonlyNote }}</div>
          </div>
        </div>

        <!-- Hero Dialog -->
        <div v-if="heroPanelVisible" class="modal-layer">
          <div class="dialog-panel hero-dialog">
            <button class="dialog-close" data-testid="hero-close" data-action="close-hero" type="button" @click="heroPanelVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>武将</h2>
            <div class="hero-list">
              <div v-if="heroRosterItems.length === 0" class="empty-list">暂无武将</div>
              <div
                v-for="hero in heroRosterItems"
                :key="hero.hid"
                class="hero-item"
              >
                <div class="hero-info">
                  <strong>{{ hero.name }}</strong>
                  <span>Lv.{{ hero.level }}</span>
                  <span>武力:{{ hero.武力 ?? hero.bravery ?? 0 }} 智力:{{ hero.智力 ?? hero.wisdom ?? 0 }} 统兵:{{ hero.统兵 ?? hero.command ?? 0 }}</span>
                  <span>{{ leftHeroState(hero) }}</span>
                </div>
                <div class="hero-actions">
                  <button type="button" disabled title="任命接口只读未接入" aria-describedby="hero-action-readonly-note">任命城守</button>
                  <button type="button" disabled title="任命接口只读未接入" aria-describedby="hero-action-readonly-note">任命军师</button>
                  <button type="button" disabled title="加点接口只读未接入" aria-describedby="hero-action-readonly-note">武力+1</button>
                  <button type="button" disabled title="加点接口只读未接入" aria-describedby="hero-action-readonly-note">智力+1</button>
                  <button type="button" disabled title="加点接口只读未接入" aria-describedby="hero-action-readonly-note">统兵+1</button>
                </div>
              </div>
            </div>
            <div class="hero-recruit">
              <span id="hero-action-readonly-note" class="readonly-note">将领任命和加点接口保持只读。</span>
              <span>可招募名额: {{ heroRecruitCapacity }}</span>
              <button data-testid="hero-recruit-button" data-action="recruit-hero" type="button" @click="handleRecruitHero">招募</button>
            </div>
          </div>
        </div>

        <!-- Barracks Dialog -->
        <div v-if="barracksPanelVisible" class="modal-layer">
          <div class="dialog-panel barracks-dialog">
            <button class="dialog-close" data-testid="barracks-close" data-action="close-barracks" type="button" @click="barracksPanelVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>兵营</h2>
            <p class="barracks-capacity">
              兵力上限: {{ troopsData?.maxCapacity ?? 0 }}
              <span>出征 {{ troopsData?.moving ?? 0 }}</span>
              <span>返回 {{ troopsData?.returning ?? 0 }}</span>
              <span>驻守 {{ troopsData?.stationed ?? 0 }}</span>
            </p>
            <div class="troops-list barracks-owned-list">
              <h3>城内兵力</h3>
              <div v-if="barracksTroopItems.length === 0" class="empty-list">暂无士兵</div>
              <div
                v-for="troop in barracksTroopItems"
                :key="troop.tid"
                class="troop-item"
              >
                <span class="troop-name">{{ troop.name }}</span>
                <span>编制: {{ troop.count }}</span>
                <span>伤兵: {{ troop.injured }}</span>
                <button
                  data-testid="barracks-train-button"
                  data-action="train-troop"
                  :data-tid="troop.tid"
                  type="button"
                  :title="`训练${troop.name}`"
                  @click="handleTrainTroop(troop.tid, 1)"
                >训练</button>
              </div>
            </div>
            <div class="troops-list barracks-draft-list">
              <h3>招募士兵</h3>
              <div
                v-for="slot in flashSoldierSlots"
                :key="`draft-${slot.sid}`"
                class="troop-item draft-option disabled"
                data-testid="barracks-draft-row"
                data-action="draft-troop-readonly"
                :data-sid="slot.sid"
              >
                <div class="draft-option-main">
                  <strong>{{ slot.name }}</strong>
                  <span>当前 {{ leftSoldierItems.find((item) => item.sid === slot.sid)?.count ?? 0 }}</span>
                  <em>当前城池尚未建造可用兵营</em>
                </div>
                <input value="0" disabled title="当前城池尚未建造可用兵营" aria-describedby="barracks-readonly-note" />
                <button type="button" disabled title="当前城池尚未建造可用兵营" aria-describedby="barracks-readonly-note">招募</button>
              </div>
              <div id="barracks-readonly-note" class="readonly-note">当前城池尚未建造可用兵营，招募接口保持只读。</div>
            </div>
            <div class="barracks-queue-list">
              <h3>招募队列</h3>
              <div class="queue-item empty-list">暂无招募队列</div>
            </div>
            <div class="barracks-external-list">
              <h3>城外部队</h3>
              <div class="barracks-external-summary">出征 {{ troopsData?.moving ?? 0 }}　返回 {{ troopsData?.returning ?? 0 }}　驻守 {{ troopsData?.stationed ?? 0 }}</div>
              <div
                v-for="item in troopsData?.items ?? []"
                :key="item.id"
                class="external-troop-item"
                data-testid="external-troop-row"
              >
                <span>{{ item.heroName || '无将领' }}</span>
                <span>{{ item.taskLabel || item.stateLabel || '行军中' }}</span>
                <span>{{ item.fromCity }} → {{ item.targetCity }}</span>
                <button data-action="recall-troop" type="button" disabled title="召回接口未接入" aria-describedby="barracks-readonly-note">召回</button>
              </div>
              <div v-if="!(troopsData?.items?.length)" class="external-troop-item empty-list">暂无城外部队</div>
            </div>
          </div>
        </div>

        <!-- College/Research Dialog -->
        <div v-if="collegePanelVisible" class="modal-layer">
          <div class="dialog-panel college-dialog">
            <button class="dialog-close" data-testid="college-close" data-action="close-college" type="button" @click="collegePanelVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>学院</h2>
            <div v-if="researchSnapshot?.researching" class="researching">
              <p>正在研究: {{ researchSnapshot.researching.name }} (Lv.{{ researchSnapshot.researching.level }})</p>
            </div>
            <div class="tech-list">
              <div v-if="researchSnapshot?.available.length === 0" class="empty-list">暂无可研究科技</div>
              <div
                v-for="tech in researchSnapshot?.available ?? []"
                :key="tech.tid"
                class="tech-item"
                :class="{ disabled: !tech.canResearch }"
                data-testid="college-tech-row"
                :data-tid="tech.tid"
              >
                <div class="tech-info">
                  <strong>{{ tech.name }}</strong>
                  <span>{{ tech.description }}</span>
                  <span>当前: Lv.{{ tech.level }} / 最大: Lv.{{ tech.maxLevel }}</span>
                  <span
                    v-if="!tech.canResearch"
                    class="tech-reason"
                    :id="`college-tech-reason-${tech.tid}`"
                  >
                    {{ tech.reason || '条件不足' }}
                  </span>
                </div>
                <button
                  data-testid="college-research-button"
                  data-action="research-tech"
                  :data-tid="tech.tid"
                  type="button"
                  :disabled="!tech.canResearch"
                  :title="tech.canResearch ? `研究${tech.name}` : (tech.reason || '条件不足')"
                  :aria-describedby="!tech.canResearch ? `college-tech-reason-${tech.tid}` : undefined"
                  @click="handleResearchTech(tech.tid)"
                >研究</button>
              </div>
            </div>
          </div>
        </div>

        <!-- Ranking Dialog -->
        <div v-if="rankPanelVisible" class="modal-layer">
          <div class="dialog-panel rank-dialog">
            <button class="dialog-close" data-testid="rank-close" data-action="close-rank" type="button" @click="rankPanelVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>排行榜</h2>
            <div class="rank-tabs">
              <button data-testid="rank-tab-power" data-action="select-rank-kind" data-rank-kind="power" :aria-pressed="rankKind === 'power'" :class="{ active: rankKind === 'power' }" type="button" @click="openRankPanel('power')">实力</button>
              <button data-testid="rank-tab-level" data-action="select-rank-kind" data-rank-kind="level" :aria-pressed="rankKind === 'level'" :class="{ active: rankKind === 'level' }" type="button" @click="openRankPanel('level')">等级</button>
              <button data-testid="rank-tab-city" data-action="select-rank-kind" data-rank-kind="city" :aria-pressed="rankKind === 'city'" :class="{ active: rankKind === 'city' }" type="button" @click="openRankPanel('city')">城池</button>
            </div>
            <div class="rank-list">
              <div v-if="rankItems.length === 0" class="empty-list">暂无排行数据</div>
              <div
                v-for="item in rankItems"
                :key="item.uid"
                class="rank-item"
              >
                <span class="rank-num">#{{ item.rank }}</span>
                <span class="rank-name">{{ item.name }}</span>
                <span class="rank-value">{{ item.value }}</span>
                <span v-if="item.detail" class="rank-detail">{{ item.detail }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Battle Dialog -->
        <div v-if="battlePanelVisible" class="modal-layer flash-panel-layer battle-layer">
          <div class="battle-stage">
            <div class="battle-bg-image" :style="{ backgroundImage: `url(${asset(`battle_map_${battleActiveMapId}.png`)})` }"></div>
            <div
              v-for="item in battleBloodItems"
              :key="item.side"
              class="battle-blooditem"
              :class="`battle-blooditem-${item.side}`"
              :style="{ width: `${item.width}px` }"
            ></div>
            <button
              v-for="item in battleFieldObjects"
              :key="item.key"
              :class="item.className"
              :data-testid="`battle-field-object-${item.key}`"
              type="button"
              aria-describedby="battle-field-object-readonly-note"
              title="查看旧服战场目标部队；派遣、巡逻、攻击仍为只读预览。"
              @click="openBattleFieldViewDialog"
            ></button>
            <span id="battle-field-object-readonly-note" class="sr-only">可查看旧服战场目标部队；写操作保持只读。</span>
            <div class="battle-canvas"></div>
            <div class="battle-flag-canvas"></div>
            <div class="battle-hero-panel">
              <button
                v-for="slot in battleHeroSlotViews"
                :key="slot.index"
                class="battle-hero-slot"
                data-testid="battle-hero-slot"
                data-action="open-battle-hero-slot"
                :data-slot-index="slot.index"
                :data-troop-id="slot.troop?.id ?? 0"
                :data-has-troop="slot.troop ? '1' : '0'"
                type="button"
                aria-describedby="battle-hero-slot-readonly-note"
                :title="slot.title"
                @click="slot.troop ? openBattleTroopDetailDialog(undefined, slot.troop) : openBattleTroopViewDialog(true)"
              >
                <span class="battle-hero-slot-canvas"></span>
                <span class="battle-hero-slot-image">
                  <img v-if="slot.image" :src="slot.image" :alt="slot.troop?.heroName || ''" @error="($event.target as HTMLImageElement).style.display = 'none'" />
                  <span v-if="!slot.image" class="battle-hero-slot-fallback">{{ slot.label }}</span>
                </span>
              </button>
              <span id="battle-hero-slot-readonly-note" class="sr-only">可查看旧服己方战场部队详情；写操作保持只读。</span>
            </div>
            <div class="battle-label"></div>
            <div class="battle-menu-panel">
              <button class="battle-menu-btn" data-testid="battle-menu-btn" data-action="open-battle-menu" type="button" aria-label="战场菜单" @click="openBattleMenu"></button>
            </div>
          </div>
        </div>

        <div v-if="battleFieldViewVisible" class="modal-layer battle-overlay-layer">
          <div class="battle-field-view-dialog" data-testid="battle-field-view-dialog">
            <div class="battle-field-view-inner">
              <div class="battle-field-title" data-testid="battle-field-title">目标部队</div>
              <button class="battle-field-send-btn" data-testid="battle-field-send-btn" type="button" aria-describedby="battle-field-readonly-note" :disabled="!battleFieldCanSendPreview" :title="battleFieldSendTitle" @click="openBattleActionDialog">派遣</button>
              <button class="battle-field-close-btn" data-testid="battle-field-close-btn" type="button" @click="closeBattleFieldViewDialog">关闭</button>
              <div class="battle-field-grid-panel" data-testid="battle-field-grid-panel">
                <div class="battle-field-grid" data-testid="battle-field-grid">
                  <div class="battle-field-grid-header">
                    <span>玩家</span>
                    <span>联盟</span>
                    <span>将领</span>
                    <span>状态</span>
                    <span>操作</span>
                  </div>
                  <div
                    v-for="row in battleFieldRows"
                    :key="row.id"
                    class="battle-field-grid-row"
                    :class="{ selected: selectedBattleFieldRow?.id === row.id }"
                    data-testid="battle-field-row"
                    :data-row-id="row.id"
                    :data-can-view="row.canView ? '1' : '0'"
                    :data-can-patrol="row.canPatrol ? '1' : '0'"
                    :data-can-attack="row.canAttack ? '1' : '0'"
                  >
                    <span :title="row.name">{{ row.name || '-' }}</span>
                    <span :title="row.union">{{ row.union || '-' }}</span>
                    <span :title="row.hero">{{ row.hero || '-' }}</span>
                    <span :title="row.stateLabel">{{ row.stateLabel || '-' }}</span>
                    <span class="battle-field-row-actions">
                      <button data-action="view" type="button" aria-describedby="battle-field-readonly-note" title="读取旧服部队详情，只读预览。" :style="{ display: row.canView ? '' : 'none' }" @click="openBattleTroopDetailDialog(row)">查看</button>
                      <button data-action="patrol" type="button" aria-describedby="battle-field-readonly-note" title="接口未接入，只读确认巡逻。" :style="{ display: row.canPatrol ? '' : 'none' }" @click="handleBattleFieldPatrol(row)">巡逻</button>
                      <button data-action="attack" type="button" aria-describedby="battle-field-readonly-note" title="读取旧服攻击预览，不会写入。" :style="{ display: row.canAttack ? '' : 'none' }" @click="selectBattleFieldRow(row); openBattleAttackDialog()">攻击</button>
                    </span>
                  </div>
                  <div v-if="battleFieldRows.length === 0" class="battle-field-empty">暂无目标部队</div>
                </div>
              </div>
              <div id="battle-field-readonly-note" class="battle-field-note">{{ battleFieldReadonlyNote }}</div>
            </div>
          </div>
        </div>

        <div v-if="battleActionVisible" class="modal-layer battle-overlay-layer">
          <div class="battle-action-dialog" data-testid="battle-action-dialog">
            <div class="battle-action-inner" data-testid="battle-action-inner">
              <div class="battle-action-title" data-testid="battle-action-title">派遣部队</div>
              <div class="battle-action-field target"><span>目标</span><span>{{ battleActionTargetName }}</span></div>
              <div class="battle-action-field hero"><span>将领</span><span>{{ battleActionHeroName }}</span></div>
              <div class="battle-action-field level"><span>等级</span><span>{{ battleActionHeroLevel }}</span></div>
              <div class="battle-action-field single-time"><span>耗时</span><span>{{ battleActionPathTime }}</span></div>
              <div class="battle-action-field arrival"><span>到达</span><span>{{ battleActionArrival }}</span></div>
              <div class="battle-action-grid-panel" data-testid="battle-action-grid-panel">
                <div class="battle-action-grid" data-testid="battle-action-soldier-grid">
                  <div class="battle-action-grid-header"><span>兵种</span><span>数量</span></div>
                  <div v-for="(soldier, index) in battleActionSoldiers" :key="battleSoldierKey(soldier, index)" class="battle-action-grid-row">
                    <span>{{ soldier.name }}</span>
                    <span>{{ formatFlashInteger(soldier.count) }}</span>
                  </div>
                  <div v-if="battleActionSoldiers.length === 0" class="battle-action-grid-empty">暂无可派遣部队</div>
                </div>
              </div>
              <div id="battle-action-readonly-note" class="battle-action-note">{{ battleActionReadonlyNote }}</div>
              <button class="battle-action-submit-btn" data-testid="battle-action-submit-btn" type="button" disabled aria-describedby="battle-action-readonly-note" title="接口未接入，暂不能派遣。">确定</button>
              <button class="battle-action-close-btn" data-testid="battle-action-close-btn" type="button" @click="closeBattleActionDialog">关闭</button>
            </div>
          </div>
        </div>

        <div v-if="battleAttackVisible" class="modal-layer battle-overlay-layer">
          <div class="battle-attack-dialog" data-testid="battle-attack-dialog">
            <div class="battle-attack-inner" data-testid="battle-attack-inner">
              <div class="battle-attack-title" data-testid="battle-attack-title">攻击部队</div>
              <div class="battle-attack-field target"><span>目标</span><span>{{ battleAttackTargetName }}</span></div>
              <div class="battle-attack-field time"><span>耗时</span><span>{{ battleAttackPathTime }}</span></div>
              <div class="battle-attack-field arrival"><span>到达</span><span>{{ battleAttackArrival }}</span></div>
              <div class="battle-attack-field my-hero"><span>我方</span><span>{{ battleAttackMyHeroName }}</span></div>
              <div class="battle-attack-field target-hero"><span>敌方</span><span>{{ battleAttackTargetHeroName }}</span></div>
              <div class="battle-attack-grid-panel my" data-testid="battle-attack-my-grid-panel">
                <div class="battle-attack-grid" data-testid="battle-attack-my-grid">
                  <div class="battle-attack-grid-header"><span>兵种</span><span>数量</span></div>
                  <div v-for="(soldier, index) in battleActionSoldiers" :key="battleSoldierKey(soldier, index)" class="battle-attack-grid-row">
                    <span>{{ soldier.name }}</span>
                    <span>{{ formatFlashInteger(soldier.count) }}</span>
                  </div>
                  <div v-if="battleActionSoldiers.length === 0" class="battle-attack-grid-empty">暂无我方部队</div>
                </div>
              </div>
              <div class="battle-attack-grid-panel target" data-testid="battle-attack-target-grid-panel">
                <div class="battle-attack-grid" data-testid="battle-attack-target-grid">
                  <div class="battle-attack-grid-header"><span>兵种</span><span>数量</span></div>
                  <div v-for="(soldier, index) in battleTargetSoldiers" :key="battleSoldierKey(soldier, index)" class="battle-attack-grid-row">
                    <span>{{ soldier.name }}</span>
                    <span>{{ formatFlashInteger(soldier.count) }}</span>
                  </div>
                  <div v-if="battleTargetSoldiers.length === 0" class="battle-attack-grid-empty">暂无目标部队</div>
                </div>
              </div>
              <div id="battle-attack-readonly-note" class="battle-attack-note">{{ battleAttackReadonlyNote }}</div>
              <button class="battle-attack-submit-btn" data-testid="battle-attack-submit-btn" type="button" disabled aria-describedby="battle-attack-readonly-note" title="接口未接入，暂不能攻击。">确定</button>
              <button class="battle-attack-close-btn" data-testid="battle-attack-close-btn" type="button" @click="closeBattleAttackDialog">关闭</button>
            </div>
          </div>
        </div>

        <div v-if="battlePatrolVisible" class="modal-layer battle-overlay-layer">
          <div class="battle-patrol-dialog" data-testid="battle-patrol-dialog">
            <div class="battle-patrol-inner" data-testid="battle-patrol-inner">
              <div class="battle-patrol-title" data-testid="battle-patrol-title">巡逻结果</div>
              <div class="battle-patrol-field target"><span>目标</span><span>{{ battlePatrolPreviewData?.targetName || battleAttackTargetName }}</span></div>
              <div class="battle-patrol-field city"><span>地点</span><span>{{ battlePatrolPreviewData?.targetCity || battleFieldName() }}</span></div>
              <div class="battle-patrol-field hero"><span>敌将</span><span>{{ battlePatrolPreviewData?.target.hero || selectedBattleFieldRow?.hero || "未知" }}</span></div>
              <div class="battle-patrol-field level"><span>等级</span><span>{{ formatFlashInteger(battlePatrolPreviewData?.target.level ?? selectedBattleFieldRow?.level) }}</span></div>
              <div class="battle-patrol-field pigeon"><span>信鸽</span><span>{{ formatFlashInteger(battlePatrolPreviewData?.pigeonCount ?? 0) }}</span></div>
              <div class="battle-patrol-grid-panel" data-testid="battle-patrol-grid-panel">
                <div class="battle-patrol-grid" data-testid="battle-patrol-grid">
                  <div class="battle-patrol-grid-header"><span>侦察报告</span></div>
                  <div v-for="(line, index) in battlePatrolReportLines" :key="`${index}-${line}`" class="battle-patrol-grid-row">
                    <span>{{ line }}</span>
                  </div>
                  <div v-if="battlePatrolReportLines.length === 0" class="battle-patrol-grid-empty">暂无侦察结果</div>
                </div>
              </div>
              <div id="battle-patrol-readonly-note" class="battle-patrol-note">{{ battlePatrolReadonlyNote }}</div>
              <button class="battle-patrol-close-btn" data-testid="battle-patrol-close-btn" type="button" @click="closeBattlePatrolDialog">关闭</button>
            </div>
          </div>
        </div>

        <div v-if="battleTroopViewVisible" class="modal-layer battle-overlay-layer">
          <div class="battle-troop-view-dialog" data-testid="battle-troop-view-dialog">
            <div class="battle-troop-view-inner" data-testid="battle-troop-view-inner">
              <div class="battle-troop-view-title" data-testid="battle-troop-view-title">部队查看</div>
              <div class="battle-troop-view-grid-panel" data-testid="battle-troop-view-grid-panel">
                <div class="battle-troop-view-grid" data-testid="battle-troop-view-grid">
                  <div class="battle-troop-grid-header"><span>将领</span><span>兵力</span></div>
                  <button
                    v-for="troop in battleCurrentTroops"
                    :key="troop.id"
                    class="battle-troop-grid-row battle-troop-row-button"
                    :class="{ selected: selectedBattleCurrentTroop?.id === troop.id }"
                    data-testid="battle-troop-row-button"
                    data-action="open-battle-troop-detail"
                    :data-troop-id="troop.id"
                    type="button"
                    @click="openBattleTroopDetailDialog(undefined, troop)"
                  >
                    <span :title="troop.heroName">{{ troop.heroName || '未知' }}</span>
                    <span>{{ formatFlashInteger(troop.soldierCount) }}</span>
                  </button>
                  <div v-if="battleCurrentTroops.length === 0" class="battle-troop-grid-empty">暂无战场部队</div>
                </div>
              </div>
              <div id="battle-troop-view-readonly-note" class="battle-troop-view-note">{{ battleTroopViewNote }}</div>
              <button class="dialog-close battle-troop-view-close-btn" data-testid="battle-troop-view-close-btn" type="button" @click="closeBattleTroopViewDialog">关闭</button>
            </div>
          </div>
        </div>

        <div v-if="battleTroopDetailVisible" class="modal-layer battle-overlay-layer">
          <div class="battle-troop-detail-dialog" data-testid="battle-troop-detail-dialog">
            <div class="battle-troop-detail-inner" data-testid="battle-troop-detail-inner">
              <div class="battle-troop-detail-title" data-testid="battle-troop-detail-title">部队详情</div>
              <div class="battle-troop-detail-field user"><span>玩家</span><span>{{ battleSelectedDetail.user }}</span></div>
              <div class="battle-troop-detail-field union"><span>联盟</span><span>{{ battleSelectedDetail.union }}</span></div>
              <div class="battle-troop-detail-field hero"><span>将领</span><span>{{ battleSelectedDetail.hero }}</span></div>
              <div class="battle-troop-detail-field level"><span>等级</span><span>{{ battleSelectedDetail.level }}</span></div>
              <div class="battle-troop-detail-field state"><span>状态</span><span>{{ battleSelectedDetail.state }}</span></div>
              <div class="battle-troop-detail-field target"><span>目标</span><span>{{ battleSelectedDetail.target }}</span></div>
              <div class="battle-troop-detail-field left-time"><span>剩余</span><span>{{ battleSelectedDetail.leftTime }}</span></div>
              <div class="battle-troop-detail-field arrival"><span>到达</span><span>{{ battleSelectedDetail.arrival }}</span></div>
              <div id="battle-troop-detail-readonly-note" class="battle-troop-detail-note">{{ battleTroopDetailNote }}</div>
              <div class="battle-troop-detail-grid-panel" data-testid="battle-troop-detail-grid-panel">
                <div class="battle-troop-detail-grid" data-testid="battle-troop-detail-grid">
                  <div class="battle-troop-grid-header"><span>兵种</span><span>数量</span></div>
                  <div v-for="(soldier, index) in battleSelectedDetail.soldiers" :key="battleSoldierKey(soldier, index)" class="battle-troop-grid-row">
                    <span>{{ soldier.name }}</span>
                    <span>{{ formatFlashInteger(soldier.count) }}</span>
                  </div>
                  <div v-if="battleSelectedDetail.soldiers.length === 0" class="battle-troop-grid-empty">暂无部队详情</div>
                </div>
              </div>
              <button v-show="battleTroopPreviewIsMyArmy" class="battle-troop-detail-call-btn" data-testid="battle-troop-detail-call-btn" type="button" disabled aria-describedby="battle-troop-detail-readonly-note" title="接口未接入，暂不能召回。">召回</button>
              <button v-show="battleTroopPreviewIsMyArmy" class="battle-troop-detail-speed-btn" data-testid="battle-troop-detail-speed-btn" type="button" disabled aria-describedby="battle-troop-detail-readonly-note" title="接口未接入，暂不能加速。">加速</button>
              <button v-show="battleTroopPreviewIsMyArmy" class="battle-troop-detail-recall-btn" data-testid="battle-troop-detail-recall-btn" type="button" disabled aria-describedby="battle-troop-detail-readonly-note" title="接口未接入，暂不能撤退。">撤退</button>
              <button class="battle-troop-detail-close-btn" data-testid="battle-troop-detail-close-btn" type="button" @click="closeBattleTroopDetailDialog">关闭</button>
            </div>
          </div>
        </div>

        <div v-if="battleMenuVisible" class="modal-layer battle-overlay-layer">
          <div class="battle-menu-dialog" data-testid="battle-menu-dialog">
            <div class="battle-menu-surface">
              <div class="battle-menu-title">战场</div>
              <button class="battle-menu-action primary" data-testid="battle-menu-action-info" data-action="info" title="战场说明" type="button" @click="openBattleInfoDialog">战场说明</button>
              <button class="battle-menu-action" data-testid="battle-menu-action-campaign" data-action="campaign" type="button" @click="openBattleCampaignDialog">战场出征</button>
              <button class="battle-menu-action" data-testid="battle-menu-action-task" data-action="task" type="button" @click="openBattleTaskDialog">战场任务</button>
              <button class="battle-menu-action" data-testid="battle-menu-action-users" data-action="users" type="button" @click="openBattleUsersDialog">战场成员</button>
              <button class="battle-menu-action" data-testid="battle-menu-action-quit" data-action="quit" type="button" @click="quitBattleField">退出战场</button>
              <button class="battle-menu-action" data-testid="battle-menu-action-close" data-action="close" type="button" @click="closeBattleMenu">关闭菜单</button>
            </div>
          </div>
        </div>

        <div v-if="battleCampaignVisible" class="modal-layer battle-overlay-layer">
          <div class="battle-simple-dialog battle-campaign-dialog" data-testid="battle-campaign-dialog">
            <div class="battle-campaign-inner" aria-hidden="true"></div>
            <button class="dialog-close" data-testid="battle-campaign-close" data-action="close-battle-campaign" type="button" @click="closeBattleCampaignDialog">关闭</button>
            <div class="battle-simple-title">{{ battleCampaignTitle }}</div>
            <div class="battle-campaign-left">
              <div class="battle-campaign-left-title">{{ battleCampaignLeftTitle }}</div>
              <button class="battle-campaign-all-btn" data-testid="battle-campaign-soldier-all" data-action="select-all-battle-soldiers" type="button" @click="takeAllBattleSoldiers">全选</button>
              <button class="battle-campaign-none-btn" data-testid="battle-campaign-soldier-none" data-action="clear-battle-soldiers" type="button" @click="takeNoBattleSoldiers">全不选</button>
              <label class="battle-campaign-left-flag" data-testid="battle-campaign-flag-row">
                <input data-testid="battle-campaign-flag" type="checkbox" :checked="battleCampaignUseFlag" @change="toggleBattleCampaignUseFlag" />
                <span class="battle-campaign-flag-box" aria-hidden="true"></span>
                <span>旗帜</span>
              </label>
              <div class="battle-campaign-soldier-list">
                <div v-for="(soldier, index) in battleCampaignSoldiers" :key="soldier.name" class="battle-campaign-soldier-row" data-testid="battle-campaign-soldier-row" :data-sid="soldier.sid">
                  <div class="battle-campaign-soldier-slot">
                    <span class="battle-campaign-soldier-fallback">{{ soldier.name.slice(0, 1) }}</span>
                    <img :src="battleSoldierIcon(soldier.sid)" :alt="soldier.name" @error="($event.target as HTMLImageElement).style.display = 'none'" />
                    <span class="battle-campaign-soldier-owned">{{ soldier.count }}</span>
                  </div>
                  <div class="battle-campaign-soldier-input">
                    <input data-action="soldier-count" :value="soldier.takecount" @input="onBattleSoldierInput(index, $event)" />
                    <button data-action="soldier-max" type="button" @click="setBattleSoldierCount(index, soldier.count)">最大</button>
                    <button data-action="soldier-min" type="button" @click="setBattleSoldierCount(index, -soldier.count)">最小</button>
                  </div>
                </div>
              </div>
            </div>
            <div class="battle-simple-row battle-campaign-target-row" data-testid="battle-campaign-target-row">
              <span>{{ battleCampaignTargetLabel }}</span>
              <span class="battle-campaign-select-shell">
                <select v-model="battleCampaignTargetId" data-testid="battle-campaign-target" :data-target-id="battleCampaignTargetId" @change="refreshBattleCampaignPreview">
                  <option v-for="target in battleCampaignTargets" :key="target.id" :value="target.id">{{ target.name }}</option>
                </select>
              </span>
            </div>
            <div class="battle-simple-row battle-campaign-hero-row" data-testid="battle-campaign-hero-row">
              <span>将领选择</span>
              <span class="battle-campaign-select-shell">
                <select v-model="battleCampaignHeroId" data-testid="battle-campaign-hero" :aria-describedby="battleCampaignHeroReadonlyReason ? 'battle-campaign-hero-note' : undefined" :title="battleCampaignHeroReadonlyReason || undefined" @change="refreshBattleCampaignPreview">
                  <option v-for="hero in battleCampaignHeroes" :key="hero.id" :value="hero.id">{{ hero.heroname }}</option>
                </select>
              </span>
            </div>
            <div id="battle-campaign-hero-note" class="battle-campaign-hero-note">{{ battleCampaignHeroReadonlyReason }}</div>
            <div class="battle-simple-row battle-campaign-food-row">
              <span>{{ battleCampaignFoodLabel }}</span>
              <span data-testid="battle-campaign-food">{{ battleCampaignFoodCarry }}</span>
            </div>
            <div class="battle-simple-row battle-campaign-field-row" data-testid="battle-campaign-field-row">
              <span>{{ battleCampaignFieldLabel }}</span>
              <span>{{ battleCampaignFieldName }}</span>
            </div>
            <div class="battle-simple-row battle-campaign-arrive-row" data-testid="battle-campaign-arrive-row">
              <span>到达时间</span>
              <span>{{ battleCampaignArriveTime }}</span>
            </div>
            <div class="battle-simple-row battle-campaign-path-row" data-testid="battle-campaign-path-row">
              <span>{{ battleCampaignPathLabel }}</span>
              <span>{{ battleCampaignPathNeedTime }}</span>
            </div>
            <button class="battle-simple-start" data-action="start-dispatch" type="button" :disabled="!!battleCampaignStartDisabledReason" :aria-describedby="battleCampaignStartNote ? 'battle-campaign-start-note' : undefined" :title="battleCampaignStartNote || undefined" @click="startBattleCampaignDispatch">{{ battleCampaignActionLabel }}</button>
            <div id="battle-campaign-start-note" class="battle-campaign-start-note">{{ battleCampaignStartNote }}</div>
          </div>
        </div>

        <div v-if="battleInfoVisible" class="modal-layer battle-overlay-layer">
          <div class="battle-simple-dialog battle-info-dialog">
            <div class="battle-info-inner">
              <button class="dialog-close" data-testid="battle-info-close" data-action="close-battle-info" type="button" @click="closeBattleInfoDialog">关闭</button>
              <div class="battle-simple-title">战场说明</div>
              <div class="battle-simple-tabs">
                <button class="battle-info-tab info" data-testid="battle-info-tab-info" data-action="select-battle-info-tab" data-tab="info" :aria-pressed="battleInfoTab === 'info'" :class="{ active: battleInfoTab === 'info' }" type="button" @click="setBattleInfoTab('info')">战场说明</button>
                <button class="battle-info-tab news" data-testid="battle-info-tab-news" data-action="select-battle-info-tab" data-tab="news" :aria-pressed="battleInfoTab === 'news'" :class="{ active: battleInfoTab === 'news' }" type="button" :title="battleNewsReadonlyTitle" @click="setBattleInfoTab('news')">战场讯息</button>
                <template v-if="battleInfoTab === 'news'">
                <div class="battle-simple-page">{{ battleInfoPage }}/{{ battleNewsPageCount }}</div>
                <button class="battle-info-page-btn prev" data-testid="battle-info-page-prev" data-action="prev-battle-info-page" type="button" @click="prevBattleInfoPage">上一页</button>
                <button class="battle-info-page-btn next" data-testid="battle-info-page-next" data-action="next-battle-info-page" type="button" @click="nextBattleInfoPage">下一页</button>
                </template>
              </div>
              <div v-if="battleInfoTab === 'info'" class="battle-simple-body battle-info-body">
                <img class="battle-info-image" :src="battleInfoImageSrc" alt="" />
                <div class="battle-info-meta">
                  <div v-for="item in battleInfoMeta" :key="item.label"><span>{{ item.label }}</span><span>{{ item.value }}</span></div>
                </div>
                <p>{{ battleInfoContent }}</p>
              </div>
              <div v-if="battleInfoTab === 'news'" class="battle-simple-body battle-info-news-body">
                <div class="battle-news-header">
                  <span>时间</span>
                  <span>内容</span>
                </div>
                <div class="battle-news-grid">
                  <div
                    v-for="item in battleInfoNewsPageItems"
                    :key="`${item.time}-${item.evtContent}`"
                    class="battle-news-row"
                    :style="{ color: `#${item.color.toString(16).padStart(6, '0')}` }"
                  >
                    <span class="battle-news-time">{{ item.time }}</span>
                    <span class="battle-news-content">{{ item.evtContent }}</span>
                  </div>
                  <div v-if="battleInfoNewsPageItems.length === 0" class="battle-news-empty">没有任何战场讯息</div>
                </div>
              </div>
              <div id="battle-info-readonly-note" class="battle-simple-note">{{ battleNewsReadonlyNote }}</div>
              <div class="battle-help-line">更多帮助请参考 http://action.uuyx.com/GameHelp/Help1/index.html</div>
            </div>
          </div>
        </div>

        <div v-if="battleUsersVisible" class="modal-layer battle-overlay-layer">
          <div class="battle-invite-users-dialog" data-testid="battle-invite-dialog">
            <button class="dialog-close" data-testid="battle-invite-close-btn" type="button" @click="closeBattleUsersDialog">关闭</button>
            <div class="battle-invite-title" data-testid="battle-invite-title">战场成员</div>
            <div class="battle-invite-count-label">战场人数</div>
            <div class="battle-invite-count-value" data-testid="battle-invite-count">{{ battleInviteDisplayCount }}</div>
            <div class="battle-invite-name-label">邀请玩家</div>
            <div class="battle-invite-name-input" data-testid="battle-invite-name-input-shell">
              <input data-testid="battle-invite-name-input" v-model="battleInviteName" maxlength="50" disabled aria-describedby="battle-invite-readonly-note" title="成员邀请接口未接入。" />
            </div>
            <button class="battle-invite-send" data-testid="battle-invite-submit-btn" type="button" disabled aria-describedby="battle-invite-readonly-note" title="成员邀请接口未接入。">邀请加入</button>
            <div id="battle-invite-readonly-note" class="battle-invite-note">成员邀请接口未接入，只读预览不会写入。</div>
            <div class="battle-invite-grid" data-testid="battle-invite-grid">
              <div class="battle-invite-grid-header">
                <span>玩家</span>
                <span>阵营</span>
                <span>状态</span>
                <span>将领数</span>
                <span>荣誉</span>
                <span>操作</span>
              </div>
              <div v-if="battleInviteDisplayUsers.length === 0" class="battle-invite-empty">暂无战场成员</div>
              <div v-for="item in battleInviteDisplayUsers" :key="item.id" class="battle-invite-grid-row" data-testid="battle-invite-row">
                <span>{{ item.name }}</span>
                <span>{{ item.camp }}</span>
                <span>{{ item.state }}</span>
                <span>{{ item.herocount }}</span>
                <span>{{ item.honour }}</span>
                <span class="battle-invite-action-cell"><button v-if="item.cancel" data-action="cancel-invite" type="button" disabled aria-describedby="battle-invite-readonly-note" title="成员邀请写接口未接入。">取消邀请</button></span>
              </div>
            </div>
          </div>
        </div>

        <div v-if="chatSendDialogVisible" class="modal-layer">
          <div class="dialog-panel chat-dialog-panel" data-testid="chat-send-dialog">
            <button class="dialog-close" data-testid="chat-send-close" type="button" @click="chatSendDialogVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>聊天发送</h2>
            <div class="chat-dialog-log">
              <div class="chat-dialog-empty">聊天读取和发送接口尚未接入旧服，当前不显示聊天记录。</div>
            </div>
            <div class="chat-dialog-form">
              <span class="chat-select-shell">
                <select disabled title="聊天发送接口未接入" aria-describedby="chat-send-readonly-note">
                  <option>{{ chatChannelLabel }}</option>
                </select>
              </span>
              <input disabled placeholder="旧服聊天发送未接入" title="聊天发送接口未接入" aria-describedby="chat-send-readonly-note" />
              <button type="button" disabled title="聊天发送接口未接入" aria-describedby="chat-send-readonly-note">发送</button>
              <span id="chat-send-readonly-note" class="chat-dialog-readonly-note">聊天读取和发送接口尚未接入旧服，当前不显示聊天记录。</span>
            </div>
          </div>
        </div>

        <div v-if="chatControlDialogVisible" class="modal-layer">
          <div class="dialog-panel chat-dialog-panel" data-testid="chat-control-dialog">
            <button class="dialog-close" data-testid="chat-control-close" type="button" @click="chatControlDialogVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>聊天过滤预览</h2>
            <div class="chat-filter-list">
              <label v-for="channel in chatChannelItems" :key="channel">
                <input type="checkbox" checked disabled title="聊天读取接口未接入" aria-describedby="chat-control-readonly-note" />
                <span class="chat-filter-check" aria-hidden="true"></span>
                <span>{{ channel }}</span>
              </label>
              <div id="chat-control-readonly-note" class="chat-control-readonly-note">聊天读取接口未接入，过滤设置仅作预览。</div>
            </div>
            <div class="chat-filter-preview">
              <div class="chat-dialog-empty">聊天读取和发送接口尚未接入旧服，当前不显示聊天记录。</div>
            </div>
          </div>
        </div>
      </template>

      <div v-if="loading" class="loading-mask">处理中...</div>
      <div v-if="error" class="error-line">{{ error }}</div>
    </section>
  </main>
</template>

