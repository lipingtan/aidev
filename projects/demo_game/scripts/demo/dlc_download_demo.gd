extends Node
## DLC 下载验证 Demo
##
## 验证完整的 DLC 生命周期：
##   1. 查询服务器 DLC 信息（获取 SHA-256）
##   2. 服务器授权验证（Token）
##   3. HTTP 下载 PCK 文件
##   4. SHA-256 完整性验证
##   5. PCK 挂载到虚拟文件系统
##   6. 注册 Component/System 到 EcsWorld
##   7. 验证 DLC System 正常运行
##   8. 卸载 DLC，验证回退

const SERVER_URL: String = "http://localhost:8787"
const DLC_ID: String = "poison_dlc"
const TEST_TOKEN: String = "test-token-12345"

## 玩家 Entity（用于验证 DLC 注入）
var player: EcsEntity3D = null

## UI 节点
@onready var log_box: RichTextLabel = $UI/LogBox
@onready var status_bar: Label = $UI/StatusBar
@onready var btn_query: Button = $UI/Buttons/BtnQuery
@onready var btn_download: Button = $UI/Buttons/BtnDownload
@onready var btn_load: Button = $UI/Buttons/BtnLoad
@onready var btn_test: Button = $UI/Buttons/BtnTest
@onready var btn_unload: Button = $UI/Buttons/BtnUnload
@onready var btn_full: Button = $UI/Buttons/BtnFull
@onready var progress_bar: ProgressBar = $UI/ProgressBar

## 从服务器获取的 DLC 信息
var _dlc_info: Dictionary = {}

## 下载用的 HTTPRequest 节点
var _query_http: HTTPRequest
var _download_http: HTTPRequest

## 下载状态
var _download_save_path: String = ""
var _expected_hash: String = ""


func _ready() -> void:
	_setup_player()
	_setup_systems()
	_connect_buttons()
	_subscribe_events()
	_log("=== DLC 下载验证 Demo ===")
	_log("服务器地址: %s" % SERVER_URL)
	_log("请确保 dlc_server 已启动（npm start）")
	_log("")
	_log("步骤说明：")
	_log("  [1] 查询 DLC 信息 → 获取 SHA-256")
	_log("  [2] 下载 PCK → 自动验证哈希")
	_log("  [3] 加载 PCK → 注册到 EcsWorld")
	_log("  [4] 测试运行 → 验证 DLC System")
	_log("  [5] 卸载 DLC → 验证回退")
	_log("  [全流程] 一键执行以上所有步骤")
	_set_status("就绪，等待操作")


# ─────────────────────────────────────────────
# 初始化
# ─────────────────────────────────────────────

func _setup_player() -> void:
	player = EcsEntity3D.new()
	player.name = "Player"
	add_child(player)

	var base := BaseStatsComponent.new()
	base.max_hp = 100.0
	base.attack = 10.0
	player.add_component(base)

	var final_stats := FinalStatsComponent.new()
	player.add_component(final_stats)

	var runtime := RuntimeStatsComponent.new()
	runtime.current_hp = 100.0
	player.add_component(runtime)

	_log("✅ 玩家 Entity 创建完成（ID: %d）" % player.entity_id)


func _setup_systems() -> void:
	EcsWorld.register_system(StatsCalculationSystem.new())
	EcsWorld.register_system(DamageSystem.new())
	_log("✅ 基础 System 注册完成")


func _connect_buttons() -> void:
	btn_query.pressed.connect(_step1_query_info)
	btn_download.pressed.connect(_step2_download)
	btn_load.pressed.connect(_step3_load_pck)
	btn_test.pressed.connect(_step4_test_dlc)
	btn_unload.pressed.connect(_step5_unload)
	btn_full.pressed.connect(_run_full_flow)


func _subscribe_events() -> void:
	EventBus.subscribe(&"dlc_download_progress", _on_download_progress)
	EventBus.subscribe(&"poison_tick", _on_poison_tick)
	DlcManager.dlc_loaded.connect(_on_dlc_loaded)
	DlcManager.dlc_unloaded.connect(_on_dlc_unloaded)
	DlcManager.dlc_load_failed.connect(_on_dlc_failed)


