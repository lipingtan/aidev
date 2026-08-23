extends Node
## T3 DB 回归测试场景：headless 运行 res://tools/test_db.tscn，全过 exit 0。
##
## 用法（APPDATA 指向临时目录；按顺序跑三遍验证持久化/退出 flush/损坏兜底）：
##   Run1: Godot --headless --path <proj> res://tools/test_db.tscn            # normal：写入 + 留挂起防抖写
##   Run2: ... res://tools/test_db.tscn -- verify                             # 验证持久化 + WILL_EXIT flush，末尾损坏 profile.json
##   Run3: ... res://tools/test_db.tscn -- corrupt-check                      # 验证损坏兜底（.corrupt 备份 + 默认值重建）

var _failures: int = 0
## records_updated 信号捕获
var _signal_gid: String = ""

func _ready() -> void:
	EventBus.records_updated.connect(_on_records_updated)
	var args := OS.get_cmdline_user_args()
	var mode := "normal"
	if not args.is_empty():
		mode = str(args[0])
	match mode:
		"verify":
			_run_verify()
		"corrupt-check":
			_run_corrupt_check()
		_:
			_run_normal()

func _on_records_updated(gid: String) -> void:
	_signal_gid = gid

func _check(cond: bool, label: String) -> void:
	if cond:
		print("[test_db] PASS ", label)
	else:
		_failures += 1
		push_error("[test_db] FAIL ", label)

## 读 user://db/ 下文件文本（不存在返回 ""）
func _disk_text(file_name: String) -> String:
	var path := "user://db/" + file_name
	if not FileAccess.file_exists(path):
		return ""
	var f := FileAccess.open(path, FileAccess.READ)
	var t := f.get_as_text()
	f.close()
	return t

## Run1：强写/防抖/信号全量测试，末尾留挂起防抖写验证 WILL_EXIT flush
func _run_normal() -> void:
	# 强写：upsert_record 立即落盘 + records_updated 信号
	DB.upsert_record("tetra_nova", {"total_playtime": 120, "best": 5000})
	var rec := _disk_text("records.json")
	_check(rec.find("tetra_nova") >= 0 and rec.find("120") >= 0, "强写 records 立即落盘")
	_check(_signal_gid == "tetra_nova", "records_updated 信号")
	# 强写：订单（自动生成 id）
	DB.put_order({"amount": 6, "item": "tetra_nova_full"})
	_check(_disk_text("orders.json").find("ord_") >= 0, "强写 orders 立即落盘")
	# save_profile force=true 立即落盘
	DB.save_profile({"theme": "dark"}, true)
	_check(_disk_text("profile.json").find("dark") >= 0, "save_profile force 立即落盘")
	# 防抖：500ms 内三次调用合并为一次落盘（内容含全部补丁）
	DB.save_profile({"a": 1})
	DB.save_profile({"b": 2})
	DB.save_profile({"c": 3})
	await get_tree().create_timer(1.0).timeout
	var pf := _disk_text("profile.json")
	_check(pf.find('"a"') >= 0 and pf.find('"b"') >= 0 and pf.find('"c"') >= 0, "防抖合并落盘")
	# 搜索历史：去重置顶 + 防抖落盘
	DB.add_search_history("tetris")
	DB.add_search_history("puzzle")
	DB.add_search_history("tetris")
	var hist := DB.get_search_history()
	_check(hist.size() == 2 and hist[0] == "tetris", "搜索历史去重置顶")
	await get_tree().create_timer(1.0).timeout
	_check(_disk_text("search_history.json").find("puzzle") >= 0, "搜索历史防抖落盘")
	# 强写：成就
	DB.unlock("tn_wave10")
	_check(_disk_text("achievements.json").find("tn_wave10") >= 0, "强写 achievements")
	# 末尾留挂起防抖写（WILL_EXIT 应 flush；Run2 verify 验证）
	DB.save_profile({"x": 99})
	if _failures == 0:
		print("[test_db] ALL PASS (normal)")
	get_tree().quit(1 if _failures > 0 else 0)

## Run2：验证跨进程持久化 + WILL_EXIT flush；末尾损坏 profile.json（Run3 验证兜底）
func _run_verify() -> void:
	var raw_rec: Variant = DB.get_record("tetra_nova")
	_check(raw_rec != null and int((raw_rec as Dictionary).get("total_playtime", -1)) == 120, "kill-重启后 records 恢复")
	var pf := DB.get_profile()
	_check(int(pf.get("x", -1)) == 99, "WILL_EXIT flush 挂起防抖写")
	_check(DB.list_orders().size() >= 1, "kill-重启后 orders 恢复")
	_check(DB.get_search_history().has("puzzle"), "kill-重启后 search_history 恢复")
	_check(DB.get_achievements().has("tn_wave10"), "kill-重启后 achievements 恢复")
	if _failures == 0:
		print("[test_db] ALL PASS (verify)")
	# 损坏 profile.json（Run3 corrupt-check 验证兜底）
	var f := FileAccess.open("user://db/profile.json", FileAccess.WRITE)
	f.store_string("{ this is CORRUPTED !!!")
	f.close()
	get_tree().quit(1 if _failures > 0 else 0)

## Run3：损坏兜底——不崩溃、.corrupt 备份、默认值重建
func _run_corrupt_check() -> void:
	_check(DB.get_profile().is_empty(), "损坏 profile 默认值重建")
	var found_backup := false
	var dir := DirAccess.open("user://db")
	if dir != null:
		dir.list_dir_begin()
		var n := dir.get_next()
		while n != "":
			if n.begins_with("profile.json.corrupt"):
				found_backup = true
			n = dir.get_next()
		dir.list_dir_end()
	_check(found_backup, ".corrupt 备份存在")
	# 其余数据不受影响
	var raw_rec: Variant = DB.get_record("tetra_nova")
	_check(raw_rec != null and int((raw_rec as Dictionary).get("best", -1)) == 5000, "损坏不影响其他文件")
	if _failures == 0:
		print("[test_db] ALL PASS (corrupt-check)")
	get_tree().quit(1 if _failures > 0 else 0)
