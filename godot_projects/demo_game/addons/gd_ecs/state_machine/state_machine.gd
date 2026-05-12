class_name StateMachine extends Node
## 通用状态机
##
## 管理状态切换，与 ECS Entity 集成。
## 状态通过读写 Entity 的 Component 来驱动行为。
## 支持动态添加/移除状态（DLC 扩展用）。
##
## 依赖：
## - EcsEntity: 祖先节点

## 当状态切换时触发
signal state_changed(old_state: StringName, new_state: StringName)

## 初始状态名称（对应子节点名称）
@export var initial_state: StringName = &""

## 当前状态
var current_state: State = null

## 当前状态名称
var current_state_name: StringName = &""

## 所属 Entity（支持任何 Entity 类型）
var _entity: Node = null

## 状态字典 { state_name: State }
var _states: Dictionary = {}


func _ready() -> void:
	# 查找祖先 EcsEntity
	_entity = _find_entity()
	if not _entity:
		push_warning("StateMachine 未找到祖先 EcsEntity 节点")
	
	# 注册所有子状态节点
	for child in get_children():
		if child is State:
			_register_state(child)
	
	# 进入初始状态
	if initial_state != &"" and initial_state in _states:
		_enter_state(initial_state)
	elif _states.size() > 0:
		_enter_state(_states.keys()[0])


func _process(delta: float) -> void:
	if current_state:
		current_state.update(delta)


func _physics_process(delta: float) -> void:
	if current_state:
		current_state.physics_update(delta)


func _unhandled_input(event: InputEvent) -> void:
	if current_state:
		current_state.handle_input(event)


## 切换到指定状态
func transition_to(state_name: StringName) -> void:
	if state_name == current_state_name:
		return
	if state_name not in _states:
		push_error("状态不存在: %s" % state_name)
		return
	
	var old_name: StringName = current_state_name
	
	# 退出当前状态
	if current_state:
		current_state.exit()
	
	# 进入新状态
	_enter_state(state_name)
	state_changed.emit(old_name, current_state_name)


## 动态添加状态（DLC 扩展用）
func add_state(state_node: State) -> void:
	add_child(state_node)
	_register_state(state_node)


## 动态移除状态
func remove_state(state_name: StringName) -> void:
	if state_name not in _states:
		return
	var state: State = _states[state_name]
	# 如果正在移除当前状态，先切换到其他状态
	if current_state == state and _states.size() > 1:
		for key in _states:
			if key != state_name:
				transition_to(key)
				break
	_states.erase(state_name)
	state.queue_free()


## 获取当前状态名称
func get_current_state_name() -> StringName:
	return current_state_name


## 检查是否处于指定状态
func is_in_state(state_name: StringName) -> bool:
	return current_state_name == state_name


## 注册状态
func _register_state(state: State) -> void:
	var state_name: StringName = StringName(state.name)
	_states[state_name] = state
	state.entity = _entity
	state.state_machine = self
	# 初始时禁用处理
	state.set_process(false)
	state.set_physics_process(false)


## 进入状态
func _enter_state(state_name: StringName) -> void:
	current_state = _states[state_name]
	current_state_name = state_name
	current_state.set_process(true)
	current_state.set_physics_process(true)
	current_state.enter()


## 查找祖先 Entity（支持 EcsEntity/EcsEntity3D/EcsEntity2D）
func _find_entity() -> Node:
	var node: Node = get_parent()
	while node:
		if node.has_method("has_all_components"):
			return node
		node = node.get_parent()
	return null
