# 任务：CR-7 M2 盒子级能力收尾（成就墙 + 本地推荐画像）

> 依据 `design.md`（2026-08-28 Review 确认，11项问题全修复）。格式：三要素（Scope/Constraints/Acceptance）。
> Shell CR 通用 Constraints（hybrid §3.2 + override §2）逐任务隐含生效，不重复列出。

## 进度摘要

| 指标 | 值 |
|------|-----|
| 总任务数 | 8 |
| 已完成 | 8 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 8/8 (100%) |
| 当前阶段 | T1-T8 全部完成，CR-7 已闭环 ✅ |

## 依赖关系

```
T1(全局成就数据 + Registry.ach_def扩展)
T2(三款游戏meta.json补成就定义, 与T1独立)
T3(AchievementEngine, 依赖T1: ach_def需覆盖全局成就)
T4(library.gd扩展 + section_achievement_wall + library.tscn, 依赖T1+T3)
T5(toast_layer扩展: color参数+消息队列, 与T3/T4独立)
T6(Recommender.for_you()画像加权, 与T1~T5独立)
T7(headless测试套件: test_achievement + test_recommender_profile, 依赖T3+T6)
T8(全量回归 + GUI双主题验收, 依赖T4+T5+T7)
```

- **关键路径**：T1 → T3 → T4 → T8（成就墙链路）；T6 → T8（推荐链路）
- T1/T2 独立并行
- T3 依赖 T1（Registry.ach_def 需覆盖全局成就）
- T4 依赖 T1+T3（Engine 就绪 + ach_def 完整）
- T5 与 T3/T4 独立（toast_layer 改动不影响 Engine/成就墙逻辑）
- T6 与 T1~T5 完全独立（Recommender.for_you 扩展不依赖成就系统）
- T7 依赖 T3+T6（测试需 Engine + Recommender 就位）
- T8 依赖 T4+T5+T7

---

### Task 1: 全局成就数据 + Registry.ach_def() 扩展

**复杂度**: 低

**Scope:**
- `data/achievements.json`（新增）：全局成就定义数组（收藏家/好评人/马拉松），按 design §3.1 格式
  - `{"id": "global_collector", "name": "收藏家", "desc": "拥有全部游戏记录", "points": 30}`
  - `{"id": "global_reviewer", "name": "好评人", "desc": "提交≥3条评价", "points": 20}`
  - `{"id": "global_marathon", "name": "马拉松", "desc": "累计游玩≥10小时", "points": 50}`
