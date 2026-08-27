extends Node
## CR-4 T11 headless 测试：Registry.query 扩展 + CategoryPage 筛选逻辑
## 运行：Godot_v4.7-stable_win64_console.exe --headless --path . --scene res://tools/test_category.tscn

var _failures: int = 0

func _check(cond: bool, label: String) -> void:
	if cond:
		print("[test_category] PASS  ", label)
	else:
		_failures += 1
		push_error("[test_category] FAIL  ", label)

func _ready() -> void:
	await get_tree().process_frame
	_run()

func _run() -> void:
	Registry.reload()
	await get_tree().process_frame

	## AC：Registry.query({}) 不报错，返回全量
	var all := Registry.query({})
	_check(all.size() > 0, "query({}) 返回全量，size=" + str(all.size()))

	## AC：Registry.query({category:""}) 等价于 all
	var all2 := Registry.query({"category": ""})
	_check(all2.size() == all.size(), "query({category:''}) 等价 all（%d==%d）" % [all2.size(), all.size()])

	## AC：price_models 过滤（free/ad/iap）
	var free_models: Array = ["free", "ad", "iap"]
	var free_games := Registry.query({"price_models": free_models})
	var expected := 0
	for m in Registry.all():
		if free_models.has((m as GameMeta).price_model):
			expected += 1
	_check(free_games.size() == expected,
		"query({price_models:[free,ad,iap]}) = %d 条（expected=%d）" % [free_games.size(), expected])

	## AC：price_models 精确过滤（只取 free）
	var only_free := Registry.query({"price_models": ["free"]})
	var all_free_ok := true
	for m in only_free:
		if (m as GameMeta).price_model != "free":
			all_free_ok = false
			break
	_check(all_free_ok, "query({price_models:[free]}) 结果全部 price_model=free")

	## AC：category 过滤（puzzle）
	var puzzle := Registry.query({"category": "puzzle"})
	var expected_puzzle := 0
	for m in Registry.all():
		if (m as GameMeta).category == "puzzle":
			expected_puzzle += 1
	_check(puzzle.size() == expected_puzzle,
		"query({category:'puzzle'}) = %d 条（expected=%d）" % [puzzle.size(), expected_puzzle])

	## AC：category + price_models 组合过滤
	var combo := Registry.query({"category": "puzzle", "price_models": ["free", "ad", "iap"]})
	var expected_combo := 0
	for m in Registry.all():
		var gm := m as GameMeta
		if gm.category == "puzzle" and free_models.has(gm.price_model):
			expected_combo += 1
	_check(combo.size() == expected_combo,
		"query 组合过滤 puzzle + free_models = %d 条（expected=%d）" % [combo.size(), expected_combo])

	## AC：空 price_models 不过滤（等价于不传 price_models）
	var empty_pm := Registry.query({"price_models": []})
	_check(empty_pm.size() == all.size(),
		"query({price_models:[]}) 不过滤，size 等于 all（%d==%d）" % [empty_pm.size(), all.size()])

	if _failures == 0:
		print("[test_category] ALL PASS")
	else:
		push_error("[test_category] FAILURES: %d" % _failures)
	get_tree().quit(1 if _failures > 0 else 0)
