# SWF FFDec Extraction

Source SWFs:

- `D:\APMServ5.2.6\www\htdocs\BloodWar.swf`
- `D:\APMServ5.2.6\www\htdocs\common2.swf`

Tools:

- FFDec/JPEXS 26.0.0 installed at `C:\Program Files (x86)\FFDec\ffdec.exe`
- Portable Java at `D:\APMServ5.2.6\new_project\tools\jre\jdk-17.0.18+8-jre\bin\java.exe`
- RABCDAsm 1.18 at `D:\APMServ5.2.6\new_project\tools\rabcdasm`

FFDec was run through portable Java:

```powershell
& 'D:\APMServ5.2.6\new_project\tools\jre\jdk-17.0.18+8-jre\bin\java.exe' -jar 'C:\Program Files (x86)\FFDec\ffdec.jar' -export script,image,shape,movie,sound,binaryData,text D:\APMServ5.2.6\new_project\artifacts\ffdec\BloodWar D:\APMServ5.2.6\www\htdocs\BloodWar.swf
& 'D:\APMServ5.2.6\new_project\tools\jre\jdk-17.0.18+8-jre\bin\java.exe' -jar 'C:\Program Files (x86)\FFDec\ffdec.jar' -export script,image,shape,movie,sound,binaryData,text D:\APMServ5.2.6\new_project\artifacts\ffdec\common2 D:\APMServ5.2.6\www\htdocs\common2.swf
```

Generated:

- `D:\APMServ5.2.6\new_project\artifacts\ffdec\BloodWar\scripts`
- `D:\APMServ5.2.6\new_project\artifacts\ffdec\BloodWar\images`
- `D:\APMServ5.2.6\new_project\artifacts\ffdec\BloodWar.xml`
- `D:\APMServ5.2.6\new_project\artifacts\ffdec\BloodWar.dumpSWF.txt`
- `D:\APMServ5.2.6\new_project\artifacts\ffdec\BloodWar.dumpAS3.txt`
- `D:\APMServ5.2.6\new_project\artifacts\ffdec\common2\scripts`
- `D:\APMServ5.2.6\new_project\artifacts\ffdec\common2\images`
- `D:\APMServ5.2.6\new_project\artifacts\ffdec\common2.xml`

## Stage

SWF display rect from FFDec XML:

- twips: `Xmax=20000`, `Ymax=12000`
- pixels: `1000 x 600`

## LoginDialog

Source: `artifacts\ffdec\BloodWar\scripts\Login\LoginDialog.as`

- Root: `Canvas 1000 x 600`
- Background images:
  - `board_login.jpg`: `x=0 y=0 w=1000 h=600`
  - `board_login2.jpg`: `x=0 y=0 w=1000 h=600`
- Username checkbox: `x=403 y=504 w=97`, style `NormalCheckBox`
- Password checkbox: `x=508 y=504 w=94`, style `NormalCheckBox`
- Login/enter button: `x=453 y=540 w=92.75 h=39`, style `NormalButton`
- Register button: `x=35 y=489 w=92.75 h=39`
- Account input wrapper: `x=444 y=439 w=140.5 h=22`, child `passportInput x=2 y=2 w=136 h=20`, style `EmptyTextInput`
- Password input wrapper: `x=444 y=469 w=140.5 h=22`, child `passwordInput x=2 y=2 w=136 h=20`, style `EmptyTextInput`
- Announcement text area: `x=42 y=200 w=235 h=255`, style `EmptyTextInput`
- PTE logo: `x=854 y=533`

## CreateRoleDialog

Source: `artifacts\ffdec\BloodWar\scripts\Login\CreateRoleDialog.as`

- Root: `Canvas 1000 x 600`
- Title image: `x=374 y=12`
- Title label: `x=418.5 y=16 w=151 h=18`
- World-map canvas `framePictureCanvas`: `x=333 y=63 w=618 h=500`, style `FramePicture`
- Map image inside frame: `x=5 y=5 w=608 h=490`, source `images/sanguo_worldmap.png`
- Left form panel: `x=50 y=63 w=262 h=500`
- Face frame: `x=14.5 y=52 w=233 h=160`
- Face image `selectHead`: `x=10 y=10 w=66 h=82` inside the face frame
- Province combo: `x=104 y=259 w=144`
- Name input: outer `x=103 y=223 w=144 h=26`, child `nameInput x=3 y=2 w=140 h=23`
- Code input: outer `x=104 y=367 w=144 h=26`, child `codeInput x=2 y=2 w=124`
- User agreement radio/button row: `x=27.5 y=417.3 w=213 h=24`
- Start button: `x=67.5 y=449.3 w=115 h=42`

