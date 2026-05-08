# 任务对话框 1:1 证据文档
日期: 2026-05-04
状态: screenshot_verified

## FFDec 源码参考
- `artifacts/ffdec/BloodWar/scripts/Task/TaskDialog.as`
- `artifacts/ffdec/BloodWar/scripts/Hotel/HotelDialog.as`
- `artifacts/ffdec/BloodWar/scripts/Hotel/RewardTaskDialog.as`

## Flash 任务对话框分析

### TaskDialog 主要布局
- 尺寸: 682 x 492
- 标题图: x=215, y=12
- 标题文字: x=264, y=16, width=151, height=18, color=15192461
- 任务类型按钮: 
  - Button1: x=21, y=41, width=75, styleName="ButtonRed"
  - Button2: x=106, y=41, width=75, styleName="ButtonGreen"
  - Button3-6: x=189-438, y=41, width=75, styleName="ButtonYellow"
- 领取按钮: x=314, y=449, width=85, styleName="ButtonYellow"

### 数据结构
```actionscript
class TaskDialog {
  taskListCtl: DataGrid  // 任务列表
  taskGroupListCtl: DataGrid  // 任务组列表
  mTasks: ArrayCollection
  mTaskGroups: ArrayCollection
  mGetRewardVisible: Boolean
  REWARD_RES_NAME: Array
}
```

### HotelDialog 任务子面板
- mRewardTaskDialog: RewardTaskDialog (奖励任务)
- mSystemRewardTaskDialog: SystemRewardTaskDialog (系统任务)
- 布局位置: x=20, y=125

## 当前 HTML5 实现

### 任务面板 (App.vue ~927-958)
```html
<div v-if="taskPanelVisible" class="task-layer">
  <div class="task-dialog">
    <button class="task-close" type="button" @click="taskPanelVisible = false; selectedTaskId = null">关闭</button>
    <img class="task-title-img" :src="asset('title.png')" alt="" />
    <h2>任务列表</h2>
    <div class="task-list">
      <template v-for="category in taskSnapshot?.categories ?? []" :key="category.type">
        <template v-for="group in category.groups ?? []" :key="group.id">
          <button
            v-for="task in group.tasks ?? []"
            :key="task.id"
            class="task-item"
            :class="{ active: selectedTaskId === task.id, completed: task.completed }"
            type="button"
            @click="selectTask(task.id)"
          >
            <span class="task-name">{{ task.name }}</span>
            <span class="task-progress">{{ task.completedGoals }}/{{ task.goalCount }}</span>
          </button>
        </template>
      </template>
    </div>
    <button
      v-if="selectedTaskId"
      class="task-claim-btn"
      type="button"
      @click="handleClaimReward"
    >
      领取奖励
    </button>
  </div>
</div>
```

### 任务API (App.vue ~450-472)
```typescript
async function openTaskPanel() {
  await withLoading(async () => {
    taskSnapshot.value = await myTasks();
    taskPanelVisible.value = true;
  });
  if (currentGuide.value?.gid === 11) setGuide(12);
}

async function selectTask(taskId: number) {
  selectedTaskId.value = taskId;
  if (currentGuide.value?.gid === 12) setGuide(13);
}

async function handleClaimReward() {
  const taskId = selectedTaskId.value;
  if (!taskId) return;
  await withLoading(async () => {
    taskSnapshot.value = await claimTaskReward(taskId);
    selectedTaskId.value = null;
    taskPanelVisible.value = false;
  });
  if (currentGuide.value?.gid === 13) setGuide(14);
}
```

## 对比分析

| 特性 | Flash | HTML5 | 状态 |
|------|-------|-------|------|
| 面板尺寸 | 682x492 | 动态 | 简化 |
| 关闭按钮 | Close style | task-close | ✓ |
| 标题图 | title.png x=215 | title.png | ✓ |
| 标题文字 | "任务" x=264 | "任务列表" | 差异 |
| 任务分类 | Button1-9 (多类型) | categories数组 | 重构 |
| 任务列表 | DataGrid | task-list div | 简化 |
| 任务项 | DataGrid row | task-item button | ✓ |
| 进度显示 | DataGrid列 | task-progress span | ✓ |
| 领取按钮 | ButtonYellow | task-claim-btn | ✓ |
| 按钮条件 | mGetRewardVisible | v-if="selectedTaskId" | ✓ |

## 已知差异

1. **任务分类**: Flash用9个固定按钮切换类型，HTML5用嵌套循环
2. **任务选择**: Flash用DataGrid单选，HTML5用v-model selectedTaskId
3. **进度显示**: Flash在DataGrid列中，HTML5在task-item内
4. **奖励预览**: Flash有详情面板，HTML5暂无
5. **Hotel任务**: Flash有HotelDialog子面板，HTML5统一task-panel

## 下一步
- 完善任务详情显示
- 实现多任务分类标签
- 添加奖励预览功能