import {describe, expect, it} from 'vitest'

import {formatTimeRange} from './formatTimeRange'
import {TIME_FORMAT} from '@/constants/timeFormat'

const now = new Date(2026, 9, 20, 12)

describe('formatTimeRange', () => {
	it('omits the year for dates in the current year', () => {
		expect(formatTimeRange(new Date(2026, 5, 7, 9), new Date(2026, 5, 7, 10), TIME_FORMAT.HOURS_24, now))
			.toBe('Jun 7 09:00 – 10:00')
	})

	it('shows the year for dates in another year', () => {
		expect(formatTimeRange(new Date(2025, 5, 7, 9), new Date(2025, 5, 7, 10), TIME_FORMAT.HOURS_24, now))
			.toBe('Jun 7, 2025 09:00 – 10:00')
	})

	it('shows both dates when the range spans midnight', () => {
		expect(formatTimeRange(new Date(2026, 5, 7, 23), new Date(2026, 5, 8, 1), TIME_FORMAT.HOURS_24, now))
			.toBe('Jun 7 23:00 – Jun 8 01:00')
	})

	it('shows both years when the range spans new year', () => {
		expect(formatTimeRange(new Date(2025, 11, 31, 23), new Date(2026, 0, 1, 1), TIME_FORMAT.HOURS_24, now))
			.toBe('Dec 31, 2025 23:00 – Jan 1 01:00')
	})

	it('leaves the end open when there is none', () => {
		expect(formatTimeRange(new Date(2026, 5, 7, 9), null, TIME_FORMAT.HOURS_24, now))
			.toBe('Jun 7 09:00 – …')
	})

	it('uses the 12 hour format', () => {
		expect(formatTimeRange(new Date(2026, 5, 7, 9), new Date(2026, 5, 7, 14), TIME_FORMAT.HOURS_12, now))
			.toBe('Jun 7 09:00 AM – 02:00 PM')
	})
})
