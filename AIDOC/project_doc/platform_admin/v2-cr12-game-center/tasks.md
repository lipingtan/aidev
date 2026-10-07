# 任务列表：mjarch 游戏中心（v2-cr12）

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 22 |
| 已完成 | 22 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 22/22 (100%) |
| 当前阶段 | Phase 6: 集成验证（通过） |

---

## Phase 1: 服务端数据模型

### Task 1: games 表字段扩展（DDL + Model） ✅

**复杂度**: 高

**设计来源:**
- design.md §2.2.1 games 表增量变更
- design.md §2.3.1 Game 模型扩展

**Scope（边界）:**
- 涉及文件:
  - `backend/app/license/models/models.go`（修改：Game 结构体追加新字段）
- 涉及模块: license
- 不触碰: 现有 Game 字段、其他 Model 定义

**Constraints（约束）:**
- 新增字段全部使用指针类型（`*int64`、`*string`）接收 NULL，is_vip 用 `int` 默认 0
- `CategoryID` 不加 `,string` json tag（普通自增 ID，前端用 number）
- `Parents` 使用 `datatypes.JSON` 类型
- AutoMigrate 只加列不删列，存量数据不受影响
- 所有新字段 gorm tag 必须包含 `default:null` 或合理默认值

**Acceptance（验证标准）:**
- AC: Game struct 包含 category_id / cover_url / size_bytes / is_vip / parents / md5 六个新字段，类型和 tag 与 design.md 一致
- AC: is_vip 默认值为 0，category_id 允许 NULL
- AC: go build ./... 零错误
- AC: 【回归】现有 games 表数据查询字段不变（RG-1）

**自测:**
- 测试文件: `backend/app/license/models/models_test.go`（新建）
- ST: 新增 Game 记录不设新字段 → is_vip=0、其余新字段为 NULL，存量兼容通过

---

### Task 2: 新增 game_categories / game_notices / game_banners / game_download_logs 模型 ✅

**复杂度**: 高

**设计来源:**
- design.md §2.2.2～2.2.5 四张新表 DDL

**Scope（边界）:**
- 涉及文件:
  - `backend/app/license/models/models.go`（修改：追加四个新 struct）
- 涉及模块: license
- 不触碰: 其他 Model 定义、现有字段

**Constraints（约束）:**
- 四个 struct 均包含 tenant_id（带 INDEX）、created_at、updated_at、deleted_at
- game_categories / game_notices / game_banners 支持软删除（GORM DeletedAt）
- game_download_logs 不需要软删除、不需要 update_by/create_by
- TableName() 方法返回对应表名

**Acceptance（验证标准）:**
- AC: GameCategory / GameNotice / GameBanner / GameDownloadLog 四个 struct 字段与 DDL 一致
- AC: AutoMigrate 后四张表结构正确（索引存在）
- AC: go build ./... 零错误

**自测:**
- 测试文件: `backend/app/license/models/models_test.go`
- ST: 四个 struct 各自 TableName() 返回正确表名
- ST: AutoMigrate 后查询空表不报错

---

### Task 3: plugin.go 追加 AutoMigrate + 路由注册骨架 ✅

**复杂度**: 中

**设计来源:**
- design.md §2.4 菜单 Seed
- design.md §2.5 文件结构

**Scope（边界）:**
- 涉及文件:
  - `backend/app/license/plugin.go`（修改：AutoMigrate 追加新 Model + 路由注册占位 + 菜单 Seed）
- 涉及模块: license
- 不触碰: 现有路由注册、现有菜单 Seed

**Constraints（约束）:**
- AutoMigrate 追加到现有调用后，不替换
- 菜单 Seed 幂等（insert-if-not-exists，模式同现有 SeedMenu 调用）
- 三个新子菜单 sort=11/12/13，挂在现有 License 管理父菜单下
- 路由注册函数调用占位（Task 6 实现具体 handler 后补充）

**Acceptance（验证标准）:**
- AC: AutoMigrate 包含 GameCategory / GameNotice / GameBanner / GameDownloadLog
- AC: 三个新子菜单 seed 可重复执行不产生重复行
- AC: go build ./... 零错误

---

## Phase 2: 服务端 service 层

### Task 4: GameInput DTO 扩展 + 现有游戏 CRUD 增量改动 ✅

**复杂度**: 高

**设计来源:**
- design.md §2.3.1 Game 模型扩展
- design.md §2.1 游戏列表支持 name/categoryId 过滤

