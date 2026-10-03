<template>
  <div class="page-container" :class="{ 'mobile-layout': isMobile }">
    <el-card shadow="never" class="main-card">
      <!-- 固定头部区域（手机端） -->
      <div class="sticky-header">
        <div class="card-header">
          <span>租户列表</span>
          <el-button type="primary" size="small" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            <span v-if="!isMobile">新增租户</span>
          </el-button>
        </div>

        <!-- 搜索栏 -->
        <el-form :model="queryParams" :inline="!isMobile" class="search-form">
          <el-form-item label="租户名称">
            <el-input v-model="queryParams.name" placeholder="请输入租户名称" clearable @keyup.enter="handleSearch" />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="queryParams.status" placeholder="全部" clearable>
              <el-option label="启用" :value="1" />
              <el-option label="禁用" :value="0" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 可滚动的列表区域 -->
      <div class="scrollable-content">
        <!-- 响应式列表 -->
        <ResponsiveList :data="tableData" :loading="loading">
        <!-- 桌面端表格 -->
        <template #table>
          <el-table-column prop="name" label="租户名称" min-width="140" />
          <el-table-column prop="tenant_code" label="租户编码" min-width="140" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-switch
                :model-value="row.status === 1"
                @change="(val: boolean) => handleStatusChange(row, val)"
                inline-prompt
                active-text="启用"
                inactive-text="禁用"
              />
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
            {{ row.name }}
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small" class="card-tag">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </div>
          <div class="card-row">
            <span class="card-label">租户编码</span>
            <span class="card-value">{{ row.tenant_code }}</span>
          </div>
          <div class="card-row">
            <span class="card-label">状态</span>
            <span class="card-value">
              <el-switch
                :model-value="row.status === 1"
                @change="(val: boolean) => handleStatusChange(row, val)"
                inline-prompt
                active-text="启用"
                inactive-text="禁用"
              />
            </span>
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

        <el-pagination
          class="pagination"
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.page_size"
          :total="total"
          :page-sizes="[10, 20, 50]"
          :layout="isMobile ? 'total, prev, pager, next' : 'total, sizes, prev, pager, next, jumper'"
          :small="isMobile"
          @size-change="fetchData"
          @current-change="fetchData"
        />
      </div>
    </el-card>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑租户' : '新增租户'" :width="isMobile ? '95%' : '500px'" destroy-on-close>
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入租户名称" />
        </el-form-item>
        <el-form-item label="编码" prop="tenant_code">
          <el-input v-model="formData.tenant_code" placeholder="请输入租户编码" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="配置">
          <el-input v-model="configJson" type="textarea" :rows="4" placeholder="JSON 格式配置（可选）" />
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
import type { FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  listTenants,
  createTenant,
  updateTenant,
  updateTenantStatus,
  deleteTenant
} from '@/api/tenant'
import type { TenantItem, TenantFormData } from '@/api/tenant'
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

// 列表状态
const loading = ref(false)
const tableData = ref<TenantItem[]>([])
const total = ref(0)

const queryParams = reactive({
  name: '',
  status: undefined as number | undefined,
  page: 1,
  page_size: 20
})

// 弹窗状态
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref('')
const formRef = ref<FormInstance>()
const configJson = ref('')

const formData = reactive<TenantFormData>({
  name: '',
  tenant_code: '',
  config: undefined
})

const formRules: FormRules = {
  name: [{ required: true, message: '请输入租户名称', trigger: 'blur' }],
  tenant_code: [
    { required: true, message: '请输入租户编码', trigger: 'blur' },
    { pattern: /^[a-zA-Z][a-zA-Z0-9_-]*$/, message: '编码须以字母开头，仅含字母数字下划线横线', trigger: 'blur' }
  ]
}

/** 获取列表数据 */
async function fetchData() {
  loading.value = true
  try {
    const res: any = await listTenants(queryParams)
    tableData.value = res.data || res.list || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  queryParams.page = 1
  fetchData()
}

function handleReset() {
  queryParams.name = ''
  queryParams.status = undefined
  handleSearch()
}

/** 新增 */
function handleAdd() {
  isEdit.value = false
  editId.value = ''
  resetForm()
  dialogVisible.value = true
}

/** 编辑 */
function handleEdit(row: TenantItem) {
  isEdit.value = true
  editId.value = row.id
  formData.name = row.name
  formData.tenant_code = row.tenant_code
  formData.config = row.config
  configJson.value = row.config ? JSON.stringify(row.config, null, 2) : ''
  dialogVisible.value = true
}

/** 状态切换 */
async function handleStatusChange(row: TenantItem, val: boolean) {
  const newStatus = val ? 1 : 0
  const action = newStatus === 1 ? '启用' : '禁用'
  try {
    await ElMessageBox.confirm(`确认${action}租户「${row.name}」？`, '提示', { type: 'warning' })
    await updateTenantStatus(row.id, newStatus)
    ElMessage.success(`${action}成功`)
    fetchData()
  } catch {
    // 用户取消，不做处理
  }
}

/** 删除 */
async function handleDelete(row: TenantItem) {
  await ElMessageBox.confirm(`确认删除租户「${row.name}」？此操作不可恢复。`, '提示', { type: 'warning' })
  await deleteTenant(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

/** 提交表单 */
async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  // 解析配置 JSON
  if (configJson.value.trim()) {
    try {
      formData.config = JSON.parse(configJson.value)
    } catch {
      ElMessage.error('配置 JSON 格式不正确')
      return
    }
  } else {
    formData.config = undefined
  }

  submitting.value = true
  try {
    if (isEdit.value) {
      await updateTenant(editId.value, formData)
      ElMessage.success('编辑成功')
    } else {
      await createTenant(formData)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    fetchData()
  } finally {
    submitting.value = false
  }
}

function resetForm() {
  formData.name = ''
  formData.tenant_code = ''
  formData.config = undefined
  configJson.value = ''
}

onMounted(fetchData)
</script>

<style scoped>
.page-container { padding: 16px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-form { margin-bottom: 0; }
.pagination { margin-top: 16px; justify-content: flex-end; }

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
    border-bottom: 1px solid var(--el-border-color-lighter);
  }

  .sticky-header .card-header {
    padding: 12px;
    border-bottom: 1px solid var(--el-border-color-light);
  }

  .scrollable-content {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
    -webkit-overflow-scrolling: touch;
  }

  .search-form {
    padding: 12px;
  }

  .search-form :deep(.el-form-item) {
    margin-bottom: 12px;
    margin-right: 0;
    width: 100%;
  }

  .search-form :deep(.el-input),
  .search-form :deep(.el-select) {
    width: 100% !important;
  }

  .pagination {
    justify-content: center;
    padding-bottom: 12px;
  }
}
</style>
