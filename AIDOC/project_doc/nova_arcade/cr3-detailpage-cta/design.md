# 设计：CR-3 详情页 CTA 状态机 + 卡片点击导航 + 首页「继续游戏」区块

> 依据：已确认的 `requirements.md`（Q1~Q6）+ `design_plan.md`（D1~D6/Q1~Q8 全采纳，2026-08-25）。
> 上游：client-design §4.0/§4.2/§4.5/§4.6/§6/§7、override §2/§5、hybrid §3.2+§3.3。
> 工程实查：services/ 已有 Launcher/TrialGuard；core/ 已有 Nav(push/pop)/DB(upsert_record→records_updated)/EventBus(records_updated/trial_consumed)；shell/pages/ 已有 home.gd(PlayButton 临时入口)；shell/components/ 已有 purchase_dialog/result_overlay/transition_overlay。

---

## 1. 文件与目录变更

```
nova-arcade/
├── services/
│   └── recommender.gd              # 新增：Recommender 服务（continue_row/for_you）
├── shell/
│   ├── components/
│   │   ├── confirm_bubble.tscn/.gd # 新增：轻量确认弹窗（Label + [确认]/[取消]）
│   │   ├── game_card.tscn/.gd      # 新增：GameCard 卡片组件（含长按计时器）
│   │   └── section_continue.gd     # 新增：继续游戏区块控制器（≤150 行）
│   └── pages/
│       ├── home.gd / home.tscn     # 改写：加继续游戏+为你推荐区块，移除 PlayButton（收尾 Task）
│       ├── detail.gd / detail.tscn # 新增：详情页（基础信息块+信号路由）
│       └── cta_bar.gd / cta_bar.tscn # 新增：CTA 固定区组件（_cta_state，≤80 行）
└── tools/
    └── test_detail.tscn / test_detail.gd  # 新增：CTA/导航/继续游戏 headless 测试
```

**Main.tscn 变更**：OverlayLayer 下增加 ConfirmBubble 子节点（静态挂载，show/hide 控制）；Services 节点下增加 Recommender 非单例节点。

---

## 2. Recommender（services/recommender.gd）

非单例，挂 `Main/Services` 节点下；home.gd 通过 `get_node("/root/Main/Services/Recommender")` 或 `$Recommender`（按 Main.tscn 挂载路径）访问。

```gdscript
class_name Recommender extends Node

## continue_row() → Array[Dictionary]
## 读 DB.list_records()，过滤 last_played > 0，按 last_played 倒序，top5
## 返回 GameMeta 的 dict 表示（Registry.lookup(gid) 转 dict，缺失 gid 跳过）
func continue_row() -> Array[Dictionary]:
    var rows: Array[Dictionary] = []
    for gid in DB.list_records():
        var rec: Variant = DB.get_record(gid)
        if rec == null: continue
        var r := rec as Dictionary
        if int(r.get("last_played", 0)) > 0:
            var meta := Registry.lookup(gid)
            if meta != null:
                rows.append({"meta": meta, "record": r})
    rows.sort_custom(func(a, b): return int(a["record"]["last_played"]) > int(b["record"]["last_played"]))
    return rows.slice(0, 5)

## for_you() → Array[Dictionary]
## 全量游戏按 title 排序，扣除 continue_row 已展示的 gid
func for_you() -> Array[Dictionary]:
    var cont_gids: Array[String] = []
    for item in continue_row():
        cont_gids.append((item["meta"] as GameMeta).id)
    var result: Array[Dictionary] = []
    var all := Registry.all()
    all.sort_custom(func(a, b): return (a as GameMeta).title < (b as GameMeta).title)
    for meta in all:
        var m := meta as GameMeta
        if not cont_gids.has(m.id):
            result.append({"meta": m, "record": DB.get_record(m.id)})
    return result
```

---

## 3. GameCard 组件（shell/components/game_card.tscn/.gd）

通用卡片，供继续游戏/为你推荐区块复用。

**场景结构**：`PanelContainer > VBoxContainer > [IconRect, NameLabel, StatusBadge]`

