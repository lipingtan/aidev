extends Node2D
## 幸运轮盘·水果机（CR-8 修订3）：goslot 1:1 移植
## - board.png 底图（900×1600 → 缩放 0.5 = 450×800 居中）
## - 24 格环形（goslot startpos/steps/steplen 原值布局）
## - 三段变速跑灯（goslot RoundNormal.startbet 节奏公式）
## - Lucky 散花（RoundLuckyTry 方向交替 + 独立计分）
## - 账本/结算卡/成就协议不变

signal on_game_over(stats: Dictionary)
signal achievement_earned(aid: String)

enum State { IDLE_BET, RUNNING, SETTLE, GUESSING }

const SAVE_PATH_DEFAULT := "user://save.cfg"
const IMG := "res://games/slot_machine/assets/images/"
const SND := "res://games/slot_machine/assets/sounds/"

## --- 账本/会话 ---
var SAVE_PATH: String = SAVE_PATH_DEFAULT
var balance: int = SlotConfig.BASE_BALANCE
var _start_balance: int = SlotConfig.BASE_BALANCE
var win_pool: int = 0                     # 中奖池（goslot tempScore）：中奖进池，下分才入余额
var _bets: Array = [0, 0, 0, 0, 0, 0, 0, 0]   # 8 可押符号（登记制）
var _bet_step: int = SlotConfig.BET_STEPS[0]
var _biggest_win: int = 0
var _last_win: int = 0                    # 上一轮中奖额（比大小用）
var _gamble_used: bool = false            # 本轮是否已比过大小
var _last_bets: Array = [0, 0, 0, 0, 0, 0, 0, 0]   # 上一轮押注（续押用）
var _session_achievements: Array = []
var _spin_count: int = 0
var state: State = State.IDLE_BET
var paused: bool = false
var _io_silent: bool = false
var _rng := RandomNumberGenerator.new()

## --- 场景 ---
var _board: Sprite2D = null
var _ring_layer: CanvasLayer = null       # 24 格图标（layer -1）
var _ui: CanvasLayer = null
var _bg_layer: CanvasLayer = null
var _center: Label = null
var _pool_label: Label = null             # 中奖池显示（板中央）
var _spin_btn: Button = null
var _bet_btns: Array = []
var _bet_count_lbls: Array = []           # 每槽黑区的押币数 Label
var _relief_btn: Button = null
var _repeat_btn: Button = null            # 续押（黄色面板）
var _gamble_big_btn: Button = null        # 比大小·大
var _gamble_small_btn: Button = null      # 比大小·小
var _guess_tween: Tween = null            # （已弃用按钮闪烁，保留变量名防外部引用）
var _guess_lights: Array = []             # 猜大小红绿灯 ×2（goslot GuessLightsLogic）
var _guess_light_tween: Tween = null      # 红绿交替动画
var _cashout_btn: Button = null           # 下分：win_pool → 余额
var _gamble_panel: Control = null
var _gamble_info: Label = null
var _overlay: Control = null
var _shine: AnimatedSprite2D = null       # 停格闪光
var _orb_sprites: Array = []
var _orb_run_player: AudioStreamPlayer = null   # 散灯跑动轰鸣（循环）
var _hit_icons: Array = []                # 命中格图标（闪烁直到重新押币）
var _icon_tween: Tween = null             # 图标闪烁动画
var _shake_tween: Tween = null            # 礼炮屏幕抖动（单实例，重触发先归位）

## 跑灯状态
var _lamp_pos := 0
var _run_aborted := false
var _sounds := {}

func _ready() -> void:
	_rng.randomize()
	_load_sounds()
	_build_board()
	_build_ring()
	_build_ui()
	_refresh_all()

func _load_sounds() -> void:
	for s in ["ding", "ding1", "pongbig", "pongsmall", "roll", "win_big", "orb_run", "firecracker"]:
		_sounds[s] = load(SND + s + ".wav")

func _play(sname: String) -> void:
	if _io_silent or not _sounds.has(sname):
		return
	var p := AudioStreamPlayer.new()
	p.stream = _sounds[sname]
	add_child(p)
	p.play()
	p.finished.connect(p.queue_free)

## 循环播放散灯跑动轰鸣（播放中重复调用不叠加）
func _play_loop(sname: String) -> void:
	if _io_silent or not _sounds.has(sname):
		return
	if _orb_run_player != null and is_instance_valid(_orb_run_player) and _orb_run_player.playing:
		return
	_orb_run_player = AudioStreamPlayer.new()
	_orb_run_player.stream = _sounds[sname]
	add_child(_orb_run_player)
	_orb_run_player.finished.connect(func() -> void:
		if _orb_run_player != null and is_instance_valid(_orb_run_player) and _orb_run_player.playing == false and state == State.RUNNING:
			_orb_run_player.play())   # 循环直到状态离开 RUNNING
	_orb_run_player.play()

func _stop_loop() -> void:
	if _orb_run_player != null and is_instance_valid(_orb_run_player):
		_orb_run_player.stop()
		_orb_run_player.queue_free()
	_orb_run_player = null

# === 账本 ===

func load_balance() -> void:
	var cf := ConfigFile.new()
	if cf.load(SAVE_PATH) == OK:
		balance = clamp(int(cf.get_value("slot", SlotConfig.SAVE_KEY, SlotConfig.BASE_BALANCE)), 0, SlotConfig.MAX_BALANCE)
	else:
		balance = SlotConfig.BASE_BALANCE
	_start_balance = SlotConfig.BASE_BALANCE

func save_balance() -> void:
	if _io_silent:
		return
	var cf := ConfigFile.new()
	cf.load(SAVE_PATH)
	cf.set_value("slot", SlotConfig.SAVE_KEY, balance)
	cf.save(SAVE_PATH)

func give_relief() -> void:
	balance = clamp(balance + SlotConfig.RELIEF_AMOUNT, 0, SlotConfig.MAX_BALANCE)
	save_balance()
	_refresh_all()

func total_bet() -> int:
	var s := 0
	for b in _bets:
		s += int(b)
	return s

# === 押注（登记制）===

