extends Node
## 本地持久化单例（design.md §3.3）——全部本地持久化的唯一出口
##
## Autoload 名 DB（注册顺序第 3），不声明 class_name（godot-engine §六）。
## - JSON 文件落 user://db/；两级写入：强写立即落盘 / 防抖写 500ms 合并
## - 强写：trial_used、orders、total_playtime/best/finish_count（records）、reviews、achievements
## - 防抖写：search_history、daily、profile（窗口内多次调用只落盘一次）
## - save_profile(patch, force := false)：默认防抖；主题切换等关键项 force=true 立即落盘（Review B-1）
## - 损坏兜底：解析失败 → 坏文件改名 .corrupt 备份 + 默认值重建 + push_warning，不崩溃
## - WILL_EXIT 强制 flush 全部防抖缓冲（RG-5；异常杀进程仍可能丢最后一次防抖写，强写不受影响）
##
## 依赖：
## - EventBus: upsert_record 后发 records_updated(gid)

## 持久化目录
const DB_DIR := "user://db/"
## 防抖窗口（毫秒）
const DEBOUNCE_MS: int = 500
## 搜索历史上限条数
const HISTORY_CAP: int = 20

## 用户资料（主题/设置等，防抖写）
var _profile: Dictionary = {}
## gid -> {total_playtime, best, finish_count, trial_used...}（强写）
var _records: Dictionary = {}
## gid -> 评价 dict（强写）
var _reviews: Dictionary = {}
## order_id -> 订单 dict（强写）
var _orders: Dictionary = {}
## 搜索历史（防抖写，新→旧）
var _search_history: Array[String] = []
## 每日数据 {date, plays...}（防抖写）
var _daily: Dictionary = {}
## 成就 id -> {unlocked_at}（强写）
var _achievements: Dictionary = {}

## 防抖挂起标记：file_name -> true
var _pending: Dictionary = {}

func _ready() -> void:
	_ensure_dir()
	_profile = _read_json("profile.json", {})
	_records = _read_json("records.json", {})
	_reviews = _read_json("reviews.json", {})
	_orders = _read_json("orders.json", {})
	var sh := _read_json("search_history.json", {"items": []})
	if (sh.get("items") as Variant) is Array:
		for s in sh["items"] as Array:
			_search_history.append(str(s))
	_daily = _read_json("daily.json", {})
	_achievements = _read_json("achievements.json", {})

## 退出（含 WILL_EXIT）强制 flush 全部防抖缓冲
func _notification(what: int) -> void:
	if what == NOTIFICATION_WM_CLOSE_REQUEST and not _pending.is_empty():
		for name in _pending.keys():
			_flush_named(str(name))

## 应用退出时 autoload 移出场景树 → 强制 flush（headless 下 WM_CLOSE_REQUEST 不触发，此钩子兜底）
func _exit_tree() -> void:
	if not _pending.is_empty():
		for name in _pending.keys():
			_flush_named(str(name))

# === profile（防抖写，可 force 强写）===

## 取用户资料快照
func get_profile() -> Dictionary:
	return _profile

## 合并写入资料；force=true 立即落盘，默认 500ms 防抖合并
func save_profile(patch: Dictionary, force: bool = false) -> void:
	for k in patch:
		_profile[k] = patch[k]
	if force:
		_write_json("profile.json", _profile)
	else:
		_debounce_write("profile.json")

# === records（强写）===

## 取单游戏记录；不存在返回 null
func get_record(gid: String) -> Variant:
	var raw: Variant = _records.get(gid)
	return raw if raw is Dictionary else null

## 合并更新游戏记录并立即落盘，发 EventBus.records_updated(gid)
func upsert_record(gid: String, patch: Dictionary) -> void:
	var raw: Variant = _records.get(gid)
	var r: Dictionary = raw if raw is Dictionary else {}
	for k in patch:
		r[k] = patch[k]
	_records[gid] = r
	_write_json("records.json", _records)
	EventBus.records_updated.emit(gid)

## 全部游戏记录（gid -> dict）
func list_records() -> Dictionary:
	return _records

# === reviews / orders（强写）===

## 取评价；不存在返回 null
func get_review(gid: String) -> Variant:
	var raw: Variant = _reviews.get(gid)
	return raw if raw is Dictionary else null

## 写入评价并立即落盘
func put_review(gid: String, review: Dictionary) -> void:
	_reviews[gid] = review
	_write_json("reviews.json", _reviews)

