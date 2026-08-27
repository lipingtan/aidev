# 任务：CR-2 Launcher 完整生命周期 + TETRA NOVA 接入 GameModule

> 依据 `design.md`（2026-08-24 确认 + review R1~R5 修订，见 design 修订记录）。格式：三要素（task-representation.md）；Shell CR + GameModule CR 通用 Constraints（hybrid §3.2/§3.3）+ override §2 逐任务隐含生效，不重复列出。
> 状态标记：⬜ 未开始 / 🔨 进行中 / ✅ 完成；验收 📋 待验 / 🎯 通过

## 进度摘要

- 总任务数：9　已完成：9　进行中：0　待开始：0（100%）
- 当前阶段：全部完成，CR-2 验收通过（acceptance_report.md 2026-08-25）

## 依赖关系

```
T1(db_io拆分) ──────────────┐(独立，任意时点)
T2(GameModule基类) → T3(tetra_nova迁入适配) ─┐
T5(TrialGuard) → T4(Launcher+转场) ──────────┤
T6(ResultOverlay/PurchaseDialog, 依赖T4) ────┼→ T8(headless test_launch) → T9(GUI双轨+全量回归)
T7(首页临时按钮, 依赖T4) ─────────────────────┘
```

- **关键路径**：T5 → T4 → {T6/T7 并行} → T8 → T9（T2→T3 与之并行，T1 独立）
- 补缺失边：T3 依赖 T2（adapter 继承基类）；T6/T7 依赖 T4（挂 OverlayLayer/调 Launcher API）
- 去多余边：T6/T7 不依赖 T3（结算卡只消费 result dict，不碰游戏场景）；T8 依赖全部功能件（全流程驱动）

---

### Task 1: db_io.gd 拆分（CR-1 遗留收口）✅ 🎯（2026-08-25：db_io.gd 95 行 / db.gd 157 行；DB 接口签名零变化；test_db×3 + RG-5 write/verify 全绿）

**复杂度**: 中

**Scope:**
- `core/db_io.gd`（新建）：JSON 读写原语、`.corrupt` 兜底、防抖 flush 调度、目录初始化
- `core/db.gd`（修改）：瘦身为接口层，委托 DBIO；接口签名全部不变；`_notification`/`_exit_tree` flush 钩子留 DB（调 DBIO.flush_all()）
- 不触碰：其他 autoload、页面

**Constraints:**
- DB 对外接口（get_profile/save_profile/upsert_record/...）签名与行为零变化
- ≤200 行/文件；类型标注完整、中文注释

**Acceptance:**
- [ ] RG-5 双相回归通过：强写数据强杀重启不丢；正常退出防抖 flush 完整恢复
- [ ] test_db 三遍全量通过（与 CR-1 基线一致）
- [ ] db.gd ≤200 行，编辑器零报错

---

### Task 2: GameModule 协议基类 ✅ 🎯（2026-08-25：core/game_module.gd 33 行；协议七项 1~4（信号+boot/pause/resume 声明）通过，POST-CHECK 全过，--import 零报错）

**复杂度**: 低

**Scope:**
- `core/game_module.gd`（新建，class_name GameModule extends Node）：quit_requested/achievement_unlocked 信号 + boot(ctx)/pause_game()/resume_game() 接口，按 design.md §2
- 不触碰：其他文件

**Acceptance:**
- [ ] 协议七项检查第 1~4 项（boot/pause/resume/quit_requested 声明）通过
- [ ] POST-CHECK 9 项通过；单文件 ≤200 行

---

### Task 3: tetra_nova 迁入 + 协议适配 ✅ 🎯（2026-08-25：src/ 27 文件迁入（排除 shot.gd/shot_driver.gd/probe*.uid）；res://scripts|scenes 旧引用零残留；SAVE_PATH 注入 main/ui 3 存档点 + game.gd 删 stage.log；adapter 85 行（回菜单按钮只读 _ui._over，src 零 diff）；meta scene→module.tscn；基线 test_runner/probe3/probe4 在 src/ 下重跑全绿 exit 0）

**复杂度**: 高

**Scope:**
- `games/tetra_nova/src/`：源工程整拷贝（scenes/scripts/.uid/.import）
- src 最小 diff（每处 ≤5 行，D1=A）：main.gd `_save_best`、ui.gd `on_run_over`/`_refresh_best` 改读 `SAVE_PATH` 变量（默认 `"user://save.cfg"`）；game.gd 删 stage.log
- dev 脚本：test_runner/probe3/probe4 无文件写入保留（基线重跑用）；`shot_driver.gd`（写 user://shot_*.png）改读 SAVE_PATH 或排除迁出（保证 grep 验收）
- `module.tscn` + `module_adapter.gd`（≤200 行）：boot 注入 SAVE_PATH+计时；run_over→结算屏"回菜单"→quit_requested({score,playtime,achievements:[],extra:{wave,lines,max_combo,bosses}})；pause/resume 转发（源无则补 set_paused ≤30 行）；reset_run()
- `meta.json`：scene → `res://games/tetra_nova/module.tscn`
- 不触碰：源工程原目录 `projects/nova_arcade/tetra-nova-godot/`（冻结参考）

