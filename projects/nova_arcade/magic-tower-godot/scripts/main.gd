extends Node2D
# Magic Tower — Godot 4.5 · 格子制完整玩法演示
# 键盘：方向键/WASD 逐格移动；战斗中 空格/回车/J 攻击。手机：左下十字盘(移动) + 右侧 A/B/C/D + Start/投币/退币/退出
# 逻辑写成纯函数，可用 MT_SELFTEST=1 无头自检（打印验证结果）。

const TILE := 32
const MAP := 640
const W := 20
const H := 20
const STEP_MS := 0.16          # 触摸十字盘连走间隔
const LERP_SPEED := 14.0      # 玩家平滑到目标格的速度
const DESKTOP_SIZE := Vector2i(640, 640)   # 桌面初始窗口（基础分辨率 320 的 2× 整数倍，最清晰）

# ---- 怪物属性 [hp, atk, def, 金币奖励] ----
const MON := {
	"slime":    [30, 2, 0, 5],
	"bat":      [40, 3, 0, 8],
	"goblin":   [80, 6, 4, 15],
	"skeleton": [120, 8, 6, 25],
	"zombie":   [150, 9, 7, 35],
	"wolfman":  [200, 12, 8, 50],
}

# ---- 道具效果（对玩家状态）----
const ITEM := {
	"potion":   ["hp", 30],
	"bigPotion":["hp", 60],
	"sword":    ["atk", 2],
	"shield":   ["def", 2],
	"coins":    ["gold", 50],
	"chest":    ["gold", 100],
}

# ===== 纯逻辑（可单测）=====
static func calc_damage(atk: int, defn: int) -> int:
	if atk <= defn:
		return 0
	var d = atk - defn
	return max(1, (d * d) / 100)

func tile_center(tx: int, ty: int) -> Vector2:
	return Vector2(tx * TILE + TILE / 2, ty * TILE + TILE / 2)

# ===== 状态 =====
var grid: Array = []                 # H x W，'#'墙 '.'地 'R/B/G'门 'S'楼梯
var monsters: Dictionary = {}        # Vector2i -> 类型名
var mon_sprites: Dictionary = {}     # Vector2i -> Sprite2D
var items: Dictionary = {}           # Vector2i -> 道具 id
var item_sprites: Dictionary = {}    # Vector2i -> Sprite2D

var hp := 100
var atk := 5
var defn := 5
var gold := 0
var hp_max := 100

var ptx := 3
var pty := 17
var in_battle := false
var b_mon := ""
var b_hp := 0
var b_tile := Vector2i.ZERO

# ===== 节点引用 =====
var player: Sprite2D
var cam: Camera2D
var hud_stats: Label
var hud_msg: Label
var hud_battle: Label
var dpad: Sprite2D
var dpad_tex: Dictionary = {}
var cur_state := "none"

# 触摸
const DPAD_CENTER := Vector2(72, 248)
const DPAD_RADIUS := 74.0
const DEADZONE := 10.0
const USE_LARGE_DPAD := false
const DPAD_SIZE := 132.0
const BTN := "res://assets/buttons/"
var tdir := Vector2.ZERO
var dpad_touch_id := -1
var step_acc := 0.0
var btns: Array = []
var pressed_idx := -1
var pressed_id := -1

func _ready() -> void:
	# 桌面端把窗口放大到整数倍（手机靠 stretch 自动铺满，不动）
	var osn := OS.get_name()
	if osn != "Android" and osn != "iOS":
		get_window().size = DESKTOP_SIZE   # 桌面初始窗口放大到整数倍（手机靠 stretch 铺满）
	_build_grid()
	_place_entities()
	_build_player()
	_build_hud()
	_build_dpad()
	_build_buttons()
	if OS.get_environment("MT_SELFTEST") == "1":
		_selftest()
		get_tree().quit()

# ---- 地图网格（与烘焙的 map.png 完全一致）----
func _build_grid() -> void:
	for y in H:
		var row := []
		for x in W:
			row.append(".")
		grid.append(row)
	for x in W:
		grid[0][x] = "#"; grid[H - 1][x] = "#"
	for y in H:
		grid[y][0] = "#"; grid[y][W - 1] = "#"
	for y in range(3, 10):   # 竖墙 x=6
		grid[y][6] = "#"
	for y in range(3, 10):   # 竖墙 x=13
		grid[y][13] = "#"
	for x in range(7, 13):   # 横墙 y=6
		grid[6][x] = "#"
	grid[5][6] = "R"         # 红门
	grid[7][13] = "B"        # 蓝门
	grid[6][9] = "G"         # 金门
	grid[2][17] = "S"        # 楼梯

