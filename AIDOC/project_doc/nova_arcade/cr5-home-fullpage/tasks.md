# 任务：CR-5 首页完整化（Banner 轮播 + 热门榜 + 每日任务）

> 依据 `design.md`（2026-08-26 确认，D1~D7 全采纳）。格式：三要素（Scope/Constraints/Acceptance）。
> Shell CR 通用 Constraints（hybrid §3.2 + override §2）逐任务隐含生效，不重复列出。

## 进度摘要

| 指标 | 值 |
|------|-----|
| 总任务数 | 9 |
| 已完成 | 9 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 9/9 (100%) |
| 当前阶段 | 已执行（T1~T9 全绿 + GUI 双主题视觉验收通过） |

## 依赖关系

```
T1(EventBus + DB 扩展)
T2(Recommender.charts, 依赖T1)
T3(DailyTaskService, 依赖T1)
T4(SectionBanner组件, 独立)
T5(SectionCharts组件, 依赖T2)
T6(SectionDaily组件, 依赖T3)
T7(home.gd/tscn 接入三区块, 依赖T4+T5+T6)
T8(Main.tscn 接线, 依赖T3+T7)
T9(测试 + 全量回归, 依赖T1~T8)
```

---

### Task 1: EventBus + DB 扩展 ✅

**复杂度**: 低

**Scope:**
- `core/event_bus.gd`（修改）：新增 `signal tasks_updated(tasks: Array[Dictionary])` 信号（预留，本 CR 由 DailyTaskService 自身发，EventBus 版本 M2 接）
- `core/db.gd`（修改）：新增 `get_daily_tasks() -> Dictionary` + `update_daily_tasks(tasks_data: Dictionary)` 两个方法，按 design.md §3
- `core/db.gd` touch_daily()（修改）：跨天重置保留 tasks_data 字段（design.md §3 共存规则，A-1 修复）
- 不触碰：touch_daily() 接口签名；plays/date 语义

**Constraints:**
- update_daily_tasks：防抖写（复用 _debounce_write("daily.json")，与 touch_daily 共用同一文件）
- get_daily_tasks：从 _daily.get("tasks_data") 读取，不存在返回空 dict
- 单行 lambda 禁止（Godot 4.5 限制）

**Acceptance:**
- AC: DB.update_daily_tasks({"progress": {}, "reward_claimed": false}) 后 DB.get_daily_tasks() 返回正确内容
- AC: DB.get_daily_tasks() 在未调用 update_daily_tasks 时返回空 dict，不崩溃
- AC: touch_daily() 跨天重置后 tasks_data 字段保留（RG：写入 tasks_data → 模拟跨天 touch_daily → get_daily_tasks() 仍可读）
- AC: EventBus.tasks_updated 信号已声明，编辑器零报错

---

### Task 2: Recommender.charts() 扩展 ✅

**复杂度**: 中

**依赖**: Task 1

**Scope:**
- `services/recommender.gd`（修改）：新增 `charts(mode)` + 三个比较器 + `_calc_players()` + `_local_rating()`，按 design.md §4
- 不触碰：continue_row() / for_you() 现有逻辑

**Constraints:**
- mode 取值 "all"/"new"/"rated"，未知 mode 按 "all" 处理
- _calc_players：sessions + 10000（Q6 基准）
- 比较器禁止单行 lambda，全部改具名方法
- M1 单款游戏时三模式结果相同，不需要特殊处理，自然收敛
- recommender.gd 行数检查：原 ~50 行 + 新增 ~60 行，若超 200 行拆 charts_helper.gd

**Acceptance:**
- AC: charts("all") 返回非空数组，每条包含 meta/record/rank/players 字段
- AC: charts("all")[0].rank == 1
- AC: charts("all")[0].players >= 10000（base 保证）
- AC: charts("rated") 有本地评价时排序正确（有评分的排前面）
- AC: 编辑器零报错

---

### Task 3: DailyTaskService ✅

**复杂度**: 高

**依赖**: Task 1

**Scope:**
- `services/daily_task_service.gd`（新增，class_name DailyTaskService extends Node），按 design.md §5
- 不触碰：DB/EventBus 内部实现

