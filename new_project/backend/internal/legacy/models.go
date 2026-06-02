package legacy

import "time"

type DatabaseStatus struct {
	Connected   bool      `json:"connected"`
	Mode        string    `json:"mode"`
	Source      string    `json:"source"`
	Message     string    `json:"message"`
	ConnectedAt time.Time `json:"connectedAt"`
}

type Counts struct {
	Users         int `json:"users"`
	Cities        int `json:"cities"`
	WorldTiles    int `json:"worldTiles"`
	ActiveTroops  int `json:"activeTroops"`
	ActiveBattles int `json:"activeBattles"`
}

type ResourceSnapshot struct {
	Wood           int64 `json:"wood"`
	Rock           int64 `json:"rock"`
	Iron           int64 `json:"iron"`
	Food           int64 `json:"food"`
	Gold           int64 `json:"gold"`
	People         int64 `json:"people"`
	PeopleMax      int64 `json:"peopleMax"`
	PeopleStable   int64 `json:"peopleStable"`
	PeopleBuilding int64 `json:"peopleBuilding"`
	FoodMax        int64 `json:"foodMax"`
	WoodMax        int64 `json:"woodMax"`
	RockMax        int64 `json:"rockMax"`
	IronMax        int64 `json:"ironMax"`
	GoldMax        int64 `json:"goldMax"`
}

type CityCard struct {
	CID       int              `json:"cid"`
	Name      string           `json:"name"`
	Owner     string           `json:"owner"`
	X         int              `json:"x"`
	Y         int              `json:"y"`
	Resources ResourceSnapshot `json:"resources"`
}

type Building struct {
	BID            int    `json:"bid"`
	Name           string `json:"name"`
	Level          int    `json:"level"`
	Position       int    `json:"position"`
	State          int    `json:"state"`
	StateStartTime int64  `json:"stateStartTime"`
	StateEndTime   int64  `json:"stateEndTime"`
}

type CityBuildingPlacementSlot struct {
	Position    int  `json:"position"`
	Inner       bool `json:"inner"`
	WallSlot    bool `json:"wallSlot"`
	Occupied    bool `json:"occupied"`
	Unlocked    bool `json:"unlocked"`
	UnlockLevel int  `json:"unlockLevel"`
}

type CityBuildingPlacementOption struct {
	BID         int    `json:"bid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       int    `json:"level"`
	Wood        int64  `json:"wood"`
	Rock        int64  `json:"rock"`
	Iron        int64  `json:"iron"`
	Food        int64  `json:"food"`
	Gold        int64  `json:"gold"`
	People      int64  `json:"people"`
	Duration    int64  `json:"duration"`
	CanBuild    bool   `json:"canBuild"`
	Reason      string `json:"reason"`
}

type CityBuildingPlacementOptions struct {
	Slot    CityBuildingPlacementSlot     `json:"slot"`
	Options []CityBuildingPlacementOption `json:"options"`
}

type CityBuildingUpgradeCondition struct {
	PreType     int    `json:"preType"`
	PreID       int    `json:"preId"`
	Type        string `json:"type"`
	UpgradeNeed int64  `json:"upgradeNeed"`
	CurrentOwn  int64  `json:"currentOwn"`
	CanUpgrade  bool   `json:"canUpgrade"`
}

type CityBuildingLevelInfo struct {
	BID              int                            `json:"bid"`
	Name             string                         `json:"name"`
	Description      string                         `json:"description"`
	Level            int                            `json:"level"`
	LevelDescription string                         `json:"levelDescription"`
	WoodNeed         int64                          `json:"woodNeed"`
	RockNeed         int64                          `json:"rockNeed"`
	IronNeed         int64                          `json:"ironNeed"`
	FoodNeed         int64                          `json:"foodNeed"`
	GoldNeed         int64                          `json:"goldNeed"`
	PeopleNeed       int64                          `json:"peopleNeed"`
	UpgradeTime      int64                          `json:"upgradeTime"`
	Conditions       []CityBuildingUpgradeCondition `json:"conditions"`
	CanUpgrade       bool                           `json:"canUpgrade"`
	Reason           string                         `json:"reason"`
}

type CityBuildingInfo struct {
	Building  Building               `json:"building"`
	Current   CityBuildingLevelInfo  `json:"current"`
	Next      *CityBuildingLevelInfo `json:"next"`
	Resources ResourceSnapshot       `json:"resources"`
}

type SpeedGoods struct {
	GID         int    `json:"gid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       int    `json:"image"`
	Count       int64  `json:"count"`
	ReduceTime  int64  `json:"reduceTime"`
	Cost        int64  `json:"cost"`
}

type BuildingSpeedGoodsSnapshot struct {
	Type      int          `json:"type"`
	Time      int64        `json:"time"`
	Cost      int64        `json:"cost"`
	Position  int          `json:"position"`
	Building  Building     `json:"building"`
	GoodsList []SpeedGoods `json:"goodsList"`
}

