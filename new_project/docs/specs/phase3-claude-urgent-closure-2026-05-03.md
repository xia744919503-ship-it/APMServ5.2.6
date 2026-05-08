# Phase 3 Claude Urgent Closure Instructions

生成时间：2026-05-03

## 现状判定

真实进度仍按 **1/13** 计算。

原因不是没有写代码，而是新增内容没有形成可验收闭环：

- tracker 只勾选了“登录/排队/建角链路闭环”。
- 当前 Go 后端编译失败：`backend/internal/legacy/troop_write.go` 有未使用的 `log` import。
- 经济验证接近完成，但 `People Max` 仍未稳定通过。
- 战斗闭环仍失败：dispatch 返回 `400`，没有有效 `troopId`，没有侦察战报。
- 旧 PHP vs Go 对照仍不可用：旧端 `http://127.0.0.1:8088` 不可达，diff 产物没有实际测试数量。

从现在开始，不再扩展新模块，不再新增大范围设计，不再做“看起来像完成”的实现。只做最短闭环。

## 总目标

把真实进度从 **1/13** 推到至少 **3/13**。

本轮只允许关闭这两个模块：

1. 城池经济结算 1:1 对齐
2. 出征与战斗 + 战报字段与语义最小闭环

其他模块先不要碰。

## P0：先恢复可编译

必须第一步完成。

### 要做

- 修复 `backend/internal/legacy/troop_write.go` 未使用 import。
- 不要顺手重构。
- 不要改无关文件。

### 验收命令

```powershell
cd D:\APMServ5.2.6\new_project\backend
$env:GOCACHE='D:\APMServ5.2.6\new_project\backend\.gocache'
go build -mod=mod -o bin\api-check.exe .\cmd\api
```

必须通过。未通过之前，不允许继续写业务逻辑。

## P1：关闭经济结算

当前经济不是全失败，而是卡在 `People Max`。不要重写经济系统，只修最后失败项。

### 已经通过的内容

最新经济报告中以下内容已经 PASS：

- Morale Stable Formula
- Food Production Rate
- Tax PATCH Method
- Tax PATCH Set Value
- Morale Stable Updated After Tax
- Morale Unchanged After Tax
- Production PATCH Method
- Production PATCH Set Values
- Food Production T+60
- Food Monotonic T+300

### 当前失败点

- `Test Data: Houses Exist`
- `People Max Formula`

### 要做

- 让经济脚本稳定准备一个有民房的测试城池。
- fixture 必须写入真实会影响 `People Max` 的旧表字段。
- 不允许只改报告 expected 值来假通过。
- 如果脚本插入了民房，必须重新读取 city detail，并确保 `summary.resources.peopleMax` 与旧公式一致。
- 如果 Go 逻辑有错，只修 `People Max` 相关逻辑，不要扩大到完整建筑系统。

### 验收命令

```powershell
cd D:\APMServ5.2.6\new_project
powershell -ExecutionPolicy Bypass -File .\scripts\smoke-economy-formula-verification.ps1 -BaseUrl http://127.0.0.1:8080
```

验收标准：

- 脚本退出码为 0。
- 最新 `artifacts\economy-formula-verification-*.json` 中 `allPassed=true`。
- 报告中 `People Max` 为 PASS。
- tracker 才能勾选“城池经济结算 1:1 对齐”。

## P2：关闭最小战斗/战报闭环

当前战斗卡在 dispatch `400`。不要先做完整战斗模拟，先做最小侦察链路。

### 当前失败点

最新战斗脚本状态：

- login PASS
- check_soldiers PASS
- troops_before PASS
- dispatch FAIL，status=400
- troopId=0
- callback 因 no troopId 跳过/失败
- reports API 可访问，但无报告
- soldier_change FAIL

### 要做

- 先查清 dispatch 返回 400 的具体原因。
- 在脚本里打印服务端返回 body，但不要打印 token/cookie。
- 侦察任务必须使用：
  - `task=2`
  - `sid=3`
  - 目标城池不能等于自己的城池
  - 目标城池必须存在
  - 出发城池必须有对应校场/出征条件
- 如果是参数名不匹配，修脚本。
- 如果是 Go 接口校验与旧 PHP 不一致，修 Go。
- 如果是 fixture 数据缺字段，补 fixture。
- 不允许绕过 dispatch 直接插 troop/report 假通过。

### 验收命令

```powershell
cd D:\APMServ5.2.6\new_project
powershell -ExecutionPolicy Bypass -File .\scripts\smoke-battle-full-flow.ps1 -BaseUrl http://127.0.0.1:8080
```

验收标准：

- dispatch 返回 200。
- 返回有效 `troopId > 0`。
- callback/settle 后能读到侦察战报。
- 战报结构检查 PASS。
- 最新 `artifacts\battle-flow-test-*.json` 中关键步骤全部 PASS。

完成这个闭环后，tracker 最多只能勾：

- 出征与战斗 1:1 对齐
- 战报字段与语义 1:1 对齐

如果只是最小侦察闭环，不要声称完整战斗系统已 1:1。

## P3：旧 PHP 对照只做恢复，不做借口

当前 `legacy-rule-diff` 无效，因为旧端不可达：

- `oldServerReachable=false`
- `totalTests=0`

### 要做

- 要么启动并修通旧 PHP 端 `http://127.0.0.1:8088`。
- 要么明确记录旧端不可达，并不要把 diff 计入完成度。

这项不能阻塞 P1/P2，但也不能拿空 diff 当验收证据。

## 禁止事项

- 禁止继续新增外围模块。
- 禁止改 tracker 勾选但没有对应 PASS 产物。
- 禁止为了 PASS 修改 expected 值。
- 禁止吞掉错误或把失败项改成 skipped。
- 禁止打印敏感 token/cookie。
- 禁止大范围重构。
- 禁止把“脚本跑过一半”算完成。

## 交付格式

完成后只提交以下证据：

1. `go build` 通过的输出摘要。
2. 最新经济验证 artifact 文件名，以及 `allPassed=true`。
3. 最新战斗验证 artifact 文件名，以及 dispatch/report 全 PASS。
4. 修改了哪些文件，每个文件一句话说明。
5. tracker 新增勾选了哪些项。

没有这些证据，就不要说完成。

