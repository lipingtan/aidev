const Layout = () => import("@/layout/index.vue");

export default {
  path: "/plugin",
  name: "PluginManage",
  component: Layout,
  redirect: "/plugin/list",
  meta: {
    icon: "ep/box",
    title: "插件管理",
    rank: 90
  },
  children: [
    {
      path: "/plugin/list",
      name: "PluginList",
      component: () => import("@/views/admin/plugin/index.vue"),
      meta: {
        title: "插件列表"
      }
    }
  ]
} satisfies RouteConfigsTable;
