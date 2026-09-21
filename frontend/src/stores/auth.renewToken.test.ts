vi.mock('@/client/generated', () => ({
	authLogin: sdk.authLogin,
	authLogout: sdk.authLogout,
	tokenRenew: sdk.tokenRenew,
	userShow: sdk.userShow,
}))
import {describe, it, expect, beforeEach, vi} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'
import {nextTick} from 'vue'

import {useAuthStore} from './auth'
import {AUTH_TYPES} from '@/constants/auth'

const sdk = vi.hoisted(() => ({
	authLogin: vi.fn(),
	authLogout: vi.fn(),
	tokenRenew: vi.fn(),
	userShow: vi.fn(),
}))

const {queryClientClearMock, refreshTokenMock, routerPushMock, getTokenMock} = vi.hoisted(() => ({
	queryClientClearMock: vi.fn(),
	refreshTokenMock: vi.fn(),
	routerPushMock: vi.fn(),
	getTokenMock: vi.fn(() => null as string | null),
}))

vi.mock('@/helpers/auth', () => ({
	refreshToken: refreshTokenMock,
	getToken: getTokenMock,
	saveToken: vi.fn(),
	removeToken: vi.fn(),
}))

vi.mock('@/router', () => ({
	default: {push: routerPushMock},
}))

vi.mock('@/client/queryClient', async () => {
	const {QueryClient} = await import('@tanstack/vue-query')
	const queryClient = new QueryClient({defaultOptions: {queries: {retry: false}}})
	const clear = queryClient.clear.bind(queryClient)
	queryClient.clear = () => {clear(); queryClientClearMock()}
	return {queryClient}
})

vi.mock('@/composables/useWebSocket', () => ({
	useWebSocket: () => ({
		disconnect: vi.fn(),
		connect: vi.fn(),
		closeStaleConnection: vi.fn(),
	}),
}))

vi.mock('@/helpers/fetcher', async importOriginal => ({
	...await importOriginal<typeof import('@/helpers/fetcher')>(),
	getApiBaseUrl: () => 'http://localhost/api/v1/',
}))

vi.mock('@/helpers/redirectToProvider', () => ({
	getRedirectUrlFromCurrentFrontendPath: vi.fn(),
	redirectToProvider: vi.fn(),
	redirectToProviderOnLogout: vi.fn(),
}))

function resetSdkMocks() {
	for (const operation of Object.values(sdk)) {
		operation.mockReset().mockResolvedValue({data: {}})
	}
}

// A refresh failure that looks like a real network/HTTP error so renewToken's
// "is this a genuine logout?" check (it inspects the error cause's status) fires.
function refreshError() {
	return new Error('Error renewing token: ', {
		cause: {status: 401},
	})
}

// A JWT carrying a not-yet-expired user session, so the checkAuth() call that
// renewToken() runs after a successful refresh treats the session as live.
function freshUserJwt() {
	const payload = {
		id: 1,
		type: AUTH_TYPES.USER,
		exp: Math.floor(Date.now() / 1000) + 3600,
	}
	const encoded = btoa(JSON.stringify(payload))
	return `header.${encoded}.signature`
}

describe('auth store renewToken retry (issue #2863)', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		resetSdkMocks()
		refreshTokenMock.mockReset()
		queryClientClearMock.mockReset()
		routerPushMock.mockReset()
		getTokenMock.mockReset().mockReturnValue(null)
	})

	function setupExpiredUserSession(store: ReturnType<typeof useAuthStore>) {
		store.setAuthenticated(true)
		// Expired exp so renewToken treats a refresh failure as a real logout.
		store.setSession({
			id: 1,
			type: AUTH_TYPES.USER,
			exp: Math.floor(Date.now() / 1000) - 60,
		})
	}

	it('does NOT log out when the first refresh fails but the retry succeeds', async () => {
		const store = useAuthStore()
		setupExpiredUserSession(store)

		// The retry "succeeds" only if it actually leaves a usable token behind:
		// renewToken() runs checkAuth() afterwards, which reads getToken(). Start
		// with no token, then hand back a fresh JWT once the refresh resolves.
		getTokenMock.mockReturnValue(null)
		refreshTokenMock
			.mockRejectedValueOnce(refreshError())
			.mockImplementationOnce(async () => {
				getTokenMock.mockReturnValue(freshUserJwt())
			})

		await store.renewToken()

		// Two refresh attempts: the initial one and the single retry.
		expect(refreshTokenMock).toHaveBeenCalledTimes(2)
		// The retry recovered the session: the user is still authenticated...
		expect(store.authenticated).toBe(true)
		// ...and was not bounced to login.
		expect(routerPushMock).not.toHaveBeenCalledWith({name: 'user.login'})
	})

	it('logs out when BOTH the refresh and its retry fail', async () => {
		const store = useAuthStore()
		setupExpiredUserSession(store)

		refreshTokenMock
			.mockRejectedValueOnce(refreshError())
			.mockRejectedValueOnce(refreshError())

		await store.renewToken()

		expect(refreshTokenMock).toHaveBeenCalledTimes(2)
		expect(routerPushMock).toHaveBeenCalledWith({name: 'user.login'})
	})

	it('does not log out when the refresh of an expired session is rate limited', async () => {
		const store = useAuthStore()
		setupExpiredUserSession(store)

		refreshTokenMock.mockRejectedValue(new Error('Error renewing token: ', {
			cause: {status: 429, detail: 'rate limit exceeded'},
		}))

		await store.renewToken()

		expect(refreshTokenMock).toHaveBeenCalledTimes(2)
		expect(store.authenticated).toBe(true)
		expect(routerPushMock).not.toHaveBeenCalled()
	})

	it('retries exactly once (no infinite loop) when the session is genuinely dead', async () => {
		const store = useAuthStore()
		setupExpiredUserSession(store)

		refreshTokenMock.mockRejectedValue(refreshError())

		await store.renewToken()

		// Initial attempt + exactly one retry — never more.
		expect(refreshTokenMock).toHaveBeenCalledTimes(2)
	})
})

