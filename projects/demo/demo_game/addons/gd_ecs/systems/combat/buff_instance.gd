class_name BuffInstance extends RefCounted
## Buff 实例（运行时状态）
##
## 存储单个 Buff 的当前激活状态，包括剩余时间、层数和计时器。

## 对应的 Buff 静态定义
var buff_data: BuffData = null

## 剩余持续时间（秒，-1 = 永久）
var remaining_time: float = 0.0

## 当前堆叠层数
var stack_count: int = 1

## 周期触发计时器
var tick_timer: float = 0.0

## 施加此 Buff 的来源 Entity ID
var source_entity_id: int = -1
