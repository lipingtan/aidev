# CR-CMS-V2：主站 CMS 架构重构与能力完善

## 状态
- **版本**: V2（V1 基础骨架已完成，commit `b5e9202`）
- **分支**: `feature/cms-v2`（从 `feature/license-phase2` 切出）
- **目标**: 内核层稳定内聚、内容层灵活扩展、能力层可插拔配置

---

## 背景与问题

V1 设计缺陷：
1. `cms_site_config` 一张表承载所有配置，字段堆积，每次加功能改表结构
2. 产品/内容字段固定，无法在不改代码的情况下调整结构
3. 无缓存层，每次页面加载都打 DB，4 个串行/并行请求
4. 第三方能力（广告/分析/登录/存储）硬编码，不可配置切换
5. 各页面存在大量硬编码文案，未完全配置化

---

## 架构设计

### 三层模型

```
┌─────────────────────────────────────────┐
│  能力层 Capabilities（可插拔）           │
│  广告 / 分析统计 / 第三方登录            │
│  支付集成 / OSS存储 / 推送 / SEO工具     │
├─────────────────────────────────────────┤
│  内容层 Content（灵活有结构）            │
│  产品服务  导航菜单  页面区块            │
│  媒体文件  留言表单  草稿/发布           │
├─────────────────────────────────────────┤
│  内核层 Core（稳定不变）                 │
│  站点标识  主题外观  SEO基础             │
│  路由配置  全局变量  租户隔离            │
└─────────────────────────────────────────┘
```

### 数据库表设计

#### 内核层（Core）

```sql
-- 站点标识（极少变）
cms_site_identity
  id, tenant_id(unique), site_name, logo_url, favicon_url,
  icp, theme_color, footer_text, updated_at

-- 全局 SEO
cms_site_seo
  id, tenant_id(unique), title_template, description, keywords,
  og_image_url, robots_txt, updated_at
```

#### 内容层（Content）

```sql
-- 页面区块配置（核心灵活化设计）
cms_page_section
  id, tenant_id, page(varchar:home/about/products/contact),
  section_key(varchar:hero/stats/features/contact/banner等),
  config_json(text, 该区块所有配置),
  sort_order, is_enabled, updated_at
  UNIQUE(tenant_id, page, section_key)

-- 产品服务（已有，保留）
cms_product: id, tenant_id, title, subtitle, description(text),
  icon_emoji, cover_url, download_url, detail_url,
  badge_text, badge_color, features_json(text), metrics_json(text),
  category, sort_order, is_visible, status(draft/published)

-- 导航菜单（已有，保留）
cms_nav_item: id, tenant_id, title, path, icon,
  sort_order, is_visible, open_new_tab

-- 联系留言（已有，保留）
cms_contact_message: id, tenant_id, name, email,
  subject, message(text), is_read, created_at
```

#### 能力层（Capabilities）

```sql
-- 集成配置（统一入口）
cms_integration
  id, tenant_id, type(storage/analytics/auth/ads/payment/notify/seo),
  provider(tencent_cos/aliyun_oss/wechat/google_analytics/baidu/...),
  config_json(text, AES加密存储), is_enabled,
  created_at, updated_at
  UNIQUE(tenant_id, type, provider)

-- 广告位定义
cms_ad_slot
  id, tenant_id, slot_key(home-top/product-card/sidebar),
  name, description, is_enabled

-- 广告内容
cms_ad
  id, tenant_id, slot_id, title, image_url, link_url,
  start_at, end_at, sort_order, is_enabled,
  impressions, clicks  -- 统计
```

### page_section config_json Schema（按 section_key 定义）

