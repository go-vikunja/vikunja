import {describe, expect, it, beforeEach, afterEach} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'
import {i18n as globalI18n} from '@/i18n'
import {useAuthStore} from '@/stores/auth'
import {formatDate} from '@/helpers/time/formatDate'
import {DATE_RANGES, formatGregorianKebab, resolveFaDateValueKebab} from '@/components/date/dateRanges'

const FIXED_NOW = new Date('2026-09-07T12:00:00.000Z')
const JALALI_DIGITS = new RegExp('[۰-۹٠-٩]')
const GREGORIAN_KEBAB = new RegExp('^\\d{4}-\\d{2}-\\d{2}$')

function dateValue(value: Date | string): string {
	if (typeof value === 'string') {
		return value
	}
	const year = value.getFullYear()
	const month = String(value.getMonth() + 1).padStart(2, '0')
	const day = String(value.getDate()).padStart(2, '0')
	return `${year}-${month}-${day}`
}

describe('TimeTracking fa range wiring', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	afterEach(() => {
		globalI18n.global.locale.value = 'en' as never
	})

	it.each([
		['Asia/Tehran', '2026-09-05'],
		['America/Los_Angeles', '2026-09-05'],
		['UTC', '2026-09-05'],
	])('week preset becomes kebab wall-clock in %s (no colons, no Jalali digits)', (timeZone, expectedStart) => {
		const start = resolveFaDateValueKebab('now/w', FIXED_NOW, timeZone)
		const end = resolveFaDateValueKebab('now/w+1w', FIXED_NOW, timeZone)
		expect(start).toBe(expectedStart)
		expect(end).toMatch(GREGORIAN_KEBAB)
		expect(start).not.toContain(':')
		expect(end).not.toContain(':')
		expect(start).not.toMatch(JALALI_DIGITS)
		expect(`${start} to ${end}`).not.toMatch(JALALI_DIGITS)
	})

	it('day presets keep datemath verbatim', () => {
		for (const timeZone of ['Asia/Tehran', 'America/Los_Angeles', 'UTC']) {
			expect(resolveFaDateValueKebab('now/d', FIXED_NOW, timeZone)).toBe('now/d')
			expect(resolveFaDateValueKebab('now/d+1d', FIXED_NOW, timeZone)).toBe('now/d+1d')
			expect(resolveFaDateValueKebab('now', FIXED_NOW, timeZone)).toBe('now')
			expect(resolveFaDateValueKebab('now+7d', FIXED_NOW, timeZone)).toBe('now+7d')
		}
	})

	it('range label uses central formatDate in fa mode', () => {
		globalI18n.global.locale.value = 'fa-IR' as never
		useAuthStore().setUserSettings({timezone: 'Asia/Tehran'} as never)
		const kebab = formatGregorianKebab(new Date('2026-09-06T20:30:00.000Z'), 'Asia/Tehran')
		expect(kebab).toBe('2026-09-07')
		const label = formatDate('2026-09-06T20:30:00.000Z', 'LL')
		expect(label).not.toBe('')
		expect(label).not.toBe('2026-09-07')
		expect(label).not.toMatch(new RegExp('now/d'))
	})

	it('en label stays byte-identical (raw kebab/datemath)', () => {
		globalI18n.global.locale.value = 'en' as never
		expect(dateValue('now/d')).toBe('now/d')
		expect(dateValue(new Date(2026, 8, 7))).toBe('2026-09-07')
		expect(DATE_RANGES.thisWeek).toEqual(['now/w', 'now/w+1w'])
	})

	it('filter query wire stays Gregorian kebab/datemath (no Jalali digits)', () => {
		const from = resolveFaDateValueKebab('now/w', FIXED_NOW, 'Asia/Tehran')
		const to = resolveFaDateValueKebab('now/w+1w', FIXED_NOW, 'Asia/Tehran')
		const filter = `start_time > ${from} && start_time < ${to}`
		expect(filter).not.toMatch(JALALI_DIGITS)
		expect(filter).toContain('2026-09-05')
	})
})
