extends Node
## 视觉验证截图：加载主场景，等渲染完成，截图到指定路径
@export var screenshot_path: String = "/c/data/run/hermes/tmp_screenshot_home.png"

func _ready() -> void:
    var res := load("res://shell/main.tscn") as PackedScene
    if res == null:
        print("[screenshot] FAIL: main scene not found")
        get_tree().quit(1)
        return
    add_child(res.instantiate())
    await get_tree().create_timer(0.5).timeout
    var img := get_viewport().get_texture().get_image()
    var err := img.save_png(screenshot_path)
    if err == OK:
        print("[screenshot] PASS: saved to %s (%dx%d)" % [screenshot_path, img.get_width(), img.get_height()])
    else:
        print("[screenshot] FAIL: error=%d" % err)
    get_tree().quit(0)
