# 前端页面路由说明

| 字段 | 内容 |
|------|------|
| 文档名称 | 前端页面路由说明 |
| 版本 | 1.0 |
| 创建日期 | _[YYYY-MM-DD]_ |
| 负责人 | _[负责人姓名]_ |
| 审核人 | _[审核人姓名]_ |

---

## 路由结构概述

本项目使用 **Vue Router 4.x** 进行路由管理。

- **路由文件位置：** `src/router/index.js`（或 `src/router/index.ts`）
- **路由模式：** `history` 模式（`createWebHistory`），生产环境需 Nginx 配置 `try_files` 支持
- **路由注册方式：** _[说明路由是静态配置还是动态注册，如：基础路由静态配置，业务路由根据用户权限动态注册]_
- **路由文件组织：** _[说明路由文件的拆分方式，如：按业务模块拆分为多个路由配置文件，在 index.js 中统一合并]_

```javascript
// src/router/index.js 基础结构示例
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/login/LoginView.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/',
      component: () => import('@/layouts/MainLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        // 业务路由在此动态注册或静态配置
      ],
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/error/NotFoundView.vue'),
    },
  ],
})

export default router
```

---

## 核心跳转方法

### router.push — 跳转并保留历史记录

```javascript
import { useRouter } from 'vue-router'

const router = useRouter()

// 通过路径跳转
router.push('/payment/list')

// 通过命名路由跳转
router.push({ name: 'payment-list' })

// 携带查询参数跳转（URL 显示为 /payment/list?status=pending&page=1）
router.push({ name: 'payment-list', query: { status: 'pending', page: 1 } })

// 携带动态路由参数跳转（URL 显示为 /payment/detail/123）
router.push({ name: 'payment-detail', params: { id: '123' } })
```

### router.replace — 跳转并替换当前历史记录

```javascript
// 适用于登录成功后跳转首页、表单提交成功后跳转列表页等场景
// 用户点击浏览器"后退"时不会回到被替换的页面

router.replace({ name: 'payment-list' })

// 等价于
router.push({ name: 'payment-list', replace: true })
```

### router.go — 在历史记录中前进或后退

```javascript
// 后退一步（等价于浏览器后退按钮）
router.go(-1)

// 前进一步
router.go(1)

// 后退两步
router.go(-2)
```

---

## 主要功能模块

_[列出系统中所有主要业务模块及其路由前缀，便于 AI 快速了解系统的功能划分。]_

| 模块名称 | 路由前缀 | 说明 |
|----------|----------|------|
| 工单管理 | `/workorder` | 报修工单的提交、派发、处理、评价全流程 |
| 缴费管理 | `/billing` | 账单查看、在线缴费、缴费记录 |
| 公告管理 | `/notice` | 社区公告发布、查看、已读管理 |
| 门禁管理 | `/access` | 访客预约、门禁记录、临时通行证 |
| 业主管理 | `/owner` | 业主信息、房产绑定、家庭成员 |
| 系统管理 | `/admin` | 用户、角色、权限、小区、楼栋等基础配置 |

---

## 路由命名规范

路由 `name` 属性采用 **模块名-功能名** 的格式，使用小写字母和连字符（kebab-case）：

```
{模块名}-{功能名}
```

**示例：**

| 路由 name | 对应路径 | 说明 |
|-----------|----------|------|
| `payment-list` | `/payment/list` | 付款申请列表页 |
| `payment-detail` | `/payment/detail/:id` | 付款申请详情页 |
| `payment-create` | `/payment/create` | 新建付款申请页 |
| `payment-approve` | `/payment/approve/:id` | 付款申请审批页 |
| `admin-user-list` | `/admin/users` | 用户管理列表页 |
| `admin-role-edit` | `/admin/roles/:id/edit` | 角色编辑页 |

**命名规则说明：**
- 模块名与路由前缀保持一致（去掉 `/`）
- 功能名使用动词或名词，如：`list`、`detail`、`create`、`edit`、`approve`
- 多级模块使用多个连字符连接，如：`admin-user-list`
- 避免使用缩写，保持可读性

---

## 常见跳转场景

### 场景一：列表页跳转详情页（携带 ID 参数）

```javascript
// 在列表页的行点击事件中
const handleRowClick = (row) => {
  router.push({ name: 'payment-detail', params: { id: row.id } })
}

// 在详情页中获取路由参数
import { useRoute } from 'vue-router'

const route = useRoute()
const paymentId = route.params.id // 获取动态路由参数
```

### 场景二：表单提交成功后跳转列表页（携带查询参数）

```javascript
const handleSubmitSuccess = () => {
  // 跳转到列表页并传递成功提示标记
  router.replace({ name: 'payment-list', query: { created: 'true' } })
}

// 在列表页中读取查询参数并展示提示
import { useRoute } from 'vue-router'
import { onMounted } from 'vue'
import { ElMessage } from 'element-plus'

const route = useRoute()

onMounted(() => {
  if (route.query.created === 'true') {
    ElMessage.success('付款申请提交成功')
  }
})
```

### 场景三：路由守卫（全局前置守卫）

