# 设计：CR-5 首页完整化（Banner 轮播 + 热门榜 + 每日任务）

> 依据：`requirements.md`（Q1~Q6）+ `design_plan.md`（D1~D7 全采纳，2026-08-26）。
> 上游：client-design §4.2/§7/§8 + override §2/§5 + hybrid §3.2。
> 工程实查：ThemeTokens.GRADS 已有 banner 渐变 ✅；EventBus 无 tasks_updated（本 CR 新增）；DB.touch_daily 结构 `{date,plays}`（本 CR 扩展）；Recommender 无 charts()（本 CR 扩展）；home.tscn ScrollView/VBox 已有两区块可直接插入。

---

## 1. 文件与目录变更

```
nova-arcade/
├── core/
│   ├── event_bus.gd         # 新增 tasks_updated 信号
│   └── db.gd                # 新增 update_daily_tasks() / get_daily_tasks()
├── services/
│   ├── recommender.gd       # 扩展 charts(mode)
│   └── daily_task_service.gd  # 新增：每日任务逻辑
├── shell/
│   ├── components/
│   │   ├── section_banner.tscn/.gd   # 新增：Banner 轮播组件
│   │   ├── section_charts.tscn/.gd   # 新增：热门榜组件
│   │   └── section_daily.tscn/.gd    # 新增：每日任务组件
│   └── pages/
│       ├── home.gd / home.tscn        # 扩展：插入三区块 + 注入 DailyTaskService
└── shell/main.tscn           # Services 下挂 DailyTaskService 节点
```

---

## 2. EventBus 新增信号

```gdscript
## 每日任务进度更新（DailyTaskService 发）；tasks 为当前任务数组
signal tasks_updated(tasks: Array[Dictionary])
```

---

## 3. DB 扩展（core/db.gd）

每日任务进度存于同一 `_daily` dict（tasks_data 字段），与 touch_daily() 共用 daily.json；共存规则：touch_daily() 跨天重置改为**保留 tasks_data 字段**（新日期下任务进度由 DailyTaskService._load_or_reset() 判定重置，DB 层不擅自清空业务字段）；_debounce_write("daily.json") 序列化整个 _daily，两路写入互不覆盖。

```gdscript
## 取每日任务进度；不存在/格式异常时返回空 dict
func get_daily_tasks() -> Dictionary:
    var d := get_daily()
    var raw: Variant = d.get("tasks_data")
    return raw if raw is Dictionary else {}

## 写入每日任务进度（防抖写，与 daily 共用同一文件）
func update_daily_tasks(tasks_data: Dictionary) -> void:
    var d := _daily
    ## 确保日期 key 存在（touch_daily 可能尚未调用）
    if not d.has("date"):
        d["date"] = Time.get_date_string_from_system()
    d["tasks_data"] = tasks_data
    _daily = d
    _debounce_write("daily.json")
```

touch_daily() 同步修订（跨天重置保留 tasks_data）：

```gdscript
## 触达每日数据：跨天重置 {date, plays:0}，防抖落盘
func touch_daily() -> void:
    var today := Time.get_date_string_from_system()
    if str(_daily.get("date", "")) != today:
        ## 跨天重置：保留 tasks_data（§3 共存规则；任务进度由 DailyTaskService._load_or_reset() 判定）
        var kept_tasks: Variant = _daily.get("tasks_data")
        _daily = {"date": today, "plays": 0}
        if kept_tasks is Dictionary:
            _daily["tasks_data"] = kept_tasks
    else:
        _daily["plays"] = int(_daily.get("plays", 0)) + 1
    _debounce_write("daily.json")
```

---

## 4. Recommender.charts() 扩展（services/recommender.gd）

