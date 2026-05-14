import { http } from "@/utils/http";

export type UserResult = {
  success: boolean;
  data: {
    avatar: string;
    username: string;
    nickname: string;
    roles: Array<string>;
    permissions: Array<string>;
    accessToken: string;
    refreshToken: string;
    expires: Date;
  };
};

export type RefreshTokenResult = {
  success: boolean;
  data: {
    accessToken: string;
    refreshToken: string;
    expires: Date;
  };
};

/** 登录 — 适配 go-admin 返回格式 */
export const getLogin = async (data?: object): Promise<UserResult> => {
  const res: any = await http.request<any>("post", "/login", { data });
  // go-admin 返回: { code, token, expire, success }
  // pure-admin 期望: { success, data: { accessToken, refreshToken, expires, roles, ... } }
  if (res?.code === 200 || res?.success) {
    // 设置一个很长的过期时间（100年），避免触发 token 刷新逻辑
    const expireDate = new Date(Date.now() + 100 * 365 * 24 * 3600 * 1000);
    return {
      success: true,
      data: {
        avatar: "",
        username: res.username || "admin",
        nickname: res.nickname || "管理员",
        roles: ["admin"],
        permissions: ["*:*:*"],
        accessToken: res.token || res.currentAuthority,
        refreshToken: res.token || res.currentAuthority,
        expires: expireDate
      }
    };
  }
  return { success: false, data: null as any };
};

/** 刷新 token — go-admin 不支持 refresh token，直接返回失败让用户重新登录 */
export const refreshTokenApi = async (_data?: object): Promise<RefreshTokenResult> => {
  // go-admin 的 token 刷新需要有效 token，pure-admin 的刷新机制与之不兼容
  // 返回失败，触发重新登录
  return Promise.reject(new Error("token expired, please re-login"));
};
