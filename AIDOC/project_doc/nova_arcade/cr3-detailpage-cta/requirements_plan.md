# 需求计划：CR-3 详情页 CTA 状态机 + 卡片点击导航 + 首页「继续游戏」区块

> 依据 `nova-arcade-client-design.md` §4.0/§4.2/§4.5/§4.6/§6 + CR-2 design §9 范围外声明（步骤13「继续游戏」区块归 CR-3）。
> 类型：Shell CR（工程 `projects/nova_arcade/nova-arcade/`）。CR-2 已交付 Launcher 完整生命周期 + GameModule 接入，本 CR 补齐「从列表到详情到开玩」的入口链路。

## 需求理解

- **目标**：把 CR-2 遗留的临时 PlayButton 替换为正式入口——卡片点击 → 详情页（CTA 状态机按权益/价格模型驱动）→ CTA 点击 → Launcher.launch / PurchaseDialog Mock；同时落地首页「继续游戏」区块（last_played 倒序 + 长按移除），打通 client-design §4.6 全流程的入口侧。
- **范围**：
  1. DetailPage（`pages/detail.tscn/.gd`）：基础信息块 + 底部固定 CTA 状态机（free/trial/paid/owned × DLC/Arcade 分支预留）+ hint 副文案行；信号驱动重算（records_updated / trial_consumed，payment_succeeded / dlc_installed 仅监听预留）。
  2. openDetail 导航：首页「为你推荐」/分类页/搜索页卡片点击 → Nav push DetailPage（转场 220ms 右滑入）；返回 = pop。
  3. 首页「继续游戏」区块：Recommender.continue_row()（last_played 倒序 top5）横滑卡片；长按卡片 → 「从记录移除」确认气泡 → 移除后立即 re-render；records_updated 后 re-render。
  4. Recommender 本地版服务（§7：continue_row / for_you）。
  5. 移除 CR-2 临时 PlayButton（home.tscn/home.gd）。
- **预期效果**：用户从首页/分类/搜索点任意卡片进入详情，CTA 按试玩余量/购买状态显示正确文案；点击 CTA 开玩（或弹 Mock 购买弹窗）；玩完回 Shell 后「继续游戏」区块即时更新。

## 假设列表

- [假设-1] 本 CR 目录内仅 tetra_nova 一款游戏，meta 无评分/评价/下载量数据 → 详情页评分分布/精选评价/「X万人玩过」等块隐藏（非空态占位），骨架保留待 M2/M3。
- [假设-2] 支付全 Mock（M4）：trial 耗尽 / paid 未拥有 → 复用 CR-2 PurchaseDialog；购买成功仅本地置 owned + Toast，无订单验签。
- [假设-3] Arcade/DLC 状态机分支仅预留接口（`_cta_state()` 返回枚举），不实现 ROM 选择面板与 BIOS 检查（目录内无 arcade 游戏）。
- [假设-4] Nav 已具备 push/pop 能力（CR-1 RG 覆盖）；openDetail = `Nav.push(DetailPage, gid)`，无需改 Nav 基类。
- [假设-5] 「为你推荐」for_you() 本地版：无画像数据时退化为 Registry 全量按 title 排序（本 CR 仅 1 款 → 单卡）。

## 澄清问题

- [Question-1] 详情页内容块范围：client-design §4.5 列了截图轮播/评分分布/精选评价/相关推荐等 8 个块，本 CR 无数据源。做全骨架（空态占位）还是最小集（图标+名称+标签+简介+CTA+相关推荐）？
  - 业界最佳实践：商店类 App（Steam/App Store）详情页内容块按「有数据才渲染」处理，避免大片空态；首版聚焦 CTA 转化路径。
  - 推荐答案及理由：**最小集**——基础信息（图标/名称/版本/标签行/简介）+ 相关推荐（similar 非空才显示，本 CR 恒隐藏）+ CTA 固定区；评分/评价/轮播块代码不写、文档标注 M2/M3。理由：无数据源时全骨架只会产出占位 UI，增加主题适配与回归面，且本 CR 核心验收在 CTA 状态机。
  [Answer-1] 采纳推荐：最小集（基础信息 + 相关推荐条件显示 + CTA 固定区）；评分/评价/轮播块不写，标注 M2/M3（2026-08-25 确认）
- [Question-2] 「继续游戏」长按移除的语义：清除该 gid 的 last_played（区块消失但保留 best/playtime/sessions）还是整条记录删除？
  - 业界最佳实践：「从继续玩列表移除」通常只清 last_played/continue 标记，不毁历史统计（Steam 的「最近游玩」隐藏、手机系统级应用卸载才删数据）。
  - 推荐答案及理由：**仅清除 last_played**（DB.upsert 置 null/0，强写分级中 records 属强写）；best/total_playtime/finish_count/trial_used 全部保留。理由：误触长按不丢历史最佳与试玩计数（trial_used 删除会导致权益复活，安全红线）。
  [Answer-2] 采纳推荐：仅清 last_played（强写），保留 best/total_playtime/finish_count/trial_used（2026-08-25 确认）
