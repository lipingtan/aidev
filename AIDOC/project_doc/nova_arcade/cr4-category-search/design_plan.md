# 设计计划：CR-4 分类页 + 搜索页 + 本地评价系统

> 依据 `requirements.md`（Q1~Q6 全部确认，2026-08-25）。
> 类型：Shell CR。工程实查前置：需读取现有代码结构再回答设计问题。

---

## 设计前提（工程现状速查）

基于 CR-1~CR-3 已交付：
- `core/registry.gd`：已有 `all()` / `lookup(gid)` / `query({category,…})` 接口（query 现有实现需确认是否支持多条件 AND 过滤）
- `core/db.gd`：已有 `get_review/put_review/add_search_history/get_search_history` 接口（add_search_history 内部去重逻辑待确认）
- `shell/components/`：已有 `game_card.tscn/.gd`（CR-3 交付，可复用）
- `services/result_overlay.gd`：已有 [✎ 评价] 按钮（CR-2 交付，按钮已置灰逻辑，需接线 ReviewEditor）
- `shell/pages/detail.gd`：已有基础信息块（CR-3 交付，需补"全部评价→"按钮）
- `core/nav.gd`：已有 `_modal_dialog_open()` 检查 ConfirmBubble/PurchaseDialog（CR-3 扩展，需再扩展 ReviewEditor）
- `data/editorial.json`：已存在（CR-1 建立），需新增 `hot_queries` 字段

---

## 设计澄清问题

- [Question-1] Registry.query() 接口现状：CR-1/CR-3 实现的 `query(filter: Dictionary)` 是否已支持多字段 AND 过滤（category + price_model + tags + min_rating + is_new）？还是需要本 CR 扩展？
  - 业界最佳实践：分类页所有筛选条件在 Registry 层内存过滤，O(n) 遍历对 ≤50 款游戏完全够用；接口设计为 `query(filter)` 传 dict，调用方无需关心实现细节。
  - 推荐答案及理由：**本 CR 扩展 Registry.query() 支持完整筛选 dict**（`{category, price_models:[], min_rating, tags:[], sort, is_new}`），若现有实现已部分支持则只补缺失字段，接口签名不变（扩展 filter key），调用方向后兼容。
  [Answer-1] 采纳推荐并修正：实查确认 query() 已支持 category/tag/price，**不支持** price_model 多选/min_rating/is_new；本 CR 扩展三个缺失维度，接口向后兼容（2026-08-25 确认）
同意
- [Question-2] GameCard 在分类页 2 列网格中的尺寸：CR-3 交付的 GameCard 是否已有多规格（S/M/L）？分类网格用哪个规格？
  - 业界最佳实践：分类页 2 列网格用方形卡（图标为主，名称+状态角标），对应 client-design §8.4 的 CardS 规格；client-design §4.3 明确"右侧 2 列网格卡片"。
  - 推荐答案及理由：**复用 GameCard 并传入 size="grid" 参数**（若 CR-3 GameCard 只有单一尺寸，则本 CR 在 setup() 中加 size 参数控制布局，不新建组件）。CategoryPage 通过 GridContainer（columns=2）排列卡片。
  [Answer-2] 采纳推荐并修正：实查 GameCard 只有单一尺寸；本 CR 在 setup() 加 size 参数（"card"默认/"grid"网格），grid 模式固定宽度适配 2 列；不新建组件（2026-08-25 确认）
同意
- [Question-3] Searcher 首字母缩写提取：pinyin 字段存全拼数组（如 `["xinxingfangzhen"]`），首字母缩写（"xxfz"）由 Searcher 运行时从全拼提取（取每段首字母）还是 meta.json 预填？
  - 业界最佳实践：client-design §4.4 明确"首字母缩写 Searcher 运行时自动从全拼提取（取每段首字母），无需单独字段"。
  - 推荐答案及理由：**Searcher 运行时提取**：`build_index()` 中对每条 pinyin 全拼串按空格/段分割取首字母拼成缩写串，存入内存索引，不写回 meta.json。
  [Answer-3] 采纳推荐：Searcher 运行时提取首字母缩写，不写回 meta.json（2026-08-25 确认）
同意
- [Question-4] ReviewsPage 评分分布条：M1 目录仅 1 款游戏最多 1 条评价，评分分布条（5 档 BarChart）是否实现？
  - 业界最佳实践：App Store 详情评分分布条数据来自云端聚合；M1 本地只有 1 条评价时分布条无意义且易误导。
  - 推荐答案及理由：**M1 不实现评分分布条**（代码占位 visible=false，标注 M3 云端聚合数据到来后显示）；仅展示"平均星 / 评价数"文本。M1 ReviewsPage 头部只有星级文本 + [写评价] 按钮。
  [Answer-4] 采纳推荐：M1 不实现评分分布条（visible=false 占位，标注 M3）；ReviewsPage 头部仅星级文本+[写评价]（2026-08-25 确认）
