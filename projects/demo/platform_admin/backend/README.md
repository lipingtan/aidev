# Platform Admin 项目启动指南

本项目包含三个子应用：后端服务、管理端前端（dev-web-admin）、C端前端（dev-web-user）。

---

## 环境要求

| 依赖 | 版本 |
|------|------|
| Go | 1.24+ |
| Node.js | 18+ |
| MySQL | 5.7+ / 8.0+ |

---

## 目录结构

```
projects/demo/platform_admin/
├── backend/          ← Go 后端服务
├── dev-web-admin/    ← 管理端前端（Vue3）
├── dev-web-user/     ← C端前端（Vue3）
└── e2e/              ← Playwright E2E 测试
```

---

## 1. 启动后端

```bash
# 进入后端目录
cd projects/demo/platform_admin/backend

# 启动（开发模式）
go run main.go server -c config/settings.yml
```

后端启动后监听：`http://localhost:8000`

**首次启动**若数据库未初始化，访问 `http://localhost:8000/setup` 完成初始化向导。

---

## 2. 启动管理端前端（dev-web-admin）

```bash
# 进入管理端目录
cd projects/demo/platform_admin/dev-web-admin

# 安装依赖（首次）
npm install

# 启动开发服务器
npm run dev
```

管理端前端启动后监听：`http://localhost:3000`

> 代理规则：`/api/*`、`/auth/*` 请求自动转发到 `http://localhost:8000`

---

## 3. 启动 C端前端（dev-web-user）

```bash
# 进入 C端目录
cd projects/demo/platform_admin/dev-web-user

# 安装依赖（首次）
npm install

# 启动开发服务器
npm run dev
```

C端前端启动后监听：`http://localhost:5174`

> 代理规则：`/api/*` 请求自动转发到 `http://localhost:8080`（需根据实际后端端口调整 vite.config.ts）

---

## 访问地址汇总

| 服务 | 地址 | 说明 |
|------|------|------|
| 后端 API | http://localhost:8000 | Go 服务 |
| 后端 Swagger | http://localhost:8000/swagger/admin/index.html | 接口文档 |
| 初始化向导 | http://localhost:8000/setup | 首次使用 |
| 管理端前端 | http://localhost:3000 | dev-web-admin |
| C端前端 | http://localhost:5174 | dev-web-user |

---

## 登录说明

### 管理端（dev-web-admin）

- **地址**：http://localhost:3000
- **登录方式**：用户名 + 密码 + 图形验证码
- **默认账号**：在初始化向导（`/setup`）中设置，或查看数据库 `admin_user` 表

### C端（dev-web-user）

- **地址**：http://localhost:5174
- **登录方式**：手机号 + 短信验证码 + 租户编码
- **开发环境**：验证码打印在后端控制台（mock 模式），无需真实短信
- **租户编码**：从管理端 http://localhost:3000 → 租户管理菜单查看，或直接查数据库 `admin_tenant` 表的 `tenant_code` 字段

---

## Docker 启动（可选）

```bash
cd projects/demo/platform_admin/backend

# 构建镜像
docker build -t go-admin:latest .

# docker-compose 启动
docker-compose up -d
```

---

## E2E 测试

```bash
cd projects/demo/platform_admin/e2e

# 安装依赖（首次）
npm install
npx playwright install chromium

# 执行 V2-CR5 测试
npx playwright test tests/v2-cr5-dual-user-auth.api.spec.ts --reporter=list

# 执行全部测试
npm test
```

> **注意**：执行 E2E 测试前，需先启动后端（端口 8000）和管理端前端（端口 3000）。

---

## 数据库配置

默认连接（`config/settings.yml`）：

```
mysql: root:root@tcp(127.0.0.1:3306)/admindb
```

可根据本地环境修改 `config/settings.yml` 中的 `database.source` 字段。

---

## 常用命令

```bash
# 编译后端
go build -o go-admin .

# 数据库迁移
go run main.go migrate -c config/settings.yml

# 查看后端版本
go run main.go version
```
