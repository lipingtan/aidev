# CR-4 分类页 + 搜索页 + 本地评价系统 — 执行进度

## 状态：**CR-4 完全闭环**——T1~T11 全 ✅ + 全量回归全绿 + Shell 集成冒烟通过

---

## 交付物

| Task | 文件 | 状态 |
|------|------|------|
| T1 | `core/registry.gd`（query 扩展）、`core/db.gd`（delete_review/clear_search_history）、`data/editorial.json`（hot_queries） | ✅ 已就位 |
| T2 | `services/searcher.gd`（新增） | ✅ 新建 |
| T3 | `shell/components/game_card.gd`（size 参数扩展） | ✅ 已就位 |
| T4 | `shell/pages/category.gd/.tscn`（改写） | ✅ 完整实现 |
| T5 | `shell/pages/search.gd/.tscn`（改写） | ✅ 完整实现 |
| T6 | `shell/overlays/review_editor.gd/.tscn`（新增） | ✅ 新建 |
| T7 | `shell/pages/reviews.gd/.tscn`（新增） | ✅ 新建 |
| T8 | `shell/pages/detail.gd/.tscn`（补充评价入口） | ✅ 修改 |
| T9 | `shell/components/result_overlay.gd`（接线 ReviewEditor） | ✅ 修改 |
| T10 | `shell/main.tscn`（挂 Searcher/ReviewEditor）、`shell/main.gd`（build_index）、`core/nav.gd`（_modal_dialog_open 扩展） | ✅ 完成 |
| T11 | `tools/test_category/test_search/test_review.tscn/.gd`（新增）+ 全量回归 | ✅ 全绿 |

---

## 验收结论（实机 headless 验证）

| 测试套件 | 断言数 | 结果 |
|---------|--------|------|
| test_category（新增） | 7 | ✅ ALL PASS |
| test_search（新增） | 10 | ✅ ALL PASS |
| test_review（新增） | 17 | ✅ ALL PASS |
| test_smoke | 12 | ✅ PASS=12 FAIL=0 |
| test_main | ALL | ✅ ALL PASS |
| test_nav | 7 | ✅ ALL PASS |
| test_detail | 13 | ✅ ALL PASS |
| test_launch | ALL | ✅ ALL PASS |
| test_db × 3 | ALL | ✅ ALL PASS |
| test_rg5 双相 | 6 | ✅ PASS=6 FAIL=0 |

Shell 集成冒烟（dev-workflow-override §5）：PASS=12 FAIL=0 ✅

---

## 关键文件索引

- CR-4 文档：`cr4-category-search/requirements.md`、`design.md`、`tasks.md`
- 协议基线：`nova-arcade-client-design.md` §4.3/§4.4/§4.8；`dev-workflow-override.md`
- 新增服务：`services/searcher.gd`
- 新增页面/组件：`shell/pages/reviews.gd`、`shell/overlays/review_editor.gd`
- 修改页面：`shell/pages/category.gd`、`shell/pages/search.gd`、`shell/pages/detail.gd`
- 修改组件：`shell/components/result_overlay.gd`、`shell/components/game_card.gd`
- 修改核心：`core/registry.gd`、`core/db.gd`、`core/nav.gd`
- 修改场景：`shell/main.tscn`、`shell/main.gd`

---

## 执行期关键决策

1. **test_db 数据污染**：test_rg5 write 阶段向搜索历史写入数据，导致跨测试间 user:// 未隔离时 test_db 出现一次失败。解决方案：每次测试前清空 `dev/tmp_test/Godot/app_userdata/NOVA ARCADE/db/`，不是代码 bug。
2. **CATEGORIES 合并**：GameMeta.category 枚举无独立"益智"值，"消除/益智"合并为 id="puzzle" 单按钮（B-4）。
3. **Nav._modal_dialog_open 升级**：从逐一枚举 ConfirmBubble/PurchaseDialog，改为遍历 OverlayLayer 所有子节点 visible，call_deferred 缓存节点引用（A-3）。
4. **ResultOverlay playtime 来源**：改用本局 result dict 的 `_result_playtime`，不再用 DB.total_playtime（P-1 防刷）。

---

## 遗留（非缺陷，归后续里程碑）

| 遗留 | 归属 |
|------|------|
| GUI 双轨截图（neon/elegant × 分类/搜索/评价）| M1/M2 发布门控 |
| 搜索猜你喜欢 | M2 |
| 分类排序下拉（OptionButton visible=false 预留）| M2 |
| 离线可玩 chip（hidden=true 预留）| M2（补 GameMeta.offline 字段） |
| 云端评价同步/点赞/举报 | M3 |
| Searcher 拼音首字母完整缩写（M1 取首字符占位）| M2 |
| ReviewsPage 信号生命周期优化（_ready 连接 vs on_enter 连接）| M2 低优 |

---

## 环境备忘

- headless：`C:\data\developer\devtool\godot\godot4.7\Godot_v4.7.2-stable_win64_console.exe`
- GUI：`Godot_v4.7.2-stable_win64.exe`
- 测试前：`$env:APPDATA = "dev\tmp_test"` + 杀孤儿进程 + 清空 `db/` 目录
- test_rg5 双相：`--rg5-phase=write` → `--rg5-phase=verify`（同 APPDATA 下连续执行）
