
extends Node
## 几何断言：24 格回字形布局（headless 可跑，不依赖渲染）

func _ready() -> void:
	var fails := 0
	var scene := load("res://games/slot_machine/module.tscn") as PackedScene
	var module := scene.instantiate()
	add_child(module)
	await get_tree().process_frame
	var src := module.get_node("Src") as Node2D
	# 收集 24 个符号格坐标（RingLayer/CanvasLayer 下，Cell{i}_{flag}，position 为绝对于层）
	var pts := {}
	for cell in module.get_node("Src/RingLayer").get_children():
		if not String(cell.name).begins_with("Cell"):
			continue
		var idx := int(String(cell.name).split("_")[0].trim_prefix("Cell"))
		pts[idx] = (cell as Node2D).position
	if pts.size() != 24:
		print("[geo] FAIL: expected 24 cells, got ", pts.size())
		get_tree().quit(1)
		return
	# 1) 无两格重叠（最小间距 > 60）
	var min_dist := 1e9
	for i in 24:
		for j in range(i + 1, 24):
			var d: float = (pts[i] as Vector2).distance_to(pts[j])
			if d < min_dist:
				min_dist = d
	print("[geo] min_dist=", min_dist)
	if min_dist < 60.0:
		fails += 1
	# 2) 四边存在：goslot 回字形，上边 y=360.8、下边 y=898.4、左列 x=86.4、右列 x=624.0
	#    （(格点+28)×1.6+(0,140)，START=(26,110)/STEPLEN=56 → x: 86.4/624.0, y: 360.8/898.4）
	var top := false
	var bottom := false
	var left := false
	var right := false
	for i in 24:
		var p: Vector2 = pts[i]
		if absf(p.y - 360.8) < 1.0: top = true
		if absf(p.y - 898.4) < 1.0: bottom = true
		if absf(p.x - 86.4) < 1.0: left = true
		if absf(p.x - 624.0) < 1.0: right = true
	print("[geo] top=", top, " bottom=", bottom, " left=", left, " right=", right)
	if not (top and bottom and left and right):
		fails += 1
	# 3) Lucky 格位置：小 Lucky 左列 idx21 (x=86.4)；大 Lucky 右列 idx9 (x=624.0)；两格同 y=629.6 对称
	var sl: Vector2 = pts[SlotConfig.LUCKY_IDX_SMALL]
	var bl: Vector2 = pts[SlotConfig.LUCKY_IDX_BIG]
	print("[geo] small_lucky=", sl, " big_lucky=", bl)
	if absf(sl.x - 86.4) > 1.0 or absf(bl.x - 624.0) > 1.0:
		fails += 1
	if absf(sl.y - 629.6) > 1.0 or absf(bl.y - 629.6) > 1.0:
		fails += 1
	print("[geo] RESULT fails=", fails)
	get_tree().quit(1 if fails > 0 else 0)