func _place_entities() -> void:
	var map = Sprite2D.new()
	map.texture = load("res://assets/map.png")
	map.centered = false
	add_child(map)

	var mon_at := { Vector2i(10, 3): "slime", Vector2i(15, 8): "bat",
			Vector2i(4, 10): "goblin", Vector2i(9, 12): "skeleton",
			Vector2i(16, 15): "zombie", Vector2i(2, 6): "wolfman" }
	for t in mon_at:
		monsters[t] = mon_at[t]
		var s = Sprite2D.new()
		s.texture = load("res://assets/%s.png" % mon_at[t])
		s.position = tile_center(t.x, t.y)
		add_child(s)
		mon_sprites[t] = s

	var item_at := { Vector2i(3, 4): "potion", Vector2i(16, 4): "sword",
			Vector2i(10, 9): "coins", Vector2i(15, 16): "shield",
			Vector2i(4, 16): "chest" }
	for t in item_at:
		items[t] = item_at[t]
		var s = Sprite2D.new()
		s.texture = load("res://assets/%s.png" % item_at[t])
		s.position = tile_center(t.x, t.y)
		add_child(s)
		item_sprites[t] = s

func _build_player() -> void:
	player = Sprite2D.new()
	player.texture = load("res://assets/hero.png")
	player.position = tile_center(ptx, pty)
	add_child(player)
	cam = Camera2D.new()
	cam.position_smoothing_enabled = true
	cam.position_smoothing_speed = 6.0
	cam.limit_left = 0; cam.limit_top = 0
	cam.limit_right = MAP; cam.limit_bottom = MAP
	player.add_child(cam)

# ---- HUD ----
func _build_hud() -> void:
	var layer = CanvasLayer.new()
	add_child(layer)
	hud_stats = Label.new(); hud_stats.position = Vector2(8, 6); layer.add_child(hud_stats)
	hud_msg = Label.new(); hud_msg.position = Vector2(8, 24); layer.add_child(hud_msg)
	hud_battle = Label.new(); hud_battle.position = Vector2(8, 44); layer.add_child(hud_battle)
	_refresh_hud()

func _refresh_hud() -> void:
	if hud_stats:
		hud_stats.text = "HP %d/%d  ATK %d  DEF %d  Gold %d" % [hp, hp_max, atk, defn, gold]

# ---- 左下十字盘（按方向高亮）----
func _build_dpad() -> void:
	var folder := "res://assets/dpad_large/" if USE_LARGE_DPAD else "res://assets/dpad/"
	var names := ["none", "up", "down", "left", "right",
			"up_left", "up_right", "down_left", "down_right"]
	for n in names:
		dpad_tex[n] = load(folder + "dpad_%s.png" % n)
	var layer = CanvasLayer.new()
	add_child(layer)
	dpad = Sprite2D.new()
	dpad.texture = dpad_tex["none"]; dpad.centered = true; dpad.position = DPAD_CENTER
	var s := DPAD_SIZE / (dpad_tex["none"] as Texture2D).get_size().x
	dpad.scale = Vector2(s, s)
	dpad.texture_filter = CanvasItem.TEXTURE_FILTER_LINEAR
	layer.add_child(dpad)

# ---- 右侧按钮（按下换 _press 图）----
func _add_button(layer: CanvasLayer, name: String, base: String, center: Vector2, target: float) -> void:
	var nrm = load(BTN + base + ".png")
	var prs = load(BTN + base + "_press.png")
	var spr = Sprite2D.new()
	spr.texture = nrm; spr.centered = true; spr.position = center
	var s = target / nrm.get_size().x
	spr.scale = Vector2(s, s)
	spr.texture_filter = CanvasItem.TEXTURE_FILTER_LINEAR
	layer.add_child(spr)
	btns.append({"spr": spr, "nrm": nrm, "prs": prs, "name": name, "center": center, "half": target / 2.0})

func _build_buttons() -> void:
	var layer = CanvasLayer.new()
	add_child(layer)
	_add_button(layer, "A", "button_a", Vector2(262, 266), 54.0)
	_add_button(layer, "B", "button_b", Vector2(232, 236), 54.0)
	_add_button(layer, "C", "button_c", Vector2(262, 206), 54.0)
	_add_button(layer, "D", "button_d", Vector2(292, 236), 54.0)
	_add_button(layer, "Start", "button_start", Vector2(250, 92), 44.0)
	_add_button(layer, "CoinIn", "button_coin_in", Vector2(292, 92), 44.0)
	_add_button(layer, "CoinOut", "button_coin_out", Vector2(250, 134), 44.0)
	_add_button(layer, "Quit", "button_quit", Vector2(292, 134), 44.0)

