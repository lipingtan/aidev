<template>
  <div class="page-container" :class="{ 'mobile-layout': isMobile }">
    <el-card class="main-card">
      <!-- 固定头部区域（手机端） -->
      <div class="sticky-header">
        <div class="card-header">
          <span>数据权限维度配置</span>
          <el-button type="primary" size="small" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            <span v-if="!isMobile">新增维度</span>
          </el-button>
        </div>
      </div>

      <!-- 可滚动的列表区域 -->
      <div class="scrollable-content">
        <ResponsiveList ref="listRef" :data="tableData" :loading="loading">
          <!-- 桌面端表格 -->
          <template #table>
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
          </template>

          <!-- 手机端卡片 -->
          <template #card="{ row }">
            <div class="card-title">
              {{ row.display_name }}
              <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small" class="card-tag">
                {{ row.status === 1 ? '启用' : '禁用' }}
              </el-tag>
            </div>
            <div class="card-row">
              <span class="card-label">维度标识</span>
              <span class="card-value">{{ row.dimension_name }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">值来源</span>
              <span class="card-value">{{ row.value_source }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">描述</span>
              <span class="card-value">{{ row.description || '-' }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">创建时间</span>
              <span class="card-value">{{ row.created_at }}</span>
            </div>
          </template>

          <!-- 手机端操作按钮 -->
          <template #actions="{ row }">
            <el-button size="small" type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </ResponsiveList>
      </div>
    </el-card>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑维度' : '新增维度'" :width="isMobile ? '95%' : '500px'" destroy-on-close>
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
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import {
  listDataScopeConfigs,
  createDataScopeConfig,
  updateDataScopeConfig,
  deleteDataScopeConfig
} from '@/api/data-scope'
import type { DataScopeConfigItem, DataScopeConfigFormData } from '@/api/data-scope'
import ResponsiveList from '@/components/ResponsiveList.vue'

// ==================== 响应式检测 ====================
const isMobile = ref(false)
const MOBILE_BREAKPOINT = 768
const listRef = ref()

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
.page-container {
  padding: 16px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color-light);
}
.scrollable-content {
  padding: 0 16px 16px;
}

/* 手机端布局 - 固定头部 */
@media (max-width: 768px) {
  .page-container {
    padding: 0;
    height: 100vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .page-container .main-card {
    flex: 1;
    display: flex;
    flex-direction: column;
    border-radius: 0;
    overflow: hidden;
  }

  .page-container .main-card :deep(.el-card__body) {
    flex: 1;
    display: flex;
    flex-direction: column;
    padding: 0;
    overflow: hidden;
  }

  .sticky-header {
    flex-shrink: 0;
    background: var(--el-bg-color);
    z-index: 10;
  }

  .scrollable-content {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
    -webkit-overflow-scrolling: touch;
  }
}
</style>
