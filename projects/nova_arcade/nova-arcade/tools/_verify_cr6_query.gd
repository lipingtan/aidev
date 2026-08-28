extends Node
## CR-6 T6 AC verification：Registry.query() + Searcher 命中三款新游戏（category/tags/aliases/pinyin）
## 运行：Godot_v4.7.2-stable_win64_console.exe --headless --path . res://tools/_verify_cr6_query.tscn

var _fails := 0

func _check(cond: bool, label: String) -> void:
	if cond:
		print("[vercr6] PASS ", label)
	else:
		_fails += 1
		push_error("[vercr6] FAIL ", label)

func _ready() -> void:
	await get_tree().process_frame
	Registry.reload()
	await get_tree().process_frame

	# --- Registry.query by category ---
	var puzzle := Registry.query({"category": "puzzle"})
	var p_ids := {}
	for m in puzzle:
		p_ids[str((m as GameMeta).id)] = true
	_check(puzzle.size() == 3 and p_ids.has("tetra_nova") and p_ids.has("magic_tower") and p_ids.has("game_2048"),
		"query(category=puzzle) 命中 tetra_nova+magic_tower+game_2048（size=%d）" % puzzle.size())

	var action := Registry.query({"category": "action"})
	var a_ids := {}
	for m in action:
		a_ids[str((m as GameMeta).id)] = true
	_check(a_ids.has("snake"), "query(category=action) 命中 snake（size=%d）" % action.size())

	# --- Registry.query by tag (neon 三款新游戏共有) ---
	var neon := Registry.query({"tag": "neon"})
	var n_ids := {}
	for m in neon:
		n_ids[str((m as GameMeta).id)] = true
	_check(n_ids.has("magic_tower") and n_ids.has("game_2048") and n_ids.has("snake"),
		"query(tag=neon) 命中三款新游戏（size=%d）" % neon.size())

	# --- Searcher by aliases / pinyin (title + alias + pinyin full/initials) ---
	var s := Searcher.new()
	add_child(s)
	s.build_index()
	_check(_has(s.query("魔塔"), "magic_tower"), "search('魔塔') 命中 magic_tower（alias/title）")
	_check(_has(s.query("MDX"), "magic_tower"), "search('MDX') 命中 magic_tower（alias）")
	_check(_has(s.query("mtx"), "magic_tower"), "search('mtx') 命中 magic_tower（pinyin initials）")
	_check(_has(s.query("2048"), "game_2048"), "search('2048') 命中 game_2048（title/alias）")
	_check(_has(s.query("数字消除"), "game_2048"), "search('数字消除') 命中 game_2048（alias）")
	_check(_has(s.query("tanchishe"), "snake"), "search('tanchishe') 命中 snake（pinyin full）")
	_check(_has(s.query("贪吃蛇"), "snake"), "search('贪吃蛇') 命中 snake（title/alias）")
	s.queue_free()

	print("[vercr6] 合计 FAIL=%d" % _fails)
	get_tree().quit(0 if _fails == 0 else 1)

func _has(arr: Array[String], gid: String) -> bool:
	return arr.has(gid)
