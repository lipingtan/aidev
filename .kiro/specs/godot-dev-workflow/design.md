# Design Document

## Overview

本设计文档将 12 项需求转化为可实施的技术方案。设计分为两大部分：

- **Part A: 工作流规范体�?*（Req 1-10）�?�?Markdown 规范文件 + steering 配置形式落地
- **Part B: 技术架构体�?*（Req 11-12）�?�?Godot addon 插件代码形式落地

---

## Part A: 工作流规范体�?

### A1. 目录结构设计（Req 1�?

#### 根目录布局

```
/
├── AIDOC/                              # 所有过程文�?
�?  ├── global-info/                    # 跨项目共�?
�?  �?  ├── knowledge/                  # 知识库（Req 5�?
�?  �?  �?  ├── engines/godot/          # Godot 引擎知识
�?  �?  �?  ├── patterns/               # 游戏设计模式
�?  �?  �?  ├── templates/              # 系统模板（RPG/动作�?
�?  �?  �?  ├── cases/                  # 案例�?
�?  �?  �?  �?  ├── success/
�?  �?  �?  �?  └── failure/
�?  �?  �?  └── steering/               # 工作流规范文�?
�?  �?  └── templates/                  # 通用文档模板
�?  �?
�?  └── projects/                       # 按游戏隔�?
�?      └── {游戏名}/
�?          ├── README.md               # 项目概述
�?          ├── design/                 # 游戏设计文档
�?          �?  ├── gdd.md              # 游戏设计文档
�?          �?  ├── worldview.md        # 世界�?
�?          �?  ├── characters/         # 角色设定
�?          �?  ├── levels/             # 关卡策划
�?          �?  └── architecture.md     # 技术架�?
�?          ├── narrative/              # 叙事文档
�?          �?  ├── outline.md          # 剧情大纲
�?          �?  ├── quests/             # 任务设计
�?          �?  ├── dialogues/          # 对话脚本
�?          �?  └── triggers.md         # 触发条件
�?          ├── iterations/             # 迭代记录
�?          �?  ├── bugs/
�?          �?  ├── performance/
�?          �?  └── gameplay/
�?          └── tracker.md              # 进度追踪
�?
├── projects/                     # Godot 工程（每游戏独立�?
�?  └── {游戏名}/
�?      ├── project.godot
�?      ├── scenes/
�?      �?  ├── levels/                 # 关卡场景
�?      �?  ├── ui/                     # UI 场景
�?      �?  ├── characters/             # 角色场景
�?      �?  └── 2d/                     # 2D 场景（按需创建�?
�?      ├── scripts/
�?      �?  ├── core/                   # 核心系统
�?      �?  ├── gameplay/               # 玩法逻辑
�?      �?  ├── ui/                     # UI 逻辑
�?      �?  ├── narrative/              # 叙事系统
�?      �?  └── ai/                     # AI 行为
�?      ├── resources/                  # Godot 资源文件
�?      �?  ├── data/                   # 数据资源�?tres�?
�?      �?  └── themes/                 # UI 主题
�?      ├── assets/
�?      �?  ├── characters/             # 按对象自包含
�?      �?  �?  └── {角色名}/
�?      �?  �?      ├── model/
�?      �?  �?      ├── textures/
�?      �?  �?      ├── animations/
�?      �?  �?      └── audio/
�?      �?  ├── environments/           # 按场景自包含
�?      �?  �?  └── {场景名}/
�?      �?  ├── enemies/
�?      �?  �?  └── {敌人名}/
�?      �?  ├── items/
�?      �?  �?  └── {物品名}/
�?      �?  └── shared/                 # 共享资产按类�?
�?      �?      ├── shaders/
�?      �?      ├── fonts/
�?      �?      ├── particles/
�?      �?      ├── audio/              # 通用音效/BGM
�?      �?      └── ui/
�?      ├── addons/
�?      �?  ├── gd_ecs/                 # ECS 框架插件（Req 11�?
�?      �?  ├── dlc_manager/            # DLC 管理插件（Req 12�?
�?      �?  └── shared/                 # 跨项目共�?addon
�?      ├── autoload/                   # 全局单例
�?      └── dlc/                        # DLC 包存放目�?
�?
├── assets_source/                      # 外部资产源文�?
�?  └── {游戏名}/
�?      ├── 3d_packages/                # 购买�?3D 素材包原始文�?
�?      ├── 2d_prompts/                 # 2D AI 生成提示词和原图
�?      └── audio_raw/                  # 原始音频文件
�?
└── .kiro/                              # 工作流配�?
    ├── steering/                       # Steering 规范
    �?  └── core.md                     # 核心规范（改造后�?
    └── specs/                          # Spec 文件
```

### A2. 核心规范文件设计（core.md 改造）

改造后�?`.kiro/steering/core.md` 结构�?

```markdown
# Godot 游戏开发工作台 �?核心规范

## 项目身份
## 语言规范
## 目录核心分离原则
## 默认引擎配置（Godot 4.x�?
## 知识库索引（按任务类型映射必读文件）
## 双轨工作流规�?
## 门控机制规则
## AI 代码生成规则
## 2D/3D 分支规则
```

### A3. 知识库文件清单（Req 5�?

```
AIDOC/global-info/knowledge/
├── engines/godot/
�?  ├── gdscript-guide.md           # GDScript 4.x 语法规范
�?  ├── node-types.md               # 节点类型参�?
�?  ├── signal-system.md            # 信号系统
�?  ├── scene-tree.md               # 场景树架�?
�?  ├── resource-system.md          # 资源系统
�?  ├── physics-system.md           # 物理系统
�?  └── rendering-pipeline.md       # 渲染管线
├── patterns/
�?  ├── state-machine.md            # 状态机模式
�?  ├── component-pattern.md        # 组件模式
�?  ├── observer-pattern.md         # 观察者模�?
�?  ├── command-pattern.md          # 命令模式
�?  ├── object-pool.md              # 对象池模�?
�?  └── behavior-tree.md            # 行为树模�?
├── templates/
�?  ├── rpg/
�?  �?  ├── attribute-system.md     # 属性系�?
�?  �?  ├── skill-system.md         # 技能系�?
�?  �?  ├── inventory-system.md     # 背包系统
�?  �?  ├── dialogue-system.md      # 对话系统
�?  �?  ├── quest-system.md         # 任务系统
�?  �?  ├── save-system.md          # 存档系统
�?  �?  └── combat-system.md        # 战斗系统
�?  └── action/
�?      ├── character-controller.md # 角色控制�?
�?      ├── animation-fsm.md        # 动画状态机
�?      ├── combo-system.md         # 连击系统
�?      ├── dodge-block.md          # 闪避/格挡
�?      ├── lock-on.md              # 锁定系统
�?      └── camera-control.md       # 相机控制
├── cases/
�?  ├── success/
�?  └── failure/
└── steering/
    ├── development-workflow.md      # 开发工作流（Req 2, 10�?
    ├── execution-protocol.md       # 执行协议（Req 3�?
    ├── code-generation.md          # 代码生成规范（Req 4�?
    ├── narrative-workflow.md        # 叙事工作流（Req 6�?
    ├── branch-routing.md           # 2D/3D 分支规范（Req 7�?
    ├── asset-pipeline.md           # 资产管线规范（Req 8�?
    ├── iteration-workflow.md       # 迭代优化规范（Req 9�?
    └── project-bootstrap.md        # 项目启动流程（Req 10�?
```

### A4. 工作流阶段设计（Req 2, 10�?

```
Phase 0: 项目初始�?
    ├── 创意轨道: 确认游戏类型/题材/平台
    └── 技术轨�? 创建目录骨架/Godot工程/技术选型
         �?
Phase 1: 核心设计
    ├── 创意轨道: GDD �?世界�?�?核心玩法设计
    └── 技术轨�? 技术架�?�?ECS框架配置 �?模块接口定义
         �?�?同步�? 世界观确�?�?触发架构设计
         �?
Phase 2: 原型验证
    ├── 创意轨道: 角色设计 �?关卡草案
    └── 技术轨�? 核心玩法原型 �?手感验证 �?AI行为验证
         �?�?同步�? 角色确认 �?触发角色系统
         �?
Phase 3: 核心系统实现
    ├── 创意轨道: 剧情大纲 �?支线设计 �?对话脚本
    └── 技术轨�? 角色系统 �?战斗系统 �?叙事系统 �?UI �?存档
         �?�?同步�? 关卡确认 �?触发场景搭建
         �?
Phase 4: 内容填充
    ├── 创意轨道: 关卡细化 �?敌人设计 �?道具设计
    └── 技术轨�? 关卡实现 �?敌人配置 �?道具数据 �?音效集成
         �?
Phase 5: 打磨优化
    ├── 创意轨道: 平衡性评�?�?体验优化建议
    └── 技术轨�? 性能优化 �?Bug修复 �?平衡调整 �?打包
```

### A5. 设计→代码生成节奏规�?

#### 核心原则

**先横向铺接口，再纵向填实现，�?可独立验�?为最小生成单元�?*

采用分层渐进（Layered Incremental）策略：

