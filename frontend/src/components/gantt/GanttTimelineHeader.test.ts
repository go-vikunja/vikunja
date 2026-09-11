import {afterEach, beforeEach, describe, expect, it} from 'vitest'
import {mount} from '@vue/test-utils'
import {createPinia, setActivePinia} from 'pinia'
import {createI18n} from 'vue-i18n'

import GanttTimelineHeader from './GanttTimelineHeader.vue'
import en from '@/i18n/lang/en.json'
import {i18n as globalI18n} from '@/i18n'
import {useAuthStore} from '@/stores/auth'
import {formatDate} from '@/helpers/time/formatDate'
import {formatJalaliDate, getJalaliMonthLength, instantToJalali, toPersianDigits} from '@/helpers/time/jalali'

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

const DAY_WIDTH = 30

function setLocaleAndTimezone(locale: string, timezone: string) {
	globalI18n.global.locale.value = locale as never
	useAuthStore().setUserSettings({timezone} as never)
}

function utcDays(startIso: string, count: number): Date[] {
	const start = new Date(startIso).getTime()
	return Array.from({length: count}, (_, i) => new Date(start + i * 24 * 60 * 60 * 1000))
}

function mountHeader(timelineData: Date[], dayWidthPixels = DAY_WIDTH) {
	return mount(GanttTimelineHeader, {
		props: {timelineData, dayWidthPixels},
		global: {plugins: [i18n]},
	})
}

function monthGroups(wrapper: ReturnType<typeof mountHeader>) {
	return wrapper.findAll('.timeunit-month')
}

function dayCells(wrapper: ReturnType<typeof mountHeader>) {
	return wrapper.findAll('.gantt-timeline-days .timeunit')
}

