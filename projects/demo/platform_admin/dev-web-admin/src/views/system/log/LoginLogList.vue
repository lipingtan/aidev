<template>
  <div class="page-container" :class="{ 'mobile-layout': isMobile }">
    <el-card class="main-card">
      <!-- 固定头部区域（手机端） -->
      <div class="sticky-header">
        <div class="card-header">
          <span>登录日志</span>
        </div>

        <!-- 搜索栏 -->
        <el-form :inline="!isMobile" :model="queryParams" class="search-form">
          <el-form-item label="用户名">
            <el-input v-model="queryParams.username" placeholder="请输入用户名" clearable @keyup.enter="handleSearch" />
          </el-form-item>
          <el-form-item label="登录IP">
            <el-input v-model="queryParams.ip" placeholder="请输入IP" clearable @keyup.enter="handleSearch" />
          </el-form-item>
          <el-form-item label="结果">
            <el-select v-model="queryParams.status" placeholder="全部" clearable>
              <el-option label="成功" :value="1" />
              <el-option label="失败" :value="0" />
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
        <!-- 响应式列表 -->
        <ResponsiveList ref="listRef" :data="tableData" :loading="loading">
          <!-- 桌面端表格 -->
          <template #table>
            <el-table-column prop="username" label="用户名" width="120" />
            <el-table-column prop="ip" label="登录IP" width="140" />
            <el-table-column prop="location" label="登录地点" width="140" />
            <el-table-column prop="browser" label="浏览器" width="120" />
            <el-table-column prop="os" label="操作系统" width="120" />
            <el-table-column prop="status" label="结果" width="80">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '成功' : '失败' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="message" label="消息" min-width="160" />
            <el-table-column prop="login_time" label="登录时间" width="180" />
          </template>

          <!-- 手机端卡片 -->
          <template #card="{ row }">
            <div class="card-title">
              {{ row.username }}
              <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small" class="card-tag">
                {{ row.status === 1 ? '成功' : '失败' }}
              </el-tag>
            </div>
            <div class="card-row">
              <span class="card-label">登录IP</span>
              <span class="card-value">{{ row.ip }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">登录地点</span>
              <span class="card-value">{{ row.location || '-' }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">浏览器</span>
              <span class="card-value">{{ row.browser || '-' }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">操作系统</span>
              <span class="card-value">{{ row.os || '-' }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">消息</span>
              <span class="card-value">{{ row.message || '-' }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">登录时间</span>
              <span class="card-value">{{ row.login_time }}</span>
            </div>
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { listLoginLogs } from '@/api/log'
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
const queryParams = reactive({ username: '', ip: '', status: undefined as number | undefined, page: 1, page_size: 10 })

async function fetchData() {
  loading.value = true
  try {
    const res: any = await listLoginLogs(queryParams)
    // 后端返回: { code, data: { list, total, page, page_size }, message }
    const payload = res.data || res
    tableData.value = payload.list || []
    total.value = payload.total || 0
  } finally { loading.value = false }
}

function handleSearch() { queryParams.page = 1; fetchData() }
function handleReset() { queryParams.username = ''; queryParams.ip = ''; queryParams.status = undefined; handleSearch() }

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
