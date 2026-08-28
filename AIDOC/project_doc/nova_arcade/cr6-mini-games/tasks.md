# 任务：CR-6 M2 小游戏批量接入（魔塔 + 2048 + 贪吃蛇）

> 依据 `design.md`（2026-08-28 Review 确认，13项问题全修复）。格式：三要素（Scope/Constraints/Acceptance）。
> Shell CR 通用 Constraints（hybrid §3.2 + override §2）逐任务隐含生效，不重复列出。

## 进度摘要

| 指标 | 值 |
|------|-----|
| 总任务数 | 6 |
| 已完成 | 2 |
| 进行中 | 0 |
| 未开始 | 4 |
| 完成率 | 2/6 (33%) |
| 当前阶段 | T1+T2 完成；T3 待执行 |

## 依赖关系

```
T1(魔塔移植: 4.5→4.7.2迁移 + meta.json + adapter)
T2(2048新建: game_2048.tscn/gd + meta.json + adapter)
T3(贪吃蛇新建: snake_game.tscn/gd + meta.json + adapter)
T4(editorial.json扩展 + icon生成, 依赖T1+T2+T3)
T5(headless测试套件: test_magic_tower/test_2048/test_snake, 依赖T1+T2+T3)
T6(全量回归: 17场景独立APPDATA + GUI双主题验收, 依赖T4+T5)
```

- **关键路径**：T1/T2/T3 并行 → T4+T5 并行 → T6
- T1/T2/T3 完全独立，可并行开发
- T4 依赖三款游戏目录就位（editorial.json 引用 gid、icon.png 已生成）
- T5 依赖三款游戏 src/ 可用

---

### Task 1: 魔塔移植（4.5→4.7.2迁移 + meta.json + adapter）✅

**复杂度**: 高

**Scope:**
- `games/magic_tower/meta.json`（新增）：按 design §2 完整字段清单（id="magic_tower", title="魔塔", category="puzzle", price_model="free", aliases=["魔塔","MDX"], pinyin=["motaxie","mtx"], scene="res://games/magic_tower/module.tscn", runtime="pck"）
- `games/magic_tower/module.tscn`（新增）：根节点 Node + module_adapter.gd
- `games/magic_tower/module_adapter.gd`（新增，extends GameModule，~50行）：按 design §3 修订版（boot注入save_dir、quit_requested装配result、pause/resume转发cur_state状态机、reset_run重开）
- `games/magic_tower/src/main.tscn`（移植自 magic-tower-godot/scenes/main.tscn，4.7.2格式转换）
- `games/magic_tower/src/main.gd`（移植自 magic-tower-godot/scripts/main.gd，按 design §3.1 R-8 修订：纯函数零改 + UI/API允许4.7兼容修复）
  - 具体修改范围：project.godot config/features 从 "4.5" → "4.7"；get_window().size → DisplayServer API；input_action() → cur_state 映射；.import 格式更新
  - 保留：十字盘 + A/B/C/D（玩法必需）
  - 移除：Start/投币/退币/退出 4个街机按钮（design §3.1 R-4）
- assets/ 目录（从 magic-tower-godot/assets/ 拷贝，.import 按 4.7.2 格式重新生成）

**Constraints:**
- 「逻辑零改动」范围（A-2/R-8）：calc_damage()/MON/ITEM 常量表 + MT_SELFTEST 断言全过；UI/API 允许 4.7 兼容修复
- adapter 只约定 boot/quit_requested 接口，不假设 src 内部结构（A-1/R-7）
- pause/resume 按 src cur_state 状态机实现（"PLAYING"/"PAUSED"/"BATTLE"等），adapter 只约定调用时机（A-5/R-11）
- adapter ≤60 行；main.gd 逻辑层行数不变
- 不触碰：tetra_nova 现有代码、Shell core/services

