extends Page
## 我的页（CR-7 FR-4）：头部信息 + 成就墙区块

@onready var _header: VBoxContainer = $RootVBox/Header
@onready var _ach_wall_root: Control = $RootVBox/ScrollView/ContentVBox/AchievementWall

var _section_ach: SectionAchievementWall = null

func on_enter(data: Dictionary) -> void:
	_render_header()
	if _section_ach:
		_section_ach.refresh()

func _render_header() -> void:
	for c: Node in _header.get_children():
		c.queue_free()
	var total_points: int = 0
	for aid: String in DB.get_achievements():
		var def: Variant = Registry.ach_def(aid)
		if def is Dictionary:
			total_points += int((def as Dictionary).get("points", 0))
	var total_hours: float = _calc_total_hours()

	var points_lbl := Label.new()
	points_lbl.text = "总成就点：%d" % total_points
	points_lbl.add_theme_color_override("font_color", ThemeTokens.color("gold"))
	_header.add_child(points_lbl)

	var hours_lbl := Label.new()
	hours_lbl.text = "累计时长：%.1f小时" % total_hours
	hours_lbl.add_theme_color_override("font_color", ThemeTokens.color("ink2"))
	_header.add_child(hours_lbl)

## B-2确认：total_playtime 为累加字段（launcher_util.finish_record 写入累加值）
func _calc_total_hours() -> float:
	var total: float = 0.0
	for gid: String in DB.list_records():
		var rec: Variant = DB.get_record(gid)
		if rec is Dictionary:
			total += float((rec as Dictionary).get("total_playtime", 0))
	return total / 3600.0
