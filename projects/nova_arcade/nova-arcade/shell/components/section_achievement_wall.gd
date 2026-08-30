class_name SectionAchievementWall extends Control
## 成就墙组件（CR-7 FR-4）：总进度条 + 按游戏分组 + 全局成就区
## 行数目标：80~150行

@onready var _root: VBoxContainer = $VBox

func refresh() -> void:
	for c: Node in _root.get_children():
		c.queue_free()
	_render_progress_bar()
	_render_by_game_groups()
	_render_global_achievements()

## 总进度条：已解锁/总数 + 百分比
func _render_progress_bar() -> void:
	var total: int = 0
	var unlocked: int = 0
	for m: GameMeta in Registry.all():
		total += m.achievements.size()
	for gid: String in DB.get_achievements():
		var def: Variant = Registry.ach_def(gid)
		if def != null:
			unlocked += 1
	# 全局成就也计入
	total += 3  # global_collector/reviewer/marathon
	for gdef in ["global_collector", "global_reviewer", "global_marathon"]:
		if DB.get_achievements().has(gdef):
			unlocked += 1

	var bar := ProgressBar.new()
	bar.anchor_right = 1.0
	bar.anchor_bottom = 1.0
	bar.grow_horizontal = 2
	bar.min_value = 0
	bar.max_value = maxi(total, 1)
	bar.value = unlocked
	bar.add_theme_color_override("fill", ThemeTokens.color("ach_line"))

	var lbl := Label.new()
	lbl.text = "成就 %d/%d (%.0f%%)" % [unlocked, total, (unlocked / float(maxi(total, 1))) * 100.0]
	lbl.add_theme_color_override("font_color", ThemeTokens.color("ink2"))

	_root.add_child(bar)
	_root.add_child(lbl)

## 按游戏分组渲染成就列表
func _render_by_game_groups() -> void:
	var achs: Dictionary = DB.get_achievements()
	for m: GameMeta in Registry.all():
		if m.achievements.size() == 0:
			continue
		# 分组标题
		var hdr := Label.new()
		hdr.text = "🎮 %s" % m.title
		hdr.anchor_right = 1.0
		hdr.grow_horizontal = 2
		hdr.add_theme_color_override("font_color", ThemeTokens.color("ink"))
		_root.add_child(hdr)
		# 成就列表
		for a in m.achievements:
			var aid: String = str(a.get("id", ""))
			var is_unlocked: bool = achs.has(aid)
			var row := HBoxContainer.new()
			row.anchor_right = 1.0
			row.grow_horizontal = 2
			var icon := Label.new()
			icon.text = "🏆" if is_unlocked else "☆"
			var name := Label.new()
			name.text = str(a.get("name", aid))
			name.anchor_right = 1.0
			name.grow_horizontal = 2
			if is_unlocked:
				name.add_theme_color_override("font_color", ThemeTokens.color("gold"))
			else:
				name.modulate = ThemeTokens.color("ink2")
			var pts := Label.new()
			pts.text = "+%d" % int(a.get("points", 0))
			pts.horizontal_alignment = 2
			if is_unlocked:
				pts.add_theme_color_override("font_color", ThemeTokens.color("gold"))
			else:
				pts.modulate = ThemeTokens.color("ink2")
			row.add_child(icon)
			row.add_child(name)
			row.add_child(pts)
			_root.add_child(row)

## 全局成就区（收藏家/好评人/马拉松）
func _render_global_achievements() -> void:
	var achs: Dictionary = DB.get_achievements()
	var gdefs: Array[Dictionary] = [
		{"id": "global_collector", "name": "收藏家", "desc": "拥有全部游戏记录", "points": 30},
		{"id": "global_reviewer", "name": "好评人", "desc": "提交≥3条评价", "points": 20},
		{"id": "global_marathon", "name": "马拉松", "desc": "累计游玩≥10小时", "points": 50},
	]
	var hdr := Label.new()
	hdr.text = "🌟 全局成就"
	hdr.anchor_right = 1.0
	hdr.grow_horizontal = 2
	hdr.add_theme_color_override("font_color", ThemeTokens.color("ink"))
	_root.add_child(hdr)

	for gdef in gdefs:
		var aid: String = str(gdef["id"])
		var is_unlocked: bool = achs.has(aid)
		var row := HBoxContainer.new()
		row.anchor_right = 1.0
		row.grow_horizontal = 2
		var icon := Label.new()
		icon.text = "🏆" if is_unlocked else "☆"
		var name := Label.new()
		name.text = "%s — %s" % [str(gdef.get("name", aid)), str(gdef.get("desc", ""))]
		name.anchor_right = 1.0
		name.grow_horizontal = 2
		if is_unlocked:
			name.add_theme_color_override("font_color", ThemeTokens.color("gold"))
		else:
			name.modulate = ThemeTokens.color("ink2")
		var pts := Label.new()
		pts.text = "+%d" % int(gdef.get("points", 0))
		pts.horizontal_alignment = 2
		if is_unlocked:
			pts.add_theme_color_override("font_color", ThemeTokens.color("gold"))
		else:
			pts.modulate = ThemeTokens.color("ink2")
		row.add_child(icon)
		row.add_child(name)
		row.add_child(pts)
		_root.add_child(row)
