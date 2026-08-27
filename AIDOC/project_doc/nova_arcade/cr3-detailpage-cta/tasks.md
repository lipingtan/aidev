# 任务：CR-3 详情页 CTA 状态机 + 卡片点击导航 + 首页「继续游戏」区块

> 依据 `design.md`（2026-08-25 确认 + 三角色 Review R1/R2 修订）。格式：四要素（task-representation.md）。
> Shell CR 通用 Constraints（hybrid §3.2 + override §2）逐任务隐含生效，不重复列出。
> **后端依赖约束**：本 CR 全部任务不依赖真实后端，涉及支付/订单均走 PurchaseDialog Mock（CR-2 已交付）。

## 进度摘要

| 指标 | 值 |
|------|-----|
| 总任务数 | 9 |
| 已完成 | 9 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 9/9 (100%) |
| 当前阶段 | 全部完成 ✅ |

## 依赖关系

```
T1(Recommender服务)
T2(GameCard组件 + ConfirmBubble) ──────────────────┐
T3(SectionContinue, 依赖T1+T2) ─────────────────────┤
T4(home.gd改写, 依赖T1+T2+T3) ─────────────────────┤
T5(DetailPage + CtaBar, 独立) ──────────────────────┤→ T7(test_detail headless) → T8(GUI双轨+全量回归)
T6(Main.tscn接线 + 导航, 依赖T2+T4+T5) ─────────────┘
T9(PlayButton收尾移除, 依赖T6+T7全绿)
```

- **关键路径**：T1+T2 并行 → T3 → T4+T5 并行 → T6 → T7 → T8；T9 最后执行
- T1/T2 独立，可并行启动
- T5（DetailPage+CtaBar）不依赖 T3/T4，可与 T3/T4 并行

---

### Task 1: Recommender 服务 ⬜

**复杂度**: 中

**Scope:**
- `services/recommender.gd`（新增，class_name Recommender extends Node）：`continue_row() -> Array[Dictionary]` + `for_you() -> Array[Dictionary]`，按 design.md §2
- 不触碰：DB/Registry/EventBus 内部实现

**Constraints:**
- 非 Autoload；由 Main.tscn Services 节点下实例挂载，home.gd 通过 get_node 取引用
- continue_row：last_played>0 过滤 → 按 last_played 倒序 → top5；返回 `[{"meta": GameMeta, "record": Dictionary}]`
- for_you：内部先调 continue_row 取已展示 gid → 全量按 title 排序 → 扣除已展示项；`record` 字段允许为 null（DB.get_record 返回 null 时正常返回，调用方自行处理）
- ≤200 行，中文注释，类型标注完整

**Acceptance:**
- AC: continue_row 返回 last_played>0 的条目，按 last_played 倒序，最多5条
- AC: last_played 全为 0 时 continue_row 返回空数组
- AC: for_you 扣除 continue_row 已展示 gid 后按 title 排序
- AC: for_you 全扣除后返回空数组（本 CR tetra_nova 仅1款，continue_row 非空时 for_you 为空）
- AC: 编辑器零报错，≤200 行

**自测:**
- 测试文件: `tools/test_detail.gd`（统一在 T7 编写，此处记录 ST 点）
- ST: DB 写入 last_played=100 → continue_row 包含该游戏
- ST: DB 写入两条 last_played>0 → continue_row 按倒序排列
- ST: DB 所有 last_played=0 → continue_row 返回空数组
- ST: for_you 扣除 continue_row 后为空 → 返回空数组

---

### Task 2: GameCard 组件 + ConfirmBubble ⬜

**复杂度**: 中

**Scope:**
- `shell/components/game_card.tscn/.gd`（新增）：card_pressed/card_long_pressed 信号 + 500ms 长按计时器 + setup(meta, record)，按 design.md §3
- `shell/components/confirm_bubble.tscn/.gd`（新增）：show_bubble(msg, on_confirm) + [确认]/[取消]，按 design.md §4
- 不触碰：其他组件

