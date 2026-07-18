<template>
  <div class="page-container">
    <!-- 提示信息 -->
    <el-alert
      title="租户插件分配功能将在多租户管理完善后启用"
      type="info"
      show-icon
      :closable="false"
      style="margin-bottom: 16px;"
    />

    <!-- 表格 -->
    <el-card shadow="never" class="table-card">
      <template #header>
        <div class="card-header">
          <span>插件列表（只读）</span>
        </div>
      </template>

      <el-table v-loading="loading" :data="tableData" row-key="name">
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
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listAllPlugins } from '@/api/plugin'

// 插件数据项接口
interface PluginItem {
  name: string
  displayName?: string
  description?: string
  version: string
  status: number
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

onMounted(fetchData)
</script>

<style scoped>
.page-container { padding: 16px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
</style>
