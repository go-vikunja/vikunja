import {afterEach, beforeEach, describe, expect, it} from 'vitest'
import {mount} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import DatepickerWithValues from './DatepickerWithValues.vue'
import {i18n as globalI18n} from '@/i18n'
import {useAuthStore} from '@/stores/auth'
import en from '@/i18n/lang/en.json'

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function mountValues(modelValue: string | Date | null) {
	return mount(DatepickerWithValues, {
		props: {modelValue},
		global: {
			plugins: [i18n],
			stubs: {
				Modal: true,
				XButton: true,
				BaseButton: true,
				Popup: {template: '<div><slot name="content" :is-open="true" /></div>'},
				'flat-pickr': true,
				DatemathHelp: true,
			},
		},
	})
}

function dayButton(wrapper: ReturnType<typeof mountValues>, label: string) {
	const found = wrapper.findAll('.jalali-calendar__day').find((button) => button.text() === label)
	if (!found) {
		throw new Error(`day button ${label} not found`)
	}
	return found
}

describe('DatepickerWithValues Jalali picker', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		globalI18n.global.locale.value = 'fa-IR'
		useAuthStore().setUserSettings({timezone: 'Asia/Tehran'} as never)
	})

	afterEach(() => {
		globalI18n.global.locale.value = 'en'
	})

	it('renders the Jalali grid instead of flatpickr', () => {
		const wrapper = mountValues('2026-09-07T12:00:00Z')
		expect(wrapper.find('.jalali-calendar').exists()).toBe(true)
		expect(wrapper.find('flat-pickr-stub').exists()).toBe(false)
	})

	it('picks a Jalali date as a Gregorian ISO string', async () => {
		const wrapper = mountValues('2026-09-07T12:00:00Z')

		await dayButton(wrapper, '۲۰').trigger('click')

		// Shahrivar 20 keeps the 15:30 Asia/Tehran wall-clock time.
		expect(wrapper.emitted('update:modelValue')?.pop()?.[0]).toBe('2026-09-11T12:00:00.000Z')
	})

	it('opens on today for datemath values without crashing', () => {
		const wrapper = mountValues('now/w')
		expect(wrapper.find('.jalali-calendar').exists()).toBe(true)
	})
})

describe('DatepickerWithValues Jalali text input', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	afterEach(() => {
		globalI18n.global.locale.value = 'en'
	})

	function mountValueInput(timezone: string, locale = 'fa-IR') {
		globalI18n.global.locale.value = locale as never
		useAuthStore().setUserSettings({timezone} as never)
		return mount(DatepickerWithValues, {
			props: {modelValue: ''},
			global: {
				plugins: [i18n],
				stubs: {
					Modal: true,
					XButton: true,
					BaseButton: true,
					Popup: {template: '<div><slot name="content" :is-open="true" /></div>'},
					'flat-pickr': true,
					DatemathHelp: true,
				},
			},
		})
	}

	it.each([
		['Asia/Tehran', '2026-09-06T20:30:00.000Z'],
		['America/Los_Angeles', '2026-09-07T07:00:00.000Z'],
		['UTC', '2026-09-07T00:00:00.000Z'],
	])('parses typed Jalali in %s to Gregorian ISO (near-midnight)', async (timezone, expected) => {
		const wrapper = mountValueInput(timezone)
		await wrapper.find('input.input').setValue('1405/06/16')
		await wrapper.vm.$nextTick()
		expect(wrapper.emitted('update:modelValue')?.pop()?.[0]).toBe(expected)
	})

	it('parses Persian digits and HH:MM', async () => {
		const wrapper = mountValueInput('Asia/Tehran')
		await wrapper.find('input.input').setValue('۱۴۰۵/۰۶/۱۶ ۱۵:۳۰')
		await wrapper.vm.$nextTick()
		expect(wrapper.emitted('update:modelValue')?.pop()?.[0]).toBe('2026-09-07T12:00:00.000Z')
	})

	it('keeps typed datemath verbatim', async () => {
		const wrapper = mountValueInput('Asia/Tehran')
		await wrapper.find('input.input').setValue('now/w')
		await wrapper.vm.$nextTick()
		expect(wrapper.emitted('update:modelValue')?.pop()?.[0]).toBe('now/w')
	})

	it('keeps ISO verbatim', async () => {
		const wrapper = mountValueInput('Asia/Tehran')
		await wrapper.find('input.input').setValue('2026-09-07')
		await wrapper.vm.$nextTick()
		expect(wrapper.emitted('update:modelValue')?.pop()?.[0]).toBe('2026-09-07')
	})

	it('leaves ambiguous day-first on the Gregorian path', async () => {
		const wrapper = mountValueInput('Asia/Tehran')
		await wrapper.find('input.input').setValue('17/06/1405')
		await wrapper.vm.$nextTick()
		expect(wrapper.emitted('update:modelValue')?.pop()?.[0]).toBe('17/06/1405')
	})

	it.each([['en'], ['de']])('leaves Jalali-looking text untouched in %s', async (locale) => {
		const wrapper = mountValueInput('Asia/Tehran', locale)
		await wrapper.find('input.input').setValue('1405/06/16')
		await wrapper.vm.$nextTick()
		expect(wrapper.emitted('update:modelValue')?.pop()?.[0]).toBe('1405/06/16')
	})
})
