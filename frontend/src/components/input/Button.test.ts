import {describe, it, expect, vi} from 'vitest'
import {mount} from '@vue/test-utils'
import Button from './Button.vue'
import {LOADING_TIMEOUT} from '@/stores/helper'

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

	it('keeps a loading button focusable but inert', async () => {
		const onClick = vi.fn()
		const wrapper = mount(Button, {
			props: {loading: true},
			attrs: {onClick},
			attachTo: document.body,
		})

		const button = wrapper.get('button')
		expect(button.attributes('disabled')).toBeUndefined()
		expect(button.attributes('aria-disabled')).toBe('true')

		button.element.focus()
		await button.trigger('click')

		expect(onClick).not.toHaveBeenCalled()
		expect(document.activeElement).toBe(button.element)

		wrapper.unmount()
	})

	it('blocks clicks immediately but delays the spinner', async () => {
		vi.useFakeTimers()
		const wrapper = mount(Button, {props: {loading: true}})
		expect(wrapper.attributes('aria-disabled')).toBe('true')
		expect(wrapper.classes()).not.toContain('is-loading')

		await vi.advanceTimersByTimeAsync(LOADING_TIMEOUT)

		expect(wrapper.classes()).toContain('is-loading')
		vi.useRealTimers()
	})

	it('emits click when not loading', async () => {
		const onClick = vi.fn()
		const wrapper = mount(Button, {attrs: {onClick}})

		await wrapper.trigger('click')

		expect(onClick).toHaveBeenCalledTimes(1)
	})

	it('natively disables a disabled button', () => {
		const wrapper = mount(Button, {props: {disabled: true}})

		expect(wrapper.attributes('disabled')).toBe('')
		expect(wrapper.attributes('aria-disabled')).toBeUndefined()
	})
})
