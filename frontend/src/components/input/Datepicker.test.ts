import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {defineComponent, h, nextTick} from 'vue'
import {mount, type VueWrapper} from '@vue/test-utils'
import {createPinia, setActivePinia} from 'pinia'
import {createI18n} from 'vue-i18n'

import en from '@/i18n/lang/en.json'
import Datepicker from './Datepicker.vue'

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

// Shows the open state Datepicker hands to its popup, so the tests need neither a popover nor its content.
const popupOpen = vi.fn()
const PopupStub = defineComponent({
	props: {open: Boolean},
	setup(props, {slots, expose}) {
		expose({openPopup: popupOpen})
		return () => h('div', {class: 'popup-stub', 'data-open': String(props.open)}, [
			slots.trigger?.({toggle: () => true, isOpen: props.open, close: () => {}}),
			slots.content?.({toggle: () => true, isOpen: props.open, close: () => {}}),
		])
	},
})

type DatepickerVm = {open: () => void}

function mountPicker(disabled: boolean) {
	return mount(Datepicker, {
		props: {modelValue: null, disabled},
		attachTo: document.body,
		global: {
			plugins: [i18n],
			stubs: {Popup: PopupStub, DatepickerInline: true},
			directives: {tooltip: {}},
		},
	})
}

function isOpen(wrapper: VueWrapper) {
	return wrapper.find('.popup-stub').attributes('data-open') === 'true'
}

async function requestOpen(wrapper: VueWrapper) {
	;(wrapper.vm as unknown as DatepickerVm).open()
	await nextTick()
}

beforeEach(() => {
	setActivePinia(createPinia())
})

afterEach(() => {
	vi.useRealTimers()
	document.body.innerHTML = ''
})

describe('Datepicker open()', () => {
	it('opens at once when enabled', async () => {
		const wrapper = mountPicker(false)
		await requestOpen(wrapper)
		expect(isOpen(wrapper)).toBe(true)
		wrapper.unmount()
	})

	it('opens once it is enabled again when asked while disabled', async () => {
		const wrapper = mountPicker(true)
		await requestOpen(wrapper)
		expect(isOpen(wrapper)).toBe(false)

		await wrapper.setProps({disabled: false})
		expect(isOpen(wrapper)).toBe(true)
		wrapper.unmount()
	})

	it('opens only once for one request', async () => {
		const wrapper = mountPicker(true)
		await requestOpen(wrapper)
		await wrapper.setProps({disabled: false})
		expect(isOpen(wrapper)).toBe(true)

		wrapper.findComponent(PopupStub).vm.$emit('update:open', false)
		await nextTick()
		await wrapper.setProps({disabled: true})
		await wrapper.setProps({disabled: false})
		expect(isOpen(wrapper)).toBe(false)
		wrapper.unmount()
	})

	it.each([
		['a key press', () => window.dispatchEvent(new KeyboardEvent('keydown', {key: 'x'}))],
		['a pointer press', () => window.dispatchEvent(new Event('pointerdown'))],
	])('forgets the request after %s', async (_name, press) => {
		const wrapper = mountPicker(true)
		await requestOpen(wrapper)
		press()
		await wrapper.setProps({disabled: false})
		expect(isOpen(wrapper)).toBe(false)
		wrapper.unmount()
	})

	it('forgets the request after ten seconds', async () => {
		vi.useFakeTimers()
		const wrapper = mountPicker(true)
		await requestOpen(wrapper)
		vi.advanceTimersByTime(10_000)
		await wrapper.setProps({disabled: false})
		expect(isOpen(wrapper)).toBe(false)
		wrapper.unmount()
	})

	it('removes its listeners and timer when unmounted with a request pending', async () => {
		vi.useFakeTimers()
		const removed = vi.spyOn(window, 'removeEventListener')
		const wrapper = mountPicker(true)
		await requestOpen(wrapper)
		expect(vi.getTimerCount()).toBe(1)

		wrapper.unmount()

		expect(vi.getTimerCount()).toBe(0)
		expect(removed.mock.calls.map(([type, , capture]) => [type, capture])).toEqual(
			expect.arrayContaining([['keydown', true], ['pointerdown', true]]),
		)
		removed.mockRestore()
	})

	it('asks the popup itself to open when the parent still thinks it is open', async () => {
		popupOpen.mockClear()
		const wrapper = mountPicker(false)
		await requestOpen(wrapper)
		expect(isOpen(wrapper)).toBe(true)

		// The browser has light-dismissed the popup, but the popup reports that to the parent only on toggle.
		;(document.activeElement as HTMLElement | null)?.blur()
		await requestOpen(wrapper)
		await nextTick()
		expect(popupOpen).toHaveBeenCalledTimes(1)
		expect(document.activeElement).toBe(wrapper.find('.show').element)
		expect(wrapper.emitted('close')).toBeUndefined()
		expect(wrapper.emitted('closeOnChange')).toBeUndefined()
		wrapper.unmount()
	})

	it('keeps a changed date for one save when reopened before the popup reports its close', async () => {
		popupOpen.mockClear()
		const wrapper = mountPicker(false)
		await requestOpen(wrapper)
		wrapper.findComponent({name: 'DatepickerInline'}).vm.$emit('update:modelValue', new Date('2026-11-02T09:00:00Z'))
		await nextTick()

		// Escape, then the shortcut before the late toggle: still one open popup, no close yet.
		await requestOpen(wrapper)
		expect(popupOpen).toHaveBeenCalledTimes(1)
		expect(wrapper.emitted('closeOnChange')).toBeUndefined()

		wrapper.findComponent(PopupStub).vm.$emit('update:open', false)
		await nextTick()
		expect(wrapper.emitted('close')).toHaveLength(1)
		expect(wrapper.emitted('closeOnChange')).toHaveLength(1)
		wrapper.unmount()
	})

	it('asks the popup itself to open for a request kept while disabled, too', async () => {
		popupOpen.mockClear()
		const wrapper = mountPicker(false)
		await requestOpen(wrapper)
		await wrapper.setProps({disabled: true})
		await requestOpen(wrapper)
		await wrapper.setProps({disabled: false})
		expect(popupOpen).toHaveBeenCalledTimes(1)
		wrapper.unmount()
	})

	it('forgets the request when the popup state changes meanwhile', async () => {
		const wrapper = mountPicker(true)
		await requestOpen(wrapper)
		wrapper.findComponent(PopupStub).vm.$emit('update:open', true)
		await nextTick()
		wrapper.findComponent(PopupStub).vm.$emit('update:open', false)
		await nextTick()
		await wrapper.setProps({disabled: false})
		expect(isOpen(wrapper)).toBe(false)
		wrapper.unmount()
	})
})
