export type SessionUser = {
  uid: number;
  name: string;
  passport: string;
  passType: string;
  cityCount: number;
  defaultCid: number;
  defaultCity: string;
  sex?: number;
  face?: number;
  usersex?: number;
  userface?: number;
  prestige?: number;
  rank?: number;
  officepos?: string;
  officeposId?: number;
  nobility?: string;
  nobilityId?: number;
  union_id?: number;
  unionname?: string;
  union_pos?: string;
  unionpos?: number;
};

export type CommanderOption = {
  uid: number;
  name: string;
  passport: string;
  passType: string;
  cityCount: number;
  defaultCid: number;
  defaultCity: string;
};

export type LoginResult = {
  logged: boolean;
  queued: boolean;
  uid?: number;
  sid?: number;
  queueCount?: number;
  raw?: unknown[];
  user?: SessionUser;
};

export type CityCard = {
  cid: number;
  name: string;
  owner: string;
  x: number;
  y: number;
  resources: ResourceSnapshot;
};

export type CityListResult = {
  items: CityCard[];
};

export type DatabaseStatus = {
  connected: boolean;
  mode: string;
  source: string;
  message: string;
  connectedAt: string;
};

export type DashboardCounts = {
  users: number;
  cities: number;
  worldTiles: number;
  activeTroops: number;
  activeBattles: number;
};

export type DashboardOverview = {
  status: DatabaseStatus;
  counts: DashboardCounts;
  featuredCities: CityCard[];
  snapshotAt: string;
};

export type ResourceSnapshot = {
  wood: number;
  rock: number;
  iron: number;
  food: number;
  gold: number;
  people: number;
  peopleMax: number;
  peopleStable: number;
  peopleBuilding: number;
  foodMax: number;
  woodMax: number;
  rockMax: number;
  ironMax: number;
  goldMax: number;
};

export type ProductionState = {
  settings: {
    foodRate: number;
    woodRate: number;
    rockRate: number;
    ironRate: number;
  };
  foodAdd: number;
  foodArmyUse: number;
  woodAdd: number;
  rockAdd: number;
  ironAdd: number;
  heroFee: number;
  goldAdd: number;
  peopleWorking: number;
};

export type Building = {
  bid: number;
  name: string;
  level: number;
  position: number;
  state: number;
  stateStartTime: number;
  stateEndTime: number;
};

export type CityDetail = {
  summary: CityCard;
  morale: number;
  moraleStable: number;
  tax: number;
  complaint: number;
  goldRate: number;
  heroCount: number;
  production: ProductionState;
  buildings: Building[];
  soldiers: Soldier[];
  defenceList: Defence[];
};

export type Soldier = {
  sid: number;
  name: string;
  count: number;
};

export type Defence = {
  did: number;
  name: string;
  count: number;
};

export type BuildingOption = {
  bid: number;
  name: string;
  description: string;
  level: number;
  wood: number;
  rock: number;
  iron: number;
  food: number;
  gold: number;
  people: number;
  duration: number;
  canBuild: boolean;
  reason: string;
};

export type BuildingOptionsResult = {
  slot: {
    position: number;
    inner: boolean;
    wallSlot: boolean;
    occupied: boolean;
    unlocked: boolean;
    unlockLevel: number;
  };
  options: BuildingOption[];
};

export type BuildingUpgradeCondition = {
  preType: number;
  preId: number;
  type: string;
  upgradeNeed: number;
  currentOwn: number;
  canUpgrade: boolean;
};

export type BuildingLevelInfo = {
  bid: number;
  name: string;
  description: string;
  level: number;
  levelDescription: string;
  woodNeed: number;
  rockNeed: number;
  ironNeed: number;
  foodNeed: number;
  goldNeed: number;
  peopleNeed: number;
  upgradeTime: number;
  conditions: BuildingUpgradeCondition[];
  canUpgrade: boolean;
  reason: string;
};

export type BuildingInfoResult = {
  building: Building;
  current: BuildingLevelInfo;
  next: BuildingLevelInfo | null;
  resources: ResourceSnapshot;
};

export type SpeedGoods = {
  gid: number;
  name: string;
  description: string;
  image: number;
  group?: number | string;
  count: number;
  reduceTime: number;
  cost: number;
};

export type BuildingSpeedGoodsResult = {
  type: number;
  time: number;
  cost: number;
  position: number;
  building: Building;
  goodsList: SpeedGoods[];
};

export type UserTypeGoodsItem = {
  gid: number;
  name: string;
  description: string;
  group?: number | string;
  count: number;
  value?: number;
};

export type UserTypeGoodsResult = {
  type: number;
  cost?: number;
  goodsList: UserTypeGoodsItem[];
};

export type Guide = {
  gid: number;
  group: number;
  pregid: number;
  name: string;
  content: string;
  triggertype: number;
  triggerdetails: string;
  showpos: string;
  distype: number;
  disdetails: string;
};

export type GuideGroup = {
  group: number;
  items: Guide[];
};

export type ActivityItem = {
  content: string;
  link: string;
  interval: string | number;
};

export type ActivityList = {
  items: ActivityItem[];
};

export type PortalLink = {
  id: string;
  label: string;
  url: string;
  note: string;
  group: string;
};

export type PortalRuleSection = {
  id: string;
  title: string;
  items: string[];
};

export type PortalNotice = {
  title: string;
  greeting: string;
  body: string;
  signature: string;
  sourceLabel: string;
  sourceUrl: string;
  updatedAt: string;
};

export type PortalBoard = {
  key: string;
  label: string;
  keeper: string;
  brief: string;
  url: string;
};

export type PortalTopic = {
  id: string;
  boardKey: string;
  title: string;
  summary: string;
  content: string;
  author: string;
  role: string;
  updatedAt: string;
  sourceLabel: string;
  sourceUrl: string;
  tags: string[];
  sticky: boolean;
};

export type PortalSnapshot = {
  passType: string;
  homeButton: string;
  supportEmail: string;
  announcementLines: string[];
  notice: PortalNotice;
  links: PortalLink[];
  rules: PortalRuleSection[];
  boards: PortalBoard[];
  topics: PortalTopic[];
};

export type TaskGoal = {
  id: number;
  taskId: number;
  sort: number;
  type: number;
  count: number;
  reduce: boolean;
  content: string;
  completed: boolean;
  trackable: boolean;
  current: number;
  target: number;
  statusLabel: string;
};

export type TaskReward = {
  sort: number;
  type: number;
  count: number;
  name: string;
};

export type TaskCard = {
  id: number;
  groupId: number;
  preTaskId: number;
  name: string;
  content: string;
  todo: string;
  completed: boolean;
  goalCount: number;
  completedGoals: number;
  goals: TaskGoal[];
  rewards: TaskReward[];
};

export type TaskGroup = {
  id: number;
  type: number;
  typeLabel: string;
  name: string;
  description: string;
  total: number;
  completed: number;
  tasks: TaskCard[];
};

export type TaskCategory = {
  type: number;
  label: string;
  groupCount: number;
  taskCount: number;
  completed: number;
  groups: TaskGroup[];
};

