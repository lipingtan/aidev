# 设计：CR-4 分类页 + 搜索页 + 本地评价系统

> 依据：已确认的 `requirements.md`（Q1~Q6）+ `design_plan.md`（D1~D8 全采纳，2026-08-25）。
> 上游：client-design §4.3/§4.4/§4.8 + override §2/§5 + hybrid §3.2。
> 工程实查结论：Registry.query 已支持 category/tag/price，不支持 price_models[]/min_rating/is_new（本 CR 扩展）；DB.add_search_history 已实现去重置顶 ✅；DB.put_review 已是覆盖语义 ✅；DB.delete_review 不存在（本 CR 补）；GameCard 单一尺寸（本 CR 加 size 参数）；result_overlay._on_review 仅 Toast（本 CR 接线 ReviewEditor）；editorial.json 无 hot_queries 字段（本 CR 补）。

---

## 1. 文件与目录变更

```
nova-arcade/
├── core/
│   ├── registry.gd          # 扩展 query()：price_models[]/min_rating/is_new 三个新维度
│   └── db.gd                # 新增 delete_review(gid)
├── data/
│   └── editorial.json       # 新增 hot_queries 字段（6 条本地热搜词）
├── services/
│   └── searcher.gd          # 新增：Searcher 服务（build_index/query/hot_queries）
├── shell/
│   ├── components/
│   │   └── game_card.gd/.tscn  # 扩展：setup() 加 size 参数（"card"/"grid"）
│   ├── pages/
│   │   ├── category.gd/.tscn   # 改写：完整分类筛选交互
│   │   ├── search.gd/.tscn     # 改写：完整搜索交互
│   │   ├── detail.gd           # 补充：添加"全部评价→"按钮
│   │   ├── reviews.gd/.tscn    # 新增：评价列表页
│   │   └── result_overlay.gd   # 修改：_on_review() 接线 ReviewEditor
│   └── overlays/
│       └── review_editor.gd/.tscn  # 新增：写/改评价模态弹窗
└── tools/
    ├── test_category.tscn/.gd  # 新增：分类页 headless 测试
    ├── test_search.tscn/.gd    # 新增：搜索页 headless 测试
    └── test_review.tscn/.gd    # 新增：评价系统 headless 测试
```

**Main.tscn 变更**：Services 节点下挂 Searcher 非单例节点；OverlayLayer 下挂 ReviewEditor 静态节点（visible=false）；Nav._modal_dialog_open() 扩展检查 ReviewEditor.visible。

---

## 2. Registry.query() 扩展（core/registry.gd）

在现有 query() 基础上追加三个过滤维度，不改接口签名（filter dict 扩展 key）：

```gdscript
## 扩展后完整 filter 支持：
## {category, tag, price, price_models:Array[String], min_rating:float, is_new:bool, sort}
func query(filters: Dictionary) -> Array[GameMeta]:
    var out := all()
    if filters.has("category") and str(filters["category"]) != "":
        out = _filter_category(out, str(filters["category"]))
    if filters.has("tag"):
        out = _filter_tag(out, str(filters["tag"]))
    if filters.has("price"):
        out = _filter_price(out, int(filters["price"]))
    # === 新增三个维度 ===
    if filters.has("price_models"):
        out = _filter_price_models(out, filters["price_models"] as Array)
    if filters.has("min_rating"):
        out = _filter_min_rating(out, float(filters["min_rating"]))
    if filters.has("is_new") and bool(filters["is_new"]):
        out = _filter_is_new(out)
    # ==================
    var sort := str(filters.get("sort", "title"))
    match sort:
        "price": out.sort_custom(_cmp_price)
        "version": out.sort_custom(_cmp_version)
        _: out.sort_custom(_cmp_title)
    return out

## 按 price_model 白名单过滤（多选叠加）
func _filter_price_models(list: Array[GameMeta], models: Array) -> Array[GameMeta]:
    if models.is_empty(): return list
    var out: Array[GameMeta] = []
    for m in list:
        if models.has(m.price_model): out.append(m)
    return out

## 按最低评分过滤（M1 meta 无评分数据，rating 默认 0 → 全部通过；M3 云端聚合后自动生效）
func _filter_min_rating(list: Array[GameMeta], min_r: float) -> Array[GameMeta]:
    var out: Array[GameMeta] = []
    for m in list:
        if float(m.get("rating", 0.0)) >= min_r: out.append(m)
    return out

## 新游：version 字段含当月日期标记（M1 简化：预留，条件恒 true，M2 补发布日期字段）
func _filter_is_new(list: Array[GameMeta]) -> Array[GameMeta]:
    return list  # M1 占位：无发布日期字段，is_new 暂不过滤（标注 M2 补 published_at 字段）
```

