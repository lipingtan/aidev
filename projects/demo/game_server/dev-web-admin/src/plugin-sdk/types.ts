// 插件 Manifest 类型
export interface PluginManifest {
  name: string
  version: string
  displayName: string
  description: string
  platforms: ('admin' | 'user' | 'both')[]
  dependencies?: Record<string, string>
  frontendEntry?: string
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
  registerExtension: (point: string, component: any, sort?: number) => void
}