export type TaskSnapshot = {
  categories: TaskCategory[];
  summary: {
    groupCount: number;
    taskCount: number;
    completedTasks: number;
    pendingTasks: number;
  };
};

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {})
    },
    ...init
  });

  const text = await response.text();
  let data: unknown = null;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = text;
    }
  }
  if (!response.ok) {
    const message = typeof data === "object" && data !== null
      ? (data as { error?: string; message?: string }).error ?? (data as { error?: string; message?: string }).message
      : typeof data === "string" && data.trim()
        ? data
        : "";
    throw new Error(message || `${response.status} ${response.statusText}`);
  }
  return data as T;
}

export function currentUser() {
  return request<{ user: SessionUser | null }>("/api/auth/me");
}

export function commanderOptions(limit = 18) {
  return request<{ items: CommanderOption[] }>(`/api/auth/commanders?limit=${Math.max(1, Math.min(40, Math.trunc(limit)))}`);
}

export function dashboardOverview() {
  return request<DashboardOverview>("/api/dashboard/overview");
}

export function legacyLogin(passport: string, password: string) {
  return request<LoginResult>("/api/legacy/login", {
    method: "POST",
    body: JSON.stringify({
      version: 0,
      loginType: 0,
      passType: "local",
      passport,
      password,
      auth: ""
    })
  });
}

export function legacyGuides(group = 1) {
  return request<GuideGroup>(`/api/legacy/guides?group=${group}`);
}

export function legacyActivities() {
  return request<ActivityList>("/api/legacy/activities");
}

export function legacyPortal() {
  return request<PortalSnapshot>("/api/legacy/portal");
}

export function createRole(payload: {
  uid: number;
  sid: number;
  userName: string;
  cityName: string;
  province: number;
  flagChar: string;
  sex: number;
  face: number;
}) {
  return request<{ uid: number; cid: number; user?: SessionUser; raw?: unknown[] }>("/api/legacy/role/create", {
    method: "POST",
    body: JSON.stringify({ ...payload, code: "" })
  });
}

export function myCities() {
  return request<CityListResult>("/api/me/cities");
}

export function cityDetail(cid: number) {
  return request<CityDetail>(`/api/cities/${cid}`);
}

export function updateCityTax(cid: number, tax: number) {
  return request<CityDetail>(`/api/cities/${cid}/tax`, {
    method: "PATCH",
    body: JSON.stringify({ tax })
  });
}

export function updateCityProduction(cid: number, settings: ProductionState["settings"]) {
  return request<CityDetail>(`/api/cities/${cid}/production`, {
    method: "PATCH",
    body: JSON.stringify(settings)
  });
}

export function buildingOptions(cid: number, position: number) {
  return request<BuildingOptionsResult>(`/api/cities/${cid}/buildings/options?position=${position}`);
}

export function buildingInfo(cid: number, position: number) {
  return request<BuildingInfoResult>(`/api/cities/${cid}/buildings/info?position=${position}`);
}

export function createBuilding(cid: number, position: number, bid: number) {
  return request<CityDetail>(`/api/cities/${cid}/buildings/create`, {
    method: "POST",
    body: JSON.stringify({ position, bid })
  });
}

export function upgradeBuilding(cid: number, position: number) {
  return request<CityDetail>(`/api/cities/${cid}/buildings/upgrade`, {
    method: "POST",
    body: JSON.stringify({ position })
  });
}

export function destroyBuilding(cid: number, position: number) {
  return request<CityDetail>(`/api/cities/${cid}/buildings/destroy`, {
    method: "POST",
    body: JSON.stringify({ position })
  });
}

export function cancelBuildingAction(cid: number, position: number) {
  return request<CityDetail>(`/api/cities/${cid}/buildings/cancel`, {
    method: "POST",
    body: JSON.stringify({ position })
  });
}

export function buildingSpeedGoods(cid: number, position: number) {
  return request<BuildingSpeedGoodsResult>(`/api/cities/${cid}/buildings/speed-goods?position=${position}`);
}

export function useBuildingSpeedGoods(cid: number, position: number, gid: number) {
  return request<CityDetail>(`/api/cities/${cid}/buildings/speed-goods/use`, {
    method: "POST",
    body: JSON.stringify({ position, gid })
  });
}

export function userTypeGoods(type: number) {
  return request<UserTypeGoodsResult>(`/api/legacy/user-type-goods?type=${type}`);
}

export function useUserTypeGoods(type: number, gid: number, cid: number) {
  return request<CityDetail>("/api/legacy/user-type-goods/use", {
    method: "POST",
    body: JSON.stringify({ type, gid, cid })
  });
}

export function myTasks() {
  return request<TaskSnapshot>("/api/me/tasks");
}

export function battleTasks(params?: { bid?: number; unionId?: number }) {
  const query = new URLSearchParams();
  if (params?.bid) query.set("bid", String(params.bid));
  if (params?.unionId) query.set("unionId", String(params.unionId));
  const suffix = query.toString();
  return request<TaskSnapshot>(suffix ? `/api/battle/tasks?${suffix}` : "/api/battle/tasks");
}

export function claimTaskReward(taskId: number) {
  return request<TaskSnapshot>("/api/me/tasks/claim", {
    method: "POST",
    body: JSON.stringify({ taskId })
  });
}

// Relation types
export type RelationCard = {
  uid: number;
  name: string;
  passport: string;
  passType: string;
  relationType: number;
  relationLabel: string;
  unionName: string;
  nobility: number;
  cityCount: number;
  defaultCid: number;
  defaultCity: string;
  updatedAt: string;
};

export type RelationPage = {
  limit: number;
  total: number;
  friendCount: number;
  enemyCount: number;
  items: RelationCard[];
};

export function myRelations() {
  return request<RelationPage>("/api/me/relations");
}

export function addRelation(name: string, relationType: number) {
  return request<RelationPage>("/api/me/relations", {
    method: "POST",
    body: JSON.stringify({ name, relationType })
  });
}

export function removeRelation(targetUid: number, relationType: number) {
  return request<RelationPage>("/api/me/relations/remove", {
    method: "POST",
    body: JSON.stringify({ targetUid, relationType })
  });
}

function parseLegacyTime(value?: string) {
  if (!value) return 0;
  const timestamp = Date.parse(value.replace(" ", "T"));
  return Number.isFinite(timestamp) ? Math.floor(timestamp / 1000) : 0;
}

function htmlToPlainText(html?: string) {
  if (!html) return "";
  if (typeof DOMParser !== "undefined") {
    const doc = new DOMParser().parseFromString(html, "text/html");
    return (doc.body.textContent ?? "").replace(/\s+/g, " ").trim();
  }
  return html.replace(/<[^>]+>/g, " ").replace(/\s+/g, " ").trim();
}

// Mail types
export type MailItem = {
  id: number;
  folder?: "inbox" | "sent" | "sys";
  type: number; // 1=inbox, 2=sent, 3=sys
  title: string;
  from: string;
  to: string;
  content: string;
  sendTime: number;
  hasAttachment: boolean;
  read: boolean;
};

export type MailPage = {
  total: number;
  page: number;
  pageSize: number;
  items: MailItem[];
};

type LegacyMailFolder = "inbox" | "outbox" | "system";

type LegacyMailSummary = {
  id: number;
  folder: LegacyMailFolder;
  fromName: string;
  toName: string;
  title: string;
  read: boolean;
  createdAt: string;
  snippet: string;
};

type LegacyMailPage = {
  folder: LegacyMailFolder;
  page: number;
  pageCount: number;
  total: number;
  items: LegacyMailSummary[];
};

