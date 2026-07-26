<script setup lang="ts">
/**
 * User PC 布局
 * 侧边栏 + 顶栏模式（Element Plus 风格）
 */
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const isCollapse = ref(false)
const activePath = computed(() => route.path)

onMounted(() => {
  userStore.fetchMenu()
})

function handleLogout() {
  userStore.logout()
}
</script>

<template>
  <el-container class="user-pc-layout">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '220px'" class="pc-aside">
      <div class="pc-logo">
        <span v-if="!isCollapse" class="logo-text">用户端</span>
        <span v-else class="logo-icon">U</span>
      </div>
      <el-menu
        :collapse="isCollapse"
        :default-active="activePath"
        router
        background-color="#1e2a3a"
        text-color="#a8b8cc"
        active-text-color="#ffffff"
        active-background-color="#667eea"
        :collapse-transition="false"
      >
        <template v-for="menu in userStore.menus" :key="menu.id">
          <el-sub-menu
            v-if="menu.children && menu.children.length"
            :index="menu.path"
          >
            <template #title>
              <el-icon><component :is="menu.icon || 'Menu'" /></el-icon>
              <span>{{ menu.title }}</span>
            </template>
            <el-menu-item
              v-for="child in menu.children"
              :key="child.id"
              :index="child.path"
            >
              {{ child.title }}
            </el-menu-item>
          </el-sub-menu>
          <el-menu-item v-else :index="menu.path">
            <el-icon><component :is="menu.icon || 'Document'" /></el-icon>
            <template #title>{{ menu.title }}</template>
          </el-menu-item>
        </template>
      </el-menu>
    </el-aside>

    <!-- 右侧容器 -->
    <el-container>
      <el-header class="pc-header">
        <div class="pc-header-left">
          <el-icon class="collapse-btn" @click="isCollapse = !isCollapse">
            <component :is="isCollapse ? 'Expand' : 'Fold'" />
          </el-icon>
        </div>
        <div class="pc-header-right">
          <span class="user-phone">{{ userStore.phone }}</span>
          <el-button link type="primary" @click="handleLogout">退出登录</el-button>
        </div>
      </el-header>
      <el-main class="pc-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.user-pc-layout {
  height: 100vh;
}

.pc-aside {
  background: #1e2a3a;
  transition: width 0.25s;
  overflow: hidden;
  flex-shrink: 0;
}

.pc-logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 17px;
  font-weight: 700;
  border-bottom: 1px solid rgba(255,255,255,0.08);
  letter-spacing: 0.5px;
}

.logo-icon {
  font-size: 22px;
}

:deep(.el-menu) {
  border-right: none;
}

:deep(.el-menu-item.is-active) {
  background-color: #667eea !important;
  border-radius: 6px;
  margin: 2px 8px;
}

.pc-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border-bottom: 1px solid #eef0f3;
  padding: 0 20px;
  height: 56px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.06);
}

.pc-header-left,
.pc-header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.collapse-btn {
  font-size: 20px;
  cursor: pointer;
  color: #666;
  transition: color 0.2s;
}

.collapse-btn:hover {
  color: #667eea;
}

.user-phone {
  font-size: 14px;
  color: #555;
}

.pc-main {
  background: #f4f6fa;
  overflow-y: auto;
}
</style>
