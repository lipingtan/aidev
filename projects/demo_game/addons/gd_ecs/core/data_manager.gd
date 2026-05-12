class_name DataManagerClass extends Node
## 全局数据表管理器
##
## 统一管理所有策划配置数据（物品表、技能表、怪物表等），
## 提供按 ID 查询、条件查询和热重载能力。

## 已加载的数据表 { table_name: { id: Resource } }
var _tables: Dictionary = {}

## 数据表路径配置 { table_name: directory_path }
var _table_paths: Dictionary = {}


## 注册数据表路径
func register_table(table_name: StringName, dir_path: String) -> void:
	_table_paths[table_name] = dir_path
	_load_table(table_name)


## 加载单张表
func _load_table(table_name: StringName) -> void:
	var dir_path: String = _table_paths.get(table_name, "")
	if dir_path.is_empty():
		return
	
	_tables[table_name] = {}
	var dir := DirAccess.open(dir_path)
	if not dir:
		push_warning("数据表目录不存在: %s" % dir_path)
		return
	
	dir.list_dir_begin()
	var file_name: String = dir.get_next()
	while file_name != "":
		if file_name.ends_with(".tres") or file_name.ends_with(".res"):
			var res: Resource = ResourceLoader.load(dir_path + "/" + file_name)
			if res and "id" in res:
				_tables[table_name][res.id] = res
		file_name = dir.get_next()


## 按 ID 获取数据
func get_data(table_name: StringName, id: StringName) -> Resource:
	var table: Dictionary = _tables.get(table_name, {})
	return table.get(id)


## 获取整张表
func get_table(table_name: StringName) -> Dictionary:
	return _tables.get(table_name, {})


## 条件查询
func query(table_name: StringName, filter: Callable) -> Array[Resource]:
	var results: Array[Resource] = []
	var table: Dictionary = _tables.get(table_name, {})
	for item in table.values():
		if filter.call(item):
			results.append(item)
	return results


## 热重载单张表（开发期调试用）
func reload_table(table_name: StringName) -> void:
	_load_table(table_name)


## 向表中动态添加数据（DLC 用）
func add_data(table_name: StringName, id: StringName, data: Resource) -> void:
	if table_name not in _tables:
		_tables[table_name] = {}
	_tables[table_name][id] = data


## 从表中移除数据（DLC 卸载用）
func remove_data(table_name: StringName, id: StringName) -> void:
	if table_name in _tables:
		_tables[table_name].erase(id)


## 获取表中数据数量
func get_count(table_name: StringName) -> int:
	return _tables.get(table_name, {}).size()
