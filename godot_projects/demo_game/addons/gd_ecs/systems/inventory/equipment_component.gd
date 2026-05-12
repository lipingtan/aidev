class_name EquipmentComponent extends EcsComponent
## 装备栏 Component
##
## 管理角色当前装备的物品，提供装备/卸下操作。
## 装备变更时自动标记 FinalStats 需要重算。

## 装备变更时触发
signal equipment_changed(slot: int, old_item: ItemData, new_item: ItemData)

## 装备槽 { EquipSlot 枚举值: ItemData }
@export var slots: Dictionary = {}


func get_component_name() -> StringName:
	return &"Equipment"


## 装备物品（返回被替换的旧物品，null 表示槽位原本为空）
func equip(item: ItemData, equip_stats: EquipmentStats) -> ItemData:
	var slot: int = equip_stats.equip_slot
	var old_item: ItemData = slots.get(slot)
	slots[slot] = item
	equipment_changed.emit(slot, old_item, item)
	return old_item


## 卸下指定槽位的装备（返回卸下的物品）
func unequip(slot: int) -> ItemData:
	if slot not in slots:
		return null
	var item: ItemData = slots[slot]
	slots.erase(slot)
	equipment_changed.emit(slot, item, null)
	return item


## 获取指定槽位的装备
func get_equipped(slot: int) -> ItemData:
	return slots.get(slot)


## 检查槽位是否有装备
func has_equipped(slot: int) -> bool:
	return slot in slots


## 获取所有装备提供的属性修饰器
func get_all_modifiers() -> Array[StatModifier]:
	var result: Array[StatModifier] = []
	for item in slots.values():
		# 从 ItemData 的 metadata 中获取 EquipmentStats
		if item and item.has_meta("equipment_stats"):
			var stats: EquipmentStats = item.get_meta("equipment_stats")
			result.append_array(stats.stat_modifiers)
	return result
