extends Node
## CR-2 T9 视觉验证（GUI 渲染）：双主题 × 三画面（首页/游戏运行中/结算卡）截图 + 溢出检查
## 用法：--scene res://tools/_shot3.tscn --theme=neon|elegant --shot_out=<绝对路径前缀>
## 产出：<前缀>_home.png / <前缀>_run.png / <前缀>_result.png

var _theme := "neon"
var _out := ""
var _fails := 0

func _ready() -> void:
	for a in OS.get_cmdline_args():
		if a.begins_with("--theme="):
			_theme = a.substr(8)
		elif a.begins_with("--shot_out="):
			_out = a.substr(11)
	get_tree().create_timer(30.0).timeout.connect(_fallback_quit)

	var root := get_tree().root
	var scene := load("res://shell/main.tscn") as PackedScene
	add_child(scene.instantiate())
	for _i in 6:
		await get_tree().process_frame
	ThemeTokens.apply_theme(_theme, root)
	await get_tree().create_timer(1.0).timeout

	print("[vis3] theme=", ThemeTokens.current, " css=", root.content_scale_size)
	# --- 画面 1：首页（含临时启动按钮）---
	_shot(root, _out + "_home.png")
	var app: Variant = _find(get_tree().root, "App")
	if app != null:
		var over := _check_overflow(app as Control, root.get_visible_rect(), "")
		print("[overflow-home] ", "none" if over.is_empty() else str(over))
		_fails += over.size()
	# --- 画面 2：游戏运行中 ---
	Launcher.launch("tetra_nova")
	await get_tree().create_timer(2.5).timeout
	print("[vis3] state=", Launcher.state)
	if Launcher.state == Launcher.State.RUNNING:
		_shot(root, _out + "_run.png")
	else:
		_fails += 1
	# --- 画面 3：结算卡 ---
	var m := Launcher.get_module()
	if m != null:
		m.quit_to_shell()
	await get_tree().create_timer(2.5).timeout
	var card: Variant = _find(get_tree().root, "ResultOverlay")
	if card is Control and (card as Control).visible:
		print("[vis3] 结算卡可见，溢出检查：")
		var over2 := _check_overflow(card as Control, root.get_visible_rect(), "")
		print("[overflow-result] ", "none" if over2.is_empty() else str(over2))
		_fails += over2.size()
		_shot(root, _out + "_result.png")
	else:
		print("[vis3] FAIL 结算卡未显示")
		_fails += 1
	get_tree().quit(0 if _fails == 0 else 1)

func _fallback_quit() -> void:
	print("[vis3] TIMEOUT 30s fallback quit")
	get_tree().quit(1)

func _shot(root: Window, path: String) -> void:
	var img: Image = null
	var tex := root.get_texture()
	if tex != null:
		img = tex.get_image()
	if img != null and img.get_size().x > 1:
		print("[shot] saved=", img.save_png(path) == OK, " ->", path)
	else:
		print("[shot] FAIL empty render ", path)
		_fails += 1

func _find(node: Node, name: String) -> Variant:
	if node.name == name:
		return node
	for ch in node.get_children():
		var r: Variant = _find(ch, name)
		if r != null:
			return r
	return null

func _check_overflow(c: Control, vp: Rect2, path: String) -> Array:
	var out := Array()
	if c.name == "ToastLayer":
		return out
	var p := path + "/" + c.name
	if c.is_visible():
		var r: Rect2 = c.get_global_rect()
		if r.position.x < vp.position.x - 1.0 or r.position.y < vp.position.y - 1.0 \
			or r.end.x > vp.end.x + 1.0 or r.end.y > vp.end.y + 1.0:
			out.append(p + " " + str(r))
	for ch in c.get_children():
		if ch is Control:
			out.append_array(_check_overflow(ch as Control, vp, p))
	return out
