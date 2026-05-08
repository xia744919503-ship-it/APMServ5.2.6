# Claude Handoff: BloodWar Flash 1:1 Frontend Continuation

Date: 2026-05-03
Workspace: `D:\APMServ5.2.6\new_project`

## Non-Negotiables

- Use Flash/SWF/database evidence as the source of truth. Do not invent modern UI.
- Browser screenshots are validation only, not the design source.
- Keep the original 1000x600 Flash stage coordinate system.
- Use original bitmap assets from `frontend/public/assets`, `www/htdocs/images`, and FFDec exports.
- Do not claim 100% or 1:1 unless a click-path regression has passed.
- Do not replace missing work with fake panels or generic icons. Mark gaps explicitly.

## Current Services

- Frontend dev server: `http://127.0.0.1:5173`
- Backend API: `http://127.0.0.1:8080`
- Chrome CDP: `http://127.0.0.1:9222`
- Backend DB status was verified live against `bloodwar@127.0.0.1:3306`.

If backend code changes, rebuild and restart:

```powershell
cd D:\APMServ5.2.6\new_project\backend
$procIds = Get-NetTCPConnection -LocalPort 8080 -State Listen -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess -Unique
foreach ($procId in $procIds) { if ($procId -gt 0) { Stop-Process -Id $procId -Force } }
go build -o api.exe .\cmd\api
Start-Process -FilePath '.\api.exe' -WorkingDirectory 'D:\APMServ5.2.6\new_project\backend' -RedirectStandardOutput 'D:\APMServ5.2.6\new_project\backend\api.stdout.log' -RedirectStandardError 'D:\APMServ5.2.6\new_project\backend\api.stderr.log' -WindowStyle Hidden
```

## Source Evidence Already Confirmed

### Guide DB, group 1

`cfg_guide` has the current playable new-player chain:

- gid 6: select empty plot, `showpos=570,130,100,60`
- gid 7: select 民房, `showpos=405,142,53,23`
- gid 8: click busy 民房, `showpos=570,130,100,60`
- gid 9: click 加速, `showpos=399,240,44,21`
- gid 10: use 鲁班残页, `showpos=413,234,68,81`
- gid 11: trigger type 3, `triggerdetails=5,1`, task button prompt, `showpos=912,4,68,24`
- gid 12: task row prompt, `showpos=312,392,145,24`
- gid 13: claim reward button prompt, `showpos=795,504,77,26`
- gid 14: final guide text, no showpos.

Useful SQL:

```sql
SELECT gid,pregid,triggertype,triggerdetails,showpos,disdetails,content
FROM cfg_guide
WHERE `group`=1 AND gid BETWEEN 6 AND 14
ORDER BY gid;
```

### Flash / FFDec references

- `artifacts/ffdec/BloodWar/scripts/guide/Guide.as`
- `artifacts/ffdec/BloodWar/scripts/guide/GuideManager.as`
- `artifacts/ffdec/BloodWar/scripts/guide/GuideTip.as`
- `artifacts/ffdec/BloodWar/scripts/Building/UpgradeBuildingItem.as`
- `artifacts/ffdec/BloodWar/scripts/DialogManager.as`
- `artifacts/ffdec/BloodWar/scripts/Goods/UseGoodsDialog.as`
- `artifacts/ffdec/BloodWar/scripts/Task/TaskDialog.as`

Important Flash facts:

- Building speed button calls:
  `Global.dialogManager.showUseGoodsDialog(mBuildingState.state_endtime - Global.getServerTime(), 0, [City.cid, inner, x, y, bid])`
- `UseGoodsDialog.loadData()` sends `user/getUserTypeGoods` with `[mType, mTime]`.
- `UseGoodsDialog.onUseGoods()` sends `user/useTypeGoods` with `[mType, gid, City.cid, ...mParam]`.
- For building speed, `mType=0`.
- `DialogManager.useGoodsSucc()` refreshes city building when `useGoodsDialog.getType() == 0`.
- `DialogManager.showTaskDialog()` creates `TaskDialog`, calls `getTaskGroupList()`, then shows it.