> **注意**：category="" 表示全部（新增判断 `!= ""`），与 D-7 对齐。

---

## 3. DB.delete_review()（core/db.gd 补充）

```gdscript
## 删除评价并立即落盘；不存在则静默忽略
func delete_review(gid: String) -> void:
    if not _reviews.has(gid): return
    _reviews.erase(gid)
    _io.write_json("reviews.json", _reviews)
```

---

## 4. editorial.json 更新

```json
{
  "banner": [],
  "featured": [],
  "categories": [],
  "hot_queries": ["TETRA NOVA", "俄罗斯方块", "消除", "益智", "街机", "roguelike"]
}
```

---

## 5. Searcher 服务（services/searcher.gd）

非单例，挂 Main/Services 节点下；冷启动时由 Main._ready() 在 Registry.reload() 后调用 build_index()。

```gdscript
class_name Searcher extends Node
## 本地搜索索引：构建 + 打分查询（client-design §4.4）
## 非单例；冷启动 Main._ready() → Registry.reload() → Searcher.build_index()

## 单游戏索引项
class IndexEntry:
    var gid: String = ""
    var title_lower: String = ""
    var aliases_lower: Array[String] = []
    var pinyin_full: Array[String] = []    ## 全拼数组（小写）
    var pinyin_initials: Array[String] = [] ## 首字母缩写数组（运行时提取）
    var tags_lower: Array[String] = []

var _index: Array = []  ## Array[IndexEntry]
var _editorial: Dictionary = {}

func _ready() -> void:
    _editorial = Registry.get_editorial()

## 从 Registry.all() 构建内存索引
## 注意：IndexEntry 数组字段必须在此处显式初始化为新数组，避免 GDScript 类默认值跨实例共享引用（A-2）
func build_index() -> void:
    _index.clear()
    _editorial = Registry.get_editorial()
    for meta in Registry.all():
        var e := IndexEntry.new()
        e.gid = (meta as GameMeta).id
        e.title_lower = (meta as GameMeta).title.to_lower()
        e.aliases_lower = []   ## 显式新数组，防止共享引用
        e.pinyin_full = []
        e.pinyin_initials = []
        e.tags_lower = []
        for alias in (meta as GameMeta).aliases:
            e.aliases_lower.append(str(alias).to_lower())
        for pf in (meta as GameMeta).pinyin:
            var p := str(pf).to_lower()
            e.pinyin_full.append(p)
            e.pinyin_initials.append(_extract_initials(p))
        for tag in (meta as GameMeta).tags:
            e.tags_lower.append(str(tag).to_lower())
        _index.append(e)

## 从全拼串提取首字母缩写（每个汉字对应一段，取各段首字母）
## 例："xinxingfangzhen" 按无分隔符不可直接提取；
## meta.json pinyin 字段若存储为单字全拼数组则每项取首字母
## M1 简化：直接取整串第一个字符作为单字母索引前缀，完整缩写 M2 补
func _extract_initials(pf: String) -> String:
    if pf == "": return ""
    return pf.left(1)  # M1 占位；M2 改为按分隔符或字数拆分

## 打分查询（同步返回）；调用方负责 150ms 去抖
func query(text: String) -> Array[String]:
    if text == "": return []
    var q := text.to_lower().strip_edges()
    var scores: Dictionary = {}  ## gid -> int
    for e in _index:
        var entry := e as IndexEntry
        var score: int = _score_entry(entry, q)
        if score > 0:
            scores[entry.gid] = score
    ## 按分数倒序排列（A-1：不可用单行 lambda，改用具名比较器 + 成员缓存）
    _sort_scores = scores
    var gids: Array[String] = []
    for gid in scores:
        gids.append(gid)
    gids.sort_custom(_cmp_score_desc)
    _sort_scores = {}
    return gids.slice(0, 20)

## 打分比较器（A-1：替代单行 lambda，规避 Godot 4.5 单行 lambda 限制）
## A-6：_sort_scores 仅在 query() 同步调用期间有效，禁止在 query() 内引入 await（会导致并发覆盖）
var _sort_scores: Dictionary = {}
func _cmp_score_desc(a: String, b: String) -> bool:
    return int(_sort_scores.get(a, 0)) > int(_sort_scores.get(b, 0))

## 单游戏打分
func _score_entry(e: IndexEntry, q: String) -> int:
    var s: int = 0
    ## 标题前缀 +100
    if e.title_lower.begins_with(q): s += 100
    ## 标题包含 +40
    elif e.title_lower.contains(q): s += 40
    ## 别名包含 +35
    for alias in e.aliases_lower:
        if alias.contains(q): s += 35; break
    ## 拼音全拼前缀 +30
    for pf in e.pinyin_full:
        if pf.begins_with(q): s += 30; break
    ## 拼音首字母前缀 +25
    for pi in e.pinyin_initials:
        if pi.begins_with(q): s += 25; break
    ## 标签命中 +15
    for tag in e.tags_lower:
        if tag.contains(q): s += 15; break
    return s

## 热搜词（editorial.json hot_queries，最多 6 条）
func hot_queries() -> Array[String]:
    var hq: Variant = _editorial.get("hot_queries", [])
    if not (hq is Array): return []
    var out: Array[String] = []
    for item in hq as Array:
        out.append(str(item))
        if out.size() >= 6: break
    return out
```

