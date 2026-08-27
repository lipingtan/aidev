class_name DBIO extends RefCounted
## JSON 持久化原语（CR-2 T1 自 db.gd 拆出；非 Autoload，由 DB 层持有）
## - read_json / write_json / .corrupt 回退 / user://db/ 目录初始化
## - debounce 调度：同文件窗口内只写一次；实际写数据映射留在 DB 层（writer Callable）
## - flush_named / flush_all 供 DB 生命周期钩子调用（WILL_EXIT/_exit_tree，RG-5）
##
## 依赖：无（纯原语；SceneTree 仅用于 create_timer 调度 debounce）

## 持久化目录
const DB_DIR := "user://db/"
## 防抖窗口（毫秒）
const DEBOUNCE_MS: int = 500

var _tree: SceneTree = null
# === 防抖挂起标记：file_name -> true（窗口内同文件只写一次）===
var _pending: Dictionary = {}
# === file_name -> 实际落盘 Callable（DB 层注入，持有数据映射）===
var _writers: Dictionary = {}

## DB._ready 时绑定 SceneTree 并确保目录存在
func bind(tree: SceneTree) -> void:
	_tree = tree
	_ensure_dir()

## 确保 user://db/ 存在
func _ensure_dir() -> void:
	if not DirAccess.dir_exists_absolute(DB_DIR):
		DirAccess.make_dir_recursive_absolute(DB_DIR)

## 读 JSON；缺失返回默认值；解析失败 → .corrupt 备份 + 默认值重建
func read_json(file_name: String, defaults: Dictionary) -> Dictionary:
	var path := DB_DIR + file_name
	if not FileAccess.file_exists(path):
		return defaults.duplicate(true)
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("DBIO: 无法读取 %s" % path)
		return defaults.duplicate(true)
	var text := file.get_as_text()
	file.close()
	var parsed: Variant = JSON.parse_string(text)
	if parsed is Dictionary:
		return parsed as Dictionary
	_backup_corrupt(file_name)
	push_warning("DBIO: %s 解析失败，已备份 .corrupt 并用默认值重建" % path)
	return defaults.duplicate(true)

## 写 JSON（覆盖）；失败 push_error
func write_json(file_name: String, data: Dictionary) -> void:
	var file := FileAccess.open(DB_DIR + file_name, FileAccess.WRITE)
	if file == null:
		push_error("DBIO: 无法写入 %s" % (DB_DIR + file_name))
		return
	file.store_string(JSON.stringify(data, "\t"))
	file.close()

## 坏文件改名 .corrupt 备份（保留现场便于排查；已存在则加时间戳后缀）
func _backup_corrupt(file_name: String) -> void:
	var dir := DirAccess.open(DB_DIR)
	if dir == null:
		return
	var bad := file_name + ".corrupt"
	if dir.file_exists(bad):
		bad = file_name + ".corrupt_%d" % int(Time.get_unix_time_from_system())
	dir.rename(file_name, bad)

## 挂起一次防抖落盘（窗口内同文件多次调用只写一次）
func schedule_debounce(file_name: String, writer: Callable) -> void:
	if _tree == null or not writer.is_valid():
		return
	if _pending.has(file_name):
		return
	_pending[file_name] = true
	_writers[file_name] = writer
	var timer := _tree.create_timer(DEBOUNCE_MS / 1000.0)
	# SceneTreeTimer 为 RefCounted，触发后自动释放（4.5 无 queue_free；A-3 无累积问题）
	timer.timeout.connect(_on_debounce_timeout.bind(file_name))

## 抖窗到期回调：落盘对应文件
func _on_debounce_timeout(file_name: String) -> void:
	flush_named(file_name)

## 按文件名落盘对应数据（防抖到期 / WILL_EXIT 调用）；无挂起则 no-op
func flush_named(file_name: String) -> void:
	if not _pending.has(file_name):
		return
	_pending.erase(file_name)
	var w: Callable = _writers.get(file_name, Callable())
	if w.is_valid():
		w.call()

## 全量落盘全部挂起项（DB 生命周期钩子调用）
func flush_all() -> void:
	for name in _pending.keys():
		flush_named(str(name))

## 是否有挂起的防抖写
func has_pending() -> bool:
	return not _pending.is_empty()