type Guide struct {
	GID            int    `json:"gid"`
	Group          int    `json:"group"`
	PreGID         int    `json:"pregid"`
	Name           string `json:"name"`
	Content        string `json:"content"`
	TriggerType    int    `json:"triggertype"`
	TriggerDetails string `json:"triggerdetails"`
	ShowPos        string `json:"showpos"`
	DisType        int    `json:"distype"`
	DisDetails     string `json:"disdetails"`
}

type GuideGroup struct {
	Group int     `json:"group"`
	Items []Guide `json:"items"`
}

type ActivityItem struct {
	Content  string `json:"content"`
	Link     string `json:"link"`
	Interval string `json:"interval"`
}

type ActivityList struct {
	Items []ActivityItem `json:"items"`
}

type Soldier struct {
	SID   int    `json:"sid"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type Defence struct {
	DID   int    `json:"did"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type CityDetail struct {
	Summary      CityCard        `json:"summary"`
	Morale       int             `json:"morale"`
	MoraleStable int             `json:"moraleStable"`
	Tax          int             `json:"tax"`
	Complaint    int             `json:"complaint"`
	GoldRate     int             `json:"goldRate"`
	HeroCount    int             `json:"heroCount"`
	Production   ProductionState `json:"production"`
	Buildings    []Building      `json:"buildings"`
	Soldiers     []Soldier       `json:"soldiers"`
	DefenceList  []Defence       `json:"defenceList"`
}

type ProductionSettings struct {
	FoodRate int `json:"foodRate"`
	WoodRate int `json:"woodRate"`
	RockRate int `json:"rockRate"`
	IronRate int `json:"ironRate"`
}

type ProductionState struct {
	Settings      ProductionSettings `json:"settings"`
	FoodAdd       int64              `json:"foodAdd"`
	FoodArmyUse   int64              `json:"foodArmyUse"`
	WoodAdd       int64              `json:"woodAdd"`
	RockAdd       int64              `json:"rockAdd"`
	IronAdd       int64              `json:"ironAdd"`
	HeroFee       int64              `json:"heroFee"`
	GoldAdd       int64              `json:"goldAdd"`
	PeopleWorking int64              `json:"peopleWorking"`
}

type CommanderOption struct {
	UID         int    `json:"uid"`
	Name        string `json:"name"`
	Passport    string `json:"passport"`
	PassType    string `json:"passType"`
	CityCount   int    `json:"cityCount"`
	DefaultCID  int    `json:"defaultCid"`
	DefaultCity string `json:"defaultCity"`
}

type SessionUser struct {
	UID         int    `json:"uid"`
	Name        string `json:"name"`
	Passport    string `json:"passport"`
	PassType    string `json:"passType"`
	CityCount   int    `json:"cityCount"`
	DefaultCID  int    `json:"defaultCid"`
	DefaultCity string `json:"defaultCity"`
	Sex         int    `json:"sex"`
	Face        int    `json:"face"`
	UserSex     int    `json:"usersex"`
	UserFace    int    `json:"userface"`
	Prestige    int64  `json:"prestige"`
	Rank        int    `json:"rank"`
	OfficePos   string `json:"officepos"`
	OfficePosID int    `json:"officeposId"`
	Nobility    string `json:"nobility"`
	NobilityID  int    `json:"nobilityId"`
	UnionID     int    `json:"union_id"`
	UnionName   string `json:"unionname"`
	UnionPos    string `json:"union_pos"`
	UnionPosID  int    `json:"unionpos"`
}

type LegacyLoginPayload struct {
	Version   int    `json:"version"`
	LoginType int    `json:"loginType"`
	PassType  string `json:"passType"`
	Passport  string `json:"passport"`
	Password  string `json:"password"`
	Auth      string `json:"auth"`
}

type LegacyQueueCheckPayload struct {
	UID int   `json:"uid"`
	SID int64 `json:"sid"`
}

type LegacyLoginResult struct {
	Raw        []any        `json:"raw"`
	Logged     bool         `json:"logged"`
	Queued     bool         `json:"queued"`
	UID        int          `json:"uid,omitempty"`
	SID        int64        `json:"sid,omitempty"`
	QueueCount int          `json:"queueCount,omitempty"`
	User       *SessionUser `json:"user,omitempty"`
}

type LegacyRoleCreatePayload struct {
	UID      int    `json:"uid"`
	SID      int64  `json:"sid"`
	UserName string `json:"userName"`
	CityName string `json:"cityName"`
	Province int    `json:"province"`
	FlagChar string `json:"flagChar"`
	Sex      int    `json:"sex"`
	Face     int    `json:"face"`
	Code     string `json:"code"`
}

type LegacyRoleCreateResult struct {
	Raw  []any        `json:"raw"`
	UID  int          `json:"uid"`
	CID  int          `json:"cid"`
	User *SessionUser `json:"user,omitempty"`
}

type RelationCard struct {
	UID           int    `json:"uid"`
	Name          string `json:"name"`
	Passport      string `json:"passport"`
	PassType      string `json:"passType"`
	RelationType  int    `json:"relationType"`
	RelationLabel string `json:"relationLabel"`
	UnionName     string `json:"unionName"`
	Nobility      int    `json:"nobility"`
	CityCount     int    `json:"cityCount"`
	DefaultCID    int    `json:"defaultCid"`
	DefaultCity   string `json:"defaultCity"`
	UpdatedAt     string `json:"updatedAt"`
}

type RelationPage struct {
	Limit       int            `json:"limit"`
	Total       int            `json:"total"`
	FriendCount int            `json:"friendCount"`
	EnemyCount  int            `json:"enemyCount"`
	Items       []RelationCard `json:"items"`
}

type UnionSummary struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	LeaderUID       int    `json:"leaderUid"`
	LeaderName      string `json:"leaderName"`
	CreatorName     string `json:"creatorName"`
	MemberCount     int    `json:"memberCount"`
	CityCount       int    `json:"cityCount"`
	Rank            int    `json:"rank"`
	Prestige        int64  `json:"prestige"`
	Intro           string `json:"intro"`
	Announcement    string `json:"announcement"`
	MyPosition      int    `json:"myPosition"`
	MyPositionLabel string `json:"myPositionLabel"`
}

