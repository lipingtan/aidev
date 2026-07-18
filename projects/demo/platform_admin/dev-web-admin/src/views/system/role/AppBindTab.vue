<template>
  <div class="perm-tab">
    <el-alert title="应用绑定由权限分配时自动管理" type="info" :closable="false" style="margin-bottom: 16px" />
    <div v-loading="loading">
      <el-tag v-for="app in boundApps" :key="app.app_code" style="margin: 4px 8px 4px 0" size="large">
        {{ app.name }}
      </el-tag>
      <el-empty v-if="!loading && boundApps.length === 0" description="该角色尚未绑定任何应用" :image-size="60" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { getRoleApps } from '@/api/role'
import { listApplications } from '@/api/application'

const props = defineProps<{ roleId: string }>()

const loading = ref(false)
const boundApps = ref<{ app_code: string; name: string }[]>([])

async function loadData() {
  loading.value = true
  try {
    const [codes, apps] = await Promise.all([
      getRoleApps(props.roleId),
      listApplications()
    ])
    const appList = Array.isArray(apps) ? apps : (apps as any).list || []
    const codeSet = new Set(Array.isArray(codes) ? codes : [])
    boundApps.value = appList.filter((a: any) => codeSet.has(a.app_code)).map((a: any) => ({ app_code: a.app_code, name: a.name }))
  } finally {
    loading.value = false
  }
}

watch(() => props.roleId, () => loadData(), { immediate: true })
</script>

<style scoped>
.perm-tab { min-height: 120px; }
</style>