---

## 6. GameCard 扩展（size 参数）

在 `setup()` 中增加可选 `size` 参数，grid 模式固定宽度适配 CategoryPage 2 列网格：

```gdscript
## size: "card"（默认，继续游戏/为你推荐横滑）/ "grid"（分类页 2 列网格）
func setup(meta: GameMeta, record: Variant, size: String = "card") -> void:
    _gid = meta.id
    _name_label.text = meta.title
    _apply_badge(meta, record)
    _apply_style()
    if size == "grid":
        ## 2 列网格：宽度充满父容器（GridContainer 均分），固定高度
        size_flags_horizontal = Control.SIZE_EXPAND_FILL
        custom_minimum_size = Vector2(0, 140)
    ## "card" 模式保持原有横滑卡尺寸，不变
```

---

## 7. CategoryPage（shell/pages/category.tscn/.gd）

**场景结构**：
```
CategoryPage (Control/Page)
├── HSplitContainer
│   ├── CategoryList (VBoxContainer)       # 左侧分类栏，固定宽度 96px
│   └── RightPanel (VBoxContainer)
│       ├── FilterRow (HBoxContainer)      # 顶部筛选 chips
│       ├── SortDropdown (OptionButton, visible=false)  # 排序下拉 M2 解锁
│       ├── GridContainer (columns=2)      # 游戏网格（复用 GameCard size="grid"）
│       └── EmptyState (VBoxContainer)     # 空态（visible=false 时隐藏）
```

```gdscript
extends Page
## 分类页：左侧分类栏 + 顶部筛选 chips + 右侧 2 列网格
## session 内筛选状态自然保留（Tab 根页常驻，on_resume 不重置）

@export var card_scene: PackedScene  # GameCard.tscn

@onready var _cat_list: VBoxContainer = $HSplit/CategoryList
@onready var _grid: GridContainer = $HSplit/RightPanel/GridContainer
@onready var _empty_state: VBoxContainer = $HSplit/RightPanel/EmptyState
@onready var _empty_clear_btn: Button = $HSplit/RightPanel/EmptyState/ClearBtn

## 当前筛选状态
var _filter: Dictionary = {
    "category": "",        # "" = 全部
    "price_models": [],    # 空 = 不过滤
    "min_rating": 0.0,
    "is_new": false,
}

## 分类定义（id → 显示名）；与 GameMeta.category 枚举对齐（B-3：消除/益智分开映射）
## B-4：GameMeta.category 枚举为 puzzle/action/arcade/casual/roguelike（无独立"益智"id）
## 合并"消除/益智"为单按钮（id="puzzle"），保留"休闲"(id="casual")，避免两按钮同效果
const CATEGORIES: Array = [
    {"id": "", "label": "全部"},
    {"id": "puzzle", "label": "消除/益智"},
    {"id": "action", "label": "动作"},
    {"id": "arcade", "label": "街机"},
    {"id": "casual", "label": "休闲"},
    {"id": "roguelike", "label": "Roguelike"},
]

func _ready() -> void:
    _build_category_list()
    _build_filter_chips()
    _empty_clear_btn.pressed.connect(_on_clear_filter)

func on_enter(_data: Dictionary) -> void:
    _render_grid()

func on_resume() -> void:
    _render_grid()  # 从详情页返回后刷新（游戏状态可能变化）

## 构建左侧分类栏
func _build_category_list() -> void:
    for item in CATEGORIES:
        var btn := Button.new()
        btn.text = str(item["label"])
        btn.pressed.connect(_on_category_selected.bind(str(item["id"])))
        _cat_list.add_child(btn)

## 构建顶部筛选 chips
## P-4：「离线可玩」chip M1 无对应 filter 维度（GameMeta 无 offline 字段），visible=false 预留 M2
func _build_filter_chips() -> void:
    pass  # 实现：按 client-design §4.3 建 chip 按钮
    # chips：免费(price_models=["free","ad","iap"]) / 付费(["paid","trial"]) / 评分≥4.0(min_rating=4.0) / 新游(is_new=true)
    # 离线可玩 chip：visible=false，M2 补 GameMeta.offline 字段后解锁

## 分类选中
func _on_category_selected(cat_id: String) -> void:
    _filter["category"] = cat_id
    _render_grid()

## chip 切换
func _on_chip_toggled(chip_key: String, value: Variant) -> void:
    _filter[chip_key] = value
    _render_grid()

## 清除筛选（空态内联按钮）
func _on_clear_filter() -> void:
    _filter = {"category": "", "price_models": [], "min_rating": 0.0, "is_new": false}
    _render_grid()

## 渲染网格
func _render_grid() -> void:
    for c in _grid.get_children(): c.queue_free()
    var results := Registry.query(_filter)
    _empty_state.visible = results.is_empty()
    _grid.visible = not results.is_empty()
    for meta in results:
        var card := card_scene.instantiate() as GameCard
        card.setup(meta as GameMeta, DB.get_record((meta as GameMeta).id), "grid")
        card.card_pressed.connect(func(gid): Nav.push("res://shell/pages/detail.tscn", {"gid": gid}))
        _grid.add_child(card)
```

