class_name SectionContinue extends VBoxContainer
## 继续游戏区块控制器（CR-3 §5）
## Recommender 非单例，由 home.gd 在 _ready 时通过 set_recommender() 注入
## records_updated 信号在 _ready 中连接一次，变更后自动重建卡片列表

@export var card_scene: PackedScene  # 指向 game_card.tscn

@onready var _scroll: ScrollContainer = $ScrollContainer
@onready var _hbox: HBoxContainer = $ScrollContainer/HBox

var _recommender: Node = null

## 注入 Recommender（由 home.gd 在 _ready 中调用）
func set_recommender(r: Node) -> void:
	_recommender = r

func _ready() -> void:
	# 信号只连一次；DB.upsert_record 每次写入都发此信号
	EventBus.records_updated.connect(_on_records_updated)
	render()

## 重建卡片列表
func render() -> void:
	# 先释放旧卡片
	for c: Node in _hbox.get_children():
		c.queue_free()
	# Recommender 未注入时直接返回，区块保持当前 visible 状态
	if _recommender == null:
		return
	var rows: Array = _recommender.call("continue_row")
	visible = rows.size() > 0
	for item: Dictionary in rows:
		var card: Node = card_scene.instantiate()
		card.connect("card_pressed",
			func(gid: String) -> void:
				Launcher.launch(gid)
		)
		card.connect("card_long_pressed",
			_on_long_press.bind(item["meta"].id, item["meta"].title)
		)
		_hbox.add_child(card)
		card.call("setup", item["meta"], item["record"])

## records_updated 触发时重建（gid 参数保留但不作过滤，整行统一刷新）
func _on_records_updated(_gid: String) -> void:
	if not is_inside_tree():
		return
	render()

## 长按卡片：弹出确认气泡，确认后清除 last_played（records_updated 自动触发 render）
func _on_long_press(gid: String, title: String) -> void:
	var bubble: Node = get_tree().root.find_child("ConfirmBubble", true, false)
	if bubble == null:
		push_error("SectionContinue: 未找到 ConfirmBubble 节点")
		return
	bubble.call("show_bubble",
		"从记录移除「%s」？" % title,
		func() -> void:
			# 只清 last_played，不手动调 render()
			# DB.upsert_record 会发 records_updated，由信号触发重建
			DB.upsert_record(gid, {"last_played": 0})
	)
