import type { PluginConfig, PluginMenu } from '@/plugin-sdk'

/** 合并后的菜单项（含来源插件标记） */
export interface MergedMenuItem extends PluginMenu {
  pluginName: string
}

/**
 * 将所有已加载插件的菜单合并为统一列表
 * - 按 sort 字段排序
 * - 按用户权限过滤（无权限不显示）
 */
export function mergePluginMenus(
  plugins: PluginConfig[],
  userPermissions: string[]
): MergedMenuItem[] {
  const menus: MergedMenuItem[] = []

  for (const plugin of plugins) {
    if (!plugin.menus || plugin.menus.length === 0) continue
    for (const menu of plugin.menus) {
      // 权限过滤：如果菜单声明了 permission 但用户没有，则跳过
      if (menu.permission && !userPermissions.includes(menu.permission)) continue
      menus.push({
        ...menu,
        pluginName: plugin.manifest.name
      })
    }
  }

  // 按 sort 升序排列
  return menus.sort((a, b) => a.sort - b.sort)
}

/**
 * 按分组聚合菜单
 */
export function groupPluginMenus(menus: MergedMenuItem[]): Record<string, MergedMenuItem[]> {
  const groups: Record<string, MergedMenuItem[]> = {}
  for (const menu of menus) {
    const group = menu.group || '其他'
    if (!groups[group]) groups[group] = []
    groups[group].push(menu)
  }
  return groups
}
