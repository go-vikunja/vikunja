/**
 * Helper functions for working with structured task repeat objects.
 * The API provides structured repeat data (ITaskRepeat) instead of raw RRULE strings.
 */

import type {ITaskRepeat} from '@/modelTypes/ITask'

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
export function repeatFromSettings(amount: number, freq: RepeatFrequency, bymonthday?: number): ITaskRepeat | null {
	if (amount <= 0) {
		return null
	}

	const repeat: ITaskRepeat = {
		freq: uiFreqToFreq(freq),
		interval: amount,
	}

	if (bymonthday !== undefined && bymonthday > 0) {
		repeat.byMonthDay = [bymonthday]
	}

	return repeat
}

/**
 * Extracts UI-friendly settings from a structured repeat object.
 */
export function repeatToSettings(repeat: ITaskRepeat | null): { amount: number; freq: RepeatFrequency; bymonthday?: number } | null {
	if (!repeat) {
		return null
	}
	return {
		amount: repeat.interval || 1,
		freq: freqToUiFreq(repeat.freq),
		bymonthday: repeat.byMonthDay?.[0],
	}
}

/**
 * Returns a human-readable description of a repeat configuration.
 */
export function describeRepeat(repeat: ITaskRepeat | null, t: (key: string, params?: Record<string, unknown>) => string): string {
	if (!repeat) {
		return ''
	}

	const freq = repeat.freq.toLowerCase()
	const interval = repeat.interval || 1
	const bymonthday = repeat.byMonthDay?.[0]

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
export function isRepeating(repeat: ITaskRepeat | null | undefined): boolean {
	return repeat != null && !!repeat.freq
}

/**
 * Returns true if the repeat rule uses options the simple editor cannot
 * represent (anything beyond freq, interval and a single byMonthDay). Such rules
 * must be treated as read-only here so the simple controls don't silently drop
 * the advanced parts (byDay, bySetPos, count, until, ...).
 */
export function isComplexRepeat(repeat: ITaskRepeat | null | undefined): boolean {
	if (!repeat) {
		return false
	}
	if ((repeat.byMonthDay?.length ?? 0) > 1) {
		return true
	}
	return (
		(repeat.byDay?.length ?? 0) > 0 ||
		(repeat.byMonth?.length ?? 0) > 0 ||
		(repeat.byYearDay?.length ?? 0) > 0 ||
		(repeat.byWeekNo?.length ?? 0) > 0 ||
		(repeat.bySetPos?.length ?? 0) > 0 ||
		(repeat.byHour?.length ?? 0) > 0 ||
		(repeat.byMinute?.length ?? 0) > 0 ||
		(repeat.bySecond?.length ?? 0) > 0 ||
		repeat.count != null ||
		(repeat.until != null && repeat.until !== '') ||
		(repeat.wkst != null && repeat.wkst !== '')
	)
}
