# 需求计划：CR-4 分类页 + 搜索页 + 本地评价系统

> 依据 `nova-arcade-client-design.md` §4.3/§4.4/§4.8 + `nova-arcade-design.md` §3.2/§3.3/§3.5/§7 + 里程碑路由 M1 剩余 CR。
> 类型：Shell CR（工程 `projects/nova_arcade/nova-arcade/`）。
> CR-3 已交付：首页继续游戏+为你推荐区块、GameCard 组件、DetailPage+CTA 状态机、Nav 模态拦截。
> 本 CR 补齐 M1 最后三块：分类页完整交互、搜索页完整交互、本地评价系统（ReviewEditor+ReviewsPage），实现 M1 全部页面闭环。

---

## 需求理解

**目标**：让用户从任意入口（分类浏览/搜索/详情页评价入口）完成完整游戏发现与评价链路，不依赖后端，全部本地运行。

**范围**：
1. **CategoryPage**（`pages/category.tscn/.gd`）：左侧分类栏 + 顶部筛选 chips + 右侧 2 列网格 + 点击→DetailPage；筛选/排序实时刷新。
2. **SearchPage**（`pages/search.tscn/.gd`）：输入框（去抖 150ms）+ 热搜词+历史行 + 实时结果列表 + 点击→DetailPage + 历史管理。
3. **Searcher 服务**（`services/searcher.gd`）：内存索引构建 + 打分查询（标题前缀/包含/别名/拼音/标签），对应 client §4.4。
4. **ReviewsPage**（`pages/reviews.tscn/.gd`）：评价列表 + 排序筛选 + 写评价入口；从 DetailPage"全部评价→"跳转（M1 暂不实现 DetailPage 内精选评价区块，仅在 DetailPage 补"全部评价"入口按钮）。
5. **ReviewEditor**（`overlays/review_editor.tscn/.gd`）：5星选择 + 文字输入 + 提交校验（≥600s 门槛/星级必选/500字限制）；DB.put_review 本地落盘，覆盖更新（不新增）。

**不在范围**：
- Banner 轮播（M2）、热门榜区块（M2）、每日任务区块（M2）
- 云端评价同步（M3）、点赞/举报（M3）
- 搜索页猜你喜欢中的「为你推荐」冷启动（本 CR 无结果时简单空态文案）
- 分类页排序下拉（M2，M1 默认最热）
- LibraryPage 内评价入口（M2 成就/订单系统一起实现）

---

## 假设列表

- [假设-1] 目录仅 tetra_nova 一款，搜索/分类结果最多 1 条，足够验证全链路逻辑。
- [假设-2] Searcher 拼音全拼/首字母缩写：meta.json 已有 `aliases`/`pinyin` 字段（CR-3 未触碰 meta.json 内容），M1 使用已有数据；若字段缺失则退化为纯标题/别名匹配，不阻塞本 CR。
- [假设-3] 评价门槛 ≥600s playtime 在 ReviewEditor 内由 `DB.get_record(gid).total_playtime` 判定（不依赖 Launcher 运行时状态）。
- [假设-4] DetailPage 内精选评价区块（2条）M1 暂不实现，仅添加"全部评价 →"导航按钮（Nav.push ReviewsPage）；此改动不破坏 test_detail 现有断言。
- [假设-5] 分类页排序下拉在 M1 隐藏（visible=false），代码已预留接口，M2 解锁。
- [假设-6] 搜索热搜词 M1 读 `data/editorial.json`（CR-1 已建该文件），新增 `hot_queries` 字段（数组，最多 6 条）。

---

## 澄清问题

- [Question-1] Searcher 索引构建时机：冷启动时 `Registry.reload()` 之后立即构建（写入 `Searcher.build_index()`），还是懒加载（首次输入时构建）？
  - 业界最佳实践：游戏量 <1k 时全量索引内存占用极小，冷启动一次构建、搜索即时响应是最优体验；懒加载会让用户首次输入有感知延迟。
  - 推荐答案及理由：**冷启动即构建**（Main._ready() 步骤 2 Registry.reload() 之后紧跟 Searcher.build_index()）。理由：本 CR 游戏数 ≤10，构建 <1ms；后续 M2 多游戏也不超过 50 款，不需要懒加载；避免搜索页首帧卡顿。
  [Answer-1] 采纳推荐：冷启动即构建（2026-08-25 确认）
懒加载（首次输入时构建）
- [Question-2] 分类页筛选 chips 状态持久化：用户切到其他 Tab 再返回时，是否保留上一次的筛选状态（如已选「免费」+「离线可玩」）？
  - 业界最佳实践：App Store/Google Play 分类页返回后筛选状态保留（当前 session 内），有助于用户继续探索；但保存到磁盘则过度设计。
  - 推荐答案及理由：**session 内保留**（on_exit 时缓存 _filter_state 到 CategoryPage 成员变量，on_resume 时恢复；Tab 切换不 queue_free 页面则天然保留，无需额外代码）。理由：Nav 已有 switch_tab 清栈回 Tab 根页语义，Tab 根页节点常驻，筛选态自然保留；重新进入 Tab = on_resume，不是 on_enter，无需重置。
  [Answer-2] 采纳推荐：session 内保留，无需额外持久化（2026-08-25 确认）
