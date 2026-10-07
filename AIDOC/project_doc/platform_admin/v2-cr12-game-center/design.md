# 设计：mjarch 游戏中心（v2-cr12）

> 版本：1.0
> 日期：2026-10-04
> 前置：design_plan.md 已于 2026-10-04 完成，10 个澄清问题全部确认。

---

## 1. 技术方案概述

本 CR 分三层：

- **服务端**：在现有 license 插件内扩展——增强 `Game` 模型字段，新增分类/公告/Banner 三张表及 CRUD，新增 CDN 签名 URL 下载接口，新增菜单 seed。
- **App 端**：严格限定 `app/src/mjarch/` 目录，改造四 Tab 容器，新增 GameDetailActivity 与 DownloadManager，本地 SQLite 缓存游戏列表/下载任务/收藏。
- **管理端前端**：增强已有游戏管理页，新增分类/公告/Banner 三个页面。

---

## 2. 服务端设计

### 2.1 API 设计

#### 管理端接口（需登录 + 权限）

| 方法 | 路径 | 权限 code | 说明 |
|------|------|-----------|------|
| GET | /api/v1/admin/license/games | license:games | 游戏列表（分页，支持 name/categoryId 过滤） |
| POST | /api/v1/admin/license/games | license:games:create | 新增游戏 |
| PUT | /api/v1/admin/license/games/:id | license:games:update | 编辑游戏 |
| DELETE | /api/v1/admin/license/games/:id | license:games:delete | 删除游戏 |
| GET | /api/v1/admin/license/game-categories | license:game-categories | 分类列表 |
| POST | /api/v1/admin/license/game-categories | license:game-categories:create | 新增分类 |
| PUT | /api/v1/admin/license/game-categories/:id | license:game-categories:update | 编辑分类 |
| DELETE | /api/v1/admin/license/game-categories/:id | license:game-categories:delete | 删除分类（含游戏归 NULL 逻辑） |
| GET | /api/v1/admin/license/game-notices | license:notices | 公告列表 |
| POST | /api/v1/admin/license/game-notices | license:notices:create | 新增公告 |
| PUT | /api/v1/admin/license/game-notices/:id | license:notices:update | 编辑公告 |
| DELETE | /api/v1/admin/license/game-notices/:id | license:notices:delete | 删除公告 |
| GET | /api/v1/admin/license/game-banners | license:banners | Banner 列表 |
| POST | /api/v1/admin/license/game-banners | license:banners:create | 新增 Banner |
| PUT | /api/v1/admin/license/game-banners/:id | license:banners:update | 编辑 Banner |
| DELETE | /api/v1/admin/license/game-banners/:id | license:banners:delete | 删除 Banner |

#### C 端公开接口（无需登录）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/user/public/license/games | 游戏列表（全量 gzip） |
| GET | /api/v1/user/public/license/game-categories | 分类列表 |
| GET | /api/v1/user/public/license/game-notices | 公告列表（分页） |
| GET | /api/v1/user/public/license/game-banners | Banner 列表 |

#### C 端认证接口（需登录）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/user/auth/license/game-download-urls | 获取主 ROM + 依赖 ROM 的 CDN 签名下载 URL |

---

### 2.2 数据库设计

#### 2.2.1 games 表（增量变更）

现有字段保留，新增以下字段（AutoMigrate 只加列不删列，所有新字段 DEFAULT NULL 或合理默认值）：

```sql
ALTER TABLE games
  ADD COLUMN category_id  BIGINT       DEFAULT NULL     COMMENT '分类ID，NULL=未分类',
  ADD COLUMN cover_url    VARCHAR(500) DEFAULT NULL     COMMENT '封面图URL（外链）',
  ADD COLUMN size_bytes   BIGINT       DEFAULT NULL     COMMENT '文件体积（字节）',
  ADD COLUMN is_vip       TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否VIP游戏：0=免费 1=VIP',
  ADD COLUMN parents      JSON         DEFAULT NULL     COMMENT 'parent ROM依赖清单（JSON数组）',
  ADD COLUMN md5          VARCHAR(64)  DEFAULT NULL     COMMENT 'ROM zip md5校验值';
```

