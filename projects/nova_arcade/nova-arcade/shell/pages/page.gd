@abstract
class_name Page extends Control
## 页面基类（四接口，CD §3.2）
##
## 所有盒子页面继承此基类；跳转只经 Nav、通知只经 EventBus，页面间互不引用。
## - on_enter：进入时由 Nav 调用，子类必须覆写（未覆写运行时 push_warning 提示）
## - on_exit / on_resume：离开 / 从下层恢复为当前页（默认空实现）
## - on_back：返回键拦截，return false → Nav 走默认 pop
##
## 引擎限制说明：Godot 4.5 无法解析「多方法类中的无函数体 @abstract 方法」（已实测），
## 故 on_enter 不用 @abstract，改为默认实现 + 运行时警告；类级 @abstract 保留（禁止直接实例化基类）。
##
## 依赖：
## - Nav: 驱动页面进出栈并调用本类接口（不直接引用，由 Nav 侧调用）
## - EventBus: 页面通知外部一律发信号

## 进入页面；data 为 Nav.push 传入的参数。子类必须覆写（未覆写时运行时警告）。
func on_enter(data: Dictionary) -> void:
	push_warning("Page.on_enter: 子类未覆写此方法（%s）" % (get_script() as String))

## 离开页面（pop / 切 Tab 清栈时调用），默认空实现。
func on_exit() -> void:
	pass

## 从下层页面恢复为当前页，默认空实现。
func on_resume() -> void:
	pass

## 返回键拦截；return true 表示已消费，return false → Nav 走默认 pop。
func on_back() -> bool:
	return false
