<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const userStore = useUserStore()

const activeTab = ref('')

onMounted(async () => {
  if (userStore.token) {
    await userStore.fetchMenu()
  }
  activeTab.value = route.path
})

const tabMenus = computed(() =>
  userStore.menus.slice(0, 5).map(menu => ({
    path: menu.path,
    title: menu.title,
    icon: menu.icon || 'apps'
  }))
)

const pageTitle = computed(() => (route.meta.title as string) || '工作台')

function handleLogout() {
  userStore.logout()
}
</script>

<template>
  <div class="h5-layout">

    <!-- 顶部导航栏 -->
    <header class="h5-topbar">
      <div class="topbar-left">
        <div class="topbar-logo">U</div>
        <span class="topbar-title">{{ pageTitle }}</span>
      </div>
      <div class="topbar-right">
        <div class="user-menu-wrap">
          <div class="topbar-avatar" @click="handleLogout" role="button" tabindex="0" aria-label="退出登录">
            {{ (userStore.phone || 'U').slice(-2) }}
          </div>
        </div>
      </div>
    </header>

    <!-- 页面内容 -->
    <main class="h5-main">
      <router-view />
    </main>

    <!-- 底部 Tab 栏 -->
    <nav v-if="tabMenus.length > 0" class="h5-tabbar" aria-label="主导航">
      <router-link
        v-for="item in tabMenus"
        :key="item.path"
        :to="item.path"
        class="tab-item"
        :class="{ 'tab-item--active': activeTab === item.path }"
        @click="activeTab = item.path"
      >
        <span class="tab-icon">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none">
            <rect x="3" y="3" width="7" height="7" rx="1.5" fill="currentColor" opacity="0.7"/>
            <rect x="14" y="3" width="7" height="7" rx="1.5" fill="currentColor"/>
            <rect x="3" y="14" width="7" height="7" rx="1.5" fill="currentColor"/>
            <rect x="14" y="14" width="7" height="7" rx="1.5" fill="currentColor" opacity="0.7"/>
          </svg>
        </span>
        <span class="tab-label">{{ item.title }}</span>
      </router-link>

      <!-- 无菜单时固定显示首页 + 退出 -->
      <template v-if="tabMenus.length === 0">
        <router-link to="/home" class="tab-item tab-item--active">
          <span class="tab-icon">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none">
              <path d="M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z" fill="currentColor"/>
            </svg>
          </span>
          <span class="tab-label">首页</span>
        </router-link>
      </template>
    </nav>

    <!-- 退出确认 Toast（点击头像提示） -->
    <Teleport to="body">
      <Transition name="slide-up">
        <div v-if="false" class="logout-sheet">
          <button class="sheet-btn sheet-btn--danger" @click="handleLogout">退出登录</button>
          <button class="sheet-btn">取消</button>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
/* ── 根布局 ── */
.h5-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  height: 100dvh;
  background: #f4f6fa;
  --primary: #4f46e5;
  --primary-rgb: 79,70,229;
}

/* ── 顶部导航栏 ── */
.h5-topbar {
  height: 56px;
  background: #fff;
  border-bottom: 1px solid #f0f0f5;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  flex-shrink: 0;
  box-shadow: 0 1px 6px rgba(0,0,0,0.05);
  /* iOS safe area */
  padding-top: env(safe-area-inset-top);
  height: calc(56px + env(safe-area-inset-top));
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.topbar-logo {
  width: 32px;
  height: 32px;
  border-radius: 9px;
  background: linear-gradient(135deg, var(--primary), #7c3aed);
  color: #fff;
  font-size: 14px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

.topbar-title {
  font-size: 17px;
  font-weight: 600;
  color: #111827;
  letter-spacing: -0.2px;
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.topbar-avatar {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  user-select: none;
  transition: opacity 0.15s;
  position: relative;
}

.topbar-avatar::after {
  content: '退出';
  position: absolute;
  bottom: -28px;
  right: 0;
  background: rgba(0,0,0,0.7);
  color: #fff;
  font-size: 11px;
  padding: 3px 7px;
  border-radius: 5px;
  white-space: nowrap;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.15s;
}

.topbar-avatar:active::after { opacity: 1; }
.topbar-avatar:active { opacity: 0.75; }

/* ── 页面内容 ── */
.h5-main {
  flex: 1;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  overscroll-behavior-y: contain;
}

/* ── 底部 Tab 栏 ── */
.h5-tabbar {
  display: flex;
  background: #fff;
  border-top: 1px solid #f0f0f5;
  flex-shrink: 0;
  /* iOS safe area bottom */
  padding-bottom: env(safe-area-inset-bottom);
  box-shadow: 0 -2px 12px rgba(0,0,0,0.06);
}

.tab-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 8px 4px;
  gap: 3px;
  text-decoration: none;
  color: #9ca3af;
  transition: color 0.15s;
  min-height: 52px;
  cursor: pointer;
  user-select: none;
  -webkit-tap-highlight-color: transparent;
}

.tab-item:active { opacity: 0.7; }

.tab-item--active {
  color: var(--primary);
}

.tab-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.tab-item--active .tab-icon::after {
  content: '';
  position: absolute;
  inset: -4px;
  background: rgba(var(--primary-rgb), 0.1);
  border-radius: 10px;
}

.tab-label {
  font-size: 11px;
  font-weight: 500;
  line-height: 1;
}

/* ── 底部弹出退出确认 ── */
.logout-sheet {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: #fff;
  border-radius: 20px 20px 0 0;
  padding: 16px 16px calc(16px + env(safe-area-inset-bottom));
  display: flex;
  flex-direction: column;
  gap: 10px;
  box-shadow: 0 -8px 32px rgba(0,0,0,0.12);
  z-index: 9999;
}

.sheet-btn {
  height: 50px;
  border: none;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  background: #f4f6fa;
  color: #374151;
  transition: background 0.15s;
}

.sheet-btn--danger {
  background: #fef2f2;
  color: #ef4444;
}

.sheet-btn:active { opacity: 0.8; }

/* ── 动画 ── */
.slide-up-enter-active, .slide-up-leave-active {
  transition: transform 0.3s cubic-bezier(.4,0,.2,1);
}
.slide-up-enter-from, .slide-up-leave-to {
  transform: translateY(100%);
}
</style>
