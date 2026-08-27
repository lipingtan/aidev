# 任务：CR-4 分类页 + 搜索页 + 本地评价系统

> 依据 `design.md`（2026-08-25 确认，D1~D8 全采纳）。格式：三要素（Scope/Constraints/Acceptance）。
> Shell CR 通用 Constraints（hybrid §3.2 + override §2）逐任务隐含生效，不重复列出。
> **后端依赖约束**：本 CR 全部任务不依赖真实后端，全本地运行。

## 进度摘要

| 指标 | 值 |
|------|-----|
| 总任务数 | 11 |
| 已完成 | 0 |
| 进行中 | 0 |
| 未开始 | 11 |
| 完成率 | 0/11 (0%) |
| 当前阶段 | 待执行 |

## 依赖关系

```
T1(Registry.query扩展 + DB补充接口 + editorial.json)
T2(Searcher服务, 依赖T1)
T3(GameCard扩展size参数)
T4(CategoryPage完整实现, 依赖T1+T3)
T5(SearchPage完整实现, 依赖T2+T3)
T6(ReviewEditor模态弹窗, 独立)
T7(ReviewsPage, 依赖T6)
T8(DetailPage补充评价入口, 依赖T6+T7)
T9(ResultOverlay接线ReviewEditor, 依赖T6)
T10(Main.tscn接线 + Nav扩展, 依赖T2+T6+T4+T5+T7)
T11(headless测试 + GUI双轨 + 全量回归, 依赖T1~T10)
```

- **关键路径**：T1 → T2+T3 → T4+T5 → T10 → T11
- T1/T3/T6 独立，可并行启动
- T2 依赖 T1；T4 依赖 T1+T3；T5 依赖 T2+T3；T7 依赖 T6；T8/T9 依赖 T6

---

### Task 1: Registry.query扩展 + DB补充接口 + editorial.json ⬜

**复杂度**: 中

**Scope:**
- `core/registry.gd`（修改）：query() 新增三个过滤维度：`price_models: Array[String]`（多选）、`min_rating: float`（M1 占位不过滤）、`is_new: bool`（M1 占位不过滤）；category="" 时跳过分类过滤；补充对应私有辅助函数 `_filter_price_models / _filter_min_rating / _filter_is_new`
- `core/db.gd`（修改）：新增 `delete_review(gid)` + `clear_search_history()` 两个方法
- `data/editorial.json`（修改）：新增 `hot_queries` 字段（6 条热搜词）
- 不触碰：其他 Registry/DB 接口签名

**Constraints:**
- 现有 query() 调用方（CategoryPage 骨架、Recommender 等）向后兼容，不传新 key 时行为不变
- delete_review：强写落盘；clear_search_history：防抖写落盘（与 add_search_history 同级）
- registry.gd 行数检查：原 ~182 行 + 新增 ~20 行，若超 200 行则拆 registry_filter.gd 辅助类
- 中文注释，类型标注完整

**Acceptance:**
- AC: Registry.query({price_models:["free"]}) 只返回 price_model="free" 的游戏
- AC: Registry.query({category:""}) 等价于 Registry.all()（不过滤分类）
- AC: Registry.query({}) 不报错，返回全量（向后兼容）
- AC: DB.delete_review(gid) 后 DB.get_review(gid) 返回 null
- AC: DB.clear_search_history() 后 DB.get_search_history() 返回空数组
- AC: editorial.json 含 hot_queries 字段，Registry.get_editorial() 可读到
- AC: 编辑器零报错；test_registry 回归全绿

---

### Task 2: Searcher 服务 ⬜

**复杂度**: 中

**依赖**: Task 1

**Scope:**
- `services/searcher.gd`（新增，class_name Searcher extends Node）：IndexEntry 内部类 + `build_index()` + `query(text) -> Array[String]` + `hot_queries() -> Array[String]`，按 design.md §5
- 不触碰：Registry/DB 内部实现

