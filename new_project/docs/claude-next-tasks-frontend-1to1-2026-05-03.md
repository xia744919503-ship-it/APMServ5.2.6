# Claude Next Tasks: BloodWar Frontend 1:1 Continuation

Date: 2026-05-03
Workspace: `D:\APMServ5.2.6\new_project`

## Current Verified Status

The current new-player chain `gid6 -> gid14` is functionally passing:

```powershell
cd D:\APMServ5.2.6\new_project
npm --prefix frontend run build
go test .\backend\...
powershell -ExecutionPolicy Bypass -File D:\APMServ5.2.6\new_project\artifacts\verify-speed-flow.ps1
```

Observed pass:

- Select empty plot.
- Build `民房`.
- Click busy `民房`.
- Open speed goods dialog.
- Use `鲁班残页`.
- Open task panel.
- Select `建造民房`.
- Claim reward.
- Reach final guide text `gid14`.

Do not overclaim: this proves the current function chain, not full 1:1 frontend parity.

## Hard Rules

- Flash/SWF/DB evidence is the source of truth.
- Screenshots are validation, not invention material.
- Keep the original `1000x600` Flash stage coordinate model.
- Use original bitmap assets. Do not introduce generic icons, modernized panels, fake maps, or placeholder battle UI.
- Do not mark a flow complete unless both click-path regression and evidence notes exist.
- If evidence is missing, write `BLOCKED: missing evidence`, not `DONE`.
- Do not refactor unrelated CSS or backend code.

## Immediate P0: Strengthen The Regression Gate

The existing `verify-speed-flow.ps1` passes, but `gid11-gid13` currently uses DOM `.click()` for task button, task row, and claim button. That is not strict enough for 1:1 guide validation.

### Task P0.1: Add Real Coordinate Click Regression

Update or add a sibling script:

```text
artifacts/verify-speed-flow-coordinates.ps1
```

It must drive the same flow using stage coordinate clicks only.

Required guide holes:

- `gid11`: task button `showpos=912,4,68,24`; click center approximately `(946,16)`.
- `gid12`: completed task row `showpos=312,392,145,24`; click center approximately `(384,404)`.
- `gid13`: claim reward button `showpos=795,504,77,26`; click center approximately `(833,517)`.

Acceptance:

- The script starts from a fresh debug passport.
- It reaches final `gid14`.
- It writes JSON output with hit target class/tag, guide text before/after each coordinate click, and error text.
- It exits non-zero if any coordinate click does not advance the guide.
- Keep `verify-speed-flow.ps1` as the functional smoke test; add the coordinate script as the stricter 1:1 gate.

### Task P0.2: Verify Backend Task State

The latest run still printed:

```text
[WARN] Could not verify task status via API
```

Fix this. The script should verify the same debug passport/session it used, not some unrelated browser state.

Acceptance:

- After claiming task 1, API/DB verification proves task 1 is no longer claimable or has claimed state.
- If auth/session prevents direct API check, query DB by the script's `debugPassport`.
- The verification must fail the script if task state is not proven.

Useful SQL:

```sql
SELECT u.uid,u.passport,t.id,t.name,ut.state
FROM sys_user u
JOIN sys_user_task ut ON ut.uid=u.uid
JOIN cfg_task t ON t.id=ut.tid
WHERE u.passport='<debugPassport>'
ORDER BY t.id;
```

## P1: Login Screen 1:1 Evidence Pass

Goal: Make login the first honest 1:1 screen, not just a usable shell.

### Evidence To Read

- FFDec exported login scripts under:
  - `artifacts\ffdec\BloodWar\scripts\Login\*.as`
  - `artifacts\ffdec\BloodWar\scripts\BloodWar.as`
- Existing screenshots:
  - `artifacts\html5-login-pass2.png`
  - `artifacts\html5-login-first-playable.png`
- Original assets in:
  - `frontend\public\assets`
  - `www\htdocs\images`

### Required Work

- Identify exact login background, logo, input frame, button states, and error style from evidence.
- Align account/password input coordinates to Flash.
- Replace CSS-looking controls with original bitmap-backed controls where the original has assets.
- Keep the existing successful `legacyLogin` behavior.
- Capture current HTML5 screenshot at `1000x600` after changes.

Acceptance:

- Existing account login enters city or create-role correctly.
- Empty username/password error behavior is original-like or documented as evidence-missing.
- Save screenshot evidence under `artifacts/`.
- Add a short evidence note under `docs/` with asset names and coordinates.

## P1: Create Role / Province Selection 1:1 Pass

Goal: Fresh account creation must match Flash enough to be playable and recognizable.

### Evidence Already Known

Use `artifacts\ffdec\BloodWar\scripts\Login\CreateRoleDialog.as`.

Important confirmed facts:

- Dialog stage: `1000x600`.
- Province map container: `x=333,y=63,w=618,h=500`.
- Base map: `images/sanguo_worldmap.png`.
- Province overlay images include:
  - `silv.png`
  - `jizhou.png`
  - `yuzhou.png`
  - `yunzhou.png`
  - `xuzhou.png`
  - `qingzhou.png`
  - `jingzhou.png`
  - `yangzhou.png`
  - `yizhou.png`
  - `liangzhou.png`
  - `bingzhou.png`
  - `youzhou.png`
  - `jiaozhou.png`
