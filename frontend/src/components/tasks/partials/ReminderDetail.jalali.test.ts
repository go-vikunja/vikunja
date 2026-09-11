import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {mount, type VueWrapper} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'

import ReminderDetail from './ReminderDetail.vue'
import {i18n as globalI18n} from '@/i18n'
import {useAuthStore} from '@/stores/auth'
import en from '@/i18n/lang/en.json'
import TaskReminderModel from '@/models/taskReminder'
import type {ITaskReminder} from '@/modelTypes/ITaskReminder'
import {REMINDER_PERIOD_RELATIVE_TO_TYPES} from '@/types/IReminderPeriodRelativeTo'
import {SECONDS_A_DAY, SECONDS_A_HOUR} from '@/constants/date'
import {
	getJalaliMonthLength,
	instantToJalali,
	jalaliToInstant,
} from '@/helpers/time/jalali'
import {toISOStringOrNull} from '@/helpers/time/toISOStringOrNull'
import {periodToSeconds} from '@/helpers/time/period'

const FROZEN_NOW = new Date('2026-09-07T12:00:00.000Z')
const JALALI_DIGITS = new RegExp('[۰-۹٠-٩]')
const GREGORIAN_ISO = new RegExp('^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}\\.\\d{3}Z$')

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function setFaSettings(timezone: string, defaultDueTime = '09:00') {
	globalI18n.global.locale.value = 'fa-IR' as never
	useAuthStore().setUserSettings({
		timezone,
		frontendSettings: {defaultDueTime},
	} as never)
}

function mountAbsolute(modelValue?: ITaskReminder) {
	return mount(ReminderDetail, {
		props: {
			modelValue: modelValue ?? (new TaskReminderModel() as ITaskReminder),
			defaultRelativeTo: null,
			allowAbsolute: true,
		},
		global: {
			plugins: [i18n],
			stubs: {
				'flat-pickr': true,
				Icon: true,
			},
			directives: {
				tooltip: () => {},
			},
		},
	})
}

function dayButton(wrapper: VueWrapper, label: string) {
	const found = wrapper.findAll('.jalali-calendar__day').find((button) => button.text() === label)
	if (!found) {
		throw new Error(`day button ${label} not found`)
	}
	return found
}

async function selectDayAndConfirm(wrapper: VueWrapper, jalaliDayLabel: string) {
	await dayButton(wrapper, jalaliDayLabel).trigger('click')
	await wrapper.vm.$nextTick()
	const confirm = wrapper.find('.reminder__close-button')
	expect(confirm.exists()).toBe(true)
	await confirm.trigger('click')
	await wrapper.vm.$nextTick()
	const emitted = wrapper.emitted('update:modelValue')?.pop()?.[0] as ITaskReminder
	if (!emitted) {
		throw new Error('expected update:modelValue to be emitted')
	}
	return emitted
}

function expectGregorianPayload(date: Date) {
	const iso = toISOStringOrNull(date)
	expect(iso).not.toBeNull()
	expect(iso!).toMatch(GREGORIAN_ISO)
	expect(iso!).not.toMatch(JALALI_DIGITS)
	return iso!
}

