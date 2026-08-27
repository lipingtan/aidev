class_name LauncherUtil extends Node
## Launcher 私有工具（CR-2 T4）：GameHost/Toast 定位 + DB 记录读写辅助
## 自 launcher.gd 拆出以守 ≤200 行/文件；无状态，由 Launcher add_child 后使用

## GameHost 节点（find_child 定位，与 Nav 定位 PageStack 同模式）
func host() -> Control:
	return get_tree().root.find_child("GameHost", true, false) as Control

## 游戏态 UI 切换（on=进入游戏）：GameHost 显隐与 Shell App 相反——
## 运行中 GameHost 可见 + 整壳隐藏防叠加；回 Shell 后反之
func set_game_ui(on: bool) -> void:
	var h := host()
	if h != null:
		h.visible = on
	var app := get_tree().root.find_child("App", true, false) as Control
	if app != null:
		app.visible = not on

## Toast（经 find_child 找 UI 层）
func toast(msg: String) -> void:
	var layer := get_tree().root.find_child("ToastLayer", true, false)
	if layer != null:
		layer.call("show_msg", msg)

## session+1 + last_played（读当前值，强写）
func bump_session(gid: String) -> void:
	DB.upsert_record(gid, {
		"last_played": int(Time.get_unix_time_from_system()),
		"sessions": int_field(gid, "sessions") + 1,
	})

## 强写本局结果：total_playtime 累加 / best 取大 / finish_count+1（读当前值，RG-7）
func finish_record(gid: String, result: Dictionary) -> void:
	DB.upsert_record(gid, {
		"total_playtime": int_field(gid, "total_playtime") + int(result.get("playtime", 0)),
		"best": maxi(int_field(gid, "best"), int(result.get("score", 0))),
		"finish_count": int_field(gid, "finish_count") + 1,
	})

## 读记录整数字段（缺失为 0）
func int_field(gid: String, key: String) -> int:
	var r: Variant = DB.get_record(gid)
	if not (r is Dictionary):
		return 0
	return int((r as Dictionary).get(key, 0))
