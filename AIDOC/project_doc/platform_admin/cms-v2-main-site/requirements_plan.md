# 需求计划：CMS 主站内容管理系统 V2（三层架构重构 + 能力完善）

## 需求理解

**目标：** V1 CMS 以单张 `cms_site_config` 大表承载所有配置，字段堆积、内容固定、无缓存、无扩展点。V2 目标是将 CMS 重构为**三层架构**（内核层/内容层/能力层），使配置内聚职责清晰、内容灵活可扩展、能力可插拔配置，并完善 C 端性能优化和主流 CMS 功能。

**范围：**
- **后端（backend/app/cms/）**：拆分数据模型为三层；新增 `cms_integration` 集成配置表；公开 API 合并为单接口 `/public/cms/all`；Redis 缓存层（管理端保存主动失效）
- **管理端前端（dev-web-admin）**：菜单重构为「站点配置/页面管理/产品服务/导航菜单/留言/集成配置」；页面管理按 page+section_key 动态表单编辑
- **C 端前端（dev-web-user）**：统一通过 `/public/cms/all` 加载；前端内存+LocalStorage 两级缓存；骨架屏
- **第三方能力**：微信登录（H5 + PC 扫码）；百度/GA 统计脚本注入；广告位系统

**预期效果：**
- 管理员可在后台配置主站每个页面的每个区块内容，无需改代码
- 新增页面区块、调整内容结构时只需改配置，不改表结构
- C 端首次加载后续切页 0 网络请求（内存缓存），刷新后 LocalStorage 命中（<10ms）
- 第三方能力通过配置表启用/切换，不需要改代码重部署

**不在范围（本期不做）：**
- 动态内容类型（管理员自定义字段 Schema）
- 内容版本历史/回滚
- 多语言
- CDN 集成与自动刷新
- SSR/SSG 预渲染

---

## 假设列表

- [假设-1] V1 的 `cms_site_config` 表存有生产数据，需要写迁移脚本将现有数据迁移到新表结构，迁移完成后废弃旧表
- [假设-2] Redis 在现有生产环境已部署可用（platform_admin 其他模块已在使用）
- [假设-3] 微信登录复用现有 `biz_user` 表，通过 `openid` 字段关联（需确认字段是否存在）
- [假设-4] 集成配置中的敏感信息（appid/appsecret/密钥）在数据库中 AES-256 加密存储，密钥从环境变量读取
- [假设-5] 广告位系统本期只做自托管（不集成第三方广告平台），支持图片广告 + 文字链接
- [假设-6] 统计脚本注入（GA/百度）只需在前端 `<head>` 动态插入 script 标签，不需要后端做数据采集

---

## 澄清问题

- [Question-1] `cms_site_config` 数据迁移策略：现有生产数据如何平滑迁移到新三层表结构？
  - 业界最佳实践：主流做法是写迁移脚本（Go 工具脚本），从旧表读数据、转换后写入新表；新旧表并存一段时间，验证无误后再删旧表。Strapi V3→V4 迁移也是类似做法。
  - 推荐答案及理由：写 `tools/cms_migrate_v2.go` 迁移脚本，先建新表（AutoMigrate），跑脚本把 `cms_site_config` 数据分拆写入 `cms_site_identity` + `cms_page_section`，保留旧表直到验证完成。本项目数据量极小（1条记录），迁移风险极低。
  [Answer-1] 写迁移脚本 tools/cms_migrate_v2.go，建新表后从旧表转换写入，旧表保留至验证完成。
cms表不用迁移，未发布，但不要动其他表的数据
- [Question-2] `cms_page_section` 的 `config_json` schema 前端如何渲染管理表单？
  - 业界最佳实践：Contentful/Strapi 的做法是 Schema 也存数据库（完全动态），但实现成本高。另一种做法是 Schema 写在前端代码（按 section_key 分发），管理端动态渲染对应表单，这是 WordPress Customizer 的做法。
  - 推荐答案及理由：本期采用「Schema 在前端代码按 section_key 分发」，每个 section_key 对应一个 Vue 表单组件（如 `HeroSectionForm.vue`），通过 `section_key` 路由到对应组件。实现简单、类型安全，后续需要动态 Schema 时再单独做。
  [Answer-2] Schema 在前端按 section_key 分发，每个 section_key 对应一个 Vue 表单组件，按 key 路由渲染。
Contentful/Strapi 的做法是 Schema 也存数据库（完全动态），但提前自己把初始化数据配置做好进去
- [Question-3] 公开 API 合并为 `/public/cms/all` 时，是否包含产品列表？产品数量增多后是否有性能问题？
  - 业界最佳实践：Headless CMS（如 Contentful）通常对内容分级，高频静态内容（导航/配置）合并返回，低频大量内容（文章列表）单独分页接口。
  - 推荐答案及理由：`/public/cms/all` 包含 identity + nav_items + sections，**不包含**产品列表（产品列表单独 `/public/cms/products`，前端按需加载）。原因：产品列表可能增多，合并会增大 all 接口体积；导航和配置真正是高频且每页都需要的内容。
  [Answer-3] /public/cms/all 只包含 identity + nav_items + sections，产品列表单独 /public/cms/products 按需加载。
同意
- [Question-4] Redis 缓存的 Key 失效策略：管理端保存时主动 DEL，还是依赖 TTL 自动过期？
  - 业界最佳实践：主流做法是「写操作主动失效 + TTL 兜底」（Cache-Aside Pattern），既保证实时性又防止缓存永不过期。纯 TTL 导致最长 N 分钟数据陈旧，不适合内容管理场景。
  - 推荐答案及理由：管理端所有写操作（保存站点配置/页面区块/导航/产品）完成后调 `InvalidateCmsCache(tenantID)` 主动 DEL，Redis 同时设 5 分钟 TTL 兜底（防止 DEL 失败场景）。
  [Answer-4] 管理端所有写操作完成后主动 DEL Redis key，同时设 5 分钟 TTL 兜底。
