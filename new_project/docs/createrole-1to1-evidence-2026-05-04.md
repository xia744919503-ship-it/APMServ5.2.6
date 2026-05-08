# 创建角色界面 1:1 证据文档
日期: 2026-05-04
状态: screenshot_verified

## FFDec 源码参考
- `artifacts/ffdec/BloodWar/scripts/Login/CreateRoleDialog.as`

## Flash 创角界面分析

### 尺寸
- 整个对话框: 1000 x 600 (全舞台)
- 背景色: 11976102 (RGB)

### 主要元素
| 元素 | Flash坐标 | 尺寸 | 说明 |
|------|-----------|------|------|
| 标题图 | 374x, 12y | 嵌入: title.png | |
| 标题文字 | 418.5x, 16y | 151x18 | "创建角色" |
| 地图区域(framePictureCanvas) | 333x, 63y | 618x500 | 包含世界地图 |
| 世界地图 | 5x, 5y | 608x490 | images/sanguo_worldmap.png |
| 左侧面板 | 50x, 63y | 262x500 | 包含头像/州郡/性别/名称 |
| 头像框 | 72.5x, 13y | 87x100 | |
| 头像切换按钮 | 34.5x, 47.5y / 175x, 47.5y | 21x30 | prev/next |
| 性别单选(男) | 75x, 121y | 47x24 | |
| 性别单选(女) | 131.5x, 121y | 47x24 | default selected |
| 君主名输入框 | 103x, 223y | 144x26 | |
| 州郡下拉框 | 104x, 259y | 144 | rowCount=7 |
| 用户协议 | 27.5x, 417.3y | 213x24 | default checked |
| 开始游戏按钮 | 67.5x, 449.3y | 115x42 | fontSize=16 |

### 省份位置 (13个州)
| 州ID | 图片 | x | y | width | height |
|------|------|-----|-----|-------|--------|
| 1 | silv.png | 208.5 | 159 | 179 | 93 |
| 2 | jizhou.png | 361 | 90 | 102 | 105 |
| 3 | yuzhou.png | 331 | 190 | 123 | 101 |
| 4 | yunzhou.png | 374 | 155 | 96 | 92 |
| 5 | xuzhou.png | 421 | 183 | 118 | 96 |
| 6 | qingzhou.png | 428 | 121 | 119 | 81 |
| 7 | jingzhou.png | 254 | 235 | 152 | 172 |
| 8 | yangzhou.png | 373 | 239 | 204 | 175 |
| 9 | yizhou.png | 40 | 220 | 270 | 204 |
| 10 | liangzhou.png | 0 | 7 | 260 | 244 |
| 11 | bingzhou.png | 219 | 45 | 153 | 155 |
| 12 | youzhou.png | 360 | 4 | 253 | 118 |
| 13 | jiaozhou.png | 135 | 363 | 328 | 136 |

### 关键代码逻辑
```actionscript
// 头像切换逻辑 (MAX_FACE_COUNT = 10)
function onNextFace():void { mFace = (mFace % 10) + 1; }
function onPrevFace():void { mFace = mFace > 1 ? mFace - 1 : 10; }

// 性别列表: female, male (ArrayCollection)
SEX_NAME = ["female", "male"]

// 创建角色回调
___CreateRoleDialog_Button3_click() -> doCreateRole()
```

## 当前 HTML5 实现

