# 需求：CR-2 Launcher 完整生命周期 + TETRA NOVA 接入 GameModule

> 依据 requirements_plan.md（Q1~Q7 已选单确认，2026-08-24）。协议基线：client-design §3.1 / §4.6、runtime-design GDScript 运行时节、override §2。

## 背景

CR-1 shell-bootstrap 已闭环（六 Autoload + 四 Tab + 主题 + DB/Registry + tetra_nova meta 占位）。M1 剩余核心：游戏能真正跑起来并走完 Launcher 全生命周期（关闭 override §5.4 挂账）。源工程 `projects/nova_arcade/tetra-nova-godot/` 零素材程序绘制，headless 基线（test_runner/probe3/probe4）交付前全绿。

## 已确认决策（2026-08-24 选单逐题确认）

| # | 决策 |
|---|---|
| Q1 | 整工程拷贝进 `games/tetra_nova/`，游戏 UI 原样保留，仅加协议适配层；源工程冻结为参考 |
| Q2 | §4.6 全时序实现；PurchaseDialog/DLC 检查/Mock 支付留桩位（M3/M4/M5 填实）；"换一个"similar() 为空按设计隐藏 |
| Q3 | result.achievements 透传 + DB.unlock + Toast 逐条；AchievementEngine 判定逻辑留 M2 |
| Q4 | 首页临时"▶ TETRA NOVA"按钮 + 测试场景驱动 Launcher.launch(gid)；正式详情 CTA 归 CR-3 |
| Q5 | 试玩耗尽 → PurchaseDialog Mock（Toast "¥X 解锁完整版（M4 接支付）"后中止启动） |
| Q6 | meta 省略 viewport_size/orientation → 继承 720×1560 竖屏；orientation 分支代码实现，本 CR 无横屏实测对象（M2 魔塔验证） |
| Q7 | 旧 user:// 根目录存档不迁移，save_dir 全新起 |

## 用户故事

- 作为玩家，我点首页"▶ TETRA NOVA"→ 转场进游戏 → 打完一局退出 → 看到结算卡（本局分数/时长/历史最高）并可"再来一局"或返回。
- 作为未购买玩家，我试玩次数用尽后再启动会被提示付费（Mock）。
- 作为开发者，新游戏按同一协议接入即可被 Launcher 拉起（TETRA NOVA 为首个实例）。

## 功能需求

| # | 需求 |
|---|---|
| FR-1 | `LauncherService.launch(gid)`：Registry.lookup → DLC 检查（桩：空跳过）→ 权益检查（free/owned 通过；trial 走 TrialGuard，剩 0 弹 PurchaseDialog Mock 中止；paid 未拥有同）→ DB.upsert_record(last_played/sessions+1) |
| FR-2 | 转场：App 淡出 300ms → **遮罩完全盖住后**读 meta.viewport_size（默认 [720,1560]）/orientation（默认 portrait）→ content_scale_size + 方向切换 → GameHost.visible=true（override §2 铁律） |
| FR-3 | runtime 分发：pck/html → `load(meta.scene).instantiate()` 入 GameHost；arcade → CoreManager 桩路径（本 CR 不实测）。`module.boot(ctx)`，ctx={save_dir: "user://saves/tetra_nova/"（示例值；通用规则 `user://saves/{gid}/` 见 design §3）, trial_mode, owned, best, viewport_size} |
| FR-4 | 退出处理：quit_requested(result) → 冻结 module → 遮罩淡入内恢复 Shell 默认视口（示例值 720×1560 竖屏，见 design §3）+ queue_free + GameHost 隐藏 → DB 强写（total_playtime/best/finish_count）→ EventBus.game_finished → 400ms 后弹 ResultOverlay |
| FR-5 | ResultOverlay 结算卡：游戏名/本局分数·时长/历史最高；新解锁成就 Toast 逐条（DB.unlock，Engine 判定留 M2）；[▶ 再来一局]（重走 FR-2~3）/[✎ 评价*]（finish_count≥2 且本次 playtime≥600s 才亮，否则置灰"还需 X 分钟"）/[返回]；"换一个"similar() 为空隐藏整区 |
| FR-6 | `TrialGuard`：left(gid)=meta.trial.plays−record.trial_used；consume 强写 + EventBus.trial_consumed(gid,left)；owned 恒 true（trial 已购） |
| FR-7 | TETRA NOVA 迁入 `games/tetra_nova/`：场景树+脚本整拷贝，新增协议适配层实现 boot/pause_game/resume_game/quit_requested(result={score,playtime,achievements,extra})；存档只写 ctx.save_dir（原 user:// 根目录写入点全部改道）；meta.json.scene 指向迁入后主场景 |
| FR-8 | GameModule 协议合规七项（hybrid §4.2）：boot/pause/resume/quit_requested 实现、result 三字段、save_dir 无硬编码 user:// 绝对路径（grep）、meta 完整、Shell 内走通生命周期 |
| FR-9 | 入口：首页临时"▶ TETRA NOVA"按钮（CR-3 详情 CTA 就绪后移除）+ `tools/test_launch.tscn` 测试场景驱动 launch/quit 全流程（headless 可验部分 + GUI 截图视觉项） |
| FR-10 | 附带收口：db.gd 拆出 `db_io.gd`（落盘原语，CR-1 验收遗留）；QualitySettings 裁剪方法恢复（resolve_texture/get_scalar 纹理分档，T2 挂账） |

## 非功能需求

- mobile 渲染器 / ≤200 行单文件 / 页面不含业务规则 / GameModule 不感知 Shell（只依赖 ctx）
- 视口切换必须在遮罩不透明后执行（override §2）；主题色零硬编码（结算卡/入口按钮走 ThemeTokens）
- 存档强写分级：trial_used、playtime/best/finish_count 立即落盘

## 范围外（明确不做）

详情页 CTA 状态机（CR-3）、真实支付/票据（M4）、DLC 下载（M5）、横屏游戏实测（M2 魔塔）、AchievementEngine 判定（M2）、评价提交链路（本地评价 M1 后段，本 CR 仅结算卡按钮态）、首页"继续游戏"区块重渲染（CR-3，client-design §4.6 步骤13）、Analytics 上报（M3，本地以 EventBus 信号为数据源）

## 验收口径（WHEN-THEN 摘要）

1. WHEN 测试场景调 launch("tetra_nova") THEN 遮罩转场后游戏画面出现、boot 收到完整 ctx
2. WHEN 游戏发 quit_requested THEN 视口恢复 Shell 默认（本 CR 即 720×1560）、module 释放、DB 强写三字段、结算卡弹出且数值正确
3. WHEN 试玩 3 次耗尽再 launch THEN PurchaseDialog Mock 提示且游戏未启动
4. WHEN "再来一局" THEN 重走转场+boot，session+1
5. WHEN 迁入后跑源工程 headless 基线（test_runner/probe3/probe4）THEN 全绿（逻辑零回归）
6. WHEN GUI 实测启动/退出/结算卡 THEN read_image 截图核验无溢出、双主题下正常
7. WHEN CR-1/CR-2 回归（RG-1~8 + test_smoke）THEN 全过
