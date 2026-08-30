class_name Game2048 extends CanvasLayer
## 2048 核心逻辑与视图（CR-6 T2）：纯函数层 grid_merge/is_game_over/spawn_tile（static，headless 可单测）
## 状态机 IDLE→PLAYING⇄PAUSED→OVER；输入=键盘方向键+触屏滑动（design §3.2）；存档 save.cfg [nova] section（tetra 同构）
## 依赖：仅 ThemeTokens；不引用 Shell 任何类型

signal on_game_over(stats: Dictionary)
signal on_restarted

enum State { IDLE, PLAYING, PAUSED, OVER }

const GRID_SIZE := 4
const CELL := 150.0
const GAP := 12.0
const BOARD_PAD := 20.0
const TILE_MARGIN := 8.0
const SPAWN_V2_WEIGHT := 90
const VIEWPORT_W := 720.0
const VIEWPORT_H := 1560.0

## 触屏滑动参数（design R-2）
@export_group("触屏滑动")
@export var swipe_threshold: float = 30.0
@export var swipe_deadzone: float = 15.0

var SAVE_PATH := "user://save.cfg"
var state: State = State.IDLE
var grid: Array = _empty_grid()
var score: int = 0
var best_score: int = 0
var rng: RandomNumberGenerator = RandomNumberGenerator.new()

var _cells: Array[ColorRect] = []
var _tile_labels: Array[Label] = []
var _score_lbl: Label
var _best_lbl: Label
var _overlay: Control
var _start_overlay: Control
var _over_score_lbl: Label
var _restart_btn: Button
var _touch_active := false
var _touch_start := Vector2.ZERO

func _ready() -> void:
	_load_save()
	_build_view()
	_refresh_all()

## 开始/继续一局（首局开局；OVER 后由 restart 调用）
func start_game() -> void:
	grid = _empty_grid()
	score = 0
	spawn_tile(grid)
	spawn_tile(grid)
	state = State.PLAYING
	_overlay.visible = false
	_refresh_all()

## 重开一局（adapter reset_run 复用模块不重建）
func restart() -> void:
	start_game()
	on_restarted.emit()

## 暂停/恢复（adapter pause/resume 转发）
func set_paused(v: bool) -> void:
	if v and state == State.PLAYING:
		state = State.PAUSED
	elif not v and state == State.PAUSED:
		state = State.PLAYING

## 纯函数：按方向合并网格，返回 {grid, score_delta}（同值合并翻倍、每格至多一次）
static func grid_merge(g: Array, dir: Vector2i) -> Dictionary:
	var n := _empty_grid()
	var delta := 0
	for i in GRID_SIZE:
		var line: Array = []
		for c in GRID_SIZE:
			line.append(_cell_at(g, i, c, dir))
		var merged: Array = []
		var k := 0
		while k < line.size():
			if int(line[k]) == 0:
				k += 1
				continue
			if k + 1 < line.size() and int(line[k + 1]) == int(line[k]):
				delta += int(line[k]) * 2
				merged.append(int(line[k]) * 2)
				k += 2
			else:
				merged.append(int(line[k]))
				k += 1
		while merged.size() < GRID_SIZE:
			merged.append(0)
		for c in GRID_SIZE:
			var pos := _pos_for(i, c, dir)
			n[pos.y][pos.x] = int(merged[c])
	return {"grid": n, "score_delta": delta}

## 纯函数：沿移动方向读第 k 格（k=0 为移动方向最前端）
static func _cell_at(g: Array, i: int, k: int, dir: Vector2i) -> int:
	if dir.x != 0:
		return int(g[i][k if dir.x == -1 else GRID_SIZE - 1 - k])
	return int(g[k if dir.y == -1 else GRID_SIZE - 1 - k][i])

## 纯函数：沿移动方向第 k 格的写回位置（与 _cell_at 对称）
static func _pos_for(i: int, k: int, dir: Vector2i) -> Vector2i:
	if dir.x != 0:
		return Vector2i(k if dir.x == -1 else GRID_SIZE - 1 - k, i)
	return Vector2i(i, k if dir.y == -1 else GRID_SIZE - 1 - k)
