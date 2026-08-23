# NOVA ARCADE 新星街机盒 — 游戏盒子应用整体设计

> 目标：一个 Godot 4.5 移动端"游戏盒子"App，内含多个小游戏，支持列表/分类/搜索/推荐/付费/评价。
> TETRA NOVA 作为首发旗舰游戏内置。视觉延续霓虹街机风格。
>
> **下游详细设计**：客户端模块与交互 → `nova-arcade-client-design.md`；
> 服务端（M3+ 启用）→ `nova-arcade-server-design.md`；
> **游戏运行时（PCK/HTML/街机ROM 三运行时）→ `nova-arcade-runtime-design.md`**（PoC 见 `nova-arcade-poc/`）。
> **交互原型**：`nova-arcade-prototype.html`（浏览器直接打开，可点击走完全部流程）。
> 内置**双主题**：霓虹·星穹 / 青瓷·素笺，首页右上角 🎨 或 设置→外观风格 可即时切换（记住选择）。
> 两套主题共用一套交互与数据结构，对应 Godot 侧 Theme 资源切换方案。

---

## 0. 核心决策总览

| 决策点 | 方案 | 理由 |
|---|---|---|
| 架构 | 本地优先（Local-first），后端分期接入 | MVP 零后端可上线；评价/推荐/支付三期再上云 |
| 游戏接入 | 统一 `Runner` 接口（PCK/HTML/Arcade 三运行时）+ 内置编译 + 动态下载 | Godot 原生支持 `load_resource_pack()`；街机最新定稿见 `projects/nova_arcade/mame-godot-plugin/DESIGN.md` v2（独立 native MAME4droid Activity，以 v2 为准；旧 myosd 直嵌表述待调整），商店可后装游戏 |
| 导航 | 竖屏 + 底部 Tab（首页/分类/搜索/我的） | 单手可达，移动端标准范式 |
| 变现 | 买断 + 试玩解锁 + 激励视频广告，三档并行 | 休闲小游戏最优解，不强制账号 |
| 评价 | 阶段1 仅本地；阶段2 云端（游玩≥600s才可评） | 防刷 + 降低冷启动合规复杂度 |

---

## 1. 产品架构

```
┌─────────────────────────────────────────────┐
│              NOVA ARCADE (Shell)             │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐        │
│  │ 首页推荐 │ │ 分类浏览 │ │  搜索   │        │
│  └─────────┘ └─────────┘ └─────────┘        │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐        │
│  │ 游戏详情 │ │ 评价中心 │ │ 我的/库 │        │
│  └─────────┘ └─────────┘ └─────────┘        │
│  通用服务：存档 / 成就 / 支付 / 事件总线 / 设置 │
├─────────────────────────────────────────────┤
│              GameModule 统一接口              │
├──────┬──────┬──────┬──────┬─────────────────┤
│tetra │game2 │game3 │ ...  │ DLC(PCK 后装)    │
│_nova │      │      │      │                 │
└──────┴──────┴──────┴──────┴─────────────────┘
```

### 1.1 GameModule 接口（Godot 落地）

```gdscript
# game_module.gd — 每个游戏必须实现
class_name GameModule extends Node

func get_meta() -> Dictionary:      # id/名称/图标/标签/价格模式/版本
    return {}

func boot(ctx: Dictionary) -> void:    # ctx: {save_dir, trial_mode, owned, best}
    pass

func pause_game() -> void:             # 系统来电/通知时由 Shell 调用
    pass

func resume_game() -> void:            # 恢复运行
    pass

signal quit_requested(result: Dictionary)  # 游戏主动退出 → 结算(result: score/playtime/achievements)
```

- **内置游戏**：与 Shell 一起编译，进 `games/<id>/`；
- **DLC 游戏**：独立打出 `.pck`，商店页显示"下载"，下载后
  `ProjectSettings.load_resource_pack(path)` + 注册 meta 进目录。
  主包体积只含 Shell + 1~2 个游戏，其余全走 DLC。

