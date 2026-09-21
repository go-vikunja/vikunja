import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {describe, it, expect, vi} from 'vitest'
import {mount, flushPromises} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import RequestPasswordReset from './RequestPasswordReset.vue'
import en from '@/i18n/lang/en.json'

const {requestResetPassword} = vi.hoisted(() => ({requestResetPassword: vi.fn()}))

vi.mock('@/client/generated', () => ({authPasswordToken: requestResetPassword}))

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

describe('RequestPasswordReset', () => {
	it('shows the error when the request fails without a response', async () => {
		requestResetPassword.mockRejectedValue(new TypeError('Failed to fetch'))
		const errors: unknown[] = []

		const wrapper = mount(RequestPasswordReset, {
			global: {
				plugins: [i18n, [VueQueryPlugin, {queryClient: new QueryClient()}]],
				directives: {focus: {}},
				stubs: {XButton: true, FormField: true},
				config: {
					errorHandler(err) {
						errors.push(err)
					},
				},
			},
		})

		await wrapper.find('form').trigger('submit')
		await flushPromises()

		expect(errors).toEqual([])
		expect(wrapper.text()).toContain('Failed to fetch')
	})
})
