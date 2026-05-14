extends Node
## 存档管理器（Autoload）
##
## 负责游戏状态的序列化和反序列化。
## 支持多存档槽、自动存档、DLC 数据兼容。
##
## 存档文件存储在 user://saves/ 目录下。

## 存档完成时触发
signal save_completed(slot: int)

## 读档完成时触发
signal load_completed(slot: int)

## 存档失败时触发
signal save_failed(slot: int, error: String)

## 读档失败时触发
signal load_failed(slot: int, error: String)

const SAVE_DIR: String = "user://saves/"
const MAX_SLOTS: int = 10
const AUTO_SAVE_SLOT: int = 0

## 当前游戏内时间（秒）
var play_time: float = 0.0

## 是否正在计时
var _timing: bool = false


func _ready() -> void:
	# 确保存档目录存在
	DirAccess.make_dir_recursive_absolute(SAVE_DIR)


func _process(delta: float) -> void:
	if _timing:
		play_time += delta


## 开始计时
func start_timing() -> void:
	_timing = true


## 停止计时
func stop_timing() -> void:
	_timing = false


## 保存游戏到指定槽位
func save_game(slot: int, save_name: String = "") -> bool:
	if slot < 0 or slot >= MAX_SLOTS:
		save_failed.emit(slot, "无效的存档槽位: %d" % slot)
		return false
	
	var data := SaveData.new()
	data.save_id = "save_%d" % slot
	data.save_name = save_name if save_name != "" else "存档 %d" % slot
	data.timestamp = int(Time.get_unix_time_from_system())
	data.play_time = play_time
	
	# 收集各系统数据
	_collect_character_data(data)
	_collect_inventory_data(data)
	_collect_skill_data(data)
	_collect_world_state(data)
	_collect_storage_data(data)
	_collect_dlc_data(data)
	
	# 写入文件
	var path: String = SAVE_DIR + "slot_%d.tres" % slot
	var err: int = ResourceSaver.save(data, path)
	if err != OK:
		save_failed.emit(slot, "写入存档文件失败，错误码: %d" % err)
		return false
	
	save_completed.emit(slot)
	return true


## 从指定槽位读取游戏
func load_game(slot: int) -> bool:
	var path: String = SAVE_DIR + "slot_%d.tres" % slot
	if not FileAccess.file_exists(path):
		load_failed.emit(slot, "存档文件不存在: %s" % path)
		return false
	
	var data: SaveData = ResourceLoader.load(path) as SaveData
	if not data:
		load_failed.emit(slot, "存档文件损坏或格式不兼容")
		return false
	
	# 验证 DLC 兼容性
	_check_dlc_compatibility(data)
	
	# 恢复各系统数据
	play_time = data.play_time
	_restore_character_data(data)
	_restore_inventory_data(data)
	_restore_skill_data(data)
	_restore_world_state(data)
	_restore_storage_data(data)
	_restore_dlc_data(data)
	
	load_completed.emit(slot)
	return true


## 删除指定槽位的存档
func delete_save(slot: int) -> bool:
	var path: String = SAVE_DIR + "slot_%d.tres" % slot
	if not FileAccess.file_exists(path):
		return false
	var err: int = DirAccess.remove_absolute(path)
	return err == OK


## 获取存档槽位信息列表
func get_save_slots() -> Array[Dictionary]:
	var slots: Array[Dictionary] = []
	for i in MAX_SLOTS:
		var path: String = SAVE_DIR + "slot_%d.tres" % i
		if FileAccess.file_exists(path):
			var data: SaveData = ResourceLoader.load(path) as SaveData
			if data:
				slots.append(data.get_display_info())
			else:
				slots.append({"save_id": "slot_%d" % i, "corrupted": true})
		else:
			slots.append({"save_id": "slot_%d" % i, "empty": true})
	return slots


## 检查存档是否存在
func has_save(slot: int) -> bool:
	return FileAccess.file_exists(SAVE_DIR + "slot_%d.tres" % slot)


## 自动存档
func auto_save() -> bool:
	return save_game(AUTO_SAVE_SLOT, "自动存档")


## 收集角色数据
func _collect_character_data(data: SaveData) -> void:
	for entity in EcsWorld.get_all_entities():
		if not entity.has_component(&"BaseStats"):
			continue
		# 使用节点名称作为持久化 key（比 entity_id 更稳定）
		var key: String = entity.name
		var entity_data: Dictionary = {}
		# 基础属性
		var base: BaseStatsComponent = entity.get_component(&"BaseStats")
		entity_data["base_stats"] = base.serialize()
		# 经验/等级
		if entity.has_component(&"Experience"):
			var exp_comp: ExperienceComponent = entity.get_component(&"Experience")
			entity_data["experience"] = exp_comp.serialize()
		# 运行时状态
		if entity.has_component(&"RuntimeStats"):
			var runtime: RuntimeStatsComponent = entity.get_component(&"RuntimeStats")
			entity_data["runtime_stats"] = runtime.serialize()
		data.character_data[key] = entity_data


