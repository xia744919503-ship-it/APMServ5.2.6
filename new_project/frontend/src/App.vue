<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import {
  buildingInfo,
  buildingOptions,
  buildingSpeedGoods,
  cancelBuildingAction,
  cityDetail,
  cityHeroes,
  cityResearchSnapshot,
  createBuilding,
  createRole,
  currentUser,
  deleteMail,
  destroyBuilding,
  legacyActivities,
  legacyGuides,
  legacyLogin,
  mailDetail,
  mailPage,
  myCities,
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
  type Building,
  type BuildingInfoResult,
  type BuildingOption,
  type BuildingOptionsResult,
  type BuildingSpeedGoodsResult,
  type CityCard,
  type CityDetail,
  type CityListResult,
  type CityResearchSnapshot,
  type Guide,
  type Hero,
  type HeroRoster,
  type LoginResult,
  type MailItem,
  type MailPage,
  type RankItem,
  type ReportItem,
  type ReportPage,
  type SessionUser,
  type ShopItem,
  type ShopSnapshot,
  type SpeedGoods,
  type TechItem,
  type TroopType,
  type UnionInfo,
  type UserTypeGoodsItem,
  type WorldMap
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
  claimTaskReward,
  myTasks,
  type TaskCard,
  type TaskSnapshot
} from "./api";

type Screen = "boot" | "login" | "create-role" | "city";
type CityView = "inner" | "outer";
type LeftInfoTab = "resource" | "commander" | "army" | "defence";

const STAGE_WIDTH = 1000;
const STAGE_HEIGHT = 600;
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
const city = ref<CityDetail | null>(null);
const cityList = ref<CityCard[]>([]);
const showBuildingLevels = ref(false);
const showTopBuildingList = ref(false);
const activeBottomFunction = ref("");
const hoveredBottomFunction = ref("");
const chatChannelMenuVisible = ref(false);
const chatChannelLabel = ref("世界");
const chatChannelItems = ["世界", "联盟", "私聊", "战场"];
const hoveredTopButton = ref("");
const pressedTopButton = ref("");
const topButtonOverAssets = new Set(["innercity", "outercity", "map", "battle"]);
const topButtonHoverOnAssets = new Set(["hero", "army", "union", "mission"]);
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
const governmentFieldRows = computed(() => [
  { type: "农田", field: "资源田", level: "1", position: "(1,1)", state: "和平" },
  { type: "木场", field: "资源田", level: "1", position: "(1,2)", state: "采集" },
  { type: "石场", field: "资源田", level: "2", position: "(2,1)", state: "战乱" },
  { type: "铁矿", field: "资源田", level: "2", position: "(2,2)", state: "和平" }
]);

// Dialog refs
const mailPanelVisible = ref(false);
const mailFolder = ref<"inbox" | "sent" | "sys">("inbox");
const mailItems = ref<MailItem[]>([]);
const mailPageInfo = ref({ total: 0, page: 1, pageSize: 10 });
const mailDetailView = ref<MailItem | null>(null);

const reportPanelVisible = ref(false);
const reportFilter = ref("all");
const reportItems = ref<ReportItem[]>([]);
const reportPageInfo = ref({ total: 0, page: 1, pageSize: 10 });
const reportDetailView = ref<ReportItem | null>(null);

const shopPanelVisible = ref(false);
const shopSnapshot = ref<ShopSnapshot | null>(null);
const selectedShopGroupId = ref<number | null>(null);

const worldMapPanelVisible = ref(false);
const worldMapData = ref<WorldMap | null>(null);
const worldMapCenter = ref<WorldMap["center"] | null>(null);
const worldMapInputX = ref("");
const worldMapInputY = ref("");
const worldMapMode = ref<"city" | "map">("city");
const worldGridTip = ref({ visible: false, x: 0, y: 0, title: "", text: "", city: false });
const selectedWorldGrid = ref<{ x: number; y: number; title: string; text: string; city: boolean; empty: boolean } | null>(null);
const worldMapAlert = ref("");
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
const unionSnapshot = ref<{ union: UnionInfo | null; applyList: { id: number; name: string }[] } | null>(null);

const heroPanelVisible = ref(false);
const heroRoster = ref<HeroRoster | null>(null);

const barracksPanelVisible = ref(false);
const troopsData = ref<{ troops: TroopType[]; maxCapacity: number } | null>(null);
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

const rankPanelVisible = ref(false);
const rankKind = ref("power");
const rankItems = ref<RankItem[]>([]);
const rankPageInfo = ref({ total: 0, page: 1, pageSize: 20 });

const battlePanelVisible = ref(false);
const battleMenuVisible = ref(false);
const battleCampaignVisible = ref(false);
const battleInfoVisible = ref(false);
const battleUsersVisible = ref(false);
const battleInfoTab = ref<"info" | "news">("info");
const battleMapId = ref(1001);
const battleCampaignUseFlag = ref(false);
const battleCampaignTargetId = ref("1");
const battleCampaignHeroId = ref("0");
const battleCampaignFieldName = ref("前线战场");
const battleCampaignFoodCarry = ref("0");
const battleCampaignPathNeedTime = ref("0:30");
const battleCampaignArriveTime = ref("0:30");
const battleInfoPage = ref(1);
const battleNewsItems = ref([
  { time: "12:00", evtContent: "暂无战场新闻。", color: 16451364 }
]);
const battleCampaignSoldiers = ref([
  { sid: 1, name: "步兵", count: 12, takecount: 0 },
  { sid: 2, name: "骑兵", count: 8, takecount: 0 },
  { sid: 3, name: "弓兵", count: 18, takecount: 0 },
  { sid: 4, name: "器械", count: 4, takecount: 0 }
]);
const battleCampaignTargets = ref([
  { id: "1", name: "城池一" },
  { id: "2", name: "城池二" }
]);
const battleCampaignHeroes = ref([
  { id: "0", heroname: "请选择" },
  { id: "1", heroname: "赵云" }
]);
const battleInfoMeta = ref([
  { label: "战场名称", value: "城池争夺" },
  { label: "难度", value: "10" },
  { label: "战场兵力", value: "1~5" },
  { label: "参加人数", value: "1~5" }
]);
const battleNewsPageCount = computed(() => Math.max(1, Math.ceil(battleNewsItems.value.length / 10)));
const battleInfoNewsPageItems = computed(() => {
  const start = (battleInfoPage.value - 1) * 10;
  return battleNewsItems.value.slice(start, start + 10);
});
const battleInviteCreator = ref(true);
const battleInviteName = ref("");
const battleInviteCountText = ref("0/10");
const battleInviteUsers = ref([
  { id: 1, name: "暂无成员", camp: "-", state: "-", herocount: 0, honour: 0, cancel: false }
]);
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
  passport: localStorage.getItem("rxsg_passport") || `codex${Math.floor(Math.random() * 9000 + 1000)}`,
  password: ""
});

const roleForm = ref({
  userName: `lord${Math.floor(Math.random() * 900 + 100)}`,
  cityName: "新城池",
  flagChar: "H",
  sex: 0,
  face: 1
});

const debugParams = new URLSearchParams(window.location.search);
const debugScreen = debugParams.get("debugScreen");
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
const taskCategories = computed(() => taskSnapshot.value?.categories ?? []);
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
const shopGroups = computed(() => shopSnapshot.value?.groups ?? []);
const selectedShopGroup = computed(() => {
  const groups = shopGroups.value;
  if (groups.length === 0) return null;
  return groups.find((group) => group.id === selectedShopGroupId.value) ?? groups[0] ?? null;
});
const selectedShopItems = computed(() => selectedShopGroup.value?.items ?? []);
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
      return;
    }
    if (me.user.cityCount <= 0) {
      screen.value = "create-role";
      return;
    }
    await enterCity();
  });
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
      userName: `lord${Math.floor(Math.random() * 9000 + 1000)}`,
      cityName: `city${Math.floor(Math.random() * 9000 + 1000)}`,
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
  await enterCity(result.user.defaultCid);
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
      cityName: "新城池",
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

async function enterCity(cid?: number) {
  const cities = await myCities();
  cityList.value = cities.items;
  const target = cid ?? user.value?.defaultCid ?? cities.items[0]?.cid;
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
  const house = city.value?.buildings.find((building) => building.bid === 5);
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
  const requestSeq = ++buildRequestSeq;
  const guideLocked = effectiveGuideVisible.value && currentGuide.value?.distype === 1;
  const targetRect = guideRect.value;
  const currentGuideID = currentGuide.value?.gid ?? 0;
  const redirectedPlot =
    guideLocked && currentGuideID === 6 ? firstGuideBuildPlot() :
    guideLocked && currentGuideID === 8 ? builtHousePlot() :
    null;
  const activePlot = redirectedPlot ?? plot;
  const bypassGuideRect =
    currentGuideID === 6 ||
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
  await openGovernmentFromCityManage();
}

