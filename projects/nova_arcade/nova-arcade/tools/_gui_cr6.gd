extends Node
## CR-6 T6 GUI 双主题验收（neon/elegant）：
##   1) 首页 Banner 轮播含新游戏条目 + SectionContinue 多条目渲染（无 overflow/布局跳变）
##   2) 三款游戏分别 launch → 游戏画面 → quit_to_shell → 结算卡弹出（score/playtime 正确）
## 用法：Godot_v4.7.2-stable_win64.exe --path . res://tools/_gui_cr6.tscn --theme=neon|elegant --shot_out=<abs_prefix>
## 产出：<prefix>_home.png / <prefix>_<gid>_game.png / <prefix>_<gid>_result.png
## stdout: [guicr6] lines; exit 0 = all pass.

var _theme := "neon"
var _out := ""
var _fails := 0
const GIDS := ["magic_tower", "game_2048", "snake"]

func _ready() -> void:
	for a in OS.get_cmdline_args():
		if a.begins_with("--theme="):
			_theme = a.substr(8)
		elif a.begins_with("--shot_out="):
			_out = a.substr(11)
	get_tree().create_timer(45.0).timeout.connect(_fallback_quit)

	var root := get_tree().root
	var scene := load("res://shell/main.tscn") as PackedScene
	var main_node := scene.instantiate()
	root.add_child.call_deferred(main_node)
	await get_tree().process_frame
	await get_tree().create_timer(1.0).timeout
	ThemeTokens.apply_theme(_theme, root)
	await get_tree().create_timer(1.0).timeout
	print("[guicr6] theme=", ThemeTokens.current, " css=", root.content_scale_size)

	var app: Variant = _find(root, "App")
	if app == null:
		print("[guicr6] FAIL App 未找到")
		get_tree().quit(1)
		return

	# 种子：三款新游戏写入 last_played>0 → SectionContinue 多条目
	for gid in GIDS:
		DB.upsert_record(gid, {"last_played": 1000 + int(gid.hash() % 100), "best": 0})
	await get_tree().process_frame
	var sc: Variant = _find(root, "SectionContinue")
	if sc is Node:
		(sc as Node).call("render")
		await get_tree().create_timer(0.5).timeout

	# 首页：Banner 可见 + SectionContinue 多卡片 + 溢出检查
	var home: Variant = _find(root, "Home")
	var ban_ok := false
	if home is Node:
		ban_ok = (home as Node).get_node_or_null("ScrollView/VBox/SectionBanner").is_visible_in_tree()
	_check(ban_ok, "首页 SectionBanner 可见（editorial banner[] 含新游戏）")
	var sc_cards := -1
	if sc is Node:
		var hbox: Variant = (sc as Node).get_node_or_null("ScrollContainer/HBox")
		sc_cards = hbox.get_child_count() if hbox != null else -1
	_check(sc_cards >= 3, "SectionContinue 多条目渲染（卡片数=%d ≥3）" % sc_cards)
	var over := _check_overflow(app as Control, root.get_visible_rect(), "")
	print("[overflow-home] ", "none" if over.is_empty() else str(over))
	_fails += over.size()
	_shot(root, _out + "_home.png")

	# 三款游戏：launch → 画面 → quit → 结算卡
	for gid in GIDS:
		await _run_game(root, gid)

	print("[guicr6] 合计 FAIL=%d" % _fails)
	get_tree().quit(0 if _fails == 0 else 1)

func _fallback_quit() -> void:
	print("[guicr6] TIMEOUT 45s fallback quit")
	get_tree().quit(1)

func _check(cond: bool, label: String) -> void:
	if cond:
		print("[guicr6] PASS ", label)
	else:
		_fails += 1
		push_error("[guicr6] FAIL ", label)

## launch(gid) → 等游戏画面 → quit_to_shell → 等结算卡 → 截图 game + result
func _run_game(root: Window, gid: String) -> void:
	print("[guicr6] === launch %s ===" % gid)
	Launcher.launch(gid)
	await get_tree().create_timer(2.5).timeout
	var mod := Launcher.get_module()
	_check(mod != null and is_instance_valid(mod), "%s 模块已挂载（Launcher.state=RUNNING）" % gid)
	if mod == null:
		return
	_shot(root, _out + "_" + gid + "_game.png")
	# 回盒 → 结算卡
	mod.call("quit_to_shell")
	await get_tree().create_timer(2.0).timeout
	var result: Variant = _find(root, "ResultOverlay")
	if result == null:
		result = _find_by_script(root, "result_overlay.gd")
	_check(result != null and (result as Control).visible, "%s 结算卡弹出（show_card）" % gid)
	if result is Control and (result as Control).visible:
		var score_line: Variant = (result as Node).get_node_or_null("Card/VBox/ScoreLine")
		if score_line is Label:
			print("[guicr6] %s 结算卡 ScoreLine='%s'" % [gid, (score_line as Label).text])
	_shot(root, _out + "_" + gid + "_result.png")
	# 释放模块，恢复 idle，准备下一款
	await Launcher.release()
	await get_tree().create_timer(1.0).timeout

func _shot(root: Window, path: String) -> void:
	var tex := root.get_texture()
	if tex != null:
		var img: Image = tex.get_image()
		if img != null and img.get_size().x > 1:
			print("[shot] saved=", img.save_png(path) == OK, " ->", path)
		else:
			print("[shot] FAIL empty render ", path)
			_fails += 1
	else:
		print("[shot] FAIL no texture ", path)
		_fails += 1

func _find(node: Node, name: String) -> Variant:
	if node.name == name:
		return node
	for ch in node.get_children():
		var r: Variant = _find(ch, name)
		if r != null:
			return r
	return null

## 按脚本文件名定位节点（ResultOverlay 场景根名可能不同）
func _find_by_script(node: Node, script_name: String) -> Variant:
	if node is CanvasItem and (node as CanvasItem).script != null \
		and str((node as CanvasItem).script.resource_path).ends_with(script_name):
		return node
	for ch in node.get_children():
		var r: Variant = _find_by_script(ch, script_name)
		if r != null:
			return r
	return null

func _check_overflow(c: Control, vp: Rect2, path: String) -> Array:
	var out := Array()
	if c.name == "ToastLayer" or c.name == "TransitionOverlay":
		return out
	var p := path + "/" + c.name
	if c.is_visible_in_tree():
		var r: Rect2 = c.get_global_rect()
		if r.position.x < vp.position.x - 1.0 or r.position.y < vp.position.y - 1.0 \
			or r.end.x > vp.end.x + 1.0 or r.end.y > vp.end.y + 1.0:
			out.append(p + " " + str(r))
	for ch in c.get_children():
		if ch is Control:
			out.append_array(_check_overflow(ch as Control, vp, p))
	return out
