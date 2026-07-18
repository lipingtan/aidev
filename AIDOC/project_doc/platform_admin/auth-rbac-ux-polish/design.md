# 设计：auth-rbac UI/UX 交互优化

## 技术方案

### 新增依赖

| 包名 | 版本 | 用途 | 体积 |
|------|------|------|------|
| jsondiffpatch | ^0.6 | JSON diff 高亮 | ~30KB gzip |

### 核心实现方案

#### 1. 全局 Loading 服务 (FR-13/14)

```typescript
// src/utils/loading.ts
import { ElLoading } from 'element-plus'

let loadingInstance: ReturnType<typeof ElLoading.service> | null = null
let count = 0

export function showLoading(text = '加载中...') {
  count++
  if (!loadingInstance) {
    loadingInstance = ElLoading.service({ fullscreen: true, text, background: 'rgba(0,0,0,0.3)' })
  }
}

export function hideLoading() {
  count--
  if (count <= 0) {
    count = 0
    loadingInstance?.close()
    loadingInstance = null
  }
}
```

集成点：
- `request.ts` 响应拦截器 token 刷新时调用 `showLoading` / `hideLoading`
- `router/guard.ts` 路由切换时调用（beforeEach show / afterEach hide）

#### 2. 网络异常兜底 (FR-15)

```
src/views/error/NetworkError.vue  — 网络异常页面（图标 + "网络异常" + 重试按钮）
src/utils/request.ts              — 响应拦截器捕获 network error 时 router.push('/network-error')
src/router/static-routes.ts       — 新增 /network-error 路由
```

#### 3. 操作列下拉菜单 (FR-1)

UserList.vue 操作列改为：

```html
<el-table-column label="操作" width="200" fixed="right">
  <template #default="{ row }">
    <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
    <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
    <el-dropdown trigger="click">
      <el-button link type="primary">更多<el-icon><ArrowDown /></el-icon></el-button>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item @click="handleTenantManage(row)">租户关联</el-dropdown-item>
          <el-dropdown-item @click="handleRoleAssign(row)">角色分配</el-dropdown-item>
          <el-dropdown-item @click="handleForceOffline(row)">强制下线</el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </template>
</el-table-column>
```

#### 4. Tab 未保存检测 (FR-2)

角色权限 4 个 Tab 组件统一暴露 `isDirty` 状态：

```typescript
// 每个 Tab 组件 defineExpose({ isDirty })
// RoleList.vue 切换 Tab 前检查：
function handleTabChange(newTab: string) {
  const currentTabRef = getCurrentTabRef()
  if (currentTabRef?.isDirty) {
    ElMessageBox.confirm('有未保存的更改，是否放弃？', '提示', { type: 'warning' })
      .then(() => { activeTab.value = newTab })
      .catch(() => {})
  } else {
    activeTab.value = newTab
  }
}
```

#### 5. 资源树动态加载 appList (FR-3)

```typescript
// ResourceTree.vue onMounted
import { listApplications } from '@/api/application'

const appList = ref<string[]>([])
onMounted(async () => {
  const apps = await listApplications()
  appList.value = apps.map(a => a.code)
  if (appList.value.length) {
    appCode.value = appList.value[0]
    fetchTree()
  }
})
```

#### 6. JSON Diff 高亮 (FR-16)

```typescript
// src/utils/json-diff.ts
import * as jsondiffpatch from 'jsondiffpatch'
import 'jsondiffpatch/formatters/styles/html.css'

const diffInstance = jsondiffpatch.create()

export function formatDiffHtml(oldVal: any, newVal: any): string {
  const delta = diffInstance.diff(oldVal, newVal)
  if (!delta) return '<span class="no-change">无变更</span>'
  return jsondiffpatch.formatters.html.format(delta, oldVal)
}
```

#### 7. 虚拟滚动 (FR-11)

接口权限右侧未分配列表：当数据 > 100 条时切换为 el-tree-v2（不支持拖拽时 fallback 为手动分页 + el-tree）。

实际方案：保留 el-tree，右侧增加搜索框 + 分页（每页 50 条），避免全量渲染。

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 现有 CRUD 功能不受影响 | 各列表页增删改查正常 |
| RG-2 | 拖拽排序功能不受影响 | 资源树拖拽后顺序正确保存 |
| RG-3 | Token 刷新机制不受影响 | 401 后自动刷新并重试 |
| RG-4 | 路由守卫认证逻辑不受影响 | 未登录跳转 /login |
| RG-5 | 暗色主题已有样式不退化 | 切换 luxury-dark 后原有组件正常 |
