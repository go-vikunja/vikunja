import {beforeEach, describe, expect, it, vi} from 'vitest'
import {mount} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'

import JalaliCalendarGrid from './JalaliCalendarGrid.vue'
import en from '@/i18n/lang/en.json'

function makeI18n() {
	return createI18n({
		legacy: false,
		locale: 'fa-IR',
		fallbackLocale: 'en',
		messages: {en},
	})
}

function mountGrid(props: Record<string, unknown> = {}) {
	return mount(JalaliCalendarGrid, {
		props: {
			modelValue: new Date('2026-09-07T12:00:00Z'),
			timeZone: 'Asia/Tehran',
			...props,
		},
		global: {
			plugins: [makeI18n()],
		},
	})
}

function dayButton(wrapper: ReturnType<typeof mountGrid>, label: string) {
	const buttons = wrapper.findAll('.jalali-calendar__day')
	const found = buttons.find((button) => button.text() === label)
	if (!found) {
		throw new Error(`day button ${label} not found`)
	}
	return found
}

describe('JalaliCalendarGrid', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		document.dir = 'rtl'
	})

	it('renders the Jalali month title and Saturday-first weekdays', () => {
		const wrapper = mountGrid()

		expect(wrapper.find('.jalali-calendar__title').text()).toBe('۱۴۰۵ شهریور')

		const weekdays = wrapper.findAll('.jalali-calendar__weekday').map((w) => w.text())
		expect(weekdays).toEqual(['شنبه', 'یک', 'دو', 'سه', 'چهار', 'پنج', 'جمعه'])
	})

	it('renders 31 days for Shahrivar with Persian digits', () => {
		const wrapper = mountGrid()
		const days = wrapper.findAll('.jalali-calendar__day')

		expect(days).toHaveLength(31)
		expect(days[0].text()).toBe('۱')
		expect(days[30].text()).toBe('۳۱')
	})

	it('selecting the already-selected day keeps the exact instant', async () => {
		const wrapper = mountGrid()

		await dayButton(wrapper, '۱۶').trigger('click')

		const emitted = wrapper.emitted('update:modelValue')?.pop()?.[0] as Date
		expect(emitted.toISOString()).toBe('2026-09-07T12:00:00.000Z')
	})

	it('selecting another day keeps the wall-clock time', async () => {
		const wrapper = mountGrid()

		await dayButton(wrapper, '۱۷').trigger('click')

		const emitted = wrapper.emitted('update:modelValue')?.pop()?.[0] as Date
		expect(emitted.toISOString()).toBe('2026-09-08T12:00:00.000Z')
	})

	it('navigates across the Esfand to Farvardin boundary', async () => {
		const wrapper = mountGrid({modelValue: new Date('2026-03-01T12:00:00Z')})
		expect(wrapper.find('.jalali-calendar__title').text()).toBe('۱۴۰۴ اسفند')

		// Nav order: prev-year, prev-month, next-month, next-year.
		const buttons = wrapper.findAll('.jalali-calendar__nav-button')
		await buttons[2].trigger('click')
		expect(wrapper.find('.jalali-calendar__title').text()).toBe('۱۴۰۵ فروردین')

		await buttons[1].trigger('click')
		expect(wrapper.find('.jalali-calendar__title').text()).toBe('۱۴۰۴ اسفند')
	})

	it('marks today in the user timezone', () => {
		const wrapper = mountGrid({modelValue: null})
		const today = wrapper.findAll('.jalali-calendar__day.is-today')

		expect(today).toHaveLength(1)
	})

	it('disables days outside min and max dates', () => {
		const wrapper = mountGrid({
			modelValue: null,
			minDate: new Date('2026-09-07T12:00:00Z'),
			maxDate: new Date('2026-09-10T12:00:00Z'),
		})

		expect(dayButton(wrapper, '۱۵').attributes('disabled')).toBeDefined()
		expect(dayButton(wrapper, '۱۶').attributes('disabled')).toBeUndefined()
		expect(dayButton(wrapper, '۲۰').attributes('disabled')).toBeDefined()
	})

	it('selects a range across month boundaries', async () => {
		const wrapper = mountGrid({modelValue: null, mode: 'range'})

		await dayButton(wrapper, '۳۰').trigger('click')
		await wrapper.setProps({modelValue: wrapper.emitted('update:modelValue')?.pop()?.[0] as Date[]})

		const buttons = wrapper.findAll('.jalali-calendar__nav-button')
		await buttons[2].trigger('click')
		expect(wrapper.find('.jalali-calendar__title').text()).toBe('۱۴۰۵ مهر')

		await dayButton(wrapper, '۲').trigger('click')
		const emitted = wrapper.emitted('update:modelValue')?.pop()?.[0] as Date[]
		expect(emitted).toHaveLength(2)
		// Start of Shahrivar 30, end of Mehr 2, both in Asia/Tehran wall-clock.
		expect(emitted[0].toISOString()).toBe('2026-09-20T20:30:00.000Z')
		expect(emitted[1].toISOString()).toBe('2026-09-24T20:29:00.000Z')
	})

	it('moves focus with arrow keys in RTL', async () => {
		// happy-dom does not move document.activeElement on focus(), so assert
		// on which element focus() was invoked instead.
		const focusSpy = vi.spyOn(window.HTMLElement.prototype, 'focus')
		const wrapper = mountGrid({modelValue: null})

		await dayButton(wrapper, '۱۶').trigger('keydown', {key: 'ArrowLeft'})
		expect(focusSpy).toHaveBeenCalled()
		expect((focusSpy.mock.instances.pop() as HTMLElement).textContent).toBe('۱۷')

		await dayButton(wrapper, '۱۷').trigger('keydown', {key: 'ArrowRight'})
		expect((focusSpy.mock.instances.pop() as HTMLElement).textContent).toBe('۱۶')

		focusSpy.mockRestore()
	})

	it('labels days with full Jalali dates', () => {
		const wrapper = mountGrid()
		expect(dayButton(wrapper, '۱۶').attributes('aria-label')).toContain('شهریور')
	})

	it('renders 30 days for Esfand in a leap year', () => {
		const wrapper = mountGrid({modelValue: new Date('2025-03-01T12:00:00Z')})
		expect(wrapper.find('.jalali-calendar__title').text()).toBe('۱۴۰۳ اسفند')
		expect(wrapper.findAll('.jalali-calendar__day')).toHaveLength(30)
	})

	it('supports 12-hour time with an AM/PM toggle', async () => {
		const wrapper = mountGrid({
			modelValue: new Date('2026-09-07T12:00:00Z'),
			enableTime: true,
			time24hr: false,
		})

		const inputs = wrapper.findAll('.jalali-calendar__time-input')
		expect(inputs).toHaveLength(2)
		expect((inputs[0].element as HTMLInputElement).value).toBe('3')
		expect(wrapper.find('.jalali-calendar__ampm').text()).toBe('PM')

		await wrapper.find('.jalali-calendar__ampm').trigger('click')
		expect(wrapper.find('.jalali-calendar__ampm').text()).toBe('AM')
	})

	it('selects a same-day range', async () => {
		const wrapper = mountGrid({modelValue: null, mode: 'range'})

		await dayButton(wrapper, '۱۶').trigger('click')
		await dayButton(wrapper, '۱۶').trigger('click')

		const emitted = wrapper.emitted('update:modelValue')?.pop()?.[0] as Date[]
		expect(emitted).toHaveLength(2)
		expect(emitted[0].toISOString()).toBe('2026-09-06T20:30:00.000Z')
		expect(emitted[1].toISOString()).toBe('2026-09-07T20:29:00.000Z')
	})
})
