<script setup lang="ts">
/**
 * 基础布局 — 响应式双模式
 * PC（≥768px）：左侧边栏 + 顶栏
 * 移动（<768px）：顶部 header + 底部 tabbar + 汉堡抽屉菜单
 */
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router    = useRouter()
const route     = useRoute()
const userStore = useUserStore()

const isCollapse    = ref(false)   // PC 侧边栏折叠
const drawerOpen    = ref(false)   // 移动端抽屉菜单
const isMobile      = ref(false)

/** 检测是否移动端 */
function checkMobile() {
  isMobile.value = window.innerWidth < 768
  // 切到 PC 时自动关闭抽屉
  if (!isMobile.value) drawerOpen.value = false
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  userStore.fetchMenu()
})
onUnmounted(() => window.removeEventListener('resize', checkMobile))

/** 顶级菜单（用于移动端 tabbar，最多显示 5 项） */
const tabbarMenus = computed(() => userStore.menus.slice(0, 5))

/** 当前激活菜单 path */
const activePath = computed(() => route.path)

function navigate(path: string) {
  router.push(path)
  drawerOpen.value = false
}

function handleLogout() {
  userStore.logout()
}
</script>

<template>
  <!-- ════════════════════════════════════
       PC 布局（≥768px）
  ════════════════════════════════════ -->
  <el-container v-if="!isMobile" class="pc-layout">
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

    <!-- 右侧 -->
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

  <!-- ════════════════════════════════════
       移动端布局（<768px）
  ════════════════════════════════════ -->
  <div v-else class="mobile-layout">
    <!-- 顶部导航栏 -->
    <header class="mobile-header">
      <button class="menu-btn" aria-label="菜单" @click="drawerOpen = true">
        <span class="menu-icon">☰</span>
      </button>
      <span class="mobile-title">用户端</span>
      <div class="mobile-header-right">
        <span class="user-phone-sm">{{ userStore.phone }}</span>
      </div>
    </header>

    <!-- 内容区 -->
    <main class="mobile-main">
      <router-view />
    </main>

    <!-- 底部 Tabbar（最多 5 个顶级菜单） -->
    <nav v-if="tabbarMenus.length" class="mobile-tabbar">
      <button
        v-for="menu in tabbarMenus"
        :key="menu.id"
        class="tab-item"
        :class="{ active: activePath.startsWith(menu.path) }"
        @click="navigate(menu.path)"
      >
        <el-icon class="tab-icon"><component :is="menu.icon || 'Document'" /></el-icon>
        <span class="tab-label">{{ menu.title }}</span>
      </button>
      <!-- 首页兜底 tab（无菜单时也显示） -->
      <button
        class="tab-item"
        :class="{ active: activePath === '/home' }"
        @click="navigate('/home')"
      >
        <el-icon class="tab-icon"><HomeFilled /></el-icon>
        <span class="tab-label">首页</span>
      </button>
    </nav>
    <!-- 无菜单时只保留底部首页 -->
    <nav v-else class="mobile-tabbar">
      <button class="tab-item active" @click="navigate('/home')">
        <el-icon class="tab-icon"><HomeFilled /></el-icon>
        <span class="tab-label">首页</span>
      </button>
      <button class="tab-item" @click="handleLogout">
        <el-icon class="tab-icon"><SwitchButton /></el-icon>
        <span class="tab-label">退出</span>
      </button>
    </nav>

    <!-- 抽屉菜单（全部菜单 + 退出） -->
    <transition name="drawer-fade">
      <div v-if="drawerOpen" class="drawer-overlay" @click="drawerOpen = false">
        <nav class="drawer-panel" @click.stop>
          <div class="drawer-header">
            <span>用户端</span>
            <button class="drawer-close" @click="drawerOpen = false">✕</button>
          </div>
          <div class="drawer-user">{{ userStore.phone }}</div>
          <ul class="drawer-menu">
            <template v-for="menu in userStore.menus" :key="menu.id">
              <!-- 有子菜单 -->
              <li v-if="menu.children && menu.children.length" class="drawer-group">
                <span class="drawer-group-title">{{ menu.title }}</span>
                <ul>
                  <li
                    v-for="child in menu.children"
                    :key="child.id"
                    class="drawer-item"
                    :class="{ active: activePath === child.path }"
                    @click="navigate(child.path)"
                  >
                    {{ child.title }}
                  </li>
                </ul>
              </li>
              <!-- 无子菜单 -->
              <li
                v-else
                class="drawer-item"
                :class="{ active: activePath === menu.path }"
                @click="navigate(menu.path)"
              >
                {{ menu.title }}
              </li>
            </template>
          </ul>
          <button class="drawer-logout" @click="handleLogout">退出登录</button>
        </nav>
      </div>
    </transition>
  </div>
