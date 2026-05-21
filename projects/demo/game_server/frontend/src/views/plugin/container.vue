<template>
  <div class="plugin-container">
    <component :is="pluginComponent" v-if="pluginComponent" />
    <div v-else-if="loading" style="padding: 40px; text-align: center">
      <el-icon class="is-loading" :size="24"><Loading /></el-icon>
      <p>加载插件页面中...</p>
    </div>
    <div v-else style="padding: 40px; text-align: center">
      <el-empty description="插件页面加载失败或不存在" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, shallowRef, defineAsyncComponent } from "vue";
import { useRoute } from "vue-router";
import { Loading } from "@element-plus/icons-vue";

const route = useRoute();
const loading = ref(true);
const pluginComponent = shallowRef(null);

async function loadPluginPage() {
  loading.value = true;
  pluginComponent.value = null;

  // 从路由 meta 中获取插件名，或从路径解析
  const pluginName = (route.meta?.pluginName as string) || route.path.split("/")[2] || "";
  if (!pluginName) {
    loading.value = false;
    return;
  }

  try {
    const url = `/static/plugins/${pluginName}/index.js`;
    const module = await import(/* @vite-ignore */ url);

    // 插件 bundle 导出的第一个路由的 component
    if (module.routes && module.routes.length > 0) {
      pluginComponent.value = module.routes[0].component;
    } else if (module.default) {
      pluginComponent.value = module.default;
    }
  } catch (err) {
    console.error(`[PluginContainer] 加载插件 ${pluginName} 页面失败:`, err);
  } finally {
    loading.value = false;
  }
}

onMounted(loadPluginPage);
watch(() => route.path, loadPluginPage);
</script>
