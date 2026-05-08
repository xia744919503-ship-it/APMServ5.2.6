# Flash 1:1 复刻审查与 Claude 完整清单

审查时间：2026-05-04

## 结论

Claude 还没有做完 Flash 1:1 复刻。

当前只能承认：

- 已经提取了旧 Flash 资源和 AS 源码。
- 已经有登录、建角、城内壳、新手引导、建民房、加速、任务领取的局部实现和证据文档。
- 后端 `go test ./...` 当前通过。

但当前不能验收：

- 前端 `npm run build` 当前失败。
- 两个前端回归脚本当前失败。
- `frontend-1to1-final-gate-report-2026-05-04.md` 里的 “P0 accepted: yes” 已经过期，当前不成立。
- World / Battle / Utility 大量 Flash 模块没有实现。
- 证据截图存在命名不可靠问题，例如 `chrome-login-final.png` 当前内容已经是城内画面，不是登录页。

## 本次实跑结果

### 1. 前端构建

命令：

```powershell
cd D:\APMServ5.2.6\new_project\frontend
npm run build
```

结果：失败。

关键错误：

```text
[vite:vue] src/App.vue (936:9): Element is missing end tag.
```

阻断文件：

- `frontend/src/App.vue`

结论：

- 当前前端源码不可构建。
- 当前任何“前端最终验收通过”的报告都不能采信。

### 2. 后端测试

命令：

```powershell
cd D:\APMServ5.2.6\new_project\backend
$env:GOCACHE='D:\APMServ5.2.6\new_project\backend\.gocache'
go test ./...
```

结果：通过。

但注意：

- 多数包是 `[no test files]`。
- 这只能证明当前 Go 包能编译和测试入口通过，不代表玩法 1:1。

### 3. 引导功能回归

命令：

```powershell
cd D:\APMServ5.2.6\new_project
powershell -ExecutionPolicy Bypass -File .\artifacts\verify-speed-flow.ps1
powershell -ExecutionPolicy Bypass -File .\artifacts\verify-speed-flow-coordinates.ps1
```

结果：均失败。

主要现象：

- guide 文案读取为 `null` / `none`。
- `taskPanelVisible=hidden`。
- `task1_state` 为空。
- 坐标脚本命中 `VITE-ERROR-OVERLAY`。

结论：

- 当前 P0 新手引导链路不可验收。
- 需要先修构建，再重跑。

## 旧 Flash 资料位置

原始 SWF：

- `D:\APMServ5.2.6\www\htdocs\BloodWar.swf`
- `D:\APMServ5.2.6\www\htdocs\common2.swf`
- `D:\APMServ5.2.6\www\htdocs\framework_3.0.0.477.swf`
- `D:\APMServ5.2.6\www\htdocs\framework_3.1.0.2710.swf`
- `D:\APMServ5.2.6\www\htdocs\framework_3.2.0.3958.swf`

已提取资料：

- `artifacts/ffdec/BloodWar/scripts`
- `artifacts/ffdec/BloodWar/images`
- `artifacts/ffdec/BloodWar.xml`
- `artifacts/ffdec/BloodWar.dumpAS3.txt`
- `artifacts/ffdec/common2/scripts`
- `artifacts/ffdec/common2/images`
- `artifacts/ffdec/common2.xml`

旧 Flash 舞台：

- `1000 x 600`
- 复刻时必须以这个 stage 坐标为基准。

## 必须使用的工具

### 1. FFDec / JPEXS

用途：

- 反编译 SWF。
- 导出 AS 源码、图片、shape、movie、text、binaryData。
- 查看组件树、资源绑定、类名、MXML/Flex 布局。

路径：

- `C:\Program Files (x86)\FFDec\ffdec.exe`
- `C:\Program Files (x86)\FFDec\ffdec.jar`

推荐命令：

```powershell
& 'D:\APMServ5.2.6\new_project\tools\jre\jdk-17.0.18+8-jre\bin\java.exe' `
  -jar 'C:\Program Files (x86)\FFDec\ffdec.jar' `
  -export script,image,shape,movie,sound,binaryData,text `
  D:\APMServ5.2.6\new_project\artifacts\ffdec\BloodWar `
  D:\APMServ5.2.6\www\htdocs\BloodWar.swf

& 'D:\APMServ5.2.6\new_project\tools\jre\jdk-17.0.18+8-jre\bin\java.exe' `
  -jar 'C:\Program Files (x86)\FFDec\ffdec.jar' `
  -export script,image,shape,movie,sound,binaryData,text `
  D:\APMServ5.2.6\new_project\artifacts\ffdec\common2 `
  D:\APMServ5.2.6\www\htdocs\common2.swf
```

### 2. RABCDAsm

用途：

