import {mount} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import {expect, it} from 'vitest'
import Password from './Password.vue'
import en from '@/i18n/lang/en.json'

it('keeps a revealed password in sync when the parent clears the form', async () => {
	const wrapper = mount(Password, {
		props: {modelValue: 'secret', id: 'reauth-password', autofocus: true},
		global: {
			plugins: [createI18n({legacy: false, locale: 'en', messages: {en}})],
			stubs: {Icon: true},
			directives: {tooltip: {}},
		},
	})

	expect(wrapper.find('[role=alert]').exists()).toBe(false)
	const input = wrapper.get('input')
	expect(input.attributes('id')).toBe('reauth-password')
	expect(input.attributes('autofocus')).toBeDefined()
	expect((input.element as HTMLInputElement).value).toBe('secret')
	await wrapper.get('button').trigger('click')
	expect(input.attributes('type')).toBe('text')
	expect((input.element as HTMLInputElement).value).toBe('secret')
	await wrapper.setProps({modelValue: ''})
	expect((input.element as HTMLInputElement).value).toBe('')
})
