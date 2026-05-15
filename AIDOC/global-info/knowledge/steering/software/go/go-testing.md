# Go 测试规范

## 测试分层

| 层级 | 覆盖范围 | 运行频率 | 依赖 |
|------|----------|----------|------|
| 单元测试 | 单个函数/方法 | 每次提交 | 无外部依赖（Mock） |
| 集成测试 | 模块间交互 | PR 合并前 | 测试数据库 |
| API 测试 | 完整接口链路 | 发布前 | 完整环境 |

---

## 单元测试规范

### 文件命名

- 测试文件与被测文件同目录：`game.go` → `game_test.go`
- 测试函数命名：`Test{函数名}_{场景}`

### 表驱动测试模板

```go
func TestCalculateDamage(t *testing.T) {
    tests := []struct {
        name     string
        baseAtk  float64
        defense  float64
        expected float64
    }{
        {"正常伤害", 100, 30, 70},
        {"防御大于攻击", 30, 100, 0},
        {"零防御", 100, 0, 100},
        {"零攻击", 0, 50, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CalculateDamage(tt.baseAtk, tt.defense)
            if result != tt.expected {
                t.Errorf("期望 %v, 实际 %v", tt.expected, result)
            }
        })
    }
}
```

### Mock 规范

使用接口 + 手写 Mock（小项目）或 `gomock`（大项目）：

```go
// 定义接口（在使用方）
type gameRepo interface {
    FindById(id int) (*models.Game, error)
}

// Mock 实现
type mockGameRepo struct {
    games map[int]*models.Game
}

func (m *mockGameRepo) FindById(id int) (*models.Game, error) {
    game, ok := m.games[id]
    if !ok {
        return nil, gorm.ErrRecordNotFound
    }
    return game, nil
}
```

---

## 集成测试规范

### 数据库测试

```go
func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatal(err)
    }
    db.AutoMigrate(&models.Game{})
    return db
}

func TestGameService_GetPage(t *testing.T) {
    db := setupTestDB(t)
    // 插入测试数据
    db.Create(&models.Game{Name: "测试游戏", TenantId: 1})

    svc := &Game{}
    svc.Orm = db

    var list []models.Game
    var count int64
    err := svc.GetPage(&dto.GameGetPageReq{}, 1, &list, &count)

    if err != nil {
        t.Fatalf("查询失败: %v", err)
    }
    if count != 1 {
        t.Errorf("期望 1 条, 实际 %d 条", count)
    }
}
```

### 测试数据规则

- 测试数据自包含，不依赖外部数据库状态
- 每个测试用例独立，不依赖执行顺序
- 使用 `t.Cleanup()` 清理资源

---

## 覆盖率要求

| 模块 | 最低覆盖率 | 说明 |
|------|-----------|------|
| Service 层 | 70% | 核心业务逻辑 |
| 公共中间件 | 80% | 影响所有请求 |
| 工具函数 | 90% | 纯函数易测试 |
| API 层 | 50% | 主要靠集成测试覆盖 |
| Model 层 | 不强制 | 主要是结构定义 |

### 查看覆盖率

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## 测试命令

```bash
# 运行所有测试
go test ./...

# 运行指定包
go test ./app/game/service/...

# 运行指定测试函数
go test -run TestCalculateDamage ./app/game/service/

# 带覆盖率
go test -cover ./...

# 竞态检测
go test -race ./...
```

---

## 禁止事项

- 禁止测试依赖网络请求（Mock 外部服务）
- 禁止测试依赖特定时间（注入 clock 接口）
- 禁止测试间共享状态（每个测试独立初始化）
- 禁止在测试中使用 `time.Sleep`（使用 channel 或条件等待）
