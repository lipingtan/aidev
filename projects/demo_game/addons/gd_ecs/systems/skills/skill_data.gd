class_name SkillData extends Resource
## 技能静态定义
##
## 策划配置数据，定义技能的效果、消耗、冷却等。

## 技能类型枚举
enum SkillType {
	ACTIVE,   ## 主动技能（需要手动释放）
	PASSIVE,  ## 被动技能（始终生效）
	TOGGLE,   ## 切换技能（开/关状态）
}

## 效果类型枚举
enum EffectType {
	DAMAGE,     ## 造成伤害
	HEAL,       ## 恢复生命
	BUFF,       ## 施加增益
	DEBUFF,     ## 施加减益
	KNOCKBACK,  ## 击退
	STUN,       ## 眩晕
}

## 目标类型枚举
enum TargetType {
	SELF,          ## 自身
	SINGLE_ENEMY,  ## 单个敌人
	AOE_CIRCLE,    ## 圆形范围
	AOE_CONE,      ## 扇形范围
	ALL_ALLIES,    ## 所有友方
}

## 唯一标识符
@export var id: StringName = &""

## 显示名称
@export var name: String = ""

## 技能图标
@export var icon: Texture2D = null

## 描述
@export_multiline var description: String = ""

## 技能类型
@export var skill_type: SkillType = SkillType.ACTIVE

## 冷却时间（秒）
@export var cooldown: float = 1.0

## MP 消耗
@export var mp_cost: float = 10.0

## 学习等级要求
@export var required_level: int = 1

## 伤害倍率（基于攻击力）
@export var damage_multiplier: float = 1.5

## 效果类型
@export var effect_type: EffectType = EffectType.DAMAGE

## 目标类型
@export var target_type: TargetType = TargetType.SINGLE_ENEMY

## 效果范围（AOE 时使用）
@export var effect_radius: float = 3.0

## 效果持续时间（Buff/Debuff 时使用）
@export var effect_duration: float = 0.0

## 被动技能提供的属性修饰器
@export var passive_modifiers: Array[StatModifier] = []

## 关联的 Buff ID（施加 Buff 时使用）
@export var buff_id: StringName = &""
