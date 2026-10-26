import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {mount, type VueWrapper} from '@vue/test-utils'
import type {Directive} from 'vue'

const auth = vi.hoisted(() => ({settings: {timezone: 'UTC', week_start: 1}}))
vi.mock('@/stores/auth', async importOriginal => ({
	...(await importOriginal<typeof import('@/stores/auth')>()),
	useAuthStore: () => auth,
}))

// Shows the key, its params and the plural count, so each assertion names exactly what reaches vue-i18n.
const t = vi.hoisted(() => (key: string, params?: unknown, plural?: number) =>
	`${key}|${JSON.stringify(params ?? {})}|${plural ?? ''}`)
vi.mock('vue-i18n', async importOriginal => ({
	...(await importOriginal<typeof import('vue-i18n')>()),
	useI18n: () => ({t, locale: {value: 'en'}}),
}))

import RepeatCustomForm from './RepeatCustomForm.vue'
import XButton from '@/components/input/Button.vue'
import CalendarMonth from '@/components/input/datepicker/CalendarMonth.vue'
import type {TaskRepeat} from '@/client/generated'
import {defaultCustomForm, formToRepeat, normalizeRepeatForm, type RepeatForm} from '@/helpers/rrule'

const D = 'task.repeat.describe'
const E = 'task.repeat.error'

// Fri Oct 20 2090, noon UTC: a Friday in UTC and New York, safely in the future for the preview.
const FRIDAY = new Date('2090-10-20T12:00:00Z')
// Fri Oct 20 2090, 20:00 in New York, which is already Saturday in UTC.
const FRIDAY_EVENING_NEW_YORK = new Date('2090-10-21T00:00:00Z')
const THURSDAY = new Date('2090-10-19T12:00:00Z')
const NEW_YORK = 'America/New_York'

const noopTooltip: Directive = {}

function form(changes: Partial<RepeatForm> = {}): RepeatForm {
	return {...defaultCustomForm(FRIDAY, 'UTC'), ...changes}
}

interface MountOptions {
	modelValue?: RepeatForm
	dueDate?: Date | null
	anchor?: Date
	repeatsFromCurrentDate?: boolean
	zone?: string
	previewZone?: string | null
	weekStart?: number
	storedRepeat?: TaskRepeat | null
	disabled?: boolean
	vModel?: boolean
}

const mounted: VueWrapper[] = []

function mountForm(options: MountOptions = {}) {
	const {vModel = false, ...props} = options
	const wrapper: VueWrapper = mount(RepeatCustomForm, {
		props: {
			modelValue: form(),
			dueDate: FRIDAY,
			anchor: props.dueDate ?? FRIDAY,
			repeatsFromCurrentDate: false,
			zone: 'UTC',
			previewZone: 'UTC',
			weekStart: 1,
			...props,
			...(vModel ? {'onUpdate:modelValue': (value: RepeatForm) => wrapper.setProps({modelValue: value})} : {}),
		},
		global: {
			stubs: {RouterLink: true, Icon: true},
			directives: {tooltip: noopTooltip},
			mocks: {$t: t},
		},
		attachTo: document.body,
	})
	mounted.push(wrapper)
	return wrapper
}

function lastEmitted(wrapper: VueWrapper): RepeatForm {
	const events = wrapper.emitted('update:modelValue') ?? []
	expect(events.length).toBeGreaterThan(0)
	return events[events.length - 1][0] as RepeatForm
}

function weekdayButtons(wrapper: VueWrapper) {
	return wrapper.findAll('.repeat-custom-weekdays button')
}

function pressedWeekdays(wrapper: VueWrapper): string[] {
	return weekdayButtons(wrapper)
		.filter(button => button.attributes('aria-pressed') === 'true')
		.map(button => button.attributes('aria-label') ?? '')
}

function weekdayButton(wrapper: VueWrapper, longName: string) {
	const button = weekdayButtons(wrapper).find(b => b.attributes('aria-label') === longName)
	expect(button, longName).toBeDefined()
	return button!
}

function select(wrapper: VueWrapper, key: string) {
	const element = wrapper.find(`select[aria-label="${key}|{}|"]`)
	expect(element.exists(), key).toBe(true)
	return element
}

function radio(wrapper: VueWrapper, value: string) {
	const element = wrapper.find(`input[type="radio"][value="${value}"]`)
	expect(element.exists(), value).toBe(true)
	return element
}

function buttonWithKey(wrapper: VueWrapper, key: string) {
	return wrapper.findAll('button').find(b => b.text() === `${key}|{}|`)
}

function doneButton(wrapper: VueWrapper) {
	const button = buttonWithKey(wrapper, 'task.repeat.custom.done')
	expect(button).toBeDefined()
	return button!
}