```
第一层：接口/骨架层（设计完一类就生成�?
    �? 设计完所有角�?�?生成角色系统接口 + 所有角色的数据骨架
    �? 设计完关卡结�?�?生成关卡加载框架 + 场景骨架
    �? 设计完道具列�?�?生成道具数据结构 + 注册�?
    �?
第二层：实现层（逐个填充�?
    �? 逐个角色：填充具体属性值、技能逻辑、动画配�?
    �? 逐个关卡：填充敌人配置、触发器、对�?
    �? 逐个道具：填充效果逻辑、视觉表�?
    �?
第三层：集成层（阶段性整合）
    �? 一个关卡的所有元素就�?�?集成测试 �?修复 �?确认
```

#### 各设计对象的生成时机

| 设计对象 | 生成时机 | 粒度 | 理由 |
|----------|----------|------|------|
| **系统框架** | 该类设计全部完成�?| 一次性生成整个系统的接口+骨架 | 确保接口一致�?|
| **角色个体** | 逐个设计完即生成 | 每个角色的数�?.tres)+专属逻辑 | 可独立验证，互不阻塞 |
| **关卡** | 单个关卡设计完整�?| 该关卡的全部配置一次性生�?| 关卡内元素耦合度高 |
| **道具/技�?* | 批量设计后批量生�?| 同类数据一起生�?| 保持格式一致�?|
| **3D 资产** | 逐个导入 | 每个资产单独适配 | 依赖外部工具，周期长 |
| **2D 资产** | 批量生成 | 同风格一起生�?| AI 批量处理效率更高 |

#### 角色设计→代码流�?

```
1. 设计所有角色的"共�?（属性结构、Component 定义�?
   �?生成：角色系统框架代码（ECS Component 基类、属性系统）
   �?验证：框架代码可编译、Component 可在 Inspector 中编�?

2. 逐个角色设计"个�?（具体数值、技能、行为）
   �?逐个生成：每个角色的数据文件(.tres) + 专属逻辑脚本
   �?每生成一个，在测试场景中验证基础行为
```

#### 关卡设计→代码流�?

```
1. 设计关卡�?结构"（区域划分、流程节点、胜利条件）
   �?生成：关卡场景骨�?.tscn) + 区域触发�?+ 流程控制脚本

2. 设计关卡�?内容"（敌人配置、道具放置、对话触发）
   �?整体生成：一个关卡设计完 �?一次性生成该关卡的全部配置数�?

3. 集成验证
   �?在编辑器中运行该关卡，验证流程完整�?
```

#### 资产设计→集成流�?

```
3D 资产：设计完一�?�?立即导入并配置（每个需单独适配碰撞/材质�?
2D 资产：批量设计提示词 �?批量生成 �?批量后处�?�?批量导入
```

#### 生成节奏与门控的关系

- 第一层（骨架）完�?�?触发架构设计 POST-CHECK
- 第二层（实现）每完成一个单�?�?触发单元�?POST-CHECK（语�?基础行为�?
- 第三层（集成）完�?�?触发集成测试 POST-CHECK（完整流程可跑通）

---

## Part B: 技术架构体�?

### B1. ECS 混合框架设计（Req 11�?

#### 核心类图

```
┌─────────────────────────────────────────────────────�?
�?EcsWorld (Autoload)                                  �?
├─────────────────────────────────────────────────────�?
�?- _entities: Dictionary[int, EcsEntity]              �?
�?- _systems: Array[EcsSystem]                         �?
�?- _component_registry: Dictionary[String, Script]    �?
�?- _query_cache: Dictionary[String, Array]            �?
├─────────────────────────────────────────────────────�?
�?+ register_entity(entity: EcsEntity) �?void          �?
�?+ unregister_entity(entity: EcsEntity) �?void        �?
�?+ register_system(system: EcsSystem) �?void          �?
�?+ unregister_system(system: EcsSystem) �?void        �?
�?+ register_component_type(name: String, script: Script) �?
�?+ query(components: Array[String]) �?Array[EcsEntity]�?
�?+ _process(delta) / _physics_process(delta)          �?
└─────────────────────────────────────────────────────�?
          �?管理                    �?调度
          �?                       �?
┌──────────────────�?   ┌──────────────────────────�?
�?EcsEntity (Node) �?   �?EcsSystem (RefCounted)    �?
├──────────────────�?   ├──────────────────────────�?
�?- _components:   �?   �?- priority: int           �?
�?  Dict[String,   �?   �?- phase: StringName       �?
�?  EcsComponent]  �?   �?- query: Array[String]    �?
├──────────────────�?   ├──────────────────────────�?
�?+ add_component()�?   �?+ process(entities, dt)   �?
�?+ remove_component�?  �?+ get_query() �?Array     �?
�?+ get_component()�?   └──────────────────────────�?
�?+ has_component()�?
└──────────────────�?
          �?持有
          �?
┌──────────────────────────�?
�?EcsComponent (Resource)   �?
├──────────────────────────�?
�?- component_name: String  �?
�?- (子类定义具体数据字段)    �?
├──────────────────────────�?
�?+ serialize() �?Dictionary�?
�?+ deserialize(data: Dict) �?
└──────────────────────────�?
```

#### 文件结构

```
addons/gd_ecs/
├── plugin.cfg                  # Godot 插件配置
├── plugin.gd                   # 插件入口
├── core/
�?  ├── ecs_world.gd            # World 管理器（Autoload�?
�?  ├── ecs_entity.gd           # Entity 基类（extends Node�?
�?  ├── ecs_component.gd        # Component 基类（extends Resource�?
�?  ├── ecs_system.gd           # System 基类（extends RefCounted�?
�?  └── ecs_query.gd            # Query 构建�?
├── debug/
�?  ├── ecs_debugger.gd         # 调试面板
�?  ├── entity_inspector.gd     # Entity 查看�?
�?  └── system_profiler.gd      # System 性能统计
└── examples/
    ├── health_component.gd     # 示例 Component
    ├── damage_system.gd        # 示例 System
    └── example_scene.tscn      # 示例场景
```

#### 关键设计决策

| 决策�?| 选择 | 理由 |
|--------|------|------|
| Entity 基类 | extends Node | 保留场景树集成、编辑器可视�?|
| Component 基类 | extends Resource | 支持 Inspector 编辑、序列化�?tres 存储 |
| System 基类 | extends RefCounted | 轻量、无节点开销、纯逻辑 |
| Query 机制 | 基于 Component 名称数组匹配 | 简单直观、支持动态注�?|
| 调度方式 | World �?_process/_physics_process 中遍�?| �?Godot 帧循环一�?|
| 动态注�?| Dictionary 注册�?+ 信号通知 | 支持 DLC 运行时注�?|
| 状态机集成 | 状态机作为 Entity 子节点，通过读写 Component 通信 | 行为决策与数据处理分�?|

#### 状态机 + ECS 集成设计

##### 职责划分

| �?| 负责什�?| 示例 |
|---|---|---|
| **状态机** | 单个角色的行为逻辑、状态转换条件、动画控�?| "血�?30%时从 Attack 切换�?Flee" |
| **ECS System** | 跨角色的批量计算、全局规则 | "所有有 PoisonBuff �?Entity 每秒�?5 血" |
| **ECS Component** | 纯数据存�?| HP、攻击力、Buff 列表、当前状态枚�?|

##### Entity 节点结构

```
Entity (EcsEntity extends Node)
├── StateMachine (Node)              �?Godot 原生节点，管理状态转�?
�?  ├── IdleState
�?  ├── RunState
�?  ├── AttackState
�?  ├── HitState
�?  └── DieState
├── Component: CharacterStats        �?ECS 纯数�?
├── Component: Velocity              �?ECS 纯数�?
├── Component: CombatState           �?ECS 纯数据（状态机写入，System 读取�?
└── Component: BuffList              �?ECS 纯数�?
```

##### 通信机制

```
状态机 ──写入──�?Component 数据 ←──读取── ECS System
                     �?
                     �?Component 是两者的"共享数据�?
                     �?
状态机读取 Component 判断转换条件
System 读取 Component 执行批量计算
```

##### CombatState Component 示例

```gdscript
class_name CombatStateComponent extends EcsComponent
## 战斗状态数据，由状态机写入，由 System 读取

@export var current_state: StringName = &"idle"
@export var state_time: float = 0.0
@export var is_invincible: bool = false
@export var combo_count: int = 0
```

##### 状态脚本示例（写入 Component�?

```gdscript
class_name AttackState extends State

func enter() -> void:
    var combat := entity.get_component("CombatState") as CombatStateComponent
    combat.current_state = &"attack"
    combat.is_invincible = true
    combat.combo_count += 1
    entity.get_node("AnimationPlayer").play("attack_01")

func exit() -> void:
    var combat := entity.get_component("CombatState") as CombatStateComponent
    combat.is_invincible = false
```

##### ECS System 示例（读�?Component 批量处理�?

```gdscript
class_name DamageSystem extends EcsSystem
## 处理所有收到伤害的 Entity

func get_query() -> Array[String]:
    return ["CharacterStats", "CombatState", "DamageEvent"]

func process(entities: Array, delta: float) -> void:
    for entity in entities:
        var stats := entity.get_component("CharacterStats")
        var combat := entity.get_component("CombatState")
        var damage := entity.get_component("DamageEvent")
        
        if combat.is_invincible:
            entity.remove_component("DamageEvent")
            continue
        
        stats.hp -= damage.amount
        entity.remove_component("DamageEvent")
```

##### 扩展新状态的流程（含 DLC 场景�?

