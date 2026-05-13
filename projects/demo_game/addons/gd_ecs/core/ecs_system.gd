@abstract
class_name EcsSystem extends RefCounted
## ECS System 基类（抽象类）
##
## System 是独立的处理器，负责批量处理匹配特定 Component 组合的 Entity。
## 每个 System 声明其关注的 Component 组合（Query），
## 在每帧或事件触发时批量处理匹配的 Entity 集合。
##
## 使用方式：
## 1. 创建子类继承 EcsSystem
## 2. 覆盖 get_query() 返回关注的 Component 名称数组
## 3. 覆盖 process() 实现批量处理逻辑
## 4. 通过 EcsWorld.register_system() 注册

## 执行优先级（数值越小越先执行）
var priority: int = 0

## 执行阶段：&"process" 或 &"physics_process"
var phase: StringName = &"process"

## 是否启用
var enabled: bool = true

## System 名称（用于调试和日志）
var system_name: StringName = &""


## 返回此 System 关注的 Component 名称数组
## @abstract 方法不能有函数体
@abstract
func get_query() -> Array[StringName]


## 批量处理匹配的 Entity 集合
## @abstract 方法不能有函数体
@abstract
func process(entities: Array, delta: float) -> void


## System 被注册到 World 时调用（可选覆盖）
func on_registered() -> void:
	pass


## System 被注销时调用（可选覆盖）
func on_unregistered() -> void:
	pass
