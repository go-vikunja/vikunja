import {describe, expect, it} from 'vitest'

import {parseDateMath} from './parseDateMath'

// Wednesday
const now = new Date(2026, 9, 21, 14, 30, 15)

describe('parseDateMath', () => {
	it.each([
		['now', new Date(2026, 9, 21, 14, 30, 15)],
		['now/d', new Date(2026, 9, 21)],
		['now/d+1d', new Date(2026, 9, 22)],
		['now/w', new Date(2026, 9, 19)],
		['now/w-1w', new Date(2026, 9, 12)],
		['now/w+2w', new Date(2026, 10, 2)],
		['now/M', new Date(2026, 9, 1)],
		['now/M+1M', new Date(2026, 10, 1)],
		['now/y+1y', new Date(2027, 0, 1)],
		['now+7d', new Date(2026, 9, 28, 14, 30, 15)],
		['now-2h/h', new Date(2026, 9, 21, 12)],
		['2024-01-31T00:00:00||+1d', new Date(2024, 1, 1)],
	])('evaluates %s', (expression, expected) => {
		expect(parseDateMath(expression, now)).toEqual(expected)
	})

	it('rounds a Sunday back to the previous Monday', () => {
		expect(parseDateMath('now/w', new Date(2026, 9, 25, 10))).toEqual(new Date(2026, 9, 19))
	})

	it.each(['', '2026-10-21', 'now+', 'now/x', 'nope||+1d'])('returns null for %s', expression => {
		expect(parseDateMath(expression, now)).toBeNull()
	})
})
