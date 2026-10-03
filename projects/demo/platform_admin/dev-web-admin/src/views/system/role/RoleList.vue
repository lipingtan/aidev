<template>
  <div class="role-page" :class="{ 'mobile-layout': isMobile }">
    <el-card shadow="never" class="role-tree-card main-card">
      <!-- 固定头部区域（手机端） -->
      <div class="sticky-header">
        <div class="card-header">
          <span>角色管理</span>
          <el-button type="primary" size="small" @click="handleAdd">新增</el-button>
        </div>

        <!-- Tab 切换：角色 / 权限集 -->
        <el-tabs v-model="activeTab" @tab-change="handleTabChange">
          <el-tab-pane label="角色" name="role" />
          <el-tab-pane label="权限集" name="permission_set" />
        </el-tabs>

        <el-input v-model="filterText" placeholder="搜索角色" clearable style="margin-bottom: 12px" />
      </div>

      <!-- 可滚动的列表区域 -->
      <div class="scrollable-content">
        <el-tree
          ref="roleTreeRef"
          v-loading="treeLoading"
          :data="roleTree"
          :props="{ label: 'role_name', children: 'children' }"
          node-key="id"
          highlight-current
          :expand-on-click-node="false"
          :filter-node-method="filterNode"
          default-expand-all
          @node-click="handleNodeClick"
        >
          <template #default="{ node, data }">
            <div class="tree-node">
              <el-tooltip :content="node.label" placement="top" :disabled="node.label.length < 12">
                <span class="tree-node-label">{{ node.label }}</span>
              </el-tooltip>
              <span class="tree-node-actions">
                <el-button link type="primary" size="small" @click.stop="handleEdit(data)">编辑</el-button>
                <el-button link type="danger" size="small" @click.stop="handleDelete(data)">删除</el-button>
              </span>
            </div>
          </template>
        </el-tree>
      </div>
    </el-card>

    <!-- 右侧：权限概览 -->
    <el-card shadow="never" class="role-detail-card">
      <template v-if="currentRole">
        <div class="role-detail-header">
          <span class="role-name">{{ currentRole.role_name }}</span>
          <el-tag size="small" :type="currentRole.status === 1 ? 'success' : 'danger'">
            {{ currentRole.status === 1 ? '启用' : '禁用' }}
          </el-tag>
          <el-tag v-if="currentRole.role_type === 'PERMISSION_SET'" size="small" type="warning">权限集</el-tag>
          <span class="role-code">{{ currentRole.role_code }}</span>
        </div>

        <!-- 应用权限汇总 - 响应式列表 -->
        <ResponsiveList :data="permissionSummary" :loading="summaryLoading" :row-key="'app_code'" :table-props="{ size: 'small' }">
          <!-- 桌面端表格 -->
          <template #table>
            <el-table-column prop="app_name" label="应用" min-width="120">
              <template #default="{ row }">
                <span>{{ row.app_name }}</span>
                <span v-if="isPluginStopped(row.app_code)" class="plugin-stopped-hint">(插件已停止)</span>
                <el-tag v-if="row.bound" type="success" size="small" style="margin-left: 8px">已绑定</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="菜单权限" width="100" align="center">
              <template #default="{ row }">
                <span :class="{ 'count-zero': row.menu_count === 0 }">{{ row.menu_count }} 项</span>
              </template>
            </el-table-column>
            <el-table-column label="接口权限" width="100" align="center">
              <template #default="{ row }">
                <span :class="{ 'count-zero': row.api_count === 0 }">{{ row.api_count }} 项</span>
              </template>
            </el-table-column>
            <el-table-column label="数据权限" width="100" align="center">
              <template #default="{ row }">
                <span :class="{ 'count-zero': row.data_scope_count === 0 }">{{ row.data_scope_count }} 维度</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="80" align="center" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link @click="openPermDrawer(row)">配置</el-button>
              </template>
            </el-table-column>
          </template>

          <!-- 手机端卡片 -->
          <template #card="{ row }">
            <div class="card-title">
              {{ row.app_name }}
              <el-tag v-if="row.bound" type="success" size="small" class="card-tag">已绑定</el-tag>
              <el-tag v-if="isPluginStopped(row.app_code)" type="danger" size="small" class="card-tag">插件已停止</el-tag>
            </div>
            <div class="card-row">
              <span class="card-label">菜单权限</span>
              <span class="card-value" :class="{ 'count-zero': row.menu_count === 0 }">{{ row.menu_count }} 项</span>
            </div>
            <div class="card-row">
              <span class="card-label">接口权限</span>
              <span class="card-value" :class="{ 'count-zero': row.api_count === 0 }">{{ row.api_count }} 项</span>
            </div>
            <div class="card-row">
              <span class="card-label">数据权限</span>
              <span class="card-value" :class="{ 'count-zero': row.data_scope_count === 0 }">{{ row.data_scope_count }} 维度</span>
            </div>
          </template>

          <!-- 手机端操作按钮 -->
          <template #actions="{ row }">
            <el-button size="small" type="primary" @click="openPermDrawer(row)">配置</el-button>
          </template>
        </ResponsiveList>
      </template>
      <template v-else>
        <el-empty description="请从左侧选择一个角色" />
      </template>
    </el-card>

    <!-- 权限配置 Drawer -->
    <el-drawer v-model="drawerVisible" :title="drawerTitle" :size="isMobile ? '95%' : '60%'" destroy-on-close>
      <el-tabs v-model="drawerTab">
        <el-tab-pane label="菜单权限" name="menu">
          <MenuPermTab :role-id="currentRole!.id" :app-code="drawerAppCode" :parent-id="currentRole!.parent_id" />
        </el-tab-pane>
        <el-tab-pane label="接口权限" name="api">
          <ApiPermTab :role-id="currentRole!.id" :app-code="drawerAppCode" :parent-id="currentRole!.parent_id" />
        </el-tab-pane>
        <el-tab-pane label="数据权限" name="data">
          <DataScopeTab :role-id="currentRole!.id" />
        </el-tab-pane>
        <el-tab-pane label="字段权限" name="field">
          <FieldPermTab :role-id="currentRole!.id" />
        </el-tab-pane>
      </el-tabs>
    </el-drawer>

    <!-- 角色新增/编辑弹窗 -->
    <RoleForm ref="formRef" :role-tree="roleTree" :is-permission-set="isPermissionSetTab" @success="loadRoleTree" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { ElTree } from 'element-plus'
