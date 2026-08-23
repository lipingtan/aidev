extends Control
## 底部四 Tab 栏（design §4）：首页/分类/搜索/我的；点击 → Nav.switch_tab，选中态用 token accent
## 子节点约定（main.tscn）：Btn0..Btn3（Button，顺序即 Tab 索引）

var _buttons: Array[Button] = []

func _ready() -> void:
	for i in 4:
		var b := get_node("Btn%d" % i) as Button
		_buttons.append(b)
		b.pressed.connect(_on_pressed.bind(i))
	EventBus.tab_changed.connect(_on_tab_changed)
	_on_tab_changed(Nav._current_tab if Nav._current_tab >= 0 else 0)

func _on_pressed(idx: int) -> void:
	Sound.click()
	Nav.switch_tab(idx)

## tab_changed → 选中态（accent）/未选中（ink2），颜色经 ThemeTokens
func _on_tab_changed(idx: int) -> void:
	var active := ThemeTokens.color("accent")
	var idle := ThemeTokens.color("ink2")
	for i in _buttons.size():
		_buttons[i].add_theme_color_override("font_color", active if i == idx else idle)
