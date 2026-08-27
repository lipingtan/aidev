extends Node
## T8 launch/quit 集成测试：实例化真实 shell/main.tscn，验证 design §3 完整时序
## 1. launch：权益→DB sessions+1→转场（vp_at_full_mask）→boot(ctx 完整)→RUNNING
## 2. quit：冻结→恢复默认视口→强写三字段→game_finished→结算卡显示
## 3. play_again：同实例复用（不重建/不耗 trial），仅 session+1
## 4. trial 耗尽 → PurchaseDialog Mock + 中止（state IDLE + 视口复原）
## 5. 坏 meta scene → 回 idle + Toast + 视口复原

const GID := "tetra_nova"
const SHELL_VP := Vector2i(720, 1560)

var _failures: int = 0
var _main: Node
var _module: GameModule

func _ready() -> void:
	var scene := load("res://shell/main.tscn") as PackedScene
	_main = scene.instantiate()
	add_child(_main)
	await get_tree().process_frame
	_run()

func _check(cond: bool, label: String) -> void:
	if cond:
		print("[test_launch] PASS ", label)
	else:
		_failures += 1
		push_error("[test_launch] FAIL ", label)

## Shell App 是否可见（RUNNING 应隐藏，退出后恢复）
func _app_visible() -> bool:
	var app := get_tree().root.find_child("App", true, false) as Control
	return app != null and app.visible

## 取记录快照（DB 返回内存字典引用，后续 upsert 会原地更新 → 必须 duplicate）
func _rec() -> Dictionary:
	var r: Variant = DB.get_record(GID)
	return (r as Dictionary).duplicate() if r is Dictionary else {}

