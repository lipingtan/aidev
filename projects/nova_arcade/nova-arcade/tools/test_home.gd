extends Node
## CR-5 T9 headless 回归：design §11 RG-21~30 + FR 验收口径 + touch_daily 共存规则。
## 用法（隔离 APPDATA）：Godot --headless --path <proj> res://tools/test_home.tscn

var _failures: int = 0
var _pass: int = 0
## RG-24 信号捕获（lambda 按值捕获局部变量，须用成员变量）
var _signal_fired: bool = false

func _ready() -> void:
	_run()

func _check(cond: bool, label: String) -> void:
	if cond:
		_pass += 1
		print("[test_home] PASS ", label)
	else:
		_failures += 1
		push_error("[test_home] FAIL ", label)

## 重置每日数据（内存 + 磁盘），避免跨 case 污染
func _reset_daily() -> void:
	DB._daily = {"date": Time.get_date_string_from_system(), "plays": 0}
	_io_flush_daily()

func _io_flush_daily() -> void:
	DB._io.flush_all()

## RG-21: banner[] 为空 → SectionBanner.visible=false
func _test_rg21() -> void:
	Registry.reload()
	Registry._editorial["banner"] = []
	var ban := load("res://shell/components/section_banner.tscn").instantiate() as SectionBanner
	add_child(ban)
	await get_tree().process_frame
	_check(ban.visible == false, "RG-21 banner[] 空 → visible=false")
	ban.queue_free()

## RG-22: banner[] 含 1 条 → visible=true，Timer 不启动（Q4）
func _test_rg22() -> void:
	Registry._editorial["banner"] = [{"image_path": "", "target_gid": "tetra_nova", "title": "TETRA NOVA"}]
	var ban := load("res://shell/components/section_banner.tscn").instantiate() as SectionBanner
	add_child(ban)
	await get_tree().process_frame
	_check(ban.visible == true, "RG-22 banner[1条] → visible=true")
	_check(ban._timer == null, "RG-22 单张不启动 Timer（Q4）")
	ban.queue_free()

## RG-23: charts("all") 非空，rank 从 1 开始，players≥10000
func _test_rg23() -> void:
	var rec := Recommender.new()
	add_child(rec)
	var rows := rec.charts("all")
	_check(not rows.is_empty(), "RG-23 charts(all) 非空")
	if not rows.is_empty():
		_check(int(rows[0].get("rank", 0)) == 1, "RG-23 rank 从 1 开始")
		_check(int(rows[0].get("players", 0)) >= 10000, "RG-23 players≥10000（Q6 base）")
	rec.queue_free()

## RG-24: on_game_launched → launch done=true + tasks_updated 信号
func _test_rg24() -> void:
	_reset_daily()
	var svc := DailyTaskService.new()
	add_child(svc)
	await get_tree().process_frame
	_signal_fired = false
	svc.tasks_updated.connect(func(_tasks): _signal_fired = true)
	EventBus.game_launched.emit("tetra_nova")
	var tasks: Array[Dictionary] = svc.get_tasks()
	_check(tasks.size() == 3, "RG-24 get_tasks 返回 3 条")
	_check(bool(tasks[0].get("done", false)), "RG-24 launch 任务 done=true")
	_check(_signal_fired, "RG-24 tasks_updated 信号发出")
	svc.queue_free()

## RG-25/RG-26: 三任务全完成 → Toast + reward_claimed；重复触发不再弹
func _test_rg25_26() -> void:
	_reset_daily()
	var svc := DailyTaskService.new()
	add_child(svc)
	await get_tree().process_frame
	EventBus.game_launched.emit("tetra_nova")
	EventBus.game_finished.emit("tetra_nova", {"playtime": 310.0})
	var tasks: Array[Dictionary] = svc.get_tasks()
	var all_done := true
	for t in tasks:
		if not bool(t.get("done", false)):
			all_done = false
	_check(all_done, "RG-25 三任务全 done")
	_check(svc.toast_shown_count == 1, "RG-25 Toast 弹出一次")
	_check(bool(DB.get_daily_tasks().get("reward_claimed", false)), "RG-25 reward_claimed=true 写入 DB")
	EventBus.game_finished.emit("tetra_nova", {"playtime": 60.0})
	_check(svc.toast_shown_count == 1, "RG-26 重复触发不再弹 Toast")
	svc.queue_free()

## RG-27: 跨自然日 → 进度全重置 + reward_claimed=false
func _test_rg27() -> void:
	_reset_daily()
	var svc := DailyTaskService.new()
	add_child(svc)
	await get_tree().process_frame
	EventBus.game_launched.emit("tetra_nova")
	EventBus.game_finished.emit("tetra_nova", {"playtime": 310.0})
	_check(bool(DB.get_daily_tasks().get("reward_claimed", false)), "RG-27 前置：reward_claimed=true")
	## 模拟跨天：date 改为昨天，保留 tasks_data（touch_daily 共存规则）
	DB._daily["date"] = "2026-01-01"
	var svc2 := DailyTaskService.new()
	add_child(svc2)
	await get_tree().process_frame
	var tasks: Array[Dictionary] = svc2.get_tasks()
	var any_done := false
	for t in tasks:
		if bool(t.get("done", false)):
			any_done = true
	_check(not any_done, "RG-27 跨天后 get_tasks() done 全 false")
	svc.queue_free()
	svc2.queue_free()