**Constraints:**
- 非 Autoload；由 Main.tscn Services 节点下实例挂载
- build_index() 在 Main._ready() 中 Registry.reload() 后调用（T10 接线）
- _extract_initials() M1 取全拼串首字符（占位），不影响基本前缀匹配
- query() 同步返回，调用方（SearchPage）负责 150ms 去抖
- GameMeta 若无 aliases/pinyin 字段则退化为纯标题匹配，不崩溃
- **A-1**：query() 中 sort_custom 禁止单行 lambda，改用具名比较器 _cmp_score_desc + _sort_scores 成员变量缓存
- **A-2**：IndexEntry 数组字段（aliases_lower/pinyin_full/pinyin_initials/tags_lower）必须在 build_index() 中显式赋 `= []`，不依赖类默认值（防跨实例共享引用）
- ≤200 行；中文注释；类型标注完整

**Acceptance:**
- AC: build_index() 后 query("tetra") 返回包含 "tetra_nova" 的 gid 数组（标题前缀命中）
- AC: query("消除") 命中标签包含 "消除" 的游戏（标签 +15 分）
- AC: query("") 返回空数组
- AC: hot_queries() 返回 editorial.json hot_queries 数组，长度 ≤6
- AC: 编辑器零报错；≤200 行

---

### Task 3: GameCard 扩展 size 参数 ⬜

**复杂度**: 低

**Scope:**
- `shell/components/game_card.gd`（修改）：setup() 新增 `size: String = "card"` 参数；size="grid" 时设置 `size_flags_horizontal = SIZE_EXPAND_FILL` + `custom_minimum_size = Vector2(0, 140)`
- `shell/components/game_card.tscn`（不改）：场景结构不变
- 不触碰：card_pressed/card_long_pressed 信号逻辑

**Constraints:**
- 默认参数 "card"：现有所有 setup(meta, record) 调用方零改动（向后兼容）
- grid 模式不改变长按/短按逻辑
- ≤原行数 +10 行

**Acceptance:**
- AC: setup(meta, record) 不传 size → 行为与 CR-3 完全一致（向后兼容）
- AC: setup(meta, record, "grid") → size_flags_horizontal = SIZE_EXPAND_FILL，custom_minimum_size.y = 140
- AC: test_detail 回归全绿（不因 GameCard 改动断言失败）

---

### Task 4: CategoryPage 完整实现 ⬜

**复杂度**: 高

**依赖**: Task 1, Task 3

**Scope:**
- `shell/pages/category.tscn`（改写）：HSplitContainer（左侧分类栏 + 右侧面板）；右侧含 FilterRow(HBox chips) + SortDropdown(visible=false) + GridContainer(columns=2) + EmptyState(VBox)
- `shell/pages/category.gd`（改写）：完整分类栏 + chips 筛选 + _render_grid() + 空态显示，按 design.md §7
- 不触碰：其他页面；Nav/DB 内部实现

**Constraints:**
- chips "全部"互斥（选中时清空 price_models/min_rating/is_new）；其余多选叠加（price_models 追加/移除）
- SortDropdown visible=false（M2 解锁），代码预留不实现逻辑
- _render_grid() 先 queue_free 旧卡片再重建（内存即时，≤50 款无性能问题）
- 空态时 GridContainer visible=false，EmptyState visible=true，内联"清除筛选"按钮可用
- category.gd ≤150 行；全字段 ThemeTokens.color()

**Acceptance:**
- AC: 点击左侧"消除/益智" → 只显示 category="puzzle" 游戏（B-4：合并后单按钮）
- AC: 选 chip"免费" → price_models=["free","ad","iap"] 过滤生效
- AC: 再选 chip"全部" → price_models 清空，所有筛选重置
- AC: 筛选无结果 → EmptyState 可见，GridContainer 不可见，"清除筛选"按钮可点
- AC: 点"清除筛选" → 恢复全量显示
- AC: 卡片点击 → Nav.push DetailPage，gid 正确
- AC: category.gd ≤150 行，编辑器零报错

---

### Task 5: SearchPage 完整实现 ⬜

**复杂度**: 高

**依赖**: Task 2, Task 3

**Scope:**
- `shell/pages/search.tscn`（改写）：TopBar(LineEdit+CancelBtn) + DefaultView(HotSection+HistorySection) + ResultList + EmptyState
- `shell/pages/search.gd`（改写）：输入去抖 150ms + Searcher.query + 热搜/历史渲染 + 结果点击导航，按 design.md §8
- 不触碰：Searcher/DB 内部实现

