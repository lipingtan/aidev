class_name SnakeGame extends CanvasLayer
## 贪吃蛇核心逻辑与视图（CR-6 T3）：纯函数层 snake_step/is_collision/spawn_food/move_interval（static，headless 可单测）
## 状态机 IDLE→PLAYING⇄PAUSED→OVER；输入=键盘方向键+右下虚拟十字盘（复用魔塔 Phoenix 贴图）
## 视口 720×1560 portrait；每吃 5 个食物提速一档；存档 save.cfg [nova] section（tetra 同构）
## 依赖：仅 ThemeTokens + dpad 贴图；不引用 Shell 任何类型

signal on_game_over(stats: Dictionary)
signal on_restarted

enum State { IDLE, PLAYING, PAUSED, OVER }

const GRID := 14
const CELL := 48.0
const SPAWN_LEN := 3
const SPEEDUP_EVERY := 5
const BASE_INTERVAL_MS := 260.0
const MIN_INTERVAL_MS := 90.0
const STEP_MS_PER_FOOD := 17.0
const VIEWPORT_W := 720.0
const VIEWPORT_H := 1560.0
const DPAD_CENTER := Vector2(560, 1340)
const DPAD_RADIUS := 96.0

var SAVE_PATH := "user://save.cfg"
var state: State = State.IDLE
var body: Array[Vector2i] = []
var dir := Vector2i(1, 0)
var food_pos := Vector2i.ZERO
var foods_eaten: int = 0
var best_score: int = 0
var rng: RandomNumberGenerator = RandomNumberGenerator.new()

var _cells: Array[ColorRect] = []
var _food_rect: ColorRect
var _score_lbl: Label
var _best_lbl: Label
var _overlay: Control
var _over_score_lbl: Label
var _restart_btn: Button
var _dpad: Sprite2D
var _dpad_tex: Dictionary = {}
var _dpad_cur := ""
var _acc: float = 0.0

func _ready() -> void:
	_load_save()
	_build_view()
	start_game()

## 开始/重start 一局（OVER 后由 restart 调用）
func start_game() -> void:
	body.clear()
	for i in SPAWN_LEN:
		body.append(Vector2i(6 - i, 7))
	dir = Vector2i(1, 0)
	foods_eaten = 0
	food_pos = spawn_food(body, rng)
	state = State.PLAYING
	_overlay.visible = false
	_refresh_all()

## 重start 一局（adapter reset_run 复用模块不重建）
func restart() -> void:
	start_game()
	on_restarted.emit()

## 暂停/resume（adapter pause/resume 转发）
func set_paused(v: bool) -> void:
	if v and state == State.PLAYING:
		state = State.PAUSED
	elif not v and state == State.PAUSED:
		state = State.PLAYING

## 纯函数：一步蛇移动。返回 {body, ate, dead}；撞墙/自撞 → dead（尾部让位规则：未吃时尾格可被头进入）
static func snake_step(body_in: Array[Vector2i], d: Vector2i, food: Vector2i, n: int) -> Dictionary:
	var head := body_in[0] + d
	var ate := head == food
	var new_body: Array[Vector2i] = []
	for i in range(0 if ate else 1, body_in.size()):
		new_body.append(body_in[i])
	new_body.push_front(head)
	return {"body": new_body, "ate": ate, "dead": is_collision(new_body, n)}

## 纯函数：撞墙/自撞判定（含边界）
static func is_collision(body_arr: Array[Vector2i], n: int) -> bool:
	var h := body_arr[0]
	if h.x < 0 or h.y < 0 or h.x >= n or h.y >= n:
		return true
	for i in range(1, body_arr.size()):
		if body_arr[i] == h:
			return true
	return false

## 纯函数：随机空位放食物（RNG 可注入，headless 确定性）；无空位 → Vector2i(-1,-1)
static func spawn_food(body_arr: Array[Vector2i], r: RandomNumberGenerator) -> Vector2i:
	var empties: Array[Vector2i] = []
	for i in GRID:
		for j in GRID:
			var p := Vector2i(i, j)
			if not body_arr.has(p):
				empties.append(p)
	if empties.is_empty():
		return Vector2i(-1, -1)
	return empties[r.randi_range(0, empties.size() - 1)]

## 纯函数：移动间隔随食物数递减（每 SPEEDUP_EVERY 个提速一档，floor MIN）
static func move_interval(eaten: int) -> float:
	return maxf(MIN_INTERVAL_MS, BASE_INTERVAL_MS - eaten * STEP_MS_PER_FOOD)

