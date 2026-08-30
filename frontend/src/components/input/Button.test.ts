import {describe, it, expect} from 'vitest'
import {mount} from '@vue/test-utils'
import Button from './Button.vue'

describe('Button', () => {
	it('wraps long labels by default', () => {
		const wrapper = mount(Button, {slots: {default: 'Alle Benachrichtigungen als gelesen markieren'}})
		expect(wrapper.attributes('style')).toContain('--button-white-space: break-spaces')
	})

	it('does not wrap when wrap is false', () => {
		const wrapper = mount(Button, {props: {wrap: false}})
		expect(wrapper.attributes('style')).toContain('--button-white-space: nowrap')
	})

	it('has a shadow by default', () => {
		const wrapper = mount(Button)
		expect(wrapper.classes()).not.toContain('has-no-shadow')
	})

	it('drops the shadow when shadow is false', () => {
		const wrapper = mount(Button, {props: {shadow: false}})
		expect(wrapper.classes()).toContain('has-no-shadow')
	})

	it('uses the primary variant by default', () => {
		const wrapper = mount(Button)
		expect(wrapper.classes()).toContain('is-primary')
	})

	it('marks tertiary buttons as shadowless', () => {
		const wrapper = mount(Button, {props: {variant: 'tertiary'}})
		expect(wrapper.classes()).toContain('is-text')
		expect(wrapper.classes()).toContain('has-no-shadow')
	})
})
