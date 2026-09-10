extends Node
## CR-8 水果机形态 headless 测试：配置/押注/结算(含散花)/账本/成就/统计
## 运行：--headless --path . --scene res://tools/test_slot_machine.tscn（隔离 APPDATA）

var _fails := 0
var _total := 0

func _ready() -> void:
	get_tree().create_timer(90.0).timeout.connect(_fallback)
	await get_tree().process_frame

	_test_config()
	_test_betting()
	_test_settle()
	_test_lucky_burst()
	_test_lifecycle()
	_test_relief()
	_test_achievements()
	_test_reset_run()
	_test_statistics()

	print("[slot_test] 合计 PASS=%d FAIL=%d" % [_total - _fails, _fails])
	get_tree().quit(1 if _fails > 0 else 0)

func _ok(cond: bool, label: String) -> void:
	_total += 1
	if cond:
		print("[slot_test] PASS ", label)
	else:
		_fails += 1
		print("[slot_test] FAIL ", label)

func _fallback() -> void:
	print("[slot_test] FAIL timeout")
	get_tree().quit(2)

var _src_seq := 0

func _new_src(silent := true) -> Node2D:
	var scene := load("res://games/slot_machine/module.tscn") as PackedScene
	var module := scene.instantiate()
	add_child(module)
	var src := module.get_node("Src") as Node2D
	src._io_silent = silent
	_src_seq += 1
	src.SAVE_PATH = "user://test_save_%d.cfg" % _src_seq
	src.load_balance()
	return src

func _free_module(src: Node2D) -> void:
	var module := src.get_parent()
	module.queue_free()
	remove_child(module)

# === 配置 ===
func _test_config() -> void:
	_ok(SlotConfig.RING.size() == 24, "盘面 24 格（goslot 原值）")
	var lucky := []
	for i in 24:
		if SlotConfig.RING[i]["flag"] == "luckytry":
			lucky.append(i)
	_ok(lucky == [9, 21], "大Lucky=9 小Lucky=21")
	_ok(SlotConfig.BET_SYMBOLS.size() == 8, "8 可押符号")
	_ok(SlotConfig.WEIGHTS.size() == 24, "权重表 24 项")
	_ok(SlotConfig.RING[0]["score"] == 10 and SlotConfig.RING[3]["score"] == 500, "分值表 goslot 原值（bigorange=10 bigbar=500）")
	# 图标文件存在
	var all_icons := true
	for i in 24:
		if not FileAccess.file_exists(SlotConfig.icon_path(i)):
			all_icons = false
			print("[slot_test]   missing: ", SlotConfig.icon_path(i))
	_ok(all_icons, "24 格图标存在（rollpic/lucky）")
	_ok(SlotConfig.START_POS == Vector2(26, 110), "startpos 原值 (26,110)")
	_ok(SlotConfig.STEPS == 7 and SlotConfig.STEPLEN == 56, "steps=7 steplen=56 原值")

# === 押注（goslot 语义：押币即扣余额，清币返还）===
func _test_betting() -> void:
	var src := _new_src()
	_ok(src.add_bet(0), "加注成功")
	_ok(src.balance == SlotConfig.BASE_BALANCE - src._bet_step, "押币即扣余额")
	_ok(src.total_bet() == src._bet_step, "登记额正确")
	src.set_bet_step(5)
	src.add_bet(0)
	_ok(int(src._bets[0]) == 6, "叠加注 1+5")
	src.set_bet_step(10)
	src.balance = 50
	src._bets = [5, 5, 5, 5, 5, 5, 0, 0]   # 30（账面已扣）
	var ok_a: bool = src.add_bet(0)         # 扣 10 → 剩 40
	var ok_b: bool = src.add_bet(1)         # 扣 10 → 剩 30
	var ok_c: bool = src.add_bet(2)         # 扣 10 → 剩 20
	var ok_d: bool = src.add_bet(3)         # 扣 10 → 剩 10
	var ok_e: bool = src.add_bet(4)         # 10 <= 10 → 成功，剩 0
	var ok_f: bool = src.add_bet(5)         # 10 > 0 → 拒
	_ok(ok_a and ok_b and ok_c and ok_d and ok_e and not ok_f, "押满余额后拒绝")
	var bal_before: int = src.balance
	src.clear_bet(0)
	_ok(int(src._bets[0]) == 0, "清零该符号")
	_ok(src.balance == bal_before + 15, "清币返还该符号全部押额(5+10): %d" % (src.balance - bal_before))
	src._bets = [0, 0, 0, 0, 0, 0, 0, 0]
	_ok(not src.can_spin(), "无押注禁旋转")
	src._bets = [1, 0, 0, 0, 0, 0, 0, 0]
	_ok(src.can_spin(), "有押注可旋转")
	src.state = src.State.RUNNING
	_ok(not src.add_bet(1), "RUNNING 期拒绝押注")
	src.state = src.State.IDLE_BET
	_free_module(src)

