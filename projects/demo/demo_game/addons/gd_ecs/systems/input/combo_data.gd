class_name ComboData extends Resource
## 连招定义
##
## 定义一个连招的输入序列和触发条件。
## 每个连招由一组 ComboInput 序列组成，按顺序匹配输入缓冲。

## 连招唯一标识
@export var id: StringName = &""

## 连招名称
@export var name: String = ""

## 输入序列（按顺序匹配，每个元素为 ComboInput 资源）
@export var sequence: Array[ComboInput] = []

## 触发的技能 ID
@export var result_skill: StringName = &""

## 可取消的动画进度区间（0~1，在此区间内可被此连招取消）
@export var cancel_window: Vector2 = Vector2(0.2, 0.8)

## 需要处于的状态（&"" = 不限状态）
@export var required_state: StringName = &""

## 优先级（数值大的优先匹配）
@export var priority: int = 0