## 收集背包/装备数据
func _collect_inventory_data(data: SaveData) -> void:
	for entity in EcsWorld.get_all_entities():
		var entity_data: Dictionary = {}
		if entity.has_component(&"ItemContainer"):
			var container: ItemContainerComponent = entity.get_component(&"ItemContainer")
			entity_data["container"] = container.serialize()
		if entity.has_component(&"Equipment"):
			var equip: EquipmentComponent = entity.get_component(&"Equipment")
			entity_data["equipment"] = equip.serialize()
		if entity_data.size() > 0:
			data.inventory_data[str(entity.entity_id)] = entity_data


## 收集技能数据
func _collect_skill_data(data: SaveData) -> void:
	for entity in EcsWorld.get_all_entities():
		if not entity.has_component(&"SkillSet"):
			continue
		var skills: SkillSetComponent = entity.get_component(&"SkillSet")
		data.skill_data[str(entity.entity_id)] = skills.serialize()


## 收集世界状态
func _collect_world_state(data: SaveData) -> void:
	# 由具体游戏逻辑扩展
	pass


## 收集储物柜数据
func _collect_storage_data(data: SaveData) -> void:
	for entity in EcsWorld.get_all_entities():
		if not entity.has_component(&"ItemContainer"):
			continue
		var container: ItemContainerComponent = entity.get_component(&"ItemContainer")
		if container.container_type == ItemContainerComponent.ContainerType.STORAGE:
			data.storage_data[container.container_id] = container.serialize()


## 收集 DLC 数据
func _collect_dlc_data(data: SaveData) -> void:
	if DlcManager:
		data.loaded_dlcs = DlcManager.get_loaded_dlcs()


## 恢复角色数据
func _restore_character_data(data: SaveData) -> void:
	for entity in EcsWorld.get_all_entities():
		var key: String = entity.name
		if key not in data.character_data:
			continue
		var entity_data: Dictionary = data.character_data[key]
		if "base_stats" in entity_data and entity.has_component(&"BaseStats"):
			entity.get_component(&"BaseStats").deserialize(entity_data["base_stats"])
		if "experience" in entity_data and entity.has_component(&"Experience"):
			entity.get_component(&"Experience").deserialize(entity_data["experience"])
		if "runtime_stats" in entity_data and entity.has_component(&"RuntimeStats"):
			entity.get_component(&"RuntimeStats").deserialize(entity_data["runtime_stats"])
		# 标记属性需要重算
		if entity.has_component(&"FinalStats"):
			(entity.get_component(&"FinalStats") as FinalStatsComponent).mark_dirty()


## 恢复背包/装备数据
func _restore_inventory_data(data: SaveData) -> void:
	for entity in EcsWorld.get_all_entities():
		var key: String = str(entity.entity_id)
		if key not in data.inventory_data:
			continue
		var entity_data: Dictionary = data.inventory_data[key]
		if "container" in entity_data and entity.has_component(&"ItemContainer"):
			entity.get_component(&"ItemContainer").deserialize(entity_data["container"])
		if "equipment" in entity_data and entity.has_component(&"Equipment"):
			entity.get_component(&"Equipment").deserialize(entity_data["equipment"])


## 恢复技能数据
func _restore_skill_data(data: SaveData) -> void:
	for entity in EcsWorld.get_all_entities():
		var key: String = str(entity.entity_id)
		if key not in data.skill_data:
			continue
		if entity.has_component(&"SkillSet"):
			entity.get_component(&"SkillSet").deserialize(data.skill_data[key])


## 恢复世界状态
func _restore_world_state(_data: SaveData) -> void:
	pass


## 恢复储物柜数据
func _restore_storage_data(data: SaveData) -> void:
	for entity in EcsWorld.get_all_entities():
		if not entity.has_component(&"ItemContainer"):
			continue
		var container: ItemContainerComponent = entity.get_component(&"ItemContainer")
		if container.container_type == ItemContainerComponent.ContainerType.STORAGE:
			var key: StringName = container.container_id
			if key in data.storage_data:
				container.deserialize(data.storage_data[key])


## 恢复 DLC 数据
func _restore_dlc_data(_data: SaveData) -> void:
	pass


## 检查 DLC 兼容性
func _check_dlc_compatibility(data: SaveData) -> void:
	if not DlcManager:
		return
	var current_dlcs: Array[String] = DlcManager.get_loaded_dlcs()
	for dlc_id in data.loaded_dlcs:
		if dlc_id not in current_dlcs:
			push_warning("存档使用了当前未加载的 DLC: %s，相关数据将被忽略" % dlc_id)
