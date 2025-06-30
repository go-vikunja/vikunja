import 'virtual:vite-plugin-sentry/sentry-config'
import type {App} from 'vue'
import type {Router} from 'vue-router'
import {AxiosError} from 'axios'

export default async function setupSentry(app: App, router: Router) {
	const Sentry = await import('@sentry/vue')
	const {Integrations} = await import('@sentry/tracing')

	Sentry.init({
		app,
		dsn: (window as any).SENTRY_DSN ?? '',
		release: (import.meta.env as any).VITE_PLUGIN_SENTRY_CONFIG?.release,
		dist: (import.meta.env as any).VITE_PLUGIN_SENTRY_CONFIG?.dist,
		integrations: [
			new (Integrations as any).BrowserTracing({
				routingInstrumentation: (Sentry as any).vueRouterInstrumentation(router),
				tracingOrigins: ['localhost', /^\//],
			}) as any,
		],
		tracesSampleRate: 1.0,
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