> 注意：`required_levels` 字段（JSON）现有含义为「所需会员等级 code 数组」，与新增 `is_vip` 字段存在语义重叠。**本 CR 保留 `required_levels`（向后兼容），新增 `is_vip` 作为简化标记**；`is_vip=1` 等价于 `required_levels` 非空。App 端以 `is_vip` 为准。

#### 2.2.2 game_categories 表（新增）

```sql
CREATE TABLE game_categories (
  id          BIGINT       NOT NULL AUTO_INCREMENT,
  tenant_id   BIGINT       NOT NULL,
  name        VARCHAR(100) NOT NULL COMMENT '分类名称',
  sort_order  INT          NOT NULL DEFAULT 0 COMMENT '排序序号，升序',
  status      TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '1=启用 0=禁用',
  created_at  DATETIME     NOT NULL,
  updated_at  DATETIME     NOT NULL,
  deleted_at  DATETIME     DEFAULT NULL,
  PRIMARY KEY (id),
  INDEX idx_tenant (tenant_id),
  INDEX idx_sort (tenant_id, status, sort_order)
);
```

#### 2.2.3 game_notices 表（新增）

```sql
CREATE TABLE game_notices (
  id           BIGINT       NOT NULL AUTO_INCREMENT,
  tenant_id    BIGINT       NOT NULL,
  title        VARCHAR(200) NOT NULL,
  content      TEXT         NOT NULL,
  status       TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '0=草稿 1=已发布',
  published_at DATETIME     DEFAULT NULL,
  created_at   DATETIME     NOT NULL,
  updated_at   DATETIME     NOT NULL,
  deleted_at   DATETIME     DEFAULT NULL,
  PRIMARY KEY (id),
  INDEX idx_tenant_status (tenant_id, status, published_at)
);
```

#### 2.2.4 game_banners 表（新增）

```sql
CREATE TABLE game_banners (
  id          BIGINT       NOT NULL AUTO_INCREMENT,
  tenant_id   BIGINT       NOT NULL,
  image_url   VARCHAR(500) NOT NULL COMMENT '图片URL',
  link        VARCHAR(500) DEFAULT NULL COMMENT '点击跳转链接或行为',
  sort_order  INT          NOT NULL DEFAULT 0,
  status      TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '1=启用 0=禁用',
  created_at  DATETIME     NOT NULL,
  updated_at  DATETIME     NOT NULL,
  deleted_at  DATETIME     DEFAULT NULL,
  PRIMARY KEY (id),
  INDEX idx_tenant_status (tenant_id, status, sort_order)
);
```

#### 2.2.5 game_download_logs 表（新增）

```sql
CREATE TABLE game_download_logs (
  id         BIGINT    NOT NULL AUTO_INCREMENT,
  tenant_id  BIGINT    NOT NULL,
  user_id    BIGINT    NOT NULL,
  rom_name   VARCHAR(100) NOT NULL,
  created_at DATETIME  NOT NULL,
  PRIMARY KEY (id),
  INDEX idx_tenant_user (tenant_id, user_id),
  INDEX idx_rom (rom_name)
);
```

---

### 2.3 核心逻辑

#### 2.3.1 Game 模型扩展（models.go）

```go
// Game 游戏条目（在现有结构体上追加字段）
type Game struct {
    // ... 现有字段保留 ...
    CategoryID *int64         `gorm:"column:category_id;default:null" json:"category_id,omitempty"`
    // 注意：CategoryID 使用 *int64 不加 ,string tag，前端以 null 表示未分类，>0 表示有分类
    // 数据库 AUTO_INCREMENT 保证分类 id 从 1 开始，不会出现 id=0 的情况
    CoverURL   *string        `gorm:"column:cover_url" json:"cover_url,omitempty"`
    SizeBytes  *int64         `gorm:"column:size_bytes" json:"size_bytes,omitempty"`
    IsVIP      int            `gorm:"column:is_vip;default:0" json:"is_vip"`
    Parents    datatypes.JSON `gorm:"column:parents;type:json" json:"parents,omitempty"`
    MD5        *string        `gorm:"column:md5" json:"md5,omitempty"`
}
```

#### 2.3.2 C 端游戏列表接口（PublicGames 扩展）

现有 `PublicGameItem` 只返回 `rom_name + required_levels`，扩展为全量字段：

