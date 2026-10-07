# 设计计划：mjarch 游戏中心（v2-cr12）

> 前置：requirements.md 已于 2026-10-04 确认。
> CR 类型：跨层集成 CR（Go 后端 + Android App mjarch flavor）

---

## 设计方向

基于确认的需求，拟采用以下技术方案：

1. **增强现有 Game 模型**：license 插件已有 `games` 表（`game_price_service.go` + `models.go` 中的 `Game` 结构体），现有字段为 `id / tenant_id / rom_name / display_name / required_levels / sort_order / status`。本 CR 在此基础上新增字段：`category_id, cover_url, size_bytes, is_vip, parents（parent ROM 依赖清单，JSON）, md5`。现有管理端 CRUD 接口（AdminListGames / AdminCreateGame / AdminUpdateGame / AdminDeleteGame）和 `GameInput` DTO 同步扩展新字段，**不重建**。

2. **扩展 C 端 PublicGames 接口**：现有 `PublicGameItem` 只返回 `rom_name + required_levels`，本 CR 扩展为全量游戏信息（含封面/体积/md5/分类/VIP 标记），gzip 压缩输出。接口路径 `GET /api/v1/user/public/license/games` 保持不变，仅扩展响应字段。

3. **新增三张内容表**：在 license 插件 `models.go` 中新增 `GameCategory（分类）/ GameNotice（公告）/ GameBanner（Banner）`，对应在 `game_price_service.go`（或新建 `game_content_service.go`）中各加 CRUD service + handler，注册到 license 插件现有路由体系，权限 seed 挂「游戏中心」菜单。

4. **新增 ROM 下载接口**：在 license 插件 handler 层新增 `GET /api/v1/user/auth/license/game/download/:rom_name`，复用现有 JWT + X-Device-ID 中间件；VIP 校验直接调用同文件内现有 `checkUserMembership` 逻辑；流式读磁盘文件，支持 HTTP Range；下载日志写新增的 `game_download_logs` 表。

5. **App 端改造严格限定 mjarch flavor**：四 Tab 容器（BottomNavigationView + Fragment）替换现有 MjarchGameListActivity；新增 GameDetailActivity（状态机按钮区）；DownloadManager 基于 okhttp4 实现断点续传队列；收藏持久化到本地 SQLite；rompath 多路径 Java 侧追加下载目录。

6. **管理端前端增强已有游戏管理页 + 新增三页面**：现有 `Games.vue` 增强（新字段表单列）；新增 Categories.vue / Notices.vue / Banners.vue，复用现有 el-table + el-dialog 模式。

---

## 技术选型

### 服务端：game 内容管理落地位置

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| 增强现有 Game 模型 + 在 license 插件内扩展（用户指定） | 复用现有 CRUD 接口骨架、路由、中间件、AutoMigrate；VIP 校验同插件内直接调用；管理端已有 Games.vue 可复用增强 | license 插件职责略扩大 | ✓ |
| 新建独立 game 插件 | 职责单一 | 需建完整骨架；与现有 Game 模型重复；已被用户否定 | ✗ |

### ROM 下载方案

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| 服务端本地磁盘 + Go HTTP Range | 实现简单；无外部依赖；接口留抽象可后续换 CDN | 带宽受服务器限制 | ✓（需求已确认） |
| 直接返回 CDN/OSS URL 给客户端 | 带宽分流 | 本期不引入 OSS；需求明确不做 | ✗ |

### App 下载实现

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| okhttp4 自实现 DownloadManager（Range 断点续传） | 已有依赖；可精细控制进度/暂停/续传；无新依赖 | 需自行管理队列和持久化 | ✓ |
| Android DownloadManager 系统 API | 系统内置 | 断点续传行为依赖 OS 版本；需要 INTERNET + WRITE_EXTERNAL_STORAGE 权限组合；难以自定义 Header（Token） | ✗ |

