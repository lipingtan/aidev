extends Page
## 详情页：基础信息块 + CTA 固定区；入页 220ms 右滑入（Nav 负责）

@onready var _cta_bar: Node = $CtaBar
@onready var _title: Label = $ScrollContainer/VBox/TitleLabel
@onready var _version: Label = $ScrollContainer/VBox/VersionLabel
@onready var _desc: Label = $ScrollContainer/VBox/DescLabel
@onready var _back_btn: Button = $BackButton
@onready var _reviews_btn: Button = $ScrollContainer/VBox/ReviewsBtn

var _gid: String = ""

func on_enter(data: Dictionary) -> void:
	_gid = data.get("gid", "")
	var meta: Variant = Registry.lookup(_gid)
	if meta == null:
		push_error("DetailPage: 未找到游戏 %s" % _gid)
		Nav.pop()
		return
	_title.text = (meta as GameMeta).title
	_version.text = "v" + (meta as GameMeta).version
	_desc.text = (meta as GameMeta).desc
	_cta_bar.call("setup", _gid)
	## 绑定评价按钮（仅首次 on_enter 连接，避免重复）
	if not _reviews_btn.pressed.is_connected(_on_reviews_pressed):
		_reviews_btn.pressed.connect(_on_reviews_pressed)

## Risk-3：结算卡关闭后回 Tab 根页走 on_resume，需刷新 CTA
func on_resume() -> void:
	if _gid != "":
		_cta_bar.call("refresh")

func _on_reviews_pressed() -> void:
	Nav.push("res://shell/pages/reviews.tscn", {"gid": _gid})

func _ready() -> void:
	_back_btn.pressed.connect(func(): Nav.pop())
