# 设计：CMS 主站内容管理系统 V2

## 技术方案

### 1. 数据库设计

#### 内核层（Core）

```sql
-- 站点标识（极少变，每个租户一条）
CREATE TABLE cms_site_identity (
  id          BIGINT AUTO_INCREMENT PRIMARY KEY,
  tenant_id   BIGINT NOT NULL,
  site_name   VARCHAR(128) NOT NULL DEFAULT '',
  logo_url    VARCHAR(512) DEFAULT '',
  favicon_url VARCHAR(512) DEFAULT '',
  icp         VARCHAR(128) DEFAULT '',
  theme_color VARCHAR(32) DEFAULT '#1a56e8',
  footer_text VARCHAR(256) DEFAULT '',
  created_at  DATETIME(3),
  updated_at  DATETIME(3),
  UNIQUE KEY uk_cms_identity_tenant (tenant_id)
);

-- 全局 SEO
CREATE TABLE cms_site_seo (
  id               BIGINT AUTO_INCREMENT PRIMARY KEY,
  tenant_id        BIGINT NOT NULL,
  title_template   VARCHAR(256) DEFAULT '%s | 千分软件',  -- %s 替换为页面标题
  description      VARCHAR(512) DEFAULT '',
  keywords         VARCHAR(512) DEFAULT '',
  og_image_url     VARCHAR(512) DEFAULT '',
  created_at       DATETIME(3),
  updated_at       DATETIME(3),
  UNIQUE KEY uk_cms_seo_tenant (tenant_id)
);
```

#### 内容层（Content）

```sql
-- 页面区块配置（核心）
CREATE TABLE cms_page_section (
  id          BIGINT AUTO_INCREMENT PRIMARY KEY,
  tenant_id   BIGINT NOT NULL,
  page        VARCHAR(64) NOT NULL,   -- home/about/products/contact
  section_key VARCHAR(64) NOT NULL,   -- hero/stats/features/contact/values
  config_json TEXT,                   -- 该区块所有配置 JSON
  sort_order  INT NOT NULL DEFAULT 0,
  is_enabled  TINYINT(1) NOT NULL DEFAULT 1,
  updated_at  DATETIME(3),
  UNIQUE KEY uk_cms_section (tenant_id, page, section_key)
);

-- 产品服务（已有 cms_product，保留）

-- 第三方登录关联表
CREATE TABLE biz_user_social_auth (
  id           BIGINT AUTO_INCREMENT PRIMARY KEY,
  tenant_id    BIGINT NOT NULL,
  user_id      BIGINT NOT NULL,
  provider     VARCHAR(32) NOT NULL,       -- wechat_h5/apple/google
  provider_uid VARCHAR(128) NOT NULL,      -- openid/sub/uid
  extra_json   TEXT,                       -- unionid/access_token等扩展
  created_at   DATETIME(3),
  UNIQUE KEY uk_social_auth (tenant_id, provider, provider_uid),
  KEY idx_social_auth_user (user_id)
);
```

#### 能力层（Capabilities）

```sql
-- 集成配置
CREATE TABLE cms_integration (
  id          BIGINT AUTO_INCREMENT PRIMARY KEY,
  tenant_id   BIGINT NOT NULL,
  type        VARCHAR(32) NOT NULL,     -- analytics/auth/ads/payment/notify/storage
  provider    VARCHAR(64) NOT NULL,     -- baidu/ga/wechat_h5/wechat_qr/...
  config_json TEXT,                     -- AES-256-GCM 加密的配置 JSON
  is_enabled  TINYINT(1) NOT NULL DEFAULT 0,
  created_at  DATETIME(3),
  updated_at  DATETIME(3),
  UNIQUE KEY uk_integration (tenant_id, type, provider)
);

-- 广告位定义
CREATE TABLE cms_ad_slot (
  id          BIGINT AUTO_INCREMENT PRIMARY KEY,
  tenant_id   BIGINT NOT NULL,
  slot_key    VARCHAR(64) NOT NULL,    -- home-banner/product-card/sidebar
  name        VARCHAR(128) NOT NULL,
  description VARCHAR(256) DEFAULT '',
  is_enabled  TINYINT(1) NOT NULL DEFAULT 1,
  UNIQUE KEY uk_ad_slot (tenant_id, slot_key)
);

-- 广告内容
CREATE TABLE cms_ad (
  id          BIGINT AUTO_INCREMENT PRIMARY KEY,
  tenant_id   BIGINT NOT NULL,
  slot_id     BIGINT NOT NULL,
  title       VARCHAR(128) NOT NULL,
  image_url   VARCHAR(512) DEFAULT '',
  link_url    VARCHAR(512) DEFAULT '',
  sort_order  INT NOT NULL DEFAULT 0,
  start_at    DATETIME(3),
  end_at      DATETIME(3),
  is_enabled  TINYINT(1) NOT NULL DEFAULT 1,
  click_count BIGINT NOT NULL DEFAULT 0,
  created_at  DATETIME(3),
  updated_at  DATETIME(3),
  KEY idx_ad_slot (slot_id)
);
```

