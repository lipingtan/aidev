class_name SkillSetComponent extends EcsComponent
## 技能栏 Component
##
## 管理角色已学习的技能、装备到快捷栏的技能和冷却状态。

## 技能学习时触发
signal skill_learned(skill_id: StringName)

## 技能使用时触发
signal skill_used(skill_id: StringName)

## 已学习的技能 ID 列表
@export var learned_skills: Array[StringName] = []

## 装备到快捷栏的技能 ID 列表
@export var equipped_skills: Array[StringName] = []

## 冷却状态 { skill_id: remaining_cd }
var cooldowns: Dictionary = {}

## 技能等级 { skill_id: level }
@export var skill_levels: Dictionary = {}


func get_component_name() -> StringName:
	return &"SkillSet"


## 学习技能
func learn_skill(skill_id: StringName) -> bool:
	if skill_id in learned_skills:
		return false
	learned_skills.append(skill_id)
	skill_levels[skill_id] = 1
	skill_learned.emit(skill_id)
	return true


## 检查技能是否已学习
func has_skill(skill_id: StringName) -> bool:
	return skill_id in learned_skills


## 检查技能是否可使用（已学习且不在冷却中）
func can_use_skill(skill_id: StringName) -> bool:
	if skill_id not in learned_skills:
		return false
	if skill_id in cooldowns and cooldowns[skill_id] > 0.0:
		return false
	return true


## 使用技能（设置冷却）
func use_skill(skill_id: StringName, cooldown_time: float) -> void:
	cooldowns[skill_id] = cooldown_time
	skill_used.emit(skill_id)


## 获取技能剩余冷却时间
func get_cooldown(skill_id: StringName) -> float:
	return cooldowns.get(skill_id, 0.0)


## 获取技能等级
func get_skill_level(skill_id: StringName) -> int:
	return skill_levels.get(skill_id, 0)


## 升级技能
func upgrade_skill(skill_id: StringName) -> int:
	if skill_id not in skill_levels:
		return 0
	skill_levels[skill_id] += 1
	return skill_levels[skill_id]


## 获取被动技能提供的属性修饰器
func get_passive_modifiers() -> Array[StatModifier]:
	var result: Array[StatModifier] = []
	for skill_id in learned_skills:
		var skill: SkillData = DataManager.get_data(&"skills", skill_id)
		if skill and skill.skill_type == SkillData.SkillType.PASSIVE:
			result.append_array(skill.passive_modifiers)
	return result
