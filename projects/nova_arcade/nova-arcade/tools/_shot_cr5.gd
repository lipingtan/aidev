extends Node
## CR-5 T9 视觉验证（GUI 渲染）：双主题 × 首页截图 + 溢出检查 + Banner 标题对比度核对（P-3）
## 用法：--scene res://tools/_shot_cr5.tscn --theme=neon|elegant --shot_out=<绝对路径前缀>
## 产出：<prefix>_home.png；stdout: [viscr5] theme / [overflow-home] / [contrast] lines

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
	## 必须挂到 root 下（绝对路径 /root/Main/...），Nav.locate_page_stack 才能找到 PageStack；
	## _ready 期 root 正 setup children，须 call_deferred add_child + 等一帧
	var main_node := scene.instantiate()
	root.add_child.call_deferred(main_node)
	await get_tree().process_frame
	await get_tree().create_timer(1.0).timeout
	ThemeTokens.apply_theme(_theme, root)
	await get_tree().create_timer(1.0).timeout

	print("[viscr5] theme=", ThemeTokens.current, " css=", root.content_scale_size)
	var app: Variant = _find(get_tree().root, "App")
	if app == null:
		print("[viscr5] FAIL App 未找到")
		get_tree().quit(1)
		return
	# 溢出检查（整 App 树）
	var over := _check_overflow(app as Control, root.get_visible_rect(), "")
	print("[overflow-home] ", "none" if over.is_empty() else str(over))
	_fails += over.size()
	# Banner 标题对比度核对（P-3）：采样 BannerRect 左上角背景 vs ink 前景
	var home: Variant = _find(get_tree().root, "Home")
	if home == null or not (home is Node):
		print("[viscr5] FAIL Home 未找到")
		_fails += 1
	else:
		var ban: Variant = (home as Node).get_node_or_null("ScrollView/VBox/SectionBanner")
		if ban == null or not (ban as Node).is_visible_in_tree():
			print("[viscr5] FAIL SectionBanner 不可见（editorial banner[] 应非空）")
			_fails += 1
		else:
			var title: Variant = (ban as Node).get_node_or_null("TitleLabel")
			var rect: Variant = (ban as Node).get_node_or_null("BannerRect")
			if title == null or rect == null or not (title is Control) or not (rect is Control):
				print("[viscr5] FAIL TitleLabel/BannerRect 节点缺失")
				_fails += 1
			else:
				var img: Image = root.get_texture().get_image()
				## 采样 BannerRect 左上角（避开居中标题文字）→ 真实背景色
				var rr: Rect2 = (rect as Control).get_global_rect()
				var bg_c := _sample(img, Vector2(rr.position.x + 16.0, rr.position.y + 16.0))
				var fg := ThemeTokens.color("ink")
				var ratio := _contrast_ratio(fg, bg_c)
				print("[contrast] theme=", _theme, " bg=", bg_c.to_html(), " fg(ink)=", fg.to_html(), " ratio=%.2f" % ratio)
				if ratio < 4.5:
					print("[viscr5] WARN 对比度 %.2f < 4.5（P-3 需人工目检）" % ratio)
	_shot(root, _out + "_home.png")
	get_tree().quit(0 if _fails == 0 else 1)

func _fallback_quit() -> void:
	print("[viscr5] TIMEOUT 30s fallback quit")
	get_tree().quit(1)

## 采样以 center 为中心的 9×9 像素均值
func _sample(img: Image, center: Vector2) -> Color:
	var cx := int(center.x)
	var cy := int(center.y)
	var sum := Vector3.ZERO
	var n := 0
	for dy in range(-4, 5):
		for dx in range(-4, 5):
			var px_x := cx + dx
			var px_y := cy + dy
			if img.get_size().x > px_x and img.get_size().y > px_y and px_x >= 0 and px_y >= 0:
				var c := img.get_pixel(px_x, px_y)
				sum += Vector3(c.r, c.g, c.b)
				n += 1
	if n == 0:
		return Color(0.0, 0.0, 0.0, 1.0)
	return Color(sum.x / float(n), sum.y / float(n), sum.z / float(n), 1.0)

## WCAG 相对亮度对比度（fg vs bg）
func _contrast_ratio(fg: Color, bg: Color) -> float:
	var l1 := _rel_luminance(fg)
	var l2 := _rel_luminance(bg)
	var hi := maxf(l1, l2)
	var lo := minf(l1, l2)
	return (hi + 0.05) / (lo + 0.05)

func _rel_luminance(c: Color) -> float:
	var r := c.r if c.r <= 0.04045 else pow((c.r + 0.055) / 1.055, 2.4)
	var g := c.g if c.g <= 0.04045 else pow((c.g + 0.055) / 1.055, 2.4)
	var b := c.b if c.b <= 0.04045 else pow((c.b + 0.055) / 1.055, 2.4)
	return 0.2126 * r + 0.7152 * g + 0.0722 * b

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
	if c.is_visible_in_tree():
		var r: Rect2 = c.get_global_rect()
		if r.position.x < vp.position.x - 1.0 or r.position.y < vp.position.y - 1.0 \
			or r.end.x > vp.end.x + 1.0 or r.end.y > vp.end.y + 1.0:
			out.append(p + " " + str(r))
	for ch in c.get_children():
		if ch is Control:
			out.append_array(_check_overflow(ch as Control, vp, p))
	return out