**Constraints:**
- 存档只写 ctx.save_dir 下（hybrid §4.2 第 5 项 grep 验证，无硬编码 user:// 绝对路径残留，含 shot_driver 等 dev 脚本）
- adapter 不引用 Shell 任何类型，只依赖 GameModule 基类 + src 节点

**Acceptance:**
- [ ] 协议七项全部通过（含 save_dir grep、生命周期运行验证留 T8）
- [ ] 源工程基线 test_runner/probe3/probe4 在 `src/` 下重跑全绿
- [ ] src diff ≤5 行/处，共 3 存档点 + 1 日志删除（shot_driver 如改道另计 ≤5 行）

---

### Task 4: LauncherService + 转场 ✅ 🎯（2026-08-25：launcher.gd 200 行（拆 launcher_transition 53 行/launcher_util 35 行）；状态机四态；视口切换仅在遮罩全不透明后（vp_at_full_mask 钩子，T8 断言通过）；play_again reset_run 复用不重建不耗 trial；launching 失败回滚 idle+Toast；Registry.similar 新增；autoload 共 8。注：4.5 屏幕方向枚举为 DisplayServer.ScreenOrientation.SCREEN_PORTRAIT/LANDSCAPE）

**复杂度**: 高

**Scope:**
- `services/launcher.gd`（Autoload）：状态机 idle→launching→running→quitting；launch/quit 编排/play_again，按 design.md §3 时序
- `services/launcher_transition.gd`：转场状态机 + 视口/方向切换（遮满后执行）
- `shell/components/transition_overlay.tscn/.gd`：fade_out/fade_in（300ms，await 信号），遮罩期拦截输入
- `core/registry.gd`（修改）：新增 `similar(gid)`（同 category+tag 匹配，无则空数组），其余接口不变
- `project.godot`：autoload +2（Launcher 在 TrialGuard 后、共 8）
- 不触碰：CoreManager 桩逻辑（arcade 分支只走桩路径）

**Constraints:**
- 视口切换必须在 `await fade_out()` alpha==1 之后（override §2 铁律，Risk-1）
- ctx={save_dir:"user://saves/{gid}/",trial_mode,owned,best,viewport_size}（通用规则，Launcher 不感知具体游戏）；GameModule 只依赖 ctx
- launcher.gd ≤200 行，超限拆 launcher_transition.gd

**Acceptance:**
- [ ] launch 时序断言：权益→DB→fade_out→遮满后切视口→boot→fade_in（T8 headless 验证）
- [ ] quit 时序断言：冻结→遮满后恢复 Shell 默认视口（当前 720×1560 portrait）→queue_free→强写三字段→game_finished
- [ ] play_again 走 reset_run 不重建 module、不消耗 trial（仅 session+1）；trial 耗尽 → PurchaseDialog Mock + 中止
- [ ] launching 失败（load/boot 异常）→ 回退 idle + Toast；sessions/finish_count 读现值递增写入（非字面 +1）
- [ ] RG-8：未 launch 时首页/Tab/主题行为与 CR-1 一致

---

### Task 5: TrialGuard ✅ 🎯（2026-08-25：services/trial_guard.gd 37 行；left=plays−trial_used（非 trial 返回 -1）；consume 强写 DB.upsert_record + EventBus.trial_consumed(gid,left)；T8 验证 trial 消耗/耗尽→Mock 链路）

**复杂度**: 低

**Scope:**
- `services/trial_guard.gd`（Autoload）：left(gid)/consume(gid)；trial_used 强写 + trial_consumed 信号，按 design.md §4；left() 仅 price_model=trial 时调用，非 trial 返回 -1 不抛错
- 不触碰：DB 接口（仅调用）

**Acceptance:**
- [ ] left=meta.trial.plays−trial_used；consume 后 DB 立即落盘（强杀重启不丢，RG-7）
- [ ] owned 恒 true 不消耗；信号参数 (gid, left) 正确

---

### Task 6: ResultOverlay + PurchaseDialog ✅ 🎯（2026-08-25：result_overlay.gd 145 行 + purchase_dialog.gd 33 行，运行时由 Launcher._ensure_cards 挂 OverlayLayer（main.tscn 零改动，RG-8）；评价按钮解读注记2（playtime<600s→「还需 X 分钟」/局数不足→「还需 N 局」）；similar 空数组整区隐藏（本 CR tetra_nova 唯一游戏→隐藏，Q2=A）；全字段 ThemeTokens.color()；按钮连接 _ready 一次性建立，show_card 可重复调用）

