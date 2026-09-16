import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {mount, flushPromises} from '@vue/test-utils'
import {h} from 'vue'

const {isMobile, floatingUi} = vi.hoisted(() => ({
	isMobile: {value: false},
	floatingUi: {
		computePosition: vi.fn(async () => ({x: 12, y: 34})),
		autoUpdate: vi.fn((anchor: unknown, popup: unknown, update: () => void) => {
			update()
			return () => {}
		}),
	},
}))

vi.mock('@/composables/useIsMobile', () => ({useIsMobile: () => isMobile}))
vi.mock('@floating-ui/dom', () => ({
	computePosition: floatingUi.computePosition,
	autoUpdate: floatingUi.autoUpdate,
	flip: () => ({name: 'flip'}),
	offset: () => ({name: 'offset'}),
	shift: () => ({name: 'shift'}),
	size: () => ({name: 'size'}),
}))

import Popup from './Popup.vue'

// happy-dom implements neither the Popover API nor :popover-open, so mirror the state the component
// drives into an attribute the tests can assert on.
const POPOVER_OPEN = 'popover-open'

function dispatchToggle(element: HTMLElement, newState: 'open' | 'closed') {
	element.dispatchEvent(Object.assign(new Event('toggle'), {newState}))
}

HTMLElement.prototype.showPopover = function showPopover(this: HTMLElement) {
	this.setAttribute(POPOVER_OPEN, '')
	dispatchToggle(this, 'open')
}

HTMLElement.prototype.hidePopover = function hidePopover(this: HTMLElement) {
	this.removeAttribute(POPOVER_OPEN)
	dispatchToggle(this, 'closed')
}

// What the browser does on an outside click or Escape.
function lightDismiss(element: HTMLElement) {
	element.removeAttribute(POPOVER_OPEN)
	dispatchToggle(element, 'closed')
}

const slots = {
	trigger: (params: {toggle: () => boolean}) => h('button', {class: 'trigger', onClick: params.toggle}, 'open'),
	content: (params: {close: () => void}) => h('button', {class: 'inner', onClick: params.close}, 'inner'),
}

const mountOptions = {
	slots,
	attachTo: document.body,
	global: {mocks: {$t: (key: string) => key}},
}

let showPopoverSpy: ReturnType<typeof vi.spyOn>
let hidePopoverSpy: ReturnType<typeof vi.spyOn>

beforeEach(() => {
	isMobile.value = false
	floatingUi.computePosition.mockClear()
	floatingUi.autoUpdate.mockClear()
	showPopoverSpy = vi.spyOn(HTMLElement.prototype, 'showPopover')
	hidePopoverSpy = vi.spyOn(HTMLElement.prototype, 'hidePopover')
})

afterEach(() => {
	showPopoverSpy.mockRestore()
	hidePopoverSpy.mockRestore()
	document.body.innerHTML = ''
})

describe('Popup', () => {
	it('renders the content in a popover the trigger stays out of', () => {
		const wrapper = mount(Popup, mountOptions)

		expect(wrapper.find('.popup').attributes('popover')).toBe('auto')
		expect(wrapper.find('.popup .inner').exists()).toBe(true)
		expect(wrapper.find('.popup .trigger').exists()).toBe(false)
		expect(wrapper.find('.trigger').exists()).toBe(true)

		wrapper.unmount()
	})

	it('shows the popover when the trigger is clicked', async () => {
		const wrapper = mount(Popup, mountOptions)

		await wrapper.find('.trigger').trigger('click')

		expect(wrapper.emitted('update:open')).toEqual([[true]])
		expect(showPopoverSpy).toHaveBeenCalledTimes(1)
		expect(wrapper.find('.popup').attributes(POPOVER_OPEN)).toBe('')

		wrapper.unmount()
	})

	it('hides the popover when the trigger is clicked again', async () => {
		const wrapper = mount(Popup, {...mountOptions, props: {open: true}})
		await flushPromises()

		await wrapper.find('.trigger').trigger('click')

		expect(wrapper.emitted('update:open')).toEqual([[false]])
		expect(hidePopoverSpy).toHaveBeenCalledTimes(1)
		expect(wrapper.find('.popup').attributes(POPOVER_OPEN)).toBeUndefined()

		wrapper.unmount()
	})

	it('hides the popover when the content closes it', async () => {
		const wrapper = mount(Popup, {...mountOptions, props: {open: true}})
		await flushPromises()

		await wrapper.find('.inner').trigger('click')

		expect(wrapper.emitted('update:open')).toEqual([[false]])
		expect(wrapper.find('.popup').attributes(POPOVER_OPEN)).toBeUndefined()

		wrapper.unmount()
	})

	it('follows the open prop in both directions', async () => {
		const wrapper = mount(Popup, {...mountOptions, props: {open: false}})

		await wrapper.setProps({open: true})
		expect(showPopoverSpy).toHaveBeenCalledTimes(1)
		expect(wrapper.find('.popup').attributes(POPOVER_OPEN)).toBe('')

		await wrapper.setProps({open: false})
		expect(hidePopoverSpy).toHaveBeenCalledTimes(1)
		expect(wrapper.find('.popup').attributes(POPOVER_OPEN)).toBeUndefined()

		wrapper.unmount()
	})

	it('emits update:open false when the browser light-dismisses the popover', async () => {
		const wrapper = mount(Popup, {...mountOptions, props: {open: true}})
		await flushPromises()

		lightDismiss(wrapper.find('.popup').element as HTMLElement)
		await flushPromises()

		expect(wrapper.emitted('update:open')).toEqual([[false]])
		expect(hidePopoverSpy).not.toHaveBeenCalled()

		wrapper.unmount()
	})

	it('does not reopen when the trigger click follows a light dismiss', async () => {
		const wrapper = mount(Popup, {...mountOptions, props: {open: true}})
		await flushPromises()

		lightDismiss(wrapper.find('.popup').element as HTMLElement)
		await wrapper.find('.trigger').trigger('click')

		expect(wrapper.emitted('update:open')).toEqual([[false]])
		expect(wrapper.find('.popup').attributes(POPOVER_OPEN)).toBeUndefined()

		wrapper.unmount()
	})

	it('renders a sheet instead of a popover on mobile', async () => {
		isMobile.value = true
		const wrapper = mount(Popup, {...mountOptions, props: {open: true, sheetOnMobile: true}})
		await flushPromises()

		expect(wrapper.find('.popup').exists()).toBe(false)
		expect(showPopoverSpy).not.toHaveBeenCalled()
		expect(document.querySelector('dialog.modal-dialog')).not.toBeNull()

		wrapper.unmount()
	})

	it('positions the popup with floating-ui when anchored', async () => {
		const anchor = document.createElement('button')
		document.body.append(anchor)

		const wrapper = mount(Popup, {
			...mountOptions,
			props: {open: true, anchor, placement: 'bottom-start' as const},
		})
		await flushPromises()

		expect(floatingUi.computePosition).toHaveBeenCalledWith(
			anchor,
			wrapper.find('.popup').element,
			expect.objectContaining({placement: 'bottom-start', strategy: 'fixed'}),
		)
		expect(wrapper.find('.popup').attributes('style')).toContain('left: 12px')

		wrapper.unmount()
	})

	it('does not position an unanchored popup', async () => {
		const wrapper = mount(Popup, {...mountOptions, props: {open: true}})
		await flushPromises()

		expect(floatingUi.computePosition).not.toHaveBeenCalled()

		wrapper.unmount()
	})
})
