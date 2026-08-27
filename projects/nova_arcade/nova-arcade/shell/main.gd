extends Node
## 主入口（design §4）：BgLayer / App(PageStack+TabBar+OverlayLayer+ToastLayer) / GameHost
## _ready：Autoload 先于场景就绪，PageStack 挂载后重定位 Nav → 恢复主题 → 进初始 Tab

func _ready() -> void:
	# A-1（修复）：游戏运行时若被强杀/崩溃，屏幕方向可能残留为横屏；主入口强制复位竖屏，避免重启后 Shell 横屏错位。
	_restore_shell_viewport()
	Nav.locate_page_stack()
	## CR-4 T10：Registry 已加载后构建搜索索引，确保 SearchPage 进入时可立即查询
	var searcher_node := get_node_or_null("Services/Searcher")
	if searcher_node != null:
		searcher_node.call("build_index")
	ThemeTokens.theme_target = $App as Control  # Fix-1：4.5 root Window theme 不传播给子 Control，落到 App 容器
	ThemeTokens.restore(get_tree().root)
	Nav.switch_tab(0)

# 仅复位屏幕方向（安全、确定生效）。
# content_scale 视口切换语义待 Services 实现时复核 Godot 4.5（review A-6：design §4.6 用 content_scale_size 模拟分辨率切换，疑似误用 API）。
func _restore_shell_viewport() -> void:
	if DisplayServer.screen_get_orientation() != DisplayServer.SCREEN_PORTRAIT:
		DisplayServer.screen_set_orientation(DisplayServer.SCREEN_PORTRAIT)
