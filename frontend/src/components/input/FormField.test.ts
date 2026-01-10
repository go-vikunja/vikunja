import {describe, it, expect, vi} from 'vitest'
import {mount} from '@vue/test-utils'
import FormField from './FormField.vue'

describe('FormField', () => {
	it('renders simple input', () => {
		const wrapper = mount(FormField)
		expect(wrapper.find('.field').exists()).toBe(true)
		expect(wrapper.find('.control').exists()).toBe(true)
		expect(wrapper.find('input.input').exists()).toBe(true)
	})

	it('supports v-model binding', async () => {
		const wrapper = mount(FormField, {
			props: {
				modelValue: 'initial',
				'onUpdate:modelValue': (val: string) => wrapper.setProps({modelValue: val}),
			},
		})
		const input = wrapper.find('input')
		expect(input.element.value).toBe('initial')

		await input.setValue('updated')
		expect(wrapper.props('modelValue')).toBe('updated')
	})

	it('renders label when provided', () => {
		const wrapper = mount(FormField, {
			props: {label: 'Username'},
		})
		const label = wrapper.find('label.label')
		expect(label.exists()).toBe(true)
		expect(label.text()).toBe('Username')
	})

	it('does not render label when not provided', () => {
		const wrapper = mount(FormField)
		expect(wrapper.find('label.label').exists()).toBe(false)
	})

	it('displays error message when provided', () => {
		const wrapper = mount(FormField, {
			props: {error: 'This field is required'},
		})
		const help = wrapper.find('.help.is-danger')
		expect(help.exists()).toBe(true)
		expect(help.text()).toBe('This field is required')
	})

	it('does not display error message when error is null', () => {
		const wrapper = mount(FormField, {
			props: {error: null},
		})
		expect(wrapper.find('.help.is-danger').exists()).toBe(false)
	})

	it('does not display error message when error is empty string', () => {
		const wrapper = mount(FormField, {
			props: {error: ''},
		})
		expect(wrapper.find('.help.is-danger').exists()).toBe(false)
	})

	it('renders addon slot when provided', () => {
		const wrapper = mount(FormField, {
			slots: {
				addon: '<button>Copy</button>',
			},
		})
		expect(wrapper.find('.field.has-addons').exists()).toBe(true)
		expect(wrapper.find('.control.is-expanded').exists()).toBe(true)
		expect(wrapper.findAll('.control').length).toBe(2)
		expect(wrapper.find('button').text()).toBe('Copy')
	})

	it('renders custom input via default slot', () => {
		const wrapper = mount(FormField, {
			props: {label: 'Custom'},
			slots: {
				default: '<select><option>Option 1</option></select>',
			},
		})
		expect(wrapper.find('input').exists()).toBe(false)
		expect(wrapper.find('select').exists()).toBe(true)
	})

	it('passes attributes through to input', () => {
		const wrapper = mount(FormField, {
			attrs: {
				type: 'email',
				placeholder: 'Enter email',
				disabled: true,
				readonly: true,
				autocomplete: 'email',
			},
		})
		const input = wrapper.find('input')
		expect(input.attributes('type')).toBe('email')
		expect(input.attributes('placeholder')).toBe('Enter email')
		expect(input.attributes('disabled')).toBe('')
		expect(input.attributes('readonly')).toBe('')
		expect(input.attributes('autocomplete')).toBe('email')
	})

	it('forwards $attrs event listeners to inner input', async () => {
		const onKeyup = vi.fn()
		const onFocusout = vi.fn()

		const wrapper = mount(FormField, {
			props: {
				modelValue: 'test',
			},
			attrs: {
				onKeyup,
				onFocusout,
			},
		})

		const input = wrapper.find('input')

		await input.trigger('keyup', {key: 'Enter'})
		expect(onKeyup).toHaveBeenCalledTimes(1)

		await input.trigger('focusout')
		expect(onFocusout).toHaveBeenCalledTimes(1)
	})

	it('uses provided id for input', () => {
		const wrapper = mount(FormField, {
			props: {id: 'my-input', label: 'My Input'},
		})
		const input = wrapper.find('input')
		const label = wrapper.find('label')
		expect(input.attributes('id')).toBe('my-input')
		expect(label.attributes('for')).toBe('my-input')
	})

	it('generates unique id when not provided', () => {
		const wrapper = mount(FormField, {
			props: {label: 'My Input'},
		})
		const input = wrapper.find('input')
		const label = wrapper.find('label')
		const inputId = input.attributes('id')
		expect(inputId).toBeTruthy()
		expect(label.attributes('for')).toBe(inputId)
	})

	it('links label to input via for attribute', () => {
		const wrapper = mount(FormField, {
			props: {label: 'Test Label'},
		})
		const label = wrapper.find('label')
		const input = wrapper.find('input')
		expect(label.attributes('for')).toBe(input.attributes('id'))
	})

	it('exposes input value for direct access', async () => {
		const wrapper = mount(FormField)
		const input = wrapper.find('input')
		await input.setValue('test value')
		expect(wrapper.vm.value).toBe('test value')
	})
})