**Constraints:**
- GameCard 长按用 `_process` 计时器实现（不依赖 ScrollContainer 事件），规避触摸滑动吞键（Risk-4）
- 抬起时先判 `_long_fired`：未触发长按 → emit card_pressed；已触发 → 不发 card_pressed
- ConfirmBubble 默认 visible=false；show_bubble 后 visible=true；[确认]/[取消] 后 visible=false
- 全字段 ThemeTokens.color()，禁止硬编码颜色
- ≤200 行/文件

**Acceptance:**
- AC: 短按（<500ms）触发 card_pressed，不触发 card_long_pressed
- AC: 长按（≥500ms）触发 card_long_pressed，不触发 card_pressed
- AC: ConfirmBubble.show_bubble 后 visible=true；[确认] 后调 on_confirm + visible=false；[取消] 后 visible=false
- AC: 编辑器零报错；ThemeTokens 颜色无硬编码

**自测:**
- 测试文件: `tools/test_detail.gd`
- ST: 模拟按下后立即抬起（<500ms）→ card_pressed.emit 触发，card_long_pressed 未触发
- ST: 模拟按下后等待 600ms → card_long_pressed.emit 触发，card_pressed 未触发
- ST: show_bubble("msg", cb) → visible=true；点 [确认] → cb 被调用 + visible=false
- ST: show_bubble("msg", cb) → 点 [取消] → cb 未被调用 + visible=false

---

### Task 3: SectionContinue 区块控制器 ⬜

**复杂度**: 中

**依赖**: Task 1, Task 2

**Scope:**
- `shell/components/section_continue.gd`（新增，class_name SectionContinue extends VBoxContainer）：render() + records_updated 信号监听 + 长按移除确认，按 design.md §5
- `shell/components/section_continue.tscn`（新增）：ScrollContainer > HBox 结构
- 不触碰：DB/EventBus 内部实现

**Constraints:**
- Recommender 通过 `set_recommender(r)` 注入（非 Autoload 不可类名调用），`_recommender == null` 时 render 直接返回
- records_updated 信号在 `_ready` 中连接一次；render() 先 free 旧卡片再重建（≤10款全重建耗时可忽略）
- 长按移除：调 ConfirmBubble.show_bubble("从记录移除「{title}」？", cb)；cb 内只调 DB.upsert_record(gid, {last_played:0})，不手动调 render()（由 records_updated 触发自动重建）
- visible = rows.size() > 0（空时整块隐藏）
- ≤150 行

**Acceptance:**
- AC: last_played>0 有数据 → SectionContinue.visible=true，卡片数与 continue_row 结果一致
- AC: last_played 全为 0 → visible=false
- AC: records_updated 信号触发 → render() 重建卡片
- AC: 长按卡片确认移除 → DB.upsert_record(last_played=0) → records_updated → render() 自动更新
- AC: game_finished → records_updated 触发 → SectionContinue 重渲染（继续游戏区块即时更新）（RG-11）
- AC: 强杀重启后 last_played=0 持久，best/trial_used/finish_count 不变（RG-12；持久化验证需强杀重启，headless 无法全自动，T8 GUI 回归时人工确认）

**自测:**
- 测试文件: `tools/test_detail.gd`
- ST: inject recommender + last_played>0 → visible=true + child count 正确
- ST: inject recommender + last_played 全为0 → visible=false
- ST: 触发 records_updated → render() 重建（child count 变化）
- ST: 长按移除确认 → DB.get_record().last_played == 0；best/trial_used 不变

---

### Task 4: home.gd 改写（加继续游戏+为你推荐区块） ⬜

**复杂度**: 中

**依赖**: Task 1, Task 2, Task 3

**Scope:**
- `shell/pages/home.gd`（修改）：新增 _recommender 注入 + on_enter/on_resume 调 _render_for_you() + _ready 中取 Recommender 节点引用并注入 SectionContinue，按 design.md §6
- `shell/pages/home.tscn`（修改）：新增 SectionContinue 节点 + SectionForYou 区块（VBox > ScrollContainer > HBox）
- 不触碰：ThemeToggle 逻辑（_on_theme_pressed 不改）、PlayButton（T9 收尾）

