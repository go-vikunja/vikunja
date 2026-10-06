import {MILLISECONDS_A_DAY} from '@/constants/date'

export function roundToNaturalDayBoundary(date: Date, isStart = false): Date {
	const d = new Date(date)
	if (isStart || d.getHours() < 12) {
		d.setHours(0, 0, 0, 0)
	} else {
		d.setHours(23, 59, 59, 999)
	}
	return d
}

/**
 * Number of whole days a Gantt bar spans. A bar always covers at least one day,
 * otherwise a task that starts and ends before noon on the same day would be
 * rounded to zero width and not be drawn at all.
 */
export function getNaturalDayCount(start: Date, end: Date): number {
	const diff = Math.ceil(
		(roundToNaturalDayBoundary(end).getTime() - roundToNaturalDayBoundary(start, true).getTime()) /
		MILLISECONDS_A_DAY,
	)
	return Math.max(1, diff)
}
