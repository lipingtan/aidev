class_name AggroTableComponent extends EcsComponent
## 仇恨表 Component
##
## 记录敌人对各目标的仇恨值，用于决定攻击优先级。
## 仇恨值随时间自然衰减。

## 仇恨条目 { entity_id: aggro_value }
var entries: Dictionary = {}

## 每秒仇恨衰减量
@export var decay_rate: float = 1.0

## 最大仇恨值上限
@export var max_aggro: float = 1000.0


func get_component_name() -> StringName:
	return &"AggroTable"


## 增加对指定目标的仇恨值
func add_aggro(entity_id: int, amount: float) -> void:
	var current: float = entries.get(entity_id, 0.0)
	entries[entity_id] = minf(current + amount, max_aggro)


## 获取仇恨值最高的目标 ID（-1 = 无目标）
func get_top_target() -> int:
	var max_val: float = 0.0
	var target: int = -1
	for id in entries:
		if entries[id] > max_val:
			max_val = entries[id]
			target = id
	return target


## 获取指定目标的仇恨值
func get_aggro(entity_id: int) -> float:
	return entries.get(entity_id, 0.0)


## 清除所有仇恨
func clear() -> void:
	entries.clear()


## 移除指定目标的仇恨
func remove_target(entity_id: int) -> void:
	entries.erase(entity_id)