```gdscript
class_name GameCard extends PanelContainer

signal card_pressed(gid: String)
signal card_long_pressed(gid: String)

## 长按阈值（毫秒）
const LONG_PRESS_MS: int = 500

var _gid: String = ""
var _press_start: int = -1
var _long_fired: bool = false

func setup(meta: GameMeta, record: Variant) -> void:
    _gid = meta.id
    $NameLabel.text = meta.title
    # icon/badge/渐变背景按 ThemeTokens 设置

## 触摸/鼠标按下开始计时
func _gui_input(event: InputEvent) -> void:
    if event is InputEventMouseButton:
        var e := event as InputEventMouseButton
        if e.button_index == MOUSE_BUTTON_LEFT:
            if e.pressed:
                _press_start = Time.get_ticks_msec()
                _long_fired = false
            else:
                if not _long_fired and _press_start >= 0:
                    var dt := Time.get_ticks_msec() - _press_start
                    if dt < LONG_PRESS_MS:
                        card_pressed.emit(_gid)
                _press_start = -1

func _process(_dt: float) -> void:
    if _press_start >= 0 and not _long_fired:
        if Time.get_ticks_msec() - _press_start >= LONG_PRESS_MS:
            _long_fired = true
            card_long_pressed.emit(_gid)
            _press_start = -1
```

> Risk-4 解法：GameCard 自行实现长按计时器，不依赖 ScrollContainer 内部事件，规避触摸滑动吞键问题。

---

## 4. ConfirmBubble（shell/components/confirm_bubble.tscn/.gd）

挂 `OverlayLayer` 下；默认 `visible = false`。

```gdscript
extends Control
## 轻量确认弹窗：show(msg, on_confirm) 弹出，[确认] 调 on_confirm 后关闭

signal confirmed
signal cancelled

@onready var _label: Label = $Panel/VBox/Label
@onready var _confirm_btn: Button = $Panel/VBox/HBox/ConfirmBtn
@onready var _cancel_btn: Button = $Panel/VBox/HBox/CancelBtn

var _on_confirm: Callable

func show_bubble(msg: String, on_confirm: Callable) -> void:
    _label.text = msg
    _on_confirm = on_confirm
    visible = true

func _ready() -> void:
    _confirm_btn.pressed.connect(_do_confirm)
    _cancel_btn.pressed.connect(_do_cancel)

func _do_confirm() -> void:
    visible = false
    confirmed.emit()
    if _on_confirm.is_valid():
        _on_confirm.call()

func _do_cancel() -> void:
    visible = false
    cancelled.emit()
```

---

## 5. SectionContinue（shell/components/section_continue.gd）

继续游戏区块控制器；挂 home.tscn 内对应 VBoxContainer 区块下。

```gdscript
class_name SectionContinue extends VBoxContainer
## 继续游戏区块：渲染卡片 + 监听 records_updated + 长按移除确认
## Recommender 为非单例，由 home.gd 在 _ready 时通过 @export 或 set_recommender() 注入

@export var card_scene: PackedScene  # GameCard.tscn
@onready var _scroll: ScrollContainer = $ScrollContainer
@onready var _hbox: HBoxContainer = $ScrollContainer/HBox

var _recommender: Recommender = null  # 由 home.gd 注入

func set_recommender(r: Recommender) -> void:
    _recommender = r

func _ready() -> void:
    EventBus.records_updated.connect(_on_records_updated)
    render()

func render() -> void:
    for c in _hbox.get_children(): c.queue_free()
    if _recommender == null: return
    var rows := _recommender.continue_row()
    visible = rows.size() > 0
    for item in rows:
        var card := card_scene.instantiate() as GameCard
        card.setup(item["meta"], item["record"])
        card.card_pressed.connect(func(gid): Nav.push("res://shell/pages/detail.tscn", {"gid": gid}))
        card.card_long_pressed.connect(_on_long_press.bind(item["meta"].id, item["meta"].title))
        _hbox.add_child(card)

func _on_records_updated(_gid: String) -> void:
    render()

func _on_long_press(gid: String, title: String) -> void:
    var bubble := get_tree().root.find_child("ConfirmBubble", true, false)
    (bubble as ConfirmBubble).show_bubble(
        "从记录移除「%s」？" % title,
        func():
            DB.upsert_record(gid, {"last_played": 0})
            # records_updated 由 DB.upsert_record 自动发出，render() 将由信号触发
    )
```

> **Risk-2 确认已解**：DB.upsert_record 实现中每次 upsert 都发 records_updated（代码已实查），长按移除路径同样触发，无需补发。

---

## 6. home.gd 改写

