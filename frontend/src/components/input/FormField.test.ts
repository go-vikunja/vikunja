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
				'onUpdate:modelValue': (val: string | number) => wrapper.setProps({modelValue: val}),
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

	it('forwards @keyup.enter to inner input', async () => {
		const onSubmit = vi.fn()

		const wrapper = mount({
			components: {FormField},
			template: `<FormField @keyup.enter="onSubmit" />`,
			setup() {
				return {onSubmit}
			},
		})

		const input = wrapper.find('input')

		// Enter key should trigger the handler
		await input.trigger('keyup', {key: 'Enter'})
		expect(onSubmit).toHaveBeenCalledTimes(1)

		// Other keys should not trigger the handler
		await input.trigger('keyup', {key: 'a'})
		expect(onSubmit).toHaveBeenCalledTimes(1)
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
		// Mount both FormFields in the same Vue app to test useId uniqueness
		const wrapper = mount({
			components: {FormField},
			template: `
				<div>
					<FormField label="First Input" />
					<FormField label="Second Input" />
				</div>
			`,
		})

		const inputs = wrapper.findAll('input')
		const labels = wrapper.findAll('label')

		const id1 = inputs[0].attributes('id')
		const id2 = inputs[1].attributes('id')

		expect(id1).toBeTruthy()
		expect(id2).toBeTruthy()
		expect(id1).not.toBe(id2)

		// Verify label linkage for both
		expect(labels[0].attributes('for')).toBe(id1)
		expect(labels[1].attributes('for')).toBe(id2)
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

	it('renders two-col layout with wrapping label', () => {
		const wrapper = mount(FormField, {
			props: {label: 'Name', layout: 'two-col'},
			slots: {
				default: '<input class="input" />',
			},
		})
		const label = wrapper.find('label.two-col')
		expect(label.exists()).toBe(true)
		expect(label.find('span').text()).toBe('Name')
		expect(label.find('input.input').exists()).toBe(true)
	})

	it('two-col layout exposes id via slot scope', () => {
		const wrapper = mount({
			components: {FormField},
			template: `
				<FormField label="X" layout="two-col" id="custom-id" v-slot="{id}">
					<input :id="id" />
				</FormField>
			`,
		})
		expect(wrapper.find('input').attributes('id')).toBe('custom-id')
	})

	it('two-col layout omits the for attribute so implicit nesting labels any slotted control', () => {
		const wrapper = mount(FormField, {
			props: {label: 'Name', layout: 'two-col'},
			slots: {
				default: '<input id="some-generated-id" />',
			},
		})
		const label = wrapper.find('label.two-col')
		// for="" would mismatch the slotted control's id; rely on the label wrapping instead.
		expect(label.attributes('for')).toBeUndefined()
		expect(label.find('input').exists()).toBe(true)
	})

	it('renders the error message in two-col layout', () => {
		const wrapper = mount(FormField, {
			props: {label: 'Name', layout: 'two-col', error: 'Required'},
		})
		const help = wrapper.find('p.help.is-danger')
		expect(help.exists()).toBe(true)
		expect(help.text()).toBe('Required')
	})

	it('renders the addon slot in two-col layout', () => {
		const wrapper = mount(FormField, {
			props: {label: 'Name', layout: 'two-col'},
			slots: {
				addon: '<button>Copy</button>',
			},
		})
		expect(wrapper.find('.field.has-addons').exists()).toBe(true)
		expect(wrapper.find('button').text()).toBe('Copy')
	})
})