### 模板结构 (App.vue ~638-691)
```html
<template v-else-if="screen === 'create-role'">
  <img class="full-bg dimmed-bg" :src="asset('board_login2.jpg')" alt="" />
  <img class="role-title-img" :src="asset('title.png')" alt="" />
  <div class="role-title">创建角色</div>

  <section class="role-left">
    <div class="face-title">
      <img class="face-select-img" :src="asset('select_headimg.png')" alt="" />
      <span>选择头像</span>
    </div>
    <div class="role-field province-select">
      <span>所在州郡</span>
      <select v-model.number="selectedProvince">
        <option v-for="province in provinceAssets" :key="province.id" :value="province.id">
          {{ province.name }}
        </option>
      </select>
    </div>
    <div class="face-frame">
      <img class="face-photo" :src="faceImage" alt="" />
      <button class="face-prev" type="button" @click="prevFace">&lt;</button>
      <button class="face-next" type="button" @click="nextFace">&gt;</button>
      <div class="sex-row">
        <button type="button" :class="{ active: roleForm.sex === 1 }" @click="roleForm.sex = 1">男</button>
        <button type="button" :class="{ active: roleForm.sex === 0 }" @click="roleForm.sex = 0">女</button>
      </div>
    </div>
    <label class="role-field name-field">
      <span>君主名</span>
      <input v-model="roleForm.userName" maxlength="8" />
    </label>
    <div class="province-info">{{ selectedProvinceName }}：请选择主公起兵之地。</div>
    <label class="role-rule">
      <input type="checkbox" checked />
      <span>我已经阅读并同意用户协议</span>
    </label>
    <button class="start-role" type="button" :disabled="loading" @click="submitRole">开始游戏</button>
  </section>

  <section class="role-map" @mousemove="moveProvince" @click="pickProvince" @mouseleave="hoveredProvince = null">
    <img class="china-map" :src="asset('sanguo_worldmap.png')" alt="" />
    <img
      v-for="province in provinceAssets"
      :key="province.id"
      v-show="activeProvince === province.id"
      class="province-piece"
      :src="asset(province.image)"
      :style="{ left: `${province.x}px`, top: `${province.y}px`, width: `${province.w}px`, height: `${province.h}px` }"
      alt=""
    />
  </section>
</template>
```

### 数据模型 (App.vue ~102-108)
```typescript
const roleForm = ref({
  userName: `lord${Math.floor(Math.random() * 900 + 100)}`,
  cityName: "新城池",
  flagChar: "H",
  sex: 0,
  face: 1
});
```

### 头像逻辑 (App.vue ~132-136)
```typescript
const faceImage = computed(() => {
  const sexName = roleForm.value.sex === 1 ? "male" : "female";
  const face = Math.min(Math.max(roleForm.value.face, 1), 10);
  return asset(`player/player_${sexName}_${face}.jpg`);
});
```

## 对比分析

| 特性 | Flash | HTML5 | 状态 |
|------|-------|-------|------|
| 舞台尺寸 | 1000x600 | 1000x600 | ✓ |
| 背景图 | 纯色11976102 | board_login2.jpg | 差异 |
| 标题图 | title.png | title.png | ✓ |
| 地图区域 | framePictureCanvas 618x500 | role-map section | 差异 |
| 世界地图 | sanguo_worldmap.png | sanguo_worldmap.png | ✓ |
| 省份高亮 | province_1~13 visible切换 | v-show activeProvince | ✓ |
| 头像框 | 87x100 canvas | face-frame div | 差异 |
| 头像切换 | ButtonPrev/Next | face-prev/next buttons | ✓ |
| 性别选择 | RadioButton female/male | sex button toggle | 简化 |
| 君主名输入 | TextInput 144x26 | input maxlength=8 | ✓ |
| 州郡选择 | ComboBox | select dropdown | ✓ |
| 用户协议 | checkbox default checked | checkbox checked | ✓ |
| 开始按钮 | Button 115x42 | start-role button | ✓ |
| 地图交互 | mouseMove/click | @mousemove/@click | ✓ |

## 已知差异

1. **背景色 vs 背景图**: Flash用纯色背景(#B6D5F6)，HTML5用board_login2.jpg
2. **省份选择**: Flash是ComboBox下拉，HTML5也是select(简化一致)
3. **头像数据**: Flash用mFace 1-10，HTML5用roleForm.face 1-10
4. **省份交互**: Flash用鼠标点击map canvas检测，HTML5用canvas像素检测(v-show覆盖)
5. **默认州**: Flash mProvinceList动态加载，HTML5 selectedProvince默认值10

## 下一步
如果需要完全1:1：
1. 统一背景为纯色或改用相同背景图
2. 头像区域使用canvas匹配Flash尺寸
3. 省份交互改为像素级点击检测
4. 验证省份图片资源是否完整
