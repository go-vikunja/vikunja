import {describe, it, expect} from 'vitest'
import {mount} from '@vue/test-utils'
import FormSelect from './FormSelect.vue'

describe('FormSelect', () => {
	it('renders the Bulma select wrapper and a native select', () => {
		const wrapper = mount(FormSelect)
		expect(wrapper.find('div.select').exists()).toBe(true)
		expect(wrapper.find('div.select > select').exists()).toBe(true)
	})

	it('renders options from the default slot', () => {
		const wrapper = mount(FormSelect, {
			slots: {
				default: '<option value="a">A</option><option value="b">B</option>',
			},
		})
		expect(wrapper.findAll('option').length).toBe(2)
	})

	it('supports v-model with string values', async () => {
		const wrapper = mount(FormSelect, {
			props: {
				modelValue: 'a',
				'onUpdate:modelValue': (val: string | number) => wrapper.setProps({modelValue: val}),
			},
			slots: {
				default: '<option value="a">A</option><option value="b">B</option>',
			},
		})
		const select = wrapper.find('select')
		expect((select.element as HTMLSelectElement).value).toBe('a')

		await select.setValue('b')
		expect(wrapper.props('modelValue')).toBe('b')
	})

	it('preserves numeric type in v-model when modelValue is a number', async () => {
		const wrapper = mount(FormSelect, {
			props: {
				modelValue: 1,
				'onUpdate:modelValue': (val: string | number) => wrapper.setProps({modelValue: val}),
			},
			slots: {
				default: '<option value="1">One</option><option value="2">Two</option>',
			},
		})
		await wrapper.find('select').setValue('2')
		expect(wrapper.props('modelValue')).toBe(2)
	})

	it('coerces to number when the .number modifier is set even if modelValue starts null', async () => {
		const wrapper = mount(FormSelect, {
			props: {
				modelValue: null,
				modelModifiers: {number: true},
				'onUpdate:modelValue': (val: string | number) => wrapper.setProps({modelValue: val}),
			},
			slots: {
				default: '<option value="1">One</option><option value="2">Two</option>',
			},
		})
		await wrapper.find('select').setValue('2')
		expect(wrapper.props('modelValue')).toBe(2)
		expect(typeof wrapper.props('modelValue')).toBe('number')
	})

	it('applies is-loading on the wrapper when loading', () => {
		const wrapper = mount(FormSelect, {props: {loading: true}})
		expect(wrapper.find('div.select').classes()).toContain('is-loading')
	})

	it('applies disabled to the native select', () => {
		const wrapper = mount(FormSelect, {props: {disabled: true}})
		expect(wrapper.find('select').attributes('disabled')).toBe('')
	})

	it('uses an explicit id prop when given, otherwise generates one', () => {
		const withProp = mount(FormSelect, {props: {id: 'explicit'}})
		expect(withProp.find('select').attributes('id')).toBe('explicit')

		const standalone = mount(FormSelect)
		expect(standalone.find('select').attributes('id')).toBeTruthy()
	})

	it('renders error message when error prop is set', () => {
		const wrapper = mount(FormSelect, {props: {error: 'Pick one'}})
		expect(wrapper.find('p.help.is-danger').text()).toBe('Pick one')
	})

	it('does not render error message when error is null or empty', () => {
		const nullErr = mount(FormSelect, {props: {error: null}})
		expect(nullErr.find('p.help.is-danger').exists()).toBe(false)

		const emptyErr = mount(FormSelect, {props: {error: ''}})
		expect(emptyErr.find('p.help.is-danger').exists()).toBe(false)
	})

	it('renders options from the options prop with object entries', () => {
		const wrapper = mount(FormSelect, {
			props: {
				options: [
					{value: 'a', label: 'Alpha'},
					{value: 'b', label: 'Bravo'},
				],
			},
		})
		const options = wrapper.findAll('option')
		expect(options).toHaveLength(2)
		expect(options[0].attributes('value')).toBe('a')
		expect(options[0].text()).toBe('Alpha')
		expect(options[1].attributes('value')).toBe('b')
		expect(options[1].text()).toBe('Bravo')
	})

	it('coerces primitive options into value/label pairs', () => {
		const wrapper = mount(FormSelect, {
			props: {options: ['one', 'two']},
		})
		const options = wrapper.findAll('option')
		expect(options).toHaveLength(2)
		expect(options[0].attributes('value')).toBe('one')
		expect(options[0].text()).toBe('one')
	})

	it('marks an option as disabled when disabled: true is given', () => {
		const wrapper = mount(FormSelect, {
			props: {
				options: [
					{value: 'a', label: 'Alpha'},
					{value: 'b', label: 'Bravo', disabled: true},
				],
			},
		})
		const options = wrapper.findAll('option')
		expect(options[0].attributes('disabled')).toBeUndefined()
		expect(options[1].attributes('disabled')).toBe('')
	})

	it('falls back to the default slot when options prop is not given', () => {
		const wrapper = mount(FormSelect, {
			slots: {
				default: '<option value="x">From slot</option>',
			},
		})
		const options = wrapper.findAll('option')
		expect(options).toHaveLength(1)
		expect(options[0].text()).toBe('From slot')
	})

	it('does not bind value when modelValue is undefined', () => {
		const wrapper = mount(FormSelect, {
			slots: {
				default: '<option value="">--</option><option value="a">A</option><option value="b">B</option>',
			},
		})
		const select = wrapper.find('select')
		// Forcing :value="undefined" would break the native default-to-first-option behavior.
		expect((select.element as HTMLSelectElement).value).toBe('')
	})

	it('ignores the slot when options prop is given', () => {
		const wrapper = mount(FormSelect, {
			props: {options: [{value: 'a', label: 'From prop'}]},
			slots: {
				default: '<option value="x">From slot</option>',
			},
		})
		const options = wrapper.findAll('option')
		expect(options).toHaveLength(1)
		expect(options[0].text()).toBe('From prop')
	})
})
