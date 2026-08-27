extends Control
## 购买弹窗 Mock（CR-2 T6 / M4）：标题 + ¥价格 + [暂不购买]
## M4 接入真实支付前仅提示；Launcher 权益检查失败时调用 show_mock

@onready var _dim: ColorRect = $Dim as ColorRect
@onready var _card: PanelContainer = $Card as PanelContainer
@onready var _title: Label = $Card/VBox/Title as Label
@onready var _price: Label = $Card/VBox/Price as Label
@onready var _btn_later: Button = $Card/VBox/BtnLater as Button

## 显示 Mock 购买弹窗
func show_mock(title: String, price: int) -> void:
	_apply_theme()
	_title.text = title
	_price.text = "¥%d 解锁完整版" % price
	_btn_later.pressed.connect(_on_later)
	visible = true

## [暂不购买]：关闭弹窗（Launcher 已 Toast 提示）
func _on_later() -> void:
	visible = false

## 主题色（仅经 ThemeTokens.color，override §2）
func _apply_theme() -> void:
	_dim.color = ThemeTokens.color("ov_bg")
	var sb := StyleBoxFlat.new()
	sb.bg_color = ThemeTokens.color("card2")
	sb.border_color = ThemeTokens.color("line")
	sb.set_border_width_all(1)
	sb.set_corner_radius_all(12)
	_card.add_theme_stylebox_override("panel", sb)
	_title.add_theme_color_override("font_color", ThemeTokens.color("ink"))
	_price.add_theme_color_override("font_color", ThemeTokens.color("gold"))
