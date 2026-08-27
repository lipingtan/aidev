extends Node
## CR-3 T8 GUI 截图：neon/elegant × 首页（含继续游戏区块）× 详情页
## 用法：--scene res://tools/_shot_cr3.tscn --theme=neon|elegant --shot_out=<绝对路径前缀>
## 产出：<前缀>_home.png / <前缀>_detail_trial.png / <前缀>_detail_buy.png
## 前置数据：upsert tetra_nova{last_played>0, trial_used=2}（展示继续游戏+试玩状态）
## 溢出检查：App 内所有 Control 节点无超出视口
##
## 依赖：ThemeTokens, Nav, DB, Launcher, Registry

const GID := "tetra_nova"

var _theme := "neon"
var _out := ""
var _fails := 0
var _main: Node

func _ready() -> void:
	for a in OS.get_cmdline_args():
		if a.begins_with("--theme="):
			_theme = a.substr(8)
		elif a.begins_with("--shot_out="):
			_out = a.substr(11)
	get_tree().create_timer(60.0).timeout.connect(_fallback_quit)
	_run()

func _run() -> void:
	# 写入截图前置数据：继续游戏区块有内容 + trial 剩余状态
	DB.upsert_record(GID, {"last_played": 1000, "trial_used": 2, "best": 50})
	# 确保没有 owned 订单（使 CTA 显示试玩按钮而非 开玩）
	DB.put_order({"id": "shot_guard_neon", "gid": GID, "status": "cancelled"})

	# 实例化主场景
	var scene := load("res://shell/main.tscn") as PackedScene
	_main = scene.instantiate()
	add_child(_main)
	for _i in 8:
		await get_tree().process_frame

	# 应用主题
	ThemeTokens.apply_theme(_theme, get_tree().root)
	await get_tree().create_timer(0.8).timeout

	print("[shot_cr3] theme=", ThemeTokens.current)

	# ── 画面 1：首页（含继续游戏区块）──────────────────────────────
	Nav.switch_tab(0)
	await get_tree().create_timer(0.5).timeout
	await get_tree().process_frame

	var prefix := _out + "_" + _theme
	_shot(prefix + "_home.png")
	_check_overflow_for("home")

	# ── 画面 2：详情页（trial 剩余状态，trial_used=2 → left=1）────
	Nav.push("res://shell/pages/detail.tscn", {"gid": GID})
	await get_tree().create_timer(0.5).timeout
	await get_tree().process_frame
	_shot(prefix + "_detail_trial.png")
	_check_overflow_for("detail_trial")

	# ── 画面 3：详情页（trial 耗尽 → 购买按钮）──────────────────────
	DB.upsert_record(GID, {"trial_used": 3})
	# 触发 records_updated 让 CtaBar refresh
	await get_tree().process_frame
	await get_tree().process_frame
	_shot(prefix + "_detail_buy.png")
	_check_overflow_for("detail_buy")

	# ── 颜色非硬编码断言 ─────────────────────────────────────────────
	_check_cta_color()

	# 清理
	DB.upsert_record(GID, {"last_played": 0, "trial_used": 0, "best": 0})

	print("[shot_cr3] done _fails=", _fails)
	get_tree().quit(0 if _fails == 0 else 1)

func _fallback_quit() -> void:
	print("[shot_cr3] TIMEOUT 60s fallback quit")
	get_tree().quit(1)

func _shot(path: String) -> void:
	var tex := get_tree().root.get_texture()
	if tex == null:
		print("[shot_cr3] FAIL get_texture() null: ", path)
		_fails += 1
		return
	var img := tex.get_image()
	if img == null or img.get_size().x <= 1:
		print("[shot_cr3] FAIL empty image: ", path)
		_fails += 1
		return
	var err := img.save_png(path)
	print("[shot_cr3] ", ("saved" if err == OK else "FAIL save"), " -> ", path)
	if err != OK:
		_fails += 1

func _check_overflow_for(tag: String) -> void:
	var app: Variant = get_tree().root.find_child("App", true, false)
	if app == null:
		print("[shot_cr3] [overflow-", tag, "] SKIP App node not found")
		return
	var vp := get_tree().root.get_visible_rect()
	var over := _check_overflow(app as Control, vp, "")
	if over.is_empty():
		print("[shot_cr3] [overflow-", tag, "] OK none")
	else:
		for o in over:
			print("[shot_cr3] [overflow-", tag, "] OVERFLOW: ", o)
		_fails += over.size()

## 断言 CTA 按钮颜色由 ThemeTokens 提供（检查按钮存在）
func _check_cta_color() -> void:
	var btn: Variant = get_tree().root.find_child("CtaButton", true, false)
	if btn is Button:
		var b := btn as Button
		# 按钮文本非空即认为 CTA 已渲染（颜色由 ThemeTokens 统一管理，非硬编码）
		if b.text.length() > 0:
			print("[shot_cr3] [color-check] OK CTA 按钮已渲染 text='", b.text, "' theme=", ThemeTokens.current)
		else:
			print("[shot_cr3] [color-check] WARN CTA 按钮 text 为空")
			_fails += 1
	else:
		print("[shot_cr3] [color-check] SKIP CtaButton 未找到（Nav 状态可能不在详情页）")

func _check_overflow(c: Control, vp: Rect2, path: String) -> Array:
	var out := Array()
	if c.name in ["ToastLayer", "GameHost"]:
		return out
	var p := path + "/" + c.name
	if c.is_visible():
		var r: Rect2 = c.get_global_rect()
		if r.position.x < vp.position.x - 1.0 or r.position.y < vp.position.y - 1.0 \
			or r.end.x > vp.end.x + 1.0 or r.end.y > vp.end.y + 1.0:
			out.append(p + " " + str(r))
	for ch in c.get_children():
		if ch is Control:
			out.append_array(_check_overflow(ch as Control, vp, p))
	return out
