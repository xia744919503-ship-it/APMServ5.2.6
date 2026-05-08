# 旧版规则对照矩阵 (Legacy Rule Matrix)

更新时间：2026-05-02

## 概述

本矩阵记录 PHP 旧版游戏规则与 Go 新版实现的对照差异，作为 1:1 复刻验收基线。

---

## 1. 任务系统 (TaskFunc.php)

### getTaskTypeGroupList
| 字段 | 规则 |
|------|------|
| 触发条件 | 无参数，读取用户所有任务 |
| 成功返回 | 任务分组列表 (daily/main/union/reward) |
| DB影响 | sys_task_user 任务进度表 |

### CompleteTask (任务奖励)
| 字段 | 规则 |
|------|------|
| 触发条件 | 任务目标全部达成 |
| 成功返回 | 奖励物品/资源写入 sys_things/sys_city_resource |
| 失败状态码 | task_not_complete |
| 失败文案 | 任务未完成，无法领取奖励 |

### DropTask (放弃任务)
| 字段 | 规则 |
|------|------|
| 触发条件 | 委托任务可放弃 |
| 成功返回 | sys_task_user state=3 |
| 失败状态码 | invalid_operation |
| 失败文案 | 任务无法放弃 |

---

## 2. 联盟系统 (UnionFunc.php)

### CreateUnion (创建联盟)
| 字段 | 规则 |
|------|------|
| 触发条件 | 金币 >= 5000，无同名联盟 |
| 成功返回 | 创建联盟，uid 成为盟主 |
| 失败状态码 | no_enough_money / name_duplicate |
| 失败文案 | 金币不足 / 联盟名称已存在 |

### ApplyJoin (申请加入)
| 字段 | 规则 |
|------|------|
| 触发条件 | 未加入联盟，非盟主 |
| 成功返回 | sys_union_apply 插入记录 |
| 失败状态码 | already_in_union |
| 失败文案 | 已在联盟中 |

### AcceptApply (审批申请)
| 字段 | 规则 |
|------|------|
| 触发条件 | 盟主/副盟主权限 |
| 成功返回 | sys_user union_id 更新，sys_union_apply 删除 |
| 失败状态码 | no_permission |
| 失败文案 | 无权审批 |

---

## 3. 商城系统 (ShopFunc.php)

### buyGoods (商城购买)
| 字段 | 规则 |
|------|------|
| 触发条件 | id 在售，cnt >= 1，paytype in (0,1) |
| 成功返回 | 物品写入 sys_things，扣除金钱/礼金 |
| 失败状态码 | invalid_pay_type / invalid_amount / stop_sale / no_enough_YuanBao / no_enough_Gift |
| 失败文案 | 支付类型错误 / 数量错误 / 商品已下架 / 元宝不足 / 礼金不足 |
| DB影响 | sys_things, sys_user(money/gift), log_shop |

### buyBattleGoods (战场商品购买)
| 字段 | 规则 |
|------|------|
| 触发条件 | 荣誉足够，勋章足够，在售商品 |
| 成功返回 | 物品/装备/将领写入对应表 |
| 失败状态码 | no_enough_Credit / no_enough_Medal |
| 失败文案 | 荣誉不足 / 勋章不足 |
| DB影响 | sys_things/sys_user_armor/sys_city_hero, sys_user(honour) |

### exchangeLiquan (礼券兑换)
| 字段 | 规则 |
|------|------|
| 触发条件 | 10位字母数字验证码，未使用，未绑定他人 |
| 成功返回 | sys_ticket uid 更新，获得物品/金币 |
| 失败状态码 | code_notNull / invalid_code / used_code / code_bind |
| 失败文案 | 验证码不能为空 / 验证码无效 / 验证码已使用 / 验证码已绑定他人 |
| DB影响 | sys_ticket, sys_things, sys_user(money/gift) |

---

## 4. 战报系统 (ReportFunc.php)

### getUnreadReport (未读战报)
| 字段 | 规则 |
|------|------|
| 触发条件 | type, page 参数 |
| 成功返回 | [总页数, 当前页, 战报列表] |
| DB影响 | sys_alarm.report=0 (标记已读) |

### getReport (战报列表)
| 字段 | 规则 |
|------|------|
| 触发条件 | type >= 0 |
| 成功返回 | [总页数, 当前页, 战报列表] |
| DB字段 | sys_report (id,origincid,origincity,happencid,happencity,title,type,time,read,battleid) |

### delReport (删除战报)
| 字段 | 规则 |
|------|------|
| 触发条件 | ids 数组 |
| 成功返回 | 删除后重新获取列表 |
| DB影响 | sys_report state=1 (软删除) |

### getReportDetail (战报详情)
| 字段 | 规则 |
|------|------|
| 触发条件 | battleid > 0 |
| 成功返回 | sys_battle_report 按 round 排序 |
| 失败状态码 | 无战斗数据 |

---

## 5. 排行系统 (RankFunc.php)