describe('ReminderDetail fa-IR absolute reminders', () => {
	beforeEach(async () => {
		setActivePinia(createPinia())
		await import('dayjs/locale/fa')
		vi.useFakeTimers()
		vi.setSystemTime(FROZEN_NOW)
	})

	afterEach(() => {
		vi.useRealTimers()
		globalI18n.global.locale.value = 'en' as never
	})

	it.each([
		['Asia/Tehran', '2026-09-11T05:30:00.000Z'],
		['America/Los_Angeles', '2026-09-11T16:00:00.000Z'],
		['UTC', '2026-09-11T09:00:00.000Z'],
	])('Shahrivar 20 09:00 in %s posts Gregorian RFC3339 %s', async (timezone, expectedIso) => {
		setFaSettings(timezone)
		const wrapper = mountAbsolute()
		expect(wrapper.find('.jalali-calendar').exists()).toBe(true)

		const emitted = await selectDayAndConfirm(wrapper, '۲۰')
		expect(emitted.relativeTo).toBeNull()
		expect(emitted.relativePeriod).toBe(0)
		expect(emitted.reminder).not.toBeNull()
		expect(emitted.reminder!.toISOString()).toBe(expectedIso)
		expect(expectGregorianPayload(emitted.reminder!)).toBe(expectedIso)
		wrapper.unmount()
	})

	it('near-midnight 23:59 wall-clock stays on the same Gregorian instant shape', async () => {
		// Helper level: exact wall-clock to instant mapping in each zone.
		expect(jalaliToInstant({year: 1405, month: 6, day: 16, hours: 23, minutes: 59}, 'Asia/Tehran')?.toISOString())
			.toBe('2026-09-07T20:29:00.000Z')
		expect(jalaliToInstant({year: 1405, month: 6, day: 16, hours: 0, minutes: 30}, 'America/Los_Angeles')?.toISOString())
			.toBe('2026-09-07T07:30:00.000Z')
		expect(jalaliToInstant({year: 1405, month: 6, day: 16, hours: 23, minutes: 59}, 'UTC')?.toISOString())
			.toBe('2026-09-07T23:59:00.000Z')

		// Component level: an existing 23:59 Tehran reminder keeps its wall-clock
		// when another day is picked, and the payload stays Gregorian RFC3339.
		setFaSettings('Asia/Tehran')
		const initial = jalaliToInstant({year: 1405, month: 6, day: 16, hours: 23, minutes: 59}, 'Asia/Tehran')
		expect(initial).not.toBeNull()
		const wrapper = mountAbsolute(new TaskReminderModel({reminder: initial!, relativeTo: null}) as ITaskReminder)
		const emitted = await selectDayAndConfirm(wrapper, '۱۷')
		expect(emitted.relativeTo).toBeNull()
		expect(emitted.reminder!.toISOString()).toBe('2026-09-08T20:29:00.000Z')
		expect(expectGregorianPayload(emitted.reminder!)).toBe('2026-09-08T20:29:00.000Z')
		wrapper.unmount()
	})

	it('Esfand boundary maps to Gregorian March and rejects Esfand 30 in common years', async () => {
		// Calendar math: 1404 Esfand has 29 days, 1403 Esfand has 30 (leap).
		expect(getJalaliMonthLength(1404, 12)).toBe(29)
		expect(getJalaliMonthLength(1403, 12)).toBe(30)
		expect(jalaliToInstant({year: 1404, month: 12, day: 30, hours: 9, minutes: 0}, 'Asia/Tehran')).toBeNull()
		expect(jalaliToInstant({year: 1404, month: 12, day: 29, hours: 9, minutes: 0}, 'Asia/Tehran')?.toISOString())
			.toBe('2026-03-20T05:30:00.000Z')
		expect(jalaliToInstant({year: 1404, month: 12, day: 29, hours: 9, minutes: 0}, 'UTC')?.toISOString())
			.toBe('2026-03-20T09:00:00.000Z')
		expect(jalaliToInstant({year: 1405, month: 1, day: 1, hours: 9, minutes: 0}, 'Asia/Tehran')?.toISOString())
			.toBe('2026-03-21T05:30:00.000Z')

		// Round-trip: the same instant formats back to the same Jalali date.
		const esfandInstant = jalaliToInstant({year: 1404, month: 12, day: 29, hours: 9, minutes: 0}, 'Asia/Tehran')!
		expect(instantToJalali(esfandInstant, 'Asia/Tehran')).toMatchObject({year: 1404, month: 12, day: 29})

		// Component level: start from an Esfand reminder so the grid opens on
		// Esfand, pick day 29, and the payload stays Gregorian RFC3339.
		setFaSettings('Asia/Tehran')
		const wrapper = mountAbsolute(new TaskReminderModel({reminder: esfandInstant, relativeTo: null}) as ITaskReminder)
		expect(wrapper.find('.jalali-calendar').exists()).toBe(true)
		expect(wrapper.findAll('.jalali-calendar__day')).toHaveLength(29)
		const emitted = await selectDayAndConfirm(wrapper, '۲۹')
		expect(emitted.reminder!.toISOString()).toBe('2026-03-20T05:30:00.000Z')
		expect(expectGregorianPayload(emitted.reminder!)).toBe('2026-03-20T05:30:00.000Z')
		wrapper.unmount()
	})

	it('absolute payload serializes via toISOStringOrNull with no Jalali digits', async () => {
		setFaSettings('UTC')
		const wrapper = mountAbsolute()
		const emitted = await selectDayAndConfirm(wrapper, '۲۰')
		const wire = toISOStringOrNull(emitted.reminder)
		expect(wire).toBe('2026-09-11T09:00:00.000Z')
		expect(wire!).toMatch(GREGORIAN_ISO)
		expect(wire!).not.toMatch(JALALI_DIGITS)
		wrapper.unmount()
	})
})