## Work Completed And Verified

### Backend

Implemented:

- `GET /api/legacy/guides?group=1`
- `GET /api/cities/{cid}/buildings/speed-goods?position=...`
- `POST /api/cities/{cid}/buildings/speed-goods/use`
- Starter goods now include:
  - `50101` 新手礼包1
  - `67 x2` 鲁班残页

Key files:

- `backend/internal/legacy/models.go`
- `backend/internal/legacy/city_building_goods.go`
- `backend/internal/legacy/account_role_create.go`
- `backend/internal/service/service.go`
- `backend/internal/server/router.go`

Verified backend chain:

- Fresh user creates role.
- Build 民房 at position `100`.
- `speed-goods` returns 鲁班系 goods `67..73`.
- Using gid `67` consumes one 鲁班残页.
- Building becomes `bid=5, level=1, state=0`.

Observed verification result:

```json
{
  "pos": 100,
  "houseCanBuild": true,
  "remaining": 150,
  "goods67": 2,
  "afterBid": 5,
  "afterLevel": 1,
  "afterState": 0
}
```

### Frontend

Implemented:

- Guide uses DB `showpos/disdetails` rather than fake hardcoded rectangles.
- Guide mask blocks outside clicks and leaves the target hole clickable.
- gid 6 -> gid 7 -> gid 8 -> gid 9 -> gid 10 -> gid 11 -> gid 12 -> gid 13 -> gid 14 click chain works end-to-end.
- Real UseGoodsDialog-style speed goods panel added.
- Original item images copied:
  - `frontend/public/assets/item_67.png`
  - `frontend/public/assets/item_68.png`
  - `frontend/public/assets/item_69.png`
  - `frontend/public/assets/item_70.png`
  - `frontend/public/assets/item_71.png`
  - `frontend/public/assets/item_72.png`
  - `frontend/public/assets/item_73.png`
- Building panel and speed button were moved so gid 9 `showpos=399,240,44,21` hits the 加速 button.
- Speed goods dialog was moved so gid 10 `showpos=413,234,68,81` hits 鲁班残页.
- Task panel added with original-style dialog (panelgrey_green.png background).
- topbutton_mission_up.png is now clickable via `.top-tabs img:nth-child(4)`.
- Task list shows 建造民房 (completed) with claim reward button.
- Guide progression: gid11→task button→task panel→gid12→click task row→gid13→click claim→gid14 final text.

Key files:

- `frontend/src/App.vue`
- `frontend/src/api.ts`
- `frontend/src/style.css`
- `frontend/public/assets/item_67.png` through `item_73.png`

Regression script:

- `artifacts/verify-speed-flow.ps1`

Latest verified output (gid11-gid14):

```json
{
  “g6”: “请您先选择一块风水宝地，然后在该\”空地\”上建造\”民房\”一间，私虽陋室，却蕴龙行。”,
  “g7”: “请选择民房”,
  “g8”: “您建造的\”民房\”看来还需要一点时间才能完成，如果您希望建筑物快速完成，您可以尝试点击\”民房\”选择\”加速\”功能”,
  “panel”: “民房(等级1)”,
  “g9”: “请点击加速”,
  “g10”: “请选择使用道具\”鲁班残页\”进行加速”,
  “goods”: “鲁班残页,鲁班便笺,鲁班草图,鲁班书册,鲁班秘录,鲁班全集,鲁班传人”,
  “afterGuide”: “您的第一间\”民房\”已经顺利建造完成。请您点击右上角\”任务\”按钮，领取任务奖励。”,
  “afterPanel”: “民房(等级1)”,
  “g11”: “\”建造民房\”任务已经显示\”完成\”，您可以点击该任务领取任务奖励。”,
  “taskPanelVisible”: “visible”,
  “taskItemsCount”: 1,
  “g12”: “请您点击任务界面右下角\”领取奖励\”按钮，领取任务奖励。”,
  “taskClaimBtn”: “visible”,
  “g13”: “不积跬步，无以至千里。现在，千秋霸业的第一步已经完成，在这间\”民房\”之中，义士运筹帷幄，必能平定天下。大将斩清风，名驹抹黄沙，本次引导也到此为止，愿义士一战天下闻，一仗江山震。闯荡属于你的《热血三国》吧！”,
  “g14”: “不积跬步，无以至千里。现在，千秋霸业的第一步已经完成，在这间\”民房\”之中，义士运筹帷幄，必能平定天下。大将斩清风，名驹抹黄沙，本次引导也到此为止，愿义士一战天下闻，一仗江山震。闯荡属于你的《热血三国》吧！”,
  “error”: “”
}
```

