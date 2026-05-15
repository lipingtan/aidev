# RPG 背包 / 容器系统

## 目标

**容器（槽位 + 堆叠/重量规则）** 与 **物品定义（ItemData）** 分离；装备位是容器的特化视图。

## 要点

| 项 | 说明 |
|----|------|
| ItemData | id、堆叠上限、类型（消耗/装备/任务）、图标、扩展属性 |
| ItemContainer | 槽数组、`can_add` / `try_stack`；装备映射身体槽位 |
| 与属性 | 装备变更 → 增删 StatModifier → invalidate Final |

## 验收

- [ ] 满包拾取、唯一物品、丢弃边界正确
- [ ] UI 与数据层：`inventory_changed` 等信号协议固定

## 常见陷阱

- UI 缓存旧的 Item 引用，数据已 swap 导致幽灵格子
