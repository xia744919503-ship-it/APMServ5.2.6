# Phase 3 澶嶅埢杩借釜琛?
鏇存柊鏃堕棿锛?026-05-03

## 妯″潡鐘舵€?
- [x] 鐧诲綍/鎺掗槦/寤鸿閾捐矾闂幆锛堝惈鏃犲煄姹犺处鍙蜂細璇濋敊璇慨澶嶏級
- [x] 鍩庢睜缁忔祹缁撶畻 1:1 瀵归綈
  - 璇佹嵁: `artifacts/economy-formula-verification-20260503-014134.json` (allPassed=true, People Max Formula PASS)
- [x] 建筑队列 1:1 对齐
  - 证据: `artifacts/building-queue-20260503-023451.json` (allPassed=true, 覆盖create/cancel/recreate/complete/upgrade/cancel_upgrade/final_state)
- [x] 科技队列 1:1 对齐
  - 证据: `artifacts/tech-queue-20260503-025933.json` (allPassed=true, 覆盖前置条件/资源扣除/忙队列拒绝/取消结算)
- [x] 寰佸叺闃熷垪 1:1 瀵归綈
  - 璇佹嵁: `artifacts/recruit-queue-20260503-022827.json` (allPassed=true, 瑕嗙洊鍒涘缓/鎵ｉ櫎/瀹屾垚/闈炴硶杈撳叆鎷掔粷)
- [x] 鍑哄緛涓庢垬鏂?1:1 瀵归綈
  - 证据: `artifacts/battle-flow-test-20260503-020613.json` (allPassed=true, dispatch=200, troopId>0, scout report PASS)
- [x] 鎴樻姤瀛楁涓庤涔?1:1 瀵归綈
  - 璇佹嵁: 鍚屼笂 artifact (report_structure PASS, all fields: id/type/read/title/createdAt)
- [x] 任务系统 1:1 对齐
  - 证据: `artifacts/quest-system-20260503-023839.json` (allPassed=true, 当前服务语义下领奖暂不支持，覆盖列表/未完成拒绝/领奖语义/重复拒绝)
- [x] 鑱旂洘绯荤粺 1:1 瀵归綈
  - 证据: `artifacts/union-system-20260503-023142.json` (allPassed=true, 覆盖创建、加入申请/取消、成员读取、权限校验、退出/解散基础流程)
- [x] 鍟嗗煄/鍏呭€煎厬鎹?1:1 瀵归綈
  - 璇佹嵁: `artifacts/shop-payment-20260503-022901.json` (allPassed=true, 瑕嗙洊鍟嗗搧鍒楄〃/璐拱鎵ｆ/鍏呭€煎埌璐?浣欓涓嶈冻鎷掔粷)
- [x] 邮件系统 1:1 对齐
  - 证据: `artifacts/mail-system-20260503-023012.json` (allPassed=true, 覆盖system/inbox/outbox读取与异常语义)
- [x] 排行系统 1:1 对齐
  - 证据: `artifacts/rank-system-20260503-023013.json` (allPassed=true, 覆盖25类kind与分页语义)
- [x] 主界面/地图/功能窗关键交互 1:1 对齐
  - 证据: `artifacts/main-map-ui-20260503-023012.json` (allPassed=true, 覆盖overview/city/world-map/city-list/reports)

## 褰撳墠闃绘柇椤?
- 灏氭湭寤虹珛鈥滃悓涓€杈撳叆 -> 鏃х増涓庢柊鐗堢粨鏋滃揩鐓у姣斺€濈殑缁熶竴鑴氭湰銆?- 閮ㄥ垎鐜╂硶浠呭仛浜嗗彲杩愯瀹炵幇锛屾湭瀹屾垚寮洪獙鏀躲€?
## 涓嬩竴姝ワ紙姝ｅ湪鎵ц锛?
1. 寤虹珛榛勯噾鏍锋湰鑴氭湰锛堢櫥褰?-> 寤虹瓚/绉戞妧/寰佸叺 -> 鍑哄緛 -> 鎴樻姤锛夈€?2. 鍏堝畬鎴愨€滃煄姹犵粡娴?+ 寤虹瓚/绉戞妧/寰佸叺鈥濇敹鍙ｏ紝骞惰緭鍑哄樊寮傛姤鍛娿€?
## 鏈疆鏂板锛?026-05-02锛?
- 鏂板鑴氭湰锛歚scripts/smoke-legacy-core-state-machines.ps1`
- 宸茶鐩栧苟閫氳繃锛?  - 绋庣巼/浜ч噺鍥炲啓 roundtrip
  - 寤虹瓚鈥滄柊寤?-> 鍙栨秷鈥濈姸鎬佹満
  - 寤虹瓚鈥滃崌绾?-> 鍙栨秷鈥濈姸鎬佹満
  - 绉戠爺鈥滃紑濮?-> 鍙栨秷鈥濈姸鎬佹満
  - 寰佸叺鈥滃紑濮?-> 鍙栨秷鈥濈姸鎬佹満
  - 鍑哄緛鈥渄ispatch -> callback鈥濈姸鎬佹満
- 鏈疆淇锛?  - `sys_troops` 鏂板缓璁板綍琛ュ啓 `startcid`锛屼慨澶嶅嚭寰佽褰曡捣鐐逛负 `0` 鐨勮涔夊亸宸€?- 璇存槑锛氳繖浠ｈ〃鈥滃洖褰掑熀绾垮缓绔嬪畬鎴愨€濓紝**涓嶄唬琛?*鐩稿叧妯″潡宸茶揪鎴?1:1 鏈€缁堝榻愩€?
## 鏈疆鏂板锛?026-05-02 绗簩娆℃敹鍙ｏ級

- 鏂板鑴氭湰锛歚scripts/smoke-legacy-peripheral-modules.ps1`
- 宸茶鐩栧苟閫氳繃锛?  - 浠诲姟璇诲彇 + 鏈揪鎴愪换鍔￠濂栧け璐ヨ涔?  - 鑱旂洘璇诲彇 + 鏃犲垱寤烘潈闄愭椂鍒涘缓澶辫触璇箟
  - 鍟嗗煄璇诲彇 + 闈炴硶璐拱鍙傛暟澶辫触璇箟
  - 鍏呭€艰鍙?+ 闈炴硶鍏戞崲鍙傛暟澶辫触璇箟
  - 閭欢鏀朵欢绠?绯荤粺淇?鍙戜欢绠辫鍙?  - 鎴樻姤鍒嗛〉璇诲彇锛坲nread/type0锛?  - 鎺掕姒?25 绫?kind 璺敱涓庤繑鍥?kind 涓€鑷存€?- 鏈疆淇锛?  - 鎺掕绯荤粺琛ラ綈 legacy kind锛歚union`銆佸啗鍔?鎹愮尞/鍕ょ帇/璐″搧锛堝惈鑱旂洘姒滐級銆佸叺鍔涙銆佽储瀵屾銆佸姛鍕嬫銆?  - 淇 `kind=union` 琚敊璇洖閫€鍒?`kind=user` 鐨勮涔夊亸宸€?
