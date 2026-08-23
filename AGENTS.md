# AI 辅助项目开发工作台 — 核心规范（注入副本）

> 本文件由 DSH `agent-instructions` 自动注入系统提示词（工作区根目录 AGENTS.md 发现机制）。
> 内容镜像自 `.kiro/steering/core.md`（单一事实源，`.cursor/rules/core.mdc` 为旧版已废弃）——**修改规范时先改 `.kiro/steering/core.md`，再同步本文件**。

## 项目身份

本工作区是 **AI 辅助项目开发工作台**，支持多类型项目的 AI 深度协作开发。

当前支持的项目类型：

| 类型 | 说明 | 规范体系 |
|------|------|----------|
| **游戏开发** | ARPG + 动作混合类型（3D 为主、2D 为辅），Godot 引擎 | `steering/game/` |
| **AI 短剧/视频制作** | 从故事策划到视频引擎提示词的全流程 | `steering/video/` |
| **软件系统开发** | Go 后端服务（Web API、管理端等） | `steering/software/` |

核心产出物：
- 可运行的工程代码（游戏/后端/前端）
- 设计文档（GDD、架构、需求、API 设计）
- 叙事/策划文档（世界观、剧本、分镜、角色设定）
- 结构化提示词（AI 视频生成用）

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
- **文件写入精简**：写入文件的内容在不丢失规则、不造成误解的前提下也保持精简，避免冗余描述

## 产出物自我审查规范（强制）

所有产出内容在提交给用户确认前，必须经过自我审查：

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
| `AIDOC/series/` | 短剧/视频项目的过程文档（策划、剧本、分镜、制作记录） |
| `AIDOC/game_doc/` | 游戏项目的过程文档（GDD、世界观、角色、关卡、叙事） |
| `AIDOC/project_doc/` | 软件项目的过程文档（需求、设计、任务、变更日志） |
| `projects/` | 所有工程代码（按解决方案分组） |
| `assets_source/` | 外部资产源文件（素材包、AI 生成原图） |
| `.kiro/` | 工作流配置（steering 规范） |

**目录使用规则：**
- 短剧/视频过程文档 → `AIDOC/series/{系列名}/`
- 游戏过程文档 → `AIDOC/game_doc/{游戏名}/`
- 软件系统过程文档 → `AIDOC/project_doc/{项目名}/`
- 所有工程代码 → `projects/{解决方案名}/{工程名}/`
- 解决方案命名：以项目名命名（英文 snake_case）
- 同一解决方案下可包含多个相关工程（游戏 + 管理端 + DLC 包等）

**解决方案层示例：**
```
projects/
├── demo/                          # demo 解决方案
│   ├── demo_game/                 # Godot 游戏工程
│   ├── dlc_pack/                  # DLC 资源包
│   └── game_server/               # Go 管理端/服务端
│
└── my_rpg/                        # 另一个解决方案（示例）
    ├── my_rpg_game/               # 游戏工程
    └── my_rpg_server/             # 配套服务端
```

**`source/` 文件夹约定：**
每个项目过程文件夹下可放置 `source/` 子目录，用于存放外部参考材料、初始想法、讨论记录等输入素材。AI 基于 `source/` 中的内容生成项目所需的基础资料。

```
AIDOC/game_doc/{游戏名}/source/       # 游戏项目的外部参考
AIDOC/project_doc/{项目名}/source/    # Go 项目的外部参考
AIDOC/series/{系列名}/source/         # 短剧项目的外部参考
```

典型内容：需求描述、想法草稿、讨论截图、竞品分析、技术调研、第三方文档等。

**禁止**将任何产出文件生成到 `.kiro/specs/` 目录下。

## 工具使用规范

### 文件写入规则（重要）

**禁止**使用 PowerShell 的 `Set-Content`、`Out-File`、`-replace | Set-Content` 写入包含中文的文件。
PowerShell 在 Windows 上默认使用 GBK/GB2312 编码，会导致中文内容乱码损坏。

