<script setup lang="ts">
/**
 * User H5 布局
 * 底部 Tabbar 导航（Vant）
 */
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const activeTab = ref('')

onMounted(async () => {
  await userStore.fetchMenu()
  // 设置当前激活 tab
  activeTab.value = route.path
})

/** 底部 Tabbar 菜单项（顶级菜单） */
const tabMenus = computed(() => {
  return userStore.menus.map(menu => ({
    path: menu.path,
    title: menu.title,
    icon: menu.icon || 'home-o'
  }))
})
</script>

<template>
  <div class="user-h5-container">
    <div class="user-h5-content">
      <router-view />
    </div>
    <!-- Vant Tabbar 底部导航 -->
    <van-tabbar v-model="activeTab" route>
      <van-tabbar-item
        v-for="item in tabMenus"
        :key="item.path"
        :to="item.path"
        :icon="item.icon"
      >
        {{ item.title }}
      </van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<style scoped>
.user-h5-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  height: 100dvh;
}

.user-h5-content {
  flex: 1;
  overflow-y: auto;
  background: #f4f6fa;
}
</style>