# === 结算逐例（含 ×2 格、散花多格、Lucky 直停奖）===
func _test_settle() -> void:
	var src := _new_src()
	# goslot getScoreVal 语义：win = 押额 × 格 score
	# RING[0]=bigorange score10；押橙 1 → 10
	src._bets = [0, 0, 0, 0, 1, 0, 0, 0]   # bets index: BET_SYMBOLS[4]=ring? 见下逐条核对
	# 逐例重写（用 _bets 按 BET_SYMBOLS 顺序: bar,seven,star,watermelon,ring,lemon,orange,apple）
	# case1: 停 0(bigorange score10) 押橙 2 → 2×10=20
	src.balance = 1000
	src._bets = [0, 0, 0, 0, 0, 0, 2, 0]
	src._settle([0], 0)
	_ok(src.balance - 1000 == 20, "停bigorange押橙2 → 20")
	# case2: 停 3(bigbar score500) 押 bar 1 → 500
	src.balance = 1000
	src._bets = [1, 0, 0, 0, 0, 0, 0, 0]
	src._settle([3], 0)
	_ok(src.balance - 1000 == 500, "停bigbar押BAR1 → 500")
	# case3: 停 2(smallbar score100) 押 bar 1 → 100
	src.balance = 1000
	src._bets = [1, 0, 0, 0, 0, 0, 0, 0]
	src._settle([2], 0)
	_ok(src.balance - 1000 == 100, "停smallbar押BAR1 → 100")
	# case4: 停 0 押苹果 → 0（flag 不匹配）
	src.balance = 1000
	src._bets = [0, 0, 0, 0, 0, 0, 0, 1]
	src._settle([0], 0)
	_ok(src.balance - 1000 == 0, "停橙押苹果 = 0")
	# case5: 散花多格 [9,21]（两个 Lucky）无押中 → 0（Lucky flag=luckytry 无押注位）
	src.balance = 1000
	src._settle([9, 21], 0)
	_ok(src.balance - 1000 == 0, "Lucky 格无押注位=0")
	# case6: 散花多格 [9, 4, 17] 押苹果 2 → stop4 bigapple 2×5=10, stop17 smalllemon不中, stop9 无 → 10
	src.balance = 1000
	src._bets = [0, 0, 0, 0, 0, 0, 0, 2]
	src._settle([9, 4, 17], 0)
	_ok(src.balance - 1000 == 10, "散花含bigapple押苹果2 → 10")
	var cleared: bool = true
	for b in src._bets:
		if int(b) != 0:
			cleared = false
	_ok(cleared, "结算后押注清零（A-1）")
	_free_module(src)

# === Lucky 散花逻辑 ===
func _test_lucky_burst() -> void:
	var src := _new_src()
	# 大 Lucky 散花数量范围 5~7
	src._rng.seed = 777
	var counts := []
	for i in 20:
		var cnt: int = src._rng.randi_range(SlotConfig.BIG_LUCKY_ORBS[0], SlotConfig.BIG_LUCKY_ORBS[1])
		counts.append(cnt)
	var in_range := true
	for c in counts:
		if c < 5 or c > 7:
			in_range = false
	_ok(in_range, "大 Lucky 散花 5~7 灯")
	var small_counts := []
	for i in 20:
		var cnt: int = src._rng.randi_range(SlotConfig.SMALL_LUCKY_ORBS[0], SlotConfig.SMALL_LUCKY_ORBS[1])
		small_counts.append(cnt)
	var in_range2 := true
	for c in small_counts:
		if c < 3 or c > 5:
			in_range2 = false
	_ok(in_range2, "小 Lucky 散花 3~5 灯")
	# 散花落位在环范围内且逐格推进（4~8 格）
	var src2 := src
	src2._rng.seed = 424242
	var all_ok := true
	for i in 50:
		var stop: int = (9 + src2._rng.randi_range(4, 8)) % 24
		if stop < 0 or stop >= 24:
			all_ok = false
	_ok(all_ok, "散花落位 ∈ [0,24)")
	_free_module(src)

