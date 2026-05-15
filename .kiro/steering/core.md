# AI 游戏开发工作台 — 核心规范

## 项目身份

本工作区是 **AI 辅助游戏开发工作台**，面向 ARPG + 动作混合类型游戏（3D 为主、2D 为辅），支持单人 + AI 深度协作模式。

核心产出物：
- 可运行的游戏工程代码
- 游戏设计文档（GDD、世界观、角色、关卡策划）
- 叙事文档（剧情大纲、任务设计、对话脚本）
- 技术架构文档（ECS 框架、DLC 系统、数值系统）

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

## 目录核心分离原则

| 目录 | 职责 |
|------|------|
| `AIDOC/global-info/` | 全局知识库、规范文件（所有领域共用） |
| `AIDOC/series/` | 短剧相关规范与过程文档（策划、剧本、分镜、制作记录） |
| `AIDOC/game_doc/` | 游戏开发过程文档（GDD、世界观、角色、关卡、叙事、制作进度） |
| `AIDOC/project_doc/` | Go 软件系统过程文档（需求、设计、任务、变更日志） |
| `projects/` | 所有工程代码（按解决方案分组，每个解决方案下含游戏引擎工程 + 配套服务端等） |
| `assets_source/` | 外部资产源文件（素材包原始文件、AI 生成原图） |
| `.kiro/` | 工作流配置（steering 规范、specs 文件） |

**目录使用规则：**
- 短剧制作的所有过程文档 → `AIDOC/series/{系列名}/`
- 游戏开发的所有过程文档 → `AIDOC/game_doc/{游戏名}/`
- Go 软件系统的所有过程文档 → `AIDOC/project_doc/{项目名}/`
- 所有工程代码 → `projects/{解决方案名}/{工程名}/`
- 解决方案命名：以创建工程时所属的项目名命名（英文 snake_case）
- 同一解决方案下可包含多个相关工程（如游戏工程 + 管理端 + DLC 包）

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

## 知识库索引

执行任务前，根据任务类型读取对应的规范文件：

**工作流导航（不确定从哪开始时读此文件）**：`AIDOC/global-info/knowledge/steering/workflow-navigator.md`

---

### 短剧/视频开发规范（独立体系）

| 任务类型 | 必读文件 |
|----------|----------|
| 短剧产品概述 | `AIDOC/global-info/knowledge/steering/video/product.md` |
| 短剧项目上下文和规范 | `AIDOC/global-info/knowledge/steering/video/project-context.md` |
| 短剧项目结构概览 | `AIDOC/global-info/knowledge/steering/video/project-overview.md` |
| 短剧目录结构规范 | `AIDOC/global-info/knowledge/steering/video/structure.md` |
| 视频引擎技术规格（KLING/WAN） | `AIDOC/global-info/knowledge/steering/video/tech.md` |
| 产出文件位置规范 | `AIDOC/global-info/knowledge/steering/video/spec-output-location.md` |
| 叙事/剧情策划工作流 | `AIDOC/global-info/knowledge/steering/video/narrative-workflow.md` |
| 叙事质量审查 | `AIDOC/global-info/knowledge/steering/video/narrative-quality-review.md` |
| 启动全新短剧系列 | `AIDOC/global-info/knowledge/steering/video/series-bootstrap-workflow.md` |
| **镜头制作工作流（逐镜头制作）** | `AIDOC/global-info/knowledge/steering/video/shot-production-workflow.md` |
| **视频迭代优化协议** | `AIDOC/global-info/knowledge/steering/video/iteration-protocol.md` |
| 高潮设计哲学 | `AIDOC/global-info/knowledge/steering/video/climax-design-philosophy.md` |
| 不可预测性设计 | `AIDOC/global-info/knowledge/steering/video/unpredictability-design.md` |
| 分镜/运镜规范 | `AIDOC/global-info/knowledge/steering/video/cinematography.md` |
| 节奏控制 | `AIDOC/global-info/knowledge/steering/video/pacing.md` |
| 角色一致性 | `AIDOC/global-info/knowledge/steering/video/character-consistency.md` |
| AI 视频提示词工程 | `AIDOC/global-info/knowledge/steering/video/prompt-engineering.md` |
| 负面提示词 | `AIDOC/global-info/knowledge/steering/video/negative-prompts.md` |
| 风格关键词 | `AIDOC/global-info/knowledge/steering/video/style-keywords.md` |

---

### 游戏开发规范（独立体系，Godot 引擎）

标注 `[common]` 的为通用游戏开发规范，适用于任何引擎；其余为 Godot 特定。

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

#### NSFW 系统模板

| 任务类型 | 必读文件 |
|----------|----------|
| 设计/实现关系数值系统（好感/堕落/羞耻等） | `AIDOC/global-info/knowledge/steering/game/templates/nsfw/relationship-stats.md` |
| 设计/实现 H-Scene 系统 | `AIDOC/global-info/knowledge/steering/game/templates/nsfw/h-scene-system.md` |
| 设计/实现 CG Gallery / 回想系统 | `AIDOC/global-info/knowledge/steering/game/templates/nsfw/cg-gallery.md` |

#### 动作系统模板

| 任务类型 | 必读文件 |
|----------|----------|
| 实现角色控制器 | `AIDOC/global-info/knowledge/steering/game/templates/action/character-controller.md` |
| 实现动画状态机 | `AIDOC/global-info/knowledge/steering/game/templates/action/animation-fsm.md` |
| 实现连击系统 | `AIDOC/global-info/knowledge/steering/game/templates/action/combo-system.md` |
| 实现闪避/格挡 | `AIDOC/global-info/knowledge/steering/game/templates/action/dodge-block.md` |
| 实现锁定系统 | `AIDOC/global-info/knowledge/steering/game/templates/action/lock-on.md` |
| 实现相机控制 | `AIDOC/global-info/knowledge/steering/game/templates/action/camera-control.md` |

#### 合规与发行

| 任务类型 | 必读文件 |
|----------|----------|
| 确认内容红线 / AI 生成 NSFW 内容前 | `AIDOC/global-info/knowledge/steering/game/compliance/content-guidelines.md` |
| 平台上架（Steam/DLsite/itch.io）前 | `AIDOC/global-info/knowledge/steering/game/compliance/platform-policies.md` |
| 规划 All-Ages vs R-18 双版本 / 打包配置 | `AIDOC/global-info/knowledge/steering/game/compliance/multi-version-strategy.md` |

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

### Go 软件项目规范（独立体系）

适用于 `projects/` 下的 Go 后端软件项目（如 game_server）。与游戏开发规范、短剧规范完全独立，互不引用。

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

#### Go 项目适用规则

- Go 软件系统过程文档 → `AIDOC/project_doc/{项目名}/`
- Go 软件系统工程代码 → `projects/{解决方案名}/{项目名}/`
- 全局规范文件 → `AIDOC/global-info/knowledge/steering/software/go/`

> **规则**：如果任务涉及多个类型，同时读取对应的多个文件。
> 知识库文件不存在时，标记为待创建并继续执行。
> 游戏开发进入任何制作阶段前，必须执行 `AIDOC/global-info/knowledge/steering/game/execution-protocol.md` 的门控检查。
> **系统模板索引与知识库写作原则**：`AIDOC/global-info/knowledge/steering/game/templates/README.md`、`AIDOC/global-info/knowledge/steering/game/GAME-DEV-INDEX.md`（不重复引擎官方文档，只收架构约束与 best practice）。
