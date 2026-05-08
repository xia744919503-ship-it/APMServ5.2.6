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
  const data = text ? JSON.parse(text) : null;
  if (!response.ok) {
    throw new Error(data?.error ?? data?.message ?? `${response.status} ${response.statusText}`);
  }
  return data as T;
}

export function currentUser() {
  return request<{ user: SessionUser | null }>("/api/auth/me");
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

export function claimTaskReward(taskId: number) {
  return request<TaskSnapshot>("/api/me/tasks/claim", {
    method: "POST",
    body: JSON.stringify({ taskId })
  });
}

// Mail types
export type MailItem = {
  id: number;
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

export function mailPage(folder = "inbox", page = 1) {
  return request<MailPage>(`/api/mail?folder=${folder}&page=${page}`);
}

export function mailDetail(id: number) {
  return request<MailItem>(`/api/mail/${id}`);
}

export function deleteMail(id: number) {
  return request<void>(`/api/mail/${id}`, { method: "DELETE" });
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

export function reports(filter = "all", page = 1) {
  return request<ReportPage>(`/api/reports?filter=${filter}&page=${page}`);
}

export function reportDetail(id: number) {
  return request<ReportItem>(`/api/reports/${id}`);
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

export function recruitHero(cid: number) {
  return request<{ hero?: Hero; cost: number }>("/api/cities/" + cid + "/heroes/recruit", {
    method: "POST"
  });
}

// Union types
export type UnionMember = {
  uid: number;
  name: string;
  level: number;
  position: number; // 1=leader, 2=officer, 3=member
  contribute: number;
  lastLogin: number;
};

export type UnionInfo = {
  id: number;
  name: string;
  leader: string;
  level: number;
  memberCount: number;
  maxMembers: number;
  announcement: string;
  members: UnionMember[];
};

export type UnionSnapshot = {
  union: UnionInfo | null;
  applyList: { id: number; name: string }[];
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
};

type LegacyUnionDirectoryItem = {
  id?: number;
  name?: string;
};

type LegacyUnionSnapshot = Partial<UnionSnapshot> & {
  joined?: boolean;
  summary?: LegacyUnionSummary | null;
  members?: LegacyUnionMember[];
  directory?: LegacyUnionDirectoryItem[];
};

function normalizeUnionSnapshot(snapshot: LegacyUnionSnapshot): UnionSnapshot {
  if ("union" in snapshot || "applyList" in snapshot) {
    return {
      union: snapshot.union ?? null,
      applyList: snapshot.applyList ?? []
    };
  }

  const summary = snapshot.summary ?? null;
  return {
    union: summary ? {
      id: Number(summary.id ?? 0),
      name: summary.name ?? "",
      leader: summary.leaderName ?? summary.leader ?? "",
      level: Number(summary.level ?? 1),
      memberCount: Number(summary.memberCount ?? snapshot.members?.length ?? 0),
      maxMembers: Number(summary.maxMembers ?? Math.max(100, snapshot.members?.length ?? 0)),
      announcement: summary.announcement ?? summary.intro ?? "",
      members: (snapshot.members ?? []).map((member) => ({
        uid: Number(member.uid ?? 0),
        name: member.name ?? "",
        level: Number(member.level ?? 0),
        position: Number(member.position ?? 1),
        contribute: Number(member.contribute ?? 0),
        lastLogin: Number(member.lastLogin ?? 0)
      }))
    } : null,
    applyList: (snapshot.directory ?? []).map((item) => ({
      id: Number(item.id ?? 0),
      name: item.name ?? ""
    })).filter((item) => item.id > 0 || item.name)
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

export function applyUnion(id: number) {
  return request<void>(`/api/me/union/apply/${id}`, { method: "POST" });
}

// Barracks/Troop types
export type TroopType = {
  tid: number;
  name: string;
  count: number;
  injured: number;
};

export type TroopPage = {
  troops: TroopType[];
  maxCapacity: number;
};

type LegacyTroopSoldier = {
  sid?: number;
  tid?: number;
  name?: string;
  count?: number;
};

type LegacyTroopCard = {
  soldiers?: LegacyTroopSoldier[];
  soldierCount?: number;
};

type LegacyTroopPage = Partial<TroopPage> & {
  total?: number;
  moving?: number;
  returning?: number;
  stationed?: number;
  battling?: number;
  gathering?: number;
  items?: LegacyTroopCard[];
};

function normalizeTroopPage(page: LegacyTroopPage): TroopPage {
  if (Array.isArray(page.troops)) {
    return {
      troops: page.troops.map((item) => ({
        tid: Number(item.tid),
        name: item.name,
        count: Number(item.count ?? 0),
        injured: Number(item.injured ?? 0)
      })),
      maxCapacity: Number(page.maxCapacity ?? page.total ?? 0)
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
    maxCapacity: Number(page.total ?? page.items?.length ?? 0)
  };
}

export async function myTroops(limit = 100): Promise<TroopPage> {
  return normalizeTroopPage(await request<LegacyTroopPage>(`/api/me/troops?limit=${limit}`));
}

export function trainTroop(tid: number, count: number) {
  return request<{ success: boolean }>("/api/me/troops/train", {
    method: "POST",
    body: JSON.stringify({ tid, count })
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
};

export type CityResearchSnapshot = {
  cityId: number;
  cityName: string;
  position: number;
  buildingLevel: number;
  researching: TechItem | null;
  available: TechItem[];
};

export function cityResearchSnapshot(cid: number, position: number) {
  return request<CityResearchSnapshot>(`/api/cities/${cid}/research?position=${position}`);
}

export function researchTech(cid: number, position: number, tid: number) {
  return request<{ success: boolean }>(`/api/cities/${cid}/research`, {
    method: "POST",
    body: JSON.stringify({ position, tid })
  });
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

export async function rankings(kind = "power", page = 1): Promise<RankingPage> {
  const result = await request<LegacyRankingPage>(`/api/rankings?kind=${kind}&page=${page}`);
  if (Array.isArray(result.items)) {
    return {
      kind: result.kind,
      title: result.title,
      page: result.page,
      pageSize: result.pageSize ?? result.items.length,
      total: result.total,
      items: result.items
    };
  }

  const rows = result.rows ?? [];
  return {
    kind: result.kind,
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

export type WorldMap = {
  center: { x: number; y: number };
  cities: MapCity[];
};

export function worldMap(cid: number, radius = 20) {
  return request<WorldMap>(`/api/world/map?cid=${cid}&radius=${radius}`);
}
