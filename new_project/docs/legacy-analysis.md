# 旧项目理解纪要

扫描时间：2026-04-15

## 旧项目结构

- 前台入口在 `www/htdocs/index.php`
- 旧前端核心是 `BloodWar.swf`
- 业务逻辑主要集中在 `www/htdocs/server/game/*.php`
- 数据库配置在 `www/htdocs/server/config/db.php`
- 实际数据位于 `MySQL5.1/data/bloodwar`

## 已确认的旧栈

- PHP 5.x
- Flash / SWF
- MySQL 5.1
- APMServ Windows 集成环境
- 大量 `mysql_*` 与 `require_once` 风格的过程式代码

## 已确认的关键表

- `sys_user`
- `sys_city`
- `mem_city_resource`
- `mem_world`
- `sys_building`
- `sys_city_soldier`
- `sys_city_hero`
- `sys_troops`
- `sys_battle`
- `sys_announce`

## 旧库快照

基于本地 `bloodwar` 实测：

- 玩家数：896
- 城池数：6930
- 世界格点：250000

## 本次重构策略

1. 不直接改旧 PHP 逻辑，先通过 Go 做兼容读取层。
2. 先把最容易验证的新界面切出来：
   - 项目概览
   - 世界地图
   - 城池详情
3. 等新前端稳定后，再逐步迁移登录、科技、部队、战斗、战报和 GM 功能。