type LegacyMailDetail = {
  folder: LegacyMailFolder;
  summary: LegacyMailSummary;
  htmlDocument: string;
};

function toLegacyMailFolder(folder: string): LegacyMailFolder {
  if (folder === "sent") return "outbox";
  if (folder === "sys") return "system";
  return "inbox";
}

function fromLegacyMailFolder(folder: string): "inbox" | "sent" | "sys" {
  if (folder === "outbox") return "sent";
  if (folder === "system") return "sys";
  return "inbox";
}

function normalizeMailItem(summary: LegacyMailSummary, content?: string): MailItem {
  const folder = fromLegacyMailFolder(summary.folder);
  return {
    id: summary.id,
    folder,
    type: folder === "sys" ? 3 : folder === "sent" ? 2 : 1,
    title: summary.title,
    from: summary.fromName,
    to: summary.toName,
    content: content ?? summary.snippet,
    sendTime: parseLegacyTime(summary.createdAt),
    hasAttachment: false,
    read: summary.read
  };
}

function normalizeMailPage(page: LegacyMailPage): MailPage {
  return {
    total: page.total,
    page: page.page + 1,
    pageSize: 10,
    items: (page.items ?? []).map((item) => normalizeMailItem(item))
  };
}

export async function mailPage(folder = "inbox", page = 1): Promise<MailPage> {
  const legacyFolder = toLegacyMailFolder(folder);
  const legacyPage = Math.max(0, page - 1);
  const result = await request<LegacyMailPage>(`/api/mail?folder=${legacyFolder}&page=${legacyPage}`);
  return normalizeMailPage(result);
}

export async function mailDetail(folder: string, id: number): Promise<MailItem> {
  const legacyFolder = toLegacyMailFolder(folder);
  const result = await request<LegacyMailDetail>(`/api/mail/${legacyFolder}/${id}`);
  return normalizeMailItem(result.summary, htmlToPlainText(result.htmlDocument) || result.summary.snippet);
}

export async function sendMail(toName: string, title: string, content: string): Promise<MailItem> {
  const result = await request<LegacyMailDetail>("/api/mail/send", {
    method: "POST",
    body: JSON.stringify({ toName, title, content })
  });
  return normalizeMailItem(result.summary, htmlToPlainText(result.htmlDocument) || result.summary.snippet);
}

export async function deleteMail(folder: string, id: number, page = 1): Promise<MailPage> {
  const result = await request<LegacyMailPage>("/api/mail/delete", {
    method: "POST",
    body: JSON.stringify({
      folder: toLegacyMailFolder(folder),
      ids: [id],
      page: Math.max(0, page - 1)
    })
  });
  return normalizeMailPage(result);
}

// Report types
export type ReportItem = {
  id: number;
  type: number; // 1=attack, 2=defend, 3=scout, etc
  title: string;
  content: string;
  reportTime: number;
  read: boolean;
};

export type ReportPage = {
  total: number;
  page: number;
  pageSize: number;
  items: ReportItem[];
};

type LegacyReportSummary = {
  id: number;
  type: number;
  title: number;
  read: boolean;
  createdAt: string;
  headline: string;
  snippet: string;
};

type LegacyReportPage = {
  filter: string;
  page: number;
  pageCount: number;
  total: number;
  items: LegacyReportSummary[];
};

type LegacyReportDetail = {
  summary: LegacyReportSummary;
  htmlDocument: string;
};

function toLegacyReportFilter(filter: string) {
  switch (filter) {
    case "all":
      return "type3";
    case "attack":
      return "type0";
    case "defend":
      return "type1";
    case "scout":
      return "type2";
    case "unread":
      return "unread";
    default:
      return "unread";
  }
}

function normalizeReportItem(summary: LegacyReportSummary, content?: string): ReportItem {
  return {
    id: summary.id,
    type: summary.type,
    title: summary.headline || `战报 #${summary.id}`,
    content: content ?? summary.snippet,
    reportTime: parseLegacyTime(summary.createdAt),
    read: summary.read
  };
}

function normalizeReportPage(page: LegacyReportPage): ReportPage {
  return {
    total: page.total,
    page: page.page + 1,
    pageSize: 10,
    items: (page.items ?? []).map((item) => normalizeReportItem(item))
  };
}

export async function reports(filter = "all", page = 1): Promise<ReportPage> {
  const legacyFilter = toLegacyReportFilter(filter);
  const legacyPage = Math.max(0, page - 1);
  const result = await request<LegacyReportPage>(`/api/reports?filter=${legacyFilter}&page=${legacyPage}`);
  return normalizeReportPage(result);
}

export async function reportDetail(id: number): Promise<ReportItem> {
  const result = await request<LegacyReportDetail>(`/api/reports/${id}`);
  return normalizeReportItem(result.summary, htmlToPlainText(result.htmlDocument) || result.summary.snippet);
}

// Shop types
export type ShopItem = {
  id: number;
  gid: number;
  groupId: number;
  groupLabel: string;
  name: string;
  description: string;
  pack: number;
  price: number;
  originalPrice: number;
  totalCount: number;
  userLimit: number;
  dayLimit: number;
  battleDayLimit: number;
  position: number;
  commended: boolean;
  hot: boolean;
  battleShop: boolean;
  creditPrice: number;
  medalPrice: number;
  medalTypeId: number;
  medalTypeLabel: string;
  battleGoodsType: number;
  boughtTotal: number;
  boughtToday: number;
};

export type ShopGroup = {
  id: number;
  label: string;
  itemCount: number;
  items: ShopItem[];
};

export type ShopSnapshot = {
  wallet: {
    userName: string;
    focusCid: number;
    focusCity: string;
    yuanbao: number;
    gift: number;
    gold: number;
    honour: number;
  };
  medals: {
    thingId: number;
    name: string;
    count: number;
  }[];
  groups: ShopGroup[];
};

export function myShop() {
  return request<ShopSnapshot>("/api/me/shop");
}

export function buyShopItem(itemId: number, count = 1, payType = 0, cityId = 0) {
  return request<ShopSnapshot>("/api/me/shop/buy", {
    method: "POST",
    body: JSON.stringify({ itemId, count, payType, cityId })
  });
}

// Charge types
export type ChargeSummary = {
  userName: string;
  focusCity: string;
  yuanbao: number;
  gift: number;
  todayPaid: number;
  totalPaid: number;
  payCount: number;
  exchangeRate: number;
  pendingOrders: number;
  readOnly: boolean;
};

export type ChargeBucket = {
  id: string;
  label: string;
  minMoney: number;
  maxMoney: number;
  yuanbao: number;
  playerCount: number;
};

export type ChargeEvent = {
  actId: number;
  name: string;
  moneyLimit: number;
  dayCount: number;
  mailTitle: string;
  startAt: string;
  endAt: string;
  active: boolean;
};

export type ChargeSnapshot = {
  summary: ChargeSummary;
  buckets: ChargeBucket[];
  events: ChargeEvent[];
};

export function myCharge() {
  return request<ChargeSnapshot>("/api/me/charge");
}

export function exchangeCharge(exchangeCount: number) {
  return request<ChargeSnapshot>("/api/me/charge/exchange", {
    method: "POST",
    body: JSON.stringify({ exchangeCount })
  });
}

