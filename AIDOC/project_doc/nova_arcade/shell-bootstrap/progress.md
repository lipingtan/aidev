# NOVA ARCADE CR-1 shell-bootstrap 执行进度（暂停检查点）

## 状态：**CR-1 完全闭环**——T1~T10 全 ✅ + §5.1 编辑器打开零报错（--editor --quit 实测，顺删废弃 _shot.gd）+ §5.5 视觉 GUI 实测（含 Fix-1）+ **三角色终版验收通过**（review_report.md；唯一低优：db.gd 232 行超限 → CR-2 拆 db_io.gd）。§5.3/§5.4 属 CR-2/CR-3 范围

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

## T10 集成冒烟 ✅ 🎯（2026-08-21，Godot 4.5 headless）
- **test_smoke 12/12 全过**（exit 0）：§5.2 四 Tab、RG-1 启动/Home、RG-2 Tab 切换、RG-3 push/pop、RG-4 主题切换+profile force 写、RG-6 CoreManager Mock
- **test_rg5 双相 6/6 全过**（exit 0）：强写（trial_used/orders/playtime/best）+ 防抖 flush（search_history）+ profile 重启不丢——RG-5 通过
- **挂账（非缺陷，待 Services/真机）**：§5.1 需 Godot 编辑器；§5.3 详情页 CTA（PayService/TrialGuard 未实现→CR-2/CR-3）；§5.4 Launcher 生命周期（LauncherService 未实现→CR-2）；§5.5 双主题无跳变/双视口溢出（Risk-1 需 GUI 截图，headless 不可验→M1/M2 发布门控）
- 交付：`smoke_report.md`（测试脚本 `test_smoke.{tscn,gd}`、`test_rg5.{tscn,gd}` 已入 tools/）

## §5.5 视觉验证 + Fix-1 ✅（2026-08-24，GUI 实测）
- **工具**：`tools/_shot2.{gd,tscn}`（args `--theme=` `--restore` `--scale=WxH` `--shot_out=`；打印 [vis]/[diag]/[rect]/[overflow]/[restore]；25s 兜底 quit 防挂死）+ 批量驱动 `dev/tmp_vis/vis_run.ps1`（GUI 版 exe，每次 APPDATA 隔离 + 杀孤儿 godot 进程）
- **矩阵全绿**：A neon/elegant ×720×1560、B neon/elegant ×720×1600 —— 溢出检查全过；关键节点 rect（App/TabBar/Home/Title/ThemeButton）跨主题同尺寸完全一致 → **无布局跳变**；C1 写 elegant + C2 `--restore` → `[restore] PASS`
- **截图证据**：`dev/tmp_vis/vis_{neon,elegant}_{1560,1600}.png`（read_image 核验：青瓷=米白底+墨字+豆绿星云+点阵）
- **Fix-1（双主题未生效，真实缺陷×2）**：① root Window 的 theme 不向子 Control 传播（4.5 实测，预设也无效）→ `theme_tokens.gd` 新增 `theme_target`，Theme 资源落到 `Main/$App`；② `bglayer.tscn` Bg 底色硬编码霓虹值且 `_apply()` 从不更新 → 按 bg token 更新（400ms Tween）。文档：`Fix-1-dual-theme-not-applied/bugfix.md`。修复后复测全绿
- 结论：§5.5 挂账**关闭**；CR-1 遗留挂账仅剩 §5.1 编辑器打开（人工）、§5.3/§5.4（Services 层 → CR-2/CR-3）

## §5.1 编辑器验证 + 三角色终版验收 ✅（2026-08-24）
- **§5.1**：GUI exe `--editor --quit` 三次实测。首跑发现 `_shot.gd:22` 解析错误（Window 无 find_node，4.5 怪癖；该工具已被 _shot2 取代）→ 删除 `_shot.{gd,tscn}` + 旧 `tmp_smoke/shot_run.ps1`，复跑**零脚本/资源错误**（剩余 stderr：TLS 证书 + resthumb 预览缓存，均为沙箱环境产物）。§5.1 挂账关闭
- **三角色终版验收**（review_report.md 追加节，打开实际代码逐项验证）：业务专家 FR/非功能/RG 证据链 ✅；产品经理 ThemeButton 入口/空态/范围一致 ✅；架构师 依赖方向（页面互不引用、场景引用仅 nav.gd L36-39）/meta 字段/db 双保险 flush/EventBus 12 信号 ✅。⚠️ 低优 1 项：db.gd 232 行 > 200（T3 约束未执行拆分）→ **CR-2 动 DB 时拆 db_io.gd**
- **结论：✅ 三角色验收通过，CR-1 shell-bootstrap 完全闭环**

