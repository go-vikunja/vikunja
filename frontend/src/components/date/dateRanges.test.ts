import {describe, expect, it} from 'vitest'
import {
	DATE_RANGES,
	DATE_VALUES,
	formatGregorianKebab,
	getFaMonthBounds,
	getFaWeekBounds,
	getFaYearBounds,
	resolveFaDateRange,
	resolveFaDateValue,
} from './dateRanges'
import {instantToJalali} from '@/helpers/time/jalali'

const MONDAY_SEP_7 = new Date('2026-09-07T12:00:00.000Z')
const JALALI_DIGITS = new RegExp('[۰-۹٠-٩]')
const GREGORIAN_ISO = new RegExp('^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}\\.\\d{3}Z$')
const GREGORIAN_KEBAB = new RegExp('^\\d{4}-\\d{2}-\\d{2}$')

describe('DATE_RANGES/DATE_VALUES stay Gregorian', () => {
	it('keeps week datemath byte-identical', () => {
		expect(DATE_RANGES.lastWeek).toEqual(['now/w-1w', 'now/w'])
		expect(DATE_RANGES.thisWeek).toEqual(['now/w', 'now/w+1w'])
		expect(DATE_RANGES.restOfThisWeek).toEqual(['now', 'now/w+1w'])
		expect(DATE_RANGES.nextWeek).toEqual(['now/w+1w', 'now/w+2w'])
	})

	it('keeps month/year datemath byte-identical', () => {
		expect(DATE_RANGES.lastMonth).toEqual(['now/M-1M', 'now/M'])
		expect(DATE_RANGES.thisMonth).toEqual(['now/M', 'now/M+1M'])
		expect(DATE_RANGES.restOfThisMonth).toEqual(['now', 'now/M+1M'])
		expect(DATE_RANGES.nextMonth).toEqual(['now/M+1M', 'now/M+2M'])
		expect(DATE_RANGES.thisYear).toEqual(['now/y', 'now/y+1y'])
		expect(DATE_RANGES.restOfThisYear).toEqual(['now', 'now/y+1y'])
	})

	it('keeps day datemath byte-identical', () => {
		expect(DATE_RANGES.today).toEqual(['now/d', 'now/d+1d'])
		expect(DATE_RANGES.tomorrow).toEqual(['now/d+1d', 'now/d+2d'])
		expect(DATE_RANGES.next7Days).toEqual(['now', 'now+7d'])
		expect(DATE_RANGES.next30Days).toEqual(['now', 'now+30d'])
		expect(DATE_VALUES.now).toBe('now')
		expect(DATE_VALUES.in7Days).toBe('now+7d')
		expect(DATE_VALUES.in30Days).toBe('now+30d')
	})
})

describe('getFaWeekBounds Saturday boundaries', () => {
	it.each([
		['Asia/Tehran', '2026-09-04T20:30:00.000Z', '2026-09-11T20:30:00.000Z'],
		['America/Los_Angeles', '2026-09-05T07:00:00.000Z', '2026-09-12T07:00:00.000Z'],
		['UTC', '2026-09-05T00:00:00.000Z', '2026-09-12T00:00:00.000Z'],
	])('this week containing 2026-09-07 in %s is Saturday→Saturday', (timeZone, expectedStart, expectedEnd) => {
		const bounds = getFaWeekBounds(MONDAY_SEP_7, timeZone)
		expect(bounds).not.toBeNull()
		expect(bounds!.start.toISOString()).toBe(expectedStart)
		expect(bounds!.end.toISOString()).toBe(expectedEnd)
		const startJalali = instantToJalali(bounds!.start, timeZone)
		const endJalali = instantToJalali(bounds!.end, timeZone)
		expect(startJalali).toMatchObject({day: 14, hours: 0, minutes: 0})
		expect(endJalali).toMatchObject({day: 21, hours: 0, minutes: 0})
		expect(bounds!.start.toISOString()).toMatch(GREGORIAN_ISO)
		expect(bounds!.start.toISOString()).not.toMatch(JALALI_DIGITS)
	})

	it('Saturday itself starts a new week', () => {
		const saturdayNoonTehran = new Date('2026-09-05T08:30:00.000Z')
		const bounds = getFaWeekBounds(saturdayNoonTehran, 'Asia/Tehran')
		expect(bounds!.start.toISOString()).toBe('2026-09-04T20:30:00.000Z')
	})

	it('Friday ends the week', () => {
		const fridayNoonTehran = new Date('2026-09-11T08:30:00.000Z')
		const bounds = getFaWeekBounds(fridayNoonTehran, 'Asia/Tehran')
		expect(bounds!.start.toISOString()).toBe('2026-09-04T20:30:00.000Z')
		expect(bounds!.end.toISOString()).toBe('2026-09-11T20:30:00.000Z')
	})
})

