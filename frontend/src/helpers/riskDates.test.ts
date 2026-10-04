import {describe, expect, it} from 'vitest'

import {dayInputToIso, dueBeforeIdentified, isoToDayInput} from './riskDates'

describe('dayInputToIso', () => {
	it('sends a day as local midnight', () => {
		const iso = dayInputToIso('2026-10-04')
		expect(iso).not.toBeNull()
		const date = new Date(iso as string)
		expect([date.getFullYear(), date.getMonth(), date.getDate()]).toEqual([2026, 9, 4])
		expect([date.getHours(), date.getMinutes(), date.getSeconds()]).toEqual([0, 0, 0])
	})

	it('sends a due day as the end of that day', () => {
		const date = new Date(dayInputToIso('2026-10-04', true) as string)
		expect([date.getFullYear(), date.getMonth(), date.getDate()]).toEqual([2026, 9, 4])
		expect([date.getHours(), date.getMinutes(), date.getSeconds()]).toEqual([23, 59, 59])
	})

	it.each(['', 'soon', '2026-13-01', '2026-02-30', '4/10/2026'])('gives null for %j', value => {
		expect(dayInputToIso(value)).toBeNull()
	})
})

describe('isoToDayInput', () => {
	it('round trips both kinds of day', () => {
		expect(isoToDayInput(dayInputToIso('2026-10-04'))).toBe('2026-10-04')
		expect(isoToDayInput(dayInputToIso('2026-10-04', true))).toBe('2026-10-04')
	})

	it('is empty without a date', () => {
		expect(isoToDayInput(null)).toBe('')
		expect(isoToDayInput(undefined)).toBe('')
		expect(isoToDayInput('')).toBe('')
		expect(isoToDayInput('nonsense')).toBe('')
	})
})

describe('dueBeforeIdentified', () => {
	it('is true only when both days are set and due is earlier', () => {
		expect(dueBeforeIdentified('2026-10-05', '2026-10-04')).toBe(true)
		expect(dueBeforeIdentified('2026-10-04', '2026-10-04')).toBe(false)
		expect(dueBeforeIdentified('2026-10-04', '2026-10-05')).toBe(false)
		expect(dueBeforeIdentified('', '2026-10-04')).toBe(false)
		expect(dueBeforeIdentified('2026-10-05', '')).toBe(false)
	})
})
