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
- **Status:** completed

### Task 2: 创建目录骨架
- **Status:** completed

### Task 3: 编写执行协议和代码生成规范
- **Status:** completed

---

## Phase 2: 工作流规范

### Task 4: 编写开发工作流规范
- **Status:** completed

### Task 5: 编写项目启动流程
- **Status:** completed

---

## Phase 3: 领域规范

### Task 6: 编写叙事工作流规范
- **Status:** completed

### Task 7: 编写 2D/3D 分支规范
- **Status:** completed

### Task 8: 编写资产管线规范
- **Status:** completed

### Task 9: 编写迭代优化规范
- **Status:** completed

### Task 10: 创建知识库骨架文件
- **Status:** completed

---

## Phase 4: 基础设施代码

### Task 11: 实现事件总线系统
- **Status:** completed
- **Output:** `addons/gd_ecs/core/event_bus.gd`

### Task 12: 实现对象池系统
- **Status:** completed
- **Output:** `addons/gd_ecs/core/object_pool.gd`

### Task 13: 实现数据表管理系统
- **Status:** completed
- **Output:** `addons/gd_ecs/core/data_manager.gd`

### Task 14: 实现存档系统
- **Status:** completed
- **Output:** `systems/save/save_manager.gd`, `systems/save/save_data.gd`

---

## Phase 5: 核心框架代码

### Task 15: 实现 ECS 混合框架核心
- **Status:** completed
- **Output:** `addons/gd_ecs/core/` (ecs_world, ecs_entity, ecs_component, ecs_system + 变体)

### Task 16: 实现状态机框架
- **Status:** completed
- **Output:** `addons/gd_ecs/state_machine/` (state_machine.gd, state.gd)

---

## Phase 6: 玩法系统代码

### Task 17: 实现 RPG 属性系统
- **Status:** completed
- **Output:** `systems/stats/`

### Task 18: 实现物品/装备/容器系统
- **Status:** completed
- **Output:** `systems/inventory/`

### Task 19: 实现技能和成长系统
- **Status:** completed
- **Output:** `systems/skills/`, `systems/growth/`

### Task 20: 实现战斗和 Buff 系统
- **Status:** completed
- **Output:** `systems/combat/`

### Task 21: 实现 AI 行为和输入/连招系统
- **Status:** completed
- **Output:** `systems/ai/`, `systems/input/`

---

## Phase 7: DLC 系统

### Task 22: 实现 DLC 管理器核心
- **Status:** completed
- **Output:** `addons/dlc_manager/core/`

### Task 23: 实现 DLC 与 ECS/资产集成层
- **Status:** completed
- **Output:** `addons/dlc_manager/integration/`
