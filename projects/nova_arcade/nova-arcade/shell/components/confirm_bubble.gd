extends Control
## ConfirmBubble 轻量确认弹窗（CR-3 §4）
## 挂 OverlayLayer 下；默认 visible=false
## show_bubble(msg, on_confirm) → 弹出；[确认]/[取消] → 关闭

signal confirmed
signal cancelled

@onready var _panel: PanelContainer = $Panel as PanelContainer
@onready var _label: Label = $Panel/VBox/Label as Label
@onready var _confirm_btn: Button = $Panel/VBox/HBox/ConfirmBtn as Button
@onready var _cancel_btn: Button = $Panel/VBox/HBox/CancelBtn as Button
@onready var _dim: ColorRect = $Dim as ColorRect

var _on_confirm: Callable

func _ready() -> void:
	_confirm_btn.pressed.connect(_do_confirm)
	_cancel_btn.pressed.connect(_do_cancel)
	_apply_style()
	visible = false

## 显示确认弹窗
func show_bubble(msg: String, on_confirm: Callable) -> void:
	_label.text = msg
	_on_confirm = on_confirm
	_apply_style()
	visible = true

## [确认] 处理
func _do_confirm() -> void:
	visible = false
	confirmed.emit()
	if _on_confirm.is_valid():
		_on_confirm.call()

## [取消] 处理
func _do_cancel() -> void:
	visible = false
	cancelled.emit()

## 应用主题色（全用 ThemeTokens.color）
func _apply_style() -> void:
	# 半透明遮罩
	_dim.color = ThemeTokens.color("ov_bg")
	# 面板背景
	var sb := StyleBoxFlat.new()
	sb.bg_color = ThemeTokens.color("card2")
	sb.border_color = ThemeTokens.color("line")
	sb.set_border_width_all(1)
	sb.set_corner_radius_all(12)
	_panel.add_theme_stylebox_override("panel", sb)
	# 文本颜色
	_label.add_theme_color_override("font_color", ThemeTokens.color("ink"))
	# 确认按钮
	var sb_confirm := StyleBoxFlat.new()
	sb_confirm.bg_color = ThemeTokens.color("accent")
	sb_confirm.set_corner_radius_all(8)
	_confirm_btn.add_theme_stylebox_override("normal", sb_confirm)
	_confirm_btn.add_theme_color_override("font_color", ThemeTokens.color("play_ink"))
	# 取消按钮
	var sb_cancel := StyleBoxFlat.new()
	sb_cancel.bg_color = ThemeTokens.color("line2")
	sb_cancel.set_corner_radius_all(8)
	_cancel_btn.add_theme_stylebox_override("normal", sb_cancel)
	_cancel_btn.add_theme_color_override("font_color", ThemeTokens.color("ink2"))