```gdscript
extends Page
## 首页：ThemeToggle + 继续游戏(SectionContinue) + 为你推荐区块 + Banner 占位

@onready var _theme_btn: Button = $ThemeButton
@onready var _section_continue: SectionContinue = $ScrollView/VBox/SectionContinue
@onready var _for_you_root: VBoxContainer = $ScrollView/VBox/SectionForYou
@onready var _for_you_hbox: HBoxContainer = $ScrollView/VBox/SectionForYou/ScrollContainer/HBox

@export var card_scene: PackedScene  # GameCard.tscn

## Recommender 非单例，_ready 时从 Services 节点取引用并注入 SectionContinue
var _recommender: Recommender = null

func _ready() -> void:
    _theme_btn.pressed.connect(_on_theme_pressed)
    _recommender = get_node("/root/Main/Services/Recommender") as Recommender
    _section_continue.set_recommender(_recommender)

func on_enter(data: Dictionary) -> void:
    _render_for_you()

func on_resume() -> void:
    _render_for_you()  # 从详情页返回时刷新

## 为你推荐区块：for_you 非空才渲染，空则隐藏
func _render_for_you() -> void:
    for c in _for_you_hbox.get_children(): c.queue_free()
    if _recommender == null: return
    var rows := _recommender.for_you()
    _for_you_root.visible = rows.size() > 0
    for item in rows:
        var card := card_scene.instantiate() as GameCard
        card.setup(item["meta"], item["record"])
        card.card_pressed.connect(func(gid): Nav.push("res://shell/pages/detail.tscn", {"gid": gid}))
        _for_you_hbox.add_child(card)

func _on_theme_pressed() -> void:
    var tw := create_tween()
    tw.tween_property(_theme_btn, "scale", Vector2(0.9, 0.9), 0.06)
    tw.tween_property(_theme_btn, "scale", Vector2.ONE, 0.12)
    Sound.toggle()
    var next := "elegant" if ThemeTokens.current == "neon" else "neon"
    ThemeTokens.apply_theme(next, get_tree().root)
```

> Risk-1 解法：home.gd 行数通过将继续游戏区块拆至 `SectionContinue` 组件控制在 ≤100 行；为你推荐使用 `_render_for_you()` 内联（≤30 行），整体不超 150 行。

> **PlayButton 移除**：`_play_btn` / `_on_play_pressed` 收尾 Task（T7）中删除，同步更新 test_main/test_nav 断言。

---

## 7. DetailPage（shell/pages/detail.tscn / detail.gd）

场景结构：
```
DetailPage (Control/Page)
├── BackButton (Button)
├── ScrollContainer
│   └── VBox
│       ├── IconRect (TextureRect)
│       ├── TitleLabel (Label)
│       ├── VersionLabel (Label)
│       ├── TagsRow (HBoxContainer)        # 标签行，ThemeTokens
│       ├── DescLabel (Label)              # 简介
│       └── SimilarSection (VBoxContainer, visible=false)  # 相关推荐占位（M2 填充数据后自动显示）
└── CtaBar (cta_bar.tscn)                  # 底部固定，anchor bottom
```

```gdscript
extends Page
## 详情页：基础信息块 + CTA 固定区；入页 220ms 右滑入（Nav 负责）

@onready var _cta_bar: CtaBar = $CtaBar
@onready var _title: Label = $ScrollContainer/VBox/TitleLabel
@onready var _version: Label = $ScrollContainer/VBox/VersionLabel
@onready var _desc: Label = $ScrollContainer/VBox/DescLabel
@onready var _back_btn: Button = $BackButton

var _gid: String = ""

func on_enter(data: Dictionary) -> void:
    _gid = data.get("gid", "")
    var meta := Registry.lookup(_gid)
    if meta == null:
        push_error("DetailPage: 未找到游戏 %s" % _gid)
        Nav.pop()  # 异常回退，避免页面卡在空态
        return
    _title.text = meta.title
    _version.text = "v" + meta.version
    _desc.text = meta.desc
    _cta_bar.setup(_gid)

## Risk-3：结算卡关闭后回 Tab 根页走 on_resume，需刷新 CTA
func on_resume() -> void:
    if _gid != "":
        _cta_bar.refresh()

func _ready() -> void:
    _back_btn.pressed.connect(func(): Nav.pop())
```

---

## 8. CtaBar（shell/pages/cta_bar.tscn / cta_bar.gd）

底部固定 CTA 区，≤80 行；`setup(gid)` 初始化，`refresh()` 重算。

