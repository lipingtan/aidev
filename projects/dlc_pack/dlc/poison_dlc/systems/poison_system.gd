class_name PoisonSystem extends EcsSystem
## 毒素伤害系统（DLC 提供）

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

		poison.remaining_time -= delta
		if poison.remaining_time <= 0.0:
			poison.is_poisoned = false
			print("[DLC PoisonSystem] 中毒结束")
			continue

		var tick_damage: float = poison.damage_per_second * delta
		runtime.take_damage(tick_damage)

		if EventBus:
			EventBus.emit_event(&"poison_tick", {
				"entity_id": entity.entity_id,
				"damage": tick_damage,
				"remaining": poison.remaining_time,
			})
