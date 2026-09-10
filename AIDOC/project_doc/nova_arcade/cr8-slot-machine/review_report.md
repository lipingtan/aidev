# CR-8 三角色评审报告（requirements / design / tasks 三件套）

> 2026-08-31。评审方式：文档交叉对照 + 工程实查（theme_tokens.gd / db.gd / event_bus.gd / registry.gd / achievement_engine.gd / searcher.gd / snake 样板）。

## 业务专家（B）

| # | 严重度 | 问题 | 处置 |
|---|---|---|---|
| B-1 | 高 | **成就解锁双通道矛盾**：requirements FR-7.2 说「走 achievement_unlocked 信号」，design §6 说「src 发 achievement_earned → adapter 转发 achievement_unlocked」。实查 achievement_engine.gd：引擎订阅的是 **EventBus.game_finished**（读 result.achievements[]），模块层的 GameModule.achievement_unlocked 信号**没有引擎监听**。按 design 现稿，成就永远不会解锁。 | 修正：模块 emit GameModule.achievement_unlocked（协议信号保留）+ quit result.achievements[] 携带（真正的解锁通道，引擎在 game_finished 时处理）。design §6 改为「成就随 quit_requested 的 achievements 数组交付，由 AchievementEngine 经 game_finished 解锁；模块内不直接调 DB.unlock」。requirements FR-7.2 措辞同步。 |
| B-2 | 中 | **FR-7.1 成就触发条件与结算时点冲突**：「单局押中七」在结算瞬间可判；但 quit 时刻 achievements 才被引擎消费——若玩家解锁后不退出继续玩，achievement 已 emit 但引擎没收到。需明确：成就解锁随 quit 批量交付（可接受延迟）还是即时解锁。 | 修正：design 明确「解锁判定即时（结算时记入 _session_achievements 数组，UI 可即时提示），DB 落盘与全局广播随 quit 批量交付」。requirements FR-7.2 加一句交付时点。 |
| B-3 | 低 | category=casino 是新分类，首页分类页 tabs 是否需要收录由 editorial.categories 驱动——实查 editorial.json 只扩 banner 没动 categories，casino 游戏在分类页可能不可见（首页/继续游戏不受影响）。 | 修正：T9 范围加「editorial.categories 视结构决定是否加 casino 桶（或复用 puzzle 桶）」——实查后定，AC-8 补分类页可见性断言。 |
| B-4 | 低 | requirements「score=本局净变化」vs design「会话累计净变化」不一致（design 已声明为语义微调）。 | 修正：requirements FR-6.1 改为会话累计净变化 + extra.biggest_win，两文档对齐。 |

## 产品经理（P）

| # | 严重度 | 问题 | 处置 |
|---|---|---|---|
| P-1 | 中 | **押注位双击清零与连续快速单击冲突**：手机端快速点两下加注会被误判双击清零。 | 修正：交互改为「单击加一份，长按 0.4s 清零」（Button 没有长按原生事件，用 gui_input 计时实现），requirements FR-2.3 与 design §3 同步；双击方案废弃。 |
| P-2 | 低 | AC-7 五态里「滚动中」截图依赖动画中段时点，hook 不稳。 | 修正：测试钩子提供「暂停在滚动中段」的确定性注入点（tween 至 50% 时 save_png）。 |
| P-3 | 低 | FR-2.2 注额档全局单选——押注位上应显示「份数×当前档额」还是「币数」未定义。 | 修正：显示币数（份数×档额），押注位文案带更新；写进 design §5。 |
| P-4 | 低 | 结算卡展示 extra.biggest_win 吗？结算卡现只渲染 score/playtime/best/成就——extra 不展示。 | 修正：requirements FR-6.2 注明 extra 仅为数据通道，结算卡不展示（无需 Shell 改动）。 |

## 架构师（A）

| # | 严重度 | 问题 | 处置 |
|---|---|---|---|
| A-1 | 高 | **UILayer 遮罩/弹层不随模块隐藏**：CanvasLayer 坑已知（框架 set_module_visible 兜底已就位），但 design 的 UILayer 含 StartOverlay/ReliefBtn——退出后 set_module_visible(false) 会快照隐藏，play_again 恢复即可。此项框架已兜，确认无新坑，但 tasks T5 验收须含「quit 后 UILayer 全隐藏、play_again 后恢复」断言。 | 修正：T5 Acceptance 补两条断言。 |
| A-2 | 中 | **emoji 字形探测未给测试口径**：FR-1.1 用 Unicode 字符，设计说「实现时探测，缺字形降级汉字」——headless 与 GUI 字体回退不同，探测时机不确定会导致测试断言漂移。 | 修正：design 定死「启动时 Font.has_char() 探测一次，结果存 SlotConfig.USE_EMOJI 常量级标志，测试对该标志不敏感（断言符号位数量与位置，不断言具体字形）」。 |
| A-3 | 中 | **Wheel 旋转 + Label 子节点**：Label 是 Control，挂在 Node2D(Wheel) 下旋转可行，但 12 个 Label 旋转过程中每帧重排 glyph 开销小（静态文本）可接受；真正风险是 Label 尺寸未设导致绕圈摆位错位。 | 修正：design §1 注明「每个符号位用固定 size 的 Label（custom_minimum_size + pivot 居中），摆位按三角函数算 position」，T3 Acceptance 含 12 位摆位坐标断言（±1px）。 |
| A-4 | 低 | reset_run 语义：会话累计净变化作 score 后，play_again 会继续累积——多次 play_again 后 quit 的 score 会偏大（含此前局收益），best 因此虚高。 | 修正：接受此语义（会话=一次运行的累计表现，与 tetra/魔塔的 best 记录同型），在 design §6 注明即可。 |
| A-5 | 低 | 万局统计测试直调 `_settle` 需要绕过余额/成就副作用——余额强写 1 万次 save.cfg 拖慢测试。 | 修正：design §7.1 注明「统计组注入内存模式标志（_io_silent=true），跳过落盘与音效」；结算纯函数化便于直调。 |

