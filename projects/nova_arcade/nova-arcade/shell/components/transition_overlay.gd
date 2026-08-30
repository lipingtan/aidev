extends Control
## 全屏转场遮罩（Launcher 专用，design §5）：300ms 淡出/淡入
## - fade_out 结束 = 遮罩全不透明（mask_solid），此后才允许切视口（override §2 铁律）
## - fade_in 结束 = 完全透明（fade_done），恢复正常显示与输入
## - 遮罩期间鼠标拦截（MOUSE_FILTER_STOP）；键盘随冻结模块的 process_mode 一并停摆
##
## 依赖：无（纯 UI；由 Launcher._ready 运行时实例化挂 root，不改 main.tscn）

## 淡入/淡出时长（秒）
const FADE_S := 0.3

signal mask_solid ## 遮罩已达全不透明（可安全切视口/换场景）
signal fade_done ## 淡入结束（恢复显示）

@onready var _mask: ColorRect = $Mask
var _tween: Tween = null

func _ready() -> void:
	visible = false
	mouse_filter = Control.MOUSE_FILTER_IGNORE

## 是否全不透明（Launcher 切视口前断言，T8 时序钩子）
func is_solid() -> bool:
	return _mask != null and _mask.color.a >= 1.0

## 淡出至全不透明；结束后发 mask_solid 并返回
func fade_out() -> void:
	_kill_tween()
	if _mask == null:
		push_error("TransitionOverlay: Mask 节点缺失")
		return
	visible = true
	mouse_filter = Control.MOUSE_FILTER_STOP
	_mask.color.a = 0.0
	var tw := create_tween()
	_tween = tw
	tw.tween_property(_mask, "color:a", 1.0, FADE_S)
	await tw.finished
	mask_solid.emit()

## 淡入至完全透明；结束后发 fade_done 并返回
func fade_in() -> void:
	_kill_tween()
	if _mask == null:
		push_error("TransitionOverlay: Mask 节点缺失")
		return
	var tw := create_tween()
	_tween = tw
	tw.tween_property(_mask, "color:a", 0.0, FADE_S).from(1.0)
	await tw.finished
	# Mask 本层须一并放行输入：alpha=0 但 mouse_filter=STOP 时是「隐形玻璃」，
	# 盖在 root 最顶会吞掉结算卡/首页全部点击（GUI 截图+hover 实测定位）
	_mask.mouse_filter = Control.MOUSE_FILTER_IGNORE
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	fade_done.emit()

## 终止进行中的淡入淡出
func _kill_tween() -> void:
	if _tween != null and _tween.is_valid():
		_tween.kill()