```go
// PublicGameItem C端游戏列表项
type PublicGameItem struct {
    RomName     string   `json:"rom_name"`
    DisplayName string   `json:"display_name"`
    CategoryID  *int64   `json:"category_id,omitempty"`
    CoverURL    *string  `json:"cover_url,omitempty"`
    SizeBytes   *int64   `json:"size_bytes,omitempty"`
    IsVIP       int      `json:"is_vip"`
    Parents     []string `json:"parents,omitempty"`
    MD5         *string  `json:"md5,omitempty"`
}
```

响应头加 `Content-Encoding: gzip`，handler 层用 `compress/gzip` 压缩后写出。

#### 2.3.3 Banner 启用数量限制

启用 Banner 时，service 层先查当前租户已启用 Banner 数量，达 5 条则返回错误拒绝启用：

```go
// EnableBanner 启用 Banner，最多允许 5 条同时启用
func (s *GameContentService) EnableBanner(tenantID, id int64) error {
    var count int64
    s.db.Model(&licenseModels.GameBanner{}).
        Where("tenant_id = ? AND status = 1 AND deleted_at IS NULL", tenantID).
        Count(&count)
    if count >= 5 {
        return fmt.Errorf("最多同时启用 5 条 Banner，请先禁用其他 Banner")
    }
    return s.db.Model(&licenseModels.GameBanner{}).
        Where("id = ? AND tenant_id = ?", id, tenantID).
        Update("status", 1).Error
}
```

#### 2.3.4 分类删除逻辑

删除分类时，将该分类下所有游戏的 `category_id` 置为 NULL（同一事务）：

```go
// DeleteCategory 删除分类，同事务将游戏归入未分类（category_id = NULL）
func (s *GameContentService) DeleteCategory(tenantID, id int64) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 1. 将该分类下游戏的 category_id 置 NULL
        if err := tx.Model(&licenseModels.Game{}).
            Where("tenant_id = ? AND category_id = ?", tenantID, id).
            Update("category_id", nil).Error; err != nil {
            return fmt.Errorf("迁移游戏失败: %w", err)
        }
        // 2. 软删除分类
        if err := tx.Where("id = ? AND tenant_id = ?", id, tenantID).
            Delete(&licenseModels.GameCategory{}).Error; err != nil {
            return fmt.Errorf("删除分类失败: %w", err)
        }
        return nil
    })
}
```

#### 2.3.5 ROM 下载 URL 批量签名接口

```
POST /api/v1/user/auth/license/game-download-urls
```

**路径设计说明：** 遵循 license 插件现有 C 端认证接口规范（`RegisterAuthUserRoutes` 下统一挂 `/license/{资源名}`），无子目录层级，与 `/license/purchase`、`/license/activate` 等保持一致。

**请求体：**
```json
{ "rom_name": "mjsangen" }
```

**处理流程：**
```
1. JWT + X-Device-ID 中间件校验（复用现有）
2. 查 games 表确认 rom_name 存在且 status=1
3. 若 is_vip=1，调用同插件内 checkUserMembership 校验会员，非会员返回 403
4. 取 parents 字段（JSON 数组），得到所有依赖 ROM 名称列表
5. 对主 ROM + 所有 parents，各自生成腾讯 CDN 签名 URL
6. 写 game_download_logs（只记录主 ROM，不重复记录 parent）
7. 返回签名 URL 集合
```

**响应结构：**
```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "main": {
      "rom_name": "mjsangen",
      "url": "https://cdn.example.com/roms/mjsangen.zip?sign=...",
      "size_bytes": 1234567,
      "md5": "abc123",
      "expires_at": "2026-10-04T16:00:00+08:00"
    },
    "parents": [
      {
        "rom_name": "mjsangu",
        "url": "https://cdn.example.com/roms/mjsangu.zip?sign=...",
        "size_bytes": 890000,
        "md5": "def456",
        "expires_at": "2026-10-04T16:00:00+08:00"
      }
    ]
  }
}
```

> 若游戏无 parent 依赖，`parents` 返回空数组 `[]`。
> VIP 校验只针对主 ROM（`is_vip` 字段），parent ROM 不单独校验权限。

