<!--
  H5 布局组件
  使用 Vant NavBar + Popup 侧滑菜单实现移动端布局
-->
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useMenuStore } from '@/store/modules/menu'
import H5Header from './H5Header.vue'
import H5MenuDrawer from './H5MenuDrawer.vue'

const route = useRoute()
const menuStore = useMenuStore()

const showMenu = ref(false)

// 当前页面标题（从路由 meta 或菜单中获取）
const currentPageTitle = computed(() => {
  return route.meta?.title as string || '管理后台'
})

// 后退处理
const onBack = () => {
  window.history.back()
}

// 菜单项
const menus = computed(() => menuStore.menus)

// H5Layout 挂载时主动加载菜单（与路由守卫的 userStore 独立）
onMounted(async () => {
  if (!menuStore.loaded && !menuStore.loading) {
    await menuStore.fetchMenus()
  }
})
</script>

<template>
  <div class="h5-layout">
    <!-- 顶部导航栏 -->
    <H5Header 
      :title="currentPageTitle"
      :show-back="route.path !== '/'"
      @back="onBack"
      @menu-click="showMenu = true"
    />

    <!-- 业务内容区 -->
    <div class="h5-content">
      <router-view />
    </div>

    <!-- 侧滑菜单 -->
    <van-popup 
      v-model:show="showMenu" 
      position="left" 
      :style="{ width: '75%', height: '100%' }"
    >
      <H5MenuDrawer :menus="menus" @close="showMenu = false" />
    </van-popup>
  </div>
</template>

<style scoped>
.h5-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

.h5-content {
  flex: 1;
  overflow-y: auto;
  background-color: var(--color-bg-page, #f5f5f5);
  -webkit-overflow-scrolling: touch;
}
</style>
