import {afterEach, beforeEach, describe, expect, it} from 'vitest'
import {createPinia, setActivePinia} from 'pinia'

import {i18n} from '@/i18n'
import {useAuthStore} from '@/stores/auth'
import {DATE_DISPLAY} from '@/constants/dateDisplay'
import {TIME_FORMAT} from '@/constants/timeFormat'
import {dateIsValid, formatDate, formatDateLong, formatDateShort, formatDateSince, formatDisplayDateFormat, formatISO, useWeekDayFromDate} from './formatDate'

describe('dateIsValid', () => {
	it.each([
		['a valid date', new Date('2021-02-06T12:00:00Z'), true],
		['a valid date string', '2021-02-06 12:00', true],
		['an invalid date', new Date('not a date'), false],
		['an unparseable string', 'not a date', false],
		['null', null, false],
		['undefined', undefined, false],
		['the api zero time', '0001-01-01T00:00:00Z', false],
	])('returns %s', (_name, date, expected) => {
		expect(dateIsValid(date)).toBe(expected)
	})
})

describe('formatISO', () => {
	it('formats a valid date', () => {
		expect(formatISO(new Date('2021-02-06T12:00:00Z'))).toBe('2021-02-06T12:00:00.000Z')
	})

	it.each([
		['an invalid date', new Date('not a date')],
		['null', null],
		['undefined', undefined],
		['the api zero time', '0001-01-01T00:00:00Z'],
	])('returns an empty string for %s', (_name, date) => {
		expect(formatISO(date)).toBe('')
	})
})

describe('formatDate', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	it('formats a valid date', () => {
		expect(formatDate(new Date(2021, 1, 6, 12, 0), 'YYYY-MM-DD')).toBe('2021-02-06')
	})

	it.each([
		['an invalid date', new Date('not a date')],
		['null', null],
		['undefined', undefined],
		['the api zero time', '0001-01-01T00:00:00Z'],
	])('returns an empty string for %s', (_name, date) => {
		expect(formatDate(date, 'YYYY-MM-DD')).toBe('')
	})
})

