# Flash HTML5 Playable Flow Spec

Scope: first playable chain only: login -> create role when required -> city interior -> click empty inner plot -> create building.

Original stage size: 1000x600. Keep all UI in this coordinate system and center/scale the root stage only outside the game area.

## Evidence

- Login baseline: `D:\APMServ5.2.6\www\htdocs\images\board_login.jpg` (1000x600)
- Alternate login/entry: `D:\APMServ5.2.6\www\htdocs\images\board_login2.jpg`, `entergame.jpg`
- Create role province assets: `sanguo_map.png`, per-province PNG files such as `bingzhou.png`, `jingzhou.png`, `yizhou.png`
- City baseline: `D:\APMServ5.2.6\new_project\artifacts\legacy-city-clear.jpg`, `legacy-city-center-full.png`
- City reusable assets: `cityinnerbg.png`, `leftboard_new.jpg`, `chatbar.png`, `topbutton_*`, `CityInnerPanel_*`, `mycity_building.png`, `guide_mm.png`, `arrow_*.png`
- Chrome click evidence: `D:\APMServ5.2.6\new_project\artifacts\screenshots\chrome-city-click-build.png`

## State Machine

- `boot`: check `/api/auth/me`. If no active user, enter `login`.
- `login`: submit `/api/legacy/login` with `{ version: 0, loginType: 0, passType: "local", passport, password, auth: "" }`.
- `create-role`: shown when logged user has `cityCount <= 0`. Submit `/api/legacy/role/create`.
- `city`: load `/api/me/cities`, then `/api/cities/{cid}`. Render city shell and building hit zones.
- `build-options`: click an empty/inner plot, call `/api/cities/{cid}/buildings/options?position={position}`.
- `building`: submit `/api/cities/{cid}/buildings/create` with selected `position` and `bid`, then reload city detail.

## Implemented Controls For First Playable

- Login passport/password fields and enter button.
- Role form: ruler name, city name, flag character, sex, face toggle, province selection and start button.
- City: top navigation buttons are visual placeholders; inner city remains active.
- City building plots: known inner slots are represented as transparent hit zones over the old city background.
- Build chooser: lists backend options, blocks disabled options, creates the selected building.
- Newbie guide: first city entry modal blocks the background until dismissed.

## Latest Regression

- `npm run build` passes in `D:\APMServ5.2.6\new_project\frontend`.
- API smoke through Vite proxy passed: legacy login -> role create -> city load -> build options at inner position `100`.
- Chrome CDP test using `C:\Users\Administrator\AppData\Local\Google\Chrome\Application\chrome.exe` passed: click at stage coordinate `(355,190)` opened the `建造建筑` panel with 14 options.

## Known Gaps

- Exact create-role Flash coordinates are still not fully decoded from Flex `UIComponentDescriptor`; current page uses real `sanguo_map.png`, province PNGs, and player face assets, but it is not 1:1 certified.
- Login form exact text-field coordinates need screenshot comparison in plugin Chrome.
- City plot coordinates are approximate over `cityinnerbg.png` and must be refined against Flash click evidence.
- Build dialog is functional and uses old building/button assets, but exact `Building.CreateBuildingDialog/CreateBuildingItem` layout is not yet reconstructed.
- Top/bottom utility buttons outside the first playable chain are visible but not function-complete.
- This version is not declared 1:1; it is a playable parity slice awaiting screenshot/click-path regression.
