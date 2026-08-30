extends Node
## 游戏注册表：启动扫描 games/*/meta.json + data/editorial.json（design.md §3.4）
##
## Autoload 名 Registry（注册顺序第 4），不声明 class_name（godot-engine §六）。
## - reload()：重扫全部 meta；DLC 预留（dlc_installed 信号触发）
## - lookup(gid)（设计稿 get，因 Object.get 冲突改名）/ all() / query({category, price, min_rating, tag, sort})
## - installed_version(gid)：内置=meta.version，未安装返回 ""
## - ach_def(aid)：跨游戏查成就定义（CR-3 成就页用）
## - 加载失败仅 push_warning 不崩溃；坏 meta.json 跳过该目录
##
## 依赖：
## - EventBus: dlc_installed → reload()
## - GameMeta: 元数据 Resource（core/game_meta.gd）

## gid -> GameMeta
var _games: Dictionary = {}
## data/editorial.json（banner/featured/categories，空模板）
var _editorial: Dictionary = {}
## 全局成就缓存（CR-7 T1：data/achievements.json）
var _global_achievements: Dictionary = {}

func _ready() -> void:
	EventBus.dlc_installed.connect(_on_dlc_installed)
	reload()

## DLC 安装完成 → 重扫（CR-1 无 DLC，预留）
func _on_dlc_installed(dlc_id: String) -> void:
	push_warning("Registry: DLC %s 已安装，重载" % dlc_id)
	reload()

## 重扫 games/*/meta.json + data/editorial.json
func reload() -> void:
	_games.clear()
	var dir := DirAccess.open("res://games")
	if dir == null:
		push_warning("Registry: res://games 不存在，跳过扫描")
	else:
		dir.list_dir_begin()
		var name := dir.get_next()
		while name != "":
			if dir.current_is_dir() and not name.begins_with("."):
				_load_meta("res://games/%s/meta.json" % name)
			name = dir.get_next()
		dir.list_dir_end()
	_load_editorial()
	_load_global_achievements()

## 加载全局成就定义（data/achievements.json）；缺失/损坏用空缓存（A-2修复）
func _load_global_achievements() -> void:
	_global_achievements.clear()
	var path := "res://data/achievements.json"
	if not FileAccess.file_exists(path):
		return
	var raw := FileAccess.get_file_as_string(path)
	var ja: Variant = JSON.parse_string(raw)
	if ja is Array:
		for ga in ja:
			if ga is Dictionary and ga.has("id"):
				_global_achievements[str(ga["id"])] = ga
			else:
				push_warning("Registry: 全局成就条目缺 id 字段，跳过", ga)
	elif ja != null:
		push_warning("Registry: achievements.json 非 Array 格式（%s），保留旧缓存" % type_string(ja))

## 加载单个 meta.json；解析失败仅告警跳过（不阻断启动）
func _load_meta(path: String) -> void:
	if not FileAccess.file_exists(path):
		return
	var file := FileAccess.open(path, FileAccess.READ)
	var text := file.get_as_text()
	file.close()
	var parsed: Variant = JSON.parse_string(text)
	if not (parsed is Dictionary):
		push_warning("Registry: meta.json 解析失败，跳过 %s" % path)
		return
	var meta := GameMeta.from_dict(parsed as Dictionary)
	if meta.id == "":
		meta.id = path.get_base_dir().get_file()
	_games[meta.id] = meta

## 读取 data/editorial.json；缺失/损坏用空模板
func _load_editorial() -> void:
	_editorial = {}
	var path := "res://data/editorial.json"
	if not FileAccess.file_exists(path):
		return
	var file := FileAccess.open(path, FileAccess.READ)
	var text := file.get_as_text()
	file.close()
	var parsed: Variant = JSON.parse_string(text)
	if parsed is Dictionary:
		_editorial = parsed as Dictionary
	else:
		push_warning("Registry: editorial.json 解析失败，用空模板")

## 按 gid 取 meta；不存在返回 null。
## 命名偏差：设计稿为 get(gid)，但 Node 基类已有 Object.get(StringName)，同名会签名冲突 → 改 lookup()。
## 返回类型用 Variant（本构建不允许非空类型注解 + return null）。
func lookup(gid: String) -> Variant:
	var m: GameMeta = _games.get(gid)
	return m if m != null else null

