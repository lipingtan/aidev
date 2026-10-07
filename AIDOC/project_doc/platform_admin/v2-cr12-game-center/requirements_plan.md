# 需求计划：mjarch 游戏中心（服务端游戏列表 + App 动态 ROM 下载 + 列表 UI 改版）

> **进度锚点（2026-10-03 收工）**：requirements_plan.md 已完成，8 个澄清问题全部有 [Answer-N]（用户已批）。**下一阶段 = requirements.md 正式需求文档（待用户确认本 plan 后动笔）**。UI 原型在 `C:\data\run\hermes\proto\mjarch_gamelist_ui.html`（浅色系、卡片不重叠、底栏四 Tab 已定稿）。同日完成：mjarch so 重编（=jdgamespace 全 ROM 范围，arm64 373MB/arm32 257MB）+ APK 313MB 重装 + App 改名「街机时代」。根因沉淀：NDK 路径必须 `C:/` 格式（见 mame-android-workbench skill pitfalls）。

## 需求理解

**目标：** mjarch App 当前把全部游戏资源打进 files.zip 随 APK 分发（APK 已 313MB），游戏更新必须重发版。本需求改为「服务端管理游戏列表 + App 动态拉取列表 + ROM 按需下载」：运营在管理端增删游戏与公告/Banner，用户端游戏大厅实时呈现，点未下载的游戏进入详情页下载，下载完即玩。配套重设计 mjarch 游戏列表 UI（公告栏、Banner 轮播、分类 Tab、小卡片、底栏四 Tab：游戏大厅/我的/下载管理/设置）。

**范围：**
- **服务端（platform_admin）**：新增 game 内容插件——游戏列表管理（CRUD：rom 名、中文名、分类、封面、体积、免费/VIP 标记、parent 依赖清单、md5）、公告管理、Banner 管理；C 端接口（游戏列表/公告/Banner 拉取 + 登录态下载接口：HTTP Range 断点续传 + md5 校验）
- **App 端（mjarch flavor）**：
  - 游戏大厅 UI：顶栏会员状态 + 公告栏 + Banner 轮播 + 分类 Tab + 双列小卡片网格
  - **游戏详情页**（点卡片进入）：大按钮（开始游戏/下载）、暂停/继续/删除下载、收藏入口；（评论/评价本期不做，预留位置）
  - 底栏四 Tab：游戏大厅 / 我的（会员状态+收藏夹） / 下载管理 / 设置
  - **收藏夹**：收藏/取消收藏游戏，收藏列表展示与快捷进入（替代原「下载管理」聚合位）
  - **设置页增强**：在现有「关于」类内容基础上，增加**已下载游戏管理**——列出已下载游戏（含占用空间），支持**清除单个游戏的包和文件夹**（rom zip + 解压后目录 + cfg/nvram 一并清），并做下载目录总占用统计
  - rompath 多路径（下载目录 + 内置目录并存）；gamelist 动态合并（files.zip 内置列表为兜底 + 服务端列表增量）
- **管理端前端（dev-web-admin）**：游戏管理页、公告管理页、Banner 管理页

**预期效果：**
- 运营在管理端上架新游戏 → 用户 App 下拉刷新即见，进详情页下载 → 完成后点「开始游戏」直接玩，全程无需发版
- 公告/Banner 运营位可随时更新
- 游戏按内容类型分类浏览（管理端可维护分类）
- 免费游戏登录即可下载直玩；VIP 游戏未开通会员时详情页引导开通（复用现有会员体系）
- 设置页可查看并清理已下载游戏占用的空间

**不在范围（本期明确不做）：**
- 游戏评论/评价功能（详情页仅预留 UI 位置）
- 其他 flavor（mjguoguan1/2、mjarch3 等）的改造——files.zip 机制保持不变
- 游戏封面图上传存储体系（本期封面 URL 由管理端填外链；管理端仅存 URL 字段）
- ROM 仓库的 CDN/OSS 分发（本期服务端本地磁盘直出，接口留抽象）
- ROM 版本更新机制（见 [Answer-5]，本期不做增量更新）
- iOS/其他平台客户端

---

## 假设列表

- [假设-1] ROM 文件仓库放服务端本地磁盘（platform_admin 配置目录下 `roms/`），服务端提供下载接口；单 ROM zip 几 MB~几十 MB，总量几十 GB
- [假设-2] 下载鉴权：所有 ROM 下载均需登录 token（免费/VIP 皆然，防爬虫批量拉库）；免费/VIP 只影响「能不能玩」，不影响「能不能下」
- [假设-3] 分类是服务端管理的表数据（管理端增删改+排序），App 动态渲染 Tab；初始分类集运营导入时指定
- [假设-4] App 端下载 ROM 存 `getExternalFilesDir()/roms/`；myosd `-rompath` 已支持 `;` 多路径（mjguoguan2 修复已具备），内置解压目录与下载目录并存搜索
- [假设-5] 游戏列表全量下发（gzip 后几百 KB）；不做分页
- [假设-6] 现有会员链路（登录/激活//status/TimeKeeper）零改动，本需求只消费 token 与会员态

