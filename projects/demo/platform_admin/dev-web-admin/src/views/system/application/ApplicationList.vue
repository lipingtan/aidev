<template>
  <div class="app-layout">
    <!-- 左侧：应用列表 -->
    <el-card shadow="never" class="app-list-card">
      <template #header>
        <div class="card-header">
          <span>应用列表</span>
          <el-button type="primary" size="small" @click="handleAdd">新增</el-button>
        </div>
      </template>
      <el-table
        v-loading="loading"
        :data="tableData"
        row-key="id"
        highlight-current-row
        @current-change="handleRowSelect"
        size="small"
      >
        <template #empty>
          <el-empty description="暂无数据" :image-size="60" />
        </template>
        <el-table-column prop="name" label="应用名称" min-width="100" />
        <el-table-column prop="app_code" label="编码" min-width="80" />
        <el-table-column label="状态" width="60">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 右侧：详情 Tabs -->
    <el-card shadow="never" class="app-detail-card">
      <el-empty v-if="!selectedApp" description="请选择一个应用" :image-size="80" />
      <template v-else>
        <el-tabs v-model="activeTab">
          <!-- Tab 1: 基本信息 -->
          <el-tab-pane label="基本信息" name="info">
            <el-form ref="infoFormRef" :model="infoForm" :rules="infoRules" label-width="80px" style="max-width: 500px">
              <el-form-item label="名称" prop="name">
                <el-input v-model="infoForm.name" />
              </el-form-item>
              <el-form-item label="编码">
                <el-input v-model="selectedApp.app_code" disabled />
              </el-form-item>
              <el-form-item label="描述">
                <el-input v-model="infoForm.description" type="textarea" :rows="3" />
              </el-form-item>
              <el-form-item label="状态">
                <el-switch v-model="infoForm.status" :active-value="1" :inactive-value="0" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :loading="infoSubmitting" @click="handleInfoSave">保存</el-button>
                <el-button type="danger" @click="handleDelete(selectedApp)">删除应用</el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>

          <!-- Tab 2: 菜单管理 -->
          <el-tab-pane label="菜单管理" name="menu">
            <ResourceTree :app-code="selectedApp.app_code" />
          </el-tab-pane>

          <!-- Tab 3: 对象API管理 -->
          <el-tab-pane label="对象API管理" name="api">
            <ApiPermissionTree :app-code="selectedApp.app_code" />
          </el-tab-pane>

          <!-- Tab 4: 租户订阅 -->
          <el-tab-pane label="租户订阅" name="subscription">
            <div class="sub-section">
              <el-form inline>
                <el-form-item label="选择租户">
                  <el-select v-model="selectedTenantId" placeholder="请选择租户" @change="handleTenantChange">
                    <el-option v-for="t in tenantList" :key="t.id" :label="t.name" :value="t.id" />
                  </el-select>
                </el-form-item>
              </el-form>
              <el-transfer
                v-model="subscribedAppCodes"
                :data="transferData"
                :titles="['未订阅', '已订阅']"
                :props="{ key: 'app_code', label: 'name' }"
                filterable
                filter-placeholder="搜索应用"
              />
              <div style="margin-top: 16px; text-align: right;">
                <el-button type="primary" :loading="subSubmitting" @click="handleSubSubmit">保存订阅</el-button>
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </template>
    </el-card>

    <!-- 新增应用弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEditDialog ? '编辑应用' : '新增应用'" width="500px" destroy-on-close>
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入应用名称" />
        </el-form-item>
        <el-form-item label="编码" prop="app_code">
          <el-input v-model="formData.app_code" placeholder="请输入应用编码" :disabled="isEditDialog" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="formData.description" type="textarea" :rows="3" placeholder="请输入描述" />
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
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  listApplications,
  createApplication,
  updateApplication,
  deleteApplication,
  getTenantApps,
  updateTenantApps
} from '@/api/application'
import type { ApplicationItem, ApplicationFormData } from '@/api/application'
import { listTenants } from '@/api/tenant'
import type { TenantItem } from '@/api/tenant'
import ResourceTree from '@/views/system/resource/ResourceTree.vue'
import ApiPermissionTree from '@/views/system/api-permission/ApiPermissionTree.vue'

// ==================== 应用列表 ====================
const loading = ref(false)
const tableData = ref<ApplicationItem[]>([])
const selectedApp = ref<ApplicationItem | null>(null)
const activeTab = ref('info')

