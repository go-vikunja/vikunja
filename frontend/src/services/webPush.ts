import {AuthenticatedHTTPFactory, apiV2Url} from '@/helpers/fetcher'

const DEVICE_ID_KEY = 'vikunja-web-push-device-id'

export type WebPushState =
	| 'enabled'
	| 'disabled'
	| 'unsupported'
	| 'permission-denied'
	| 'requires-install'
	| 'configuration-error'

interface WebPushKeys {
	p256dh: string
	auth: string
}

interface WebPushSubscriptionRequest {
	endpoint: string
	expiration_time: number | null
	keys: WebPushKeys
}

function isIOS(): boolean {
	return /iPad|iPhone|iPod/.test(navigator.userAgent)
		|| (navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1)
}

function isStandalone(): boolean {
	return window.matchMedia('(display-mode: standalone)').matches
		|| Boolean((navigator as Navigator & {standalone?: boolean}).standalone)
}

function supportsWebPush(): boolean {
	return window.isSecureContext
		&& 'serviceWorker' in navigator
		&& 'PushManager' in window
		&& 'Notification' in window
}

function getDeviceID(): string {
	let deviceID = localStorage.getItem(DEVICE_ID_KEY)
	if (!deviceID) {
		deviceID = crypto.randomUUID()
		localStorage.setItem(DEVICE_ID_KEY, deviceID)
	}
	return deviceID
}

function urlBase64ToUint8Array(value: string): Uint8Array<ArrayBuffer> {
	const padding = '='.repeat((4 - value.length % 4) % 4)
	const base64 = (value + padding).replace(/-/g, '+').replace(/_/g, '/')
	const bytes = atob(base64)
	return Uint8Array.from(bytes, character => character.charCodeAt(0))
}

function serializeSubscription(subscription: PushSubscription): WebPushSubscriptionRequest {
	const json = subscription.toJSON()
	if (!json.endpoint || !json.keys?.p256dh || !json.keys?.auth) {
		throw new Error('The browser returned an incomplete push subscription.')
	}

	return {
		endpoint: json.endpoint,
		expiration_time: subscription.expirationTime,
		keys: {
			p256dh: json.keys.p256dh,
			auth: json.keys.auth,
		},
	}
}

async function getRegistration(): Promise<ServiceWorkerRegistration> {
	return navigator.serviceWorker.ready
}

async function registerSubscription(subscription: PushSubscription): Promise<void> {
	const HTTP = AuthenticatedHTTPFactory()
	await HTTP.put(
		apiV2Url(`user/settings/web-push/subscriptions/${encodeURIComponent(getDeviceID())}`),
		serializeSubscription(subscription),
	)
}

function isEndpointOwnershipConflict(error: unknown): boolean {
	return typeof error === 'object'
		&& error !== null
		&& 'response' in error
		&& (error as {response?: {status?: number}}).response?.status === 409
}

export function getWebPushAvailability(serverEnabled: boolean, publicKey: string): WebPushState {
	if (!serverEnabled || !publicKey) {
		return 'configuration-error'
	}
	if (!supportsWebPush()) {
		return 'unsupported'
	}
	if (isIOS() && !isStandalone()) {
		return 'requires-install'
	}
	if (Notification.permission === 'denied') {
		return 'permission-denied'
	}
	return 'disabled'
}

export async function getWebPushState(serverEnabled: boolean, publicKey: string): Promise<WebPushState> {
	const availability = getWebPushAvailability(serverEnabled, publicKey)
	if (availability !== 'disabled') {
		return availability
	}

	const registration = await getRegistration()
	return await registration.pushManager.getSubscription() ? 'enabled' : 'disabled'
}

export async function enableWebPush(publicKey: string): Promise<void> {
	const permission = await Notification.requestPermission()
	if (permission !== 'granted') {
		throw new Error('Push notification permission was not granted.')
	}

	const registration = await getRegistration()
	let subscription = await registration.pushManager.getSubscription()
	if (!subscription) {
		subscription = await registration.pushManager.subscribe({
			userVisibleOnly: true,
			applicationServerKey: urlBase64ToUint8Array(publicKey),
		})
	}

	try {
		await registerSubscription(subscription)
	} catch (error) {
		if (!isEndpointOwnershipConflict(error)) {
			throw error
		}

		await subscription.unsubscribe()
		const replacement = await registration.pushManager.subscribe({
			userVisibleOnly: true,
			applicationServerKey: urlBase64ToUint8Array(publicKey),
		})
		await registerSubscription(replacement)
	}
}

export async function reconcileWebPushSubscription(serverEnabled: boolean, publicKey: string): Promise<void> {
	if (getWebPushAvailability(serverEnabled, publicKey) !== 'disabled') {
		return
	}

	const registration = await getRegistration()
	const subscription = await registration.pushManager.getSubscription()
	if (!subscription) {
		return
	}

	try {
		await registerSubscription(subscription)
	} catch (error) {
		if (!isEndpointOwnershipConflict(error)) {
			throw error
		}
		await subscription.unsubscribe()
		const replacement = await registration.pushManager.subscribe({
			userVisibleOnly: true,
			applicationServerKey: urlBase64ToUint8Array(publicKey),
		})
		await registerSubscription(replacement)
	}
}

export async function disableWebPush(): Promise<void> {
	const registration = supportsWebPush() ? await getRegistration() : null
	const subscription = await registration?.pushManager.getSubscription()
	let serverError: unknown

	try {
		const HTTP = AuthenticatedHTTPFactory()
		await HTTP.delete(apiV2Url(`user/settings/web-push/subscriptions/${encodeURIComponent(getDeviceID())}`))
	} catch (error) {
		serverError = error
	} finally {
		await subscription?.unsubscribe()
	}

	if (serverError) {
		throw serverError
	}
}

export async function unsubscribeWebPushLocally(): Promise<void> {
	if (!supportsWebPush()) {
		return
	}
	const registration = await getRegistration()
	await (await registration.pushManager.getSubscription())?.unsubscribe()
}

export async function sendWebPushTest(): Promise<void> {
	const HTTP = AuthenticatedHTTPFactory()
	await HTTP.post(apiV2Url(`user/settings/web-push/subscriptions/${encodeURIComponent(getDeviceID())}/test`))
}
