# 需求文档：CR-8 幸运轮盘押注机（转圈押币）

> 依据 requirements_plan.md（2026-08-31 用户确认定稿）。
> 工程路径：`projects/nova_arcade/nova-arcade/games/slot_machine/`。
> 前置：GameModule 协议（core/game_module.gd）、结算卡、成就引擎、editorial banner 均已就位，Shell 零改动。

---

## 1. FR 清单

### FR-1 面板与符号
- FR-1.1 正方形面板（边长 = 视口宽 - 左右边距），边缘均匀分布 12 个符号位，每符号 1 个居中 Label（符号用 Unicode 字符 + ThemeTokens 颜色，不引外部素材）
- FR-1.2 符号构成：6 种 × 各 2 位，对称交错分布（樱桃/柠檬/铃铛/星/钻石/七）
- FR-1.3 面板中央区：余额显示、开奖结果提示（普通/连击×2）、旋转按钮的容器
- FR-1.4 指针指示器：面板顶部固定三角标记，标记开奖位

### FR-2 押注
- FR-2.1 底部押注区：6 个符号押注位（横向一排，等宽），每个位显示 符号+当前押注额
- FR-2.2 注额选择：1/5/10 三档切换（分段按钮组），全局单选影响所有符号
- FR-2.3 押注操作：点押注位 → 当前注额 +1 份（可重复叠加）；长按 0.4s → 清零该符号押注（A-2：押注为登记制不入账，仅校验总押 ≤ 余额；旋转时一次性扣款，清零天然无退款问题）
- FR-2.4 押注约束：总押注（登记额）≤ 当前余额；余额不足时该次点击无效 + Toast 提示
- FR-2.5 至少押 1 注才能旋转；未押点旋转 → 按钮置灰

### FR-3 滚动与开奖
- FR-3.1 点「旋转」：一次性扣除总押注 → 转圈动画（缓出减速 ease out，~2.5s）→ 结算完成后清零押注表（A-1：清零须在结算读取之后）
- FR-3.2 结果先抽后演：动画启动前先按均匀分布抽中停位（1/12），动画终点对准该位
- FR-3.3 滚动期间：押注区与旋转按钮锁定（mouse_filter 忽略 + 置灰），防连点竞态
- FR-3.4 开奖：停在符号 S → 结算（FR-4）→ 中央显示结果 + 音效

### FR-4 赔付与连环奖
- FR-4.1 赔付表（const，test 共享）：樱桃×5 / 柠檬×6 / 铃铛×8 / 星×10 / 钻石×14 / 七×20
- FR-4.2 基础结算：押中符号的每一份注 → 押额×赔率 返还；未押中符号的注没收
- FR-4.3 连环奖：上一局停符号 == 本局停符号 → 押中该符号的赔付整体 ×2；中央提示「连击！×2」
- FR-4.4 streak 状态：模块内存变量（会话内有效，不落盘、不随 reset_run 重置——reset_run 仅重开一局，余额与 streak 延续）；quit 后模块释放自然清零
- FR-4.5 RTP 目标：基础 ≈88%（±3% design 微调空间）；含连环奖实际落点 ≤95%

### FR-5 虚拟币账本
- FR-5.1 余额初始 100；存 `ctx.save_dir/save.cfg`（key: balance），每局结算后立即强写（对齐 override §2 best 类强写）
- FR-5.2 破产救济：结算后余额 < 1 → 中央出现「领取救济 50 币」按钮，点击 +50 强写；无冷却无上限
- FR-5.3 余额上限：999,999（溢出不加，防数值溢出显示）

### FR-6 结算卡对接
- FR-6.1 quit_requested({score, playtime, achievements, extra})：
  - score = 会话累计净变化（当前余额 - 初始 100，可负；跨局累积，best 有意义）；单局最大赢额入 extra.biggest_win
  - playtime = 本模块会话秒数
  - extra = {balance, biggest_win, streak_symbol, streak_count}
- FR-6.2 best 记录：DB best = max(best, score)，负数不影响
- FR-6.2a extra 为数据通道，结算卡不展示 extra 字段（Shell 零改动）
- FR-6.3 回菜单按钮：adapter 侧常驻（FOCUS_NONE，右上角，对齐四游戏既有约定）；开始遮罩（标题+玩法说明+开始按钮，boot 不自动开局；说明含连击规则「连续停同符号，押中翻倍」P-4）

### FR-7 成就
- FR-7.1 ×2 条目（meta.achievements）：
  - sl_jackpot「头奖手气」：单局押中「七」并赔付（desc: 押中大奖符号七）
  - sl_rich300「一夜暴富」：单局 win ≥300（B-2：原 500 数学不可达——注额上限下最高单局 win=10×20×2=400，正常单注七=200，300 可达且需高赔率/连击才触发）