---

### 2. API 设计

#### 公开接口（C 端，无需登录）

| 方法 | 路径 | 描述 | 缓存 |
|------|------|------|------|
| GET  | `/api/v1/public/cms/all` | 返回 identity+seo+nav_items+sections | Redis 5min |
| GET  | `/api/v1/public/cms/products` | 产品列表（visible） | Redis 5min |
| POST | `/api/v1/public/cms/contact` | 提交留言 | 无 |
| GET  | `/api/v1/public/cms/ads` | 广告位内容（?slot=xxx） | Redis 1min |
| POST | `/api/v1/public/cms/ads/:id/track` | 广告点击统计 | 无 |
| GET  | `/api/v1/user/auth/wechat/h5` | 微信H5登录跳转 | 无 |
| GET  | `/api/v1/user/auth/wechat/h5/callback` | 微信H5回调 | 无 |

`/public/cms/all` 响应结构：
```json
{
  "identity": { "site_name": "千分软件", "logo_url": "...", "theme_color": "#1a56e8" },
  "seo": { "title_template": "%s | 千分软件", "description": "..." },
  "nav_items": [{ "title": "首页", "path": "/home", "sort_order": 1 }],
  "sections": {
    "home": {
      "hero": { "badge_text": "...", "title": "...", "stats": [...] },
      "contact": { "title": "...", "email": "...", "phone": "..." }
    },
    "about": {
      "hero": { "title": "...", "subtitle": "..." },
      "stats": { "items": [...] },
      "values": { "items": [...] }
    }
  },
  "integrations": {
    "analytics": [{ "provider": "baidu", "script": "...", "enabled": true }]
  }
}
```

#### 管理端接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET/PUT | `/api/v1/admin/cms/identity` | 站点标识 |
| GET/PUT | `/api/v1/admin/cms/seo` | SEO 配置 |
| GET     | `/api/v1/admin/cms/sections` | 所有页面区块列表 |
| GET/PUT | `/api/v1/admin/cms/sections/:page/:key` | 单个区块配置 |
| GET/POST/PUT/DELETE | `/api/v1/admin/cms/products/:id` | 产品 CRUD（已有） |
| GET/POST/PUT/DELETE | `/api/v1/admin/cms/nav-items/:id` | 导航 CRUD（已有） |
| GET     | `/api/v1/admin/cms/contact-messages` | 留言列表（已有） |
| GET/POST/PUT | `/api/v1/admin/cms/integrations` | 集成配置 CRUD |
| GET/POST/PUT/DELETE | `/api/v1/admin/cms/ad-slots/:id` | 广告位 CRUD |
| GET/POST/PUT/DELETE | `/api/v1/admin/cms/ads/:id` | 广告 CRUD |

---

### 3. 缓存架构

```
写操作流程：
  PUT /admin/cms/* → Service 层写 DB → 成功后 Redis DEL cms:public:{tenantID}
                                                          DEL cms:products:{tenantID}

读操作流程（/public/cms/all）：
  Handler → Redis GET cms:public:{tenantID}
    ├── HIT → 直接返回（1-5ms）
    └── MISS → DB 查询 → Redis SET (TTL 5min) → 返回

前端缓存（utils/cms-cache.ts）：
  getCmsSiteData()
    ├── L1: memCache（Map，当次会话）→ HIT: 0ms
    ├── L2: localStorage['cms_site_data']（5min TTL）→ HIT: <10ms
    └── L3: fetch /public/cms/all → 写 L1+L2

降级：Redis 不可用时直接查 DB，TTL 标记失败不阻断请求
```

---

### 4. 微信 H5 登录流程