function toggleTopBuildingList() {
  showTopBuildingList.value = !showTopBuildingList.value;
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
    selected ? "_on" :
    pressedTopButton.value === name && hasDown ? "_down" :
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

function showMissingFlashDialog(name: string) {
  closeMainSceneOverlays();
  status(`${name}窗口还没有接入。`);
}

function selectChatChannel(channel: string) {
  chatChannelLabel.value = channel;
  chatChannelMenuVisible.value = false;
}

function handleChatSend() {
  status("聊天发送还没有接入。");
}

function handleChatControl() {
  status("聊天控制还没有接入。");
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

async function useGeneralGoods(item: UserTypeGoodsItem) {
  if (!city.value) return;
  if (item.count <= 0) {
    status(`${item.name}数量不足。`);
    return;
  }
  if (!window.confirm(`确定使用${item.name}?`)) return;
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

function plotIntersectsGuide(plot: Plot, rect: { x: number; y: number; w: number; h: number }) {
  return plot.x < rect.x + rect.w && plot.x + plot.w > rect.x && plot.y < rect.y + rect.h && plot.y + plot.h > rect.y;
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
  if (!window.confirm(`确定拆除${buildingPanel.value.name}(等级${buildingPanel.value.level})?`)) return;
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

async function selectTask(taskId: number) {
  selectedTaskId.value = taskId;
  if (currentGuide.value?.gid === 12) setGuide(13);
}

async function handleClaimReward() {
  const taskId = selectedTaskId.value;
  if (!taskId) return;
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
  if (!window.confirm(`确定使用${item.name}?`)) return;
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

function closeUtilityDialogs() {
  closeMailPanel();
  closeReportPanel();
  closeShopPanel();
  closeRankPanel();
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
  if (!window.confirm("确定删除此邮件?")) return;
  await withLoading(async () => {
    await deleteMail(id);
    mailItems.value = mailItems.value.filter((m) => m.id !== id);
    if (mailDetailView.value?.id === id) mailDetailView.value = null;
  });
}

// Report dialog
async function openReportPanel(filter = "all") {
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
  battleInfoVisible.value = false;
  battleUsersVisible.value = false;
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
  closeMainSceneOverlays();
  await withLoading(async () => {
    unionSnapshot.value = await myUnion();
    unionPanelVisible.value = true;
  });
}

// Hero dialog
async function openHeroPanel() {
  if (!city.value) return;
  closeMainSceneOverlays();
  await withLoading(async () => {
    heroRoster.value = await cityHeroes(city.value!.summary.cid, 100);
    heroPanelVisible.value = true;
  });
}

async function handleRecruitHero() {
  if (!city.value) return;
  await withLoading(async () => {
    const result = await recruitHero(city.value!.summary.cid);
    if (result.hero) {
      heroRoster.value = await cityHeroes(city.value!.summary.cid, 100);
    }
  });
}

// Barracks dialog
async function openBarracksPanel() {
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
async function openRankPanel(kind = "power") {
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
function openBattlePanel() {
  leaveGuideForManualNavigation();
  closeMainSceneOverlays();
  worldMapPanelVisible.value = false;
  battlePanelVisible.value = true;
}

function openBattleMenu() {
  battleMenuVisible.value = true;
}

function closeBattleMenu() {
  battleMenuVisible.value = false;
}

function openBattleCampaignDialog() {
  battleMenuVisible.value = false;
  battleCampaignVisible.value = true;
}

function openBattleInfoDialog() {
  battleMenuVisible.value = false;
  battleInfoVisible.value = true;
  battleInfoTab.value = "info";
  battleInfoPage.value = 1;
}

function openBattleUsersDialog() {
  battleMenuVisible.value = false;
  battleUsersVisible.value = true;
}

function openBattleTaskDialog() {
  battleMenuVisible.value = false;
  void openTaskPanel();
}

function quitBattleField() {
  battleMenuVisible.value = false;
  battlePanelVisible.value = false;
}

function closeBattleCampaignDialog() {
  battleCampaignVisible.value = false;
}

function closeBattleInfoDialog() {
  battleInfoVisible.value = false;
}

function closeBattleUsersDialog() {
  battleUsersVisible.value = false;
}

function inviteBattleUserLocal() {
  const name = battleInviteName.value.trim();
  if (!name) return;
  battleInviteUsers.value = battleInviteUsers.value.filter((item) => item.name !== "暂无成员");
  battleInviteUsers.value.push({
    id: Date.now(),
    name,
    camp: "我方",
    state: "已邀请",
    herocount: 0,
    honour: 0,
    cancel: true
  });
  battleInviteName.value = "";
  battleInviteCountText.value = `${battleInviteUsers.value.length}/10`;
}

function cancelBattleInviteLocal(id: number) {
  battleInviteUsers.value = battleInviteUsers.value.filter((item) => item.id !== id);
  if (battleInviteUsers.value.length === 0) {
    battleInviteUsers.value.push({ id: 1, name: "暂无成员", camp: "-", state: "-", herocount: 0, honour: 0, cancel: false });
  }
  const count = battleInviteUsers.value.filter((item) => item.name !== "暂无成员").length;
  battleInviteCountText.value = `${count}/10`;
}

function setBattleInfoTab(tab: "info" | "news") {
  battleInfoTab.value = tab;
}

function prevBattleInfoPage() {
  battleInfoPage.value = Math.max(1, battleInfoPage.value - 1);
}

function nextBattleInfoPage() {
  battleInfoPage.value = Math.min(battleNewsPageCount.value, battleInfoPage.value + 1);
}

function toggleBattleCampaignUseFlag() {
  battleCampaignUseFlag.value = !battleCampaignUseFlag.value;
}

function setBattleSoldierCount(index: number, delta: number) {
  const item = battleCampaignSoldiers.value[index];
  if (!item) return;
  item.takecount = Math.max(0, Math.min(item.count, item.takecount + delta));
}

function onBattleSoldierInput(index: number, value: string) {
  const item = battleCampaignSoldiers.value[index];
  if (!item) return;
  const parsed = Number.parseInt(value, 10);
  item.takecount = Number.isNaN(parsed) ? 0 : Math.max(0, Math.min(item.count, parsed));
}

function battleSoldierIcon(sid: number) {
  return asset(`images/army_${sid}.png`);
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
}

function takeNoBattleSoldiers() {
  for (const item of battleCampaignSoldiers.value) {
    item.takecount = 0;
  }
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
  const baseX = WORLD_GRID_CENTER_X + WORLD_GRID_OFFSET_X;
  const baseY = WORLD_GRID_CENTER_Y + WORLD_GRID_IMAGE_HEIGHT - WORLD_GRID_OFFSET_X;
  const x = Math.floor((2 * (localY - baseY) + (localX - baseX)) / (WORLD_GRID_OFFSET_X * 2)) + center.x;
  const y = Math.floor((2 * (localY - baseY) - (localX - baseX)) / (WORLD_GRID_OFFSET_X * 2)) + center.y;
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
      empty: false
    };
  }

  const type = worldTerrainType(x, y);
  if (type === 0) {
    return {
      title: `空地:(${x},${y})`,
      text: "这是一块尚未开发的空地。",
      city: false,
      empty: true
    };
  }

  const level = worldTerrainLevel(x, y);
  return {
    title: `${WORLD_TERRAIN_NAMES[type] ?? "野地"}:(${x},${y})  ${level}级`,
    text: worldFieldResourceText(type, level),
    city: false,
    empty: false
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

function selectWorldCity(cityItem: WorldMap["cities"][number]) {
  const point = flashCityPoint(cityItem);
  selectWorldGrid(point.x, point.y);
}

function flashCityPoint(cityItem: WorldMap["cities"][number]) {
  const cidX = cityItem.cid % 1000;
  const cidY = Math.floor(cityItem.cid / 1000);
  return {
    x: Number.isFinite(cidX) && cidX > 0 ? cidX : cityItem.x,
    y: Number.isFinite(cidY) && cidY > 0 ? cidY : cityItem.y
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
    left: `${point.x + 8}px`,
    top: `${point.y + 8}px`
  };
}

function worldMapCityLabelStyle(cityItem: WorldMap["cities"][number]) {
  const point = flashCityPoint(cityItem);
  const left = Math.min(Math.max(point.x + 9, 0), 472);
  const top = Math.min(Math.max(point.y + 9, 0), 500);
  return {
    left: `${left}px`,
    top: `${top}px`
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
          <button class="login-submit" type="submit" :disabled="loading">进入游戏</button>
        </form>
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
            <select v-model.number="selectedProvince">
              <option v-for="province in provinceAssets" :key="province.id" :value="province.id">
                {{ province.name }}
              </option>
            </select>
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
          <div class="province-info">{{ selectedProvinceName }}：请选择主公起兵之地。</div>
          <label class="role-rule">
            <input type="checkbox" checked />
            <span>我已经阅读并同意用户协议</span>
          </label>
          <button class="start-role" type="button" :disabled="loading" @click="submitRole">开始游戏</button>
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
          <button class="top-tab-btn city-tab" type="button" @mouseenter="hoveredTopButton = 'innercity'" @mouseleave="hoveredTopButton = ''; pressedTopButton = ''" @pointerdown="pressedTopButton = 'innercity'; openCityScene('inner')" @pointerup="pressedTopButton = ''">
            <img :src="topButtonImage('innercity', cityView === 'inner' && !worldMapPanelVisible && !battlePanelVisible, false)" alt="城内" />
          </button>
          <button class="top-tab-btn city-tab" type="button" @mouseenter="hoveredTopButton = 'outercity'" @mouseleave="hoveredTopButton = ''; pressedTopButton = ''" @pointerdown="pressedTopButton = 'outercity'; openCityScene('outer')" @pointerup="pressedTopButton = ''">
            <img :src="topButtonImage('outercity', cityView === 'outer' && !worldMapPanelVisible && !battlePanelVisible, false)" alt="城池" />
          </button>
          <button class="top-tab-btn city-tab" type="button" @mouseenter="hoveredTopButton = 'map'" @mouseleave="hoveredTopButton = ''; pressedTopButton = ''" @pointerdown="pressedTopButton = 'map'; openWorldMapPanel()" @pointerup="pressedTopButton = ''">
            <img :src="topButtonImage('map', worldMapPanelVisible, false)" alt="地图" />
          </button>
          <button class="top-tab-btn city-tab" type="button" @mouseenter="hoveredTopButton = 'battle'" @mouseleave="hoveredTopButton = ''; pressedTopButton = ''" @pointerdown="pressedTopButton = 'battle'; openBattlePanel()" @pointerup="pressedTopButton = ''">
            <img :src="topButtonImage('battle', battlePanelVisible, false)" alt="战场" />
          </button>
        </div>
        <div class="top-mini-functions">
          <button
            class="mini-function-btn build"
            :class="{ selected: showTopBuildingList }"
            type="button"
            @click="toggleTopBuildingList"
          ></button>
          <button class="mini-function-btn craft" type="button" @click="showMissingFlashDialog('工匠')"></button>
          <button class="mini-function-btn tactic" type="button" @click="showMissingFlashDialog('计谋')"></button>
          <button
            class="mini-function-btn level"
            :class="{ selected: showBuildingLevels }"
            type="button"
            @click="showBuildingLevels = !showBuildingLevels"
          ></button>
        </div>
        <div v-if="showTopBuildingList" class="top-building-list-panel">
          <template v-if="topBuildingQueueItems.length > 0">
            <button
              v-for="item in topBuildingQueueItems"
              :key="`top-building-${item.position}-${item.state}`"
              class="top-building-list-row"
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
        <div class="top-actions">
          <button class="top-tab-btn func-tab" type="button" @mouseenter="hoveredTopButton = 'hero'" @mouseleave="hoveredTopButton = ''; pressedTopButton = ''" @pointerdown="pressedTopButton = 'hero'" @pointerup="pressedTopButton = ''" @click="openHeroPanel"><img :src="topButtonImage('hero', heroPanelVisible)" alt="将领" /></button>
          <button class="top-tab-btn func-tab" type="button" @mouseenter="hoveredTopButton = 'army'" @mouseleave="hoveredTopButton = ''; pressedTopButton = ''" @pointerdown="pressedTopButton = 'army'" @pointerup="pressedTopButton = ''" @click="openBarracksPanel"><img :src="topButtonImage('army', barracksPanelVisible)" alt="军队" /></button>
          <button class="top-tab-btn func-tab" type="button" @mouseenter="hoveredTopButton = 'union'" @mouseleave="hoveredTopButton = ''; pressedTopButton = ''" @pointerdown="pressedTopButton = 'union'" @pointerup="pressedTopButton = ''" @click="openUnionPanel"><img :src="topButtonImage('union', unionPanelVisible)" alt="联盟" /></button>
          <button class="top-tab-btn func-tab" type="button" @mouseenter="hoveredTopButton = 'mission'" @mouseleave="hoveredTopButton = ''; pressedTopButton = ''" @pointerdown="pressedTopButton = 'mission'" @pointerup="pressedTopButton = ''" @click="openTaskPanel"><img :src="topButtonImage('mission', taskPanelVisible)" alt="任务" /></button>
        </div>
        <div class="left-data">
          <div class="left-bottom-panel" aria-hidden="true"></div>
          <div class="left-view-tabs">
            <button
              class="left-vert-btn"
              :class="{ selected: leftInfoTab === 'resource' }"
              type="button"
              @mousedown.prevent="setLeftInfoTab('resource')"
              @keydown.enter.prevent="setLeftInfoTab('resource')"
              @keydown.space.prevent="setLeftInfoTab('resource')"
            >
              资源
            </button>
            <button
              class="left-vert-btn"
              :class="{ selected: leftInfoTab === 'commander' }"
              type="button"
              @mousedown.prevent="setLeftInfoTab('commander')"
              @keydown.enter.prevent="setLeftInfoTab('commander')"
              @keydown.space.prevent="setLeftInfoTab('commander')"
            >
              将领
            </button>
            <button
              class="left-vert-btn"
              :class="{ selected: leftInfoTab === 'army' }"
              type="button"
              @mousedown.prevent="setLeftInfoTab('army')"
              @keydown.enter.prevent="setLeftInfoTab('army')"
              @keydown.space.prevent="setLeftInfoTab('army')"
            >
              军队
            </button>
            <button
              class="left-vert-btn"
              :class="{ selected: leftInfoTab === 'defence' }"
              type="button"
              @mousedown.prevent="setLeftInfoTab('defence')"
              @keydown.enter.prevent="setLeftInfoTab('defence')"
              @keydown.space.prevent="setLeftInfoTab('defence')"
            >
              城防
            </button>
          </div>
          <div class="left-rank-strip">声望 {{ userPrestige }}&nbsp;&nbsp;排名 {{ userRank }}</div>
          <div class="left-hero-panel">
            <div class="lord-portrait-frame">
              <img class="lord-portrait" :src="faceImage" alt="" @click="showMissingFlashDialog('君主信息')" />
            </div>
            <div class="lord-meta-panel">
              <div class="lord-meta-row row-king"><span @click="showMissingFlashDialog('君主信息')">君主:</span><strong>{{ user?.name || city.summary.owner }}</strong></div>
              <div class="lord-meta-row row-prestige"><span @click="openRankPanel()">声望:</span><strong>{{ userPrestige }}</strong></div>
              <div class="lord-meta-row row-rank"><span @click="openRankPanel()">排名:</span><strong @click="openRankPanel()">{{ userRank }}</strong></div>
              <div class="lord-meta-row row-office"><span @click="openTaskPanel">官职:</span><strong>{{ userOffice }}</strong></div>
              <div class="lord-meta-row row-title"><span @click="openTaskPanel">爵位:</span><strong>{{ userNobility }}</strong></div>
              <div class="lord-meta-row row-union"><span @click="openUnionPanel">联盟:</span><strong>{{ userUnionName }}</strong></div>
              <div class="lord-meta-row row-post"><span @click="openUnionPanel">职位:</span><strong>{{ userUnionPosition }}</strong></div>
            </div>
            <div class="left-action-column">
              <button class="playerinfo-btn king" type="button" @click="showMissingFlashDialog('君主信息')">君主</button>
              <button class="playerinfo-btn armor" type="button" @click="showMissingFlashDialog('装备')">装备</button>
              <button class="playerinfo-btn inventory" type="button" @click="showMissingFlashDialog('宝物')">宝物</button>
            </div>
          </div>
          <div class="city-summary-strip">
            <img class="summary-real-icon morale-icon" :src="asset('city_popularity.png')" alt="" @mouseenter="showMoraleTip" @mouseleave="hideLeftInfoTip" @click="openUseGoodsDialog(4)" />
            <span class="summary-value morale-value" @mouseenter="showFlashToolTip($event, '民心')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ city.morale }}</span>
            <span class="summary-slash">/</span>
            <span class="summary-value complaint-value" @mouseenter="showFlashToolTip($event, '民怨')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ city.complaint }}</span>
            <img class="summary-real-icon tax-icon" :src="asset('city_tax.png')" alt="" @mouseenter="showFlashToolTip($event, '税率')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openLeftTax" />
            <span class="summary-value tax-value" @mouseenter="showFlashToolTip($event, '税率')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ city.tax }}%</span>
            <img class="summary-real-icon gold-icon" :src="asset('city_gold.png')" alt="" @mouseenter="showGoldTip" @mouseleave="hideLeftInfoTip" @click="openUseGoodsDialog(5)" />
            <span class="summary-value gold-value" @mouseenter="showFlashToolTip($event, '黄金数量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashRoundedInteger(resources?.gold) }}</span>
            <span class="summary-value gold-add" :class="{ negative: leftGoldAdd < 0 }" @mouseenter="showFlashToolTip($event, '黄金产量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(leftGoldAdd) }}</span>
            <img class="summary-real-icon people-icon" :src="asset('city_population.png')" alt="" @mouseenter="showPeopleTip" @mouseleave="hideLeftInfoTip" @click="openUseGoodsDialog(6)" />
            <span class="summary-value people-value" @mouseenter="showFlashToolTip($event, '当前人口')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(resources?.people) }}</span>
            <span class="summary-value idle-value" :class="{ negative: leftIdlePeople < 0 }" @mouseenter="showFlashToolTip($event, '空闲人口')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(leftIdlePeople) }}</span>
            <button class="left-plus-btn morale-plus" type="button" @mouseenter="showFlashToolTip($event, '使用宝物提高民心，消除民怨')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openUseGoodsDialog(4)"></button>
            <button class="left-plus-btn gold-plus" type="button" @mouseenter="showFlashToolTip($event, '使用宝物增加黄金税收')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openUseGoodsDialog(5)"></button>
            <button class="left-plus-btn people-plus" type="button" @mouseenter="showFlashToolTip($event, '使用宝物增加人口')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openUseGoodsDialog(6)"></button>
            <button class="mycity-btn mycity-building" type="button" @mouseenter="showFlashToolTip($event, '建筑信息')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openLeftBuildingList"></button>
            <button class="mycity-btn mycity-labor" type="button" @mouseenter="showFlashToolTip($event, '资源生产')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openLeftResourceProduce"></button>
            <button class="mycity-btn mycity-field" type="button" @mouseenter="showFlashToolTip($event, '附属野地')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openLeftCityFields"></button>
          </div>
          <div class="left-city-header">
            <img class="left-city-title" :src="asset('title.png')" alt="" />
            <div class="flash-city-combo">
              <button class="city-list" :class="{ open: cityDropdownOpen }" type="button" @click="toggleCityDropdown">
                {{ selectedCityLabel }}
              </button>
              <div v-if="cityDropdownOpen" class="city-list-menu">
                <button
                  v-for="item in cityList.slice(0, 10)"
                  :key="item.cid"
                  class="city-list-option"
                  :class="{ selected: item.cid === selectedCityId }"
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
              <div class="resource-icon-box food"><img :src="asset('resource_food.png')" alt="" @mouseenter="showFoodTip" @mouseleave="hideLeftInfoTip" @click="openUseGoodsDialog(7)" /></div>
              <div class="resource-icon-box wood"><img :src="asset('resource_wood.png')" alt="" @mouseenter="showResourceTip('wood')" @mouseleave="hideLeftInfoTip" @click="openUseGoodsDialog(8)" /></div>
              <div class="resource-icon-box rock"><img :src="asset('resource_rock.png')" alt="" @mouseenter="showResourceTip('rock')" @mouseleave="hideLeftInfoTip" @click="openUseGoodsDialog(9)" /></div>
              <div class="resource-icon-box iron"><img :src="asset('resource_iron.png')" alt="" @mouseenter="showResourceTip('iron')" @mouseleave="hideLeftInfoTip" @click="openUseGoodsDialog(10)" /></div>
              <div class="resource-value-box food" @mouseenter="showFlashToolTip($event, '粮食数量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(leftFoodValue) }}</div>
              <div class="resource-value-box wood" @mouseenter="showFlashToolTip($event, '木材数量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(leftWoodValue) }}</div>
              <div class="resource-value-box rock" @mouseenter="showFlashToolTip($event, '石料数量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(leftRockValue) }}</div>
              <div class="resource-value-box iron" @mouseenter="showFlashToolTip($event, '铁锭数量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(leftIronValue) }}</div>
              <div class="resource-add-box food" :class="{ negative: leftFoodAdd < 0 }" @mouseenter="showFlashToolTip($event, '粮食产量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(leftFoodAdd) }}</div>
              <div class="resource-add-box wood" @mouseenter="showFlashToolTip($event, '木材产量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(production?.woodAdd) }}</div>
              <div class="resource-add-box rock" @mouseenter="showFlashToolTip($event, '石料产量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(production?.rockAdd) }}</div>
              <div class="resource-add-box iron" @mouseenter="showFlashToolTip($event, '铁锭产量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip">{{ formatFlashInteger(production?.ironAdd) }}</div>
              <button class="left-plus-btn resource-plus food" type="button" @mouseenter="showFlashToolTip($event, '使用宝物增加粮食产量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openUseGoodsDialog(7)"></button>
              <button class="left-plus-btn resource-plus wood" type="button" @mouseenter="showFlashToolTip($event, '使用宝物增加木材产量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openUseGoodsDialog(8)"></button>
              <button class="left-plus-btn resource-plus rock" type="button" @mouseenter="showFlashToolTip($event, '使用宝物增加石料产量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openUseGoodsDialog(9)"></button>
              <button class="left-plus-btn resource-plus iron" type="button" @mouseenter="showFlashToolTip($event, '使用宝物增加铁锭产量')" @mousemove="moveFlashToolTip" @mouseleave="hideFlashToolTip" @click="openUseGoodsDialog(10)"></button>
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
            <button
              class="cityhall-hit"
              data-testid="cityhall-hit"
              :class="{
                occupied: isOccupied(cityHallPlot.position),
                busy: isBusy(positionBuilding(cityHallPlot.position)),
                hovered: hoveredInnerPlot?.position === cityHallPlot.position,
                selected: selectedPlot?.position === cityHallPlot.position
              }"
              :style="{
                left: `${cityHallPlot.x}px`,
                top: `${cityHallPlot.y}px`,
                width: `${cityHallPlot.w}px`,
                height: `${cityHallPlot.h}px`
              }"
              type="button"
            >
              <img class="cityhall-img" :src="asset('building_cityhall.png')" alt="" />
              <img
                v-if="showBuildingLevels && positionBuilding(cityHallPlot.position)"
                class="building-level-img cityhall-level"
                :src="buildingLevelImage(positionBuilding(cityHallPlot.position))"
                alt=""
              />
            </button>
            <button
              v-for="plot in innerPlots"
              :key="plot.position"
              class="plot-hit"
              data-testid="inner-plot-hit"
              :class="{
                occupied: isOccupied(plot.position),
                busy: isBusy(positionBuilding(plot.position)),
                hovered: hoveredInnerPlot?.position === plot.position,
                selected: selectedPlot?.position === plot.position
              }"
              :style="{ left: `${plot.x}px`, top: `${plot.y}px`, width: `${plot.w}px`, height: `${plot.h}px` }"
              type="button"
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
            </button>
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
            <div class="outer-citywall" aria-hidden="true">
              <img :src="asset(positionBuilding(cityWallPlot.position) ? 'building_outercity.png' : 'building_outertown.png')" alt="" />
            </div>
            <button
              v-for="plot in visibleOuterPlots"
              :key="`outer-${plot.position}`"
              class="outer-plot-hit"
              :class="{ occupied: isOccupied(plot.position), busy: isBusy(positionBuilding(plot.position)) }"
              :style="{ left: `${plot.x}px`, top: `${plot.y}px`, width: `${plot.w}px`, height: `${plot.h}px` }"
              type="button"
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
            </button>
          </template>
        </div>
        <div class="bottom-chat">
          <div class="flash-chat-shell">
            <div class="chat-console">
              <div class="chat-entry">
                <button class="chat-channel-btn" type="button" @click="chatChannelMenuVisible = !chatChannelMenuVisible">{{ chatChannelLabel }}</button>
                <input class="chat-input" type="text" value="" readonly />
                <button class="chat-send-btn" type="button" aria-label="发送" @click="handleChatSend"></button>
                <button class="chat-control-btn first" type="button" aria-label="聊天控制" @click="handleChatControl"></button>
                <div v-if="chatChannelMenuVisible" class="chat-channel-menu">
                  <button
                    v-for="channel in chatChannelItems"
                    :key="channel"
                    class="chat-channel-option"
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
              <button class="function-btn friend" type="button" @pointerenter="hoveredBottomFunction = 'friend'" @pointerdown="activeBottomFunction = 'friend'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="showMissingFlashDialog('好友')"><img :src="bottomFunctionImage('friend')" alt="好友" /></button>
              <button class="function-btn report" type="button" @pointerenter="hoveredBottomFunction = 'report'" @pointerdown="activeBottomFunction = 'report'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="openReportPanel()"><img :src="bottomFunctionImage('report')" alt="报告" /></button>
              <button class="function-btn mail" type="button" @pointerenter="hoveredBottomFunction = 'mail'" @pointerdown="activeBottomFunction = 'mail'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="openMailPanel('inbox')"><img :src="bottomFunctionImage('mail')" alt="信件" /></button>
              <button class="function-btn rank" type="button" @pointerenter="hoveredBottomFunction = 'rank'" @pointerdown="activeBottomFunction = 'rank'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="openRankPanel()"><img :src="bottomFunctionImage('rank')" alt="排行" /></button>
              <button class="function-btn stat" type="button" @pointerenter="hoveredBottomFunction = 'stat'" @pointerdown="activeBottomFunction = 'stat'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="showMissingFlashDialog('统计')"><img :src="bottomFunctionImage('stat')" alt="统计" /></button>
              <button class="function-btn forum" type="button" @pointerenter="hoveredBottomFunction = 'forum'" @pointerdown="activeBottomFunction = 'forum'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="showMissingFlashDialog('论坛')"><img :src="bottomFunctionImage('forum')" alt="论坛" /></button>
              <button class="function-btn website" type="button" @pointerenter="hoveredBottomFunction = 'website'" @pointerdown="activeBottomFunction = 'website'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="showMissingFlashDialog('官网')"><img :src="bottomFunctionImage('website')" alt="官网" /></button>
              <button class="function-btn help" type="button" @pointerenter="hoveredBottomFunction = 'help'" @pointerdown="activeBottomFunction = 'help'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="showMissingFlashDialog('帮助')"><img :src="bottomFunctionImage('help')" alt="帮助" /></button>
              <button class="function-btn charge" type="button" @pointerenter="hoveredBottomFunction = 'charge'" @pointerdown="activeBottomFunction = 'charge'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="showMissingFlashDialog('充值')"><img :src="bottomFunctionImage('charge')" alt="充值" /></button>
              <button class="function-btn shop" type="button" @pointerenter="hoveredBottomFunction = 'shop'" @pointerdown="activeBottomFunction = 'shop'" @pointerup="activeBottomFunction = ''" @pointerleave="hoveredBottomFunction = ''; activeBottomFunction = ''" @click="openShopPanel()"><img :src="bottomFunctionImage('shop')" alt="商城" /></button>
            </div>
          </div>
        </div>
        <div v-if="effectiveGuideVisible" class="guide-layer">
          <button
            v-if="guideRect"
            class="guide-hotspot"
            type="button"
            :style="{
              left: `${guideRect.x}px`,
              top: `${guideRect.y}px`,
              width: `${guideRect.w}px`,
              height: `${guideRect.h}px`
            }"
            @click="handleGuideHotspotClick"
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

        <div v-if="announcementVisible" class="modal-layer announcement-layer">
          <div class="announcement-dialog">
            <img class="announcement-title-img" :src="asset('title.png')" alt="" />
            <h2>最新公告</h2>
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
                  <button type="button">条件不足</button>
                </div>
                <div class="announcement-line line-1"></div>
                <div class="announcement-reward-row row-2 alt">
                  <span class="announcement-reward-icon"></span>
                  <span class="announcement-reward-content">需要&lt;5元宝&gt;</span>
                  <button type="button">条件不足</button>
                </div>
                <div class="announcement-line line-2"></div>
                <div class="announcement-reward-row row-3">
                  <span class="announcement-reward-icon"></span>
                  <span class="announcement-reward-content special">需要&lt;5元宝&gt;</span>
                  <button type="button">条件不足</button>
                </div>
                <div class="announcement-line line-3"></div>
                <div class="announcement-reward-row row-4 alt">
                  <span class="announcement-reward-icon"></span>
                  <span class="announcement-reward-content special">充值后可免费领取</span>
                  <button type="button">条件不足</button>
                </div>
                <div class="announcement-line line-4"></div>
                <div class="announcement-reward-row row-5">
                  <span class="announcement-reward-icon hidden"></span>
                  <span class="announcement-reward-content special"></span>
                  <button type="button"></button>
                </div>
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
            <button class="announcement-close" type="button" @click="announcementVisible = false">关闭</button>
          </div>
        </div>

        <div v-if="buildingPanel && selectedPlot" class="modal-layer">
          <div class="building-panel" :class="{ 'government-building-panel': buildingPanel.bid === 6 }">
            <button class="building-close" type="button" @click="closeBuild">关闭</button>
            <img class="building-title-img" :src="asset('title.png')" alt="" />
            <h2>{{ buildingPanel.name }}(等级{{ displayBuildingLevel(buildingPanel) }})</h2>
            <div class="building-item">
              <div class="building-icon-frame">
                <img :src="buildingDialogImage(buildingPanel)" :alt="buildingPanel.name" />
              </div>
              <button
                v-if="buildingPanel.state === 0"
                class="building-upgrade-btn"
                type="button"
                @click="upgradeSelectedBuilding"
              >
                升级
              </button>
              <button
                v-if="buildingPanel.state === 0"
                class="building-destroy-btn"
                type="button"
                @click="requestDestroySelectedBuilding"
              >
                拆除
              </button>
              <button
                v-if="buildingPanel.state !== 0"
                class="building-speed-btn"
                type="button"
                @click="requestSpeedSelectedBuilding"
              >
                加速
              </button>
              <button
                v-if="buildingPanel.state !== 0"
                class="building-cancel-btn"
                type="button"
                @click="cancelSelectedBuildingAction"
              >
                取消
              </button>
              <div class="building-info-board">
                <div class="building-description">
                  {{ buildingDescriptionText }}
                </div>
                <div class="building-state">
                  {{ buildingNeedText || buildingStateText }}
                </div>
              </div>
            </div>
            <div v-if="buildingPanel.bid === 6" class="government-panel">
              <div class="government-actions">
                <button type="button">建筑总览</button>
                <button type="button" @click="openGovernmentSubDialog('tax')">调整税率</button>
                <button type="button" @click="openGovernmentSubDialog('pacify')">安抚百姓</button>
                <button type="button" @click="openGovernmentSubDialog('levy')">征收物资</button>
                <button type="button" @click="openGovernmentSubDialog('produce')">资源生产</button>
                <button type="button" @click="openGovernmentSubDialog('name')">城池改名</button>
                <button type="button" @click="openGovernmentSubDialog('fields')">附属野地</button>
                <button type="button" @click="openGovernmentSubDialog('cities')">所有城池</button>
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

              <div v-if="governmentSubDialog === 'name'" class="government-popup change-name-popup">
                <button class="popup-close" type="button" @click="closeGovernmentSubDialog">关闭</button>
                <button class="popup-confirm name-confirm" type="button" @click="closeGovernmentSubDialog">确定</button>
                <div class="popup-title">城池改名</div>
                <div class="popup-label name-label-current">城池</div>
                <div class="popup-value name-value-current">{{ city?.summary.name }}</div>
                <div class="popup-label name-label-new">新名称</div>
                <div class="popup-value name-input-wrap">
                  <input v-model="governmentCityName" maxlength="20" />
                </div>
                <div class="name-tip">每天只能更改一次城池名，请慎重考虑！</div>
              </div>

              <div v-if="governmentSubDialog === 'pacify'" class="government-popup pacify-popup">
                <button class="popup-close" type="button" @click="closeGovernmentSubDialog">关闭</button>
                <button class="popup-confirm pacify-confirm" type="button" @click="closeGovernmentSubDialog">确定</button>
                <div class="popup-title">安抚百姓</div>
                <div class="popup-label pacify-label-people">人口</div>
                <div class="popup-value pacify-value-people">{{ formatNumber(cityPopulation) }}</div>
                <div class="popup-label pacify-label-morale">民心</div>
                <div class="popup-value pacify-value-morale">{{ city?.morale ?? 0 }}</div>
                <div class="popup-label pacify-label-action">行为</div>
                <select v-model="governmentPacifyAction" class="pacify-select">
                  <option>赈灾</option>
                  <option>祈福</option>
                  <option>祭天</option>
                  <option>增丁</option>
                </select>
                <textarea class="pacify-preview" readonly :value="governmentPacifyText"></textarea>
              </div>

              <div v-if="governmentSubDialog === 'levy'" class="government-popup levy-popup">
                <button class="popup-close" type="button" @click="closeGovernmentSubDialog">关闭</button>
                <button class="popup-confirm levy-confirm" type="button" @click="closeGovernmentSubDialog">确定</button>
                <div class="popup-title">征收物资</div>
                <div class="popup-label levy-label-people">人口</div>
                <div class="popup-value levy-value-people">{{ formatNumber(cityPopulation) }}</div>
                <div class="popup-label levy-label-morale">民心</div>
                <div class="popup-value levy-value-morale">{{ city?.morale ?? 0 }}</div>
                <div class="popup-label levy-label-resource">征收</div>
                <select v-model="governmentLevyResource" class="levy-select">
                  <option>黄金</option>
                  <option>粮食</option>
                  <option>木材</option>
                  <option>石料</option>
                  <option>铁锭</option>
                </select>
                <div class="levy-preview">{{ governmentLevyPreview }}</div>
                <textarea class="levy-textarea" readonly></textarea>
              </div>

              <div v-if="governmentSubDialog === 'produce'" class="government-full-dialog produce-dialog-panel">
                <button class="dialog-close government-full-close" type="button" @click="closeGovernmentSubDialog">关闭</button>
                <div class="government-full-title wide">资源生产</div>
                <div class="government-list-panel produce-board-panel">
                  <div class="produce-head">
                    <span>资源</span>
                    <span>比例</span>
                    <span>存量</span>
                    <span>修改</span>
                  </div>
                  <div v-for="row in governmentProduceRows" :key="row.label" class="produce-row">
                    <span>{{ row.label }}</span>
                    <span><input type="number" :value="row.rate" readonly /></span>
                    <span>{{ formatNumber(row.value) }}</span>
                    <span><button type="button" class="city-list-action">修改</button></span>
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
                <button class="government-return-btn produce-confirm-btn" type="button" @click="closeGovernmentSubDialog">修改</button>
              </div>

              <div v-if="governmentSubDialog === 'fields'" class="government-full-dialog city-field-dialog-panel">
                <button class="dialog-close government-full-close field-close" type="button" @click="closeGovernmentSubDialog">关闭</button>
                <div class="government-full-title">附属野地</div>
                <div class="government-list-panel field-list-panel">
                  <div class="city-field-head">
                    <span>野地</span>
                    <span>位置</span>
                    <span>等级</span>
                    <span>状态</span>
                    <span>查看</span>
                  </div>
                  <div v-for="(row, index) in governmentFieldRows" :key="index" class="city-field-row">
                    <span>{{ row.field }}</span>
                    <span>{{ row.position }}</span>
                    <span>{{ row.level }}</span>
                    <span>{{ row.state }}</span>
                    <span><button type="button" class="city-field-action">查看</button></span>
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
                  <div v-for="item in governmentCityItems" :key="item.cid" class="city-list-row">
                    <span>{{ item.name }}</span>
                    <span>{{ formatCityCode(item.cid) }}</span>
                    <span>{{ getCityChiefName(item) }}</span>
                    <span>{{ formatNumber(item.resources.people) }}</span>
                    <span>{{ cityMorale(item) }}</span>
                    <span>
                      <button type="button" class="city-list-action">查看</button>
                    </span>
                  </div>
                  <div v-if="governmentCityItems.length === 0" class="government-empty list-empty">暂无城池数据。</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-if="buildPanel && selectedPlot" class="modal-layer">
          <div class="build-panel">
            <button class="build-close" type="button" @click="closeBuild">关闭</button>
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
                :class="{ active: selectedBid === option.bid }"
                :aria-disabled="!option.canBuild"
                @click="selectBuildOption(option)"
              >
                <img :src="buildingIntro(option)" alt="" />
                <button
                  class="create-building-btn"
                  data-testid="create-building-btn"
                  type="button"
                  :disabled="!option.canBuild"
                  @click.stop="confirmBuild(option)"
                >建造</button>
                <strong>{{ option.name }}</strong>
                <span>{{ buildOptionDescription(option) }}</span>
                <small>{{ option.canBuild ? `建造时间 ${formatDuration(option.duration)}` : option.reason }}</small>
              </div>
            </div>
          </div>
        </div>

        <div v-if="speedGoodsPanel" class="speed-goods-layer">
          <div class="speed-goods-dialog">
            <button class="speed-goods-close" type="button" @click="speedGoodsPanel = null">关闭</button>
            <img class="speed-goods-title-img" :src="asset('title.png')" alt="" />
            <h2>使用宝物</h2>
            <div class="speed-goods-list">
              <button
                v-for="item in speedGoodsPanel.goodsList"
                :key="item.gid"
                class="speed-goods-item"
                :disabled="item.count <= 0"
                type="button"
                @click.stop="useSpeedGoods(item)"
              >
                <span class="speed-goods-frame">
                  <img :src="speedGoodsImage(item)" :alt="item.name" @error="($event.target as HTMLImageElement).src = asset('board_listdown.png')" />
                  <span class="speed-goods-count">x{{ item.count }}</span>
                </span>
                <span class="speed-goods-name">{{ item.name }}</span>
                <span class="speed-goods-effect">{{ speedGoodsEffect(item) }}</span>
              </button>
            </div>
            <button class="speed-goods-buy" type="button">购买</button>
          </div>
        </div>

        <div v-if="useGoodsPanel.visible" class="use-goods-layer">
          <div class="use-goods-dialog" :style="useGoodsDialogStyle">
            <button class="use-goods-close" type="button" @click="closeUseGoodsDialog">关闭</button>
            <img class="use-goods-title-img" :src="asset('title.png')" alt="" />
            <h2>使用宝物</h2>
            <div class="use-goods-list">
              <div v-if="useGoodsPanel.loading" class="use-goods-message">加载中...</div>
              <div v-else-if="useGoodsPanel.error" class="use-goods-message">{{ useGoodsPanel.error }}</div>
              <div v-else-if="useGoodsPanel.goodsList.length === 0" class="use-goods-message">暂无可用宝物</div>
              <template v-else>
                <button
                  v-for="item in useGoodsPanel.goodsList"
                  :key="item.gid"
                  class="use-goods-item"
                  :disabled="item.count <= 0"
                  type="button"
                  :title="item.description"
                  @click.stop="useGeneralGoods(item)"
                >
                  <span class="use-goods-frame">
                    <img :src="useGoodsImage(item)" :alt="item.name" @error="($event.target as HTMLImageElement).src = asset('frame_pic.png')" />
                    <span class="use-goods-count">x{{ item.count }}</span>
                  </span>
                  <span class="use-goods-name">{{ item.name }}</span>
                </button>
              </template>
            </div>
            <button class="use-goods-buy" type="button" @click="openUseGoodsShopPanel">购买</button>
          </div>
        </div>

        <div v-if="taskPanelVisible" class="modal-layer task-layer" :class="{ 'battle-overlay-layer': battlePanelVisible }">
          <div class="task-dialog">
            <button class="task-close" type="button" @click="taskPanelVisible = false; selectedTaskId = null; selectedTaskGroupId = null; selectedTaskCategoryType = null">关闭</button>
            <img class="task-title-img" :src="asset('title.png')" alt="" />
            <h2>任务</h2>
            <div class="task-category-tabs">
              <button
                v-for="(category, index) in taskCategories.slice(0, 6)"
                :key="category.type"
                class="task-category-tab"
                :class="[`kind-${index}`, { active: selectedTaskCategoryType === category.type }]"
                type="button"
                @click="selectedTaskCategoryType = category.type; selectedTaskGroupId = category.groups[0]?.id ?? null; selectedTaskId = category.groups[0]?.tasks[0]?.id ?? null"
              >
                {{ category.label }}
              </button>
            </div>
            <div class="task-groups">
              <button
                v-for="group in selectedTaskGroups"
                :key="group.id"
                class="task-group-item"
                :class="{ active: selectedTaskGroupId === group.id }"
                type="button"
                @click="selectedTaskGroupId = group.id; selectedTaskId = group.tasks[0]?.id ?? null"
              >
                <span class="task-group-name">{{ group.name }}</span>
                <span class="task-group-progress">{{ group.completed }}/{{ group.total }}</span>
              </button>
            </div>
            <div class="task-list">
              <button
                v-for="task in selectedTaskGroup?.tasks ?? []"
                :key="task.id"
                class="task-item"
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
              type="button"
              @click="handleClaimReward"
            >
              领取奖励
            </button>
          </div>
        </div>

        <!-- Mail Dialog -->
        <div v-if="mailPanelVisible" class="modal-layer">
          <div class="dialog-panel mail-dialog">
            <button class="dialog-close" type="button" @click="mailPanelVisible = false; mailDetailView = null">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>邮件</h2>
            <div class="mail-tabs">
              <button :class="{ active: mailFolder === 'inbox' }" type="button" @click="openMailPanel('inbox')">收件箱</button>
              <button :class="{ active: mailFolder === 'sent' }" type="button" @click="openMailPanel('sent')">发件箱</button>
              <button :class="{ active: mailFolder === 'sys' }" type="button" @click="openMailPanel('sys')">系统邮件</button>
            </div>
            <div v-if="mailDetailView" class="mail-detail">
              <h3>{{ mailDetailView.title }}</h3>
              <div class="mail-meta">
                <span>{{ mailDetailView.type === 3 ? '系统' : (mailDetailView.type === 2 ? '收件人' : '发件人') }}: {{ mailDetailView.type === 3 ? '' : (mailDetailView.type === 2 ? mailDetailView.to : mailDetailView.from) }}</span>
                <span>{{ new Date(mailDetailView.sendTime * 1000).toLocaleString('zh-CN') }}</span>
              </div>
              <p class="mail-content">{{ mailDetailView.content }}</p>
              <button type="button" @click="mailDetailView = null">返回列表</button>
              <button v-if="mailDetailView.type !== 3" type="button" @click="deleteMailItem(mailDetailView.id)">删除</button>
            </div>
            <div v-else class="mail-list">
              <div v-if="mailItems.length === 0" class="empty-list">暂无邮件</div>
              <button
                v-for="item in mailItems"
                :key="item.id"
                class="mail-item"
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
            <button class="dialog-close" type="button" @click="reportPanelVisible = false; reportDetailView = null">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>战报</h2>
            <div class="report-tabs">
              <button :class="{ active: reportFilter === 'all' }" type="button" @click="openReportPanel('all')">全部</button>
              <button :class="{ active: reportFilter === 'attack' }" type="button" @click="openReportPanel('attack')">进攻</button>
              <button :class="{ active: reportFilter === 'defend' }" type="button" @click="openReportPanel('defend')">防御</button>
              <button :class="{ active: reportFilter === 'scout' }" type="button" @click="openReportPanel('scout')">侦察</button>
            </div>
            <div v-if="reportDetailView" class="report-detail">
              <h3>{{ reportDetailView.title }}</h3>
              <p class="report-content">{{ reportDetailView.content }}</p>
              <button type="button" @click="reportDetailView = null">返回列表</button>
            </div>
            <div v-else class="report-list">
              <div v-if="reportItems.length === 0" class="empty-list">暂无战报</div>
              <button
                v-for="item in reportItems"
                :key="item.id"
                class="report-item"
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
          <div class="dialog-panel shop-dialog">
            <button class="dialog-close" type="button" @click="shopPanelVisible = false">关闭</button>
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
              >
                <img :src="asset(`item_${item.gid}.png`)" alt="" @error="($event.target as HTMLImageElement).src = asset('board_listdown.png')" />
                <div class="shop-item-info">
                  <strong>{{ item.name }}</strong>
                  <span>{{ item.description }}</span>
                  <span class="shop-item-price">{{ item.price }} {{ item.battleShop ? item.medalTypeLabel : '元宝' }}</span>
                  <span class="shop-item-stock">库存: {{ item.totalCount === -1 ? '不限' : item.totalCount }} / 已购: {{ item.boughtToday }}</span>
                </div>
                <button type="button" @click="handleBuyItem(item)">购买</button>
              </div>
            </div>
          </div>
        </div>

        <!-- World Map Dialog -->
        <div v-if="worldMapPanelVisible" class="modal-layer flash-panel-layer worldmap-layer">
          <div class="worldmap-stage">
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
                type="button"
                aria-label="地图地块"
                @mousemove="handleWorldGridMove"
                @mouseleave="hideWorldGridTip"
                @click="handleWorldGridClick"
              ></button>
              <div v-if="selectedWorldGrid" class="worldmap-selected-grid" :style="selectedWorldGridMarkerStyle()"></div>
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
              <button class="worldmap-action-close" type="button" aria-label="关闭地图动作" @click="closeWorldGridAction"></button>
              <div class="worldmap-action-title">{{ selectedWorldGrid.title }}</div>
              <div class="worldmap-action-body">
                <div>{{ selectedWorldGrid.text }}</div>
              </div>
              <div class="worldmap-action-buttons">
                <button v-if="selectedWorldGrid.city" type="button">查看</button>
                <button v-if="selectedWorldGrid.city" type="button">出征</button>
                <button v-if="!selectedWorldGrid.city" type="button">占领</button>
                <button v-if="!selectedWorldGrid.city" type="button">侦察</button>
              </div>
            </div>
            <div v-if="worldMapAlert" class="worldmap-alert">
              <div class="worldmap-alert-message">{{ worldMapAlert }}</div>
              <button type="button" @click="closeWorldMapAlert">确定</button>
            </div>
            <div class="worldmap-view-shell">
              <div v-show="worldMapMode === 'map'" class="worldmap-map-panel">
                <button class="worldmap-mini-map-hit" type="button" aria-label="地图定位" @click="handleWorldMiniMapClick">
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
                  type="button"
                  :style="worldMapCityDotStyle(cityItem)"
                  :aria-label="`地图城池点 ${cityItem.name}`"
                  :title="`${cityItem.name}[${cityItem.x},${cityItem.y}]\n君主:${cityItem.owner}`"
                  @click="selectWorldCity(cityItem)"
                ></button>
                <button
                  v-for="cityItem in worldMapCities"
                  :key="`label-${cityItem.cid}`"
                  class="worldmap-city"
                  type="button"
                  :style="worldMapCityLabelStyle(cityItem)"
                  :aria-label="`地图城池 ${cityItem.name}`"
                  :title="`${cityItem.name}[${cityItem.x},${cityItem.y}]\n君主:${cityItem.owner}`"
                  @click="selectWorldCity(cityItem)"
                >
                  <span class="city-name">{{ cityItem.name }}</span>
                </button>
              </div>
            </div>
            <div class="worldmap-switch-panel">
              <div class="worldmap-side-strip top">
                <button
                  class="worldmap-side-btn zoom-in"
                  :class="{ active: worldMapMode === 'city' }"
                  type="button"
                  aria-label="城池视图"
                  @click="worldMapMode = 'city'"
                ></button>
                <button
                  class="worldmap-side-btn zoom-out"
                  :class="{ active: worldMapMode === 'map' }"
                  type="button"
                  aria-label="地图视图"
                  @click="worldMapMode = 'map'"
                ></button>
              </div>
            </div>
            <div class="worldmap-close-panel">
              <div class="worldmap-side-strip bottom">
                <button class="worldmap-side-btn focus-city" type="button" aria-label="关闭地图覆盖层" @click="worldMapMode = 'city'"></button>
              </div>
            </div>
            <div class="worldmap-control-panel">
              <button class="worldmap-control upleft" type="button" aria-label="左上" @click="moveWorldMap(-WORLD_MOVE_OBLIQUE, 0)"></button>
              <button class="worldmap-control up" type="button" aria-label="上移" @click="moveWorldMap(-WORLD_MOVE_VERT_STRAIGHT, -WORLD_MOVE_VERT_STRAIGHT)"></button>
              <button class="worldmap-control upright" type="button" aria-label="右上" @click="moveWorldMap(0, -WORLD_MOVE_OBLIQUE)"></button>
              <button class="worldmap-control left" type="button" aria-label="左移" @click="moveWorldMap(-WORLD_MOVE_HORI_STRAIGHT, WORLD_MOVE_HORI_STRAIGHT)"></button>
              <button class="worldmap-control mycity" type="button" title="返回" aria-label="返回本城" @click="resetWorldMapToCity"></button>
              <button class="worldmap-control right" type="button" aria-label="右移" @click="moveWorldMap(WORLD_MOVE_HORI_STRAIGHT, -WORLD_MOVE_HORI_STRAIGHT)"></button>
              <button class="worldmap-control downleft" type="button" aria-label="左下" @click="moveWorldMap(0, WORLD_MOVE_OBLIQUE)"></button>
              <button class="worldmap-control down" type="button" aria-label="下移" @click="moveWorldMap(WORLD_MOVE_VERT_STRAIGHT, WORLD_MOVE_VERT_STRAIGHT)"></button>
              <button class="worldmap-control downright" type="button" aria-label="右下" @click="moveWorldMap(WORLD_MOVE_OBLIQUE, 0)"></button>
              <button class="worldmap-control move" type="button" aria-label="跳转坐标" @click="submitWorldMapMove"></button>
              <input v-model="worldMapInputX" class="worldmap-input x" type="text" inputmode="numeric" maxlength="3" aria-label="X" @keydown.enter.prevent="submitWorldMapMove" @blur="submitWorldMapMove" />
              <input v-model="worldMapInputY" class="worldmap-input y" type="text" inputmode="numeric" maxlength="3" aria-label="Y" @keydown.enter.prevent="submitWorldMapMove" @blur="submitWorldMapMove" />
            </div>
          </div>
        </div>

        <!-- Union Dialog -->
        <div v-if="unionPanelVisible" class="modal-layer">
          <div class="dialog-panel union-dialog">
            <button class="dialog-close" type="button" @click="unionPanelVisible = false">关闭</button>
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
              <div v-if="unionSnapshot?.applyList.length" class="union-apply-list">
                <h4>可申请联盟</h4>
                <button
                  v-for="u in unionSnapshot.applyList"
                  :key="u.id"
                  type="button"
                >
                  {{ u.name }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Hero Dialog -->
        <div v-if="heroPanelVisible" class="modal-layer">
          <div class="dialog-panel hero-dialog">
            <button class="dialog-close" type="button" @click="heroPanelVisible = false">关闭</button>
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
              </div>
            </div>
            <div class="hero-recruit">
              <span>可招募名额: {{ heroRecruitCapacity }}</span>
              <button type="button" @click="handleRecruitHero">招募</button>
            </div>
          </div>
        </div>

        <!-- Barracks Dialog -->
        <div v-if="barracksPanelVisible" class="modal-layer">
          <div class="dialog-panel barracks-dialog">
            <button class="dialog-close" type="button" @click="barracksPanelVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>兵营</h2>
            <p class="barracks-capacity">兵力上限: {{ troopsData?.maxCapacity ?? 0 }}</p>
            <div class="troops-list">
              <div v-if="barracksTroopItems.length === 0" class="empty-list">暂无士兵</div>
              <div
                v-for="troop in barracksTroopItems"
                :key="troop.tid"
                class="troop-item"
              >
                <span class="troop-name">{{ troop.name }}</span>
                <span>编制: {{ troop.count }}</span>
                <span>伤兵: {{ troop.injured }}</span>
                <button type="button" @click="handleTrainTroop(troop.tid, 1)">训练</button>
              </div>
            </div>
          </div>
        </div>

        <!-- College/Research Dialog -->
        <div v-if="collegePanelVisible" class="modal-layer">
          <div class="dialog-panel college-dialog">
            <button class="dialog-close" type="button" @click="collegePanelVisible = false">关闭</button>
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
              >
                <div class="tech-info">
                  <strong>{{ tech.name }}</strong>
                  <span>{{ tech.description }}</span>
                  <span>当前: Lv.{{ tech.level }} / 最大: Lv.{{ tech.maxLevel }}</span>
                </div>
                <button type="button" :disabled="!tech.canResearch" @click="handleResearchTech(tech.tid)">研究</button>
              </div>
            </div>
          </div>
        </div>

        <!-- Ranking Dialog -->
        <div v-if="rankPanelVisible" class="modal-layer">
          <div class="dialog-panel rank-dialog">
            <button class="dialog-close" type="button" @click="rankPanelVisible = false">关闭</button>
            <img class="dialog-title-img" :src="asset('title.png')" alt="" />
            <h2>排行榜</h2>
            <div class="rank-tabs">
              <button :class="{ active: rankKind === 'power' }" type="button" @click="openRankPanel('power')">实力</button>
              <button :class="{ active: rankKind === 'level' }" type="button" @click="openRankPanel('level')">等级</button>
              <button :class="{ active: rankKind === 'city' }" type="button" @click="openRankPanel('city')">城池</button>
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
            <div class="battle-bg-image" :style="{ backgroundImage: `url(${asset(`battle_map_${battleMapId}.png`)})` }"></div>
            <div class="battle-blooditem battle-blooditem-left"></div>
            <div class="battle-blooditem battle-blooditem-right"></div>
            <div class="battle-canvas"></div>
            <div class="battle-flag-canvas"></div>
            <div class="battle-hero-panel"></div>
            <div class="battle-label">剩余时间: 30秒</div>
            <div class="battle-menu-panel">
              <button class="battle-menu-btn" type="button" aria-label="战场菜单" @click="openBattleMenu"></button>
            </div>
          </div>
        </div>

        <div v-if="battleMenuVisible" class="modal-layer battle-overlay-layer">
          <div class="battle-menu-dialog">
            <div class="battle-menu-surface">
              <div class="battle-menu-title">战场</div>
              <button class="battle-menu-action primary" type="button" @click="openBattleInfoDialog">申请</button>
              <button class="battle-menu-action" type="button" @click="openBattleCampaignDialog">邀请</button>
              <button class="battle-menu-action" type="button" @click="openBattleTaskDialog">任务</button>
              <button class="battle-menu-action" type="button" @click="openBattleUsersDialog">简介</button>
              <button class="battle-menu-action" type="button" @click="quitBattleField">外交</button>
              <button class="battle-menu-action" type="button" @click="closeBattleMenu">关闭</button>
            </div>
          </div>
        </div>

        <div v-if="battleCampaignVisible" class="modal-layer battle-overlay-layer">
          <div class="battle-simple-dialog battle-campaign-dialog">
            <button class="dialog-close" type="button" @click="closeBattleCampaignDialog">关闭</button>
            <div class="battle-simple-title">战场出征</div>
            <div class="battle-campaign-left">
              <div class="battle-campaign-left-title">出征部队</div>
              <button class="battle-campaign-all-btn" type="button" @click="takeAllBattleSoldiers">全选</button>
              <button class="battle-campaign-none-btn" type="button" @click="takeNoBattleSoldiers">全不选</button>
              <label class="battle-campaign-left-flag">
                <input type="checkbox" :checked="battleCampaignUseFlag" @change="toggleBattleCampaignUseFlag" />
                <span>旗帜</span>
              </label>
              <div class="battle-campaign-soldier-list">
                <div v-for="(soldier, index) in battleCampaignSoldiers" :key="soldier.name" class="battle-campaign-soldier-row">
                  <div class="battle-campaign-soldier-slot">
                    <span class="battle-campaign-soldier-fallback">{{ soldier.name.slice(0, 1) }}</span>
                    <img :src="battleSoldierIcon(soldier.sid)" :alt="soldier.name" @error="($event.target as HTMLImageElement).style.display = 'none'" />
                    <span class="battle-campaign-soldier-owned">{{ soldier.count }}</span>
                  </div>
                  <div class="battle-campaign-soldier-input">
                    <input :value="soldier.takecount" @input="onBattleSoldierInput(index, ($event.target as HTMLInputElement).value)" />
                    <button type="button" @click="setBattleSoldierCount(index, soldier.count)">最大</button>
                    <button type="button" @click="setBattleSoldierCount(index, -soldier.count)">最小</button>
                  </div>
                </div>
              </div>
            </div>
            <div class="battle-simple-row battle-campaign-target-row">
              <span>目标选择</span>
              <select v-model="battleCampaignTargetId">
                <option v-for="target in battleCampaignTargets" :key="target.id" :value="target.id">{{ target.name }}</option>
              </select>
            </div>
            <div class="battle-simple-row battle-campaign-hero-row">
              <span>将领选择</span>
              <select v-model="battleCampaignHeroId">
                <option v-for="hero in battleCampaignHeroes" :key="hero.id" :value="hero.id">{{ hero.heroname }}</option>
              </select>
            </div>
            <div class="battle-simple-row battle-campaign-food-row">
              <span>携带粮草</span>
              <span>{{ battleCampaignFoodCarry }}</span>
            </div>
            <div class="battle-simple-row battle-campaign-field-row">
              <span>战场名称</span>
              <span>{{ battleCampaignFieldName }}</span>
            </div>
            <div class="battle-simple-row battle-campaign-arrive-row">
              <span>到达时间</span>
              <span>{{ battleCampaignArriveTime }}</span>
            </div>
            <div class="battle-simple-row battle-campaign-path-row">
              <span>行军耗时</span>
              <span>{{ battleCampaignPathNeedTime }}</span>
            </div>
            <button class="battle-simple-start" type="button" @click="battleCampaignVisible = false">出征</button>
          </div>
        </div>

        <div v-if="battleInfoVisible" class="modal-layer battle-overlay-layer">
          <div class="battle-simple-dialog battle-info-dialog">
            <button class="dialog-close" type="button" @click="closeBattleInfoDialog">关闭</button>
            <div class="battle-simple-title">战场信息</div>
            <div class="battle-simple-tabs">
              <button class="battle-info-tab info" :class="{ active: battleInfoTab === 'info' }" type="button" @click="setBattleInfoTab('info')">信息</button>
              <button class="battle-info-tab news" :class="{ active: battleInfoTab === 'news' }" type="button" @click="setBattleInfoTab('news')">新闻</button>
              <template v-if="battleInfoTab === 'news'">
                <div class="battle-simple-page">{{ battleInfoPage }}/{{ battleNewsPageCount }}</div>
                <button class="battle-info-page-btn prev" type="button" @click="prevBattleInfoPage">上一页</button>
                <button class="battle-info-page-btn next" type="button" @click="nextBattleInfoPage">下一页</button>
              </template>
            </div>
            <div v-if="battleInfoTab === 'info'" class="battle-simple-body battle-info-body">
              <img class="battle-info-image" :src="asset('battle_header.jpg')" alt="" />
              <div class="battle-info-meta">
                <div v-for="item in battleInfoMeta" :key="item.label"><span>{{ item.label }}</span><span>{{ item.value }}</span></div>
              </div>
              <p>暂无战场详情。</p>
            </div>
            <div v-if="battleInfoTab === 'news'" class="battle-simple-body battle-info-news-body">
              <div class="battle-news-header">
                <span>时间</span>
                <span>内容</span>
              </div>
              <div class="battle-news-grid">
                <div v-for="item in battleInfoNewsPageItems" :key="item.time" class="battle-news-row">
                  <span class="battle-news-time" :style="{ color: `#${item.color.toString(16).padStart(6, '0')}` }">{{ item.time }}</span>
                  <span class="battle-news-content" :style="{ color: `#${item.color.toString(16).padStart(6, '0')}` }">{{ item.evtContent }}</span>
                </div>
              </div>
            </div>
            <div class="battle-help-line">更多帮助请参考 http://action.uuyx.com/GameHelp/Help1/index.html</div>
          </div>
        </div>

        <div v-if="battleUsersVisible" class="modal-layer battle-overlay-layer">
          <div class="battle-invite-users-dialog">
            <button class="dialog-close" type="button" @click="closeBattleUsersDialog">关闭</button>
            <div class="battle-invite-title">邀请玩家</div>
            <div class="battle-invite-count-label">人数</div>
            <div class="battle-invite-count-value">{{ battleInviteCountText }}</div>
            <template v-if="battleInviteCreator">
              <div class="battle-invite-name-label">邀请玩家</div>
              <div class="battle-invite-name-input">
                <input v-model="battleInviteName" maxlength="50" @keyup.enter="inviteBattleUserLocal" />
              </div>
              <button class="battle-invite-send" type="button" @click="inviteBattleUserLocal">邀请</button>
            </template>
            <div class="battle-invite-grid">
              <div class="battle-invite-grid-header">
                <span>玩家</span>
                <span>阵营</span>
                <span>状态</span>
                <span>将领</span>
                <span>荣誉</span>
                <span>操作</span>
              </div>
              <div v-for="item in battleInviteUsers" :key="item.id" class="battle-invite-grid-row">
                <span>{{ item.name }}</span>
                <span>{{ item.camp }}</span>
                <span>{{ item.state }}</span>
                <span>{{ item.herocount }}</span>
                <span>{{ item.honour }}</span>
                <span><button v-if="item.cancel" type="button" @click="cancelBattleInviteLocal(item.id)">取消</button></span>
              </div>
            </div>
          </div>
        </div>
      </template>

      <div v-if="loading" class="loading-mask">处理中...</div>
      <div v-if="error" class="error-line">{{ error }}</div>
    </section>
  </main>
</template>

