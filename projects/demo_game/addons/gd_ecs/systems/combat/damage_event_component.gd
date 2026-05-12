class_name DamageEventComponent extends EcsComponent
## 伤害事件 Component（一次性，处理后移除）
##
## 由攻击方生成并挂载到受击方 Entity 上，
## DamageSystem 处理后自动移除。

## 伤害类型枚举
enum DamageType {
	PHYSICAL,  ## 物理伤害（受防御减免）
	MAGICAL,   ## 魔法伤害（受魔防减免）
	TRUE,      ## 真实伤害（无视防御）
}

## 元素类型枚举
enum ElementType {
	NONE,       ## 无元素
	FIRE,       ## 火
	ICE,        ## 冰
	LIGHTNING,  ## 雷
	SHADOW,     ## 暗
	HOLY,       ## 圣
}

## 基础伤害量
@export var damage_amount: float = 0.0

## 伤害类型
@export var damage_type: DamageType = DamageType.PHYSICAL

## 元素类型
@export var element: ElementType = ElementType.NONE

## 攻击方 Entity ID
@export var attacker_id: int = -1

## 触发此伤害的技能 ID
@export var skill_id: StringName = &""

## 是否暴击
@export var is_crit: bool = false

## 击退力（Vector3.ZERO = 不击退）
@export var knockback_force: Vector3 = Vector3.ZERO

## 命中位置（用于特效生成）
@export var hit_position: Vector3 = Vector3.ZERO


func get_component_name() -> StringName:
	return &"DamageEvent"
