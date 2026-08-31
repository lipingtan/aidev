extends Node
## LauncherService（Autoload 名 Launcher，顺序第 8；CR-2 T4）：状态机 idle→launching→running→quitting（design §3）
## launch：权益检查 → sessions+1 → 转场（全不透明后切视口）→ boot(ctx) → game_launched → 淡入
## quit：冻结（保留节点供 play_again 复用）→ 全不透明后恢复默认视口 → 强写三字段 → 400ms 后结算卡
## play_again：不重建模块（reset_run），仅 session+1，不耗 trial；launching 失败 → 回 idle + Toast
## Mock 桩：PurchaseDialog（M4）；arcade 走 CoreManager 桩路径（M2）。转场细节见 LauncherTransition
## 依赖：Registry/GameMeta、DB、TrialGuard、EventBus、GameModule 基类

enum State { IDLE, LAUNCHING, RUNNING, QUITTING }

## 退出后延迟显示结算卡（秒；design §3：400ms）
const RESULT_DELAY_S := 0.4

var state: State = State.IDLE
var current_gid: String = ""
var _module: GameModule = null
var _tr := LauncherTransition.new()
var _u := LauncherUtil.new()
var _cards_layer: Node = null
var _result: Control = null
var _purchase: Control = null

## autoload 就绪期 root 正忙于装配子节点 → 全部延迟到下一帧挂载
func _ready() -> void:
	add_child.call_deferred(_tr)
	add_child.call_deferred(_u)
	_tr.attach.call_deferred()

## 挂载结算卡/购买弹窗到 OverlayLayer（懒加载：autoload 先于主场景就绪）
func _ensure_cards() -> void:
	if _result != null and is_instance_valid(_result):
		return
	_cards_layer = get_tree().root.find_child("OverlayLayer", true, false) as Control
	if _cards_layer == null:
		_cards_layer = get_tree().root
	var rs := load("res://shell/components/result_overlay.tscn") as PackedScene
	_result = rs.instantiate()
	_cards_layer.add_child(_result)
	var ps := load("res://shell/components/purchase_dialog.tscn") as PackedScene
	_purchase = ps.instantiate()
	_cards_layer.add_child(_purchase)

## 当前游戏模块实例（只读访问器，供测试/外部使用）
func get_module() -> GameModule:
	return _module

## 最近一次视口切换是否发生在遮罩全不透明后（时序钩子，T8 断言用）
func vp_switched_at_full_mask() -> bool:
	return _tr.vp_at_full_mask

# === 启动 ===

## 启动游戏（公开入口；design §3 完整序列）
func launch(gid: String) -> void:
	if state != State.IDLE:
		_u.toast("当前操作繁忙，请稍候")
		return
	var meta: GameMeta = Registry.lookup(gid)
	if meta == null:
		_u.toast("游戏不存在：%s" % gid)
		return
	if not _check_entitlement(meta):
		return
	state = State.LAUNCHING
	current_gid = gid
	_ensure_cards()
	_u.bump_session(gid)
	await _enter(meta)

## 权益检查：free 放行；trial 余量>0 消耗一次；其余未拥有 → Mock 购买弹窗（M4）
func _check_entitlement(meta: GameMeta) -> bool:
	match meta.price_model:
		"free":
			return true
		"trial":
			if TrialGuard.left(meta.id) <= 0:
				_show_purchase(meta)
				return false
			TrialGuard.consume(meta.id)
			return true
		_:
			_show_purchase(meta)
			return false

## Mock 购买弹窗（M4 接入真实支付前仅提示）
func _show_purchase(meta: GameMeta) -> void:
	if is_instance_valid(_purchase):
		(_purchase as Node).call("show_mock", meta.title, meta.price)
	_u.toast("%s ¥%d 解锁完整版（M4 接入支付）" % [meta.title, meta.price])

