import mitt from 'mitt'
import type { EventBus } from '@/plugin-sdk/types'

const emitter = mitt()
export const pluginEventBus: EventBus = {
  emit: (event, payload) => emitter.emit(event, payload),
  on: (event, handler) => emitter.on(event, handler),
  off: (event, handler) => emitter.off(event, handler as any)
}