- 当 FFDec 不能还原 Flex `UIComponentDescriptor` 细节时，用 ABC 字节码继续追。
- 追事件处理、服务命令字符串、组件初始化闭包。

路径：

- `D:\APMServ5.2.6\new_project\tools\rabcdasm`

### 3. 原始 Flash 运行/截图工具

用途：

- 跑旧 SWF。
- 截旧版每个窗口状态。
- 给 HTML5 做视觉对照。

优先级：

1. 能跑旧 Flash projector 就用 projector。
2. projector 不稳定时，用 FFDec 预览和导出资源补足。
3. Ruffle 只能作为辅助观察，不能替代 FFDec 源码依据。

### 4. 浏览器自动化

用途：

- 跑 HTML5。
- 坐标点击。
- 截图。
- 对比元素命中。

必须继续使用：

- Chrome DevTools Protocol / Playwright / 现有 `verify-speed-flow*.ps1`。

要求：

- 不能用隐藏 helper `.click()` 假装坐标通过。
- 必须记录 `elementFromPoint`、viewport 坐标、stage 坐标、截图路径、DB 状态。

### 5. 图片/像素对比

用途：

- 旧 Flash 截图 vs HTML5 截图。

建议工具：

- ImageMagick `compare`
- pixelmatch
- Playwright screenshot diff

验收要求：

- 每个窗口至少有旧图、新图、diff 图。
- 不能只写“看起来像”。

### 6. DB/API 对照工具

用途：

- 验证旧 PHP / Go / HTML5 状态一致。

必须用：

- MySQL 查询旧库快照。
- Go API smoke。
- PHP 源码对照。
- `compare-legacy-rules.ps1` 或新的 diff 脚本。

## 先修阻断项

### P0-0：修前端构建

目标：

- `npm run build` 必须通过。

当前错误：

- `frontend/src/App.vue:936:9`
- `Element is missing end tag`

验收：

```powershell
cd D:\APMServ5.2.6\new_project\frontend
npm run build
```

必须 PASS。

### P0-1：重跑现有回归

修完构建后立即跑：

```powershell
cd D:\APMServ5.2.6\new_project
powershell -ExecutionPolicy Bypass -File .\artifacts\verify-speed-flow.ps1
powershell -ExecutionPolicy Bypass -File .\artifacts\verify-speed-flow-coordinates.ps1
```

必须满足：

- exit code 0。
- 没有 `VITE-ERROR-OVERLAY`。
- 没有 `null` / `none` guide 文案。
- `taskPanelVisible=visible`。
- DB 中 task1 已领取。

### P0-2：清理错误证据

必须修正：

- 过期的 `frontend-1to1-final-gate-report-2026-05-04.md`。
- 截图命名和内容不一致的问题。
- 文档里乱码严重的中文内容。

要求：

- 每个证据文档必须写清：旧 Flash 来源、HTML5 截图、回归命令、当前状态。
- 不能把失败状态写成 accepted。

## 完整 Flash 1:1 模块清单

下面所有模块必须逐项对照旧 Flash，不允许只实现“差不多的现代弹窗”。

每个模块的验收模板：

1. FFDec 源码路径。
2. 旧 Flash 截图。
3. HTML5 截图。
4. 坐标/尺寸表。
5. 图片资源映射表。
6. 事件处理映射。
7. API/DB 状态映射。
8. 正常流程。
9. 失败流程。
10. 自动化回归脚本。

### A. 登录与建角

- [ ] LoginDialog：背景、账号、密码、记住账号/密码、公告、登录、注册、排队。
- [ ] CreateRoleDialog：头像、性别、君主名、州选择、地图 hover、地图点击、协议、开始游戏。
- [ ] 队列状态：旧 Flash 登录排队 UI。
- [ ] 错误提示：账号错误、封禁、重名、非法名、服务器忙。

旧源码：

- `artifacts/ffdec/BloodWar/scripts/Login`

### B. 主城壳

- [ ] TopPanel：城内、城池、地图、任务、军务、战斗、武将、联盟、邮件、商店等按钮状态。
- [ ] BottomPanel：聊天条、系统消息、功能按钮。
- [ ] UserInfo.InfoPanel：君主、城名、坐标、民心、税率、资源、人口。
- [ ] 城内背景和 1000x600 stage 对齐。
- [ ] 缩放/居中规则。

旧源码：

- `Bar`
- `UserInfo`
- `Building`

### C. 城内建筑

- [ ] BuildingGrid：空地、选中、高亮、灰态、等级图标。
- [ ] 城墙、市政厅特殊位置。
- [ ] 所有内城建筑图：
  - 官府
  - 民房
  - 仓库
  - 市场
  - 兵营
  - 铁匠铺
  - 马厩
  - 客栈
  - 招贤馆
  - 书院
  - 寺庙
  - 校场
  - 烽火台
  - 城墙
