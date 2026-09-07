import {describe, expect, it} from 'vitest'

import {
	addJalaliDays,
	formatJalaliDate,
	formatJalaliParts,
	getJalaliMonthLength,
	gregorianToJalali,
	instantToJalali,
	isJalaliLocale,
	jalaliToGregorian,
	jalaliToInstant,
	jalaliWeekday,
	JALALI_CALENDAR_LATIN_DIGITS_LOCALE,
	normalizePersianDigits,
	parseJalaliDateInput,
	resolveUserTimezone,
	toPersianDigits,
	weekdayInTimezone,
} from './jalali'

// Fixed instants + explicit timeZone:'UTC' keep every assertion independent
// of the machine's local timezone.
const NOWRUZ_1405 = new Date('2026-03-21T00:00:00Z')
const SHAHRIVAR_1405 = new Date('2026-09-07T00:00:00Z')
const NUMERIC_UTC = {timeZone: 'UTC', year: 'numeric', month: '2-digit', day: '2-digit'} as const

describe('isJalaliLocale', () => {
	it.each([
		['fa-IR'],
		['fa-ir'],
		['fa'],
		[' fa-IR '],
	])('recognizes %s as Jalali', (locale) => {
		expect(isJalaliLocale(locale)).toBe(true)
	})

	it.each([
		['en-US'],
		['en'],
		['de-DE'],
		['ar-SA'],
		['he-IL'],
		[''],
		[null],
		[undefined],
	])('does not recognize %s as Jalali', (locale) => {
		expect(isJalaliLocale(locale)).toBe(false)
	})
})

describe('normalizePersianDigits', () => {
	it('converts Persian digits to ASCII', () => {
		expect(normalizePersianDigits('۰۱۲۳۴۵۶۷۸۹')).toBe('0123456789')
	})

	it('converts Arabic-Indic digits to ASCII', () => {
		expect(normalizePersianDigits('٠١٢٣٤٥٦٧٨٩')).toBe('0123456789')
	})

	it('handles mixed strings and leaves other characters unchanged', () => {
		expect(normalizePersianDigits('ساعت ۱۲:۳۰')).toBe('ساعت 12:30')
		expect(normalizePersianDigits('۱۴۰۳/۰۵/۱۷')).toBe('1403/05/17')
		expect(normalizePersianDigits('۱۲٠٣')).toBe('1203')
	})

	it('leaves ASCII strings unchanged', () => {
		expect(normalizePersianDigits('2026-09-07 12:30')).toBe('2026-09-07 12:30')
		expect(normalizePersianDigits('')).toBe('')
		expect(normalizePersianDigits('hello')).toBe('hello')
	})
})

describe('toPersianDigits', () => {
	it('converts ASCII digits to Persian digits', () => {
		expect(toPersianDigits('0123456789')).toBe('۰۱۲۳۴۵۶۷۸۹')
		expect(toPersianDigits(16)).toBe('۱۶')
	})

	it('leaves other characters unchanged', () => {
		expect(toPersianDigits('1405/06/16')).toBe('۱۴۰۵/۰۶/۱۶')
		expect(toPersianDigits('')).toBe('')
	})

	it('round-trips with normalizePersianDigits', () => {
		expect(normalizePersianDigits(toPersianDigits('1405/06/16 15:30'))).toBe('1405/06/16 15:30')
	})
})

describe('resolveUserTimezone', () => {
	it('prefers the configured timezone', () => {
		expect(resolveUserTimezone('Asia/Tehran')).toBe('Asia/Tehran')
	})

	it.each([[null], [undefined], [''], ['   ']])('falls back for %s', (value) => {
		const resolved = resolveUserTimezone(value)
		expect(typeof resolved).toBe('string')
		expect(resolved.length).toBeGreaterThan(0)
	})

	it('falls back for an invalid timezone instead of throwing', () => {
		const resolved = resolveUserTimezone('Not/AZone')
		expect(typeof resolved).toBe('string')
		expect(resolved.length).toBeGreaterThan(0)
	})
})

describe('formatJalaliDate', () => {
	it('formats 2026-03-21 as 1405-01-01', () => {
		expect(formatJalaliDate(NOWRUZ_1405, NUMERIC_UTC, JALALI_CALENDAR_LATIN_DIGITS_LOCALE)).toBe('1405/01/01')
	})

	it('formats 2026-09-07 as 1405-06-16', () => {
		expect(formatJalaliDate(SHAHRIVAR_1405, NUMERIC_UTC, JALALI_CALENDAR_LATIN_DIGITS_LOCALE)).toBe('1405/06/16')
	})

	it('uses Persian digits with the default locale', () => {
		expect(formatJalaliDate(SHAHRIVAR_1405, NUMERIC_UTC)).toBe('۱۴۰۵/۰۶/۱۶')
	})

	it('honours partial options without injecting date components', () => {
		expect(formatJalaliDate(new Date('2026-03-21T00:00:00Z'), {timeZone: 'UTC', weekday: 'short'})).toBe('شنبه')
	})

	it.each([
		['an invalid date', new Date('not a date')],
		['null', null],
		['undefined', undefined],
		['an empty string', ''],
		['the api zero time', '0001-01-01T00:00:00Z'],
	])('returns an empty string for %s instead of throwing', (_name, date) => {
		expect(formatJalaliDate(date, NUMERIC_UTC, JALALI_CALENDAR_LATIN_DIGITS_LOCALE)).toBe('')
	})
})