| 操作 | 正确工具 | 禁止方式 |
|------|----------|----------|
| 创建/写入文件 | `fs_write` 工具 | PowerShell Set-Content |
| 修改文件内容 | `str_replace` 工具 | PowerShell -replace \| Set-Content |
| 追加文件内容 | `fs_append` 工具 | PowerShell Add-Content |
| 创建目录 | PowerShell `New-Item -ItemType Directory` | — |
| 移动/重命名文件 | PowerShell `Move-Item` / `Rename-Item` | — |
| 读取文件验证 | PowerShell `Get-Content -Encoding UTF8` | — |

**规则**：PowerShell 只用于目录操作、文件移动/重命名、读取验证。所有文件内容写入必须通过 `fs_write` / `str_replace` / `fs_append`。

---

## 执行前强制检查（违反即为错误）

**任何开发任务开始前，必须按以下顺序操作，不得跳过：**

### Go 软件项目任务
1. 读取 `AIDOC/global-info/knowledge/steering/software/go/go-development-workflow.md`
2. 按规范流程执行：Requirement → Design → Task → 执行
3. 过程文档必须创建到 `AIDOC/project_doc/{项目名}/{功能名}/`，**禁止**创建到 `.kiro/specs/`
4. 文档格式必须符合规范（requirements.md 含用户故事/FR-N/WHEN-SHALL；design.md 含不变行为清单；tasks.md 含三要素）
5. 需求阶段必须先创建 `requirements_plan.md` 提出澄清问题，用户确认后再输出 `requirements.md`
6. 设计阶段必须先创建 `design_plan.md` 提出设计相关澄清问题，用户确认后再输出 `design.md`
7. **每个阶段切换时（Requirement → Design → Task → 执行），必须重新读取 `go-development-workflow.md` 中对应阶段的流程步骤，确认该阶段的产出物格式和前置条件，不得凭记忆行动**
8. **执行阶段编写测试时，必须遵循 `go-testing.md` 中「自测真实性强制规范」，禁止编写空测试/虚假测试/跳过验收场景的测试。违反即视为任务未完成。**

### 软件开发强制流程规范（对所有 AI 实体生效）

> **本规范对主代理、子代理、任何被委托执行软件开发任务的 AI 实体均具有约束力。**
> 子代理不得以"已由上级代理确认"为由跳过用户确认步骤。

**核心规则：**
- 复杂 CR（逻辑复杂/跨多模块/涉及 DDL+业务/安全相关/完整 CRUD/第三方集成）必须走完整流程
- 完整流程：需求计划 → 用户确认 → 需求 → 用户确认 → 设计计划 → 用户确认 → 设计 → 用户确认 → 任务 → 用户确认 → 执行
- **每个阶段产出物必须经用户明确确认后才能进入下一阶段，违反即回退**
- 不确定复杂度时按复杂 CR 处理
- 详细判定标准、门控规则、模板见 `go-development-workflow.md`

### 游戏开发任务
1. 读取 `AIDOC/global-info/knowledge/steering/game/execution-protocol.md` 执行门控检查
2. 读取对应功能的规范文件（见下方知识库索引）

### 通用规则
- **禁止**在未读规范的情况下直接创建文档或代码
- **禁止**将过程文档（requirements.md / design.md / tasks.md）放到 `.kiro/specs/` 目录
- `.kiro/specs/` 仅用于 Kiro 工作流内部控制文件（.config.kiro 等）
- **禁止**跳过 plan 文件直接输出正式文档（requirements_plan → requirements、design_plan → design）
- **禁止**子代理跳过门控确认步骤直接执行开发任务

---

## Kiro 资产索引（`.kiro/`）

DSH 无 Kiro 式自动激活机制，按以下规则手动激活：

**Skills**（任务匹配触发条件时，先读对应 `SKILL.md` 再按其流程执行）：

