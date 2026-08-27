extends Page
## 搜索页：输入去抖 150ms + 热搜 + 历史（去重置顶）+ 实时结果

@export var card_scene: PackedScene

@onready var _input: LineEdit = $TopBar/SearchInput
@onready var _cancel_btn: Button = $TopBar/CancelBtn
@onready var _default_view: VBoxContainer = $DefaultView
@onready var _result_list: VBoxContainer = $ResultList
@onready var _empty_state: Label = $EmptyState
@onready var _hot_row: HBoxContainer = $DefaultView/HotSection/HotRow
@onready var _hist_row: HBoxContainer = $DefaultView/HistorySection/HistRow
@onready var _hist_clear_btn: Button = $DefaultView/HistorySection/HistHeader/ClearBtn

var _searcher: Searcher = null
var _debounce_active: bool = false

func _ready() -> void:
	_input.text_changed.connect(_on_text_changed)
	_cancel_btn.pressed.connect(_on_cancel_pressed)
	_hist_clear_btn.pressed.connect(_on_clear_history)
	_searcher = get_node_or_null("/root/Main/Services/Searcher") as Searcher

func _on_cancel_pressed() -> void:
	Nav.pop()

## P-5：on_enter = 新入口，重置搜索状态
func on_enter(_data: Dictionary) -> void:
	_input.grab_focus()
	_input.text = ""
	_debounce_active = false
	_show_default()

## P-5：on_resume = Tab 切回，保持输入状态，仅刷新历史
func on_resume() -> void:
	_render_history()

func _on_text_changed(new_text: String) -> void:
	_debounce_active = false
	if new_text.strip_edges() == "":
		_show_default()
		return
	_debounce_active = true
	var timer := get_tree().create_timer(0.15)
	await timer.timeout
	if not _debounce_active:
		return
	_debounce_active = false
	_run_search(new_text)

func _run_search(q: String) -> void:
	_default_view.visible = false
	for c in _result_list.get_children():
		c.queue_free()
	if _searcher == null:
		push_warning("SearchPage: Searcher 未找到")
		return
	var gids := _searcher.query(q)
	if gids.is_empty():
		_result_list.visible = false
		_empty_state.text = "没有找到「%s」" % q
		_empty_state.visible = true
		return
	_empty_state.visible = false
	_result_list.visible = true
	for gid in gids:
		var meta: Variant = Registry.lookup(gid)
		if meta == null:
			continue
		if card_scene == null:
			continue
		var card := card_scene.instantiate() as GameCard
		card.setup(meta as GameMeta, DB.get_record(gid))
		## P-2：bind(q) 追加 q 为第二个参数；gid 来自 card_pressed 信号（第一个参数）
		card.card_pressed.connect(_on_result_pressed.bind(q))
		_result_list.add_child(card)

## P-2：gid 来自 card_pressed 信号，q 来自 bind 追加
func _on_result_pressed(gid: String, q: String) -> void:
	DB.add_search_history(q)
	Nav.push("res://shell/pages/detail.tscn", {"gid": gid})

func _show_default() -> void:
	_result_list.visible = false
	_empty_state.visible = false
	_default_view.visible = true
	_render_default_view()

func _render_default_view() -> void:
	_render_hot_words()
	_render_history()

func _render_hot_words() -> void:
	for c in _hot_row.get_children():
		c.queue_free()
	if _searcher == null:
		return
	for q in _searcher.hot_queries():
		var btn := Button.new()
		btn.text = q
		## A-1：不可用单行多语句 lambda，改用 bind + 具名方法
		btn.pressed.connect(_on_hot_or_hist_pressed.bind(q))
		_hot_row.add_child(btn)

func _render_history() -> void:
	for c in _hist_row.get_children():
		c.queue_free()
	for q in DB.get_search_history():
		var btn := Button.new()
		btn.text = q
		btn.pressed.connect(_on_hot_or_hist_pressed.bind(q))
		_hist_row.add_child(btn)

## A-1：热搜/历史词点击，替代单行多语句 lambda
func _on_hot_or_hist_pressed(q: String) -> void:
	_input.text = q
	_run_search(q)

func _on_clear_history() -> void:
	DB.clear_search_history()
	_render_history()
