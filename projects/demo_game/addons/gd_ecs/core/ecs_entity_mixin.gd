class_name EcsEntityMixin extends RefCounted
## ECS Entity 能力混入（核心逻辑）
##
## 封装所有 Component 管理逻辑，不继承任何 Node。
## 被 EcsEntity、EcsEntity3D、EcsEntity2D、EcsEntityNode 共同持有，
## 避免代码重复。
##
## 使用方式：
## 各 Entity 类内部创建 _mixin = EcsEntityMixin.new()，
## 然后将公开方法委托给 _mixin。

## 当 Component 被添加时触发（由宿主节点转发信号）
signal component_added(component_name: StringName)

## 当 Component 被移除时触发
signal component_removed(component_name: StringName)

## 实体唯一 ID（由 EcsWorld 分配）
var entity_id: int = -1

## 宿主节点引用（挂载 ECS 能力的实际 Node）
var host_node: Node = null

## 已挂载的 Component 字典 { component_name: EcsComponent }
var _components: Dictionary = {}


## 初始化，绑定宿主节点
func setup(host: Node) -> void:
	host_node = host


## 添加 Component
func add_component(component: EcsComponent) -> void:
	var comp_name: StringName = component.get_component_name()
	if comp_name == &"":
		push_error("Component 未定义 get_component_name()")
		return
	_components[comp_name] = component
	component_added.emit(comp_name)
	# 通知 EcsWorld 更新查询缓存（通过宿主节点获取 Autoload）
	var world: Node = _get_ecs_world()
	if world:
		world._on_entity_component_changed(host_node)


## 移除 Component
func remove_component(component_name: StringName) -> EcsComponent:
	if component_name not in _components:
		return null
	var component: EcsComponent = _components[component_name]
	_components.erase(component_name)
	component_removed.emit(component_name)
	var world: Node = _get_ecs_world()
	if world:
		world._on_entity_component_changed(host_node)
	return component


## 获取 Component
func get_component(component_name: StringName) -> EcsComponent:
	return _components.get(component_name)


## 检查是否拥有 Component
func has_component(component_name: StringName) -> bool:
	return component_name in _components


## 检查是否拥有所有指定的 Component
func has_all_components(component_names: Array[StringName]) -> bool:
	for comp_name in component_names:
		if comp_name not in _components:
			return false
	return true


## 获取所有 Component 名称列表
func get_component_names() -> Array[StringName]:
	var names: Array[StringName] = []
	for key in _components:
		names.append(key)
	return names


## 序列化所有 Component（用于存档）
func serialize_components() -> Dictionary:
	var data: Dictionary = {}
	for comp_name in _components:
		data[comp_name] = _components[comp_name].serialize()
	return data


## 反序列化所有 Component（用于读档）
func deserialize_components(data: Dictionary) -> void:
	for comp_name in data:
		if comp_name in _components:
			_components[comp_name].deserialize(data[comp_name])


## 安全获取 EcsWorld Autoload（通过宿主节点的场景树）
func _get_ecs_world() -> Node:
	if host_node and host_node.is_inside_tree():
		return host_node.get_node_or_null("/root/EcsWorld")
	return null