**CDN 文件路径规范（扁平化）：**
所有 ROM 文件路径统一为 `https://{cdn_domain}/roms/{rom_name}.zip`，与 MAME 的 rompath 机制天然一致（MAME 按 `rom_name.zip` 查找文件）。rom_name 即文件名，无需数据库额外存路径字段。

**CDN 签名算法（腾讯 CDN TypeA）：**

```go
// GenerateCDNSignURL 生成腾讯 CDN TypeA 鉴权签名 URL
// 来源：mcbtcola260907 PHP 旧服务 buildTypeASignUrl，已验证可用
// URL 格式：{rawURL}?sign={timestamp}-{rand}-0-{md5(path-timestamp-rand-0-secret)}
func GenerateCDNSignURL(rawURL, secret string, expireSecs int64) (string, time.Time, error) {
    u, err := url.Parse(rawURL)
    if err != nil {
        return "", time.Time{}, fmt.Errorf("无效 URL: %w", err)
    }
    now := time.Now().Unix()
    randStr := randAlphaNum(12)
    uid := "0"
    path := u.Path
    raw := fmt.Sprintf("%s-%d-%s-%s-%s", path, now, randStr, uid, secret)
    h := md5.New()
    _, _ = io.WriteString(h, raw)
    hash := hex.EncodeToString(h.Sum(nil))
    signedURL := fmt.Sprintf("%s?sign=%d-%s-%s-%s", rawURL, now, randStr, uid, hash)
    expiresAt := time.Unix(now+expireSecs, 0)
    return signedURL, expiresAt, nil
}
```

**配置项（settings.yml）：**
```yaml
game:
  cdn_secret: "cdntlping"       # 腾讯 CDN TypeA 密钥（固定字符串）
  cdn_domain: "your.cdn.domain" # CDN 域名
  cdn_sign_expire_secs: 7200    # 签名有效期 2 小时
```

---

### 2.4 新增菜单 Seed（plugin.go）

在现有「License 管理」一级菜单下追加三个子菜单（sort 11/12/13）：

```go
// 子菜单：游戏分类（新增）
{Name: "游戏分类", PermissionCode: "license:game-categories", Path: "/license/game-categories",
 Icon: "Menu", Sort: 11, ...},
// 子菜单：公告管理（新增）
{Name: "公告管理", PermissionCode: "license:notices", Path: "/license/game-notices",
 Icon: "Bell", Sort: 12, ...},
// 子菜单：Banner管理（新增）
{Name: "Banner管理", PermissionCode: "license:banners", Path: "/license/game-banners",
 Icon: "Picture", Sort: 13, ...},
```

ApiPermissions 同步追加对应的 ENDPOINT 权限项（CREATE/UPDATE/DELETE/LIST 各一条）。

---

### 2.5 文件结构（服务端）

```
backend/app/license/
├── models/
│   └── models.go                  ← 追加 Game 新字段 + 新增 GameCategory/GameNotice/GameBanner/GameDownloadLog
├── service/
│   ├── game_price_service.go      ← 扩展 GameInput/PublicGameItem + 现有游戏 CRUD 增量改动
│   └── game_content_service.go    ← 新增：分类/公告/Banner CRUD + CDN 签名 URL 逻辑
├── apis/
│   └── game_content_api.go        ← 新增：对应 handler（参照 membership_api.go 风格）
└── plugin.go                      ← 追加路由注册 + AutoMigrate + 菜单 Seed
```

---

## 3. App 端设计（mjarch flavor）

> **硬规则：所有改动严格限定在 `app/src/mjarch/` 目录，禁止触碰其他 flavor 及基类。**

### 3.1 本地数据库（SQLite）

新增 `GameDatabase`（原生 SQLiteOpenHelper，不引入 Room）：

```
数据库名：game_center.db
版本：1

表：
  game_list        — 游戏列表缓存（rom_name, display_name, category_id, cover_url, size_bytes, is_vip, md5, parents_json, ...）
  game_categories  — 分类缓存（id, name, sort_order）
  download_tasks   — 下载文件任务（job_rom_name, rom_name, is_main, url, expires_at,
                     total_bytes, downloaded_bytes, status, md5, updated_at）
  favorites        — 收藏（rom_name, added_at）

SharedPreferences（名：game_center_meta）：
  game_list_last_sync  — 游戏列表最后刷新时间戳（毫秒）
```