| 技能 | 触发场景 |
|------|----------|
| `.kiro/skills/story-design/` | 新短剧/电影创意从零策划（故事+分镜） |
| `.kiro/skills/create-task/` | 故事策划/分镜拆解为可执行制作任务 |
| `.kiro/skills/prompt-generation/` | 分镜完成后生成目标引擎提示词 |
| `.kiro/skills/refactor/` | 优化已有提示词质量/统一风格（KLING/WAN） |
| `.kiro/skills/write-changelog/` | 迭代优化后记录提示词变更 |
| `.kiro/skills/multi-role-review/` | "多角色review/三角色review/review一下"等 |
| `.kiro/skills/project-manager-agent/` | 项目管理类：进展/风险/里程碑/WBS/回顾 |
| `.kiro/skills/test-case-gen/` | "写测试用例/生成用例/test case" |
| `.kiro/skills/e2e-playwright/` | "跑e2e/回归测试/自动化测试"（Playwright） |
| `.kiro/skills/platform-admin-plugin-dev/` | platform_admin 插件开发/部署/升级 |
| `.kiro/skills/tech-graph/` | "画图/架构图/流程图/出图"（SVG+PNG） |
| `.kiro/skills/ui-ux-design/` | 页面构建/UI 风格/配色/字体/设计系统审查 |
| `.kiro/skills/pptx/` | PPT 创建与编辑（python-pptx + 校验脚本） |

**Hooks**（Kiro 事件驱动配置，DSH 中由我模拟触发）：

| Hook | 规则 |
|------|------|
| `reply-in-chinese` | 始终中文回复（已含于核心规范） |
| `consistency-check` | 提示词文件被编辑后：对照项目 `story/characters.md` + `style-bible.md` 检查角色关键词/风格锚定词一致性，只提醒不自动改 |
| `storyboard-to-prompt-reminder` | 分镜文件创建后：简短提醒可生成提示词，不自动生成 |
| `post-task-acceptance` | tasks.md 完成率 100% 时：立即执行三角色验收（业务专家/产品经理/架构师清单，打开实际代码验证），通过输出「✅ 三角色验收通过」 |

**Steering**：`.kiro/steering/core.md` = 本文件镜像源（单一事实源）。

---

## 知识库索引

执行任务前，根据任务类型和项目领域读取对应的规范文件。

**不确定从哪开始？** → 读 `AIDOC/global-info/knowledge/steering/workflow-navigator.md`

---

### 零、混合项目规范（同时含 Go 后端 + Godot App + 游戏工程的项目）

| 任务类型 | 必读文件 |
|----------|----------|
| **混合项目任何 CR**（CR 类型判定 + 激活层路由） | `AIDOC/global-info/knowledge/steering/hybrid-project-workflow.md` |
| **nova_arcade 项目专属约束**（项目特有 Constraints / 冒烟清单 / 里程碑路由） | `AIDOC/project_doc/nova_arcade/dev-workflow-override.md` |

**使用规则：**
- 执行 nova_arcade 任何 CR 前：先读 `hybrid-project-workflow.md` 判定 CR 类型，再读 `dev-workflow-override.md` 叠加项目专属约束
- 新增混合项目时：在 `AIDOC/project_doc/{项目名}/dev-workflow-override.md` 创建项目 override，引用 `hybrid-project-workflow.md` 通用规范

---

### 一、短剧/视频制作规范（`steering/video/`）

