# 设计：CR-2 Launcher 完整生命周期 + TETRA NOVA 接入 GameModule

> 依据：已确认的 `requirements.md`（Q1~Q7）+ `design_plan.md`（D1~D5 全 A，2026-08-24）。
> 上游：client-design §3.1/§4.6、runtime-design GDScript 运行时节、override §2/§5、hybrid §3.2+§3.3。

## 1. 文件与目录变更

```
nova-arcade/
├── project.godot            # autoload +2（Launcher/TrialGuard，共 8）
├── core/
│   ├── game_module.gd       # GameModule 协议基类（class_name，非 autoload）
│   ├── db_io.gd             # 从 db.gd 拆出落盘原语（class_name，非 autoload）
│   └── db.gd                # 瘦身为接口层（≤200 行），委托 DBIO
├── services/
│   ├── launcher.gd          # Autoload：launch/quit/play_again 编排
│   ├── launcher_transition.gd  # 转场状态机+视口切换（launcher 超限时拆出）
│   └── trial_guard.gd       # Autoload：left/consume
├── shell/components/
│   ├── transition_overlay.tscn/.gd   # 全屏遮罩 + fade（Launcher 专用）
│   ├── result_overlay.tscn/.gd       # 结算卡（挂 OverlayLayer）
│   └── purchase_dialog.tscn/.gd      # Mock：标题+价格+按钮，M4 换内容
├── shell/pages/home.gd/.tscn         # 临时"▶ TETRA NOVA"按钮（CR-3 移除）
├── games/tetra_nova/
│   ├── src/                 # 源工程 scenes/scripts 整拷贝（.uid/.import 随迁）
│   ├── module.tscn          # 实例化 src/Main.tscn + module_adapter.gd
│   ├── module_adapter.gd    # GameModule 协议实现（≤200 行）
│   └── meta.json            # scene → res://games/tetra_nova/module.tscn
└── tools/test_launch.tscn/.gd        # headless 全流程驱动
```

## 2. GameModule 协议基类（core/game_module.gd）

```gdscript
class_name GameModule extends Node
signal quit_requested(result: Dictionary)   # {score, playtime, achievements, extra}
signal achievement_unlocked(aid: String)
func boot(ctx: Dictionary) -> void          # ctx={save_dir,trial_mode,owned,best,viewport_size}
func pause_game() -> void
func resume_game() -> void
```

所有游戏场景脚本继承它；Launcher 只依赖此基类，不感知具体游戏（hybrid §3.3 约束）。

> `achievement_unlocked` 映射：module 信号带 `aid`，Shell 侧经 `Registry.ach_def(aid)` 映射到 EventBus.achievement_unlocked(def, points)（M2 Engine 接入时生效；本 CR achievements 恒空数组）。

## 3. LauncherService（services/launcher.gd）

状态机：`idle → launching → running → quitting`；launching 失败（load/boot 异常）→ 回退 idle + Toast 报错，视口/遮罩复原。

**launch(gid) 时序**（FR-1~3，§4.6 步骤 1~9）：
1. `Registry.lookup(gid)` 为 null → Toast 报错返回
2. DLC 检查桩：跳过（M5）
3. 权益：`price_model=free`/owned → 通过；`trial` → `TrialGuard.left(gid)==0` → `PurchaseDialog.show_mock(meta.price)` + Toast"¥X 解锁完整版（M4 接支付）"→ 中止；paid 未拥有同理；trial 且 left>0 → `TrialGuard.consume(gid)`（强写，boot 前）
4. `DB.get_record(gid)` 读现值 → sessions+1 → `DB.upsert_record(gid, {last_played: now, sessions: n+1})`（upsert 为浅合并，禁止字面 +1）
5. `TransitionOverlay.fade_out(300ms)` → **alpha==1 后**（显式 await，Risk-1）：调 `meta.get_viewport_size()`/`get_orientation()`（GameMeta 内置缺省 720×1560/portrait，不裸读字段）→ `get_viewport().content_scale_*` + `DisplayServer.screen_set_orientation` → `GameHost.visible = true`
6. runtime 分发：`pck/html` → `load(meta.scene).instantiate()` 入 GameHost；`arcade` → CoreManager 桩路径（不实测）
7. `module.boot(ctx)`，ctx.save_dir=`"user://saves/{gid}/"`（通用规则，Launcher 不感知具体游戏；如 tetra_nova → `user://saves/tetra_nova/`）；发 `game_launched(gid)`
8. `TransitionOverlay.fade_in(300ms)`

