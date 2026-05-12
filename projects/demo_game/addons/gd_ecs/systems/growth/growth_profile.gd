class_name GrowthProfile extends Resource
## 成长曲线定义
##
## 每个角色/职业一份，定义升级时各属性的增长值和技能解锁表。

## 每升一级各属性增长值 { stat_name: growth_per_level }
@export var stat_growth: Dictionary = {
	"max_hp": 15.0,
	"max_mp": 8.0,
	"attack": 3.0,
	"defense": 2.0,
	"magic_attack": 2.5,
	"magic_defense": 1.5,
	"speed": 0.5,
}

## 升级所需经验曲线（X轴=等级比例0~1，Y轴=所需经验）
@export var exp_curve: Curve = null

## 技能解锁表 { level: [skill_id] }
@export var skill_unlock_table: Dictionary = {}

## 最大等级
@export var max_level: int = 99

## 基础经验值（1级升2级所需）
@export var base_exp: float = 100.0

## 经验增长系数
@export var exp_growth_rate: float = 1.2


## 获取指定等级升级所需经验
func get_required_exp(level: int) -> float:
	if exp_curve:
		var t: float = float(level) / float(max_level)
		return exp_curve.sample(t) * base_exp * max_level
	# 无曲线时使用指数公式
	return base_exp * pow(exp_growth_rate, level - 1)


## 获取指定等级解锁的技能列表
func get_unlocked_skills(level: int) -> Array:
	return skill_unlock_table.get(level, [])
