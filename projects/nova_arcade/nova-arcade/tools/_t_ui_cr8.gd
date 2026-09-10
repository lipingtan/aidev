extends Node
## 临时验证：goslot 机制对齐（押币即扣/中奖池/下分/比大小 1~10 同箱/救济/quit_stats）
var fails := 0
func ck(cond: bool, msg: String) -> void:
	if not cond:
		fails += 1
		print("[t] FAIL: ", msg)

func _ready() -> void:
	var scene := load("res://games/slot_machine/module.tscn") as PackedScene
	var m := scene.instantiate()
	add_child(m)
	await get_tree().process_frame
	m = m.get_node("Src")   # adapter 根下才是 slot_main
	m._io_silent = true
	# 0) 押注热区 = 白格+黑格整体（y 1048，高 132），计数 Label 挂在热区按钮内
	ck(m._bet_btns.size() == 8, "8 bet btns")
	for i in 8:
		ck(absf(m._bet_btns[i].position.y - 1048.0) < 1.0 and absf(m._bet_btns[i].size.y - 132.0) < 1.0,
			"btn%d hotzone y=1048 h=132, got y=%s h=%s" % [i, m._bet_btns[i].position.y, m._bet_btns[i].size.y])
	ck(m._bet_count_lbls.size() == 8, "8 count labels")
	ck(m._bet_count_lbls[0].get_parent() == m._bet_btns[0], "count label inside hot zone")
	# 1) 押币即扣余额
	m.balance = 100
	m._bet_step = 5
	m._bets = [0,0,0,0,0,0,0,0]
	var ok: bool = m.add_bet(0)
	ck(ok, "add_bet ok")
	ck(m.balance == 95, "bet deducts balance immediately: %d" % m.balance)
	ck(m._bets[0] == 5, "bet registered 5")
	# 2) 清币返还
	m.clear_bet(0)
	ck(m.balance == 100, "clear_bet refunds: %d" % m.balance)
	ck(m._bets[0] == 0, "bets cleared")
	# 3) 中奖进池不动余额 + 池显示
	m._bets = [0,0,0,0,0,0,0,0]
	m.add_bet(7)   # apple 5
	m._settle([4]) # apple 格 → 5×5=25
	ck(m.win_pool == 25, "win goes to pool: %d" % m.win_pool)
	ck(m.balance == 95, "balance untouched by win: %d" % m.balance)
	ck(m._pool_label != null, "pool label exists")
	# 4) 下分：池→余额，池清零
	m.cash_out()
	ck(m.balance == 120, "cashout adds pool to balance: %d" % m.balance)
	ck(m.win_pool == 0, "pool cleared after cashout")
	# 5) 比大小：无池 → 点大/小无效（保持 IDLE）
	ck(m.win_pool == 0, "pool empty")
	m._on_gamble_pressed(true)
	ck(m.state == m.State.IDLE_BET, "gamble rejected with empty pool")
	# 6) 比大小：大/小按钮不闪；红绿灯（面板中间偏下 ×2）开牌时交替闪、停牌落在开出侧
	m.win_pool = 50
	m._refresh_all()
	ck(m._gamble_big_btn != null and m._gamble_small_btn != null, "big/small btns exist")
	ck(not m._gamble_big_btn.disabled, "big btn enabled with pool")
	ck(m._guess_lights.size() == 2, "guess lights x2")
	ck(m._guess_light_tween == null, "no light anim at idle")
	var pool_before: int = m.win_pool
	m._on_gamble_pressed(true)
	ck(m.state == m.State.GUESSING, "reveal started")
	ck(m._guess_light_tween != null and m._guess_light_tween.is_valid(), "lights swapping during reveal")
	# 等开牌完成（1.2s 快闪 + 余量）
	for wait_i in 40:
		if m.state == m.State.IDLE_BET:
			break
		await get_tree().create_timer(0.1).timeout
	ck(m.state == m.State.IDLE_BET, "back to IDLE after reveal")
	ck(m._guess_light_tween == null, "light anim stopped after reveal")
	var l0_red: bool = m._guess_lights[0].texture.resource_path.contains("light1")
	var l1_green: bool = m._guess_lights[1].texture.resource_path.contains("light2")
	ck(l0_red or l1_green, "lamp parked on revealed side")
	ck(m.win_pool == pool_before * 2 or m.win_pool == 0, "pool doubled or lost: %d" % m.win_pool)
	ck(m._gamble_info.text.contains("开出"), "reveal info: " + m._gamble_info.text)
	# 7) 完整 spin 链路：押 apple 停 idx4 → 池增加、主灯闪烁、金币生成
	m._bets = [0,0,0,0,0,0,0,0]
	m.balance = 500
	m.win_pool = 0
	m._bet_step = 1
	m.add_bet(7)
	await m.spin(4)
	ck(m.state == m.State.IDLE_BET, "spin back to IDLE")
	ck(m.win_pool == 5, "spin win into pool (apple×5): %d" % m.win_pool)
	ck(m._shine.animation == "shine1" and m._shine.is_playing(), "main lamp blinking after hit")
	ck(m._hit_icons.size() >= 1, "hit cell icon registered for blink: %d" % m._hit_icons.size())
	var icon_ok := true
	for ic in m._hit_icons:
		if not is_instance_valid(ic) or ic.modulate.a >= 1.0:
			icon_ok = false   # 闪烁中被 tween 拉低过透明度
	ck(icon_ok or m._hit_icons.is_empty(), "hit icon blinking (alpha pulsed)")
	# 音效资源存在
	for s in ["win_big", "orb_run", "firecracker"]:
		ck(m._sounds.has(s) and m._sounds[s] != null, "sfx loaded: " + s)
	await get_tree().create_timer(0.2).timeout
	var coin_cnt := 0
	for n in m._ring_layer.get_children():
		if n is Sprite2D and n.texture == m._coin_tex():
			coin_cnt += 1
	ck(coin_cnt >= 4, "coins flying: %d" % coin_cnt)
	print("[t] coins=", coin_cnt, " pool=", m.win_pool)
	# 7.5) 散花轮：礼炮音就绪 + board 底图归位；散花次数在 5~7（大 Lucky）
	ck(m._sounds.has("firecracker") and m._sounds["firecracker"] != null, "firecracker sfx ready")
	m._bets = [0,0,0,0,0,0,0,0]
	m.balance = 500
	m.add_bet(0)   # bar（不押 lucky 路径符号）
	await m.spin(9)   # 大 Lucky
	ck(m.state == m.State.IDLE_BET, "lucky round back to IDLE")
	ck(m._orb_sprites.size() >= 5 and m._orb_sprites.size() <= 7, "big lucky orbs 5..7: %d" % m._orb_sprites.size())
	await get_tree().create_timer(0.6).timeout   # 等礼炮抖动 tween（0.3s+）走完归位
	ck(m._board.position == Vector2(0, 140), "board home restored after shake: %s" % str(m._board.position))
	print("[t] lucky orbs=", m._orb_sprites.size(), " lamp=", m._lamp_pos)
	# 8) 押币打断：池滚动立即终值、余额立即终值 + 闪烁停止
	await get_tree().create_timer(0.25).timeout
	m.add_bet(7)
	ck(m._center.text == "💰 %d" % m.balance, "balance final now: " + m._center.text)
	ck(m._pool_label.text == "中奖 %d" % m.win_pool, "pool final now: " + m._pool_label.text)
	ck(m._hit_icons.is_empty(), "bet stops icon blink")
	# 9) 续押：完整复刻上轮；余额不足 → 整体拒绝并提示手动押币
	m._last_bets = [10, 4, 0, 0, 0, 0, 0, 0]
	m._bets = [0, 0, 0, 0, 0, 0, 0, 0]
	m.balance = 14        # 需 14：够
	m.repeat_last_bets()
	ck(m._bets[0] == 10 and m._bets[1] == 4, "rebet restores exact amounts: %s" % str(m._bets))
	ck(m.balance == 0, "rebet deducts exact: %d" % m.balance)
	# 余额不足：上轮 [6,0,...] 需 6，余额只给 5 → 拒绝且押注不变
	m._last_bets = [6, 0, 0, 0, 0, 0, 0, 0]
	m._bets = [0, 0, 0, 0, 0, 0, 0, 0]
	m.balance = 5
	m.repeat_last_bets()
	ck(m._bets[0] == 0, "insufficient rebet rejected, bets unchanged")
	ck(m._gamble_info.text.contains("余额不够"), "rebet hint: " + m._gamble_info.text)
	# 10) quit_stats 含池
	var st: Dictionary = m.quit_stats()
	ck(int(st["score"]) == m.balance + m.win_pool - m._start_balance, "quit_stats includes pool")
	print("[t] RESULT fails=", fails)
	get_tree().quit(1 if fails > 0 else 0)
