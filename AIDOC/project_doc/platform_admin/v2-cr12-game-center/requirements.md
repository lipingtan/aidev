# 需求：mjarch 游戏中心（v2-cr12）

> 版本：1.1
> 日期：2026-10-04（2026-10-05 修订：底栏 Tab 调整 + 首页顶栏改版 + 分页）
> CR 类型：跨层集成 CR（Go 后端 + Android App mjarch flavor）
> 来源：requirements_plan.md（2026-10-03 完成，8 个澄清问题全部用户批准）

---

## 背景

mjarch App（「街机时代」）当前把全部游戏资源打入 files.zip 随 APK 分发（APK 已 313MB），游戏更新必须重新发版。本需求改为「服务端管理游戏列表 + App 动态拉取 + ROM 按需下载」，配套重设计游戏大厅 UI。

**目标：**
- 运营在管理端上架新游戏 → 用户 App 下拉刷新即见，进详情页下载后直接玩，全程无需发版
- 公告/Banner 运营位可随时更新
- 游戏按内容类型分类浏览

---

## 用户故事

- 作为**运营管理员**，我希望在管理端增删游戏条目（含封面、分类、VIP 标记），以便控制 App 端游戏上下架，无需发版。
- 作为**运营管理员**，我希望维护分类、公告、Banner，以便灵活运营 App 内容。
- 作为**已登录免费用户**，我希望在游戏大厅浏览游戏列表并下载免费游戏，以便即时游玩新上架的游戏。
- 作为**VIP 会员**，我希望能下载并游玩标记为 VIP 的游戏，以便享受会员权益。
- 作为**未登录游客**，我希望能浏览游戏列表和公告/Banner，以便了解可玩的游戏内容。
- 作为**已登录用户**，我希望收藏喜欢的游戏并在收藏夹快速找到，以便管理个人游戏库。
- 作为**已登录用户**，我希望在设置页查看并清理已下载游戏占用的空间，以便管理设备存储。

---

## 功能需求

### FR-01：游戏管理（管理端）

**描述：** 管理员在管理端对游戏条目进行增删改查。

**验收标准：**
- WHEN 管理员填写 rom 名称、中文名、分类、封面图 URL、体积、VIP 标记、parent 依赖清单、md5 并提交 THEN 系统 SHALL 创建游戏条目并持久化
- WHEN 管理员编辑现有游戏条目并保存 THEN 系统 SHALL 更新对应字段，不影响其他条目
- WHEN 管理员删除游戏条目 THEN 系统 SHALL 软删除该条目，C 端列表接口不再返回该游戏
- WHEN 管理员在列表页按名称或分类搜索 THEN 系统 SHALL 返回匹配结果（支持分页）
- WHEN 管理员提交 rom 名称与已有条目重复的游戏 THEN 系统 SHALL 返回错误提示，拒绝创建

---

### FR-02：分类管理（管理端）

**描述：** 管理员维护游戏分类表，App 端 Tab 栏以服务端返回为准，不硬编码。

**验收标准：**
- WHEN 管理员新增分类（名称、排序序号）并保存 THEN 系统 SHALL 创建分类，App 端下次拉取时呈现新 Tab
- WHEN 管理员调整分类排序 THEN 系统 SHALL 保存新排序，App 端 Tab 顺序随之变化
- WHEN 管理员删除某分类 THEN 系统 SHALL 弹出确认提示「该分类下 N 个游戏将自动归入未分类」，确认后删除分类并将游戏 category_id 置为 NULL（同一事务）
- WHEN 管理员禁用分类 THEN 系统 SHALL 使该分类不出现在 C 端接口返回中

---

### FR-03：公告管理（管理端）

**描述：** 管理员发布、编辑、下线公告。

**验收标准：**
- WHEN 管理员填写标题、正文并将状态设为「已发布」保存 THEN 系统 SHALL 创建已发布公告，C 端接口立即可查到
- WHEN 管理员将公告状态改为「草稿」 THEN 系统 SHALL 使该公告不出现在 C 端接口返回中
- WHEN 管理员删除公告 THEN 系统 SHALL 软删除，C 端不再返回

