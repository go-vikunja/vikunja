import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest'

import {getToken, refreshToken, removeToken} from './auth'

let resolvePost: ((value: unknown) => void) | null = null

const post = vi.hoisted(() => vi.fn(() => {
	return new Promise((resolve) => {
		resolvePost = resolve
	})
}))

vi.mock('@/helpers/fetcher', () => ({
	apiV2Url: (path: string) => `/api/v2/${path}`,
	HTTPFactory: () => ({post}),
}))

const desktop = vi.hoisted(() => ({
	isDesktop: false,
	refreshDesktopToken: vi.fn(),
}))

vi.mock('@/helpers/desktopAuth', () => ({
	isDesktopApp: () => desktop.isDesktop,
	refreshDesktopToken: desktop.refreshDesktopToken,
}))

const FAKE_TOKEN = 'header.payload.signature'

function settlePost() {
	resolvePost?.({data: {token: FAKE_TOKEN}})
}

describe('refreshToken in-flight dedup', () => {
	const originalLocks = navigator.locks

	beforeEach(() => {
		resolvePost = null
		post.mockClear()
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
		expect(post).toHaveBeenCalledTimes(1)

		settlePost()
		await Promise.all([p1, p2])

		// The Web Locks branch actually ran...
		expect(requestSpy).toHaveBeenCalledWith('vikunja-token-refresh', expect.any(Function))
		// ...and the in-flight dedup still collapsed both calls into one POST.
		expect(post).toHaveBeenCalledTimes(1)
		expect(post).toHaveBeenCalledWith('/api/v2/user/token/refresh')
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

		expect(post).toHaveBeenCalledTimes(1)

		settlePost()
		await Promise.all([p1, p2, p3])

		expect(post).toHaveBeenCalledTimes(1)
	})

	it('allows a fresh refresh after the previous one settled', async () => {
		const p1 = refreshToken(true)
		settlePost()
		await p1
		expect(post).toHaveBeenCalledTimes(1)

		// The in-flight promise was reset, so a later refresh runs anew.
		const p2 = refreshToken(true)
		expect(post).toHaveBeenCalledTimes(2)
		settlePost()
		await p2
	})

	it('does not re-persist the token when logout happens during an in-flight refresh', async () => {
		const p1 = refreshToken(true)
		expect(post).toHaveBeenCalledTimes(1)

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
		expect(post).toHaveBeenCalledTimes(1)
		const resolveA = resolvePost

		// User logs out, which drops the in-flight reference to A.
		removeToken()

		// Refresh B starts; it must claim the in-flight slot.
		const pB = refreshToken(true)
		expect(post).toHaveBeenCalledTimes(2)
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
		expect(post).toHaveBeenCalledTimes(2)

		resolveB?.({data: {token: FAKE_TOKEN}})
		await Promise.all([pB, pB2])
	})
})

describe('refreshToken v1 cookie fallback', () => {
	beforeEach(() => {
		post.mockClear()
		removeToken()
		localStorage.clear()
	})

	it.each([
		['401', {response: {status: 401}}],
		['404 (e.g. misconfigured API_URL)', {response: {status: 404}}],
		['no response (e.g. network/CORS error)', new Error('Network Error')],
	])('retries against v1 when the v2 refresh fails with %s', async (_label, rejection) => {
		post.mockRejectedValueOnce(rejection)
		post.mockResolvedValueOnce({data: {token: FAKE_TOKEN}})

		await refreshToken(true)

		expect(post).toHaveBeenNthCalledWith(1, '/api/v2/user/token/refresh')
		expect(post).toHaveBeenNthCalledWith(2, 'user/token/refresh')
		expect(localStorage.getItem('token')).toBe(FAKE_TOKEN)
	})

	it('does not retry against v1 when the v2 refresh is rate limited (429)', async () => {
		post.mockRejectedValueOnce({response: {status: 429}})

		await expect(refreshToken(true)).rejects.toThrow('Error renewing token')

		expect(post).toHaveBeenCalledTimes(1)
		expect(localStorage.getItem('token')).toBeNull()
	})

	it('does not fall back to v1 when logout happens between the v2 failure and the fallback call', async () => {
		// removeToken() runs synchronously as part of the v2 call rejecting, simulating
		// a logout landing in the gap before the v1 fallback would otherwise fire.
		post.mockImplementationOnce(() => {
			removeToken()
			return Promise.reject({response: {status: 404}})
		})

		await refreshToken(true)

		expect(post).toHaveBeenCalledTimes(1)
		expect(localStorage.getItem('token')).toBeNull()
	})
})

