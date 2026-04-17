import {describe, it, expect, vi} from 'vitest'
import {mount} from '@vue/test-utils'
import FormInput from './FormInput.vue'

describe('FormInput', () => {
	it('renders a Bulma-classed input', () => {
		const wrapper = mount(FormInput)
		const input = wrapper.find('input')
		expect(input.exists()).toBe(true)
		expect(input.classes()).toContain('input')
	})

	it('supports v-model', async () => {
		const wrapper = mount(FormInput, {
			props: {
				modelValue: 'hello',
				'onUpdate:modelValue': (val: string | number) => wrapper.setProps({modelValue: val}),
			},
		})
		const input = wrapper.find('input')
		expect(input.element.value).toBe('hello')

		await input.setValue('world')
		expect(wrapper.props('modelValue')).toBe('world')
	})

	it('preserves numeric type in v-model when modelValue is a number', async () => {
		const wrapper = mount(FormInput, {
			props: {
				modelValue: 42,
				'onUpdate:modelValue': (val: number | string) => wrapper.setProps({modelValue: val as number}),
			},
		})
		await wrapper.find('input').setValue('7')
		expect(wrapper.props('modelValue')).toBe(7)
	})

	it('coerces to number when the .number modifier is set even if modelValue starts null', async () => {
		const wrapper = mount(FormInput, {
			props: {
				modelValue: null,
				modelModifiers: {number: true},
				'onUpdate:modelValue': (val: number | string) => wrapper.setProps({modelValue: val as number}),
			},
		})
		await wrapper.find('input').setValue('42')
		expect(wrapper.props('modelValue')).toBe(42)
		expect(typeof wrapper.props('modelValue')).toBe('number')
	})

	it('applies is-loading class when loading', () => {
		const wrapper = mount(FormInput, {props: {loading: true}})
		expect(wrapper.find('input').classes()).toContain('is-loading')
	})

	it('applies disabled class and attribute when disabled', () => {
		const wrapper = mount(FormInput, {props: {disabled: true}})
		const input = wrapper.find('input')
		expect(input.classes()).toContain('disabled')
		expect(input.attributes('disabled')).toBe('')
	})

	it('uses an explicit id prop when given', () => {
		const wrapper = mount(FormInput, {props: {id: 'my-id'}})
		expect(wrapper.find('input').attributes('id')).toBe('my-id')
	})

	it('generates a unique id when no id prop is given', () => {
		const wrapper = mount({
			components: {FormInput},
			template: '<div><FormInput /><FormInput /></div>',
		})
		const inputs = wrapper.findAll('input')
		const id1 = inputs[0].attributes('id')
		const id2 = inputs[1].attributes('id')
		expect(id1).toBeTruthy()
		expect(id2).toBeTruthy()
		expect(id1).not.toBe(id2)
	})

	it('forwards $attrs (type, placeholder, autocomplete) to the input', () => {
		const wrapper = mount(FormInput, {
			attrs: {
				type: 'email',
				placeholder: 'Enter email',
				autocomplete: 'email',
			},
		})
		const input = wrapper.find('input')
		expect(input.attributes('type')).toBe('email')
		expect(input.attributes('placeholder')).toBe('Enter email')
		expect(input.attributes('autocomplete')).toBe('email')
	})

	it('forwards event listeners', async () => {
		const onKeyup = vi.fn()
		const wrapper = mount(FormInput, {attrs: {onKeyup}})
		await wrapper.find('input').trigger('keyup', {key: 'Enter'})
		expect(onKeyup).toHaveBeenCalledTimes(1)
	})

	it('renders error message when error prop is set', () => {
		const wrapper = mount(FormInput, {props: {error: 'Required'}})
		const help = wrapper.find('p.help.is-danger')
		expect(help.exists()).toBe(true)
		expect(help.text()).toBe('Required')
	})

	it('does not render error message when error is null or empty', () => {
		const nullErr = mount(FormInput, {props: {error: null}})
		expect(nullErr.find('p.help.is-danger').exists()).toBe(false)

		const emptyErr = mount(FormInput, {props: {error: ''}})
		expect(emptyErr.find('p.help.is-danger').exists()).toBe(false)
	})

	it('exposes value and focus()', async () => {
		const wrapper = mount(FormInput)
		await wrapper.find('input').setValue('test value')
		expect(wrapper.vm.value).toBe('test value')
		expect(() => wrapper.vm.focus()).not.toThrow()
	})
})