**Acceptance:**
- AC: Registry.reload() 扫描后 lookup("magic_tower") 返回完整 GameMeta（id/title/category/scene 字段正确）
- AC: Launcher.launch("magic_tower") 走通 boot → 游戏画面 → quit_requested 完整生命周期
- AC: MT_SELFTEST=1 headless run 全部断言通过（移植后逻辑零改动验证）
- AC: src/main.gd grep 无硬编码 user:// 绝对路径（save.cfg 写 ctx.save_dir）
- AC: adapter pause/resume 转发正确（Launcher.pause_game() → cur_state="PAUSED"，resume → cur_state="PLAYING"）
- AC: 编辑器零报错（4.7.2 兼容）

---

### Task 2: 2048新建 ✅（2026-08-28 headless 验证：module scene exit 0；tools/_probe_2048.gd 纯函数层 11/11 断言全绿；game_2048.gd=300行、adapter=75行）

**复杂度**: 高

**Scope:**
- `games/game_2048/meta.json`（新增）：按 design §2 完整字段清单（id="game_2048", title="2048", subtitle="数字消除 · 经典益智", category="puzzle", aliases=["2048","数字消除"], pinyin=["erling","ersifba"], price_model="free"）
- `games/game_2048/module.tscn`（新增）：根节点 Node + module_adapter.gd
- `games/game_2048/module_adapter.gd`（新增，extends GameModule，~50行）：tetra模式
- `games/game_2048/src/game_2048.tscn`（新增）：2048游戏场景（CanvasLayer + TileMap/Grid 节点 + HUD 分数标签）
- `games/game_2048/src/game_2048.gd`（新增，≤300行）：核心逻辑
  - 纯函数层（RefCounted/静态方法）：grid_merge(direction: Vector2i) → new_grid + score_delta；is_game_over(grid) → bool
  - 视图层：渲染 grid（TileMap 或 LabelGrid）、分数 HUD、游戏结束遮罩
  - 输入层：键盘方向键 + 触屏滑动（SWIPE_THRESHOLD=30, SWIPE_DEADZONE=15，@export 常量）

**Constraints:**
- 逻辑与视图分离：core rules 纯函数可 headless 单测
- 2048 grid 为 4×4；合并规则：同值合并翻倍、每格至多合并一次
- 输入方案（design Q1=A）：键盘方向键 + 触屏滑动（非 D-pad）
- 视口规格（design Q3=A）：720×1560 portrait，内容区居中 + 上下留白放 HUD
- ThemeTokens 全 token（禁硬编码色值）
- game_2048.gd ≤300 行；adapter ≤60 行
- best_score 强写立即落盘（design §4 R-3：沿用 tetra_nova save.cfg 格式，不强制 JSON）
- 不触碰：tetra_nova 现有代码、Shell core/services

**Acceptance:**
- AC: Registry.reload() lookup("game_2048") 返回完整 GameMeta
- AC: grid_merge 规则正确（同值合并翻倍、每格至多一次、分数累计）
- AC: is_game_over() 在网格无合法移动时返回 true
- AC: 键盘方向键 + 触屏滑动均能触发 grid 移动
- AC: quit_requested 含 score/playtime/achievements/extra 字段，score = 本局最高分
- AC: best_score 写入 ctx.save_dir/save.cfg（adapter 侧比较并写入）
- AC: headless run 10+ 轮随机滑动无崩溃

---

### Task 3: 贪吃蛇新建 ✅（2026-08-28：snake_game.gd=292行、adapter=75行；tools/test_snake.gd 15断言全绿）

**复杂度**: 高

**Scope:**
- `games/snake/meta.json`（新增）：按 design §2 完整字段清单（id="snake", title="贪吃蛇", subtitle="经典街机 · 越吃越长", category="action", aliases=["贪吃蛇","she"], pinyin=["tanshishe","tcs"], price_model="free"）
- `games/snake/module.tscn`（新增）：根节点 Node + module_adapter.gd
- `games/snake/module_adapter.gd`（新增，extends GameModule，~50行）：tetra模式
- `games/snake/src/snake_game.tscn`（新增）：贪吃蛇游戏场景（CanvasLayer + TileMap/Grid 节点 + HUD 分数标签 + D-pad 虚拟按键）
- `games/snake/src/snake_game.gd`（新增，≤300行）：核心逻辑
  - 纯函数层（RefCounted/静态方法）：snake_move(snake_body: Array[Vector2i], direction: Vector2i, food_pos: Vector2i) → new_body + ate_food: bool；is_collision(snake_body) → bool
  - 视图层：渲染蛇身（TileMap/ColorRect）、食物、分数 HUD、死亡遮罩
  - 输入层：键盘方向键 + 触屏虚拟 D-pad（复用魔塔十字盘视觉风格）
  - 速度递增曲线：每吃 N=5 个食物提升移动速度

