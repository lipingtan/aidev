<template>
  <div class="page-container">
    <!-- 顶部筛选（仅当未外部传入 appCode 时显示） -->
    <el-card v-if="!props.appCode" shadow="never" class="search-card">
      <el-form inline>
        <el-form-item label="应用">
          <el-select v-model="internalAppCode" placeholder="请选择应用" @change="fetchTree">
            <el-option v-for="app in appList" :key="app" :label="app" :value="app" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleAdd(null)">新增根菜单</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 资源树 -->
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>菜单/资源树</span>
        </div>
      </template>

      <el-tree
        ref="treeRef"
        v-loading="loading"
        :data="treeData"
        node-key="id"
        default-expand-all
        draggable
        :allow-drop="allowDrop"
        @node-drop="handleDrop"
      >
        <template #default="{ node, data }">
          <div class="tree-node">
            <el-icon class="drag-handle"><Rank /></el-icon>
            <span class="tree-node__label">
              <el-tag :type="data.type === 'MENU' ? 'primary' : 'warning'" size="small" class="tree-node__tag">
                {{ data.type === 'MENU' ? '菜单' : '按钮' }}
              </el-tag>
              {{ data.name }}
              <span v-if="data.permission_code" class="tree-node__perm">[{{ data.permission_code }}]</span>
            </span>
            <span class="tree-node__actions">
              <el-button link type="primary" size="small" @click.stop="handleAdd(data)">新增子级</el-button>
              <el-button link type="primary" size="small" @click.stop="handleEdit(data)">编辑</el-button>
              <el-button link type="danger" size="small" @click.stop="handleDelete(data)">删除</el-button>
            </span>
          </div>
        </template>
      </el-tree>
    </el-card>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑资源' : '新增资源'" width="520px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="类型" prop="type">
          <el-radio-group v-model="form.type">
            <el-radio value="MENU">菜单</el-radio>
            <el-radio value="BUTTON">按钮</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入名称" />
        </el-form-item>
        <el-form-item v-if="form.type === 'MENU'" label="路径" prop="path">
          <el-input v-model="form.path" placeholder="请输入路由路径" />
        </el-form-item>
        <el-form-item v-if="form.type === 'MENU'" label="图标" prop="icon">
          <el-input v-model="form.icon" placeholder="请输入图标名称" />
        </el-form-item>
        <el-form-item label="权限码" prop="permission_code">
          <el-input v-model="form.permission_code" placeholder="如 system:user:list" />
        </el-form-item>
        <el-form-item label="应用编码" prop="app_code">
          <el-input v-model="form.app_code" placeholder="应用编码" disabled />
        </el-form-item>
        <el-form-item label="排序" prop="sort_order">
          <el-input-number v-model="form.sort_order" :min="0" :max="9999" />
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
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Rank } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import type Node from 'element-plus/es/components/tree/src/model/node'
import {
  getResourceTree,
  createResource,
  updateResource,
  deleteResource,
  sortResources,
} from '@/api/resource'
import { listApplications } from '@/api/application'
import type { ResourceTreeNode, ResourceForm, ResourceSortItem } from '@/api/resource'

// ===== Props =====
const props = defineProps<{
  appCode?: string
}>()

// ===== 状态 =====
const loading = ref(false)
const treeData = ref<ResourceTreeNode[]>([])
const treeRef = ref()
const internalAppCode = ref('')
const appList = ref<string[]>([])

// 外部传入 appCode 时使用外部值
const currentAppCode = () => props.appCode || internalAppCode.value

// ===== 弹窗 =====
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()
const editingId = ref<string | null>(null)

const defaultForm = (): ResourceForm => ({
  parent_id: null,
  name: '',
  type: 'MENU',
  path: '',
  icon: '',
  permission_code: '',
  app_code: currentAppCode(),
  sort_order: 0,
})

const form = reactive<ResourceForm>(defaultForm())

const rules: FormRules = {
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  app_code: [{ required: true, message: '应用编码不能为空', trigger: 'blur' }],
}

// ===== 数据加载 =====
async function fetchTree() {
  const code = currentAppCode()
  if (!code) return
  loading.value = true
  try {
    treeData.value = await getResourceTree({ app_code: code })
  } finally {
    loading.value = false
  }
}

// ===== 拖拽排序 =====
function allowDrop(draggingNode: Node, dropNode: Node, type: string): boolean {
  // 按钮不能作为父节点
  if (type === 'inner' && dropNode.data.type === 'BUTTON') return false
  return true
}

async function handleDrop() {
  // 收集当前树的排序信息
  const sortList: ResourceSortItem[] = []
  function walk(nodes: ResourceTreeNode[], parentId: string | null) {
    nodes.forEach((node, index) => {
      sortList.push({ id: node.id, parent_id: parentId, sort_order: index })
      if (node.children?.length) {
        walk(node.children, node.id)
      }
    })
  }
  walk(treeData.value, null)

  try {
    await sortResources(sortList)
    ElMessage.success('排序已保存')
  } catch {
    // 排序失败时刷新树
    fetchTree()
  }
}

// ===== 新增/编辑 =====
function handleAdd(parent: ResourceTreeNode | null) {
  isEdit.value = false
  editingId.value = null
  Object.assign(form, defaultForm())
  form.parent_id = parent?.id || null
  form.app_code = currentAppCode()
  dialogVisible.value = true
}

function handleEdit(row: ResourceTreeNode) {
  isEdit.value = true
  editingId.value = row.id
  Object.assign(form, {
    parent_id: row.parent_id,
    name: row.name,
    type: row.type,
    path: row.path || '',
    icon: row.icon || '',
    permission_code: row.permission_code || '',
    app_code: row.app_code,
    sort_order: row.sort_order || 0,
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEdit.value && editingId.value) {
      await updateResource(editingId.value, { ...form })
      ElMessage.success('编辑成功')
    } else {
      await createResource({ ...form })
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    fetchTree()
  } finally {
    submitLoading.value = false
  }
}

// ===== 删除 =====
async function handleDelete(row: ResourceTreeNode) {
  await ElMessageBox.confirm(`确认删除「${row.name}」？删除后不可恢复。`, '提示', { type: 'warning' })
  await deleteResource(row.id)
  ElMessage.success('删除成功')
  fetchTree()
}

// ===== 初始化 =====
watch(() => props.appCode, (val) => {
  if (val) fetchTree()
})

onMounted(async () => {
  if (props.appCode) {
    // 外部传入了 appCode，直接加载
    fetchTree()
    return
  }
  try {
    const apps = await listApplications()
    const list = Array.isArray(apps) ? apps : (apps as any).list || []
    appList.value = list.map((a: any) => a.app_code)
  } catch {
    appList.value = []
  }
  if (appList.value.length) {
    internalAppCode.value = appList.value[0]
    fetchTree()
  }
})
</script>

<style scoped>
.page-container { padding: 16px; }
.search-card { margin-bottom: 16px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.tree-node { display: flex; align-items: center; justify-content: space-between; width: 100%; padding-right: 8px; }
.tree-node__label { display: flex; align-items: center; gap: 6px; }
.tree-node__tag { margin-right: 4px; }
.tree-node__perm { color: #909399; font-size: 12px; margin-left: 4px; }
.tree-node__actions { flex-shrink: 0; }
.drag-handle { cursor: move; color: #c0c4cc; margin-right: 6px; font-size: 14px; }
.drag-handle:hover { color: #409eff; }
</style>