describe('getFaMonthBounds 29/30/31-day months', () => {
	it.each([
		['Asia/Tehran', '2026-08-22T20:30:00.000Z', '2026-09-22T20:30:00.000Z'],
		['America/Los_Angeles', '2026-08-23T07:00:00.000Z', '2026-09-23T07:00:00.000Z'],
		['UTC', '2026-08-23T00:00:00.000Z', '2026-09-23T00:00:00.000Z'],
	])('Shahrivar (31d) in %s', (timeZone, expectedStart, expectedEnd) => {
		const bounds = getFaMonthBounds(MONDAY_SEP_7, timeZone)
		expect(bounds!.start.toISOString()).toBe(expectedStart)
		expect(bounds!.end.toISOString()).toBe(expectedEnd)
		expect(instantToJalali(bounds!.start, timeZone)).toMatchObject({month: 6, day: 1})
		expect(instantToJalali(bounds!.end, timeZone)).toMatchObject({month: 7, day: 1})
	})

	it('Mehr (30d) in Asia/Tehran', () => {
		const ref = new Date('2026-10-07T12:00:00.000Z')
		const bounds = getFaMonthBounds(ref, 'Asia/Tehran')
		expect(bounds!.start.toISOString()).toBe('2026-09-22T20:30:00.000Z')
		expect(bounds!.end.toISOString()).toBe('2026-10-22T20:30:00.000Z')
		expect(instantToJalali(bounds!.start, 'Asia/Tehran')).toMatchObject({month: 7, day: 1})
		expect(instantToJalali(bounds!.end, 'Asia/Tehran')).toMatchObject({month: 8, day: 1})
	})

	it('Esfand 29d (common year 1404) in Asia/Tehran', () => {
		const ref = new Date('2026-03-10T12:00:00.000Z')
		const bounds = getFaMonthBounds(ref, 'Asia/Tehran')
		expect(bounds!.start.toISOString()).toBe('2026-02-19T20:30:00.000Z')
		expect(bounds!.end.toISOString()).toBe('2026-03-20T20:30:00.000Z')
		expect(instantToJalali(bounds!.start, 'Asia/Tehran')).toMatchObject({year: 1404, month: 12, day: 1})
		expect(instantToJalali(bounds!.end, 'Asia/Tehran')).toMatchObject({year: 1405, month: 1, day: 1})
	})

	it('Esfand 30d (leap year 1403) in Asia/Tehran', () => {
		const ref = new Date('2025-03-10T12:00:00.000Z')
		const bounds = getFaMonthBounds(ref, 'Asia/Tehran')
		expect(bounds!.start.toISOString()).toBe('2025-02-18T20:30:00.000Z')
		expect(bounds!.end.toISOString()).toBe('2025-03-20T20:30:00.000Z')
		expect(instantToJalali(bounds!.start, 'Asia/Tehran')).toMatchObject({year: 1403, month: 12, day: 1})
		expect(instantToJalali(bounds!.end, 'Asia/Tehran')).toMatchObject({year: 1404, month: 1, day: 1})
	})
})

describe('Esfand→Farvardin rollover', () => {
	it('year bounds in Asia/Tehran roll from Farvardin 1 to next Farvardin 1', () => {
		const bounds = getFaYearBounds(MONDAY_SEP_7, 'Asia/Tehran')
		expect(bounds!.start.toISOString()).toBe('2026-03-20T20:30:00.000Z')
		expect(instantToJalali(bounds!.start, 'Asia/Tehran')).toMatchObject({year: 1405, month: 1, day: 1})
		expect(instantToJalali(bounds!.end, 'Asia/Tehran')).toMatchObject({year: 1406, month: 1, day: 1})
		expect(bounds!.end.toISOString()).toMatch(GREGORIAN_ISO)
		expect(bounds!.end.toISOString()).not.toMatch(JALALI_DIGITS)
	})

	it.each([
		['Asia/Tehran'],
		['America/Los_Angeles'],
		['UTC'],
	])('lastMonth from Farvardin lands in Esfand in %s', (timeZone) => {
		const farvardinRef = new Date('2026-03-25T12:00:00.000Z')
		const month = getFaMonthBounds(farvardinRef, timeZone)!
		const prev = getFaMonthBounds(new Date(month.start.getTime() - 1), timeZone)!
		expect(instantToJalali(prev.start, timeZone)).toMatchObject({month: 12})
		expect(prev.end.toISOString()).toBe(month.start.toISOString())
	})

	it.each([
		['Asia/Tehran'],
		['America/Los_Angeles'],
		['UTC'],
	])('nextMonth from Esfand lands in Farvardin in %s', (timeZone) => {
		const esfandRef = new Date('2026-03-10T12:00:00.000Z')
		const month = getFaMonthBounds(esfandRef, timeZone)!
		const next = getFaMonthBounds(new Date(month.end.getTime() + 1), timeZone)!
		expect(instantToJalali(next.start, timeZone)).toMatchObject({month: 1, day: 1})
		expect(next.start.toISOString()).toBe(month.end.toISOString())
	})
})

