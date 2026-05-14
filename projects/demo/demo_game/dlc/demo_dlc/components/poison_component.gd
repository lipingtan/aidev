class_name PoisonComponent extends EcsComponent
## 毒素 Component（由 Demo DLC 提供）
##
## 存储角色当前的中毒状态。
## 由 PoisonSystem 每秒读取并造成持续伤害。

## 是否处于中毒状态
var is_poisoned: bool = false

## 毒素伤害（每秒）
@export var damage_per_second: float = 5.0

## 中毒剩余时间（秒）
var remaining_time: float = 0.0

## 毒素来源 Entity ID
var source_id: int = -1


func get_component_name() -> StringName:
	return &"Poison"


## 施加中毒效果
func apply_poison(duration: float, dps: float, from_id: int = -1) -> void:
	is_poisoned = true
	remaining_time = duration
	damage_per_second = dps
	source_id = from_id
	print("[DLC] 毒素施加：持续 %.1f 秒，每秒 %.1f 伤害" % [duration, dps])
