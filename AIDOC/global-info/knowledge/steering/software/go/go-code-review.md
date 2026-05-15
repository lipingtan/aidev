# Go 代码审查规范

## 审查维度

| 维度 | 检查项 |
|------|--------|
| **代码规范** | 命名规范、gofmt 格式、注释完整性 |
| **安全性** | SQL 注入、敏感信息泄露、权限校验、租户隔离 |
| **性能** | N+1 查询、无索引查询、goroutine 泄漏、内存泄漏 |
| **可维护性** | 函数复杂度、模块耦合度、重复代码 |
| **错误处理** | 错误是否被处理、错误信息是否有意义 |
| **并发安全** | 共享状态是否加锁、goroutine 是否有退出机制 |

## 问题级别

| 级别 | 标识 | 说明 | 处理要求 |
|------|:----:|------|----------|
| 严重 | 🔴 | 安全漏洞、数据泄露、系统崩溃风险 | 必须立即修复 |
| 重要 | 🟠 | 功能缺陷、性能问题、租户隔离缺失 | 必须修复后才能合并 |
| 一般 | 🟡 | 代码规范问题、可读性问题 | 建议修复 |
| 建议 | 🟢 | 优化建议、最佳实践 | 可选修复 |

## Go 特有检查项

### 🔴 严重问题

**租户隔离缺失：**
```go
// 错误：查询未加租户过滤
db.Find(&list)

// 正确
if tenantId > 0 {
    db = db.Where("tenant_id = ?", tenantId)
}
db.Find(&list)
```

**goroutine 泄漏：**
```go
// 错误：goroutine 无法退出
go func() {
    for {
        doWork()  // 没有退出条件
    }
}()

// 正确
go func(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            doWork()
        }
    }
}(ctx)
```

**使用不安全的随机数：**
```go
// 错误：math/rand 不适合密钥生成
import "math/rand"
secret := rand.Int63()

// 正确：crypto/rand
import "crypto/rand"
b := make([]byte, 32)
rand.Read(b)
```

### 🟠 重要问题

**忽略错误：**
```go
// 错误
db.Create(&data)  // 忽略错误

// 正确
if err := db.Create(&data).Error; err != nil {
    return fmt.Errorf("创建失败: %w", err)
}
```

**N+1 查询：**
```go
// 错误：循环中查询数据库
for _, game := range games {
    db.Where("game_id = ?", game.Id).Find(&dlcs)
}

// 正确：批量查询
gameIds := make([]int, len(games))
for i, g := range games { gameIds[i] = g.Id }
db.Where("game_id IN ?", gameIds).Find(&dlcs)
```

**未释放资源：**
```go
// 错误
resp, err := http.Get(url)
// 忘记 resp.Body.Close()

// 正确
resp, err := http.Get(url)
if err != nil { return err }
defer resp.Body.Close()
```

### 🟡 一般问题

**函数过长（超过 50 行）：**
- 提取子函数，每个函数只做一件事

**错误信息不够具体：**
```go
// 不好
return errors.New("error")

// 好
return fmt.Errorf("获取游戏列表失败，tenantId=%d: %w", tenantId, err)
```

**魔法数字：**
```go
// 不好
if status == 2 { ... }

// 好
const StatusEnabled = 1
const StatusDisabled = 2
if status == StatusDisabled { ... }
```

## 审查模板

```markdown
# CR-{YYYY}-{QN}-{NNN} - {审查标题}

## 基本信息
| 项目 | 内容 |
|------|------|
| 审查日期 | YYYY-MM-DD |
| 审查范围 | [涉及的模块/文件] |
| 审查人 | [姓名] |

## 审查摘要
| 维度 | 评分 | 说明 |
|------|:----:|------|
| 代码规范 | ⭐⭐⭐⭐⭐ | |
| 安全性 | ⭐⭐⭐⭐⭐ | |
| 性能 | ⭐⭐⭐⭐⭐ | |
| 错误处理 | ⭐⭐⭐⭐⭐ | |
| 并发安全 | ⭐⭐⭐⭐⭐ | |

## 问题列表

### 🔴 严重问题
| 序号 | 文件 | 行号 | 问题描述 | 状态 |
|:----:|------|:----:|----------|:----:|

### 🟠 重要问题
| 序号 | 文件 | 行号 | 问题描述 | 状态 |
|:----:|------|:----:|----------|:----:|

### 🟡 一般问题
| 序号 | 文件 | 行号 | 问题描述 | 状态 |
|:----:|------|:----:|----------|:----:|

## 审查结论
- [ ] ✅ 通过，可以合并
- [ ] ⚠️ 有条件通过，需修复重要问题后合并
- [ ] ❌ 不通过，需重新审查
```

## 提交前检查清单

```bash
# 格式检查
gofmt -l .

# 静态分析
go vet ./...

# 编译检查
go build ./...

# 测试
go test ./...
```

- [ ] `gofmt` 无差异
- [ ] `go vet` 无警告
- [ ] `go build ./...` 零错误
- [ ] 无 🔴 严重问题
- [ ] 无 🟠 重要问题
- [ ] 租户隔离逻辑正确
- [ ] 错误处理完整
