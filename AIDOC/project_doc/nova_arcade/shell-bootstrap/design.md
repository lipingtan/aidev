# 设计：CR-1 shell-bootstrap（Shell 工程初始化 + Autoload 骨架）

> 依据：已确认的 `design_plan.md`（Q1 确认 / Q2=A Android先行 / Q3=A CoreManager接口桩 / Q4=A 含DB/Registry骨架）。
> 上游：`nova-arcade-client-design.md`（下称 CD）、`dev-workflow-override.md` §2、`hybrid-project-workflow.md` §3.2。

## 1. 工程与目录

新工程 `projects/nova_arcade/nova-arcade/`，结构按 CD §10：

```
nova-arcade/
├── project.godot            # autoload×6 注册
├── shell/
│   ├── main.tscn          # 主场景
│   ├── theme/             # theme_neon.tres / theme_elegant.tres / theme_tokens.gd
│   ├── components/        # tab_bar.gd 等
│   ├── pages/             # page.gd（基类）+ home/category/search/library 四根页骨架
│   └── layers/            # bg_layer.gd / game_host.gd / overlay_layer.gd / toast_layer.gd
├── core/                  # event_bus / quality_settings / db / registry / nav / sound（autoload 脚本）
├── services/              # core_manager.gd（Mock 桩）
├── games/
│   └── tetra_nova/        # meta.json + 占位（游戏本体 CR-2 迁入）
└── data/                  # editorial.json（Banner/精选/分类配置，空模板）
```

**偏离记录**：目录结构以 CD §10 为准，不采用 godot-bootstrap 通用模板（scenes/scripts/resources/autoload）——盒子工程按"壳层/单例/服务/游戏"组织，与三运行时 Runner 分发对齐。

## 2. project.godot

- `config/name="NOVA ARCADE"`，`run/main_scene="res://shell/main.tscn"`，features 4.5
- `[display]`：720×1560、`stretch/mode=canvas_items`、`stretch/aspect=expand`（override §2）
- `[rendering]`：`renderer/rendering_method="mobile"`（Q2=A Android 先行；桌面仅调试运行）
- `[autoload]` 注册顺序（被依赖者在前，EventBus 最先，QualitySettings 早于数据类单例）：

| # | 名称 | 路径 |
|---|------|------|
| 1 | EventBus | `res://core/event_bus.gd` |
| 2 | QualitySettings | `res://core/quality_settings.gd` |
| 3 | DB | `res://core/db.gd` |
| 4 | Registry | `res://core/registry.gd` |
| 5 | Nav | `res://core/nav.gd` |
| 6 | Sound | `res://core/sound.gd` |

**命名规则**（godot-engine §六）：autoload 脚本一律不声明 `class_name`，避免与注册名冲突。

## 3. Autoload 设计

### 3.1 EventBus
纯信号枢纽，无状态。信号表（CD §3.3 + 主题）：

| 信号 | 参数 |
|---|---|
| `game_launched` / `game_finished` | gid / (gid, result) |
| `trial_consumed` | gid, left |
| `payment_succeeded` | gid, order_id |
| `review_submitted` / `review_deleted` | gid |
| `achievement_unlocked` | def, points |
| `dlc_installed` | gid |
| `tab_changed` | idx |
| `settings_changed` | key, value |
| `theme_changed` | theme_name |
| `records_updated` | gid |

### 3.2 QualitySettings
复制 `projects/demo/demo_game/autoload/quality_settings.gd`（实查唯一参考实现；godot-engine.md §3.1 中 `projects/demo_game/` 为旧路径）后裁剪：保留档位检测（`OS.has_feature("mobile")` → mobile_high）、`resolve_texture()`、`get_scalar()`；Shell 几乎无纹理资产，`assets/textures/tier_desktop|tier_mobile/` 建空目录占位。**裁剪理由**：盒子 UI 为矢量/ColorRect 绘制，纹理分档在 CR-2 迁入游戏后按需启用（Risk-3）。

### 3.3 DB
全部本地持久化唯一出口。JSON 文件落 `user://db/`；两级写入（override §2）：
- **强写**（立即落盘）：`trial_used`、`orders`、`total_playtime/best/finish_count`
- **防抖写**（500ms 合并）：`search_history`、`daily`、`profile`

写级别冲突消解（Review B-1）：`save_profile(patch, force := false)`——默认防抖；主题切换等关键项传 `force=true` 立即落盘。

接口（CD §2.1）：`get_profile()/save_profile()`、`get_record(gid)/upsert_record(gid, patch)`、`list_records()`、`get_review(gid)/put_review(gid, dict)`、`list_orders()/put_order(dict)`、`add_search_history(q)/get_search_history()`、`get_achievements()/unlock(id)`、`get_daily()/touch_daily()`。

**损坏兜底**：JSON 解析失败 → 坏文件改名 `.corrupt` 备份 + 默认值重建 + 日志告警，不崩溃。
**退出 flush**：WILL_EXIT 时强制 flush 防抖缓冲（Review 低优项，RG-5 完整可靠）。

### 3.4 Registry
游戏目录：启动时扫描 `games/*/meta.json` + 读取 `data/editorial.json`；DLC meta 预留（`dlc_installed` 信号触发 `reload()`）。接口：`reload()`、`get(gid) -> GameMeta?`、`all() -> Array[GameMeta]`、`query({category, price, min_rating, tag, sort})`、`installed_version(gid)`、`ach_def(aid)`。