```
扩展新状态（�?DLC 添加"暗影形�?）：
1. 新建 ShadowFormState 脚本 �?注册到角色的 StateMachine
2. 新建 ShadowComponent（暗影能量值等数据）→ 注册�?ECS
3. 可选：新建 ShadowDrainSystem（每秒消耗暗影能量）�?注册�?ECS

无需修改任何现有代码�?
- 状态机通过添加新状态节点扩�?
- ECS 通过注册�?Component + System 扩展
- DLC 通过 DlcManager 自动完成上述注册
```

#### 核心流程

```
游戏启动
    �?
    �?
EcsWorld._ready()
    ├── 加载已注册的 System（按 priority 排序�?
    └── 扫描场景树中�?EcsEntity 节点并注�?
    �?
    �?
每帧循环
    �?
    ├── EcsWorld._process(delta)
    �?  └── 遍历 phase="process" �?System
    �?      └── 对每�?System: query �?获取匹配 Entity �?system.process(entities, delta)
    �?
    └── EcsWorld._physics_process(delta)
        └── 遍历 phase="physics_process" �?System
            └── 同上
    �?
    �?
Entity 生命周期
    ├── 进入场景�?�?_enter_tree() �?EcsWorld.register_entity(self)
    └── 离开场景�?�?_exit_tree() �?EcsWorld.unregister_entity(self)
```

### B2. DLC 动态挂接系统设计（Req 12�?

#### DLC 包结�?

```
dlc/
└── {dlc_id}/
    ├── manifest.json           # DLC 清单
    ├── components/             # �?Component 定义
    �?  └── *.gd
    ├── systems/                # �?System 定义
    �?  └── *.gd
    ├── data/                   # 数据资源
    �?  ├── characters/         # 角色扩展数据
    �?  ├── items/              # 新道�?
    �?  ├── quests/             # 新任�?
    �?  └── dialogues/          # 新对�?
    ├── assets/                 # 新资�?
    �?  ├── models/
    �?  ├── textures/
    �?  ├── audio/
    �?  └── scenes/
    └── scripts/                # DLC 专用逻辑脚本
```

#### manifest.json 格式

```json
{
  "id": "dlc_dark_realm",
  "version": "1.0.0",
  "display_name": "暗域扩展�?,
  "description": "新增暗域地图�?个角色、暗影技能树",
  "min_game_version": "0.5.0",
  "max_game_version": "1.x",
  "dependencies": [],
  "conflicts": [],
  "content": {
    "components": ["components/shadow_affinity.gd"],
    "systems": ["systems/shadow_damage_system.gd"],
    "character_extensions": [
      {
        "target_entity": "warrior",
        "add_components": ["shadow_affinity"],
        "add_data": "data/characters/warrior_shadow.tres"
      }
    ],
    "assets": ["assets/"],
    "scenes": ["assets/scenes/dark_realm.tscn"],
    "quests": ["data/quests/shadow_quest_01.json"],
    "dialogues": ["data/dialogues/shadow_npc.json"]
  },
  "load_order": 100
}
```

#### 核心类图

```
┌─────────────────────────────────────────────────────�?
�?DlcManager (Autoload)                                �?
├─────────────────────────────────────────────────────�?
�?- _loaded_dlcs: Dictionary[String, DlcPackage]       �?
�?- _dlc_scan_paths: Array[String]                     �?
├─────────────────────────────────────────────────────�?
�?+ scan_dlcs() �?Array[DlcManifest]                   �?
�?+ load_dlc(dlc_id: String) �?Result                  �?
�?+ unload_dlc(dlc_id: String) �?Result                �?
�?+ get_loaded_dlcs() �?Array[String]                  �?
�?+ is_dlc_loaded(dlc_id: String) �?bool               �?
�?                                                     �?
�?signal dlc_loaded(dlc_id: String)                    �?
�?signal dlc_unloaded(dlc_id: String)                  �?
�?signal dlc_load_failed(dlc_id: String, error: String)�?
└─────────────────────────────────────────────────────�?
          �?管理
          �?
┌──────────────────────────────────�?
�?DlcPackage (RefCounted)           �?
├──────────────────────────────────�?
�?- manifest: DlcManifest           �?
�?- registered_components: Array    �?
�?- registered_systems: Array       �?
�?- registered_assets: Dictionary   �?
�?- entity_extensions: Array        �?
├──────────────────────────────────�?
�?+ apply() �?void                  �?
�?+ revert() �?void                 �?
└──────────────────────────────────�?
          �?解析
          �?
┌──────────────────────────────────�?
�?DlcManifest (Resource)            �?
├──────────────────────────────────�?
�?- id: String                      �?
�?- version: String                 �?
�?- min_game_version: String        �?
�?- dependencies: Array[String]     �?
�?- conflicts: Array[String]        �?
�?- content: Dictionary             �?
└──────────────────────────────────�?
```

#### DLC 加载流程

```
DlcManager.load_dlc(dlc_id)
    �?
    ├── 1. 读取 manifest.json �?解析�?DlcManifest
    �?
    ├── 2. 版本兼容性检�?
    �?  └── IF 不兼�?�?emit dlc_load_failed �?return
    �?
    ├── 3. 依赖检�?
    �?  └── IF 依赖未加�?�?递归加载依赖
    �?
    ├── 4. 冲突检�?
    �?  └── IF 存在冲突 �?emit dlc_load_failed �?return
    �?
    ├── 5. 注册 Components
    �?  └── 遍历 manifest.content.components
    �?      └── EcsWorld.register_component_type(name, script)
    �?
    ├── 6. 注册 Systems
    �?  └── 遍历 manifest.content.systems
    �?      └── EcsWorld.register_system(system_instance)
    �?
    ├── 7. 注册资产
    �?  └── 遍历 manifest.content.assets
    �?      └── AssetRegistry.register(path, dlc_id)
    �?
    ├── 8. 应用角色扩展
    �?  └── 遍历 manifest.content.character_extensions
    �?      └── 找到目标 Entity �?add_component(dlc_component)
    �?
    └── 9. emit dlc_loaded(dlc_id)
```

#### DLC 卸载流程

```
DlcManager.unload_dlc(dlc_id)
    �?
    ├── 1. 检查是否有其他 DLC 依赖�?DLC
    �?  └── IF �?�?先卸载依赖方
    �?
    ├── 2. 回退角色扩展
    �?  └── 遍历 entity_extensions �?remove_component
    �?
    ├── 3. 注销 Systems
    �?  └── EcsWorld.unregister_system(...)
    �?
    ├── 4. 注销 Components
    �?  └── EcsWorld.unregister_component_type(...)
    �?
    ├── 5. 注销资产
    �?  └── AssetRegistry.unregister(dlc_id)
    �?
    └── 6. emit dlc_unloaded(dlc_id)
```

#### 文件结构

```
addons/dlc_manager/
├── plugin.cfg
├── plugin.gd
├── core/
�?  ├── dlc_manager.gd          # DLC Manager（Autoload�?
�?  ├── dlc_package.gd          # DLC 包封�?
�?  ├── dlc_manifest.gd         # 清单解析
�?  ├── dlc_validator.gd        # 版本/依赖/冲突验证
�?  └── dlc_loader.gd           # 文件加载（目�?PCK�?
├── integration/
�?  ├── ecs_integration.gd      # �?ECS 框架的集成层
�?  └── asset_integration.gd    # 与资产注册表的集成层
└── examples/
    ├── example_dlc/             # 示例 DLC �?
    �?  ├── manifest.json
    �?  ├── components/
    �?  └── systems/
    └── dlc_test_scene.tscn
```

### B3. RPG 数值系统设计（属�?装备/技�?成长/容器�?

#### 系统总览

```
┌─────────────────────────────────────────────────────────────�?
�?RPG 数值系�?�?Component/System/Resource 分层                 �?
├─────────────────────────────────────────────────────────────�?
�?                                                             �?
�? Resource（策划配置数据）        Component（运行时可变数据�?    �?
�? ├── ItemData                  ├── BaseStats                 �?
�? ├── EquipmentStats            ├── FinalStats                �?
�? ├── SkillData                 ├── Equipment                 �?
�? ├── SkillEffect               ├── Inventory                 �?
�? ├── BuffData                  ├── SkillSet                  �?
�? ├── GrowthProfile             ├── BuffList                  �?
�? └── StatModifier              ├── Experience                �?
�?                               └── ItemContainer             �?
�?                                                             �?
�? System（计算规则）                                           �?
�? ├── StatsCalculationSystem    # 属性汇总计�?                �?
�? ├── LevelUpSystem             # 升级判定                     �?
�? ├── SkillCooldownSystem       # 技能冷却递减                 �?
�? ├── BuffTickSystem            # Buff 持续/过期处理           �?
�? └── ContainerSystem           # 容器间物品转�?              �?
�?                                                             �?
└─────────────────────────────────────────────────────────────�?
```

#### 属性计算管�?

```
触发重算（装备变�?升级/Buff变化/技能学习）
    �?
    �?设置 FinalStats.is_dirty = true
    �?
StatsCalculationSystem.process()
    �?
    ├── 1. 复制 BaseStats �?FinalStats（基础值）
    �?
    ├── 2. 收集所�?StatModifier 来源
    �?  ├── Equipment.slots �?每件装备�?stat_modifiers
    �?  ├── BuffList.active_buffs �?每个 Buff �?modifiers
    �?  └── SkillSet.learned_skills �?被动技能的 passive_modifiers
    �?
    ├── 3. 按类型分组计�?
    �?  ├── FLAT_ADD: 所有固定值求�?
    �?  ├── PERCENT_ADD: 所有百分比求和后乘�?
    �?  └── PERCENT_MULT: 逐个独立乘算
    �?
    ├── 4. 最终公�?
    �?  final = (base + flat_sum) * (1 + percent_add_sum) * percent_mult_product
    �?
    └── 5. FinalStats.is_dirty = false
```

