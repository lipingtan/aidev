# 需求计划：CR-5 首页完整化（Banner 轮播 + 热门榜 + 每日任务）

> 依据 `nova-arcade-client-design.md` §4.2 + `nova-arcade-design.md` §3.1/§5/§10。
> 类型：Shell CR（工程 `projects/nova_arcade/nova-arcade/`）。
> CR-1~CR-4 已交付：Shell 骨架 + 四Tab + 首页继续游戏/为你推荐两区块。
> 本 CR 补齐首页剩余三区块：**Banner 轮播** + **热门榜 Top10** + **每日任务**，完成首页五区块全量实现。

---

## 需求理解

**目标**：让首页呈现完整的推荐流体验，不依赖后端，全部本地数据驱动。

**范围**：
1. **Banner 轮播**：3张横幅，4s 自动轮播，激活指示点；点击跳转 DetailPage；数据来自 `editorial.json.banner[]`。
2. **热门榜**：综合/新游/好评 三个子 Tab；列表行：排名+图标+名+★+人数；点击→DetailPage；数据来自 `Recommender.charts(mode)`。
3. **每日任务**：本地日期自然日重置；任务完成打勾；全部完成→成就点+Toast；数据来自 `DB.get_daily()` / `DB.touch_daily()`。
4. **editorial.json** 补充 banner 配置字段（M1 为空数组，本 CR 补数据结构）。
5. **DailyTaskService**（或直接在 home.gd 中实现）：每日任务业务逻辑（内置任务定义、进度更新、重置检测）。

**已有基础**（CR-3/CR-4 交付）：
- 首页继续游戏 + 为你推荐两区块已运行
- `Recommender.charts()` 接口已在 client-design 中定义，但 services/recommender.gd 目前只实现了 `continue_row()` + `for_you()`
- `DB.get_daily()` / `DB.touch_daily()` 已在 db.gd 中实现

---

## 假设列表

- [假设-1] Banner M1 目录只有 tetra_nova 一款游戏，banner[] 最多 1 条有效数据（或用编辑占位）；轮播组件仍需支持 0~N 张，0 张时隐藏整个 Banner 区块。
- [假设-2] 热门榜数据源：M1 无云端数据，`charts(mode)` 用本地规则计算（参见 client-design §7：综合=hot_score、新游=version 日期加权、好评=rating×count）；M1 仅 tetra_nova 一款时三个子 Tab 内容相同，但结构完整，M2 多游戏后自动生效。
- [假设-3] 每日任务 M1 内置固定任务集（如：开玩一局、完成一局），不依赖后端下发；任务完成回调通过 EventBus.game_finished 信号驱动。
- [假设-4] 每日任务奖励 M1 仅 Toast 反馈（成就点逻辑 M2 接入 AchievementEngine 后生效），本 CR 不实现跨游戏成就解锁。
- [假设-5] 下拉刷新（client-design §4.2 标注"下拉刷新"）M1 不实现，标注 M2。

---

## 澄清问题

- [Question-1] Banner 轮播数据格式：editorial.json 的 `banner[]` 每项应包含哪些字段？建议结构是什么？
  - 业界最佳实践：Banner 条目至少包含 `image_path`（图片资源路径）+ `target_gid`（点击跳转游戏，空=不跳转）+ `title`（标题文字，可选）。
  - 推荐答案及理由：`banner: [{image_path:"res://...", target_gid:"tetra_nova", title:""}]`；M1 允许 image_path 为空（Banner 用游戏截图/纯色占位）；target_gid 空时 Banner 不可点。M1 预置 1 条以 tetra_nova 截图为 image_path 的条目验证轮播功能。
  [Answer-1] 采纳推荐：`{image_path, target_gid, title}`；image_path 空时用 banner-grad 渐变色块占位；M1 预置 1 条 tetra_nova 条目（2026-08-26 确认）

- [Question-2] 热门榜子 Tab 点击切换是否需要转场动画？
  - 业界最佳实践：App Store 热门榜 Tab 切换用即时刷新（无动画），避免打断视觉流。
  - 推荐答案及理由：**即时刷新**，无转场。子 Tab 按钮选中态变色，列表直接重建。
  [Answer-2] 采纳推荐：即时刷新，无转场（2026-08-26 确认）

