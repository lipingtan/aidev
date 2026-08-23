extends Page
## 首页骨架（CR-1）：标题 + EmptyState 占位；on_enter 打印参数供冒烟验证
## 右上角 ThemeToggleButton（🎨 占位，CD §8.4）：按压缩放 + Sound.toggle() + 主题切换

@onready var _theme_btn: Button = $ThemeButton

func on_enter(data: Dictionary) -> void:
	print("[home] on_enter ", str(data))

func _ready() -> void:
	_theme_btn.pressed.connect(_on_theme_pressed)

## 🎨 按下：scale(.9) 反馈 + 切换音 + neon/elegant 切换（design §5 流程）
func _on_theme_pressed() -> void:
	var tw := create_tween()
	tw.tween_property(_theme_btn, "scale", Vector2(0.9, 0.9), 0.06)
	tw.tween_property(_theme_btn, "scale", Vector2.ONE, 0.12)
	Sound.toggle()
	var next := "elegant" if ThemeTokens.current == "neon" else "neon"
	ThemeTokens.apply_theme(next, get_tree().root)
