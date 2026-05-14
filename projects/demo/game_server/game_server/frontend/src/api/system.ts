import { http } from "@/utils/http";

/** 通用分页参数 */
export interface PageQuery {
  pageIndex?: number;
  pageSize?: number;
  [key: string]: any;
}

/** 通用响应 */
export interface Result<T = any> {
  code: number;
  msg: string;
  data: T;
}

export interface PageResult<T = any> {
  code: number;
  msg: string;
  data: T[];
  count: number;
}

/** 适配 go-admin PageOK 格式：{ code, data: { list, count, pageIndex, pageSize } } */
function adaptPage(res: any): PageResult {
  if (!res) return { code: 500, msg: "", data: [], count: 0 };
  // PageOK 格式
  if (res.data && Array.isArray(res.data.list)) {
    return { code: res.code, msg: res.msg, data: res.data.list, count: res.data.count || 0 };
  }
  // OK 格式（data 直接是数组）
  if (Array.isArray(res.data)) {
    return { code: res.code, msg: res.msg, data: res.data, count: res.data.length };
  }
  return { code: res.code, msg: res.msg, data: [], count: 0 };
}

// ─── 用户管理 ───────────────────────────────────────────
export const getUserList = (params?: PageQuery) =>
  http.request<any>("get", "/api/v1/sys-user", { params }).then(adaptPage);

export const getUserById = (id: number) =>
  http.request<Result>("get", `/api/v1/sys-user/${id}`);

export const createUser = (data: any) =>
  http.request<Result>("post", "/api/v1/sys-user", { data });

export const updateUser = (data: any) =>
  http.request<Result>("put", "/api/v1/sys-user", { data });

export const deleteUser = (data: { ids: number[] }) =>
  http.request<Result>("delete", "/api/v1/sys-user", { data });

export const resetUserPwd = (data: any) =>
  http.request<Result>("put", "/api/v1/user/pwd/reset", { data });

export const updateUserStatus = (data: any) =>
  http.request<Result>("put", "/api/v1/user/status", { data });

// ─── 角色管理 ───────────────────────────────────────────
export const getRoleList = (params?: PageQuery) =>
  http.request<any>("get", "/api/v1/role", { params }).then(adaptPage);

export const getRoleById = (id: number) =>
  http.request<Result>("get", `/api/v1/role/${id}`);

export const createRole = (data: any) =>
  http.request<Result>("post", "/api/v1/role", { data });

export const updateRole = (id: number, data: any) =>
  http.request<Result>("put", `/api/v1/role/${id}`, { data });

export const deleteRole = (data: { ids: number[] }) =>
  http.request<Result>("delete", "/api/v1/role", { data });

export const getRoleMenuTree = (roleId: number) =>
  http.request<Result>("get", `/api/v1/roleMenuTreeselect/${roleId}`);

// ─── 菜单管理 ───────────────────────────────────────────
export const getMenuList = (params?: any) =>
  http.request<Result>("get", "/api/v1/menu", { params });

export const getMenuById = (id: number) =>
  http.request<Result>("get", `/api/v1/menu/${id}`);

export const createMenu = (data: any) =>
  http.request<Result>("post", "/api/v1/menu", { data });

export const updateMenu = (id: number, data: any) =>
  http.request<Result>("put", `/api/v1/menu/${id}`, { data });

export const deleteMenu = (data: { ids: number[] }) =>
  http.request<Result>("delete", "/api/v1/menu", { data });

// ─── 部门管理 ───────────────────────────────────────────
export const getDeptList = (params?: any) =>
  http.request<Result>("get", "/api/v1/dept", { params });

export const getDeptById = (id: number) =>
  http.request<Result>("get", `/api/v1/dept/${id}`);

export const createDept = (data: any) =>
  http.request<Result>("post", "/api/v1/dept", { data });

export const updateDept = (id: number, data: any) =>
  http.request<Result>("put", `/api/v1/dept/${id}`, { data });

export const deleteDept = (id: number) =>
  http.request<Result>("delete", `/api/v1/dept/${id}`);

export const getDeptTree = () =>
  http.request<Result>("get", "/api/v1/deptTree");

// ─── 岗位管理 ───────────────────────────────────────────
export const getPostList = (params?: PageQuery) =>
  http.request<any>("get", "/api/v1/post", { params }).then(adaptPage);

export const createPost = (data: any) =>
  http.request<Result>("post", "/api/v1/post", { data });

export const updatePost = (id: number, data: any) =>
  http.request<Result>("put", `/api/v1/post/${id}`, { data });

export const deletePost = (id: number) =>
  http.request<Result>("delete", `/api/v1/post/${id}`);

// ─── 字典管理 ───────────────────────────────────────────
export const getDictTypeList = (params?: PageQuery) =>
  http.request<any>("get", "/api/v1/dict/type", { params }).then(adaptPage);

export const createDictType = (data: any) =>
  http.request<Result>("post", "/api/v1/dict/type", { data });

export const updateDictType = (id: number, data: any) =>
  http.request<Result>("put", `/api/v1/dict/type/${id}`, { data });

export const deleteDictType = (data: { ids: number[] }) =>
  http.request<Result>("delete", "/api/v1/dict/type", { data });

export const getDictDataList = (params?: PageQuery) =>
  http.request<any>("get", "/api/v1/dict/data", { params }).then(adaptPage);

export const createDictData = (data: any) =>
  http.request<Result>("post", "/api/v1/dict/data", { data });

export const updateDictData = (dictCode: string, data: any) =>
  http.request<Result>("put", `/api/v1/dict/data/${dictCode}`, { data });

export const deleteDictData = (data: { ids: number[] }) =>
  http.request<Result>("delete", "/api/v1/dict/data", { data });

// ─── 系统配置 ───────────────────────────────────────────
export const getConfigList = (params?: PageQuery) =>
  http.request<any>("get", "/api/v1/config", { params }).then(adaptPage);

export const createConfig = (data: any) =>
  http.request<Result>("post", "/api/v1/config", { data });

export const updateConfig = (id: number, data: any) =>
  http.request<Result>("put", `/api/v1/config/${id}`, { data });

export const deleteConfig = (data: { ids: number[] }) =>
  http.request<Result>("delete", "/api/v1/config", { data });

// ─── 接口管理 ───────────────────────────────────────────
export const getApiList = (params?: PageQuery) =>
  http.request<any>("get", "/api/v1/sys-api", { params }).then(adaptPage);

// ─── 登录日志 ───────────────────────────────────────────
export const getLoginLogList = (params?: PageQuery) =>
  http.request<any>("get", "/api/v1/sys-login-log", { params }).then(adaptPage);

export const deleteLoginLog = (data: { ids: number[] }) =>
  http.request<Result>("delete", "/api/v1/sys-login-log", { data });

// ─── 操作日志 ───────────────────────────────────────────
export const getOperaLogList = (params?: PageQuery) =>
  http.request<any>("get", "/api/v1/sys-opera-log", { params }).then(adaptPage);

export const deleteOperaLog = (data: { ids: number[] }) =>
  http.request<Result>("delete", "/api/v1/sys-opera-log", { data });
