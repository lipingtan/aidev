extends Page
## 首页：ThemeToggle + Banner/继续游戏/为你推荐/热门榜/每日任务五区块（CR-5）

@onready var _theme_btn: Button = $ThemeButton
## Banner 轮播区块（CR-4 起组件自驱动，home 仅持有引用）
@onready var _section_banner: SectionBanner = $ScrollView/VBox/SectionBanner
## 继续游戏区块（SectionContinue 组件）
@onready var _section_continue: Node = $ScrollView/VBox/SectionContinue
## 为你推荐区块根节点
@onready var _for_you_root: VBoxContainer = $ScrollView/VBox/SectionForYou
## 为你推荐横向卡片容器
@onready var _for_you_hbox: HBoxContainer = $ScrollView/VBox/SectionForYou/ScrollContainer/HBox
## 热门榜区块（CR-5）
@onready var _section_charts: SectionCharts = $ScrollView/VBox/SectionCharts
## 每日任务区块（CR-5）
@onready var _section_daily: SectionDaily = $ScrollView/VBox/SectionDaily

@export var card_scene: PackedScene  # GameCard.tscn

## Recommender 非单例，_ready 时从 Services 节点取引用并注入
var _recommender: Node = null
## DailyTaskService 非 Autoload，Main/Services 节点下（FR-8）
var _daily_svc: DailyTaskService = null

func _ready() -> void:
	_theme_btn.pressed.connect(_on_theme_pressed)
	# 取 Recommender 引用并注入继续游戏区块（get_node_or_null 避免节点不存在时触发 native ERROR）
	_recommender = get_node_or_null("/root/Main/Services/Recommender")
	if _recommender == null:
		push_warning("home.gd: Recommender 节点未找到")
	_section_continue.call("set_recommender", _recommender)
	# CR-5：注入热门榜 Recommender + 每日任务服务
	_section_charts.set_recommender(_recommender as Recommender)
	_daily_svc = get_node_or_null("/root/Main/Services/DailyTaskService") as DailyTaskService
	if _daily_svc != null:
		_section_daily.set_service(_daily_svc)

func on_enter(_data: Dictionary) -> void:
	_render_for_you()

func on_resume() -> void:
	# 从详情页返回时刷新为你推荐区块（三区块各自内部驱动，RG-28）
	_render_for_you()

## 为你推荐区块：for_you 非空才渲染，空则隐藏
func _render_for_you() -> void:
	for c: Node in _for_you_hbox.get_children():
		c.queue_free()
	if _recommender == null:
		return
	var rows: Array = _recommender.call("for_you")
	_for_you_root.visible = rows.size() > 0
	for item: Dictionary in rows:
		var card: Node = card_scene.instantiate()
		## 先入树再 setup：@onready 节点在 add_child 时同步 _ready，避免 setup 访问空引用
		_for_you_hbox.add_child(card)
		card.call("setup", item["meta"], item["record"])
		card.connect("card_pressed",
			func(gid: String) -> void:
				Nav.push("res://shell/pages/detail.tscn", {"gid": gid})
		)

## 🎨 按下：scale(.9) 反馈 + 切换音 + neon/elegant 切换
func _on_theme_pressed() -> void:
	var tw := create_tween()
	tw.tween_property(_theme_btn, "scale", Vector2(0.9, 0.9), 0.06)
	tw.tween_property(_theme_btn, "scale", Vector2.ONE, 0.12)
	Sound.toggle()
	var next := "elegant" if ThemeTokens.current == "neon" else "neon"
	ThemeTokens.apply_theme(next, get_tree().root)
