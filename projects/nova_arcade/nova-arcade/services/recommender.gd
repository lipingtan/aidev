class_name Recommender extends Node
## 推荐服务：继续游戏行 + 为你推荐（画像加权）+ 热门榜
## 非单例；挂 Main/Services 节点下，由 home.gd 通过 get_node 取引用。

var _cached_result: Array[Dictionary] = []
var _cache_dirty: bool = true

func _ready() -> void:
	EventBus.records_updated.connect(func(_gid: String) -> void: _cache_dirty = true)

# ── 公开 API ──────────────────────────────────────────────

## 继续游戏行
## 读 DB.list_records()，过滤 last_played > 0，按 last_played 倒序，返回 top5
## 返回格式：[{"meta": GameMeta, "record": Dictionary}, ...]
func continue_row() -> Array[Dictionary]:
	var rows: Array[Dictionary] = []
	for gid: String in DB.list_records():
		var rec: Variant = DB.get_record(gid)
		if rec == null:
			continue
		var r := rec as Dictionary
		if int(r.get("last_played", 0)) > 0:
			var meta: Variant = Registry.lookup(gid)
			if meta != null:
				rows.append({"meta": meta as GameMeta, "record": r})
	rows.sort_custom(func(a: Dictionary, b: Dictionary) -> bool:
		return int(a["record"]["last_played"]) > int(b["record"]["last_played"])
	)
	return rows.slice(0, 5)

## 为你推荐（CR-7 T6：画像加权 + 未玩过优先 + 冷启动回退）
## 返回 top 4，格式 [{"meta": GameMeta, "record": Variant}, ...]
func for_you() -> Array[Dictionary]:
	if not _cache_dirty:
		return _cached_result
	var profile: Dictionary = _build_profile()
	var scored: Array[Dictionary] = []
	for meta: GameMeta in Registry.all():
		var score: float = _score_for(meta, profile)
		var rec: Variant = DB.get_record(meta.id)
		var played: bool = (rec is Dictionary) and int((rec as Dictionary).get("last_played", 0)) > 0
		scored.append({"meta": meta, "record": rec, "score": score, "played": played})
	scored.sort_custom(func(a: Dictionary, b: Dictionary) -> bool:
		if a["played"] != b["played"]:
			return not a["played"]  # 未玩过排前
		if a["score"] != b["score"]:
			return a["score"] > b["score"]
		return (a["meta"] as GameMeta).title.to_lower() < (b["meta"] as GameMeta).title.to_lower()
	)
	var result: Array[Dictionary] = []
	if profile.is_empty():
		result = _fallback_editorial()
	else:
		for item in scored.slice(0, 4):
			result.append({"meta": item["meta"], "record": item["record"]})
	_cached_result = result
	_cache_dirty = false
	return result

## 构建用户画像（类目偏好 + 标签偏好）
func _build_profile() -> Dictionary:
	var cats: Dictionary = {}
	var tags: Dictionary = {}
	for gid: String in DB.list_records():
		var rec: Variant = DB.get_record(gid)
		if rec is Dictionary and int((rec as Dictionary).get("last_played", 0)) > 0:
			var meta: Variant = Registry.lookup(gid)
			if meta is GameMeta:
				cats[str(meta.category)] = cats.get(str(meta.category), 0) + 1
				for tag in meta.tags:
					tags[str(tag)] = tags.get(str(tag), 0) + 1
	for gid: String in DB.list_records():
		if DB.get_review(gid) != null:
			var meta: Variant = Registry.lookup(gid)
			if meta is GameMeta:
				cats[str(meta.category)] = cats.get(str(meta.category), 0) + 2
	return {"cats": cats, "tags": tags}

## 计算画像匹配分（类目权重 5x + 标签权重 1x）
func _score_for(meta: GameMeta, profile: Dictionary) -> float:
	if profile.is_empty():
		return 0.0
	var score: float = 0.0
	var cats: Dictionary = profile.get("cats", {})
	var tags: Dictionary = profile.get("tags", {})
	score += float(cats.get(str(meta.category), 0)) * 5.0
	for tag in meta.tags:
		score += float(tags.get(str(tag), 0))
	return score

## 冷启动回退：charts("all") 转 for_you() 格式 {"meta","record"}（B-3修复）
func _fallback_editorial() -> Array[Dictionary]:
	var charts_rows: Array[Dictionary] = charts("all")
	var result: Array[Dictionary] = []
	for row in charts_rows.slice(0, 4):
		result.append({"meta": row["meta"], "record": row["record"]})
	return result

# ── 热门榜（CR-5 FR-3）──────────────────────────────────────

## 热门榜（mode: "all"=综合 / "new"=新游 / "rated"=好评；未知 mode 按 "all"）
func charts(mode: String) -> Array[Dictionary]:
	var all_metas: Array[GameMeta] = Registry.all()
	var rows: Array[Dictionary] = []
	for meta: GameMeta in all_metas:
		var rec: Variant = DB.get_record(meta.id)
		rows.append({"meta": meta, "record": rec, "rank": 0, "players": _calc_players(meta.id, rec)})
	match mode:
		"new":
			rows.sort_custom(_cmp_version_desc)
		"rated":
			rows.sort_custom(_cmp_rating_desc)
		_:
			rows.sort_custom(_cmp_sessions_desc)
	var result: Array[Dictionary] = rows.slice(0, 10)
	for i in result.size():
		result[i]["rank"] = i + 1
	return result

func _calc_players(gid: String, rec: Variant) -> int:
	var sessions: int = 0
	if rec is Dictionary:
		sessions = int((rec as Dictionary).get("sessions", 0))
	return sessions + 10000

func _cmp_version_desc(a: Dictionary, b: Dictionary) -> bool:
	return (a["meta"] as GameMeta).version > (b["meta"] as GameMeta).version

func _cmp_rating_desc(a: Dictionary, b: Dictionary) -> bool:
	return _local_rating((a["meta"] as GameMeta).id) > _local_rating((b["meta"] as GameMeta).id)

func _cmp_sessions_desc(a: Dictionary, b: Dictionary) -> bool:
	return int(a["players"]) > int(b["players"])

func _local_rating(gid: String) -> float:
	var rv: Variant = DB.get_review(gid)
	if rv is Dictionary:
		return float((rv as Dictionary).get("stars", 0))
	return 0.0
