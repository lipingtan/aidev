class_name GameMeta extends Resource
## 标准游戏元数据（CD §3.4）；数据源 games/*/meta.json，由 Registry 扫描加载
##
## - 内置游戏 size_kb=0；DLC 阶段填 dlc_url/size_kb（CR-1 仅内置 tetra_nova）
## - viewport_size/orientation 可省略：省略时继承 Shell 默认 720×1560 竖屏
## - aliases/pinyin 由内容编辑预填写，随资源包发布（Searcher 统一 lowercase / 首字母缩写运行时提取）
## - Arcade 特有字段（runtime="arcade" 时必填 core/roms）CR-1 不涉及
##
## 依赖：无（纯数据 Resource；加载方为 Registry）

## 游戏唯一 id（目录名一致）
var id: String = ""
## 显示标题
var title: String = ""
## 副标题（一句话卖点）
var subtitle: String = ""
## 分类：puzzle/action/arcade/casual/roguelike
var category: String = ""
## 标签（检索/推荐用）
var tags: Array[String] = []
## 定价模式：free|ad|trial|paid|iap
var price_model: String = ""
## 价格（元；free/ad/iap 为 0）
var price: int = 0
## 试玩限制：{plays:3} 或 {minutes:10}
var trial: Dictionary = {}
## 版本号
var version: String = ""
## 资源包大小 KB（DLC 用；内置为 0）
var size_kb: int = 0
## 图标路径
var icon: String = ""
## 截图路径列表
var screenshots: Array[String] = []
## 游戏主场景路径
var scene: String = ""
## DLC CDN 地址（DLC 阶段填）
var dlc_url: String = ""
## 逻辑设计分辨率 [w,h]；ZERO 表示继承 Shell 默认
var viewport_size: Vector2i = Vector2i.ZERO
## 屏幕方向 portrait|landscape；空表示继承 Shell（portrait）
var orientation: String = ""
## 中英文别名数组（大小写不敏感）
var aliases: Array[String] = []
## 无声调全拼数组（首字母缩写运行时自动提取）
var pinyin: Array[String] = []
## 成就定义 [{id, name, points}]
var achievements: Array[Dictionary] = []
## 游戏描述
var desc: String = ""
## 运行时：pck|html|arcade
var runtime: String = "pck"
## 运行时依赖核心包名（arcade="myosd-0.288"，其余为空）
var core: String = ""

## Shell 默认设计分辨率（720×1560 竖屏）
const SHELL_VIEWPORT := Vector2i(720, 1560)

## 视口尺寸：meta 未声明时继承 Shell 默认
func get_viewport_size() -> Vector2i:
	return viewport_size if viewport_size != Vector2i.ZERO else SHELL_VIEWPORT

## 屏幕方向：meta 未声明时继承 Shell（portrait）
func get_orientation() -> String:
	return orientation if orientation != "" else "portrait"

## 从 meta.json 字典构造（缺失字段取默认值，类型不匹配安全降级）
static func from_dict(d: Dictionary) -> GameMeta:
	var m := GameMeta.new()
	m.id = str(d.get("id", ""))
	m.title = str(d.get("title", ""))
	m.subtitle = str(d.get("subtitle", ""))
	m.category = str(d.get("category", ""))
	m.tags = _str_array(d.get("tags"))
	m.price_model = str(d.get("price_model", ""))
	m.price = int(d.get("price", 0))
	m.trial = d.get("trial", {}) as Dictionary
	m.version = str(d.get("version", ""))
	m.size_kb = int(d.get("size_kb", 0))
	m.icon = str(d.get("icon", ""))
	m.screenshots = _str_array(d.get("screenshots"))
	m.scene = str(d.get("scene", ""))
	m.dlc_url = str(d.get("dlc_url", ""))
	# d.get() 返回 Variant，显式判型后取用（避免 Variant 推断告警）
	var vp_raw: Variant = d.get("viewport_size")
	if vp_raw is Array and (vp_raw as Array).size() == 2:
		m.viewport_size = Vector2i(int((vp_raw as Array)[0]), int((vp_raw as Array)[1]))
	m.orientation = str(d.get("orientation", ""))
	m.aliases = _str_array(d.get("aliases"))
	m.pinyin = _str_array(d.get("pinyin"))
	var achs_raw: Variant = d.get("achievements")
	if achs_raw is Array:
		for a in achs_raw as Array:
			if a is Dictionary:
				m.achievements.append(a as Dictionary)
	m.desc = str(d.get("desc", ""))
	m.runtime = str(d.get("runtime", "pck"))
	m.core = str(d.get("core", ""))
	return m

## 安全转字符串数组（null/非数组返回空）
static func _str_array(v: Variant) -> Array[String]:
	var out: Array[String] = []
	if v is Array:
		for s in v as Array:
			out.append(str(s))
	return out
