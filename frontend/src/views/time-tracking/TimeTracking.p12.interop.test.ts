import {describe, expect, it} from 'vitest'
import {formatGregorianKebab, resolveFaDateValueKebab} from '@/components/date/dateRanges'

const FIXED_NOW = new Date('2026-09-07T12:00:00.000Z')
const JALALI_DIGITS = new RegExp('[۰-۹٠-٩]')
const GREGORIAN_KEBAB = new RegExp('^\\d{4}-\\d{2}-\\d{2}$')

// Mirrors TimeTracking.vue dateValue: datemath verbatim, Date to kebab.
function dateValue(value: Date | string): string {
	if (typeof value === 'string') {
		return value
	}
	const year = value.getFullYear()
	const month = String(value.getMonth() + 1).padStart(2, '0')
	const day = String(value.getDate()).padStart(2, '0')
	return `${year}-${month}-${day}`
}

// P12: TimeTracking filterQuery stays Gregorian kebab/datemath, never Jalali.
describe('P12 TimeTracking filterQuery stays Gregorian', () => {
	it('filterQuery keeps datemath verbatim', () => {
		expect(dateValue('now/w')).toBe('now/w')
		expect(dateValue('now/w+1w')).toBe('now/w+1w')
		expect(dateValue('now/d')).toBe('now/d')
		expect(dateValue('now+7d')).toBe('now+7d')
	})

	it('fa week preset resolves to kebab for the filter (ASCII only)', () => {
		const from = resolveFaDateValueKebab('now/w', FIXED_NOW, 'Asia/Tehran')
		const to = resolveFaDateValueKebab('now/w+1w', FIXED_NOW, 'Asia/Tehran')
		expect(from).toBe('2026-09-05')
		expect(to).toBe('2026-09-12')
		expect(from).toMatch(GREGORIAN_KEBAB)
		expect(to).toMatch(GREGORIAN_KEBAB)
		expect(`${from} ${to}`).not.toMatch(JALALI_DIGITS)
		const filter = `start_time > ${from} && start_time < ${to}`
		expect(filter).not.toMatch(JALALI_DIGITS)
		expect(filter).not.toContain(':')
	})

	it('custom Date becomes kebab (no colons, no Jalali)', () => {
		const kebab = dateValue(new Date(2026, 8, 7))
		expect(kebab).toBe('2026-09-07')
		expect(kebab).toMatch(GREGORIAN_KEBAB)
		expect(kebab).not.toMatch(JALALI_DIGITS)
	})

	it('wall-clock kebab is ASCII Gregorian', () => {
		const kebab = formatGregorianKebab(new Date('2026-09-06T20:30:00.000Z'), 'Asia/Tehran')
		expect(kebab).toBe('2026-09-07')
		expect(kebab).toMatch(GREGORIAN_KEBAB)
		expect(kebab).not.toMatch(JALALI_DIGITS)
	})

	it('day presets stay verbatim in kebab path', () => {
		for (const tz of ['Asia/Tehran', 'America/Los_Angeles', 'UTC']) {
			expect(resolveFaDateValueKebab('now/d', FIXED_NOW, tz)).toBe('now/d')
			expect(resolveFaDateValueKebab('now', FIXED_NOW, tz)).toBe('now')
			expect(resolveFaDateValueKebab('now+7d', FIXED_NOW, tz)).toBe('now+7d')
		}
	})

	it('ISO kebab stays preserved', () => {
		expect(dateValue('2026-09-07')).toBe('2026-09-07')
		expect(resolveFaDateValueKebab('2026-09-07', FIXED_NOW, 'Asia/Tehran')).toBe('2026-09-07')
	})
})
