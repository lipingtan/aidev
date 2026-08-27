# 设计计划：CR-3 详情页 CTA 状态机 + 卡片点击导航 + 首页「继续游戏」区块

> 阶段：设计计划（用户确认后才输出 design.md）。格式：go-development-workflow.md 附录 B。
> 前置：requirements.md 已定稿（Q1~Q6 全部采纳推荐，2026-08-25）。
> 源工程实查结论（设计依据）：CR-2 已交付 Launcher/TrialGuard/DB/Registry/Nav/EventBus；EventBus 已有 records_updated/trial_consumed 信号；DB.upsert_record 在 upsert 后发 records_updated；home.tscn 有临时 PlayButton 节点（CR-3 移除）。

## 设计方向

新增 `services/recommender.gd`（Recommender 本地版，continue_row/for_you）；改写 `pages/home.tscn/.gd`（加继续游戏 + 为你推荐区块，含长按确认气泡组件）；新增 `pages/detail.tscn/.gd`（DetailPage：基础信息块 + CTA 状态机）；导航接线（卡片点击 → Nav.push(DetailPage, gid)）；收尾移除 PlayButton。CTA 状态机逻辑内聚到 DetailPage 私有函数 `_cta_state(gid)`，不外散到其他页面。

## 技术选型

| # | 决策点 | 方案 A | 方案 B | 推荐 |
|---|---|---|---|---|
| D1 | Recommender 挂载位置 | `Services` 节点下非单例（与 Launcher 同层，client-design §2.2） | Autoload 单例 | **A**（client-design §1.1 架构已规定，Recommender 属服务层；Autoload 留给跨模块共享单例） |
| D2 | 长按移除确认气泡实现方式 | 复用 Toast/Dialog 体系：在 OverlayLayer 弹一个带"移除/取消"按钮的小弹窗（类 Android ContextMenu） | 在「继续游戏」区块内 inline 浮层（绑定到被长按卡片位置） | **A**（OverlayLayer 已存在且统一管理；inline 浮层需计算位置溢出，且单独为一个交互写新组件成本高；Toast 体系已有弹出动画复用） |
| D3 | DetailPage ≤200 行拆分策略 | 拆为 detail.gd（页面控制器，≤200行）+ `components/cta_bar.gd`（CTA 区，≤80行，含 `_cta_state()` 私有函数） | 全写在 detail.gd，靠 #region 分区，允许突破 200 行 | **A**（非功能需求明确 ≤200 行单文件；CTA 状态机是独立的业务单元，适合拆出；`_cta_state()` 测试时也更易 mock） |
| D4 | CTA 状态机触发方式 | DetailPage.on_enter/on_resume 主动拉取 + 监听 EventBus 信号（records_updated/trial_consumed 过滤 gid） | 轮询（每帧或定时） | **A**（事件驱动是 client-design §3.3 既定模式；trial_consumed/records_updated 信号已存在） |
| D5 | 首页「继续游戏」/「为你推荐」区块渲染方式 | 两块各用独立 `HBoxContainer` + 动态 `instantiate` 卡片（滚动用 ScrollContainer 横向）；records_updated 信号触发 `_render_continue_row()` 局部重绘 | 整个首页走 rebuild，重建所有区块 | **A**（局部重绘只动一个区块，避免 Banner 轮播被打断；client-design §4.2 明确「游戏结束后 renderHome() 重渲染首页继续游戏区块」而非全页重建） |
| D6 | Nav.push(DetailPage, gid) 的 gid 传参方式 | 通过 `on_enter(data: Dictionary)` 接收 `data["gid"]`（Page 基类协议，client-design §3.2） | DetailPage 暴露 `set_gid(gid)` 方法，push 后调用 | **A**（Page 基类协议已有 on_enter(data)，D6=方案A 零改动 Nav，与 CR-1 模式一致） |

## 澄清问题

- [Question-1] D1 Recommender 挂载位置（方案见上方技术选型表 D1 行）？
  - 业界最佳实践：推荐/个性化服务通常为无状态或轻状态的服务对象，不需要跨模块共享时挂载于局部更合适（可复用/可测试/无全局污染）
  - 推荐答案及理由：**A**。client-design §1.1 架构已划定 Services 层，Recommender 已在 §2.2 服务表内；本 CR 只有 home 消费 Recommender，无需 Autoload
  [Answer-1]A

- [Question-2] D2 长按确认气泡实现方式（方案见上方技术选型表 D2 行）？
  - 业界最佳实践：移动端长按菜单（Android ContextMenu / iOS ActionSheet）均走全局弹层，不做 inline 位置计算；避免卡片边缘裁剪问题
  - 推荐答案及理由：**A**。OverlayLayer 已存在于 Main.tscn；弹一个带两个按钮的确认弹窗（类 PurchaseDialog 结构，更轻），调用方只需 `ConfirmBubble.show(title, on_confirm)`；长按信号从 GameCard 发出，home.gd 处理后调 OverlayLayer
  [Answer-2]a

