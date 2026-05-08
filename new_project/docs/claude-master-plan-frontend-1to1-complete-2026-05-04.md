# Claude Master Plan: BloodWar Frontend 1:1 Completion

Date: 2026-05-04
Workspace: `D:\APMServ5.2.6\new_project`
Max parallel workers: 9

## Mission

Finish the Flash/SWF-derived HTML5 reconstruction honestly. Do not produce fake panels, generic icons, modern UI, or "looks close" claims. The source of truth is:

1. FFDec ActionScript and exported assets.
2. DB config and live backend state.
3. Original Flash screenshots/runtime observations.
4. HTML5 regression output and screenshots.

The final target is not "the current tutorial sort of works". The final target is:

- login works and visually follows Flash evidence,
- create role / province selection works and visually follows Flash evidence,
- city shell is aligned enough that guide holes land on real visible controls,
- build -> speed goods -> task reward tutorial chain passes pure coordinate regression,
- core panels used in that chain are Flash-derived,
- remaining world/battle/utility gaps are explicitly tracked and not hidden.

## Current Truth As Of 2026-05-04

Already passing:

- `npm run build`
- `go test ./...`
- `verify-speed-flow.ps1` functionally reaches `gid14`
- `verify-speed-flow-coordinates.ps1` exits 0 and proves DB task state in latest runs

Still not acceptable:

- `verify-speed-flow.ps1` still has warning-only task state verification in some runs.
- `verify-speed-flow-coordinates.ps1` recently used helper behavior around `g13` (`Call-Claim-Reward` / `task-claim-btn-clicked`) instead of pure mouse/coordinate proof.
- `verify-speed-flow-coordinates-check.py` has allowed stale backend error text such as `driver: bad connection`.
- Evidence docs are mostly `partial`; they are useful drafts, not completion certificates.
- `evidence-index-2026-05-04.md` must not call partial evidence `1to1_accepted`.

Therefore P0 is **not accepted** until all gates below are clean.

## Non-Negotiable Final Gates

These commands must pass in a fresh sequential run:

```powershell
cd D:\APMServ5.2.6\new_project\frontend
npm run build

cd D:\APMServ5.2.6\new_project\backend
go test ./...

cd D:\APMServ5.2.6\new_project
powershell -ExecutionPolicy Bypass -File .\artifacts\verify-speed-flow.ps1
powershell -ExecutionPolicy Bypass -File .\artifacts\verify-speed-flow-coordinates.ps1
```

Final gate requirements:

- no `WARN`
- no `forbidden`
- no MySQL help text
- no stale error text
- no `driver: bad connection`
- no helper `.click()` for `gid11/gid12/gid13`
- `g11` must advance via real coordinate/mouse target
- `g12` must hit task row and advance via real coordinate/mouse target
- `g13` must hit claim button and advance via real coordinate/mouse target
- DB proves task 1 is claimed/no longer claimable for the generated passport

If any condition fails, P0 remains incomplete.

## Parallel Work Model

Only one worker may edit `frontend\src\App.vue` and `frontend\src\style.css` at a time. Evidence-only workers can run in parallel. Backend workers must avoid touching frontend files.

Recommended concurrency:

- Worker 1 and Worker 2 start immediately and block P0 acceptance.
- Workers 3-9 run in parallel on docs/screenshots/patch plans.
- Implementation workers after P0 should work in disjoint slices.

## Worker 1: Regression And Gate Owner

Ownership:

