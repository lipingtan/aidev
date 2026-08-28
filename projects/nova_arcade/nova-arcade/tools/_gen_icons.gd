extends Node
## T4 图标生成（CR-6）：headless GDScript Image 脚本，纯 set_pixel/fill API 无外部资源。
## 三款 icon.png 512×512 霓虹风（渐变底 + 主题 glyph），色值走 ThemeTokens(neon)。
## 注：4.7 headless Image 无 draw_* 方法 → glyphs built via pixel math (rect/circle/line).

const SZ := 512
var _failures: int = 0

func _ready() -> void:
	ThemeTokens.current = "neon"
	_gen("magic_tower", _glyph_tower)
	_gen("game_2048", _glyph_2048)
	_gen("snake", _glyph_snake)
	if _failures == 0:
		print("[gen_icons] PASS (3 icons written)")
	get_tree().quit(1 if _failures > 0 else 0)

## 渐变底（neon banner grad）+ vignette ring + glyph；落盘 res://games/{gid}/icon.png
func _gen(gid: String, glyph: Callable) -> void:
	var img := Image.create(SZ, SZ, false, Image.FORMAT_RGBA8)
	var stops: Array = ((ThemeTokens.GRADS["neon"] as Dictionary)["banner"] as Array)
	var c0: Color = stops[0] as Color
	var c1: Color = stops[stops.size() - 1] as Color
	for y in SZ:
		for x in SZ:
			img.set_pixel(x, y, c0.lerp(c1, (float(x) + float(y)) / (2.0 * SZ)))
	_ring(img, 40, 40, SZ - 80, SZ - 80, Color(0, 0, 0, 0.18), 24)
	glyph.call(img)
	var path := "res://games/%s/icon.png" % gid
	if img.save_png(path) != OK:
		_failures += 1
		push_error("[gen_icons] FAIL save %s" % path)
	else:
		print("[gen_icons] wrote %s (%dx%d)" % [path, img.get_width(), img.get_height()])

## filled rect (pixel math)
func _rect(img: Image, x0: int, y0: int, w: int, h: int, col: Color) -> void:
	for y in range(y0, y0 + h):
		for x in range(x0, x0 + w):
			if x >= 0 and y >= 0 and x < SZ and y < SZ:
				img.set_pixel(x, y, col)

## filled circle (pixel math)
func _circle(img: Image, cx: int, cy: int, r: int, col: Color) -> void:
	for y in range(cy - r, cy + r + 1):
		for x in range(cx - r, cx + r + 1):
			if (x - cx) * (x - cx) + (y - cy) * (y - cy) <= r * r:
				img.set_pixel(x, y, col)

## thick line via radius sampling
func _line(img: Image, a: Vector2i, b: Vector2i, col: Color, w: int) -> void:
	var d := b - a
	var steps := maxi(absi(d.x), absi(d.y))
	for i in range(steps + 1):
		var t := float(i) / float(maxi(steps, 1))
		var p := Vector2(a).lerp(Vector2(b), t)
		_circle(img, int(p.x), int(p.y), w, col)

## hollow ring (border only)
func _ring(img: Image, x0: int, y0: int, w: int, h: int, col: Color, t: int) -> void:
	_rect(img, x0, y0, w, t, col)
	_rect(img, x0, y0 + h - t, w, t, col)
	_rect(img, x0, y0, t, h, col)
	_rect(img, x0 + w - t, y0, t, h, col)

## 魔塔 glyph：tower silhouette (gold) + crenellations + gem window + door
func _glyph_tower(img: Image) -> void:
	var gold := ThemeTokens.color("gold")
	var cx := SZ / 2
	_rect(img, cx - 90, 150, 180, 300, gold)
	for i in range(-1, 2):
		_rect(img, cx + i * 70 - 26, 118, 52, 40, gold)
	_rect(img, cx - 34, 360, 68, 90, ThemeTokens.color("bg"))
	_rect(img, cx - 26, 210, 52, 52, ThemeTokens.color("accent"))

## 2048 glyph：rounded tile (alt purple) + inner lighter tile
func _glyph_2048(img: Image) -> void:
	var alt := ThemeTokens.color("alt")
	var bg := ThemeTokens.color("bg")
	var cx := SZ / 2
	_rect(img, cx - 130, 150, 260, 260, alt)
	for c in [Vector2i(cx - 130, 150), Vector2i(cx + 130, 150), Vector2i(cx - 130, 410), Vector2i(cx + 130, 410)]:
		for y in range(-48, 0):
			for x in range(-48, 0):
				var dx := absi(x + 48) - 48
				var dy := absi(y + 48) - 48
				if dx * dx + dy * dy > 48 * 48:
					var px: int = c.x + x
					var py: int = c.y + y
					if px >= 0 and py >= 0 and px < SZ and py < SZ:
						img.set_pixel(px, py, bg)
	_rect(img, cx - 108, 172, 216, 216, Color(alt.r, alt.g, alt.b, 0.55))

## 蛇 glyph：S-curve body (accent cyan) + gold head + food dot (alt)
func _glyph_snake(img: Image) -> void:
	var accent := ThemeTokens.color("accent")
	var pts := [Vector2i(150, 360), Vector2i(200, 400), Vector2i(300, 400), Vector2i(360, 340), Vector2i(360, 240), Vector2i(300, 180), Vector2i(200, 180)]
	for i in range(pts.size() - 1):
		_line(img, pts[i], pts[i + 1], accent, 22)
	_circle(img, 150, 360, 40, ThemeTokens.color("gold"))
	_circle(img, 138, 348, 8, ThemeTokens.color("bg"))
	_circle(img, 420, 250, 26, ThemeTokens.color("alt"))
