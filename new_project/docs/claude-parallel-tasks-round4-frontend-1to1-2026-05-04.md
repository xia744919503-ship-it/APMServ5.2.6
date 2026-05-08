# Claude Round 4 Parallel Task Board: BloodWar Frontend 1:1

Date: 2026-05-04
Workspace: `D:\APMServ5.2.6\new_project`
Max parallel workers allowed: 9

## Current Verified Status

Checked on 2026-05-04:

```powershell
cd D:\APMServ5.2.6\new_project\frontend
npm run build

cd D:\APMServ5.2.6\new_project\backend
go test ./...

cd D:\APMServ5.2.6\new_project
powershell -ExecutionPolicy Bypass -File .\artifacts\verify-speed-flow.ps1
powershell -ExecutionPolicy Bypass -File .\artifacts\verify-speed-flow-coordinates.ps1
```

Results:

- `npm run build`: PASS
- `go test ./...`: PASS, but no test files
- `verify-speed-flow.ps1`: PASS for gid6->gid14 functional chain, but still prints `[WARN] Could not verify task status via API`
- `verify-speed-flow-coordinates.ps1`: exits 0 and DB proves `task1_state=1`, but g13 uses a direct claim helper (`Call-Claim-Reward`) instead of pure coordinate/mouse click

Important: **P0 is not honestly complete yet.**

## Work Completed Since Round 3

Documentation/evidence files were created:

- `docs\login-1to1-evidence-2026-05-04.md`
- `docs\createrole-1to1-evidence-2026-05-04.md`
- `docs\city-1to1-evidence-2026-05-04.md`
- `docs\building-panel-1to1-evidence-2026-05-04.md`
- `docs\usegoods-dialog-1to1-evidence-2026-05-04.md`
- `docs\task-dialog-1to1-evidence-2026-05-04.md`
- `docs\evidence-index-2026-05-04.md`

These are useful, but most are marked `partial`. Do not treat them as implementation completion.

Screenshots/assets observed:

- `artifacts\screenshots\chrome-login-ffdec-align.png`
- `artifacts\screenshots\chrome-login-ffdec-align-2.png`
- `artifacts\screenshots\chrome-create-role-ffdec-align.png`
- `artifacts\screenshots\chrome-create-role-ffdec-align-2.png`
- `artifacts\screenshots\chrome-build-dialog-ffdec-align.png`
- extracted FFDec UI/button assets under `artifacts\ffdec-*`

## Blocking Problems

### Blocker A: Functional Smoke Still Has Warning-Only Task State Check

`verify-speed-flow.ps1` still passes while printing:

```text
[WARN] Could not verify task status via API
```

That means the functional smoke test does not prove the final DB/API state.

### Blocker B: Coordinate Regression Is Not Pure

`verify-speed-flow-coordinates.ps1` contains and uses:

```powershell
Call-Claim-Reward
```

Latest output:

```text
[g13] Call-Claim-Reward result: task-claim-btn-clicked
g13_hit: task-claim-btn-clicked
```

This is not a pure coordinate click. It is a helper-triggered action.

### Blocker C: Coordinate Regression Ignores Stale Backend Error Text

`verify-speed-flow-coordinates-check.py` still allows:

```text
driver: bad connection
```

Do not allow backend errors in final gates. Clear the UI error state or fix the backend/session issue.

### Blocker D: Evidence Index Is Too Optimistic

`docs\evidence-index-2026-05-04.md` says workers are complete, while the detailed evidence docs say `status: partial`.

Fix the index to distinguish:

- `evidence drafted`
- `implemented`
- `screenshot verified`
- `click-path verified`
- `1:1 accepted`

## Parallel Work Rules

- Only Workers 1-3 may edit executable code this round.
- Workers 4-9 should produce docs, screenshots, and patch plans unless explicitly unblocked.
- Do not let multiple workers edit `frontend\src\App.vue` or `frontend\src\style.css` at the same time.
- Run regression scripts sequentially, not concurrently.
- No "done" without command output, screenshot path, or exact evidence citation.