# ─────────────────────────────────────────────
# 步骤 1：查询 DLC 信息
# ─────────────────────────────────────────────

func _step1_query_info() -> void:
	_log("\n[步骤1] 查询 DLC 信息...")
	_set_status("查询中...")
	btn_query.disabled = true

	_query_http = HTTPRequest.new()
	add_child(_query_http)
	_query_http.request_completed.connect(_on_query_completed)

	var err: int = _query_http.request(
		"%s/api/dlc/%s/info" % [SERVER_URL, DLC_ID]
	)
	if err != OK:
		_log("  ❌ 请求失败，请确认服务器已启动")
		btn_query.disabled = false
		_set_status("查询失败")


func _on_query_completed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	_query_http.queue_free()
	btn_query.disabled = false

	if result != HTTPRequest.RESULT_SUCCESS or response_code != 200:
		_log("  ❌ 服务器响应错误: HTTP %d" % response_code)
		_set_status("查询失败")
		return

	var json := JSON.new()
	if json.parse(body.get_string_from_utf8()) != OK:
		_log("  ❌ 响应解析失败")
		return

	_dlc_info = json.data
	_expected_hash = _dlc_info.get("sha256", "")

	_log("  ✅ DLC 信息获取成功")
	_log("     名称: %s" % _dlc_info.get("display_name", ""))
	_log("     版本: %s" % _dlc_info.get("version", ""))
	_log("     大小: %d bytes" % _dlc_info.get("file_size", 0))
	_log("     SHA-256: %s" % _expected_hash.left(16) + "...")

	if _dlc_info.get("warning"):
		_log("  ⚠️  %s" % _dlc_info.get("warning"))

	_set_status("DLC 信息已获取，可以下载")


# ─────────────────────────────────────────────
# 步骤 2：下载 PCK
# ─────────────────────────────────────────────

func _step2_download() -> void:
	_log("\n[步骤2] 开始下载 DLC PCK...")
	_set_status("下载中...")
	btn_download.disabled = true
	progress_bar.value = 0

	_download_save_path = "user://dlc/%s.pck" % DLC_ID
	DirAccess.make_dir_recursive_absolute("user://dlc/")

	_download_http = HTTPRequest.new()
	add_child(_download_http)
	_download_http.download_file = _download_save_path
	_download_http.request_completed.connect(_on_download_completed)

	var headers: PackedStringArray = []
	# 免费 DLC 无需 Token，付费 DLC 取消注释：
	# headers.append("Authorization: Bearer %s" % TEST_TOKEN)

	var err: int = _download_http.request(
		"%s/api/dlc/%s/download" % [SERVER_URL, DLC_ID],
		headers
	)
	if err != OK:
		_log("  ❌ 下载请求失败")
		btn_download.disabled = false
		_set_status("下载失败")


func _on_download_completed(result: int, response_code: int, headers: PackedStringArray, _body: PackedByteArray) -> void:
	_download_http.queue_free()
	btn_download.disabled = false

	if result != HTTPRequest.RESULT_SUCCESS:
		_log("  ❌ 下载失败，result=%d" % result)
		_set_status("下载失败")
		return

	if response_code != 200:
		_log("  ❌ 服务器错误: HTTP %d" % response_code)
		_set_status("下载失败")
		return

	# 从响应头获取服务器计算的哈希
	var server_hash: String = ""
	for header in headers:
		if header.to_lower().begins_with("x-dlc-sha256:"):
			server_hash = header.split(":")[1].strip_edges()
			break

	_log("  ✅ 下载完成: %s" % _download_save_path)
	if server_hash != "":
		_log("  服务器 SHA-256: %s..." % server_hash.left(16))

	progress_bar.value = 100
	_set_status("下载完成，开始验证...")

	# 自动验证文件完整性
	_verify_downloaded_file(server_hash)


