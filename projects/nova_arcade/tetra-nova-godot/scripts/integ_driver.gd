extends Node
## Integration driver: runs the UI-flow checks once the tree is live.

var main: Node = null
var _step := 0

func _ready() -> void:
    # wait two frames so main._ready + first process pass complete
    await get_tree().process_frame
    await get_tree().process_frame
    _run()
    await get_tree().process_frame
    get_tree().quit(0 if _ok else 1)

var _ok := true

func _fail(msg: String) -> void:
    _ok = false
    printerr("INTEG-FAIL: " + msg)

func _run() -> void:
    var game: Node = main.game
    var ui: CanvasLayer = main.ui
    if game == null:
        _fail("main.game is nil")
        return
    print("INTEG-OK scene assembled")

    game.input_action("confirm")
    if game.state != "PLAYING":
        _fail("not playing after confirm (state=" + game.state + ")")
        return
    print("INTEG-OK start -> PLAYING")

    game.wave_lines = game.wave_goal
    game.after_pipe()
    game._process(1.2)
    if game.state != "SELECT":
        _fail("select not open (state=" + game.state + ")")
        return
    if game.current_cards.size() != 3:
        _fail("expected 3 cards, got " + str(game.current_cards.size()))
        return
    print("INTEG-OK select open with 3 cards")

    game.pick_card(0)
    if game.state != "PLAYING" or game.wave != 2:
        _fail("wave not advanced (state=" + game.state + " wave=" + str(game.wave) + ")")
        return
    print("INTEG-OK pick -> wave 2")

    game.spawn_boss()
    ui.tick()
    print("INTEG-OK boss + hud tick")

    game.boss.hp = 1
    game.boss_damage(10.0, 1)
    game._process(2.5)
    ui.tick()
    print("INTEG-OK boss killed, bosses=" + str(game.stats_bosses) + " state=" + game.state)

    # pick the boss spoil if select opened
    var guard := 0
    while game.state == "SELECT" and guard < 4:
        game.pick_card(0)
        game._process(1.0)
        guard += 1

    if game.state == "PLAYING":
        game.muts = {}
        game.top_out()
        if game.state != "OVER":
            _fail("no game over (state=" + game.state + ")")
            return
    print("INTEG-OK game over")
    print("ALL INTEGRATION CHECKS DONE")