**Scope（边界）:**
- 涉及文件:
  - `backend/app/license/service/game_price_service.go`（修改）
- 涉及模块: license
- 不触碰: PriceConfig 相关逻辑、其他 service 文件

**Constraints（约束）:**
- `GameInput` 追加 CategoryID / CoverURL / SizeBytes / IsVIP / Parents / MD5 字段
- `ListGames` 支持 `name`（模糊）、`categoryId` 过滤参数
- `CreateGame` / `UpdateGame` 保存新字段
- 租户隔离：所有查询必须带 `tenant_id` 过滤
- 错误用 `fmt.Errorf` 包装返回

**Acceptance（验收标准）:**
- AC: GameInput 包含六个新字段
- AC: 按 name 模糊搜索返回匹配结果
- AC: 按 categoryId 过滤返回对应分类游戏
- AC: Create/Update 正确保存新字段到数据库
- AC: go build ./... 零错误
- AC: 【回归】现有游戏列表/创建/更新接口响应格式不变（RG-1）

**自测:**
- 测试文件: `backend/app/license/service/game_price_service_test.go`（新建）
- ST: ListGames name="麻将" → 仅返回名称含"麻将"的游戏
- ST: ListGames categoryId=1 → 仅返回该分类游戏
- ST: CreateGame 含 is_vip=1 → 数据库 is_vip=1
- ST: 【回归】ListGames 不传过滤参数 → 返回全量（RG-1）

---

### Task 5: PublicGameItem 扩展 + gzip 输出 ✅

**复杂度**: 中

**设计来源:**
- design.md §2.3.2 C 端游戏列表接口扩展

**Scope（边界）:**
- 涉及文件:
  - `backend/app/license/service/game_price_service.go`（修改：PublicGameItem + ListPublicGames）
- 涉及模块: license
- 不触碰: 其他 handler、其他 service、game_content_api.go（由 Task 8 统一建立）

**Constraints（约束）:**
- PublicGameItem 追加 category_id / cover_url / size_bytes / is_vip / parents / md5
- handler 层用 `compress/gzip` + `json.NewEncoder` 流式写出，响应头加 `Content-Encoding: gzip`
- 只返回 status=1 的游戏（已上架）
- 租户通过 ResolveTenantByHost() 获取，不从请求参数接收

**Acceptance（验证标准）:**
- AC: GET /api/v1/user/public/license/games 返回包含新字段的游戏列表
- AC: 响应头含 Content-Encoding: gzip
- AC: 未登录客户端可正常访问（无 token）
- AC: go build ./... 零错误
- AC: 【回归】旧版 App 解析响应不崩溃（RG-8）

**自测:**
- 测试文件: `backend/app/license/service/game_price_service_test.go`
- ST: ListPublicGames → 返回包含 is_vip / md5 等新字段的列表
- ST: 下架游戏（status=0） → 不出现在公开列表

---

### Task 6: game_content_service.go 新增（分类/公告/Banner CRUD） ✅

**复杂度**: 高

**设计来源:**
- design.md §2.3.3 Banner 启用数量限制
- design.md §2.3.4 分类删除逻辑
- design.md §2.1 分类/公告/Banner 管理端接口

**Scope（边界）:**
- 涉及文件:
  - `backend/app/license/service/game_content_service.go`（新建）
- 涉及模块: license
- 不触碰: game_price_service.go、其他 service 文件

**Constraints（约束）:**
- GameContentService 使用同一个 `*gorm.DB` 实例（与 GameService 保持一致）
- 分类删除必须在单一事务内：先 UPDATE games.category_id=NULL，再软删除分类
- Banner 启用前检查 status=1 数量，≥5 条返回错误
- 所有查询带 tenant_id 过滤
- 软删除用 GORM 的 Delete（走 deleted_at），不物理删除

**Acceptance（验证标准）:**
- AC: 分类 CRUD 四个方法均可正常调用
- AC: 删除分类时同事务将游戏 category_id 置 NULL
- AC: 启用第 6 条 Banner 时返回错误，已有 5 条不变
- AC: 公告状态切换（草稿↔已发布）正确
- AC: go build ./... 零错误