describe('refreshToken in desktop mode', () => {
	const originalLocks = navigator.locks

	let resolveRefresh: ((tokens: unknown) => void) | null = null

	// Runs the lock callback only once the returned release function is called,
	// so a test can act as the other renderer window while we sit in the queue.
	function stubQueuedLocks() {
		let openGate: (() => void) | null = null
		const gate = new Promise<void>((resolve) => {
			openGate = resolve
		})
		Object.defineProperty(navigator, 'locks', {
			value: {
				request: async (_name: string, cb: () => unknown) => {
					await gate
					return cb()
				},
			},
			configurable: true,
			writable: true,
		})
		return () => openGate?.()
	}

	beforeEach(() => {
		resolveRefresh = null
		removeToken()
		localStorage.clear()

		desktop.isDesktop = true
		desktop.refreshDesktopToken.mockReset()
		desktop.refreshDesktopToken.mockImplementation(() => new Promise((resolve) => {
			resolveRefresh = resolve
		}))

		Object.defineProperty(navigator, 'locks', {
			value: {request: (_name: string, cb: () => unknown) => cb()},
			configurable: true,
			writable: true,
		})
	})

	afterEach(() => {
		desktop.isDesktop = false
		Object.defineProperty(navigator, 'locks', {
			value: originalLocks,
			configurable: true,
			writable: true,
		})
	})

	it('coalesces concurrent calls into a single IPC refresh', async () => {
		localStorage.setItem('desktopOAuthRefreshToken', 'refresh-1')

		const p1 = refreshToken(true)
		const p2 = refreshToken(true)

		expect(desktop.refreshDesktopToken).toHaveBeenCalledTimes(1)

		resolveRefresh?.({access_token: FAKE_TOKEN, refresh_token: 'refresh-2'})
		await Promise.all([p1, p2])

		expect(desktop.refreshDesktopToken).toHaveBeenCalledTimes(1)
		expect(localStorage.getItem('token')).toBe(FAKE_TOKEN)
		expect(localStorage.getItem('desktopOAuthRefreshToken')).toBe('refresh-2')
	})

	it('adopts the token another renderer window refreshed instead of spending the rotated one', async () => {
		localStorage.setItem('desktopOAuthRefreshToken', 'refresh-1')
		localStorage.setItem('token', 'old-token')

		const openGate = stubQueuedLocks()

		// This window queues for the lock...
		const p = refreshToken(true)

		// ...while the other window wins it and rotates both tokens.
		localStorage.setItem('desktopOAuthRefreshToken', 'refresh-2')
		localStorage.setItem('token', FAKE_TOKEN)

		openGate()
		await p

		expect(desktop.refreshDesktopToken).not.toHaveBeenCalled()
		expect(getToken()).toBe(FAKE_TOKEN)
		expect(localStorage.getItem('desktopOAuthRefreshToken')).toBe('refresh-2')
	})

	it('does not re-persist tokens when logout happens during an in-flight refresh', async () => {
		localStorage.setItem('desktopOAuthRefreshToken', 'refresh-1')

		const p = refreshToken(true)
		expect(desktop.refreshDesktopToken).toHaveBeenCalledTimes(1)

		removeToken()

		resolveRefresh?.({access_token: FAKE_TOKEN, refresh_token: 'refresh-2'})
		await p

		expect(localStorage.getItem('token')).toBeNull()
		expect(localStorage.getItem('desktopOAuthRefreshToken')).toBeNull()
	})
})