- `core/registry.gd`（修改）：
  - 新增 `_global_achievements: Dictionary = {}` 缓存
  - `reload()` 中加载 `data/achievements.json`（file_exists 检查 + JSON.parse_string 类型校验 + ga.has("id") 字段校验 + push_warning 兜底，见 design §3.2 A-2修复）
  - `ach_def(aid)` 扩展：先查全局成就缓存，再遍历 games/*/meta.json

**Constraints:**
- reload() 错误处理（A-2）：FileAccess.file_exists → parse_string 返回值类型检查 → for ga in ja 循环内 assert ga.has("id") → 解析失败 push_warning 并保留旧缓存
- aid 命名空间隔离：`global_*` vs `{game_id}_*`（design §7）
- Registry 文件行数 ≤200（扩展后复核）
- 不触碰：现有 ach_def 对游戏级成就的查询逻辑（仅前置全局查询）

**Acceptance:**
- AC: `data/achievements.json` 存在且 JSON 格式合法（3条全局成就）
- AC: Registry.reload() 无报错，`_global_achievements` 含 3 条记录
- AC: `Registry.ach_def("global_collector")` 返回完整 def（id/name/desc/points）
- AC: `Registry.ach_def("tn_wave10")` 仍返回 tetra_nova 游戏级成就（原有逻辑不受影响）
- AC: `Registry.ach_def("nonexistent")` 返回 null
- AC: achievements.json 格式异常时 reload() push_warning + 旧缓存保留

---

### Task 2: 三款新游戏 meta.json 补成就定义

**复杂度**: 低

**依赖**: 与 T1 独立（可并行）

**Scope:**
- `games/magic_tower/meta.json`（修改）：achievements[] 补 2 个成就
  - `{"id": "mt_floor50", "name": "勇攀高峰", "desc": "到达第50层", "points": 15}`
  - `{"id": "mt_boss_slayer", "name": "屠龙者", "desc": "击败最终Boss", "points": 25}`
- `games/game_2048/meta.json`（修改）：achievements[] 补 2 个成就
  - `{"id": "g2048_tile512", "name": "初露锋芒", "desc": "合成512方块", "points": 10}`
  - `{"id": "g2048_tile2048", "name": "终极目标", "desc": "合成2048方块", "points": 30}`
- `games/snake/meta.json`（修改）：achievements[] 补 2 个成就
  - `{"id": "sn_eat50", "name": "大胃王", "desc": "吃到50个食物", "points": 10}`
  - `{"id": "sn_length30", "name": "巨蛇", "desc": "身长达到30节", "points": 20}`

**Constraints:**
- 成就格式与 tetra_nova 同构（id/name/points），desc 为可选字段（B-4）
- points 范围 10~30（design §3.3）
- JSON 格式合法，Registry.reload() 解析零报错
- 不触碰：meta.json 其他字段（仅修改 achievements[] 空数组 → 填充）

**Acceptance:**
- AC: Registry.reload() 后 `Registry.lookup("magic_tower")` 返回非空 achievements[]（≥2个，含 id/name/points）
- AC: 同上验证 game_2048/snake
- AC: `Registry.ach_def("mt_floor50")` / `ach_def("g2048_tile2048")` / `ach_def("sn_length30")` 均返回完整 def
- AC: tetra_nova meta.json 零改动（仅修改三款新游戏）

---

### Task 3: AchievementEngine（services/achievement_engine.gd）

**复杂度**: 中

**依赖**: T1（Registry.ach_def 需覆盖全局成就）

**Scope:**
- `services/achievement_engine.gd`（新增，≤150行）：按 design §2.2 类签名
  - `class_name AchievementEngine extends Node`
  - `_enter_tree()` 连接 EventBus.game_finished / review_submitted（A-1修复）
  - `_on_game_finished(gid, result)` → 遍历 result.achievements[] → _try_unlock()
  - `_try_unlock(aid)` → DB.get_achievements().has(aid) 幂等检查 → Registry.ach_def(aid) → DB.unlock() → EventBus.achievement_unlocked.emit()
  - `_on_review_submitted(gid)` → _evaluate_global()
  - `_evaluate_global()` → 遍历 GLOBAL_DEFS → _check_global() → emit achievement_unlocked
  - `_check_global(gdef)` → match id: global_collector (DB.list_records().size() >= maxi(Registry.all().size(), 1)) / global_reviewer (_count_reviews() >= 3) / global_marathon (_total_playtime_hours() >= 10.0)（P-3修复：动态阈值）
  - `_count_reviews()` / `_total_playtime_hours()` 辅助函数
- `shell/main.tscn`（修改）：Services 节点下新增 AchievementEngine 子节点

**Constraints:**
- 行数 ≤150（design §非功能需求）
- 信号订阅用 _enter_tree() 非 _ready()（A-1）
- 幂等保证：_try_unlock 先检查 has(aid) → DB.unlock() 内部再 check → 双重保护
- GLOBAL_DEFS 收藏家阈值动态 = Registry.all().size()（P-3修复）
- total_playtime 为累加字段（B-2确认：launcher_util.finish_record 写入累加值）
- 不新增 EventBus 信号（仅复用 achievement_unlocked）
- 不触碰：现有 Services 节点下其他服务

**Acceptance:**
- AC: main.tscn 含 `Services/AchievementEngine` 节点（type=Node, script=achievement_engine.gd）
- AC: Engine._try_unlock("tn_wave10") → DB.unlock() 被调用 + EventBus.achievement_unlocked.emit() 参数正确
- AC: 同一 aid 调 _try_unlock 3次 → DB.write_json 仅触发 1 次（幂等）
- AC: game_finished result.achievements=["tn_wave10"] → Engine 正确解锁并发射信号
- AC: 收藏家条件判定用动态阈值（Registry.all().size()），非写死 ≥5
- AC: achievement_engine.gd ≤150 行

---

### Task 4: library.gd 扩展 + section_achievement_wall + library.tscn

**复杂度**: 中

**依赖**: T1 + T3（Engine 就绪 + ach_def 完整）

**Scope:**
- `shell/components/section_achievement_wall.gd`（新增，80~150行）：
  - `class_name SectionAchievementWall extends Control`
  - `refresh()` → _clear_children() → _render_progress_bar() → _render_by_game_groups() → _render_global_achievements()
  - _render_progress_bar(): ProgressBar + Label（已解锁/总数 + 百分比）
  - _render_by_game_groups(): 遍历 Registry.all()，有 achievements[] 的游戏建分组标题 + 成就列表（已解锁高亮/未解锁置灰）
  - _render_global_achievements(): 渲染 global_collector/global_reviewer/global_marathon（已解锁高亮/未解锁置灰）
  - 颜色全走 ThemeTokens（--ach-*/--gold，override §2）
