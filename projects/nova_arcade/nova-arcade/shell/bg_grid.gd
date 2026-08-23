extends Control
## 背景纹理层（CD §8.2）：霓虹=透视网格线（底部 34%），青瓷=点阵（底部 30%）
## 颜色经 ThemeTokens，无硬编码；由 BgLayer 按主题切换 visible

## neon | elegant
@export var style: String = "neon"

func _draw() -> void:
	if style == "neon":
		_draw_grid()
	else:
		_draw_dots()

## 透视网格线（青色细线，底部 34%；横线间距递增模拟透视）
func _draw_grid() -> void:
	var c := ThemeTokens.color("accent")
	c.a = 0.25
	var h := size.y
	var top := h * 0.66
	var y := top
	var gap := 8.0
	while y < h:
		draw_line(Vector2(0, y), Vector2(size.x, y), c, 1.0)
		y += gap
		gap *= 1.35
	# 竖线向顶部中点透视收敛
	var cx := size.x * 0.5
	for i in range(-6, 7):
		draw_line(Vector2(cx + float(i) * 40.0, top), Vector2(cx + float(i) * 90.0, h), c, 1.0)

## 点阵（墨色低透明，底部 30%，22px 间距）
func _draw_dots() -> void:
	var c := ThemeTokens.color("ink")
	c.a = 0.05
	var step := 22
	var top := int(size.y * 0.7)
	for y in range(top, int(size.y), step):
		for x in range(0, int(size.x), step):
			draw_circle(Vector2(x, y), 1.5, c)
