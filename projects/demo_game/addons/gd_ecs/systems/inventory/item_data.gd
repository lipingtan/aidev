class_name ItemData extends Resource
## 物品静态定义
##
## 策划配置数据，所有同类物品共享同一份 ItemData。
## 通过 DataManager 按 ID 加载。

## 物品类型枚举
enum ItemType {
	WEAPON,      ## 武器
	ARMOR,       ## 防具
	ACCESSORY,   ## 饰品
	CONSUMABLE,  ## 消耗品
	MATERIAL,    ## 材料
	QUEST,       ## 任务物品
}

## 唯一标识符
@export var id: StringName = &""

## 显示名称
@export var name: String = ""

## 物品图标
@export var icon: Texture2D = null

## 物品类型
@export var item_type: ItemType = ItemType.MATERIAL

## 稀有度（1~5）
@export_range(1, 5) var rarity: int = 1

## 是否可堆叠
@export var stackable: bool = false

## 最大堆叠数
@export var max_stack: int = 1

## 描述文本
@export_multiline var description: String = ""

## 购买价格（0 = 不可购买）
@export var buy_price: int = 0

## 出售价格（0 = 不可出售）
@export var sell_price: int = 0

## 使用等级要求
@export var required_level: int = 1
