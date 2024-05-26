import { defineConfig, loadEnv } from "vite";
import solidPlugin from "vite-plugin-solid";
import { resolve } from "node:path";

export default ({ mode }: { mode: string }) => {
  const env = loadEnv(mode, import.meta.url, "");

  return defineConfig({
    plugins: [solidPlugin()],
    resolve: {
      alias: {
        "@": resolve("src"),
        "@config": env.VITE_CONFIG_PATH || resolve("..", "config.json")
      }
    }
  });
};