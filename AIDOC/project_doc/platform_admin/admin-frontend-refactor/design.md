# 技术设计文档：admin-frontend-refactor

## 概述

将 `frontend`（pure-admin-thin）中已实现的系统管理模块迁移到 `dev-web-admin` 框架中。迁移范围包括：部门管理、岗位管理、字典管理、系统配置、接口管理、权限演示、服务监控、定时任务、插件容器、开发工具共 10 个功能模块，以及配套的路由注册、API 层封装、旧业务模块清理和样式规范统一。

设计核心原则：
- 复用 dev-web-admin 现有的页面模式（搜索卡片 + 表格卡片 + 表单对话框）
- API 层适配 go-admin 的 PageOK 响应格式
- 所有新页面遵循已有 UserList.vue 确立的代码组织和样式模式

## 架构

### 整体架构不变

```
┌─────────────────────────────────────────────┐
│              AppLayout（保留）                │
│  ┌─────────┐ ┌────────────────────────────┐ │
│  │AppSidebar│ │  路由视图（router-view）    │ │
│  │  (菜单)  │ │  ┌──────────────────────┐  │ │
│  │          │ │  │ 各模块页面组件        │  │ │
│  │          │ │  └──────────────────────┘  │ │
│  └─────────┘ └────────────────────────────┘ │
└─────────────────────────────────────────────┘
```

### 新增模块在架构中的位置

```
src/
├── api/
│   ├── dept.ts          # 新增：部门管理 API
│   ├── post.ts          # 新增：岗位管理 API
│   ├── dict.ts          # 新增：字典管理 API
│   ├── config.ts        # 新增：系统配置 API
│   ├── sys-api.ts       # 新增：接口管理 API
│   ├── monitor.ts       # 新增：服务监控 API
│   ├── job.ts           # 新增：定时任务 API
│   └── (已有文件保留)
├── views/
│   ├── system/
│   │   ├── dept/        # 新增：部门管理页面
│   │   ├── post/        # 新增：岗位管理页面
│   │   ├── dict/        # 新增：字典管理页面
│   │   ├── config/      # 新增：系统配置页面
│   │   ├── api/         # 新增：接口管理页面
│   │   └── (user/role/menu/log 保留)
│   ├── monitor/
│   │   └── server/      # 新增：服务监控页面
│   ├── job/
│   │   ├── list/        # 新增：任务列表页面
│   │   └── log/         # 新增：任务日志页面
│   ├── permission/
│   │   ├── button/      # 新增：按钮权限演示
│   │   └── page/        # 新增：页面权限演示
│   ├── plugin/
│   │   └── PluginContainer.vue  # 新增：插件容器
│   ├── devtools/
│   │   ├── codegen/     # 新增：代码生成占位
│   │   ├── build/       # 新增：构建工具占位
│   │   └── swagger/     # 新增：Swagger 页面
│   └── (删除: base/ owner/ workorder/ billing/ notice/)
└── router/
    └── static-routes.ts # 修改：移除旧路由、注册新路由
```

## 组件与接口

### 页面组件设计模式

所有 CRUD 模块页面统一采用已有模式：

```
ModuleList.vue        # 主页面：搜索栏 + 表格 + 分页
ModuleForm.vue        # 表单对话框组件：新增/编辑复用
```

**主页面结构**（与 UserList.vue 一致）：
- `<el-card class="search-card">` — 搜索表单
- `<el-card class="table-card">` — 操作按钮 + 表格 + 分页
- `<ModuleForm ref="formRef" @success="fetchData" />` — 表单弹窗

**表单对话框模式**：
- 通过 `ref` 暴露 `open(row?)` 方法
- 无参数 = 新增模式，传参 = 编辑模式
- 提交成功后 emit `success` 事件刷新列表

### 特殊模块组件