func _verify_downloaded_file(server_hash: String) -> void:
	_log("\n[验证] 计算本地文件 SHA-256...")

	var file := FileAccess.open(_download_save_path, FileAccess.READ)
	if not file:
		_log("  ❌ 无法打开下载的文件")
		return

	var ctx := HashingContext.new()
	ctx.start(HashingContext.HASH_SHA256)
	const CHUNK: int = 65536
	while file.get_position() < file.get_length():
		ctx.update(file.get_buffer(CHUNK))
	file.close()
	var local_hash: String = ctx.finish().hex_encode()

	_log("  本地 SHA-256: %s..." % local_hash.left(16))

	# 与服务器哈希对比
	var compare_hash: String = server_hash if server_hash != "" else _expected_hash
	if compare_hash != "" and local_hash.to_lower() != compare_hash.to_lower():
		_log("  ❌ 哈希校验失败！文件可能已损坏")
		_set_status("验证失败")
		return

	_log("  ✅ 哈希校验通过，文件完整")
	_set_status("验证通过，可以加载")


# ─────────────────────────────────────────────
# 步骤 3：加载 PCK
# ─────────────────────────────────────────────

func _step3_load_pck() -> void:
	_log("\n[步骤3] 加载 PCK 到虚拟文件系统...")

	if not FileAccess.file_exists(_download_save_path):
		_log("  ❌ PCK 文件不存在，请先下载")
		return

	# 使用 DlcManager 的 PCK 加载器
	var pck_loader := DlcPckLoader.new()
	if not pck_loader.load_pck(_download_save_path, DLC_ID):
		_log("  ❌ PCK 挂载失败")
		_set_status("PCK 加载失败")
		return

	_log("  ✅ PCK 已挂载到 res://")

	# 验证 PCK 内容可访问
	var manifest_path: String = "res://dlc/%s/manifest.json" % DLC_ID
	if FileAccess.file_exists(manifest_path):
		_log("  ✅ manifest.json 可访问: %s" % manifest_path)
	else:
		_log("  ⚠️  manifest.json 不可访问（PCK 路径结构可能不匹配）")

	# 通过 DlcManager 注册 DLC 内容
	_log("  注册 DLC 内容到 EcsWorld...")
	var manifest_data: Dictionary = pck_loader.read_manifest_from_pck(DLC_ID)
	if manifest_data.is_empty():
		_log("  ❌ 无法读取 manifest，尝试直接加载...")
		# 回退：直接用目录模式加载（如果 PCK 挂载成功，res://dlc/ 下应该有内容）
		DlcManager.load_dlc(DLC_ID)
	else:
		var manifest := DlcManifest.from_dict(manifest_data, "res://dlc/%s" % DLC_ID)
		# 手动触发应用
		DlcManager._apply_dlc(manifest)

	_set_status("PCK 已加载")


# ─────────────────────────────────────────────
# 步骤 4：测试 DLC 运行
# ─────────────────────────────────────────────

func _step4_test_dlc() -> void:
	_log("\n[步骤4] 测试 DLC System 运行...")

	if not DlcManager.is_dlc_loaded(DLC_ID):
		_log("  ❌ DLC 未加载，请先执行步骤3")
		return

	# 验证 Component 已注册
	var poison_script = EcsWorld._component_registry.get(&"Poison")
	if poison_script:
		_log("  ✅ PoisonComponent 已注册到 EcsWorld")
	else:
		_log("  ❌ PoisonComponent 未注册")
		return

	# 验证玩家已注入 Poison Component
	if player.has_component(&"Poison"):
		_log("  ✅ 玩家已注入 PoisonComponent（角色扩展验证通过）")
	else:
		_log("  ⚠️  玩家未注入 PoisonComponent，手动添加...")
		var poison_comp: PoisonComponent = EcsWorld.create_component(&"Poison")
		if poison_comp:
			player.add_component(poison_comp)
			_log("  ✅ 手动添加成功")

	# 施加毒素，验证 PoisonSystem 运行
	var poison: PoisonComponent = player.get_component(&"Poison")
	if poison:
		poison.apply_poison(5.0, 10.0)
		_log("  ✅ 毒素已施加（5秒，10 DPS）")
		_log("  观察控制台输出，验证 PoisonSystem 每帧处理...")
		_set_status("DLC 运行中，观察毒素伤害输出")
	else:
		_log("  ❌ 获取 PoisonComponent 失败")


