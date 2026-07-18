<!--
  侧边菜单栏
  - 默认 220px，折叠后 64px
  - 优先从 menu store 动态加载菜单树（后端 GET /api/v1/resources/user-menu）
  - 未加载时 fallback 到静态路由
  - 支持递归子菜单渲染
  - 高亮当前路由对应的菜单项
-->
<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/store/modules/app'
import { useMenuStore } from '@/store/modules/menu'
import type { MenuItem } from '@/store/modules/menu'
import { asyncRoutes } from '@/router/static-routes'
import {
  HomeFilled, Setting, User, UserFilled, Tickets, Lock,
  Menu, Document, Notebook, Avatar, Key, Box, Connection, DataAnalysis, Monitor as MonitorIcon
} from '@element-plus/icons-vue'
import type { Component } from 'vue'

const route = useRoute()
const appStore = useAppStore()
const menuStore = useMenuStore()

const iconMap: Record<string, Component> = {
  HomeFilled, Setting, User, UserFilled, Tickets, Lock,
  Menu, Document, Notebook, Avatar, Key, Box, Connection, DataAnalysis, Monitor: MonitorIcon
}

/** 动态菜单树（仅后端数据，无权限则不显示菜单） */
const dynamicMenuTree = computed<MenuItem[]>(() => {
  if (menuStore.loaded) {
    return menuStore.menuTree || []
  }
  return []
})

/** 从静态路由构建菜单数据（fallback） */
function buildMenuFromRoutes(): MenuItem[] {
  const layoutRoute = asyncRoutes[0]
  if (!layoutRoute?.children) return []
  const children = layoutRoute.children.filter((r) => !r.meta?.hidden)

  // 按路径前缀分组
  const systemItems = children.filter(r => r.path?.startsWith('system/'))
  const logItems = children.filter(r => r.path?.startsWith('log/'))
  const monitorItems = children.filter(r => r.path?.startsWith('monitor/'))
  const permissionItems = children.filter(r => r.path?.startsWith('permission/'))
  const otherItems = children.filter(r =>
    !r.path?.startsWith('system/') &&
    !r.path?.startsWith('log/') &&
    !r.path?.startsWith('monitor/') &&
    !r.path?.startsWith('permission/')
  )

  const result: MenuItem[] = []

  // 独立菜单项
  otherItems.forEach(item => {
    result.push({
      id: 0,
      name: (item.meta?.title as string) || '',
      path: '/' + item.path,
      icon: (item.meta?.icon as string) || '',
      permission_code: item.meta?.permission as string
    })
  })

  // 分组子菜单
  if (systemItems.length) {
    result.push({
      id: 0,
      name: '系统管理',
      path: '',
      icon: 'Setting',
      children: systemItems.map(r => ({
        id: 0,
        name: (r.meta?.title as string) || '',
        path: '/' + r.path,
        icon: (r.meta?.icon as string) || '',
        permission_code: r.meta?.permission as string
      }))
    })
  }
  if (logItems.length) {
    result.push({
      id: 0,
      name: '日志管理',
      path: '',
      icon: 'Document',
      children: logItems.map(r => ({
        id: 0,
        name: (r.meta?.title as string) || '',
        path: '/' + r.path,
        icon: (r.meta?.icon as string) || '',
        permission_code: r.meta?.permission as string
      }))
    })
  }
  if (monitorItems.length) {
    result.push({
      id: 0,
      name: '服务监控',
      path: '',
      icon: 'DataAnalysis',
      children: monitorItems.map(r => ({
        id: 0,
        name: (r.meta?.title as string) || '',
        path: '/' + r.path,
        icon: (r.meta?.icon as string) || '',
        permission_code: r.meta?.permission as string
      }))
    })
  }
  if (permissionItems.length) {
    result.push({
      id: 0,
      name: '权限演示',
      path: '',
      icon: 'Lock',
      children: permissionItems.map(r => ({
        id: 0,
        name: (r.meta?.title as string) || '',
        path: '/' + r.path,
        icon: (r.meta?.icon as string) || '',
        permission_code: r.meta?.permission as string
      }))
    })
  }

  return result
}

const activeMenu = computed(() => route.path)

onMounted(() => {
  // 如果菜单未加载过，尝试加载
  if (!menuStore.loaded && !menuStore.loading) {
    menuStore.fetchMenus().catch(() => {
      // 加载失败使用 fallback
    })
  }
})
</script>

