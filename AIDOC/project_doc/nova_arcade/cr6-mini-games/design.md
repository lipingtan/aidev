# 设计：CR-6 M2 小游戏批量接入（魔塔 + 2048 + 贪吃蛇）

> 依据：`requirements.md`（FR1~FR5）+ `design_plan.md`（Q1~Q5 Answer 已确认，D1=A tetra模式）。
> 上游：client-design §4.2/§7 + override §1/§2 + hybrid §3.2。
> 工程实查：GameModule 基类已有 boot()/quit_requested()/pause/resume ✅；Registry.reload() 扫描 games/ ✅；tetra_nova adapter 先例已定 ✅；SectionContinue 走 records_updated 信号重建列表 ✅；editorial.json 当前仅 tetra_nova 1条 banner。

---

## 1. 文件与目录变更

```
nova-arcade/
├── games/
│   ├── magic_tower/          # 新增
│   │   ├── meta.json         # 魔塔元数据（aliases含"魔塔/MDX"等）
│   │   ├── module.tscn       # 根节点 Node + module_adapter.gd
│   │   ├── module_adapter.gd # tetra模式：boot注入save_dir、quit装配result、pause转发
│   │   ├── src/              # 移植自 magic-tower-godot（4.5→4.7.2）
│   │   │   ├── main.tscn     # 原工程 scenes/main.tscn 转换后
│   │   │   └── main.gd       # 原工程 scripts/main.gd 转换后（逻辑零改动）
│   │   └── icon.png          # headless GDScript 生成 512×512 霓虹风格
│   ├── game_2048/            # 新增
│   │   ├── meta.json         # aliases含"2048/erling"等
│   │   ├── module.tscn       # 根节点 Node + module_adapter.gd
│   │   ├── module_adapter.gd # tetra模式
│   │   ├── src/
│   │   │   ├── game_2048.tscn    # 2048游戏场景
│   │   │   └── game_2048.gd     # 核心逻辑（纯函数：grid merge + view渲染）
│   │   └── icon.png          # headless GDScript 生成
│   └── snake/                # 新增
│       ├── meta.json         # aliases含"贪吃蛇/she"等
│       ├── module.tscn       # 根节点 Node + module_adapter.gd
│       ├── module_adapter.gd # tetra模式
│       ├── src/
│       │   ├── snake_game.tscn    # 贪吃蛇游戏场景
│       │   └── snake_game.gd      # 核心逻辑（纯函数：snake move + collision）
│       └── icon.png          # headless GDScript 生成
├── data/
│   └── editorial.json        # 扩展：banner[] 补充三款新游戏条目
├── tools/
│   ├── test_magic_tower.tscn    # 魔塔headless测试（≥10断言）
│   ├── test_2048.tscn           # 2048 headless测试（≥10断言）
│   └── test_snake.tscn          # 贪吃蛇headless测试（≥10断言）
└── core/                     # Shell零改动（仅复用 GameModule 基类）
```

---

## 2. meta.json 规范（三款统一格式）

所有 meta.json 均遵循 tetra_nova 先例，新增字段：

| 字段 | 说明 |
|------|------|
| `id` | 游戏唯一标识（magic_tower / game_2048 / snake） |
| `title` | 中文标题（"魔塔" / "2048" / "贪吃蛇"） |
| `subtitle` | 副标题（玩法简述） |
| `category` | 分类（puzzle / action） |
| `tags` | 标签数组（含英文/中文关键词） |
| `aliases` | 别名数组（含中文名、拼音、简称） |
| `pinyin` | 拼音数组（用于搜索匹配） |
| `scene` | 模块场景路径 `res://games/{id}/module.tscn` |
| `runtime` | `"pck"`（三款均为 GDScript PCK 模式） |
| `achievments` | 成就列表（本 CR 可空数组，M2 AchievementEngine 后续接入） |

---

## 3. Adapter 模式（统一模板）

三款 adapter 均 extends GameModule，遵循 tetra_nova 先例：

