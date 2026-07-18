<template>
  <div class="api-permission-page">
    <el-tabs v-model="activeSubTab" type="border-card">
      <!-- Sub Tab 1: 对象API树 -->
      <el-tab-pane label="对象API树" name="tree">
        <div class="tab-header">
          <span class="tree-stats" v-if="treeData.length">
            {{ groupCount }} 个对象 / {{ endpointCount }} 个接口
          </span>
          <el-button type="primary" size="small" @click="handleAddGroup()">新增分组</el-button>
        </div>
        <el-tree
          ref="treeRef"
          v-loading="treeLoading"
          :data="treeData"
          node-key="id"
          :props="treeProps"
          default-expand-all
          draggable
          :allow-drop="allowDrop"
          @node-drop="handleNodeDrop"
        >
          <template #default="{ node, data }">
            <span class="tree-node">
              <el-icon class="drag-handle"><Rank /></el-icon>
              <el-tag :type="data.type === 'GROUP' ? 'primary' : 'success'" size="small" class="node-tag">
                {{ data.type === 'GROUP' ? '分组' : 'API' }}
              </el-tag>
              <span class="node-label">{{ data.type === 'GROUP' ? (data.display_name || data.name) : '' }}</span>
              <span v-if="data.type === 'ENDPOINT'" class="node-method">
                <span v-if="data.display_name" class="node-display-name">{{ data.display_name }}</span>
                <el-tag size="small" :type="getMethodTagType(data.http_method)">{{ data.http_method }}</el-tag>
                <span class="node-path">{{ data.url_pattern }}</span>
              </span>
              <span class="node-actions">
                <el-switch
                  v-model="data.visible"
                  :active-value="1"
                  :inactive-value="0"
                  size="small"
                  style="margin-right: 8px"
                  @change="handleVisibleChange(data)"
                  @click.stop
                />
                <el-button link size="small" type="primary" @click.stop="handleEdit(data)">编辑</el-button>
                <el-button
                  v-if="data.type === 'GROUP'"
                  link
                  size="small"
                  type="primary"
                  @click.stop="handleAddEndpoint(data)"
                >新增接口</el-button>
                <el-button link size="small" type="danger" @click.stop="handleDelete(node, data)">删除</el-button>
              </span>
            </span>
          </template>
        </el-tree>
      </el-tab-pane>

      <!-- Sub Tab 2: 待分组接口 -->
      <el-tab-pane name="unassigned">
        <template #label>
          待分组接口
          <el-badge v-if="unassignedList.length" :value="unassignedList.length" :max="999" class="badge-margin" />
        </template>
        <div class="tab-header">
          <el-input v-model="unassignedFilter" placeholder="搜索接口名称/路径" clearable style="width: 300px" />
          <el-button size="small" @click="fetchUnassigned">刷新</el-button>
        </div>
        <p class="drag-tip">拖拽接口到「对象API树」Tab 中的分组节点完成归类</p>
        <el-table v-loading="unassignedLoading" :data="pagedUnassigned" size="small" row-key="id">
          <el-table-column prop="http_method" label="方法" width="80">
            <template #default="{ row }">
              <el-tag size="small" :type="getMethodTagType(row.http_method)">{{ row.http_method }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="url_pattern" label="路径" min-width="250" show-overflow-tooltip />
          <el-table-column prop="name" label="名称" min-width="180" show-overflow-tooltip />
          <el-table-column label="操作" width="120">
            <template #default="{ row }">
              <el-button link size="small" type="primary" @click="handleEdit(row)">编辑</el-button>
              <el-button link size="small" type="danger" @click="handleDeleteUnassigned(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-pagination
          v-if="filteredUnassigned.length > unassignedPageSize"
          small
          layout="total, prev, pager, next"
          :total="filteredUnassigned.length"
          :page-size="unassignedPageSize"
          v-model:current-page="unassignedPage"
          style="margin-top: 12px; justify-content: flex-end"
        />
      </el-tab-pane>
    </el-tabs>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="520px" destroy-on-close>
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="100px">
        <el-form-item label="节点类型" prop="type">
          <el-radio-group v-model="formData.type" :disabled="!!formData.id">
            <el-radio value="GROUP">分组</el-radio>
            <el-radio value="ENDPOINT">接口</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入名称" />
        </el-form-item>
        <el-form-item label="显示名称">
          <el-input v-model="formData.display_name" placeholder="中文显示名称（可选）" />
        </el-form-item>
        <el-form-item label="权限码">
          <el-input v-model="formData.permission_code" placeholder="请输入权限码（可选）" />
        </el-form-item>
        <el-form-item v-if="formData.type === 'ENDPOINT'" label="请求方法" prop="http_method">
          <el-select v-model="formData.http_method" placeholder="请选择">
            <el-option v-for="m in httpMethods" :key="m" :label="m" :value="m" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="formData.type === 'ENDPOINT'" label="URL模式" prop="url_pattern">
          <el-input v-model="formData.url_pattern" placeholder="/api/v1/xxx" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="formData.sort_order" :min="0" :max="9999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Rank } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getApiPermissionTree,
  getUnassignedEndpoints,
  createApiPermission,
  updateApiPermission,
  deleteApiPermission,
  moveApiPermission,
  toggleApiPermissionVisible
} from '@/api/api-permission'
import type { ApiPermissionNode, ApiPermissionNodeType } from '@/api/api-permission'

