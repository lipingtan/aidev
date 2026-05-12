class_name SkillCooldownSystem extends EcsSystem
## 技能冷却递减系统
##
## 每帧递减所有技能的剩余冷却时间，到 0 时移除。

func _init() -> void:
	system_name = &"SkillCooldown"
	priority = 5


func get_query() -> Array[StringName]:
	return [&"SkillSet"]


func process(entities: Array[EcsEntity], delta: float) -> void:
	for entity in entities:
		var skill_set: SkillSetComponent = entity.get_component(&"SkillSet")
		var expired: Array[StringName] = []
		
		for skill_id in skill_set.cooldowns:
			skill_set.cooldowns[skill_id] -= delta
			if skill_set.cooldowns[skill_id] <= 0.0:
				expired.append(skill_id)
		
		for skill_id in expired:
			skill_set.cooldowns.erase(skill_id)