**Constraints:**
- 去抖：text_changed 时设 `_debounce_active=false`（取消旧）+ 新建 SceneTreeTimer(0.15)；timeout 后检查 flag 再执行查询
- Searcher 通过 `get_node("/root/Main/Services/Searcher")` 取引用，null 时 push_warning 跳过不崩溃
- 点击结果：先 DB.add_search_history(q) 再 Nav.push（历史含本次搜索词）
- 无结果：EmptyState.text = "没有找到「{q}」"，不实现猜你喜欢（M2）
- [取消] = Nav.pop()
- **A-1**：热搜/历史 btn.pressed 不可用单行多语句 lambda，改用 bind + 具名方法 `_on_hot_or_hist_pressed(q)`
- **P-2**：card.card_pressed.connect(_on_result_pressed.bind(q)) 需验证 bind 语义：测试断言收到的 gid 与 card 对应游戏一致，q 与搜索框文字一致；若有歧义改用 lambda 分两行
- search.gd ≤150 行

**Acceptance:**
- AC: 进入搜索页（on_enter）→ LineEdit 自动聚焦，输入框清空，显示热搜+历史默认视图（P-5）
- AC: Tab 切回（on_resume）→ 输入框保持上次内容，历史行刷新，不重置搜索状态（P-5）
- AC: 输入 "tetra" → 150ms 后结果列表显示 TETRA NOVA
- AC: 输入 "" → 结果列表隐藏，DefaultView（热搜+历史）显示
- AC: 点击结果 → DetailPage 打开，DB.get_search_history() 包含该搜索词且置顶
- AC: 重复搜索同一词 → 历史无重复，该词仍置顶
- AC: 无结果 → "没有找到「xx」"文案，无崩溃
- AC: 点 [取消] → Nav.pop() 回上页
- AC: search.gd ≤150 行，编辑器零报错

---

### Task 6: ReviewEditor 模态弹窗 ⬜

**复杂度**: 高

**Scope:**
- `shell/overlays/review_editor.tscn`（新增）：Dim(ColorRect) + Card(PanelContainer > VBox: TitleLabel/StarRow/TextEdit/CharCount/BtnRow)
- `shell/overlays/review_editor.gd`（新增，class_name ReviewEditor extends Control）：show_modal(gid, playtime) + 5星选择 + 500字限 + 提交校验 + DB.put_review + EventBus.review_submitted，按 design.md §9
- 不触碰：DB/EventBus 内部实现

**Constraints:**
- 默认 visible=false；show_modal 时 visible=true；提交/取消后 visible=false
- 提交校验：stars>0 且 playtime≥600.0（SubmitBtn.disabled 实时更新）
- text_changed 截断：超 500 字时截断并更新 caret，CharCount 实时显示 "xxx/500"
- 回填：show_modal 时读 DB.get_review(gid)，若存在则回填 stars 和 text
- 提交：DB.put_review（覆盖语义）+ EventBus.review_submitted.emit(gid) + visible=false
- review_editor.gd ≤120 行；全字段 ThemeTokens.color()

**Acceptance:**
- AC: show_modal(gid, 700.0) → visible=true，TitleLabel 含游戏名
- AC: playtime<600.0 → SubmitBtn.disabled=true
- AC: playtime≥600.0 且 stars=0 → SubmitBtn.disabled=true
- AC: playtime≥600.0 且 stars>0 → SubmitBtn.disabled=false
- AC: 输入超 500 字 → 自动截断，CharCount 显示 "500/500"
- AC: 点提交 → DB.get_review(gid) 存在且 stars/text 正确，status="local"
- AC: 点取消 → visible=false，DB 无写入
- AC: 已有评价时 show_modal → 回填旧 stars/text
- AC: 编辑器零报错；≤120 行

---

### Task 7: ReviewsPage ⬜

**复杂度**: 中

**依赖**: Task 6

**Scope:**
- `shell/pages/reviews.tscn`（新增）：BackButton + VBox(Header: AvgLabel+WriteBtn / ListView / EmptyState)
- `shell/pages/reviews.gd`（新增，extends Page）：on_enter(data={gid}) + _render() + 监听 review_submitted/review_deleted，按 design.md §10
- 不触碰：ReviewEditor/DB 内部实现

