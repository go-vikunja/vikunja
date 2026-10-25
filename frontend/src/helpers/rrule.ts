/**
 * Helper functions for working with structured task repeat objects.
 * The API provides structured repeat data (TaskRepeat) instead of raw RRULE strings.
 */

import type {TaskRepeat} from '@/client/generated'

export const REPEAT_FREQUENCIES = {
	Hours: 'hours',
	Days: 'days',
	Weeks: 'weeks',
	Months: 'months',
	Years: 'years',
} as const

export type RepeatFrequency = typeof REPEAT_FREQUENCIES[keyof typeof REPEAT_FREQUENCIES]

const FREQ_TO_UI: Record<string, RepeatFrequency> = {
	hourly: 'hours',
	daily: 'days',
	weekly: 'weeks',
	monthly: 'months',
	yearly: 'years',
}

const UI_TO_FREQ: Record<RepeatFrequency, string> = {
	hours: 'hourly',
	days: 'daily',
	weeks: 'weekly',
	months: 'monthly',
	years: 'yearly',
}

/**
 * Whether the simple editor can represent this freq. Freqs it can't map
 * (MINUTELY/SECONDLY, or anything unknown) have no matching unit in the dropdown.
 */
export function isMappableFreq(freq: string | null | undefined): boolean {
	return !!freq && freq.toLowerCase() in FREQ_TO_UI
}

/**
 * Maps an API freq string to a UI frequency.
 */
export function freqToUiFreq(freq: string): RepeatFrequency {
	return FREQ_TO_UI[freq.toLowerCase()] || 'days'
}

/**
 * Maps a UI frequency to an API freq string.
 */
export function uiFreqToFreq(freq: RepeatFrequency): string {
	return UI_TO_FREQ[freq] || 'daily'
}

/**
 * Creates a structured repeat object from UI-friendly settings.
 */
export function repeatFromSettings(amount: number, freq: RepeatFrequency, bymonthday?: number): TaskRepeat | null {
	if (amount <= 0) {
		return null
	}

	const repeat: TaskRepeat = {
		freq: uiFreqToFreq(freq),
		interval: amount,
	}

	if (bymonthday !== undefined && bymonthday > 0) {
		repeat.by_month_day = [bymonthday]
	}

	return repeat
}

/**
 * Returns a human-readable description of a repeat configuration.
 */
export function describeRepeat(
	repeat: TaskRepeat | null | undefined,
	t: (key: string, params?: Record<string, unknown>) => string,
): string {
	if (!repeat) {
		return ''
	}

	const freq = (repeat.freq ?? '').toLowerCase()
	const interval = repeat.interval || 1
	const bymonthday = repeat.by_month_day?.[0]

	// Special cases for interval=1
	if (interval === 1) {
		switch (freq) {
			case 'hourly':
				return t('task.repeat.everyHour')
			case 'daily':
				return t('task.repeat.everyDay')
			case 'weekly':
				return t('task.repeat.everyWeek')
			case 'monthly':
				if (bymonthday) {
					return t('task.repeat.everyMonthOnDay', {day: bymonthday})
				}
				return t('task.repeat.everyMonth')
			case 'yearly':
				return t('task.repeat.everyYear')
		}
	}

	// General case with interval
	const freqKey = freqToUiFreq(freq)
	return t('task.repeat.everyN', {n: interval, unit: t(`task.repeat.${freqKey}`)})
}

/**
 * Checks if a task has a valid repeat configuration.
 */
export function isRepeating(repeat: TaskRepeat | null | undefined): boolean {
	return repeat != null && !!repeat.freq
}

/**
 * Returns true if the repeat rule uses options the simple editor cannot
 * represent (anything beyond freq, interval and, on a monthly rule, a single
 * by_month_day from 1 to 31). Such rules
 * must be treated as read-only here so the simple controls don't silently drop
 * the advanced parts (by_day, by_set_pos, count, until, ...).
 */
export function isComplexRepeat(repeat: TaskRepeat | null | undefined): boolean {
	if (!repeat) {
		return false
	}
	// The server flags rules whose stored text has parts this structure can't hold (BYEASTER, DTSTART).
	if (repeat.read_only) {
		return true
	}
	// A freq the simple editor doesn't map (MINUTELY/SECONDLY or anything unknown)
	// would fall back to "daily" and silently rewrite the rule, so treat it as complex.
	if (repeat.freq && !isMappableFreq(repeat.freq)) {
		return true
	}
	// The day-of-month control offers days 1-31 and only for monthly rules, so a single by_month_day
	// is simple only there. Any other (a negative day such as -1 for the last day, or a by_month_day
	// on another freq) would be regenerated without it.
	const monthDays = repeat.by_month_day ?? []
	if (monthDays.length > 1) {
		return true
	}
	if (monthDays.length === 1) {
		const day = monthDays[0]!
		if ((repeat.freq ?? '').toLowerCase() !== 'monthly' || !Number.isInteger(day) || day < 1 || day > 31) {
			return true
		}
	}
	return (
		(repeat.by_day?.length ?? 0) > 0 ||
		(repeat.by_month?.length ?? 0) > 0 ||
		(repeat.by_year_day?.length ?? 0) > 0 ||
		(repeat.by_week_no?.length ?? 0) > 0 ||
		(repeat.by_set_pos?.length ?? 0) > 0 ||
		(repeat.by_hour?.length ?? 0) > 0 ||
		(repeat.by_minute?.length ?? 0) > 0 ||
		(repeat.by_second?.length ?? 0) > 0 ||
		repeat.count != null ||
		(repeat.until != null && repeat.until !== '') ||
		(repeat.wkst != null && repeat.wkst !== '')
	)
}
