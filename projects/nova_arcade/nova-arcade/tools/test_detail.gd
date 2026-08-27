extends Node
## CR-3 T7 headless 全覆盖测试（design.md §11）
## 覆盖：CTA 五分支 + best hint / 继续游戏区块 / 长按移除
##       records_updated 重渲染 / Nav push/pop / on_resume 刷新
## 运行：Godot_v4.7-stable_win64_console.exe --headless --path . --scene res://tools/test_detail.tscn

const GID := "tetra_nova"          # price_model=trial, trial.plays=3
const GID_FREE := "_test_free"     # 临时注入：price_model=free
const GID_PAID := "_test_paid"     # 临时注入：price_model=paid, price=10

var _failures: int = 0
var _main: Node

# ── 工具 ──────────────────────────────────────────────────────────────────────

func _check(cond: bool, label: String) -> void:
	if cond:
		print("[test_detail] PASS  ", label)
	else:
		_failures += 1
		push_error("[test_detail] FAIL  ", label)

## 安全取记录快照（duplicate 避免引用污染）
func _rec(gid: String = GID) -> Dictionary:
	var r: Variant = DB.get_record(gid)
	return (r as Dictionary).duplicate() if r is Dictionary else {}

func _f() -> void:
	await get_tree().process_frame

func _t(sec: float) -> void:
	await get_tree().create_timer(sec).timeout

# ── 启动 ──────────────────────────────────────────────────────────────────────

func _ready() -> void:
	var scene := load("res://shell/main.tscn") as PackedScene
	_main = scene.instantiate()
	add_child(_main)
	await _f()
	await _f()
	Nav.locate_page_stack()
	# 写入临时 meta（free/paid），重载 Registry
	_write_temp_metas()
	Registry.reload()
	await _f()
	_run()

func _run() -> void:
	await test_cta_free()
	await test_cta_trial()
	await test_cta_trial_exhausted()
	await test_cta_paid()
	await test_cta_owned()
	await test_hint_first_play()
	await test_hint_best()
	await test_continue_row_visible()
	await test_continue_row_empty()
	await test_long_press_remove()
	await test_records_updated_rerender()
	await test_nav_push_pop()
	await test_on_resume_refresh()
	# 清理临时 meta
	_remove_temp_metas()
	Registry.reload()

	if _failures == 0:
		print("[test_detail] ALL PASS (%d 组断言通过)" % 13)
	else:
		push_error("[test_detail] FAILURES: %d" % _failures)
	get_tree().quit(1 if _failures > 0 else 0)

# ── 临时 Meta 管理 ────────────────────────────────────────────────────────────

## 写入临时 free/paid meta.json，供 Registry.reload 扫入
func _write_temp_metas() -> void:
	_ensure_dir("res://games/" + GID_FREE)
	_ensure_dir("res://games/" + GID_PAID)
	var free_meta := {
		"id": GID_FREE, "title": "Free Test Game", "category": "puzzle",
		"tags": ["test"], "price_model": "free", "price": 0, "version": "0.1.0",
		"scene": "res://games/tetra_nova/module.tscn", "runtime": "pck", "core": "",
		"trial": {}, "achievements": []
	}
	var paid_meta := {
		"id": GID_PAID, "title": "Paid Test Game", "category": "puzzle",
		"tags": ["test"], "price_model": "paid", "price": 10, "version": "0.1.0",
		"scene": "res://games/tetra_nova/module.tscn", "runtime": "pck", "core": "",
		"trial": {}, "achievements": []
	}
	_write_json("res://games/%s/meta.json" % GID_FREE, free_meta)
	_write_json("res://games/%s/meta.json" % GID_PAID, paid_meta)

func _ensure_dir(path: String) -> void:
	var dir := DirAccess.open("res://games")
	if dir == null:
		return
	var name := path.get_file()
	if not dir.dir_exists(name):
		dir.make_dir(name)

func _write_json(path: String, data: Dictionary) -> void:
	var f := FileAccess.open(path, FileAccess.WRITE)
	if f == null:
		push_error("test_detail: 无法写入 " + path)
		return
	f.store_string(JSON.stringify(data))
	f.close()

