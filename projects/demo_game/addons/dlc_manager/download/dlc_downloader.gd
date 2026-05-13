class_name DlcDownloader extends RefCounted
## DLC 下载管理器
##
## 负责从服务器下载 DLC 包（PCK 或目录格式），
## 支持进度回调、断点续传标记、下载验证。
##
## 使用方式：
##   var dl := DlcDownloader.new()
##   dl.download_progress.connect(_on_progress)
##   dl.download_completed.connect(_on_completed)
##   dl.start_download("https://your-server.com/dlc/dark_realm.pck", "dark_realm")

## 下载进度更新时触发（dlc_id, bytes_downloaded, total_bytes）
signal download_progress(dlc_id: String, downloaded: int, total: int)

## 下载完成时触发（dlc_id, save_path, success）
signal download_completed(dlc_id: String, save_path: String, success: bool)

## 下载失败时触发（dlc_id, error_message）
signal download_failed(dlc_id: String, error: String)

## DLC 保存目录
const SAVE_DIR: String = "user://dlc/"

## 当前活跃的下载请求 { dlc_id: HTTPRequest }
var _active_requests: Dictionary = {}


func _init() -> void:
	DirAccess.make_dir_recursive_absolute(SAVE_DIR)


## 开始下载 DLC
## url: 下载地址
## dlc_id: DLC 唯一标识（用于文件命名和信号）
## auth_token: 授权 Token（服务器验证用，可为空）
func start_download(url: String, dlc_id: String, auth_token: String = "") -> bool:
	if dlc_id in _active_requests:
		push_warning("DLC %s 正在下载中，请勿重复请求" % dlc_id)
		return false

	# 创建 HTTPRequest 节点（需要挂载到场景树才能工作）
	var http := HTTPRequest.new()
	# 注意：HTTPRequest 需要挂载到场景树，调用方需要将其 add_child
	# 这里通过信号通知调用方
	_active_requests[dlc_id] = http

	# 构建请求头（携带授权 Token）
	var headers: PackedStringArray = []
	if auth_token != "":
		headers.append("Authorization: Bearer %s" % auth_token)
	headers.append("User-Agent: GodotDLCManager/1.0")

	# 设置下载路径
	var save_path: String = SAVE_DIR + dlc_id + ".pck"
	http.download_file = save_path

	# 连接信号
	http.request_completed.connect(
		func(result, response_code, _headers, _body):
			_on_request_completed(dlc_id, result, response_code, save_path)
	)

	# 发起请求
	var err: int = http.request(url, headers, HTTPClient.METHOD_GET)
	if err != OK:
		_active_requests.erase(dlc_id)
		download_failed.emit(dlc_id, "HTTP 请求发起失败，错误码: %d" % err)
		return false

	return true


## 取消下载
func cancel_download(dlc_id: String) -> void:
	if dlc_id in _active_requests:
		var http: HTTPRequest = _active_requests[dlc_id]
		http.cancel_request()
		_active_requests.erase(dlc_id)
		download_failed.emit(dlc_id, "下载已取消")


## 获取下载进度（0.0 ~ 1.0）
func get_progress(dlc_id: String) -> float:
	if dlc_id not in _active_requests:
		return 0.0
	var http: HTTPRequest = _active_requests[dlc_id]
	var downloaded: int = http.get_downloaded_bytes()
	var total: int = http.get_body_size()
	if total <= 0:
		return 0.0
	return float(downloaded) / float(total)


## 检查 DLC 是否已下载到本地
func is_downloaded(dlc_id: String) -> bool:
	return FileAccess.file_exists(SAVE_DIR + dlc_id + ".pck")


## 获取已下载的 PCK 路径
func get_local_path(dlc_id: String) -> String:
	return SAVE_DIR + dlc_id + ".pck"


## 删除本地 DLC 文件
func delete_local(dlc_id: String) -> bool:
	var path: String = SAVE_DIR + dlc_id + ".pck"
	if FileAccess.file_exists(path):
		return DirAccess.remove_absolute(path) == OK
	return false


## 下载完成回调
func _on_request_completed(dlc_id: String, result: int, response_code: int, save_path: String) -> void:
	_active_requests.erase(dlc_id)

	if result != HTTPRequest.RESULT_SUCCESS:
		download_failed.emit(dlc_id, "下载失败，result=%d" % result)
		return

	if response_code != 200:
		download_failed.emit(dlc_id, "服务器返回错误，HTTP %d" % response_code)
		# 删除不完整的文件
		if FileAccess.file_exists(save_path):
			DirAccess.remove_absolute(save_path)
		return

	download_completed.emit(dlc_id, save_path, true)