#### 属性系�?

```gdscript
# ===== 属性修饰器 =====
class_name StatModifier extends Resource
## 一条属性修饰规�?

@export var stat_name: StringName      # 目标属性名（attack/defense/...�?
@export var mod_type: ModType          # 加成类型
@export var value: float               # 数�?
@export var source: StringName = &""   # 来源标识（用于移除）

enum ModType { 
    FLAT_ADD,        # 固定值加成：+50
    PERCENT_ADD,     # 百分比加成：+10%（可叠加�?
    PERCENT_MULT     # 百分比乘算：x1.1（独立乘算）
}


# ===== 基础属�?Component =====
class_name BaseStatsComponent extends EcsComponent
## 角色的基础属性值（由等级和成长曲线决定�?

@export var max_hp: float = 100.0
@export var max_mp: float = 50.0
@export var attack: float = 10.0
@export var defense: float = 5.0
@export var speed: float = 3.0
@export var crit_rate: float = 0.05
@export var crit_damage: float = 1.5


# ===== 最终属�?Component =====
class_name FinalStatsComponent extends EcsComponent
## 经过所有修饰器计算后的最终属�?

@export var max_hp: float = 0.0
@export var max_mp: float = 0.0
@export var attack: float = 0.0
@export var defense: float = 0.0
@export var speed: float = 0.0
@export var crit_rate: float = 0.0
@export var crit_damage: float = 0.0
@export var is_dirty: bool = true   # 标记是否需要重�?
```

#### 物品/装备系统

```gdscript
# ===== 物品定义（数据资源）=====
class_name ItemData extends Resource
## 物品的静态定义（策划配置，所有同类物品共享）

@export var id: StringName
@export var name: String
@export var icon: Texture2D
@export var item_type: ItemType
@export var rarity: int = 1
@export var stackable: bool = false
@export var max_stack: int = 1
@export var description: String
@export var sell_price: int = 0
@export var buy_price: int = 0

enum ItemType { WEAPON, ARMOR, ACCESSORY, CONSUMABLE, MATERIAL, QUEST }


# ===== 装备属性加�?=====
class_name EquipmentStats extends Resource
## 装备提供的属性修�?

@export var stat_modifiers: Array[StatModifier] = []
@export var equip_slot: EquipSlot
@export var required_level: int = 1

enum EquipSlot { WEAPON, HEAD, BODY, LEGS, FEET, ACCESSORY_1, ACCESSORY_2 }


# ===== 物品实例数据（区分同类物品的个体差异�?====
class_name ItemInstance extends Resource
## 物品的个体数据（同一把剑可能强化等级不同�?

@export var uid: String = ""
@export var enhance_level: int = 0
@export var enchantments: Array[StringName] = []
@export var durability: float = 100.0
@export var custom_data: Dictionary = {}


# ===== 装备�?Component =====
class_name EquipmentComponent extends EcsComponent
## 角色当前装备的物�?

@export var slots: Dictionary = {}  # { EquipSlot: ItemData }

func equip(item: ItemData, equip_stats: EquipmentStats) -> ItemData:
    var slot := equip_stats.equip_slot
    var old_item := slots.get(slot)
    slots[slot] = item
    return old_item

func unequip(slot: EquipSlot) -> ItemData:
    var item := slots.get(slot)
    slots.erase(slot)
    return item

func get_all_modifiers() -> Array[StatModifier]:
    var result: Array[StatModifier] = []
    for item in slots.values():
        var stats := item.get_meta("equipment_stats") as EquipmentStats
        if stats:
            result.append_array(stats.stat_modifiers)
    return result
```

#### 技能体�?

```gdscript
# ===== 技能定�?=====
class_name SkillData extends Resource
## 技能的静态定�?

@export var id: StringName
@export var name: String
@export var icon: Texture2D
@export var description: String
@export var skill_type: SkillType
@export var cooldown: float = 1.0
@export var mp_cost: float = 10.0
@export var required_level: int = 1
@export var damage_formula: String = "attack * 1.5"
@export var effects: Array[SkillEffect] = []
@export var passive_modifiers: Array[StatModifier] = []

enum SkillType { ACTIVE, PASSIVE, TOGGLE }


# ===== 技能效�?=====
class_name SkillEffect extends Resource
## 技能触发的效果

@export var effect_type: EffectType
@export var target: TargetType
@export var value: float
@export var duration: float = 0.0

enum EffectType { DAMAGE, HEAL, BUFF, DEBUFF, KNOCKBACK, STUN }
enum TargetType { SELF, SINGLE_ENEMY, AOE_CIRCLE, AOE_CONE, ALL_ALLIES }


# ===== 技能栏 Component =====
class_name SkillSetComponent extends EcsComponent
## 角色已学习的技能和冷却状�?

@export var learned_skills: Array[StringName] = []
@export var equipped_skills: Array[StringName] = []
@export var cooldowns: Dictionary = {}
@export var skill_levels: Dictionary = {}

func get_passive_modifiers() -> Array[StatModifier]:
    var result: Array[StatModifier] = []
    for skill_id in learned_skills:
        var skill := _load_skill(skill_id)
        if skill and skill.skill_type == SkillData.SkillType.PASSIVE:
            result.append_array(skill.passive_modifiers)
    return result
```

#### 成长体系

```gdscript
# ===== 成长曲线定义 =====
class_name GrowthProfile extends Resource
## 角色的成长曲线（每个角色/职业一份）

@export var stat_growth: Dictionary = {
    "max_hp": 15.0,
    "max_mp": 8.0,
    "attack": 3.0,
    "defense": 2.0,
    "speed": 0.5,
}
@export var exp_curve: Curve
@export var skill_unlock_table: Dictionary = {}  # { level: [skill_id] }


# ===== 经验/等级 Component =====
class_name ExperienceComponent extends EcsComponent
## 角色的经验和等级数据

@export var level: int = 1
@export var current_exp: float = 0.0
@export var total_exp: float = 0.0
@export var growth_profile: GrowthProfile


# ===== 升级 System =====
class_name LevelUpSystem extends EcsSystem

func get_query() -> Array[String]:
    return ["Experience", "BaseStats"]

func process(entities: Array, delta: float) -> void:
    for entity in entities:
        var exp := entity.get_component("Experience") as ExperienceComponent
        var required := exp.growth_profile.exp_curve.sample(exp.level)
        
        while exp.current_exp >= required and exp.level < 99:
            exp.current_exp -= required
            exp.level += 1
            _apply_level_up(entity, exp)
            required = exp.growth_profile.exp_curve.sample(exp.level)

func _apply_level_up(entity: EcsEntity, exp: ExperienceComponent) -> void:
    var base := entity.get_component("BaseStats") as BaseStatsComponent
    var growth := exp.growth_profile.stat_growth
    
    base.max_hp += growth.get("max_hp", 0.0)
    base.attack += growth.get("attack", 0.0)
    base.defense += growth.get("defense", 0.0)
    base.speed += growth.get("speed", 0.0)
    
    # 解锁新技�?
    var unlocks := exp.growth_profile.skill_unlock_table.get(exp.level, [])
    if unlocks.size() > 0 and entity.has_component("SkillSet"):
        var skills := entity.get_component("SkillSet") as SkillSetComponent
        for skill_id in unlocks:
            if skill_id not in skills.learned_skills:
                skills.learned_skills.append(skill_id)
    
    entity.get_component("FinalStats").is_dirty = true
```

#### 统一容器系统（背�?储物�?宝箱/商店/掉落物）

##### 核心思路

背包、储物柜、商店货架、宝箱、掉落物都是**物品容器**的不同实例，共享同一套数据结构和操作逻辑�?

##### 容器类型配置

| 容器类型 | 容量 | 过滤�?| 只读 | 持久�?| 特殊规则 |
|----------|------|--------|------|--------|----------|
| 背包 | 40（可扩展�?| �?| �?| 存档 | 负重限制（可选） |
| 装备�?| 7（固定槽位） | 按槽位类�?| �?| 存档 | 装备/卸下触发属性重�?|
| 储物�?| 100 | �?| �?| 存档 | 多角色共�?|
| 商店 | 无限 | �?| �?| 不存 | 有价格、可刷新库存 |
| 宝箱 | 1~10 | �?| �?| 场景状�?| 开启后标记已开 |
| 掉落�?| 1~5 | �?| �?| 不存 | 有过期时�?|
| 交易窗口 | 临时 | �?| �?| 不存 | 双方确认后执�?|

##### 数据结构

