class_name DlcEcsIntegration extends RefCounted
## DLC 与 ECS 框架的集成层
##
## 负责将 DLC 包中的 Component 类型和 System 实例
## 注册到 EcsWorld，并在卸载时清理。

## 从 DLC 包路径加载并注册所有 Component 类型
static func register_components(component_paths: Array, base_path: String) -> Array[StringName]:
	var registered: Array[StringName] = []
	for rel_path in component_paths:
		var full_path: String = base_path + "/" + rel_path
		if not ResourceLoader.exists(full_path):
			push_warning("DLC Component 文件不存在: %s" % full_path)
			continue
		var script: Script = load(full_path)
		if not script:
			push_warning("DLC Component 脚本加载失败: %s" % full_path)
			continue
		# 创建临时实例获取组件名
		var instance: EcsComponent = script.new()
		var comp_name: StringName = instance.get_component_name()
		if comp_name == &"":
			push_warning("DLC Component 未定义 get_component_name(): %s" % full_path)
			continue
		EcsWorld.register_component_type(comp_name, script)
		registered.append(comp_name)
	return registered


## 从 DLC 包路径加载并注册所有 System
static func register_systems(system_paths: Array, base_path: String) -> Array[EcsSystem]:
	var registered: Array[EcsSystem] = []
	for rel_path in system_paths:
		var full_path: String = base_path + "/" + rel_path
		if not ResourceLoader.exists(full_path):
			push_warning("DLC System 文件不存在: %s" % full_path)
			continue
		var script: Script = load(full_path)
		if not script:
			push_warning("DLC System 脚本加载失败: %s" % full_path)
			continue
		var system: EcsSystem = script.new()
		EcsWorld.register_system(system)
		registered.append(system)
	return registered


## 注销所有已注册的 Component 类型
static func unregister_components(component_names: Array[StringName]) -> void:
	for comp_name in component_names:
		EcsWorld.unregister_component_type(comp_name)


## 注销所有已注册的 System
static func unregister_systems(systems: Array[EcsSystem]) -> void:
	for system in systems:
		EcsWorld.unregister_system(system)


## 将 DLC 的 Component 注入到现有 Entity（角色扩展）
static func apply_entity_extensions(extensions: Array, base_path: String) -> Array[Dictionary]:
	var applied: Array[Dictionary] = []
	for ext in extensions:
		if not ext is Dictionary:
			continue
		var target_name: String = ext.get("target_entity", "")
		var add_components: Array = ext.get("add_components", [])
		var add_data_path: String = ext.get("add_data", "")
		
		# 查找目标 Entity（按节点名匹配）
		for entity in EcsWorld.get_all_entities():
			if entity.name != target_name:
				continue
			
			# 挂载新 Component
			for comp_name in add_components:
				var component: EcsComponent = EcsWorld.create_component(StringName(comp_name))
				if component:
					# 如果有额外数据文件，加载并应用
					if add_data_path != "":
						var data_path: String = base_path + "/" + add_data_path
						if ResourceLoader.exists(data_path):
							var data: Resource = load(data_path)
							if data and data.has_method("apply_to_component"):
								data.apply_to_component(component)
					entity.add_component(component)
					applied.append({
						"entity_id": entity.entity_id,
						"component_name": StringName(comp_name),
					})
	return applied


## 回退 Entity 扩展
static func revert_entity_extensions(extensions: Array[Dictionary]) -> void:
	for ext in extensions:
		var entity_id: int = ext.get("entity_id", -1)
		var comp_name: StringName = ext.get("component_name", &"")
		if entity_id < 0 or comp_name == &"":
			continue
		for entity in EcsWorld.get_all_entities():
			if entity.entity_id == entity_id:
				entity.remove_component(comp_name)
				break
