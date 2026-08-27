extends SceneTree

func _initialize() -> void:
    for path in ["fx_layer", "board_view", "audio_manager", "ui", "mini_piece", "touch_controls", "star_bg", "main"]:
        var s = load("res://games/tetra_nova/src/scripts/" + path + ".gd")
        if s == null:
            print("LOAD-FAIL " + path)
        else:
            print("LOAD-OK " + path)
    quit(0)