```json
// page=home, section_key=hero
{
  "badge_text": "全新 AI 产品矩阵正式发布",
  "title": "智能科技，赋能未来",
  "subtitle": "专注于人工智能与互联网应用开发...",
  "cta_text": "了解产品",
  "cta_link": "/products",
  "stats": [{"val":"7+","label":"年行业经验"},{"val":"100万+","label":"服务用户"}]
}

// page=home, section_key=features
{
  "title": "核心优势",
  "subtitle": "为什么选择千分软件",
  "items": [
    {"icon":"🧠","title":"大模型驱动","desc":"基于最新 LLM 技术"},
    {"icon":"🔒","title":"企业级安全","desc":"私有化部署选项"}
  ]
}

// page=contact, section_key=info
{
  "title": "准备好开始了吗？",
  "desc": "欢迎随时联系我们的团队",
  "email": "business@btcola.com",
  "phone": "400-888-9999",
  "address": "深圳市南山区"
}

// page=about, section_key=hero
{
  "title": "用技术创新，创造真实价值",
  "subtitle": "专注于人工智能与互联网应用开发..."
}

// page=about, section_key=stats
{
  "items": [{"val":"7+","label":"年行业经验"},{"val":"100万+","label":"服务用户"}]
}

// page=about, section_key=values
{
  "items": [{"icon":"🎯","title":"技术驱动","desc":"持续投入 AI 研发"}]
}
```

### API 设计

#### 公开接口（C端，无需登录）

```
GET /api/v1/public/cms/all
  → 一次返回：identity + seo + nav_items + products(visible) + sections(all pages)
  → Redis 缓存 5分钟，Cache-Control: public, max-age=300
  → 管理端保存时主动 InvalidateCache

GET /api/v1/public/cms/ads?slot=home-top   -- 广告位（后续）
POST /api/v1/public/cms/contact            -- 留言提交（已有）
POST /api/v1/public/cms/ads/:id/track      -- 广告点击/曝光统计（后续）
```

#### 管理端接口

```
-- 内核层
GET/PUT /api/v1/admin/cms/identity
GET/PUT /api/v1/admin/cms/seo

-- 内容层：页面区块
GET  /api/v1/admin/cms/sections?page=home
PUT  /api/v1/admin/cms/sections/:page/:section_key
POST /api/v1/admin/cms/sections/batch   -- 批量保存

-- 内容层：产品（已有）
GET/POST/PUT/DELETE /api/v1/admin/cms/products/:id

-- 内容层：导航（已有）
GET/POST/PUT/DELETE /api/v1/admin/cms/nav-items/:id

-- 能力层：集成配置
GET/POST/PUT /api/v1/admin/cms/integrations
GET          /api/v1/admin/cms/integrations/:type  -- 按类型查

-- 能力层：广告（后续）
CRUD /api/v1/admin/cms/ad-slots
CRUD /api/v1/admin/cms/ads
```

### 缓存架构

```
C端请求链路：
  浏览器 LocalStorage (5min TTL)
    └─ HIT → 直接渲染（0ms）
    └─ MISS → GET /public/cms/all
                └─ Redis (5min TTL)
                     └─ HIT → 返回（1-5ms）
                     └─ MISS → DB 查询 → 写 Redis → 返回

管理端保存触发：
  PUT /admin/cms/* → 业务逻辑 → InvalidateCmsCache(tenantID)
                                 └─ Redis DEL cms:public:{tenantID}
                                    LocalStorage 靠 TTL 自然过期（max 5min）
```

### 前端缓存实现

```typescript
// utils/cms-cache.ts
const CACHE_KEY = 'cms_site_data'
const TTL_MS = 5 * 60 * 1000  // 5分钟

export async function getCmsSiteData() {
  // L1: 内存（当次会话，切页面复用）
  if (memCache && Date.now() < memCache.expireAt) return memCache.data

  // L2: LocalStorage（跨刷新）
  try {
    const raw = localStorage.getItem(CACHE_KEY)
    if (raw) {
      const { data, expireAt } = JSON.parse(raw)
      if (Date.now() < expireAt) {
        memCache = { data, expireAt }
        return data
      }
    }
  } catch {}

  // L3: 网络请求
  const res = await fetch('/api/v1/public/cms/all')
  const { data } = await res.json()
  const expireAt = Date.now() + TTL_MS
  memCache = { data, expireAt }
  try { localStorage.setItem(CACHE_KEY, JSON.stringify({ data, expireAt })) } catch {}
  return data
}

export function invalidateCmsCache() {
  memCache = null
  localStorage.removeItem(CACHE_KEY)
}
```

---

## 登录集成设计

### 微信登录（国内主流）

