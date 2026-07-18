<template>
  <div class="plugin-container">
    <van-loading v-if="loading" size="32px" vertical>加载中...</van-loading>
    <component v-else-if="pluginComponent" :is="pluginComponent" />
    <van-empty v-else :description="error || '插件页面加载失败或不存在'" />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, shallowRef, type Component } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const loading = ref(false)
const error = ref('')
const pluginComponent = shallowRef<Component | null>(null)

function resolvePluginName(): string {
  if (route.meta?.pluginName) return route.meta.pluginName as string
  const segments = route.path.split('/')
  const idx = segments.indexOf('plugin')
  return (idx >= 0 && segments[idx + 1]) ? segments[idx + 1] : ''
}

async function loadPlugin(pluginName: string) {
  if (!pluginName) { error.value = '插件页面加载失败或不存在'; return }
  loading.value = true
  error.value = ''
  pluginComponent.value = null
  try {
    const module = await import(/* @vite-ignore */ `/static/plugins/${pluginName}/index.js`)
    if (module.routes?.length > 0 && module.routes[0].component) {
      pluginComponent.value = module.routes[0].component
    } else if (module.default) {
      pluginComponent.value = module.default
    } else {
      error.value = '插件页面加载失败或不存在'
    }
  } catch { error.value = '插件页面加载失败或不存在' }
  finally { loading.value = false }
}

watch(() => route.path, () => loadPlugin(resolvePluginName()), { immediate: true })
</script>

<style scoped>
.plugin-container { padding: 16px; min-height: 200px; }
</style>