---

## 8. SearchPage（shell/pages/search.tscn/.gd）

**场景结构**：
```
SearchPage (Control/Page)
├── TopBar (HBoxContainer)
│   ├── SearchInput (LineEdit)
│   └── CancelBtn (Button)
├── DefaultView (VBoxContainer)       # 无输入时显示
│   ├── HotSection (VBoxContainer)    # 热搜词
│   └── HistorySection (VBoxContainer) # 历史行（含清空按钮）
├── ResultList (VBoxContainer)        # 有输入时显示
└── EmptyState (Label)                # 无结果时显示
```

```gdscript
extends Page
## 搜索页：输入去抖 150ms + 热搜 + 历史（去重置顶）+ 实时结果

@export var card_scene: PackedScene

@onready var _input: LineEdit = $TopBar/SearchInput
@onready var _cancel_btn: Button = $TopBar/CancelBtn
@onready var _default_view: VBoxContainer = $DefaultView
@onready var _result_list: VBoxContainer = $ResultList
@onready var _empty_state: Label = $EmptyState
@onready var _hot_row: HBoxContainer = $DefaultView/HotSection/HotRow
@onready var _hist_row: HBoxContainer = $DefaultView/HistorySection/HistRow
@onready var _hist_clear_btn: Button = $DefaultView/HistorySection/ClearBtn

var _searcher: Searcher = null
var _debounce_active: bool = false  # 去抖 flag

func _ready() -> void:
    _input.text_changed.connect(_on_text_changed)
    _cancel_btn.pressed.connect(func(): Nav.pop())
    _hist_clear_btn.pressed.connect(_on_clear_history)
    _searcher = get_node("/root/Main/Services/Searcher") as Searcher

func on_enter(_data: Dictionary) -> void:
    _input.grab_focus()
    _input.text = ""          ## on_enter = 新入口（push 进来），重置搜索状态
    _debounce_active = false
    _show_default()

## on_resume：Tab 切回时保持上次输入状态，仅刷新历史行
func on_resume() -> void:
    _render_history()         ## 历史可能被其他页面更新，需刷新

## 输入变化：取消旧 Timer，新建 150ms 去抖 Timer
func _on_text_changed(new_text: String) -> void:
    _debounce_active = false
    if new_text.strip_edges() == "":
        _show_default()
        return
    _debounce_active = true
    var timer := get_tree().create_timer(0.15)
    await timer.timeout
    if not _debounce_active: return  # 被更新的输入取消
    _debounce_active = false
    _run_search(new_text)

## 执行搜索并渲染结果
func _run_search(q: String) -> void:
    _default_view.visible = false
    for c in _result_list.get_children(): c.queue_free()
    if _searcher == null:
        push_warning("SearchPage: Searcher 未找到")
        return
    var gids := _searcher.query(q)
    if gids.is_empty():
        _result_list.visible = false
        _empty_state.text = "没有找到「%s」" % q
        _empty_state.visible = true
        return
    _empty_state.visible = false
    _result_list.visible = true
    for gid in gids:
        var meta: Variant = Registry.lookup(gid)
        if meta == null: continue
        var card := card_scene.instantiate() as GameCard
        card.setup(meta as GameMeta, DB.get_record(gid))
        card.card_pressed.connect(_on_result_pressed.bind(q))
        _result_list.add_child(card)

## 点击结果：导航到详情页 + 记录历史
func _on_result_pressed(gid: String, q: String) -> void:
    DB.add_search_history(q)
    Nav.push("res://shell/pages/detail.tscn", {"gid": gid})

## 无输入时默认视图（热搜 + 历史）
func _show_default() -> void:
    _result_list.visible = false
    _empty_state.visible = false
    _default_view.visible = true
    _render_default_view()

func _render_default_view() -> void:
    _render_hot_words()
    _render_history()

func _render_hot_words() -> void:
    for c in _hot_row.get_children(): c.queue_free()
    if _searcher == null: return
    for q in _searcher.hot_queries():
        var btn := Button.new()
        btn.text = q
        ## A-1：不可用单行多语句 lambda，改用 bind + 显式方法
        btn.pressed.connect(_on_hot_or_hist_pressed.bind(q))
        _hot_row.add_child(btn)

func _render_history() -> void:
    for c in _hist_row.get_children(): c.queue_free()
    for q in DB.get_search_history():
        var btn := Button.new()
        btn.text = q
        btn.pressed.connect(_on_hot_or_hist_pressed.bind(q))
        _hist_row.add_child(btn)

## 热搜/历史词点击（A-1：替代单行多语句 lambda）
func _on_hot_or_hist_pressed(q: String) -> void:
    _input.text = q
    _run_search(q)

func _on_clear_history() -> void:
    ## DB 无 clear_search_history，调用空写覆盖
    ## M1 临时方案：直接访问内部（设计上应补 DB.clear_search_history 接口）
    ## 本 CR 在 db.gd 补充 clear_search_history()
    DB.clear_search_history()
    _render_history()
```

