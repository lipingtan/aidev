import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig(({ mode }) => {
  const isH5 = mode === 'h5'
  return {
    plugins: [vue()],
    resolve: {
      alias: { '@': resolve(__dirname, 'src') }
    },
    build: {
      outDir: isH5 ? 'dist-h5' : 'dist-pc',
      rollupOptions: {
        input: isH5
          ? resolve(__dirname, 'index-h5.html')
          : resolve(__dirname, 'index.html')
      }
    },
    define: {
      // 编译期常量，供 App.vue tree-shaking 使用
      __IS_H5__: isH5
    },
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
