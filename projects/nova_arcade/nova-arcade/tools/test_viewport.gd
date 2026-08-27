extends Node
## A-6 挂账验证：content_scale_size 运行时切换是否生效（headless）
## 判据：canvas_items + expand 下，切换后 root.get_visible_rect() 应随 content_scale_size 变化

func _ready() -> void:
	var root := get_tree().root
	print("window_size=", root.size)
	print("before: css=", root.content_scale_size, " visible=", root.get_visible_rect())
	root.content_scale_size = Vector2i(1920, 1080)
	await get_tree().process_frame
	await get_tree().process_frame
	print("after:  css=", root.content_scale_size, " visible=", root.get_visible_rect())
	# 恢复 Shell 基准，验证双向可切
	root.content_scale_size = Vector2i(720, 1560)
	await get_tree().process_frame
	await get_tree().process_frame
	print("restored: css=", root.content_scale_size, " visible=", root.get_visible_rect())
	get_tree().quit(0)