type UnionMember struct {
	UID           int    `json:"uid"`
	Name          string `json:"name"`
	Passport      string `json:"passport"`
	PassType      string `json:"passType"`
	Position      int    `json:"position"`
	PositionLabel string `json:"positionLabel"`
	Rank          int    `json:"rank"`
	Nobility      int    `json:"nobility"`
	CityCount     int    `json:"cityCount"`
	DefaultCID    int    `json:"defaultCid"`
	DefaultCity   string `json:"defaultCity"`
	LastOnline    string `json:"lastOnline"`
}

type UnionRelation struct {
	UnionID       int    `json:"unionId"`
	Name          string `json:"name"`
	LeaderName    string `json:"leaderName"`
	RelationType  int    `json:"relationType"`
	RelationLabel string `json:"relationLabel"`
	MemberCount   int    `json:"memberCount"`
	Rank          int    `json:"rank"`
	Prestige      int64  `json:"prestige"`
}

type UnionEvent struct {
	Type      int    `json:"type"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

type UnionPermissions struct {
	CanCreate          bool `json:"canCreate"`
	CanApply           bool `json:"canApply"`
	CanCancelApply     bool `json:"canCancelApply"`
	CanLeave           bool `json:"canLeave"`
	CanEditProfile     bool `json:"canEditProfile"`
	CanManageRelations bool `json:"canManageRelations"`
}

type UnionDirectoryCard struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	LeaderUID   int    `json:"leaderUid"`
	LeaderName  string `json:"leaderName"`
	MemberCount int    `json:"memberCount"`
	Rank        int    `json:"rank"`
	Prestige    int64  `json:"prestige"`
	Intro       string `json:"intro"`
	IsApplied   bool   `json:"isApplied"`
}

type UnionApplication struct {
	UnionID   int    `json:"unionId"`
	UnionName string `json:"unionName"`
	CreatedAt string `json:"createdAt"`
}

type UnionSnapshot struct {
	Joined      bool                 `json:"joined"`
	Summary     *UnionSummary        `json:"summary"`
	Members     []UnionMember        `json:"members"`
	Relations   []UnionRelation      `json:"relations"`
	Events      []UnionEvent         `json:"events"`
	Directory   []UnionDirectoryCard `json:"directory"`
	Application *UnionApplication    `json:"application"`
	Permissions UnionPermissions     `json:"permissions"`
}

type TaskSummary struct {
	CategoryCount   int `json:"categoryCount"`
	GroupCount      int `json:"groupCount"`
	TaskCount       int `json:"taskCount"`
	CompletedTasks  int `json:"completedTasks"`
	PendingTasks    int `json:"pendingTasks"`
	CompletedGroups int `json:"completedGroups"`
}

type TaskGoal struct {
	ID          int    `json:"id"`
	TaskID      int    `json:"taskId"`
	Sort        int    `json:"sort"`
	Type        int    `json:"type"`
	Count       int64  `json:"count"`
	Reduce      bool   `json:"reduce"`
	Content     string `json:"content"`
	Completed   bool   `json:"completed"`
	Trackable   bool   `json:"trackable"`
	Current     int64  `json:"current"`
	Target      int64  `json:"target"`
	StatusLabel string `json:"statusLabel"`
}

type TaskReward struct {
	Sort  int    `json:"sort"`
	Type  int    `json:"type"`
	Count int64  `json:"count"`
	Name  string `json:"name"`
}

type TaskCard struct {
	ID             int          `json:"id"`
	GroupID        int          `json:"groupId"`
	PreTaskID      int          `json:"preTaskId"`
	Name           string       `json:"name"`
	Content        string       `json:"content"`
	Todo           string       `json:"todo"`
	Completed      bool         `json:"completed"`
	GoalCount      int          `json:"goalCount"`
	CompletedGoals int          `json:"completedGoals"`
	Goals          []TaskGoal   `json:"goals"`
	Rewards        []TaskReward `json:"rewards"`
}

type TaskGroup struct {
	ID          int        `json:"id"`
	Type        int        `json:"type"`
	TypeLabel   string     `json:"typeLabel"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Total       int        `json:"total"`
	Completed   int        `json:"completed"`
	Tasks       []TaskCard `json:"tasks"`
}

