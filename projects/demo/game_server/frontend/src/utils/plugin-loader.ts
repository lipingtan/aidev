import { usePluginStoreHook } from "@/store/modules/plugin";
import { getPluginList } from "@/api/plugin";
import { router } from "@/router";
import type { RouteRecordRaw } from "vue-router";

/**
 * 加载所有运行中插件的前端 bundle
 * 在应用初始化时调用（路由守卫中）
 */
export async function loadPlugins(): Promise<void> {
  const pluginStore = usePluginStoreHook();

  try {
    const res = await getPluginList();
    if (res?.code !== 200 || !res?.data) return;

    const plugins = res.data;
    pluginStore.setPlugins(plugins);

    // 只加载运行中（status=1）且有前端路径的插件
    const runningPlugins = plugins.filter(
      (p: any) => p.status === 1 && p.frontendPath
    );

    await Promise.allSettled(
      runningPlugins.map((p: any) => loadSinglePlugin(p.name))
    );
  } catch (err) {
    console.error("[PluginLoader] 获取插件列表失败:", err);
  }
}

/**
 * 加载单个插件的前端 bundle
 */
async function loadSinglePlugin(name: string): Promise<void> {
  const pluginStore = usePluginStoreHook();

  try {
    // 使用 dynamic import 加载插件 bundle
    const url = `/static/plugins/${name}/index.js`;
    const module = await import(/* @vite-ignore */ url);

    const pluginModule = {
      routes: module.routes || [],
      menus: module.menus || []
    };

    // 注册路由到 Vue Router
    if (pluginModule.routes.length > 0) {
      pluginModule.routes.forEach((route: RouteRecordRaw) => {
        router.addRoute(route);
      });
    }

    // 存储到 store
    pluginStore.addLoadedModule(name, pluginModule);

    console.log(`[PluginLoader] 插件 ${name} 前端加载成功`);
  } catch (err) {
    // 加载失败不阻塞 Host 前端
    console.error(`[PluginLoader] 插件 ${name} 前端加载失败:`, err);
  }
}

/**
 * 卸载单个插件的前端（移除路由和菜单）
 */
export function unloadPlugin(name: string): void {
  const pluginStore = usePluginStoreHook();
  const module = pluginStore.loadedModules[name];

  if (module?.routes) {
    module.routes.forEach((route: RouteRecordRaw) => {
      if (route.name) {
        router.removeRoute(route.name);
      }
    });
  }

  pluginStore.removeLoadedModule(name);
  console.log(`[PluginLoader] 插件 ${name} 前端已卸载`);
}