describe('Jalali display (fa-IR)', () => {
	// Fixed instants; the user timezone below makes every assertion
	// independent of the machine's local timezone.
	const NOWRUZ_1405 = new Date('2026-03-21T00:00:00Z')
	const SHAHRIVAR_1405 = new Date('2026-09-07T12:00:00Z')
	// 00:00 in Tehran, still the previous day in Los Angeles.
	const MIDNIGHT_EDGE = new Date('2026-03-20T20:30:00Z')

	// The store exposes settings as readonly, so tests go through the action.
	function setTimezone(timezone: string) {
		useAuthStore().setUserSettings({timezone} as never)
	}

	beforeEach(async () => {
		setActivePinia(createPinia())
		// dayjs falls back to english when the locale module was never loaded;
		// the app loads it via useDayjsLanguageSync, tests load it directly.
		await import('dayjs/locale/de')
		await import('dayjs/locale/fa')
		i18n.global.locale.value = 'fa-IR'
		setTimezone('Asia/Tehran')
	})

	afterEach(() => {
		i18n.global.locale.value = 'en'
	})

	it('formats 2026-03-21 as 1405-01-01', () => {
		expect(formatDate(NOWRUZ_1405, 'LL')).toBe('۱ فروردین ۱۴۰۵')
	})

	it('formats 2026-09-07 as 1405-06-16', () => {
		expect(formatDate(SHAHRIVAR_1405, 'LL')).toBe('۱۶ شهریور ۱۴۰۵')
	})

	it('formats long dates with Persian weekday and month names', () => {
		expect(formatDateLong(NOWRUZ_1405)).toBe('۱۴۰۵ فروردین ۱, شنبه ساعت ۰۳:۳۰')
	})

	it('formats short dates', () => {
		expect(formatDateShort(SHAHRIVAR_1405)).toBe('۱۶ شهریور ۱۴۰۵ ساعت ۱۵:۳۰')
	})

	it('formats weekday names in Persian', () => {
		expect(formatDate(NOWRUZ_1405, 'ddd')).toBe('شنبه')
		expect(useWeekDayFromDate().value(NOWRUZ_1405)).toBe('شنبه')
	})

	it('shows a different Jalali date per user timezone around midnight', () => {
		expect(formatDate(MIDNIGHT_EDGE, 'LL')).toBe('۱ فروردین ۱۴۰۵')

		setTimezone('America/Los_Angeles')
		expect(formatDate(MIDNIGHT_EDGE, 'LL')).toBe('۲۹ اسفند ۱۴۰۴')
	})

	it('keeps time formatting unchanged across locales', () => {
		const jalaliTime = formatDate(SHAHRIVAR_1405, 'HH:mm')

		i18n.global.locale.value = 'en'
		expect(formatDate(SHAHRIVAR_1405, 'HH:mm')).toBe(jalaliTime)
		expect(jalaliTime).toMatch(/^\d{2}:\d{2}$/)
	})

	it('renders numeric display modes as Jalali with the mode separator', () => {
		expect(formatDisplayDateFormat(SHAHRIVAR_1405, DATE_DISPLAY.YYYY_SLASH_MM_DD, TIME_FORMAT.HOURS_24))
			.toMatch(/^۱۴۰۵\/۰۶\/۱۶ \d{2}:\d{2}$/)
		expect(formatDisplayDateFormat(SHAHRIVAR_1405, DATE_DISPLAY.YYYY_MM_DD, TIME_FORMAT.HOURS_24))
			.toMatch(/^۱۴۰۵-۰۶-۱۶ \d{2}:\d{2}$/)
	})

	it('renders day-month-year modes with Persian names', () => {
		expect(formatDisplayDateFormat(SHAHRIVAR_1405, DATE_DISPLAY.DAY_MONTH_YEAR, TIME_FORMAT.HOURS_24))
			.toBe('۱۶ شهریور ۱۴۰۵ ساعت ۱۵:۳۰')
		expect(formatDisplayDateFormat(SHAHRIVAR_1405, DATE_DISPLAY.WEEKDAY_DAY_MONTH_YEAR, TIME_FORMAT.HOURS_24))
			.toBe('۱۴۰۵ شهریور ۱۶, دوشنبه ساعت ۱۵:۳۰')
	})

	it('keeps relative time working with Persian wording', () => {
		const threeDaysAgo = new Date(Date.now() - 3 * 24 * 60 * 60 * 1000)
		expect(formatDateSince(threeDaysAgo)).toContain('روز پیش')
		expect(formatDisplayDateFormat(threeDaysAgo, DATE_DISPLAY.RELATIVE)).toContain('روز پیش')
	})

	it.each([
		['an invalid date', new Date('not a date')],
		['null', null],
		['undefined', undefined],
		['the api zero time', '0001-01-01T00:00:00Z'],
	])('returns an empty string for %s', (_name, date) => {
		expect(formatDate(date, 'LL')).toBe('')
		expect(formatDateLong(date)).toBe('')
		expect(formatDisplayDateFormat(date, DATE_DISPLAY.YYYY_SLASH_MM_DD, TIME_FORMAT.HOURS_24)).toBe('')
	})
})

describe('non-Persian locales stay Gregorian', () => {
	const LOCAL_SEPTEMBER = new Date(2026, 8, 7, 12, 0)

	beforeEach(async () => {
		setActivePinia(createPinia())
		await import('dayjs/locale/de')
		await import('dayjs/locale/fa')
	})

	afterEach(() => {
		i18n.global.locale.value = 'en'
	})

	it('keeps english output unchanged', () => {
		i18n.global.locale.value = 'en'
		expect(formatDate(LOCAL_SEPTEMBER, 'LL')).toBe('September 7, 2026')
		expect(formatDateLong(LOCAL_SEPTEMBER)).toBe('Monday, September 7, 2026 12:00 PM')
		expect(formatDisplayDateFormat(LOCAL_SEPTEMBER, DATE_DISPLAY.YYYY_SLASH_MM_DD, TIME_FORMAT.HOURS_24))
			.toBe('2026/09/07 12:00')
		expect(formatDisplayDateFormat(LOCAL_SEPTEMBER, DATE_DISPLAY.DAY_MONTH_YEAR, TIME_FORMAT.HOURS_24))
			.toBe('September 7, 2026 at 12:00')
	})

	it('keeps german output unchanged', () => {
		i18n.global.locale.value = 'de-DE'
		expect(formatDate(LOCAL_SEPTEMBER, 'LL')).toBe('7. September 2026')
		expect(formatDisplayDateFormat(LOCAL_SEPTEMBER, DATE_DISPLAY.YYYY_SLASH_MM_DD, TIME_FORMAT.HOURS_24))
			.toBe('2026/09/07 12:00')
	})
})