Province hover overlays are hidden by default. Hit-testing loops `province_1` to `province_13`, checks `hitTestPoint`, then reads the PNG pixel under the mouse and treats nonzero color as inside the province. On mouse move, only the current province image is visible. On click, `provinceSelect.selectedIndex` becomes the province index.

Province overlays inside `framePictureCanvas`:

| Province id | Source | x | y | w | h |
| --- | --- | ---: | ---: | ---: | ---: |
| 1 | `images/silv.png` | 208.5 | 159 | 179 | 93 |
| 2 | `images/jizhou.png` | 361 | 90 | 102 | 105 |
| 3 | `images/yuzhou.png` | 331 | 190 | 123 | 101 |
| 4 | `images/yunzhou.png` | 374 | 155 | 96 | 92 |
| 5 | `images/xuzhou.png` | 421 | 183 | 118 | 96 |
| 6 | `images/qingzhou.png` | 428 | 121 | 119 | 81 |
| 7 | `images/jingzhou.png` | 254 | 235 | 152 | 172 |
| 8 | `images/yangzhou.png` | 373 | 239 | 204 | 175 |
| 9 | `images/yizhou.png` | 40 | 220 | 270 | 204 |
| 10 | `images/liangzhou.png` | 0 | 7 | 260 | 244 |
| 11 | `images/bingzhou.png` | 219 | 45 | 153 | 155 |
| 12 | `images/youzhou.png` | 360 | 4 | 253 | 118 |
| 13 | `images/jiaozhou.png` | 135 | 363 | 328 | 136 |

## BuildingGrid

Source: `artifacts\ffdec\BloodWar\scripts\Building\BuildingGrid.as`

- Root grid: `w=108 h=79`
- Empty image: `x=2 y=20 w=104 h=50`, source embedded `empty_shadow`
- Building image: `x=0 y=0 w=103 h=79`
- Bitmap: `x=0 y=0 w=103 h=79`
- Level image: `x=40 y=40 w=19 h=14`
- `setOuter()`: root `112x66`, empty `x=6 y=4 w=104 h=50`, building `112x66`, level `x=46 y=30`
- `setCityHall()`: source `images/building_cityhall.png`, building `195x105`, level `x=150 y=36`
- `setWall(true)`: source `images/building_citywall.png`, building `731x519`, level `x=660 y=165`
- `setWall(false)`: source `images/building_townwall.png`, building `731x519`, level `x=562 y=138`
- `setCityWall(true/false)`: sources `images/building_outercity.png` / `images/building_outertown.png`, building `356x222`

## CreateBuildingDialog

Source: `artifacts\ffdec\BloodWar\scripts\Building\CreateBuildingDialog.as`

- Root: `Canvas 682 x 492`, style `PanelGreyGreen`
- Close button: `x=603 y=445`, style `Close`
- Item container: `y=44 h=386`, style `BoardListDown`, left/right style `20`, vertical scroll on
- Title image: `y=10`, horizontally centered, source `title.png`
- Title label: `y=15`, horizontally centered
- Items are `CreateBuildingItem`, positioned `x=10`, `y=10 + index * 75`
- Start build command sends `city/startUpgradeBuilding` with `[City.cid, City.current_inner, City.current_x, City.current_y, bid]`

## CreateBuildingItem

Source: `artifacts\ffdec\BloodWar\scripts\Building\CreateBuildingItem.as`

- Root: `Canvas 605 x 70`, style `BoardListUp`
- Create button `btnCreate`: `x=84 y=32 w=57 h=30`, style `ButtonRed`
- Text description: `y=10 w=444 h=48`, right style `10`
- Name label: `x=89 y=9 w=57`, style `ItemTitle`
- Building image: `x=8 y=6 w=72 h=56`, source binding `images/building_intro_{bid}.png`
- `setItemInfo` binds `bid`, `name`, `description`, `levelDescription`, and `btnCreate.enabled = canUpgrade`
