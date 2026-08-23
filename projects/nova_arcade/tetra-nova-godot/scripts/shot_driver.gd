extends Node
## Screenshot capture driver: boots the real scene, plays a few seconds,
## saves PNGs so we can see exactly what a player sees.

var main: Node = null
var shots := 0

func _ready() -> void:
    await get_tree().process_frame
    await get_tree().process_frame
    print("[shot] scene up, starting game")
    main.game.input_action("confirm")
    await _after(30, "01_just_started")     # piece near top
    await _after(120, "02_piece_falling")   # ~2s later, mid-board
    for k in 3:
        main.game.input_action("drop")
        await get_tree().process_frame
        await get_tree().process_frame
    await _after(30, "03_after_drops")      # stack on floor
    print("[shot] piece=", main.game.piece, " stack_height=", _stack_h())
    print("[shot] cell=", main.board_view.cell_size, " origin=", main.board_view.board_origin,
        " board=", main.board_view.board_size, " vp=", get_viewport().get_visible_rect().size)
    print("SHOTS DONE")
    get_tree().quit(0)

func _after(frames: int, tag: String) -> void:
    for i in frames:
        await get_tree().process_frame
    _shot(tag)

func _shot(tag: String) -> void:
    shots += 1
    var img := get_viewport().get_texture().get_image()
    var path := "user://shot_" + tag + ".png"
    img.save_png(path)
    print("[shot] saved ", path, " ", img.get_size())

func _stack_h() -> int:
    var h := 0
    for y in main.game.ROWS:
        for x in main.game.W:
            if main.game.grid[y][x] != null:
                h = main.game.ROWS - y
                return h
    return h
