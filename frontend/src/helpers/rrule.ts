/**
 * RRULE helper functions for parsing and generating RFC 5545 recurrence rules.
 * This is a simplified implementation for the UI - the backend uses the full rrule-go library.
 */

export type RRuleFrequency = 'HOURLY' | 'DAILY' | 'WEEKLY' | 'MONTHLY' | 'YEARLY'

export interface ParsedRRule {
	freq: RRuleFrequency
	interval: number
	bymonthday?: number
}

/**
 * Parses an RRULE string into a structured object.
 * Example: "FREQ=DAILY;INTERVAL=2" -> { freq: 'DAILY', interval: 2 }
 */
export function parseRRule(rrule: string): ParsedRRule | null {
	if (!rrule || rrule.trim() === '') {
		return null
	}

	const parts: Record<string, string> = {}
	rrule.split(';').forEach(part => {
		const [key, value] = part.split('=')
		if (key && value) {
			parts[key.toUpperCase()] = value
		}
	})

	if (!parts.FREQ) {
		return null
	}

	const freq = parts.FREQ as RRuleFrequency
	const interval = parts.INTERVAL ? parseInt(parts.INTERVAL, 10) : 1
	const bymonthday = parts.BYMONTHDAY ? parseInt(parts.BYMONTHDAY, 10) : undefined

	return {
		freq,
		interval,
		bymonthday,
	}
}

/**
 * Generates an RRULE string from structured parameters.
 */
export function generateRRule(freq: RRuleFrequency, interval: number, bymonthday?: number): string {
	let rrule = `FREQ=${freq};INTERVAL=${interval}`
	if (bymonthday !== undefined && bymonthday > 0) {
		rrule += `;BYMONTHDAY=${bymonthday}`
	}
	return rrule
}

/**
 * UI-friendly frequency types that map to RRULE frequencies.
 */
export const REPEAT_FREQUENCIES = {
	Hours: 'hours',
	Days: 'days',
	Weeks: 'weeks',
	Months: 'months',
	Years: 'years',
} as const

export type RepeatFrequency = typeof REPEAT_FREQUENCIES[keyof typeof REPEAT_FREQUENCIES]

/**
 * Maps UI frequency to RRULE frequency.
 */
export function uiFreqToRRuleFreq(freq: RepeatFrequency): RRuleFrequency {
	switch (freq) {
		case 'hours':
			return 'HOURLY'
		case 'days':
			return 'DAILY'
		case 'weeks':
			return 'WEEKLY'
		case 'months':
			return 'MONTHLY'
		case 'years':
			return 'YEARLY'
		default:
			return 'DAILY'
	}
}

/**
 * Maps RRULE frequency to UI frequency.
 */
export function rruleFreqToUiFreq(freq: RRuleFrequency): RepeatFrequency {
	switch (freq) {
		case 'HOURLY':
			return 'hours'
		case 'DAILY':
			return 'days'
		case 'WEEKLY':
			return 'weeks'
		case 'MONTHLY':
			return 'months'
		case 'YEARLY':
			return 'years'
		default:
			return 'days'
	}
}

/**
 * Converts UI-friendly repeat settings to an RRULE string.
 */
export function repeatSettingsToRRule(amount: number, freq: RepeatFrequency, bymonthday?: number): string {
	if (amount <= 0) {
		return ''
	}
	return generateRRule(uiFreqToRRuleFreq(freq), amount, bymonthday)
}

/**
 * Parses an RRULE string to UI-friendly repeat settings.
 */
export function rruleToRepeatSettings(rrule: string): { amount: number; freq: RepeatFrequency; bymonthday?: number } | null {
	const parsed = parseRRule(rrule)
	if (!parsed) {
		return null
	}
	return {
		amount: parsed.interval,
		freq: rruleFreqToUiFreq(parsed.freq),
		bymonthday: parsed.bymonthday,
	}
}

/**
 * Returns a human-readable description of an RRULE.
 */
export function describeRRule(rrule: string, t: (key: string, params?: Record<string, unknown>) => string): string {
	const parsed = parseRRule(rrule)
	if (!parsed) {
		return ''
	}

	const { freq, interval, bymonthday } = parsed

	// Special cases for interval=1
	if (interval === 1) {
		switch (freq) {
			case 'HOURLY':
				return t('task.repeat.everyHour')
			case 'DAILY':
				return t('task.repeat.everyDay')
			case 'WEEKLY':
				return t('task.repeat.everyWeek')
			case 'MONTHLY':
				if (bymonthday) {
					return t('task.repeat.everyMonthOnDay', { day: bymonthday })
				}
				return t('task.repeat.everyMonth')
			case 'YEARLY':
				return t('task.repeat.everyYear')
		}
	}

	// General case with interval
	const freqKey = rruleFreqToUiFreq(freq)
	return t('task.repeat.everyN', { n: interval, unit: t(`task.repeat.${freqKey}`) })
}

/**
 * Checks if a task has a valid repeat configuration.
 */
export function isRepeating(repeats: string | undefined | null): boolean {
	return !!repeats && repeats.trim() !== ''
}
