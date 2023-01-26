import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import path from "path";
import loadVersion from "vite-plugin-package-version";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue(), loadVersion()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
      vue: "vue/dist/vue.esm-bundler.js",
    },
  },
});
