import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig(({ mode }) => {
  return {
    plugins: [vue()],
    resolve: {
      alias: { '@': resolve(__dirname, 'src') }
    },
    build: {
      // 统一输出到 dist/，由运行时 isMobile() 决定 PC/H5 布局
      outDir: 'dist',
      rollupOptions: {
        input: resolve(__dirname, 'index.html')
      }
    },
    // 不再需要 __IS_H5__ 编译期常量，布局切换改为运行时检测
    server: {
      port: 3000,
      proxy: {
        '/api':    { target: 'http://localhost:8000', changeOrigin: true },
        '/auth':   { target: 'http://localhost:8000', changeOrigin: true },
        '/setup':  { target: 'http://localhost:8000', changeOrigin: true },
        '/static': { target: 'http://localhost:8000', changeOrigin: true }
      }
    }
  }
})
