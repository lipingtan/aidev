<template>
  <div class="page-container">
    <el-card shadow="never" class="search-card">
      <el-form :model="queryParams" inline>
        <el-form-item label="接口路径">
          <el-input v-model="queryParams.path" placeholder="请输入接口路径" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="接口标题">
          <el-input v-model="queryParams.title" placeholder="请输入接口标题" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="table-card">
      <template #header>
        <div class="card-header">
          <span>接口列表</span>
        </div>
      </template>

      <el-table v-loading="loading" :data="tableData" row-key="id">
        <el-table-column prop="path" label="接口路径" min-width="240" />
        <el-table-column prop="method" label="请求方式" width="100">
          <template #default="{ row }">
            <el-tag :type="methodTagType(row.method)" size="small">{{ row.method }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="接口标题" min-width="180" />
        <el-table-column prop="group" label="所属模块" width="140" />
      </el-table>

      <el-pagination
        class="pagination"
        v-model:current-page="queryParams.pageNum"
        v-model:page-size="queryParams.pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchData"
        @current-change="fetchData"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { listSysApis } from '@/api/sys-api'
import type { SysApiItem } from '@/api/sys-api'

const loading = ref(false)
const tableData = ref<SysApiItem[]>([])
const total = ref(0)

const queryParams = reactive({
  path: '',
  title: '',
  pageNum: 1,
  pageSize: 10
})

function methodTagType(method: string) {
  const map: Record<string, string> = { GET: '', POST: 'success', PUT: 'warning', DELETE: 'danger' }
  return map[method?.toUpperCase()] || 'info'
}

async function fetchData() {
  loading.value = true
  try {
    const { list, total: t } = await listSysApis(queryParams)
    tableData.value = list
    total.value = t
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  queryParams.pageNum = 1
  fetchData()
}

function handleReset() {
  queryParams.path = ''
  queryParams.title = ''
  handleSearch()
}

onMounted(fetchData)
</script>

<style scoped>
.page-container { padding: 16px; }
.search-card { margin-bottom: 16px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