type TaskCategory struct {
	Type       int         `json:"type"`
	Label      string      `json:"label"`
	GroupCount int         `json:"groupCount"`
	TaskCount  int         `json:"taskCount"`
	Completed  int         `json:"completed"`
	Groups     []TaskGroup `json:"groups"`
}

type TaskSnapshot struct {
	Summary    TaskSummary    `json:"summary"`
	Categories []TaskCategory `json:"categories"`
}

type ShopWallet struct {
	UserName  string `json:"userName"`
	FocusCID  int    `json:"focusCid"`
	FocusCity string `json:"focusCity"`
	Yuanbao   int64  `json:"yuanbao"`
	Gift      int64  `json:"gift"`
	Gold      int64  `json:"gold"`
	Honour    int64  `json:"honour"`
}

type ShopMedal struct {
	ThingID int    `json:"thingId"`
	Name    string `json:"name"`
	Count   int64  `json:"count"`
}

type ShopItem struct {
	ID              int    `json:"id"`
	GID             int    `json:"gid"`
	GroupID         int    `json:"groupId"`
	GroupLabel      string `json:"groupLabel"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Pack            int    `json:"pack"`
	Price           int64  `json:"price"`
	OriginalPrice   int64  `json:"originalPrice"`
	TotalCount      int64  `json:"totalCount"`
	UserLimit       int    `json:"userLimit"`
	DayLimit        int    `json:"dayLimit"`
	BattleDayLimit  int    `json:"battleDayLimit"`
	Position        int    `json:"position"`
	Commended       bool   `json:"commended"`
	Hot             bool   `json:"hot"`
	BattleShop      bool   `json:"battleShop"`
	CreditPrice     int64  `json:"creditPrice"`
	MedalPrice      int64  `json:"medalPrice"`
	MedalTypeID     int    `json:"medalTypeId"`
	MedalTypeLabel  string `json:"medalTypeLabel"`
	BattleGoodsType int    `json:"battleGoodsType"`
	BoughtTotal     int64  `json:"boughtTotal"`
	BoughtToday     int64  `json:"boughtToday"`
}

type ShopGroup struct {
	ID        int        `json:"id"`
	Label     string     `json:"label"`
	ItemCount int        `json:"itemCount"`
	Items     []ShopItem `json:"items"`
}

type ShopSnapshot struct {
	Wallet ShopWallet  `json:"wallet"`
	Medals []ShopMedal `json:"medals"`
	Groups []ShopGroup `json:"groups"`
}

type ChargeSummary struct {
	UserName      string `json:"userName"`
	FocusCity     string `json:"focusCity"`
	Yuanbao       int64  `json:"yuanbao"`
	Gift          int64  `json:"gift"`
	TodayPaid     int64  `json:"todayPaid"`
	TotalPaid     int64  `json:"totalPaid"`
	PayCount      int    `json:"payCount"`
	ExchangeRate  int    `json:"exchangeRate"`
	PendingOrders int    `json:"pendingOrders"`
	ReadOnly      bool   `json:"readOnly"`
}

type ChargeBucket struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	MinMoney    int64  `json:"minMoney"`
	MaxMoney    int64  `json:"maxMoney"`
	Yuanbao     int64  `json:"yuanbao"`
	PlayerCount int    `json:"playerCount"`
}

type ChargeEvent struct {
	ActID      int    `json:"actId"`
	Name       string `json:"name"`
	MoneyLimit int64  `json:"moneyLimit"`
	DayCount   int64  `json:"dayCount"`
	MailTitle  string `json:"mailTitle"`
	StartAt    string `json:"startAt"`
	EndAt      string `json:"endAt"`
	Active     bool   `json:"active"`
}

type ChargeSnapshot struct {
	Summary ChargeSummary  `json:"summary"`
	Buckets []ChargeBucket `json:"buckets"`
	Events  []ChargeEvent  `json:"events"`
}

type PortalLink struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	URL   string `json:"url"`
	Note  string `json:"note"`
	Group string `json:"group"`
}

type PortalRuleSection struct {
	ID    string   `json:"id"`
	Title string   `json:"title"`
	Items []string `json:"items"`
}

type PortalNotice struct {
	Title       string `json:"title"`
	Greeting    string `json:"greeting"`
	Body        string `json:"body"`
	Signature   string `json:"signature"`
	SourceLabel string `json:"sourceLabel"`
	SourceURL   string `json:"sourceUrl"`
	UpdatedAt   string `json:"updatedAt"`
}

type PortalBoard struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Keeper string `json:"keeper"`
	Brief  string `json:"brief"`
	URL    string `json:"url"`
}

type PortalTopic struct {
	ID          string   `json:"id"`
	BoardKey    string   `json:"boardKey"`
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	Content     string   `json:"content"`
	Author      string   `json:"author"`
	Role        string   `json:"role"`
	UpdatedAt   string   `json:"updatedAt"`
	SourceLabel string   `json:"sourceLabel"`
	SourceURL   string   `json:"sourceUrl"`
	Tags        []string `json:"tags"`
	Sticky      bool     `json:"sticky"`
}

type PortalSnapshot struct {
	PassType          string              `json:"passType"`
	HomeButton        string              `json:"homeButton"`
	SupportEmail      string              `json:"supportEmail"`
	AnnouncementLines []string            `json:"announcementLines"`
	Notice            PortalNotice        `json:"notice"`
	Links             []PortalLink        `json:"links"`
	Rules             []PortalRuleSection `json:"rules"`
	Boards            []PortalBoard       `json:"boards"`
	Topics            []PortalTopic       `json:"topics"`
}

type WorldTile struct {
	WID       int    `json:"wid"`
	CID       int    `json:"cid"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Type      int    `json:"type"`
	TypeName  string `json:"typeName"`
	Level     int    `json:"level"`
	State     int    `json:"state"`
	OwnerCID  int    `json:"ownerCid"`
	CityName  string `json:"cityName"`
	OwnerName string `json:"ownerName"`
	Province  int    `json:"province"`
	Jun       int    `json:"jun"`
}