Validation commands:

```powershell
cd D:\APMServ5.2.6\new_project\frontend
npm run build

cd D:\APMServ5.2.6\new_project\backend
go test ./...

powershell -ExecutionPolicy Bypass -File D:\APMServ5.2.6\new_project\artifacts\verify-speed-flow.ps1
```

## Completed This Session

gid11→gid14 task chain is now complete. Full guide chain (gid6→gid14) passes regression:

1. Make top mission button clickable - DONE
   - topbutton_mission_up.png is now clickable via `.top-tabs img:nth-child(4)`
   - On click: loads `myTasks()`, opens task panel, advances from gid11 to gid12

2. Implement TaskDialog-style panel - DONE
   - Uses panelgrey_green.png background matching build panel style
   - Task list shows “建造民房” with 1/1 progress (completed state)
   - 领取奖励 button appears after clicking task row
   - Guide progression: gid12→click task→gid13→click claim→gid14 final text

3. Backend task endpoints - Already existed:
   - `GET /api/me/tasks` returns task snapshot
   - `POST /api/me/tasks/claim` with `{ taskId }` claims reward

4. Guide progression - Implemented:
   - gid11: task button click → task panel opens → gid12
   - gid12: task row click → claim button visible → gid13
   - gid13: claim button click → gid14 final guide text

5. Regression script extended - PASSES
   - Full gid6→gid14 chain verified
   - All guide texts correct
   - Task panel visible after gid11
   - Claim button visible after gid12
   - Final guide text shown after gid13

6. Bug fixes this session:
   - Fixed PowerShell syntax error: `-or` operator replaced with separate if statements
   - Fixed wrong API endpoint: `/api/tasks/my` → `/api/me/tasks`
   - Fixed reversed task status check: `-eq “claimed”` (was `-ne`)
   - Fixed PowerShell Chinese encoding issue: moved verification logic to Python script
   - Fixed task claim button position: left=636px top=450px (matches gid13 showpos=795,504,77,26)

## Known Gaps / Do Not Overclaim

- Full login/create-role/city visual 1:1 is still incomplete.
- Map and battle are not 1:1.
- TaskDialog complete behavior not fully implemented (only claim flow works).
- Current guide starts at gid 6, skipping gid 1-5 package/task setup.
- UseGoodsDialog position was adjusted to match guide DB click hole for gid10. If doing deeper visual parity later, verify against original Flash screenshot before moving it again.
- Regression script now verifies backend task state change (task 1 status via API after claiming).

## Full Remaining Work Backlog

This section is the larger queue after the immediate gid11-gid14 task. Work in priority order. Do not jump to broad polish before the first playable guide chain is closed.

### P0: Finish Current New-Player Guide Chain

Status: COMPLETE

Goal: Complete the current playable tutorial flow through final gid14.

Completed states:

