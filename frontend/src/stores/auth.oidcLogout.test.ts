import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'

import {useAuthStore} from './auth'

const {authLogoutMock, routerPushMock} = vi.hoisted(() => ({
	authLogoutMock: vi.fn(),
	routerPushMock: vi.fn(),
}))

vi.mock('@/client/generated', async (importOriginal) => ({
	...await importOriginal<typeof import('@/client/generated')>(),
	authLogout: authLogoutMock,
}))

vi.mock('@/router', () => ({
	default: {push: routerPushMock},
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
	redirectToProviderOnLogout: vi.fn(() => false),
}))

describe('auth store OIDC logout redirect', () => {
	const originalLocation = window.location

	beforeEach(() => {
		setActivePinia(createPinia())
		localStorage.clear()
		authLogoutMock.mockReset()
		routerPushMock.mockReset().mockResolvedValue(undefined)
		Object.defineProperty(window, 'location', {
			configurable: true,
			value: {...originalLocation, href: 'http://localhost/'},
		})
	})

	afterEach(() => {
		Object.defineProperty(window, 'location', {configurable: true, value: originalLocation})
	})

	it('redirects to an https oidc_logout_url', async () => {
		authLogoutMock.mockResolvedValue({data: {oidc_logout_url: 'https://op.example.com/logout?client_id=x'}})

		await useAuthStore().logout()

		expect(window.location.href).toBe('https://op.example.com/logout?client_id=x')
		expect(routerPushMock).not.toHaveBeenCalled()
	})

	it('does not navigate to a javascript: oidc_logout_url', async () => {
		authLogoutMock.mockResolvedValue({data: {oidc_logout_url: 'javascript:alert(document.domain)//?client_id=x'}})

		await useAuthStore().logout()

		expect(window.location.href).toBe('http://localhost/')
		expect(routerPushMock).toHaveBeenCalledWith({name: 'user.login'})
	})
})