**Constraints:**
- 非 Autoload；挂 Main/Services 节点下（T8 接线）
- _ready() 中连接 EventBus.game_launched / game_finished
- 跨天检测：_load_or_reset() 对比 DB.get_daily().date 与 today
- 防重复触发：_update_task 的 cap_at_target 参数控制
- reward_claimed=true 后不再弹 Toast（_show_reward_toast 前检查）
- 单行 lambda 禁止；≤150 行

**Acceptance:**
- AC: on_game_launched 调用后 get_tasks()[0].done == true（launch 任务）
- AC: on_game_finished(result={playtime:310.0}) 后 get_tasks()[2].done == true（累计 5 分钟）
- AC: 三任务全 done → _show_reward_toast 调用，reward_claimed=true 写入 DB
- AC: 再次调用完成事件 → 不再触发 Toast（reward_claimed 防重）
- AC: 跨天后 _load_or_reset → get_tasks() 全部 done=false
- AC: 编辑器零报错；≤150 行

---

### Task 4: SectionBanner 组件 ✅

**复杂度**: 中

**Scope:**
- `data/editorial.json`（修改）：banner[] 预置 1 条 `{image_path: "", target_gid: "tetra_nova", title: "TETRA NOVA"}`（FR-1，B-3 修复）
- `shell/components/section_banner.tscn`（新增）：BannerRect(PanelContainer) + TitleLabel + DotsRow(HBoxContainer)
- `shell/components/section_banner.gd`（新增，class_name SectionBanner extends Control），按 design.md §6
- 不触碰：其他组件

**Constraints:**
- banner[] 为空时 visible=false；不为空时 visible=true
- _load_banners() 幂等：重载前 `_items.clear()` + 旧 timer 引用置空（design §12，A-2 修复）
- 单张时不启动 Timer（Q4）；多张时 4s Timer 循环
- 切换动画：淡入淡出 0.3s（Q1），BannerRect modulate.a 0→1
- image_path 为空时用 `ThemeTokens.grad("banner")` + GradientTexture2D 渐变色块（P-2 修复：无 banner_grad_start token；grad() 已存在，M1 即真渐变）
- 点击整体 BannerRect → Nav.push(DetailPage)（target_gid 空则不响应）
- 指示点：当前宽 14px，非激活 5px
- section_banner.gd ≤100 行

**Acceptance:**
- AC: editorial.json banner[] 为空 → SectionBanner.visible=false
- AC: banner[] 含 1 条 → visible=true，Timer 不启动
- AC: banner[] 含 2 条 → visible=true，Timer 启动
- AC: target_gid="" 时点击 BannerRect 不崩溃（不 push）
- AC: 渐变色块颜色走 ThemeTokens（不硬编码）
- AC: 编辑器零报错；≤100 行

---

### Task 5: SectionCharts 组件 ✅

**复杂度**: 中

**依赖**: Task 2

**Scope:**
- `shell/components/section_charts.tscn`（新增）：Header(Title+TabRow) + ListView + EmptyState
- `shell/components/section_charts.gd`（新增，class_name SectionCharts extends VBoxContainer），按 design.md §7（含图标节点，P-1 修复）
- 不触碰：Recommender 内部实现

**Constraints:**
- set_recommender(r) 注入后须立即 `_render()`（is_node_ready 时，A-3 修复）；null 时 EmptyState 可见
- Tab 切换即时刷新，无动画（Q2）
- 列表行用 _make_row() 动态构建，点击走 _on_row_input
- 排名 1/2/3 用 rank1/rank2/rank3 颜色 token（ThemeTokens）
- 人数格式：≥10000 显示"X.X万+"（Q2-format）
- 无评价时不显示★（review 为 null 时跳过 stars_lbl）
- section_charts.gd ≤150 行

**Acceptance:**
- AC: set_recommender(r) 后 _render() 执行，列表行数 = charts() 返回条数
- AC: 点击"新游" Tab → 列表立即重建（无延迟）
- AC: 玩过人数显示"1.0万+"（1款 sessions=0，base=10000）
- AC: 无评价时列表行无★标签
- AC: 点击行 → Nav.push DetailPage，gid 正确
- AC: 编辑器零报错；≤150 行

---

### Task 6: SectionDaily 组件 ✅

**复杂度**: 中

**依赖**: Task 3

**Scope:**
- `shell/components/section_daily.tscn`（新增）：Header(Title+ProgressLabel) + TaskList + CompletedLabel
- `shell/components/section_daily.gd`（新增，class_name SectionDaily extends VBoxContainer），按 design.md §8
- 不触碰：DailyTaskService 内部实现

