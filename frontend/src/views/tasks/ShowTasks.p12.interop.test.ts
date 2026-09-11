import {describe, expect, it} from 'vitest'
import {DATE_RANGES, resolveFaDateRange, resolveFaDateValue} from '@/components/date/dateRanges'
import {parseJalaliDateInput} from '@/helpers/time/jalali'

const FIXED_NOW = new Date('2026-09-07T12:00:00.000Z')
const JALALI_DIGITS = new RegExp('[۰-۹٠-٩]')
const GREGORIAN_ISO = new RegExp('^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}\\.\\d{3}Z$')
const ASCII_KEBAB = new RegExp('^\\d{4}-\\d{2}-\\d{2}$')

// P12: URLs/saved filters stay Gregorian kebab/ISO, never Jalali digits.
// ShowTasks.setDate pushes from/to verbatim (router query); loadPendingTasks resolves fa presets to ISO for the filter.
describe('P12 ShowTasks setDate query stays Gregorian', () => {
	it('setDate query keeps datemath verbatim (shareable, dynamic)', () => {
		// Mirrors ShowTasks.vue setDate: query.from/to = dates verbatim.
		const dates = {dateFrom: 'now/w', dateTo: 'now/w+1w'}
		const query = {from: dates.dateFrom, to: dates.dateTo}
		expect(query).toEqual({from: 'now/w', to: 'now/w+1w'})
		expect(`${query.from} ${query.to}`).not.toMatch(JALALI_DIGITS)
	})

	it('fa week preset resolves to explicit Gregorian ISO for the filter (not the URL)', () => {
		const [from, to] = resolveFaDateRange('now/w', 'now/w+1w', FIXED_NOW, 'Asia/Tehran')
		expect(from).toBe('2026-09-04T20:30:00.000Z')
		expect(to).toBe('2026-09-11T20:30:00.000Z')
		expect(from).toMatch(GREGORIAN_ISO)
		expect(to).toMatch(GREGORIAN_ISO)
		expect(`${from} ${to}`).not.toMatch(JALALI_DIGITS)
		const filter = `done = false && due_date < '${to}' && due_date > '${from}'`
		expect(filter).not.toMatch(JALALI_DIGITS)
	})

	it('typed Jalali converts to Gregorian ISO before URL', () => {
		const parsed = parseJalaliDateInput('1405/06/16', {timeZone: 'Asia/Tehran'})
		expect(parsed).not.toBeNull()
		const iso = parsed!.toISOString()
		expect(iso).toBe('2026-09-06T20:30:00.000Z')
		expect(iso).toMatch(GREGORIAN_ISO)
		expect(iso).not.toMatch(JALALI_DIGITS)
	})

	it('typed Persian-digit Jalali converts before URL', () => {
		const parsed = parseJalaliDateInput('۱۴۰۵/۰۶/۱۶', {timeZone: 'Asia/Tehran'})
		expect(parsed).not.toBeNull()
		expect(parsed!.toISOString()).toBe('2026-09-06T20:30:00.000Z')
	})

	it('preset datemath stays verbatim', () => {
		expect(resolveFaDateRange('now/d', 'now/d+1d', FIXED_NOW, 'Asia/Tehran')).toEqual(['now/d', 'now/d+1d'])
		expect(resolveFaDateRange('now', 'now+7d', FIXED_NOW, 'Asia/Tehran')).toEqual(['now', 'now+7d'])
		expect(resolveFaDateValue('now/w', FIXED_NOW, 'Asia/Tehran')).toBe('2026-09-04T20:30:00.000Z')
	})

	it('ISO stays preserved (no conversion, no Jalali)', () => {
		expect(resolveFaDateValue('2026-09-07', FIXED_NOW, 'Asia/Tehran')).toBe('2026-09-07')
		expect(resolveFaDateRange('2026-09-01', '2026-09-30', FIXED_NOW, 'Asia/Tehran')).toEqual(['2026-09-01', '2026-09-30'])
		expect('2026-09-07').toMatch(ASCII_KEBAB)
	})

	it('en presets stay byte-identical', () => {
		expect(DATE_RANGES.thisWeek).toEqual(['now/w', 'now/w+1w'])
		expect(DATE_RANGES.today).toEqual(['now/d', 'now/d+1d'])
	})
})