// Hero types
export type Hero = {
  hid: number;
  uid?: number;
  cid?: number;
  name: string;
  sex?: number;
  face?: number;
  state?: number;
  stateLabel?: string;
  level: number;
  exp: number;
  loyalty?: number;
  command?: number;
  affairs?: number;
  bravery?: number;
  wisdom?: number;
  affairsAdd?: number;
  braveryAdd?: number;
  wisdomAdd?: number;
  availablePoints?: number;
  force?: number;
  forceMax?: number;
  energy?: number;
  energyMax?: number;
  智力?: number;
  武力?: number;
  统兵?: number;
  status?: number; // legacy frontend fixture shape
  cityId?: number | null;
  cityName?: string | null;
  statename?: string;
  stateName?: string;
};

export type HeroRecruit = {
  id: number;
  name: string;
  sex: number;
  face: number;
  cid: number;
  level: number;
  affairsBase: number;
  braveryBase: number;
  wisdomBase: number;
  affairsAdd: number;
  braveryAdd: number;
  wisdomAdd: number;
  loyalty: number;
  goldNeed: number;
};

export type HeroRoster = {
  cid?: number;
  cityName?: string;
  owner?: string;
  count?: number;
  items?: Hero[];
  hotelLevel?: number;
  recruitCapacity?: number;
  recruits?: HeroRecruit[];
  heroes?: Hero[];
  maxCount?: number;
  recruitFreeCount?: number;
  nextRecruitTime?: number;
};

export function cityHeroes(cid: number, limit = 100) {
  return request<HeroRoster>(`/api/cities/${cid}/heroes?limit=${limit}`);
}

export function recruitHero(cid: number, recruitId: number) {
  return request<HeroRoster>("/api/cities/" + cid + "/heroes/recruit", {
    method: "POST",
    body: JSON.stringify({ recruitId })
  });
}

export function updateCityChief(cid: number, hid: number) {
  return request<HeroRoster>(`/api/cities/${cid}/chief`, {
    method: "PATCH",
    body: JSON.stringify({ hid })
  });
}

export function updateCityGeneral(cid: number, hid: number) {
  return request<HeroRoster>(`/api/cities/${cid}/general`, {
    method: "PATCH",
    body: JSON.stringify({ hid })
  });
}

export function updateCityCounsellor(cid: number, hid: number) {
  return request<HeroRoster>(`/api/cities/${cid}/counsellor`, {
    method: "PATCH",
    body: JSON.stringify({ hid })
  });
}

export function addHeroPoints(cid: number, hid: number, stat: string, amount: number) {
  return request<HeroRoster>(`/api/cities/${cid}/heroes/${hid}/points`, {
    method: "PATCH",
    body: JSON.stringify({ stat, amount })
  });
}

export type HeroArmorAttribute = {
  type: number;
  label: string;
  value: number;
};

export type HeroArmorItem = {
  sid: number;
  armorId: number;
  name: string;
  part: number;
  partLabel: string;
  slotKey: number;
  slotLabel: string;
  type: number;
  heroLevel: number;
  durability: number;
  durabilityMax: number;
  originalDurabilityMax: number;
  repairGoldNeed: number;
  renovateMoneyNeed: number;
  recycleGold: number;
  equipped: boolean;
  attributes: HeroArmorAttribute[];
};

export type HeroArmorSlot = {
  spart: number;
  part: number;
  partLabel: string;
  slotLabel: string;
  equipped: HeroArmorItem | null;
};

export type HeroArmorSnapshot = {
  hid: number;
  cid: number;
  heroName: string;
  heroLevel: number;
  heroState: number;
  heroStateLabel: string;
  slots: HeroArmorSlot[];
  inventory: HeroArmorItem[];
};

export function heroArmorSnapshot(cid: number, hid: number) {
  return request<HeroArmorSnapshot>(`/api/cities/${cid}/heroes/${hid}/armor`);
}

export function equipHeroArmor(cid: number, hid: number, sid: number, spart: number) {
  return request<HeroArmorSnapshot>(`/api/cities/${cid}/heroes/${hid}/armor/equip`, {
    method: "POST",
    body: JSON.stringify({ sid, spart })
  });
}

export function offloadHeroArmor(cid: number, hid: number, spart: number) {
  return request<HeroArmorSnapshot>(`/api/cities/${cid}/heroes/${hid}/armor/offload`, {
    method: "POST",
    body: JSON.stringify({ spart })
  });
}

export function repairHeroArmor(cid: number, hid: number, sid: number) {
  return request<HeroArmorSnapshot>(`/api/cities/${cid}/heroes/${hid}/armor/repair`, {
    method: "POST",
    body: JSON.stringify({ sid })
  });
}

export function repairAllHeroArmor(cid: number, hid: number, sids: number[]) {
  return request<HeroArmorSnapshot>(`/api/cities/${cid}/heroes/${hid}/armor/repair-all`, {
    method: "POST",
    body: JSON.stringify({ sids })
  });
}

export function renovateHeroArmor(cid: number, hid: number, sid: number) {
  return request<HeroArmorSnapshot>(`/api/cities/${cid}/heroes/${hid}/armor/renovate`, {
    method: "POST",
    body: JSON.stringify({ sid })
  });
}

export function renovateAllHeroArmor(cid: number, hid: number, sids: number[]) {
  return request<HeroArmorSnapshot>(`/api/cities/${cid}/heroes/${hid}/armor/renovate-all`, {
    method: "POST",
    body: JSON.stringify({ sids })
  });
}

export function recycleHeroArmor(cid: number, hid: number, sid: number) {
  return request<HeroArmorSnapshot>(`/api/cities/${cid}/heroes/${hid}/armor/recycle`, {
    method: "POST",
    body: JSON.stringify({ sid })
  });
}

// Union types
export type UnionMember = {
  uid: number;
  name: string;
  level: number;
  position: number; // 1=leader, 2=officer, 3=member
  positionLabel?: string;
  cityCount?: number;
  defaultCid?: number;
  defaultCity?: string;
  contribute: number;
  lastLogin: number | string;
};

export type UnionRelation = {
  unionId: number;
  name: string;
  leaderName: string;
  relationType: number;
  relationLabel: string;
  memberCount: number;
  rank: number;
  prestige: number;
};

export type UnionApplication = {
  unionId: number;
  unionName: string;
  createdAt: string;
};

export type UnionPermissions = {
  canCreate: boolean;
  canApply: boolean;
  canCancelApply: boolean;
  canLeave: boolean;
  canEditProfile: boolean;
  canManageRelations: boolean;
};

export type UnionInfo = {
  id: number;
  name: string;
  leader: string;
  level: number;
  memberCount: number;
  maxMembers: number;
  intro?: string;
  announcement: string;
  myPosition?: number;
  myPositionLabel?: string;
  members: UnionMember[];
};

export type UnionSnapshot = {
  union: UnionInfo | null;
  applyList: { id: number; name: string; leaderName?: string; memberCount?: number; rank?: number; prestige?: number; intro?: string; isApplied?: boolean }[];
  application?: UnionApplication | null;
  permissions?: UnionPermissions;
  relations?: UnionRelation[];
};

type LegacyUnionMember = Partial<UnionMember> & {
  positionLabel?: string;
  cityCount?: number;
  lastOnline?: number | string;
};