describe('formatJalaliParts', () => {
	it('returns Jalali parts for a valid date', () => {
		const parts = formatJalaliParts(NOWRUZ_1405, NUMERIC_UTC, JALALI_CALENDAR_LATIN_DIGITS_LOCALE)
		const byType = Object.fromEntries(parts.map((part) => [part.type, part.value]))

		expect(byType).toMatchObject({year: '1405', month: '01', day: '01'})
	})

	it('returns an empty array for unusable dates', () => {
		expect(formatJalaliParts(null)).toEqual([])
		expect(formatJalaliParts(undefined)).toEqual([])
		expect(formatJalaliParts(new Date('not a date'))).toEqual([])
		expect(formatJalaliParts('0001-01-01T00:00:00Z')).toEqual([])
	})
})

describe('gregorianToJalali / jalaliToGregorian', () => {
	it.each([
		[2026, 3, 21, 1405, 1, 1],
		[2026, 9, 7, 1405, 6, 16],
		[2026, 3, 20, 1404, 12, 29],
		[2024, 3, 19, 1402, 12, 29],
		[2024, 3, 20, 1403, 1, 1],
		[2025, 3, 20, 1403, 12, 30],
		[2025, 3, 21, 1404, 1, 1],
	])('converts %i-%i-%i to %i/%i/%i and back', (gy, gm, gd, jy, jm, jd) => {
		expect(gregorianToJalali(gy, gm, gd)).toEqual({year: jy, month: jm, day: jd})
		expect(jalaliToGregorian({year: jy, month: jm, day: jd})).toEqual({year: gy, month: gm, day: gd})
	})

	it.each([
		[{year: 1405, month: 13, day: 1}],
		[{year: 1405, month: 0, day: 1}],
		[{year: 1405, month: 6, day: 32}],
		[{year: 1404, month: 12, day: 30}],
		[{year: 1405.5, month: 1, day: 1}],
	])('returns null for impossible date %o', (date) => {
		expect(jalaliToGregorian(date)).toBeNull()
	})
})

describe('getJalaliMonthLength / jalaliWeekday', () => {
	it.each([
		[1405, 1, 31],
		[1405, 6, 31],
		[1405, 7, 30],
		[1405, 11, 30],
		[1404, 12, 29],
		[1403, 12, 30],
	])('month %i of %i has %i days', (year, month, length) => {
		expect(getJalaliMonthLength(year, month)).toBe(length)
	})

	it('1405/01/01 is a Saturday', () => {
		expect(jalaliWeekday({year: 1405, month: 1, day: 1})).toBe(6)
	})

	it('returns null for impossible dates', () => {
		expect(jalaliWeekday({year: 1405, month: 13, day: 1})).toBeNull()
	})
})

