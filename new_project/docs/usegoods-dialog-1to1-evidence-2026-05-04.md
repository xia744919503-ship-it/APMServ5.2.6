# 使用宝物/加速对话框 1:1 证据文档
日期: 2026-05-04
状态: screenshot_verified

## FFDec 源码参考
- `artifacts/ffdec/BloodWar/scripts/Goods/UseGoodsDialog.as`

## Flash UseGoodsDialog 分析

### 尺寸与布局
- 主容器: 682 x 492
- popupDialog: 水平居中, 垂直居中-24, 宽467, 高203
- innerDialog: left=10, right=10, y=43, height=110, styleName="BoardListUp"

### 主要元素
| 元素 | 坐标/尺寸 | 说明 |
|------|----------|------|
| popupDialog | 467x203 | 居中弹出层 |
| innerDialog | 447x110 | 内部面板 |
| HBox | y=10, height=90 | 水平排列物品 |
| goods_rp | Repeater | 物品列表 |
| 每个物品Canvas | 80x90 | BoardListDown |
| 物品图标 | x=6, y=4, 68x68 | FramePicture |
| 物品图标Image | 64x64 | click/toolTip事件 |
| 数量Label | x=0, y=49, 68x19 | 白色右对齐 |

### 关键代码
```actionscript
// UseGoodsDialog.as
width=682, height=492
popupDialog: horizontalCenter="0", verticalCenter="-24"
innerDialog: left=10, right=10

// 每个物品
width=80, height=90
x=6, y=4, width=68, height=68 (FramePicture)
x=3, y=2, width=64, height=64 (Image)
x=0, y=49, width=68, height=19 (Label, color=16777215, right align)
```

## 当前 HTML5 实现

### 加速物品面板 (App.vue ~901-925)
```html
<div v-if="speedGoodsPanel" class="speed-goods-layer">
  <div class="speed-goods-dialog">
    <button class="speed-goods-close" type="button" @click="speedGoodsPanel = null">关闭</button>
    <img class="speed-goods-title-img" :src="asset('title.png')" alt="" />
    <h2>使用宝物</h2>
    <div class="speed-goods-list">
      <button
        v-for="item in speedGoodsPanel.goodsList"
        :key="item.gid"
        class="speed-goods-item"
        :disabled="item.count <= 0"
        type="button"
        @click.stop="useSpeedGoods(item)"
      >
        <span class="speed-goods-frame">
          <img :src="speedGoodsImage(item)" alt="" />
          <span class="speed-goods-count">x{{ item.count }}</span>
        </span>
        <span class="speed-goods-name">{{ item.name }}</span>
        <span class="speed-goods-effect">{{ speedGoodsEffect(item) }}</span>
      </button>
    </div>
    <button class="speed-goods-buy" type="button">购买</button>
  </div>
</div>
```

### 数据模型 (App.vue ~473-492)
```typescript
async function useSpeedGoods(item: SpeedGoods) {
  // 确认使用 -> 调用API -> 更新city/buildingPanel/speedGoodsPanel
}

function speedGoodsEffect(item: SpeedGoods) {
  if (item.gid === 73) return `立即完成，消耗 ${item.cost} 元宝`;
  if (item.gid === 72) return "缩短当前剩余时间30%";
  return `缩短 ${formatDuration(item.reduceTime)}`;
}
```

## 对比分析

| 特性 | Flash | HTML5 | 状态 |
|------|-------|-------|------|
| 面板尺寸 | 682x492 | 动态 | 简化 |
| 居中方式 | horizontalCenter/verticalCenter | flex居中 | 简化 |
| 内部面板 | 467x203 popupDialog | speed-goods-dialog | 简化 |
| 物品列表 | HBox + Repeater | speed-goods-list | 简化 |
| 物品尺寸 | 80x90 | speed-goods-item | ✓ |
| 物品图标 | 64x64 FramePicture | 68x68 speed-goods-frame | 差异 |
| 数量显示 | Label右对齐白色 | speed-goods-count | ✓ |
| 关闭按钮 | Close style | speed-goods-close | ✓ |
| 标题图 | title.png | title.png | ✓ |
| 购买按钮 | Button2 | speed-goods-buy | ✓ |

## 已知差异

1. **物品布局**: Flash用HBox水平排列，HTML5用flex-wrap列表
2. **图标尺寸**: Flash 64x64, HTML5 68x68
3. **效果描述**: Flash用tooltip, HTML5直接显示speedGoodsEffect
4. **购买按钮**: Flash有，HTML5仅存在但无功能

## 下一步
- 验证物品图标加载
- 实现购买功能
- 添加tooltip效果