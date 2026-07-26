<script setup lang="ts">
/**
 * 根组件 — 双模式支持
 * 根据构建模式（__IS_H5__）注入对应布局
 */
import { defineAsyncComponent } from 'vue'

// @ts-ignore - Vite define 注入的全局常量
const isH5 = typeof __IS_H5__ !== 'undefined' ? __IS_H5__ : false

// 根据构建模式动态导入布局（实现 tree-shaking）
const Layout = isH5
  ? defineAsyncComponent(() => import('@/layout/h5/UserH5Layout.vue'))
  : defineAsyncComponent(() => import('@/layout/pc/UserPcLayout.vue'))
</script>

<template>
  <Layout />
</template>

<style>
html, body, #app {
  margin: 0;
  padding: 0;
  height: 100%;
}
</style>
