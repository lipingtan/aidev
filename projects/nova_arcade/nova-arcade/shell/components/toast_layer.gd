extends Control
## 顶部 Toast 层（design §4）：300ms 滑入 / 2200ms 停留 / 滑出
## 子节点约定（main.tscn）：Label（居中，初始隐藏）
## P-1修复：show_msg(msg, color) 增加可选 color 参数（金色 Toast 用 ThemeTokens.color("gold")）
## P-2修复：消息队列防重叠（_busy 时入队，tween.finished → _pop_queue）

var _label: Label
var _tween: Tween
var _queue: Array[String] = []
var _busy: bool = false

func _ready() -> void:
	_label = $Label as Label
	_label.visible = false

## 显示一条消息（公开入口；页面/服务调用）
## color != null 时覆盖字体颜色（如金色 Toast）；null 时恢复默认
func show_msg(msg: String, color: Variant = null) -> void:
	if _busy:
		_queue.append(msg)
		return
	_label.text = msg
	_label.visible = true
	if color != null:
		_label.add_theme_color_override("font_color", color as Color)
	else:
		_label.remove_theme_color_override("font_color")
	if _tween != null and _tween.is_valid():
		_tween.kill()
	_busy = true
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
	_tween.connect("finished", _on_tween_finished.bind())

## tween 完成 → 出队显示下一条（P-2修复）
func _on_tween_finished() -> void:
	_busy = false
	if _queue.size() > 0:
		show_msg(_queue.pop_front())