**Constraints:**
- Recommender 访问路径：`get_node("/root/Main/Services/Recommender")` 取引用，Tasks 执行时按 main.tscn 实际 Services 子节点名确认；取到 null 时 push_warning 并跳过（不崩溃）
- home.gd ≤150 行（SectionContinue 逻辑已拆出）
- _render_for_you 在 on_enter + on_resume 均调用（从详情页返回时刷新）
- for_you 为空 → SectionForYou 整块 visible=false

**Acceptance:**
- AC: 冷启动首页：last_played>0 时 SectionContinue 可见；last_played 全为0时 SectionContinue/SectionForYou 均隐藏（RG-9）
- AC: on_resume 时 for_you 区块刷新（详情页 pop 返回后区块状态正确）
- AC: home.gd ≤150 行，编辑器零报错
- AC: 【回归】ThemeToggle 功能不受影响（RG-8）

**自测:**
- 测试文件: `tools/test_detail.gd`（集成验证，T7 统一编写）
- ST: last_played>0 → SectionContinue visible=true
- ST: last_played 全为0 → SectionContinue/SectionForYou 均 visible=false
- ST: on_resume 调用 → _render_for_you 执行（for_you HBox 重建）

---

### Task 5: DetailPage + CtaBar ⬜

**复杂度**: 高

**Scope:**
- `shell/pages/detail.tscn`（新增）：BackButton + ScrollContainer(VBox: IconRect/TitleLabel/VersionLabel/TagsRow/DescLabel/SimilarSection[visible=false]) + CtaBar（底部固定 anchor bottom），按 design.md §7
- `shell/pages/detail.gd`（新增，extends Page）：on_enter(data) 绑定 meta + setup CtaBar；on_resume() 调 _cta_bar.refresh()；meta=null 时 push_error + Nav.pop()，按 design.md §7
- `shell/pages/cta_bar.tscn`（新增）：VBox(HintLabel + CtaButton)
- `shell/pages/cta_bar.gd`（新增，class_name CtaBar）：_ready 连信号；setup(gid)/refresh()；_cta_state() 五分支；_is_owned() 遍历 orders；_show_purchase() 调 PurchaseDialog.show_mock(title, price)，按 design.md §8
- 不触碰：Nav/Launcher/DB/Registry 内部实现

**Constraints:**
- CtaBar 信号（records_updated/trial_consumed）在 `_ready()` 中连接一次，不在 setup() 中重复连接
- `_btn.pressed` 连接改用具名方法 `_on_cta_pressed`，refresh() 中先 disconnect 再 connect（防 action 变化后旧回调残留）
- payment_succeeded/dlc_installed 仅留注释占位（M3/M4 接线）
- CTA 状态机逻辑内聚于 cta_bar.gd，禁止跨页散落
- detail.gd ≤200 行；cta_bar.gd ≤80 行；全字段 ThemeTokens.color()

**Acceptance:**
- AC: on_enter gid 正确 → TitleLabel/VersionLabel/DescLabel 展示 meta 对应字段
- AC: Nav.push("res://shell/pages/detail.tscn", {"gid": gid}) 后 DetailPage 入栈并完成 220ms 右滑入转场（由 Nav._animate_in 负责，AC 验证页面已入栈且 on_enter 被调用）（FR-3/FR-4/FR-7）
- AC: CTA 五分支文案与 requirements.md FR-5 一致：free→[▶ 开玩]；trial N>0→[▶ 试玩·剩N次] + hint；trial 0→[¥X 解锁完整版]；paid 未拥有→[¥X 购买]；owned→[▶ 开玩] + hint 含 best 分
- AC: owned best=0 时 hint 显示「首次开玩」；best>0 时显示「上次最高 {best} 分 · 继续」
- AC: CTA 开玩类点击 → Launcher.launch(gid)；购买类点击 → PurchaseDialog.show_mock(title, price)
- AC: records_updated/trial_consumed 触发后 CTA 文案即时更新（on_resume 刷新）（FR-6）
- AC: meta=null 时 push_error + Nav.pop() 回退，不卡空页
- AC: 编辑器零报错；detail.gd ≤200 行，cta_bar.gd ≤80 行

