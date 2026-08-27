# 需求：CR-5 首页完整化（Banner 轮播 + 热门榜 + 每日任务）

> 依据 requirements_plan.md（Q1~Q6 全部确认，2026-08-26）。协议基线：client-design §4.2/§7 + nova-arcade-design §3.1/§5 + override §2/§5。
> CR-1~CR-4 已交付：Shell 骨架 + 四Tab + 首页继续游戏/为你推荐两区块。
> 本 CR 补齐首页剩余三区块，实现首页五区块全量。

---

## 背景

CR-4 后首页已有继续游戏 + 为你推荐两区块。Banner/热门榜/每日任务三个区块仍是未实现占位。editorial.json.banner[] 为空数组。Recommender 尚无 charts() 实现。DB.get_daily/touch_daily 已实现但未使用。

---

## 已确认决策（2026-08-26）

| # | 决策 |
|---|------|
| Q1 | Banner 条目格式 `{image_path, target_gid, title}`；image_path 空时用 banner-grad 渐变色块占位；M1 预置 1 条 tetra_nova 条目 |
| Q2 | 热门榜子 Tab 切换即时刷新，无动画 |
| Q3 | 每日任务 3 条固定：① 启动游戏（game_launched≥1）② 完成 1 局（game_finished≥1）③ 累计游玩 5 分钟（playtime≥300s） |
| Q4 | 全完成奖励：Toast「今日任务全部完成 🎉」+ DB daily 写入 reward_claimed=true 防重复 |
| Q5 | Banner 无图 fallback：ThemeTokens banner-grad 渐变色块 + 游戏标题文字叠加 |
| Q6 | 热门榜人数：M1 本地基准 1万；M3 获取云端数据后叠加累加；M1 显示"1万+"或"1.X万+" |

---

## 用户故事

- 作为玩家，我打开首页看到 Banner 轮播，4s 自动切换，点击跳入游戏详情。
- 作为玩家，我在热门榜看到综合/新游/好评三个分类，列表展示排名、图标、评分和人气，点击进详情。
- 作为每日登录的玩家，我看到今日任务列表，完成启动/游玩/时长三个目标，全完成后弹 Toast 庆祝。

---

## 功能需求

| # | 需求 |
|---|------|
| FR-1 | **editorial.json Banner 数据**：新增 `banner[]` 数组，每项 `{image_path: String, target_gid: String, title: String}`；M1 预置 1 条（tetra_nova 截图路径或空 image_path，target_gid="tetra_nova"）；`Registry.get_editorial()` 已可读取，无需改接口 |
- FR-2 | **SectionBanner 组件**（`shell/components/section_banner.tscn/.gd`）：从 editorial.json 读取 banner[]；0条时 visible=false；1+ 条时展示轮播；**4s** 自动轮播（SceneTreeTimer 循环）；底部激活指示点（当前宽 14px，非激活 5px）；点击 → `Nav.push(DetailPage, {gid: target_gid})`（target_gid 为空时不响应）；image_path 为空时用 `ThemeTokens.grad("banner")` + GradientTexture2D 渐变色块 + title 文字叠加 |
| FR-3 | **Recommender.charts(mode) 实现**（`services/recommender.gd` 扩展）：`charts(mode: String) -> Array[Dictionary]`，mode 取值 "all"（综合）/"new"（新游）/"rated"（好评）；返回 `[{"meta": GameMeta, "record": Variant, "rank": int}]`；综合=hot_score 排序；新游=version 字符串倒序；好评=本地评价均分倒序（无评价时归到末尾）；最多 10 条；M1 仅 tetra_nova 一款时三模式结果相同 |
| FR-4 | **SectionCharts 组件**（`shell/components/section_charts.tscn/.gd`）：头部三 Tab 按钮（综合/新游/好评）；选中 Tab 即时重建列表；列表行：排名数字 + 图标 + 游戏名 + ★均分（无评价时不显示）+ 人数（本地 sessions 之和 + 1万基准，格式"X.X万+"）；点击行 → `Nav.push(DetailPage, {gid})`；无数据时显示空态 |
- FR-5 | **DailyTaskService**（`services/daily_task_service.gd`，非 Autoload）：内置 3 条任务定义（常量数组）；`get_tasks() -> Array[Dictionary]`：读 DB.get_daily()，跨天则重置进度；`on_game_launched(gid)` + `on_game_finished(gid, result)`：更新对应任务进度，防重复完成；全部完成时 Toast + 写 reward_claimed=true；get_tasks() 返回结构：`{id, label, done: bool, progress: int, target: int}` |
- FR-6 | **DB.daily 结构扩展**：现有 touch_daily() 维护 `{date, plays}`；本 CR 在同一 dict 新增 `tasks_data: Dictionary`（结构 `{progress: {task_id: int}, reward_claimed: bool}`，与 design §3 一致）；touch_daily() 跨天重置时保留/显式清空 tasks_data（共存规则见 design §3），触达逻辑不变 |
| FR-7 | **SectionDaily 组件**（`shell/components/section_daily.tscn/.gd`）：展示 3 条任务（label + 完成勾选状态）+ 进度指示（●●○ 今日2/3 样式）；全部完成时区块底部显示"🎉 今日已全部完成"；每日任务由 DailyTaskService 驱动，组件只负责展示 |
| FR-8 | **home.gd 接入三区块**：首页滚动容器新增 SectionBanner / SectionCharts / SectionDaily 三个区块；顺序：Banner → 继续游戏 → 为你推荐 → 热门榜 → 每日任务；DailyTaskService 通过 Main/Services 节点注入；EventBus.game_launched/game_finished 由 DailyTaskService 在 _ready() 中连接 |
| FR-9 | **Main.tscn 接线**：Services 节点下挂 DailyTaskService 节点 |
| FR-10 | **headless 测试**：`tools/test_home.tscn/.gd`（Banner 0条隐藏/1条可见；charts() 三模式返回非空；每日任务重置/进度更新/全完成 Toast 触发；reward_claimed 防重复）；全量回归 |

