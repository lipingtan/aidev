# 设计计划：CR-5 首页完整化（Banner 轮播 + 热门榜 + 每日任务）

> 依据 `requirements.md`（Q1~Q6 全部确认，2026-08-26）。
> 类型：Shell CR。工程实查前置：已读取 recommender.gd / theme_tokens.gd。

---

## 设计前提（工程现状速查）

- `ThemeTokens.GRADS` 中已有 `banner` 渐变 token（neon: 深蓝紫；elegant: 浅绿紫米）✅ 无需新增
- `Recommender` 已有 `continue_row()` / `for_you()`；**尚无 `charts()`** → 本 CR 扩展
- `DB.get_daily()` / `touch_daily()` 已实现，daily 结构 `{date, plays}` → 本 CR 扩展为 `{date, plays, tasks[], reward_claimed}`
- `EventBus` 已有 `game_launched` / `game_finished` 信号 → DailyTaskService 直接连接
- `home.gd` 当前有继续游戏 + 为你推荐两区块，ScrollView/VBox 结构可直接插入新区块
- `section_continue.gd` 组件模式可作为 SectionBanner/SectionCharts/SectionDaily 的开发参考

---

## 设计澄清问题

- [Question-1] Banner 轮播多张切换的动画方式：横向滑动 vs 淡入淡出？
  - 业界最佳实践：移动端 Banner 以横向滑动为主（用户感知连续性更强）；淡入淡出实现更简单。
  - 推荐答案及理由：**淡入淡出**（alpha 0.3s Tween）。理由：M1 横向滑动需 HBoxContainer + 多节点管理，增加复杂度；淡入淡出用单 TextureRect + Tween 即可，代码量小且 headless 可验。
  [Answer-1] 采纳推荐：淡入淡出，0.3s Tween（2026-08-26 确认）

- [Question-2] 热门榜"人数"的精确格式：M1 本地 sessions + 1万基准，如何呈现？
  - 推荐答案：`sessions_total + 10000`，然后格式化：≥10000 显示 "X.X万+"；< 10000 显示原数字。1款游戏 sessions=0 时显示"1万+"；sessions=5 显示"1万+"；sessions=15000 显示"2.5万+"。
  [Answer-2] 采纳推荐：base=10000 + sessions，≥10000 显示"X.X万+"（2026-08-26 确认）

- [Question-3] DailyTaskService 与 home.gd 的通信：DailyTaskService 发信号 → home 重渲染 SectionDaily，还是 SectionDaily 直接监听 EventBus？
  - 推荐答案：DailyTaskService 发 `tasks_updated` 信号 → SectionDaily 连接后自行 `_render()`，职责分离更清晰；home.gd 只负责注入 DailyTaskService 引用。
  [Answer-3] 采纳推荐：DailyTaskService.tasks_updated → SectionDaily._render()（2026-08-26 确认）

- [Question-4] Banner 轮播单张时是否还需要 4s Timer？
  - 推荐答案：**不需要**。单张时禁用 Timer，不循环；多张时启动 Timer。
  [Answer-4] 采纳推荐：单张禁用 Timer，多张启动（2026-08-26 确认）

- [Question-5] daily 数据结构扩展：`DB.touch_daily()` 现在按日期重置并写 `{date, plays}`。扩展为含 tasks[] 后，touch_daily() 接口是否改变？
  - 推荐答案：**touch_daily() 不改接口**，仍负责日期重置检测；DailyTaskService 调用 `DB.update_daily_tasks(tasks_data)` 写入 tasks 进度（新增方法），与 touch_daily 分离，避免互相覆盖。
  [Answer-5] 采纳推荐：新增 `DB.update_daily_tasks(dict)` + `DB.get_daily_tasks() -> Dictionary`，touch_daily 不改（2026-08-26 确认）

- [Question-6] 热门榜 hot_score 计算：M1 只有 tetra_nova 一款，norm() 归一化无意义（只有 1 个值）。如何处理？
  - 推荐答案：M1 直接用原始值排序（不做归一化），games.size()≤1 时直接返回全量。归一化 M2 多游戏后再启用。
  [Answer-6] 采纳推荐：M1 直接原始值排序，size≤1 时不归一化（2026-08-26 确认）

- [Question-7] SectionBanner 的 Banner 高度设计：几 px 合适？
  - 推荐答案：`custom_minimum_size.y = 200`（占首页约 1/7 高度，视觉比例合理）；宽度 SIZE_EXPAND_FILL 充满。
  [Answer-7] 采纳推荐：高度 200px（2026-08-26 确认）

---

## 非功能设计约束

- 单文件 ≤200 行（SectionCharts 若超限拆 charts_row.gd 行组件）
- 所有颜色走 ThemeTokens；banner 渐变用 `ThemeTokens.grad("banner")`（GRADS 已定义）
- 单行 lambda 禁止（Godot 4.5 限制，沿用 CR-4 经验）
- DailyTaskService 非 Autoload，挂 Services 节点下
- 新增 EventBus 信号：`tasks_updated`（DailyTaskService 发）需在 event_bus.gd 声明

---

## 修订记录

- 2026-08-26 初版（D1~D7 全采纳，基于工程实查）
