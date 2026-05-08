# Phase 3 Claude Complete Remaining Plan

生成时间：2026-05-03

## 当前真实进度

当前按可验收证据计算为 **2/13**。

已完成：

- 登录/排队/建角链路闭环
- 城池经济结算 1:1 对齐

不能再按旧的 1/13 口径推进。经济已经有可验收证据：

- `artifacts/economy-formula-verification-20260503-014134.json`
- `allPassed=true`
- `People Max Formula PASS`
- Go 后端 `go build` 已恢复通过

但 tracker 还没有同步勾选经济项。先同步 tracker，再继续后续模块。

## 总原则

目标：把剩余 **11/13** 全部做完。

每个模块只有满足下面三项，才允许算完成：

1. 有对应 smoke/diff 脚本。
2. 最新 artifact 明确 PASS，不允许空测试、不允许 skipped 冒充通过。
3. `docs/specs/phase3-1to1-tracker.md` 对应项已勾选，并写明产物文件名。

禁止事项：

- 不许只改 tracker。
- 不许只改报告 expected。
- 不许吞掉失败项。
- 不许把接口 200 当业务闭环。
- 不许新增外围模块来逃避当前失败项。
- 不许打印 token/cookie。
- 不许大范围重构。

## 第 0 步：同步当前事实

### 要做

- 在 `phase3-1to1-tracker.md` 勾选“城池经济结算 1:1 对齐”。
- 在 tracker 下方记录证据：
  - `economy-formula-verification-20260503-014134.json`
  - `allPassed=true`
  - `People Max Formula PASS`

### 验收

```powershell
cd D:\APMServ5.2.6\new_project\backend
$env:GOCACHE='D:\APMServ5.2.6\new_project\backend\.gocache'
go build -mod=mod -o bin\api-check.exe .\cmd\api
```

必须通过。

## 第 1 步：完成出征与战斗 + 战报

当前状态：

- dispatch 已经从 `400` 修到 `200`
- 有有效 `troopId`
- DB 校验 PASS
- 士兵变化 PASS
- 但没有侦察战报，`report_structure` 仍失败

### 要做

- 修复侦察战报生成。
- `reports` 步骤不能再出现 `passed=true` 但 `scout_report=false`。
- 如果旧 PHP 侦察不会立即生成战报，要在脚本中调用正确的 settle/callback/finish 流程，而不是硬插报告。
- 战报字段必须至少校验：
  - report id
  - type/category
  - source city
  - target city
  - time
  - read/unread 状态
  - 侦察结果主体字段

### 验收

```powershell
cd D:\APMServ5.2.6\new_project
powershell -ExecutionPolicy Bypass -File .\scripts\smoke-battle-full-flow.ps1 -BaseUrl http://127.0.0.1:8080
```

完成后可勾选：

- 出征与战斗 1:1 对齐
- 战报字段与语义 1:1 对齐

前提是 artifact 中 dispatch、verify_db、settle/callback、report_structure 全 PASS。

## 第 2 步：建筑队列 1:1 对齐

### 范围

- 建筑升级/新建
- 队列创建
- 时间推进
- 资源扣除
- 完成后城池状态变化
- 与旧 PHP 字段语义一致

### 脚本

新增或完善：

```text
scripts/smoke-building-queue-1to1.ps1
artifacts/building-queue-*.json
```

### 验收

- 创建建筑任务 PASS
- 资源扣除 PASS
- 队列状态 PASS
- 强制结算/到期结算 PASS
- 城池详情刷新后建筑等级 PASS

## 第 3 步：科技队列 1:1 对齐

### 范围

- 科技升级
- 前置条件
- 资源扣除
- 队列状态
- 到期完成

### 脚本

```text
scripts/smoke-tech-queue-1to1.ps1
artifacts/tech-queue-*.json
```

### 验收

- 可升级科技 PASS
- 非法前置条件拒绝 PASS
- 资源扣除 PASS
- 到期完成后等级变化 PASS

## 第 4 步：征兵队列 1:1 对齐

### 范围

- 征兵创建
- 人口/资源扣除
- 兵种数量变化
- 队列完成
- 取消/异常条件

### 脚本

```text
scripts/smoke-recruit-queue-1to1.ps1
artifacts/recruit-queue-*.json
```

### 验收

- 创建征兵队列 PASS
- 资源与人口扣除 PASS
- 到期后兵力增加 PASS
- 非法兵种/资源不足拒绝 PASS

