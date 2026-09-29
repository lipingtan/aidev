<template>
  <div class="page-container">
    <el-alert
      v-if="error"
      :title="'数据获取失败：' + error"
      type="error"
      show-icon
      :closable="false"
      style="margin-bottom: 16px"
    />
    <el-alert
      v-if="stale"
      title="数据已过期，正在重试..."
      type="warning"
      show-icon
      :closable="false"
      style="margin-bottom: 16px"
    />

    <el-row :gutter="16">
      <!-- CPU 信息 -->
      <el-col :span="8">
        <el-card shadow="never" :body-style="{ padding: '16px' }">
          <template #header><span>CPU 信息</span></template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="核心数">{{ data.cpu.cores || '-' }}</el-descriptions-item>
            <el-descriptions-item label="使用率">
              <template v-if="data.cpu.usagePercent !== null && data.cpu.usagePercent !== undefined">
                <el-progress :percentage="data.cpu.usagePercent" :stroke-width="16" :text-inside="true" />
              </template>
              <template v-else>-</template>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>

      <!-- 内存信息 -->
      <el-col :span="8">
        <el-card shadow="never" :body-style="{ padding: '16px' }">
          <template #header><span>内存信息</span></template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="总内存">{{ formatBytes(data.mem.total) }}</el-descriptions-item>
            <el-descriptions-item label="已用内存">{{ formatBytes(data.mem.used) }}</el-descriptions-item>
            <el-descriptions-item label="使用率">
              <template v-if="data.mem.usagePercent !== null && data.mem.usagePercent !== undefined">
                <el-progress :percentage="data.mem.usagePercent" :stroke-width="16" :text-inside="true" />
              </template>
              <template v-else>-</template>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>

      <!-- 磁盘信息 -->
      <el-col :span="8">
        <el-card shadow="never" :body-style="{ padding: '16px' }">
          <template #header><span>磁盘信息</span></template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="总空间">{{ formatBytes(data.disk.total) }}</el-descriptions-item>
            <el-descriptions-item label="已用空间">{{ formatBytes(data.disk.used) }}</el-descriptions-item>
            <el-descriptions-item label="使用率">
              <template v-if="data.disk.usagePercent !== null && data.disk.usagePercent !== undefined">
                <el-progress :percentage="data.disk.usagePercent" :stroke-width="16" :text-inside="true" />
              </template>
              <template v-else>-</template>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" style="margin-top: 16px">
      <!-- 服务器信息 -->
      <el-col :span="12">
        <el-card shadow="never" :body-style="{ padding: '16px' }">
          <template #header><span>服务器信息</span></template>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="服务器名称">{{ data.os.hostName || '-' }}</el-descriptions-item>
            <el-descriptions-item label="操作系统">{{ data.os.os || '-' }}</el-descriptions-item>
            <el-descriptions-item label="服务器IP">{{ data.os.ip || '-' }}</el-descriptions-item>
            <el-descriptions-item label="运行时长">{{ formatUptime(data.process.uptimeSecs) }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>

      <!-- 进程信息 -->
      <el-col :span="12">
        <el-card shadow="never" :body-style="{ padding: '16px' }">
          <template #header><span>进程信息</span></template>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="Goroutine 数">{{ data.process.goroutines || '-' }}</el-descriptions-item>
            <el-descriptions-item label="最后更新">{{ lastUpdated || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted, onUnmounted } from 'vue'
import { getServerMonitor } from '@/api/monitor'

const data = reactive({
  cpu: { cores: 0, usagePercent: null as number | null },
  mem: { total: 0, used: 0, usagePercent: null as number | null },
  os: { hostName: '', os: '', ip: '' },
  disk: { total: 0, used: 0, usagePercent: null as number | null },
  process: { goroutines: 0, uptimeSecs: 0 }
})

const error = ref('')
const stale = ref(false)
const lastUpdated = ref('')
let timer: number | null = null

function formatBytes(bytes: number): string {
  if (!bytes) return '-'
  const gb = bytes / (1024 * 1024 * 1024)
  return gb >= 1 ? `${gb.toFixed(2)} GB` : `${(bytes / (1024 * 1024)).toFixed(2)} MB`
}

function formatUptime(seconds: number): string {
  if (!seconds) return '-'
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (days > 0) return `${days}天 ${hours}小时`
  if (hours > 0) return `${hours}小时 ${minutes}分钟`
  return `${minutes}分钟`
}

async function fetchData() {
  try {
    const res: any = await getServerMonitor()
    if (res) {
      data.cpu.cores = res.cpu?.cores ?? 0
      data.cpu.usagePercent = res.cpu?.usagePercent ?? null
      data.mem.total = res.mem?.total ?? 0
      data.mem.used = res.mem?.used ?? 0
      data.mem.usagePercent = res.mem?.usagePercent ?? null
      data.os.hostName = res.os?.hostName ?? ''
      data.os.os = res.os?.os ?? ''
      data.os.ip = res.os?.ip ?? ''
      data.disk.total = res.disk?.total ?? 0
      data.disk.used = res.disk?.used ?? 0
      data.disk.usagePercent = res.disk?.usagePercent ?? null
      data.process.goroutines = res.process?.goroutines ?? 0
      data.process.uptimeSecs = res.process?.uptimeSecs ?? 0
      lastUpdated.value = new Date().toLocaleTimeString()
      error.value = ''
      stale.value = false
    }
  } catch (e: any) {
    error.value = e.message || '请求失败'
    stale.value = true
  }
}

onMounted(() => {
  fetchData()
  timer = window.setInterval(fetchData, 5000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.page-container { padding: 16px; }
</style>
