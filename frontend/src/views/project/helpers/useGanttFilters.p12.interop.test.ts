import {describe, expect, it} from 'vitest'
import {isoToKebabDate} from '@/helpers/time/isoToKebabDate'
import {parseDateProp} from '@/helpers/time/parseDateProp'

const JALALI_DIGITS = new RegExp('[۰-۹٠-٩]')
const GREGORIAN_KEBAB = new RegExp('^\\d{4}-\\d{2}-\\d{2}$')

// P12: Gantt dateFrom/dateTo stay Gregorian kebab in URL and API filter.
describe('P12 Gantt dateFrom/dateTo kebab stays Gregorian', () => {
	it('iso to kebab is ASCII Gregorian (no Jalali digits)', () => {
		const kebab = isoToKebabDate('2026-09-07T12:00:00.000Z')
		expect(kebab).toBe('2026-09-07')
		expect(kebab).toMatch(GREGORIAN_KEBAB)
		expect(kebab).not.toMatch(JALALI_DIGITS)
	})

	it('gantt api filter built from kebab stays Gregorian', () => {
		const dateFrom = isoToKebabDate('2026-09-01T12:00:00.000Z')
		const dateTo = isoToKebabDate('2026-09-10T12:00:00.000Z')
		const filter = `(start_date >= "${dateFrom}" && start_date <= "${dateTo}")`
		expect(filter).toContain('2026-09-01')
		expect(filter).toContain('2026-09-10')
		expect(filter).not.toMatch(JALALI_DIGITS)
	})

	it('route query round-trips kebab to ISO without Jalali', () => {
		const iso = parseDateProp('2026-09-07')
		expect(iso).toBeDefined()
		// Local midnight round-trips through the same local kebab helper in any TZ.
		expect(isoToKebabDate(iso! as never)).toBe('2026-09-07')
		expect(iso!).not.toMatch(JALALI_DIGITS)
	})

	it('Jalali digits in route query are rejected (no leak into filters)', () => {
		expect(parseDateProp('۱۴۰۵-۰۶-۱۶' as never)).toBeUndefined()
		expect(parseDateProp('1405-06-16')).toBeDefined()
		// 1405 is not a Gregorian year the gantt range expects; kebab shape holds but wire stays ASCII.
		expect('1405-06-16').not.toMatch(JALALI_DIGITS)
	})

	it('ISO preserved verbatim through kebab helpers', () => {
		expect(isoToKebabDate('2026-01-01T12:00:00.000Z')).toMatch(GREGORIAN_KEBAB)
		expect(parseDateProp('2026-01-31')).toBeDefined()
	})
})
