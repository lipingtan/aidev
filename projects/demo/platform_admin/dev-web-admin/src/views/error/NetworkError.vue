<!--
  网络异常兜底页
  请求超时或 network error 时展示，提供重试按钮
-->
<script setup lang="ts">
import { useRouter } from 'vue-router'
import { WarnTriangleFilled } from '@element-plus/icons-vue'

const router = useRouter()

function handleRetry() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.replace('/')
  }
}
</script>

<template>
  <div class="network-error-page">
    <div class="network-error-content">
      <el-icon :size="80" color="#E6A23C">
        <WarnTriangleFilled />
      </el-icon>
      <h1 class="network-error-title">网络异常</h1>
      <p class="network-error-subtitle">请检查网络连接后重试</p>
      <el-button type="primary" @click="handleRetry">重试</el-button>
    </div>
  </div>
</template>

<style scoped>
.network-error-page {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
}

.network-error-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.network-error-title {
  margin: 0;
  font-size: 24px;
  font-weight: 500;
  color: #303133;
}

.network-error-subtitle {
  margin: 0;
  font-size: 14px;
  color: #909399;
}
</style>
