# Go 安全编码规范

## 认证与授权

**JWT 配置：**
- 密钥从环境变量读取，不硬编码
- Token 有效期生产环境不超过 24 小时
- 刷新 Token 有效期不超过 7 天
- 生产环境关闭 dev 模式（`mode: prod`）

**接口权限：**
- 所有业务接口必须经过 JWT 验证中间件
- 使用 Casbin 做接口级权限控制
- `admin` 角色绕过 Casbin，其他角色必须显式授权
- 公开接口（如 `/api/v1/app-config`）必须明确标注

**多租户隔离：**
```go
// 所有业务查询必须加租户过滤
if tenantId > 0 {
    db = db.Where("tenant_id = ?", tenantId)
}
// tenant_id=0 为超级管理员，可查看所有数据
```

## SQL 安全

**禁止 SQL 拼接：**
```go
// 错误示例（严禁）
db.Raw("SELECT * FROM game WHERE name = '" + name + "'")

// 正确示例（参数化查询）
db.Where("name LIKE ?", "%"+name+"%").Find(&list)
db.Raw("SELECT * FROM game WHERE name = ?", name).Scan(&result)
```

**GORM 安全使用：**
```go
// 使用 GORM 的结构化查询，避免原始 SQL
db.Model(&Game{}).
    Where("tenant_id = ? AND status = ?", tenantId, 1).
    Find(&list)
```

## 敏感数据处理

**密码存储：**
```go
// 使用 bcrypt，cost 不低于 10
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// 验证
err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(inputPassword))
```

**API 密钥：**
```go
// 生成随机密钥
func generateSecret() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil
}

// 返回时隐藏密钥
object.AppSecret = "***"  // 列表接口不返回明文
```

**敏感字段不序列化：**
```go
type SysUser struct {
    Password string `json:"-" gorm:"size:128"` // json:"-" 不序列化
    Salt     string `json:"-" gorm:"size:255"`
}
```

## 输入验证

**参数绑定必须验证：**
```go
// 使用 binding tag 做基础验证
type GameInsertReq struct {
    Name   string `json:"name" binding:"required,min=1,max=128"`
    AppKey string `json:"appKey" binding:"required,alphanum"`
    Status int    `json:"status" binding:"oneof=1 2"`
}

// 绑定失败时返回 400
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(400, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
    return
}
```

**业务层二次验证：**
```go
// 不信任前端传来的 tenant_id，从 JWT 中获取
tenantId := middleware.GetTenantId(c)  // 从 context 获取，不从请求参数读取
```

## 配置安全

**禁止硬编码密钥：**
```go
// 错误示例（严禁）
const jwtSecret = "my-secret-key"

// 正确示例
secret := config.JwtConfig.Secret  // 从配置文件读取
// 配置文件中使用环境变量
// jwt.secret: ${JWT_SECRET:default-dev-secret}
```

**生产环境检查清单：**
- [ ] `mode: prod`（关闭 Swagger、调试接口）
- [ ] JWT secret 使用随机生成的强密钥
- [ ] 数据库密码不在代码中
- [ ] 日志级别设为 `info` 或 `warn`
- [ ] HTTPS 已启用

## 文件上传安全

```go
// 限制文件类型
allowedExts := map[string]bool{".pck": true, ".zip": true}
ext := filepath.Ext(header.Filename)
if !allowedExts[ext] {
    c.JSON(400, gin.H{"code": 400, "msg": "不支持的文件类型"})
    return
}

// 限制文件大小
const maxFileSize = 500 * 1024 * 1024 // 500MB
if header.Size > maxFileSize {
    c.JSON(400, gin.H{"code": 400, "msg": "文件大小超过限制"})
    return
}

// 使用安全的文件路径（防止路径穿越）
filename := filepath.Base(header.Filename)  // 只取文件名，去掉路径
savePath := filepath.Join(uploadDir, filename)
```

## 代码质量安全检查

**必须通过的检查：**
```bash
# 静态分析
go vet ./...

# 安全扫描（推荐）
gosec ./...

# 依赖漏洞扫描
govulncheck ./...
```

**禁止的模式：**
- 使用 `math/rand` 生成密钥（应使用 `crypto/rand`）
- 使用 `MD5`/`SHA1` 存储密码（应使用 `bcrypt`）
- 在日志中输出密码、token、密钥
- 在错误信息中暴露内部实现细节