```gdscript
## 热门榜（mode: "all"=综合 / "new"=新游 / "rated"=好评）
## M1 单款游戏时三模式结果相同；M2 多游戏后自动生效
## 返回 [{"meta": GameMeta, "record": Variant, "rank": int, "players": int}]
func charts(mode: String) -> Array[Dictionary]:
    var all_metas: Array[GameMeta] = Registry.all()
    var rows: Array[Dictionary] = []
    for meta: GameMeta in all_metas:
        var rec: Variant = DB.get_record(meta.id)
        var players: int = _calc_players(meta.id, rec)
        rows.append({"meta": meta, "record": rec, "rank": 0, "players": players})

    match mode:
        "new":
            ## 新游：version 字符串倒序（M2 接入真实 version 后需复核排序语义：空 version 当前排最前）
            rows.sort_custom(_cmp_version_desc)
        "rated":
            ## 好评：本地评价均分倒序（无评价的排末尾）
            rows.sort_custom(_cmp_rating_desc)
        _:
            ## 综合：hot_score = sessions 倒序（M1 简化，不归一化）
            rows.sort_custom(_cmp_sessions_desc)

    ## 取前 10，写入 rank
    var result: Array[Dictionary] = rows.slice(0, 10)
    for i in result.size():
        result[i]["rank"] = i + 1
    return result

## 玩过人数：本地 sessions + 1万基准（Q6：M3 获取云端数据后叠加）
func _calc_players(gid: String, rec: Variant) -> int:
    var sessions: int = 0
    if rec is Dictionary:
        sessions = int((rec as Dictionary).get("sessions", 0))
    return sessions + 10000

## 比较器：version 字符串倒序
func _cmp_version_desc(a: Dictionary, b: Dictionary) -> bool:
    return (a["meta"] as GameMeta).version > (b["meta"] as GameMeta).version

## 比较器：本地评价均分倒序（无评价=0，排末尾）
func _cmp_rating_desc(a: Dictionary, b: Dictionary) -> bool:
    return _local_rating((a["meta"] as GameMeta).id) > _local_rating((b["meta"] as GameMeta).id)

## 比较器：sessions 倒序
func _cmp_sessions_desc(a: Dictionary, b: Dictionary) -> bool:
    return a["players"] > b["players"]

## 本地评价均分（0~5.0）
func _local_rating(gid: String) -> float:
    var rv: Variant = DB.get_review(gid)
    if rv is Dictionary:
        return float((rv as Dictionary).get("stars", 0))
    return 0.0
```

---

## 5. DailyTaskService（services/daily_task_service.gd）

```gdscript
class_name DailyTaskService extends Node
## 每日任务逻辑服务（FR-5）
## 非 Autoload；挂 Main/Services 节点下

signal tasks_updated(tasks: Array[Dictionary])

## 任务定义（硬编码，M2 可扩展为后端下发）
const TASK_DEFS: Array[Dictionary] = [
    {"id": "launch", "label": "今日启动游戏", "target": 1},
    {"id": "finish", "label": "今日完成 1 局", "target": 1},
    {"id": "playtime", "label": "今日累计游玩 5 分钟", "target": 300},
]

## 当日进度缓存 {task_id: progress_value}
var _progress: Dictionary = {}
## 今日已领奖
var _reward_claimed: bool = false

func _ready() -> void:
    _load_or_reset()
    EventBus.game_launched.connect(_on_game_launched)
    EventBus.game_finished.connect(_on_game_finished)

## 加载/重置每日进度
func _load_or_reset() -> void:
    var today := Time.get_date_string_from_system()
    var daily := DB.get_daily()
    if str(daily.get("date", "")) != today:
        ## 跨天：重置
        _progress = {}
        _reward_claimed = false
        _save()
    else:
        var saved := DB.get_daily_tasks()
        _progress = saved.get("progress", {}) if saved is Dictionary else {}
        _reward_claimed = bool(saved.get("reward_claimed", false)) if saved is Dictionary else false

## 获取当前任务状态（供 SectionDaily 渲染）
func get_tasks() -> Array[Dictionary]:
    var result: Array[Dictionary] = []
    for def in TASK_DEFS:
        var progress: int = int(_progress.get(str(def["id"]), 0))
        result.append({
            "id": def["id"],
            "label": str(def["label"]),
            "done": progress >= int(def["target"]),
            "progress": progress,
            "target": int(def["target"]),
        })
    return result

func _on_game_launched(_gid: String) -> void:
    _update_task("launch", 1, true)  ## 只需 1 次，设为至少=1

func _on_game_finished(_gid: String, result: Dictionary) -> void:
    _update_task("finish", 1, true)
    var pt: float = float(result.get("playtime", 0.0))
    var cur_pt: int = int(_progress.get("playtime", 0))
    _update_task("playtime", cur_pt + int(pt), false)

## 更新任务进度（cap_at_target=true 时最多写到 target，防止重复触发）
func _update_task(task_id: String, new_val: int, cap_at_target: bool) -> void:
    var def: Dictionary = _find_def(task_id)
    if def.is_empty(): return
    var target: int = int(def["target"])
    var val: int = mini(new_val, target) if cap_at_target else new_val
    _progress[task_id] = val
    _save()
    tasks_updated.emit(get_tasks())
    ## 全部完成 + 未领奖 → Toast + 标记
    if not _reward_claimed and _all_done():
        _reward_claimed = true
        _save()
        _show_reward_toast()

func _all_done() -> bool:
    for def in TASK_DEFS:
        if int(_progress.get(str(def["id"]), 0)) < int(def["target"]):
            return false
    return true

func _find_def(task_id: String) -> Dictionary:
    for d in TASK_DEFS:
        if str(d["id"]) == task_id: return d
    return {}

func _save() -> void:
    DB.update_daily_tasks({"progress": _progress, "reward_claimed": _reward_claimed})

func _show_reward_toast() -> void:
    var layer := get_tree().root.find_child("ToastLayer", true, false)
    if layer != null:
        layer.call("show_msg", "今日任务全部完成 🎉")
```