type WorldMap struct {
	Center CityCard    `json:"center"`
	FocusX int         `json:"focusX"`
	FocusY int         `json:"focusY"`
	Radius int         `json:"radius"`
	Tiles  []WorldTile `json:"tiles"`
}

type MailCounts struct {
	Inbox        int `json:"inbox"`
	Outbox       int `json:"outbox"`
	System       int `json:"system"`
	UnreadInbox  int `json:"unreadInbox"`
	UnreadSystem int `json:"unreadSystem"`
}

type MailSummary struct {
	ID        int    `json:"id"`
	Folder    string `json:"folder"`
	FromName  string `json:"fromName"`
	ToName    string `json:"toName"`
	Title     string `json:"title"`
	Read      bool   `json:"read"`
	CreatedAt string `json:"createdAt"`
	Snippet   string `json:"snippet"`
}

type MailPage struct {
	Folder    string        `json:"folder"`
	Page      int           `json:"page"`
	PageCount int           `json:"pageCount"`
	Total     int           `json:"total"`
	Counts    MailCounts    `json:"counts"`
	Items     []MailSummary `json:"items"`
}

type MailDetail struct {
	Folder       string      `json:"folder"`
	Summary      MailSummary `json:"summary"`
	HTMLDocument string      `json:"htmlDocument"`
}

type ReportSummary struct {
	ID         int    `json:"id"`
	OriginCID  int    `json:"originCid"`
	OriginCity string `json:"originCity"`
	HappenCID  int    `json:"happenCid"`
	HappenCity string `json:"happenCity"`
	Title      int    `json:"title"`
	Type       int    `json:"type"`
	Read       bool   `json:"read"`
	BattleID   int    `json:"battleId"`
	CreatedAt  string `json:"createdAt"`
	Headline   string `json:"headline"`
	Snippet    string `json:"snippet"`
}

type ReportPage struct {
	Filter    string          `json:"filter"`
	Page      int             `json:"page"`
	PageCount int             `json:"pageCount"`
	Total     int             `json:"total"`
	Items     []ReportSummary `json:"items"`
}

type ReportDetail struct {
	Summary      ReportSummary `json:"summary"`
	HTMLDocument string        `json:"htmlDocument"`
}

type RankingColumn struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type RankingPage struct {
	Kind      string              `json:"kind"`
	Title     string              `json:"title"`
	UpdatedAt string              `json:"updatedAt"`
	Page      int                 `json:"page"`
	PageCount int                 `json:"pageCount"`
	Total     int                 `json:"total"`
	Columns   []RankingColumn     `json:"columns"`
	Rows      []map[string]string `json:"rows"`
}