- `artifacts\verify-speed-flow.ps1`
- `artifacts\verify-speed-flow-check.py`
- `artifacts\verify-speed-flow-coordinates.ps1`
- `artifacts\verify-speed-flow-coordinates-check.py`
- optional helper files under `artifacts\regression\`

Tasks:

1. Add DB state verification to `verify-speed-flow.ps1`.
2. Remove all warning-only success paths.
3. Remove/forbid helper clicks for `gid11/gid12/gid13`.
4. Use CDP `Input.dispatchMouseEvent` or equivalent real mouse coordinates.
5. Record stage coordinate, viewport coordinate, `elementFromPoint`, guide text before/after, and DB state.
6. Fail on any non-empty `.error-line`.
7. Add a script lock or clear isolation so two smoke scripts cannot corrupt each other.

Acceptance:

- Both regression scripts pass sequentially.
- Running them concurrently either safely isolates or clearly exits with a lock message.
- JSON outputs are written and easy to audit.

## Worker 2: Tutorial Coordinate Alignment Owner

Ownership:

- `frontend\src\App.vue`
- `frontend\src\style.css`
- `docs\guide-coordinate-model-2026-05-04.md`

Tasks:

1. Make `gid11` pure coordinate click open task panel.
2. Make `gid12` pure coordinate click hit `.task-item` and advance to `gid13`.
3. Make `gid13` pure coordinate click hit `.task-claim-btn` and claim reward.
4. Align task row/claim button to Flash `showpos` without fake invisible click proxies unless documented as Flash mask equivalent.
5. Document stage vs viewport vs `getBoundingClientRect()` math.

Required showpos:

- `gid11`: `912,4,68,24`
- `gid12`: `312,392,145,24`
- `gid13`: `795,504,77,26`

Acceptance:

- `verify-speed-flow-coordinates.ps1` logs real targets, not helper results.
- `docs\guide-coordinate-model-2026-05-04.md` has gid6-gid14 table with expected showpos, current rect, hit target, status.

## Worker 3: Login And Create Role Implementation Owner

Ownership:

- `frontend\src\App.vue`
- `frontend\src\style.css`
- `frontend\src\api.ts` only if needed
- `docs\login-1to1-evidence-2026-05-04.md`
- `docs\createrole-1to1-evidence-2026-05-04.md`
- screenshots under `artifacts\screenshots\`

Start only after Worker 2 releases frontend files.

Tasks:

1. Login:
   - align background, account input, password input, login button using Flash evidence,
   - use bitmap button assets where available,
   - preserve existing `legacyLogin`.
2. Create role:
   - align left panel,
   - province map container `333,63,618,500`,
   - province overlays,
   - portrait previous/next,
   - gender controls,
   - name input,
   - province selector,
   - start button `67.5,449.3,115,42`,
   - preserve create-role API behavior.
3. Save screenshots:
   - `artifacts\screenshots\chrome-login-final.png`
   - `artifacts\screenshots\chrome-create-role-final.png`

Acceptance:

- fresh account can create role and enter city,
- existing account can login,
- docs include exact Flash references and current screenshot paths.

## Worker 4: City Shell Implementation Owner

Ownership:

- `frontend\src\App.vue`
- `frontend\src\style.css`
- `docs\city-1to1-evidence-2026-05-04.md`
- `docs\city-shell-patch-plan-2026-05-04.md`

Start after P0 coordinates are stable.

Tasks:

1. Align city shell to Flash layout:
   - top bar,
   - left resource panel,
   - city map stage,
   - inner background,
   - building grid and hit zones,
   - bottom/chat/function bar if available.
2. Keep gid6-gid14 regression passing.
3. Do not introduce broad responsive redesign.

Acceptance:

- `artifacts\screenshots\chrome-city-final.png`
- city evidence doc has expected/current/delta table,
- guide holes land on visible controls.

## Worker 5: Build Panel And Building Info Owner

Ownership:

- frontend files only after Worker 4 releases them
- `docs\building-panel-1to1-evidence-2026-05-04.md`
- `docs\build-usegoods-patch-plan-2026-05-04.md`

Tasks:

1. Align `CreateBuildingDialog`.
2. Align `CreateBuildingItem` row for `民房`.
3. Align occupied building info panel.
4. Align speed button.
5. Implement/mark blocked close, cancel, destroy, upgrade states.

Acceptance:

- empty tile opens build panel,
- `民房` build works,
- busy `民房` opens info panel,
- speed button opens goods dialog,
- screenshot saved: `artifacts\screenshots\chrome-build-panel-final.png`.

## Worker 6: UseGoodsDialog And Goods Effects Owner

Ownership:

- `frontend\src\App.vue`
- `frontend\src\style.css`
- backend goods files only if needed:
  - `backend\internal\legacy\city_building_goods.go`
  - `backend\internal\legacy\models.go`
  - service/router files if routes change
- `docs\usegoods-dialog-1to1-evidence-2026-05-04.md`

Tasks:

1. Align UseGoodsDialog panel, item list, item icon frame, close/buy buttons.
2. Preserve gid10 click.
3. Verify goods behavior:
   - `67`: 15 minutes
   - `68`: 1 hour
   - `69`: 2.5 hours
   - `70`: 8 hours
   - `71`: random 10-30 hours if confirmed
   - `72`: 30% remaining if confirmed
   - `73`: instant/yuanbao if confirmed
4. Non-owned goods must be disabled/blocked like Flash.

Acceptance:

- gid10 still passes,
- backend mutation correct,
- screenshot saved: `artifacts\screenshots\chrome-usegoods-final.png`,
- evidence doc says exact/current status per goods id.

## Worker 7: TaskDialog Full Window Owner

Ownership:

- `frontend\src\App.vue`
- `frontend\src\style.css`
- `frontend\src\api.ts`
- task backend only if needed
- `docs\task-dialog-1to1-evidence-2026-05-04.md`

Tasks:

1. Separate tutorial-critical task subset from full task window.
2. Implement full task dialog elements:
   - categories,
   - groups,
   - task rows,
   - selected task detail,
   - goals,
   - rewards,
   - claim button,
   - close button,
   - completed/incomplete states.
3. Preserve gid12/gid13 coordinate acceptance.

Acceptance:

- first reward tutorial still passes,
- after gid14 task window does not dead-end,
- screenshot saved: `artifacts\screenshots\chrome-task-dialog-final.png`.

## Worker 8: World / Battle / Utility Evidence And Implementation Backlog Owner

Ownership:

- docs only unless P0-P1 are already accepted
- `docs\world-battle-utility-backlog-2026-05-04.md`

Tasks:

1. Create final backlog for:
   - outer city,
   - world map,
   - battle,
   - report,
   - mail,
   - union,
   - rank,
   - shop,
   - hero/recruit,
   - research,
   - barracks/training.
2. For each:
   - FFDec source files,
   - assets,
   - backend route/state needs,
   - minimum playable click path,
   - acceptance test.

Acceptance:

- no fake implementation,
- every future area has evidence-first task list.

## Worker 9: Status Index / Screenshot / Final Report Owner

Ownership:

- `docs\evidence-index-2026-05-04.md`
- `docs\screenshot-index-2026-05-04.md`
- final report doc:
  - `docs\frontend-1to1-final-gate-report-2026-05-04.md`

Tasks:

1. Fix status index so no partial doc is marked accepted.
2. Capture and catalog all final screenshots.
3. Maintain status vocabulary:
   - `not_started`
   - `evidence_drafted`
   - `implemented_partial`
   - `functional_verified`
   - `coordinate_verified`
   - `screenshot_verified`
   - `1to1_accepted`
4. Produce final gate report.

Acceptance:

- final report includes all command outputs,
- every screenshot path exists,
- every accepted item has evidence + screenshot + click-path.

## Execution Order

Phase 0:

1. Worker 1 fixes gates.
2. Worker 2 fixes coordinates.
3. Workers 4-9 prepare docs and plans in parallel.

Phase 1:

1. Worker 3 implements login/create-role.
2. Worker 4 implements city shell.
3. Worker 5 implements build panel.
4. Worker 6 implements UseGoods.
5. Worker 7 implements TaskDialog.

Phase 2:

1. Full regression.
2. Screenshot capture.
3. Evidence index update.
4. Final gate report.

Phase 3:

Only after the above is accepted, start world/battle/utility implementation.

## Final Report Required From Claude

Claude must report exactly:

```text
Global Gate:
- npm run build:
- go test ./...:
- verify-speed-flow.ps1:
- verify-speed-flow-coordinates.ps1:
- P0 accepted: yes/no

Implemented:
- Login:
- Create role:
- City shell:
- Build panel:
- UseGoods:
- TaskDialog:

Evidence:
- docs:
- screenshots:
- JSON outputs:

Known gaps:
- ...

Files changed:
- ...

Next backlog:
- world:
- battle:
- utility windows:
```

If any command has warnings, helper clicks, stale errors, or missing DB proof, report `P0 accepted: no`.
