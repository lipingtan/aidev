# 需求：CR-4 分类页 + 搜索页 + 本地评价系统

> 依据 requirements_plan.md（Q1~Q6 全部采纳推荐，2026-08-25 确认）。协议基线：client-design §4.3/§4.4/§4.8 + nova-arcade-design §3.2/§3.3/§3.5/§7 + override §2/§5 + hybrid §3.2。
> CR-1~CR-3 已交付：Shell 骨架 + Launcher + GameModule + 首页入口链路 + DetailPage+CTA。
> 本 CR 补齐 M1 最后三块，实现全部页面闭环：分类浏览 / 搜索发现 / 本地评价。

---

## 背景

CR-3 后 CategoryPage / SearchPage / ReviewsPage 仍是空骨架（EmptyState 占位）。Searcher 服务未实现。本地评价（ReviewEditor + ReviewsPage）未落地，ResultOverlay 的 [✎ 评价] 按钮未接线。M1 目标：无后端依赖，用户可通过分类/搜索发现游戏并写评价，全链路本地闭环。

---

## 已确认决策（2026-08-25 全部采纳推荐）

| # | 决策 |
|---|------|
| Q1 | Searcher 索引冷启动即构建（Main._ready Registry.reload 后立即 build_index），避免首次输入延迟 |
| Q2 | 分类筛选状态 session 内保留（Tab 根页常驻，on_resume 自然保持，无需额外持久化） |
| Q3 | 搜索历史置顶+去重：已存在的 q 删旧位置插头部，超 20 条删尾部；本 CR 确认/补全 DB.add_search_history 内部逻辑 |
| Q4 | 评价双入口：结算卡 [✎ 评价] → ReviewEditor 模态（playtime 从 result.playtime 注入）；DetailPage"全部评价→" → ReviewsPage → ReviewEditor |
| Q5 | 搜索无结果 M1 仅空态文案"没有找到「xx」"，猜你喜欢 M2 补 |
| Q6 | ReviewEditor 挂 OverlayLayer 静态子节点（默认 visible=false），Nav._modal_dialog_open() 扩展检查 ReviewEditor.visible |

---

## 用户故事

- 作为玩家，我点底部分类 Tab，左侧选「益智」后右侧只显示益智类游戏；选「免费」chip 后进一步过滤；筛选无结果时看到空态提示和"清除筛选"按钮。
- 作为玩家，我点搜索 Tab，输入"tetra"后实时看到 TETRA NOVA；点击进入详情；历史记录保留，再次进搜索页看到上次的词，可以清空。
- 作为玩了 10 分钟的玩家，退出游戏时结算卡 [✎ 评价] 按钮可点（playtime≥600s）；写 3 星评价提交后立即在评价列表可见。
- 作为玩家，我从详情页点"全部评价→"进入评价列表，看到自己的评价置顶，可以修改或删除。

---

## 功能需求

