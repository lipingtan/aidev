extends Node
## DLC 管理器（Autoload）
##
## 负责扫描、验证、加载和卸载 DLC 包。
## 支持目录格式和 PCK 格式两种分发方式。
##
## 依赖：
## - EcsWorld: 注册 DLC 的 Component/System
## - DataManager: 注册 DLC 的数据资源

## DLC 加载成功时触发
signal dlc_loaded(dlc_id: String)

## DLC 卸载时触发
signal dlc_unloaded(dlc_id: String)

## DLC 加载失败时触发
signal dlc_load_failed(dlc_id: String, error: String)

## 已加载的 DLC { dlc_id: DlcPackage }
var _loaded_dlcs: Dictionary = {}

## DLC 扫描路径列表
var _scan_paths: Array[String] = ["res://dlc/", "user://dlc/"]

## 游戏版本（用于兼容性检查）
var game_version: String = "0.1.0"


func _ready() -> void:
	# 延迟到下一帧加载 DLC，确保 EcsWorld 等 Autoload 已就绪
	call_deferred("_deferred_load_dlcs")


## 延迟加载所有 DLC
func _deferred_load_dlcs() -> void:
	var manifests: Array[DlcManifest] = scan_dlcs()
	for manifest in manifests:
		load_dlc(manifest.id)


## 扫描所有 DLC 目录，返回发现的清单列表
func scan_dlcs() -> Array[DlcManifest]:
	var manifests: Array[DlcManifest] = []
	
	for scan_path in _scan_paths:
		var dir := DirAccess.open(scan_path)
		if not dir:
			continue
		
		dir.list_dir_begin()
		var dir_name: String = dir.get_next()
		while dir_name != "":
			if dir.current_is_dir() and dir_name != "." and dir_name != "..":
				var manifest_path: String = scan_path + dir_name + "/manifest.json"
				var manifest: DlcManifest = _load_manifest(manifest_path, scan_path + dir_name)
				if manifest:
					manifests.append(manifest)
			dir_name = dir.get_next()
	
	# 按 load_order 排序
	manifests.sort_custom(func(a, b): return a.load_order < b.load_order)
	return manifests


## 加载指定 DLC
func load_dlc(dlc_id: String) -> bool:
	# 检查是否已加载
	if dlc_id in _loaded_dlcs:
		return true
	
	# 查找清单
	var manifest: DlcManifest = _find_manifest(dlc_id)
	if not manifest:
		dlc_load_failed.emit(dlc_id, "未找到 DLC: %s" % dlc_id)
		return false
	
	# 版本兼容性检查
	if not _check_version_compatibility(manifest):
		dlc_load_failed.emit(dlc_id, "版本不兼容: 需要 %s ~ %s" % [manifest.min_game_version, manifest.max_game_version])
		return false
	
	# 依赖检查
	for dep in manifest.dependencies:
		if dep not in _loaded_dlcs:
			# 尝试递归加载依赖
			if not load_dlc(dep):
				dlc_load_failed.emit(dlc_id, "依赖未满足: %s" % dep)
				return false
	
	# 冲突检查
	for conflict in manifest.conflicts:
		if conflict in _loaded_dlcs:
			dlc_load_failed.emit(dlc_id, "与已加载的 DLC 冲突: %s" % conflict)
			return false
	
	# 创建并应用 DLC 包
	var package := DlcPackage.new()
	package.manifest = manifest
	
	if package.apply():
		_loaded_dlcs[dlc_id] = package
		dlc_loaded.emit(dlc_id)
		return true
	else:
		dlc_load_failed.emit(dlc_id, "DLC 应用失败")
		return false


## 卸载指定 DLC
func unload_dlc(dlc_id: String) -> bool:
	if dlc_id not in _loaded_dlcs:
		return false
	
	# 检查是否有其他 DLC 依赖此 DLC
	for other_id in _loaded_dlcs:
		if other_id == dlc_id:
			continue
		var other: DlcPackage = _loaded_dlcs[other_id]
		if dlc_id in other.manifest.dependencies:
			# 先卸载依赖方
			unload_dlc(other_id)
	
	# 回退 DLC 内容
	var package: DlcPackage = _loaded_dlcs[dlc_id]
	package.revert()
	_loaded_dlcs.erase(dlc_id)
	dlc_unloaded.emit(dlc_id)
	return true


## 获取已加载的 DLC 列表
func get_loaded_dlcs() -> Array[String]:
	var result: Array[String] = []
	for key in _loaded_dlcs:
		result.append(key)
	return result


## 检查 DLC 是否已加载
func is_dlc_loaded(dlc_id: String) -> bool:
	return dlc_id in _loaded_dlcs


## 添加 DLC 扫描路径
func add_scan_path(path: String) -> void:
	if path not in _scan_paths:
		_scan_paths.append(path)


## 加载 manifest.json
func _load_manifest(path: String, base_path: String) -> DlcManifest:
	if not FileAccess.file_exists(path):
		return null
	
	var file := FileAccess.open(path, FileAccess.READ)
	if not file:
		return null
	
	var json_text: String = file.get_as_text()
	file.close()
	
	var json := JSON.new()
	if json.parse(json_text) != OK:
		push_warning("DLC manifest 解析失败: %s" % path)
		return null
	
	var data: Dictionary = json.data
	var manifest: DlcManifest = DlcManifest.from_dict(data, base_path)
	
	if not manifest.is_valid():
		push_warning("DLC manifest 无效: %s" % path)
		return null
	
	return manifest


## 查找已扫描的 manifest
func _find_manifest(dlc_id: String) -> DlcManifest:
	var manifests: Array[DlcManifest] = scan_dlcs()
	for manifest in manifests:
		if manifest.id == dlc_id:
			return manifest
	return null


## 版本兼容性检查（简化版：字符串比较）
func _check_version_compatibility(manifest: DlcManifest) -> bool:
	# 简化实现：只检查 min_game_version
	return game_version >= manifest.min_game_version