- FR-7.2 解锁判定即时（结算时记入会话成就数组，UI 可即时提示）；DB 落盘与解锁广播随 quit_requested 的 achievements 数组批量交付（AchievementEngine 在 game_finished 时统一处理）

### FR-8 音效
- FR-8.1 复用 core/sound.gd（实查可用：click/success/error/coin/toggle）：押注 click、起转 click、开奖中奖 coin+success、输 error；旋转过程不循环音效（A-3：两点式，防高频刷屏）

## 2. 数据契约

- meta.json：id=slot_machine, category=casual（B-3：复用分类页既有「休闲」桶，Shell 零改动；casino 桶不存在于 category.gd CATEGORIES）, price_model=free, runtime=pck, viewport 720×1560 portrait, pinyin [xingyunlunpan, xl], aliases [轮盘, 幸运轮盘, 老虎机]
- 赔付表/符号配置：`src/slot_config.gd`（class_name SlotConfig，const SYMBOLS/ODDS/COLOR_KEYS），src 与 test 共享
- 存档 key：save.cfg → balance:int

## 3. AC（验收标准）

- AC-1 headless 生命周期：boot → start(点开始) → 押注 → 旋转 → 开奖 → quit_to_shell → 结算卡字段齐全（≥12 断言）
- AC-2 赔付正确性：给定停位与押注组合（含连击 ×2 路径），结算值逐例断言（≥10 例）
- AC-3 统计断言：模拟 10000 局（固定 seed），符号级命中频率 ∈ 1/6±3%（P-2：约 ±1.6σ，误杀率 <1%；位级不设硬断言）、基础 RTP ∈ [85%,92%]、6 符号全部有命中、连击发生次数 > 0
- AC-4 存档：结算后 balance 落盘，重 boot 读回一致；救济领取后余额 +50 落盘
- AC-5 破产流程：余额清空 → 救济按钮出现 → 领取 → 可继续押注
- AC-6 成就：押中「七」解锁 sl_jackpot；单局 win ≥300 解锁 sl_rich300（headless 直调 _settle 构造局验证）
- AC-7 GUI 截图（双主题）：开始遮罩 / 押注态 / 滚动中 / 开奖结果（含连击提示）/ 破产救济，五态无布局跳变、无元素截断
- AC-8 Shell 集成：Registry.query 命中（category/tags/pinyin）、banner 轮播出现、launch→play→quit→result_card 全链路（GUI）
- AC-9 全量回归：既有 launch/smoke/nav/四游戏测试全绿 + 无 Godot 孤儿进程

## 4. 不在范围（重申）

真实支付、币售卖、联网奖池、多面板、横屏、Shell 改动、streak 落盘持久化

---

## 修订记录

- 2026-08-31 三角色评审修正：成就交付时点/通道明确（B-1/B-2）、score 语义对齐 design（B-4）、长按清零替代双击（P-1）、extra 不展示说明（P-4）。
- 2026-08-31 二轮评审修正：押注改登记制旋转一次性扣款（A-2）、清零时序后移到结算后（A-1）、统计断言口径符号级 1/6±3%（P-2）、成就阈值 500→300 改名 sl_rich300（B-2）、category casino→casual（B-3）、音效两点式（A-3）、错别字与措辞统一（P-1）、遮罩补连击说明（P-4）、AC-6 同步（P-3）。

---

## 修订 2（2026-08-31，用户对照实物水果机图重定义形态）

用户确认改为经典水果机形态，以下覆盖本文与 12 格旋转方案冲突的条目：

- **盘面**：24 格环形（矩形环排布），图标贴图格（程序化预绘 PNG，非文字）
- **跑灯**：单灯逐格步进（非整轮旋转），快→慢减速停格，停格=预抽加权结果
- **Lucky 散花**：灯经大Lucky左邻/小Lucky右邻触发散花，飞出灯珠随机停格（大5~7/小3~4 个）；**结算 = 主停格 + 所有散花停格，逐格判定押中即倍数相乘，合计得分**（用户确认）
- **押注**：8 个可押符号（苹果×5/橙×10/柠×10/西瓜×20/番茄×5/铃铛×10/星×20/77×20），BAR25/BAR50/Lucky 不可押；登记制不变；数字键 1~8 快捷
- **结算卡对接/账本/成就协议**：不变（score=会话累计净变化）
- AC-2/AC-3 按新概率表在 T7 重算；AC-7 五态不变（含散花态截图）