---

## 6. SectionBanner（shell/components/section_banner.tscn/.gd）

**场景结构**：
```
SectionBanner (Control, min_height=200)
├── BannerRect (PanelContainer, 全屏锚点)  # 背景色块/图片
│   └── TitleLabel (Label, 居中)           # 游戏标题文字
└── DotsRow (HBoxContainer, bottom-center)  # 指示点
    └── [Dot1, Dot2, …]（Button，动态生成）
```

```gdscript
class_name SectionBanner extends Control
## Banner 轮播组件（FR-2）：淡入淡出，4s Timer，指示点

@onready var _banner_rect: PanelContainer = $BannerRect
@onready var _title_lbl: Label = $BannerRect/TitleLabel
@onready var _dots_row: HBoxContainer = $DotsRow

var _items: Array[Dictionary] = []   ## banner[] 条目
var _current: int = 0
var _timer: SceneTreeTimer = null

func _ready() -> void:
    visible = false
    _load_banners()

func _load_banners() -> void:
    ## 幂等：重载前先清空（§12）；旧 timer 由 SceneTreeTimer 一次性特性自然失效，_timer 引用置空
    _items.clear()
    _timer = null
    var editorial: Dictionary = Registry.get_editorial()
    var raw: Variant = editorial.get("banner", [])
    if not (raw is Array):
        return
    for item in raw as Array:
        if item is Dictionary:
            _items.append(item as Dictionary)
    if _items.is_empty():
        visible = false
        return
    visible = true
    _build_dots()
    _show(0)
    if _items.size() > 1:
        _start_timer()

func _build_dots() -> void:
    for c in _dots_row.get_children():
        c.queue_free()
    for i in _items.size():
        var dot := Button.new()
        dot.custom_minimum_size = Vector2(5, 5)
        dot.flat = true
        _dots_row.add_child(dot)
        ## D4（D-3）：不在此处连接 pressed，点击 Banner 整体区域响应跳转
    _update_dots()

func _update_dots() -> void:
    for i in _dots_row.get_child_count():
        var dot: Button = _dots_row.get_child(i) as Button
        if dot == null: continue
        dot.custom_minimum_size = Vector2(14 if i == _current else 5, 5)

func _show(idx: int) -> void:
    _current = idx
    var item: Dictionary = _items[idx]
    ## image_path 为空时用渐变色块
    var img_path: String = str(item.get("image_path", ""))
    if img_path != "" and ResourceLoader.exists(img_path):
        pass  ## TODO M2: 加载 Texture
    ## 渐变色块（banner grad token → GradientTexture2D）
    var tex := GradientTexture2D.new()
    tex.gradient = ThemeTokens.grad("banner")
    _banner_rect.add_theme_stylebox_override("panel", StyleBoxEmpty.new())
    _banner_rect.texture = tex
    _title_lbl.text = str(item.get("title", ""))
    _title_lbl.add_theme_color_override("font_color", ThemeTokens.color("ink"))
    _update_dots()

func _start_timer() -> void:
    _timer = get_tree().create_timer(4.0)
    _timer.timeout.connect(_on_timer_tick)

func _on_timer_tick() -> void:
    var next: int = (_current + 1) % _items.size()
    ## 淡入淡出 0.3s（A-1）
    var tw := create_tween()
    tw.tween_property(_banner_rect, "modulate:a", 0.0, 0.15)
    tw.tween_callback(_show.bind(next))
    tw.tween_property(_banner_rect, "modulate:a", 1.0, 0.15)
    ## 循环
    _start_timer()

func _gui_input(event: InputEvent) -> void:
    if event is InputEventMouseButton:
        var e := event as InputEventMouseButton
        if e.button_index == MOUSE_BUTTON_LEFT and e.pressed:
            var gid: String = str(_items[_current].get("target_gid", ""))
            if gid != "":
                Nav.push("res://shell/pages/detail.tscn", {"gid": gid})
```

