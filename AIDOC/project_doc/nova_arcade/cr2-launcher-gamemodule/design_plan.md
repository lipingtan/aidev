# 设计计划：CR-2 Launcher 完整生命周期 + TETRA NOVA 接入 GameModule

> 阶段：设计计划（用户确认后才输出 design.md）。格式：go-development-workflow.md 附录 B。前置：requirements.md 已定稿（Q1~Q7，2026-08-24）。
> 源工程实查结论（设计依据）：存档点仅 3 处——`user://save.cfg`（main.gd `_save_best` / ui.gd `on_run_over`+`_refresh_best`）、`user://stage.log`（game.gd dev 日志）；退出时机 = `game.game_over()` → `run_over` 信号 → 结算屏（ui.gd `on_run_over`，含菜单按钮）。

## 设计方向

Shell 侧新增 Launcher/TrialGuard 两个 Autoload + 转场遮罩与结算卡组件，按 client-design §4.6 全时序编排（权益→DB→遮罩转场→视口切换→boot→quit→恢复→强写→ResultOverlay）；tetra_nova 整拷贝进 `games/tetra_nova/src/`，顶层 `module.tscn` 包装 + `module_adapter.gd` 实现 GameModule 协议基类（ctx 注入 SAVE_PATH、run_over 结算屏"退出"→quit_requested），src 脚本最小 diff；附带收口 db_io.gd 拆分。

## 技术选型

| # | 决策点 | 方案 A | 方案 B | 推荐 |
|---|---|---|---|---|
| D1 | tetra_nova 适配方式 | 最小 diff 改拷贝（src/ 原样 + module.tscn 包装 + adapter 注入 SAVE_PATH，src 改动 ≤5 行/处） | wrapper 场景 + 运行时 monkey-patch（不改 src 一行，靠信号转发/路径重写钩子） | **A**（拷贝即自有代码；B 的 patch 链路在 4.5 下难测且脆弱） |
| D2 | ResultOverlay / PurchaseDialog 形态 | 均 OverlayLayer 独立场景（result_overlay + purchase_dialog），M4 支付只填 purchase_dialog 内容 | 结算卡用 Dialog 组件、购买仅 Toast | **A**（§4.6 步骤 14 字段多，独立场景好维护；购买弹窗 M4 直接升级） |
| D3 | QualitySettings 裁剪方法恢复时机 | CR-2 不做（tetra_nova 零素材无纹理分档需求），挂账至首个有纹理的 CR | 本 CR 顺手恢复 resolve_texture/get_scalar | **A**（无消费方，恢复=无验收对象的死代码） |
| D4 | Autoload 注册方式 | Launcher/TrialGuard 注册为 autoload（共 8 个，页面/组件直连） | 非单例：Main._ready 实例化注入 | **A**（与 CR-1 六 autoload 模式一致；client-design §4 服务表即按单例设计） |
| D5 | 退出触发点 | 结算屏"回菜单"按钮 → quit_requested（玩家主动）；游戏内无其他退出路径 | 结算屏自动 N 秒后退出 + 按钮双通道 | **A**（协议语义=游戏主动退出；自动退出让玩家丢结算信息） |

## 澄清问题

- [Question-1] D1 tetra_nova 适配方式（方案见上方技术选型表 D1 行）？
  - 业界最佳实践：Steam/Epic 等平台的 SDK 接入模式=拷贝自有代码后直接改造（adapter 薄层），运行时 patch 仅用于无法取得源码的第三方二进制
  - 推荐答案及理由：**A**。源工程已整拷贝为自有代码，直接改最干净；monkey-patch 在 Godot 4.5 GDScript 下无成熟机制，信号转发/路径钩子链路难测且脆弱
  [Answer-1] A（2026-08-24 选单确认）
- [Question-2] D2 ResultOverlay / PurchaseDialog 形态（方案见上方技术选型表 D2 行）？
  - 业界最佳实践：结算页（post-game screen）在手游中均为全屏/大卡片独立页面（含多按钮+数据展示）；付费弹窗为模态对话框，两者 UI 复杂度都超出 Toast 承载
  - 推荐答案及理由：**A**。§4.6 步骤 14 字段多（分数/时长/最高/成就/3 个 CTA），独立场景好维护且双主题走 ThemeTokens；purchase_dialog 独立后 M4 接真实支付只换内容不动结构
  [Answer-2] A（2026-08-24 选单确认）
- [Question-3] D3 QualitySettings 裁剪方法恢复时机（方案见上方技术选型表 D3 行）？
  - 业界最佳实践：按需实现（YAGNI）——无消费方的接口恢复属于死代码，增加回归面却不产生验收价值
  - 推荐答案及理由：**A**。tetra_nova 纯程序绘制零纹理，resolve_texture/get_scalar 分档无任何调用方；挂账至首个有素材的 CR（魔塔/真实游戏接入时）再恢复并带验收
  [Answer-3] A（2026-08-24 选单确认）
- [Question-4] D4 Autoload 注册方式（方案见上方技术选型表 D4 行）？
  - 业界最佳实践：Godot 项目服务层惯例=Autoload 单例（CR-1 已有 Nav/DB/Registry/EventBus/ToastLayer/QualitySettings 六个）；client-design §4 服务表本身即按单例设计
  - 推荐答案及理由：**A**。与 CR-1 模式一致，页面/组件直连无注入链；8 个 autoload 在 4.5 下无性能问题，且避免 Main 持有并转发服务的样板代码
  [Answer-4] A（2026-08-24 选单确认）
- [Question-5] D5 退出触发点（方案见上方技术选型表 D5 行）？
  - 业界最佳实践：roguelite 类（Hades/Dead Cells）结算/死亡界面均由玩家主动选择继续或返回，无自动跳转——自动退出会让玩家丢失结算信息
  - 推荐答案及理由：**A**。协议语义即"游戏主动退出"，按钮触发最贴切；自动 N 秒与 §4.6 步骤 14"展示结算卡"矛盾（刚展示就要消失）
  [Answer-5] A（2026-08-24 选单确认）

## 风险点

- [Risk-1] 视口切换时机：遮罩未盖满就切 content_scale_size 会闪帧（override §2 铁律）→ 转场状态机显式等待 alpha=1 后再切，test_launch 断言时序
- [Risk-2] db.gd 拆 db_io.gd 属重构 → RG-5 双相 + test_db 三遍全量回归，接口签名不变
- [Risk-3] 源工程迁入后 headless 基线可能因路径/uid 变化失败 → test_runner/probe3/probe4 在 `games/tetra_nova/src/` 下重跑，失败项逐个定位（.import/.uid 随拷贝）
- [Risk-4] tetra_nova 游戏画面与 Shell 主题无关（自带霓虹风），双主题下仅结算卡/遮罩需走 ThemeTokens → 视觉验收只断言 Shell 侧元素

## 用户确认记录

| 日期 | 确认内容 | 方式 |
|---|---|---|
| 2026-08-24 | D1~D5 全部采纳推荐 A，设计计划通过；design.md 据此产出 | 选单确认（ask_user_question） |

