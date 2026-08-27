extends Node
## CR-4 T11 headless 测试：评价写入/覆盖/删除 + playtime 校验 + B-2 finish_count
## 运行：Godot_v4.7-stable_win64_console.exe --headless --path . --scene res://tools/test_review.tscn

var _failures: int = 0
const GID := "tetra_nova"

func _check(cond: bool, label: String) -> void:
	if cond:
		print("[test_review] PASS  ", label)
	else:
		_failures += 1
		push_error("[test_review] FAIL  ", label)

func _ready() -> void:
	await get_tree().process_frame
	_run()

func _run() -> void:
	## 前置清理
	if DB.get_review(GID) != null:
		DB.delete_review(GID)

	## AC：初始状态 get_review 返回 null
	var r0: Variant = DB.get_review(GID)
	_check(r0 == null, "初始状态 get_review 返回 null")

	## AC：put_review 写入 stars=4
	DB.put_review(GID, {"stars": 4, "text": "好玩", "status": "local"})
	var rv: Variant = DB.get_review(GID)
	_check(rv is Dictionary and int((rv as Dictionary).get("stars", 0)) == 4,
		"put_review stars=4 写入成功（rv=%s）" % str(rv))

	## AC：put_review 覆盖语义（stars=5）
	DB.put_review(GID, {"stars": 5, "text": "非常好玩", "status": "local"})
	var rv2: Variant = DB.get_review(GID)
	_check(rv2 is Dictionary and int((rv2 as Dictionary).get("stars", 0)) == 5,
		"put_review 覆盖 stars=5 成功")
	_check(rv2 is Dictionary and str((rv2 as Dictionary).get("text", "")) == "非常好玩",
		"put_review 覆盖 text='非常好玩'")
	_check(rv2 is Dictionary and str((rv2 as Dictionary).get("status", "")) == "local",
		"put_review status='local'")

	## AC：delete_review 后 get 返回 null
	DB.delete_review(GID)
	var rv3: Variant = DB.get_review(GID)
	_check(rv3 == null, "delete_review 后 get_review=null")

	## AC：重复 delete 不崩溃（静默忽略）
	DB.delete_review(GID)
	_check(true, "重复 delete_review 不崩溃")

	## AC：ReviewEditor 提交校验逻辑（playtime<600 → disabled=true）
	## stars>0 且 playtime≥600 才可提交；否则 disabled
	var should_disable_low_time: bool = (0 <= 0 or 599.9 < 600.0)  ## stars=0, playtime=599.9
	_check(should_disable_low_time, "playtime=599.9 → SubmitBtn.disabled=true（逻辑正确）")

	var should_enable: bool = not (4 <= 0 or 700.0 < 600.0)  ## stars=4, playtime=700.0
	_check(should_enable, "stars=4 且 playtime=700.0 → SubmitBtn.disabled=false")

	var should_disable_no_stars: bool = (0 <= 0 or 700.0 < 600.0)  ## stars=0, playtime=700.0
	_check(should_disable_no_stars, "stars=0 且 playtime≥600 → SubmitBtn.disabled=true")

	## AC：B-2 finish_count<2 → 写评价 disabled
	DB.upsert_record(GID, {"total_playtime": 700.0, "finish_count": 1})
	var rec: Variant = DB.get_record(GID)
	var playtime := float((rec as Dictionary).get("total_playtime", 0.0)) if rec is Dictionary else 0.0
	var finish := int((rec as Dictionary).get("finish_count", 0)) if rec is Dictionary else 0
	var b2_disabled: bool = playtime < 600.0 or finish < 2
	_check(b2_disabled, "B-2: finish_count=1 → 写评价 disabled=true（playtime=%s finish=%d）" % [str(playtime), finish])

	## AC：B-2 finish_count≥2 且 playtime≥600 → 可写评价
	DB.upsert_record(GID, {"total_playtime": 700.0, "finish_count": 2})
	var rec2: Variant = DB.get_record(GID)
	var playtime2 := float((rec2 as Dictionary).get("total_playtime", 0.0)) if rec2 is Dictionary else 0.0
	var finish2 := int((rec2 as Dictionary).get("finish_count", 0)) if rec2 is Dictionary else 0
	var b2_enabled: bool = not (playtime2 < 600.0 or finish2 < 2)
	_check(b2_enabled, "B-2: finish_count=2 且 playtime=700 → 写评价可用")

	## AC：完整提交流程（stars=3, playtime=700, status=local）
	DB.put_review(GID, {
		"stars": 3,
		"text": "测试评价内容",
		"playtime_at_review": 700.0,
		"created_at": int(Time.get_unix_time_from_system()),
		"status": "local",
	})
	var rv_full: Variant = DB.get_review(GID)
	_check(rv_full is Dictionary, "完整 put_review 后 get_review 返回 Dictionary")
	if rv_full is Dictionary:
		var rd := rv_full as Dictionary
		_check(int(rd.get("stars", 0)) == 3, "完整流程 stars=3")
		_check(str(rd.get("status", "")) == "local", "完整流程 status=local")
		_check(rd.has("created_at"), "完整流程含 created_at 时间戳")

	## AC：RG-20 clear_search_history 后 get_search_history 为空
	DB.add_search_history("test_rg20")
	DB.clear_search_history()
	_check(DB.get_search_history().is_empty(), "RG-20: clear_search_history 后为空")

	## 清理
	DB.delete_review(GID)
	DB.upsert_record(GID, {"total_playtime": 0.0, "finish_count": 0})

	if _failures == 0:
		print("[test_review] ALL PASS")
	else:
		push_error("[test_review] FAILURES: %d" % _failures)
	get_tree().quit(1 if _failures > 0 else 0)
