extends Node
## T10 集成冒烟（§5 清单子集 + RG-1~6）：驱动真实 shell/main.tscn。
## 覆盖：§5.2 四 Tab、RG-1 启动/Home、RG-2 Tab 切换、RG-3 push/pop、RG-4 主题切换+persist、RG-6 CoreManager Mock。
## §5.1 需 Godot 编辑器（仅 headless 可验 autoload 无错）；§5.3 详情页 CTA / §5.4 Launcher 生命周期 / §5.5 视觉跳变
## → 依赖 Services 层/详情页/真机，CR-1 骨架不可验，见 smoke_report.md。

var _failures := 0
var _pass := 0

func _ready() -> void:
	var scene := load("res://shell/main.tscn") as PackedScene
	var main := scene.instantiate()
	add_child(main)
	await get_tree().process_frame
	await get_tree().process_frame
	_run()

func _check(cond: bool, label: String) -> void:
	if cond:
		_pass += 1
		print("[smoke] PASS ", label)
		return
	_failures += 1
	push_error("[smoke] FAIL ", label)

func _run() -> void:
	print("=== §5 冒烟 + RG-1~6（CR-1 骨架）===")
	# §5.2 / RG-1：启动进入 Home
	_check(Nav.current() == "res://shell/pages/home.tscn", "RG-1 启动进入 Home（§5.2）")
	# RG-2：四 Tab 切换无崩溃
	var tabs := [["1","category"],["2","search"],["3","library"],["0","home"]]
	for t in tabs:
		Nav.switch_tab(int(t[0]))
		_check(Nav.current() == "res://shell/pages/%s.tscn" % t[1], "RG-2 切 Tab(%s)" % t[1])
	# RG-3：push/pop 页面（占位 page_a，Page 实例）
	Nav.push("res://tools/dummy_page_a.tscn", {"title":"push"})
	_check(Nav.current() == "res://tools/dummy_page_a.tscn", "RG-3 push 入栈")
	Nav.pop()
	_check(Nav.current() == "res://shell/pages/home.tscn", "RG-3 pop 回 Home")
	# RG-4：主题切换 + profile force 写（复用首页 🎨 按钮）
	var btn := get_tree().root.find_child("ThemeButton", true, false) as Button
	_check(btn != null, "RG-4 首页 🎨 按钮存在")
	if btn != null:
		var before := ThemeTokens.current
		btn.pressed.emit()
		_check(ThemeTokens.current != before, "RG-4 主题切换生效")
		_check(str(DB.get_profile().get("theme","")) == ThemeTokens.current, "RG-4 profile force 写（可持久化）")
	# RG-5：强写接口（实际持久化见 test_rg5 双相）
	DB.upsert_record("tetra_nova", {"trial_used":5,"total_playtime":1000,"best":999,"finish_count":3})
	_check(int(DB.get_record("tetra_nova").get("trial_used",-1)) == 5, "RG-5 upsert_record 强写可落盘")
	# RG-6：CoreManager Mock
	var cm := CoreManager.new()
	_check(cm.is_installed("0.288") == false, "RG-6 CoreManager Mock 返回未安装")
	print("[smoke] 合计 PASS=%d FAIL=%d" % [_pass, _failures])
	get_tree().quit(1 if _failures > 0 else 0)