**Constraints:**
- 逻辑与视图分离：core rules 纯函数可 headless 单测
- 网格制移动（非像素级），方向不可反向（当前向上则按向下忽略）
- 输入方案（design Q1=A）：键盘方向键 + 虚拟 D-pad（非滑动）
- 视口规格（design Q3=A）：720×1560 portrait，内容区居中 + 上下留白放 HUD/D-pad
- ThemeTokens 全 token（禁硬编码色值）
- snake_game.gd ≤300 行；adapter ≤60 行
- best_score 强写立即落盘（design §4 R-3：沿用 tetra_nova save.cfg 格式）
- 不触碰：tetra_nova 现有代码、Shell core/services

**Acceptance:**
- AC: Registry.reload() lookup("snake") 返回完整 GameMeta
- AC: snake_move 规则正确（吃到食物增长1节、撞墙/自撞死亡、方向不可反向）
- AC: 键盘方向键 + D-pad 均能控制蛇移动方向
- AC: 每吃5个食物速度提升（headless 可验证 move_interval 缩短）
- AC: quit_requested 含 score/playtime/achievements/extra 字段，score = 本局长度
- AC: best_score 写入 ctx.save_dir/save.cfg
- AC: headless run 10+ 轮随机方向无崩溃

---

### Task 4: editorial.json扩展 + icon生成 ✅（2026-08-28：banner[] 扩至4条；_gen_icons.gd headless 生成三款 512×512 icon.png）

**复杂度**: 低

**依赖**: Task 1 + Task 2 + Task 3

**Scope:**
- `data/editorial.json`（修改）：按 design §5 R-1 补 banner[] 至4条（tetra_nova/magic_tower/game_2048/snake 均衡覆盖）；featured 标注为预留字段（R-5）；hot_queries 补充新游戏关键词
- `tools/_gen_icons.gd`（新增，≤80行）：headless GDScript Image 脚本生成三款 icon.png（512×512，渐变底 + 主题 glyph，neon 配色走 ThemeTokens 色值）
- `tools/_gen_icons.tscn`（新增）：配套场景定义（根节点 Node + _gen_icons.gd），_ready() 生成后 get_tree().quit()（A-6/R-12）
- 运行 icon 生成脚本落盘三款 icon.png

**Constraints:**
- banner[] 条目含 image_path/target_gid/title 字段（image_path 空时使用 meta.icon 占位）
- icon.png 512×512 PNG 格式，霓虹风格与 tetra icon 一致
- _gen_icons.gd 不依赖外部资源（纯 Image.draw() API）
- editorial.json JSON 格式合法，Registry._load_editorial() 解析零报错
- 不触碰：tetra_nova meta.json、现有 games/ 目录结构

**Acceptance:**
- AC: Registry.reload() 后 _editorial.banner[] 含4条新游戏条目（target_gid 均能 lookup 到）
- AC: 三款 icon.png 文件存在（games/{id}/icon.png，512×512 PNG）
- AC: 首页 Banner 轮播出现新游戏条目（GUI 截图核对）
- AC: _gen_icons.tscn headless run 后三张 icon.png 落盘成功

---

### Task 5: headless测试套件 ✅（2026-08-28：test_magic_tower 12断言 / test_2048 13断言 / test_snake 15断言，全绿 exit 0）

**复杂度**: 中

**依赖**: Task 1 + Task 2 + Task 3