func add_bet(sym: int) -> bool:
	if state != State.IDLE_BET or paused:
		return false
	if sym < 0 or sym >= SlotConfig.BET_SYMBOLS.size():
		return false
	if _bet_step > balance:
		return false
	_bets[sym] = int(_bets[sym]) + _bet_step
	balance -= _bet_step                     # goslot：押币即扣 credit
	_stop_orb_blink()
	_finish_balance_roll()
	_finish_pool_roll()
	_refresh_all()
	return true

func clear_bet(sym: int) -> void:
	if state != State.IDLE_BET or sym < 0 or sym >= _bets.size():
		return
	balance = clamp(balance + int(_bets[sym]), 0, SlotConfig.MAX_BALANCE)   # 清币返还
	_bets[sym] = 0
	_stop_orb_blink()
	_finish_balance_roll()
	_finish_pool_roll()
	_refresh_all()

func set_bet_step(v: int) -> void:
	_bet_step = v
	_refresh_all()

func can_spin() -> bool:
	return state == State.IDLE_BET and not paused and total_bet() > 0

# === 跑灯（goslot RoundNormal.startbet 节奏公式 1:1）===

## 计算每格的 (sec1, sec2)：sec1=灯亮起始时刻, sec2=持续时长
func _compute_schedule(nodes_count: int, startindex: int) -> Array:
	var o := SlotConfig.RUN_OPTIONS
	var slowcount: int = o["slowcount"]
	var slowwait1: float = o["slowwait1"]
	var slowwait2: float = o["slowwait2"]
	var fastwait1: float = o["fastwait1"]
	var fastwait2: float = o["fastwait2"]
	var slowendwait1: float = o["slowendwait1"]
	var slowendwait2: float = o["slowendwait2"]
	var slowendcount: int = o["slowendcount"]
	var delta: float = o["delta"]
	var delta2: float = o["delta2"]
	var schedule := []
	var sec1 := 0.0
	var sec2 := 0.0
	for idx in range(nodes_count):
		var idx2 := idx - startindex
		if idx2 <= slowcount:
			sec1 = slowwait1 * idx2 - delta * idx2 + 0.0001
			sec2 = slowwait2 - delta * idx2
			if idx2 == slowcount:
				sec2 = fastwait1 + 0.1
		elif idx >= nodes_count - slowendcount:
			sec1 = slowwait1 * slowcount + fastwait1 * (nodes_count - slowcount - slowendcount - startindex) \
				+ slowendwait1 * (idx - (nodes_count - slowendcount) + 1) + 0.0001 \
				+ (idx - (nodes_count - slowendcount) + 1) * delta2
			sec2 = slowendwait2
		else:
			sec1 = slowwait1 * slowcount + fastwait1 * idx2
			sec2 = fastwait2
		schedule.append([sec1, sec2])
	return schedule

## 旋转入口
func spin(force_stop: int = -1) -> void:
	if not can_spin():
		return
	state = State.RUNNING
	_set_locked(true)
	_clear_orbs()   # 清上一轮散花灯珠
	_play("roll")
	var stop := SlotConfig.weighted_stop(_rng) if force_stop < 0 else force_stop % 24
	await _run_main(stop)   # 押币时已扣款，spin 不再扣

## 主轮跑灯：从上次停位起跑（goslot startindex 语义）×4 圈 + 收尾跑到停位
func _run_main(stop: int) -> void:
	var nodes := []
	var cur: int = _lamp_pos % 24
	for i in range(96):                       # 4 整圈，从当前停位连续起跑
		nodes.append((cur + i) % 24)
	var tail_len := (stop - cur + 24) % 24    # 收尾：继续走到停位
	if tail_len == 0:
		tail_len = 24                         # 停位=起跑位 → 再走一整圈
	for i in range(1, tail_len + 1):
		nodes.append((cur + i) % 24)
	var startindex: int = _lamp_pos % 24
	var schedule := _compute_schedule(nodes.size(), startindex)
	_run_aborted = false
	# 播放：每节点在 sec1 亮灯、sec1+sec2 熄灯（goslot TimerManager startElapsed/endElapsed）
	for idx in range(nodes.size()):
		if _run_aborted:
			return
		var pos: int = nodes[idx]
		var sec1: float = schedule[idx][0]
		var sec2: float = schedule[idx][1]
		_lamp_pos = pos
		_move_lamp(pos)
		_play("ding1")
		var is_end: bool = idx == nodes.size() - 1
		var prev_sec1: float = schedule[idx - 1][0] if idx > 0 else 0.0
		var wait: float = sec1 - prev_sec1
		if wait > 0:
			await get_tree().create_timer(maxf(wait, 0.001)).timeout
		if _run_aborted:
			return
		if is_end:
			# 停格闪光（shine2 常亮）；命中押注符号 → 主灯闪烁（重新押币才停）
			_play_shine(pos)
			var hit_stop := false
			var stop_flag: String = SlotConfig.RING[pos]["flag"]
			for si in SlotConfig.BET_SYMBOLS.size():
				if SlotConfig.BET_SYMBOLS[si]["key"] == stop_flag and int(_bets[si]) > 0:
					hit_stop = true
					break
			if hit_stop:
				_shine.play("shine1")
				_blink_cell_icon(pos)
			_play("ding")
			await get_tree().create_timer(0.5).timeout
	# 主轮结算；若停位是 Lucky → 散花轮
	if stop == SlotConfig.LUCKY_IDX_BIG or stop == SlotConfig.LUCKY_IDX_SMALL:
		var orb_stops := await _run_lucky_try(stop)
		var all_stops := [stop]
		all_stops.append_array(orb_stops)
		_settle(all_stops)
	else:
		_settle([stop])

