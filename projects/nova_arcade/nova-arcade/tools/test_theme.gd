extends Node
## T7 主题系统回归测试场景：headless 运行 res://tools/test_theme.tscn，全过 exit 0。
## 视觉项（无布局跳变/重启恢复截图）移交 T10 在完整场景树后执行（Review P-5）。

var _failures: int = 0
## 信号捕获
var _settings_sig := ""
var _theme_sig := ""
var _bg: Control

func _ready() -> void:
	EventBus.settings_changed.connect(_on_settings)
	EventBus.theme_changed.connect(_on_theme)
	_run()

func _on_settings(key: String, value: Variant) -> void:
	_settings_sig = "%s=%s" % [key, str(value)]

func _on_theme(name: String) -> void:
	_theme_sig = name

func _check(cond: bool, label: String) -> void:
	if cond:
		print("[test_theme] PASS ", label)
	else:
		_failures += 1
		push_error("[test_theme] FAIL ", label)

func _run() -> void:
	# 1. token 双套差异 + icon_grad
	var bg_neon := ThemeTokens.color("bg")
	ThemeTokens.current = "elegant"
	var bg_el := ThemeTokens.color("bg")
	_check(bg_neon != bg_el, "token 双套差异")
	var g := ThemeTokens.icon_grad(Color(0.5, 0.8, 1))
	_check(g.colors.size() == 2 and g.colors[1] == ThemeTokens.color("icon_end"), "icon_grad 渐变")
	ThemeTokens.current = "neon"

	# 2. BgLayer 实例化（初始 neon：网格线可见/点阵隐藏）
	var scene := load("res://shell/bglayer.tscn") as PackedScene
	_bg = scene.instantiate()
	add_child(_bg)
	await get_tree().process_frame
	_check(_bg.get_node("GridNeon").visible and not _bg.get_node("DotGrid").visible, "初始 neon 纹理层")

	# 3. apply_theme 信号流：Token → root.theme → profile force 写 → settings_changed + theme_changed
	ThemeTokens.apply_theme("elegant", get_tree().root)
	_check(ThemeTokens.current == "elegant", "apply_theme Token 切换")
	_check(get_tree().root.theme != null, "root.theme 换资源")
	_check(str(DB.get_profile().get("theme", "")) == "elegant", "profile force 写")
	_check(_settings_sig == "theme=elegant", "settings_changed 信号")
	_check(_theme_sig == "elegant", "theme_changed 信号")

	# 4. BgLayer 仅响应 theme_changed：400ms Tween 后纹理层切换
	await get_tree().create_timer(0.6).timeout
	_check(not _bg.get_node("GridNeon").visible and _bg.get_node("DotGrid").visible, "elegant 纹理层切换")

	# 5. restore（RG-4）：profile=neon → 恢复 neon
	DB.save_profile({"theme": "neon"}, true)
	ThemeTokens.restore(get_tree().root)
	_check(ThemeTokens.current == "neon", "restore 恢复上次选择")
	_check(get_tree().root.theme != null, "restore root.theme")

	if _failures == 0:
		print("[test_theme] ALL PASS")
	get_tree().quit(1 if _failures > 0 else 0)
