class_name ObjectPoolClass extends Node
## 全局对象池管理器
##
## 避免高频创建/销毁节点导致的 GC 压力。
## 适用于：子弹、伤害数字、粒子特效、掉落物。

## 池数据结构
class PoolData:
	var scene: PackedScene
	var available: Array[Node] = []
	var in_use: Array[Node] = []
	var max_size: int = 50
	var auto_expand: bool = true
	var parent: Node = null

## 池注册表 { pool_name: PoolData }
var _pools: Dictionary = {}


## 注册一个对象池
func register_pool(pool_name: StringName, scene: PackedScene,
					initial_size: int = 10, max_size: int = 50,
					parent: Node = null) -> void:
	var pool := PoolData.new()
	pool.scene = scene
	pool.max_size = max_size
	pool.parent = parent if parent else self
	_pools[pool_name] = pool
	
	# 预创建实例
	for i in initial_size:
		var instance: Node = scene.instantiate()
		instance.set_process(false)
		instance.set_physics_process(false)
		if instance is CanvasItem:
			instance.visible = false
		elif instance is Node3D:
			(instance as Node3D).visible = false
		pool.parent.add_child(instance)
		pool.available.append(instance)


## 从池中获取一个对象
func acquire(pool_name: StringName) -> Node:
	if pool_name not in _pools:
		push_error("对象池不存在: %s" % pool_name)
		return null
	
	var pool: PoolData = _pools[pool_name]
	var instance: Node
	
	if pool.available.size() > 0:
		instance = pool.available.pop_back()
	elif pool.auto_expand and pool.in_use.size() < pool.max_size:
		instance = pool.scene.instantiate()
		pool.parent.add_child(instance)
	else:
		# 池已满，回收最早使用的
		if pool.in_use.size() > 0:
			instance = pool.in_use.pop_front()
			_reset_instance(instance)
		else:
			push_warning("对象池 %s 已满且无可回收对象" % pool_name)
			return null
	
	# 激活实例
	instance.set_process(true)
	instance.set_physics_process(true)
	if instance is CanvasItem:
		instance.visible = true
	elif instance is Node3D:
		(instance as Node3D).visible = true
	pool.in_use.append(instance)
	
	if instance.has_method("on_pool_acquire"):
		instance.on_pool_acquire()
	
	return instance


## 归还对象到池中
func release(pool_name: StringName, instance: Node) -> void:
	if pool_name not in _pools:
		return
	var pool: PoolData = _pools[pool_name]
	pool.in_use.erase(instance)
	_reset_instance(instance)
	pool.available.append(instance)


## 获取池的使用统计
func get_pool_stats(pool_name: StringName) -> Dictionary:
	if pool_name not in _pools:
		return {}
	var pool: PoolData = _pools[pool_name]
	return {
		"available": pool.available.size(),
		"in_use": pool.in_use.size(),
		"max_size": pool.max_size,
	}


## 重置实例状态
func _reset_instance(instance: Node) -> void:
	instance.set_process(false)
	instance.set_physics_process(false)
	if instance is CanvasItem:
		instance.visible = false
	elif instance is Node3D:
		(instance as Node3D).visible = false
	if instance.has_method("on_pool_release"):
		instance.on_pool_release()
