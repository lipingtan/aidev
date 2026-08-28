extends Node
## T2 临时探针：验证 game_2048.gd 纯函数层（grid_merge / is_game_over / spawn_tile）
func _ready() -> void:
	var fail := 0
	fail += check(Game2048.grid_merge([[2, 2, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0]], Vector2i(1, 0))["grid"] == [[0, 0, 0, 4], [0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0]], "merge right: [2,2]→[4] at right edge")
	fail += check(int(Game2048.grid_merge([[2, 2, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0]], Vector2i(1, 0))["score_delta"]) == 4, "merge right: score_delta=4")
	fail += check(Game2048.grid_merge([[2, 2, 2, 2], [0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0]], Vector2i(1, 0))["grid"] == [[0, 0, 4, 4], [0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0]], "merge right: each cell merges at most once")
	fail += check(Game2048.grid_merge([[0, 0, 0, 2], [0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0]], Vector2i(-1, 0))["grid"] == [[2, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0]], "merge left")
	fail += check(Game2048.grid_merge([[2, 0, 0, 0], [2, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0]], Vector2i(0, 1))["grid"] == [[0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0], [4, 0, 0, 0]], "merge down: at bottom edge")
	fail += check(Game2048.grid_merge([[2, 4, 8, 16], [2, 4, 8, 16], [0, 0, 0, 0], [0, 0, 0, 0]], Vector2i(0, -1))["grid"] == [[4, 8, 16, 32], [0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0]], "merge up: column pairs merge")
	fail += check(Game2048.is_game_over([[2, 4, 8, 16], [4, 8, 16, 32], [8, 16, 32, 64], [16, 32, 64, 128]]) == true, "over: no empty + no mergeable pair")
	fail += check(Game2048.is_game_over([[2, 4, 8, 16], [4, 8, 16, 32], [8, 16, 32, 64], [16, 32, 64, 64]]) == false, "not over: full grid with mergeable pair")
	var g: Array = Game2048._empty_grid()
	var r := RandomNumberGenerator.new()
	r.seed = 12345
	fail += check(Game2048.spawn_tile(g, r) == true and _count_tiles(g) == 1, "spawn_tile: exactly one tile placed")
	fail += check(not Game2048.is_game_over(g), "single tile grid not over")
	var full := [[2, 4, 8, 16], [4, 8, 16, 32], [8, 16, 32, 64], [16, 32, 64, 2]]
	fail += check(Game2048.spawn_tile(full) == false, "spawn_tile: full grid returns false")
	print("[probe_2048] %s (failures=%d)" % ["PASS" if fail == 0 else "FAIL", fail])
	get_tree().quit(1 if fail > 0 else 0)

func check(cond: bool, label: String) -> int:
	if cond:
		print("[probe_2048] PASS ", label)
	else:
		push_error("[probe_2048] FAIL ", label)
		return 1
	return 0

func _count_tiles(g: Array) -> int:
	var n := 0
	for i in g.size():
		for j in (g[i] as Array).size():
			if int((g[i] as Array)[j]) > 0:
				n += 1
	return n
