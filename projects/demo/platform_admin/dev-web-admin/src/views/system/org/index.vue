<template>
  <div class="org-management">
    <el-row :gutter="16">
      <!-- 左侧：组织架构树 -->
      <el-col :span="10">
        <el-card shadow="never">
          <template #header>
            <div class="card-header">
              <span>组织架构</span>
              <el-button type="primary" size="small" @click="handleAdd(null)">
                新增顶级节点
              </el-button>
            </div>
          </template>
          <el-tree
            ref="treeRef"
            :data="treeData"
            :props="{ label: 'name', children: 'children' }"
            node-key="id"
            highlight-current
            default-expand-all
            @node-click="handleNodeClick"
          >
            <template #default="{ node, data }">
              <span class="tree-node">
                <span>{{ data.name }}（{{ nodeTypeLabel(data.node_type) }}）</span>
                <span class="tree-actions">
                  <el-button type="primary" link size="small" @click.stop="handleAdd(data)">新增</el-button>
                  <el-button type="primary" link size="small" @click.stop="handleEdit(data)">编辑</el-button>
                  <el-button type="danger" link size="small" @click.stop="handleDelete(data)">删除</el-button>
                </span>
              </span>
            </template>
          </el-tree>
        </el-card>
      </el-col>

      <!-- 右侧：节点详情 + 用户管理 -->
      <el-col :span="14">
        <el-card v-if="currentNode" shadow="never">
          <template #header>
            <span>{{ currentNode.name }} - 人员管理</span>
          </template>
          <el-transfer
            v-model="selectedUserIds"
            :data="allUsers"
            :titles="['可选用户', '已分配用户']"
            :props="{ key: 'id', label: 'username' }"
            filterable
            filter-placeholder="搜索用户"
          />
          <div style="margin-top: 16px; text-align: right;">
            <el-button type="primary" @click="saveNodeUsers">保存分配</el-button>
          </div>
        </el-card>
        <el-empty v-else description="请在左侧选择组织节点" />
      </el-col>
    </el-row>

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="100px">
        <el-form-item label="节点类型" prop="node_type">
          <el-select v-model="formData.node_type" placeholder="选择节点类型">
            <el-option label="公司" value="COMPANY" />
            <el-option label="分公司" value="BRANCH" />
            <el-option label="部门" value="DEPARTMENT" />
            <el-option label="小组" value="GROUP" />
            <el-option label="团队" value="TEAM" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入名称" />
        </el-form-item>
        <el-form-item label="编码" prop="code">
          <el-input v-model="formData.code" placeholder="可选，租户内唯一" />
        </el-form-item>
        <el-form-item label="排序" prop="sort_order">
          <el-input-number v-model="formData.sort_order" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import {
  getOrgTree,
  createOrgUnit,
  updateOrgUnit,
  deleteOrgUnit,
  getNodeUsers,
  setNodeUsers,
  type OrgUnitItem
} from '@/api/org'
import { listUsers } from '@/api/user'

const treeRef = ref()
const treeData = ref<OrgUnitItem[]>([])
const currentNode = ref<OrgUnitItem | null>(null)
const selectedUserIds = ref<string[]>([])
const allUsers = ref<any[]>([])

// 对话框
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref<FormInstance>()
const isEdit = ref(false)
const editingId = ref('')
const editingVersion = ref(1)
const parentId = ref<string | null>(null)

const formData = ref({
  node_type: 'DEPARTMENT',
  name: '',
  code: '',
  sort_order: 0,
})

const formRules = {
  node_type: [{ required: true, message: '请选择节点类型', trigger: 'change' }],
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
}

const nodeTypeLabel = (type: string) => {
  const map: Record<string, string> = {
    COMPANY: '公司',
    BRANCH: '分公司',
    DEPARTMENT: '部门',
    GROUP: '小组',
    TEAM: '团队',
  }
  return map[type] || type
}

const loadTree = async () => {
  const res: any = await getOrgTree()
  treeData.value = res?.data || res || []
}

const loadAllUsers = async () => {
  try {
    const res: any = await listUsers({ page: 1, page_size: 1000 })
    allUsers.value = (res?.list || res?.data?.list || []).map((u: any) => ({
      id: u.id,
      username: u.nickname || u.username,
    }))
  } catch {
    allUsers.value = []
  }
}

const handleNodeClick = async (data: OrgUnitItem) => {
  currentNode.value = data
  try {
    const res: any = await getNodeUsers(data.id)
    const users = res?.data || res || []
    selectedUserIds.value = users.map((u: any) => u.user_id)
  } catch {
    selectedUserIds.value = []
  }
}

const saveNodeUsers = async () => {
  if (!currentNode.value) return
  try {
    await setNodeUsers(currentNode.value.id, {
      user_ids: selectedUserIds.value,
      is_primary: 1,
    })
    ElMessage.success('保存成功')
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  }
}

const handleAdd = (parent: OrgUnitItem | null) => {
  isEdit.value = false
  dialogTitle.value = '新增组织节点'
  parentId.value = parent?.id || null
  formData.value = { node_type: 'DEPARTMENT', name: '', code: '', sort_order: 0 }
  dialogVisible.value = true
}

const handleEdit = (data: OrgUnitItem) => {
  isEdit.value = true
  dialogTitle.value = '编辑组织节点'
  editingId.value = data.id
  editingVersion.value = data.version
  parentId.value = data.parent_id
  formData.value = {
    node_type: data.node_type,
    name: data.name,
    code: data.code || '',
    sort_order: data.sort_order,
  }
  dialogVisible.value = true
}

const handleDelete = async (data: OrgUnitItem) => {
  await ElMessageBox.confirm(`确定删除「${data.name}」？`, '提示', { type: 'warning' })
  try {
    await deleteOrgUnit(data.id)
    ElMessage.success('删除成功')
    loadTree()
  } catch (e: any) {
    ElMessage.error(e?.message || '删除失败，可能存在子节点')
  }
}

const submitForm = async () => {
  await formRef.value?.validate()
  try {
    if (isEdit.value) {
      await updateOrgUnit(editingId.value, {
        ...formData.value,
        version: editingVersion.value,
      })
    } else {
      await createOrgUnit({
        ...formData.value,
        parent_id: parentId.value,
      })
    }
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    dialogVisible.value = false
    loadTree()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

onMounted(() => {
  loadTree()
  loadAllUsers()
})
</script>

<style scoped>
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
.tree-actions {
  display: none;
}
.tree-node:hover .tree-actions {
  display: inline-flex;
}
</style>
