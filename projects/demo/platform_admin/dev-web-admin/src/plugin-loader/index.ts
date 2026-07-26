import type { PluginConfig, TeardownFn, UnregisterFn } from '../plugin-sdk/types'
import { doRegisterExtension } from './extension-registry'
import { registerPluginRoutes, unregisterPluginRoutes } from './route-manager'
import { globalEventBus } from './event-bus'
import { useUserStore } from '../store/modules/user'
import { useAuthStore } from '../store/modules/auth'

// 扩展点注册表（plugin-name → unregisterFn[]）
const extensionRegistry = new Map<string, UnregisterFn[]>()

// teardown 队列（plugin-name → TeardownFn[]）
const teardownQueues = new Map<string, TeardownFn[]>()

// 已加载插件的 PluginConfig 缓存（用于 V1 兼容）
const loadedConfigs = new Map<string, PluginConfig>()

const TEARDOWN_TIMEOUT_MS = 5000

/**
 * 创建插件上下文
 */
function createPluginContext(
  pluginName: string,
  platform: 'admin' | 'user',
  device: 'pc' | 'h5'
) {
  teardownQueues.set(pluginName, [])
  extensionRegistry.set(pluginName, [])

  // 运行时获取 store 数据（store 在 pinia 初始化后才可调用）
  const getAuthInfo = () => {
    try {
      const userStore = useUserStore()
      const authStore = useAuthStore()

      // 找到当前租户信息
      const currentTenant = authStore.tenants?.find(
        (t: any) => String(t.id) === String(authStore.currentTenantId)
      ) as any ?? {}

      return {
        currentUser: {
          id: Number(userStore.userId) || 0,
          username: userStore.username || '',
          realName: userStore.realName || '',
          roles: userStore.roles || []
        },
        currentTenant: {
          id: Number(currentTenant.id) || 0,
          name: (currentTenant as any).name || ''
        },
        permissions: userStore.permissions || [] as string[]
      }
    } catch {
      return {
        currentUser: { id: 0, username: '', realName: '', roles: [] as string[] },
        currentTenant: { id: 0, name: '' },
        permissions: [] as string[]
      }
    }
  }

  const mergeLocaleMessage = (locale: string, messages: Record<string, any>) => {
    try {
      // 动态导入 i18n 实例（避免循环依赖，运行时获取）
      import('../locales').then(mod => {
        const i18nInstance = mod.default
        if (i18nInstance?.global?.mergeLocaleMessage) {
          i18nInstance.global.mergeLocaleMessage(locale, messages)
        }
      }).catch(() => {
        console.warn('[plugin-sdk] i18n 不可用，跳过 mergeLocale')
      })
    } catch {
      console.warn('[plugin-sdk] i18n 不可用，跳过 mergeLocale')
    }
  }

  return {
    ...getAuthInfo(),
    eventBus: globalEventBus,
    platform,
    device,
    getPlatform: () => platform,
    getDevice: () => device,
    registerExtension: (point: string, component: any, sort?: number): UnregisterFn => {
      const unregister = doRegisterExtension(point, component, sort)
      extensionRegistry.get(pluginName)!.push(unregister)
      return unregister
    },
    onTeardown: (fn: TeardownFn) => {
      teardownQueues.get(pluginName)!.push(fn)
    },
    i18n: {
      mergeLocale: mergeLocaleMessage
    }
  }
}

/**
 * 加载插件 bundle（UMD 格式，通过 <script> 标签注入）
 */
export async function loadPlugin(
  pluginName: string,
  platform: 'admin' | 'user',
  device: 'pc' | 'h5'
): Promise<void> {
  const bundlePath = `/static/plugins/${pluginName}/${platform}-${device}/bundle.js`

  return new Promise((resolve, reject) => {
    // 防重复加载
    if (document.querySelector(`script[data-plugin="${pluginName}"]`)) {
      resolve()
      return
    }

    const script = document.createElement('script')
    script.src = bundlePath
    script.setAttribute('data-plugin', pluginName)

    script.onload = async () => {
      const globalKey = `__PLUGIN_${pluginName.toUpperCase().replace(/-/g, '_')}__`
      const pluginConfig: PluginConfig = (window as any)[globalKey]

      if (!pluginConfig) {
        reject(new Error(`[plugin-loader] 插件 ${pluginName} 未正确注册到 window.${globalKey}`))
        return
      }

      loadedConfigs.set(pluginName, pluginConfig)
      const ctx = createPluginContext(pluginName, platform, device)

      try {
        if (pluginConfig.setup) {
          await pluginConfig.setup(ctx as any)
        }
        if (pluginConfig.routes) {
          registerPluginRoutes(pluginName, pluginConfig.routes)
        }
        resolve()
      } catch (e) {
        reject(e)
      }
    }

    script.onerror = () => {
      reject(new Error(`[plugin-loader] 加载插件 bundle 失败: ${bundlePath}`))
    }

    document.head.appendChild(script)
  })
}

/**
 * 卸载插件：执行 teardown（含 V1 兼容）→ 清理扩展点 → 移除路由
 */
export async function unloadPlugin(pluginName: string): Promise<void> {
  const fns = [...(teardownQueues.get(pluginName) ?? [])]

  // V1 兼容
  const config = loadedConfigs.get(pluginName)
  if (config?.teardown) {
    const ctx = { getPlatform: () => 'admin' as const, getDevice: () => 'pc' as const } as any
    fns.push(() => config.teardown!(ctx))
  }

  const teardownPromise = Promise.all(fns.map(fn => fn()))
  const timeoutPromise = new Promise<void>((_, reject) =>
    setTimeout(() => reject(new Error(`[plugin-loader] ${pluginName} teardown 超时`)), TEARDOWN_TIMEOUT_MS)
  )

  try {
    await Promise.race([teardownPromise, timeoutPromise])
  } catch (e) {
    console.warn(e)
  }

  // 清理扩展点
  const unregisters = extensionRegistry.get(pluginName) ?? []
  unregisters.forEach(fn => fn())
  extensionRegistry.delete(pluginName)
  teardownQueues.delete(pluginName)
  loadedConfigs.delete(pluginName)

  // 移除路由
  unregisterPluginRoutes(pluginName)

  // 移除 script 标签
  const scriptEl = document.querySelector(`script[data-plugin="${pluginName}"]`)
  if (scriptEl) {
    scriptEl.remove()
  }

  // 清理 window 全局变量
  const globalKey = `__PLUGIN_${pluginName.toUpperCase().replace(/-/g, '_')}__`
  delete (window as any)[globalKey]
}
