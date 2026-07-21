<template>
  <div class="config-management">
    <!-- Scope 切换 Tab -->
    <el-tabs v-model="activeScope" @tab-change="loadList">
      <el-tab-pane label="系统配置" name="SYSTEM" />
      <el-tab-pane label="租户配置" name="TENANT" />
      <el-tab-pane label="用户配置" name="USER" />
    </el-tabs>

    <!-- 工具栏 -->
    <div class="toolbar">
      <el-input v-model="queryKey" placeholder="按 key 搜索" clearable style="width: 240px" @clear="loadList" @keyup.enter="loadList" />
      <el-checkbox v-model="onlyFeatureFlag" label="仅功能开关" @change="loadList" />
      <el-button type="primary" @click="handleAdd">新增配置</el-button>
    </div>

    <!-- 表格 -->
    <el-table :data="list" border stripe style="width: 100%; margin-top: 12px">
      <el-table-column prop="config_key" label="Key" min-width="180" />
      <el-table-column prop="config_value" label="Value" min-width="150" show-overflow-tooltip />
      <el-table-column prop="config_type" label="类型" width="90" />
      <el-table-column prop="display_name" label="显示名" min-width="120" />
      <el-table-column prop="is_feature_flag" label="功能开关" width="90" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.is_feature_flag" type="warning" size="small">是</el-tag>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <el-pagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :total="total"
      layout="total, prev, pager, next"
      style="margin-top: 12px"
      @current-change="loadList"
    />

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑配置' : '新增配置'" width="560px">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="100px">
        <el-form-item label="Scope" prop="scope">
          <el-select v-model="formData.scope" :disabled="isEdit">
            <el-option label="系统" value="SYSTEM" />
            <el-option label="租户" value="TENANT" />
            <el-option label="用户" value="USER" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="formData.scope === 'USER'" label="目标用户">
          <el-select v-model="formData.scope_id" filterable placeholder="搜索用户">
            <el-option v-for="u in userOptions" :key="u.id" :label="u.username" :value="u.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="Key" prop="config_key">
          <el-input v-model="formData.config_key" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="Value" prop="config_value">
          <el-input v-model="formData.config_value" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="类型" prop="config_type">
          <el-select v-model="formData.config_type">
            <el-option label="字符串" value="string" />
            <el-option label="数字" value="number" />
            <el-option label="布尔" value="boolean" />
            <el-option label="JSON" value="json" />
          </el-select>
        </el-form-item>
        <el-form-item label="显示名">
          <el-input v-model="formData.display_name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="formData.description" type="textarea" />
        </el-form-item>
        <el-form-item label="功能开关">
          <el-switch v-model="formData.is_feature_flag" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <!-- 覆盖关系提示 -->
      <el-alert v-if="isEdit && systemDefault !== null" type="info" :closable="false" style="margin-top: 8px">
        系统默认值：{{ systemDefault || '（未设置系统默认）' }}
      </el-alert>
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
  listAdminConfigs,
  createAdminConfig,
  updateAdminConfig,
  deleteAdminConfig,
  resolveConfig,
  type AdminConfigItem,
} from '@/api/admin-config'
import { listUsers } from '@/api/user'

const activeScope = ref('SYSTEM')
const queryKey = ref('')
const onlyFeatureFlag = ref(false)
const list = ref<AdminConfigItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const dialogVisible = ref(false)
const isEdit = ref(false)
const editingId = ref('')
const systemDefault = ref<string | null>(null)
const formRef = ref<FormInstance>()
const userOptions = ref<any[]>([])

const formData = ref({
  scope: 'SYSTEM',
  scope_id: '0',
  config_key: '',
  config_value: '',
  config_type: 'string',
  display_name: '',
  description: '',
  is_feature_flag: 0,
})

const formRules = {
  scope: [{ required: true, message: '请选择 Scope', trigger: 'change' }],
  config_key: [{ required: true, message: '请输入 Key', trigger: 'blur' }],
}

const loadList = async () => {
  const params: any = {
    scope: activeScope.value,
    page: page.value,
    page_size: pageSize.value,
  }
  if (queryKey.value) params.key = queryKey.value
  if (onlyFeatureFlag.value) params.is_feature_flag = 1

  const res = await listAdminConfigs(params)
  list.value = res.list || []
  total.value = res.total || 0
}

const handleAdd = () => {
  isEdit.value = false
  systemDefault.value = null
  formData.value = {
    scope: activeScope.value,
    scope_id: '0',
    config_key: '',
    config_value: '',
    config_type: 'string',
    display_name: '',
    description: '',
    is_feature_flag: 0,
  }
  dialogVisible.value = true
}

const handleEdit = async (row: AdminConfigItem) => {
  isEdit.value = true
  editingId.value = row.id
  formData.value = {
    scope: row.scope,
    scope_id: row.scope_id,
    config_key: row.config_key,
    config_value: row.config_value,
    config_type: row.config_type,
    display_name: row.display_name || '',
    description: row.description || '',
    is_feature_flag: row.is_feature_flag,
  }
  // 查询系统默认值作为参考
  if (row.scope !== 'SYSTEM') {
    try {
      const res: any = await resolveConfig(row.config_key)
      systemDefault.value = res?.data?.value || '（未设置系统默认）'
    } catch {
      systemDefault.value = '（未设置系统默认）'
    }
  } else {
    systemDefault.value = null
  }
  dialogVisible.value = true
}

const handleDelete = async (row: AdminConfigItem) => {
  await ElMessageBox.confirm(`确定删除配置「${row.config_key}」？`, '提示', { type: 'warning' })
  try {
    await deleteAdminConfig(row.id)
    ElMessage.success('删除成功')
    loadList()
  } catch (e: any) {
    ElMessage.error(e?.message || '删除失败')
  }
}

const submitForm = async () => {
  await formRef.value?.validate()
  try {
    if (isEdit.value) {
      await updateAdminConfig(editingId.value, {
        config_value: formData.value.config_value,
        config_type: formData.value.config_type,
        display_name: formData.value.display_name,
        description: formData.value.description,
        is_feature_flag: formData.value.is_feature_flag,
      })
    } else {
      await createAdminConfig({
        ...formData.value,
        tenant_id: formData.value.scope === 'SYSTEM' ? '0' : formData.value.scope_id,
      } as any)
    }
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    dialogVisible.value = false
    loadList()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

const loadUserOptions = async () => {
  try {
    const res: any = await listUsers({ page: 1, page_size: 500 })
    userOptions.value = (res?.list || res?.data?.list || []).map((u: any) => ({
      id: u.id,
      username: u.nickname || u.username,
    }))
  } catch {
    userOptions.value = []
  }
}

onMounted(() => {
  loadList()
  loadUserOptions()
})
</script>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
}
</style>
