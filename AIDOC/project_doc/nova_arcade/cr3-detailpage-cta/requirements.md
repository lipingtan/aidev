# 需求：CR-3 详情页 CTA 状态机 + 卡片点击导航 + 首页「继续游戏」区块

> 依据 requirements_plan.md（Q1~Q6 全部采纳推荐，2026-08-25 确认）。协议基线：client-design §4.0/§4.2/§4.5/§4.6/§6、override §2。
> CR-2 已交付 Launcher 完整生命周期 + GameModule 接入；本 CR 补齐「列表 → 详情 → 开玩」入口链路并移除临时 PlayButton。

## 背景

CR-2 后唯一开玩入口是首页临时「▶ TETRA NOVA」按钮（Q4=CR-3 移除）。client-design §4.6 全流程的入口侧（卡片点击 → 详情 CTA → launch）与步骤 13（游戏结束后首页「继续游戏」区块重渲染）在本 CR 落地。目录内仅 tetra_nova 一款游戏，无评分/评价/下载量数据源。

## 已确认决策（2026-08-25 选单逐题确认）

| # | 决策 |
|---|------|
| Q1 | 详情页最小集：基础信息块 + 相关推荐（similar 非空才显示，本 CR 恒隐藏）+ CTA 固定区；评分分布/精选评价/截图轮播不写代码，标注 M2/M3 |
| Q2 | 「继续游戏」长按移除仅清 last_played（强写立即落盘），保留 best/total_playtime/finish_count/trial_used（权益红线） |
| Q3 | CTA 四分支全实现：free/owned→launch；trial 剩 N>0→launch；trial 剩 0 / paid 未拥有→PurchaseDialog Mock（复用 CR-2） |
| Q4 | CTA 重算接 records_updated + trial_consumed（均已存在：DB.upsert_record 后发前者、TrialGuard.consume 发后者）；payment_succeeded/dlc_installed 注释占位不接线（M3/M4） |
| Q5 | PlayButton 单独收尾 Task：卡片导航 + CTA 全绿后移除，同步更新 home 回归断言 |
| Q6 | 「为你推荐」实现但条件渲染：for_you() 扣除「继续游戏」已展示项后为空 → 整块隐藏；本 CR 卡片点击验收由「继续游戏」承载 |

## 用户故事

- 作为玩家，我点任意游戏卡片进入详情页，看到按我的权益显示的正确按钮（试玩剩几次 / 购买价 / 直接开玩），点一下就开玩。
- 作为玩过游戏的玩家，回 Shell 后首页「继续游戏」立刻出现该游戏；长按可把它从列表移除而不丢历史成绩。
- 作为试玩耗尽的玩家，详情页 CTA 变为「¥X 解锁完整版」，点击看到购买弹窗（Mock）。

## 功能需求