## 第 5 步：任务系统 1:1 对齐

### 范围

- 任务列表
- 完成条件
- 奖励领取
- 重复领取拒绝

### 脚本

```text
scripts/smoke-quest-system-1to1.ps1
artifacts/quest-system-*.json
```

### 验收

- 任务列表字段 PASS
- 任务完成检测 PASS
- 奖励发放 PASS
- 重复领取拒绝 PASS

## 第 6 步：联盟系统 1:1 对齐

### 范围

- 创建联盟
- 加入/退出
- 成员列表
- 职位/权限
- 公告或基础信息

### 脚本

```text
scripts/smoke-union-system-1to1.ps1
artifacts/union-system-*.json
```

### 验收

- 创建联盟 PASS
- 加入联盟 PASS
- 成员列表 PASS
- 权限校验 PASS
- 退出/解散基础流程 PASS

## 第 7 步：商城/充值兑换 1:1 对齐

### 范围

- 商品列表
- 购买
- 元宝/资源变化
- 兑换点券/充值相关只做旧版可验证路径

### 脚本

```text
scripts/smoke-shop-payment-1to1.ps1
artifacts/shop-payment-*.json
```

### 验收

- 商品列表字段 PASS
- 购买扣款 PASS
- 道具/资源到账 PASS
- 余额不足拒绝 PASS

## 第 8 步：邮件系统 1:1 对齐

### 范围

- 邮件列表
- 读邮件
- 发送邮件
- 删除邮件
- 附件/奖励如旧版支持则必须验证

### 脚本

```text
scripts/smoke-mail-system-1to1.ps1
artifacts/mail-system-*.json
```

### 验收

- 发送邮件 PASS
- 列表读取 PASS
- 已读状态 PASS
- 删除 PASS
- 附件领取 PASS 或明确旧版无附件路径

## 第 9 步：排行系统 1:1 对齐

### 范围

- 玩家排行
- 城池排行
- 联盟排行
- 分页/排序字段

### 脚本

```text
scripts/smoke-rank-system-1to1.ps1
artifacts/rank-system-*.json
```

### 验收

- 各榜单接口 PASS
- 排序方向 PASS
- 分页 PASS
- 关键字段 PASS

## 第 10 步：主界面/地图/功能窗关键交互 1:1 对齐

### 范围

- 主界面进入
- 城池资源刷新
- 地图基础数据
- 功能窗打开
- 关键按钮状态

### 脚本

可用 API smoke + 前端 smoke 组合：

```text
scripts/smoke-main-map-ui-1to1.ps1
artifacts/main-map-ui-*.json
```

### 验收

- 登录后主界面数据 PASS
- 地图数据 PASS
- 城池资源展示 PASS
- 关键功能窗数据 PASS
- 前端无关键 JS 错误 PASS

## 第 11 步：旧 PHP vs Go 对照恢复

这一步贯穿所有模块，但不能阻塞单模块 smoke。

当前问题：

- `oldServerReachable=false`
- `totalTests=0`

### 要做

- 修通旧 PHP 服务 `http://127.0.0.1:8088`，或写明本机旧服务启动方式。
- `compare-legacy-rules.ps1` 不能在旧端不可达时生成“看起来成功”的产物。
- 旧端不可达时，artifact 必须标记为 invalid，不能计入完成度。

### 验收

```powershell
cd D:\APMServ5.2.6\new_project
powershell -ExecutionPolicy Bypass -File .\scripts\compare-legacy-rules.ps1 -OldBaseUrl http://127.0.0.1:8088 -NewBaseUrl http://127.0.0.1:8080
```

验收标准：

- `oldServerReachable=true`
- `loginSuccess=true`
- `totalTests > 0`
- `mismatches=0`
- `errors=0`

## 推荐执行顺序

不要并行乱开。按下面顺序一直做完：

1. 同步 tracker，把经济标为已完成。
2. 修战斗战报，把进度推到 4/13。
3. 建筑队列。
4. 科技队列。
5. 征兵队列。
6. 任务系统。
7. 联盟系统。
8. 商城/充值兑换。
9. 邮件系统。
10. 排行系统。
11. 主界面/地图/功能窗关键交互。
12. 恢复旧 PHP vs Go 全量对照。

## 每完成一项必须回报

每项完成后只输出下面格式：

```text
完成项：
证据 artifact：
脚本命令：
关键 PASS 字段：
修改文件：
tracker 已勾选：
剩余项：
```

没有 artifact，不算完成。

