class_name RuntimeStatsComponent extends EcsComponent
## 运行时状态 Component
##
## 存储角色的当前运行时状态（当前 HP/MP 等），
## 区别于 FinalStats（最大值/能力上限）。

## 当生命值变化时触发
signal hp_changed(old_value: float, new_value: float)

## 当魔法值变化时触发
signal mp_changed(old_value: float, new_value: float)

## 当角色死亡时触发
signal died()

## 当前生命值
@export var current_hp: float = 100.0

## 当前魔法值
@export var current_mp: float = 50.0

## 是否存活
var is_alive: bool = true


func get_component_name() -> StringName:
	return &"RuntimeStats"


## 扣减生命值（返回实际扣减量）
func take_damage(amount: float) -> float:
	if not is_alive:
		return 0.0
	var old_hp: float = current_hp
	current_hp = maxf(0.0, current_hp - amount)
	hp_changed.emit(old_hp, current_hp)
	if current_hp <= 0.0:
		is_alive = false
		died.emit()
	return old_hp - current_hp


## 恢复生命值（需要 FinalStats 提供上限）
func heal(amount: float, max_hp: float) -> float:
	if not is_alive:
		return 0.0
	var old_hp: float = current_hp
	current_hp = minf(max_hp, current_hp + amount)
	hp_changed.emit(old_hp, current_hp)
	return current_hp - old_hp


## 消耗魔法值（返回是否成功）
func spend_mp(amount: float) -> bool:
	if current_mp < amount:
		return false
	var old_mp: float = current_mp
	current_mp -= amount
	mp_changed.emit(old_mp, current_mp)
	return true


## 恢复魔法值
func restore_mp(amount: float, max_mp: float) -> float:
	var old_mp: float = current_mp
	current_mp = minf(max_mp, current_mp + amount)
	mp_changed.emit(old_mp, current_mp)
	return current_mp - old_mp


## 重置为满状态（复活/关卡开始时）
func reset_to_full(max_hp: float, max_mp: float) -> void:
	current_hp = max_hp
	current_mp = max_mp
	is_alive = true