**Constraints:**
- set_service(svc) 注入，连接 svc.tasks_updated → _render()
- _render() 先 queue_free 旧行再重建
- 全部完成时 CompletedLabel.visible=true（文案"🎉 今日已全部完成"）
- 完成的任务：check="✓"，颜色 ThemeTokens.color("ok")；未完成："○"，颜色 ink3
- ≤80 行

**Acceptance:**
- AC: set_service(svc) 后 TaskList 显示 3 条任务行
- AC: svc 触发 tasks_updated([...done=true...]) → 对应行 check="✓"
- AC: 三任务全 done → CompletedLabel.visible=true
- AC: 编辑器零报错；≤80 行

---

### Task 7: home.gd / home.tscn 接入三区块 ✅

**复杂度**: 中

**依赖**: Task 4, Task 5, Task 6

**Scope:**
- `shell/pages/home.tscn`（修改）：ScrollView/VBox 中按顺序插入 SectionBanner / SectionCharts / SectionDaily 三节点
  - 顺序：SectionBanner（最顶）→ SectionContinue → SectionForYou → SectionCharts → SectionDaily
- `shell/pages/home.gd`（修改）：新增三区块 @onready + _ready() 中注入 _recommender / _daily_svc
- 不触碰：ThemeButton 逻辑；继续游戏/为你推荐区块逻辑

**Constraints:**
- DailyTaskService 访问路径：`get_node_or_null("/root/Main/Services/DailyTaskService")`
- Recommender 路径沿用现有 `/root/Main/Services/Recommender`
- home.gd 行数：原 ~50 行 + 新增 ~20 行，目标 ≤80 行
- 单行 lambda 禁止

**Acceptance:**
- AC: home.gd _ready() 中 _section_banner/_section_charts/_section_daily 均不为 null
- AC: SectionBanner 在首页最顶可见（banner[] 有数据时）
- AC: SectionCharts set_recommender 被调用，热门榜有数据
- AC: SectionDaily set_service 被调用，任务列表可见
- AC: ThemeToggle 功能不受影响（回归）
- AC: home.gd ≤80 行，编辑器零报错

---

### Task 8: Main.tscn 接线 ✅

**复杂度**: 低

**依赖**: Task 3, Task 7

**Scope:**
- `shell/main.tscn`（修改）：Services 节点下新增 DailyTaskService 子节点（type=Node，script=daily_task_service.gd）
- 不触碰：其他节点；nav.gd；main.gd

**Constraints:**
- load_steps 计数需对应更新（+1 ext_resource）
- DailyTaskService 节点名与 home.gd get_node_or_null 路径精确对齐

**Acceptance:**
- AC: home.gd 中 get_node_or_null("/root/Main/Services/DailyTaskService") 不返回 null
- AC: 主场景编辑器可打开，无报错（--import 验证）

---

### Task 9: headless 测试 + 全量回归 ✅

**复杂度**: 高

**依赖**: Task 1 ~ Task 8

**Scope:**
- `tools/test_home.tscn/.gd`（新增）：覆盖 design.md §11 RG-21~30 + FR 验收口径
- 全量回归：test_category/test_search/test_review/test_smoke/test_main/test_nav/test_detail/test_launch/test_db×3/test_rg5 双相

**Constraints:**
- headless 可跑（console exe + 隔离 APPDATA）
- DB 操作前 duplicate() 避免引用污染
- 每 case 重置 DB.daily 数据
- 退出码 0，无脚本错误

**Acceptance:**
- AC: banner[]=[] → SectionBanner.visible=false（RG-21）
- AC: banner[1条] → visible=true，Timer 未启动（RG-22）
- AC: charts("all") 返回 rank=1 且 players≥10000（RG-23）
- AC: DailyTaskService on_game_launched → launch 任务 done=true（RG-24）
- AC: 三任务全完成 → reward_claimed=true（RG-25）；重复触发不再写 Toast（RG-26）
- AC: 跨日后 get_tasks() done 全 false（RG-27）
- AC: target_gid="" 点击不响应（RG-29）；charts 空数组 → EmptyState 可见（RG-30）
- AC: test_home 全绿；全量回归 RG-1~30 全过
- AC: GUI 双主题截图时核对 Banner 标题在霓虹/青瓷两主题下对比度可读（P-3）
- AC: 退出码 0，无脚本错误
