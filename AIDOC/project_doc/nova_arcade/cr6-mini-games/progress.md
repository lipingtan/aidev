# CR-6 小游戏批量集成（魔塔 + 2048 + 贪吃蛇）— 执行进度

> 门控：需求/设计/任务三文档已于 2026-08-28 用户确认（含多角色 review R1~R13，全部修复，见 review_report.md）。
> 执行顺序（依赖）：T1 → T2/T3 → T4/T5 → T6

## 环境速查（继承 CR-5 progress.md）

- headless：`C:\data\developer\devtool\godot\godot4.7\Godot_v4.7.2-stable_win64_console.exe --headless --path projects\nova_arcade\nova-arcade res://tools/xxx.tscn`
- GUI：`Godot_v4.7.2-stable_win64.exe`（同目录）；截图驱动 `tools/_shot_cr5.*` 模式（Main 挂 root + call_deferred add_child）
- headless 前：`$env:APPDATA` 指向临时目录隔离 + 杀孤儿 godot 进程；**每个测试场景须用独立 APPDATA，跨场景共享会串状态误判 FAIL**
- GDScript 4.7 坑（本 CR 踩过）：① lambda 按值捕获局部变量 → 信号回调写回须用成员变量 `_got: Dictionary`；② `Vector2i` 参数按值传递，不能当 out-param，改返回值；③ 无 Python 风格列表推导，须显式 for + `.append()`；④ 魔塔 main.gd 无 `class_name Main`，静态/常量须走实例访问（`src.calc_damage(...)`）

## 执行记录

### T1 魔塔移植 ✅
- [x] `games/magic_tower/`：meta.json + module.tscn + module_adapter.gd（extends GameModule）+ src/{main,main.tscn,touch_view}.gd + assets/（hero/bat/goblin/skeleton/zombie/wolfman/slime/chest/potion/shield/coins/map + buttons A/B/C/D + dpad 十字盘）
- [x] meta.json：id="magic_tower", title="魔塔", subtitle, category="puzzle", aliases=["魔塔","MDX"], pinyin=["motaxie","mtx"], price_model="free"
- [x] 移除 4 个街机实体按钮（R-4/P-1）；回菜单放顶部 HUD 区
- [x] save.cfg 沿用 tetra_nova 格式（best_score 强写立即落盘，R-3/B-3）
- [x] 逻辑纯函数层可 headless：calc_damage(atk,def)（atk<=def→0 / atk>def→>=1）+ MON/ITEM 常量表 + MT_SELFTEST

### T2 2048 新建 ✅
- [x] `games/game_2048/`：meta.json + module.tscn + module_adapter.gd（extends GameModule，tetra 模式）+ src/{game_2048,game_2048.tscn}.gd
- [x] meta.json：id="game_2048", title="2048", subtitle, category="puzzle", aliases=["2048","数字消除"], pinyin=["erling","ersifba"], price_model="free"
- [x] game_2048.gd（299 行 ≤300）：纯函数 grid_merge（全同值/无合并/对角线移动边界）+ 分数累计 + is_game_over；视图层渲染 4×4 网格；键盘方向键 + D-pad
- [x] save.cfg 沿用 tetra_nova 格式

### T3 贪吃蛇新建 ✅
- [x] `games/snake/`：meta.json + module.tscn + module_adapter.gd（extends GameModule，tetra 模式）+ src/{snake_game,snake_game.tscn}.gd + assets/dpad/（复用魔塔十字盘视觉风格）
- [x] meta.json：id="snake", title="贪吃蛇", subtitle="经典街机 · 越吃越长", category="action", aliases=["贪吃蛇","she"], pinyin=["tanchishe","tcs"], price_model="free"
- [x] snake_game.gd（292 行 ≤300）：纯函数层 snake_move(body,dir,food_pos)→new_body+ate_food + is_collision(body)（自撞/撞墙）；视图层蛇身/食物/分数 HUD/死亡遮罩；键盘方向键 + D-pad
- [x] 网格制移动，方向不可反向（向上时按向下忽略）；每吃 N=5 食物提升速度（move_interval 缩短）
- [x] best_score 强写立即落盘（save.cfg，tetra_nova 格式）

### T4 editorial.json 扩展 + icon 生成 ✅
- [x] `data/editorial.json`：banner[] 扩至 4 条 `[tetra_nova, magic_tower, game_2048, snake]`（均衡覆盖，R-1）；featured 标注为预留字段（R-5）；hot_queries 补新游戏关键词
- [x] `tools/_gen_icons.gd`（102 行）+ `_gen_icons.tscn`：headless GDScript Image 脚本，纯 `set_pixel`/`fill`（Godot 4.7 headless 无 draw_* API），FORMAT_RGBA8，渐变底 + 主题 glyph，neon 配色走 ThemeTokens 色值
- [x] 运行生成三款 icon.png：games/{magic_tower,game_2048,snake}/icon.png 均 512×512 RGBA（PIL 核对）

