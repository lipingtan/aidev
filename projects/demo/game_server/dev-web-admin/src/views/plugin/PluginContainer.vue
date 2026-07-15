<template>
  <div class="page-container">
    <div v-if="loading" v-loading="true" style="min-height: 300px;" />
    <component v-else-if="pluginComponent" :is="pluginComponent" />
    <el-empty v-else :description="error || '插件页面加载失败或不存在'" />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, shallowRef, type Component } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const loading = ref(false)
const error = ref('')
const pluginComponent = shallowRef<Component | null>(null)

/** 从路由中解析插件名称 */
function resolvePluginName(): string {
  // 优先从 meta.pluginName 读取
  if (route.meta?.pluginName) return route.meta.pluginName as string
  // 否则从路径段中解析：/plugin/{pluginName}/...
  const segments = route.path.split('/')
  const pluginIdx = segments.indexOf('plugin')
  if (pluginIdx >= 0 && segments[pluginIdx + 1]) {
    return segments[pluginIdx + 1]
  }
  return ''
}

async function loadPlugin(pluginName: string) {
  if (!pluginName) {
    error.value = '插件页面加载失败或不存在'
    return
  }
  loading.value = true
  error.value = ''
  pluginComponent.value = null
  try {
    const module = await import(/* @vite-ignore */ `/static/plugins/${pluginName}/index.js`)
    if (module.routes && Array.isArray(module.routes)) {
      const currentPath = route.path
      const matched = module.routes.find((r: any) => currentPath.includes(r.path))
      if (matched && matched.component) {
        pluginComponent.value = matched.component
      } else if (module.routes.length > 0 && module.routes[0].component) {
        // 默认取第一个路由组件
        pluginComponent.value = module.routes[0].component
      } else {
        error.value = '插件页面加载失败或不存在'
      }
    } else if (module.default) {
      pluginComponent.value = module.default
    } else {
      error.value = '插件页面加载失败或不存在'
    }
  } catch {
    error.value = '插件页面加载失败或不存在'
  } finally {
    loading.value = false
  }
}

watch(() => route.path, () => {
  const name = resolvePluginName()
  loadPlugin(name)
}, { immediate: true })
</script>

<style scoped>
.page-container { padding: 16px; min-height: 300px; }
</style>
