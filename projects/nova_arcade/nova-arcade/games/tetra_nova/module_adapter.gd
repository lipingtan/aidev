extends GameModule
## TETRA NOVA 协议适配器（CR-2 T3；D1=A 最小 diff / D5=A「回菜单」为唯一主动退出）
## - boot(ctx)：注入 SAVE_PATH（ctx.save_dir+"save.cfg"）到 main/ui + 开局计时
## - src game.run_over(stats) → 记录本局统计；结算屏「回菜单」按钮（adapter 侧追加，src 零 diff）
##   → quit_requested({score, playtime, achievements:[], extra:{wave,lines,max_combo,bosses}})
## - pause/resume 转发 src PAUSED 状态机（input_action("pause") 切换）
## - reset_run()：重开一局（Launcher.play_again 复用模块不重建）
##
## 依赖：仅 GameModule 基类 + src 节点；不引用 Shell 任何类型

var _main: Node = null
var _game: Node = null
var _ui: CanvasLayer = null
var _run_start_ms: int = 0
var _last_stats: Dictionary = {}
var _menu_btn: Button = null

func _ready() -> void:
	# 子节点 _ready 先于本节点：此处 main 已完成组装（game/ui 就绪）
	_main = $Main as Node
	_game = _main.game
	_ui = _main.ui
	_game.run_over.connect(_on_run_over)
	_install_menu_button()

## 启动：注入存档路径 + 开局计时（play_again 复用时无需重走，仅 reset_run）
func boot(ctx_in: Dictionary) -> void:
	ctx = ctx_in
	var path := str(ctx.get("save_dir", "user://")) + "save.cfg"
	_main.SAVE_PATH = path
	_ui.SAVE_PATH = path
	_set_ui_visible(true)
	_run_start_ms = Time.get_ticks_msec()

## 回菜单（D5=A 唯一主动退出）：隐藏 UI 层 + 发 quit_requested（本局结果）
func quit_to_shell() -> void:
	_set_ui_visible(false)
	quit_requested.emit({
		"score": int(_last_stats.get("score", 0)),
		"playtime": _run_seconds(),
		"achievements": [],
		"extra": {
			"wave": int(_last_stats.get("wave", 1)),
			"lines": int(_last_stats.get("lines", 0)),
			"max_combo": int(_last_stats.get("max_combo", 0)),
			"bosses": int(_last_stats.get("bosses", 0)),
		},
	})

## 暂停（转发 src PAUSED 状态机）
func pause_game() -> void:
	if _game != null and str(_game.state) == "PLAYING":
		_game.input_action("pause")

## 恢复（转发 src 恢复机制）
func resume_game() -> void:
	if _game != null and str(_game.state) == "PAUSED":
		_game.input_action("pause")

## 重开一局（Launcher.play_again 复用模块不重建）
func reset_run() -> void:
	_set_ui_visible(true)
	if _game != null:
		_game.start_game()
		_run_start_ms = Time.get_ticks_msec()

## UI 层显隐：ui 是 CanvasLayer，不随祖先 Control.visible 隐藏，须逐子节点管理
func _set_ui_visible(v: bool) -> void:
	if _ui == null:
		return
	for ch in _ui.get_children():
		ch.visible = v

## 本局已玩秒数
func _run_seconds() -> int:
	return (Time.get_ticks_msec() - _run_start_ms) / 1000

## 一局结束：记录统计（结算屏由 src 自身展示）
func _on_run_over(stats: Dictionary) -> void:
	_last_stats = stats

## 结算屏追加「回菜单」按钮（src OVER overlay box；只读访问 src 节点，src 零 diff）
func _install_menu_button() -> void:
	if _menu_btn != null or _ui == null:
		return
	var over: Control = _ui._over
	var box := over.get_meta("box") as VBoxContainer
	if box == null:
		return
	_menu_btn = Button.new()
	_menu_btn.text = "↩ 回菜单"
	_menu_btn.add_theme_font_size_override("font_size", 20)
	_menu_btn.custom_minimum_size = Vector2(240, 56)
	_menu_btn.pressed.connect(quit_to_shell)
	box.add_child(_menu_btn)
