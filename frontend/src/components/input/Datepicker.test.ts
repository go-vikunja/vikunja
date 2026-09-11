import {describe, it, expect, beforeEach} from 'vitest'
import {mount} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import Datepicker from './Datepicker.vue'
import en from '@/i18n/lang/en.json'
import {useAuthStore} from '@/stores/auth'
import {DATE_DISPLAY} from '@/constants/dateDisplay'

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})
const pinia = createPinia()

function mountPicker(props: Record<string, any> = {}) {
	return mount(Datepicker, {
		props: {
			modelValue: null,
			...props,
		},
		global: {
			plugins: [i18n, pinia],
			stubs: ['DatepickerInline', 'XButton', 'CustomTransition', 'RouterLink'],
			directives: {
				cy: () => {},
			},
		},
	})
}

describe('Datepicker', () => {
	beforeEach(() => {
		setActivePinia(pinia)
		const authStore = useAuthStore()
		authStore.setUserSettings({
			frontendSettings: {
				dateDisplay: DATE_DISPLAY.RELATIVE,
			},
		} as any)
	})

	it('renders choose date label when modelValue is null', () => {
		const wrapper = mountPicker({modelValue: null})
		expect(wrapper.text()).toContain('Choose a date')
	})

	it('renders normal formatted date when showExactDate is false', () => {
		const testDate = new Date(Date.UTC(2026, 11, 1, 12))
		const wrapper = mountPicker({
			modelValue: testDate,
			showExactDate: false,
		})
		expect(wrapper.find('.datepicker__exact-date').exists()).toBe(false)
		expect(wrapper.text()).toBeTruthy()
	})

	it('renders capitalized relative date and exact date with lighter text when showExactDate is true and dateDisplay is relative', () => {
		const testDate = new Date(Date.UTC(2026, 11, 1, 12))
		const wrapper = mountPicker({
			modelValue: testDate,
			showExactDate: true,
		})

		const relativeSpan = wrapper.find('.datepicker__relative-date')
		const exactSpan = wrapper.find('.datepicker__exact-date')

		expect(relativeSpan.exists()).toBe(true)
		expect(exactSpan.exists()).toBe(true)
		expect(exactSpan.text()).toBe('(2026-12-01)')
		expect(wrapper.text()).toContain('(2026-12-01)')
		// Ensure first letter is capitalized
		const firstChar = relativeSpan.text().charAt(0)
		expect(firstChar).toBe(firstChar.toUpperCase())
	})

	it('does not duplicate exact date when dateDisplay setting is absolute', async () => {
		const authStore = useAuthStore()
		authStore.setUserSettings({
			frontendSettings: {
				dateDisplay: DATE_DISPLAY.YYYY_MM_DD,
			},
		} as any)

		const testDate = new Date(Date.UTC(2026, 11, 1, 12))
		const wrapper = mountPicker({
			modelValue: testDate,
			showExactDate: true,
		})

		expect(wrapper.find('.datepicker__exact-date').exists()).toBe(false)
		expect(wrapper.text()).toContain('2026-12-01')
	})
})