---

### FR-04：Banner 管理（管理端）

**描述：** 管理员维护 App 内 Banner 轮播内容。

**验收标准：**
- WHEN 管理员新增 Banner（图片 URL、跳转链接、排序序号）并启用 THEN 系统 SHALL 使该 Banner 出现在 C 端接口
- WHEN 已启用 Banner 数量已达 5 条时管理员再次启用新 Banner THEN 系统 SHALL 返回提示，拒绝超出限制
- WHEN 管理员禁用 Banner THEN 系统 SHALL 使该 Banner 不出现在 C 端接口返回中

---

### FR-05：C 端公开接口（服务端）

**描述：** 提供 App 端拉取游戏列表、分类、公告、Banner 的公开接口，无需登录。

**验收标准：**
- WHEN App 调用 `GET /api/v1/user/public/license/games` THEN 系统 SHALL 返回全量已上架游戏列表（gzip 压缩，含 rom_name、display_name、category_id、cover_url、size_bytes、is_vip、md5）
- WHEN App 调用 `GET /api/v1/user/public/license/game-categories` THEN 系统 SHALL 返回已启用分类列表，按 sort_order 升序
- WHEN App 调用 `GET /api/v1/user/public/license/game-notices` THEN 系统 SHALL 返回已发布公告列表，支持分页，按发布时间倒序
- WHEN App 调用 `GET /api/v1/user/public/license/game-banners` THEN 系统 SHALL 返回已启用 Banner 列表，按 sort_order 升序
- WHEN 以上接口被未登录客户端调用 THEN 系统 SHALL 正常返回数据，不要求 token

---

### FR-06：ROM 下载接口（服务端）

**描述：** 提供已登录用户获取 ROM 下载链接的接口，服务端完成鉴权后返回腾讯 CDN 签名 URL，App 直接请求 CDN 下载，支持断点续传。

**验收标准：**
- WHEN 已登录用户请求游戏下载链接 `POST /api/v1/user/auth/license/game-download-urls` THEN 系统 SHALL 返回主 ROM 及所有依赖 parent ROM 的 CDN 签名 URL 集合及过期时间
- WHEN 未登录用户请求下载链接 THEN 系统 SHALL 返回 401，拒绝
- WHEN 已登录非 VIP 用户请求 VIP 游戏的下载链接 THEN 系统 SHALL 返回 403，拒绝
- WHEN 已登录 VIP 用户请求 VIP 游戏的下载链接 THEN 系统 SHALL 校验会员身份通过后返回签名 URL 集合
- WHEN App 拿到签名 URL 后携带 Range 头请求 CDN THEN CDN SHALL 支持断点续传（206 Partial Content）
- WHEN 服务端签发 URL 时 THEN 系统 SHALL 将下载记录写入 download_log（user_id、rom_name、tenant_id、时间，只记主 ROM）

---

### FR-10：底栏四 Tab 容器（App mjarch）

**描述：** App 底部导航改为四 Tab 结构，替换现有游戏列表入口。

> **Flavor 隔离硬规则：** 本 FR 及后续所有 App 端 FR（FR-10 ～ FR-18）的改动，严格限定在 `app/src/mjarch/` 目录下，禁止触碰其他 flavor（mjguoguan1/2、mjarch3 等）代码与资源。若必须修改共享代码（baseuilib 等），改动处必须用 `instanceof` 或 flavor 运行时判断保护，确保其他 flavor 行为零变化。

**验收标准：**
- WHEN 用户打开 mjarch App THEN 系统 SHALL 展示底栏四 Tab：游戏大厅 / 消息 / 我的 / 设置（2026-10-05 用户定稿，原「收藏夹」Tab 由「消息」替代）
- WHEN 用户切换 Tab THEN 系统 SHALL 无闪烁地切换对应 Fragment，Fragment 保活不重新加载
- WHEN 其他 flavor 安装运行时 THEN 系统 SHALL 维持原有界面与行为，零差异

