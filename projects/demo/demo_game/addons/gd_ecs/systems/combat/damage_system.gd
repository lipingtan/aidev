class_name DamageSystem extends EcsSystem
## 伤害计算系统
##
## 处理所有挂载了 DamageEvent 的 Entity，执行完整伤害管线：
## 命中判定 → 伤害计算 → 应用伤害 → 后处理
##
## 伤害管线：
## 基础伤害 → 防御减伤 → 元素克制 → 暴击 → 格挡 → 最终伤害
##
## 依赖 Component：
## - FinalStats: 攻防数值
## - CombatState: 无敌帧/格挡状态
## - RuntimeStats: 扣减 HP
## - DamageEvent: 伤害事件（处理后移除）

## 元素克制表（可通过 DataManager 加载自定义表）
var element_table: ElementTable = ElementTable.new()

## 防御公式常数（防御/(防御+常数) = 减伤比例）
const DEFENSE_CONSTANT: float = 100.0


func _init() -> void:
	system_name = &"Damage"
	priority = 10
	phase = &"physics_process"


func get_query() -> Array[StringName]:
	return [&"DamageEvent", &"FinalStats", &"RuntimeStats"]


func process(entities: Array, delta: float) -> void:
	for entity in entities:
		var damage_event: DamageEventComponent = entity.get_component(&"DamageEvent")
		var final_stats: FinalStatsComponent = entity.get_component(&"FinalStats")
		var runtime_stats: RuntimeStatsComponent = entity.get_component(&"RuntimeStats")
		
		# 无敌帧检查
		if entity.has_component(&"CombatState"):
			var combat: CombatStateComponent = entity.get_component(&"CombatState")
			if combat.is_invincible:
				entity.remove_component(&"DamageEvent")
				continue
		
		# 计算最终伤害
		var final_damage: float = _calculate_damage(damage_event, final_stats, entity)
		
		# 应用伤害
		var actual_damage: float = runtime_stats.take_damage(final_damage)
		
		# 发送伤害事件
		if EventBus:
			EventBus.emit_event(&"damage_dealt", {
				"target_id": entity.entity_id,
				"attacker_id": damage_event.attacker_id,
				"damage": actual_damage,
				"is_crit": damage_event.is_crit,
				"element": damage_event.element,
				"hit_position": damage_event.hit_position,
			})
		
		# 消费事件
		entity.remove_component(&"DamageEvent")


## 计算最终伤害
func _calculate_damage(event: DamageEventComponent, stats: FinalStatsComponent, entity: Node) -> float:
	var damage: float = event.damage_amount
	
	# 防御减伤
	match event.damage_type:
		DamageEventComponent.DamageType.PHYSICAL:
			var defense: float = stats.defense
			damage *= (1.0 - defense / (defense + DEFENSE_CONSTANT))
		DamageEventComponent.DamageType.MAGICAL:
			var m_defense: float = stats.magic_defense
			damage *= (1.0 - m_defense / (m_defense + DEFENSE_CONSTANT))
		DamageEventComponent.DamageType.TRUE:
			pass  # 真实伤害无视防御
	
	# 元素克制
	var element_resist: float = _get_element_resist(event.element, stats)
	var element_mult: float = element_table.get_multiplier(event.element, 0)
	damage *= element_mult * (1.0 - element_resist)
	
	# 暴击
	if event.is_crit:
		damage *= stats.crit_damage
	
	# 格挡减伤
	if entity.has_component(&"CombatState"):
		var combat: CombatStateComponent = entity.get_component(&"CombatState")
		if combat.is_blocking:
			damage *= (1.0 - combat.block_reduction)
	
	return maxf(1.0, damage)  # 最低造成 1 点伤害


## 获取元素抗性
func _get_element_resist(element: int, stats: FinalStatsComponent) -> float:
	match element:
		DamageEventComponent.ElementType.FIRE: return stats.fire_resist
		DamageEventComponent.ElementType.ICE: return stats.ice_resist
		DamageEventComponent.ElementType.LIGHTNING: return stats.lightning_resist
		DamageEventComponent.ElementType.SHADOW: return stats.shadow_resist
	return 0.0