- `shell/pages/library.gd`（修改）：
  - 扩展为完整"我的"页：@onready var _header: VBoxContainer + @onready var _ach_wall_root: Control
  - `on_enter(data)` → _render_header() → _section_ach.refresh()
  - `_render_header()` → 总成就点（Σ已解锁 points）+ 累计时长（_calc_total_hours()）→ 渲染到 _header Label
  - `_calc_total_hours()` → 遍历 DB.list_records() 累加 total_playtime / 3600（B-2确认：total_playtime 为累加字段）
- `shell/pages/library.tscn`（修改）：场景树从 CR-1 骨架扩展为完整结构
  - 新增 RootVBox/ScrollView + Header VBoxContainer + AchievementWall Control（script = SectionAchievementWall）

**Constraints:**
- section_achievement_wall.gd 80~150 行（design §非功能需求）
- library.gd ≤80 行（超出则拆分）
- 颜色全走 ThemeTokens.color()（--ach_bg/--ach_line/--gold）
- library.tscn 场景树与 library.gd @onready 引用对齐（P-5修复：预建节点，非运行时 instantiate）
- 空态处理：无解锁成就时显示引导文案（非崩溃）
- 不触碰：home.gd/其他 page 文件

**Acceptance:**
- AC: library.tscn 含 RootVBox/Header/AchievementWall 节点结构
- AC: library.gd on_enter() → _header 显示总成就点 + 累计时长
- AC: AchievementWall.refresh() → 按游戏分组渲染成就列表（已解锁/未解锁样式区分）
- AC: 全局成就区显示 3 个成就（收藏家/好评人/马拉松）
- AC: 无任何解锁成就时显示空态引导文案
- AC: neon/elegant 双主题下无布局跳变

---

### Task 5: toast_layer 扩展（color 参数 + 消息队列）

**复杂度**: 低

**依赖**: 与 T3/T4 独立

**Scope:**
- `shell/components/toast_layer.gd`（修改）：
  - `show_msg(msg, color := null)` 增加可选 color 参数（P-1修复）
  - 新增 `_queue: Array[String] = []` + `_busy: bool = false` 消息队列（P-2修复）
  - _busy 时消息入队，tween.finished → _pop_queue() → 出队显示下一条
  - color != null 时 `_label.add_theme_color_override("font_color", color)`；null 时 remove_override

**Constraints:**
- show_msg() 签名向后兼容（color 默认 null）
- 队列机制不影响原有 300ms 滑入 / 2200ms 停留 / 250ms 滑出时间线
- 不触碰：其他调用 toast_layer.show_msg() 的代码（单参数调用仍有效）
- toast_layer.gd 行数 ≤80

**Acceptance:**
- AC: `show_msg("test")` 无 color 参数 → 正常显示（原有行为不变）
- AC: `show_msg("test", Color(1, 0.824, 0.247))` → Label 字体颜色为金色
- AC: 连续调用 show_msg() 3次 → 消息按序展示，前一条滑出后再显示下一条（队列不丢消息）
- AC: tween.kill() 后队列中待显示消息不受影响

---

### Task 6: Recommender.for_you() 画像加权扩展

**复杂度**: 高

**依赖**: 与 T1~T5 独立

**Scope:**
- `services/recommender.gd`（修改，扩展后 ≤150行）：
  - 替换现有 `for_you()` 实现为画像加权算法（design §5.2）
  - 新增 `_build_profile()` → 遍历 DB.list_records() + Registry.lookup(gid) → 类目/标签偏好分
  - 新增 `_score_for(meta, profile)` → 类目权重 5x + 标签权重 1x
  - 新增 `_fallback_editorial()` → charts("all") 转 `{"meta","record"}` 格式 + slice(0,4)（B-3修复）
  - 新增结果缓存：`_cached_result: Array[Dictionary]` + `_cache_dirty: bool = true` + EventBus.records_updated.connect() 失效（A-3修复）
  - 保留 `continue_row()` / `charts()` 不变

**Constraints:**
- 扩展后 ≤150 行（design §非功能需求）
- for_you() 返回 top 4（推荐层截断）
- 冷启动回退：无游玩记录 → charts("all") 转格式（B-3修复）
- 结果缓存：_cache_dirty flag + records_updated 信号失效（A-3修复）
- 纯本地 DB 读取，无网络调用
- continue_row() / charts() 行为不变（零副作用）
- 不触碰：home.gd（接口签名不变 → home.gd 零改动）

