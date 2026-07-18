<template>
  <div class="page-container">
    <!-- 搜索栏 -->
    <el-card shadow="never" class="search-card">
      <el-form :model="queryParams" inline>
        <el-form-item label="部门名称">
          <el-input v-model="queryParams.deptName" placeholder="请输入部门名称" clearable @keyup.enter="handleSearch" />
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
          <span>部门列表</span>
          <el-button type="primary" @click="handleAdd">新增部门</el-button>
        </div>
      </template>

      <el-table v-loading="loading" :data="tableData" row-key="deptId" :tree-props="{ children: 'children' }" default-expand-all>
        <el-table-column prop="deptName" label="部门名称" min-width="200" />
        <el-table-column prop="leader" label="负责人" width="120" />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 0 ? 'success' : 'danger'" size="small">
              {{ row.status === 0 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <DeptForm ref="formRef" @success="fetchData" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listDepts, deleteDept } from '@/api/dept'
import type { DeptItem } from '@/api/dept'
import DeptForm from './DeptForm.vue'

const loading = ref(false)
const tableData = ref<DeptItem[]>([])
const formRef = ref()

const queryParams = reactive({
  deptName: ''
})

async function fetchData() {
  loading.value = true
  try {
    const res: any = await listDepts(queryParams.deptName ? { deptName: queryParams.deptName } : undefined)
    tableData.value = Array.isArray(res) ? res : (res?.list || [])
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  fetchData()
}

function handleReset() {
  queryParams.deptName = ''
  fetchData()
}

function handleAdd() {
  formRef.value?.open()
}

function handleEdit(row: DeptItem) {
  formRef.value?.open(row)
}

async function handleDelete(row: DeptItem) {
  await ElMessageBox.confirm(`确认删除部门「${row.deptName}」？`, '提示', { type: 'warning' })
  await deleteDept(row.deptId)
  ElMessage.success('删除成功')
  fetchData()
}

onMounted(fetchData)
</script>

<style scoped>
.page-container { padding: 16px; }
.search-card { margin-bottom: 16px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
</style>
