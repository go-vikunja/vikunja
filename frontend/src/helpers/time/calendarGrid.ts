import {addDays} from '@/helpers/time/dateMath'

export interface CalendarCell {
	date: Date
	inMonth: boolean
}

// Always 6 rows so the grid never jumps in height between months.
export function buildMonthGrid(year: number, month: number, weekStart: number): CalendarCell[] {
	const first = new Date(year, month, 1)
	const lead = (first.getDay() - weekStart + 7) % 7
	const gridStart = addDays(first, -lead)

	return Array.from({length: 42}, (_, i) => {
		const date = addDays(gridStart, i)
		return {date, inMonth: date.getMonth() === month}
	})
}

export function weekdayOrder(weekStart: number): number[] {
	return Array.from({length: 7}, (_, i) => (i + weekStart) % 7)
}
