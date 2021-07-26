/* eslint-disable no-console */

import {register} from 'register-service-worker'
import {getToken} from './helpers/auth'

if (import.meta.env.PROD) {
	register('/sw.js', {
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

if (navigator && navigator.serviceWorker) {
	navigator.serviceWorker.addEventListener('message', event => {
		// for every message we expect an action field
		// determining operation that we should perform
		const {action} = event.data
		// we use 2nd port provided by the message channel
		const port = event.ports[0]

		if (action === 'getBearerToken') {
			console.debug('Token request from sw')
			port.postMessage({
				authToken: getToken(),
			})
		} else {
			console.error('Unknown event', event)
			port.postMessage({
				error: 'Unknown request',
			})
		}
	})
}

