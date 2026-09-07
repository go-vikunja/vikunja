import {describe, expect, it} from 'vitest'
import {DATE_RANGES, resolveFaDateRange, resolveFaDateValue} from '@/components/date/dateRanges'

const FIXED_NOW = new Date('2026-09-07T12:00:00.000Z')
const JALALI_DIGITS = new RegExp('[۰-۹٠-٩]')
const GREGORIAN_ISO = new RegExp('^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}\\.\\d{3}Z$')

describe('ShowTasks fa upcoming ranges', () => {
	it.each([
		['Asia/Tehran', '2026-09-04T20:30:00.000Z', '2026-09-11T20:30:00.000Z'],
		['America/Los_Angeles', '2026-09-05T07:00:00.000Z', '2026-09-12T07:00:00.000Z'],
		['UTC', '2026-09-05T00:00:00.000Z', '2026-09-12T00:00:00.000Z'],
	])('thisWeek in %s becomes explicit Saturday bounds', (timeZone, expectedFrom, expectedTo) => {
		const [from, to] = resolveFaDateRange(DATE_RANGES.thisWeek[0], DATE_RANGES.thisWeek[1], FIXED_NOW, timeZone)
		expect(from).toBe(expectedFrom)
		expect(to).toBe(expectedTo)
		expect(from).toMatch(GREGORIAN_ISO)
		expect(`${from}?to=${to}`).not.toMatch(JALALI_DIGITS)
	})

	it.each([
		['Asia/Tehran'],
		['America/Los_Angeles'],
		['UTC'],
	])('day presets keep datemath verbatim in %s', (timeZone) => {
		expect(resolveFaDateRange('now/d', 'now/d+1d', FIXED_NOW, timeZone)).toEqual(['now/d', 'now/d+1d'])
		expect(resolveFaDateRange('now', 'now+7d', FIXED_NOW, timeZone)).toEqual(['now', 'now+7d'])
		expect(resolveFaDateRange('now', 'now+30d', FIXED_NOW, timeZone)).toEqual(['now', 'now+30d'])
	})

	it('filter wire uses quoted Gregorian ISO (no Jalali digits)', () => {
		const [from, to] = resolveFaDateRange('now/w', 'now/w+1w', FIXED_NOW, 'Asia/Tehran')
		const filter = `done = false && due_date < '${to}' && due_date > '${from}'`
		expect(filter).not.toMatch(JALALI_DIGITS)
		expect(filter).toContain('2026-09-04T20:30:00.000Z')
		expect(filter).toContain('2026-09-11T20:30:00.000Z')
	})

	it('single overdue bound resolves like the range end', () => {
		const end = resolveFaDateValue('now/w+1w', FIXED_NOW, 'Asia/Tehran')
		const [, rangeEnd] = resolveFaDateRange('now/w', 'now/w+1w', FIXED_NOW, 'Asia/Tehran')
		expect(end).toBe(rangeEnd)
	})

	it('en/de path stays byte-identical (no conversion)', () => {
		expect(DATE_RANGES.thisWeek).toEqual(['now/w', 'now/w+1w'])
		expect(DATE_RANGES.today).toEqual(['now/d', 'now/d+1d'])
	})
})