### 1.2 目录结构（建议新工程 `nova-arcade/`，TETRA NOVA 迁入）

```
nova-arcade/
  shell/            # 外壳：导航/页面/通用服务
    scenes/  scripts/  theme/
  games/
    tetra_nova/     # 现有工程迁入，实现 GameModule
    ...
  services/         # save / payment / review / recommend / event_bus
  dlcs/             # 下载的 .pck 存放 (user://dlcs)
```

---

## 2. 数据模型

```yaml
GameMeta:            # 游戏元数据（本地内置表 + 云端可更新）
  id: tetra_nova
  title: TETRA NOVA / 新星方阵
  category: puzzle        # 消除/益智/动作/街机/休闲/roguelike
  tags: [tetris, roguelike, neon, offline]
  price_model: free|paid|trial|iap|ad
  price: 6                # 买断价（分）
  trial_limit: {plays: 3} # 试玩制：3次
  version: 1.0.0
  size_kb: 4200           # DLC 用
  icon / screenshots[]: res://...
  min_playtime_to_review: 600   # 秒（10分钟）

  # 视口与方向（可选；省略则继承 Shell 默认值 720×1560 竖屏）：
  viewport_size: [720, 1560]    # 游戏的逻辑设计分辨率 [w, h]
  orientation: "portrait"       # "portrait" | "landscape"
  # Launcher 在转场遮罩完全盖住屏幕后切换，退出时自动恢复 Shell 值。
  # 横屏游戏填 [1920, 1080] + "landscape"；DLC 开发者按自己习惯的基准填即可。

  # Arcade 特有字段（runtime="arcade" 时）：
  runtime: "pck"               # "pck" | "html" | "arcade"
  core: ""                     # 依赖的核心包名（arcade="myosd-0.288"）
  roms: [{
    name: "pacman",            # MAME 驱动名
    title: "PAC-MAN",
    parent: "",                # 父驱动名（clone 时填 parent）
    bios: "",                  # BIOS 依赖名
    zips: ["pacman.zip"],      # 该驱动需要的 rom zip 列表
    controllerType: "virtualkey",
    buttonNumber: 4,
    cansl: true,
    buttonConfig: {...},
    hiscore: "plugin"
  }]

PlayRecord:          # 本地游玩记录（驱动推荐/继续游戏）
  game_id / last_played / total_playtime / sessions / best_score

DailyTask:           # 每日任务（本地，按自然日重置；详细字段见 client-design §4.2）
  date / tasks[{id, done}] / reward_claimed

RatingReview:
  game_id / user_id / stars(1-5) / text / playtime_at_review
  created_at / likes / status: local|pending|published

UserProfile:
  guest_id / nickname / avatar_idx
  owned_games[] / orders[] / wallet
  achievements{} / settings{}
```

---

## 3. 页面设计（竖屏，720×1280 画布，详见 `nova-arcade-client-design.md` §8）

### 3.1 首页（推荐流）

```
┌──────────────────────────┐
│ NOVA ARCADE      [🔍搜索] │ ← 顶栏：Logo + 搜索入口 + 🎨主题切换
│ ┌──────────────────────┐ │
│ │   Banner 轮播(3张)    │ │ ← 编辑推荐/活动/新游；4s自动轮播
│ │ ●○○ (激活点=14px宽)   │ │
│ └──────────────────────┘ │
│ ▶ 继续游戏               │ ← 最近玩过；长按卡片→"从记录移除"
│ [TETRA NOVA▾] [2048▾] →  │
│ ▶ 为你推荐               │ ← 个性化横滑卡
│ [卡1] [卡2] [卡3] →      │
│ ▶ 热门榜 Top10           │ ← 综合/新游/好评 三个子Tab
│ 1 TETRA NOVA  ★4.8 12w人 │
│ ▶ 每日任务  ●●○ 今日2/3   │
├──────────────────────────┤
│ [🛸首页] [▦分类] [🔍] [👾我的] │ ← 底部Tab；选中项显示16×2.5px底部高亮条
└──────────────────────────┘
```

