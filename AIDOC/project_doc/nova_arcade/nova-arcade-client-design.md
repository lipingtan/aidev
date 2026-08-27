# NOVA ARCADE 客户端详细设计（模块 + 交互）

> 上游文档：`nova-arcade-design.md`（总览）。本文聚焦客户端模块划分、接口协议与全部交互流。
> 技术底座：Godot 4.5 + GDScript，竖屏画布，本地优先（全部功能离线可用）。
>
> **Shell 基准分辨率：`720 × 1560`**，`stretch/mode=canvas_items`，`stretch/aspect=expand`。
> DLC 游戏通过 `meta.json` 的 `viewport_size` / `orientation` 字段声明自己的设计基准，
> Launcher 在转场遮罩盖住屏幕后切换，退出时自动恢复 Shell 值（详见 §4.6）。
> 省略两字段时默认继承 Shell 值，零迁移成本。

---

## 1. 总体架构

### 1.1 场景树（运行时形态）

```
Main.tscn (Node)
├── EventBus / QualitySettings / DB / Registry / Nav / Sound  ← Autoload 单例（不在此树）
├── BgLayer (Control)          # 星空/网格动态背景，全局常驻
├── App (Control)
│   ├── PageStack (Control)    # 页面栈容器（滑入/滑出动画）
│   │   └── <当前页面链>
│   ├── TabBar (Control)       # 底部四 Tab：首页/分类/搜索/我的
│   ├── OverlayLayer (Control) # 结果结算卡/购买弹窗/下载弹窗/评价编辑器
│   └── ToastLayer (Control)   # 全局提示（顶部滑入）
├── GameHost (Control)         # 游戏全屏运行容器（玩时隐藏 App）
└── Services (Node)            # Launcher/Pay/Searcher/Recommender/Analytics...
```

### 1.2 分层原则

```
┌─ 页面层 Pages（只管展示与转发，不写业务规则）
├─ 服务层 Services（业务规则：试玩判定/推荐排序/订单状态机）
├─ 单例层 Autoloads（跨模块共享：事件/存储/目录/导航/音效）
└─ 数据层 user://（唯一持久化出口，全部经 DB 单例读写）
```

- 页面之间**禁止互相引用**，一切跳转走 `Nav`，一切通知走 `EventBus`。
- 游戏模块（GameModule）**只依赖 ctx 注入**，不感知 Shell 存在。

---

## 2. 模块清单

### 2.1 单例层（project.godot 注册 Autoload）

| 模块 | 职责 | 关键接口 |
|---|---|---|
| **EventBus** | 全局信号枢纽，模块解耦 | 见 §3.3 信号表 |
| **QualitySettings** | 设备档位检测（`--aidev-tier=` CLI 覆盖 + 平台判定）与资源分档解析；须先于数据类单例注册 | `detect_tier()`<br>`resolve_texture(path)`<br>`get_scalar(key)` |
| **DB** | 全部本地持久化的唯一出口；JSON/ConfigFile 封装，写入分两级：**强写**（立即落盘，用于关键数据）/ **防抖写**（500ms 合并落盘，用于非关键数据）。强写范围：`trial_used`（消耗次数）、`orders`（订单状态）、`total_playtime/best/finish_count`（game_finished 时）；防抖写范围：`search_history`、`daily`、`profile`（昵称/设置）。崩溃恢复可靠性由强写保证。 | `get_profile() / save_profile()`<br>`get_record(gid) / upsert_record(gid, patch)`<br>`list_records()`<br>`get_review(gid) / put_review(gid, dict)`<br>`list_orders() / put_order(dict)`<br>`add_search_history(q) / get_search_history()`<br>`get_achievements() / unlock(id)`<br>`get_daily() / touch_daily()` |
| **Registry** | 游戏目录：内置 meta + DLC meta；查询/筛选/排序/版本比较；成就定义表 | `reload()`<br>`lookup(gid) -> GameMeta?`（偏差：原设计 `get(gid)` 与 Node 基类 Object.get 签名冲突，实现改名）<br>`all() -> [GameMeta]`<br>`query({category, price, min_rating, tag, sort}) -> [GameMeta]`<br>`installed_version(gid)`<br>`ach_def(aid)` |
| **Nav** | 页面栈管理 + Tab 切换 + 转场动画 + 返回键拦截 | `push(page_path, data)`<br>`pop()` / `pop_to_root()`<br>`switch_tab(idx)`（清栈回 Tab 根页）<br>`current() -> String` |
| **Sound** | UI 音效/震动（尊重设置开关）；预合成 PCM，延续 TETRA NOVA 方案 | `click() / toggle() / success() / error() / coin()`<br>`haptic(str)` |

### 2.2 服务层（Shell 下 Services 节点，非单例）

| 模块 | 职责 | 关键接口 |
|---|---|---|
| **Launcher** | 游戏生命周期总管（详见 §4.6 时序） | `launch(gid)`<br>`_on_game_quit_requested(result)` |
| **TrialGuard** | 试玩资格判定与计数 | `left(gid) -> int`<br>`consume(gid) -> int`（返回剩余） |
| **PayService** | 订单状态机：创建→支付→票据→权益；现阶段 Mock，后期接渠道 SDK | `price(gid)`<br>`create_order(gid) -> order`<br>`finalize(order)`<br>`restore()`<br>`owned(gid) -> bool`（free/trial 已购/iap 恒 true） |
| **Searcher** | 搜索索引构建与打分查询（见 §4.4） | `build_index()`<br>`query(text) -> [gid]`（前缀>包含>别名>拼音>标签）<br>`hot_queries() -> [String]` |
| **Recommender** | 首页各区块数据源（见 §7） | `continue_row() -> [gid]`<br>`for_you() -> [gid]`<br>`charts(mode) -> [gid]`<br>`similar(gid, n=3) -> [gid]` |
| **AchievementEngine** | 盒子级成就判定（监听 EventBus + PlayRecord） | `evaluate(event)` |
| **Analytics** | 埋点：内存 buffer + JSONL 追加 user://logs/；后期批量上报 | `track(event, props)` |
| **DLCManager**（后期） | 下载/校验/挂载 PCK | `download(gid)` `mount(gid)` |
| **CoreManager**（Arcade） | Arcade 核心 .so 与 ROM 状态管理；Android 委托 MameRuntime 插件（v2），桌面返回「不可用」（arcade 仅真机可玩，CTA 置灰）。现有 Mock 桩待 M2 spike 后替换 | `ensure_native()`<br>`is_ready() -> bool`<br>`launch(rom, extras)`<br>`download_progress() -> float`<br>signal `game_exited` |

### 2.3 页面层

| 页面 | 场景路径 | 入口 | 说明 |
|---|---|---|---|
| HomePage | `pages/home.tscn` | Tab0 | 推荐流五区块 |
| CategoryPage | `pages/category.tscn` | Tab1 | 分类+筛选+网格 |
| SearchPage | `pages/search.tscn` | Tab2 | 搜索全流程 |
| LibraryPage | `pages/library.tscn` | Tab3 | 我的/游戏库 |
| DetailPage | `pages/detail.tscn` | 卡片点击 | 详情+CTA |
| ReviewsPage | `pages/reviews.tscn` | 详情"全部评价" | 评价中心 |
| AchievementsPage | `pages/achievements.tscn` | 我的→成就墙 | 成就总览 |
| WalletOrdersPage | `pages/orders.tscn` | 我的→钱包订单 | 订单列表/恢复购买 |
| SettingsPage | `pages/settings.tscn` | 我的→设置 | 全局设置 |
| ResultOverlay | `overlays/result.tscn` | 游戏退出时 | 结算卡（模态） |
| PurchaseDialog | `overlays/purchase.tscn` | 购买 CTA | 支付流程 |
| ReviewEditor | `overlays/review_editor.tscn` | 评价入口 | 写/改评价 |
| DownloadDialog | `overlays/download.tscn` | DLC 下载 | 进度+校验 |