describe('instantToJalali / jalaliToInstant', () => {
	it('splits an instant into Jalali date and wall-clock time per timezone', () => {
		expect(instantToJalali(new Date('2026-03-20T20:30:00Z'), 'Asia/Tehran'))
			.toEqual({year: 1405, month: 1, day: 1, hours: 0, minutes: 0})
		expect(instantToJalali(new Date('2026-03-20T20:30:00Z'), 'America/Los_Angeles'))
			.toEqual({year: 1404, month: 12, day: 29, hours: 13, minutes: 30})
		expect(instantToJalali(new Date('2026-09-07T12:00:00Z'), 'Asia/Tehran'))
			.toEqual({year: 1405, month: 6, day: 16, hours: 15, minutes: 30})
	})

	it('builds the correct instant for a Jalali wall-clock time', () => {
		expect(jalaliToInstant({year: 1405, month: 1, day: 1, hours: 0, minutes: 0}, 'Asia/Tehran')?.toISOString())
			.toBe('2026-03-20T20:30:00.000Z')
		expect(jalaliToInstant({year: 1404, month: 12, day: 29, hours: 0, minutes: 0}, 'America/Los_Angeles')?.toISOString())
			.toBe('2026-03-20T07:00:00.000Z')
		expect(jalaliToInstant({year: 1405, month: 6, day: 16, hours: 15, minutes: 30}, 'Asia/Tehran')?.toISOString())
			.toBe('2026-09-07T12:00:00.000Z')
		expect(jalaliToInstant({year: 1405, month: 6, day: 16, hours: 15, minutes: 30}, 'UTC')?.toISOString())
			.toBe('2026-09-07T15:30:00.000Z')
	})

	it('round-trips across timezones including a DST zone', () => {
		const instants = [
			new Date('2026-01-15T12:00:00Z'),
			new Date('2026-03-20T20:30:00Z'),
			new Date('2026-07-01T12:00:00Z'),
			new Date('2026-09-07T12:00:00Z'),
			new Date('2024-03-19T20:30:00Z'),
		]
		for (const instant of instants) {
			for (const timeZone of ['Asia/Tehran', 'America/Los_Angeles', 'UTC']) {
				const parts = instantToJalali(instant, timeZone)
				expect(parts).not.toBeNull()
				expect(jalaliToInstant(parts!, timeZone)?.toISOString()).toBe(instant.toISOString())
			}
		}
	})

	it.each([
		['null', null],
		['undefined', undefined],
		['an invalid date', new Date('not a date')],
		['the api zero time', '0001-01-01T00:00:00Z'],
	])('instantToJalali returns null for %s', (_name, date) => {
		expect(instantToJalali(date, 'Asia/Tehran')).toBeNull()
	})

	it.each([
		['Esfand 30 in a common year', {year: 1404, month: 12, day: 30, hours: 0, minutes: 0}],
		['month 13', {year: 1405, month: 13, day: 1, hours: 0, minutes: 0}],
		['hour 24', {year: 1405, month: 1, day: 1, hours: 24, minutes: 0}],
		['minute 60', {year: 1405, month: 1, day: 1, hours: 0, minutes: 60}],
	])('jalaliToInstant returns null for %s', (_name, date) => {
		expect(jalaliToInstant(date, 'Asia/Tehran')).toBeNull()
	})

	it('falls back instead of throwing for a bad timezone', () => {
		expect(() => jalaliToInstant({year: 1405, month: 1, day: 1, hours: 0, minutes: 0}, 'Not/AZone')).not.toThrow()
		expect(() => instantToJalali(new Date('2026-03-21T00:00:00Z'), 'Not/AZone')).not.toThrow()
	})
})