同意
- [Question-5] 微信登录本期做 H5 还是 PC 扫码？两者都做还是只选一个？
  - 业界最佳实践：移动端 App 内打开网页用 H5 OAuth（code 换 token 流），PC 浏览器用扫码登录（轮询 state 状态）。两者 SDK 不同，H5 用公众号 appid，PC 扫码用开放平台 appid。
  - 推荐答案及理由：本期优先实现 **H5 微信登录**（移动端用户更多，且 btcola.com 主要通过微信分享传播）。PC 扫码列入后续 Phase，不在本次范围内。
  [Answer-5] 本期只做 H5 微信登录（移动端优先），PC 扫码列入后续 Phase。
同意
- [Question-6] 广告位系统本期是否需要投放时间控制和曝光/点击统计？
  - 业界最佳实践：完整的广告系统包含投放时间（start_at/end_at）、频次控制、曝光/点击统计、A/B 测试。但 MVP 阶段通常只需时间控制 + 基础统计。
  - 推荐答案及理由：本期实现「投放时间控制（start_at/end_at）+ 基础点击统计（click_count 计数）」。曝光统计（每次展示都打点）性能成本较高，本期不做，只做点击统计。A/B 测试不在本期范围。
  [Answer-6] 本期实现投放时间控制（start_at/end_at）+ 基础点击统计（click_count），不做曝光统计和 A/B 测试。
同意
- [Question-7] `biz_user` 表是否已有 `openid` 字段？微信登录绑定方式是新增字段还是单独关联表？
  - 业界最佳实践：主流做法是「第三方登录关联表」（user_social_auth），存 provider + provider_uid + user_id，一个用户可绑多个第三方账号，比直接在 user 表加字段更灵活。
  - 推荐答案及理由：新建 `biz_user_social_auth` 关联表（provider, provider_uid, user_id, created_at），不修改 `biz_user` 表结构。原因：未来可能支持多种第三方登录（Apple/Google），关联表扩展性更好；同时不破坏现有 `biz_user` 表结构，降低迁移风险。
  [Answer-7] 新建 `biz_user_social_auth` 关联表，`biz_user` 不加字段（表中无 openid 字段，有 ext_fields json 可留作其他扩展）。三方登录统一走关联表。
同意
- [Question-8] 集成配置（`cms_integration`）的管理端是单独菜单还是合并到「站点配置」中？
  - 业界最佳实践：Strapi/WordPress 将「集成/插件」单独放在一级菜单下（Settings → Integrations），与内容管理分开，职责更清晰。
  - 推荐答案及理由：独立为「能力中心」一级菜单，下有「集成配置」和「广告管理」子菜单。原因：集成配置涉及密钥等敏感信息，权限控制应独立；且集成的功能不属于内容管理范畴，职责分离更清晰。
  [Answer-8] 独立为「能力中心」一级菜单，下含「集成配置」和「广告管理」子菜单，与内容管理职责分离。
同意
---

## 非功能需求建议

- **性能**：`/public/cms/all` 接口 P99 < 50ms（Redis 命中时）；前端内存缓存命中 0ms；LocalStorage 命中 < 10ms
- **安全**：集成配置中的 appid/secret 等敏感字段 AES-256-GCM 加密存储；微信回调接口做 state 校验防 CSRF；管理端集成配置页面需要更高权限（单独 permission code）
- **多租户**：所有 CMS 表均按 tenant_id 隔离；Redis 缓存 key 含 tenant_id（`cms:public:{tenantID}`）
- **向后兼容**：V1 已有的公开 API（`/public/cms/site-config`、`/public/cms/nav-items`、`/public/cms/products`）在迁移期间保持可用，待前端切换到 `/public/cms/all` 后再废弃
- **幂等性**：所有管理端写操作幂等；迁移脚本幂等（重跑不产生重复数据）

---

## 影响范围预判

**后端（backend）：**
- `backend/app/cms/models/models.go` — 新增/修改模型
- `backend/app/cms/service/service.go` — 服务层重构
- `backend/app/cms/apis/handler.go` — 接口层重构
- `backend/app/cms/plugin.go` — 路由注册更新
- `backend/common/auth/auth.go` — 无需改动（插件注册机制不变）
- 新建：`backend/app/cms/models/integration.go`（集成配置模型）
- 新建：`backend/tools/cms_migrate_v2.go`（迁移脚本）

**管理端前端（dev-web-admin）：**
- `src/views/cms/CmsSettings.vue` → 拆分为多个 section 表单组件
- `src/views/cms/CmsProducts.vue` — 保留，小调整
- 新增：`src/views/cms/CmsPageManager.vue`（页面区块编辑器）
- 新增：`src/views/cms/CmsIntegrations.vue`（集成配置）
- 新增：`src/views/cms/CmsAds.vue`（广告管理）
- `src/router/static-routes.ts` — 菜单更新

**C 端前端（dev-web-user）：**
- `src/views/cms/HomePage.vue` / `ProductsPage.vue` / `AboutPage.vue` — 切换 API + 加缓存层
- 新增：`src/utils/cms-cache.ts`（统一缓存 composable）
- 新增：骨架屏组件

**可能副作用：**
- V1 的 `cms_site_config` 表字段在迁移完成前需保持可读（兼容期）
- 微信登录新增公开接口，需确认路由不被认证中间件拦截
