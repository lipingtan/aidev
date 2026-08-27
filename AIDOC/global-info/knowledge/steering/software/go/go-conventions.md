# Go 编码规范

## 文件大小规范（强制）

- **小文件上限**：≤300 行 / ≤10KB；适用于大多数 service、handler、repository 文件
- **大文件上限**：≤800 行 / ≤30KB；仅适用于极端复杂、逻辑不可拆分的场景
- **超限处理**：必须按设计模式拆分为更多子类/接口实现（如策略模式、装饰器、分层 Repository）；禁止以"暂时"为由跳过拆分
- **函数级**：单函数建议 ≤50 行，极端不超过 100 行

## 命名规范

**包名：**
- 全部小写，简短，无下划线，无驼峰
- 包名应与目录名一致
- 示例：`middleware`、`models`、`service`、`dto`

**类型名（struct/interface）：**
- 使用 UpperCamelCase（大驼峰）
- 接口名通常以 `-er` 结尾（`Reader`、`Writer`、`Handler`）
- 示例：`GameService`、`TenantMiddleware`、`UserRepository`

**函数/方法名：**
- 导出函数使用 UpperCamelCase
- 非导出函数使用 lowerCamelCase
- 布尔返回值函数使用 `Is`、`Has`、`Can` 前缀
- 示例：`GetUserById`、`createOrder`、`IsActive`、`HasPermission`

**变量名：**
- 使用 lowerCamelCase，名称应有实际含义
- 缩写词全大写（`userID`、`httpURL`、`apiKey`）
- 循环变量可用单字母（`i`、`j`、`k`）
- 示例：`userID`、`orderStatus`、`pageSize`

**常量：**
- 使用 UpperCamelCase（Go 惯例，不用全大写）
- 枚举常量使用类型前缀
- 示例：`MaxRetryCount`、`StatusPending`、`RoleAdmin`

**错误变量：**
- 包级错误变量以 `Err` 开头
- 示例：`ErrNotFound`、`ErrUnauthorized`、`ErrInvalidParam`

## 代码格式规范

**强制使用 `gofmt`：**
- 所有代码必须通过 `gofmt` 格式化
- CI 中加入 `gofmt -l .` 检查，有差异则失败
- 推荐使用 `goimports`（自动管理 import）

**行长度：**
- 无硬性限制，但建议不超过 120 个字符
- 函数参数过多时换行，每个参数独占一行

**import 分组（按顺序）：**
```go
import (
    // 1. 标准库
    "context"
    "fmt"
    "time"

    // 2. 第三方库
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    // 3. 内部包
    "go-admin/common/models"
    "go-admin/app/game/service"
)
```

**空行规则：**
- 函数之间用 1 个空行分隔
- 函数内逻辑分组之间用 1 个空行
- 不得连续出现 2 个以上空行

## 注释规范

**包注释：**
```go
// Package service 提供游戏管理相关的业务逻辑实现。
package service
```

**导出类型/函数注释（必须）：**
```go
// Game 游戏服务，提供游戏的增删改查和密钥管理功能。
type Game struct {
    service.Service
}

// GetPage 获取游戏分页列表，支持按名称模糊搜索和租户隔离。
// tenantId 为 0 时返回所有租户数据（超级管理员）。
func (e *Game) GetPage(c *dto.GameGetPageReq, p *actions.DataPermission, list *[]models.Game, count *int64) error {
```

**行内注释：**
- 复杂业务逻辑、特殊处理必须添加注释说明原因
- 注释使用中文，与代码保持同步
- 禁止无意义注释，如 `// 获取用户` 对应 `getUser()`

**TODO/FIXME：**
```go
// TODO(username): 后续需要加缓存优化
// FIXME: 并发场景下可能有竞态条件
```

## 错误处理规范

