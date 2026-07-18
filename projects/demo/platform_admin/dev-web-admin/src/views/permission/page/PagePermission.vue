<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header><span>页面权限演示</span></template>

      <el-alert
        title="页面级权限说明"
        description="页面权限通过路由 meta.permission 字段配置。当用户不具备对应权限时，该菜单项在侧边栏中不会显示，用户无法访问该页面。"
        type="info"
        show-icon
        :closable="false"
        style="margin-bottom: 20px;"
      />

      <h4>当前用户角色：</h4>
      <div style="margin-bottom: 20px;">
        <el-tag v-for="role in roles" :key="role" type="success" style="margin: 4px;">{{ role }}</el-tag>
      </div>

      <h4>页面访问控制示例：</h4>
      <el-table :data="pagePermissionData" border>
        <el-table-column prop="page" label="页面" width="200" />
        <el-table-column prop="permission" label="所需权限" width="200" />
        <el-table-column prop="accessible" label="当前用户可访问" width="140">
          <template #default="{ row }">
            <el-tag :type="row.accessible ? 'success' : 'danger'" size="small">
              {{ row.accessible ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useUserStore } from '@/store/modules/user'

const userStore = useUserStore()
const roles = computed(() => userStore.roles || [])
const permissions = computed(() => userStore.permissions || [])

const pagePermissionData = computed(() => [
  { page: '用户管理', permission: 'user:user:list', accessible: permissions.value.includes('user:user:list') },
  { page: '角色管理', permission: 'user:role:list', accessible: permissions.value.includes('user:role:list') },
  { page: '菜单管理', permission: 'user:menu:list', accessible: permissions.value.includes('user:menu:list') },
  { page: '部门管理', permission: 'system:dept:list', accessible: permissions.value.includes('system:dept:list') },
  { page: '岗位管理', permission: 'system:post:list', accessible: permissions.value.includes('system:post:list') }
])
</script>

<style scoped>
.page-container { padding: 16px; }
</style>