type LegacyUnionSummary = {
  id?: number;
  name?: string;
  leader?: string;
  leaderName?: string;
  level?: number;
  memberCount?: number;
  maxMembers?: number;
  announcement?: string;
  intro?: string;
  myPosition?: number;
  myPositionLabel?: string;
};

type LegacyUnionDirectoryItem = {
  id?: number;
  name?: string;
  leaderName?: string;
  memberCount?: number;
  rank?: number;
  prestige?: number;
  intro?: string;
  isApplied?: boolean;
};

type LegacyUnionSnapshot = Partial<UnionSnapshot> & {
  joined?: boolean;
  summary?: LegacyUnionSummary | null;
  members?: LegacyUnionMember[];
  directory?: LegacyUnionDirectoryItem[];
  application?: UnionApplication | null;
  permissions?: UnionPermissions;
  relations?: UnionRelation[];
};

function normalizeUnionSnapshot(snapshot: LegacyUnionSnapshot): UnionSnapshot {
  if ("union" in snapshot || "applyList" in snapshot) {
    return {
      union: snapshot.union ?? null,
      applyList: snapshot.applyList ?? [],
      application: snapshot.application ?? null,
      permissions: snapshot.permissions,
      relations: snapshot.relations ?? []
    };
  }

  const summary = snapshot.summary ?? null;
  const defaultPermissions = {
    canCreate: !summary,
    canApply: !summary,
    canCancelApply: !!snapshot.application,
    canLeave: !!summary,
    canEditProfile: false,
    canManageRelations: false
  };
  return {
    union: summary ? {
      id: Number(summary.id ?? 0),
      name: summary.name ?? "",
      leader: summary.leaderName ?? summary.leader ?? "",
      level: Number(summary.level ?? 1),
      memberCount: Number(summary.memberCount ?? snapshot.members?.length ?? 0),
      maxMembers: Number(summary.maxMembers ?? Math.max(100, snapshot.members?.length ?? 0)),
      intro: summary.intro ?? "",
      announcement: summary.announcement ?? summary.intro ?? "",
      myPosition: Number(summary.myPosition ?? 0),
      myPositionLabel: summary.myPositionLabel ?? "",
      members: (snapshot.members ?? []).map((member) => ({
        uid: Number(member.uid ?? 0),
        name: member.name ?? "",
        level: Number(member.level ?? 0),
        position: Number(member.position ?? 1),
        positionLabel: member.positionLabel ?? "",
        cityCount: Number(member.cityCount ?? 0),
        defaultCid: Number((member as LegacyUnionMember & { defaultCid?: number }).defaultCid ?? 0),
        defaultCity: String((member as LegacyUnionMember & { defaultCity?: string }).defaultCity ?? ""),
        contribute: Number(member.contribute ?? 0),
        lastLogin: member.lastOnline ?? member.lastLogin ?? 0
      }))
    } : null,
    applyList: (snapshot.directory ?? []).map((item) => ({
      id: Number(item.id ?? 0),
      name: item.name ?? "",
      leaderName: item.leaderName ?? "",
      memberCount: Number(item.memberCount ?? 0),
      rank: Number(item.rank ?? 0),
      prestige: Number(item.prestige ?? 0),
      intro: item.intro ?? "",
      isApplied: Boolean(item.isApplied)
    })).filter((item) => item.id > 0 || item.name),
    application: snapshot.application ?? null,
    permissions: snapshot.permissions ?? defaultPermissions,
    relations: snapshot.relations ?? []
  };
}

export async function myUnion(): Promise<UnionSnapshot> {
  return normalizeUnionSnapshot(await request<LegacyUnionSnapshot>("/api/me/union"));
}

export function createUnion(name: string) {
  return request<{ id: number }>("/api/me/union/create", {
    method: "POST",
    body: JSON.stringify({ name })
  });
}

export async function applyUnion(id: number): Promise<UnionSnapshot> {
  return normalizeUnionSnapshot(await request<LegacyUnionSnapshot>("/api/me/union/apply", {
    method: "POST",
    body: JSON.stringify({ unionId: id })
  }));
}

export async function cancelUnionApply(): Promise<UnionSnapshot> {
  return normalizeUnionSnapshot(await request<LegacyUnionSnapshot>("/api/me/union/apply/cancel", {
    method: "POST"
  }));
}

export async function leaveUnion(): Promise<UnionSnapshot> {
  return normalizeUnionSnapshot(await request<LegacyUnionSnapshot>("/api/me/union/leave", {
    method: "POST"
  }));
}

export async function updateUnionProfile(name: string, intro: string, announcement: string): Promise<UnionSnapshot> {
  return normalizeUnionSnapshot(await request<LegacyUnionSnapshot>("/api/me/union/profile", {
    method: "POST",
    body: JSON.stringify({ name, intro, announcement })
  }));
}

export async function setUnionRelation(targetUnionId: number, relationType: number): Promise<UnionSnapshot> {
  return normalizeUnionSnapshot(await request<LegacyUnionSnapshot>("/api/me/union/relations", {
    method: "POST",
    body: JSON.stringify({ targetUnionId, relationType })
  }));
}

export async function removeUnionRelation(targetUnionId: number, relationType: number): Promise<UnionSnapshot> {
  return normalizeUnionSnapshot(await request<LegacyUnionSnapshot>("/api/me/union/relations/remove", {
    method: "POST",
    body: JSON.stringify({ targetUnionId, relationType })
  }));
}

export type CityBarracksDraftCondition = {
  type: string;
  name: string;
  requiredLevel: number;
  currentLevel: number;
  satisfied: boolean;
};

export type CityBarracksDraftOption = {
  sid: number;
  name: string;
  description: string;
  count: number;
  hp: number;
  ap: number;
  dp: number;
  range: number;
  speed: number;
  carry: number;
  foodUse: number;
  woodNeed: number;
  rockNeed: number;
  ironNeed: number;
  foodNeed: number;
  goldNeed: number;
  peopleNeed: number;
  draftDuration: number;
  canDraft: boolean;
  reason: string;
  conditions: CityBarracksDraftCondition[];
};

export type CityBarracksQueueItem = {
  id: number;
  sid: number;
  name: string;
  count: number;
  state: number;
  stateLabel: string;
  draftInterval: number;
  needTime: number;
  stateStartTime: number;
  endTime: number;
  secondsLeft: number;
  accMark: boolean;
};

export type CityBarracksSnapshot = {
  cid: number;
  position: number;
  level: number;
  queueCapacity: number;
  queueCount: number;
  freePeople: number;
  options: CityBarracksDraftOption[];
  queue: CityBarracksQueueItem[];
};

export function cityBarracksSnapshot(cid: number, position: number) {
  return request<CityBarracksSnapshot>(`/api/cities/${cid}/barracks?position=${position}`);
}

export function startCitySoldierDraft(cid: number, position: number, sid: number, count: number) {
  return request<CityBarracksSnapshot>(`/api/cities/${cid}/barracks/draft/start`, {
    method: "POST",
    body: JSON.stringify({ position, sid, count })
  });
}

export function cancelCitySoldierDraft(cid: number, position: number, queueId: number) {
  return request<CityBarracksSnapshot>(`/api/cities/${cid}/barracks/draft/cancel`, {
    method: "POST",
    body: JSON.stringify({ position, queueId })
  });
}