type HeroCard struct {
	HID        int    `json:"hid"`
	UID        int    `json:"uid"`
	CID        int    `json:"cid"`
	Name       string `json:"name"`
	Sex        int    `json:"sex"`
	Face       int    `json:"face"`
	State      int    `json:"state"`
	StateName  string `json:"statename"`
	StateLabel string `json:"stateLabel"`
	Level      int    `json:"level"`
	Loyalty    int    `json:"loyalty"`
	Exp        int64  `json:"exp"`
	Command    int    `json:"command"`
	Affairs    int    `json:"affairs"`
	Bravery    int    `json:"bravery"`
	Wisdom     int    `json:"wisdom"`
	AffairsAdd int    `json:"affairsAdd"`
	BraveryAdd int    `json:"braveryAdd"`
	WisdomAdd  int    `json:"wisdomAdd"`
	Available  int    `json:"availablePoints"`
	Force      int    `json:"force"`
	ForceMax   int    `json:"forceMax"`
	Energy     int    `json:"energy"`
	EnergyMax  int    `json:"energyMax"`
}

type HeroRecruit struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Sex         int    `json:"sex"`
	Face        int    `json:"face"`
	CID         int    `json:"cid"`
	Level       int    `json:"level"`
	AffairsBase int    `json:"affairsBase"`
	BraveryBase int    `json:"braveryBase"`
	WisdomBase  int    `json:"wisdomBase"`
	AffairsAdd  int    `json:"affairsAdd"`
	BraveryAdd  int    `json:"braveryAdd"`
	WisdomAdd   int    `json:"wisdomAdd"`
	Loyalty     int    `json:"loyalty"`
	GoldNeed    int64  `json:"goldNeed"`
}

type HeroRoster struct {
	CID             int           `json:"cid"`
	CityName        string        `json:"cityName"`
	Owner           string        `json:"owner"`
	Count           int           `json:"count"`
	Items           []HeroCard    `json:"items"`
	HotelLevel      int           `json:"hotelLevel"`
	RecruitCapacity int           `json:"recruitCapacity"`
	Recruits        []HeroRecruit `json:"recruits"`
}

type HeroArmorAttribute struct {
	Type  int    `json:"type"`
	Label string `json:"label"`
	Value int    `json:"value"`
}

type HeroArmorItem struct {
	SID                   int                  `json:"sid"`
	ArmorID               int                  `json:"armorId"`
	Name                  string               `json:"name"`
	Part                  int                  `json:"part"`
	PartLabel             string               `json:"partLabel"`
	SlotKey               int                  `json:"slotKey"`
	SlotLabel             string               `json:"slotLabel"`
	Type                  int                  `json:"type"`
	HeroLevel             int                  `json:"heroLevel"`
	Durability            int                  `json:"durability"`
	DurabilityMax         int                  `json:"durabilityMax"`
	OriginalDurabilityMax int                  `json:"originalDurabilityMax"`
	RepairGoldNeed        int64                `json:"repairGoldNeed"`
	RenovateMoneyNeed     int64                `json:"renovateMoneyNeed"`
	RecycleGold           int64                `json:"recycleGold"`
	Equipped              bool                 `json:"equipped"`
	Attributes            []HeroArmorAttribute `json:"attributes"`
}

type HeroArmorSlot struct {
	Spart     int            `json:"spart"`
	Part      int            `json:"part"`
	PartLabel string         `json:"partLabel"`
	SlotLabel string         `json:"slotLabel"`
	Equipped  *HeroArmorItem `json:"equipped"`
}

type HeroArmorSnapshot struct {
	HID            int             `json:"hid"`
	CID            int             `json:"cid"`
	HeroName       string          `json:"heroName"`
	HeroLevel      int             `json:"heroLevel"`
	HeroState      int             `json:"heroState"`
	HeroStateLabel string          `json:"heroStateLabel"`
	Slots          []HeroArmorSlot `json:"slots"`
	Inventory      []HeroArmorItem `json:"inventory"`
}

type TroopResource struct {
	Gold int64 `json:"gold"`
	Food int64 `json:"food"`
	Wood int64 `json:"wood"`
	Rock int64 `json:"rock"`
	Iron int64 `json:"iron"`
}

type TroopCard struct {
	ID           int           `json:"id"`
	UID          int           `json:"uid"`
	CID          int           `json:"cid"`
	StartCID     int           `json:"startCid"`
	TargetCID    int           `json:"targetCid"`
	FromCity     string        `json:"fromCity"`
	TargetCity   string        `json:"targetCity"`
	HeroID       int           `json:"heroId"`
	HeroName     string        `json:"heroName"`
	HeroLevel    int           `json:"heroLevel"`
	Task         int           `json:"task"`
	TaskLabel    string        `json:"taskLabel"`
	State        int           `json:"state"`
	StateLabel   string        `json:"stateLabel"`
	StartTime    int64         `json:"startTime"`
	EndTime      int64         `json:"endTime"`
	PathTime     int           `json:"pathTime"`
	SecondsLeft  int64         `json:"secondsLeft"`
	People       int64         `json:"people"`
	FoodUse      float64       `json:"foodUse"`
	SoldiersRaw  string        `json:"soldiersRaw"`
	ResourceRaw  string        `json:"resourceRaw"`
	Resources    TroopResource `json:"resources"`
	Resource     TroopResource `json:"resource"`
	Soldiers     []Soldier     `json:"soldiers"`
	SoldierCount int64         `json:"soldierCount"`
}

