extends Node
## CR-8 T8 GUI 验证（环道版）：真实 Shell → launch slot_machine → 五态截图
## 用法：--scene res://tools/_gui_cr8.tscn [--theme=neon|elegant]

var _theme := "neon"
var _fails := 0

func _ready() -> void:
	for a in OS.get_cmdline_args():
		if a.begins_with("--theme="):
			_theme = a.substr(8)
	get_tree().create_timer(150.0).timeout.connect(_fallback)
	var scene := load("res://shell/main.tscn") as PackedScene
	var main := scene.instantiate()
	add_child(main)
	for _i in 10:
		await get_tree().process_frame
	await get_tree().create_timer(1.5).timeout
	ThemeTokens.apply_theme(_theme, get_tree().root)
	await get_tree().create_timer(1.0).timeout

	# 1. launch
	Launcher.launch("slot_machine")
	await get_tree().create_timer(2.5).timeout
	var src := _find_src()
	if src == null:
		print("[cr8] FAIL Src not found")
		_fails += 1
		get_tree().quit(1)
		return
	await _shot("cr8_%s_1_overlay.png" % _theme)
	_ok(src._overlay.visible, "遮罩可见")

	# 2. 点开始 → 押注态
	_press_start(src)
	await get_tree().process_frame
	_ok(not src._overlay.visible, "点开始后遮罩隐藏")
	src.add_bet(0)
	src.add_bet(2)
	src.add_bet(5)
	await get_tree().process_frame
	await _shot("cr8_%s_2_bet.png" % _theme)
	_ok(src.total_bet() == 3, "押注 3 份登记")

	# 3. 真实跑灯：快跑段截图 → 等结束
	src._spin_btn.pressed.emit()
	_ok(src.state == src.State.RUNNING, "进入 RUNNING")
	_ok(src._spin_btn.disabled, "跑灯期旋转按钮禁用")
	await get_tree().create_timer(1.2).timeout
	await _shot("cr8_%s_3_spinning.png" % _theme)
	var waited := 0.0
	while src.state != src.State.IDLE_BET and waited < 25.0:
		await get_tree().create_timer(0.25).timeout
		waited += 0.25
	_ok(src.state == src.State.IDLE_BET, "跑灯完成回 IDLE_BET (waited=%.1fs)" % waited)

	# 4. 开奖结果态：构造结算（押苹果停苹果格 1）
	src._bets = [1, 0, 0, 0, 0, 0, 0, 0]
	src.balance = max(src.balance, 10)
	src._settle_from_stops([1], 1)
	await get_tree().process_frame
	await _shot("cr8_%s_4_result.png" % _theme)
	_ok(true, "结果态截图完成")

	# 5. 破产救济态
	src.balance = 0
	src._refresh_all()
	await get_tree().process_frame
	_ok(src._relief_btn.visible, "救济按钮可见")
	await _shot("cr8_%s_5_relief.png" % _theme)

	print("[cr8] RESULT fails=", _fails)
	get_tree().quit(1 if _fails > 0 else 0)

func _find_src() -> Node2D:
	var host := get_tree().root.find_child("GameHost", true, false)
	if host == null:
		return null
	for c in host.get_children():
		var s := c.find_child("Src", true, false)
		if s != null and s.get_script() != null and str(s.get_script().resource_path).contains("slot_main"):
			return s as Node2D
	return null

func _press_start(src: Node2D) -> void:
	var overlay := src._overlay as Control
	for child in overlay.get_children():
		if child is VBoxContainer:
			for sub in (child as VBoxContainer).get_children():
				if sub is Button:
					(sub as Button).pressed.emit()
					return

func _ok(cond: bool, label: String) -> void:
	if cond:
		print("[cr8] PASS ", label)
	else:
		_fails += 1
		print("[cr8] FAIL ", label)

func _shot(fname: String) -> void:
	await get_tree().process_frame
	await get_tree().process_frame
	var tex := get_tree().root.get_texture()
	if tex == null:
		_fails += 1
		return
	var img := tex.get_image()
	if img == null or img.get_size().x < 2:
		_fails += 1
		return
	var err := img.save_png(ProjectSettings.globalize_path("res://dev/" + fname))
	print("[cr8] shot ", fname, " err=", err)

func _fallback() -> void:
	print("[cr8] timeout")
	get_tree().quit(2)