**download_tasks 表字段说明：**
- `job_rom_name`：所属游戏任务的主 ROM 名，用于聚合一个游戏的所有文件任务
- `rom_name`：当前文件任务对应的 ROM 名（主 ROM 或 parent ROM）
- `is_main`：1=主 ROM，0=parent ROM
- `status`：`PENDING=0 / DOWNLOADING=1 / PAUSED=2 / COMPLETED=3 / FAILED=4`

**parent ROM 共享策略：**
同一个 parent ROM 可能被多个游戏依赖（如 `mjsangu` 被多个麻将游戏共用）。download_tasks 表按 `(job_rom_name, rom_name)` 联合唯一，即每个游戏任务独立记录自己的 FileTask，**不共享**。但磁盘上的文件只下载一次：下载前先检查文件是否已存在且 md5 匹配，若已存在则直接标记该 FileTask 为 COMPLETED，跳过下载。

**清除游戏时 parent ROM 的处理：**
清除某游戏（主 ROM）时，对每个 parent ROM 文件，满足以下**任一条件**则保留文件，不删除：

1. **有其他已下载游戏把它作为 parent ROM 依赖**：查 download_tasks 表，其他 `job_rom_name` 下存在 `rom_name=该parent` 且 `status=COMPLETED` 的记录
2. **它本身是某个已下载游戏的主 ROM**：查 download_tasks 表，存在 `job_rom_name=该parent` 且 `is_main=1` 且 `status=COMPLETED` 的记录

两个条件均不满足时，才删除该 parent ROM 文件。

只删除被清除游戏专属的文件：主 ROM zip + 解压目录 + 该游戏的 cfg/nvram。

**游戏级下载状态由 download_tasks 聚合推导：**
- 所有 FileTask COMPLETED → 游戏已下载
- 任意 FileTask FAILED → 游戏下载失败
- 有 FileTask DOWNLOADING → 游戏下载中
- 有 FileTask PAUSED，无 DOWNLOADING → 游戏已暂停

App 启动时将所有 `DOWNLOADING` 状态的 FileTask 自动降级为 `PAUSED`，恢复队列。

### 3.2 模块结构

```
app/src/mjarch/java/.../gamecenter/
├── api/
│   └── GameCenterApi.java          — HTTP 接口封装（复用现有 MembershipApi 的 OkHttpClient）
├── db/
│   └── GameDatabase.java           — SQLiteOpenHelper 单例
├── model/
│   ├── GameItem.java               — 游戏列表 POJO
│   ├── GameCategory.java
│   ├── GameNotice.java
│   ├── GameBanner.java
│   └── DownloadTask.java
├── repository/
│   └── GameRepository.java         — 列表合并（服务端 + files.zip 内置）、收藏、下载记录
├── download/
│   └── DownloadManager.java        — okhttp4 断点续传队列（并发上限 2）
└── ui/
    ├── GameHallFragment.java       — 游戏大厅 Tab
    ├── GameDetailActivity.java     — 详情页（状态机按钮区）
    ├── FavoritesFragment.java      — 收藏夹 Tab
    ├── MyFragment.java             — 我的 Tab
    ├── NoticeListActivity.java     — 公告列表页
    └── DownloadedGamesFragment.java — 设置页已下载游戏管理（嵌入 SettingsFragment）
```

### 3.3 MjarchGameListActivity 改造

直接改造为四 Tab 容器（BottomNavigationView + FragmentManager）：

```
改造前：继承基类，setContentView 为游戏列表布局
改造后：setContentView 为四 Tab 布局，底栏切换 Fragment

Tab 索引：0=游戏大厅 1=我的 2=收藏夹 3=设置
Fragment 保活：show/hide 方式切换，不 replace（避免重新加载）
```

**关键约束：**
- 不修改基类（`BaseActivity` 等共享代码）任何方法
- 改动仅限 MjarchGameListActivity.java 本身 + 新增 Fragment/Activity
- 若需覆盖基类行为，在 mjarch 子类里 `@Override`，不改基类

### 3.4 DownloadManager 核心流程

每个「游戏下载任务（Job）」由一组 FileTask 组成（主 ROM + 所有 parent ROM）：

