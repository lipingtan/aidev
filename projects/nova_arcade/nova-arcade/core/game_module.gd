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
## 两个引擎坑（GUI 截图实测）：
## 1. CanvasLayer 不随祖先 Control.visible 隐藏 → 逐层显式处理
## 2. 模块根是普通 Node（非 CanvasItem），GameHost.visible 的传播在此断链，
##    下属 Node2D（棋盘等）即使 GameHost 已隐藏仍会渲染 → 必须逐 CanvasItem 快照+隐藏
## 恢复时只点亮隐藏前可见的项，避免复活游戏内本就隐藏的实体（已拾取道具/已击败怪物）
var _vis_snapshot: Array = []

func set_module_visible(v: bool) -> void:
	if v:
		for it in _vis_snapshot:
			if is_instance_valid(it):
				(it as CanvasItem).visible = true
		_vis_snapshot.clear()
	else:
		_vis_snapshot.clear()
		_snap_and_hide(self)
	_set_canvas_layers_visible(self, v)

func _snap_and_hide(n: Node) -> void:
	for ch in n.get_children():
		if ch is CanvasItem and (ch as CanvasItem).visible:
			_vis_snapshot.append(ch)
			(ch as CanvasItem).visible = false
		_snap_and_hide(ch)

func _set_canvas_layers_visible(n: Node, v: bool) -> void:
	if n is CanvasLayer:
		(n as CanvasLayer).visible = v
	for ch in n.get_children():
		_set_canvas_layers_visible(ch, v)