- [Question-3] D3 DetailPage 拆分策略（方案见上方技术选型表 D3 行）？
  - 业界最佳实践：UI 组件化拆分以「能独立测试」为标准；CTA 区含状态机逻辑，是自然的组件边界
  - 推荐答案及理由：**A**。detail.gd 负责入页数据绑定（meta 显示）+ 信号路由，cta_bar.gd 封装 `_cta_state()/refresh_cta()`；两文件均 ≤200 行，test_detail.gd 可直接注入 mock record 断言 CTA 状态
  [Answer-3]a

- [Question-4] D4 CTA 状态机触发方式（方案见上方技术选型表 D4 行）？
  - 业界最佳实践：事件驱动优于轮询，避免每帧 O(1) 检查造成不必要 CPU 占用
  - 推荐答案及理由：**A**。on_enter 做初始化渲染；EventBus.records_updated / trial_consumed 在 `_ready` 中连接（带 gid 过滤，非本页游戏的信号忽略）；payment_succeeded / dlc_installed 留注释占位（M3/M4）
  [Answer-4]a

- [Question-5] D5 首页区块渲染方式（方案见上方技术选型表 D5 行）？
  - 业界最佳实践：局部刷新是列表类 UI 的标准实践（RecyclerView notifyItemChanged 等），整页重建代价高且产生闪烁
  - 推荐答案及理由：**A**。继续游戏区块 `_render_continue_row()` 函数在 records_updated 时独立调用；为你推荐区块 `_render_for_you()` 在 on_enter/on_resume 时调用（本 CR 该区块内容相对静态）；两块均先 `free()` 旧卡片再重建，不做 diff（本地≤10款，全重建耗时可忽略）
  [Answer-5]a

- [Question-6] D6 Nav.push 传参方式（方案见上方技术选型表 D6 行）？
  - 业界最佳实践：Page 基类协议统一入口，避免各页面有各自的初始化方法导致调用方需感知页面内部
  - 推荐答案及理由：**A**。`Nav.push("res://pages/detail.tscn", {"gid": gid})`，detail.gd.on_enter(data) 读 `data["gid"]`；零改 Nav 基类
  [Answer-6]a

- [Question-7] `best` 字段在 owned hint 行的格式：client-design §4.5 写「上次最高 X · 继续」，X 是 record.best（int，单位：分）。best=0 时 hint 显示「首次开玩」，best>0 时显示「上次最高 {best} 分 · 继续」。test_detail 需要断言文案，格式是否就用整数直接拼接（如「上次最高 3200 分 · 继续」），不做千分符格式化？
  - 业界最佳实践：游戏类 App（Steam/App Store 成就）中高分通常带千分符（3,200），提升可读性；但若游戏分数本身不大（TETRA NOVA 波次分），整数亦可接受
  - 推荐答案及理由：**整数直接拼接，不做千分符**。理由：TETRA NOVA 当前得分为波次/回合计数（量级 ≤100），不需要千分符；且 GDScript str(int) 直接可用，避免引入格式化函数的测试复杂度；后续多游戏若有高分需求再按游戏 meta 声明 score_fmt 字段
  [Answer-7]同意

- [Question-8] ConfirmBubble（长按移除确认弹窗）是否复用 PurchaseDialog overlay 结构，还是新建轻量 `overlays/confirm_bubble.tscn`？
  - 业界最佳实践：确认类弹窗（二次确认）通常比支付弹窗轻量得多（仅标题 + 两个按钮），复用支付弹窗结构会引入不相关字段；通用 ConfirmDialog 是常见组件
  - 推荐答案及理由：**新建 `overlays/confirm_bubble.tscn`**（轻量：Label 标题 + [确认]/[取消] 两按钮，共用 OverlayLayer 弹出动画），暴露 `show(msg: String, on_confirm: Callable)` 接口。理由：与 PurchaseDialog 职责不同，复用会导致两者互相污染；且 ConfirmBubble 后续长按删除/清空历史等场景可复用
  [Answer-8]同意

## 风险点

- [Risk-1] home.gd 行数：加继续游戏 + 为你推荐两区块 + Banner 轮播 + 原有区块，home.gd 可能突破 200 行 → 拆 `components/section_continue.gd` + `section_for_you.gd` 承接区块逻辑，home.gd 只做组合
- [Risk-2] records_updated 信号发射时机：DB.upsert_record 已在 CR-2 game_finished 路径中强写并发信号；需确认 FR-2 长按移除（DB.upsert_record(last_played=0)）同样会触发 records_updated → home 区块自动 re-render；若 DB 实现中 upsert 只在 best/playtime 路径发信号则需补发
- [Risk-3] DetailPage on_resume vs on_enter：结算卡关闭后回 Tab 根页，不走 on_enter 而走 on_resume；CTA 需在 on_resume 时也触发 refresh_cta()，否则试玩耗尽一局后回到详情页 CTA 不更新
- [Risk-4] ScrollContainer 横向 + 长按手势冲突：Godot ScrollContainer 默认拦截触摸滑动事件，InputEventLongPress 或 _input 长按检测可能被吞 → 卡片需显式 mouse_filter + _unhandled_input 处理长按，或在 GameCard 内部实现长按计时器（500ms 阈值）
- [Risk-5] PlayButton 移除收尾 Task：test_main 和 test_nav 有断言 PlayButton 节点存在的用例 → 移除前必须先定位所有引用并同步更新断言，否则 headless 跑红
