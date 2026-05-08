# Flash SWF Readout First Pass

Source files:

- `D:\APMServ5.2.6\www\htdocs\BloodWar.swf`
- `D:\APMServ5.2.6\www\htdocs\common2.swf`

Generated probes:

- `D:\APMServ5.2.6\new_project\artifacts\swf-probe\BloodWar.json`
- `D:\APMServ5.2.6\new_project\artifacts\swf-probe\BloodWar.md`
- `D:\APMServ5.2.6\new_project\artifacts\swf-probe\common2.json`
- `D:\APMServ5.2.6\new_project\artifacts\swf-probe\common2.md`

## Confirmed SWF Classes And Assets

Login:

- `Login.LoginDialog`
- `Login.LoginDialog__embed_mxml_images_board_login_jpg_357357848` -> `board_login.jpg`
- `Login.LoginDialog__embed_mxml_images_board_login2_jpg_1760462718` -> `board_login2.jpg`
- `Login.LoginDialog__embed_mxml_images_sequence_jpg_1184055864` -> `sequence.jpg`
- `Login.LoginDialog__embed_mxml_images_PTELOGO_png_136645992` -> `PTELOGO.png`
- `Login.LoginDialog__embed_css_images_button_red_*` -> red login button states
- `Login.LoginDialog__embed_css_images_button_selected_png_1348947203`
- `Login.LoginDialog__embed_css_images_button_unselected_png_1136992691`

Create role:

- `Login.CreateRoleDialog`
- `Login.CreateRoleDialog__embed_mxml_images_title_png_1437309784` -> shared `title.png`
- `_Login_CreateRoleDialogWatcherSetupUtil`

City shell:

- `Bar.TopPanel`
- `Bar.BottomPanel`
- `UserInfo.InfoPanel`
- `UserInfo.InfoPanel__embed_mxml_images_leftboard_new_jpg_1497586600` -> `leftboard_new.jpg`
- `Bar.TopPanel__embed_mxml_images_board_topbar_png_1752022900` -> `board_topbar.png`
- `Bar.BottomPanel__embed_mxml_images_chatbar_png_731973978` -> `chatbar.png`
- `Bar.BottomPanel__embed_css_images_function_*` -> bottom function buttons

City/building:

- `Building.BuildingGrid`
- `Building.BuildingGrid__embed_mxml____swf_scene_swf_empty_shadow_539463457`
- `Building.CreateBuildingDialog`
- `Building.BuildingListDialog`
- `Building.BuildingTip`
- `Building.CreateBuildingItem`
- `CityInnerPanel_cityhall`
- `CityInnerPanel_house`
- `CityInnerPanel_barn`
- `CityInnerPanel_market`
- `CityInnerPanel_barrack`
- `CityInnerPanel_warhouse`
- `CityInnerPanel_blacksmith`
- `CityInnerPanel_stable`
- `CityInnerPanel_inn`
- `CityInnerPanel_hotel`
- `CityInnerPanel_institute`
- `CityInnerPanel_temple`
- `CityInnerPanel_recruithall`
- `CityInnerPanel_forceyard`
- `CityInnerPanel_tower`
- `CityInnerPanel_townwall`
- `CityInnerPanel_citywall`
- `BloodWar__embed_css_images_mycity_building_*`
- `BloodWar__embed_css_images_mycity_field_*`
- `BloodWar__embed_css_images_mycity_labor_*`

Guide:

- `guide.GuideTip`
- `guide.GuideTip__embed_css_images_tipBg_png_1589532066` -> `tipBg.png`
- `_guide_GuideTipWatcherSetupUtil`

## Implementation Changes From This Readout

- Frontend building image mapping now uses `CityInnerPanel_*` SWF symbols instead of temporary `inbuilding_*` assets.
- City empty/selected plot feedback now uses `mycity_building.png` from SWF symbol inventory.
- Guide panel now includes `tipBg.png`, tied to `guide.GuideTip`.
- City interior background now uses `cityinnerbg.png` (738x557) instead of the lower-fidelity temporary `map_innercity.jpg`.
- The SWF probe script is checked in at `scripts/swf_probe.py` so future work can repeat symbol extraction.