- [ ] 建造面板 CreateBuildingDialog。
- [ ] 建筑详情 BuildingListDialog / BuildingTip。
- [ ] 升级、拆除、取消、加速、条件不足。
- [ ] 建筑倒计时和完成刷新。

旧源码：

- `Building`
- `Government`
- `Barn`
- `Market`
- `BlackSmith`
- `Hotel`
- `Wall`

### D. 城外视图

- [ ] 城外背景。
- [ ] 农田、伐木场、采石场、铁矿。
- [ ] 城外空地/已建/升级态。
- [ ] 城内/城外 tab 切换。
- [ ] 城外建筑建造、升级、拆除、加速。

旧源码：

- `Building`
- `Ground`
- `Store`

### E. 世界地图

- [ ] WorldMapDialog。
- [ ] WorldMapCanvas。
- [ ] 地图瓦片。
- [ ] 城池、野地、州郡、NPC、玩家标记。
- [ ] 拖动/移动/缩放/方向按钮。
- [ ] 坐标输入和跳转。
- [ ] 点击目标后的操作菜单。

旧源码：

- `World`
- `Ground`

### F. 新手引导

- [ ] GuideTip 背景、箭头、遮罩。
- [ ] gid6-gid14 坐标完全对齐。
- [ ] 不能用隐藏 helper 点击替代真实坐标。
- [ ] 失败/重入/刷新恢复。
- [ ] DB guide/task 状态一致。

旧源码：

- `guide`
- `Task`

### G. 任务系统

- [ ] TaskDialog 外观。
- [ ] 任务分类。
- [ ] 任务列表。
- [ ] 任务详情。
- [ ] 完成状态。
- [ ] 奖励领取。
- [ ] 奖励发放 DB 对齐。
- [ ] 新手任务和普通任务都覆盖。

旧源码：

- `Task`

### H. 物品与加速

- [ ] Goods 使用窗口。
- [ ] 鲁班类建筑加速。
- [ ] 资源包。
- [ ] 增益道具。
- [ ] 物品数量变化。
- [ ] 背包/道具格子布局。

旧源码：

- `Goods`
- `Buffer`

### I. 邮件

- [ ] MailDialog。
- [ ] 收件箱、发件箱、系统邮件。
- [ ] 邮件列表。
- [ ] 邮件阅读。
- [ ] 发送邮件。
- [ ] 删除邮件。
- [ ] 未读状态。

旧源码：

- `Mail`

### J. 战报/报告

- [ ] ReportDialog。
- [ ] 战报列表。
- [ ] 战报详情。
- [ ] 侦察报告。
- [ ] 掠夺/战斗报告。
- [ ] 已读/未读。
- [ ] 删除。

旧源码：

- `Report`
- `Battle`
- `Troop`

### K. 军务/军队/兵营

- [ ] TroopDialog。
- [ ] ArmyDialog。
- [ ] 出征列表。
- [ ] 驻军列表。
- [ ] 召回。
- [ ] 运输、派遣、侦察、掠夺。
- [ ] 兵营训练。
- [ ] 征兵队列。
- [ ] 士兵详情和图标。

旧源码：

- `Troop`
- `Army`

### L. 战斗系统

- [ ] BattleDialog。
- [ ] BattleAction。
- [ ] BattleFormation。
- [ ] 阵法。
- [ ] 回合动画。
- [ ] 兵种位置。
- [ ] 战斗结果。
- [ ] 战斗报告生成。

旧源码：

- `Battle`

### M. 武将/客栈/招募

- [ ] HeroDialog。
- [ ] Hero 详情。
- [ ] 招募列表。
- [ ] 客栈。
- [ ] 武将装备。
- [ ] 加点。
- [ ] 状态、体力、忠诚、统率。

旧源码：

- `Hero`
- `Hotel`
- `Armor`

### N. 书院/科技

- [ ] CollegeDialog。
- [ ] 科技列表。
- [ ] 研究条件。
- [ ] 研究队列。
- [ ] 加速/取消。
- [ ] 科技等级和效果。

旧源码：

- `College`

### O. 联盟/鸿胪寺

- [ ] UnionDialog。
- [ ] 联盟列表。
- [ ] 创建联盟。
- [ ] 申请加入。
- [ ] 成员列表。
- [ ] 权限、职位、公告。
- [ ] 城池/名城相关信息。

旧源码：

- `Union`
- `Honglu`

### P. 商城/充值/活动

- [ ] ShopDialog。
- [ ] 商品分类。
- [ ] 商品详情。
- [ ] 购买。
- [ ] 货币扣除。
- [ ] 限购/失败状态。
- [ ] 充值兑换入口。
- [ ] 抽奖 Lottery。

