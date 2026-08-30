# 设计：CR-7 M2 盒子级能力收尾（成就墙 + 本地推荐画像）

> 依据：`requirements.md`（FR1~FR6）+ `design_plan.md`（Q1~Q3 Answer 已确认）。
> 上游：client-design §4.9 + design §10 M2 + override §2/§7。
> 工程实查：event_bus 有 `achievement_unlocked(def, points)` ✅；Registry.ach_def(aid) 跨游戏查询 ✅；db.gd `_achievements` + `unlock(aid)` 强写 ✅；Recommender.for_you() 当前字母排序 ✅；home.gd 已消费 for_you() ✅；library.gd CR-1 骨架 ✅。

---

## 修订记录

| 日期 | 变更 | 触发源 |
|------|------|--------|
| 2026-08-28 | 初稿 | design_plan 确认 |
| 2026-08-28 | 多角色review修复（B1/B2/B3/P1/P2/P3/P4/P5/A1/A2/A3/A5） | 三角色review汇总

---

## 1. 文件与目录变更

```
nova-arcade/
├── services/
│   ├── achievement_engine.gd      # 新增：成就判定引擎（≤150行）
│   └── recommender.gd             # 扩展：for_you() 画像加权（扩展后≤150行）
├── core/
│   └── registry.gd                # 修改：ach_def() 扩展全局成就查询
├── data/
│   ├── achievements.json          # 新增：全局成就定义
│   └── editorial.json             # 可选：补充推荐位配置
├── games/
│   ├── magic_tower/meta.json      # 修改：补 achievements[]（3个）
│   ├── game_2048/meta.json       # 修改：补 achievements[]（3个）
│   └── snake/meta.json           # 修改：补 achievements[]（3个）
├── shell/
│   ├── main.tscn                  # 修改：Services 下新增 AchievementEngine 节点
│   └── pages/
│       ├── library.gd             # 扩展：成就墙 UI（拆为 section_achievement_wall.gd）
│       └── library.tscn           # 修改：场景结构适配
├── shell/components/
│   └── section_achievement_wall.gd  # 新增：成就墙组件（80~150行）
└── tools/
    ├── test_achievement.tscn       # 新增：Engine headless测试（≥10断言）
    ├── test_achievement.gd         # 新增
    ├── test_recommender_profile.tscn # 新增：推荐画像测试（≥10断言）
    └── test_recommender_profile.gd   # 新增
```

---

## 2. AchievementEngine（services/achievement_engine.gd）

### 2.1 架构

- **节点类型**：`Node`，非单例，挂 `Main/Services/AchievementEngine`
- **信号订阅**：
  - `EventBus.game_finished(gid, result)` → 游戏级成就判定
  - `EventBus.review_submitted(gid)` → 好评人全局成就判定
- **信号发射**：复用 `EventBus.achievement_unlocked(def, points)`（零新增）
- **幂等保证**：`db.gd.unlock(aid)` 内部 `has(aid)` 检查，Engine 仅调用 unlock() → emit

### 2.2 类签名