**复杂度**: 中

**Scope:**
- `shell/components/result_overlay.tscn/.gd`：游戏名/本局分数·时长/历史最高/成就 Toast 逐条/按钮（再来一局/评价*/返回）/"换一个"区（消费 Registry.similar，空则整区隐藏；非空时卡片点击弹 Toast"详情页即将上线"，openDetail 归 CR-3），按 design.md §5
- `shell/components/purchase_dialog.tscn/.gd`：Mock（名称+¥price+[暂不购买]）
- 挂 OverlayLayer；全字段 ThemeTokens.color()
- 不触碰：Launcher 内部时序（仅消费其 API/信号）

**Acceptance:**
- [ ] 评价* 亮灭逻辑：finish_count≥2 且本次 playtime≥600s，否则置灰"还需 X 分钟"
- [ ] 再来一局→Launcher.play_again；返回→Nav 回首页；成就 Toast 逐条（DB.unlock+achievement_unlocked）
- [ ] 双主题下无硬编码色、无布局溢出（T9 GUI 验证）

---

### Task 7: 首页临时启动按钮 ✅ 🎯（2026-08-25：home.tscn PlayButton「▶ TETRA NOVA」y=0.74 + home.gd _on_play_pressed→Launcher.launch("tetra_nova")（注释标注 CR-3 移除）；ThemeTokens 样式；首页其余元素零改动（RG-8 test_main/test_nav 全绿）

**复杂度**: 低

**Scope:**
- `shell/pages/home.gd/.tscn`：临时"▶ TETRA NOVA"按钮 → `Launcher.launch("tetra_nova")`（CR-3 详情页就绪后移除，注释标注）
- 不触碰：其他页面

**Acceptance:**
- [ ] 点击触发 launch 全流程；按钮样式走 ThemeTokens
- [ ] RG-8 回归：首页其余元素无变化

---

### Task 8: headless test_launch 全流程 ✅ 🎯（2026-08-25：tools/test_launch.tscn/.gd 137 行；五步全链路 22 断言 ALL PASS exit 0——launch(ctx 完整/遮满后切视口/sessions+1/trial 消耗)、quit(冻结/视口复原/finish_count+1/playtime 累加/结算卡显示)、play_again(同实例复用/不耗 trial/session+1)、trial 耗尽(PurchaseDialog Mock+IDLE+视口复原)、坏 meta(回退；headless res:// 只读时优雅 SKIP)。修复记录：DB.get_record 返回内存引用→测试 _rec() 改 duplicate 快照；headless playtime<1s→launch 后等待 2.0s）

**复杂度**: 高

**Scope:**
- `tools/test_launch.tscn/.gd`：驱动 launch→boot→quit 全链路断言，按 design.md §10；headless quit 触发：直接调 src `game.game_over()`（→run_over→结算屏）驱动 quit_requested 链路
- 不触碰：功能代码（发现缺陷回退对应 Task）

**Acceptance:**
- [ ] ctx 完整注入（save_dir/trial_mode/owned/best/viewport_size）
- [ ] 视口切换发生在遮罩 alpha==1 之后（时序断言，Risk-1）
- [ ] DB 强写三字段（sessions/finish_count 读现值递增，二次 launch 不重置）+ trial 计数正确；play_again session+1 且不消耗 trial；trial 耗尽中止并弹 Mock
- [ ] launching 失败分支回退 idle（临时坏 meta：scene 指向不存在路径，模拟 load 失败断言）
- [ ] 退出码 0，无脚本错误

---

### Task 9: GUI 双轨验收 + 全量回归 ✅ 🎯（2026-08-25：neon/elegant × home/run/result 截图 read_image 逐张断言通过；视觉检查发现并修复 3 缺陷——RUNNING 时 App 未隐藏（LauncherUtil.set_game_ui）、OverlayLayer 父层恒关（show_card 开/关父层）、tetra_nova CanvasLayer HUD 泄漏（adapter._set_ui_visible）；test_launch 25 断言 ALL PASS + RG-1~8 + 源基线全绿；acceptance_report.md 落盘含遗留项登记）

**复杂度**: 高

**Scope:**
- GUI：启动画面/游戏运行中/结算卡截图 read_image（neon/elegant × 720×1560）；断言 Shell 侧元素（遮罩/结算卡/按钮）无溢出、主题正确；游戏画面属 tetra_nova 自带风格不判
- 回归：RG-1~8 逐条 + test_smoke + 源基线（T3 已跑则复核）
- 产出 `acceptance_report.md`（含截图路径、RG 结果表）

**Acceptance:**
- [x] 双主题×三画面截图通过 read_image 断言
- [x] RG-1~8 全绿；test_smoke/test_db/源基线全绿
- [x] acceptance_report.md 落盘，遗留项（D3 挂账/Mock 桩位）显式登记
