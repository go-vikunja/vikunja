import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {mount, type VueWrapper} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'

import DatepickerInline from './DatepickerInline.vue'
import {i18n as globalI18n} from '@/i18n'
import {useAuthStore} from '@/stores/auth'
import en from '@/i18n/lang/en.json'

const FROZEN_NOW = new Date('2026-09-07T12:00:00Z')

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function setPersianSettings() {
	useAuthStore().setUserSettings({
		timezone: 'Asia/Tehran',
		frontendSettings: {defaultDueTime: '09:00'},
	} as never)
}

function mountInline(modelValue: Date | null = null) {
	return mount(DatepickerInline, {
		props: {modelValue},
		global: {
			plugins: [i18n],
			stubs: {
				'flat-pickr': true,
				Icon: true,
			},
		},
	})
}

function shortcutButtons(wrapper: VueWrapper) {
	return wrapper.findAll('.datepicker__quick-select-date')
}

describe('DatepickerInline Jalali picker', () => {
	beforeEach(async () => {
		setActivePinia(createPinia())
		await import('dayjs/locale/fa')
		globalI18n.global.locale.value = 'fa-IR'
		setPersianSettings()
		vi.useFakeTimers()
		vi.setSystemTime(FROZEN_NOW)
	})

	afterEach(() => {
		vi.useRealTimers()
		globalI18n.global.locale.value = 'en'
	})

	it('renders the Jalali grid instead of flatpickr', () => {
		const wrapper = mountInline()

		expect(wrapper.find('.jalali-calendar').exists()).toBe(true)
		expect(wrapper.find('flat-pickr-stub').exists()).toBe(false)
	})

	it('selecting a day emits the correct Gregorian instant', async () => {
		const wrapper = mountInline()

		await wrapper.findAll('.jalali-calendar__day')
			.find((button) => button.text() === '۲۰')!
			.trigger('click')

		const emitted = wrapper.emitted('update:modelValue')?.pop()?.[0] as Date
		// Shahrivar 20, 09:00 in Asia/Tehran.
		expect(emitted.toISOString()).toBe('2026-09-11T05:30:00.000Z')
	})

	it('shortcuts select the correct instant in the user timezone', async () => {
		const wrapper = mountInline()
		const buttons = shortcutButtons(wrapper)

		// Tomorrow: Shahrivar 17, 09:00 Asia/Tehran.
		await buttons[1].trigger('click')
		expect((wrapper.emitted('update:modelValue')?.pop()?.[0] as Date).toISOString())
			.toBe('2026-09-08T05:30:00.000Z')

		// Today: Shahrivar 16, 09:00 Asia/Tehran.
		await buttons[0].trigger('click')
		expect((wrapper.emitted('update:modelValue')?.pop()?.[0] as Date).toISOString())
			.toBe('2026-09-07T05:30:00.000Z')
	})

	it('keeps the selected time when picking another day', async () => {
		const wrapper = mountInline(new Date('2026-09-07T12:00:00Z'))

		await wrapper.findAll('.jalali-calendar__day')
			.find((button) => button.text() === '۱۷')!
			.trigger('click')

		const emitted = wrapper.emitted('update:modelValue')?.pop()?.[0] as Date
		expect(emitted.toISOString()).toBe('2026-09-08T12:00:00.000Z')
	})
})

describe('DatepickerInline Gregorian picker', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		globalI18n.global.locale.value = 'en'
	})

	afterEach(() => {
		globalI18n.global.locale.value = 'en'
	})

	it('renders flatpickr instead of the Jalali grid', () => {
		const wrapper = mountInline()

		expect(wrapper.find('flat-pickr-stub').exists()).toBe(true)
		expect(wrapper.find('.jalali-calendar').exists()).toBe(false)
	})
})
