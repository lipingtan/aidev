<template>
  <div class="plugin-container" :class="{ 'mobile-layout': isMobile }">
    <el-card class="main-card">
      <!-- 固定头部区域（手机端） -->
      <div class="sticky-header">
        <div class="card-header">
          <span>插件管理</span>
          <div class="header-actions">
            <el-upload
              :auto-upload="false"
              :show-file-list="false"
              accept=".zip,.tar.gz"
              :on-change="handleInstallFile"
            >
              <el-button type="primary" size="small" :icon="Upload">
                <span v-if="!isMobile">安装插件</span>
              </el-button>
            </el-upload>
            <el-button size="small" :icon="Refresh" @click="fetchList">
              <span v-if="!isMobile">刷新</span>
            </el-button>
          </div>
        </div>
      </div>

      <!-- 可滚动的列表区域 -->
      <div class="scrollable-content">
        <!-- 响应式列表 -->
        <ResponsiveList ref="listRef" :data="pluginList" :loading="loading" row-key="name" :table-props="{ border: true, stripe: true }">
          <!-- 桌面端表格 -->
          <template #table>
            <el-table-column prop="name" label="名称" min-width="120" />
            <el-table-column prop="version" label="版本" width="100" />
            <el-table-column prop="description" label="描述" min-width="200" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.runStatus)">
                  {{ statusLabel(row.runStatus) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button v-if="row.runStatus === 'stopped' || row.status === 0" size="small" type="success" @click="handleStart(row)">启动</el-button>
                <el-button v-if="row.runStatus === 'running'" size="small" type="warning" @click="handleStop(row)">停止</el-button>
                <el-button v-if="row.runStatus === 'running' || row.runStatus === 'stopped'" size="small" type="primary" @click="openUpgradeDialog(row)">升级</el-button>
                <el-button v-if="row.runStatus === 'stopped' || row.status === 0" size="small" type="danger" @click="handleUninstall(row)">卸载</el-button>
              </template>
            </el-table-column>
          </template>

          <!-- 手机端卡片 -->
          <template #card="{ row }">
            <div class="card-title">
              {{ row.name }}
              <el-tag :type="statusTagType(row.runStatus)" size="small" class="card-tag">
                {{ statusLabel(row.runStatus) }}
              </el-tag>
            </div>
            <div class="card-row">
              <span class="card-label">版本</span>
              <span class="card-value">{{ row.version }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">描述</span>
              <span class="card-value">{{ row.description || '-' }}</span>
            </div>
          </template>

          <!-- 手机端操作按钮 -->
          <template #actions="{ row }">
            <el-button v-if="row.runStatus === 'stopped' || row.status === 0" size="small" type="success" @click="handleStart(row)">启动</el-button>
            <el-button v-if="row.runStatus === 'running'" size="small" type="warning" @click="handleStop(row)">停止</el-button>
            <el-button v-if="row.runStatus === 'running' || row.runStatus === 'stopped'" size="small" type="primary" @click="openUpgradeDialog(row)">升级</el-button>
            <el-button v-if="row.runStatus === 'stopped' || row.status === 0" size="small" type="danger" @click="handleUninstall(row)">卸载</el-button>
          </template>
        </ResponsiveList>
      </div>
    </el-card>

    <!-- 升级对话框 -->
    <el-dialog v-model="upgradeDialogVisible" title="升级插件" :width="isMobile ? '95%' : '500px'">
      <div v-if="!upgradeNeedConfirm">
        <p>请选择升级包文件：</p>
        <el-upload
          :auto-upload="false"
          :show-file-list="true"
          :limit="1"
          accept=".zip,.tar.gz"
          :on-change="handleUpgradeFileChange"
        >
          <el-button type="primary">选择文件</el-button>
        </el-upload>
      </div>
      <div v-else class="upgrade-confirm">
        <el-alert type="warning" :closable="false" show-icon>
          <template #title>升级需要确认</template>
          <p>{{ upgradeOldVersion }} → {{ upgradeNewVersion }}</p>
        </el-alert>
        <div class="migration-notes" v-if="upgradeMigrationNotes">
          <p><strong>迁移说明：</strong></p>
          <pre>{{ upgradeMigrationNotes }}</pre>
        </div>
      </div>
      <template #footer>
        <el-button @click="upgradeDialogVisible = false">取消</el-button>
        <el-button
          v-if="!upgradeNeedConfirm"
          type="primary"
          :disabled="!upgradeFile"
          :loading="upgradeLoading"
          @click="doUpgrade(false)"
        >上传升级</el-button>
        <el-button
          v-else
          type="warning"
          :loading="upgradeLoading"
          @click="doUpgrade(true)"
        >确认升级</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Upload, Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadFile } from 'element-plus'
import {
  getPluginList,
  uploadPlugin,
  startPlugin,
  stopPlugin,
  uninstallPlugin,
  upgradePlugin,
  getPluginHealth,
  type PluginItem
} from '@/api/plugin'
import ResponsiveList from '@/components/ResponsiveList.vue'

// 响应式检测
const isMobile = ref(false)
const MOBILE_BREAKPOINT = 768
const listRef = ref()

function checkMobile() {
  isMobile.value = window.innerWidth <= MOBILE_BREAKPOINT
}

const loading = ref(false)
const pluginList = ref<PluginItem[]>([])

// 升级相关状态
const upgradeDialogVisible = ref(false)
const upgradeLoading = ref(false)
const upgradeFile = ref<File | null>(null)
const upgradeTargetName = ref('')
const upgradeNeedConfirm = ref(false)
const upgradeMigrationNotes = ref('')
const upgradeOldVersion = ref('')
const upgradeNewVersion = ref('')

/** 获取插件列表 */
async function fetchList() {
  loading.value = true
  try {
    const res: any = await getPluginList()
    const data = res?.data || res
    pluginList.value = data?.list || []
  } catch {
    ElMessage.error('获取插件列表失败')
  } finally {
    loading.value = false
  }
}

/** 状态标签类型 */
function statusTagType(status: string) {
  const map: Record<string, string> = {
    running: 'success',
    stopped: 'warning',
    error: 'danger',
    unknown: 'info'
  }
  return map[status] || 'info'
}

/** 状态文本 */
function statusLabel(status: string) {
  const map: Record<string, string> = {
    running: '运行中',
    stopped: '已停止',
    error: '异常',
    unknown: '未知'
  }
  return map[status] || '已安装'
}

/** 安装插件 - 文件选择 */
async function handleInstallFile(uploadFile: UploadFile) {
  if (!uploadFile.raw) return
  try {
    await uploadPlugin(uploadFile.raw)
    ElMessage.success('插件安装成功')
    fetchList()
  } catch {
    ElMessage.error('插件安装失败')
  }
}

/** 启动插件 */
async function handleStart(row: PluginItem) {
  try {
    await startPlugin(row.name)
    ElMessage.success(`插件 ${row.name} 已启动`)
    fetchList()
  } catch {
    ElMessage.error('启动失败')
  }
}

/** 停止插件 */
async function handleStop(row: PluginItem) {
  try {
    await ElMessageBox.confirm(`确定停止插件 "${row.name}" 吗？`, '提示', { type: 'warning' })
    await stopPlugin(row.name)
    ElMessage.success(`插件 ${row.name} 已停止`)
    fetchList()
  } catch {
    // 用户取消或请求失败
  }
}

/** 卸载插件 */
async function handleUninstall(row: PluginItem) {
  try {
    await ElMessageBox.confirm(`确定卸载插件 "${row.name}" 吗？此操作不可恢复。`, '警告', { type: 'error' })
    await uninstallPlugin(row.name)
    ElMessage.success(`插件 ${row.name} 已卸载`)
    fetchList()
  } catch {
    // 用户取消或请求失败
  }
}

/** 打开升级对话框 */
function openUpgradeDialog(row: PluginItem) {
  upgradeTargetName.value = row.name
  upgradeFile.value = null
  upgradeNeedConfirm.value = false
  upgradeMigrationNotes.value = ''
  upgradeOldVersion.value = ''
  upgradeNewVersion.value = ''
  upgradeDialogVisible.value = true
}

/** 升级文件选择 */
function handleUpgradeFileChange(uploadFile: UploadFile) {
  upgradeFile.value = uploadFile.raw || null
}

/** 执行升级 */
async function doUpgrade(confirm: boolean) {
  if (!upgradeFile.value) return
  upgradeLoading.value = true
  try {
    const res: any = await upgradePlugin(upgradeTargetName.value, upgradeFile.value, confirm)
    const data = res?.data || res
    if (data?.needConfirm && !confirm) {
      // 需要二次确认
      upgradeNeedConfirm.value = true
      upgradeMigrationNotes.value = data.migrationNotes || ''
      upgradeOldVersion.value = data.oldVersion || ''
      upgradeNewVersion.value = data.newVersion || ''
    } else {
      // 升级成功，轮询健康检查
      ElMessage.success(data?.message || '升级成功')
      upgradeDialogVisible.value = false
      pollHealth(upgradeTargetName.value)
    }
  } catch {
    ElMessage.error('升级失败')
  } finally {
    upgradeLoading.value = false
  }
}

/** 升级后轮询健康检查 */
async function pollHealth(name: string, retries = 10) {
  for (let i = 0; i < retries; i++) {
    await new Promise(r => setTimeout(r, 2000))
    try {
      const res: any = await getPluginHealth(name)
      const data = res?.data || res
      if (data?.status === 'healthy' || data?.status === 'running') {
        ElMessage.success(`插件 ${name} 已恢复正常`)
        fetchList()
        return
      }
    } catch {
      // 继续轮询
    }
  }
  ElMessage.warning(`插件 ${name} 健康检查超时，请手动确认`)
  fetchList()
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  fetchList()
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

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.scrollable-content {
  padding: 0 16px 16px;
}

.upgrade-confirm {
  padding: 8px 0;
}

.migration-notes {
  margin-top: 12px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

.migration-notes pre {
  white-space: pre-wrap;
  word-break: break-all;
  margin: 8px 0 0;
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

  .scrollable-content {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
    -webkit-overflow-scrolling: touch;
  }
}
</style>