### App 收藏持久化

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| SharedPreferences（Set<String>） | 简单零依赖 | 大量收藏时性能一般 | ✓（本期收藏量有限） |
| SQLite / Room | 查询灵活；适合大数据量 | 引入额外依赖；本期无查询需求 | 备选（后续云同步时升级） |

---


## 澄清问题

- [Question-1] game 插件的权限 seed 与菜单归属：游戏管理页面挂在管理端哪个一级菜单下？
  - 业界最佳实践：内容运营类页面（游戏/公告/Banner）通常独立成「内容管理」一级菜单，与系统配置分开，便于运营角色单独授权。
  - 推荐答案及理由：新增「游戏中心」一级菜单，下挂游戏管理、分类管理、公告管理、Banner 管理四个子菜单；权限 code 前缀统一 `game:`。
  [Answer-1]
  都在 license 下实现。即在现有「License 管理」一级菜单下新增三个子菜单：游戏分类（sort=11）、公告管理（sort=12）、Banner 管理（sort=13），权限 code 沿用 `license:` 前缀（`license:game-categories` / `license:notices` / `license:banners`）。已有「游戏管理」子菜单（sort=9）直接增强，不新建。

- [Question-2] VIP 校验方式：下载接口校验用户是否 VIP，采用何种方式？
  - 业界最佳实践：同模块内调用内部 service 函数最简洁，无额外接口开销；直接查表也可接受（同一数据库）。
  - 推荐答案及理由：直接在 license 插件内调用现有 `checkUserMembership`（或等效的会员判定函数），传入 tenantID、userID、levelCode，返回是否有效会员。同插件调用，无跨模块耦合。
  [Answer-2]
  下载使用 CDN 提供，VIP 校验通过服务端签发签名 URL 实现（见 Question-9）。

- [Question-3] 游戏列表接口是否需要分页？需求确认全量下发，但列表可能增长——全量 gzip 后是否足够？
  - 业界最佳实践：Steam/Google Play 等平台对移动端游戏目录通常采用全量增量更新包（类似 manifest），gzip 后几十 KB 到几百 KB；只要单次响应 < 1MB 即可接受全量方案。mjarch 当前 12633 个 ROM，全量 JSON（每条约 150 字节）= ~1.8MB，gzip 后约 200~400KB，可接受。
  - 推荐答案及理由：维持全量下发，服务端 gzip 压缩，客户端内存解析后本地缓存 6 小时。后续 ROM 量级超过 5 万条再考虑增量更新。
  [Answer-3]
  全量下发。

- [Question-4] download_log 表是否需要 tenant_id 隔离？
  - 业界最佳实践：多租户系统中，业务操作日志均需携带 tenant_id，便于后续统计和隔离。
  - 推荐答案及理由：download_log 加 tenant_id 字段，与现有业务表隔离模式一致。
  [Answer-4]
  需要。

- [Question-5] App 端游戏列表本地缓存格式：缓存到 SharedPreferences（JSON 字符串）还是 SQLite？
  - 业界最佳实践：游戏列表量大（~12633 条）且需要按分类过滤，SQLite 查询更高效；SharedPreferences 适合小量 KV 存储，不适合大列表。
  - 推荐答案及理由：游戏列表缓存到 SQLite（Room 或原生 SQLiteOpenHelper），支持按 category_id 过滤查询；缓存元数据（最后刷新时间）存 SharedPreferences。
  [Answer-5]
  SQLite。游戏列表、下载任务、收藏均存同一个 SQLite 库（原生 SQLiteOpenHelper，不引入 Room 新依赖）。
  另外原先游戏中有数据清除功能默认不要清除 sqlite数据，这个要改原先逻辑的注意一下；

- [Question-6] 下载任务持久化：App 重启后如何恢复下载队列？
  - 业界最佳实践：下载队列持久化到 SQLite，记录 rom_name、已下载字节数、状态（等待/下载中/暂停/完成/失败）；App 启动时扫描未完成任务自动恢复。
  - 推荐答案及理由：下载任务表存 SQLite（与游戏列表缓存同库），App onCreate 时 DownloadManager 扫描并恢复「暂停」状态任务到队列，不自动恢复「下载中」（上次可能异常中断，改为「暂停」让用户手动恢复或自动续传）。
  [Answer-6]
  同意。下载任务表与游戏列表缓存存同一 SQLite 库，App 启动时将「下载中」状态自动降级为「暂停」后恢复队列。

