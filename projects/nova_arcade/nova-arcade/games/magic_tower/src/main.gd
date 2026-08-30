extends Node2D
## 魔塔 — CR-6 T1 移植（magic-tower-godot 4.5 → 4.7.2，接入盒子）
## 逻辑层零改动：calc_damage / MON / ITEM / 网格 / 战斗 / 拾取 / MT_SELFTEST 断言全保留
## UI 层适配盒子 720×1560 竖屏：棋盘 640×640 居中，顶部 HUD 区，底部十字盘 + A/B/C/D（TouchView）
## 协议：adapter 注入 SAVE_PATH；死亡发 run_over(stats)；set_paused 供 adapter pause/resume 转发
## 输入：键盘方向键/WASD + 触屏十字盘/A-B-C-D（TouchView）

# ===== 纯逻辑常量（零改动，MT_SELFTEST 覆盖）=====
const TILE := 32
const MAP := 640
const W := 20
const H := 20
const STEP_MS := 0.16          # 触屏十字盘连走间隔（秒）
const LERP_SPEED := 14.0       # 玩家平滑到目标格的速度

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

# ===== 盒子布局（720×1560 竖屏，棋盘 640×640）=====
const BOARD_OFFSET := Vector2(40, 170)   # 棋盘左上角（视口坐标）
const ASSET_DIR := "res://games/magic_tower/assets/"

# ===== 信号 =====
## 本局结束（死亡），stats={score,gold}
signal run_over(stats: Dictionary)

# ===== 状态（零改动）=====
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

# ===== 盒子协议状态 =====
var paused := false
var dead := false
## 存档路径（adapter boot 注入 ctx.save_dir + "save.cfg"；独立运行默认值）
var SAVE_PATH := "user://save.cfg"

# ===== 节点引用 =====
var board: Node2D = null             # 棋盘容器（BOARD_OFFSET 偏移，内部坐标=原工程坐标）
var player: Sprite2D = null
var view: Node = null                # TouchView（HUD/十字盘/按钮）

func _ready() -> void:
	board = $Board
	_build_grid()
	_place_entities()
	_build_player()
	view = $TouchView
	view.attach(self)
	if OS.get_environment("MT_SELFTEST") == "1":
		_selftest()
		get_tree().quit()

# ---- 地图网格（与烘焙的 map.png 完全一致，零改动）----
func _build_grid() -> void:
	grid.clear()
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
	map.texture = load(ASSET_DIR + "map.png")
	map.centered = false
	board.add_child(map)

	var mon_at := { Vector2i(10, 3): "slime", Vector2i(15, 8): "bat",
			Vector2i(4, 10): "goblin", Vector2i(9, 12): "skeleton",
			Vector2i(16, 15): "zombie", Vector2i(2, 6): "wolfman" }
	for t in mon_at:
		monsters[t] = mon_at[t]
		var s = Sprite2D.new()
		s.texture = load("%s%s.png" % [ASSET_DIR, mon_at[t]])
		s.position = tile_center(t.x, t.y)
		board.add_child(s)
		mon_sprites[t] = s

	var item_at := { Vector2i(3, 4): "potion", Vector2i(16, 4): "sword",
			Vector2i(10, 9): "coins", Vector2i(15, 16): "shield",
			Vector2i(4, 16): "chest" }
	for t in item_at:
		items[t] = item_at[t]
		var s = Sprite2D.new()
		s.texture = load("%s%s.png" % [ASSET_DIR, item_at[t]])
		s.position = tile_center(t.x, t.y)
		board.add_child(s)
		item_sprites[t] = s

func _build_player() -> void:
	if player != null and is_instance_valid(player):
		player.queue_free()
	player = Sprite2D.new()
	player.texture = load(ASSET_DIR + "hero.png")
	player.position = tile_center(ptx, pty)
	board.add_child(player)

# ===== 纯逻辑（零改动）=====
static func calc_damage(atk: int, defn: int) -> int:
	if atk <= defn:
		return 0
	var d = atk - defn
	return max(1, (d * d) / 100)

func tile_center(tx: int, ty: int) -> Vector2:
	return Vector2(tx * TILE + TILE / 2, ty * TILE + TILE / 2)

# ===== 移动（逐格，零改动）=====
func _step(dx: int, dy: int) -> void:
	if in_battle or paused or dead:
		return
	var nx = ptx + dx
	var ny = pty + dy
	if nx < 0 or ny < 0 or nx >= W or ny >= H:
		return
	var t = Vector2i(nx, ny)
	var c = grid[ny][nx]
	if c == "#":
		view.show_msg("撞墙了")
		return
	if monsters.has(t):
		_start_battle(t)
		return
	if items.has(t):
		_pickup(t)
	if c == "S":
		view.show_msg("到达楼梯！(单层演示)")
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

# ===== 战斗（零改动）=====
func _start_battle(t: Vector2i) -> void:
	in_battle = true
	b_mon = monsters[t]
	b_tile = t
	b_hp = MON[b_mon][0]
	view.show_msg("遭遇 %s！(空格/A 攻击)" % b_mon)
	_refresh_hud()

