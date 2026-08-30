extends Page
## 评价列表页（FR-7）：头部平均星 + [写评价] + 列表

@onready var _back_btn: Button = $BackButton
@onready var _avg_label: Label = $VBox/Header/AvgLabel
@onready var _write_btn: Button = $VBox/Header/WriteBtn
@onready var _list_vbox: VBoxContainer = $VBox/ListView
@onready var _empty_state: Label = $VBox/EmptyState

var _gid: String = ""

func _ready() -> void:
	_back_btn.pressed.connect(_on_back_pressed)
	_write_btn.pressed.connect(_on_write_review)
	EventBus.review_submitted.connect(_on_review_changed)
	EventBus.review_deleted.connect(_on_review_changed)

func _on_back_pressed() -> void:
	Nav.pop()

func on_enter(data: Dictionary) -> void:
	_gid = data.get("gid", "")
	_render()

func on_resume() -> void:
	_render()

func _on_write_review() -> void:
	var rec: Variant = DB.get_record(_gid)
	var playtime := float((rec as Dictionary).get("total_playtime", 0.0)) if rec is Dictionary else 0.0
	var editor := get_tree().root.find_child("ReviewEditor", true, false)
	if editor != null:
		editor.call("show_modal", _gid, playtime)

func _on_review_changed(_gid_changed: String) -> void:
	if _gid_changed == _gid:
		_render()

func _render() -> void:
	for c in _list_vbox.get_children():
		c.queue_free()
	var review: Variant = DB.get_review(_gid)
	var rec: Variant = DB.get_record(_gid)
	var playtime := float((rec as Dictionary).get("total_playtime", 0.0)) if rec is Dictionary else 0.0
	var finish := int((rec as Dictionary).get("finish_count", 0)) if rec is Dictionary else 0
	## B-2：playtime<600s 或 finish_count<2 时写评价按钮置灰
	_write_btn.disabled = playtime < 600.0 or finish < 2
	if review == null:
		_empty_state.visible = true
		_list_vbox.visible = false
		_avg_label.text = "暂无评分"
		return
	_empty_state.visible = false
	_list_vbox.visible = true
	var rv := review as Dictionary
	## P-3：明确文案为"你的评分"
	_avg_label.text = "你的评分：★%d" % int(rv.get("stars", 0))
	_add_review_item(rv, true)

func _add_review_item(rv: Dictionary, is_own: bool) -> void:
	var item := VBoxContainer.new()
	var stars_lbl := Label.new()
	stars_lbl.text = "★".repeat(int(rv.get("stars", 0)))
	stars_lbl.add_theme_color_override("font_color", ThemeTokens.color("gold"))
	item.add_child(stars_lbl)
	var text_lbl := Label.new()
	text_lbl.text = str(rv.get("text", ""))
	text_lbl.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	item.add_child(text_lbl)
	if is_own:
		var btn_row := HBoxContainer.new()
		var edit_btn := Button.new()
		edit_btn.text = "修改"
		edit_btn.pressed.connect(_on_write_review)
		var del_btn := Button.new()
		del_btn.text = "删除"
		## B-1：删除前二次确认
		del_btn.pressed.connect(_on_delete_review_confirm)
		btn_row.add_child(edit_btn)
		btn_row.add_child(del_btn)
		item.add_child(btn_row)
	_list_vbox.add_child(item)

## B-1：删除确认弹窗（ConfirmBubble）
func _on_delete_review_confirm() -> void:
	var bubble := get_tree().root.find_child("ConfirmBubble", true, false)
	if bubble == null:
		_on_delete_review()
		return
	bubble.call("show_bubble", "确认删除这条评价？", Callable(self, "_on_delete_review"))

func _on_delete_review() -> void:
	DB.delete_review(_gid)
	EventBus.review_deleted.emit(_gid)
