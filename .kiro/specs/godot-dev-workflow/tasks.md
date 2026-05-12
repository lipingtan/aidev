# Tasks

## Task Dependency Graph

```
Phase 1 (基础规范)
  T1 → T2 → T3

Phase 2 (工作流规范)
  T3 → T4 → T5

Phase 3 (领域规范)
  T5 → T6, T7, T8, T9, T10 (可并行)

Phase 4 (基础设施代码)
  T3 → T11 → T12 → T13 → T14

Phase 5 (核心框架代码)
  T14 → T15 → T16

Phase 6 (玩法系统代码)
  T16 → T17 → T18 → T19 → T20 → T21

Phase 7 (DLC 系统)
  T15 + T14 → T22 → T23
```

---

## Phase 1: 基础规范

### Task 1: 改造 core.md 核心规范文件
- **Status:** not started
- **Requirements:** Req 1, 3, 4
- **Description:** 将现有 `.kiro/steering/core.md` 从 AI 短剧工作台改造为 Godot 游戏开发工作台核心规范，包含项目身份、语言规范、目录分离原则、默认引擎配置、知识库索引、双轨工作流规则、门控机制规则、AI 代码生成规则、2D/3D 分支规则
- **Output:** `.kiro/steering/core.md`（改造后）
- **Acceptance Criteria:**
  - 文件包含所有 9 个规范章节
  - 知识库索引表覆盖所有任务类型
  - AI 代码生成规则包含 GDScript 4.x 规范要求

### Task 2: 创建目录骨架
- **Status:** not started
- **Requirements:** Req 1
- **Description:** 按设计文档 A1 创建完整的目录结构骨架，包含 AIDOC/global-info/knowledge/ 全部子目录、AIDOC/projects/ 模板、godot_projects/ 模板结构
- **Output:** 目录骨架 + 各目录的 README.md
- **Acceptance Criteria:**
  - AIDOC/global-info/knowledge/ 下 engines/godot/、patterns/、templates/rpg/、templates/action/、cases/、steering/ 目录存在
  - 每个目录有 README.md 说明用途

### Task 3: 编写执行协议和代码生成规范
- **Status:** not started
- **Requirements:** Req 3, 4
- **Description:** 编写 `AIDOC/global-info/knowledge/steering/execution-protocol.md`（适配游戏开发的 PRE-CHECK/POST-CHECK 模板）和 `code-generation.md`（AI 代码生成规范详细版）
- **Output:** `steering/execution-protocol.md`, `steering/code-generation.md`
- **Acceptance Criteria:**
  - 执行协议包含游戏开发各 Phase 的具体检查清单
  - 代码生成规范包含 GDScript 文件头模板、碰撞层分配方案、Shader 注释规范

---

## Phase 2: 工作流规范

### Task 4: 编写开发工作流规范
- **Status:** not started
- **Requirements:** Req 2, 10
- **Description:** 编写 `steering/development-workflow.md`，定义双轨并行工作流的完整流程、同步点、门控机制、Phase 0~5 各阶段产出物和验收标准
- **Output:** `steering/development-workflow.md`
- **Acceptance Criteria:**
  - 包含创意轨道和技术轨道的完整阶段定义
  - 包含 3 个同步点的触发规则
  - 包含 Phase 0~5 的产出物清单和验收检查项

### Task 5: 编写项目启动流程
- **Status:** not started
- **Requirements:** Req 10
- **Description:** 编写 `steering/project-bootstrap.md`，定义从零启动一个新游戏项目的完整流程（类似原工程的 series-bootstrap-workflow），包含澄清问题、目录创建、技术选型
- **Output:** `steering/project-bootstrap.md`
- **Acceptance Criteria:**
  - 包含 Phase 0 的完整澄清问题列表
  - 包含目录骨架自动创建规则
  - 包含快捷模式（已有部分设计时的跳过规则）

---

## Phase 3: 领域规范

### Task 6: 编写叙事工作流规范
- **Status:** not started
- **Requirements:** Req 6
- **Description:** 编写 `steering/narrative-workflow.md`，定义 RPG 叙事从策划到实现的完整流程，包含对话数据格式、任务数据格式、触发条件设计规范
- **Output:** `steering/narrative-workflow.md`
- **Acceptance Criteria:**
  - 包含对话 JSON 格式示例
  - 包含任务数据格式示例
  - 包含叙事与玩法集成点定义

### Task 7: 编写 2D/3D 分支规范
- **Status:** not started
- **Requirements:** Req 7
- **Description:** 编写 `steering/branch-routing.md`，定义在场景搭建、角色控制器、相机、物理、光照、UI 各节点的 2D/3D 分支处理规则
- **Output:** `steering/branch-routing.md`
- **Acceptance Criteria:**
  - 包含 6 个分支节点的 2D/3D 对比表
  - 包含 3D 资产导入流程步骤
  - 包含 2D AI 生成流程步骤
  - 包含混合模式规范

