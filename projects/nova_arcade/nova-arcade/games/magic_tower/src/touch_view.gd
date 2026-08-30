extends Node
## 魔塔 TouchView（CR-6 T1）：盒子 720×1560 竖屏 UI 层
## - 顶部 HUD（HP/ATK/DEF/Gold + 消息 + 战斗行）
## - 底部左下十字盘（9 图切换）+ 右侧 A/B/C/D 玩法按钮（Start/投币/退出 已移除，盒内无投币语义）
## - 触屏：十字盘连走（写 main.tdir）+ 按钮 tap（A=攻击 in battle）
## CanvasLayer 渲染于视口坐标，HUD/控件固定不随棋盘偏移

const ASSET_DIR := "res://games/magic_tower/assets/"
const DPAD_CENTER := Vector2(140, 1360)
const DPAD_RADIUS := 84.0
const DEADZONE := 10.0
const BTN_DIR := "res://games/magic_tower/assets/buttons/"

var _main: Node = null
var _hud_layer: CanvasLayer = null
var _stats: Label = null
var _msg: Label = null
var _battle: Label = null
var _dpad: Sprite2D = null
var _dpad_tex: Dictionary = {}
var _btns: Array = []
var tdir := Vector2.ZERO
var _dpad_touch_id := -1
var _pressed_idx := -1
var _pressed_id := -1

func _ready() -> void:
	_build_hud()
	_build_dpad()
	_build_buttons()
	_build_start()

## 开始遮罩：标题 + 玩法说明 + 开始按钮（覆盖全屏，点击后隐藏；main.gd 的 _process 靠 paused 判定无需改）
func _build_start() -> void:
	var ov := Control.new()
	ov.set_anchors_preset(Control.PRESET_FULL_RECT)
	ov.mouse_filter = Control.MOUSE_FILTER_STOP
	var dim := ColorRect.new()
	dim.color = Color(0.01, 0.016, 0.05, 0.88)
	dim.set_anchors_preset(Control.PRESET_FULL_RECT)
	ov.add_child(dim)
	var box := VBoxContainer.new()
	box.alignment = BoxContainer.ALIGNMENT_CENTER
	box.set_anchors_preset(Control.PRESET_FULL_RECT)
	box.add_theme_constant_override("separation", 20)
	var title := Label.new()
	title.text = "魔 塔"
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	title.add_theme_font_size_override("font_size", 56)
	title.add_theme_color_override("font_color", Color(1, 0.824, 0.247))
	box.add_child(title)
	var d1 := Label.new()
	d1.text = "方向键/十字盘移动 · 撞怪物进入战斗"
	d1.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	d1.add_theme_font_size_override("font_size", 19)
	d1.add_theme_color_override("font_color", Color(0.812, 0.918, 1))
	box.add_child(d1)
	var d2 := Label.new()
	d2.text = "战斗中按空格/A/方向键攻击 · 拾取道具变强"
	d2.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	d2.add_theme_font_size_override("font_size", 17)
	d2.add_theme_color_override("font_color", Color(0.6, 0.72, 0.85))
	box.add_child(d2)
	var btn := Button.new()
	btn.text = "▶ 开始游戏"
	btn.focus_mode = Control.FOCUS_NONE
	btn.custom_minimum_size = Vector2(260, 64)
	btn.add_theme_font_size_override("font_size", 24)
	btn.pressed.connect(func() -> void:
		ov.visible = false
		ov.mouse_filter = Control.MOUSE_FILTER_IGNORE
		if _main != null:
			_main.set_paused(false))
	box.add_child(btn)
	box.add_child(Control.new())
	ov.add_child(box)
	_hud_layer.add_child(ov)

## 由 main._ready 注入（子节点 _ready 先于 main）
func attach(main: Node) -> void:
	_main = main
	main.paused = true   # 开始遮罩期间锁定游戏，点「开始游戏」解锁

# ---- HUD ----
func _build_hud() -> void:
	_hud_layer = CanvasLayer.new()
	add_child(_hud_layer)
	_stats = _make_label(8, 10, 26)
	_msg = _make_label(8, 46, 22)
	_battle = _make_label(8, 78, 22)