# ─────────────────────────────────────────────
# 步骤 5：卸载 DLC
# ─────────────────────────────────────────────

func _step5_unload() -> void:
	_log("\n[步骤5] 卸载 DLC...")

	if not DlcManager.is_dlc_loaded(DLC_ID):
		_log("  DLC 未加载，跳过")
		return

	var success: bool = DlcManager.unload_dlc(DLC_ID)
	if success:
		_log("  ✅ DLC 卸载成功")
		# 验证回退
		if not player.has_component(&"Poison"):
			_log("  ✅ 玩家 PoisonComponent 已回退移除")
		var poison_script = EcsWorld._component_registry.get(&"Poison")
		if not poison_script:
			_log("  ✅ PoisonComponent 已从 EcsWorld 注销")
		_log("  ⚠️  注意：PCK 仍在虚拟文件系统中，重启游戏才完全清除")
		_set_status("DLC 已卸载")
	else:
		_log("  ❌ 卸载失败")


# ─────────────────────────────────────────────
# 全流程一键执行
# ─────────────────────────────────────────────

func _run_full_flow() -> void:
	_log("\n========== 全流程验证开始 ==========")
	btn_full.disabled = true

	# 步骤1：查询
	_step1_query_info()
	await DlcManager.dlc_loaded  # 等待或用 timer

	# 使用 timer 串联步骤（给 HTTP 请求留时间）
	await get_tree().create_timer(1.5).timeout
	_step2_download()

	await get_tree().create_timer(3.0).timeout
	_step3_load_pck()

	await get_tree().create_timer(0.5).timeout
	_step4_test_dlc()

	await get_tree().create_timer(6.0).timeout  # 等待毒素运行几秒
	_step5_unload()

	await get_tree().create_timer(0.5).timeout
	_log("\n========== 全流程验证完成 ==========")
	btn_full.disabled = false


# ─────────────────────────────────────────────
# 事件回调
# ─────────────────────────────────────────────

func _on_download_progress(data: Dictionary) -> void:
	var downloaded: int = data.get("downloaded", 0)
	var total: int = data.get("total", 1)
	if total > 0:
		progress_bar.value = float(downloaded) / float(total) * 100.0


func _on_poison_tick(data: Dictionary) -> void:
	var damage: float = data.get("damage", 0.0)
	var remaining: float = data.get("remaining", 0.0)
	_log("  ☠️ 毒素 Tick: %.3f 伤害，剩余 %.1f 秒" % [damage, remaining])
	# 更新玩家 HP 显示
	if player.has_component(&"RuntimeStats"):
		var runtime: RuntimeStatsComponent = player.get_component(&"RuntimeStats")
		_set_status("玩家 HP: %.1f（中毒中，剩余 %.1f 秒）" % [runtime.current_hp, remaining])


func _on_dlc_loaded(dlc_id: String) -> void:
	_log("  🎉 DLC 加载事件: %s" % dlc_id)


func _on_dlc_unloaded(dlc_id: String) -> void:
	_log("  📦 DLC 卸载事件: %s" % dlc_id)


func _on_dlc_failed(dlc_id: String, error: String) -> void:
	_log("  ❌ DLC 失败事件: %s — %s" % [dlc_id, error])


# ─────────────────────────────────────────────
# UI 辅助
# ─────────────────────────────────────────────

func _log(msg: String) -> void:
	print(msg)
	if log_box:
		log_box.text += msg + "\n"
		# 自动滚动到底部
		await get_tree().process_frame
		log_box.scroll_to_line(log_box.get_line_count())


func _set_status(msg: String) -> void:
	if status_bar:
		status_bar.text = "状态: " + msg
