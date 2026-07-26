import type { UnregisterFn } from '../plugin-sdk/types'

// 扩展点条目
interface ExtensionEntry {
  component: any
  sort: number
}

// 全局扩展点注册表：point → entries[]
const extensionMap = new Map<string, ExtensionEntry[]>()

/**
 * 注册扩展点组件
 * @returns UnregisterFn - 调用后从注册表移除该组件
 */
export function doRegisterExtension(
  point: string,
  component: any,
  sort = 0
): UnregisterFn {
  if (!extensionMap.has(point)) {
    extensionMap.set(point, [])
  }
  const entry: ExtensionEntry = { component, sort }
  extensionMap.get(point)!.push(entry)

  return () => {
    const list = extensionMap.get(point)
    if (list) {
      const idx = list.indexOf(entry)
      if (idx !== -1) list.splice(idx, 1)
      if (list.length === 0) extensionMap.delete(point)
    }
  }
}

/**
 * 获取扩展点已注册的组件列表（按 sort 排序）
 */
export function getExtensions(point: string): any[] {
  const list = extensionMap.get(point) || []
  return [...list].sort((a, b) => a.sort - b.sort).map(e => e.component)
}

/** 清空指定扩展点（用于测试） */
export function clearExtensionPoint(point: string): void {
  extensionMap.delete(point)
}
