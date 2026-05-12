class_name StatModifier extends Resource
## 属性修饰器
##
## 定义一条属性修饰规则，可来自装备、Buff、被动技能等。
## 支持三种计算类型：固定值加成、百分比加成、百分比乘算。

## 加成类型枚举
enum ModType {
	FLAT_ADD,       ## 固定值加成：最终值 += value
	PERCENT_ADD,    ## 百分比加成：最终值 *= (1 + 所有百分比之和)
	PERCENT_MULT,   ## 百分比乘算：最终值 *= (1 + value)，独立乘算
}

## 目标属性名（如 &"attack"、&"defense"、&"max_hp"）
@export var stat_name: StringName = &""

## 加成类型
@export var mod_type: ModType = ModType.FLAT_ADD

## 数值（FLAT_ADD: 绝对值, PERCENT: 小数形式如 0.1 = 10%）
@export var value: float = 0.0

## 来源标识（用于按来源批量移除，如 "buff_poison"、"equip_sword"）
@export var source: StringName = &""

## 修饰器优先级（同类型内的应用顺序，数值小的先应用）
@export var order: int = 0