**Constraints:**
- on_enter 和 on_resume 均调 _render()（从 ReviewEditor 返回后刷新）
- M1 本地最多 1 条评价（自己的），置顶显示 [修改][删除]
- [修改] = ReviewEditor.show_modal(gid, playtime)（find_child 查找 ReviewEditor）
- **B-1**：[删除] 必须先弹 ConfirmBubble 二次确认（find_child("ConfirmBubble")），确认后调 DB.delete_review + EventBus.review_deleted；找不到 ConfirmBubble 时降级直接删
- **B-2**：[写评价] 按钮置灰条件：playtime<600s **或** finish_count<2（与 result_overlay.REVIEW_RUNS=2 对齐）
- 空时 EmptyState 可见："还没有评价，来抢沙发"
- reviews.gd ≤120 行

**Acceptance:**
- AC: on_enter {gid:"tetra_nova"} → 页面正确加载，无崩溃
- AC: 无评价时 → EmptyState 可见，ListView 不可见
- AC: 有评价时 → 显示星级+文字，[修改][删除] 可见；平均星文案为"你的评分：★N"（P-3）
- AC: 点 [修改] → ReviewEditor.show_modal 被调用（visible=true）
- AC: 点 [删除] → ConfirmBubble 弹出；确认后 DB.get_review(gid)=null；_render() 重建后显示空态（B-1）
- AC: [写评价] playtime<600s 或 finish_count<2 → disabled=true（B-2）
- AC: review_submitted/review_deleted 信号触发 → 自动重渲染
- AC: 编辑器零报错；≤120 行

---

### Task 8: DetailPage 补充评价入口 ⬜

**复杂度**: 低

**依赖**: Task 6, Task 7

**Scope:**
- `shell/pages/detail.tscn`（修改）：ScrollContainer/VBox 末尾添加 `ReviewsBtn: Button`（文案"全部评价 →"）
- `shell/pages/detail.gd`（修改）：on_enter 中绑定 ReviewsBtn.pressed → Nav.push ReviewsPage {gid}
- 不触碰：CtaBar/CTA 状态机；test_detail 现有 23 断言

**Constraints:**
- 连接方式：检查 is_connected 避免重复连接（on_enter 可能多次调用）
- 不改 _cta_bar 引用路径，不改 on_resume 逻辑
- detail.gd 改动 ≤10 行

**Acceptance:**
- AC: 详情页底部出现"全部评价 →"按钮
- AC: 点击 → Nav.push("res://shell/pages/reviews.tscn", {"gid": gid})，ReviewsPage 正确加载
- AC: test_detail 原有 23 断言全绿（不因本改动断言失败）
- AC: 编辑器零报错

---

### Task 9: ResultOverlay 接线 ReviewEditor ⬜

**复杂度**: 低

**依赖**: Task 6

**Scope:**
- `shell/components/result_overlay.gd`（修改）：`_on_review()` 由 Toast 改为调用 `ReviewEditor.show_modal(gid, playtime)`，按 design.md §12
- 不触碰：_fill_review_button 灰态逻辑（不改）；其余按钮逻辑

**Constraints:**
- 通过 `get_tree().root.find_child("ReviewEditor", true, false)` 查找（与 _toast 保持同模式）
- **P-1**：playtime 必须来自本局 result dict（`_result_playtime` 成员变量，show_card 时赋值），不得用 DB.total_playtime
- result_overlay.gd 改动：新增成员变量 `var _result_playtime: float = 0.0`，show_card 中赋值 `_result_playtime = float(result.get("playtime", 0.0))`；_on_review 中传 `_result_playtime`
- 改动 ≤15 行

**Acceptance:**
- AC: [✎ 评价] 可点时（playtime≥600s + finish_count≥2）→ ReviewEditor.visible=true
- AC: [✎ 评价] 置灰逻辑不变（_fill_review_button 不改）
- AC: test_launch 回归全绿

---

### Task 10: Main.tscn 接线 + Nav 扩展 ⬜

**复杂度**: 中

