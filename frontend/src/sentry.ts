import 'virtual:vite-plugin-sentry/sentry-config'
import type {App} from 'vue'
import type {Router} from 'vue-router'
import {AxiosError} from 'axios'

export default async function setupSentry(app: App, router: Router) {
	const Sentry = await import('@sentry/vue')
	const {Integrations} = await import('@sentry/tracing')

	Sentry.init({
		app,
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		dsn: (window as any).SENTRY_DSN ?? '',
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		release: (import.meta.env as any).VITE_PLUGIN_SENTRY_CONFIG?.release,
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		dist: (import.meta.env as any).VITE_PLUGIN_SENTRY_CONFIG?.dist,
		integrations: [
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			new (Integrations as any).BrowserTracing({
				// eslint-disable-next-line @typescript-eslint/no-explicit-any
				routingInstrumentation: (Sentry as any).vueRouterInstrumentation(router),
				tracingOrigins: ['localhost', /^\//],
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			}) as any,
		],
		tracesSampleRate: 1.0,
		// @ts-expect-error: Sentry event and hint types are complex
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		beforeSend(event: any, hint: any) {

			if ((typeof hint.originalException?.code !== 'undefined' && 
				typeof hint.originalException?.message !== 'undefined')
			|| hint.originalException instanceof AxiosError) {
				return null
			}

			return event
		},
	})
}