```gdscript
# ===== 物品槽位 =====
class_name ItemSlot extends Resource
## 容器中的一个格�?

@export var item: ItemData = null
@export var count: int = 0
@export var instance_data: ItemInstance = null


# ===== 物品容器 Component =====
class_name ItemContainerComponent extends EcsComponent
## 通用物品容器

@export var container_id: StringName = &""
@export var container_type: ContainerType = ContainerType.BACKPACK
@export var capacity: int = 40
@export var slots: Array[ItemSlot] = []
@export var filter: Array[ItemData.ItemType] = []
@export var is_readonly: bool = false
@export var owner_entity_id: int = -1

enum ContainerType {
    BACKPACK,
    EQUIPMENT,
    STORAGE,
    SHOP,
    LOOT,
    TRADE
}

func can_add(item: ItemData) -> bool:
    if is_readonly:
        return false
    if filter.size() > 0 and item.item_type not in filter:
        return false
    return _find_available_slot(item) >= 0

func add_item(item: ItemData, count: int = 1, instance: ItemInstance = null) -> int:
    # 返回实际添加的数�?
    # 先尝试堆叠到已有同类物品
    # 再尝试放入空�?
    pass

func remove_item(slot_index: int, count: int = 1) -> ItemSlot:
    # 从指定格子移除物品，返回移除的内�?
    pass

func sort() -> void:
    # 按物品类型和稀有度排序，合并可堆叠物品
    pass
```

##### 容器操作事件

```gdscript
# ===== 容器操作事件 Component =====
class_name ContainerActionComponent extends EcsComponent
## 容器间物品转移请求（�?UI 层生成，�?ContainerSystem 消费�?

@export var action_type: ActionType
@export var source_container_id: StringName
@export var source_slot: int
@export var target_container_id: StringName
@export var target_slot: int = -1  # -1 = 自动寻找空位
@export var count: int = 1

enum ActionType { MOVE, SPLIT, USE, DROP, SORT, BUY, SELL }


# ===== 容器操作 System =====
class_name ContainerSystem extends EcsSystem

func get_query() -> Array[String]:
    return ["ContainerAction"]

func process(entities: Array, delta: float) -> void:
    for entity in entities:
        var action := entity.get_component("ContainerAction") as ContainerActionComponent
        var result := _execute_action(action)
        entity.remove_component("ContainerAction")
        if not result.success:
            push_warning("容器操作失败: %s" % result.error)

func _execute_action(action: ContainerActionComponent) -> Dictionary:
    var source := _find_container(action.source_container_id)
    if not source:
        return { "success": false, "error": "源容器不存在" }
    
    match action.action_type:
        ActionType.MOVE:
            var target := _find_container(action.target_container_id)
            return _move_item(source, action.source_slot, target, action.target_slot, action.count)
        ActionType.USE:
            return _use_item(source, action.source_slot)
        ActionType.DROP:
            return _drop_item(source, action.source_slot, action.count)
        ActionType.SORT:
            source.sort()
            return { "success": true }
        ActionType.BUY:
            return _buy_item(source, action.source_slot, action.count)
        ActionType.SELL:
            return _sell_item(source, action.source_slot, action.count)
    return { "success": false, "error": "未知操作类型" }
```

##### 容器间交互流�?

```
玩家操作：从背包拖物品到储物�?
    �?
    �?
UI 层：生成 ContainerAction 事件
    �? action_type = MOVE
    �? source = 背包容器ID, slot = 3
    �? target = 储物柜容器ID, slot = 7
    �?
    �?
ContainerSystem.process()
    �?
    ├── 验证：source 格子有物品？
    ├── 验证：target 容器未满�?
    ├── 验证：target 过滤器允许该物品类型�?
    ├── 验证：source 非只读？target 非只读？
    �?
    ├── IF 目标格有同类可堆叠物�?�?合并堆叠
    ├── IF 目标格有不同物品 �?交换位置
    └── IF 目标格为�?�?直接移动
    �?
    �?
发出信号：container_changed(container_id)
    �?
    �?
UI 层：监听信号 �?刷新显示
属性系统：如果是装备栏变更 �?FinalStats.is_dirty = true
```

##### 储物�?Entity 示例

```
Entity: StorageChest (EcsEntity + Node3D)
├── Component: ItemContainer (type=STORAGE, capacity=100)
├── Component: Interactable (interaction_range=2.0, prompt="打开储物�?)
├── MeshInstance3D
└── CollisionShape3D

# 玩家交互时：
# 1. UI 打开双栏界面（左=背包，右=储物柜）
# 2. 拖拽操作生成 ContainerAction
# 3. ContainerSystem 处理转移
# 4. 存档时序列化 STORAGE 类型容器�?slots 数据
```

#### 系统间关系图

```
┌─────────────�?    装备变更信号     ┌──────────────────────�?
�?装备�?      �?──────────────────�?�?StatsCalculationSystem�?
�?(Equipment) �?                     �?(重算 FinalStats)     �?
└─────────────�?                     └──────────────────────�?
       �?物品转移                            �?
┌─────────────�?                     ┌──────────────────────�?
�?背包         �?── 使用经验物品 ──�? �?LevelUpSystem         �?
�?(Backpack)  �?                     �?(升级→属性成�?       �?
└─────────────�?                     └──────────────────────�?
       �?物品转移                            �?
┌─────────────�?                     ┌──────────────────────�?
�?储物�?      �?                     �?SkillCooldownSystem   �?
�?(Storage)   �?                     �?(技能CD递减)          �?
└─────────────�?                     └──────────────────────�?
       �?拾取                                �?
┌─────────────�?                     ┌──────────────────────�?
�?掉落�?宝箱  �?                     �?BuffTickSystem        �?
�?(Loot)      �?                     �?(Buff持续/过期)       �?
└─────────────�?                     └──────────────────────�?
       �?买卖
┌─────────────�?
�?商店         �?
�?(Shop)      �?
└─────────────�?
```

#### DLC 扩展�?

| 扩展内容 | 实现方式 |
|----------|----------|
| 新属性维度（�?暗影抗�?�?| �?Component 或扩�?BaseStats + 修改计算公式 |
| 新装备槽�?| EquipmentComponent.slots �?Dictionary，直接加�?key |
| 新技能类�?| �?SkillData 资源 + 可选的�?SkillEffect 类型 |
| 新成长路�?| �?GrowthProfile 资源文件 |
| 新物品类�?| �?ItemType 枚举�?+ 对应处理逻辑 |
| 新容器类�?| �?ContainerType + 对应 UI 面板 |
| 新物品效�?| �?UseEffect 脚本注册到效果处理器 |

---

### B4. 战斗/伤害系统

#### 伤害管线

```
攻击发起
    �?
    ├── 1. 命中判定（Hitbox/Hurtbox 碰撞检测）
    �?  └── Hitbox(Area3D) �?Hurtbox(Area3D) 重叠 �?生成 DamageEvent
    �?
    ├── 2. 伤害计算
    �?  ├── 基础伤害 = 攻击�?FinalStats.attack * 技能倍率
    �?  ├── 防御减伤 = 基础伤害 * (1 - 防御�?defense / (防御�?defense + 常数))
    �?  ├── 元素克制 = 减伤�?* 元素倍率表[攻击元素][防御元素]
    �?  ├── 暴击判定 = rand() < 攻击�?crit_rate �?伤害 * crit_damage
    �?  └── 格挡判定 = 防御�?is_blocking �?伤害 * block_reduction
    �?
    ├── 3. 伤害应用
    �?  └── 防御�?CharacterStats.hp -= final_damage
    �?
    └── 4. 后处�?
        ├── 生成伤害数字 UI
        ├── 触发受击状态（状态机 �?HitState�?
        ├── 应用击退/击飞
        └── 触发 Buff 效果（如吸血、反伤）
```

#### 核心数据结构

```gdscript
# ===== 伤害事件 Component（一次性，处理后移除）=====
class_name DamageEventComponent extends EcsComponent

@export var damage_amount: float = 0.0
@export var damage_type: DamageType = DamageType.PHYSICAL
@export var element: ElementType = ElementType.NONE
@export var attacker_id: int = -1
@export var skill_id: StringName = &""
@export var is_crit: bool = false
@export var knockback_force: Vector3 = Vector3.ZERO
@export var hit_position: Vector3 = Vector3.ZERO

enum DamageType { PHYSICAL, MAGICAL, TRUE }
enum ElementType { NONE, FIRE, ICE, LIGHTNING, SHADOW, HOLY }


# ===== 元素克制表（全局配置 Resource�?====
class_name ElementTable extends Resource

@export var multipliers: Dictionary = {
    # [攻击元素][防御元素] = 倍率
    # 1.0=正常, 1.5=克制, 0.5=抵抗, 0.0=免疫
}


# ===== 战斗配置 Component =====
class_name CombatConfigComponent extends EcsComponent

@export var hitbox_layers: int = 0
@export var hurtbox_layers: int = 0
@export var block_reduction: float = 0.7
@export var block_angle: float = 120.0  # 格挡有效角度
@export var invincible_on_dodge: bool = true
@export var dodge_i_frames: float = 0.3
```

#### Hitbox/Hurtbox 设计

```
角色 Entity
├── Hurtbox (Area3D, 始终激�?     �?受击区域
├── Hitbox_Weapon (Area3D, 默认关闭) �?武器攻击区域
�?  └── 由状态机在攻击动画关键帧激�?关闭
└── Hitbox_Skill (Area3D, 动态生�?  �?技能特效区�?

# 碰撞层分配：
# Layer 1: 玩家 Hurtbox
# Layer 2: 敌人 Hurtbox
# Layer 3: 玩家 Hitbox（mask=Layer2�?
# Layer 4: 敌人 Hitbox（mask=Layer1�?
# Layer 5: 环境碰撞
```

### B5. Buff/Debuff 系统

#### Buff 生命周期

