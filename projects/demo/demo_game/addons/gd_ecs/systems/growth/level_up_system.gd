class_name LevelUpSystem extends EcsSystem
## 升级判定系统
##
## 检查经验值是否满足升级条件，执行升级逻辑：
## - 增加基础属性（按成长曲线）
## - 解锁新技能
## - 标记属性需要重算
##
## 依赖 Component：
## - Experience: 经验和等级数据
## - BaseStats: 基础属性（升级时增加）
## - FinalStats: 标记 dirty
## - SkillSet（可选）: 解锁新技能

func _init() -> void:
	system_name = &"LevelUp"
	priority = 1  # 在属性计算之后检查


func get_query() -> Array[StringName]:
	return [&"Experience", &"BaseStats"]


func process(entities: Array, delta: float) -> void:
	for entity in entities:
		var exp_comp: ExperienceComponent = entity.get_component(&"Experience")
		if not exp_comp.growth_profile:
			continue
		
		var required: float = exp_comp.get_required_exp()
		var max_level: int = exp_comp.growth_profile.max_level
		
		# 循环检查是否可升级（可能一次获得大量经验连升多级）
		while exp_comp.current_exp >= required and exp_comp.level < max_level:
			exp_comp.current_exp -= required
			exp_comp.level += 1
			_apply_level_up(entity, exp_comp)
			required = exp_comp.get_required_exp()


## 执行升级逻辑
func _apply_level_up(entity: Node, exp_comp: ExperienceComponent) -> void:
	var base: BaseStatsComponent = entity.get_component(&"BaseStats")
	var growth: Dictionary = exp_comp.growth_profile.stat_growth
	
	# 按成长曲线增加基础属性
	for stat_name in growth:
		var current: float = base.get_stat(StringName(stat_name))
		base.set_stat(StringName(stat_name), current + growth[stat_name])
	
	# 解锁新技能
	var unlocks: Array = exp_comp.growth_profile.get_unlocked_skills(exp_comp.level)
	if unlocks.size() > 0 and entity.has_component(&"SkillSet"):
		var skills: SkillSetComponent = entity.get_component(&"SkillSet")
		for skill_id in unlocks:
			skills.learn_skill(StringName(skill_id))
	
	# 标记属性需要重算
	if entity.has_component(&"FinalStats"):
		var final_stats: FinalStatsComponent = entity.get_component(&"FinalStats")
		final_stats.mark_dirty()
	
	# 发送升级事件
	exp_comp.leveled_up.emit(exp_comp.level)
	if EventBus:
		EventBus.emit_event(&"player_leveled_up", {
			"entity_id": entity.entity_id,
			"new_level": exp_comp.level,
		})
