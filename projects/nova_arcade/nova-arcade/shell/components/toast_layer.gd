extends Control
## 顶部 Toast 层（design §4）：300ms 滑入 / 2200ms 停留 / 滑出
## 子节点约定（main.tscn）：Label（居中，初始隐藏）
## 时间线：set_parallel(true) 并行滑入（position+alpha）→ set_parallel(false) + interval 停留 → 顺序滑出。
## 经验：本构建无 tween_parallel()/嵌套 create_tween()；tween_interval 之后勿重复 set_parallel(true)（会压平停留）。

var _label: Label
var _tween: Tween

func _ready() -> void:
	_label = $Label as Label
	_label.visible = false

## 显示一条消息（公开入口；页面/服务调用）。命名偏差：show() 与 CanvasItem.show() 冲突 → show_msg()
func show_msg(msg: String) -> void:
	_label.text = msg
	_label.visible = true
	if _tween != null and _tween.is_valid():
		_tween.kill()
	_tween = create_tween()
	# 滑入：position + alpha 并行 300ms（ease-out）
	_tween.set_parallel(true)
	_tween.tween_property(_label, "position:y", 24.0, 0.3).from(-80.0) \
		.set_trans(Tween.TRANS_CUBIC).set_ease(Tween.EASE_OUT)
	_tween.tween_property(_label, "modulate:a", 1.0, 0.3).from(0.0)
	# 停留 2200ms（顺序）
	_tween.set_parallel(false)
	_tween.tween_interval(2.2)
	# 滑出：position + alpha 串行 250ms（ease-in）
	_tween.tween_property(_label, "position:y", -80.0, 0.25) \
		.set_trans(Tween.TRANS_CUBIC).set_ease(Tween.EASE_IN)
	_tween.tween_property(_label, "modulate:a", 0.0, 0.25)