```
Buff 施加
    �?
    ├── 堆叠检�?
    �?  ├── 不可堆叠 �?刷新持续时间
    �?  ├── 可堆叠（层数叠加）→ stack_count += 1
    �?  └── 可堆叠（独立计时）→ 添加新实�?
    �?
    ├── 免疫检�?
    �?  └── 目标有对应免疫标�?�?施加失败
    �?
    �?
Buff 生效中（每帧/每秒 tick�?
    �?
    ├── 持续时间递减
    ├── 周期性效果触发（如每秒回血、每秒中毒）
    └── 属性修饰器生效（FinalStats.is_dirty = true�?
    �?
    �?
Buff 结束
    ├── 持续时间到期 �?自动移除
    ├── 被净�?�?强制移除
    ├── 达到最大触发次�?�?移除
    └── 移除时：撤销属性修饰器、触发结束效�?
```

#### 数据结构

```gdscript
# ===== Buff 定义（策划配置）=====
class_name BuffData extends Resource

@export var id: StringName
@export var name: String
@export var icon: Texture2D
@export var buff_type: BuffType
@export var duration: float = 5.0
@export var tick_interval: float = 1.0       # 周期触发间隔�?=不触发）
@export var max_stacks: int = 1
@export var stack_mode: StackMode = StackMode.REFRESH
@export var stat_modifiers: Array[StatModifier] = []  # 属性修�?
@export var tick_effects: Array[SkillEffect] = []     # 周期效果
@export var on_apply_effects: Array[SkillEffect] = [] # 施加时效�?
@export var on_remove_effects: Array[SkillEffect] = [] # 移除时效�?
@export var immunity_tags: Array[StringName] = []     # 免疫�?Buff 的标�?
@export var dispel_type: DispelType = DispelType.MAGIC

enum BuffType { BUFF, DEBUFF }
enum StackMode { REFRESH, STACK_COUNT, STACK_INDEPENDENT }
enum DispelType { MAGIC, PHYSICAL, UNDISPELLABLE }


# ===== Buff 实例（运行时状态）=====
class_name BuffInstance extends Resource

@export var buff_data: BuffData
@export var remaining_time: float = 0.0
@export var stack_count: int = 1
@export var tick_timer: float = 0.0
@export var source_entity_id: int = -1


# ===== Buff 列表 Component =====
class_name BuffListComponent extends EcsComponent

@export var active_buffs: Array[BuffInstance] = []
@export var immunity_tags: Array[StringName] = []  # 当前免疫标签

func add_buff(buff: BuffData, source_id: int) -> bool:
    # 免疫检�?
    for tag in buff.immunity_tags:
        if tag in immunity_tags:
            return false
    # 堆叠逻辑
    var existing := _find_buff(buff.id)
    if existing:
        match buff.stack_mode:
            BuffData.StackMode.REFRESH:
                existing.remaining_time = buff.duration
            BuffData.StackMode.STACK_COUNT:
                existing.stack_count = mini(existing.stack_count + 1, buff.max_stacks)
                existing.remaining_time = buff.duration
    else:
        var instance := BuffInstance.new()
        instance.buff_data = buff
        instance.remaining_time = buff.duration
        instance.source_entity_id = source_id
        active_buffs.append(instance)
    return true

func get_all_modifiers() -> Array[StatModifier]:
    var result: Array[StatModifier] = []
    for buff_inst in active_buffs:
        for mod in buff_inst.buff_data.stat_modifiers:
            var scaled_mod := mod.duplicate()
            scaled_mod.value *= buff_inst.stack_count
            result.append(scaled_mod)
    return result


# ===== Buff Tick System =====
class_name BuffTickSystem extends EcsSystem

func get_query() -> Array[String]:
    return ["BuffList"]

func process(entities: Array, delta: float) -> void:
    for entity in entities:
        var buff_list := entity.get_component("BuffList") as BuffListComponent
        var expired: Array[BuffInstance] = []
        
        for buff_inst in buff_list.active_buffs:
            buff_inst.remaining_time -= delta
            
            # 周期触发
            if buff_inst.buff_data.tick_interval > 0:
                buff_inst.tick_timer += delta
                if buff_inst.tick_timer >= buff_inst.buff_data.tick_interval:
                    buff_inst.tick_timer -= buff_inst.buff_data.tick_interval
                    _apply_tick_effects(entity, buff_inst)
            
            # 过期检�?
            if buff_inst.remaining_time <= 0:
                expired.append(buff_inst)
        
        # 移除过期 Buff
        for buff_inst in expired:
            _remove_buff(entity, buff_list, buff_inst)
```

### B6. AI/敌人行为系统

#### 架构：行为树 + ECS 集成

```
敌人 Entity
├── StateMachine (高层状态：巡�?战斗/逃跑/死亡)
�?  └── CombatState 内部使用行为树做决策
├── BehaviorTree (Node)          �?战斗决策
�?  ├── Selector: 选择行为
�?  �?  ├── Sequence: 远程攻击（距�?5 �?有弹药）
�?  �?  ├── Sequence: 近战攻击（距�?2�?
�?  �?  ├── Sequence: 接近目标（距�?2�?
�?  �?  └── Action: 待机
�?  └── 每帧 tick �?输出决策 �?写入 Component
├── Component: AiState           �?AI 决策数据
├── Component: AggroTable        �?仇恨�?
├── Component: PatrolRoute       �?巡逻路�?
└── Component: CombatState       �?共用战斗状�?
```

#### 核心数据结构

```gdscript
# ===== AI 状�?Component =====
class_name AiStateComponent extends EcsComponent

@export var ai_mode: AiMode = AiMode.IDLE
@export var target_entity_id: int = -1
@export var alert_range: float = 10.0
@export var chase_range: float = 15.0
@export var attack_range: float = 2.0
@export var retreat_hp_threshold: float = 0.2
@export var current_decision: StringName = &""

enum AiMode { IDLE, PATROL, ALERT, COMBAT, RETREAT, DEAD }


# ===== 仇恨�?Component =====
class_name AggroTableComponent extends EcsComponent

@export var entries: Dictionary = {}  # { entity_id: aggro_value }
@export var decay_rate: float = 1.0   # 每秒仇恨衰减

func add_aggro(entity_id: int, amount: float) -> void:
    entries[entity_id] = entries.get(entity_id, 0.0) + amount

func get_top_target() -> int:
    var max_aggro := 0.0
    var target := -1
    for id in entries:
        if entries[id] > max_aggro:
            max_aggro = entries[id]
            target = id
    return target


# ===== Boss 阶段 Component =====
class_name BossPhaseComponent extends EcsComponent

@export var current_phase: int = 1
@export var phase_thresholds: Array[float] = [0.7, 0.4, 0.15]  # HP 百分�?
@export var phase_skills: Dictionary = {}  # { phase: [skill_ids] }


# ===== AI 决策 System =====
class_name AiDecisionSystem extends EcsSystem

func get_query() -> Array[String]:
    return ["AiState", "CombatState", "CharacterStats"]

func process(entities: Array, delta: float) -> void:
    for entity in entities:
        var ai := entity.get_component("AiState") as AiStateComponent
        var stats := entity.get_component("CharacterStats")
        
        match ai.ai_mode:
            AiMode.PATROL:
                _check_alert(entity, ai)
            AiMode.COMBAT:
                _update_combat(entity, ai, stats)
            AiMode.RETREAT:
                _update_retreat(entity, ai)

# 仇恨衰减 System
class_name AggroDecaySystem extends EcsSystem

func get_query() -> Array[String]:
    return ["AggroTable"]

func process(entities: Array, delta: float) -> void:
    for entity in entities:
        var aggro := entity.get_component("AggroTable") as AggroTableComponent
        var to_remove: Array[int] = []
        for id in aggro.entries:
            aggro.entries[id] -= aggro.decay_rate * delta
            if aggro.entries[id] <= 0:
                to_remove.append(id)
        for id in to_remove:
            aggro.entries.erase(id)
```

### B7. 输入/连招系统

#### 输入缓冲设计

```
玩家按键
    �?
    �?
InputBuffer（环形缓冲区，记录最�?N 帧的输入�?
    �?
    ├── 当前帧输入：{ action: "attack", timestamp: 1234, direction: Vector2 }
    ├── 上一�?..
    └── N帧前...
    �?
    �?
ComboMatcher（连招匹配器�?
    �?
    ├── 遍历连招表，检查缓冲区是否匹配某个连招序列
    ├── 检查取消窗口（当前动画是否允许被取消）
    └── 输出：匹配到的技能ID �?null
    �?
    �?
写入 Component �?状态机响应
```

#### 数据结构

