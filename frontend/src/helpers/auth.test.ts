import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest'

import {refreshToken, removeToken} from './auth'

// Count how many times the refresh endpoint is actually POSTed. The whole point
// of the in-flight dedup is that concurrent refreshToken() calls share a single
// underlying POST, independent of the Web Locks API.
let postCallCount = 0
let resolvePost: ((value: unknown) => void) | null = null

vi.mock('@/helpers/fetcher', () => ({
	HTTPFactory: () => ({
		post: vi.fn(() => {
			postCallCount++
			return new Promise((resolve) => {
				resolvePost = resolve
			})
		}),
	}),
}))

vi.mock('@/helpers/desktopAuth', () => ({
	isDesktopApp: () => false,
	refreshDesktopToken: vi.fn(),
}))

const FAKE_TOKEN = 'header.payload.signature'

function settlePost() {
	resolvePost?.({data: {token: FAKE_TOKEN}})
}

describe('refreshToken in-flight dedup', () => {
	const originalLocks = navigator.locks

	beforeEach(() => {
		postCallCount = 0
		resolvePost = null
		removeToken()
		localStorage.clear()
	})

	afterEach(() => {
		Object.defineProperty(navigator, 'locks', {
			value: originalLocks,
			configurable: true,
			writable: true,
		})
	})

	it('coalesces concurrent calls into a single POST when Web Locks is available', async () => {
		const p1 = refreshToken(true)
		const p2 = refreshToken(true)

		// Both calls share one underlying request.
		expect(postCallCount).toBe(1)

		settlePost()
		await Promise.all([p1, p2])

		expect(postCallCount).toBe(1)
	})

	it('coalesces concurrent calls into a single POST on insecure HTTP (no Web Locks)', async () => {
		// Simulate an insecure HTTP context where navigator.locks is undefined.
		Object.defineProperty(navigator, 'locks', {
			value: undefined,
			configurable: true,
			writable: true,
		})

		const p1 = refreshToken(true)
		const p2 = refreshToken(true)
		const p3 = refreshToken(true)

		expect(postCallCount).toBe(1)

		settlePost()
		await Promise.all([p1, p2, p3])

		expect(postCallCount).toBe(1)
	})

	it('allows a fresh refresh after the previous one settled', async () => {
		const p1 = refreshToken(true)
		settlePost()
		await p1
		expect(postCallCount).toBe(1)

		// The in-flight promise was reset, so a later refresh runs anew.
		const p2 = refreshToken(true)
		expect(postCallCount).toBe(2)
		settlePost()
		await p2
	})
})
