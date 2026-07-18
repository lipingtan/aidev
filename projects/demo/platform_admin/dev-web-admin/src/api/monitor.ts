/**
 * 服务监控 API（对接 go-admin）
 * 后端路由：/api/v1/server-monitor
 */
import request from '@/utils/request'

export interface ServerMonitorData {
  cpu: { cores: number; usagePercent: number }
  mem: { total: number; used: number; usagePercent: number }
  os: { hostName: string; os: string; ip: string }
  disk: { total: number; used: number; usagePercent: number }
}

/** 获取服务器监控数据 */
export function getServerMonitor() {
  return request.get<ServerMonitorData>('/server-monitor')
}