> **DB 接口补充**：需在 db.gd 中新增 `clear_search_history()` 方法（清空 _search_history + 防抖写）。

---

## 9. ReviewEditor（shell/overlays/review_editor.tscn/.gd）

OverlayLayer 静态子节点，默认 visible=false。

**场景结构**：
```
ReviewEditor (Control)
├── Dim (ColorRect)                    # 半透明遮罩
└── Card (PanelContainer)
    └── VBox
        ├── TitleLabel (Label)         # "评价 {游戏名}"
        ├── StarRow (HBoxContainer)    # 5 颗星按钮
        ├── TextEdit (TextEdit)        # 最多 500 字
        ├── CharCount (Label)          # 实时字数 "xxx/500"
        └── BtnRow (HBoxContainer)
            ├── SubmitBtn (Button)
            └── CancelBtn (Button)
```

```gdscript
extends Control
## ReviewEditor 模态弹窗：写/改评价（FR-8）
## show_modal(gid, playtime) → 弹出；提交 → DB.put_review + EventBus.review_submitted + 关闭

const MAX_CHARS: int = 500
const MIN_PLAYTIME: float = 600.0

@onready var _title_lbl: Label = $Card/VBox/TitleLabel
@onready var _text_edit: TextEdit = $Card/VBox/TextEdit
@onready var _char_count: Label = $Card/VBox/CharCount
@onready var _submit_btn: Button = $Card/VBox/BtnRow/SubmitBtn
@onready var _cancel_btn: Button = $Card/VBox/BtnRow/CancelBtn
@onready var _dim: ColorRect = $Dim

var _gid: String = ""
var _playtime: float = 0.0
var _stars: int = 0

func _ready() -> void:
    visible = false
    _submit_btn.pressed.connect(_on_submit)
    _cancel_btn.pressed.connect(_on_cancel)
    _text_edit.text_changed.connect(_on_text_changed)

## 弹出编辑器；若已有评价则回填内容
func show_modal(gid: String, playtime: float) -> void:
    _gid = gid
    _playtime = playtime
    _stars = 0
    _text_edit.text = ""
    var meta: Variant = Registry.lookup(gid)
    _title_lbl.text = "评价 " + ((meta as GameMeta).title if meta != null else gid)
    ## 回填已有评价
    var existing: Variant = DB.get_review(gid)
    if existing is Dictionary:
        var ex := existing as Dictionary
        _stars = int(ex.get("stars", 0))
        _text_edit.text = str(ex.get("text", ""))
    _update_star_ui()
    _update_submit_state()
    _on_text_changed()
    visible = true

## 实时字数统计
func _on_text_changed() -> void:
    var txt := _text_edit.text
    if txt.length() > MAX_CHARS:
        _text_edit.text = txt.left(MAX_CHARS)
        _text_edit.set_caret_column(MAX_CHARS)
    _char_count.text = "%d/%d" % [_text_edit.text.length(), MAX_CHARS]
    _update_submit_state()

## 提交按钮状态：stars>0 且 playtime≥600s
func _update_submit_state() -> void:
    _submit_btn.disabled = (_stars <= 0 or _playtime < MIN_PLAYTIME)

func _on_star_pressed(n: int) -> void:
    _stars = n
    _update_star_ui()
    _update_submit_state()

func _update_star_ui() -> void:
    pass  # 更新 StarRow 各按钮样式（选中/未选中颜色，ThemeTokens.color("gold"/"star_off")）

## 提交：DB.put_review（覆盖语义）+ 信号 + 关闭
func _on_submit() -> void:
    var review := {
        "stars": _stars,
        "text": _text_edit.text,
        "playtime_at_review": _playtime,
        "created_at": int(Time.get_unix_time_from_system()),
        "status": "local",
    }
    DB.put_review(_gid, review)
    EventBus.review_submitted.emit(_gid)
    visible = false

func _on_cancel() -> void:
    visible = false
```

