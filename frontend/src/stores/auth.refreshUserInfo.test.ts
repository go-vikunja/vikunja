vi.mock('@/client/generated', () => ({userShow: httpGetMock}))
import {describe, it, expect, beforeEach, vi} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'

import {useAuthStore} from './auth'
import {shouldDropEvent} from '@/helpers/sentryFilters'
import {getErrorText} from '@/message'

const {httpGetMock, routerPushMock, getTokenMock} = vi.hoisted(() => ({
	httpGetMock: vi.fn(),
	routerPushMock: vi.fn(),
	getTokenMock: vi.fn(),
}))

vi.mock('@/helpers/auth', () => ({
	refreshToken: vi.fn(),
	getToken: getTokenMock,
	saveToken: vi.fn(),
	removeToken: () => getTokenMock.mockReturnValue(null),
}))

vi.mock('@/router', () => ({
	default: {push: routerPushMock},
}))

vi.mock('@/client/queryClient', async () => {
	const {QueryClient} = await import('@tanstack/vue-query')
	const queryClient = new QueryClient({defaultOptions: {queries: {retry: false}}})
	return {queryClient}
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

async function refreshError(): Promise<unknown> {
	try {
		await useAuthStore().refreshUserInfo()
	} catch (e) {
		return e
	}
	throw new Error('refreshUserInfo did not throw')
}

describe('auth store refreshUserInfo failures', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		httpGetMock.mockReset()
		routerPushMock.mockReset()
		getTokenMock.mockReset().mockReturnValue('token')
		vi.spyOn(console, 'error').mockImplementation(() => {})
	})

	it('throws an error sentry drops on a network error', async () => {
		httpGetMock.mockRejectedValue(new TypeError('Failed to fetch'))

		expect(shouldDropEvent(await refreshError())).toBe(true)
	})

	it('throws an error sentry reports and that shows the server message on a 5xx', async () => {
		httpGetMock.mockRejectedValue({status: 500, detail: 'Internal server error'})

		const e = await refreshError()

		expect(shouldDropEvent(e)).toBe(false)
		expect(getErrorText(e)).toBe('Error while refreshing user info: Internal server error')
	})

	it.each([401, 403])('logs out on a %i', async status => {
		const store = useAuthStore()
		store.setAuthenticated(true)
		httpGetMock.mockRejectedValue({status, detail: 'invalid token'})

		await expect(store.refreshUserInfo()).resolves.toBeUndefined()

		expect(store.authenticated).toBe(false)
		expect(routerPushMock).toHaveBeenCalledWith({name: 'user.login'})
	})

	it.each([404, 429])('keeps the session and throws on a %i', async status => {
		const store = useAuthStore()
		store.setAuthenticated(true)
		httpGetMock.mockRejectedValue({status, detail: 'rejected'})

		const e = await refreshError()

		expect(getErrorText(e)).toBe('Error while refreshing user info: rejected')
		expect(store.authenticated).toBe(true)
		expect(routerPushMock).not.toHaveBeenCalled()
	})
})
