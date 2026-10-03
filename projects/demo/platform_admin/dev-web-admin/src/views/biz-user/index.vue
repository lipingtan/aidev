<template>
  <div class="page-container" :class="{ 'mobile-layout': isMobile }">
    <el-card shadow="never" class="main-card">
      <!-- 固定头部区域（手机端） -->
      <div class="sticky-header">
        <div class="card-header">
          <span>C端用户列表</span>
          <el-button type="primary" size="small" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            <span v-if="!isMobile">新增用户</span>
          </el-button>
        </div>

        <!-- 搜索栏 -->
        <el-form :model="queryParams" :inline="!isMobile" class="search-form">
          <el-form-item label="手机号">
            <el-input v-model="queryParams.phone" placeholder="请输入手机号" clearable @keyup.enter="handleSearch" />
          </el-form-item>
          <el-form-item label="邮箱">
            <el-input v-model="queryParams.email" placeholder="请输入邮箱" clearable @keyup.enter="handleSearch" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 可滚动的列表区域 -->
      <div class="scrollable-content">
        <ResponsiveList :data="tableData" :loading="loading">
        <!-- 桌面端表格列 -->
        <template #table>
          <el-table-column prop="id" label="ID" width="200" show-overflow-tooltip />
          <el-table-column prop="phone" label="手机号" width="140">
            <template #default="{ row }">{{ row.phone || '-' }}</template>
          </el-table-column>
          <el-table-column prop="email" label="邮箱" width="200" show-overflow-tooltip>
            <template #default="{ row }">{{ row.email || '-' }}</template>
          </el-table-column>
          <el-table-column prop="nickname" label="昵称" width="120">
            <template #default="{ row }">{{ row.nickname || '-' }}</template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="110">
            <template #default="{ row }">
              <el-tooltip v-if="row.status === 0" placement="top">
                <template #content>
                  禁用时间：{{ row.disabled_at || '-' }}<br />
                  禁用原因（内部）：{{ row.disabled_reason || '-' }}<br />
                  用户展示原因：{{ row.disabled_reason_user || '-' }}
                </template>
                <el-tag type="danger" size="small">已禁用</el-tag>
              </el-tooltip>
              <el-tooltip v-else-if="row.enabled_at" placement="top">
                <template #content>
                  解禁时间：{{ row.enabled_at || '-' }}
                </template>
                <el-tag type="success" size="small">启用</el-tag>
              </el-tooltip>
              <el-tag v-else :type="row.status === 1 ? 'success' : 'info'" size="small">启用</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="last_login_at" label="最后登录" width="170">
            <template #default="{ row }">{{ row.last_login_at || '-' }}</template>
          </el-table-column>
          <el-table-column prop="last_login_ip" label="登录IP" width="140">
            <template #default="{ row }">{{ row.last_login_ip || '-' }}</template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="170" />
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
              <el-dropdown trigger="click">
                <el-button link type="primary">更多<el-icon class="el-icon--right"><ArrowDown /></el-icon></el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="handleToggleStatus(row, row.status === 1)">
                      {{ row.status === 1 ? '禁用' : '启用' }}
                    </el-dropdown-item>
                    <el-dropdown-item @click="handleResetPassword(row)">重置密码</el-dropdown-item>
                    <el-dropdown-item @click="handleForceLogout(row)">强制登出</el-dropdown-item>
                    <el-dropdown-item divided @click="handleDelete(row)">
                      <span style="color: var(--el-color-danger)">删除</span>
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
          </el-table-column>
        </template>

        <!-- 手机端卡片内容 -->
        <template #card="{ row }">
          <div class="card-title">
            {{ row.nickname || '未设置昵称' }}
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small" class="card-tag">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </div>
          <div class="card-row">
            <span class="card-label">ID</span>
            <span class="card-value">{{ row.id }}</span>
          </div>
          <div class="card-row">
            <span class="card-label">手机号</span>
            <span class="card-value">{{ row.phone || '-' }}</span>
          </div>
          <div class="card-row">
            <span class="card-label">邮箱</span>
            <span class="card-value">{{ row.email || '-' }}</span>
          </div>
          <div class="card-row">
            <span class="card-label">最后登录</span>
            <span class="card-value">{{ row.last_login_at || '-' }}</span>
          </div>
          <div class="card-row">
            <span class="card-label">登录IP</span>
            <span class="card-value">{{ row.last_login_ip || '-' }}</span>
          </div>
          <div class="card-row">
            <span class="card-label">创建时间</span>
            <span class="card-value">{{ row.created_at }}</span>
          </div>
        </template>

        <!-- 手机端卡片操作 -->
        <template #actions="{ row }">
          <el-button size="small" type="primary" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" :type="row.status === 1 ? 'warning' : 'success'" @click="handleToggleStatus(row, row.status === 1)">
            {{ row.status === 1 ? '禁用' : '启用' }}
          </el-button>
          <el-dropdown trigger="click">
            <el-button size="small">更多<el-icon class="el-icon--right"><ArrowDown /></el-icon></el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="handleResetPassword(row)">重置密码</el-dropdown-item>
                <el-dropdown-item @click="handleForceLogout(row)">强制登出</el-dropdown-item>
                <el-dropdown-item divided @click="handleDelete(row)">
                  <span style="color: var(--el-color-danger)">删除</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </ResponsiveList>

      <!-- 分页 -->
      <el-pagination
        class="pagination"
        v-model:current-page="queryParams.page"
        v-model:page-size="queryParams.page_size"
        :total="total"
        :page-sizes="[20, 50, 100]"
        :layout="isMobile ? 'total, prev, pager, next' : 'total, sizes, prev, pager, next, jumper'"
        :small="isMobile"
        @size-change="fetchData"
        @current-change="fetchData"
      />
      </div>
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

    <!-- FR-21: 禁用确认弹框（填写管理端原因 + 用户展示原因） -->
    <el-dialog v-model="statusDialogVisible" title="禁用账号" width="480px" destroy-on-close>
      <el-alert type="warning" :closable="false" show-icon
                title="禁用后该用户无法购买套餐、使用激活码和游玩 VIP 游戏（登录不受影响）"
                style="margin-bottom: 16px" />
      <el-form label-width="110px">
        <el-form-item label="内部原因" required>
          <el-input v-model="statusForm.adminReason" type="textarea" :rows="2" maxlength="500"
                    show-word-limit placeholder="管理端内部原因（仅后台可见）" />
        </el-form-item>
        <el-form-item label="用户展示原因" required>
          <el-input v-model="statusForm.userReason" type="textarea" :rows="2" maxlength="500"
                    show-word-limit placeholder="将提示给用户（App 端可见）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="statusDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmDisable">确定禁用</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, Plus } from '@element-plus/icons-vue'