```gdscript
class_name AchievementEngine extends Node
## 成就判定引擎（CR-7 FR-1）
## 挂 Main/Services 节点下，非单例。
## 订阅 game_finished / review_submitted，evaluate() 统一判定游戏级+全局成就。

const GLOBAL_DEFS: Array[Dictionary] = [
    {"id": "global_collector", "name": "收藏家", "desc": "拥有≥5款游戏记录", "points": 30},
    {"id": "global_reviewer", "name": "好评人", "desc": "提交≥3条评价", "points": 20},
    {"id": "global_marathon", "name": "马拉松", "desc": "累计游玩≥10小时", "points": 50},
]

func _enter_tree() -> void:
    EventBus.game_finished.connect(_on_game_finished)
    EventBus.review_submitted.connect(_on_review_submitted)

## 游戏结束事件：判定 result.achievements[] 中的游戏级成就
func _on_game_finished(gid: String, result: Dictionary) -> void:
    var aids: Array = result.get("achievements", [])
    for aid in aids:
        _try_unlock(str(aid))

## 评价提交事件：判定好评人全局成就
func _on_review_submitted(gid: String) -> void:
    _evaluate_global()

## 尝试解锁某成就（幂等）
func _try_unlock(aid: String) -> void:
    if DB.get_achievements().has(aid):
        return  # 已解锁，跳过
    var def: Variant = Registry.ach_def(aid)
    if def == null:
        return  # 定义不存在，跳过
    DB.unlock(aid)
    EventBus.achievement_unlocked.emit(def as Dictionary, int((def as Dictionary).get("points", 0)))

## 判定全局成就（收藏家/好评人/马拉松）
func _evaluate_global() -> void:
    for gdef in GLOBAL_DEFS:
        var aid: String = str(gdef["id"])
        if DB.get_achievements().has(aid):
            continue
        if _check_global(gdef):
            EventBus.achievement_unlocked.emit(gdef, int(gdef["points"]))

## 全局成就条件检查
func _check_global(gdef: Dictionary) -> bool:
    match gdef["id"]:
        "global_collector":
            return DB.list_records().size() >= maxi(Registry.all().size(), 1)
        "global_reviewer":
            return _count_reviews() >= 3
        "global_marathon":
            return _total_playtime_hours() >= 10.0
    return false

## 统计有效评价数
func _count_reviews() -> int:
    var count: int = 0
    for gid: String in DB.list_records():
        if DB.get_review(gid) != null:
            count += 1
    return count

## 累计游玩时长（小时）
func _total_playtime_hours() -> float:
    var total: float = 0.0
    for gid: String in DB.list_records():
        var rec: Variant = DB.get_record(gid)
        if rec is Dictionary:
            total += float((rec as Dictionary).get("total_playtime", 0))
    return total / 3600.0
```

### 2.3 与 result_overlay.gd 的关系

现有 result_overlay.gd 在 `_fill_achievements()` 内做 `DB.unlock(aid)` + 显示 UI，但**不 emit EventBus.achievement_unlocked**。本 CR 两种方案：

| 方案 | 优点 | 缺点 |
|------|------|------|
| A: Engine 订阅 game_finished，result_overlay 保留 DB.unlock | 职责分离清晰 | result_overlay 与 Engine 各自调 unlock() → 需确保幂等 |
| B: result_overlay emit EventBus.achievement_unlocked，Engine 仅监听此信号 | 单发射源 | result_overlay 需改代码 |

**推荐方案 A**：Engine 订阅 `game_finished`，result_overlay 保留现有逻辑。db.gd.unlock() 已有 `has(aid)` 幂等保护，两者调 unlock() 不冲突。Engine emit `achievement_unlocked` → Toast；result_overlay 显示结算卡内成就列表。

---

## 3. 数据层变更

### 3.1 data/achievements.json（全局成就定义）

```json
[
    {"id": "global_collector", "name": "收藏家", "desc": "拥有≥5款游戏记录", "points": 30},
    {"id": "global_reviewer", "name": "好评人", "desc": "提交≥3条评价", "points": 20},
    {"id": "global_marathon", "name": "马拉松", "desc": "累计游玩≥10小时", "points": 50}
]
```

### 3.2 Registry.ach_def(aid) 扩展

