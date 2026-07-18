<template>
  <div class="role-page">
    <!-- 左侧：角色树 -->
    <el-card shadow="never" class="role-tree-card">
      <template #header>
        <div class="card-header">
          <span>角色列表</span>
          <el-button type="primary" size="small" @click="handleAdd">新增</el-button>
        </div>
      </template>
      <el-input v-model="filterText" placeholder="搜索角色" clearable style="margin-bottom: 12px" />
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
    </el-card>

    <!-- 右侧：权限概览 -->
    <el-card shadow="never" class="role-detail-card">
      <template v-if="currentRole">
        <div class="role-detail-header">
          <span class="role-name">{{ currentRole.role_name }}</span>
          <el-tag size="small" :type="currentRole.status === 1 ? 'success' : 'danger'">
            {{ currentRole.status === 1 ? '启用' : '禁用' }}
          </el-tag>
          <span class="role-code">{{ currentRole.role_code }}</span>
        </div>

        <!-- 应用权限汇总表格 -->
        <el-table v-loading="summaryLoading" :data="permissionSummary" row-key="app_code" size="small">
          <el-table-column prop="app_name" label="应用" min-width="120">
            <template #default="{ row }">
              <span>{{ row.app_name }}</span>
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
          <el-table-column label="操作" width="80" align="center">
            <template #default="{ row }">
              <el-button type="primary" link @click="openPermDrawer(row)">配置</el-button>
            </template>
          </el-table-column>
        </el-table>
      </template>
      <template v-else>
        <el-empty description="请从左侧选择一个角色" />
      </template>
    </el-card>

    <!-- 权限配置 Drawer -->
    <el-drawer v-model="drawerVisible" :title="drawerTitle" size="60%" destroy-on-close>
      <el-tabs v-model="drawerTab">
        <el-tab-pane label="菜单权限" name="menu">
          <MenuPermTab :role-id="currentRole!.id" :app-code="drawerAppCode" />
        </el-tab-pane>
        <el-tab-pane label="接口权限" name="api">
          <ApiPermTab :role-id="currentRole!.id" :app-code="drawerAppCode" />
        </el-tab-pane>
        <el-tab-pane label="数据权限" name="data">
          <DataScopeTab :role-id="currentRole!.id" />
        </el-tab-pane>
      </el-tabs>
    </el-drawer>

    <!-- 角色新增/编辑弹窗 -->
    <RoleForm ref="formRef" :role-tree="roleTree" @success="loadRoleTree" />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { ElTree } from 'element-plus'
import { getRoleList, deleteRole, getRolePermissionSummary } from '@/api/role'
import type { RoleItem, AppPermissionSummary } from '@/api/role'
import RoleForm from './RoleForm.vue'
import MenuPermTab from './MenuPermTab.vue'
import ApiPermTab from './ApiPermTab.vue'
import DataScopeTab from './DataScopeTab.vue'

const treeLoading = ref(false)
const roleTree = ref<RoleItem[]>([])
const currentRole = ref<RoleItem | null>(null)
const filterText = ref('')
const roleTreeRef = ref<InstanceType<typeof ElTree>>()
const formRef = ref()

// 权限汇总
const permissionSummary = ref<AppPermissionSummary[]>([])
const summaryLoading = ref(false)

// Drawer 状态
const drawerVisible = ref(false)
const drawerAppCode = ref('')
const drawerTitle = ref('')
const drawerTab = ref('menu')

/** 加载角色树 */
async function loadRoleTree() {
  treeLoading.value = true
  try {
    roleTree.value = await getRoleList()
  } finally {
    treeLoading.value = false
  }
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
</style>
