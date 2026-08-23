extends SceneTree
## headless 冒烟启动脚本（T1 验收门槛，后续任务回归复用）
##
## 用法：Godot --headless --path <工程目录> -s res://tools/smoke_boot.gd
## 跑通 Autoload 初始化后 quit(0)；任何脚本错误都会使退出码非 0。

func _initialize() -> void:
	print("[smoke] autoload 加载完成：")
	for name in ["EventBus", "QualitySettings", "DB", "Registry", "Nav", "Sound"]:
		var node := root.get_node_or_null(NodePath(name))
		if node == null:
			push_error("[smoke] 缺少 Autoload: %s" % name)
			quit(1)
			return
		print("  - %s OK" % name)
	print("[smoke] boot OK")
	quit(0)
