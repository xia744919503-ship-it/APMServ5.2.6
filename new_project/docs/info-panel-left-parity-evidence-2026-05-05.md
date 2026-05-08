# InfoPanel Left Parity Evidence - 2026-05-05

Scope: city screen left `InfoPanel` only.

Flash evidence:
- `artifacts/ffdec/BloodWar/scripts/UserInfo/InfoPanel.as`
- `artifacts/ffdec/BloodWar/scripts/UserInfo/InfoTip.as`
- `artifacts/ffdec/BloodWar/binaryData/386_utils.LocaleChinese.bin`

Verified / fixed:
- Rechecked the Flash root left-board bitmap. `InfoPanel.as` creates the first root child as `Image` with `x=1,y=0,source=_embed_mxml_images_leftboard_new_jpg_1497586600`; the exported asset `leftboard_new.jpg` is `263x600`, and the city template already renders exactly one `<img class="left-board">` at `left=1,top=0,width=263,height=600`. A duplicate `.left-data::before` background was avoided because it would double-darken the panel.
- City ComboBox matches Flash `x=47,y=8,w=171,h=18,rowCount=10`; fixed parent clipping so all 10 rows can open.
- City ComboBox dropdown colors now match Flash `.InfoComboBoxDropdown`: background `#001e32`, border `#147273`, selection `#002f3e`, rollover `#199ebf`, selected text `#e1dc5a`; evidence `test-results/info-panel-city-dropdown-after-style.png`.
- City title image now uses the Flash asset's native displayed size `252x35` at `x=0,y=3` inside the `left=7,y=203` header canvas; evidence `test-results/info-panel-current-resource-after-title-size.png`.
- No-union position row now matches Flash `Utils.getUnionPos()`: `User.unionid <= 0` returns an empty string, so `职位:` has no value while `联盟:` remains `无联盟`.
- Left root text shadow removed; Flash `InfoPanel` labels use explicit colors/font sizes without inherited text shadow.
- Resource tab coordinates verified against Flash:
  - ViewStack absolute origin: left panel `6,344`.
  - Resource icons: internal `x=11`, y `8/38/68/98`, size `35x25`.
  - Resource icon child images match Flash local offsets: food/wood/iron `x=5,y=-1,w=30,h=27`, rock `x=5,y=0,w=30,h=27`.
  - Value boxes: internal `x=49,w=102`; add boxes `x=155,w=75`.
  - Resource value labels match Flash `y=3,right=2,textAlign=right`; resource add labels match Flash `y=3,right=2,textAlign=right` except food add `y=2`.
  - Plus buttons: internal `x=232`, y `16/47/77/106`.
  - Divider: internal `x=6,y=126`.
  - ActivityBar: internal left/right/bottom `6`, height `113`, alpha `.5`.
  - Divider asset `line_leftboard.png` verified at native `240x3`; Flash sets only `x=6,y=126`, so browser native image sizing is correct.
- Four tab states render without broken images; current clean evidence:
  - `test-results/info-panel-current-resource-after-title-size.png`
  - `test-results/info-panel-current-commander-after-title-size.png`
  - `test-results/info-panel-current-army-after-title-size.png`
  - `test-results/info-panel-current-defence-after-title-size.png`
- `InfoHeroItem`, `InfoSoldierItem`, and `InfoDefenceItem` internal dimensions/coords match Flash descriptors:
  - Hero item `122x40`, frame `x=4,y=0,w=33,h=40`, portrait `x=2,y=2,w=29,h=36`, text board `x=40,w=80`.
  - Soldier/defence item `122x39`, icon board `x=4,y=0,w=47,h=38`, text board `x=53,y=0,w=66,h=37`.
  - Soldier icon `x=3,y=2,w=40,h=32`; defence icon `x=3,y=3,w=40,h=32`.
  - Soldier/defence name label `x=0,y=0,color=#cec36d,bold`; count label `x=0,y=14,w=57,textAlign=right`.
  - Placement matches Flash: hero `x=8/125,y=floor(i/2)*47+8`; army/defence `x=6/124,y=floor(i/2)*40+8`.