type TroopPage struct {
	Total     int         `json:"total"`
	Moving    int         `json:"moving"`
	Returning int         `json:"returning"`
	Stationed int         `json:"stationed"`
	Battling  int         `json:"battling"`
	Gathering int         `json:"gathering"`
	Items     []TroopCard `json:"items"`
}

type BattleFieldTroopRow struct {
	ID            int       `json:"id"`
	UID           int       `json:"uid"`
	CID           int       `json:"cid"`
	TargetCID     int       `json:"targetCid"`
	StartCID      int       `json:"startCid"`
	BattlefieldID int       `json:"battlefieldId"`
	BattleUnionID int       `json:"battleUnionId"`
	HeroID        int       `json:"heroId"`
	Name          string    `json:"name"`
	Union         string    `json:"union"`
	Hero          string    `json:"hero"`
	Level         int       `json:"level"`
	State         int       `json:"state"`
	StateLabel    string    `json:"stateLabel"`
	SoldiersRaw   string    `json:"soldiersRaw"`
	Soldiers      []Soldier `json:"soldiers"`
	SoldierCount  int64     `json:"soldierCount"`
	CanView       bool      `json:"canView"`
	CanPatrol     bool      `json:"canPatrol"`
	CanAttack     bool      `json:"canAttack"`
}

type BattleFieldCityInfo struct {
	CID           int    `json:"cid"`
	BattlefieldID int    `json:"battlefieldId"`
	Name          string `json:"name"`
	UID           int    `json:"uid"`
	UnionID       int    `json:"unionId"`
	HasUser       bool   `json:"hasUser"`
	Flag          int    `json:"flag"`
	FlagLabel     string `json:"flagLabel"`
	FlagChar      string `json:"flagChar"`
}

type BattleFieldCurrentTroop struct {
	ID           int       `json:"id"`
	CID          int       `json:"cid"`
	SoldiersRaw  string    `json:"soldiersRaw"`
	Soldiers     []Soldier `json:"soldiers"`
	SoldierCount int64     `json:"soldierCount"`
	HeroName     string    `json:"heroName"`
	HeroLevel    int       `json:"heroLevel"`
	Face         int       `json:"face"`
	Sex          int       `json:"sex"`
}