卡片信息：图标（动态渐变背景，霓虹主题用 `linear-gradient(150deg, col, #0a1030)`，青瓷用 `linear-gradient(150deg, colE, #f4f1ea)`）/ 名称 / ★评分 / 付费模式 / 状态角标（已装▶、未装⬇、试玩N）。

### 3.2 分类页
- 左侧竖列：全部/消除/益智/动作/街机/休闲/Roguelike；选中项左边缘显示3px强调色竖条
- 顶部筛选 chips："全部"**互斥**（选中时清空其余），其余**多选叠加**；`免费` `付费` `离线可玩` `评分≥4.0` `新游`
- 右侧 2 列网格卡片 + 顶部排序下拉：`最热`（默认） `最新` `评分`；下拉 M2 实现，M1 默认最热
- 筛选无结果时：空态插画 + "该筛选下暂无游戏" + 内联"清除筛选"按钮（`reset to all`）

### 3.3 搜索页
```
┌──────────────────────────┐
│ [← 输入关键词…    取消]   │
│ 🔥 热搜：俄罗斯方块 2048 … │
│ 🕘 历史：[tetris] [rogue] │
│ ── 实时结果（输入即出）── │
│ TETRA NOVA ★4.8 [消除/RL] │
│ 2048霓虹版   ★4.6 [益智]  │
└──────────────────────────┘
```
匹配优先级：标题前缀 > 标题包含 > 别名(拼音全/首字母) > 标签 > 开发者名。
无结果时展示"猜你喜欢"（按用户标签画像）。游戏量 <1k 时本地内存索引即可，无需搜索引擎。

### 3.4 游戏详情页
```
┌──────────────────────────┐
│ [←]  截图轮播(可横滑)      │
│ ┌──┐ TETRA NOVA          │
│ │图标│ 新星方阵 v1.0       │
│ └──┘ ★4.8 (2.3万评价) 12万人玩过 │
│ [标签] 消除·Roguelike·离线      │
│ ═══════════ 概况 ═══════════ │
│ 简介/更新日志                 │
│ ── 评分分布 ▇▇▇▇▂▂ 好评91% ── │
│ 精选评论 2条 … [全部评价→]     │
│ ── 相关推荐 ──               │
│ [▶ 免费开玩]  或  [¥6 购买]   │
│        或  [试玩 (剩2次)]     │
└──────────────────────────┘
```
主按钮状态机：`免费开玩` → `继续` / `试玩(n)` → `购买¥6` → `已拥有▶` / DLC：`下载(4MB)` → `开玩`。

### 3.5 评价中心
- 列表排序：最有用（点赞数）/ 最新 / 好评优先 / 差评优先
- 写评价：5星选择 + 文字(≤500字)；**游玩时长 <600s 按钮置灰**
- 一人一游戏一条，可修改（修改刷新时间）
- 点赞/举报；云端阶段加敏感词过滤 + 审核队列
- 评分展示双口径：**近30天** 与 **全部**（Steam式），防刷分翻盘

### 3.6 我的
- 头像/昵称（本地随机生成，上云后可改）
- 我的游戏库（已拥有/试玩中/最近）、游玩时长统计图
- 成就墙（跨游戏成就点数总览）
- 钱包/订单/恢复购买
- 设置：音效/震动/语言/深色/清理缓存/关于

---

## 4. 搜索实现（本地索引）

```gdscript
# services/search_index.gd
构建: games[] → 每个游戏产出 keys = [title, title_lower, aliases[],
      pinyin_full, pinyin_initials, tags[], dev]
查询: 输入串 tokenize → 对每个 key 计算得分
      前缀命中 +100 / 包含 +40 / 拼音首字母 +30 / 标签 +15
      → 总分排序，取前 20
热词: 本地统计输入→点击，云端阶段下发热搜榜
```