func _remove_temp_metas() -> void:
	for gid in [GID_FREE, GID_PAID]:
		var mp := "res://games/%s/meta.json" % gid
		if FileAccess.file_exists(mp):
			DirAccess.remove_absolute(ProjectSettings.globalize_path(mp))
		var dp := DirAccess.open("res://games")
		if dp != null and dp.dir_exists(gid):
			dp.remove(gid)

# ── 辅助：实例化 CtaBar ────────────────────────────────────────────────────────

func _new_cta_bar(gid: String = GID) -> Node:
	var cta: Node = (load("res://shell/pages/cta_bar.tscn") as PackedScene).instantiate()
	_main.add_child(cta)
	await _f()
	cta.call("setup", gid)
	await _f()
	return cta

func _free_node(n: Node) -> void:
	if n != null and is_instance_valid(n):
		n.get_parent().remove_child(n)
		n.queue_free()

# ── CTA 测试组（五分支）────────────────────────────────────────────────────────

## 分支1: free → btn.text == "▶ 开玩"
func test_cta_free() -> void:
	DB.upsert_record(GID_FREE, {"trial_used": 0, "best": 0, "last_played": 0})
	await _f()
	var cta := await _new_cta_bar(GID_FREE)
	var btn: Button = cta.get_node("VBox/CtaButton") as Button
	_check(btn.text == "▶ 开玩", "test_cta_free: free → btn.text == '▶ 开玩'")
	_free_node(cta)

## 分支2: trial N>0（trial_used=0, plays=3）→ btn.text 含 "3"
func test_cta_trial() -> void:
	DB.upsert_record(GID, {"trial_used": 0, "best": 0})
	await _f()
	var cta := await _new_cta_bar(GID)
	var btn: Button = cta.get_node("VBox/CtaButton") as Button
	_check(btn.text.contains("3"), "test_cta_trial: trial_used=0 left=3 → btn.text 含 '3'")
	_free_node(cta)

## 分支3: trial 耗尽（trial_used=3）→ btn.text 含 "解锁"
func test_cta_trial_exhausted() -> void:
	DB.upsert_record(GID, {"trial_used": 3, "best": 0})
	await _f()
	var cta := await _new_cta_bar(GID)
	var btn: Button = cta.get_node("VBox/CtaButton") as Button
	_check(btn.text.contains("解锁"), "test_cta_trial_exhausted: trial 耗尽 → btn.text 含 '解锁'")
	DB.upsert_record(GID, {"trial_used": 0})
	_free_node(cta)

## 分支4: paid 未拥有（无 paid order）→ btn.text 含 "购买"
func test_cta_paid() -> void:
	# 取消任何遗留的 paid 订单
	DB.put_order({"id": "test_ord_paid_guard", "gid": GID_PAID, "status": "cancelled"})
	DB.upsert_record(GID_PAID, {"trial_used": 0, "best": 0})
	await _f()
	var cta := await _new_cta_bar(GID_PAID)
	var btn: Button = cta.get_node("VBox/CtaButton") as Button
	_check(btn.text.contains("购买"), "test_cta_paid: paid 未拥有 → btn.text 含 '购买'")
	_free_node(cta)

## 分支5: owned（paid + order.status=paid）→ btn.text == "▶ 开玩"
func test_cta_owned() -> void:
	DB.put_order({"id": "test_ord_owned", "gid": GID_PAID, "status": "paid"})
	DB.upsert_record(GID_PAID, {"trial_used": 0, "best": 0})
	await _f()
	var cta := await _new_cta_bar(GID_PAID)
	var btn: Button = cta.get_node("VBox/CtaButton") as Button
	_check(btn.text == "▶ 开玩", "test_cta_owned: paid+order.paid → btn.text == '▶ 开玩'")
	DB.put_order({"id": "test_ord_owned", "gid": GID_PAID, "status": "cancelled"})
	_free_node(cta)

# ── best 字段 hint ────────────────────────────────────────────────────────────