**自测:**
- 测试文件: `backend/app/license/service/game_content_service_test.go`（新建）
- ST: DeleteCategory 有游戏 → 游戏 category_id=NULL + 分类软删除（同事务）
- ST: DeleteCategory 事务失败 → 两侧均回滚，数据不变
- ST: EnableBanner 已有 5 条 → 返回错误，第 6 条未启用
- ST: EnableBanner 已有 4 条 → 成功，变为 5 条

---

### Task 7: CDN 签名 URL 批量接口（game-download-urls） ✅

**复杂度**: 高

**设计来源:**
- design.md §2.3.5 ROM 下载 URL 批量签名接口

**Scope（边界）:**
- 涉及文件:
  - `backend/app/license/service/game_content_service.go`（修改：追加 GenerateCDNSignURL + GetDownloadURLs）
- 涉及模块: license
- 不触碰: 其他 service 文件、现有 license 服务

**Constraints（约束）:**
- `GenerateCDNSignURL` 为纯函数，不依赖外部状态，可直接单测
- 算法严格按腾讯 CDN TypeA：`{path}-{timestamp}-{rand}-0-{secret}` 的 MD5
- `GetDownloadURLs` 流程：JWT 校验 → 查 games → VIP 校验（同插件内 checkUserMembership）→ 展开 parents → 批量签名 → 写 download_log（只记主 ROM）→ 返回
- CDN 文件路径：`https://{cdn_domain}/roms/{rom_name}.zip`
- 配置从 settings.yml `game.cdn_secret / cdn_domain / cdn_sign_expire_secs` 读取
- 签名失败返回 HTTP 500 + 友好 msg，不暴露内部错误

**Acceptance（验证标准）:**
- AC: POST /api/v1/user/auth/license/game-download-urls 返回 main + parents[] 结构
- AC: 未登录返回 401
- AC: 非 VIP 会员请求 VIP 游戏返回 403
- AC: 无 parent 游戏返回 parents=[]
- AC: download_log 只写主 ROM 记录
- AC: GenerateCDNSignURL 生成的 sign 参数格式正确
- AC: go build ./... 零错误

**自测:**
- 测试文件: `backend/app/license/service/game_content_service_test.go`
- ST: GenerateCDNSignURL("https://cdn/roms/test.zip", "cdntlping", 7200) → sign 格式 `{ts}-{rand}-0-{md5}` 正确
- ST: GenerateCDNSignURL 两次调用 → sign 不同（rand 不重复）
- ST: GetDownloadURLs 游戏有 2 个 parent → 返回 1 个 main + 2 个 parent
- ST: GetDownloadURLs VIP 游戏 + 非会员 → 403
- ST: download_log 写入 → 只有主 ROM 记录，无 parent 记录

---

## Phase 3: 服务端 handler + 路由

### Task 8: game_content_api.go 实现（管理端接口） ✅

**复杂度**: 高

**设计来源:**
- design.md §2.1 管理端接口（16 个）
- design.md §2.4 菜单 Seed 权限 code

**Scope（边界）:**
- 涉及文件:
  - `backend/app/license/apis/game_content_api.go`（新建）
  - `backend/app/license/plugin.go`（修改：RegisterAdminRoutes 追加路由）
- 涉及模块: license
- 不触碰: 现有 handler 文件、现有路由

**Constraints（约束）:**
- 参照 membership_api.go 风格实现（参数绑定/错误处理/响应格式）
- 所有管理端接口从 context 获取 tenant_id（middleware.GetTenantId），不从请求参数读
- Banner PUT 接口在 status 变为 1 时调用 EnableBanner（含数量校验）
- 分类 DELETE 接口调用 DeleteCategory（含事务）

**Acceptance（验证标准）:**
- AC: 分类/公告/Banner 各 CRUD 接口（12 个）返回正确响应格式
- AC: 游戏管理接口（4 个）支持 name/categoryId 过滤
- AC: tenant_id 从 context 获取，不可伪造
- AC: go build ./... 零错误
- AC: 【回归】现有管理端路由不受影响（RG-4）

**自测:**
- 测试文件: `backend/app/license/apis/game_content_api_test.go`（新建）
- ST: DELETE /game-categories/:id → 分类软删除 + 游戏归未分类
- ST: PUT /game-banners/:id（启用第 6 条）→ 400 含错误 msg
- ST: 【回归】现有 /games CRUD → 响应格式不变（RG-1）

---

### Task 9: C 端公开接口 + 下载接口路由注册 ✅

**复杂度**: 中

**设计来源:**
- design.md §2.1 C 端公开接口 + C 端认证接口

