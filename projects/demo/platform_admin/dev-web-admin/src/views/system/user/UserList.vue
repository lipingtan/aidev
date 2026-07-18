<template>
  <div class="page-container">
    <!-- 搜索栏 -->
    <el-card shadow="never" class="search-card">
      <el-form :model="queryParams" inline>
        <el-form-item label="用户名">
          <el-input v-model="queryParams.username" placeholder="请输入用户名" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="queryParams.email" placeholder="请输入邮箱" clearable @keyup.enter="handleSearch" />
        </el-form-item>
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
          <span>用户列表</span>
          <el-button type="primary" @click="handleAdd">新增用户</el-button>
        </div>
      </template>

      <el-table v-loading="loading" :data="tableData" row-key="id">
        <template #empty>
          <el-empty description="暂无数据" />
        </template>
        <el-table-column prop="username" label="用户名" width="140" />
        <el-table-column prop="email" label="邮箱" min-width="180" />
        <el-table-column prop="phone" label="手机号" width="140" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
            <el-dropdown trigger="click">
              <el-button link type="primary">更多<el-icon class="el-icon--right"><ArrowDown /></el-icon></el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="handleTenantManage(row)">租户关联</el-dropdown-item>
                  <el-dropdown-item @click="handleRoleAssign(row)">角色分配</el-dropdown-item>
                  <el-dropdown-item @click="handleForceOffline(row)">强制下线</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
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
    <UserForm ref="formRef" @success="fetchData" />

    <!-- 租户关联管理对话框 -->
    <el-dialog v-model="tenantDialogVisible" title="租户关联管理" width="700px" destroy-on-close>
      <el-transfer
        v-model="userTenantIds"
        :data="allTenantOptions"
        :titles="['可选租户', '已关联租户']"
        :props="{ key: 'id', label: 'name' }"
        filterable
        filter-placeholder="搜索租户"
      />
      <template #footer>
        <el-button @click="tenantDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="tenantSaving" @click="handleTenantSave">确定</el-button>
      </template>
    </el-dialog>

    <!-- 角色分配对话框 -->
    <el-dialog v-model="roleDialogVisible" title="角色分配" width="500px" destroy-on-close>
      <el-checkbox-group v-model="selectedRoleIds">
        <el-checkbox
          v-for="role in currentTenantRoles"
          :key="role.id"
          :label="role.id"
          :value="role.id"
        >
          {{ role.name }}（{{ role.code }}）
        </el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="roleSaving" @click="handleRoleSave">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import {
  listUsers,
  deleteUser,
  getUserTenants,
  addUserTenants,
  removeUserTenant,
  assignUserRoles,
  forceOfflineUser
} from '@/api/user'
import type { UserPageItem, UserTenant } from '@/api/user'
import { listTenants } from '@/api/tenant'
import type { TenantItem } from '@/api/tenant'
import { getRoleList } from '@/api/role'
import type { RoleItem } from '@/api/role'
import UserForm from './UserForm.vue'

const loading = ref(false)
const tableData = ref<UserPageItem[]>([])
const total = ref(0)
const formRef = ref()

const queryParams = reactive({
  username: '',
  email: '',
  phone: '',
  page: 1,
  page_size: 20
})

// ==================== 列表逻辑 ====================

async function fetchData() {
  loading.value = true
  try {
    const res: any = await listUsers({ ...queryParams })
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
  queryParams.username = ''
  queryParams.email = ''
  queryParams.phone = ''
  handleSearch()
}

function handleAdd() {
  formRef.value?.open()
}

function handleEdit(row: UserPageItem) {
  formRef.value?.open(row)
}

async function handleDelete(row: UserPageItem) {
  await ElMessageBox.confirm(`确认删除用户「${row.username}」？`, '提示', { type: 'warning' })
  await deleteUser(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

async function handleForceOffline(row: UserPageItem) {
  await ElMessageBox.confirm(`确认强制下线用户「${row.username}」？`, '提示', { type: 'warning' })
  await forceOfflineUser(row.id)
  ElMessage.success('已强制下线')
}

// ==================== 租户关联管理 ====================

const tenantDialogVisible = ref(false)
const tenantSaving = ref(false)
const currentUserId = ref('')
const userTenantIds = ref<string[]>([])
const originalTenantIds = ref<string[]>([])
const allTenantOptions = ref<{ id: string; name: string }[]>([])

async function handleTenantManage(row: UserPageItem) {
  currentUserId.value = row.id
  tenantDialogVisible.value = true
  // 并行加载全量租户和用户已关联租户
  const [tenantsRes, userTenants] = await Promise.all([
    listTenants({ page: 1, page_size: 1000 }) as any,
    getUserTenants(row.id)
  ])
  const tenantList: TenantItem[] = tenantsRes.list || tenantsRes.data || tenantsRes || []
  allTenantOptions.value = tenantList.map((t: TenantItem) => ({ id: t.id, name: t.name }))
  const ids = (userTenants as UserTenant[]).map((t: UserTenant) => t.id)
  userTenantIds.value = [...ids]
  originalTenantIds.value = [...ids]
}

async function handleTenantSave() {
  tenantSaving.value = true
  try {
    const toAdd = userTenantIds.value.filter(id => !originalTenantIds.value.includes(id))
    const toRemove = originalTenantIds.value.filter(id => !userTenantIds.value.includes(id))
    // 新增关联
    if (toAdd.length) {
      await addUserTenants(currentUserId.value, { tenant_ids: toAdd })
    }
    // 移除关联
    for (const tid of toRemove) {
      await removeUserTenant(currentUserId.value, tid)
    }
    ElMessage.success('租户关联已更新')
    tenantDialogVisible.value = false
  } finally {
    tenantSaving.value = false
  }
}

// ==================== 角色分配 ====================

const roleDialogVisible = ref(false)
const roleSaving = ref(false)
const roleUserId = ref('')
const selectedRoleIds = ref<string[]>([])
const currentTenantRoles = ref<{ id: string; name: string; code: string }[]>([])

async function handleRoleAssign(row: UserPageItem) {
  roleUserId.value = row.id
  roleDialogVisible.value = true
  selectedRoleIds.value = []
  // 加载当前租户下角色列表
  const roles: RoleItem[] = await getRoleList()
  // 展平树形为平铺列表
  const flat: { id: string; name: string; code: string }[] = []
  function flatten(nodes: RoleItem[]) {
    for (const n of nodes) {
      flat.push({ id: n.id, name: n.role_name, code: n.role_code })
      if (n.children) flatten(n.children)
    }
  }
  flatten(roles)
  currentTenantRoles.value = flat
}

async function handleRoleSave() {
  roleSaving.value = true
  try {
    const tenantId = localStorage.getItem('current_tenant_id') || ''
    await assignUserRoles(roleUserId.value, {
      tenant_id: tenantId,
      roles: selectedRoleIds.value.map(id => ({ role_id: id }))
    })
    ElMessage.success('角色分配成功')
    roleDialogVisible.value = false
  } finally {
    roleSaving.value = false
  }
}

onMounted(fetchData)
</script>

<style scoped>
.page-container { padding: 16px; }
.search-card { margin-bottom: 16px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
