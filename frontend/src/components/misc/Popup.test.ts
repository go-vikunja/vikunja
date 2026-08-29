import {describe, it, expect} from 'vitest'
import {mount} from '@vue/test-utils'
import {h, nextTick} from 'vue'
import Popup from './Popup.vue'

const slots = {
	trigger: (params: {toggle: () => boolean}) => h('button', {class: 'trigger', onClick: params.toggle}, 'open'),
	content: (params: {close: () => void}) => h('button', {class: 'inner', onClick: params.close}, 'inner'),
}

const escapeEvent = () => new KeyboardEvent('keydown', {key: 'Escape', bubbles: true, cancelable: true})

describe('Popup', () => {
	it('marks the content wrapper inert while closed', () => {
		const wrapper = mount(Popup, {slots})
		expect(wrapper.find('.popup').attributes('inert')).toBe('')
	})

	it('removes inert from the content wrapper when open', async () => {
		const wrapper = mount(Popup, {slots})

		await wrapper.setProps({open: true})

		expect(wrapper.find('.popup').attributes('inert')).toBeUndefined()
	})

	it('leaves the trigger outside the inert wrapper', () => {
		const wrapper = mount(Popup, {slots})
		expect(wrapper.find('.popup .trigger').exists()).toBe(false)
		expect(wrapper.find('.trigger').exists()).toBe(true)
	})

	it('returns focus to the trigger before inert is applied when closing from inside', async () => {
		const trigger = document.createElement('button')
		document.body.append(trigger)
		trigger.focus()

		const wrapper = mount(Popup, {props: {open: true}, attachTo: document.body})

		const popup = wrapper.find('.popup').element
		let inertWhenFocusReturned: boolean | null = null
		trigger.addEventListener('focus', () => {
			inertWhenFocusReturned = popup.hasAttribute('inert')
		})

		const inner = document.createElement('button')
		popup.append(inner)
		inner.focus()
		expect(document.activeElement).toBe(inner)

		inertWhenFocusReturned = null
		await wrapper.setProps({open: false})

		expect(document.activeElement).toBe(trigger)
		// focus lands back on the trigger before the DOM patch sets inert
		expect(inertWhenFocusReturned).toBe(false)

		trigger.remove()
		wrapper.unmount()
	})

	it('returns focus to the trigger when the closing click already blurred focus to the body', async () => {
		const trigger = document.createElement('button')
		document.body.append(trigger)
		trigger.focus()

		const wrapper = mount(Popup, {props: {open: true}, attachTo: document.body})

		const inner = document.createElement('button')
		wrapper.find('.popup').element.append(inner)
		inner.focus()
		expect(document.activeElement).toBe(inner)

		// the browser blurs the focused control on mousedown, before onClickOutside closes the popup on click
		inner.blur()
		expect(document.activeElement).toBe(document.body)

		await wrapper.setProps({open: false})

		expect(document.activeElement).toBe(trigger)

		trigger.remove()
		wrapper.unmount()
	})

	it('does not steal focus when the closing click moved focus to another element', async () => {
		const trigger = document.createElement('button')
		const elsewhere = document.createElement('button')
		document.body.append(trigger, elsewhere)
		trigger.focus()

		const wrapper = mount(Popup, {props: {open: true}, attachTo: document.body})

		const inner = document.createElement('button')
		wrapper.find('.popup').element.append(inner)
		inner.focus()

		elsewhere.focus()
		await wrapper.setProps({open: false})

		expect(document.activeElement).toBe(elsewhere)

		trigger.remove()
		elsewhere.remove()
		wrapper.unmount()
	})

	it('does not pull focus back to the trigger when focus never entered the popup', async () => {
		const trigger = document.createElement('button')
		document.body.append(trigger)
		trigger.focus()

		const wrapper = mount(Popup, {props: {open: true}, attachTo: document.body})

		trigger.blur()
		expect(document.activeElement).toBe(document.body)

		await wrapper.setProps({open: false})

		expect(document.activeElement).toBe(document.body)

		trigger.remove()
		wrapper.unmount()
	})

	it('does not steal focus when the popup closes while focus is elsewhere', async () => {
		const trigger = document.createElement('button')
		const elsewhere = document.createElement('button')
		document.body.append(trigger, elsewhere)
		trigger.focus()

		const wrapper = mount(Popup, {props: {open: true}, attachTo: document.body})

		elsewhere.focus()
		await wrapper.setProps({open: false})

		expect(document.activeElement).toBe(elsewhere)

		trigger.remove()
		elsewhere.remove()
		wrapper.unmount()
	})

	it('closes on escape and returns focus to the trigger', async () => {
		const trigger = document.createElement('button')
		document.body.append(trigger)
		trigger.focus()

		const wrapper = mount(Popup, {props: {open: true}, attachTo: document.body})

		const inner = document.createElement('button')
		wrapper.find('.popup').element.append(inner)
		inner.focus()

		const event = escapeEvent()
		inner.dispatchEvent(event)
		await nextTick()

		expect(wrapper.emitted('update:open')).toEqual([[false]])
		expect(wrapper.find('.popup').attributes('inert')).toBe('')
		expect(document.activeElement).toBe(trigger)
		// a wrapping native <dialog> must not treat the same Escape as its own close request
		expect(event.defaultPrevented).toBe(true)

		trigger.remove()
		wrapper.unmount()
	})

	it('does not open a closed popup on escape', async () => {
		const wrapper = mount(Popup, {attachTo: document.body})

		const inner = document.createElement('button')
		wrapper.find('.popup').element.append(inner)
		inner.focus()

		inner.dispatchEvent(escapeEvent())
		await nextTick()

		expect(wrapper.emitted('update:open')).toBeUndefined()
		expect(wrapper.find('.popup').attributes('inert')).toBe('')

		wrapper.unmount()
	})

	it('ignores escape pressed outside the popup and its trigger', async () => {
		const trigger = document.createElement('button')
		const elsewhere = document.createElement('button')
		document.body.append(trigger, elsewhere)
		trigger.focus()

		const wrapper = mount(Popup, {props: {open: true}, attachTo: document.body})

		elsewhere.focus()
		elsewhere.dispatchEvent(escapeEvent())
		await nextTick()

		expect(wrapper.emitted('update:open')).toBeUndefined()

		trigger.remove()
		elsewhere.remove()
		wrapper.unmount()
	})

	it('leaves escape alone when a control inside already handled it', async () => {
		const wrapper = mount(Popup, {props: {open: true}, attachTo: document.body})

		const inner = document.createElement('button')
		wrapper.find('.popup').element.append(inner)
		inner.addEventListener('keydown', event => event.preventDefault())
		inner.focus()

		inner.dispatchEvent(escapeEvent())
		await nextTick()

		expect(wrapper.emitted('update:open')).toBeUndefined()

		wrapper.unmount()
	})

	it('closes an open popup when its trigger is clicked', async () => {
		const wrapper = mount(Popup, {props: {open: true}, slots, attachTo: document.body})

		await wrapper.find('.trigger').trigger('click')

		expect(wrapper.emitted('update:open')).toEqual([[false]])
		expect(wrapper.find('.popup').attributes('inert')).toBe('')

		wrapper.unmount()
	})

	it('opens a closed popup when its trigger is clicked', async () => {
		const wrapper = mount(Popup, {slots, attachTo: document.body})

		await wrapper.find('.trigger').trigger('click')

		expect(wrapper.emitted('update:open')).toEqual([[true]])
		expect(wrapper.find('.popup').attributes('inert')).toBeUndefined()

		wrapper.unmount()
	})
})