### getRank (排行分页)
| 字段 | 规则 |
|------|------|
| 触发条件 | start 页码, type 排行类型 |
| 成功返回 | [更新时间, 总页数, 当前页, type, 排行列表] |
| 失败返回 | [更新时间, 0, 0, type, []] |
| type 类型 | user/union/hero_level/hero_affairs/hero_bravery/hero_wisdom/city_people/city_type/jungong/juanxian/military/rich/battle_total 等 |

### getNameRank (名称搜索)
| 字段 | 规则 |
|------|------|
| 触发条件 | name 模糊匹配 |
| 成功返回 | [更新时间, 1页, 0页, type, 匹配列表] |

### getMyRank (我的排名)
| 字段 | 规则 |
|------|------|
| 触发条件 | 用户名精确匹配 |
| 成功返回 | 定位到用户所在页 |
| 失败处理 | 未找到则 rank=1 |

---

## 6. 出征系统 (TroopFunc.php)

### getArmyTroops (我方队列)
| 字段 | 规则 |
|------|------|
| 触发条件 | uid 自己的部队 (state < 4) |
| 成功返回 | 部队列表含: fromcity, resource, soldier, wtype, targetcity |
| DB影响 | sys_alarm.troops=0 |

### getEnemyTroops (敌方队列)
| 字段 | 规则 |
|------|------|
| 触发条件 | 攻击自己城池/属地的部队 (task in 2,3,4) |
| 成功返回 | 部队列表含: targetcity, wtype, viewLevel, enemyuser, origincity |
| 权限控制 | 烽火台 >= 4级 显示敌人名, >= 5级 显示起点城 |

### callBackTroop (召回军队)
| 字段 | 规则 |
|------|------|
| 触发条件 | task=0 (前往中) 可召回，task=4/2 (驻军/侦查) 可召回 |
| 成功返回 | state=1, 计算返回时间 |
| 失败状态码 | gather / invalid_army / on_back / army_in_battle / army_on_way_back |
| 失败文案 | 采集中的军队无法召回 / 无效军队 / 禁止回城 / 战斗中 / 正在返回 |
| DB影响 | sys_troops, sys_city_hero(state), 双方资源更新 |

### setSoldierTactics (设置战术)
| 字段 | 规则 |
|------|------|
| 触发条件 | battleid, userIsAttack, stype, action, target |
| 成功返回 | 写入 mem_battle_tactics |
| 失败状态码 | cant_change_enemy_tactics |
| 失败文案 | 不能更改敌方战术 |

---

## 7. 城池经济 (CityFunc.php)

### levyResource (征收资源)
| 字段 | 规则 |
|------|------|
| 触发条件 | resid 0-4 (金/粮/木/石/铁), 冷却 3 秒, 民心 > 10 |
| 成功返回 | 资源 += people * rate * GAME_SPEED_RATE |
| 失败状态码 | time_limit / not_enough_morale |
| 失败文案 | 冷却时间未到 / 民心不足 |
| DB影响 | mem_city_resource 资源，民心-10，稳定人口重算 |

### changeTax (调整税率)
| 字段 | 规则 |
|------|------|
| 触发条件 | 0-100 整数 |
| 成功返回 | mem_city_resource tax 更新 |
| 副作用 | morale_stable = max(0, min(100, 100 - tax - complaint)) |

### getCityProduct (获取产量)
| 字段 | 规则 |
|------|------|
| 触发条件 | 无参数 |
| 成功返回 | [资源信息, 劳力人口, 基础产量, 科技加成, 军队消耗, 城守加成, 增益到期时间] |
| 关键计算 | food_add = GLOBAL_FOOD_RATE * food_all_people * GAME_SPEED_RATE * 10 |

---

## 8. 关键常量对照

| 常量名 | 值 | 用途 |
|--------|-----|------|
| REPORT_PAGE_CPP | 10 | 战报每页条数 |
| RANK_PAGE_CPP | 10 | 排行每页条数 |
| GAME_SPEED_RATE | 1 | 游戏速度倍率 |
| GLOBAL_FOOD_RATE | 1 | 粮食基础产量 |
| GLOBAL_WOOD_RATE | 1 | 木材基础产量 |
| GLOBAL_ROCK_RATE | 1 | 石料基础产量 |
| GLOBAL_IRON_RATE | 1 | 铁锭基础产量 |
| GLOBAL_GOLD_RATE | 0.2 | 黄金基础产量 |

---

## 9. 状态码对照表

| 状态值 | 含义 |
|--------|------|
| troop state=0 | 前往中 |
| troop state=1 | 返回中 |
| troop state=2 | 侦查中 |
| troop state=3 | 战斗中 |
| troop state=4 | 驻军中 |
| troop state=5 | 采集中 (特殊) |
| task=0 | 城内移动 |
| task=1 | 增援 |
| task=2 | 掠夺 |
| task=3 | 攻击 |
| task=4 | 占领 |
| task=7/8/9 | 战场相关 |

---

## 10. 差异跟踪

### 已知不一致项
1. 排行榜 kind=union 错误回退到 kind=user (已修复)
2. sys_troops.startcid 未写入 (已修复)
3. 联盟权限提示语义 (已修复)

---

更新时间：2026-05-02
维护者：Claude