type BattleFieldInfo struct {
	Name        string `json:"name"`
	BID         int    `json:"bid"`
	MinPeople   int    `json:"minPeople"`
	MaxPeople   int    `json:"maxPeople"`
	MaxLevel    int    `json:"maxLevel"`
	Level       int    `json:"level"`
	State       int    `json:"state"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
	Winner      int    `json:"winner"`
	Content     string `json:"content"`
	PeopleTotal int    `json:"peopleTotal"`
	Image       string `json:"image"`
}

type BattleFieldWinPoint struct {
	BattlefieldID int    `json:"battlefieldId"`
	UnionID       int    `json:"unionId"`
	Point         int    `json:"point"`
	NextReset     int64  `json:"nextReset"`
	Interval      int    `json:"interval"`
	BID           int    `json:"bid"`
	PointCount    int    `json:"pointCount"`
	PointName     string `json:"pointName"`
	State         int    `json:"state"`
}

type BattleFieldNewsItem struct {
	ID       int    `json:"id"`
	BattleID int    `json:"battleId"`
	UnionID  int    `json:"unionId"`
	Content  string `json:"content"`
	LogTime  int64  `json:"logTime"`
	Time     string `json:"time"`
	Color    int    `json:"color"`
	OwnUnion bool   `json:"ownUnion"`
}

type BattleQuitPreview struct {
	Result      int    `json:"result"`
	Message     string `json:"message"`
	HonourDelta int    `json:"honourDelta"`
	ReadOnly    bool   `json:"readOnly"`
}

type BattleTroopDetail struct {
	ID            int                 `json:"id"`
	UID           int                 `json:"uid"`
	CID           int                 `json:"cid"`
	TargetCID     int                 `json:"targetCid"`
	BattleUnionID int                 `json:"battleUnionId"`
	Name          string              `json:"name"`
	Union         string              `json:"union"`
	HeroID        int                 `json:"heroId"`
	Hero          string              `json:"hero"`
	Level         int                 `json:"level"`
	State         int                 `json:"state"`
	StateLabel    string              `json:"stateLabel"`
	TargetName    string              `json:"targetName"`
	PathTime      int                 `json:"pathTime"`
	EndTime       int64               `json:"endTime"`
	SecondsLeft   int64               `json:"secondsLeft"`
	SoldiersRaw   string              `json:"soldiersRaw"`
	Soldiers      []Soldier           `json:"soldiers"`
	SoldierCount  int64               `json:"soldierCount"`
	Buffers       []BattleTroopBuffer `json:"buffers"`
	ReadOnly      bool                `json:"readOnly"`
}

type BattleTroopBuffer struct {
	BufType  int `json:"bufType"`
	BufParam int `json:"bufParam"`
}

type BattleArmySendPreview struct {
	Troop    BattleTroopDetail `json:"troop"`
	TargetID int               `json:"targetId"`
	Target   string            `json:"target"`
	PathTime int               `json:"pathTime"`
	Arrival  int64             `json:"arrival"`
	ReadOnly bool              `json:"readOnly"`
	Message  string            `json:"message"`
}

type BattleCampaignPreview struct {
	CID                 int       `json:"cid"`
	TargetCID           int       `json:"targetCid"`
	Target              string    `json:"target"`
	FieldName           string    `json:"fieldName"`
	HeroID              int       `json:"heroId"`
	Soldiers            []Soldier `json:"soldiers"`
	SoldierCount        int64     `json:"soldierCount"`
	People              int64     `json:"people"`
	FoodUse             int64     `json:"foodUse"`
	PathTime            int       `json:"pathTime"`
	Arrival             int64     `json:"arrival"`
	GroundLevel         int       `json:"groundLevel"`
	Capacity            int64     `json:"capacity"`
	CurrentBattleTroops int       `json:"currentBattleTroops"`
	CurrentCityTroops   int       `json:"currentCityTroops"`
	FlagCount           int64     `json:"flagCount"`
	UseFlag             bool      `json:"useFlag"`
	Blocked             bool      `json:"blocked"`
	Reason              string    `json:"reason"`
	ReadOnly            bool      `json:"readOnly"`
	Message             string    `json:"message"`
}

type BattleArmyAttackPreview struct {
	Troop      BattleTroopDetail `json:"troop"`
	Target     BattleTroopDetail `json:"target"`
	TargetID   int               `json:"targetId"`
	TargetName string            `json:"targetName"`
	PathTime   int               `json:"pathTime"`
	Arrival    int64             `json:"arrival"`
	ReadOnly   bool              `json:"readOnly"`
	Message    string            `json:"message"`
}

type BattlePatrolPreview struct {
	Troop       BattleTroopDetail `json:"troop"`
	Target      BattleTroopDetail `json:"target"`
	TargetID    int               `json:"targetId"`
	TargetName  string            `json:"targetName"`
	TargetCity  string            `json:"targetCity"`
	Report      string            `json:"report"`
	ReportLines []string          `json:"reportLines"`
	PigeonCount int64             `json:"pigeonCount"`
	Blocked     bool              `json:"blocked"`
	ReadOnly    bool              `json:"readOnly"`
	Message     string            `json:"message"`
}

type BattleMemberRow struct {
	ID        int    `json:"id"`
	UID       int    `json:"uid"`
	Name      string `json:"name"`
	Camp      string `json:"camp"`
	State     string `json:"state"`
	HeroCount int    `json:"herocount"`
	Honour    int64  `json:"honour"`
	Cancel    bool   `json:"cancel"`
	InviteID  int    `json:"inviteId"`
	Invited   bool   `json:"invited"`
}

type BattleMembersSnapshot struct {
	Rows      []BattleMemberRow `json:"rows"`
	InCount   int               `json:"inCount"`
	IsCreator bool              `json:"isCreator"`
	ReadOnly  bool              `json:"readOnly"`
	Message   string            `json:"message"`
}

type BattleFieldNewsPage struct {
	Page      int                   `json:"page"`
	PageSize  int                   `json:"pageSize"`
	Total     int                   `json:"total"`
	PageCount int                   `json:"pageCount"`
	Items     []BattleFieldNewsItem `json:"items"`
	ReadOnly  bool                  `json:"readOnly"`
	Message   string                `json:"message"`
}

type BattleFieldState struct {
	FieldName     string                    `json:"fieldName"`
	CID           int                       `json:"cid"`
	BattlefieldID int                       `json:"battlefieldId"`
	BID           int                       `json:"bid"`
	UnionID       int                       `json:"unionId"`
	CanSend       bool                      `json:"canSend"`
	Info          BattleFieldInfo           `json:"info"`
	NewsTotal     int                       `json:"newsTotal"`
	News          []BattleFieldNewsItem     `json:"news"`
	Rows          []BattleFieldTroopRow     `json:"rows"`
	Cities        []BattleFieldCityInfo     `json:"cities"`
	CurrentTroops []BattleFieldCurrentTroop `json:"currentTroops"`
	WinPoints     []BattleFieldWinPoint     `json:"winPoints"`
}
