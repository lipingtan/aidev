extends Node
## T3 贪吃蛇 headless 测试（CR-6）：≥10 断言（逻辑 ≥6 + 协议 ≥4）。
## 用法：Godot --headless res://tools/_probe_snake.tscn  → PASS (failures=0) exit 0

var _failures: int = 0
var _over_stats: Dictionary = {}

func _check(cond: bool, label: String) -> void:
	if cond:
		print("[probe_snake] PASS ", label)
	else:
		_failures += 1
		push_error("[probe_snake] FAIL ", label)

func _ready() -> void:
	var s := load("res://games/snake/src/snake_game.gd")
	var G := s as GDScript

	# ---- logic layer (pure functions, deterministic RNG) ----
	var body: Array[Vector2i] = [Vector2i(3, 3), Vector2i(2, 3), Vector2i(1, 3)]
	var food := Vector2i(4, 3)
	var r: Dictionary = G.snake_step(body, Vector2i(1, 0), food, G.GRID)
	_check(bool(r["ate"]), "L1 step onto food → ate=true")
	var nb: Array[Vector2i] = r["body"] as Array[Vector2i]
	_check(nb.size() == 4, "L2 growth on eat (3→4)")
	_check(nb[0] == Vector2i(4, 3), "L3 head advances to food cell")

	var r2: Dictionary = G.snake_step(body, Vector2i(1, 0), Vector2i(9, 9), G.GRID)
	var nb2: Array[Vector2i] = r2["body"] as Array[Vector2i]
	_check(nb2.size() == 3 and not bool(r2["ate"]), "L4 no-eat step keeps length")

	var wall: Array[Vector2i] = [Vector2i(13, 3), Vector2i(12, 3), Vector2i(11, 3)]
	var rw: Dictionary = G.snake_step(wall, Vector2i(1, 0), Vector2i(0, 0), G.GRID)
	_check(bool(rw["dead"]), "L5 head hits right wall → dead")

	var selfc: Array[Vector2i] = [Vector2i(5, 5), Vector2i(6, 5), Vector2i(6, 6), Vector2i(5, 6), Vector2i(4, 5)]
	var rs: Dictionary = G.snake_step(selfc, Vector2i(1, 0), Vector2i(0, 0), G.GRID)
	_check(bool(rs["dead"]), "L6 self-collision → dead")

	var rng := RandomNumberGenerator.new()
	rng.seed = 12345
	var full: Array[Vector2i] = []
	for i in G.GRID:
		for j in G.GRID:
			full.append(Vector2i(i, j))
	var fpos: Vector2i = G.spawn_food(full, rng)
	_check(fpos == Vector2i(-1, -1), "L7 spawn_food on full board → (-1,-1) (no empty cell)")
	var sp: Vector2i = G.spawn_food([Vector2i(0, 0)] as Array[Vector2i], rng)
	_check(sp != Vector2i(0, 0), "L8 spawn_food avoids occupied cell")

	_check(absf(G.move_interval(0) - G.BASE_INTERVAL_MS) < 0.01, "L9 move_interval(0)=base")
	_check(G.move_interval(100) == G.MIN_INTERVAL_MS, "L10 speed floors at MIN_INTERVAL")
	_check(G.move_interval(5) < G.move_interval(0), "L11 speed increases with food")

	# ---- protocol layer (instantiate module scene, boot via adapter pattern) ----
	var mod: Node = load("res://games/snake/module.tscn").instantiate()
	add_child(mod)
	await get_tree().process_frame
	var src: Node = mod.get_node("Src")
	src.on_game_over.connect(func(st): _over_stats = st)
	src.SAVE_PATH = "user://save.cfg"
	src.start_game()
	_check(src.state == 1, "P1 start_game → PLAYING (state machine)")
	_check(src.body.size() == G.SPAWN_LEN, "P2 start_game spawns initial length")
	src.set_paused(true)
	_check(src.state == 2, "P3 set_paused → PAUSED")
	src.set_paused(false)
	_check(src.state == 1, "P4 resume → PLAYING")

	print("[probe_snake] %s (failures=%d)" % ["PASS" if _failures == 0 else "FAIL", _failures])
	get_tree().quit(1 if _failures > 0 else 0)