- [Question-7] 管理端分类删除时，已归属该分类的游戏如何处理？
  - 业界最佳实践：强制提示迁移（如 Google Play 的分类删除）或自动归入「未分类」；直接拒绝删除是最保守方案，用户体验稍差但数据安全。
  - 推荐答案及理由：后端校验——分类下有游戏时返回错误提示「请先将该分类下 N 个游戏迁移到其他分类」，阻止删除；分类下游戏为零时允许删除。前端在删除前调一次游戏数接口做预检提示。
  [Answer-7]
  允许删除，删除时自动将该分类下游戏的 category_id 置为 NULL（归入「未分类」），前端删除确认弹窗说明「该分类下 N 个游戏将自动归入未分类」。「未分类」不是数据库记录，category_id = NULL 即表示未分类，App 端 Tab 栏固定展示「全部」+ 各分类，NULL 的游戏只出现在「全部」Tab 中。

- [Question-8] App 四 Tab 容器：MjarchGameListActivity 是否直接改造为 Tab 容器，还是新建 Activity 并让旧的跳转？
  - 业界最佳实践：直接改造现有 Activity 可减少 AndroidManifest 改动，且不影响启动 intent；新建 Activity 更干净但需迁移所有启动路径。
  - 推荐答案及理由：直接改造 MjarchGameListActivity 为四 Tab 容器（替换 setContentView 布局，Fragment 管理替换原列表逻辑），原 Activity 名称和 intent-filter 保持不变，对系统启动路径零影响。
  [Answer-8]
  直接改造。注意改动前 grep 确认基类方法调用，所有新逻辑写在 mjarch flavor 的子类 / Fragment 中，不修改基类，确保其他 flavor 零影响。

- [Question-9] ROM 下载使用 CDN 时，VIP 校验如何实现？
  - 业界最佳实践：CDN 分发场景下鉴权的主流做法是「服务端签名 URL」：客户端先调服务端接口做业务鉴权（登录 + VIP 校验），通过后服务端生成带时效的 CDN 签名 URL 返回给客户端，客户端直接请求 CDN URL 下载，CDN 只校验 URL 签名不感知业务身份。
  - 推荐答案及理由：新增接口 `GET /api/v1/user/auth/license/game/download-url/:rom_name`（需登录），服务端完成：① 登录 token 校验；② VIP 游戏校验会员身份（同插件内 `checkUserMembership`）；③ 生成腾讯 CDN TypeA 签名 URL（有效期 2 小时）；④ 返回 `{url, expires_at}` 给 App。App 拿到 URL 后用 okhttp4 直接请求 CDN 下载，支持 Range 断点续传。download_log 在服务端签发 URL 时记录。
  [Answer-9]
  方案可以，CDN 为腾讯 CDN，签名算法已从 PHP 旧服务（`mcbtcola260907/html/addons/tradeorder/controller/Index.php::buildTypeASignUrl`）确认，见 Answer-10。

