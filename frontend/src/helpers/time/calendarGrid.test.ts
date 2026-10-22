import {describe, expect, it} from 'vitest'
import {buildMonthGrid, weekdayOrder} from './calendarGrid'

describe('buildMonthGrid', () => {
	it('starts the grid on the configured week start', () => {
		// November 2022 starts on a Tuesday
		const monday = buildMonthGrid(2022, 10, 1)
		expect(monday[0].date.getDay()).toBe(1)
		expect(monday[0].date.getDate()).toBe(31)

		const sunday = buildMonthGrid(2022, 10, 0)
		expect(sunday[0].date.getDay()).toBe(0)
		expect(sunday[0].date.getDate()).toBe(30)
	})

	it('always yields 42 cells and marks out-of-month days', () => {
		const grid = buildMonthGrid(2024, 1, 1)
		expect(grid).toHaveLength(42)
		expect(grid.filter(c => c.inMonth)).toHaveLength(29)
	})
})

describe('weekdayOrder', () => {
	it('rotates so the week start comes first', () => {
		expect(weekdayOrder(1)).toEqual([1, 2, 3, 4, 5, 6, 0])
		expect(weekdayOrder(0)).toEqual([0, 1, 2, 3, 4, 5, 6])
	})
})
