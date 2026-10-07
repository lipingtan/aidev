<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const userStore = useUserStore()

const sidebarCollapsed = ref(false)
const activePath = computed(() => route.path)
const showUserMenu = ref(false)

// 是否显示侧边栏：已登录且不在首页
const showSidebar = computed(() => {
  return userStore.isLoggedIn && route.path !== '/home'
})

// CMS standalone：body 自然高度；管理后台：用 layout-fixed class 锁定
function syncBodyScroll() {
  // 先清掉可能残留的内联样式
  document.body.style.overflow = ''
  document.body.style.height = ''
  document.documentElement.style.overflow = ''
  document.documentElement.style.height = ''
  if (showSidebar.value) {
    document.body.classList.add('layout-fixed')
  } else {
    document.body.classList.remove('layout-fixed')
  }
}

watch(showSidebar, syncBodyScroll)

onMounted(() => {
  syncBodyScroll()
  if (userStore.token) {
    userStore.fetchMenu()
  }
  document.addEventListener('click', handleDocumentClick)
})

onUnmounted(() => {
  document.body.classList.remove('layout-fixed')
  document.removeEventListener('click', handleDocumentClick)
})

function handleDocumentClick(e: MouseEvent) {
  if (showUserMenu.value) {
    const target = e.target as HTMLElement
    const userMenuWrap = document.querySelector('.user-menu-wrap')
    const userDropdown = document.querySelector('.user-dropdown')
    if (userMenuWrap && !userMenuWrap.contains(target) && userDropdown && !userDropdown.contains(target)) {
      showUserMenu.value = false
    }
  }
}

function handleLogout() {
  userStore.logout()
}

function toggleUserMenu() {
  showUserMenu.value = !showUserMenu.value
}
</script>