## 纯函数：网格无空位且无任何可合并对 → 游戏结束
static func is_game_over(g: Array) -> bool:
	for i in GRID_SIZE:
		for j in GRID_SIZE:
			var v := int(g[i][j])
			if v == 0:
				return false
			if j + 1 < GRID_SIZE and int(g[i][j + 1]) == v:
				return false
			if i + 1 < GRID_SIZE and int(g[i + 1][j]) == v:
				return false
	return true
## 纯函数：随机空位生成 2（90%）/4（10%），返回是否成功
static func spawn_tile(g: Array, r: RandomNumberGenerator = null) -> bool:
	var empties: Array[Vector2i] = []
	for i in GRID_SIZE:
		for j in GRID_SIZE:
			if int(g[i][j]) == 0:
				empties.append(Vector2i(i, j))
	if empties.is_empty():
		return false
	var rr := r if r != null else RandomNumberGenerator.new()
	var idx := empties[rr.randi_range(0, empties.size() - 1)]
	g[idx.x][idx.y] = 4 if rr.randi() < 100 - SPAWN_V2_WEIGHT else 2
	return true

## 执行一次滑动（状态守卫 + 合并 + 生成 + 结束判定）
func _move(dir: Vector2i) -> void:
	if state != State.PLAYING:
		return
	var r := grid_merge(grid, dir)
	var new_grid: Array = r["grid"]
	if new_grid == grid:
		return
	grid = new_grid
	score += int(r["score_delta"])
	spawn_tile(grid, rng)
	if is_game_over(grid):
		state = State.OVER
		_save_best()
		on_game_over.emit({"score": score, "best": best_score, "max_tile": _max_tile()})
		_show_over()
	_refresh_all()

## 键盘方向键（触屏滑动走 _input）
func _unhandled_input(event: InputEvent) -> void:
	if event is InputEventKey and event.pressed and not event.echo:
		match (event as InputEventKey).keycode:
			KEY_UP: _move(Vector2i(0, -1))
			KEY_DOWN: _move(Vector2i(0, 1))
			KEY_LEFT: _move(Vector2i(-1, 0))
			KEY_RIGHT: _move(Vector2i(1, 0))

## 触屏滑动：位移过死区且方向占优时触发
func _input(event: InputEvent) -> void:
	if event is InputEventScreenTouch:
		var t := event as InputEventScreenTouch
		if t.pressed:
			_touch_active = true
			_touch_start = t.position
		elif _touch_active:
			_touch_active = false
			var d := t.position - _touch_start
			if absf(d.x) < swipe_deadzone and absf(d.y) < swipe_deadzone:
				return
			_move(Vector2i(signi(int(d.x)), 0)) if absf(d.x) > absf(d.y) else _move(Vector2i(0, signi(int(d.y))))

## 视图构建：背景板 + HUD + 4×4 网格 + 结束遮罩（全 ThemeTokens，禁硬编码色值）
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

	var board_size := GRID_SIZE * CELL + (GRID_SIZE - 1) * GAP + BOARD_PAD * 2.0
	# 棋盘容器：普通 Control + 手动 position 布局。
	# 不能用 GridContainer：容器会忽略子节点手动 position，把所有 cell 叠到同一点
	# （GUI 截图实测：棋盘格全部重叠，画面只剩裸数字）
	var grid_node := Control.new()
	grid_node.position = Vector2((VIEWPORT_W - board_size) / 2.0, VIEWPORT_H * 0.34)
	for i in GRID_SIZE:
		for j in GRID_SIZE:
			var cell := ColorRect.new()
			cell.size = Vector2(CELL, CELL)
			cell.position = Vector2(j * (CELL + GAP), i * (CELL + GAP))
			cell.color = ThemeTokens.color("card")
			var lbl := Label.new()
			lbl.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
			lbl.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
			lbl.add_theme_font_size_override("font_size", 34)
			lbl.size = Vector2(CELL - TILE_MARGIN * 2.0, CELL - TILE_MARGIN * 2.0)
			lbl.position = Vector2(TILE_MARGIN, TILE_MARGIN)
			cell.add_child(lbl)
			grid_node.add_child(cell)
			_cells.append(cell)
			_tile_labels.append(lbl)
	add_child(grid_node)
	_overlay = _build_over()
	_overlay.visible = false
	add_child(_overlay)
	_start_overlay = _build_start()
	add_child(_start_overlay)   # 启动时显示：用户点「开始游戏」才开局