| 任务类型 | 必读文件 |
|----------|----------|
| 短剧产品概述 | `AIDOC/global-info/knowledge/steering/video/product.md` |
| 短剧项目上下文和规范 | `AIDOC/global-info/knowledge/steering/video/project-context.md` |
| 短剧项目结构概览 | `AIDOC/global-info/knowledge/steering/video/project-overview.md` |
| 短剧目录结构规范 | `AIDOC/global-info/knowledge/steering/video/structure.md` |
| 视频引擎技术规格（KLING/WAN） | `AIDOC/global-info/knowledge/steering/video/tech.md` |
| 产出文件位置规范 | `AIDOC/global-info/knowledge/steering/video/spec-output-location.md` |
| 叙事质量审查 | `AIDOC/global-info/knowledge/steering/video/narrative-quality-review.md` |
| 启动全新短剧系列 | `AIDOC/global-info/knowledge/steering/video/series-bootstrap-workflow.md` |
| **镜头制作工作流（逐镜头制作）** | `AIDOC/global-info/knowledge/steering/video/shot-production-workflow.md` |
| **元素就绪检查（制作前置）** | `AIDOC/global-info/knowledge/steering/video/element-readiness.md` |
| **视频迭代优化协议** | `AIDOC/global-info/knowledge/steering/video/iteration-protocol.md` |
| **后期剪辑与衔接规范** | `AIDOC/global-info/knowledge/steering/video/post-production.md` |
| 高潮设计哲学 | `AIDOC/global-info/knowledge/steering/video/climax-design-philosophy.md` |
| 不可预测性设计 | `AIDOC/global-info/knowledge/steering/video/unpredictability-design.md` |
| 分镜/运镜规范 | `AIDOC/global-info/knowledge/steering/video/cinematography.md` |
| 节奏控制 | `AIDOC/global-info/knowledge/steering/video/pacing.md` |
| 角色一致性 | `AIDOC/global-info/knowledge/steering/video/character-consistency.md` |
| **空间视觉一致性（场景 DNA 体系）** | `AIDOC/global-info/knowledge/steering/video/spatial-consistency.md` |
| AI 视频提示词工程 | `AIDOC/global-info/knowledge/steering/video/prompt-engineering.md` |
| 负面提示词 | `AIDOC/global-info/knowledge/steering/video/negative-prompts.md` |
| 风格关键词 | `AIDOC/global-info/knowledge/steering/video/style-keywords.md` |

---

### 二、游戏开发规范（`steering/game/` + `steering/game/godot/`）

通用规范适用于任何引擎；`godot/` 下为 Godot 特定。

#### 通用游戏规范 [common]

| 任务类型 | 必读文件 |
|----------|----------|
| **写代码前：多平台画质与性能定稿** | `AIDOC/global-info/knowledge/steering/game/pre-development-defaults.md` |
| **启动全新游戏项目** | `AIDOC/global-info/knowledge/steering/game/project-bootstrap.md` |
| **开发具体功能/模块的完整流程** | `AIDOC/global-info/knowledge/steering/game/feature-development-flow.md` |
| **任务拆解与三要素格式** | `AIDOC/global-info/knowledge/steering/game/task-representation.md` |
| **游戏类型蓝图（标准功能清单）** | `AIDOC/global-info/knowledge/steering/game/game-type-blueprints.md` |
| **体验基准参考（手感参数范围）** | `AIDOC/global-info/knowledge/steering/game/experience-benchmarks.md` |
| 执行开发工作流（双轨并行） | `AIDOC/global-info/knowledge/steering/game/development-workflow.md` |
| 每步执行时的依赖检查和产出验证 | `AIDOC/global-info/knowledge/steering/game/execution-protocol.md` |
| 2D/3D 分支路由规则 | `AIDOC/global-info/knowledge/steering/game/branch-routing.md` |
| 资产管线（导入/生成/集成） | `AIDOC/global-info/knowledge/steering/game/asset-pipeline.md` |
| 迭代优化（调试/性能/平衡） | `AIDOC/global-info/knowledge/steering/game/iteration-workflow.md` |
| 多平台性能/画质预算 | `AIDOC/global-info/knowledge/steering/game/performance-budget.md` |
| RPG 叙事策划工作流（对话/任务/触发器） | `AIDOC/global-info/knowledge/steering/game/narrative-workflow.md` |

#### Godot 引擎特定

| 任务类型 | 必读文件 |
|----------|----------|
| Godot 引擎特定规范和约束 | `AIDOC/global-info/knowledge/steering/game/godot/godot-engine.md` |
| QualitySettings 实现与 API 定稿 | `AIDOC/global-info/knowledge/steering/game/godot/quality-settings-spec.md` |
| Godot 代码生成规范 | `AIDOC/global-info/knowledge/steering/game/godot/code-generation.md` |
| Godot 工作流补充（POST-CHECK等） | `AIDOC/global-info/knowledge/steering/game/godot/godot-workflow.md` |
| Godot 项目启动配置 | `AIDOC/global-info/knowledge/steering/game/godot/godot-bootstrap.md` |