| 模块 | 组件 | 特殊点 |
|------|------|--------|
| 部门管理 | DeptList.vue + DeptForm.vue | 树形表格（el-table row-key + tree-props），非分页 |
| 字典管理 | DictTypeList.vue + DictTypeForm.vue + DictDataList.vue + DictDataForm.vue | 两级列表（类型→数据），通过路由参数或对话框切换 |
| 服务监控 | ServerMonitor.vue | 纯展示卡片，无 CRUD |
| 权限演示 | ButtonPermission.vue + PagePermission.vue | 演示页，无后端交互 |
| 插件容器 | PluginContainer.vue | 动态加载外部 JS 模块 |
| 开发工具 | CodegenPlaceholder.vue + BuildPlaceholder.vue + SwaggerView.vue | 占位页 + iframe |

### API 层接口设计

统一适配模式（与 frontend 的 system.ts 中 `adaptPage` 等价）：

```typescript
// src/api/helpers.ts — 新增公共适配函数
import request from '@/utils/request'

export interface PageQuery {
  pageNum: number
  pageSize: number
  [key: string]: any
}

/**
 * 适配 go-admin PageOK 响应格式
 * 后端返回: { code: 200, data: { list: T[], count: number }, message: '' }
 * 适配为: { list: T[], total: number }
 */
export function adaptPageResponse<T>(res: any): { list: T[]; total: number } {
  if (res && Array.isArray(res.list)) {
    return { list: res.list, total: res.count || res.list.length }
  }
  if (Array.isArray(res)) {
    return { list: res, total: res.length }
  }
  return { list: [], total: 0 }
}
```

各模块 API 文件结构示例（以 dept.ts 为例）：

```typescript
// src/api/dept.ts
import request from '@/utils/request'

export interface DeptItem {
  deptId: number
  deptName: string
  parentId: number
  sort: number
  leader: string
  status: number
  children?: DeptItem[]
  createTime: string
}

export interface DeptCreateParams {
  deptName: string
  parentId: number
  sort: number
  leader?: string
  status?: number
}

export function listDepts(params?: { deptName?: string }) {
  return request.get('/api/v1/dept', { params })
}

export function getDept(id: number) {
  return request.get(`/api/v1/dept/${id}`)
}

export function createDept(data: DeptCreateParams) {
  return request.post('/api/v1/dept', data)
}

export function updateDept(id: number, data: Partial<DeptCreateParams>) {
  return request.put(`/api/v1/dept/${id}`, data)
}

export function deleteDept(id: number) {
  return request.delete(`/api/v1/dept/${id}`)
}

export function getDeptTree() {
  return request.get('/api/v1/deptTree')
}
```

### 插件容器核心逻辑

```typescript
// PluginContainer.vue 核心流程
async function loadPlugin(pluginName: string) {
  loading.value = true
  error.value = ''
  try {
    const module = await import(/* @vite-ignore */ `/static/plugins/${pluginName}/index.js`)
    if (module.routes) {
      // 根据当前路径匹配插件内部路由，渲染对应组件
      const matched = module.routes.find(r => currentPath.includes(r.path))
      if (matched) {
        pluginComponent.value = matched.component
      } else {
        error.value = '插件页面加载失败或不存在'
      }
    }
  } catch (e) {
    error.value = '插件页面加载失败或不存在'
  } finally {
    loading.value = false
  }
}
```

## 数据模型

### API 请求/响应格式（go-admin 标准）

**分页请求参数**：
```typescript
interface PageQuery {
  pageNum: number    // 页码，从 1 开始
  pageSize: number   // 每页条数
  [key: string]: any // 筛选条件
}
```

**分页响应（PageOK 格式）**：
```typescript
interface GoAdminPageResponse<T> {
  code: number       // 200 = 成功
  message: string
  data: {
    list: T[]        // 数据列表
    count: number    // 总记录数
  }
}
```

**单条响应（OK 格式）**：
```typescript
interface GoAdminResponse<T> {
  code: number
  message: string
  data: T
}
```

### 各模块数据模型

