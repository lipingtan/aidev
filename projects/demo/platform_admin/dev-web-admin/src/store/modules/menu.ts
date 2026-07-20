/**
 * 菜单状态管理
 * - 从后端 GET /api/v1/common/user-menu?platform=admin 动态加载菜单树
 * - 管理侧边栏菜单渲染数据
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { fetchUserMenu } from '@/api/menu'

/** 菜单树节点 */
export interface MenuItem {
  id: number
  name: string
  path: string
  icon: string
  permission_code?: string
  children?: MenuItem[]
}

export const useMenuStore = defineStore('menu', () => {
  /** 原始菜单树 */
  const menuTree = ref<MenuItem[]>([])
  /** 是否已加载 */
  const loaded = ref(false)
  /** 加载中 */
  const loading = ref(false)

  /** 扁平化菜单列表（用于权限匹配等） */
  const flatMenus = computed(() => {
    const result: MenuItem[] = []
    function flatten(items: MenuItem[]) {
      for (const item of items) {
        result.push(item)
        if (item.children?.length) {
          flatten(item.children)
        }
      }
    }
    flatten(menuTree.value)
    return result
  })

  /** 从后端加载菜单树 */
  async function fetchMenus(): Promise<MenuItem[]> {
    if (loading.value) return menuTree.value
    loading.value = true
    try {
      const tree = await fetchUserMenu()
      menuTree.value = tree
      loaded.value = true
      return tree
    } finally {
      loading.value = false
    }
  }

  /** 重置菜单状态（登出时调用） */
  function resetMenu(): void {
    menuTree.value = []
    loaded.value = false
  }

  return {
    menuTree,
    loaded,
    loading,
    flatMenus,
    fetchMenus,
    resetMenu
  }
})
