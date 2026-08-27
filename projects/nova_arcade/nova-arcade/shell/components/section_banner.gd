class_name SectionBanner extends Control
## Banner 轮播组件（CR-5 FR-2）：渐变色块，4s Timer，指示点，点击跳详情
##
## - banner[] 为空 → visible=false；单张不启动 Timer（Q4）
## - image_path 空 → ThemeTokens.grad("banner") + GradientTexture2D 色块（P-2）
## - 主题切换时重绘当前条目颜色（token 驱动，禁硬编码）

@onready var _banner_rect: TextureRect = $BannerRect
@onready var _title_lbl: Label = $TitleLabel
@onready var _dots_row: HBoxContainer = $DotsRow

var _items: Array[Dictionary] = []
var _current: int = 0
var _timer: SceneTreeTimer = null

func _ready() -> void:
	visible = false
	## 延迟到 autoload 全部 _ready 后再读 editorial（Registry.reload 在其 _ready 执行）
	call_deferred("_load_banners")
	EventBus.theme_changed.connect(_on_theme_changed)

## 幂等：重载前清空（design §12）；旧 timer 引用置空（A-2）
func _load_banners() -> void:
	_items.clear()
	_timer = null
	var raw: Variant = Registry.get_editorial().get("banner", [])
	if not (raw is Array):
		return
	for item in raw as Array:
		if item is Dictionary:
			_items.append(item)
	visible = not _items.is_empty()
	if _items.is_empty():
		return
	_build_dots()
	_show(0)
	if _items.size() > 1:
		_start_timer()

func _build_dots() -> void:
	for c in _dots_row.get_children():
		c.queue_free()
	for i in _items.size():
		var dot := Button.new()
		dot.flat = true
		dot.custom_minimum_size = Vector2(5, 5)
		_dots_row.add_child(dot)
	_update_dots()

## 指示点：当前宽 14px / 非激活 5px；颜色走 token（accent/ink3）
func _update_dots() -> void:
	for i in _dots_row.get_child_count():
		var dot := _dots_row.get_child(i) as Button
		if dot == null:
			continue
		var active: bool = i == _current
		dot.custom_minimum_size = Vector2(14 if active else 5, 5)
		var sb := StyleBoxFlat.new()
		sb.bg_color = ThemeTokens.color("accent" if active else "ink3")
		sb.set_corner_radius_all(3)
		dot.add_theme_stylebox_override("normal", sb)

## 展示第 idx 条（渐变色块 + 标题）
func _show(idx: int) -> void:
	_current = idx
	var item: Dictionary = _items[idx]
	## image_path 非空时 M2 加载 Texture；M1 一律渐变色块
	var tex := GradientTexture2D.new()
	tex.gradient = ThemeTokens.grad("banner")
	_banner_rect.texture = tex
	_title_lbl.text = str(item.get("title", ""))
	_title_lbl.add_theme_color_override("font_color", ThemeTokens.color("ink"))
	_update_dots()

func _start_timer() -> void:
	_timer = get_tree().create_timer(4.0)
	_timer.timeout.connect(_on_timer_tick)

## 淡出 0.15s → 切换 → 淡入 0.15s（Q1），循环
func _on_timer_tick() -> void:
	var next: int = (_current + 1) % _items.size()
	var tw := create_tween()
	tw.tween_property(_banner_rect, "modulate:a", 0.0, 0.15)
	tw.tween_callback(_show.bind(next))
	tw.tween_property(_banner_rect, "modulate:a", 1.0, 0.15)
	_start_timer()

## 点击整体 BannerRect → 详情（target_gid 空不响应，RG-29）
func _gui_input(event: InputEvent) -> void:
	if event is InputEventMouseButton:
		var e := event as InputEventMouseButton
		if e.button_index == MOUSE_BUTTON_LEFT and e.pressed:
			var gid: String = str(_items[_current].get("target_gid", ""))
			if gid != "":
				Nav.push("res://shell/pages/detail.tscn", {"gid": gid})

func _on_theme_changed(_name: String) -> void:
	if not _items.is_empty():
		_show(_current)
