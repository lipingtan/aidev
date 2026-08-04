/**
 * 静态路由配置
 * - constantRoutes：不需要布局的路由（登录页、404）
 * - asyncRoutes：需要布局的业务路由（框架搭建阶段静态配置，后续改为动态注册）
 */

import type { RouteRecordRaw } from 'vue-router'

/** 不需要布局的路由 */
export const constantRoutes: RouteRecordRaw[] = [
  {
    path: '/init',
    name: 'system-init',
    component: () => import('@/views/init/InitPage.vue'),
    meta: { title: '系统初始化', requiresAuth: false, hidden: true }
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/login/LoginView.vue'),
    meta: { title: '登录', requiresAuth: false, hidden: true }
  },
  {
    path: '/tenant-select',
    name: 'tenant-select',
    component: () => import('@/views/tenant-select/index.vue'),
    meta: { title: '选择租户', requiresAuth: false, hidden: true }
  },
  {
    path: '/network-error',
    name: 'network-error',
    component: () => import('@/views/error/NetworkError.vue'),
    meta: { title: '网络异常', requiresAuth: false, hidden: true }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/views/error/NotFoundView.vue'),
    meta: { title: '404', requiresAuth: false, hidden: true }
  }
]

/** 需要布局的业务路由（后续改为动态注册） */
export const asyncRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layout/AppLayout.vue'),
    redirect: '/home',
    children: [
      {
        path: 'home',
        name: 'home',
        component: () => import('@/views/home/HomeView.vue'),
        meta: { title: '首页', icon: 'HomeFilled', affix: true }
      },
      // 系统管理
      {
        path: 'system/users',
        name: 'system-users',
        component: () => import('@/views/system/user/UserList.vue'),
        meta: { title: '用户管理', icon: 'User', permission: 'user:user:list' }
      },
      {
        path: 'system/roles',
        name: 'system-roles',
        component: () => import('@/views/system/role/RoleList.vue'),
        meta: { title: '角色管理', icon: 'UserFilled', permission: 'user:role:list' }
      },
      {
        path: 'system/menus',
        name: 'system-menus',
        component: () => import('@/views/system/menu/MenuList.vue'),
        meta: { title: '菜单管理', icon: 'Menu', permission: 'user:menu:list', hidden: true }
      },
      {
        path: 'system/tenants',
        name: 'system-tenants',
        component: () => import('@/views/system/tenant/TenantList.vue'),
        meta: { title: '租户管理', icon: 'Tickets', permission: 'system:tenant:list' }
      },
      {
        path: 'system/applications',
        name: 'system-applications',
        component: () => import('@/views/system/application/ApplicationList.vue'),
        meta: { title: '应用管理', icon: 'Box', permission: 'system:app:list' }
      },
      {
        path: 'system/data-scope',
        name: 'system-data-scope',
        component: () => import('@/views/system/data-scope/DataScopeConfigList.vue'),
        meta: { title: '数据权限配置', icon: 'Key', permission: 'system:data-scope:list' }
      },
      {
        path: 'system/abac',
        name: 'system-abac',
        component: () => import('@/views/system/abac/PolicyList.vue'),
        meta: { title: 'ABAC 策略', icon: 'Lock', permission: 'system:abac:list' }
      },
      {
        path: 'system/org',
        name: 'system-org',
        component: () => import('@/views/system/org/index.vue'),
        meta: { title: '组织架构', icon: 'OfficeBuilding', permission: 'org:tree' }
      },
      // 日志管理
      {
        path: 'log/login-logs',
        name: 'login-logs',
        component: () => import('@/views/system/log/LoginLogList.vue'),
        meta: { title: '登录日志', icon: 'Document', permission: 'user:log:list' }
      },
      {
        path: 'log/operation-logs',
        name: 'operation-logs',
        component: () => import('@/views/system/log/OperationLogList.vue'),
        meta: { title: '操作日志', icon: 'Notebook', permission: 'user:log:list' }
      },
      // 个人中心
      {
        path: 'profile',
        name: 'profile',
        component: () => import('@/views/profile/ProfilePage.vue'),
        meta: { title: '个人中心', icon: 'Avatar', hidden: true }
      },

      // 系统配置
      {
        path: 'system/config',
        name: 'system-config',
        component: () => import('@/views/system/config/ConfigList.vue'),
        meta: { title: '系统配置', icon: 'Setting', permission: 'system:config:list' }
      },
      // 接口管理
      {
        path: 'system/api',
        name: 'system-api',
        component: () => import('@/views/system/api/ApiList.vue'),
        meta: { title: '接口管理', icon: 'Connection', permission: 'system:api:list' }
      },
      // 服务监控
      {
        path: 'monitor/server',
        name: 'monitor-server',
        component: () => import('@/views/monitor/server/ServerMonitor.vue'),
        meta: { title: '服务监控', icon: 'Monitor', permission: 'monitor:server' }
      },

      // 权限演示（hidden，仅供字段对象管理页帮助 Drawer 引用）
      {
        path: 'permission/button',
        name: 'permission-button',
        component: () => import('@/views/permission/button/ButtonPermission.vue'),
        meta: { title: '按钮权限', icon: 'Lock', permission: 'permission:button', hidden: true }
      },
      {
        path: 'permission/page',
        name: 'permission-page',
        component: () => import('@/views/permission/page/PagePermission.vue'),
        meta: { title: '页面权限', icon: 'Key', permission: 'permission:page', hidden: true }
      },
      {
        path: 'permission/field',
        name: 'permission-field',
        component: () => import('@/views/permission/field-permission/index.vue'),
        meta: { title: '字段权限', icon: 'Grid', permission: 'permission:field:list', hidden: true }
      },
      {
        path: 'system/field-objects',
        name: 'system-field-objects',
        component: () => import('@/views/system/field-objects/index.vue'),
        meta: { title: '字段对象管理', icon: 'Grid', permission: 'system:field-object:list' }
      },
      // 插件容器
      {
        path: 'plugin/:pluginName/:pathMatch(.*)*',
        name: 'plugin-container',
        component: () => import('@/views/plugin/PluginContainer.vue'),
        meta: { title: '插件页面', icon: 'Box', hidden: true }
      },

      // 应用目录
      {
        path: 'system/app-catalog',
        name: 'AppCatalog',
        component: () => import('@/views/app-catalog/AppCatalog.vue'),
        meta: { title: '应用目录', icon: 'Grid', permission: 'system:app-catalog:list' }
      },

      // 插件管理（新版）
      {
        path: 'system/plugin',
        name: 'PluginManage',
        component: () => import('@/views/plugin/PluginManage.vue'),
        meta: { title: '插件管理', icon: 'Box', permission: 'system:plugin:list' }
      },
      // 插件管理（旧版，保留兼容）
      {
        path: 'system/plugins',
        name: 'system-plugins',
        component: () => import('@/views/system/plugin/PluginList.vue'),
        meta: { title: '插件管理(旧)', icon: 'Box', permission: 'system:plugin:list', hidden: true }
      },
      {
        path: 'system/tenant-plugins',
        name: 'system-tenant-plugins',
        component: () => import('@/views/system/plugin/TenantPluginList.vue'),
        meta: { title: '租户插件', icon: 'Connection', permission: 'system:plugin:tenant' }
      },

      // C端用户管理
      {
        path: 'system/biz-user',
        name: 'biz-user',
        component: () => import('@/views/biz-user/index.vue'),
        meta: { title: 'C端用户管理', icon: 'Avatar', permission: 'biz:user:list' }
      },

      // 域名管理
      {
        path: 'system/tenant-domain',
        name: 'system-tenant-domain',
        component: () => import('@/views/system/tenant-domain/index.vue'),
        meta: { title: '域名管理', icon: 'Link', permission: 'system:tenant-domain:list' }
      },

      // 审批管理
      {
        path: 'approval',
        name: 'approval',
        component: () => import('@/views/approval/index.vue'),
        meta: { title: '审批管理', icon: 'Stamp' }
      },
      {
        path: 'approval/detail/:id',
        name: 'approval-detail',
        component: () => import('@/views/approval/detail.vue'),
        meta: { title: '审批详情', icon: 'Stamp', hidden: true }
      },
      {
        path: 'approval/flow-config',
        name: 'approval-flow-config',
        component: () => import('@/views/approval/flow-config.vue'),
        meta: { title: '审批流配置', icon: 'SetUp', permission: 'approval:flow:list' }
      },

    ]
  }
]
