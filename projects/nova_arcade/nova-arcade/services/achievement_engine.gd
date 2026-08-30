class_name AchievementEngine extends Node
## 成就判定引擎（CR-7 FR-1）
## 挂 Main/Services 节点下，非单例。
## 订阅 game_finished / review_submitted，evaluate() 统一判定游戏级+全局成就。

const GLOBAL_DEFS: Array[Dictionary] = [
	{"id": "global_collector", "name": "收藏家", "desc": "拥有全部游戏记录", "points": 30},
	{"id": "global_reviewer", "name": "好评人", "desc": "提交≥3条评价", "points": 20},
	{"id": "global_marathon", "name": "马拉松", "desc": "累计游玩≥10小时", "points": 50},
]

## _enter_tree() 连接信号（A-1修复：早于 _ready，防信号丢失）
func _enter_tree() -> void:
	EventBus.game_finished.connect(_on_game_finished)
	EventBus.review_submitted.connect(_on_review_submitted)

## 游戏结束事件：判定 result.achievements[] 中的游戏级成就
func _on_game_finished(gid: String, result: Dictionary) -> void:
	var aids: Array = result.get("achievements", [])
	for aid in aids:
		_try_unlock(str(aid))

## 评价提交事件：判定好评人全局成就
func _on_review_submitted(gid: String) -> void:
	_evaluate_global()

## 尝试解锁某成就（幂等：先检查 has(aid)，再调 DB.unlock）
func _try_unlock(aid: String) -> void:
	if DB.get_achievements().has(aid):
		return  # 已解锁，跳过
	var def: Variant = Registry.ach_def(aid)
	if def == null:
		return  # 定义不存在，跳过
	DB.unlock(aid)
	EventBus.achievement_unlocked.emit(def as Dictionary, int((def as Dictionary).get("points", 0)))

## 判定全局成就（收藏家/好评人/马拉松）
func _evaluate_global() -> void:
	for gdef in GLOBAL_DEFS:
		var aid: String = str(gdef["id"])
		if DB.get_achievements().has(aid):
			continue
		if _check_global(gdef):
			DB.unlock(aid)
			EventBus.achievement_unlocked.emit(gdef, int(gdef["points"]))

## 全局成就条件检查（P-3修复：收藏家阈值动态 = Registry.all().size()）
func _check_global(gdef: Dictionary) -> bool:
	match gdef["id"]:
		"global_collector":
			return DB.list_records().size() >= maxi(Registry.all().size(), 1)
		"global_reviewer":
			return _count_reviews() >= 3
		"global_marathon":
			return _total_playtime_hours() >= 10.0
	return false

## 统计有效评价数
func _count_reviews() -> int:
	var count: int = 0
	for gid: String in DB.list_records():
		if DB.get_review(gid) != null:
			count += 1
	return count

## 累计游玩时长（小时）；B-2确认：total_playtime 为累加字段（launcher_util.finish_record 写入）
func _total_playtime_hours() -> float:
	var total: float = 0.0
	for gid: String in DB.list_records():
		var rec: Variant = DB.get_record(gid)
		if rec is Dictionary:
			total += float((rec as Dictionary).get("total_playtime", 0))
	return total / 3600.0
