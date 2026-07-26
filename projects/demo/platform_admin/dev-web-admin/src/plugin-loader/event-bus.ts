import mitt from 'mitt'
import type { EventBus } from '../plugin-sdk/types'

// 使用 mitt 创建事件总线
const emitter = mitt<Record<string, any>>()

export const globalEventBus: EventBus = {
  emit: (event, payload) => emitter.emit(event, payload),
  on: (event, handler) => emitter.on(event, handler),
  off: (event, handler) => emitter.off(event, handler)
}
