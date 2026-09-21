import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'

import {AUTH_TYPES} from '@/constants/auth'

// Issue #4006: exp is set by the server clock; the browser clock may disagree by more than the JWT TTL.

const SERVER_START_MS = Date.UTC(2026, 0, 1, 12)
const TTL_SECONDS = 600
const MINUTE_MS = 60 * 1000

const server = vi.hoisted(() => ({
	nowMs: 0,
	includeIat: true,
	refreshFails: false,
	post: vi.fn(),
}))

function issueToken(): string {
	const iat = Math.floor(server.nowMs / 1000)
	const payload: Record<string, unknown> = {
		id: 1,
		type: AUTH_TYPES.USER,
		username: 'user1',
		exp: iat + TTL_SECONDS,
		sid: 'session-1',
	}
	if (server.includeIat) {
		payload.iat = iat
	}
	return `header.${btoa(JSON.stringify(payload))}.signature`
}

function fakeHttp() {
	return {
		post: server.post,
		get: vi.fn().mockResolvedValue({
			data: {
				id: 1,
				username: 'user1',
				settings: {},
			},
		}),
		interceptors: {
			request: {use: vi.fn()},
			response: {use: vi.fn()},
		},
	}
}

vi.mock('@/helpers/fetcher', () => ({
	apiV2Url: (path: string) => `/api/v2/${path}`,
	getApiV2BaseUrl: () => '/api/v2/',
	HTTPFactory: () => fakeHttp(),
	AuthenticatedHTTPFactory: () => fakeHttp(),
}))

vi.mock('@/client/generated', async (importOriginal) => ({
	...await importOriginal<typeof import('@/client/generated')>(),
	authLogin: () => server.post('login'),
	authLogout: () => server.post('user/logout'),
	authRefreshToken: () => server.post('user/token/refresh'),
	tokenRenew: () => server.post('user/token'),
	userShow: async () => ({
		data: {
			id: 1,
			username: 'user1',
		},
	}),
}))

vi.mock('@/router', () => ({
	default: {push: vi.fn()},
}))

vi.mock('@/client/queryClient', () => ({
	queryClient: {clear: vi.fn()},
}))

vi.mock('@/composables/useWebSocket', () => ({
	useWebSocket: () => ({
		disconnect: vi.fn(),
		connect: vi.fn(),
		closeStaleConnection: vi.fn(),
	}),
}))

vi.mock('@/helpers/redirectToProvider', () => ({
	getRedirectUrlFromCurrentFrontendPath: vi.fn(),
	redirectToProvider: vi.fn(),
	redirectToProviderOnLogout: vi.fn(),
}))

// Fresh module state (in-memory token, pinia) per call, like a page load. localStorage survives.
async function loadApp() {
	vi.resetModules()
	const {createPinia, setActivePinia} = await import('pinia')
	setActivePinia(createPinia())
	const {useAuthStore} = await import('./auth')
	return useAuthStore()
}

function setBrowserSkew(skewMs: number) {
	vi.setSystemTime(server.nowMs + skewMs)
}

function advance(ms: number) {
	server.nowMs += ms
	vi.setSystemTime(Date.now() + ms)
}

function refreshCalls() {
	return server.post.mock.calls.filter(([url]) => String(url).includes('token/refresh'))
}

describe('auth store with a skewed browser clock', () => {
	beforeEach(() => {
		vi.useFakeTimers({toFake: ['Date']})
		localStorage.clear()
		server.nowMs = SERVER_START_MS
		server.includeIat = true
		server.refreshFails = false
		server.post.mockReset().mockImplementation(async (url: string) => {
			if (url.includes('token/refresh') && server.refreshFails) {
				throw {status: 401}
			}
			if (url === 'user/logout') {
				return {data: {}}
			}
			return {data: {token: issueToken()}}
		})
	})

	afterEach(() => {
		vi.useRealTimers()
		vi.restoreAllMocks()
	})

	describe.each([
		['ahead', 20 * MINUTE_MS],
		['behind', -20 * MINUTE_MS],
	])('browser clock 20 minutes %s', (_, skewMs) => {
		beforeEach(() => setBrowserSkew(skewMs))

		it('stays authenticated after login', async () => {
			const store = await loadApp()

			await store.login({
				username: 'user1',
				password: 'pw',
			})

			expect(store.authenticated).toBe(true)
			expect(refreshCalls()).toHaveLength(0)
		})

		it('stays authenticated after a reload', async () => {
			await (await loadApp()).login({
				username: 'user1',
				password: 'pw',
			})
			advance(5000)

			const store = await loadApp()
			await store.checkAuth()

			expect(store.authenticated).toBe(true)
			expect(refreshCalls()).toHaveLength(0)
		})

		it('refreshes once the token expired by server time and stays authenticated', async () => {
			await (await loadApp()).login({
				username: 'user1',
				password: 'pw',
			})
			advance((TTL_SECONDS + 60) * 1000)

			const store = await loadApp()
			await store.checkAuth()

			expect(refreshCalls()).toHaveLength(1)
			expect(store.authenticated).toBe(true)
		})

		it('keeps a session that is valid by server time when a proactive refresh fails', async () => {
			const store = await loadApp()
			await store.login({
				username: 'user1',
				password: 'pw',
			})
			server.refreshFails = true

			await store.renewToken()

			expect(server.post).not.toHaveBeenCalledWith('user/logout')
			expect(store.authenticated).toBe(true)
		})
	})

	it('warns about the clock skew once per session', async () => {
		const warn = vi.spyOn(console, 'warn').mockImplementation(() => {})
		setBrowserSkew(20 * MINUTE_MS)
		const store = await loadApp()

		await store.login({
			username: 'user1',
			password: 'pw',
		})
		await store.renewToken()

		expect(warn.mock.calls.filter(([message]) => String(message).includes('clock'))).toHaveLength(1)
	})

	it('does not warn when the clocks agree', async () => {
		const warn = vi.spyOn(console, 'warn').mockImplementation(() => {})
		setBrowserSkew(0)

		await (await loadApp()).login({
			username: 'user1',
			password: 'pw',
		})

		expect(warn.mock.calls.filter(([message]) => String(message).includes('clock'))).toHaveLength(0)
	})

	it('trusts a token the server just refreshed even without iat', async () => {
		server.includeIat = false
		setBrowserSkew(20 * MINUTE_MS)
		const store = await loadApp()

		await store.login({
			username: 'user1',
			password: 'pw',
		})

		expect(refreshCalls()).toHaveLength(1)
		expect(store.authenticated).toBe(true)
	})

	describe('token without iat', () => {
		beforeEach(() => {
			server.includeIat = false
			setBrowserSkew(0)
		})

		it('judges a stored token by the browser clock', async () => {
			await (await loadApp()).login({
				username: 'user1',
				password: 'pw',
			})
			advance(5000)

			const store = await loadApp()
			await store.checkAuth()

			expect(store.authenticated).toBe(true)
			expect(refreshCalls()).toHaveLength(0)
		})

		it('treats a stored token past exp by the browser clock as expired', async () => {
			await (await loadApp()).login({
				username: 'user1',
				password: 'pw',
			})
			advance((TTL_SECONDS + 60) * 1000)
			server.refreshFails = true

			const store = await loadApp()
			await store.checkAuth()

			expect(refreshCalls().length).toBeGreaterThan(0)
			expect(store.authenticated).toBe(false)
		})
	})
})
