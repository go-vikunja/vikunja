vi.mock('@/client/generated', () => ({userShow: httpGetMock}))
import {describe, it, expect, beforeEach, vi} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'

import {useAuthStore} from './auth'
import {shouldDropEvent} from '@/helpers/sentryFilters'
import {getErrorText} from '@/message'

const {httpGetMock} = vi.hoisted(() => ({
	httpGetMock: vi.fn(),
}))

vi.mock('@/helpers/auth', () => ({
	refreshToken: vi.fn(),
	getToken: () => 'token',
	saveToken: vi.fn(),
	removeToken: vi.fn(),
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

function fakeHttp() {
	return {
		get: httpGetMock,
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
		vi.spyOn(console, 'error').mockImplementation(() => {})
	})

	it('throws an error sentry drops on a network error', async () => {
		httpGetMock.mockRejectedValue(new TypeError('Failed to fetch'))

		expect(shouldDropEvent(await refreshError())).toBe(true)
	})

	it('throws an error that shows the server message on a 5xx', async () => {
		httpGetMock.mockRejectedValue({status: 500, detail: 'Internal server error'})

		const e = await refreshError()

		expect(shouldDropEvent(e)).toBe(true)
		expect(getErrorText(e)).toBe('Error while refreshing user info: Internal server error')
	})
})
