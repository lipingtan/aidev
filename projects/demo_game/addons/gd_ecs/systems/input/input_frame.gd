class_name InputFrame extends RefCounted
## 单帧输入记录
##
## 存储一次输入事件的完整信息，用于输入缓冲和连招匹配。

## 动作名称
var action: StringName = &""

## 输入时间戳（游戏时间，秒）
var timestamp: float = 0.0

## 输入方向（摇杆/方向键）
var direction: Vector2 = Vector2.ZERO

## 是否为长按状态
var is_held: bool = false