---

## 10. ReviewsPage（shell/pages/reviews.tscn/.gd）

从 DetailPage"全部评价→"入口（Nav.push，data={gid}）进入。

```gdscript
extends Page
## 评价列表页（FR-7）：头部平均星 + [写评价] + 列表（自己的评价置顶）

@onready var _back_btn: Button = $BackButton
@onready var _avg_label: Label = $VBox/Header/AvgLabel
@onready var _write_btn: Button = $VBox/Header/WriteBtn
@onready var _list_vbox: VBoxContainer = $VBox/ListView
@onready var _empty_state: Label = $VBox/EmptyState

var _gid: String = ""

func _ready() -> void:
    _back_btn.pressed.connect(func(): Nav.pop())
    _write_btn.pressed.connect(_on_write_review)
    EventBus.review_submitted.connect(_on_review_changed)
    EventBus.review_deleted.connect(_on_review_changed)

func on_enter(data: Dictionary) -> void:
    _gid = data.get("gid", "")
    _render()

func on_resume() -> void:
    _render()

func _on_write_review() -> void:
    var rec: Variant = DB.get_record(_gid)
    var playtime := float((rec as Dictionary).get("total_playtime", 0.0)) if rec is Dictionary else 0.0
    var editor := get_tree().root.find_child("ReviewEditor", true, false)
    if editor != null:
        editor.call("show_modal", _gid, playtime)

func _on_review_changed(_gid_changed: String) -> void:
    if _gid_changed == _gid: _render()

func _render() -> void:
    for c in _list_vbox.get_children(): c.queue_free()
    var review: Variant = DB.get_review(_gid)
    ## M1 本地只有 1 条评价（自己的）
    var rec: Variant = DB.get_record(_gid)
    var playtime := float((rec as Dictionary).get("total_playtime", 0.0)) if rec is Dictionary else 0.0
    var finish := int((rec as Dictionary).get("finish_count", 0)) if rec is Dictionary else 0
    ## B-2：[写评价] 按钮条件与结算卡对齐：playtime≥600s 且 finish_count≥2 才可用
    _write_btn.disabled = playtime < 600.0 or finish < 2
    if review == null:
        _empty_state.visible = true
        _list_vbox.visible = false
        _avg_label.text = "暂无评分"
        return
    _empty_state.visible = false
    _list_vbox.visible = true
    var rv := review as Dictionary
    ## P-3：明确文案为"你的评分"，避免误解为多人平均分
    _avg_label.text = "你的评分：★%d" % int(rv.get("stars", 0))
    ## 自己的评价置顶（M1 恒为唯一一条，带 [修改][删除]）
    _add_review_item(rv, true)

func _add_review_item(rv: Dictionary, is_own: bool) -> void:
    var item := VBoxContainer.new()
    var stars_lbl := Label.new()
    stars_lbl.text = "★" * int(rv.get("stars", 0))
    stars_lbl.add_theme_color_override("font_color", ThemeTokens.color("gold"))
    item.add_child(stars_lbl)
    var text_lbl := Label.new()
    text_lbl.text = str(rv.get("text", ""))
    text_lbl.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
    item.add_child(text_lbl)
    if is_own:
        var btn_row := HBoxContainer.new()
        var edit_btn := Button.new()
        edit_btn.text = "修改"
        edit_btn.pressed.connect(_on_write_review)
        var del_btn := Button.new()
        del_btn.text = "删除"
        ## B-1：删除前二次确认，防误触（复用 ConfirmBubble，已在 OverlayLayer 静态挂载）
        del_btn.pressed.connect(_on_delete_review_confirm)
        btn_row.add_child(edit_btn)
        btn_row.add_child(del_btn)
        item.add_child(btn_row)
    _list_vbox.add_child(item)

## B-1：删除确认弹窗（ConfirmBubble）
func _on_delete_review_confirm() -> void:
    var bubble := get_tree().root.find_child("ConfirmBubble", true, false)
    if bubble == null:
        _on_delete_review()  # 降级：找不到弹窗直接删
        return
    bubble.call("show_bubble", "确认删除这条评价？", Callable(self, "_on_delete_review"))

func _on_delete_review() -> void:
    DB.delete_review(_gid)
    EventBus.review_deleted.emit(_gid)
```

