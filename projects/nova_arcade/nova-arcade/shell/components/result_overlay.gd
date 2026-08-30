extends Control
## 结算卡（CR-2 T6）：本局得分/用时/历史最佳 + 成就提示 + [再来一局][评价*][返回] + 「换一个」区
## - 评价按钮（解读注记2）：finish_count≥2 且本局 playtime≥600s → 可点亮；否则灰显
##   （playtime 不足 →「还需 X 分钟」；局数不足 →「还需 N 局」），提交链路 CR-3/M1
## - 「换一个」（Q2=A）：Registry.similar(gid)；空数组 → 隐藏整个区域；非空 → 迷你卡，
##   点击 → Toast「详情页即将上线」（详情页本身 CR-3）
## - [返回]：Launcher.release() + Nav 回首页
## - 成就提示逐条弹出（DB.unlock + achievement_unlocked；本 CR 恒空数组 → 不弹）

## 评价解锁：累计时长阈值（秒）
const REVIEW_MIN_S := 600.0
## 评价解锁：最少局数
const REVIEW_RUNS := 2

@onready var _dim: ColorRect = $Dim as ColorRect
@onready var _card: PanelContainer = $Card as PanelContainer
@onready var _title: Label = $Card/VBox/Title as Label
@onready var _score_line: Label = $Card/VBox/ScoreLine as Label
@onready var _best_line: Label = $Card/VBox/BestLine as Label
@onready var _ach_box: VBoxContainer = $Card/VBox/AchBox as VBoxContainer
@onready var _btn_again: Button = $Card/VBox/BtnRow/BtnAgain as Button
@onready var _btn_review: Button = $Card/VBox/BtnRow/BtnReview as Button
@onready var _btn_back: Button = $Card/VBox/BtnRow/BtnBack as Button
@onready var _switch_box: VBoxContainer = $Card/VBox/SwitchBox as VBoxContainer
@onready var _switch_row: HBoxContainer = $Card/VBox/SwitchBox/SwitchRow as HBoxContainer

## 当前展示的游戏 id（按钮回调复用；连接在 _ready 一次性建立，show_card 可重复调用）
var _gid: String = ""
## P-1：缓存本局 playtime，用于传递给 ReviewEditor（非 total_playtime，防止绕过防刷）
var _result_playtime: float = 0.0

func _ready() -> void:
	visible = false
	_btn_again.pressed.connect(_on_again)
	_btn_review.pressed.connect(_on_review)
	_btn_back.pressed.connect(_on_back)
	_dim.gui_input.connect(_on_dim_tap)

## 点击遮罩关闭结算卡
func _on_dim_tap(event: InputEvent) -> void:
	if event is InputEventMouseButton and (event as InputEventMouseButton).pressed:
		visible = false
		_set_layer_visible(false)

## 显示结算卡（Launcher 退出链 400ms 后调用；可重复）
func show_card(gid: String, result: Dictionary) -> void:
	_gid = gid
	_result_playtime = float(result.get("playtime", 0.0))
	_apply_theme()
	var meta: GameMeta = Registry.lookup(gid)
	_title.text = meta.title if meta != null else gid
	var score := int(result.get("score", 0))
	var playtime := float(result.get("playtime", 0.0))
	_score_line.text = "得分 %d    用时 %s" % [score, _fmt_time(playtime)]
	var rec: Variant = DB.get_record(gid)
	var best := 0
	if rec is Dictionary:
		best = int((rec as Dictionary).get("best", 0))
	_best_line.text = "历史最佳 %d%s" % [maxi(best, score), "（新纪录）" if score > best and score > 0 else ""]
	_fill_achievements(result)
	_fill_review_button(gid, playtime)
	_fill_switch_area(gid)
	_set_layer_visible(true)  # 父层 OverlayLayer 默认隐藏，展示前必须打开
	visible = true

## 显隐父层（OverlayLayer）：只在自己展示时打开层；关闭时只藏自己——
## OverlayLayer 还承载 ConfirmBubble/ReviewEditor 等同层弹窗，关闭连坐会把它们一起藏掉
func _set_layer_visible(v: bool) -> void:
	if not v:
		return
	var p := get_parent()
	if p is Control and (p as Control).name != "root":
		(p as Control).visible = v