## Lucky 散花轮（goslot RoundLuckyTry：方向交替 + 每灯独立跑到目标格）
func _run_lucky_try(stop: int) -> Array:
	_play_loop("orb_run")   # 低频轰鸣贯穿散灯跑动
	var orb_range: Array = SlotConfig.BIG_LUCKY_ORBS if stop == SlotConfig.LUCKY_IDX_BIG else SlotConfig.SMALL_LUCKY_ORBS
	var count: int = _rng.randi_range(int(orb_range[0]), int(orb_range[1]))
	var lucky_numbers: Array = []
	var excluded := [stop]
	for i in count:
		var n: int = SlotConfig.weighted_stop(_rng)
		while n in excluded or n in lucky_numbers:
			n = (n + 1) % 24
		lucky_numbers.append(n)
		excluded.append(n)
	# 方向交替（goslot: direction = (tryIndex+1)%2 - tryIndex%2 → +1,-1,+1...）
	# 主灯留在幸运格常亮（_run_main is_end 已 shine2）；每次散花从幸运格飞出一个新灯珠
	var orb_stops := []
	var direction := 1
	var positions := _ring_positions()
	for try_i in lucky_numbers.size():
		direction = (try_i + 1) % 2 - try_i % 2
		var target: int = lucky_numbers[try_i]
		# 新灯珠从停格跑到 target（goslot getLuckyTryItems 路径）
		var orb := _make_orb(stop)
		var path := _lucky_path(stop, target, direction)
		for pos in path:
			if _run_aborted:
				return orb_stops
			orb.position = positions[pos]
			await get_tree().create_timer(SlotConfig.RUN_OPTIONS["fastwait2"]).timeout
		# 到达：礼炮爆响 + 屏幕抖动 + 手机震动；命中格图标闪烁（直到重新押币）
		orb.position = positions[target]
		_stop_loop()
		_play("firecracker")
		_shake_screen(0.3, 14.0)
		_vibrate(120)
		var hit := false
		var flag: String = SlotConfig.RING[target]["flag"]
		for si in SlotConfig.BET_SYMBOLS.size():
			if SlotConfig.BET_SYMBOLS[si]["key"] == flag and int(_bets[si]) > 0:
				hit = true
				break
		orb.play("shine1" if hit else "shine2")
		if hit:
			_blink_cell_icon(target)
		orb_stops.append(target)
	return orb_stops

## 屏幕抖动（goslot CameraShaker 语义）：duration 秒内随机偏移，结束归零
## 抖 board 底图 Sprite2D（CanvasLayer 无 offset 属性）
## 多颗灯珠连触发：新抖动先 kill 旧 tween 并归位，避免结束值互相覆盖
func _shake_screen(duration: float, max_off: float) -> void:
	if _board == null:
		return
	if _shake_tween != null and _shake_tween.is_valid():
		_shake_tween.kill()
	_board.position = BOARD_OFFSET   # 先归位再抖
	var tw := create_tween()
	_shake_tween = tw
	var steps := int(duration / 0.03)
	for i in steps:
		var decay := 1.0 - float(i) / steps
		var ox := _rng.randf_range(-max_off, max_off) * decay
		var oy := _rng.randf_range(-max_off, max_off) * decay
		tw.tween_property(_board, "position", BOARD_OFFSET + Vector2(ox, oy), 0.03)
	tw.tween_property(_board, "position", BOARD_OFFSET, 0.03)

## 手机震动（Android/iOS；桌面无感忽略）
func _vibrate(ms: int) -> void:
	if _io_silent:
		return
	Input.vibrate_handheld(ms)

## 命中格图标呼吸闪烁（透明度脉冲），重新押币时停
func _blink_cell_icon(idx: int) -> void:
	var icon := _cell_icon(idx)
	if icon == null or icon in _hit_icons:
		return
	_hit_icons.append(icon)
	if _icon_tween == null or not _icon_tween.is_valid():
		_icon_tween = create_tween().set_loops()
		_icon_tween.tween_method(_set_icons_alpha, 1.0, 0.25, 0.22)
		_icon_tween.tween_method(_set_icons_alpha, 0.25, 1.0, 0.22)

func _set_icons_alpha(a: float) -> void:
	for ic in _hit_icons:
		if is_instance_valid(ic):
			ic.modulate = Color(1, 1, 1, a)

## 取环上第 idx 格的图标 Sprite2D
func _cell_icon(idx: int) -> Sprite2D:
	for cell in _ring_layer.get_children():
		if cell is Node2D and String(cell.name).begins_with("Cell%d_" % idx):
			for ch in cell.get_children():
				if ch is Sprite2D:
					return ch
	return null

## 重新押币后停掉命中灯珠/主灯/格图标的闪烁（转常亮）
func _stop_orb_blink() -> void:
	for o in _orb_sprites:
		if is_instance_valid(o) and o.animation == "shine1" and o.is_playing():
			o.play("shine2")
	if _shine != null and _shine.animation == "shine1" and _shine.is_playing():
		_shine.play("shine2")
	if _icon_tween != null and _icon_tween.is_valid():
		_icon_tween.kill()
	_icon_tween = null
	for ic in _hit_icons:
		if is_instance_valid(ic):
			ic.modulate = Color(1, 1, 1, 1.0)
	_hit_icons.clear()

## 中奖金币动画：N 枚金币从停格飞向中奖池显示区（缩放消失），池数字滚动
func _coin_burst(stop_idx: int, amount: int) -> void:
	if amount <= 0:
		return
	var from: Vector2 = _ring_positions()[stop_idx]
	var to := Vector2(360, 440)                     # 中奖池 Label 附近
	var n: int = clampi(4 + amount / 50, 4, 12)     # 奖金越多币越多
	for i in n:
		var c := Sprite2D.new()
		c.texture = _coin_tex()
		c.position = from
		c.scale = Vector2(0.8, 0.8)
		c.z_index = 50
		_ring_layer.add_child(c)
		_fly_coin(c, to, 0.5 + 0.05 * i)
	if _pool_tween != null and _pool_tween.is_valid():
		_pool_tween.kill()   # 上一轮滚动还在 → 直接替掉
	_pool_tween = create_tween()
	_pool_tween.tween_method(func(v: float) -> void:
		_pool_label.text = "中奖 %d" % int(round(v)),
		float(win_pool - amount), float(win_pool), 0.6)
	_pool_tween.tween_callback(func() -> void: _pool_tween = null)

## 押币打断池数字滚动：杀 tween，池显示立即终值
func _finish_pool_roll() -> void:
	if _pool_tween != null and _pool_tween.is_valid():
		_pool_tween.kill()
	_pool_tween = null
	_refresh_pool_label()

func _refresh_pool_label() -> void:
	if _pool_label != null:
		_pool_label.text = "中奖 %d" % win_pool if win_pool > 0 else ""

## 押币打断数字滚动：杀掉 tween，余额立即显示最终值
func _finish_balance_roll() -> void:
	if _balance_tween != null and _balance_tween.is_valid():
		_balance_tween.kill()
	_balance_tween = null
	if _center != null:
		_center.text = "💰 %d" % balance