## ABC Class/Method Findings

Target class traits are exported to:

- `D:\APMServ5.2.6\new_project\artifacts\swf-probe\target-classes.md`

Create role:

- `Login.CreateRoleDialog` extends `mx.containers.Canvas`.
- Fields include `nameInput`, `codeInput`, `provinceSelect`, `radiogroup1`, `framePictureCanvas`, `mPlayerFaceImage`, `mFace`, `mSex`, `mProvinceInfo`, `mProvinceList`.
- Province images are real component fields: `province_1` through `province_13`.
- Button/event mapping from traits:
  - `___CreateRoleDialog_Button1_click` -> `onPrevFace`
  - `___CreateRoleDialog_Button2_click` -> `onNextFace`
  - `___CreateRoleDialog_Button3_click` -> `onSubmit`
  - `___CreateRoleDialog_Canvas1_creationComplete` -> `onCreate`
  - `__provinceSelect_change` -> province change handler
  - `__framePictureCanvas_mouseMove` and `__framePictureCanvas_click` handle province hover/click on the map canvas
- `onSubmit` references command strings and services: `createRole`, `createCity`, `sendGlobalCommand`, `openService`, `Command`, `user`, `新城池`.
- `makeFaceImage` builds paths from `images/player/player_`, `_`, `.jpg`.

Login:

- `Login.LoginDialog` extends `mx.containers.Canvas`.
- Fields include `passportInput`, `passwordInput`, `enterGameButton`, `savePassword`, `saveUsername`, `announcementText`, `loginMask`, queue fields.
- `onCreate` references service/config calls: `initLoginService`, `getLoginAnnouncement`, `getSavedUsername`, `getSavedPassword`, `doLogin`, `getPassType`, `getHomeButton`, `getHomeUrl`, `getRegisterUrl`, `getUserRule`.

Building:

- `Building.BuildingGrid` extends `mx.containers.Canvas`.
- Fields include `buildingBitmap`, `buildingImage`, `emptyImage`, `levelImage`, `mBid`, `mHilight`, `buildingTitle`.
- Methods include `setOuter`, `setCityHall`, `setCityWall`, `setWall`, `showLevel`, `setGrey`, `resetFilter`, `onCreate`.
- `Building.CreateBuildingDialog` extends `mx.containers.Canvas`.
- Methods include `loadValidBuildings`, `startUpgradeBuilding`, `onData`, `onCreate`, `onShow`, `onClose`.
- `Building.CreateBuildingItem` has `setItemInfo`, `showCreateBuildingTip`, `startUpgradeBuilding`, `__btnCreate_click`, `btnCreate`, `mBuildingID`, `mBuildingName`, `mDescription`, `mLevelDescription`.

Guide:

- `guide.GuideTip` extends `mx.containers.Canvas`.
- Confirmed asset dependency: `tipBg.png`.

## Still Missing For True 1:1

- Exact MXML component layout coordinates for `LoginDialog`, `CreateRoleDialog`, `BuildingGrid`, and `CreateBuildingDialog`.
- Button hitArea definitions and ActionScript event handlers.
- Exact city grid placement transform from `BuildingGrid` to stage coordinates.
- Exact create-role province hit zones and selection feedback.
- Exact guide sequence order and lock rules.

The current probe can read SymbolClass, ExportAssets, DoABC strings, ABC multinames, traits, method bodies, and best-effort bytecode references. It still does not reconstruct Flex `UIComponentDescriptor` constructor trees into exact x/y/width/height. A deeper pass should decode `newclass/newfunction` closures and `UIComponentDescriptor` property arrays or use FFDec/JPEXS for visual MXML reconstruction.
