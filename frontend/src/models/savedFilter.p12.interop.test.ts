import {describe, expect, it} from 'vitest'
import {transformFilterStringForApi, transformFilterStringFromApi} from '@/helpers/filters'
import {parseJalaliDateInput} from '@/helpers/time/jalali'
import SavedFilterModel from '@/models/savedFilter'

const JALALI_DIGITS = new RegExp('[۰-۹٠-٩]')

function nullResolver() {
	return null
}

// P12: saved-filter serialization stays Gregorian kebab/ISO/datemath, never Jalali.
describe('P12 saved-filter serialization stays Gregorian', () => {
	it('typed Jalali converts to ISO before serialization', () => {
		const parsed = parseJalaliDateInput('1405/06/16', {timeZone: 'Asia/Tehran'})
		expect(parsed).not.toBeNull()
		const api = transformFilterStringForApi(`dueDate > ${parsed!.toISOString()}`, nullResolver, nullResolver)
		expect(api).toBe('due_date > 2026-09-06T20:30:00.000Z')
		expect(api).not.toMatch(JALALI_DIGITS)
	})

	it('preset datemath preserved through api transform', () => {
		expect(transformFilterStringForApi('dueDate > now/w', nullResolver, nullResolver)).toBe('due_date > now/w')
		expect(transformFilterStringForApi('dueDate = now/d', nullResolver, nullResolver)).toBe('due_date = now/d')
		expect(transformFilterStringForApi('dueDate > now+7d', nullResolver, nullResolver)).toBe('due_date > now+7d')
	})

	it('ISO preserved through api and back', () => {
		const api = transformFilterStringForApi('dueDate > 2026-09-07', nullResolver, nullResolver)
		expect(api).toBe('due_date > 2026-09-07')
		expect(api).not.toMatch(JALALI_DIGITS)
		const back = transformFilterStringFromApi(api, () => null, () => null)
		expect(back).toBe('dueDate > 2026-09-07')
		expect(back).not.toMatch(JALALI_DIGITS)
	})

	it('SavedFilterModel stores Gregorian filter without Jalali', () => {
		const model = new SavedFilterModel({filters: {filter: 'due_date > 2026-09-07T12:00:00.000Z'}} as never)
		expect(model.filters.filter).toContain('2026-09-07T12:00:00.000Z')
		expect(model.filters.filter).not.toMatch(JALALI_DIGITS)
		const dynamic = new SavedFilterModel({filters: {filter: 'due_date > now/w'}} as never)
		expect(dynamic.filters.filter).toBe('due_date > now/w')
	})
})
