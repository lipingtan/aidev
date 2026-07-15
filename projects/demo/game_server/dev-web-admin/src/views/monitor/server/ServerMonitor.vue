<template>
  <div class="page-container">
    <el-row :gutter="16">
      <!-- CPU 信息 -->
      <el-col :span="8">
        <el-card shadow="never">
          <template #header><span>CPU 信息</span></template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="核心数">{{ data.cpu.cores || '-' }}</el-descriptions-item>
            <el-descriptions-item label="使用率">
              <el-progress :percentage="data.cpu.usagePercent || 0" :stroke-width="16" :text-inside="true" />
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>

      <!-- 内存信息 -->
      <el-col :span="8">
        <el-card shadow="never">
          <template #header><span>内存信息</span></template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="总内存">{{ formatBytes(data.mem.total) }}</el-descriptions-item>
            <el-descriptions-item label="已用内存">{{ formatBytes(data.mem.used) }}</el-descriptions-item>
            <el-descriptions-item label="使用率">
              <el-progress :percentage="data.mem.usagePercent || 0" :stroke-width="16" :text-inside="true" />
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>

      <!-- 服务器信息 -->
      <el-col :span="8">
        <el-card shadow="never">
          <template #header><span>服务器信息</span></template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="服务器名称">{{ data.os.hostName || '-' }}</el-descriptions-item>
            <el-descriptions-item label="操作系统">{{ data.os.os || '-' }}</el-descriptions-item>
            <el-descriptions-item label="服务器IP">{{ data.os.ip || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { getServerMonitor } from '@/api/monitor'

const data = reactive({
  cpu: { cores: 0, usagePercent: 0 },
  mem: { total: 0, used: 0, usagePercent: 0 },
  os: { hostName: '', os: '', ip: '' },
  disk: { total: 0, used: 0, usagePercent: 0 }
})

function formatBytes(bytes: number): string {
  if (!bytes) return '-'
  const gb = bytes / (1024 * 1024 * 1024)
  return gb >= 1 ? `${gb.toFixed(2)} GB` : `${(bytes / (1024 * 1024)).toFixed(2)} MB`
}

async function fetchData() {
  try {
    const res: any = await getServerMonitor()
    if (res) {
      Object.assign(data.cpu, res.cpu || {})
      Object.assign(data.mem, res.mem || {})
      Object.assign(data.os, res.os || {})
      Object.assign(data.disk, res.disk || {})
    }
  } catch {
    // 接口失败时展示空数据，不崩溃
  }
}

onMounted(fetchData)
</script>

<style scoped>
.page-container { padding: 16px; }
</style>
