import {getFullBaseUrl} from './helpers/getFullBaseUrl'

declare let self: ServiceWorkerGlobalScope & {
	__precacheManifest?: Parameters<typeof precacheAndRoute>[0],
}
declare const workbox: {
	core: typeof import('workbox-core'),
	routing: typeof import('workbox-routing'),
	strategies: typeof import('workbox-strategies'),
	precaching: typeof import('workbox-precaching'),
	setConfig(options: {modulePathPrefix: string}): void,
}
declare const __WORKBOX_VERSION__: string

const fullBaseUrl = getFullBaseUrl()
const workboxVersion = __WORKBOX_VERSION__

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

// Construct pattern with full base URL
const apiRoutePattern = new RegExp(`${fullBaseUrl.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}api\\/v1\\/.*$`)
// Always send api requests through the network and bypass the browser's HTTP cache
workbox.routing.registerRoute(
	apiRoutePattern,
	new workbox.strategies.NetworkOnly({
		fetchOptions: {
			cache: 'no-store',
		},
	}),
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
			self.clients.openWindow(`${fullBaseUrl}tasks/${taskId}`)
			break
	}
})

workbox.core.clientsClaim()
// The precaching code provided by Workbox.
self.__precacheManifest = [...(self.__precacheManifest ?? [])]
workbox.precaching.precacheAndRoute(self.__precacheManifest, {})

