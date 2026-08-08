import {describe, expect, it} from 'vitest'

import {
	webPushNotificationTarget,
	type WebPushPayload,
} from './webPushNotification'

const payload: WebPushPayload = {
	title: 'Vikunja',
	body: 'A task changed',
	url: '/tasks/42',
	tag: 'stable-tag',
	notification_id: 123,
}

describe('Web Push service worker decisions', () => {
	it('creates a same-origin deep link that carries the notification id', () => {
		expect(webPushNotificationTarget(payload, 'https://example.test/vikunja/', 'https://example.test'))
			.toBe('https://example.test/vikunja/tasks/42?vikunja_notification=123')
	})

	it('refuses cross-origin deep links', () => {
		expect(webPushNotificationTarget({...payload, url: 'https://evil.test/'}, 'https://example.test/', 'https://example.test'))
			.toBe('https://example.test/')
	})
})
