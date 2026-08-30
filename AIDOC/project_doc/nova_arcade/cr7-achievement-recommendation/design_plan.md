# 设计计划：CR-7 M2 盒子级能力收尾（成就墙 + 本地推荐画像）

## 设计方向

两部分独立设计，共享回归基线。

### ① 成就墙（AchievementEngine + 数据 + UI）

**核心策略：** AchievementEngine 挂 Services 节点下（非单例），订阅 `game_finished` / `review_submitted` / `tasks_updated`，`evaluate(event)` 统一判定游戏级/全局成就。现有信号零新增：
- `game_finished(result.achievements[])` → 游戏级成就判定
- `review_submitted(gid)` → 好评人全局成就判定
- `tasks_updated(tasks[])`（DailyTaskService 自身信号）→ 每日任务完成事件源（可选，Answer-4 提及"每日任务完成"但 TASK_DEFS 无成就映射，本 CR 仅做 game_finished + review_submitted 两条线，每日任务完成顺延至 M3 云端化时再接入）
- `achievement_unlocked(def, points)` → Toast 消费者（result_overlay.gd 已接线）

**数据层：**
- `data/achievements.json` — 全局成就定义（收藏家/好评人/马拉松），Registry.reload() 统一加载
- 三款新游戏 meta.json 补 `achievements[]`（魔塔/2048/贪吃蛇各 2~3 个）
- Registry.ach_def(aid) 扩展：除搜 games/*/meta.json 外，也搜索 data/achievements.json
- db.gd `_achievements` + `unlock(aid)` 已就位（强写），本 CR 零改动

**UI层：**
- library.gd 从 CR-1 骨架扩展为完整「我的」页：头部（总成就点 + 累计时长）+ 成就墙区块（按游戏分组 + 全局成就区）
- 金色 Toast 复用 result_overlay.gd 既有逻辑（`achievement_unlocked` → Toast），本 CR 确保 Engine 不重复触发

### ② 本地推荐画像（Recommender.for_you 扩展）

**核心策略：** `for_you()` 从「字母排序」升级为「画像加权 + 未玩过优先 + 最热兜底」：
1. 读 DB.list_records() 构建用户画像（游玩偏好/评价偏好/成就偏好）
2. 对所有游戏计算画像匹配分（同类目 + 同标签加权）
3. 未玩过优先排序
4. 冷启动（无记录）回退编辑推荐/热榜
5. home.gd 零改动（已消费 for_you() 接口）

## 技术选型

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| A: Engine 挂 Services 节点 + 订阅既有信号 | 与 Recommender/Searcher 同构；零新增 EventBus 信号 | Engine 需手动接入（Main.tscn 加节点） | ✓ |
| B: AchievementEngine 做 Autoload | 全局可访问，无需手动接入 | 增加加载时间；现有服务均非单例，打破一致性 | ✗ |
| C: 成就判定嵌入 game_finished 信号 handler | 少一个文件 | 事件总线耦合业务逻辑；多消费者时难维护 | ✗ |

## 澄清问题

- [Question-1] Engine 接入方式？
  - 推荐答案：**挂 Main/Services 节点下**（非单例），Main.tscn 新增子节点。与 Recommender/Searcher 同构。
  [Answer-1] A: Services 节点下，2026-08-28

- [Question-2] 每日任务完成是否本 CR 接入成就判定？
  - 推荐答案：**暂不接入**。TASK_DEFS 无成就映射字段，且 Answer-4 提及"每日任务完成"作为未来扩展。本 CR 仅 game_finished + review_submitted 两条线。
  [Answer-2] 暂不接入，顺延 M3，2026-08-28

- [Question-3] Recommender.for_you() 返回多少条？
  - 推荐答案：**top 4**（与首页一屏宽度匹配；CR-5 的 for_you 返回全量由 UI 截断，本 CR 在推荐层截断更高效）。
  [Answer-3] top 4，2026-08-28

## 风险点

- [Risk-1] Registry.ach_def(aid) 当前仅搜 games/*/meta.json；扩展至 data/achievements.json 需保持接口不变（aid 命名空间：游戏级 = `{game_id}_{suffix}`，全局级 = `global_{name}`，互不冲突）
- [Risk-2] AchievementEngine 与 result_overlay.gd 共用 `achievement_unlocked` 信号——result_overlay 已在 game_finished handler 内做 DB.unlock + emit；Engine 需确保幂等（db.gd.unlock 内部 has(aid) 检查防重）
- [Risk-3] library.gd 扩展后行数可能超 80~150 行上限——成就墙 UI 渲染逻辑拆为独立 section 组件文件
