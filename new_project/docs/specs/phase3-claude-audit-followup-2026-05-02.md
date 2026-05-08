# Phase 3 Claude 跟进审计与修正指引

审计时间：2026-05-02 21:15 Asia/Shanghai

## 当前结论

这轮有进展，但还不能算 1:1 复刻闭环。

已确认：

- 新增了 `scripts/smoke-economy-formula-verification.ps1`。
- 新增了 `scripts/smoke-battle-full-flow.ps1`。
- 新增了 `backend/artifacts/economy-formula-verification-20260502-203117.*`。
- 新增了 `backend/artifacts/battle-flow-test-*.json`。
- 修改了 `backend/internal/legacy/city_write.go`。
- Go 后端可编译通过。

未达标：

- `docs/specs/phase3-1to1-tracker.md` 仍然只有登录/排队/建角链路为 `[x]`，其余 12 项仍未闭环。
- 经济验证脚本存在 false positive。
- 战斗流程脚本存在 false positive。
- `city_write.go` 的税率/民心行为与旧 PHP 疑似不一致。
- 当前经济公式没有覆盖旧版完整资源结算。

## P0：修正税率行为，不要改错旧版语义

旧 PHP：

- `www/htdocs/server/game/CityFunc.php:216-228`
- `www/htdocs/server/amfphp/game/CityFunc.php:216-228`

旧版 `changeTax` 只做：

```sql
update mem_city_resource set tax='$newtax' where cid='$cid';
update mem_city_resource set morale_stable=GREATEST(0,LEAST(100-tax-complaint,100)) where cid='$cid';
```

它没有直接更新当前 `morale`。

当前 Go：

- `backend/internal/legacy/city_write.go:42-57`

Go 现在同时写了 `morale` 和 `morale_stable`：

```sql
update mem_city_resource set morale = ?, morale_stable = ? where cid = ?
```

要求：

1. 对齐旧版：设置税率时只改 `tax` 和 `morale_stable`。
2. 不要在 `UpdateCityTax` 里直接把当前 `morale` 改成稳定值，除非能拿出旧 PHP 同步行为证据。
3. 补验证：改税前记录 `tax/morale/morale_stable/complaint`，改税后断言 `morale` 不变、`morale_stable = clamp(100-tax-complaint, 0, 100)`。

## P0：经济公式不能只验局部

旧版完整资源结算主要在：

- `www/htdocs/server/game/sevice.php:40-174`
- `www/htdocs/server/game/CityFunc.php:75-165`
- `www/htdocs/server/game/common.php:73-77`

旧版资源结算至少包含：

- 四类资源建筑 `using_people`。
- `GLOBAL_*_RATE` 与 `GAME_SPEED_RATE`。
- 生产分配比例 `food_rate/wood_rate/rock_rate/iron_rate`。
- 人口不足倍率 `bl = min(1, people / people_working)`。
- 固定自然产出 `+100`。
- 科技加成 `sys_city_technic level*10`。
- 太守/城守内政加成。
- 道具 buff 加成。
- 军队耗粮 `food_army_use`。
- 金币增长 `people * GAME_SPEED_RATE * GLOBAL_GOLD_RATE * 0.1 * elapsedHours`。
- 资源按 `lastupdate` 到当前时间结算。
- 人口趋向 `people_max * morale * 0.01`。

当前 Go 的 `recalculateCityProduction` 只覆盖了工人、比例和人口倍率，未覆盖上面多数项，也没有在 `CityDetail` 读取时按 `lastupdate` 结算资源。

要求：

1. 新增或补全 Go 的资源结算函数，语义对齐 `sevice.php:40-174`。
2. `CityDetail` 或对应读城池入口必须触发资源结算，不能只读取旧值。
3. 产物必须输出 T0/T60/T300 的旧版预期、Go 实际、差异。
4. 不能只用 Go 源码公式当 expected；expected 必须来自旧 PHP 公式复写、旧库快照，或旧接口响应。

## P0：经济验证脚本 false positive 必须修

当前问题：

- `scripts/smoke-economy-formula-verification.ps1:112-129` 的 `Add-Formula` 不会影响 `allPassed`。
- `scripts/smoke-economy-formula-verification.ps1:427` 对 `People Max Formula` 使用了：

```powershell
($resT0.peopleMax -eq $expectedPeopleMax -or $expectedPeopleMax -eq 0)
```

这会把 `expected=0 actual=100` 放过。

最新报告已经出现矛盾：

- Markdown 里 `People Max` 是 `FAIL`。
- Checks 里 `People Max Formula expected=0 actual=100` 却是 `PASS`。
- JSON 顶层仍然 `allPassed=true`。

要求：

1. `Add-Formula` 的 expected/actual 不一致必须让脚本失败。
2. 删除 `expectedPeopleMax -eq 0` 放行逻辑。
3. 如果接口没有返回建筑列表或房屋数据，脚本应标记为失败，而不是把 expected 置 0 后通过。
4. 报告里 `allPassed`、Markdown 状态、checks 状态必须一致。

## P0：战斗流程脚本不能跳过也算通过

复跑 `scripts/smoke-battle-full-flow.ps1` 的实际结果：

- 没有士兵。
- dispatch 返回 400。
- 没有 troopId。
- callback 跳过。
- 没有战报。
- 战报结构跳过。
- 士兵变化跳过。
- 最后仍然 `9/9 passed`。

这不能算战斗/战报复刻。

要求：

1. 脚本启动时必须准备或选择一个有士兵的测试城池。
2. 没有士兵时直接失败，不能跳过后计入通过。
3. dispatch 必须返回 200 且有有效 `troopId`。
4. callback 必须实际执行，且校验状态迁移。
5. 必须产生或读取战报，校验 `title/type/time/read/battleid/content/origincid/happencid` 等核心字段。
6. 必须校验派兵前后 `sys_city_soldier`、`sys_troops`、资源携带/返还或掠夺结果。
7. P0 验收脚本中不允许把 400 当成功。

## P0：旧版对照基线仍要补齐

之前 `compare-legacy-rules.ps1` 的产物是旧服务 `8088` 连接失败，不是有效对照。

要求：

1. 先保证旧 PHP 端可请求，或者改成直接用旧 PHP 方法/数据库快照作为 expected。
2. 对每个接口输出：
   - oldStatus
   - newStatus
   - oldBody/raw
   - newBody/raw
   - normalizedDiff
   - legacy source file/line
   - go source file/line
3. 旧端不可达时脚本必须失败，不能生成“可验收”报告。

## Tracker 更新规则

只能在满足下面条件后改 `phase3-1to1-tracker.md`：

1. 有真实旧版依据。
2. 有 Go 实现对应源码。
3. 有脚本产物证明通过。
4. 脚本不存在“跳过也算通过”的 P0 项。
5. 报告里没有自相矛盾的 PASS/FAIL。

当前不要把“城池经济结算”“出征与战斗”“战报字段与语义”改为 `[x]`。