**quit 时序**（FR-4，§4.6 步骤 10~16）：监听 `module.quit_requested(result)` → 冻结 module（process_mode=DISABLED）→ fade_out 遮罩 → **alpha==1 后**恢复 Shell 默认视口（当前 720×1560 portrait，常量集中在 launcher.gd；M2 横屏游戏不受影响）+ orientation portrait → queue_free + GameHost.visible=false → 读现值 finish_count+1 → `DB.upsert_record` 强写 {total_playtime, best, finish_count: n+1} → 发 `game_finished(gid, result)` → fade_in → 400ms 后 `ResultOverlay.show(result)`。

**play_again(gid)**（FR-5，§4.6 步骤 15）：不重建 module——`module.reset_run()`（adapter 重启一局）+ 重走步骤 5~8；**不消耗 trial**（重玩免费，仅 session+1 读现值递增），不重走权益检查。

**"换一个"**：`Registry.similar(gid)`（本 CR 新增：按同 category+tag 匹配，无则空数组）为空 → 整区隐藏（Q2=A）；非空时迷你卡区横列，卡片点击 CR-2 内弹 Toast"详情页即将上线"（回退策略，client-design §4.6 步骤14 openDetail 归 CR-3）。

## 4. TrialGuard（services/trial_guard.gd）

- `left(gid) -> int`：meta.trial.plays − record.trial_used（record 缺省 0）
- `consume(gid)`：trial_used+1 **强写** + 发 `trial_consumed(gid, left)`；owned 恒 true
- 消耗时机：launch 通过权益检查后、boot 前

## 5. TransitionOverlay / ResultOverlay / PurchaseDialog

- **TransitionOverlay**：全屏 ColorRect（黑，alpha 0→1 Tween 300ms）；`fade_out() -> await Signal`、`fade_in()`；遮罩期间拦截输入。视口切换只允许在 `await fade_out()` 返回后执行（override §2 铁律，Risk-1）
- **ResultOverlay**（挂 OverlayLayer，ThemeTokens 全字段）：游戏名 / 本局分数·时长 / 历史最高；新解锁成就逐条 Toast（`DB.unlock` + `achievement_unlocked`，Engine 判定留 M2）；按钮：[▶ 再来一局]→`Launcher.play_again`；[✎ 评价*]→finish_count≥2 且本次 playtime≥600s 亮，否则置灰"还需 X 分钟"（X=`ceil((600 − playtime) / 60)`，playtime 单位秒，≤0 点亮；评价提交链路 CR-3/M1 后段）；[返回]→Nav 回首页
- **PurchaseDialog**（Mock）：游戏名 + ¥price + [暂不购买]；M4 接支付只换按钮行为

## 6. tetra_nova 迁入适配（D1=A，最小 diff）

- `src/` 整拷贝源工程（含 .uid/.import）；dev 脚本处理：test_runner/probe3/probe4 无文件写入保留（基线重跑用），`shot_driver.gd`（写 `user://shot_*.png`）改读 SAVE_PATH（≤5 行）或排除迁出——保证 T3 `user://` grep 验收通过
- 3 处存档点改道：
  - main.gd `_save_best`、ui.gd `on_run_over`/`_refresh_best`：字面量 `"user://save.cfg"` → 读节点变量 `SAVE_PATH`（默认值不变，adapter boot 时注入 `ctx.save_dir + "save.cfg"`）——每处改动 ≤5 行
  - game.gd `stage.log` dev 日志删除
