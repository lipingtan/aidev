# CR-CMS：主站内容管理系统插件

## 状态

- **Phase**: 实现中（backend + admin + user 骨架已完成）
- **分支**: `feature/license-phase2`
- **关键 commit**: `1b199c3`

---

## 目标

在现有 platform_admin 架构中以**内置插件**方式新增 CMS 模块，实现对 btcola.com 主站内容的可视化管理。

主站用户（`dev-web-user`）与 mjarch 使用同一租户用户体系，`biz_user` 表，登录互通。

---

## 原型参考

三套风格（已生成 HTML 文件，可直接浏览器打开）：

| 方案 | 风格 | 文件 |
|------|------|------|
| A | 科技感深色（Dark Tech）| `C:\data\run\kirocli\tmp\cms_proto_a.html` |
| B | 现代简约白色（Modern Minimal）| `C:\data\run\kirocli\tmp\cms_proto_b.html` |
| C | 商务渐变（Business Gradient）⭐ 已实现 | `C:\data\run\kirocli\tmp\cms_proto_c.html` |

---

## 已实现内容

### 后端（`backend/app/cms/`）

| 文件 | 说明 |
|------|------|
| `models/models.go` | 5张表，全部有 `TableName()` |
| `service/service.go` | 完整 CRUD 服务 |
| `apis/handler.go` | 公开接口 + 管理端接口 |
| `plugin.go` | 内置插件，实现 `proto.PluginService` |

**数据库表**（GORM AutoMigrate 自动建表）：
- `cms_site_config` — 网站基础配置（名称、Logo、Hero、主题色、SEO）
- `cms_nav_item` — 导航菜单项
- `cms_product` — 产品/服务（含下载链接、说明链接、功能标签、指标亮点）
- `cms_team_member` — 团队成员
- `cms_contact_message` — 用户联系留言

**API 路由**：
```
公开（无需登录）：
  GET /api/v1/public/cms/site-config
  GET /api/v1/public/cms/nav-items
  GET /api/v1/public/cms/products
  GET /api/v1/public/cms/team-members
  POST /api/v1/public/cms/contact

管理端（需登录）：
  GET/PUT /api/v1/admin/cms/site-config
  GET/POST/PUT/DELETE /api/v1/admin/cms/nav-items/:id
  GET/POST/PUT/DELETE /api/v1/admin/cms/products/:id
  GET/POST/PUT/DELETE /api/v1/admin/cms/team-members/:id
  GET /api/v1/admin/cms/contact-messages
  PUT /api/v1/admin/cms/contact-messages/:id/read
```

### 管理端（`dev-web-admin/src/views/cms/CmsManager.vue`）

5个 Tab：
1. **基础配置** — 网站名称、Logo、Hero 区、主题色、SEO、页脚
2. **导航菜单** — 增删改排序，显示/隐藏，新标签开关
3. **产品管理** — 卡片式列表，完整表单（下载链接、说明链接、功能标签JSON、指标JSON）
4. **团队成员** — 头像、姓名、职位、简介
5. **用户留言** — 联系表单提交记录，标记已读

路由已注册到 `static-routes.ts`（`/cms/*`）。

### 主站前端（`dev-web-user/src/views/cms/`）

| 文件 | 路由 | 说明 |
|------|------|------|
| `HomePage.vue` | `/home` | 首页：Hero + 产品预览 + 团队 + 联系表单 |
| `ProductsPage.vue` | `/products` | 产品详情页：分类筛选 + 完整产品列表 |
| `AboutPage.vue` | `/about` | 关于页：公司介绍 + 团队成员 |
| `components/CmsNav.vue` | — | 共用导航栏组件 |
| `components/CmsFooter.vue` | — | 共用 Footer 组件 |

所有页面：
- `standalone: true`（不套侧边栏 layout，独立全屏展示）
- 响应式（768px 断点，移动端汉堡菜单）
- 数据从 `/api/v1/public/cms/*` 动态加载
- 主题色由 `cms_site_config.theme_color` 控制（CSS 变量注入）

---

## 技术约束

1. **插件方式**：`backend/app/cms/`，参照 bulletin/license 结构
2. **多租户**：所有数据按 `tenant_id` 隔离
3. **用户体系**：共用 `biz_user` 表，主站登录与 mjarch 互通
4. **GORM 表名**：所有 model 显式声明 `TableName()`
5. **雪花ID**：无雪花ID字段（CMS 用自增ID，不存在精度问题）
6. **响应式**：所有页面支持 320px~1440px

---

## 待完成事项

### Phase 1（已完成 ✅）
- [x] 后端 CMS 插件骨架
- [x] 5张数据表 + 完整 API
- [x] admin 管理页面（CmsManager.vue）
- [x] user 主站页面（HomePage + ProductsPage + AboutPage）
- [x] 共用导航/Footer 组件
- [x] go build + vite build 验证通过

### Phase 2（待做）
- [ ] admin 菜单注册到后端 plugin syncer（让菜单自动出现在管理端左侧菜单）
- [ ] 初始化种子数据（默认导航菜单项、示例产品）
- [ ] 产品封面图走 media 模块上传（图片选择器接入）
- [ ] 联系表单邮件通知（可选）
- [ ] E2E 测试（`cms.api.spec.ts`）

---

## 待确认

1. **风格选择**：用方案C还是希望调整？
2. **初始导航**：默认菜单项是哪些（首页/产品/关于/联系）？
3. **产品图片**：是否接入现有 media 模块上传？