// ===== Props =====
const props = defineProps<{
  appCode?: string
}>()

// ==================== 树相关 ====================
const treeRef = ref()
const unassignedTreeRef = ref()
const treeLoading = ref(false)
const unassignedLoading = ref(false)
const treeData = ref<ApiPermissionNode[]>([])
const unassignedList = ref<ApiPermissionNode[]>([])
const treeProps = { label: 'name', children: 'children' }
const activeSubTab = ref('tree')

// 统计：对象（GROUP）数量和接口（ENDPOINT）总数
const groupCount = computed(() => {
  let count = 0
  function walk(nodes: ApiPermissionNode[]) {
    for (const n of nodes) {
      if (n.type === 'GROUP') count++
      if (n.children?.length) walk(n.children)
    }
  }
  walk(treeData.value)
  return count
})

const endpointCount = computed(() => {
  let count = 0
  function walk(nodes: ApiPermissionNode[]) {
    for (const n of nodes) {
      if (n.type === 'ENDPOINT') count++
      if (n.children?.length) walk(n.children)
    }
  }
  walk(treeData.value)
  // 加上待分组的
  count += unassignedList.value.length
  return count
})
const httpMethods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']

// ==================== 未分配列表搜索+分页 ====================
const unassignedFilter = ref('')
const unassignedPage = ref(1)
const unassignedPageSize = 50

const filteredUnassigned = computed(() => {
  let list = unassignedList.value
  if (unassignedFilter.value) {
    const kw = unassignedFilter.value.toLowerCase()
    list = list.filter(item =>
      item.name.toLowerCase().includes(kw) ||
      (item.url_pattern || '').toLowerCase().includes(kw)
    )
  }
  return list
})

const pagedUnassigned = computed(() => {
  const start = (unassignedPage.value - 1) * unassignedPageSize
  return filteredUnassigned.value.slice(start, start + unassignedPageSize)
})

async function fetchTree() {
  if (!props.appCode) return
  treeLoading.value = true
  try {
    treeData.value = await getApiPermissionTree(props.appCode)
  } finally {
    treeLoading.value = false
  }
}

async function fetchUnassigned() {
  if (!props.appCode) return
  unassignedLoading.value = true
  try {
    unassignedList.value = await getUnassignedEndpoints(props.appCode)
  } finally {
    unassignedLoading.value = false
  }
}