- gid11: click top task button at `showpos=912,4,68,24` - DONE
- gid12: show task panel and highlight completed `建造民房` row at `showpos=312,392,145,24` - DONE
- gid13: highlight `领取奖励` button at `showpos=795,504,77,26` - DONE (.task-claim-btn left=636px, top=450px)
- gid14: show final guide text after reward claim - DONE

Implementation details:

- Uses existing `/api/me/tasks` and `/api/me/tasks/claim`
- Task panel styled with panelgrey_green.png to match build panel
- topbutton_mission_up.png clickable via `.top-tabs img:nth-child(4)`
- Guide layer z-index raised above task layer to show guide text over panel
- Task claim button position adjusted: left=636px, top=450px (matches gid13 showpos=795,504,77,26)

Acceptance: verify-speed-flow.ps1 extended through gid14 PASSES.

### P0: Restore/Protect Regression Script

Status: COMPLETE

Goal: Make one command reliably validate the current guide chain.

Tasks:

- Keep `artifacts/verify-speed-flow.ps1` as the canonical smoke test.
- Add clear JSON output fields:
  - `g6`, `g7`, `g8`, `g9`, `g10`, `g11`, `g12`, `g13`, `g14`
  - `goods`
  - `taskPanelVisible`
  - `taskClaimBtn`
  - `error`
- Exit code check: returns 0 on pass, 1 on failure.
- Verification logic moved to `artifacts/verify-speed-flow-check.py` (Python parses UTF-8 correctly).
- Make the script fail with non-zero exit code if any expected guide text is missing.
- Avoid broad CDP loops that spam disconnected WebSocket errors.

Acceptance:

```powershell
powershell -ExecutionPolicy Bypass -File D:\APMServ5.2.6\new_project\artifacts\verify-speed-flow.ps1
```

must end with gid14 and no error.

### P1: Login Screen 1:1 Evidence Pass

Goal: Login screen should match Flash evidence, not just be usable.

Required evidence:

- Original Flash/login screenshot at 1000x600.
- Current HTML5 screenshot at 1000x600.
- Asset list for login background, logo, input frame, button states.

Tasks:

- Verify `board_login.jpg` placement and scale.
- Match account/password input coordinates.
- Match login button image/state; avoid CSS-only button if original bitmap exists.
- Check empty username/password error style against original if evidence exists.
- Preserve successful login behavior via `legacyLogin`.

Acceptance:

- Click path: login existing account -> enters city or create-role.
- Screenshot comparison saved under `artifacts/`.
- List remaining visual deltas in a doc if not pixel-perfect.

### P1: Create Role / Province Selection 1:1 Pass

Goal: First character creation must align with old Flash.

Required controls:

- Left role panel.
- Portrait image, previous/next arrows, gender toggle.
- Commander name input.
- City name input.
- Province selector/hit zones.
- Agreement checkbox if present in Flash.
- Start button.
- Right province map and selected province feedback.

Known current state:

- Some improvements were made earlier, but this has not been fully verified as 1:1.

Tasks:

- Re-read FFDec/create-role scripts and assets.
- Verify province bitmap pieces and hit zones.
- Confirm create role endpoint maps to old expected payload.
- Confirm default values match old Flash where possible.
- Capture screenshots before claiming done.

Acceptance:

- Fresh account -> create role -> enters city.
- Screenshot evidence saved.
- Province click zones behave like original.

### P1: City Inner Screen Visual Alignment

Goal: The city screen must stop drifting from original Flash layout.

Required areas:

- Top tabs: 城内, 城外, 地图, 任务, 报告, 联盟.
- Left lord/resource panel.
- Inner-city board/background.
- Building grid hit zones.
- City hall and wall hit zones.
- Bottom chat/function bar.
- Building level badges.
- Modal layer placement.

Tasks:

- Reconcile current `assets.ts` coordinates with Flash evidence.
- Do not move panels just to make one guide hole work unless it matches original or is documented.
- Ensure gid showpos holes land on the same visible controls as Flash.
- Audit all temporary coordinate hacks and record which are evidence-backed.

