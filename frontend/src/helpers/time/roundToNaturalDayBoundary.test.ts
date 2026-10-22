import {describe, it, expect} from 'vitest'

import {getNaturalDayCount, roundToNaturalDayBoundary} from './roundToNaturalDayBoundary'
import {MILLISECONDS_A_DAY} from '@/constants/date'

describe('roundToNaturalDayBoundary', () => {
	it('rounds start dates to start of day regardless of time', () => {
		const morning = new Date('2024-01-01T08:00:00')
		const afternoon = new Date('2024-01-01T13:00:00')
		expect(roundToNaturalDayBoundary(morning, true)).toEqual(new Date('2024-01-01T00:00:00.000'))
		expect(roundToNaturalDayBoundary(afternoon, true)).toEqual(new Date('2024-01-01T00:00:00.000'))
	})

	it('rounds end dates to natural boundaries', () => {
		const morning = new Date('2024-01-01T08:00:00')
		const afternoon = new Date('2024-01-01T13:00:00')
		expect(roundToNaturalDayBoundary(morning)).toEqual(new Date('2024-01-01T00:00:00.000'))
		expect(roundToNaturalDayBoundary(afternoon)).toEqual(new Date('2024-01-01T23:59:59.999'))
	})

	it('counts inclusive days for same-day evening end', () => {
		const start = new Date('2024-01-01T15:00:00')
		const end = new Date('2024-01-01T18:00:00')
		const diff = Math.ceil(
			(roundToNaturalDayBoundary(end).getTime() - roundToNaturalDayBoundary(start, true).getTime()) /
				MILLISECONDS_A_DAY,
		)
		expect(diff).toBe(1)
	})

	it('gives same-day tasks ending before noon a width of one day', () => {
		const start = new Date('2024-01-01T08:00:00')
		const end = new Date('2024-01-01T10:00:00')
		expect(getNaturalDayCount(start, end)).toBe(1)
	})

	it('gives tasks with identical start and end at midnight a width of one day', () => {
		const date = new Date('2024-01-01T00:00:00')
		expect(getNaturalDayCount(date, date)).toBe(1)
	})

	it('still counts multi-day spans', () => {
		const start = new Date('2024-01-01T08:00:00')
		expect(getNaturalDayCount(start, new Date('2024-01-03T18:00:00'))).toBe(3)
		expect(getNaturalDayCount(start, new Date('2024-01-03T10:00:00'))).toBe(2)
	})
})