### Task 8: 编写资产管线规范
- **Status:** not started
- **Requirements:** Req 8
- **Description:** 编写 `steering/asset-pipeline.md`，定义 3D 素材包导入和 2D AI 生成的完整生命周期、兼容性检查清单、优化规范
- **Output:** `steering/asset-pipeline.md`
- **Acceptance Criteria:**
  - 包含 3D 资产 7 阶段生命周期
  - 包含 2D 资产 7 阶段生命周期
  - 包含资产注册表格式定义
  - 包含性能预算阈值模板

### Task 9: 编写迭代优化规范
- **Status:** not started
- **Requirements:** Req 9
- **Description:** 编写 `steering/iteration-workflow.md`，定义 Bug 修复、性能优化、玩法调优三类迭代的标准流程和日志格式
- **Output:** `steering/iteration-workflow.md`
- **Acceptance Criteria:**
  - 包含三类迭代的启动条件和完成条件
  - 包含迭代日志模板
  - 包含性能预算基准模板

### Task 10: 创建知识库骨架文件
- **Status:** not started
- **Requirements:** Req 5
- **Description:** 创建 `knowledge/engines/godot/` 下 7 个知识文件、`patterns/` 下 6 个模式文件、`templates/rpg/` 下 7 个模板文件、`templates/action/` 下 6 个模板文件的骨架（含标题和章节结构，内容待后续填充）
- **Output:** 26 个骨架 .md 文件
- **Acceptance Criteria:**
  - 每个文件包含标题、概念说明、适用场景、代码示例、常见陷阱的章节标题
  - 知识库索引文件可正确映射到所有文件路径

---

## Phase 4: 基础设施代码

### Task 11: 实现事件总线系统
- **Status:** not started
- **Requirements:** Req 11 (基础设施)
- **Design:** B9
- **Description:** 实现 EventBus Autoload，包含 subscribe/unsubscribe/emit_event/emit_deferred 方法，支持优先级排序和调试历史
- **Output:** `autoload/event_bus.gd`
- **Acceptance Criteria:**
  - 可订阅/取消订阅事件
  - 支持优先级排序
  - emit_deferred 延迟到帧末执行
  - 调试模式下记录事件历史

### Task 12: 实现对象池系统
- **Status:** not started
- **Requirements:** Req 11 (基础设施)
- **Design:** B10
- **Description:** 实现 ObjectPool Autoload，包含 register_pool/acquire/release 方法，支持预创建、自动扩展、池满回收
- **Output:** `autoload/object_pool.gd`
- **Acceptance Criteria:**
  - 可注册池并预创建实例
  - acquire 返回可用实例
  - release 归还并重置实例
  - 支持 on_pool_acquire/on_pool_release 回调

### Task 13: 实现数据表管理系统
- **Status:** not started
- **Requirements:** Req 11 (基础设施)
- **Design:** B11
- **Description:** 实现 DataManager Autoload，包含自动加载数据目录、按 ID 查询、条件查询、热重载功能
- **Output:** `autoload/data_manager.gd`
- **Acceptance Criteria:**
  - 启动时自动扫描 resources/data/ 目录加载所有 .tres
  - get_item/get_skill/get_buff 按 ID 查询
  - query() 支持 Callable 过滤
  - reload_table() 可热重载单张表

### Task 14: 实现存档系统
- **Status:** not started
- **Requirements:** Req 11 (基础设施)
- **Design:** B8
- **Description:** 实现 SaveManager Autoload 和 SaveData Resource，支持多存档槽、序列化/反序列化、DLC 数据兼容
- **Output:** `autoload/save_manager.gd`, `resources/save_data.gd`
- **Acceptance Criteria:**
  - save_game(slot) 序列化所有系统数据到 .tres 文件
  - load_game(slot) 反序列化并恢复状态
  - 支持最多 10 个存档槽
  - DLC 数据独立存储，缺失 DLC 时不报错

---

## Phase 5: 核心框架代码

### Task 15: 实现 ECS 混合框架核心
- **Status:** not started
- **Requirements:** Req 11
- **Design:** B1
- **Description:** 实现 gd_ecs addon 的核心类：EcsWorld（Autoload）、EcsEntity（extends Node）、EcsComponent（extends Resource）、EcsSystem（extends RefCounted）、EcsQuery
- **Output:** `addons/gd_ecs/core/` 下 5 个文件 + plugin.cfg + plugin.gd
- **Acceptance Criteria:**
  - Entity 进入场景树自动注册到 World
  - Component 可通过 Inspector 编辑
  - System 按 priority 排序执行
  - Query 正确过滤匹配 Entity
  - 支持运行时动态注册 Component/System

### Task 16: 实现状态机框架
- **Status:** not started
- **Requirements:** Req 11
- **Design:** B1 (状态机 + ECS 集成)
- **Description:** 实现通用状态机框架：StateMachine（Node）、State（基类），与 ECS Entity 集成，状态通过读写 Component 通信
- **Output:** `addons/gd_ecs/state_machine/` 下 state_machine.gd, state.gd
- **Acceptance Criteria:**
  - StateMachine 管理状态切换（enter/exit/update）
  - State 可访问父 Entity 的 Component
  - 支持动态添加/移除状态（DLC 扩展用）