```
1. 用户点击「微信登录」
   → GET /api/v1/user/auth/wechat/h5
   → 后端生成 state（UUID，存 Redis 60s），返回微信授权 URL
   → 前端 redirect 到微信 OAuth 页

2. 用户授权，微信回调
   → GET /api/v1/user/auth/wechat/h5/callback?code=xxx&state=xxx
   → 后端：验证 state 有效 → 用 code 换 openid（调微信 API）
   → 查 biz_user_social_auth WHERE provider=wechat_h5 AND provider_uid=openid
     ├── 找到 → 取 user_id → 签发 JWT
     └── 未找到 → 创建 biz_user（nickname=微信昵称, avatar=头像）
                → 创建 biz_user_social_auth 记录
                → 签发 JWT

3. 前端收到 token，存 localStorage，完成登录

配置来源：cms_integration WHERE type=auth AND provider=wechat_h5
  config_json（解密后）: { "appid": "wx...", "appsecret": "...", "redirect_uri": "..." }
前提：该集成 is_enabled=true，否则接口返回 400「微信登录未启用」
```

---

### 5. 集成配置加密

```go
// 加密：保存时调用
func EncryptConfig(plaintext string) (string, error) {
    key := []byte(os.Getenv("CMS_INTEGRATION_KEY"))  // 32字节 AES-256
    // AES-256-GCM 加密，返回 base64(nonce + ciphertext)
}

// 解密：读取时调用
func DecryptConfig(ciphertext string) (string, error) { ... }

// 管理端展示脱敏
func MaskSensitiveFields(configJSON string) string {
    // appid/appsecret/secret 等字段只显示前4位 + ****
}
```

---

### 6. 管理端菜单结构（重构后）

```
内容管理
  ├── 站点配置（identity + seo，基础信息）
  ├── 页面管理（page_section，按 page+key 渲染动态表单）
  ├── 产品服务（cms_product，已有重构）
  ├── 导航菜单（cms_nav_item，已有）
  └── 用户留言（cms_contact_message，已有）

能力中心
  ├── 集成配置（cms_integration：analytics/auth/ads/notify）
  └── 广告管理（cms_ad_slot + cms_ad）
```

页面管理 UI（页面区块编辑器）：
```
首页(home)
  ├── [Hero 区] ← HeroSectionForm.vue
  ├── [特点区] ← FeaturesSectionForm.vue
  └── [联系区] ← ContactSectionForm.vue

关于页(about)
  ├── [Hero 区] ← PageHeroForm.vue（通用）
  ├── [统计区] ← StatsSectionForm.vue
  └── [价值观区] ← ValuesSectionForm.vue

产品页(products)
  └── [Hero 区] ← PageHeroForm.vue（复用）
```

---

### 7. 数据迁移脚本（tools/cms_migrate_v2.go）

```
执行步骤（幂等）：
1. 检查 cms_site_identity 是否已有数据 → 有则跳过
2. 从 cms_site_config 读取数据
3. 写入 cms_site_identity（site_name/logo_url/theme_color/footer_text/icp）
4. 写入 cms_site_seo（description/keywords）
5. 写入 cms_page_section（按字段分拆到对应 page+section_key）：
   - home/hero: hero_title/hero_subtitle/hero_badge_text/hero_cta_text/hero_cta_link/hero_stats_json
   - home/contact: contact_title/contact_email/contact_phone
   - about/hero: pages_hero_json.about
   - about/stats: about_stats_json
   - about/values: values_json
   - products/hero: pages_hero_json.products
6. 打印迁移摘要
```

---

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 现有 biz_user 登录（手机号/邮箱）不受影响 | POST /user/auth/login 正常返回 token |
| RG-2 | 现有 `/public/cms/products` 接口在迁移期间继续可用 | GET 返回 200 + 产品列表 |
| RG-3 | 现有 `/public/cms/nav-items` 接口在迁移期间继续可用 | GET 返回 200 + 导航列表 |
| RG-4 | cms_product/cms_nav_item/cms_contact_message 现有数据不丢失 | 迁移后记录数一致 |
| RG-5 | 管理端产品管理、导航管理、留言查看功能不受影响 | CRUD 操作正常 |
| RG-6 | 多租户隔离不被破坏 | 不同 tenant_id 数据不互通 |

## 正确性属性

- 所有 CMS 公开接口写入 `Cache-Control: public, max-age=300` 响应头
- 管理端写操作成功后必须调 `InvalidateCmsCache(tenantID)` 失效缓存
- 集成配置中含 appid/appsecret/secret/key 的字段在存入前必须加密，读取后展示前必须脱敏
- 微信回调必须验证 state 参数（Redis 中存在且未过期），防止 CSRF
- `biz_user_social_auth` 的 (tenant_id, provider, provider_uid) 组合唯一，防重复绑定
