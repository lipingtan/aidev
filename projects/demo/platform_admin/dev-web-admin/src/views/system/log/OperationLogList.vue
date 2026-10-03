<template>
  <div class="page-container" :class="{ 'mobile-layout': isMobile }">
    <el-card class="main-card">
      <!-- 固定头部区域（手机端） -->
      <div class="sticky-header">
        <div class="card-header">
          <span>操作日志</span>
        </div>

        <!-- 搜索栏 -->
        <el-form :inline="!isMobile" :model="queryParams" class="search-form">
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
      </div>

      <!-- 可滚动的列表区域 -->
      <div class="scrollable-content">
        <!-- 响应式列表 -->
        <ResponsiveList ref="listRef" :data="tableData" :loading="loading">
          <!-- 桌面端表格 -->
          <template #table>
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
          </template>

          <!-- 手机端卡片 -->
          <template #card="{ row }">
            <div class="card-title">
              {{ row.module }} / {{ row.action }}
            </div>
            <div class="card-row">
              <span class="card-label">摘要</span>
              <span class="card-value">{{ row.summary || '-' }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">目标类型</span>
              <span class="card-value">{{ row.target_type || '-' }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">目标 ID</span>
              <span class="card-value">{{ row.target_id || '-' }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">IP</span>
              <span class="card-value">{{ row.client_ip || '-' }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">操作时间</span>
              <span class="card-value">{{ row.created_at }}</span>
            </div>
          </template>

          <!-- 手机端操作按钮 -->
          <template #actions="{ row }">
            <el-button size="small" type="primary" @click="handleDetail(row)">详情</el-button>
          </template>
        </ResponsiveList>

        <!-- 分页 -->
        <el-pagination
          class="pagination"
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.page_size"
          :total="total"
          :page-sizes="[10, 20, 50]"
          :layout="isMobile ? 'total, prev, pager, next' : 'total, sizes, prev, pager, next, jumper'"
          :small="isMobile"
          @size-change="fetchData"
          @current-change="fetchData"
        />
      </div>
    </el-card>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="操作日志详情" :width="isMobile ? '95%' : '700px'" destroy-on-close>
      <el-descriptions :column="isMobile ? 1 : 2" border size="small">
        <el-descriptions-item label="模块">{{ detailRow?.module }}</el-descriptions-item>
        <el-descriptions-item label="操作">{{ detailRow?.action }}</el-descriptions-item>
        <el-descriptions-item label="摘要" :span="isMobile ? 1 : 2">{{ detailRow?.summary }}</el-descriptions-item>
        <el-descriptions-item label="目标类型">{{ detailRow?.target_type || '-' }}</el-descriptions-item>
        <el-descriptions-item label="目标 ID">{{ detailRow?.target_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="操作人 ID">{{ detailRow?.user_id }}</el-descriptions-item>
        <el-descriptions-item label="IP">{{ detailRow?.client_ip }}</el-descriptions-item>
        <el-descriptions-item label="操作时间" :span="isMobile ? 1 : 2">{{ detailRow?.created_at }}</el-descriptions-item>
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
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { listOperationLogs } from '@/api/log'
import ResponsiveList from '@/components/ResponsiveList.vue'

// 响应式检测
const isMobile = ref(false)
const MOBILE_BREAKPOINT = 768
const listRef = ref()

function checkMobile() {
  isMobile.value = window.innerWidth <= MOBILE_BREAKPOINT
}

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

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  fetchData()
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
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