### 2.4 游戏运行时层

`GameHost`（Main 下全屏容器）+ `GameModule` 协议（见 §3.1）。玩游戏时 `App.visible=false`，退出恢复。

三种运行时实现统一 `Runner` 接口（PckRunner / HtmlRunner / ArcadeRunner），由 `meta.json.runtime` 分发——详见 `nova-arcade-runtime-design.md`。ArcadeRunner = MAME4droid native Activity + MameRuntime 插件桥（v2，参照 mjarch3 内容管线，不用 libretro）：首次需下载核心 .so（~74MB/ABI）+ 该游戏的 rom zip 集与说明文件（专属文件夹），退出协议/结算卡与其他运行时完全同构。

> **状态（v2 定稿）**：Arcade 运行时以 `projects/nova_arcade/mame-godot-plugin/DESIGN.md` v2 为准（标准 Godot 导出 + 独立 native Activity + MameRuntime 桥接）。本文 ArcadeRunner 相关小节（§2.2 CoreManager、§4.6 Arcade 分支、§5 核心目录）已按 v2 对齐；行为细节（帧率/音频/进程语义/插件打包）待 M2 真机 spike 验证后定稿。

---

## 3. 核心接口协议

### 3.1 GameModule（每个游戏实现）

```gdscript
class_name GameModule extends Node

# ---- 由游戏提供 ----
signal quit_requested(result: Dictionary)   # 游戏主动退出时发
signal achievement_unlocked(id: String)

func boot(ctx: Dictionary) -> void:
    # ctx = {
    #   save_dir: "user://saves/tetra_nova/",  # 游戏私有存档区
    #   trial_mode: bool,                       # 本次是否试玩
    #   owned: bool,
    #   best: int,                              # 历史最好成绩
    # }
    pass

func pause_game() -> void: pass    # 系统通知/弹窗时由 Shell 调用
func resume_game() -> void: pass

# result = {
#   score: int, playtime: float,
#   achievements: [String],        # 本次解锁的成就 id
#   extra: Dictionary              # 游戏自定义（如 wave/boss 数）
# }
```

**游戏接入清单**（TETRA NOVA 已满足）：实现 meta.json + boot/pause/resume + 主动退出上报 + 存档只写自己的 save_dir。

### 3.2 Page 基类

```gdscript
class_name Page extends Control
func on_enter(data: Dictionary) -> void    # 入栈，接收参数（如 {game_id}）
func on_exit() -> void                     # 出栈，释放资源
func on_resume() -> void                   # 上层页面弹出、重新可见
func on_back() -> bool                     # 返回键；return false 走默认 pop
```

### 3.3 EventBus 信号契约

| 信号 | 参数 | 触发者 → 监听者 |
|---|---|---|
| `game_launched` | gid | Launcher → Analytics/成就 |
| `game_finished` | gid, result | Launcher → ResultOverlay/DB/成就/Analytics |
| `trial_consumed` | gid, left | TrialGuard → DetailPage(刷新CTA) |
| `payment_succeeded` | gid, order_id | PayService → DetailPage/Toast/成就 |
| `review_submitted` | gid | ReviewEditor → ReviewsPage/成就 |
| `review_deleted` | gid | ReviewEditor → ReviewsPage |
| `achievement_unlocked` | def, points | AchievementEngine → Toast/成就页 |
| `dlc_installed` | gid | DLCManager → Registry.reload/刷新UI |
| `tab_changed` | idx | Nav → TabBar/Analytics |
| `settings_changed` | key, value | SettingsPage → Sound（非主题类设置） |
| `theme_changed` | theme_name | ThemeTokens.apply_theme → BgLayer/页面重渲染 |
| `records_updated` | gid | DB → 首页"继续游戏" |

### 3.4 数据结构（Registry 里的标准 GameMeta）

```yaml
id: tetra_nova
title: "TETRA NOVA"
subtitle: "新星方阵 · 变异消除"
category: puzzle          # puzzle/action/arcade/casual/roguelike
tags: [tetris, roguelike, neon, offline]
price_model: trial        # free|ad|trial|paid|iap
price: 6                  # 元；free/ad/iap 为 0
trial: {plays: 3}         # 或 {minutes: 10}
version: "1.0.0"
size_kb: 4200             # DLC 用；内置游戏为 0
icon: "res://games/tetra_nova/icon.png"
screenshots: [...]
scene: "res://games/tetra_nova/main.tscn"
dlc_url: ""               # DLC 阶段填 CDN 地址
# 视口与方向（可选；省略则继承 Shell 默认值 720×1560 竖屏）：
viewport_size: [720, 1560]  # 游戏的逻辑设计分辨率 [w, h]；DLC 开发者按自己的设计基准填
orientation: "portrait"     # "portrait" | "landscape"
                            # 横屏游戏填 [1920, 1080] + "landscape"
                            # Launcher 在转场遮罩完全盖住屏幕后切换，退出时自动恢复 Shell 值
aliases: ["俄罗斯方块", "els", "tetris"]
pinyin: ["eluosifangkuai", "elsfk"]
# aliases / pinyin 由内容编辑在编辑阶段预填写，编译进 meta.json 随资源包发布。
# aliases: 中英文别名数组（大小写不敏感，Searcher 内部统一 lowercase）
# pinyin:  无声调全拼数组，每个元素对应 aliases[i] 或 title 的全拼；
#          首字母缩写 Searcher 运行时自动从全拼提取（取每段首字母），无需单独字段。
#          例：title="新星方阵" → pinyin 增加 "xinxingfangzhen" → 缩写 Searcher 提取 "xxfz"
achievements:             # 该游戏成就定义
  - {id: tn_wave10, name: "突破十波", points: 10}
  - {id: tn_boss3, name: "屠龙者", points: 20}
desc: "..."
# Arcade 特有字段（runtime: "arcade" 时必填）：
runtime: "pck"            # pck|html|arcade
core: ""                  # 运行时依赖的核心包名（arcade="mame4droid-0.288"，其余为空）
roms: []                  # [{
  # 每个 ROM 对应一个 MAME 驱动（romset）
  name: "pacman",         # MAME 驱动名（romset 名），对应 argv[0]
  title: "PAC-MAN",
  parent: "",             # 父驱动名（空=无父；clone 时填 parent 名；对应 gamedrv.h::parent / myosd_game_info.parent）
  bios: "",               # 依赖的 BIOS 驱动名（空=无 BIOS 依赖）
  zips: ["pacman.zip"],   # 该驱动需要的 rom zip 文件列表（每个 zip 对应 rompath 下的一个文件）
  controllerType: "virtualkey",   # "virtualkey" | "joystick"
  buttonNumber: 4,
  cansl: true,            # 是否允许存状态
  buttonConfig: {...},
  hiscore: "plugin"       # "plugin" | "nvram" | "none"
# }]
```

---

## 4. 页面与交互流

### 4.0 导航图

```
TabBar ── Home ────┬── DetailPage ─┬── ReviewsPage ── ReviewEditor(模态)
                   │               ├── PurchaseDialog(模态)
                   │               └── DownloadDialog(模态, DLC)
       ── Category ─┘(共享 DetailPage)
       ── Search ────┘
       ── Library ─┬── AchievementsPage
                   ├── WalletOrdersPage
                   └── SettingsPage
ResultOverlay：游戏退出后模态覆盖全屏，关闭后回到原 Tab 根页
```

转场规范：
- 页面推入：右滑入 **220ms**（opacity 0→1 + translateX 24px→0），`cubic-bezier(.25,.46,.45,.94)`
- 页面弹出：左滑出 220ms（同曲线反向）
- Tab 切换：交叉淡入 150ms
- 进入游戏：App `scale(.97)` + fade-out 300ms；游戏退出：App fade-in 300ms
- 结算卡弹出：`translateY(40px) scale(.92) opacity:0` → normal，280ms，**`cubic-bezier(.34,1.5,.64,1)`**（弹簧过冲）
- Toast 弹出：`translateY(-70px)` → `translateY(0)`，300ms，`cubic-bezier(.34,1.4,.64,1)`；显示 **2200ms** 后收起
- Tab 选中指示器：底部 **16×2.5px** 圆角胶囊，颜色 `--accent`（不是下划线，不是点）
- 底部 Tab 图标：选中 `filter:none; opacity:1`；未选中霓虹 `grayscale(.7) brightness(.85)`，未选中青瓷 `grayscale(.35); opacity:.75`