import { getRoleList, deleteRole, getRolePermissionSummary } from '@/api/role'
import type { RoleItem, AppPermissionSummary } from '@/api/role'
import { getPluginList } from '@/api/plugin'
import type { PluginItem } from '@/api/plugin'
import RoleForm from './RoleForm.vue'
import MenuPermTab from './MenuPermTab.vue'
import ApiPermTab from './ApiPermTab.vue'
import DataScopeTab from './DataScopeTab.vue'
import FieldPermTab from './FieldPermTab.vue'
import ResponsiveList from '@/components/ResponsiveList.vue'

// 响应式检测
const isMobile = ref(false)
const MOBILE_BREAKPOINT = 768
function checkMobile() {
  isMobile.value = window.innerWidth <= MOBILE_BREAKPOINT
}
onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})
onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})

const treeLoading = ref(false)
const roleTree = ref<RoleItem[]>([])
const currentRole = ref<RoleItem | null>(null)
const filterText = ref('')
const roleTreeRef = ref<InstanceType<typeof ElTree>>()
const formRef = ref()

// Tab 切换
const activeTab = ref<'role' | 'permission_set'>('role')
const isPermissionSetTab = computed(() => activeTab.value === 'permission_set')

// 权限汇总
const permissionSummary = ref<AppPermissionSummary[]>([])
const summaryLoading = ref(false)

// Drawer 状态
const drawerVisible = ref(false)
const drawerAppCode = ref('')
const drawerTitle = ref('')
const drawerTab = ref('menu')

// 插件状态 map：name → runStatus
const pluginStatusMap = ref<Record<string, string>>({})

/** 加载插件状态列表，构建 name → runStatus 的映射 */
async function loadPluginStatus() {
  try {
    const res: any = await getPluginList()
    const data = res?.data || res
    const list: PluginItem[] = data?.list || []
    const map: Record<string, string> = {}
    list.forEach((p) => {
      map[p.name] = p.runStatus
    })
    pluginStatusMap.value = map
  } catch {
    pluginStatusMap.value = {}
  }
}

/** 判断某 app_code 对应的插件是否已停止 */
function isPluginStopped(appCode: string): boolean {
  const status = pluginStatusMap.value[appCode]
  return !!status && status !== 'running'
}

/** 加载角色树 */
async function loadRoleTree() {
  treeLoading.value = true
  try {
    if (activeTab.value === 'permission_set') {
      roleTree.value = await getRoleList('PERMISSION_SET')
    } else {
      // 角色 Tab：获取全部后过滤掉 PERMISSION_SET
      const all = await getRoleList()
      roleTree.value = filterOutPermissionSet(all)
    }
  } finally {
    treeLoading.value = false
  }
}