```
用户点击「下载」（GameDetailActivity）
  ↓
调用 POST /license/game-download-urls，body={ rom_name }
  → 服务端返回 main + parents[] 的签名 URL 集合
  ↓
DownloadManager.enqueueJob(mainRomName, main, parents[])
  → SQLite 写入多条 FileTask（job_rom_name=主ROM，先写 parents，再写 main）
  → 入队顺序：parents 先于 main（确保依赖先就位）
  ↓
工作线程（最多 2 个并发 FileTask，跨 Job）
  → 取队首 FileTask
  → 检查 expires_at，若距过期 ≤ 5 分钟则重新请求签名 URL 刷新
  → OkHttp Range 请求 CDN（从 downloaded_bytes 断点续传）
  → 每 500ms 回调进度 → 更新 SQLite downloaded_bytes
  → 下载完成 → md5 校验
      ✅ → FileTask status = COMPLETED
      ❌ → 删文件，FileTask status = FAILED，整个 Job 标记 FAILED，通知用户
  ↓
同一 Job 所有 FileTask COMPLETED → Job 完成
  → 通知 UI 刷新（Observer/Callback）→ 主按钮切换为「开始游戏」
```

**进度计算（用户侧整体进度）：**
```
总进度 = Σ(所有 FileTask.downloaded_bytes) / Σ(所有 FileTask.total_bytes)
```
用户看到的是整游进度（如 67%），不感知内部多文件。

**大厅游戏卡片状态角标 + 顶部进度提示：**
- 下载中的游戏卡片显示进度角标（百分比或转圈动画）
- 游戏大厅顶部展示一个细进度条（有活跃下载任务时显示），点击跳到对应游戏详情页
- 无活跃下载时进度条隐藏，不占空间

### 3.5 rompath 多路径

```java
// MjarchGameListActivity 或 GameRepository 中
String builtinDir = getFilesDir().getAbsolutePath() + "/roms"; // files.zip 解压目录
String downloadDir = getExternalFilesDir(Environment.DIRECTORY_DOWNLOADS) + "/roms";
String romPath = builtinDir + ";" + downloadDir;
// 传给 myosd 启动参数（现有 -rompath 机制已支持 ; 分隔）
```

---

## 4. 管理端前端设计

### 4.1 文件结构

```
dev-web-admin/src/
├── api/
│   └── license.ts          ← 追加游戏分类/公告/Banner 的 API 函数
├── views/license/
│   ├── Games.vue           ← 已有文件，增强新字段（category_id/cover_url/size_bytes/is_vip/parents/md5）
│   ├── GameCategories.vue  ← 新增：分类管理页
│   ├── GameNotices.vue     ← 新增：公告管理页
│   └── GameBanners.vue     ← 新增：Banner 管理页
└── router/
    └── static-routes.ts    ← 追加三个新页面路由（在 license 路由组内）
```

### 4.2 前后端字段对齐

| 后端 JSON tag | 前端字段 | 类型 | 说明 |
|---------------|----------|------|------|
| `id` | `id` | string | 雪花 ID，加 `json:",string"` tag，前端 string 类型 |
| `rom_name` | `romName` | string | |
| `display_name` | `displayName` | string | |
| `category_id` | `categoryId` | number \| null | *int64，null=未分类，>0=有效分类ID（非雪花ID，普通自增，前端用 number） |
| `cover_url` | `coverUrl` | string | |
| `size_bytes` | `sizeBytes` | number | |
| `is_vip` | `isVip` | number | 0/1 |
| `md5` | `md5` | string | |
| `sort_order` | `sortOrder` | number | |
| `status` | `status` | number | |
| `required_levels` | `requiredLevels` | string[] | JSON 数组 |

> 所有 BIGINT 类型 ID 字段后端必须加 `json:",string"` tag，前端对应 `string` 类型，防止 JS number 精度丢失。

### 4.3 Games.vue 增量改动

在现有表格和表单基础上：
- 表格新增列：分类（显示分类名，需 JOIN 或前端映射）、封面（缩略图）、体积、VIP 标记
- 编辑表单新增字段：分类（el-select，数据来自分类接口）、封面 URL、文件体积、is_vip（开关）、md5

