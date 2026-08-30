1|# 需求计划：CR-7 M2 盒子级能力收尾（成就墙 + 本地推荐画像）
2|
3|> 依据 `nova-arcade-design.md` §10 里程碑 M2「+成就墙 + 本地推荐画像」+ `dev-workflow-override.md` §7 路由（M2：成就/每日任务系统 = Shell CR）。
4|> 类型：**Shell CR**（工程 `projects/nova_arcade/nova-arcade/`，激活层 hybrid-project-workflow §3.2）。
5|> 前置状态：M1 全部闭环；CR-6 三款小游戏已接入（tetra_nova/魔塔/2048/贪吃蛇共 4 款）；event_bus 已有 `achievement_unlocked` 信号、Registry 已有 `ach_def(aid)`，但无 AchievementEngine/成就页/全局成就数据。
6|> 用户决定（2026-08-28）：本 CR = ①成就墙 + ②本地推荐画像；④Arcade PoC Android 真机验证 **拆出为独立 CR**（后续单独排期，不阻塞本 CR）。
7|
8|---
9|
10|## 需求理解
11|
12|**目标**：把 M2「内容扩充」的盒子级能力收尾——补齐盒子级成就系统（成就墙）与本地规则推荐画像，使首页/我的页数据更完整，M1+M2 成为「完整可玩、可分发的免费盒子」（design §10 注）。
13|
14|**范围（两部分）**：
15|1. **成就墙（Shell 层）**：AchievementEngine（监听 EventBus + PlayRecord，`evaluate(event)`）+ `data/achievements.json` 全局成就（收藏家/好评人/马拉松）+ 三款新游戏补游戏级成就定义 + 我的页「成就墙」区块（总进度条 + 按游戏分组 + 全局成就）+ 解锁 Toast（金色，与结算卡 900ms 错开防重叠）+ `db.gd` 成就解锁记录存储
16|2. **本地推荐画像（Shell 层）**：Recommender 从「仅继续游戏」扩展为阶段1本地规则推荐——读 DB 游玩记录/成就/评价构建本地画像，`for_you()` 返回个性化推荐行（最热兜底），首页「为你推荐」区块接入，纯本地无后端
17|
18|**不在范围**：
19|- **Arcade PoC Android 真机验证**（MameRuntime 插件 spike + CoreManager 真实接线）——拆出为独立 CR，单独排期
20|- 打砖块/扫雷/弹幕等更多小游戏（CR-6 已交付 4 款，design 要 4~6 → 达标下限；顺延）
21|- M3 云端化（云评价/云推荐/排行榜/云存档）、M4 商业化（支付SDK/广告SDK）
22|- PCK/HTML 运行时扩展、DLC 管线（M5）
23|
24|## 假设列表
25|
26|- [假设-1] 成就墙数据走 `data/achievements.json`（全局）+ 各游戏 meta.json `achievements[]`（游戏级），Registry.ach_def(aid) 统一查询（已就位，client-design §4.9）。
27|- [假设-2] 成就解锁状态存 `db.gd` 新增记录（gid→[aid] map + 全局成就独立键），强写（解锁即时落盘）；总成就点 = Σ 已解锁 points。
28|- [假设-3] 本地推荐画像纯本地：读 DB.list_records() + 成就 + 评价构建画像，规则推荐（同类目/同标签偏好加权），最热兜底；M3 再切云端。
29|- [假设-4] 首页「为你推荐」区块当前数据源为编辑推荐（editorial.json）或占位，本 CR 接入 `Recommender.for_you()` 后优先展示画像推荐、无画像时回退编辑推荐。
30|- [假设-5] 两部分共享同一回归基线：全量 headless 回归 + GUI 双主题验收（neon/elegant）+ Shell 集成冒烟 §5。
31|
32|## 澄清问题
33|
34|- [Question-1] 成就墙放哪个入口？
35|  - 业界最佳实践：成就墙是「我的」页的固定区块（client-design §4.9 已定：头部总成就点 + 累计时长，下方成就墙区块），非独立 Tab。
36|  - 推荐答案及理由：**我的/Library 页新增「成就墙」区块**（总进度条 + 按游戏分组 + 全局成就区）。理由：design §4.9 已明确布局，独立 Tab 增加导航深度且无 design 依据。
37|  [Answer-1] A: 我的/Library 页固定区块（总进度条 + 按游戏分组 + 全局成就区），2026-08-28
38|- [Question-2] 三款新游戏（魔塔/2048/贪吃蛇）补哪些游戏级成就？
39|  - 业界最佳实践：每款 2~3 个轻量成就（进度里程碑 + 隐藏挑战），与 tetra_nova 的 wave10/boss3 同构。
40|  - 推荐答案及理由：**魔塔**（到达高层/击败Boss）、**2048**（达成 512/2048 tile）、**贪吃蛇**（吃到 N 食物/最长身长）各 2~3 个，points 10~30。理由：与 tetra_nova 成就密度一致，验证「多游戏成就聚合」能力。
41|  [Answer-2] A: 魔塔(到达高层/击败Boss)、2048(达成512/2048 tile)、贪吃蛇(吃到N食物/最长身长)各2~3个，points 10~30，2026-08-28
42|- [Question-3] 本地推荐画像的推荐策略？
43|  - 业界最佳实践：阶段1无后端，用本地游玩/成就/评价信号做规则加权（同类目偏好 + 同标签亲和），最热兜底；避免「只推玩过的」。
44|  - 推荐答案及理由：**`Recommender.for_you()` = 画像加权（类目/标签偏好分）+ 未玩过优先 + 最热兜底**，返回 top-N。纯本地 DB 读取，无网络。理由：design §5 阶段1「本地规则推荐（无后端）」，M3 再切云端个性化。
45|  [Answer-3] A: for_you() = 画像加权(类目/标签偏好分) + 未玩过优先 + 最热兜底，纯本地DB读取，2026-08-28
46|- [Question-4] 成就解锁的判定入口与事件源？
47|  - 业界最佳实践：AchievementEngine 挂 Services 节点，订阅 EventBus `game_finished`（result.achievements[]）+ `review_submitted` + 时长累计，evaluate(event) 统一判定游戏级/全局成就，解锁走 `achievement_unlocked` 信号 → Toast。
48|  - 推荐答案及理由：**Engine 订阅 `game_finished` / `review_submitted` / 每日任务完成**，evaluate() 内判定（游戏级：result.achievements[] 命中 def；全局：游戏数/评价数/时长阈值）；解锁后强写 DB + emit `achievement_unlocked`。理由：client-design §2.2/§3.3 已定信号契约，零新增信号。
49|  [Answer-4] A: Engine 挂 Services 节点下，订阅 game_finished / review_submitted；evaluate() 判游戏级+全局成就，解锁强写DB + emit achievement_unlocked（零新增信号），2026-08-28
50|- [Question-5] 两部分验收标准怎么定？
51|  - 业界最佳实践：与 CR-3/CR-4/CR-5/CR-6 口径一致——headless test suite（logic + lifecycle）+ GUI 双主题 screenshot + Shell 集成冒烟 §5 + full regression.
52|  - 推荐答案及理由：**`tools/test_achievement.tscn`**（engine evaluate/unlock/总点/全局成就判定 ≥10 断言）+ **`tools/test_recommender_profile.tscn`**（for_you 画像加权/未玩过优先/最热兜底/空画像回退 ≥10 断言）+ GUI 双主题（成就墙区块 + 推荐行渲染，neon/elegant）+ Shell 冒烟 §5 + full regression.
53|  [Answer-5] A: test_achievement.tscn (>=10断言) + test_recommender_profile.tscn (>=10断言) + GUI双主题(neon/elegant) + Shell冒烟§5 + full regression，2026-08-28
54|
55|## 非功能需求建议
56|
57|- 性能：成就 evaluate 在 EventBus 回调内 O(1)~O(n)（n=已解锁），无阻塞；推荐画像读 DB 缓存，不每帧重算
58|- 一致性：主题色全走 ThemeTokens（`--ach-*` / `--gold` 已就位）；存档写入分级遵循 override §2（成就解锁 → 强写；profile → 防抖）
59|- 目录：成就墙组件 section 80~150 行；AchievementEngine ≤150 行；Recommender 扩展后 ≤150 行
60|
61|## 影响范围预判
62|
63|- 涉及模块：`nova-arcade/services/`（+achievement_engine.gd, recommender.gd 扩展）、`nova-arcade/data/achievements.json`（新增）、`nova-arcade/games/*/meta.json`（+成就定义 ×3）、`nova-arcade/shell/pages/library.gd`（+成就墙区块）、`nova-arcade/shell/pages/home.gd`（推荐行数据源接入）、`nova-arcade/core/db.gd`（+成就解锁记录）
64|- 涉及文件（预估）：~8 新增 + ~6 修改；Shell core 少量改动（db.gd 成就记录）
65|- 可能的副作用：event_bus achievement_unlocked 已有消费者（Toast），需确保 AchievementEngine 不重复触发；首页「为你推荐」区块数据源切换需回归首页布局
66|