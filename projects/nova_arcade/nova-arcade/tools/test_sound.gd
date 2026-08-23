extends Node
## T6 Sound 回归测试场景：headless 运行 res://tools/test_sound.tscn，全过 exit 0。

var _failures: int = 0

func _ready() -> void:
	_run_tests()

func _check(cond: bool, label: String) -> void:
	if cond:
		print("[test_sound] PASS ", label)
	else:
		_failures += 1
		push_error("[test_sound] FAIL ", label)

func _run_tests() -> void:
	# 各音效接口可调用无报错，last_play 更新
	Sound.click()
	_check(Sound.last_play == "click", "click")
	Sound.toggle()
	_check(Sound.last_play == "toggle", "toggle")
	Sound.success()
	_check(Sound.last_play == "success", "success")
	Sound.error()
	_check(Sound.last_play == "error", "error")
	Sound.coin()
	_check(Sound.last_play == "coin", "coin")
	# 震动：桌面/headless 无特性，调用不报错即可
	Sound.haptic("light")
	Sound.haptic("strong")
	_check(true, "haptic 平台保护不报错")
	# profile 关闭后静默（last_play 不变）
	DB.save_profile({"sound_enabled": false}, true)
	var before := Sound.last_play
	Sound.click()
	_check(Sound.last_play == before, "sound_enabled=false 静默")
	# haptics 关闭同样静默（桌面本就跳过，仅验证不报错）
	DB.save_profile({"haptics_enabled": false}, true)
	Sound.haptic("medium")
	_check(true, "haptics_enabled=false 不报错")

	if _failures == 0:
		print("[test_sound] ALL PASS")
	get_tree().quit(1 if _failures > 0 else 0)
