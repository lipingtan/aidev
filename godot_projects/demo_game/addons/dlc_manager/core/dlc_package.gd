class_name DlcPackage extends RefCounted
## DLC 包封装
##
## 管理单个 DLC 的加载状态和注册内容，支持应用和回退。

var manifest: DlcManifest = null
var is_loaded: bool = false

## 已注册的内容（用于卸载时回退）
var registered_components: Array[StringName] = []
var registered_systems: Array[EcsSystem] = []
var registered_data: Array[Dictionary] = []  # [{ table, id }]
var entity_extensions: Array[Dictionary] = []  # [{ entity_id, component_name }]


## 应用 DLC 内容
func apply() -> bool:
	if not manifest or not manifest.is_valid():
		return false
	
	var base: String = manifest.base_path
	
	# 注册 Components
	var components: Array = manifest.content.get("components", [])
	for comp_path in components:
		var full_path: String = base + "/" + comp_path
		var script: Script = load(full_path)
		if script:
			var instance: EcsComponent = script.new()
			var comp_name: StringName = instance.get_component_name()
			if comp_name != &"":
				EcsWorld.register_component_type(comp_name, script)
				registered_components.append(comp_name)
	
	# 注册 Systems
	var systems: Array = manifest.content.get("systems", [])
	for sys_path in systems:
		var full_path: String = base + "/" + sys_path
		var script: Script = load(full_path)
		if script:
			var system: EcsSystem = script.new()
			EcsWorld.register_system(system)
			registered_systems.append(system)
	
	# 注册数据资源
	var data_dirs: Array = manifest.content.get("data", [])
	for data_entry in data_dirs:
		if data_entry is Dictionary:
			var table: StringName = StringName(data_entry.get("table", ""))
			var file_path: String = base + "/" + data_entry.get("path", "")
			var res: Resource = load(file_path)
			if res and "id" in res and table != &"":
				DataManager.add_data(table, res.id, res)
				registered_data.append({"table": table, "id": res.id})
	
	# 应用角色扩展
	var extensions: Array = manifest.content.get("character_extensions", [])
	for ext in extensions:
		_apply_extension(ext, base)
	
	is_loaded = true
	return true


## 回退 DLC 内容
func revert() -> void:
	# 回退角色扩展
	for ext in entity_extensions:
		var entity: EcsEntity = _find_entity_by_id(ext.entity_id)
		if entity:
			entity.remove_component(ext.component_name)
	entity_extensions.clear()
	
	# 注销 Systems
	for system in registered_systems:
		EcsWorld.unregister_system(system)
	registered_systems.clear()
	
	# 注销 Components
	for comp_name in registered_components:
		EcsWorld.unregister_component_type(comp_name)
	registered_components.clear()
	
	# 移除数据
	for entry in registered_data:
		DataManager.remove_data(entry.table, entry.id)
	registered_data.clear()
	
	is_loaded = false


## 应用角色扩展
func _apply_extension(ext: Dictionary, base: String) -> void:
	var target: String = ext.get("target_entity", "")
	var add_components: Array = ext.get("add_components", [])
	
	# 查找目标 Entity（通过名称匹配）
	for entity in EcsWorld.get_all_entities():
		if entity.name == target:
			for comp_name in add_components:
				var component: EcsComponent = EcsWorld.create_component(StringName(comp_name))
				if component:
					entity.add_component(component)
					entity_extensions.append({
						"entity_id": entity.entity_id,
						"component_name": StringName(comp_name)
					})


## 查找 Entity
func _find_entity_by_id(id: int) -> EcsEntity:
	var entities: Array[EcsEntity] = EcsWorld.get_all_entities()
	for entity in entities:
		if entity.entity_id == id:
			return entity
	return null