func _fly_coin(c: Sprite2D, to: Vector2, delay: float) -> void:
	var tw := create_tween()
	tw.tween_interval(delay)
	tw.tween_callback(func() -> void: c.visible = true)
	tw.tween_property(c, "position", to, 0.45).set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_IN)
	tw.parallel().tween_property(c, "scale", Vector2(0.4, 0.4), 0.45)
	tw.tween_callback(func() -> void: c.queue_free())

var _coin_tex_cache: ImageTexture = null
var _balance_tween: Tween = null            # 余额数字滚动（押币可打断）
var _pool_tween: Tween = null               # 中奖池数字滚动（押币可打断）
## 程序生成金币贴图（金色圆 + 深金描边 + 高光），缓存复用
func _coin_tex() -> ImageTexture:
	if _coin_tex_cache != null:
		return _coin_tex_cache
	var img := Image.create(32, 32, false, Image.FORMAT_RGBA8)
	for x in 32:
		for y in 32:
			var dx := float(x) - 15.5
			var dy := float(y) - 15.5
			var d := sqrt(dx * dx + dy * dy)
			if d <= 14.0:
				var col := Color(1.0, 0.84, 0.0)       # 金
				if d > 11.5:
					col = Color(0.85, 0.62, 0.0)       # 边
				elif dx < -3.0 and dy < -3.0 and d < 9.0:
					col = col.lightened(0.35)          # 高光
				img.set_pixel(x, y, col)
	_coin_tex_cache = ImageTexture.create_from_image(img)
	return _coin_tex_cache

## 散花灯珠：与主灯同帧图，跑动 shine1 / 到位 shine2 常亮
func _make_orb(from_idx: int) -> AnimatedSprite2D:
	var orb := AnimatedSprite2D.new()
	orb.sprite_frames = _shine.sprite_frames
	orb.scale = _shine.scale
	orb.position = _ring_positions()[from_idx]
	_ring_layer.add_child(orb)
	orb.play("shine1")
	_orb_sprites.append(orb)
	return orb

## 清散花灯珠（下一轮 spin 开始时）
func _clear_orbs() -> void:
	for o in _orb_sprites:
		if is_instance_valid(o):
			o.queue_free()
	_orb_sprites.clear()

## goslot getLuckyTryItems 路径生成（顺/逆时针环形）
func _lucky_path(fire_index: int, lucky_number: int, direction: int) -> Array:
	var path := []
	if direction >= 1:
		if lucky_number > fire_index:
			for i in range(fire_index, lucky_number + 1):
				path.append(i)
		elif lucky_number < fire_index:
			for i in range(fire_index, 24):
				path.append(i)
			for i in range(0, lucky_number + 1):
				path.append(i)
	else:
		if lucky_number > fire_index:
			for i in range(fire_index, -1, -1):
				path.append(i)
			for i in range(23, lucky_number - 1, -1):
				path.append(i)
		elif lucky_number < fire_index:
			for i in range(fire_index, lucky_number - 1, -1):
				path.append(i)
	return path

## 停格闪光：跑灯动画切 shine2 常亮
func _play_shine(pos: int) -> void:
	_move_lamp(pos)
	if _shine != null:
		_shine.play("shine2")

## 结算：主停位 + 散花停位逐格判定（goslot computeTempScore 语义）
## 中奖进 win_pool（goslot tempScore），不动余额；由下分/比大小处理
func _settle(stops: Array) -> void:
	var win := 0
	var detail := []
	for stop in stops:
		var idx := int(stop)
		var flag: String = SlotConfig.RING[idx]["flag"]
		# flag 映射到押注符号
		for si in SlotConfig.BET_SYMBOLS.size():
			if SlotConfig.BET_SYMBOLS[si]["key"] == flag and int(_bets[si]) > 0:
				var payout: int = SlotConfig.cell_payout(idx, int(_bets[si]))
				win += payout
				detail.append("%s×%d" % [flag, payout])
	_last_win_detail = detail
	if OS.get_environment("SLOT_DEBUG") == "1":
		print("[settle] stops=", stops, " bets=", _bets, " win=", win)
	if win > _biggest_win:
		_biggest_win = win
	_last_bets = _bets.duplicate()
	_bets = [0, 0, 0, 0, 0, 0, 0, 0]
	_spin_count += 1
	if win > 0:
		win_pool += win
		_play("win_big")   # 中奖震撼音
		if not _io_silent:
			_show_result(win, detail)
		_coin_burst(int(stops[0]), win)   # 金币飞向中奖池显示区 + 池数字滚动
	else:
		if not _io_silent:
			_center.text = "未中，再接再厉"
	_check_achievements(win)
	state = State.IDLE_BET
	_set_locked(false)
	_refresh_all()

func _show_result(win: int, detail: Array) -> void:
	if win > 0:
		Sound.coin()
		_center.text = "中奖 %d 币！%s" % [win, ", ".join(PackedStringArray(detail))]
	else:
		_center.text = "未中，再接再厉"

var _last_win_detail: Array = []
func win_detail_all() -> Array:
	return _last_win_detail

func _check_achievements(win: int) -> void:
	var hit_seven := false
	for d in _last_win_detail:
		if String(d).begins_with("seven") or String(d).begins_with("ring"):
			hit_seven = true
	if hit_seven and not _session_achievements.has("sl_jackpot"):
		_session_achievements.append("sl_jackpot")
		achievement_earned.emit("sl_jackpot")
	if win >= 300 and not _session_achievements.has("sl_rich300"):
		_session_achievements.append("sl_rich300")
		achievement_earned.emit("sl_rich300")

func quit_stats() -> Dictionary:
	return {
		"score": balance + win_pool - _start_balance,   # 中奖池未下分也算成绩
		"playtime": 0.0,
		"achievements": _session_achievements.duplicate(),
		"extra": {"balance": balance, "win_pool": win_pool, "biggest_win": _biggest_win, "spin_count": _spin_count},
	}

func set_paused(v: bool) -> void:
	paused = v

func reset_run() -> void:
	_bets = [0, 0, 0, 0, 0, 0, 0, 0]
	_refresh_all()

# === 渲染 ===