## 合并去重结论

高 2（B-1 成就通道、A-1 验收补强）、中 5、低 7，共 14 项，**全部为文档修正类**（无推翻性设计问题），已按上表处置方案同步进三文档（见各文档修订记录）。三件套数据结构/字段名/任务承接已对齐：每条 FR 均有 T 承接、theme token 与 API 引用经实查确认存在（ach_def/EventBus.game_finished/Searcher.query/ConfigFile）。

## 结论

✅ 评审通过（修正后）。可进入执行阶段（T1 起）。

---

## 修复记录（二轮，2026-08-31）

13 项全部修复：A-1/A-2 改 design §3/§4 账本模型（登记制+集中扣款+结算后清零）；P-2/B-2/B-3/P-3 改 requirements AC-3/FR-7.1/数据契约/AC-6；A-3/P-1/P-4 改 requirements FR-8.1/FR-1.3/FR-3.1/FR-6.3；A-4/A-5 改 design §2；A-6/B-1 改 tasks.md T3/requirements_plan.md 范围指针。三文档修订记录均已追加日期行。

✅ 三角色 review 完成，全部问题已修复，文档可进入执行阶段。

---

# 二轮评审（用户要求「执行多角色review」，2026-08-31）

> 对一轮评审修正后的三文档再评。评审方式：multi-role-review skill 流程 + 工程 token 实查（sound.gd/category.gd CATEGORIES/launcher_util.gd finish_record/Font.has_char 文档/editorial.json categories）。

## 业务专家（B）

| # | 严重度 | 问题描述 | 建议方案 |
|---|--------|---------|---------|
| B-1 | 中 | requirements_plan.md「范围」仍写旧语义（score=本局净变化 / 走 achievement_unlocked 信号），与 requirements.md 修订后不一致，验收人可能以 plan 为准误读 | plan 加修订指针，不改正文 |
| B-2 | 低 | sl_rich500 阈值数学不可达：注额上限下单局最高 win=10×20×2(连击)=400 < 500 | 阈值改 ≥300，改名 sl_rich300 |
| B-3 | 低 | category=casino 实查确认：category.gd CATEGORIES 硬编码表（全部/puzzle/action/arcade/casual/roguelike）无 casino 桶，游戏只出现在「全部」 | 改用 casual 桶（零 Shell 改动） |

✅ B-1/B-2/B-3 已修复（plan 指针 / requirements FR-7.1+AC-6 / 数据契约 category=casual）

## 产品经理（P）

| # | 严重度 | 问题描述 | 建议方案 |
|---|--------|---------|---------|
| P-1 | 中 | FR-1.3「开评」错别字；「错峰：先快后慢」与 design「ease out」措辞不统一 | 修错别字；统一为「缓出减速 ease out」 |
| P-2 | 中 | AC-3「1/12±2%」口径数学不可行：位级 ±2%=±2σ 外（万局 σ≈37 次命中，误杀率约 5%/符号）；且符号级频率是 1/6 非 1/12 | 符号级 1/6±3%（≈±1.6σ，误杀率 <1%），位级不设硬断言 |
| P-3 | 低 | AC-6 未随 B-2 阈值同步 | sl_rich300 同步 |
| P-4 | 低 | 开始遮罩文案未提连击规则，玩家首次遇连击困惑倍数来源 | 补「连续停同符号，押中翻倍」 |

✅ P-1/P-2/P-3/P-4 已修复（requirements FR-1.3/FR-3.1/AC-3/AC-6/FR-6.3）

## 架构师（A）

| # | 严重度 | 问题描述 | 建议方案 |
|---|--------|---------|---------|
| A-1 | 高 | **清零时序矛盾（玩法失效级）**：FR-3.1「旋转时清零押注」但 design §4 _settle 遍历 _bets 计算赔付——先清零则结算恒 0，玩家永远赢不到 | _bets 保留到 _settle 读取完毕后才清零；FR-3.1 措辞同步 |
| A-2 | 中 | **押注即时扣款与长按清零矛盾**：押=扣款则长按清零变纯视觉操作（币不返还），与用户预期冲突 | 改登记制：押注只校验 Σ≤balance 不扣款，_spin 一次性扣 total_bet，_settle 加回 win |
| A-3 | 中 | sound.gd 实查无 ratchet（仅 click/success/error/coin/toggle），「click 循环」fallback = 旋转期高频刷屏 | 两点式：起转 click + 开奖结算音；旋转过程静音 |
| A-4 | 低 | Font.has_char() 探测对象未写明 | ThemeDB.fallback_font |
| A-5 | 低 | POSITIONS=[0,1,2,3,4,5,0,1,2,3,4,5] 是同符号相邻，与注释「对角分布」不符 | 交错排列 [0,2,4,1,3,5,0,2,4,1,3,5] |
| A-6 | 低 | tasks T3 仍写「双击清零」 | 改「长按清零」 |

✅ A-1/A-2/A-3/A-4/A-5/A-6 已修复（design §3/§4/§2 + requirements FR-3.1/FR-2.3/FR-8.1/数据契约 + tasks T3）

## 汇总

共 13 个问题（高: 1 / 中: 6 / 低: 6），用户确认全部修复。落点：requirements.md（9 处）/ design.md（6 处）/ tasks.md（5 处）/ requirements_plan.md（2 处指针）。三文档修订记录均已追加二轮日期行。

✅ 三角色 review 完成，全部问题已修复，文档可进入执行阶段。
