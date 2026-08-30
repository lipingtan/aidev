class_name GameModule extends Node
## GameModule 协议基类（design §2）——所有游戏场景继承本类，Launcher 只依赖此基类
##
## - boot(ctx)：Launcher 注入 ctx={save_dir,trial_mode,owned,best,viewport_size} 后接管入口
## - quit_requested(result)：请求回盒；result 必含 {score, playtime(秒), achievements, extra}
## - achievement_unlocked(aid)：成就解锁（本 CR 恒空数组，M2 AchievementEngine 接入后启用）
## - pause_game / resume_game：Shell 暂停键转发到游戏自身暂停机制
##
## 约束：适配器只依赖本基类 + src 节点，禁止引用 Shell 类型（override §1）

## Launcher 注入的游戏上下文（save_dir/trial_mode/owned/best/viewport_size）
var ctx: Dictionary = {}

## 请求返回盒；result 含 score/playtime/achievements/extra
signal quit_requested(result: Dictionary)
## 成就解锁；aid 为成就 id
signal achievement_unlocked(aid: String)

## 启动：接收 ctx 并接管游戏入口（子类必须覆写）
func boot(ctx_in: Dictionary) -> void:
	ctx = ctx_in

## 恢复（子类转发到游戏自身恢复机制）
func resume_game() -> void:
	pass

## 模块整体显隐（退出/回壳时由 Launcher 调用）：
## CanvasLayer 不随祖先 Control.visible 隐藏，须显式遍历子树逐一处理。
## 默认实现：递归隐藏/恢复子树中所有 CanvasLayer（含各自子节点）；
## src 中普通 Node2D/Control 随 GameHost.visible 走，无需在此处理。
func set_module_visible(v: bool) -> void:
	_set_canvas_layers_visible(self, v)

func _set_canvas_layers_visible(n: Node, v: bool) -> void:
	if n is CanvasLayer:
		(n as CanvasLayer).visible = v
	for ch in n.get_children():
		_set_canvas_layers_visible(ch, v)
