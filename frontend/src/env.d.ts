/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly CONFIG_PATH: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
