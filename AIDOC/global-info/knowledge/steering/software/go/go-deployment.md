# Go 部署与发布规范

## 构建

### 构建脚本（build.ps1）

```powershell
# 设置版本信息
$version = git describe --tags --always
$buildTime = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

# 构建后端
Set-Location backend
go build -ldflags "-X main.Version=$version -X main.BuildTime=$buildTime" -o ../dist/server.exe ./cmd/api/
Set-Location ..

# 构建前端（如有）
Set-Location frontend
npm run build
Copy-Item -Recurse dist ../backend/web/dist
Set-Location ..
```

### 构建检查清单

- [ ] `go build ./...` 零错误
- [ ] `go vet ./...` 无警告
- [ ] `go test ./...` 全部通过
- [ ] 前端构建成功（如有）
- [ ] 版本号已更新

---

## Docker 部署

### Dockerfile 模板

```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -o server ./cmd/api/

# 运行阶段
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/config/settings.example.yml ./config/settings.yml
EXPOSE 8000
CMD ["./server"]
```

### docker-compose.yml 模板

```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8000:8000"
    environment:
      - DB_DSN=user:pass@tcp(db:3306)/dbname?charset=utf8mb4&parseTime=True
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - db
    restart: unless-stopped

  db:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: ${MYSQL_DATABASE}
    volumes:
      - db_data:/var/lib/mysql
    ports:
      - "3306:3306"

volumes:
  db_data:
```

---

## 环境变量管理

| 变量 | 说明 | 必须 |
|------|------|:---:|
| `DB_DSN` | 数据库连接字符串 | ✅ |
| `JWT_SECRET` | JWT 签名密钥（≥32字符随机串） | ✅ |
| `APP_MODE` | 运行模式（dev/prod） | ✅ |
| `APP_PORT` | 监听端口 | 默认 8000 |
| `LOG_LEVEL` | 日志级别 | 默认 info |

**规则：**
- 开发环境使用 `.env` 文件（gitignore）
- 生产环境使用系统环境变量或 Docker secrets
- 禁止在代码或配置文件中硬编码敏感信息

---

## 发布检查清单

### 发布前

- [ ] 所有测试通过
- [ ] 代码审查通过
- [ ] changelog.md 已更新
- [ ] 版本号已更新（git tag）
- [ ] 数据库迁移脚本已准备（如有 DDL 变更）
- [ ] 配置变更已通知运维

### 发布时

- [ ] 备份当前数据库
- [ ] 执行数据库迁移
- [ ] 部署新版本
- [ ] 验证健康检查接口
- [ ] 验证核心功能（登录、主要业务接口）

### 发布后

- [ ] 监控错误日志 15 分钟
- [ ] 确认无异常报警
- [ ] 通知相关人员发布完成

---

## 回滚策略

| 问题级别 | 响应 | 操作 |
|----------|------|------|
| 启动失败 | 立即回滚 | 切换到上一版本镜像 |
| 核心功能异常 | 5分钟内回滚 | 切换镜像 + 回滚数据库（如需） |
| 非核心功能异常 | 评估后决定 | 热修复或下一版本修复 |

**数据库回滚规则：**
- DDL 变更必须提供回滚 SQL
- 数据迁移必须可逆
- 不可逆变更需要额外审批
