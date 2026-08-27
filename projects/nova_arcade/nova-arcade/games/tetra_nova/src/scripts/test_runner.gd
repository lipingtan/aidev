extends SceneTree
## Headless logic test runner:
##   godot --headless --path . -s res://games/tetra_nova/src/scripts/test_runner.gd

func _initialize() -> void:
    var game_script := load("res://games/tetra_nova/src/scripts/game.gd")
    var game: Node = game_script.new()
    var ok: bool = game.dev_test()
    game.queue_free()
    quit(0 if ok else 1)