**自测:**
- 测试文件: `tools/test_detail.gd`
- ST: on_enter {gid:"tetra_nova"} → TitleLabel.text == "TETRA NOVA"
- ST: price_model=free → btn.text == "▶ 开玩"
- ST: price_model=trial, trial_used=0（left=3）→ btn.text 含 "3"
- ST: price_model=trial, trial_used=3（left=0）→ btn.text 含 "¥" 且含 "解锁"
- ST: price_model=paid, no order → btn.text 含 "购买"
- ST: price_model=paid + order.status=paid → btn.text == "▶ 开玩"
- ST: best=0 → hint == "首次开玩"
- ST: best=42 → hint 含 "42"
- ST: records_updated emit(gid) → refresh() 被调用 → 文案更新
- ST: trial_consumed emit(gid, 2) → refresh() 被调用

---

### Task 6: Main.tscn 接线 + 导航路由 ⬜

**复杂度**: 中

**依赖**: Task 2, Task 4, Task 5

**Scope:**
- `shell/main.tscn`（修改）：Services 节点下挂 Recommender 子节点；OverlayLayer 下挂 ConfirmBubble 子节点（static，默认 visible=false）
- `core/nav.gd`（修改）：`_modal_dialog_open()` 实现：查 OverlayLayer 下 ConfirmBubble.visible || PurchaseDialog.visible，有模态弹窗时拦截返回键
- 不触碰：Nav push/pop/switch_tab 主流程；Launcher 内部

**Constraints:**
- Recommender 节点名与 home.gd 中 get_node 路径对齐（执行时按 main.tscn 实际路径确认）
- ConfirmBubble 节点名与 SectionContinue._on_long_press 中 find_child("ConfirmBubble") 对齐
- nav.gd 改动仅限 `_modal_dialog_open()` 函数体，其余接口签名不变；`_modal_dialog_open` 只在 `handle_back()` 调用时执行，不在 `_ready` 初始化时调用，场景树完整时访问 OverlayLayer 无时序问题
- ≤5 行实质改动（nav.gd 函数体填充）

**Acceptance:**
- AC: home.gd `_ready()` 中 get_node Recommender 不返回 null
- AC: SectionContinue._on_long_press 中 find_child("ConfirmBubble") 找到正确节点
- AC: ConfirmBubble 或 PurchaseDialog visible=true 时，返回键被拦截（不触发 Nav.pop）
- AC: 【回归】Nav push/pop/switch_tab 主流程不受影响（RG-10）
- AC: 编辑器零报错

**自测:**
- 测试文件: `tools/test_detail.gd`
- ST: Nav.push(detail) 后 pop → 返回首页（RG-10）
- ST: ConfirmBubble visible=true → handle_back() 返回 true（不 pop）
- ST: ConfirmBubble visible=false → handle_back() 正常 pop

---

### Task 7: test_detail headless 全覆盖 ⬜

**复杂度**: 高

**依赖**: Task 1, Task 2, Task 3, Task 4, Task 5, Task 6

**Scope:**
- `tools/test_detail.tscn/.gd`（新增）：headless 驱动，覆盖 design.md §11 全部测试组
- 不触碰：功能代码（发现缺陷回退对应 Task）

**Constraints:**
- headless 可跑（console exe + 隔离 APPDATA，沿用 CR-2 环境备忘）
- DB 操作前必须 `DB.get_record(gid).duplicate()` 避免引用污染断言
- 测试间独立：每 case 重置 DB 相关字段（upsert 置初始值）
- 退出码 0，无脚本错误