describe('parseJalaliDateInput', () => {
	it.each([
		['1405/06/16'],
		['1405-06-16'],
		['1405.06.16'],
		['1405/6/16'],
		['  1405/06/16  '],
	])('parses %s at midnight in Asia/Tehran', (raw) => {
		expect(parseJalaliDateInput(raw, {timeZone: 'Asia/Tehran'})?.toISOString())
			.toBe('2026-09-06T20:30:00.000Z')
	})

	it('parses single-digit day 1405/06/6', () => {
		expect(parseJalaliDateInput('1405/06/6', {timeZone: 'Asia/Tehran'})?.toISOString())
			.toBe('2026-08-27T20:30:00.000Z')
	})

	it.each([
		['1405/06/16'],
		['1405-06-16'],
		['1405.06.16'],
	])('parses %s at midnight in UTC', (raw) => {
		expect(parseJalaliDateInput(raw, {timeZone: 'UTC'})?.toISOString())
			.toBe('2026-09-07T00:00:00.000Z')
	})

	it('parses 1405/06/16 at midnight in America/Los_Angeles', () => {
		expect(parseJalaliDateInput('1405/06/16', {timeZone: 'America/Los_Angeles'})?.toISOString())
			.toBe('2026-09-07T07:00:00.000Z')
	})

	it('parses 1405/01/01 per timezone', () => {
		expect(parseJalaliDateInput('1405/01/01', {timeZone: 'Asia/Tehran'})?.toISOString())
			.toBe('2026-03-20T20:30:00.000Z')
		expect(parseJalaliDateInput('1405/01/01', {timeZone: 'UTC'})?.toISOString())
			.toBe('2026-03-21T00:00:00.000Z')
		expect(parseJalaliDateInput('1405/01/01', {timeZone: 'America/Los_Angeles'})?.toISOString())
			.toBe('2026-03-21T07:00:00.000Z')
	})

	it.each([
		['۱۴۰۵/۰۶/۱۶'],
		['١٤٠٥/٠٦/١٦'],
		['۱۴۰۵-۰۶-۱۶'],
	])('parses non-ASCII digits %s', (raw) => {
		expect(parseJalaliDateInput(raw, {timeZone: 'UTC'})?.toISOString())
			.toBe('2026-09-07T00:00:00.000Z')
	})

	it('parses Persian digits with time', () => {
		expect(parseJalaliDateInput('۱۴۰۵/۰۶/۱۶ ۱۵:۳۰', {timeZone: 'Asia/Tehran'})?.toISOString())
			.toBe('2026-09-07T12:00:00.000Z')
	})

	it('parses an HH:MM variant in the given timezone', () => {
		expect(parseJalaliDateInput('1405/06/16 15:30', {timeZone: 'Asia/Tehran'})?.toISOString())
			.toBe('2026-09-07T12:00:00.000Z')
		expect(parseJalaliDateInput('1405/06/16 15:30', {timeZone: 'UTC'})?.toISOString())
			.toBe('2026-09-07T15:30:00.000Z')
		expect(parseJalaliDateInput('1405/06/16 00:00', {timeZone: 'UTC'})?.toISOString())
			.toBe('2026-09-07T00:00:00.000Z')
	})

	it.each([
		['month 13', '1405/13/01'],
		['month 0', '1405/00/01'],
		['day 32', '1405/06/32'],
		['day 0', '1405/06/00'],
		['Esfand 30 in common year 1404', '1404/12/30'],
		['hour 24', '1405/06/16 24:00'],
		['minute 60', '1405/06/16 12:60'],
		['year 1299 below window', '1299/01/01'],
		['year 1501 above window', '1501/01/01'],
		['Gregorian year 2026', '2026/09/07'],
		['Gregorian year with dashes', '2026-09-07'],
	])('returns null for invalid %s (%s)', (_name, raw) => {
		expect(parseJalaliDateInput(raw, {timeZone: 'Asia/Tehran'})).toBeNull()
		expect(parseJalaliDateInput(raw, {timeZone: 'UTC'})).toBeNull()
	})

	it.each([
		['ISO date', '2026-09-07'],
		['ISO datetime Z', '2026-09-07T12:00:00.000Z'],
		['ISO datetime', '2026-09-07T12:00:00Z'],
		['ISO with slashes and time', '2026/09/07 12:00'],
		['datemath now', 'now'],
		['datemath week', 'now/w'],
		['datemath range', 'now/w+1w'],
		['day-first ambiguous', '17/06/1405'],
		['day-first dashes', '16-06-1405'],
		['partial date', '1405/06'],
		['time only', '15:30'],
	])('returns null for %s %s instead of a Gregorian reading', (_name, raw) => {
		expect(parseJalaliDateInput(raw, {timeZone: 'Asia/Tehran'})).toBeNull()
		expect(parseJalaliDateInput(raw, {timeZone: 'UTC'})).toBeNull()
	})

	it.each([
		['empty string', ''],
		['blank string', '   '],
		['garbage', 'not a date'],
		['garbage with digits', 'foo 123'],
		['null', null],
		['undefined', undefined],
	])('returns null for %s without throwing', (_name, raw) => {
		expect(() => parseJalaliDateInput(raw as string, {timeZone: 'UTC'})).not.toThrow()
		expect(parseJalaliDateInput(raw as string, {timeZone: 'UTC'})).toBeNull()
	})

	it('falls back instead of throwing for a bad timezone', () => {
		expect(() => parseJalaliDateInput('1405/06/16', {timeZone: 'Not/AZone'})).not.toThrow()
		expect(parseJalaliDateInput('1405/06/16', {timeZone: 'Not/AZone'})).not.toBeNull()
	})
})

describe('addJalaliDays / weekdayInTimezone', () => {
	it.each([
		[{year: 1405, month: 1, day: 1}, 0, {year: 1405, month: 1, day: 1}],
		[{year: 1405, month: 1, day: 1}, 1, {year: 1405, month: 1, day: 2}],
		[{year: 1405, month: 1, day: 31}, 1, {year: 1405, month: 2, day: 1}],
		[{year: 1405, month: 6, day: 31}, 1, {year: 1405, month: 7, day: 1}],
		[{year: 1404, month: 12, day: 29}, 1, {year: 1405, month: 1, day: 1}],
		[{year: 1403, month: 12, day: 30}, 1, {year: 1404, month: 1, day: 1}],
		[{year: 1405, month: 1, day: 1}, -1, {year: 1404, month: 12, day: 29}],
		[{year: 1405, month: 6, day: 16}, 7, {year: 1405, month: 6, day: 23}],
	])('adds %i days to %o', (date, days, expected) => {
		expect(addJalaliDays(date, days)).toEqual(expected)
	})

	it('anchors the weekday to the given timezone', () => {
		// 2026-03-20T20:30Z is Saturday in Tehran, Friday in Los Angeles.
		expect(weekdayInTimezone(new Date('2026-03-20T20:30:00Z'), 'Asia/Tehran')).toBe(6)
		expect(weekdayInTimezone(new Date('2026-03-20T20:30:00Z'), 'America/Los_Angeles')).toBe(5)
	})
})
