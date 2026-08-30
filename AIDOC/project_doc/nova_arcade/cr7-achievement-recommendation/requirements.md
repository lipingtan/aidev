# 需求：CR-7 M2 盒子级能力收尾（成就墙 + 本地推荐画像）

## 背景

M1 已闭环，CR-6 接入三款小游戏后盒子共 4 款可玩内容。M2「内容扩充」还剩两项盒子级能力未落地：**成就系统**（AchievementEngine + 成就墙 + 全局成就）与**本地推荐画像**（阶段1 本地规则推荐）。event_bus 已有 `achievement_unlocked` 信号、Registry 已有 `ach_def(aid)`、theme 已有 `--ach-*/--gold` token，但无 AchievementEngine/成就页/全局成就数据；Recommender 目前仅 `continue_row()`，`home.gd` 已调用 `for_you()`（空则隐藏区块）。

本 CR 补齐这两项，使 M1+M2 成为「完整可玩、可分发的免费盒子」（design §10 注）。Arcade PoC 真机验证拆出为独立 CR，不在本 CR。

## 用户故事

- 作为玩家，我希望看到自己解锁的成就（按游戏分组 + 全局成就 + 总成就点），以便 有收集/挑战目标
- 作为玩家，我希望解锁成就时有金色 Toast 提示，以便 即时获得正反馈
- 作为玩家，我希望首页「为你推荐」根据我的游玩/成就/评价偏好个性化推荐，而非千篇一律的热榜
- 作为玩家，我希望推荐在无画像时回退到编辑推荐/热榜，以便 冷启动也有内容

## 功能需求

### FR-1: AchievementEngine（盒子级成就判定）
**描述：** 新增 `services/achievement_engine.gd`（≤150 行），挂 Main/Services 节点下（非单例）。订阅 EventBus `game_finished`(result.achievements[]) / `review_submitted` / 每日任务完成，`evaluate(event)` 统一判定游戏级 + 全局成就；解锁后强写 DB + emit `achievement_unlocked`。
**验收标准：**
- WHEN `game_finished` 携带 result.achievements[] 命中某游戏 def THEN 系统 SHALL 标记该成就解锁（若未已解锁）并强写 DB
- WHEN 全局成就条件满足（收藏家=拥有≥5款 / 好评人=提交≥3条评价 / 马拉松=累计时长≥10h）THEN 系统 SHALL 判定解锁并 emit `achievement_unlocked`
- WHEN 同一成就重复触发 THEN 系统 SHALL 不重复解锁/不重复 emit（幂等）
- WHEN evaluate() 执行 THEN 系统 SHALL 无阻塞（O(已解锁)），不新增 EventBus 信号

### FR-2: 全局成就数据 + Registry 统一加载
**描述：** 新增 `data/achievements.json`（全局成就：收藏家/好评人/马拉松，含 id/name/desc/points）；Registry.reload() 统一加载并与游戏级成就共用 `ach_def(aid)` 查询接口。
**验收标准：**
- WHEN Registry.reload() THEN 系统 SHALL 解析 achievements.json 无报错，`ach_def("global_collector")` 返回完整 def
- WHEN 任一新游戏 meta.json 补游戏级成就后 Registry.reload() THEN 系统 SHALL `ach_def(aid)` 命中该游戏级 def（与全局共用接口）

### FR-3: 三款新游戏补游戏级成就定义
**描述：** magic_tower / game_2048 / snake 的 meta.json `achievements[]` 各补 2~3 个轻量成就（进度里程碑 + 隐藏挑战），points 10~30，与 tetra_nova 同构。
**验收标准：**
- WHEN Registry.lookup("magic_tower")/("game_2048")/("snake") THEN 系统 SHALL 返回非空 achievements[]（每款 ≥2 个，含 id/name/points）
- WHEN 游戏 quit_requested.result.achievements[] 携带本局解锁的 aid THEN FR-1 engine SHALL 能据此判定解锁

### FR-4: 我的页「成就墙」区块
**描述：** `shell/pages/library.gd`（当前 CR-1 骨架：标题+EmptyState）新增头部（总成就点 + 累计时长）+ 成就墙区块（总进度条 + 按游戏分组 + 全局成就区），主题色走 `--ach-*`/`--gold`。
**验收标准：**
- WHEN 玩家进入我的页 THEN 系统 SHALL 渲染头部总成就点（=Σ已解锁 points）与累计时长
- WHEN 成就墙区块渲染 THEN 系统 SHALL 按游戏分组展示各游戏成就（已解锁高亮/未解锁置灰）+ 全局成就区，neon/elegant 双主题无布局跳变
- WHEN 无任何解锁成就 THEN 系统 SHALL 显示空态引导（非崩溃）

### FR-5: 成就解锁 Toast
**描述：** 成就解锁走既有 `achievement_unlocked` 信号 → 金色 Toast（`--ach-toast-bg`/`--ach-toast-ink`），与结算卡 900ms 错开防重叠；确保 AchievementEngine 不重复触发既有 Toast 消费者。
**验收标准：**
- WHEN 成就解锁 THEN 系统 SHALL 弹出金色 Toast「🏆 成就解锁：XXX +N 点」（展示 2200ms）
- WHEN 结算卡与成就 Toast 同时触发 THEN 系统 SHALL 错开 ≥900ms 防重叠
- WHEN 同一成就重复触发 THEN 系统 SHALL 不弹多次 Toast

### FR-6: 本地推荐画像（Recommender.for_you）
**描述：** `services/recommender.gd`（扩展后 ≤150 行）新增 `for_you() -> Array[Dictionary]`：读 DB.list_records() + 成就 + 评价构建本地画像，规则加权（类目/标签偏好分）+ 未玩过优先 + 最热兜底，返回 top-N `{"meta","record"}`；纯本地无网络。home.gd 已消费该接口（零改动）。
**验收标准：**
- WHEN 有游玩记录 THEN 系统 SHALL `for_you()` 返回画像加权推荐（同类目/同标签偏好游戏靠前）且优先未玩过游戏
- WHEN 无游玩记录（冷启动）THEN 系统 SHALL 回退编辑推荐/热榜（非空）
- WHEN 画像推荐 + 兜底合并 THEN 系统 SHALL 去重、≤N 条、每条含有效 meta（lookup 命中）
- WHEN `for_you()` 执行 THEN 系统 SHALL 纯本地 DB 读取，无网络调用

## 非功能需求

- 性能：成就 evaluate O(已解锁) 无阻塞；推荐画像读 DB 缓存不每帧重算
- 一致性：主题色全走 ThemeTokens（`--ach-*`/`--gold` 已就位）；存档写入分级遵循 override §2（成就解锁 → 强写；profile → 防抖）
- 兼容性：home.gd/library.gd 现有结构最小改动；Recommender 扩展保持 continue_row() 行为不变
- 目录规范：AchievementEngine ≤150 行、Recommender ≤150 行、成就墙组件 section 80~150 行