---

## 7. SectionCharts（shell/components/section_charts.tscn/.gd）

**场景结构**：
```
SectionCharts (VBoxContainer)
├── Header (HBoxContainer)
│   ├── TitleLabel (Label, "热门榜")
│   └── TabRow (HBoxContainer)           # 综合/新游/好评 三 Tab
├── ListView (VBoxContainer)             # 列表行（动态生成）
└── EmptyState (Label, "暂无数据")
```

```gdscript
class_name SectionCharts extends VBoxContainer
## 热门榜区块（FR-4）：三 Tab 即时切换，列表行按排名展示

@onready var _tab_row: HBoxContainer = $Header/TabRow
@onready var _list_vbox: VBoxContainer = $ListView
@onready var _empty_state: Label = $EmptyState

const MODES: Array[String] = ["all", "new", "rated"]
const MODE_LABELS: Array[String] = ["综合", "新游", "好评"]
var _current_mode: String = "all"
var _recommender: Recommender = null

func set_recommender(r: Recommender) -> void:
    _recommender = r
    if is_node_ready():
        _render()

func _ready() -> void:
    _build_tabs()
    _render()

func _build_tabs() -> void:
    for i in MODES.size():
        var btn := Button.new()
        btn.text = MODE_LABELS[i]
        btn.pressed.connect(_on_tab_pressed.bind(MODES[i]))
        _tab_row.add_child(btn)

func _on_tab_pressed(mode: String) -> void:
    _current_mode = mode
    _render()

func _render() -> void:
    for c in _list_vbox.get_children():
        c.queue_free()
    if _recommender == null:
        _empty_state.visible = true
        return
    var rows := _recommender.charts(_current_mode)
    _empty_state.visible = rows.is_empty()
    _list_vbox.visible = not rows.is_empty()
    for row in rows:
        _list_vbox.add_child(_make_row(row))

func _make_row(row: Dictionary) -> HBoxContainer:
    var meta := row["meta"] as GameMeta
    var rank: int = int(row.get("rank", 0))
    var players: int = int(row.get("players", 10000))

    var hbox := HBoxContainer.new()
    ## 排名数字
    var rank_lbl := Label.new()
    rank_lbl.text = str(rank)
    rank_lbl.custom_minimum_size = Vector2(40, 0)
    var rank_color: String = "rank1" if rank == 1 else ("rank2" if rank == 2 else ("rank3" if rank == 3 else "ink2"))
    rank_lbl.add_theme_color_override("font_color", ThemeTokens.color(rank_color))
    hbox.add_child(rank_lbl)
    ## 图标（meta.icon；空路径或资源不存在时隐藏）
    var icon_rect := TextureRect.new()
    if meta.icon != "" and ResourceLoader.exists(meta.icon):
        icon_rect.texture = load(meta.icon)
    else:
        icon_rect.visible = false
    icon_rect.custom_minimum_size = Vector2(36, 36)
    icon_rect.expand_mode = TextureRect.EXPAND_FIT_WIDTH_PROPORTIONAL
    hbox.add_child(icon_rect)
    ## 游戏名
    var name_lbl := Label.new()
    name_lbl.text = meta.title
    name_lbl.size_flags_horizontal = Control.SIZE_EXPAND_FILL
    name_lbl.add_theme_color_override("font_color", ThemeTokens.color("ink"))
    hbox.add_child(name_lbl)
    ## 评分（有评价才显示）
    var rv: Variant = DB.get_review(meta.id)
    if rv is Dictionary:
        var stars_lbl := Label.new()
        stars_lbl.text = "★%.1f" % float((rv as Dictionary).get("stars", 0))
        stars_lbl.add_theme_color_override("font_color", ThemeTokens.color("gold"))
        hbox.add_child(stars_lbl)
    ## 人数（A-2：base=10000+sessions，≥10000 显示"X.X万+"）
    var players_lbl := Label.new()
    players_lbl.text = _fmt_players(players)
    players_lbl.add_theme_color_override("font_color", ThemeTokens.color("ink2"))
    hbox.add_child(players_lbl)
    ## 点击
    hbox.gui_input.connect(_on_row_input.bind(meta.id))
    return hbox

func _fmt_players(n: int) -> String:
    if n >= 10000:
        return "%.1f万+" % (float(n) / 10000.0)
    return str(n) + "人"

func _on_row_input(event: InputEvent, gid: String) -> void:
    if event is InputEventMouseButton:
        var e := event as InputEventMouseButton
        if e.button_index == MOUSE_BUTTON_LEFT and e.pressed:
            Nav.push("res://shell/pages/detail.tscn", {"gid": gid})
```