---

## 5. 推荐系统（三阶段演进）

### 阶段1：本地规则推荐（无后端）
```
热度分 = 0.5*norm(游玩人数) + 0.3*norm(评分*评人数) + 0.2*norm(新游加权)
为你推荐 = 用户玩过游戏的标签并集 → 同标签游戏按热度分排序
          （排除已拥有/最近曝光过的，7日频控）
继续游戏 = PlayRecord.last_played 倒序 top5
冷启动(无记录) = 精选榜单 + 分类热门
```

### 阶段2：云端个性化
```
召回: ItemCF 协同过滤(玩过A也玩B) + 标签向量 + 热度兜底
排序: 简单 CTR 模型(特征: 标签匹配/价格档/历史点击率)
重排: 多样性(同分类≤2) / 新游扶持 / 已玩过滤 / 曝光频控
指标: 曝光→点击率→首游戏时长→7日留存
```

### 阶段3：场景化推荐
- 玩完某局退出时："换一个？"（同类型快速切换，转化最高的场景）
- 断网/排队场景推荐离线小游戏

---

## 6. 付费设计

### 6.1 付费模式（每游戏四选一）
| 模式 | 说明 | 示例 |
|---|---|---|
| free | 完全免费 | 引流游戏 |
| ad | 免费+激励视频广告（复活/双倍分） | 休闲游戏 |
| trial | 试玩 N 次后付费解锁（转化最好的模式） | 前3次免费 |
| paid | 直接买断 ¥1/¥3/¥6/¥12 | 旗舰游戏 |
| iap | 游戏免费 + 去广告/皮肤/扩展关卡 | 长线游戏 |

### 6.2 支付通道
- **海外**：Google Play Billing（Godot 官方 `InAppStore` Android 插件）
- **国内**：华为/小米/OPPO/vivo/应用宝 各渠道 SDK，做一层
  `PaymentProvider` 抽象接口，按渠道包接入
- **流程**：
```
点击购买 → 创建本地订单(pending) → 调起渠道支付 → 回调成功
→ 票据本地暂存 → 服务端验签(阶段3) / 本地校验(阶段1)
→ 发放权益(owned_games) → 订单完成 → 恢复购买可重放
```
- 关键保障：**断网重试队列**（票据未验证的订单本地持久化，启动重试）、恢复购买、退款回收权益（云端阶段）。

### 6.3 广告
- 激励视频为主（玩家主动看，体验最好）：复活/道具/试玩+1次
- 穿山甲/AdMob，Godot Android 插件形式接入
- 频控：每局最多 1 次，每日上限

---

## 7. 评价系统细节

```
提交: 星级必选 + 文字可选 → 本地立即可见(local状态)
     ↑ 游玩时长门槛: <600s 按钮置灰
云端阶段: 敏感词过滤 → 机审(pending) → 人工审核队列 → published
聚合: game.rating_avg / count / last30d_avg  由服务端定时算好下发
排序: 有用(点赞) > 最新；点赞需登录(阶段3)
防刷: 设备指纹+账号 双维度限频；同IP段批量差评自动折叠
```

---

## 8. 通用服务（Shell 层）

| 服务 | 职责 |
|---|---|
| EventBus | 全局信号：game_launched/finished/rated/purchased… |
| SaveService | `user://saves/<game_id>/` 按游戏隔离；盒子级 profile 单独存 |
| AchievementService | 跨游戏成就定义/解锁/展示，游戏内通过 EventBus 上报 |
| PaymentService | 订单/票据/权益/恢复购买 |
| ReviewService | 本地评价存储 + 云端同步(阶段3) |
| RecommendService | 上述三阶段推荐 |
| AnalyticsService | 埋点：曝光/点击/开玩/时长/付费（本地留存，云端上报） |
| LauncherService | 游戏生命周期：launch → boot(ctx) → 暂停Shell → 退出恢复Shell+上报 |