## 全部订单（order_id -> dict）
func list_orders() -> Dictionary:
	return _orders

## 写入订单并立即落盘（以 order["id"] 为键，缺省自动生成）
func put_order(order: Dictionary) -> void:
	var oid := str(order.get("id", ""))
	if oid == "":
		oid = "ord_%d" % int(Time.get_unix_time_from_system())
		order["id"] = oid
	_orders[oid] = order
	_write_json("orders.json", _orders)

# === search_history / daily（防抖写）===

## 追加搜索词（去重置顶，超上限截断），防抖落盘
func add_search_history(q: String) -> void:
	if q == "":
		return
	if _search_history.has(q):
		_search_history.erase(q)
	_search_history.push_front(q)
	while _search_history.size() > HISTORY_CAP:
		_search_history.pop_back()
	_debounce_write("search_history.json")

## 取搜索历史（新→旧）
func get_search_history() -> Array[String]:
	return _search_history

## 取每日数据
func get_daily() -> Dictionary:
	return _daily

## 触达每日数据：跨天重置 {date, plays:0}，防抖落盘
func touch_daily() -> void:
	var today := Time.get_date_string_from_system()
	if str(_daily.get("date", "")) != today:
		_daily = {"date": today, "plays": 0}
	else:
		_daily["plays"] = int(_daily.get("plays", 0)) + 1
	_debounce_write("daily.json")

# === achievements（强写）===

## 全部成就解锁记录
func get_achievements() -> Dictionary:
	return _achievements

## 解锁成就并立即落盘
func unlock(aid: String) -> void:
	if _achievements.has(aid):
		return
	_achievements[aid] = {"unlocked_at": int(Time.get_unix_time_from_system())}
	_write_json("achievements.json", _achievements)

# === 防抖与读写底层 ===

## 挂起一次防抖落盘（窗口内同文件多次调用只写一次）
func _debounce_write(file_name: String) -> void:
	if _pending.has(file_name):
		return
	_pending[file_name] = true
	var timer := get_tree().create_timer(DEBOUNCE_MS / 1000.0)
	timer.timeout.connect(_on_debounce_timeout.bind(timer, file_name))

## 抖窗到期回调：释放 Timer 后落盘（A-3 修复：避免死 Timer 长会话累积）
func _on_debounce_timeout(timer: Timer, file_name: String) -> void:
	timer.queue_free()
	_flush_named(file_name)

## 按文件名落盘对应数据（防抖到期 / WILL_EXIT 调用）
func _flush_named(file_name: String) -> void:
	_pending.erase(file_name)
	match file_name:
		"profile.json":
			_write_json("profile.json", _profile)
		"search_history.json":
			_write_json("search_history.json", {"items": _search_history})
		"daily.json":
			_write_json("daily.json", _daily)

## 确保 user://db/ 存在
func _ensure_dir() -> void:
	if not DirAccess.dir_exists_absolute(DB_DIR):
		DirAccess.make_dir_recursive_absolute(DB_DIR)

## 读 JSON；缺失返回默认值；解析失败 → .corrupt 备份 + 默认值重建
func _read_json(file_name: String, defaults: Dictionary) -> Dictionary:
	var path := DB_DIR + file_name
	if not FileAccess.file_exists(path):
		return defaults.duplicate(true)
	var file := FileAccess.open(path, FileAccess.READ)
	var text := file.get_as_text()
	file.close()
	var parsed: Variant = JSON.parse_string(text)
	if parsed is Dictionary:
		return parsed as Dictionary
	_backup_corrupt(file_name)
	push_warning("DB: %s 解析失败，已备份 .corrupt 并用默认值重建" % path)
	return defaults.duplicate(true)

## 坏文件改名 .corrupt 备份（保留现场便于排查）
func _backup_corrupt(file_name: String) -> void:
	var dir := DirAccess.open(DB_DIR)
	if dir == null:
		return
	var bad := file_name + ".corrupt"
	if dir.file_exists(bad):
		bad = file_name + ".corrupt_%d" % int(Time.get_unix_time_from_system())
	dir.rename(file_name, bad)

## 写 JSON（覆盖）；失败 push_error
func _write_json(file_name: String, data: Dictionary) -> void:
	var file := FileAccess.open(DB_DIR + file_name, FileAccess.WRITE)
	if file == null:
		push_error("DB: 无法写入 %s" % (DB_DIR + file_name))
		return
	file.store_string(JSON.stringify(data, "\t"))
	file.close()