旧源码：

- `Shop`
- `Lottery`
- `Goods`

### Q. 排行

- [ ] RankDialog。
- [ ] 排行种类。
- [ ] 分页。
- [ ] 排名字段。
- [ ] 联盟/个人/军功/财富等旧版 kind。

旧源码：

- `Rank`

### R. 市场/仓库/资源

- [ ] MarketDialog。
- [ ] 买卖资源。
- [ ] 运输资源。
- [ ] Barn 仓库容量。
- [ ] Store 野地/产量。
- [ ] 资源变动与旧 PHP 一致。

旧源码：

- `Market`
- `Barn`
- `Store`

### S. 官府/官职/安抚

- [ ] GovernmentDialog。
- [ ] 税率。
- [ ] 安抚。
- [ ] 征收。
- [ ] 城市改名。
- [ ] 官职/爵位。

旧源码：

- `Government`
- `Office`

### T. 好友/聊天/统计/计谋

- [ ] Friend。
- [ ] Chat。
- [ ] Stat。
- [ ] Trick。
- [ ] Tooltip。
- [ ] 通用弹窗、确认框、错误框。

旧源码：

- `friend`
- `Chat`
- `Stat`
- `Trick`
- `Tooltip`
- `Basic`

## 每个模块的完成标准

不能只写“已实现”。必须同时满足：

```text
旧 Flash 源码定位：有
旧 Flash 截图：有
HTML5 截图：有
视觉 diff：有
坐标差异表：有
资源映射表：有
事件/API 映射表：有
正常流程脚本：PASS
失败流程脚本：PASS
DB 状态证明：有
文档状态：1to1_accepted
```

## 不允许的验收方式

- 不允许仅凭截图说完成。
- 不允许用现代 UI 替代 Flash UI。
- 不允许只实现一个空弹窗。
- 不允许 helper click 伪造坐标点击。
- 不允许把失败脚本写成 warning 后继续通过。
- 不允许截图文件名和实际画面不一致。
- 不允许文档乱码。
- 不允许 final gate 报告和当前实跑结果不一致。

## 给 Claude 的执行顺序

### 第 1 步：修当前破损

1. 修 `frontend/src/App.vue:936` 缺标签。
2. 跑 `npm run build`。
3. 跑两个 speed flow 脚本。
4. 修证据文档和截图命名。

### 第 2 步：恢复 P0 引导闭环

1. 登录。
2. 建角。
3. 进城。
4. 建民房。
5. 加速。
6. 打开任务。
7. 领取奖励。
8. DB 证明 task1 已领取。

### 第 3 步：补齐 Flash 主 UI

1. TopPanel。
2. BottomPanel。
3. 左侧资源栏。
4. 城内/城外/地图切换。
5. 所有功能按钮可打开对应旧版窗口。

### 第 4 步：按模块逐个 1:1

优先顺序：

1. 任务。
2. 建筑。
3. 物品/加速。
4. 邮件。
5. 战报。
6. 军队/兵营。
7. 世界地图。
8. 武将/客栈。
9. 书院/科技。
10. 联盟。
11. 商城。
12. 排行。
13. 市场/仓库/资源。
14. 官府/官职/安抚。
15. 好友/聊天/统计/计谋。

### 第 5 步：最终总验收

必须一次性通过：

```powershell
cd D:\APMServ5.2.6\new_project\frontend
npm run build

cd D:\APMServ5.2.6\new_project\backend
$env:GOCACHE='D:\APMServ5.2.6\new_project\backend\.gocache'
go test ./...

cd D:\APMServ5.2.6\new_project
powershell -ExecutionPolicy Bypass -File .\artifacts\verify-speed-flow.ps1
powershell -ExecutionPolicy Bypass -File .\artifacts\verify-speed-flow-coordinates.ps1
```

并新增：

```powershell
powershell -ExecutionPolicy Bypass -File .\artifacts\verify-flash-ui-all-panels.ps1
powershell -ExecutionPolicy Bypass -File .\artifacts\verify-flash-visual-diff.ps1
powershell -ExecutionPolicy Bypass -File .\scripts\compare-legacy-rules.ps1
```

这些脚本还没齐，需要 Claude 建起来。

## 当前缺口估算

以完整 Flash 客户端为标准：

- 当前可采信完成度：约 15%-25%。
- 如果按当前源码可构建可运行标准：当前是 0%，因为前端构建失败。
- 剩余主要工作：75%-85%。

重点：

- P0 新手链路不是完整 Flash。
- 当前 World / Battle / Mail / Report / Union / Hero / College / Barracks / Rank / Shop 等都没有达到 1:1。
- 必须逐模块用 FFDec 和旧 Flash 截图做闭环。
