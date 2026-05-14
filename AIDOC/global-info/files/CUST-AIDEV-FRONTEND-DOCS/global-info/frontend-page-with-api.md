# 前端页面与 API 对应关系

| 字段 | 内容 |
|------|------|
| 文档名称 | 前端页面与 API 对应关系 |
| 版本 | 1.0 |
| 创建日期 | _[YYYY-MM-DD]_ |
| 负责人 | _[负责人姓名]_ |
| 审核人 | _[审核人姓名]_ |

---

## 页面列表

_[列出系统中所有主要页面及其对应调用的后端 API，便于 AI 快速定位页面与接口的关联关系。]_

| 页面名称 | 路由路径 | 使用的 API | 说明 |
|----------|----------|-----------|------|
| 付款申请列表 | `/payment/list` | `GET /api/v1/payments` | 分页查询付款申请，支持多条件筛选 |
| 付款申请详情 | `/payment/detail/:id` | `GET /api/v1/payments/:id` | 查询单条付款申请的完整信息 |
| _[页面名称，如：新建付款申请]_ | _[路由路径，如：/payment/create]_ | _[API，如：POST /api/v1/payments]_ | _[说明，如：提交新的付款申请]_ |
| _[页面名称，如：付款申请审批]_ | _[路由路径，如：/payment/approve/:id]_ | _[API，如：POST /api/v1/payments/:id/approve]_ | _[说明，如：审批人对付款申请进行审批操作]_ |
| _[页面名称，如：用户管理]_ | _[路由路径，如：/admin/users]_ | _[API，如：GET /api/v1/admin/users]_ | _[说明，如：管理员查询和管理系统用户]_ |
| _[页面名称，如：角色权限配置]_ | _[路由路径，如：/admin/roles]_ | _[API，如：GET /api/v1/admin/roles]_ | _[说明，如：配置角色与菜单/按钮权限的映射关系]_ |
| _[页面名称]_ | _[路由路径]_ | _[API 路径]_ | _[说明]_ |

> **说明：** 部分页面可能调用多个 API（如：详情页同时加载主体信息和关联数据），请在说明列中注明所有相关接口，或拆分为多行记录。

---

## 公共 API

_[描述所有页面或大多数页面共用的基础 API，这些接口通常在应用初始化或布局组件中调用，无需在每个页面重复声明。]_

### 获取当前用户信息

- **接口路径：** `GET /api/v1/auth/me`
- **调用时机：** 应用启动时（`App.vue` 或路由守卫中）
- **说明：** 获取当前登录用户的基本信息（用户名、角色、权限列表），结果存入 Pinia Store，供全局使用。

### 获取数据字典

- **接口路径：** `GET /api/v1/dict/items?dictCode={dictCode}`
- **调用时机：** 需要下拉选项的页面初始化时，或统一在应用启动时预加载常用字典
- **说明：** 获取指定字典编码的选项列表（如：付款类型、审批状态、币种等），建议结合本地缓存（sessionStorage）减少重复请求。

### 获取菜单权限

- **接口路径：** `GET /api/v1/auth/menus`
- **调用时机：** 用户登录成功后，动态注册路由前
- **说明：** 根据当前用户角色返回可访问的菜单树，前端据此动态生成侧边栏导航和路由配置。

### 文件上传

- **接口路径：** `POST /api/v1/files/upload`
- **调用时机：** 任何包含附件上传功能的页面
- **说明：** 统一的文件上传接口，返回文件 ID 和访问 URL，支持图片、PDF、Excel 等格式，单文件大小限制 _[如：20MB]_。

### _[其他公共 API 名称]_

- **接口路径：** _[如：GET /api/v1/organizations/tree]_
- **调用时机：** _[如：需要选择组织机构的页面]_
- **说明：** _[如：获取组织机构树形数据，用于部门选择器组件]_

---

## API 版本说明

_[说明项目中 API 版本的管理规则，帮助 AI 在生成代码时使用正确的接口路径前缀。]_

### 版本前缀规则

本项目所有后端 API 均以版本号作为路径前缀，当前主版本为 **v1**：

```
/api/v1/{模块路径}/{资源路径}
```

示例：
- `GET /api/v1/payments` — 查询付款申请列表
- `POST /api/v1/payments` — 创建付款申请
- `GET /api/v1/payments/{id}` — 查询单条付款申请
- `PUT /api/v1/payments/{id}` — 更新付款申请
- `DELETE /api/v1/payments/{id}` — 删除付款申请

### 版本升级策略

- 当接口发生**不兼容变更**时，升级主版本号（如：`/api/v2/`），旧版本接口保留一段时间后下线
- 当接口发生**向后兼容的变更**（如：新增可选参数）时，不升级版本号，直接在原接口上扩展
- _[如有特殊版本规则，在此补充说明]_

### Axios 基础配置

前端通过统一的 axios 实例发起请求，基础 URL 和版本前缀在环境变量中配置：

```javascript
// src/utils/request.js
import axios from 'axios'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL, // 如：https://api.example.com/api/v1
  timeout: 10000,
})

// 请求拦截器：统一添加 Authorization 头
request.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器：统一处理错误码
request.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      // Token 过期，跳转登录页
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default request
```

### 环境变量配置

```
# .env.development
VITE_API_BASE_URL=http://localhost:8080/api/v1

# .env.production
VITE_API_BASE_URL=https://api.example.com/api/v1
```

> **注意：** _[如有特殊的接口鉴权方式（如：行内统一网关 Token、证书认证等），在此补充说明具体配置方法。]_