---

## 非功能需求

- **性能**：Banner 轮播 Timer 用 SceneTreeTimer，不用 _process；热门榜重建 ≤10 行，内存即时
- **主题**：三区块全字段 ThemeTokens.color()；Banner 色块用 `ThemeTokens.GRADS["banner"]` 停色（已存在，无需新增 token；对应 client-design §8.1 `--banner-grad`）
- **存档**：daily 任务进度防抖写 500ms（touch_daily 已实现），reward_claimed 同级
- **信号驱动**：DailyTaskService 监听 EventBus，home.gd 监听 DailyTaskService 发出的 tasks_updated 信号刷新 SectionDaily

## 范围外（明确不做）

- Banner 轮播自动指示点动画（CSS transition 等效的 Tween，M2 精细化）
- 下拉刷新（M2）
- 热门榜服务端数据（M3）
- 每日任务奖励成就解锁（M2 接 AchievementEngine）
- 每日任务后端下发（M3+）

---

## 验收口径（WHEN-THEN 摘要）

1. WHEN editorial.json banner[] 为空 THEN SectionBanner.visible=false，首页无空白区块
2. WHEN banner[] 含 1 条 target_gid="tetra_nova" THEN Banner 可见，点击跳入 DetailPage
3. WHEN 4s 经过 THEN Banner 自动切换到下一张（单张时不切换）
4. WHEN 热门榜点"好评" Tab THEN 列表按本地评价均分排序，即时刷新无动画
5. WHEN 无评价数据 THEN 热门榜行不显示★，人数显示"1万+"
6. WHEN EventBus.game_launched 发出 THEN 每日任务①进度 done=true
7. WHEN EventBus.game_finished(result.playtime≥300s) THEN 每日任务③进度 done=true
8. WHEN 三条任务均 done THEN Toast「今日任务全部完成 🎉」弹出，reward_claimed=true 写入 DB
9. WHEN 重启应用 THEN reward_claimed=true 时不再重复弹 Toast
10. WHEN 跨自然日重启 THEN 每日任务进度全部重置为 done=false，reward_claimed=false
11. WHEN GUI 双主题截图 THEN Banner 渐变色块/热门榜行/任务勾选 全走 ThemeTokens，无溢出
