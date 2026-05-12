class_name ItemSlot extends Resource
## 物品槽位
##
## 容器中的一个格子，存储物品引用和数量。

## 物品数据引用（null = 空格）
@export var item: ItemData = null

## 堆叠数量
@export var count: int = 0

## 物品实例数据（强化等级、附魔等个体差异，可选）
@export var instance_uid: String = ""


## 是否为空
func is_empty() -> bool:
	return item == null or count <= 0


## 清空槽位
func clear() -> void:
	item = null
	count = 0
	instance_uid = ""


## 能否与另一个槽位堆叠
func can_stack_with(other: ItemSlot) -> bool:
	if is_empty() or other.is_empty():
		return false
	if item.id != other.item.id:
		return false
	if not item.stackable:
		return false
	return count < item.max_stack