```gdscript
# ===== 输入缓冲 Component =====
class_name InputBufferComponent extends EcsComponent

@export var buffer_size: int = 10           # 缓冲帧数
@export var buffer: Array[InputFrame] = []
@export var buffer_window: float = 0.3      # 缓冲有效时间（秒�?

class InputFrame:
    var action: StringName
    var timestamp: float
    var direction: Vector2
    var is_held: bool


# ===== 连招表定�?=====
class_name ComboData extends Resource

@export var id: StringName
@export var name: String
@export var sequence: Array[ComboInput] = []  # 输入序列
@export var result_skill: StringName          # 触发的技�?
@export var cancel_window: Vector2 = Vector2(0.2, 0.8)  # 可取消的动画进度区间
@export var required_state: StringName = &""  # 需要在什么状态下才能触发

class ComboInput:
    var action: StringName        # "attack" / "special" / "dodge"
    var max_interval: float = 0.5 # 与上一个输入的最大间�?
    var direction: StringName = &""  # 可选方向要求："forward"/"back"


# ===== 连招匹配�?Component =====
class_name ComboMatcherComponent extends EcsComponent

@export var combo_table: Array[ComboData] = []
@export var current_combo_progress: Dictionary = {}  # { combo_id: matched_count }
@export var matched_combo: StringName = &""          # 本帧匹配到的连招


# ===== 输入处理 System =====
class_name InputProcessSystem extends EcsSystem

func get_query() -> Array[String]:
    return ["InputBuffer", "ComboMatcher", "CombatState"]

func process(entities: Array, delta: float) -> void:
    for entity in entities:
        var buffer := entity.get_component("InputBuffer") as InputBufferComponent
        var matcher := entity.get_component("ComboMatcher") as ComboMatcherComponent
        var combat := entity.get_component("CombatState") as CombatStateComponent
        
        matcher.matched_combo = &""
        
        # 清理过期输入
        _clean_expired(buffer)
        
        # 尝试匹配连招
        for combo in matcher.combo_table:
            if _try_match(buffer, combo, combat):
                matcher.matched_combo = combo.result_skill
                break  # 优先匹配第一个（按优先级排序�?
```

### B8. 存档系统

#### 需要存档的数据

| 数据类别 | 来源 | 存档方式 |
|----------|------|----------|
| 角色属�?| BaseStats, Experience | 序列�?Component |
| 装备/背包 | Equipment, Inventory | 序列化物品ID+实例数据 |
| 技能状�?| SkillSet | 已学技能列�?等级 |
| 任务进度 | QuestTracker | 任务状态表 |
| 世界状�?| WorldState | 已开宝箱、已触发事件、NPC状�?|
| 储物�?| Storage containers | 序列�?slots |
| 玩家位置 | Transform | 场景ID + 坐标 |
| 游戏设置 | Settings | 独立文件，不随存�?|

#### 架构设计

```gdscript
# ===== 存档数据结构 =====
class_name SaveData extends Resource

@export var save_id: String = ""
@export var save_name: String = ""
@export var timestamp: int = 0
@export var play_time: float = 0.0
@export var scene_id: StringName = &""
@export var player_position: Vector3 = Vector3.ZERO

# 各系统数�?
@export var character_data: Dictionary = {}    # 角色属�?等级/经验
@export var inventory_data: Dictionary = {}    # 背包/装备
@export var skill_data: Dictionary = {}        # 技�?
@export var quest_data: Dictionary = {}        # 任务
@export var world_state: Dictionary = {}       # 世界状�?
@export var storage_data: Dictionary = {}      # 储物�?
@export var dlc_data: Dictionary = {}          # DLC 扩展数据


# ===== 存档管理器（Autoload�?====
class_name SaveManager extends Node

const SAVE_DIR := "user://saves/"
const MAX_SLOTS := 10

func save_game(slot: int) -> bool:
    var data := SaveData.new()
    data.save_id = "save_%d" % slot
    data.timestamp = int(Time.get_unix_time_from_system())
    
    # 收集各系统数�?
    _collect_character_data(data)
    _collect_inventory_data(data)
    _collect_skill_data(data)
    _collect_quest_data(data)
    _collect_world_state(data)
    _collect_storage_data(data)
    _collect_dlc_data(data)
    
    # 序列化写�?
    var path := SAVE_DIR + "slot_%d.tres" % slot
    return ResourceSaver.save(data, path) == OK

func load_game(slot: int) -> bool:
    var path := SAVE_DIR + "slot_%d.tres" % slot
    if not FileAccess.file_exists(path):
        return false
    
    var data := ResourceLoader.load(path) as SaveData
    if not data:
        return false
    
    # 恢复各系统数�?
    _restore_character_data(data)
    _restore_inventory_data(data)
    _restore_skill_data(data)
    _restore_quest_data(data)
    _restore_world_state(data)
    _restore_storage_data(data)
    _restore_dlc_data(data)
    
    # 加载场景
    _load_scene(data.scene_id, data.player_position)
    return true


# ===== 可存档接口（各系统实现）=====
# 每个需要存档的系统实现 ISaveable 接口�?
# func collect_save_data() -> Dictionary
# func restore_save_data(data: Dictionary) -> void
```

#### 存档�?ECS/DLC 的集�?

```
存档时：
    ├── EcsWorld 遍历所有标记为 "saveable" �?Entity
    �?  └── 序列化其所�?Component 数据
    ├── DlcManager 记录当前已加载的 DLC 列表
    �?  └── DLC 扩展�?Component 数据也一并序列化
    └── 写入 .tres 文件

读档时：
    ├── 先加�?DLC（确�?Component 类型已注册）
    ├── 创建/恢复 Entity
    �?  └── 反序列化 Component 数据
    └── �?System 自动开始处理恢复后的数�?

DLC 兼容性：
    ├── 存档中记�?DLC 版本
    ├── 读档时检�?DLC 是否仍然可用
    └── 缺失 DLC 的数据标记为 "orphaned"，不加载但不报错
```

---

### B9. 事件总线系统

#### 设计目标

提供全局的发�?订阅机制，让系统间通过事件名解耦通信，避免直接引用�?

#### 架构

```gdscript
# ===== EventBus（Autoload�?====
class_name EventBus extends Node
## 全局事件总线，所有系统间通信的中�?

# 事件注册表：{ event_name: Array[Callable] }
var _listeners: Dictionary = {}

# 事件历史（调试用�?
var _history: Array[Dictionary] = []
var _history_enabled: bool = false
const MAX_HISTORY := 100


func subscribe(event_name: StringName, callback: Callable, priority: int = 0) -> void:
    ## 订阅事件
    if event_name not in _listeners:
        _listeners[event_name] = []
    _listeners[event_name].append({ "callback": callback, "priority": priority })
    _listeners[event_name].sort_custom(func(a, b): return a.priority > b.priority)


func unsubscribe(event_name: StringName, callback: Callable) -> void:
    ## 取消订阅
    if event_name in _listeners:
        _listeners[event_name] = _listeners[event_name].filter(
            func(entry): return entry.callback != callback
        )


func emit_event(event_name: StringName, data: Dictionary = {}) -> void:
    ## 发布事件
    if _history_enabled:
        _history.append({ "event": event_name, "data": data, "time": Time.get_ticks_msec() })
        if _history.size() > MAX_HISTORY:
            _history.pop_front()
    
    if event_name in _listeners:
        for entry in _listeners[event_name]:
            entry.callback.call(data)


func emit_deferred(event_name: StringName, data: Dictionary = {}) -> void:
    ## 延迟到帧末发布（避免在遍历中修改状态）
    call_deferred("emit_event", event_name, data)
```

#### 预定义事件清�?

| 事件�?| 触发时机 | data 字段 |
|--------|----------|-----------|
| `player_damaged` | 玩家受伤 | { amount, source, element } |
| `player_healed` | 玩家回血 | { amount, source } |
| `player_died` | 玩家死亡 | { killer_id } |
| `player_leveled_up` | 升级 | { new_level, old_level } |
| `enemy_killed` | 击杀敌人 | { enemy_id, enemy_type, position } |
| `item_acquired` | 获得物品 | { item_id, count, source } |
| `item_used` | 使用物品 | { item_id, target_id } |
| `equipment_changed` | 装备变更 | { slot, old_item, new_item } |
| `skill_learned` | 学习技�?| { skill_id, level } |
| `skill_used` | 使用技�?| { skill_id, target_ids } |
| `quest_started` | 任务开�?| { quest_id } |
| `quest_completed` | 任务完成 | { quest_id, rewards } |
| `scene_entered` | 进入场景 | { scene_id } |
| `dialogue_started` | 对话开�?| { npc_id, dialogue_id } |
| `buff_applied` | Buff 施加 | { buff_id, target_id, source_id } |
| `buff_removed` | Buff 移除 | { buff_id, target_id, reason } |
| `dlc_loaded` | DLC 加载 | { dlc_id } |

#### �?ECS 的集�?

```
ECS System 内部处理完逻辑�?�?EventBus.emit_event() 通知外部
外部系统（UI、音效、成就）�?EventBus.subscribe() 监听并响�?

示例�?
LevelUpSystem 升级�?�?emit("player_leveled_up", { new_level: 10 })
    ├── UI 系统监听 �?显示升级特效
    ├── 音效系统监听 �?播放升级音效
    ├── 成就系统监听 �?检�?达到10�?成就
    └── 任务系统监听 �?检�?升到10�?任务目标
```

### B10. 对象池系�?

#### 设计目标

避免高频创建/销毁节点导致的 GC 压力和帧率波动。适用于：子弹、伤害数字、粒子特效、掉落物、音效播放器�?

#### 架构

