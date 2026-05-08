# Claude Round 3 Parallel Task Board: BloodWar Frontend 1:1

Date: 2026-05-03
Workspace: `D:\APMServ5.2.6\new_project`
Max parallel workers allowed: 9

## Current Check Result

Commands just checked:

```powershell
cd D:\APMServ5.2.6\new_project\frontend
npm run build

cd D:\APMServ5.2.6\new_project\backend
go test ./...
```

Both pass.

But P0 is not clean:

- `verify-speed-flow.ps1` can fail under concurrent script runs with `error="forbidden"`.
- `verify-speed-flow.ps1` still ends with warning-only task state verification:
  - `[WARN] Could not verify task status via API`
- `verify-speed-flow-coordinates.ps1` now tries to fail if fallback is used, but fallback code is still present.
- DB check inside `verify-speed-flow-coordinates.ps1` is broken: MySQL prints help text instead of query result, likely because CLI args are malformed.
- `debug-pure-coordinates.ps1` proves pure coordinates are not actually aligned:
  - `gid11` coordinate hits `IMG` and advances.
  - `gid12` coordinate hits `DIV.task-list`, not `task-item`, and does not advance.
  - `gid13` coordinate hits `BUTTON.task-close`, not `task-claim-btn`, and does not claim.

Therefore:

- Functional chain is still useful.
- Coordinate chain is not truly 1:1 clean.
- Do not claim P0 complete until the pure coordinate script passes without fallback and task state proof works.

## Parallel Execution Rules

- Do not let multiple workers edit the same files.
- Do not run smoke scripts in parallel against the same browser/backend unless the script creates isolated browser targets and isolated debug passports.
- Evidence-only workers may run in parallel freely.
- Code-changing workers must list changed files.
- No worker may claim "1:1 complete" without evidence path, screenshot path, and click-path output.

## Worker 1: P0 Regression Owner

Ownership:

- `artifacts\verify-speed-flow.ps1`
- `artifacts\verify-speed-flow-check.py`
- `artifacts\verify-speed-flow-coordinates.ps1`
- `artifacts\verify-speed-flow-coordinates-check.py`
- optional new helper under `artifacts\regression\`

Tasks:

1. Fix MySQL invocation for DB task-state verification.
2. Make task-state verification work in both smoke scripts for the exact generated passport.
3. Remove warning-only pass behavior.
4. Make concurrent runs impossible or clearly blocked:
   - Either serialize via a lock file.
   - Or make scripts fully isolated.
5. Remove direct DOM fallback clicks from coordinate regression, or fail before fallback can mutate state.

Acceptance:

```powershell
powershell -ExecutionPolicy Bypass -File D:\APMServ5.2.6\new_project\artifacts\verify-speed-flow.ps1
powershell -ExecutionPolicy Bypass -File D:\APMServ5.2.6\new_project\artifacts\verify-speed-flow-coordinates.ps1
```

Both must pass sequentially with:

- no `WARN`
- no `forbidden`
- no MySQL help text
- no fallback used
- DB proves task 1 claimed/no longer claimable

## Worker 2: P0 Coordinate Alignment Owner

Ownership:

- `frontend\src\App.vue`
- `frontend\src\style.css`
- `docs\guide-coordinate-model-2026-05-03.md`

Tasks:

1. Fix `gid12` pure coordinate hit.
   - Required click: guide showpos center around `(384,404)`.
   - Current pure debug hit: `DIV.task-list`.
   - Required result: `task-item` and guide advances to `gid13`.
2. Fix `gid13` pure coordinate hit.
   - Required click: guide showpos center around `(833,517)`.
   - Current pure debug hit: `BUTTON.task-close`.
   - Required result: `task-claim-btn` and guide advances to final text.
3. Explain coordinate systems:
   - viewport coordinate
   - `.stage` top/left offset
   - Flash-stage coordinate
   - CSS absolute coordinate
   - `getBoundingClientRect()` coordinate
4. Keep `gid10` and earlier guide holes working.

Acceptance:

- `debug-pure-coordinates.ps1` must show:
  - `gid12` result `BUTTON.task-item` or equivalent clickable task item.
  - `gid13` result `BUTTON.task-claim-btn`.
- `verify-speed-flow-coordinates.ps1` must pass without fallback.
- `docs\guide-coordinate-model-2026-05-03.md` must include a rect table for gid6-gid14.

## Worker 3: Login 1:1 Evidence Owner

Ownership:

- `docs\login-1to1-evidence-2026-05-03.md`
- `artifacts\html5-login-round3.png`
- Read-only on frontend unless Worker 2 is done and no conflict is pending.

Tasks:

1. Read FFDec login evidence:
   - `artifacts\ffdec\BloodWar\scripts\Login\*.as`
   - `artifacts\ffdec\BloodWar\scripts\BloodWar.as`
2. Identify login background, logo, input frames, button states, and coordinates.
3. Compare current HTML5 login screenshot to evidence.
4. Record exact/approximate/unknown deltas.

Acceptance:

- Evidence doc exists.
- Screenshot exists.
- Login click path is described.
- No code changes unless explicitly coordinated after P0.

## Worker 4: Create Role / Province Evidence Owner

Ownership:

- `docs\create-role-1to1-evidence-2026-05-03.md`
- `artifacts\html5-create-role-round3.png`

Tasks:

1. Re-read:
   - `artifacts\ffdec\BloodWar\scripts\Login\CreateRoleDialog.as`
2. Extract exact:
   - dialog size
   - province map container
   - province overlay assets and coords
   - portrait coords/assets
   - sex toggle behavior
   - name input behavior
   - city name behavior
   - start button position
3. Compare with current HTML5.
4. Produce a patch plan, but do not edit `App.vue`/`style.css` in parallel with Worker 2.

Acceptance:

- Evidence doc includes a table of controls and coordinates.
- Fresh account create-role click path is documented.
- Screenshot saved if currently reachable.

## Worker 5: City Shell Evidence Owner

Ownership:

- `docs\city-shell-1to1-evidence-2026-05-03.md`
- `artifacts\html5-city-round3.png`

Tasks:

1. Read:
   - `artifacts\ffdec\BloodWar\scripts\CityInnerPanel.as`
   - `artifacts\ffdec\BloodWar\scripts\CityPanel.as`
   - `artifacts\ffdec\BloodWar\scripts\Bar\TopPanel.as`
   - `artifacts\ffdec\BloodWar\scripts\Bar\BottomPanel.as`
   - `artifacts\ffdec\BloodWar\scripts\Building\BuildingGrid.as`
2. Build a city shell delta table:
   - top bar
   - left resource panel
   - inner background
   - building grid
   - bottom bar
   - guide overlay
3. Include current rects and Flash expected rects.

Acceptance:

- Evidence doc exists.
- Screenshot exists.
- Table has `expected`, `current`, `delta`, `status`.

## Worker 6: Build Panel Evidence Owner

Ownership:

- `docs\build-panel-1to1-evidence-2026-05-03.md`
- optional screenshots under `artifacts\build-panel-*.png`

Tasks:

1. Read:
   - `CreateBuildingDialog.as`
   - `CreateBuildingItem.as`
   - `UpgradeBuildingDialog.as`
   - `UpgradeBuildingItem.as`
   - `DialogManager.as`
2. Map:
   - panel size
   - item list position
   - `民房` item row
   - build button
   - busy building info panel
   - speed button
3. Compare with current HTML5.

Acceptance:

- Evidence doc exists.
- Do not modify shared frontend files in this worker.

## Worker 7: UseGoodsDialog Evidence / Backend Behavior Owner

Ownership:

- `docs\use-goods-dialog-1to1-evidence-2026-05-03.md`
- optional backend notes only, no code unless assigned separately.

Tasks:

1. Read:
   - `artifacts\ffdec\BloodWar\scripts\Goods\UseGoodsDialog.as`
   - `artifacts\ffdec\BloodWar\scripts\DialogManager.as`
2. Verify item behavior for goods `67..73`.
3. Compare current backend speed effects with Flash evidence.
4. Identify exact gaps.

Acceptance:

- Evidence doc lists each goods id, original effect, current effect, status.
- Do not change backend in this worker unless P0 regression is already clean.

## Worker 8: TaskDialog Evidence Owner

Ownership:

- `docs\task-dialog-1to1-evidence-2026-05-03.md`

Tasks:

1. Read:
   - `artifacts\ffdec\BloodWar\scripts\Task\TaskDialog.as`
   - any related Task item/render classes in FFDec export
2. Map:
   - dialog size and root position
   - category/group list
   - task row
   - selected task detail
   - reward list
   - claim button
   - close button
3. Compare with current simplified task panel.

Acceptance:

- Evidence doc lists exactly what is implemented, stubbed, missing, or blocked.
- Include how current `gid12/gid13` placement should be fixed for 1:1.

## Worker 9: Evidence Index / Final Gate Owner

Ownership:

- `docs\frontend-1to1-status-index-2026-05-03.md`

Tasks:

1. Create a single index of all frontend parity areas:
   - login
   - create role
   - city shell
   - guide chain
   - build panel
   - use goods
   - task dialog
   - world map
   - battle
   - utility windows
2. For each area, record:
   - status: `verified`, `partial`, `blocked`, `not started`
   - owner worker
   - evidence doc path
   - screenshot path
   - regression path
   - next action
3. Keep this as the source of truth for progress.

Acceptance:

- Index exists and is honest.
- P0 remains `partial` until Worker 1 and Worker 2 pass cleanly.

## Recommended Parallel Schedule

Start immediately in parallel:

- Worker 1: regression scripts.
- Worker 2: coordinate alignment.
- Worker 3: login evidence.
- Worker 4: create-role evidence.
- Worker 5: city shell evidence.
- Worker 6: build panel evidence.
- Worker 7: UseGoods evidence.
- Worker 8: TaskDialog evidence.
- Worker 9: status index.

Only Worker 1 and Worker 2 should edit executable code right now.

Evidence workers should not edit `frontend\src\App.vue` or `frontend\src\style.css` until Worker 2 is done, otherwise conflicts will slow everyone down.

## Final Report Required From Claude

Report by worker:

```text
Worker 1 Regression:
- Status:
- Commands run:
- Files changed:
- Remaining blocker:

Worker 2 Coordinate Alignment:
- Status:
- gid12 hit:
- gid13 hit:
- Files changed:
- Remaining blocker:

Worker 3 Login Evidence:
- Status:
- Evidence doc:
- Screenshot:
- Gaps:

...

Global Gate:
- npm run build:
- go test ./...:
- verify-speed-flow.ps1:
- verify-speed-flow-coordinates.ps1:
- P0 status:
```

No vague "mostly done". Every completion needs a path and a command or screenshot.
