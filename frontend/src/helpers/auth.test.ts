import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest'

import {
	getAuthSessionEpoch,
	getToken,
	getTokenIdentity,
	getTokenType,
	isTokenExpired,
	MAX_RETRY_AFTER_MS,
	onTokenRefreshed,
	refreshToken,
	RefreshTokenError,
	removeToken,
	type RefreshFailure,
} from './auth'

let resolvePost: ((value: unknown) => void) | null = null

const post = vi.hoisted(() => vi.fn(() => {
	return new Promise((resolve) => {
		resolvePost = resolve
	})
}))

const apiUrls = vi.hoisted(() => ({base: '/api/v2'}))

vi.mock('@/helpers/apiUrl', () => ({
	getApiBaseUrl: () => apiUrls.base,
}))

vi.mock('@/client/generated', () => ({authRefreshToken: post}))

const desktop = vi.hoisted(() => ({
	isDesktop: false,
	refreshDesktopToken: vi.fn(),
}))

vi.mock('@/helpers/desktopAuth', () => ({
	isDesktopApp: () => desktop.isDesktop,
	refreshDesktopToken: desktop.refreshDesktopToken,
}))

const FAKE_TOKEN = 'header.payload.signature'

describe('getTokenType', () => {
	it('reads the numeric JWT type claim', () => {
		const payload = btoa(JSON.stringify({type: 2}))

		expect(getTokenType(`header.${payload}.signature`)).toBe(2)
	})

	it('returns null for missing or malformed tokens', () => {
		expect(getTokenType(null)).toBeNull()
		expect(getTokenType('not-a-jwt')).toBeNull()
	})
})

describe('getTokenIdentity', () => {
	it('reads numeric identity claims from a JWT', () => {
		const payload = btoa(JSON.stringify({id: 42, type: 1}))

		expect(getTokenIdentity(`header.${payload}.signature`)).toEqual({id: 42, type: 1})
	})

	it('returns null for incomplete identity claims', () => {
		const payload = btoa(JSON.stringify({type: 1}))

		expect(getTokenIdentity(`header.${payload}.signature`)).toBeNull()
		expect(getTokenIdentity(null)).toBeNull()
	})
})

describe('isTokenExpired', () => {
	function tokenExpiringIn(seconds: number) {
		const payload = btoa(JSON.stringify({exp: Date.now() / 1000 + seconds}))
		return `header.${payload}.signature`
	}

	it('treats a token as valid until its exp', () => {
		expect(isTokenExpired(tokenExpiringIn(1))).toBe(false)
		expect(isTokenExpired(tokenExpiringIn(-1))).toBe(true)
	})

	it('treats a token expiring within the margin as expired', () => {
		expect(isTokenExpired(tokenExpiringIn(1), 5)).toBe(true)
		expect(isTokenExpired(tokenExpiringIn(10), 5)).toBe(false)
	})
})

describe('getAuthSessionEpoch', () => {
	it('advances when tokens are removed', () => {
		const before = getAuthSessionEpoch()

		removeToken()

		expect(getAuthSessionEpoch()).toBe(before + 1)
	})
})

function settlePost() {
	resolvePost?.({data: {token: FAKE_TOKEN}})
}

function rateLimitedResponse(headers: Record<string, string>) {
	return {
		error: {message: 'Too Many Requests'},
		response: new Response(null, {status: 429, headers}),
	}
}

