class_name StatsCalculationSystem extends EcsSystem
## 属性计算系统
##
## 汇总所有修饰器来源（装备、Buff、被动技能），计算最终属性。
## 仅在 FinalStats.is_dirty 时执行重算，避免每帧计算。
##
## 计算公式：
## final = (base + flat_sum) * (1 + percent_add_sum) * percent_mult_product
##
## 依赖 Component：
## - BaseStats: 基础属性值
## - FinalStats: 计算结果存储
## - Equipment（可选）: 装备提供的修饰器
## - BuffList（可选）: Buff 提供的修饰器
## - SkillSet（可选）: 被动技能提供的修饰器


func _init() -> void:
	system_name = &"StatsCalculation"
	priority = 0  # 最先执行，确保其他系统读到最新属性


func get_query() -> Array[StringName]:
	return [&"BaseStats", &"FinalStats"]


func process(entities: Array, delta: float) -> void:
	for entity in entities:
		var final_stats: FinalStatsComponent = entity.get_component(&"FinalStats")
		if not final_stats.is_dirty:
			continue
		
		var base_stats: BaseStatsComponent = entity.get_component(&"BaseStats")
		_recalculate(entity, base_stats, final_stats)
		final_stats.is_dirty = false


## 重新计算所有属性
func _recalculate(entity: Node, base: BaseStatsComponent, final: FinalStatsComponent) -> void:
	# 收集所有修饰器
	var all_modifiers: Array[StatModifier] = _collect_modifiers(entity)
	
	# 对每个属性执行计算
	for stat_name in base.get_all_stat_names():
		var base_value: float = base.get_stat(stat_name)
		var final_value: float = _calculate_stat(base_value, stat_name, all_modifiers)
		final.set_stat(stat_name, final_value)


## 收集所有来源的修饰器
func _collect_modifiers(entity: Node) -> Array[StatModifier]:
	var modifiers: Array[StatModifier] = []
	
	# 从装备收集
	if entity.has_component(&"Equipment"):
		var equipment = entity.get_component(&"Equipment")
		if equipment.has_method("get_all_modifiers"):
			modifiers.append_array(equipment.get_all_modifiers())
	
	# 从 Buff 收集
	if entity.has_component(&"BuffList"):
		var buff_list = entity.get_component(&"BuffList")
		if buff_list.has_method("get_all_modifiers"):
			modifiers.append_array(buff_list.get_all_modifiers())
	
	# 从被动技能收集
	if entity.has_component(&"SkillSet"):
		var skill_set = entity.get_component(&"SkillSet")
		if skill_set.has_method("get_passive_modifiers"):
			modifiers.append_array(skill_set.get_passive_modifiers())
	
	return modifiers


## 计算单个属性的最终值
func _calculate_stat(base_value: float, stat_name: StringName, modifiers: Array[StatModifier]) -> float:
	var flat_sum: float = 0.0
	var percent_add_sum: float = 0.0
	var percent_mult_product: float = 1.0
	
	for mod in modifiers:
		if mod.stat_name != stat_name:
			continue
		match mod.mod_type:
			StatModifier.ModType.FLAT_ADD:
				flat_sum += mod.value
			StatModifier.ModType.PERCENT_ADD:
				percent_add_sum += mod.value
			StatModifier.ModType.PERCENT_MULT:
				percent_mult_product *= (1.0 + mod.value)
	
	# 最终公式
	var result: float = (base_value + flat_sum) * (1.0 + percent_add_sum) * percent_mult_product
	return result
