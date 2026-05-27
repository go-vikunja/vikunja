import {beforeEach, describe, expect, it, vi} from 'vitest'
import {mount} from '@vue/test-utils'

import Login from './Login.vue'

const login = vi.fn()

vi.mock('vue-i18n', () => ({
	useI18n: () => ({
		t: (key: string) => key,
	}),
}))

vi.mock('vue-router', () => ({
	useRouter: () => ({
		push: vi.fn(),
	}),
}))

vi.mock('@/composables/useTitle', () => ({
	useTitle: vi.fn(),
}))

vi.mock('@/composables/useRedirectToLastVisited', () => ({
	useRedirectToLastVisited: () => ({
		redirectIfSaved: vi.fn(),
	}),
}))

vi.mock('@/helpers/desktopAuth', () => ({
	isDesktopApp: () => false,
}))

vi.mock('@/helpers/redirectToProvider', () => ({
	redirectToProvider: vi.fn(),
}))

vi.mock('@/message', () => ({
	getErrorText: vi.fn(),
}))

vi.mock('@/stores/auth', () => ({
	useAuthStore: () => ({
		authenticated: false,
		isLoading: false,
		needsTotpPasscode: false,
		verifyEmail: () => Promise.resolve(false),
		login,
		setNeedsTotpPasscode: vi.fn(),
	}),
}))

vi.mock('@/stores/config', () => ({
	useConfigStore: () => ({
		auth: {
			local: {
				enabled: true,
				registrationEnabled: false,
			},
			ldap: {
				enabled: false,
			},
			openidConnect: {
				enabled: false,
				providers: [],
			},
		},
	}),
}))

describe('Login', () => {
	beforeEach(() => {
		login.mockReset()
		sessionStorage.clear()
		document.body.innerHTML = ''
	})

	function mountLogin() {
		return mount(Login, {
			attachTo: document.body,
			global: {
				directives: {
					focus: {mounted: vi.fn()},
					tooltip: {mounted: vi.fn()},
				},
				mocks: {
					$t: (key: string) => key,
				},
				stubs: {
					DesktopLogin: true,
					Icon: true,
					RouterLink: true,
					XButton: {
						template: '<button type="button" @click="$emit(\'click\', $event)"><slot /></button>',
					},
				},
			},
		})
	}

	it('mirrors the username while still submitting the DOM value', async () => {
		login.mockResolvedValue(undefined)

		const wrapper = mountLogin()

		const username = wrapper.get<HTMLInputElement>('input#username')
		await username.setValue('alice')
		expect(username.element.value).toBe('alice')

		// Simulate browser autofill after Vue's input event: the submit path must
		// keep using the DOM value for browsers which do not update bindings.
		username.element.value = 'alice@example.com'
		await wrapper.get<HTMLInputElement>('input#password').setValue('secret')
		await wrapper.get('form').trigger('submit')

		expect(login).toHaveBeenCalledWith({
			username: 'alice@example.com',
			password: 'secret',
			longToken: false,
		})
	})

	it('restores the username if the login form remounts', async () => {
		sessionStorage.setItem('loginUsername', 'alice')

		const wrapper = mountLogin()

		expect(wrapper.get<HTMLInputElement>('input#username').element.value).toBe('alice')
	})
})
