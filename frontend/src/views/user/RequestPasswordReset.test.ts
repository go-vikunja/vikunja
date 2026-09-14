import {describe, it, expect, vi} from 'vitest'
import {mount, flushPromises} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import RequestPasswordReset from './RequestPasswordReset.vue'
import en from '@/i18n/lang/en.json'

const requestResetPassword = vi.fn()

vi.mock('@/services/passwordReset', () => ({
	default: class {
		loading = false
		requestResetPassword = requestResetPassword
	},
}))

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

describe('RequestPasswordReset', () => {
	it('shows the error when the request fails without a response', async () => {
		// Axios network errors (unreachable API, CORS) carry no `response`.
		requestResetPassword.mockRejectedValue(Object.assign(new Error('Network Error'), {code: 'ERR_NETWORK'}))
		const errors: unknown[] = []

		const wrapper = mount(RequestPasswordReset, {
			global: {
				plugins: [i18n],
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
		expect(wrapper.text()).toContain('Network Error')
	})
})
