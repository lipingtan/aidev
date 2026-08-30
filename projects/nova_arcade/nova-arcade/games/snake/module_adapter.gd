extends GameModule
## 贪吃蛇 协议适配器（CR-6 T3；tetra 模式）
## - boot(ctx)：注入 SAVE_PATH = ctx.save_dir + "save.cfg" → src
## - src on_game_over(stats) → quit_requested({score, playtime, achievements:[], extra:{length}})
## - pause/resume 转发 src.set_paused()（R-11 只约定接口）
## - reset_run()：重开一局（Launcher.play_again 复用模块不重建）
## - adapter 侧追加「回菜单」按钮（D5=A 唯一 active exit；src zero diff）
## 依赖：仅 GameModule base class + src nodes; no reference to any Shell type

var _src: Node = null
var _run_start_ms: int = 0
var _last_stats: Dictionary = {}
var _menu_btn: Button = null

func _ready() -> void:
	# 子节点 _ready 先于本节点：此处 src 已完成组装
	_src = $Src as Node
	_src.on_game_over.connect(_on_game_over)
	_install_menu_button()

## 启动：注入存档路径 + 开局计时（play_again reuse only needs reset_run）
func boot(ctx_in: Dictionary) -> void:
	ctx = ctx_in
	var path := str(ctx.get("save_dir", "user://")) + "save.cfg"
	_src.SAVE_PATH = path
	# 不在此调 start_game：等玩家点开始遮罩的「开始游戏」
	_run_start_ms = Time.get_ticks_msec()

## 回菜单（D5=A unique active exit）：emit quit_requested（本局结果）
func quit_to_shell() -> void:
	quit_requested.emit({
		"score": int(_last_stats.get("score", 0)),
		"playtime": _run_seconds(),
		"achievements": [],
		"extra": {"length": int(_last_stats.get("length", 0))},
	})

## 暂停（forward to src.set_paused）
func pause_game() -> void:
	if _src != null:
		_src.set_paused(true)

## Resume（forward to src.set_paused）
func resume_game() -> void:
	if _src != null:
		_src.set_paused(false)

## Restart run (Launcher.play_again reuses module without rebuilding)
func reset_run() -> void:
	if _src != null:
		_src.restart()
	_last_stats = {}
	_run_start_ms = Time.get_ticks_msec()

## 本局已玩秒数
func _run_seconds() -> int:
	return (Time.get_ticks_msec() - _run_start_ms) / 1000

## 游戏结束：record stats（Shell 统一结算卡，adapter 不展示）
func _on_game_over(stats: Dictionary) -> void:
	_last_stats = stats

## Append "回菜单" button on adapter side (top-right HUD area; src zero diff)
func _install_menu_button() -> void:
	if _menu_btn != null or _src == null:
		return
	var layer := CanvasLayer.new()
	add_child(layer)
	_menu_btn = Button.new()
	_menu_btn.focus_mode = Control.FOCUS_NONE   # 防抢键盘焦点（方向键/空格被吞）
	_menu_btn.text = "↩ 回菜单"
	_menu_btn.add_theme_font_size_override("font_size", 20)
	_menu_btn.custom_minimum_size = Vector2(150, 48)
	_menu_btn.position = Vector2(720 - 160, 10)
	_menu_btn.pressed.connect(quit_to_shell)
	layer.add_child(_menu_btn)
