<!--
  顶部导航栏
  - 64px 高，白色背景 + 底部边框
  - 左侧：Logo + 系统名称 + 折叠按钮
  - 右侧：当前租户名称 + 切换租户按钮 + 主题切换 + 用户头像 + 登出
-->
<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/store/modules/app'
import { useUserStore } from '@/store/modules/user'
import { useAuthStore } from '@/store/modules/auth'
import { useMenuStore } from '@/store/modules/menu'
import { ElMessageBox } from 'element-plus'
import { Fold, Expand, UserFilled, Brush, SwitchButton, Switch } from '@element-plus/icons-vue'

const router = useRouter()
const appStore = useAppStore()
const userStore = useUserStore()
const authStore = useAuthStore()
const menuStore = useMenuStore()

/** 当前租户名称 */
const currentTenantName = computed(() => {
  const tenantId = authStore.currentTenantId
  if (!tenantId) return ''
  const tenant = authStore.tenants.find(t => String(t.id) === String(tenantId))
  return tenant?.name || tenantId
})

/** 切换租户 */
async function handleSwitchTenant() {
  try {
    await ElMessageBox.confirm('切换租户将离开当前页面，是否继续？', '提示', { type: 'warning' })
    router.push('/tenant-select')
  } catch {
    // 用户取消
  }
}

/** 退出登录 */
function handleLogout() {
  menuStore.resetMenu()
  authStore.logout()
  userStore.resetState()
  router.push('/login')
}

function switchTheme(mode: 'business' | 'luxury-dark' | 'aurora-local') {
  appStore.applyTheme(mode)
}
</script>

<template>
  <el-header class="app-header" :class="{ 'aurora-local': appStore.themeMode === 'aurora-local' }">
    <!-- 左侧区域 -->
    <div class="app-header__left">
      <div class="app-header__logo">
        <span class="app-header__title">智慧物业管理平台</span>
      </div>
      <el-icon
        class="app-header__collapse-btn"
        @click="appStore.toggleSidebar()"
      >
        <Fold v-if="!appStore.sidebarCollapsed" />
        <Expand v-else />
      </el-icon>
    </div>

    <!-- 右侧区域 -->
    <div class="app-header__right">
      <!-- 当前租户 + 切换 -->
      <div v-if="currentTenantName" class="app-header__tenant">
        <span class="app-header__tenant-label">租户：</span>
        <span class="app-header__tenant-name">{{ currentTenantName }}</span>
        <el-button text size="small" class="tenant-switch-btn" @click="handleSwitchTenant">
          <el-icon><Switch /></el-icon>
          <span>切换</span>
        </el-button>
      </div>

      <!-- 主题切换 -->
      <el-dropdown trigger="click">
        <el-button class="theme-btn" text>
          <el-icon><Brush /></el-icon>
          <span>
            {{
              appStore.themeMode === 'luxury-dark'
                ? '暗夜奢华'
                : appStore.themeMode === 'aurora-local'
                  ? '极光科技'
                  : '商务经典'
            }}
          </span>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item @click="switchTheme('business')">商务经典</el-dropdown-item>
            <el-dropdown-item @click="switchTheme('luxury-dark')">暗夜奢华</el-dropdown-item>
            <el-dropdown-item @click="switchTheme('aurora-local')">极光科技（布局专属）</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>

      <!-- 用户信息 + 登出 -->
      <el-dropdown trigger="click">
        <div class="app-header__user">
          <el-avatar :size="32" :icon="UserFilled" />
          <span class="app-header__username">{{ userStore.username || '用户' }}</span>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item @click="handleSwitchTenant">
              <el-icon><Switch /></el-icon>切换租户
            </el-dropdown-item>
            <el-dropdown-item divided @click="handleLogout">
              <el-icon><SwitchButton /></el-icon>退出登录
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </el-header>
</template>

<style scoped>
.app-header {
  height: var(--header-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--spacing-xl);
  background: var(--glass-bg);
  border-bottom: 1px solid var(--glass-border);
  backdrop-filter: blur(12px);
  box-shadow: var(--shadow-sm);
  z-index: 1000;
}

.app-header__left {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
}

.app-header__logo {
  display: flex;
  align-items: center;
}

.app-header__title {
  font-size: 18px;
  font-weight: 600;
  color: var(--color-primary);
  white-space: nowrap;
}

.app-header__collapse-btn {
  font-size: 20px;
  cursor: pointer;
  color: var(--color-text-regular);
  transition: color 0.2s;
}

.app-header__collapse-btn:hover {
  color: var(--color-primary);
}

.app-header__right {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.app-header__tenant {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
  border-radius: 4px;
  background: color-mix(in oklab, var(--color-primary) 8%, transparent);
}

.app-header__tenant-label {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.app-header__tenant-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-primary);
}

.tenant-switch-btn {
  margin-left: 4px;
  font-size: 12px;
  color: var(--color-primary);
}

.app-header__user {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  cursor: pointer;
}

.app-header__username {
  font-size: 14px;
  color: var(--color-text-regular);
}

.theme-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--color-text-regular);
}

.app-header.aurora-local {
  background:
    linear-gradient(105deg, rgba(8, 27, 58, 0.88), rgba(17, 46, 96, 0.84)),
    radial-gradient(120% 100% at 0% 0%, rgba(52, 212, 255, 0.2), transparent 60%);
  border-bottom-color: rgba(86, 208, 255, 0.32);
  box-shadow: 0 12px 30px rgba(5, 20, 46, 0.34);
}

.app-header.aurora-local .app-header__title {
  color: #9de8ff;
  text-shadow: 0 0 16px rgba(92, 233, 255, 0.44);
}

.app-header.aurora-local .app-header__username,
.app-header.aurora-local .theme-btn,
.app-header.aurora-local .app-header__collapse-btn {
  color: #d3ecff;
}

.app-header.aurora-local .app-header__tenant-label {
  color: #8ecfee;
}

.app-header.aurora-local .app-header__tenant-name {
  color: #e4f6ff;
}

.app-header.aurora-local .app-header__tenant {
  background: rgba(52, 160, 220, 0.12);
}
</style>