## 环境铁律
- Godot：`C:\data\developer\devtool\godot\godot4.7\Godot_v4.7-stable_win64_console.exe`（headless）；GUI 渲染用 `Godot_v4.7-stable_win64.exe`（Vulkan 可用，RTX 5090），每次运行前杀孤儿 godot 进程（脚本错误在 quit() 前会留活进程）
- 每次 headless 前 `$env:APPDATA` 指临时目录（user:// 隔离）；stderr "Failed to read the root certificate store" 无害
- 验证一律测试场景（.tscn + quit(0/1)），不用 -s 读 autoload 状态；新增 class_name 先 `--import`
- 级联解析错误时用 `-s res://<file>` 直接运行逼出根错误

## Godot 4.5 实测经验（累计 21 条，新子代理/后续 CR 必读）
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
21. **root Window 的 `theme` 属性不向子 Control 传播**（GUI 实测：root.theme=青瓷后子 Label `get_theme()==null`、颜色落内置白默认；场景加载前预设同样无效）→ Theme 资源须同时赋给 Control 容器（本项目 `ThemeTokens.theme_target = Main/$App`，见 Fix-1）；诊断手段：GUI 运行打印节点 `get_theme()`/`font_color` 解析值

## 偏差登记（tasks.md Constraints 已同步）
- T5：on_enter 运行时 push_warning 替代 @abstract（引擎限制）
- T4：get→lookup；无单行 lambda / Type? / natural_compare
- T3：_exit_tree 补 WILL_EXIT flush（headless 不触发 WM_CLOSE_REQUEST）
- T6：haptic 仅能力保护（M2 接插件）；音效 gen_sfx.gd 生成
- T7：Gradient 用 offsets+colors 构造；restore() 发 theme_changed（BgLayer _ready 先于 Main._ready，需借此重渲染）
- T8：show→show_msg

## 三角色 review 落实纪要（全部修复：高1/中4/低4，共9项）

| # | 类型 | 落地方式 |
|---|---|---|
| A-2 | 架构 | Services 层（Launcher/TrialGuard/PayService/Recommender/AchievementEngine/Searcher/Analytics）未实现，当前代码仅覆盖核心单例+页面骨架；design §4.6/§6/§7 需**独立 CR** 补码后再做设计级 review 的代码验证 |
| P-2 | 产品 | 双主题视觉验收设为 M1/M2 **发布门控**，不得整体后移到 T10 之后（T10 冒烟前必须完成） |
| B-1 | 业务 | runtime-design §3.6 已加 ROM 合规加固（授权/公有域白名单 + 自导入隔离 + 版权尽调）✅ |
| P-1 | 产品 | client-design §4.6 评价引导改为"按被评价游戏、累计达门槛" ✅ |
| P-3 | 产品 | client-design §4.4 搜索历史统一 top-20（与 HISTORY_CAP 一致）✅ |
| P-4 | 产品 | client-design §4.6 Arcade 再来一局标注依赖未实现组件（M2 真机前置）✅ |
| A-4 | 架构 | client-design §5 注明 DB 采用 `user://db/` 合并布局（与实现一致）✅ |
| A-1 | 代码 | main.gd 主入口恢复屏幕方向到竖屏（崩溃/强杀兜底）✅ |
| A-3 | 代码 | db.gd 防抖 Timer 到期 queue_free，消除累积 ✅ |
| A-6 | 架构 | ✅ 已关闭（实测）：`root.content_scale_size` 运行时切换在 4.5 headless 生效（tools/test_viewport.gd：visible rect 1560×1560→1920×1920→恢复，expand 公式精确吻合），client §4.6 参考实现成立；`screen_set_orientation` 真机行为归 M2 |

## 挂账项闭环记录（CR-1 后文档收口）

| 挂账 | 状态 | 处理 |
|---|---|---|
| A-6 content_scale_size 语义存疑 | ✅ 已关闭 | 4.5 headless 实测生效（tools/test_viewport.gd），client §4.6 加注记 |
| Registry.get→lookup 偏差未回写设计 | ✅ 已关闭 | client §2.1/§4.6 改为 `lookup(gid)` + 偏差注记（与 Object.get 签名冲突） |
| Arcade 运行时 myosd 直嵌 vs v2 双态 | ✅ 已收口（架构层） | runtime-design 按 v2 重写（§3.2 MameRuntime 插件/§3.3 生命周期/§4 对照表），myosd 直嵌降级为附录 A；client §2.2/§2.4/§4.6/§5、server §4.4、总览 §0 同步 |
| 核心包体积 80~150MB vs 74MB | ✅ 已关闭 | 统一为 v2：libMAME4droid.so ~74MB 按 ABI + jni .so 0.1MB 随包 |

**遗留（非文档问题，归 M2）**：v2 spike 验证 DESIGN.md §6 风险项（插件打包钩子/manifest 合并/JNI 包名绑定/.so 下载）+ Arcade 真机帧率/音频/进程语义。
