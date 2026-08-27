# CR-5 首页完整化（Banner 轮播 + 热门榜 + 每日任务）— 执行进度

> 门控：需求/设计/任务三文档已于 2026-08-27 用户确认（含多角色 review R1~R11，全部修复）。
> 执行顺序（依赖）：T1 → T2/T3 → T4/T5/T6 → T7 → T8 → T9

## 环境速查（继承 CR-2 progress.md）

- headless：`C:\data\developer\devtool\godot\godot4.7\Godot_v4.7.2-stable_win64_console.exe --headless --path projects\nova_arcade\nova-arcade res://tools/xxx.tscn`
- GUI：`Godot_v4.7.2-stable_win64.exe`（同目录）；截图驱动 `tools/_shot_cr5.*` 模式（Main 挂 root + call_deferred add_child）
- headless 前：`$env:APPDATA` 指向临时目录隔离 + 杀孤儿 godot 进程；**每个测试场景须用独立 APPDATA，跨场景共享会串状态误判 FAIL**
- 新增 class_name 文件需先跑一次 `--headless --import` 刷新全局类缓存

## 执行记录

### T1 EventBus + DB 扩展 ✅
- [x] core/event_bus.gd：+`signal tasks_updated(tasks: Array[Dictionary])`（预留，M2 接）
- [x] core/db.gd：+`get_daily_tasks()` / `update_daily_tasks()`；touch_daily() 跨天重置保留 tasks_data（design §3 共存规则，A-1）
- [x] test_home touch_daily 共存断言全绿

### T2 Recommender.charts() ✅
- [x] services/recommender.gd：+`charts(mode)` + 三个具名比较器（hot/new/rating）+ `_calc_players()`（sessions+10000，Q6）；返回 `Array[Dictionary]{meta, record}`
- [x] 100 行（≤200 无需拆分）；continue_row()/for_you() 零改动

### T3 DailyTaskService ✅
- [x] services/daily_task_service.gd（99 行，class_name extends Node，非 Autoload）
- [x] _ready 连 EventBus.game_launched/game_finished；跨天检测 _load_or_reset；cap_at_target 防重；reward_claimed=true 后不再弹 Toast
- [x] +`toast_shown_count` 测试钩子（RG-25/26）

### T4 SectionBanner 组件 ✅
- [x] data/editorial.json：banner[] 预置 1 条 `{image_path:"", target_gid:"tetra_nova", title:"TETRA NOVA"}`（B-3）
- [x] shell/components/section_banner.tscn/.gd（99 行，class_name extends Control）
- [x] image_path 空 → `ThemeTokens.grad("banner")` + GradientTexture2D 真渐变色块（P-2：grad() 已存在，M1 即真渐变）
- [x] 单张不启动 Timer（Q4）；多张 4s 循环；指示点当前 14px/非激活 5px（token accent/ink3）
- [x] **执行期修复**：① `set_corner_radius(3)` 4.7 需 2 参 → `set_corner_radius_all(3)`；② BannerRect 由 Panel 改 TextureRect（Panel 无 texture 属性，赋值崩）；③ `_load_banners` 须 `call_deferred`（Registry.reload 在 autoload _ready，同步读 editorial 为空）；④ stretch_mode=2 是 KEEP 非 SCALE → 0（否则渐变只占 64×64 左上角）

### T5 SectionCharts 组件 ✅
- [x] shell/components/section_charts.tscn/.gd（110 行，class_name extends VBoxContainer）
- [x] set_recommender() 末尾必调 _render()（A-3）；三 Tab 即时刷新；rank1/2/3 token；人数 ≥10000 → "X.X万+"；图标 TextureRect（空 icon 隐藏）

### T6 SectionDaily 组件 ✅
- [x] shell/components/section_daily.tscn/.gd（56 行，class_name extends VBoxContainer）
- [x] set_service() 连 svc.tasks_updated → _render()；先 queue_free 旧行再重建；全完成 CompletedLabel "🎉 今日已全部完成"

### T7 home.gd / home.tscn ✅
- [x] home.tscn（83 行，load_steps=7）：ScrollView/VBox 顺序 SectionBanner→SectionContinue→SectionForYou→SectionCharts→SectionDaily
- [x] home.gd（70 行 ≤80）：@onready 三区块 + _ready 注入 /root/Main/Services/{Recommender,DailyTaskService}
- [x] **执行期修复**：card.setup() 须 add_child 之后调（@onready 节点在入树时同步 _ready，先 setup 后入树 → "text on Nil"）；同修 category.gd:106 既有隐患

### T8 Main.tscn ✅
- [x] main.tscn：Services 下 +DailyTaskService（type=Node，script=daily_task_service.gd），load_steps 同步

### T9 headless 测试 + 全量回归 + GUI 双主题 ✅
- [x] tools/test_home.tscn/.gd（207 行）：RG-21~30 + FR 验收 + touch_daily 共存 = 27 断言 ALL PASS exit 0
- [x] **执行期修复**：① `.instantiate()` 须 `as SectionBanner/SectionCharts` 类型转换（否则后续方法调用解析失败致挂起）；② lambda 按值捕获局部变量 → tasks_updated 信号改用成员变量 `_signal_fired`；③ RG-28/T8 测试须把 Main 挂到 `get_tree().root`（绝对路径 /root/Main/Services... 才解析，挂在 TestRoot 下必失败）
- [x] 全量回归 14 场景独立 APPDATA 全绿：test_db(3相)/test_home(27)/test_category/test_detail/test_launch/test_main/test_nav/test_registry/test_review/test_search/test_smoke(12)/test_sound/test_theme/test_viewport
- [x] GUI 双主题视觉验收（tools/_shot_cr5.*）：neon/elegant × home 截图 + overflow=none + Banner 标题对比度 neon 12.38 / elegant 9.80（均 >4.5，P-3 通过）
- [x] 截图路径：`C:\data\run\hermes\cr5_neon_home.png` / `cr5_elegant_home.png`（banner crop 同名 _banner_crop.png）

## 设计解释记录（执行期决策）

1. **DailyTaskService 非 Autoload**，挂 Main/Services 节点树；home.gd 经 `get_node_or_null("/root/Main/Services/DailyTaskService")` 取引用（design §5 约束）。
2. **SectionBanner 用 TextureRect 承载渐变**（非 PanelContainer）：Panel 系无 texture 属性，GradientTexture2D 只能挂 TextureRect；TitleLabel/DotsRow 为同级兄弟节点。
3. **_load_banners 延迟执行**：Registry.reload() 在 autoload _ready 内跑，页面组件 _ready 早于其完成，同步读 editorial 必空 → call_deferred。
4. **card.setup() 时序**：instantiate → add_child（触发 @onready）→ setup()；home.gd 与 category.gd 统一此顺序。
5. **测试隔离**：跨场景共享 APPDATA 会串 DB/主题状态误判 FAIL，每场景独立目录；test_db 三相须同目录连续跑。

## 遗留项（M2）

- banner[] image_path 非空时加载真实 Texture（M1 一律渐变色块）
- EventBus.tasks_updated 信号 M2 由 DailyTaskService 改走 EventBus 广播
- charts "rated" 模式依赖本地评价数据，M1 单款游戏自然收敛为 all
