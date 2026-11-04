import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {beforeEach, describe, it, expect, vi} from 'vitest'
import {mount, flushPromises} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import OpenIdAuth from './OpenIdAuth.vue'
import en from '@/i18n/lang/en.json'

const {sdk, mockRoute, router, authStore, redirectIfSaved} = vi.hoisted(() => ({
	sdk: {
		totpGet: vi.fn(),
		userIdentitiesLink: vi.fn(),
	},
	mockRoute: {
		params: {provider: 'dex'} as Record<string, string>,
		query: {} as Record<string, string>,
	},
	router: {push: vi.fn()},
	authStore: {
		authUser: false,
		isLoading: false,
		openIdAuth: vi.fn(),
	},
	redirectIfSaved: vi.fn(),
}))

vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', async importOriginal => ({
	...await importOriginal<typeof import('@/message')>(),
	success: vi.fn(),
	error: vi.fn(),
}))
vi.mock('vue-router', () => ({
	useRoute: () => mockRoute,
	useRouter: () => router,
}))
vi.mock('@/stores/auth', () => ({useAuthStore: () => authStore}))
vi.mock('@/stores/config', () => ({
	useConfigStore: () => ({
		totp_enabled: true,
		auth: {
			openid_connect: {
				enabled: true,
				providers: [{
					key: 'dex',
					name: 'Dex',
					auth_url: 'http://dex:5556/auth',
					client_id: 'vikunja',
					scope: 'openid',
					logout_url: '',
				}],
			},
		},
	}),
}))
vi.mock('@/composables/useRedirectToLastVisited', () => ({
	useRedirectToLastVisited: () => ({redirectIfSaved}),
}))

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function mountCallback() {
	return mount(OpenIdAuth, {
		global: {
			plugins: [i18n, [VueQueryPlugin, {queryClient: new QueryClient()}]],
			stubs: {XButton: true},
			directives: {focus: {}},
		},
	})
}

beforeEach(() => {
	vi.clearAllMocks()
	localStorage.clear()
	sessionStorage.clear()
	localStorage.setItem('state', 'expected-state')
	mockRoute.query = {code: 'auth-code', state: 'expected-state'}
	authStore.authUser = false
	authStore.openIdAuth.mockResolvedValue(undefined)
	sdk.totpGet.mockResolvedValue({data: {enabled: false}})
})

describe('OpenIdAuth login', () => {
	it('refuses a callback whose state does not match', async () => {
		mockRoute.query = {code: 'auth-code', state: 'forged'}
		const wrapper = mountCallback()
		await flushPromises()

		expect(authStore.openIdAuth).not.toHaveBeenCalled()
		expect(wrapper.text()).toContain('State does not match')
	})

	it('logs in with the code', async () => {
		mountCallback()
		await flushPromises()

		expect(authStore.openIdAuth).toHaveBeenCalledWith({
			provider: 'dex',
			code: 'auth-code',
			totpPasscode: undefined,
		})
		expect(redirectIfSaved).toHaveBeenCalled()
	})

	it('asks for a TOTP passcode when the login requires one', async () => {
		authStore.openIdAuth.mockRejectedValue({code: 1017})
		const wrapper = mountCallback()
		await flushPromises()

		expect(wrapper.find('#openIdTotpPasscode').exists()).toBe(true)
		expect(redirectIfSaved).not.toHaveBeenCalled()
	})
})
