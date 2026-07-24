<script setup lang="ts">
/**
 * 首页 — 欢迎页，响应式
 */
import { useUserStore } from '@/stores/user'
const userStore = useUserStore()
</script>

<template>
  <div class="home-page">
    <div class="welcome-card">
      <div class="welcome-icon">👋</div>
      <h2 class="welcome-title">欢迎回来</h2>
      <p class="welcome-sub">{{ userStore.phone }}</p>
      <p class="welcome-hint">请通过菜单导航到对应功能页面</p>
    </div>
  </div>
</template>

<style scoped>
.home-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - 56px - 56px); /* 减去顶栏和底栏 */
  padding: 24px 16px;
  box-sizing: border-box;
}

.welcome-card {
  text-align: center;
  padding: 40px 32px;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.06);
  max-width: 380px;
  width: 100%;
}

.welcome-icon {
  font-size: 48px;
  margin-bottom: 12px;
}

.welcome-title {
  font-size: 22px;
  font-weight: 700;
  color: #1a1a2e;
  margin: 0 0 8px;
}

.welcome-sub {
  font-size: 14px;
  color: #667eea;
  margin: 0 0 12px;
  font-weight: 500;
}

.welcome-hint {
  font-size: 13px;
  color: #aaa;
  margin: 0;
}

@media (max-width: 480px) {
  .welcome-card {
    padding: 32px 20px;
  }
  .welcome-title { font-size: 20px; }
}
</style>