/** 递归过滤掉 PERMISSION_SET 类型角色 */
function filterOutPermissionSet(nodes: RoleItem[]): RoleItem[] {
  return nodes
    .filter(n => n.role_type !== 'PERMISSION_SET')
    .map(n => ({
      ...n,
      children: n.children ? filterOutPermissionSet(n.children) : undefined
    }))
}

/** Tab 切换处理 */
function handleTabChange() {
  currentRole.value = null
  permissionSummary.value = []
  loadRoleTree()
}

/** 加载权限汇总 */
async function loadPermissionSummary() {
  if (!currentRole.value) return
  summaryLoading.value = true
  try {
    permissionSummary.value = await getRolePermissionSummary(currentRole.value.id)
  } finally {
    summaryLoading.value = false
  }
}

/** 搜索过滤 */
function filterNode(value: string, data: RoleItem) {
  if (!value) return true
  return data.role_name.includes(value) || data.role_code.includes(value)
}

watch(filterText, (val) => {
  roleTreeRef.value?.filter(val)
})

/** 选中角色节点 */
function handleNodeClick(data: RoleItem) {
  currentRole.value = data
  loadPermissionSummary()
}

/** 打开权限配置 Drawer */
function openPermDrawer(row: AppPermissionSummary) {
  drawerAppCode.value = row.app_code
  drawerTitle.value = `${row.app_name} - 权限配置`
  drawerTab.value = 'menu'
  drawerVisible.value = true
}

/** 新增角色 */
function handleAdd() {
  formRef.value?.open()
}

/** 编辑角色 */
function handleEdit(data: RoleItem) {
  formRef.value?.open(data)
}

/** 删除角色 */
async function handleDelete(data: RoleItem) {
  try {
    await ElMessageBox.confirm(`确认删除角色「${data.role_name}」？`, '提示', { type: 'warning' })
    await deleteRole(data.id)
    ElMessage.success('删除成功')
    if (currentRole.value?.id === data.id) {
      currentRole.value = null
    }
    loadRoleTree()
  } catch (e: any) {
    const code = e?.response?.data?.code
    if (code === 40006) {
      ElMessage.warning('该角色已绑定用户，无法删除')
    }
  }
}

onMounted(() => {
  loadRoleTree()
  loadPluginStatus()
})
</script>

<style scoped>
.role-page {
  display: flex;
  gap: 16px;
  padding: 16px;
  height: calc(100vh - 100px);
}
.role-tree-card {
  width: 320px;
  flex-shrink: 0;
  overflow-y: auto;
}
.role-detail-card {
  flex: 1;
  overflow-y: auto;
}

.search-form {
  margin-bottom: 0;
}

/* 窄屏（< 768px）：固定头部 + 可滚动内容 */
@media (max-width: 767px) {
  .role-page {
    flex-direction: column;
    gap: 0;
    padding: 0;
    height: 100vh;
    overflow: hidden;
  }

  .role-tree-card.main-card {
    width: 100%;
    display: flex;
    flex-direction: column;
    border-radius: 0;
    overflow: hidden;
    flex: 1;
    min-height: 0;
  }

  .role-tree-card.main-card :deep(.el-card__body) {
    flex: 1;
    display: flex;
    flex-direction: column;
    padding: 0;
    overflow: hidden;
    min-height: 0;
  }

  .sticky-header {
    flex-shrink: 0;
    background: var(--el-bg-color);
    z-index: 10;
    padding: 0 12px;
    border-bottom: 1px solid var(--el-border-color-lighter);
  }

  .sticky-header .card-header {
    padding: 12px 0;
    border-bottom: 1px solid var(--el-border-color-light);
  }

  .scrollable-content {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
    -webkit-overflow-scrolling: touch;
    min-height: 0;
  }

  .role-detail-card {
    flex-shrink: 0;
    overflow-y: auto;
    max-height: 40vh;
    border-top: 8px solid var(--el-fill-color-light);
  }

  /* 移动端树节点操作按钮始终显示 */
  .tree-node-actions {
    opacity: 1 !important;
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.tree-node {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding-right: 8px;
}
.tree-node-label {
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: inline-block;
}
.tree-node-actions {
  opacity: 0;
  transition: opacity 0.2s;
}
.tree-node:hover .tree-node-actions {
  opacity: 1;
}
.role-detail-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.role-name {
  font-size: 16px;
  font-weight: 600;
}
.role-code {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.count-zero {
  color: var(--el-text-color-secondary);
}
.plugin-stopped-hint {
  color: #f56c6c;
  font-size: 12px;
  margin-left: 4px;
}
</style>
