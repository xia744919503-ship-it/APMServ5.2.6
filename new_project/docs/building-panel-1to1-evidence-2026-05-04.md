# 建筑面板 1:1 证据文档
日期: 2026-05-04
状态: screenshot_verified

## FFDec 源码参考
- `artifacts/ffdec/BloodWar/scripts/Building/CreateBuildingDialog.as`
- `artifacts/ffdec/BloodWar/scripts/Building/UpgradeBuildingDialog.as`
- `artifacts/ffdec/BloodWar/scripts/Building/BuildingInfo.as`

## Flash 建筑面板分析

### CreateBuildingDialog (建造面板)
- 尺寸: 682 x 492
- 主要元素:
  - 关闭按钮: x=603, y=445, styleName="Close"
  - itemContainer: y=44, height=386, 内部滚动列表
  - 标题图: y=10, 水平居中
  - 标题文字: y=15

### UpgradeBuildingDialog (升级面板)
- 尺寸: 682 x 166
- 主要元素:
  - 标题图: x=218, y=11
  - 标题文字: x=266, y=16, width=151, height=18
  - 关闭按钮: styleName="Close", bottom=10, right=20, fontSize=16
  - buildingItem: x=18, y=46, width=646

### BuildingInfo (建筑信息数据模型)
```actionscript
class BuildingInfo {
  bid, level, name, levelDescription
  woodNeed, rockNeed, ironNeed, foodNeed, goldNeed
  peopleNeed, upgradeTime, canUpgrade
  conditions: Array
}
```

## 当前 HTML5 实现

### 建造面板 (build-panel) - App.vue ~868-899
```html
<div v-if="buildPanel && selectedPlot" class="modal-layer">
  <div class="build-panel">
    <button class="build-close" type="button" @click="closeBuild">关闭</button>
    <img class="build-title-img" :src="asset('title.png')" alt="" />
    <h2>建造建筑</h2>
    <div v-if="buildPanel.slot.occupied" class="build-message">这里已经有建筑...</div>
    <div v-else-if="!buildPanel.slot.unlocked" class="build-message">官府等级不足...</div>
    <div v-else class="build-list">
      <div v-for="option in buildPanel.options" :key="option.bid" class="build-option">
        <img :src="buildingIntro(option)" alt="" />
        <button class="create-building-btn" type="button" @click.stop="confirmBuild(option)">建造</button>
        <strong>{{ option.name }}</strong>
        <span>{{ buildOptionDescription(option) }}</span>
        <small>{{ option.canBuild ? `耗时 ${option.duration} 秒` : option.reason }}</small>
      </div>
    </div>
  </div>
</div>
```

### 建筑信息面板 (building-panel) - App.vue ~815-866
```html
<div v-if="buildingPanel && selectedPlot" class="modal-layer">
  <div class="building-panel">
    <button class="building-close" type="button" @click="closeBuild">关闭</button>
    <img class="building-title-img" :src="asset('title.png')" alt="" />
    <h2>{{ buildingPanel.name }}(等级{{ displayBuildingLevel(buildingPanel) }})</h2>
    <div class="building-item">
      <div class="building-icon-frame">
        <img :src="buildingDialogImage(buildingPanel)" :alt="buildingPanel.name" />
      </div>
      <button v-if="buildingPanel.state === 0" class="building-upgrade-btn">升级</button>
      <button v-if="buildingPanel.state === 0" class="building-destroy-btn">拆除</button>
      <button v-if="buildingPanel.state !== 0" class="building-speed-btn">加速</button>
      <button v-if="buildingPanel.state !== 0" class="building-cancel-btn">取消</button>
      <div class="building-info-board">
        <div class="building-description">{{ buildingDescriptionText }}</div>
        <div class="building-state">{{ buildingNeedText || buildingStateText }}</div>
      </div>
    </div>
  </div>
</div>
```

## 对比分析

| 特性 | Flash | HTML5 | 状态 |
|------|-------|-------|------|
| 面板宽度 | 682px | 动态/响应式 | 简化 |
| 关闭按钮 | Close style | build-close class | ✓ |
| 标题图 | title.png 居中 | title.png | ✓ |
| 建筑列表 | itemContainer with scroll | build-list div | 简化 |
| 建造按钮 | 创建按钮 | create-building-btn | ✓ |
| 升级面板 | UpgradeBuildingDialog | building-panel | 差异 |
| 升级按钮 | buildingItem内 | building-upgrade-btn | ✓ |
| 拆除按钮 | buildingItem内 | building-destroy-btn | ✓ |
| 加速按钮 | buildingItem内 | building-speed-btn | ✓ |
| 取消按钮 | buildingItem内 | building-cancel-btn | ✓ |

## 已知差异

1. **面板尺寸**: Flash固定682px，HTML5响应式
2. **状态显示**: Flash用BuildingInfo模型，HTML5用buildingPanel/buildingInfoPanel
3. **条件检查**: Flash conditions数组，HTML5 option.canBuild/option.reason
4. **升级时间**: Flash有upgradeTime，HTML5用buildingRemaining计算
5. **建筑图标**: Flash用BuildingInfo图片，HTML5用buildingIntro()

## 下一步
- 统一建筑面板尺寸匹配Flash
- 完善升级条件显示
- 实现建筑动画效果