// Barracks/Troop types
export type TroopType = {
  tid: number;
  name: string;
  count: number;
  injured: number;
};

export type TroopCard = {
  id: number;
  uid: number;
  cid: number;
  startCid: number;
  targetCid: number;
  fromCity: string;
  targetCity: string;
  heroId: number;
  heroName: string;
  heroLevel: number;
  task: number;
  taskLabel: string;
  state: number;
  stateLabel: string;
  startTime: number;
  endTime: number;
  pathTime: number;
  secondsLeft: number;
  people: number;
  soldierCount: number;
  soldiers: TroopType[];
};

export type TroopPage = {
  troops: TroopType[];
  maxCapacity: number;
  total?: number;
  moving?: number;
  returning?: number;
  stationed?: number;
  battling?: number;
  gathering?: number;
  items?: TroopCard[];
};

export type BattleFieldTroopRow = {
  id: number;
  uid: number;
  cid: number;
  targetCid: number;
  startCid: number;
  battlefieldId: number;
  battleUnionId: number;
  heroId: number;
  name: string;
  union: string;
  hero: string;
  level: number;
  state: number;
  stateLabel: string;
  soldiersRaw: string;
  soldiers: TroopType[];
  soldierCount: number;
  canView: boolean;
  canPatrol: boolean;
  canAttack: boolean;
};

export type BattleFieldCityInfo = {
  cid: number;
  battlefieldId: number;
  name: string;
  uid: number;
  unionId: number;
  hasUser: boolean;
  flag: number;
  flagLabel: string;
  flagChar: string;
};

export type BattleFieldCurrentTroop = {
  id: number;
  cid: number;
  soldiersRaw: string;
  soldiers: TroopType[];
  soldierCount: number;
  heroName: string;
  heroLevel: number;
  face: number;
  sex: number;
};

export type BattleFieldInfo = {
  name: string;
  bid: number;
  minPeople: number;
  maxPeople: number;
  maxLevel: number;
  level: number;
  state: number;
  startTime: number;
  endTime: number;
  winner: number;
  content: string;
  peopleTotal: number;
  image: string;
};

export type BattleFieldWinPoint = {
  battlefieldId: number;
  unionId: number;
  point: number;
  nextReset: number;
  interval: number;
  bid: number;
  pointCount: number;
  pointName: string;
  state: number;
};

export type BattleFieldNewsItem = {
  id: number;
  battleId: number;
  unionId: number;
  content: string;
  logTime: number;
  time: string;
  color: number;
  ownUnion: boolean;
};

export type BattleFieldNewsPage = {
  page: number;
  pageSize: number;
  total: number;
  pageCount: number;
  items: BattleFieldNewsItem[];
  readOnly: boolean;
  message: string;
};

export type BattleFieldState = {
  fieldName: string;
  cid: number;
  battlefieldId: number;
  bid?: number;
  unionId: number;
  canSend: boolean;
  info?: BattleFieldInfo;
  newsTotal?: number;
  news?: BattleFieldNewsItem[];
  rows: BattleFieldTroopRow[];
  cities: BattleFieldCityInfo[];
  currentTroops: BattleFieldCurrentTroop[];
  winPoints?: BattleFieldWinPoint[];
};

export type BattleQuitPreview = {
  result: number;
  message: string;
  honourDelta: number;
  readOnly: boolean;
};

export type BattleTroopBuffer = {
  bufType: number;
  bufParam: number;
};

export type BattleTroopDetail = {
  id: number;
  uid: number;
  cid: number;
  targetCid: number;
  battleUnionId: number;
  name: string;
  union: string;
  heroId: number;
  hero: string;
  level: number;
  state: number;
  stateLabel: string;
  targetName: string;
  pathTime: number;
  endTime: number;
  secondsLeft: number;
  soldiersRaw: string;
  soldiers: TroopType[];
  soldierCount: number;
  buffers: BattleTroopBuffer[];
  readOnly: boolean;
};

export type BattleArmySendPreview = {
  troop: BattleTroopDetail;
  targetId: number;
  target: string;
  pathTime: number;
  arrival: number;
  readOnly: boolean;
  message: string;
};

export type BattleCampaignPreview = {
  cid: number;
  targetCid: number;
  target: string;
  fieldName: string;
  heroId: number;
  soldiers: TroopType[];
  soldierCount: number;
  people: number;
  foodUse: number;
  pathTime: number;
  arrival: number;
  groundLevel: number;
  capacity: number;
  currentBattleTroops: number;
  currentCityTroops: number;
  flagCount: number;
  useFlag: boolean;
  blocked: boolean;
  reason: string;
  readOnly: boolean;
  message: string;
};

export type BattleArmyAttackPreview = {
  troop: BattleTroopDetail;
  target: BattleTroopDetail;
  targetId: number;
  targetName: string;
  pathTime: number;
  arrival: number;
  readOnly: boolean;
  message: string;
};

export type BattlePatrolPreview = {
  troop: BattleTroopDetail;
  target: BattleTroopDetail;
  targetId: number;
  targetName: string;
  targetCity: string;
  report: string;
  reportLines: string[];
  pigeonCount: number;
  blocked: boolean;
  readOnly: boolean;
  message: string;
};

export type BattleMemberRow = {
  id: number;
  uid: number;
  name: string;
  camp: string;
  state: string;
  herocount: number;
  honour: number;
  cancel: boolean;
  inviteId: number;
  invited: boolean;
};

export type BattleMembersSnapshot = {
  rows: BattleMemberRow[];
  inCount: number;
  isCreator: boolean;
  readOnly: boolean;
  message: string;
};

type LegacyTroopSoldier = {
  sid?: number;
  tid?: number;
  name?: string;
  count?: number;
};

type LegacyTroopCard = {
  id?: number;
  uid?: number;
  cid?: number;
  startCid?: number;
  targetCid?: number;
  fromCity?: string;
  targetCity?: string;
  heroId?: number;
  heroName?: string;
  heroLevel?: number;
  task?: number;
  taskLabel?: string;
  state?: number;
  stateLabel?: string;
  startTime?: number;
  endTime?: number;
  pathTime?: number;
  secondsLeft?: number;
  people?: number;
  soldiers?: LegacyTroopSoldier[];
  soldierCount?: number;
};

type LegacyTroopPage = Omit<Partial<TroopPage>, "items" | "troops"> & {
  total?: number;
  moving?: number;
  returning?: number;
  stationed?: number;
  battling?: number;
  gathering?: number;
  items?: LegacyTroopCard[];
  troops?: Partial<TroopType>[];
};

