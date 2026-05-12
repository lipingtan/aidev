class_name AiStateComponent extends EcsComponent
## AI 状态 Component
##
## 存储 AI 的当前模式、目标和行为参数。
## 由 AiDecisionSystem 写入，由状态机读取执行。

## AI 模式枚举
enum AiMode {
	IDLE,     ## 待机
	PATROL,   ## 巡逻
	ALERT,    ## 警戒（发现玩家但未进入战斗）
	COMBAT,   ## 战斗
	RETREAT,  ## 撤退（低血量）
	DEAD,     ## 死亡
}

## 当前 AI 模式
var ai_mode: AiMode = AiMode.IDLE

## 当前目标 Entity ID（-1 = 无目标）
var target_entity_id: int = -1

## 警戒范围（进入此范围触发 ALERT）
@export var alert_range: float = 10.0

## 追击范围（超出此范围放弃追击）
@export var chase_range: float = 15.0

## 攻击范围（进入此范围触发攻击）
@export var attack_range: float = 2.0

## 撤退血量阈值（HP 低于此比例时撤退）
@export var retreat_hp_threshold: float = 0.2

## 当前帧的决策结果（由 AiDecisionSystem 写入）
var current_decision: StringName = &""

## 上次看到目标的位置
var last_known_target_pos: Vector3 = Vector3.ZERO

## 警戒计时器（超时后回到 IDLE）
var alert_timer: float = 0.0

## 警戒超时时间
@export var alert_timeout: float = 5.0


func get_component_name() -> StringName:
	return &"AiState"