## 24 格位置（goslot genPoints 1:1，goslot 视口 450×800 原生坐标）：
## startpos(26,110) 每边7格 步长56 → 7*56=392 跨度
## 我们的视口 720×1560：board 900×1600 × 0.8 = 720×1280，水平铺满、垂直居中(偏移140)
## 所有 goslot 坐标 × 0.8（= BoardScale）+ 板偏移
const BOARD_SCALE := 0.8
const BOARD_OFFSET := Vector2(0, 140)

func _ring_positions() -> Array:
	var pts := []
	var start: Vector2 = SlotConfig.START_POS
	pts.append(start)
	# top
	var prev: Vector2 = pts[pts.size() - 1]
	for i in range(1, SlotConfig.STEPS):
		pts.append(prev + Vector2(SlotConfig.STEPLEN * i, 0))
	# right
	prev = pts[pts.size() - 1]
	for i in range(1, SlotConfig.STEPS):
		pts.append(prev + Vector2(0, SlotConfig.STEPLEN * i))
	# bottom
	prev = pts[pts.size() - 1]
	for i in range(1, SlotConfig.STEPS):
		pts.append(prev + Vector2(-1 * SlotConfig.STEPLEN * i, 0))
	# left
	prev = pts[pts.size() - 1]
	for i in range(1, SlotConfig.STEPS):
		pts.append(prev + Vector2(0, -1 * SlotConfig.STEPLEN * i))
	# 对齐白格：BetItem 有 (28,28) 自身偏移，board 图内白格中心 = (格点+(28,28))×2
	# 我们显示 board ×0.8 → 图标位置 = (格点+(28,28))×2×0.8 + OFFSET = (格点+28)×1.6 + OFFSET
	var item_off := Vector2(28, 28)
	for i in pts.size():
		pts[i] = (pts[i] + item_off) * (BOARD_SCALE * 2.0) + BOARD_OFFSET
	return pts

func _build_board() -> void:
	# 底图层（-2）
	_bg_layer = CanvasLayer.new()
	_bg_layer.name = "BgLayer"
	_bg_layer.layer = -2
	add_child(_bg_layer)
	# board.png 900×1600 缩放 0.5 → 450×800，水平居中（720-450)/2=135，垂直 60 起
	_board = Sprite2D.new()
	_board.name = "Board"
	_board.texture = load(IMG + "board.png")
	_board.centered = false
	_board.scale = Vector2(BOARD_SCALE, BOARD_SCALE)
	_board.position = BOARD_OFFSET
	_bg_layer.add_child(_board)

func _build_ring() -> void:
	_ring_layer = CanvasLayer.new()
	_ring_layer.name = "RingLayer"
	_ring_layer.layer = -1
	add_child(_ring_layer)
	var positions := _ring_positions()
	for i in 24:
		var entry: Dictionary = SlotConfig.RING[i]
		var cell := Node2D.new()
		cell.name = "Cell%d_%s" % [i, entry["flag"]]
		cell.position = positions[i]
		# bgitem（goslot BetItem bgitem scale 0.5 + imagescale）
		# 白格显示宽 87px（109 board 系 ×0.8）；rollpic 112px → scale 0.78 满格
		var img_scale: float = 0.78 if bool(entry["big"]) else 0.78 * SlotConfig.SMALL_SCALE
		var icon := Sprite2D.new()
		icon.texture = load(SlotConfig.icon_path(i))
		icon.scale = Vector2(img_scale, img_scale)
		cell.add_child(icon)
		_ring_layer.add_child(cell)
	# 跑灯（goslot 原行为：shineitem 播 shine1 循环帧跟格跑，终点切 shine2 常亮）
	_shine = AnimatedSprite2D.new()
	_shine.name = "LampShine"
	var frames := SpriteFrames.new()
	frames.add_animation("shine1")
	frames.set_animation_speed("shine1", 10.0)
	frames.add_frame("shine1", load(IMG + "shine03.png"))
	frames.add_frame("shine1", load(IMG + "shine02.png"))
	frames.add_animation("shine2")
	frames.set_animation_speed("shine2", 3.0)
	frames.add_frame("shine2", load(IMG + "shine04.png"))
	_shine.sprite_frames = frames
	_shine.scale = Vector2(1.35, 1.35)   # shine 72px → 97px，盖住白格（87px 显示宽）
	_ring_layer.add_child(_shine)
	_shine.visible = true
	_shine.play("shine1")
	_move_lamp(0)

func _move_lamp(idx: int) -> void:
	var positions := _ring_positions()
	_shine.position = positions[idx % 24]

func _build_ui() -> void:
	_ui = CanvasLayer.new()
	_ui.name = "UILayer"
	_ui.layer = 1
	add_child(_ui)
	# 中央信息（board 中央白区）
	_center = Label.new()
	_center.name = "CenterInfo"
	_center.text = "💰 %d" % balance
	_center.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_center.add_theme_font_size_override("font_size", 44)
	_center.add_theme_color_override("font_color", Color(0.5, 0.05, 0.05))
	_center.size = Vector2(360, 200)
	_center.position = Vector2(180, 480)
	_ui.add_child(_center)
	# 中奖池显示（goslot DigitWin：板中央白区、余额上方，金色）
	_pool_label = Label.new()
	_pool_label.name = "WinPool"
	_pool_label.text = ""
	_pool_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_pool_label.add_theme_font_size_override("font_size", 52)
	_pool_label.add_theme_color_override("font_color", ThemeTokens.color("gold"))
	_pool_label.add_theme_color_override("font_outline_color", Color(0.35, 0.05, 0))
	_pool_label.add_theme_constant_override("outline_size", 8)
	_pool_label.size = Vector2(360, 80)
	_pool_label.position = Vector2(180, 400)
	_ui.add_child(_pool_label)
	# 旋转按钮（中央下方）
	_spin_btn = Button.new()
	_spin_btn.name = "SpinBtn"
	_spin_btn.text = "开始"
	_spin_btn.focus_mode = Control.FOCUS_NONE
	_spin_btn.size = Vector2(220, 70)
	_spin_btn.position = Vector2(250, 660)
	_spin_btn.add_theme_font_size_override("font_size", 30)
	_spin_btn.pressed.connect(func() -> void: spin())
	_ui.add_child(_spin_btn)
	# 注额档（黄色面板第 1 行：面板屏幕 y≈1180..1356）
	_steps_row(Vector2(125, 1188))
	# 押注位 8 个（白格图标 + 黑格计数）
	_build_bet_pads()
	# 黄色面板第 2 行：续押 + 比大小
	_build_panel_actions(Vector2(125, 1268))
	# 救济
	_relief_btn = Button.new()
	_relief_btn.name = "ReliefBtn"
	_relief_btn.text = "领取救济 %d 币" % SlotConfig.RELIEF_AMOUNT
	_relief_btn.focus_mode = Control.FOCUS_NONE
	_relief_btn.size = Vector2(320, 70)
	_relief_btn.position = Vector2(200, 940)
	_relief_btn.visible = false
	_relief_btn.pressed.connect(func() -> void:
		give_relief()
		_relief_btn.visible = false)
	_ui.add_child(_relief_btn)
	_build_overlay()

