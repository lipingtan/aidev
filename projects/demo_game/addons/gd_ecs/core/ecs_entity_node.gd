class_name EcsEntityNode extends Node
## ECS Entity 节点（组合模式）
##
## 作为子节点挂载到任何已有节点上，赋予其 ECS 能力。
## 适用于不想改变现有节点继承关系的场景。
##
## 使用方式：
## 1. 将 EcsEntityNode 作为子节点添加到目标节点
## 2. 通过 EcsEntityNode 操作 Component
## 3. EcsWorld 注册的是父节点（host），而非 EcsEntityNode 自身
##
## 示例场景树：
## StaticBody3D（宝箱）
## ├── EcsEntityNode        ← 赋予 ECS 能力
## ├── MeshInstance3D
## └── CollisionShape3D

## 当 Component 被添加时触发
signal component_added(component_name: StringName)

## 当 Component 被移除时触发
signal component_removed(component_name: StringName)

## 内部混入
var _mixin: EcsEntityMixin = EcsEntityMixin.new()

## 实体唯一 ID
var entity_id: int:
	get: return _mixin.entity_id
	set(value): _mixin.entity_id = value

## 宿主节点（父节点）
var host: Node:
	get: return get_parent()


func _init() -> void:
	_mixin.component_added.connect(func(n): component_added.emit(n))
	_mixin.component_removed.connect(func(n): component_removed.emit(n))


func _enter_tree() -> void:
	# 绑定到父节点（如果有）
	var parent: Node = get_parent()
	if parent:
		_mixin.setup(parent)
	else:
		_mixin.setup(self)
		push_warning("EcsEntityNode 没有父节点，将自身作为宿主")
	if EcsWorld:
		# 注册自身为 Entity
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
