import type {App} from 'vue'
import type {Router} from 'vue-router'
import {AxiosError} from 'axios'
import {VERSION} from './version.json'

export default async function setupSentry(app: App, router: Router) {
	const Sentry = await import('@sentry/vue')

	Sentry.init({
		app,
		dsn: window.SENTRY_DSN ?? '',

		// cache offline errors
		transport: Sentry.makeBrowserOfflineTransport(Sentry.makeFetchTransport),
		integrations: [
			Sentry.browserTracingIntegration({ router }),
			Sentry.replayIntegration(),
		],

		// vue
		trackComponents: true,

		// Set tracesSampleRate to 1.0 to capture 100%
		// of transactions for tracing.
		// We recommend adjusting this value in production
		tracesSampleRate: 1.0,

		// Set `tracePropagationTargets` to control for which URLs trace propagation should be enabled
		tracePropagationTargets: [
			'localhost',
			/^\//,
			// /^https:\/\/yourserver\.io\/api/,
		],

		// Capture Replay for 10% of all sessions,
		// plus for 100% of sessions with an error
		replaysSessionSampleRate: 0.1,
		replaysOnErrorSampleRate: 1.0,


		beforeSend(event, hint) {

			if ((typeof hint.originalException?.code !== 'undefined' && 
				typeof hint.originalException?.message !== 'undefined')
			|| hint.originalException instanceof AxiosError) {
				return null
			}

			return event
		},
	})

	// from https://docs.sentry.io/platforms/javascript/guides/vue/troubleshooting/
	// under "Capturing resource 404s"
	document.body.addEventListener(
		'error',
		(event) => {
			if (!event.target) return
	
			if (event.target.tagName === 'IMG') {
				Sentry.captureMessage(
					`Failed to load image: ${event.target.src}`,
					'warning',
				)
			} else if (event.target.tagName === 'LINK') {
				Sentry.captureMessage(
					`Failed to load css: ${event.target.href}`,
					'warning',
				)
			}
		},
		true, // useCapture - necessary for resource loading errors
	)
}
