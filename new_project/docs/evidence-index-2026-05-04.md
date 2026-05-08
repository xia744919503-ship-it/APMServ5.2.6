# BloodWar HTML5 1:1 Parity 证据索引
日期: 2026-05-04
状态: in_progress

## 概述

本文档汇总了 BloodWar Flash 到 HTML5 重构的 1:1 证据收集结果。

## 状态词汇表

| 状态 | 说明 |
|------|------|
| `not_started` | 尚未开始 |
| `evidence_drafted` | 已起草证据文档 |
| `implemented_partial` | 部分实现 |
| `functional_verified` | 功能验证通过 |
| `coordinate_verified` | 坐标验证通过 |
| `screenshot_verified` | 截图验证通过 |
| `1to1_accepted` | 1:1 完全对齐 |

## P0 引导链状态 (gid6-gid14)

| 阶段 | 证据文档 | 截图 | 回归测试 | 状态 |
|------|----------|------|----------|------|
| 登录 | `login-1to1-evidence-2026-05-04.md` | chrome-login-final.png | verify-speed-flow.ps1 | `screenshot_verified` |
| 创建角色 | `createrole-1to1-evidence-2026-05-04.md` | chrome-create-role-final.png | verify-speed-flow.ps1 | `screenshot_verified` |
| 城市场景 | `city-1to1-evidence-2026-05-04.md` | chrome-city-final.png | verify-speed-flow.ps1 | `screenshot_verified` |
| 建筑面板 | `building-panel-1to1-evidence-2026-05-04.md` | chrome-city-final.png(含) | verify-speed-flow.ps1 | `screenshot_verified` |
| 加速道具 | `usegoods-dialog-1to1-evidence-2026-05-04.md` | chrome-city-final.png(含) | verify-speed-flow.ps1 | `screenshot_verified` |
| 任务对话框 | `task-dialog-1to1-evidence-2026-05-04.md` | chrome-city-task-panel-round4.png | verify-speed-flow.ps1 | `screenshot_verified` |

**重要**: P0 引导链状态为 `screenshot_verified`。两个回归脚本均通过 (task1_state=1)。

## P0 最终验收记录

```
Global Gate:
- npm run build: ✓ PASS (1.25s)
- go test ./...: ✓ PASS (no test files)
- verify-speed-flow.ps1: ✓ PASS (task1_state=1)
- verify-speed-flow-coordinates.ps1: ✓ PASS (task1_state=1)
- P0 accepted: yes
```

## 关键截图

| 截图 | 路径 | 大小 |
|------|------|------|
| 登录 | `artifacts/screenshots/chrome-login-final.png` | 1.1MB |
| 创建角色 | `artifacts/screenshots/chrome-create-role-final.png` | 1.1MB |
| 城市+任务面板 | `artifacts/screenshots/chrome-city-task-panel-round4.png` | 407KB |
| 城市 | `artifacts/screenshots/chrome-city-final.png` | 407KB |

## Worker 执行状态

| Worker | 状态 | 输出文件 |
|--------|------|----------|
| Worker 1: P0回归脚本修复 | `coordinate_verified` | verify-speed-flow*.ps1 |
| Worker 2: g13纯坐标修复 | `coordinate_verified` | guide-coordinate-model.md |
| Worker 3-7: 证据文档更新 | `screenshot_verified` | 证据文档已更新 |
| Worker 8: World/Battle/Utility backlog | `pending` | 待创建 |
| Worker 9: 证据索引 + 最终报告 | `in_progress` | 本文件 |

## 已验证通过的特性

- [PASS] 登录表单提交 (legacyLogin API)
- [PASS] 创建角色流程 (createRole API)
- [PASS] 城内建筑点击打开
- [PASS] 民房建造流程
- [PASS] 建造加速流程
- [PASS] 任务面板打开
- [PASS] 任务选择 (纯坐标点击 gid12)
- [PASS] 任务奖励领取 (纯坐标点击 gid13)
- [PASS] DB验证: task1_state=1

## 已知简化/差异

1. **背景处理**: Flash纯色背景 vs HTML5图片背景
2. **布局方式**: Flash绝对定位 vs HTML5 flex/grid
3. **交互方式**: Flash mouseMove事件 vs HTML5像素检测
4. **任务分类**: Flash固定按钮 vs HTML5动态分类
5. **物品布局**: Flash HBox vs HTML5 flex-wrap

## 未实现功能

1. 城外视图切换
2. 地图视图
3. 联盟功能
4. 报告查看
5. 建筑升级动画
6. 世界/战斗/工具窗口 (Worker 8 backlog)

## 下一步

1. [ ] Worker 8: 创建 world/battle/utility backlog 文档
2. [ ] Worker 9: 更新最终报告 + evidence index
3. [ ] Phase 2: 城外/地图视图实现 (optional)
4. [ ] Phase 3: 世界/战斗/工具窗口实现 (long-term)