### 4.1 冷启动

```
Splash(0.8s 品牌动画)
→ Main._ready():
   1. DB.load_all()          # profile/records/reviews/orders/achievements
   2. Registry.reload()      # 扫描 games/*/meta.json + user://dlcs
   3. PayService.retry_pending_orders():
        - 有 pending 订单且网络可用 → 后台静默重试验签
        - 重试成功 → 顶部 Toast "上次购买已恢复 ▸"
        - 无网络 / 重试失败 → 顶部横幅"有一笔订单待确认 [重试]"（带关闭按钮，不阻塞启动）
        - 多次失败的订单仍保留 pending 状态，下次启动继续重试
   4. 构建 Shell，Nav.switch_tab(0)
   5. Analytics.track(app_open)
```

### 4.2 首页（五区块，下拉刷新）

| 区块 | 数据源 | 交互 |
|---|---|---|
| Banner 轮播(3张) | 编辑配置（本地 json，后期云端下发） | **4s** 自动轮播；激活指示点宽度 14px（非激活 5px），CSS transition 300ms；点击→DetailPage/活动页 |
| ▶ 继续游戏 | `Recommender.continue_row()` = PlayRecord.last_played 倒序 top5 | 横滑卡片；**长按卡片→弹"从记录移除"确认气泡**；移除后立即 re-render 本区块 |
| ▶ 为你推荐 | `for_you()`（§7） | 横滑；卡片点击→DetailPage；游戏结束后 `renderHome()` 立即重渲染（更新"继续游戏"区块） |
| ▶ 热门榜 | `charts(mode)`，子Tab：综合/新游/好评 | 列表行：排名+图标+名+★+玩过人数；点击→Detail |
| ▶ 每日任务 | DB.get_daily()（本地日期判断重置） | 完成打勾；全部完成→成就点+Toast |

卡片（三规格）统一信息：图标（动态渐变背景，见 §8.1 token）/名称/★评分/标签角标(付费·离线·新)/状态角标(已装▶、未装⬇、试玩N、已拥有)。

### 4.3 分类页

- 左侧竖向分类栏：全部/消除/益智/动作/街机/休闲/Roguelike（数量角标）；选中项左侧显示 3px 强调色竖条（`::before`，高度 26%–74%）。
- 顶部筛选 chips：**"全部"互斥**（选中时清空所有其他过滤器）；其余**多选叠加**（`免费` `付费` `离线` `评分≥4.0` `新游`）；选任意非"全部"时"全部"自动取消选中。
- 右侧 2 列网格；顶部排序下拉：`最热`（默认）/ `最新` / `评分`（M2 实现，M1 只用最热）。
- 数据一次拉全（本地目录小），内存过滤即时刷新；无结果时显示空态 + 内联"清除筛选"按钮。

### 4.4 搜索页

```
进入 → 自动聚焦输入框 → 展示:
  热搜（M1 用本地 editorial.json 预配置 6 条；M2+ 改为本地词频统计 top-8）
  历史（top-20，可清空；与 DB HISTORY_CAP=20 一致，review P-3）
输入(去抖 150ms) → Searcher.query:
    分词 → 每游戏打分:
      标题前缀命中 +100 │ 标题包含 +40 │ 别名包含 +35
      拼音全拼前缀 +30 │ 拼音首字母前缀 +25 │ 标签命中 +15
    总分排序取 20 → 实时列表(图标/名/★/标签/状态角标)
无结果 → "没有找到「xx」" + 猜你喜欢(Recommender.for_you 前8)
点击结果 → DetailPage，并 add_search_history(q)、词频+1
```

> **双语标题（t2）的搜索处理**：游戏在青瓷主题下显示的别名（如"2048 青瓷版"、"竹影贪吃蛇"）在内容编辑时写入 `aliases[]` 字段，Searcher 不需要感知 `t2` 的存在——直接通过别名命中即可（得分 +35）。这样 Searcher 与主题系统完全解耦。

### 4.5 详情页（CTA 状态机 = 客户端最核心交互）

```
┌ 截图横滑轮播(指示点)
├ 图标+名称+版本 │ ★4.8(2.3万) │ 12万人玩过
├ 标签行: 消除 · Roguelike · 离线
├ 概况: 简介 / 更新日志(折叠，默认收起)
├ 评分分布条(5→1) + 好评率 + 近30天标记
├ 精选评价 2条 → [全部评价 →]
├ 相关推荐(Recommender.similar)
└ 底部固定 CTA（状态机见下）
     ┌──────────────────────┐
     │  [副文案 hint 行]     │ ← 始终可见的单行说明文字
     │  [     CTA 按钮    ]  │
     └──────────────────────┘
```

**CTA 状态机**（每次 `records_updated / payment_succeeded / trial_consumed / dlc_installed` 后重算）：

```
状态             CTA 按钮文字              hint 副文案
─────────────────────────────────────────────────────────
DLC 未安装     → [⬇ 下载 · X MB]          "DLC 下载后即可开玩（Wi-Fi 优先）"
free/ad/iap   → [▶ 开玩]                  首次: "首次开玩" / 已玩: "上次最高 X · 继续"
trial 剩N>0   → [▶ 试玩 · 剩 N 次]        "试玩结束后 ¥X 解锁全部"
trial 剩0     → [¥X 解锁完整版]           "试玩次数已用完"
paid 未拥有   → [¥X 购买]                 "买断制 · 一次购买永久拥有"
paid 已拥有   → [▶ 开玩]                  "上次最高 X · 继续" / "首次开玩"
arcade        → 见下方 Arcade 额外逻辑
```

hint 行在 CTA 固定区域内，按钮上方，始终可见，字号 10px，颜色 `--ink3`；按钮按下时 `scale(.97)`，卡片 S 按下 `scale(.95)`，卡片 M 按下 `scale(.96)`。

**Arcade 游戏详情页额外逻辑：**
- 若 `meta.roms[]` 非空（多 ROM/克隆体/子游戏）：CTA 区域改为子游戏选择面板，列出每个 ROM 的 title + controllerType；用户选择后 CTA 显示 `[▶ 开玩 <rom_title>]`
- 每个子游戏有独立的 `cansl`（存状态权限）、`buttonConfig`（按钮几何）、`hiscore` 规则
- 子游戏选择面板可滚动；当前选择高亮；每行显示标题 + 简短说明
- 若为单 ROM（roms 长度=1），直接显示标准 CTA 不展示子面板
- **Parent/Clone 标注**：若 `roms[].parent` 非空，子条目标题旁标注 "(clone of <parent.title>)"；若 `roms[].bios` 非空，标注 "(needs <bios_title>)"
- **BIOS 依赖检查**：每个子条目启动前检查 `roms[].zips[]` 中每个文件是否存在（含 parent zip 和 bios zip）；缺少的文件在详情页显示红色提示 "缺少 <file_name>，[下载]"

### 4.6 启动游戏全流程（时序）

