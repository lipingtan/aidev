class_name SectionCharts extends VBoxContainer
## 热门榜区块（CR-5 FR-4）：三 Tab 即时切换，列表行按排名展示
##
## - set_recommender(r) 注入后立即刷新（A-3）；null → EmptyState
## - 排名 1/2/3 用 rank1/rank2/rank3 token；人数 ≥10000 显示 "X.X万+"（Q6）

@onready var _tab_row: HBoxContainer = $Header/TabRow
@onready var _list_vbox: VBoxContainer = $ListView
@onready var _empty_state: Label = $EmptyState

const MODES: Array[String] = ["all", "new", "rated"]
const MODE_LABELS: Array[String] = ["综合", "新游", "好评"]
var _current_mode: String = "all"
var _recommender: Recommender = null

## 注入 Recommender；is_node_ready 时立即重建列表（A-3）
func set_recommender(r: Recommender) -> void:
	_recommender = r
	if is_node_ready():
		_render()

func _ready() -> void:
	_build_tabs()
	_render()
	EventBus.theme_changed.connect(_on_theme_changed)

func _build_tabs() -> void:
	for i in MODES.size():
		var btn := Button.new()
		btn.text = MODE_LABELS[i]
		btn.pressed.connect(_on_tab_pressed.bind(MODES[i]))
		_tab_row.add_child(btn)

## Tab 切换即时重建，无动画（Q2）
func _on_tab_pressed(mode: String) -> void:
	_current_mode = mode
	_render()

func _render() -> void:
	for c in _list_vbox.get_children():
		c.queue_free()
	if _recommender == null:
		_empty_state.visible = true
		_list_vbox.visible = false
		return
	var rows := _recommender.charts(_current_mode)
	_empty_state.visible = rows.is_empty()
	_list_vbox.visible = not rows.is_empty()
	for row in rows:
		_list_vbox.add_child(_make_row(row))

## 行：排名 + 图标 + 游戏名 + ★均分（无评价不显示）+ 人数（P-1 含图标）
func _make_row(row: Dictionary) -> HBoxContainer:
	var meta := row["meta"] as GameMeta
	var rank: int = int(row.get("rank", 0))
	var players: int = int(row.get("players", 10000))

	var hbox := HBoxContainer.new()
	## 排名数字（rank1/rank2/rank3 token）
	var rank_lbl := Label.new()
	rank_lbl.text = str(rank)
	rank_lbl.custom_minimum_size = Vector2(40, 0)
	var rank_color: String = "rank1" if rank == 1 else ("rank2" if rank == 2 else ("rank3" if rank == 3 else "ink2"))
	rank_lbl.add_theme_color_override("font_color", ThemeTokens.color(rank_color))
	hbox.add_child(rank_lbl)
	## 图标（meta.icon；空路径或资源不存在时隐藏）
	var icon_rect := TextureRect.new()
	if meta.icon != "" and ResourceLoader.exists(meta.icon):
		icon_rect.texture = load(meta.icon)
	else:
		icon_rect.visible = false
	icon_rect.custom_minimum_size = Vector2(36, 36)
	icon_rect.expand_mode = TextureRect.EXPAND_FIT_WIDTH_PROPORTIONAL
	hbox.add_child(icon_rect)
	## 游戏名
	var name_lbl := Label.new()
	name_lbl.text = meta.title
	name_lbl.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	name_lbl.add_theme_color_override("font_color", ThemeTokens.color("ink"))
	hbox.add_child(name_lbl)
	## ★均分（有评价才显示）
	var rv: Variant = DB.get_review(meta.id)
	if rv is Dictionary:
		var stars_lbl := Label.new()
		stars_lbl.text = "★%.1f" % float((rv as Dictionary).get("stars", 0))
		stars_lbl.add_theme_color_override("font_color", ThemeTokens.color("gold"))
		hbox.add_child(stars_lbl)
	## 人数（Q6：base=10000+sessions，≥10000 显示 "X.X万+"）
	var players_lbl := Label.new()
	players_lbl.text = _fmt_players(players)
	players_lbl.add_theme_color_override("font_color", ThemeTokens.color("ink2"))
	hbox.add_child(players_lbl)
	## 行点击 → 详情
	hbox.gui_input.connect(_on_row_input.bind(meta.id))
	return hbox

## 人数格式：≥10000 → "X.X万+"；否则 "N人"
func _fmt_players(n: int) -> String:
	if n >= 10000:
		return "%.1f万+" % (float(n) / 10000.0)
	return str(n) + "人"

func _on_row_input(event: InputEvent, gid: String) -> void:
	if event is InputEventMouseButton:
		var e := event as InputEventMouseButton
		if e.button_index == MOUSE_BUTTON_LEFT and e.pressed:
			Launcher.launch(gid)

func _on_theme_changed(_name: String) -> void:
	_render()
