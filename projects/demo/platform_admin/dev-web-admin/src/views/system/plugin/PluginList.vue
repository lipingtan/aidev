<template>
  <div class="plugin-container" :class="{ 'mobile-layout': isMobile }">
    <el-card class="main-card">
      <!-- 固定头部区域（手机端） -->
      <div class="sticky-header">
        <div class="card-header">
          <span>插件列表</span>
          <el-button type="primary" size="small" @click="showInstallDialog = true">
            <el-icon><Plus /></el-icon>
            <span v-if="!isMobile">安装插件</span>
          </el-button>
        </div>

        <!-- 搜索栏 -->
        <el-form :inline="!isMobile" :model="queryParams" class="search-form">
          <el-form-item label="插件名称">
            <el-input v-model="queryParams.name" placeholder="请输入插件名称" clearable @keyup.enter="handleSearch" style="width: 180px" />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="queryParams.status" placeholder="全部" clearable style="width: 120px">
              <el-option label="已安装" :value="0" />
              <el-option label="运行中" :value="1" />
              <el-option label="已停止" :value="2" />
              <el-option label="异常" :value="3" />
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
        <ResponsiveList ref="listRef" :data="filteredData" :loading="loading" row-key="name">
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
            <el-table-column label="操作" width="140" fixed="right">
              <template #default="{ row }">
                <el-button v-if="row.status !== 1" link type="success" @click="handleEnable(row)">启用</el-button>
                <el-button v-if="row.status === 1" link type="warning" @click="handleDisable(row)">禁用</el-button>
                <el-button link type="danger" @click="handleUninstall(row)">卸载</el-button>
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

          <!-- 手机端操作按钮 -->
          <template #actions="{ row }">
            <el-button v-if="row.status !== 1" size="small" type="success" @click="handleEnable(row)">启用</el-button>
            <el-button v-if="row.status === 1" size="small" type="warning" @click="handleDisable(row)">禁用</el-button>
            <el-button size="small" type="danger" @click="handleUninstall(row)">卸载</el-button>
          </template>
        </ResponsiveList>
      </div>
    </el-card>

    <!-- 安装插件对话框 -->
    <el-dialog v-model="showInstallDialog" title="安装插件" :width="isMobile ? '95%' : '500px'">
      <el-tabs v-model="installTab">
        <el-tab-pane label="文件上传" name="file">
          <el-form label-width="80px">
            <el-form-item label="插件名称">
              <el-input v-model="installForm.name" placeholder="请输入插件标识名" />
            </el-form-item>
            <el-form-item label="插件文件">
              <el-upload
                :auto-upload="false"
                :limit="1"
                accept=".zip,.tar.gz"
                @change="(file: any) => installForm.file = file.raw"
              >
                <el-button type="primary">选择文件</el-button>
              </el-upload>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="URL安装" name="url">
          <el-form label-width="80px">
            <el-form-item label="插件名称">
              <el-input v-model="installForm.name" placeholder="请输入插件标识名" />
            </el-form-item>
            <el-form-item label="下载地址">
              <el-input v-model="installForm.url" placeholder="https://..." />
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="showInstallDialog = false">取消</el-button>
        <el-button type="primary" :loading="installing" @click="handleInstall">安装</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { listAllPlugins, enablePlugin, disablePlugin, uninstallPlugin, installPluginByUrl, installPluginByFile } from '@/api/plugin'
import ResponsiveList from '@/components/ResponsiveList.vue'

// 插件数据项接口
interface PluginItem {
  name: string
  displayName?: string
  description?: string
  version: string
  status: number // 0=已安装, 1=运行中, 2=已停止, 3=异常
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

const queryParams = ref({
  name: '',
  status: undefined as number | undefined
})

// 前端过滤
const filteredData = computed(() => {
  return tableData.value.filter(item => {
    if (queryParams.value.name && !item.name.includes(queryParams.value.name)) return false
    if (queryParams.value.status !== undefined && item.status !== queryParams.value.status) return false
    return true
  })
})

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

function handleSearch() {
  // 前端过滤，无需重新请求
}

function handleReset() {
  queryParams.value.name = ''
  queryParams.value.status = undefined
}

// 启用插件
async function handleEnable(row: PluginItem) {
  await enablePlugin(row.name)
  ElMessage.success(`插件「${row.name}」已启用`)
  fetchData()
}

// 禁用插件
async function handleDisable(row: PluginItem) {
  await disablePlugin(row.name)
  ElMessage.success(`插件「${row.name}」已禁用`)
  fetchData()
}

// 卸载插件
async function handleUninstall(row: PluginItem) {
  await ElMessageBox.confirm(`确认卸载插件「${row.name}」？此操作不可恢复。`, '提示', { type: 'warning' })
  await uninstallPlugin(row.name)
  ElMessage.success(`插件「${row.name}」已卸载`)
  fetchData()
}

// ========== 安装插件 ==========
const showInstallDialog = ref(false)
const installTab = ref('file')
const installing = ref(false)
const installForm = reactive({
  name: '',
  url: '',
  file: null as File | null
})

async function handleInstall() {
  if (!installForm.name) {
    ElMessage.warning('请输入插件名称')
    return
  }
  installing.value = true
  try {
    if (installTab.value === 'url') {
      if (!installForm.url) { ElMessage.warning('请输入下载地址'); installing.value = false; return }
      await installPluginByUrl({ name: installForm.name, url: installForm.url })
    } else {
      if (!installForm.file) { ElMessage.warning('请选择插件文件'); installing.value = false; return }
      await installPluginByFile(installForm.name, installForm.file)
    }
    ElMessage.success('插件安装成功')
    showInstallDialog.value = false
    installForm.name = ''
    installForm.url = ''
    installForm.file = null
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.message || '安装失败')
  } finally {
    installing.value = false
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

.search-form {
  padding: 12px 16px;
  margin-bottom: 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
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

  .search-form :deep(.el-select),
  .search-form :deep(.el-input) {
    width: 100% !important;
  }
}
</style>
