# 任务列表：auth-rbac-integration（主服务接入）

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 4 |
| 已完成 | 3 |
| 进行中 | 0 |
| 未开始 | 1 |
| 完成率 | 3/4 (75%) |
| 当前阶段 | Task 4: 端到端验证（需手动测试） |

### 状态说明

| 标记 | 状态 | 含义 |
|------|------|------|
| ✅ | 已完成 | 代码修改完成 + 编译通过 |
| 🔨 | 进行中 | 正在编码 |
| ⬜ | 未开始 | 尚未启动 |

---

### Task 1: 导出 AutoMigrate 函数 + 初始数据种子 ✅

**复杂度**: 中

**Scope:**
- 涉及文件: backend/common/auth/auth.go (增加导出函数), backend/common/auth/seed.go (新建)

**Acceptance:**
- AC: auth 包导出 AutoMigratePublic(db *gorm.DB) error 函数供 setup 调用
- AC: 新建 seed.go 提供 SeedInitialData(db *gorm.DB) error 函数
- AC: SeedInitialData 创建 admin 用户(bcrypt加密 admin123) + default 租户 + SUPER_ADMIN 角色 + 关联
- AC: go build ./common/auth/... 零错误

---

### Task 2: 修改 setup.go 对接新表 ✅

**依赖**: Task 1

**复杂度**: 高

**Scope:**
- 涉及文件: backend/app/setup/setup.go

**Constraints:**
- 保留 /setup 路由注册逻辑不变
- 保留 createDatabaseIfNotExists / writeConfig / InstallMiddleware 不变
- 仅修改 runMigrations 和初始数据部分

**Acceptance:**
- AC: runMigrations 调用 auth.AutoMigratePublic(db) 创建 admin_* 表
- AC: runMigrations 保留插件管理表迁移
- AC: doInstall 中初始数据部分改为调用 auth.SeedInitialData(db)
- AC: 不再检查 sys_user 表（改为检查 admin_user）
- AC: go build ./... 零错误

---

### Task 3: 修改 server.go + init_router.go 接入新路由 ✅

**依赖**: Task 2

**复杂度**: 高

**Scope:**
- 涉及文件: backend/cmd/api/server.go, backend/app/admin/router/init_router.go, backend/common/middleware/init.go

**Constraints:**
- 旧路由文件保留不删除
- 旧中间件文件保留不删除

**Acceptance:**
- AC: server.go 已安装时调用 auth.Init(cfg, db, engine) 注册新路由（DB 通过遍历 GetDb() map 获取）
- AC: server.go RegisterOnInstalled 回调中也调用 auth.Init
- AC: init_router.go InitRouter 函数体清空（保留空函数避免 AppRouters 调用报错）
- AC: common/middleware/init.go 仅移除旧 JWT/Casbin 3 行注册（保留 WithContextDb/LoggerToFile/CustomError/CORS 等通用中间件）
- AC: go build ./... 零错误
- AC: 服务启动后 POST /auth/login 可用

---

### Task 4: 端到端验证 ⬜

**依赖**: Task 3

**复杂度**: 中

**Scope:**
- 涉及文件: 无新文件（手动测试 + 确认）

**Acceptance:**
- AC: 删除 config/settings.yml → 启动服务 → 访问 /setup 正常
- AC: 执行初始化 → admin_* 表创建 + admin 用户 + default 租户存在
- AC: POST /auth/login {username: admin, password: admin123} → 返回 token
- AC: 使用 token 访问 GET /api/v1/tenants → 200
- AC: 不带 token 访问 GET /api/v1/tenants → 401
