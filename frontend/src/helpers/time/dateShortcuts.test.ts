import {describe, expect, it} from 'vitest'
import {buildDateShortcuts} from './dateShortcuts'

const keys = (date: Date) => buildDateShortcuts(date).map(s => s.key)

describe('buildDateShortcuts', () => {
	it('drops shortcuts that land on the same day as an earlier one', () => {
		// Friday: tomorrow is the weekend, "later this week" is today
		expect(keys(new Date(2026, 8, 11, 10))).toEqual(['today', 'tomorrow', 'nextMonday', 'nextWeek', 'nextWeekend', 'endOfMonth', 'inAMonth'])
		// Monday: "next Monday" resolves to today
		expect(keys(new Date(2026, 8, 14, 10))).toEqual(['today', 'tomorrow', 'thisWeekend', 'laterThisWeek', 'nextWeek', 'nextWeekend', 'endOfMonth', 'inAMonth'])
	})

	it('keeps distinct days on a midweek day', () => {
		// Tuesday
		expect(keys(new Date(2026, 8, 15, 10))).toEqual(['today', 'tomorrow', 'nextMonday', 'thisWeekend', 'laterThisWeek', 'nextWeek', 'nextWeekend', 'endOfMonth', 'inAMonth'])
	})

	it('resolves the month based shortcuts', () => {
		const byKey = Object.fromEntries(buildDateShortcuts(new Date(2026, 8, 15, 10)).map(s => [s.key, s.date]))
		expect(byKey.nextWeekend).toEqual(new Date(2026, 8, 26))
		expect(byKey.endOfMonth).toEqual(new Date(2026, 8, 30))
		expect(byKey.inAMonth).toEqual(new Date(2026, 9, 15))
	})

	it('resolves next weekend to the very next Saturday on a Sunday', () => {
		const byKey = Object.fromEntries(buildDateShortcuts(new Date(2026, 8, 13, 10)).map(s => [s.key, s.date]))
		expect(byKey.nextWeekend).toEqual(new Date(2026, 8, 19))
	})

	it('clamps "in a month" to the last day of the target month', () => {
		const byKey = Object.fromEntries(buildDateShortcuts(new Date(2026, 0, 31, 10)).map(s => [s.key, s.date]))
		expect(byKey.inAMonth).toEqual(new Date(2026, 1, 28))
	})

	it('hides end of month on the last day of the month', () => {
		expect(keys(new Date(2026, 8, 30, 10))).not.toContain('endOfMonth')
	})

	it('hides today late in the evening', () => {
		expect(keys(new Date(2026, 8, 15, 22))).not.toContain('today')
	})

	it('resolves each shortcut to the start of its day', () => {
		const shortcuts = buildDateShortcuts(new Date(2026, 8, 15, 10, 30))
		for (const {date} of shortcuts) {
			expect(date.getHours()).toBe(0)
		}
	})
})
