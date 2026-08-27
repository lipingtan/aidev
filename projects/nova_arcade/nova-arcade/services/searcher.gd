class_name Searcher extends Node
## 本地搜索索引：构建 + 打分查询（client-design §4.4）
## 非单例；冷启动 Main._ready() → Registry.reload() → Searcher.build_index()

## 单游戏索引项
class IndexEntry:
	var gid: String = ""
	var title_lower: String = ""
	var aliases_lower: Array[String] = []
	var pinyin_full: Array[String] = []
	var pinyin_initials: Array[String] = []
	var tags_lower: Array[String] = []

var _index: Array = []
var _editorial: Dictionary = {}

func _ready() -> void:
	_editorial = Registry.get_editorial()

## 从 Registry.all() 构建内存索引
## A-2：IndexEntry 数组字段必须显式初始化为新数组，避免跨实例共享引用
func build_index() -> void:
	_index.clear()
	_editorial = Registry.get_editorial()
	for meta in Registry.all():
		var e := IndexEntry.new()
		e.gid = (meta as GameMeta).id
		e.title_lower = (meta as GameMeta).title.to_lower()
		e.aliases_lower = []
		e.pinyin_full = []
		e.pinyin_initials = []
		e.tags_lower = []
		for alias in (meta as GameMeta).aliases:
			e.aliases_lower.append(str(alias).to_lower())
		for pf in (meta as GameMeta).pinyin:
			var p := str(pf).to_lower()
			e.pinyin_full.append(p)
			e.pinyin_initials.append(_extract_initials(p))
		for tag in (meta as GameMeta).tags:
			e.tags_lower.append(str(tag).to_lower())
		_index.append(e)

## 从全拼串提取首字母（M1：取整串第一个字符；M2 改为完整缩写）
func _extract_initials(pf: String) -> String:
	if pf == "":
		return ""
	return pf.left(1)

## A-6：_sort_scores 仅在 query() 同步调用期间有效，禁止在 query() 内引入 await
var _sort_scores: Dictionary = {}

## 打分查询（同步返回，最多 20 条）；调用方负责 150ms 去抖
## A-1：sort_custom 使用具名比较器 _cmp_score_desc，禁止单行 lambda
func query(text: String) -> Array[String]:
	if text == "":
		return []
	var q := text.to_lower().strip_edges()
	var scores: Dictionary = {}
	for e in _index:
		var entry := e as IndexEntry
		var score: int = _score_entry(entry, q)
		if score > 0:
			scores[entry.gid] = score
	_sort_scores = scores
	var gids: Array[String] = []
	for gid in scores:
		gids.append(gid)
	gids.sort_custom(_cmp_score_desc)
	_sort_scores = {}
	return gids.slice(0, 20)

## A-1：具名比较器，替代单行 lambda
func _cmp_score_desc(a: String, b: String) -> bool:
	return int(_sort_scores.get(a, 0)) > int(_sort_scores.get(b, 0))

## 单游戏打分
func _score_entry(e: IndexEntry, q: String) -> int:
	var s: int = 0
	## 标题前缀 +100；标题包含 +40（互斥）
	if e.title_lower.begins_with(q):
		s += 100
	elif e.title_lower.contains(q):
		s += 40
	## 别名包含 +35
	for alias in e.aliases_lower:
		if alias.contains(q):
			s += 35
			break
	## 拼音全拼前缀 +30
	for pf in e.pinyin_full:
		if pf.begins_with(q):
			s += 30
			break
	## 拼音首字母前缀 +25
	for pi in e.pinyin_initials:
		if pi.begins_with(q):
			s += 25
			break
	## 标签命中 +15
	for tag in e.tags_lower:
		if tag.contains(q):
			s += 15
			break
	return s

## 热搜词（editorial.json hot_queries，最多 6 条）
func hot_queries() -> Array[String]:
	var hq: Variant = _editorial.get("hot_queries", [])
	if not (hq is Array):
		return []
	var out: Array[String] = []
	for item in hq as Array:
		out.append(str(item))
		if out.size() >= 6:
			break
	return out