---

## 11. DetailPage 补充（detail.gd / detail.tscn）

在 ScrollContainer/VBox 末尾加"全部评价→"按钮节点（`ReviewsBtn: Button`）；on_enter 中绑定导航：

```gdscript
## detail.gd 新增
@onready var _reviews_btn: Button = $ScrollContainer/VBox/ReviewsBtn

func on_enter(data: Dictionary) -> void:
    # ... 原有逻辑不变 ...
    ## 绑定评价按钮（仅首次 on_enter 连接，避免重复）
    if not _reviews_btn.pressed.is_connected(_on_reviews_pressed):
        _reviews_btn.pressed.connect(_on_reviews_pressed)

func _on_reviews_pressed() -> void:
    Nav.push("res://shell/pages/reviews.tscn", {"gid": _gid})
```

---

## 12. ResultOverlay 接线（result_overlay.gd）

将 `_on_review()` 由 Toast 改为 ReviewEditor.show_modal：

```gdscript
## 替换原有 _on_review()
## P-1：playtime 应来自本局 result dict，而非 DB.total_playtime（防止累计时长绕过防刷）
## ResultOverlay 已在 show_card(gid, result) 时持有 result，新增 _result 成员变量缓存
func _on_review() -> void:
    var editor := get_tree().root.find_child("ReviewEditor", true, false)
    if editor != null:
        ## 使用本局 playtime（show_card 时从 result dict 写入 _result_playtime）
        editor.call("show_modal", _gid, _result_playtime)

## show_card 中需同时缓存本局 playtime（在 show_card 函数内补充赋值）：
## _result_playtime = float(result.get("playtime", 0.0))
## 并在类顶部声明：var _result_playtime: float = 0.0
```

---

## 13. Main.tscn + Nav 更新

- Services 节点下增加 `Searcher`（Node，挂 searcher.gd）
- OverlayLayer 下增加 `ReviewEditor`（Control，挂 review_editor.tscn，visible=false）
- `nav.gd` `_modal_dialog_open()` 扩展：

```gdscript
## A-3：OverlayLayer 路径在 T10 执行时实查确认；初始化时缓存引用避免每帧 get_node
## 缓存方式：在 Nav._ready() 中赋值 _overlay_layer
var _overlay_layer: Control = null

func _ready() -> void:
    ## 延迟到下一帧取，确保 Main 场景树完整
    call_deferred("_cache_overlay_layer")

func _cache_overlay_layer() -> void:
    ## 执行时按 main.tscn 实际路径确认（T10 Constraint 要求实查）
    _overlay_layer = get_node_or_null("/root/Main/App/OverlayLayer") as Control

func _modal_dialog_open() -> bool:
    if _overlay_layer == null: return false
    for child in _overlay_layer.get_children():
        if child.visible: return true
    return false
```

---

## 14. DB 补充接口汇总

| 方法 | 说明 |
|------|------|
| `delete_review(gid)` | 删除评价，强写落盘 |
| `clear_search_history()` | 清空历史数组，防抖写落盘 |

