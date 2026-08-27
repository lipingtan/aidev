class_name LauncherTransition extends Node
## 转场状态机 + 视口切换（CR-2 T4，自 launcher.gd 拆出以守 ≤200 行；无状态，由 Launcher add_child）
## - to_game(meta)：fade_out → **遮罩全不透明后**切游戏视口/方向
## - to_shell()：fade_out → 全不透明后恢复 Shell 默认视口 + portrait（override §2 铁律）
## - reveal()：fade_in 至完全透明
## - vp_at_full_mask：时序钩子（T8 断言最近一次切换发生在遮罩全不透明之后）
##
## 依赖：TransitionOverlay 场景（shell/components）、GameMeta

## Shell 默认设计分辨率（720×1560 竖屏；游戏未声明 viewport 时继承）
const SHELL_VIEWPORT := Vector2i(720, 1560)
## Shell 默认屏幕方向
const SHELL_ORIENTATION := "portrait"

var _overlay: Control = null
## 时序钩子（T8 断言）：最近一次视口切换是否发生在遮罩全不透明之后
var vp_at_full_mask := false

## 实例化遮罩并挂 root（Launcher._ready 调用；不改 main.tscn，保 RG-8）
func attach() -> void:
	_overlay = (load("res://shell/components/transition_overlay.tscn") as PackedScene).instantiate()
	get_tree().root.add_child(_overlay)

## 切游戏视口（必须在遮罩全不透明后调用，override §2 铁律）；记录时序钩子
func _switch_viewport(size: Vector2i, orientation: String) -> void:
	get_tree().root.content_scale_size = size
	DisplayServer.screen_set_orientation(_orientation_enum(orientation))
	vp_at_full_mask = (_overlay as Node).call("is_solid")

## 恢复 Shell 默认视口 + portrait（必须在遮罩全不透明后调用）
func _restore_viewport() -> void:
	get_tree().root.content_scale_size = SHELL_VIEWPORT
	DisplayServer.screen_set_orientation(DisplayServer.ScreenOrientation.SCREEN_PORTRAIT)

## 方向字符串 → DisplayServer 枚举（4.5 常量名：SCREEN_LANDSCAPE / SCREEN_PORTRAIT）
func _orientation_enum(o: String) -> DisplayServer.ScreenOrientation:
	if o == "landscape":
		return DisplayServer.ScreenOrientation.SCREEN_LANDSCAPE
	return DisplayServer.ScreenOrientation.SCREEN_PORTRAIT

## 转场到游戏：fade_out → 全不透明后切视口（meta 未声明时继承 Shell 默认）
func to_game(meta: GameMeta) -> void:
	await (_overlay as Node).call("fade_out")
	_switch_viewport(meta.get_viewport_size(), meta.get_orientation())

## 转场回 Shell：fade_out → 全不透明后恢复默认视口/方向
func to_shell() -> void:
	await (_overlay as Node).call("fade_out")
	_restore_viewport()

## 淡入恢复显示（启动完成 / 退出完成后调用）
func reveal() -> void:
	await (_overlay as Node).call("fade_in")
