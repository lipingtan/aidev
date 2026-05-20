import { http } from "@/utils/http";

/** 获取插件列表 */
export function getPluginList() {
  return http.get<any, any>("/api/v1/plugins");
}

/** 安装插件（URL 方式） */
export function installPluginByUrl(data: { name: string; url: string }) {
  return http.post<any, any>("/api/v1/plugins/install", { data });
}

/** 安装插件（文件上传方式） */
export function installPluginByFile(name: string, file: File) {
  const formData = new FormData();
  formData.append("name", name);
  formData.append("file", file);
  return http.post<any, any>("/api/v1/plugins/install", {
    data: formData,
    headers: { "Content-Type": "multipart/form-data" }
  });
}

/** 启动插件 */
export function startPlugin(name: string) {
  return http.post<any, any>(`/api/v1/plugins/${name}/start`);
}

/** 停止插件 */
export function stopPlugin(name: string) {
  return http.post<any, any>(`/api/v1/plugins/${name}/stop`);
}

/** 卸载插件 */
export function uninstallPlugin(name: string, cleanData = false) {
  return http.request<any>("delete", `/api/v1/plugins/${name}`, {
    params: { clean_data: cleanData }
  });
}

/** 健康检查 */
export function healthCheck(name: string) {
  return http.get<any, any>(`/api/v1/plugins/${name}/health`);
}
