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

## 全部 meta
func all() -> Array[GameMeta]:
	var out: Array[GameMeta] = []
	for k in _games:
		out.append(_games[k] as GameMeta)
	return out

## 查询：category 精确 / tag 包含 / price 上限 / min_rating 预留（CR-1 无评分数据）/ sort 排序键
## 注意：本 Godot 4.5 构建解析器不支持单行 lambda，过滤/比较用显式辅助函数
func query(filters: Dictionary) -> Array[GameMeta]:
	var out := all()
	if filters.has("category"):
		out = _filter_category(out, str(filters["category"]))
	if filters.has("tag"):
		out = _filter_tag(out, str(filters["tag"]))
	if filters.has("price"):
		out = _filter_price(out, int(filters["price"]))
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

## 跨游戏查成就定义；不存在返回 null（返回类型用 Variant，理由同 lookup）
func ach_def(aid: String) -> Variant:
	for m in all():
		for a in m.achievements:
			if str(a.get("id", "")) == aid:
				return a
	return null

## 编辑位配置（banner/featured/categories）
func get_editorial() -> Dictionary:
	return _editorial
