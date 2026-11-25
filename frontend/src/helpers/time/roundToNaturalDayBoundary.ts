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
 * Number of local calendar days from `from` to `to`, ignoring the time of day.
 * Dividing elapsed milliseconds by a day length is off by one hour across a DST
 * change (the fall-back day has 25 hours), so this compares the local Y/M/D values.
 */
export function getCalendarDayDiff(from: Date, to: Date): number {
	const toDay = Date.UTC(to.getFullYear(), to.getMonth(), to.getDate())
	const fromDay = Date.UTC(from.getFullYear(), from.getMonth(), from.getDate())
	return Math.round((toDay - fromDay) / MILLISECONDS_A_DAY)
}

/**
 * Number of whole days a Gantt bar spans. A bar always covers at least one day,
 * otherwise a task that starts and ends before noon on the same day would be
 * rounded to zero width and not be drawn at all.
 */
export function getNaturalDayCount(start: Date, end: Date): number {
	// An end before noon rounds to that day's midnight and does not count the day itself.
	const endDayIncluded = end.getHours() < 12 ? 0 : 1
	return Math.max(1, getCalendarDayDiff(start, end) + endDayIncluded)
}
