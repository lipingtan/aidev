<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>数据权限维度配置</span>
          <el-button type="primary" @click="handleAdd">新增维度</el-button>
        </div>
      </template>

      <el-table v-loading="loading" :data="tableData" row-key="id">
        <template #empty>
          <el-empty description="暂无数据" />
        </template>
        <el-table-column prop="dimension_name" label="维度标识" min-width="140" />
        <el-table-column prop="display_name" label="维度名称" min-width="140" />
        <el-table-column prop="value_source" label="值来源" min-width="200" show-overflow-tooltip />
        <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑维度' : '新增维度'" width="500px" destroy-on-close>
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="90px">
        <el-form-item label="维度标识" prop="dimension_name">
          <el-input v-model="formData.dimension_name" placeholder="如 dept / region" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="维度名称" prop="display_name">
          <el-input v-model="formData.display_name" placeholder="如 部门 / 区域" />
        </el-form-item>
        <el-form-item label="值来源" prop="value_source">
          <el-input v-model="formData.value_source" placeholder="如 table:departments 或 api:/depts" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="formData.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="formData.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  listDataScopeConfigs,
  createDataScopeConfig,
  updateDataScopeConfig,
  deleteDataScopeConfig
} from '@/api/data-scope'
import type { DataScopeConfigItem, DataScopeConfigFormData } from '@/api/data-scope'

const loading = ref(false)
const tableData = ref<DataScopeConfigItem[]>([])

async function fetchData() {
  loading.value = true
  try {
    const res = await listDataScopeConfigs()
    tableData.value = Array.isArray(res) ? res : (res as any).list || []
  } finally {
    loading.value = false
  }
}

// ==================== CRUD 弹窗 ====================
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref('')
const formRef = ref<FormInstance>()

const formData = reactive<DataScopeConfigFormData>({
  dimension_name: '',
  display_name: '',
  value_source: '',
  description: '',
  status: 1
})

const formRules: FormRules = {
  dimension_name: [
    { required: true, message: '请输入维度标识', trigger: 'blur' },
    { pattern: /^[a-zA-Z][a-zA-Z0-9_]*$/, message: '标识须以字母开头，仅含字母数字下划线', trigger: 'blur' }
  ],
  display_name: [{ required: true, message: '请输入维度名称', trigger: 'blur' }],
  value_source: [{ required: true, message: '请输入值来源', trigger: 'blur' }]
}

function handleAdd() {
  isEdit.value = false
  editId.value = ''
  resetForm()
  dialogVisible.value = true
}

function handleEdit(row: DataScopeConfigItem) {
  isEdit.value = true
  editId.value = row.id
  formData.dimension_name = row.dimension_name
  formData.display_name = row.display_name
  formData.value_source = row.value_source
  formData.description = row.description || ''
  formData.status = row.status
  dialogVisible.value = true
}

async function handleDelete(row: DataScopeConfigItem) {
  await ElMessageBox.confirm(`确认删除维度「${row.display_name}」？`, '提示', { type: 'warning' })
  await deleteDataScopeConfig(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (isEdit.value) {
      await updateDataScopeConfig(editId.value, formData)
      ElMessage.success('编辑成功')
    } else {
      await createDataScopeConfig(formData)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    fetchData()
  } finally {
    submitting.value = false
  }
}

function resetForm() {
  formData.dimension_name = ''
  formData.display_name = ''
  formData.value_source = ''
  formData.description = ''
  formData.status = 1
}

onMounted(fetchData)
</script>

<style scoped>
.page-container { padding: 16px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
</style>