```
用户点 ▶ (任意卡片/详情CTA)
1. Launcher.launch(gid)
2.   meta = Registry.lookup(gid)
3.   DLC未装? → DownloadDialog → 失败中止，成功继续
4.   权益检查 PayGate:
       free/ad/iap/owned → 通过
       trial → TrialGuard.consume(gid) → EventBus.trial_consumed
                剩0 时弹 PurchaseDialog(文案"试玩次数已用完") → 中止
       paid未拥有 → 弹 PurchaseDialog → 中止
5.   DB.upsert_record(gid, {last_played=now, sessions+1})
6.   转场: App 淡出（300ms fade） → 遮罩完全盖住屏幕后：
       a. 读取 meta.viewport_size（默认 [720,1560]）和 meta.orientation（默认 "portrait"）
       b. get_tree().root.content_scale_size = Vector2i(vp[0], vp[1])
       c. 若 orientation=="landscape":
            DisplayServer.screen_set_orientation(SCREEN_LANDSCAPE)
          否则保持竖屏不变
       d. GameHost.visible = true
     ⚠️ 分辨率切换必须在遮罩完全不透明后执行，避免用户看到布局跳变的一帧
7.   按运行时分发:
      ├─ pck/html: module = load(meta.scene).instantiate(); GameHost.add_child
      └─ arcade:  RunnerRegistry.dispatch → ArcadeRunner 接管(见下方 Arcade 分支)
8.   module.boot(ctx)        # ctx 含 save_dir/trial_mode/owned/best/viewport_size
9.   Analytics.track(game_launch)
─── 游戏运行中（系统来电/通知 → Shell 调 module.pause_game/resume_game）───
10.  module 发 quit_requested(result)
11.  Launcher 冻结 module（暂停其 _process）
     转场遮罩淡入（遮住分辨率恢复过程）：
       a. get_tree().root.content_scale_size = Vector2i(720, 1560)  # 恢复 Shell 基准
       b. DisplayServer.screen_set_orientation(SCREEN_PORTRAIT)      # 恢复竖屏
       c. module.queue_free()；GameHost.visible = false；App.visible = true
     遮罩淡出，用户看到的始终是正确布局
12.  DB.upsert_record(gid, {total_playtime+=result.playtime,
                            best=max(best, result.score),
                            finish_count+1})  ← 强写
13.  AchievementEngine 消化 result.achievements → Toast 逐条弹出
     renderHome() 立即重渲染首页"继续游戏"区块（更新 last_played 顺序）
14.  **等待 400ms**（给游戏退出动画/音效留尾气），然后弹出 ResultOverlay：
        ┌ 游戏名 / 本局: 分数、时长 / 历史最高
        ├ 新解锁成就(如有)
        ├ [▶ 再来一局]  [✎ 评价*]  [返回]
        │   *playtime 门槛≥600s 才亮，否则置灰+"还需X分钟"
        ├ "换一个" 迷你卡区：similar(gid,3) 横列；每张卡点击 = closeOv + openDetail(id)
        │   similar() 为空时隐藏整个迷你卡区
15.  用户选择:
       再来一局 → 重走步骤 6b-8（重新设置分辨率/方向，因为恢复时已经重置）；
                  Arcade = CoreManager.launch(rom) 重新拉起 MAME Activity（v2）
                   ⚠️ Arcade 拉起/收回依赖 MameRuntime 插件（v2 spike 未开始，M2 真机验证前置，review P-4）
       其他     → 回 Tab 根页（分辨率已在步骤11恢复，直接显示）
16.  **结算卡评价引导（按被评价游戏，review P-1）**：该游戏累计 `finish_count` 达门槛（如 ≥2）且本次 `playtime≥600s` 时，在该游戏结算卡引导写评价；计数按游戏独立维护，消除"仅全局第 2 次"的错位与一次性触发
```

**视口切换 GDScript 参考实现：**

```gdscript
# services/launcher.gd

const SHELL_VP  := Vector2i(720, 1560)
const SHELL_ORI := DisplayServer.SCREEN_PORTRAIT

func _apply_game_viewport(meta: Dictionary) -> void:
    var vp_arr: Array = meta.get("viewport_size", [720, 1560])
    var vp := Vector2i(vp_arr[0], vp_arr[1])
    var ori: String = meta.get("orientation", "portrait")
    
    get_tree().root.content_scale_size = vp
    if ori == "landscape":
        DisplayServer.screen_set_orientation(DisplayServer.SCREEN_LANDSCAPE)
    # portrait 时不调用，保持当前竖屏方向（避免不必要的重排）

func _restore_shell_viewport() -> void:
    get_tree().root.content_scale_size = SHELL_VP
    DisplayServer.screen_set_orientation(SHELL_ORI)
```

`_apply_game_viewport` 在 `TransitionLayer.fade_out()` 的 `await` 之后调用，`_restore_shell_viewport` 在 `module.queue_free()` 之前、遮罩不透明时调用。

> **A-6 已关闭（4.5 headless 实测）**：`root.content_scale_size` 运行时切换有效——`tools/test_viewport.gd` 实测 visible rect 1560×1560 →（切 1920×1080）→ 1920×1920 →（恢复）→ 1560×1560，canvas 变换按 expand 公式即时重算，参考实现成立。`screen_set_orientation` 真机行为 headless 不可验，归 M2 真机项。

**ArcadeRunner 分支（替代上述 7→8→9→10→11 步，其余一致；v2）：**

```
7b. ArcadeRunner.launch(ctx, rom, meta):
      - CoreManager.ensure_native() → 未就绪则 DownloadDialog 显示进度（libMAME4droid.so ~74MB/ABI）
      - MameRuntime.launch(rom, extras)   # startActivity→MAME 游戏 Activity（extra=ROM 路径/cfg/buttonConfig）
      - Godot 主 Activity 后台 pause；游戏中 native 侧独立渲染（GLSurfaceView + OpenSL），Godot 不参与
8b. 游戏中: native 侧独立渲染（GLSurfaceView + OpenSL 音频），Godot 不参与帧/音频/输入
10b. 游戏内退出 → Activity finish() → MameRuntime 发 game_exited → Launcher 接收，走步骤 11-16
12b. (同 12): 写 DB.upsert_record；score 读 hiscore/nvram（MAME 侧写 game_dir），提取不到 score:-1
14b. (同 14): 结算卡显示; [再来一局] = CoreManager.launch(rom) 重新拉起 Activity（内置/自导入 ROM 同此路径）
```

**"再来一局"语义（v2）：**

| ROM 来源 | [再来一局] | [返回] |
|---|---|---|
| 内置/授权 ROM | `launch(rom)` 重新拉起 Activity（每局一次 Activity 生命周期） | `game_exited` → 回盒子 |
| 用户自导入 ROM | 同上（路径经 Intent extra，runtime §3.4 自导入路线） | 同左 |
| PCK/HTML 游戏 | module 复位重 boot()（不重建场景） | module.queue_free() + App 淡入 |

> v2 无 mjarch3 killProcess 分支：Godot 进程全程不杀、无进程内 native 状态，「再来一局」= 重新 launch。

### 4.7 购买流程（现 Mock、后接真渠道）

```
[¥6 购买] → PurchaseDialog(模态):
   展示: 图标/名/价格/渠道选择(现仅"模拟支付")
   [确认支付]
→ PayService.create_order(gid):
   order = {id, gid, amount, status:pending, created_at}
   DB.put_order()  ← 先持久化再支付（崩溃可恢复）
→ Mock: 900ms spinner 动画 → 成功 / 真渠道: 调起SDK支付回调
→ PayService.finalize(order):
   status=paid, receipt=票据存档 → owned 判定通过
→ EventBus.payment_succeeded
   → 关闭 PurchaseDialog → Toast "支付成功 · 已解锁 ▸"
   → CTA 变 [▶ 开玩] + renderLib('owned')
   → **900ms 后**（与 Toast 错开防重叠）再弹成就 Toast "🏆 成就解锁：XXX +N 点"
取消/失败 → order 停留 pending → 下次启动 retry_pending_orders()
我的→钱包订单: [恢复购买] = 遍历 pending 重放
```

**DownloadDialog UX 序列（DLC / Arcade rom zip）：**
```
打开弹窗 → 显示"游戏图标 + 名称 + 扩展包 · 校验 SHA-256"
→ 进度条每约 220ms 随机递进（上限 16%/tick，模拟；真实用 HTTPRequest.get_downloaded_bytes）
→ 进度文案："下载中 XX% · 预计 Xs"（倒计时 = ceil((100-p)/速度)）
→ p=100：进度文案变 "✓ 安装完成"，等待 500ms
→ 自动关闭弹窗 + Toast "已安装 · 可开玩" + DetailPage CTA 自动刷新
失败 → 进度文案变 "⚠ 下载失败" + [重试] 按钮（支持 HTTP Range 断点续传）
```

