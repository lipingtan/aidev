# Go 项目启动流程

## 概述

从零启动一个新 Go 后端项目的标准流程，确保目录结构规范、基础配置正确、可立即进入开发。

---

## 一、澄清问题列表

### 1.1 必填项

| # | 问题 | 格式 | 默认值 |
|---|------|------|--------|
| 1 | 项目名称（英文 kebab-case） | ≤32 字符 | 无（必填） |
| 2 | 项目类型 | Web API / CLI 工具 / 混合 | Web API |
| 3 | 基础框架 | Gin+GORM / Echo+Ent / 纯标准库 | Gin+GORM |
| 4 | 数据库 | MySQL / PostgreSQL / SQLite | MySQL |
| 5 | 是否需要多租户 | 是 / 否 | 是 |
| 6 | 是否需要前端 | 是（Vue/React）/ 否 | 是（Vue） |
| 7 | 认证方式 | JWT / Session / OAuth2 | JWT |
| 8 | 目标部署环境 | Docker / 裸机 / K8s | Docker |
| 9 | 一句话描述项目用途 | 自由文本 | 无（必填） |

### 1.2 可选项

| # | 问题 | 默认值 |
|---|------|--------|
| 10 | 是否需要文件上传 | 否 |
| 11 | 是否需要 WebSocket | 否 |
| 12 | 是否需要定时任务 | 否 |
| 13 | 是否需要消息队列 | 否 |
| 14 | Go 版本要求 | 1.21+ |

### 1.3 完成标准

- 必填项 1~9 全部有明确答案
- 结果写入 `AIDOC/project_doc/{项目名}/requirements.md` 的背景章节

---

## 二、目录创建

澄清完成后自动创建：

```
# 过程文档
AIDOC/project_doc/{项目名}/
├── requirements.md          # 需求文档（含澄清结果）
├── design.md                # 设计文档
├── tasks.md                 # 任务列表
├── changelog.md             # 变更日志
└── source/                  # 外部参考材料

# 工程代码
projects/{解决方案名}/{项目名}/
├── backend/
│   ├── app/                 # 业务域
│   ├── cmd/                 # 入口
│   ├── common/              # 公共组件
│   │   ├── middleware/
│   │   ├── models/
│   │   └── global/
│   ├── config/
│   │   ├── settings.example.yml
│   │   └── extend.go
│   ├── web/                 # 前端静态文件（如需）
│   ├── docs/                # Swagger
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── frontend/                # 前端工程（如需）
├── deploy/                  # 部署配置
│   ├── Dockerfile
│   └── docker-compose.yml
├── dist/                    # 构建产物（gitignore）
├── build.ps1                # 构建脚本
├── .gitignore
└── README.md
```

---

## 三、基础配置生成

### 3.1 go.mod 初始化

```bash
go mod init {module-path}
```

### 3.2 核心依赖安装

```bash
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/mysql
go get github.com/golang-jwt/jwt/v5
```

### 3.3 配置文件模板

生成 `config/settings.example.yml`，包含所有配置项和注释说明。

### 3.4 .gitignore 生成

按 `go-git-workflow.md` 中的规范生成。

---

## 四、首个接口验证

Phase 0 完成前必须验证：

```
1. go build ./... 零错误
2. 启动服务，GET /api/v1/health 返回 200
3. 数据库连接正常（如配置了数据库）
4. JWT 中间件可正常拦截未认证请求
```

---

## 五、完成检查清单

```markdown
## Phase 0 完成检查

### 澄清
- [ ] 必填项全部回答
- [ ] 结果写入 requirements.md

### 目录
- [ ] AIDOC/project_doc/{项目名}/ 创建完整
- [ ] projects/{项目名}/ 创建完整
- [ ] .gitignore 已配置

### 工程
- [ ] go.mod 初始化完成
- [ ] 核心依赖已安装
- [ ] main.go 可编译运行
- [ ] 配置文件模板已生成

### 验证
- [ ] go build ./... 零错误
- [ ] 健康检查接口可访问
- [ ] 用户确认项目方向正确
```

---

## 六、启动后下一步

```
✅ 项目 {项目名} 初始化完成！

下一步：
1. 编写第一个功能的 requirements.md
2. 按 go-development-workflow.md 进入 Design 阶段
```
