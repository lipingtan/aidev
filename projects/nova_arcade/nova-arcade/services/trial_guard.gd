extends Node
## 试玩次数守卫（Autoload 名 TrialGuard，注册顺序第 7；CR-2 T5）
## - left(gid)：meta.trial.plays − record.trial_used；非 trial / minutes 型返回 -1 不抛错
## - consume(gid)：trial_used+1 强写（upsert_record 立即落盘，强杀重启不丢 RG-7）
##   + 发 EventBus.trial_consumed(gid, left)；非 trial 无操作（已拥有恒真不消耗）
##
## 依赖：Registry/GameMeta、DB（records）、EventBus（trial_consumed）

## 取剩余试玩次数；非 trial 游戏或 minutes 型限制返回 -1（调用方仅在 price_model=trial 时消费）
func left(gid: String) -> int:
	var m: GameMeta = Registry.lookup(gid)
	if m == null or m.price_model != "trial" or not _is_plays(m):
		return -1
	return _plays(m) - _used(gid)

## 消耗一次试玩：强写 trial_used+1 + 发 trial_consumed(gid, left)；非 trial 无操作
func consume(gid: String) -> void:
	var m: GameMeta = Registry.lookup(gid)
	if m == null or m.price_model != "trial" or not _is_plays(m):
		return
	DB.upsert_record(gid, {"trial_used": _used(gid) + 1})
	EventBus.trial_consumed.emit(gid, left(gid))

## 是否 plays 型试玩限制（minutes 型本 CR 不涉及，返回 false 不消耗）
func _is_plays(m: GameMeta) -> bool:
	return m.trial.has("plays")

## 试玩总次数（meta.trial.plays）
func _plays(m: GameMeta) -> int:
	return int(m.trial.get("plays", 0))

## 已消耗次数（record.trial_used；无记录为 0）
func _used(gid: String) -> int:
	var r: Variant = DB.get_record(gid)
	if not (r is Dictionary):
		return 0
	return int((r as Dictionary).get("trial_used", 0))