| # | 需求 |
|---|------|
| FR-1 | **Searcher 服务**（`services/searcher.gd`）：`build_index()` 从 Registry.all() 构建内存索引（每游戏：title/title_lower/aliases[]/pinyin[]/tags[]）；`query(text) -> Array[String]`（gid 列表）打分规则：标题前缀 +100、标题包含 +40、别名包含 +35、拼音全拼前缀 +30、拼音首字母前缀 +25、标签命中 +15，去抖调用方负责（SearchPage 内 150ms 去抖），Searcher.query 同步返回；`hot_queries() -> Array[String]`（读 data/editorial.json 的 hot_queries 字段，最多 6 条）；非单例，挂 Main/Services 节点下 |
| FR-2 | **data/editorial.json 新增 hot_queries 字段**：数组，M1 预置 ≤6 条本地热搜词（如 `["TETRA NOVA", "俄罗斯方块", "消除", "益智", "街机", "roguelike"]`） |
| FR-3 | **CategoryPage**（`pages/category.tscn/.gd`）：左侧竖向分类栏（全部/消除/益智/动作/街机/休闲/Roguelike）+ 顶部筛选 chips（"全部"互斥清除其余；免费/付费/离线/评分≥4.0/新游 多选叠加）+ 右侧 2 列网格（复用 GameCard 组件）；点击卡片 → Nav.push(DetailPage, gid)；排序下拉 M1 隐藏（visible=false 预留 M2）；数据来源 Registry.query({category, …filters}) 内存过滤即时刷新 |
| FR-4 | **CategoryPage 空态**：筛选无结果时显示插画（大 emoji）+ "该筛选下暂无游戏" + 内联"清除筛选"按钮（重置分类=全部、chips=全部清除） |
| FR-5 | **SearchPage**（`pages/search.tscn/.gd`）：进入自动聚焦输入框；展示热搜词（editorial.json hot_queries，最多 6 条）+ 历史行（DB.get_search_history()，最多 20 条，可整体清空）；输入去抖 150ms → Searcher.query → 实时列表（图标/名称/★评分/标签行/状态角标）；点击结果 → Nav.push(DetailPage, gid) + DB.add_search_history(q)；无结果 → 空态文案"没有找到「{q}」"（M1 不做猜你喜欢）；[取消] 按钮 → Nav.pop() 回上一页 |
| FR-6 | **DB.add_search_history 去重置顶**：已存在 q → 删旧位置插头部（最近搜索置顶）；超 20 条 → 删尾部；本 CR 确认实现，接口签名不变 |
| FR-7 | **ReviewsPage**（`pages/reviews.tscn/.gd`）：从 DetailPage"全部评价→"入口进入（Nav.push，data={gid}）；头部展示平均星/总数/[写评价] 按钮；列表展示 DB.get_review(gid)（M1 仅本地 1 条，自己的评价置顶带 [修改][删除]）；playtime<600s 时 [写评价] 按钮置灰；列表空时显示"还没有评价，来抢沙发" + 可用时 [写评价] |
| FR-8 | **ReviewEditor**（`overlays/review_editor.tscn/.gd`）：5 星选择 + 文字输入框（500 字限，实时计数）+ [提交] 按钮；提交校验：stars>0 且 playtime≥600s（否则置灰）；`show_modal(gid, playtime)` 弹出（OverlayLayer 子节点，visible=false→true）；提交 → DB.put_review(gid, {stars, text, playtime_at_review, created_at, status:"local"})（强写）→ EventBus.review_submitted(gid) → 关闭 modal；修改 = 覆盖同一条并刷新 created_at；[删除] → DB 删除条目 → EventBus.review_deleted(gid) |
| FR-9 | **DetailPage 补充**：在基础信息块下方添加 [全部评价 →] 导航按钮（Nav.push ReviewsPage，data={gid}）；M1 不渲染精选评价 2 条（标注 M2）；此改动不破坏 test_detail 现有 23 断言 |
| FR-10 | **ResultOverlay 接线**：[✎ 评价] 按钮接线至 ReviewEditor.show_modal(gid, result.playtime)；playtime<600s 时按钮置灰（文案"还需 X 分钟"）；playtime≥600s 且 finish_count<2 时置灰（文案"还需 N 局"）——沿用 CR-2 design §3 解读注记2 逻辑 |
| FR-11 | **Main.tscn 接线**：Services 节点下挂 Searcher 非单例节点；OverlayLayer 下挂 ReviewEditor 静态节点（visible=false）；Nav._modal_dialog_open() 扩展检查 ReviewEditor.visible |
| FR-12 | **headless 测试**：`tools/test_category.tscn/.gd`（分类筛选/空态/卡片点击）+ `tools/test_search.tscn/.gd`（索引构建/打分/历史去重/空态）+ `tools/test_review.tscn/.gd`（ReviewEditor 提交/校验/覆盖/删除；ResultOverlay 按钮灰态）；全部 headless 可跑，退出码 0 |
| FR-13 | **GUI 双轨截图 + 全量回归**：neon/elegant × 分类页/搜索页（有结果+无结果）/ReviewsPage 截图 read_image 断言（无溢出，ThemeTokens 合规）；RG-1~12 + test_detail/test_launch/test_db×3/test_rg5 全量回归 |

---

## 非功能需求

- **性能**：本地目录 ≤50 款（M2），Searcher.query <5ms；CategoryPage 筛选内存即时（<16ms）
- **主题**：CategoryPage/SearchPage/ReviewsPage/ReviewEditor 全字段 ThemeTokens.color()，双主题截图验收
- **存档分级**：reviews → 强写（立即落盘）；search_history → 防抖写 500ms（沿用 DB 现有分级）
- **单文件 ≤200 行**，中文注释，类型标注完整
- **页面不含业务规则**：筛选逻辑内聚 Registry.query，打分逻辑内聚 Searcher，评价校验内聚 ReviewEditor

---

## 范围外（明确不做）

- 搜索猜你喜欢（M2）、分类排序下拉（M2）、Banner/热门榜/每日任务（M2）
- 云端评价同步/点赞/举报（M3）、评价敏感词过滤（M3）
- LibraryPage 内评价入口（M2）、AchievementEngine 评价成就（M2）
- Analytics 上报（M3）、分页（M3，本地目录无需分页）

---

## 验收口径（WHEN-THEN 摘要）

1. WHEN 分类页选「益智」+ chip「免费」THEN 右侧只显示 price_model=free 的益智游戏；无结果时出空态+清除筛选按钮
2. WHEN 搜索输入"tetra"THEN 150ms 后实时列表出现 TETRA NOVA；点击进详情后历史记录出现"tetra"置顶
3. WHEN 再次进搜索页输入已有历史词 THEN 该词置顶不重复；历史超 20 条时最旧条目自动删除
4. WHEN 游玩 <600s 退出 THEN 结算卡 [✎ 评价] 置灰，文案提示还需时长
5. WHEN 游玩 ≥600s 且 finish_count≥2 退出 THEN [✎ 评价] 可点；提交 3 星后立即在 ReviewsPage 可见（status=local）
6. WHEN ReviewsPage 点 [修改] THEN 覆盖同一条记录，created_at 刷新，列表置顶
7. WHEN ReviewsPage 点 [删除] THEN DB 条目移除，列表变空态
8. WHEN DetailPage 点"全部评价→"THEN ReviewsPage 220ms 右滑入，gid 正确
9. WHEN GUI 双主题截图（分类/搜索/评价 × neon/elegant）THEN read_image 断言无溢出、配色走 ThemeTokens
10. WHEN 全量回归（RG-1~12 + 各套件）THEN 全过
