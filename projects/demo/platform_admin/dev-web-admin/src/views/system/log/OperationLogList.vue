<template>
  <div class="page-container">
    <el-card shadow="never" class="search-card">
      <el-form :model="queryParams" inline>
        <el-form-item label="模块">
          <el-input v-model="queryParams.module" placeholder="请输入模块" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="操作">
          <el-input v-model="queryParams.action" placeholder="请输入操作" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <template #header><span>操作日志</span></template>
      <el-table v-loading="loading" :data="tableData">
        <el-table-column prop="module" label="模块" width="100" />
        <el-table-column prop="action" label="操作" width="100" />
        <el-table-column prop="target_type" label="目标类型" width="120" />
        <el-table-column prop="target_id" label="目标 ID" width="160" show-overflow-tooltip />
        <el-table-column prop="summary" label="摘要" min-width="180" show-overflow-tooltip />
        <el-table-column prop="client_ip" label="IP" width="130" />
        <el-table-column prop="created_at" label="操作时间" width="180" />
      </el-table>
      <el-pagination class="pagination" v-model:current-page="queryParams.page" v-model:page-size="queryParams.page_size"
        :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchData" @current-change="fetchData" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { listOperationLogs } from '@/api/log'

const loading = ref(false)
const tableData = ref<any[]>([])
const total = ref(0)
const queryParams = reactive({ module: '', action: '', page: 1, page_size: 10 })

async function fetchData() {
  loading.value = true
  try {
    const res: any = await listOperationLogs(queryParams)
    tableData.value = res.data || res.list || []
    total.value = res.total || 0
  } finally { loading.value = false }
}

function handleSearch() { queryParams.page = 1; fetchData() }
function handleReset() { queryParams.module = ''; queryParams.action = ''; handleSearch() }

onMounted(fetchData)
</script>

<style scoped>
.page-container { padding: 16px; }
.search-card { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