- [Question-10] CDN 厂商及签名算法？
  - 业界最佳实践：腾讯 CDN TypeA 鉴权算法，URL 格式：`{原始URL}?sign={timestamp}-{rand}-{uid}-{md5hash}`，其中 md5hash = MD5(`{path}-{timestamp}-{rand}-{uid}-{密钥}`)。
  - 推荐答案及理由：直接沿用旧 PHP 服务的签名算法（已验证可用），用 Go 重新实现：
  [Answer-10]
  腾讯 CDN，TypeA 鉴权算法，已从旧服务确认。Go 实现方案如下：

  ```go
  // GenerateCDNSignURL 生成腾讯 CDN TypeA 鉴权签名 URL
  // 算法来源：mcbtcola260907/html/addons/tradeorder/controller/Index.php::buildTypeASignUrl
  // URL 格式：{rawURL}?sign={timestamp}-{rand}-0-{md5(path-timestamp-rand-0-secret)}
  func GenerateCDNSignURL(rawURL, secret string, expireSecs int64) (signedURL string, expiresAt time.Time, err error) {
      u, err := url.Parse(rawURL)
      if err != nil {
          return "", time.Time{}, fmt.Errorf("无效 URL: %w", err)
      }

      now := time.Now().Unix()
      // rand：12 位随机字母数字（与旧服务行为一致）
      randStr := randAlphaNum(12)
      uid := "0"
      path := u.Path

      // md5(path-timestamp-rand-uid-secret)
      raw := fmt.Sprintf("%s-%d-%s-%s-%s", path, now, randStr, uid, secret)
      h := md5.New()
      io.WriteString(h, raw)
      hash := hex.EncodeToString(h.Sum(nil))

      signedURL = fmt.Sprintf("%s?sign=%d-%s-%s-%s", rawURL, now, randStr, uid, hash)
      expiresAt = time.Unix(now+expireSecs, 0)
      return signedURL, expiresAt, nil
  }

  // randAlphaNum 生成 n 位随机字母数字字符串
  func randAlphaNum(n int) string {
      const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
      b := make([]byte, n)
      for i := range b {
          b[i] = charset[rand.Intn(len(charset))]
      }
      return string(b)
  }
  ```

  依赖：`crypto/md5`、`encoding/hex`、`io`、`math/rand`、`net/url`、`fmt`、`time`（均标准库，零新增依赖）。

  配置项（写入 settings.yml）：
  - `cdn.secret`：固定字符串密钥（旧服务硬编码为 `cdntlping`，直接写 settings.yml 即可，无需 key 管理服务）
  - `cdn.domain`：CDN 域名（ROM 原始 URL 前缀）
  - `cdn.sign_expire_secs`：签名有效期秒数（建议 7200，即 2 小时）

  ROM 原始 URL 构造：`https://{cdn.domain}/roms/{rom_name}.zip`，由 handler 拼接后传入 `GenerateCDNSignURL`。
---

## 风险点

- [Risk-1] **games 表 DDL 变更兼容性**：现有 `games` 表已有数据，新增字段（category_id / cover_url / size_bytes / is_vip / parents / md5）必须有合理默认值，AutoMigrate 只加列不删列，存量数据不受影响。缓解：所有新增字段设 `DEFAULT NULL` 或合理默认值（is_vip DEFAULT 0），Go 端指针类型接收 NULL。

- [Risk-2] **CDN 签名 URL 有效期与大文件下载时长冲突**：腾讯 CDN TypeA 算法的 timestamp 是签发时间，CDN 侧通过后台配置的「时间窗口」（如 ±2 小时）判断是否有效，而非 URL 自带截止时间。只要 CDN 后台窗口 ≥ 单次下载耗时即可。缓解：服务端配置 `cdn_sign_expire_secs=7200`（2 小时），App 端在 `expires_at` 前 5 分钟重新获取 URL 并继续 Range 续传。

- [Risk-3] **App 端 SQLite 并发写**：DownloadManager 在后台线程更新下载进度，同时 UI 线程读取列表，需注意数据库并发访问。缓解：使用单例 SQLiteOpenHelper，写操作加事务，Room 本身已处理线程安全。

- [Risk-4] **下载目录权限（Android 11+）**：`getExternalFilesDir()` 不需要 WRITE_EXTERNAL_STORAGE 权限（API 29+），但需确认 minSdk 21 的低版本行为。缓解：使用 `getExternalFilesDir()` 而非 `Environment.getExternalStorageDirectory()`，前者无需运行时权限。

- [Risk-5] **MjarchGameListActivity 改造范围蔓延**：改造为 Tab 容器时可能无意间修改到共享基类逻辑。缓解：改动前先 grep 确认基类方法调用，所有新逻辑写在 mjarch flavor 子类 / Fragment 中，不改基类。
