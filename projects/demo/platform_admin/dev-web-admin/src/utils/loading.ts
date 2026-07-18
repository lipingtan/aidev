/**
 * 全局 Loading 服务
 * - 引用计数机制，支持多处并发调用
 * - 基于 ElLoading.service 全屏遮罩
 */

import { ElLoading } from 'element-plus'

let loadingInstance: ReturnType<typeof ElLoading.service> | null = null
let count = 0

export function showLoading(text = '加载中...') {
  count++
  if (!loadingInstance) {
    loadingInstance = ElLoading.service({ fullscreen: true, text, background: 'rgba(0,0,0,0.3)' })
  }
}

export function hideLoading() {
  count--
  if (count <= 0) {
    count = 0
    loadingInstance?.close()
    loadingInstance = null
  }
}