**Acceptance:**
- AC: 有游玩记录时 `for_you()` 返回画像加权推荐（同类目游戏靠前）
- AC: 未玩过游戏优先于已玩过游戏排序
- AC: 冷启动（无记录）`for_you()` 返回非空（charts 兜底，≤4条）
- AC: for_you() 返回 ≤4 条，每条含有效 meta（Registry.lookup 命中）
- AC: 缓存机制：连续调 for_you() 2次 → 第二次直接返回缓存（不调 _build_profile）
- AC: EventBus.records_updated.emit(gid) → 下次 for_you() 重新计算
- AC: recommender.gd ≤150 行

---

### Task 7: headless 测试套件

**复杂度**: 中

**依赖**: T3 + T6（Engine + Recommender 就位）

**Scope:**
- `tools/test_achievement.tscn` + `test_achievement.gd`（新增，≥10断言）：
  - 1-3: Engine._try_unlock() 幂等（同一 aid 调 3 次，DB 仅记录 1 条）
  - 4-5: game_finished result.achievements[] → Engine 正确解锁
  - 6-7: _evaluate_global() 收藏家条件（动态阈值）判定正确
  - 8-9: _evaluate_global() 好评人条件（≥3条评价）判定正确
  - 10: _evaluate_global() 马拉松条件（≥10h时长）判定正确
  - 11: achievement_unlocked 信号发射参数正确（def + points）
- `tools/test_recommender_profile.tscn` + `test_recommender_profile.gd`（新增，≥10断言）：
  - 1-3: _build_profile() 从游玩记录构建类目/标签偏好
  - 4-5: _score_for() 同类目游戏得分高于不同类目
  - 6-7: for_you() 未玩过优先排序
  - 8-9: 冷启动回退（无记录时返回非空，≤4条）
  - 10: for_you() 返回格式正确（每条含有效 meta + record）

**Constraints:**
- 测试架构：直接 instantiate test_achievement.tscn（Node + Engine 节点 + 断言脚本），不走 Main.tscn 全场景（tools/api_probe.tscn 模式）
- save_dir 用独立 APPDATA 隔离路径
- 每款测试脚本 ≤100 行
- headless run quit(0/1) 退出码标记通过/失败

**Acceptance:**
- AC: test_achievement headless run ≥10断言全绿 + exit 0
- AC: test_recommender_profile headless run ≥10断言全绿 + exit 0
- AC: 独立 APPDATA 隔离无交叉污染
- AC: Godot 4.7 console 进程在测试完成后显式退出（get_tree().quit()），无孤儿进程占用服务器资源。CI/运行脚本需验证：测试执行后 `Get-Process -Name "Godot*" -ErrorAction SilentlyContinue` 返回空（或 kill + 确认退出码）。

---

### Task 8: 全量回归 + GUI双主题验收

**复杂度**: 中

**依赖**: T4 + T5 + T7

**Scope:**
- 全量 headless 回归：既有场景（test_home/test_db/test_smoke/test_magic_tower/test_2048/test_snake等）+ 新场景（test_achievement/test_recommender_profile）= **全量场景**，独立 APPDATA 全绿
- GUI 双主题验收（neon/elegant）：
  - library.gd 成就墙页面：总成就点/累计时长显示正确、进度条渲染正常、分组标题可见、已解锁/未解锁样式区分
  - ToastLayer 金色 Toast 显示正常（color 参数生效）
  - home.gd「为你推荐」区块渲染正常（画像推荐/冷启动回退）
- progress.md 记录执行过程

**Constraints:**
- 独立 APPDATA 隔离（每场景一个子目录，防交叉污染——CR-5/6 教训）
- GUI 截图用 Godot_v4.7.2-stable_win64.exe --scene res://shell/main.tscn + 视口截图
- Shell 集成冒烟 §5（override.md）
- 不触碰：现有代码（仅新增/修改 design.md 列出的文件）

**Acceptance:**
- AC: 全量场景 headless 全绿（独立 APPDATA）
- AC: GUI 双主题截图通过（neon/elegant library.gd 成就墙 + home.gd 推荐区块均正常）
- AC: ToastLayer 金色 Toast 显示正确（color 参数生效）
- AC: Shell 集成冒烟 §5 全过

---

## 修订记录

基于 review_report.md（2026-08-28，11项问题全修复），本 tasks.md 已同步修订：
- T3 Scope 补 _enter_tree() 替代 _ready() + 动态收藏家阈值（A-1/P-3）
- T1 Scope 补 Registry.reload() 错误处理（file_exists + parse_string 类型校验 + push_warning）（A-2）
- T4 Scope 补 library.tscn 场景树变更 + section_achievement_wall 纯脚本 class_name（P-5/A-5）
- T5 Scope 补 color 参数 + 消息队列机制（P-1/P-2）
- T6 Scope 补 _fallback_editorial 完整实现 + 结果缓存（B-3/A-3）
- T7 test_achievement 收藏家断言改为动态阈值（P-3）
