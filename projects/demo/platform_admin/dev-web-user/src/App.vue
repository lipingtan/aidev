<script setup lang="ts">
/**
 * 根组件 — 响应式双模式支持
 * /login 等不需要认证的页面：直接渲染 router-view（无 layout）
 * 登录后的页面：根据屏幕宽度切换 H5/PC layout
 */
import { ref, computed, onMounted, onUnmounted, defineAsyncComponent } from 'vue'
import { useRoute } from 'vue-router'

const UserH5Layout = defineAsyncComponent(() => import('@/layout/h5/UserH5Layout.vue'))
const UserPcLayout = defineAsyncComponent(() => import('@/layout/pc/UserPcLayout.vue'))

const MOBILE_BP = 768
const isNarrow = ref(window.innerWidth < MOBILE_BP)
const mq = window.matchMedia(`(max-width: ${MOBILE_BP - 1}px)`)

function onMQChange(e: MediaQueryListEvent) {
  isNarrow.value = e.matches
}

onMounted(() => mq.addEventListener('change', onMQChange))
onUnmounted(() => mq.removeEventListener('change', onMQChange))

const route = useRoute()
// 不需要认证的页面（登录页等）直接用空白路由视图，不套 layout
const noLayout = computed(() => route.meta.requiresAuth === false)
</script>

<template>
  <!-- 登录页：无 layout 直接渲染 -->
  <router-view v-if="noLayout" />
  <!-- 登录后页面：按屏幕宽度选择布局 -->
  <component v-else :is="isNarrow ? UserH5Layout : UserPcLayout" />
</template>

<style>
html, body, #app {
  margin: 0;
  padding: 0;
  height: 100%;
}
</style>
