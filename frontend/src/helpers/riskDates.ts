import dayjs from 'dayjs'

import {parseDateOrNull} from '@/helpers/parseDateOrNull'

// The form picks calendar days (<input type="date">), the API stores points in time.
// A day is sent as local midnight. A due date is sent as the end of the day instead, so a risk that
// is due today is not overdue until the day is over.

const DAY_FORMAT = 'YYYY-MM-DD'
const DAY_PATTERN = /^\d{4}-\d{2}-\d{2}$/

// '' or an invalid day gives null.
export function dayInputToIso(value: string, endOfDay = false): string | null {
	if (!DAY_PATTERN.test(value)) {
		return null
	}
	const [year, month, day] = value.split('-').map(Number)
	const date = endOfDay
		? new Date(year, month - 1, day, 23, 59, 59)
		: new Date(year, month - 1, day, 0, 0, 0)
	// 2026-02-30 would silently become March 2: only a day that exists is accepted.
	if (date.getFullYear() !== year || date.getMonth() !== month - 1 || date.getDate() !== day) {
		return null
	}
	return date.toISOString()
}

// The local calendar day of a stored date, '' when there is none.
export function isoToDayInput(value: string | null | undefined): string {
	const date = parseDateOrNull(value)
	return date ? dayjs(date).format(DAY_FORMAT) : ''
}

// Due before identified is a mistake. Either day missing is fine.
export function dueBeforeIdentified(identified: string, due: string): boolean {
	return DAY_PATTERN.test(identified) && DAY_PATTERN.test(due) && due < identified
}
