import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {flushPromises} from '@vue/test-utils'

import {AUTH_TYPES} from '@/constants/auth'

// postMock only counts calls into the real refresh; tests stub fetch.
const {postMock, userShowMock, openidCallbackMock, loginMock, routerMock} = vi.hoisted(() => ({
	postMock: vi.fn(),
	userShowMock: vi.fn(),
	openidCallbackMock: vi.fn(),
	loginMock: vi.fn(),
	routerMock: {
		push: vi.fn(),
		currentRoute: {value: {name: 'home'}},
	},
}))

vi.mock('@/client/generated', async (importOriginal) => ({
	...await importOriginal<typeof import('@/client/generated')>(),
	authRefreshToken: postMock,
	userShow: userShowMock,
	authOpenidCallback: openidCallbackMock,
	authLogin: loginMock,
}))

vi.mock('@/router', () => ({
	default: routerMock,
}))

vi.mock('@/client/queryClient', async () => {
	const {QueryClient} = await import('@tanstack/vue-query')
	return {queryClient: new QueryClient({defaultOptions: {queries: {retry: false}}})}
})

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

const STALE_JWT = `header.${btoa(JSON.stringify({
	id: 1,
	type: AUTH_TYPES.USER,
	exp: Math.floor(Date.now() / 1000) - 3600,
}))}.signature`

const FRESH_JWT = `header.${btoa(JSON.stringify({
	id: 1,
	type: AUTH_TYPES.USER,
	exp: Math.floor(Date.now() / 1000) + 3600,
}))}.signature`

function json(status: number, body: object, headers: Record<string, string> = {}) {
	return new Response(JSON.stringify(body), {
		status,
		headers: {
			'Content-Type': 'application/json',
			...headers,
		},
	})
}

function rateLimited(headers: Record<string, string> = {}) {
	return json(429, {message: 'Too Many Requests'}, headers)
}

function problem(status: number, code?: number) {
	return new Response(JSON.stringify({title: 'Error', status, detail: 'Error', code}), {
		status,
		headers: {'Content-Type': 'application/problem+json'},
	})
}

function htmlPage(status: number) {
	return new Response(`<html>${status}</html>`, {status, headers: {'Content-Type': 'text/html'}})
}

function stubFetch(respond: () => Response) {
	vi.stubGlobal('fetch', vi.fn(async () => respond()))
}

function refreshCalls() {
	return postMock.mock.calls.length
}

let useAuthStore: typeof import('./auth').useAuthStore
let getToken: typeof import('@/helpers/auth').getToken

// The refresh rate-limit gate deliberately survives logout.
async function loadApp() {
	vi.resetModules()
	const {createPinia, setActivePinia} = await import('pinia')
	setActivePinia(createPinia())
	;({useAuthStore} = await import('./auth'))
	;({getToken} = await import('@/helpers/auth'))
}