function normalizeTroopPage(page: LegacyTroopPage): TroopPage {
  const normalizedItems = (page.items ?? []).map((item) => ({
    id: Number(item.id ?? 0),
    uid: Number(item.uid ?? 0),
    cid: Number(item.cid ?? 0),
    startCid: Number(item.startCid ?? item.cid ?? 0),
    targetCid: Number(item.targetCid ?? 0),
    fromCity: item.fromCity ?? "",
    targetCity: item.targetCity ?? "",
    heroId: Number(item.heroId ?? 0),
    heroName: item.heroName ?? "",
    heroLevel: Number(item.heroLevel ?? 0),
    task: Number(item.task ?? 0),
    taskLabel: item.taskLabel ?? "",
    state: Number(item.state ?? 0),
    stateLabel: item.stateLabel ?? "",
    startTime: Number(item.startTime ?? 0),
    endTime: Number(item.endTime ?? 0),
    pathTime: Number(item.pathTime ?? 0),
    secondsLeft: Number(item.secondsLeft ?? 0),
    people: Number(item.people ?? 0),
    soldierCount: Number(item.soldierCount ?? 0),
    soldiers: (item.soldiers ?? []).map((soldier) => ({
      tid: Number(soldier.tid ?? soldier.sid ?? 0),
      name: soldier.name ?? "",
      count: Number(soldier.count ?? 0),
      injured: 0
    })).filter((soldier) => soldier.tid > 0)
  }));

  if (Array.isArray(page.troops)) {
    return {
      troops: page.troops.map((item) => ({
        tid: Number(item.tid),
        name: item.name ?? "",
        count: Number(item.count ?? 0),
        injured: Number(item.injured ?? 0)
      })),
      maxCapacity: Number(page.maxCapacity ?? page.total ?? 0),
      total: Number(page.total ?? page.troops.length),
      moving: Number(page.moving ?? 0),
      returning: Number(page.returning ?? 0),
      stationed: Number(page.stationed ?? 0),
      battling: Number(page.battling ?? 0),
      gathering: Number(page.gathering ?? 0),
      items: normalizedItems
    };
  }

  const byID = new Map<number, TroopType>();
  for (const card of page.items ?? []) {
    for (const soldier of card.soldiers ?? []) {
      const tid = Number(soldier.tid ?? soldier.sid ?? 0);
      if (!Number.isFinite(tid) || tid <= 0) continue;
      const current = byID.get(tid) ?? { tid, name: soldier.name ?? `#${tid}`, count: 0, injured: 0 };
      current.count += Number(soldier.count ?? 0);
      byID.set(tid, current);
    }
  }

  return {
    troops: [...byID.values()],
    maxCapacity: Number(page.total ?? page.items?.length ?? 0),
    total: Number(page.total ?? normalizedItems.length),
    moving: Number(page.moving ?? 0),
    returning: Number(page.returning ?? 0),
    stationed: Number(page.stationed ?? 0),
    battling: Number(page.battling ?? 0),
    gathering: Number(page.gathering ?? 0),
    items: normalizedItems
  };
}

export async function myTroops(limit = 100): Promise<TroopPage> {
  return normalizeTroopPage(await request<LegacyTroopPage>(`/api/me/troops?limit=${limit}`));
}

export function battleFieldState(params: { battlefieldId: number; unionId: number; cid: number; name?: string }) {
  const query = new URLSearchParams({
    battlefieldId: String(params.battlefieldId),
    unionId: String(params.unionId),
    cid: String(params.cid)
  });
  if (params.name) query.set("name", params.name);
  return request<BattleFieldState>(`/api/battle/field-state?${query.toString()}`);
}

export function battleQuitPreview() {
  return request<BattleQuitPreview>("/api/battle/quit-preview");
}

export function battleTroopDetail(troopId: number) {
  return request<BattleTroopDetail>(`/api/battle/troop-detail?troopId=${troopId}`);
}

export function battleArmySendPreview(params: { troopId: number; targetCid: number; targetName?: string }) {
  const query = new URLSearchParams({
    troopId: String(params.troopId),
    targetCid: String(params.targetCid)
  });
  if (params.targetName) query.set("targetName", params.targetName);
  return request<BattleArmySendPreview>(`/api/battle/army-send-preview?${query.toString()}`);
}

export function battleCampaignPreview(params: {
  cid: number;
  targetCid: number;
  heroId: number;
  soldiers: { sid: number; count: number }[];
  useFlag: boolean;
}) {
  const query = new URLSearchParams({
    cid: String(params.cid),
    targetCid: String(params.targetCid),
    heroId: String(params.heroId),
    useFlag: params.useFlag ? "1" : "0"
  });
  for (const soldier of params.soldiers) {
    query.append("soldiers", `${soldier.sid}:${soldier.count}`);
  }
  return request<BattleCampaignPreview>(`/api/battle/campaign-preview?${query.toString()}`);
}

export function battleArmyAttackPreview(params: { troopId: number; targetTroopId: number; targetName?: string }) {
  const query = new URLSearchParams({
    troopId: String(params.troopId),
    targetTroopId: String(params.targetTroopId)
  });
  if (params.targetName) query.set("targetName", params.targetName);
  return request<BattleArmyAttackPreview>(`/api/battle/army-attack-preview?${query.toString()}`);
}

export function battlePatrolPreview(params: { troopId: number; targetTroopId: number }) {
  const query = new URLSearchParams({
    troopId: String(params.troopId),
    targetTroopId: String(params.targetTroopId)
  });
  return request<BattlePatrolPreview>(`/api/battle/patrol-preview?${query.toString()}`);
}

export function battleMembers() {
  return request<BattleMembersSnapshot>("/api/battle/members");
}

export function battleNews(params: { battlefieldId: number; unionId: number; page: number; pageSize?: number }) {
  const query = new URLSearchParams({
    battlefieldId: String(params.battlefieldId),
    unionId: String(params.unionId),
    page: String(params.page),
    pageSize: String(params.pageSize ?? 10)
  });
  return request<BattleFieldNewsPage>(`/api/battle/news?${query.toString()}`);
}

export async function callbackTroop(troopId: number): Promise<TroopPage> {
  return normalizeTroopPage(await request<LegacyTroopPage>(`/api/troops/${troopId}/callback`, {
    method: "POST"
  }));
}

