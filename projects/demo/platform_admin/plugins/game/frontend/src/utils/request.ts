// 获取登录 token（admin 端，key 为 access_token）
export function getToken(): string {
  return localStorage.getItem("access_token") || "";
}

// 通用 JSON 请求
export async function request(method: string, url: string, body?: any) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${getToken()}`
  };
  const opts: RequestInit = { method, headers };
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
    opts.body = JSON.stringify(body);
  }
  const res = await fetch(url, opts);
  return res.json();
}

// 带 Authorization 头的原始 fetch（用于 FormData 上传等）
export async function authFetch(url: string, opts: RequestInit = {}): Promise<Response> {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${getToken()}`,
    ...(opts.headers as Record<string, string>)
  };
  return fetch(url, { ...opts, headers });
}
