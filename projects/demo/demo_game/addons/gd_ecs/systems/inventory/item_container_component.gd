class_name ItemContainerComponent extends EcsComponent
## 通用物品容器 Component
##
## 背包、储物柜、商店、宝箱、掉落物都使用此 Component。
## 通过 container_type 和配置参数区分不同容器的行为。

## 容器内容变更时触发
signal container_changed()

## 容器类型枚举
enum ContainerType {
	BACKPACK,    ## 角色背包
	EQUIPMENT,   ## 装备栏
	STORAGE,     ## 储物柜
	SHOP,        ## 商店
	LOOT,        ## 掉落物/宝箱
	TRADE,       ## 交易窗口
}

## 容器唯一标识
@export var container_id: StringName = &""

## 容器类型
@export var container_type: ContainerType = ContainerType.BACKPACK

## 最大格子数
@export var capacity: int = 40

## 格子列表
@export var slots: Array[ItemSlot] = []

## 允许放入的物品类型过滤器（空 = 不限制）
@export var type_filter: Array[ItemData.ItemType] = []

## 是否只读（商店货架）
@export var is_readonly: bool = false


func get_component_name() -> StringName:
	return &"ItemContainer"


func _init() -> void:
	# 初始化空格子
	_ensure_slots()


## 确保格子数量与容量一致
func _ensure_slots() -> void:
	while slots.size() < capacity:
		slots.append(ItemSlot.new())


## 添加物品（返回实际添加的数量）
func add_item(item: ItemData, count: int = 1) -> int:
	if is_readonly:
		return 0
	if type_filter.size() > 0 and item.item_type not in type_filter:
		return 0
	
	var remaining: int = count
	
	# 先尝试堆叠到已有同类物品
	if item.stackable:
		for slot in slots:
			if remaining <= 0:
				break
			if slot.item and slot.item.id == item.id and slot.count < item.max_stack:
				var can_add: int = mini(remaining, item.max_stack - slot.count)
				slot.count += can_add
				remaining -= can_add
	
	# 再放入空格
	for slot in slots:
		if remaining <= 0:
			break
		if slot.is_empty():
			slot.item = item
			if item.stackable:
				var can_add: int = mini(remaining, item.max_stack)
				slot.count = can_add
				remaining -= can_add
			else:
				slot.count = 1
				remaining -= 1
	
	if remaining < count:
		container_changed.emit()
	return count - remaining


## 从指定格子移除物品（返回移除的 ItemSlot 副本）
func remove_at(slot_index: int, count: int = 1) -> ItemSlot:
	if slot_index < 0 or slot_index >= slots.size():
		return null
	var slot: ItemSlot = slots[slot_index]
	if slot.is_empty():
		return null
	
	var result := ItemSlot.new()
	result.item = slot.item
	result.count = mini(count, slot.count)
	
	slot.count -= result.count
	if slot.count <= 0:
		slot.clear()
	
	container_changed.emit()
	return result


## 检查是否能添加物品
func can_add(item: ItemData) -> bool:
	if is_readonly:
		return false
	if type_filter.size() > 0 and item.item_type not in type_filter:
		return false
	# 检查是否有空间（空格或可堆叠格）
	for slot in slots:
		if slot.is_empty():
			return true
		if item.stackable and slot.item and slot.item.id == item.id and slot.count < item.max_stack:
			return true
	return false


## 获取物品总数（按 ID）
func get_item_count(item_id: StringName) -> int:
	var total: int = 0
	for slot in slots:
		if slot.item and slot.item.id == item_id:
			total += slot.count
	return total


## 排序（按类型 → 稀有度 → 名称）
func sort_items() -> void:
	# 收集非空槽位
	var items: Array[ItemSlot] = []
	for slot in slots:
		if not slot.is_empty():
			items.append(slot.duplicate())
	
	# 排序
	items.sort_custom(func(a, b):
		if a.item.item_type != b.item.item_type:
			return a.item.item_type < b.item.item_type
		if a.item.rarity != b.item.rarity:
			return a.item.rarity > b.item.rarity
		return a.item.name < b.item.name
	)
	
	# 重新填充
	for i in slots.size():
		if i < items.size():
			slots[i] = items[i]
		else:
			slots[i] = ItemSlot.new()
	
	container_changed.emit()
