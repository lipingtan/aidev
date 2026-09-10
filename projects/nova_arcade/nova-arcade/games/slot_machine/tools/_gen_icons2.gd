extends Node
## CR-8 图标生成 v2：参照实物水果机风格——鲜艳立体水果、高光、金边黄底格
## 用法：--headless --path . --scene res://games/slot_machine/tools/_gen_icons2.tscn

const OUT := "res://games/slot_machine/assets/icons/"
const S := 128

func _ready() -> void:
	DirAccess.make_dir_recursive_absolute(ProjectSettings.globalize_path(OUT))
	_apple()
	_orange()
	_lemon()
	_tomato()
	_bell()
	_star()
	_bar25()
	_bar50()
	_seven()
	_lucky("lucky_big", Color(0.78, 0.05, 0.05), Color(1.0, 0.85, 0.2), "大LUCKY")
	_lucky("lucky_small", Color(0.75, 0.1, 0.1), Color(0.3, 0.95, 0.4), "LUCKY")
	print("[icons] all done")
	get_tree().quit(0)

func _img() -> Image:
	var img := Image.create(S, S, false, Image.FORMAT_RGBA8)
	img.fill(Color(0, 0, 0, 0))
	return img

## 实物风格底板：放射黄底 + 金色粗边框
func _plate(img: Image) -> void:
	var gold := Color(1.0, 0.78, 0.15)
	var yellow_c := Color(1.0, 0.92, 0.35)
	var yellow_e := Color(0.98, 0.72, 0.1)
	for x in range(4, S - 4):
		for y in range(4, S - 4):
			# 放射渐变（中心亮）
			var d := Vector2(x - 64, y - 64).length() / 90.0
			var c := yellow_c.lerp(yellow_e, clampf(d, 0.0, 1.0))
			img.set_pixel(x, y, c)
	# 金色边框（4 层粗边）
	for i in range(4):
		for x in range(4 + i, S - 4 - i):
			img.set_pixel(x, 4 + i, gold)
			img.set_pixel(x, S - 5 - i, gold)
		for y in range(4 + i, S - 4 - i):
			img.set_pixel(4 + i, y, gold)
			img.set_pixel(S - 5 - i, y, gold)

func _circle(img: Image, cx: float, cy: float, r: float, c: Color) -> void:
	for x in range(maxi(0, int(cx - r)), mini(S, int(cx + r) + 1)):
		for y in range(maxi(0, int(cy - r)), mini(S, int(cy + r) + 1)):
			if Vector2(x - cx, y - cy).length() <= r:
				img.set_pixel(x, y, c)

func _ellipse(img: Image, cx: float, cy: float, rx: float, ry: float, c: Color) -> void:
	for x in range(maxi(0, int(cx - rx)), mini(S, int(cx + rx) + 1)):
		for y in range(maxi(0, int(cy - ry)), mini(S, int(cy + ry) + 1)):
			var dx := (x - cx) / rx
			var dy := (y - cy) / ry
			if dx * dx + dy * dy <= 1.0:
				img.set_pixel(x, y, c)

## 立体高光：左上白色大椭圆 + 环形阴影
func _shine3d(img: Image, cx: float, cy: float, rx: float, ry: float) -> void:
	_ellipse(img, cx, cy, rx, ry, Color(1, 1, 1, 0.75))
	_ellipse(img, cx + rx * 0.25, cy + ry * 0.3, rx * 0.55, ry * 0.5, Color(1, 1, 1, 0.4))

func _leaf(img: Image, cx: float, cy: float, flip := 1.0) -> void:
	for i in range(22):
		_ellipse(img, cx + flip * i * 0.9, cy - i * 0.55, 13 - i * 0.45, 5.5 - i * 0.18, Color(0.12, 0.62, 0.12))
	_ellipse(img, cx + flip * 6, cy - 3, 9, 3.5, Color(0.3, 0.8, 0.25))

func _save(img: Image, name_: String) -> void:
	var err := img.save_png(ProjectSettings.globalize_path(OUT + name_ + ".png"))
	print("[icons] ", name_, " err=", err)

