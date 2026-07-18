<!--
  此组件已废弃，数据权限配置已迁移到 DataScopeTab.vue（RoleList 页面 Tab 内）。
  保留文件避免其他路由引用报错，实际不再使用。
-->
<template>
  <DataScopeTab v-if="visible" :role-id="roleId" />
</template>

<script setup lang="ts">
import { ref } from 'vue'
import DataScopeTab from './DataScopeTab.vue'

const visible = ref(false)
const roleId = ref('')

function open(id: string) {
  roleId.value = id
  visible.value = true
}

defineExpose({ open })
</script>
