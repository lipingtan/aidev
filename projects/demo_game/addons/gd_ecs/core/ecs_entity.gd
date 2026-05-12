class_name EcsEntity extends Node
## ECS Entity（extends Node）
##
## 纯逻辑 Entity，适用于不需要物理/渲染的游戏对象，
## 如：游戏管理器、触发器逻辑、全局状态容器。
##
## 如需 3D 物理角色，使用 EcsEntity3D。
## 如需 2D 物理角色，使用 EcsEntity2D。
## 如需挂载到已有节点，使用 EcsEntityNode。

## 当 Component 被添加时触发
signal component_added(component_name: StringName)

## 当 Component 被移除时触发
signal component_removed(component_name: StringName)

## 内部混入（封装所有 Component 管理逻辑）
var _mixin: EcsEntityMixin = EcsEntityMixin.new()

## 实体唯一 ID
var entity_id: int:
	get: return _mixin.entity_id
	set(value): _mixin.entity_id = value


func _init() -> void:
	_mixin.setup(self)
	_mixin.component_added.connect(func(n): component_added.emit(n))
	_mixin.component_removed.connect(func(n): component_removed.emit(n))


func _enter_tree() -> void:
	if EcsWorld:
		EcsWorld.register_entity(self)


func _exit_tree() -> void:
	if EcsWorld:
		EcsWorld.unregister_entity(self)


## 添加 Component
func add_component(component: EcsComponent) -> void:
	_mixin.add_component(component)


## 移除 Component
func remove_component(component_name: StringName) -> EcsComponent:
	return _mixin.remove_component(component_name)


## 获取 Component
func get_component(component_name: StringName) -> EcsComponent:
	return _mixin.get_component(component_name)


## 检查是否拥有 Component
func has_component(component_name: StringName) -> bool:
	return _mixin.has_component(component_name)


## 检查是否拥有所有指定的 Component
func has_all_components(component_names: Array[StringName]) -> bool:
	return _mixin.has_all_components(component_names)


## 获取所有 Component 名称列表
func get_component_names() -> Array[StringName]:
	return _mixin.get_component_names()


## 序列化所有 Component
func serialize_components() -> Dictionary:
	return _mixin.serialize_components()


## 反序列化所有 Component
func deserialize_components(data: Dictionary) -> void:
	_mixin.deserialize_components(data)
