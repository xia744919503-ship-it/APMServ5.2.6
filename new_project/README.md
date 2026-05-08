# 热血三国 `new_project`

这是基于旧版【热血三国】单机服做的现代化重构工作区。

当前已经完成的重构切片：

- Go API 直接读取旧库 `bloodwar`，并提供项目概览、城池详情、世界地图接口
- Vue 3 + TypeScript + Vite + Pinia 前端骨架已经接上真实 API
- Element Plus 已经承载概览页和城池详情页
- PixiJS 已经用于世界地图的格点渲染
- 旧库不可用时，会自动退回到内置样例数据，前端仍然可以运行

## 目录

- `backend`: Go 后端
- `frontend`: Vue 3 前端
- `docs`: 项目理解与迁移分析
- `scripts`: 本地验证脚本

## 运行

1. 确保旧库可用。

   默认配置会连接：

   - 地址：`127.0.0.1:3306`
   - 数据库：`bloodwar`
   - 用户：`root`
   - 密码：空

2. 启动后端：

```powershell
cd d:\APMServ5.2.6\new_project\backend
go run ./cmd/api
```

3. 启动前端开发服务器：

```powershell
cd d:\APMServ5.2.6\new_project\frontend
npm install
npm run dev
```

4. 生产构建：

```powershell
cd d:\APMServ5.2.6\new_project\frontend
npm run build
```

构建完成后，Go 后端会自动托管 `frontend/dist`。

## 验证

启动后端后，可以执行运行态烟测脚本：

```powershell
cd d:\APMServ5.2.6\new_project
.\scripts\smoke-api.ps1
```

这个脚本会验证：

- 主公列表
- 登录 / 退出登录
- 当前会话
- 我的城池
- 城池详情
- 税率 PATCH（使用原值回写）
- 资源分配 PATCH（使用原值回写）
- 后端首页是否正确托管 `frontend/dist`

## 主要接口

- `GET /api/health`
- `GET /api/dashboard/overview`
- `GET /api/cities?limit=24`
- `GET /api/cities/:cid`
- `GET /api/world/map?cid=215265&radius=8`