---

## 澄清问题

- [Question-1] APK 体积策略：现有 files.zip（含 12633 个 ROM）如何处理？
  [Answer-1]
  **APK 保留现状 ROM 打包不变**（files.zip 照旧全量内置）。本期动态下载机制的直接受益者是新游戏上架（不用发版），而非 APK 瘦身；APK 瘦身留待后续版本单独评估。

- [Question-2] 免费游戏 ROM 下载是否需要登录？
  [Answer-2]
  **需要登录**。免费/VIP 均凭登录 token 下载（防爬虫），与假设-2 一致；未登录用户点下载引导先登录。

- [Question-3] VIP 游戏判定粒度？
  [Answer-3]
  **整游 VIP 二元制**（复用现有 MembershipAccessControl 判定点），不做试玩计时。

- [Question-4] 分类体系？
  [Answer-4]
  同意推荐：**管理端可维护分类表**（增删改+排序），App 端动态渲染。

- [Question-5] 已下载游戏的 ROM 更新感知？
  [Answer-5]
  **本期不做 ROM 更新机制**。服务端列表仍下发 md5 字段（数据结构预留），App 暂不比对、不显示更新角标；后续要做时数据链路已通。

- [Question-6] 底栏「下载管理」是否保留为独立 Tab？
  [Answer-6]
  **改为「我的收藏夹」**：底栏四 Tab = 游戏大厅 / 我的 / 收藏夹 / 设置。下载队列相关操作收敛到**游戏详情页**（暂停/继续/删除），不设独立下载管理页。

- [Question-7]（新增）详情页能力边界？
  [Answer-7]
  详情页包含：游戏封面大图 + 基本信息（分类/体积/免费VIP）+ **主按钮区**（状态机驱动：未下载=「下载」、下载中=「暂停」、暂停=「继续」、已下载免费=「开始游戏」大按钮、VIP未开通=「开通会员」引导）+ 次操作（删除下载、收藏/取消收藏）+ **评论区预留位置**（本期不实现评论，仅留「敬请期待」占位）。

- [Question-8]（新增）设置页清理能力的边界？
  [Answer-8]
  设置页新增「已下载游戏管理」：列表显示已下载游戏（名称+占用空间），支持**逐个清除**（删除 rom zip + 解压目录 + 该游戏 cfg/nvram，一次删干净），显示下载目录总占用。不做「一键全部清理」（误删风险，逐个操作足够）。

---

## 非功能需求建议

- **性能：** 列表接口 gzip；下载接口 HTTP Range 断点续传；单设备下载并发限 1-2
- **安全：** 下载接口挂双 token 鉴权；管理端 CRUD 挂 `game:*` 权限点（新增权限 seed 复用 RBAC）
- **存储：** ROM 仓库目录可配置；download_log 落库（本期不做报表）
- **App 兼容：** minSdk 21 不变；okhttp4 已有依赖；UI 改版仅动 mjarch flavor，baseuilib 共享改动必须 instanceof 防护（硬规则）
- **降级：** 列表拉取失败回退内置 files.zip 列表；下载中断点续传，App 重启恢复队列

## 影响范围预判

**新增（服务端 platform_admin）：**
- `app/game/` 插件：model（game/category/notice/banner/download_log）+ service + handler + plugin.go（参照 license 插件结构）
- ROM 仓库目录 + 下载 handler（Range/md5）
- 权限 seed：`game:games:*`、`game:category:*`、`game:notice:*`、`game:banner:*`
- C 端接口：`GET /public/game/list|categories|notices|banners`、`GET /public/game/download/:rom`（登录态）
- 收藏：App 本地存储（SharedPreferences/SQLite），**服务端本期不做收藏同步接口**

**新增（App mjarch）：**
- `app/src/mjarch/java/.../gamecenter/`：GameListApi、DownloadManager（okhttp 队列+暂停/续传）、GameRepository（列表合并/下载记录/收藏持久化）
- 四 Tab 容器 + 大厅 Fragment（GameCardView/CategoryTabBar/NoticeBar/BannerViewPager）
- **GameDetailActivity**（详情页：状态机按钮区 + 删除下载 + 收藏 + 评论占位）
- FavoritesFragment（收藏夹 Tab）
- 设置页新增「已下载游戏管理」模块（列表+逐个清除+总占用统计）

**修改：**
- MjarchGameListActivity → 改造为四 Tab 容器（旧列表逻辑并入大厅 Fragment）
- myosd rompath 传参（Java 侧追加下载目录，机制已存在）
- 设置 Fragment（mjarch 侧）加已下载管理入口

**不触碰：**
- 会员体系（license 插件）全链路、TimeKeeper、双 token——纯消费方
- 其他 flavor 的任何代码与资源
- MAME 内核（so 不需重编；rompath 机制已具备）
- APK 内 files.zip（保留现状）
