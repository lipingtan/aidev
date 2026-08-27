class_name ReviewEditor extends Control
## ReviewEditor 模态弹窗（CR-4 T6）
## show_modal(gid, playtime) → 5星选择 + 500字限 + 提交校验 + DB.put_review

const MAX_CHARS: int = 500
const MIN_PLAYTIME: float = 600.0

@onready var _title_lbl: Label = $Card/VBox/TitleLabel
@onready var _text_edit: TextEdit = $Card/VBox/TextEdit
@onready var _char_count: Label = $Card/VBox/CharCount
@onready var _submit_btn: Button = $Card/VBox/BtnRow/SubmitBtn
@onready var _cancel_btn: Button = $Card/VBox/BtnRow/CancelBtn
@onready var _dim: ColorRect = $Dim

var _gid: String = ""
var _playtime: float = 0.0
var _stars: int = 0

func _ready() -> void:
	visible = false
	_submit_btn.pressed.connect(_on_submit)
	_cancel_btn.pressed.connect(_on_cancel)
	_text_edit.text_changed.connect(_on_text_changed)
	var star_row: HBoxContainer = $Card/VBox/StarRow
	for i in range(star_row.get_child_count()):
		var btn: Button = star_row.get_child(i) as Button
		if btn != null:
			btn.pressed.connect(_on_star_pressed.bind(i + 1))

## 显示模态：填充数据 + 回填已有评价
func show_modal(gid: String, playtime: float) -> void:
	_gid = gid
	_playtime = playtime
	_stars = 0
	_text_edit.text = ""
	var meta: Variant = Registry.lookup(gid)
	_title_lbl.text = "评价 " + ((meta as GameMeta).title if meta != null else gid)
	var existing: Variant = DB.get_review(gid)
	if existing is Dictionary:
		var ex := existing as Dictionary
		_stars = int(ex.get("stars", 0))
		_text_edit.text = str(ex.get("text", ""))
	_update_star_ui()
	_update_submit_state()
	_on_text_changed()
	_apply_theme()
	visible = true

func _on_text_changed() -> void:
	var txt := _text_edit.text
	if txt.length() > MAX_CHARS:
		_text_edit.text = txt.left(MAX_CHARS)
		_text_edit.set_caret_column(MAX_CHARS)
	_char_count.text = "%d/%d" % [_text_edit.text.length(), MAX_CHARS]
	_update_submit_state()

func _update_submit_state() -> void:
	_submit_btn.disabled = (_stars <= 0 or _playtime < MIN_PLAYTIME)

func _on_star_pressed(n: int) -> void:
	_stars = n
	_update_star_ui()
	_update_submit_state()

func _update_star_ui() -> void:
	var star_row: HBoxContainer = $Card/VBox/StarRow
	for i in range(star_row.get_child_count()):
		var btn: Button = star_row.get_child(i) as Button
		if btn == null:
			continue
		if i < _stars:
			btn.add_theme_color_override("font_color", ThemeTokens.color("gold"))
		else:
			btn.add_theme_color_override("font_color", ThemeTokens.color("star_off"))

func _on_submit() -> void:
	var review := {
		"stars": _stars,
		"text": _text_edit.text,
		"playtime_at_review": _playtime,
		"created_at": int(Time.get_unix_time_from_system()),
		"status": "local",
	}
	DB.put_review(_gid, review)
	EventBus.review_submitted.emit(_gid)
	visible = false

func _on_cancel() -> void:
	visible = false

## 主题色（全经 ThemeTokens.color，override §2）
func _apply_theme() -> void:
	_dim.color = ThemeTokens.color("ov_bg")
	var sb := StyleBoxFlat.new()
	sb.bg_color = ThemeTokens.color("card2")
	sb.border_color = ThemeTokens.color("line")
	sb.set_border_width_all(1)
	sb.set_corner_radius_all(12)
	($Card as PanelContainer).add_theme_stylebox_override("panel", sb)
	_title_lbl.add_theme_color_override("font_color", ThemeTokens.color("ink"))
	_char_count.add_theme_color_override("font_color", ThemeTokens.color("ink2"))
	_submit_btn.add_theme_color_override("font_color", ThemeTokens.color("play_ink"))
	_cancel_btn.add_theme_color_override("font_color", ThemeTokens.color("ink2"))
	var sb_submit := StyleBoxFlat.new()
	sb_submit.bg_color = ThemeTokens.color("accent")
	sb_submit.set_corner_radius_all(8)
	_submit_btn.add_theme_stylebox_override("normal", sb_submit)
	var sb_cancel := StyleBoxFlat.new()
	sb_cancel.bg_color = ThemeTokens.color("line2")
	sb_cancel.set_corner_radius_all(8)
	_cancel_btn.add_theme_stylebox_override("normal", sb_cancel)
	_update_star_ui()
