import { http } from "@/utils/http";
import type { PageQuery, Result } from "./system";

/** 适配 go-admin PageOK 格式 */
function adaptPage(res: any) {
  if (!res) return { code: 500, msg: "", data: [], count: 0 };
  if (res.data && Array.isArray(res.data.list)) {
    return { code: res.code, msg: res.msg, data: res.data.list, count: res.data.count || 0 };
  }
  if (Array.isArray(res.data)) {
    return { code: res.code, msg: res.msg, data: res.data, count: res.data.length };
  }
  return { code: res.code, msg: res.msg, data: [], count: 0 };
}

// ─── 游戏管理 ───────────────────────────────────────────
export const getGameList = (params?: PageQuery) =>
  http.request<any>("get", "/api/v1/game", { params }).then(adaptPage);

export const getGameById = (id: number) =>
  http.request<Result>("get", `/api/v1/game/${id}`);

export const createGame = (data: any) =>
  http.request<Result>("post", "/api/v1/game", { data });

export const updateGame = (id: number, data: any) =>
  http.request<Result>("put", `/api/v1/game/${id}`, { data });

export const deleteGame = (id: number) =>
  http.request<Result>("delete", `/api/v1/game/${id}`);

export const regenerateSecret = (id: number) =>
  http.request<Result>("post", `/api/v1/game/${id}/secret`);

// ─── DLC 管理 ───────────────────────────────────────────
export const getDlcList = (params?: PageQuery) =>
  http.request<any>("get", "/api/v1/game-dlc", { params }).then(adaptPage);

export const getDlcById = (id: number) =>
  http.request<Result>("get", `/api/v1/game-dlc/${id}`);

export const createDlc = (data: any) =>
  http.request<Result>("post", "/api/v1/game-dlc", { data });

export const updateDlc = (id: number, data: any) =>
  http.request<Result>("put", `/api/v1/game-dlc/${id}`, { data });

export const deleteDlc = (id: number) =>
  http.request<Result>("delete", `/api/v1/game-dlc/${id}`);

// ─── 玩家管理 ───────────────────────────────────────────
export const getPlayerList = (params?: PageQuery) =>
  http.request<any>("get", "/api/v1/game-player", { params }).then(adaptPage);

export const getPlayerById = (id: number) =>
  http.request<Result>("get", `/api/v1/game-player/${id}`);

export const banPlayer = (id: number, data: any) =>
  http.request<Result>("put", `/api/v1/game-player/${id}/ban`, { data });

// ─── 订单管理 ───────────────────────────────────────────
export const getOrderList = (params?: PageQuery) =>
  http.request<any>("get", "/api/v1/game-order", { params }).then(adaptPage);

export const getOrderById = (id: number) =>
  http.request<Result>("get", `/api/v1/game-order/${id}`);

export const refundOrder = (id: number) =>
  http.request<Result>("post", `/api/v1/game-order/${id}/refund`);

// ─── 支付配置 ───────────────────────────────────────────
export const getPaymentConfigList = (params?: PageQuery) =>
  http.request<any>("get", "/api/v1/game-payment-config", { params }).then(adaptPage);

export const savePaymentConfig = (data: any) =>
  http.request<Result>("post", "/api/v1/game-payment-config", { data });

// ─── H5 页面 ───────────────────────────────────────────
export const getH5PageList = (params?: PageQuery) =>
  http.request<any>("get", "/api/v1/game-h5", { params }).then(adaptPage);

export const getH5PageById = (id: number) =>
  http.request<Result>("get", `/api/v1/game-h5/${id}`);

export const createH5Page = (data: any) =>
  http.request<Result>("post", "/api/v1/game-h5", { data });

export const updateH5Page = (id: number, data: any) =>
  http.request<Result>("put", `/api/v1/game-h5/${id}`, { data });

export const deleteH5Page = (id: number) =>
  http.request<Result>("delete", `/api/v1/game-h5/${id}`);