- [Question-3] 每日任务的内置任务集 M1 定义：任务数量和触发条件？
  - 业界最佳实践：每日任务数 3~5 条，触发条件基于已有埋点事件（game_launched/game_finished），避免引入新埋点。
  - 推荐答案及理由：**3 条固定任务**：① 今日启动游戏（触发：game_launched，次数≥1）② 今日完成 1 局（触发：game_finished，次数≥1）③ 今日游玩累计 5 分钟（触发：game_finished 累加 result.playtime≥300s）。任务定义硬编码在 DailyTaskService（或常量），不依赖后端下发（M2 可扩展）。
  [Answer-3] 采纳推荐：3条固定任务（启动/完成1局/累计5分钟），硬编码定义（2026-08-26 确认）

- [Question-4] 每日任务全部完成后的奖励：Toast 文案 + 是否写入 DB？
  - 业界最佳实践：全完成奖励立即反馈（Toast）+ 记录已领取标记（防重复领取）；M1 无成就系统时 Toast 即可。
  - 推荐答案及理由：**Toast「今日任务全部完成 🎉」+ DB.touch_daily() 写入 `reward_claimed=true`**；重启后不再重复弹 Toast（通过 DB.get_daily() 检查 reward_claimed 字段）。
  [Answer-4] 采纳推荐：Toast「今日任务全部完成 🎉」+ DB 写 reward_claimed=true 防重复弹出（2026-08-26 确认）

- [Question-5] Banner 轮播在没有游戏截图资产时的 fallback：纯色占位还是直接隐藏？
  - 业界最佳实践：无图时用游戏主色作为渐变背景占位（比隐藏整个轮播区更好看），类似 App Store 的色块 Banner。
  - 推荐答案及理由：**占位用 ThemeTokens 渐变色块**（霓虹用 `banner-grad` token，青瓷同理）；image_path="" 时不加载 Texture，直接用 StyleBoxFlat 填充渐变背景色，叠加游戏标题文字。这样 Banner 区块始终有内容，不闪烁。
  [Answer-5] 采纳推荐：image_path="" 时用 ThemeTokens banner-grad 渐变色块 + 游戏标题文字叠加（2026-08-26 确认）

- [Question-6] 热门榜行中的"★评分"和"X万人玩过"：M1 目录只有 1 款游戏，这两个字段从哪来？
  - 业界最佳实践：M1 本地无云端评分/人数数据，应降级展示。
  - 推荐答案及理由：★评分来自 **DB.list_reviews()** 本地评价均分（无评价时不显示星）；"X人玩过"来自 **DB.list_records() 的 sessions 字段之和**（记录有多少玩过的设备数；本地只有 1 台设备时等于当前设备是否玩过，显示"1人玩过"或不显示）。M3 云端聚合后替换为服务端字段。
  [Answer-6] 用户答案：M1 本地默认给 1万 基准值；获取到云端数据（M3）后在此基础上叠加累加（2026-08-26 确认）
---

## 非功能需求建议

- **性能**：Banner 轮播用 Tween + Timer，不用 AnimationPlayer；热门榜列表最多 10 条，内存即时渲染；每日任务最多 5 条。
- **主题**：Banner/热门榜/每日任务区块全字段 ThemeTokens.color()；双主题截图验收。
- **存档分级**：daily（每日任务进度）→ 防抖写 500ms（已在 DB.touch_daily 实现）；奖励已领取标记 → 防抖写（不影响核心权益，丢失只导致 Toast 重复）。
- **信号驱动**：每日任务进度更新通过监听 EventBus.game_launched / game_finished 信号，不在游戏内部主动调用 Home 页方法。

## 影响范围预判

- **涉及模块**：shell/pages/home.gd/.tscn（改写扩展）、services/recommender.gd（新增 charts()）、data/editorial.json（补 banner 字段）、新增 services/daily_task_service.gd（每日任务逻辑）、shell/main.tscn（挂 DailyTaskService 节点）。
- **新增组件**：`shell/components/section_banner.tscn/.gd`（Banner 轮播）、`shell/components/section_charts.tscn/.gd`（热门榜）、`shell/components/section_daily.tscn/.gd`（每日任务）。
- **不触碰**：继续游戏/为你推荐两区块（CR-3/CR-4 已交付），Nav/Launcher/DB 内部实现。
