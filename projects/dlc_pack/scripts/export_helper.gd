extends Node
## DLC 导出辅助脚本
##
## 在 Godot 编辑器中运行此场景，自动将 dlc/ 目录下的内容
## 导出为 .pck 文件，输出到 exported/ 目录。
##
## 使用方式：
## 1. 在 Godot 4.5 中打开 dlc_pack 工程
## 2. 运行此场景（F5）
## 3. 查看 exported/ 目录下生成的 .pck 文件
## 4. 将 .pck 文件复制到 dlc_server/dlc_files/ 目录

@onready var status_label: Label = $StatusLabel
@onready var export_btn: Button = $ExportBtn
@onready var hash_label: Label = $HashLabel


func _ready() -> void:
	export_btn.pressed.connect(_on_export_pressed)
	status_label.text = "准备就绪。点击「导出 DLC PCK」开始。"


func _on_export_pressed() -> void:
	status_label.text = "正在导出..."
	export_btn.disabled = true

	# 确保输出目录存在
	DirAccess.make_dir_recursive_absolute("res://exported/")

	# 导出 poison_dlc 为 PCK
	var output_path: String = "res://exported/poison_dlc.pck"
	var files_to_pack: PackedStringArray = _collect_files("res://dlc/poison_dlc/")

	if files_to_pack.is_empty():
		status_label.text = "❌ 未找到 dlc/poison_dlc/ 目录下的文件"
		export_btn.disabled = false
		return

	# 使用 PCKPacker 打包
	var packer := PCKPacker.new()
	var err: int = packer.pck_start(output_path)
	if err != OK:
		status_label.text = "❌ PCK 创建失败，错误码: %d" % err
		export_btn.disabled = false
		return

	for file_path in files_to_pack:
		# 保持 res:// 路径结构
		packer.add_file(file_path, file_path)

	err = packer.flush(true)
	if err != OK:
		status_label.text = "❌ PCK 写入失败，错误码: %d" % err
		export_btn.disabled = false
		return

	# 计算 SHA-256 哈希（供服务器配置使用）
	var hash_str: String = _compute_sha256(output_path)

	status_label.text = "✅ 导出成功！\n路径: %s\n文件数: %d" % [output_path, files_to_pack.size()]
	hash_label.text = "SHA-256: %s\n\n请将此哈希值填入 dlc_server/config.json 的 hash 字段" % hash_str
	export_btn.disabled = false

	print("=== DLC PCK 导出完成 ===")
	print("输出路径: ", output_path)
	print("SHA-256: ", hash_str)
	print("包含文件:")
	for f in files_to_pack:
		print("  ", f)


## 递归收集目录下所有文件
func _collect_files(dir_path: String) -> PackedStringArray:
	var result: PackedStringArray = []
	var dir := DirAccess.open(dir_path)
	if not dir:
		return result

	dir.list_dir_begin()
	var name: String = dir.get_next()
	while name != "":
		var full_path: String = dir_path + name
		if dir.current_is_dir():
			if name != "." and name != "..":
				result.append_array(_collect_files(full_path + "/"))
		else:
			result.append(full_path)
		name = dir.get_next()

	return result


## 计算文件 SHA-256
func _compute_sha256(file_path: String) -> String:
	var file := FileAccess.open(file_path, FileAccess.READ)
	if not file:
		return ""
	var ctx := HashingContext.new()
	ctx.start(HashingContext.HASH_SHA256)
	const CHUNK: int = 65536
	while file.get_position() < file.get_length():
		ctx.update(file.get_buffer(CHUNK))
	file.close()
	return ctx.finish().hex_encode()