**不得忽略错误：**
```go
// 错误示例（严禁）
os.Remove(tmpFile)  // 忽略错误

// 正确示例
if err := os.Remove(tmpFile); err != nil {
    log.Errorf("删除临时文件失败: %v", err)
}
```

**错误包装（使用 `fmt.Errorf` + `%w`）：**
```go
// 正确：保留错误链
if err := db.Create(&data).Error; err != nil {
    return fmt.Errorf("创建游戏失败: %w", err)
}

// 调用方可以用 errors.Is / errors.As 检查
if errors.Is(err, gorm.ErrRecordNotFound) {
    return ErrNotFound
}
```

**自定义业务错误：**
```go
// 定义错误变量
var (
    ErrGameNotFound    = errors.New("游戏不存在")
    ErrTenantMismatch  = errors.New("无权访问该租户数据")
    ErrInvalidAppKey   = errors.New("AppKey 格式非法")
)

// 使用
if game.TenantId != tenantId {
    return fmt.Errorf("%w: gameId=%d", ErrTenantMismatch, gameId)
}
```

**提前返回（减少嵌套）：**
```go
// 错误示例（嵌套过深）
func processGame(game *models.Game) error {
    if game != nil {
        if game.Status == 1 {
            if game.TenantId > 0 {
                // 处理逻辑
            }
        }
    }
    return nil
}

// 正确示例（提前返回）
func processGame(game *models.Game) error {
    if game == nil {
        return ErrGameNotFound
    }
    if game.Status != 1 {
        return fmt.Errorf("游戏状态异常: status=%d", game.Status)
    }
    if game.TenantId <= 0 {
        return ErrTenantMismatch
    }
    // 处理逻辑
    return nil
}
```

## 日志规范

**使用项目统一 logger：**
```go
// 使用 go-admin-core 的 logger
import log "github.com/go-admin-team/go-admin-core/logger"

// 或在 service 中使用 e.Log
e.Log.Errorf("db error: %s", err)
e.Log.Infof("创建游戏成功: gameId=%d", game.Id)
```

**禁止使用 `fmt.Println`：**
- 严禁在任何环境使用 `fmt.Println` 或 `fmt.Printf` 输出日志
- 调试时使用 `log.Debug()`，完成后及时清理

**日志级别选择：**

| 级别 | 使用场景 |
|------|----------|
| `Error` | 系统异常、需要立即处理的错误 |
| `Warn` | 潜在问题、降级处理、重试等 |
| `Info` | 关键业务节点、接口调用记录、状态变更 |
| `Debug` | 开发调试信息，生产环境不输出 |

**日志格式要求：**
- 使用格式化字符串，包含业务标识（如 gameId、tenantId）
- 记录异常时必须包含错误信息：`log.Errorf("操作失败: %v", err)`
- 关键操作需记录操作人和操作对象

## 并发规范

**goroutine 必须有退出机制：**
```go
// 正确：使用 context 控制退出
go func(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            // 处理逻辑
        }
    }
}(ctx)
```

**共享状态必须加锁：**
```go
type Cache struct {
    mu   sync.RWMutex
    data map[string]interface{}
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    v, ok := c.data[key]
    return v, ok
}
```

**使用 `sync.Once` 实现单例：**
```go
var (
    instance *Service
    once     sync.Once
)

func GetInstance() *Service {
    once.Do(func() {
        instance = &Service{}
    })
    return instance
}
```

## 接口设计规范

**接口应小而专一：**
```go
// 好的设计：小接口
type GameReader interface {
    GetGame(id int) (*models.Game, error)
}

type GameWriter interface {
    CreateGame(game *models.Game) error
    UpdateGame(game *models.Game) error
}

// 组合接口
type GameRepository interface {
    GameReader
    GameWriter
}
```

**接口定义在使用方（消费者）：**
```go
// service 层定义自己需要的接口
// 而不是在 repository 层定义接口让 service 依赖
type gameRepo interface {
    FindById(id int) (*models.Game, error)
}
```