describe('auth store checkAuth refresh (issue #4023)', () => {
	beforeEach(async () => {
		await loadApp()
		const {authRefreshToken} = await vi.importActual<typeof import('@/client/generated')>('@/client/generated')
		postMock.mockReset().mockImplementation(authRefreshToken)
		userShowMock.mockReset()
		routerMock.push.mockReset()
		routerMock.currentRoute.value = {name: 'home'}
		localStorage.clear()
	})

	afterEach(() => {
		vi.unstubAllGlobals()
		vi.useRealTimers()
	})

	it('does not refresh without a stored token', async () => {
		const store = useAuthStore()

		await store.checkAuth()

		expect(refreshCalls()).toBe(0)
		expect(store.authenticated).toBe(false)
	})

	it.each([
		['a rotated-away refresh token', () => problem(401, 16002)],
		['a proxy 401 page', () => htmlPage(401)],
	])('retries once, then drops the stale token and stops refreshing after %s', async (_, respond) => {
		localStorage.setItem('token', STALE_JWT)
		stubFetch(respond)
		const store = useAuthStore()

		await store.checkAuth()

		expect(refreshCalls()).toBe(2)
		expect(localStorage.getItem('token')).toBeNull()
		expect(store.authenticated).toBe(false)

		// The router guard runs checkAuth again for the /login redirect.
		await store.checkAuth()

		expect(refreshCalls()).toBe(2)
	})

	it('refreshes once and not again this boot when no refresh cookie was sent', async () => {
		localStorage.setItem('token', STALE_JWT)
		stubFetch(() => problem(401, 16005))
		const store = useAuthStore()

		await store.checkAuth()

		expect(refreshCalls()).toBe(1)
		expect(localStorage.getItem('token')).toBeNull()

		await store.checkAuth()

		expect(refreshCalls()).toBe(1)
	})

	it('keeps the token on a rate limit without Retry-After but does not refresh it again this boot', async () => {
		vi.useFakeTimers({toFake: ['setTimeout', 'clearTimeout', 'Date']})
		localStorage.setItem('token', STALE_JWT)
		stubFetch(() => rateLimited())
		const store = useAuthStore()

		await store.checkAuth()

		expect(refreshCalls()).toBe(1)
		expect(localStorage.getItem('token')).toBe(STALE_JWT)

		await vi.advanceTimersByTimeAsync(60 * 60 * 1000)
		await store.checkAuth()

		expect(refreshCalls()).toBe(1)
	})

	describe('rate limit with Retry-After', () => {
		beforeEach(() => {
			vi.useFakeTimers({toFake: ['setTimeout', 'clearTimeout', 'Date']})
			userShowMock.mockResolvedValue({data: {id: 1, username: 'user1'}})
		})

		it('skips refreshes until Retry-After, then retries on its own and restores the session', async () => {
			localStorage.setItem('token', STALE_JWT)
			stubFetch(() => rateLimited({'Retry-After': '30'}))
			const store = useAuthStore()

			await store.checkAuth()

			expect(refreshCalls()).toBe(1)
			expect(localStorage.getItem('token')).toBe(STALE_JWT)
			expect(store.authenticated).toBe(false)

			await vi.advanceTimersByTimeAsync(29_000)
			await store.checkAuth()

			expect(refreshCalls()).toBe(1)

			stubFetch(() => json(200, {token: FRESH_JWT}))
			await vi.advanceTimersByTimeAsync(1_000)
			await flushPromises()

			expect(refreshCalls()).toBe(2)
			expect(getToken()).toBe(FRESH_JWT)
			expect(store.authenticated).toBe(true)
			expect(routerMock.push).not.toHaveBeenCalled()
		})

		it('leaves the login page for the saved route when the scheduled retry restores the session', async () => {
			localStorage.setItem('token', STALE_JWT)
			stubFetch(() => rateLimited({'Retry-After': '30'}))
			const store = useAuthStore()

			await store.checkAuth()
			routerMock.currentRoute.value = {name: 'user.login'}
			localStorage.setItem('lastVisited', JSON.stringify({
				name: 'project.index',
				params: {projectId: '1'},
				query: {},
			}))

			stubFetch(() => json(200, {token: FRESH_JWT}))
			await vi.advanceTimersByTimeAsync(30_000)
			await flushPromises()

			expect(store.authenticated).toBe(true)
			expect(routerMock.push).toHaveBeenCalledExactlyOnceWith({
				name: 'project.index',
				params: {projectId: '1'},
				query: {},
			})
		})

		it('retries after ten minutes when Retry-After is further out', async () => {
			localStorage.setItem('token', STALE_JWT)
			stubFetch(() => rateLimited({'Retry-After': '900'}))
			const store = useAuthStore()

			await store.checkAuth()
			await vi.advanceTimersByTimeAsync(599_000)
			await store.checkAuth()

			expect(refreshCalls()).toBe(1)

			stubFetch(() => json(200, {token: FRESH_JWT}))
			await vi.advanceTimersByTimeAsync(1_000)
			await flushPromises()

			expect(refreshCalls()).toBe(2)
			expect(store.authenticated).toBe(true)
		})

		it('waits at least a second before retrying on Retry-After: 0', async () => {
			localStorage.setItem('token', STALE_JWT)
			stubFetch(() => rateLimited({'Retry-After': '0'}))
			const store = useAuthStore()

			await store.checkAuth()
			await vi.advanceTimersByTimeAsync(999)

			expect(refreshCalls()).toBe(1)

			stubFetch(() => json(200, {token: FRESH_JWT}))
			await vi.advanceTimersByTimeAsync(1)
			await flushPromises()

			expect(refreshCalls()).toBe(2)
			expect(store.authenticated).toBe(true)
		})

		it('still retries when Retry-After runs out while a check is held', async () => {
			localStorage.setItem('token', STALE_JWT)
			stubFetch(() => rateLimited({'Retry-After': '30'}))
			const store = useAuthStore()
			const retryAt = Date.now() + 30_000

			await store.checkAuth()
			await vi.advanceTimersByTimeAsync(29_999)
			// The hold is checked synchronously; the retry is rescheduled after this.
			const checking = store.checkAuth()
			vi.setSystemTime(retryAt)
			await checking

			stubFetch(() => json(200, {token: FRESH_JWT}))
			await vi.advanceTimersByTimeAsync(1_000)
			await flushPromises()

			expect(refreshCalls()).toBe(2)
			expect(store.authenticated).toBe(true)
		})

		it('leaves the redirect to a login that lands while the scheduled retry refreshes', async () => {
			localStorage.setItem('token', STALE_JWT)
			stubFetch(() => rateLimited({'Retry-After': '30'}))
			loginMock.mockResolvedValue({data: {token: FRESH_JWT}})
			const store = useAuthStore()

			await store.checkAuth()
			routerMock.currentRoute.value = {name: 'user.login'}
			localStorage.setItem('lastVisited', JSON.stringify({
				name: 'project.index',
				params: {projectId: '1'},
				query: {},
			}))
			let answerRefresh: (response: Response) => void = () => {}
			vi.stubGlobal('fetch', vi.fn(() => new Promise<Response>(resolve => {
				answerRefresh = resolve
			})))
			await vi.advanceTimersByTimeAsync(30_000)

			expect(refreshCalls()).toBe(2)

			await store.login({username: 'user1', password: 'password'})
			answerRefresh(json(200, {token: FRESH_JWT}))
			await flushPromises()

			expect(store.authenticated).toBe(true)
			expect(routerMock.push).not.toHaveBeenCalled()
		})

		it('does not retry after logout when the rate limit lands while logging out', async () => {
			localStorage.setItem('token', STALE_JWT)
			const answers: Record<string, (response: Response) => void> = {}
			vi.stubGlobal('fetch', vi.fn((request: Request) => new Promise<Response>(resolve => {
				answers[request.url.endsWith('/logout') ? 'logout' : 'refresh'] = resolve
			})))
			const store = useAuthStore()

			const checking = store.checkAuth()
			await flushPromises()
			const loggingOut = store.logout()
			answers.refresh(rateLimited({'Retry-After': '30'}))
			await checking
			await flushPromises()
			answers.logout(json(200, {}))
			await loggingOut
			localStorage.setItem('token', STALE_JWT)
			stubFetch(() => json(200, {token: FRESH_JWT}))
			await vi.advanceTimersByTimeAsync(30_000)
			await flushPromises()

			expect(refreshCalls()).toBe(1)
		})

		it('retries again when the scheduled retry is rate limited too', async () => {
			localStorage.setItem('token', STALE_JWT)
			const responses = [
				rateLimited({'Retry-After': '30'}),
				rateLimited({'Retry-After': '30'}),
				json(200, {token: FRESH_JWT}),
			]
			stubFetch(() => responses.shift() as Response)
			const store = useAuthStore()

			await store.checkAuth()
			await vi.advanceTimersByTimeAsync(30_000)
			await vi.advanceTimersByTimeAsync(30_000)
			await flushPromises()

			expect(refreshCalls()).toBe(3)
			expect(store.authenticated).toBe(true)
		})

	})

	it.each([
		['a network error', 1, () => {
			throw new TypeError('Failed to fetch')
		}],
		['a server error', 2, () => problem(503)],
		['an empty-body proxy 502', 2, () => new Response(null, {status: 502})],
	])('refreshes again on the next check after %s', async (_, calls, respond) => {
		localStorage.setItem('token', STALE_JWT)
		stubFetch(respond)
		const store = useAuthStore()

		await store.checkAuth()

		expect(refreshCalls()).toBe(calls)
		expect(localStorage.getItem('token')).toBe(STALE_JWT)

		await store.checkAuth()

		expect(refreshCalls()).toBe(calls * 2)
	})

	it('recovers within one check when a server error clears on the retry', async () => {
		localStorage.setItem('token', STALE_JWT)
		const responses = [problem(503), json(200, {token: FRESH_JWT})]
		stubFetch(() => responses.shift() as Response)
		userShowMock.mockResolvedValue({data: {id: 1, username: 'user1'}})
		const store = useAuthStore()

		await store.checkAuth()

		expect(refreshCalls()).toBe(2)
		expect(getToken()).toBe(FRESH_JWT)
		expect(store.authenticated).toBe(true)
	})

	it('does not refresh again this boot after a response without a token', async () => {
		localStorage.setItem('token', STALE_JWT)
		stubFetch(() => json(200, {}))
		const store = useAuthStore()

		await store.checkAuth()

		expect(refreshCalls()).toBe(2)
		expect(localStorage.getItem('token')).toBe(STALE_JWT)

		await store.checkAuth()

		expect(refreshCalls()).toBe(2)
	})

	it('keeps a token another tab stored while the refresh was failing', async () => {
		const otherTabJwt = `header.${btoa(JSON.stringify({
			id: 1,
			type: AUTH_TYPES.USER,
			exp: Math.floor(Date.now() / 1000) + 3600,
		}))}.signature`
		localStorage.setItem('token', STALE_JWT)
		stubFetch(() => {
			localStorage.setItem('token', otherTabJwt)
			return problem(401, 16004)
		})
		userShowMock.mockResolvedValue({data: {id: 1, username: 'user1'}})
		const store = useAuthStore()

		await store.checkAuth()

		expect(refreshCalls()).toBe(1)
		expect(localStorage.getItem('token')).toBe(otherTabJwt)
		expect(getToken()).toBe(otherTabJwt)
		expect(store.authenticated).toBe(true)
	})
})

describe('OpenID provider lookup', () => {
	it('rejects an unknown provider and clears loading', async () => {
		await loadApp()
		const store = useAuthStore()
		openidCallbackMock.mockReset()
		await expect(store.openIdAuth({provider: 'missing', code: 'code'}))
			.rejects.toThrow('Unknown OpenID provider: missing')
		expect(openidCallbackMock).not.toHaveBeenCalled()
		expect(store.isLoading).toBe(false)
	})
})
