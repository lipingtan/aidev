<!--
  主布局容器
  使用 el-container 组合顶部导航、侧边栏、标签页、面包屑和内容区
-->
<script setup lang="ts">
import { useAppStore } from '@/store/modules/app'
import AppHeader from './AppHeader.vue'
import AppSidebar from './AppSidebar.vue'
import TagsView from './TagsView.vue'
import AppBreadcrumb from './AppBreadcrumb.vue'

const appStore = useAppStore()
</script>

<template>
  <el-container class="app-layout" :class="{ 'aurora-local': appStore.themeMode === 'aurora-local' }">
    <AppHeader />
    <el-container class="app-layout__body">
      <AppSidebar />
      <el-main class="app-main">
        <TagsView />
        <AppBreadcrumb />
        <div class="app-main__content">
          <router-view />
        </div>
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.app-layout {
  height: 100vh;
  flex-direction: column;
  overflow: hidden;
}

.app-layout__body {
  flex: 1;
  overflow: hidden;    /* 防止任何子元素撑开整体 */
  display: flex;
}

.app-main {
  padding: 0;
  flex: 1;
  display: flex;
  flex-direction: column;  /* TagsView 固定在顶，内容区独立滚动 */
  min-width: 0;
  overflow: hidden;
  background-color: var(--color-bg-page);
}

/* TagsView 和 Breadcrumb 固定，不参与滚动 */
.app-main :deep(.tags-view),
.app-main :deep(.app-breadcrumb) {
  flex-shrink: 0;
}

.app-main__content {
  flex: 1;
  overflow-y: auto;        /* 只有内容区上下滚动 */
  padding: var(--content-padding);
  position: relative;
  z-index: 1;
}

.app-layout.aurora-local .app-main {
  background:
    radial-gradient(130% 100% at 0% 0%, rgba(88, 220, 255, 0.14), transparent 58%),
    linear-gradient(160deg, #071327, #0a1b36 60%, #08172f);
}

.app-layout.aurora-local .app-main__content {
  backdrop-filter: blur(2px);
}
</style>
