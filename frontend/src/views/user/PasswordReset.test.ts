import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {describe, it, expect, vi} from 'vitest'
import {mount, flushPromises} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import PasswordReset from './PasswordReset.vue'
import Password from '@/components/input/Password.vue'
import en from '@/i18n/lang/en.json'

const {resetPassword} = vi.hoisted(() => ({resetPassword: vi.fn()}))

vi.mock('@/client/generated', () => ({authPasswordReset: resetPassword}))

vi.mock('vue-router', () => ({
	useRoute: () => ({query: {userPasswordReset: 'token'}}),
}))

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

describe('PasswordReset', () => {
	it('shows the error when the request fails without a response', async () => {
		resetPassword.mockRejectedValue(new TypeError('Failed to fetch'))
		const errors: unknown[] = []

		const wrapper = mount(PasswordReset, {
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

		wrapper.findComponent(Password).vm.$emit('update:modelValue', 'new-password')
		await wrapper.find('form').trigger('submit')
		await flushPromises()

		expect(errors).toEqual([])
		expect(wrapper.text()).toContain('Failed to fetch')
	})
})
