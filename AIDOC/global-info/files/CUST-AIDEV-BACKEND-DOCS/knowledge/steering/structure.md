---
inclusion: always
---

# Project Structure

## 文件位置说明

本工作区采用**通用规则**与**项目信息**分离的组织方式：

| 文件类型 | 位置 | 说明 |
|---|---|---|
| **通用规则** | `.trae/rules/` | 可复用于多个项目的规则，自动加载 |
| **项目信息** | `CUST-AIDEV-BACKEND-DOCS/global-info/` | SPMP 项目特定的详细信息 |

### 项目信息文件位置

以下项目特定文件位于 `.trae/rules/` 目录（自动加载）：

| 文件 | 位置 | 内容 |
|---|---|---|
| product.md | `.trae/rules/product.md` | 产品概述、业务领域、用户角色 |
| project-context.md | `.trae/rules/project-context.md` | SPMP 项目上下文和规范 |
| project-overview.md | `.trae/rules/project-overview.md` | 项目概览 |

以下项目特定文件位于 `CUST-AIDEV-BACKEND-DOCS/global-info/` 目录（需要时主动读取）：

| 文件 | 位置 | 内容 |
|---|---|---|
| business-brief.md | `global-info/business-brief.md` | 业务背景、用户、挑战 |
| backend-tech-stack-components.md | `global-info/backend-tech-stack-components.md` | 框架版本、依赖、命令 |
| backend-tech-stack-structure.md | `global-info/backend-tech-stack-structure.md` | 包结构、命名约定 |
| backend-service-api-call.md | `global-info/backend-service-api-call.md` | Feign Client 定义、外部服务调用 |
| backend-code-analysis.md | `global-info/backend-code-analysis.md` | 后端代码分析报告 |
| info-tobe-clarified.md | `global-info/info-tobe-clarified.md` | 待澄清信息 |

---

## Workspace Root

```
/
├── src/                         # 源代码目录（前后端分离）
│   ├── spmp-backend/            # 后端单体应用（Spring Boot + DDD 模块化）
│   ├── spmp-web-pc/             # PC 端前端（Vue 3 + Element Plus）
│   └── spmp-web-h5/             # H5 端前端（Vue 3 + Vant）
├── CUST-AIDEV-BACKEND-DOCS/     # 后端 AI 开发文档
├── CUST-AIDEV-FRONTEND-DOCS/    # 前端 AI 开发文档
├── .trae/                       # Trae 配置（rules, skills）
└── .kiro/                       # Kiro 配置（steering, skills, hooks）
```

---

## Backend Docs (`CUST-AIDEV-BACKEND-DOCS/`)

```
CUST-AIDEV-BACKEND-DOCS/
├── global-info/                          # 项目全局信息（SPMP 特定）
│   ├── business-brief.md                 # 业务背景、用户、挑战
│   ├── backend-tech-stack-components.md  # 框架版本、依赖、命令
│   ├── backend-tech-stack-structure.md   # 包结构、命名约定
│   ├── backend-service-api-call.md       # Feign Client 定义、外部服务调用
│   ├── backend-code-analysis.md          # 后端代码分析报告
│   └── info-tobe-clarified.md            # 待澄清信息
├── logs/                                 # 交互日志和变更日志
│   ├── interaction-log.md                # 交互日志
│   └── change-log.md                     # 变更日志
├── domain/                               # 域规范目录
│   ├── {domain}-center/                  # 各域规范
│   │   ├── domain-share/                 # 域共享文档
│   │   │   ├── api.md                    # API 文档
│   │   │   ├── database.md               # 数据库设计
│   │   │   ├── flow.md                   # 业务流程
│   │   │   └── integration.md            # 系统集成
│   │   ├── spec/                         # 需求规范
│   │   │   ├── normalCR/                # 常规 CR 需求（默认存放位置）
│   │   │   │   └── CRxx-{需求名}/       # 各 CR 子目录
│   │   │   └── minorCR/                 # 小型 CR 需求
│   │   │       └── MinorCR-{N}-{需求名}/ # 各小型 CR 子目录
│   │   ├── vibe/                         # Vibe 开发目录
│   │   ├── change-log/                   # 变更日志
│   │   └── code-review/                  # 代码审查
│   ├── templates/                        # 模板目录
│   └── README.md
└── skills/
    └── README.md
```

---

## Frontend Docs (`CUST-AIDEV-FRONTEND-DOCS/`)

```
CUST-AIDEV-FRONTEND-DOCS/
├── global-info/
│   ├── business-brief.md                 # 前端业务上下文
│   ├── frontend-pages-route.md           # 路由结构、命名、导航模式
│   ├── frontend-page-with-api.md         # 页面 → API 映射表
│   └── frontend-code-analysis.md         # 前端代码分析报告
├── spec/                                 # 前端 CR 规范（同后端结构）
└── README.md
```

---

## Trae Config (`.trae/`)

```
.trae/
├── rules/                            # 通用规则（可复用于多个项目）
│   ├── structure.md                  # 项目结构文档（本文件）
│   ├── auto-sync.md                  # 自动同步规则（Trae → Kiro）
│   ├── git-best-practices.md         # Git 工作流规则
│   ├── java-conventions.md           # Java 编码规范
│   ├── java-standards.md             # Java 开发标准
│   ├── spring-boot-patterns.md       # Spring Boot 模式
│   ├── security-best-practices.md    # 安全编码规则
│   ├── automated-api-testing.md      # 自动化测试标准
│   ├── development-workflow.md       # 开发工作流模板
│   ├── task-representation.md       # 任务可执行表示规范（RDD 三要素）
│   ├── regression-guard.md          # 回归防护规范
│   ├── log-analyzer.md               # 日志分析器引用
│   ├── sonarqube-quality.md          # SonarQube 质量标准
│   ├── change-log.md                 # 变更日志标准
│   └── code-review.md                # 代码审查标准
├── skills/                           # 技能目录
└── settings/                         # MCP 和其他设置
    └── mcp.json
```

---

## Kiro Config (`.kiro/`)

`.kiro/` 目录是 `.trae/` 的同步副本，供使用 Kiro 工具的团队成员使用：

```
.kiro/
├── steering/                        # 从 .trae/rules/ 同步
├── skills/                          # 从 .trae/skills/ 同步（排除 interaction-logger）
├── hooks/                           # Kiro 钩子
└── settings/                        # 设置
```

**同步规则**：详见 `.trae/rules/auto-sync.md`

---

## Key Conventions

- CR naming: `CRxx-{需求名}` for normal CRs, `MinorCR-{N}-{需求名}` for minor CRs
- CR 目录层级：`spec/normalCR/CRxx-{需求名}/` 用于常规 CR（默认），`spec/minorCR/MinorCR-{N}-{需求名}/` 用于小型 CR
- All spec documents include a metadata header table (title, CR number, version, author, dates, status)
- Acceptance criteria use SHALL / WHEN-IF-THEN format in `requirements.md`
- `design.md` references `sql/` and `api/` subdirectories by relative links — do not inline SQL or API specs in design
- `tasks.md` organizes tasks into parallel groups (A, B, C...) with explicit dependency chains
- Each `Task-{N}.md` lists input documents, implementation scope by layer (Controller/Service/Mapper), acceptance criteria, and parallelism notes
- DDL must be executed before DML; both must run before service deployment
- Production DDL requires DBA approval and off-peak execution window
