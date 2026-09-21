import {describe, it, expect, vi, beforeAll, beforeEach} from 'vitest'
import type {App} from 'vue'
import type {Router} from 'vue-router'

const captureMessage = vi.fn()

vi.mock('@sentry/vue', () => ({
	init: vi.fn(),
	captureMessage,
	makeBrowserOfflineTransport: vi.fn(),
	makeFetchTransport: vi.fn(),
	browserTracingIntegration: vi.fn(),
	replayIntegration: vi.fn(),
}))

import setupSentry from './sentry'

function failImage(src: string | null, parent: HTMLElement = document.body) {
	const img = document.createElement('img')
	if (src !== null) {
		img.setAttribute('src', src)
	}
	parent.appendChild(img)
	img.dispatchEvent(new Event('error'))
	img.remove()
}

beforeAll(async () => {
	await setupSentry({} as App, {} as Router)
})

beforeEach(() => {
	captureMessage.mockClear()
})

describe('sentry image load errors', () => {
	it('reports an image that failed to load', () => {
		failImage('https://example.com/missing.png')

		expect(captureMessage).toHaveBeenCalledWith('Failed to load image: https://example.com/missing.png', 'warning')
	})

	it.each([
		['missing', null],
		['empty', ''],
		['whitespace', ' '],
		['fragment', '#'],
		['page url', window.location.href],
	])('skips a %s src resolving to the page itself', (_, src) => {
		failImage(src)

		expect(captureMessage).not.toHaveBeenCalled()
	})

	it('skips a broken image inside user content', () => {
		const container = document.createElement('div')
		container.setAttribute('data-user-content', '')
		const paragraph = document.createElement('p')
		container.appendChild(paragraph)
		document.body.appendChild(container)

		// <img src="x"> in a task description at /tasks/3434 (FRONTEND-OSS-263)
		failImage('x', paragraph)
		container.remove()

		expect(captureMessage).not.toHaveBeenCalled()
	})
})
