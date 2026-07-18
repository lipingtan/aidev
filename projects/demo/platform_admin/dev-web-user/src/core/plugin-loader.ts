import type { PluginConfig } from '@/plugin-sdk'

const loadedPlugins = new Map<string, PluginConfig>()

async function loadRuntime(name: string): Promise<PluginConfig> {
  const mod = await import(/* @vite-ignore */ `/static/plugins/${name}/index.js`)
  return mod.default || mod
}

export async function loadAllPlugins(enabledNames: string[]): Promise<PluginConfig[]> {
  const results: PluginConfig[] = []
  for (const name of enabledNames) {
    try {
      if (loadedPlugins.has(name)) {
        results.push(loadedPlugins.get(name)!)
        continue
      }
      const config = await loadRuntime(name)
      loadedPlugins.set(name, config)
      results.push(config)
    } catch (e) {
      console.error(`[PluginLoader] 加载插件 ${name} 失败:`, e)
    }
  }
  return results
}

export function clearPluginCache() { loadedPlugins.clear() }
