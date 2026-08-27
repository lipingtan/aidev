extends SceneTree
## Boots the real Main scene with an integration driver attached.
##   godot --headless --path . -s res://games/tetra_nova/src/scripts/probe4.gd

func _initialize() -> void:
    var main = load("res://games/tetra_nova/src/scenes/Main.tscn").instantiate()
    root.add_child(main)
    var driver = load("res://games/tetra_nova/src/scripts/integ_driver.gd").new()
    driver.main = main
    root.add_child(driver)