## 相似推荐（Q2=A）：同类目且标签交集，排除自身；无则空数组（标题序）
func similar(gid: String) -> Array[GameMeta]:
	var base: GameMeta = lookup(gid)
	if base == null or base.category == "":
		return []
	var out: Array[GameMeta] = []
	for m in all():
		if m.id == gid or m.category != base.category:
			continue
		if _tags_overlap(m.tags, base.tags):
			out.append(m)
	out.sort_custom(_cmp_title)
	return out

## 两个标签集是否有交集
func _tags_overlap(a: Array[String], b: Array[String]) -> bool:
	for t in a:
		if b.has(t):
			return true
	return false

## 全部 meta
func all() -> Array[GameMeta]:
	var out: Array[GameMeta] = []
	for k in _games:
		out.append(_games[k] as GameMeta)
	return out

## 查询：category 精确（空字符串跳过）/ tag 包含 / price 上限
## 新增：price_models 多选 / min_rating M1 占位 / is_new M1 占位 / sort 排序键
## 向后兼容：不传新 key 时行为不变
func query(filters: Dictionary) -> Array[GameMeta]:
	var out := all()
	if filters.has("category") and str(filters["category"]) != "":
		out = _filter_category(out, str(filters["category"]))
	if filters.has("tag"):
		out = _filter_tag(out, str(filters["tag"]))
	if filters.has("price"):
		out = _filter_price(out, int(filters["price"]))
	if filters.has("price_models"):
		out = _filter_price_models(out, filters["price_models"] as Array)
	if filters.has("min_rating"):
		out = _filter_min_rating(out, float(filters["min_rating"]))
	if filters.has("is_new") and bool(filters["is_new"]):
		out = _filter_is_new(out)
	var sort := str(filters.get("sort", "title"))
	match sort:
		"price":
			out.sort_custom(_cmp_price)
		"version":
			out.sort_custom(_cmp_version)
		_:
			out.sort_custom(_cmp_title)
	return out

## 按分类精确过滤
func _filter_category(list: Array[GameMeta], cat: String) -> Array[GameMeta]:
	var out: Array[GameMeta] = []
	for m in list:
		if m.category == cat:
			out.append(m)
	return out

## 按标签包含过滤
func _filter_tag(list: Array[GameMeta], tag: String) -> Array[GameMeta]:
	var out: Array[GameMeta] = []
	for m in list:
		if m.tags.has(tag):
			out.append(m)
	return out

## 按价格上限过滤
func _filter_price(list: Array[GameMeta], max_price: int) -> Array[GameMeta]:
	var out: Array[GameMeta] = []
	for m in list:
		if m.price <= max_price:
			out.append(m)
	return out

## 按 price_model 多选过滤；models 为空时直接返回（不过滤）
func _filter_price_models(list: Array[GameMeta], models: Array) -> Array[GameMeta]:
	if models.is_empty():
		return list
	var out: Array[GameMeta] = []
	for m in list:
		if models.has(m.price_model):
			out.append(m)
	return out

## 按最低评分过滤；M1 占位：GameMeta 无 rating 字段时直接透传
func _filter_min_rating(list: Array[GameMeta], _min_r: float) -> Array[GameMeta]:
	# M1 GameMeta 无 rating，占位透传；M2 补充 rating 字段后再实现
	return list

## 按"新上架"过滤；M1 占位：直接透传不过滤
func _filter_is_new(list: Array[GameMeta]) -> Array[GameMeta]:
	return list

## 比较器：标题（本构建 String 无 natural_compare，用大小写不敏感字典序）
func _cmp_title(a: GameMeta, b: GameMeta) -> bool:
	return a.title.to_lower() < b.title.to_lower()

## 比较器：价格
func _cmp_price(a: GameMeta, b: GameMeta) -> bool:
	return a.price < b.price

## 比较器：版本
func _cmp_version(a: GameMeta, b: GameMeta) -> bool:
	return a.version < b.version

## 已安装版本：内置游戏=meta.version，未安装返回 ""
func installed_version(gid: String) -> String:
	var m: GameMeta = lookup(gid)
	return m.version if m != null else ""

## 跨游戏查成就定义：先全局缓存，再遍历游戏级（CR-7 T1）；不存在返回 null
func ach_def(aid: String) -> Variant:
	if _global_achievements.has(aid):
		return _global_achievements[aid]
	for m in all():
		for a in m.achievements:
			if str(a.get("id", "")) == aid:
				return a
	return null

## 编辑位配置（banner/featured/categories）
func get_editorial() -> Dictionary:
	return _editorial
