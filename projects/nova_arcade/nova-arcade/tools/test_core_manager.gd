extends SceneTree
## CoreManager 桩临时验证脚本（Task 9 验收，RG-6）
##
## 用法：Godot --headless --path <工程目录> -s res://tools/test_core_manager.gd
## 实例化 CoreManager 调用四接口，确认无报错且 Mock 恒返回未安装。

func _initialize() -> void:
	var cm := CoreManager.new()
	var installed: bool = cm.is_installed("1.0.0")
	var checked: bool = cm.check_installed("1.0.0")
	cm.download("1.0.0", "abc123def456")
	cm.remove("1.0.0")
	print("[test] is_installed=%s check_installed=%s" % [installed, checked])
	if installed or checked:
		push_error("[test] Mock 应恒返回未安装")
		quit(1)
		return
	print("[test] CoreManager 四接口调用无报错，Mock 行为正确")
	quit(0)
