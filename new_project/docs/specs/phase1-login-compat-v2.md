# Phase 1 登录兼容（V2）

## 目标

把旧版登录入口链路完整搬到 Go 侧，确保可复刻旧流程：

1. 登录公告  
2. 登录验证  
3. 排队检查  
4. 首次登录自动建号（`state=3`，等待建角）

## 接口

- `GET /api/legacy/login/announcement`
- `POST /api/legacy/login`
- `POST /api/legacy/login/queue`

## 说明：注册语义

旧项目里“账号注册”主要在通行证系统（站外）完成；游戏服 `Login.doLogin` 的关键职责是：

- 首次登录时，如果 `sys_user` 不存在，则自动创建账号记录并进入 `state=3`。

当前 Go 兼容层保持这个语义不变，因此“注册+登录”闭环可由：

1. 通行证创建账号（或直接使用新 passport）
2. `POST /api/legacy/login`
3. `POST /api/legacy/role/create`（Phase 2）

## 请求示例

### 1) 登录公告

```http
GET /api/legacy/login/announcement
```

### 2) 登录

```json
{
  "version": 0,
  "loginType": 0,
  "passType": "local",
  "passport": "autotest_20260501",
  "password": "",
  "auth": ""
}
```

### 3) 排队检查

```json
{
  "uid": 1007,
  "sid": 815203404
}
```

## 返回语义（保留 legacy raw）

- `raw=[2]`：版本/状态拒绝（按旧链路）
- `raw=[0, "..."]`：登录失败
- `raw=[1,1,uid,sid,queueCount]`：进入排队
- `raw=[1,2,uid,sid]`：直接登录成功
- queue:
  - `raw=[0]`：无排队记录
  - `raw=[1,1,queueOrder]`：继续排队
  - `raw=[1,0]`：排队通过

## 2026-05-01 验证结果

- 登录公告接口可返回非空内容。
- 新 passport 登录成功：`raw=[1,2,uid,sid]`。
- 可继续进入 Phase 2 建角建城链路，完成首次入服闭环。