<template>
  <el-aside class="app-sidebar" :class="{ 'aurora-local': appStore.themeMode === 'aurora-local' }" :width="appStore.sidebarCollapsed ? '64px' : '220px'">
    <el-menu :default-active="activeMenu" :collapse="appStore.sidebarCollapsed" :collapse-transition="false" router class="app-sidebar__menu">
      <template v-for="item in dynamicMenuTree" :key="item.id || item.path || item.name">
        <!-- 有子菜单：渲染为 sub-menu -->
        <el-sub-menu v-if="item.children && item.children.length" :index="item.path || item.name">
          <template #title>
            <el-icon v-if="item.icon && iconMap[item.icon]"><component :is="iconMap[item.icon]" /></el-icon>
            <span>{{ item.name }}</span>
          </template>
          <el-menu-item v-for="child in item.children" :key="child.id || child.path" :index="child.path">
            <el-icon v-if="child.icon && iconMap[child.icon]"><component :is="iconMap[child.icon]" /></el-icon>
            <template #title>{{ child.name }}</template>
          </el-menu-item>
        </el-sub-menu>
        <!-- 无子菜单：渲染为独立菜单项 -->
        <el-menu-item v-else :index="item.path">
          <el-icon v-if="item.icon && iconMap[item.icon]"><component :is="iconMap[item.icon]" /></el-icon>
          <template #title>{{ item.name }}</template>
        </el-menu-item>
      </template>
    </el-menu>
  </el-aside>
</template>

<style scoped>
.app-sidebar {
  overflow-x: hidden;
  overflow-y: auto;
  background: linear-gradient(180deg, var(--color-surface), var(--color-surface-elevated));
  border-right: 1px solid var(--color-border);
  transition: width 0.3s;
  box-shadow: var(--shadow-sm);
}
.app-sidebar__menu {
  height: 100%;
  border-right: none;
  background: transparent;
}
.app-sidebar__menu:not(.el-menu--collapse) {
  width: 220px;
}

.app-sidebar :deep(.el-menu-item),
.app-sidebar :deep(.el-sub-menu__title) {
  color: var(--color-text-primary);
  font-weight: 600;
}

.app-sidebar :deep(.el-menu-item:hover),
.app-sidebar :deep(.el-sub-menu__title:hover) {
  color: var(--color-primary);
  background: color-mix(in oklab, var(--color-primary) 10%, transparent);
}

.app-sidebar :deep(.el-menu-item.is-active) {
  color: var(--color-primary);
  background: color-mix(in oklab, var(--color-primary) 14%, transparent);
  box-shadow: inset 3px 0 0 var(--color-primary);
}

.app-sidebar.aurora-local {
  background:
    linear-gradient(180deg, rgba(10, 24, 48, 0.96), rgba(8, 16, 34, 0.98)),
    radial-gradient(120% 100% at 20% 0%, rgba(79, 228, 255, 0.16), transparent 65%);
  border-right-color: rgba(82, 204, 255, 0.28);
  box-shadow: 14px 0 28px rgba(2, 14, 36, 0.42);
}

.app-sidebar.aurora-local :deep(.el-menu-item),
.app-sidebar.aurora-local :deep(.el-sub-menu__title) {
  color: #ecf7ff;
  font-weight: 600;
  text-shadow: 0 1px 10px rgba(80, 220, 255, 0.18);
}

.app-sidebar.aurora-local :deep(.el-menu-item .el-icon),
.app-sidebar.aurora-local :deep(.el-sub-menu__title .el-icon),
.app-sidebar.aurora-local :deep(.el-sub-menu__icon-arrow) {
  color: #bceeff;
}

.app-sidebar.aurora-local :deep(.el-menu-item:hover),
.app-sidebar.aurora-local :deep(.el-sub-menu__title:hover) {
  color: #f4fbff;
  background: linear-gradient(90deg, rgba(35, 84, 130, 0.52), rgba(28, 65, 104, 0.34));
}

.app-sidebar.aurora-local :deep(.el-menu-item.is-active) {
  background: linear-gradient(90deg, rgba(22, 68, 112, 0.92), rgba(14, 42, 75, 0.84));
  color: #ffffff;
  box-shadow: inset 3px 0 0 #5ce9ff, 0 0 10px rgba(22, 108, 178, 0.32);
}

.app-sidebar.aurora-local :deep(.el-menu-item.is-active .el-icon) {
  color: #ffffff;
}

.app-sidebar.aurora-local :deep(.el-sub-menu.is-opened > .el-sub-menu__title) {
  color: #ffffff;
  background: linear-gradient(90deg, rgba(18, 58, 98, 0.82), rgba(12, 36, 64, 0.76));
}

.app-sidebar.aurora-local :deep(.el-sub-menu .el-menu-item) {
  background-color: rgba(7, 22, 43, 0.42);
}
</style>