func _run() -> void:
	await _t(0.4)
	var base := _rec()
	# ---- 1. launch：权益消耗 + sessions+1 + ctx 完整 + 遮满后切视口 ----
	# 等待 ≥2s：headless 下 boot→quit 需跨满整秒，playtime（int 秒）才 >0
	Launcher.launch(GID)
	await _t(2.0)
	_check(Launcher.state == Launcher.State.RUNNING, "launch → RUNNING")
	_module = Launcher.get_module()
	_check(_module != null, "模块实例化成功（GameModule）")
	if _module != null:
		var ctx := _module.ctx
		_check(ctx.has("save_dir") and ctx.has("trial_mode") and ctx.has("owned"), "ctx 基础键完整")
		_check(ctx.has("best") and ctx.has("viewport_size"), "ctx best/viewport_size 完整")
		_check(str(ctx.get("save_dir", "")) == "user://saves/%s/" % GID, "ctx.save_dir 规范")
	_check(Launcher.vp_switched_at_full_mask(), "视口切换发生在遮罩全不透明后")
	_check(not _app_visible(), "运行中 Shell App 隐藏（防 UI 叠加）")
	var rec1 := _rec()
	_check(int(rec1.get("sessions", 0)) == int(base.get("sessions", 0)) + 1, "sessions+1（读现值）")
	_check(int(rec1.get("trial_used", 0)) == int(base.get("trial_used", 0)) + 1, "trial 消耗一次")
	# ---- 2. quit：冻结 → 恢复视口 → 强写三字段 → 结算卡 ----
	var host := get_tree().root.find_child("GameHost", true, false) as Control
	if _module != null:
		_module.quit_to_shell()
	await _t(1.8)
	_check(Launcher.state == Launcher.State.IDLE, "quit → IDLE")
	_check(host != null and not host.visible, "GameHost 隐藏")
	_check(_app_visible(), "退出后 Shell App 恢复")
	_check(get_tree().root.content_scale_size == SHELL_VP, "视口恢复 Shell 默认")
	var rec2 := _rec()
	_check(int(rec2.get("finish_count", 0)) == int(rec1.get("finish_count", 0)) + 1, "finish_count+1（读现值）")
	_check(float(rec2.get("total_playtime", 0.0)) > float(rec1.get("total_playtime", 0.0)), "total_playtime 累加")
	var result_ov := get_tree().root.find_child("ResultOverlay", true, false) as Control
	var card_layer_ok := false
	if result_ov != null:
		var p := result_ov.get_parent()
		card_layer_ok = p is Control and (p as Control).visible
	_check(result_ov != null and result_ov.visible and card_layer_ok, "结算卡显示（含父层 OverlayLayer 打开）")
	if result_ov != null:
		result_ov.visible = false
		var p := result_ov.get_parent()
		if p is Control and (p as Control).name != "root":
			(p as Control).visible = false  # 同步隐藏 OverlayLayer，防 Dim 遮罩残留
	# ---- 3. play_again：同实例复用，不耗 trial ----
	var m1 := Launcher.get_module()
	Launcher.play_again(GID)
	await _t(1.0)
	_check(Launcher.state == Launcher.State.RUNNING, "play_again → RUNNING")
	_check(Launcher.get_module() == m1, "play_again 复用同实例（不重建）")
	var rec3 := _rec()
	_check(int(rec3.get("trial_used", 0)) == int(rec2.get("trial_used", 0)), "play_again 不耗 trial")
	_check(int(rec3.get("sessions", 0)) == int(rec2.get("sessions", 0)) + 1, "play_again session+1")
	if _module != null:
		_module.quit_to_shell()
	await _t(1.8)
	# ---- 4. trial 耗尽 → PurchaseDialog Mock + 中止 ----
	DB.upsert_record(GID, {"trial_used": 3})
	Launcher.launch(GID)
	await _t(0.6)
	var purchase := get_tree().root.find_child("PurchaseDialog", true, false) as Control
	_check(purchase != null and purchase.visible, "耗尽 → PurchaseDialog Mock")
	_check(Launcher.state == Launcher.State.IDLE, "耗尽中止 → IDLE")
	_check(get_tree().root.content_scale_size == SHELL_VP, "耗尽中止视口复原")
	if purchase != null:
		purchase.visible = false
	DB.upsert_record(GID, {"trial_used": 0})
	# ---- 5. 坏 meta scene → 回 idle + Toast + 视口复原（headless 下 res:// 只读则 SKIP）----
	if _make_broken_meta():
		Registry.reload()
		Launcher.launch("_test_broken")
		await _t(1.0)
		_check(Launcher.state == Launcher.State.IDLE, "坏场景 → IDLE（回退）")
		_check(get_tree().root.content_scale_size == SHELL_VP, "坏场景视口复原")
		_kill_broken_meta()
	else:
		print("[test_launch] SKIP 坏场景步骤（headless res:// 只读）")
	Registry.reload()

	if _failures == 0:
		print("[test_launch] ALL PASS")
	get_tree().quit(1 if _failures > 0 else 0)

## 等待 t 秒（转场 fade 300ms×2 + 结算延迟 400ms 的缓冲）
func _t(sec: float) -> void:
	await get_tree().create_timer(sec).timeout

## 造坏 meta（free + 不存在的 scene），验证 launching 失败回滚；res:// 只读时返回 false
func _make_broken_meta() -> bool:
	var dir := DirAccess.open("res://games")
	if dir == null or dir.dir_exists("_test_broken"):
		return false
	dir.make_dir("_test_broken")
	var meta := {
		"id": "_test_broken", "title": "坏场景测试", "category": "puzzle",
		"tags": ["test"], "price_model": "free", "price": 0, "version": "0.0.1",
		"scene": "res://games/_test_broken/nope.tscn", "runtime": "pck", "core": "",
	}
	var f := FileAccess.open("res://games/_test_broken/meta.json", FileAccess.WRITE)
	if f == null:
		return false
	f.store_string(JSON.stringify(meta))
	f.close()
	return true

## 清理坏 meta
func _kill_broken_meta() -> void:
	var d := DirAccess.open("res://games/_test_broken")
	if d != null:
		d.remove("meta.json")
	var parent := DirAccess.open("res://games")
	if parent != null and parent.dir_exists("_test_broken"):
		parent.remove("_test_broken")  # 4.5：remove() 可删空目录（无 rmdir）