describe('resolveFaDateRange preset mapping', () => {
	it.each([
		['Asia/Tehran'],
		['America/Los_Angeles'],
		['UTC'],
	])('day presets keep datemath verbatim in %s', (timeZone) => {
		expect(resolveFaDateRange('now/d', 'now/d+1d', MONDAY_SEP_7, timeZone)).toEqual(['now/d', 'now/d+1d'])
		expect(resolveFaDateRange('now/d+1d', 'now/d+2d', MONDAY_SEP_7, timeZone)).toEqual(['now/d+1d', 'now/d+2d'])
		expect(resolveFaDateRange('now', 'now+7d', MONDAY_SEP_7, timeZone)).toEqual(['now', 'now+7d'])
		expect(resolveFaDateRange('now', 'now+30d', MONDAY_SEP_7, timeZone)).toEqual(['now', 'now+30d'])
		expect(resolveFaDateValue('now', MONDAY_SEP_7, timeZone)).toBe('now')
		expect(resolveFaDateValue('now/d', MONDAY_SEP_7, timeZone)).toBe('now/d')
		expect(resolveFaDateValue('now+7d', MONDAY_SEP_7, timeZone)).toBe('now+7d')
	})

	it('thisWeek becomes explicit Saturday bounds', () => {
		const [from, to] = resolveFaDateRange('now/w', 'now/w+1w', MONDAY_SEP_7, 'Asia/Tehran')
		expect(from).toBe('2026-09-04T20:30:00.000Z')
		expect(to).toBe('2026-09-11T20:30:00.000Z')
		expect(from).toMatch(GREGORIAN_ISO)
		expect(from).not.toMatch(JALALI_DIGITS)
	})

	it('rest presets keep now and resolve the end', () => {
		const [from, to] = resolveFaDateRange('now', 'now/w+1w', MONDAY_SEP_7, 'Asia/Tehran')
		expect(from).toBe('now')
		expect(to).toBe('2026-09-11T20:30:00.000Z')
		const [, monthTo] = resolveFaDateRange('now', 'now/M+1M', MONDAY_SEP_7, 'Asia/Tehran')
		expect(monthTo).toBe('2026-09-22T20:30:00.000Z')
	})

	it('month presets become explicit Jalali month bounds', () => {
		const [from, to] = resolveFaDateRange('now/M', 'now/M+1M', MONDAY_SEP_7, 'Asia/Tehran')
		expect(from).toBe('2026-08-22T20:30:00.000Z')
		expect(to).toBe('2026-09-22T20:30:00.000Z')
	})

	it('year presets span Farvardin 1 to Esfand end', () => {
		const [from, to] = resolveFaDateRange('now/y', 'now/y+1y', MONDAY_SEP_7, 'Asia/Tehran')
		expect(from).toBe('2026-03-20T20:30:00.000Z')
		expect(instantToJalali(new Date(to), 'Asia/Tehran')).toMatchObject({year: 1406, month: 1, day: 1})
	})

	it('single values resolve consistently with ranges', () => {
		const weekStart = resolveFaDateValue('now/w', MONDAY_SEP_7, 'Asia/Tehran')
		const weekEnd = resolveFaDateValue('now/w+1w', MONDAY_SEP_7, 'Asia/Tehran')
		const [rangeFrom, rangeTo] = resolveFaDateRange('now/w', 'now/w+1w', MONDAY_SEP_7, 'Asia/Tehran')
		expect(weekStart).toBe(rangeFrom)
		expect(weekEnd).toBe(rangeTo)
	})
})

describe('Gregorian wire (no Jalali digits)', () => {
	it('resolved bounds are Gregorian ISO/kebab only', () => {
		const pairs: [string, string][] = [
			['now/w-1w', 'now/w'],
			['now/w', 'now/w+1w'],
			['now', 'now/w+1w'],
			['now/w+1w', 'now/w+2w'],
			['now/M-1M', 'now/M'],
			['now/M', 'now/M+1M'],
			['now', 'now/M+1M'],
			['now/M+1M', 'now/M+2M'],
			['now/y', 'now/y+1y'],
			['now', 'now/y+1y'],
		]
		for (const timeZone of ['Asia/Tehran', 'America/Los_Angeles', 'UTC']) {
			for (const [from, to] of pairs) {
				const [resolvedFrom, resolvedTo] = resolveFaDateRange(from, to, MONDAY_SEP_7, timeZone)
				for (const value of [resolvedFrom, resolvedTo]) {
					if (value === 'now') {
						continue
					}
					expect(value).toMatch(GREGORIAN_ISO)
					expect(value).not.toMatch(JALALI_DIGITS)
				}
			}
		}
	})

	it('kebab formatter stays Gregorian wall-clock', () => {
		expect(formatGregorianKebab(new Date('2026-09-06T20:30:00.000Z'), 'Asia/Tehran')).toBe('2026-09-07')
		expect(formatGregorianKebab(new Date('2026-09-07T00:00:00.000Z'), 'UTC')).toBe('2026-09-07')
		expect(formatGregorianKebab(new Date('2026-09-07T07:00:00.000Z'), 'America/Los_Angeles')).toBe('2026-09-07')
		expect(formatGregorianKebab(new Date('2026-09-06T20:30:00.000Z'), 'Asia/Tehran')).toMatch(GREGORIAN_KEBAB)
		expect(formatGregorianKebab(new Date('2026-09-06T20:30:00.000Z'), 'Asia/Tehran')).not.toMatch(JALALI_DIGITS)
	})
})
