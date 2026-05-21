import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

export default defineConfig({
  plugins: [vue()],
  define: {
    "process.env.NODE_ENV": JSON.stringify("production"),
    "process.env": JSON.stringify({})
  },
  build: {
    lib: {
      entry: "src/index.ts",
      name: "DlcPlugin",
      fileName: () => "index.js",
      formats: ["es"]
    },
    rollupOptions: {
      external: ["vue", "element-plus"],
      output: {
        // 将 external 的 import 路径映射为全局变量访问
        paths: {
          vue: "/plugin-shims/vue.js",
          "element-plus": "/plugin-shims/element-plus.js"
        }
      }
    },
    outDir: "dist"
  }
});
