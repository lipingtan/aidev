<!--
  AppLayout — 响应式布局路由出口
  根据屏幕宽度（< 768px = H5，>= 768px = PC）动态切换布局组件
  使用 window.matchMedia 监听，实时响应窗口 resize
-->
<script setup lang="ts">
import { ref, onMounted, onUnmounted, defineAsyncComponent } from 'vue'

const H5Layout = defineAsyncComponent(() => import('./h5/H5Layout.vue'))
const PcLayout = defineAsyncComponent(() => import('./pc/PcLayout.vue'))

const MOBILE_BP = 768

const isNarrow = ref(window.innerWidth < MOBILE_BP)

const mq = window.matchMedia(`(max-width: ${MOBILE_BP - 1}px)`)

function onMQChange(e: MediaQueryListEvent) {
  isNarrow.value = e.matches
}

onMounted(() => {
  mq.addEventListener('change', onMQChange)
})

onUnmounted(() => {
  mq.removeEventListener('change', onMQChange)
})
</script>

<template>
  <component :is="isNarrow ? H5Layout : PcLayout" />
</template>
