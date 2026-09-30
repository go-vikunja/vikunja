import {describe, it, expect, beforeEach, vi} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'

import {useAuthStore} from './auth'
import {getToken, removeToken, saveToken} from '@/helpers/auth'
import {queryClient} from '@/client/queryClient'
import {accountKeys} from '@/client/queries/account'
import {recordServerClock, serverNowSeconds} from '@/helpers/serverClock'
import {AUTH_TYPES} from '@/constants/auth'

// Real @/helpers/auth so the counts cover its dedupe; only the transport is mocked.
const {postMock, openidCallbackMock, userShowMock} = vi.hoisted(() => ({
	postMock: vi.fn(),
	openidCallbackMock: vi.fn(),
	userShowMock: vi.fn(),
}))

vi.mock('@/client/generated', async (importOriginal) => ({
	...await importOriginal<typeof import('@/client/generated')>(),
	authRefreshToken: postMock,
	authOpenidCallback: openidCallbackMock,
	userShow: userShowMock,
}))

vi.mock('@/router', () => ({
	default: {push: vi.fn()},
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
	sid: 'fresh-session',
}))}.signature`

function refreshRejection(status: number, code?: number) {
	return {status, code}
}

function refreshCalls() {
	return postMock.mock.calls.length
}

describe('auth store checkAuth refresh (issue #4023)', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		queryClient.clear()
		postMock.mockReset()
		userShowMock.mockReset().mockResolvedValue({data: {id: 1, username: 'user1'}})
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

	it('refreshes once and not again this boot when no refresh cookie was sent', async () => {
		localStorage.setItem('token', STALE_JWT)
		postMock.mockRejectedValue(refreshRejection(401, 16005))
		const store = useAuthStore()

		await store.checkAuth()

		expect(refreshCalls()).toBe(1)
		expect(localStorage.getItem('token')).toBeNull()

		await store.checkAuth()

		expect(refreshCalls()).toBe(1)
	})

	it('retries once when the refresh token was rotated away', async () => {
		localStorage.setItem('token', STALE_JWT)
		postMock.mockRejectedValue(refreshRejection(401, 16002))
		const store = useAuthStore()

		await store.checkAuth()

		expect(refreshCalls()).toBe(2)
	})

	it.each([
		['network error', new TypeError('Failed to fetch')],
		['server error', refreshRejection(503)],
		['rate limit', refreshRejection(429)],
	])('recovers on a later auth check after a %s', async (_label, failure) => {
		localStorage.setItem('token', STALE_JWT)
		postMock.mockRejectedValue(failure)
		const store = useAuthStore()

		await store.checkAuth()

		expect(store.authenticated).toBe(false)
		expect(localStorage.getItem('token')).toBe(STALE_JWT)
		postMock.mockResolvedValue({data: {token: FRESH_JWT}})

		await store.checkAuth()

		expect(store.authenticated).toBe(true)
		expect(getToken()).toBe(FRESH_JWT)
		expect(store.currentSessionId).toBe('fresh-session')
	})

	it.each([401, 503])('adopts another tab token after a %s refresh failure', async status => {
		localStorage.setItem('token', STALE_JWT)
		postMock.mockImplementation(async () => {
			localStorage.setItem('token', FRESH_JWT)
			throw refreshRejection(status, 16004)
		})
		const store = useAuthStore()

		await store.checkAuth()

		expect(localStorage.getItem('token')).toBe(FRESH_JWT)
		expect(getToken()).toBe(FRESH_JWT)
		expect(store.authenticated).toBe(true)
		expect(store.currentSessionId).toBe('fresh-session')
		const callsAfterRecovery = refreshCalls()

		await store.checkAuth()

		expect(store.authenticated).toBe(true)
		expect(refreshCalls()).toBe(callsAfterRecovery)
	})

	it('reconciles the account and clears old data when refresh returns another user', async () => {
		localStorage.setItem('token', STALE_JWT)
		const otherUserJwt = `header.${btoa(JSON.stringify({
			id: 2, type: AUTH_TYPES.USER, exp: Math.floor(Date.now() / 1000) + 3600,
			sid: 'other-session',
		}))}.signature`
		postMock.mockResolvedValue({data: {token: otherUserJwt}})
		userShowMock.mockResolvedValue({data: {id: 2, username: 'user2'}})
		const store = useAuthStore()
		store.setSession({id: 1, type: AUTH_TYPES.USER, exp: 1})
		store.setAuthenticated(true)
		queryClient.setQueryData(accountKeys.user(1), {id: 1, username: 'user1'})
		queryClient.setQueryData(['private-projects'], [{id: 123, title: 'Private to user1'}])

		await store.checkAuth()

		expect(getToken()).toBe(otherUserJwt)
		expect(store.authenticated).toBe(true)
		expect(store.identityKey).toBe(`2:${AUTH_TYPES.USER}`)
		expect(store.currentSessionId).toBe('other-session')
		expect(store.info).toMatchObject({id: 2, username: 'user2'})
		expect(queryClient.getQueryData(accountKeys.user(1))).toBeUndefined()
		expect(queryClient.getQueryData(['private-projects'])).toBeUndefined()
	})

	it('preserves a link share opened before a successful refresh resolves', async () => {
		localStorage.setItem('token', STALE_JWT)
		const linkToken = `header.${btoa(JSON.stringify({
			id: 2, type: AUTH_TYPES.LINK_SHARE, exp: 9999999999,
		}))}.signature`
		const store = useAuthStore()
		postMock.mockImplementation(async () => {
			saveToken(linkToken, false)
			store.setSession({id: 2, type: AUTH_TYPES.LINK_SHARE, exp: 9999999999})
			store.setAuthenticated(true)
			queryClient.setQueryData(['link-projects'], [{id: 456}])
			return {data: {token: FRESH_JWT}}
		})

		await store.checkAuth()

		expect(getToken()).toBe(linkToken)
		expect(store.authLinkShare).toBe(true)
		expect(queryClient.getQueryData(['link-projects'])).toEqual([{id: 456}])
		expect(userShowMock).not.toHaveBeenCalled()
	})

	it.each([
		['malformed', 'not-a-jwt'],
		['expired', `header.${btoa(JSON.stringify({id: 1, type: AUTH_TYPES.USER, exp: 1}))}.signature`],
		['link share', `header.${btoa(JSON.stringify({id: 1, type: AUTH_TYPES.LINK_SHARE, exp: 9999999999}))}.signature`],
		['missing identity', `header.${btoa(JSON.stringify({type: AUTH_TYPES.USER, exp: 9999999999}))}.signature`],
	])('does not authenticate with a %s replacement token', async (_label, replacement) => {
		localStorage.setItem('token', STALE_JWT)
		postMock.mockImplementation(async () => {
			localStorage.setItem('token', replacement)
			throw refreshRejection(401, 16004)
		})
		const store = useAuthStore()

		await store.checkAuth()

		expect(store.authenticated).toBe(false)
		expect(localStorage.getItem('token')).toBe(replacement)
	})

	it('preserves the server clock when adopting an older stored token', async () => {
		const now = Math.floor(Date.now() / 1000)
		const otherTabJwt = `header.${btoa(JSON.stringify({
			id: 1, type: AUTH_TYPES.USER, iat: now - 300, exp: now + 300,
		}))}.signature`
		recordServerClock(now, true)
		localStorage.setItem('token', STALE_JWT)
		postMock.mockImplementation(async () => {
			localStorage.setItem('token', otherTabJwt)
			throw refreshRejection(401, 16004)
		})
		const store = useAuthStore()

		await store.checkAuth()

		expect(store.authenticated).toBe(true)
		expect(Math.abs(serverNowSeconds() - now)).toBeLessThan(5)
	})

	it('does not adopt another token after a logout during refresh', async () => {
		localStorage.setItem('token', STALE_JWT)
		postMock.mockImplementation(async () => {
			removeToken()
			localStorage.setItem('token', FRESH_JWT)
			throw refreshRejection(401, 16005)
		})
		const store = useAuthStore()

		await store.checkAuth()

		expect(store.authenticated).toBe(false)
		expect(refreshCalls()).toBe(1)
	})

	it('does not retry or adopt a token after the server changes', async () => {
		localStorage.setItem('token', STALE_JWT)
		const apiUrl = window.API_URL
		postMock.mockImplementation(async () => {
			window.API_URL = 'https://other.example/api/v2'
			localStorage.setItem('token', FRESH_JWT)
			throw refreshRejection(401, 16004)
		})
		const store = useAuthStore()

		try {
			await store.checkAuth()

			expect(store.authenticated).toBe(false)
			expect(getToken()).toBe(STALE_JWT)
			expect(refreshCalls()).toBe(1)
		} finally {
			window.API_URL = apiUrl
		}
	})

	it('does not overwrite a link share opened during refresh', async () => {
		localStorage.setItem('token', STALE_JWT)
		const linkToken = `header.${btoa(JSON.stringify({
			id: 2, type: AUTH_TYPES.LINK_SHARE, exp: 9999999999,
		}))}.signature`
		const store = useAuthStore()
		postMock.mockImplementation(async () => {
			saveToken(linkToken, false)
			store.setSession({id: 2, type: AUTH_TYPES.LINK_SHARE, exp: 9999999999})
			store.setAuthenticated(true)
			localStorage.setItem('token', FRESH_JWT)
			throw refreshRejection(401, 16004)
		})

		await store.checkAuth()

		expect(getToken()).toBe(linkToken)
		expect(store.authLinkShare).toBe(true)
		expect(refreshCalls()).toBe(1)
	})

})


describe('OpenID provider lookup', () => {
	it('rejects an unknown provider and clears loading', async () => {
		setActivePinia(createPinia())
		const store = useAuthStore()
		openidCallbackMock.mockReset()
		await expect(store.openIdAuth({provider: 'missing', code: 'code'}))
			.rejects.toThrow('Unknown OpenID provider: missing')
		expect(openidCallbackMock).not.toHaveBeenCalled()
		expect(store.isLoading).toBe(false)
	})
})