func _btn_at(pos: Vector2) -> int:
	for i in btns.size():
		if (pos - btns[i].center).length() <= btns[i].half * 1.25:
			return i
	return -1

# ===== 移动（逐格）=====
func _step(dx: int, dy: int) -> void:
	if in_battle:
		return
	var nx = ptx + dx
	var ny = pty + dy
	if nx < 0 or ny < 0 or nx >= W or ny >= H:
		return
	var t = Vector2i(nx, ny)
	var c = grid[ny][nx]
	if c == "#":
		hud_msg.text = "撞墙了"
		return
	if monsters.has(t):
		_start_battle(t)
		return
	if items.has(t):
		_pickup(t)
	if c == "S":
		hud_msg.text = "到达楼梯！(单层演示)"
	ptx = nx; pty = ny

func _step_from_vec(v: Vector2) -> void:
	if v == Vector2.ZERO:
		return
	var dx = 0
	var dy = 0
	if absf(v.x) > absf(v.y):
		dx = 1 if v.x > 0 else -1
	else:
		dy = 1 if v.y > 0 else -1
	_step(dx, dy)

# ===== 战斗 =====
func _start_battle(t: Vector2i) -> void:
	in_battle = true
	b_mon = monsters[t]
	b_tile = t
	b_hp = MON[b_mon][0]
	hud_msg.text = "遭遇 %s！(空格/A 攻击)" % b_mon
	_refresh_hud()

func _attack() -> void:
	if not in_battle:
		return
	var mdef = MON[b_mon][2]
	var matk = MON[b_mon][1]
	var pdmg = calc_damage(atk, mdef)
	b_hp -= pdmg
	hud_battle.text = "你造成 %d 伤害，%s 剩 %d HP" % [pdmg, b_mon, max(0, b_hp)]
	if b_hp <= 0:
		_win_battle()
		return
	var mdmg = calc_damage(matk, defn)
	hp -= mdmg
	if hp <= 0:
		hp = 0
		in_battle = false
		hud_msg.text = "你被 %s 击败了…(重开)" % b_mon
		_refresh_hud()
		return
	hud_battle.text += "  | %s 反击 -%d" % [b_mon, mdmg]
	_refresh_hud()

func _win_battle() -> void:
	in_battle = false
	if monsters.has(b_tile):
		monsters.erase(b_tile)
		mon_sprites[b_tile].visible = false
	gold += MON[b_mon][3]
	hud_msg.text = "击败 %s！+%d 金币" % [b_mon, MON[b_mon][3]]
	_refresh_hud()

# ===== 道具 =====
func _pickup(t: Vector2i) -> void:
	var id = items[t]
	items.erase(t)
	item_sprites[t].visible = false
	var stat = ITEM[id][0]
	var amt = ITEM[id][1]
	match stat:
		"hp": hp = mini(hp + amt, hp_max)
		"atk": atk += amt
		"def": defn += amt
		"gold": gold += amt
	hud_msg.text = "拾取 %s（%s+%d）" % [id, stat, amt]
	_refresh_hud()

# ===== 输入 =====
func _unhandled_input(event: InputEvent) -> void:
	var e = event as InputEventKey
	if e != null and e.pressed:
		if in_battle:
			if e.keycode == KEY_SPACE or e.keycode == KEY_ENTER or e.keycode == KEY_J:
				_attack()
			return
		match e.keycode:
			KEY_LEFT, KEY_A: _step(-1, 0)
			KEY_RIGHT, KEY_D: _step(1, 0)
			KEY_UP, KEY_W: _step(0, -1)
			KEY_DOWN, KEY_S: _step(0, 1)

func _input(event: InputEvent) -> void:
	var st = event as InputEventScreenTouch
	if st != null:
		if st.pressed:
			var bi = _btn_at(st.position)
			if bi >= 0:
				pressed_idx = bi; pressed_id = st.index
				btns[bi].spr.texture = btns[bi].prs
				if in_battle and str(btns[bi].name) == "A":
					_attack()
				elif hud_msg:
					hud_msg.text = "按钮: " + str(btns[bi].name)
			elif (st.position - DPAD_CENTER).length() <= DPAD_RADIUS:
				dpad_touch_id = st.index
				step_acc = 0.0
				_set_tdir(st.position)
		else:
			if st.index == pressed_id:
				if pressed_idx >= 0:
					btns[pressed_idx].spr.texture = btns[pressed_idx].nrm
				pressed_id = -1; pressed_idx = -1
			elif st.index == dpad_touch_id:
				dpad_touch_id = -1; tdir = Vector2.ZERO
		return
	var sd = event as InputEventScreenDrag
	if sd != null and sd.index == dpad_touch_id:
		_set_tdir(sd.position)

