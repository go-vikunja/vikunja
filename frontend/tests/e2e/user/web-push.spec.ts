import {test, expect} from '../../support/fixtures'

test.describe('Web Push settings', () => {
	test('enables the current browser only after the user clicks Enable', async ({authenticatedPage: page}) => {
		let registeredSubscription: Record<string, unknown> | null = null
		let registeredDeviceID = ''

		await page.addInitScript(() => {
			let subscription: PushSubscription | null = null
			let permissionRequests = 0
			const browserSubscription = {
				endpoint: 'https://push.example.test/device',
				expirationTime: null,
				toJSON: () => ({
					endpoint: 'https://push.example.test/device',
					keys: {p256dh: 'test-p256dh', auth: 'test-auth'},
				}),
				unsubscribe: async () => true,
			} as unknown as PushSubscription

			Object.defineProperty(window, 'Notification', {
				configurable: true,
				value: class FakeNotification {
					static permission: NotificationPermission = 'default'
					static async requestPermission() {
						permissionRequests++
						FakeNotification.permission = 'granted'
						return FakeNotification.permission
					}
				},
			})
			Object.defineProperty(window, 'PushManager', {configurable: true, value: class FakePushManager {}})
			Object.defineProperty(navigator, 'serviceWorker', {
				configurable: true,
				value: {
					addEventListener: () => undefined,
					ready: Promise.resolve({
						pushManager: {
							getSubscription: async () => subscription,
							subscribe: async () => {
								subscription = browserSubscription
								return subscription
							},
						},
					}),
				},
			})
			Object.defineProperty(window, '__webPushPermissionRequests', {
				get: () => permissionRequests,
			})
		})

		await page.route('**/api/v1/info', async route => {
			const response = await route.fetch()
			const body = await response.json() as Record<string, unknown>
			body.web_push_enabled = true
			body.web_push_public_key = 'BExampleVapidPublicKey'
			await route.fulfill({response, json: body})
		})
		await page.route('**/api/v2/user/settings/web-push/subscriptions/*', async route => {
			const request = route.request()
			if (request.method() !== 'PUT') {
				await route.fallback()
				return
			}
			registeredSubscription = request.postDataJSON() as Record<string, unknown>
			registeredDeviceID = new URL(request.url()).pathname.split('/').at(-1) ?? ''
			await route.fulfill({
				status: 200,
				json: {id: 1, device_id: registeredDeviceID, created: new Date().toISOString(), updated: new Date().toISOString()},
			})
		})

		await page.goto('/user/settings/general')
		const setting = page.getByText('Push notifications on this device', {exact: true}).locator('..')
		const enable = setting.locator('button').filter({hasText: /^Enable$/})
		await expect(enable).toBeVisible()
		expect(await page.evaluate(() => (window as typeof window & {__webPushPermissionRequests: number}).__webPushPermissionRequests)).toBe(0)

		await enable.click()

		await expect(setting.locator('button').filter({hasText: 'Send test notification'})).toBeVisible()
		expect(await page.evaluate(() => (window as typeof window & {__webPushPermissionRequests: number}).__webPushPermissionRequests)).toBe(1)
		expect(registeredDeviceID).toMatch(/^[0-9a-f-]{36}$/)
		expect(registeredSubscription).toEqual({
			endpoint: 'https://push.example.test/device',
			expiration_time: null,
			keys: {p256dh: 'test-p256dh', auth: 'test-auth'},
		})
	})
})