```javascript
// src/router/index.js
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()

  // 不需要登录的页面直接放行
  if (to.meta.requiresAuth === false) {
    return next()
  }

  // 检查登录状态
  if (!authStore.isLoggedIn) {
    // 未登录，跳转登录页并记录目标路由
    return next({ name: 'login', query: { redirect: to.fullPath } })
  }

  // 检查页面权限
  if (to.meta.permission && !authStore.hasPermission(to.meta.permission)) {
    // 无权限，跳转 403 页面
    return next({ name: 'forbidden' })
  }

  next()
})
```

### 场景四：登录成功后跳转原目标页

```javascript
// 在登录页的登录成功回调中
const handleLoginSuccess = () => {
  const redirect = route.query.redirect
  if (redirect && typeof redirect === 'string') {
    router.replace(redirect)
  } else {
    router.replace({ name: 'home' })
  }
}
```

### 场景五：在非 setup 环境中使用路由（如：Pinia Store）

```javascript
// 在 Pinia Store 或工具函数中，不能使用 useRouter()，需直接导入 router 实例
import router from '@/router'

export const useAuthStore = defineStore('auth', {
  actions: {
    async logout() {
      // 清除登录状态...
      router.replace({ name: 'login' })
    },
  },
})
```

---

## 开发注意事项

### 路由懒加载

所有业务页面组件必须使用动态导入（懒加载），避免首屏加载过多资源：

```javascript
// ✅ 正确：使用动态导入
component: () => import('@/views/payment/PaymentListView.vue')

// ❌ 错误：直接导入会打包进主 bundle
import PaymentListView from '@/views/payment/PaymentListView.vue'
component: PaymentListView
```

### 权限控制

- **菜单级权限：** 根据后端返回的菜单权限数据，动态注册路由，未授权路由不注册，直接 404
- **按钮级权限：** 通过自定义指令 `v-permission` 或计算属性控制按钮的显示/隐藏
- **接口级权限：** 后端接口同样需要鉴权，前端权限控制仅为用户体验优化，不可替代后端鉴权

```javascript
// 自定义权限指令示例
// src/directives/permission.js
export const permission = {
  mounted(el, binding) {
    const authStore = useAuthStore()
    if (!authStore.hasPermission(binding.value)) {
      el.parentNode?.removeChild(el)
    }
  },
}

// 使用方式
// <el-button v-permission="'payment:approve'">审批</el-button>
```

### 404 处理

路由配置中必须包含通配符路由，捕获所有未匹配的路径：

```javascript
// 必须放在路由配置的最后
{
  path: '/:pathMatch(.*)*',
  name: 'not-found',
  component: () => import('@/views/error/NotFoundView.vue'),
}
```

### History 模式 Nginx 配置

使用 `createWebHistory` 时，生产环境 Nginx 需配置 `try_files`，否则刷新页面会出现 404：

```nginx
location / {
    try_files $uri $uri/ /index.html;
}
```

### PC 管理端侧边栏菜单分组规范

PC 管理端（spmp-web-pc）的侧边栏菜单按路由路径前缀自动分组为子菜单。每个业务域的路由必须使用统一的路径前缀，侧边栏组件（`AppSidebar.vue`）根据前缀自动归类。

**分组规则：**

| 路由前缀 | 菜单分组名称 | 图标 | 说明 |
|----------|-------------|------|------|
| `base/` | 基础数据 | OfficeBuilding | 片区、小区、楼栋、单元、房屋管理 |
| `owner/` | 业主管理 | UserFilled | 业主列表、认证审批 |
| `workorder/` | 工单管理 | Tickets | 报修工单管理（待开发） |
| `billing/` | 缴费管理 | Wallet | 账单缴费管理（待开发） |
| `notice/` | 公告管理 | Bell | 社区公告管理（待开发） |
| `access/` | 门禁管理 | Lock | 门禁访客管理（待开发） |
| `system/` | 系统管理 | Setting | 用户、角色、菜单管理 |
| `log/` | 日志管理 | Document | 登录日志、操作日志 |
| 其他 | 独立菜单项 | — | 首页等不归属任何分组的路由 |

**新增业务域时的前端要求：**

1. 路由路径必须以域前缀开头（如 `workorder/orders`、`billing/bills`）
2. 在 `AppSidebar.vue` 的分组逻辑中添加对应前缀的过滤和分组
3. 在 `iconMap` 中导入并注册该域使用的图标
4. 详情页等不需要在菜单中显示的路由，设置 `meta.hidden: true`
5. 路由 `meta.permission` 与后端 `@PreAuthorize` 权限标识保持一致

**实现参考（AppSidebar.vue 分组逻辑）：**

```typescript
// 按路由前缀分组
const baseItems = children.filter(r => r.path?.startsWith('base/'))
const ownerItems = children.filter(r => r.path?.startsWith('owner/'))
const workorderItems = children.filter(r => r.path?.startsWith('workorder/'))
// ... 其他域类似

// 每个分组渲染为 el-sub-menu
if (ownerItems.length) {
  groups.push({
    type: 'submenu',
    title: '业主管理',
    icon: 'UserFilled',
    index: 'owner',
    children: ownerItems
  })
}
```