## touch_daily 共存规则（design §3）：跨天重置保留 tasks_data
func _test_touch_daily_coexist() -> void:
	_reset_daily()
	DB.update_daily_tasks({"progress": {"launch": 1}, "reward_claimed": true})
	DB._daily["date"] = "2026-01-01"
	DB.touch_daily()
	var daily := DB.get_daily().duplicate(true)
	_check(daily.get("plays") == 0, "touch_daily 跨天 plays=0")
	var kept: Variant = daily.get("tasks_data")
	_check(kept is Dictionary and bool((kept as Dictionary).get("reward_claimed", false)), "touch_daily 跨天保留 tasks_data（design §3）")

## RG-29: banner target_gid="" → 点击不响应（不 push、不崩溃）
func _test_rg29() -> void:
	Registry._editorial["banner"] = [{"image_path": "", "target_gid": "", "title": "NO TARGET"}]
	var ban := load("res://shell/components/section_banner.tscn").instantiate() as SectionBanner
	add_child(ban)
	await get_tree().process_frame
	var before := Nav.current()
	var click := InputEventMouseButton.new()
	click.button_index = MOUSE_BUTTON_LEFT
	click.pressed = true
	ban._gui_input(click)
	_check(Nav.current() == before, "RG-29 target_gid 空点击不 push")
	ban.queue_free()

## RG-30: charts 空数据 → EmptyState 可见、列表区隐藏
func _test_rg30() -> void:
	var charts := load("res://shell/components/section_charts.tscn").instantiate() as SectionCharts
	add_child(charts)
	await get_tree().process_frame
	_check(charts._empty_state.visible == true, "RG-30 无 Recommender → EmptyState 可见")
	_check(charts._list_vbox.visible == false, "RG-30 列表区隐藏")
	charts.queue_free()

## RG-28: 真实 Main 树内 home.on_resume 刷新（不破坏 CR-3/CR-4）+ ThemeToggle 回归
func _test_rg28() -> void:
	var main := load("res://shell/main.tscn").instantiate() as Node
	## 挂到 root 下使绝对路径 /root/Main/Services/... 可解析（home.gd 注入依赖此）
	get_tree().root.add_child(main)
	await get_tree().process_frame
	await get_tree().process_frame
	var home: Variant = main.get_node_or_null("App/PageStack/Home")
	if home == null:
		home = get_tree().root.find_child("Home", true, false)
	_check(home != null, "RG-28 真实 Main 树内 Home 存在")
	if home != null and home.has_method("on_resume"):
		var for_you: Node = home.get_node_or_null("ScrollView/VBox/SectionForYou")
		var hbox: Node = home.get_node_or_null("ScrollView/VBox/SectionForYou/ScrollContainer/HBox")
		home.call("on_enter", {})
		home.call("on_resume")
		await get_tree().process_frame
		_check(home.get("_section_banner") != null, "RG-28 _section_banner 非 null（T7 AC）")
		_check(home.get("_section_charts") != null, "RG-28 _section_charts 非 null（T7 AC）")
		_check(home.get("_daily_svc") != null, "T7 AC: home._daily_svc 注入成功")
		if for_you != null and hbox != null:
			_check(for_you.visible == true and hbox.get_child_count() >= 1, "RG-28 on_resume 刷新为你推荐（CR-4）")
	# ThemeToggle 回归：🎨 按钮切换主题
	var btn := main.get_node_or_null("App/PageStack/Home/ThemeButton")
	if btn != null:
		var before_theme := ThemeTokens.current
		btn.pressed.emit()
		await get_tree().process_frame
		_check(ThemeTokens.current != before_theme, "RG-28 ThemeToggle 不受影响（regression）")
	main.queue_free()

## T8 AC: Main.tscn DailyTaskService 接线
func _test_t7_ac() -> void:
	var main := load("res://shell/main.tscn").instantiate() as Node
	get_tree().root.add_child(main)
	await get_tree().process_frame
	await get_tree().process_frame
	var svc := main.get_node_or_null("Services/DailyTaskService") as DailyTaskService
	_check(svc != null, "T8 AC: Main/Services/DailyTaskService 存在")
	main.queue_free()

func _run() -> void:
	print("=== CR-5 T9 headless 回归（RG-21~30 + FR）===")
	await _test_rg21()
	await _test_rg22()
	await _test_rg23()
	await _test_rg24()
	await _test_rg25_26()
	await _test_rg27()
	await _test_touch_daily_coexist()
	await _test_rg29()
	await _test_rg30()
	await _test_rg28()
	await _test_t7_ac()
	print("[test_home] 合计 PASS=%d FAIL=%d" % [_pass, _failures])
	get_tree().quit(1 if _failures > 0 else 0)
