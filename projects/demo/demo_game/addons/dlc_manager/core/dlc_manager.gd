extends Node
## DLC 管理器（Autoload）
##
## 统一管理 DLC 的完整生命周期：
##   下载 → 验证 → PCK加载 → 内容注册 → 运行时卸载
##
## 支持两种分发模式：
##   模式 A：平台托管（Steam/iOS/主机）— 平台验证后直接加载本地 PCK
##   模式 B：自建服务器 — 服务器授权 → 下载加密PCK → 验证 → 加载
##
## 依赖：
## - EcsWorld: 注册 DLC 的 Component/System
## - DataManager: 注册 DLC 的数据资源

## DLC 下载进度更新（dlc_id, downloaded_bytes, total_bytes）
signal dlc_download_progress(dlc_id: String, downloaded: int, total: int)

## DLC 下载完成（dlc_id）
signal dlc_download_completed(dlc_id: String)

## DLC 加载成功（dlc_id）
signal dlc_loaded(dlc_id: String)

## DLC 卸载（dlc_id）
signal dlc_unloaded(dlc_id: String)

## DLC 加载失败（dlc_id, error）
signal dlc_load_failed(dlc_id: String, error: String)

## 已加载的 DLC { dlc_id: DlcPackage }
var _loaded_dlcs: Dictionary = {}

## DLC 扫描路径（目录模式，开发期使用）
var _scan_paths: Array[String] = ["res://dlc/", "user://dlc/"]

## 游戏版本（用于兼容性检查）
var game_version: String = "0.1.0"

## 授权服务器地址（空 = 不做服务器授权验证）
var auth_server_url: String = ""

## 子模块
var _pck_loader: DlcPckLoader = DlcPckLoader.new()
var _verifier: DlcVerifier = DlcVerifier.new()
var _downloader: DlcDownloader = DlcDownloader.new()

## 待处理的下载队列 { dlc_id: { url, token, expected_hash } }
var _pending_downloads: Dictionary = {}


func _ready() -> void:
	# 连接子模块信号
	_downloader.download_completed.connect(_on_download_completed)
	_downloader.download_failed.connect(_on_download_failed)
	_verifier.verify_completed.connect(_on_verify_completed)
	_verifier.auth_completed.connect(_on_auth_completed)
	_verifier.auth_server_url = auth_server_url

	# 延迟加载，确保 EcsWorld 等 Autoload 已就绪
	call_deferred("_deferred_load_local_dlcs")


## 延迟加载本地已有的 DLC（目录模式）
func _deferred_load_local_dlcs() -> void:
	var manifests: Array[DlcManifest] = scan_dlcs()
	for manifest in manifests:
		load_dlc(manifest.id)


# ─────────────────────────────────────────────
# 模式 A：直接加载本地 DLC（目录或已下载的 PCK）
# ─────────────────────────────────────────────

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

	manifests.sort_custom(func(a, b): return a.load_order < b.load_order)
	return manifests


## 直接加载指定 DLC（目录模式，开发期使用）
func load_dlc(dlc_id: String) -> bool:
	if dlc_id in _loaded_dlcs:
		return true

	var manifest: DlcManifest = _find_manifest(dlc_id)
	if not manifest:
		dlc_load_failed.emit(dlc_id, "未找到 DLC: %s" % dlc_id)
		return false

	return _apply_dlc(manifest)


## 从已下载的 PCK 文件加载 DLC（模式 A：平台托管）
func load_dlc_from_pck(pck_path: String) -> bool:
	# 1. 加载 PCK 到虚拟文件系统
	if not _pck_loader.load_pck(pck_path, ""):
		return false

	# 2. 从 PCK 中读取 manifest（需要先知道 dlc_id）
	# 临时加载后读取 manifest 获取 dlc_id
	# PCK 内约定 manifest 路径：res://dlc/{dlc_id}/manifest.json
	# 由于不知道 dlc_id，先扫描 user://dlc/ 目录
	var manifests: Array[DlcManifest] = scan_dlcs()
	for manifest in manifests:
		if not (manifest.id in _loaded_dlcs):
			return _apply_dlc(manifest)

	return false


