# AI 辅助项目开发工作台 — 核心规范

## 项目身份

本工作区是 **AI 辅助项目开发工作台**，支持多类型项目的 AI 深度协作开发。

当前支持的项目类型：

| 类型 | 说明 | 规范体系 |
|------|------|----------|
| **游戏开发** | ARPG + 动作混合类型（3D 为主、2D 为辅），Godot 引擎 | `steering/game/` |
| **AI 短剧/视频制作** | 从故事策划到视频引擎提示词的全流程 | `steering/video/` |
| **软件系统开发** | Go 后端服务（Web API、管理端等） | `steering/software/` |

## 语言规范

- **所有沟通、文档、回复必须使用中文**
- 代码注释使用中文
- 代码本身（变量名、函数名、类名）使用英文
- 文件/目录命名使用英文 snake_case

## 输出规范（强制）

- **极简输出**：回复必须简洁明了，直接给出结果，禁止输出思考过程、推理链、内心独白
- 不解释"为什么这样做"，除非用户明确要求
- 不列举多个方案供选择，除非用户明确要求
- 不输出"让我来分析一下"、"首先我们需要"等过渡性废话
- 列表/要点仅在用户明确要求"列出关键点"时使用
- 执行类任务：只报告完成状态和关键变更，不复述操作步骤
- **文件写入精简**：写入文件的内容在不丢失规则、不造成误解的前提下也保持精简

## 产出物自我审查规范（强制）

| 维度 | 检查要点 |
|------|----------|
| 准确性 | 数据/逻辑/引用是否正确 |
| 完整性 | 是否覆盖所有需求点 |
| 专业性 | 技术方案合理，无明显缺陷 |
| 格式规范 | 符合项目文档结构规范 |
| 一致性 | 命名风格与项目规范对齐 |

## 目录核心分离原则

| 目录 | 职责 |
|------|------|
| `AIDOC/global-info/` | 全局知识库与规范（`knowledge/steering/` 下按领域分类） |
| `AIDOC/series/` | 短剧/视频项目的过程文档 |
| `AIDOC/game_doc/` | 游戏项目的过程文档 |
| `AIDOC/project_doc/` | 软件项目的过程文档 |
| `projects/` | 所有工程代码（按解决方案分组） |
| `assets_source/` | 外部资产源文件 |
| `.trae/` | 工作流配置（rules 规范） |

**目录使用规则：**
- 短剧/视频过程文档 → `AIDOC/series/{系列名}/`
- 游戏过程文档 → `AIDOC/game_doc/{游戏名}/`
- 软件系统过程文档 → `AIDOC/project_doc/{项目名}/`
- 所有工程代码 → `projects/{解决方案名}/{工程名}/`

**禁止**将任何产出文件生成到 `.trae/` 内部目录下。

## 工具使用规范

**禁止**使用 PowerShell 的 `Set-Content`、`Out-File` 写入包含中文的文件（GBK 乱码问题）。

| 操作 | 正确方式 |
|------|----------|
| 创建/写入文件 | IDE 文件写入工具 |
| 修改文件内容 | IDE str_replace 工具 |
| 追加文件内容 | IDE append 工具 |
| 创建目录 | PowerShell `New-Item -ItemType Directory` |

---

## 执行前强制检查（违反即为错误）

### Go 软件项目任务
1. 读取 `AIDOC/global-info/knowledge/steering/software/go/go-development-workflow.md`
2. 按规范流程执行：Requirement → Design → Task → 执行
3. 过程文档必须创建到 `AIDOC/project_doc/{项目名}/{功能名}/`
4. 需求阶段：先创建 `requirements_plan.md` 提出澄清 → 用户确认 → 输出 `requirements.md`
5. 设计阶段：先创建 `design_plan.md` 提出设计澄清 → 用户确认 → 输出 `design.md`
6. 每个阶段切换时必须重新读取 `go-development-workflow.md` 对应阶段步骤

**软件开发强制流程（对所有 AI 实体生效）：**
- 复杂 CR 必须走完整流程：需求计划 → 确认 → 需求 → 确认 → 设计计划 → 确认 → 设计 → 确认 → 任务 → 确认 → 执行
- 每个阶段产出物必须经用户明确确认后才能进入下一阶段
- 子代理不得跳过用户确认步骤

### 游戏开发任务
1. 读取 `AIDOC/global-info/knowledge/steering/game/execution-protocol.md` 执行门控检查
2. 读取对应功能的规范文件

### 通用规则
- **禁止**在未读规范的情况下直接创建文档或代码
- **禁止**跳过 plan 文件直接输出正式文档

---

## 知识库索引

**不确定从哪开始？** → 读 `AIDOC/global-info/knowledge/steering/workflow-navigator.md`

### 零、混合项目规范