**Scope:**
- `tools/test_magic_tower.tscn`（新增）+ 配套脚本：≥10断言（逻辑规则≥6 + 协议走查≥4 + MT_SELFTEST全量），按 design §7.1 R-9 测试架构（直接 instantiate module.tscn → boot(ctx)）
- `tools/test_2048.tscn`（新增）+ 配套脚本：≥10断言（逻辑规则≥6 + 协议走查≥4）
  - 逻辑层：grid_merge 边界值测试（全同值、无合并、对角线移动）、分数累计正确性、is_game_over 判定
  - 协议层：boot(ctx) → src.SAVE_PATH 注入 / pause/resume 状态切换 / quit_requested result 字段完整 / save.cfg 写入
- `tools/test_snake.tscn`（新增）+ 配套脚本：≥10断言（逻辑规则≥6 + 协议走查≥4）
  - 逻辑层：snake_move 增长/死亡/方向不可反向、is_collision 边界（自撞/撞墙）、速度递增曲线验证
  - 协议层：同上

**Constraints:**
- 测试架构（A-3/R-9）：直接 instantiate `games/{id}/module.tscn`，手动调用 boot(ctx={save_dir:...})；不走 Registry/Launcher 全链路（由 T6 覆盖）
- save_dir 用独立 APPDATA 隔离路径（与 CR-5 测试同模式：`dev/tmp_test/Godot/app_userdata/CR6_{game}/db/`）
- 每款测试脚本 ≤100 行
- 不触碰：现有测试场景

**Acceptance:**
- AC: test_magic_tower headless run ≥10断言全绿 + MT_SELFTEST 全量通过
- AC: test_2048 headless run ≥10断言全绿
- AC: test_snake headless run ≥10断言全绿
- AC: 独立 APPDATA 隔离无交叉污染（三款并行跑互不影响）

---

### Task 6: 全量回归 + GUI双主题验收 ✅（2026-08-28：17场景独立APPDATA全绿；neon/elegant GUI截图通过；Registry.query+Searcher 命中三款新游戏）

**复杂度**: 中

**依赖**: Task 4 + Task 5

**Scope:**
- 全量 headless 回归：既有14场景（test_home/test_db/test_smoke等）+ 新3场景（test_magic_tower/test_2048/test_snake）= **17场景**，独立 APPDATA 全绿
- GUI 双主题验收（neon/elegant）：
  - 首页 Banner 轮播含新游戏条目（无 overflow/布局跳变）
  - SectionContinue 多条目渲染（A-4/R-10：neon/elegant 双主题截图核对，确认无布局跳变/overflow）
  - 三款游戏分别启动 → 游戏画面正常 → 「回菜单」按钮回盒 → 结算卡弹出（score/playtime 正确）
- progress.md 记录执行过程

**Constraints:**
- 独立 APPDATA 隔离（每场景一个子目录，防交叉污染——CR-5 教训）
- GUI 截图用 Godot_v4.7.2-stable_win64.exe --scene res://shell/main.tscn + 视口截图
- 对比度验证（WCAG ≥4.5）：Banner 标题/结算卡文字
- 不触碰：现有代码（仅新增 games/ 目录 + editorial.json + tools/ 测试）

**Acceptance:**
- AC: 17场景 headless 全绿（独立 APPDATA）
- AC: GUI 双主题截图通过（neon/elegant 首页/Banner/SectionContinue/结算卡均正常，无 overflow/布局跳变）
- AC: 三款游戏分别走通 launch → play → quit_to_shell → result_card 完整链路
- AC: Registry.query() 能命中三款新游戏（按 category/tags/aliases/pinyin 搜索）

---

## 修订记录

基于 review_report.md（2026-08-28，13项问题全修复），本 tasks.md 已同步修订：
- T1 Scope 补魔塔 UI移除4个街机按钮 + 回菜单放顶部 HUD 区（R-4/P-1）
- T2/T3 meta.json 字段按 R-2/R-6 补完整清单
- T2/T3 save.cfg 格式改为沿用 tetra_nova（R-3/B-3）
- T5 测试架构明确为直接 instantiate module.tscn（R-9/A-3）
- T6 补 SectionContinue 多条目双主题 GUI 验收（R-10/A-4）