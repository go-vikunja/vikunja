import {describe, expect, it} from 'vitest'
import {isDayBetween, isSameDay} from './dateMath'

describe('isSameDay / isDayBetween', () => {
	it('ignores the time part', () => {
		expect(isSameDay(new Date(2024, 0, 1, 8), new Date(2024, 0, 1, 22))).toBe(true)
		expect(isSameDay(new Date(2024, 0, 1), null)).toBe(false)
	})

	it('is exclusive on both ends', () => {
		const start = new Date(2024, 0, 1)
		const end = new Date(2024, 0, 5)
		expect(isDayBetween(new Date(2024, 0, 1, 12), start, end)).toBe(false)
		expect(isDayBetween(new Date(2024, 0, 3), start, end)).toBe(true)
		expect(isDayBetween(end, start, end)).toBe(false)
	})
})