describe('GanttTimelineHeader Jalali', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	afterEach(() => {
		globalI18n.global.locale.value = 'en'
	})

	it('regroups across Nowruz (Esfand -> Farvardin) in Asia/Tehran', () => {
		setLocaleAndTimezone('fa-IR', 'Asia/Tehran')
		// 2026-03-18..24 UTC midnights: first three are Esfand 1404, rest Farvardin 1405 in Tehran
		const dates = utcDays('2026-03-18T00:00:00Z', 7)
		const wrapper = mountHeader(dates)

		const groups = monthGroups(wrapper)
		expect(groups).toHaveLength(2)
		expect(groups[0].attributes('style')).toContain(`${3 * DAY_WIDTH}px`)
		expect(groups[1].attributes('style')).toContain(`${4 * DAY_WIDTH}px`)

		const total = groups.reduce((sum, g) => {
			const match = g.attributes('style')?.match(/(\d+)px/)
			return sum + Number(match?.[1] ?? 0)
		}, 0)
		expect(total).toBe(7 * DAY_WIDTH)
	})

	it('regroups across Nowruz in UTC as well', () => {
		setLocaleAndTimezone('fa-IR', 'UTC')
		const dates = utcDays('2026-03-18T00:00:00Z', 7)
		const wrapper = mountHeader(dates)

		const groups = monthGroups(wrapper)
		expect(groups).toHaveLength(2)
		const total = groups.reduce((sum, g) => {
			const match = g.attributes('style')?.match(/(\d+)px/)
			return sum + Number(match?.[1] ?? 0)
		}, 0)
		expect(total).toBe(7 * DAY_WIDTH)
	})

	it('shows a different Jalali day per timezone around midnight', () => {
		// Same instant is Farvardin 1 in Tehran, Esfand 29 in UTC
		const edge = new Date('2026-03-20T20:30:00Z')
		expect(instantToJalali(edge, 'Asia/Tehran')).toMatchObject({year: 1405, month: 1, day: 1})
		expect(instantToJalali(edge, 'UTC')).toMatchObject({year: 1404, month: 12, day: 29})

		setLocaleAndTimezone('fa-IR', 'Asia/Tehran')
		const tehranWrapper = mountHeader([edge])
		expect(dayCells(tehranWrapper)[0].find('span').text()).toBe('۱')

		setLocaleAndTimezone('fa-IR', 'UTC')
		const utcWrapper = mountHeader([edge])
		expect(dayCells(utcWrapper)[0].find('span').text()).toBe('۲۹')
	})

	it('renders a single group for a 31-day Jalali month (Farvardin 1405)', () => {
		expect(getJalaliMonthLength(1405, 1)).toBe(31)
		setLocaleAndTimezone('fa-IR', 'UTC')
		// Farvardin 1405 in UTC: 2026-03-21..2026-04-20
		const dates = utcDays('2026-03-21T00:00:00Z', 31)
		const wrapper = mountHeader(dates)

		const groups = monthGroups(wrapper)
		expect(groups).toHaveLength(1)
		expect(groups[0].attributes('style')).toContain(`${31 * DAY_WIDTH}px`)
		expect(groups[0].text()).toBe(formatJalaliDate(dates[0], {timeZone: 'UTC', month: 'long', year: 'numeric'}))
	})

	it('renders a single group for a 30-day Jalali month (Mehr 1405)', () => {
		expect(getJalaliMonthLength(1405, 7)).toBe(30)
		setLocaleAndTimezone('fa-IR', 'UTC')
		// Mehr 1405 in UTC: 2026-09-23..2026-10-22
		const dates = utcDays('2026-09-23T00:00:00Z', 30)
		const wrapper = mountHeader(dates)

		const groups = monthGroups(wrapper)
		expect(groups).toHaveLength(1)
		expect(groups[0].attributes('style')).toContain(`${30 * DAY_WIDTH}px`)
	})

	it('renders a single group for a 29-day Esfand (1404)', () => {
		expect(getJalaliMonthLength(1404, 12)).toBe(29)
		setLocaleAndTimezone('fa-IR', 'UTC')
		// Esfand 1404 in UTC: 2026-02-20..2026-03-20
		const dates = utcDays('2026-02-20T00:00:00Z', 29)
		const wrapper = mountHeader(dates)

		const groups = monthGroups(wrapper)
		expect(groups).toHaveLength(1)
		expect(groups[0].attributes('style')).toContain(`${29 * DAY_WIDTH}px`)
	})

	it('renders a single group for a 30-day Esfand (1403)', () => {
		expect(getJalaliMonthLength(1403, 12)).toBe(30)
		setLocaleAndTimezone('fa-IR', 'UTC')
		// Esfand 1403 in UTC: 2025-02-19..2025-03-20
		const dates = utcDays('2025-02-19T00:00:00Z', 30)
		const wrapper = mountHeader(dates)

		const groups = monthGroups(wrapper)
		expect(groups).toHaveLength(1)
		expect(groups[0].attributes('style')).toContain(`${30 * DAY_WIDTH}px`)
	})

	it('shows Jalali day numbers in Asia/Tehran', () => {
		setLocaleAndTimezone('fa-IR', 'Asia/Tehran')
		const dates = utcDays('2026-03-18T00:00:00Z', 7)
		const wrapper = mountHeader(dates)
		const cells = dayCells(wrapper)

		expect(cells).toHaveLength(7)
		dates.forEach((date, i) => {
			const expected = instantToJalali(date, 'Asia/Tehran')?.day ?? date.getDate()
			expect(cells[i].find('span').text()).toBe(toPersianDigits(expected))
		})
	})

	it('shows Jalali day numbers in UTC', () => {
		setLocaleAndTimezone('fa-IR', 'UTC')
		const dates = utcDays('2026-09-07T00:00:00Z', 3)
		const wrapper = mountHeader(dates)
		const cells = dayCells(wrapper)

		dates.forEach((date, i) => {
			const expected = instantToJalali(date, 'UTC')?.day ?? date.getDate()
			expect(cells[i].find('span').text()).toBe(toPersianDigits(expected))
		})
	})

	it('uses the central Jalali formatter for day aria', () => {
		setLocaleAndTimezone('fa-IR', 'Asia/Tehran')
		const dates = utcDays('2026-03-18T00:00:00Z', 3)
		const wrapper = mountHeader(dates)
		const cells = dayCells(wrapper)

		dates.forEach((date, i) => {
			const aria = cells[i].attributes('aria-label') ?? ''
			expect(aria).toContain(formatDate(date, 'LL'))
			expect(aria).not.toContain(date.toLocaleDateString())
		})
	})

	it('keeps pixel totals equal to days * dayWidth', () => {
		setLocaleAndTimezone('fa-IR', 'Asia/Tehran')
		const dates = utcDays('2026-02-20T00:00:00Z', 45)
		const wrapper = mountHeader(dates, 42)

		const total = monthGroups(wrapper).reduce((sum, g) => {
			const match = g.attributes('style')?.match(/(\d+)px/)
			return sum + Number(match?.[1] ?? 0)
		}, 0)
		expect(total).toBe(45 * 42)
	})
})

describe('GanttTimelineHeader Gregorian unchanged', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		globalI18n.global.locale.value = 'en'
	})

	afterEach(() => {
		globalI18n.global.locale.value = 'en'
	})

	it('groups by Gregorian month with byte-identical labels and aria', () => {
		const dates = utcDays('2026-03-28T00:00:00Z', 7)
		const wrapper = mountHeader(dates)
		const groups = monthGroups(wrapper)

		// March 28..31 + April 1..3 in machine TZ may vary, so derive expected the same way the component does
		const expectedKeys: string[] = []
		for (const d of dates) {
			const key = `${d.getFullYear()}-${d.getMonth()}`
			if (expectedKeys[expectedKeys.length - 1] !== key) {
				expectedKeys.push(key)
			}
		}
		expect(groups).toHaveLength(expectedKeys.length)

		const cells = dayCells(wrapper)
		dates.forEach((date, i) => {
			expect(cells[i].find('span').text()).toBe(String(date.getDate()))
			expect(cells[i].attributes('aria-label')).toContain(date.toLocaleDateString())
		})

		const total = groups.reduce((sum, g) => {
			const match = g.attributes('style')?.match(/(\d+)px/)
			return sum + Number(match?.[1] ?? 0)
		}, 0)
		expect(total).toBe(dates.length * DAY_WIDTH)
	})
})