**Acceptance:**
- AC: CTA 五分支逐一断言（btn.text + hint.text）
- AC: 继续游戏区块：last_played>0 → visible+card 数；全为0 → visible=false
- AC: 长按移除：DB.get_record().last_played==0；best/trial_used/finish_count 不变
- AC: records_updated 信号触发 → SectionContinue 重渲染
- AC: Nav.push(detail, {gid}) → DetailPage.on_enter → gid 正确；Nav.pop → 回首页
- AC: trial_consumed emit 后调 on_resume → CTA 文案更新
- AC: 退出码 0，无脚本错误

**自测:**
- 测试文件: `tools/test_detail.gd`（本任务即为测试文件实现）
- ST: 所有上述 AC 均对应测试函数（test_cta_free/test_cta_trial/test_cta_trial_exhausted/test_cta_paid/test_cta_owned/test_continue_row_visible/test_continue_row_empty/test_long_press_remove/test_records_updated_rerender/test_nav_push_pop/test_on_resume_refresh）

---

### Task 8: GUI 双轨截图 + 全量回归 ⬜

**复杂度**: 高

**依赖**: Task 7

**Scope:**
- GUI exe（隔离 APPDATA + 清 tmp_gui\Godot）：neon/elegant × 首页（含继续游戏区块）× 详情页截图 read_image 断言
- 全量回归：RG-1~12 + test_launch/test_db×3/test_rg5 双相/源基线
- 截图前置数据：upsert last_played>0 + trial_used=2（展示试玩状态）；额外一轮 trial_used=meta.trial.plays（展示购买按钮）

**Constraints:**
- GUI exe 必须隔离 `$env:APPDATA = dev\tmp_gui`（CR-2 踩坑，signal 11 崩溃防护）
- 截图前清 `tmp_gui\Godot`（trial 3次/主题，防耗尽弹 Mock 影响截图）
- read_image 断言：无布局溢出，按钮/标签/hint 颜色非硬编码

**Acceptance:**
- AC: neon/elegant × 首页/详情页 × 3状态（继续游戏有数据/trial剩余/trial耗尽购买按钮）截图 read_image 全通（FR-10）
- AC: RG-1~12 逐条通过；特别核查 RG-9（冷启动首页区块隐藏）+ RG-11（launch quit 后继续游戏区块即时更新）
- AC: test_detail headless 全绿（复跑 T7）
- AC: test_launch/test_db×3/test_rg5 全绿

**自测:**
- 测试文件: GUI 驱动脚本（沿用 CR-2 _shot3.gd 体系扩展）
- ST: neon 首页截图 → read_image 无异常色块 / 无溢出节点
- ST: elegant 详情页截图（trial 状态）→ CTA 按钮颜色为 ThemeTokens buy-grad

---

### Task 9: PlayButton 收尾移除 ⬜

**复杂度**: 低

**依赖**: Task 6, Task 7（全绿后执行）；**T4 已修改 home.gd，本 Task 在 T4 完成后才可操作 home.gd，禁止并行**

**Scope:**
- `shell/pages/home.gd`（修改）：删 `@onready var _play_btn` + `_on_play_pressed` 方法及「CR-3 移除」注释行
- `shell/pages/home.tscn`（修改）：删 `PlayButton` 节点
- `tools/test_main.gd`（修改）：将 PlayButton 存在断言改为 SectionContinue 存在断言
- `tools/test_nav.gd`（修改）：移除 PlayButton 点击触发 launch 测试用例（改由 test_detail 覆盖）
- 不触碰：其他节点和逻辑

**Constraints:**
- 执行前先 grep `test_main.gd` 和 `test_nav.gd` 确认 PlayButton 相关断言行数，逐行移除或替换，不盲改
- home.gd 删除后行数应 ≤130 行（T4 改写后基准）

**Acceptance:**
- AC: home.tscn 中 PlayButton 节点已删除；home.gd 中无 _play_btn/_on_play_pressed 残留
- AC: test_main/test_nav headless 重跑全绿
- AC: 首页仍有开玩入口（SectionContinue 卡片点击）（RG-9）
- AC: 编辑器零报错
