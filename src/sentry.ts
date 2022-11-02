import type { App } from 'vue'
import type { Router } from 'vue-router'
import {VERSION} from './version.json'

export default async function setupSentry(app: App, router: Router) {
	const Sentry = await import('@sentry/vue')
	const {Integrations} = await import('@sentry/tracing')

	Sentry.init({
		release: VERSION,
		app,
		dsn: window.SENTRY_DSN,
		integrations: [
			new Integrations.BrowserTracing({
				routingInstrumentation: Sentry.vueRouterInstrumentation(router),
				tracingOrigins: ['localhost', /^\//],
			}),
		],
		tracesSampleRate: 1.0,
	})
}
