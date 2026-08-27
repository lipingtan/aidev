extends Node
## 本地持久化单例（design.md §3.3）——全部本地持久化的唯一出口
##
## Autoload 名 DB（注册顺序第 3），不声明 class_name（godot-engine §六）。
## - JSON 文件落 user://db/；两级写入：强写立即落盘 / 防抖写 500ms 合并
## - 强写：trial_used、orders、total_playtime/best/finish_count（records）、reviews、achievements
## - 防抖写：search_history、daily、profile（窗口内多次调用只落盘一次）
## - save_profile(patch, force := false)：默认防抖；主题切换等关键项 force=true 立即落盘（Review B-1）
## - 损坏兜底 / 防抖调度 / 目录初始化 → DBIO（core/db_io.gd，CR-2 T1 拆出）
## - WILL_EXIT 强制 flush 全部防抖缓冲（RG-5；异常杀进程仍可能丢最后一次防抖写，强写不受影响）
##
## 依赖：
## - EventBus: upsert_record 后发 records_updated(gid)

## 搜索历史上限条数
const HISTORY_CAP: int = 20

var _io := DBIO.new()
# === 内存缓存（_ready 时从 JSON 载入）===
var _profile: Dictionary = {}
var _records: Dictionary = {}
var _reviews: Dictionary = {}
var _orders: Dictionary = {}
var _search_history: Array[String] = []
var _daily: Dictionary = {}
var _achievements: Dictionary = {}

func _ready() -> void:
	_io.bind(get_tree())
	_profile = _io.read_json("profile.json", {})
	_records = _io.read_json("records.json", {})
	_reviews = _io.read_json("reviews.json", {})
	_orders = _io.read_json("orders.json", {})
	var sh := _io.read_json("search_history.json", {"items": []})
	if (sh.get("items") as Variant) is Array:
		for s in sh["items"] as Array:
			_search_history.append(str(s))
	_daily = _io.read_json("daily.json", {})
	_achievements = _io.read_json("achievements.json", {})

## 退出（含 WILL_EXIT）强制 flush 全部防抖缓冲
func _notification(what: int) -> void:
	if what == NOTIFICATION_WM_CLOSE_REQUEST and _io.has_pending():
		_io.flush_all()

## 应用退出时 autoload 移出场景树 → 强制 flush（headless 下 WM_CLOSE_REQUEST 不触发，此钩子兜底）
func _exit_tree() -> void:
	if _io.has_pending():
		_io.flush_all()

# === profile（防抖写，可 force 强写）===

## 取用户资料快照
func get_profile() -> Dictionary:
	return _profile

## 合并写入资料；force=true 立即落盘，默认 500ms 防抖合并
func save_profile(patch: Dictionary, force: bool = false) -> void:
	for k in patch:
		_profile[k] = patch[k]
	if force:
		_io.write_json("profile.json", _profile)
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
	_io.write_json("records.json", _records)
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
	_io.write_json("reviews.json", _reviews)

## 删除评价并立即落盘；gid 不存在时静默忽略
func delete_review(gid: String) -> void:
	if not _reviews.has(gid):
		return
	_reviews.erase(gid)
	_io.write_json("reviews.json", _reviews)

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
	_io.write_json("orders.json", _orders)

# === search_history / daily（防抖写）===

## 清空搜索历史，防抖落盘
func clear_search_history() -> void:
	_search_history.clear()
	_debounce_write("search_history.json")

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

## 触达每日数据：跨天重置 {date, plays:0}（保留 tasks_data，design §3 共存规则），防抖落盘
func touch_daily() -> void:
	var today := Time.get_date_string_from_system()
	if str(_daily.get("date", "")) != today:
		var kept_tasks: Variant = _daily.get("tasks_data")
		_daily = {"date": today, "plays": 0}
		if kept_tasks is Dictionary:
			_daily["tasks_data"] = kept_tasks
	else:
		_daily["plays"] = int(_daily.get("plays", 0)) + 1
	_debounce_write("daily.json")

## 取每日任务进度；不存在/格式异常时返回空 dict（CR-5 design §3）
func get_daily_tasks() -> Dictionary:
	var raw: Variant = _daily.get("tasks_data")
	return raw if raw is Dictionary else {}

## 写入每日任务进度（防抖写，与 touch_daily 共用 daily.json；确保 date key 存在）
func update_daily_tasks(tasks_data: Dictionary) -> void:
	if not _daily.has("date"):
		_daily["date"] = Time.get_date_string_from_system()
	_daily["tasks_data"] = tasks_data
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
	_io.write_json("achievements.json", _achievements)

# === 防抖（调度在 DBIO；落盘映射在本层）===

## 挂起一次防抖落盘（窗口内同文件多次调用只写一次）
func _debounce_write(file_name: String) -> void:
	_io.schedule_debounce(file_name, _flush_named.bind(file_name))

## 按文件名落盘对应数据（防抖到期 / WILL_EXIT 调用）
func _flush_named(file_name: String) -> void:
	match file_name:
		"profile.json":
			_io.write_json("profile.json", _profile)
		"search_history.json":
			_io.write_json("search_history.json", {"items": _search_history})
		"daily.json":
			_io.write_json("daily.json", _daily)
