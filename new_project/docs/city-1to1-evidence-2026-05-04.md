# 城市场景 1:1 证据文档
日期: 2026-05-04
状态: screenshot_verified

## FFDec 源码参考
- `artifacts/ffdec/BloodWar/scripts/Government/CityFieldDialog.as` (城内建筑区域)
- `artifacts/ffdec/BloodWar/scripts/BloodWar.as` (主场景)

## Flash 城市场景分析

### 尺寸
- 整个舞台: 1000 x 600
- 城内面板区域: 相对布局

### 主要元素 (CityFieldDialog)
| 元素 | Flash坐标/尺寸 | 说明 |
|------|---------------|------|
| 外景面板 | 1000x600全舞台 | 包含地图和点击区域 |
| 内城列表按钮 | topbutton_innercity | 切换到城内视图 |
| 外城按钮 | topbutton_outercity | 城外视图 |
| 地图按钮 | topbutton_map | 地图视图 |
| 任务按钮 | topbutton_mission | 打开任务面板 |
| 报告按钮 | topbutton_report | 报告 |
| 联盟按钮 | topbutton_union | 联盟 |
| 资源显示 | 左侧面板 | 黄金/粮食/木材等 |
| 内城地图 | 地图区域 | map_innercity.jpg |

### 关键代码
```actionscript
// BloodWar.as 中的内城切换
function openInnerCity():void { ... }
function openOuterCity():void { ... }

// CityFieldDialog.as - 城内建筑视图
function onShow():void {
  Command(Global.cmd).sendGlobalCommand("city","getCityField",_loc1_);
}
```

## 当前 HTML5 实现

### 城市界面模板 (App.vue ~693-782)
```html
<template v-else-if="screen === 'city' && city">
  <img class="full-bg city-bg" :src="asset('board_login2.jpg')" alt="" />
  <img class="left-board" :src="asset('leftboard_new.jpg')" alt="" />
  <div class="top-tabs">
    <img :src="asset('topbutton_innercity_on.png')" alt="城内" />
    <img :src="asset('topbutton_outercity_up.png')" alt="城外" />
    <img :src="asset('topbutton_map_up.png')" alt="地图" />
    <img :src="asset('topbutton_mission_up.png')" alt="任务" @click="openTaskPanel" style="cursor:pointer" />
    <img :src="asset('topbutton_report_up.png')" alt="报告" />
    <img :src="asset('topbutton_union_up.png')" alt="联盟" />
  </div>
  <div class="left-data">
    <div class="lord">{{ user?.name || city.summary.owner }}</div>
    <div class="city-name">{{ city.summary.name }} ({{ city.summary.x }},{{ city.summary.y }})</div>
    <div>民心 {{ city.morale }} 税率 {{ city.tax }}%</div>
    <div class="res"><img :src="asset('resource_food.png')" alt="" />粮食 {{ formatNumber(resources?.food) }}</div>
    <div class="res"><img :src="asset('resource_wood.png')" alt="" />木材 {{ formatNumber(resources?.wood) }}</div>
    <div class="res"><img :src="asset('resource_rock.png')" alt="" />石料 {{ formatNumber(resources?.rock) }}</div>
    <div class="res"><img :src="asset('resource_iron.png')" alt="" />铁锭 {{ formatNumber(resources?.iron) }}</div>
    <div class="res"><img :src="asset('city_gold.png')" alt="" />黄金 {{ formatNumber(resources?.gold) }}</div>
    <div>人口 {{ formatNumber(resources?.people) }}/{{ formatNumber(resources?.peopleMax) }}</div>
  </div>
  <div class="city-view">
    <img class="inner-map" :src="asset(isOccupied(cityWallPlot.position) ? 'map_innercity_high.jpg' : 'map_innercity.jpg')" alt="" />
  </div>
  <!-- 城墙/官府/内城地块点击区域 -->
  ...
</template>
```

## 对比分析

| 特性 | Flash | HTML5 | 状态 |
|------|-------|-------|------|
| 舞台尺寸 | 1000x600 | 1000x600 | ✓ |
| 顶部标签 | topbutton_*.png | topbutton_*.png | ✓ |
| 任务按钮交互 | 点击打开任务 | @click="openTaskPanel" | ✓ |
| 资源面板 | 左侧显示 | left-data div | ✓ |
| 城内地图 | map_innercity.jpg | map_innercity.jpg | ✓ |
| 内城地块 | 绝对定位按钮 | plot-hit buttons | ✓ |
| 建筑显示 | 动态加载 | isOccupied check | ✓ |

## 已知差异

1. **切换动画**: Flash使用视图切换动画，HTML5当前为单视图
2. **城外视图**: 尚未实现外城视图切换
3. **地图视图**: 尚未实现地图视图
4. **建筑交互**: Flash通过CityFieldDialog，HTML5使用buildPanel overlay

## 下一步
- 实现城外视图切换
- 实现地图视图
- 完善内城/外城切换动画