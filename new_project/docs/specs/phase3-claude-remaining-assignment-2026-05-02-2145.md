# Phase 3 Claude 剩余任务指派单

审计时间：2026-05-02 21:45 Asia/Shanghai

## 当前完成度

正式 tracker 完成度：1 / 13。

已完成：

- 登录/排队/建角链路闭环。

有进展但不能算完成：

- `backend/internal/legacy/city_write.go` 已把税率更新改回旧 PHP 方向：只写 `tax` 和 `morale_stable`，不再直接写当前 `morale`。
- `scripts/smoke-economy-formula-verification.ps1` 已经不再把明显错误全部放过，最新产物 `artifacts/economy-formula-verification-20260502-213350.*` 为失败。
- `scripts/smoke-battle-full-flow.ps1` 已经不再把没兵、400、无战报当成功，最新产物 `artifacts/battle-flow-test-20260502-213943.json` 为失败。

仍未完成：

- 城池经济结算 1:1。
- 建筑/科技/征兵强验收。
- 出征与战斗 1:1。
- 战报字段与语义 1:1。
- 旧 PHP vs Go 对照基线。
- 任务/联盟/商城/充值/邮件/排行/主界面强验收。

## P0-1：修正经济验证脚本的证据口径

当前问题：

- 脚本写的是 `morale_stable`，但实际取的是 API 的 `morale` 字段。
- `CityDetail` 目前没有返回 `moraleStable`。
- `Add-Formula` 本身仍不影响 `allPassed`，如果未来某个公式没有对应 `Add-Check`，可能再次出现报告自相矛盾。
- `People Max` 的 expected 来自 API building 列表；如果列表缺失或样本城没有房屋，会得到 `expected=0 actual=100`，这个失败是对的，但还需要定位到底是样本数据问题、API 字段问题，还是旧库数据问题。

要求：

1. 给经济验证脚本补真实 `morale_stable` 来源：
   - 方案 A：Go `CityDetail` 增加 `moraleStable` 字段，并从 `mem_city_resource.morale_stable` 读取。
   - 方案 B：验证脚本直接查询数据库快照。
2. 税率 PATCH 验收必须断言：
   - `tax` 改为目标值。
   - 当前 `morale` 不被 PATCH 直接改动。
   - `morale_stable = clamp(100 - tax - complaint, 0, 100)`。
3. `Add-Formula` 的 expected/actual 不一致也必须让脚本失败。
4. `People Max` 必须同时输出：
   - API buildings 中 bid=5 的列表。
   - DB `sys_building where cid=? and bid=5` 的列表。
   - DB `mem_city_resource.people_max`。
   - 旧 PHP 公式结果。

验收：

- 重新生成 `artifacts/economy-formula-verification-*.json/md`。
- 报告内 formula 状态、checks 状态、`allPassed` 必须一致。

## P0-2：实现旧版完整经济结算，不要只做局部公式

旧版依据：

- `www/htdocs/server/game/sevice.php:40-174`
- `www/htdocs/server/game/CityFunc.php:75-165`
- `www/htdocs/server/game/common.php:73-77`
- `www/htdocs/server/game/utils.php:465-483`

当前 Go `recalculateCityProduction` 只覆盖：

- 工人需求。
- 分配比例。
- 人口不足倍率。
- 基础四资源产出。

缺失：

- 固定自然产出 `+100`。
- 科技加成。
- 太守/城守内政加成。
- 道具 buff 加成。
- 军队耗粮。
- 金币增长。
- `lastupdate` 时间差结算。
- 人口向稳定值推进。
- 资源读城池时主动结算。

要求：

1. 在 Go 中新增或补全旧版资源结算函数。
2. `CityDetail` 或对应读城池入口必须先按 `lastupdate` 结算资源。
3. 经济验证脚本必须比较 T0/T60/T300：
   - expected 来源于旧 PHP 公式复写、旧 PHP 接口或旧库快照。
   - actual 来源于 Go API/DB。
   - 输出 delta。
4. 不允许只拿 Go 代码自身公式当 expected。

验收：

- `go build -mod=mod -o bin/api.exe ./cmd/api` 通过。
- 经济脚本通过。
- tracker 的“城池经济结算 1:1 对齐”只有在上述证据齐全后才能打勾。

## P0-3：战斗脚本必须准备真实测试数据

当前最新战斗产物失败原因：

- 测试城没有士兵。
- dispatch 返回 400。
- 没有 troopId。
- callback 没执行。
- 没有战报。

要求：

1. 脚本启动时必须准备一个有兵力的测试城：
   - 优先走注册/建角/征兵 API。
   - 如果 API 链路太长，允许建立明确标注的 DB fixture，但必须写在脚本里，不能依赖手工前置状态。
2. dispatch 必须返回 200 和有效 `troopId`。
3. 必须实际执行：
   - dispatch
   - troops list 状态变化
   - callback 或到达/返回流程
   - reports unread/all/detail
4. 必须校验：
   - `sys_city_soldier` 前后变化。
   - `sys_troops` 的 `cid/startcid/targetcid/task/state/starttime/endtime/soldiers/resource`。
   - 战报 `title/type/time/read/battleid/content/origincid/happencid`。
5. P0 脚本里禁止把 400 当成功。

验收：

- `scripts/smoke-battle-full-flow.ps1` 能在干净样本上通过。
- 产物里不能有 `skipped` 关键步骤。

## P0-4：旧 PHP vs Go 对照基线必须重新跑通

当前只有旧产物：

- `artifacts/legacy-rule-diff-20260502-165913.json`

这份旧端 `8088` 没连上，不能作为验收证据。

要求：

1. 启动旧 PHP 服务到 `http://127.0.0.1:8088`，或改脚本直接调用旧 PHP/旧库快照。
2. 每个对照项输出：
   - oldStatus
   - newStatus
   - oldBody/raw
   - newBody/raw
   - normalizedDiff
   - legacy source file/line
   - go source file/line
3. 旧端不可达时脚本必须失败。

验收：

- 重新生成新的 `artifacts/legacy-rule-diff-*.json`。
- 不能再出现 old server unreachable。

## P1：队列系统强验收

范围：

- 建筑队列。
- 科技队列。
- 征兵队列。

要求：

1. 以旧 PHP 为基线，比对 start/cancel/finish 的资源、状态、时间、返还比例。
2. 覆盖资源不足、队列忙、非法 position/bid/tid/sid。
3. 产物要能证明字段与状态机一致。

## P1：外围模块强验收

范围：

- 任务系统。
- 联盟系统。
- 商城/充值兑换。
- 邮件系统。
- 排行系统。
- 主界面/地图/功能窗关键交互。

要求：

1. 不能只 smoke read endpoint。
2. 每个模块至少有读、成功写、失败写、边界权限四类证据。
3. 对照旧 PHP 的返回结构、状态码、失败语义。

## Tracker 更新规则

不要提前改 `phase3-1to1-tracker.md`。

只有满足以下条件才能把某项改成 `[x]`：

1. 有旧 PHP 或旧 DB 快照作为 expected。
2. 有 Go 源码实现位置。
3. 有脚本产物证明通过。
4. 脚本失败时 exit code 非 0。
5. 关键步骤没有 skipped。
6. 报告没有 PASS/FAIL 自相矛盾。

当前可保持：

- 登录/排队/建角链路 `[x]`。

其余 12 项继续保持 `[ ]`。
