class_name ComboData extends Resource
## 连招定义
##
## 定义一个连招的输入序列和触发条件。

## 单个连招输入要求
class ComboInput:
	## 动作名称（如 &"attack"、&"heavy_attack"）
	var action: StringName = &""
	## 与上一个输入的最大间隔（秒）
	var max_interval: float = 0.5
	## 可选的方向要求（&"" = 不限方向）
	var direction_hint: StringName = &""

## 连招唯一标识
@export var id: StringName = &""

## 连招名称
@export var name: String = ""

## 输入序列（按顺序匹配）
@export var sequence: Array[ComboInput] = []

## 触发的技能 ID
@export var result_skill: StringName = &""

## 可取消的动画进度区间（0~1，在此区间内可被此连招取消）
@export var cancel_window: Vector2 = Vector2(0.2, 0.8)

## 需要处于的状态（&"" = 不限状态）
@export var required_state: StringName = &""

## 优先级（数值大的优先匹配）
@export var priority: int = 0