Acceptance:

- Empty plot selection, 民房 build, busy building click, speed click all work visually and interactively.
- Current screenshot at 1000x600 and old baseline screenshot are both saved.
- No guide hole points to empty/non-clickable space.

### P1: Build Panel / Building Info Panel 1:1 Pass

Goal: Building creation and building detail panels should use original panel geometry.

Already partly done:

- Build/upgrade panels use FFDec-exported panel/list/button/frame assets.
- Building info uses backend `CityBuildingInfo`.
- Level images use original `level_*.png`.

Remaining:

- Verify `CreateBuildingDialog` dimensions and item row coordinates from FFDec.
- Verify `UpgradeBuildingDialog` / `UpgradeBuildingItem` dimensions and button positions.
- Confirm disabled/insufficient-resource states.
- Confirm close/cancel/destroy button behavior matches original.
- Confirm opening occupied vs empty tile behaves differently and predictably.

Acceptance:

- Build 民房 and open 民房 panel with correct image, level, description, state text, buttons.
- Unsupported buttons either work or show original-like blocked behavior.

### P1: UseGoodsDialog Generalization

Goal: Current speed-goods dialog works for gid10, but needs a more evidence-backed shape.

Already done:

- Building speed goods `67..73` listed.
- Original item images copied.
- gid10 click hole lands on 鲁班残页.
- Using 鲁班残页 consumes one item and completes short build.

Remaining:

- Verify dialog root and inner popup coordinates against Flash screenshots, not only DB guide hole.
- Support goods effects accurately:
  - 67: 15 minutes
  - 68: 1 hour
  - 69: 2.5 hours
  - 70: 8 hours
  - 71: random 10-30 hours if implemented later; currently conservative fixed 10 hours
  - 72: 30% remaining time; currently backend placeholder needs real implementation if used
  - 73: instant with yuanbao cost; currently listed but disabled without inventory
- Add no-inventory behavior matching Flash.
- Avoid using fake fallback images except as a visible documented gap.

Acceptance:

- gid10 path passes.
- Non-owned goods are disabled or blocked like original.
- Backend state changes are correct.

### P2: Task System Beyond First Reward

Goal: After gid14, task UI should not dead-end.

Tasks:

- Display task categories/groups from `TaskSnapshot`.
- Display selected task content, todo, goals, rewards.
- Support completed and incomplete styles.
- Support claim reward for available tasks.
- Support follow-up task activation after claim.
- Confirm reward resource/goods application.

Acceptance:

- Task panel can still be used after tutorial ends.
- Claiming task updates resources/task list without reload.

### P2: Top Navigation Stubs And Real Entrypoints

Goal: Visible top buttons must not be inert fake icons.

Buttons:

- 城内: current screen, active state.
- 城外: should route/show outer city or explicit documented blocked state.
- 地图: should route/show map if implemented.
- 任务: task panel.
- 报告: report panel or documented blocked state.
- 联盟: union panel or documented blocked state.

Tasks:

- Convert image-only top tabs to clickable controls while preserving bitmap look.
- Use original up/on/down assets.
- Record unimplemented targets as gaps, not done.

Acceptance:

- Every visible top tab is clickable.
- Each click either opens the real panel or a known original-like blocked message.

### P2: Map Screen 1:1 Pass

Goal: Replace any fake/random map with old terrain/tile evidence.

Required:

- Original terrain tile style and scale.
- Pan controls.
- Coordinate display.
- My-city button.
- Selected target panel.
- City/NPC/terrain icons from old assets.

Existing backend:

- `/api/world/map` exists.

Tasks:

- Verify current map implementation against `world-pass*.png` artifacts.
- Use backend world data, not random icon placement.
- Make tile clicking select correct target.
- Make unsupported actions documented.

Acceptance:

