# 系统模板 · 选择与使用指南

> AI 在 feature-development-flow 的"设计计划"步骤中，根据功能需求查本表选择对应模板。
> 模板提供：目标、核心抽象、与工程对齐方式、验收标准、常见陷阱。

---

## 功能→模板映射表

| 蓝图中的功能模块 | 对应模板 | 何时读取 |
|-----------------|----------|----------|
| 属性系统（HP/MP/攻防） | `rpg/attribute-system.md` | 设计计划阶段 |
| 技能系统（主动/被动/冷却） | `rpg/skill-system.md` | 设计计划阶段 |
| 物品/装备/背包 | `rpg/inventory-system.md` | 设计计划阶段 |
| 对话系统 | `rpg/dialogue-system.md` | 设计计划阶段 |
| 任务系统（主线/支线） | `rpg/quest-system.md` | 设计计划阶段 |
| 存档系统 | `rpg/save-system.md` | 设计计划阶段 |
| 伤害计算/战斗系统 | `rpg/combat-system.md` | 设计计划阶段 |
| 角色移动控制器 | `action/character-controller.md` | 设计计划阶段 |
| 动画状态机 | `action/animation-fsm.md` | 设计计划阶段 |
| 连击系统 | `action/combo-system.md` | 设计计划阶段 |
| 闪避/格挡 | `action/dodge-block.md` | 设计计划阶段 |
| 锁定系统 | `action/lock-on.md` | 设计计划阶段 |
| 相机控制 | `action/camera-control.md` | 设计计划阶段 |
| 关系数值（好感/堕落） | `nsfw/relationship-stats.md` | 设计计划阶段 |
| H-Scene 系统 | `nsfw/h-scene-system.md` | 设计计划阶段 |
| CG Gallery | `nsfw/cg-gallery.md` | 设计计划阶段 |

---

## 使用流程

```
1. feature-development-flow → 2.3 设计计划
2. 确定功能属于哪个模块 → 查上表
3. 读取对应模板 → 获取：
   - 核心抽象（该用什么数据结构）
   - 与工程对齐（如何与 gd_ecs/EventBus 集成）
   - 验收标准（怎么算做完）
   - 常见陷阱（不要踩的坑）
4. 写入设计摘要的"架构方案"字段
5. 同时查 patterns/README.md 确定用什么设计模式
```

---

## 模板与其他规范的关系

| 模板提供 | 其他规范提供 |
|----------|-------------|
| 数据结构和核心抽象 | `patterns/README.md` → 设计模式选择 |
| 验收标准 | `experience-benchmarks.md` → 具体参数范围 |
| 与 ECS 的集成方式 | `godot/godot-engine.md` → Autoload 顺序 |
| 常见陷阱 | `iteration-workflow.md` → 踩坑后的修复流程 |

---

## 模板不覆盖的内容

模板只提供**架构约束**，以下内容由其他规范负责：

- 具体代码怎么写 → `godot/code-generation.md`
- 文件怎么组织 → `godot/godot-engine.md` 第六节
- 参数填什么值 → `experience-benchmarks.md`
- 性能预算多少 → `performance-budget.md`