import {
  getBizUsers,
  deleteBizUser,
  resetPassword,
  forceLogout,
  toggleStatus
} from '@/api/biz-user'
import type { BizUser } from '@/api/biz-user'
import BizUserForm from './components/BizUserForm.vue'
import ResponsiveList from '@/components/ResponsiveList.vue'

// ==================== 响应式检测 ====================
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

// ==================== 数据 ====================
const loading = ref(false)
const tableData = ref<BizUser[]>([])
const total = ref(0)
const formRef = ref()

const queryParams = reactive({
  phone: '',
  email: '',
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
  queryParams.email = ''
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
  const name = row.nickname || row.phone || row.email || '该用户'
  await ElMessageBox.confirm(`确认删除用户「${name}」？`, '提示', { type: 'warning' })
  await deleteBizUser(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

// ==================== 重置密码 ====================

const resetPwdVisible = ref(false)
const resetPwdResult = ref('')

async function handleResetPassword(row: BizUser) {
  const name = row.nickname || row.phone || row.email || '该用户'
  await ElMessageBox.confirm(`确认重置用户「${name}」的密码？`, '提示', { type: 'warning' })
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
  const name = row.nickname || row.phone || row.email || '该用户'
  await ElMessageBox.confirm(`确认强制登出用户「${name}」？此操作将使该用户所有会话失效。`, '提示', { type: 'warning' })
  await forceLogout(row.id)
  ElMessage.success('已强制登出')
}

// ==================== 启用/禁用 ====================

// FR-21: 禁用/启用弹框（禁用时填写两个原因）
const statusDialogVisible = ref(false)
const statusDialogMode = ref<'disable' | 'enable'>('disable')
const statusForm = reactive({ adminReason: '', userReason: '' })
const statusTargetRow = ref<BizUser | null>(null)

function handleToggleStatus(row: BizUser, val: boolean) {
  statusTargetRow.value = row
  if (val) {
    // 禁用：弹框填两个原因
    statusForm.adminReason = ''
    statusForm.userReason = ''
    statusDialogMode.value = 'disable'
    statusDialogVisible.value = true
  } else {
    // 启用：确认即可
    ElMessageBox.confirm('确定启用该账号吗？启用后恢复正常使用。', '启用确认', {
      confirmButtonText: '确定', cancelButtonText: '取消', type: 'info'
    }).then(async () => {
      await doToggle(row, 1, '', '')
    }).catch(() => {})
  }
}

async function doToggle(row: BizUser, newStatus: number, adminReason: string, userReason: string) {
  await toggleStatus(row.id, newStatus, adminReason, userReason)
  row.status = newStatus
  statusDialogVisible.value = false
  ElMessage.success(newStatus === 1 ? '已启用' : '已禁用')
}

async function confirmDisable() {
  if (!statusForm.adminReason.trim() || !statusForm.userReason.trim()) {
    ElMessage.warning('请填写两个原因')
    return
  }
  if (statusTargetRow.value) {
    await doToggle(statusTargetRow.value, 0, statusForm.adminReason.trim(), statusForm.userReason.trim())
  }
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

.search-form {
  padding: 12px 16px;
  margin-bottom: 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.scrollable-content {
  padding: 0 16px 16px;
}

.pagination {
  margin-top: 16px;
  justify-content: flex-end;
}

.reset-pwd-display {
  margin-top: 12px;
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

  .search-form {
    padding: 12px;
  }

  .search-form :deep(.el-form-item) {
    margin-bottom: 12px;
    margin-right: 0;
    width: 100%;
  }

  .search-form :deep(.el-input) {
    width: 100%;
  }

  .pagination {
    justify-content: center;
    padding-bottom: 12px;
  }
}
</style>