```gdscript
extends GameModule
## {GAME_TITLE} 协议适配器（CR-6；D1=A tetra模式 / D5=A「回菜单」为唯一主动退出）
## - boot(ctx)：注入 SAVE_PATH = ctx.save_dir + "save.cfg" → src
## - src game_over(stats) → quit_requested({score, playtime, achievements:[], extra:{...}})
## - pause/resume 转发 src 暂停状态机
## - reset_run()：重开一局（Launcher.play_again 复用模块不重建）
## 依赖：仅 GameModule 基类 + src 节点；不引用 Shell 任何类型

var _src: Node = null
var _run_start_ms: int = 0
var _last_stats: Dictionary = {}
var _menu_btn: Button = null

func _ready() -> void:
    # 子节点 _ready 先于本节点：此处 src 已完成组装
    _src = $Src as Node
    _install_menu_button()
    # 连接 src 的游戏结束信号（各游戏不同）
    _connect_src_signals()

func boot(ctx_in: Dictionary) -> void:
    ctx = ctx_in
    var path := str(ctx.get("save_dir", "user://")) + "save.cfg"
    _src.SAVE_PATH = path
    _run_start_ms = Time.get_ticks_msec()

func quit_to_shell() -> void:
    quit_requested.emit({
        "score": int(_last_stats.get("score", 0)),
        "playtime": _run_seconds(),
        "achievements": [],
        "extra": {},
    })

# pause/resume/reset_run 各按 src 状态机实现
```

### 3.1 魔塔 adapter 特殊处理

Magic Tower demo（Godot 4.5）UI 改造：
- **保留**：十字盘 + A/B/C/D 按钮（玩法必需）
- **移除**：Start/投币/退出 三个街机按钮（盒内无投币语义）
- **追加**：adapter 侧「回菜单」按钮（tetra D5=A 模式，src 最小 diff）

4.5→4.7.2 迁移要点：
- `project.godot` 中 `config/features` 从 `4.5` → `4.7`
- `get_window().size` API 在 4.7 仍可用，但推荐用 `DisplayServer`
- 触屏输入 API（`Input.get_pointer_count()` 等）未变
- 逻辑层零改动原则：`main.gd` 仅做 API 兼容修复，`MT_SELFTEST` 全量通过

### 3.2 2048 adapter 特殊处理

2048 输入方案（Q1=A）：
- **键盘方向键**：上下左右直接映射 grid 移动方向
- **触屏滑动**：自实现手势判定（位移阈值 ≥30px，方向 dominance = |dx| > |dy| 判断水平/垂直）
- 游戏结束判定：网格无合法移动时触发结算入口

### 3.3 贪吃蛇 adapter 特殊处理

贪吃蛇输入方案（Q1=A）：
- **键盘方向键**：上下左右映射蛇移动方向
- **触屏虚拟 D-pad**：右下角十字盘，复用魔塔 demo 的十字盘视觉风格（Phoenix 主题图）
- 速度递增曲线：每吃 N 个食物提升速度（按设计参数）

---

## 4. 存档同构

三款游戏均写 `save.cfg`（JSON）到 `ctx.save_dir`，格式统一：

```json
{
    "best_score": 0,
    "last_played": "",
    "game_state": {}
}
```

- **best_score**：历史最高分（adapter 在 quit_requested 时比较并写入）
- **last_played**：上次游玩时间戳（供继续游戏区块排序）
- **game_state**：游戏特有存档数据（各游戏自行定义结构）

Shell DB 记录 total_playtime/best 供结算卡与继续游戏使用，与 tetra_nova 同构。

---

## 5. editorial.json 扩展

当前 `editorial.json` 仅 tetra_nova 1条 banner。本 CR 补充：

```json
{
    "banner": [
        {"image_path": "", "target_gid": "tetra_nova", "title": "TETRA NOVA"},
        {"image_path": "", "target_gid": "magic_tower", "title": "魔塔 · 经典闯关"},
        {"image_path": "", "target_gid": "game_2048", "title": "2048 · 数字消除"}
    ],
    "featured": [
        {"gid": "tetra_nova"},
        {"gid": "magic_tower"},
        {"gid": "game_2048"},
        {"gid": "snake"}
    ],
    ...
}
```

注意：`image_path` 为空时 Banner/推荐区块使用 meta.icon 作为占位。

---

## 6. icon.png 生成方案（Q4=A）

headless GDScript Image 脚本 `tools/_gen_icons.gd`：
- 512×512 尺寸，渐变底 + 主题 glyph（文字/简单图形）
- neon 配色走 ThemeTokens 色值（与 tetra icon 先例一致）
- 一次跑三张落盘到各自 `games/{id}/icon.png`

```gdscript
# 运行方式：
# Godot_v4.7.2-stable_win64_console.exe --headless --quit-after 30 res://tools/_gen_icons.tscn
```

---

## 7. 测试结构（Q5=A）

### 7.1 每款 headless 测试套件

每款 `tools/test_{game}.tscn`：≥10 断言（逻辑规则 ≥6 + 协议走查 ≥4）

**test_magic_tower.tscn**：
- 逻辑层：calc_damage() 边界值、怪物属性表、道具效果、门/楼梯机制
- 协议层：boot(ctx) → src.SAVE_PATH 注入 / pause/resume 转发 / quit_requested result 字段完整 / save_dir 写入
- MT_SELFTEST 全量（移植后逻辑零改动）

