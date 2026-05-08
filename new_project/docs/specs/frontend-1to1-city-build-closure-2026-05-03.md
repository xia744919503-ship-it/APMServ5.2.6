# Frontend 1:1 Closure: Login/Role/City/Build

Date: 2026-05-03

## Scope

This note covers only the currently verified frontend chain:

- login entry
- create role / choose province
- enter inner city
- new-player guide target lock
- click empty inner plot
- open create-building dialog
- create building
- click busy occupied building
- open upgrade/building dialog

It does not certify world map, battle, or all utility windows.

## SWF Evidence Used

- Original SWF: `D:\APMServ5.2.6\www\htdocs\BloodWar.swf`
- FFDec scripts: `D:\APMServ5.2.6\new_project\artifacts\ffdec\BloodWar\scripts`
- FFDec images: `D:\APMServ5.2.6\new_project\artifacts\ffdec\BloodWar\images`
- Native stage: `1000x600`

Key ActionScript evidence:

- `Building/CreateBuildingDialog.as`
  - dialog `682x492`
  - item container `left/right=20`, `y=44`, `h=386`, `styleName=BoardListDown`
  - close button `x=603`, `y=445`, `styleName=Close`
- `Building/CreateBuildingItem.as`
  - item `605x70`, `styleName=BoardListUp`
  - image `x=8`, `y=6`, `72x56`
  - create button `x=84`, `y=32`, `57x30`, `styleName=ButtonRed`
  - title `x=89`, `y=9`, `w=57`
  - description `y=10`, `w=444`, `h=48`, `right=10`
- `Building/UpgradeBuildingDialog.as`
  - dialog `682x166`
  - title image `x=218`, `y=11`
  - title label `x=266`, `y=16`, `151x18`
  - close button `right=20`, `bottom=10`, `styleName=Close`
  - item `x=18`, `y=46`, `w=646`
- `Building/UpgradeBuildingItem.as`
  - item `646x73`, `styleName=BoardListUp`
  - image frame `x=8`, `y=7`, `76x60`
  - image `x=2`, `y=2`, `72x56`
  - upgrade/speed button `x=90`, `y=7`, `50x30`
  - destroy/cancel button `x=90`, `y=37`, `50x30`
  - info board `x=147`, `y=7`, `491x59`, `styleName=BoardListDown`
- `World/BuildingGrid.as`
  - building image `x=0`, `y=0`, `103x79`
  - level image `x=40`, `y=40`, `19x14`
  - cityhall level image `x=150`, `y=36`

## Assets Now Using Original SWF Exports

Direct FFDec exports already in frontend:

- `level_0.png` through `level_10.png`
- `CityInnerPanel_*` inner-city building images
- `building_intro_*.png`
- `button_red.png`, `button_red_hl.png`, `button_red_pr.png`
- `button_close.png`, `button_close_hl.png`, `button_close_pr.png`

Newly exported from original SWF symbol classes:

- `panelgrey_green.png`
  - Source class: `BloodWar__embed_css_swf_UIParts_swf_panelgrey_green_1436267812`
  - Character id: `25`
- `board_listdown.png`
  - Source class: `BloodWar__embed_css_swf_UIParts_swf_board_listdown_181322118`
  - Character id: `197`
- `board_listup.png`
  - Source class: `BloodWar__embed_css_swf_UIParts_swf_board_listup_1311296035`
  - Character id: `207`
- `button_green.png`, `button_green_hl.png`, `button_green_pr.png`
  - Source classes: `BloodWar__embed_css_images_button_green_*`
  - Source shape ids: `38`, `47`, `141`
- `button_yellow.png`, `button_yellow_hl.png`, `button_yellow_pr.png`
  - Source classes: `BloodWar__embed_css_images_button_yellow_*`
  - Source shape ids: `138`, `7`, `79`
- `frame_pic.png`
  - Source class: `BloodWar__embed_css_images_frame_pic_png_1043465936`
  - Source shape id: `22`

## Browser Validation

Validated against running services:

- frontend: `http://127.0.0.1:5173`
- backend: `http://127.0.0.1:8080`
- Chrome CDP: `http://127.0.0.1:9222`

Latest CDP smoke result:

- create-building dialog
  - `682x492`
  - list `642x386`, relative `20,44`
  - first item `605x70`, relative `10,10`
  - create button `57x30`, relative `84,32`
  - background image is `panelgrey_green.png`
  - list background image is `board_listdown.png`
  - item background image is `board_listup.png`
- created building
  - busy occupied plot appears
  - level image source `/assets/level_1.png`
  - level image `19x14`, relative `40,40`
- busy building dialog
  - title `民房(等级1)`
  - speed button exists
  - cancel button exists
  - dialog `682x166`
  - panel background image is `panelgrey_green.png`
  - item background image is `board_listup.png`
  - info board background image is `board_listdown.png`
- idle cityhall dialog
  - title `官府(等级1)`
  - upgrade button background image is `button_green.png`
  - destroy button background image is `button_red.png`
  - icon frame background image is `frame_pic.png`
- busy building dialog button/frame skin validation
  - speed button background image is `button_yellow.png`
  - cancel button background image is `button_red.png`
  - icon frame background image is `frame_pic.png`
- backend text validation
  - fixed building names through deterministic `bid -> Chinese name` mapping
  - cityhall title no longer displays mojibake
  - create-building option names begin with `民房`, `书院`, `校场`, `兵营`, `客栈`

## Verified Commands

- `npm run build` in `D:\APMServ5.2.6\new_project\frontend`
- `go test ./...` in `D:\APMServ5.2.6\new_project\backend`

Both passed after the latest frontend changes.

## Known Gaps

These are not complete and must not be reported as 1:1 finished:

- upgrade dialog content does not yet use original `getBuildingInfo` current/next-level details.
- speed-up flow is still stubbed/disabled, not the original goods dialog.
- map and battle frontend parity are not certified by this note.
- `docs/specs/phase3-1to1-tracker.md` contains encoding damage and should not be used as the sole truth without checking artifacts.

## Current Verdict

The inner-city build chain is now substantially closer to Flash evidence and click-complete for the tested path, with original SWF-derived panel/list/level assets loaded in-browser.

It is not a full-game 1:1 completion certificate.
