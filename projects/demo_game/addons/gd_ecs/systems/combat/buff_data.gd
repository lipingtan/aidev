class_name BuffData extends Resource
## Buff/Debuff 静态定义
##
## 策划配置数据，通过 DataManager 按 ID 加载。

## 堆叠模式枚举
enum StackMode {
	REFRESH,           ## 刷新持续时间（不叠加层数）
	STACK_COUNT,       ## 叠加层数（刷新时间）
	STACK_INDEPENDENT, ## 独立计时（每次施加都是独立实例）
}

## 净化类型枚举
enum DispelType {
	MAGIC,        ## 可被魔法净化
	PHYSICAL,     ## 可被物理净化
	UNDISPELLABLE, ## 不可净化
}

## 唯一标识符
@export var id: StringName = &""

## 显示名称
@export var name: String = ""

## 图标
@export var icon: Texture2D = null

## 是否为 Debuff
@export var is_debuff: bool = false

## 持续时间（秒，-1 = 永久）
@export var duration: float = 5.0

## 周期触发间隔（秒，0 = 不触发）
@export var tick_interval: float = 0.0

## 最大堆叠层数
@export var max_stacks: int = 1

## 堆叠模式
@export var stack_mode: StackMode = StackMode.REFRESH

## 净化类型
@export var dispel_type: DispelType = DispelType.MAGIC

## 属性修饰器（持续生效）
@export var stat_modifiers: Array[StatModifier] = []

## 周期触发效果（每次 tick 时）
@export var tick_damage: float = 0.0
@export var tick_heal: float = 0.0
@export var tick_element: DamageEventComponent.ElementType = DamageEventComponent.ElementType.NONE

## 施加时触发的效果
@export var on_apply_stun: bool = false
@export var on_apply_knockback: Vector3 = Vector3.ZERO

## 免疫此 Buff 所需的标签（目标有这些标签则免疫）
@export var immunity_tags: Array[StringName] = []
