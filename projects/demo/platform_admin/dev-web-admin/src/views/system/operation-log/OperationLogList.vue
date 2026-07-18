<template>
  <div class="page-container">
    <!-- 搜索区域 -->
    <el-card shadow="never" class="search-card">
      <el-form :model="queryParams" inline>
        <el-form-item label="模块">
          <el-input v-model="queryParams.module" placeholder="请输入模块名" clearable />
        </el-form-item>
        <el-form-item label="操作人">
          <el-input v-model="queryParams.user_id" placeholder="用户 ID" clearable />
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="timeRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="YYYY-MM-DD HH:mm:ss"
          />
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
        <span>操作日志</span>
      </template>

      <el-table v-loading="loading" :data="tableData" row-key="id">
        <template #empty>
          <el-empty description="暂无数据" />
        </template>
        <el-table-column prop="module" label="模块" width="120" />
        <el-table-column prop="action" label="操作" width="120" />
        <el-table-column prop="summary" label="摘要" width="200" show-overflow-tooltip />
        <el-table-column prop="target_type" label="目标类型" width="120" />
        <el-table-column prop="target_id" label="目标 ID" width="160" show-overflow-tooltip />
        <el-table-column prop="client_ip" label="IP" width="140" />
        <el-table-column prop="created_at" label="操作时间" width="180" />
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDetail(row)">详情</el-button>
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

    <!-- 详情弹窗（JSON diff） -->
    <el-dialog v-model="detailVisible" title="操作日志详情" width="700px" destroy-on-close>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="模块">{{ detailRow?.module }}</el-descriptions-item>
        <el-descriptions-item label="操作">{{ detailRow?.action }}</el-descriptions-item>
        <el-descriptions-item label="操作人">{{ detailRow?.user_id }}</el-descriptions-item>
        <el-descriptions-item label="操作时间">{{ detailRow?.created_at }}</el-descriptions-item>
        <el-descriptions-item label="目标类型">{{ detailRow?.target_type }}</el-descriptions-item>
        <el-descriptions-item label="目标 ID">{{ detailRow?.target_id }}</el-descriptions-item>
        <el-descriptions-item label="IP">{{ detailRow?.client_ip }}</el-descriptions-item>
      </el-descriptions>

      <div class="diff-container">
        <div v-if="diffHtml" class="diff-html" v-html="diffHtml"></div>
        <div v-else class="no-diff-tip">无变更数据</div>
      </div>

      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { listOperationLogs } from '@/api/operation-log'
import type { OperationLogItem } from '@/api/operation-log'
import { formatDiffHtml } from '@/utils/json-diff'

const loading = ref(false)
const tableData = ref<OperationLogItem[]>([])
const total = ref(0)
const timeRange = ref<[string, string] | null>(null)

const queryParams = reactive({
  module: '',
  action: '',
  user_id: '',
  start_time: '',
  end_time: '',
  page: 1,
  page_size: 20
})

async function fetchData() {
  // 同步时间范围到查询参数
  if (timeRange.value && timeRange.value.length === 2) {
    queryParams.start_time = timeRange.value[0]
    queryParams.end_time = timeRange.value[1]
  } else {
    queryParams.start_time = ''
    queryParams.end_time = ''
  }

  loading.value = true
  try {
    const res: any = await listOperationLogs(queryParams)
    tableData.value = res.data || res.list || []
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
  queryParams.module = ''
  queryParams.action = ''
  queryParams.user_id = ''
  timeRange.value = null
  handleSearch()
}

// ==================== 详情弹窗 ====================
const detailVisible = ref(false)
const detailRow = ref<OperationLogItem | null>(null)

function handleDetail(row: OperationLogItem) {
  detailRow.value = row
  detailVisible.value = true
}

const diffHtml = computed(() => {
  if (!detailRow.value) return ''
  return formatDiffHtml(detailRow.value.old_value, detailRow.value.new_value)
})

onMounted(fetchData)
</script>

<style scoped>
.page-container { padding: 16px; }
.search-card { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }
.diff-container {
  margin-top: 16px;
  max-height: 400px;
  overflow: auto;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  padding: 12px;
}
</style>
