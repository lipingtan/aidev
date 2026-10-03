<template>
  <div class="plugin-container" :class="{ 'mobile-layout': isMobile }">
    <el-card class="main-card">
      <!-- 固定头部区域（手机端） -->
      <div class="sticky-header">
        <div class="card-header">
          <span>插件列表（只读）</span>
        </div>

        <!-- 提示信息 -->
        <el-alert
          title="租户插件分配功能将在多租户管理完善后启用"
          type="info"
          show-icon
          :closable="false"
          class="header-alert"
        />
      </div>

      <!-- 可滚动的列表区域 -->
      <div class="scrollable-content">
        <!-- 响应式列表 -->
        <ResponsiveList ref="listRef" :data="tableData" :loading="loading" row-key="name">
          <!-- 桌面端表格 -->
          <template #table>
            <el-table-column prop="name" label="插件名称" min-width="150" />
            <el-table-column prop="displayName" label="显示名" min-width="180">
              <template #default="{ row }">
                {{ row.displayName || row.description || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="version" label="版本" width="120" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</el-tag>
              </template>
            </el-table-column>
          </template>

          <!-- 手机端卡片 -->
          <template #card="{ row }">
            <div class="card-title">
              {{ row.name }}
              <el-tag :type="statusTagType(row.status)" size="small" class="card-tag">
                {{ statusLabel(row.status) }}
              </el-tag>
            </div>
            <div class="card-row">
              <span class="card-label">显示名</span>
              <span class="card-value">{{ row.displayName || row.description || '-' }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">版本</span>
              <span class="card-value">{{ row.version }}</span>
            </div>
          </template>
        </ResponsiveList>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { listAllPlugins } from '@/api/plugin'
import ResponsiveList from '@/components/ResponsiveList.vue'

// 插件数据项接口
interface PluginItem {
  name: string
  displayName?: string
  description?: string
  version: string
  status: number
}

// 响应式检测
const isMobile = ref(false)
const MOBILE_BREAKPOINT = 768
const listRef = ref()

function checkMobile() {
  isMobile.value = window.innerWidth <= MOBILE_BREAKPOINT
}

const loading = ref(false)
const tableData = ref<PluginItem[]>([])

// 状态标签类型
function statusTagType(status: number) {
  const map: Record<number, string> = { 0: '', 1: 'success', 2: 'info', 3: 'danger' }
  return map[status] || ''
}

// 状态文字
function statusLabel(status: number) {
  const map: Record<number, string> = { 0: '已安装', 1: '运行中', 2: '已停止', 3: '异常' }
  return map[status] || '未知'
}

// 获取插件列表
async function fetchData() {
  loading.value = true
  try {
    const res: any = await listAllPlugins()
    tableData.value = Array.isArray(res) ? res : (res?.data || res?.list || [])
  } finally {
    loading.value = false
  }
}

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
.plugin-container {
  padding: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color-light);
}

.header-alert {
  margin: 12px 16px;
  width: auto;
}

.scrollable-content {
  padding: 0 16px 16px;
}

/* 手机端布局 - 固定头部 */
@media (max-width: 768px) {
  .plugin-container {
    padding: 0;
    height: 100vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .plugin-container .main-card {
    flex: 1;
    display: flex;
    flex-direction: column;
    border-radius: 0;
    overflow: hidden;
  }

  .plugin-container .main-card :deep(.el-card__body) {
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

  .header-alert {
    margin: 12px;
  }

  .scrollable-content {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
    -webkit-overflow-scrolling: touch;
  }
}
</style>