func _steps_row(origin: Vector2) -> void:
	var row := HBoxContainer.new()
	row.name = "BetSteps"
	row.position = origin
	row.add_theme_constant_override("separation", 16)
	_ui.add_child(row)
	for step in SlotConfig.BET_STEPS:
		var b := Button.new()
		b.name = "Step%d" % step
		b.text = "%d 币" % step
		b.focus_mode = Control.FOCUS_NONE
		b.custom_minimum_size = Vector2(190, 56)
		b.add_theme_font_size_override("font_size", 22)
		var val := int(step)
		b.pressed.connect(func() -> void: set_bet_step(val))
		row.add_child(b)
		_step_btns.append(b)

var _step_btns: Array = []

## 黄色面板功能行：续押 + 比大小 + 下分
func _build_panel_actions(origin: Vector2) -> void:
	_gamble_panel = Control.new()
	_gamble_panel.name = "PanelActions"
	_gamble_panel.position = origin
	_gamble_panel.size = Vector2(600, 90)
	_gamble_panel.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_ui.add_child(_gamble_panel)
	var row := HBoxContainer.new()
	row.position = Vector2(0, 0)   # 相对 _gamble_panel（panel 已定位在 origin）
	row.add_theme_constant_override("separation", 16)
	_gamble_panel.add_child(row)
	_repeat_btn = Button.new()
	_repeat_btn.name = "RepeatBtn"
	_repeat_btn.text = "续押"
	_repeat_btn.focus_mode = Control.FOCUS_NONE
	_repeat_btn.custom_minimum_size = Vector2(150, 64)
	_repeat_btn.add_theme_font_size_override("font_size", 24)
	_repeat_btn.pressed.connect(_on_repeat_pressed)
	row.add_child(_repeat_btn)
	# 比大小 = 大/小 双按钮（goslot betbig/betsmall）：奖池>0 慢闪待选，点后快闪开牌
	_gamble_big_btn = Button.new()
	_gamble_big_btn.name = "BetBigBtn"
	_gamble_big_btn.text = "大"
	_gamble_big_btn.focus_mode = Control.FOCUS_NONE
	_gamble_big_btn.custom_minimum_size = Vector2(120, 64)
	_gamble_big_btn.add_theme_font_size_override("font_size", 26)
	_gamble_big_btn.pressed.connect(_on_gamble_pressed.bind(true))
	row.add_child(_gamble_big_btn)
	_gamble_small_btn = Button.new()
	_gamble_small_btn.name = "BetSmallBtn"
	_gamble_small_btn.text = "小"
	_gamble_small_btn.focus_mode = Control.FOCUS_NONE
	_gamble_small_btn.custom_minimum_size = Vector2(120, 64)
	_gamble_small_btn.add_theme_font_size_override("font_size", 26)
	_gamble_small_btn.pressed.connect(_on_gamble_pressed.bind(false))
	row.add_child(_gamble_small_btn)
	_cashout_btn = Button.new()
	_cashout_btn.name = "CashoutBtn"
	_cashout_btn.text = "下分"
	_cashout_btn.focus_mode = Control.FOCUS_NONE
	_cashout_btn.custom_minimum_size = Vector2(150, 64)
	_cashout_btn.add_theme_font_size_override("font_size", 24)
	_cashout_btn.pressed.connect(_on_cashout_pressed)
	row.add_child(_cashout_btn)
	_gamble_info = Label.new()
	_gamble_info.name = "GambleInfo"
	_gamble_info.text = ""
	_gamble_info.add_theme_font_size_override("font_size", 20)
	_gamble_info.add_theme_color_override("font_color", Color(0.35, 0.1, 0))
	row.add_child(_gamble_info)
	_build_guess_lights()

## 猜大小红绿灯（goslot GuessLightsLogic 1:1）：面板中间偏下两颗灯
## light1=红 light2=绿；guessing 动画 = 红亮绿灭 ↔ 红灭绿亮 交替（0.5s 相位）
func _build_guess_lights() -> void:
	_guess_lights = []
	for i in 2:
		var s := Sprite2D.new()
		s.texture = load(IMG + ("light1.png" if i == 0 else "light2.png"))
		s.position = Vector2(310 + i * 100, 1130)   # 面板中央偏下（黄面板上沿）
		s.scale = Vector2(1.6, 1.6)
		s.z_index = 30
		_ui.add_child(s)
		_guess_lights.append(s)
	_set_guess_lights("idle")   # 初始：双灰暗

## 灯态："idle"=双灰 / "swap"=红绿交替闪（开牌中）/ "big"=停大（左红亮）/ "small"=停小（右绿亮）
func _set_guess_lights(mode: String) -> void:
	if _guess_light_tween != null and _guess_light_tween.is_valid():
		_guess_light_tween.kill()
	_guess_light_tween = null
	var gray: Texture2D = load(IMG + "lightgray1.png")
	match mode:
		"idle":
			_guess_lights[0].texture = gray
			_guess_lights[1].texture = gray
			for s in _guess_lights:
				s.modulate = Color(1, 1, 1, 0.4)
		"swap":
			for s in _guess_lights:
				s.modulate = Color(1, 1, 1, 1.0)
			_guess_light_tween = create_tween().set_loops()
			_guess_light_tween.tween_method(_set_guess_light_phase, 0.0, 1.0, 0.5)
		"big":
			_guess_lights[0].texture = load(IMG + "light1.png")
			_guess_lights[1].texture = gray
			for s in _guess_lights:
				s.modulate = Color(1, 1, 1, 1.0)
		"small":
			_guess_lights[0].texture = gray
			_guess_lights[1].texture = load(IMG + "light2.png")
			for s in _guess_lights:
				s.modulate = Color(1, 1, 1, 1.0)

