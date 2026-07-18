# 设计：auth-rbac 主服务接入

## 技术方案

### 变更文件清单

| 文件 | 操作 | 说明 |
|------|------|------|
| app/setup/setup.go | 修改 | runMigrations 调用 auth.autoMigrate + 初始数据创建 admin/default 租户 |
| cmd/api/server.go | 修改 | preRun/run 中使用 auth.Init 代替旧路由注册 |
| app/admin/router/init_router.go | 修改 | InitRouter 不再注册旧路由，改为调用 auth.Init |
| common/middleware/init.go | 修改 | 移除旧 JWT/Casbin 中间件注册（保留文件） |
| common/auth/auth.go | 不变 | 已有完整 Init 函数 |

### 核心逻辑

#### 1. setup.go — runMigrations 修改

```go
import authPkg "go-admin/common/auth"

func runMigrations(db *gorm.DB) error {
    // 新 RBAC 表（替代旧 sys_* 表）
    if err := authPkg.AutoMigratePublic(db); err != nil {
        return fmt.Errorf("auth-rbac 表迁移失败: %w", err)
    }
    // 保留插件管理表
    if err := db.AutoMigrate(&pluginModels.SysPlugin{}); err != nil {
        return fmt.Errorf("插件表迁移失败: %w", err)
    }
    return nil
}
```

#### 2. setup.go — 初始数据（替代 adminModels.InitDb）

```go
func seedInitialData(db *gorm.DB) error {
    // 1. 创建超级管理员
    hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
    admin := &authModel.User{Username: "admin", Password: string(hashedPwd), Nickname: "超级管理员", Status: 1, Version: 1}
    db.Create(admin)

    // 2. 创建默认租户
    tenant := &authModel.Tenant{TenantCode: "default", Name: "默认租户", Status: 1, Version: 1}
    db.Create(tenant)

    // 3. 关联 admin → default 租户
    db.Create(&authModel.UserTenant{UserID: admin.ID, TenantID: tenant.ID})

    // 4. 创建 SUPER_ADMIN 角色
    role := &authModel.Role{TenantID: tenant.ID, RoleCode: "SUPER_ADMIN", RoleName: "超级管理员", RoleType: "SUPER_ADMIN", Status: 1, Version: 1}
    db.Create(role)

    // 5. 分配角色
    db.Create(&authModel.UserRole{UserID: admin.ID, RoleID: role.ID, TenantID: tenant.ID})

    return nil
}
```

#### 3. server.go — 启动时集成

```go
// 在 run() 函数中，已安装时注册业务路由的位置：
if setup.IsInstalled() {
    // 获取 DB 连接（GetDb 返回 map，取第一个可用连接）
    var db *gorm.DB
    for _, d := range sdk.Runtime.GetDb() {
        if d != nil {
            db = d
            break
        }
    }
    r := sdk.Runtime.GetEngine().(*gin.Engine)

    // 初始化 auth-rbac（注册路由+中间件）
    authCfg := authConfig.DefaultConfig()
    if config.JwtConfig != nil && config.JwtConfig.Secret != "" {
        authCfg.JWT.Secret = config.JwtConfig.Secret  // 复用 settings.yml 中的 jwt secret
    }
    if err := authPkg.Init(authCfg, db, r); err != nil {
        log.Fatalf("auth-rbac 初始化失败: %v", err)
    }
}
```

注意：RegisterOnInstalled 回调中同样使用相同逻辑获取 DB。

#### 4. init_router.go — 不注册旧路由

```go
func InitRouter() {
    // 旧路由已迁移到 auth-rbac 模块，由 server.go 中 auth.Init() 统一注册
    // 保留空函数体避免 AppRouters 调用报错
}
```

#### 5. common/middleware/init.go — 仅移除 JWT/Casbin 注册

```go
func InitMiddleware(r *gin.Engine) {
    r.Use(DemoEvn())
    r.Use(WithContextDb)       // 保留：DB 注入
    r.Use(LoggerToFile())      // 保留：日志
    r.Use(CustomError)         // 保留：错误处理
    r.Use(NoCache)             // 保留
    r.Use(Options)             // 保留：CORS
    r.Use(Secure)              // 保留：安全头
    // 以下 3 行移除（旧 JWT + Casbin）：
    // sdk.Runtime.SetMiddleware(JwtTokenCheck, (*jwt.GinJWTMiddleware).MiddlewareFunc)
    // sdk.Runtime.SetMiddleware(RoleCheck, AuthCheckRole())
    // sdk.Runtime.SetMiddleware(PermissionCheck, actions.PermissionAction())
}
```

### 不变行为清单

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | /setup 安装向导流程不变 | 清 DB 后访问 /setup 正常显示 |
| RG-2 | 前端 /init 页面不变 | 初始化页面正常渲染和提交 |
| RG-3 | 插件系统不受影响 | 插件管理表正常迁移 |
| RG-4 | go build 零错误 | 旧文件保留但不影响编译 |
