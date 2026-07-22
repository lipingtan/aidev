<template>
  <div class="app-catalog-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>应用目录</span>
          <el-tag type="info" size="small">共 {{ catalogList.length }} 个应用</el-tag>
        </div>
      </template>

      <!-- 应用卡片网格 -->
      <el-row :gutter="16">
        <el-col
          v-for="app in catalogList"
          :key="app.id"
          :xs="24"
          :sm="12"
          :md="8"
          :lg="6"
        >
          <el-card class="app-card" shadow="hover">
            <!-- 图标 + 名称 -->
            <div class="app-card__header">
              <el-icon class="app-card__icon" :size="36">
                <component :is="app.icon || 'Box'" />
              </el-icon>
              <div class="app-card__title">
                <span class="app-name">{{ app.name }}</span>
                <!-- 状态标签 -->
                <div class="app-card__tags">
                  <el-tag v-if="app.subscribed" type="success" size="small">已开通</el-tag>
                  <el-tag
                    v-if="app.app_type === 'PLUGIN' && app.plugin_status === 'STOPPED'"
                    type="warning"
                    size="small"
                  >
                    维护中
                  </el-tag>
                  <el-tag
                    v-if="app.app_type === 'PLUGIN' && app.plugin_status === 'ERROR'"
                    type="danger"
                    size="small"
                  >
                    异常
                  </el-tag>
                </div>
              </div>
            </div>

            <!-- 描述 -->
            <p class="app-card__desc">{{ app.description || '暂无描述' }}</p>

            <!-- 操作按钮 -->
            <div class="app-card__actions">
              <template v-if="app.subscribed">
                <el-button size="small" type="primary" @click="openModuleDrawer(app)">
                  模块配置
                </el-button>
                <!-- BUILTIN 类型不展示退订按钮 -->
                <el-button
                  v-if="app.app_type !== 'BUILTIN'"
                  size="small"
                  type="danger"
                  plain
                  @click="handleUnsubscribe(app)"
                >
                  退订
                </el-button>
              </template>
              <template v-else>
                <el-button size="small" type="success" @click="handleSubscribe(app)">
                  申请开通
                </el-button>
              </template>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <!-- 模块配置抽屉 -->
    <el-drawer
      v-model="drawerVisible"
      :title="`${currentApp?.name ?? ''} - 模块配置`"
      size="400px"
    >
      <div v-if="currentApp" class="module-config">
        <el-form label-position="left" label-width="auto">
          <el-form-item
            v-for="mod in currentApp.modules"
            :key="mod.code"
            :label="mod.name"
          >
            <el-switch
              v-model="moduleStates[mod.code]"
            />
          </el-form-item>
        </el-form>
        <div class="module-config__footer">
          <el-button type="primary" :loading="saving" @click="saveModules">
            保存
          </el-button>
          <el-button @click="drawerVisible = false">取消</el-button>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getAppCatalog,
  subscribeApp,
  unsubscribeApp,
  updateModules,
  type AppCatalogItem
} from '@/api/app-catalog'

/** 应用列表 */
const catalogList = ref<AppCatalogItem[]>([])
const loading = ref(false)

/** 抽屉相关 */
const drawerVisible = ref(false)
const currentApp = ref<AppCatalogItem | null>(null)
const moduleStates = reactive<Record<string, boolean>>({})
const saving = ref(false)

/** 加载应用目录 */
async function loadCatalog() {
  loading.value = true
  try {
    const res: any = await getAppCatalog()
    catalogList.value = Array.isArray(res) ? res : (res?.data || [])
  } catch (e: any) {
    ElMessage.error('加载应用目录失败')
  } finally {
    loading.value = false
  }
}

/** 申请开通 */
async function handleSubscribe(app: AppCatalogItem) {
  try {
    await subscribeApp(app.app_code)
    ElMessage.success(`已开通「${app.name}」`)
    app.subscribed = true
  } catch (e: any) {
    // 配额超限提示
    const msg = e?.response?.data?.message || e?.message || ''
    if (msg.includes('配额') || msg.includes('quota') || e?.response?.status === 429) {
      ElMessageBox.alert('当前租户应用订阅已达配额上限，请联系管理员扩容。', '配额超限', {
        type: 'warning'
      })
    } else {
      ElMessage.error(msg || '开通失败')
    }
  }
}

/** 退订 */
async function handleUnsubscribe(app: AppCatalogItem) {
  try {
    await ElMessageBox.confirm(
      '退订后将清理相关权限配置，确定退订？',
      '退订确认',
      { type: 'warning', confirmButtonText: '确定退订', cancelButtonText: '取消' }
    )
    await unsubscribeApp(app.app_code)
    ElMessage.success(`已退订「${app.name}」`)
    app.subscribed = false
  } catch (e: any) {
    // 用户取消不提示
    if (e === 'cancel' || e?.toString?.().includes('cancel')) return
    ElMessage.error('退订失败')
  }
}

/** 打开模块配置抽屉 */
function openModuleDrawer(app: AppCatalogItem) {
  currentApp.value = app
  // 初始化模块开关状态（默认全部开启）
  Object.keys(moduleStates).forEach((k) => delete moduleStates[k])
  for (const mod of app.modules) {
    moduleStates[mod.code] = true
  }
  drawerVisible.value = true
}

/** 保存模块配置 */
async function saveModules() {
  if (!currentApp.value) return
  saving.value = true
  try {
    const enabledModules = Object.entries(moduleStates)
      .filter(([, v]) => v)
      .map(([k]) => k)
    await updateModules(currentApp.value.app_code, enabledModules)
    ElMessage.success('模块配置已保存')
    drawerVisible.value = false
  } catch (e: any) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadCatalog()
})
</script>

<style scoped>
.app-catalog-page {
  padding: 16px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.app-card {
  margin-bottom: 16px;
  min-height: 200px;
  display: flex;
  flex-direction: column;
}

.app-card__header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 8px;
}

.app-card__icon {
  flex-shrink: 0;
  color: var(--el-color-primary);
}

.app-card__title {
  flex: 1;
  min-width: 0;
}

.app-name {
  font-size: 15px;
  font-weight: 600;
  display: block;
  margin-bottom: 4px;
}

.app-card__tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.app-card__desc {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  line-height: 1.5;
  margin: 8px 0 12px;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.app-card__actions {
  display: flex;
  gap: 8px;
}

.module-config {
  padding: 0 16px;
}

.module-config__footer {
  margin-top: 24px;
  display: flex;
  gap: 8px;
}
</style>