```gdscript
class_name CtaBar extends Control

@onready var _hint: Label = $VBox/HintLabel
@onready var _btn: Button = $VBox/CtaButton

var _gid: String = ""

func _ready() -> void:
    # 信号在 _ready 中连接一次，避免 setup() 多次调用时重复连接
    EventBus.records_updated.connect(_on_records_updated)
    EventBus.trial_consumed.connect(_on_trial_consumed)
    # payment_succeeded / dlc_installed → 占位注释（M3/M4 接线）

func setup(gid: String) -> void:
    _gid = gid
    refresh()

func refresh() -> void:
    if _gid == "": return
    var state := _cta_state()
    _hint.text = state["hint"]
    _btn.text = state["label"]
    # 断开旧连接后重连，防止 push 同一页面时 action 变化
    if _btn.pressed.is_connected(_on_cta_pressed):
        _btn.pressed.disconnect(_on_cta_pressed)
    _btn.pressed.connect(_on_cta_pressed)

func _on_cta_pressed() -> void:
    var state := _cta_state()
    if state["action"] == "launch":
        Launcher.launch(_gid)
    else:
        _show_purchase()

## CTA 状态机（FR-5）
func _cta_state() -> Dictionary:
    var meta := Registry.lookup(_gid)
    if meta == null: return {"label": "—", "hint": "", "action": "none"}
    var rec: Variant = DB.get_record(_gid)
    var r := rec if rec is Dictionary else {}
    var model: String = meta.price_model
    var price: int = meta.price
    match model:
        "free", "ad", "iap":
            return _open_play_state(r)
        "trial":
            var left: int = TrialGuard.left(_gid)
            if left > 0:
                return {"label": "▶ 试玩 · 剩 %d 次" % left,
                        "hint": "试玩结束后 ¥%d 解锁全部" % price,
                        "action": "launch"}
            else:
                return {"label": "¥%d 解锁完整版" % price,
                        "hint": "试玩次数已用完", "action": "buy"}
        "paid":
            if _is_owned():
                return _open_play_state(r)
            return {"label": "¥%d 购买" % price,
                    "hint": "买断制 · 一次购买永久拥有", "action": "buy"}
        _:  # DLC/Arcade 预留
            return {"label": "—", "hint": "", "action": "none"}

func _open_play_state(r: Dictionary) -> Dictionary:
    var best: int = int(r.get("best", 0))
    var hint := "首次开玩" if best == 0 else "上次最高 %d 分 · 继续" % best
    return {"label": "▶ 开玩", "hint": hint, "action": "launch"}

func _is_owned() -> bool:
    for order in DB.list_orders().values():
        var o := order as Dictionary
        if o.get("gid", "") == _gid and o.get("status", "") == "paid":
            return true
    return false

func _show_purchase() -> void:
    var meta := Registry.lookup(_gid)
    if meta == null: return
    PurchaseDialog.show_mock(meta.title, meta.price)  # show_mock(title, price) 签名

func _on_records_updated(gid: String) -> void:
    if gid == _gid: refresh()

func _on_trial_consumed(gid: String, _left: int) -> void:
    if gid == _gid: refresh()
```

> **Button.pressed.disconnect_all()**：Godot 4.5 `disconnect_all()` 不存在；实现时改用 `_btn.pressed.disconnect` 显式断开已连接回调，或用 flag 防重连。Tasks 阶段细化。

---

## 9. 导航接线

| 触发点 | 代码 |
|--------|------|
| SectionContinue 卡片点击 | `Nav.push("res://shell/pages/detail.tscn", {"gid": gid})` |
| home.gd 为你推荐卡片点击 | 同上 |
| 详情页返回键 / 系统返回 | `Nav.pop()` |
| CTA → launch | `Launcher.launch(gid)`（CR-2 转场链路，遮罩+视口+boot） |
| CTA → 购买 | `PurchaseDialog.show_mock(price)` |

`Nav._modal_dialog_open()`（CR-1 预留桩）：接入 ConfirmBubble/PurchaseDialog 后，此函数返回是否有模态弹窗打开，ConfirmBubble/PurchaseDialog 的 visible 变化时通知 Nav（或 Nav 直接查 OverlayLayer 子节点 visible）。

---

## 10. PlayButton 移除（收尾 Task T7）

- 删 `home.gd` 中 `@onready var _play_btn` + `_on_play_pressed` 方法
- 删 `home.tscn` 中 `PlayButton` 节点
- `tools/test_main.gd`：将"PlayButton 存在"断言改为"继续游戏区块存在"断言
- `tools/test_nav.gd`：移除 PlayButton 点击触发 launch 的测试用例（改由 test_detail 覆盖）

