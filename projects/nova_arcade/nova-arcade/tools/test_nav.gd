extends Node
## T5 Nav 回归测试场景：headless 运行 res://tools/test_nav.tscn，全过 exit 0，任一失败 exit 1。
##
## 用法（APPDATA 指向临时目录）：
##   Godot --headless --path <proj> res://tools/test_nav.tscn
##
## 说明：T5 用占位页验证栈行为（P-4 决策），最终形态 T10 复验。

var _failures: int = 0
var _tab_events: Array[int] = []
const A_PATH := "res://tools/dummy_page_a.tscn"
const B_PATH := "res://tools/dummy_page_b.tscn"

func _ready() -> void:
	EventBus.tab_changed.connect(_on_tab_changed)
	# 测试场景在 Nav._ready 之后挂载，PageStack 需手动再定位一次
	Nav.locate_page_stack()
	await get_tree().process_frame
	_run_tests()

func _on_tab_changed(idx: int) -> void:
	_tab_events.append(idx)

## 断言：通过打印 PASS，失败计数并 push_error
func _check(cond: bool, label: String) -> void:
	if cond:
		print("[test_nav] PASS ", label)
	else:
		_failures += 1
		push_error("[test_nav] FAIL ", label)

## 等待转场（220ms）完成
func _settle() -> void:
	await get_tree().create_timer(0.35).timeout

func _run_tests() -> void:
	# Tab 根页覆盖为占位场景（Nav._tab_roots 实例变量，T8 前真实页面尚不存在）
	Nav._tab_roots = {0: A_PATH, 1: B_PATH}

	# 1. switch_tab(0)：根页入栈 + tab_changed(0)
	Nav.switch_tab(0)
	await _settle()
	_check(Nav.current() == A_PATH, "switch_tab(0) current")
	_check(_tab_events == [0], "tab_changed(0)")

	# 2. push：on_enter 收到 data、栈深 2
	Nav.push(B_PATH, {"k": 1})
	await _settle()
	_check(Nav.current() == B_PATH, "push current")
	var top_b := Nav._pages[1] as DummyPageB
	_check(top_b != null and top_b.enter_data.get("k", -1) == 1, "on_enter data")

	# 3. pop：顶层 on_exit、下层 on_resume、回根页
	var root_a := Nav._pages[0] as DummyPageA
	Nav.pop()
	# on_exit 同步调用先取值（页面在 220ms 转场结束后 queue_free，之后不可再访问）
	var b_exited: int = top_b.exit_count
	await _settle()
	_check(b_exited == 1, "pop on_exit")
	_check(root_a.resume_count == 1, "pop on_resume")
	_check(Nav.current() == A_PATH, "pop current")

	# 4. switch_tab(1)：清栈回目标根页 + tab_changed 顺序
	Nav.switch_tab(1)
	await _settle()
	_check(Nav._pages.size() == 1 and Nav.current() == B_PATH, "switch_tab(1) 清栈")
	_check(_tab_events == [0, 1], "tab_changed 顺序")

	# 5. pop_to_root：压两层后清空回根
	Nav.push(A_PATH, {})
	await _settle()
	Nav.pop_to_root()
	await _settle()
	_check(Nav._pages.size() == 1 and Nav.current() == B_PATH, "pop_to_root")

	# 6. handle_back：非根页默认 pop（消费）；根页返回 false
	Nav.push(A_PATH, {})
	await _settle()
	var consumed := Nav.handle_back()
	await _settle()
	_check(consumed and Nav._pages.size() == 1, "handle_back 消费 pop")
	_check(not Nav.handle_back(), "handle_back 根页返回 false")

	# 7. 转场时长常量 220ms
	_check(Nav.TRANSITION_MS == 0.22, "转场 220ms")

	if _failures == 0:
		print("[test_nav] ALL PASS (7 组断言)")
		get_tree().quit(0)
	else:
		push_error("[test_nav] FAILURES: %d" % _failures)
		get_tree().quit(1)
