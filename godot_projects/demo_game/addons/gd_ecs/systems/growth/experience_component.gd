class_name ExperienceComponent extends EcsComponent
## 经验/等级 Component
##
## 管理角色的经验值和等级，关联成长曲线。

## 升级时触发
signal leveled_up(new_level: int)

## 当前等级
@export var level: int = 1

## 当前经验值
@export var current_exp: float = 0.0

## 累计总经验
@export var total_exp: float = 0.0

## 成长曲线引用
@export var growth_profile: GrowthProfile = null


func get_component_name() -> StringName:
	return &"Experience"


## 增加经验值（可能触发升级）
func add_exp(amount: float) -> void:
	current_exp += amount
	total_exp += amount


## 获取当前等级升级所需经验
func get_required_exp() -> float:
	if growth_profile:
		return growth_profile.get_required_exp(level)
	return 100.0 * level


## 获取升级进度百分比（0~1）
func get_progress() -> float:
	var required: float = get_required_exp()
	if required <= 0.0:
		return 1.0
	return clampf(current_exp / required, 0.0, 1.0)