### 3.5 Nav
页面栈 + Tab：`push(page_path, data)`（220ms 滑入）、`pop()` / `pop_to_root()`、`switch_tab(idx)`（清栈回 Tab 根页，发 `tab_changed`）、`current() -> String`；返回键拦截：模态 Dialog 优先关闭 → 否则当前页 `on_back()` 返回 false 时默认 pop。

### 3.6 Sound
UI 音效/震动：`click()/toggle()/success()/error()/coin()`、`haptic(str)`；尊重 profile 开关（设置项缺省为开）。CR-1 用占位短音资源，M2 换正式音效。

## 4. 场景树与页面骨架

`Main.tscn`（CD §1.1）：

```
Main (Node)
├── BgLayer (Control)        # nebula 径向渐变 + GridLines(霓虹)/DotGrid(青瓷)，监听 theme_changed，400ms Tween
├── App (Control)
│   ├── PageStack (Control)  # 页面容器（当前页链）
│   ├── TabBar (Control)     # 底部四 Tab：首页/分类/搜索/我的
│   ├── OverlayLayer         # CR-1 空容器占位
│   └── ToastLayer           # 顶部滑入 Toast（300ms，展示 2200ms）
└── GameHost (Control)       # 游戏运行容器占位（CR-4 Launcher 接管）
```

**Page 基类**（`shell/pages/page.gd`，CD §3.2）：

```gdscript
class_name Page extends Control
func on_enter(data: Dictionary) -> void
func on_exit() -> void
func on_resume() -> void
func on_back() -> bool      # return false 走默认 pop
```

四根页骨架（home/category/search/library）：仅标题 + EmptyState 空态占位，`on_enter` 打印 gid/参数供冒烟验证。

## 5. 主题系统

- `shell/theme/theme_tokens.gd`：`TOKENS = {"neon": {...}, "elegant": {...}}`（token 值按 CD §8.1 表格），`static func color(key) -> Color`、`icon_grad(game_col)`、`current`；**主题色禁止硬编码**，非 Theme 控件一律 `ThemeTokens.color(key)`（override §2）
- `theme_neon.tres` / `theme_elegant.tres`：Panel/Label/Button/LineEdit 的 StyleBoxFlat 样式（圆角/边框/投影按 CD §8.5 映射）
- 切换流：`apply_theme(name)` → `ThemeTokens.current = name` → `root.theme` 换资源 → `DB.save_profile({"theme": name}, force=true)` → 发 `settings_changed` + `theme_changed` → BgLayer/页面重渲染；启动时读 profile 恢复
- **信号分工**（Review B-2）：BgLayer 只连 `theme_changed`（400ms Tween 仅此一处触发）；`settings_changed` 仅承载声音/震动等非主题设置，BgLayer 不监听
- **切换入口**：首页右上角 ThemeToggleButton（🎨 占位，CD §8.4），按下 `scale(.9)` + Sound.toggle()（Review 高优项补入）
- **禁用 hdr_2d/glow**（override §2，mobile 渲染器实心矩形 bug）：霓虹效果全部用分层半透明 ColorRect / StyleBoxFlat shadow 模拟

## 6. CoreManager 桩（Q3=A）

`services/core_manager.gd`：接口 `is_installed(version) -> bool`、`check_installed(version) -> bool`、`download(version, sha256)`、`remove(version)`；Mock 实现恒返回"未安装"并记录日志，**不阻塞任何流程**。真实下载/校验待 mame-godot-plugin v2 spike 验证后填（v2 已定稿，见 runtime-design §3.2；桩接口暂留至 M2）。

## 7. GameMeta 与内置 tetra_nova meta（Q4=A）

- `GameMeta` 字段按 CD §3.4（id/title/category/tags/price_model/price/trial/version/icon/scene/aliases/pinyin/achievements/desc/runtime/viewport_size/orientation）；`viewport_size`/`orientation` 省略时继承 Shell 720×1560 竖屏
- `games/tetra_nova/meta.json`：runtime="pck"、price_model="trial"（plays:3）、icon/scene 指向占位资源；游戏本体与 GameModule 适配在 CR-2 完成，CR-1 仅保证 Registry 可加载该 meta（页面展示在 CR-3 实现）

## 8. 不变行为清单（回归防护）

| # | WHEN | THEN 系统 SHALL |
|---|------|-----------------|
| RG-1 | 应用启动 | 进入 Home Tab，编辑器/运行无脚本错误 |
| RG-2 | 切换任一 Tab | 清栈回该 Tab 根页，发 `tab_changed`，无崩溃 |
| RG-3 | push/pop 页面 | 220ms 转场；返回键：Dialog 优先 → `on_back()` false 时默认 pop（CR-1 无 Dialog，Dialog 分支待 CR-3 补验） |
| RG-4 | 切换主题 | 所有颜色经 ThemeTokens，布局无跳变，选择持久化，重启后恢复 |
| RG-5 | 强写数据（trial_used/orders/playtime）变更后强杀进程（kill -9/断电）重启 | 数据不丢失；防抖数据此时允许丢失未落盘部分。**正常退出**时防抖缓冲必须已 flush，重启后完整恢复 |
| RG-6 | CoreManager Mock 查询 | 返回未安装，其余流程不受影响 |

## 9. 验收

- 代码级 POST-CHECK 9 项（hybrid §4.1）逐文件执行
- 集成冒烟（override §5，CR-1 适用子集）：① 工程编辑器打开无报错 ② 四 Tab 切换无崩溃 ⑤ 双主题切换无布局跳变；两种视口比例验证（Risk-1）
- RG-1~RG-6 逐条运行验证
