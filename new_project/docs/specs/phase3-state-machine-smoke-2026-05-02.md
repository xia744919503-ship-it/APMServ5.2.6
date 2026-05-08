# Phase 3 状态机冒烟记录（2026-05-02）

## 脚本

- `scripts/smoke-legacy-core-state-machines.ps1`

## 覆盖范围

1. 城池经济
   - `/api/cities/:cid/tax` PATCH 回写
   - `/api/cities/:cid/production` PATCH 回写
2. 建筑
   - `/api/cities/:cid/buildings/create` -> `/cancel`
   - `/api/cities/:cid/buildings/upgrade` -> `/cancel`
3. 科研
   - `/api/cities/:cid/research/start` -> `/cancel`
4. 征兵
   - `/api/cities/:cid/barracks/draft/start` -> `/cancel`
5. 出征
   - `/api/cities/:cid/troops/dispatch` -> `/api/troops/:id/callback`

## 实测结果

脚本输出 `ok=true`，关键字段如下：

- building:
  - `create_state=1`
  - `exists_after_cancel=false`
  - `gov_state_after_upgrade=1`
  - `gov_state_after_cancel=0`
- research:
  - `active_after_start=tid`
  - `active_after_cancel=0`
  - `state_after_start=1`
  - `state_after_cancel=0`
- draft:
  - `queue_after_start = queue_before + 1`
  - `queue_after_cancel = queue_before`
- troop:
  - `startCid` 按出发城正确写入（不再是 `0`）
  - `state_after_dispatch=0`（行军）
  - callback 后允许两种结果：
    - 返回列表中仍可见该 troop 且 `state=1`（返程）
    - 或 callback 后立即结算完成，列表中不可见（`state_after_callback=-1`）

## 结论

- 当前“建筑/科研/征兵/出征”核心状态机链路已具备可回归的自动化基线。
- 该结果仅说明“能稳定跑通并防回退”，不等于 1:1 细节完全一致。