```
移动端 H5（微信内打开）：
  用户点击「微信登录」
  → 跳转微信授权 https://open.weixin.qq.com/connect/oauth2/...
  → 微信回调 /api/v1/user/auth/wechat/h5/callback?code=xxx
  → 后端用 code 换 openid/unionid
  → 查/创建 biz_user（关联 openid）
  → 返回 JWT token，前端存储并跳转

PC 扫码（网站内嵌）：
  后端生成含 state 的二维码 URL
  → 前端展示二维码 + 轮询 /api/v1/user/auth/wechat/qr/status?state=xxx
  → 用户微信扫码授权
  → 后端收回调，更新 state 对应的 token
  → 前端轮询拿到 token，完成登录

配置存 cms_integration 表：
  type=auth, provider=wechat
  config_json: {"appid":"wx...","appsecret":"...","redirect_uri":"..."}
```

### 其他登录（扩展）

```
同样通过 cms_integration 配置：
  provider=apple    → Sign in with Apple（海外/iOS）
  provider=google   → Google OAuth2（海外）
  provider=phone    → 手机验证码（已有 biz_user 体系）
```

---

## 管理端 UI 设计

### 菜单结构（重构后）

```
内容管理
  ├── 站点配置（identity + seo，基础信息）
  ├── 页面管理（page_section，可视化区块编辑）
  ├── 产品服务（cms_product）
  ├── 导航菜单（cms_nav_item）
  └── 用户留言（cms_contact_message）

能力中心
  ├── 集成配置（cms_integration：分析/广告/登录/存储）
  └── 广告管理（cms_ad_slot + cms_ad）[后续]
```

### 页面管理：区块编辑器

管理端「页面管理」按页面+区块展示，每个区块展开对应的表单：

```
首页(home)
  ├── [Hero 区] badge_text / title / subtitle / cta / stats
  ├── [特点区] title / subtitle / items[]
  └── [联系区] title / desc / email / phone

关于页(about)
  ├── [Hero 区] title / subtitle
  ├── [统计区] items[]
  └── [价值观区] items[]

产品页(products)
  └── [Hero 区] title / sub
```

---

## 实现计划

### Phase 1：内核重构（1-2天）

- [ ] 新建 `cms_site_identity` 表（从 cms_site_config 迁移）
- [ ] 新建 `cms_page_section` 表（替换所有页面硬编码字段）
- [ ] 新建 `cms_integration` 表（能力层入口）
- [ ] 合并公开 API → `GET /public/cms/all`
- [ ] 后端 Redis 缓存层（5min TTL + 主动失效）
- [ ] 前端统一缓存 composable（内存 + LocalStorage）
- [ ] 迁移现有数据（cms_site_config → identity + sections）

### Phase 2：管理端重构（1天）

- [ ] 管理端拆分：站点配置 / 页面管理 / 产品 / 导航 / 留言
- [ ] 页面管理：按 page+section_key 渲染动态表单
- [ ] 保存时主动失效 Redis 缓存

### Phase 3：能力层（按优先级逐步实现）

- [ ] 集成配置 CRUD（加密存储 config_json）
- [ ] 微信登录集成（H5 + PC 扫码）
- [ ] 百度/GA 统计 script 注入
- [ ] 广告位系统
- [ ] Webhook 推送（新留言 → 企业微信）

### Phase 4：内容层完善

- [ ] 富文本编辑器（产品描述/公司简介）
- [ ] 产品草稿/发布状态
- [ ] 媒体上传直接对接 media 模块
- [ ] 内容版本历史

### Phase 5：性能与 SEO

- [ ] 各页面独立 `<title>/<meta>` 动态注入
- [ ] Sitemap.xml 自动生成
- [ ] Nginx Cache-Control 头
- [ ] 骨架屏

---

## 待确认

1. `cms_site_config` 是否直接废弃还是保留兼容？建议直接废弃，做一次性迁移
2. 微信登录优先 H5 还是 PC 扫码？
3. 集成配置的 `config_json` 加密方式：AES-256-GCM，密钥从环境变量读
4. Phase 1 完成后是否需要保证现有生产数据可用（迁移脚本）
