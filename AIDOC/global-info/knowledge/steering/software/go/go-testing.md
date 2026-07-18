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
- 禁止在测试中使用 `time.Sleep`（使用 channel 或条件等待，TTL 过期测试除外且 TTL 必须 ≤200ms）

---

## 自测真实性强制规范（对所有 AI 实体生效）

> **本规范对主代理、子代理、任何被委托执行开发任务的 AI 实体均具有约束力。**
> 违反任何一条即视为任务未完成，必须回退重做。

### 核心原则

**测试必须证明业务逻辑按预期工作，而非仅证明代码能编译运行。**

### 强制要求

| # | 要求 | 违反示例 | 正确做法 |
|---|------|----------|----------|
| T-1 | 测试必须验证业务结果 | `assert err == nil` 后不检查返回值 | 断言返回数据的每个关键字段与预期一致 |
| T-2 | 测试必须覆盖核心分支 | 只测正常流程 | 同时覆盖异常分支（参数错误/权限不足/数据不存在/并发冲突） |
| T-3 | 测试必须验证副作用 | 调用 Create 后不查 DB | 调用后直接查 DB 验证记录存在且字段正确 |
| T-4 | 测试必须验证约束生效 | 测循环引用但不构造实际循环 | 构造 A→B→A 的循环后验证返回 ErrCyclicHierarchy |
| T-5 | 测试必须使用真实数据层 | Mock 掉所有 repo 方法 | 使用 SQLite 内存 DB + AutoMigrate 执行真实 SQL |
| T-6 | 测试必须验证级联行为 | 删除关联后不检查级联删除 | 验证级联目标表记录确实被删除 |
| T-7 | 测试数据不可硬编码预期 | `assert count == 1` 但未先插入数据 | 测试内自行准备数据，再验证操作结果 |

### 禁止的虚假测试模式

以下模式视为"虚假测试"，等同于未写测试：

```go
// ❌ 禁止：空测试
func TestXxx(t *testing.T) {
    t.Log("TODO")
}

// ❌ 禁止：只检查 nil error 不验证结果
func TestCreate(t *testing.T) {
    result, err := svc.Create(req)
    if err != nil {
        t.Fatal(err)
    }
    _ = result // 完全不验证 result 内容
}

// ❌ 禁止：Mock 返回预设值然后断言该值（自循环验证）
func TestGetUser(t *testing.T) {
    mockRepo.On("FindByID", 1).Return(&User{Name: "test"}, nil)
    user, _ := svc.GetUser(1)
    assert.Equal(t, "test", user.Name) // 只是验证 mock 返回值，未测试任何逻辑
}

// ❌ 禁止：跳过验收标准中的测试场景
// tasks.md 要求测试"乐观锁冲突返回错误"，但测试文件中没有此用例

// ❌ 禁止：使用 t.Skip() 跳过关键测试
func TestCriticalLogic(t *testing.T) {
    t.Skip("暂时跳过") // 除非标注 //go:build integration
}
```

### 正确的测试模式

```go
// ✅ 正确：验证业务结果 + 副作用 + 约束
func TestCreateUser_PasswordBcryptEncrypted(t *testing.T) {
    db := setupTestDB(t)
    svc := setupService(db)

    user, err := svc.CreateUser(&CreateUserRequest{
        Username: "test_user",
        Password: "plaintext123",
    })
    if err != nil {
        t.Fatalf("创建用户失败: %v", err)
    }

    // 验证返回值
    if user.ID == 0 {
        t.Fatal("期望返回有效 ID")
    }
    if user.Username != "test_user" {
        t.Fatalf("期望 username=test_user, 实际: %s", user.Username)
    }

    // 验证副作用：DB 中密码不是明文
    var dbUser model.User
    db.First(&dbUser, user.ID)
    if dbUser.Password == "plaintext123" {
        t.Fatal("密码以明文存储，安全漏洞")
    }

    // 验证约束：bcrypt 可逆验证
    err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte("plaintext123"))
    if err != nil {
        t.Fatalf("bcrypt 验证失败: %v", err)
    }
}

// ✅ 正确：验证异常分支返回正确错误码
func TestDeleteRole_HasUsers_Rejected(t *testing.T) {
    db := setupTestDB(t)
    // ... 准备数据：创建角色 + 绑定用户 ...

    err := svc.DeleteRole(roleID)

    // 验证错误类型和错误码
    authErr, ok := err.(*errors.AuthError)
    if !ok {
        t.Fatalf("期望 AuthError, 实际: %T", err)
    }
    if authErr.Code != errors.ErrRoleHasUsers {
        t.Fatalf("期望 code=%d, 实际: %d", errors.ErrRoleHasUsers, authErr.Code)
    }

    // 验证角色未被删除（副作用未发生）
    var role model.Role
    if err := db.First(&role, roleID).Error; err != nil {
        t.Fatal("角色不应被删除")
    }
}
```

### 验收标准对照规则

tasks.md 中 `自测（验证通过才算完成）` 列出的每个测试场景，**必须在测试文件中有对应的 Test 函数**。对照检查：

1. 逐条读取 tasks.md 中 `自测` 部分的测试点
2. 在 `_test.go` 文件中查找对应的 Test 函数
3. 验证 Test 函数确实按描述构造了场景并验证了预期行为
4. 缺少任何一个测试点 = 任务未完成
