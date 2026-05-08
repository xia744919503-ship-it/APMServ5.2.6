# World / Battle / Utility Implementation Backlog
日期: 2026-05-04
状态: evidence_drafted

## 概述

本文档记录 BloodWar HTML5 重构中尚未实现的功能区域，基于 FFDec 源码分析。

## 未实现功能清单

### 1. 世界地图 (World)

**FFDec 源码**: `artifacts/ffdec/BloodWar/scripts/World/`
- WorldMapDialog.as
- WorldMapCanvas.as
- WorldMapEvent.as

**Assets**:
- 地图瓦片图片
- 州郡边界线
- 城市标记图标

**Backend 需求**:
- GET /api/city/list 或类似获取世界城市列表
- 坐标转换 (x,y) -> 瓦片索引

**实现任务:
1. [ ] 创建 WorldMapDialog 组件
2. [ ] 实现瓦片网格渲染
3. [ ] 添加城市标记和点击检测
4. [ ] 连接世界城市API
5. [ ] 添加坐标转换逻辑

**Acceptance Test**:
- 用户可以在世界地图上查看所有城市
- 点击城市可以进入 (需要先实现跳转逻辑)

---

### 2. 战斗系统 (Battle)

**FFDec 源码**: `artifacts/ffdec/BloodWar/scripts/Battle/`
- BattleDialog.as
- BattleAction.as
- BattleFormation.as

**Assets**:
- 战斗背景图
- 兵种图标
- 阵法图片

**Backend 需求**:
- POST /api/battle/start
- GET /api/battle/report
- 战斗动画数据

**实现任务:
1. [ ] 创建 BattleDialog 组件
2. [ ] 实现战斗场景渲染
3. [ ] 兵种选择和阵法配置
4. [ ] 战斗动画序列
5. [ ] 战斗结果展示

**Acceptance Test**:
- 用户可以发起战斗
- 战斗动画正常播放
- 战斗结果正确显示

---

### 3. 联盟系统 (Union)

**FFDec 源码**: `artifacts/ffdec/BloodWar/scripts/Union/`
- UnionDialog.as
- UnionList.as
- UnionMember.as

**Assets**:
- 联盟图标
- 联盟旗帜

**Backend 需求**:
- GET /api/union/list
- POST /api/union/create
- POST /api/union/join
- 联盟成员管理API

**实现任务:
1. [ ] 创建 UnionDialog 组件
2. [ ] 实现联盟列表
3. [ ] 创建联盟表单
4. [ ] 成员管理界面

**Acceptance Test**:
- 用户可以查看联盟列表
- 用户可以创建/加入联盟

---

### 4. 邮件系统 (Mail)

**FFDec 源码**: `artifacts/ffdec/BloodWar/scripts/Mail/`
- MailDialog.as
- MailItem.as

**Backend 需求**:
- GET /api/mail/list
- POST /api/mail/send
- POST /api/mail/delete

**实现任务:
1. [ ] 创建 MailDialog 组件
2. [ ] 邮件列表显示
3. [ ] 邮件阅读界面
4. [ ] 邮件发送功能

**Acceptance Test**:
- 用户可以查看邮件列表
- 用户可以阅读邮件
- 用户可以发送邮件

---

### 5. 报告系统 (Report)

**FFDec 源码**: `artifacts/ffdec/BloodWar/scripts/Report/`
- ReportDialog.as
- BattleReport.as

**Backend 需求**:
- GET /api/report/list
- GET /api/report/detail

**实现任务:
1. [ ] 创建 ReportDialog 组件
2. [ ] 报告列表
3. [ ] 报告详情查看

**Acceptance Test**:
- 用户可以查看战斗报告
- 报告详情正确显示

---

### 6. 排行榜 (Rank)

**FFDec 源码**: `artifacts/ffdec/BloodWar/scripts/Rank/`
- RankDialog.as

**Backend 需求**:
- GET /api/rank/list

**实现任务:
1. [ ] 创建 RankDialog 组件
2. [ ] 排行榜数据获取
3. [ ] 排名展示

**Acceptance Test**:
- 用户可以查看排行榜

---

### 7. 商店 (Shop)

**FFDec 源码**: `artifacts/ffdec/BloodWar/scripts/Shop/`
- ShopDialog.as
- ShopItem.as

**Backend 需求**:
- GET /api/shop/list
- POST /api/shop/buy

**实现任务:
1. [ ] 创建 ShopDialog 组件
2. [ ] 商品列表
3. [ ] 购买功能

**Acceptance Test**:
- 用户可以查看商品
- 用户可以购买商品

---

### 8. 英雄/招募 (Hero)

**FFDec 源码**: `artifacts/ffdec/BloodWar/scripts/Hero/`
- HeroDialog.as
- RecruitDialog.as

**Assets**:
- 英雄头像
- 招募动画

**Backend 需求**:
- GET /api/hero/list
- POST /api/hero/recruit
- 武将卡池数据

**实现任务:
1. [ ] 创建 HeroDialog 组件
2. [ ] 武将列表显示
3. [ ] 招募功能
4. [ ] 武将详情

**Acceptance Test**:
- 用户可以查看武将
- 用户可以招募武将

---

### 9. 科技/学院 (College)

**FFDec 源码**: `artifacts/ffdec/BloodWar/scripts/College/`
- CollegeDialog.as
- ResearchDialog.as

**Backend 需求**:
- GET /api/tech/list
- POST /api/tech/research

**实现任务:
1. [ ] 创建 CollegeDialog 组件
2. [ ] 科技列表
3. [ ] 研究功能

**Acceptance Test**:
- 用户可以查看科技
- 用户可以研究科技

---

### 10. 兵营/训练 (Troop/Barracks)

**FFDec 源码**: `artifacts/ffdec/BloodWar/scripts/Troop/`
- TroopDialog.as
- BarracksDialog.as

**Backend 需求**:
- GET /api/army/list
- POST /api/army/train
- 兵种数据

**实现任务:
1. [ ] 创建 BarracksDialog 组件
2. [ ] 兵种列表
3. [ ] 训练功能

**Acceptance Test**:
- 用户可以查看兵种
- 用户可以训练士兵

---

## FFDec 源码清单

```
artifacts/ffdec/BloodWar/scripts/
├── Battle/           # 战斗系统
├── College/          # 科技学院
├── Hero/             # 英雄招募
├── Hotel/            # 客栈(已有部分)
├── Mail/             # 邮件系统
├── Market/           # 市场
├── Rank/             # 排行榜
├── Report/           # 报告系统
├── Shop/             # 商店
├── Troop/            # 兵营训练
├── Union/            # 联盟系统
└── World/            # 世界地图
```

## 优先级建议

| 优先级 | 功能 | 理由 |
|--------|------|------|
| P1 | 邮件系统 | 基础通信功能 |
| P1 | 报告系统 | 查看战斗结果 |
| P2 | 商店 | 购买道具 |
| P2 | 世界地图 | 探索游戏世界 |
| P3 | 联盟系统 | 社交功能 |
| P3 | 英雄招募 | 核心玩法 |
| P4 | 科技/兵营 | 后期功能 |

## 下一步

1. 选择下一个功能区域 (建议从 P1 开始)
2. 深入分析 FFDec 源码
3. 准备 evidence 文档
4. 实现功能