## 成就提示逐条弹出（DB.unlock + achievement_unlocked；本 CR 恒空 → 隐藏区域）
func _fill_achievements(result: Dictionary) -> void:
	var aids: Array = result.get("achievements", [])
	for child in _ach_box.get_children():
		child.queue_free()
	if aids.is_empty():
		_ach_box.visible = false
		return
	_ach_box.visible = true
	for aid in aids:
		DB.unlock(str(aid))
		var lbl := Label.new()
		lbl.text = "★ 成就解锁：%s" % str(aid)
		lbl.add_theme_color_override("font_color", ThemeTokens.color("gold"))
		_ach_box.add_child(lbl)

## 评价按钮：finish_count≥2 且本局时长达标 → 可点亮；否则灰显（解读注记2）
func _fill_review_button(gid: String, playtime: float) -> void:
	var rec: Variant = DB.get_record(gid)
	var finish := 0
	if rec is Dictionary:
		finish = int((rec as Dictionary).get("finish_count", 0))
	if finish >= REVIEW_RUNS and playtime >= REVIEW_MIN_S:
		_btn_review.text = "评价"
		_btn_review.disabled = false
	elif playtime < REVIEW_MIN_S:
		_btn_review.text = "还需 %d 分钟" % ceili(REVIEW_MIN_S - playtime)
		_btn_review.disabled = true
	else:
		_btn_review.text = "还需 %d 局" % (REVIEW_RUNS - finish)
		_btn_review.disabled = true

## 「换一个」区：similar 空数组隐藏；非空迷你卡点击 → Toast（Q2=A）
func _fill_switch_area(gid: String) -> void:
	for child in _switch_row.get_children():
		child.queue_free()
	var sims := Registry.similar(gid)
	if sims.is_empty():
		_switch_box.visible = false
		return
	_switch_box.visible = true
	for m in sims:
		var card := Button.new()
		card.text = m.title
		card.custom_minimum_size = Vector2(180, 64)
		card.add_theme_font_size_override("font_size", 20)
		card.pressed.connect(_on_similar_pressed)
		_switch_row.add_child(card)

## 迷你卡点击：详情页即将上线（CR-3）
func _on_similar_pressed() -> void:
	_toast("详情页即将上线")

## [再来一局]：隐藏卡片 + Launcher.play_again（复用模块，不耗 trial）
func _on_again() -> void:
	visible = false
	_set_layer_visible(false)
	Launcher.play_again(_gid)

## [评价]：弹出 ReviewEditor，注入本局 playtime（P-1：非 total_playtime，防绕过防刷）
func _on_review() -> void:
	var editor := get_tree().root.find_child("ReviewEditor", true, false)
	if editor != null:
		editor.call("show_modal", _gid, _result_playtime)

## [返回]：释放模块 + 回首页
func _on_back() -> void:
	visible = false
	_set_layer_visible(false)
	Launcher.release()
	Nav.switch_tab(0)

## 主题色（仅经 ThemeTokens.color，override §2）
func _apply_theme() -> void:
	_dim.color = ThemeTokens.color("ov_bg")
	var sb := StyleBoxFlat.new()
	sb.bg_color = ThemeTokens.color("card2")
	sb.border_color = ThemeTokens.color("line")
	sb.set_border_width_all(1)
	sb.set_corner_radius_all(12)
	_card.add_theme_stylebox_override("panel", sb)
	_title.add_theme_color_override("font_color", ThemeTokens.color("ink"))
	_score_line.add_theme_color_override("font_color", ThemeTokens.color("ink"))
	_best_line.add_theme_color_override("font_color", ThemeTokens.color("ink2"))

## 秒 → mm:ss
func _fmt_time(sec: float) -> String:
	var m := int(sec) / 60
	return "%d:%02d" % [m, int(sec) % 60]

## Toast（经 find_child 找 UI 层）
func _toast(msg: String) -> void:
	var layer := get_tree().root.find_child("ToastLayer", true, false)
	if layer != null:
		layer.call("show_msg", msg)