**依赖**: Task 2, Task 4, Task 5, Task 6, Task 7

**Scope:**
- `shell/main.tscn`（修改）：Services 节点下挂 Searcher 子节点；OverlayLayer 下挂 ReviewEditor 子节点（visible=false）
- `core/nav.gd`（修改）：`_modal_dialog_open()` 改为遍历 OverlayLayer 所有子节点 visible，按 design.md §13
- `services/launcher.gd` 或 `main.gd`（修改）：Main._ready() 中 Registry.reload() 后调用 Searcher.build_index()
- 不触碰：Nav push/pop/switch_tab 主流程；Launcher 内部

**Constraints:**
- Searcher 节点名与 SearchPage.get_node 路径对齐（执行时确认 main.tscn 实际路径）
- ReviewEditor 节点名与 result_overlay/reviews.gd find_child("ReviewEditor") 对齐
- **A-3**：nav.gd `_modal_dialog_open()` 改为在 `_ready()` 中 call_deferred 缓存 `_overlay_layer`（避免每次 get_node），T10 执行前实查 main.tscn 确认 OverlayLayer 实际路径，写入注释；接口签名不变
- Main._ready() 中 Searcher.build_index() 调用时机：Registry.reload() 之后、Nav.switch_tab(0) 之前

**Acceptance:**
- AC: SearchPage `get_node("/root/Main/Services/Searcher")` 不返回 null
- AC: result_overlay find_child("ReviewEditor") 找到正确节点
- AC: ConfirmBubble/PurchaseDialog/ReviewEditor 任一 visible=true 时，返回键被拦截
- AC: Nav push/pop/switch_tab 主流程不受影响（test_nav/test_main 回归全绿）
- AC: 冷启动时 Searcher.build_index() 被调用（SearchPage 可立即查询）
- AC: 编辑器零报错

---

### Task 11: headless 测试 + GUI 双轨 + 全量回归 ⬜

**复杂度**: 高

**依赖**: Task 1 ~ Task 10

**Scope:**
- `tools/test_category.tscn/.gd`（新增）：分类筛选/空态/卡片点击逻辑测试
- `tools/test_search.tscn/.gd`（新增）：索引构建/打分/历史去重/空态测试
- `tools/test_review.tscn/.gd`（新增）：ReviewEditor 提交/校验/回填/删除；ReviewsPage 渲染；ResultOverlay 灰态
- GUI 双轨：neon/elegant × 分类页/搜索页（有结果）/ReviewsPage 截图，read_image 断言
- 全量回归：RG-1~12（沿用）+ RG-13~19（本 CR 新增）+ test_detail/test_launch/test_db×3/test_rg5

**Constraints:**
- headless 可跑（console exe + 隔离 APPDATA，沿用 CR-3 环境备忘）
- DB 操作前 duplicate() 避免引用污染断言
- 测试间独立：每 case 重置相关 DB 字段
- GUI exe 必须隔离 `$env:APPDATA = dev\tmp_gui`；截图前清 tmp_gui\Godot
- 退出码 0，无脚本错误

**Acceptance:**
- AC: test_category：筛选 price_models=["free"] → 正确过滤；空态出现；卡片点击 Nav.push 正确
- AC: test_search：build_index 后 query("tetra") 含 tetra_nova；add_search_history 去重置顶；无结果文案正确；_on_hot_or_hist_pressed bind 语义验证（gid/q 参数顺序正确）
- AC: test_review：ReviewEditor playtime<600s → SubmitBtn.disabled；finish_count<2 → [写评价] disabled（B-2）；提交后 DB.get_review 存在；点[删除]弹 ConfirmBubble，确认后 null（B-1）；ResultOverlay [✎ 评价] 注入本局 playtime 而非 total（P-1）；ReviewsPage 平均星文案为"你的评分：★N"（P-3）
- AC: RG-20：clear_search_history 后正常退出 → 重启后 get_search_history() 返回空数组（A-5）
- AC: GUI 截图：neon/elegant × 分类/搜索/评价 各页面无溢出，ThemeTokens 颜色正确
- AC: RG-1~20 全部通过；test_detail/test_launch/test_db×3/test_rg5 全绿
- AC: 全部测试退出码 0，无脚本错误