## best=0 → hint == "首次开玩"
func test_hint_first_play() -> void:
	DB.put_order({"id": "test_ord_hint1", "gid": GID_PAID, "status": "paid"})
	DB.upsert_record(GID_PAID, {"trial_used": 0, "best": 0})
	await _f()
	var cta := await _new_cta_bar(GID_PAID)
	var hint: Label = cta.get_node("VBox/HintLabel") as Label
	_check(hint.text == "首次开玩", "test_hint_first_play: best=0 → hint == '首次开玩'")
	DB.put_order({"id": "test_ord_hint1", "gid": GID_PAID, "status": "cancelled"})
	_free_node(cta)

## best=42 → hint 含 "42"
func test_hint_best() -> void:
	DB.put_order({"id": "test_ord_hint2", "gid": GID_PAID, "status": "paid"})
	DB.upsert_record(GID_PAID, {"trial_used": 0, "best": 42})
	await _f()
	var cta := await _new_cta_bar(GID_PAID)
	var hint: Label = cta.get_node("VBox/HintLabel") as Label
	_check(hint.text.contains("42"), "test_hint_best: best=42 → hint 含 '42'")
	DB.upsert_record(GID_PAID, {"best": 0})
	DB.put_order({"id": "test_ord_hint2", "gid": GID_PAID, "status": "cancelled"})
	_free_node(cta)

# ── 继续游戏区块 ──────────────────────────────────────────────────────────────

func _new_section_continue() -> Node:
	var sc: Node = (load("res://shell/components/section_continue.tscn") as PackedScene).instantiate()
	_main.add_child(sc)
	await _f()
	var recommender: Node = get_tree().root.find_child("Recommender", true, false)
	sc.call("set_recommender", recommender)
	sc.call("render")
	await _f()
	return sc

## last_played>0 → visible=true + 卡片数正确
func test_continue_row_visible() -> void:
	DB.upsert_record(GID, {"last_played": 1000, "best": 0, "trial_used": 0})
	await _f()
	var sc := await _new_section_continue()
	_check(sc.visible == true, "test_continue_row_visible: last_played>0 → visible=true")
	var hbox: Node = sc.get_node("ScrollContainer/HBox")
	var recommender: Node = get_tree().root.find_child("Recommender", true, false)
	var rows: Array = recommender.call("continue_row") as Array
	_check(hbox.get_child_count() == rows.size(),
		"test_continue_row_visible: 卡片数 == continue_row 数量 (%d)" % rows.size())
	_free_node(sc)
	DB.upsert_record(GID, {"last_played": 0})

## last_played 全为0 → visible=false
func test_continue_row_empty() -> void:
	DB.upsert_record(GID, {"last_played": 0, "best": 0})
	await _f()
	var sc := await _new_section_continue()
	_check(sc.visible == false, "test_continue_row_empty: last_played 全为0 → visible=false")
	_free_node(sc)

# ── 长按移除（权益保护）────────────────────────────────────────────────────────

## DB.upsert_record(last_played=0) 只清 last_played，best/trial_used/finish_count 不变
func test_long_press_remove() -> void:
	DB.upsert_record(GID, {"last_played": 500, "best": 99, "trial_used": 1, "finish_count": 2})
	await _f()
	var snap_before := _rec()
	DB.upsert_record(GID, {"last_played": 0})
	await _f()
	var snap_after := _rec()
	_check(int(snap_after.get("last_played", -1)) == 0,
		"test_long_press_remove: last_played == 0")
	_check(int(snap_after.get("best", 0)) == int(snap_before.get("best", 0)),
		"test_long_press_remove: best 不变（%d）" % int(snap_before.get("best", 0)))
	_check(int(snap_after.get("trial_used", 0)) == int(snap_before.get("trial_used", 0)),
		"test_long_press_remove: trial_used 不变")
	_check(int(snap_after.get("finish_count", 0)) == int(snap_before.get("finish_count", 0)),
		"test_long_press_remove: finish_count 不变")
	DB.upsert_record(GID, {"best": 0, "trial_used": 0, "finish_count": 0, "last_played": 0})

# ── records_updated 信号重渲染 ────────────────────────────────────────────────

