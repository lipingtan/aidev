// 获取登录 token
export function getToken(): string {
  try {
    // 优先从 Cookie 获取
    const cookieMatch = document.cookie.match(/authorized-token=([^;]+)/);
    if (cookieMatch) {
      const data = JSON.parse(decodeURIComponent(cookieMatch[1]));
      return data?.accessToken || "";
    }
    // 兜底从 localStorage 获取（pure-admin 使用 responsive- 前缀）
    const stored = localStorage.getItem("responsive-user-info");
    if (stored) {
      const data = JSON.parse(stored);
      return data?.accessToken || "";
    }
  } catch {}
  return "";
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