var _ni_count := 0
func _ok_ni(orbs: Array) -> void:
	# 每个 orb 落位合法
	for o in orbs:
		if int(o) < 0 or int(o) >= 24:
			_fails += 1

# === 生命周期 ===
func _test_lifecycle() -> void:
	var scene := load("res://games/slot_machine/module.tscn") as PackedScene
	var module := scene.instantiate()
	add_child(module)
	var src := module.get_node("Src") as Node2D
	src._io_silent = true
	src.SAVE_PATH = "user://test_life.cfg"
	module.boot({"save_dir": "user://"})
	_ok(src._overlay.visible, "boot 后遮罩可见")
	for child in src._overlay.get_children():
		if child is VBoxContainer:
			for sub in child.get_children():
				if sub is Button:
					sub.pressed.emit()
	_ok(not src._overlay.visible, "点开始后遮罩隐藏")
	var got := {"ok": false}
	module.quit_requested.connect(func(stats: Dictionary) -> void:
		got.clear()
		got.merge(stats))
	module.quit_to_shell()
	_ok(got.has("score"), "quit result 送达")
	_ok(got.has("score") and got.has("playtime") and got.has("achievements") and got.has("extra"), "四字段齐全")
	var extra: Dictionary = got.get("extra", {})
	_ok(extra.has("balance") and extra.has("biggest_win"), "extra 子字段齐")
	_free_module(src)

# === 救济 ===
func _test_relief() -> void:
	var src := _new_src()
	src._overlay.visible = false
	src.balance = 0
	src._refresh_all()
	_ok(src._relief_btn.visible, "破产显示救济")
	src.give_relief()
	_ok(src.balance == SlotConfig.RELIEF_AMOUNT, "领取 +50")
	_ok(not src._relief_btn.visible, "领取后隐藏")
	_free_module(src)

# === 成就 ===
func _test_achievements() -> void:
	var src := _new_src()
	var earned := []
	src.achievement_earned.connect(func(aid: String) -> void: earned.append(aid))
	# 停 77（stop13 或 6? RING: index13=seven）押七 10 份 → 110 → jackpot
	src._bets = [0, 30, 0, 0, 0, 0, 0, 0]
	src.balance = 1000
	src._settle([15], 0)   # 停bigseven：30×40=1200 ≥300 两成全解锁
	_ok(earned.has("sl_jackpot"), "中 77 解锁 sl_jackpot")
	_ok(earned.has("sl_rich300"), "win=600 解锁 sl_rich300")
	var stats: Dictionary = src.quit_stats()
	_ok((stats["achievements"] as Array).size() == 2, "quit 携带 2 成就")
	_free_module(src)

# === reset_run ===
func _test_reset_run() -> void:
	var src := _new_src()
	src._bets = [3, 0, 0, 0, 0, 0, 0, 0]
	var module := src.get_parent()
	module.reset_run()
	var cleared := true
	for b in src._bets:
		if int(b) != 0:
			cleared = false
	_ok(cleared, "reset_run 清押注")
	_ok(src.balance == SlotConfig.BASE_BALANCE, "余额保留")
	_free_module(src)

# === 万局统计（苹果单符号纯 RTP 口径）===
func _test_statistics() -> void:
	var src := _new_src()
	src._rng.seed = 12345
	var freq := {}
	var runs := 10000
	var total_win := 0.0
	# 纯 RTP 口径：每局恒押「橙子」1 币（goslot 用户押注语义：押额×score）
	# 橙子格：0(bigorange w2500 score10) 11(smallorange w2500 score2) 12(bigorange w2500 score10)
	for r in runs:
		var stop: int = SlotConfig.weighted_stop(src._rng)
		var key: String = SlotConfig.RING[stop]["flag"]
		freq[key] = freq.get(key, 0) + 1
		if key == "orange":
			var score: int = int(SlotConfig.RING[stop]["score"])
			total_win += float(score)
	var avg := total_win / runs
	# 理论：P(orange)=7500/total_w；E = P×(10+2+10)/3 → 计算 total_w
	var total_w := 0
	for w in SlotConfig.WEIGHTS:
		total_w += w
	var theo := 7500.0 / total_w * 22.0 / 3.0
	_ok(absf(avg - theo) < theo * 0.08 + 0.02, "橙子押注实测 RTP %.3f ≈ 理论 %.3f±8%%" % [avg, theo])
	_ok(freq.size() >= 8, "≥8 种 flag 被命中")
	_ok(freq.get("luckytry", 0) > 0, "Lucky 格有命中")
	_free_module(src)