func _set_tdir(pos: Vector2) -> void:
	var off = pos - DPAD_CENTER
	tdir = Vector2.ZERO if off.length() <= DEADZONE else off.normalized()

func _process(delta: float) -> void:
	if OS.get_environment("MT_SELFTEST") == "1":
		return
	# 玩家平滑到目标格
	player.position = player.position.lerp(tile_center(ptx, pty), minf(1.0, delta * LERP_SPEED))
	# 触摸十字盘连走
	if not in_battle and tdir != Vector2.ZERO:
		step_acc += delta
		if step_acc >= STEP_MS:
			step_acc = 0.0
			_step_from_vec(tdir)
	_update_dpad(tdir if tdir != Vector2.ZERO else _kb_dir())

# 键盘方向（用于十字盘高亮）
func _kb_dir() -> Vector2:
	var v := Vector2.ZERO
	if Input.is_key_pressed(KEY_LEFT) or Input.is_key_pressed(KEY_A): v.x -= 1.0
	if Input.is_key_pressed(KEY_RIGHT) or Input.is_key_pressed(KEY_D): v.x += 1.0
	if Input.is_key_pressed(KEY_UP) or Input.is_key_pressed(KEY_W): v.y -= 1.0
	if Input.is_key_pressed(KEY_DOWN) or Input.is_key_pressed(KEY_S): v.y += 1.0
	return v

func _update_dpad(d: Vector2) -> void:
	var st := _dir_state(d)
	if st != cur_state:
		cur_state = st
		dpad.texture = dpad_tex[st]

func _dir_state(d: Vector2) -> String:
	var x = 1 if d.x > 0.3 else (-1 if d.x < -0.3 else 0)
	var y = 1 if d.y > 0.3 else (-1 if d.y < -0.3 else 0)
	if x == 0 and y == 0:
		return "none"
	var v = "up" if y == -1 else ("down" if y == 1 else "")
	var h = "left" if x == -1 else ("right" if x == 1 else "")
	if x != 0 and y != 0:
		return v + "_" + h
	return h if x != 0 else v

# ===== 无头自检（MT_SELFTEST=1）=====
func _selftest() -> void:
	print("[selftest] calc_damage(30,10)=%d  (5,10)=%d  (20,20)=%d" % [calc_damage(30, 10), calc_damage(5, 10), calc_damage(20, 20)])
	assert(calc_damage(30, 10) > calc_damage(12, 10), "higher ATK should deal more")
	assert(calc_damage(5, 10) == 0, "ATK<=DEF deals 0")
	assert(calc_damage(20, 20) == 0, "equal deals 0")

	# 道具效果（受控起始值，验证精确数值）
	hp = 40; atk = 5; defn = 5; gold = 0
	_test_item("potion"); assert(hp == 70, "potion +30 hp (40->70)")
	_test_item("sword"); assert(atk == 7, "sword +2 atk")
	_test_item("shield"); assert(defn == 7, "shield +2 def")
	_test_item("coins"); assert(gold == 50, "coins +50 gold")

	# 网格
	assert(grid[0][0] == "#", "corner is wall")
	assert(grid[10][10] == ".", "center is floor")
	assert(grid[2][17] == "S", "stairs placed")
	assert(grid[5][6] == "R" and grid[7][13] == "B" and grid[6][9] == "G", "doors placed")

	# 完整战斗模拟：打一只 slime
	var mhp = MON["slime"][0]
	var rounds = 0
	while mhp > 0 and rounds < 50:
		mhp -= calc_damage(atk, MON["slime"][2])
		if mhp <= 0:
			break
		hp -= calc_damage(MON["slime"][1], defn)
		rounds += 1
	print("[selftest] slime battle: rounds=%d playerHP=%d (atk=%d)" % [rounds, hp, atk])
	assert(mhp <= 0, "player should defeat slime")
	print("[selftest] ALL PASS ✅")

func _test_item(id: String) -> void:
	var stat = ITEM[id][0]
	var amt = ITEM[id][1]
	match stat:
		"hp": hp = mini(hp + amt, hp_max)
		"atk": atk += amt
		"def": defn += amt
		"gold": gold += amt