---

## 11. test_detail（tools/test_detail.tscn/.gd）

headless 可跑，覆盖 FR-9：

| 测试组 | 覆盖场景 |
|--------|----------|
| CTA 状态机 | free → [▶ 开玩]；trial N>0 → 按钮文案含剩余次数；trial 0 → [¥X 解锁完整版]；paid 未拥有 → [¥X 购买]；owned(paid+order) → [▶ 开玩] hint 含分数 |
| 继续游戏区块 | last_played>0 → 区块可见 + 卡片数正确；last_played 全为 0 → visible=false |
| 长按移除 | 触发 _on_long_press → DB.upsert_record(last_played=0) → records_updated → render() |
| openDetail push/pop | Nav.push → DetailPage 入栈；on_enter gid 正确；Nav.pop → 回首页 |
| on_resume 刷新 | 模拟 trial_consumed 后调 on_resume → CTA 文案更新 |

---

## 12. GUI 双轨（FR-10）

继承 CR-2 流程（GUI exe + 隔离 APPDATA + 清 tmp_gui\Godot）：
- neon/elegant × 首页（含继续游戏区块）× 详情页截图 read_image 断言
- 断言：无溢出、按钮/标签/hint 颜色走 ThemeTokens（非硬编码）
- **截图前置**：test_db 强写 last_played>0（确保继续游戏区块有内容）+ trial_used=2（确保 CTA 显示试玩状态）
- trial 耗尽截图（CTA 显示购买按钮）单独一轮：upsert trial_used=meta.trial.plays

---

## 13. 不变行为清单（回归防护）

| # | WHEN | THEN 系统 SHALL |
|---|------|-----------------|
| RG-1~8 | （沿用 CR-2 design.md §8 全部条目） | 不破坏：启动/Tab/push-pop/主题强写/CoreManager Mock/launch链路 |
| RG-9 | home 改写后冷启动 | 四 Tab 正常切换；继续游戏区块空时隐藏，首页无空白块 |
| RG-10 | Nav.push(DetailPage) 后 pop | 返回首页，首页状态保持（Banner/区块不重建） |
| RG-11 | launch 后 quit + 回 Shell | ResultOverlay 正常弹出；首页继续游戏区块即时更新（records_updated 触发） |
| RG-12 | 长按移除后强杀重启 | last_played=0 持久；best/trial_used/finish_count 不变 |

---

## 14. 正确性属性

- **CTA 幂等性**：同一 gid 多次调用 `_cta_state()` 结果相同（无副作用，纯读 DB+Registry）
- **权益红线**：长按移除只清 last_played，trial_used/best/finish_count 不被改写（DB.upsert 浅合并保证）
- **信号过滤**：DetailPage 只响应本 gid 的 records_updated/trial_consumed，外部游戏状态变化不触发刷新
- **视口铁律**（override §2）：CTA→launch 走 Launcher.launch，转场遮罩完全盖住后才切视口，本 CR 不绕过此路径
- **区块隐藏完整性**：continue_row 为空时 SectionContinue.visible=false；for_you 为空时 SectionForYou.visible=false；两者均空时首页无空白区块

---

## 15. 决策记录

- `_btn.pressed.disconnect_all()` → GDScript 无此方法，Tasks 阶段改用 flag 或显式 disconnect
- `Recommender` 访问路径：home.gd 通过 `get_node("/root/Main/Services/Recommender")` 取；Tasks 阶段按 main.tscn 实际结构确认路径
- `CtaBar._is_owned()`：M4 前通过遍历 orders 判断；M4 接支付后改由 `PayService.owned(gid)`
- Mock 桩位：PurchaseDialog.show_mock（M4）；similar 区块（M2，结算卡"换一个"归 CR-2 已处理）

## 16. 修订记录

- 2026-08-25 初版（D1~D6/Q1~Q8 全采纳）
- 2026-08-25 R1：三角色 Review 第一轮修复：CtaBar 信号连接移至 _ready()（高优）；DetailPage SimilarSection 占位节点补入（中优）
- 2026-08-25 R2：三角色 Review 第二轮修复：PurchaseDialog.show_mock 补 title 参数（高优）；SectionContinue/home.gd Recommender 改为注入访问（高优）；detail.gd on_enter meta=null 补 Nav.pop() 回退（中优）；for_you() double-traverse 低优标注
