# Claude Round 2 Tasks: BloodWar 1:1 Frontend

Date: 2026-05-03
Workspace: `D:\APMServ5.2.6\new_project`

## What Was Just Verified

I re-ran the current checks sequentially.

Passed:

```powershell
cd D:\APMServ5.2.6\new_project\frontend
npm run build

cd D:\APMServ5.2.6\new_project\backend
go test ./...

cd D:\APMServ5.2.6\new_project
powershell -ExecutionPolicy Bypass -File D:\APMServ5.2.6\new_project\artifacts\verify-speed-flow.ps1
powershell -ExecutionPolicy Bypass -File D:\APMServ5.2.6\new_project\artifacts\verify-speed-flow-coordinates.ps1
```

Current good news:

- Functional guide chain `gid6 -> gid14` passes.
- Coordinate chain also reaches `gid14`.
- The strict coordinate output shows the intended targets are hit:
  - `g11_hit = guide-task-btn-clickable`
  - `g12_hit = task-item`
  - `g13_hit = task-claim-btn`
- `claimBtnRect.left = 795`, matching `gid13 showpos=795,504,77,26` on X/width.

Do not overclaim:

- This only closes the current tutorial chain around build/speed/task claim.
- It does not mean full frontend 1:1 is done.
- Login, create-role, city shell, world map, battle, task system beyond first reward, and utility windows are not fully 1:1.

## Problems Still Present

### Problem 1: Functional Smoke Still Cannot Prove Backend Task State

`verify-speed-flow.ps1` still ends with:

```text
[WARN] Could not verify task status via API
```

This is not acceptable as a final gate. After task 1 is claimed, the test must prove task 1 is claimed/no longer claimable for the same debug user.

### Problem 2: Coordinate Script Is Not Pure Enough

`artifacts\verify-speed-flow-coordinates.ps1` says it uses only stage coordinates, but it still contains direct DOM fallback helpers:

- `Click-GuideBtn`
- `Click-TaskRow`
- `Click-ClaimBtn`

The latest successful output did not need these fallbacks for `gid11-gid13`, but the script should prove that directly.

### Problem 3: Coordinate Script Allows A Backend Error Text

The latest passing coordinate output included:

```text
debug.errorText = driver: bad connection
```

The Python checker currently allows this. Do not allow backend error text in a clean 1:1 gate unless it is proven to be a harmless stale UI message and cleared before final state.

### Problem 4: Gid13 Y Coordinate Needs Explanation

Flash guide `gid13` has `showpos=795,504,77,26`, but current debug rect says:

```text
claimBtnRect = x:795, y:544.5, width:77, height:26
```

Yet clicking `(833,517)` still hits `task-claim-btn`. Explain this by inspecting stage offset, guide overlay behavior, and elementFromPoint coordinates. If it is a test coordinate bug, fix it. If it is CSS transform/stage offset, document it.

## Immediate P0 Tasks

### P0.1 Fix Backend Task State Verification

Update both:

- `artifacts\verify-speed-flow.ps1`
- `artifacts\verify-speed-flow-coordinates.ps1`

Acceptance:

- Both scripts verify the task state for the exact passport generated in that run.
- Prefer direct DB check by `debugPassport` if the browser session auth cannot be reused reliably.
- Scripts fail non-zero if task 1 remains claimable or state cannot be proven.
- No more warning-only verification.

Suggested DB check:

```sql
SELECT u.uid,u.passport,t.id,t.name,ut.state
FROM sys_user u
JOIN sys_user_task ut ON ut.uid=u.uid
JOIN cfg_task t ON t.id=ut.tid
WHERE u.passport='<passport>'
ORDER BY t.id;
```

Expected: task 1 should be claimed/no longer claimable according to current backend model.

### P0.2 Make Coordinate Regression Truly Coordinate-Based

Update:

- `artifacts\verify-speed-flow-coordinates.ps1`
- `artifacts\verify-speed-flow-coordinates-check.py`

Required:

- Remove direct DOM fallback click helpers, or make the script fail if fallback is used.
- Record actual `elementFromPoint` target at each coordinate.
- For `gid11`, `gid12`, `gid13`, fail unless coordinate hit is exactly the intended clickable:
  - `gid11`: `guide-task-btn-clickable` or the actual top task image if the overlay is removed.
  - `gid12`: `task-item`.
  - `gid13`: `task-claim-btn`.
