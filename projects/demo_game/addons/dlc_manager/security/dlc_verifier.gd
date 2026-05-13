class_name DlcVerifier extends RefCounted
## DLC 安全验证器
##
## 提供两层安全机制：
## 1. 文件完整性验证（SHA-256 哈希校验）
## 2. 服务器授权验证（Token 验证，可选）
##
## 使用方式：
##   var verifier := DlcVerifier.new()
##   verifier.verify_completed.connect(_on_verify_done)
##   verifier.verify_file("user://dlc/dark_realm.pck", "dark_realm", expected_hash)

## 验证完成时触发（dlc_id, success, error_message）
signal verify_completed(dlc_id: String, success: bool, error: String)

## 服务器授权验证完成时触发（dlc_id, authorized）
signal auth_completed(dlc_id: String, authorized: bool)

## 授权服务器地址（空 = 不做服务器验证）
var auth_server_url: String = ""

## 当前活跃的授权请求
var _auth_requests: Dictionary = {}


## 验证本地 PCK 文件的完整性
## expected_sha256: 期望的 SHA-256 哈希值（十六进制字符串，由服务器下发）
## 如果 expected_sha256 为空，跳过哈希验证
func verify_file(file_path: String, dlc_id: String, expected_sha256: String = "") -> void:
	if not FileAccess.file_exists(file_path):
		verify_completed.emit(dlc_id, false, "文件不存在: %s" % file_path)
		return

	# 文件大小检查
	var file := FileAccess.open(file_path, FileAccess.READ)
	if not file:
		verify_completed.emit(dlc_id, false, "无法打开文件: %s" % file_path)
		return

	var file_size: int = file.get_length()
	if file_size == 0:
		file.close()
		verify_completed.emit(dlc_id, false, "文件为空: %s" % file_path)
		return

	# SHA-256 哈希验证
	if expected_sha256 != "":
		var actual_hash: String = _compute_sha256(file)
		file.close()
		if actual_hash.to_lower() != expected_sha256.to_lower():
			verify_completed.emit(dlc_id, false,
				"哈希校验失败：期望 %s，实际 %s" % [expected_sha256, actual_hash])
			return
	else:
		file.close()

	verify_completed.emit(dlc_id, true, "")


## 向服务器验证授权（Token 是否有效）
## device_id: 设备唯一标识（防止 Token 共享）
func verify_auth(dlc_id: String, auth_token: String, device_id: String = "") -> void:
	if auth_server_url == "":
		# 未配置服务器，跳过授权验证
		auth_completed.emit(dlc_id, true)
		return

	if dlc_id in _auth_requests:
		push_warning("DLC %s 授权验证正在进行中" % dlc_id)
		return

	var http := HTTPRequest.new()
	_auth_requests[dlc_id] = http

	# 构建验证请求体
	var body: Dictionary = {
		"dlc_id": dlc_id,
		"token": auth_token,
		"device_id": device_id if device_id != "" else _get_device_id(),
	}

	var headers: PackedStringArray = [
		"Content-Type: application/json",
		"Authorization: Bearer %s" % auth_token,
	]

	http.request_completed.connect(
		func(result, response_code, _headers, response_body):
			_on_auth_completed(dlc_id, result, response_code, response_body)
	)

	var err: int = http.request(
		auth_server_url + "/api/dlc/verify",
		headers,
		HTTPClient.METHOD_POST,
		JSON.stringify(body)
	)

	if err != OK:
		_auth_requests.erase(dlc_id)
		auth_completed.emit(dlc_id, false)


## 计算文件的 SHA-256 哈希
func _compute_sha256(file: FileAccess) -> String:
	var ctx := HashingContext.new()
	ctx.start(HashingContext.HASH_SHA256)

	# 分块读取，避免大文件占用过多内存
	const CHUNK_SIZE: int = 65536  # 64KB
	while file.get_position() < file.get_length():
		var chunk: PackedByteArray = file.get_buffer(CHUNK_SIZE)
		ctx.update(chunk)

	var hash_bytes: PackedByteArray = ctx.finish()
	return hash_bytes.hex_encode()


## 授权验证回调
func _on_auth_completed(dlc_id: String, result: int, response_code: int, body: PackedByteArray) -> void:
	_auth_requests.erase(dlc_id)

	if result != HTTPRequest.RESULT_SUCCESS or response_code != 200:
		auth_completed.emit(dlc_id, false)
		return

	# 解析服务器响应
	var json := JSON.new()
	if json.parse(body.get_string_from_utf8()) != OK:
		auth_completed.emit(dlc_id, false)
		return

	var data: Dictionary = json.data
	var authorized: bool = data.get("authorized", false)
	auth_completed.emit(dlc_id, authorized)


## 获取设备唯一标识（用于绑定授权）
func _get_device_id() -> String:
	# 优先使用持久化的设备 ID
	var config := ConfigFile.new()
	var config_path: String = "user://device.cfg"
	if config.load(config_path) == OK:
		var stored_id: String = config.get_value("device", "id", "")
		if stored_id != "":
			return stored_id

	# 首次运行：生成并持久化设备 ID
	var new_id: String = _generate_device_id()
	config.set_value("device", "id", new_id)
	config.save(config_path)
	return new_id


## 生成设备 ID（基于系统信息的哈希）
func _generate_device_id() -> String:
	var raw: String = "%s_%s_%d" % [
		OS.get_unique_id(),
		OS.get_model_name(),
		Time.get_unix_time_from_system(),
	]
	var ctx := HashingContext.new()
	ctx.start(HashingContext.HASH_SHA256)
	ctx.update(raw.to_utf8_buffer())
	return ctx.finish().hex_encode().left(32)
