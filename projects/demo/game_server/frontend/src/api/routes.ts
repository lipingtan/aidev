import { http } from "@/utils/http";

type Result = {
  success: boolean;
  data: Array<any>;
};

/** 将 go-admin 菜单格式转换为 pure-admin 路由格式（树形结构） */
function convertMenuToRoute(menus: any[]): any[] {
  if (!menus || menus.length === 0) return [];
  const result: any[] = [];

  menus
    .filter(m => m.menuType !== "F")
    .forEach(m => {
      const route: any = {
        path: m.path,
        name: m.menuName,
        meta: {
          title: m.title || m.menuName,
          icon: m.icon || "",
          rank: m.sort || 0,
          showLink: m.visible === "0" || m.visible === true,
          roles: ["admin"]
        }
      };
      // 只有菜单类型（C）且 component 不是 Layout 才设置组件路径
      if (m.menuType === "C" && m.component && m.component !== "Layout") {
        route.component = m.component.replace(/^\//, "");
      }
      if (m.children && m.children.length > 0) {
        const children = convertMenuToRoute(m.children);
        if (children.length > 0) {
          // 父级 path 为空时，子菜单直接提升到当前层级
          if (!m.path) {
            result.push(...children);
            return;
          }
          route.children = children;
          // 目录类型：给第一个子菜单加 showParent，强制显示父级目录
          if (m.menuType === "M" && children[0]?.meta) {
            children[0].meta.showParent = true;
          }
        }
      }
      if (m.path) result.push(route);
    });

  return result;
}

/** 获取异步路由（对接 go-admin /api/v1/menurole） */
export const getAsyncRoutes = async (): Promise<Result> => {
  try {
    const res: any = await http.request<any>("get", "/api/v1/menurole");
    if (res?.code === 200 && res?.data) {
      const routes = convertMenuToRoute(res.data);
      console.log("[routes] converted:", JSON.stringify(routes, null, 2));
      return { success: true, data: routes };
    }
  } catch (e) {
    console.warn("获取菜单失败，使用空路由", e);
  }
  return { success: true, data: [] };
};
