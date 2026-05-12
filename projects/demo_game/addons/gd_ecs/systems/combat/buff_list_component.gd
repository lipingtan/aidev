class_name BuffListComponent extends EcsComponent
## Buff 列表 Component
##
## 管理角色当前所有激活的 Buff/Debuff 实例。

## Buff 施加时触发
signal buff_applied(buff_id: StringName)

## Buff 移除时触发
signal buff_removed(buff_id: StringName, reason: StringName)

## Buff 实例（运行时状态）
class BuffInstance:
	var buff_data: BuffData
	var remaining_time: float = 0.0
	var stack_count: int = 1
	var tick_timer: float = 0.0
	var source_entity_id: int = -1

## 当前激活的 Buff 列表
var active_buffs: Array[BuffInstance] = []

## 免疫标签（有这些标签则免疫对应 Buff）
@export var immunity_tags: Array[StringName] = []


func get_component_name() -> StringName:
	return &"BuffList"


## 施加 Buff（返回是否成功）
func add_buff(buff: BuffData, source_id: int = -1) -> bool:
	# 免疫检查
	for tag in buff.immunity_tags:
		if tag in immunity_tags:
			return false
	
	# 查找已有同类 Buff
	var existing: BuffInstance = _find_buff(buff.id)
	
	if existing:
		match buff.stack_mode:
			BuffData.StackMode.REFRESH:
				existing.remaining_time = buff.duration
			BuffData.StackMode.STACK_COUNT:
				if existing.stack_count < buff.max_stacks:
					existing.stack_count += 1
				existing.remaining_time = buff.duration
			BuffData.StackMode.STACK_INDEPENDENT:
				# 独立实例，直接添加新的
				_create_instance(buff, source_id)
				buff_applied.emit(buff.id)
				return true
	else:
		_create_instance(buff, source_id)
	
	buff_applied.emit(buff.id)
	return true


## 移除指定 Buff
func remove_buff(buff_id: StringName, reason: StringName = &"manual") -> bool:
	for i in active_buffs.size():
		if active_buffs[i].buff_data.id == buff_id:
			active_buffs.remove_at(i)
			buff_removed.emit(buff_id, reason)
			return true
	return false


## 净化指定类型的 Buff
func dispel(dispel_type: BuffData.DispelType) -> int:
	var removed: int = 0
	var to_remove: Array[int] = []
	for i in active_buffs.size():
		var inst: BuffInstance = active_buffs[i]
		if inst.buff_data.dispel_type == dispel_type:
			to_remove.append(i)
	# 倒序移除避免索引错位
	for i in range(to_remove.size() - 1, -1, -1):
		var idx: int = to_remove[i]
		var buff_id: StringName = active_buffs[idx].buff_data.id
		active_buffs.remove_at(idx)
		buff_removed.emit(buff_id, &"dispelled")
		removed += 1
	return removed


## 检查是否有指定 Buff
func has_buff(buff_id: StringName) -> bool:
	return _find_buff(buff_id) != null


## 获取 Buff 的当前层数
func get_stack_count(buff_id: StringName) -> int:
	var inst: BuffInstance = _find_buff(buff_id)
	return inst.stack_count if inst else 0


## 获取所有 Buff 提供的属性修饰器
func get_all_modifiers() -> Array[StatModifier]:
	var result: Array[StatModifier] = []
	for inst in active_buffs:
		for mod in inst.buff_data.stat_modifiers:
			# 按层数缩放修饰值
			var scaled_mod: StatModifier = mod.duplicate()
			scaled_mod.value *= inst.stack_count
			result.append(scaled_mod)
	return result


## 创建 Buff 实例
func _create_instance(buff: BuffData, source_id: int) -> BuffInstance:
	var inst := BuffInstance.new()
	inst.buff_data = buff
	inst.remaining_time = buff.duration
	inst.source_entity_id = source_id
	active_buffs.append(inst)
	return inst


## 查找已有 Buff 实例
func _find_buff(buff_id: StringName) -> BuffInstance:
	for inst in active_buffs:
		if inst.buff_data.id == buff_id:
			return inst
	return null
