extends Control
## 背景层（design §4/§5，CD §8.2）：nebula 色块 + 纹理层（霓虹网格线 / 青瓷点阵）
##
## - **仅监听 EventBus.theme_changed**（信号分工 Review B-2；settings_changed 不监听）
## - 主题切换 400ms Tween 过渡（唯一触发点），避免瞬切
## - 禁用 hdr_2d/glow：nebula 用半透明 ColorRect 模拟光晕
##
## 子节点约定（bglayer.tscn）：Bg / NebulaA|B|C（ColorRect）/ GridNeon / DotGrid（bg_grid.gd）

var _nebula: Array[ColorRect] = []
var _grid_neon: Control
var _dot_grid: Control
var _tween: Tween

func _ready() -> void:
	for n in get_children():
		if n is ColorRect and (n as Node).name.begins_with("Nebula"):
			_nebula.append(n as ColorRect)
	_grid_neon = get_node_or_null("GridNeon")
	_dot_grid = get_node_or_null("DotGrid")
	EventBus.theme_changed.connect(_on_theme_changed)
	_apply(ThemeTokens.current, true)

## theme_changed → 400ms Tween（仅此一处触发）
func _on_theme_changed(theme_name: String) -> void:
	_apply(theme_name, false)

## 应用主题：纹理层切换 + nebula 颜色/透明度（instant=true 时直设，用于初始状态）
func _apply(theme_name: String, instant: bool) -> void:
	var is_neon := theme_name == "neon"
	if _grid_neon != null:
		_grid_neon.visible = is_neon
		_grid_neon.queue_redraw()
	if _dot_grid != null:
		_dot_grid.visible = not is_neon
		_dot_grid.queue_redraw()
	# nebula（CD §8.2）：霓虹 2 块 opacity .14 / 青瓷 3 块 .5；色块取 accent/alt/bg_out
	var alpha := 0.14 if is_neon else 0.5
	var targets: Array[Color] = []
	for key in ["accent", "alt", "bg_out"]:
		var c := ThemeTokens.color(key)
		c.a = alpha
		targets.append(c)
	if _nebula.size() == 0:
		return
	if instant:
		for i in _nebula.size():
			_nebula[i].color = targets[i % targets.size()]
		return
	if _tween != null and _tween.is_valid():
		_tween.kill()
	_tween = create_tween()
	for i in _nebula.size():
		_tween.tween_property(_nebula[i], "color", targets[i % targets.size()], 0.4)
