<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const sidebarCollapsed = ref(false)
const activePath = computed(() => route.path)

onMounted(() => {
  if (userStore.token) {
    userStore.fetchMenu()
  }
})

function handleLogout() {
  userStore.logout()
}
</script>

<template>
  <div class="pc-layout" :class="{ 'pc-layout--collapsed': sidebarCollapsed }">

    <!-- 侧边栏 -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <div class="sidebar-logo">
          <div class="logo-icon">U</div>
          <Transition name="fade">
            <span v-if="!sidebarCollapsed" class="logo-text">用户工作台</span>
          </Transition>
        </div>
        <button class="collapse-btn" @click="sidebarCollapsed = !sidebarCollapsed" :title="sidebarCollapsed ? '展开' : '收起'">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
            <path v-if="!sidebarCollapsed" d="M15 18l-6-6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path v-else d="M9 18l6-6-6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </button>
      </div>

      <nav class="sidebar-nav">
        <template v-if="userStore.menus.length > 0">
          <template v-for="menu in userStore.menus" :key="menu.id">
            <router-link
              v-if="!menu.children || menu.children.length === 0"
              :to="menu.path"
              class="nav-item"
              :class="{ 'nav-item--active': activePath === menu.path }"
              :title="sidebarCollapsed ? menu.title : ''"
            >
              <span class="nav-icon">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                  <rect x="3" y="3" width="7" height="7" rx="1.5" fill="currentColor" opacity="0.7"/>
                  <rect x="14" y="3" width="7" height="7" rx="1.5" fill="currentColor"/>
                  <rect x="3" y="14" width="7" height="7" rx="1.5" fill="currentColor"/>
                  <rect x="14" y="14" width="7" height="7" rx="1.5" fill="currentColor" opacity="0.7"/>
                </svg>
              </span>
              <Transition name="fade"><span v-if="!sidebarCollapsed" class="nav-label">{{ menu.title }}</span></Transition>
              <span v-if="!sidebarCollapsed" class="nav-active-bar"/>
            </router-link>
          </template>
        </template>
        <div v-else class="nav-empty">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" opacity="0.3">
            <path d="M3 12h18M3 6h18M3 18h18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
        </div>
      </nav>

      <!-- 用户信息 -->
      <div class="sidebar-user">
        <div class="user-avatar">{{ (userStore.phone || 'U').slice(-2) }}</div>
        <Transition name="fade">
          <div v-if="!sidebarCollapsed" class="user-info">
            <span class="user-phone">{{ userStore.phone }}</span>
            <button class="logout-btn" @click="handleLogout">退出登录</button>
          </div>
        </Transition>
        <button v-if="sidebarCollapsed" class="logout-icon-btn" @click="handleLogout" title="退出登录">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
            <path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4M16 17l5-5-5-5M21 12H9" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </button>
      </div>
    </aside>

    <!-- 主内容区 -->
    <div class="main-wrap">
      <!-- 顶部导航栏 -->
      <header class="topbar">
        <div class="topbar-left">
          <nav class="breadcrumb" aria-label="breadcrumb">
            <span class="breadcrumb-home">首页</span>
            <span class="breadcrumb-sep">/</span>
            <span class="breadcrumb-current">{{ route.meta.title || '工作台' }}</span>
          </nav>
        </div>
        <div class="topbar-right">
          <div class="topbar-avatar">{{ (userStore.phone || 'U').slice(-2) }}</div>
          <span class="topbar-phone">{{ userStore.phone }}</span>
        </div>
      </header>

      <!-- 页面内容 -->
      <main class="page-content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<style scoped>
/* ── 根布局 ── */
.pc-layout {
  display: flex;
  height: 100vh;
  height: 100dvh;
  background: #f4f6fa;
  --sidebar-w: 220px;
  --sidebar-w-collapsed: 68px;
  --sidebar-bg: #1a1d2e;
  --primary: #4f46e5;
  --primary-light: rgba(79,70,229,0.12);
}