## 入场序列：转场切视口 → boot(ctx) → game_launched → 淡入
func _enter(meta: GameMeta) -> void:
	await _tr.to_game(meta)
	var host := _u.host()
	if host == null:
		push_error("Launcher: 未找到 GameHost 节点")
		_abort()
		return
	_u.set_game_ui(true)
	var module := _instantiate_module(meta)
	if module == null:
		_abort()
		return
	_module = module
	host.add_child(module)
	module.boot({
		"save_dir": "user://saves/%s/" % meta.id,
		"trial_mode": meta.price_model == "trial",
		"owned": false,
		"best": _u.int_field(meta.id, "best"),
		"viewport_size": meta.get_viewport_size(),
	})
	module.quit_requested.connect(_on_quit_requested.bind(meta.id))
	state = State.RUNNING
	EventBus.game_launched.emit(meta.id)
	await _tr.reveal()

## 实例化游戏模块；arcade 走 CoreManager 桩路径（M2，本 CR 不实际执行）
func _instantiate_module(meta: GameMeta) -> GameModule:
	if meta.runtime == "arcade":
		push_warning("Launcher: arcade runtime 走 CoreManager 桩（core=%s）" % meta.core)
		return null
	var res := load(meta.scene) as PackedScene
	if res == null:
		push_error("Launcher: 场景加载失败 %s" % meta.scene)
		return null
	return res.instantiate() as GameModule

## launching 失败回滚：恢复视口 + 隐藏 GameHost + 回 idle + Toast
func _abort() -> void:
	await _tr.to_shell()
	_u.set_game_ui(false)
	state = State.IDLE
	current_gid = ""
	_u.toast("启动失败，请重试")

# === 退出 / 再来一局 ===

## 游戏请求回盒：冻结 → 转场回 Shell → 强写三字段 → game_finished → 结算卡
func _on_quit_requested(result: Dictionary, gid: String) -> void:
	if state != State.RUNNING:
		return
	state = State.QUITTING
	if is_instance_valid(_module):
		_module.process_mode = Node.PROCESS_MODE_DISABLED
		_module.set_module_visible(false)   # CanvasLayer 不随 GameHost 隐藏，须显式处理
	await _tr.to_shell()
	_u.set_game_ui(false)
	_u.finish_record(gid, result)
	EventBus.game_finished.emit(gid, result)
	await _tr.reveal()
	state = State.IDLE
	await get_tree().create_timer(RESULT_DELAY_S).timeout
	if is_instance_valid(_result):
		(_result as Node).call("show_card", gid, result)

## 再来一局：不重建模块（reset_run 复用），仅 session+1，不耗 trial
func play_again(gid: String) -> void:
	if state != State.IDLE or not is_instance_valid(_module):
		return
	var meta: GameMeta = Registry.lookup(gid)
	if meta == null:
		return
	state = State.LAUNCHING
	current_gid = gid
	_u.bump_session(gid)
	# 复用入场：reset_run → 恢复可见+解冻 → 转场切视口（不重走权益/不 boot）
	(_module as Node).call("reset_run")
	_module.set_module_visible(true)
	_module.process_mode = Node.PROCESS_MODE_INHERIT
	await _tr.to_game(meta)
	_u.set_game_ui(true)
	state = State.RUNNING
	EventBus.game_launched.emit(meta.id)
	await _tr.reveal()

## 释放游戏实例（结算卡「返回」/「换一个」调用）：free 模块 + 恢复视口
## QUITTING 门禁（2026-08-31）：转场期间 state=IDLE 裸奔，launch 可并发进入与 to_shell 竞态
func release() -> void:
	if state == State.QUITTING or state == State.LAUNCHING:
		return
	state = State.QUITTING
	if is_instance_valid(_module):
		_module.set_module_visible(false)   # free 前先摘掉 CanvasLayer，防释放帧闪现
		_module.queue_free()
		_module = null
	await _tr.to_shell()
	_u.set_game_ui(false)
	await _tr.reveal()   # 缺此步则 to_shell 的黑罩（alpha=1）永驻屏幕顶层 → 全屏黑 + 点击全吞
	state = State.IDLE

## Shell 暂停键转发到当前模块
func pause_game() -> void:
	if state == State.RUNNING and is_instance_valid(_module):
		_module.pause_game()

## Shell 恢复键转发到当前模块
func resume_game() -> void:
	if state == State.RUNNING and is_instance_valid(_module):
		_module.resume_game()

