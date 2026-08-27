extends Node
## §5.5 视觉验证（GUI 渲染，非 headless）：双主题截图 / 关键节点 rect 对比（无跳变）/ 视口溢出检查 / 重启恢复主题
## 用法：--scene res://tools/_shot2.tscn [--theme=neon|elegant] [--restore] [--scale=720x1600] [--shot_out=<绝对路径>]
## --theme：显式切换（模拟用户点击 🎨）；--restore：不显式切换，验证启动从 profile 恢复

var _theme := "neon"
var _restore := false
var _scale := Vector2i.ZERO
var _out := ""
var _fails := 0

func _ready() -> void:
	for a in OS.get_cmdline_args():
		if a.begins_with("--theme="):
			_theme = a.substr(8)
		elif a == "--restore":
			_restore = true
		elif a.begins_with("--scale="):
			var p := a.substr(8).split("x")
			_scale = Vector2i(int(p[0]), int(p[1]))
		elif a.begins_with("--shot_out="):
			_out = a.substr(11)

	# 兜底：任何未捕获错误导致流程中断时，25s 后强制退出，避免窗口挂死
	get_tree().create_timer(25.0).timeout.connect(_fallback_quit)

	var root := get_tree().root
	if _scale != Vector2i.ZERO:
		root.content_scale_size = _scale
		await get_tree().process_frame
	var scene := load("res://shell/main.tscn") as PackedScene
	var main := scene.instantiate()
	add_child(main)
	for _i in 6:
		await get_tree().process_frame
	if not _restore:
		ThemeTokens.apply_theme(_theme, root)
		await get_tree().create_timer(1.0).timeout  # 等 BgLayer 400ms tween 落定（桌面帧率不限速，按帧数等不准）

	print("[vis] theme=", ThemeTokens.current, " css=", root.content_scale_size, " visible=", root.get_visible_rect())
	# 诊断：Theme 链路生效验证（Label 解析色 + Bg 底色，应随主题变化）
	var bgc: Variant = _find(main, "Bg")
	if bgc is ColorRect:
		print("[diag] Bg.color=", (bgc as ColorRect).color)
	var title: Variant = _find(main, "Title")
	if title is Label:
		print("[diag] Title.font_color=", (title as Label).get_theme_color("font_color"))
	# 关键节点 rect（同 scale 下跨主题对比，验证无布局跳变）
	for n in ["App", "TabBar", "Home", "Title", "ThemeButton"]:
		var c: Variant = _find(main, n)
		if c is Control and is_instance_valid(c):
			print("[rect] ", n, " ", (c as Control).get_global_rect())
		elif c != null:
			print("[rect] ", n, " FREED")
		else:
			print("[rect] ", n, " NOT FOUND")
	# 溢出检查：App 下所有可见 Control（ToastLayer 设计上初始屏外，整棵跳过）
	var app: Variant = _find(main, "App")
	if app != null:
		var over := _check_overflow(app as Control, root.get_visible_rect(), "")
		if over.is_empty():
			print("[overflow] none")
		else:
			for o in over:
				print("[overflow] ", o)
			_fails = over.size()
	if _restore:
		var expect := str(DB.get_profile().get("theme", "neon"))
		if ThemeTokens.current == expect:
			print("[restore] PASS current=", ThemeTokens.current, "== profile=", expect)
		else:
			print("[restore] FAIL current=", ThemeTokens.current, " profile=", expect)
			_fails += 1
	if _out != "":
		var tex := root.get_texture()
		var img: Image = null
		if tex != null:
			img = tex.get_image()
		if img != null and img.get_size().x > 1:
			print("[shot] saved=", img.save_png(_out) == OK, " ->", _out)
		else:
			print("[shot] FAIL empty render")
			_fails += 1
	get_tree().quit(0 if _fails == 0 else 1)

func _fallback_quit() -> void:
	print("[vis] TIMEOUT 25s fallback quit")
	get_tree().quit(1)

# Window 无 find_node，递归按名查找
func _find(node: Node, name: String) -> Variant:
	if node.name == name:
		return node
	for ch in node.get_children():
		var r: Variant = _find(ch, name)
		if r != null:
			return r
	return null

# 递归检查可见 Control 是否超出视口（1px 容差）；ToastLayer 子树跳过
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