describe('ReminderDetail fa-IR relative reminders unchanged', () => {
	beforeEach(async () => {
		setActivePinia(createPinia())
		await import('dayjs/locale/fa')
		vi.useFakeTimers()
		vi.setSystemTime(FROZEN_NOW)
	})

	afterEach(() => {
		vi.useRealTimers()
		globalI18n.global.locale.value = 'en' as never
	})

	it('presets emit seconds plus enum independent of locale and timezone', async () => {
		for (const timezone of ['Asia/Tehran', 'America/Los_Angeles', 'UTC']) {
			setFaSettings(timezone)
			const wrapper = mount(ReminderDetail, {
				props: {
					modelValue: new TaskReminderModel() as ITaskReminder,
					defaultRelativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE,
					allowAbsolute: true,
				},
				global: {
					plugins: [i18n],
					stubs: {
						'flat-pickr': true,
						Icon: true,
					},
					directives: {
						tooltip: () => {},
					},
				},
			})

			const options = wrapper.findAll('.option-button')
			// 6 presets + custom + dateAndTime.
			expect(options).toHaveLength(8)

			await options[1].trigger('click')
			await wrapper.vm.$nextTick()
			const first = wrapper.emitted('update:modelValue')?.pop()?.[0] as ITaskReminder
			expect(first.relativePeriod).toBe(-2 * SECONDS_A_HOUR)
			expect(first.relativeTo).toBe(REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE)
			expect(first.reminder).toBeNull()

			await options[2].trigger('click')
			await wrapper.vm.$nextTick()
			const second = wrapper.emitted('update:modelValue')?.pop()?.[0] as ITaskReminder
			expect(second.relativePeriod).toBe(-1 * SECONDS_A_DAY)
			expect(second.relativeTo).toBe(REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE)
			expect(second.reminder).toBeNull()

			wrapper.unmount()
		}
	})

	it('period helpers stay Gregorian seconds in fa-IR mode', () => {
		setFaSettings('Asia/Tehran')
		expect(periodToSeconds(2, 'hours')).toBe(2 * SECONDS_A_HOUR)
		expect(periodToSeconds(3, 'days')).toBe(3 * SECONDS_A_DAY)
		expect(periodToSeconds(1, 'minutes')).toBe(60)
		expect(periodToSeconds(1, 'weeks')).toBe(7 * SECONDS_A_DAY)

		const model = new TaskReminderModel({
			reminder: null,
			relativePeriod: -1 * SECONDS_A_DAY,
			relativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE,
		}) as ITaskReminder
		expect(model.relativePeriod).toBe(-86400)
		expect(model.relativeTo).toBe('due_date')
		expect(toISOStringOrNull(model.reminder)).toBeNull()
	})
})