```gdscript
# ===== ObjectPool（Autoload�?====
class_name ObjectPool extends Node
## 全局对象池管理器

# 池注册表：{ pool_name: PoolData }
var _pools: Dictionary = {}


class PoolData:
    var scene: PackedScene
    var available: Array[Node] = []
    var in_use: Array[Node] = []
    var max_size: int = 50
    var auto_expand: bool = true
    var parent: Node = null


func register_pool(pool_name: StringName, scene: PackedScene, 
                   initial_size: int = 10, max_size: int = 50,
                   parent: Node = null) -> void:
    ## 注册一个对象池
    var pool := PoolData.new()
    pool.scene = scene
    pool.max_size = max_size
    pool.parent = parent if parent else self
    _pools[pool_name] = pool
    
    # 预创�?
    for i in initial_size:
        var instance := scene.instantiate()
        instance.set_process(false)
        instance.visible = false
        pool.parent.add_child(instance)
        pool.available.append(instance)


func acquire(pool_name: StringName) -> Node:
    ## 从池中获取一个对�?
    var pool := _pools.get(pool_name) as PoolData
    if not pool:
        push_error("对象池不存在: %s" % pool_name)
        return null
    
    var instance: Node
    if pool.available.size() > 0:
        instance = pool.available.pop_back()
    elif pool.auto_expand and pool.in_use.size() < pool.max_size:
        instance = pool.scene.instantiate()
        pool.parent.add_child(instance)
    else:
        # 池已满，回收最早使用的
        instance = pool.in_use.pop_front()
        _reset_instance(instance)
    
    instance.set_process(true)
    instance.visible = true
    pool.in_use.append(instance)
    
    if instance.has_method("on_pool_acquire"):
        instance.on_pool_acquire()
    
    return instance


func release(pool_name: StringName, instance: Node) -> void:
    ## 归还对象到池�?
    var pool := _pools.get(pool_name) as PoolData
    if not pool:
        return
    
    pool.in_use.erase(instance)
    _reset_instance(instance)
    pool.available.append(instance)


func _reset_instance(instance: Node) -> void:
    instance.set_process(false)
    instance.visible = false
    if instance.has_method("on_pool_release"):
        instance.on_pool_release()
```

#### 使用示例

```gdscript
# 注册池（游戏启动时）
ObjectPool.register_pool(&"damage_number", preload("res://scenes/ui/damage_number.tscn"), 20, 50)
ObjectPool.register_pool(&"hit_particle", preload("res://scenes/fx/hit_particle.tscn"), 10, 30)
ObjectPool.register_pool(&"projectile", preload("res://scenes/combat/projectile.tscn"), 15, 40)

# 使用（战斗中�?
var dmg_num := ObjectPool.acquire(&"damage_number") as DamageNumber
dmg_num.show_damage(150, hit_position, is_crit)

# 归还（动画播放完毕后�?
func _on_animation_finished():
    ObjectPool.release(&"damage_number", self)
```

#### 池化对象接口约定

```gdscript
# 任何需要池化的场景根节点应实现�?

func on_pool_acquire() -> void:
    ## 从池中取出时调用（重置状态、启动逻辑�?
    pass

func on_pool_release() -> void:
    ## 归还到池中时调用（停止逻辑、清理引用）
    pass
```

### B11. 数据�?配置管理系统

#### 设计目标

统一管理所有策划配置数据（物品表、技能表、怪物表、关卡配置等），提供高效索引和热重载能力�?

#### 架构

```gdscript
# ===== DataManager（Autoload�?====
class_name DataManager extends Node
## 全局数据表管理器

# 已加载的数据表：{ table_name: Dictionary[id, Resource] }
var _tables: Dictionary = {}

# 数据表路径配�?
const TABLE_PATHS := {
    &"items": "res://resources/data/items/",
    &"skills": "res://resources/data/skills/",
    &"buffs": "res://resources/data/buffs/",
    &"enemies": "res://resources/data/enemies/",
    &"quests": "res://resources/data/quests/",
    &"dialogues": "res://resources/data/dialogues/",
    &"growth_profiles": "res://resources/data/growth/",
    &"combo_tables": "res://resources/data/combos/",
    &"element_tables": "res://resources/data/elements/",
}


func _ready() -> void:
    _load_all_tables()


func _load_all_tables() -> void:
    for table_name in TABLE_PATHS:
        _tables[table_name] = {}
        var dir := DirAccess.open(TABLE_PATHS[table_name])
        if not dir:
            continue
        dir.list_dir_begin()
        var file_name := dir.get_next()
        while file_name != "":
            if file_name.ends_with(".tres"):
                var res := ResourceLoader.load(TABLE_PATHS[table_name] + file_name)
                if res and res.has_method("get") and "id" in res:
                    _tables[table_name][res.id] = res
            file_name = dir.get_next()


func get_item(id: StringName) -> ItemData:
    ## 获取物品数据
    return _tables.get(&"items", {}).get(id)


func get_skill(id: StringName) -> SkillData:
    ## 获取技能数�?
    return _tables.get(&"skills", {}).get(id)


func get_buff(id: StringName) -> BuffData:
    ## 获取 Buff 数据
    return _tables.get(&"buffs", {}).get(id)


func get_enemy(id: StringName) -> Resource:
    ## 获取敌人配置
    return _tables.get(&"enemies", {}).get(id)


func get_all(table_name: StringName) -> Dictionary:
    ## 获取整张�?
    return _tables.get(table_name, {})


func query(table_name: StringName, filter: Callable) -> Array:
    ## 条件查询
    var results: Array = []
    var table := _tables.get(table_name, {})
    for item in table.values():
        if filter.call(item):
            results.append(item)
    return results


func reload_table(table_name: StringName) -> void:
    ## 热重载单张表（开发期调试用）
    _tables[table_name] = {}
    var dir := DirAccess.open(TABLE_PATHS[table_name])
    if not dir:
        return
    dir.list_dir_begin()
    var file_name := dir.get_next()
    while file_name != "":
        if file_name.ends_with(".tres"):
            var res := ResourceLoader.load(TABLE_PATHS[table_name] + file_name, "", ResourceLoader.CACHE_MODE_IGNORE)
            if res and "id" in res:
                _tables[table_name][res.id] = res
        file_name = dir.get_next()
    EventBus.emit_event(&"data_table_reloaded", { "table": table_name })
```

#### 数据文件组织

```
resources/data/
├── items/
�?  ├── sword_iron.tres         # ItemData
�?  ├── potion_hp_small.tres
�?  └── ...
├── skills/
�?  ├── slash_basic.tres        # SkillData
�?  ├── fireball.tres
�?  └── ...
├── buffs/
�?  ├── poison.tres             # BuffData
�?  ├── attack_up.tres
�?  └── ...
├── enemies/
�?  ├── goblin.tres             # EnemyConfig
�?  ├── boss_dragon.tres
�?  └── ...
├── growth/
�?  ├── warrior_growth.tres     # GrowthProfile
�?  ├── mage_growth.tres
�?  └── ...
└── combos/
    ├── warrior_combos.tres     # ComboData[]
    └── ...
```

#### �?DLC 的集�?

```
DLC 加载时：
    ├── DlcManager 扫描 DLC �?data/ 目录
    ├── �?DLC 数据注册�?DataManager 对应表中
    �?  └── DataManager._tables["items"]["dlc_shadow_sword"] = dlc_item_data
    └── emit("data_table_updated", { table: "items", source: "dlc_dark_realm" })

DLC 卸载时：
    ├── �?DataManager 中移除该 DLC 注册的所有条�?
    └── emit("data_table_updated", ...)

冲突处理�?
    ├── 如果 DLC 物品 ID 与主游戏冲突 �?加载失败并报�?
    └── DLC 数据 ID 建议使用前缀：dlc_{dlc_id}_{item_name}
```

#### 使用示例

```gdscript
# 获取单个物品
var sword := DataManager.get_item(&"sword_iron")

# 条件查询：所有稀有度>=3的武�?
var rare_weapons := DataManager.query(&"items", func(item: ItemData):
    return item.rarity >= 3 and item.item_type == ItemData.ItemType.WEAPON
)

# 热重载（开发期按快捷键触发�?
DataManager.reload_table(&"skills")
```

---

## Part C: 实施计划

| 序号 | 产出�?| 类型 | 对应需�?|
|------|--------|------|----------|
| 1 | `.kiro/steering/core.md`（改造） | 规范文件 | Req 1-10 |
| 2 | `AIDOC/global-info/knowledge/steering/` 全部规范文件 | 规范文件 | Req 2-10 |
| 3 | `AIDOC/global-info/knowledge/engines/godot/` 知识文件 | 知识�?| Req 5 |
| 4 | `AIDOC/global-info/knowledge/patterns/` 模式文件 | 知识�?| Req 5 |
| 5 | `AIDOC/global-info/knowledge/templates/` 模板文件 | 知识�?| Req 5 |
| 6 | `addons/gd_ecs/` 插件代码 | GDScript | Req 11 |
| 7 | `addons/dlc_manager/` 插件代码 | GDScript | Req 12 |

### 实施顺序

```
Phase 1: 基础规范（Req 1, 3, 4�?
    ├── 改�?core.md
    ├── 创建目录骨架
    └── 编写执行协议和代码生成规�?

Phase 2: 工作流规范（Req 2, 10�?
    ├── 编写开发工作流
    └── 编写项目启动流程

Phase 3: 领域规范（Req 5, 6, 7, 8, 9�?
    ├── 创建知识库骨架文�?
    ├── 编写叙事工作�?
    ├── 编写 2D/3D 分支规范
    ├── 编写资产管线规范
    └── 编写迭代优化规范

Phase 4: ECS 框架实现（Req 11�?
    ├── 实现 core/ 核心�?
    ├── 实现 debug/ 调试工具
    └── 编写示例和文�?

Phase 5: DLC 系统实现（Req 12�?
    ├── 实现 core/ 核心�?
    ├── 实现 integration/ 集成�?
    └── 编写示例 DLC 和文�?
```