## records_updated emit(gid) → SectionContinue.render() 重建卡片数变化
func test_records_updated_rerender() -> void:
	DB.upsert_record(GID, {"last_played": 0, "best": 0})
	await _f()
	var sc := await _new_section_continue()
	var hbox: Node = sc.get_node("ScrollContainer/HBox")
	var count_before: int = hbox.get_child_count()
	# DB.upsert_record 内部发 records_updated → SectionContinue._on_records_updated → render()
	DB.upsert_record(GID, {"last_played": 999})
	await _f()
	await _f()
	var count_after: int = hbox.get_child_count()
	_check(count_after > count_before,
		"test_records_updated_rerender: records_updated → 卡片数增加（%d→%d）" % [count_before, count_after])
	_check(sc.visible == true,
		"test_records_updated_rerender: 信号后 visible=true")
	_free_node(sc)
	DB.upsert_record(GID, {"last_played": 0})

# ── Nav push/pop ──────────────────────────────────────────────────────────────

## Nav.push(detail, {gid}) → 栈 +1；on_enter _gid 正确；Nav.pop → 栈恢复
func test_nav_push_pop() -> void:
	Nav.locate_page_stack()
	# 覆盖 tab_roots 为占位页，避免依赖真实首页场景
	Nav._tab_roots = {0: "res://tools/dummy_page_a.tscn"}
	Nav.switch_tab(0)
	await _t(0.4)
	var size_before: int = Nav._pages.size()
	Nav.push("res://shell/pages/detail.tscn", {"gid": GID})
	await _t(0.4)
	var size_after: int = Nav._pages.size()
	_check(size_after == size_before + 1,
		"test_nav_push_pop: push → 栈 +1（%d→%d）" % [size_before, size_after])
	if Nav._pages.size() > 0:
		var top: Node = Nav._pages.back()
		# _gid 是私有成员，通过 get() 访问
		var gid_val: Variant = top.get("_gid")
		_check(str(gid_val) == GID,
			"test_nav_push_pop: on_enter _gid == 'tetra_nova'（实际='%s'）" % str(gid_val))
	Nav.pop()
	await _t(0.4)
	_check(Nav._pages.size() == size_before,
		"test_nav_push_pop: pop → 栈恢复（%d）" % size_before)

# ── on_resume 刷新 ────────────────────────────────────────────────────────────

## trial_consumed emit 后 on_resume → CTA 文案更新（trial_used=1 left=2 → 含 "2"）
func test_on_resume_refresh() -> void:
	DB.upsert_record(GID, {"trial_used": 0, "best": 0})
	# 确保 tetra_nova 不被 owned（paid 分支会走 open_play，忽略 trial_used）
	DB.put_order({"id": "test_ord_resume_guard", "gid": GID, "status": "cancelled"})
	await _f()
	var ps: Node = get_tree().root.find_child("PageStack", true, false)
	if ps == null:
		ps = _main
	var detail: Node = (load("res://shell/pages/detail.tscn") as PackedScene).instantiate()
	ps.add_child(detail)
	await _f()
	detail.call("on_enter", {"gid": GID})
	await _f()
	var cta_bar: Node = detail.get_node("CtaBar")
	var btn: Button = cta_bar.get_node("VBox/CtaButton") as Button
	# trial_used=0 时文案含 "3"
	_check(btn.text.contains("3"),
		"test_on_resume_refresh: trial_used=0 → btn 含 '3'（实际='%s'）" % btn.text)
	# 消耗一次后 trial_consumed emit
	DB.upsert_record(GID, {"trial_used": 1})
	EventBus.trial_consumed.emit(GID, 2)
	await _f()
	# CtaBar._on_trial_consumed → refresh()，按钮文案应含 "2"
	_check(btn.text.contains("2"),
		"test_on_resume_refresh: trial_consumed emit → CTA refresh 含 '2'（实际='%s'）" % btn.text)
	# 再消耗一次（left=1）后调 on_resume
	DB.upsert_record(GID, {"trial_used": 2})
	EventBus.trial_consumed.emit(GID, 1)
	await _f()
	detail.call("on_resume")
	await _f()
	_check(btn.text.contains("1"),
		"test_on_resume_refresh: on_resume 后 CTA 含 '1'（实际='%s'）" % btn.text)
	_free_node(detail)
	DB.upsert_record(GID, {"trial_used": 0})