- Do not accept `none`, `task-list`, `guide-tip-content`, or fallback-click result as pass.

Acceptance:

```powershell
powershell -ExecutionPolicy Bypass -File D:\APMServ5.2.6\new_project\artifacts\verify-speed-flow-coordinates.ps1
```

must pass without any fallback and without error text.

### P0.3 Fix Or Explain The Gid13 Y Offset

Current visible rect and guide showpos disagree on Y:

- Guide: `y=504`
- Debug rect: `top=544.5`
- Click center used: `y=517`

Required:

- Add a short note to docs explaining coordinate systems:
  - viewport coordinate
  - `.stage` top offset
  - Flash-stage Y coordinate
  - `getBoundingClientRect()` Y
- If the button is visually lower than the guide hole, fix CSS.
- If the script is mixing viewport/stage coordinates, fix script.

Acceptance:

- The coordinate test logs both viewport and stage-relative rects.
- `gid13` showpos center and claim button stage-relative rect overlap.

## P1 Tasks After P0 Is Clean

Only start these after P0 scripts pass cleanly.

### P1.1 Login Screen 1:1 Evidence Pass

Goal: first screen must match Flash evidence, not just be usable.

Do:

- Read FFDec login classes under `artifacts\ffdec\BloodWar\scripts\Login`.
- Identify exact login background, input frames, button assets, and coordinates.
- Update HTML/CSS only where evidence supports it.
- Preserve working login behavior.
- Save current screenshot at `artifacts\html5-login-round2.png`.
- Create `docs\login-1to1-evidence-2026-05-03.md`.

Acceptance:

- Login with existing account works.
- Empty input behavior is documented or matched.
- Evidence doc lists Flash files, asset names, coordinates, screenshot path, known gaps.

### P1.2 Create Role / Province Selection 1:1 Pass

Goal: fresh account creation must be a real Flash-derived screen.

Do:

- Re-read `artifacts\ffdec\BloodWar\scripts\Login\CreateRoleDialog.as`.
- Align province map container to `x=333,y=63,w=618,h=500`.
- Use `images/sanguo_worldmap.png` and province overlay images.
- Verify portrait rule `images/player/player_{sex}_{faceIndex}.jpg`.
- Verify start button `x=67.5,y=449.3,w=115,h=42`.
- Confirm city-name behavior from AS before adding/removing inputs.
- Save screenshot at `artifacts\html5-create-role-round2.png`.
- Create `docs\create-role-1to1-evidence-2026-05-03.md`.

Acceptance:

- Fresh debug account reaches create-role.
- Province selection via map works.
- Create role enters city.
- Evidence doc lists what is exact, approximate, and blocked.

### P1.3 City Shell And Guide-Hole Evidence Pass

Goal: the current city screen should stop being judged by "it works" only.

Do:

- Read:
  - `artifacts\ffdec\BloodWar\scripts\CityInnerPanel.as`
  - `artifacts\ffdec\BloodWar\scripts\Bar\TopPanel.as`
  - `artifacts\ffdec\BloodWar\scripts\Bar\BottomPanel.as`
  - `artifacts\ffdec\BloodWar\scripts\Building\BuildingGrid.as`
- Save screenshot at `artifacts\html5-city-round2.png`.
- Create `docs\city-shell-1to1-evidence-2026-05-03.md`.
- List current guide holes `gid6-gid14`, target UI element, and whether the visible control overlaps the hole.

Acceptance:

- The doc includes a table:
  - `gid`
  - `showpos`
  - expected Flash control
  - current HTML element/class
  - stage-relative rect
  - pass/fail
- No fake "aligned" claims without rect evidence.

## Do Not Work On Yet

- World map.
- Battle.
- Mail/report/rank/shop/union/friend windows.
- Broad CSS cleanup.
- Responsive layout.

Those come after login/create-role/city shell are evidence-locked.

## Required Final Report Format

Claude must report exactly:

```text
Build/test:
- npm run build:
- go test ./...:
- verify-speed-flow.ps1:
- verify-speed-flow-coordinates.ps1:

P0 fixes:
- Backend task state proof:
- Pure coordinate regression:
- gid13 coordinate explanation:

P1 progress:
- Login:
- Create role:
- City shell:

Evidence files:
- ...

Screenshots:
- ...

Known gaps:
- ...

Files changed:
- ...
```

No "done" unless the command passes in a fresh run.
