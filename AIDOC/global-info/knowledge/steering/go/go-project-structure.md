# Go 项目目录结构规范

## 架构风格

采用**模块化单体架构**，按业务域划分为独立包（package），模块间通过接口进行松耦合通信，便于后续按需拆分为微服务。

基础框架：**Gin + GORM + go-admin-core**

---

## 顶层目录结构

```
projects/{project-name}/
├── backend/                    # Go 后端
├── frontend/                   # 前端（Vue 3 / pure-admin）
├── dist/                       # 构建产物
├── deploy/                     # 部署配置（Docker、k8s）
└── build.ps1                   # 一键构建脚本
```

---

## 后端目录结构

```
backend/
├── app/                        # 业务应用层
│   ├── {domain}/               # 业务域（如 game、admin、tenant）
│   │   ├── apis/               # HTTP 接口层（Controller）
│   │   ├── models/             # 数据模型（Entity）
│   │   ├── router/             # 路由注册
│   │   └── service/            # 业务逻辑层
│   │       ├── {service}.go    # Service 实现
│   │       └── dto/            # 数据传输对象
│   └── setup/                  # 安装向导模块
├── cmd/                        # 命令行入口
│   ├── api/                    # HTTP 服务启动
│   │   └── server.go
│   └── cobra.go                # CLI 根命令
├── common/                     # 公共组件
│   ├── actions/                # 数据权限
│   ├── database/               # 数据库初始化
│   ├── middleware/             # 中间件
│   │   ├── auth.go             # JWT 认证
│   │   ├── tenant.go           # 租户隔离
│   │   ├── permission.go       # 权限检查
│   │   └── handler/            # 认证处理器
│   ├── models/                 # 公共 Model（ControlBy、ModelTime、TenantBy）
│   └── global/                 # 全局常量
├── config/                     # 配置文件
│   ├── settings.yml            # 运行时配置（gitignore）
│   └── extend.go               # 扩展配置结构
├── web/                        # 前端静态文件（embed）
│   ├── dist/                   # 构建产物（gitignore）
│   └── embed.go                # go:embed 声明
├── docs/                       # Swagger 文档（自动生成）
├── go.mod
├── go.sum
└── main.go                     # 程序入口
```

---

## 业务域内部结构

每个业务域（domain）采用分层架构，以 `game` 域为例：

```
app/game/
├── apis/                       # HTTP 接口层
│   ├── game.go                 # 游戏管理接口
│   ├── dlc.go                  # DLC 管理接口
│   ├── player.go               # 玩家管理接口
│   └── order.go                # 订单管理接口
├── models/                     # 数据模型
│   ├── game.go                 # Game 实体
│   ├── dlc.go                  # Dlc 实体
│   ├── player.go               # Player 实体
│   └── order.go                # Order 实体
├── router/                     # 路由注册
│   ├── router.go               # 路由入口
│   └── dlc.go                  # DLC 相关路由
└── service/                    # 业务逻辑
    ├── game.go                 # Game Service
    ├── dlc.go                  # Dlc Service
    └── dto/                    # DTO 定义
        ├── game.go
        └── dlc.go
```

---

## 分层职责说明

| 层 | 包路径 | 职责 |
|----|--------|------|
| **Router** | `app/{domain}/router/` | 路由注册、中间件绑定 |
| **API（Controller）** | `app/{domain}/apis/` | 参数绑定、权限注入、调用 Service、返回响应 |
| **Service** | `app/{domain}/service/` | 业务逻辑、事务控制、调用 ORM |
| **DTO** | `app/{domain}/service/dto/` | 请求/响应数据结构，与 Model 解耦 |
| **Model（Entity）** | `app/{domain}/models/` | 数据库表映射，包含 GORM tag |
| **Middleware** | `common/middleware/` | 横切关注点（认证、租户、权限、日志） |
| **Common Models** | `common/models/` | 公共嵌入结构（ControlBy、ModelTime、TenantBy） |

---

## 公共 Model 规范

所有业务 Model 必须嵌入以下公共结构：

```go
// Model 基础主键
type Model struct {
    Id int `json:"id" gorm:"primaryKey;autoIncrement"`
}

// ModelTime 时间戳
type ModelTime struct {
    CreatedAt time.Time  `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
    DeletedAt *time.Time `json:"deletedAt,omitempty" gorm:"index"`
}

// ControlBy 操作人
type ControlBy struct {
    CreateBy int `json:"createBy" gorm:"comment:创建人"`
    UpdateBy int `json:"updateBy" gorm:"comment:更新人"`
}

// TenantBy 租户隔离（业务表必须嵌入）
type TenantBy struct {
    TenantId int `json:"tenantId" gorm:"index;default:0;comment:租户ID"`
}
```

**业务 Model 示例：**
```go
type Game struct {
    models.Model
    Name   string `json:"name" gorm:"size:128;not null;comment:游戏名称"`
    Status int    `json:"status" gorm:"default:1;comment:状态"`

    models.ModelTime
    models.ControlBy
    models.TenantBy  // 需要租户隔离的表必须嵌入
}
```

---

## 数据库表命名规范

| 模块 | 表前缀 | 示例 |
|------|--------|------|
| 系统管理 | `sys_` | `sys_user`、`sys_role`、`sys_menu` |
| 游戏管理 | `game_` | `game`、`game_dlc`、`game_player` |
| 租户管理 | `tenant` | `tenant` |

**通用字段（所有表必须包含）：**
- `id` — 主键（INT，自增）
- `created_at` — 创建时间
- `updated_at` — 更新时间
- `create_by` — 创建人 ID
- `update_by` — 更新人 ID
- `tenant_id` — 租户 ID（业务表，0=超级管理员）

---

## 配置文件规范

```yaml
# config/settings.yml
settings:
  application:
    name: game-server
    mode: prod          # dev/prod
    host: 0.0.0.0
    port: 8000
  logger:
    path: temp/logs
    level: info
  jwt:
    secret: ${JWT_SECRET}
    timeout: 86400      # 24小时（秒）
  database:
    driver: mysql
    source: ${DB_DSN}
```

**规则：**
- 敏感信息（密码、密钥）使用环境变量，不硬编码
- 生产环境 `mode: prod`，关闭 Swagger 和调试接口
- 配置文件不提交到 git（加入 `.gitignore`）

---

## 路由注册规范

```go
// 路由注册函数签名统一
func RegisterXxxRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
    api := apis.Xxx{}
    r := v1.Group("/xxx").
        Use(authMiddleware.MiddlewareFunc()).  // JWT 认证
        Use(middleware.WithTenantId()).         // 租户注入（业务接口必须）
        Use(middleware.AuthCheckRole())         // 权限检查
    {
        r.GET("", api.GetPage)
        r.GET("/:id", api.Get)
        r.POST("", api.Insert)
        r.PUT("/:id", api.Update)
        r.DELETE("/:id", api.Delete)
    }
}
```

---

## 模块间通信规范

- 禁止模块间直接引用对方的 Service 实现
- 跨模块调用通过接口（interface）进行
- 禁止模块间共享 Model 实体（通过 DTO 传递）
