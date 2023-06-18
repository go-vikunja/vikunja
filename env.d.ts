/// <reference types="vite/client" />
/// <reference types="vite-svg-loader" />
/// <reference types="cypress" />
/// <reference types="@histoire/plugin-vue/components" />

declare module 'postcss-focus-within/browser' {
	import focusWithinInit from 'postcss-focus-within/browser'
	export default focusWithinInit
}

declare module 'css-has-pseudo/browser' {
	import cssHasPseudo from 'css-has-pseudo/browser'
	export default cssHasPseudo
}

interface ImportMetaEnv {
	readonly VIKUNJA_API_URL?: string
	readonly VIKUNJA_HTTP_PORT?: number
	readonly VIKUNJA_HTTPS_PORT?: number

	readonly VIKUNJA_SENTRY_ENABLED?: boolean
	readonly VIKUNJA_SENTRY_DSN?: string

	readonly SENTRY_AUTH_TOKEN?: string
	readonly SENTRY_ORG?: string
	readonly SENTRY_PROJECT?: string

	readonly VITE_WORKBOX_DEBUG?: boolean
	readonly VITE_IS_ONLINE: boolean
}

interface ImportMeta {
	readonly env: ImportMetaEnv
}