---

### FR-11：游戏大厅 Tab（App mjarch）

**描述：** 游戏大厅是 App 首屏，展示会员状态、公告、Banner 轮播、分类 Tab、游戏卡片网格。

**验收标准：**
- WHEN 已登录用户进入游戏大厅 THEN 系统 SHALL 在顶部展示当前会员状态（免费/VIP）及到期时间
- WHEN 未登录用户进入游戏大厅 THEN 系统 SHALL 在顶部展示「未登录」，点击可跳转登录页
- WHEN 公告数据加载完成 THEN 系统 SHALL 在搜索栏下方展示最新一条已发布公告（标题+摘要），点击「更多」跳转公告列表页
- WHEN Banner 数据加载完成 THEN 系统 SHALL 展示水平自动轮播 Banner
- WHEN 分类数据加载完成 THEN 系统 SHALL 展示分类 Tab 栏（含「全部」Tab），用户切换 Tab 时过滤游戏列表
- WHEN 游戏列表加载完成 THEN 系统 SHALL 以双列卡片网格展示游戏，每张卡片含封面图、中文名、VIP 标记（如适用）、下载状态角标
- WHEN 用户下拉刷新 THEN 系统 SHALL 重新拉取服务端游戏列表、公告与 Banner
- WHEN 服务端游戏列表接口请求失败 THEN 系统 SHALL 回退使用 files.zip 内置列表，不崩溃，不显示错误页
- WHEN 用户点击游戏卡片 THEN 系统 SHALL 跳转到该游戏的详情页（FR-12）

---

### FR-12：游戏详情页（App mjarch）

**描述：** 展示游戏详情，通过状态机按钮驱动下载与游玩流程。

**验收标准：**
- WHEN 用户进入详情页 THEN 系统 SHALL 展示游戏封面大图、中文名、分类、文件体积、免费/VIP 标记
- WHEN 游戏未下载且用户已登录（免费游戏） THEN 系统 SHALL 展示「下载」主按钮
- WHEN 游戏未下载且用户未登录 THEN 系统 SHALL 展示「下载（需登录）」按钮，点击跳转登录页
- WHEN 游戏为 VIP 且用户非会员 THEN 系统 SHALL 展示「开通会员」按钮，点击跳转会员购买页
- WHEN 游戏为 VIP 且用户已是 VIP 会员 THEN 系统 SHALL 展示「下载」主按钮
- WHEN 用户点击「下载」且下载开始 THEN 系统 SHALL 将主按钮切换为「暂停」，展示下载进度条与百分比
- WHEN 用户点击「暂停」 THEN 系统 SHALL 暂停下载，主按钮切换为「继续」
- WHEN 用户点击「继续」 THEN 系统 SHALL 从断点恢复下载，主按钮切换为「暂停」
- WHEN 下载完成且 md5 校验通过 THEN 系统 SHALL 主按钮切换为「开始游戏」
- WHEN 下载完成但 md5 校验失败 THEN 系统 SHALL 自动删除损坏文件，提示用户重新下载，主按钮回到「下载」
- WHEN 游戏已下载且用户点击「开始游戏」 THEN 系统 SHALL 启动 MAME 运行该游戏
- WHEN 游戏已下载时 THEN 系统 SHALL 展示次操作「删除下载」；点击后二次确认，确认后删除 rom zip + 解压目录 + 对应 cfg/nvram
- WHEN 用户点击收藏图标 THEN 系统 SHALL 切换本地收藏状态（已收藏↔未收藏），并持久化到本地存储
- WHEN 用户查看详情页底部 THEN 系统 SHALL 展示「评论功能敬请期待」占位文案（本期不实现评论）

---

### FR-13：下载队列管理（App mjarch）