**Scope（边界）:**
- 涉及文件:
  - `backend/app/license/plugin.go`（修改：RegisterPublicUserRoutes + RegisterAuthUserRoutes 追加路由）
  - `backend/app/license/apis/game_content_api.go`（修改：补充 handler 方法）
- 涉及模块: license
- 不触碰: 现有路由、现有 handler

**Constraints（约束）:**
- 公开接口（games/game-categories/game-notices/game-banners）挂 RegisterPublicUserRoutes
- 下载接口（game-download-urls）挂 RegisterAuthUserRoutes，复用现有 JWT + X-Device-ID 中间件
- 路径格式遵循现有风格：`g.GET("/game-categories", ...)` 等

**Acceptance（验证标准）:**
- AC: GET /api/v1/user/public/license/game-categories 无 token 可访问
- AC: POST /api/v1/user/auth/license/game-download-urls 无 token 返回 401
- AC: go build ./... 零错误
- AC: 【回归】现有 /public/license/games 接口不受影响（RG-1）

---

## Phase 4: 管理端前端

### Task 10: license.ts 追加 API 函数 ✅

**复杂度**: 中

**设计来源:**
- design.md §4.1 文件结构
- design.md §4.2 前后端字段对齐

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/api/license.ts`（修改）
- 涉及模块: license 前端
- 不触碰: 现有 API 函数

**Constraints（约束）:**
- 必须使用 `@/utils/request`（axios wrapper）
- category_id 前端类型为 `number | null`（非 string）
- id 字段前端类型为 `string`（雪花 ID）
- 函数命名遵循现有风格（如 `getGameCategories`、`createGameCategory`）

**Acceptance（验证标准）:**
- AC: 分类/公告/Banner 各 CRUD API 函数定义完整（共 12 个）
- AC: TypeScript 编译无类型错误
- AC: vite build 零错误

---

### Task 11: Games.vue 增强新字段 ✅

**复杂度**: 中

**设计来源:**
- design.md §4.3 Games.vue 增量改动

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/views/license/Games.vue`（修改）
  - `dev-web-admin/src/api/license.ts`（依赖 Task 10）
- 涉及模块: license 前端
- 不触碰: 其他 Vue 页面

**Constraints（约束）:**
- 页面初始化时并行调用分类列表接口，构建 `categoryMap: Record<number, string>`
- 分类列下拉用 el-select，数据来源同上 categoryMap
- is_vip 用 el-switch 展示/编辑
- cover_url 用 el-input（文本 URL），表格列显示缩略图（img 标签，加载失败时显示占位图）
- 【回归】现有表格列和表单字段不被删除或改变

**Acceptance（验证标准）:**
- AC: 表格新增分类名（前端映射）、封面缩略图、体积、VIP 标记四列
- AC: 编辑表单包含分类下拉（el-select）、封面 URL、体积、is_vip、md5 字段
- AC: vite build 零错误
- AC: 【回归】现有表格列（rom_name/display_name/status/操作）不变（RG-4）

---

### Task 12: GameCategories.vue 新增 ✅

**复杂度**: 中

**设计来源:**
- design.md §4.1 文件结构
- requirements.md FR-21

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/views/license/GameCategories.vue`（新建）
  - `dev-web-admin/src/router/static-routes.ts`（修改：追加路由）
- 涉及模块: license 前端

**Constraints（约束）:**
- 复用现有 el-table + el-dialog 表单模式（参照 Games.vue）
- 路由 path `/license/game-categories`，permission: `license:game-categories`
- 删除操作前展示确认弹窗，说明游戏将归入未分类
- 排序字段 sort_order 可编辑

**Acceptance（验证标准）:**
- AC: 分类列表正常加载，支持新增/编辑/删除
- AC: 删除确认弹窗说明游戏归未分类
- AC: 路由可达，无权限时不显示
- AC: vite build 零错误

---

### Task 13: GameNotices.vue 新增 ✅

**复杂度**: 中

**设计来源:**
- design.md §4.1；requirements.md FR-22

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/views/license/GameNotices.vue`（新建）
  - `dev-web-admin/src/router/static-routes.ts`（修改）

**Constraints（约束）:**
- 状态字段用 el-tag 展示（草稿/已发布）
- 表单包含标题、正文（el-input textarea）、状态切换

