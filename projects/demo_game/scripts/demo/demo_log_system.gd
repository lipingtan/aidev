class_name DemoLogSystem extends EcsSystem
## Demo 日志系统
##
## 每秒输出一次所有 Entity 的状态，用于验证 ECS 框架运行正常。

var _log_timer: float = 0.0
const LOG_INTERVAL: float = 2.0  # 每 2 秒输出一次


func _init() -> void:
	system_name = &"DemoLog"
	priority = 99  # 最后执行，确保读到最新数据


func get_query() -> Array[StringName]:
	return [&"DemoTag", &"RuntimeStats", &"FinalStats"]


func process(entities: Array, delta: float) -> void:
	_log_timer += delta
	if _log_timer < LOG_INTERVAL:
		return
	_log_timer = 0.0

	print("\n=== ECS 状态快照 ===")
	for entity in entities:
		var tag: DemoTagComponent = entity.get_component(&"DemoTag")
		var runtime: RuntimeStatsComponent = entity.get_component(&"RuntimeStats")
		var final_stats: FinalStatsComponent = entity.get_component(&"FinalStats")
		print("  [%s] HP: %.0f/%.0f  ATK: %.1f  DEF: %.1f  存活: %s" % [
			tag.display_name,
			runtime.current_hp,
			final_stats.max_hp,
			final_stats.attack,
			final_stats.defense,
			"是" if runtime.is_alive else "否",
		])
	print("===================")
