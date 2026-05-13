class_name DlcPckLoader extends RefCounted
## DLC PCK 动态加载器
##
## 负责将下载好的 .pck 文件挂载到 Godot 虚拟文件系统，
## 使其内容可通过 res:// 路径访问。
##
## PCK 加载后，其中的文件会挂载到 res:// 下，
## 例如 PCK 内的 dlc/dark_realm/manifest.json
## 加载后可通过 res://dlc/dark_realm/manifest.json 访问。

## 已加载的 PCK 记录 { dlc_id: pck_path }
var _loaded_pcks: Dictionary = {}


## 加载 PCK 文件到虚拟文件系统
## pck_path: PCK 文件的本地路径（user:// 或绝对路径）
## dlc_id: DLC 唯一标识
## 返回：是否加载成功
func load_pck(pck_path: String, dlc_id: String) -> bool:
	if dlc_id in _loaded_pcks:
		push_warning("DLC PCK 已加载: %s" % dlc_id)
		return true

	if not FileAccess.file_exists(pck_path):
		push_error("PCK 文件不存在: %s" % pck_path)
		return false

	# 核心：将 PCK 挂载到 Godot 虚拟文件系统
	var success: bool = ProjectSettings.load_resource_pack(pck_path)
	if not success:
		push_error("PCK 加载失败: %s（文件可能损坏或格式不兼容）" % pck_path)
		return false

	_loaded_pcks[dlc_id] = pck_path
	print("[DlcPckLoader] PCK 已挂载: %s → res://" % pck_path)
	return true


## 检查 PCK 是否已加载
func is_pck_loaded(dlc_id: String) -> bool:
	return dlc_id in _loaded_pcks


## 获取已加载的 PCK 列表
func get_loaded_pcks() -> Array[String]:
	var result: Array[String] = []
	for key in _loaded_pcks:
		result.append(key)
	return result


## 注意：Godot 不支持卸载已加载的 PCK
## 一旦 PCK 挂载到虚拟文件系统，只能重启游戏才能完全卸载
## 但可以通过不再引用其中的资源来"逻辑卸载"
func unload_pck(dlc_id: String) -> void:
	if dlc_id not in _loaded_pcks:
		return
	_loaded_pcks.erase(dlc_id)
	push_warning(
		"[DlcPckLoader] PCK '%s' 已从记录中移除，但 Godot 不支持真正卸载 PCK。\n" +
		"资源仍在虚拟文件系统中，重启游戏后才会完全清除。" % dlc_id
	)


## 从 PCK 中读取 manifest.json
## 假设 PCK 内的 manifest 路径为 res://dlc/{dlc_id}/manifest.json
func read_manifest_from_pck(dlc_id: String) -> Dictionary:
	var manifest_path: String = "res://dlc/%s/manifest.json" % dlc_id
	if not FileAccess.file_exists(manifest_path):
		push_error("PCK 中未找到 manifest: %s" % manifest_path)
		return {}

	var file := FileAccess.open(manifest_path, FileAccess.READ)
	if not file:
		return {}

	var json := JSON.new()
	if json.parse(file.get_as_text()) != OK:
		file.close()
		return {}

	file.close()
	return json.data