## 苹果：双圆心形、深红渐变、高光、褐柄绿叶
func _apple() -> void:
	var img := _img()
	_plate(img)
	var red := Color(0.88, 0.08, 0.08)
	var red_d := Color(0.6, 0.03, 0.03)
	var red_l := Color(1.0, 0.25, 0.2)
	# 底层暗红（立体感）
	_circle(img, 48, 74, 34, red_d)
	_circle(img, 80, 74, 34, red_d)
	# 主色
	_circle(img, 48, 70, 31, red)
	_circle(img, 80, 70, 31, red)
	_ellipse(img, 64, 72, 14, 32, red)
	# 亮面（右上偏移红）
	_ellipse(img, 74, 58, 16, 14, red_l)
	_shine3d(img, 46, 54, 15, 10)
	# 柄
	for i in range(16):
		for w in range(4):
			var x := 63 + int(i * 0.25) + w
			var y := 38 - i
			if x < S and y >= 0:
				img.set_pixel(x, y, Color(0.4, 0.25, 0.1))
	_leaf(img, 72, 32)
	_save(img, "apple")

## 橙子：亮橙 + 脐点 + 光泽
func _orange() -> void:
	var img := _img()
	_plate(img)
	var c := Color(1.0, 0.6, 0.05)
	var c_d := Color(0.8, 0.42, 0.02)
	_circle(img, 64, 74, 40, c_d)
	_circle(img, 64, 70, 38, c)
	_ellipse(img, 64, 86, 28, 16, c_d)  # 下部暗面
	_circle(img, 64, 34, 5, Color(0.85, 0.5, 0.05))  # 脐
	_leaf(img, 66, 32)
	_shine3d(img, 48, 54, 17, 12)
	_save(img, "orange")

## 柠檬：两头尖的椭圆、亮黄
func _lemon() -> void:
	var img := _img()
	_plate(img)
	var c := Color(0.92, 0.92, 0.12)
	var c_d := Color(0.7, 0.72, 0.05)
	_ellipse(img, 64, 72, 46, 34, c_d)
	_ellipse(img, 64, 68, 44, 31, c)
	# 两头凸起
	_ellipse(img, 18, 66, 8, 6, c)
	_ellipse(img, 110, 62, 8, 6, c)
	_ellipse(img, 64, 84, 30, 12, c_d)
	_shine3d(img, 46, 54, 18, 10)
	_leaf(img, 106, 54, -1.0)
	_save(img, "lemon")

## 番茄：扁圆多瓣 + 绿星蒂
func _tomato() -> void:
	var img := _img()
	_plate(img)
	var c := Color(0.92, 0.12, 0.1)
	var c_d := Color(0.65, 0.05, 0.05)
	_ellipse(img, 64, 76, 44, 36, c_d)
	_ellipse(img, 64, 72, 42, 34, c)
	# 瓣状竖弧
	for i in [-1, 0, 1]:
		_ellipse(img, 64 + i * 22, 74, 10, 30, c.lerp(c_d, 0.25))
	_ellipse(img, 64, 56, 12, 10, Color(1.0, 0.4, 0.35))
	_shine3d(img, 46, 60, 14, 9)
	# 星形蒂
	for i in range(6):
		var a := deg_to_rad(i * 60.0 - 90.0)
		_ellipse(img, 64 + cos(a) * 14, 44 + sin(a) * 7, 11, 4.5, Color(0.15, 0.6, 0.12))
	_circle(img, 64, 44, 5, Color(0.1, 0.45, 0.1))
	_save(img, "tomato")

## 铃铛：金钟立体 + 双高光
func _bell() -> void:
	var img := _img()
	_plate(img)
	var gold := Color(1.0, 0.82, 0.1)
	var gold_d := Color(0.85, 0.6, 0.02)
	var gold_l := Color(1.0, 0.95, 0.55)
	# 钟体（馒头形）
	_ellipse(img, 64, 66, 38, 36, gold_d)
	_ellipse(img, 64, 62, 36, 33, gold)
	_ellipse(img, 52, 52, 14, 12, gold_l)
	# 扩口缘
	_ellipse(img, 64, 96, 44, 10, gold_d)
	_ellipse(img, 64, 94, 42, 8, gold)
	# 摆锤
	_circle(img, 64, 108, 8, gold_d)
	# 顶钮
	_circle(img, 64, 26, 8, gold_d)
	_shine3d(img, 46, 44, 13, 9)
	_save(img, "bell")

