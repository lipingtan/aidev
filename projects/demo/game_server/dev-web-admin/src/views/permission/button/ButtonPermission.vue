<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header><span>按钮权限演示</span></template>

      <h4>当前用户权限列表：</h4>
      <div style="margin-bottom: 20px;">
        <el-tag v-for="code in permissionCodes" :key="code" style="margin: 4px;">{{ code }}</el-tag>
        <el-empty v-if="!permissionCodes.length" description="暂无权限" />
      </div>

      <el-divider />

      <h4>v-auth 指令方式（无权限时按钮隐藏）：</h4>
      <div style="margin-bottom: 20px;">
        <el-button v-auth="'system:user:add'" type="primary">新增用户（system:user:add）</el-button>
        <el-button v-auth="'system:user:edit'" type="warning">编辑用户（system:user:edit）</el-button>
        <el-button v-auth="'system:user:delete'" type="danger">删除用户（system:user:delete）</el-button>
        <el-button v-auth="'fake:permission'" type="info">无效权限按钮（不可见）</el-button>
      </div>

      <el-divider />

      <h4>hasAuth 函数方式（无权限时按钮禁用）：</h4>
      <div>
        <el-button type="primary" :disabled="!hasAuth('system:user:add')">新增用户</el-button>
        <el-button type="warning" :disabled="!hasAuth('system:user:edit')">编辑用户</el-button>
        <el-button type="danger" :disabled="!hasAuth('system:user:delete')">删除用户</el-button>
        <el-button type="info" :disabled="!hasAuth('fake:permission')">无效权限按钮（禁用）</el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useUserStore } from '@/store/modules/user'

const userStore = useUserStore()
const permissionCodes = computed(() => userStore.permissions || [])

function hasAuth(code: string): boolean {
  return permissionCodes.value.includes(code)
}
</script>

<style scoped>
.page-container { padding: 16px; }
</style>
