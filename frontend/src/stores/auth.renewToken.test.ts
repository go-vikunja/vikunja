import {describe, it, expect, beforeEach, vi} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'

import {useAuthStore} from './auth'
import {AUTH_TYPES} from '@/modelTypes/IUser'

const {refreshTokenMock, routerPushMock, getTokenMock} = vi.hoisted(() => ({
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

vi.mock('@/composables/useWebSocket', () => ({
	useWebSocket: () => ({disconnect: vi.fn(), connect: vi.fn()}),
}))

function fakeHttp() {
	return {
		post: vi.fn().mockResolvedValue({data: {}}),
		get: vi.fn().mockResolvedValue({data: {}}),
		request: vi.fn().mockResolvedValue({data: {}}),
		interceptors: {
			request: {use: vi.fn()},
			response: {use: vi.fn()},
		},
	}
}

vi.mock('@/helpers/fetcher', () => ({
	HTTPFactory: () => fakeHttp(),
	AuthenticatedHTTPFactory: () => fakeHttp(),
	getApiBaseUrl: () => 'http://localhost/api/v1/',
}))

vi.mock('@/helpers/redirectToProvider', () => ({
	getRedirectUrlFromCurrentFrontendPath: vi.fn(),
	redirectToProvider: vi.fn(),
	redirectToProviderOnLogout: vi.fn(),
}))

// A refresh failure that looks like a real network/HTTP error so renewToken's
// "is this a genuine logout?" check (it inspects the error cause's status) fires.
function refreshError() {
	return new Error('Error renewing token: ', {
		cause: {response: {status: 401}},
	})
}

describe('auth store renewToken retry (issue #2863)', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		refreshTokenMock.mockReset()
		routerPushMock.mockReset()
		getTokenMock.mockReset().mockReturnValue(null)
	})

	function setupExpiredUserSession(store: ReturnType<typeof useAuthStore>) {
		store.setAuthenticated(true)
		// Expired exp so renewToken treats a refresh failure as a real logout.
		store.setUser({
			id: 1,
			type: AUTH_TYPES.USER,
			exp: Math.floor(Date.now() / 1000) - 60,
		} as never, false)
	}

	it('does NOT log out when the first refresh fails but the retry succeeds', async () => {
		const store = useAuthStore()
		setupExpiredUserSession(store)

		refreshTokenMock
			.mockRejectedValueOnce(refreshError())
			.mockResolvedValueOnce(undefined)

		await store.renewToken()

		// Two refresh attempts: the initial one and the single retry.
		expect(refreshTokenMock).toHaveBeenCalledTimes(2)
		// The retry succeeded, so the user is not bounced to login.
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

	it('retries exactly once (no infinite loop) when the session is genuinely dead', async () => {
		const store = useAuthStore()
		setupExpiredUserSession(store)

		refreshTokenMock.mockRejectedValue(refreshError())

		await store.renewToken()

		// Initial attempt + exactly one retry — never more.
		expect(refreshTokenMock).toHaveBeenCalledTimes(2)
	})
})
