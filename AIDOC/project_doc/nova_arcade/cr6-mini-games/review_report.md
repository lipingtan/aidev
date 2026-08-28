# CR-6 三角色 Review 报告（2026-08-28）

> 对象：requirements.md / design_plan.md / design.md；对照工程实查（registry.gd/launcher.gd/game_module.gd/tetra_adapter/editorial.json/db.gd）。
> 结果：13 项问题（高:2 中:5 低:6），用户确认后修复落盘。

---

## 业务专家

| # | 严重度 | 问题 | 处置 |
|---|--------|------|------|
| B-1 | 中 | design §5 editorial.json banner 仅3条（tetra_nova/magic_tower/game_2048），snake 只进 featured 不进 banner——四款新游戏齐接入，banner 应均衡覆盖 | ✅ design §5 banner 补 snake 条目，共4条；featured 保持4款 |
| B-2 | 低 | design §2 meta.json 字段清单缺 price_model/price/trial/version/screenshots/desc 等 GameMeta 必填字段——Registry.from_dict() 会赋空值，运行时可能导致 Launcher 权益检查异常（trial 字段空 dict vs tetra_nova 有 `{plays:3}`） | ✅ design §2 补完整字段清单；三款均为 free/price=0/trial={}/version="1.0.0"/screenshots=[]/desc=玩法简述 |
| B-3 | 低 | design §4 save.cfg 写 "JSON"，但 tetra_nova adapter 实际用 `FileAccess.store_string()` 存自定义文本格式（非标准 JSON）——三款应沿用同格式而非另起 JSON | ✅ design §4 改为「沿用 tetra_nova save.cfg 格式（adapter 侧读写，不强制 JSON 解析）」 |

---

## 产品经理

| # | 严重度 | 问题 | 处置 |
|---|--------|------|------|
| P-1 | 中 | design §2.1 魔塔 adapter 说「移除 Start/投币/退出三个街机按钮」，但实查 magic-tower-godot/scripts/main.gd 的 `_build_buttons()` 建的是 A/B/C/D + Start/投币/退币/退出共8个——adapter 追加「回菜单」后总按钮数可能超出触屏可视区（720×1560 竖屏下右侧按钮列过密） | ✅ T1 Scope 增加「魔塔 UI 适配：移除 Start/投币/退币/退出4个街机按钮，仅保留十字盘+A/B/C/D」；adapter 追加回菜单按钮放顶部 HUD 区（非右侧） |
| P-2 | 低 | design §5 featured 数组用 `{"gid": "..."}` 格式，但实查 editorial.json 当前无 featured 字段——Registry._load_editorial() 加载后谁消费 featured？Recommender.for_you() 走 Registry.all() 扣除 continue_row gid，不读 editorial.json.featured | ✅ design §5 补 featured 消费方说明（当前为预留字段，M2 Recommender 扩展时启用；本 CR 主要靠 banner[] 驱动首页展示） |
| P-3 | 低 | 贪吃蛇 meta aliases 含 "she"（拼音首字母），但 design_plan Q1=A 说 aliases 应含中文名/拼音——"she" 与 "2048" 的 "erling" 语义不统一（erling = 二零四八？还是其他？） | ✅ 统一 aliases 规范：含中文名 + 常用简称；pinyin 字段承载无声调全拼。2048 aliases=["2048","数字消除"] pinyin=["erling","ersifba"]；snake aliases=["贪吃蛇","she"] pinyin=["tanshishe","tcs"] |

---

## 架构师

