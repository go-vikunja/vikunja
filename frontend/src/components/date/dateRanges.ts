import {addJalaliDays, getJalaliMonthLength, instantToJalali, jalaliToInstant, weekdayInTimezone} from '@/helpers/time/jalali'

export const DATE_RANGES = {
	// Format: 
	// Key is the title, as a translation string, the first entry of the value array 
	// is the "from" date, the second one is the "to" date.
	'today': ['now/d', 'now/d+1d'],
	'tomorrow': ['now/d+1d', 'now/d+2d'],

	'lastWeek': ['now/w-1w', 'now/w'],
	'thisWeek': ['now/w', 'now/w+1w'],
	'restOfThisWeek': ['now', 'now/w+1w'],
	'nextWeek': ['now/w+1w', 'now/w+2w'],
	'next7Days': ['now', 'now+7d'],

	'lastMonth': ['now/M-1M', 'now/M'],
	'thisMonth': ['now/M', 'now/M+1M'],
	'restOfThisMonth': ['now', 'now/M+1M'],
	'nextMonth': ['now/M+1M', 'now/M+2M'],
	'next30Days': ['now', 'now+30d'],
	
	'thisYear': ['now/y', 'now/y+1y'],
	'restOfThisYear': ['now', 'now/y+1y'],
} as const

export const DATE_VALUES = {
	'now': 'now',
	'startOfToday': 'now/d',
	'endOfToday': 'now/d+1d',

	'beginningOflastWeek': 'now/w-1w',
	'endOfLastWeek': 'now/w',
	'beginningOfThisWeek': 'now/w',
	'endOfThisWeek': 'now/w+1w',
	'startOfNextWeek': 'now/w+1w',
	'endOfNextWeek': 'now/w+2w',
	'in7Days': 'now+7d',

	'beginningOfLastMonth': 'now/M-1M',
	'endOfLastMonth': 'now/M',
	'startOfThisMonth': 'now/M',
	'endOfThisMonth': 'now/M+1M',
	'startOfNextMonth': 'now/M+1M',
	'endOfNextMonth': 'now/M+2M',
	'in30Days': 'now+30d',

	'startOfThisYear': 'now/y',
	'endOfThisYear': 'now/y+1y',
} as const

export interface FaBounds {
	start: Date,
	end: Date,
}

// Saturday 00:00 containing ref → next Saturday 00:00 (Friday 24:00) in timeZone.
export function getFaWeekBounds(ref: Date, timeZone: string): FaBounds | null {
	const parts = instantToJalali(ref, timeZone)
	if (parts === null) {
		return null
	}
	const weekday = weekdayInTimezone(ref, timeZone)
	const offset = (weekday + 1) % 7
	const saturday = addJalaliDays({year: parts.year, month: parts.month, day: parts.day}, -offset)
	const start = jalaliToInstant({year: saturday.year, month: saturday.month, day: saturday.day, hours: 0, minutes: 0}, timeZone)
	const nextSaturday = addJalaliDays(saturday, 7)
	const end = jalaliToInstant({year: nextSaturday.year, month: nextSaturday.month, day: nextSaturday.day, hours: 0, minutes: 0}, timeZone)
	if (start === null || end === null) {
		return null
	}
	return {start, end}
}

// First day 00:00 of the Jalali month containing ref → first day 00:00 of the next month.
export function getFaMonthBounds(ref: Date, timeZone: string): FaBounds | null {
	const parts = instantToJalali(ref, timeZone)
	if (parts === null) {
		return null
	}
	const start = jalaliToInstant({year: parts.year, month: parts.month, day: 1, hours: 0, minutes: 0}, timeZone)
	if (start === null) {
		return null
	}
	const monthLength = getJalaliMonthLength(parts.year, parts.month)
	const nextStart = addJalaliDays({year: parts.year, month: parts.month, day: 1}, monthLength)
	const end = jalaliToInstant({year: nextStart.year, month: nextStart.month, day: nextStart.day, hours: 0, minutes: 0}, timeZone)
	if (end === null) {
		return null
	}
	return {start, end}
}

// Farvardin 1 00:00 containing ref → next Farvardin 1 00:00 (Esfand 29/30 leap aware).
export function getFaYearBounds(ref: Date, timeZone: string): FaBounds | null {
	const parts = instantToJalali(ref, timeZone)
	if (parts === null) {
		return null
	}
	const start = jalaliToInstant({year: parts.year, month: 1, day: 1, hours: 0, minutes: 0}, timeZone)
	if (start === null) {
		return null
	}
	const esfandLength = getJalaliMonthLength(parts.year, 12)
	const totalDays = 6 * 31 + 5 * 30 + esfandLength
	const nextStart = addJalaliDays({year: parts.year, month: 1, day: 1}, totalDays)
	const end = jalaliToInstant({year: nextStart.year, month: nextStart.month, day: nextStart.day, hours: 0, minutes: 0}, timeZone)
	if (end === null) {
		return null
	}
	return {start, end}
}

