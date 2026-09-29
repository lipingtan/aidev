<script setup lang="ts">
/**
 * 根组件 — 模式 A（访问时触发登录）
 * - 所有页面均可访问（requiresAuth: false / 'soft'）
 * - 未登录时布局顶栏显示"登录"按钮
 * - 全局登录弹窗由 userStore.showLoginDialog 控制
 */
import { ref, computed, onMounted, onUnmounted, defineAsyncComponent } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'

const UserH5Layout = defineAsyncComponent(() => import('@/layout/h5/UserH5Layout.vue'))
const UserPcLayout = defineAsyncComponent(() => import('@/layout/pc/UserPcLayout.vue'))
const LoginDialog = defineAsyncComponent(() => import('@/views/login/LoginDialog.vue'))

const MOBILE_BP = 768
const isNarrow = ref(window.innerWidth < MOBILE_BP)
const mq = window.matchMedia(`(max-width: ${MOBILE_BP - 1}px)`)

function onMQChange(e: MediaQueryListEvent) { isNarrow.value = e.matches }
onMounted(() => {
  mq.addEventListener('change', onMQChange)
  // 始终调用 fetchMenu：
  // - 已登录：从认证接口获取完整菜单
  // - 未登录：从公开接口获取插件动态菜单（如「通知栏」）
  const store = useUserStore()
  store.fetchMenu()
})
onUnmounted(() => mq.removeEventListener('change', onMQChange))

const route = useRoute()
const userStore = useUserStore()

// 登录页 或 standalone 页面不套 layout
const noLayout = computed(() =>
  (route.meta.requiresAuth === false && route.path === '/login') ||
  route.meta.standalone === true
)
</script>

<template>
  <!-- 登录页：无 layout -->
  <router-view v-if="noLayout" />
  <!-- 其他页面：套布局 -->
  <component v-else :is="isNarrow ? UserH5Layout : UserPcLayout" />

  <!-- 全局登录弹窗（模式 A：原地登录不跳页） -->
  <LoginDialog v-if="userStore.showLoginDialog" @close="userStore.showLoginDialog = false" />
</template>

<style>
html, body, #app {
  margin: 0;
  padding: 0;
  height: 100%;
}
</style>

