import 'virtual:vite-plugin-sentry/sentry-config'
import type {App} from 'vue'
import type {Router} from 'vue-router'
import {AxiosError} from 'axios'

export default async function setupSentry(app: App, router: Router) {
	const Sentry = await import('@sentry/vue')

	Sentry.init({
		app,
		dsn: window.SENTRY_DSN ?? '',
		release: import.meta.env.VITE_PLUGIN_SENTRY_CONFIG.release,
		dist: import.meta.env.VITE_PLUGIN_SENTRY_CONFIG.dist,
		integrations: [
			Sentry.browserTracingIntegration({
				router,
			}),
		],
		tracesSampleRate: 1.0,
		beforeSend(event, hint) {
			const originalException = hint.originalException as Error & { code?: number; message?: string }

			if ((typeof originalException?.code !== 'undefined' &&
				typeof originalException?.message !== 'undefined')
			|| originalException instanceof AxiosError) {
				return null
			}

			return event
		},
	})
}