func _process(delta: float) -> void:
	if state != State.PLAYING:
		return
	_acc += delta * 1000.0
	var interval := move_interval(foods_eaten)
	while _acc >= interval and state == State.PLAYING:
		_acc -= interval
		_step()

func _step() -> void:
	var r := snake_step(body, dir, food_pos, GRID)
	body = r["body"] as Array[Vector2i]
	if bool(r["ate"]):
		foods_eaten += 1
		food_pos = spawn_food(body, rng)
	_refresh_all()
	if bool(r["dead"]):
		state = State.OVER
		_save_best()
		on_game_over.emit({"score": score(), "best": best_score, "length": body.size()})
		_show_over()

## 本局得分 = 蛇身长度 - 初始长度（T3 AC：score=本局长度）
func score() -> int:
	return maxi(0, body.size() - SPAWN_LEN)

## 键盘方向键（虚拟 D-pad 走 _input）
func _unhandled_input(event: InputEvent) -> void:
	if event is InputEventKey and event.pressed and not event.echo:
		var d := Vector2i.ZERO
		match (event as InputEventKey).keycode:
			KEY_UP: d = Vector2i(0, -1)
			KEY_DOWN: d = Vector2i(0, 1)
			KEY_LEFT: d = Vector2i(-1, 0)
			KEY_RIGHT: d = Vector2i(1, 0)
		if d != Vector2i.ZERO:
			_set_dir(d)

## 触屏虚拟 D-pad：十字盘区域内触点 → 方向（对角 snaps to dominant axis）
func _input(event: InputEvent) -> void:
	if event is InputEventScreenTouch and event.pressed:
		var t := event as InputEventScreenTouch
		var off := t.position - DPAD_CENTER
		if off.length() <= DPAD_RADIUS and off.length() > 12.0:
			var d := Vector2i(signi(int(off.x)), signi(int(off.y)))
			if d != Vector2i.ZERO:
				_set_dir(Vector2i(d.x, 0)) if absf(off.x) >= absf(off.y) else _set_dir(Vector2i(0, d.y))

## 方向不可反向（当前向上则 press downward ignored）
func _set_dir(d: Vector2i) -> void:
	if body.size() > 1 and d == -dir:
		return
	dir = d
	_update_dpad_tex()

## 视图构建：背景板 + HUD + 网格 + 食物 + 十字盘 + death overlay（all ThemeTokens，no hardcoded colors）
func _build_view() -> void:
	# 全屏不透明背景板（最底层）：游戏态 BgLayer 隐藏，无此板则露出 root 深灰清屏色（黑屏感）
	var bg := ColorRect.new()
	bg.color = ThemeTokens.color("bg")
	bg.size = Vector2(VIEWPORT_W, VIEWPORT_H)
	add_child(bg)
	var top := HBoxContainer.new()
	top.position = Vector2(30, 96)
	_score_lbl = _mk_label("SCORE 0", 26, ThemeTokens.color("ink"))
	_best_lbl = _mk_label("BEST 0", 26, ThemeTokens.color("gold"))
	var spacer := Control.new()
	spacer.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	top.add_child(_score_lbl)
	top.add_child(spacer)
	top.add_child(_best_lbl)
	add_child(top)

	var board := GRID * CELL
	# 棋盘容器：普通 Control + 手动 position 布局。
	# 不能用 GridContainer：容器会忽略子节点手动 position，把所有 cell 叠到同一点
	# （GUI 截图实测：棋盘格全部重叠、只剩裸数字，即此因）
	var grid_node := Control.new()
	grid_node.position = Vector2((VIEWPORT_W - board) / 2.0, VIEWPORT_H * 0.30)
	for i in GRID:
		for j in GRID:
			var cell := ColorRect.new()
			cell.size = Vector2(CELL, CELL)
			cell.position = Vector2(j * CELL, i * CELL)
			cell.color = ThemeTokens.color("card")
			grid_node.add_child(cell)
			_cells.append(cell)
	add_child(grid_node)

	var fx := (VIEWPORT_W - board) / 2.0
	var fy := VIEWPORT_H * 0.30
	_food_rect = ColorRect.new()
	_food_rect.size = Vector2(CELL - 8, CELL - 8)
	_food_rect.color = ThemeTokens.color("gold")
	_food_rect.position = Vector2(fx + food_pos.x * CELL + 4, fy + food_pos.y * CELL + 4)
	add_child(_food_rect)

	_build_dpad()
	_overlay = _build_over()
	_overlay.visible = false
	add_child(_overlay)

