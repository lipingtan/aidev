class_name ElementTable extends Resource
## 元素克制表
##
## 定义元素间的伤害倍率关系。
## 1.0 = 正常，1.5 = 克制，0.5 = 抵抗，0.0 = 免疫

## 克制倍率表 { "攻击元素_防御元素": 倍率 }
@export var multipliers: Dictionary = {
	# 火克冰，冰克雷，雷克水，暗克圣，圣克暗
	"FIRE_ICE": 1.5,
	"ICE_FIRE": 0.5,
	"ICE_LIGHTNING": 1.5,
	"LIGHTNING_ICE": 0.5,
	"SHADOW_HOLY": 1.5,
	"HOLY_SHADOW": 1.5,
	# 同元素抵抗
	"FIRE_FIRE": 0.5,
	"ICE_ICE": 0.5,
	"LIGHTNING_LIGHTNING": 0.5,
	"SHADOW_SHADOW": 0.5,
	"HOLY_HOLY": 0.5,
}


## 获取元素克制倍率
func get_multiplier(attack_element: int, defense_element: int) -> float:
	if attack_element == DamageEventComponent.ElementType.NONE:
		return 1.0
	if defense_element == DamageEventComponent.ElementType.NONE:
		return 1.0
	
	var key: String = "%s_%s" % [
		DamageEventComponent.ElementType.keys()[attack_element],
		DamageEventComponent.ElementType.keys()[defense_element],
	]
	return multipliers.get(key, 1.0)
