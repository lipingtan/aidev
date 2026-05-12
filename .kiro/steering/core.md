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
| `AIDOC/` | 所有过程文档（策划、设计、迭代记录、知识库） |
| `projects/` | 游戏引擎工程（每个游戏一个独立工程目录） |
| `assets_source/` | 外部资产源文件（素材包原始文件、AI 生成原图） |
| `.kiro/` | 工作流配置（steering 规范、specs 文件） |

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

### 工作流与规范（通用）

| 任务类型 | 必读文件 |
|----------|----------|
| **写代码前：多平台画质与性能定稿** | `AIDOC/global-info/knowledge/steering/pre-development-defaults.md` |
| **启动全新游戏项目** | `AIDOC/global-info/knowledge/steering/project-bootstrap.md` |
| 执行开发工作流（双轨并行） | `AIDOC/global-info/knowledge/steering/development-workflow.md` |
| 每步执行时的依赖检查和产出验证 | `AIDOC/global-info/knowledge/steering/execution-protocol.md` |
| 叙事/剧情策划工作流 | `AIDOC/global-info/knowledge/steering/narrative-workflow.md` |
| 2D/3D 分支路由规则 | `AIDOC/global-info/knowledge/steering/branch-routing.md` |
| 资产管线（导入/生成/集成）；多平台画质档、纹理分档 | `AIDOC/global-info/knowledge/steering/asset-pipeline.md` |
| 迭代优化（调试/性能/平衡） | `AIDOC/global-info/knowledge/steering/iteration-workflow.md` |
| 多平台性能/画质预算 | `AIDOC/global-info/knowledge/steering/performance-budget.md` |

### 引擎特定规范（Godot）

| 任务类型 | 必读文件 |
|----------|----------|
| Godot 引擎特定规范和约束 | `AIDOC/global-info/knowledge/steering/godot/godot-engine.md` |
| QualitySettings 实现与 API 定稿 | `AIDOC/global-info/knowledge/steering/godot/quality-settings-spec.md` |
| Godot 代码生成规范 | `AIDOC/global-info/knowledge/steering/godot/code-generation.md` |
| Godot 工作流补充（POST-CHECK等） | `AIDOC/global-info/knowledge/steering/godot/godot-workflow.md` |
| Godot 项目启动配置 | `AIDOC/global-info/knowledge/steering/godot/godot-bootstrap.md` |

### 引擎知识（Godot）

| 任务类型 | 必读文件 |
|----------|----------|
| 编写脚本代码 | `AIDOC/global-info/knowledge/engines/godot/gdscript-guide.md` |
| 选择/使用节点类型 | `AIDOC/global-info/knowledge/engines/godot/node-types.md` |
| 设计信号通信 | `AIDOC/global-info/knowledge/engines/godot/signal-system.md` |
| 组织场景树结构 | `AIDOC/global-info/knowledge/engines/godot/scene-tree.md` |
| 使用资源系统 | `AIDOC/global-info/knowledge/engines/godot/resource-system.md` |
| 配置物理/碰撞 | `AIDOC/global-info/knowledge/engines/godot/physics-system.md` |
| 渲染/着色器/视觉效果 | `AIDOC/global-info/knowledge/engines/godot/rendering-pipeline.md` |

### 游戏设计模式

| 任务类型 | 必读文件 |
|----------|----------|
| 实现状态机 | `AIDOC/global-info/knowledge/patterns/state-machine.md` |
| 实现组件模式 | `AIDOC/global-info/knowledge/patterns/component-pattern.md` |
| 实现观察者/事件系统 | `AIDOC/global-info/knowledge/patterns/observer-pattern.md` |
| 实现命令模式 | `AIDOC/global-info/knowledge/patterns/command-pattern.md` |
| 实现对象池 | `AIDOC/global-info/knowledge/patterns/object-pool.md` |
| 实现行为树 | `AIDOC/global-info/knowledge/patterns/behavior-tree.md` |

### RPG 系统模板