export function dispatchCityTroop(cid: number, payload: {
  targetCid: number;
  soldierSid?: number;
  soldierCount?: number;
  soldiers?: { sid: number; count: number }[];
  heroId?: number;
  task?: number;
  resources?: { gold?: number; food?: number; wood?: number; rock?: number; iron?: number };
}) {
  return request<LegacyTroopPage>(`/api/cities/${cid}/troops/dispatch`, {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

// Research/College types
export type TechItem = {
  tid: number;
  name: string;
  level: number;
  maxLevel: number;
  cost: number;
  duration: number;
  description: string;
  canResearch: boolean;
  nextLevel?: number;
  reason?: string;
  state?: number;
  stateLabel?: string;
  secondsLeft?: number;
  woodNeed?: number;
  rockNeed?: number;
  ironNeed?: number;
  foodNeed?: number;
  goldNeed?: number;
};

export type CityResearchSnapshot = {
  cityId: number;
  cityName: string;
  position: number;
  buildingLevel: number;
  researching: TechItem | null;
  available: TechItem[];
};

type LegacyResearchOption = {
  tid: number;
  name?: string;
  description?: string;
  level?: number;
  nextLevel?: number;
  nextLevelDescription?: string;
  levelDescription?: string;
  state?: number;
  stateLabel?: string;
  researchCid?: number;
  secondsLeft?: number;
  woodNeed?: number;
  rockNeed?: number;
  ironNeed?: number;
  foodNeed?: number;
  goldNeed?: number;
  upgradeDuration?: number;
  canUpgrade?: boolean;
  reason?: string;
};

type LegacyResearchSnapshot = Partial<CityResearchSnapshot> & {
  cid?: number;
  position?: number;
  level?: number;
  activeTid?: number;
  options?: LegacyResearchOption[];
};

function normalizeResearchOption(option: LegacyResearchOption): TechItem {
  const level = Number(option.level ?? 0);
  const nextLevel = Number(option.nextLevel ?? level + 1);
  return {
    tid: Number(option.tid ?? 0),
    name: option.name ?? "",
    level,
    maxLevel: 10,
    cost: Number(option.goldNeed ?? 0),
    duration: Number(option.upgradeDuration ?? 0),
    description: option.nextLevelDescription || option.levelDescription || option.description || "",
    canResearch: Boolean(option.canUpgrade),
    nextLevel,
    reason: option.reason ?? "",
    state: Number(option.state ?? 0),
    stateLabel: option.stateLabel ?? "",
    secondsLeft: Number(option.secondsLeft ?? 0),
    woodNeed: Number(option.woodNeed ?? 0),
    rockNeed: Number(option.rockNeed ?? 0),
    ironNeed: Number(option.ironNeed ?? 0),
    foodNeed: Number(option.foodNeed ?? 0),
    goldNeed: Number(option.goldNeed ?? 0)
  };
}

function normalizeResearchSnapshot(snapshot: LegacyResearchSnapshot): CityResearchSnapshot {
  if ("available" in snapshot || "researching" in snapshot) {
    return {
      cityId: Number(snapshot.cityId ?? snapshot.cid ?? 0),
      cityName: snapshot.cityName ?? "",
      position: Number(snapshot.position ?? 0),
      buildingLevel: Number(snapshot.buildingLevel ?? snapshot.level ?? 0),
      researching: snapshot.researching ?? null,
      available: snapshot.available ?? []
    };
  }

  const available = (snapshot.options ?? []).map(normalizeResearchOption);
  const activeTid = Number(snapshot.activeTid ?? 0);
  const researching = available.find((option) => option.tid === activeTid) ?? null;
  return {
    cityId: Number(snapshot.cid ?? 0),
    cityName: "",
    position: Number(snapshot.position ?? 0),
    buildingLevel: Number(snapshot.level ?? 0),
    researching,
    available
  };
}

export async function cityResearchSnapshot(cid: number, position: number): Promise<CityResearchSnapshot> {
  return normalizeResearchSnapshot(await request<LegacyResearchSnapshot>(`/api/cities/${cid}/research?position=${position}`));
}

export async function researchTech(cid: number, position: number, tid: number): Promise<CityResearchSnapshot> {
  return normalizeResearchSnapshot(await request<LegacyResearchSnapshot>(`/api/cities/${cid}/research/start`, {
    method: "POST",
    body: JSON.stringify({ position, tid })
  }));
}

export async function cancelResearchTech(cid: number, position: number, tid: number): Promise<CityResearchSnapshot> {
  return normalizeResearchSnapshot(await request<LegacyResearchSnapshot>(`/api/cities/${cid}/research/cancel`, {
    method: "POST",
    body: JSON.stringify({ position, tid })
  }));
}

// Ranking types
export type RankItem = {
  rank: number;
  uid: number;
  name: string;
  value: number;
  detail?: string;
};

export type RankingPage = {
  kind: string;
  title?: string;
  page: number;
  pageSize: number;
  total: number;
  items: RankItem[];
};

type LegacyRankingPage = {
  kind: string;
  title?: string;
  page: number;
  pageCount?: number;
  pageSize?: number;
  total: number;
  columns?: { key: string; label: string }[];
  rows?: Record<string, string | number>[];
  items?: RankItem[];
};

function toLegacyRankingKind(kind: string) {
  switch (kind) {
    case "power":
      return "user";
    case "level":
      return "hero_level";
    case "city":
      return "city_people";
    default:
      return kind;
  }
}

export async function rankings(kind = "power", page = 1): Promise<RankingPage> {
  const legacyKind = toLegacyRankingKind(kind);
  const result = await request<LegacyRankingPage>(`/api/rankings?kind=${legacyKind}&page=${Math.max(0, page - 1)}`);
  if (Array.isArray(result.items)) {
    return {
      kind,
      title: result.title,
      page: result.page,
      pageSize: result.pageSize ?? result.items.length,
      total: result.total,
      items: result.items
    };
  }

  const rows = result.rows ?? [];
  return {
    kind,
    title: result.title,
    page: result.page + 1,
    pageSize: result.pageSize ?? Math.max(rows.length, 1),
    total: result.total,
    items: rows.map((row, index) => {
      const rank = Number(row.rank ?? index + 1);
      const prestige = Number(row.prestige ?? row.value ?? 0);
      const city = row.city ? `城池 ${row.city}` : "";
      const people = row.people ? `人口 ${row.people}` : "";
      const union = row.union_name ? `联盟 ${row.union_name}` : "";
      return {
        rank: Number.isFinite(rank) ? rank : index + 1,
        uid: Number(row.uid ?? rank ?? index + 1),
        name: String(row.name ?? ""),
        value: Number.isFinite(prestige) ? prestige : 0,
        detail: [row.nobility, city, people, union].filter(Boolean).join("  ")
      };
    })
  };
}

// World map types
export type MapCity = {
  cid: number;
  name: string;
  owner: string;
  ownerId: number;
  x: number;
  y: number;
  level: number;
  flagChar: string;
};

export type WorldTile = {
  wid: number;
  cid: number;
  x: number;
  y: number;
  type: number;
  typeName: string;
  level: number;
  state: number;
  ownerCid: number;
  cityName: string;
  ownerName: string;
  province: number;
  jun: number;
};

export type WorldMap = {
  center: { x: number; y: number };
  focusX: number;
  focusY: number;
  radius: number;
  tiles: WorldTile[];
  cities: MapCity[];
};

type LegacyWorldMap = Partial<WorldMap> & {
  center?: Partial<CityCard> & { resources?: ResourceSnapshot };
  tiles?: WorldTile[];
};

function normalizeWorldMap(payload: LegacyWorldMap): WorldMap {
  const centerX = Number(payload.focusX ?? payload.center?.x ?? 0);
  const centerY = Number(payload.focusY ?? payload.center?.y ?? 0);
  const tiles = payload.tiles ?? [];
  const cities = payload.cities ?? tiles
    .filter((tile) => Number(tile.type ?? -1) === 0 && (tile.ownerCid > 0 || tile.cityName))
    .map((tile) => ({
      cid: tile.ownerCid || tile.cid,
      name: tile.cityName || `city${tile.cid}`,
      owner: tile.ownerName || "",
      ownerId: 0,
      x: tile.x,
      y: tile.y,
      level: Math.max(1, Number(tile.level ?? 1)),
      flagChar: ""
    }));

  return {
    center: { x: centerX, y: centerY },
    focusX: centerX,
    focusY: centerY,
    radius: Number(payload.radius ?? 0),
    tiles,
    cities
  };
}

export async function worldMap(cid: number, radius = 20, x?: number, y?: number): Promise<WorldMap> {
  const params = new URLSearchParams({
    cid: String(cid),
    radius: String(radius)
  });
  if (x !== undefined && y !== undefined) {
    params.set("x", String(x));
    params.set("y", String(y));
  }
  return normalizeWorldMap(await request<LegacyWorldMap>(`/api/world/map?${params.toString()}`));
}

export function trainTroop(tid: number, count: number) {
  return request<{ success: boolean }>(`/api/troops/train`, {
    method: "POST",
    body: JSON.stringify({ tid, count })
  });
}