- Soldier/defence list names now use fixed Flash slot names for the 12 army slots and 5 defence slots while using backend data only for counts. This prevents fixture/backend labels such as `Worker`, `Militia`, or other aliases from replacing the Flash-localized left-panel names.
- Left tab switching now follows Flash's event timing: `btn_resource`, `btn_commmander`, `btn_army`, and `btn_defence` are wired from `mouseDown` to `onChangeViewStack(index)` in `InfoPanel.as`, so the HTML buttons now switch on `mousedown` instead of waiting for `click`; keyboard Enter/Space uses the same transition for browser accessibility.
- Top lord metadata click scope rechecked against `InfoPanel.as`: Flash gives click/hand-cursor to the field labels `君主/声望/排名/官职/爵位/联盟/职位` and additionally to the rank value label only. Current frontend keeps values non-clickable except the rank value, matching that descriptor instead of broadening hit zones.
- Removed obsolete `.left-bottom-tab` / `.left-bottom-tabs` CSS from an earlier left-tab implementation; the live template uses only `.left-view-tabs .left-vert-btn`, which maps to the Flash `btn_resource/btn_commmander/btn_army/btn_defence` descriptors.
- Rechecked the top lord metadata rows and the `King/Armor/Inventory` buttons against `InfoPanel.as`: rows stay at `x=84,y=12/39/66/93/120/147/174,w=161,h=24`; action buttons stay at `x=13,y=102/135/168` with `King`, `Armor`, and `Inventory` skins.
- Runtime DOM check confirms the ViewStack and tab geometry after browser layout:
  - Tab hit boxes render at root-relative `x=12/72/132/192,y=321,w=60,h=26`, matching Flash panel `left=6,top=312` plus button `x=6/66/126/186,y=9`.
  - Resource stack renders at `x=6,y=344,w=252,h=250`; first food icon board at `x=17,y=352,w=35,h=25`; first resource value box at `x=55,y=352,w=102,h=25`; resource plus buttons at `x=238,y=360/391/421/450,w=11,h=11`.
  - Commander stack renders at `x=6,y=344,w=252,h=250`; visible hero items at `x=14/131,y=352` and `x=14,y=399`, matching Flash local `8/125` and `floor(i/2)*47+8`.
  - Army stack renders 12 items and defence stack renders 5 items, matching Flash `initSoldier()` and `initDefence()` slot counts.
- Left tab SVG assets are FFDec shape exports (`ffdec:objectType="shape"`) and not hand-drawn replacements.
- Combo dropdown evidence:
  - `test-results/info-panel-city-dropdown.png`
- InfoTip matches Flash dimensions and fixed position `x=45,y=39,w=168,h=42+16*n`; labels now use the Flash locale text.
- InfoTip separator rows now match Flash `addHorizoneLine()`: name column gets locale `UserInfo_InfoTip_1_1` (`────────────`), value column gets a blank row, and the separator contributes one 16px line to tooltip height.
  - Morale evidence: `test-results/info-tip-morale-with-separator.png`, CDP text includes `当前民怨`, separator, then `民心变化`; rect `x=45,y=39,w=168,h=106`.
  - Gold evidence: `test-results/info-tip-gold-with-separator.png`, CDP text includes `将领俸禄`, separator, then `实际收入`; rect `x=45,y=39,w=168,h=138`.
  - People evidence: `test-results/info-tip-people-with-separator.png`, CDP text includes separators after `人口上限` and `空闲人口`; rect `x=45,y=39,w=168,h=170`.
- Food InfoTip now matches Flash `addTipItem(..., City.food_army_use, City.food_army_use > 0, "-")`: the army food-use row always renders the `-` prefix, including `-0` when usage is zero; evidence `test-results/info-tip-food-army-use-prefix.png`, rect `x=45,y=39,w=168,h=122`.
- Replaced browser-native `title` attributes on left InfoPanel values/buttons with a controlled Flex-like tooltip:
  - Flash style source: `_ToolTipStyle.as`, `fontSize=9`, `backgroundColor=16777164`, `backgroundAlpha=0.95`, `borderColor=9542041`.
  - Locale text source: `UserInfo_InfoPanel_9_1` through `UserInfo_InfoPanel_34_1`.
  - Resource/morale/gold/people icons still use custom `InfoTip`; numeric labels and plus/mycity buttons use the plain tooltip.
  - Evidence: `test-results/flash-tooltip-gold-value.png` and `test-results/info-tip-gold-icon.png`.
- Matched Flash's 1px exception for the food add label (`foodAdd y=2`, other resource add labels `y=3`).
- Matched Flash top summary gold formatting: `Math.round(City.gold)` for the summary strip, while resource tab values continue to use `Math.floor(*_real)`.
- Matched Flash skin-native size for the three `MyCity*` buttons: `24x19` at `x=218,y=31/52/73` instead of stretching them to `30x18`; evidence `test-results/info-panel-current-resource-after-mycity-size.png`.
- Matched Flash `Button` hand-cursor behavior for the three `MyCity*` buttons by explicitly setting the browser cursor to `pointer`.
- Restored Flash avatar click target: portrait `x=2,y=2,w=64,h=80` inside the `68x84` frame now receives clicks and routes to the same current `showUserInfoDialog` stub as the `君主` button; verified via CDP hit-test on `.lord-portrait`.
- Matched Flash hand-cursor state for the rank value label (`_InfoPanel_Label6`, `x=50,y=2,w=105,useHandCursor=true,buttonMode=true` inside the rank row); CDP verified `.lord-meta-row.row-rank strong` computes `cursor: pointer`.
- ActivityBar now follows Flash `scrollContent()` display mechanics: a rotating window of at most six items, starting at `mCurrentIndex`, wrapping modulo list length, and using each last visible item's `interval` as the next delay. Its visible TextArea geometry remains `x=4,w=360,top=0,bottom=-5` inside the `left/right/bottom=6,height=113,alpha=.5` panel; evidence `test-results/info-panel-current-after-activitybar-rotation.png`.
- ActivityBar data now follows Flash `refreshContent()` command source. Flash sends `sendGlobalCommand("activity","getActivityList")`; legacy PHP reads `select content, link, interval from sys_activity where inuse=1 order by id`. The Go adapter exposes the same payload at `GET /api/legacy/activities`, and the frontend loads it when entering the city screen while preserving the placeholder only if the adapter is unavailable. Smoke evidence: `STATUS=200` returned current legacy rows `欢迎使用我爱热三v1.2版本` and `更新了占领的战报以及道具获得`, both with `interval="60"`.
- Commander tab data contract now exposes Flash's original `InfoHeroItem` state field `statename` in addition to the modern `stateLabel`; the frontend reads `statename` first. Flash evidence: `InfoHeroItem.as` binds `mItem.statename`, `mItem.sex`, and `mItem.face` directly. Current evidence: `test-results/info-panel-commander-after-statename-contract.png`; CDP sample returned first hero `sex=1,face=3,statename=城守,stateLabel=城守`, and rendered image `/assets/images/herox/hero_boy_3.jpg` with natural size `29x36`.
- Latest left screenshot:
  - `test-results/info-panel-current-after-activitybar-rotation.png`
  - `test-results/info-panel-commander-after-statename-contract.png`