- [Question-3] CTA 点击后的完整分支矩阵是否按 client-design §6 全实现（含 trial 剩 N>0 → launch、剩 0 → Mock 购买弹窗、free/owned → launch、paid → Mock 购买弹窗）？
  - 业界最佳实践：CTA 是唯一转化入口，状态机与点击行为必须一一对应可测；支付未接入时统一弹 Mock 并 Toast。
  - 推荐答案及理由：**是，四分支全实现**。launch 分支走 CR-2 Launcher.launch(gid)（权益检查/视口切换/结算卡链路已就绪）；购买分支复用 PurchaseDialog（暂不购买 → 关闭）。理由：与 RG-7/权益强写闭环，且 test_detail 可逐分支断言。
  [Answer-3] 采纳推荐：四分支全实现；开玩分支走 Launcher.launch，购买分支复用 PurchaseDialog Mock（2026-08-25 确认）
- [Question-4] CTA 重算信号面：本 CR 只接 records_updated + trial_consumed 两个真实信号，payment_succeeded / dlc_installed 是否只声明监听不实现处理？
  - 业界最佳实践：事件驱动 UI 按「信号已存在才接线」原则，避免为不存在的事件源写死代码。
  - 推荐答案及理由：**是**——EventBus 已有 trial_consumed（CR-2）；records_updated 需确认 DB 是否已发（无则本 CR 补发，属 DB 接口零签名变化）；payment_succeeded/dlc_installed 在 DetailPage 留注释占位不接线。理由：保持事件面与 M3/M4 真实源对齐，防止幽灵信号。
  [Answer-4] 采纳推荐：本 CR 接 records_updated + trial_consumed；payment_succeeded/dlc_installed 注释占位不接线（2026-08-25 确认）
- [Question-5] 移除临时 PlayButton 的时机：与本 CR 其他任务并行（同批提交）还是单独一个收尾 Task？
  - 业界最佳实践：临时入口与正式入口并存会造成双路径歧义，应在正式链路自测通过后立即移除。
  - 推荐答案及理由：**单独收尾 Task**（依赖卡片点击导航 + 详情页 CTA 均绿后执行），移除时同步更新 home 回归断言。理由：保证任何时点首页都有可用入口，避免中间态不可玩。
  [Answer-5] 采纳推荐：单独收尾 Task，卡片导航 + CTA 全绿后移除并同步回归断言（2026-08-25 确认）
- [Question-6] 「为你推荐」区块本 CR 是否实现（client-design §4.2 五区块之二）？仅 tetra_nova 一款时该区块与「继续游戏」内容重复。
  - 业界最佳实践：推荐位数据不足时隐藏整块（而非重复展示同一条），避免用户困惑。
  - 推荐答案及理由：**实现区块但条件渲染**——for_you() 结果扣除「继续游戏」已展示项后为空 → 整块隐藏；本 CR 实际表现为只有「继续游戏」一块。理由：区块代码一次到位（M2 多游戏后自动生效），且卡片点击→详情导航的验收需要「为你推荐」或「继续游戏」任一横滑卡承载（由「继续游戏」承载）。
  [Answer-6] 采纳推荐：实现区块但条件渲染（扣除继续游戏已展示项后为空 → 整块隐藏）；本 CR 卡片点击验收由「继续游戏」承载（2026-08-25 确认）

## 非功能需求建议

- **性能**：本地目录 ≤10 款，列表过滤/排序内存即时（<16ms）；CTA 重算 O(1)（读单条 record + meta）。
- **主题**：DetailPage 全字段走 ThemeTokens.color()；双主题截图验收沿用 CR-2 流程（GUI exe + 隔离 APPDATA）。
- **存档分级**：last_played 清除属 records 强写立即落盘；search_history 防抖 500ms（本 CR 不涉及搜索页改动则不动）。
- **视口**：DetailPage 为 Shell 内页面，不触发游戏视口切换（仅 CTA→launch 走 CR-2 遮罩时序）。

## 影响范围预判

- **涉及模块**：shell/pages（detail 新增、home 改）、services/recommender（新增）、core/nav（openDetail 包装，若 push API 足够则零改动）、core/db（records_updated 信号补发，若缺）、shell/components（确认气泡可复用 Toast 体系或新增小组件）。
- **涉及文件（预估）**：`pages/detail.tscn/.gd`、`pages/home.tscn/.gd`、`services/recommender.gd`、`core/nav.gd`、`core/db.gd`（可能）、`tools/test_detail.tscn/.gd`（新增）、`tools/test_launch.gd`（PlayButton 移除后回归复核）。
- **可能的副作用**：DB 补发 records_updated 信号 → RG-5 双相 + test_db 回归；home 结构变动 → RG-8 首页行为回归（test_main/test_nav）；DetailPage push/pop → CR-1 转场回归。
