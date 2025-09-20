import {getFullBaseUrl} from './helpers/getFullBaseUrl'

declare let self: ServiceWorkerGlobalScope & {
	__WB_MANIFEST: string[]
	__precacheManifest: string[]
	addEventListener: (type: string, listener: (event: Event) => void) => void
	skipWaiting: () => Promise<void>
}

declare const workbox: {
	setConfig: (config: Record<string, unknown>) => void
	routing: {
		registerRoute: (route: unknown, strategy: unknown) => void
	}
	strategies: {
		StaleWhileRevalidate: new () => unknown
		NetworkOnly: new () => unknown
	}
	precaching: {
		precacheAndRoute: (manifest: unknown, options?: unknown) => void
	}
	core: {
		skipWaiting: () => void
		clientsClaim: () => void
	}
}

declare const clients: {
	claim: () => Promise<void>
	openWindow: (url: string) => Promise<WindowClient | null>
}

declare function importScripts(...urls: string[]): void

const fullBaseUrl = getFullBaseUrl()
const workboxVersion = 'v7.3.0'

importScripts(`${fullBaseUrl}workbox-${workboxVersion}/workbox-sw.js`)
workbox.setConfig({
	modulePathPrefix: `${fullBaseUrl}workbox-${workboxVersion}`,
})

import { precacheAndRoute } from 'workbox-precaching'
precacheAndRoute(self.__WB_MANIFEST)

// Cache assets
workbox.routing.registerRoute(
	// This regexp matches all files in precache-manifest
	new RegExp('.+\\.(css|json|js|svg|woff2|png|html|txt|wav)$'),
	new workbox.strategies.StaleWhileRevalidate(),
)

// Always send api requests through the network
workbox.routing.registerRoute(
	new RegExp('api\\/v1\\/.*$'),
	new workbox.strategies.NetworkOnly(),
)

// This code listens for the user's confirmation to update the app.
self.addEventListener('message', (e) => {
	if (!e.data) {
		return
	}

	switch (e.data) {
		case 'skipWaiting':
			self.skipWaiting()
			break
		default:
			// NOOP
			break
	}
})

// Notification action
self.addEventListener('notificationclick', function (event) {
	const taskId = event.notification.data.taskId
	event.notification.close()

	switch (event.action) {
		case 'show-task':
			clients.openWindow(`${fullBaseUrl}tasks/${taskId}`)
			break
	}
})

workbox.core.clientsClaim()
// The precaching code provided by Workbox.
self.__precacheManifest = ([] as string[]).concat(self.__precacheManifest || [])
workbox.precaching.precacheAndRoute(self.__precacheManifest, {})

