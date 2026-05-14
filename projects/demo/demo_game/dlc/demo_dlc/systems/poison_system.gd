class_name PoisonSystem extends EcsSystem
## 毒素伤害系统（由 Demo DLC 提供）
##
## 处理所有中毒状态的 Entity，每秒造成持续伤害。
## 验证 DLC System 可以正常注册到 EcsWorld 并参与调度。


func _init() -> void:
	system_name = &"PoisonDamage"
	priority = 15


func get_query() -> Array[StringName]:
	return [&"Poison", &"RuntimeStats"]


func process(entities: Array, delta: float) -> void:
	for entity in entities:
		var poison: PoisonComponent = entity.get_component(&"Poison")
		var runtime: RuntimeStatsComponent = entity.get_component(&"RuntimeStats")

		if not poison.is_poisoned or not runtime.is_alive:
			continue

		# 递减中毒时间
		poison.remaining_time -= delta
		if poison.remaining_time <= 0.0:
			poison.is_poisoned = false
			print("[DLC PoisonSystem] 中毒状态结束")
			continue

		# 造成毒素伤害（每帧按比例）
		var tick_damage: float = poison.damage_per_second * delta
		var actual: float = runtime.take_damage(tick_damage)
		if actual > 0.0:
			print("[DLC PoisonSystem] 毒素伤害: %.2f，剩余 HP: %.1f" % [actual, runtime.current_hp])

		# 通过 EventBus 广播（验证 DLC System 可以访问 Autoload）
		if EventBus:
			EventBus.emit_event(&"poison_tick", {
				"entity_id": entity.entity_id,
				"damage": tick_damage,
			})
