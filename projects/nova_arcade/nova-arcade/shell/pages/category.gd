extends Page
## 分类页：左侧分类栏 + 顶部筛选 chips + 右侧 2 列网格

@export var card_scene: PackedScene  # 编辑器赋值 GameCard.tscn

@onready var _cat_list: VBoxContainer = $HSplit/CategoryList
@onready var _filter_row: HBoxContainer = $HSplit/RightPanel/FilterRow
@onready var _grid: GridContainer = $HSplit/RightPanel/GridContainer
@onready var _empty_state: VBoxContainer = $HSplit/RightPanel/EmptyState
@onready var _empty_clear_btn: Button = $HSplit/RightPanel/EmptyState/ClearBtn

var _filter: Dictionary = {
	"category": "",
	"price_models": [],
	"min_rating": 0.0,
	"is_new": false,
}

## B-4：GameMeta.category 枚举为 puzzle/action/arcade/casual/roguelike
## 合并"消除/益智"为单按钮(id="puzzle")，避免两个按钮同效果
const CATEGORIES: Array = [
	{"id": "", "label": "全部"},
	{"id": "puzzle", "label": "消除/益智"},
	{"id": "action", "label": "动作"},
	{"id": "arcade", "label": "街机"},
	{"id": "casual", "label": "休闲"},
	{"id": "roguelike", "label": "Roguelike"},
]

## chips 定义（P-4：离线可玩 hidden=true，M2 补 GameMeta.offline 字段后解锁）
const CHIPS: Array = [
	{"key": "all", "label": "全部", "exclusive": true},
	{"key": "free", "label": "免费", "price_models": ["free", "ad", "iap"]},
	{"key": "paid", "label": "付费", "price_models": ["paid", "trial"]},
	{"key": "offline", "label": "离线可玩", "hidden": true},
	{"key": "rating", "label": "评分≥4.0", "min_rating": 4.0},
	{"key": "new", "label": "新游", "is_new": true},
]

func _ready() -> void:
	_build_category_list()
	_build_filter_chips()
	_empty_clear_btn.pressed.connect(_on_clear_filter)

func on_enter(_data: Dictionary) -> void:
	_render_grid()

func on_resume() -> void:
	_render_grid()

func _build_category_list() -> void:
	for item in CATEGORIES:
		var btn := Button.new()
		btn.text = str(item["label"])
		btn.pressed.connect(_on_category_selected.bind(str(item["id"])))
		_cat_list.add_child(btn)

func _build_filter_chips() -> void:
	for chip in CHIPS:
		if chip.get("hidden", false):
			continue  # P-4：离线 chip M1 跳过
		var btn := Button.new()
		btn.text = str(chip["label"])
		btn.pressed.connect(_on_chip_pressed.bind(chip))
		_filter_row.add_child(btn)

func _on_category_selected(cat_id: String) -> void:
	_filter["category"] = cat_id
	_render_grid()

func _on_chip_pressed(chip: Dictionary) -> void:
	if chip.get("exclusive", false):
		## "全部" chip：清除所有筛选
		_filter["price_models"] = []
		_filter["min_rating"] = 0.0
		_filter["is_new"] = false
	else:
		## 非"全部" chip：按类型切换
		if chip.has("price_models"):
			var models: Array = chip["price_models"] as Array
			if _filter["price_models"] == models:
				_filter["price_models"] = []
			else:
				_filter["price_models"] = models
		if chip.has("min_rating"):
			var mr: float = float(chip["min_rating"])
			_filter["min_rating"] = 0.0 if _filter["min_rating"] == mr else mr
		if chip.has("is_new"):
			_filter["is_new"] = not bool(_filter.get("is_new", false))
	_render_grid()

func _on_clear_filter() -> void:
	_filter = {"category": "", "price_models": [], "min_rating": 0.0, "is_new": false}
	_render_grid()

func _render_grid() -> void:
	for c in _grid.get_children():
		c.queue_free()
	var results := Registry.query(_filter)
	_empty_state.visible = results.is_empty()
	_grid.visible = not results.is_empty()
	if card_scene == null:
		return
	for meta in results:
		var card := card_scene.instantiate() as GameCard
		## 先入树再 setup：@onready 节点在 add_child 时同步 _ready，避免 setup 访问空引用
		_grid.add_child(card)
		card.setup(meta as GameMeta, DB.get_record((meta as GameMeta).id), "grid")
		card.card_pressed.connect(_on_card_pressed)

func _on_card_pressed(gid: String) -> void:
	Nav.push("res://shell/pages/detail.tscn", {"gid": gid})