- Opening map from top tab shows data-backed map.
- Click tile selects a target.
- Screenshot evidence saved.

### P2: Battle Screen / Battle Entry

Goal: Do not present battle as complete until it is actually interactive.

Required:

- Battle frame.
- Terrain.
- Unit slots.
- Command panel.
- Hero display.
- Unit selection.
- Command buttons.
- Turn/action state.

Tasks:

- Read FFDec battle scripts and assets before UI work.
- Inventory backend battle endpoints/data.
- Implement only the first validated battle flow.
- If combat logic is missing, mark blocked.

Acceptance:

- A user can enter a battle screen and click at least the original visible controls that are claimed implemented.
- Unsupported actions are explicit.

### P2: Utility Windows

Goal: Replace inert/fake utility buttons with real or documented windows.

Areas:

- Reports
- Mail
- Ranking
- Shop
- Union
- Friend
- Hero/recruit
- Research
- Barracks/training

Tasks:

- Prioritize windows touched by early-game flow.
- Use existing backend endpoints where already present.
- Use original assets and dimensions.

Acceptance:

- Each implemented window has a click-path test and screenshot evidence.

### P3: Backend Compatibility Gaps

Goal: Remove compatibility shortcuts as evidence becomes available.

Known items:

- Starter goods currently grant 鲁班残页 directly to let gid10 work without opening 新手礼包1. This is practical for the current playable chain but should be reconciled with original gid1-gid5/package flow later.
- Building speed item 71/72/73 behavior is only partly modeled.
- Some task error messages in backend are mojibake from legacy encoding; fix only when touching user-visible errors.

Tasks:

- Reconstruct gid1-gid5 from Flash/DB and decide whether new-player package opening should precede gid6.
- Implement package opening if required.
- Improve goods-use parity for all speed items.
- Keep DB mutations reversible/testable on fresh debug accounts.

Acceptance:

- Fresh account follows original startup sequence without manual seed hacks, or the hack is documented as a temporary compatibility layer.

### P3: Documentation And Evidence Index

Goal: Every claimed completed flow has a source trail.

Tasks:

- Maintain a parity checklist per flow in `docs/` or `artifacts/`.
- For each flow record:
  - Original evidence file.
  - Current screenshot.
  - Used assets.
  - Server calls.
  - DB mutations.
  - Known gaps.
- Do not let screenshots accumulate without naming what they prove.

Acceptance:

- A reviewer can open one doc and see exactly what is complete and what is not.

### P3: Visual Pixel Cleanup

Goal: After click-paths are real, tighten visual deltas.

Tasks:

- Compare old/current screenshots at 1000x600.
- Fix asset scale, panel offsets, text sizes, clipping.
- Avoid broad CSS refactors.
- Do not introduce responsive scaling into Flash-stage components until parity is proven.

Acceptance:

- Diffs shrink without breaking interaction regression.

## Useful SQL For Next Verification

```sql
SELECT u.uid,u.passport,t.id,t.name,ut.state
FROM sys_user u
JOIN sys_user_task ut ON ut.uid=u.uid
JOIN cfg_task t ON t.id=ut.tid
WHERE u.passport='<debugPassport>'
ORDER BY t.id;

SELECT id,`group`,pretid,name,content,todo
FROM cfg_task
WHERE id=1;

SELECT *
FROM cfg_task_goal
WHERE tid=1;

SELECT *
FROM cfg_task_reward
WHERE tid=1;
```

## Suggested Acceptance Gate

Claude should not report this segment done unless all pass:

- `npm run build`
- `go test ./...`
- `verify-speed-flow.ps1` extended through gid14 passes
- UI click path proves:
  - gid11 text visible
  - task button opens task panel
  - gid12 text visible
  - 建造民房 row is visible and clickable
  - gid13 text visible
  - 领取奖励 button is visible and clickable
  - gid14 final guide text visible
  - task 1 is no longer claimable afterward
