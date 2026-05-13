class_name BuffTickSystem extends EcsSystem
## Buff Tick 系统
##
## 每帧处理所有 Buff 的持续时间递减、周期效果触发和过期移除。
## 同时处理无敌帧和连击重置计时。
##
## 依赖 Component：
## - BuffList: Buff 列表
## - CombatState（可选）: 无敌帧/连击计时
## - RuntimeStats（可选）: 周期伤害/治疗
## - FinalStats（可选）: 属性重算标记


func _init() -> void:
	system_name = &"BuffTick"
	priority = 5
	phase = &"process"


func get_query() -> Array[StringName]:
	return [&"BuffList"]


func process(entities: Array, delta: float) -> void:
	for entity in entities:
		var buff_list: BuffListComponent = entity.get_component(&"BuffList")
		var stats_dirty: bool = false
		var expired: Array[int] = []
		
		for i in buff_list.active_buffs.size():
			var inst: BuffInstance = buff_list.active_buffs[i]
			
			# 永久 Buff 不递减
			if inst.buff_data.duration < 0.0:
				continue
			
			# 递减持续时间
			inst.remaining_time -= delta
			
			# 周期效果触发
			if inst.buff_data.tick_interval > 0.0:
				inst.tick_timer += delta
				if inst.tick_timer >= inst.buff_data.tick_interval:
					inst.tick_timer -= inst.buff_data.tick_interval
					_apply_tick_effect(entity, inst)
			
			# 过期检查
			if inst.remaining_time <= 0.0:
				expired.append(i)
				# 如果有属性修饰器，需要重算
				if inst.buff_data.stat_modifiers.size() > 0:
					stats_dirty = true
		
		# 倒序移除过期 Buff
		for i in range(expired.size() - 1, -1, -1):
			var idx: int = expired[i]
			var buff_id: StringName = buff_list.active_buffs[idx].buff_data.id
			buff_list.active_buffs.remove_at(idx)
			buff_list.buff_removed.emit(buff_id, &"expired")
		
		# 标记属性需要重算
		if stats_dirty and entity.has_component(&"FinalStats"):
			(entity.get_component(&"FinalStats") as FinalStatsComponent).mark_dirty()
		
		# 更新战斗状态计时
		if entity.has_component(&"CombatState"):
			_tick_combat_state(entity.get_component(&"CombatState"), delta)


## 应用周期效果
func _apply_tick_effect(entity: Node, inst: BuffInstance) -> void:
	if not entity.has_component(&"RuntimeStats"):
		return
	var runtime: RuntimeStatsComponent = entity.get_component(&"RuntimeStats")
	
	# 周期伤害
	if inst.buff_data.tick_damage > 0.0:
		var dmg_event := DamageEventComponent.new()
		dmg_event.damage_amount = inst.buff_data.tick_damage * inst.stack_count
		dmg_event.damage_type = DamageEventComponent.DamageType.MAGICAL
		dmg_event.element = inst.buff_data.tick_element
		dmg_event.attacker_id = inst.source_entity_id
		entity.add_component(dmg_event)
	
	# 周期治疗
	if inst.buff_data.tick_heal > 0.0 and entity.has_component(&"FinalStats"):
		var final_stats: FinalStatsComponent = entity.get_component(&"FinalStats")
		runtime.heal(inst.buff_data.tick_heal * inst.stack_count, final_stats.max_hp)


## 更新战斗状态计时
func _tick_combat_state(combat: CombatStateComponent, delta: float) -> void:
	combat.state_time += delta
	combat.tick_invincible(delta)
	
	# 连击重置
	if combat.combo_count > 0:
		combat.combo_reset_timer -= delta
		if combat.combo_reset_timer <= 0.0:
			combat.reset_combo()