**游戏退出协议**：游戏调 `ctx.quit(result)` → Shell 弹"成绩结算卡"（分数/成就/评价引导/推荐下一款）→ 返回首页。评价引导时机 = 第2次游玩结束时，转化最高且不打扰。

---

## 9. 后端设计（阶段3 才需要）

```
技术选型(轻量): Supabase/Firebase(BaaS省事) 或 Go+Postgres+Redis
接口:
  POST /reviews            提交评价
  GET  /games/:id/reviews  分页评论
  POST /payments/verify    渠道票据验签
  GET  /recommend/:uid     推荐列表
  GET  /games/catalog      元数据+榜单(带版本号增量)
  POST /sync/save          云存档
账号: 游客(设备ID) → 可选绑定 手机号/Google
合规: 隐私政策+未成年人防沉迷(国内上架必需)
```

---

## 10. 里程碑

| 阶段 | 内容 | 后端 | 工作量级 |
|---|---|---|---|
| **M1 盒子骨架** | Shell+底部Tab+列表/分类/搜索/详情（本地数据）+ TETRA NOVA 迁入接入 GameModule + 本地评价 + 继续游戏 | ✗ | 2-3天 |
| **M2 内容扩充** | +4~6 个小游戏(2048/贪吃蛇/打砖块/扫雷/弹幕…) + 成就墙 + 每日任务 + 本地推荐画像 + **Arcade PoC Android 真机验证**（跑通 1 个 ROM，按 mame-godot-plugin/DESIGN.md v2 验证 MameRuntime 拉起/收回与帧率/音频，锁定进程语义） | ✗ | 每游戏0.5-1天；Arcade PoC 1-2天 |
| **M3 云端化** | 账号/云评价/云推荐/排行榜/云存档（PCK/HTML 游戏的 ≤64KB 存档 blob） | ✓ BaaS | 3-5天 |
| **M4 商业化** | 支付SDK(渠道/GooglePlay) + 试玩付费 + 广告SDK | ✓ | 3-5天 |
| **M5 DLC 商店** | PCK 打包流水线 + 下载安装 + 增量更新；街机核心包 + rom zip 集/说明文件下发（runtime §3.4）+ ArcadeRunner 真机验证（多 ROM/parent/clone/BIOS）；支持 parent/clone/BIOS 多 zip 下载与校验 | ✓ CDN | 2-3天 |

> M1+M2 就是完整可玩、可分发的免费盒子；付费与评论完整形态从 M3 起。

---

## 11. 已知风险与对策

| 风险 | 对策 |
|---|---|
| Godot mobile 渲染器 2D glow 后处理破坏实心矩形(已踩坑) | 盒子统一禁用 hdr_2d/glow，霓虹用分层半透明绘制 |
| 渠道支付 SDK 接入碎片化 | PaymentProvider 抽象 + 先只上 GooglePlay/单渠道验证闭环 |
| PCK DLC 被篡改/盗版 | 阶段5再加签名校验；前期游戏免费不痛 |
| 评价冷启动空页 | 预置编辑评价 + "成为第一个评价的人"引导 |
| 主包膨胀 | 主包只装 Shell+1游戏，其余全 DLC |
| 街机核心体积大 / ROM 版权 | 核心按 ABI 运行时下载（v2：`libMAME4droid.so` 约 74MB，见 mame-godot-plugin/DESIGN.md）；ROM 只走授权下发或用户自导入（标注），详见 runtime-design §3.4/§3.6（待按 v2 调整） |
| 多 zip romset（parent/clone/BIOS）/ 校验复杂 | meta.json 每个 ROM 声明 `parent`/`bios`/`zips[]`；服务端每个 zip 独立签名；客户端逐文件校验后 MAME 原生 romset 校验兜底（沿 parent/BIOS 链）；详见 runtime-design §3.4 |
