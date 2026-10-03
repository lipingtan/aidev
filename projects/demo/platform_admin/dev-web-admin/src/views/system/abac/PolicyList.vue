<template>
  <div class="page-container" :class="{ 'mobile-layout': isMobile }">
    <el-card class="main-card">
      <!-- 固定头部区域（手机端） -->
      <div class="sticky-header">
        <div class="card-header">
          <span>ABAC 策略列表</span>
          <el-button type="primary" size="small" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            <span v-if="!isMobile">新增策略</span>
          </el-button>
        </div>

        <!-- 搜索栏 -->
        <el-form :model="queryParams" :inline="!isMobile" class="search-form">
          <el-form-item label="资源类型">
            <el-select v-model="queryParams.resource_type" placeholder="全部" clearable style="width: 160px">
              <el-option
                v-for="res in resourceList"
                :key="res.type"
                :label="res.display_name"
                :value="res.type"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="主体类型">
            <el-select v-model="queryParams.subject_type" placeholder="全部" clearable style="width: 160px">
              <el-option label="角色" value="ROLE" />
              <el-option label="权限集" value="PERMISSION_SET" />
              <el-option label="用户" value="USER" />
              <el-option label="部门" value="DEPT" />
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
        <ResponsiveList ref="listRef" :data="tableData" :loading="loading">
          <!-- 桌面端表格 -->
          <template #table>
            <el-table-column prop="name" label="策略名称" min-width="160" show-overflow-tooltip />
            <el-table-column prop="resource_type" label="资源类型" width="120">
              <template #default="{ row }">
                <span>{{ getResourceDisplayName(row.resource_type) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="主体" min-width="160">
              <template #default="{ row }">
                <el-tag size="small" type="info">{{ subjectTypeLabel(row.subject_type) }}</el-tag>
                <span class="ml-1">{{ row.subject_display_name || row.subject_id }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="effect" label="效果" width="80">
              <template #default="{ row }">
                <el-tag :type="row.effect === 'ALLOW' ? 'success' : 'danger'" size="small">
                  {{ row.effect === 'ALLOW' ? '允许' : '拒绝' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="priority" label="优先级" width="80" align="center" />
            <el-table-column prop="status" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
                  {{ row.status === 1 ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="160" />
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
              <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small" class="card-tag">
                {{ row.status === 1 ? '启用' : '禁用' }}
              </el-tag>
            </div>
            <div class="card-row">
              <span class="card-label">资源类型</span>
              <span class="card-value">{{ getResourceDisplayName(row.resource_type) }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">主体</span>
              <span class="card-value">
                <el-tag size="small" type="info">{{ subjectTypeLabel(row.subject_type) }}</el-tag>
                <span class="ml-1">{{ row.subject_display_name || row.subject_id }}</span>
              </span>
            </div>
            <div class="card-row">
              <span class="card-label">效果</span>
              <span class="card-value">
                <el-tag :type="row.effect === 'ALLOW' ? 'success' : 'danger'" size="small">
                  {{ row.effect === 'ALLOW' ? '允许' : '拒绝' }}
                </el-tag>
              </span>
            </div>
            <div class="card-row">
              <span class="card-label">优先级</span>
              <span class="card-value">{{ row.priority }}</span>
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
    <PolicyForm ref="formRef" :resource-list="resourceList" :subject-attrs="subjectAttrs" @success="fetchData" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  listAbacPolicies,
  deleteAbacPolicy,
  listAbacResources,
  type AbacPolicyItem,
  type ResourceVO,
  type SubjectAttrVO
} from '@/api/abac'
import PolicyForm from './PolicyForm.vue'
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
const tableData = ref<AbacPolicyItem[]>([])
const total = ref(0)
const formRef = ref()
const resourceList = ref<ResourceVO[]>([])
const subjectAttrs = ref<SubjectAttrVO[]>([])

const queryParams = reactive({
  resource_type: '',
  subject_type: '',
  page: 1,
  page_size: 20
})

// 获取资源显示名
function getResourceDisplayName(type: string): string {
  const res = resourceList.value.find(r => r.type === type)
  return res?.display_name || type
}

// 主体类型显示
function subjectTypeLabel(type: string): string {
  const map: Record<string, string> = {
    ROLE: '角色', PERMISSION_SET: '权限集', USER: '用户', DEPT: '部门'
  }
  return map[type] || type
}

async function fetchData() {
  loading.value = true
  try {
    const params: any = { page: queryParams.page, page_size: queryParams.page_size }
    if (queryParams.resource_type) params.resource_type = queryParams.resource_type
    if (queryParams.subject_type) params.subject_type = queryParams.subject_type
    const res: any = await listAbacPolicies(params)
    tableData.value = res.list || res.data?.list || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

async function loadResources() {
  try {
    const res = await listAbacResources()
    resourceList.value = res.resources || []
    subjectAttrs.value = res.subject_attrs || []
  } catch { /* 静默 */ }
}

function handleSearch() {
  queryParams.page = 1
  fetchData()
}

function handleReset() {
  queryParams.resource_type = ''
  queryParams.subject_type = ''
  handleSearch()
}

function handleAdd() {
  formRef.value?.open()
}

function handleEdit(row: AbacPolicyItem) {
  formRef.value?.open(row)
}

async function handleDelete(row: AbacPolicyItem) {
  await ElMessageBox.confirm(`确认删除策略「${row.name}」？`, '提示', { type: 'warning' })
  await deleteAbacPolicy(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

onMounted(() => {
  loadResources()
  fetchData()
})
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
.ml-1 {
  margin-left: 4px;
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

  .search-form :deep(.el-select) {
    width: 100% !important;
  }

  .pagination {
    justify-content: center;
    padding-bottom: 12px;
  }
}
</style>