### 4.8 评价中心与写评价

```
ReviewsPage:
  头部: ★平均分 + 评分分布条(5档, 可点筛选) + [写评价]
  筛选chips: 最有用(赞数) / 最新 / 好评 / 差评
  列表项: ★ + 文字 + [游玩3.2小时]徽章 + 日期 + 👍数 + [本地]标记
  自己的评价置顶, 带 [修改][删除]

ReviewEditor(模态):
  5星选择(大, 点第3颗=3星) + 文本框(500字, 实时计数)
  提交校验:
    ① playtime ≥ 600s (页面级门槛, 编辑器内再校验)
    ② 星级必选; 文字可为空(纯星级评价)
    ③ 敏感词/长度 → 本地基础过滤(云端阶段服务端复审)
  → DB.put_review() 立即可见(status=local)
  → 修改=覆盖同一条并刷新时间; 删除=移除并 EventBus.review_deleted
```

### 4.9 我的（Library Tab）

```
头部: 随机头像+昵称(可改, 本地) │ 总成就点 │ 累计时长
游戏库(三子Tab): 已拥有 / 试玩中(剩次角标) / 最近（按 last_played 倒序，含已卸载 DLC 带下载角标）
成就墙: 总进度条 + 按游戏分组 + 全局成就(收藏家/好评人/马拉松)
钱包订单: 订单列表(状态) + [恢复购买]
设置: 音效/震动/语言(zh,en)/动效开关/清理缓存/核心管理(Arcade: 版本/删除重下)
      ROM 管理(Arcade): 扫描用户 ROM（[扫描 user://roms/ 目录]按钮，手动触发）
                        → 发现的 ROM zip 列表（标注「自导入」）→ 点击可直接启动
      关于/隐私政策
```

> **Library 第3 Tab 说明**："最近"显示所有有过游玩记录的游戏（按 last_played 倒序），不限拥有状态；已卸载的 DLC 游戏也在此显示，带 ⬇ 角标方便重新下载。这和"已拥有"Tab 的区别在于：已拥有 = 权益层面，最近 = 游玩历史层面。

**全局成就定义**位于 `data/achievements.json`，由 Registry 统一加载（与游戏级成就共用 `ach_def(aid)` 接口）。全局成就示例：

```json
[
  {"id": "global_collector",  "name": "收藏家",   "desc": "拥有 5 款游戏",     "points": 30},
  {"id": "global_reviewer",   "name": "好评人",   "desc": "提交 3 条游戏评价",  "points": 20},
  {"id": "global_marathon",   "name": "马拉松",   "desc": "累计游玩时长超 10小时", "points": 50}
]
```

AchievementEngine 在 `evaluate(event)` 中额外消化全局成就触发条件（游戏数/评价数/时长跨游戏累加），判定逻辑与游戏级成就相同，解锁后走同一 Toast 通知流。

### 4.10 空态/错态/加载态（每页必备）

| 页面 | 加载 | 空态 | 错态 |
|---|---|---|---|
| 首页 | 骨架屏 shimmer | 不会空（冷启动有内置游戏+编辑榜） | — （本地） |
| 分类 | 即时 | 无结果→"该筛选下暂无游戏"+清筛选按钮 | — |
| 搜索 | 即时 | 无结果+猜你喜欢 | — |
| 详情 | 骨架屏 | meta缺失→"游戏已下架" | — |
| 评价 | — | "还没有评价, 来抢沙发" + 写评价CTA | — |
| DLC下载 | 进度%+速度 | — | 失败→[重试]（断点续传Range） |
| 订单 | — | "暂无订单" | pending 横幅提示重试 |
| ResultOverlay "换一个" | — | `similar(gid)` 为空时（目录只有 1 款游戏或无同类）→ 隐藏"换一个"按钮，仅保留[再来一局]/[评价]/[返回] | — |

---

## 5. 本地存储设计（user://）

> A-4（review）：下方为概念布局 Sketch；实现采用 `core/db.gd` 的合并布局 `user://db/{profile,records,reviews,orders,daily,search_history,achievements}.json`，功能等价、写入分级一致。

```
user://
  profile.cfg            # ConfigFile: [user]昵称/头像idx  [settings]音效/震动/语言
  records/<gid>.json     # PlayRecord: last_played/total_playtime/sessions/best/finish_count/trial_used
  reviews/<gid>.json     # 本地评价(含编辑历史)
  orders.json            # 订单数组(pending/paid)
  achievements.json      # {aid: unlocked_at}
  daily.json             # 每日任务进度 + date
  search.json            # 历史 + 词频统计
  saves/<gid>/           # 各游戏私有区（Shell 不读内容）
  <filesDir>/mame/       # 核心 .so（v2：libMAME4droid.so 按 ABI ~74MB + 版本标记，Android app 私有目录，见 runtime-design §3.5）
  games/<gid>/           # Arcade 游戏专属文件夹（roms/cfg/states/nvram，见 runtime-design §3.4-§3.5）
  dlcs/<gid>.pck
  logs/analytics-YYYYMMDD.jsonl
```

写策略：DB 内部脏标记分两级——**强写**（立即落盘）用于 `trial_used / orders / playtime / best / finish_count`；**防抖写**（500ms 合并）用于 `search_history / daily / profile`。关键节点触发强写：`game_finished`（时长/分数）、`trial_consumed`（试玩计数）、`payment_*`（订单状态/权益）。

---

## 6. 试玩与权益判定（TrialGuard/PayGate 汇总）

```
owned(gid):
  price_model in [free, ad, iap]        → true
  paid: order.status==paid              → true
  trial: 已购同上；否则走试玩计数
left(gid): meta.trial.plays - record.trial_used
consume(gid): record.trial_used += 1 → 强写 → 返回剩余
弹购买的两个触发点: ①CTA点击时判定不过 ②试玩耗尽当次结算卡提示
```

---

## 7. 客户端推荐（Recommender 本地版）

```
hot_score(g) = .5*norm(players) + .3*norm(rating*rating_count) + .2*new_boost
continue_row : records 按 last_played 倒序 top5（DLC 未装也显示, 带下载角标）
for_you      : 用户玩过游戏 tags 并集 → 候选(去已拥有) → hot_score 排序
               曝光频控: exposed_at 7天内不重复 → 存 daily.json
charts       : 综合=hot_score / 新游=version 日期加权 / 好评=rating*count
similar(gid) : 同 category + 共享tag≥1 → hot_score → top3（结算卡"换一个"用）
冷启动       : 无任何记录 → 精选编辑榜(本地json)
```

---

## 8. 主题与组件库

两套主题共用一套场景树与交互逻辑，只切换 Theme 资源（Godot 侧）或 CSS 变量集（原型侧）。切换入口：首页右上角 🎨 按钮 / 设置→外观风格，即时生效，记住选择（写 `profile.cfg [settings] theme`）。

> **注意：两套主题均禁用 hdr_2d/glow**（mobile 渲染器实心矩形 bug，已踩坑）。发光效果通过分层半透明矩形模拟（霓虹：叠加高亮色；青瓷：省略）。

### 8.1 主题 Token 对照表

以下 token 名与原型 `nova-arcade-prototype.html` CSS 变量一一对应，Godot 侧按同名定义 Theme 常量。

