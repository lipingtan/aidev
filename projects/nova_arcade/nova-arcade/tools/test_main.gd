extends Node
## T8 集成回归：实例化真实 shell/main.tscn，验证启动/四 Tab/Toast/🎨 主题切换。
## 视觉项（双视口比例无溢出）移交 T10 截图验证（Risk-1）。

var _failures: int = 0
var _main: Node

func _ready() -> void:
	var scene := load("res://shell/main.tscn") as PackedScene
	_main = scene.instantiate()
	add_child(_main)
	await get_tree().process_frame
	_run()

func _check(cond: bool, label: String) -> void:
	if cond:
		print("[test_main] PASS ", label)
	else:
		_failures += 1
		push_error("[test_main] FAIL ", label)

func _run() -> void:
	# 1. 启动进入 Home（RG-1）
	_check(Nav.current() == "res://shell/pages/home.tscn", "启动进入 Home")
	# 2. 四 Tab 切换无崩溃（RG-2）
	Nav.switch_tab(1)
	_check(Nav.current() == "res://shell/pages/category.tscn", "切分类")
	Nav.switch_tab(2)
	_check(Nav.current() == "res://shell/pages/search.tscn", "切搜索")
	Nav.switch_tab(3)
	_check(Nav.current() == "res://shell/pages/library.tscn", "切我的")
	Nav.switch_tab(0)
	_check(Nav.current() == "res://shell/pages/home.tscn", "回首页")
	# 3. Toast 时序：300ms 滑入 / 2200ms 后消失
	var label := _main.get_node("App/ToastLayer/Label") as Label
	(_main.get_node("App/ToastLayer") as Node).call("show_msg", "测试消息")
	await get_tree().create_timer(0.4).timeout
	_check(label.visible and label.modulate.a > 0.9, "Toast 滑入展示")
	await get_tree().create_timer(2.8).timeout
	_check(label.modulate.a < 0.1, "Toast 2200ms 后消失")
	# 4. 🎨 主题切换（RG-4）：音效 + 缩放反馈 + Token/profile/信号
	var before := ThemeTokens.current
	var btn := get_tree().root.find_child("ThemeButton", true, false) as Button
	_check(btn != null, "首页 🎨 按钮存在")
	if btn != null:
		btn.pressed.emit()
	_check(ThemeTokens.current != before, "🎨 主题切换生效")
	_check(Sound.last_play == "toggle", "🎨 播放切换音")
	_check(str(DB.get_profile().get("theme", "")) == ThemeTokens.current, "🎨 profile force 写")

	if _failures == 0:
		print("[test_main] ALL PASS")
	get_tree().quit(1 if _failures > 0 else 0)
