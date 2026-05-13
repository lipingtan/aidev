class_name ComboInput extends Resource
## 单个连招输入要求
##
## 定义连招序列中一个输入步骤的匹配条件。

## 动作名称（如 &"attack"、&"heavy_attack"）
@export var action: StringName = &""

## 与上一个输入的最大间隔（秒）
@export var max_interval: float = 0.5

## 可选的方向要求（&"" = 不限方向）
@export var direction_hint: StringName = &""
