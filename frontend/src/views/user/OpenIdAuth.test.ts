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
		userIdentitiesPreview: vi.fn(),
	},
	mockRoute: {
		params: {provider: 'dex'} as Record<string, string>,
		query: {} as Record<string, string | null>,
	},
	router: {
		push: vi.fn(),
		replace: vi.fn(),
	},
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

const i18n = createI18n({
	legacy: false,
	locale: 'en',
	messages: {en},
})

const REDIRECT_URL = 'http://localhost:3000/auth/openid/dex'

const queryClientConfig = {
	defaultOptions: {
		queries: {retry: false},
	},
}

function mountCallback() {
	return mount(OpenIdAuth, {
		global: {
			plugins: [i18n, [VueQueryPlugin, {queryClient: new QueryClient(queryClientConfig)}]],
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
	mockRoute.query = {
		code: 'auth-code',
		state: 'expected-state',
	}
	authStore.authUser = false
	authStore.openIdAuth.mockResolvedValue(undefined)
	sdk.totpGet.mockResolvedValue({data: {enabled: false}})
	sdk.userIdentitiesPreview.mockResolvedValue({data: {account: 'test@example.com'}})
})

describe('OpenIdAuth login', () => {
	it('refuses a callback whose state does not match', async () => {
		mockRoute.query = {
			code: 'auth-code',
			state: 'forged',
		}
		const wrapper = mountCallback()
		await flushPromises()

		expect(authStore.openIdAuth).not.toHaveBeenCalled()
		expect(wrapper.text()).toContain('State does not match')
	})

	it('refuses a bare state when none is stored', async () => {
		localStorage.clear()
		authStore.authUser = true
		mockRoute.query = {
			code: 'x',
			state: null,
		}
		const wrapper = mountCallback()
		await flushPromises()

		expect(wrapper.text()).toContain('State does not match')
		expect(wrapper.find('#linkIdentityPassword').exists()).toBe(false)
		expect(authStore.openIdAuth).not.toHaveBeenCalled()
		expect(sdk.userIdentitiesLink).not.toHaveBeenCalled()
	})

	it('does not enter link mode on a provider error with a bare state', async () => {
		localStorage.clear()
		authStore.authUser = true
		mockRoute.query = {
			error: 'access_denied',
			state: null,
		}
		const wrapper = mountCallback()
		await flushPromises()

		expect(wrapper.text()).toContain('An error occurred while authenticating')
		expect(wrapper.findComponent({name: 'XButton'}).exists()).toBe(false)
	})

	it.each([
		['missing', undefined],
		['bare', null],
		['empty', ''],
	])('refuses a %s state while a flow is pending', async (_, state) => {
		mockRoute.query = {code: 'auth-code'}
		if (state !== undefined) {
			mockRoute.query.state = state
		}
		const wrapper = mountCallback()
		await flushPromises()

		expect(authStore.openIdAuth).not.toHaveBeenCalled()
		expect(wrapper.text()).toContain('State does not match')
	})

	it('refuses an empty state when an empty one is stored', async () => {
		localStorage.setItem('state', '')
		mockRoute.query = {
			code: 'auth-code',
			state: '',
		}
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

	it.each([
		['succeeds', () => authStore.openIdAuth.mockResolvedValue(undefined)],
		['fails', () => authStore.openIdAuth.mockRejectedValue({code: 1011})],
	])('consumes the stored state when the login %s', async (_, setup) => {
		setup()
		mountCallback()
		await flushPromises()

		expect(localStorage.getItem('state')).toBeNull()

		authStore.openIdAuth.mockClear()
		const revisit = mountCallback()
		await flushPromises()

		expect(authStore.openIdAuth).not.toHaveBeenCalled()
		expect(revisit.text()).toContain('State does not match')
	})

	it('consumes the stored state when the provider reports an error', async () => {
		mockRoute.query = {
			error: 'access_denied',
			state: 'expected-state',
		}
		mountCallback()
		await flushPromises()

		expect(localStorage.getItem('state')).toBeNull()
	})

	it('keeps the pending flow when the state does not match', async () => {
		mockRoute.query = {
			code: 'auth-code',
			state: 'forged',
		}
		mountCallback()
		await flushPromises()

		expect(localStorage.getItem('state')).toBe('expected-state')
	})

	it('asks for a TOTP passcode when the login requires one', async () => {
		authStore.openIdAuth.mockRejectedValue({code: 1017})
		const wrapper = mountCallback()
		await flushPromises()

		expect(wrapper.find('#openIdTotpPasscode').exists()).toBe(true)
		expect(redirectIfSaved).not.toHaveBeenCalled()
	})
})

describe('OpenIdAuth linking', () => {
	beforeEach(() => {
		localStorage.setItem('openidLinking', 'expected-state')
		authStore.authUser = true
		sdk.userIdentitiesLink.mockResolvedValue({
			data: {
				id: 1,
				provider_key: 'dex',
			},
		})
	})

	async function submitLink(wrapper: ReturnType<typeof mountCallback>, password: string) {
		await wrapper.find('#linkIdentityPassword').setValue(password)
		await wrapper.find('form').trigger('submit')
		await flushPromises()
	}

	it('asks for the password instead of logging in', async () => {
		const wrapper = mountCallback()
		await flushPromises()

		expect(authStore.openIdAuth).not.toHaveBeenCalled()
		expect(wrapper.text()).toContain('Link Dex to your account')
		expect(wrapper.text()).toContain('test@example.com')
		expect(wrapper.find('#linkIdentityPasscode').exists()).toBe(false)
	})

	it('waits for the verified account before accepting a password', async () => {
		let resolvePreview!: (value: {data: {account: string}}) => void
		sdk.userIdentitiesPreview.mockReturnValue(new Promise(resolve => { resolvePreview = resolve }))
		const wrapper = mountCallback()
		await flushPromises()

		expect(wrapper.find('#linkIdentityPassword').exists()).toBe(false)
		await wrapper.find('form').trigger('submit')
		expect(sdk.userIdentitiesLink).not.toHaveBeenCalled()

		resolvePreview({data: {account: 'alice@example.com'}})
		await flushPromises()
		expect(wrapper.text()).toContain('alice@example.com')
		expect(wrapper.find('#linkIdentityPassword').exists()).toBe(true)
	})

	it('ends the flow when the account cannot be previewed', async () => {
		sdk.userIdentitiesPreview.mockRejectedValue({code: 15001, message: 'Provider unavailable'})
		const wrapper = mountCallback()
		await flushPromises()

		expect(wrapper.find('#linkIdentityPassword').exists()).toBe(false)
		expect(wrapper.getComponent({name: 'XButton'}).vm.$attrs.to).toEqual({name: 'user.settings.linked-accounts'})
		expect(localStorage.getItem('openidLinking')).toBeNull()
		expect(localStorage.getItem('state')).toBeNull()
		expect(sdk.userIdentitiesLink).not.toHaveBeenCalled()
	})

	it('ignores a preview completed after cancelling', async () => {
		let resolvePreview!: (value: {data: {account: string}}) => void
		sdk.userIdentitiesPreview.mockReturnValue(new Promise(resolve => { resolvePreview = resolve }))
		const wrapper = mountCallback()
		await flushPromises()
		await wrapper.find('x-button-stub[variant="tertiary"]').trigger('click')

		resolvePreview({data: {account: 'alice@example.com'}})
		await flushPromises()
		expect(wrapper.find('#linkIdentityPassword').exists()).toBe(false)
		expect(wrapper.text()).not.toContain('alice@example.com')
		expect(localStorage.getItem('openidLinking')).toBeNull()
		expect(router.replace).toHaveBeenCalledWith({name: 'user.settings.linked-accounts'})
	})

	it('does not log in when the callback is revisited after linking', async () => {
		const wrapper = mountCallback()
		await flushPromises()
		await submitLink(wrapper, 'secret')

		expect(sdk.userIdentitiesLink).toHaveBeenCalledWith({
			body: {
				provider: 'dex',
				code: 'auth-code',
				redirect_url: REDIRECT_URL,
				password: 'secret',
				totp_passcode: undefined,
			},
		})
		expect(router.replace).toHaveBeenCalledWith({name: 'user.settings.linked-accounts'})
		expect(localStorage.getItem('openidLinking')).toBeNull()

		const revisit = mountCallback()
		await flushPromises()

		expect(authStore.openIdAuth).not.toHaveBeenCalled()
		expect(revisit.text()).toContain('State does not match')
	})

	it('does not log in when the callback is revisited after cancelling', async () => {
		const wrapper = mountCallback()
		await flushPromises()
		await wrapper.find('x-button-stub[variant="tertiary"]').trigger('click')

		expect(localStorage.getItem('openidLinking')).toBeNull()
		expect(router.replace).toHaveBeenCalledWith({name: 'user.settings.linked-accounts'})

		const revisit = mountCallback()
		await flushPromises()

		expect(authStore.openIdAuth).not.toHaveBeenCalled()
		expect(revisit.text()).toContain('State does not match')
	})

	it('disables cancel while the link request is pending', async () => {
		sdk.userIdentitiesLink.mockReturnValue(new Promise(() => {}))
		const wrapper = mountCallback()
		await flushPromises()
		await submitLink(wrapper, 'secret')

		expect(wrapper.find('x-button-stub[variant="tertiary"]').attributes('disabled')).toBe('true')
	})

	it('offers the way back when the provider reports an error', async () => {
		mockRoute.query = {
			error: 'access_denied',
			state: 'expected-state',
		}
		const wrapper = mountCallback()
		await flushPromises()

		expect(authStore.openIdAuth).not.toHaveBeenCalled()
		expect(wrapper.text()).toContain('An error occurred while authenticating')
		expect(wrapper.getComponent({name: 'XButton'}).vm.$attrs.to).toEqual({name: 'user.settings.linked-accounts'})
		expect(localStorage.getItem('openidLinking')).toBeNull()
		expect(localStorage.getItem('state')).toBeNull()
	})

	it('offers the login when the provider reports an error without a session', async () => {
		authStore.authUser = false
		mockRoute.query = {
			error: 'access_denied',
			state: 'expected-state',
		}
		const wrapper = mountCallback()
		await flushPromises()

		expect(wrapper.text()).toContain('An error occurred while authenticating')
		expect(wrapper.getComponent({name: 'XButton'}).vm.$attrs.to).toEqual({name: 'user.login'})
		expect(localStorage.getItem('openidLinking')).toBeNull()
		expect(localStorage.getItem('state')).toBeNull()
	})

	it('sends the TOTP passcode when TOTP is enabled', async () => {
		sdk.totpGet.mockResolvedValue({data: {enabled: true}})
		const wrapper = mountCallback()
		await flushPromises()
		await wrapper.find('#linkIdentityPasscode').setValue('123456')
		await submitLink(wrapper, 'secret')

		expect(sdk.userIdentitiesLink).toHaveBeenCalledWith({
			body: expect.objectContaining({
				password: 'secret',
				totp_passcode: '123456',
			}),
		})
	})

	it('marks the submit button loading while the TOTP status is loading', async () => {
		sdk.totpGet.mockReturnValue(new Promise(() => {}))
		const wrapper = mountCallback()
		await flushPromises()

		expect(wrapper.find('x-button-stub[type="submit"]').attributes('loading')).toBe('true')
		await submitLink(wrapper, 'secret')
		expect(sdk.userIdentitiesLink).not.toHaveBeenCalled()
	})

	it('asks for the TOTP passcode when the TOTP status cannot be loaded', async () => {
		sdk.totpGet.mockRejectedValue({status: 500})
		const wrapper = mountCallback()
		await flushPromises()

		expect(wrapper.find('#linkIdentityPasscode').exists()).toBe(true)
		expect(wrapper.find('x-button-stub[type="submit"]').attributes('loading')).toBe('false')
	})

	it('keeps the form for a retry after a retryable error', async () => {
		sdk.userIdentitiesLink.mockRejectedValueOnce({
			status: 403,
			code: 1039,
		})
		const wrapper = mountCallback()
		await flushPromises()
		await submitLink(wrapper, 'secret')

		expect(wrapper.find('form').exists()).toBe(true)
		expect(localStorage.getItem('openidLinking')).toBe('expected-state')
		expect(localStorage.getItem('state')).toBe('expected-state')
		expect(router.replace).not.toHaveBeenCalled()
	})

	it('drops the form once the code was used up', async () => {
		sdk.userIdentitiesLink.mockRejectedValue({
			status: 409,
			code: 15002,
		})
		const wrapper = mountCallback()
		await flushPromises()
		await submitLink(wrapper, 'secret')

		expect(wrapper.text()).toContain('already linked to another Vikunja account')
		expect(wrapper.find('form').exists()).toBe(false)
		expect(localStorage.getItem('openidLinking')).toBeNull()
	})

	it('refuses to link or log in without a session', async () => {
		authStore.authUser = false
		const wrapper = mountCallback()
		await flushPromises()

		expect(authStore.openIdAuth).not.toHaveBeenCalled()
		expect(sdk.userIdentitiesLink).not.toHaveBeenCalled()
		expect(wrapper.find('form').exists()).toBe(false)
		expect(wrapper.text()).toContain('You need to be logged in to link an account.')
		expect(wrapper.getComponent({name: 'XButton'}).vm.$attrs.to).toEqual({name: 'user.login'})
		expect(localStorage.getItem('openidLinking')).toBeNull()
	})

	it('logs in when the marker belongs to another flow', async () => {
		localStorage.setItem('openidLinking', 'other-state')
		const wrapper = mountCallback()
		await flushPromises()

		expect(authStore.openIdAuth).toHaveBeenCalledWith({
			provider: 'dex',
			code: 'auth-code',
			totpPasscode: undefined,
		})
		expect(wrapper.find('#linkIdentityPassword').exists()).toBe(false)
		expect(localStorage.getItem('openidLinking')).toBe('other-state')
	})

	it('checks the state before entering link mode', async () => {
		mockRoute.query = {
			code: 'auth-code',
			state: 'forged',
		}
		const wrapper = mountCallback()
		await flushPromises()

		expect(wrapper.find('form').exists()).toBe(false)
		expect(wrapper.text()).toContain('State does not match')
	})
})
