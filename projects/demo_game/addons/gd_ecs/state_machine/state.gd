class_name State extends Node
## 状态基类
##
## 状态机中的单个状态。通过读写父 Entity 的 Component 来驱动行为。
## 子类覆盖 enter/exit/update/physics_update 实现具体逻辑。
##
## 依赖：
## - StateMachine: 父节点，管理状态切换
## - EcsEntity: 祖先节点，提供 Component 访问

## 所属的 Entity（由 StateMachine 在 _ready 时设置，可以是任何 Entity 类型）
var entity: Node = null

## 所属的状态机（由 StateMachine 在 _ready 时设置）
var state_machine: StateMachine = null


## 进入状态时调用
func enter() -> void:
	pass


## 退出状态时调用
func exit() -> void:
	pass


## 每帧更新（对应 _process）
func update(delta: float) -> void:
	pass


## 物理帧更新（对应 _physics_process）
func physics_update(delta: float) -> void:
	pass


## 处理输入事件
func handle_input(event: InputEvent) -> void:
	pass


## 获取 Entity 的 Component（便捷方法）
func get_comp(component_name: StringName) -> EcsComponent:
	if entity and entity.has_method("get_component"):
		return entity.get_component(component_name)
	return null