---

## 8. SectionDaily（shell/components/section_daily.tscn/.gd）

**场景结构**：
```
SectionDaily (VBoxContainer)
├── Header (HBoxContainer)
│   ├── TitleLabel (Label, "每日任务")
│   └── ProgressLabel (Label, "0/3")
├── TaskList (VBoxContainer)            # 任务行（动态生成）
└── CompletedLabel (Label, "🎉 今日已全部完成", visible=false)
```

```gdscript
class_name SectionDaily extends VBoxContainer
## 每日任务区块（FR-7）

@onready var _progress_lbl: Label = $Header/ProgressLabel
@onready var _task_list: VBoxContainer = $TaskList
@onready var _completed_lbl: Label = $CompletedLabel

var _daily_svc: DailyTaskService = null

func set_service(svc: DailyTaskService) -> void:
    _daily_svc = svc
    _daily_svc.tasks_updated.connect(_on_tasks_updated)
    _render(_daily_svc.get_tasks())

func _on_tasks_updated(tasks: Array[Dictionary]) -> void:
    _render(tasks)

func _render(tasks: Array[Dictionary]) -> void:
    for c in _task_list.get_children():
        c.queue_free()
    var done_count: int = 0
    for task in tasks:
        if bool(task.get("done", false)):
            done_count += 1
        _task_list.add_child(_make_task_row(task))
    _progress_lbl.text = "%d/%d" % [done_count, tasks.size()]
    _completed_lbl.visible = done_count == tasks.size() and tasks.size() > 0

func _make_task_row(task: Dictionary) -> HBoxContainer:
    var hbox := HBoxContainer.new()
    var check := Label.new()
    check.text = "✓" if bool(task.get("done", false)) else "○"
    check.add_theme_color_override("font_color",
        ThemeTokens.color("ok") if bool(task.get("done", false)) else ThemeTokens.color("ink3"))
    hbox.add_child(check)
    var lbl := Label.new()
    lbl.text = str(task.get("label", ""))
    lbl.add_theme_color_override("font_color", ThemeTokens.color("ink"))
    lbl.size_flags_horizontal = Control.SIZE_EXPAND_FILL
    hbox.add_child(lbl)
    return hbox
```

