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
        <el-table-column prop="action" label="操作" width="140" />
        <el-table-column prop="target_type" label="目标类型" width="100" />
        <el-table-column prop="target_id" label="目标 ID" width="160" show-overflow-tooltip />
        <el-table-column prop="summary" label="摘要" min-width="180" show-overflow-tooltip />
        <el-table-column prop="client_ip" label="IP" width="130" />
        <el-table-column prop="created_at" label="操作时间" width="180" />
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination class="pagination" v-model:current-page="queryParams.page" v-model:page-size="queryParams.page_size"
        :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchData" @current-change="fetchData" />
    </el-card>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="操作日志详情" width="700px" destroy-on-close>
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="模块">{{ detailRow?.module }}</el-descriptions-item>
        <el-descriptions-item label="操作">{{ detailRow?.action }}</el-descriptions-item>
        <el-descriptions-item label="摘要" :span="2">{{ detailRow?.summary }}</el-descriptions-item>
        <el-descriptions-item label="目标类型">{{ detailRow?.target_type || '-' }}</el-descriptions-item>
        <el-descriptions-item label="目标 ID">{{ detailRow?.target_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="操作人 ID">{{ detailRow?.user_id }}</el-descriptions-item>
        <el-descriptions-item label="IP">{{ detailRow?.client_ip }}</el-descriptions-item>
        <el-descriptions-item label="操作时间" :span="2">{{ detailRow?.created_at }}</el-descriptions-item>
      </el-descriptions>

      <el-divider content-position="left">请求入参</el-divider>
      <div class="json-block">
        <pre v-if="reqJson">{{ reqJson }}</pre>
        <span v-else class="no-data">无入参</span>
      </div>

      <el-divider content-position="left">执行结果</el-divider>
      <div class="json-block">
        <pre v-if="respJson">{{ respJson }}</pre>
        <span v-else class="no-data">无响应数据</span>
      </div>

      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { listOperationLogs } from '@/api/log'

const loading = ref(false)
const tableData = ref<any[]>([])
const total = ref(0)
const queryParams = reactive({ module: '', action: '', page: 1, page_size: 10 })

async function fetchData() {
  loading.value = true
  try {
    const res: any = await listOperationLogs(queryParams)
    const payload = res.data || res
    tableData.value = payload.list || []
    total.value = payload.total || 0
  } finally { loading.value = false }
}

function handleSearch() { queryParams.page = 1; fetchData() }
function handleReset() { queryParams.module = ''; queryParams.action = ''; handleSearch() }

// 详情弹窗
const detailVisible = ref(false)
const detailRow = ref<any>(null)

function handleDetail(row: any) {
  detailRow.value = row
  detailVisible.value = true
}

function formatJson(val: any): string {
  if (!val) return ''
  try {
    const obj = typeof val === 'string' ? JSON.parse(val) : val
    return JSON.stringify(obj, null, 2)
  } catch {
    return String(val)
  }
}

const reqJson = computed(() => formatJson(detailRow.value?.old_value))
const respJson = computed(() => formatJson(detailRow.value?.new_value))

onMounted(fetchData)
</script>

<style scoped>
.page-container { padding: 16px; }
.search-card { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
.json-block {
  background: #f5f7fa;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 12px;
  max-height: 260px;
  overflow: auto;
}
.json-block pre {
  margin: 0;
  font-size: 12px;
  font-family: 'Consolas', 'Monaco', monospace;
  white-space: pre-wrap;
  word-break: break-all;
  color: #303133;
}
.no-data { color: #909399; font-size: 13px; }
</style>
