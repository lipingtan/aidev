import type { PluginConfig } from '@/plugin-sdk'
import { listEnabledPlugins, type EnabledPluginInfo } from '@/api/plugin'

/** 已加载插件缓存 */
const loadedPlugins = new Map<string, PluginConfig>()

/**
 * 加载源码模式插件
 * 使用 import.meta.glob 扫描 src/plugins/xxx/index.ts
 */
async function loadSource(name: string): Promise<PluginConfig> {
  const modules = import.meta.glob<{ default: PluginConfig }>('/src/plugins/*/index.ts')
  const entry = modules[`/src/plugins/${name}/index.ts`]
  if (!entry) {
    throw new Error(`[PluginLoader] 源码模式插件 ${name} 未找到`)
  }
  const mod = await entry()
  return mod.default
}

/**
 * 加载运行时模式插件
 * 从 /static/plugins/{name}/index.js 动态 import
 * 支持两种 bundle 格式：
 * 1. PluginConfig 格式（有 manifest 字段）
 * 2. 简单 routes 格式（只有 routes 数组，PluginContainer 使用）
 */
async function loadRuntime(name: string): Promise<PluginConfig | null> {
  const mod = await import(/* @vite-ignore */ `/static/plugins/${name}/index.js`)
  const config = mod.default || mod
  // 校验是否是有效的 PluginConfig（需要有 manifest 字段）
  if (config && config.manifest) {
    return config as PluginConfig
  }
  // routes 格式的 bundle 由 PluginContainer 直接加载，不走 PluginConfig 体系
  return null
}

/**
 * 加载单个插件（带缓存）
 */
async function loadPlugin(name: string, mode: 'source' | 'runtime'): Promise<PluginConfig | null> {
  if (loadedPlugins.has(name)) {
    return loadedPlugins.get(name)!
  }
  const config = mode === 'source' ? await loadSource(name) : await loadRuntime(name)
  if (config) {
    loadedPlugins.set(name, config)
  }
  return config
}

/**
 * 批量加载已启用插件
 * 单个插件加载失败不影响其他插件
 */
export async function loadAllPlugins(): Promise<PluginConfig[]> {
  let enabledList: EnabledPluginInfo[] = []
  try {
    enabledList = await listEnabledPlugins()
  } catch (e) {
    console.error('[PluginLoader] 获取已启用插件列表失败:', e)
    return []
  }

  const results: PluginConfig[] = []
  for (const info of enabledList) {
    try {
      const config = await loadPlugin(info.name, info.mode)
      if (config) {
        results.push(config)
      }
    } catch (e) {
      console.error(`[PluginLoader] 加载插件 ${info.name} 失败:`, e)
    }
  }
  return results
}

/** 清除插件缓存（用于热重载） */
export function clearPluginCache() {
  loadedPlugins.clear()
}
