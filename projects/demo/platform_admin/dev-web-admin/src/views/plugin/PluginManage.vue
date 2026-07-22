<template>
  <div class="plugin-manage">
    <!-- 顶部操作栏 -->
    <div class="toolbar">
      <el-upload
        :auto-upload="false"
        :show-file-list="false"
        accept=".zip,.tar.gz"
        :on-change="handleInstallFile"
      >
        <el-button type="primary" :icon="Upload">安装插件</el-button>
      </el-upload>
      <el-button :icon="Refresh" @click="fetchList">刷新</el-button>
    </div>

    <!-- 插件列表表格 -->
    <el-table :data="pluginList" v-loading="loading" border stripe>
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
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <!-- 已停止/已安装：启动 -->
          <el-button
            v-if="row.runStatus === 'stopped' || row.status === 0"
            size="small"
            type="success"
            @click="handleStart(row)"
          >启动</el-button>
          <!-- 运行中：停止 -->
          <el-button
            v-if="row.runStatus === 'running'"
            size="small"
            type="warning"
            @click="handleStop(row)"
          >停止</el-button>
          <!-- 运行中/已停止：升级 -->
          <el-button
            v-if="row.runStatus === 'running' || row.runStatus === 'stopped'"
            size="small"
            type="primary"
            @click="openUpgradeDialog(row)"
          >升级</el-button>
          <!-- 已停止/已安装：卸载 -->
          <el-button
            v-if="row.runStatus === 'stopped' || row.status === 0"
            size="small"
            type="danger"
            @click="handleUninstall(row)"
          >卸载</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 升级对话框 -->
    <el-dialog v-model="upgradeDialogVisible" title="升级插件" width="500px">
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
import { ref, onMounted } from 'vue'
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
  fetchList()
})
</script>

<style scoped>
.plugin-manage {
  padding: 20px;
}
.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
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
</style>
