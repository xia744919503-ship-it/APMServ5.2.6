# 热血三国一比一复刻重构规划

## 结论先行

后续重构不能继续按“现代化管理后台 + 部分旧库读取”的方向做。这个方向只能做出“很像一个三国页游后台”，做不出一比一复刻。

新的目标必须改成：

1. 后端统一使用 Go。
2. 旧 PHP 只作为规格样本和行为对照，不再继续扩展。
3. `BloodWar.swf`、旧 AMF 协议、旧数据库状态变更、旧资源图片共同组成复刻基准。
4. 所有登录、注册、建号、城池、建筑、科技、征兵、地图、出征、战斗、战报等行为，必须先有“金样本”，再写 Go 实现。
5. 当前单机服里的魔改数值、测试赠送、半成品修复逻辑不能直接视为原版标准。

## 当前项目判断

可信输入：

- `D:\APMServ5.2.6\www\htdocs\BloodWar.swf`
- `D:\APMServ5.2.6\www\htdocs\common2.swf`
- `D:\APMServ5.2.6\www\htdocs\images`
- `D:\APMServ5.2.6\www\htdocs\server\amfphp\gateway.php`
- `D:\APMServ5.2.6\www\htdocs\server\game\*.php`
- `D:\APMServ5.2.6\MySQL5.1\data\bloodwar`

当前 Go/Vue 工程可保留的部分：

- `backend/internal/legacy` 里已经做过的旧库读取、城池归属校验、部分写操作。
- `backend/internal/server` 里已有的本地服务框架。
- `frontend/src/assets` 里已经搬过来的旧资源。
- `scripts/smoke-api.ps1` 这种可重复验证脚本。

必须降级为“临时参考”的部分：

- 现代 dashboard 风格页面。
- 与旧 AMF 返回结构不一致的自定义 JSON 结构。
- 靠截图硬贴但没有真实状态覆盖的假 1:1 界面。
- 当前内置 fallback 样例数据。
- 只验证“接口能跑”，不验证“行为和旧端一致”的烟测。

## 一比一的定义

一比一不是“布局像”“接口差不多”“玩法能点”。一比一需要同时满足四层：

### 1. 协议一比一

- 旧登录入口：`Login.doLogin`。
- 旧命令入口：`Command.sendCommand`。
- 返回数组的顺序、类型、错误消息、状态码含义必须一致。
- 旧客户端依赖的字段名、下标、空值表现必须一致。

Go 后端可以额外提供现代 JSON API，但现代 API 只能是适配层，不能成为玩法标准。真正标准应来自旧 AMF/命令协议。

### 2. 数据一比一

- 优先沿用旧表结构：`sys_user`、`sys_city`、`mem_city_resource`、`mem_world`、`sys_building`、`sys_city_soldier`、`sys_city_hero`、`sys_troops`、`sys_battle`、`sys_report` 等。
- Go 写操作产生的数据库 diff，要和旧 PHP 在同一初始快照下产生的 diff 对齐。
- 新增表只能做旁路审计、迁移记录、测试夹具，不得改变旧玩法表语义。

### 3. 时序一比一

- 建筑完成、科技完成、征兵完成、部队到达、部队返回、战斗结算，都要按旧逻辑的触发时机复刻。
- 旧端是懒结算还是定时结算，必须先记录，再决定 Go 里用同等行为还是兼容封装。
- 所有时间相关测试必须可注入固定时间，不能靠当前系统时间碰运气。

### 4. 前端一比一

- 主舞台尺寸、旧按钮状态、弹窗位置、tab 切换、提示文案、错误弹窗、等待状态都要以旧 SWF 运行效果为准。
- 前端不能再自由发挥成“比较像”的页面。
- 可使用 Vue 重新实现，但每个页面必须有旧端截图作为参考，并通过截图 diff 验收。

## Go 后端目标架构

建议后端重构成以下结构：

```text
backend/
  cmd/api/
  internal/config/
  internal/transport/
    http/          # 现代 JSON API，仅作调试和新前端适配
    amf/           # 旧客户端协议兼容入口，目标是复刻 Login/Command
  internal/app/
    account/
    city/
    resource/
    building/
    technic/
    soldier/
    hero/
    troop/
    battle/
    report/
    mail/
    union/
    shop/
    task/
  internal/legacydb/
    repository.go  # 只做 SQL 和旧表映射
  internal/rules/
    clock.go       # 可注入时间
    formulas.go    # 公式集中管理
  internal/parity/
    fixtures/
    golden/
    tests/
```

核心原则：

- transport 层只负责协议，不写玩法。
- app 层负责状态机和用例。
- legacydb 层负责旧表 SQL，不夹杂业务判断。
- rules 层放公式、时间、数值规则，必须可测试。
- parity 层是生命线，所有重写模块都要有金样本对比。

## 复刻方法

### 第一步：建立金样本

每个功能都要用同一套流程采样：

1. 还原旧 PHP + 旧 MySQL 快照。
2. 用旧 SWF 执行动作。
3. 记录 AMF 请求和响应。
4. 记录动作前后的数据库 diff。
5. 截图记录 UI 初始态、操作中、成功态、失败态。
6. 把样本固化到 `docs/specs` 和 `backend/internal/parity/fixtures`。

