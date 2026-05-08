# Phase 1 登录兼容说明

## 目标

先在 Go 后端提供可调用的旧登录语义入口，确保后续注册、建号、建城可以沿同一协议链路推进。

## 新增接口

- `POST /api/legacy/login`
- `POST /api/legacy/login/queue`

这两个接口会返回 `raw` 字段，用于保持旧 PHP `Login.doLogin` / `Login.checkQueue` 的数组语义。

## 请求结构

### `/api/legacy/login`

```json
{
  "version": 0,
  "loginType": 0,
  "passType": "local",
  "passport": "your-account",
  "password": "your-password",
  "auth": ""
}
```

当前阶段已实现：`loginType = 0`（账号密码入口）。

### `/api/legacy/login/queue`

```json
{
  "uid": 123,
  "sid": 456789
}
```

## 响应结构

统一响应示例：

```json
{
  "raw": [1, 2, 123, 456789],
  "logged": true,
  "queued": false,
  "uid": 123,
  "sid": 456789,
  "user": {
    "uid": 123,
    "name": "xxx"
  }
}
```

`raw` 是旧协议核心，当前已覆盖：

- `raw = [2]`: 版本不匹配且 `loginType=2`。
- `raw = [0, "..."]`: 登录失败。
- `raw = [1, 1, uid, sid, queueCount]`: 进入排队。
- `raw = [1, 2, uid, sid]`: 直接登录成功。
- `queue raw = [0]`: 无排队记录。
- `queue raw = [1, 1, queueOrder]`: 继续排队。
- `queue raw = [1, 0]`: 排队通过并登录成功。

## 当前阶段范围

已完成：

- 版本状态校验（`mem_state`）。
- 服务器状态校验（`mem_state` + 公告回退文案）。
- 账号查找（`passport + passtype`）。
- 新账号基础创建（`sys_user/sys_sessions/sys_online`）。
- 在线人数与排队判定（`sys_online/mem_queue`）。
- 排队检查与放行登录。
- 成功登录后桥接当前 cookie 会话（便于新前端共存调试）。

未完成（后续补齐）：

- `loginType=1` 的授权登录语义。
- 平台级 passport 验证脚本兼容（旧 `passport/*.php`）。
- 旧端完整错误文案逐条对齐。
- 成人防沉迷、每日礼包、合服补偿等旧链路副作用。

