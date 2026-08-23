extends SceneTree
## Arcade runtime PoC launcher:
##   godot --headless --path . -s res://poc/arcade_poc.gd

func _initialize() -> void:
    var driver = load("res://poc/arcade_runtime_driver.gd").new()
    root.add_child(driver)