没有金样本的功能，不允许标记为一比一完成。

### 第二步：Go 实现同一行为

Go 实现必须对同一输入产生：

- 同样的返回结构。
- 同样的数据库变化。
- 同样的错误消息。
- 同样的时间推进结果。
- 同样的副作用，例如邮件、战报、告警、任务进度。

### 第三步：前端复刻旧交互

前端不是先做新页面，而是按旧端页面分解：

- 固定舞台。
- 背景层。
- 按钮层。
- 热区层。
- 弹窗层。
- 文本和状态覆盖层。

视觉还原可以复用旧图片，也可以从旧截图切片，但动态数据必须由真实状态覆盖，不能长期停留在静态截图。

## 分阶段规划

### Phase 0：冻结当前方向，建立复刻基线

目标：

- 明确当前 `new_project` 不是最终形态，只是旧库适配和资源试验场。
- 建立一比一复刻目录、金样本格式、验收规则。

交付：

- `docs/one-to-one-reconstruction-plan.md`
- `docs/specs/protocol-inventory.md`
- `docs/specs/ui-screens.md`
- `docs/specs/db-diff-rules.md`
- `backend/internal/parity` 测试骨架

验收：

- 能列出旧 `Login` 和 `Command` 的入口。
- 能列出第一批必须采样的 AMF 命令。
- 能对一个最小动作生成“请求、响应、DB diff、截图”四件套。

### Phase 1：协议入口和账号体系

优先级最高，因为登录、注册、建号是所有玩法入口。

必须复刻：

- `Login.getLoginAnnouncement`
- `Login.doLogin`
- `Login.checkQueue`
- 新用户注册流程。
- 创建角色：`UserFunc.php::createRole`
- 创建城池：`UserFunc.php::doCreateCity`
- `sys_sessions`、`sys_online`、`mem_queue` 行为。
- 版本检查、服务器状态检查、排队逻辑、封禁逻辑、未建城状态。

Go 实现：

- 不再只支持当前的 uid 快捷登录。
- 必须支持 passport/password/passtype 语义。
- 必须返回旧端期望的数组结构。
- 现代 cookie session 可以保留给 Vue，但不能替代旧 `uid + sid` 协议。

验收：

- 老样本中成功登录、错误密码、旧版本、服务器维护、排队、新号未建城的响应完全匹配。
- 建号后 `sys_user`、`sys_city`、`mem_city_resource`、`sys_building`、`sys_city_soldier` 等初始数据 diff 匹配金样本。

### Phase 2：城池基础循环

必须复刻：

- `CityFunc.php::getCityInfo`
- `CityFunc.php::getCityBaseInfo`
- `CityFunc.php::getCityProduct`
- `CityFunc.php::setCityProductRate`
- `CityFunc.php::changeTax`
- `utils.php::doGetCityBaseInfo`
- `utils.php::doGetCityAllInfo`
- `utils.php::updateCityResourceAdd`
- `utils.php::upbuilding`

重点：

- 资源产量不能再出现前端显示和后台累计两套算法。
- 懒结算触发点要和旧端一致。
- 税率、民心、人口、黄金、资源上限要按旧公式。

验收：

- 同一城池同一时间点，旧 PHP 和 Go 的资源、人口、税率、产量、告警一致。
- 修改税率、修改生产比例后的返回和 DB diff 一致。

### Phase 3：建筑、科技、征兵、城防时序

必须复刻：

- `BuildingFunc.php`
- `BuildingCron.php`
- `TechnicFunc.php`
- `TechnicCron.php`
- `SoldierFunc.php`
- `SoldierCron.php`
- `DefenceFunc.php`

重点：

- 建筑升级、取消、拆除、创建。
- 科技研究、取消、完成。
- 征兵排队、取消、完成。
- 城防建造、解散、加速。
- 队列数量、资源扣除、取消返还、加速道具。

验收：

- 开始动作的资源扣除一致。
- 取消动作的返还一致。
- 到点结算的目标等级、士兵数量、科技等级一致。
- 错误条件文案一致，例如资源不足、前置不足、队列已满。

### Phase 4：英雄、装备、道具、buff

必须复刻：

- `HeroFunc.php`
- `HotelFunc.php`
- `OfficeFunc.php`
- `ArmorFunc.php`
- `EquipmentFunc.php`
- `GoodsFunc.php`
- `BufferFunc.php`
- `BarnFunc.php`

重点：

- 招募、解雇、任命太守/将军/军师。
- 英雄属性点、经验、忠诚、体力、精力。
- 装备穿戴、卸下、修理、翻新、回收。
- 道具使用、礼包、buff 对建筑/征兵/战斗公式的影响。

验收：

- 任命后的 `sys_city`、`sys_city_hero`、`mem_city_resource` 更新一致。
- 道具减少、buff 生效时间、属性加成一致。

### Phase 5：世界地图与出征

必须复刻：

