class_name DemoTagComponent extends EcsComponent
## Demo 标签 Component
##
## 用于标记 Entity 的角色类型，方便 System 区分玩家和敌人。

## 标签枚举
enum Tag {
	PLAYER,  ## 玩家
	ENEMY,   ## 敌人
	CHEST,   ## 宝箱
}

## 当前标签
@export var tag: Tag = Tag.PLAYER

## Entity 显示名称（用于 UI 和日志）
@export var display_name: String = "未命名"


func get_component_name() -> StringName:
	return &"DemoTag"
