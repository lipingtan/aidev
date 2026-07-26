// 新增：取消注册函数
export type UnregisterFn = () => void

// 新增：teardown 清理函数
export type TeardownFn = () => void | Promise<void>

// 插件 Manifest 类型
export interface PluginManifest {
  name: string
  version: string
  displayName: string
  description: string
  platforms: ('admin' | 'user' | 'both')[]
  dependencies?: Record<string, string>
  frontendEntry?: string  // V1 兼容保留
  // V2 新增：多端 bundle 配置
  frontends?: Array<{
    platform: 'admin' | 'user'
    device: 'pc' | 'h5'
    entry: string
  }>
}

// 插件路由
export interface PluginRoute {
  path: string
  name: string
  component: () => Promise<any>
  meta: {
    title: string
    icon?: string
    permission?: string
    hidden?: boolean
  }
}

// 插件菜单项
export interface PluginMenu {
  title: string
  icon: string
  path: string
  sort: number
  group?: string
  permission?: string
}

// 插件配置（definePlugin 参数）
export interface PluginConfig {
  manifest: PluginManifest
  routes?: PluginRoute[]
  menus?: PluginMenu[]
  permissions?: string[]
  extensions?: Record<string, any>
  setup?: (ctx: PluginContext) => void | Promise<void>
  // teardown 字段保留用于兼容旧写法，推荐用 ctx.onTeardown()
  teardown?: (ctx: PluginContext) => void | Promise<void>
}

// 用户信息（只读）
export interface UserInfo {
  id: number
  username: string
  realName: string
  roles: string[]
}

// 租户信息（只读）
export interface TenantInfo {
  id: number
  name: string
}

// 事件总线接口
export interface EventBus {
  emit: (event: string, payload?: any) => void
  on: (event: string, handler: (payload: any) => void) => void
  off: (event: string, handler?: (payload: any) => void) => void
}

// 插件上下文
export interface PluginContext {
  currentUser: Readonly<UserInfo>
  currentTenant: Readonly<TenantInfo>
  permissions: Readonly<string[]>
  eventBus: EventBus
  // V1 已有，返回值改为 UnregisterFn
  registerExtension: (point: string, component: any, sort?: number) => UnregisterFn
  // V2 新增
  platform: 'admin' | 'user'
  device: 'pc' | 'h5'
  getPlatform: () => 'admin' | 'user'
  getDevice: () => 'pc' | 'h5'
  onTeardown: (fn: TeardownFn) => void
  i18n: {
    mergeLocale: (locale: string, messages: Record<string, any>) => void
  }
}