**描述：** 管理下载任务的持久化与并发控制，无独立下载管理页。

**验收标准：**
- WHEN App 重启时存在未完成下载任务 THEN 系统 SHALL 自动恢复下载队列，继续断点续传
- WHEN 同时触发多个下载任务 THEN 系统 SHALL 限制并发下载数不超过 2，多余任务排队
- WHEN 下载完成 THEN 系统 SHALL 自动校验 md5，校验失败删除文件并通知用户重新下载

---

### FR-14：我的 Tab（App mjarch）

**描述：** 展示当前用户账号与会员状态，提供激活码、购买会员、登出入口。

**验收标准：**
- WHEN 已登录用户进入「我的」Tab THEN 系统 SHALL 展示邮箱、会员等级、会员到期时间
- WHEN 未登录用户进入「我的」Tab THEN 系统 SHALL 展示「立即登录」入口
- WHEN 用户点击「激活码」 THEN 系统 SHALL 跳转激活码输入页
- WHEN 用户点击「购买会员」 THEN 系统 SHALL 跳转会员购买页
- WHEN 已登录用户点击「登出」 THEN 系统 SHALL 清除本地 session，顶部会员状态更新为「未登录」

---

### FR-15：收藏管理（App mjarch，2026-10-05 调整）

**描述：** 收藏功能保留，但不再占用底栏独立 Tab（底栏改为 游戏大厅/消息/我的/设置）。收藏列表收进「我的」Tab 作为入口。

**验收标准：**
- WHEN 用户进入「我的」Tab 并点击「我的收藏」入口 THEN 系统 SHALL 展示本地收藏的游戏列表（卡片样式与大厅一致）
- WHEN 用户点击收藏列表中的游戏卡片 THEN 系统 SHALL 跳转到该游戏详情页（FR-12）
- WHEN 收藏列表为空 THEN 系统 SHALL 展示空状态提示
- WHEN 用户在详情页取消收藏后返回收藏列表 THEN 系统 SHALL 从列表中移除该游戏

---

### FR-16：设置页——已下载游戏管理（App mjarch）

**描述：** 设置页新增游戏下载管理入口，支持逐个清除已下载游戏占用的空间。

**验收标准：**
- WHEN 用户进入「设置 → 已下载游戏管理」 THEN 系统 SHALL 展示已下载游戏列表，每项含游戏中文名和占用磁盘空间
- WHEN 用户点击某游戏的「清除」按钮 THEN 系统 SHALL 弹出二次确认对话框
- WHEN 用户确认清除 THEN 系统 SHALL 删除该游戏的 rom zip + 解压目录 + 对应 cfg/nvram，并从列表中移除
- WHEN 页面加载完成 THEN 系统 SHALL 在顶部展示下载目录总占用空间
- WHEN 已下载游戏列表为空 THEN 系统 SHALL 展示空状态提示

---

### FR-17：rompath 多路径（App mjarch）

**描述：** MAME 内核接收 rompath 时包含内置目录与动态下载目录，两路径以 `;` 分隔。

**验收标准：**
- WHEN App 启动并拉起 MAME 时 THEN 系统 SHALL 将 rompath 设为「内置解压目录;getExternalFilesDir()/roms/」
- WHEN 游戏 ROM 存在于下载目录中 THEN 系统 SHALL 能正常启动并游玩该游戏
- WHEN 游戏 ROM 仅存在于内置目录中 THEN 系统 SHALL 仍能正常启动（兜底不受影响）

---

### FR-18：公告列表页（App mjarch）

**描述：** 独立公告列表页，展示历史公告，支持分页。

**验收标准：**
- WHEN 用户从游戏大厅公告栏点击「更多」 THEN 系统 SHALL 跳转到公告列表页
- WHEN 公告列表页加载完成 THEN 系统 SHALL 按发布时间倒序展示已发布公告（标题+摘要+时间），支持上滑加载更多
- WHEN 用户点击某条公告 THEN 系统 SHALL 展示公告详情（标题、正文、发布时间）