`clear_search_history()` 实现：
```gdscript
func clear_search_history() -> void:
    _search_history.clear()
    _debounce_write("search_history.json")
```

---

## 15. 不变行为清单（回归防护）

| # | WHEN | THEN 系统 SHALL |
|---|------|-----------------|
| RG-1~12 | 沿用 CR-3 全部条目 | 不破坏现有所有链路 |
| RG-13 | CategoryPage 筛选变化 | 网格即时更新，无崩溃；空态正常显示 |
| RG-14 | SearchPage 输入"tetra" | 150ms 后出现 TETRA NOVA；点击进详情，历史记录更新 |
| RG-15 | 同一词重复搜索 | 历史中该词置顶，无重复条目 |
| RG-16 | ReviewEditor 提交 | DB.get_review(gid) 存在且 status="local"；ReviewsPage 立即可见 |
| RG-17 | ResultOverlay [✎ 评价] playtime<600s | 按钮置灰，不弹 ReviewEditor |
| RG-18 | ResultOverlay [✎ 评价] playtime≥600s + finish_count≥2 | ReviewEditor.show_modal 被调用，注入本局 playtime（非 total） |
| RG-19 | DetailPage [全部评价→] | Nav.push ReviewsPage，gid 正确 |
| RG-20 | clear_search_history() 后正常退出 | 重启后 get_search_history() 返回空数组（A-5：clear 使用防抖写，正常退出可落盘） |

---

## 16. 正确性属性

- **Searcher 幂等性**：query(q) 无副作用，纯读内存索引，多次调用结果相同
- **评价一人一条**：put_review 按 gid 键覆盖，天然保证唯一性
- **删除安全性**：delete_review 删键后立即强写，无 dangling reference
- **搜索历史上限**：add_search_history 已保证 HISTORY_CAP=20，clear 后为空
- **筛选空状态完整性**：Registry.query 返回空数组时 CategoryPage 展示空态+清除按钮，不显示空白网格

---

## 17. 决策记录

- `GameCard.setup()` 加 size 参数：向后兼容，现有调用方不传 size 默认 "card"
- `_extract_initials()` M1 简化取首字符：M2 meta.json 改为存储单字全拼数组后重新实现完整缩写
- `is_new` 过滤 M1 占位：不过滤，M2 补 `published_at` 字段后实现
- `ReviewEditor` find_child 查找：与 result_overlay._toast() 保持一致的访问模式
- `_modal_dialog_open()` 遍历 OverlayLayer 所有子节点 visible：比逐一枚举更健壮，后续新增弹窗无需改 Nav

## 18. 修订记录

- 2026-08-25 初版（D1~D8 全采纳，基于工程实查结论）
- 2026-08-25 R2（三角色 Review 第二轮，3项修复）：
  - 中-P5：SearchPage 补 on_resume（Tab 切回保持输入状态，仅刷新历史行）；on_enter 重置输入
  - 低-B4：CATEGORIES 合并"消除/益智"为单按钮(id="puzzle")，消除"益智"/"休闲"同映射 "casual" 的重复
  - 低-A6：_sort_scores 加注释"禁止在 query() 内引入 await"
  - 高-A1：Searcher.query 单行 lambda → 具名 _cmp_score_desc + _sort_scores 成员；SearchPage 热搜/历史单行 lambda → _on_hot_or_hist_pressed
  - 高-A2：IndexEntry 数组字段 build_index 中显式初始化为新数组
  - 高-P1：ResultOverlay._on_review 改用本局 _result_playtime（新增成员变量），不再用 total_playtime
  - 中-B1：ReviewsPage [删除] 接入 ConfirmBubble 二次确认，降级兜底直接删
  - 中-B2：ReviewsPage [写评价] 条件补 finish_count≥2（与结算卡 REVIEW_RUNS=2 对齐）
  - 中-P3：ReviewsPage 平均星文案改为"你的评分：★N"
  - 中-A3：Nav._modal_dialog_open 改为 _ready 中 call_deferred 缓存 _overlay_layer，T10 执行时实查路径
  - 低-B3：CATEGORIES 中"益智"改映射 "casual"，消除/益智不再同映射 "puzzle"
  - 中-P4：_build_filter_chips 注释标注"离线可玩" chip M1 visible=false，M2 补 GameMeta.offline 字段
  - 中-P2：T5 执行时补断言验证 bind(q) 语义（gid/q 参数顺序）
  - 中-A4：ReviewsPage 信号生命周期问题标注低优 M2 优化
  - 低-A5：新增 RG-20（clear_search_history 持久化验证）
