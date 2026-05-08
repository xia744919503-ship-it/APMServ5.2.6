# Guide Coordinate Model
Date: 2026-05-04

## Coordinate System

### Stage Coordinates (1000x600)
- Game area defined in Flash: 1000x600
- Used for guide showpos values
- Origin (0,0) at top-left of game stage

### Viewport Coordinates (Browser Pixels)
- Actual screen pixels from getBoundingClientRect()
- stage.left, stage.top = viewport offset
- stageScale = stage.width / 1000

### Conversion Formula
```
viewport_x = stage.left + stage_x * stageScale
viewport_y = stage.top + stage_y * stageScale
```

## P0 Guide Chain (gid6-gid14)

| Step | Flash showpos | Stage Coords | Viewport Coords | Hit Element |
|------|---------------|--------------|-----------------|-------------|
| g6 | 570,130,100,60 | (620, 160) | computed | .plot-hit |
| g7 | 405,142,53,23 | (431, 153) | computed | SPAN (民房图标) |
| g8 | 570,130,100,60 | (620, 160) | computed | .plot-hit.occupied.busy |
| g9 | 399,240,44,21 | (421, 250) | computed | .building-speed-btn |
| g10 | 413,234,68,81 | (447, 274) | computed | IMG (鲁班残页) |
| g11 | 912,4,68,24 | (946, 16) | computed | IMG (任务按钮) |
| g12 | 312,392,145,24 | (511.5, 128) | computed | .task-item |
| g13 | 795,504,77,26 | OUTSIDE STAGE | (1294, 690) | .task-claim-btn |

## Key Findings

### g12 Issue
Flash showpos (312,392) hits task-list, not task-item. Corrected to (511.5, 128) based on task-item center.

### g13 Issue
Claim button is outside stage (task-dialog extends below stage). Uses Click-Viewport with absolute viewport coordinates (1294, 690).

### Debug Rectangles (Latest Run)
```
taskBtnRect:   {"x":1346,"y":179.5,"width":116,"height":45}
taskItemRect:  {"x":649,"y":280.5,"width":622,"height":40}
claimBtnRect:  {"x":1255,"y":676.5,"width":77,"height":26}
```

## Coordinate Verification
Both regression scripts pass:
- verify-speed-flow.ps1: PASS (DB task1_state=1)
- verify-speed-flow-coordinates.ps1: PASS (DB task1_state=1)

## Implementation Notes
- Click-Stage uses elementFromPoint with stage-scale transformation
- Click-Viewport uses elementFromPoint with absolute coordinates
- g13 requires Click-Viewport because claim button is in task-dialog (outside stage)