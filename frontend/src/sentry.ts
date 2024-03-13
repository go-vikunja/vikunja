import 'virtual:vite-plugin-sentry/sentry-config'
import type {App} from 'vue'
import type {Router} from 'vue-router'
import {AxiosError} from 'axios'

export default async function setupSentry(app: App, router: Router) {
	const Sentry = await import('@sentry/vue')
	const {Integrations} = await import('@sentry/tracing')

	Sentry.init({
		app,
		dsn: window.SENTRY_DSN ?? '',
		release: import.meta.env.VITE_PLUGIN_SENTRY_CONFIG.release,
		dist: import.meta.env.VITE_PLUGIN_SENTRY_CONFIG.dist,
		integrations: [
			new Integrations.BrowserTracing({
				routingInstrumentation: Sentry.vueRouterInstrumentation(router),
				tracingOrigins: ['localhost', /^\//],
			}),
		],
		tracesSampleRate: 1.0,
		beforeSend(event, hint) {

			if ((typeof hint.originalException?.code !== 'undefined' && 
				typeof hint.originalException?.message !== 'undefined')
			|| hint.originalException instanceof AxiosError) {
				return null
			}

			return event
		},
	})
}