/** 仅允许拖入 GROUP 节点 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function allowDrop(_draggingNode: any, dropNode: any, type: string): boolean {
  if (type === 'inner') {
    return (dropNode.data as ApiPermissionNode).type === 'GROUP'
  }
  return true
}

/** 拖拽完成：调用 move 接口 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
async function handleNodeDrop(draggingNode: any, dropNode: any, type: string) {
  const dragData = draggingNode.data as ApiPermissionNode
  let targetParentId: string

  if (type === 'inner') {
    targetParentId = (dropNode.data as ApiPermissionNode).id
  } else {
    targetParentId = (dropNode.data as ApiPermissionNode).parent_id || ''
  }

  try {
    await moveApiPermission(dragData.id, { target_parent_id: targetParentId })
    ElMessage.success('移动成功')
    fetchTree()
    fetchUnassigned()
  } catch {
    // 移动失败，刷新恢复原状
    fetchTree()
    fetchUnassigned()
  }
}

// ==================== 弹窗相关 ====================
const dialogVisible = ref(false)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()

interface FormState {
  id?: string
  type: ApiPermissionNodeType
  parent_id: string | null
  name: string
  display_name: string
  permission_code: string
  http_method: string
  url_pattern: string
  sort_order: number
}

const formData = reactive<FormState>({
  id: undefined,
  type: 'GROUP',
  parent_id: null,
  name: '',
  display_name: '',
  permission_code: '',
  http_method: 'GET',
  url_pattern: '',
  sort_order: 0
})

const dialogTitle = computed(() => {
  if (formData.id) return '编辑节点'
  return formData.type === 'GROUP' ? '新增分组' : '新增接口'
})

const formRules: FormRules = {
  type: [{ required: true, message: '请选择节点类型', trigger: 'change' }],
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  http_method: [{ required: true, message: '请选择请求方法', trigger: 'change' }],
  url_pattern: [{ required: true, message: '请输入URL模式', trigger: 'blur' }]
}

function resetForm() {
  formData.id = undefined
  formData.type = 'GROUP'
  formData.parent_id = null
  formData.name = ''
  formData.display_name = ''
  formData.permission_code = ''
  formData.http_method = 'GET'
  formData.url_pattern = ''
  formData.sort_order = 0
}

function handleAddGroup() {
  resetForm()
  formData.type = 'GROUP'
  formData.parent_id = null
  dialogVisible.value = true
}

function handleAddEndpoint(parentNode: ApiPermissionNode) {
  resetForm()
  formData.type = 'ENDPOINT'
  formData.parent_id = parentNode.id
  dialogVisible.value = true
}

function handleEdit(data: ApiPermissionNode) {
  resetForm()
  formData.id = data.id
  formData.type = data.type
  formData.parent_id = data.parent_id
  formData.name = data.name
  formData.display_name = data.display_name || ''
  formData.permission_code = data.permission_code || ''
  formData.http_method = data.http_method || 'GET'
  formData.url_pattern = data.url_pattern || ''
  formData.sort_order = data.sort_order
  dialogVisible.value = true
}

async function handleDelete(_node: any, data: ApiPermissionNode) {
  const label = data.type === 'GROUP' ? '分组' : '接口'
  await ElMessageBox.confirm(`确认删除${label}「${data.name}」？`, '提示', { type: 'warning' })
  await deleteApiPermission(data.id)
  ElMessage.success('删除成功')
  fetchTree()
  fetchUnassigned()
}

async function handleDeleteUnassigned(row: ApiPermissionNode) {
  await ElMessageBox.confirm(`确认删除接口「${row.name}」？`, '提示', { type: 'warning' })
  await deleteApiPermission(row.id)
  ElMessage.success('删除成功')
  fetchUnassigned()
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    const payload: any = {
      parent_id: formData.parent_id,
      type: formData.type,
      name: formData.name,
      display_name: formData.display_name || undefined,
      permission_code: formData.permission_code || undefined,
      http_method: formData.type === 'ENDPOINT' ? formData.http_method : undefined,
      url_pattern: formData.type === 'ENDPOINT' ? formData.url_pattern : undefined,
      app_code: props.appCode || 'admin',
      sort_order: formData.sort_order
    }

    if (formData.id) {
      await updateApiPermission(formData.id, payload)
      ElMessage.success('编辑成功')
    } else {
      await createApiPermission(payload)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    fetchTree()
    fetchUnassigned()
  } finally {
    submitLoading.value = false
  }
}

// ==================== 工具方法 ====================

async function handleVisibleChange(data: ApiPermissionNode) {
  try {
    await toggleApiPermissionVisible(data.id, data.visible)
    ElMessage.success(data.visible ? '已显示' : '已隐藏')
    // GROUP 切换后刷新树以反映子节点变化
    if (data.type === 'GROUP') {
      fetchTree()
    }
  } catch {
    // 回滚
    data.visible = data.visible === 1 ? 0 : 1
    ElMessage.error('操作失败')
  }
}

function getMethodTagType(method?: string): 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  switch (method) {
    case 'GET': return 'success'
    case 'POST': return 'primary'
    case 'PUT': return 'warning'
    case 'DELETE': return 'danger'
    default: return 'info'
  }
}

watch(() => props.appCode, (val) => {
  if (val) {
    fetchTree()
    fetchUnassigned()
  }
})

onMounted(() => {
  if (props.appCode) {
    fetchTree()
    fetchUnassigned()
  }
})
</script>

<style scoped>
.api-permission-page {
  padding: 0;
}
.tab-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.tree-stats {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
.drag-tip {
  color: #909399;
  font-size: 12px;
  margin: 0 0 12px;
}
.tree-node {
  display: flex;
  align-items: center;
  flex: 1;
  font-size: 14px;
  padding-right: 8px;
}
.node-tag {
  margin-right: 8px;
}
.node-label {
  margin-right: 8px;
}
.node-method {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.node-display-name {
  font-weight: 500;
  margin-right: 4px;
}
.node-path {
  color: #909399;
  font-size: 12px;
}
.node-actions {
  margin-left: auto;
}
.drag-handle { cursor: move; color: #c0c4cc; margin-right: 6px; font-size: 14px; }
.drag-handle:hover { color: #409eff; }
.badge-margin { margin-left: 6px; }
</style>
