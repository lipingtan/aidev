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
      name: "GamePlugin",
      fileName: () => "index.js",
      formats: ["es"]
    },
    rollupOptions: {
      external: ["vue", "element-plus"],
      output: {
        paths: {
          vue: "/plugin-shims/vue.js",
          "element-plus": "/plugin-shims/element-plus.js"
        }
      }
    },
    outDir: "dist"
  }
});