#### 引擎知识（Godot）

> 以下知识已整合到 `godot-engine.md` 和 `code-generation.md` 中，不单独维护空文件。
> 遵循 GAME-DEV-INDEX.md 原则：不重复官方文档，只收架构约束与 best practice。

| 任务类型 | 必读文件 |
|----------|----------|
| GDScript 编码规范 / 类型标注 / 文件组织 | `AIDOC/global-info/knowledge/steering/game/godot/code-generation.md` |
| Godot 引擎约束（版本/渲染/物理/Autoload） | `AIDOC/global-info/knowledge/steering/game/godot/godot-engine.md` |
| 信号使用规范 / 何时用信号 vs 直接调用 | `AIDOC/global-info/knowledge/steering/game/godot/code-generation.md` 第八节 |
| 碰撞层分配 / 场景树组织 | `AIDOC/global-info/knowledge/steering/game/godot/code-generation.md` 第三、五节 |
| Shader 注释规范 | `AIDOC/global-info/knowledge/steering/game/godot/code-generation.md` 第四节 |
| 导出变量 / 异步操作 / 错误处理 | `AIDOC/global-info/knowledge/steering/game/godot/code-generation.md` 第六、七节 |

#### 游戏设计模式

| 任务类型 | 必读文件 |
|----------|----------|
| 实现状态机 | `AIDOC/global-info/knowledge/steering/game/patterns/state-machine.md` |

> 其余模式（观察者、命令、对象池、行为树、组件）在实际开发需要时按 state-machine.md 格式创建。

#### RPG 系统模板

| 任务类型 | 必读文件 |
|----------|----------|
| 设计/实现属性系统 | `AIDOC/global-info/knowledge/steering/game/templates/rpg/attribute-system.md` |
| 设计/实现技能系统 | `AIDOC/global-info/knowledge/steering/game/templates/rpg/skill-system.md` |
| 设计/实现背包/容器系统 | `AIDOC/global-info/knowledge/steering/game/templates/rpg/inventory-system.md` |
| 设计/实现对话系统 | `AIDOC/global-info/knowledge/steering/game/templates/rpg/dialogue-system.md` |
| 设计/实现任务系统 | `AIDOC/global-info/knowledge/steering/game/templates/rpg/quest-system.md` |
| 设计/实现存档系统 | `AIDOC/global-info/knowledge/steering/game/templates/rpg/save-system.md` |
| 设计/实现战斗系统 | `AIDOC/global-info/knowledge/steering/game/templates/rpg/combat-system.md` |

#### NSFW 规范（独立目录，非 NSFW 项目忽略）

> 详见 `AIDOC/global-info/knowledge/steering/game/nsfw/README.md`

| 任务类型 | 必读文件 |
|----------|----------|
| NSFW 项目完整规范索引 | `AIDOC/global-info/knowledge/steering/game/nsfw/README.md` |

#### 动作系统模板

| 任务类型 | 必读文件 |
|----------|----------|
| 实现角色控制器 | `AIDOC/global-info/knowledge/steering/game/templates/action/character-controller.md` |
| 实现动画状态机 | `AIDOC/global-info/knowledge/steering/game/templates/action/animation-fsm.md` |
| 实现连击系统 | `AIDOC/global-info/knowledge/steering/game/templates/action/combo-system.md` |
| 实现闪避/格挡 | `AIDOC/global-info/knowledge/steering/game/templates/action/dodge-block.md` |
| 实现锁定系统 | `AIDOC/global-info/knowledge/steering/game/templates/action/lock-on.md` |
| 实现相机控制 | `AIDOC/global-info/knowledge/steering/game/templates/action/camera-control.md` |

#### 项目级资产

