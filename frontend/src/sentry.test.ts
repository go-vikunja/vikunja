import {describe, it, expect, vi, beforeAll, beforeEach} from 'vitest'
import type {App} from 'vue'
import type {Router} from 'vue-router'

const captureMessage = vi.fn()

vi.mock('@sentry/vue', () => ({
	init: vi.fn(),
	addEventProcessor: vi.fn(),
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
		const src = new URL('/assets/missing.png', window.location.href).href
		failImage(src)

		expect(captureMessage).toHaveBeenCalledWith(`Failed to load image: ${src}`, 'warning')
	})

	it('skips an image from another origin', () => {
		// FRONTEND-OSS-263: injected by a browser extension
		failImage('https://news.search.yahoo.co.jp/favicon.ico')

		expect(captureMessage).not.toHaveBeenCalled()
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

	it.each([
		// FRONTEND-OSS-2KK: the generated default avatar, handed to the <img> as a data url
		['data', 'data:image/svg+xml;base64,PHN2ZyB2aWV3Qm94PSIwIDAgMTAwIDEwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIi8+'],
		// FRONTEND-OSS-26S: an avatar or attachment preview blob url
		['blob', 'blob:https://app.vikunja.cloud/47479b89-bed9-427b-a859-a447f21d5034'],
		// FRONTEND-OSS-2FZ: a mail client's inline attachment
		['cid', 'cid:part1.abcdef@example.com'],
		['filesystem', 'filesystem:https://app.vikunja.cloud/temporary/avatar.png'],
	])('skips a %s src, whose bytes never went over the network', (_, src) => {
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

describe('sentry css load errors', () => {
	function failStylesheet(href: string) {
		const link = document.createElement('link')
		// No rel="stylesheet": happy-dom would try to fetch it for real.
		link.setAttribute('href', href)
		document.body.appendChild(link)
		link.dispatchEvent(new Event('error'))
		link.remove()
	}

	it('reports a stylesheet that failed to load', () => {
		const href = new URL('/assets/missing.css', window.location.href).href
		failStylesheet(href)

		expect(captureMessage).toHaveBeenCalledWith(`Failed to load css: ${href}`, 'warning')
	})

	it('skips a stylesheet from another origin', () => {
		failStylesheet('https://example.com/injected.css')

		expect(captureMessage).not.toHaveBeenCalled()
	})

	it('skips a blob href, whose bytes never went over the network', () => {
		failStylesheet('blob:https://app.vikunja.cloud/47479b89-bed9-427b-a859-a447f21d5034')

		expect(captureMessage).not.toHaveBeenCalled()
	})
})
