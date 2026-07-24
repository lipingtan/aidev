<template>
  <div class="page-container">
    <!-- 搜索栏 -->
    <el-card shadow="never" class="search-card">
      <el-form :model="queryParams" inline>
        <el-form-item label="手机号">
          <el-input v-model="queryParams.phone" placeholder="请输入手机号" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 操作栏 + 表格 -->
    <el-card shadow="never" class="table-card">
      <template #header>
        <div class="card-header">
          <span>C端用户列表</span>
          <el-button type="primary" @click="handleAdd">新增用户</el-button>
        </div>
      </template>

      <el-table v-loading="loading" :data="tableData" row-key="id">
        <template #empty>
          <el-empty description="暂无数据" />
        </template>
        <el-table-column prop="id" label="ID" width="200" show-overflow-tooltip />
        <el-table-column prop="phone" label="手机号" width="140" />
        <el-table-column prop="nickname" label="昵称" width="140" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-switch
              :model-value="row.status === 1"
              @change="(val: boolean) => handleToggleStatus(row, val)"
              inline-prompt
              active-text="启用"
              inactive-text="禁用"
            />
          </template>
        </el-table-column>
        <el-table-column prop="last_login_at" label="最后登录时间" width="180" />
        <el-table-column prop="last_login_ip" label="最后登录IP" width="140" />
        <el-table-column label="操作" min-width="280" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="warning" @click="handleResetPassword(row)">重置密码</el-button>
            <el-button link type="info" @click="handleForceLogout(row)">强制登出</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="pagination"
        v-model:current-page="queryParams.page"
        v-model:page-size="queryParams.page_size"
        :total="total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchData"
        @current-change="fetchData"
      />
    </el-card>

    <!-- 新增/编辑对话框 -->
    <BizUserForm ref="formRef" @success="fetchData" />

    <!-- 重置密码结果弹窗 -->
    <el-dialog v-model="resetPwdVisible" title="重置密码成功" width="420px" destroy-on-close>
      <el-alert type="success" :closable="false" show-icon>
        <template #title>
          <span>新密码已生成，请妥善保管：</span>
        </template>
      </el-alert>
      <div class="reset-pwd-display">
        <el-input :model-value="resetPwdResult" readonly>
          <template #append>
            <el-button @click="handleCopyPassword">复制</el-button>
          </template>
        </el-input>
      </div>
      <template #footer>
        <el-button type="primary" @click="resetPwdVisible = false">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getBizUsers,
  deleteBizUser,
  resetPassword,
  forceLogout,
  toggleStatus
} from '@/api/biz-user'
import type { BizUser } from '@/api/biz-user'
import BizUserForm from './components/BizUserForm.vue'

const loading = ref(false)
const tableData = ref<BizUser[]>([])
const total = ref(0)
const formRef = ref()

const queryParams = reactive({
  phone: '',
  page: 1,
  page_size: 20
})

// ==================== 列表逻辑 ====================

async function fetchData() {
  loading.value = true
  try {
    const res: any = await getBizUsers({ ...queryParams })
    tableData.value = res.list || res.data || []
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
  queryParams.phone = ''
  handleSearch()
}

function handleAdd() {
  formRef.value?.open()
}

function handleEdit(row: BizUser) {
  formRef.value?.open(row)
}

// ==================== 删除 ====================

async function handleDelete(row: BizUser) {
  await ElMessageBox.confirm(`确认删除用户「${row.phone}」？`, '提示', { type: 'warning' })
  await deleteBizUser(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

// ==================== 重置密码 ====================

const resetPwdVisible = ref(false)
const resetPwdResult = ref('')

async function handleResetPassword(row: BizUser) {
  await ElMessageBox.confirm(`确认重置用户「${row.phone}」的密码？`, '提示', { type: 'warning' })
  const res = await resetPassword(row.id)
  resetPwdResult.value = res.password
  resetPwdVisible.value = true
}

function handleCopyPassword() {
  navigator.clipboard.writeText(resetPwdResult.value).then(() => {
    ElMessage.success('已复制到剪贴板')
  }).catch(() => {
    ElMessage.warning('复制失败，请手动复制')
  })
}

// ==================== 强制登出 ====================

async function handleForceLogout(row: BizUser) {
  await ElMessageBox.confirm(`确认强制登出用户「${row.phone}」？此操作将使该用户所有会话失效。`, '提示', { type: 'warning' })
  await forceLogout(row.id)
  ElMessage.success('已强制登出')
}

// ==================== 启用/禁用 ====================

async function handleToggleStatus(row: BizUser, val: boolean) {
  const newStatus = val ? 1 : 0
  await toggleStatus(row.id, newStatus)
  row.status = newStatus
  ElMessage.success(val ? '已启用' : '已禁用')
}

onMounted(fetchData)
</script>

<style scoped>
.page-container { padding: 16px; }
.search-card { margin-bottom: 16px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.pagination { margin-top: 16px; justify-content: flex-end; }
.reset-pwd-display { margin-top: 12px; }
</style>
