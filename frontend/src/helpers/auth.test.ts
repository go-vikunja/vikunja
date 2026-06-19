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
		// Stub a minimal Web Locks API: happy-dom leaves navigator.locks
		// undefined, so without this the test would silently fall through to
		// the insecure-HTTP branch and never exercise navigator.locks.request.
		const requestSpy = vi.fn((_name: string, cb: () => unknown) => cb())
		Object.defineProperty(navigator, 'locks', {
			value: {request: requestSpy},
			configurable: true,
			writable: true,
		})

		const p1 = refreshToken(true)
		const p2 = refreshToken(true)

		// Both calls share one underlying request.
		expect(postCallCount).toBe(1)

		settlePost()
		await Promise.all([p1, p2])

		// The Web Locks branch actually ran...
		expect(requestSpy).toHaveBeenCalledWith('vikunja-token-refresh', expect.any(Function))
		// ...and the in-flight dedup still collapsed both calls into one POST.
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

	it('does not re-persist the token when logout happens during an in-flight refresh', async () => {
		const p1 = refreshToken(true)
		expect(postCallCount).toBe(1)

		// User logs out while the refresh POST is still in flight.
		removeToken()

		// The in-flight POST resolves afterwards — it must not undo the logout.
		settlePost()
		await p1

		expect(localStorage.getItem('token')).toBeNull()
	})

	it('an older refresh settling does not clobber a newer in-flight one', async () => {
		// Refresh A starts and stays in flight.
		const pA = refreshToken(true)
		expect(postCallCount).toBe(1)
		const resolveA = resolvePost

		// User logs out, which drops the in-flight reference to A.
		removeToken()

		// Refresh B starts; it must claim the in-flight slot.
		const pB = refreshToken(true)
		expect(postCallCount).toBe(2)
		const resolveB = resolvePost

		// A settles after B started. Its cleanup must NOT null the in-flight
		// slot, since that slot now belongs to B. Without the `=== p` guard,
		// A's .finally would clobber B and let a concurrent caller fire a
		// second parallel POST.
		resolveA?.({data: {token: FAKE_TOKEN}})
		await pA

		// A concurrent caller while B is still in flight must dedup to B —
		// no third POST.
		const pB2 = refreshToken(true)
		expect(postCallCount).toBe(2)

		resolveB?.({data: {token: FAKE_TOKEN}})
		await Promise.all([pB, pB2])
	})
})