**Acceptance（验证标准）:**
- AC: 公告列表含标题/状态/发布时间/操作列
- AC: 新增/编辑/发布/下线操作正常
- AC: vite build 零错误

---

### Task 14: GameBanners.vue 新增 ✅

**复杂度**: 中

**设计来源:**
- design.md §4.1；requirements.md FR-23

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/views/license/GameBanners.vue`（新建）
  - `dev-web-admin/src/router/static-routes.ts`（修改）

**Constraints（约束）:**
- 启用操作调用 PUT 接口将 status 设为 1，后端返回 400 时展示错误提示（最多 5 条限制）
- 图片 URL 字段预览缩略图

**Acceptance（验证标准）:**
- AC: Banner 列表含图片预览/排序/状态/操作
- AC: 启用第 6 条时前端展示后端返回的错误 msg
- AC: vite build 零错误

---

## Phase 5: App 端（mjarch flavor）

### Task 15: GameDatabase + SQLite 表结构 ✅

**复杂度**: 高

**设计来源:**
- design.md §3.1 本地数据库

**Scope（边界）:**
- 涉及文件:
  - `app/src/mjarch/java/.../gamecenter/db/GameDatabase.java`（新建）
- 涉及模块: mjarch gamecenter
- 不触碰: 其他 flavor 代码、现有 MembershipStorage

**Constraints（约束）:**
- 原生 SQLiteOpenHelper 单例，不引入 Room
- download_tasks 表包含 job_rom_name / rom_name / is_main / url / expires_at / total_bytes / downloaded_bytes / status / md5 / updated_at
- `(job_rom_name, rom_name)` 联合唯一约束
- App 启动时将 DOWNLOADING(1) → PAUSED(2) 自动降级

**Acceptance（验证标准）:**
- AC: 四张表（game_list / game_categories / download_tasks / favorites）创建成功
- AC: download_tasks 联合唯一约束存在
- AC: App 冷启动后 DOWNLOADING 记录自动降级为 PAUSED

---

### Task 16: GameCenterApi.java ✅

**复杂度**: 中

**设计来源:**
- design.md §3.2 模块结构；design.md §2.1 C 端接口

**Scope（边界）:**
- 涉及文件:
  - `app/src/mjarch/java/.../gamecenter/api/GameCenterApi.java`（新建）
- 涉及模块: mjarch gamecenter
- 不触碰: MembershipApi、其他 flavor

**Constraints（约束）:**
- 复用 MembershipApi 的 OkHttpClient（单例，含 Authorization + X-Device-ID header）
- 接口：getGameList / getCategories / getNotices / getBanners / getDownloadUrls
- getDownloadUrls 需要登录 token，其余无需 token
- 网络失败返回 null 或空列表，不抛异常

**Acceptance（验证标准）:**
- AC: 五个接口方法定义完整
- AC: getDownloadUrls 携带 Authorization header
- AC: 公开接口无 Authorization header
- AC: 编译通过

---

### Task 17: GameRepository.java ✅

**复杂度**: 高

**设计来源:**
- design.md §3.2 模块结构；design.md §3.1 parent ROM 共享策略

**Scope（边界）:**
- 涉及文件:
  - `app/src/mjarch/java/.../gamecenter/repository/GameRepository.java`（新建）
- 涉及模块: mjarch gamecenter
- 不触碰: 其他 flavor、现有 MembershipStorage

**Constraints（约束）:**
- 游戏列表合并：服务端列表（优先）+ files.zip 内置列表（兜底），服务端拉取失败时只用内置
- 服务端列表缓存 6 小时（game_center_meta SharedPreferences）
- 收藏 CRUD 操作 favorites 表
- 清除游戏：删主 ROM 文件 + 解压目录 + cfg/nvram；对每个 parent ROM 检查：① 被其他游戏作为 parent 依赖，或 ② 本身是某已下载游戏的主 ROM——满足任一条件则跳过删除，两个条件均不满足才删除

**Acceptance（验证标准）:**
- AC: 服务端拉取失败时返回内置列表，不崩溃
- AC: 6 小时内重复调用使用本地缓存
- AC: 清除游戏时 parent ROM 被其他游戏作为 parent 依赖 → 保留文件
- AC: 清除游戏时 parent ROM 本身是某已下载游戏的主 ROM → 保留文件
- AC: 清除游戏时 parent ROM 无任何依赖且非主 ROM → 删除文件
- AC: 收藏/取消收藏正确持久化

---

### Task 18: DownloadManager.java ✅

**复杂度**: 高

**设计来源:**
- design.md §3.4 DownloadManager 核心流程

**Scope（边界）:**
- 涉及文件:
  - `app/src/mjarch/java/.../gamecenter/download/DownloadManager.java`（新建）
- 涉及模块: mjarch gamecenter
- 不触碰: 其他 flavor、现有网络代码

**Constraints（约束）:**
- 并发下载上限 2 个 FileTask（跨 Job）
- parent ROM 先于 main ROM 入队
- 下载前检查文件已存在且 md5 匹配 → 直接标记 COMPLETED，跳过下载
- expires_at 距过期 ≤ 5 分钟时，重新调 getDownloadUrls 刷新 URL
- 下载完成后强制 md5 校验，失败则删文件 + status=FAILED
- 进度 = Σ downloaded_bytes / Σ total_bytes（所有 FileTask 合并）

**Acceptance（验证标准）:**
- AC: enqueueJob 写入 download_tasks，parents 排在 main 前
- AC: 并发不超过 2 个 FileTask
- AC: 已存在且 md5 匹配的文件直接 COMPLETED，不重复下载
- AC: URL 临近过期时自动刷新
- AC: md5 校验失败 → 文件删除 + status=FAILED
- AC: 进度回调聚合所有 FileTask

---

### Task 19: MjarchGameListActivity 四 Tab 改造 ✅

**复杂度**: 高

**设计来源:**
- design.md §3.3 MjarchGameListActivity 改造

**Scope（边界）:**
- 涉及文件:
  - `app/src/mjarch/java/.../MjarchGameListActivity.java`（修改）
  - `app/src/mjarch/res/layout/activity_mjarch_game_list.xml`（修改）
  - `app/src/mjarch/java/.../gamecenter/ui/GameHallFragment.java`（新建）
  - `app/src/mjarch/java/.../gamecenter/ui/MyFragment.java`（新建）
  - `app/src/mjarch/java/.../gamecenter/ui/FavoritesFragment.java`（新建）
- 涉及模块: mjarch
- 不触碰: 基类（BaseActivity 等）、其他 flavor 任何文件

**Constraints（约束）:**
- Fragment 用 show/hide 保活切换，不 replace
- 改动只在 MjarchGameListActivity.java 本身 + 新增 Fragment
- 若需覆盖基类行为，只在 mjarch 子类 @Override，不改基类
- GameHallFragment 包含大厅顶部细进度条（有活跃下载任务时显示，点击跳详情页）

**Acceptance（验证标准）:**
- AC: mjarch App 展示底栏四 Tab：游戏大厅/消息/我的/设置（2026-10-05 用户修订，收藏夹入口移入「我的」Tab）
- AC: Tab 切换无闪烁，Fragment 保活
- AC: 有下载任务时大厅顶部展示细进度条，点击跳转对应游戏详情页
- AC: 已登录用户大厅顶部展示会员状态（免费/VIP）及到期时间
- AC: 未登录用户大厅顶部展示「未登录」，点击跳转登录页
- AC: 服务端游戏列表拉取失败时回退 files.zip 内置列表，不崩溃不显示错误页
- AC: 【回归】其他 flavor 安装运行时行为零变化（RG-3）

---

### Task 20: GameDetailActivity.java ✅

**复杂度**: 高

**设计来源:**
- design.md §3.2；requirements.md FR-12

**Scope（边界）:**
- 涉及文件:
  - `app/src/mjarch/java/.../gamecenter/ui/GameDetailActivity.java`（新建）
  - `app/src/mjarch/res/layout/activity_game_detail.xml`（新建）
- 涉及模块: mjarch gamecenter
- 不触碰: 其他 flavor、基类

**Constraints（约束）:**
- 主按钮区为状态机（7 种状态：未下载已登录/未登录/VIP非会员/VIP会员/下载中/暂停/已下载）
- 进度条聚合所有 FileTask 总进度（Σbytes）
- 删除下载：删主 ROM + parent ROM（引用计数保护）+ cfg/nvram
- 收藏图标切换本地状态，持久化到 SQLite favorites 表

**Acceptance（验证标准）:**
- AC: 七种主按钮状态正确显示与切换
- AC: 下载进度条实时更新（500ms 间隔）
- AC: md5 失败时主按钮回到「下载」并提示用户
- AC: 「删除下载」点击后弹出二次确认对话框，确认后删主 ROM + cfg/nvram；每个 parent ROM 检查两个条件（被其他游戏依赖 OR 本身是主 ROM），均不满足才删除
- AC: 「开始游戏」正确启动 MAME

---

### Task 21: rompath 多路径 + 其余 UI 模块 ✅

**复杂度**: 中

**设计来源:**
- design.md §3.5；requirements.md FR-16/FR-17/FR-18

**Scope（边界）:**
- 涉及文件:
  - `app/src/mjarch/java/.../MjarchGameListActivity.java`（修改：rompath 传参）
  - `app/src/mjarch/java/.../gamecenter/ui/DownloadedGamesFragment.java`（新建）
  - `app/src/mjarch/java/.../gamecenter/ui/NoticeListActivity.java`（新建）
- 涉及模块: mjarch gamecenter
- 不触碰: 其他 flavor、内核

**Constraints（约束）:**
- rompath = `{builtinDir};{downloadDir}`，使用 `;` 分隔
- DownloadedGamesFragment 列出 download_tasks status=COMPLETED 的游戏，显示总体积（主 ROM + parent ROM 文件合计）
- 清除时调用 GameRepository.deleteGame（含引用计数保护）

**Acceptance（验证标准）:**
- AC: MAME 启动时 rompath 包含下载目录
- AC: 已下载 ROM 在下载目录可正常游玩
- AC: 设置页已下载管理列表正确显示
- AC: 清除游戏后文件被删除，列表更新

---

## Phase 6: 集成验证

### Task 22: 构建验证 + 全链路集成验证 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及范围: 后端编译/vet/测试；前端构建；App 编译

**Acceptance（验证标准）:**
- AC: go build ./... 零错误
- AC: go vet ./... 零警告
- AC: go test ./backend/app/license/... 全部通过
- AC: vite build 零错误
- AC: mjarch APK 编译成功
- AC: 【回归】RG-1 现有游戏列表接口向后兼容
- AC: 【回归】RG-2 license 插件激活/登录/购买链路正常
- AC: 【回归】RG-3 其他 flavor App 行为零变化
- AC: 【回归】RG-4 管理端现有页面正常
- AC: 【回归】RG-5 mjarch 现有会员/购买流程正常
- AC: 【回归】RG-8 旧版 App 调用新接口不崩溃

---

## 并行执行建议

```
Phase 1（可完全并行）:
  Task 1 ──┐
  Task 2 ──┼── 并行（无文件交叉）
  Task 3 ──┘

