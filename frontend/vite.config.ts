import { defineConfig } from "vite";
import solidPlugin from "vite-plugin-solid";
import { resolve } from "path";

export default defineConfig({
  plugins: [solidPlugin()],
  build: {
    target: "esnext",
    polyfillDynamicImport: false
  },
  resolve: {
    alias: {
      "@": resolve("src"),
      "@config": resolve("..", "config.json")
    }
  }
});