| Token | 霓虹·星穹 (neon) | 青瓷·素笺 (elegant) | 用途 |
|---|---|---|---|
| `--bg` | `#05060f` | `#f6f4ef` | 全局背景 |
| `--bg-out` | `#020308` | `#eae6dd` | 手机壳外背景 |
| `--card` | `rgba(10,16,34,.72)` | `#ffffff` | 卡片/面板底色 |
| `--card2` | `rgba(14,22,48,.92)` | `#ffffff` | 次级卡片/输入框底色 |
| `--ink` | `#cfeaff` | `#2f3a35` | 主文字 |
| `--ink2` | `#8fb8d8` | `#6f7a74` | 次级文字 |
| `--ink3` | `#5b7d99` | `#9aa5a0` | 辅助/占位文字 |
| `--accent` | `#2be8ff` | `#4a7d64` | 主强调色（Tab 选中/Chip 选中/进度条） |
| `--accent-soft` | `rgba(43,232,255,.10)` | `#e7f0ea` | 强调色浅底（Chip 选中背景等） |
| `--alt` | `#c86bff` | `#9d86b5` | 次强调色（DL 按钮/稀有卡边框） |
| `--alt-soft` | `rgba(200,107,255,.12)` | `#efe9f5` | 次强调色浅底 |
| `--gold` | `#ffd23f` | `#b8934a` | 金色（评分星/成就/传说稀有） |
| `--gold-soft` | `rgba(255,210,63,.16)` | `#f6efdd` | 金色浅底 |
| `--ok` | `#45ff88` | `#4a7d64` | 正向状态色（好评率/游玩时长徽章） |
| `--play-grad` | `linear-gradient(90deg,#2be8ff,#c86bff)` | `#4a7d64`（纯色） | 开玩按钮背景 |
| `--play-ink` | `#04101c` | `#fdfcf9` | 开玩按钮文字 |
| `--play-sh` | `0 0 20px rgba(43,232,255,.35)` | `0 4px 14px rgba(74,125,100,.3)` | 开玩按钮投影 |
| `--buy-grad` | `linear-gradient(90deg,#ffd23f,#ff9838)` | `#c98f5f`（纯色） | 购买按钮背景 |
| `--buy-ink` | `#241300` | `#fdfaf5` | 购买按钮文字 |
| `--buy-sh` | `0 0 20px rgba(255,210,63,.3)` | `0 4px 14px rgba(201,143,95,.3)` | 购买按钮投影 |
| `--line` | `rgba(43,232,255,.18)` | `#e6e1d6` | 常规分割线/边框 |
| `--line2` | `rgba(255,255,255,.08)` | `#efece4` | 次级分割线 |
| `--ic-border` | `1px solid rgba(255,255,255,.1)` | `1px solid rgba(255,255,255,.7)` | 游戏图标边框 |
| `--star-off` | `#3a5268` | `#dcd6c8` | 星级未选中颜色 |
| `--rank1/2/3` | `#ffd23f / #dfe8f2 / #ff9838` | `#b8934a / #6f7a74 / #c98f5f` | 排行榜前3名颜色 |
| `--banner-grad` | `linear-gradient(120deg,#12224d,#3a1560)` | `linear-gradient(115deg,#e3ece5,#ece3f0,#f4ead9)` | Banner 背景渐变 |
| `--banner-line` | `rgba(200,107,255,.35)` | `rgba(255,255,255,.8)` | Banner 边框 |
| `--shot-grad` | `linear-gradient(140deg,#101c3f,#241040)` | `linear-gradient(140deg,#e5ece6,#e9e1ef)` | 截图占位背景 |
| `--av-grad` | `linear-gradient(140deg,#2a3c6e,#5a2a80)` | `linear-gradient(140deg,#dbe7dd,#e6dcee)` | 头像渐变背景 |
| `--libhead-grad` | `linear-gradient(130deg,rgba(43,232,255,.09),rgba(200,107,255,.09))` | `linear-gradient(120deg,#e9f0ea,#efe9f3)` | Library 头部卡片背景 |
| `--ov-bg` | `rgba(3,4,12,.78)` | `rgba(58,66,60,.32)` | 遮罩层背景 |
| `--ov-blur` | `5px` | `4px` | 遮罩层模糊 |
| `--toast-bg` | `rgba(10,18,38,.95)` | `rgba(47,58,53,.92)` | 普通 Toast 背景 |
| `--toast-ink` | `#dff4ff` | `#f2f0e9` | 普通 Toast 文字 |
| `--toast-line` | `rgba(43,232,255,.3)` | `rgba(47,58,53,.2)` | 普通 Toast 边框 |
| `--ach-toast-bg` | `#ffd23f` | `#b8934a` | 成就 Toast 背景 |
| `--ach-toast-ink` | `#241300` | `#fdfaf0` | 成就 Toast 文字 |
| `--sw-off/on` | `rgba(255,255,255,.12) / rgba(43,232,255,.5)` | `#e6e1d6 / #4a7d64` | 开关轨道颜色 |
| `--bar-grad` | `linear-gradient(90deg,#2be8ff,#c86bff)` | `linear-gradient(90deg,#4a7d64,#9d86b5)` | 进度条渐变 |
| `--rbar-bg` | `rgba(255,255,255,.07)` | `#efece4` | 评分分布条背景 |
| `--rbar-grad` | `linear-gradient(90deg,#ffd23f,#ff9838)` | `linear-gradient(90deg,#b8934a,#c98f5f)` | 评分分布条填充 |
| `--ach-bg/line` | `rgba(255,210,63,.07) / rgba(255,210,63,.35)` | `#f6efdd / #e8ddc0` | 成就区块背景/边框 |
| `--myrev-bg/line` | `rgba(255,210,63,.04) / rgba(255,210,63,.4)` | `#fdfbf4 / #dfd6bd` | 自己的评价背景/边框 |
| `--score-grad` | `linear-gradient(180deg,#fff,#2be8ff)` | `linear-gradient(180deg,#4a7d64,#4a7d64)` | 结算卡分数渐变文字 |
| `--sechead-mark` | `"▶ "` | `"❋ "` | SectionHeader 前缀符号 |
| `--serif` | `inherit` | `Georgia,serif` | 衬线字体（排名数字/分数） |
| `--sh` | `none` | `0 2px 10px rgba(90,105,98,.07)` | 卡片投影 |
| `--sh2` | `0 0 60px rgba(43,232,255,.12),0 0 120px rgba(200,107,255,.08)` | `0 6px 24px rgba(90,105,98,.12)` | 外层容器投影 |
| `--iconEnd` | `#0a1030` | `#f4f1ea` | 游戏图标渐变终止色（非 CSS 变量，GDScript 常量） |

### 8.2 背景层差异

| | 霓虹·星穹 | 青瓷·素笺 |
|---|---|---|
| 背景粒子 | 两个大色块 nebula（青/紫），`opacity:.14`，`blur:50px` | 三个色块 nebula（青绿/淡紫/米白），`opacity:.5`，`blur:64px`（柔和光晕） |
| 纹理层 | 透视网格线（底部 34% 高度，`perspective rotateX(56deg)`，青色细线） | 点阵（底部 30% 高度，22×22px，墨绿色 `opacity:.05`） |
| 扫描线 | 全屏 `repeating-linear-gradient`（见 `tetra-nova.html`，Arcade Shell 级启用） | 无 |

### 8.3 Godot 落地方案