## 交替相位：t<0.5 红亮绿灰，t>=0.5 红灰绿亮（goslot guessing 轨道互换）
func _set_guess_light_phase(t: float) -> void:
	var red: Texture2D = load(IMG + "light1.png")
	var green: Texture2D = load(IMG + "light2.png")
	var gray: Texture2D = load(IMG + "lightgray1.png")
	if t < 0.5:
		_guess_lights[0].texture = red
		_guess_lights[1].texture = gray
	else:
		_guess_lights[0].texture = gray
		_guess_lights[1].texture = green

func _on_cashout_pressed() -> void:
	cash_out()

func _on_gamble_pressed(guess_big: bool) -> void:
	if win_pool > 0 and state == State.IDLE_BET and not paused:
		gamble_big_small(guess_big)

func _on_repeat_pressed() -> void:
	repeat_last_bets()

## 比大小（goslot betbig/betsmall）：押全部奖池，红绿灯交替闪 1.2s → 停在开出侧
func gamble_big_small(guess_big: bool) -> void:
	if state != State.IDLE_BET or paused or win_pool <= 0:
		return
	state = State.GUESSING
	_set_locked(true)
	_gamble_info.text = "押上全部中奖 %d，开！" % win_pool
	_set_guess_lights("swap")                # 红绿交替闪
	var hop := 0.06
	for i in 20:                             # 1.2s 开牌闪烁
		await get_tree().create_timer(hop).timeout
		if _run_aborted:
			return
	var n := _rng.randi_range(1, 10)
	var sys_big := n >= 6                    # goslot boxNumber：1~5 小箱 / 6~10 大箱
	if sys_big == guess_big:
		win_pool *= 2
		_gamble_info.text = "开出 %d（%s）→ 押中！奖池翻倍 %d" % [n, "大" if sys_big else "小", win_pool]
	else:
		win_pool = 0
		_refresh_pool_label()
		_gamble_info.text = "开出 %d（%s）→ 未押中，奖池清零" % [n, "大" if sys_big else "小"]
	# 灯停在开出侧：大=左红灯亮 / 小=右绿灯亮
	_set_guess_lights("big" if sys_big else "small")
	state = State.IDLE_BET
	_set_locked(false)
	_refresh_all()

## 下分（goslot getscore）：中奖池滚入余额，池清空
func cash_out() -> void:
	if state != State.IDLE_BET or paused or win_pool <= 0:
		return
	var amount := win_pool
	win_pool = 0
	balance = clamp(balance + amount, 0, SlotConfig.MAX_BALANCE)
	_gamble_info.text = "下分 %d" % amount
	if not _io_silent:
		Sound.coin()
		save_balance()
	_pool_tween = null
	_refresh_pool_label()
	var tw := create_tween()
	tw.tween_method(func(v: float) -> void:
		_center.text = "💰 %d" % int(round(v)),
		float(balance - amount), float(balance), 0.4)
	tw.tween_callback(func() -> void:
		_balance_tween = null
		_refresh_all())
	_balance_tween = tw
	_stop_orb_blink()
	_refresh_all()

## 续押（goslot rebet）：完整复刻上一轮的押注符号与数量
## 余额不足以补齐差额 → 整体拒绝并提示需手动押币（不部分续押）
func repeat_last_bets() -> void:
	if state != State.IDLE_BET or paused:
		return
	var need := 0
	for si in _last_bets.size():
		need += maxi(0, int(_last_bets[si]) - int(_bets[si]))
	if need <= 0:
		_gamble_info.text = "当前押注已与上轮相同"
		return
	if need > balance:
		_gamble_info.text = "余额不够（续押需 %d，余额 %d），请手动押币" % [need, balance]
		return
	for si in _last_bets.size():
		var add := int(_last_bets[si]) - int(_bets[si])
		if add > 0:
			_bets[si] = int(_bets[si]) + add
			balance -= add
	_stop_orb_blink()
	_finish_balance_roll()
	_refresh_all()
	_play("ding1")

# === 比大小（goslot RoundGuess/GuessTaiSai：开 1~10，1~5 小箱 / 6~10 大箱，同箱即赢 ×2）===
# 实现在 _on_gamble_pressed / gamble_big_small（黄面板区，约 774 行）

## 押注位：8 个热区 = 白格图标 + 下方黑格整体（y 1048..1180），点白块或黑块同效
## 图标（白格中心 1090）与实时押币数（黑格 y≈1136）都挂在热区按钮内
func _build_bet_pads() -> void:
	var cell_centers := [84.0, 188.0, 293.0, 397.0, 502.0, 606.0, 711.0, 813.0]
	for si in SlotConfig.BET_SYMBOLS.size():
		var sym: Dictionary = SlotConfig.BET_SYMBOLS[si]
		var bx: float = cell_centers[si] * 0.8                    # 屏幕 x（同板槽列对齐）
		# --- 热区：白格顶(1048) 到 黑格底(1180) ---
		var b := Button.new()
		b.name = "Bet_" + str(sym["key"])
		b.focus_mode = Control.FOCUS_NONE
		b.flat = true
		b.size = Vector2(90, 132)
		b.position = Vector2(bx - 45, 1048.0)
		var icon := TextureRect.new()
		icon.texture = load(SlotConfig.icon_path(_flag_to_ring_idx(str(sym["key"]))))
		icon.expand_mode = TextureRect.EXPAND_IGNORE_SIZE
		icon.stretch_mode = TextureRect.STRETCH_KEEP_ASPECT_CENTERED
		icon.size = Vector2(60, 64)
		icon.position = Vector2(15, 21)          # 白格中心 (45, 42) 对齐
		icon.mouse_filter = Control.MOUSE_FILTER_IGNORE
		b.add_child(icon)
		var odd_lbl := Label.new()
		odd_lbl.name = "Odd"
		odd_lbl.text = "×%d" % int(sym["odd"])
		odd_lbl.add_theme_font_size_override("font_size", 17)
		odd_lbl.add_theme_color_override("font_color", Color(0.5, 0.05, 0.05))
		odd_lbl.position = Vector2(8, 4)
		odd_lbl.mouse_filter = Control.MOUSE_FILTER_IGNORE
		b.add_child(odd_lbl)
		# 黑格押币数（按钮内子节点，相对热区顶部 y=1136-1048=88）
		var cnt := Label.new()
		cnt.name = "BetCnt_" + str(sym["key"])
		cnt.text = ""
		cnt.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
		cnt.add_theme_font_size_override("font_size", 26)
		cnt.add_theme_color_override("font_color", Color(1.0, 0.85, 0.2))
		cnt.add_theme_color_override("font_outline_color", Color(0.2, 0, 0))
		cnt.add_theme_constant_override("outline_size", 4)
		cnt.size = Vector2(90, 44)
		cnt.position = Vector2(0, 88)
		cnt.mouse_filter = Control.MOUSE_FILTER_IGNORE
		b.add_child(cnt)
		_attach_long_press(b, si)
		_ui.add_child(b)
		_bet_btns.append(b)
		_bet_count_lbls.append(cnt)