</template>

<style scoped>
/* ═══════════════════════════
   PC 布局
═══════════════════════════ */
.pc-layout {
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

/* 覆盖 el-menu 默认背景，保持一致 */
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
.collapse-btn:hover { color: #667eea; }

.user-phone {
  font-size: 14px;
  color: #555;
}

.pc-main {
  background: #f4f6fa;
  overflow-y: auto;
}

/* ═══════════════════════════
   移动端布局
═══════════════════════════ */
.mobile-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  height: 100dvh;
}

/* 顶部 Header */
.mobile-header {
  height: 52px;
  min-height: 52px;
  display: flex;
  align-items: center;
  padding: 0 16px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: #fff;
  gap: 12px;
  position: sticky;
  top: 0;
  z-index: 100;
}

.menu-btn {
  background: none;
  border: none;
  color: #fff;
  padding: 6px;
  cursor: pointer;
  border-radius: 6px;
  line-height: 1;
}
.menu-icon { font-size: 20px; }

.mobile-title {
  flex: 1;
  font-size: 17px;
  font-weight: 600;
}

.mobile-header-right {
  display: flex;
  align-items: center;
}
.user-phone-sm {
  font-size: 12px;
  opacity: 0.85;
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 内容区 */
.mobile-main {
  flex: 1;
  overflow-y: auto;
  background: #f4f6fa;
  /* 底部 tabbar 高度留白 */
  padding-bottom: env(safe-area-inset-bottom);
}

/* 底部 Tabbar */
.mobile-tabbar {
  height: 56px;
  min-height: 56px;
  display: flex;
  align-items: center;
  background: #fff;
  border-top: 1px solid #eef0f3;
  box-shadow: 0 -2px 8px rgba(0,0,0,0.06);
  padding-bottom: env(safe-area-inset-bottom);
}

.tab-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  background: none;
  border: none;
  cursor: pointer;
  color: #999;
  padding: 6px 0;
  transition: color 0.2s;
}
.tab-item.active { color: #667eea; }
.tab-item:active { opacity: 0.7; }

.tab-icon { font-size: 20px; }
.tab-label { font-size: 10px; line-height: 1.2; }

/* 抽屉遮罩 */
.drawer-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
  z-index: 200;
  display: flex;
}

/* 抽屉面板 */
.drawer-panel {
  width: 240px;
  max-width: 75vw;
  height: 100%;
  background: #fff;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.drawer-header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: #fff;
  font-size: 16px;
  font-weight: 600;
  flex-shrink: 0;
}
.drawer-close {
  background: none;
  border: none;
  color: #fff;
  font-size: 18px;
  cursor: pointer;
  padding: 4px;
}

.drawer-user {
  padding: 12px 20px;
  font-size: 13px;
  color: #888;
  border-bottom: 1px solid #f0f0f0;
  flex-shrink: 0;
}

.drawer-menu {
  list-style: none;
  margin: 0;
  padding: 8px 0;
  flex: 1;
}

.drawer-group { padding: 4px 0; }
.drawer-group-title {
  display: block;
  padding: 8px 20px 4px;
  font-size: 11px;
  color: #bbb;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.drawer-group ul {
  list-style: none;
  margin: 0;
  padding: 0;
}

.drawer-item {
  padding: 13px 20px;
  font-size: 15px;
  color: #333;
  cursor: pointer;
  transition: background 0.15s;
  border-radius: 0;
}
.drawer-item:hover { background: #f5f7ff; }
.drawer-item.active {
  color: #667eea;
  background: #f0f3ff;
  font-weight: 600;
}

.drawer-logout {
  margin: 8px 16px 24px;
  padding: 12px;
  width: calc(100% - 32px);
  background: #fff1f0;
  border: 1px solid #ffccc7;
  border-radius: 8px;
  color: #ff4d4f;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.2s;
}
.drawer-logout:hover { background: #ffe8e8; }

/* 抽屉动画 */
.drawer-fade-enter-active,
.drawer-fade-leave-active {
  transition: opacity 0.2s;
}
.drawer-fade-enter-active .drawer-panel,
.drawer-fade-leave-active .drawer-panel {
  transition: transform 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}
.drawer-fade-enter-from,
.drawer-fade-leave-to {
  opacity: 0;
}
.drawer-fade-enter-from .drawer-panel {
  transform: translateX(-100%);
}
.drawer-fade-leave-to .drawer-panel {
  transform: translateX(-100%);
}
</style>