---

## Phase 6: 玩法系统代码

### Task 17: 实现 RPG 属性系统
- **Status:** not started
- **Requirements:** Req 11
- **Design:** B3 (属性部分)
- **Description:** 实现 BaseStatsComponent、FinalStatsComponent、StatModifier、StatsCalculationSystem
- **Output:** `scripts/core/stats/` 下相关文件
- **Acceptance Criteria:**
  - FinalStats 在 is_dirty 时自动重算
  - 支持 FLAT_ADD/PERCENT_ADD/PERCENT_MULT 三种修饰类型
  - 计算公式：final = (base + flat) * (1 + pct_add) * pct_mult

### Task 18: 实现物品/装备/容器系统
- **Status:** not started
- **Requirements:** Req 11
- **Design:** B3 (物品/装备/容器部分)
- **Description:** 实现 ItemData、EquipmentStats、ItemInstance、ItemSlot、ItemContainerComponent、EquipmentComponent、ContainerSystem
- **Output:** `scripts/core/inventory/` 下相关文件 + `resources/data/items/` 示例数据
- **Acceptance Criteria:**
  - 容器支持 6 种类型（背包/装备栏/储物柜/商店/掉落物/交易）
  - 物品可在容器间转移
  - 装备变更触发属性重算
  - 支持堆叠和拆分

### Task 19: 实现技能和成长系统
- **Status:** not started
- **Requirements:** Req 11
- **Design:** B3 (技能/成长部分)
- **Description:** 实现 SkillData、SkillEffect、SkillSetComponent、SkillCooldownSystem、GrowthProfile、ExperienceComponent、LevelUpSystem
- **Output:** `scripts/core/skills/` + `scripts/core/growth/` 下相关文件
- **Acceptance Criteria:**
  - 技能冷却每帧递减
  - 被动技能提供属性加成
  - 升级时按成长曲线增加基础属性
  - 升级时自动解锁技能

### Task 20: 实现战斗和 Buff 系统
- **Status:** not started
- **Requirements:** Req 11
- **Design:** B4, B5
- **Description:** 实现 DamageEventComponent、DamageSystem（伤害管线）、Hitbox/Hurtbox 碰撞检测、BuffData、BuffInstance、BuffListComponent、BuffTickSystem
- **Output:** `scripts/core/combat/` + `scripts/core/buffs/` 下相关文件
- **Acceptance Criteria:**
  - 伤害管线：命中→计算→应用→后处理
  - 支持元素克制倍率
  - 无敌帧检查
  - Buff 支持堆叠/刷新/独立三种模式
  - Buff 周期 tick 和过期自动移除

### Task 21: 实现 AI 行为和输入/连招系统
- **Status:** not started
- **Requirements:** Req 11
- **Design:** B6, B7
- **Description:** 实现 AiStateComponent、AggroTableComponent、AiDecisionSystem、InputBufferComponent、ComboData、ComboMatcherComponent、InputProcessSystem
- **Output:** `scripts/core/ai/` + `scripts/core/input/` 下相关文件
- **Acceptance Criteria:**
  - AI 支持巡逻/警戒/战斗/撤退模式切换
  - 仇恨表支持添加/衰减/获取最高目标
  - 输入缓冲记录最近 N 帧输入
  - 连招匹配器按优先级匹配连招表

---

## Phase 7: DLC 系统

### Task 22: 实现 DLC 管理器核心
- **Status:** not started
- **Requirements:** Req 12
- **Design:** B2
- **Description:** 实现 dlc_manager addon 的核心类：DlcManager（Autoload）、DlcPackage、DlcManifest、DlcValidator、DlcLoader
- **Output:** `addons/dlc_manager/core/` 下 5 个文件 + plugin.cfg + plugin.gd
- **Acceptance Criteria:**
  - 启动时扫描 dlc/ 目录发现 DLC 包
  - 解析 manifest.json 并验证版本兼容性
  - 按依赖顺序加载 DLC
  - 冲突检测阻止互斥 DLC 同时加载
  - 支持目录和 PCK 两种格式

### Task 23: 实现 DLC 与 ECS/资产集成层
- **Status:** not started
- **Requirements:** Req 12
- **Design:** B2 (integration/)
- **Description:** 实现 EcsIntegration（DLC Component/System 注册到 EcsWorld）和 AssetIntegration（DLC 资产注册到 DataManager），以及角色扩展机制
- **Output:** `addons/dlc_manager/integration/` 下 2 个文件 + 示例 DLC 包
- **Acceptance Criteria:**
  - DLC 的 Component 自动注册到 EcsWorld
  - DLC 的 System 自动注册并参与调度
  - DLC 资产可通过 DataManager 统一访问
  - 角色扩展数据可注入到现有 Entity
  - DLC 卸载时清理所有注册内容
