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

## 目录核心分离原则

| 目录 | 职责 |
|------|------|
| `AIDOC/` | 所有过程文档（策划、设计、迭代记录、知识库） |
| `godot_projects/` | 游戏引擎工程（每个游戏一个独立工程目录） |
| `assets_source/` | 外部资产源文件（素材包原始文件、AI 生成原图） |
| `.kiro/` | 工作流配置（steering 规范、specs 文件） |

**禁止**将任何产出文件生成到 `.kiro/specs/` 目录下。

## 知识库索引

执行任务前，根据任务类型读取对应的规范文件：

### 工作流与规范（通用）

| 任务类型 | 必读文件 |
|----------|----------|
| **启动全新游戏项目** | `AIDOC/global-info/knowledge/steering/project-bootstrap.md` |
| 执行开发工作流（双轨并行） | `AIDOC/global-info/knowledge/steering/development-workflow.md` |
| 每步执行时的依赖检查和产出验证 | `AIDOC/global-info/knowledge/steering/execution-protocol.md` |
| 叙事/剧情策划工作流 | `AIDOC/global-info/knowledge/steering/narrative-workflow.md` |
| 2D/3D 分支路由规则 | `AIDOC/global-info/knowledge/steering/branch-routing.md` |
| 资产管线（导入/生成/集成） | `AIDOC/global-info/knowledge/steering/asset-pipeline.md` |
| 迭代优化（调试/性能/平衡） | `AIDOC/global-info/knowledge/steering/iteration-workflow.md` |

### 引擎特定规范（Godot）

| 任务类型 | 必读文件 |
|----------|----------|
| Godot 引擎特定规范和约束 | `AIDOC/global-info/knowledge/steering/godot/godot-engine.md` |
| Godot 代码生成规范 | `AIDOC/global-info/knowledge/steering/godot/code-generation.md` |
| Godot 工作流补充（POST-CHECK等） | `AIDOC/global-info/knowledge/steering/godot/godot-workflow.md` |
| Godot 项目启动配置 | `AIDOC/global-info/knowledge/steering/godot/godot-bootstrap.md` |

### 引擎知识

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
