class_name BaseStatsComponent extends EcsComponent
## 基础属性 Component
##
## 存储角色的基础属性值（由等级和成长曲线决定）。
## 这些值是属性计算的起点，不含任何加成。

@export_group("生命与魔法")
@export var max_hp: float = 100.0
@export var max_mp: float = 50.0

@export_group("攻防")
@export var attack: float = 10.0
@export var defense: float = 5.0
@export var magic_attack: float = 8.0
@export var magic_defense: float = 4.0

@export_group("速度与暴击")
@export var speed: float = 3.0
@export var attack_speed: float = 1.0
@export var crit_rate: float = 0.05
@export var crit_damage: float = 1.5

@export_group("抗性")
@export var fire_resist: float = 0.0
@export var ice_resist: float = 0.0
@export var lightning_resist: float = 0.0
@export var shadow_resist: float = 0.0


func get_component_name() -> StringName:
	return &"BaseStats"


## 获取指定属性的基础值
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


## 设置指定属性的基础值
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


## 获取所有属性名列表
func get_all_stat_names() -> Array[StringName]:
	return [
		&"max_hp", &"max_mp",
		&"attack", &"defense", &"magic_attack", &"magic_defense",
		&"speed", &"attack_speed", &"crit_rate", &"crit_damage",
		&"fire_resist", &"ice_resist", &"lightning_resist", &"shadow_resist",
	]
