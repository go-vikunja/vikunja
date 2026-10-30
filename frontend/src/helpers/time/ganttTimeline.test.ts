import {describe, expect, it} from 'vitest'

import {
	GANTT_ZOOM_MAX,
	GANTT_ZOOM_MIN,
	groupTimelineDates,
	isoWeekNumber,
	parseGanttZoom,
	stepGanttZoom,
	timelineGroupOffsets,
	timelineUnitsForDayWidth,
} from './ganttTimeline'

function datesBetween(from: string, to: string): Date[] {
	const dates: Date[] = []
	const current = new Date(from)
	const end = new Date(to)

	while (current <= end) {
		dates.push(new Date(current))
		current.setDate(current.getDate() + 1)
	}

	return dates
}

describe('ganttTimeline', () => {
	it('parses a zoom from the route and rejects garbage', () => {
		expect(parseGanttZoom('8')).toBe(8)
		expect(parseGanttZoom('12.5')).toBe(12.5)
		expect(parseGanttZoom('1000')).toBe(GANTT_ZOOM_MAX)
		expect(parseGanttZoom('0')).toBeUndefined()
		expect(parseGanttZoom('week')).toBeUndefined()
		expect(parseGanttZoom(undefined)).toBeUndefined()
	})

	it('steps the zoom geometrically within the bounds', () => {
		expect(stepGanttZoom(8, 1)).toBe(12)
		expect(stepGanttZoom(12, -1)).toBe(8)
		expect(stepGanttZoom(GANTT_ZOOM_MIN, -1)).toBe(GANTT_ZOOM_MIN)
		expect(stepGanttZoom(GANTT_ZOOM_MAX, 1)).toBe(GANTT_ZOOM_MAX)
	})

	it('picks timeline units from the day width', () => {
		expect(timelineUnitsForDayWidth(30)).toEqual({upper: 'month', lower: 'day', showWeekday: true})
		expect(timelineUnitsForDayWidth(25)).toEqual({upper: 'month', lower: 'day', showWeekday: false})
		expect(timelineUnitsForDayWidth(8)).toEqual({upper: 'month', lower: 'week', showWeekday: false})
		expect(timelineUnitsForDayWidth(3)).toEqual({upper: 'year', lower: 'month', showWeekday: false})
	})

	it('groups days into iso weeks, keeping partial weeks at the edges', () => {
		// 2026-01-01 is a Thursday
		const groups = groupTimelineDates(datesBetween('2026-01-01T00:00', '2026-01-14T00:00'), 'week')

		expect(groups.map(g => g.dayCount)).toEqual([4, 7, 3])
		expect(groups.map(g => isoWeekNumber(g.start))).toEqual([1, 2, 3])
	})

	it('groups days into months and years', () => {
		const dates = datesBetween('2025-12-15T00:00', '2026-02-03T00:00')

		expect(groupTimelineDates(dates, 'month').map(g => [g.key, g.dayCount])).toEqual([
			['2025-12', 17],
			['2026-01', 31],
			['2026-02', 3],
		])
		expect(groupTimelineDates(dates, 'year').map(g => g.dayCount)).toEqual([17, 34])
	})

	it('computes pixel offsets from the cumulative day count', () => {
		const groups = groupTimelineDates(datesBetween('2026-01-01T00:00', '2026-01-14T00:00'), 'week')

		expect(timelineGroupOffsets(groups, 10)).toEqual([0, 40, 110])
	})
})
