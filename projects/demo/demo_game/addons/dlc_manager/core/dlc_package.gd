class_name DlcPackage extends RefCounted
## DLC 包封装
##
## 管理单个 DLC 的加载状态和注册内容，支持应用和回退。
## 通过集成层与 ECS 框架和资产系统交互。

var manifest: DlcManifest = null
var is_loaded: bool = false

## 已注册的内容（用于卸载时回退）
var registered_components: Array[StringName] = []
var registered_systems: Array[EcsSystem] = []
var registered_data: Array[Dictionary] = []
var entity_extensions: Array[Dictionary] = []


## 应用 DLC 内容
func apply() -> bool:
	if not manifest or not manifest.is_valid():
		push_error("DLC manifest 无效: %s" % (manifest.id if manifest else "null"))
		return false
	
	var base: String = manifest.base_path
	var content: Dictionary = manifest.content
	
	# 1. 注册 Components（通过 ECS 集成层）
	var comp_paths: Array = content.get("components", [])
	if comp_paths.size() > 0:
		registered_components = DlcEcsIntegration.register_components(comp_paths, base)
	
	# 2. 注册 Systems（通过 ECS 集成层）
	var sys_paths: Array = content.get("systems", [])
	if sys_paths.size() > 0:
		registered_systems = DlcEcsIntegration.register_systems(sys_paths, base)
	
	# 3. 注册数据资源（通过资产集成层）
	var data_entries: Array = content.get("data", [])
	if data_entries.size() > 0:
		registered_data = DlcAssetIntegration.register_data(data_entries, base, manifest.id)
	
	# 4. 注册数据目录（通过资产集成层）
	var data_dirs: Array = content.get("data_dirs", [])
	if data_dirs.size() > 0:
		var dir_data: Array[Dictionary] = DlcAssetIntegration.register_data_dirs(data_dirs, base, manifest.id)
		registered_data.append_array(dir_data)
	
	# 5. 应用角色扩展（通过 ECS 集成层）
	var extensions: Array = content.get("character_extensions", [])
	if extensions.size() > 0:
		entity_extensions = DlcEcsIntegration.apply_entity_extensions(extensions, base)
	
	is_loaded = true
	return true


## 回退 DLC 内容
func revert() -> void:
	# 回退角色扩展
	DlcEcsIntegration.revert_entity_extensions(entity_extensions)
	entity_extensions.clear()
	
	# 注销 Systems
	DlcEcsIntegration.unregister_systems(registered_systems)
	registered_systems.clear()
	
	# 注销 Components
	DlcEcsIntegration.unregister_components(registered_components)
	registered_components.clear()
	
	# 移除数据
	DlcAssetIntegration.unregister_data(registered_data)
	registered_data.clear()
	
	is_loaded = false
