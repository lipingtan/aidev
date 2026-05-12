class_name DlcAssetIntegration extends RefCounted
## DLC 与资产注册表的集成层
##
## 负责将 DLC 包中的数据资源注册到 DataManager，
## 并在卸载时清理。

## 注册 DLC 数据资源到 DataManager
## data_entries 格式：[{ "table": "items", "path": "data/items/sword.tres" }]
static func register_data(data_entries: Array, base_path: String, dlc_id: String) -> Array[Dictionary]:
	var registered: Array[Dictionary] = []
	for entry in data_entries:
		if not entry is Dictionary:
			continue
		var table: StringName = StringName(entry.get("table", ""))
		var rel_path: String = entry.get("path", "")
		if table == &"" or rel_path == "":
			continue
		
		var full_path: String = base_path + "/" + rel_path
		if not ResourceLoader.exists(full_path):
			push_warning("DLC 数据文件不存在: %s" % full_path)
			continue
		
		var res: Resource = load(full_path)
		if not res:
			push_warning("DLC 数据加载失败: %s" % full_path)
			continue
		
		if not "id" in res:
			push_warning("DLC 数据资源缺少 id 字段: %s" % full_path)
			continue
		
		# 检查 ID 冲突
		var existing: Resource = DataManager.get_data(table, res.id)
		if existing:
			push_warning("DLC 数据 ID 冲突: table=%s, id=%s, dlc=%s" % [table, res.id, dlc_id])
			continue
		
		DataManager.add_data(table, res.id, res)
		registered.append({"table": table, "id": res.id})
	
	return registered


## 批量注册目录下的所有数据资源
## dir_entries 格式：[{ "table": "items", "dir": "data/items/" }]
static func register_data_dirs(dir_entries: Array, base_path: String, dlc_id: String) -> Array[Dictionary]:
	var registered: Array[Dictionary] = []
	for entry in dir_entries:
		if not entry is Dictionary:
			continue
		var table: StringName = StringName(entry.get("table", ""))
		var rel_dir: String = entry.get("dir", "")
		if table == &"" or rel_dir == "":
			continue
		
		var full_dir: String = base_path + "/" + rel_dir
		var dir := DirAccess.open(full_dir)
		if not dir:
			continue
		
		dir.list_dir_begin()
		var file_name: String = dir.get_next()
		while file_name != "":
			if file_name.ends_with(".tres") or file_name.ends_with(".res"):
				var file_path: String = full_dir + file_name
				var res: Resource = load(file_path)
				if res and "id" in res:
					var existing: Resource = DataManager.get_data(table, res.id)
					if not existing:
						DataManager.add_data(table, res.id, res)
						registered.append({"table": table, "id": res.id})
					else:
						push_warning("DLC 数据 ID 冲突: table=%s, id=%s, dlc=%s" % [table, res.id, dlc_id])
			file_name = dir.get_next()
	
	return registered


## 注销 DLC 注册的所有数据
static func unregister_data(registered: Array[Dictionary]) -> void:
	for entry in registered:
		var table: StringName = entry.get("table", &"")
		var id: StringName = entry.get("id", &"")
		if table != &"" and id != &"":
			DataManager.remove_data(table, id)