func _attack() -> void:
	if not in_battle or paused or dead:
		return
	var mdef = MON[b_mon][2]
	var matk = MON[b_mon][1]
	var pdmg = calc_damage(atk, mdef)
	b_hp -= pdmg
	view.show_battle("你造成 %d 伤害，%s 剩 %d HP" % [pdmg, b_mon, max(0, b_hp)])
	if b_hp <= 0:
		_win_battle()
		return
	var mdmg = calc_damage(matk, defn)
	hp -= mdmg
	if hp <= 0:
		hp = 0
		_lose_battle()
		return
	view.show_battle(view.battle_text() + "  | %s 反击 -%d" % [b_mon, mdmg])
	_refresh_hud()

func _win_battle() -> void:
	in_battle = false
	if monsters.has(b_tile):
		monsters.erase(b_tile)
		mon_sprites[b_tile].visible = false
	gold += MON[b_mon][3]
	view.show_msg("击败 %s！+%d 金币" % [b_mon, MON[b_mon][3]])
	_refresh_hud()

## 本局结束：死亡 → 强写 best + 发 run_over（adapter 装配 quit_requested）
func _lose_battle() -> void:
	in_battle = false
	dead = true
	view.show_msg("你被 %s 击败了…(回菜单结束本局)" % b_mon)
	_refresh_hud()
	_save_best()
	run_over.emit({"score": gold, "gold": gold})

func _save_best() -> void:
	var cfg := ConfigFile.new()
	cfg.load(SAVE_PATH)
	var best: int = cfg.get_value("tower", "best", 0)
	if gold > best:
		cfg.set_value("tower", "best", gold)
		cfg.save(SAVE_PATH)

# ===== 道具（零改动）=====
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
	view.show_msg("拾取 %s（%s+%d）" % [id, stat, amt])
	_refresh_hud()

func _refresh_hud() -> void:
	view.show_stats("HP %d/%d  ATK %d  DEF %d  Gold %d" % [hp, hp_max, atk, defn, gold])

# ===== 盒子协议（adapter 调用入口）=====
## 暂停/恢复（adapter pause/resume 转发；R-11 只约定接口）
func set_paused(v: bool) -> void:
	paused = v
	view.show_msg("已暂停" if v else "继续")

## 重开一局（Launcher.play_again 复用模块不重建）
func reset_run() -> void:
	hp = 100; atk = 5; defn = 5; gold = 0; hp_max = 100
	ptx = 3; pty = 17
	in_battle = false; b_mon = ""; b_hp = 0
	paused = false; dead = false
	monsters.clear(); items.clear()
	for s in mon_sprites.values():
		if is_instance_valid(s):
			s.queue_free()
	for s in item_sprites.values():
		if is_instance_valid(s):
			s.queue_free()
	mon_sprites.clear(); item_sprites.clear()
	_place_entities()
	_build_player()
	view.show_msg("重开一局")
	_refresh_hud()

# ===== 输入（键盘，零改动）=====
func _unhandled_input(event: InputEvent) -> void:
	var e = event as InputEventKey
	if e != null and e.pressed:
		if in_battle:
			# 空格/回车/J 攻击；方向键也视为攻击（防「方向键失灵」体感：战斗中只认空格）
			match e.keycode:
				KEY_SPACE, KEY_ENTER, KEY_J, \
				KEY_LEFT, KEY_A, KEY_RIGHT, KEY_D, KEY_UP, KEY_W, KEY_DOWN, KEY_S:
					_attack()
			return
		match e.keycode:
			KEY_LEFT, KEY_A: _step(-1, 0)
			KEY_RIGHT, KEY_D: _step(1, 0)
			KEY_UP, KEY_W: _step(0, -1)
			KEY_DOWN, KEY_S: _step(0, 1)

# ===== 帧循环（触屏方向由 TouchView 写入 view.tdir）=====
var step_acc := 0.0

func _process(delta: float) -> void:
	if OS.get_environment("MT_SELFTEST") == "1":
		return
	if paused or dead:
		return
	player.position = player.position.lerp(tile_center(ptx, pty), minf(1.0, delta * LERP_SPEED))
	var tdir: Vector2 = view.tdir if view != null else Vector2.ZERO
	if not in_battle and tdir != Vector2.ZERO:
		step_acc += delta
		if step_acc >= STEP_MS:
			step_acc = 0.0
			_step_from_vec(tdir)
	view.update_dpad(tdir if tdir != Vector2.ZERO else _kb_dir())

# 键盘方向（用于十字盘高亮）
func _kb_dir() -> Vector2:
	var v := Vector2.ZERO
	if Input.is_key_pressed(KEY_LEFT) or Input.is_key_pressed(KEY_A): v.x -= 1.0
	if Input.is_key_pressed(KEY_RIGHT) or Input.is_key_pressed(KEY_D): v.x += 1.0
	if Input.is_key_pressed(KEY_UP) or Input.is_key_pressed(KEY_W): v.y -= 1.0
	if Input.is_key_pressed(KEY_DOWN) or Input.is_key_pressed(KEY_S): v.y += 1.0
	return v

# ===== 无头自检（MT_SELFTEST=1，断言零改动）=====
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
