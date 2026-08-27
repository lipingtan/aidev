class_name SectionDaily extends VBoxContainer
## 每日任务区块（CR-5 FR-7）：3 条任务行 + 进度指示，由 DailyTaskService 驱动
##
## - set_service(svc) 注入并连接 svc.tasks_updated → _render()
## - 完成 ✓ token ok / 未完成 ○ token ink3；全完成显示 CompletedLabel

@onready var _progress_lbl: Label = $Header/ProgressLabel
@onready var _task_list: VBoxContainer = $TaskList
@onready var _completed_lbl: Label = $CompletedLabel

var _daily_svc: DailyTaskService = null

## 注入服务并连接刷新信号（A-3 同模式：注入后立即渲染）
func set_service(svc: DailyTaskService) -> void:
	_daily_svc = svc
	_daily_svc.tasks_updated.connect(_on_tasks_updated)
	if is_node_ready():
		_render(_daily_svc.get_tasks())

func _ready() -> void:
	_render([])
	EventBus.theme_changed.connect(_on_theme_changed)

func _on_tasks_updated(tasks: Array[Dictionary]) -> void:
	_render(tasks)

## 先 queue_free 旧行再重建（design §8）
func _render(tasks: Array[Dictionary]) -> void:
	for c in _task_list.get_children():
		c.queue_free()
	var done_count: int = 0
	for task in tasks:
		if bool(task.get("done", false)):
			done_count += 1
		_task_list.add_child(_make_task_row(task))
	_progress_lbl.text = "%d/%d" % [done_count, tasks.size()]
	_completed_lbl.visible = done_count == tasks.size() and tasks.size() > 0

## 任务行：✓/○ + label（token ok / ink3）
func _make_task_row(task: Dictionary) -> HBoxContainer:
	var done := bool(task.get("done", false))
	var hbox := HBoxContainer.new()
	var check := Label.new()
	check.text = "✓" if done else "○"
	check.add_theme_color_override("font_color", ThemeTokens.color("ok" if done else "ink3"))
	hbox.add_child(check)
	var lbl := Label.new()
	lbl.text = str(task.get("label", ""))
	lbl.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	lbl.add_theme_color_override("font_color", ThemeTokens.color("ink"))
	hbox.add_child(lbl)
	return hbox

func _on_theme_changed(_name: String) -> void:
	if _daily_svc != null:
		_render(_daily_svc.get_tasks())
