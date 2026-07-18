<!--
  租户选择页
  - 展示可用租户列表（卡片形式）
  - 点击租户卡片调用 POST /auth/tenant/select 获取 access_token
  - 成功后跳转至首页
-->
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/modules/auth'
import { ElMessage } from 'element-plus'
import type { TenantInfo } from '@/store/modules/auth'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const selectedId = ref('')
const tenantList = ref<TenantInfo[]>([])

onMounted(() => {
  tenantList.value = authStore.tenants
  // 如果 store 中无数据，尝试从 localStorage 恢复
  if (!tenantList.value.length) {
    const cached = localStorage.getItem('tenant_list')
    if (cached) {
      try {
        tenantList.value = JSON.parse(cached)
      } catch { /* ignore */ }
    }
  }
  // 仍然无数据，回退到登录页
  if (!tenantList.value.length) {
    router.replace('/login')
  }
})

/** 选择租户 */
async function handleSelect(tenant: TenantInfo) {
  if (loading.value) return
  loading.value = true
  selectedId.value = tenant.id
  try {
    await authStore.selectTenant(tenant.id)
    ElMessage.success(`已进入租户：${tenant.name}`)
    router.replace('/home')
  } catch {
    ElMessage.error('租户选择失败，请重试')
  } finally {
    loading.value = false
    selectedId.value = ''
  }
}
</script>

<template>
  <div class="tenant-select-page">
    <div class="tenant-select-container">
      <h1 class="page-title">选择租户</h1>
      <p class="page-subtitle">请选择要进入的租户空间</p>

      <div class="tenant-grid">
        <div
          v-for="tenant in tenantList"
          :key="tenant.id"
          class="tenant-card"
          :class="{ 'is-loading': selectedId === tenant.id }"
          @click="handleSelect(tenant)"
        >
          <div class="tenant-logo">
            <img v-if="tenant.logo" :src="tenant.logo" :alt="tenant.name" />
            <el-icon v-else :size="32"><OfficeBuilding /></el-icon>
          </div>
          <div class="tenant-info">
            <h3 class="tenant-name">{{ tenant.name }}</h3>
            <p v-if="tenant.description" class="tenant-desc">{{ tenant.description }}</p>
          </div>
          <el-icon v-if="selectedId === tenant.id" class="loading-icon" :size="20">
            <Loading />
          </el-icon>
        </div>
      </div>

      <div class="back-link">
        <el-button link type="primary" @click="router.replace('/login')">
          返回登录
        </el-button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { OfficeBuilding, Loading } from '@element-plus/icons-vue'
export default { components: { OfficeBuilding, Loading } }
</script>

<style scoped>
.tenant-select-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: var(--color-bg-page, #f5f7fa);
  padding: 20px;
}

.tenant-select-container {
  width: 100%;
  max-width: 640px;
  text-align: center;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--color-primary, #409eff);
  margin-bottom: 8px;
}

.page-subtitle {
  font-size: 14px;
  color: var(--color-text-secondary, #909399);
  margin-bottom: 32px;
}

.tenant-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.tenant-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: var(--color-bg-white, #fff);
  border-radius: var(--border-radius-md, 8px);
  box-shadow: var(--shadow-sm, 0 2px 8px rgba(0,0,0,0.06));
  cursor: pointer;
  transition: all 0.2s;
  border: 2px solid transparent;
  position: relative;
}

.tenant-card:hover {
  border-color: var(--color-primary, #409eff);
  box-shadow: var(--shadow-md, 0 4px 16px rgba(0,0,0,0.1));
  transform: translateY(-2px);
}

.tenant-card.is-loading {
  opacity: 0.7;
  pointer-events: none;
}

.tenant-logo {
  flex-shrink: 0;
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg-page, #f5f7fa);
  border-radius: 8px;
}

.tenant-logo img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 8px;
}

.tenant-info {
  flex: 1;
  text-align: left;
}

.tenant-name {
  font-size: 16px;
  font-weight: 500;
  color: var(--color-text-primary, #303133);
  margin: 0 0 4px 0;
}

.tenant-desc {
  font-size: 12px;
  color: var(--color-text-secondary, #909399);
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.loading-icon {
  position: absolute;
  right: 16px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.back-link {
  margin-top: 24px;
}
</style>