---

## 9. home.gd / home.tscn 扩展

home.tscn 在 ScrollView/VBox 中新增三个区块（顺序：Banner → 继续游戏 → 为你推荐 → 热门榜 → 每日任务）：

```gdscript
## home.gd 新增（在现有 _ready 基础上追加）

@onready var _section_banner: SectionBanner = $ScrollView/VBox/SectionBanner
@onready var _section_charts: SectionCharts = $ScrollView/VBox/SectionCharts
@onready var _section_daily: SectionDaily = $ScrollView/VBox/SectionDaily

var _daily_svc: DailyTaskService = null

func _ready() -> void:
    ## … 原有逻辑 …
    _daily_svc = get_node_or_null("/root/Main/Services/DailyTaskService") as DailyTaskService
    if _daily_svc != null:
        _section_daily.set_service(_daily_svc)
    _section_charts.set_recommender(_recommender)

## on_enter / on_resume 保持不变（三区块各自内部驱动，无需 home.gd 主动刷新）
```

---

## 10. Main.tscn + EventBus 变更

- Main.tscn：Services 节点下新增 `DailyTaskService`（Node，挂 daily_task_service.gd）
- event_bus.gd：新增 `signal tasks_updated(tasks: Array[Dictionary])`
  - 注：DailyTaskService 直接 `tasks_updated.emit()` 自身信号，不通过 EventBus；EventBus 的 tasks_updated 预留给未来 M2 跨系统使用，本 CR 暂不接线 EventBus

---

## 11. 不变行为清单（回归防护）

| # | WHEN | THEN 系统 SHALL |
|---|------|-----------------|
| RG-1~20 | 沿用 CR-4 全部条目 | 不破坏现有所有链路 |
| RG-21 | banner[] 为空 | SectionBanner.visible=false，首页无空白区块 |
| RG-22 | banner[] 含 1 条 | SectionBanner 可见，4s Timer 不启动（单张不轮播） |
| RG-23 | charts("all") 调用 | 返回非空数组，rank 从 1 开始，players≥10000 |
| RG-24 | 每日任务启动一局 | "launch" 任务 done=true，tasks_updated 信号发出 |
| RG-25 | 三任务均完成 | Toast 弹出，reward_claimed=true 写入 DB |
| RG-26 | 重启后 reward_claimed=true | Toast 不重复弹出 |
| RG-27 | 跨自然日重启 | 所有任务进度重置，reward_claimed=false |
| RG-28 | home.gd on_resume | 继续游戏/为你推荐区块正常刷新（不破坏 CR-3/CR-4 逻辑） |
| RG-29 | Banner 条目 target_gid="" | 点击 BannerRect 不响应（不 push、不崩溃） |
| RG-30 | charts() 返回空数组 | SectionCharts 显示 EmptyState，列表区隐藏 |

---

## 12. 正确性属性

- **Banner 幂等性**：多次调用 _load_banners() 不重复添加条目（调用前已 clear）
- **每日任务防重**：cap_at_target=true 防止同一任务多次完成；reward_claimed=true 防止 Toast 重复
- **热门榜无副作用**：charts() 纯读 DB + Registry，不写任何状态

---

## 13. 修订记录

- 2026-08-26 初版（D1~D7 全采纳，基于工程实查）
- 2026-08-27 三角色 review 修复：B-1/B-2/B-3、P-1/P-2/P-3、A-1~A-5（§3 touch_daily 共存规则、§4 new 模式排序注记、§6 _load_banners 幂等、§7 set_recommender 刷新 + 图标节点、§11 RG-29/RG-30）