同意
- [Question-5] SearchPage 输入框去抖实现：Godot 4.5 中 LineEdit 的 text_changed 信号 + SceneTreeTimer 去抖，还是 Tween 延迟？
  - 业界最佳实践：搜索去抖标准实现：text_changed 触发时取消旧 Timer、创建新 SceneTreeTimer(0.15s)，到期执行查询；Godot 4.5 SceneTreeTimer 用 `get_tree().create_timer(0.15)` + `await timer.timeout` 实现（CR-2 progress.md 经验：SceneTreeTimer 为 RefCounted，无 queue_free）。
  - 推荐答案及理由：**SceneTreeTimer 去抖**：`text_changed` 时先 `_debounce_timer` 若存活则取消（set a flag），再 create_timer(0.15)，await timeout 后执行 Searcher.query。注意经验#1：验证场景用 tscn 不用 `-s`，去抖测试用 await 模拟时序。
  [Answer-5] 采纳推荐：SceneTreeTimer(0.15s) 去抖，text_changed 时 flag 取消旧 Timer（2026-08-25 确认）
同意
- [Question-6] ResultOverlay 现有 [✎ 评价] 置灰逻辑（CR-2 注记2）：`playtime<600s → 「还需X分钟」；finish_count<2 → 「还需N局」`，本 CR 接线 ReviewEditor 时是否改变这个判定逻辑？
  - 业界最佳实践：ReviewEditor 弹出后自行做二次校验（playtime≥600s 才可提交），ResultOverlay 的置灰是前置 UX 引导，两层校验互补。
  - 推荐答案及理由：**不改变 ResultOverlay 置灰逻辑**，仅在原来"置灰+文案"的代码上补连线：按钮可点时 → `ReviewEditor.show_modal(gid, result.playtime)`；ReviewEditor 内部再做提交前 playtime 校验（二次防护）。接线不改 CR-2 已有的灰态文案逻辑。
  [Answer-6] 采纳推荐并修正：实查确认 _on_review() 目前只弹 Toast，本 CR 改为调 ReviewEditor.show_modal(gid, playtime)；置灰逻辑（_fill_review_button）不变（2026-08-25 确认）
同意
- [Question-7] CategoryPage 分类栏选中"全部"时的过滤行为：category="" 还是特殊值？Registry.query 如何区分"全部"vs"具体分类"？
  - 业界最佳实践：filter dict 中 category 键缺失或为空字符串表示"不过滤分类"；具体分类传对应字符串（"puzzle"/"action" 等）。
  - 推荐答案及理由：**category="" 表示全部**（不传 category 或传空字符串，Registry.query 内部：category 为空则不过滤该维度）；分类栏"全部"按钮选中时 `_filter.category = ""`，其他按钮选中时传对应枚举值（与 GameMeta.category 字段一致）。
  [Answer-7] 采纳推荐：category="" 表示全部，Registry.query 内部 category 为空跳过该维度过滤（2026-08-25 确认）
同意
- [Question-8] ReviewEditor 覆盖更新语义：DB.put_review(gid, dict) 是否已是覆盖（PUT）语义，还是追加？M1 需要"一人一游戏一条"。
  - 业界最佳实践：client-design §2.1 明确 `put_review(gid, dict)` 语义；设计文档 §7 明确"一人一游戏一条 active 记录（PUT 覆盖更新）"。
  - 推荐答案及理由：**put_review 已是覆盖语义**（按 gid 键存储，再次 put 覆盖旧值）；本 CR 执行时实查 db.gd 代码确认，若非覆盖则修复（接口签名不变）。删除操作用 `DB.delete_review(gid)` 或 put 一个 `{status:"deleted"}` 标记——执行前确认 DB 是否有 delete_review，若无则本 CR 补充。
  [Answer-8] 采纳推荐并修正：实查 put_review 已是覆盖语义 ✅；delete_review **不存在**，本 CR 在 db.gd 补充 `delete_review(gid)` 方法（从 _reviews 删键 + 强写落盘）（2026-08-25 确认）
同意
---

## 非功能设计约束

- 所有新增页面/组件严格遵守 override §2：禁用 hdr_2d/glow，颜色全走 ThemeTokens.color()
- 单文件 ≤200 行（Searcher 若超限拆 searcher_index.gd）
- 新增 Autoload 节点：无；Searcher/ReviewEditor 均非 Autoload
- Main._ready() 步骤顺序：DB.load_all → Registry.reload → **Searcher.build_index()（新增）** → PayService.retry → Nav.switch_tab(0)

---

## 修订记录

- 2026-08-25 初版（D1~D8 待用户确认后输出 design.md）
