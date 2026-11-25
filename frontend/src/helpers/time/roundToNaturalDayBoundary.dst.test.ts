process.env.TZ = 'America/New_York'

import {describe, it, expect} from 'vitest'

import {getCalendarDayDiff, getNaturalDayCount} from './roundToNaturalDayBoundary'

describe('Gantt day math across DST changes', () => {
	it('runs in a timezone with DST', () => {
		expect(new Date('2026-10-28T12:00:00').getTimezoneOffset()).toBe(240)
		expect(new Date('2026-11-03T12:00:00').getTimezoneOffset()).toBe(300)
	})

	describe('getNaturalDayCount', () => {
		it('does not add a day to a span crossing the fall-back', () => {
			const start = new Date('2026-10-28T00:59:00-04:00')
			const end = new Date('2026-11-03T23:59:00-05:00')
			expect(getNaturalDayCount(start, end)).toBe(7)
		})

		it('gives a same-day bar after the fall-back a width of one day', () => {
			const start = new Date('2026-11-03T11:00:00-05:00')
			const end = new Date('2026-11-03T18:00:00-05:00')
			expect(getNaturalDayCount(start, end)).toBe(1)
		})

		it('counts the fall-back day itself as one day', () => {
			const start = new Date('2026-11-01T00:00:00-04:00')
			const end = new Date('2026-11-01T23:59:00-05:00')
			expect(getNaturalDayCount(start, end)).toBe(1)
		})

		it('does not drop a day from a span crossing the spring-forward', () => {
			const start = new Date('2026-03-05T09:00:00-05:00')
			const end = new Date('2026-03-10T18:00:00-04:00')
			expect(getNaturalDayCount(start, end)).toBe(6)
		})

		it('does not count the end day when the end is before noon across the fall-back', () => {
			const start = new Date('2026-10-30T09:00:00-04:00')
			const end = new Date('2026-11-02T10:00:00-05:00')
			expect(getNaturalDayCount(start, end)).toBe(3)
		})

		it('is unchanged for spans without a DST change', () => {
			const start = new Date('2026-10-23T23:59:59-04:00')
			const end = new Date('2026-10-30T23:59:59-04:00')
			expect(getNaturalDayCount(start, end)).toBe(8)
		})
	})

	describe('getCalendarDayDiff (bar x offset)', () => {
		const timelineStart = new Date('2026-10-15T00:00:00-04:00')

		it('places a bar after the fall-back in the right column', () => {
			expect(getCalendarDayDiff(timelineStart, new Date('2026-11-03T11:00:00-05:00'))).toBe(19)
		})

		it('places a bar before the fall-back in the right column', () => {
			expect(getCalendarDayDiff(timelineStart, new Date('2026-10-28T00:59:00-04:00'))).toBe(13)
		})

		it('places a bar after the spring-forward in the right column', () => {
			const springStart = new Date('2026-03-01T00:00:00-05:00')
			expect(getCalendarDayDiff(springStart, new Date('2026-03-10T08:00:00-04:00'))).toBe(9)
		})

		it('ignores the time of day', () => {
			expect(getCalendarDayDiff(
				new Date('2026-11-01T23:30:00-05:00'),
				new Date('2026-11-02T00:30:00-05:00'),
			)).toBe(1)
		})

		it('is negative for dates before the timeline start', () => {
			expect(getCalendarDayDiff(timelineStart, new Date('2026-10-13T15:00:00-04:00'))).toBe(-2)
		})
	})
})