func _attach_long_press(btn: Button, sym: int) -> void:
	var press_t := 0
	var long_fired := false
	btn.gui_input.connect(func(ev: InputEvent) -> void:
		if ev is InputEventMouseButton and ev.button_index == MOUSE_BUTTON_LEFT:
			if ev.pressed:
				press_t = Time.get_ticks_msec()
				long_fired = false
			else:
				if not long_fired:
					add_bet(sym)
				press_t = 0
		elif ev is InputEventMouseMotion and press_t > 0 and not long_fired:
			if Time.get_ticks_msec() - press_t >= 400:
				long_fired = true
				clear_bet(sym))

func _build_overlay() -> void:
	_overlay = Control.new()
	_overlay.name = "StartOverlay"
	_overlay.size = Vector2(720, 1560)
	_overlay.mouse_filter = Control.MOUSE_FILTER_STOP
	var dim := ColorRect.new()
	dim.color = ThemeTokens.color("ov_bg")
	dim.size = Vector2(720, 1560)
	dim.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_overlay.add_child(dim)
	var box := VBoxContainer.new()
	box.size = Vector2(640, 600)
	box.position = Vector2(40, 480)
	box.add_theme_constant_override("separation", 24)
	_overlay.add_child(box)
	var title := Label.new()
	title.text = "幸运轮盘"
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	title.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	title.add_theme_font_size_override("font_size", 64)
	title.add_theme_color_override("font_color", ThemeTokens.color("gold"))
	box.add_child(title)
	var d1 := Label.new()
	d1.text = "押注符号，跑灯停在对应格即按倍数赔付"
	d1.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	d1.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	d1.add_theme_font_size_override("font_size", 24)
	d1.add_theme_color_override("font_color", ThemeTokens.color("ink"))
	box.add_child(d1)
	var d2 := Label.new()
	d2.text = "灯停在大/小 Lucky 会散花，散到的格也计奖"
	d2.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	d2.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	d2.add_theme_font_size_override("font_size", 20)
	d2.add_theme_color_override("font_color", ThemeTokens.color("ink2"))
	box.add_child(d2)
	var go := Button.new()
	go.text = "▶ 开始游戏"
	go.focus_mode = Control.FOCUS_NONE
	go.custom_minimum_size = Vector2(300, 80)
	go.size_flags_horizontal = Control.SIZE_SHRINK_CENTER
	go.add_theme_font_size_override("font_size", 28)
	go.pressed.connect(func() -> void:
		_overlay.mouse_filter = Control.MOUSE_FILTER_IGNORE
		_overlay.visible = false)
	box.add_child(go)
	_ui.add_child(_overlay)

# === 刷新 ===

func _refresh_all() -> void:
	if _center != null:
		_center.text = "💰 %d" % balance
	_refresh_pool_label()
	if _spin_btn != null:
		_spin_btn.disabled = not can_spin()
	for i in _bet_btns.size():
		if _bet_btns[i] != null:
			# 押注反馈：押了该符号 → 按钮加白色高亮描边（_modulate 闪烁太花，用透明度）
			_bet_btns[i].modulate = Color(1, 1, 1, 1.0) if int(_bets[i]) > 0 else Color(1, 1, 1, 0.75)
		# 黑区实时押币数（押币即更新；结算清零后随 _refresh_all 变空白）
		if i < _bet_count_lbls.size() and _bet_count_lbls[i] != null:
			_bet_count_lbls[i].text = "%d" % int(_bets[i]) if int(_bets[i]) > 0 else ""
	if _relief_btn != null:
		# 救济：余额空 且 中奖池也空（池里有钱应先下分）
		_relief_btn.visible = balance < 1 and win_pool <= 0 and state == State.IDLE_BET and not _overlay.visible
	# 面板按钮状态（显式状态机）
	var guessing := state == State.GUESSING
	if _repeat_btn != null:
		_repeat_btn.disabled = guessing or state != State.IDLE_BET
	if _cashout_btn != null:
		_cashout_btn.disabled = guessing or win_pool <= 0 or state != State.IDLE_BET
	if _gamble_big_btn != null and _gamble_small_btn != null:
		# 大/小按钮本身不闪：奖池>0 可点，开牌中/无池禁用；闪烁全交给红绿灯
		var bettable := win_pool > 0 and state == State.IDLE_BET and not guessing
		_gamble_big_btn.disabled = not bettable
		_gamble_small_btn.disabled = not bettable
		if not bettable and not guessing and state != State.RUNNING:
			for b in [_gamble_big_btn, _gamble_small_btn]:
				b.modulate = Color(1, 1, 1, 1.0)

func _set_locked(v: bool) -> void:
	if _spin_btn != null:
		_spin_btn.disabled = v or not can_spin()
	for b in _bet_btns:
		b.disabled = v


## 押注位图标取该 flag 的第一个大图标格 index
func _flag_to_ring_idx(flag: String) -> int:
	for i in 24:
		if SlotConfig.RING[i]["flag"] == flag and bool(SlotConfig.RING[i]["big"]):
			return i
	return 0
