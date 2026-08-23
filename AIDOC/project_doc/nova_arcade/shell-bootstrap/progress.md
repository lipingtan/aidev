# NOVA ARCADE CR-1 shell-bootstrap 执行进度（暂停检查点）

## 状态：T8 已完成（test_main 13/13 全过）；T10 集成冒烟待开始

## 已完成任务（tasks.md 同步 ✅🎯）
| 任务 | 验证 | 证据 |
|---|---|---|
| T1 工程初始化 | smoke_boot exit 0 | project.godot（720×1560 canvas_items/expand，renderer=mobile，6 autoload） |
| T2 EventBus+QualitySettings | test_tier | event_bus.gd（12 信号）、quality_settings.gd（detect_tier() 公开，裁剪方法待 CR-2 恢复） |
| T5 Nav+Page | test_nav 13 断言 exit 0 | core/nav.gd、shell/pages/page.gd（@abstract 引擎限制→运行时 push_warning）、tools/dummy_page_a/b |
| T9 CoreManager 桩 | test_core_manager | services/core_manager.gd |
| T4 Registry+GameMeta+meta | test_registry 12 断言 + 坏 meta 容错 exit 0 | core/game_meta.gd、core/registry.gd（get→lookup）、games/tetra_nova/{meta.json,icon.png,main.tscn}、data/editorial.json |
| T3 DB | test_db 三遍（normal/verify/corrupt-check）全过 exit 0 | core/db.gd（强写/防抖/.corrupt 兜底/_exit_tree flush）、tools/test_db |
| T6 Sound | test_sound 8 断言 exit 0 | core/sound.gd、assets/sfx/*.wav×5（gen_sfx.gd 生成）、tools/gen_sfx.gd |
| T7 主题系统 | test_theme 11 断言 exit 0 + grep 验收过 | shell/theme/{theme_tokens.gd,theme_neon.tres,theme_elegant.tres}、shell/{bg_grid.gd,bglayer.gd,bglayer.tscn}；视觉项移交 T10 |
| T8 四Tab骨架 + Main.tscn | test_main 13/13 ✅ | shell/main.tscn + pages/{home,category,search,library} + tab_bar/toast_layer；**Toast 根因修复**：`tween_interval` 后重复 `set_parallel(true)` 压平停留 → 已去掉；`position:y` 在锚定 FILL 控制体干扰时间线（滑出改串行）；两种视口溢出项移交 T10（Risk-1） |

## T8（进行中）四 Tab 骨架 + Main.tscn
已创建：
- shell/pages/{home,category,search,library}.gd/.tscn（标题+EmptyState；home 含 🎨 ThemeButton：scale(.9)+Sound.toggle()+apply_theme 切换）
- shell/components/tab_bar.gd（Btn0..3 → Nav.switch_tab，选中态 token accent/ink2）、shell/components/toast_layer.gd（show_msg：300ms 滑入/2200ms/滑出，set_parallel）
- shell/main.gd（_ready：Nav.locate_page_stack → ThemeTokens.restore → Nav.switch_tab(0)）+ shell/main.tscn 重写（BgLayer/App{PageStack,OverlayLayer,ToastLayer,TabBar}/GameHost）
- tools/test_main.{gd,tscn}：13/13 全过（启动 Home、四 Tab 切换、Toast 滑入+消失时序、🎨 按钮存在/主题切换/音效/profile 写）

**失败项（唯一）**：`Toast 滑入展示`——t=0.4s 时 Label.modulate.a > 0.9 不成立（但 t=3.2s 的"滑出消失"反而通过）。

**诊断已完成（根因定位）**：
- 探针 `tools/_probe_toast.tscn`（Node 版，`--scene` 运行；SceneTree 版 `-s` 会 `await process_frame` 死锁）实测原版 `toast_layer.gd`：
  - t=0.2 a=0.651 / t=0.4 a=0.615 / t=0.6 a=0.0、pos.y 24→-80 —— **slide-out（a 1→0、pos 24→-80）从 t=0 直接运行，slide-in(0→1) 与 2.2s 停留被压平**。
- 对照实验 v1（**纯 alpha 顺序**：`tween_property a 0→1 .from(0.0)` → `tween_interval(2.2)` → `tween_property a 1→0`，无 set_parallel、无 position）：探针显示 0.2s a=0.65、0.4s a=1.0、2.4s 前恒为 1.0、2.6s a=0.58、2.8s a=0.0 —— **完全正确**。
- **结论**：失败非 alpha 本身，而是原版的 `set_parallel(true/false)` 切换 + `tween_interval` 组合把 2.2s 停留压平；且 `position:y` 参与时时间线错乱。

**最终修复**：`toast_layer.gd` 定稿——`set_parallel(true)` 并行滑入（position+alpha 300ms ease-out）→ `set_parallel(false)` + `tween_interval(2.2)` 停留 → 顺序滑出（ease-in）；**关键**：`tween_interval` 之后不再重复 `set_parallel(true)`（原版压平停留的根因）。
**验收**：test_main 13/13 全过（含原失败项「Toast 滑入展示」）；临时探针 `_probe_toast.{gd,tscn}` 已删除。
**视口溢出**（两种视口比例无错位）为视觉项，headless 不可验，移交 T10（Risk-1）。

## T10（未开始）集成冒烟
- RG-1~RG-6 回归 + T5 栈复验 + **视觉验证**（GUI 截图/read_image：双主题切换无布局跳变、双视口比例 9:19.5 与 9:20 无溢出 Risk-1、重启恢复上次主题 RG-4）+ smoke_report.md
- 需 `pnpm run dev:web` 类流程不适用——Godot GUI 截图走 DSH Web GUI 或 Godot editor 截图（T10 时定方案）

## 环境铁律
- Godot：`C:\data\developer\devtool\godot\godot4.5\Godot_v4.5-stable_win64_console.exe`
- 每次 headless 前 `$env:APPDATA` 指临时目录（user:// 隔离）；stderr "Failed to read the root certificate store" 无害
- 验证一律测试场景（.tscn + quit(0/1)），不用 -s 读 autoload 状态；新增 class_name 先 `--import`
- 级联解析错误时用 `-s res://<file>` 直接运行逼出根错误

## Godot 4.5 实测经验（累计 20 条，新子代理/后续 CR 必读）
1. `-s` 脚本模式：Autoload `_ready()` 晚于脚本 `_initialize()` → 验证用测试场景
2. 多方法类中无函数体的 `@abstract func` 不可解析（根错误被级联掩盖）→ 默认实现 + 运行时 push_warning（T5 Page.on_enter）
3. 新增 class_name 文件需一次 `--headless --import` 重建全局类缓存
4. headless 必须重定向 `$env:APPDATA`（曾污染工程目录 Godot/app_userdata/，已删）
5. stderr TLS 证书告警无害
6. 级联解析错误：`-s res://<file>` 直接运行逼出根错误
7. **Node 子类不能定义 `func get(...)`**（与 Object.get(StringName) 签名冲突 + warning-as-error）→ Registry 用 lookup()；同理 `show(msg)` 与 CanvasItem.show() 冲突 → toast 用 show_msg()
8. **`push_info()` 不存在**（仅 push_error/push_warning）
9. **String.natural_compare 此构建不可用** → to_lower() 字典序
10. **单行 lambda 不支持**（`func(x: T) -> bool: return ...` 一行 → 解析错）→ 全部改显式辅助函数 + `Callable.bind()`
11. **可空返回注解 `Type?` 不可解析；非空类型注解 + return null = 编译错误** → 用 `-> Variant`，调用方显式 `var m: GameMeta = ...`
12. **Gradient 无 set_point_color/point_count/points 属性**（此构建仅 `offsets`(PackedFloat32Array)+`colors`）；add_point 是插入语义且默认自带 2 点 → 直接赋值 colors/offsets（用 get_property_list 探测确认）
13. **无内置 haptic API**（DisplayServer.haptic_pulse 不存在）→ Sound.haptic 仅 mobile 能力保护 + _haptic_impl 实现位（M2 接插件）
14. **headless 下 NOTIFICATION_WM_CLOSE_REQUEST 不触发** → DB._exit_tree() flush 兜底
15. **Variant 推断 `:=` 遇未类型化调用 = warning-as-error** → 全显式类型标注
16. **`Tween.parallel()` 返回独立 Tween 对象**（无引用即被释放，分支动画丢失）→ 用 `set_parallel(true/false)`
17. PowerShell 生成二进制资产不可靠（BinaryWriter 方法解析失败、-band 对数组报错）→ 用 Godot `-s` SceneTree 脚本生成（tools/gen_sfx.gd 模式：PackedByteArray 手工拼 WAV 头）
18. 子代理（background subagent）本会话两度批量中断且未写任何文件 → 批次3/4 起全部改主会话直接执行
19. **`Tween.set_parallel` + `tween_interval`：`tween_interval()` 之后再 `set_parallel(true)` 会把停留压平**（slide-in 后直接 slide-out，2.2s 间隔丢失）→ 修复=滑入用 `set_parallel(true)` 建并行对、`set_parallel(false)` 后 `tween_interval` 停留、之后**不再 toggle**；`position:y` 在锚定 FILL Control 上 tween 本身**无碍**（v2/v5 探针验证时间线正确，无需改用 offset/纯 alpha）
20. **此构建 `Tween` 无 `tween_parallel()`、`Tween.create_tween()`（嵌套）不可用**；`create_tween()`/`tween_parallel()` 返回 Variant → `var tw: Tween = ...` 须显式类型标注，否则 `:=` 推断报「Cannot infer type」编译错误（经验 #15 的延伸）

## 偏差登记（tasks.md Constraints 已同步）
- T5：on_enter 运行时 push_warning 替代 @abstract（引擎限制）
- T4：get→lookup；无单行 lambda / Type? / natural_compare
- T3：_exit_tree 补 WILL_EXIT flush（headless 不触发 WM_CLOSE_REQUEST）
- T6：haptic 仅能力保护（M2 接插件）；音效 gen_sfx.gd 生成
- T7：Gradient 用 offsets+colors 构造；restore() 发 theme_changed（BgLayer _ready 先于 Main._ready，需借此重渲染）
- T8：show→show_msg