| 任务类型 | 必读文件 |
|----------|----------|
| 设计/实现属性系统 | `AIDOC/global-info/knowledge/templates/rpg/attribute-system.md` |
| 设计/实现技能系统 | `AIDOC/global-info/knowledge/templates/rpg/skill-system.md` |
| 设计/实现背包/容器系统 | `AIDOC/global-info/knowledge/templates/rpg/inventory-system.md` |
| 设计/实现对话系统 | `AIDOC/global-info/knowledge/templates/rpg/dialogue-system.md` |
| 设计/实现任务系统 | `AIDOC/global-info/knowledge/templates/rpg/quest-system.md` |
| 设计/实现存档系统 | `AIDOC/global-info/knowledge/templates/rpg/save-system.md` |
| 设计/实现战斗系统 | `AIDOC/global-info/knowledge/templates/rpg/combat-system.md` |

### NSFW 系统模板

| 任务类型 | 必读文件 |
|----------|----------|
| 设计/实现关系数值系统（好感/堕落/羞耻等） | `AIDOC/global-info/knowledge/templates/nsfw/relationship-stats.md` |
| 设计/实现 H-Scene 系统 | `AIDOC/global-info/knowledge/templates/nsfw/h-scene-system.md` |
| 设计/实现 CG Gallery / 回想系统 | `AIDOC/global-info/knowledge/templates/nsfw/cg-gallery.md` |

### 合规与发行

| 任务类型 | 必读文件 |
|----------|----------|
| 确认内容红线 / AI 生成 NSFW 内容前 | `AIDOC/global-info/knowledge/compliance/content-guidelines.md` |
| 平台上架（Steam/DLsite/itch.io）前 | `AIDOC/global-info/knowledge/compliance/platform-policies.md` |
| 规划 All-Ages vs R-18 双版本 / 打包配置 | `AIDOC/global-info/knowledge/compliance/multi-version-strategy.md` |

### 动作系统模板

| 任务类型 | 必读文件 |
|----------|----------|
| 实现角色控制器 | `AIDOC/global-info/knowledge/templates/action/character-controller.md` |
| 实现动画状态机 | `AIDOC/global-info/knowledge/templates/action/animation-fsm.md` |
| 实现连击系统 | `AIDOC/global-info/knowledge/templates/action/combo-system.md` |
| 实现闪避/格挡 | `AIDOC/global-info/knowledge/templates/action/dodge-block.md` |
| 实现锁定系统 | `AIDOC/global-info/knowledge/templates/action/lock-on.md` |
| 实现相机控制 | `AIDOC/global-info/knowledge/templates/action/camera-control.md` |

### 项目级资产

| 任务类型 | 必读文件 |
|----------|----------|
| 查看游戏设计文档 | `AIDOC/projects/{游戏名}/design/gdd.md` |
| 查看世界观设定 | `AIDOC/projects/{游戏名}/design/worldview.md` |
| 查看角色设定 | `AIDOC/projects/{游戏名}/design/characters/` |
| 查看关卡策划 | `AIDOC/projects/{游戏名}/design/levels/` |
| 查看技术架构 | `AIDOC/projects/{游戏名}/design/architecture.md` |
| 查看剧情大纲 | `AIDOC/projects/{游戏名}/narrative/outline.md` |
| 查看任务设计 | `AIDOC/projects/{游戏名}/narrative/quests/` |
| 查看对话脚本 | `AIDOC/projects/{游戏名}/narrative/dialogues/` |
| 查看制作进度 | `AIDOC/projects/{游戏名}/tracker.md` |

> **规则**：如果任务涉及多个类型，同时读取对应的多个文件。
> 知识库文件不存在时，标记为待创建并继续执行。
> 进入任何制作阶段前，必须执行 `execution-protocol.md` 的门控检查。
> **系统模板索引与知识库写作原则**：`AIDOC/global-info/knowledge/templates/README.md`、`AIDOC/global-info/knowledge/GAME-DEV-INDEX.md`（不重复引擎官方文档，只收架构约束与 best practice）。
