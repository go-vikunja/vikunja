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