## 星星：金色五角星 + 内层高光
func _star() -> void:
	var img := _img()
	_plate(img)
	var gold := Color(1.0, 0.85, 0.1)
	var gold_d := Color(0.9, 0.62, 0.0)
	var pts: Array[Vector2] = []
	for i in range(10):
		var a := deg_to_rad(i * 36.0 - 90.0)
		var r := 48.0 if i % 2 == 0 else 20.0
		pts.append(Vector2(64, 68) + Vector2.from_angle(a) * r)
	var centroid := Vector2(64, 68)
	for x in range(10, 118):
		for y in range(14, 122):
			var p := Vector2(x, y)
			var inside := false
			for i in range(5):
				if _tri_contains(centroid, pts[i * 2], pts[(i * 2 + 2) % 10], p):
					inside = true
					break
			if inside:
				img.set_pixel(x, y, gold)
	# 内层亮星
	var pts2: Array[Vector2] = []
	for i in range(10):
		var a := deg_to_rad(i * 36.0 - 90.0)
		var r := 28.0 if i % 2 == 0 else 12.0
		pts2.append(Vector2(62, 64) + Vector2.from_angle(a) * r)
	for x in range(30, 100):
		for y in range(34, 100):
			var p := Vector2(x, y)
			var inside := false
			for i in range(5):
				if _tri_contains(Vector2(62, 64), pts2[i * 2], pts2[(i * 2 + 2) % 10], p):
					inside = true
					break
			if inside:
				img.set_pixel(x, y, Color(1.0, 0.98, 0.6))
	_save(img, "star")

func _tri_contains(a: Vector2, b: Vector2, c: Vector2, p: Vector2) -> bool:
	var d1 := _sign(p, a, b)
	var d2 := _sign(p, b, c)
	var d3 := _sign(p, c, a)
	return not (((d1 < 0) or (d2 < 0) or (d3 < 0)) and ((d1 > 0) or (d2 > 0) or (d3 > 0)))

func _sign(p1: Vector2, p2: Vector2, p3: Vector2) -> float:
	return (p1.x - p3.x) * (p2.y - p3.y) - (p2.x - p3.x) * (p1.y - p3.y)

## BAR：黑圆角块 + 三重立体白条 + 金倍数
func _bar_img(mult: String, file: String) -> void:
	var img := _img()
	_plate(img)
	# 黑色圆角块
	var black := Color(0.06, 0.06, 0.07)
	for x in range(14, 114):
		for y in range(18, 88):
			# 圆角检测
			var dx := minf(x - 14, 113 - x)
			var dy := minf(y - 18, 87 - y)
			if dx >= 0 and dy >= 0 and (dx > 10 or dy > 10 or Vector2(10 - dx, 10 - dy).length() <= 10):
				img.set_pixel(x, y, black)
	# 三重白条（带灰底立体）
	for row in range(3):
		var y0 := 24 + row * 20
		for x in range(22, 106):
			for y in range(y0, y0 + 14):
				img.set_pixel(x, y, Color(0.75, 0.75, 0.78))
			for y in range(y0, y0 + 11):
				img.set_pixel(x, y, Color(0.97, 0.97, 0.99))
		# 字母镂空 B-A-R（黑竖条简化）
		for gx in [34, 50, 66, 82, 94]:
			for y in range(y0 + 1, y0 + 10):
				img.set_pixel(gx, y, black)
	# 金色倍数横幅
	var gold := Color(1.0, 0.8, 0.1)
	for x in range(24, 104):
		for y in range(92, 118):
			img.set_pixel(x, y, gold)
	# ×N 黑字点阵
	_dot_x(img, 34, 96)
	_digit_dot(img, 58, 94, mult)
	_save(img, file)

func _bar25() -> void:
	_bar_img("25", "bar25")

func _bar50() -> void:
	_bar_img("50", "bar50")

## ×号（点阵）
func _dot_x(img: Image, ox: int, oy: int) -> void:
	for i in range(5):
		for w in range(3):
			_set_px(img, ox + i * 3 + w, oy + i * 3 + w)
			_set_px(img, ox + 12 - i * 3 - w, oy + i * 3 + w)