## 开始遮罩：标题 + 玩法说明 + 开始按钮（boot 时覆盖，点击后 start_game）
func _build_start() -> Control:
	var ov := Control.new()
	ov.size = Vector2(VIEWPORT_W, VIEWPORT_H)   # CanvasLayer 下无 Control 父，anchors 无参照，手动定尺寸
	ov.mouse_filter = Control.MOUSE_FILTER_STOP
	var dim := ColorRect.new()
	dim.color = ThemeTokens.color("ov_bg")
	dim.size = Vector2(VIEWPORT_W, VIEWPORT_H)
	ov.add_child(dim)
	var box := VBoxContainer.new()
	box.alignment = BoxContainer.ALIGNMENT_CENTER
	box.size = Vector2(VIEWPORT_W, VIEWPORT_H)
	box.add_theme_constant_override("separation", 20)
	for it in [["2048", 56, ThemeTokens.color("gold")],
			["滑动或方向键合并相同数字", 20, ThemeTokens.color("ink")],
			["合成 2048 即获胜", 18, ThemeTokens.color("ink2")]]:
		var l := _mk_label(it[0], it[1], it[2])
		l.size_flags_horizontal = Control.SIZE_EXPAND_FILL   # 占满行宽，配合水平居中
		l.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
		box.add_child(l)
	var btn := Button.new()
	btn.text = "▶ 开始游戏"
	btn.focus_mode = Control.FOCUS_NONE
	btn.custom_minimum_size = Vector2(260, 64)
	btn.size_flags_horizontal = Control.SIZE_SHRINK_CENTER
	btn.add_theme_font_size_override("font_size", 24)
	btn.pressed.connect(func() -> void:
		ov.visible = false
		ov.mouse_filter = Control.MOUSE_FILTER_IGNORE
		start_game())
	box.add_child(btn)
	ov.add_child(box)
	return ov

## 结束遮罩：暗底 + 分数 + 再来一局（「回菜单」由 adapter 侧追加）
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
	_restart_btn.text = "↻ 再来一局"
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

## 网格值 → 主题色（空位用 card；数值越高越亮）
func _tile_color(v: int) -> Color:
	match v:
		0: return ThemeTokens.color("card")
		2, 4: return ThemeTokens.color("accent_soft")
		8, 16, 32: return ThemeTokens.color("accent")
		64, 128: return ThemeTokens.color("alt")
		_: return ThemeTokens.color("gold")

func _tile_ink(v: int) -> Color:
	return ThemeTokens.color("play_ink") if v >= 8 else ThemeTokens.color("ink")

func _refresh_all() -> void:
	for i in GRID_SIZE:
		for j in GRID_SIZE:
			var idx := i * GRID_SIZE + j
			var v := int(grid[i][j])
			_cells[idx].color = _tile_color(v)
			_tile_labels[idx].text = str(v) if v > 0 else ""
			_tile_labels[idx].add_theme_color_override("font_color", _tile_ink(v))
	_score_lbl.text = "SCORE %d" % score
	_best_lbl.text = "BEST %d" % best_score

func _show_over() -> void:
	_over_score_lbl.text = "本局得分 %d · 历史最高 %d" % [score, best_score]
	_overlay.visible = true

## 存档（tetra 同构：ConfigFile [nova] section；best_score 强写立即落盘）
func _load_save() -> void:
	var cfg := ConfigFile.new()
	if cfg.load(SAVE_PATH) == OK:
		best_score = int(cfg.get_value("nova", "best", 0))

func _save_best() -> void:
	if score > best_score:
		best_score = score
	var cfg := ConfigFile.new()
	cfg.set_value("nova", "best", best_score)
	cfg.set_value("nova", "last_played", Time.get_datetime_string_from_system())
	cfg.save(SAVE_PATH)

func _max_tile() -> int:
	var m := 0
	for i in GRID_SIZE:
		for j in GRID_SIZE:
			m = maxi(m, int(grid[i][j]))
	return m

## 纯函数：构造空网格（4×4 全零）
static func _empty_grid() -> Array:
	var g: Array = []
	for i in GRID_SIZE:
		var row: Array = []
		for _j in GRID_SIZE:
			row.append(0)
		g.append(row)
	return g
