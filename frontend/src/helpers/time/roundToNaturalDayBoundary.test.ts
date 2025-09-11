import {describe, it, expect} from 'vitest'

import {roundToNaturalDayBoundary} from './roundToNaturalDayBoundary'
import {MILLISECONDS_A_DAY} from '@/constants/date'

describe('roundToNaturalDayBoundary', () => {
	it('rounds morning dates to start of day', () => {
		const d = new Date('2024-01-01T08:00:00')
		expect(roundToNaturalDayBoundary(d)).toEqual(new Date('2024-01-01T00:00:00.000'))
	})

	it('rounds afternoon dates to end of day', () => {
		const d = new Date('2024-01-01T13:00:00')
		expect(roundToNaturalDayBoundary(d)).toEqual(new Date('2024-01-01T23:59:59.999'))
	})

	it('counts inclusive days for same-day evening end', () => {
		const start = new Date('2024-01-01T08:00:00')
		const end = new Date('2024-01-01T18:00:00')
		const diff = Math.ceil(
			(roundToNaturalDayBoundary(end).getTime() - roundToNaturalDayBoundary(start).getTime()) /
			MILLISECONDS_A_DAY,
		)
		expect(diff).toBe(1)
	})
})