Phase 2（有依赖）:
  Task 1 → Task 4（GameInput 依赖 Game 新字段）
  Task 2 → Task 6（GameContentService 依赖新 Model）
  Task 4 → Task 5（PublicGameItem 依赖扩展后的 Game）
  Task 6 → Task 7（CDN 接口在同一 service 文件）

Phase 3（依赖 Phase 2）:
  Task 8 + Task 9 可并行（不同路由组）

Phase 4（可完全并行，依赖 Task 10）:
  Task 10 → Task 11/12/13/14（前端 API 先行）

Phase 5（可部分并行）:
  Task 15 → Task 16/17/18（SQLite 先行）
  Task 17 + Task 18 可并行
  Task 19 → Task 20/21（Tab 容器先行）

Phase 6（最后执行）:
  Task 22（依赖所有前置任务）
```

---

## 工时汇总

| Phase | 任务数 | 说明 |
|-------|--------|------|
| Phase 1: 服务端数据模型 | 3 | 可并行，约 1 天 |
| Phase 2: 服务端 service | 4 | 串行为主，约 2.5 天 |
| Phase 3: 服务端 handler + 路由 | 2 | 约 1 天 |
| Phase 4: 管理端前端 | 5 | 约 1.5 天 |
| Phase 5: App 端 | 7 | 约 4 天 |
| Phase 6: 集成验证 | 1 | 约 0.5 天 |
| **合计** | **22** | **约 10.5 天（考虑并行可压缩至 7 天）** |

---

## 变更记录

### 2026-10-05 UI 修订（用户定稿，原型 mjarch_home_v2.html）

1. **底栏 Tab 调整**：游戏大厅 / 消息 / 我的 / 设置（原「收藏夹」→「消息」；收藏列表入口移入「我的」Tab，FR-15 已同步修订）
2. **首页顶栏改版**（参照原型 mjarch_home_v2.html，方案C布局）：
   - 第一行：左上角「街机时代社区」title（strings.xml 资源）+ 右侧登录状态（未登录=「登录/注册」chip；已登录=「尊享会员·N天」）
   - 第二行：搜索框通栏（占位「搜索 N 款街机游戏」；搜索接口本期不实现，入口预留）
3. **首页信息结构**（自上而下）：title行 → 搜索框 → 公告栏 → Banner → 分类标签 → 分页器（上） → 游戏卡片网格 → 分页器（下） → 底部TabBar
4. **滚动模式**：整页上下滚动（title 连带滚动），仅底部 TabBar 固定
5. **游戏卡片样式**：参照 hall_brand_proto.html —— thumb + 游戏名 + 标签行（分类/免费/VIP）+ 右下「详情 ›」；卡片无直接下载/进入按钮，行动统一收敛到详情页
6. **分页**：C 端公开游戏列表接口需补 page/page_size 参数透传（后端 ListGames 已支持分页）

---

### 2026-10-05 后续迭代（游戏大厅深度优化）

**标签区交互改版：**
- 默认折叠2行，超出用 `FrameLayout` 截断（不截字）
- 底部操作行独立：左「更多▾/收起▴」，右🔍搜索图标（各自点击区域 ≥ 44dp）
- 搜索栏默认 GONE，点🔍展开，聚焦显示历史浮层（最近15条）
- 搜索：联网走服务器（5秒超时）→ 失败 fallback 本地 + Toast；结果直接以卡片形式展示（与 Tab 筛选相同）

**顶栏：**
- 独立白底 topbar（从 Banner 内移出），去掉搜索栏占位（搜索移至标签区）
- Logo 使用 SpannableString「街机时代社区」（「时代」粉红 #ff6b9d）
- 三态 chip：未登录灰「登录/注册›」/ 普通用户蓝灰「免费用户」/ VIP 金黄「👑 尊享会员」

**Banner：**
- 圆角 `dp(12)` + 左右 margin `dp(10)` + `setClipToOutline(true)`

**其他：**
- 公告栏右侧加「›」，点击跳消息 Tab
- 游戏详情页标签显示中文名（异步加载 tagNameMap）
- 数字分页控件（1/2/3…），卡片上下各一排
- `SeedDefaultTags` 只保留 `majiang`，删除其他默认标签
- 后端 `/games` 支持 `?q=关键词` 模糊搜索

**原型：** `C:\data\run\hermes\proto\mjarch_home_v2.html`（含交互演示）


---

## 附录：客服系统 VAPID Web Push 配置

### VAPID 密钥对（2026-10-05 生成，生产用）

> ⚠️ 私钥敏感，部署时通过环境变量注入，不要硬编码进代码仓库

| 字段 | 值 |
|---|---|
| **公钥（publicKey）** | `BEFfT2og8-3von_HQzptULmpu0KH8NMDQwGljVGtQI0lkIQ2kbmf2cZz_-ZnE47GH04YPtIzIzYOgvZSUmZmVeU` |
| **私钥（privateKey）** | `93UKa_DS4_UvAr1Pj1u2kzHwP9fElqxqVpnC3olzKfM` |
| 曲线 | P-256 (SECP256R1) |
| 格式 | Base64url（无填充） |
| 生成工具 | Python cryptography 库（`cryptography.hazmat.primitives.asymmetric.ec`） |
| 生成脚本 | `C:\data\run\kirocli\gen_vapid.py` |

### 后端配置方式

密钥对初始化写入 `cs_config` 表（`key=vapid_public_key` / `key=vapid_private_key`），
管理端「客服配置 → 推送配置」Tab 可查看公钥 / 重新生成。

### 重新生成方法

```bash
python C:\data\run\kirocli\gen_vapid.py
```

重新生成后需：
1. 管理端「推送配置」更新公钥
2. 所有已订阅的客服需重新点「开启推送通知」（旧 subscription 失效）

### Web Push 推送流程

```
客服浏览器                  后端                     浏览器推送服务(FCM/Mozilla等)
    |                        |                              |
    |-- POST /cs/push-sub -->|                              |
    |   (endpoint+keys)      |-- 存 cs_push_subscriptions --|
    |                        |                              |
    |  (有新会话/消息)         |                              |
    |                        |-- web-push(VAPID签名) ------->|
    |<------ 推送通知 ---------|                              |
    |  [新会话：用户xxx]        |                              |
```