/* ── 侧边栏 ── */
.sidebar {
  width: var(--sidebar-w);
  background: var(--sidebar-bg);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  transition: width 0.25s cubic-bezier(.4,0,.2,1);
  overflow: hidden;
  border-right: 1px solid rgba(255,255,255,0.06);
}

.pc-layout--collapsed .sidebar {
  width: var(--sidebar-w-collapsed);
}

/* 头部 logo */
.sidebar-header {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  border-bottom: 1px solid rgba(255,255,255,0.06);
  flex-shrink: 0;
}

.sidebar-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.logo-icon {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  background: linear-gradient(135deg, var(--primary), #7c3aed);
  color: #fff;
  font-size: 15px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(79,70,229,0.4);
}

.logo-text {
  font-size: 14px;
  font-weight: 700;
  color: #fff;
  white-space: nowrap;
}

.collapse-btn {
  width: 28px;
  height: 28px;
  border: none;
  background: rgba(255,255,255,0.08);
  border-radius: 7px;
  color: rgba(255,255,255,0.5);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: background 0.15s, color 0.15s;
}
.collapse-btn:hover { background: rgba(255,255,255,0.14); color: #fff; }

/* 导航 */
.sidebar-nav {
  flex: 1;
  overflow-y: auto;
  padding: 12px 10px;
  scrollbar-width: none;
}
.sidebar-nav::-webkit-scrollbar { display: none; }

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 42px;
  padding: 0 10px;
  border-radius: 10px;
  color: rgba(255,255,255,0.55);
  text-decoration: none;
  margin-bottom: 4px;
  transition: background 0.15s, color 0.15s;
  position: relative;
  white-space: nowrap;
  cursor: pointer;
}

.nav-item:hover {
  background: rgba(255,255,255,0.07);
  color: rgba(255,255,255,0.85);
}

.nav-item--active {
  background: var(--primary-light);
  color: #a5b4fc;
}

.nav-item--active .nav-active-bar {
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 20px;
  background: var(--primary);
  border-radius: 2px 0 0 2px;
}

.nav-icon { display: flex; align-items: center; flex-shrink: 0; }
.nav-label { font-size: 14px; font-weight: 500; }

.nav-empty {
  display: flex;
  justify-content: center;
  padding-top: 32px;
}

/* 用户信息 */
.sidebar-user {
  padding: 14px 12px;
  border-top: 1px solid rgba(255,255,255,0.06);
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.user-info {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
}

.user-phone {
  font-size: 13px;
  color: rgba(255,255,255,0.75);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.logout-btn {
  background: none;
  border: none;
  color: rgba(255,255,255,0.35);
  font-size: 11px;
  cursor: pointer;
  padding: 0;
  text-align: left;
  transition: color 0.15s;
}
.logout-btn:hover { color: #f87171; }

.logout-icon-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: rgba(255,255,255,0.06);
  border-radius: 8px;
  color: rgba(255,255,255,0.4);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s, color 0.15s;
}
.logout-icon-btn:hover { background: rgba(248,113,113,0.15); color: #f87171; }

/* ── 主内容区 ── */
.main-wrap {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}

/* 顶部导航栏 */
.topbar {
  height: 64px;
  background: #fff;
  border-bottom: 1px solid #f0f0f5;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  flex-shrink: 0;
  box-shadow: 0 1px 4px rgba(0,0,0,0.04);
}

.topbar-left { display: flex; align-items: center; }

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
}

.breadcrumb-home { color: #9ca3af; }
.breadcrumb-sep { color: #d1d5db; }
.breadcrumb-current { color: #374151; font-weight: 500; }

.topbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.topbar-avatar {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: linear-gradient(135deg, #4f46e5, #7c3aed);
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}

.topbar-phone {
  font-size: 14px;
  color: #374151;
}

/* 页面内容 */
.page-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

/* ── 过渡动画 ── */
.fade-enter-active, .fade-leave-active { transition: opacity 0.15s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
