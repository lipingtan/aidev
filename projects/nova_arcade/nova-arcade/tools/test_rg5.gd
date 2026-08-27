extends Node
## T10 RG-5 强写持久化（双相）：write 写盘后退出 → verify 重启读回断言。
## 用法：-s res://tools/test_rg5.gd --rg5-phase=write | --rg5-phase=verify
## RG-5 语义：强写（trial_used/orders/playtime 等）强杀/重启不丢；正常退出时防抖缓冲已 flush。

var _failures := 0

func _ready() -> void:
	var phase := ""
	for a in OS.get_cmdline_args():
		if a.begins_with("--rg5-phase="):
			phase = a.substr(12)
	print("[rg5] phase=", phase)
	match phase:
		"write": _write()
		"verify": _verify()
		_:
			push_error("[rg5] 缺 --rg5-phase (write|verify)")
			_failures += 1
	get_tree().quit(1 if _failures > 0 else 0)

func _write() -> void:
	DB.upsert_record("tetra_nova", {"trial_used":5,"total_playtime":1000,"best":999,"finish_count":3})
	DB.put_order({"id":"ord_test","gid":"tetra_nova","amount_cents":600,"status":"paid"})
	DB.add_search_history("tetra")
	DB.save_profile({"theme":"neon"}, true)          # force 强写
	DB.add_search_history("hello")                    # 触发一次防抖写
	await get_tree().create_timer(0.6).timeout        # 等 500ms 防抖窗口到期 → _exit_tree flush
	print("[rg5] 强写+防抖已落盘并 flush，退出")
	get_tree().quit(0)

func _verify() -> void:
	var rec: Variant = DB.get_record("tetra_nova")
	_check(rec is Dictionary and int(rec.get("trial_used",-1)) == 5, "RG-5 trial_used=5 重启不丢")
	_check(rec is Dictionary and int(rec.get("total_playtime",-1)) == 1000, "RG-5 total_playtime=1000 重启不丢")
	_check(rec is Dictionary and int(rec.get("best",-1)) == 999, "RG-5 best=999 重启不丢")
	_check(list_orders().get("ord_test", {}).get("status") == "paid", "RG-5 订单 ord_test 重启不丢")
	var sh: Array[String] = DB.get_search_history()
	_check(sh.has("hello") and sh.has("tetra"), "RG-5 防抖写搜索历史重启不丢（正常退出已 flush）")
	_check(str(DB.get_profile().get("theme","")) == "neon", "RG-5 profile 重启不丢")

func _check(cond: bool, label: String) -> void:
	if cond:
		print("[rg5] PASS ", label)
	else:
		_failures += 1
		push_error("[rg5] FAIL ", label)

func list_orders() -> Dictionary:
	return DB.list_orders()