async function fetchData() {
  loading.value = true
  try {
    const res = await listApplications()
    tableData.value = Array.isArray(res) ? res : (res as any).list || []
  } finally {
    loading.value = false
  }
}

function handleRowSelect(row: ApplicationItem | null) {
  selectedApp.value = row
  if (row) {
    infoForm.name = row.name
    infoForm.description = row.description || ''
    infoForm.status = row.status
  }
}

// ==================== 基本信息编辑 ====================
const infoFormRef = ref<FormInstance>()
const infoSubmitting = ref(false)
const infoForm = reactive({
  name: '',
  description: '',
  status: 1
})
const infoRules: FormRules = {
  name: [{ required: true, message: '请输入应用名称', trigger: 'blur' }]
}

async function handleInfoSave() {
  const valid = await infoFormRef.value?.validate().catch(() => false)
  if (!valid || !selectedApp.value) return

  infoSubmitting.value = true
  try {
    await updateApplication(selectedApp.value.id, {
      name: infoForm.name,
      app_code: selectedApp.value.app_code,
      description: infoForm.description,
      status: infoForm.status
    })
    ElMessage.success('保存成功')
    fetchData()
  } finally {
    infoSubmitting.value = false
  }
}

// ==================== 新增弹窗 ====================
const dialogVisible = ref(false)
const isEditDialog = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const formData = reactive<ApplicationFormData>({
  name: '',
  app_code: '',
  description: ''
})
const formRules: FormRules = {
  name: [{ required: true, message: '请输入应用名称', trigger: 'blur' }],
  app_code: [
    { required: true, message: '请输入应用编码', trigger: 'blur' },
    { pattern: /^[a-zA-Z][a-zA-Z0-9_-]*$/, message: '编码须以字母开头，仅含字母数字下划线横线', trigger: 'blur' }
  ]
}

function handleAdd() {
  isEditDialog.value = false
  formData.name = ''
  formData.app_code = ''
  formData.description = ''
  dialogVisible.value = true
}

async function handleDelete(row: ApplicationItem) {
  await ElMessageBox.confirm(`确认删除应用「${row.name}」？此操作不可恢复。`, '提示', { type: 'warning' })
  await deleteApplication(row.id)
  ElMessage.success('删除成功')
  selectedApp.value = null
  fetchData()
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    await createApplication(formData)
    ElMessage.success('新增成功')
    dialogVisible.value = false
    fetchData()
  } finally {
    submitting.value = false
  }
}

// ==================== 租户订阅管理 ====================
const subSubmitting = ref(false)
const tenantList = ref<TenantItem[]>([])
const selectedTenantId = ref('')
const subscribedAppCodes = ref<string[]>([])

const transferData = computed(() =>
  tableData.value.map((app) => ({ app_code: app.app_code, name: app.name }))
)

async function handleTenantChange(tenantId: string) {
  if (!tenantId) {
    subscribedAppCodes.value = []
    return
  }
  try {
    const apps = await getTenantApps(tenantId)
    subscribedAppCodes.value = Array.isArray(apps) ? apps.map((a: any) => a.app_code || a) : []
  } catch {
    subscribedAppCodes.value = []
  }
}

async function handleSubSubmit() {
  if (!selectedTenantId.value) {
    ElMessage.warning('请先选择租户')
    return
  }
  subSubmitting.value = true
  try {
    await updateTenantApps(selectedTenantId.value, subscribedAppCodes.value)
    ElMessage.success('订阅保存成功')
  } finally {
    subSubmitting.value = false
  }
}

// ==================== 初始化 ====================
onMounted(async () => {
  await fetchData()
  // 预加载租户列表
  try {
    const res: any = await listTenants({ page: 1, page_size: 999 })
    tenantList.value = res.data || res.list || []
  } catch {
    tenantList.value = []
  }
})
</script>

<style scoped>
.app-layout {
  display: flex;
  flex-direction: row;
  gap: 16px;
  padding: 16px;
  height: calc(100vh - 100px);
  min-height: 500px;
}
.app-list-card {
  width: 360px;
  min-width: 360px;
  flex-shrink: 0;
  overflow-y: auto;
}
.app-detail-card {
  flex: 1;
  min-width: 0;
  overflow-y: auto;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.sub-section {
  padding: 8px 0;
}
/* 窄屏：上下堆叠 */
@media (max-width: 767px) {
  .app-layout {
    flex-direction: column;
    height: auto;
  }
  .app-list-card {
    width: 100%;
    min-width: 0;
  }
}
</style>
