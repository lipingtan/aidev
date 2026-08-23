class_name ThemeTokens
## 主题 Token（design.md §5 / CD §8.1）——两套主题共用场景树，仅切换 Token + Theme 资源
##
## - TOKENS：颜色 token 双套（霓虹·星穹 neon / 青瓷·素笺 elegant），值按 CD §8.1 对照表
## - GRADS：渐变 token（停色数组；elegant 多为纯色）
## - EXTRA：非颜色 token（SectionHeader 前缀符号/衬线字体）
## - **主题色禁止硬编码**：非 Theme 控件一律 ThemeTokens.color(key)（override §2）
## - apply_theme(name, root)：Token → root.theme → profile force 写 → settings_changed + theme_changed
## - restore(root)：启动读 profile 恢复上次选择（RG-4）
##
## 依赖：DB（profile）、EventBus（settings_changed/theme_changed）

## 默认主题
const DEFAULT := "neon"

## 当前主题名
static var current: String = DEFAULT

## 颜色 token（CD §8.1）
const TOKENS: Dictionary = {
	"neon": {
		"bg": Color(0.02, 0.024, 0.059),
		"bg_out": Color(0.008, 0.012, 0.031),
		"card": Color(0.039, 0.063, 0.133, 0.72),
		"card2": Color(0.055, 0.086, 0.188, 0.92),
		"ink": Color(0.812, 0.918, 1),
		"ink2": Color(0.561, 0.722, 0.847),
		"ink3": Color(0.357, 0.49, 0.6),
		"accent": Color(0.169, 0.91, 1),
		"accent_soft": Color(0.169, 0.91, 1, 0.1),
		"alt": Color(0.784, 0.42, 1),
		"alt_soft": Color(0.784, 0.42, 1, 0.12),
		"gold": Color(1, 0.824, 0.247),
		"gold_soft": Color(1, 0.824, 0.247, 0.16),
		"ok": Color(0.271, 1, 0.533),
		"play_ink": Color(0.016, 0.063, 0.11),
		"buy_ink": Color(0.141, 0.075, 0),
		"line": Color(0.169, 0.91, 1, 0.18),
		"line2": Color(1, 1, 1, 0.08),
		"star_off": Color(0.227, 0.322, 0.408),
		"rank1": Color(1, 0.824, 0.247),
		"rank2": Color(0.875, 0.91, 0.949),
		"rank3": Color(1, 0.596, 0.22),
		"banner_line": Color(0.784, 0.42, 1, 0.35),
		"ov_bg": Color(0.012, 0.016, 0.047, 0.78),
		"toast_bg": Color(0.039, 0.071, 0.149, 0.95),
		"toast_ink": Color(0.875, 0.957, 1),
		"toast_line": Color(0.169, 0.91, 1, 0.3),
		"ach_bg": Color(1, 0.824, 0.247, 0.07),
		"ach_line": Color(1, 0.824, 0.247, 0.35),
		"myrev_bg": Color(1, 0.824, 0.247, 0.04),
		"myrev_line": Color(1, 0.824, 0.247, 0.4),
		"rbar_bg": Color(1, 1, 1, 0.07),
		"icon_end": Color(0.039, 0.063, 0.188),
		"sw_off": Color(1, 1, 1, 0.12),
		"sw_on": Color(0.169, 0.91, 1, 0.5)
	},
	"elegant": {
		"bg": Color(0.965, 0.957, 0.937),
		"bg_out": Color(0.918, 0.902, 0.867),
		"card": Color(1, 1, 1),
		"card2": Color(1, 1, 1),
		"ink": Color(0.184, 0.227, 0.208),
		"ink2": Color(0.435, 0.478, 0.455),
		"ink3": Color(0.604, 0.647, 0.627),
		"accent": Color(0.29, 0.49, 0.392),
		"accent_soft": Color(0.906, 0.941, 0.918),
		"alt": Color(0.616, 0.525, 0.71),
		"alt_soft": Color(0.937, 0.914, 0.961),
		"gold": Color(0.722, 0.576, 0.29),
		"gold_soft": Color(0.965, 0.937, 0.867),
		"ok": Color(0.29, 0.49, 0.392),
		"play_ink": Color(0.992, 0.988, 0.976),
		"buy_ink": Color(0.992, 0.98, 0.961),
		"line": Color(0.902, 0.882, 0.839),
		"line2": Color(0.937, 0.925, 0.894),
		"star_off": Color(0.863, 0.839, 0.784),
		"rank1": Color(0.722, 0.576, 0.29),
		"rank2": Color(0.435, 0.478, 0.455),
		"rank3": Color(0.788, 0.561, 0.373),
		"banner_line": Color(1, 1, 1, 0.8),
		"ov_bg": Color(0.227, 0.259, 0.235, 0.32),
		"toast_bg": Color(0.184, 0.227, 0.208, 0.92),
		"toast_ink": Color(0.949, 0.941, 0.914),
		"toast_line": Color(0.184, 0.227, 0.208, 0.2),
		"ach_bg": Color(0.965, 0.937, 0.867),
		"ach_line": Color(0.91, 0.867, 0.753),
		"myrev_bg": Color(0.992, 0.984, 0.957),
		"myrev_line": Color(0.875, 0.839, 0.741),
		"rbar_bg": Color(0.937, 0.925, 0.894),
		"icon_end": Color(0.957, 0.945, 0.918),
		"sw_off": Color(0.902, 0.882, 0.839),
		"sw_on": Color(0.29, 0.49, 0.392)
	}
}