### T5 headless 测试套件 ✅
- [x] `tools/test_magic_tower.tscn/.gd`（53 行 ≤100）：协议 P1~P9 + 逻辑 L1~L3 = **12 断言全绿**（含 MT_SELFTEST）
- [x] `tools/test_2048.tscn/.gd`（75 行 ≤100）：协议 P1~P9 + 逻辑 L1~L4 = **13 断言全绿**（grid_merge 边界/分数累计/is_game_over）
- [x] `tools/test_snake.tscn/.gd`（68 行 ≤100）：协议 P1~P9 + 逻辑 L1~L6 = **15 断言全绿**（snake_move 增长/死亡/方向不可反向、is_collision 边界、速度递增）
- [x] 测试架构（A-3/R-9）：直接 instantiate `games/{id}/module.tscn` → boot(ctx={save_dir:...})，不走 Registry/Launcher 全链路；独立 APPDATA 隔离

### T6 全量回归 + GUI 双主题验收 ✅
- [x] 全量 headless 回归 **17 场景独立 APPDATA 全绿（exit 0）**：test_2048/test_category/test_db/test_detail/test_home/test_launch/test_magic_tower/test_main/test_nav/test_registry/test_review/test_search/test_smoke/test_snake/test_sound/test_theme/test_viewport
- [x] **执行期修复**：`tools/test_registry.gd` 陈旧断言更新（puzzle query size 1→3，all() size 1→4）；删除 `games/_test_free/`、`games/_test_paid/` 未跟踪残留 fixture（test_detail.gd 运行时自建自删，磁盘副本为污染）
- [x] GUI 双主题验收（tools/_gui_cr6.*，非 headless 渲染）：neon/elegant × 首页 Banner 轮播 + SectionContinue 多条目 + 三款游戏分别 launch→游戏画面→quit_to_shell→结算卡弹出（score/playtime 正确），FAIL=0 exit 0
- [x] `tools/_verify_cr6_query.tscn/.gd`：Registry.query() + Searcher 命中三款新游戏（category/tags/aliases/pinyin）全绿，FAIL=0 exit 0
- [x] GUI 截图路径：`C:\data\developer\studio\games\aidev\dev\tmp_t6_gui_{neon,elegant}\shot_*.png`（home / magic_tower_game / game_2048_game / snake_game / *_result）

## 设计解释记录（执行期决策）

1. **魔塔 main.gd 无 class_name**：测试须走实例访问 `src.calc_damage(...)`，不能 `Main.calc_damage(...)`。
2. **Godot 4 lambda 按值捕获**：quit_requested / run_over 信号回调写回结果用成员变量 `_got: Dictionary`（局部变量被按值拷贝，回调内赋值不生效）。
3. **Vector2i 按值传递**：spawn_food 等改返回值（`-> Vector2i`），不能用 out-param。
4. **Array.duplicate() 返回无类型 Array**：typed `Array[Vector2i]` 复制须显式 `_copy()` helper。
5. **test_registry 断言随 CR-6 更新**：新增三款游戏后 puzzle category=3（tetra_nova+magic_tower+game_2048，snake 是 action），all()=4；循环变量 `m`→`gm` 避免与 tetra_nova 查找回调变量名冲突。
6. **test_rg5 排除在回归计数外**：两阶段测试（须 `--rg5-phase=write|verify`），非标准单跑场景，不计入 17 场景。

## 验收结果汇总

- AC: Registry.reload() lookup 三款新游戏返回完整 GameMeta ✅
- AC: 三款逻辑规则 headless 正确（calc_damage / grid_merge / snake_move）✅
- AC: 键盘方向键 + D-pad 均可控制 ✅
- AC: quit_requested 含 score/playtime/achievements/extra，结算卡弹出 ✅
- AC: best_score 写入 ctx.save_dir/save.cfg ✅
- AC: banner[] 4 条新游戏条目（target_gid 均可 lookup）✅
- AC: 三款 icon.png 存在（512×512 PNG）✅
- AC: 首页 Banner 轮播出现新游戏条目（GUI 截图核对，neon/elegant）✅
- AC: 17 场景 headless 全绿（独立 APPDATA）✅
- AC: GUI 双主题截图通过（neon/elegant 首页/Banner/SectionContinue/结算卡均正常，无 overflow/布局跳变）✅
- AC: 三款游戏分别走通 launch → play → quit_to_shell → result_card 完整链路 ✅
- AC: Registry.query() 能命中三款新游戏（category/tags/aliases/pinyin）✅

## 遗留项（M2）

- banner[] image_path 非空时加载真实 Texture（M1 一律渐变色块，target_gid 走 meta.icon 占位）
- featured 字段为预留，M2 接推荐位
- min_rating / is_new 过滤器 M1 占位透传（GameMeta 无 rating 字段）