## Worker 1: P0 Regression Gate Owner

Ownership:

- `artifacts\verify-speed-flow.ps1`
- `artifacts\verify-speed-flow-check.py`
- `artifacts\verify-speed-flow-coordinates.ps1`
- `artifacts\verify-speed-flow-coordinates-check.py`
- optional helper scripts under `artifacts\regression\`

Tasks:

1. Add DB state verification to `verify-speed-flow.ps1`, same as coordinate script.
2. Remove warning-only pass behavior.
3. Remove `Call-Claim-Reward` from coordinate regression.
4. Use real browser mouse events or CDP `Input.dispatchMouseEvent` at Flash-stage coordinates.
5. Fail if any `.click()` helper is used for gid11/gid12/gid13.
6. Fail if `debug.errorText` is non-empty.

Acceptance:

```powershell
powershell -ExecutionPolicy Bypass -File D:\APMServ5.2.6\new_project\artifacts\verify-speed-flow.ps1
powershell -ExecutionPolicy Bypass -File D:\APMServ5.2.6\new_project\artifacts\verify-speed-flow-coordinates.ps1
```

Both must:

- exit 0
- have no WARN
- have no `driver: bad connection`
- prove task 1 state for the generated passport
- use no direct DOM helper for gid11/gid12/gid13

## Worker 2: Gid13 Pure Coordinate Fix Owner

Ownership:

- `frontend\src\App.vue`
- `frontend\src\style.css`
- `docs\guide-coordinate-model-2026-05-04.md`

Tasks:

1. Make the Flash showpos center for gid13 hit the actual claim button with `elementFromPoint` or CDP mouse event.
2. Current evidence:
   - guide showpos: `795,504,77,26`
   - desired center: approximately `(833,517)`
   - current claim button rect in output: `x=806.5,y=544.5,width=77,height=26`
   - current script bypasses with `Call-Claim-Reward`
3. Fix layout or coordinate conversion. Do not use helper click.
4. Keep gid6-gid12 passing.

Acceptance:

- `verify-speed-flow-coordinates.ps1` logs `g13_hit=task-claim-btn` from coordinate/mouse target, not `task-claim-btn-clicked`.
- Final guide text appears.
- DB state is `1`.
- `docs\guide-coordinate-model-2026-05-04.md` explains stage vs viewport coordinates and includes gid6-gid14 rects.

## Worker 3: P1 Login/Create-Role Implementation Owner

Ownership:

- `frontend\src\App.vue`
- `frontend\src\style.css`
- `frontend\src\api.ts` only if needed
- screenshots under `artifacts\screenshots\`

Only start after Worker 2 has either finished or confirmed no conflict in `App.vue/style.css`.

Tasks:

1. Convert the login screen from `partial evidence` toward real alignment:
   - absolute Flash-stage positioning
   - original input/button coordinates
   - bitmap button states where available
2. Convert create-role screen toward Flash evidence:
   - province map container `333,63,618,500`
   - left panel and start button coordinates
   - portrait controls and gender controls
3. Preserve working login/create role behavior.

Acceptance:

- Save:
  - `artifacts\screenshots\chrome-login-round4.png`
  - `artifacts\screenshots\chrome-create-role-round4.png`
- Add click-path notes to the corresponding docs.
- Do not claim 1:1 unless screenshot and coordinate table are present.

## Worker 4: Evidence Index Truth Owner

Ownership:

- `docs\evidence-index-2026-05-04.md`
- optional `docs\frontend-1to1-status-index-2026-05-04.md`

Tasks:

1. Rewrite the index so it does not mark partial docs as complete.
2. Use this status vocabulary:
   - `not_started`
   - `evidence_drafted`
   - `implemented_partial`
   - `functional_verified`
   - `coordinate_verified`
   - `screenshot_verified`
   - `1to1_accepted`
3. For each area, record:
   - evidence doc
   - screenshot path
   - regression path
   - implementation status
   - next action

Acceptance:

- P0 guide chain is no higher than `functional_verified` until Worker 1/2 pass pure coordinate.
- Login/create-role/city/build/usegoods/task remain `evidence_drafted` or `implemented_partial`, not accepted.

## Worker 5: City Shell Patch Plan Owner

Ownership:

- `docs\city-1to1-evidence-2026-05-04.md`
- `docs\city-shell-patch-plan-2026-05-04.md`

Tasks:

1. Turn the city evidence doc into an actionable patch plan.
2. Include:
   - current rect
   - Flash rect
   - delta
   - asset name
   - file to edit
   - risk to gid6-gid14
3. Prioritize first:
   - top bar
   - left resource panel
   - inner city map position
   - bottom bar

Acceptance:

- No code changes.
- Patch plan can be handed to a coding worker without re-reading all AS files.

## Worker 6: Build/UseGoods Patch Plan Owner

Ownership:

- `docs\building-panel-1to1-evidence-2026-05-04.md`
- `docs\usegoods-dialog-1to1-evidence-2026-05-04.md`
- `docs\build-usegoods-patch-plan-2026-05-04.md`

Tasks:

1. Convert evidence into patchable tasks.
2. Separate:
   - CreateBuildingDialog visual fixes
   - Building info panel fixes
   - UseGoodsDialog layout fixes
   - goods 67-73 backend behavior gaps
3. Mark which items are safe before P0 and which must wait.

Acceptance:

- No code changes.
- Each planned change has file path and acceptance check.

## Worker 7: TaskDialog Patch Plan Owner

Ownership:

- `docs\task-dialog-1to1-evidence-2026-05-04.md`
- `docs\task-dialog-patch-plan-2026-05-04.md`

Tasks:

1. Turn TaskDialog evidence into implementation slices:
   - dialog root/position
   - category buttons
   - task rows
   - selected task detail
   - rewards
   - claim button
2. Explicitly separate:
   - gid12/gid13 tutorial-critical subset
   - full task window after gid14

Acceptance:

- No code changes unless Worker 2 asks for TaskDialog-specific coordinate help.
- Patch plan includes acceptance tests.

## Worker 8: Screenshot / Visual Evidence Owner

Ownership:

- `artifacts\screenshots\*round4*.png`
- `docs\screenshot-index-2026-05-04.md`

Tasks:

1. Capture current HTML5 screenshots at `1000x600` for:
   - login
   - create-role
   - city
   - build panel
   - speed goods dialog
   - task dialog
2. Record exact URL/debug params used.
3. Do not use screenshots as proof of 1:1 by themselves.

Acceptance:

- `docs\screenshot-index-2026-05-04.md` maps screenshot path to screen/state/debug URL.
- Screenshots are outside Chrome profile cache folders.

## Worker 9: Backend Compatibility / Risk Owner

Ownership:

- `docs\backend-compatibility-gaps-2026-05-04.md`

Tasks:

1. Review recent backend changes:
   - `backend\internal\legacy\account_role_create.go`
   - `city_building_goods.go`
   - `city_building_info.go`
   - `city_building_create.go`
   - `guide.go`
   - task claim files
2. Document:
   - debug-only behavior
   - temporary starter goods hacks
   - goods 67-73 behavior gaps
   - task claim assumptions
   - any divergence from Flash AMF semantics

Acceptance:

- No code changes.
- Produce risk table with severity and suggested owner.

## Required Final Report Format

Claude must report by worker:

```text
Worker 1 Regression:
- Status:
- Commands:
- Remaining blocker:

Worker 2 Coordinate:
- Status:
- g13 hit:
- DB state:
- Remaining blocker:

Worker 3 Login/CreateRole:
- Status:
- Screenshots:
- Files changed:

Worker 4 Index:
- Status:
- Index path:

...

Global Gate:
- npm run build:
- go test ./...:
- verify-speed-flow.ps1:
- verify-speed-flow-coordinates.ps1:
- P0 accepted: yes/no
```

If any gate has warnings, `P0 accepted` must be `no`.
