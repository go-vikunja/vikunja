import {describe, it, expect, beforeEach, vi} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'

import {useAuthStore} from './auth'
import {removeToken} from '@/helpers/auth'
import {AUTH_TYPES} from '@/modelTypes/IUser'

// Real @/helpers/auth so the counts cover its dedupe; only the transport is mocked.
const {postMock} = vi.hoisted(() => ({
	postMock: vi.fn(),
}))

function fakeHttp() {
	return {
		post: postMock,
		get: vi.fn().mockResolvedValue({data: {}}),
		interceptors: {
			request: {use: vi.fn()},
			response: {use: vi.fn()},
		},
	}
}

vi.mock('@/helpers/fetcher', () => ({
	apiV2Url: (path: string) => `/api/v2/${path}`,
	HTTPFactory: () => fakeHttp(),
	AuthenticatedHTTPFactory: () => fakeHttp(),
	getApiBaseUrl: () => 'http://localhost/api/v1/',
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

const STALE_JWT = `header.${btoa(JSON.stringify({
	id: 1,
	type: AUTH_TYPES.USER,
	exp: Math.floor(Date.now() / 1000) - 3600,
}))}.signature`

function refreshRejection(status: number, code?: number) {
	return {
		response: {
			status,
			data: {code},
		},
	}
}

function refreshCalls() {
	return postMock.mock.calls.filter(([url]) => url === '/api/v2/user/token/refresh').length
}

describe('auth store checkAuth refresh (issue #4023)', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		postMock.mockReset()
		removeToken()
		localStorage.clear()
	})

	it('does not refresh without a stored token', async () => {
		const store = useAuthStore()

		await store.checkAuth()

		expect(refreshCalls()).toBe(0)
		expect(store.authenticated).toBe(false)
	})

	it('drops a stale token once its refresh is rejected', async () => {
		localStorage.setItem('token', STALE_JWT)
		postMock.mockRejectedValue(refreshRejection(401, 16003))
		const store = useAuthStore()

		await store.checkAuth()
		const callsAfterFirstCheck = refreshCalls()

		expect(callsAfterFirstCheck).toBeGreaterThan(0)
		expect(localStorage.getItem('token')).toBeNull()
		expect(store.authenticated).toBe(false)

		// The router guard runs checkAuth again for the /login redirect.
		await store.checkAuth()

		expect(refreshCalls()).toBe(callsAfterFirstCheck)
	})

	it('keeps the token on a rate limit but does not refresh it again this boot', async () => {
		localStorage.setItem('token', STALE_JWT)
		postMock.mockRejectedValue(refreshRejection(429))
		const store = useAuthStore()

		await store.checkAuth()
		const callsAfterFirstCheck = refreshCalls()

		expect(localStorage.getItem('token')).toBe(STALE_JWT)

		await store.checkAuth()

		expect(refreshCalls()).toBe(callsAfterFirstCheck)
	})

	it('keeps a token another tab stored while the refresh was failing', async () => {
		const otherTabJwt = `header.${btoa(JSON.stringify({
			id: 1,
			type: AUTH_TYPES.USER,
			exp: Math.floor(Date.now() / 1000) + 3600,
		}))}.signature`
		localStorage.setItem('token', STALE_JWT)
		postMock.mockImplementation(async () => {
			localStorage.setItem('token', otherTabJwt)
			throw refreshRejection(401, 16004)
		})
		const store = useAuthStore()

		await store.checkAuth()

		expect(localStorage.getItem('token')).toBe(otherTabJwt)
	})
})
