# 需求：CMS 主站内容管理系统 V2

## 背景

V1 CMS 以单张大表承载所有配置，字段堆积、无扩展点、无缓存。V2 重构为三层架构，使系统配置内聚、内容灵活扩展、能力可插拔，同时完善 C 端性能优化和主流 CMS 核心功能。

## 用户故事

- 作为**运营人员**，我希望在管理后台配置主站每个页面每个区块的内容（标题/副标题/统计数字/联系方式等），以便无需改代码即可更新官网内容
- 作为**运营人员**，我希望管理产品服务列表（含下载链接/说明链接/功能标签），以便随时上下架产品介绍
- 作为**运营人员**，我希望在「能力中心」配置第三方集成（微信登录/统计脚本/广告），以便灵活启用/切换外部能力
- 作为**C 端用户**，我希望主站页面加载快、切换流畅，以便获得良好的浏览体验
- 作为**C 端用户**，我希望使用微信登录主站（H5），以便无需注册即可快速访问登录态功能

---

## 功能需求

### FR-1: 三层数据模型重构

**描述：** 将 `cms_site_config` 拆分为三层独立表结构，每层职责单一。

**验收标准：**
- WHEN 系统启动 THEN 自动 AutoMigrate 创建 `cms_site_identity`、`cms_site_seo`、`cms_page_section`、`cms_integration`、`cms_ad_slot`、`cms_ad` 表
- WHEN 运行迁移脚本 THEN 旧 `cms_site_config` 数据完整迁移到新表，无数据丢失
- WHEN 查询新表 THEN 每个表只含本层职责的字段（内核/内容/能力互不混用）

### FR-2: 页面区块配置（Page Section）

**描述：** 主站每个页面（home/about/products）的每个区块（hero/stats/features/contact等）独立配置，存储为 JSON，管理端按 section_key 渲染对应表单。

**验收标准：**
- WHEN 管理员访问「页面管理」 THEN 看到按页面分组的区块列表（home/about/products）
- WHEN 管理员编辑 home/hero 区块 THEN 表单显示该区块专属字段（badge_text/title/subtitle/cta/stats[]）
- WHEN 管理员保存区块配置 THEN config_json 更新，Redis 缓存失效，C 端下次请求获取最新内容
- WHEN C 端加载首页 THEN Hero 区标题/副标题/统计数字/CTA 全部来自数据库，无硬编码

### FR-3: 公开 API 合并与缓存

**描述：** C 端通过单一接口 `/public/cms/all` 获取 identity + nav_items + sections，后端 Redis 缓存 5 分钟，前端内存+LocalStorage 两级缓存。

**验收标准：**
- WHEN C 端首次加载 THEN 只发一个 `/public/cms/all` 请求（不含产品列表）
- WHEN C 端同一会话内切换页面 THEN 不发网络请求（内存缓存命中）
- WHEN C 端刷新页面后 5 分钟内 THEN LocalStorage 命中，不发网络请求
- WHEN 管理员保存任意配置 THEN Redis `cms:public:{tenantID}` 立即失效
- WHEN Redis 不可用 THEN 降级直接查 DB，不报错
- WHEN `/public/cms/all` Redis 命中 THEN 接口响应时间 < 50ms（P99）

### FR-4: 产品服务管理

**描述：** 产品列表 CRUD，含下载链接、说明链接、功能标签、指标亮点、草稿/发布状态。

**验收标准：**
- WHEN 管理员创建产品 THEN 可填写：名称/副标题/描述/图标/封面URL/下载链接/说明链接/功能标签JSON/指标JSON/分类/排序/显示状态
- WHEN 产品 is_visible=false THEN C 端产品列表不返回该产品
- WHEN C 端请求 `/public/cms/products` THEN 返回 is_visible=true 的产品列表，按 sort_order 升序

### FR-5: 集成配置管理

**描述：** 通过 `cms_integration` 表配置第三方集成（统计/广告/登录），敏感 config_json AES-256-GCM 加密存储。

**验收标准：**
- WHEN 管理员访问「集成配置」 THEN 看到按类型分组的集成列表（analytics/auth/ads/notify）
- WHEN 管理员启用百度统计集成（type=analytics, provider=baidu）并填入 site_id THEN C 端 `<head>` 自动注入百度统计 script
- WHEN 管理员启用微信登录集成（type=auth, provider=wechat_h5）并填入 appid/appsecret THEN 主站登录页出现「微信登录」按钮
- WHEN 保存集成配置 THEN config_json 中的 appid/appsecret 字段在数据库中 AES-256-GCM 加密存储
- WHEN 读取集成配置（管理端展示） THEN 敏感字段脱敏显示（仅显示前4位+星号）

### FR-6: 微信 H5 登录

**描述：** 主站 C 端支持微信网页授权登录（移动端 H5 模式），绑定关系存 `biz_user_social_auth` 关联表。

**验收标准：**
- WHEN 用户在微信内访问主站并点击「微信登录」 THEN 跳转微信 OAuth 授权页
- WHEN 用户授权后微信回调 `/api/v1/user/auth/wechat/h5/callback` THEN 系统用 code 换取 openid，查找或创建 biz_user，返回 JWT token
- WHEN 同一微信用户二次登录 THEN 通过 `biz_user_social_auth` 关联找到已有 biz_user，不重复创建账号
- WHEN 微信登录未在集成配置中启用 THEN 登录页不显示「微信登录」按钮
- WHEN 微信 OAuth 失败（code 无效/过期）THEN 返回友好错误提示，不暴露内部错误

### FR-7: 广告位系统

**描述：** 自托管广告位，支持投放时间控制和点击统计。

**验收标准：**
- WHEN 管理员创建广告位（slot_key: home-banner） THEN C 端可通过 `/public/cms/ads?slot=home-banner` 获取当前有效广告
- WHEN 广告 start_at 未到或 end_at 已过 THEN 该广告不在 C 端返回结果中
- WHEN 用户点击广告 THEN POST `/public/cms/ads/:id/track`，click_count+1
- WHEN 广告位无有效广告 THEN 接口返回空数组，不报错

### FR-8: 骨架屏

**描述：** C 端页面在 CMS 数据加载前显示骨架屏，避免白屏。

**验收标准：**
- WHEN 首次加载 `/home` 且无缓存 THEN 显示骨架屏（Hero 区/产品卡片区）直至数据返回
- WHEN 数据返回后 THEN 骨架屏平滑替换为真实内容（transition 过渡）

---

## 非功能需求

- **性能**：`/public/cms/all` Redis 命中时 P99 < 50ms；前端内存缓存命中 0ms；LocalStorage 命中 < 10ms
- **安全**：集成配置敏感字段 AES-256-GCM 加密；微信回调 state 防 CSRF；管理端集成配置独立权限（`cms:integrations:manage`）
- **多租户**：所有 CMS 表按 tenant_id 隔离；Redis key 含 tenant_id
- **向后兼容**：迁移期间旧版 `/public/cms/site-config` 等接口继续可用，前端切换完成后再废弃
- **幂等性**：迁移脚本重跑不产生重复数据；所有写操作幂等
