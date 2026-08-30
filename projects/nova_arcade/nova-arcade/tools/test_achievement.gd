extends Node
## 成就判定引擎 headless 测试（CR-7 T7）
## 断言≥10，PASS/FAIL 计数，quit(0/1)
## HEADLESS 测试规范：测试完成后 get_tree().quit() 显式退出 Godot 4.7 console 进程，释放服务器资源。

var _pass: int = 0
var _fail: int = 0
var _sig_test_aid: String = ""
var _sig_test_pts: int = 0

func assert_true(cond: bool, label: String) -> void:
	if cond:
		_pass += 1
		print("PASS %s" % label)
	else:
		_fail += 1
		push_error("FAIL %s" % label)

func _on_sig_test(def2: Dictionary, pts: int) -> void:
	_sig_test_aid = str(def2.get("id", ""))
	_sig_test_pts = pts

func _ready() -> void:
	var engine: Node = AchievementEngine.new()
	add_child(engine)
	await get_tree().create_timer(0.1).timeout

	# === 断言 1-3: _try_unlock 幂等 ===
	var aid: String = "tn_wave10"
	engine._try_unlock(aid)
	assert_true(DB.get_achievements().has(aid), "T7-A1: _try_unlock 首次解锁成功")

	var count_before: int = DB.get_achievements()[aid]["unlocked_at"]
	engine._try_unlock(aid)
	assert_true(DB.get_achievements()[aid]["unlocked_at"] == count_before, "T7-A2: _try_unlock 第二次幂等跳过")

	var size_before: int = DB.get_achievements().size()
	engine._try_unlock(aid)
	assert_true(DB.get_achievements().size() == size_before, "T7-A3: _try_unlock 第三次仍幂等（size不变）")

	# === 断言 4-5: game_finished result.achievements[] → Engine 正确解锁 ===
	var new_aid: String = "tn_boss3"
	DB.get_achievements().erase(aid)
	engine._on_game_finished("tetra_nova", {"achievements": [new_aid], "score": 100, "playtime": 60})
	assert_true(DB.get_achievements().has(new_aid), "T7-A4: game_finished → Engine 解锁成就")

	var def: Variant = Registry.ach_def(new_aid)
	assert_true(def != null and str(def.get("id", "")) == new_aid, "T7-A5: ach_def 返回正确定义")

	# === 断言 6-7: 收藏家条件（动态阈值）===
	for gid in ["tetra_nova", "magic_tower", "game_2048"]:
		DB.upsert_record(gid, {"last_played": 1})
	var game_count: int = Registry.all().size()
	assert_true(game_count >= 3, "T7-A6: Registry.all() ≥3 款游戏")
	engine._evaluate_global()
	var collector_unlocked: bool = DB.get_achievements().has("global_collector")
	assert_true(collector_unlocked or DB.list_records().size() < game_count, "T7-A7: 收藏家条件判定正确（动态阈值）")

	# === 断言 8-9: 好评人条件 ≥3条评价 ===
	for gid2 in ["tetra_nova", "magic_tower", "game_2048"]:
		DB.put_review(gid2, {"stars": 5})
	engine._evaluate_global()
	assert_true(DB.get_achievements().has("global_reviewer"), "T7-A8: 好评人条件（≥3条评价）解锁")

	# === 断言 10: 马拉松条件 ≥10h ===
	DB.upsert_record("tetra_nova", {"total_playtime": 36000})
	engine._evaluate_global()
	assert_true(DB.get_achievements().has("global_marathon"), "T7-A9: 马拉松条件（≥10h）解锁")

	# === 断言 11: achievement_unlocked 信号参数正确 ===
	EventBus.achievement_unlocked.connect(_on_sig_test)
	DB.get_achievements().erase("tn_wave10")
	engine._try_unlock("tn_wave10")
	assert_true(_sig_test_aid == "tn_wave10" and _sig_test_pts == 10, "T7-A10: achievement_unlocked 信号参数正确（def + points）")

	print("--- T7 achievement: %d PASS, %d FAIL ---" % [_pass, _fail])
	get_tree().quit(1 if _fail > 0 else 0)