**test_2048.tscn**：
- 逻辑层：grid merge 规则（同值合并翻倍、每格至多一次）、分数累计、游戏结束判定（无合法移动）、滑动方向映射
- 协议层：同上 + score/playtime 上报 / best_score 写入 save.cfg

**test_snake.tscn**：
- 逻辑层：蛇身增长、食物生成、撞墙/自撞死亡、速度递增曲线、D-pad 输入映射
- 协议层：同上 + score/playtime 上报 / best_score 写入 save.cfg

### 7.2 全量回归

既有 14 场景（test_home/test_db/test_smoke 等）+ 新 3 场景 = **17 场景**，独立 APPDATA 全绿。

---

## 8. Registry.reload() + SectionContinue 兼容性

Registry.reload() 扫描 games/ 后自动注册三款新游戏，SectionContinue 走 records_updated 信号重建列表。

**风险点 R3 应对**：多游戏后「继续游戏」区块出现多条目的渲染布局需 GUI 截图核对（参照 CR-5 视觉验收方式）。

---

## 9. 任务依赖关系

```
T1(魔塔移植: 4.5→4.7.2迁移 + meta.json + adapter)
T2(2048新建: game_2048.tscn/gd + meta.json + adapter)
T3(贪吃蛇新建: snake_game.tscn/gd + meta.json + adapter)
T4(editorial.json扩展 + icon生成, 依赖T1+T2+T3)
T5(headless测试套件: test_magic_tower/test_2048/test_snake, 依赖T1+T2+T3)
T6(全量回归: 17场景独立APPDATA + GUI验收, 依赖T4+T5)
```

- **关键路径**：T1/T2/T3 并行 → T4+T5 并行 → T6
- T1/T2/T3 完全独立，可并行开发
- T4 依赖三款游戏目录就位（editorial.json 引用 gid、icon.png 已生成）
- T5 依赖三款游戏 src/ 可用

---

## 10. 风险点应对

| 风险 | 应对 |
|------|------|
| R1: Godot 4.5→4.7.2 API 变更 | 按 kb/godot-4.7-api-facts.md 逐项修复；逻辑层零改动；MT_SELFTEST 全量验证 |
| R2: 触屏滑动阈值不当 | 参数走 @export 常量（SWIPE_THRESHOLD=30, SWIPE_DEADZONE=15）；GUI 验收实测调整 |
| R3: SectionContinue 多条目布局跳变 | T6 GUI 双主题截图核对（neon/elegant），与 CR-5 视觉验收同口径 |

---

## 11. 修订记录（Review 后更新）

基于 review_report.md（2026-08-28，13项问题全修复）：

| 修订号 | 来源 | 修改内容 |
|--------|------|----------|
| R-1 | B-1 | §5 banner 补 snake 条目至4条均衡覆盖 |
| R-2 | B-2 | §2 meta.json 补完整字段清单（price_model/free, price=0, trial={}, version="1.0.0", screenshots=[], desc=玩法简述） |
| R-3 | B-3 | §4 save.cfg 改为沿用 tetra_nova 格式（不强制 JSON 解析） |
| R-4 | P-1 | §3.1 魔塔 UI：移除 Start/投币/退币/退出 4个按钮；回菜单放顶部 HUD 区（非右侧） |
| R-5 | P-2 | §5 featured 标注为预留字段（M2 Recommender 扩展时启用，本 CR 主要靠 banner[] 驱动首页） |
| R-6 | P-3 | §2 aliases/pinyin 统一规范：aliases=中文名+常用简称，pinyin=无声调全拼 |
| R-7 | A-1(高) | §3 adapter 统一模板改为「只约定 boot/quit_requested 接口，不假设 src 内部结构」 |
| R-8 | A-2(高) | §3.1 魔塔「逻辑零改动」明确范围：纯函数（calc_damage/MON/ITEM + MT_SELFTEST）零改；UI/API 允许 4.7 兼容修复 |
| R-9 | A-3 | §7 测试架构：直接 instantiate module.tscn → boot(ctx)，不走全链路（由 T6 覆盖） |
| R-10 | A-4 | §8 + T6 补 SectionContinue 多条目 GUI 验收（neon/elegant 双主题截图核对） |
| R-11 | A-5 | §3 pause/resume 转发按 src 各自实现，adapter 只约定对外接口调用时机 |
| R-12 | A-6 | §6 补 _gen_icons.tscn 配套场景；脚本生成后 get_tree().quit() |