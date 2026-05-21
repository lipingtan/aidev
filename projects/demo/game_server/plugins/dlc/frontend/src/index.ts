import DlcList from "./views/DlcList.vue";

// 插件路由
export const routes = [
  {
    path: "/plugin/dlc",
    name: "PluginDlc",
    component: DlcList,
    meta: {
      title: "DLC管理",
      icon: "ep/box"
    }
  }
];

// 插件菜单（注入到侧边栏）
export const menus = [
  {
    path: "/plugin/dlc",
    name: "PluginDlc",
    meta: {
      title: "DLC管理",
      icon: "ep/box",
      rank: 50
    }
  }
];
