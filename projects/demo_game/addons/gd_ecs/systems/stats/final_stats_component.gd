class_name FinalStatsComponent extends EcsComponent
## 最终属性 Component
##
## 经过所有修饰器（装备、Buff、被动技能）计算后的最终属性值。
## 由 StatsCalculationSystem 在 is_dirty 时自动重算。
## 其他系统应读取此 Component 获取角色的实际属性。

@export_group("生命与魔法")
@export var max_hp: float = 0.0
@export var max_mp: float = 0.0

@export_group("攻防")
@export var attack: float = 0.0
@export var defense: float = 0.0
@export var magic_attack: float = 0.0
@export var magic_defense: float = 0.0

@export_group("速度与暴击")
@export var speed: float = 0.0
@export var attack_speed: float = 0.0
@export var crit_rate: float = 0.0
@export var crit_damage: float = 0.0

@export_group("抗性")
@export var fire_resist: float = 0.0
@export var ice_resist: float = 0.0
@export var lightning_resist: float = 0.0
@export var shadow_resist: float = 0.0

## 标记是否需要重算（装备变更/升级/Buff 变化时设为 true）
var is_dirty: bool = true


func get_component_name() -> StringName:
	return &"FinalStats"


## 获取指定属性的最终值
func get_stat(stat_name: StringName) -> float:
	match stat_name:
		&"max_hp": return max_hp
		&"max_mp": return max_mp
		&"attack": return attack
		&"defense": return defense
		&"magic_attack": return magic_attack
		&"magic_defense": return magic_defense
		&"speed": return speed
		&"attack_speed": return attack_speed
		&"crit_rate": return crit_rate
		&"crit_damage": return crit_damage
		&"fire_resist": return fire_resist
		&"ice_resist": return ice_resist
		&"lightning_resist": return lightning_resist
		&"shadow_resist": return shadow_resist
	return 0.0


## 设置指定属性的最终值（由 StatsCalculationSystem 调用）
func set_stat(stat_name: StringName, value: float) -> void:
	match stat_name:
		&"max_hp": max_hp = value
		&"max_mp": max_mp = value
		&"attack": attack = value
		&"defense": defense = value
		&"magic_attack": magic_attack = value
		&"magic_defense": magic_defense = value
		&"speed": speed = value
		&"attack_speed": attack_speed = value
		&"crit_rate": crit_rate = value
		&"crit_damage": crit_damage = value
		&"fire_resist": fire_resist = value
		&"ice_resist": ice_resist = value
		&"lightning_resist": lightning_resist = value
		&"shadow_resist": shadow_resist = value


## 标记需要重算
func mark_dirty() -> void:
	is_dirty = true