同意
- [Question-3] 搜索历史的去重与上限：每次点击结果调 `DB.add_search_history(q)` 时，若 q 已在历史中应置顶还是忽略？上限 20 条时溢出行为？
  - 业界最佳实践：搜索历史置顶最近搜过的词是通用做法（微信/淘宝），上限溢出时删最旧一条。
  - 推荐答案及理由：**置顶 + 去重**：已存在的 q 先删旧位置再插头部；超 20 条时删尾部。DB 层已有 `add_search_history` 接口（client §2.1），本 CR 需确认其内部逻辑是否已实现去重置顶，若未实现则在本 CR 修复（DB 接口不变，内部逻辑补全）。
  [Answer-3] 采纳推荐：置顶+去重，DB.add_search_history 内部逻辑本 CR 补全确认（2026-08-25 确认）
同意
- [Question-4] ReviewsPage 的进入时机：M1 仅从 DetailPage"全部评价→"进入，还是结算卡的 [✎ 评价] 按钮也进入 ReviewEditor（而非 ReviewsPage）？
  - 业界最佳实践：结算卡直接弹 ReviewEditor（模态，点评完回结算卡）效率更高；ReviewsPage 则是浏览场景。
  - 推荐答案及理由：**结算卡→ReviewEditor（模态）；DetailPage→ReviewsPage→ReviewEditor**。结算卡 [✎ 评价] 走 `ReviewEditor.show_modal(gid)` 直接覆盖结算卡（OverlayLayer 子节点，与 ResultOverlay 同级，playtime 从 result.playtime 注入）；DetailPage"全部评价→"走 `Nav.push(ReviewsPage, {gid})`。两条路径写入同一 DB 条目（DB.put_review 覆盖语义）。
  [Answer-4] 采纳推荐：结算卡→ReviewEditor 模态；DetailPage→ReviewsPage→ReviewEditor（2026-08-25 确认）
同意
- [Question-5] 搜索无结果时「猜你喜欢」区块：M1 使用 Recommender.for_you() 还是简单显示"暂无匹配，试试其他关键词"空态文案？
  - 业界最佳实践：无结果+引导是商店类搜索的标准模式；但 for_you 在本 CR 只有 1 款游戏时与热搜重叠。
  - 推荐答案及理由：**M1 仅空态文案**（"没有找到「xx」"）；为你推荐区块 M2 多游戏后再补。理由：避免 1 款游戏场景下重复展示同一 tetra_nova，且不引入额外 Recommender 依赖使 SearchPage 独立可测。
  [Answer-5] 采纳推荐：M1 仅空态文案，猜你喜欢 M2 补（2026-08-25 确认）
同意
- [Question-6] ReviewEditor 的 OverlayLayer 挂载：与 PurchaseDialog/ConfirmBubble 同级（OverlayLayer 子节点），还是动态实例化挂 root？
  - 业界最佳实践：所有模态弹窗走统一 OverlayLayer，避免 z-order 竞争和返回键拦截混乱。
  - 推荐答案及理由：**OverlayLayer 静态子节点**（与 PurchaseDialog 同挂载方式，Main.tscn 中默认 visible=false，调用 show_modal 时显示）；CR-4 在 main.tscn 的 OverlayLayer 下新增 ReviewEditor 节点；同时 Nav.`_modal_dialog_open()` 扩展检查 ReviewEditor.visible（防返回键关闭穿透）。
  [Answer-6] 采纳推荐：OverlayLayer 静态子节点，Nav 模态检查扩展 ReviewEditor.visible（2026-08-25 确认）
同意
---

## 非功能需求建议

- **性能**：本地目录 ≤50 款（M2 目标），筛选/搜索内存即时（<16ms）；Searcher 全量重建 <5ms。
- **主题**：CategoryPage/SearchPage/ReviewsPage/ReviewEditor 全字段 ThemeTokens.color()；双主题截图验收（GUI exe + 隔离 APPDATA）。
- **存档分级**：search_history → 防抖写 500ms；reviews → 强写（评价写入立即落盘，防数据丢失）；records 改动（last_played 等）延续 CR-3 已有规则。
- **测试分层**：headless test_category + test_search + test_review 覆盖核心逻辑；GUI 双轨截图覆盖主题渲染。

## 影响范围预判

- **涉及模块**：shell/pages（category/search 改写为完整实现、reviews 新增）、shell/overlays（review_editor 新增）、services/searcher（新增）、core/db（add_search_history 去重逻辑补全若缺）、data/editorial.json（新增 hot_queries 字段）、shell/main.tscn（挂 Searcher + ReviewEditor 节点）、core/nav.gd（模态检查扩展）、shell/pages/detail.gd（新增"全部评价→"入口按钮）、services/result_overlay.gd（[✎ 评价] 接线至 ReviewEditor）。
- **可能的副作用**：main.tscn 改动 → test_smoke/test_main 回归；detail.gd 改动 → test_detail 回归；result_overlay.gd 改动 → test_launch 回归。