<template>
  <div class="pc-layout" :class="{ 'pc-layout--collapsed': sidebarCollapsed, 'pc-layout--no-sidebar': !showSidebar }">

    <!-- 侧边栏（仅登录后且非首页显示） -->
    <aside v-if="showSidebar" class="sidebar">
      <div class="sidebar-header">
        <div class="sidebar-logo">
          <div class="logo-icon">千</div>
          <Transition name="fade">
            <span v-if="!sidebarCollapsed" class="logo-text">千分软件</span>
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
    </aside>

    <!-- 主内容区 -->
    <div class="main-wrap">
      <!-- 顶部导航栏 -->
      <header class="topbar">
        <div class="topbar-left">
          <!-- 未登录或首页时显示 Logo -->
          <template v-if="!showSidebar">
            <router-link to="/home" class="topbar-brand">
              <img src="@/assets/logo.png" alt="千分软件" class="brand-logo" />
            </router-link>
          </template>
          <!-- 已登录非首页显示面包屑 -->
          <template v-else>
            <router-link to="/home" class="topbar-brand-mini">
              <img src="@/assets/logo.png" alt="千分软件" class="brand-logo-mini" />
            </router-link>
            <nav class="breadcrumb" aria-label="breadcrumb">
              <router-link to="/home" class="breadcrumb-home">首页</router-link>
              <span class="breadcrumb-sep">/</span>
              <span class="breadcrumb-current">{{ route.meta.title || '工作台' }}</span>
            </nav>
          </template>
        </div>
        <div class="topbar-right">
          <!-- 导航菜单（首页时显示） -->
          <nav v-if="!showSidebar" class="topbar-nav">
            <router-link to="/home" class="nav-link">首页</router-link>
            <a href="https://btcola.com/c/product.html" target="_blank" class="nav-link">产品中心</a>
            <a href="https://btcola.com/d/message.html" target="_blank" class="nav-link">我要咨询</a>
          </nav>
          <!-- 已登录：用户菜单 -->
          <template v-if="userStore.isLoggedIn">
            <span v-if="userStore.tenantName" class="topbar-tenant">{{ userStore.tenantName }}</span>
            <div class="user-menu-wrap" @click="toggleUserMenu">
              <div class="topbar-avatar">{{ (userStore.phone || 'U').slice(-2) }}</div>
              <span class="topbar-phone">{{ userStore.phone }}</span>
              <svg class="dropdown-arrow" width="12" height="12" viewBox="0 0 24 24" fill="none">
                <path d="M6 9l6 6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </div>
            <Transition name="dropdown">
              <div v-if="showUserMenu" class="user-dropdown">
                <div class="dropdown-header">
                  <span class="dropdown-phone">{{ userStore.phone }}</span>
                  <span v-if="userStore.tenantName" class="dropdown-tenant">{{ userStore.tenantName }}</span>
                </div>
                <div class="dropdown-divider"></div>
                <button class="dropdown-item" @click="handleLogout">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                    <path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4M16 17l5-5-5-5M21 12H9" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                  退出登录
                </button>
              </div>
            </Transition>
          </template>
          <!-- 未登录：登录按钮 -->
          <button v-else class="topbar-login-btn" @click="userStore.requireLogin()">登录</button>
        </div>
      </header>

      <!-- 页面内容 -->
      <main class="page-content" :class="{ 'page-content--full': !showSidebar }">
        <router-view />
      </main>

      <!-- 底部备案信息 -->
      <footer class="pc-footer">
        <div class="footer-content">
          <span class="copyright">© 2019. All Rights Reserved. 千分软件</span>
          <div class="beian-links">
            <a href="https://beian.miit.gov.cn" target="_blank" rel="noopener noreferrer">渝ICP备19013468号-1</a>
            <a href="http://www.beian.gov.cn/portal/registerSystemInfo?recordcode=50010502001858" target="_blank" rel="noopener noreferrer">
              <img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABQAAAAUCAYAAACNiR0NAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAADfSURBVDhPY2AYBYMeMDIy/oeygAAggJgImf3//78BkO0PhCxQGhCALGRkZBQA0v8ZGRn/AwVAYiC5/4yMDEA1DFAaLgcSBwJGRsYGIP0fJAbSA9ULkgMCBqgaoCRYHCQGFAPLQeVAciB9QDYDVAzEBwJQQqDKQPKgwAPxQeIgsv8/AwMQg8RA4kAxIJ8BKgaig+RAhgDlQBgUyCAxEA0KdJAcSA5Eg+RA8kAaJAfCIL0gDFIHsoFqQYYB1TFA+WA+SA4kD8IgDJIDqQOxQfJAPogPlAMZCpQDsUFskNwoGOIAALOgU+fv0HdAAAAAAElFTkSuQmCC" alt="公安备案" class="beian-icon"/>
              渝公网安备 50010502001858号
            </a>
          </div>
        </div>
      </footer>
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

/* 无侧边栏模式 */
.pc-layout--no-sidebar {
  --sidebar-w: 0px;
  --sidebar-w-collapsed: 0px;
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

/* ── 主内容区 ── */
.main-wrap {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}

/* 顶部导航栏 - 黑色风格 */
.topbar {
  height: 50px;
  background: #222;
  border-bottom: none;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(0,0,0,0.15);
}

/* standalone模式（CMS主站页面）隐藏系统topbar，CMS页面有自己的CmsNav */
.pc-layout--no-sidebar .topbar {
  display: none;
}

.topbar-left { display: flex; align-items: center; gap: 16px; }

/* 顶部 Logo */
.topbar-brand {
  display: flex;
  align-items: center;
  text-decoration: none;
}

.brand-logo {
  height: 30px;
  width: auto;
}

.topbar-brand-mini {
  display: flex;
  align-items: center;
  text-decoration: none;
  margin-right: 12px;
}

.brand-logo-mini {
  height: 24px;
  width: auto;
}

/* 顶部导航菜单 */
.topbar-nav {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-left: 24px;
}

.nav-link {
  padding: 8px 16px;
  color: rgba(255,255,255,0.85);
  text-decoration: none;
  font-size: 14px;
  border-radius: 4px;
  transition: background 0.15s, color 0.15s;
}

.nav-link:hover {
  background: rgba(255,255,255,0.1);
  color: #fff;
}

.nav-link.router-link-active {
  color: #fff;
  background: rgba(255,255,255,0.15);
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
}

.breadcrumb-home { 
  color: rgba(255,255,255,0.6); 
  text-decoration: none;
  transition: color 0.15s;
}
.breadcrumb-home:hover { color: #fff; }
.breadcrumb-sep { color: rgba(255,255,255,0.3); }
.breadcrumb-current { color: rgba(255,255,255,0.9); font-weight: 500; }

.topbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.topbar-tenant {
  font-size: 13px;
  color: rgba(255,255,255,0.7);
  background: rgba(255,255,255,0.1);
  padding: 4px 10px;
  border-radius: 6px;
  white-space: nowrap;
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
  color: rgba(255,255,255,0.9);
}

/* 用户菜单 */
.user-menu-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s;
}

.user-menu-wrap:hover {
  background: rgba(255,255,255,0.1);
}

.dropdown-arrow {
  color: rgba(255,255,255,0.6);
  transition: transform 0.15s;
}

.user-menu-wrap:hover .dropdown-arrow {
  transform: translateY(1px);
}

.user-dropdown {
  position: absolute;
  top: 60px;
  right: 24px;
  width: 220px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.12);
  border: 1px solid #f0f0f5;
  overflow: hidden;
  z-index: 100;
}

