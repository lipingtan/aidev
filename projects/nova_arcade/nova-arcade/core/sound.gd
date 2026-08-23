extends Node
## UI 音效/震动单例（design.md §3.6）
##
## Autoload 名 Sound（注册顺序第 6），不声明 class_name（godot-engine §六）。
## - click()/toggle()/success()/error()/coin()：UI 音效；haptic(strength)：震动
## - 尊重 profile 开关 sound_enabled/haptics_enabled（缺省为开）
## - CR-1 用占位短音 assets/sfx/*.wav（tools/gen_sfx.gd 生成），M2 换正式音效
## - 震动受平台能力保护：非 mobile 平台静默跳过不报错（4.5 无内置 haptic API，M2 接插件）
##
## 依赖：DB（读 profile 开关）

## 流缓存：name -> AudioStream（首次播放时加载）
var _streams: Dictionary = {}
## 最近一次实际播放的音效名（测试/调试用；被开关拦截时不变）
var last_play: String = ""

func _ready() -> void:
	pass # 资源惰性加载，避免启动期无谓 IO

# === 音效接口 ===

## 点击音
func click() -> void:
	_play("click")

## 切换音（主题/开关类交互）
func toggle() -> void:
	_play("toggle")

## 成功音
func success() -> void:
	_play("success")

## 错误音
func error() -> void:
	_play("error")

## 金币/获得音
func coin() -> void:
	_play("coin")

## 震动；strength: "light"|"medium"|"strong"。
## 偏差（实测）：Godot 4.5 无内置跨平台 haptic API，CR-1 仅做能力保护（非 mobile 静默跳过），
## 真机震动待 M2 接入移动端插件后在 _haptic_impl 实现
func haptic(strength: String = "light") -> void:
	if not _enabled("haptics_enabled"):
		return
	if not OS.has_feature("mobile"):
		return
	_haptic_impl(strength)

## 真机震动实现位（M2 接插件后填实）
func _haptic_impl(strength: String) -> void:
	var amp := 0.3
	match strength:
		"medium":
			amp = 0.6
		"strong":
			amp = 1.0
	push_warning("Sound: haptic 待 M2 移动端插件接入（amp=%s）" % str(amp))

# === 底层 ===

## profile 开关（缺省为开）
func _enabled(key: String) -> bool:
	var p := DB.get_profile()
	return bool(p.get(key, true))

## 播放占位音效（受 sound_enabled 拦截；资源缺失仅告警）
func _play(name: String) -> void:
	if not _enabled("sound_enabled"):
		return
	var stream: AudioStream = _streams.get(name)
	if stream == null:
		stream = load("res://assets/sfx/%s.wav" % name) as AudioStream
		if stream == null:
			push_warning("Sound: 音效资源缺失 %s" % name)
			return
		_streams[name] = stream
	var player := AudioStreamPlayer.new()
	player.stream = stream
	add_child(player)
	player.play()
	player.finished.connect(player.queue_free)
	last_play = name
