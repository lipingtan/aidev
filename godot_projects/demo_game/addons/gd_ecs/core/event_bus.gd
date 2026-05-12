class_name EventBusClass extends Node
## 全局事件总线
##
## 提供发布/订阅机制，让系统间通过事件名解耦通信。
## 支持优先级排序和调试历史记录。

## 事件监听器注册表 { event_name: Array[{ callback, priority }] }
var _listeners: Dictionary = {}

## 事件历史（调试用）
var _history: Array[Dictionary] = []
var history_enabled: bool = false
const MAX_HISTORY: int = 100


## 订阅事件
func subscribe(event_name: StringName, callback: Callable, priority: int = 0) -> void:
	if event_name not in _listeners:
		_listeners[event_name] = []
	_listeners[event_name].append({"callback": callback, "priority": priority})
	_listeners[event_name].sort_custom(func(a, b): return a.priority > b.priority)


## 取消订阅
func unsubscribe(event_name: StringName, callback: Callable) -> void:
	if event_name not in _listeners:
		return
	_listeners[event_name] = _listeners[event_name].filter(
		func(entry): return entry.callback != callback
	)


## 发布事件（立即执行）
func emit_event(event_name: StringName, data: Dictionary = {}) -> void:
	if history_enabled:
		_history.append({"event": event_name, "data": data, "time": Time.get_ticks_msec()})
		if _history.size() > MAX_HISTORY:
			_history.pop_front()
	
	if event_name not in _listeners:
		return
	for entry in _listeners[event_name]:
		entry.callback.call(data)


## 延迟发布事件（帧末执行，避免在遍历中修改状态）
func emit_deferred(event_name: StringName, data: Dictionary = {}) -> void:
	call_deferred("emit_event", event_name, data)


## 清除指定事件的所有监听器
func clear_event(event_name: StringName) -> void:
	_listeners.erase(event_name)


## 清除所有监听器
func clear_all() -> void:
	_listeners.clear()


## 获取事件历史（调试用）
func get_history() -> Array[Dictionary]:
	return _history