- **module_adapter.gd**（继承 GameModule）：
  - `boot(ctx)`：记录 ctx；注入 SAVE_PATH；启动计时（Time.get_ticks_msec）
  - 监听 src `game.run_over(stats)` → 结算屏"回菜单"按钮（D5=A，玩家主动唯一退出路径）→ emit `quit_requested({score, playtime: 秒, achievements: [], extra: {wave, lines, max_combo, bosses}})`
  - `pause_game()/resume_game()`：转发 src 暂停机制（源无 pause 则补最小 set_paused：process_mode + 遮罩，≤30 行）
  - `reset_run()`：重启一局（src 重开局入口），供 play_again
- meta.json：`scene → res://games/tetra_nova/module.tscn`，其余字段不动

## 7. db_io.gd 拆分（CR-1 遗留收口）

- 迁出至 `core/db_io.gd`：JSON 读/写原语、`.corrupt` 兜底、防抖 flush 调度、目录初始化、`flush_all()` 原语；DB 接口签名**全部不变**；`_notification`/`_exit_tree` 生命周期钩子（正常退出 flush，RG-5 双相）留 DB 层，调 DBIO.flush_all()
- RG-5 双相（强杀/正常退出）+ test_db 三遍全量回归

## 8. 不变行为清单（回归防护）

| # | WHEN | THEN 系统 SHALL |
|---|------|-----------------|
| RG-1~6 | （沿用 CR-1 design.md §8 全部条目） | 不破坏：启动/Tab/push-pop/主题强写/CoreManager Mock |
| RG-7 | launch 后强杀进程重启 | trial_used/playtime/best/finish_count 不丢（强写生效）；save.cfg 在 save_dir 下完整 |
| RG-8 | 未 launch 时打开应用 | 首页/四 Tab/主题切换行为与 CR-1 完全一致（临时按钮除外） |

## 9. 决策记录

- D3=A：QualitySettings resolve_texture/get_scalar **不恢复**，挂账首个有纹理的 CR
- Mock 桩位：PurchaseDialog（M4）、DLC 检查（M5）、AchievementEngine 判定（M2）、arcade runtime（CoreManager v2 spike 后）
- 范围外声明（client-design §4.6 对应步骤，防执行时漏做/误加）：首页"继续游戏"区块重渲染（步骤13）归 CR-3；Analytics.track 上报（步骤9）归 M3，本地以 EventBus.game_launched/game_finished 为数据源

## 修订记录

- 2026-08-24 初版（Q1~Q7 + D1~D5 确认）
- review R1~R5 修订：upsert 递增口径、Registry.lookup/similar、consume 入时序、play_again trial 语义、dev 脚本 grep 策略、Shell 默认视口、launching 失败分支、achievement 映射、X 分钟公式、db_io 钩子归属、save_dir 通用规则、GameMeta 访问器、"换一个"任务归属与点击回退、headless quit 触发、left() 非 trial 行为、范围外声明（继续游戏区块/Analytics）

## 10. 验收

- 代码级 POST-CHECK 9 项（hybrid §4.1）逐文件；GameModule 协议合规七项（hybrid §4.2，含 save_dir grep 验证）
- headless：`tools/test_launch.tscn` 驱动 launch→boot→quit 全流程断言（ctx 完整 / 视口切换在遮满后 / DB 强写三字段 / trial 计数 / play_again / 耗尽中止）
- 迁入回归：源工程 test_runner/probe3/probe4 在 `games/tetra_nova/src/` 下重跑全绿
- GUI 双轨：启动画面/游戏运行/结算卡截图 read_image（neon/elegant × 720×1560，断言 Shell 侧元素无溢出）；RG-1~8 + test_smoke 回归