.dropdown-header {
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.dropdown-phone {
  font-size: 14px;
  font-weight: 600;
  color: #111827;
}

.dropdown-tenant {
  font-size: 12px;
  color: #9ca3af;
}

.dropdown-divider {
  height: 1px;
  background: #f0f0f5;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 12px 16px;
  border: none;
  background: none;
  font-size: 14px;
  color: #374151;
  cursor: pointer;
  transition: background 0.15s;
}

.dropdown-item:hover {
  background: #f3f4f6;
}

.dropdown-item svg {
  color: #9ca3af;
}

/* 未登录顶栏登录按钮 */
.topbar-login-btn {
  height: 32px;
  padding: 0 16px;
  border: 1.5px solid rgba(255,255,255,0.5);
  border-radius: 6px;
  background: transparent;
  color: #fff;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
}

.topbar-login-btn:hover {
  background: rgba(255,255,255,0.1);
  border-color: rgba(255,255,255,0.8);
}

/* 下拉菜单动画 */
.dropdown-enter-active, .dropdown-leave-active {
  transition: opacity 0.15s, transform 0.15s;
}

.dropdown-enter-from, .dropdown-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

/* 页面内容 */
.page-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

/* 全屏内容（无侧边栏时，CMS主站页面） */
.page-content--full {
  padding: 0;
  flex: 1;
  overflow: visible; /* 让内容自然撑开，不截断 */
}

/* standalone 模式：整个 main-wrap 和 pc-layout 不限制高度，允许页面自然滚动 */
.pc-layout--no-sidebar {
  height: auto;
  min-height: 100vh;
  overflow-y: auto;
}
.pc-layout--no-sidebar .main-wrap {
  overflow: visible;
  height: auto;
  min-height: 100vh;
}

/* ── 底部备案信息（仅非standalone模式显示）── */
.pc-footer {
  background: #fff;
  border-top: 1px solid #f0f0f5;
  padding: 16px 24px;
  flex-shrink: 0;
}

/* standalone模式（CMS页面）隐藏系统页脚，CMS页面有自己的Footer */
.pc-layout--no-sidebar .pc-footer {
  display: none;
}

.footer-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.copyright {
  font-size: 13px;
  color: #9ca3af;
}

.beian-links {
  display: flex;
  align-items: center;
  gap: 20px;
}

.beian-links a {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: #9ca3af;
  text-decoration: none;
  transition: color 0.15s;
}

.beian-links a:hover {
  color: #6b7280;
}

.beian-icon {
  width: 14px;
  height: 14px;
}

/* ── 过渡动画 ── */
.fade-enter-active, .fade-leave-active { transition: opacity 0.15s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
