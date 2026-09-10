extends GameModule
## 幸运轮盘 协议适配器（CR-8；snake 模式）
## - boot(ctx)：注入 SAVE_PATH → src；显示开始遮罩（src 自建）；不自动开局
## - src quit_stats() → quit_requested({score, playtime, achievements, extra})
## - pause/resume → src.set_paused()；reset_run()：余额/streak 保留，清当局押注
## - adapter 侧常驻「回菜单」按钮（CanvasLayer layer=20，FOCUS_NONE）
## 依赖：仅 GameModule base class + src nodes; no reference to any Shell type

var _src: Node = null
var _run_start_ms: int = 0
var _menu_btn: Button = null

func _ready() -> void:
	_src = $Src as Node
	_install_menu_button()

func boot(ctx_in: Dictionary) -> void:
	ctx = ctx_in
	_src.SAVE_PATH = str(ctx.get("save_dir", "user://")) + "save.cfg"
	_src.load_balance()
	_run_start_ms = Time.get_ticks_msec()

## 回菜单：组装会话统计 emit quit_requested（Shell 统一结算卡）
func quit_to_shell() -> void:
	var stats: Dictionary = _src.quit_stats()
	stats["playtime"] = (Time.get_ticks_msec() - _run_start_ms) / 1000.0
	quit_requested.emit(stats)

func pause_game() -> void:
	_src.set_paused(true)

func resume_game() -> void:
	_src.set_paused(false)

## 重开一局：余额/streak 保留（会话语义），清当局押注
func reset_run() -> void:
	_src.reset_run()
	_run_start_ms = Time.get_ticks_msec()

## 常驻回菜单按钮（右上角，layer=20 高于 ui(10)）
func _install_menu_button() -> void:
	if _menu_btn != null:
		return
	var layer := CanvasLayer.new()
	layer.layer = 20
	add_child(layer)
	_menu_btn = Button.new()
	_menu_btn.focus_mode = Control.FOCUS_NONE
	_menu_btn.text = "↩ 回菜单"
	_menu_btn.add_theme_font_size_override("font_size", 20)
	_menu_btn.custom_minimum_size = Vector2(150, 48)
	_menu_btn.position = Vector2(720 - 160, 10)
	_menu_btn.pressed.connect(quit_to_shell)
	layer.add_child(_menu_btn)
