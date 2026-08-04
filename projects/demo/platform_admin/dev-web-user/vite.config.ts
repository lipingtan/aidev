import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

/**
 * dev-web-user Vite 配置
 * 已改为响应式单入口架构：
 * - 统一入口 index.html / main.ts
 * - 运行时根据 window.innerWidth < 768 自动切换 H5Layout / PcLayout
 * - 不再需要 dev:h5 / build:h5 双模式
 */
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: { '@': resolve(__dirname, 'src') }
  },
  build: {
    outDir: 'dist',
    rollupOptions: {
      input: resolve(__dirname, 'index.html')
    }
  },
  server: {
    port: 5174,
    proxy: {
      '/api':    { target: 'http://localhost:8000', changeOrigin: true },
      '/auth':   { target: 'http://localhost:8000', changeOrigin: true },
      '/static': { target: 'http://localhost:8000', changeOrigin: true }
    }
  }
})
