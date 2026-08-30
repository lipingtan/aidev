class_name GameCard extends PanelContainer
## GameCard 卡片组件（CR-3 §3）：通用游戏卡片，供继续游戏/为你推荐区块复用
## 长按计时器用 _process 实现，规避 ScrollContainer 触摸滑动吞键（Risk-4）

signal card_pressed(gid: String)
signal card_long_pressed(gid: String)

## 长按阈值（毫秒）
const LONG_PRESS_MS: int = 500

@onready var _icon_rect: TextureRect = $VBox/IconRect as TextureRect
@onready var _name_label: Label = $VBox/NameLabel as Label
@onready var _status_badge: Label = $VBox/StatusBadge as Label

var _gid: String = ""
var _press_start: int = -1
var _long_fired: bool = false

## 初始化卡片数据；size: "card"（默认横滑卡）/ "grid"（分类页 2 列网格）
func setup(meta: GameMeta, record: Variant, size: String = "card") -> void:
	_gid = meta.id
	_name_label.text = meta.title
	_apply_badge(meta, record)
	_apply_style()
	if size == "grid":
		size_flags_horizontal = Control.SIZE_EXPAND_FILL
		custom_minimum_size = Vector2(0, 140)

## 根据 record 设置状态角标
func _apply_badge(meta: GameMeta, record: Variant) -> void:
	var r := record as Dictionary if record is Dictionary else {}
	var best: int = int(r.get("best", 0))
	if best > 0:
		_status_badge.text = "最高 %d" % best
		_status_badge.visible = true
	elif meta.price_model == "trial":
		var left: int = 0
		if TrialGuard != null:
			left = TrialGuard.left(meta.id)
		if left > 0:
			_status_badge.text = "试玩 %d 次" % left
		else:
			_status_badge.text = "试玩结束"
		_status_badge.visible = true
	else:
		_status_badge.visible = false

## 应用主题色（全用 ThemeTokens.color，禁止硬编码）
func _apply_style() -> void:
	# 卡片背景
	var sb := StyleBoxFlat.new()
	sb.bg_color = ThemeTokens.color("card")
	sb.border_color = ThemeTokens.color("line")
	sb.set_border_width_all(1)
	sb.set_corner_radius_all(10)
	add_theme_stylebox_override("panel", sb)
	# 标题颜色
	_name_label.add_theme_color_override("font_color", ThemeTokens.color("ink"))
	# 角标颜色
	_status_badge.add_theme_color_override("font_color", ThemeTokens.color("ink2"))

## 触摸/鼠标按下开始计时
func _gui_input(event: InputEvent) -> void:
	if event is InputEventMouseButton:
		var e := event as InputEventMouseButton
		if e.button_index == MOUSE_BUTTON_LEFT:
			if e.pressed:
				_press_start = Time.get_ticks_msec()
				_long_fired = false
			else:
				# 抬起时先判 _long_fired：未触发长按 → emit card_pressed
				if not _long_fired and _press_start >= 0:
					var dt := Time.get_ticks_msec() - _press_start
					if dt < LONG_PRESS_MS:
						card_pressed.emit(_gid)
				_press_start = -1

## _process 中轮询长按计时，不依赖 ScrollContainer 事件
func _process(_dt: float) -> void:
	if _press_start >= 0 and not _long_fired:
		if Time.get_ticks_msec() - _press_start >= LONG_PRESS_MS:
			_long_fired = true
			card_long_pressed.emit(_gid)
			_press_start = -1
