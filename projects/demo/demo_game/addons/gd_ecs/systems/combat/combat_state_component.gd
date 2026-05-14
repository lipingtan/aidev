class_name CombatStateComponent extends EcsComponent
## 战斗状态 Component
##
## 由状态机写入，由 DamageSystem 读取。
## 存储战斗相关的即时状态（无敌帧、格挡、连击等）。

## 当前状态名称（由状态机写入）
var current_state: StringName = &"idle"

## 当前状态持续时间（秒）
var state_time: float = 0.0

## 是否处于无敌帧（闪避/受击硬直期间）
var is_invincible: bool = false

## 无敌帧剩余时间
var invincible_timer: float = 0.0

## 是否正在格挡
var is_blocking: bool = false

## 格挡有效角度（度，相对于面朝方向）
@export var block_angle: float = 120.0

## 格挡减伤比例（0~1）
@export var block_reduction: float = 0.7

## 当前连击计数
var combo_count: int = 0

## 连击重置计时器
var combo_reset_timer: float = 0.0

## 连击重置时间（秒）
@export var combo_reset_time: float = 2.0


func get_component_name() -> StringName:
	return &"CombatState"


## 开始无敌帧
func start_invincible(duration: float) -> void:
	is_invincible = true
	invincible_timer = duration


## 更新无敌帧计时（由 CombatStateSystem 每帧调用）
func tick_invincible(delta: float) -> void:
	if is_invincible:
		invincible_timer -= delta
		if invincible_timer <= 0.0:
			is_invincible = false
			invincible_timer = 0.0


## 增加连击数
func add_combo() -> int:
	combo_count += 1
	combo_reset_timer = combo_reset_time
	return combo_count


## 重置连击
func reset_combo() -> void:
	combo_count = 0
	combo_reset_timer = 0.0
