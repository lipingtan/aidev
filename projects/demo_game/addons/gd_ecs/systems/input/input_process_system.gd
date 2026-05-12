class_name InputProcessSystem extends EcsSystem
## 输入处理系统
##
## 每帧清理过期输入，尝试匹配连招表，
## 将匹配结果写入 CombatStateComponent 供状态机读取。
##
## 依赖 Component：
## - InputBuffer: 输入缓冲队列
## - CombatState（可选）: 当前状态和连招结果写入目标

## 连招表（按优先级降序排列）
var combo_table: Array[ComboData] = []

## 当前帧时间（由外部更新）
var _current_time: float = 0.0


func _init() -> void:
	system_name = &"InputProcess"
	priority = 0
	phase = &"process"


func get_query() -> Array[StringName]:
	return [&"InputBuffer"]


func process(entities: Array, delta: float) -> void:
	_current_time += delta
	
	for entity in entities:
		var buffer: InputBufferComponent = entity.get_component(&"InputBuffer")
		buffer.current_time = _current_time
		
		# 清理过期输入
		buffer.clean_expired()
		
		# 尝试匹配连招
		var matched: StringName = _match_combo(buffer, entity)
		
		# 将结果写入 CombatState
		if entity.has_component(&"CombatState"):
			var combat: CombatStateComponent = entity.get_component(&"CombatState")
			if matched != &"":
				# 触发技能
				if entity.has_component(&"SkillSet"):
					var skill_set: SkillSetComponent = entity.get_component(&"SkillSet")
					var skill: SkillData = DataManager.get_data(&"skills", matched)
					if skill and skill_set.can_use_skill(matched):
						skill_set.use_skill(matched, skill.cooldown)
						combat.add_combo()
						buffer.clear()  # 消费输入
						EventBus.emit_event(&"skill_used", {
							"entity_id": entity.entity_id,
							"skill_id": matched,
						})


## 尝试匹配连招（返回匹配到的技能 ID，&"" 表示无匹配）
func _match_combo(buffer: InputBufferComponent, entity: Node) -> StringName:
	# 按优先级排序的连招表
	var sorted_combos: Array[ComboData] = combo_table.duplicate()
	sorted_combos.sort_custom(func(a, b): return a.priority > b.priority)
	
	for combo in sorted_combos:
		# 检查状态要求
		if combo.required_state != &"" and entity.has_component(&"CombatState"):
			var combat: CombatStateComponent = entity.get_component(&"CombatState")
			if combat.current_state != combo.required_state:
				continue
		
		if _try_match_sequence(buffer, combo):
			return combo.result_skill
	
	return &""


## 尝试匹配单个连招序列
func _try_match_sequence(buffer: InputBufferComponent, combo: ComboData) -> bool:
	if combo.sequence.is_empty():
		return false
	if buffer.buffer.size() < combo.sequence.size():
		return false
	
	# 从缓冲末尾向前匹配
	var buf_idx: int = buffer.buffer.size() - 1
	var seq_idx: int = combo.sequence.size() - 1
	
	while seq_idx >= 0 and buf_idx >= 0:
		var required: ComboData.ComboInput = combo.sequence[seq_idx]
		var frame: InputBufferComponent.InputFrame = buffer.buffer[buf_idx]
		
		if frame.action != required.action:
			buf_idx -= 1
			continue
		
		# 检查时间间隔
		if seq_idx < combo.sequence.size() - 1:
			var next_frame: InputBufferComponent.InputFrame = buffer.buffer[buf_idx + 1]
			if next_frame.timestamp - frame.timestamp > required.max_interval:
				return false
		
		seq_idx -= 1
		buf_idx -= 1
	
	return seq_idx < 0  # 所有序列都匹配了


## 注册连招
func register_combo(combo: ComboData) -> void:
	combo_table.append(combo)


## 注销连招（DLC 卸载时使用）
func unregister_combo(combo_id: StringName) -> void:
	combo_table = combo_table.filter(func(c): return c.id != combo_id)
