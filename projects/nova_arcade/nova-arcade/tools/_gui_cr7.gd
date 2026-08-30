extends Node
## CR-7 T8 GUI 双主题验收（简化版）：library 成就墙 + home 推荐区块 + ToastLayer
## 用法：Godot_v4.7.2-stable_win64.exe --path . res://tools/_gui_cr7.tscn --theme=neon|elegant --shot_out=<abs_prefix>

var _theme := "neon"
var _out := ""
var _fails := 0

func _ready() -> void:
	for a in OS.get_cmdline_args():
		if a.begins_with("--theme="):
			_theme = a.substr(8)
		elif a.begins_with("--shot_out="):
			_out = a.substr(11)
	
	get_tree().create_timer(25.0).timeout.connect(_fallback_quit)

	var root := get_tree().root
	var scene := load("res://shell/main.tscn") as PackedScene
	var main_node := scene.instantiate()
	root.add_child.call_deferred(main_node)
	await get_tree().process_frame
	await get_tree().create_timer(1.5).timeout
	
	ThemeTokens.apply_theme(_theme, root)
	await get_tree().create_timer(1.0).timeout
	print("[guicr7] theme=", ThemeTokens.current)

	var app: Variant = _find(root, "App")
	if app == null:
		print("[guicr7] FAIL App 未找到")
		get_tree().quit(1)
		return

	# === 1. Library 成就墙页面 (tab_idx=3) ===
	print("[guicr7] === library page ===")
	var nav: Variant = _find(root, "Nav")
	if nav is Node:
		(nav as Node).call("switch_tab", 3)
		await get_tree().create_timer(1.5).timeout
	
	var lib: Variant = _find(root, "Library")
	_check(lib != null, "Library 页面存在")
	
	var ach_wall: Variant = _find_by_script(root, "section_achievement_wall.gd")
	_check(ach_wall != null, "SectionAchievementWall 节点存在")
	
	_shot(root, _out + "_library.png")

	# === 2. Home 推荐区块 (tab_idx=0) ===
	print("[guicr7] === home page ===")
	if nav is Node:
		(nav as Node).call("switch_tab", 0)
		await get_tree().create_timer(1.5).timeout
	
	var home: Variant = _find(root, "Home")
	_check(home != null, "Home 页面存在")
	
	_shot(root, _out + "_home.png")

	# === 3. ToastLayer 金色 Toast ===
	print("[guicr7] === toast golden ===")
	var toast: Variant = _find_by_script(root, "toast_layer.gd")
	if toast is Node:
		toast.call("show_msg", "金色测试 Toast", Color(1, 0.824, 0.247))
		await get_tree().create_timer(0.5).timeout
		_shot(root, _out + "_toast.png")

	print("[guicr7] 合计 FAIL=%d" % _fails)
	get_tree().quit(0 if _fails == 0 else 1)

func _fallback_quit() -> void:
	print("[guicr7] TIMEOUT 25s fallback quit")
	get_tree().quit(1)

func _check(cond: bool, label: String) -> void:
	if cond:
		print("[guicr7] PASS", label)
	else:
		_fails += 1
		push_error("[guicr7] FAIL", label)

func _shot(root: Window, path: String) -> void:
	await get_tree().create_timer(0.5).timeout
	var tex := root.get_texture()
	if tex != null:
		var img: Image = tex.get_image()
		if img != null and img.get_size().x > 1:
			print("[shot] saved=", img.save_png(path) == OK, " ->", path)
		else:
			print("[shot] FAIL empty render", path)
			_fails += 1
	else:
		print("[shot] FAIL no texture", path)
		_fails += 1

func _find(node: Node, name: String) -> Variant:
	if node.name == name:
		return node
	for ch in node.get_children():
		var r: Variant = _find(ch, name)
		if r != null:
			return r
	return null

func _find_by_script(node: Node, script_name: String) -> Variant:
	if node is CanvasItem and (node as CanvasItem).script != null \
		and str((node as CanvasItem).script.resource_path).ends_with(script_name):
		return node
	for ch in node.get_children():
		var r: Variant = _find_by_script(ch, script_name)
		if r != null:
			return r
	return null
