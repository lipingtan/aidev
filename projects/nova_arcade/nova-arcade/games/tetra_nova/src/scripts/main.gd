extends Node2D
## Main assembly: game core + renderers + audio + ui + input + neon environment.

const GameClass := preload("res://games/tetra_nova/src/scripts/game.gd")

## Save file path (GameModule adapter injects ctx.save_dir + "save.cfg" at boot; standalone default below)
var SAVE_PATH := "user://save.cfg"

var game: Node
var board_view: Node2D
var fx: Node2D
var bg: Node2D
var au: Node
var ui: CanvasLayer
var touch: Control
var shaker: Node2D

func _ready() -> void:
    randomize()
    # NOTE: no WorldEnvironment glow — on the mobile renderer, 2D HDR glow
    # postfx mangles solid fills (verified by rect repro). Neon look is done
    # with layered translucent draws in board_view instead.
    shaker = Node2D.new()
    add_child(shaker)

    bg = preload("res://games/tetra_nova/src/scripts/star_bg.gd").new()
    add_child(bg)

    board_view = preload("res://games/tetra_nova/src/scripts/board_view.gd").new()
    shaker.add_child(board_view)

    fx = preload("res://games/tetra_nova/src/scripts/fx_layer.gd").new()
    shaker.add_child(fx)
    fx.board_view = board_view

    au = preload("res://games/tetra_nova/src/scripts/audio_manager.gd").new()
    add_child(au)

    game = GameClass.new()
    add_child(game)
    game.fx = fx
    game.au = au
    game.board_view = board_view
    board_view.game = game   # board renderer needs the game state to draw!

    ui = preload("res://games/tetra_nova/src/scripts/ui.gd").new()
    add_child(ui)
    ui.game = game
    ui.action.connect(_on_ui_action)
    game.state_changed.connect(ui.on_state_changed)
    game.select_opened.connect(ui.on_select_opened)
    game.run_over.connect(ui.on_run_over)
    game.run_over.connect(_save_best)

    touch = preload("res://games/tetra_nova/src/scripts/touch_controls.gd").new()
    add_child(touch)
    touch.board_view = board_view
    touch.action.connect(_on_ui_action)

    au.ensure()  # prebake all PCM during load, not on first input

    get_viewport().size_changed.connect(_layout)
    _layout()

func _layout() -> void:
    var v := get_viewport_rect().size
    board_view.layout(v)
    bg.layout(v)

# ===================================================================== input routing
func _on_ui_action(a: String) -> void:
    au.ensure()
    if a == "mute":
        au.muted = not au.muted
        return
    if a == "toggle_buttons":
        touch.toggle_buttons()
        return
    game.input_action(a)

func _unhandled_key_input(ev: InputEvent) -> void:
    if not (ev is InputEventKey) or not ev.pressed or ev.echo:
        return
    var key: Key = ev.keycode
    var a := ""
    match key:
        KEY_LEFT: a = "left"
        KEY_RIGHT: a = "right"
        KEY_DOWN: a = "down"
        KEY_UP, KEY_X: a = "rot_cw"
        KEY_Z: a = "rot_ccw"
        KEY_SPACE, KEY_ENTER: a = "drop"
        KEY_C, KEY_SHIFT: a = "hold"
        KEY_P, KEY_ESCAPE: a = "pause"
        KEY_R:
            if game.state == "SELECT":
                a = "reroll"
            else:
                a = "rerun"
        KEY_M: a = "mute"
        KEY_B: a = "toggle_buttons"
        KEY_1: a = "card1"
        KEY_2: a = "card2"
        KEY_3: a = "card3"
    if a != "":
        _on_ui_action(a)

func _unhandled_input(ev: InputEvent) -> void:
    if ev is InputEventKey and not ev.pressed:
        match ev.keycode:
            KEY_LEFT: game.input_action("left_up")
            KEY_RIGHT: game.input_action("right_up")
            KEY_DOWN: game.input_action("down_up")

# ===================================================================== frame
func _process(delta: float) -> void:
    var frozen: bool = fx.hitstop > 0.0 and game.state == "PLAYING"
    fx.tick(delta, frozen)
    bg.tick(delta)
    shaker.position = fx.shake_offset()
    _music_drive()
    ui.tick()
    touch.visible = game.state == "PLAYING" or game.state == "PAUSED"

func _music_drive() -> void:
    match game.state:
        "PLAYING":
            if not au.is_processing():
                au.start_music()
        "MENU":
            au.stop_music()
        _:
            pass

func _save_best(stats: Dictionary) -> void:
    var cfg := ConfigFile.new()
    cfg.load(SAVE_PATH)
    var best: int = cfg.get_value("nova", "best", 0)
    var wv: int = cfg.get_value("nova", "wave", 1)
    var runs: int = cfg.get_value("nova", "runs", 0)
    cfg.set_value("nova", "best", maxi(best, stats.score))
    cfg.set_value("nova", "wave", maxi(wv, stats.wave))
    cfg.set_value("nova", "runs", runs + 1)
    cfg.save(SAVE_PATH)