## 数字点阵（3x5）
func _digit_dot(img: Image, ox: int, oy: int, digits: String) -> void:
	var glyphs := {
		"2": ["111", "001", "111", "100", "111"],
		"5": ["111", "100", "111", "001", "111"],
	}
	for d_i in digits.length():
		var g: Array = glyphs.get(digits[d_i], ["000", "000", "000", "000", "000"])
		var dx := ox + d_i * 14
		for ry in range(5):
			for rx in range(3):
				if g[ry][rx] == "1":
					for px_ in range(4):
						for py in range(4):
							_set_px(img, dx + rx * 4 + px_, oy + ry * 4 + py)

func _set_px(img: Image, x: int, y: int) -> void:
	if x >= 0 and x < S and y >= 0 and y < S:
		img.set_pixel(x, y, Color(0.1, 0.1, 0.1))

## 77：双立体红七
func _seven() -> void:
	var img := _img()
	_plate(img)
	var red := Color(0.92, 0.08, 0.1)
	var red_d := Color(0.6, 0.02, 0.04)
	for off in [14, 70]:
		# 暗红底版（偏移 2px 立体感）
		_seven_stroke(img, off + 3, 4, red_d)
		_seven_stroke(img, off, 0, red)
	_save(img, "seven")

func _seven_stroke(img: Image, off: int, down: int, c: Color) -> void:
	for i in range(34):
		for w in range(11):
			var x: int = off + i + w
			var y: int = 28 + down + w
			if x < S and y < S:
				img.set_pixel(x, y, c)   # 顶横
			var x2: int = off + 34 - i - w
			var y2: int = 30 + down + i
			if x2 > 0 and x2 < S and y2 < 120:
				img.set_pixel(x2, y2, c)   # 斜笔

## Lucky：红底 + 金字横幅 + 星光底纹
func _lucky(file: String, bg: Color, accent: Color, label: String) -> void:
	var img := _img()
	_plate(img)
	# 中央红底块（圆角）
	var dark := bg.darkened(0.4)
	for x in range(12, 116):
		for y in range(12, 116):
			var dx := minf(x - 12, 115 - x)
			var dy := minf(y - 12, 115 - y)
			if dx >= 0 and dy >= 0 and (dx > 12 or dy > 12 or Vector2(12 - dx, 12 - dy).length() <= 12):
				img.set_pixel(x, y, bg)
	# 星光放射（暗红条）
	for i in range(8):
		var a := deg_to_rad(i * 45.0 + 22.5)
		for rr in range(40, 62):
			for w in range(4):
				var x := int(64 + cos(a) * rr) - w / 2
				var y := int(48 + sin(a) * rr) - w / 2
				if x >= 0 and x < S and y >= 0 and y < S:
					img.set_pixel(x, y, dark)
	# 金圈
	for rr in range(30, 34):
		for a_deg in range(360):
			var a := deg_to_rad(float(a_deg))
			var x := int(64 + cos(a) * rr)
			var y := int(50 + sin(a) * rr)
			if x >= 0 and x < S and y >= 0 and y < S:
				img.set_pixel(x, y, accent)
	# 金色横幅字条
	for x in range(10, 118):
		for y in range(88, 116):
			img.set_pixel(x, y, accent)
	# 字母块（白/黑点阵简化：LUCKY 5 字母 + 大L 的"大"）
	if label == "大LUCKY":
		for x in range(14, 38):
			for y in range(92, 112):
				img.set_pixel(x, y, Color(0.1, 0.1, 0.1))  # 「大」占位块
		_dot_lucky(img, 44, 92)
	else:
		_dot_lucky(img, 20, 92)
	_save(img, file)

func _dot_lucky(img: Image, ox: int, oy: int) -> void:
	# L U C K Y 五字母点阵（每个 3x5，间隔 1 列）→ 简化为 L̶ 三笔 + 块
	var glyphs := {
		"L": ["100", "100", "100", "100", "111"],
		"U": ["101", "101", "101", "101", "111"],
		"C": ["111", "100", "100", "100", "111"],
		"K": ["101", "101", "110", "101", "101"],
		"Y": ["101", "101", "010", "010", "010"],
	}
	var word := "LUCKY"
	var cx := ox
	for ch in word:
		var g: Array = glyphs[ch]
		for ry in range(5):
			for rx in range(3):
				if g[ry][rx] == "1":
					for px_ in range(3):
						for py in range(3):
							var x := cx + rx * 3 + px_
							var y := oy + ry * 3 + py
							if x < S and y < S:
								img.set_pixel(x, y, Color(0.1, 0.1, 0.1))
		cx += 12
