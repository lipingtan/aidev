class_name DailyTaskService extends Node
## 每日任务逻辑服务（CR-5 FR-5）
##
## - 非 Autoload；挂 Main/Services 节点下，home.gd 通过 get_node_or_null 取引用
## - _ready() 连接 EventBus.game_launched / game_finished
## - 跨天检测：DB.get_daily().date != today → 重置进度与 reward_claimed
## - 全部完成 → Toast + reward_claimed=true（防重复弹 Toast）
## - 进度经 DB.update_daily_tasks() 防抖落盘（design §3）

signal tasks_updated(tasks: Array[Dictionary])

## 任务定义（硬编码；M2 可扩展为后端下发）
const TASK_DEFS: Array[Dictionary] = [
	{"id": "launch", "label": "今日启动游戏", "target": 1},
	{"id": "finish", "label": "今日完成 1 局", "target": 1},
	{"id": "playtime", "label": "今日累计游玩 5 分钟", "target": 300},
]

## 当日进度缓存 {task_id: progress_value}
var _progress: Dictionary = {}
## 今日已领奖
var _reward_claimed: bool = false
## 测试钩子：_show_reward_toast 累计调用次数（RG-25/RG-26）
var toast_shown_count: int = 0

func _ready() -> void:
	_load_or_reset()
	EventBus.game_launched.connect(_on_game_launched)
	EventBus.game_finished.connect(_on_game_finished)

## 加载/重置每日进度（跨天 → 全量重置 + 落盘）
func _load_or_reset() -> void:
	var today := Time.get_date_string_from_system()
	var daily := DB.get_daily()
	if str(daily.get("date", "")) != today:
		_progress = {}
		_reward_claimed = false
		_save()
	else:
		var saved := DB.get_daily_tasks()
		_progress = saved.get("progress", {}) if saved is Dictionary else {}
		_reward_claimed = bool(saved.get("reward_claimed", false)) if saved is Dictionary else false

## 获取当前任务状态（供 SectionDaily 渲染）
func get_tasks() -> Array[Dictionary]:
	var result: Array[Dictionary] = []
	for def in TASK_DEFS:
		var progress: int = int(_progress.get(str(def["id"]), 0))
		result.append({
			"id": def["id"],
			"label": str(def["label"]),
			"done": progress >= int(def["target"]),
			"progress": progress,
			"target": int(def["target"]),
		})
	return result

func _on_game_launched(_gid: String) -> void:
	_update_task("launch", 1, true)

func _on_game_finished(_gid: String, result: Dictionary) -> void:
	_update_task("finish", 1, true)
	var pt: float = float(result.get("playtime", 0.0))
	var cur_pt: int = int(_progress.get("playtime", 0))
	_update_task("playtime", cur_pt + int(pt), false)

## 更新任务进度（cap_at_target=true 时封顶 target，防重复触发）
func _update_task(task_id: String, new_val: int, cap_at_target: bool) -> void:
	var def := _find_def(task_id)
	if def.is_empty():
		return
	var target: int = int(def["target"])
	var val: int = mini(new_val, target) if cap_at_target else new_val
	_progress[task_id] = val
	_save()
	tasks_updated.emit(get_tasks())
	## 全部完成 + 未领奖 → Toast + 标记（防重）
	if not _reward_claimed and _all_done():
		_reward_claimed = true
		_save()
		_show_reward_toast()

func _all_done() -> bool:
	for def in TASK_DEFS:
		if int(_progress.get(str(def["id"]), 0)) < int(def["target"]):
			return false
	return true

func _find_def(task_id: String) -> Dictionary:
	for d in TASK_DEFS:
		if str(d["id"]) == task_id:
			return d
	return {}

func _save() -> void:
	DB.update_daily_tasks({"progress": _progress, "reward_claimed": _reward_claimed})

## Toast 层在 Main/App/ToastLayer；find_child 找不到时静默跳过（测试场景无 Main 时）
func _show_reward_toast() -> void:
	toast_shown_count += 1
	var layer := get_tree().root.find_child("ToastLayer", true, false)
	if layer != null:
		layer.call("show_msg", "今日任务全部完成 🎉")