| 任务类型 | 必读文件 |
|----------|----------|
| 查看游戏设计文档 | `AIDOC/game_doc/{游戏名}/design/gdd.md` |
| 查看世界观设定 | `AIDOC/game_doc/{游戏名}/design/worldview.md` |
| 查看角色设定 | `AIDOC/game_doc/{游戏名}/design/characters/` |
| 查看关卡策划 | `AIDOC/game_doc/{游戏名}/design/levels/` |
| 查看技术架构 | `AIDOC/game_doc/{游戏名}/design/architecture.md` |
| 查看剧情大纲 | `AIDOC/game_doc/{游戏名}/narrative/outline.md` |
| 查看任务设计 | `AIDOC/game_doc/{游戏名}/narrative/quests/` |
| 查看对话脚本 | `AIDOC/game_doc/{游戏名}/narrative/dialogues/` |
| 查看制作进度 | `AIDOC/game_doc/{游戏名}/tracker.md` |

### 三、软件系统开发规范（`steering/software/go/`）

适用于 Go 后端软件项目。后续可扩展 Java/Python 等语言。

#### Go 开发规范

| 任务类型 | 必读文件 |
|----------|----------|
| **启动全新 Go 项目** | `AIDOC/global-info/knowledge/steering/software/go/go-project-bootstrap.md` |
| **Go 编码规范**（命名/格式/注释/错误处理/日志） | `AIDOC/global-info/knowledge/steering/software/go/go-conventions.md` |
| **Go 项目目录结构**（分层架构/模块划分/数据库规范） | `AIDOC/global-info/knowledge/steering/software/go/go-project-structure.md` |
| **Go 软件开发工作流**（需求→设计→任务→执行→Bugfix） | `AIDOC/global-info/knowledge/steering/software/go/go-development-workflow.md` |
| **Go 任务三要素**（Scope/Constraints/Acceptance 详细定义） | `AIDOC/global-info/knowledge/steering/software/go/go-task-representation.md` |
| **Go 回归防护**（不变行为清单、回归验证规范） | `AIDOC/global-info/knowledge/steering/software/go/go-regression-guard.md` |
| **Go API 设计规范**（URL/分页/响应格式/DTO） | `AIDOC/global-info/knowledge/steering/software/go/go-api-design.md` |
| **Go 测试规范**（单元测试/集成测试/覆盖率/Mock） | `AIDOC/global-info/knowledge/steering/software/go/go-testing.md` |
| **Go 部署与发布**（构建/Docker/环境变量/发布检查清单） | `AIDOC/global-info/knowledge/steering/software/go/go-deployment.md` |
| **Go 前后端联调**（接口对接/CORS/Token/联调检查清单） | `AIDOC/global-info/knowledge/steering/software/go/go-frontend-integration.md` |
| **Go Git 工作流**（分支策略/Commit规范/PR流程） | `AIDOC/global-info/knowledge/steering/software/go/go-git-workflow.md` |
| **Go 安全编码**（认证/SQL安全/敏感数据/输入验证） | `AIDOC/global-info/knowledge/steering/software/go/go-security.md` |
| **Go 代码审查**（审查维度/问题级别/检查清单） | `AIDOC/global-info/knowledge/steering/software/go/go-code-review.md` |
| **Go 排查方法论**（排查优先级/高频陷阱/经验教训） | `AIDOC/global-info/knowledge/steering/software/go/go-debugging.md` |
| **Go 插件开发**（插件工程结构/打包规范/前端bundle/构建脚本） | `AIDOC/global-info/knowledge/steering/software/go/go-plugin-development.md` |

#### Go 项目适用规则

- Go 软件系统过程文档 → `AIDOC/project_doc/{项目名}/`
- Go 软件系统工程代码 → `projects/{解决方案名}/{项目名}/`
- 全局规范文件 → `AIDOC/global-info/knowledge/steering/software/go/`

> **规则**：如果任务涉及多个类型，同时读取对应的多个文件。
> 知识库文件不存在时，标记为待创建并继续执行。
> 游戏开发进入任何制作阶段前，必须执行 `AIDOC/global-info/knowledge/steering/game/execution-protocol.md` 的门控检查。
> **系统模板索引与知识库写作原则**：`AIDOC/global-info/knowledge/steering/game/templates/README.md`、`AIDOC/global-info/knowledge/steering/game/GAME-DEV-INDEX.md`（不重复引擎官方文档，只收架构约束与 best practice）。