- Portrait rule: `images/player/player_{sex}_{faceIndex}.jpg`.
- Start button position from AS: `x=67.5,y=449.3,w=115,h=42`.

### Required Work

- Rebuild the create-role screen from `CreateRoleDialog.as`, not from memory.
- Fix province overlays and hit zones.
- Implement portrait previous/next and gender toggle using original behavior.
- Confirm whether city name is a real input or legacy constant in the original. If still ambiguous, document it before changing behavior.
- Verify create role endpoint payload matches the old client order as closely as the backend supports.

Acceptance:

- Fresh debug account reaches create-role.
- User can select province by map and/or dropdown.
- User can create role and enter city.
- Save screenshot evidence at `1000x600`.
- Add a click-path note in docs.

## P1: City Inner Screen 1:1 Alignment

Goal: Stop city UI drift. The city should look and click like Flash for the first playable screen.

### Evidence To Use

- `artifacts\ffdec\BloodWar\scripts\CityInnerPanel.as`
- `artifacts\ffdec\BloodWar\scripts\CityPanel.as`
- `artifacts\ffdec\BloodWar\scripts\Building\BuildingGrid.as`
- `artifacts\ffdec\BloodWar\scripts\Bar\TopPanel.as`
- `artifacts\ffdec\BloodWar\scripts\Bar\BottomPanel.as`

Known facts:

- Inner city panel: `736x556`.
- Inner background low/high at `(3,3)`:
  - `images/map_innercity_low.jpg`
  - `images/map_innercity_high.jpg`
- Top panel button positions in Flash `TopPanel`:
  - inner `(10,4)`
  - outer `(80,4)`
  - world `(150,4)`
  - battle `(220,4)`
  - build `(305,4)`
  - level `(407,4)`
  - hero `(436,4)`
  - army `(507,4)`
  - union `(577,4)`
  - mission `(647,4)`
- Building grid base: `108x79`.
- Empty shadow: `(2,20)` size `104x50`.

### Required Work

- Reconcile current city CSS/assets with the Flash `736x556` panel inside the `1000x600` stage.
- Audit current top tabs. If they are not Flash top panel positions/assets, fix or document the compatibility reason.
- Ensure every guide hole from gid6-gid14 lands on a visible clickable control.
- Confirm empty plot, house build, busy house click, speed button, task button, task row, claim button with coordinate regression.
- Do not move UI just to satisfy one guide hole unless the move is supported by Flash evidence or DB guide coordinates.

Acceptance:

- Functional smoke test passes.
- Coordinate click test passes.
- Current screenshot at `1000x600` saved.
- Evidence doc lists remaining visual deltas.

## P1: Build Panel / Building Info Panel 1:1 Pass

Goal: Building dialogs should use original geometry and behavior.

### Evidence To Use

- `artifacts\ffdec\BloodWar\scripts\Building\CreateBuildingDialog.as`
- `artifacts\ffdec\BloodWar\scripts\Building\CreateBuildingItem.as`
- `artifacts\ffdec\BloodWar\scripts\Building\UpgradeBuildingDialog.as`
- `artifacts\ffdec\BloodWar\scripts\Building\UpgradeBuildingItem.as`
- `artifacts\ffdec\BloodWar\scripts\DialogManager.as`

### Required Work

- Verify create-building dialog dimensions and item row coordinates.
- Verify `民房` row image, title, resource requirements, build button, disabled state.
- Verify occupied building info panel title, image, level, state text, speed button, close button.
- Unsupported buttons must either work or show original-like blocked behavior.

Acceptance:

- Empty tile opens build panel.
- `民房` build works.
- Busy `民房` opens info panel.
- Speed button opens goods dialog.
- Screenshots and click-path notes saved.

## P1: UseGoodsDialog Generalization

Goal: The current speed goods flow works for `鲁班残页`, but the dialog should stop being a single-guide hack.

### Required Work

- Verify dialog root and item list coordinates against `UseGoodsDialog.as`.
- Keep gid10 click hole valid.
- Implement accurate backend behavior for:
  - `67`: 15 minutes.
  - `68`: 1 hour.
  - `69`: 2.5 hours.
  - `70`: 8 hours.
  - `71`: random 10-30 hours if original confirms.
  - `72`: 30% remaining time if original confirms.
  - `73`: instant behavior/yuanbao cost only if evidence exists.
- Non-owned goods must be disabled or blocked like original.

Acceptance:

- `鲁班残页` path still passes.
- Other listed goods do not lie about behavior.
- Backend mutation is correct and documented.

## Do Not Start Yet Unless P0/P1 Above Are Green

- World map 1:1.
- Battle 1:1.
- Mail/report/rank/shop/union/friend windows.
- Broad visual polish.
- Responsive redesign.

These are important, but they must wait until login/create-role/current city/tutorial chain are honestly locked down.

## Required Final Report From Claude

When done, report in this exact structure:

```text
Build/test:
- npm run build: pass/fail
- go test ./...: pass/fail
- verify-speed-flow.ps1: pass/fail
- verify-speed-flow-coordinates.ps1: pass/fail

Completed:
- ...

Evidence:
- Flash/AS paths read:
- Screenshots saved:
- JSON outputs saved:

Known gaps:
- ...

Files changed:
- ...
```

No vague "basically done". No "1:1 complete" unless the coordinate regression, screenshots, and evidence notes all pass.
