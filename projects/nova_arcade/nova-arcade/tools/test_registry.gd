extends Node
## T4 Registry 回归测试场景：headless 运行 res://tools/test_registry.tscn，全过 exit 0。
##
## 用法（APPDATA 指向临时目录）：
##   Godot --headless --path <proj> res://tools/test_registry.tscn

var _failures: int = 0

func _ready() -> void:
	_run_tests()

func _check(cond: bool, label: String) -> void:
	if cond:
		print("[test_registry] PASS ", label)
	else:
		_failures += 1
		push_error("[test_registry] FAIL ", label)

func _run_tests() -> void:
	# 1. lookup("tetra_nova") 返回完整 meta（Registry._ready 已扫描）
	var m: GameMeta = Registry.lookup("tetra_nova")
	_check(m != null, "lookup(tetra_nova)")
	if m != null:
		print("[test_registry] meta: id=%s title=%s cat=%s runtime=%s trial=%s ver=%s" % [m.id, m.title, m.category, m.runtime, str(m.trial), m.version])
		_check(m.category == "puzzle" and m.runtime == "pck", "meta 关键字段")
		_check(int((m.trial as Dictionary).get("plays", -1)) == 3, "trial plays=3")
		_check(m.get_viewport_size() == Vector2i(720, 1560), "viewport 继承 Shell 默认")
		_check(m.get_orientation() == "portrait", "orientation 继承 portrait")

	# 2. query({category:"puzzle"}) 命中 tetra_nova
	var q: Array[GameMeta] = Registry.query({"category": "puzzle"})
	_check(q.size() == 1 and (q[0] as GameMeta).id == "tetra_nova", "query category=puzzle")

	# 3. all / installed_version / ach_def
	_check(Registry.all().size() == 1, "all size=1")
	_check(Registry.installed_version("tetra_nova") == "1.0.0", "installed_version")
	var ach: Dictionary = Registry.ach_def("tn_wave10")
	_check(ach != null and str(ach.get("name", "")) == "突破十波", "ach_def")
	_check(Registry.ach_def("not_exist") == null, "ach_def 未命中返回 null")

	# 4. editorial 空模板可读
	_check(Registry.get_editorial().has("banner"), "editorial 模板")

	if _failures == 0:
		print("[test_registry] ALL PASS")
		get_tree().quit(0)
	else:
		push_error("[test_registry] FAILURES: %d" % _failures)
		get_tree().quit(1)
