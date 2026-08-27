# 需求计划：CR-2 Launcher 完整生命周期 + TETRA NOVA 接入 GameModule

> 阶段：需求计划（用户确认后才输出 requirements.md）。格式：go-development-workflow.md 附录 A。
> CR 类型判定（hybrid §1）：同时涉及 Shell 主工程（Launcher/TrialGuard/Services）与游戏接入协议 → **Shell CR + GameModule CR 双激活层**。
> 依据：client-design §3.1（GameModule 协议）、§4.6（启动全流程时序）、override §2/§5；源工程 `projects/nova_arcade/tetra-nova-godot/`（零素材程序绘制，headless 基线全绿）。

## 需求理解

- **目标**：关闭 CR-1 挂账 §5.4——游戏真正跑起来并走完 Launcher 全生命周期（§4.6），TETRA NOVA 作为首个 GameModule 实例接入协议
- **范围**：Shell 层 `services/launcher.gd`、`trial_guard.gd`、ResultOverlay 结算卡；GameModule 层 tetra_nova 迁入 `games/tetra_nova/` + 协议适配；附带收口 db_io.gd 拆分（CR-1 验收遗留）
- **预期效果**：玩家点首页"▶ TETRA NOVA"→ 转场进游戏 → 打完一局退出 → 结算卡（本局/历史最高/成就 Toast）可再来一局或返回；试玩 3 次耗尽提示付费（Mock）

## 假设列表

- [假设-1] tetra_nova 整拷贝后源工程冻结为参考，后续修改只发生在 `games/tetra_nova/`
- [假设-2] tetra_nova meta 省略 viewport_size/orientation → 继承 Shell 720×1560 竖屏；orientation 分支代码实现但本 CR 无横屏实测对象（M2 魔塔验证）
- [假设-3] 旧 user:// 根目录存档不迁移，save_dir 全新起（开发期数据可弃）
- [假设-4] AchievementEngine 判定逻辑留 M2；本 CR 只做 result.achievements 透传 + DB.unlock + Toast
- [假设-5] PurchaseDialog/DLC 检查/Mock 支付留桩位（M3/M4/M5 填实），本 CR 仅 Mock Toast

## 澄清问题

- [Question-1] 迁入粒度：A 整工程拷贝场景树+脚本进 `games/tetra_nova/`，游戏 UI（星空/HUD/选卡）原样保留仅加协议适配层；B 抽核心逻辑重写 UI 跟随 Shell 双主题？
  - 业界最佳实践：平台 SDK 接入（Steam/Epic 式）=拷贝自有工程后直接改造+薄适配层；重写 UI 仅当品牌强统一要求时
  - 推荐答案及理由：**A**。tetra_nova 零素材程序绘制自成一体，B 工作量翻倍且违反"GameModule 不感知 Shell"约束
  [Answer-1] A（2026-08-24 选单确认）
- [Question-2] Launcher 完整度：A §4.6 全时序实现，PurchaseDialog/DLC/Mock 支付留桩位，"换一个"similar() 为空按设计隐藏；B 仅 launch→boot→quit→结算卡最小闭环？
  - 业界最佳实践：盒子类客户端（Steam/Epic/Galaxy）启动管线均为完整权益+转场+结算链路，未就绪环节以明确桩位占位而非砍时序
  - 推荐答案及理由：**A**。client-design §4.6 已定稿，桩位标注清晰即可，避免 CR-3/M4 返工改骨架
  [Answer-2] A（2026-08-24 选单确认）
- [Question-3] 成就处理：A result.achievements 透传 + DB.unlock + Toast 逐条，Engine 判定留 M2；B 不接成就占位空数组？
  - 业界最佳实践：解锁提示（Toast/横幅）与判定逻辑解耦是常见分层——客户端先做展示链路，判定引擎后补
  - 推荐答案及理由：**A**。CR-1 的 DB.unlock 接口已有只差消费方，透传+Toast 成本低且 M2 接 Engine 时零改动
  [Answer-3] A（2026-08-24 选单确认）
- [Question-4] 启动入口：A 首页临时"▶ TETRA NOVA"按钮 + 测试场景驱动 Launcher.launch(gid)，正式详情 CTA 归 CR-3；B CR-2 先做详情页？
  - 业界最佳实践：里程碑式交付先用最小入口打通全链路，正式 UI 按依赖顺序后置（避免为入口重做阻塞核心路径）
  - 推荐答案及理由：**A**。与 CR-1 挂账划分一致（§5.4 本 CR、§5.3 CR-3），临时按钮 CR-3 就绪后移除
  [Answer-4] A（2026-08-24 选单确认）
- [Question-5] 试玩耗尽行为：trial 剩 0 → PurchaseDialog Mock（Toast"¥X 解锁完整版（M4 接支付）"后中止启动），与 §4.6 步骤 4 一致？
  - 业界最佳实践：Freemium 手游（Royal Match/Homescapes 类）试玩耗尽弹付费墙并中止，不静默放行
  - 推荐答案及理由：**照此执行**。与 §4.6 步骤 4 一致，Mock 仅 Toast+中止，M4 换真实支付不动流程
  [Answer-5] 确认（2026-08-24 选单确认）
- [Question-6] 视口：tetra_nova meta 省略 viewport_size/orientation → 继承 Shell 720×1560 竖屏；orientation 分支代码实现但本 CR 无横屏实测对象（M2 魔塔验证）？
  - 业界最佳实践：盒子类应用默认继承宿主视口，游戏 meta 显式声明才覆盖；分支逻辑先实现、实测随首个横屏游戏补齐
  - 推荐答案及理由：**照此执行**。tetra_nova 竖屏玩法与 Shell 一致；orientation 分支代码就位，M2 魔塔做首个横屏实测
  [Answer-6] 确认（2026-08-24 选单确认）
- [Question-7] 存档迁移：源工程旧 user:// 根目录存档不迁移，新 save_dir 全新起（开发期数据可弃）？
  - 业界最佳实践：正式版本才需要存档迁移/兼容策略；开发期数据直接弃用是通行做法
  - 推荐答案及理由：**照此执行**。本 CR 处于开发期，save_dir（user://saves/tetra_nova/）全新起，省去迁移代码
  [Answer-7] 确认（2026-08-24 选单确认）

## 非功能需求建议

- 性能：视口切换必须在遮罩不透明后执行（override §2 铁律）；转场淡出 300ms
- 数据：存档强写分级——trial_used、playtime/best/finish_count 立即落盘；GameModule 只写 ctx.save_dir
- 架构：GameModule 不感知 Shell（只依赖 ctx 注入）；≤200 行单文件；主题色零硬编码

## 影响范围预判

- 涉及模块：project.godot（autoload +2）、services/（launcher/trial_guard 新增）、shell/components/（result_overlay/transition_overlay 新增）、games/tetra_nova/（迁入+适配）、core/db.gd（拆 db_io.gd）
- 涉及文件（预估）：新增 ~10、修改 ~6（meta.json.scene、home.tscn 临时按钮、project.godot、db.gd 拆分）
- 可能的副作用：DB 拆分需 RG-5 回归；源工程 headless 基线（test_runner/probe3/probe4）迁入后须重跑全绿

## 用户确认记录

| 日期 | 确认内容 | 方式 |
|---|---|---|
| 2026-08-24 | Q1~Q7 全部采纳推荐，需求计划通过；requirements.md 定稿 | 选单逐题确认（ask_user_question） |