Left InfoPanel visible interaction inventory, Flash evidence from `InfoPanel.as`:
- Portrait image `___InfoPanel_Image2_click`, `君主` label `___InfoPanel_Label1_click`, and `King` button `___InfoPanel_Button1_click` all call `onClickUserInfoBtn()` -> `Global.dialogManager.showUserInfoDialog()`. Current frontend routes all three to the documented `君主信息` stub.
- `Armor` button `___InfoPanel_Button2_click` calls `Global.dialogManager.showArmorDialog()`. Current frontend routes to the documented `装备` stub.
- `Inventory` button `___InfoPanel_Button3_click` calls `Global.dialogManager.showMyGoodsDialog()`. Current frontend routes to the documented `宝物` stub.
- Prestige label `_InfoPanel_Label3` and rank label `_InfoPanel_Label5` call `showRankDialog()`. Rank value `_InfoPanel_Label6` is also hand-cursor/clickable in Flash and current frontend now binds rank value click to `openRankPanel()`.
- Office label `_InfoPanel_Label7` and nobility label `_InfoPanel_Label9` call `showJiaGuan()`. Current frontend opens the existing task/office panel; this is not yet verified as the Flash `showJiaGuan()` window.
- Union label `_InfoPanel_Label11` and union-position label `_InfoPanel_Label13` call `showUnionDialog()`. Current frontend opens the existing union panel; this is not yet verified as the Flash `showUnionDialog()` window.
- Morale icon and plus call `showUseGoodsDialog(0,4,null)`, gold icon and plus call `showUseGoodsDialog(0,5,null)`, people icon and plus call `showUseGoodsDialog(0,6,null)`. Current frontend routes to the documented use-goods stub with the same `gid`.
- Tax icon `___InfoPanel_Image4_click` calls `gotoTax()`, which creates/reuses `ChangeTaxDialog`. Current frontend opens the government tax subdialog; visual/window parity is still unverified.
- MyCity buttons call `showBuildingListDialog()`, `showResourceProductDialog()`, and `showCityFieldDialog()`. Current frontend opens existing government/manage subdialogs; visual/window parity is still unverified.
- Resource icons and plus buttons call `showUseGoodsDialog(0,7+param,null)`: food `7`, wood `8`, rock `9`, iron `10`. Current frontend routes to the documented use-goods stub with those `gid` values.

Plain tooltip locale inventory, Flash evidence from `386_utils.LocaleChinese.bin` and `InfoPanel.as` bindings:
- Summary strip: `9_1=民心`, `10_1=民怨`, `11_1/16_1=税率`, `12_1=黄金数量`, `13_1=当前人口`, `14_1=黄金产量`, `15_1=空闲人口`, `17_1=使用宝物增加黄金税收`, `18_1=使用宝物提高民心，消除民怨`, `19_1=使用宝物增加人口`, `20_1=建筑信息`, `21_1=资源生产`, `22_1=附属野地`.
- Resource tab: `23_1=粮食数量`, `24_1=石料数量`, `25_1=木材数量`, `26_1=铁锭数量`, `27_1=铁锭产量`, `28_1=木材产量`, `29_1=石料产量`, `30_1=粮食产量`, `31_1=使用宝物增加粮食产量`, `32_1=使用宝物增加木材产量`, `33_1=使用宝物增加石料产量`, `34_1=使用宝物增加铁锭产量`.

Known non-complete areas:
- Browser screenshot verification was not refreshed in this pass: `npm run build` passes, but local `vite dev/preview` fails in this sandbox while loading `vite.config.ts` because esbuild child process spawn returns `EPERM`; an escalated start attempt also did not expose the port.
- User info / armor / goods dialogs opened from the top-left commander buttons are still status stubs unless implemented elsewhere.
- Real-time resource tick parity is not complete: Flash updates `*_real` every second from `*_add_real`, while the current frontend displays backend snapshot values and refreshes elsewhere. This is a data-contract/runtime behavior gap, not a coordinate/UI asset gap.
- Do not claim full city 1:1 parity from this file; this covers only the left InfoPanel pass.