func _make_label(x: int, y: int, size: int) -> Label:
	var l := Label.new()
	l.position = Vector2(x, y)
	l.add_theme_font_size_override("font_size", size)
	l.modulate = Color(1, 0.95, 0.8)
	_hud_layer.add_child(l)
	return l

func show_stats(text: String) -> void:
	if _stats != null:
		_stats.text = text

func show_msg(text: String) -> void:
	if _msg != null:
		_msg.text = text

func show_battle(text: String) -> void:
	if _battle != null:
		_battle.text = text

func battle_text() -> String:
	return _battle.text if _battle != null else ""

# ---- 十字盘（按方向切换贴图）----
func _build_dpad() -> void:
	var names := ["none", "up", "down", "left", "right",
			"up_left", "up_right", "down_left", "down_right"]
	for n in names:
		_dpad_tex[n] = load(ASSET_DIR + "dpad/dpad_%s.png" % n)
	_dpad = Sprite2D.new()
	_dpad.texture = _dpad_tex["none"]
	_dpad.centered = true
	_dpad.position = DPAD_CENTER
	var s := 150.0 / (_dpad_tex["none"] as Texture2D).get_size().x
	_dpad.scale = Vector2(s, s)
	_dpad.texture_filter = CanvasItem.TEXTURE_FILTER_LINEAR
	_hud_layer.add_child(_dpad)

func update_dpad(d: Vector2) -> void:
	var st := _dir_state(d)
	if _dpad != null and st != _cur_state:
		_cur_state = st
		_dpad.texture = _dpad_tex[st]

var _cur_state := "none"

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

# ---- A/B/C/D 玩法按钮（右侧，按下换 _press 图）----
func _build_buttons() -> void:
	_add_button("A", "button_a", Vector2(560, 1300), 64.0)
	_add_button("B", "button_b", Vector2(500, 1380), 64.0)
	_add_button("C", "button_c", Vector2(560, 1460), 64.0)
	_add_button("D", "button_d", Vector2(620, 1380), 64.0)

func _add_button(name: String, base: String, center: Vector2, target: float) -> void:
	var nrm = load(BTN_DIR + base + ".png")
	var prs = load(BTN_DIR + base + "_press.png")
	var spr = Sprite2D.new()
	spr.texture = nrm; spr.centered = true; spr.position = center
	var s = target / nrm.get_size().x
	spr.scale = Vector2(s, s)
	spr.texture_filter = CanvasItem.TEXTURE_FILTER_LINEAR
	_hud_layer.add_child(spr)
	_btns.append({"spr": spr, "nrm": nrm, "prs": prs, "name": name, "center": center, "half": target / 2.0})

func _btn_at(pos: Vector2) -> int:
	for i in _btns.size():
		if (pos - _btns[i].center).length() <= _btns[i].half * 1.3:
			return i
	return -1

# ---- 触屏输入（十字盘 + buttons）----
func _input(event: InputEvent) -> void:
	var st = event as InputEventScreenTouch
	if st != null:
		if st.pressed:
			var bi = _btn_at(st.position)
			if bi >= 0:
				_pressed_idx = bi; _pressed_id = st.index
				_btns[bi].spr.texture = _btns[bi].prs
				_on_button(_btns[bi].name)
			elif (st.position - DPAD_CENTER).length() <= DPAD_RADIUS:
				_dpad_touch_id = st.index
				tdir = _dir_of(st.position)
		else:
			if st.index == _pressed_id:
				if _pressed_idx >= 0:
					_btns[_pressed_idx].spr.texture = _btns[_pressed_idx].nrm
				_pressed_id = -1; _pressed_idx = -1
			elif st.index == _dpad_touch_id:
				_dpad_touch_id = -1; tdir = Vector2.ZERO
		return
	var sd = event as InputEventScreenDrag
	if sd != null and sd.index == _dpad_touch_id:
		tdir = _dir_of(sd.position)

func _on_button(name: String) -> void:
	if _main == null:
		return
	if name == "A":
		_main._attack()
	elif _stats != null:
		show_msg("按钮: " + name)

func _dir_of(pos: Vector2) -> Vector2:
	var off = pos - DPAD_CENTER
	return Vector2.ZERO if off.length() <= DEADZONE else off.normalized()
