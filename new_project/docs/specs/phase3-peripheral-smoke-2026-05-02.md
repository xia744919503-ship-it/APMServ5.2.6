# Phase 3 外围系统冒烟记录（2026-05-02）

## 脚本

- `scripts/smoke-legacy-peripheral-modules.ps1`

## 覆盖范围

1. 任务系统
   - `GET /api/me/tasks`
   - `POST /api/me/tasks/claim`（失败语义）
2. 联盟系统
   - `GET /api/me/union`
   - `POST /api/me/union/create`（无权限失败语义）
3. 商城与充值
   - `GET /api/me/shop`
   - `POST /api/me/shop/buy`（非法参数失败语义）
   - `GET /api/me/charge`
   - `POST /api/me/charge/exchange`（非法参数失败语义）
4. 邮件与战报
   - `GET /api/mail?folder=system|inbox|outbox`
   - `GET /api/reports?filter=unread|type0`
5. 排行系统
   - `GET /api/rankings` 25 类 kind

## 实测结果

- `ok=true`
- 关键失败语义（test 账号）：
  - 任务领奖未完成：`400` + `任务条件尚未完成。`
  - 联盟创建权限不足：`400` + `你的鸿胪寺等级不足2级，不能创建联盟。`
  - 商城非法购买数量：`400` + `商品或数量无效。`
  - 非法充值兑换数量：`400` + `充值数量无效。`
- 排行 kind 一致性：
  - `user / union / hero_* / city_* / jungong* / juanxian* / qinwang* / gongpin* / military* / rich* / battle*`
  - 全部返回 `200` 且 `response.kind == request.kind`

## 结论

- 外围系统已具备“可回归的读写失败语义基线”。
- 该结果仅说明接口行为可稳定回放，不等于 1:1 最终验收完成。