## 渐变 token（停色数组；elegant 纯色单停）
const GRADS: Dictionary = {
	"neon": {
		"play": [Color(0.169, 0.91, 1), Color(0.784, 0.42, 1)],
		"buy": [Color(1, 0.824, 0.247), Color(1, 0.596, 0.22)],
		"banner": [Color(0.071, 0.133, 0.302), Color(0.227, 0.082, 0.376)],
		"shot": [Color(0.063, 0.11, 0.247), Color(0.141, 0.063, 0.251)],
		"av": [Color(0.165, 0.235, 0.431), Color(0.353, 0.165, 0.502)],
		"libhead": [Color(0.169, 0.91, 1, 0.09), Color(0.784, 0.42, 1, 0.09)],
		"bar": [Color(0.169, 0.91, 1), Color(0.784, 0.42, 1)],
		"rbar": [Color(1, 0.824, 0.247), Color(1, 0.596, 0.22)],
		"score": [Color(1, 1, 1), Color(0.169, 0.91, 1)]
	},
	"elegant": {
		"play": [Color(0.29, 0.49, 0.392)],
		"buy": [Color(0.788, 0.561, 0.373)],
		"banner": [Color(0.89, 0.925, 0.898), Color(0.925, 0.89, 0.941), Color(0.957, 0.918, 0.851)],
		"shot": [Color(0.898, 0.925, 0.898), Color(0.914, 0.882, 0.937)],
		"av": [Color(0.859, 0.906, 0.867), Color(0.902, 0.863, 0.933)],
		"libhead": [Color(0.914, 0.941, 0.918), Color(0.937, 0.914, 0.953)],
		"bar": [Color(0.29, 0.49, 0.392), Color(0.616, 0.525, 0.71)],
		"rbar": [Color(0.722, 0.576, 0.29), Color(0.788, 0.561, 0.373)],
		"score": [Color(0.29, 0.49, 0.392), Color(0.29, 0.49, 0.392)]
	}
}

## 非颜色 token
const EXTRA: Dictionary = {
	"neon": {"sechead_mark": "▶ ", "serif": ""},
	"elegant": {"sechead_mark": "❋ ", "serif": "Georgia"}
}

## 取当前主题颜色 token；未知 key 告警返回白
static func color(key: String) -> Color:
	var t: Variant = TOKENS.get(current)
	if t is Dictionary and (t as Dictionary).has(key):
		return (t as Dictionary)[key] as Color
	push_warning("ThemeTokens: 未知颜色 token %s" % key)
	return Color.WHITE

## 取当前主题字符串 token（sechead_mark/serif）
static func extra(key: String) -> String:
	var t: Variant = EXTRA.get(current)
	if t is Dictionary and (t as Dictionary).has(key):
		return str((t as Dictionary)[key])
	return ""

## 取当前主题渐变（停色数组 → Gradient；4.5 API：offsets + colors 直接赋值）
static func grad(key: String) -> Gradient:
	var g := Gradient.new()
	var stops_raw: Variant = GRADS.get(current)
	if stops_raw is Dictionary and (stops_raw as Dictionary).has(key):
		var stops: Array = ((stops_raw as Dictionary)[key]) as Array
		var cols := PackedColorArray()
		var offs := PackedFloat32Array()
		for i in stops.size():
			cols.append(stops[i] as Color)
			offs.append(float(i) / float(maxi(stops.size() - 1, 1)))
		g.colors = cols
		g.offsets = offs
	else:
		push_warning("ThemeTokens: 未知渐变 token %s" % key)
	return g

## 游戏图标动态渐变：game_col → icon_end（CD §8.4 GameCard）
static func icon_grad(game_col: Color) -> Gradient:
	var g := Gradient.new()
	g.colors = PackedColorArray([game_col, color("icon_end")])
	g.offsets = PackedFloat32Array([0.0, 1.0])
	return g

## 切换主题入口（design §5 流程）：Token → root.theme → profile force 写 → 双信号
static func apply_theme(name: String, root: Window) -> void:
	if not TOKENS.has(name):
		push_warning("ThemeTokens: 未知主题 %s" % name)
		return
	current = name
	root.theme = load("res://shell/theme/theme_%s.tres" % name) as Theme
	DB.save_profile({"theme": name}, true)
	EventBus.settings_changed.emit("theme", name)
	EventBus.theme_changed.emit(name)

## 启动恢复：读 profile 上次选择（缺省 neon，RG-4）。
## 偏差：恢复后发 theme_changed（BgLayer _ready 先于 Main._ready，需借此信号按恢复值重渲染）
static func restore(root: Window) -> void:
	var t := str(DB.get_profile().get("theme", DEFAULT))
	if not TOKENS.has(t):
		t = DEFAULT
	current = t
	root.theme = load("res://shell/theme/theme_%s.tres" % t) as Theme
	EventBus.theme_changed.emit(t)
