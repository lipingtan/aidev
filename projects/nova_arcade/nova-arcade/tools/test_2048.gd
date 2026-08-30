extends Node
## T5 2048 测试（CR-6）：headless 实例化 module.tscn（adapter+src）→ GameModule 协议 + 纯函数逻辑。
## 断言 ≥10：boot SAVE_PATH、pause/resume、reset_run、quit_requested 结构、grid_merge/is_game_over/spawn_tile。

var _failures := 0
var _pass := 0
var _got: Dictionary = {}

func _ready() -> void:
	var scene := load("res://games/game_2048/module.tscn") as PackedScene
	var mod := scene.instantiate()
	add_child(mod)
	await get_tree().process_frame
	_run(mod)

func _check(cond: bool, label: String) -> void:
	if cond:
		_pass += 1
		print("[2048_test] PASS ", label)
	else:
		_failures += 1
		push_error("[2048_test] FAIL ", label)

func _run(mod: Node) -> void:
	print("=== 2048 T5（协议 + 逻辑）===")
	var src := mod.get_node("Src") as Node
	_check(mod is GameModule, "P1 module 是 GameModule")
	_check(src != null and src.has_signal("on_game_over"), "P2 src 存在且 on_game_over 信号")
	mod.boot({"save_dir": "user://", "trial_mode": true, "owned": false, "best": 0, "viewport_size": Vector2i(720, 1560)})
	_check(src.get("SAVE_PATH") == "user://save.cfg", "P3 boot 注入 SAVE_PATH")
	# boot 后停在开始遮罩（等待玩家点「开始游戏」），不自动开局
	src.start_game()   # 模拟玩家点击开始
	_check(int(src.get("state")) == int(src.State.PLAYING), "P4 boot 后 state=PLAYING")
	mod.pause_game()
	_check(int(src.get("state")) == int(src.State.PAUSED), "P5 pause_game → PAUSED")
	mod.resume_game()
	_check(int(src.get("state")) == int(src.State.PLAYING), "P6 resume_game → PLAYING")
	_got = {}
	mod.quit_requested.connect(func(r: Dictionary): _got = r)
	mod.quit_to_shell()
	_check(_got.has("score") and _got.has("playtime") and _got.has("achievements") and _got.has("extra"), "P7 quit_requested 四键齐全")
	_check((_got.get("achievements", []) as Array).is_empty(), "P8 achievements 恒空数组（M2 预留）")
	_check(_got.get("extra", {}).has("max_tile"), "P9 extra.max_tile 存在（2048 结算字段）")
	# 纯函数逻辑
	var g: Array = []
	for i in 4:
		g.append([0, 0, 0, 0])
	g[3][2] = 2
	var r := Game2048.grid_merge(g.duplicate(true), Vector2i(-1, 0))
	_check(int((r["grid"] as Array)[3][0]) == 2 and int(r["score_delta"]) == 0, "L1 grid_merge 左移单 tile 滑到边")
	var g2: Array = []
	for i in 4:
		g2.append([0, 0, 0, 0])
	g2[2][3] = 2
	g2[2][2] = 2
	var r2 := Game2048.grid_merge(g2.duplicate(true), Vector2i(-1, 0))
	_check(int((r2["grid"] as Array)[2][0]) == 4 and int(r2["score_delta"]) == 4, "L2 grid_merge 相邻同值合并 +delta")
	var g3: Array = []
	for i in 4:
		var row: Array = []
		for j in 4:
			row.append(2 if (i + j) % 2 == 0 else 4)
		g3.append(row)
	_check(Game2048.is_game_over(g3), "L3 is_game_over 满盘无相邻同值 → true")
	var g4: Array = []
	for i in 4:
		g4.append([0, 0, 0, 0])
	var rng := RandomNumberGenerator.new()
	rng.seed = 7
	var ok := Game2048.spawn_tile(g4, rng)
	var sum := 0
	for i in 4:
		for j in 4:
			sum += int(g4[i][j])
	_check(ok and (sum == 2 or sum == 4), "L4 spawn_tile 空盘生成 2/4")
	print("[2048_test] 合计 PASS=%d FAIL=%d" % [_pass, _failures])
	get_tree().quit(1 if _failures > 0 else 0)
