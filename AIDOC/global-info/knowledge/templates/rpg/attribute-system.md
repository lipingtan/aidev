# RPG 属性系统

## 目标

统一 **基础属性（Base）**、**运行时修饰（Modifier）**、**最终战斗用数值（Final）**，避免技能/装备/Buff 各算一套公式。

## 核心抽象

| 概念 | 职责 |
|------|------|
| BaseStats | 等级表、职业成长、裸装面板 |
| StatModifier | 来源（装备/Buff/环境）、加算或乘算、持续帧或永久 |
| FinalStats | 每帧或事件前由结算系统从 Base + Modifiers 生成，**只读**供战斗/AI |

**约定**：伤害、治疗、暴击等只读 FinalStats；改装备/Buff 时 invalidate，下一 tick 重算。

## 与工程对齐

- 与 `addons/gd_ecs`：`BaseStatsComponent` / `FinalStatsComponent` / `StatModifier` 资源同形态即可。
- **NSFW**：好感、堕落等走 `templates/nsfw/relationship-stats.md`，勿与 attack/defense 混用同一表，除非剧情技能明确需要。

## 验收

- [ ] 换装/上 Buff 后 Final 可预测、可单测
- [ ] 无「读 Base 直接进伤害公式」的旁路
- [ ] 存档存 Base 与 Modifier 来源，不存 Final（或仅存校验快照）

## 常见陷阱

- 修饰器顺序未文档化（加算/乘算顺序须全局统一）
- 死亡/切图未清空临时 Modifier 导致数值泄漏
