extends Control
## Touch input: swipe to move / soft-drop, flick down = hard drop, tap = rotate
## (right 2/3 CW, left 1/3 CCW), long-press = hold. Plus optional button pad.

signal action(a: String)

var board_view: Node2D = null
var show_buttons := true

var _active := -1
var _start_pos := Vector2.ZERO
var _last_pos := Vector2.ZERO
var _start_t := 0.0
var _accum_x := 0.0
var _accum_y := 0.0
var _vel_y := 0.0
var _moved := false
var _soft := false
var _long_pressed := false
var _hold_timer := 0.0

func _ready() -> void:
    set_anchors_preset(Control.PRESET_FULL_RECT)
    mouse_filter = Control.MOUSE_FILTER_STOP
    _build_buttons()

func _build_buttons() -> void:
    var pad := HBoxContainer.new()
    pad.name = "Pad"
    pad.set_anchors_preset(Control.PRESET_BOTTOM_WIDE)
    pad.offset_left = 12
    pad.offset_right = -12
    pad.offset_bottom = -14
    pad.offset_top = -78
    pad.alignment = BoxContainer.ALIGNMENT_CENTER
    pad.add_theme_constant_override("separation", 10)
    add_child(pad)
    var defs := [
        ["◀", "left", "left_up"], ["▼", "down", "down_up"], ["▶", "right", "right_up"],
        ["⟳", "rot_cw", ""], ["⤓", "drop", ""], ["⇄", "hold", ""],
    ]
    for d in defs:
        var b := Button.new()
        b.text = d[0]
        b.add_theme_font_size_override("font_size", 26)
        b.custom_minimum_size = Vector2(64, 64)
        var dn: String = d[1]
        var up: String = d[2]
        b.button_down.connect(func(): action.emit(dn))
        if up != "":
            b.button_up.connect(func(): action.emit(up))
        pad.add_child(b)

func toggle_buttons() -> void:
    show_buttons = not show_buttons
    var pad := get_node_or_null("Pad")
    if pad:
        pad.visible = show_buttons

func _input(ev: InputEvent) -> void:
    if ev is InputEventScreenTouch:
        if ev.pressed:
            if _active == -1:
                _active = ev.index
                _start_pos = ev.position
                _last_pos = ev.position
                _start_t = Time.get_ticks_msec() / 1000.0
                _accum_x = 0.0
                _accum_y = 0.0
                _vel_y = 0.0
                _moved = false
                _soft = false
                _long_pressed = false
                _hold_timer = 0.0
        else:
            if ev.index == _active:
                _release()
                _active = -1
    elif ev is InputEventScreenDrag and ev.index == _active:
        var d: Vector2 = ev.position - _last_pos
        _last_pos = ev.position
        _accum_x += d.x
        _accum_y += d.y
        _vel_y = ev.velocity.y
        if absf(ev.position.x - _start_pos.x) > 18.0 or absf(ev.position.y - _start_pos.y) > 18.0:
            _moved = true
        var cell := 34.0
        if board_view:
            cell = board_view.cell_size * 0.9
        while _accum_x > cell:
            _accum_x -= cell
            action.emit("right")
        while _accum_x < -cell:
            _accum_x += cell
            action.emit("left")
        if _accum_y > 46.0 and not _soft:
            _soft = true
            action.emit("down")

func _process(delta: float) -> void:
    if _active != -1 and not _moved:
        _hold_timer += delta
        if _hold_timer > 0.45 and not _long_pressed:
            _long_pressed = true
            action.emit("hold")

func _release() -> void:
    if _soft:
        action.emit("down_up")
        _soft = false
    var dur := Time.get_ticks_msec() / 1000.0 - _start_t
    if not _moved and not _long_pressed and dur < 0.30:
        # tap: left sixth = CCW, else CW
        var vp := get_viewport_rect().size
        if _start_pos.x < vp.x * 0.166:
            action.emit("rot_ccw")
        else:
            action.emit("rot_cw")
    elif _moved and _vel_y > 1300.0 and _accum_y > 60.0:
        action.emit("drop")
    _vel_y = 0.0
