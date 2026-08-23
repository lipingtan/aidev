extends SceneTree
## Windowed screenshot run:
##   godot --path . -s res://scripts/shot.gd
## (NOT headless — needs a real renderer to capture pixels)

func _initialize() -> void:
    var main = load("res://scenes/Main.tscn").instantiate()
    root.add_child(main)
    var driver = load("res://scripts/shot_driver.gd").new()
    driver.main = main
    root.add_child(driver)
