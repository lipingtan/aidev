class_name SaveData extends Resource
## 存档数据结构
##
## 序列化游戏状态的完整快照，以 .tres 格式存储。
## 各系统通过 collect/restore 接口参与存档。

## 存档唯一标识
@export var save_id: String = ""

## 存档显示名称
@export var save_name: String = ""

## 存档时间戳（Unix 时间）
@export var timestamp: int = 0

## 游戏内累计时间（秒）
@export var play_time: float = 0.0

## 存档时所在场景 ID
@export var scene_id: StringName = &""

## 玩家位置
@export var player_position: Vector3 = Vector3.ZERO

## 玩家朝向
@export var player_rotation: float = 0.0

## 角色数据（属性、等级、经验）
@export var character_data: Dictionary = {}

## 背包/装备数据
@export var inventory_data: Dictionary = {}

## 技能数据
@export var skill_data: Dictionary = {}

## 任务进度数据
@export var quest_data: Dictionary = {}

## 世界状态（已开宝箱、已触发事件等）
@export var world_state: Dictionary = {}

## 储物柜数据
@export var storage_data: Dictionary = {}

## DLC 扩展数据（DLC 卸载后保留，重新加载时恢复）
@export var dlc_data: Dictionary = {}

## 已加载的 DLC 列表（用于读档时验证）
@export var loaded_dlcs: Array[String] = []


## 生成存档显示信息
func get_display_info() -> Dictionary:
	return {
		"save_id": save_id,
		"save_name": save_name,
		"timestamp": timestamp,
		"play_time": play_time,
		"scene_id": scene_id,
	}
