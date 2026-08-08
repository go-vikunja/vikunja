import {beforeEach, describe, expect, it, vi} from 'vitest'

const {put, deleteRequest, post} = vi.hoisted(() => ({
	put: vi.fn(),
	deleteRequest: vi.fn(),
	post: vi.fn(),
}))

vi.mock('@/helpers/fetcher', () => ({
	apiV2Url: (path: string) => `/api/v2/${path}`,
	AuthenticatedHTTPFactory: () => ({
		put,
		delete: deleteRequest,
		post,
	}),
}))

import {
	disableWebPush,
	enableWebPush,
	getWebPushAvailability,
	getWebPushState,
	reconcileWebPushSubscription,
} from './webPush'

function subscription(endpoint = 'https://push.example.test/one'): PushSubscription {
	return {
		endpoint,
		expirationTime: 123456789,
		toJSON: () => ({
			endpoint,
			keys: {p256dh: 'public-key', auth: 'auth-key'},
		}),
		unsubscribe: vi.fn().mockResolvedValue(true),
		getKey: vi.fn(),
		options: {} as PushSubscriptionOptions,
	} as unknown as PushSubscription
}

function installBrowserMocks(existing: PushSubscription | null = null) {
	const subscribe = vi.fn().mockResolvedValue(subscription('https://push.example.test/new'))
	const getSubscription = vi.fn().mockResolvedValue(existing)
	Object.defineProperty(window, 'isSecureContext', {configurable: true, value: true})
	Object.defineProperty(window, 'PushManager', {configurable: true, value: class PushManager {}})
	Object.defineProperty(window, 'matchMedia', {
		configurable: true,
		value: vi.fn().mockReturnValue({matches: true}),
	})
	Object.defineProperty(navigator, 'serviceWorker', {
		configurable: true,
		value: {ready: Promise.resolve({pushManager: {getSubscription, subscribe}})},
	})
	Object.defineProperty(navigator, 'userAgent', {configurable: true, value: 'Android Chrome'})
	Object.defineProperty(globalThis, 'Notification', {
		configurable: true,
		value: {permission: 'default', requestPermission: vi.fn().mockResolvedValue('granted')},
	})
	return {getSubscription, subscribe}
}

describe('Web Push browser integration', () => {
	beforeEach(() => {
		localStorage.clear()
		put.mockReset().mockResolvedValue({data: {}})
		deleteRequest.mockReset().mockResolvedValue({data: {}})
		post.mockReset().mockResolvedValue({data: {}})
		installBrowserMocks()
	})

	it('reports server and permission states before offering enable', () => {
		expect(getWebPushAvailability(false, '')).toBe('configuration-error')
		Object.defineProperty(globalThis, 'Notification', {
			configurable: true,
			value: {permission: 'denied', requestPermission: vi.fn()},
		})
		expect(getWebPushAvailability(true, 'key')).toBe('permission-denied')
	})

	it('requires an installed Home Screen app on iOS', () => {
		Object.defineProperty(navigator, 'userAgent', {configurable: true, value: 'iPhone'})
		Object.defineProperty(window, 'matchMedia', {
			configurable: true,
			value: vi.fn().mockReturnValue({matches: false}),
		})
		expect(getWebPushAvailability(true, 'key')).toBe('requires-install')
	})

	it('serializes and registers a subscription only after enable', async () => {
		const {subscribe} = installBrowserMocks()
		await enableWebPush('BElwdWJsaWMta2V5')

		expect(Notification.requestPermission).toHaveBeenCalledOnce()
		expect(subscribe).toHaveBeenCalledWith({
			userVisibleOnly: true,
			applicationServerKey: expect.any(Uint8Array),
		})
		expect(put).toHaveBeenCalledOnce()
		expect(put.mock.calls[0][1]).toEqual({
			endpoint: 'https://push.example.test/new',
			expiration_time: 123456789,
			keys: {p256dh: 'public-key', auth: 'auth-key'},
		})
	})

	it('reconciles an existing subscription without creating one', async () => {
		const existing = subscription()
		const {subscribe} = installBrowserMocks(existing)
		await reconcileWebPushSubscription(true, 'key')

		expect(put).toHaveBeenCalledOnce()
		expect(subscribe).not.toHaveBeenCalled()
		expect(await getWebPushState(true, 'key')).toBe('enabled')
	})

	it('unsubscribes locally even if disabling on the server fails', async () => {
		const existing = subscription()
		installBrowserMocks(existing)
		deleteRequest.mockRejectedValueOnce(new Error('offline'))

		await expect(disableWebPush()).rejects.toThrow('offline')
		expect(existing.unsubscribe).toHaveBeenCalledOnce()
	})
})