| 任务类型 | 必读文件 |
|----------|----------|
| 混合项目任何 CR | `AIDOC/global-info/knowledge/steering/hybrid-project-workflow.md` |
| nova_arcade 项目专属约束 | `AIDOC/project_doc/nova_arcade/dev-workflow-override.md` |

### 一、短剧/视频制作规范

| 任务类型 | 必读文件 |
|----------|----------|
| 短剧产品概述 | `AIDOC/global-info/knowledge/steering/video/product.md` |
| 短剧项目上下文和规范 | `AIDOC/global-info/knowledge/steering/video/project-context.md` |
| 短剧目录结构规范 | `AIDOC/global-info/knowledge/steering/video/structure.md` |
| 视频引擎技术规格（KLING/WAN） | `AIDOC/global-info/knowledge/steering/video/tech.md` |
| 叙事质量审查 | `AIDOC/global-info/knowledge/steering/video/narrative-quality-review.md` |
| 启动全新短剧系列 | `AIDOC/global-info/knowledge/steering/video/series-bootstrap-workflow.md` |
| 镜头制作工作流 | `AIDOC/global-info/knowledge/steering/video/shot-production-workflow.md` |
| 元素就绪检查 | `AIDOC/global-info/knowledge/steering/video/element-readiness.md` |
| 视频迭代优化协议 | `AIDOC/global-info/knowledge/steering/video/iteration-protocol.md` |
| 后期剪辑与衔接规范 | `AIDOC/global-info/knowledge/steering/video/post-production.md` |
| 分镜/运镜规范 | `AIDOC/global-info/knowledge/steering/video/cinematography.md` |
| 角色一致性 | `AIDOC/global-info/knowledge/steering/video/character-consistency.md` |
| 空间视觉一致性 | `AIDOC/global-info/knowledge/steering/video/spatial-consistency.md` |
| AI 视频提示词工程 | `AIDOC/global-info/knowledge/steering/video/prompt-engineering.md` |

### 二、游戏开发规范

#### 通用游戏规范

| 任务类型 | 必读文件 |
|----------|----------|
| 写代码前：多平台画质与性能定稿 | `AIDOC/global-info/knowledge/steering/game/pre-development-defaults.md` |
| 启动全新游戏项目 | `AIDOC/global-info/knowledge/steering/game/project-bootstrap.md` |
| 开发具体功能/模块的完整流程 | `AIDOC/global-info/knowledge/steering/game/feature-development-flow.md` |
| 任务拆解与三要素格式 | `AIDOC/global-info/knowledge/steering/game/task-representation.md` |
| 游戏类型蓝图 | `AIDOC/global-info/knowledge/steering/game/game-type-blueprints.md` |
| 执行开发工作流 | `AIDOC/global-info/knowledge/steering/game/development-workflow.md` |
| 执行门控检查 | `AIDOC/global-info/knowledge/steering/game/execution-protocol.md` |

#### Godot 引擎特定

| 任务类型 | 必读文件 |
|----------|----------|
| Godot 引擎特定规范和约束 | `AIDOC/global-info/knowledge/steering/game/godot/godot-engine.md` |
| Godot 代码生成规范 | `AIDOC/global-info/knowledge/steering/game/godot/code-generation.md` |
| Godot 工作流补充 | `AIDOC/global-info/knowledge/steering/game/godot/godot-workflow.md` |
| Godot 项目启动配置 | `AIDOC/global-info/knowledge/steering/game/godot/godot-bootstrap.md` |

### 三、软件系统开发规范（Go）

| 任务类型 | 必读文件 |
|----------|----------|
| 启动全新 Go 项目 | `AIDOC/global-info/knowledge/steering/software/go/go-project-bootstrap.md` |
| Go 编码规范 | `AIDOC/global-info/knowledge/steering/software/go/go-conventions.md` |
| Go 项目目录结构 | `AIDOC/global-info/knowledge/steering/software/go/go-project-structure.md` |
| Go 软件开发工作流 | `AIDOC/global-info/knowledge/steering/software/go/go-development-workflow.md` |
| Go 任务三要素 | `AIDOC/global-info/knowledge/steering/software/go/go-task-representation.md` |
| Go 回归防护 | `AIDOC/global-info/knowledge/steering/software/go/go-regression-guard.md` |
| Go API 设计规范 | `AIDOC/global-info/knowledge/steering/software/go/go-api-design.md` |
| Go 测试规范 | `AIDOC/global-info/knowledge/steering/software/go/go-testing.md` |
| Go 部署与发布 | `AIDOC/global-info/knowledge/steering/software/go/go-deployment.md` |
| Go 安全编码 | `AIDOC/global-info/knowledge/steering/software/go/go-security.md` |
| Go 排查方法论 | `AIDOC/global-info/knowledge/steering/software/go/go-debugging.md` |
| Go 插件开发 | `AIDOC/global-info/knowledge/steering/software/go/go-plugin-development.md` |
