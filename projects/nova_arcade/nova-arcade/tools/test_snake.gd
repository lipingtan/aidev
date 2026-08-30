extends Node
## T5 贪吃蛇测试（CR-6）：headless 实例化 module.tscn（adapter+src）→ GameModule 协议 + 纯函数逻辑。
## 断言 ≥10：boot SAVE_PATH、pause/resume、reset_run、quit_requested 结构、snake_step/is_collision/spawn_food/move_interval。

var _failures := 0
var _pass := 0
var _got: Dictionary = {}

func _ready() -> void:
	var scene := load("res://games/snake/module.tscn") as PackedScene
	var mod := scene.instantiate()
	add_child(mod)
	await get_tree().process_frame
	_run(mod)

func _check(cond: bool, label: String) -> void:
	if cond:
		_pass += 1
		print("[snake_test] PASS ", label)
	else:
		_failures += 1
		push_error("[snake_test] FAIL ", label)

func _run(mod: Node) -> void:
	print("=== 贪吃蛇 T5（协议 + 逻辑）===")
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
	_check(_got.get("extra", {}).has("length"), "P9 extra.length 存在（蛇结算字段）")
	# 纯函数逻辑（显式 Array[Vector2i]，避免 duplicate() 返回非类型化数组）
	var body: Array[Vector2i] = [Vector2i(5, 7), Vector2i(4, 7), Vector2i(3, 7)]
	var body_copy := _copy(body)
	var r := SnakeGame.snake_step(body_copy, Vector2i(1, 0), Vector2i.ZERO, 14)
	_check(int(r["body"].size()) == 3 and int((r["body"] as Array[Vector2i])[0].x) == 6, "L1 snake_step 前进一步 head+1 长度不变")
	var r2 := SnakeGame.snake_step(_copy(body), Vector2i(1, 0), Vector2i(6, 7), 14)
	_check(int(r2["body"].size()) == 4 and bool(r2["ate"]), "L2 snake_step 吃 food → 长度+1 ate=true")
	_check(SnakeGame.is_collision([Vector2i(-1, 0), Vector2i(0, 0)], 14), "L3 is_collision 出界 → true")
	var rng := RandomNumberGenerator.new()
	rng.seed = 99
	var fp := SnakeGame.spawn_food(_copy(body), rng)
	_check(fp.x >= 0 and fp.y >= 0 and not body.has(fp), "L4 spawn_food 落在空格非蛇身")
	var full: Array[Vector2i] = []
	for i in 14:
		for j in 14:
			full.append(Vector2i(i, j))
	_check(SnakeGame.spawn_food(full, rng) == Vector2i(-1, -1), "L5 spawn_food 满盘 → (-1,-1)")
	_check(absf(SnakeGame.move_interval(0) - 260.0) < 0.01 and absf(SnakeGame.move_interval(100) - 90.0) < 0.01, "L6 move_interval base=260 floor=90")
	print("[snake_test] 合计 PASS=%d FAIL=%d" % [_pass, _failures])
	get_tree().quit(1 if _failures > 0 else 0)

## 拷贝为类型化 Array[Vector2i]（duplicate() 返回非类型化数组）
func _copy(src_arr: Array[Vector2i]) -> Array[Vector2i]:
	var out: Array[Vector2i] = []
	for p in src_arr:
		out.append(p)
	return out
