extends Node
## 截图验证：加载主场景，等渲染完成，截图并退出
@export var screenshot_path: String = "/c/data/run/hermes/tmp_gui_verify.png"

func _ready() -> void:
    # 导入 autoloads 后加载主场景
    var res := load("res://shell/main.tscn") as PackedScene
    if res == null:
        print("[verify] FAIL: main scene not found")
        get_tree().quit(1)
        return
    add_child(res.instantiate())
    # 等渲染完成（0.3s 足够 GPU 渲染一帧）
    await get_tree().create_timer(0.3).timeout
    var img := get_viewport().get_texture().get_image()
    var err := img.save_png(screenshot_path)
    print("[verify] saved %s (%dx%d) err=%d" % [screenshot_path, img.get_width(), img.get_height(), err])
    get_tree().quit(0)