// Gregorian YYYY-MM-DD wall-clock in timeZone (kebab, no Jalali digits, no colons).
export function formatGregorianKebab(date: Date, timeZone: string): string {
	try {
		const parts = new Intl.DateTimeFormat('en-CA', {timeZone, year: 'numeric', month: '2-digit', day: '2-digit'}).formatToParts(date)
		const byType: Record<string, string> = {}
		for (const part of parts) {
			byType[part.type] = part.value
		}
		return `${byType.year}-${byType.month}-${byType.day}`
	} catch {
		return date.toISOString().split('T')[0] ?? ''
	}
}

// Single datemath value → explicit Gregorian ISO in fa mode. Day counts stay verbatim.
export function resolveFaDateValue(value: string, now: Date, timeZone: string): string {
	if (value === DATE_VALUES.beginningOflastWeek) {
		const week = getFaWeekBounds(now, timeZone)
		if (week === null) {
			return value
		}
		const prev = getFaWeekBounds(new Date(week.start.getTime() - 1), timeZone)
		if (prev === null) {
			return value
		}
		return prev.start.toISOString()
	}
	if (value === DATE_VALUES.beginningOfThisWeek || value === DATE_VALUES.endOfLastWeek) {
		const week = getFaWeekBounds(now, timeZone)
		return week === null ? value : week.start.toISOString()
	}
	if (value === DATE_VALUES.endOfThisWeek || value === DATE_VALUES.startOfNextWeek) {
		const week = getFaWeekBounds(now, timeZone)
		return week === null ? value : week.end.toISOString()
	}
	if (value === DATE_VALUES.endOfNextWeek) {
		const week = getFaWeekBounds(now, timeZone)
		if (week === null) {
			return value
		}
		const next = getFaWeekBounds(new Date(week.end.getTime() + 1), timeZone)
		if (next === null) {
			return value
		}
		return next.end.toISOString()
	}
	if (value === DATE_VALUES.beginningOfLastMonth) {
		const month = getFaMonthBounds(now, timeZone)
		if (month === null) {
			return value
		}
		const prev = getFaMonthBounds(new Date(month.start.getTime() - 1), timeZone)
		if (prev === null) {
			return value
		}
		return prev.start.toISOString()
	}
	if (value === DATE_VALUES.endOfLastMonth || value === DATE_VALUES.startOfThisMonth) {
		const month = getFaMonthBounds(now, timeZone)
		return month === null ? value : month.start.toISOString()
	}
	if (value === DATE_VALUES.endOfThisMonth || value === DATE_VALUES.startOfNextMonth) {
		const month = getFaMonthBounds(now, timeZone)
		return month === null ? value : month.end.toISOString()
	}
	if (value === DATE_VALUES.endOfNextMonth) {
		const month = getFaMonthBounds(now, timeZone)
		if (month === null) {
			return value
		}
		const next = getFaMonthBounds(new Date(month.end.getTime() + 1), timeZone)
		if (next === null) {
			return value
		}
		return next.end.toISOString()
	}
	if (value === DATE_VALUES.startOfThisYear) {
		const year = getFaYearBounds(now, timeZone)
		return year === null ? value : year.start.toISOString()
	}
	if (value === DATE_VALUES.endOfThisYear) {
		const year = getFaYearBounds(now, timeZone)
		return year === null ? value : year.end.toISOString()
	}
	return value
}

// Range pair → explicit Gregorian ISO bounds in fa mode. Day presets keep datemath.
export function resolveFaDateRange(from: string, to: string, now: Date, timeZone: string): [string, string] {
	return [resolveFaDateValue(from, now, timeZone), resolveFaDateValue(to, now, timeZone)]
}

// Single datemath value → explicit Gregorian kebab (YYYY-MM-DD wall-clock) in fa mode for filters that cannot carry ':'.
export function resolveFaDateValueKebab(value: string, now: Date, timeZone: string): string {
	const iso = resolveFaDateValue(value, now, timeZone)
	if (iso === value) {
		return value
	}
	const parsed = new Date(iso)
	if (Number.isNaN(parsed.getTime())) {
		return value
	}
	return formatGregorianKebab(parsed, timeZone)
}

// Range pair → explicit Gregorian kebab bounds in fa mode. Day presets keep datemath.
export function resolveFaDateRangeKebab(from: string, to: string, now: Date, timeZone: string): [string, string] {
	return [resolveFaDateValueKebab(from, now, timeZone), resolveFaDateValueKebab(to, now, timeZone)]
}