- `WorldFunc.php`
- `TroopFunc.php`
- `utilsExtend.php`
- 相关 `sys_troops`、`mem_world`、`sys_alarm`、战报写入。

重点：

- 地图格子信息、城池信息、野地信息。
- 侦察、掠夺、占领、运输、派驻、召回。
- 行军时间、负重、粮耗、目标合法性。
- 到达、返回、驻守、采集的状态流。

验收：

- 同一兵种、数量、目标，行军时间和粮耗一致。
- 出征后士兵扣除、部队记录、告警一致。
- 召回、到达、返回后的 DB diff 一致。

### Phase 6：战斗与战报

必须复刻：

- `BattleFunc.php`
- `BattleCommand.php`
- `cfg_js.php`
- `bscfg_js.php`
- `ReportFunc.php`
- `report.php`

重点：

- 战场创建。
- 部队进入战场。
- 回合推进。
- 兵种速度、射程、攻击、防御、生命、阵型、战术。
- 战损、掉落、经验、声望、荣誉。
- 战报 HTML 内容和标题。

验收：

- 固定随机种子下，同一战斗输入得到同一回合结果。
- 战损、奖励、战报内容与金样本一致。
- 不能只做“简化战斗”，否则不算一比一。

### Phase 7：外围系统

必须复刻：

- `MailFunc.php`
- `TaskFunc.php`
- `UnionFunc.php`
- `ShopFunc.php`
- `RankFunc.php`
- `FriendFunc.php`
- `MarketFunc.php`
- `RewardFunc.php`
- `TrickFunc.php`
- `StatFunc.php`

顺序建议：

1. 邮件和战报，因为它们是其他系统副作用。
2. 任务，因为它被登录、建筑、征兵、联盟、战斗触发。
3. 商城/道具，因为它影响加速、buff、装备。
4. 联盟、好友、排行、市场。
5. 计谋、统计、活动奖励。

验收：

- 每个模块至少有读取、成功写入、失败写入、跨模块副作用四类样本。

### Phase 8：前端全量一比一

必须复刻页面：

- 登录公告。
- 登录框。
- 排队页。
- 注册/建号/建城。
- 主城内城。
- 外城/资源田。
- 世界地图。
- 建筑弹窗。
- 科技弹窗。
- 兵营弹窗。
- 英雄/客栈/官府/装备。
- 部队/战场/战报。
- 邮件、任务、联盟、商城、排行。

验收：

- 固定 `1000x600` 基准截图 diff。
- 关键按钮 hover/down/on 状态一致。
- 弹窗位置、尺寸、遮罩、关闭行为一致。
- 同一操作链 UI 状态跳转一致。
- 动态文字不能遮挡旧 UI，不能靠现代组件风格替代旧窗口。

## 具体执行顺序

第一批必须做小闭环：

1. 登录公告。
2. 账号登录。
3. 新账号注册。
4. 创建角色。
5. 创建第一座城。
6. 进入主城。
7. 读取城池基础信息。
8. 修改税率。
9. 修改生产比例。

这批跑通以后，才继续建筑和征兵。原因是它覆盖了协议、会话、DB 写入、UI 入口，是整个复刻工程的地基。

第二批：

1. 建筑信息。
2. 建筑升级。
3. 取消升级。
4. 到点完成。
5. 科技研究。
6. 征兵。
7. 取消征兵。

第三批：

1. 世界地图。
2. 目标城池/野地详情。
3. 出征。
4. 召回。
5. 到达结算。
6. 返回结算。
7. 战报生成。

第四批：

1. 战斗创建。
2. 战斗回合。
3. 战斗结束。
4. 战损与奖励。
5. 战报一比一。

## 验收标准

每个功能完成必须同时满足：

- Go 单元测试通过。
- Go 集成测试通过。
- 与旧 PHP 金样本的 AMF/返回结构对比通过。
- 与旧 PHP 金样本的 DB diff 对比通过。
- 前端截图对比通过。
- 手动完整链路通过。

不满足这些，只能标记为：

- `read-only`
- `approximate`
- `prototype`
- `needs-golden`

不能标记为 `1:1 done`。

## 立刻要停掉的做法

- 不再新增“看起来高级”的现代 dashboard。
- 不再用自定义 JSON 结构替代旧协议作为唯一实现。
- 不再靠静态截图冒充动态 UI。
- 不再把当前魔改单机服数值直接当原版。
- 不再先写一大堆功能再回头验收。

## 立刻要开始的做法

- 每个功能先采样，再实现。
- 每次实现只打一个小闭环。
- 每个闭环都有 DB 快照和 diff。
- 每个页面都有旧端截图。
- 每个 Go 写操作都能回放测试。

## 成功状态

最终成功状态不是“新项目能运行”，而是：

- 原旧端的核心操作链能在 Go 后端上按同样规则运行。
- 新 Vue 前端在视觉和交互上与旧 SWF 运行效果一致。
- 旧数据可以被 Go 服务稳定读取和写入。
- 所有核心玩法都有金样本测试保护。
- 任何人改动建筑、征兵、出征、战斗公式时，测试能立刻发现偏差。