describe('auth store logout query lifecycle', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		localStorage.clear()
		resetSdkMocks()
		queryClientClearMock.mockReset()
		routerPushMock.mockReset().mockResolvedValue(undefined)
	})

	it('clears server data before navigating away', async () => {
		const store = useAuthStore()
		store.setSession({
			id: 1,
			type: AUTH_TYPES.USER,
			exp: 0,
		})
		queryClientClearMock.mockReset()

		await store.logout()

		expect(queryClientClearMock).toHaveBeenCalledOnce()
		expect(queryClientClearMock.mock.invocationCallOrder[0]).toBeLessThan(routerPushMock.mock.invocationCallOrder[0])
	})

	it('clears browser data after reactive logout cleanup', async () => {
		const store = useAuthStore()
		store.setAuthenticated(true)
		store.setSession({
			id: 1,
			type: AUTH_TYPES.USER,
			exp: 0,
		})
		queryClientClearMock.mockReset()
		localStorage.setItem('projectHistory', '[{"id":1}]')

		const events: string[] = []
		const clear = localStorage.clear.bind(localStorage)
		const clearSpy = vi.spyOn(localStorage, 'clear').mockImplementation(() => {
			events.push('storage')
			clear()
		})
		queryClientClearMock.mockImplementation(() => {
			events.push(`query:${store.authenticated}`)
			localStorage.setItem('projectHistory', '[{"id":1}]')
		})

		await store.logout()
		clearSpy.mockRestore()

		expect(events).toEqual(['query:false', 'storage'])
		expect(localStorage.getItem('projectHistory')).toBeNull()
	})
})

describe('auth store query identity lifecycle', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		resetSdkMocks()
		queryClientClearMock.mockReset()
	})

	async function seedIdentity(id: number, type: AUTH_TYPES) {
		useAuthStore().setSession({
			id,
			type,
			exp: 0,
		})
		await nextTick()
		queryClientClearMock.mockReset()
	}

	it('clears server data when changing users', async () => {
		await seedIdentity(1, AUTH_TYPES.USER)

		useAuthStore().setSession({
			id: 2,
			type: AUTH_TYPES.USER,
			exp: 0,
		})
		await nextTick()

		expect(queryClientClearMock).toHaveBeenCalledOnce()
	})

	it('clears server data when changing from user to link share', async () => {
		await seedIdentity(1, AUTH_TYPES.USER)

		useAuthStore().setSession({
			id: 1,
			type: AUTH_TYPES.LINK_SHARE,
			exp: 0,
		})
		await nextTick()

		expect(queryClientClearMock).toHaveBeenCalledOnce()
	})

	it('preserves server data when authentication fails without an identity transition', async () => {
		await seedIdentity(1, AUTH_TYPES.USER)
		sdk.authLogin.mockRejectedValueOnce(new Error('invalid credentials'))

		await expect(useAuthStore().login({username: 'user', password: 'wrong'})).rejects.toThrow('invalid credentials')
		await nextTick()

		expect(queryClientClearMock).not.toHaveBeenCalled()
	})

	it('preserves server data when renewing the same identity', async () => {
		await seedIdentity(1, AUTH_TYPES.USER)

		useAuthStore().setSession({
			id: 1,
			type: AUTH_TYPES.USER,
			exp: 42,
		})
		await nextTick()

		expect(queryClientClearMock).not.toHaveBeenCalled()
	})
})
