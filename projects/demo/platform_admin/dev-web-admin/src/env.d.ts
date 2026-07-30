/// <reference types="vite/client" />

// Vue 单文件组件类型声明
declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<object, object, unknown>
  export default component
}

// Vite define 注入的编译期常量（已废弃，保留声明防止旧引用报错）
// 实际布局切换请使用 @/utils/device.ts 的 isMobile() 运行时检测
declare const __IS_H5__: boolean
