extends Node
## CR-4 T11 headless 测试：Searcher 索引构建 + 打分查询 + 历史去重
## 运行：Godot_v4.7-stable_win64_console.exe --headless --path . --scene res://tools/test_search.tscn

var _failures: int = 0

func _check(cond: bool, label: String) -> void:
	if cond:
		print("[test_search] PASS  ", label)
	else:
		_failures += 1
		push_error("[test_search] FAIL  ", label)

func _ready() -> void:
	await get_tree().process_frame
	_run()

func _run() -> void:
	Registry.reload()
	await get_tree().process_frame

	## 构建索引
	var s := Searcher.new()
	add_child(s)
	s.build_index()

	## AC：query("") 返回空
	var r0 := s.query("")
	_check(r0.is_empty(), "query('') 返回空数组（size=%d）" % r0.size())

	## AC：query("tetra") 命中 tetra_nova（标题前缀）
	var r1 := s.query("tetra")
	_check(r1.has("tetra_nova"), "query('tetra') 命中 tetra_nova（结果=%s）" % str(r1))

	## AC：多次调用结果幂等（无副作用）
	var r1b := s.query("tetra")
	_check(r1b.has("tetra_nova"), "query('tetra') 第二次调用仍命中（幂等性）")

	## AC：大写查询折叠（to_lower）
	var r_upper := s.query("TETRA")
	_check(r_upper.has("tetra_nova"), "query('TETRA') 大写折叠命中 tetra_nova")

	## AC：hot_queries() 返回 ≤6 条
	var hq := s.hot_queries()
	_check(hq.size() <= 6, "hot_queries() size=%d（≤6）" % hq.size())

	## AC：hot_queries() 全部为字符串
	var hq_ok := true
	for item in hq:
		if not (item is String):
			hq_ok = false
			break
	_check(hq_ok, "hot_queries() 元素全为 String")

	## AC：DB.add_search_history 去重置顶
	DB.clear_search_history()
	DB.add_search_history("tetris")
	DB.add_search_history("rogue")
	DB.add_search_history("tetris")  ## 重复，应置顶
	var hist := DB.get_search_history()
	_check(hist.size() == 2, "add_search_history 去重：size=2（实际=%d）" % hist.size())
	_check(hist.size() > 0 and hist[0] == "tetris", "add_search_history 置顶：hist[0]='tetris'（实际='%s'）" % (hist[0] if hist.size() > 0 else ""))

	## AC：DB.clear_search_history 后为空
	DB.clear_search_history()
	_check(DB.get_search_history().is_empty(), "clear_search_history 后 get_search_history() 为空")

	## AC：历史上限 HISTORY_CAP=20
	DB.clear_search_history()
	for i in range(25):
		DB.add_search_history("q%d" % i)
	var hist2 := DB.get_search_history()
	_check(hist2.size() <= 20, "历史上限 ≤20，实际=%d" % hist2.size())
	DB.clear_search_history()

	s.queue_free()

	if _failures == 0:
		print("[test_search] ALL PASS")
	else:
		push_error("[test_search] FAILURES: %d" % _failures)
	get_tree().quit(1 if _failures > 0 else 0)
