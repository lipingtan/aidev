import { defineStore } from "pinia";
import { store } from "../utils";

export interface PluginInfo {
  id: number;
  name: string;
  version: string;
  description: string;
  status: number; // 0=已安装 1=运行中 2=已停止 3=异常
  frontendPath: string;
}

export interface PluginModule {
  routes?: any[];
  menus?: any[];
}

export const usePluginStore = defineStore("plugin", {
  state: () => ({
    // 所有插件列表
    plugins: [] as PluginInfo[],
    // 已加载的插件模块
    loadedModules: {} as Record<string, PluginModule>,
    // 插件菜单（合并后注入侧边栏）
    pluginMenus: [] as any[]
  }),
  actions: {
    setPlugins(plugins: PluginInfo[]) {
      this.plugins = plugins;
    },
    addLoadedModule(name: string, module: PluginModule) {
      this.loadedModules[name] = module;
      if (module.menus) {
        this.pluginMenus = [...this.pluginMenus, ...module.menus];
      }
    },
    removeLoadedModule(name: string) {
      const module = this.loadedModules[name];
      if (module?.menus) {
        this.pluginMenus = this.pluginMenus.filter(
          m => !module.menus.includes(m)
        );
      }
      delete this.loadedModules[name];
    }
  }
});

export function usePluginStoreHook() {
  return usePluginStore(store);
}
