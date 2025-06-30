import {getFullBaseUrl} from './helpers/getFullBaseUrl'

declare let self: ServiceWorkerGlobalScope & {
	__WB_MANIFEST: any
	__precacheManifest: any
}

declare const workbox: any
declare const clients: any

const fullBaseUrl = getFullBaseUrl()
const workboxVersion = 'v7.3.0'

importScripts(`${fullBaseUrl}workbox-${workboxVersion}/workbox-sw.js`)
;(workbox as any).setConfig({
	modulePathPrefix: `${fullBaseUrl}workbox-${workboxVersion}`,
})

import { precacheAndRoute } from 'workbox-precaching'
precacheAndRoute(self.__WB_MANIFEST)

// Cache assets
;(workbox as any).routing.registerRoute(
	// This regexp matches all files in precache-manifest
	new RegExp('.+\\.(css|json|js|svg|woff2|png|html|txt|wav)$'),
	new (workbox as any).strategies.StaleWhileRevalidate(),
)

// Always send api requests through the network
;(workbox as any).routing.registerRoute(
	new RegExp('api\\/v1\\/.*$'),
	new (workbox as any).strategies.NetworkOnly(),
)

// This code listens for the user's confirmation to update the app.
self.addEventListener('message', (e: any) => {
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
self.addEventListener('notificationclick', function (event: any) {
	const taskId = event.notification.data.taskId
	event.notification.close()

	switch (event.action) {
		case 'show-task':
			clients.openWindow(`${fullBaseUrl}tasks/${taskId}`)
			break
	}
})

;(workbox as any).core.clientsClaim()
// The precaching code provided by Workbox.
self.__precacheManifest = [].concat(self.__precacheManifest || [])
;(workbox as any).precaching.precacheAndRoute(self.__precacheManifest, {})

