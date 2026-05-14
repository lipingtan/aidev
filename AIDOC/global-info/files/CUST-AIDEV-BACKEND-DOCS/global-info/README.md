# AIStart 项目启动与初始化 README

## 1. 适用范围

本文档用于本项目本地/演示环境快速启动，覆盖：

- 后端启动（`spmp-backend`）
- 前端启动（`spmp-web-pc`）
- 首次初始化流程
- 默认账号密码说明

---

## 2. 目录位置

- 后端：`src/spmp-backend`
- 前端：`src/spmp-web-pc`
- 数据库迁移脚本：`src/spmp-backend/src/main/resources/db/migration`

---

## 3. 环境要求

- JDK 8
- Maven 3.6+
- Node.js 18+（建议）+ npm
- MySQL 8.x
- Redis 6.x+

---

## 4. 后端启动（spmp-backend）

### 4.1 进入目录

```bash
cd src/spmp-backend
```

### 4.2 构建与启动命令（jar 包方式）

```bash
mvn package -DskipTests
java -jar target/spmp-backend-1.0.0-SNAPSHOT.jar
```

### 4.3 默认端口与配置（`application.yml`）

- 后端端口：`8080`
- 默认数据库：`jdbc:mysql://localhost:3306/spmp`
- 默认 DB 用户：`liping`
- 默认 DB 密码：`ss123456`
- 默认 Redis：`localhost:6379`，`database=0`
- 后端接口基址：`http://localhost:8080/api/v1`

> 说明：首次部署建议走初始化流程，初始化成功后以 `init-config.yml` 为准。

---

## 5. 前端启动（spmp-web-pc）

### 5.1 进入目录

```bash
cd src/spmp-web-pc
```

### 5.2 安装依赖（首次）

```bash
npm install
```

### 5.3 启动命令

```bash
npm run dev
```

- 前端端口：`3000`
- 开发环境 API 前缀：`/api/v1`（来自 `.env.development`）
- 前端访问地址：`http://localhost:3000`

---

## 6. 首次初始化流程

系统未初始化时，后端仅开放初始化相关接口（`/api/v1/init/*`），核心接口包括：

- `GET /api/v1/init/status`
- `POST /api/v1/init/test-connection`
- `POST /api/v1/init/execute`

### 6.1 操作步骤

1. 启动后端与前端；
2. 打开前端页面：`http://localhost:3000`（进入初始化页面 `InitPage`）；
3. 填写 MySQL 与 Redis 连接信息；
4. 点击“测试连接”；
5. 点击“执行初始化”；
6. 初始化完成后**手动重启后端服务**。

初始化接口访问地址示例（后端）：`http://localhost:8080/api/v1/init/status`

### 6.2 初始化产物

- `config/init-config.yml`：初始化后的连接配置（密码加密）
- `config/init.lock`：初始化完成标记

---

## 7. 默认账号密码

初始化脚本 `V2__init_user_data.sql` 会创建超级管理员：

- 用户名：`admin`
- 密码：`Spmp@2026`

> 若环境中已修改密码，请以实际数据库为准。

---

## 8. 常见问题

### 8.1 初始化成功后仍无法正常访问业务接口

请确认是否已重启后端。初始化完成后必须重启，系统才会加载新配置并退出最小启动模式。

### 8.2 前端请求接口失败

请确认：

- 前端是否运行在 `3000`
- 后端是否运行在 `8080`
- API 前缀是否为 `/api/v1`
- 访问地址是否正确（前端 `http://localhost:3000`，后端 `http://localhost:8080/api/v1`）

### 8.3 登录失败

请先确认使用默认管理员账号（`admin / Spmp@2026`），并排除数据库中密码已变更情况。