```
shell/theme/
  theme_neon.tres     # 霓虹 Theme 资源（Panel/Label/Button/LineEdit 等控件样式）
  theme_elegant.tres  # 青瓷 Theme 资源
  theme_tokens.gd     # const 字典，存上表所有 token 值（两套），供非 Theme 控件（Canvas/Shader/ColorRect）读取

# ThemeTokens.gd 示例结构:
const TOKENS = {
    "neon": {
        "bg": Color(0.02, 0.024, 0.059),
        "accent": Color(0.169, 0.91, 1.0),
        "iconEnd": Color(0.039, 0.063, 0.188),
        # ... 完整 token 列表同 §8.1 表格
    },
    "elegant": {
        "bg": Color(0.965, 0.957, 0.937),
        "accent": Color(0.29, 0.49, 0.392),
        "iconEnd": Color(0.957, 0.945, 0.918),
        # ...
    }
}
static var current: String = "neon"
static func color(key: String) -> Color:
    return TOKENS[current].get(key, Color.WHITE)
static func icon_grad(game_col: Color) -> Gradient:
    var g = Gradient.new()
    g.add_point(0.0, game_col)
    g.add_point(1.0, color("iconEnd"))
    return g

# 切换逻辑（在 Autoload Sound 或独立 ThemeManager 单例中）:
func apply_theme(name: String) -> void:
    ThemeTokens.current = name
    get_tree().root.theme = load("res://shell/theme/theme_%s.tres" % name)
    DB.save_profile({"theme": name})
    EventBus.emit_signal("settings_changed", "theme", name)
    rerenderAll()   # 同步重渲染所有活跃页面

func rerenderAll() -> void:
    # 通知各页面自行重渲染（通过 EventBus 或直接调用 Nav.current_page.on_theme_changed()）
    # 背景层通过 settings_changed 信号处理，其余页面监听此信号重建 icon_grad 等动态颜色
    EventBus.emit_signal("theme_changed", ThemeTokens.current)
```

背景层（BgLayer）监听 `settings_changed("theme", ...)` 或 `theme_changed` 信号，切换时：
- 更新 3 个 nebula ColorRect 的 color + modulate.a
- 切换 GridLines（霓虹可见）和 DotGrid（青瓷可见）的 `visible`
- `Tween` 过渡颜色 400ms，避免瞬切

### 8.4 组件库

| 组件 | 要点 |
|---|---|
| GameCard S/M/L | 三规格；图标背景为**动态渐变**：`linear-gradient(150deg, game.col, --iconEnd)`（霓虹 `--iconEnd=#0a1030`，青瓷 `=f4f1ea`）；浅色 soft 变体：`game.col+'55'`（33% alpha）用于 Library 行；角标系统(状态/标签)统一组件；卡片底色用 `--card`，投影用 `--sh` |
| StarRow | 只读展示与可交互两种；支持半星渲染；星色用 `--gold`，未选用 `--star-off` |
| TagChip | 颜色按类型(分类/付费/离线)；边框/文字随主题 `--line`/`--ink2`；dim 变体用于次要标签（付费模式/DLC） |
| SectionHeader | 前缀符号来自 `--sechead-mark`（霓虹"▶ "，青瓷"❋ "） |
| EmptyState | 插画位（大号 emoji，`opacity:.45`）+ 文案 + CTA 按钮 |
| Toast | 顶部滑入，弹簧曲线 `cubic-bezier(.34,1.4,.64,1)`，300ms；**展示 2200ms**；成就 Toast 用 `--ach-toast-bg`（霓虹金色背景深色字，青瓷棕金） |
| ShimmerSkeleton | 加载占位；shimmer 动效颜色基于 `--line`（深色主题低对比，浅色主题白色光带） |
| PullRefresh | 首页/分类；下拉阈值 60px；松手触发 `renderHome()` / `renderCat()`；转圈动画 400ms |
| Dialog 基类 | 弹出动画：`translateY(40px) scale(.92) opacity:0 → normal`，**280ms**，`cubic-bezier(.34,1.5,.64,1)`；模态背景 `--ov-bg` + `blur(--ov-blur)`；返回键关模态优先于 pop 页面 |
| ThemeToggleButton | 首页右上角 🎨 / 设置页外观风格选择；按下 `scale(.9)`；切换后立即调用 `rerenderAll()`（6 个渲染函数同步重渲染：home/charts/cat/search/lib/detail）；切换时播放 Sound.toggle() |
| CTA 固定区 | 按钮 + hint 行容器，固定在 DetailPage 底部；渐变淡出遮罩（`linear-gradient(0deg, --bg 55%, transparent)`）覆盖在内容滚动层上方 |
| ProgressBar | DLC 下载/成就进度；轨道背景 `--rbar-bg`，填充渐变 `--bar-grad` |

### 8.5 CSS → Godot Control 映射

原型用 CSS 描述的每一类视觉效果，在 Godot 4.5 中的对应实现方案。

#### 背景色 / 纯色填充

```
CSS:    background: #05060f
Godot:  ColorRect  (color = Color(0.02,0.024,0.059))
        或 PanelContainer + StyleBoxFlat.bg_color
```

`StyleBoxFlat` 是 Panel/Button/LineEdit 等控件的 Theme 样式资源，承载背景色、圆角、边框、投影。绝大多数容器用它。`ColorRect` 只用于纯色无交互的背景层（如 BgLayer nebula）。

#### 圆角

```
CSS:    border-radius: 20px
Godot:  StyleBoxFlat.corner_radius_top_left    = 20
        StyleBoxFlat.corner_radius_top_right   = 20
        StyleBoxFlat.corner_radius_bottom_left = 20
        StyleBoxFlat.corner_radius_bottom_right= 20
```

四个角可以单独设，卡片/图标 20px，按钮 14px，Chip 15px（全圆），对照 §8.1 的 token 换算（1.92×CSS 值）。

#### 半透明背景（毛玻璃卡片）

```
CSS:    background: rgba(10,16,34,.72); backdrop-filter: blur(6px)
Godot:  StyleBoxFlat.bg_color = Color(0.039,0.063,0.133, 0.72)
        # backdrop-filter blur：Godot mobile 渲染器无原生支持
        # → 用视觉近似：深色半透明底 + 细边框即可，视觉效果足够
        StyleBoxFlat.border_width_* = 1
        StyleBoxFlat.border_color = Color(accent, 0.18)
```

`backdrop-filter: blur` 在 Godot mobile 渲染器里没有对应实现，但实际效果上深色半透明 + 细边框已经足够接近原型视觉，不需要真实模糊。

#### 线性渐变背景

```
CSS:    background: linear-gradient(90deg, #2be8ff, #c86bff)
Godot:  方案A（推荐，纯代码）：
          var gt := GradientTexture2D.new()
          gt.gradient = Gradient.new()
          gt.gradient.set_color(0, Color("#2be8ff"))
          gt.gradient.set_color(1, Color("#c86bff"))
          gt.fill_from = Vector2(0, 0.5)   # 90deg = 水平
          gt.fill_to   = Vector2(1, 0.5)
          $Button.icon = gt   # 或 TextureRect / StyleBoxTexture

        方案B（编辑器资源）：
          GradientTexture2D.tres 直接赋给控件的 texture 属性
```

按钮渐变背景用 `TextureButton` 或在 `_draw()` 里手动 `draw_texture_rect`。  
卡片/Banner 渐变背景用 `TextureRect`（stretch_mode = STRETCH_SCALE）叠在内容下方。

#### 投影（box-shadow）

```
CSS:    box-shadow: 0 0 20px rgba(43,232,255,.35)
Godot:  StyleBoxFlat:
          shadow_color  = Color(0.169, 0.91, 1.0, 0.35)
          shadow_size   = 20        # 对应 blur radius
          shadow_offset = Vector2(0, 0)
```

霓虹主题的外发光（`box-shadow` 无偏移）用 `shadow_offset=Vector2(0,0)` 即可。  
青瓷主题的方向性投影（`0 2px 10px`）设 `shadow_offset=Vector2(0,2)`。  
⚠️ StyleBoxFlat 的投影是矩形 blur，不完全等同 CSS box-shadow，但在 mobile 上视觉效果可接受。

#### 渐变文字（-webkit-background-clip: text）

```
CSS:    background: linear-gradient(180deg,#fff,#2be8ff);
        -webkit-background-clip: text; color: transparent
Godot:  Label + ShaderMaterial（最干净的方案）:
```

```glsl
// gradient_text.gdshader
shader_type canvas_item;
uniform vec4 color_top : source_color = vec4(1.0);
uniform vec4 color_bot : source_color = vec4(0.169, 0.91, 1.0, 1.0);
void fragment() {
    vec4 txt = texture(TEXTURE, UV);
    COLOR = mix(color_top, color_bot, UV.y) * vec4(1.0, 1.0, 1.0, txt.a);
}
```

