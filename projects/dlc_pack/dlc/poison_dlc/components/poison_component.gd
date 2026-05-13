class_name PoisonComponent extends Resource
## 毒素 Component（DLC 提供）
##
## 注意：此脚本在 DLC PCK 中，加载后通过 res://dlc/poison_dlc/components/poison_component.gd 访问

var is_poisoned: bool = false
@export var damage_per_second: float = 5.0
var remaining_time: float = 0.0
var source_id: int = -1


func get_component_name() -> StringName:
	return &"Poison"


func apply_poison(duration: float, dps: float, from_id: int = -1) -> void:
	is_poisoned = true
	remaining_time = duration
	damage_per_second = dps
	source_id = from_id
	print("[DLC PoisonComponent] 毒素施加：%.1f 秒，%.1f DPS" % [duration, dps])