| # | 严重度 | 问题 | 处置 |
|---|--------|------|------|
| A-1 | **高** | design §3 Adapter 统一模板用 `_src = $Src as Node`，但实查 tetra_nova adapter 的 src 结构是 `$Main`（内含 `.game`/`.ui` 子节点），magic_tower demo 的 src 是 `$Main`（Node2D，含 player/cam/hud/dpad/buttons 子节点）——两者结构完全不同，统一模板 `_src = $Src` 无法适配两种不同的 src 内部结构 | ✅ design §3 改为「adapter 不假设 src 内部结构：每款 adapter 的 `_connect_src_signals()` / `pause/resume/reset_run` 按各自 src 实际节点实现；统一模板只约定 boot(ctx)→注入SAVE_PATH + quit_requested()→装配result」 |
| A-2 | **高** | design §3.1 魔塔移植说「逻辑层零改动」，但实查 magic-tower-godot project.godot `config/features=PackedStringArray("4.5")` → Godot 4.5 工程导入 4.7.2 时：(a) `get_window().size` 在 4.7 仍可用但被 `DisplayServer.window_get_size()` 取代；(b) `input_action()` 非标准 API，实查 main.gd 用 `cur_state` 状态机而非 input_map；(c) assets/ 目录含 Sprite2D 资源，.import 文件在 4.7.2 格式可能变更——「逻辑零改动」需明确定义范围（仅 main.gd 的纯函数部分零改，UI/输入层允许适配） | ✅ design §3.1 明确：「逻辑零改动」指 calc_damage()/MON/ITEM 常量表 + MT_SELFTEST 断言全过；main.gd 的 UI/输入/API 调用允许做 4.7 兼容修复（含 get_window→DisplayServer、input_action→cur_state 映射、.import 格式更新） |
| A-3 | 中 | design §7.1 test_2048.tscn 说「协议层：boot(ctx)→src.SAVE_PATH 注入」，但实查 Registry._load_meta() 用 `GameMeta.from_dict(parsed)` 而非直接实例化场景——headless 测试如何模拟 Launcher.launch(gid)→boot(ctx) 完整链路？需确认测试架构（直接 instantiate module.tscn → call boot() vs 走 Registry.lookup()+Launcher.launch()） | ✅ design §7 明确测试架构：「每款 test_{game}.tscn 直接 instantiate `games/{id}/module.tscn`，手动调用 boot(ctx={save_dir:...})，验证 src.SAVE_PATH 注入 + quit_requested 信号发射 + save.cfg 写入；不走 Registry/Launcher 全链路（该链路由 T6 全量回归覆盖）」 |
| A-4 | 中 | design §8 说 SectionContinue 走 records_updated 信号重建列表，但实查 SectionContinue._on_records_updated() 接收 gid 参数却不作过滤——当三款新游戏同时触发 upsert_record（比如 T6 批量写入），信号会触发3次 render()，每次清空 _hbox 重建全部卡片——性能可接受（卡片数≤5）但未评估多条目下的布局跳变视觉 | ✅ design §8 + T6 Scope 补「SectionContinue 多条目 GUI 验收：neon/elegant 双主题截图核对，确认无布局跳变/overflow」 |
| A-5 | 中 | design §3 adapter 说 `pause/resume 转发 src 暂停状态机`，但实查 tetra_nova adapter 用 `_game.input_action("pause")`（非标准 API），magic_tower demo 用 `cur_state` 状态机（"PLAYING"/"PAUSED"）——2048/贪吃蛇的暂停机制尚未设计，adapter pause/resume 转发逻辑可能因 src 不同而各异 | ✅ design §3 补「每款游戏的暂停机制由 src 自行实现（2048=暂停遮罩+状态切换；贪吃蛇=蛇身冻结+暂停提示），adapter 只约定对外接口 `pause_game()/resume_game()` 的调用时机，不强制内部实现方式」 |
| A-6 | 低 | design §6 icon 生成脚本 `_gen_icons.gd` 用 headless GDScript Image，但实查 Godot 4.7 headless 模式加载场景需 `--headless --quit-after N res://tools/_gen_icons.tscn`——需确认 .tscn 配套文件（根节点 Node + _gen_icons.gd）存在 | ✅ design §6 补 `_gen_icons.tscn` 配套场景定义；脚本在 `_ready()` 生成三张 icon.png 后 `get_tree().quit()` |

---

## 结论

✅ 三角色 review 完成，13/13 已修复落盘（design.md / requirements.md 同步修订）。Q1~Q5/D1=A 决策未受影响。

**高优先级修复 2 项**：
- A-1: adapter 统一模板改为「只约定 boot/quit_requested 接口，不假设 src 内部结构」
- A-2: 魔塔「逻辑零改动」明确范围（纯函数零改，UI/API 允许兼容修复）

**中优先级修复 5 项**：
- B-1: banner 补 snake 条目至4条均衡覆盖
- P-1: 魔塔 UI 移除4个街机按钮，回菜单放顶部 HUD 区
- A-3: 测试架构明确（直接 instantiate module.tscn，不走全链路）
- A-4: SectionContinue 多条目 GUI 验收补双主题截图核对
- A-5: pause/resume 转发逻辑按 src 各自实现

**低优先级修复 6 项**：均已落盘。