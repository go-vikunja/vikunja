import {describe, it, expect} from 'vitest'
import {mount} from '@vue/test-utils'
import FormCheckbox from './FormCheckbox.vue'

describe('FormCheckbox', () => {
	it('renders a Bulma-classed checkbox label', () => {
		const wrapper = mount(FormCheckbox, {props: {label: 'Enable thing'}})
		const label = wrapper.find('label.checkbox')
		expect(label.exists()).toBe(true)
		expect(label.text()).toContain('Enable thing')
		expect(label.find('input[type="checkbox"]').exists()).toBe(true)
	})

	it('supports v-model (boolean)', async () => {
		const wrapper = mount(FormCheckbox, {
			props: {
				label: 'Toggle',
				modelValue: false,
				'onUpdate:modelValue': (val: boolean) => wrapper.setProps({modelValue: val}),
			},
		})
		const input = wrapper.find('input[type="checkbox"]')
		expect((input.element as HTMLInputElement).checked).toBe(false)

		await input.setValue(true)
		expect(wrapper.props('modelValue')).toBe(true)
	})

	it('applies disabled', () => {
		const wrapper = mount(FormCheckbox, {
			props: {label: 'X', disabled: true},
		})
		expect(wrapper.find('input').attributes('disabled')).toBe('')
	})

	it('renders slot content instead of label prop when slot is provided', () => {
		const wrapper = mount(FormCheckbox, {
			slots: {default: '<span>Custom <b>content</b></span>'},
		})
		expect(wrapper.find('label.checkbox').html()).toContain('<b>content</b>')
	})
})
