import {getFullBaseUrl} from './helpers/getFullBaseUrl'

declare let self: ServiceWorkerGlobalScope & {
	__WB_MANIFEST: unknown[]
	__precacheManifest: unknown[]
}

// @ts-expect-error: Workbox is injected globally via importScripts
declare const workbox: {
	setConfig: (config: { modulePathPrefix: string }) => void
	routing: {
		registerRoute: (matcher: RegExp, strategy: unknown) => void
	}
	strategies: {
		StaleWhileRevalidate: new () => unknown
		NetworkOnly: new () => unknown
	}
	core: {
		clientsClaim: () => void
	}
	precaching: {
		precacheAndRoute: (manifest: unknown[], options: Record<string, unknown>) => void
	}
}

// @ts-expect-error: Clients API is part of service worker global scope
declare const clients: {
	openWindow: (url: string) => void
}

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
self.addEventListener('message', (e: MessageEvent) => {
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
self.addEventListener('notificationclick', function (event: NotificationEvent) {
	const taskId = (event.notification.data as { taskId: string }).taskId
	event.notification.close()

	switch (event.action) {
		case 'show-task':
			clients.openWindow(`${fullBaseUrl}tasks/${taskId}`)
			break
	}
})

workbox.core.clientsClaim()
// The precaching code provided by Workbox.
self.__precacheManifest = [].concat(self.__precacheManifest || [])
workbox.precaching.precacheAndRoute(self.__precacheManifest, {})