**分类名称映射加载时机：** Games.vue 初始化时同时调用分类列表接口 `GET /api/v1/admin/license/game-categories`，将结果存为 `categoryMap: Record<string, string>`（key=id，value=name），表格列渲染时用此 map 转换显示名称。el-select 下拉也使用同一份分类列表数据。

---

## 5. 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 现有游戏管理接口（List/Create/Update/Delete）返回字段向后兼容 | 旧 App 版本调用 `/public/license/games` 仍能解析 |
| RG-2 | license 插件现有会员激活/登录/购买全链路不受影响 | 调用激活/登录/购买接口正常返回 |
| RG-3 | 其他 flavor（mjguoguan1/2、mjarch3）App 行为零变化 | 安装其他 flavor APK，游戏列表/启动正常 |
| RG-4 | 管理端现有激活码/会员/订单等页面不受影响 | 管理端各页面正常加载无报错 |
| RG-5 | mjarch App 现有会员激活、购买、登录流程不受影响 | 走一遍激活码激活 + 购买流程 |
| RG-6 | go build ./... 零错误 | 编译检查 |
| RG-7 | 前端 vite build 零错误 | 构建检查 |
| RG-8 | PublicGames 接口字段扩展后，旧版 App 调用不崩溃（JSON 新增字段旧端忽略） | 用旧版 App 调用新接口，游戏列表正常解析展示 |

---

## 6. 正确性属性

- **租户隔离**：所有新增表（game_categories / game_notices / game_banners / game_download_logs）的查询必须带 `tenant_id` 过滤。C 端接口通过 `ResolveTenantByHost()` 获取当前租户，不从请求参数接收。
- **VIP 下载鉴权服务端强制**：`is_vip=1` 的游戏，签名 URL 接口必须在服务端完成会员校验，不信任 App 端标记。
- **分类删除事务一致性**：删除分类与游戏归 NULL 必须在同一数据库事务中完成，不允许分步操作。
- **games 表向后兼容**：新增字段全部 DEFAULT NULL 或有合理默认值，AutoMigrate 不删现有列，现有 App 版本解析不报错。
- **CDN 签名安全**：`cdn_secret` 仅存 settings.yml（服务端配置），不下发给客户端，不写入数据库。
- **CDN 签名失败容错**：`GenerateCDNSignURL` 返回 error 时，handler 层返回 HTTP 500 + 用户友好 msg「下载链接生成失败，请稍后重试」，不暴露内部错误详情。
- **category_id 不使用 0 值**：分类表 AUTO_INCREMENT 从 1 开始，category_id=NULL 表示未分类，不存在 id=0 的分类记录，前端以 null/undefined 判断未分类。
- **App flavor 隔离**：mjarch flavor 的任何改动不得影响其他 flavor 的编译和运行时行为；共享基类不修改。
- **下载 md5 强制校验**：下载完成后必须校验 md5，失败时删除文件并通知用户，不允许静默忽略损坏文件。
- **parent ROM 引用计数保护**：清除某游戏时，对每个 parent ROM 文件，满足「被其他游戏作为 parent 依赖」或「本身是某已下载游戏的主 ROM」任一条件时必须保留文件，不允许删除；两个条件均不满足时才可删除。



## Web User 端路径规划（2026-10-06 确定）

所有 Web User 端页面统一在 `https://btcola.com` 域名下，路径规划如下：

| 路径 | 页面说明 | 认证要求 | 状态 |
|------|----------|----------|------|
| `/home` | 首页（游戏大厅） | 公开 | 已有 |
| `/bulletin` | 公告栏 | 公开 | 已有 |
| `/recommend` | 应用推荐 | 公开 | 已有 |
| `/login` | 登录/注册 | 公开 | 已有 |
| `/profile` | 个人中心 | 强制登录 | 占位 |
| `/app/upgrade` | App 升级页（展示版本信息+CDN签名下载） | 公开 | 新增 |
| `/games/:rom` | 游戏详情页 | 软鉴权 | 预留 |
| `/cs` | 在线客服（Web端） | 软鉴权 | 预留 |

**路径分组原则**：
- `/app/*` — App 相关（升级、下载）
- `/games/*` — 游戏内容相关
- `/cs/*` — 客服相关
- 顶级路径（`/home`、`/login`、`/profile`）— 核心导航页

**禁止随意新增顶级路径**，新功能按分组原则归入对应子路径。