---

### FR-20：游戏管理页（管理端前端）

**描述：** dev-web-admin 新增游戏管理页面，对接 FR-01 后端接口。

**验收标准：**
- WHEN 管理员进入游戏管理页 THEN 系统 SHALL 展示游戏列表（rom 名、中文名、分类、封面缩略图、体积、VIP 标记、操作列）
- WHEN 管理员点击「新增」并填写表单提交 THEN 系统 SHALL 调用创建接口，刷新列表
- WHEN 管理员点击「编辑」某行 THEN 系统 SHALL 打开编辑表单，回填现有数据，保存后刷新列表
- WHEN 管理员点击「删除」某行并确认 THEN 系统 SHALL 调用删除接口，列表移除该条目

---

### FR-21：分类管理页（管理端前端）

**验收标准：**
- WHEN 管理员进入分类管理页 THEN 系统 SHALL 展示分类列表及操作（新增/编辑/排序/删除）
- WHEN 管理员调整排序并保存 THEN 系统 SHALL 更新分类顺序

---

### FR-22：公告管理页（管理端前端）

**验收标准：**
- WHEN 管理员进入公告管理页 THEN 系统 SHALL 展示公告列表（标题、状态、发布时间、操作）
- WHEN 管理员新增/编辑公告并发布 THEN 系统 SHALL 调用接口，状态切换为「已发布」

---

### FR-23：Banner 管理页（管理端前端）

**验收标准：**
- WHEN 管理员进入 Banner 管理页 THEN 系统 SHALL 展示 Banner 列表（图片预览、排序序号、状态、操作）
- WHEN 管理员新增/编辑 Banner 并启用 THEN 系统 SHALL 调用接口，Banner 在 App 端即时生效

---

## 非功能需求

- **性能：**
  - 游戏列表接口 gzip 压缩后预期 < 500KB；App 端缓存 6 小时，后台静默刷新
  - 下载接口支持 HTTP Range 断点续传，单设备并发下载 ≤ 2
  - App 冷启动首屏（本地缓存命中时）< 2s

- **安全：**
  - ROM 下载接口需有效 JWT token 鉴权（access_token + X-Device-ID），与 license 插件现有认证体系一致
  - VIP 游戏下载接口服务端必须二次校验会员身份，不依赖客户端标记
  - 管理端 CRUD 接口挂 `license:games`、`license:game-categories`、`license:notices`、`license:banners` 权限点，新增权限 seed 复用现有 RBAC

- **多租户：**
  - 游戏数据按 tenant_id 隔离；C 端公开接口按当前租户上下文过滤

- **App 兼容性：**
  - minSdk 21 不变；不引入新网络库（复用现有 okhttp4）
  - 下载目录使用 `getExternalFilesDir(Environment.DIRECTORY_DOWNLOADS)/roms/`

- **容错与降级：**
  - 服务端列表拉取失败，App 回退 files.zip 内置列表，不崩溃、不显示错误页
  - App 重启后自动恢复下载队列（断点续传）
  - md5 校验失败自动删除损坏文件，不静默忽略

---

## 范围外（本期明确不做）

| 项目 | 备注 |
|------|------|
| APK 瘦身（移除 files.zip） | 留待后续版本 |
| ROM 版本增量更新 | md5 字段已预留，逻辑本期不做 |
| 游戏评论/评价 | 详情页仅留「敬请期待」占位 |
| 收藏云同步 | 本期本地存储，后续同步至服务端 |
| App 厂商推送通道（FCM/小米/华为） | 本期不做 |
| ROM CDN/OSS 分发 | 本期服务端本地磁盘直出，接口已留抽象 |
| 一键全部清理下载 | 误删风险，逐个操作即可 |
| 其他 flavor（mjguoguan1/2、mjarch3）改造 | 严格不触碰 |
| iOS 客户端 | 不在范围 |
