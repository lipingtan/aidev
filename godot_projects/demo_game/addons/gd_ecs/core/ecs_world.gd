class_name EcsWorld extends Node
## ECS World 管理器（Autoload）
##
## 负责管理所有 Entity 和 System 的生命周期，
## 提供 Query 机制和 System 调度。
##
## 职责：
## - Entity 注册/注销
## - Component 类型注册
## - System 注册/注销/调度
## - Query 缓存管理

## 当新 Entity 注册时触发
signal entity_registered(entity: Node)

## 当 Entity 注销时触发
signal entity_unregistered(entity: Node)

## 当新 System 注册时触发
signal system_registered(system: EcsSystem)

## 已注册的 Entity { entity_id: Node }（支持 EcsEntity/EcsEntity3D/EcsEntity2D/EcsEntityNode）
var _entities: Dictionary = {}

## 已注册的 System（按 priority 排序）
var _process_systems: Array[EcsSystem] = []
var _physics_systems: Array[EcsSystem] = []

## Component 类型注册表 { component_name: Script }
var _component_registry: Dictionary = {}

## Query 缓存 { query_key: Array[EcsEntity] }
var _query_cache: Dictionary = {}

## 缓存是否需要刷新
var _cache_dirty: bool = false

## 下一个可用的 Entity ID
var _next_entity_id: int = 0


func _process(delta: float) -> void:
	_flush_cache_if_dirty()
	for system in _process_systems:
		if system.enabled:
			var entities: Array = _get_matching_entities(system)
			if entities.size() > 0:
				system.process(entities, delta)


func _physics_process(delta: float) -> void:
	_flush_cache_if_dirty()
	for system in _physics_systems:
		if system.enabled:
			var entities: Array = _get_matching_entities(system)
			if entities.size() > 0:
				system.process(entities, delta)


## 注册 Entity（支持 EcsEntity/EcsEntity3D/EcsEntity2D/EcsEntityNode）
func register_entity(entity: Node) -> void:
	if not entity.has_method("has_all_components"):
		push_error("注册的节点不具备 ECS Entity 接口")
		return
	entity.entity_id = _next_entity_id
	_next_entity_id += 1
	_entities[entity.entity_id] = entity
	_cache_dirty = true
	entity_registered.emit(entity)


## 注销 Entity
func unregister_entity(entity: Node) -> void:
	if entity.entity_id in _entities:
		_entities.erase(entity.entity_id)
		_cache_dirty = true
		entity_unregistered.emit(entity)


## 注册 System
func register_system(system: EcsSystem) -> void:
	if system.phase == &"physics_process":
		_physics_systems.append(system)
		_physics_systems.sort_custom(_sort_by_priority)
	else:
		_process_systems.append(system)
		_process_systems.sort_custom(_sort_by_priority)
	system.on_registered()
	system_registered.emit(system)


## 注销 System
func unregister_system(system: EcsSystem) -> void:
	_process_systems.erase(system)
	_physics_systems.erase(system)
	system.on_unregistered()


## 注册 Component 类型（用于动态创建和 DLC 扩展）
func register_component_type(component_name: StringName, script: Script) -> void:
	_component_registry[component_name] = script


## 注销 Component 类型
func unregister_component_type(component_name: StringName) -> void:
	_component_registry.erase(component_name)


## 创建已注册类型的 Component 实例
func create_component(component_name: StringName) -> EcsComponent:
	if component_name not in _component_registry:
		push_error("未注册的 Component 类型: %s" % component_name)
		return null
	var script: Script = _component_registry[component_name]
	return script.new()


## 查询匹配指定 Component 组合的所有 Entity
func query(component_names: Array[StringName]) -> Array:
	var key: String = _make_query_key(component_names)
	if key in _query_cache:
		return _query_cache[key]
	var result: Array = _execute_query(component_names)
	_query_cache[key] = result
	return result


## 获取所有已注册的 Entity
func get_all_entities() -> Array:
	var result: Array = []
	for entity in _entities.values():
		result.append(entity)
	return result


## 获取 Entity 数量
func get_entity_count() -> int:
	return _entities.size()


## 获取所有已注册的 System
func get_all_systems() -> Array[EcsSystem]:
	var result: Array[EcsSystem] = []
	result.append_array(_process_systems)
	result.append_array(_physics_systems)
	return result


## Entity Component 变更时的回调（由 Entity 调用）
func _on_entity_component_changed(_entity: Node) -> void:
	_cache_dirty = true


## 刷新查询缓存
func _flush_cache_if_dirty() -> void:
	if _cache_dirty:
		_query_cache.clear()
		_cache_dirty = false


## 执行查询
func _execute_query(component_names: Array[StringName]) -> Array:
	var result: Array = []
	for entity in _entities.values():
		if entity.has_all_components(component_names):
			result.append(entity)
	return result


## 获取 System 匹配的 Entity 列表
func _get_matching_entities(system: EcsSystem) -> Array:
	var query_components: Array[StringName] = system.get_query()
	if query_components.is_empty():
		return []
	return query(query_components)


## 生成查询缓存 key
func _make_query_key(component_names: Array[StringName]) -> String:
	var sorted: Array[StringName] = component_names.duplicate()
	sorted.sort()
	return ",".join(sorted)


## 按优先级排序
func _sort_by_priority(a: EcsSystem, b: EcsSystem) -> bool:
	return a.priority < b.priority