function clearButton(wrapper: VueWrapper) {
	return buttonWithKey(wrapper, 'task.repeat.custom.clearEndDate')
}

function countInput(wrapper: VueWrapper) {
	return wrapper.find('input[aria-label="task.repeat.custom.count|{}|"]')
}

function errorText(wrapper: VueWrapper): string[] {
	return wrapper.findAll('.repeat-custom-error').map(error => error.text())
}

function summary(wrapper: VueWrapper): string {
	return wrapper.find('.repeat-custom-summary').text()
}

function previewDatesShown(wrapper: VueWrapper): string[] {
	return wrapper.findAll('.repeat-custom-preview li').map(li => li.text())
}

function calendar(wrapper: VueWrapper) {
	return wrapper.findComponent(CalendarMonth)
}

function dayButtons(wrapper: VueWrapper) {
	return wrapper.findAll('.repeat-custom-until-picker .calendar-month__day')
}

function day(wrapper: VueWrapper, ymd: string) {
	const element = wrapper.find(`.repeat-custom-until-picker .calendar-month__day[data-date="${ymd}"]`)
	expect(element.exists(), ymd).toBe(true)
	return element
}

function tabbableDays(wrapper: VueWrapper) {
	return dayButtons(wrapper).filter(button => button.attributes('tabindex') === '0')
}

function selectedDays(wrapper: VueWrapper): string[] {
	return wrapper.findAll('.repeat-custom-until-picker [aria-selected="true"] .calendar-month__day')
		.map(button => button.attributes('data-date') ?? '')
}

function accessibleName(root: Element, element: Element): string {
	const ariaLabel = element.getAttribute('aria-label')
	if (ariaLabel) {
		return ariaLabel
	}
	if (element.id) {
		const label = root.querySelector(`label[for="${element.id}"]`)
		if (label) {
			return label.textContent?.trim() ?? ''
		}
	}
	const wrappingLabel = element.closest('label')
	if (wrappingLabel) {
		return wrappingLabel.textContent?.trim() ?? ''
	}
	return element.tagName === 'BUTTON' ? element.textContent?.trim() ?? '' : ''
}

const originalTZ = process.env.TZ
afterEach(() => {
	process.env.TZ = originalTZ
	mounted.splice(0).forEach(wrapper => wrapper.unmount())
	vi.useRealTimers()
})

