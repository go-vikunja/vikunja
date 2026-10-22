import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {beforeEach, describe, it, expect, vi} from 'vitest'
import {mount, flushPromises} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import PasswordReset from './PasswordReset.vue'
import Password from '@/components/input/Password.vue'
import en from '@/i18n/lang/en.json'

const {resetPassword, mockRoute} = vi.hoisted(() => ({
	resetPassword: vi.fn(),
	mockRoute: {query: {} as Record<string, string>, hash: ''},
}))

vi.mock('@/client/generated', () => ({authPasswordReset: resetPassword}))

vi.mock('vue-router', () => ({
	useRoute: () => mockRoute,
}))

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function mountPasswordReset(errors: unknown[] = []) {
	return mount(PasswordReset, {
		global: {
			plugins: [i18n, [VueQueryPlugin, {queryClient: new QueryClient()}]],
			stubs: {XButton: true, Password: true},
			config: {
				errorHandler(err) {
					errors.push(err)
				},
			},
		},
	})
}

describe('PasswordReset', () => {
	beforeEach(() => {
		resetPassword.mockReset()
		mockRoute.query = {userPasswordReset: 'token'}
		mockRoute.hash = ''
	})

	it('reads the token from the fragment', async () => {
		resetPassword.mockResolvedValue({data: {message: 'ok'}})
		mockRoute.query = {}
		mockRoute.hash = '#userPasswordReset=fragment-token'

		const wrapper = mountPasswordReset()
		wrapper.findComponent(Password).vm.$emit('update:modelValue', 'new-password')
		await wrapper.find('form').trigger('submit')
		await flushPromises()

		expect(resetPassword).toHaveBeenCalledWith({body: {new_password: 'new-password', token: 'fragment-token'}})
	})

	it('shows the error when the request fails without a response', async () => {
		resetPassword.mockRejectedValue(new TypeError('Failed to fetch'))
		const errors: unknown[] = []

		const wrapper = mountPasswordReset(errors)

		wrapper.findComponent(Password).vm.$emit('update:modelValue', 'new-password')
		await wrapper.find('form').trigger('submit')
		await flushPromises()

		expect(errors).toEqual([])
		expect(resetPassword).toHaveBeenCalledWith({body: {new_password: 'new-password', token: 'token'}})
		expect(wrapper.text()).toContain('Failed to fetch')
	})
})