结算卡大分数、Logo 文字用这个 shader。其他不需要渐变的 Label 直接设 `font_color`。

#### 圆形/圆角图片裁切（border-radius on image）

```
CSS:    border-radius: 50%;  overflow: hidden  (头像圆形裁切)
Godot:  TextureRect + ClipContents = true 放在圆角 Panel 里
        或 SubViewport 渲染后裁切（较重，不推荐）
        推荐方案：
          Panel (StyleBoxFlat, corner=radius, clip_contents=true)
            └── TextureRect (expand=true, stretch=KEEP_ASPECT_COVERED)
```

#### 边框（border）

```
CSS:    border: 1px solid rgba(43,232,255,.18)
Godot:  StyleBoxFlat:
          border_width_top/right/bottom/left = 1
          border_color = Color(0.169, 0.91, 1.0, 0.18)
```

#### Tween 动画替代 CSS transition / animation

```
CSS:    transition: transform .15s
Godot:  var tw := create_tween()
        tw.tween_property(node, "scale", Vector2(0.97, 0.97), 0.15)

CSS:    animation: dlgIn .28s cubic-bezier(.34,1.5,.64,1)
Godot:  tw.set_trans(Tween.TRANS_SPRING).set_ease(Tween.EASE_OUT)
        # 或手动用 TRANS_CUBIC + 自定义曲线
        tw.tween_property(node, "position:y", target_y, 0.28)
        tw.parallel().tween_property(node, "scale", Vector2.ONE, 0.28)
        tw.parallel().tween_property(node, "modulate:a", 1.0, 0.28)
```

Godot `Tween.TRANS_SPRING` 近似 `cubic-bezier(.34,1.5,.64,1)`（弹簧过冲）。  
精确曲线用 `tween_method` + 自定义插值函数。

| CSS 动画场景 | Godot Tween 方案 |
|---|---|
| 页面推入（translateX + opacity，220ms） | `TRANS_CUBIC EASE_OUT`，`position.x` + `modulate.a` |
| Dialog 弹出（translateY + scale + opacity，280ms） | `TRANS_SPRING EASE_OUT`（弹跳感）|
| Toast 滑入（translateY，300ms） | `TRANS_BACK EASE_OUT` |
| 按钮按下（scale .97，instant） | `TRANS_LINEAR`，duration 0.08s |
| Tab 切换（opacity，150ms） | `TRANS_SINE EASE_IN_OUT` |
| 主题切换背景色（400ms） | `tween_property(color_rect, "color", ...)` |

#### 横向滚动容器（HScrollContainer）

```
CSS:    display:flex; overflow-x:auto; scrollbar-width:none
Godot:  ScrollContainer (horizontal_scroll_mode=SCROLL_MODE_AUTO,
                         vertical_scroll_mode=SCROLL_MODE_DISABLED)
          └── HBoxContainer
        隐藏滚动条：ScrollContainer 的 theme override
          scrollbar_h_separation = 0  并设 HScrollBar 的 minimum_size.y = 0
```

#### 固定定位底部 CTA（position: fixed; bottom: 0）

```
CSS:    position: fixed; bottom: 0; left: 0; right: 0
Godot:  Control (anchor_left=0, anchor_right=1,
                 anchor_bottom=1, anchor_top=1,
                 offset_top=-120)   # 120px 高度
        放在页面场景树的最后（最高 z-index）
        渐变遮罩：GradientTexture2D（下不透明→上透明）赋给 TextureRect
```

#### 背景层 nebula（大色块模糊）

```
CSS:    .neb { border-radius:50%; filter:blur(50px); opacity:.14 }
Godot:  ColorRect (color=accent_color, modulate.a=0.14)
          + BackBufferCopy / CanvasGroup 实现模糊
        简化方案（推荐）：
          不用真实模糊，直接用大半径 StyleBoxFlat shadow 或
          GradientTexture2D（中心不透明→边缘透明的径向渐变）
          visual 效果等价，性能更好
```

```gdscript
# BgLayer.gd — nebula 实现
func _setup_nebula() -> void:
    for neb in $Nebulas.get_children():
        var gt := GradientTexture2D.new()
        gt.fill = GradientTexture2D.FILL_RADIAL
        gt.fill_from = Vector2(0.5, 0.5)
        gt.fill_to   = Vector2(1.0, 0.5)
        gt.gradient  = Gradient.new()
        gt.gradient.set_color(0, neb_color)           # 中心不透明
        gt.gradient.set_color(1, Color(neb_color, 0)) # 边缘透明
        neb.texture = gt   # TextureRect
```

#### 透视网格线（霓虹主题底部）

```
CSS:    background: repeating-linear-gradient(...);
        transform: perspective(300px) rotateX(56deg) scale(1.6)
Godot:  _draw() 在 Control 子类里手动绘制网格线：
```

```gdscript
# GridLines.gd
extends Control
func _draw() -> void:
    var w := size.x * 1.6
    var h := size.y
    var col := Color(ThemeTokens.color("accent"), 0.05)
    # 横线
    var step_h := 26.0
    for y in range(0, int(h), int(step_h)):
        draw_line(Vector2(-w*0.3, y), Vector2(w*1.3, y), col, 1.0)
    # 竖线（加透视变形 - 用 draw_set_transform 模拟）
    var step_v := 34.0
    for x in range(0, int(w), int(step_v)):
        var xf := (x / w - 0.5) * 1.6
        draw_line(Vector2(x - w*0.3, 0), Vector2(w*0.5 + xf*w, h), col, 1.0)
```

透视变形靠顶点计算模拟，不用真实 3D。主题切换时 `visible = (theme == "neon")`。

---



| 埋点事件 | 属性 |
|---|---|
| app_open / app_close | duration |
| tab_switch / page_view / card_expose / card_click | page, gid, section |
| game_launch / game_finish | gid, runtime, playtime, score, trial |
| arcade_launch / arcade_finish | gid, rom_name, core_version, playtime, score |
| core_installed / core_upgrade / core_removed | version, result |
| rom_imported / rom_scan_done | count, duration |
| purchase_click / purchase_success / purchase_fail | gid, amount, channel |
| review_open / review_submit | gid, stars, has_text |
| search_query | q_len, result_n, click_gid |
| dlc_download | gid, result |
| achievement_unlock | aid |

本地 JSONL 落盘；M3 起批量上报（WiFi+充电时优先）。

---

## 10. 工程目录（新工程 nova-arcade/）

```
nova-arcade/
  project.godot            # autoload×6 注册（EventBus/QualitySettings/DB/Registry/Nav/Sound）
  shell/
    main.tscn / theme/  components/  pages/  overlays/
  games/
    tetra_nova/            # 迁入现有工程, 加 meta.json + GameModule 适配层
  services/                # launcher/pay/trial/search/recommend/analytics/ach/dlc
  core/                    # event_bus/db/registry/nav/sound (autoload 脚本)
  data/editorial.json      # Banner/精选/分类配置
```

---

## 11. M1 任务拆解（可开工粒度）

- [ ] 新工程 + autoload×6 + 主题 Token + TabBar/PageStack/转场
- [ ] DB/Registry + GameMeta 定义 + 内置 tetra_nova meta
- [ ] CoreManager（Android 委托 MameRuntime 插件 v2；桌面返回「不可用」，见 runtime §3.2）
- [ ] TETRA NOVA 迁入 + GameModule 适配（boot/quit/result/存档目录）
- [ ] Launcher 全时序（含权益门/转场/结算卡骨架）
- [ ] HomePage 五区块 + GameCard 组件
- [ ] CategoryPage / SearchPage(+Searcher) / DetailPage(CTA状态机)
- [ ] ReviewsPage + ReviewEditor（含时长门槛）
- [ ] PurchaseDialog + PayService(Mock) + 订单持久化/重试
- [ ] LibraryPage(库/成就/订单/设置/核心管理) + AchievementEngine
- [ ] 空态/加载态/Toast/埋点补齐 + 全流程回归脚本（延续 shot.gd 截图法）
```
