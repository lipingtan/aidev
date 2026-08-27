class_name Recommender extends Node
## 推荐服务：继续游戏行 + 为你推荐
## 非单例；挂 Main/Services 节点下，由 home.gd 通过 get_node 取引用。

# ── 公开 API ──────────────────────────────────────────────

## 继续游戏行
## 读 DB.list_records()，过滤 last_played > 0，按 last_played 倒序，返回 top5
## 返回格式：[{"meta": GameMeta, "record": Dictionary}, ...]
func continue_row() -> Array[Dictionary]:
	var rows: Array[Dictionary] = []
	# 遍历所有本地记录，筛选曾经游玩过的条目
	for gid: String in DB.list_records():
		var rec: Variant = DB.get_record(gid)
		if rec == null:
			continue
		var r := rec as Dictionary
		if int(r.get("last_played", 0)) > 0:
			# 查 Registry，meta 缺失则跳过（避免孤儿记录污染结果）
			var meta: Variant = Registry.lookup(gid)
			if meta != null:
				rows.append({"meta": meta as GameMeta, "record": r})
	# 按 last_played 倒序
	rows.sort_custom(func(a: Dictionary, b: Dictionary) -> bool:
		return int(a["record"]["last_played"]) > int(b["record"]["last_played"])
	)
	# 最多返回 5 条
	return rows.slice(0, 5)


## 为你推荐
## 全量游戏按 title 排序，扣除 continue_row 已展示的 gid
## record 字段允许为 null（DB.get_record 返回 null 时照常追加，调用方自行处理）
func for_you() -> Array[Dictionary]:
	# 先取继续游戏行，收集已展示 gid
	var cont_gids: Array[String] = []
	for item: Dictionary in continue_row():
		cont_gids.append((item["meta"] as GameMeta).id)

	# 全量 meta 按 title 排序
	var all_metas: Array[GameMeta] = Registry.all()
	all_metas.sort_custom(func(a: GameMeta, b: GameMeta) -> bool:
		return a.title.to_lower() < b.title.to_lower()
	)

	# 扣除已展示条目，组装结果
	var result: Array[Dictionary] = []
	for meta: GameMeta in all_metas:
		if not cont_gids.has(meta.id):
			result.append({"meta": meta, "record": DB.get_record(meta.id)})
	return result

# ── 热门榜（CR-5 FR-3）──────────────────────────────────────

## 热门榜（mode: "all"=综合 / "new"=新游 / "rated"=好评；未知 mode 按 "all"）
## M1 单款游戏时三模式结果相同；M2 多游戏后自动生效
## 返回 [{"meta": GameMeta, "record": Variant, "rank": int, "players": int}]，最多 10 条
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

## 玩过人数：本地 sessions + 1万基准（Q6；M3 获取云端数据后叠加）
func _calc_players(gid: String, rec: Variant) -> int:
	var sessions: int = 0
	if rec is Dictionary:
		sessions = int((rec as Dictionary).get("sessions", 0))
	return sessions + 10000

## 比较器：version 字符串倒序（M2 接入真实 version 后需复核排序语义：空 version 当前排最前）
func _cmp_version_desc(a: Dictionary, b: Dictionary) -> bool:
	return (a["meta"] as GameMeta).version > (b["meta"] as GameMeta).version

## 比较器：本地评价均分倒序（无评价=0，排末尾）
func _cmp_rating_desc(a: Dictionary, b: Dictionary) -> bool:
	return _local_rating((a["meta"] as GameMeta).id) > _local_rating((b["meta"] as GameMeta).id)

## 比较器：sessions（players）倒序
func _cmp_sessions_desc(a: Dictionary, b: Dictionary) -> bool:
	return int(a["players"]) > int(b["players"])

## 本地评价均分（0~5.0；无评价返回 0.0）
func _local_rating(gid: String) -> float:
	var rv: Variant = DB.get_review(gid)
	if rv is Dictionary:
		return float((rv as Dictionary).get("stars", 0))
	return 0.0