# ─────────────────────────────────────────────
# 模式 B：从服务器下载并加载 DLC
# ─────────────────────────────────────────────

## 从服务器下载 DLC（模式 B：自建服务器）
## url: DLC 下载地址
## dlc_id: DLC 唯一标识
## auth_token: 授权 Token（由服务器在购买后颁发）
## expected_sha256: 期望的文件哈希（由服务器下发，用于完整性验证）
func download_and_load_dlc(
		url: String,
		dlc_id: String,
		auth_token: String = "",
		expected_sha256: String = "") -> void:

	if dlc_id in _loaded_dlcs:
		push_warning("DLC %s 已加载" % dlc_id)
		return

	# 如果已下载，直接验证并加载
	if _downloader.is_downloaded(dlc_id):
		_pending_downloads[dlc_id] = {
			"token": auth_token,
			"expected_hash": expected_sha256,
		}
		_start_verification(dlc_id, expected_sha256)
		return

	# 记录待处理信息
	_pending_downloads[dlc_id] = {
		"url": url,
		"token": auth_token,
		"expected_hash": expected_sha256,
	}

	# 先做服务器授权验证（如果配置了服务器）
	if auth_server_url != "" and auth_token != "":
		_verifier.verify_auth(dlc_id, auth_token)
	else:
		# 无需授权，直接下载
		_start_download(dlc_id)


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
			unload_dlc(other_id)

	var package: DlcPackage = _loaded_dlcs[dlc_id]
	package.revert()
	_loaded_dlcs.erase(dlc_id)

	# PCK 模式：记录已卸载（但 PCK 仍在虚拟文件系统中）
	if _pck_loader.is_pck_loaded(dlc_id):
		_pck_loader.unload_pck(dlc_id)

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


## 检查 DLC 是否已下载到本地
func is_dlc_downloaded(dlc_id: String) -> bool:
	return _downloader.is_downloaded(dlc_id)


## 添加 DLC 扫描路径
func add_scan_path(path: String) -> void:
	if path not in _scan_paths:
		_scan_paths.append(path)


## 配置授权服务器
func set_auth_server(url: String) -> void:
	auth_server_url = url
	_verifier.auth_server_url = url


# ─────────────────────────────────────────────
# 内部方法
# ─────────────────────────────────────────────

## 应用 DLC 内容（目录模式和 PCK 模式共用）
func _apply_dlc(manifest: DlcManifest) -> bool:
	# 版本兼容性检查
	if not _check_version_compatibility(manifest):
		dlc_load_failed.emit(manifest.id,
			"版本不兼容: 需要 %s ~ %s，当前 %s" % [
				manifest.min_game_version, manifest.max_game_version, game_version
			])
		return false

	# 依赖检查
	for dep in manifest.dependencies:
		if dep not in _loaded_dlcs:
			if not load_dlc(dep):
				dlc_load_failed.emit(manifest.id, "依赖未满足: %s" % dep)
				return false

	# 冲突检查
	for conflict in manifest.conflicts:
		if conflict in _loaded_dlcs:
			dlc_load_failed.emit(manifest.id, "与已加载的 DLC 冲突: %s" % conflict)
			return false

	# 创建并应用 DLC 包
	var package := DlcPackage.new()
	package.manifest = manifest

	if package.apply():
		_loaded_dlcs[manifest.id] = package
		dlc_loaded.emit(manifest.id)
		return true
	else:
		dlc_load_failed.emit(manifest.id, "DLC 内容应用失败")
		return false


## 开始下载
func _start_download(dlc_id: String) -> void:
	var info: Dictionary = _pending_downloads.get(dlc_id, {})
	var url: String = info.get("url", "")
	var token: String = info.get("token", "")

	if url == "":
		dlc_load_failed.emit(dlc_id, "下载地址为空")
		return

	# 将 HTTPRequest 节点添加到场景树（必须）
	var http := HTTPRequest.new()
	add_child(http)
	_downloader._active_requests[dlc_id] = http

	var headers: PackedStringArray = []
	if token != "":
		headers.append("Authorization: Bearer %s" % token)

	var save_path: String = DlcDownloader.SAVE_DIR + dlc_id + ".pck"
	http.download_file = save_path
	http.request_completed.connect(
		func(result, response_code, _headers, _body):
			_downloader._on_request_completed(dlc_id, result, response_code, save_path)
	)

	var err: int = http.request(url, headers, HTTPClient.METHOD_GET)
	if err != OK:
		_downloader._active_requests.erase(dlc_id)
		dlc_load_failed.emit(dlc_id, "HTTP 请求发起失败")


