import type { PluginConfig, PluginContext } from './types'
export type * from './types'

export function definePlugin(config: PluginConfig): PluginConfig { return config }
export function usePluginContext(): PluginContext {
  throw new Error('[PluginSDK] usePluginContext 必须在插件组件内部调用')
}
