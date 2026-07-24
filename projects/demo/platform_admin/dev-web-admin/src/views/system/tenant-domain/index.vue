<template>
  <div class="page-container">
    <!-- 搜索区域 -->
    <el-card shadow="never" class="search-card">
      <el-form :model="queryParams" inline>
        <el-form-item label="域名">
          <el-input v-model="queryParams.domain" placeholder="请输入域名" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 表格区域 -->
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>域名映射列表</span>
          <el-button type="primary" @click="handleAdd">新增映射</el-button>
        </div>
      </template>

      <el-table v-loading="loading" :data="tableData" row-key="id">
        <template #empty>
          <el-empty description="暂无数据" />
        </template>
        <el-table-column prop="domain" label="域名" min-width="200" />
        <el-table-column prop="tenant_name" label="租户名称" min-width="140" />
        <el-table-column prop="tenant_code" label="租户编码" min-width="120" />
        <el-table-column prop="match_type" label="匹配方式" width="100" />
        <el-table-column prop="remark" label="备注" min-width="140" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="pagination"
        v-model:current-page="queryParams.page"
        v-model:page-size="queryParams.page_size"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchData"
        @current-change="fetchData"
      />
    </el-card>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑域名映射' : '新增域名映射'" width="500px" destroy-on-close>
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="80px">
        <el-form-item label="域名" prop="domain">
          <el-input v-model="formData.domain" placeholder="请输入域名（如 app.example.com）" />
        </el-form-item>
        <el-form-item label="租户" prop="tenant_id">
          <el-select
            v-model="formData.tenant_id"
            filterable
            placeholder="请选择租户"
            style="width: 100%"
            :loading="tenantLoading"
          >
            <el-option
              v-for="t in tenantOptions"
              :key="t.id"
              :label="`${t.name}（${t.tenant_code}）`"
              :value="t.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="formData.remark" type="textarea" :rows="3" placeholder="备注信息（可选）" />
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
  listTenantDomains,
  createTenantDomain,
  updateTenantDomain,
  deleteTenantDomain
} from '@/api/tenant-domain'
import type { TenantDomainItem } from '@/api/tenant-domain'
import request from '@/utils/request'

// 租户下拉选项
interface TenantOption { id: string; name: string; tenant_code: string }
const tenantOptions = ref<TenantOption[]>([])
const tenantLoading = ref(false)

/** 加载租户列表（供下拉选择） */
async function fetchTenantOptions() {
  tenantLoading.value = true
  try {
    const res: any = await request.get('/api/v1/admin/tenants', { params: { page: 1, page_size: 200 } })
    const list = res.data?.list || res.list || []
    tenantOptions.value = list.map((t: any) => ({
      id: String(t.id),
      name: t.name,
      tenant_code: t.tenant_code
    }))
  } finally {
    tenantLoading.value = false
  }
}

// 列表状态
const loading = ref(false)
const tableData = ref<TenantDomainItem[]>([])
const total = ref(0)

const queryParams = reactive({
  domain: '',
  page: 1,
  page_size: 20
})

// 弹窗状态
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref('')
const editVersion = ref(0)
const formRef = ref<FormInstance>()

const formData = reactive({
  domain: '',
  tenant_id: '',
  remark: ''
})

const formRules: FormRules = {
  domain: [
    { required: true, message: '请输入域名', trigger: 'blur' },
    { max: 255, message: '域名长度不能超过 255', trigger: 'blur' }
  ],
  tenant_id: [
    { required: true, message: '请选择租户', trigger: 'change' }
  ]
}

/** 获取列表数据 */
async function fetchData() {
  loading.value = true
  try {
    const res: any = await listTenantDomains(queryParams)
    tableData.value = res.data?.list || res.list || []
    total.value = res.data?.total || res.total || 0
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  queryParams.page = 1
  fetchData()
}

function handleReset() {
  queryParams.domain = ''
  handleSearch()
}

/** 新增 */
function handleAdd() {
  isEdit.value = false
  editId.value = ''
  editVersion.value = 0
  resetForm()
  dialogVisible.value = true
}

/** 编辑 */
function handleEdit(row: TenantDomainItem) {
  isEdit.value = true
  editId.value = row.id
  editVersion.value = row.version
  formData.domain = row.domain
  formData.tenant_id = row.tenant_id
  formData.remark = row.remark
  dialogVisible.value = true
}

/** 删除 */
async function handleDelete(row: TenantDomainItem) {
  await ElMessageBox.confirm(`确认删除域名「${row.domain}」的映射？`, '提示', { type: 'warning' })
  await deleteTenantDomain(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

/** 提交表单 */
async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (isEdit.value) {
      await updateTenantDomain(editId.value, {
        domain: formData.domain,
        tenant_id: formData.tenant_id,
        remark: formData.remark,
        version: editVersion.value
      })
      ElMessage.success('编辑成功')
    } else {
      await createTenantDomain({
        domain: formData.domain,
        tenant_id: formData.tenant_id,
        remark: formData.remark
      })
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    fetchData()
  } finally {
    submitting.value = false
  }
}

function resetForm() {
  formData.domain = ''
  formData.tenant_id = ''
  formData.remark = ''
}

onMounted(() => {
  fetchData()
  fetchTenantOptions()
})
</script>

<style scoped>
.page-container { padding: 16px; }
.search-card { margin-bottom: 16px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
