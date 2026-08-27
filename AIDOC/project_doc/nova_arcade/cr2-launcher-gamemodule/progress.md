# CR-2 LauncherService + GameModule 协议 — 执行进度

> 门控：需求/设计/任务三文档已于 2026-08-24 用户确认（含 review R1~R5），直接进入任务执行。
> 执行顺序（依赖）：T1 → T2 → T5 → T4 → T3 → T6 → T7 → T8 → T9

## 环境速查（继承 CR-1 progress.md）

- headless：`C:\data\developer\devtool\godot\godot4.5\Godot_v4.5-stable_win64_console.exe --headless --path projects\nova_arcade\nova-arcade --scene res://tools/xxx.tscn`
- GUI：`Godot_v4.5-stable_win64.exe`（Vulkan，RTX 5090）；截图驱动 `dev/tmp_vis/vis_run.ps1` + `tools/_shot2.*` 模式
- headless 前：`$env:APPDATA` 指向临时目录隔离 + 杀孤儿 godot 进程
- 新增 class_name 文件需先跑一次 `--headless --import` 刷新全局类缓存

## 执行记录

### T1 db_io.gd 拆分（CR-1 遗留）✅
- [x] core/db_io.gd（class_name DBIO extends RefCounted：读写原语/.corrupt 回退/目录初始化/debounce 调度）
- [x] core/db.gd 瘦身 ≤200 行，接口签名零变化；_notification/_exit_tree 钩子留 DB 层调 DBIO.flush_all()
- [x] test_db 三遍 + RG-5 回归（write/verify 全绿）
- 经验：4.5 SceneTreeTimer（RefCounted）无 queue_free，抖窗到期直接落盘即可；RG-5 双相须用 `--scene res://tools/test_rg5.tscn --rg5-phase=write|verify`（`-s .gd` 不载 autoload）

### T2 GameModule 协议基类 ✅
- [x] core/game_module.gd（33 行；协议七项 1~4 + POST-CHECK 全过，--import 零报错）

### T5 TrialGuard ✅
- [x] services/trial_guard.gd（37 行；left 非 trial 返回 -1；consume 强写 + trial_consumed(gid,left)）

### T4 LauncherService + 转场 ✅
- [x] services/launcher.gd（200 行整，超限拆出 launcher_transition.gd 53 行 / launcher_util.gd 35 行；后两者 extends Node 由 Launcher add_child.call_deferred——autoload 就绪期 root busy，同步 add_child 必失败）
- [x] shell/components/transition_overlay.tscn/.gd（300ms 淡出/淡入；vp_at_full_mask 时序钩子 T8 断言通过）
- [x] core/registry.gd +similar(gid)（182 行；同类目+标签交集排序，空数组）
- [x] project.godot autoload 共 8（TrialGuard 第 7、Launcher 第 8）
- 经验：4.5 屏幕方向枚举改名 `DisplayServer.ScreenOrientation.SCREEN_PORTRAIT/SCREEN_LANDSCAPE`（旧 SCREEN_ORIENTATION_* 解析报错；批量探针法：多候选名同脚本，解析器逐一报缺失项）

### T3 tetra_nova 迁入 + 适配层 ✅
- [x] games/tetra_nova/src/ 27 文件（排除 shot.gd/shot_driver.gd/probe*.uid 孤儿引用）
- [x] res://scripts|scenes 旧前缀零残留（grep 验证）；SAVE_PATH 变量化 3 存档点 + game.gd 删 stage.log
- [x] module.tscn + module_adapter.gd（85 行；只读 _ui._over 追加「↩ 回菜单」，src 零 diff）
- [x] meta.json scene → module.tscn；基线 test_runner/probe3/probe4 在 src/ 下重跑全绿 exit 0

### T6 ResultOverlay + PurchaseDialog ✅
- [x] result_overlay.gd（145 行；按钮连接 _ready 一次性建立防重复 connect，show_card 可重复；评价灰态按解读注记2）
- [x] purchase_dialog.gd（33 行 Mock：¥price+[暂不购买]）

### T7 首页临时启动按钮 ✅
- [x] home.tscn PlayButton「▶ TETRA NOVA」+ home.gd → Launcher.launch("tetra_nova")（注释标注 CR-3 移除）

### T8 headless test_launch ✅
- [x] tools/test_launch.tscn/.gd（137 行；22 断言 ALL PASS exit 0；坏 meta 步 headless res:// 只读时优雅 SKIP）
- 排障：DB.get_record 返回**内存字典引用**（后续 upsert 原地更新）→ 测试 _rec() 必须 duplicate() 快照，否则「前值==后值」误判 FAIL；headless 下 boot→quit <1s → playtime(int 秒)=0 → launch 后等待放宽到 2.0s

### T9 GUI 双轨验收 ✅
- [x] neon/elegant × home/run/result 截图 + read_image 断言（GUI exe 需隔离 APPDATA，否则默认 %APPDATA% 不可写 → signal 11 崩溃）
- [x] 全量回归（test_launch 25 断言 / test_smoke / RG-1~8 / 源基线 ×3 复跑）
- [x] acceptance_report.md（截图路径 + RG 表 + 自测矩阵 + 解读注记 + 遗留项）
- 视觉检查发现并修复 3 缺陷：① RUNNING 时 Shell App 未隐藏 → LauncherUtil.set_game_ui（GameHost 与 App 显隐相反）；② OverlayLayer 恒 visible=false 结算卡不渲染 → show_card 开父层、关闭路径关回；③ tetra_nova ui 是 CanvasLayer 不随祖先 visible 隐藏，HUD 泄漏到结算画面 → adapter._set_ui_visible（quit 隐 / boot、reset_run 显）

## 设计解释记录（执行期决策，写入验收报告）

1. quit 序列：模块 freeze（process_mode=DISABLED）后**保留节点不立即 free**，`release()`（结算卡「返回」）才 queue_free —— 满足 play_again「不重建模块、reset_run 复用」（design §3 两处约束的唯一自洽解释）。
2. 评价按钮灰态文案：playtime<600s → 「还需 X 分钟」；playtime≥600s 但 finish_count<2 → 「还需 N 局」。
3. 转场遮罩/结算卡/购买弹窗由 Launcher._ready 运行时实例化（遮罩挂 root，卡片挂 OverlayLayer，无则 fallback root）——不改 main.tscn，保 RG-8。
