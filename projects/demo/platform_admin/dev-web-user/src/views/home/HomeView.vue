<!--
  H5 首页
  - 顶部欢迎语 + 用户名
  - 中部 van-grid 快捷入口卡片
-->
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/store/modules/user'
import { showToast } from 'vant'
import { getProfile } from '@/api/owner'

const router = useRouter()
const userStore = useUserStore()
const communityName = ref('欢迎使用')

/** 快捷入口配置 */
const shortcuts = [
  { text: '个人信息', icon: 'user-o', path: '/mine/profile', needLogin: true }
]

/** 点击快捷入口 */
function onShortcutClick(item: typeof shortcuts[number]) {
  if (!item.path) {
    showToast('功能开发中，敬请期待')
    return
  }
  if (item.needLogin && !userStore.token) {
    router.push({ path: '/login', query: { redirect: item.path } })
    return
  }
  router.push(item.path)
}

async function loadCommunityName() {
  if (!userStore.token) {
    communityName.value = '访客'
    return
  }
  try {
    const profile = await getProfile()
    communityName.value = profile.name || '欢迎使用'
  } catch {
    communityName.value = '欢迎使用'
  }
}

onMounted(() => {
  loadCommunityName()
})
</script>

<template>
  <div class="home-page">
    <!-- 欢迎区域 -->
    <div class="home-welcome">
      <p class="home-welcome__greeting">
        你好，{{ userStore.username || '用户' }}
      </p>
      <p class="home-welcome__community">{{ communityName }}</p>
    </div>

    <!-- 快捷入口 -->
    <div class="home-shortcuts">
      <van-grid :column-num="3" :gutter="10">
        <van-grid-item
          v-for="item in shortcuts"
          :key="item.text"
          :icon="item.icon"
          :text="item.text"
          @click="onShortcutClick(item)"
        />
      </van-grid>
    </div>
  </div>
</template>

<style scoped>
.home-page {
  padding: var(--spacing-lg);
}

.home-welcome {
  position: relative;
  overflow: hidden;
  background: linear-gradient(135deg, var(--color-brand-gradient-start), var(--color-brand-gradient-end));
  border-radius: var(--border-radius-lg);
  padding: var(--spacing-xl);
  margin-bottom: var(--spacing-lg);
  color: #fff;
  box-shadow: var(--shadow-lg);
}

.home-welcome::after {
  content: '';
  position: absolute;
  right: -30%;
  top: -40%;
  width: 180px;
  height: 180px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.15);
  filter: blur(4px);
}

.home-welcome__greeting {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: var(--spacing-xs);
}

.home-welcome__community {
  font-size: 14px;
  opacity: 0.85;
}

.home-shortcuts {
  margin-bottom: var(--spacing-lg);
}

.home-shortcuts :deep(.van-grid-item__content) {
  border-radius: var(--border-radius-md);
  background: var(--color-surface);
  box-shadow: var(--shadow-sm);
}

.home-shortcuts :deep(.van-grid-item__text) {
  color: var(--color-text-primary);
  font-weight: 600;
  letter-spacing: 0.2px;
  text-shadow: 0 1px 1px rgba(0, 0, 0, 0.08);
}
</style>