describe('RepeatCustomForm', () => {
	describe('interval and unit', () => {
		it('emits a new draft object and leaves the prop untouched', async () => {
			const draft = form()
			const snapshot = structuredClone(draft)
			const wrapper = mountForm({modelValue: draft})

			await wrapper.find('input[type="number"]').setValue('3')

			const emitted = lastEmitted(wrapper)
			expect(emitted).not.toBe(draft)
			expect(emitted).toEqual({...snapshot, interval: 3})
			expect(draft).toEqual(snapshot)
		})

		it('labels the units in the plural that matches the interval', async () => {
			const wrapper = mountForm({modelValue: form({interval: 1}), vModel: true})
			const unitLabels = () => select(wrapper, 'task.repeat.custom.unit').findAll('option').map(option => option.text())
			expect(unitLabels()).toEqual(['hour', 'day', 'week', 'month', 'year'].map(unit => `task.repeat.unit.${unit}|1|`))

			await wrapper.find('input[type="number"]').setValue('2')

			expect(unitLabels()).toEqual(['hour', 'day', 'week', 'month', 'year'].map(unit => `task.repeat.unit.${unit}|2|`))
		})

		it('keeps every other field when the unit changes, so month starts on the due day', async () => {
			const wrapper = mountForm({modelValue: form({interval: 2}), vModel: true})

			await select(wrapper, 'task.repeat.custom.unit').setValue('month')

			expect(lastEmitted(wrapper)).toEqual(form({interval: 2, unit: 'month'}))
			expect(lastEmitted(wrapper)).toMatchObject({monthMode: 'day', monthDay: null})
			expect((radio(wrapper, 'day').element as HTMLInputElement).checked).toBe(true)
			expect((select(wrapper, 'task.repeat.custom.monthDay').element as HTMLSelectElement).value).toBe('20')
		})

		it('hides the weekday and month controls for hourly repeats', () => {
			const wrapper = mountForm({modelValue: form({unit: 'hour'})})

			expect(weekdayButtons(wrapper)).toHaveLength(0)
			expect(wrapper.find('input[value="nthWeekday"]').exists()).toBe(false)
			expect(wrapper.find('select[aria-label="task.repeat.custom.monthDay|{}|"]').exists()).toBe(false)
			expect(summary(wrapper)).toBe(`${D}.hourly|{"n":1}|1`)
		})
	})

	describe('weekday grid', () => {
		it.each([
			[0, ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']],
			[1, ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday']],
		])('orders the days from the week start %i', (weekStart, expected) => {
			const wrapper = mountForm({weekStart})

			expect(weekdayButtons(wrapper).map(b => b.attributes('aria-label'))).toEqual(expected)
			expect(weekdayButtons(wrapper)[0].text()).toBe(expected[0].slice(0, 3))
		})

		it('shows the due weekday pressed while the selection is implicit', () => {
			const wrapper = mountForm({modelValue: form({weekdays: null})})

			expect(pressedWeekdays(wrapper)).toEqual(['Friday'])
			expect(weekdayButtons(wrapper).every(b => b.attributes('type') === 'button')).toBe(true)
		})

		it('makes the selection explicit on the first toggle, starting from the due weekday', async () => {
			const wrapper = mountForm({modelValue: form({weekdays: null})})

			await weekdayButton(wrapper, 'Monday').trigger('click')

			expect(lastEmitted(wrapper).weekdays).toEqual(['mo', 'fr'])
		})

		it('keeps an emptied selection empty in the draft and shows the fallback hint', async () => {
			const wrapper = mountForm({modelValue: form({weekdays: ['mo']}), vModel: true})
			const group = wrapper.find('fieldset')
			expect(wrapper.find('.repeat-custom-weekday-hint').exists()).toBe(false)

			await weekdayButton(wrapper, 'Monday').trigger('click')

			expect(lastEmitted(wrapper).weekdays).toEqual([])
			expect(pressedWeekdays(wrapper)).toEqual([])
			const hint = wrapper.find('.repeat-custom-weekday-hint')
			expect(hint.text()).toBe('task.repeat.custom.weekdayFallback|{"weekday":"Friday"}|')
			expect(group.attributes('aria-describedby')).toBe(hint.attributes('id'))
			// The summary reads the draft the way it will be saved: the due weekday.
			expect(summary(wrapper)).toBe(`${D}.weeklyOn|{"n":1,"days":"Friday"}|1`)

			await weekdayButton(wrapper, 'Monday').trigger('click')

			expect(lastEmitted(wrapper).weekdays).toEqual(['mo'])
			expect(pressedWeekdays(wrapper)).toEqual(['Monday'])
		})

		it('empties the selection when the implicit due weekday is toggled off', async () => {
			const wrapper = mountForm({modelValue: form({weekdays: null})})

			await weekdayButton(wrapper, 'Friday').trigger('click')

			expect(lastEmitted(wrapper).weekdays).toEqual([])
		})
	})

	describe('month mode', () => {
		const monthly = () => form({unit: 'month'})

		it('picks a day of the month explicitly', async () => {
			const wrapper = mountForm({modelValue: monthly()})

			await select(wrapper, 'task.repeat.custom.monthDay').setValue('15')

			expect(lastEmitted(wrapper)).toMatchObject({monthMode: 'day', monthDay: 15})
		})

		it('selects the nth weekday, including the last one', async () => {
			const wrapper = mountForm({modelValue: monthly(), vModel: true})
			const ordinals = select(wrapper, 'task.repeat.custom.ordinal').findAll('option').map(option => option.text())
			expect(ordinals).toEqual([
				`${D}.ordinal.one|{"n":1}|`,
				`${D}.ordinal.two|{"n":2}|`,
				`${D}.ordinal.few|{"n":3}|`,
				`${D}.ordinal.other|{"n":4}|`,
				'task.repeat.custom.lastOrdinal|{}|',
			])

			await select(wrapper, 'task.repeat.custom.ordinal').setValue('-1')

			const draft = lastEmitted(wrapper)
			expect(draft).toMatchObject({monthMode: 'nthWeekday', ordinal: -1, monthWeekday: 'fr'})
			expect(formToRepeat(normalizeRepeatForm(draft), FRIDAY, 'UTC')).toEqual({freq: 'monthly', by_day: ['-1fr']})
			expect(summary(wrapper)).toBe(`${D}.monthlyOnLastWeekday|{"n":1,"weekday":"Friday"}|1`)

			await select(wrapper, 'task.repeat.custom.monthWeekday').setValue('tu')
			await select(wrapper, 'task.repeat.custom.ordinal').setValue('2')

			expect(lastEmitted(wrapper)).toMatchObject({monthMode: 'nthWeekday', ordinal: 2, monthWeekday: 'tu'})
			expect(summary(wrapper)).toBe(`${D}.monthlyOnNth|{"n":1,"ordinal":"${D}.ordinal.two|{\\"n\\":2}|","weekday":"Tuesday"}|1`)
		})

		it('switches modes with the radios', async () => {
			const wrapper = mountForm({modelValue: monthly()})

			await radio(wrapper, 'lastDay').setValue(true)
			expect(lastEmitted(wrapper).monthMode).toBe('lastDay')

			await radio(wrapper, 'nthWeekday').setValue(true)
			expect(lastEmitted(wrapper).monthMode).toBe('nthWeekday')
		})

		it('groups the month radios under a legend', () => {
			const wrapper = mountForm({modelValue: monthly()})

			const fieldset = radio(wrapper, 'lastDay').element.closest('fieldset')
			expect(fieldset?.querySelector('legend')?.textContent?.trim()).toBe('task.repeat.custom.monthMode|{}|')
		})
	})

	describe('end condition', () => {
		it('words the count as times including the current one', async () => {
			const wrapper = mountForm({modelValue: form({end: 'count', count: 3}), vModel: true})
			const count = countInput(wrapper)
			const hint = wrapper.findAll('.repeat-custom-hint').find(p => p.text() === 'task.repeat.custom.countIncludesCurrent|{}|')

			expect(hint).toBeDefined()
			expect(count.attributes('aria-describedby')).toBe(hint!.attributes('id'))
			expect(wrapper.find('.repeat-custom-times').text()).toBe('task.repeat.custom.times|3|')
			expect(summary(wrapper)).toBe(`${D}.weeklyOn|{"n":1,"days":"Friday"}|1${D}.endCount|{"n":3}|3`)

			await count.setValue('1')

			expect(lastEmitted(wrapper)).toMatchObject({end: 'count', count: 1})
			expect(wrapper.find('.repeat-custom-times').text()).toBe('task.repeat.custom.times|1|')
		})

		it('shows the date-only calendar only for an end date', async () => {
			const wrapper = mountForm({modelValue: form(), vModel: true})
			expect(calendar(wrapper).exists()).toBe(false)

			await radio(wrapper, 'count').setValue(true)
			expect(calendar(wrapper).exists()).toBe(false)

			await radio(wrapper, 'until').setValue(true)
			expect(lastEmitted(wrapper).end).toBe('until')
			expect(calendar(wrapper).exists()).toBe(true)
			expect(calendar(wrapper).props('large')).toBe(true)
			expect(calendar(wrapper).props('minDate')).toBeNull()
			expect(wrapper.find('.repeat-custom-until-picker input[type="time"], .repeat-custom-until-picker .time-control').exists()).toBe(false)
		})

		it('hands the calendar the stored calendar date, whatever the browser zone', () => {
			process.env.TZ = 'America/Los_Angeles'
			const wrapper = mountForm({modelValue: form({end: 'until', untilDate: '2090-11-06'}), dueDate: null, anchor: FRIDAY})

			const shown = calendar(wrapper).props('selected') as Date
			expect([shown.getFullYear(), shown.getMonth() + 1, shown.getDate()]).toEqual([2090, 11, 6])
			expect(selectedDays(wrapper)).toEqual(['2090-11-06'])
		})

		it('stores the calendar day clicked, without a zone shift', async () => {
			process.env.TZ = 'Pacific/Kiritimati'
			const wrapper = mountForm({modelValue: form({end: 'until', untilDate: '2090-11-03'}), zone: NEW_YORK})

			await day(wrapper, '2090-11-10').trigger('click')

			expect(lastEmitted(wrapper)).toMatchObject({end: 'until', untilDate: '2090-11-10'})
		})

		it('shows a stored end time that is not the end of a day', () => {
			const wrapper = mountForm({
				modelValue: form({end: 'until', untilDate: '2090-11-03', untilInstant: '2090-11-03T12:00:00Z'}),
			})

			expect(wrapper.find('.repeat-custom-until-text').text()).toBe('Nov 3, 2090, 12:00 PM')
		})

		it('reads a stored end time in the recurrence zone', () => {
			process.env.TZ = 'UTC'
			// 22:00 on Nov 3 in New York is already Nov 4 in UTC.
			const wrapper = mountForm({
				modelValue: form({end: 'until', untilDate: '2090-11-03', untilInstant: '2090-11-04T02:00:00Z'}),
				zone: NEW_YORK,
			})

			expect(wrapper.find('.repeat-custom-until-text').text()).toBe('Nov 3, 2090, 10:00 PM')
		})

		it.each([null, '2090-11-03T23:59:59Z'])('shows an end date picked here as a plain date (stored %s)', untilInstant => {
			const wrapper = mountForm({modelValue: form({end: 'until', untilDate: '2090-11-03', untilInstant})})

			expect(wrapper.find('.repeat-custom-until-text').text()).toBe('Nov 3, 2090')
		})

		it('requires a day when "On a date" is chosen, and selects none until one is picked', async () => {
			// Without an end date the calendar opens on today's month.
			vi.useFakeTimers({toFake: ['Date']})
			vi.setSystemTime(new Date(2090, 9, 15, 9))
			const wrapper = mountForm({modelValue: form(), vModel: true})

			await radio(wrapper, 'until').setValue(true)

			expect(errorText(wrapper)).toEqual([`${E}.untilMissing|{}|`])
			expect(doneButton(wrapper).attributes('disabled')).toBeDefined()
			expect(selectedDays(wrapper)).toEqual([])
			expect(clearButton(wrapper)).toBeUndefined()

			await day(wrapper, '2090-10-27').trigger('click')

			expect(lastEmitted(wrapper)).toMatchObject({end: 'until', untilDate: '2090-10-27'})
			expect(selectedDays(wrapper)).toEqual(['2090-10-27'])
			expect(errorText(wrapper)).toEqual([])
			expect(doneButton(wrapper).attributes('disabled')).toBeUndefined()
		})

		it('clears the end date, requires a new one and moves focus to the end choice', async () => {
			const wrapper = mountForm({modelValue: form({end: 'until', untilDate: '2090-11-03'}), vModel: true})
			expect(selectedDays(wrapper)).toEqual(['2090-11-03'])

			await clearButton(wrapper)!.trigger('click')
			await wrapper.vm.$nextTick()

			expect(lastEmitted(wrapper)).toMatchObject({end: 'until', untilDate: null})
			const error = wrapper.find('.repeat-custom-error')
			expect(error.text()).toBe(`${E}.untilMissing|{}|`)
			expect(radio(wrapper, 'until').attributes('aria-describedby')).toBe(error.attributes('id'))
			expect(wrapper.find('.repeat-custom-until-picker').attributes('aria-describedby')).toBe(error.attributes('id'))
			expect(doneButton(wrapper).attributes('disabled')).toBeDefined()
			expect(wrapper.find('.repeat-custom-until-text').exists()).toBe(false)
			expect(clearButton(wrapper)).toBeUndefined()
			expect(selectedDays(wrapper)).toEqual([])
			expect(document.activeElement).toBe(radio(wrapper, 'until').element)

			await doneButton(wrapper).trigger('click')
			expect(wrapper.emitted('done')).toBeUndefined()
		})
	})

	describe('end date calendar from the keyboard', () => {
		beforeEach(() => {
			vi.useFakeTimers({toFake: ['Date']})
			vi.setSystemTime(new Date(2090, 9, 15, 9))
		})

		it('enters on an enabled day when there is no end date', async () => {
			const wrapper = mountForm({modelValue: form({end: 'until'})})

			const tabbable = tabbableDays(wrapper)
			expect(tabbable.map(button => button.attributes('data-date'))).toEqual(['2090-10-15'])
			expect(tabbable[0].attributes('disabled')).toBeUndefined()
			expect(selectedDays(wrapper)).toEqual([])
		})

		it('enters on the end date when one is set', () => {
			const wrapper = mountForm({modelValue: form({end: 'until', untilDate: '2090-12-24'})})

			const tabbable = tabbableDays(wrapper)
			expect(tabbable.map(button => button.attributes('data-date'))).toEqual(['2090-12-24'])
			expect(tabbable[0].attributes('disabled')).toBeUndefined()
		})

		it('leaves every day enabled, so a day before the due date is refused by validation', async () => {
			const wrapper = mountForm({modelValue: form({end: 'until'}), vModel: true})
			expect(dayButtons(wrapper).filter(button => button.attributes('disabled') !== undefined)).toEqual([])

			await day(wrapper, '2090-10-16').trigger('click')

			expect(lastEmitted(wrapper).untilDate).toBe('2090-10-16')
			const error = wrapper.find('.repeat-custom-error')
			expect(error.text()).toBe(`${E}.untilBeforeDue|{}|`)
			expect(radio(wrapper, 'until').attributes('aria-describedby')).toBe(error.attributes('id'))
			expect(doneButton(wrapper).attributes('disabled')).toBeDefined()
		})

		it('moves with the arrows, picks, clears and enters again', async () => {
			const wrapper = mountForm({modelValue: form({end: 'until'}), vModel: true})
			const grid = wrapper.find('.repeat-custom-until-picker [role="grid"]')

			;(tabbableDays(wrapper)[0].element as HTMLButtonElement).focus()
			await grid.trigger('keydown', {key: 'ArrowDown'})
			await wrapper.vm.$nextTick()
			await grid.trigger('keydown', {key: 'ArrowRight'})
			await wrapper.vm.$nextTick()

			expect((document.activeElement as HTMLElement).dataset.date).toBe('2090-10-23')
			expect(tabbableDays(wrapper).map(button => button.attributes('data-date'))).toEqual(['2090-10-23'])

			await (wrapper.find('.calendar-month__day[data-date="2090-10-23"]')).trigger('click')
			expect(lastEmitted(wrapper).untilDate).toBe('2090-10-23')
			expect(selectedDays(wrapper)).toEqual(['2090-10-23'])

			await clearButton(wrapper)!.trigger('click')
			await wrapper.vm.$nextTick()

			expect(selectedDays(wrapper)).toEqual([])
			const again = tabbableDays(wrapper)
			expect(again).toHaveLength(1)
			expect(again[0].attributes('disabled')).toBeUndefined()

			;(again[0].element as HTMLButtonElement).focus()
			await grid.trigger('keydown', {key: 'ArrowRight'})
			await wrapper.vm.$nextTick()
			await (wrapper.find(`.calendar-month__day[data-date="${(document.activeElement as HTMLElement).dataset.date}"]`)).trigger('click')

			expect(lastEmitted(wrapper).untilDate).toBe('2090-10-24')
			expect(errorText(wrapper)).toEqual([])
		})

		it('stays on the month being browsed while another field changes', async () => {
			const wrapper = mountForm({modelValue: form({end: 'until', untilDate: '2090-11-03'}), vModel: true})
			const next = wrapper.find('.calendar-month__nav-button[aria-label="input.datepicker.nextMonth|{}|"]')

			await next.trigger('click')
			await next.trigger('click')
			expect(wrapper.find('.calendar-month__month-select').element).toHaveProperty('value', '0')

			await wrapper.find('input[type="number"]').setValue('2')

			expect(wrapper.find('.calendar-month__month-select').element).toHaveProperty('value', '0')
			expect(wrapper.find('.calendar-month__year-input').element).toHaveProperty('value', '2091')
		})
	})

	describe('validation', () => {
		it('shows the interval error, links it to the field and blocks Done', async () => {
			const wrapper = mountForm({modelValue: form(), vModel: true})
			const interval = wrapper.find('input[type="number"]')

			await interval.setValue('')

			expect(Number.isNaN(lastEmitted(wrapper).interval)).toBe(true)
			const error = wrapper.find('.repeat-custom-error')
			expect(error.text()).toBe(`${E}.interval|{}|`)
			expect(interval.attributes('aria-invalid')).toBe('true')
			expect(interval.attributes('aria-describedby')).toBe(error.attributes('id'))
			expect(doneButton(wrapper).attributes('disabled')).toBeDefined()

			await doneButton(wrapper).trigger('click')
			expect(wrapper.emitted('done')).toBeUndefined()
		})

		it('refuses an end date before the due date', () => {
			const wrapper = mountForm({modelValue: form({end: 'until', untilDate: '2090-10-19'})})

			const error = wrapper.find('.repeat-custom-error')
			expect(error.text()).toBe(`${E}.untilBeforeDue|{}|`)
			expect(radio(wrapper, 'until').attributes('aria-describedby')).toBe(error.attributes('id'))
			expect(doneButton(wrapper).attributes('disabled')).toBeDefined()
		})

		it('refuses a count below one', () => {
			const wrapper = mountForm({modelValue: form({end: 'count', count: 0})})

			const count = countInput(wrapper)
			expect(errorText(wrapper)).toEqual([`${E}.count|{}|`])
			expect(count.attributes('aria-invalid')).toBe('true')
			expect(count.attributes('aria-describedby')?.split(' ')).toContain(wrapper.find('.repeat-custom-error').attributes('id'))
		})

		it.each([
			['hour', 876_601, 876_600],
			['day', 36_526, 36_525],
			['week', 5_219, 5_218],
			['month', 1_201, 1_200],
			['year', 101, 100],
		] as const)('refuses an interval above the %s limit with its maximum', async (unit, interval, max) => {
			const wrapper = mountForm({modelValue: form({unit, interval: max}), vModel: true})
			expect(errorText(wrapper)).toEqual([])

			await wrapper.find('input[type="number"]').setValue(String(interval))

			expect(errorText(wrapper)).toEqual([`${E}.intervalMax|{"max":${max}}|`])
			expect(wrapper.find('input[type="number"]').attributes('aria-invalid')).toBe('true')
			expect(doneButton(wrapper).attributes('disabled')).toBeDefined()
		})

		it('refuses a count above the limit with its maximum', async () => {
			const wrapper = mountForm({modelValue: form({end: 'count', count: 10_000}), vModel: true})
			expect(errorText(wrapper)).toEqual([])

			await countInput(wrapper).setValue('10001')

			expect(errorText(wrapper)).toEqual([`${E}.countMax|{"max":10000}|`])
			expect(countInput(wrapper).attributes('aria-invalid')).toBe('true')
			expect(doneButton(wrapper).attributes('disabled')).toBeDefined()
		})

		it('refuses raising only the count of the stored repeat, and allows lowering it', async () => {
			const storedRepeat: TaskRepeat = {freq: 'weekly', interval: 1, count: 3}
			const wrapper = mountForm({modelValue: form({end: 'count', count: 3}), storedRepeat, vModel: true})
			expect(errorText(wrapper)).toEqual([])

			await countInput(wrapper).setValue('5')

			expect(errorText(wrapper)).toEqual([`${E}.countRaise|{}|`])
			expect(countInput(wrapper).attributes('aria-describedby')?.split(' ')).toContain(wrapper.find('.repeat-custom-error').attributes('id'))
			expect(doneButton(wrapper).attributes('disabled')).toBeDefined()

			await countInput(wrapper).setValue('2')

			expect(errorText(wrapper)).toEqual([])
			expect(doneButton(wrapper).attributes('disabled')).toBeUndefined()
		})

		it('allows a higher count together with another change', async () => {
			const storedRepeat: TaskRepeat = {freq: 'weekly', interval: 1, count: 3}
			const wrapper = mountForm({modelValue: form({end: 'count', count: 5, interval: 2}), storedRepeat})

			expect(errorText(wrapper)).toEqual([])
			expect(doneButton(wrapper).attributes('disabled')).toBeUndefined()
		})

		it('emits done for a valid draft, and back without validating', async () => {
			const valid = mountForm()
			await doneButton(valid).trigger('click')
			expect(valid.emitted('done')).toHaveLength(1)

			const invalid = mountForm({modelValue: form({interval: 0})})
			await buttonWithKey(invalid, 'task.repeat.custom.back')!.trigger('click')
			expect(invalid.emitted('back')).toHaveLength(1)
			expect(invalid.emitted('done')).toBeUndefined()
		})
	})

	describe('summary and preview', () => {
		it('recompute when the draft changes', async () => {
			const wrapper = mountForm({modelValue: form()})
			expect(summary(wrapper)).toBe(`${D}.weeklyOn|{"n":1,"days":"Friday"}|1`)
			expect(previewDatesShown(wrapper)).toEqual(['Fri, Oct 20', 'Fri, Oct 27', 'Fri, Nov 3'])

			await wrapper.setProps({modelValue: form({interval: 2, end: 'count', count: 2})})

			expect(summary(wrapper)).toBe(`${D}.weeklyOn|{"n":2,"days":"Friday"}|2${D}.endCount|{"n":2}|2`)
			expect(previewDatesShown(wrapper)).toEqual(['Fri, Oct 20', 'Fri, Nov 3'])
		})

		it('updates the summary and preview well within the feedback budget', async () => {
			const wrapper = mountForm({modelValue: form({weekdays: ['mo', 'we', 'fr']})})
			const timings: number[] = []
			for (let i = 0; i < 20; i++) {
				const start = performance.now()
				await wrapper.setProps({modelValue: form({weekdays: ['mo', 'we', 'fr'], interval: 1 + (i % 3)})})
				timings.push(performance.now() - start)
				expect(previewDatesShown(wrapper)).toHaveLength(3)
			}
			timings.sort((a, b) => a - b)
			expect(timings[10]).toBeLessThan(50)
		})

		it('derives the implicit weekday in the recurrence zone, not the browser zone', () => {
			process.env.TZ = 'UTC'
			const wrapper = mountForm({
				modelValue: defaultCustomForm(FRIDAY_EVENING_NEW_YORK, NEW_YORK),
				dueDate: FRIDAY_EVENING_NEW_YORK,
				zone: NEW_YORK,
				previewZone: NEW_YORK,
			})

			expect(pressedWeekdays(wrapper)).toEqual(['Friday'])
			expect(summary(wrapper)).toBe(`${D}.weeklyOn|{"n":1,"days":"Friday"}|1`)
			expect(previewDatesShown(wrapper)[0]).toBe('Fri, Oct 20')
		})

		it('formats the preview dates in the recurrence zone', () => {
			// Already Saturday in Kiritimati (UTC+14) while it is still Friday in every browser zone west of UTC+13.
			const due = new Date('2090-10-20T11:00:00Z')
			const zone = 'Pacific/Kiritimati'
			const wrapper = mountForm({modelValue: defaultCustomForm(due, zone), dueDate: due, zone, previewZone: zone})

			expect(pressedWeekdays(wrapper)).toEqual(['Saturday'])
			expect(previewDatesShown(wrapper)).toEqual(['Sat, Oct 21', 'Sat, Oct 28', 'Sat, Nov 4'])
		})

		it('hides the preview and hints at the time zone setting without one', async () => {
			const wrapper = mountForm({modelValue: form(), zone: NEW_YORK, previewZone: null})
			expect(previewDatesShown(wrapper)).toEqual([])
			expect(wrapper.find('.repeat-custom-preview-hint').text()).toBe('task.repeat.custom.noPreview|{}|')

			await wrapper.setProps({previewZone: NEW_YORK})

			expect(previewDatesShown(wrapper)).toHaveLength(3)
			expect(wrapper.find('.repeat-custom-preview-hint').exists()).toBe(false)
		})

		it('shows no preview and no hint for other hidden reasons', () => {
			const wrapper = mountForm({repeatsFromCurrentDate: true})

			expect(wrapper.find('.repeat-custom-preview').exists()).toBe(false)
			expect(wrapper.find('.repeat-custom-preview-hint').exists()).toBe(false)
		})

		it('derives the weekday from the anchor when there is no due date', () => {
			const wrapper = mountForm({
				modelValue: defaultCustomForm(THURSDAY, 'UTC'),
				dueDate: null,
				anchor: THURSDAY,
			})

			expect(pressedWeekdays(wrapper)).toEqual(['Thursday'])
			expect(summary(wrapper)).toBe(`${D}.weeklyOn|{"n":1,"days":"Thursday"}|1`)
			expect(previewDatesShown(wrapper)).toEqual([])
		})
	})

	it('builds Mon, Wed, Fri from the default form with two toggles and Done', async () => {
		const wrapper = mountForm({modelValue: defaultCustomForm(FRIDAY, 'UTC'), vModel: true})

		await weekdayButton(wrapper, 'Monday').trigger('click')
		await weekdayButton(wrapper, 'Wednesday').trigger('click')
		await doneButton(wrapper).trigger('click')

		expect(wrapper.emitted('done')).toHaveLength(1)
		expect(formToRepeat(normalizeRepeatForm(lastEmitted(wrapper)), FRIDAY, 'UTC')).toEqual({freq: 'weekly', by_day: ['mo', 'we', 'fr']})
		expect(summary(wrapper)).toBe(`${D}.weeklyOn|{"n":1,"days":"Monday, Wednesday, and Friday"}|1`)
	})

	it.each([
		['weekly', form()],
		['monthly', form({unit: 'month'})],
		['with an end date', form({end: 'until', untilDate: '2090-11-03'})],
	])('gives every control an accessible name (%s)', (_, draft) => {
		const wrapper = mountForm({modelValue: draft})
		const root = wrapper.element as Element
		const controls = Array.from(root.querySelectorAll('input, select, button'))
		expect(controls.length).toBeGreaterThan(5)

		const unnamed = controls.filter(control => accessibleName(root, control) === '').map(control => control.outerHTML)
		expect(unnamed).toEqual([])
		for (const group of Array.from(root.querySelectorAll('fieldset'))) {
			expect(group.querySelector('legend')?.textContent?.trim()).toBeTruthy()
		}
		if (draft.end === 'until') {
			expect(wrapper.find('.repeat-custom-until-picker').attributes()).toMatchObject({
				role: 'group',
				'aria-label': 'task.repeat.custom.untilDate|{}|',
			})
		}
	})

	it.each([
		['monthly', form({unit: 'month'})],
		['with an end date', form({end: 'until', untilDate: '2090-11-03'})],
	])('disables every control when disabled (%s)', async (_, draft) => {
		const wrapper = mountForm({modelValue: draft, disabled: true})

		const controls = wrapper.findAll('input, select, button')
			.filter(control => !control.element.closest('.repeat-custom-until-picker'))
		expect(controls.length).toBeGreaterThan(5)
		const enabled = controls.filter(control => control.attributes('disabled') === undefined && control.text() !== 'task.repeat.custom.back|{}|')
		expect(enabled.map(control => control.html())).toEqual([])

		if (draft.end === 'until') {
			expect(wrapper.find('.repeat-custom-until-picker').attributes('inert')).toBeDefined()
			await day(wrapper, '2090-11-10').trigger('click')
			await clearButton(wrapper)!.trigger('click')
			expect(wrapper.emitted('update:modelValue')).toBeUndefined()
		}
	})

	describe('Done through the button component event', () => {
		function emitDoneClick(wrapper: VueWrapper) {
			const done = wrapper.findAllComponents(XButton).find(button => button.text() === 'task.repeat.custom.done|{}|')
			expect(done).toBeDefined()
			done!.vm.$emit('click', new MouseEvent('click'))
		}

		it('does not emit done for an invalid draft', () => {
			const wrapper = mountForm({modelValue: form({interval: 0})})
			emitDoneClick(wrapper)
			expect(wrapper.emitted('done')).toBeUndefined()
		})

		it('does not emit done while disabled', () => {
			const wrapper = mountForm({disabled: true})
			emitDoneClick(wrapper)
			expect(wrapper.emitted('done')).toBeUndefined()
		})

		it('emits done for a valid draft', () => {
			const wrapper = mountForm()
			emitDoneClick(wrapper)
			expect(wrapper.emitted('done')).toHaveLength(1)
		})
	})
})
