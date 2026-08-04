<template>
  <div class="page-container">
    <!-- 搜索栏 -->
    <el-card shadow="never" class="search-card">
      <el-form :model="queryParams" inline>
        <el-form-item label="资源类型">
          <el-select v-model="queryParams.resource_type" placeholder="全部" clearable>
            <el-option
              v-for="res in resourceList"
              :key="res.type"
              :label="res.display_name"
              :value="res.type"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="主体类型">
          <el-select v-model="queryParams.subject_type" placeholder="全部" clearable>
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
    </el-card>

    <!-- 表格 -->
    <el-card shadow="never" class="table-card">
      <template #header>
        <div class="card-header">
          <span>ABAC 策略列表</span>
          <el-button type="primary" @click="handleAdd">新增策略</el-button>
        </div>
      </template>

      <div class="table-scroll-wrapper">
        <el-table v-loading="loading" :data="tableData" row-key="id" style="min-width: 800px;">
          <template #empty>
            <el-empty description="暂无策略" />
          </template>
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
        </el-table>
      </div>

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
    <PolicyForm ref="formRef" :resource-list="resourceList" :subject-attrs="subjectAttrs" @success="fetchData" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  listAbacPolicies,
  deleteAbacPolicy,
  listAbacResources,
  type AbacPolicyItem,
  type ResourceVO,
  type SubjectAttrVO
} from '@/api/abac'
import PolicyForm from './PolicyForm.vue'

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
.page-container { padding: 16px; }
.search-card { margin-bottom: 16px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.pagination { margin-top: 16px; justify-content: flex-end; }
.table-scroll-wrapper { overflow-x: auto; -webkit-overflow-scrolling: touch; }
.ml-1 { margin-left: 4px; }
</style>
