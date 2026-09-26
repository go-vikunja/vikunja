import {getFullBaseUrl} from './helpers/getFullBaseUrl'
import {
	webPushNotificationTarget,
	type WebPushPayload,
} from './helpers/webPushNotification'

declare let self: ServiceWorkerGlobalScope
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

self.addEventListener('push', event => {
	event.waitUntil((async () => {
		let payload: WebPushPayload
		try {
			payload = event.data?.json() as WebPushPayload
		} catch {
			return
		}
		if (!payload?.title || !payload.body || !payload.tag) {
			return
		}

		// Refresh the in-app notification list for any open, visible tab, and
		// always surface the OS notification too. Task reminders should appear
		// even while a Vikunja tab is focused, so a user watching one project
		// still sees a reminder for another.
		const openClients = await self.clients.matchAll({type: 'window', includeUncontrolled: true})
		openClients
			.filter(client => client.visibilityState === 'visible')
			.forEach(client => client.postMessage({type: 'web-push-received'}))

		await self.registration.showNotification(payload.title, {
			body: payload.body,
			tag: payload.tag,
			icon: `${fullBaseUrl}images/icons/android-chrome-192x192.png`,
			badge: `${fullBaseUrl}images/icons/badge-monochrome.png`,
			data: payload,
		})
	})())
})

self.addEventListener('notificationclick', event => {
	event.notification.close()
	const payload = event.notification.data as WebPushPayload | undefined
	if (!payload) {
		return
	}

	event.waitUntil((async () => {
		const target = webPushNotificationTarget(payload, self.registration.scope, self.location.origin)
		const openClients = await self.clients.matchAll({type: 'window', includeUncontrolled: true})
		const existing = openClients[0]
		if (existing) {
			await existing.navigate(target)
			await existing.focus()
			return
		}
		await self.clients.openWindow(target)
	})())
})

workbox.core.clientsClaim()
// The precaching code provided by Workbox.
self.__precacheManifest = [].concat(self.__precacheManifest || [])
workbox.precaching.precacheAndRoute(self.__precacheManifest, {})
