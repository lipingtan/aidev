class_name EquipmentStats extends Resource
## 装备属性加成定义
##
## 附加在 ItemData 上，定义装备提供的属性修饰和槽位信息。

## 装备槽位枚举
enum EquipSlot {
	WEAPON,       ## 武器
	HEAD,         ## 头部
	BODY,         ## 身体
	LEGS,         ## 腿部
	FEET,         ## 脚部
	ACCESSORY_1,  ## 饰品1
	ACCESSORY_2,  ## 饰品2
}

## 装备槽位
@export var equip_slot: EquipSlot = EquipSlot.WEAPON

## 属性修饰器列表
@export var stat_modifiers: Array[StatModifier] = []

## 装备等级要求
@export var required_level: int = 1
