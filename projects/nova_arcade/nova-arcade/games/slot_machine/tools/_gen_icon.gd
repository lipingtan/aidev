extends Node
## icon.png 生成（CR-8 T1）：512×512 自绘——方形面板+圆环+指针意象，neon 风格
## 用法：--headless --path . --scene res://games/slot_machine/tools/_gen_icon.tscn 之类；
## 此处按 CR-6 _gen_icons 模式放 tools 内跑一次

func _ready() -> void:
	var img := Image.create(512, 512, false, Image.FORMAT_RGBA8)
	var bg := Color(0.02, 0.024, 0.059)
	var panel := Color(0.039, 0.063, 0.133)
	var ring := Color(0.169, 0.91, 1)
	var gold := Color(1, 0.824, 0.247)
	# 背景
	img.fill(bg)
	# 方形面板（居中 400×400 圆角感：先填方块再描边）
	var ox := 56
	var oy := 56
	var side := 400
	for x in range(side):
		for y in range(side):
			img.set_pixel(ox + x, oy + y, panel)
	# 圆环（半径 150，宽 10）
	var cx := 256.0
	var cy := 256.0
	var r := 150.0
	for x in range(512):
		for y in range(512):
			var d := sqrt(float(x - cx) * (x - cx) + float(y - cy) * (y - cy))
			if absf(d - r) < 5.0:
				img.set_pixel(x, y, ring)
			# 12 个符号位圆点
			for i in 12:
				var a := deg_to_rad(i * 30.0 - 90.0)
				var px := cx + cos(a) * r
				var py := cy + sin(a) * r
				if sqrt(float(x - px) * (x - px) + float(y - py) * (y - py)) < 12.0:
					img.set_pixel(x, y, gold)
	# 顶部指针
	for x in range(242, 270):
		for y in range(66, 96):
			if absi(x - 256) < (y - 66) / 2:
				img.set_pixel(x, y, gold)
	var err := img.save_png(ProjectSettings.globalize_path("res://games/slot_machine/icon.png"))
	print("[icon] saved err=", err)
	get_tree().quit(0 if err == OK else 1)
