extends Node
## 推荐画像 headless 测试（CR-7 T7）
## 断言≥10，PASS/FAIL 计数，quit(0/1)
## HEADLESS 测试规范：测试完成后 get_tree().quit() 显式退出 Godot 4.7 console 进程，释放服务器资源。

var _pass: int = 0
var _fail: int = 0

func assert_true(cond: bool, label: String) -> void:
	if cond:
		_pass += 1
		print("PASS %s" % label)
	else:
		_fail += 1
		push_error("FAIL %s" % label)

func _ready() -> void:
	var rec: Node = Recommender.new()
	add_child(rec)
	await get_tree().create_timer(0.1).timeout

	# === 断言 1-3: _build_profile 从游玩记录构建类目/标签偏好 ===
	DB.upsert_record("tetra_nova", {"last_played": 1})
	var profile: Dictionary = rec._build_profile()
	assert_true(profile.has("cats") and profile.has("tags"), "T7-R1: _build_profile 返回 cats+tags")

	var cats: Dictionary = profile.get("cats", {}) as Dictionary
	assert_true(cats.has("puzzle") and int(cats["puzzle"]) > 0, "T7-R2: 类目偏好包含 puzzle")

	var tags: Dictionary = profile.get("tags", {}) as Dictionary
	assert_true(tags.size() > 0, "T7-R3: 标签偏好非空")

	# === 断言 4-5: _score_for 同类目游戏得分高于不同类目 ===
	var puzzle_meta: Variant = Registry.lookup("tetra_nova")
	var action_meta: Variant = Registry.lookup("snake")
	assert_true(puzzle_meta != null and action_meta != null, "T7-R4: meta lookup 成功")

	var score_puzzle: float = rec._score_for(puzzle_meta as GameMeta, profile)
	var score_action: float = rec._score_for(action_meta as GameMeta, profile)
	assert_true(score_puzzle > score_action, "T7-R5: 同类目游戏得分高于不同类目")

	# === 断言 6-7: for_you 未玩过优先排序 ===
	DB.upsert_record("magic_tower", {"last_played": 1})
	rec._cache_dirty = true
	var results: Array[Dictionary] = rec.for_you()
	assert_true(results.size() > 0, "T7-R6: for_you 返回非空")

	if results.size() > 0:
		var first_rec: Variant = results[0].get("record")
		var first_played: bool = (first_rec is Dictionary) and int((first_rec as Dictionary).get("last_played", 0)) > 0
		assert_true(not first_played or results.size() <= 2, "T7-R7: for_you 未玩过优先排序")

	# === 断言 8-9: 冷启动回退（无记录时返回非空）===
	for gid in DB.list_records():
		DB.upsert_record(gid, {})
	rec._cache_dirty = true
	var cold_results: Array[Dictionary] = rec.for_you()
	assert_true(cold_results.size() > 0, "T7-R8: 冷启动回退返回非空")
	assert_true(cold_results.size() <= 4, "T7-R9: 冷启动回退 ≤4条")

	# === 断言 10: for_you 格式正确（每条含有效 meta + record）===
	DB.upsert_record("tetra_nova", {"last_played": 1})
	rec._cache_dirty = true
	var final_results: Array[Dictionary] = rec.for_you()
	var all_valid: bool = true
	for item in final_results:
		var meta: Variant = item.get("meta")
		if meta == null or not (meta is GameMeta):
			all_valid = false
			break
	assert_true(all_valid and final_results.size() <= 4, "T7-R10: for_you 格式正确（每条含有效 meta）")

	print("--- T7 recommender: %d PASS, %d FAIL ---" % [_pass, _fail])
	get_tree().quit(1 if _fail > 0 else 0)
