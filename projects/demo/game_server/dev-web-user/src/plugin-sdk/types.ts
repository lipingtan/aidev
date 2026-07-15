export interface PluginManifest {
  name: string
  version: string
  displayName: string
  description: string
  platforms: ('admin' | 'user' | 'both')[]
  dependencies?: Record<string, string>
  frontendEntry?: string
}

export interface PluginRoute {
  path: string
  name: string
  component: () => Promise<any>
  meta: { title: string; icon?: string; permission?: string; hidden?: boolean }
}

export interface PluginMenu {
  title: string
  icon: string
  path: string
  sort: number
  group?: string
  permission?: string
}

export interface PluginConfig {
  manifest: PluginManifest
  routes?: PluginRoute[]
  menus?: PluginMenu[]
  permissions?: string[]
  extensions?: Record<string, any>
  setup?: (ctx: PluginContext) => void | Promise<void>
}

export interface UserInfo { id: number; username: string; realName: string; roles: string[] }
export interface TenantInfo { id: number; name: string }
export interface EventBus {
  emit: (event: string, payload?: any) => void
  on: (event: string, handler: (payload: any) => void) => void
  off: (event: string, handler?: (payload: any) => void) => void
}
export interface PluginContext {
  currentUser: Readonly<UserInfo>
  currentTenant: Readonly<TenantInfo>
  permissions: Readonly<string[]>
  eventBus: EventBus
  registerExtension: (point: string, component: any, sort?: number) => void
}
