import {describe, expect, it} from 'vitest'

import {
	buildTimelineTiers,
	gridLineXs,
	isGanttZoom,
	resolveDayWidth,
	startOfWeek,
	todayX,
	weekendBands,
	ZOOM_DAY_WIDTH,
} from './ganttZoom'

// 2026-11-02 is a Monday.
function days(count: number, from = new Date(2026, 10, 2)): Date[] {
	return Array.from({length: count}, (_, i) => new Date(from.getFullYear(), from.getMonth(), from.getDate() + i))
}

describe('resolveDayWidth', () => {
	it('fit stretches to the container but never below the minimum', () => {
		expect(resolveDayWidth('fit', 55)).toBe(55)
		expect(resolveDayWidth('fit', 4)).toBe(30)
	})

	it('uses a fixed width for every other zoom and ignores the container', () => {
		expect(resolveDayWidth('day', 999)).toBe(ZOOM_DAY_WIDTH.day)
		expect(resolveDayWidth('week', 1)).toBe(ZOOM_DAY_WIDTH.week)
		expect(resolveDayWidth('month', 1)).toBe(ZOOM_DAY_WIDTH.month)
		expect(resolveDayWidth('quarter', 1)).toBe(ZOOM_DAY_WIDTH.quarter)
	})

	it('widths shrink as the zoom gets coarser', () => {
		expect(ZOOM_DAY_WIDTH.day).toBeGreaterThan(ZOOM_DAY_WIDTH.week)
		expect(ZOOM_DAY_WIDTH.week).toBeGreaterThan(ZOOM_DAY_WIDTH.month)
		expect(ZOOM_DAY_WIDTH.month).toBeGreaterThan(ZOOM_DAY_WIDTH.quarter)
	})
})

describe('isGanttZoom', () => {
	it('accepts the known values only', () => {
		expect(isGanttZoom('week')).toBe(true)
		expect(isGanttZoom('fit')).toBe(true)
		expect(isGanttZoom('year')).toBe(false)
		expect(isGanttZoom(undefined)).toBe(false)
	})
})

describe('startOfWeek', () => {
	it('returns the Monday, also for a Sunday', () => {
		expect(startOfWeek(new Date(2026, 10, 4)).getDate()).toBe(2)
		expect(startOfWeek(new Date(2026, 10, 8)).getDate()).toBe(2)
		expect(startOfWeek(new Date(2026, 10, 2)).getDate()).toBe(2)
		expect(startOfWeek(new Date(2026, 10, 9)).getDate()).toBe(9)
	})
})

describe('buildTimelineTiers', () => {
	it('day and fit have months on top and no lower tier', () => {
		const tiers = buildTimelineTiers(days(40), 40, 'day')
		expect(tiers.upper.map(c => c.unit)).toEqual(['month', 'month'])
		expect(tiers.lower).toEqual([])
		expect(buildTimelineTiers(days(40), 40, 'fit').lower).toEqual([])
	})

	it('week zoom: months over weeks, widths add up to the total', () => {
		const dates = days(28) // 4 whole weeks starting on a Monday
		const tiers = buildTimelineTiers(dates, 14, 'week')
		expect(tiers.lower).toHaveLength(4)
		expect(tiers.lower.every(c => c.width === 7 * 14)).toBe(true)
		expect(tiers.lower.map(c => c.x)).toEqual([0, 98, 196, 294])
		expect(tiers.upper.reduce((sum, c) => sum + c.width, 0)).toBe(28 * 14)
	})

	it('a partial week at the start is a narrower cell', () => {
		const tiers = buildTimelineTiers(days(10, new Date(2026, 10, 4)), 10, 'week') // Wed 4th
		expect(tiers.lower[0].width).toBe(5 * 10) // Wed..Sun
		expect(tiers.lower[1].width).toBe(5 * 10) // Mon 9th..Fri 13th
	})

	it('month zoom: years over months; quarter zoom: years over quarters', () => {
		const dates = days(200, new Date(2026, 9, 1))
		const month = buildTimelineTiers(dates, 5, 'month')
		expect(month.upper.map(c => c.label)).toEqual(['2026', '2027'])
		expect(month.lower[0].width).toBe(31 * 5)
		const quarter = buildTimelineTiers(dates, 2, 'quarter')
		expect(quarter.lower.map(c => c.key)).toEqual(['2026-Q3', '2027-Q0', '2027-Q1'])
	})

	it('labels come from the formatter', () => {
		const tiers = buildTimelineTiers(days(3), 10, 'day', (_d, unit) => `L:${unit}`)
		expect(tiers.upper[0].label).toBe('L:month')
	})

	it('an empty range is empty', () => {
		expect(buildTimelineTiers([], 10, 'week')).toEqual({upper: [], lower: []})
	})
})

describe('gridLineXs', () => {
	it('one line per day for fit and day', () => {
		expect(gridLineXs(days(3), 40, 'day')).toEqual([0, 40, 80])
		expect(gridLineXs(days(3), 30, 'fit')).toEqual([0, 30, 60])
	})

	it('lines follow the lower tier at coarser zooms', () => {
		expect(gridLineXs(days(14), 14, 'week')).toEqual([0, 98])
	})
})

describe('weekendBands', () => {
	it('merges Saturday and Sunday into one band', () => {
		const bands = weekendBands(days(14), 40, 'day')
		expect(bands).toEqual([{x: 5 * 40, width: 80}, {x: 12 * 40, width: 80}])
	})

	it('a range that starts on a Sunday begins with a one-day band', () => {
		const bands = weekendBands(days(3, new Date(2026, 10, 1)), 10, 'day')
		expect(bands[0]).toEqual({x: 0, width: 10})
	})

	it('no bands at month and quarter zoom', () => {
		expect(weekendBands(days(30), 5, 'month')).toEqual([])
		expect(weekendBands(days(30), 2, 'quarter')).toEqual([])
	})
})

describe('todayX', () => {
	it('places now inside its day', () => {
		const x = todayX(days(10), 40, new Date(2026, 10, 4, 12, 0))
		expect(x).toBeCloseTo(2 * 40 + 20, 5)
	})

	it('null when now is outside the shown days', () => {
		expect(todayX(days(10), 40, new Date(2026, 9, 30))).toBeNull()
		expect(todayX(days(10), 40, new Date(2026, 10, 12))).toBeNull()
		expect(todayX([], 40, new Date(2026, 10, 4))).toBeNull()
	})

	it('the first and last shown day are inside', () => {
		expect(todayX(days(10), 40, new Date(2026, 10, 2, 0, 0))).toBeCloseTo(0, 5)
		expect(todayX(days(10), 40, new Date(2026, 10, 11, 18, 0))).toBeGreaterThan(9 * 40)
	})
})