// A fresh module, like a page load: the rate-limit gate deliberately survives removeToken().
async function loadFreshAuth() {
	vi.resetModules()
	return import('./auth')
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
		expect(post).toHaveBeenCalledWith(expect.objectContaining({baseUrl: 'http://localhost:3000/api/v2'}))
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

describe('onTokenRefreshed', () => {
	const listener = vi.fn()
	let unsubscribe: () => void

	beforeEach(() => {
		unsubscribe = onTokenRefreshed(listener)
		post.mockClear()
		listener.mockClear()
		removeToken()
		localStorage.clear()
	})

	afterEach(() => {
		unsubscribe()
	})

	it('notifies once per coalesced refresh after the new token is saved', async () => {
		const seenTokens: (string | null)[] = []
		listener.mockImplementation(() => seenTokens.push(getToken()))
		const p1 = refreshToken(true)
		const p2 = refreshToken(true)
		settlePost()
		await Promise.all([p1, p2])

		expect(listener).toHaveBeenCalledTimes(1)
		expect(seenTokens).toEqual([FAKE_TOKEN])
	})

	it('keeps notifying other listeners when one throws', async () => {
		const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
		const unsubscribeThrowing = onTokenRefreshed(() => {
			throw new Error('boom')
		})
		try {
			const p = refreshToken(true)
			settlePost()
			await p
			await Promise.resolve()

			expect(errorSpy).toHaveBeenCalledTimes(1)
			expect(listener).toHaveBeenCalledTimes(1)
		} finally {
			unsubscribeThrowing()
			errorSpy.mockRestore()
		}
	})

	it('does not notify after unsubscribing', async () => {
		unsubscribe()
		const p = refreshToken(true)
		settlePost()
		await p

		expect(listener).not.toHaveBeenCalled()
	})

	it('does not notify when the refresh fails', async () => {
		post.mockResolvedValueOnce({error: {}, response: new Response(null, {status: 401})})

		await refreshToken(true).catch(() => {})

		expect(listener).not.toHaveBeenCalled()
	})
})

describe('refreshToken across a server switch', () => {
	beforeEach(() => {
		resolvePost = null
		post.mockClear()
		removeToken()
		localStorage.clear()
		apiUrls.base = 'http://first/api/v2'
	})

	afterEach(() => {
		apiUrls.base = '/api/v2'
	})

	it('does not save the token when the user switched servers while the refresh was in flight', async () => {
		const p = refreshToken(true)
		expect(post).toHaveBeenCalledTimes(1)

		apiUrls.base = 'http://second/api/v2'

		settlePost()
		await p

		expect(localStorage.getItem('token')).toBeNull()
		expect(getToken()).toBeNull()
	})
})

describe('refreshToken failure', () => {
	beforeEach(() => {
		post.mockClear()
		removeToken()
		localStorage.clear()
	})

	const bodyErrorCases: [string, Error, RefreshFailure | undefined][] = [
		['a body parse', new SyntaxError('bad json'), undefined],
		['a mid-body network', new TypeError('network'), {kind: 'network'}],
	]

	it.each(bodyErrorCases)('rethrows %s error unchanged instead of stamping the 200 status', async (_, err, failure) => {
		post.mockResolvedValueOnce({error: err, response: new Response(null, {status: 200})})

		const caught = await refreshToken(true).catch((e: unknown) => e)

		expect((caught as Error).cause).toBe(err)
		expect((caught as Partial<RefreshTokenError>).failure).toEqual(failure)
	})

	const statusCases: [string, number, unknown, RefreshFailure, Record<string, string>?][] = [
		['a rejected refresh token', 401, {code: 16002, detail: 'gone'}, {kind: 'rejected', status: 401, code: 16002}],
		['a proxy error page', 400, '<html>400</html>', {kind: 'rejected', status: 400}],
		['a rate limit', 429, {message: 'Too Many Requests'}, {kind: 'rate-limited'}],
		['a rate limit with an HTTP-date Retry-After', 429, {message: 'Too Many Requests'}, {kind: 'rate-limited'}, {'Retry-After': 'Wed, 21 Oct 2015 07:28:00 GMT'}],
		['a server error', 502, {}, {kind: 'server', status: 502}],
	]

	it.each(statusCases)('classifies %s by the response status', async (_, status, body, failure, headers) => {
		post.mockResolvedValueOnce({error: body, response: new Response(null, {status, headers})})

		const caught = await refreshToken(true).catch((e: unknown) => e)

		expect(post).toHaveBeenCalledWith(expect.objectContaining({
			baseUrl: 'http://localhost:3000/api/v2',
			throwOnError: false,
		}))
		expect(caught).toBeInstanceOf(RefreshTokenError)
		expect((caught as RefreshTokenError).failure).toEqual(failure)
		expect((caught as RefreshTokenError).cause).toMatchObject({status})
	})
	describe('Retry-After on a rate limit', () => {
		const NOW = Date.parse('2015-10-21T07:27:00Z')
		let auth: typeof import('./auth')

		beforeEach(async () => {
			vi.useFakeTimers({toFake: ['Date'], now: NOW})
			auth = await loadFreshAuth()
		})

		afterEach(() => {
			vi.useRealTimers()
		})

		async function rateLimitedRetryAt(headers: Record<string, string>) {
			post.mockResolvedValueOnce(rateLimitedResponse(headers))
			const caught = await auth.refreshToken(true).catch((e: unknown) => e)
			expect(caught).toBeInstanceOf(auth.RefreshTokenError)
			const {failure} = caught as RefreshTokenError
			expect(failure.kind).toBe('rate-limited')
			return failure.kind === 'rate-limited' ? failure.retryAt : 'not rate-limited'
		}

		it.each([
			['an hour', '3600'],
			['an overflowing value', '9'.repeat(400)],
		])('caps %s', async (_, value) => {
			expect(await rateLimitedRetryAt({'Retry-After': value})).toBe(NOW + MAX_RETRY_AFTER_MS)
		})
	})
})

describe('refreshToken after a rate limit with Retry-After', () => {
	const NOW = Date.parse('2015-10-21T07:27:00Z')
	let auth: typeof import('./auth')

	beforeEach(async () => {
		vi.useFakeTimers({toFake: ['Date'], now: NOW})
		resolvePost = null
		post.mockClear()
		localStorage.clear()
		auth = await loadFreshAuth()
	})

	afterEach(() => {
		vi.useRealTimers()
		apiUrls.base = '/api/v2'
	})

	it('rethrows the rate limit without a request until Retry-After passed, across a logout', async () => {
		post.mockResolvedValueOnce(rateLimitedResponse({'Retry-After': '30'}))
		const rateLimit = await auth.refreshToken(true).catch((e: unknown) => e)
		expect(rateLimit).toBeInstanceOf(auth.RefreshTokenError)

		vi.setSystemTime(NOW + 29_999)
		auth.removeToken()

		await expect(auth.refreshToken(true)).rejects.toBe(rateLimit)
		expect(post).toHaveBeenCalledOnce()

		vi.setSystemTime(NOW + 30_000)
		post.mockResolvedValueOnce({data: {token: FAKE_TOKEN}})

		await auth.refreshToken(true)

		expect(post).toHaveBeenCalledTimes(2)
		expect(auth.getToken()).toBe(FAKE_TOKEN)
	})

	it('refreshes against another API server before Retry-After passed', async () => {
		apiUrls.base = 'http://first/api/v2'
		post.mockResolvedValueOnce(rateLimitedResponse({'Retry-After': '30'}))
		await auth.refreshToken(true).catch(() => {})

		apiUrls.base = 'http://second/api/v2'
		post.mockResolvedValueOnce({data: {token: FAKE_TOKEN}})

		await auth.refreshToken(true)

		expect(post).toHaveBeenCalledTimes(2)
		expect(auth.getToken()).toBe(FAKE_TOKEN)
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
