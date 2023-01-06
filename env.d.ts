/// <reference types="vite/client" />
/// <reference types="vite-svg-loader" />
/// <reference types="cypress" />
/// <reference types="@histoire/plugin-vue/components" />

interface ImportMetaEnv {
  readonly VITE_IS_ONLINE: boolean
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}