## 十字盘（reuse magic-tower Phoenix dpad textures；9-way switch）
func _build_dpad() -> void:
	for n in ["none", "up", "down", "left", "right"]:
		_dpad_tex[n] = load("res://games/snake/assets/dpad/dpad_%s.png" % n)
	_dpad = Sprite2D.new()
	_dpad.texture = _dpad_tex["none"]
	_dpad.centered = true
	_dpad.position = DPAD_CENTER
	var s := 190.0 / (_dpad_tex["none"] as Texture2D).get_size().x
	_dpad.scale = Vector2(s, s)
	_dpad.texture_filter = CanvasItem.TEXTURE_FILTER_LINEAR
	add_child(_dpad)

func _update_dpad_tex() -> void:
	var st := "none"
	if dir == Vector2i(0, -1):
		st = "up"
	elif dir == Vector2i(0, 1):
		st = "down"
	elif dir == Vector2i(-1, 0):
		st = "left"
	elif dir == Vector2i(1, 0):
		st = "right"
	if st != _dpad_cur:
		_dpad_cur = st
		_dpad.texture = _dpad_tex[st]

## death overlay：dark base + score + play again（adapter side adds "return to menu"）
func _build_over() -> Control:
	var ov := Control.new()
	ov.set_anchors_preset(Control.PRESET_FULL_RECT)
	var dim := ColorRect.new()
	dim.color = ThemeTokens.color("ov_bg")
	dim.set_anchors_preset(Control.PRESET_FULL_RECT)
	ov.add_child(dim)
	var box := VBoxContainer.new()
	box.alignment = BoxContainer.ALIGNMENT_CENTER
	box.set_anchors_preset(Control.PRESET_FULL_RECT)
	box.add_child(_mk_label("GAME OVER", 44, ThemeTokens.color("ink")))
	_over_score_lbl = _mk_label("", 26, ThemeTokens.color("gold"))
	box.add_child(_over_score_lbl)
	_restart_btn = Button.new()
	_restart_btn.focus_mode = Control.FOCUS_NONE   # 防抢键盘焦点（方向键被吞）
	_restart_btn.text = "↻ Play Again"
	_restart_btn.add_theme_font_size_override("font_size", 20)
	_restart_btn.custom_minimum_size = Vector2(240, 56)
	_restart_btn.pressed.connect(restart)
	box.add_child(_restart_btn)
	ov.add_child(box)
	return ov

func _mk_label(text: String, size: int, col: Color) -> Label:
	var l := Label.new()
	l.text = text
	l.add_theme_font_size_override("font_size", size)
	l.add_theme_color_override("font_color", col)
	return l

func _refresh_all() -> void:
	for i in GRID:
		for j in GRID:
			_cells[i * GRID + j].color = ThemeTokens.color("card")
	for k in body.size():
		var seg := body[k]
		if seg.x >= 0 and seg.y >= 0 and seg.x < GRID and seg.y < GRID:
			_cells[seg.y * GRID + seg.x].color = ThemeTokens.color("accent") if k == 0 else ThemeTokens.color("alt")
	var board_x := (VIEWPORT_W - GRID * CELL) / 2.0
	var board_y := VIEWPORT_H * 0.30
	_food_rect.position = Vector2(board_x + food_pos.x * CELL + 4, board_y + food_pos.y * CELL + 4)
	_score_lbl.text = "SCORE %d" % score()
	_best_lbl.text = "BEST %d" % best_score

func _show_over() -> void:
	_over_score_lbl.text = "This run score %d · all-time high %d" % [score(), best_score]
	_overlay.visible = true

## save (tetra isomorphic: ConfigFile [nova] section; best_score force-written immediately to disk)
func _load_save() -> void:
	var cfg := ConfigFile.new()
	if cfg.load(SAVE_PATH) == OK:
		best_score = int(cfg.get_value("nova", "best", 0))

func _save_best() -> void:
	if score() > best_score:
		best_score = score()
	var cfg := ConfigFile.new()
	cfg.set_value("nova", "best", best_score)
	cfg.set_value("nova", "last_played", Time.get_datetime_string_from_system())
	cfg.save(SAVE_PATH)
