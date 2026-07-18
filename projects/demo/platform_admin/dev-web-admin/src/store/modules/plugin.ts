/**
 * 插件 Store — 管理已加载的插件状态
 */
import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { PluginConfig, PluginMenu } from '@/plugin-sdk'

export interface PluginInfo {
  name: string
  version: string
  displayName?: string
  description?: string
  status: number
  frontendPath?: string
}

export const usePluginStore = defineStore('plugin', () => {
  /** 所有插件列表（从后端获取） */
  const plugins = ref<PluginInfo[]>([])

  /** 已加载的插件模块 */
  const loadedModules = ref<Record<string, PluginConfig>>({})

  /** 插件菜单（合并后供侧边栏使用） */
  const pluginMenus = ref<PluginMenu[]>([])

  function setPlugins(list: PluginInfo[]) {
    plugins.value = list
  }

  function addLoadedModule(name: string, config: PluginConfig) {
    loadedModules.value[name] = config
    if (config.menus) {
      pluginMenus.value = [...pluginMenus.value, ...config.menus]
    }
  }

  function removeLoadedModule(name: string) {
    const module = loadedModules.value[name]
    if (module?.menus) {
      pluginMenus.value = pluginMenus.value.filter(
        m => !module.menus!.includes(m)
      )
    }
    delete loadedModules.value[name]
  }

  function clearAll() {
    plugins.value = []
    loadedModules.value = {}
    pluginMenus.value = []
  }

  return {
    plugins,
    loadedModules,
    pluginMenus,
    setPlugins,
    addLoadedModule,
    removeLoadedModule,
    clearAll
  }
})
