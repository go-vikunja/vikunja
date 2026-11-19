import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {nextTick} from 'vue'
import {flushPromises, mount} from '@vue/test-utils'
import {createPinia, setActivePinia} from 'pinia'
import {createI18n} from 'vue-i18n'

import en from '@/i18n/lang/en.json'
import Datepicker from './Datepicker.vue'
import DatepickerInline from './DatepickerInline.vue'

const {isMobile, flip, shift, computePosition} = vi.hoisted(() => ({
	isMobile: {value: false},
	flip: vi.fn(() => ({name: 'flip'})),
	shift: vi.fn(() => ({name: 'shift'})),
	computePosition: vi.fn(async () => ({
		x: 12,
		y: 34,
	})),
}))

vi.mock('@/composables/useIsMobile', async () => {
	const {ref} = await import('vue')
	return {useIsMobile: () => ref(isMobile.value)}
})
vi.mock('@floating-ui/dom', () => ({
	computePosition,
	flip,
	shift,
	autoUpdate: (_anchor: unknown, _popup: unknown, update: () => void) => {
		update()
		return () => {}
	},
	offset: () => ({name: 'offset'}),
	size: () => ({name: 'size'}),
}))

const i18n = createI18n({
	legacy: false,
	locale: 'en',
	messages: {en},
})

function mountPicker(compact?: boolean, showShortcuts = true) {
	return mount(Datepicker, {
		props: {
			modelValue: null,
			compact,
			showShortcuts,
		},
		attachTo: document.body,
		global: {
			plugins: [i18n],
			stubs: {
				DatepickerInline: true,
				XButton: true,
				Icon: true,
			},
			directives: {
				tooltip: {},
				cy: {},
			},
		},
	})
}

beforeEach(() => {
	setActivePinia(createPinia())
	isMobile.value = false
	vi.clearAllMocks()
})

afterEach(() => {
	document.body.innerHTML = ''
})

describe('Datepicker layout', () => {
	it.each([
		[true, 'chips', true, 'bestFit'],
		[false, 'sidebar', false, 'initialPlacement'],
		[undefined, 'sidebar', false, 'initialPlacement'],
	] as const)('with compact %s uses the %s layout', async (compact, layout, narrow, fallbackStrategy) => {
		const wrapper = mountPicker(compact)
		wrapper.vm.open()
		await flushPromises()

		expect(wrapper.findComponent(DatepickerInline).props('shortcutsLayout')).toBe(layout)
		expect(wrapper.find('.datepicker-popup').classes('datepicker-popup--compact')).toBe(narrow)
		expect(flip).toHaveBeenCalledWith({
			padding: {
				top: 72,
				right: 8,
				bottom: 8,
				left: 8,
			},
			fallbackStrategy,
		})
		expect(shift).toHaveBeenCalledWith({
			padding: {
				top: 72,
				right: 8,
				bottom: 8,
				left: 8,
			},
			crossAxis: true,
		})
		wrapper.unmount()
	})

	it('keeps the settings picker without shortcuts non-compact', async () => {
		const wrapper = mountPicker(undefined, false)
		await wrapper.find('button.show').trigger('click')
		await flushPromises()

		expect(wrapper.find('.datepicker-popup--no-shortcuts').exists()).toBe(true)
		expect(wrapper.find('.datepicker-popup--compact').exists()).toBe(false)
		expect(wrapper.findComponent(DatepickerInline).props('showShortcuts')).toBe(false)
		wrapper.unmount()
	})

	it.each([true, false])('keeps the phone bottom sheet with compact %s', async compact => {
		isMobile.value = true
		const wrapper = mountPicker(compact)
		wrapper.vm.open()
		await flushPromises()

		expect(document.querySelector('.bottom-sheet')).not.toBeNull()
		expect(document.querySelector('.popup')).toBeNull()
		expect(computePosition).not.toHaveBeenCalled()
		const picker = wrapper.findComponent(DatepickerInline)
		expect(picker.props('shortcutsLayout')).toBe('chips')
		expect(picker.props('large')).toBe(true)
		picker.vm.$emit('confirm')
		await nextTick()
		expect(wrapper.emitted('close')).toEqual([[false]])
		wrapper.unmount()
	})
})
