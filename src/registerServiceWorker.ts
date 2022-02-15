/* eslint-disable no-console */

import {register} from 'register-service-worker'

if (import.meta.env.PROD) {
	register('/sw.ts', {
		ready() {
			console.log('App is being served from cache by a service worker.')
		},
		registered() {
			console.log('Service worker has been registered.')
		},
		cached() {
			console.log('Content has been cached for offline use.')
		},
		updatefound() {
			console.log('New content is downloading.')
		},
		updated(registration) {
			console.log('New content is available; please refresh.')
			// Send an event with the updated info
			document.dispatchEvent(
				new CustomEvent('swUpdated', {detail: registration}),
			)
		},
		offline() {
			console.log('No internet connection found. App is running in offline mode.')
		},
		error(error) {
			console.error('Error during service worker registration:', error)
		},
	})
}
