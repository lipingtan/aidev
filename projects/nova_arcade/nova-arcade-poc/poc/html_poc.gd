extends SceneTree
## HTML runtime PoC launcher:
##   godot --headless --path . -s res://poc/html_poc.gd

func _initialize() -> void:
    var driver = load("res://poc/html_runtime_driver.gd").new()
    root.add_child(driver)
