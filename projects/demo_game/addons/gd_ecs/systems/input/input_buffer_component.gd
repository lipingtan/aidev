class_name InputBufferComponent extends EcsComponent
## 输入缓冲 Component
##
## 记录最近 N 帧的输入，用于连招匹配。
## 支持输入宽容窗口（在窗口内的输入都有效）。

## 单帧输入记录
class InputFrame:
	var action: StringName = &""
	var timestamp: float = 0.0
	var direction: Vector2 = Vector2.ZERO
	var is_held: bool = false

## 缓冲有效时间窗口（秒）
@export var buffer_window: float = 0.3

## 最大缓冲帧数
@export var max_frames: int = 10

## 输入缓冲队列
var buffer: Array[InputFrame] = []

## 当前帧时间戳
var current_time: float = 0.0


func get_component_name() -> StringName:
	return &"InputBuffer"


## 记录一次输入
func push_input(action: StringName, direction: Vector2 = Vector2.ZERO, is_held: bool = false) -> void:
	var frame := InputFrame.new()
	frame.action = action
	frame.timestamp = current_time
	frame.direction = direction
	frame.is_held = is_held
	buffer.append(frame)
	# 超出最大帧数时移除最旧的
	if buffer.size() > max_frames:
		buffer.pop_front()


## 清理过期输入
func clean_expired() -> void:
	var cutoff: float = current_time - buffer_window
	while buffer.size() > 0 and buffer[0].timestamp < cutoff:
		buffer.pop_front()


## 获取最近的指定动作输入（在窗口内）
func get_recent_input(action: StringName) -> InputFrame:
	var cutoff: float = current_time - buffer_window
	for i in range(buffer.size() - 1, -1, -1):
		var frame: InputFrame = buffer[i]
		if frame.timestamp >= cutoff and frame.action == action:
			return frame
	return null


## 清空缓冲
func clear() -> void:
	buffer.clear()
