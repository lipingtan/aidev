extends Node
## T5 魔塔测试（CR-6）：headless 实例化 module.tscn（adapter+src）→ GameModule 协议 + 纯函数逻辑。
## 断言 ≥10：boot SAVE_PATH、pause/resume、reset_run、quit_requested 结构、calc_damage/MON/ITEM。

var _failures := 0
var _pass := 0
var _got: Dictionary = {}

func _ready() -> void:
	var scene := load("res://games/magic_tower/module.tscn") as PackedScene
	var mod := scene.instantiate()
	add_child(mod)
	await get_tree().process_frame
	_run(mod)

func _check(cond: bool, label: String) -> void:
	if cond:
		_pass += 1
		print("[mt_test] PASS ", label)
	else:
		_failures += 1
		push_error("[mt_test] FAIL ", label)

func _run(mod: Node) -> void:
	print("=== 魔塔 T5（协议 + 逻辑）===")
	var src := mod.get_node("Src") as Node
	_check(mod is GameModule, "P1 module 是 GameModule")
	_check(src != null and src.has_signal("run_over"), "P2 src 存在且 run_over 信号")
	mod.boot({"save_dir": "user://", "trial_mode": true, "owned": false, "best": 0, "viewport_size": Vector2i(720, 1560)})
	_check(src.get("SAVE_PATH") == "user://save.cfg", "P3 boot 注入 SAVE_PATH")
	var paused_before: bool = src.get("paused")
	mod.pause_game()
	_check(bool(src.get("paused")) != paused_before, "P4 pause_game 转发 set_paused(true)")
	mod.resume_game()
	_check(bool(src.get("paused")) == paused_before, "P5 resume_game 转发 set_paused(false)")
	src.set("hp", 37)
	src.set("gold", 999)
	mod.reset_run()
	_check(int(src.get("hp")) == 100 and int(src.get("gold")) == 0, "P6 reset_run 重置玩家状态")
	_got = {}
	mod.quit_requested.connect(func(r: Dictionary): _got = r)
	mod.quit_to_shell()
	_check(_got.has("score") and _got.has("playtime") and _got.has("achievements") and _got.has("extra"), "P7 quit_requested 四键齐全")
	_check((_got.get("achievements", []) as Array).is_empty(), "P8 achievements 恒空数组（M2 预留）")
	_check(_got.get("extra", {}).has("gold"), "P9 extra.gold 存在（魔塔结算字段）")
	# 纯函数逻辑（魔塔 main.gd 无 class_name → 经实例访问 static/const）
	_check(src.calc_damage(5, 5) == 0, "L1 calc_damage atk<=def → 0")
	_check(int(src.calc_damage(6, 5)) >= 1, "L2 calc_damage atk>def → >=1")
	var mon: Dictionary = src.get("MON")
	var item: Dictionary = src.get("ITEM")
	_check(int(mon["slime"][0]) == 30 and int(item["potion"][1]) == 30, "L3 MON/ITEM 常量表")
	print("[mt_test] 合计 PASS=%d FAIL=%d" % [_pass, _failures])
	get_tree().quit(1 if _failures > 0 else 0)