| 模块 | 主要字段 | API 路径 |
|------|----------|----------|
| 部门 | deptId, deptName, parentId, sort, leader, status | /api/v1/dept |
| 岗位 | postId, postCode, postName, sort, status | /api/v1/post |
| 字典类型 | dictId, dictName, dictType, status, remark | /api/v1/dict/type |
| 字典数据 | dictCode, dictSort, dictLabel, dictValue, dictType, status | /api/v1/dict/data |
| 系统配置 | configId, configName, configKey, configValue, configType, remark | /api/v1/config |
| 接口 | id, path, method, title, group | /api/v1/sys-api |
| 服务监控 | cpu{}, mem{}, disk{}, os{} | /api/v1/server-monitor |
| 定时任务 | jobId, jobName, jobGroup, cronExpression, invokeTarget, status | /api/v1/sysjob |
| 任务日志 | logId, jobName, jobGroup, invokeTarget, status, createTime | /api/v1/sysjob/log |

### 路由配置数据模型

```typescript
// 新增路由统一使用 AppLayout 作为父容器
{
  path: '/system/dept',
  name: 'system-dept',
  component: () => import('@/views/system/dept/DeptList.vue'),
  meta: { title: '部门管理', icon: 'OfficeBuilding', permission: 'system:dept:list' }
}
```

## 错误处理

### API 层错误处理

- 所有 API 调用通过已有的 `request.ts` 响应拦截器统一处理
- 业务状态码 ≠ 200 时，自动弹出 ElMessage.error
- 401 状态自动清除 Token 并跳转登录页
- 页面组件中 try/catch 仅处理特殊逻辑（如刷新列表）

### 模块级错误处理

| 场景 | 处理方式 |
|------|----------|
| 列表加载失败 | loading 状态结束，表格显示空数据，拦截器已提示错误 |
| 表单提交失败 | 对话框保持打开，拦截器已提示错误信息 |
| 删除失败 | 确认框关闭，拦截器已提示错误 |
| 服务监控接口失败 | 卡片展示空数据/默认值，不崩溃 |
| 插件加载失败 | 展示 "插件页面加载失败或不存在" 空状态 |
| 路由无权限 | 侧边栏不显示该菜单项 |

### 业务模块清理后的错误防护

- 清理完成后执行 `vue-tsc` 编译检查，确保无遗留 import 引用
- 路由移除后，已保存的标签页（TagsView）中可能存在失效路由，需在路由守卫中过滤无效标签

## 测试策略

### 为什么不使用属性基测试

本项目属于前端 UI 重构/迁移，涉及的主要内容为：
- Vue 组件渲染和布局（UI rendering）
- 简单的 CRUD 操作（list/create/update/delete）
- API 层为薄封装，无复杂数据转换逻辑
- 插件容器为副作用操作（动态加载 JS）

这些场景不适合属性基测试，应使用示例基测试和集成测试。

### 测试方案

**1. 单元测试（vitest）**

| 测试对象 | 测试内容 |
|----------|----------|
| adaptPageResponse | PageOK 格式解析、数组格式兜底、空数据处理 |
| API 函数 | 使用 MSW 或手动 mock axios，验证请求路径/方法/参数正确 |
| 权限指令 v-auth | 有权限时渲染、无权限时隐藏 |
| hasAuth 函数 | 权限列表包含/不包含指定 code 的判断 |

**2. 组件测试（vitest + @vue/test-utils）**

| 测试对象 | 测试内容 |
|----------|----------|
| DeptList | 树形表格渲染、搜索过滤、新增/编辑/删除按钮交互 |
| PluginContainer | 加载中状态、加载失败显示空状态、成功渲染组件 |
| ServerMonitor | 数据正常展示卡片、接口失败展示空数据不崩溃 |

**3. 编译验证**

- `vue-tsc` 类型检查通过
- `vite build` 构建无报错
- 清理旧模块后无遗留引用

**4. 手动验收测试**

- 各模块 CRUD 全流程可用
- 侧边栏菜单导航正确
- 权限控制生效
- 插件容器正常加载外部插件
- 旧模块路由不可访问（404）