## 开始验证
func _start_verification(dlc_id: String, expected_hash: String) -> void:
	var pck_path: String = DlcDownloader.SAVE_DIR + dlc_id + ".pck"
	_verifier.verify_file(pck_path, dlc_id, expected_hash)


## 下载完成回调
func _on_download_completed(dlc_id: String, save_path: String, success: bool) -> void:
	if not success:
		_pending_downloads.erase(dlc_id)
		return

	dlc_download_completed.emit(dlc_id)

	# 开始文件验证
	var info: Dictionary = _pending_downloads.get(dlc_id, {})
	var expected_hash: String = info.get("expected_hash", "")
	_start_verification(dlc_id, expected_hash)


## 下载失败回调
func _on_download_failed(dlc_id: String, error: String) -> void:
	_pending_downloads.erase(dlc_id)
	dlc_load_failed.emit(dlc_id, "下载失败: %s" % error)


## 文件验证完成回调
func _on_verify_completed(dlc_id: String, success: bool, error: String) -> void:
	if not success:
		_pending_downloads.erase(dlc_id)
		dlc_load_failed.emit(dlc_id, "文件验证失败: %s" % error)
		return

	# 验证通过，加载 PCK
	var pck_path: String = DlcDownloader.SAVE_DIR + dlc_id + ".pck"
	if not _pck_loader.load_pck(pck_path, dlc_id):
		_pending_downloads.erase(dlc_id)
		dlc_load_failed.emit(dlc_id, "PCK 加载失败")
		return

	# 从 PCK 中读取 manifest 并应用
	var manifest_data: Dictionary = _pck_loader.read_manifest_from_pck(dlc_id)
	if manifest_data.is_empty():
		_pending_downloads.erase(dlc_id)
		dlc_load_failed.emit(dlc_id, "PCK 中未找到有效 manifest")
		return

	var manifest: DlcManifest = DlcManifest.from_dict(
		manifest_data,
		"res://dlc/%s" % dlc_id
	)
	_pending_downloads.erase(dlc_id)
	_apply_dlc(manifest)


## 服务器授权验证完成回调
func _on_auth_completed(dlc_id: String, authorized: bool) -> void:
	if not authorized:
		_pending_downloads.erase(dlc_id)
		dlc_load_failed.emit(dlc_id, "服务器授权验证失败，请确认已购买此 DLC")
		return

	# 授权通过，开始下载
	_start_download(dlc_id)


## 加载 manifest.json（目录模式）
func _load_manifest(path: String, base_path: String) -> DlcManifest:
	if not FileAccess.file_exists(path):
		return null

	var file := FileAccess.open(path, FileAccess.READ)
	if not file:
		return null

	var json := JSON.new()
	if json.parse(file.get_as_text()) != OK:
		file.close()
		push_warning("DLC manifest 解析失败: %s" % path)
		return null
	file.close()

	var manifest: DlcManifest = DlcManifest.from_dict(json.data, base_path)
	if not manifest.is_valid():
		push_warning("DLC manifest 无效: %s" % path)
		return null

	return manifest


## 查找已扫描的 manifest
func _find_manifest(dlc_id: String) -> DlcManifest:
	for scan_path in _scan_paths:
		var manifest_path: String = scan_path + dlc_id + "/manifest.json"
		var manifest: DlcManifest = _load_manifest(manifest_path, scan_path + dlc_id)
		if manifest:
			return manifest
	return null


## 版本兼容性检查（语义化版本简化比较）
func _check_version_compatibility(manifest: DlcManifest) -> bool:
	return game_version >= manifest.min_game_version
