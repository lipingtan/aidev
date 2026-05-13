@abstract
class_name EcsComponent extends Resource
## ECS Component 基类（抽象类）
##
## Component 是纯数据容器，不包含行为逻辑。
## 所有游戏数据（属性、状态、配置）都通过 Component 存储。
## 可通过编辑器 Inspector 面板直接编辑。
##
## 使用方式：
## 1. 创建子类继承 EcsComponent
## 2. 用 @export 定义数据字段
## 3. 覆盖 get_component_name() 返回唯一名称
## 4. 通过 EcsEntity.add_component() 挂载到 Entity

## 返回组件唯一名称（子类必须覆盖）
## @abstract 方法不能有函数体
@abstract
func get_component_name() -> StringName


## 序列化为字典（用于存档）
func serialize() -> Dictionary:
	var data: Dictionary = {}
	for prop in get_property_list():
		if prop.usage & PROPERTY_USAGE_STORAGE and prop.name != "resource_local_to_scene":
			data[prop.name] = get(prop.name)
	return data


## 从字典反序列化（用于读档）
func deserialize(data: Dictionary) -> void:
	for key in data:
		if key in self:
			set(key, data[key])