当前实现仅遍历 `all()`（games/*/meta.json）。扩展后增加全局成就查询：

```gdscript
## 修改前
func ach_def(aid: String) -> Variant:
    for m in all():
        for a in m.achievements:
            if str(a.get("id", "")) == aid:
                return a
    return null

## 修改后：增加全局成就字典缓存
var _global_achievements: Dictionary = {}

## reload() 中新增加载 data/achievements.json
# ... existing reload code ...
_global_achievements.clear()
if FileAccess.file_exists("res://data/achievements.json"):
    var raw := FileAccess.get_file_as_string("res://data/achievements.json")
    var ja := JSON.parse_string(raw)
    if ja is Array:
        for ga in ja:
            if ga is Dictionary and ga.has("id"):
                _global_achievements[str(ga["id"])] = ga
            else:
                push_warning("Registry: 全局成就条目缺 id 字段，跳过", ga)
    elif ja != null:
        push_warning("Registry: achievements.json 非 Array 格式（%s），保留旧缓存" % type_string(ja))

## ach_def 扩展
func ach_def(aid: String) -> Variant:
    # 先查全局成就
    if _global_achievements.has(aid):
        return _global_achievements[aid]
    # 再查游戏级成就
    for m in all():
        for a in m.achievements:
            if str(a.get("id", "")) == aid:
                return a
    return null
```

### 3.3 三款新游戏补 achievements[]

**magic_tower/meta.json：**
```json
"achievements": [
    {"id": "mt_floor50", "name": "勇攀高峰", "desc": "到达第50层", "points": 15},
    {"id": "mt_boss_slayer", "name": "屠龙者", "desc": "击败最终Boss", "points": 25}
]
```

**game_2048/meta.json：**
```json
"achievements": [
    {"id": "g2048_tile512", "name": "初露锋芒", "desc": "合成512方块", "points": 10},
    {"id": "g2048_tile2048", "name": "终极目标", "desc": "合成2048方块", "points": 30}
]
```

**snake/meta.json：**
```json
"achievements": [
    {"id": "sn_eat50", "name": "大胃王", "desc": "吃到50个食物", "points": 10},
    {"id": "sn_length30", "name": "巨蛇", "desc": "身长达到30节", "points": 20}
]
```

### 3.4 db.gd 成就记录（已就位，零改动）

现有 `_achievements: Dictionary` + `unlock(aid)` + `get_achievements()` 已满足需求：
- unlock() 幂等（has(aid) 检查）
- 强写落盘（`_io.write_json("achievements.json", ...)`）
- 存储格式：`{"aid": {"unlocked_at": unix_timestamp}}`

---

## 4. 成就墙 UI（library.gd + section_achievement_wall.gd）

### 4.1 library.gd 扩展

当前 CR-1 骨架仅 `on_enter(print)`。扩展后结构：

```gdscript
extends Page
## 我的页（CR-7 FR-4）：头部信息 + 成就墙区块
@onready var _header: VBoxContainer = $RootVBox/Header
@onready var _ach_wall_root: Control = $RootVBox/AchievementWall

var _section_ach: SectionAchievementWall = null

func on_enter(data: Dictionary) -> void:
    _render_header()
    if _section_ach:
        _section_ach.refresh()

func _render_header() -> void:
    # 总成就点 = Σ 已解锁成就 points
    var total_points: int = 0
    for aid: String in DB.get_achievements():
        var def: Variant = Registry.ach_def(aid)
        if def is Dictionary:
            total_points += int((def as Dictionary).get("points", 0))
    # 累计时长（小时）
    var total_hours: float = _calc_total_hours()
    # ... 渲染到 _header Label ...

func _calc_total_hours() -> float:
    # B-2确认：launcher_util.finish_record() 写入 total_playtime（累加式）
    # "total_playtime": int_field(gid, "total_playtime") + int(result.get("playtime", 0))
    var total: float = 0.0
    for gid: String in DB.list_records():
        var rec: Variant = DB.get_record(gid)
        if rec is Dictionary:
            total += float((rec as Dictionary).get("total_playtime", 0))
    return total / 3600.0
```

### 4.2 section_achievement_wall.gd（独立组件，纯脚本 class_name）

**A-5修复：** library.tscn 预建 AchievementWall 节点并挂载此脚本（非运行时 instantiate），避免 @onready 引用报错。

```gdscript
class_name SectionAchievementWall extends Control
## 成就墙组件（CR-7 FR-4）：总进度条 + 按游戏分组 + 全局成就区
## 行数目标：80~150行

func refresh() -> void:
    _clear_children()
    _render_progress_bar()
    _render_by_game_groups()
    _render_global_achievements()

func _render_progress_bar() -> void:
    # 总成就数 / 已解锁数 → ProgressBar + 百分比 Label
    pass

func _render_by_game_groups() -> void:
    # 遍历 Registry.all()，有 achievements[] 的游戏建分组 Header + 成就列表
    pass

func _render_global_achievements() -> void:
    # 渲染全局成就（收藏家/好评人/马拉松）
    pass
```

### 4.3 Toast 机制

**P-1修复：** toast_layer.show_msg() 增加可选 color 参数：
```gdscript
## 修改后
func show_msg(msg: String, color: Variant = null) -> void:
    _label.text = msg
    _label.visible = true
    if color != null:
        _label.add_theme_color_override("font_color", color as Color)
    else:
        _label.remove_theme_color_override("font_color")
    # ... tween logic ...
```

**P-2修复：** toast_layer 增加消息队列防重叠：
```gdscript
var _queue: Array[String] = []
var _busy: bool = false
func show_msg(msg: String, color: Variant = null) -> void:
    if _busy:
        _queue.append(msg)
        return
    # ... existing tween logic ...
    _tween.connect("finished", _pop_queue.bind())
func _pop_queue() -> void:
    _busy = false
    if _queue.size() > 0:
        show_msg(_queue.pop_front())
```

Engine 监听 `achievement_unlocked` 信号时调用：
```gdscript
# Engine _enter_tree() 中连接
EventBus.achievement_unlocked.connect(_on_achievement_unlocked)
func _on_achievement_unlocked(def: Dictionary, points: int) -> void:
    var layer := get_tree().root.find_child("ToastLayer", true, false)
    if layer != null:
        layer.call("show_msg", "🏆 成就解锁：%s +%d点" % [def.get("name", ""), points], ThemeTokens.color("gold"))
```

result_overlay 结算卡 400ms 后弹出，Toast 队列自动排队 → 错开 ≥900ms。

---

## 5. Recommender.for_you() 扩展

### 5.1 当前实现 vs 目标

| 维度 | 当前 | 目标 |
|------|------|------|
| 排序依据 | 字母顺序 | 画像加权分 + 未玩过优先 |
| 冷启动 | 返回全量（扣除continue_row） | 回退编辑推荐/热榜 |
| 返回条数 | 全量（UI截断） | top 4（推荐层截断） |
| 网络依赖 | 无 | 纯本地DB |

### 5.2 新算法

```gdscript
## for_you() 扩展（替换现有实现）
func for_you() -> Array[Dictionary]:
    var profile: Dictionary = _build_profile()
    var scored: Array[Dictionary] = []
    
    for meta: GameMeta in Registry.all():
        var score: float = _score_for(meta, profile)
        var rec: Variant = DB.get_record(meta.id)
        var played: bool = (rec is Dictionary) and int((rec as Dictionary).get("last_played", 0)) > 0
        scored.append({"meta": meta, "record": rec, "score": score, "played": played})
    
    # 排序：未玩过优先 → 画像分降序 → title字母序
    scored.sort_custom(func(a: Dictionary, b: Dictionary) -> bool:
        if a["played"] != b["played"]:
            return not a["played"]  # 未玩过排前
        if a["score"] != b["score"]:
            return a["score"] > b["score"]
        return (a["meta"] as GameMeta).title.to_lower() < (b["meta"] as GameMeta).title.to_lower()
    )
    
    # 冷启动回退：无游玩记录 → 编辑推荐/热榜
    if profile.is_empty():
        return _fallback_editorial()
    
    return scored.slice(0, 4)

## 构建用户画像（类目偏好 + 标签偏好）
func _build_profile() -> Dictionary:
    var cats: Dictionary = {}  # category -> weight
    var tags: Dictionary = {}  # tag -> weight
    for gid: String in DB.list_records():
        var rec: Variant = DB.get_record(gid)
        if rec is Dictionary and int((rec as Dictionary).get("last_played", 0)) > 0:
            var meta: Variant = Registry.lookup(gid)
            if meta is GameMeta:
                cats[str(meta.category)] = cats.get(str(meta.category), 0) + 1
                for tag in meta.tags:
                    tags[str(tag)] = tags.get(str(tag), 0) + 1
    # 评价加分
    for gid: String in DB.list_records():
        if DB.get_review(gid) != null:
            var meta: Variant = Registry.lookup(gid)
            if meta is GameMeta:
                cats[str(meta.category)] = cats.get(str(meta.category), 0) + 2
    return {"cats": cats, "tags": tags}

## 计算画像匹配分
func _score_for(meta: GameMeta, profile: Dictionary) -> float:
    if profile.is_empty():
        return 0.0
    var score: float = 0.0
    var cats: Dictionary = profile.get("cats", {})
    var tags: Dictionary = profile.get("tags", {})
    # 类目偏好分（权重 5x）
    score += float(cats.get(str(meta.category), 0)) * 5.0
    # 标签偏好分（权重 1x）
    for tag in meta.tags:
        score += float(tags.get(str(tag), 0))
    return score

## 冷启动回退：编辑推荐/热榜（B-3修复：转 for_you() 格式 {"meta","record"}）
func _fallback_editorial() -> Array[Dictionary]:
    # charts("all") 返回 {"meta","record","rank","players"} → 转为 {"meta","record"}
    var charts_rows: Array[Dictionary] = charts("all")
    var result: Array[Dictionary] = []
    for row in charts_rows.slice(0, 4):
        result.append({"meta": row["meta"], "record": row["record"]})
    return result

## 结果缓存（A-3修复：records_updated 信号失效）
var _cached_result: Array[Dictionary] = []
var _cache_dirty: bool = true

func _ready() -> void:
    EventBus.records_updated.connect(func(_gid: String) -> void: _cache_dirty = true)

func for_you() -> Array[Dictionary]:
    if not _cache_dirty:
        return _cached_result
    # ... existing logic ...
    _cached_result = scored.slice(0, 4) if not profile.is_empty() else _fallback_editorial()
    _cache_dirty = false
    return _cached_result
```

### 5.3 home.gd 零改动

home.gd 已消费 `for_you()` 接口：
```gdscript
func _render_for_you() -> void:
    if _recommender == null:
        return
    var items: Array[Dictionary] = _recommender.for_you()
    # ... 渲染逻辑 ...
```
接口签名不变（`-> Array[Dictionary]`），home.gd 零改动。

---

## 6. 测试方案

### 6.1 test_achievement.tscn（≥10断言）

| 断言 | 验证内容 |
|------|----------|
| 1-3 | Engine._try_unlock() 幂等：同一aid调3次，DB仅记录1条 |
| 4-5 | game_finished result.achievements[] → Engine 正确解锁 |
| 6-7 | _evaluate_global() 收藏家条件（动态阈值=Registry.all().size()）判定正确 |
| 8-9 | _evaluate_global() 好评人条件（≥3条评价）判定正确 |
| 10 | _evaluate_global() 马拉松条件（≥10h时长）判定正确 |
| 11 | achievement_unlocked 信号发射参数正确（def + points） |

### 6.2 test_recommender_profile.tscn（≥10断言）

| 断言 | 验证内容 |
|------|----------|
| 1-3 | _build_profile() 从游玩记录构建类目/标签偏好 |
| 4-5 | _score_for() 同类目游戏得分高于不同类目 |
| 6-7 | for_you() 未玩过优先排序 |
| 8-9 | 冷启动回退（无记录时返回非空） |
| 10 | for_you() 返回 ≤4 条，每条含有效 meta |

### 6.3 GUI 双主题验收

- neon/elegant 下分别截图 library.gd 成就墙页面
- 验证：总成就点/累计时长显示正确、进度条渲染正常、分组标题可见、已解锁/未解锁样式区分、Toast 金色主题色正确

---

## 7. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Registry.ach_def 扩展后 aid 命名空间冲突 | 游戏级/全局级同aid | aid 前缀隔离：`global_*` vs `{game_id}_*` |
| Engine 与 result_overlay 重复调 unlock() | Toast 重复弹出 | db.gd.unlock() 幂等 + Engine 先检查 has(aid) |
| library.gd 超行数上限 | 维护困难 | 成就墙渲染拆为 section_achievement_wall.gd（独立组件） |
| Recommender.for_you() 算法性能 | 全量遍历 O(N_games × N_tags) | 当前 ≤10款游戏，O(1)；M3 云端化时换索引方案 |
