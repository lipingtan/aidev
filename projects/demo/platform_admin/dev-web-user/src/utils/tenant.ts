/**
 * 域名 → 租户编码解析工具
 *
 * 通过调用后端 /api/v1/public/tenant-domain 接口动态解析当前域名对应的租户。
 * 超时 3s，失败时 fallback 到环境变量 VITE_TENANT_CODE 或 'default'。
 */

/** 默认租户编码（接口失败时的 fallback） */
const DEFAULT_TENANT_CODE = import.meta.env.VITE_TENANT_CODE || 'default'

/** 解析结果缓存（避免重复请求） */
let cachedResult: { tenantCode: string; tenantName: string } | null = null

/**
 * 异步解析当前域名对应的租户编码
 * - 调用 /api/v1/public/tenant-domain?domain={hostname}
 * - 超时 3s，失败 fallback 到 VITE_TENANT_CODE || 'default'
 */
export async function resolveTenantCodeAsync(): Promise<{ tenantCode: string; tenantName: string }> {
  if (cachedResult) return cachedResult

  const hostname = window.location.hostname

  try {
    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), 3000)

    const resp = await fetch(`/api/v1/public/tenant-domain?domain=${encodeURIComponent(hostname)}`, {
      signal: controller.signal
    })
    clearTimeout(timer)

    if (!resp.ok) throw new Error(`HTTP ${resp.status}`)

    const json = await resp.json()
    if (json.code === 0 && json.data) {
      cachedResult = {
        tenantCode: json.data.tenant_code || DEFAULT_TENANT_CODE,
        tenantName: json.data.tenant_name || ''
      }
      return cachedResult
    }
  } catch {
    // 超时或网络异常，使用 fallback
  }

  cachedResult = { tenantCode: DEFAULT_TENANT_CODE, tenantName: '' }
  return cachedResult
}

/**
 * 同步获取租户编码（兼容旧调用方式）
 * 如果缓存已有结果则直接返回，否则返回默认值
 */
export function resolveTenantCode(): string {
  if (cachedResult) return cachedResult.tenantCode
  return DEFAULT_TENANT_CODE
}
