# Phase 3 Claude 加速闭环指令

审计时间：2026-05-03 00:08 Asia/Shanghai

## 当前真实进度

正式 tracker：1 / 13。

已闭环：

- 登录/排队/建角链路。

这轮新增的有效进展：

- `CityDetail` 已补 `moraleStable`。
- `UpdateCityTax` 已对齐旧 PHP 方向：PATCH 税率只更新 `tax` 与 `morale_stable`，不直接改当前 `morale`。
- Go 当前源码可编译。
- 战斗脚本已能准备士兵，并且失败时 exit 1。
- 经济脚本已能识别 `peopleMax` 样本问题，并且失败时 exit 1。

还不能算完成：

- 城池经济结算仍未 1:1。
- 出征/战斗/战报仍未 1:1。
- 旧 PHP vs Go 对照仍未跑通。

## 今天不要再发散

目标不是继续补大文档，而是先把 tracker 从 1/13 推到 2/13 或 3/13。

只做下面三件事，做完再开新模块。

## P0-1：修经济报告自相矛盾

文件：

- `scripts/smoke-economy-formula-verification.ps1`

当前 bug：

- `Add-Formula` 已经把 `expected/actual/passed` 放到 `$f.expected/$f.actual/$f.passed`。
- Markdown 生成处仍然用 `$f.result.expected` 和 `$f.result.actual`。
- 结果是 People Max 公式表显示 PASS，但 Checks 显示 FAIL。

必须修：

```powershell
foreach ($f in $formulasList) {
  $status = if ($f.passed) { 'PASS' } else { 'FAIL' }
  $md += "`n| $($f.name) | $($f.goFile) | $status |"
}
```

验收：

- 重新跑经济脚本。
- 如果 People Max 仍不匹配，Formula 和 Check 必须同时 FAIL。
- 报告不能再出现 Formula PASS / Check FAIL 的矛盾。

## P0-2：先用正确样本闭环 People Max

当前最新经济失败：

- `Data: Test Data: Houses Exist expected >0 actual 0`
- `People Max Formula expected 0 actual 100`

说明当前样本城 `266010` 没有 `bid=5` 民房，不适合验 peopleMax。

必须二选一：

1. 优先选择已有民房的城市作为经济样本：

```sql
select c.cid
from sys_city c
join sys_building b on b.cid = c.cid and b.bid = 5
join mem_city_resource r on r.cid = c.cid
where c.uid = 1010
limit 1;
```

2. 如果 `test` 账号没有这种城，就在脚本 fixture 中插入一个最小民房样本，并同步 `mem_city_resource.people_max`：

```sql
insert into sys_building (cid, bid, level, xy)
values ($CityId, 5, 1, <free_xy>);

update mem_city_resource
set people_max = 1 * (1 + 1) * 500
where cid = $CityId;
```

要求：

- fixture 必须写在脚本里，不能靠手工预置。
- 输出 API houses、DB houses、DB people_max、formula expected。
- People Max 通过后，经济脚本仍然只能声明“经济公式样本闭环”，不能直接把完整“城池经济结算 1:1”打勾，除非完整 `sevice.php` 资源结算也做完。

## P0-3：战斗脚本改成最小侦察闭环

当前战斗失败不是深层战斗问题，是参数错：

- 脚本默认 `TargetCityId = 266010`，等于自己的 `CityId`。
- 脚本用 `task=2`，但 Go 里 `task=2` 只允许侦察兵 `sid=3`。
- 脚本选到了 `sid=2`，所以 dispatch 必然 400。

必须修：

1. 选择非本人、非本城、存在的目标：

```sql
select cid
from sys_city
where cid <> $CityId and uid <> $uid
limit 1;
```

如果没有城市目标，用 `mem_world` 选一个 `type > 0` 的野地，并按 Go 的 `cidToWid` 映射生成可用 targetCid。

2. 对 `task=2` 固定使用侦察兵：

```powershell
$availableSid = 3
$dispatchTask = 2
```

3. fixture 必须确保源城有侦察兵：

```sql
insert into sys_city_soldier (cid, sid, count)
values ($CityId, 3, 50)
on duplicate key update count = greatest(count, 50);
```

4. dispatch 必须满足：

- status = 200
- troopId > 0
- `sys_troops.startcid = $CityId`
- state 从 0 到 1 或按当前实现进入返回链路
- 产生 attacker scout report

5. 等待最小路径时间：

```powershell
Start-Sleep -Seconds 6
```

然后调用 `/api/me/troops` 或其他会触发 `settleDueTroops` 的接口，再查 reports。

验收：

- `scripts/smoke-battle-full-flow.ps1` 退出码 0。
- 最新 `artifacts/battle-flow-test-*.json` 不能有关键步骤 skipped。
- dispatch/callback/report/soldier_change 都必须 passed。

## P0-4：清掉临时 debug 日志

当前源码里有大量 `DEBUG` 日志：

- `backend/internal/server/session.go`
- `backend/internal/server/router.go`
- `backend/internal/legacy/troop_write.go`

闭环前必须删除或改成受环境变量控制。

要求：

- 不要在正常服务日志里打印 session token/cookies。
- 不要提交 `DEBUG currentUID`、`DEBUG sessions`、`DEBUG handleCityTroopDispatch` 这类日志。

## P0-5：旧 PHP 对照不要继续拖

目前最新 `legacy-rule-diff` 还是：

- `artifacts/legacy-rule-diff-20260502-165913.json`

这份旧端不可达，不算。

最短闭环要求：

- 要么启动旧 PHP 到 `8088`，重新跑 `compare-legacy-rules.ps1`。
- 要么把 expected 改成直接从旧 PHP 源码公式和 DB 快照产生。

完成标准：

- 新增一份 `artifacts/legacy-rule-diff-*.json`。
- oldStatus/newStatus 都有真实值。
- 不允许 old server unreachable。

## Tracker 规则

现在不要改大项。

可以新增“小闭环证据”备注，但只有满足这些条件才能打 `[x]`：

- 脚本 exit 0。
- artifact allPassed true。
- 没有关键 skipped。
- 报告无 PASS/FAIL 矛盾。
- 有旧 PHP 源码/旧 DB 依据。

今天优先达成：

1. 经济公式样本报告干净通过。
2. 侦察 dispatch -> settle -> report 最小闭环通过。
3. 旧 PHP 对照恢复可运行。
