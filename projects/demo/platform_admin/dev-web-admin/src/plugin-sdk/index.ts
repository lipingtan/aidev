import type { PluginConfig, PluginContext } from './types'
export type * from './types'

/** 定义插件入口 */
export function definePlugin(config: PluginConfig): PluginConfig {
  return config
}

/** 获取插件上下文（由框架 provide/inject 注入） */
export function usePluginContext(): PluginContext {
  // 实际注入逻辑在 plugin-loader 中通过 provide 实现
  // 这里提供类型安全的 stub，运行时由 inject 覆盖
  throw new Error('[PluginSDK] usePluginContext 必须在插件组件内部调用')
}
