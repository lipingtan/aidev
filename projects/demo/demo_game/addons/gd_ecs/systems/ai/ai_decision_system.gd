class_name AiDecisionSystem extends EcsSystem
## AI 决策系统
##
## 每帧更新 AI 状态机的决策，处理仇恨衰减。
## 决策结果写入 AiStateComponent，由状态机读取执行。
##
## 依赖 Component：
## - AiState: AI 状态和参数
## - AggroTable（可选）: 仇恨管理
## - RuntimeStats（可选）: 血量检查（撤退判断）
## - FinalStats（可选）: 最大血量


func _init() -> void:
	system_name = &"AiDecision"
	priority = 20
	phase = &"process"


func get_query() -> Array[StringName]:
	return [&"AiState"]


func process(entities: Array, delta: float) -> void:
	for entity in entities:
		var ai: AiStateComponent = entity.get_component(&"AiState")
		
		# 死亡状态不处理
		if ai.ai_mode == AiStateComponent.AiMode.DEAD:
			continue
		
		# 仇恨衰减
		if entity.has_component(&"AggroTable"):
			_decay_aggro(entity.get_component(&"AggroTable"), delta)
		
		# 根据当前模式更新决策
		match ai.ai_mode:
			AiStateComponent.AiMode.IDLE:
				_update_idle(entity, ai, delta)
			AiStateComponent.AiMode.PATROL:
				_update_patrol(entity, ai, delta)
			AiStateComponent.AiMode.ALERT:
				_update_alert(entity, ai, delta)
			AiStateComponent.AiMode.COMBAT:
				_update_combat(entity, ai, delta)
			AiStateComponent.AiMode.RETREAT:
				_update_retreat(entity, ai, delta)


## 待机状态：检测是否有目标进入警戒范围
func _update_idle(entity: Node, ai: AiStateComponent, _delta: float) -> void:
	var target_id: int = _find_nearest_target(entity, ai.alert_range)
	if target_id >= 0:
		ai.target_entity_id = target_id
		ai.ai_mode = AiStateComponent.AiMode.ALERT
		ai.alert_timer = ai.alert_timeout


## 巡逻状态：同待机，但有移动路径
func _update_patrol(entity: Node, ai: AiStateComponent, delta: float) -> void:
	_update_idle(entity, ai, delta)


## 警戒状态：确认目标后进入战斗，超时回到待机
func _update_alert(entity: Node, ai: AiStateComponent, delta: float) -> void:
	ai.alert_timer -= delta
	
	# 检查目标是否在攻击范围内
	var target: Node = _get_entity_by_id(ai.target_entity_id)
	if target:
		var dist: float = _get_distance(entity, target)
		if dist <= ai.attack_range:
			ai.ai_mode = AiStateComponent.AiMode.COMBAT
			return
	
	# 超时回到待机
	if ai.alert_timer <= 0.0:
		ai.target_entity_id = -1
		ai.ai_mode = AiStateComponent.AiMode.IDLE


## 战斗状态：检查撤退条件，更新仇恨目标
func _update_combat(entity: Node, ai: AiStateComponent, _delta: float) -> void:
	# 检查撤退条件
	if _should_retreat(entity, ai):
		ai.ai_mode = AiStateComponent.AiMode.RETREAT
		return
	
	# 从仇恨表更新目标
	if entity.has_component(&"AggroTable"):
		var aggro: AggroTableComponent = entity.get_component(&"AggroTable")
		var top_target: int = aggro.get_top_target()
		if top_target >= 0:
			ai.target_entity_id = top_target
	
	# 目标丢失检查
	var target: Node = _get_entity_by_id(ai.target_entity_id)
	if not target:
		ai.target_entity_id = -1
		ai.ai_mode = AiStateComponent.AiMode.ALERT
		ai.alert_timer = ai.alert_timeout
		return
	
	# 超出追击范围
	var dist: float = _get_distance(entity, target)
	if dist > ai.chase_range:
		ai.ai_mode = AiStateComponent.AiMode.ALERT
		ai.alert_timer = ai.alert_timeout


## 撤退状态：远离目标
func _update_retreat(entity: Node, ai: AiStateComponent, _delta: float) -> void:
	# 如果血量恢复则回到战斗
	if not _should_retreat(entity, ai):
		ai.ai_mode = AiStateComponent.AiMode.COMBAT


## 仇恨衰减
func _decay_aggro(aggro: AggroTableComponent, delta: float) -> void:
	var to_remove: Array[int] = []
	for id in aggro.entries:
		aggro.entries[id] -= aggro.decay_rate * delta
		if aggro.entries[id] <= 0.0:
			to_remove.append(id)
	for id in to_remove:
		aggro.entries.erase(id)


## 检查是否应该撤退
func _should_retreat(entity: Node, ai: AiStateComponent) -> bool:
	if not entity.has_component(&"RuntimeStats") or not entity.has_component(&"FinalStats"):
		return false
	var runtime: RuntimeStatsComponent = entity.get_component(&"RuntimeStats")
	var final_stats: FinalStatsComponent = entity.get_component(&"FinalStats")
	if final_stats.max_hp <= 0.0:
		return false
	return (runtime.current_hp / final_stats.max_hp) < ai.retreat_hp_threshold


## 查找最近的目标（简化实现，实际应通过物理查询）
func _find_nearest_target(_entity: Node, _range: float) -> int:
	# 子类或具体游戏逻辑中实现
	# 这里返回 -1 表示无目标
	return -1


## 通过 ID 获取 Entity
func _get_entity_by_id(entity_id: int) -> Node:
	if entity_id < 0 or not EcsWorld:
		return null
	for entity in EcsWorld.get_all_entities():
		if entity.entity_id == entity_id:
			return entity
	return null


## 获取两个节点间的距离
func _get_distance(a: Node, b: Node) -> float:
	if a is Node3D and b is Node3D:
		return (a as Node3D).global_position.distance_to((b as Node3D).global_position)
	if a is Node2D and b is Node2D:
		return (a as Node2D).global_position.distance_to((b as Node2D).global_position)
	return INF