| # | 需求 |
|---|------|
| FR-1 | `Recommender`（services/recommender.gd）：`continue_row() -> Array[GameMeta]` = records 中 last_played>0 按倒序 top5；`for_you() -> Array[GameMeta]` = Registry 全量按 title 排序，扣除 continue_row 已展示项 |
| FR-2 | 首页「继续游戏」区块：横滑卡片（图标/名称/状态角标）；空 → 整块隐藏；**长按卡片** → 「从记录移除？」确认气泡 [移除] → DB.upsert_record(last_played=0)（强写）→ 立即 re-render 本区块 |
| FR-3 | 首页「为你推荐」区块：for_you() 为空 → 整块隐藏；卡片点击 → openDetail(gid) |
| FR-4 | DetailPage（`pages/detail.tscn/.gd`）：基础信息块（图标/名称/版本/标签行/简介，数据全来自 meta）+ 相关推荐区（Registry.similar 非空才渲染，本 CR 恒隐藏）+ 底部固定 CTA 区（hint 行 + CTA 按钮）；入页转场 220ms 右滑入（client-design §4.0），返回 pop |
| FR-5 | CTA 状态机 `_cta_state()`：free/ad/iap → [▶ 开玩]；trial 剩 N>0 → [▶ 试玩 · 剩 N 次] hint「试玩结束后 ¥X 解锁全部」；trial 剩 0 → [¥X 解锁完整版] hint「试玩次数已用完」；paid 未拥有 → [¥X 购买] hint「买断制 · 一次购买永久拥有」；owned → [▶ 开玩] hint「上次最高 {best} 分 · 继续」（record.best = 0 → hint 显示「首次开玩」；best 字段类型为 int，单位为分，格式化为整数直接显示，无小数）；DLC/Arcade 分支仅预留枚举不实现。开玩类点击 → `Launcher.launch(gid)`；购买类点击 → PurchaseDialog Mock |
| FR-6 | CTA 重算：DetailPage 监听 EventBus.records_updated + trial_consumed（本游戏 gid 过滤）后重算 CTA；payment_succeeded/dlc_installed 注释占位（M3/M4 接线） |
| FR-7 | openDetail 导航：卡片点击 → `Nav.push(DetailPage, gid)`（220ms 右滑入，同曲线弹出）；详情页返回键/系统返回 → pop 回原 Tab；CTA→launch 复用 CR-2 转场链路（遮罩全不透明后切视口），结算卡返回后回原 Tab |
| FR-8 | 移除首页临时 PlayButton（home.tscn/home.gd，含「CR-3 移除」注释代码）；同步更新 test_main/test_nav 相关断言 |
| FR-9 | `tools/test_detail.tscn/.gd`：CTA 状态机逐分支断言（free/trial N>0/trial 0/paid/owned × 文案 + 点击行为）+ 「继续游戏」渲染/长按移除/重渲染 + openDetail push/pop；headless 可跑 |
| FR-10 | GUI 双轨：neon/elegant × 首页（含「继续游戏」区块）/详情页截图 read_image 断言（无溢出、ThemeTokens 合规）；RG-1~8 + test_launch/test_db×3 全量回归 |

## 非功能需求

- mobile 渲染器 / ≤200 行单文件 / 页面不含业务规则（CTA 判定可下沉 Recommender 或 DetailPage 内私有函数，禁跨页散落）/ 主题色零硬编码（ThemeTokens.color）
- 存档分级：last_played 清除属 records 强写立即落盘；本 CR 不新增防抖写数据
- 性能：本地目录 ≤10 款，continue_row/for_you 内存即时；CTA 重算 O(1)
- 视口：DetailPage 为 Shell 内页面不切视口；仅 CTA→launch 走 CR-2 遮罩时序（override §2 铁律不变）

## 范围外（明确不做）

详情页评分分布/精选评价/截图轮播块（M2/M3）、真实支付与订单验签（M4）、DLC 下载与安装（M5）、Arcade ROM 选择面板与 BIOS 检查（M2）、搜索页/分类页筛选逻辑改动（仅卡片点击接线）、推荐算法真实画像（for_you 本 CR 为 title 排序退化版）、Analytics 上报（M3）

## 验收口径（WHEN-THEN 摘要）

1. WHEN 首页「继续游戏」卡片点击 THEN DetailPage 220ms 推入，CTA 按当前权益显示正确文案
2. WHEN tetra_nova trial 剩 N>0 点 CTA THEN Launcher.launch 正常开玩；一局结束回 Shell 后「继续游戏」区块出现该游戏且 CTA 重算为 N-1
3. WHEN trial 耗尽（N=0）THEN CTA 显示 [¥6 解锁完整版]，点击弹 PurchaseDialog Mock，游戏未启动
4. WHEN 长按「继续游戏」卡片并确认移除 THEN last_played 清除（best/total_playtime/finish_count/trial_used 不变）、区块立即 re-render、强杀重启后清除持久
5. WHEN 无任何 last_played>0 记录 THEN 「继续游戏」「为你推荐」整块隐藏，首页其余区块正常
6. WHEN PlayButton 移除后跑 test_main/test_nav + RG-8 THEN 全绿且首页仍有开玩入口（卡片点击）
7. WHEN GUI 双主题截图（首页/详情页 × neon/elegant）THEN read_image 断言无溢出、配色走 ThemeTokens
8. WHEN CR-1/CR-2 全量回归（RG-1~8 + test_launch/test_db×3/test_rg5 双相/源基线）THEN 全过
